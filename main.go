package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var config struct {
	token     string
	ip        string
	debug     bool
	wanDevice string
	seconds   int64
}

func updateSensorTX(speed int64) {
	updateSensor("sensor.wan_tx", speed)
}

func updateSensorRX(speed int64) {
	updateSensor("sensor.wan_rx", speed)
}

func updateSensor(sensor string, speed int64) {
	url := "http://" + config.ip + ":8123/api/states/" + sensor

	str := `{"state": "` + strconv.FormatInt(speed, 10) + `", "attributes": {"unit_of_measurement": "MBps"}}`

	var jsonStr = []byte(str)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("Authorization", "Bearer "+config.token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 1 * time.Second,
	}
	resp, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	//|| resp.StatusCode != http.StatusOK
	if config.debug {
		fmt.Println("response Status:", resp.Status)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
	}
}

func readTX() int64 {
	return read("/sys/class/net/" + config.wanDevice + "/statistics/tx_bytes")
}

func readRX() int64 {
	return read("/sys/class/net/" + config.wanDevice + "/statistics/rx_bytes")
}

func read(file string) int64 {
	trafficBytes, err := ioutil.ReadFile(file)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	trafficInt, err := strconv.ParseInt(strings.TrimSpace(string(trafficBytes)), 10, 64)
	if err != nil {
		fmt.Println(err)
		return -1
	}

	return trafficInt
}

func readTokenFromConfigFile() {
	tokenBytes, err := ioutil.ReadFile("/etc/config/ha-token")
	if err != nil {
		return
	}

	config.token = strings.TrimSpace(string(tokenBytes))
}

func run() {
	var (
		lastTX      int64 = -1
		lastRX      int64 = -1
		lastTXspeed int64 = -1
		lastRXspeed int64 = -1
	)

	for {
		start := time.Now()
		tx := readTX()

		if tx >= 0 && lastTX >= 0 {
			txSpeed := ((tx - lastTX) / 131072) / config.seconds

			if txSpeed != lastTXspeed {
				updateSensorTX(txSpeed)
				lastTXspeed = txSpeed
			}
		}

		rx := readRX()

		if rx >= 0 && lastRX >= 0 {
			rxSpeed := ((rx - lastRX) / 131072) / config.seconds

			if rxSpeed != lastRXspeed {
				updateSensorRX(rxSpeed)
				lastRXspeed = rxSpeed
			}
		}

		lastTX = tx
		lastRX = rx
		elapsed := time.Since(start)
		sleep := (time.Duration(config.seconds) * time.Second) - elapsed

		if sleep > 0 {
			time.Sleep(sleep)
		}
	}
}

func main() {
	flag.Int64Var(&config.seconds, "seconds", 2, "Update in seconds.")
	flag.StringVar(&config.token, "token", "", "Token to access HA.")
	flag.StringVar(&config.ip, "ha", "192.168.1.100", "HA ip address or hostname.")
	flag.BoolVar(&config.debug, "debug", false, "Debug output.")
	flag.StringVar(&config.wanDevice, "wan", "br-wan", "Your wan device.")
	flag.Parse()

	if config.token == "" {
		readTokenFromConfigFile()

		if config.token == "" {
			flag.PrintDefaults()
			fmt.Println("\nYou can also add token to /etc/config/ha-token file.")
			os.Exit(1)
		}
	}

	if config.seconds < 1 {
		flag.PrintDefaults()
		os.Exit(1)
	}

	run()
}
