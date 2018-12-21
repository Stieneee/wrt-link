package main

import (
	// "fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

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
					log.Println("Adding Ip rules for " + feilds[0])
					_ = exec.Command("iptables", "-I", "WRTLINK", "-d", feilds[0], "-j", "RETURN").Run()
					_ = exec.Command("iptables", "-I", "WRTLINK", "-s", feilds[0], "-j", "RETURN").Run()
				}
			}
		}
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
			feilds := strings.Split(line, " ")
			// for _, feild := range feilds {
			// 	// fmt.Println(feild)
			// }
			if len(feilds) >= 6 {
				dev, ok := arpData[feilds[0]]
				if !ok {
					arpData[feilds[0]] = Netfilter{
						Ip:  feilds[0],
						Mac: feilds[3],
						Out: 0,
						In:  0,
					}
				} else {
					dev.Ip = feilds[0]
					dev.Mac = feilds[3]
					dev.Out = 0
					dev.In = 0
				}
			}
		}
	}

	return arpData
}

func readIptable() []*Netfilter {
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
					dev.In, _ = strconv.ParseUint(fields[1], 10, 32)
					// fmt.Println(dev)
				}

			} else if fields[8] == "0.0.0.0/0" {
				// Upload
				dev, ok := arpData[fields[7]]
				if !ok {
					// hmmmm device missing
				} else {
					dev.Out, _ = strconv.ParseUint(fields[1], 10, 32)
				}
			} else {
				// TODO hmmmm ipv6?
			}
		}
	}

	// Turn map into a array and return
	var iptableResult []*Netfilter
	for _, value := range arpData {
		iptableResult = append(iptableResult, &value)
	}

	return iptableResult
}
