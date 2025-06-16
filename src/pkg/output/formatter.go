package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gonzalezfj/memrollup-stats/src/internal/config"
	"github.com/gonzalezfj/memrollup-stats/src/pkg/memory"
)

// Statistics holds the calculated statistics for each metric.
type Statistics struct {
	Min   int64   `json:"min"`
	Max   int64   `json:"max"`
	Avg   float64 `json:"avg"`
	Std   float64 `json:"std"`
	Count int     `json:"count"`
	Unit  string  `json:"unit"`
}

// OutputData holds the final output structure.
type OutputData struct {
	Metadata struct {
		Command   string    `json:"command"`
		PID       int       `json:"pid"`
		Frequency float64   `json:"frequency"`
		Duration  float64   `json:"duration"`
		Samples   int       `json:"samples"`
		StartTime time.Time `json:"start_time"`
	} `json:"metadata"`
	Statistics map[string]Statistics `json:"statistics"`
}

// OutputFormatter defines the interface for output formatting.
type OutputFormatter interface {
	Format(metrics *memory.MemoryMetrics, config *config.Config) ([]byte, error)
	Write(w io.Writer, data []byte) error
}

// JSONFormatter implements OutputFormatter for JSON output.
type JSONFormatter struct{}

// CSVFormatter implements OutputFormatter for CSV output.
type CSVFormatter struct{}

// NewFormatter creates a new formatter based on the configuration.
func NewFormatter(cfg *config.Config) OutputFormatter {
	if cfg.JSONOutput {
		return &JSONFormatter{}
	}
	return &CSVFormatter{}
}

// Format implements OutputFormatter interface for JSON.
func (f *JSONFormatter) Format(metrics *memory.MemoryMetrics, config *config.Config) ([]byte, error) {
	duration := metrics.Timestamps[len(metrics.Timestamps)-1].Sub(metrics.Timestamps[0]).Seconds()

	data := OutputData{}
	data.Metadata.Command = config.Command
	data.Metadata.PID = config.PID
	data.Metadata.Frequency = config.Frequency
	data.Metadata.Duration = duration
	data.Metadata.Samples = len(metrics.Timestamps)
	data.Metadata.StartTime = config.StartTime

	data.Statistics = make(map[string]Statistics)
	data.Statistics["RSS"] = calculateStats(metrics.RSS)
	data.Statistics["PSS"] = calculateStats(metrics.PSS)
	data.Statistics["USS"] = calculateStats(metrics.USS)
	data.Statistics["SHARED_CLEAN"] = calculateStats(metrics.SharedClean)
	data.Statistics["SHARED_DIRTY"] = calculateStats(metrics.SharedDirty)
	data.Statistics["PRIVATE_CLEAN"] = calculateStats(metrics.PrivateClean)
	data.Statistics["PRIVATE_DIRTY"] = calculateStats(metrics.PrivateDirty)

	return json.MarshalIndent(data, "", "  ")
}

// Write implements OutputFormatter interface for JSON.
func (f *JSONFormatter) Write(w io.Writer, data []byte) error {
	_, err := w.Write(data)
	return err
}

// Format implements OutputFormatter interface for CSV.
func (f *CSVFormatter) Format(metrics *memory.MemoryMetrics, config *config.Config) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	header := []string{"metric", "min_kb", "max_kb", "avg_kb", "std_kb", "samples", "duration_sec"}
	if err := writer.Write(header); err != nil {
		return nil, err
	}

	duration := metrics.Timestamps[len(metrics.Timestamps)-1].Sub(metrics.Timestamps[0]).Seconds()
	metricsMap := map[string][]int64{
		"RSS":           metrics.RSS,
		"PSS":           metrics.PSS,
		"USS":           metrics.USS,
		"SHARED_CLEAN":  metrics.SharedClean,
		"SHARED_DIRTY":  metrics.SharedDirty,
		"PRIVATE_CLEAN": metrics.PrivateClean,
		"PRIVATE_DIRTY": metrics.PrivateDirty,
	}

	for metric, values := range metricsMap {
		stats := calculateStats(values)
		record := []string{
			metric,
			strconv.FormatInt(stats.Min, 10),
			strconv.FormatInt(stats.Max, 10),
			fmt.Sprintf("%.2f", stats.Avg),
			fmt.Sprintf("%.2f", stats.Std),
			strconv.Itoa(stats.Count),
			fmt.Sprintf("%.3f", duration),
		}
		if err := writer.Write(record); err != nil {
			return nil, err
		}
	}

	writer.Flush()
	return []byte(buf.String()), writer.Error()
}

// Write implements OutputFormatter interface for CSV.
func (f *CSVFormatter) Write(w io.Writer, data []byte) error {
	_, err := w.Write(data)
	return err
}

// WriteResults writes the results to the appropriate output.
func WriteResults(metrics *memory.MemoryMetrics, cfg *config.Config) error {
	if len(metrics.Timestamps) == 0 {
		fmt.Println("No data collected")
		return nil
	}

	formatter := NewFormatter(cfg)
	data, err := formatter.Format(metrics, cfg)
	if err != nil {
		return fmt.Errorf("failed to format output: %w", err)
	}

	if cfg.OutputFile != "" {
		file, err := os.Create(cfg.OutputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		return formatter.Write(file, data)
	}

	return formatter.Write(os.Stdout, data)
}

// calculateStats calculates statistics for a slice of values.
func calculateStats(values []int64) Statistics {
	if len(values) == 0 {
		return Statistics{Unit: "kB"}
	}

	min := values[0]
	max := values[0]
	sum := int64(0)

	for _, v := range values {
		if v < min {
			min = v
		}
		if v > max {
			max = v
		}
		sum += v
	}

	avg := float64(sum) / float64(len(values))

	// Calculate standard deviation
	var sumSq float64
	for _, v := range values {
		diff := float64(v) - avg
		sumSq += diff * diff
	}
	variance := sumSq / float64(len(values))
	std := math.Sqrt(variance)

	return Statistics{
		Min:   min,
		Max:   max,
		Avg:   avg,
		Std:   std,
		Count: len(values),
		Unit:  "kB",
	}
}
