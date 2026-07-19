// Command hermesbench runs repeatable local workloads and optionally writes CPU
// and heap profiles for pprof analysis.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/sidwiskers/hermes"
	telegram "github.com/sidwiskers/hermes/types"
)

var fixture = []byte(`{"update_id":781234567,"message":{"message_id":42,"from":{"id":123456789,"is_bot":false,"first_name":"Ada"},"chat":{"id":123456789,"type":"private"},"date":1710000000,"text":"/start benchmark"}}`)

func main() {
	workload := flag.String("workload", "decode", "decode or route")
	iterations := flag.Int("n", 1_000_000, "number of operations")
	preserveRaw := flag.Bool("raw", false, "preserve raw update JSON in decode workload")
	cpuProfile := flag.String("cpuprofile", "", "write CPU profile")
	memProfile := flag.String("memprofile", "", "write heap profile")
	flag.Parse()

	if *iterations < 1 {
		fmt.Fprintln(os.Stderr, "-n must be positive")
		os.Exit(2)
	}
	stopCPU, err := startCPUProfile(*cpuProfile)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer stopCPU()

	started := time.Now()
	switch *workload {
	case "decode":
		runDecode(*iterations, *preserveRaw)
	case "route":
		runRoute(*iterations)
	default:
		fmt.Fprintf(os.Stderr, "unknown workload %q\n", *workload)
		os.Exit(2)
	}
	elapsed := time.Since(started)
	fmt.Printf("workload=%s operations=%d elapsed=%s ops_per_second=%.0f\n",
		*workload, *iterations, elapsed, float64(*iterations)/elapsed.Seconds())

	if *memProfile != "" {
		if err := writeHeapProfile(*memProfile); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}
}

func runDecode(iterations int, preserve bool) {
	for index := 0; index < iterations; index++ {
		update, err := telegram.DecodeUpdate(fixture, preserve)
		if err != nil || update.UpdateID == 0 {
			panic(err)
		}
	}
}

func runRoute(iterations int) {
	bot := hermes.New("TOKEN")
	bot.Command("start", func(*hermes.Context) error { return nil })
	update := &hermes.Update{Message: &hermes.Message{
		MessageID: 42,
		Chat:      hermes.Chat{ID: 123, Type: "private"},
		Text:      "/start benchmark",
	}}
	ctx := context.Background()
	for index := 0; index < iterations; index++ {
		if err := bot.Handle(ctx, update); err != nil {
			panic(err)
		}
	}
}

func startCPUProfile(path string) (func(), error) {
	if path == "" {
		return func() {}, nil
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create CPU profile: %w", err)
	}
	if err := pprof.StartCPUProfile(file); err != nil {
		_ = file.Close()
		return nil, fmt.Errorf("start CPU profile: %w", err)
	}
	return func() {
		pprof.StopCPUProfile()
		_ = file.Close()
	}, nil
}

func writeHeapProfile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create heap profile: %w", err)
	}
	defer file.Close()
	runtime.GC()
	if err := pprof.WriteHeapProfile(file); err != nil {
		return fmt.Errorf("write heap profile: %w", err)
	}
	return nil
}
