package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"telepight/handlers"
	"telepight/processors"
	"telepight/templates"
	"telepight/utils"
)

func main() {

	// load flags and configs
	primaryHost := flag.String("m", "", "The primary server hostname.")
	isVideoSource := flag.Bool("s", false, "Whether this instance serves as a camera source.")
	flag.Parse()

	// load all the templates
	basePath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}
	templates := templates.ParseAllTemplates(basePath)

	// shared state
	var sources []processors.StreamSource
	if *isVideoSource {
		sources = append(sources, processors.StreamSource{
			Hostname: utils.GetHostName(),
		})
	}
	streamSources := processors.StreamSources{
		Sources: sources,
	}
	streamChan := make(chan processors.StreamSource)
	defer close(streamChan)

	// view handlers
	http.HandleFunc("/stream/", handlers.StreamViewHandler(templates, &streamSources))
	http.HandleFunc("/watch/", handlers.VideosWatchHandler(templates, &streamSources))
	// api handlers
	http.HandleFunc("/", handlers.LandingViewHandler(templates))
	http.HandleFunc("/listvideos/", handlers.VideosListHandler(basePath))
	http.HandleFunc("/subscribe/", handlers.RequestSubscriptionHandler(streamChan))
	// serve static video files
	http.Handle("/videos/", http.FileServer(http.Dir(basePath)))

	// start background routines
	go processors.RequestSubscription(*primaryHost)
	go processors.AcceptSubscription(&streamSources, streamChan)
	go processors.StartAutoRecording(basePath)

	log.Fatal(http.ListenAndServe(":8080", nil))
}
