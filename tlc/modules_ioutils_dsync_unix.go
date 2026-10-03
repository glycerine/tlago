//go:build aix || darwin || linux || netbsd || openbsd || solaris

package tlc

import "syscall"

func ioUtilsDataSyncFlag() int { return syscall.O_DSYNC }
