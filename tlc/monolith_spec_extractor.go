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
	active := false
	var out strings.Builder
	for _, line := range monolithLines(source) {
		if active && monolithEnd(line) {
			out.WriteString(line)
			out.WriteByte('\n')
			break
		}
		if !active && isMonolithModuleStart(line, moduleName) {
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

func isMonolithModuleStart(line, moduleName string) bool {
	pattern := `^-{4,}` + monolithWhitespace + `*MODULE` + monolithWhitespace + `+` + regexp.QuoteMeta(moduleName) + monolithWhitespace + `*-{3,}$`
	return regexp.MustCompile(pattern).MatchString(line)
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
	f, err := openFileInputStream(input)
	if err != nil {
		return nil, err
	}
	namedInputStreamReferences.Lock()
	if namedInputStreamReferences.count < 0 {
		namedInputStreamReferences.count = 0
	}
	namedInputStreamReferences.count++
	namedInputStreamReferences.Unlock()
	return &NamedInputStream{File: f, fileName: file, moduleName: module, inputFile: input}, nil
}
func (s *NamedInputStream) require() *NamedInputStream {
	if s == nil {
		panic(NewNullPointerException())
	}
	return s
}

func (s *NamedInputStream) GetName() string       { return s.require().fileName }
func (s *NamedInputStream) GetFileName() string   { return s.require().fileName }
func (s *NamedInputStream) GetModuleName() string { return s.require().moduleName }
func (s *NamedInputStream) SourceFile() string    { return s.require().inputFile }
func (s *NamedInputStream) GetAbsoluteResolvedPath() (string, error) {
	s.require()
	path, err := filepath.EvalSymlinks(s.inputFile)
	if err != nil {
		return "", bufferedRandomAccessFileIOError(err)
	}
	return filepath.Abs(path)
}
func (s *NamedInputStream) String() string {
	s.require()
	return "[ fileName: " + s.fileName + ", moduleName: " + s.moduleName + " ]"
}
func (s *NamedInputStream) Close() error {
	s.require()
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
func MonolithModule(inputFile, moduleName string) (result *NamedInputStream, err error) {
	safeName := distributedJavaFileName(moduleName) + ".tla"
	parent, err := os.MkdirTemp(Globals.MetaDir, "")
	if err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	}
	output := filepath.Join(parent, safeName)
	registerDistributedDeleteOnExit(parent)
	registerDistributedDeleteOnExit(output)
	writer, err := openFileOutputStream(output)
	if err != nil {
		return nil, err
	}
	input, err := openFileInputStream(inputFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		failure := recover()
		closeErr := bufferedRandomAccessFileIOError(input.Close())
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if err == nil {
			if closeErr != nil {
				result = nil
				err = closeErr
			}
		} else {
			javaSuppressCloseError(err, closeErr)
		}
	}()
	reader, err := newJavaReplacingLineReader(input, javaDefaultCharset())
	if err != nil {
		return nil, err
	}
	// PrintWriter(FileWriter) adds its own character buffer around FileWriter's
	// default-charset encoder and suppresses write/flush/close IOExceptions.
	printWriter := &statePrintWriter{buffer: NewJavaBufferedWriter(writer)}
	found := false
	for {
		line, readErr := reader.ReadLine()
		if readErr != nil {
			return nil, bufferedRandomAccessFileIOError(readErr)
		}
		if line == nil {
			break
		}
		if found && monolithEnd(*line) {
			_, _ = printWriter.WriteString(*line + javaFormatLineSeparator())
			break
		}
		if !found && isMonolithModuleStart(*line, moduleName) {
			found = true
			_, _ = printWriter.WriteString(*line + javaFormatLineSeparator())
			continue
		}
		if found {
			_, _ = printWriter.WriteString(*line + javaFormatLineSeparator())
		}
	}
	printWriter.Flush()
	printWriter.closed = true
	_ = writer.Close()
	if !found {
		return nil, nil
	}
	return NewNamedInputStream(filepath.Base(output), moduleName, output)
}
