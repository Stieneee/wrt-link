package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type conn struct {
	spackets int
	sbytes   int
	dpackets int
	dbytes   int
}

var m = make(map[string]conn)

// These are expensive
var srcR = regexp.MustCompile("src=[0-9\\.]+")
var dstR = regexp.MustCompile("dst=[0-9\\.]+")
var sportR = regexp.MustCompile("sport=[0-9\\.]+")
var dportR = regexp.MustCompile("dport=[0-9\\.]+")
var packetsR = regexp.MustCompile("packets=[0-9\\.]+")
var bytesR = regexp.MustCompile("bytes=[0-9\\.]+")

func main() {
	go ReadConnTrack()
	for true {
		time.Sleep(time.Second)
	}

}

func ReadConnTrack() {
	for range time.Tick(time.Second) {
		start := time.Now()
		file, err := os.Open("/proc/net/ip_conntrack")
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			var temp string
			text := scanner.Text()
			cType := text[0:3]

			if cType == "tcp" || cType == "udp" {
				// fmt.Println(scanner.Text())
				src := strings.Replace(srcR.FindString(text), "src=", "", -1)
				dst := strings.Replace(dstR.FindString(text), "dst=", "", -1)
				sport, _ := strconv.Atoi(strings.Replace(sportR.FindString(text), "sport=", "", -1))
				dport, _ := strconv.Atoi(strings.Replace(dportR.FindString(text), "dport=", "", -1))

				packets := packetsR.FindAllString(text, -1)
				temp = strings.Replace(packets[0], "packets=", "", -1)
				spackets, _ := strconv.Atoi(temp)
				temp = strings.Replace(packets[1], "packets=", "", -1)
				dpackets, _ := strconv.Atoi(temp)

				bytes := bytesR.FindAllString(text, -1)
				temp = strings.Replace(bytes[0], "bytes=", "", -1)
				sbytes, _ := strconv.Atoi(temp)
				temp = strings.Replace(bytes[1], "bytes=", "", -1)
				dbytes, _ := strconv.Atoi(temp)

				// fmt.Printf("%s %s %s %d %d %d %d %d %d \n", cType, src, dst, sport, dport, spackets, dpackets, sbytes, dbytes)

				hash := src + dst + strconv.Itoa(sport) + strconv.Itoa(dport)
				c, ok := m[hash]
				if !ok {
					m[hash] = conn{spackets: spackets, dpackets: dpackets, sbytes: sbytes, dbytes: dbytes}
					// fmt.Println("new")
				} else if c.spackets > spackets || c.dpackets > dpackets {
					// replace old
					c.spackets = spackets
					c.dpackets = dpackets
					c.sbytes = sbytes
					c.dbytes = dbytes
					// fmt.Println("replace")
				} else {
					// spacketsDelta := spackets - c.spackets
					// dpacketsDelta := dpackets - c.dpackets
					// sbytresDelta := sbytes - c.sbytes
					// dbytesDelta := dbytes - c.dbytes

					c.spackets = spackets
					c.dpackets = dpackets
					c.sbytes = sbytes
					c.dbytes = dbytes
					// if spacketsDelta != 0 || dpacketsDelta != 0 || sbytresDelta != 0 || dbytesDelta != 0 {
					// 	fmt.Println(spacketsDelta, dpacketsDelta, sbytresDelta, dbytesDelta)
					// }
				}
			}
		}

		if err := scanner.Err(); err != nil {
			log.Fatal(err)
		}
		elapsed := time.Since(start)
		log.Printf("Run took %s", elapsed)
		PrintMemUsage()
	}
}

func PrintMemUsage() {
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
