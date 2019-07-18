package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type speedStats struct {
	MaxRxR uint64
	MaxTxR uint64
}

func speedMonitor(speedMonitorChan chan<- speedStats, requestSpeedMonitorChan <-chan bool) {
	var maxRxR, maxTxR uint64 = 0, 0
	prevRx, prevTx := getByteValues()

	for range time.Tick(time.Second) {
		rx, tx := getByteValues()
		rxr := rx - prevRx
		txr := tx - prevTx
		if rxr > maxRxR {
			maxRxR = rxr
		}
		if txr > maxTxR {
			maxTxR = txr
		}
		prevRx, prevTx = rx, tx
		select {
		case _, ok := <-requestSpeedMonitorChan:
			if ok {
				report := speedStats{
					MaxRxR: maxRxR,
					MaxTxR: maxTxR,
				}
				speedMonitorChan <- report
				maxRxR, maxTxR = 0, 0
			} else {
				fmt.Println("requestSpeedMonitorChan closed!")
			}
		default:
		}
	}
}

func getByteValues() (uint64, uint64) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		items := strings.Fields(text)
		if items[0] == "vlan2:" {
			rx, _ := strconv.ParseUint(items[1], 10, 64)
			tx, _ := strconv.ParseUint(items[9], 10, 64)
			return rx, tx
		}
	}
	log.Println("Missing Vlan2 Interface")
	return 0, 0
}
