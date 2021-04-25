package processors

import (
	"fmt"
	"log"
	"net/http"
	"telesight/utils"
	"time"
)

type StreamSources struct {
	Sources map[string]StreamSource
}

type StreamSource struct {
	Hostname string
	IP       string
	Port     int
}

func (s *StreamSources) addStreamSource(newSource StreamSource) {
	// check for duplicated source
	if existing, ok := s.Sources[newSource.Hostname]; ok {
		if existing.IP == newSource.IP {
			return
		}
	}
	log.Printf("Accepting new subscription from host=%s\n", newSource.Hostname)
	s.Sources[newSource.Hostname] = newSource
}

func RequestSubscription(primaryHost, ip string) {
	if utils.IsPrimaryHost(primaryHost) {
		log.Println("This is the primary instance, no need to request subscription.")
		return
	}

	if len(primaryHost) == 0 {
		log.Println("Primary host is not specified. Stop subscribing to it.")
	}

	host := utils.GetHostName()
	subscriptionURL := RequestSubscriptionURL(primaryHost, host, ip)
	for {
		resp, err := http.Get(subscriptionURL)
		if err != nil {
			log.Fatalf("Failed to request subscription from the primary host %s\n", primaryHost)
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

func RequestSubscriptionURL(primaryHost, host, ip string) string {
	return fmt.Sprintf("http://%s.local/telesight/subscribe/?host=%s&ip=%s", primaryHost, host, ip)
}
