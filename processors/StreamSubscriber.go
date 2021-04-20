package processors

import (
	"fmt"
	"log"
	"net/http"
	"telesight/utils"
	"time"
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
	log.Printf("Accepting new subscription from host=%s", newSource.Hostname)
	s.Sources = append(s.Sources, newSource)
}

func RequestSubscription(primaryHost string) {
	if utils.IsPrimaryHost(primaryHost) {
		log.Println("This is the primary instance, no need to request subscription.")
		return
	}

	host := utils.GetHostName()
	subscriptionURL := RequestSubscriptionURL(primaryHost, host)
	for {
		resp, err := http.Get(subscriptionURL)
		if err != nil {
			log.Fatalf("Failed to request subscription from the primary host %s", primaryHost)
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

func RequestSubscriptionURL(primaryHost, host string) string {
	return fmt.Sprintf("http://%s.local/subscribe/?host=%s", primaryHost, host)
}
