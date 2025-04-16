package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DinnieJ/docker-container-stats-prometheus/pkg/docker"
	"github.com/DinnieJ/docker-container-stats-prometheus/pkg/logger"
	"github.com/DinnieJ/docker-container-stats-prometheus/pkg/prometheus"
	"github.com/docker/docker/api/types/container"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM)
	logging := logger.GetLogger(&logger.LoggerConfig{
		Name:  "Application",
		Level: logger.TRACE,
	})
	// wg := sync.WaitGroup{}
	go func() {
		logging.Warn("Waiting for interrupt signal")
		<-sigChannel
		logging.Fatal("Killing process")
		cancel()
		os.Exit(0)
		// wg.Done()
	}()

	channelStats := make(chan container.StatsResponse, 1)
	channelContainers := make(chan []container.Summary, 1)

	// Workflow
	// Docker Container Scan -> Docker Container Stat  check -> Prom update
	go docker.ChannelFetchDockerContainers(channelContainers, rootCtx)
	go docker.ChannelWatchContainerStat(channelContainers, channelStats, rootCtx)
	go prometheus.BackgroundMetricHandler(rootCtx, channelStats)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
