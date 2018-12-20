package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"time"
)

var m = make(map[string]Conntrack)

func reportConntract() []*Conntrack {
	var connTrackResult []*Conntrack
	for key, value := range m {
		connTrackResult = append(connTrackResult, &value)
		delete(m, key)
	}

	return connTrackResult
}

func readConntrack(filename string) {
	start := time.Now()
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
			// fmt.Println(scanner.Text())
			Src := ""
			Dst := ""
			DportS := ""
			SportS := ""
			var packets [2]string
			packets[0] = ""
			packets[1] = ""
			var bytes [2]string
			bytes[0] = ""
			bytes[1] = ""

			for i := 0; i < len(text); i++ {
				switch text[i] {
				case 's':
					if Src == "" && string(text[i:i+4]) == "Src=" {
						j := i + 4
						for ; text[j] != ' '; j++ {
						}
						Src = string(text[i+4 : j])
					}
					if SportS == "" && string(text[i:i+6]) == "Sport=" {
						j := i + 6
						for ; text[j] != ' '; j++ {
						}
						SportS = string(text[i+6 : j])
					}
				case 'd':
					if Dst == "" && string(text[i:i+4]) == "Dst=" {
						j := i + 4
						for ; text[j] != ' '; j++ {
						}
						Dst = string(text[i+4 : j])
					}
					if DportS == "" && string(text[i:i+6]) == "Dport=" {
						j := i + 6
						for ; text[j] != ' '; j++ {
						}
						DportS = string(text[i+6 : j])
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

			Sport, _ := strconv.Atoi(SportS)
			Dport, _ := strconv.Atoi(DportS)
			tmp, _ := strconv.ParseUint(packets[0], 10, 32)
			Srcp := uint32(tmp)
			tmp, _ = strconv.ParseUint(packets[1], 10, 32)
			Dstp := uint32(tmp)
			Out, _ := strconv.ParseUint(bytes[0], 10, 64)
			In, _ := strconv.ParseUint(bytes[1], 10, 64)

			// fmt.Printf("%s %s %s %d %d %d %d %d %d \n", cType, Src, Dst, Sport, Dport, Srcp, Dstp, Out, In)

			hash := Src + Dst + strconv.Itoa(Sport) + strconv.Itoa(Dport)
			c, ok := m[hash]
			if !ok {
				m[hash] = Conntrack{Srcp: Srcp, Dstp: Dstp, Out: Out, In: In}
				// fmt.Println("new")
			} else if c.Srcp > Srcp || c.Dstp > Dstp {
				// replace old
				c.Srcp = Srcp
				c.Dstp = Dstp
				c.Out = Out
				c.In = In
				// fmt.Println("replace")
			} else {
				// SrcpDelta := Srcp - c.Srcp
				// DstpDelta := Dstp - c.Dstp
				// sbytresDelta := Out - c.Out
				// inDelta := In - c.In

				c.Srcp = Srcp
				c.Dstp = Dstp
				c.Out = Out
				c.In = In
				// if SrcpDelta != 0 || DstpDelta != 0 || sbytresDelta != 0 || inDelta != 0 {
				// 	fmt.Println(SrcpDelta, DstpDelta, sbytresDelta, inDelta)
				// }
			}
		}
	}

	file.Close()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)
	log.Printf("Run took %s", elapsed)
	// PrintMemUsage()
}
