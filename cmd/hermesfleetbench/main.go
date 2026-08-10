// Command hermesfleetbench compares one multi-bot Fleet process with the same
// number of one-bot processes. Linux /proc supplies steady-idle RSS, PSS, and
// file-descriptor evidence; child goroutine counts come from the Go runtime.
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sidwiskers/hermes"
	"github.com/sidwiskers/hermes/fleet"
)

type measurement struct {
	Processes       int   `json:"processes"`
	Bots            int   `json:"bots"`
	RSSBytes        int64 `json:"rss_bytes"`
	PSSBytes        int64 `json:"pss_bytes,omitempty"`
	Goroutines      int   `json:"goroutines"`
	FileDescriptors int   `json:"file_descriptors"`
}

type savings struct {
	RSSBytes        int64   `json:"rss_bytes"`
	RSSPercent      float64 `json:"rss_percent"`
	PSSBytes        int64   `json:"pss_bytes,omitempty"`
	PSSPercent      float64 `json:"pss_percent,omitempty"`
	Goroutines      int     `json:"goroutines"`
	FileDescriptors int     `json:"file_descriptors"`
}

type report struct {
	Timestamp string      `json:"timestamp"`
	GoVersion string      `json:"go_version"`
	GOOS      string      `json:"goos"`
	GOARCH    string      `json:"goarch"`
	CPU       string      `json:"cpu"`
	Workload  string      `json:"workload"`
	Bots      int         `json:"bots"`
	Samples   int         `json:"samples"`
	Fleet     measurement `json:"fleet_median"`
	Separate  measurement `json:"separate_median"`
	Savings   savings     `json:"median_savings"`
}

type child struct {
	command     *exec.Cmd
	input       io.WriteCloser
	output      io.ReadCloser
	bots        int
	rssBytes    int64
	pssBytes    int64
	goroutines  int
	descriptors int
	stopOnce    sync.Once
}

type childReady struct {
	RSSBytes        int64 `json:"rss_bytes"`
	PSSBytes        int64 `json:"pss_bytes"`
	Goroutines      int   `json:"goroutines"`
	FileDescriptors int   `json:"file_descriptors"`
}

func main() {
	mode := flag.String("mode", "report", "report or child")
	bots := flag.Int("bots", 5, "number of idle webhook bots")
	samples := flag.Int("samples", 5, "number of process samples")
	flag.Parse()
	if *bots < 1 {
		fmt.Fprintln(os.Stderr, "bots must be positive")
		os.Exit(2)
	}
	var err error
	if *mode == "child" {
		err = runChild(*bots)
	} else {
		if *samples < 1 {
			fmt.Fprintln(os.Stderr, "samples must be positive")
			os.Exit(2)
		}
		err = runReport(*bots, *samples)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runReport(botCount, sampleCount int) error {
	if runtime.GOOS != "linux" {
		return errors.New("hermesfleetbench requires Linux /proc")
	}
	executable, err := os.Executable()
	if err != nil {
		return err
	}
	hostedSamples := make([]measurement, 0, sampleCount)
	separateSamples := make([]measurement, 0, sampleCount)
	for range sampleCount {
		hosted, separate, err := takeSample(executable, botCount)
		if err != nil {
			return err
		}
		hostedSamples = append(hostedSamples, hosted)
		separateSamples = append(separateSamples, separate)
	}
	hostedMeasurement := medianMeasurement(hostedSamples)
	separateMeasurement := medianMeasurement(separateSamples)
	hostedMeasurement.Bots = botCount
	separateMeasurement.Bots = botCount
	result := report{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
		CPU:       cpuModel(),
		Workload:  "idle webhook bots with independent Hermes routers and dispatchers",
		Bots:      botCount,
		Samples:   sampleCount,
		Fleet:     hostedMeasurement,
		Separate:  separateMeasurement,
		Savings: savings{
			RSSBytes:        separateMeasurement.RSSBytes - hostedMeasurement.RSSBytes,
			RSSPercent:      percentSaved(separateMeasurement.RSSBytes, hostedMeasurement.RSSBytes),
			PSSBytes:        separateMeasurement.PSSBytes - hostedMeasurement.PSSBytes,
			PSSPercent:      percentSaved(separateMeasurement.PSSBytes, hostedMeasurement.PSSBytes),
			Goroutines:      separateMeasurement.Goroutines - hostedMeasurement.Goroutines,
			FileDescriptors: separateMeasurement.FileDescriptors - hostedMeasurement.FileDescriptors,
		},
	}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func takeSample(executable string, botCount int) (measurement, measurement, error) {
	hosted, err := startChild(executable, botCount)
	if err != nil {
		return measurement{}, measurement{}, err
	}
	defer hosted.stop()
	separate := make([]*child, 0, botCount)
	defer func() {
		for _, process := range separate {
			process.stop()
		}
	}()
	for range botCount {
		process, err := startChild(executable, 1)
		if err != nil {
			return measurement{}, measurement{}, err
		}
		separate = append(separate, process)
	}
	hostedMeasurement, err := measure([]*child{hosted})
	if err != nil {
		return measurement{}, measurement{}, err
	}
	separateMeasurement, err := measure(separate)
	if err != nil {
		return measurement{}, measurement{}, err
	}
	return hostedMeasurement, separateMeasurement, nil
}

func runChild(botCount int) error {
	host := fleet.New(fleet.WithWebhookAddress("127.0.0.1:0"))
	for index := range botCount {
		bot := host.NewBot(
			fmt.Sprintf("TOKEN_%d", index),
			hermes.WithBotUsername(fmt.Sprintf("fleet_%d_bot", index)),
		)
		bot.OnUpdate(func(*hermes.Context) error { return nil })
		if err := host.Mount(
			fmt.Sprintf("bot-%d", index),
			bot,
			fleet.WithWebhook(fmt.Sprintf("/bot-%d", index), hermes.WebhookOptions{}),
		); err != nil {
			return err
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- host.Run(ctx) }()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	for host.WebhookAddress() == "" {
		select {
		case err := <-done:
			return fmt.Errorf("Fleet stopped before ready: %w", err)
		case <-deadline.C:
			cancel()
			return errors.New("Fleet readiness timed out")
		case <-time.After(time.Millisecond):
		}
	}
	runtime.GC()
	debug.FreeOSMemory()
	ready, err := currentProcessMeasurement()
	if err != nil {
		cancel()
		return err
	}
	if err := json.NewEncoder(os.Stdout).Encode(ready); err != nil {
		cancel()
		return err
	}
	_, _ = io.Copy(io.Discard, os.Stdin)
	cancel()
	return <-done
}

func startChild(executable string, bots int) (*child, error) {
	command := exec.Command(executable, "-mode", "child", "-bots", strconv.Itoa(bots))
	input, err := command.StdinPipe()
	if err != nil {
		return nil, err
	}
	output, err := command.StdoutPipe()
	if err != nil {
		input.Close()
		return nil, err
	}
	command.Stderr = os.Stderr
	if err := command.Start(); err != nil {
		input.Close()
		return nil, err
	}
	result := &child{command: command, input: input, output: output, bots: bots}
	scanner := bufio.NewScanner(output)
	if !scanner.Scan() {
		result.stop()
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, errors.New("Fleet child stopped before readiness")
	}
	var ready childReady
	if err := json.Unmarshal(scanner.Bytes(), &ready); err != nil {
		result.stop()
		return nil, err
	}
	result.rssBytes = ready.RSSBytes
	result.pssBytes = ready.PSSBytes
	result.goroutines = ready.Goroutines
	result.descriptors = ready.FileDescriptors
	return result, nil
}

func (c *child) stop() {
	if c == nil {
		return
	}
	c.stopOnce.Do(func() {
		if c.input != nil {
			_ = c.input.Close()
		}
		if c.command != nil {
			_ = c.command.Wait()
		}
		if c.output != nil {
			_ = c.output.Close()
		}
	})
}

func measure(children []*child) (measurement, error) {
	result := measurement{Processes: len(children)}
	for _, process := range children {
		result.RSSBytes += process.rssBytes
		result.PSSBytes += process.pssBytes
		result.Goroutines += process.goroutines
		result.FileDescriptors += process.descriptors
	}
	return result, nil
}

func medianMeasurement(samples []measurement) measurement {
	result := measurement{}
	if len(samples) == 0 {
		return result
	}
	result.Processes = samples[0].Processes
	result.Bots = samples[0].Bots
	rss := make([]int64, len(samples))
	pss := make([]int64, len(samples))
	goroutines := make([]int, len(samples))
	descriptors := make([]int, len(samples))
	for index, sample := range samples {
		rss[index] = sample.RSSBytes
		pss[index] = sample.PSSBytes
		goroutines[index] = sample.Goroutines
		descriptors[index] = sample.FileDescriptors
	}
	sort.Slice(rss, func(i, j int) bool { return rss[i] < rss[j] })
	sort.Slice(pss, func(i, j int) bool { return pss[i] < pss[j] })
	sort.Ints(goroutines)
	sort.Ints(descriptors)
	middle := len(samples) / 2
	result.RSSBytes = rss[middle]
	result.PSSBytes = pss[middle]
	result.Goroutines = goroutines[middle]
	result.FileDescriptors = descriptors[middle]
	return result
}

func currentProcessMeasurement() (childReady, error) {
	rss, err := procValue("status", "VmRSS:")
	if err != nil {
		return childReady{}, err
	}
	pss, _ := procValue("smaps_rollup", "Pss:")
	entries, err := os.ReadDir("/proc/self/fd")
	if err != nil {
		return childReady{}, err
	}
	return childReady{
		RSSBytes:        rss,
		PSSBytes:        pss,
		Goroutines:      runtime.NumGoroutine(),
		FileDescriptors: len(entries),
	}, nil
}

func procValue(file, prefix string) (int64, error) {
	path := filepath.Join("/proc/self", file)
	handle, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer handle.Close()
	scanner := bufio.NewScanner(handle)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, prefix) {
			continue
		}
		fields := strings.Fields(strings.TrimPrefix(line, prefix))
		if len(fields) == 0 {
			break
		}
		kilobytes, err := strconv.ParseInt(fields[0], 10, 64)
		if err != nil {
			return 0, err
		}
		return kilobytes * 1024, nil
	}
	if err := scanner.Err(); err != nil {
		return 0, err
	}
	return 0, fmt.Errorf("%s not found in %s", prefix, path)
}

func percentSaved(separate, hosted int64) float64 {
	if separate <= 0 {
		return 0
	}
	return float64(separate-hosted) * 100 / float64(separate)
}

func cpuModel() string {
	handle, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "unknown"
	}
	defer handle.Close()
	scanner := bufio.NewScanner(handle)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "model name") {
			continue
		}
		if index := strings.IndexByte(line, ':'); index >= 0 {
			return strings.TrimSpace(line[index+1:])
		}
	}
	return "unknown"
}
