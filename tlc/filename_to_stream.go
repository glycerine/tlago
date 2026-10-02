package tlc

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
)

const TLALibraryProperty = "TLA-Library"
const StandardModulesClasspath = "tla2sany/StandardModules"

// Both local and remote resolvers are used by TLAFile's override lookup.
type FilenameToStream interface {
	Resolve(string, bool) *TLAFile
	GetFullPath() string
	IsStandardModule(string) bool
}

// TLAFile retains FilenameToStream.TLAFile's provenance, including the original
// resource URI for modules copied from an archive or embedded classpath.
type TLAFile struct {
	path        string
	library     bool
	libraryPath *string
	resolver    FilenameToStream
}

func NewTLAFile(path string, library bool, resolver FilenameToStream) *TLAFile {
	file := &TLAFile{path: filenameNormalizeFile(path), library: library, resolver: resolver}
	if file.Exists() {
		uri := filenameFileURI(file.path)
		file.libraryPath = &uri
	}
	return file
}

func NewTLAFileInDirectory(parent, child string, resolver FilenameToStream) *TLAFile {
	pattern := `^([a-zA-Z]+:)?/`
	if runtime.GOOS == "windows" {
		pattern = `^([a-zA-Z]+:)?\\`
	}
	if regexp.MustCompile(pattern).MatchString(parent) && strings.HasPrefix(child, parent) {
		child = child[len(parent):]
	}
	return NewTLAFile(filenameFileJoin(parent, child), false, resolver)
}

func newResourceTLAFile(path, location string, library bool, resolver FilenameToStream) *TLAFile {
	file := &TLAFile{path: filenameNormalizeFile(path), library: library, resolver: resolver}
	// URL.toURI rejects spaces/control characters rather than encoding them.
	if _, err := url.Parse(location); err == nil && !strings.ContainsAny(location, " \t\r\n") {
		file.libraryPath = &location
	}
	return file
}

func (f *TLAFile) String() string  { return f.path }
func (f *TLAFile) GetPath() string { return f.path }
func (f *TLAFile) GetName() string { return distributedJavaFileName(f.path) }
func (f *TLAFile) GetAbsolutePath() string {
	if filepath.IsAbs(f.path) {
		return f.path
	}
	cwd, err := os.Getwd()
	if err != nil {
		panic(NewRuntimeExceptionFromCause(NewIOException(distributedIOMessage(err))))
	}
	return filenameFileJoin(cwd, f.path)
}
func (f *TLAFile) Exists() bool            { _, err := os.Stat(f.path); return err == nil }
func (f *TLAFile) IsLibraryModule() bool   { return f.library }
func (f *TLAFile) HasLibraryPath() bool    { return f.libraryPath != nil }
func (f *TLAFile) GetLibraryPath() *string { return copyJavaMessage(f.libraryPath) }
func (f *TLAFile) GetModuleOverride() *TLAFile {
	if f.resolver == nil {
		panic(NewNullPointerException())
	}
	// Java's .tla$ pattern deliberately has a wildcard, not an escaped dot.
	name := regexp.MustCompile(`[^\n\r\x{0085}\x{2028}\x{2029}]tla$`).ReplaceAllString(f.GetName(), ".class")
	file := f.resolver.Resolve(name, false)
	if file.Exists() {
		return file
	}
	return nil
}

func FilenameIsInJar(path string) bool {
	return strings.HasPrefix(path, "jar:") || strings.HasSuffix(path, ".jar")
}
func FilenameIsArchive(path string) bool {
	return FilenameIsInJar(path) || strings.HasSuffix(path, ".zip")
}

func filenameTempDirectory() *string {
	dir, err := os.MkdirTemp("", "tlc-")
	if err != nil {
		printDistributedFileException(NewIOException(distributedIOMessage(err)))
		return nil
	}
	return &dir
}

var filenameUserDirectory struct {
	sync.Mutex
	value *string
}

func SetFilenameUserDirectory(path *string) {
	filenameUserDirectory.Lock()
	defer filenameUserDirectory.Unlock()
	filenameUserDirectory.value = copyJavaMessage(path)
}
func GetFilenameUserDirectory() *string {
	filenameUserDirectory.Lock()
	defer filenameUserDirectory.Unlock()
	return copyJavaMessage(filenameUserDirectory.value)
}

func filenameNormalizeFile(path string) string {
	unc := false
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, "/", `\`)
		unc = strings.HasPrefix(path, `\\`)
	}
	separator := string(filepath.Separator)
	for strings.Contains(path, separator+separator) {
		path = strings.ReplaceAll(path, separator+separator, separator)
	}
	if unc {
		path = separator + path
	}
	minimum := 1
	if runtime.GOOS == "windows" {
		minimum = len(filepath.VolumeName(path)) + 1
	}
	for len(path) > minimum && strings.HasSuffix(path, separator) {
		path = path[:len(path)-1]
	}
	return path
}
func filenameFileJoin(parent, child string) string {
	if parent == "" {
		parent = string(filepath.Separator)
	}
	if child == "" {
		return filenameNormalizeFile(parent)
	}
	// java.io.File(parent, child) retains the parent even for absolute children.
	return filenameNormalizeFile(parent + string(filepath.Separator) + child)
}
func filenameFileURI(path string) string {
	if !filepath.IsAbs(path) {
		cwd, err := os.Getwd()
		if err != nil {
			panic(NewRuntimeExceptionFromCause(err))
		}
		path = filenameFileJoin(cwd, path)
	}
	path = filepath.ToSlash(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if info, err := os.Stat(path); err == nil && info.IsDir() && !strings.HasSuffix(path, "/") {
		path += "/"
	}
	var escaped strings.Builder
	for _, char := range path {
		if char >= 128 {
			escaped.WriteRune(char)
		} else {
			escaped.WriteString((&url.URL{Path: string(char)}).EscapedPath())
		}
	}
	return "file:" + escaped.String()
}

func filenameCheckPath(path string) {
	if strings.ContainsRune(path, 0) {
		panic(NewInvalidPathException(path, "Nul character not allowed"))
	}
}

// ModuleFilename applies FileUtil.createNamedInputStream's logical filename
// normalization before resolution and parse-unit metadata are recorded.
func ModuleFilename(name string) string { return filenameNormalizeModule(name, true) }

func filenameNormalizeModule(name string, isModule bool) string {
	if n := strings.IndexByte(name, '\n'); n >= 0 {
		fmt.Fprintf(os.Stdout, "*** Warning: module name '%s' contained NEWLINE; Only the part before NEWLINE is considered.\n", name)
		name = name[:n]
	}
	if isModule {
		if strings.HasSuffix(strings.ToLower(name), ".tla") {
			name = name[:len(name)-4]
		}
		name += ".tla"
	}
	return name
}
