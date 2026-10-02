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
// OpenJDK default InputStreamReader/OutputStreamWriter operations used by the
// bundled text handlers. UTF-8/ASCII/Latin-1 codec paths retain StreamEncoder
// buffering, pending surrogates and failed-buffer state. Other charsets need the
// explicit reader/writer provider, independently of charset-name metadata.
package tlc

import "io"

type MailTextWriter struct {
	WriteStringFunc func(*string) error
	FlushFunc       func() error
}

func (w *MailTextWriter) WriteString(text *string) error {
	_, err := mailInvoke(func() (bool, error) {
		if w == nil {
			return false, NewNullPointerException()
		}
		if w.WriteStringFunc == nil {
			return false, NewUnsupportedOperationException("OutputStreamWriter write provider required")
		}
		return false, w.WriteStringFunc(text)
	})
	return err
}
func (w *MailTextWriter) Flush() error {
	_, err := mailInvoke(func() (bool, error) {
		if w == nil {
			return false, NewNullPointerException()
		}
		if w.FlushFunc == nil {
			return false, NewUnsupportedOperationException("OutputStreamWriter flush provider required")
		}
		return false, w.FlushFunc()
	})
	return err
}
func mailReadDefaultText(input *MailInputStream) (*string, error) {
	if input == nil {
		return nil, NewNullPointerException()
	}
	charset := mailCharsetName(mailFlavorDefaultCharset())
	switch charset {
	case "utf-8", "us-ascii", "iso-8859-1":
	default:
		return nil, NewUnsupportedOperationException("JVM default InputStreamReader provider required for " + charset)
	}
	decoder := mailJDKStreamDecoder{input: input, charset: charset}
	result := make([]uint16, 0, 1024)
	buffer := make([]uint16, 32768)
	for {
		n, err := decoder.readChars(buffer)
		if err != nil {
			return nil, err
		}
		if n == -1 {
			break
		}
		result = append(result, buffer[:n]...)
	}
	return javaString(mailAddressUTF16String(result)), nil
}

type mailJDKStreamEncoder struct {
	mu              distributedServerMonitor
	out             io.Writer
	charset         string
	bytes           []byte
	position, limit int
	haveLeftover    bool
	leftover        uint16
}

func NewMailDefaultTextWriter(out io.Writer) (*MailTextWriter, error) {
	return mailInvoke(func() (*MailTextWriter, error) {
		if mailHandlerNull(out) {
			return nil, NewNullPointerException()
		}
		var charset string
		if p, ok := out.(interface{ MailPrintStreamCharset() string }); ok {
			charset = p.MailPrintStreamCharset()
		} else {
			charset = mailFlavorDefaultCharset()
		}
		charset = mailCharsetName(charset)
		switch charset {
		case "utf-8", "us-ascii", "iso-8859-1":
		default:
			return nil, NewUnsupportedOperationException("JVM default OutputStreamWriter provider required for " + charset)
		}
		e := &mailJDKStreamEncoder{out: out, charset: charset, bytes: make([]byte, 512), limit: 512}
		return &MailTextWriter{
			WriteStringFunc: func(s *string) error {
				if s == nil {
					return NewNullPointerException()
				}
				units := mailAddressUTF16(*s)
				e.mu.Lock()
				defer e.mu.Unlock()
				if len(units) == 0 {
					return nil
				}
				return e.write(units)
			},
			FlushFunc: func() error {
				e.mu.Lock()
				defer e.mu.Unlock()
				if e.position > 0 {
					if err := e.writeBytes(); err != nil {
						return err
					}
				}
				if out, ok := e.out.(interface{ Flush() error }); ok {
					return out.Flush()
				}
				return nil
			},
		}, nil
	})
}
func (e *mailJDKStreamEncoder) writeBytes() error {
	count := e.position
	e.position = 0
	e.limit = count // ByteBuffer.flip precedes the call
	if count > 0 {
		n, err := e.out.Write(e.bytes[:count])
		if err != nil {
			return err
		}
		if n != count {
			return NewIOException("short write")
		}
	}
	e.position = 0
	e.limit = len(e.bytes) // clear only on success
	return nil
}

// encode is CharsetEncoder.encode(..., false). Underflow keeps a final high
// surrogate unconsumed; replacement is '?' for all three source encoders.
func (e *mailJDKStreamEncoder) encode(chars []uint16) (used int, overflow bool) {
	for used < len(chars) {
		// ISO_8859_1's array loop bounds its initial scan by output space.
		// With no space it reports overflow before parsing a final surrogate;
		// UTF_8 and US_ASCII parse the surrogate first.
		if e.charset == "iso-8859-1" && e.position == e.limit {
			return used, true
		}
		c := chars[used]
		count := 1
		var data [4]byte
		n := 1
		if c >= 0xd800 && c <= 0xdbff {
			if used+1 == len(chars) {
				return used, false
			}
			low := chars[used+1]
			if low >= 0xdc00 && low <= 0xdfff {
				count = 2
				if e.charset == "utf-8" {
					cp := uint32(c-0xd800)*1024 + uint32(low-0xdc00) + 0x10000
					n = 4
					data = [4]byte{byte(0xf0 | cp>>18), byte(0x80 | cp>>12&63), byte(0x80 | cp>>6&63), byte(0x80 | cp&63)}
				} else {
					data[0] = '?'
				}
			} else {
				data[0] = '?'
			}
		} else if c >= 0xdc00 && c <= 0xdfff {
			data[0] = '?'
		} else if e.charset != "utf-8" {
			max := uint16(127)
			if e.charset == "iso-8859-1" {
				max = 255
			}
			if c > max {
				data[0] = '?'
			} else {
				data[0] = byte(c)
			}
		} else if c < 128 {
			data[0] = byte(c)
		} else if c < 2048 {
			n = 2
			data[0] = byte(0xc0 | c>>6)
			data[1] = byte(0x80 | c&63)
		} else {
			n = 3
			data[0] = byte(0xe0 | c>>12)
			data[1] = byte(0x80 | c>>6&63)
			data[2] = byte(0x80 | c&63)
		}
		if e.limit-e.position < n {
			return used, true
		}
		copy(e.bytes[e.position:], data[:n])
		e.position += n
		used += count
	}
	return used, false
}
func (e *mailJDKStreamEncoder) flushLeftover(chars []uint16) (remaining []uint16, err error) {
	if !e.haveLeftover {
		return chars, nil
	}
	buf := []uint16{e.leftover}
	if len(chars) > 0 {
		buf = append(buf, chars[0])
		chars = chars[1:]
	}
	for len(buf) > 0 {
		n, overflow := e.encode(buf)
		buf = buf[n:]
		if overflow {
			if err := e.writeBytes(); err != nil {
				return chars, err
			}
			continue
		}
		if len(buf) > 0 {
			e.leftover = buf[0]
			if len(chars) > 0 {
				buf = []uint16{e.leftover, chars[0]}
				chars = chars[1:]
				continue
			}
			return chars, nil
		}
		break
	}
	e.haveLeftover = false
	return chars, nil
}
func (e *mailJDKStreamEncoder) write(chars []uint16) error {
	var err error
	if e.haveLeftover {
		chars, err = e.flushLeftover(chars)
		if err != nil {
			return err
		}
	}
	if len(e.bytes) < 8192 {
		maxBytes := int32(len(chars))
		if e.charset == "utf-8" {
			maxBytes *= 3
		}
		capacity := maxBytes
		if capacity > 8192 {
			capacity = 8192
		}
		if int(capacity) > len(e.bytes) {
			if e.position > 0 {
				if err := e.writeBytes(); err != nil {
					return err
				}
			}
			e.bytes = make([]byte, int(capacity))
			e.position = 0
			e.limit = len(e.bytes)
		}
	}
	for len(chars) > 0 {
		n, overflow := e.encode(chars)
		chars = chars[n:]
		if overflow {
			if err := e.writeBytes(); err != nil {
				return err
			}
			continue
		}
		if len(chars) == 1 {
			e.haveLeftover = true
			e.leftover = chars[0]
		}
		break
	}
	return nil
}
