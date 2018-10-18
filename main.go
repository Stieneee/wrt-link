package main

import (
	"fmt"
	"runtime"
	"time"
)

type conn struct {
	spackets int
	sbytes   int
	dpackets int
	dbytes   int
}

func main() {
	go readStatsScheduler()
	for true {
		time.Sleep(time.Second)
	}
}

func readStatsScheduler() {
	for range time.Tick(time.Second) {
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
