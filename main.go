package main

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

var (
	// BuildVersion - populated during go build
	BuildVersion string
	// BuildTime - populated during go build
	BuildTime string
)

// os.Args[1] // API Address
// os.Args[2] // UID
// os.Args[3] // 256 byte RSA Private key path

type routerInfoReport struct {
}

type dataReport struct {
	T         uint64
	Nf        []Netfilter
	Ct        []Conntrack
	Hostnames []hostname
	WanIP     string
	Ping      pingStats
	Speed     speedStats
	Device    deviceStats
	MemUsage  memUsageStats
	AppUptime uint64
}

var signKey *rsa.PrivateKey

var startTime time.Time

var sfe bool
var lanInterface string
var wanInterface string
var wanIP string
var uptime string

func main() {
	startTime = time.Now()
	fmt.Printf("wrt-link %v %v \n", BuildVersion, BuildTime)

	setupHTTPClient()

	signBytes, err := ioutil.ReadFile(os.Args[3])
	if err != nil {
		log.Fatal(err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Fatal(err)
	}

	testAuthCreds()

	collectStartupInfo()

	log.Println("Setting up Ip tables")
	setupIptable()

	conntrackResultChan := make(chan []Conntrack, 2)
	requestConntrackChan := make(chan bool, 2)
	speedMonitorChan := make(chan speedStats, 2)
	requestSpeedMonitorChan := make(chan bool, 2)

	go readConntrackScheduler(conntrackResultChan, requestConntrackChan)
	go speedMonitor(speedMonitorChan, requestSpeedMonitorChan)
	go pingForInterval()
	go reporter()
	go upgradeChecker()

	for range time.Tick(time.Minute) {
		requestConntrackChan <- true
		requestSpeedMonitorChan <- true
		collectReport(
			conntrackResultChan,
			speedMonitorChan,
		)
	}
}

// testAuthCreds - Test creds before starting. ~10 mins to pass test before exit
func testAuthCreds() {
	for i := 0; i < 20; i++ {
		req, err := http.NewRequest("GET", fullURL("authCheck"), nil)
		if err != nil {
			log.Println(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+createToken())
		resp, err := client.Do(req)
		if err != nil {
			log.Println(err)
			time.Sleep(30 * time.Second)
			continue
		}
		if resp.StatusCode != 200 {
			s := []string{"API returned status code", resp.Status, fullURL("authCheck")}
			err = errors.New(strings.Join(s, " "))
			log.Println(err)
		}
		if resp.StatusCode == 200 {
			log.Println("Auth OK! Starting Logging.")
			return
		}

		defer resp.Body.Close()
		time.Sleep(30 * time.Second)
	}

	log.Fatal("Failed AuthCheck")
}

func collectStartupInfo() {

	out, err := exec.Command("nvram", "get", "DD_BOARD").Output()
	if err != nil {
		log.Fatal(err)
	}
	routerModel := string(strings.TrimSuffix(string(out), "\n"))
	log.Printf("Router Model is %s\n", routerModel)

	out, err = exec.Command("nvram", "get", "os_version").Output()
	if err != nil {
		log.Fatal(err)
	}
	osVersion := string(strings.TrimSuffix(string(out), "\n"))
	log.Printf("Firmware build %s\n", osVersion)

	out, err = exec.Command("nvram", "get", "sfe").Output()
	if err != nil {
		log.Fatal(err)
	}
	if out[0] == '1' {
		sfe = true
	}
	log.Printf("SFE is %t\n", sfe)

	out, err = exec.Command("nvram", "get", "lan_ifname").Output()
	if err != nil {
		log.Fatal(err)
	}
	lanInterface = string(strings.TrimSuffix(string(out), "\n"))
	log.Printf("Lan Interface is %s\n", lanInterface)

	out, err = exec.Command("nvram", "get", "wan_ifname").Output()
	if err != nil {
		log.Fatal(err)
	}
	wanInterface = string(strings.TrimSuffix(string(out), "\n"))
	log.Printf("Wan Interface is %s\n", wanInterface)

	out, err = exec.Command("uptime", "-s").Output()
	if err != nil {
		log.Fatal(err)
	}
	uptime = string(strings.TrimSuffix(string(out), "\n"))

	message := map[string]interface{}{
		"routerModel":  routerModel,
		"osVersion":    osVersion,
		"sfe":          sfe,
		"lanInterface": lanInterface,
		"uptime":       uptime,
	}
	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return
	}

	sendReport("POST", true, "startup", bytes)
}

func collectReport(
	conntrackResultChan <-chan []Conntrack,
	speedMonitorChan <-chan speedStats,
) {
	conntrackResult := <-conntrackResultChan
	speedStats := <-speedMonitorChan

	// Iptables
	iptableResult := readIptable(conntrackResult)
	setupIptable()

	tmpByteArr, err := exec.Command("nvram", "get", "wan_ipaddr").Output()
	if err != nil {
		log.Fatal(err)
		wanIP = "Error"
	} else {
		wanIP = strings.TrimSuffix(string(tmpByteArr), "\n")
	}

	// Create message
	message := dataReport{
		T:         uint64(time.Now().Unix()),
		Nf:        iptableResult,
		Ct:        conntrackResult,
		Hostnames: getHostnames(),
		WanIP:     wanIP,
		Ping:      getPingStats(),
		Speed:     speedStats,
		Device:    getDeviceStats(),
		MemUsage:  getMemUsage(),
		AppUptime: uint64(time.Since(startTime).Seconds()),
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return
	}

	sendReport("POST", true, "report", bytes)

	go pingForInterval()
}
