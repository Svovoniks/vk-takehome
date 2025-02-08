package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"pinger/config"
	"pinger/ping"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/prometheus-community/pro-bing"
	"golang.org/x/sys/unix"
)

type NetNS struct {
	Path string
	Fd   int
}

type Container struct {
	Ip  string
	Pid string
}

func getCurrentNS() (*NetNS, error) {
	nsPath := fmt.Sprintf("/proc/%v/ns/net", os.Getpid())
	fd, err := unix.Open(nsPath, unix.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open current namespace: %v", err)
	}

	return &NetNS{
		Path: nsPath,
		Fd:   fd,
	}, nil
}
func getContainerNS(sandboxKey string) (*NetNS, error) {
	path := fmt.Sprintf("/docker/proc/%v/ns/net", sandboxKey)
	fd, err := unix.Open(path, unix.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open container namespace: %v", err)
	}

	log.Println(path)

	return &NetNS{
		Path: path,
		Fd:   fd,
	}, nil
}

func setNS(ns *NetNS) error {
	runtime.LockOSThread()
	if err := unix.Setns(ns.Fd, syscall.CLONE_NEWNET); err != nil {
		runtime.UnlockOSThread()
		log.Println(ns.Fd)
		return fmt.Errorf("failed to switch namespace: %v", err)
	}
	return nil
}

func pingInNS(container *Container) (float64, error) {
	originalNS, err := getCurrentNS()
	if err != nil {
		return -1, err
	}
	defer unix.Close(originalNS.Fd)

	ns, err := getContainerNS(container.Pid)
	if err != nil {
		return -1, err
	}

	if err := setNS(ns); err != nil {
		return -1, err
	}

	ms, err := doPing(container.Ip)

	if err != nil {
		if nsErr := setNS(originalNS); nsErr != nil {
			err = fmt.Errorf("Error 1: %v\nError 2: %v", err, nsErr)
		}
		return -1, err
	}

	if err := setNS(originalNS); err != nil {
		return -1, err
	}

	return ms, nil
}

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
	return float64(stats.MaxRtt.Microseconds()) / 1000, nil
}

func getContainers() ([]*Container, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}
	defer cli.Close()

	containerData, err := cli.ContainerList(context.Background(), container.ListOptions{})
	if err != nil {
		return nil, err
	}

	var containers []*Container
	for _, container := range containerData {
		inspect, err := cli.ContainerInspect(context.Background(), container.ID)
		if err != nil {
			continue
		}
		if inspect.NetworkSettings != nil {
			for _, network := range inspect.NetworkSettings.Networks {
				containers = append(containers,
					&Container{
						Ip:  network.IPAddress,
						Pid: strconv.Itoa(inspect.State.Pid),
					})
			}
		}
	}

	return containers, nil
}

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("No config file")
	}

	for {
		time.Sleep(time.Second * 5)

		containers, err := getContainers()
		if err != nil {
			log.Println("Failed to get containers")
			continue
		}

		pingEvents := []ping.PingEvent{}

		for _, cn := range containers {
			ms, err := pingInNS(cn)
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
			log.Printf("Couldn't post ping results\n%v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("Server did not accept ping update\n%v", err)
		}

	}
}
