package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Conntrack - delivery structure of conntrack based information
type conntrack struct {
	proto      string
	src        string
	dst        string
	srcp       uint32
	dstp       uint32
	in         uint64
	out        uint64
	inPackets  uint32
	outPackets uint32
}

type conntrackLog struct {
	time            uint64
	proto           string
	src             string
	dst             string
	srcp            uint32
	dstp            uint32
	in              uint64
	out             uint64
	inPackets       uint32
	outPackets      uint32
	inDelta         uint64
	outDelta        uint64
	inPacketsDelta  uint32
	outPacketsDelta uint32
}

var m = make(map[string]conntrackLog)

func readConntrackScheduler(conntrackResultChan chan<- []conntrack, requestConntrackChan <-chan bool) {
	for range time.Tick(time.Second) {
		if len(requestConntrackChan) > 0 {
			log.Println("Conntrack report requested")
			_ = <-requestConntrackChan
			conntrackResultChan <- reportConntract()
		}
		readConntrack("/proc/net/ip_conntrack")
	}
}

func reportConntract() []conntrack {
	var connTrackResult []conntrack
	for _, value := range m {
		connTrackResult = append(connTrackResult, conntrack{
			proto:      value.proto,
			src:        value.src,
			dst:        value.dst,
			srcp:       value.srcp,
			dstp:       value.dstp,
			in:         value.inDelta,
			out:        value.outDelta,
			inPackets:  value.inPacketsDelta,
			outPackets: value.outPacketsDelta,
		})

		value.inDelta = 0
		value.outDelta = 0
		value.inPacketsDelta = 0
		value.outPacketsDelta = 0
	}

	return connTrackResult
}

func readConntrack(filename string) {
	// start := time.Now()
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := scanner.Text()
		cType := text[0:3]

		if cType == "tcp" || cType == "udp" {
			// log.Println(scanner.Text())
			src := ""
			dst := ""
			dportS := ""
			sportS := ""
			var packets [2]string
			packets[0] = ""
			packets[1] = ""
			var bytes [2]string
			bytes[0] = ""
			bytes[1] = ""

			for i := 0; i < len(text); i++ {
				switch text[i] {
				case 's':
					if src == "" && string(text[i:i+4]) == "src=" {
						j := i + 4
						for ; text[j] != ' '; j++ {
						}
						src = string(text[i+4 : j])
					}
					if sportS == "" && string(text[i:i+6]) == "sport=" {
						j := i + 6
						for ; text[j] != ' '; j++ {
						}
						sportS = string(text[i+6 : j])
					}
				case 'd':
					if dst == "" && string(text[i:i+4]) == "dst=" {
						j := i + 4
						for ; text[j] != ' '; j++ {
						}
						dst = string(text[i+4 : j])
					}
					if dportS == "" && string(text[i:i+6]) == "dport=" {
						j := i + 6
						for ; text[j] != ' '; j++ {
						}
						dportS = string(text[i+6 : j])
					}
				case 'p':
					if (packets[0] == "" || packets[1] == "") && string(text[i:i+8]) == "packets=" {
						j := i + 8
						for ; text[j] != ' '; j++ {
						}
						if packets[0] == "" {
							packets[0] = string(text[i+8 : j])
						} else {
							packets[1] = string(text[i+8 : j])
						}
					}
				case 'b':
					if (bytes[0] == "" || bytes[1] == "") && string(text[i:i+6]) == "bytes=" {
						j := i + 6
						for ; text[j] != ' '; j++ {
						}
						if bytes[0] == "" {
							bytes[0] = string(text[i+6 : j])
						} else {
							bytes[1] = string(text[i+6 : j])
						}
					}
				}
			}

			srcSlice := strings.Split(src, ".")
			dstSlice := strings.Split(dst, ".")

			if srcSlice[0] == dstSlice[0] && srcSlice[1] == dstSlice[1] && srcSlice[2] == dstSlice[2] {
				// same network don't care to log atm
				continue
			}

			tmp, _ := strconv.ParseUint(sportS, 10, 32)
			srcp := uint32(tmp)
			tmp, _ = strconv.ParseUint(dportS, 10, 32)
			dstp := uint32(tmp)
			tmp, _ = strconv.ParseUint(packets[0], 10, 32)
			srcPackets := uint32(tmp)
			tmp, _ = strconv.ParseUint(packets[1], 10, 32)
			dstPackets := uint32(tmp)
			in, _ := strconv.ParseUint(bytes[0], 10, 64)
			out, _ := strconv.ParseUint(bytes[1], 10, 64)

			// log.Printf("%s %s %s %d %d %d %d %d %d \n", cType, src, dst, sport, dport, spackets, dpackets, sbytes, dbytes)

			hash := src + dst + sportS + dportS
			c, ok := m[hash]
			if !ok {
				// log.Println("track new")
				m[hash] = conntrackLog{
					time:            uint64(time.Now().Unix()),
					proto:           cType,
					src:             src,
					dst:             dst,
					srcp:            srcp,
					dstp:            dstp,
					in:              in,
					out:             out,
					inPackets:       srcPackets,
					outPackets:      dstPackets,
					inDelta:         in,
					outDelta:        out,
					inPacketsDelta:  srcPackets,
					outPacketsDelta: dstPackets,
				}
				// log.Println("new")
			} else if c.inPackets > srcPackets || c.outPackets > dstPackets {
				// the connection seems to have less packets then previously seen the connection must have been reset
				log.Println("connection restart overwrite")
				c.time = uint64(time.Now().Unix())

				c.inDelta = c.inDelta + in
				c.outDelta = c.outDelta + out
				c.inPacketsDelta = c.inPacketsDelta + srcPackets
				c.outPacketsDelta = c.outPacketsDelta + dstPackets

				c.in = in
				c.out = out
				c.inPackets = srcPackets
				c.outPackets = dstPackets
				m[hash] = c
			} else if c.inPackets < srcPackets || c.outPackets < dstPackets {
				// new packets have arrived update the connection deltas and last seen states
				// log.Println("new packets update")
				c.time = uint64(time.Now().Unix())

				c.inDelta = c.inDelta + (in - c.in)
				c.outDelta = c.outDelta + (out - c.out)
				c.inPacketsDelta = c.inPacketsDelta + (srcPackets - c.inPackets)
				c.outPacketsDelta = c.outPacketsDelta + (dstPackets - c.outPackets)

				c.in = in
				c.out = out
				c.inPackets = srcPackets
				c.outPackets = dstPackets
				m[hash] = c
			} else {
				// log.Println("Update Time")
				// We saw this connection refresh the last seen time
				c.time = uint64(time.Now().Unix())
				m[hash] = c
			}
		}

	}

	file.Close()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	// elapsed := time.Since(start)
	// log.Printf("Run took %s", elapsed)
	// PrintMemUsage()
}
