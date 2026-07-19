// Command hermessoak runs a reproducible in-process webhook/runtime soak and
// emits one JSON report suitable for release evidence.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/observe"
)

var latencyBounds = [...]time.Duration{
	50 * time.Microsecond,
	100 * time.Microsecond,
	250 * time.Microsecond,
	500 * time.Microsecond,
	time.Millisecond,
	2 * time.Millisecond,
	5 * time.Millisecond,
	10 * time.Millisecond,
}

type latencyHistogram struct {
	buckets [len(latencyBounds) + 1]atomic.Uint64
}

func (h *latencyHistogram) observe(duration time.Duration) {
	for index, bound := range latencyBounds {
		if duration <= bound {
			h.buckets[index].Add(1)
			return
		}
	}
	h.buckets[len(latencyBounds)].Add(1)
}

func (h *latencyHistogram) snapshot() map[string]uint64 {
	result := make(map[string]uint64, len(h.buckets))
	for index, bound := range latencyBounds {
		result["le_"+bound.String()] = h.buckets[index].Load()
	}
	result["gt_"+latencyBounds[len(latencyBounds)-1].String()] = h.buckets[len(latencyBounds)].Load()
	return result
}

type report struct {
	Timestamp           time.Time         `json:"timestamp"`
	GoVersion           string            `json:"go_version"`
	Duration            string            `json:"duration"`
	GOMAXPROCS          int               `json:"gomaxprocs"`
	Workers             int               `json:"workers"`
	Routes              int               `json:"routes"`
	MaxConcurrent       int               `json:"max_concurrent_updates"`
	HandlerDelay        string            `json:"handler_delay"`
	Requests            uint64            `json:"requests"`
	Accepted            uint64            `json:"accepted"`
	Overloaded          uint64            `json:"overloaded"`
	Unexpected          uint64            `json:"unexpected_status"`
	RequestsPerSecond   float64           `json:"requests_per_second"`
	ResponseLatency     map[string]uint64 `json:"response_latency_buckets"`
	Metrics             observe.Snapshot  `json:"metrics"`
	StartHeapBytes      uint64            `json:"start_heap_bytes"`
	EndHeapBytes        uint64            `json:"end_heap_bytes"`
	HeapDeltaBytes      int64             `json:"heap_delta_bytes"`
	Mallocs             uint64            `json:"mallocs"`
	TotalAllocatedBytes uint64            `json:"total_allocated_bytes"`
	StartGoroutines     int               `json:"start_goroutines"`
	EndGoroutines       int               `json:"end_goroutines"`
	PeakGoroutines      int64             `json:"peak_goroutines"`
	Drained             bool              `json:"drained"`
}

func main() {
	duration := flag.Duration("duration", 10*time.Minute, "soak duration")
	workers := flag.Int("workers", runtime.GOMAXPROCS(0)*4, "concurrent request generators")
	routes := flag.Int("routes", 1000, "exact command routes")
	maximum := flag.Int("max-concurrent", runtime.GOMAXPROCS(0)*8, "maximum executing handlers")
	handlerDelay := flag.Duration("handler-delay", 0, "optional work simulated by each accepted handler")
	flag.Parse()
	if *duration <= 0 || *workers <= 0 || *routes <= 0 || *maximum <= 0 || *handlerDelay < 0 {
		log.Fatal("duration, workers, routes, and max-concurrent must be positive; handler-delay cannot be negative")
	}

	metrics := new(observe.Metrics)
	bot := hermes.New(
		"SOAK_TOKEN",
		hermes.WithBotUsername("soak_bot"),
		hermes.WithMaxConcurrentUpdates(*maximum),
	)
	bot.Use(observe.Middleware(metrics))
	for index := range *routes {
		bot.Command("r"+strconv.Itoa(index), func(*hermes.Context) error {
			if *handlerDelay > 0 {
				time.Sleep(*handlerDelay)
			}
			return nil
		})
	}
	handler := bot.WebhookHandler(hermes.WebhookOptions{})
	payloads := make([]string, *routes)
	for index := range payloads {
		payloads[index] = fmt.Sprintf(
			`{"update_id":%d,"message":{"message_id":1,"from":{"id":2,"is_bot":false,"first_name":"Soak"},"chat":{"id":1,"type":"private"},"text":"/r%d"}}`,
			index+1,
			index,
		)
	}

	runtime.GC()
	var startMemory runtime.MemStats
	runtime.ReadMemStats(&startMemory)
	startGoroutines := runtime.NumGoroutine()
	var peakGoroutines atomic.Int64
	peakGoroutines.Store(int64(startGoroutines))

	ctx, cancel := context.WithTimeout(context.Background(), *duration)
	defer cancel()
	started := time.Now()
	var requests, accepted, overloaded, unexpected atomic.Uint64
	var latencies latencyHistogram
	var group sync.WaitGroup
	for worker := range *workers {
		group.Add(1)
		go func() {
			defer group.Done()
			index := worker % len(payloads)
			for ctx.Err() == nil {
				requestStarted := time.Now()
				request := httptest.NewRequest(http.MethodPost, "/telegram", strings.NewReader(payloads[index]))
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, request)
				latencies.observe(time.Since(requestStarted))
				requests.Add(1)
				switch response.Code {
				case http.StatusOK:
					accepted.Add(1)
				case http.StatusServiceUnavailable:
					overloaded.Add(1)
				default:
					unexpected.Add(1)
				}
				index++
				if index == len(payloads) {
					index = 0
				}
			}
		}()
	}
	group.Add(1)
	go func() {
		defer group.Done()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				current := int64(runtime.NumGoroutine())
				for peak := peakGoroutines.Load(); current > peak; peak = peakGoroutines.Load() {
					if peakGoroutines.CompareAndSwap(peak, current) {
						break
					}
				}
			}
		}
	}()
	group.Wait()
	bot.Wait()
	elapsed := time.Since(started)

	runtime.GC()
	var endMemory runtime.MemStats
	runtime.ReadMemStats(&endMemory)
	result := report{
		Timestamp:           time.Now().UTC(),
		GoVersion:           runtime.Version(),
		Duration:            elapsed.String(),
		GOMAXPROCS:          runtime.GOMAXPROCS(0),
		Workers:             *workers,
		Routes:              *routes,
		MaxConcurrent:       *maximum,
		HandlerDelay:        handlerDelay.String(),
		Requests:            requests.Load(),
		Accepted:            accepted.Load(),
		Overloaded:          overloaded.Load(),
		Unexpected:          unexpected.Load(),
		RequestsPerSecond:   float64(requests.Load()) / elapsed.Seconds(),
		ResponseLatency:     latencies.snapshot(),
		Metrics:             metrics.Snapshot(),
		StartHeapBytes:      startMemory.HeapAlloc,
		EndHeapBytes:        endMemory.HeapAlloc,
		HeapDeltaBytes:      int64(endMemory.HeapAlloc) - int64(startMemory.HeapAlloc),
		Mallocs:             endMemory.Mallocs - startMemory.Mallocs,
		TotalAllocatedBytes: endMemory.TotalAlloc - startMemory.TotalAlloc,
		StartGoroutines:     startGoroutines,
		EndGoroutines:       runtime.NumGoroutine(),
		PeakGoroutines:      peakGoroutines.Load(),
		Drained:             metrics.Snapshot().UpdatesInFlight == 0,
	}
	encoder := json.NewEncoder(log.Writer())
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(result); err != nil {
		log.Fatal(err)
	}
	if result.Unexpected != 0 || !result.Drained {
		log.Fatal("soak invariant failed")
	}
}
