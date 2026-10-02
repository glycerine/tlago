/*
 * Copyright (c) 1997, 2018 Oracle and/or its affiliates. All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0, which is available at
 * http://www.eclipse.org/legal/epl-2.0.
 *
 * This Source Code may also be made available under the following Secondary
 * Licenses when the conditions for such availability set forth in the
 * Eclipse Public License v. 2.0 are satisfied: GNU General Public License,
 * version 2 with the GNU Classpath Exception, which is available at
 * https://www.gnu.org/software/classpath/license.html.
 *
 * SPDX-License-Identifier: EPL-2.0 OR GPL-2.0 WITH Classpath-exception-2.0
 */
// Java InputStream boundary for JavaMail. ReadJava returns -1 at EOF; the
// io.Reader adapter delays a simultaneous Go read error until after its bytes.
package tlc

import (
	"io"
	"os"
)

type MailInputStream struct {
	ReadBytes      func([]byte, int32, int32) (int, error)
	AvailableBytes func() (int32, error)
	CloseStream    func() error
	Shared         *MailSharedByteArrayInputStream
}

func (s *MailInputStream) ReadJava(b []byte, off, n int32) (int, error) {
	if s == nil {
		return 0, NewNullPointerException()
	}
	return s.ReadBytes(b, off, n)
}

// The base Java InputStream.available implementation returns zero.
func (s *MailInputStream) AvailableJava() (int32, error) {
	if s == nil {
		return 0, NewNullPointerException()
	}
	if s.AvailableBytes != nil {
		return s.AvailableBytes()
	}
	return 0, nil
}
func (s *MailInputStream) Read(b []byte) (int, error) {
	n, e := s.ReadJava(b, 0, int32(len(b)))
	if n == -1 && e == nil {
		return 0, io.EOF
	}
	return n, e
}
func (s *MailInputStream) Close() error {
	if s == nil {
		return NewNullPointerException()
	}
	if s.CloseStream != nil {
		return s.CloseStream()
	}
	return nil
}
func NewMailInputStream(reader io.Reader, closer io.Closer) *MailInputStream {
	var pending error
	s := &MailInputStream{}
	if r, ok := reader.(interface{ Len() int }); ok {
		s.AvailableBytes = func() (int32, error) { return int32(r.Len()), nil }
	} else if r, ok := reader.(interface {
		Stat() (os.FileInfo, error)
		Seek(int64, int) (int64, error)
	}); ok {
		s.AvailableBytes = func() (int32, error) {
			info, e := r.Stat()
			if e != nil {
				return 0, e
			}
			if !info.Mode().IsRegular() {
				return 0, NewUnsupportedOperationException("FileInputStream.available provider required for nonregular file")
			}
			pos, e := r.Seek(0, io.SeekCurrent)
			if e != nil {
				return 0, e
			}
			n := info.Size() - pos
			if n < 0 {
				n = 0
			}
			if n > 2147483647 {
				n = 2147483647
			}
			return int32(n), nil
		}
	}
	s.ReadBytes = func(b []byte, off, length int32) (int, error) {
		if b == nil {
			return 0, NewNullPointerException()
		}
		if off < 0 || length < 0 || int64(off)+int64(length) > int64(len(b)) {
			return 0, mailReadBounds(off, length, len(b))
		}
		if length == 0 {
			return 0, nil
		}
		if pending != nil {
			e := pending
			pending = nil
			if e == io.EOF {
				return -1, nil
			}
			return 0, e
		}
		n, e := reader.Read(b[int(off):int(off+length)])
		if n > 0 {
			pending = e
			return n, nil
		}
		if e == io.EOF {
			return -1, nil
		}
		return n, e
	}
	if closer != nil {
		s.CloseStream = closer.Close
	}
	return s
}

// JavaMail subclasses ByteArrayInputStream without changing its read/mark/skip
// behavior. The inherited OpenJDK implementation is in mail_jdk_byte_array.go.
func NewMailSharedByteArrayInputStream(data []byte, bounds ...int32) *MailSharedByteArrayInputStream {
	return newMailByteArrayInputStream(data, bounds...)
}
func (s *MailSharedByteArrayInputStream) InputStream() *MailInputStream {
	return &MailInputStream{ReadBytes: s.ReadJava, AvailableBytes: func() (int32, error) { return s.Available(), nil }, Shared: s}
}
func (s *MailSharedByteArrayInputStream) GetPosition() int64 { return int64(s.position - s.start) }
func (s *MailSharedByteArrayInputStream) NewStream(start, end int64) *MailSharedByteArrayInputStream {
	if start < 0 {
		panic(NewIllegalArgumentException("start < 0"))
	}
	if end == -1 {
		end = int64(s.count - s.start)
	}
	return NewMailSharedByteArrayInputStream(s.buffer, s.start+int32(start), int32(end-start))
}
