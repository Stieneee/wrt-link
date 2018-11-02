package main

import (
	"log"
	"os/exec"
	"strings"
)

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

func readIptable() {

}
