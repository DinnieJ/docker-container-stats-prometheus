package docker_test

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/DinnieJ/docker-container-stats-prometheus/pkg/docker"
	"github.com/docker/docker/api/types/container"
	"github.com/golang/mock/gomock"
)

const STATS_API_INTERVAL = 1 // or import it if it's defined elsewhere

func TestLoopFetchContainerStatInfo_CtxCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := docker.NewMockDockerClientInterface(ctrl)

	// Simulated stats response (empty JSON)
	statResponse := container.StatsResponse{}
	jsonData, _ := json.Marshal(statResponse)
	body := io.NopCloser(strings.NewReader(string(jsonData)))

	// Expect ContainerStats call
	mockClient.EXPECT().
		ContainerStats(gomock.Any(), "test-container", false).
		Return(container.StatsResponseReader{
			Body: body,
		}, nil).
		Times(1) // Should only be called once before ctx is cancelled

	ch := make(chan container.StatsResponse, 1)
	ctx, cancel := context.WithCancel(context.Background())
	appCtx := context.Background()

	// Trigger cancellation shortly after starting
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := docker.LoopFetchContainerStatInfo(mockClient, "test-container", ch, ctx, appCtx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	select {
	case <-ch:
		// success
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected a stats response to be sent to the channel")
	}
}
