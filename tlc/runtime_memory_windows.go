package tlc

import (
	"syscall"
	"unsafe"
)

func tlcPhysicalMemoryBytes() (int64, error) {
	var status struct {
		Length, Load                                                                                         uint32
		TotalPhys, AvailPhys, TotalPageFile, AvailPageFile, TotalVirtual, AvailVirtual, AvailExtendedVirtual uint64
	}
	status.Length = uint32(unsafe.Sizeof(status))
	proc := syscall.NewLazyDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx")
	ok, _, err := proc.Call(uintptr(unsafe.Pointer(&status)))
	if ok == 0 {
		return 0, err
	}
	return int64(status.TotalPhys), nil
}
