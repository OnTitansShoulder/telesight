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
	videosURLTemplate = "http://%s/telesight/listvideos/"
)

type VideoListPageData struct {
	VideoSources          []HostData
	SelectedVideoSourceIP string
	Videos                []string
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
		hostParam := params["host"]
		if len(hostParam) == 0 || len(hostParam[0]) == 0 {
			respondWithTemplate(t, w, VideoListPageData{
				VideoSources: videoSources,
				Videos:       []string{},
			})
			return
		}
		host := hostParam[0]
		ip := streamSources.Sources[host].IP
		resp, err := http.Get(fmt.Sprintf(videosURLTemplate, ip))
		if err != nil {
			log.Printf("Failed to request video list from host=%s ip=%s: %v\n", host, ip, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Failed to read video list from host=%s ip=%s: %v\n", host, ip, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		var videoFileNames []string
		err = json.Unmarshal(body, &videoFileNames)
		if err != nil {
			log.Printf("Failed to unmarshall video list from host=%s ip=%s: %v\n", host, ip, err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if len(videoFileNames) > 0 {
			// display the files in reverse-chronological order
			utils.ReverseAny(videoFileNames)
		}

		respondWithTemplate(t, w, VideoListPageData{
			VideoSources:          videoSources,
			SelectedVideoSourceIP: ip,
			Videos:                videoFileNames,
		})
	}
}

func respondWithTemplate(t *template.Template, w http.ResponseWriter, data VideoListPageData) {
	err := t.ExecuteTemplate(w, videolistTemplate, data)
	if err != nil {
		log.Printf("Error processing template=%s: %v\n", videolistTemplate, err)
	}
}

func VideosListHandler(basePath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		videosPath := filepath.Join(basePath, videoDirName)
		files, err := ioutil.ReadDir(videosPath)
		if err != nil {
			log.Printf("Failed to read files from videosPath=%s: %v\n", videosPath, err)
			fmt.Fprint(w, "[]")
			return
		}
		var fileNames []string
		for _, file := range files {
			fileNames = append(fileNames, file.Name())
		}
		bytes, err := json.Marshal(fileNames)
		if err != nil {
			log.Printf("Failed to marshall fileNames into JSON\n")
			fmt.Fprint(w, "[]")
			return
		}
		fmt.Fprint(w, string(bytes))
	}
}
