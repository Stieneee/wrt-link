package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type DataReport struct {
	Time uint64
	Nf   []*Netfilter
	Ct   []*Conntrack
}

var sfe bool
var lanInterface string

func main() {
	// collectRouterInfo()

	log.Println("Setting up Ip tables")
	setupIptable()

	conntrackResultChan := make(chan []*Conntrack, 2)
	requestConntrackChan := make(chan bool, 2)

	go readConntrackScheduler(conntrackResultChan, requestConntrackChan)

	for range time.Tick(time.Minute) {
		log.Println("Time to Report")

		// Call the Conntrack thread to report current totals via channel.
		requestConntrackChan <- true

		// Iptables
		var iptableResult = readIptable()
		setupIptable()
		log.Println("Report got iptables")

		// Grad other stats

		// Grab results from other go routines
		var conntrackResult = <-conntrackResultChan
		log.Println("Report got conntrackResults")

		message := &DataReport{
			Time: uint64(time.Now().Unix()),
			Nf:   iptableResult,
			Ct:   conntrackResult,
		}

		bytesRepresentation, err := json.Marshal(message)
		if err != nil {
			log.Fatalln(err)
		}

		resp, err := http.Post("http://192.168.0.142:5000/logmyio-203720/us-central1/wrtLink/report", "application/json", bytes.NewBuffer(bytesRepresentation))
		if err != nil {
			log.Fatalln(err)
		}

		var result map[string]interface{}

		json.NewDecoder(resp.Body).Decode(&result)

		log.Println(result)
		log.Println(result["data"])
	}
}

func readConntrackScheduler(conntrackResultChan chan<- []*Conntrack, requestConntrackChan <-chan bool) {
	for range time.Tick(time.Second) {
		if len(requestConntrackChan) > 0 {
			log.Println("Conntrack report requested")
			_ = <-requestConntrackChan
			conntrackResultChan <- reportConntract()
		}
		readConntrack("/proc/net/ip_conntrack")
	}
}

func collectRouterInfo() {
	out, err := exec.Command("nvram", "get", "sfe").Output()
	if err != nil {
		log.Fatal(err)
	}
	if out[0] == '1' {
		sfe = true
	}
	fmt.Printf("SFE is %t\n", sfe)

	lanInterface, err := exec.Command("nvram", "get", "lan_ifname").Output()
	if err != nil {
		log.Fatal(err)
	}
	lanInterface = []byte(strings.TrimSuffix(string(lanInterface), "\n"))
	fmt.Printf("Lan Interface is %s\n", lanInterface)
}

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
