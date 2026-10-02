/*
 * Copyright (c) 1996, 2023, Oracle and/or its affiliates. All rights reserved.
 * Copyright (c) 2001, 2023, Oracle and/or its affiliates. All rights reserved.
 * Copyright (c) 2000, 2021, Oracle and/or its affiliates. All rights reserved.
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
// Activation's unmarked BufferedReader.readLine and InputStream StreamDecoder
// paths, with the OpenJDK UTF-8/ASCII/Latin-1 replacement decoders. This adapter
// has separate 8192-byte and 8192-UTF-16-unit buffers, as in the Java source.
// Sources: OpenJDK jdk21u java.io.BufferedReader, sun.nio.cs.StreamDecoder/UTF_8.
package tlc

// mailDecodeReaderCharacter returns one decoded code point and its byte width.
// Width zero means that the decoder needs more input before end-of-input.
func mailDecodeReaderCharacter(b []byte, charset string, eof bool) (uint32, int) {
	c := b[0]
	if charset == "iso-8859-1" || c < 128 {
		return uint32(c), 1
	}
	if charset == "us-ascii" {
		return 0xfffd, 1
	}
	width := 0
	switch {
	case c >= 0xc2 && c <= 0xdf:
		width = 2
	case c >= 0xe0 && c <= 0xef:
		width = 3
	case c >= 0xf0 && c <= 0xf4:
		width = 4
	}
	if width == 0 {
		return 0xfffd, 1
	}
	n := 1
	for n < width && n < len(b) {
		next := b[n]
		if next&0xc0 != 0x80 || (n == 1 && ((c == 0xe0 && next < 0xa0) || (c == 0xf0 && next < 0x90) || (c == 0xf4 && next > 0x8f))) {
			return 0xfffd, n
		}
		n++
	}
	if n < width {
		if !eof {
			return 0, 0
		}
		return 0xfffd, n
	}
	var cp uint32
	switch width {
	case 2:
		cp = uint32(c&0x1f)<<6 | uint32(b[1]&0x3f)
	case 3:
		cp = uint32(c&0xf)<<12 | uint32(b[1]&0x3f)<<6 | uint32(b[2]&0x3f)
	case 4:
		cp = uint32(c&7)<<18 | uint32(b[1]&0x3f)<<12 | uint32(b[2]&0x3f)<<6 | uint32(b[3]&0x3f)
	}
	if cp >= 0xd800 && cp <= 0xdfff {
		return 0xfffd, width
	}
	return cp, width
}

type mailJDKStreamDecoder struct {
	input           *MailInputStream
	charset         string
	bytes           [8192]byte
	position, limit int
}

// readBytes compacts unread bytes and flips the buffer even when reading throws.
func (d *mailJDKStreamDecoder) readBytes() (n int, err error) {
	left := copy(d.bytes[:], d.bytes[d.position:d.limit])
	d.position, d.limit = 0, left
	defer func() { d.position = 0 }()
	n, err = d.input.ReadJava(d.bytes[:], int32(left), int32(len(d.bytes)-left))
	if err != nil {
		return 0, err
	}
	if n < 0 {
		return n, nil
	}
	if n == 0 {
		return 0, NewIOException("Underlying input stream returned zero bytes")
	}
	d.limit = left + n
	return d.limit, nil
}

// A failed implRead does not publish any of the characters it already decoded.
// BufferedReader.fill advances its character bounds only after a successful read.
func (d *mailJDKStreamDecoder) readChars(out []uint16) (int, error) {
	used := 0
	eof := false
	for {
		overflow := false
		for d.position < d.limit {
			cp, n := mailDecodeReaderCharacter(d.bytes[d.position:d.limit], d.charset, eof)
			if n == 0 {
				break
			}
			units := 1
			if cp > 0xffff {
				units = 2
			}
			if len(out)-used < units {
				overflow = true
				break
			}
			if units == 1 {
				out[used] = uint16(cp)
			} else {
				cp -= 0x10000
				out[used], out[used+1] = uint16(0xd800+(cp>>10)), uint16(0xdc00+(cp&0x3ff))
			}
			used += units
			d.position += n
		}
		if overflow || eof || used == len(out) {
			break
		}
		if used > 0 {
			ready, e := d.input.AvailableJava()
			if e != nil {
				if !isJavaIOException(e) {
					return 0, e
				}
				ready = 0
			}
			if ready <= 0 {
				break
			}
		}
		n, e := d.readBytes()
		if e != nil {
			return 0, e
		}
		if n < 0 {
			eof = true
			if used == 0 && d.position == d.limit {
				break
			}
		}
	}
	// These three decoders are stateless; their source reset at EOF is a no-op.
	if used == 0 && eof {
		return -1, nil
	}
	return used, nil
}

type mailJDKBufferedReader struct {
	decoder     mailJDKStreamDecoder
	chars       [8192]uint16
	next, count int
	skipLF      bool
}

func (r *mailJDKBufferedReader) readLine() (*string, error) {
	var line []uint16
	omitLF := r.skipLF
	for {
		if r.next >= r.count {
			n, e := r.decoder.readChars(r.chars[:])
			if e != nil {
				return nil, e
			}
			if n > 0 {
				r.next, r.count = 0, n
			}
		}
		if r.next >= r.count {
			if len(line) > 0 {
				return javaString(mailAddressUTF16String(line)), nil
			}
			return nil, nil
		}
		if omitLF && r.chars[r.next] == '\n' {
			r.next++
		}
		r.skipLF, omitLF = false, false
		start := r.next
		for r.next < r.count && r.chars[r.next] != '\r' && r.chars[r.next] != '\n' {
			r.next++
		}
		line = append(line, r.chars[start:r.next]...)
		if r.next < r.count {
			r.skipLF = r.chars[r.next] == '\r'
			r.next++
			return javaString(mailAddressUTF16String(line)), nil
		}
	}
}

func mailReadActivationLines(s *MailInputStream, charset string, consume func(string)) error {
	if s == nil {
		return NewNullPointerException()
	}
	r := mailJDKBufferedReader{decoder: mailJDKStreamDecoder{input: s, charset: charset}}
	for {
		line, e := r.readLine()
		if e != nil {
			return e
		}
		if line == nil {
			return nil
		}
		consume(*line)
	}
}
