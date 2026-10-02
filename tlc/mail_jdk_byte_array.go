/*
 * Copyright (c) 1994, 2023, Oracle and/or its affiliates. All rights reserved.
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS FILE HEADER.
 *
 * This code is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License version 2 only, as
 * published by the Free Software Foundation.  Oracle designates this
 * particular file as subject to the "Classpath" exception as provided
 * by Oracle in the LICENSE file that accompanied this code.
 *
 * This code is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License
 * version 2 for more details (a copy is included in the LICENSE file that
 * accompanied this code).
 *
 * You should have received a copy of the GNU General Public License version
 * 2 along with this work; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin St, Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * Please contact Oracle, 500 Oracle Parkway, Redwood Shores, CA 94065 USA
 * or visit www.oracle.com if you need additional information or have any
 * questions.
 */
// ByteArrayInputStream behavior inherited by JavaMail SharedByteArrayInputStream.
// Source: https://github.com/openjdk/jdk21u/blob/master/src/java.base/share/classes/java/io/ByteArrayInputStream.java
package tlc

import (
	"fmt"
	"sync"
)

func mailReadBounds(off, n int32, size int) error {
	e := NewIndexOutOfBoundsException(0, size)
	e.Message = javaString(fmt.Sprintf("Range [%d, %d + %d) out of bounds for length %d", off, off, n, size))
	return e
}

// ByteArrayInputStream storage is shared, close is a no-op, and positions/counts
// retain Java's unchecked signed-int constructor arithmetic.
type MailSharedByteArrayInputStream struct {
	mu                           sync.Mutex
	buffer                       []byte
	position, mark, count, start int32
}

func newMailByteArrayInputStream(data []byte, bounds ...int32) *MailSharedByteArrayInputStream {
	if data == nil {
		panic(NewNullPointerException())
	}
	s := &MailSharedByteArrayInputStream{buffer: data, count: int32(len(data))}
	if len(bounds) > 0 {
		offset, length := bounds[0], bounds[1]
		s.position = offset
		s.mark = offset
		s.start = offset
		s.count = offset + length
		if s.count > int32(len(data)) {
			s.count = int32(len(data))
		}
	}
	return s
}
func (s *MailSharedByteArrayInputStream) ReadJava(b []byte, off, length int32) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if b == nil {
		return 0, NewNullPointerException()
	}
	if off < 0 || length < 0 || int64(off)+int64(length) > int64(len(b)) {
		return 0, mailReadBounds(off, length, len(b))
	}
	if s.position >= s.count {
		return -1, nil
	}
	available := s.count - s.position
	if length > available {
		length = available
	}
	if length <= 0 {
		return 0, nil
	}
	if s.position < 0 || int64(s.position)+int64(length) > int64(len(s.buffer)) {
		e := NewArrayIndexOutOfBoundsException(int(s.position), len(s.buffer))
		e.Message = javaString(fmt.Sprintf("arraycopy: source index %d out of bounds for byte[%d]", s.position, len(s.buffer)))
		return 0, e
	}
	copy(b[int(off):int(off+length)], s.buffer[int(s.position):int(s.position+length)])
	s.position += length
	return int(length), nil
}
func (s *MailSharedByteArrayInputStream) ReadByteJava() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.position >= s.count {
		return -1, nil
	}
	p := s.position
	s.position++
	if p < 0 || int(p) >= len(s.buffer) {
		return 0, NewArrayIndexOutOfBoundsException(int(p), len(s.buffer))
	}
	return int(s.buffer[p]), nil
}
func (s *MailSharedByteArrayInputStream) Available() int32 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.count - s.position
}
func (s *MailSharedByteArrayInputStream) Skip(n int64) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	k := int64(s.count - s.position)
	if n < k {
		if n < 0 {
			k = 0
		} else {
			k = n
		}
	}
	s.position += int32(k)
	return k
}
func (s *MailSharedByteArrayInputStream) Mark(readLimit int32) {
	s.mu.Lock()
	s.mark = s.position
	s.mu.Unlock()
}
func (s *MailSharedByteArrayInputStream) Reset()              { s.mu.Lock(); s.position = s.mark; s.mu.Unlock() }
func (s *MailSharedByteArrayInputStream) MarkSupported() bool { return true }
func (s *MailSharedByteArrayInputStream) Close() error        { return nil }
