package utils

import (
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const (
	machineIpRegex = `^192.168.1.\d{1,3}$`
)

func DirMustExist(dirPath string) {
	err := os.Mkdir(dirPath, 0755)
	if err != nil && !os.IsExist(err) {
		log.Fatalf("Error: Unable to create directory %s: %v\n", dirPath, err)
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
		log.Fatalf("Failed to get hostname! Output: %s; Error: %v\n", output, err)
	}
	return strings.TrimSpace(string(output))
}

func GetIpAddr() string {
	ip, exist := os.LookupEnv("IP_ADDR")
	if exist {
		return ip
	}
	ipRegex := regexp.MustCompile(machineIpRegex)
	ifaces, _ := net.Interfaces()
	// handle err
	for _, i := range ifaces {
		addrs, _ := i.Addrs()
		// handle err
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			// process IP address
			if ipRegex.MatchString(ip.String()) {
				return ip.String()
			}
		}
	}
	log.Fatalln("Unable to find this machine's ip address, the network might be down.")
	return ""
}

func IsPrimaryHost(hostname string) bool {
	if strings.Contains(hostname, GetHostName()) {
		return true
	}
	return false
}
