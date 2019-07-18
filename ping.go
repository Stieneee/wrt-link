package main

import (
	"log"
	"time"

	ping "github.com/sparrc/go-ping"
)

// Pinger - curent ping send receive handler
var Pinger *ping.Pinger

type pingStats struct {
	PacketsRecv int
	PacketsSent int
	MinRtt      time.Duration
	MaxRtt      time.Duration
	AvgRtt      time.Duration
	StdDevRtt   time.Duration
}

func pingForInterval() {
	var err error
	Pinger, err = ping.NewPinger("1.1.1.1")
	if err != nil {
		log.Println(err)
	}
	Pinger.SetPrivileged(true)
	Pinger.Count = 29
	Pinger.Interval, err = time.ParseDuration("2s")
	if err != nil {
		log.Println(err)
	}
	// Pinger.OnRecv = func(pkt *ping.Packet) {
	// 	fmt.Printf("%d bytes from %s: icmp_seq=%d time=%v\n",
	// 		pkt.Nbytes, pkt.IPAddr, pkt.Seq, pkt.Rtt)
	// }
	Pinger.Run()
}

func getPingStats() pingStats {
	var stats pingStats
	if Pinger != nil {
		liveStats := Pinger.Statistics()
		stats.PacketsRecv = liveStats.PacketsRecv
		stats.PacketsSent = liveStats.PacketsSent
		stats.MinRtt = liveStats.MinRtt / 1000
		stats.MaxRtt = liveStats.MaxRtt / 1000
		stats.AvgRtt = liveStats.AvgRtt / 1000
		stats.StdDevRtt = liveStats.StdDevRtt / 1000
	}
	return stats
}
