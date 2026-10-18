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

func (e *FileSystemException) GetFile() *string {
	e.require()
	return copyJavaMessage(e.file)
}
func (e *FileSystemException) GetOtherFile() *string {
	e.require()
	return copyJavaMessage(e.other)
}
func (e *FileSystemException) GetReason() *string {
	e.require()
	return e.IOException.GetMessage()
}
func (e *FileSystemException) GetMessage() *string {
	e.require()
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
func (e *FileSystemException) GetCause() error {
	e.require()
	return e.IOException.GetCause()
}
func (e *FileSystemException) Error() string { return javaThrowableMessage(e) }
func (e *FileSystemException) require() {
	if e == nil || e.IOException == nil {
		panic(NewNullPointerException())
	}
}

type NoSuchFileException struct{ *FileSystemException }

func (e *NoSuchFileException) Error() string { return javaThrowableMessage(e) }
func (e *NoSuchFileException) GetFile() *string {
	e.require()
	return e.FileSystemException.GetFile()
}
func (e *NoSuchFileException) GetOtherFile() *string {
	e.require()
	return e.FileSystemException.GetOtherFile()
}
func (e *NoSuchFileException) GetReason() *string {
	e.require()
	return e.FileSystemException.GetReason()
}
func (e *NoSuchFileException) GetMessage() *string {
	e.require()
	return e.FileSystemException.GetMessage()
}
func (e *NoSuchFileException) GetCause() error {
	e.require()
	return e.FileSystemException.GetCause()
}
func (e *NoSuchFileException) require() {
	if e == nil || e.FileSystemException == nil {
		panic(NewNullPointerException())
	}
}

func NewNoSuchFileException(file *string) *NoSuchFileException {
	return &NoSuchFileException{NewFileSystemException(file)}
}
func NewNoSuchFileExceptionWithReason(file, other, reason *string) *NoSuchFileException {
	return &NoSuchFileException{NewFileSystemExceptionWithReason(file, other, reason)}
}

type AccessDeniedException struct{ *FileSystemException }

func (e *AccessDeniedException) Error() string { return javaThrowableMessage(e) }
func (e *AccessDeniedException) GetFile() *string {
	e.require()
	return e.FileSystemException.GetFile()
}
func (e *AccessDeniedException) GetOtherFile() *string {
	e.require()
	return e.FileSystemException.GetOtherFile()
}
func (e *AccessDeniedException) GetReason() *string {
	e.require()
	return e.FileSystemException.GetReason()
}
func (e *AccessDeniedException) GetMessage() *string {
	e.require()
	return e.FileSystemException.GetMessage()
}
func (e *AccessDeniedException) GetCause() error {
	e.require()
	return e.FileSystemException.GetCause()
}
func (e *AccessDeniedException) require() {
	if e == nil || e.FileSystemException == nil {
		panic(NewNullPointerException())
	}
}

func NewAccessDeniedException(file *string) *AccessDeniedException {
	return &AccessDeniedException{NewFileSystemException(file)}
}
func NewAccessDeniedExceptionWithReason(file, other, reason *string) *AccessDeniedException {
	return &AccessDeniedException{NewFileSystemExceptionWithReason(file, other, reason)}
}

type FileAlreadyExistsException struct{ *FileSystemException }

func (e *FileAlreadyExistsException) Error() string { return javaThrowableMessage(e) }
func (e *FileAlreadyExistsException) GetFile() *string {
	e.require()
	return e.FileSystemException.GetFile()
}
func (e *FileAlreadyExistsException) GetOtherFile() *string {
	e.require()
	return e.FileSystemException.GetOtherFile()
}
func (e *FileAlreadyExistsException) GetReason() *string {
	e.require()
	return e.FileSystemException.GetReason()
}
func (e *FileAlreadyExistsException) GetMessage() *string {
	e.require()
	return e.FileSystemException.GetMessage()
}
func (e *FileAlreadyExistsException) GetCause() error {
	e.require()
	return e.FileSystemException.GetCause()
}
func (e *FileAlreadyExistsException) require() {
	if e == nil || e.FileSystemException == nil {
		panic(NewNullPointerException())
	}
}

func NewFileAlreadyExistsException(file *string) *FileAlreadyExistsException {
	return &FileAlreadyExistsException{NewFileSystemException(file)}
}
func NewFileAlreadyExistsExceptionWithReason(file, other, reason *string) *FileAlreadyExistsException {
	return &FileAlreadyExistsException{NewFileSystemExceptionWithReason(file, other, reason)}
}

// DirectoryNotEmptyException is the replacement-copy/delete boundary for a
// directory that still contains entries.
type DirectoryNotEmptyException struct{ *FileSystemException }

func (e *DirectoryNotEmptyException) Error() string { return javaThrowableMessage(e) }
func (e *DirectoryNotEmptyException) GetFile() *string {
	e.require()
	return e.FileSystemException.GetFile()
}
func (e *DirectoryNotEmptyException) GetOtherFile() *string {
	e.require()
	return e.FileSystemException.GetOtherFile()
}
func (e *DirectoryNotEmptyException) GetReason() *string {
	e.require()
	return e.FileSystemException.GetReason()
}
func (e *DirectoryNotEmptyException) GetMessage() *string {
	e.require()
	return e.FileSystemException.GetMessage()
}
func (e *DirectoryNotEmptyException) GetCause() error {
	e.require()
	return e.FileSystemException.GetCause()
}
func (e *DirectoryNotEmptyException) require() {
	if e == nil || e.FileSystemException == nil {
		panic(NewNullPointerException())
	}
}

func NewDirectoryNotEmptyException(file *string) *DirectoryNotEmptyException {
	return &DirectoryNotEmptyException{NewFileSystemException(file)}
}
