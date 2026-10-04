package tlc

import (
	"fmt"
	"io"
	"os"
	"sync"
)

const (
	ToolIOSystem = 0
	ToolIOTool   = 1
)

// ToolIO's two streams share one message buffer. A println call stores one
// message even when its argument contains newlines; print accumulates a prefix.
var toolIO = struct {
	sync.Mutex
	mode        int
	out, err    io.Writer
	messages    []string
	nextMessage string
}{out: os.Stdout, err: os.Stderr, messages: make([]string, 0, 1)}

func ToolIOGetMode() int {
	toolIO.Lock()
	defer toolIO.Unlock()
	return toolIO.mode
}

func ToolIOSetMode(mode int) bool {
	toolIO.Lock()
	defer toolIO.Unlock()
	if mode != ToolIOSystem && mode != ToolIOTool {
		return false
	}
	toolIO.mode = mode
	if mode == ToolIOSystem {
		toolIO.out, toolIO.err = os.Stdout, os.Stderr
	}
	return true
}

// Reset retains the mode and streams, exactly as util.ToolIO.reset does.
func ToolIOReset() {
	toolIO.Lock()
	defer toolIO.Unlock()
	toolIO.messages = make([]string, 0, 1)
	toolIO.nextMessage = ""
}

func ToolIOGetAllMessages() []string {
	toolIO.Lock()
	defer toolIO.Unlock()
	messages := make([]string, len(toolIO.messages))
	copy(messages, toolIO.messages)
	if toolIO.nextMessage != "" {
		// Source arraycopy uses retLen, including the pending message. Preserve
		// its bounds failure when the completed-message array is exactly full.
		if len(toolIO.messages)+1 > cap(toolIO.messages) {
			text := fmt.Sprintf("arraycopy: last source index %d out of bounds for object array[%d]", len(toolIO.messages)+1, cap(toolIO.messages))
			panic(&ArrayIndexOutOfBoundsException{javaExceptionBase: newJavaExceptionBase(javaString(text), nil)})
		}
		messages = append(messages, toolIO.nextMessage)
	}
	return messages
}

func ToolIOPrint(text string)      { toolIOWrite(text, false, false) }
func ToolIOPrintln(text string)    { toolIOWrite(text, true, false) }
func ToolIOErrPrint(text string)   { toolIOWrite(text, false, true) }
func ToolIOErrPrintln(text string) { toolIOWrite(text, true, true) }

func toolIOWrite(text string, newline, toError bool) {
	toolIO.Lock()
	defer toolIO.Unlock()
	if toolIO.mode == ToolIOTool {
		if newline {
			if len(toolIO.messages) == cap(toolIO.messages) {
				grown := make([]string, len(toolIO.messages), 2*cap(toolIO.messages))
				copy(grown, toolIO.messages)
				toolIO.messages = grown
			}
			toolIO.messages = append(toolIO.messages, toolIO.nextMessage+text)
			toolIO.nextMessage = ""
		} else {
			toolIO.nextMessage += text
		}
		return
	}
	writer := toolIO.out
	if toError {
		writer = toolIO.err
	}
	// Java PrintStream swallows IOExceptions rather than changing MP's result.
	if newline {
		_, _ = fmt.Fprintln(writer, text)
	} else {
		_, _ = fmt.Fprint(writer, text)
	}
}

// The Go command boundary supplies process streams explicitly. TOOL buffering
// remains independent of those streams, as in the Java embedding boundary.
func ToolIOSetSystemStreams(out, err io.Writer) func() {
	toolIO.Lock()
	oldOut, oldErr := toolIO.out, toolIO.err
	toolIO.out, toolIO.err = out, err
	toolIO.Unlock()
	return func() { toolIO.Lock(); toolIO.out, toolIO.err = oldOut, oldErr; toolIO.Unlock() }
}
