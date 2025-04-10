package prometheus

import (
	"context"
	"os"

	"github.com/docker/docker/api/types/container"
	"github.com/prometheus/client_golang/prometheus"
)

const PREFIX_METRIC = "dsp_"

var mHostname string

func init() {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	mHostname = hostname
}

var (
	DockerCpuUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: PREFIX_METRIC + "docker_cpu_percent_usage",
			Help: "Docker cpu usage from host machine",
			ConstLabels: prometheus.Labels{
				"hostname": mHostname,
			},
		}, []string{"containerId", "containerName"},
	)

	DockerMemoryUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: PREFIX_METRIC + "docker_memory_percent_usage",
			Help: "Docker memory usage from host machine",
			ConstLabels: prometheus.Labels{
				"hostname": mHostname,
			},
		}, []string{"containerId", "containerName"},
	)

	DockerIOUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: PREFIX_METRIC + "docker_io_percent_usage",
			Help: "Docker IO usage from host machine",
			ConstLabels: prometheus.Labels{
				"hostname": mHostname,
			},
		}, []string{"containerId", "containerName", "type"},
	)
)

func init() {
	prometheus.MustRegister(DockerCpuUsage, DockerMemoryUsage, DockerIOUsage)
}

func getCpuUsageFromDockerStat(stat container.StatsResponse) float64 {
	// fmt.Println(stat)
	cpuDelta := stat.CPUStats.CPUUsage.TotalUsage - stat.PreCPUStats.CPUUsage.TotalUsage
	systemCpuDelta := stat.CPUStats.SystemUsage - stat.PreCPUStats.SystemUsage
	numberOfCpu := stat.CPUStats.OnlineCPUs
	if systemCpuDelta == 0 {
		return 0
	}
	return (float64(cpuDelta) / float64(systemCpuDelta)) * float64(numberOfCpu) * 100
}

func getMemoryUsageFromDockerStat(stat container.StatsResponse) float64 {
	return float64(stat.MemoryStats.Usage) / float64(stat.MemoryStats.Limit) * 100
}

func getIOUsageFromDockerStat(stat container.StatsResponse) (float64, float64) {
	var vRead float64
	var vWrite float64

	for _, v := range stat.BlkioStats.IoServiceBytesRecursive {
		if v.Op == "read" {
			vRead = float64(v.Value)
		} else if v.Op == "write" {
			vWrite = float64(v.Value)
		}
	}
	return vRead, vWrite
}
func BackgroundMetricHandler(ctx context.Context, ch chan container.StatsResponse) {
	for {
		select {
		case <-ctx.Done():
			return
		case stat := <-ch:
			if stat.ID == "" {
				continue
			}
			cpuUsagePercent := getCpuUsageFromDockerStat(stat)
			DockerCpuUsage.With(prometheus.Labels{
				"containerId":   stat.ID,
				"containerName": stat.Name,
			}).Set(cpuUsagePercent)
			memoryUsage := getMemoryUsageFromDockerStat(stat)
			DockerMemoryUsage.With(prometheus.Labels{
				"containerId":   stat.ID,
				"containerName": stat.Name,
			}).Set(memoryUsage)
			r, w := getIOUsageFromDockerStat(stat)
			DockerIOUsage.With(
				prometheus.Labels{
					"type":          "read",
					"containerId":   stat.ID,
					"containerName": stat.Name,
				},
			).Set(r)
			DockerIOUsage.With(
				prometheus.Labels{
					"type":          "write",
					"containerId":   stat.ID,
					"containerName": stat.Name,
				},
			).Set(w)

			// fmt.Println(cpuUsagePercent, memoryUsage, r, w)
		}
	}
}
