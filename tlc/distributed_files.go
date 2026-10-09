package tlc

import (
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"unicode/utf16"
)

// DistributedFileServer supplies coordinator files to workers through either
// a local coordinator or the native Go network endpoint.
type DistributedFileServer interface {
	GetFile(string) ([]byte, error)
}

// DistributedFilenameToStreamResolver fetches coordinator files into a private
// directory and caches the local files. It preserves the upstream worker
// resolver's behavior independently of transport. Constructor library paths
// and the resolve isModule flag do not affect this resolver.
type DistributedFilenameToStreamResolver struct {
	server    DistributedFileServer
	fileCache *InsMap[string, *TLAFile]
	tmpDir    string
}

func NewDistributedFilenameToStreamResolver(libraryPaths ...[]string) *DistributedFilenameToStreamResolver {
	dir, err := os.MkdirTemp("", "tlc-")
	if err != nil {
		printDistributedFileException(NewIOException(distributedIOMessage(err)))
		panic(NewNullPointerException()) // newExclusiveTemporaryDirectory dereferences null.
	}
	registerDistributedDeleteOnExit(dir)
	return &DistributedFilenameToStreamResolver{fileCache: NewInsMap[string, *TLAFile](), tmpDir: dir}
}

func (r *DistributedFilenameToStreamResolver) SetTLCServer(server DistributedFileServer) {
	r.server = server
}

func (r *DistributedFilenameToStreamResolver) Resolve(filename string, isModule bool) *TLAFile {
	name := distributedJavaFileName(filename)
	file, found := r.fileCache.Get2(name)
	if !found || !file.Exists() {
		bs := r.fetch(name)
		file = r.writeToNewTempFile(name, bs)
		r.fileCache.Set(name, file)
	}
	return file
}

func (r *DistributedFilenameToStreamResolver) fetch(name string) (bs []byte) {
	bs = []byte{} // Java's initial byte[0] is not null.
	defer func() {
		if failure := recover(); failure != nil {
			if err, ok := failure.(error); ok && isDistributedRemoteFailure(err) {
				printDistributedFileException(err)
				return
			}
			panic(failure)
		}
	}()
	if r.server == nil {
		panic(NewNullPointerException())
	}
	data, err := r.server.GetFile(name)
	if err != nil {
		panic(err)
	}
	return data
}

func (r *DistributedFilenameToStreamResolver) writeToNewTempFile(name string, bs []byte) *TLAFile {
	file := NewTLAFile(filenamePathResolve(r.tmpDir, name), false, r)
	registerDistributedDeleteOnExit(file.GetPath())
	out, err := os.Create(file.GetPath())
	if err != nil {
		printDistributedFileException(distributedFileOpenException(file.GetPath(), err))
		return file
	}
	defer func() {
		if err := out.Close(); err != nil {
			printDistributedFileException(NewIOException(distributedIOMessage(err)))
		}
	}()
	// FileOutputStream.write(null) throws after the file has been created.
	if bs == nil {
		panic(NewNullPointerException())
	}
	if _, err := out.Write(bs); err != nil {
		printDistributedFileException(NewIOException(distributedIOMessage(err)))
	}
	return file
}

func (r *DistributedFilenameToStreamResolver) IsStandardModule(moduleName string) bool { return false }

func (r *DistributedFilenameToStreamResolver) GetFullPath() string {
	var buf strings.Builder
	i := 0
	for name := range r.fileCache.All() {
		buf.WriteString(name)
		// Java compares against this key's UTF-16 length, not the map size.
		if i < len(utf16.Encode([]rune(name)))-1 {
			buf.WriteString(", ")
		}
		i++
	}
	return buf.String()
}

// DistributedServerFiles supplies SimpleFilenameToStream's filename-only
// search used by TLCServer.getFile. The superclass's one-argument resolve calls
// the virtual two-argument method, so InJar's /model/ resources take precedence.
type DistributedServerFiles struct {
	UserDirectory  string
	LibraryPaths   []string
	ModelResources fs.FS
	Resources      fs.FS
	ResourcePrefix string
	Classpath      []FilenameClasspathEntry // Non-null supplies the complete class-loader roots.
}

func NewDistributedServerFiles(userDirectory string, libraryPaths []string, resources fs.FS, resourcePrefix string) *DistributedServerFiles {
	var captured []string
	if libraryPaths != nil {
		captured = append([]string{}, libraryPaths...)
	}
	return &DistributedServerFiles{UserDirectory: userDirectory, LibraryPaths: captured, ModelResources: resources, Resources: resources, ResourcePrefix: resourcePrefix}
}

func (s *DistributedServerFiles) resolve(name string) string {
	// TLCServer.getFile constructs InJarFilenameToStream for each call.
	// Only explicit native search overrides are captured; defaults and resource
	// temporary ownership belong to this request's fresh resolver.
	options := FilenameResolverOptions{Classpath: s.filenameClasspath()}
	if s.UserDirectory != "" {
		options.UserDirectory = &s.UserDirectory
	}
	resolver := NewSimpleFilenameToStream(s.LibraryPaths, options)
	if resolver.tmpDir != nil {
		// Own each request's temporary directory as well as copied files.
		registerDistributedDeleteOnExit(*resolver.tmpDir)
	}
	prefix := "/model/"
	resolver.modelPrefix = &prefix
	return resolver.Resolve(name, false).GetPath()
}

func (s *DistributedServerFiles) filenameClasspath() []FilenameClasspathEntry {
	if s.Classpath != nil {
		return append([]FilenameClasspathEntry{}, s.Classpath...)
	}
	entries := []FilenameClasspathEntry{}
	if s.ModelResources != nil {
		entries = append(entries, FilenameClasspathEntry{Files: s.ModelResources})
	}
	if s.Resources != nil {
		if resources, err := fs.Sub(s.Resources, strings.Trim(s.ResourcePrefix, "/")); s.ResourcePrefix != "" && err == nil {
			entries = append(entries, FilenameClasspathEntry{Files: resources, Prefix: StandardModulesClasspath, Location: "embedded:/" + s.ResourcePrefix})
		}
		entries = append(entries, FilenameClasspathEntry{Files: s.Resources})
	}
	return append(entries, filenameDefaultClasspath()...)
}

func (s *TLCServer) GetFile(file string) ([]byte, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	files := s.Files
	if files == nil {
		if s.Tool != nil {
			files = s.Tool.DistributedFiles
		}
		if files == nil {
			files = NewDistributedServerFiles("", nil, nil, "")
		}
	}
	return readDistributedServerFile(files.resolve(distributedJavaFileName(file))), nil
}

func readDistributedServerFile(file string) []byte {
	abs, err := filepath.Abs(file)
	if err != nil {
		abs = file
	}
	var size int64
	if info, err := os.Stat(file); err == nil {
		if info.IsDir() {
			panic(NewRuntimeException("Unsupported operation, file " + abs + " is a directory"))
		}
		size = info.Size()
	}
	if size > math.MaxInt32 {
		panic(NewRuntimeException("Unsupported operation, file " + abs + " is too big"))
	}
	buffer := make([]byte, int(size))
	in, err := os.Open(file)
	var pending error
	if err != nil {
		pending = NewRuntimeExceptionWithCause(javaString("Exception occured on reading file "+abs), distributedFileOpenException(file, err))
	} else {
		// Java performs a single read and ignores its result, retaining zeros
		// if the file shrinks. EOF is a return value, not an IOException.
		if _, err := in.Read(buffer); err != nil && err != io.EOF {
			pending = NewRuntimeExceptionWithCause(javaString("Exception occured on reading file "+abs), NewIOException(distributedIOMessage(err)))
		}
		if err := in.Close(); err != nil && pending == nil {
			pending = NewRuntimeExceptionWithCause(javaString("Exception occured on closing file"+abs), NewIOException(distributedIOMessage(err)))
		}
	}
	if pending != nil {
		panic(NewRuntimeExceptionFromCause(pending))
	}
	return buffer
}

func distributedJavaFileName(path string) string {
	// File normalizes repeated/trailing native separators, but retains . and ..
	// components. filepath.Base("") and filepath.Base("/") differ from Java.
	if runtime.GOOS == "windows" {
		path = strings.ReplaceAll(path, "/", "\\")
		if len(path) >= 2 && path[1] == ':' {
			path = path[2:]
		}
	}
	path = strings.TrimRight(path, string(filepath.Separator))
	if i := strings.LastIndexByte(path, byte(filepath.Separator)); i >= 0 {
		return path[i+1:]
	}
	return path
}

func distributedIOMessage(err error) string {
	if failure, ok := err.(*os.PathError); ok {
		err = failure.Err
	}
	message := err.Error()
	if len(message) > 0 {
		message = strings.ToUpper(message[:1]) + message[1:]
	}
	return message
}

func distributedFileOpenException(file string, err error) *FileNotFoundException {
	return NewFileNotFoundException(file + " (" + distributedIOMessage(err) + ")")
}

func printDistributedFileException(err error) { fmt.Fprint(os.Stderr, javaThrowableStackTrace(err)) }

var distributedDeleteOnExit struct {
	sync.Mutex
	paths *InsMap[string, bool]
}

func registerDistributedDeleteOnExit(path string) {
	distributedDeleteOnExit.Lock()
	defer distributedDeleteOnExit.Unlock()
	if distributedDeleteOnExit.paths == nil {
		distributedDeleteOnExit.paths = NewInsMap[string, bool]()
	}
	distributedDeleteOnExit.paths.Set(path, true)
}

// CleanupDistributedFiles is the process-exit hook, not a worker-exit hook.
// Like Java DeleteHook, deletion is non-recursive and reverses registration.
func CleanupDistributedFiles() {
	distributedDeleteOnExit.Lock()
	defer distributedDeleteOnExit.Unlock()
	if distributedDeleteOnExit.paths == nil {
		return
	}
	var paths []string
	for path := range distributedDeleteOnExit.paths.All() {
		paths = append(paths, path)
	}
	for i := len(paths) - 1; i >= 0; i-- {
		_ = os.Remove(paths[i])
	}
	distributedDeleteOnExit.paths.DeleteAll()
}
