package dto

import "time"

type ComprehensivePodMetrics struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Status     string            `json:"status"`
	StartTime  string            `json:"startTime"`
	TimePoints []PodMetricsPoint `json:"timePoints"`
}

type PodMetricsPoint struct {
	Timestamp       time.Time `json:"timestamp"`
	CPUUsage        float64   `json:"cpuUsage"`        // cores
	CPURequests     float64   `json:"cpuRequests"`     // cores
	CPULimits       float64   `json:"cpuLimits"`       // cores
	MemoryUsage     float64   `json:"memoryUsage"`     // MiB
	MemoryRequests  float64   `json:"memoryRequests"`  // MiB
	MemoryLimits    float64   `json:"memoryLimits"`    // MiB
	NetworkRxBytes  float64   `json:"networkRxBytes"`  // bytes/sec
	NetworkTxBytes  float64   `json:"networkTxBytes"`  // bytes/sec
	NetworkRxErrors float64   `json:"networkRxErrors"` // errors/sec
	NetworkTxErrors float64   `json:"networkTxErrors"` // errors/sec
	RestartCount    int32     `json:"restartCount"`
}
