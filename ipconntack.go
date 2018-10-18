package main

import (
	"bufio"
	"log"
	"os"
	"strconv"
	"time"
)

var m = make(map[string]conn)

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

			sport, _ := strconv.Atoi(sportS)
			dport, _ := strconv.Atoi(dportS)
			spackets, _ := strconv.Atoi(packets[0])
			dpackets, _ := strconv.Atoi(packets[1])
			sbytes, _ := strconv.Atoi(bytes[0])
			dbytes, _ := strconv.Atoi(bytes[1])

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

	file.Close()
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	elapsed := time.Since(start)
	log.Printf("Run took %s", elapsed)
	// PrintMemUsage()
}
