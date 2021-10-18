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
	"strings"
	"time"
)

var (
	// config values
	streamURL          = "http://localhost:8080/?action=snapshot"
	framesDirName      = "frames"
	frameFileNameLimit = 10 // fill frame filename up to this length with '0's
	fps                = 1
	duration           = 900 // in seconds
	videosDirName      = "videos"

	frameFetchFailureThreshold = 10
	maxNumberOfVideos          = 4 * 24 * 3 // keep three days worth of videos

	// derived
	frameGap    = 1000 / fps
	totalFrames = fps * duration

	// video encoding retry
	videoEncodeRetryLimit = 5
)

// StartAutoRecording calls the local stream URL and saves
// jpg frames to filesystem.
// It also spawns processes to use ffmpeg to create videos out of the frames
// and clean up old videos and the used frames
func StartAutoRecording(basePath string) {
	status, body, err := get(streamURL)
	if err != nil || status > 399 {
		log.Printf("Error: unabled to reach the stream server for saving frames. status=%d body=%v err=%v\n", status, string(body), err)
		return
	}

	// init frame dir
	framesDir := filepath.Join(basePath, framesDirName)
	ensureDirExists(framesDir)

	// init video dir
	videosDir := filepath.Join(basePath, videosDirName)
	ensureDirExists(videosDir)

	videoCounter := initNextVideoCounter(videosDir)
	lastPrefix := getFramesPrefix(videoCounter)
	frameCounter := initNextFrameCounter(framesDir, lastPrefix)
	log.Printf("starting with prefix=%s videoCounter=%d frameCounter=%d\n", lastPrefix, videoCounter, frameCounter)
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

	try := 0
	for try < videoEncodeRetryLimit {
		try++
		outputVideo := filepath.Join(videosDir, fmt.Sprintf("%s.mp4", prefix))
		ffmpegCmd := getFFmpegCommand()
		ffmpegArgs := []string{"-framerate", "4", "-i", filepath.Join(framesDir, "%"+strconv.Itoa(frameFileNameLimit)+"d.jpg"), "-c:v", "libx264", "-pix_fmt", "yuv420p", outputVideo}
		cmdOutput, err := exec.Command(ffmpegCmd, ffmpegArgs...).CombinedOutput()
		if err == nil {
			break
		}
		log.Printf("Failed encoding video from %s (try %d): %v: %s\n", framesDir, try, err, string(cmdOutput))
	}
	if try == videoEncodeRetryLimit {
		log.Printf("Retry has exhausted for encoding video from %s\n", framesDir)
	} else {
		log.Printf("Encoding video for %s was successful.\n", framesDir)
	}
}

func postEncodingCleanup(framesDir, videosDir string) {
	err := os.RemoveAll(framesDir)
	if err != nil {
		log.Printf("Failed to remove framesDir=%s: %v\n", framesDir, err)
		log.Printf("Falling back to use command 'rm -rf'\n")
		// TODO write rm -rf command
	}

	// remove old videos that takes space
	files, err := ioutil.ReadDir(videosDir)
	if err != nil {
		log.Printf("Failed to check files from videosDir=%s: %v\n", videosDir, err)
	}
	videosToRemove := len(files) - maxNumberOfVideos
	for i := 0; i < videosToRemove; i++ {
		fileName := files[i].Name()
		err = os.Remove(filepath.Join(videosDir, fileName))
		if err != nil {
			log.Printf("Failed to remove file=%s under dir=%s : %v\n", fileName, videosDir, err)
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
		log.Fatalf("Error: Unable to create directory %s: %v\n", dirPath, err)
	}
}

func getFramesPrefix(videoCounter int) string {
	now := time.Now()
	return now.Format("2006-01-02_15") + "_" + strconv.Itoa(videoCounter) // gives month-day_hour_videoCounter
}

func initNextFrameCounter(frameDir string, prefix string) int {
	counter := 0
	frameFile := filepath.Join(frameDir, prefix, getFrameFileName(counter))
	for {
		if _, err := os.Stat(frameFile); os.IsNotExist(err) {
			return counter
		}
		counter++
		frameFile = filepath.Join(frameDir, prefix, getFrameFileName(counter))
	}
}

func initNextVideoCounter(videosDir string) int {
	files, err := ioutil.ReadDir(videosDir)
	if err != nil {
		log.Printf("Failed to extract video counter: %v\n", err)
		return 0
	}
	now := time.Now()
	nowTimeStr := now.Format("2006-01-02_15")
	var filesInCurrentHour []os.FileInfo
	for _, f := range files {
		if strings.HasPrefix(f.Name(), nowTimeStr) {
			filesInCurrentHour = append(filesInCurrentHour, f)
		}
	}
	if len(filesInCurrentHour) == 0 {
		return 0
	}
	lastFileInCurrentHour := filesInCurrentHour[len(filesInCurrentHour)-1]
	t := strings.TrimPrefix(lastFileInCurrentHour.Name(), nowTimeStr)
	t = strings.TrimSuffix(t, ".mp4")
	counter, err := strconv.Atoi(t)
	if err != nil {
		log.Printf("Failed to extract video counter: %v\n", err)
		return 0
	}
	return counter
}

func SaveFrame(dir string, frameNumber int) {
	var imageData []byte
	for try := 0; try < frameFetchFailureThreshold; try++ {
		status, body, err := get(streamURL)
		if err != nil {
			log.Printf("Error (try=%d): Failed to save frame: %v\n", try, err)
		}
		if status > 399 {
			log.Printf("Error (try=%d): Got status code > 399, status=%d\n", try, status)
		}
		imageData = body
		break
	}

	frameFile := filepath.Join(dir, getFrameFileName(frameNumber))
	err := ioutil.WriteFile(frameFile, imageData, 0644)
	if err != nil {
		log.Fatalf("Error: Unable to write frame to file=%s\n", frameFile)
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
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	return resp.StatusCode, body, err
}
