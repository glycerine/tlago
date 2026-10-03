// java.nio.file exceptions used by core TLC file I/O. OpenJDK's GPLv2 with
// Classpath Exception notices are retained in tlc/licenses/.
package tlc

type FileSystemException struct {
	*IOException
	file, other *string
}

func NewFileSystemException(file *string) *FileSystemException {
	return NewFileSystemExceptionWithReason(file, nil, nil)
}

func NewFileSystemExceptionWithReason(file, other, reason *string) *FileSystemException {
	return &FileSystemException{
		IOException: &IOException{javaExceptionBase: newJavaExceptionBase(reason, nil)},
		file:        copyJavaMessage(file), other: copyJavaMessage(other),
	}
}

func (e *FileSystemException) GetFile() *string      { return copyJavaMessage(e.file) }
func (e *FileSystemException) GetOtherFile() *string { return copyJavaMessage(e.other) }
func (e *FileSystemException) GetReason() *string    { return e.IOException.GetMessage() }
func (e *FileSystemException) GetMessage() *string {
	if e.file == nil && e.other == nil {
		return e.GetReason()
	}
	message := ""
	if e.file != nil {
		message = *e.file
	}
	if e.other != nil {
		message += " -> " + *e.other
	}
	if reason := e.GetReason(); reason != nil {
		message += ": " + *reason
	}
	return javaString(message)
}
func (e *FileSystemException) Error() string { return javaThrowableMessage(e) }

type NoSuchFileException struct{ *FileSystemException }

func (e *NoSuchFileException) Error() string { return javaThrowableMessage(e) }

func NewNoSuchFileException(file *string) *NoSuchFileException {
	return &NoSuchFileException{NewFileSystemException(file)}
}
func NewNoSuchFileExceptionWithReason(file, other, reason *string) *NoSuchFileException {
	return &NoSuchFileException{NewFileSystemExceptionWithReason(file, other, reason)}
}

type AccessDeniedException struct{ *FileSystemException }

func (e *AccessDeniedException) Error() string { return javaThrowableMessage(e) }

func NewAccessDeniedException(file *string) *AccessDeniedException {
	return &AccessDeniedException{NewFileSystemException(file)}
}
func NewAccessDeniedExceptionWithReason(file, other, reason *string) *AccessDeniedException {
	return &AccessDeniedException{NewFileSystemExceptionWithReason(file, other, reason)}
}

type FileAlreadyExistsException struct{ *FileSystemException }

func (e *FileAlreadyExistsException) Error() string { return javaThrowableMessage(e) }

func NewFileAlreadyExistsException(file *string) *FileAlreadyExistsException {
	return &FileAlreadyExistsException{NewFileSystemException(file)}
}
func NewFileAlreadyExistsExceptionWithReason(file, other, reason *string) *FileAlreadyExistsException {
	return &FileAlreadyExistsException{NewFileSystemExceptionWithReason(file, other, reason)}
}
