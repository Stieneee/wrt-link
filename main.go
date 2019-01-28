package main

import (
	"crypto/rsa"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/getsentry/raven-go"
)

// os.Args[1] // API Address
// os.Args[2] // UID
// os.Args[3] // 256 byte RSA Private key path

type routerInfoReport struct {
}

type dataReport struct {
	Time uint64
	Nf   []*netfilter
	Ct   []*Conntrack
}

var signKey *rsa.PrivateKey

var sfe bool
var lanInterface string

func main() {
	ravenInit()

	signBytes, err := ioutil.ReadFile(os.Args[3])
	if err != nil {
		raven.CaptureErrorAndWait(err, ravenContext)
		log.Panicln(err)
	}
	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		raven.CaptureErrorAndWait(err, ravenContext)
		log.Panicln(err)
	}

	testAuthCreds()

	// collectStartupInfo()

	log.Println("Setting up Ip tables")
	setupIptable()

	conntrackResultChan := make(chan []*Conntrack, 2)
	requestConntrackChan := make(chan bool, 2)

	go readConntrackScheduler(conntrackResultChan, requestConntrackChan)
	go reporter()

	for range time.Tick(time.Minute) {
		collectReport(conntrackResultChan, requestConntrackChan)
	}
}

// testAuthCreds - Test credientals before starting. ~10 mins to pass test before exit
func testAuthCreds() {
	for i := 0; i < 20; i++ {
		req, err := http.NewRequest("GET", fullURL("authCheck"), nil)
		if err != nil {
			// raven.CaptureError(err, ravenContext)
			log.Println(err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+createToken())
		resp, err := client.Do(req)
		if err != nil {
			// raven.CaptureError(err, ravenContext)
			log.Println(err)
		}
		if resp.StatusCode != 200 {
			s := []string{"API returned status code", resp.Status, fullURL("authCheck")}
			err = errors.New(strings.Join(s, " "))
			// raven.CaptureError(err, ravenContext)
			log.Println(err)
		}
		if resp.StatusCode == 200 {
			log.Println("Auth OK! Starting Logging.")
			return
		}

		defer resp.Body.Close()
		time.Sleep(30 * time.Second)
	}

	raven.CaptureMessageAndWait("Failed AuthCheck", ravenContext)
	log.Fatalln("Failed AuithCheck")
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
		raven.CaptureError(err, ravenContext)
		log.Println(err)
		return
	}

	sendReport("POST", true, "startup", bytes)
}

func collectReport(conntrackResultChan <-chan []*Conntrack, requestConntrackChan chan<- bool) {
	log.Println("Time to generate report")

	// Call the Conntrack thread to report current totals via channel.
	requestConntrackChan <- true

	// Iptables
	var iptableResult = readIptable()
	setupIptable()

	// Grab results from other go routines
	var conntrackResult = <-conntrackResultChan

	//	TODO Grad other stats

	// Create message
	message := &dataReport{
		Time: uint64(time.Now().Unix()),
		Nf:   iptableResult,
		Ct:   conntrackResult,
	}
	bytes, err := json.Marshal(message)
	if err != nil {
		raven.CaptureError(err, ravenContext)
		log.Println(err)
		return
	}

	sendReport("POST", true, "report", bytes)
}
