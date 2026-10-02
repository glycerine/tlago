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
// Go port of the bundled Geronimo MailcapCommandMap registry operations.
// Source: geronimo-activation_1.1_spec/src/main/java/javax/activation/MailcapCommandMap.java.
package tlc

import (
	"strings"
	"sync"
)

// CommandInfo metadata is immutable. Bean instantiation and Externalizable
// command initialization remain part of the JVM object-provider port.
type MailCommandInfo struct{ name, class *string }

func NewMailCommandInfo(name, class *string) *MailCommandInfo {
	return &MailCommandInfo{name: name, class: class}
}
func (c *MailCommandInfo) GetCommandName() *string  { return c.name }
func (c *MailCommandInfo) GetCommandClass() *string { return c.class }

type mailCommandTable = mailParameterHashTable[*MailCommandInfo]

// This ports the registry instance; CommandMap superclass initialization and
// class-loader content-handler creation are separate pending runtime features.
type MailMailcapCommandMap struct {
	mu                  sync.Mutex
	environment         MailActivationEnvironment
	mimeTypes           *mailParameterHashTable[string]
	preferred, fallback map[string]*mailCommandTable
	all                 map[string][]*MailCommandInfo
	native              map[string][]string
}

func NewMailMailcapCommandMap(env ...MailActivationEnvironment) *MailMailcapCommandMap {
	e := DefaultMailActivationEnvironment
	if len(env) > 0 {
		e = env[0]
	}
	e = mailActivationEnvironment(e)
	m := &MailMailcapCommandMap{
		environment: e, mimeTypes: newMailParameterHashTable[string](),
		preferred: map[string]*mailCommandTable{}, fallback: map[string]*mailCommandTable{},
		all: map[string][]*MailCommandInfo{}, native: map[string][]string{},
	}
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
	_, err := mailInvoke(func() (bool, error) {
		s, err := e.DefaultResource("/META-INF/mailcap.default")
		if err != nil || s == nil {
			return false, err
		}
		return false, m.loadAndClose(s)
	})
	ignore(err, false)
	_, err = mailInvoke(func() (bool, error) {
		openers, err := e.Resources("META-INF/mailcap")
		if err != nil {
			return false, err
		}
		for _, open := range openers {
			_, err := mailInvoke(func() (bool, error) {
				s, err := open()
				if err != nil {
					return false, err
				}
				return false, m.loadAndClose(s)
			})
			if err != nil && !isJavaIOException(err) {
				return false, err
			}
		}
		return false, nil
	})
	ignore(err, true)
	for _, location := range []struct{ property, child string }{{"java.home", "lib/mailcap"}, {"user.home", ".mailcap"}} {
		_, err = mailInvoke(func() (bool, error) {
			parent := e.Property(location.property)
			path := location.child
			if parent != nil {
				path = filenameFileJoin(*parent, path)
			}
			s, err := e.OpenFile(path)
			if err != nil {
				return false, err
			}
			return false, m.loadAndClose(s)
		})
		ignore(err, true)
	}
	return m
}
func (m *MailMailcapCommandMap) parseStream(s *MailInputStream) error {
	_, err := mailInvoke(func() (bool, error) {
		return false, m.environment.ReadLines(s, func(line string) { m.AddMailcap(javaString(line)) })
	})
	if isJavaIOException(err) {
		return nil
	}
	return err
}
func (m *MailMailcapCommandMap) loadAndClose(s *MailInputStream) error {
	body := m.parseStream(s)
	_, close := mailInvoke(func() (bool, error) { return false, s.Close() })
	if close != nil {
		return close
	}
	return body
}
func NewMailMailcapCommandMapStream(s *MailInputStream, env ...MailActivationEnvironment) *MailMailcapCommandMap {
	m := NewMailMailcapCommandMap(env...)
	if err := m.parseStream(s); err != nil {
		panic(err)
	}
	return m
}
func NewMailMailcapCommandMapFile(path *string, env ...MailActivationEnvironment) (*MailMailcapCommandMap, error) {
	return mailInvoke(func() (*MailMailcapCommandMap, error) {
		m := NewMailMailcapCommandMap(env...)
		if path == nil {
			return nil, NewNullPointerException()
		}
		s, err := m.environment.OpenFile(*path)
		if err != nil {
			return nil, err
		}
		// FileReader's finally-close overrides any body failure. Unlike the
		// InputStream overload, Reader parsing does not swallow IOException.
		_, body := mailInvoke(func() (bool, error) {
			return false, m.environment.ReadLines(s, func(line string) { m.AddMailcap(javaString(line)) })
		})
		_, close := mailInvoke(func() (bool, error) { return false, s.Close() })
		if close != nil {
			return nil, close
		}
		if body != nil {
			return nil, body
		}
		return m, nil
	})
}
func (m *MailMailcapCommandMap) lower(value *string) string {
	if value == nil {
		panic(NewNullPointerException())
	}
	if f := m.environment.Lowercase; f != nil {
		return f(*value)
	}
	return mailJDKEnglishLower(*value)
}
func mailcapSkipSpace(s []uint16, index int) int {
	for index < len(s) && mailActivationWhitespace(s[index]) {
		index++
	}
	return index
}
func mailcapToken(s []uint16, index int) int {
	for index < len(s) && s[index] != '#' && !mailActivationSpecial(s[index]) {
		index++
	}
	return index
}
func mailcapMText(s []uint16, index int) int {
	for index < len(s) {
		c := s[index]
		if c == '#' || c == ';' || mailActivationISOControl(c) {
			return index
		}
		if c == '\\' {
			index++
			if index == len(s) {
				return index
			}
		}
		index++
	}
	return index
}
func (m *MailMailcapCommandMap) AddMailcap(value *string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if value == nil {
		panic(NewNullPointerException())
	}
	s := mailAddressUTF16(*value)
	index := mailcapSkipSpace(s, 0)
	if index == len(s) || s[index] == '#' {
		return
	}
	start := index
	index = mailcapToken(s, index)
	if start == index {
		return
	}
	mime := mailAddressUTF16String(s[start:index])
	index = mailcapSkipSpace(s, index)
	if index == len(s) || s[index] == '#' {
		return
	}
	if s[index] == '/' {
		index = mailcapSkipSpace(s, index+1)
		start = index
		index = mailcapToken(s, index)
		mime += "/" + mailAddressUTF16String(s[start:index])
	} else {
		mime += "/*"
	}
	mime = m.lower(javaString(mime))
	index = mailcapSkipSpace(s, index)
	if index == len(s) || s[index] != ';' {
		return
	}
	index = mailcapSkipSpace(s, index+1)
	if index == len(s) || s[index] != ';' {
		m.native[mime] = append(m.native[mime], *value)
		index = mailcapMText(s, index)
	}
	var commands []*MailCommandInfo
	fallback := false
	for index < len(s) && s[index] == ';' {
		index = mailcapSkipSpace(s, index+1)
		start = index
		index = mailcapToken(s, index)
		field := m.lower(javaString(mailAddressUTF16String(s[start:index])))
		index = mailcapSkipSpace(s, index)
		if index < len(s) && s[index] == '=' {
			index = mailcapSkipSpace(s, index+1)
			start = index
			index = mailcapMText(s, index)
			text := mailAddressUTF16String(s[start:index])
			index = mailcapSkipSpace(s, index)
			if strings.HasPrefix(field, "x-java-") && len(field) > 7 {
				command := field[7:]
				text = mailAddressTrim(text)
				if command == "fallback-entry" {
					if text == "true" {
						fallback = true
					}
				} else {
					commands = append(commands, NewMailCommandInfo(javaString(command), javaString(text)))
				}
			}
		}
	}
	m.mimeTypes.set(mime, mime)
	target := m.preferred
	if fallback {
		target = m.fallback
	}
	for _, command := range commands {
		table := target[mime]
		if table == nil {
			table = newMailParameterHashTable[*MailCommandInfo]()
			target[mime] = table
		}
		table.set(*command.name, command)
		if !fallback {
			m.all[mime] = append(m.all[mime], command)
		}
	}
}
func mailcapWildcard(mime string) string {
	if index := strings.IndexByte(mime, '/'); index >= 0 {
		return mime[:index+1] + "*"
	}
	return mime + "/*"
}
func mailcapMerge(main, fallback *mailCommandTable) *mailCommandTable {
	result := copyMailParameterHashTable(fallback)
	// In particular, a wildcard-only fallback passes nil main to putAll.
	result.putAll(main)
	return result
}
func (m *MailMailcapCommandMap) fallbackCommands(mime string) *mailCommandTable {
	commands, wildcard := m.fallback[mime], m.fallback[mailcapWildcard(mime)]
	if wildcard == nil {
		return commands
	}
	return mailcapMerge(commands, wildcard)
}
func mailcapValues(table *mailCommandTable) []*MailCommandInfo {
	result := []*MailCommandInfo{}
	if table != nil {
		for _, name := range table.names() {
			result = append(result, table.items.Get(name))
		}
	}
	return result
}
func (m *MailMailcapCommandMap) GetPreferredCommands(mimeType *string, source ...*MailDataSource) []*MailCommandInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	mime := m.lower(mimeType)
	commands := m.preferred[mime]
	if commands == nil {
		commands = m.preferred[mailcapWildcard(mime)]
	}
	fallback := m.fallbackCommands(mime)
	if fallback != nil {
		if commands == nil {
			commands = fallback
		} else {
			commands = mailcapMerge(commands, fallback)
		}
	}
	return mailcapValues(commands)
}
func (m *MailMailcapCommandMap) GetAllCommands(mimeType *string, source ...*MailDataSource) []*MailCommandInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	mime := m.lower(mimeType)
	exact, wildcard := m.all[mime], m.all[mailcapWildcard(mime)]
	fallback := m.fallbackCommands(mime)
	result := make([]*MailCommandInfo, 0, len(exact)+len(wildcard))
	result = append(result, exact...)
	result = append(result, wildcard...)
	return append(result, mailcapValues(fallback)...)
}
func (m *MailMailcapCommandMap) GetCommand(mimeType, commandName *string, source ...*MailDataSource) *MailCommandInfo {
	m.mu.Lock()
	defer m.mu.Unlock()
	mime := m.lower(mimeType)
	if index := strings.IndexByte(mime, ';'); index >= 0 {
		mime = mailAddressTrim(mime[:index])
	}
	commands := m.preferred[mime]
	if commands == nil {
		commands = m.preferred[mailcapWildcard(mime)]
	}
	if commands == nil {
		commands = m.fallback[mime]
	}
	if commands == nil {
		commands = m.fallback[mailcapWildcard(mime)]
	}
	if commands == nil {
		return nil
	}
	return commands.items.Get(m.lower(commandName))
}
func (m *MailMailcapCommandMap) GetMIMETypes() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.mimeTypes.names()
}
func (m *MailMailcapCommandMap) GetNativeCommands(mimeType *string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.native[m.lower(mimeType)]...)
}
