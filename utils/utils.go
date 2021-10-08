package utils

import (
	"log"
	"net"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

const (
	machineIpRegex = `^\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}$`
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
	ip_addrs := getValidIpAddrs()
	if len(ip_addrs) == 0 {
		log.Fatalln("Unable to find this machine's ip address, the network might be down.")
	}
	sort.Strings(ip_addrs)
	return ip_addrs[0]
}

func getValidIpAddrs() (ip_addrs []string) {
	ip, exist := os.LookupEnv("IP_ADDR")
	if exist {
		ip_addrs = append(ip_addrs, ip)
	} else {
		ipRegex := regexp.MustCompile(machineIpRegex)
		ifaces, err := net.Interfaces()
		if err != nil {
			log.Fatal("Unable to get the Network interfaces")
		}
		for _, i := range ifaces {
			addrs, err := i.Addrs()
			if err != nil {
				log.Fatalf("Unable to get the Network addresses from interface %v", i)
			}
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				ip_string := ip.String()
				// process IP address
				if ipRegex.MatchString(ip_string) {
					if ip_string != "127.0.0.1" {
						ip_addrs = append(ip_addrs, ip_string)
					}
				}
			}
		}
	}
	return ip_addrs
}

func IsPrimaryHost(target string) bool {
	if strings.Contains(target, GetHostName()) {
		return true
	}
	for _, ip_addr := range getValidIpAddrs() {
		if target == ip_addr {
			return true
		}
	}
	return false
}

func ReverseAny(s interface{}) {
	n := reflect.ValueOf(s).Len()
	swap := reflect.Swapper(s)
	for i, j := 0, n-1; i < j; i, j = i+1, j-1 {
		swap(i, j)
	}
}
