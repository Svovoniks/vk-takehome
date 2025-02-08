package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"pinger/config"
	"pinger/docker"
	"pinger/netspace"
	"pinger/ping"
	"time"
)

func main() { cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("No config file")
	}

	for {
		time.Sleep(time.Second * 5)

		containers, err := docker.GetContainers()
		if err != nil {
			log.Println("Failed to get containers")
			continue
		}

		pingEvents := []ping.PingEvent{}

		for _, cn := range containers {
			ms, err := netspace.PingInNS(cn)
			if err != nil {
				log.Printf("Couldn't ping '%v'\n", cn.Ip)
				log.Println(err)
				continue
			}

			event := ping.PingEvent{
				Ip:        cn.Ip,
				Ping_ms:   ms,
				Pinged_at: time.Now(),
			}

			pingEvents = append(pingEvents, event)
		}

		pingJson, err := json.Marshal(pingEvents)
		if err != nil {
			panic("Marshal failed")
		}

		resp, err := http.Post(cfg.ApiUrl, "applicatin/json", bytes.NewReader(pingJson))
		if err != nil {
			log.Printf("Couldn't post ping results\n%v\n", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Server did not accept ping update\n%v\n", err)
		}

	}
}
