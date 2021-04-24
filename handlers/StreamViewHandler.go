package handlers

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"telesight/processors"
)

const (
	StreamURLPath     = "/stream/"
	streamTemplate    = "stream.gtpl"
	streamURLTemplate = "http://%s.local/webcam/?action=stream"
)

type StreamPageData struct {
	StreamSources []HostData
}

type HostData struct {
	DisplayText string
	URL         string
}

func StreamViewHandler(t *template.Template, streamSources *processors.StreamSources) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != StreamURLPath {
			http.NotFound(w, r)
			return
		}
		pageData := generateStreamPage(streamSources)
		err := t.ExecuteTemplate(w, streamTemplate, pageData)
		if err != nil {
			log.Printf("Error processing template=%s: %v\n", streamTemplate, err)
		}
	}
}

func generateStreamPage(streamSources *processors.StreamSources) StreamPageData {
	var pageData []HostData
	for _, streamSource := range streamSources.Sources {
		pageData = append(pageData, HostData{
			DisplayText: streamSource.Hostname,
			URL:         fmt.Sprintf(streamURLTemplate, streamSource.Hostname),
		})
	}
	return StreamPageData{
		StreamSources: pageData,
	}
}
