package main

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

// These are expensive
var srcR = regexp.MustCompile("src=[0-9\\.]+")
var dstR = regexp.MustCompile("dst=[0-9\\.]+")
var sportR = regexp.MustCompile("sport=[0-9\\.]+")
var dportR = regexp.MustCompile("dport=[0-9\\.]+")
var packetsR = regexp.MustCompile("packets=[0-9\\.]+")
var bytesR = regexp.MustCompile("bytes=[0-9\\.]+")

func _firstParse() {
	file, err := os.Open(".samples/ip_conntrack")
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
			_ = regexp.MustCompile("src=[0-9\\.]+").FindString(text)[4:]
			_ = regexp.MustCompile("dst=[0-9\\.]+").FindString(text)[4:]
			_, _ = strconv.Atoi(strings.Replace(regexp.MustCompile("sport=[0-9\\.]+").FindString(text), "sport=", "", -1))
			_, _ = strconv.Atoi(strings.Replace(regexp.MustCompile("dport=[0-9\\.]+").FindString(text), "dport=", "", -1))

			packets := regexp.MustCompile("packets=[0-9\\.]+").FindAllString(text, 2)
			_, _ = strconv.Atoi(strings.Replace(packets[0], "packets=", "", -1))
			_, _ = strconv.Atoi(strings.Replace(packets[1], "packets=", "", -1))

			bytes := regexp.MustCompile("bytes=[0-9\\.]+").FindAllString(text, 2)
			_, _ = strconv.Atoi(strings.Replace(bytes[0], "bytes=", "", -1))
			_, _ = strconv.Atoi(strings.Replace(bytes[1], "bytes=", "", -1))
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func TestFirstParse(t *testing.T) {
	start := time.Now()

	for i := 0; i < 100; i++ {
		_firstParse()
	}

	elapsed := time.Since(start)
	log.Printf("First took %s", elapsed)
}

func _secondParse() {
	file, err := os.Open(".samples/ip_conntrack")
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
			_ = srcR.FindString(text)[4:]
			_ = dstR.FindString(text)[4:]
			_, _ = strconv.Atoi(strings.Replace(sportR.FindString(text), "sport=", "", -1))
			_, _ = strconv.Atoi(strings.Replace(dportR.FindString(text), "dport=", "", -1))

			packets := packetsR.FindAllString(text, 2)
			_, _ = strconv.Atoi(strings.Replace(packets[0], "packets=", "", -1))
			_, _ = strconv.Atoi(strings.Replace(packets[1], "packets=", "", -1))

			bytes := bytesR.FindAllString(text, 2)
			_, _ = strconv.Atoi(strings.Replace(bytes[0], "bytes=", "", -1))
			_, _ = strconv.Atoi(strings.Replace(bytes[1], "bytes=", "", -1))
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func TestSecondParse(t *testing.T) {
	start := time.Now()

	for i := 0; i < 100; i++ {
		_secondParse()
	}

	elapsed := time.Since(start)
	log.Printf("Second took %s", elapsed)
}

func _thirdParse() {
	file, err := os.Open(".samples/ip_conntrack")
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
			_ = srcR.FindString(text)[4:]
			_ = dstR.FindString(text)[4:]
			_, _ = strconv.Atoi(sportR.FindString(text)[6:])
			_, _ = strconv.Atoi(dportR.FindString(text)[6:])

			packets := packetsR.FindAllString(text, 2)
			_, _ = strconv.Atoi(packets[0][8:])
			_, _ = strconv.Atoi(packets[1][8:])

			bytes := bytesR.FindAllString(text, 2)
			_, _ = strconv.Atoi(bytes[0][6:])
			_, _ = strconv.Atoi(bytes[1][6:])
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func TestThirdParse(t *testing.T) {
	start := time.Now()

	for i := 0; i < 100; i++ {
		_thirdParse()
	}

	elapsed := time.Since(start)
	log.Printf("Third took %s", elapsed)
}

func _forthParse() {
	file, err := os.Open(".samples/ip_conntrack")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		text := []byte(scanner.Text())
		cType := string(text[0:3])

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

			_, _ = strconv.Atoi(sportS)
			_, _ = strconv.Atoi(dportS)

			_, _ = strconv.Atoi(packets[0])
			_, _ = strconv.Atoi(packets[1])

			_, _ = strconv.Atoi(bytes[0])
			_, _ = strconv.Atoi(bytes[1])
		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func TestFourParse(t *testing.T) {
	start := time.Now()

	for i := 0; i < 100; i++ {
		_forthParse()
	}

	elapsed := time.Since(start)
	log.Printf("Forth took %s", elapsed)
}
