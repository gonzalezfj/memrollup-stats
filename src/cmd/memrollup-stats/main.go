package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gonzalezfj/memrollup-stats/src/internal/config"
	"github.com/gonzalezfj/memrollup-stats/src/pkg/memory"
	"github.com/gonzalezfj/memrollup-stats/src/pkg/output"
	"github.com/gonzalezfj/memrollup-stats/src/pkg/process"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handlers
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Parse command line arguments
	if err := config.ParseArgs(); err != nil {
		log.Fatalf("Failed to parse arguments: %v", err)
	}

	// Validate environment
	if err := config.ValidateEnvironment(); err != nil {
		log.Fatalf("Environment validation failed: %v", err)
	}

	// Create process manager
	pm := process.New(config.Get())
	if err := pm.Start(ctx); err != nil {
		log.Fatalf("Failed to start process: %v", err)
	}
	defer pm.Stop()

	// Create memory collector
	mc := memory.New(config.Get().PID)

	// Monitor memory
	metrics, err := monitorMemory(ctx, mc, pm)
	if err != nil {
		log.Printf("Memory monitoring error: %v", err)
	}

	// Output results
	if err := output.WriteResults(metrics, config.Get()); err != nil {
		log.Printf("Failed to output results: %v", err)
	}
}

func monitorMemory(ctx context.Context, mc memory.MemoryCollector, pm process.ProcessManager) (*memory.MemoryMetrics, error) {
	cfg := config.Get()
	sleepInterval := time.Duration(float64(time.Second) / cfg.Frequency)
	ticker := time.NewTicker(sleepInterval)
	defer ticker.Stop()

	metrics := &memory.MemoryMetrics{}
	sampleCount := 0
	if cfg.Verbose {
		log.Printf("Starting memory monitoring at frequency %.1f Hz...", cfg.Frequency)
	}

	// Create a channel to signal process termination
	processDone := make(chan error, 1)

	// Start a goroutine to wait for the process
	go func() {
		processDone <- pm.Wait()
	}()

	for {
		select {
		case <-ctx.Done():
			return metrics, ctx.Err()
		case err := <-processDone:
			if cfg.Verbose {
				log.Printf("Process has terminated: %v", err)
			}
			return metrics, nil
		case <-ticker.C:
			newMetrics, err := mc.Collect(ctx)
			if err != nil {
				return metrics, err
			}

			// Append new metrics
			metrics.Timestamps = append(metrics.Timestamps, newMetrics.Timestamps...)
			metrics.RSS = append(metrics.RSS, newMetrics.RSS...)
			metrics.PSS = append(metrics.PSS, newMetrics.PSS...)
			metrics.USS = append(metrics.USS, newMetrics.USS...)
			metrics.SharedClean = append(metrics.SharedClean, newMetrics.SharedClean...)
			metrics.SharedDirty = append(metrics.SharedDirty, newMetrics.SharedDirty...)
			metrics.PrivateClean = append(metrics.PrivateClean, newMetrics.PrivateClean...)
			metrics.PrivateDirty = append(metrics.PrivateDirty, newMetrics.PrivateDirty...)

			sampleCount++
			if cfg.Verbose {
				log.Printf("Collected sample #%d", sampleCount)
			}
		}
	}
}
