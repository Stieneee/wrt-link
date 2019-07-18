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
	T     uint64
	Nf    []Netfilter
	Ct    []Conntrack
	WanIP string
	Ping  pingStats
	Speed speedStats
}

var signKey *rsa.PrivateKey

var sfe bool
var lanInterface string
var wanIP string

func main() {
	fmt.Printf("wrt-link %v %v \n", BuildVersion, BuildTime)

	setupHTTPClient()

	signBytes, err := ioutil.ReadFile(os.Args[3])
	if err != nil {
		log.Panicln(err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		log.Panicln(err)
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

	log.Fatalln("Failed AuthCheck")
}

func collectStartupInfo() {
	out, err := exec.Command("nvram", "get", "sfe").Output()
	if err != nil {
		log.Fatal(err)
	}
	if out[0] == '1' {
		sfe = true
	}
	log.Printf("SFE is %t\n", sfe)

	lanInterface, err := exec.Command("nvram", "get", "lan_ifname").Output()
	if err != nil {
		log.Fatal(err)
	}
	lanInterface = []byte(strings.TrimSuffix(string(lanInterface), "\n"))
	log.Printf("Lan Interface is %s\n", lanInterface)

	message := map[string]interface{}{
		"sfe":          sfe,
		"lanInterface": lanInterface,
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
		T:     uint64(time.Now().Unix()),
		Nf:    iptableResult,
		Ct:    conntrackResult,
		WanIP: wanIP,
		Ping:  getPingStats(),
		Speed: speedStats,
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return
	}

	sendReport("POST", true, "report", bytes)

	go pingForInterval()
}
