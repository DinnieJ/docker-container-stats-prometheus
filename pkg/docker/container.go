package docker

import (
	"context"

	"github.com/docker/docker/api/types/container"
)

func GetAllContainers(client DockerClientInterface) ([]container.Summary, error) {
	return client.ContainerList(context.Background(), container.ListOptions{})
}
