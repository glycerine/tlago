package tlc

import (
	"archive/zip"
	"errors"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// FilenameClasspathEntry represents one ordered Java class-loader search root.
// Path names a directory or archive. Files supplies a bundled fs.FS mounted at
// Prefix; Location supplies its original URI (otherwise embedded:/ is used).
type FilenameClasspathEntry struct {
	Path     string
	Files    fs.FS
	Prefix   string
	Location string
}

type FilenameResolverOptions struct {
	UserDirectory *string
	Classpath     []FilenameClasspathEntry
}

type SimpleFilenameToStream struct {
	tmpDir      *string
	userPaths   []string
	userDir     *string
	classpath   []FilenameClasspathEntry
	modelPrefix *string
}

func NewSimpleFilenameToStream(libraryPaths []string, options ...FilenameResolverOptions) *SimpleFilenameToStream {
	resolver := &SimpleFilenameToStream{tmpDir: filenameTempDirectory()}
	if libraryPaths == nil {
		if paths, ok := tlcLookupSystemProperty(TLALibraryProperty); ok {
			libraryPaths = filenameSplitPaths(paths)
		}
	}
	var configuration FilenameResolverOptions
	if len(options) > 0 {
		configuration = options[0]
	}
	resolver.userDir = copyJavaMessage(configuration.UserDirectory)
	dir := configuration.UserDirectory
	if dir == nil {
		dir = GetFilenameUserDirectory()
	}
	if dir == nil {
		if cwd, ok := tlcLookupSystemProperty("user.dir"); ok {
			dir = &cwd
		} else if cwd, err := os.Getwd(); err == nil {
			dir = &cwd
		}
	}
	if dir != nil {
		filenameCheckPath(*dir)
		resolver.userPaths = append(resolver.userPaths, filenameNormalizeFile(*dir))
	}
	for _, path := range libraryPaths {
		filenameCheckPath(path)
		resolver.userPaths = append(resolver.userPaths, filenameNormalizeFile(path))
	}
	resolver.classpath = append([]FilenameClasspathEntry(nil), configuration.Classpath...)
	if configuration.Classpath == nil {
		resolver.classpath = filenameDefaultClasspath()
	}
	return resolver
}

// DefaultFilenameClasspath snapshots the process classpath in source order.
func DefaultFilenameClasspath() []FilenameClasspathEntry { return filenameDefaultClasspath() }

func filenameDefaultClasspath() []FilenameClasspathEntry {
	paths, ok := tlcLookupSystemProperty("java.class.path")
	if !ok {
		paths, ok = os.LookupEnv("CLASSPATH")
	}
	if !ok {
		paths = "."
	}
	var entries []FilenameClasspathEntry
	for _, path := range filenameSplitPaths(paths) {
		entries = append(entries, FilenameClasspathEntry{Path: path})
	}
	return entries
}

func NewInJarFilenameToStream(prefix string, options ...FilenameResolverOptions) *SimpleFilenameToStream {
	resolver := NewSimpleFilenameToStream(nil, options...)
	resolver.modelPrefix = &prefix
	return resolver
}

func filenameSplitPaths(paths string) []string {
	// String.split removes trailing empties, except that "" gives one empty path.
	parts := strings.Split(paths, string(os.PathListSeparator))
	if paths != "" {
		for len(parts) > 0 && parts[len(parts)-1] == "" {
			parts = parts[:len(parts)-1]
		}
	}
	return parts
}

func (r *SimpleFilenameToStream) GetFullPath() string {
	// There are two nested sequential user locators, so even an empty library
	// locator contributes its separator in the historical description.
	user, libraries := "", ""
	if len(r.userPaths) > 0 {
		user = r.userPaths[0]
		libraries = strings.Join(r.userPaths[1:], ", ")
	}
	return user + ", " + libraries + ", classpath:/" + StandardModulesClasspath + ", classpath:/"
}

func (r *SimpleFilenameToStream) IsStandardModule(name string) bool {
	file := r.Resolve(name, true)
	return file.Exists() && file.IsLibraryModule()
}

func (r *SimpleFilenameToStream) Resolve(name string, isModule bool) *TLAFile {
	if r.modelPrefix != nil {
		// Class.getResource resolves a relative prefix in package model.
		resource := *r.modelPrefix + name
		if strings.HasPrefix(resource, "/") {
			resource = resource[1:]
		} else {
			resource = "model/" + resource
		}
		if location := r.findClasspath(resource); location != nil {
			if file, err := r.copyResource(name, location, false); err == nil {
				return file
			} else if file != nil {
				printDistributedFileException(err)
			}
		}
	}
	name = filenameNormalizeModule(name, isModule)
	filenameCheckPath(name)
	for _, dir := range r.userPaths {
		path := filenamePathResolve(dir, name)
		if _, err := os.Stat(path); err == nil {
			return NewTLAFile(path, false, r)
		}
	}
	for _, prefix := range []string{StandardModulesClasspath + "/", ""} {
		if location := r.findClasspath(prefix + name); location != nil {
			if location.path != "" {
				return NewTLAFile(location.path, true, r)
			}
			file, err := r.copyResource(name, location, true)
			if err != nil {
				if file == nil {
					panic(NewRuntimeExceptionFromCause(err))
				}
				printDistributedFileException(err)
			}
			// Simple.read returns its attempted file even after a copy IOException.
			return file
		}
	}
	dir := r.userDir
	if dir == nil {
		dir = GetFilenameUserDirectory()
	}
	parent := "."
	if dir != nil {
		parent = *dir
	}
	return NewTLAFileInDirectory(parent, name, r)
}

func filenamePathResolve(parent, child string) string {
	if filepath.IsAbs(child) || parent == "" {
		return filenameNormalizeFile(child)
	}
	return filenameFileJoin(parent, child)
}

type filenameResource struct {
	path, location string
	files          fs.FS
	name, archive  string
}

func (r *SimpleFilenameToStream) findClasspath(name string) *filenameResource {
	if strings.HasPrefix(name, "/") {
		return nil
	}
	for _, entry := range r.classpath {
		if entry.Files != nil {
			prefix := strings.Trim(entry.Prefix, "/")
			resource := name
			if prefix != "" {
				if !strings.HasPrefix(resource, prefix+"/") {
					continue
				}
				resource = resource[len(prefix)+1:]
			}
			_, err := fs.Stat(entry.Files, resource)
			if err != nil {
				if os.IsNotExist(err) || errors.Is(err, fs.ErrInvalid) {
					continue
				}
				panic(NewRuntimeExceptionFromCause(NewIOException(distributedIOMessage(err))))
			}
			base := entry.Location
			if base == "" {
				base = "embedded:/" + prefix
			}
			return &filenameResource{location: strings.TrimRight(base, "/") + "/" + (&url.URL{Path: resource}).EscapedPath(), files: entry.Files, name: resource}
		}
		if entry.Path == "" {
			entry.Path = "." // An empty Java classpath component is the CWD.
		}
		info, err := os.Stat(entry.Path)
		if err != nil {
			continue
		}
		if info.IsDir() {
			path := filenamePathResolve(entry.Path, name)
			if _, err := os.Stat(path); err == nil {
				return &filenameResource{path: path, location: filenameFileURI(path)}
			}
			continue
		}
		archive, err := zip.OpenReader(entry.Path)
		if err != nil {
			continue
		} // ClassLoader ignores inaccessible/invalid archives.
		var found *filenameResource
		for _, file := range archive.File {
			if file.Name != name {
				continue
			}
			found = &filenameResource{location: "jar:" + filenameFileURI(entry.Path) + "!/" + (&url.URL{Path: name}).EscapedPath(), archive: entry.Path, name: name}
			break
		}
		_ = archive.Close()
		if found != nil {
			return found
		}
	}
	return nil
}

// filenameArchiveStream closes the zip file after its entry stream.
type filenameArchiveStream struct {
	io.ReadCloser
	archive *zip.ReadCloser
	closed  sync.Once
}

func (s *filenameArchiveStream) Close() error {
	var err error
	s.closed.Do(func() {
		err = s.ReadCloser.Close()
		if closeErr := s.archive.Close(); err == nil {
			err = closeErr
		}
	})
	return err
}

type filenameOSStream struct {
	*os.File
	closed sync.Once
}

func (s *filenameOSStream) Close() error {
	var err error
	s.closed.Do(func() { err = s.File.Close() })
	return err
}
func (resource *filenameResource) open() (io.ReadCloser, error) {
	if resource.path != "" {
		file, err := os.Open(resource.path)
		if err != nil {
			return nil, err
		}
		return &filenameOSStream{File: file}, nil
	}
	if resource.files != nil {
		return resource.files.Open(resource.name)
	}
	archive, err := zip.OpenReader(resource.archive)
	if err != nil {
		return nil, err
	}
	for _, file := range archive.File {
		if file.Name == resource.name {
			stream, err := file.Open()
			if err != nil {
				_ = archive.Close()
				return nil, err
			}
			return &filenameArchiveStream{ReadCloser: stream, archive: archive}, nil
		}
	}
	_ = archive.Close()
	return nil, fs.ErrNotExist
}
func (r *SimpleFilenameToStream) copyResource(name string, resource *filenameResource, library bool) (file *TLAFile, err error) {
	stream, err := resource.open()
	if err != nil {
		return nil, NewIOException(distributedIOMessage(err))
	}
	defer func() {
		if library && file != nil && err != nil {
			// Simple.read catches and prints before the outer resource close.
			printDistributedFileException(err)
			err = nil
		}
		if closeErr := stream.Close(); library && closeErr != nil {
			panic(NewRuntimeExceptionFromCause(NewIOException(distributedIOMessage(closeErr))))
		}
	}()
	if r.tmpDir == nil {
		panic(NewNullPointerException())
	}
	path := filenamePathResolve(*r.tmpDir, name)
	file = newResourceTLAFile(path, resource.location, library, r)
	registerDistributedDeleteOnExit(file.path)
	out, err := os.Create(path)
	if err != nil {
		return file, distributedFileOpenException(path, err)
	}
	defer out.Close()
	buffer := make([]byte, 1024)
	for {
		count, readErr := stream.Read(buffer)
		if count > 0 {
			if _, err := out.Write(buffer[:count]); err != nil {
				return file, NewIOException(distributedIOMessage(err))
			}
		}
		if readErr != nil && readErr != io.EOF {
			return file, NewIOException(distributedIOMessage(readErr))
		}
		if count <= 0 || readErr == io.EOF {
			break
		}
	}
	if err := out.Close(); err != nil {
		return file, NewIOException(distributedIOMessage(err))
	}
	if err := stream.Close(); err != nil {
		return file, NewIOException(distributedIOMessage(err))
	}
	return file, nil
}
