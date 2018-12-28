package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
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
	fmt.Printf("SFE is %t\n", sfe)

	lanInterface, err := exec.Command("nvram", "get", "lan_ifname").Output()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Lan Interface is %q\n", lanInterface)

	conn, err := grpc.Dial("192.168.0.141:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	log.Println("Setting up Ip tables")
	setupIptable()
	client := NewReporterClient(conn)

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

		// Send full report
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		response, err := client.ReportData(ctx, &DataReport{
			Time: uint64(time.Now().Unix()),
			Nf:   iptableResult,
			Ct:   conntrackResult,
		})
		if err != nil {
			log.Println(err)
		} else {
			log.Printf("Response from server: %t", response.Success)
		}
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
