package tlc

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"
)

const (
	ToolIOSystem = 0
	ToolIOTool   = 1
)

var toolIOSystemOutHigh, toolIOSystemErrHigh uint16

// ToolIO's two streams share one message buffer. A println call stores one
// message even when its argument contains newlines; print accumulates a prefix.
var toolIO = struct {
	sync.Mutex
	mode                   int
	captureOut, captureErr bool
	out, err               io.Writer
	outHigh, errHigh       *uint16
	messages               []string
	nextMessage            string
}{out: os.Stdout, err: os.Stderr, outHigh: &toolIOSystemOutHigh, errHigh: &toolIOSystemErrHigh, messages: make([]string, 0, 1)}

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
	toolIO.captureOut, toolIO.captureErr = mode == ToolIOTool, mode == ToolIOTool
	if mode == ToolIOSystem {
		toolIO.out, toolIO.err = os.Stdout, os.Stderr
		toolIO.outHigh, toolIO.errHigh = &toolIOSystemOutHigh, &toolIOSystemErrHigh
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
	capture := toolIO.captureOut
	if toError {
		capture = toolIO.captureErr
	}
	if capture {
		if newline {
			if len(toolIO.messages) == cap(toolIO.messages) {
				grown := make([]string, len(toolIO.messages), 2*cap(toolIO.messages))
				copy(grown, toolIO.messages)
				toolIO.messages = grown
			}
			toolIO.messages = append(toolIO.messages, javaStringConcat(toolIO.nextMessage, text))
			toolIO.nextMessage = ""
		} else {
			toolIO.nextMessage = javaStringConcat(toolIO.nextMessage, text)
		}
		return
	}
	writer := toolIO.out
	high := toolIO.outHigh
	if toError {
		writer = toolIO.err
		high = toolIO.errHigh
	}
	// Java PrintStream swallows IOExceptions rather than changing MP's result.
	_, _ = fmt.Fprint(writer, toolIOEncodeUTF8(text, newline, high))
}

// UTF-8 PrintStreams retain a trailing high surrogate across print calls. A
// following low surrogate completes it; println's newline forces replacement.
func toolIOEncodeUTF8(text string, newline bool, high *uint16) string {
	units := javaStringUTF16(text)
	if *high != 0 {
		units = append([]uint16{*high}, units...)
		*high = 0
	}
	if newline {
		units = append(units, '\n')
	}
	if len(units) != 0 && units[len(units)-1] >= 0xd800 && units[len(units)-1] <= 0xdbff {
		*high = units[len(units)-1]
		units = units[:len(units)-1]
	}
	return javaStringUTF8(javaStringFromUTF16(units))
}

// Assigning native out/err streams overrides the buffered stream instances while
// retaining ToolIO.mode, just as assigning ToolIO.out and ToolIO.err does in Java.
func ToolIOSetSystemStreams(out, err io.Writer) func() {
	toolIO.Lock()
	oldOut, oldErr := toolIO.out, toolIO.err
	oldOutHigh, oldErrHigh := toolIO.outHigh, toolIO.errHigh
	oldCaptureOut, oldCaptureErr := toolIO.captureOut, toolIO.captureErr
	toolIO.out, toolIO.err = out, err
	toolIO.outHigh, toolIO.errHigh = new(uint16), new(uint16)
	if out != nil && reflect.TypeOf(out).Comparable() && out == err {
		toolIO.errHigh = toolIO.outHigh
	}
	toolIO.captureOut, toolIO.captureErr = false, false
	toolIO.Unlock()
	return func() {
		toolIO.Lock()
		toolIO.out, toolIO.err = oldOut, oldErr
		toolIO.outHigh, toolIO.errHigh = oldOutHigh, oldErrHigh
		toolIO.captureOut, toolIO.captureErr = oldCaptureOut, oldCaptureErr
		toolIO.Unlock()
	}
}

// ToolIORawOutputStream is the byte-writing boundary of the current PrintStream.
// Source ToolPrintStream captures string print/println only; inherited byte
// writes go to an unconnected pipe and their IOException is swallowed.
func ToolIORawOutputStream() io.Writer {
	toolIO.Lock()
	defer toolIO.Unlock()
	if toolIO.captureOut {
		return io.Discard
	}
	return toolIO.out
}
