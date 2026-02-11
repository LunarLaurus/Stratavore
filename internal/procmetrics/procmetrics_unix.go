//go:build !linux

package procmetrics

import (
	"fmt"
	"io"
	"os/exec"
	"strings"
)

// runPS spawns `ps -p <pid> -o %cpu,rss` and returns an io.Reader over stdout.
func runPS(pid int) (io.Reader, error) {
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "%cpu,rss")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("procmetrics: ps failed for pid %d: %w", pid, err)
	}
	return strings.NewReader(string(out)), nil
}
