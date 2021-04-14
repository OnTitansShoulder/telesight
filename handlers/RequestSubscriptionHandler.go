package handlers

import (
	"fmt"
	"net/http"
	"telepight/processors"
)

var ()

func RequestSubscriptionHandler(streamChan chan processors.StreamSource) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := r.URL.Query()
		host := params["host"]
		if len(host) == 0 || len(host[0]) == 0 {
			http.Error(w, "must provide a host param", http.StatusBadRequest)
		}

		streamSource := processors.StreamSource{
			Hostname: host[0],
		}
		streamChan <- streamSource

		fmt.Fprint(w, "ok")
	}
}
