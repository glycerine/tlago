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
// Go port of JavaMail 1.6.8 MimeUtility.getEncoding, stream checkAscii and
// AsciiOutputStream. Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/main/java/javax/mail/internet/MimeUtility.java
package tlc

import (
	"fmt"
	"sync"
)

const (
	MailAllASCII       = 1
	MailMostlyASCII    = 2
	MailMostlyNonASCII = 3
)

func mailNonASCII(b byte) bool { return b >= 0x7f || (b < 0x20 && b != '\r' && b != '\n' && b != '\t') }
func mailInvoke[T any](call func() (T, error)) (value T, err error) {
	defer func() {
		if x := recover(); x != nil {
			err = panicValueAsError(x)
		}
	}()
	return call()
}
func mailErrorIsJavaError(err error) bool { return isJavaError(err) }

func CheckMailStreamASCII(stream *MailInputStream, max int32, breakOnNonASCII bool) (int, error) {
	initializeMailMIMEProperties()
	return mailInvoke(func() (int, error) {
		var ascii, nonASCII, line int32
		longLine, badEOL := false, false
		strict := mailMIMEClass.encodeEOLStrict && breakOnNonASCII
		block := int32(4096)
		var buffer []byte
		if max != 0 {
			if max != -1 && max < block {
				block = max
			}
			if block < 0 {
				return 0, NewNegativeArraySizeException(fmt.Sprint(block))
			}
			buffer = make([]byte, int(block))
		}
		for max != 0 {
			n, err := stream.ReadJava(buffer, 0, block)
			if err != nil {
				if isJavaIOException(err) {
					break
				}
				return 0, err
			}
			if n == -1 {
				break
			}
			last := byte(0)
			for i := 0; i < n; i++ {
				if i >= len(buffer) {
					return 0, NewArrayIndexOutOfBoundsException(i, len(buffer))
				}
				b := buffer[i]
				if strict && ((last == '\r' && b != '\n') || (last != '\r' && b == '\n')) {
					badEOL = true
				}
				if b == '\r' || b == '\n' {
					line = 0
				} else {
					line++
					if line > 998 {
						longLine = true
					}
				}
				if mailNonASCII(b) {
					if breakOnNonASCII {
						return MailMostlyNonASCII, nil
					}
					nonASCII++
				} else {
					ascii++
				}
				last = b
			}
			if max != -1 {
				max -= int32(n)
			}
		}
		if max == 0 && breakOnNonASCII {
			return MailMostlyNonASCII, nil
		}
		if nonASCII == 0 {
			if badEOL {
				return MailMostlyNonASCII, nil
			}
			if longLine {
				return MailMostlyASCII, nil
			}
			return MailAllASCII, nil
		}
		if ascii > nonASCII {
			return MailMostlyASCII, nil
		}
		return MailMostlyNonASCII, nil
	})
}

type MailASCIIOutputStream struct {
	breakOnNonASCII, checkEOL        bool
	ascii, nonASCII, line, last, ret int32
	longLine, badEOL                 bool
}

func NewMailASCIIOutputStream(breakOnNonASCII, strict bool) *MailASCIIOutputStream {
	return &MailASCIIOutputStream{breakOnNonASCII: breakOnNonASCII, checkEOL: strict && breakOnNonASCII}
}
func (s *MailASCIIOutputStream) WriteByteJava(value int32) error {
	b := value & 0xff
	if s.checkEOL && ((s.last == '\r' && b != '\n') || (s.last != '\r' && b == '\n')) {
		s.badEOL = true
	}
	if b == '\r' || b == '\n' {
		s.line = 0
	} else {
		s.line++
		if s.line > 998 {
			s.longLine = true
		}
	}
	if mailNonASCII(byte(b)) {
		s.nonASCII++
		if s.breakOnNonASCII {
			s.ret = MailMostlyNonASCII
			return NewEOFException()
		}
	} else {
		s.ascii++
	}
	s.last = b
	return nil
}
func (s *MailASCIIOutputStream) Write(data []byte) (int, error) {
	if data == nil {
		return 0, NewNullPointerException()
	}
	for i, b := range data {
		if e := s.WriteByteJava(int32(b)); e != nil {
			return i, e
		}
	}
	return len(data), nil
}
func (s *MailASCIIOutputStream) WriteRange(data []byte, off, length int32) error {
	end := length + off
	for i := off; i < end; i++ {
		if data == nil {
			return NewNullPointerException()
		}
		if i < 0 || int(i) >= len(data) {
			return NewArrayIndexOutOfBoundsException(int(i), len(data))
		}
		if e := s.WriteByteJava(int32(data[i])); e != nil {
			return e
		}
	}
	return nil
}
func (s *MailASCIIOutputStream) GetASCII() int {
	if s.ret != 0 {
		return int(s.ret)
	}
	if s.badEOL {
		return MailMostlyNonASCII
	}
	if s.nonASCII == 0 {
		if s.longLine {
			return MailMostlyASCII
		}
		return MailAllASCII
	}
	if s.ascii > s.nonASCII {
		return MailMostlyASCII
	}
	return MailMostlyNonASCII
}

var mailNonASCIICharsetCache struct {
	sync.Mutex
	values map[string]bool
}

func (m MailMIMECodec) nonASCIICharset(ct *MailContentType) (bool, error) {
	charset := ct.GetParameter("charset")
	if charset == nil {
		return false, nil
	}
	name := mailCharsetLower(*charset)
	mailNonASCIICharsetCache.Lock()
	v, found := mailNonASCIICharsetCache.values[name]
	mailNonASCIICharsetCache.Unlock()
	if found {
		return v, nil
	}
	encoder := m.EncodeCharset
	if encoder == nil {
		encoder = encodeMailCharset
	}
	b, e := mailInvoke(func() ([]byte, error) { return encoder("\r\n", name) })
	if e != nil {
		if _, ok := e.(*UnsupportedEncodingException); ok {
			v = false
		} else if mailErrorIsJavaError(e) {
			return false, e
		} else if isJavaIOException(e) {
			return false, e
		} else {
			v = true
		}
	} else {
		v = len(b) != 2 || b[0] != '\r' || b[1] != '\n'
	}
	mailNonASCIICharsetCache.Lock()
	if mailNonASCIICharsetCache.values == nil {
		mailNonASCIICharsetCache.values = map[string]bool{}
	}
	mailNonASCIICharsetCache.values[name] = v
	mailNonASCIICharsetCache.Unlock()
	return v, nil
}
func GetMailDataSourceEncoding(ds *MailDataSource) (string, error) {
	return DefaultMailMIMECodec.GetDataSourceEncoding(ds)
}
func (m MailMIMECodec) GetDataSourceEncoding(ds *MailDataSource) (string, error) {
	initializeMailMIMEProperties()
	return mailInvoke(func() (string, error) {
		if ds != nil && ds.EncodingFunc != nil {
			v := ds.EncodingFunc()
			if v != nil {
				return *v, nil
			}
		}
		var stream *MailInputStream
		result, err := mailInvoke(func() (string, error) {
			ct, e := NewMailContentType(ds.GetContentType())
			if e != nil {
				return "", e
			}
			stream, e = ds.GetInputStream()
			if e != nil {
				return "", e
			}
			text := ct.MatchString(javaString("text/*"))
			kind, e := CheckMailStreamASCII(stream, -1, !text)
			if e != nil {
				return "", e
			}
			switch kind {
			case MailAllASCII:
				return "7bit", nil
			case MailMostlyASCII:
				if text {
					non, e := m.nonASCIICharset(ct)
					if e != nil {
						return "", e
					}
					if non {
						return "base64", nil
					}
				}
				return "quoted-printable", nil
			default:
				return "base64", nil
			}
		})
		if err != nil && !mailErrorIsJavaError(err) {
			result = "base64"
			err = nil
		}
		if stream != nil {
			_, e := mailInvoke(func() (bool, error) { return false, stream.Close() })
			if e != nil && !isJavaIOException(e) {
				return "", e
			}
		}
		return result, err
	})
}
func GetMailDataHandlerEncoding(dh *MailDataHandler) (string, error) {
	return DefaultMailMIMECodec.GetDataHandlerEncoding(dh)
}
func (m MailMIMECodec) GetDataHandlerEncoding(dh *MailDataHandler) (string, error) {
	initializeMailMIMEProperties()
	return mailInvoke(func() (string, error) {
		if dh.GetName() != nil {
			return m.GetDataSourceEncoding(dh.GetDataSource())
		}
		ct, e := mailInvoke(func() (*MailContentType, error) { return NewMailContentType(dh.GetContentType()) })
		if e != nil {
			if mailErrorIsJavaError(e) {
				return "", e
			}
			return "base64", nil
		}
		text := ct.MatchString(javaString("text/*"))
		out := NewMailASCIIOutputStream(!text, mailMIMEClass.encodeEOLStrict)
		_, e = mailInvoke(func() (bool, error) { return false, dh.WriteTo(out) })
		if e != nil && !isJavaIOException(e) {
			return "", e
		}
		kind := out.GetASCII()
		if kind == MailAllASCII {
			return "7bit", nil
		}
		if text && kind == MailMostlyASCII {
			return "quoted-printable", nil
		}
		return "base64", nil
	})
}
