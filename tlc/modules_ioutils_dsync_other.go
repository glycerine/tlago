//go:build !aix && !darwin && !linux && !netbsd && !openbsd && !solaris

package tlc

import "os"

// Platforms without a separate data-sync open flag use synchronous writes.
func fileDataSyncFlag() int { return os.O_SYNC }
