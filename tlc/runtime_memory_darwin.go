package tlc

import (
	"encoding/binary"
	"fmt"
	"syscall"
)

func tlcPhysicalMemoryBytes() (int64, error) {
	raw, err := syscall.Sysctl("hw.memsize")
	if err != nil {
		return 0, err
	}
	// Sysctl strips terminal zero bytes; restore the native uint64 width.
	if len(raw) > 8 {
		return 0, fmt.Errorf("invalid hw.memsize width %d", len(raw))
	}
	var data [8]byte
	copy(data[:], raw)
	return int64(binary.LittleEndian.Uint64(data[:])), nil
}
