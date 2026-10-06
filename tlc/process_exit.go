package tlc

import "fmt"

// ProcessExit carries a source System.exit through the Go library boundary.
// The command exits normally with the mapped status; the library does not kill
// its host process or run Java finally actions that System.exit bypasses.
type ProcessExit struct {
	ErrorCode int
}

func (e *ProcessExit) Error() string {
	return fmt.Sprintf("TLC exited with status %d", ExitStatusForErrorCode(e.ErrorCode))
}

func ExitTLCProcess(errorCode int) {
	panic(&ProcessExit{ErrorCode: errorCode})
}
