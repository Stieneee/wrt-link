package main

import (
	"encoding/json"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
)

type routerInfoReport struct {
}

type dataReport struct {
	Time uint64
	Nf   []*netfilter
	Ct   []*Conntrack
}

var sfe bool
var lanInterface string

func main() {
	ravenInit()

	// TODO Do an args check
	// os.Args[1] // Address
	// os.Args[2] // UID
	// os.Args[3] // Secret

	// testAuthCreds()
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

func testAuthCreds() {

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

	// message := map[string]interface{}{
	// 	"hello": "world",
	// 	"life":  42,
	// 	"embedded": map[string]string{
	// 		"yes": "of course!",
	// 	},
	// }
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
