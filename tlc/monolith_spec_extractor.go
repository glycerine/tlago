// Copyright (c) 2020 Microsoft Research. All rights reserved.
package tlc

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

const monolithWhitespace = `[ \t\n\x0B\f\r]`

func MonolithGetConfig(configFile string) string {
	if strings.HasSuffix(configFile, ".tla") {
		return configFile
	}
	return configFile + ".cfg"
}

func monolithLines(source string) []string {
	source = strings.ReplaceAll(strings.ReplaceAll(source, "\r\n", "\n"), "\r", "\n")
	lines := strings.Split(source, "\n")
	// BufferedReader.readLine does not invent a final line after a terminator.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return lines
}

func monolithEnd(line string) bool {
	// Java's default dot excludes these line terminators; readLine already
	// consumed CR and LF, but preserves NEL and the Unicode line separators.
	return strings.HasPrefix(line, "====") && !strings.ContainsAny(line, "\u0085\u2028\u2029")
}

func ExtractMonolithModuleSource(source, moduleName string) (string, bool) {
	start := regexp.MustCompile(`^-{4,}` + monolithWhitespace + `*MODULE` + monolithWhitespace + `+` + regexp.QuoteMeta(moduleName) + monolithWhitespace + `*-{3,}$`)
	active := false
	var out strings.Builder
	for _, line := range monolithLines(source) {
		if active && monolithEnd(line) {
			out.WriteString(line)
			out.WriteByte('\n')
			break
		}
		if !active && start.MatchString(line) {
			active = true
			out.WriteString(line)
			out.WriteByte('\n')
			continue
		}
		if active {
			out.WriteString(line)
			out.WriteByte('\n')
		}
	}
	return out.String(), active
}

var namedInputStreamReferences struct {
	sync.Mutex
	count int
}

type NamedInputStream struct {
	*os.File
	fileName, moduleName, inputFile string
	closed                          bool
	closeMu                         sync.Mutex
}

func NewNamedInputStream(file, module, input string) (*NamedInputStream, error) {
	f, err := os.Open(input)
	if err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	}
	namedInputStreamReferences.Lock()
	if namedInputStreamReferences.count < 0 {
		namedInputStreamReferences.count = 0
	}
	namedInputStreamReferences.count++
	namedInputStreamReferences.Unlock()
	return &NamedInputStream{File: f, fileName: file, moduleName: module, inputFile: input}, nil
}
func (s *NamedInputStream) GetName() string       { return s.fileName }
func (s *NamedInputStream) GetFileName() string   { return s.fileName }
func (s *NamedInputStream) GetModuleName() string { return s.moduleName }
func (s *NamedInputStream) SourceFile() string    { return s.inputFile }
func (s *NamedInputStream) GetAbsoluteResolvedPath() (string, error) {
	path, err := filepath.EvalSymlinks(s.inputFile)
	if err != nil {
		return "", bufferedRandomAccessFileIOError(err)
	}
	return filepath.Abs(path)
}
func (s *NamedInputStream) String() string {
	return "[ fileName: " + s.fileName + ", moduleName: " + s.moduleName + " ]"
}
func (s *NamedInputStream) Close() error {
	s.closeMu.Lock()
	defer s.closeMu.Unlock()
	namedInputStreamReferences.Lock()
	namedInputStreamReferences.count--
	namedInputStreamReferences.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	return bufferedRandomAccessFileIOError(s.File.Close())
}
func NamedInputStreamNumberOfReferences() int {
	namedInputStreamReferences.Lock()
	defer namedInputStreamReferences.Unlock()
	return namedInputStreamReferences.count
}

// MonolithModule retains real file-backed NamedInputStream metadata, including
// the original temporary basename and delete-on-exit registration.
func MonolithModule(inputFile, moduleName string) (*NamedInputStream, error) {
	safeName := distributedJavaFileName(moduleName) + ".tla"
	parent, err := os.MkdirTemp(Globals.MetaDir, "")
	if err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	}
	output := filepath.Join(parent, safeName)
	registerDistributedDeleteOnExit(parent)
	registerDistributedDeleteOnExit(output)
	writer, err := os.Create(output)
	if err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	}
	data, err := os.ReadFile(inputFile)
	if err != nil {
		_ = writer.Close()
		return nil, bufferedRandomAccessFileIOError(err)
	}
	source, err := javaCharsetDecode(data, javaDefaultCharset(), true)
	if err != nil {
		_ = writer.Close()
		return nil, err
	}
	module, found := ExtractMonolithModuleSource(source, moduleName)
	if os.PathSeparator == '\\' {
		module = strings.ReplaceAll(module, "\n", "\r\n")
	}
	encoded, err := javaCharsetEncode(module, javaDefaultCharset(), false)
	if err != nil {
		_ = writer.Close()
		return nil, err
	}
	// Source PrintWriter suppresses write/close IOExceptions.
	_, _ = writer.Write(encoded)
	_ = writer.Close()
	if !found {
		return nil, nil
	}
	return NewNamedInputStream(filepath.Base(output), moduleName, output)
}
