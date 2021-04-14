package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"telepight/processors"
)

const (
	StreamPath        = "/stream/"
	streamURLTemplate = "http://%s.local/webcam/?action=stream"
)

type StreamPageData struct {
	StreamSources []StreamSourceData
}

type StreamSourceData struct {
	DisplayText string
	StreamURL   string
}

func StreamViewHandler(t *template.Template, streamSources *processors.StreamSources) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validatePath(w, r) {
			return
		}
		pageData := generateStreamPage(streamSources)
		err := t.ExecuteTemplate(w, "stream.gtpl", pageData)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

func validatePath(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != StreamPath {
		http.NotFound(w, r)
		return false
	}
	return true
}

func generateStreamPage(streamSources *processors.StreamSources) StreamPageData {
	var pageData []StreamSourceData
	for _, streamSource := range streamSources.Sources {
		pageData = append(pageData, StreamSourceData{
			DisplayText: streamSource.Hostname,
			StreamURL:   fmt.Sprintf(streamURLTemplate, streamSource.Hostname),
		})
	}
	return StreamPageData{
		StreamSources: pageData,
	}
}
