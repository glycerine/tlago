/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
// Go port of the bundled Apache Geronimo Activation DataHandler DataSource
// constructor, metadata, stream and writeTo paths; FileDataSource constructors
// and metadata. Object/URL/command-map/AWT content-handler providers remain
// separate required features. Source: https://github.com/apache/geronimo-specs/tree/trunk/geronimo-activation_1.1_spec/src/main/java/javax/activation
package tlc

import (
	"errors"
	"io"
	"os"
	"syscall"
)

type MailDataHandler struct {
	source         *MailDataSource
	flavorMIMEType *string
	WriteToFunc    func(io.Writer) error
}

func NewMailDataHandler(source *MailDataSource) (*MailDataHandler, error) {
	return mailInvoke(func() (*MailDataHandler, error) {
		return &MailDataHandler{source: source, flavorMIMEType: copyJavaMessage(source.GetContentType())}, nil
	})
}
func (dh *MailDataHandler) GetDataSource() *MailDataSource {
	if dh == nil {
		panic(NewNullPointerException())
	}
	return dh.source
}
func (dh *MailDataHandler) GetName() *string        { return dh.GetDataSource().GetName() }
func (dh *MailDataHandler) GetContentType() *string { return dh.GetDataSource().GetContentType() }
func (dh *MailDataHandler) GetInputStream() (*MailInputStream, error) {
	return dh.GetDataSource().GetInputStream()
}
func (dh *MailDataHandler) GetOutputStream() (io.WriteCloser, error) {
	return dh.GetDataSource().GetOutputStream()
}
func (dh *MailDataHandler) WriteTo(out io.Writer) error {
	_, err := mailInvoke(func() (bool, error) {
		if dh.WriteToFunc != nil {
			return false, dh.WriteToFunc(out)
		}
		buffer := make([]byte, 1024)
		stream, e := dh.GetInputStream()
		if e != nil {
			return false, e
		}
		_, bodyError := mailInvoke(func() (bool, error) {
			for {
				n, e := stream.ReadJava(buffer, 0, 1024)
				if e != nil {
					return false, e
				}
				if n == -1 {
					return true, nil
				}
				if n < 0 || n > len(buffer) {
					return false, mailReadBounds(0, int32(n), len(buffer))
				}
				if out == nil {
					return false, NewNullPointerException()
				}
				written, e := out.Write(buffer[:n])
				if e != nil {
					return false, e
				}
				if written != n {
					return false, io.ErrShortWrite
				}
			}
		})
		_, closeError := mailInvoke(func() (bool, error) { return false, stream.Close() })
		if closeError != nil {
			return false, closeError
		}
		return false, bodyError
	})
	return err
}
func NewMailFileDataSource(file *TLAFile) *MailDataSource {
	return &MailDataSource{file: file, isFile: true}
}
func NewMailFilePathDataSource(path string) *MailDataSource {
	return NewMailFileDataSource(&TLAFile{path: filenameNormalizeFile(path)})
}
func (ds *MailDataSource) GetFile() *TLAFile                          { return ds.file }
func (ds *MailDataSource) SetFileTypeMap(m *MailMimetypesFileTypeMap) { ds.fileTypeMap = m }

// FileInputStream rejects directories when opening, before any read occurs.
func mailOpenFileInputStream(path string) (*MailInputStream, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, distributedFileOpenException(path, e)
	}
	if info, x := f.Stat(); x == nil && info.IsDir() {
		_ = f.Close()
		return nil, distributedFileOpenException(path, syscall.EISDIR)
	}
	stream := &mailFileStream{filenameOSStream: filenameOSStream{File: f}}
	return NewMailInputStream(stream, stream), nil
}

// Native FileInputStream/FileOutputStream keep Java I/O throwable carriers and
// close idempotently. OS-specific errors use the existing file-message adapter.
type mailFileStream struct{ filenameOSStream }

func mailFileIOException(e error) error {
	if e == nil || e == io.EOF {
		return e
	}
	if errors.Is(e, os.ErrClosed) {
		return NewIOException("Stream Closed")
	}
	return NewIOException(distributedIOMessage(e))
}
func (s *mailFileStream) Read(b []byte) (int, error) {
	n, e := s.File.Read(b)
	return n, mailFileIOException(e)
}
func (s *mailFileStream) Write(b []byte) (int, error) {
	n, e := s.File.Write(b)
	return n, mailFileIOException(e)
}
func (s *mailFileStream) Stat() (os.FileInfo, error) {
	n, e := s.File.Stat()
	return n, mailFileIOException(e)
}
func (s *mailFileStream) Seek(off int64, whence int) (int64, error) {
	n, e := s.File.Seek(off, whence)
	return n, mailFileIOException(e)
}
func (s *mailFileStream) Close() error { return mailFileIOException(s.filenameOSStream.Close()) }
