package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/certifi/gocertifi"
)

type msgContainer struct {
	method   string
	auth     bool
	endPoint string
	json     []byte
}

// TODO investigate other properties
var tr = &http.Transport{}

var client = &http.Client{Transport: tr}

var sendChan = make(chan msgContainer, 10000)

func setupHTTPClient() {
	rootCAs, err := gocertifi.CACerts()
	if err != nil {
		log.Println("failed to load root TLS certificates:", err)
	} else {
		client = &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyFromEnvironment,
				TLSClientConfig: &tls.Config{RootCAs: rootCAs},
			},
		}
	}

	// Trust the augmented cert pool in our client
	config := &tls.Config{
		RootCAs: rootCAs,
	}

	tr = &http.Transport{
		TLSClientConfig: config,
		MaxIdleConns:    10,
		IdleConnTimeout: 30 * time.Second,
	}

	client = &http.Client{Transport: tr}
}

// Add the endpoint to url of the API
func fullURL(endPoint string) string {
	u, err := url.Parse(os.Args[1])
	if err != nil {
		log.Println(err)
		return ""
	}
	u.Path = path.Join(u.Path, endPoint)
	s := u.String()
	return s
}

// Queue the report to be sent by reporter
func sendReport(method string, auth bool, endPoint string, json []byte) {
	msg := msgContainer{
		method:   method,
		auth:     auth,
		endPoint: endPoint,
		json:     json,
	}
	sendChan <- msg
}

func attemptReport(msg msgContainer) bool {
	req, err := http.NewRequest(msg.method, fullURL(msg.endPoint), bytes.NewBuffer(msg.json))
	if err != nil {
		log.Println(err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+createToken())
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return false
	}
	if resp.StatusCode != 200 {
		s := []string{"API returned status code", resp.Status, fullURL(msg.endPoint)}
		err = errors.New(strings.Join(s, " "))
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	// log.Println(result)

	if result["restart"] == true {
		log.Println("Restart Requested. Going down now....")
		exec.Command("reboot").Run()
	}
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
			// log.Println("msg sent")
		}
	}
}
