package utils

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func DirMustExist(dirPath string) {
	err := os.Mkdir(dirPath, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("Error: Unable to create directory %s: %v", dirPath, err)
	}
}

func GetHostName() string {
	host, exist := os.LookupEnv("HOSTNAME")
	if exist {
		return host
	}
	log.Println("HOSTNAME is not set in environment, fall back to use hostname command.")

	path, err := exec.LookPath("hostname")
	if err != nil {
		log.Fatal("Cannot find command hostname from PATH")
	}
	output, err := exec.Command(path).Output()
	if err != nil {
		log.Fatalf("Failed to get hostname! Output: %s; Error: %v", output, err)
	}
	return strings.TrimSpace(string(output))
}

func IsPrimaryHost(hostname string) bool {
	if strings.Contains(hostname, GetHostName()) {
		return true
	}
	return false
}
