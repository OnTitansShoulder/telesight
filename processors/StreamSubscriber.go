package processors

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	PrimaryHost = "zk-gatekeeper"
)

type StreamSources struct {
	Sources []StreamSource
}

type StreamSource struct {
	Hostname string
	IP       string
	Port     int
}

func (s *StreamSources) addStreamSource(newSource StreamSource) {
	// check for duplicated source
	for _, source := range s.Sources {
		if source.Hostname == newSource.Hostname {
			return // TODO add check for IP later, and override if IP is different
		}
	}
	s.Sources = append(s.Sources, newSource)
}

func RequestSubscription() {
	if IsPrimaryHost() {
		log.Println("This is the primary instance, no need to request subscription.")
		return
	}

	host := GetHostName()
	subscriptionURL := RequestSubscriptionURL(host)
	for {
		resp, err := http.Get(subscriptionURL)
		if err != nil {
			log.Printf("Failed to request subscription from the primary host %s", PrimaryHost)
		}

		log.Println(resp.Body) // TODO parse the response and populate local stream list

		time.Sleep(time.Minute)
	}
}

func AcceptSubscription(streamSources *StreamSources, streamChan chan StreamSource) {
	for {
		streamSource := <-streamChan
		streamSources.addStreamSource(streamSource)
	}
}

//
// Helper functions
//

func GetHostName() string {
	host, exist := os.LookupEnv("HOSTNAME")
	if !exist {
		log.Fatal("HOSTNAME is not set!")
	}
	return host
}

func IsPrimaryHost() bool {
	if PrimaryHost == GetHostName() {
		return true
	}
	return false
}

func RequestSubscriptionURL(host string) string {
	return fmt.Sprintf("http://%s.local/subscribe/?host=%s", PrimaryHost, host)
}
