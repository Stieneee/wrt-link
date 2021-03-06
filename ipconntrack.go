package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

// Conntrack - delivery structure of Conntrack based information
type Conntrack struct {
	Proto          string
	Src            string
	Dst            string
	Srcp           uint32
	Dstp           uint32
	In             uint64
	Out            uint64
	InPackets      uint32
	OutPackets     uint32
	MaxActiveCount uint32
	MinActiveCount uint32
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
var maxCount uint32 = 0
var minCount uint32 = ^uint32(0)

func readConntrackScheduler(conntrackResultChan chan<- []Conntrack, requestConntrackChan <-chan bool) {
	for range time.Tick(time.Second) {
		if len(requestConntrackChan) > 0 {
			_ = <-requestConntrackChan
			conntrackResultChan <- reportConntract()
		}
		readConntrack("/proc/net/ip_conntrack")
	}
}

func reportConntract() []Conntrack {
	var connTrackResult []Conntrack
	expireTime := (uint64)(time.Now().Unix() - 5)
	for key, value := range m {
		if value.time < expireTime {
			delete(m, key)
		} else {
			if minCount == ^uint32(0) {
				minCount = 0
			}

			connTrackResult = append(connTrackResult, Conntrack{
				Proto:          value.proto,
				Src:            value.src,
				Dst:            value.dst,
				Srcp:           value.srcp,
				Dstp:           value.dstp,
				In:             value.inDelta,
				Out:            value.outDelta,
				InPackets:      value.inPacketsDelta,
				OutPackets:     value.outPacketsDelta,
				MaxActiveCount: maxCount,
				MinActiveCount: minCount,
			})

			value.inDelta = 0
			value.outDelta = 0
			value.inPacketsDelta = 0
			value.outPacketsDelta = 0

			m[key] = value
		}
	}

	maxCount = 0
	minCount = ^uint32(0)

	return connTrackResult
}

func readConntrack(filename string) {
	// start := time.Now()
	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var activeCount uint32 = 0

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		activeCount++
		text := scanner.Text()
		cType := text[0:3]

		if cType == "tcp" || cType == "udp" {
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
			out, _ := strconv.ParseUint(bytes[0], 10, 64) //first src is used, first bytes src -> dst is out
			in, _ := strconv.ParseUint(bytes[1], 10, 64)

			hash := src + dst + sportS + dportS
			c, ok := m[hash]
			if !ok {
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

	if activeCount > maxCount {
		maxCount = activeCount
	}

	if activeCount < minCount {
		minCount = activeCount
	}
}
