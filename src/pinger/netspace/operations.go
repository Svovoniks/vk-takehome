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
		return nil, fmt.Errorf("Failed to open current namespace: %v", err)
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
		return nil, fmt.Errorf("Failed to open container namespace: %v", err)
	}

	return &NetNS{
		Path: path,
		Fd:   fd,
	}, nil
}

func PingInNS(container *docker.Container) (float64, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	originalNS, err := getCurrentNS()
	if err != nil {
		return -1, err
	}
	defer unix.Close(originalNS.Fd)

	ns, err := getContainerNS(container.Pid)
	if err != nil {
		return -1, err
	}
	defer unix.Close(ns.Fd)

	if err := unix.Setns(ns.Fd, syscall.CLONE_NEWNET); err != nil {
		return -1, err
	}
	defer func() {
		if err := unix.Setns(originalNS.Fd, syscall.CLONE_NEWNET); err != nil {
			log.Printf("Failed to switch to the originalNS\n%v\n", err)
		}

	}()

	ms, err := ping.DoPing(container.Ip)
	if err != nil {
		return -1, err
	}

	return ms, nil
}
