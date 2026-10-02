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

// DistributedFileServer is the getFile portion of TLCServerRMI. The local
// TLCServer implements the same contract used by a future network client.
type DistributedFileServer interface {
	GetFile(string) ([]byte, error)
}

// RMIFilenameToStreamResolver mirrors the worker's resolver, including its
// cache of files rather than bytes. Neither constructor library paths nor the
// resolve isModule flag affect this particular Java resolver.
type RMIFilenameToStreamResolver struct {
	server    DistributedFileServer
	fileCache *InsMap[string, string]
	tmpDir    string
}

func NewRMIFilenameToStreamResolver(libraryPaths ...[]string) *RMIFilenameToStreamResolver {
	dir, err := os.MkdirTemp("", "tlc-")
	if err != nil {
		printDistributedFileException(NewIOException(distributedIOMessage(err)))
		panic(NewNullPointerException()) // newExclusiveTemporaryDirectory dereferences null.
	}
	registerDistributedDeleteOnExit(dir)
	return &RMIFilenameToStreamResolver{fileCache: NewInsMap[string, string](), tmpDir: dir}
}

func (r *RMIFilenameToStreamResolver) SetTLCServer(server DistributedFileServer) { r.server = server }

func (r *RMIFilenameToStreamResolver) Resolve(filename string, isModule bool) string {
	name := distributedJavaFileName(filename)
	file, found := r.fileCache.Get2(name)
	if _, err := os.Stat(file); !found || err != nil {
		bs := r.fetch(name)
		file = r.writeToNewTempFile(name, bs)
		r.fileCache.Set(name, file)
	}
	return file
}

func (r *RMIFilenameToStreamResolver) fetch(name string) (bs []byte) {
	bs = []byte{} // Java's initial byte[0] is not null.
	defer func() {
		if failure := recover(); failure != nil {
			if err, ok := failure.(error); ok && javaRemoteException(err) != nil {
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

func (r *RMIFilenameToStreamResolver) writeToNewTempFile(name string, bs []byte) string {
	file := filepath.Join(r.tmpDir, name)
	registerDistributedDeleteOnExit(file)
	out, err := os.Create(file)
	if err != nil {
		printDistributedFileException(distributedFileOpenException(file, err))
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

func (r *RMIFilenameToStreamResolver) IsStandardModule(moduleName string) bool { return false }

func (r *RMIFilenameToStreamResolver) GetFullPath() string {
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
}

func NewDistributedServerFiles(userDirectory string, libraryPaths []string, resources fs.FS, resourcePrefix string) *DistributedServerFiles {
	if userDirectory == "" {
		userDirectory, _ = os.Getwd()
	}
	if libraryPaths == nil {
		if paths := os.Getenv("TLA-Library"); paths != "" {
			libraryPaths = filepath.SplitList(paths)
		}
	}
	return &DistributedServerFiles{UserDirectory: userDirectory, LibraryPaths: append([]string(nil), libraryPaths...), ModelResources: resources, Resources: resources, ResourcePrefix: resourcePrefix}
}

func (s *DistributedServerFiles) resolve(name string) string {
	var tmpDir string
	copyResource := func(data []byte) (string, error) {
		if tmpDir == "" {
			var err error
			tmpDir, err = os.MkdirTemp("", "tlc-")
			if err != nil {
				printDistributedFileException(NewIOException(distributedIOMessage(err)))
				panic(NewNullPointerException())
			}
		}
		// Simple/InJar schedule the copied file, not their temporary directory.
		file := filepath.Join(tmpDir, name)
		registerDistributedDeleteOnExit(file)
		return file, os.WriteFile(file, data, 0666)
	}
	if s.ModelResources != nil {
		// InJar tries the raw name before Simple strips a newline.
		if data, err := fs.ReadFile(s.ModelResources, "model/"+name); err == nil {
			file, err := copyResource(data)
			if err == nil {
				return file
			}
			// InJar's copy IOException prints and falls back to Simple.
			printDistributedFileException(NewIOException(distributedIOMessage(err)))
		}
	}
	if n := strings.IndexByte(name, '\n'); n >= 0 {
		fmt.Fprintf(os.Stdout, "*** Warning: module name '%s' contained NEWLINE; Only the part before NEWLINE is considered.\n", name)
		name = name[:n]
	}
	for _, dir := range append([]string{s.UserDirectory}, s.LibraryPaths...) {
		path := filepath.Join(dir, name)
		if filepath.IsAbs(name) {
			path = name
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	if s.Resources != nil {
		resource := strings.TrimSuffix(s.ResourcePrefix, "/") + "/" + name
		if s.ResourcePrefix == "" {
			resource = name
		}
		data, err := fs.ReadFile(s.Resources, resource)
		if os.IsNotExist(err) && s.ResourcePrefix != "" {
			// Java searches the bare classpath after StandardModules.
			data, err = fs.ReadFile(s.Resources, name)
		}
		if err == nil {
			file, err := copyResource(data)
			if err != nil {
				// Simple returns its file even when the copy failed.
				printDistributedFileException(NewIOException(distributedIOMessage(err)))
			}
			return file
		}
		if !os.IsNotExist(err) {
			panic(NewRuntimeExceptionFromCause(NewIOException(distributedIOMessage(err))))
		}
	}
	return filepath.Join(s.UserDirectory, name)
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
