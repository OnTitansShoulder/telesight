package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"telesight/processors"
	"telesight/utils"
)

const (
	VideoListURLPath  = "/watch/"
	videolistTemplate = "video-list.gtpl"
	videoDirName      = "videos"
	videosURLTemplate = "http://%s.local/listvideos/"
)

type VideoListPageData struct {
	VideoSources []HostData
	Videos       []string
}

func VideosWatchHandler(t *template.Template, streamSources *processors.StreamSources) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != VideoListURLPath {
			http.NotFound(w, r)
			return
		}
		var videoSources []HostData
		for _, streamSource := range streamSources.Sources {
			videoSources = append(videoSources, HostData{
				DisplayText: streamSource.Hostname,
			})
		}
		params := r.URL.Query()
		host := params["host"]
		if len(host) == 0 || len(host[0]) == 0 {
			respondWithTemplate(t, w, VideoListPageData{
				VideoSources: videoSources,
				Videos:       []string{},
			})
			return
		}
		resp, err := http.Get(VideosURL(host[0]))
		if err != nil {
			log.Printf("Failed to request video list from host=%s: %v", host, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read video list from host=%s: %v", host, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		var videoFileNames []string
		err = json.Unmarshal(body, &videoFileNames)
		if err != nil {
			log.Printf("Failed to unmarshall video list from host=%s: %v", host, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		respondWithTemplate(t, w, VideoListPageData{
			VideoSources: videoSources,
			Videos:       videoFileNames,
		})
	}
}

func respondWithTemplate(t *template.Template, w http.ResponseWriter, data VideoListPageData) {
	err := t.ExecuteTemplate(w, videolistTemplate, data)
	if err != nil {
		log.Printf("Error processing template=%s: %v", videolistTemplate, err)
	}
}

func VideosURL(host string) string {
	if host == utils.GetHostName() {
		return "http://localhost:8080/listvideos/"
	}
	return fmt.Sprintf(videosURLTemplate, host)
}

func VideosListHandler(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videosPath := filepath.Join(basePath, videoDirName)
		files, err := ioutil.ReadDir(videosPath)
		if err != nil {
			log.Printf("Failed to read files from videosPath=%s: %v", videosPath, err)
			fmt.Fprint(w, "[]")
			return
		}
		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}
		bytes, err := json.Marshal(fileNames)
		if err != nil {
			log.Printf("Failed to marshall fileNames into JSON")
			fmt.Fprint(w, "[]")
			return
		}
		fmt.Fprint(w, string(bytes))
	}
}
