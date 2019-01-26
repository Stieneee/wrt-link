package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"time"
)

type msgContainer struct {
	metheod  string
	auth     bool
	endPoint string
	json     []byte
}

var tr = &http.Transport{
	MaxIdleConns:    10,
	IdleConnTimeout: 30 * time.Second,
	// DisableCompression: true,
}

var client = &http.Client{Transport: tr}

var sendChan = make(chan msgContainer, 10000)

// Add the endpoint to url of the API
func fullURL(endPoint string) string {
	u, err := url.Parse(os.Args[1])
	if err != nil {
		log.Println(err)
	}
	u.Path = path.Join(u.Path, endPoint)
	s := u.String()
	return s
}

// Queue the message to be sent by reporter
func sendMessage(metheod string, auth bool, endPoint string, json []byte) {
	msg := msgContainer{
		metheod:  metheod,
		auth:     auth,
		endPoint: endPoint,
		json:     json,
	}
	sendChan <- msg
}

func attemptReport(msg msgContainer) bool {
	req, err := http.NewRequest(msg.metheod, fullURL(msg.endPoint), bytes.NewBuffer(msg.json))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(os.Args[2], os.Args[3])
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}
	if resp.StatusCode != 200 {
		log.Printf("API returned status code %d \n", resp.StatusCode)
		return false
	}
	// defer resp.Body.Close()
	var result map[string]interface{}

	json.NewDecoder(resp.Body).Decode(&result)
	log.Println(result)
	return true
}

func reporter() {
	var retryTime int32

	for {
		var msg = <-sendChan

		if attemptReport(msg) == false {
			sendChan <- msg
			if retryTime == 0 {
				retryTime = 5
			} else {
				// backoff capped at a 600 seconds
				retryTime = retryTime * 2
				if retryTime > 60 {
					retryTime = 60
				}
			}
			log.Printf("Message Send Failed Retrying in %d seconds\n", retryTime)
			time.Sleep(time.Duration(retryTime) * time.Second)
			// TODO could decay retry time on success to ease load back in.
		} else {
			retryTime = 0
			log.Println("msg sent")
		}
	}
}
