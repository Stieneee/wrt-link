package main

import (
	"log"
	"os/exec"
	"strings"
)

type hostname struct {
	Mac      string
	Hostname string
}

func getHostnames() []hostname {
	var hostnames []hostname

	out, err := exec.Command("arp", "-a").Output()
	if err != nil {
		log.Fatal(err)
		return hostnames
	}

	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 3 {
			continue
		}
		hostnames = append(hostnames, hostname{
			Mac:      fields[3],
			Hostname: fields[0],
		})
	}

	return hostnames
}
