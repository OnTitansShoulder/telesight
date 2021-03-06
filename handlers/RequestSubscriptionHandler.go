package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"telesight/processors"
)

const (
	SubscribeURLPath = "/subscribe/"
)

func RequestSubscriptionHandler(streamChan chan processors.StreamSource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, SubscribeURLPath) {
			http.NotFound(w, r)
			return
		}
		params := r.URL.Query()
		host := params["host"]
		if len(host) == 0 || len(host[0]) == 0 {
			http.Error(w, "must provide a host param", http.StatusBadRequest)
		}
		ip := params["ip"]
		if len(ip) == 0 || len(ip[0]) == 0 {
			http.Error(w, "must provide an ip param", http.StatusBadRequest)
		}

		streamSource := processors.StreamSource{
			Hostname: host[0],
			IP:       ip[0],
		}
		streamChan <- streamSource

		fmt.Fprint(w, "ok")
	}
}
