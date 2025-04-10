package docker_test

import (
	"errors"
	"testing"

	"github.com/DinnieJ/docker-container-stats-prometheus/pkg/docker"
	"github.com/docker/docker/api/types/container"
	gomock "github.com/golang/mock/gomock"
)

func TestGetAllContainers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := docker.NewMockDockerClientInterface(ctrl)

	expected := []container.Summary{
		{ID: "c1"},
		{ID: "c2"},
	}

	// Set up expectation
	mockClient.EXPECT().
		ContainerList(gomock.Any(), container.ListOptions{}).
		Return(expected, nil)

	containers, err := docker.GetAllContainers(mockClient)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(containers) != 2 {
		t.Errorf("expected 2 containers, got %d", len(containers))
	}
	if containers[0].ID != "c1" {
		t.Errorf("unexpected first container ID: %s", containers[0].ID)
	}
}

func TestGetAllContainers_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := docker.NewMockDockerClientInterface(ctrl)

	mockClient.EXPECT().
		ContainerList(gomock.Any(), container.ListOptions{}).
		Return(nil, errors.New("docker failure"))

	containers, err := docker.GetAllContainers(mockClient)
	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if containers != nil {
		t.Errorf("expected nil containers, got %v", containers)
	}
}
