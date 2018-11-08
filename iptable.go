package main

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type iptableData struct {
	ip   string
	mac  string
	up   int64
	down int64
}

// TODO investigate if we are making repeate entries in iptable
// TODO what the behvaiour of iptables if there are multiple entries. first match? should we be adding? are later entries replacing already read values.

func setupIptable() {
	// Create tables (it doesn't matter if it already exists).
	err := exec.Command("iptables", "-N", "WRTLINK").Run()
	if err != nil {
		// log.Println(err)
	}

	// Add the WRTLINK CHAIN to the FORWARD chain (if non existing).
	err = exec.Command("sh", "-c", "'iptables -L FORWARD --line-numbers -n | grep \"WRTLINK\" | grep \"1\"'").Run()
	if err != nil {
		// if that command errors the chain is not where is should be
		log.Println("iptables chain out of place")

		err = exec.Command("sh", "-c", "'iptables -L FORWARD -n | grep \"WRTLINK\"'").Run()
		if err == nil {
			// the chain exsists but is in the wrong spot
			log.Println("iptables chain misplaced, recreating it...")
			// delete the chain
			err = exec.Command("sh", "-c", "'iptables -D FORWARD -j WRTLINK'").Run()
		}
		_ = exec.Command("sh", "-c", "'iptables -I FORWARD -j WRTLINK'").Run()
	}

	out, err := exec.Command("grep", lanInterface, "/proc/net/arp").Output()
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		if len(line) >= 1 {
			if line[0] == 'I' {
				continue
			}
			feilds := strings.Split(line, " ")
			if len(feilds) >= 1 {
				err = exec.Command("sh", "-c", "'iptables -nL WRTLINK | grep \""+feilds[0]+"\"'").Run()
				if err != nil {
					log.Println("Adding ip rules for " + feilds[0])
					_ = exec.Command("iptables", "-I", "WRTLINK", "-d", feilds[0], "-j", "RETURN").Run()
					_ = exec.Command("iptables", "-I", "WRTLINK", "-s", feilds[0], "-j", "RETURN").Run()
				}
			}
		}
	}
}

func readArp() map[string]iptableData {

	var arpData = make(map[string]iptableData)

	out, err := exec.Command("grep", "-v", "\"0x0\"", "/proc/net/arp").Output()
	if err != nil {
		log.Fatal(err)
	}
	arpLines := strings.Split(string(out), "\n")

	for _, line := range arpLines {
		if len(line) >= 1 {
			if line[0] == 'I' {
				continue
			}
			feilds := strings.Split(line, " ")
			for _, feild := range feilds {
				fmt.Println(feild)
			}
			if len(feilds) >= 6 {
				dev, ok := arpData[feilds[0]]
				if !ok {
					arpData[feilds[0]] = iptableData{
						ip:   feilds[0],
						mac:  feilds[3],
						up:   0,
						down: 0,
					}
				} else {
					dev.ip = feilds[0]
					dev.mac = feilds[3]
					dev.up = 0
					dev.down = 0
				}
			}
		}
	}

	return arpData
}

func readIptable() {
	arpData := readArp()

	out, err := exec.Command("iptables", "-L", "WRTLINK", "-vnxZ").Output()
	if err != nil {
		log.Fatal(err)
	}
	iptableLines := strings.Split(string(out), "\n")

	for _, line := range iptableLines {
		fields := strings.Fields(line)
		if len(fields) == 9 && fields[0] != "pkts" {
			// fmt.Println(fields)
			if fields[7] == "0.0.0.0/0" {
				// Download
				dev, ok := arpData[fields[8]]
				if !ok {
					// hmmmm device missing
				} else {
					dev.down, _ = strconv.ParseInt(fields[1], 10, 64)
					fmt.Println(dev)
				}

			} else if fields[8] == "0.0.0.0/0" {
				// Upload
				dev, ok := arpData[fields[7]]
				if !ok {
					// hmmmm device missing
				} else {
					dev.up, _ = strconv.ParseInt(fields[1], 10, 64)
				}
			} else {
				// TODO hmmmm ipv6?
			}
		}
	}
}
