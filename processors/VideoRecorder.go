package processors

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

var (
	// config values
	streamURL    = "http://192.168.1.36/webcam/?action=snapshot"
	frameDirName = "frames"
	fps          = 1
	duration     = 900 // in seconds
	videoDirName = "camrecord"

	// derived
	frameGap    = 1000 / fps
	totalFrames = fps * duration
)

// StartAutoRecording calls the local stream URL and saves
// jpg frames to filesystem.
// It also spawns processes to use ffmpeg to create videos out of the frames
// and clean up old videos and the used frames
func StartAutoRecording(basePath string) {
	status, body, err := get(streamURL)
	if err != nil || status > 399 {
		log.Printf("Error: unabled to reach the stream server for saving frames. status=%d body=%v err=%v", status, body, err)
		return
	}

	// init frame dir
	frameDir := filepath.Join(basePath, frameDirName)
	ensureDirExists(frameDir)

	// init video dir
	videoDir := filepath.Join(basePath, videoDirName)
	ensureDirExists(videoDir)

	// TODO start video cleaning process

	videoCounter := 0 // the ith video within current hour
	frameCounter := 0
	lastDir := ""
	// start saving frames
	for {
		dirPrefix := getImageDatePrefix()
		if lastDir != dirPrefix {
			// TODO start process to create video
			videoCounter++
			frameCounter = 0
			lastDir = dirPrefix
		}

		tempDir := filepath.Join(frameDir, dirPrefix, strconv.Itoa(videoCounter))
		if frameCounter == 0 {
			ensureDirExists(tempDir)
		}

		// save next frame
		saveFrame(tempDir, frameCounter)
		frameCounter++

		// time gap before next iteration
		time.Sleep(time.Millisecond * time.Duration(frameGap))
		if frameCounter >= totalFrames { // need to switch to new video
			// TODO start process to create video
			videoCounter++
			frameCounter = 0
		}
	}
}

func ensureDirExists(dirPath string) {
	err := os.Mkdir(dirPath, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("Error: Unable to create directory %s: %v", dirPath, err)
	}
}

func getImageDatePrefix() string {
	now := time.Now()
	return now.Format("01-02_15") // gives month-day_hour
}

func saveFrame(dir string, frameNumber int) {
	status, body, err := get(streamURL)
	if err != nil {
		log.Fatalf("Error: Failed to save frame: %v", err)
	}
	if status > 399 {
		log.Fatalf("Error: Got status code > 399, status=%d", status)
	}

	frameFile := filepath.Join(dir, strconv.Itoa(frameNumber)+".jpg")
	err = ioutil.WriteFile(frameFile, body, 0644)
	if err != nil {
		log.Fatalf("Error: Unable to write frame to file=%s", frameFile)
	}
}

func get(url string) (int, []byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	return resp.StatusCode, body, err
}
