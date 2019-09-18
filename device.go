package main

import (
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type loadStats struct {
	one     float64
	five    float64
	fifteen float64
	running uint64
	total   uint64
}

func getLoad() loadStats {
	var load loadStats
	data, err := ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		log.Println(err)
		load.one = 0
		load.five = 0
		load.fifteen = 0
		load.running = 0
		load.total = 0
		return load
	}

	strArr := strings.Split(string(data), " ")

	load.one, _ = strconv.ParseFloat(strArr[0], 64)
	load.five, _ = strconv.ParseFloat(strArr[1], 64)
	load.fifteen, _ = strconv.ParseFloat(strArr[2], 64)

	strArr2 := strings.Split(strArr[3], "/")

	load.running, _ = strconv.ParseUint(strArr2[0], 10, 64)
	load.total, _ = strconv.ParseUint(strArr2[1], 10, 64)

	return load
}
