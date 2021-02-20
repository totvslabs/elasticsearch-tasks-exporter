package collector

import (
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/totvslabs/elasticsearch-tasks-exporter/client"
)

const (
	namespace = "elasticsearch"
	subsystem = "pending_tasks"
)

// NewCollector collector
func NewCollector(client client.Client) prometheus.Collector {
	return &collector{
		client: client,

		// default metrics
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "up"),
			"API is responding",
			nil,
			nil,
		),
		scrapeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "scrape_duration_seconds"),
			"Scrape duration in seconds",
			nil,
			nil,
		),
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "total"),
			"Total pending tasks by source and priority",
			[]string{"source", "priority"},
			nil,
		),
	}
}

type collector struct {
	mutex  sync.Mutex
	client client.Client

	up             *prometheus.Desc
	scrapeDuration *prometheus.Desc
	total          *prometheus.Desc
}

// Describe all metrics
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.scrapeDuration
	ch <- c.total
}

// Collect all metrics
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	defer func() {
		ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
	}()

	log.Info("Collecting ES Pending Tasks metrics...")
	tasks, err := c.client.Tasks()
	log.With("tasks", tasks).With("error", err).Debug("collected")
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		log.With("error", err).Error("failed to scrape ES")
		return
	}

	log.Debugf("tasks: %d", len(tasks))
	groupedTasks := map[string][]client.Task{}

	for _, task := range tasks {
		key := task.Source + "/" + task.Priority
		groupedTasks[key] = append(groupedTasks[key], task)
	}

	for _, tasks := range groupedTasks {
		ch <- prometheus.MustNewConstMetric(c.total, prometheus.CounterValue, float64(len(tasks)), strings.ToLower(tasks[0].Source), strings.ToLower(tasks[0].Priority))
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
}
