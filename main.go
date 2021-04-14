package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"telepight/handlers"
	"telepight/processors"
	"telepight/templates"
)

func main() {

	// load all the templates
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	templates := templates.ParseAllTemplates(basePath)

	// shared state
	streamSources := processors.StreamSources{
		Sources: []processors.StreamSource{
			{
				Hostname: processors.GetHostName(),
			},
		},
	}
	streamChan := make(chan processors.StreamSource)
	defer close(streamChan)

	// view handlers
	http.HandleFunc("/", handlers.LandingViewHandler(templates))
	http.HandleFunc("/stream/", handlers.StreamViewHandler(templates, &streamSources))
	// api handlers
	http.HandleFunc("/subscribe/", handlers.RequestSubscriptionHandler(streamChan))

	// start background routines
	go processors.RequestSubscription()
	go processors.AcceptSubscription(&streamSources, streamChan)
	go processors.StartAutoRecording(basePath)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
