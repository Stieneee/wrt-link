package main

import (
	"encoding/json"
	"log"
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
	// TODO Do an args check
	// os.Args[1] // Address
	// os.Args[2] // UID
	// os.Args[3] // Secret

	// collectRouterInfo()

	log.Println("Setting up Ip tables")
	setupIptable()

	conntrackResultChan := make(chan []*Conntrack, 2)
	requestConntrackChan := make(chan bool, 2)

	go readConntrackScheduler(conntrackResultChan, requestConntrackChan)
	go reporter()

	for range time.Tick(time.Minute) {
		log.Println("Time to generate report")

		// Call the Conntrack thread to report current totals via channel.
		requestConntrackChan <- true

		// Iptables
		var iptableResult = readIptable()
		setupIptable()
		// log.Println("Report got iptables")

		// Grab results from other go routines
		var conntrackResult = <-conntrackResultChan

		// TODO Grad other stats

		message := &DataReport{
			Time: uint64(time.Now().Unix()),
			Nf:   iptableResult,
			Ct:   conntrackResult,
		}

		bytes, err := json.Marshal(message)
		if err != nil {
			log.Fatalln(err)
		}

		sendMessage("POST", true, "report", bytes)
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

func printMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// For info on each, see: https://golang.org/pkg/runtime/#MemStats
	log.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	log.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	log.Printf("\tSys = %v MiB", bToMb(m.Sys))
	log.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}
