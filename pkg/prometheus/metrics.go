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
)
