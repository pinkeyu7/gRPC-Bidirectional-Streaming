package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	memoryAlloc = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_alloc",
		Help: "memory_alloc in bytes",
	})

	memoryTotalAlloc = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_total_alloc",
		Help: "memory_total_alloc in bytes",
	})

	memoryHeapAlloc = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_heap_alloc",
		Help: "memory_heap_alloc in bytes",
	})

	memorySys = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "memory_sys",
		Help: "memory_sys in bytes",
	})

	goroutineNum = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "goroutine_num",
		Help: "goroutine_num",
	})

	RequestNum = promauto.NewCounter(prometheus.CounterOpts{
		Name: "request_num",
		Help: "request_num",
	})

	WorkerRequestNum = promauto.NewCounter(prometheus.CounterOpts{
		Name: "worker_request_num",
		Help: "worker_request_num",
	})

	ResponseNum = promauto.NewCounter(prometheus.CounterOpts{
		Name: "response_num",
		Help: "response_num",
	})

	InputChanNum = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "input_chan_num",
		Help: "input_chan_num",
	})

	OutputChanNum = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "output_chan_num",
		Help: "output_chan_num",
	})

	ResponseTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "response_time",
		Help:    "Histogram of response time for handler in seconds",
		Buckets: []float64{.01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"status"})
)
