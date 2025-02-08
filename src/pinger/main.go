package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"pinger/config"
	"pinger/ping"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/prometheus-community/pro-bing"
)

func doPing(ip string) (float64, error) {
	pinger, err := probing.NewPinger(ip)
	if err != nil {
		return -1, err
	}

	pinger.Count = 1
	pinger.Timeout = time.Second

	err = pinger.Run()
	if err != nil {
		return -1, err
	}

	stats := pinger.Statistics()
	return float64(stats.MaxRtt.Microseconds()), nil
}

func getContainers() ([]string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	containers, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}

	var ips []string
	for _, container := range containers {
		inspect, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			continue
		}
		if inspect.NetworkSettings != nil {
			for _, network := range inspect.NetworkSettings.Networks {
				ips = append(ips, network.IPAddress)
			}
		}
	}

	return ips, nil
}

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("No config file")
	}

	for {
		time.Sleep(time.Second * 5)

		ls, err := getContainers()
		if err != nil {
			log.Println("Failed to get containers")
			continue
		}

		res := []ping.PingEvent{}

		for _, ip := range ls {
			ms, err := doPing(ip)
			if err != nil {
				log.Printf("Couldn't ping '%v'\n", ip)
				continue
			}

			event := ping.PingEvent{
				Ip:        ip,
				Ping_ms:   ms,
				Pinged_at: time.Now(),
			}

			res = append(res, event)

		}

		resJosn, err := json.Marshal(res)
		if err != nil {
			panic("Should never happend")
		}

		if rs, err := http.Post(cfg.ApiUrl, "applicatin/json", bytes.NewReader(resJosn)); err != nil || rs.StatusCode != 200 {
			log.Printf("Couldn't post ping results\n %v\n%v", err, rs)
		}
	}
}
