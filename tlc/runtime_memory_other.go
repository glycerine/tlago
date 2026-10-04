//go:build !linux && !darwin && !windows

package tlc

import "fmt"

func tlcPhysicalMemoryBytes() (int64, error) {
	return 0, fmt.Errorf("physical-memory discovery is unavailable on this platform")
}
