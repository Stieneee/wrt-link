package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func upgradeChecker() {
	for range time.Tick(time.Minute) {
		queryUpgrade()
	}
}

func queryUpgrade() bool {
	req, err := http.NewRequest("GET", fullURL("queryUpgrade"), nil)
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
		s := []string{"API returned status code", resp.Status, "upgradeQuery"}
		err = errors.New(strings.Join(s, " "))
		log.Println(err)
		return false
	}
	defer resp.Body.Close()
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)

	if result["requested"] == true {
		// TODO validate
		log.Println("Upgrade requested!")
		log.Println(result)
		doUpgrade(result["url"].(string), result["hash"].(string), result["attemptID"].(string))
		return true
	}
	return false
}

func doUpgrade(url string, hash string, attemptID string) {
	upgradeReport("log", "starting upgrade", attemptID)

	// dDownload File ///////////////
	err := downloadFile("/tmp/wrtlink-upgrade.bin", url)
	if err != nil {
		upgradeReport("fail", "download failure", attemptID)
		log.Fatal(err)
		return
	}
	upgradeReport("log", "download complete", attemptID)

	// Check Hash /////////////////
	out, err := exec.Command("md5sum", "/tmp/wrtlink-upgrade.bin").Output()
	if err != nil {
		upgradeReport("fail", "download error", attemptID)
		log.Fatal(err)
		return
	}
	md5Output := string(strings.TrimSuffix(string(out), "\n"))
	fields := strings.Fields(md5Output)

	if len(fields) < 2 {
		// MD5 error
		upgradeReport("fail", "error parsing md5Output", attemptID)
		log.Fatal("error parsing md5Output")
		return
	}
	if !strings.Contains(fields[0], hash) {
		upgradeReport("fail", "hash dose not match", attemptID)
		log.Fatal("hash dose not match")
		return
	}
	upgradeReport("log", "hash good", attemptID)

	// Do upgrade/////////////////
	out, err = exec.Command("write", "/tmp/wrtlink-upgrade.bin", "linux").Output()
	if err != nil {
		upgradeReport("fail", "write error", attemptID)
		log.Fatal(err)
		return
	}
	upgradeReport("success", string(out), attemptID)

	// Reboot ////////////////////
	exec.Command("reboot").Run()
}

// Do not queue these messages. Attempt to send immediately and block progress

func upgradeReport(level string, msg string, attemptID string) {
	message := map[string]interface{}{
		"level":     level,
		"attemptID": attemptID,
		"msg":       msg,
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		log.Println(err)
		return
	}

	m := msgContainer{
		method:   "POST",
		auth:     true,
		endPoint: "upgradeLog",
		json:     bytes,
	}
	attemptReport(m)
}

func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}
