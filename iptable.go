package main

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// TODO investigate if we are making repeate entries in iptable
// TODO what the behvaiour of iptables if there are multiple entries. first match? should we be adding? are later entries replacing already read values.

type netfilter struct {
	mac string
	ip  string
	in  uint64
	out uint64
}

func setupIptable() {
	// Create tables (it doesn't matter if it already exists).
	err := exec.Command("iptables", "-N", "WRTLINK").Run()
	if err != nil {
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
					_, err = exec.Command("sh", "-c", "'iptables -nL WRTLINK | grep \""+feilds[0]+"\"' | echo $?").Output()
					if err != nil {
						log.Println("Adding Ip rules for " + feilds[0])
						_ = exec.Command("iptables", "-I", "WRTLINK", "-d", feilds[0], "-j", "RETURN").Run()
						_ = exec.Command("iptables", "-I", "WRTLINK", "-s", feilds[0], "-j", "RETURN").Run()
					}
				}
			}
		}
	} else {
		log.Println(err)
	}
}

func readArp() map[string]netfilter {

	var arpData = make(map[string]netfilter)

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
			feilds := strings.Fields(line)
			log.Println(feilds)
			if len(feilds) >= 6 {
				dev, ok := arpData[feilds[0]]
				if !ok {
					arpData[feilds[0]] = netfilter{
						ip:  feilds[0],
						mac: feilds[3],
						out: 0,
						in:  0,
					}
				} else {
					dev.ip = feilds[0]
					dev.mac = feilds[3]
					dev.out = 0
					dev.in = 0
				}
			}
		}
	}

	log.Println(arpData)

	return arpData
}

func readIptable(conntrackResult []conntrack) []netfilter {
	arpData := readArp()

	out, err := exec.Command("iptables", "-L", "WRTLINK", "-vnxZ").Output()
	if err != nil {
		log.Fatal(err)
	}
	iptableLines := strings.Split(string(out), "\n")

	for _, line := range iptableLines {
		fields := strings.Fields(line)
		if len(fields) == 9 && fields[0] != "pkts" {
			// log.Println(fields)
			if fields[0] != "0" && fields[1] != "0" {
				if fields[7] == "0.0.0.0/0" {
					// Download
					dev, ok := arpData[fields[8]]
					if !ok {
						log.Println("Error looking up device 7")
					} else {
						tmp, _ := strconv.ParseUint(fields[1], 10, 32)
						dev.in = dev.in + tmp
						arpData[fields[8]] = dev
					}

				} else if fields[8] == "0.0.0.0/0" {
					// Upload
					dev, ok := arpData[fields[7]]
					if !ok {
						log.Println("Error looking up device 8")
					} else {
						tmp, _ := strconv.ParseUint(fields[1], 10, 32)
						dev.out = dev.out + tmp
						arpData[fields[7]] = dev
					}
				} else {
					log.Println("iptable line missing 0.0.0.0/0")
				}
			}
		}
	}

	if sfe {
		for _, cr := range conntrackResult {
			dev, ok := arpData[cr.src]
			if !ok {
				dev, ok := arpData[cr.dst]
				if !ok {
					// log.Println("No match conntrack result in arp data", dev)
				} else {
					dev.in += cr.out
					dev.out += cr.in
					arpData[cr.dst] = dev
				}
			} else {
				dev.in += cr.in
				dev.out += cr.out
				arpData[cr.src] = dev
			}
		}
	}

	// Turn map into a array and return
	var iptableResult []netfilter
	for _, value := range arpData {
		log.Println("iptables result", value)
		vCopy := value
		iptableResult = append(iptableResult, vCopy)
	}

	return iptableResult
}
