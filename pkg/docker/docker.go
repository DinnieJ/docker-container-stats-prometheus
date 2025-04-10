package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

//go:generate mockgen -source=./docker.go -destination=./docker_mock.go -package=docker
type DockerClientInterface interface {
	ContainerStats(ctx context.Context, containerID string, stream bool) (container.StatsResponseReader, error)
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
}

var dockerClient *client.Client

const STATS_API_INTERVAL = 15
const CONTAINER_SCAN_INTERVAL = 30

func init() {
	c, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic("Failed to initialize docker client, maybe your system Docker is not running ?")
	}
	dockerClient = c
}
