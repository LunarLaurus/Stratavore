// Package procmetrics provides lightweight CPU and memory sampling for a
// running OS process. It reads directly from /proc on Linux and falls back to
// a `ps` subprocess on other UNIX-like platforms (macOS, BSDs).
//
// No third-party dependencies are required.
package procmetrics

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Sample holds a single CPU/memory snapshot for a process.
type Sample struct {
	PID       int
	CPUPercent float64 // 0–100 (per-core; may exceed 100 on multi-core)
	MemoryMB  int64   // resident set size in megabytes
	Timestamp time.Time
}

// Sampler takes repeated measurements for a single PID and computes CPU usage
// as the delta between successive samples.
type Sampler struct {
	pid      int
	prevTick uint64
	prevTime time.Time
}

// NewSampler creates a Sampler for the given PID.
func NewSampler(pid int) *Sampler {
	return &Sampler{pid: pid}
}

// Sample collects one metrics snapshot.
//
// On the first call, CPUPercent will always be 0 because there is no prior
// measurement to diff against.
func (s *Sampler) Sample() (Sample, error) {
	now := time.Now()

	cpuPct := 0.0
	var memMB int64

	if runtime.GOOS == "linux" {
		tick, rss, err := readProcStat(s.pid)
		if err != nil {
			return Sample{}, err
		}
		memMB = rss

		if !s.prevTime.IsZero() {
			elapsed := now.Sub(s.prevTime).Seconds()
			tickDelta := float64(tick - s.prevTick)
			// Linux: clock ticks per second is typically 100 (SC_CLK_TCK)
			ticksPerSec := 100.0
			if elapsed > 0 {
				cpuPct = (tickDelta / ticksPerSec) / elapsed * 100.0
			}
		}
		s.prevTick = tick
	} else {
		// macOS / other UNIX: fall back to `ps`
		var err error
		cpuPct, memMB, err = sampleViaPS(s.pid)
		if err != nil {
			return Sample{}, err
		}
	}

	s.prevTime = now
	return Sample{
		PID:        s.pid,
		CPUPercent: cpuPct,
		MemoryMB:   memMB,
		Timestamp:  now,
	}, nil
}

// ─── Linux: /proc/<pid>/stat + /proc/<pid>/statm ─────────────────────────────

// readProcStat parses /proc/<pid>/stat for total CPU ticks and
// /proc/<pid>/statm for resident memory pages → MB.
func readProcStat(pid int) (totalTicks uint64, rssBytes int64, err error) {
	// --- CPU ticks from /proc/<pid>/stat ---
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	statData, err := os.ReadFile(statPath)
	if err != nil {
		return 0, 0, fmt.Errorf("procmetrics: read %s: %w", statPath, err)
	}

	// The format is: pid (comm) state ppid ... utime(14) stime(15) ...
	// The comm field can contain spaces and parentheses, so parse after the
	// closing ')'.
	raw := string(statData)
	closeParen := strings.LastIndex(raw, ")")
	if closeParen < 0 {
		return 0, 0, fmt.Errorf("procmetrics: malformed stat: %s", statPath)
	}
	fields := strings.Fields(raw[closeParen+1:])
	// fields[0] = state, fields[1] = ppid, ... fields[11] = utime, fields[12] = stime
	if len(fields) < 13 {
		return 0, 0, fmt.Errorf("procmetrics: too few fields in %s", statPath)
	}
	utime, err := strconv.ParseUint(fields[11], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("procmetrics: parse utime: %w", err)
	}
	stime, err := strconv.ParseUint(fields[12], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("procmetrics: parse stime: %w", err)
	}
	totalTicks = utime + stime

	// --- RSS from /proc/<pid>/statm ---
	statmPath := fmt.Sprintf("/proc/%d/statm", pid)
	statmData, err := os.ReadFile(statmPath)
	if err != nil {
		return totalTicks, 0, fmt.Errorf("procmetrics: read %s: %w", statmPath, err)
	}
	statmFields := strings.Fields(string(statmData))
	if len(statmFields) < 2 {
		return totalTicks, 0, fmt.Errorf("procmetrics: malformed statm: %s", statmPath)
	}
	rssPages, err := strconv.ParseInt(statmFields[1], 10, 64)
	if err != nil {
		return totalTicks, 0, fmt.Errorf("procmetrics: parse rss pages: %w", err)
	}
	pageSize := int64(os.Getpagesize())
	rssBytes = (rssPages * pageSize) / (1024 * 1024) // pages → MB

	return totalTicks, rssBytes, nil
}

// ─── macOS / other UNIX: `ps` fallback ───────────────────────────────────────

func sampleViaPS(pid int) (cpuPct float64, memMB int64, err error) {
	// ps -p <pid> -o %cpu,rss
	f, execErr := runPS(pid)
	if execErr != nil {
		return 0, 0, execErr
	}
	data, readErr := io.ReadAll(f)
	if readErr != nil {
		return 0, 0, readErr
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) < 2 {
		return 0, 0, fmt.Errorf("procmetrics: ps returned no data for pid %d", pid)
	}
	fields := strings.Fields(lines[1])
	if len(fields) < 2 {
		return 0, 0, fmt.Errorf("procmetrics: unexpected ps output: %q", lines[1])
	}
	cpuPct, err = strconv.ParseFloat(fields[0], 64)
	if err != nil {
		return 0, 0, fmt.Errorf("procmetrics: parse cpu: %w", err)
	}
	rssKB, err := strconv.ParseInt(fields[1], 10, 64)
	if err != nil {
		return 0, 0, fmt.Errorf("procmetrics: parse rss: %w", err)
	}
	memMB = rssKB / 1024 // kB → MB
	return cpuPct, memMB, nil
}
