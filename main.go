package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"time"
)

var sfe bool
var lanInterface string

func main() {
	out, err := exec.Command("nvram", "get", "sfe").Output()
	if err != nil {
		log.Fatal(err)
	}
	if string(out) == "1" {
		sfe = true
	}
	fmt.Printf("SFE is %d\n", sfe)

	lanInterface, err := exec.Command("nvram", "get", "lan_ifname").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Lan Interface is %s\n", lanInterface)

	log.Println("Setting up Ip tables")
	setupIptable()

	go readConntrackScheduler()
	go readStatsScheduler()
	for true {
		time.Sleep(time.Minute)

	}
}

func readConntrackScheduler() {
	for range time.Tick(time.Second) {
		readConntrack("/proc/net/ip_conntrack")
	}
}

func readStatsScheduler() {
	for range time.Tick(time.Minute) {
		readConntrack("/proc/net/ip_conntrack")
	}
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
