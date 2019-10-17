package main

import (
	"log"
	"os/exec"
	"strconv"
	"strings"
)

// TODO investigate if we are making repeate entries in iptable
// TODO what the behvaiour of iptables if there are multiple entries. first match? should we be adding? are later entries replacing already read values.

// Netfilter - hold device ip mac and bandwidth
type Netfilter struct {
	Mac string
	IP  string
	In  uint64
	Out uint64
}

func setupIptable() {
	// Create tables (it doesn't matter if it already exists).
	err := exec.Command("iptables", "-N", "WRTLINK").Run()
	if err != nil {
		// Add the WRTLINK CHAIN to the FORWARD chain (if non existing).
		err = exec.Command("/bin/sh", "-c", "iptables -L FORWARD --line-numbers -n | grep WRTLINK | grep 1").Run()
		if err != nil {
			// if that command errors the chain is not where is should be
			log.Println(err)
			log.Println("iptables chain out of place")

			err = exec.Command("/bin/sh", "-c", "iptables -L FORWARD -n | grep WRTLINK").Run()
			if err == nil {
				// the chain exists but is in the wrong spot
				log.Println("iptables chain misplaced, deleting it...")
				// delete the chain
				//TODO This only deletes one
				err = exec.Command("iptables", "-D", "FORWARD", "-j", "WRTLINK").Run()
			}
			log.Println("creating iptables chain")
			_ = exec.Command("iptables", "-I", "FORWARD", "-j", "WRTLINK").Run()
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
				fields := strings.Split(line, " ")
				if len(fields) >= 1 {
					_, err = exec.Command("sh", "-c", "'iptables -nL WRTLINK | grep \""+fields[0]+"\"' | echo $?").Output()
					if err != nil {
						log.Println("Adding Ip rules for " + fields[0])
						_ = exec.Command("iptables", "-I", "WRTLINK", "-d", fields[0], "-j", "RETURN").Run()
						_ = exec.Command("iptables", "-I", "WRTLINK", "-s", fields[0], "-j", "RETURN").Run()
					}
				}
			}
		}
	} else {
		log.Println(err)
	}
}

func readArp() map[string]Netfilter {

	var arpData = make(map[string]Netfilter)

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
			fields := strings.Fields(line)
			if len(fields) >= 6 {
				dev, ok := arpData[fields[0]]
				if !ok {
					arpData[fields[0]] = Netfilter{
						IP:  fields[0],
						Mac: fields[3],
						Out: 0,
						In:  0,
					}
				} else {
					dev.IP = fields[0]
					dev.Mac = fields[3]
					dev.Out = 0
					dev.In = 0
				}
			}
		}
	}
	return arpData
}

func readIptable(conntrackResult []Conntrack) []Netfilter {
	arpData := readArp()

	out, err := exec.Command("iptables", "-L", "WRTLINK", "-vnxZ").Output()
	if err != nil {
		log.Fatal(err)
	}
	iptableLines := strings.Split(string(out), "\n")

	for _, line := range iptableLines {
		fields := strings.Fields(line)
		if len(fields) == 9 && fields[0] != "pkts" {
			if fields[0] != "0" && fields[1] != "0" {
				if fields[7] == "0.0.0.0/0" {
					// Download
					dev, ok := arpData[fields[8]]
					if !ok {
						log.Println("Error looking up device 7")
					} else {
						tmp, _ := strconv.ParseUint(fields[1], 10, 32)
						dev.In = dev.In + tmp
						arpData[fields[8]] = dev
					}

				} else if fields[8] == "0.0.0.0/0" {
					// Upload
					dev, ok := arpData[fields[7]]
					if !ok {
						log.Println("Error looking up device 8")
					} else {
						tmp, _ := strconv.ParseUint(fields[1], 10, 32)
						dev.Out = dev.Out + tmp
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
			dev, ok := arpData[cr.Src]
			if !ok {
				dev, ok := arpData[cr.Dst]
				if !ok {
					// log.Println("No match conntrack result in arp data", dev)
				} else {
					dev.In += cr.Out
					dev.Out += cr.In
					arpData[cr.Dst] = dev
				}
			} else {
				dev.In += cr.In
				dev.Out += cr.Out
				arpData[cr.Src] = dev
			}
		}
	}

	// Turn map into a array and return
	var iptableResult []Netfilter
	for _, value := range arpData {
		vCopy := value
		iptableResult = append(iptableResult, vCopy)
	}

	return iptableResult
}
