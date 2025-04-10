package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/docker/docker/api/types/container"
)

func ChannelFetchDockerContainers(ch chan []container.Summary, ctx context.Context) error {
	defer close(ch)
	for {
		fmt.Println("Scan docker containers is running")
		select {
		case <-ctx.Done():
			fmt.Println("Receive cancel signal")
			close(ch)
			return nil
		default:
			containers, err := GetAllContainers(dockerClient)
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				fmt.Println("Receive cancel signal")
				return nil
			case ch <- containers:
			}
			time.Sleep(CONTAINER_SCAN_INTERVAL * time.Second)
		}
	}
}

func ChannelWatchContainerStat(ch chan []container.Summary, chStat chan container.StatsResponse, rootCtx context.Context) {
	monitorData := make(map[string]struct {
		IsRunning  bool
		CancelFunc context.CancelFunc
	})
	for {
		select {
		case <-rootCtx.Done():
			fmt.Println("Receive root cancel signal, exiting application")
			return
		case containers := <-ch:
			receivedContainersID := make([]string, 0, len(containers))
			for _, k := range containers {
				receivedContainersID = append(receivedContainersID, k.ID)
			}
			for k := range monitorData {
				if slices.Contains(receivedContainersID, k) {
					continue
				} else {
					fmt.Println("Container removed", k)
					monitorData[k].CancelFunc()
					delete(monitorData, k)
				}
			}
			for _, k := range containers {
				if _, ok := monitorData[k.ID]; ok {
					continue
				}
				ctx, cancelFunc := context.WithCancel(rootCtx)
				monitorData[k.ID] = struct {
					IsRunning  bool
					CancelFunc context.CancelFunc
				}{
					IsRunning:  true,
					CancelFunc: cancelFunc,
				}
				fmt.Println("Container added", k.ID, k.Names)
				go LoopFetchContainerStatInfo(dockerClient, k.ID, chStat, ctx, rootCtx)
			}
		}
	}
}

// loopFetchContainerStatInfo is a function that will run in a goroutine
// Looping in interval to fetch container stat and push to prometheus client receiver
// The function will wait for STATS_API_INTERVAL seconds before fetching
// the stat again.
func LoopFetchContainerStatInfo(client DockerClientInterface, containerID string, ch chan container.StatsResponse, ctx context.Context, appCtx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Receive cancel signal, container removed")
			return nil
		case <-appCtx.Done():
			fmt.Println("Receive app cancel signal, exiting application")
			close(ch)
			return nil
		default:
			stat, err := client.ContainerStats(context.Background(), containerID, false)
			if err != nil {
				return err
			}
			defer stat.Body.Close()

			resp, err := io.ReadAll(stat.Body)
			if err != nil {
				return err
			}
			var statInfo container.StatsResponse
			err = json.Unmarshal(resp, &statInfo)
			if err != nil {
				return err
			}
			select {
			case <-ctx.Done():
				fmt.Println("Receive cancel signal")
				return nil
			case ch <- statInfo:
			}
			time.Sleep(STATS_API_INTERVAL * time.Second)
		}
	}
}
