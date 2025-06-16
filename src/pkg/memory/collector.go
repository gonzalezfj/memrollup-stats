package memory

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// MemoryMetrics holds the memory statistics.
type MemoryMetrics struct {
	RSS          []int64
	PSS          []int64
	USS          []int64
	SharedClean  []int64
	SharedDirty  []int64
	PrivateClean []int64
	PrivateDirty []int64
	Timestamps   []time.Time
}

// MemoryCollector defines the interface for memory statistics collection.
type MemoryCollector interface {
	Collect(ctx context.Context) (*MemoryMetrics, error)
}

// RealMemoryCollector implements MemoryCollector for actual memory collection.
type RealMemoryCollector struct {
	pid int
}

// New creates a new memory collector.
func New(pid int) MemoryCollector {
	return &RealMemoryCollector{pid: pid}
}

// Collect implements MemoryCollector interface.
func (mc *RealMemoryCollector) Collect(ctx context.Context) (*MemoryMetrics, error) {
	smapsPath := fmt.Sprintf("/proc/%d/smaps_rollup", mc.pid)
	data, err := os.ReadFile(smapsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read smaps_rollup: %w", err)
	}

	metrics := &MemoryMetrics{
		Timestamps: []time.Time{time.Now()},
	}

	// Parse memory stats
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		value, err := strconv.ParseInt(fields[1], 10, 64)
		if err != nil {
			continue
		}

		switch fields[0] {
		case "Rss:":
			metrics.RSS = append(metrics.RSS, value)
		case "Pss:":
			metrics.PSS = append(metrics.PSS, value)
		case "Shared_Clean:":
			metrics.SharedClean = append(metrics.SharedClean, value)
		case "Shared_Dirty:":
			metrics.SharedDirty = append(metrics.SharedDirty, value)
		case "Private_Clean:":
			metrics.PrivateClean = append(metrics.PrivateClean, value)
		case "Private_Dirty:":
			metrics.PrivateDirty = append(metrics.PrivateDirty, value)
		}
	}

	// Calculate USS
	if len(metrics.PrivateClean) > 0 && len(metrics.PrivateDirty) > 0 {
		metrics.USS = append(metrics.USS, metrics.PrivateClean[0]+metrics.PrivateDirty[0])
	}

	return metrics, nil
}
