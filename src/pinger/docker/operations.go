package docker

import (
	"context"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

func GetContainers() ([]*Container, error) {
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
