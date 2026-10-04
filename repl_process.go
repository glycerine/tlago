// Copyright (c) 2024, Oracle and/or its affiliates.
package tlago

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/glycerine/tlago/tlc"
)

// The registry holds the native optional CommunityModules
// index entry point. Installing the provider corresponds to placing that index
// on Java's class path; a nil provider means the optional class is absent.
var communityModulesIndex struct {
	sync.RWMutex
	provider func() (string, error)
}

func SetCommunityModulesPopularModules(provider func() (string, error)) {
	communityModulesIndex.Lock()
	defer communityModulesIndex.Unlock()
	communityModulesIndex.provider = provider
}

func replPopularModules() (modules string, available bool) {
	communityModulesIndex.RLock()
	provider := communityModulesIndex.provider
	communityModulesIndex.RUnlock()
	if provider == nil {
		return "", false
	}
	defer func() {
		if failure := recover(); failure != nil {
			if _, ok := failure.(*tlc.NoClassDefFoundError); ok {
				modules, available = "", false
				return
			}
			panic(failure)
		}
	}()
	modules, err := provider()
	return modules, err == nil
}

type REPL struct {
	tempDir     string
	specFile    string
	specFileSet bool
	writer      io.Writer
}

// Java constructs a buffered PrintWriter with autoFlush=false. Its write and
// flush methods retain IO failures internally rather than throwing IOException.
// REPL never calls checkError, so this adapter preserves that boundary.
type replPrintWriter struct{ buffered *bufio.Writer }

func (w *replPrintWriter) Write(p []byte) (int, error) {
	_, _ = w.buffered.Write(p)
	return len(p), nil
}
func (w *replPrintWriter) Flush() error { _ = w.buffered.Flush(); return nil }
func NewREPL(tempDir string) *REPL {
	return &REPL{tempDir: tempDir, writer: &replPrintWriter{buffered: bufio.NewWriterSize(os.Stdout, 8192)}}
}
func (r *REPL) SetSpecFile(file string) { r.specFile = file; r.specFileSet = true }

var replErrorLocation = regexp.MustCompile(`\nline [0-9]+, col [0-9]+ to line [0-9]+, col [0-9]+ of module tlarepl$`)

func replJavaTrim(s string) string {
	return strings.TrimFunc(s, func(r rune) bool { return r <= 0x20 })
}

// SpecProcessor passes the individual SANY errors to Assert.fail, rather than
// dropping them and returning only the generic parsing-failed message.
func replParsingFailure(diags Diagnostics) *tlc.TLCError {
	params := make([]string, 0)
	for _, d := range diags.Errors() {
		message := d.Message
		if d.Code == "E4200" {
			for _, prefix := range []string{"undefined identifier ", "undefined operator "} {
				if strings.HasPrefix(message, prefix) {
					message = "Unknown operator: `" + strings.TrimPrefix(message, prefix) + "'."
					break
				}
			}
		}
		end := d.Pos.SourceEnd()
		module := strings.TrimSuffix(filepath.Base(d.Pos.File), ".tla")
		params = append(params, fmt.Sprintf("line %d, col %d to line %d, col %d of module %s\n\n%s", d.Pos.Line, d.Pos.Column, end.Line, end.Column, module, message))
	}
	return tlc.NewTLCRuntimeException(tlc.ECTLCParsingFailed, params...)
}

func printREPLEvaluationFailure(expression string, err error) bool {
	eval, runtime := tlc.ClassifyEvaluationFailure(err)
	if eval {
		fmt.Fprintf(os.Stdout, "Error evaluating expression: '%s'\ntlc2.tool.EvalException: %s\n", expression, err.Error())
		return true
	}
	if runtime == nil {
		return false
	}
	fmt.Fprintf(os.Stdout, "Error evaluating expression: '%s'\n", expression)
	if len(runtime.Params) > 0 {
		params := append([]string(nil), runtime.Params...)
		if runtime.NullableParams != nil {
			for i, p := range runtime.NullableParams {
				if p == nil {
					params[i] = "null"
				} else {
					params[i] = *p
				}
			}
		}
		fmt.Fprintf(os.Stdout, "[%s]\n", strings.Join(params, ", "))
	} else if message := runtime.GetMessage(); message != nil {
		message := replErrorLocation.ReplaceAllString(replJavaTrim(*message), "")
		fmt.Fprintln(os.Stdout, replJavaTrim(strings.ReplaceAll(message, "\n", " ")))
	}
	return true
}

func (r *REPL) ProcessInput(expression string) (result string) {
	innerStarted := false
	defer func() {
		defer func() {
			if innerStarted {
				if flusher, ok := r.writer.(interface{ Flush() error }); ok {
					_ = flusher.Flush()
				}
				tlc.TLCOutput = nil
				tlc.TLCOutputToUserFile = false
			}
		}()
		if failure := recover(); failure != nil {
			if err, ok := failure.(error); innerStarted && ok && printREPLEvaluationFailure(expression, err) {
				result = ""
				return
			}
			panic(failure)
		}
	}()
	result, _, err := evaluateREPLExpression(expression, REPLEvalOptions{
		TempDir: r.tempDir, SpecFile: r.specFile, specFileSet: r.specFileSet, PrintOutput: r.writer, KeepTempFiles: true,
	}, true, &innerStarted)
	if err == nil {
		return result
	}
	if printREPLEvaluationFailure(expression, err) {
		return ""
	}
	if _, ok := err.(*os.PathError); ok {
		fmt.Fprintf(os.Stderr, "%s\n%s", err, debug.Stack())
		return ""
	}
	panic(err)
}
