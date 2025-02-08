package netspace

import (
	"fmt"
	"log"
	"os"
	"pinger/docker"
	"pinger/ping"
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"
)

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

func PingInNS(container *docker.Container) (float64, error) {
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

	ms, err := ping.DoPing(container.Ip)

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
