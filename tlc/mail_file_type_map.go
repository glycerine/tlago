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
// Go port of the bundled Geronimo MimetypesFileTypeMap and FileTypeMap.
// Source: https://github.com/apache/geronimo-specs/tree/trunk/geronimo-activation_1.1_spec/src/main/java/javax/activation
package tlc

import (
	"bytes"
	_ "embed"
	"os"
	"strings"
	"sync"
	"sync/atomic"
)

//go:embed resources/activation/mimetypes.default
var mailActivationDefaultMIMETypes []byte

// Resource and default-reader callbacks represent the JVM class-loader and
// InputStreamReader providers. Resource names retain their source leading slash.
// Configure the process environment before first use of the default map.
type MailActivationEnvironment struct {
	Classpath       []FilenameClasspathEntry
	DefaultResource func(string) (*MailInputStream, error)
	Resources       func(string) ([]func() (*MailInputStream, error), error)
	Property        func(string) *string
	OpenFile        func(string) (*MailInputStream, error)
	ReadLines       func(*MailInputStream, func(string)) error
	CheckSetFactory func() error
	Lowercase       func(string) string
}

var DefaultMailActivationEnvironment MailActivationEnvironment

func mailActivationEnvironment(e MailActivationEnvironment) MailActivationEnvironment {
	if e.Classpath == nil {
		e.Classpath = filenameDefaultClasspath()
	}
	if e.Property == nil {
		e.Property = func(name string) *string {
			if v, ok := tlcLookupSystemProperty(name); ok {
				return javaString(v)
			}
			if name == "user.home" {
				if h, x := os.UserHomeDir(); x == nil {
					return javaString(h)
				}
			}
			return nil
		}
	}
	if e.OpenFile == nil {
		e.OpenFile = mailOpenFileInputStream
	}
	if e.DefaultResource == nil {
		e.DefaultResource = func(name string) (*MailInputStream, error) {
			name = strings.TrimPrefix(name, "/")
			r := (&SimpleFilenameToStream{classpath: e.Classpath}).findClasspath(name)
			if r != nil {
				s, x := r.open()
				if x != nil {
					return nil, x
				}
				return NewMailInputStream(s, s), nil
			}
			if name == "META-INF/mimetypes.default" {
				return NewMailInputStream(bytes.NewReader(mailActivationDefaultMIMETypes), nil), nil
			}
			return nil, nil
		}
	}
	if e.Resources == nil {
		e.Resources = func(name string) ([]func() (*MailInputStream, error), error) {
			openers := []func() (*MailInputStream, error){}
			// ClassLoader.getResources does not strip the slash, unlike Class.getResource.
			for _, entry := range e.Classpath {
				r := (&SimpleFilenameToStream{classpath: []FilenameClasspathEntry{entry}}).findClasspath(name)
				if r != nil {
					openers = append(openers, func() (*MailInputStream, error) {
						s, x := r.open()
						if x != nil {
							return nil, x
						}
						return NewMailInputStream(s, s), nil
					})
				}
			}
			return openers, nil
		}
	}
	if e.ReadLines == nil {
		e.ReadLines = func(s *MailInputStream, consume func(string)) error {
			charset := "UTF-8"
			if cs := e.Property("file.encoding"); cs != nil {
				charset = *cs
			}
			name := mailCharsetName(charset)
			if name != "utf-8" && name != "us-ascii" && name != "iso-8859-1" {
				return NewUnsupportedOperationException("JDK InputStreamReader provider required for Activation charset " + charset)
			}
			return mailReadActivationLines(s, name, consume)
		}
	}
	return e
}

type MailMimetypesFileTypeMap struct {
	mu                  sync.Mutex
	types               map[string]string
	ContentTypeFunc     func(string) string
	FileContentTypeFunc func(*TLAFile) string
	environment         MailActivationEnvironment
}

func NewMailMimetypesFileTypeMap(env ...MailActivationEnvironment) *MailMimetypesFileTypeMap {
	e := DefaultMailActivationEnvironment
	if len(env) > 0 {
		e = env[0]
	}
	e = mailActivationEnvironment(e)
	m := &MailMimetypesFileTypeMap{types: map[string]string{}, environment: e}
	ignore := func(err error, security bool) {
		if err == nil || isJavaIOException(err) {
			return
		}
		if security {
			if _, ok := err.(*SecurityException); ok {
				return
			}
		}
		panic(err)
	}
	_, x := mailInvoke(func() (bool, error) {
		s, x := e.DefaultResource("/META-INF/mimetypes.default")
		if x != nil {
			return false, x
		}
		if s == nil {
			return false, nil
		}
		return false, m.loadAndClose(s)
	})
	ignore(x, false)
	_, x = mailInvoke(func() (bool, error) {
		openers, x := e.Resources("/META-INF/mime.types")
		if x != nil {
			return false, x
		}
		for _, open := range openers {
			_, x := mailInvoke(func() (bool, error) {
				s, x := open()
				if x != nil {
					return false, x
				}
				return false, m.loadAndClose(s)
			})
			if x != nil && !isJavaIOException(x) {
				return false, x
			}
		}
		return false, nil
	})
	ignore(x, true)
	for _, location := range []struct{ property, child string }{{"java.home", "lib/mime.types"}, {"user.home", ".mime.types"}} {
		_, x = mailInvoke(func() (bool, error) {
			parent := e.Property(location.property)
			path := location.child
			if parent != nil {
				path = filenameFileJoin(*parent, path)
			}
			s, x := e.OpenFile(path)
			if x != nil {
				return false, x
			}
			return false, m.loadAndClose(s)
		})
		ignore(x, true)
	}
	return m
}
func (m *MailMimetypesFileTypeMap) loadAndClose(s *MailInputStream) error {
	_, body := mailInvoke(func() (bool, error) { return false, m.environment.ReadLines(s, m.AddMIMETypes) })
	_, close := mailInvoke(func() (bool, error) { return false, s.Close() })
	if close != nil {
		return close
	}
	return body
}
func NewMailMimetypesFileTypeMapStream(s *MailInputStream, env ...MailActivationEnvironment) *MailMimetypesFileTypeMap {
	m := NewMailMimetypesFileTypeMap(env...)
	_, x := mailInvoke(func() (bool, error) { return false, m.environment.ReadLines(s, m.AddMIMETypes) })
	if x != nil && !isJavaIOException(x) {
		panic(x)
	}
	return m
}
func NewMailMimetypesFileTypeMapFile(path string, env ...MailActivationEnvironment) (*MailMimetypesFileTypeMap, error) {
	return mailInvoke(func() (*MailMimetypesFileTypeMap, error) {
		m := NewMailMimetypesFileTypeMap(env...)
		s, x := m.environment.OpenFile(path)
		if x != nil {
			return nil, x
		}
		// BufferedReader.close clears its reader in finally, even if close fails.
		closed := false
		closeReader := func() error {
			if closed {
				return nil
			}
			closed = true
			return s.Close()
		}
		_, x = mailInvoke(func() (bool, error) {
			if x := m.environment.ReadLines(s, m.AddMIMETypes); x != nil {
				return false, x
			}
			return false, closeReader()
		})
		if x != nil {
			if isJavaIOException(x) {
				_, close := mailInvoke(func() (bool, error) { return false, closeReader() })
				if close != nil && !isJavaIOException(close) {
					return nil, close
				}
			}
			return nil, x
		}
		return m, nil
	})
}
func (m *MailMimetypesFileTypeMap) AddMIMETypes(value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if i := strings.IndexByte(value, '#'); i >= 0 {
		value = value[:i]
	}
	tokens := strings.FieldsFunc(value, func(c rune) bool { return c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == '\f' })
	if len(tokens) == 0 {
		return
	}
	for _, extension := range tokens[1:] {
		m.types[extension] = tokens[0]
	}
}
func (m *MailMimetypesFileTypeMap) GetContentType(filename string) string {
	if m.ContentTypeFunc != nil {
		return m.ContentTypeFunc(filename)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	i := strings.LastIndexByte(filename, '.')
	if i < 0 || i == len(filename)-1 {
		return "application/octet-stream"
	}
	if v, ok := m.types[filename[i+1:]]; ok {
		return v
	}
	return "application/octet-stream"
}
func (m *MailMimetypesFileTypeMap) GetFileContentType(file *TLAFile) string {
	if m.FileContentTypeFunc != nil {
		return m.FileContentTypeFunc(file)
	}
	if file == nil {
		panic(NewNullPointerException())
	}
	return m.GetContentType(file.GetName())
}

var mailDefaultFileTypeMap struct {
	sync.Mutex
	value atomic.Pointer[MailMimetypesFileTypeMap]
}

func GetDefaultMailFileTypeMap() *MailMimetypesFileTypeMap {
	mailDefaultFileTypeMap.Lock()
	defer mailDefaultFileTypeMap.Unlock()
	if mailDefaultFileTypeMap.value.Load() == nil {
		mailDefaultFileTypeMap.value.Store(NewMailMimetypesFileTypeMap())
	}
	return mailDefaultFileTypeMap.value.Load()
}
func SetDefaultMailFileTypeMap(value *MailMimetypesFileTypeMap) error {
	if f := DefaultMailActivationEnvironment.CheckSetFactory; f != nil {
		if e := f(); e != nil {
			return e
		}
	}
	// Only the source getter is synchronized; a constructor provider can call
	// this setter while the getter holds the initialization monitor.
	mailDefaultFileTypeMap.value.Store(value)
	return nil
}
