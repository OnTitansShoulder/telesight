package processors

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

var (
	// config values
	streamURL          = "http://zk-cam02.local/webcam/?action=snapshot"
	framesDirName      = "frames"
	frameFileNameLimit = 10 // fill frame filename up to this length with '0's
	fps                = 1
	duration           = 600 // in seconds
	videosDirName      = "videos"

	frameFetchFailureThreshold = 10
	maxNumberOfVideos          = 36

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
		log.Printf("Error: unabled to reach the stream server for saving frames. status=%d body=%v err=%v", status, string(body), err)
		return
	}

	// init frame dir
	framesDir := filepath.Join(basePath, framesDirName)
	ensureDirExists(framesDir)

	// init video dir
	videosDir := filepath.Join(basePath, videosDirName)
	ensureDirExists(videosDir)

	// TODO start video cleaning process

	videoCounter := 0 // the ith video within current hour
	frameCounter := 0
	lastPrefix := getFramesPrefix(videoCounter)
	// start saving frames
	for {
		prefix := getFramesPrefix(videoCounter)
		if lastPrefix != prefix && frameCounter > 0 {
			go StartEncodingVideo(lastPrefix, framesDir, videosDir)
			videoCounter = 0
			frameCounter = 0
			// need to reset the video counter since this is a new hour
			prefix = getFramesPrefix(videoCounter)
			lastPrefix = prefix
		}

		tempDir := filepath.Join(framesDir, prefix)
		if frameCounter == 0 {
			ensureDirExists(tempDir)
		}
		// TODO check for existing frames within the tempDir and increase base from there

		// save next frame
		SaveFrame(tempDir, frameCounter)
		frameCounter++

		// time gap before next iteration
		time.Sleep(time.Millisecond * time.Duration(frameGap))
		if frameCounter >= totalFrames { // need to switch to new video
			go StartEncodingVideo(lastPrefix, framesDir, videosDir)
			videoCounter++
			frameCounter = 0
			prefix = getFramesPrefix(videoCounter)
			lastPrefix = prefix
		}
	}
}

func StartEncodingVideo(prefix, srcDir, videosDir string) {
	framesDir := filepath.Join(srcDir, prefix)
	defer postEncodingCleanup(framesDir, videosDir)

	outputVideo := filepath.Join(videosDir, fmt.Sprintf("%s.mp4", prefix))
	ffmpegCmd := getFFmpegCommand()
	ffmpegArgs := []string{"-framerate", "4", "-i", filepath.Join(framesDir, "%"+strconv.Itoa(frameFileNameLimit)+"d.jpg"), "-c:v", "libx264", "-pix_fmt", "yuv420p", outputVideo}
	err := exec.Command(ffmpegCmd, ffmpegArgs...).Run()
	if err != nil {
		log.Fatalf("Failed encoding video from %s: %v", framesDir, err)
	}
	log.Printf("Encoding video for %s was successful.", framesDir)
}

func postEncodingCleanup(framesDir, videosDir string) {
	err := os.RemoveAll(framesDir)
	if err != nil {
		log.Printf("Failed to remove framesDir=%s: %v", framesDir, err)
		log.Printf("Falling back to use command 'rm -rf'")
		// TODO write rm -rf command
	}

	// remove old videos that takes space
	files, err := ioutil.ReadDir(videosDir)
	if err != nil {
		log.Printf("Failed to check files from videosDir=%s: %v", videosDir, err)
	}
	videosToRemove := len(files) - maxNumberOfVideos
	for i := 0; i < videosToRemove; i++ {
		fileName := files[i].Name()
		err = os.Remove(filepath.Join(videosDir, fileName))
		if err != nil {
			log.Printf("Failed to remove file=%s under dir=%s : %v", fileName, videosDir, err)
		}
	}
}

func getFFmpegCommand() string {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		log.Fatal("Cannot find ffmpeg from PATH")
	}
	return path
}

func ensureDirExists(dirPath string) {
	err := os.Mkdir(dirPath, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("Error: Unable to create directory %s: %v", dirPath, err)
	}
}

func getFramesPrefix(videoCounter int) string {
	now := time.Now()
	return now.Format("2006-01-02_15") + "_" + strconv.Itoa(videoCounter) // gives month-day_hour_videoCounter
}

func SaveFrame(dir string, frameNumber int) {
	var imageData []byte
	for try := 0; try < frameFetchFailureThreshold; try++ {
		status, body, err := get(streamURL)
		if err != nil {
			log.Printf("Error (try=%d): Failed to save frame: %v", try, err)
		}
		if status > 399 {
			log.Printf("Error (try=%d): Got status code > 399, status=%d", try, status)
		}
		imageData = body
		break
	}

	frameFile := filepath.Join(dir, getFrameFileName(frameNumber))
	err := ioutil.WriteFile(frameFile, imageData, 0644)
	if err != nil {
		log.Fatalf("Error: Unable to write frame to file=%s", frameFile)
	}
}

func getFrameFileName(counter int) string {
	numbDigits := 1
	remain := counter / 10
	for remain > 0 {
		numbDigits++
		remain = remain / 10
	}

	fillLength := frameFileNameLimit - numbDigits
	frameFileName := strconv.Itoa(counter) + ".jpg"
	for fillLength > 0 {
		frameFileName = "0" + frameFileName
		fillLength--
	}
	return frameFileName
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
