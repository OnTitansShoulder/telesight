package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"telesight/handlers"
	"telesight/processors"
	"telesight/templates"
	"telesight/utils"
	"time"
)

func main() {

	// load flags and configs
	primaryHost := flag.String("m", "", "The primary server hostname.")
	base := flag.String("b", "", "The basePath to save runtime buffer and images.")
	isVideoSource := flag.Bool("s", false, "Whether this instance serves as a camera source.")
	subscribePrimaryHost := flag.Bool("r", false, "Whether this instance should report to primary instance.")
	warmUpSeconds := flag.Int("w", 0, "Warm up seconds for starting up mjpg_streamer")
	flag.Parse()

	ipAddr := utils.GetIpAddr()

	if *isVideoSource {
		if *warmUpSeconds > 0 {
			log.Printf("Sleep %d seconds before starting up...\n", *warmUpSeconds)
			time.Sleep(time.Duration(*warmUpSeconds) * time.Second)
		}
	}

	// load all the templates
	var basePath string
	var err error
	if base != nil && len(*base) > 0 {
		basePath = *base
	} else {
		basePath, err = filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			log.Fatal(err)
		}
	}
	templates := templates.ParseAllTemplates(basePath)

	// shared state
	sources := make(map[string]processors.StreamSource)
	if *isVideoSource {
		host := utils.GetHostName()
		sources[host] = processors.StreamSource{
			Hostname: host,
			IP:       ipAddr,
		}
	}
	streamSources := processors.StreamSources{
		Sources: sources,
	}
	streamChan := make(chan processors.StreamSource)
	defer close(streamChan)

	http.DefaultClient.Timeout = time.Minute
	// view handlers
	http.HandleFunc("/stream/", handlers.StreamViewHandler(templates, &streamSources))
	http.HandleFunc("/watch/", handlers.VideosWatchHandler(templates, &streamSources))
	// api handlers
	http.HandleFunc("/health/", handlers.LandingViewHandler(templates))
	http.HandleFunc("/listvideos/", handlers.VideosListHandler(basePath))
	http.HandleFunc("/subscribe/", handlers.RequestSubscriptionHandler(streamChan))
	http.HandleFunc("/", handlers.StreamViewRedirectHandler())
	// serve static video files
	http.Handle("/videos/", http.FileServer(http.Dir(basePath)))

	// start background routines
	go processors.AcceptSubscription(&streamSources, streamChan)
	if utils.IsPrimaryHost(*primaryHost) {
		go processors.CheckHeartBeats(&streamSources)
	}
	if *subscribePrimaryHost {
		go processors.RequestSubscription(*primaryHost, ipAddr)
	}
	if *isVideoSource {
		go processors.StartAutoRecording(basePath)
	}

	log.Fatal(http.ListenAndServe(":8089", nil))
}
