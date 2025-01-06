package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type MetricsController struct {
	client        client.Client
	metricsClient *metrics.Clientset
	scheme        *runtime.Scheme
	promURL       string
}

type HistoricalMetrics struct {
	PodMetrics []PodMetricTimeSeries `json:"podMetrics"`
	TimeRange  string                `json:"timeRange"`
	StartTime  time.Time             `json:"startTime"`
	EndTime    time.Time             `json:"endTime"`
}

type PodMetricTimeSeries struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Metrics   []MetricDataPoint `json:"metrics"`
}

type MetricDataPoint struct {
	Timestamp     time.Time `json:"timestamp"`
	CPUUsage      float64   `json:"cpuUsage"`      // cores
	MemoryUsage   float64   `json:"memoryUsage"`   // MiB
	NetworkRxRate float64   `json:"networkRxRate"` // bytes/sec
	NetworkTxRate float64   `json:"networkTxRate"` // bytes/sec
}

func NewMetricsController(client client.Client, scheme *runtime.Scheme, config *rest.Config) (*MetricsController, error) {
	metricsClient, err := metrics.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics client: %v", err)
	}

	// Get Prometheus URL from config or environment
	promURL := os.Getenv("PROMETHEUS_URL")
	if promURL == "" {
		promURL = "http://prometheus-server:9090" // default
	}

	return &MetricsController{
		client:        client,
		metricsClient: metricsClient,
		scheme:        scheme,
		promURL:       promURL,
	}, nil
}

func (c *MetricsController) GetHistoricalMetrics(ctx context.Context, timeRange string) (*HistoricalMetrics, error) {
	// Calculate time range
	endTime := time.Now()
	var startTime time.Time
	var step time.Duration

	switch timeRange {
	case "1d":
		startTime = endTime.Add(-24 * time.Hour)
		step = 5 * time.Minute
	case "3d":
		startTime = endTime.Add(-72 * time.Hour)
		step = 15 * time.Minute
	case "1w":
		startTime = endTime.Add(-168 * time.Hour)
		step = 1 * time.Hour
	default:
		return nil, fmt.Errorf("invalid time range: %s", timeRange)
	}

	// Get list of pods
	var pods corev1.PodList
	if err := c.client.List(ctx, &pods); err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	var podMetrics []PodMetricTimeSeries

	for _, pod := range pods.Items {
		// Query Prometheus for each metric
		cpuQuery := fmt.Sprintf("rate(container_cpu_usage_seconds_total{pod=\"%s\", namespace=\"%s\"}[5m])",
			pod.Name, pod.Namespace)
		memQuery := fmt.Sprintf("container_memory_working_set_bytes{pod=\"%s\", namespace=\"%s\"}",
			pod.Name, pod.Namespace)
		rxQuery := fmt.Sprintf("rate(container_network_receive_bytes_total{pod=\"%s\", namespace=\"%s\"}[5m])",
			pod.Name, pod.Namespace)
		txQuery := fmt.Sprintf("rate(container_network_transmit_bytes_total{pod=\"%s\", namespace=\"%s\"}[5m])",
			pod.Name, pod.Namespace)

		cpuData, err := c.QueryPrometheus(cpuQuery, startTime, endTime, step)
		if err != nil {
			return nil, err
		}

		memData, err := c.QueryPrometheus(memQuery, startTime, endTime, step)
		if err != nil {
			return nil, err
		}

		rxData, err := c.QueryPrometheus(rxQuery, startTime, endTime, step)
		if err != nil {
			return nil, err
		}

		txData, err := c.QueryPrometheus(txQuery, startTime, endTime, step)
		if err != nil {
			return nil, err
		}

		// Combine metrics into time series
		var metrics []MetricDataPoint
		if len(cpuData.Data.Result) > 0 && len(memData.Data.Result) > 0 &&
			len(rxData.Data.Result) > 0 && len(txData.Data.Result) > 0 {

			for i, values := range cpuData.Data.Result[0].Values {
				timestamp := int64(values[0].(float64))

				metrics = append(metrics, MetricDataPoint{
					Timestamp:     time.Unix(timestamp, 0),
					CPUUsage:      values[1].(float64),
					MemoryUsage:   memData.Data.Result[0].Values[i][1].(float64) / (1024 * 1024), // Convert to MiB
					NetworkRxRate: rxData.Data.Result[0].Values[i][1].(float64),
					NetworkTxRate: txData.Data.Result[0].Values[i][1].(float64),
				})
			}
		}

		podMetrics = append(podMetrics, PodMetricTimeSeries{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Metrics:   metrics,
		})
	}

	return &HistoricalMetrics{
		PodMetrics: podMetrics,
		TimeRange:  timeRange,
		StartTime:  startTime,
		EndTime:    endTime,
	}, nil
}

type PrometheusResponse struct {
	Status string         `json:"status"`
	Data   PrometheusData `json:"data"`
}

type PrometheusData struct {
	ResultType string             `json:"resultType"`
	Result     []PrometheusResult `json:"result"`
}

type PrometheusResult struct {
	Metric map[string]string `json:"metric"`
	Values [][]interface{}   `json:"values"`
}

func (c *MetricsController) QueryPrometheus(query string, start, end time.Time, step time.Duration) (*PrometheusResponse, error) {
	url := fmt.Sprintf("%s/api/v1/query_range", c.promURL)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("query", query)
	q.Add("start", fmt.Sprintf("%d", start.Unix()))
	q.Add("end", fmt.Sprintf("%d", end.Unix()))
	q.Add("step", fmt.Sprintf("%ds", int(step.Seconds())))
	req.URL.RawQuery = q.Encode()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var promResp PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&promResp); err != nil {
		return nil, err
	}

	return &promResp, nil
}
