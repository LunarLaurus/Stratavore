//go:build linux

package procmetrics

import (
	"fmt"
	"io"
)

// runPS is not used on Linux (we read /proc directly), but must exist to
// satisfy the compiler when sampleViaPS is referenced in the shared file.
// In practice, sampleViaPS is only called on non-Linux builds.
func runPS(_ int) (io.Reader, error) {
	return nil, fmt.Errorf("procmetrics: runPS not supported on Linux (uses /proc instead)")
}
