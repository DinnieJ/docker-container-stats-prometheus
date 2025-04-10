package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)

type DockerContainerStatInfo struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	CpuStats     *CpuStat      `json:"cpu_stats"`
	PreCpuStats  *CpuStat      `json:"precpu_stats"`
	BlockIOStats *BlockIOStats `json:"blkio_stats"`
	MemoryStats  *MemoryStat   `json:"memory_stats"`
}

type BlockIOStats struct {
	IoServiceBytesRecursive []struct {
		Major    int64  `json:"major"`
		Minor    int64  `json:"minor"`
		Value    int64  `json:"value"`
		Operator string `json:"op"`
	} `json:"io_service_bytes_recursive"`
}

type CpuStat struct {
	CpuUsage struct {
		TotalUsage        int64 `json:"total_usage"`
		UsageInKernelmode int64 `json:"usage_in_kernelmode"`
		UsageInUserMode   int64 `json:"usage_in_usermode"`
	} `json:"cpu_usage"`

	ThrottlingData struct {
		Periods          uint64 `json:"periods"`
		ThrottledPeriods uint64 `json:"throttled_periods"`
		ThrottledTime    int64  `json:"throttled_time"`
	} `json:"throttling_data"`
	OnlineCpus  int64 `json:"online_cpus"`
	SystemUsage int64 `json:"system_cpu_usage"`
}

type MemoryStat struct {
	Usage uint64 `json:"usage"`
	Stats struct {
		ActiveAnon uint64 `json:"active_anon"`
		ActiveFile uint64 `json:"active_file"`
		Anon       uint64 `json:"anon"`
		AnonThp    uint64 `json:"anon_thp"`
		File       uint64 `json:"file"`
		FileDirty  uint64 `json:"file_dirty"`
		FileMapped uint64 `json:"file_mapped"`
	}
	Limit uint64 `json:"limit"`
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

func GetAllContainers() ([]container.Summary, error) {
	return dockerClient.ContainerList(context.Background(), container.ListOptions{})
	// return containers, err
}

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
			containers, err := GetAllContainers()
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

func ChannelWatchContainerStat(ch chan []container.Summary, chStat chan DockerContainerStatInfo, rootCtx context.Context) {
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
				go ChannelDockerStatInfo(k.ID, chStat, ctx, rootCtx)
			}
		}
	}
}

func ChannelDockerStatInfo(containerID string, ch chan DockerContainerStatInfo, ctx context.Context, appCtx context.Context) error {
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
			stat, err := dockerClient.ContainerStats(context.Background(), containerID, false)
			if err != nil {
				return err
			}
			defer stat.Body.Close()

			resp, err := io.ReadAll(stat.Body)
			if err != nil {
				return err
			}
			var statInfo DockerContainerStatInfo
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
