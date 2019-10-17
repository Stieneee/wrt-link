package main

import (
	"log"
	"runtime"
	"syscall"
)

type deviceStats struct {
	Uptime    uint64
	Loads     [3]uint64
	Totalram  uint64
	Freeram   uint64
	Sharedram uint64
	Bufferram uint64
	Procs     uint16
}
type memUsageStats struct {
	Alloc      uint64
	TotalAlloc uint64
	Sys        uint64
	NumGC      uint32
}

func getDeviceStats() deviceStats {
	var in syscall.Sysinfo_t
	err := syscall.Sysinfo(&in)
	if err != nil {
		log.Println(err)
	}

	var res deviceStats

	res.Uptime = uint64(in.Uptime)
	res.Loads = [3]uint64{uint64(in.Loads[0]), uint64(in.Loads[1]), uint64(in.Loads[2])}
	res.Totalram = uint64(in.Totalram)
	res.Freeram = uint64(in.Freeram)
	res.Sharedram = uint64(in.Sharedram)
	res.Bufferram = uint64(in.Bufferram)
	res.Procs = in.Procs

	return res
}

func getMemUsage() memUsageStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	var res memUsageStats
	res.Alloc = m.Alloc
	res.TotalAlloc = m.TotalAlloc
	res.Sys = m.Sys
	res.NumGC = m.NumGC

	return res
}
