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

// OpenJDK Hashtable string-key/value buckets and legacy keys Enumeration used
// by AWT MimeTypeParameterList. This is separate from Activation's HashMap.
package tlc

import "sync"

type mailAWTHashEntry struct {
	name, value string
	hash        int32
	next        *mailAWTHashEntry
}
type mailAWTHashtable struct {
	mu       sync.Mutex
	table    []*mailAWTHashEntry
	size     int
	modCount int32
}

func newMailAWTHashtable() *mailAWTHashtable {
	return &mailAWTHashtable{table: make([]*mailAWTHashEntry, 11)}
}
func mailAWTStringHash(value string) int32 {
	var h int32
	for _, c := range mailAddressUTF16(value) {
		h = 31*h + int32(c)
	}
	return h
}
func (m *mailAWTHashtable) get(name string) *string {
	m.mu.Lock()
	defer m.mu.Unlock()
	h := mailAWTStringHash(name)
	index := int(uint32(h)&0x7fffffff) % len(m.table)
	for e := m.table[index]; e != nil; e = e.next {
		if e.hash == h && e.name == name {
			return javaString(e.value)
		}
	}
	return nil
}
func (m *mailAWTHashtable) set(name string, value *string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if value == nil {
		panic(NewNullPointerException())
	}
	h := mailAWTStringHash(name)
	index := int(uint32(h)&0x7fffffff) % len(m.table)
	for e := m.table[index]; e != nil; e = e.next {
		if e.hash == h && e.name == name {
			e.value = *value
			return
		}
	}
	if m.size >= int(float32(len(m.table))*float32(.75)) {
		m.modCount++ // rehash is a structural change of its own
		old := m.table
		capacity := len(old)*2 + 1
		m.table = make([]*mailAWTHashEntry, capacity)
		for i := len(old) - 1; i >= 0; i-- {
			for e := old[i]; e != nil; {
				next := e.next
				j := int(uint32(e.hash)&0x7fffffff) % capacity
				e.next = m.table[j]
				m.table[j] = e
				e = next
			}
		}
		index = int(uint32(h)&0x7fffffff) % capacity
	}
	m.modCount++
	m.table[index] = &mailAWTHashEntry{name: name, value: *value, hash: h, next: m.table[index]}
	m.size++
}
func (m *mailAWTHashtable) remove(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	h := mailAWTStringHash(name)
	i := int(uint32(h)&0x7fffffff) % len(m.table)
	var previous *mailAWTHashEntry
	for e := m.table[i]; e != nil; e = e.next {
		if e.hash == h && e.name == name {
			if previous == nil {
				m.table[i] = e.next
			} else {
				previous.next = e.next
			}
			m.size--
			m.modCount++
			return
		}
		previous = e
	}
}
func (m *mailAWTHashtable) length() int { m.mu.Lock(); defer m.mu.Unlock(); return m.size }
func (m *mailAWTHashtable) clone() *mailAWTHashtable {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := &mailAWTHashtable{table: make([]*mailAWTHashEntry, len(m.table)), size: m.size}
	for i, head := range m.table {
		tail := &result.table[i]
		for e := head; e != nil; e = e.next {
			x := *e
			x.next = nil
			*tail = &x
			tail = &x.next
		}
	}
	return result
}

// Unlike HashMap iterators, Hashtable keys() is a live non-fail-fast Enumeration.
// It retains the original table and follows entry links changed by a rehash.
type MailAWTParameterNames struct {
	owner            *mailAWTHashtable
	table            []*mailAWTHashEntry
	index            int
	entry            *mailAWTHashEntry
	empty            bool
	expectedModCount int32
}

func (m *mailAWTHashtable) names() *MailAWTParameterNames {
	m.mu.Lock()
	defer m.mu.Unlock()
	return &MailAWTParameterNames{owner: m, table: m.table, index: len(m.table), empty: m.size == 0, expectedModCount: m.modCount}
}
func (e *MailAWTParameterNames) find() bool {
	if e.empty {
		return false
	}
	for e.entry == nil && e.index > 0 {
		e.index--
		e.entry = e.table[e.index]
	}
	return e.entry != nil
}
func (e *MailAWTParameterNames) HasMoreElements() bool {
	if e == nil {
		panic(NewNullPointerException())
	}

	e.owner.mu.Lock()
	defer e.owner.mu.Unlock()
	return e.find()
}
func (e *MailAWTParameterNames) NextElement() string {
	if e == nil {
		panic(NewNullPointerException())
	}

	e.owner.mu.Lock()
	defer e.owner.mu.Unlock()
	if !e.find() {
		if e.empty {
			panic(NewNoSuchElementException())
		}
		panic(&NoSuchElementException{javaExceptionBase: newJavaExceptionBase(javaString("Hashtable Enumerator"), nil)})
	}
	entry := e.entry
	e.entry = entry.next
	return entry.name
}

// Parameter-list equality uses entrySet().iterator(), whose next() checks for
// structural changes. The public names Enumeration deliberately does not.
func (e *MailAWTParameterNames) nextEntry() (string, *string) {
	e.owner.mu.Lock()
	defer e.owner.mu.Unlock()
	if !e.empty && e.owner.modCount != e.expectedModCount {
		panic(NewConcurrentModificationException())
	}
	if !e.find() {
		if e.empty {
			panic(NewNoSuchElementException())
		}
		panic(&NoSuchElementException{javaExceptionBase: newJavaExceptionBase(javaString("Hashtable Enumerator"), nil)})
	}
	entry := e.entry
	e.entry = entry.next
	return entry.name, javaString(entry.value)
}
