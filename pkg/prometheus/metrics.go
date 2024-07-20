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

	RequestNum = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "request_num",
		Help: "request_num",
	})

	TaskIdWorkerNum = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "task_id_worker_num",
		Help: "task_id_worker_num",
	})

	InputChanNum = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "input_chan_num",
		Help: "input_chan_num",
	})

	OutputChanNum = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "output_chan_num",
		Help: "output_chan_num",
	})
)
