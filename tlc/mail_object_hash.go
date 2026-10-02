/*
 * Copyright (c) 1997, 2023, Oracle and/or its affiliates. All rights reserved.
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
// Object-key HashMap operations used by ImageIO ServiceRegistry and its graph.
// Bucket lists, collision trees and iterator removal follow OpenJDK 21 HashMap.
// Identity/class/comparable metadata belongs to the native object carriers.
package tlc

type mailObjectHashKey interface {
	mailObjectHashCode() int32
	mailObjectEquals(mailObjectHashKey) bool
	mailObjectClass() *MailActivationClass
	mailObjectClassName() string
	mailObjectIdentityHash() int32
	mailObjectComparableClass() *MailActivationClass
	mailObjectCompare(mailObjectHashKey) int32
}

func (c *MailActivationClass) mailObjectHashCode() int32                       { return c.HashCode() }
func (c *MailActivationClass) mailObjectEquals(k mailObjectHashKey) bool       { return c == k }
func (c *MailActivationClass) mailObjectClass() *MailActivationClass           { return nil } // Class is not Comparable.
func (c *MailActivationClass) mailObjectClassName() string                     { return "java.lang.Class" }
func (c *MailActivationClass) mailObjectIdentityHash() int32                   { return c.HashCode() }
func (c *MailActivationClass) mailObjectComparableClass() *MailActivationClass { return nil }
func (c *MailActivationClass) mailObjectCompare(mailObjectHashKey) int32 {
	panic(NewClassCastException())
}
func (p *MailImageSPI) mailObjectHashCode() int32 {
	if p.HashCodeFunc != nil {
		return p.HashCodeFunc()
	}
	return p.mailObjectIdentityHash()
}
func (p *MailImageSPI) mailObjectEquals(k mailObjectHashKey) bool {
	if p.EqualsFunc != nil {
		return p.EqualsFunc(k)
	}
	return p == k
}
func (p *MailImageSPI) mailObjectClass() *MailActivationClass {
	if p.Class == nil {
		panic(NewUnsupportedOperationException("Image SPI leaf Class provider required"))
	}
	return p.Class
}
func (p *MailImageSPI) mailObjectClassName() string { return p.mailObjectClass().GetName() }
func (p *MailImageSPI) mailObjectIdentityHash() int32 {
	if p.IdentityHashCodeFunc != nil {
		return p.IdentityHashCodeFunc()
	}
	if h := p.identityHash.Load(); h != 0 {
		return h
	}
	h := mailFlavorIdentityHash.Add(1) & 0x7fffffff
	if h == 0 {
		h = 1
	}
	p.identityHash.CompareAndSwap(0, h)
	return p.identityHash.Load()
}
func (p *MailImageSPI) mailObjectComparableClass() *MailActivationClass { return p.ComparableClass }
func (p *MailImageSPI) mailObjectCompare(k mailObjectHashKey) int32 {
	if p.CompareToFunc == nil {
		panic(NewUnsupportedOperationException("Image SPI Comparable provider required"))
	}
	return p.CompareToFunc(k)
}
func mailObjectHash(k mailObjectHashKey) int32 {
	if k == nil {
		return 0
	}
	h := uint32(k.mailObjectHashCode())
	return int32(h ^ (h >> 16))
}
func mailObjectKey(k mailObjectHashKey) mailObjectHashKey {
	if mailHandlerNull(k) {
		return nil
	}
	return k
}
func mailObjectKeysEqual(k, stored mailObjectHashKey) bool {
	return k == stored || (k != nil && k.mailObjectEquals(stored))
}
func mailObjectCompare(kc *MailActivationClass, k, other mailObjectHashKey) int32 {
	if other == nil || other.mailObjectClass() != kc {
		return 0
	}
	return k.mailObjectCompare(other)
}
func mailObjectTieBreak(a, b mailObjectHashKey) int {
	if a != nil && b != nil {
		if d := mailParameterHashDirection(&mailParameterHashNode{name: a.mailObjectClassName()}, &mailParameterHashNode{name: b.mailObjectClassName()}); d != 0 {
			return d
		}
	}
	var ah, bh int32
	if a != nil {
		ah = a.mailObjectIdentityHash()
	}
	if b != nil {
		bh = b.mailObjectIdentityHash()
	}
	if ah <= bh {
		return -1
	}
	return 1
}

// Structural red-black rotations/balancing reuse the already-ported HashMap
// node implementation. Object keys and values are separate from string keys.
type mailObjectHashMap struct {
	table                     []*mailParameterHashNode
	size, threshold, modCount int32
}

func (m *mailObjectHashMap) front(table []*mailParameterHashNode, root *mailParameterHashNode) {
	o := mailParameterHashOrder{table: table}
	o.front(root)
}
func mailObjectTreeRoot(p *mailParameterHashNode) *mailParameterHashNode {
	for p.parent != nil {
		p = p.parent
	}
	return p
}
func mailObjectTreeFind(p *mailParameterHashNode, h int32, k mailObjectHashKey, kc *MailActivationClass) *mailParameterHashNode {
	for p != nil {
		pl, pr := p.left, p.right
		if p.hash > h {
			p = pl
		} else if p.hash < h {
			p = pr
		} else if mailObjectKeysEqual(k, p.objectKey) {
			return p
		} else if pl == nil {
			p = pr
		} else if pr == nil {
			p = pl
		} else {
			if kc == nil && k != nil {
				kc = k.mailObjectComparableClass()
			}
			if kc != nil {
				d := mailObjectCompare(kc, k, p.objectKey)
				if d != 0 {
					if d < 0 {
						p = pl
					} else {
						p = pr
					}
					continue
				}
			}
			if q := mailObjectTreeFind(pr, h, k, kc); q != nil {
				return q
			}
			p = pl
		}
	}
	return nil
}
func (m *mailObjectHashMap) node(key mailObjectHashKey) *mailParameterHashNode {
	k := mailObjectKey(key)
	tab := m.table
	if len(tab) == 0 {
		return nil
	}
	h := mailObjectHash(k)
	p := tab[int(uint32(h)&uint32(len(tab)-1))]
	if p == nil {
		return nil
	}
	if p.hash == h && mailObjectKeysEqual(k, p.objectKey) {
		return p
	}
	e := p.next
	if e == nil {
		return nil
	}
	if p.tree {
		return mailObjectTreeFind(mailObjectTreeRoot(p), h, k, nil)
	}
	for ; e != nil; e = e.next {
		if e.hash == h && mailObjectKeysEqual(k, e.objectKey) {
			return e
		}
	}
	return nil
}
func (m *mailObjectHashMap) get(k mailObjectHashKey) any {
	if p := m.node(k); p != nil {
		return p.objectValue
	}
	return nil
}
func (m *mailObjectHashMap) contains(k mailObjectHashKey) bool { return m.node(k) != nil }
func (m *mailObjectHashMap) put(key mailObjectHashKey, value any) any {
	k := mailObjectKey(key)
	h := mailObjectHash(k)
	tab := m.table
	if len(tab) == 0 {
		tab = m.resize()
	}
	index := int(uint32(h) & uint32(len(tab)-1))
	p := tab[index]
	if p == nil {
		tab[index] = &mailParameterHashNode{objectKey: k, objectValue: value, hash: h}
	} else {
		var e *mailParameterHashNode
		if p.hash == h && mailObjectKeysEqual(k, p.objectKey) {
			e = p
		} else if p.tree {
			e = m.putTree(p, tab, h, k, value)
		} else {
			for count := 0; ; count++ {
				e = p.next
				if e == nil {
					p.next = &mailParameterHashNode{objectKey: k, objectValue: value, hash: h}
					if count >= 7 {
						m.treeifyBin(tab, index)
					}
					break
				}
				if e.hash == h && mailObjectKeysEqual(k, e.objectKey) {
					break
				}
				p = e
			}
		}
		if e != nil {
			old := e.objectValue
			e.objectValue = value
			return old
		}
	}
	m.modCount++
	m.size++
	if m.size > m.threshold {
		m.resize()
	}
	return nil
}
func (m *mailObjectHashMap) putTree(head *mailParameterHashNode, table []*mailParameterHashNode, h int32, k mailObjectHashKey, value any) *mailParameterHashNode {
	var kc *MailActivationClass
	searched := false
	root := mailObjectTreeRoot(head)
	for p := root; ; {
		var d int
		if p.hash > h {
			d = -1
		} else if p.hash < h {
			d = 1
		} else if mailObjectKeysEqual(k, p.objectKey) {
			return p
		} else {
			if kc == nil && k != nil {
				kc = k.mailObjectComparableClass()
			}
			if kc != nil {
				d = int(mailObjectCompare(kc, k, p.objectKey))
			}
			if kc == nil || d == 0 {
				if !searched {
					searched = true
					if p.left != nil {
						if q := mailObjectTreeFind(p.left, h, k, kc); q != nil {
							return q
						}
					}
					if p.right != nil {
						if q := mailObjectTreeFind(p.right, h, k, kc); q != nil {
							return q
						}
					}
				}
				d = mailObjectTieBreak(k, p.objectKey)
			}
		}
		xp := p
		if d <= 0 {
			p = p.left
		} else {
			p = p.right
		}
		if p == nil {
			xpn := xp.next
			x := &mailParameterHashNode{objectKey: k, objectValue: value, hash: h, next: xpn, prev: xp, parent: xp, tree: true}
			if d <= 0 {
				xp.left = x
			} else {
				xp.right = x
			}
			xp.next = x
			if xpn != nil {
				xpn.prev = x
			}
			m.front(table, mailParameterBalanceInsertion(root, x))
			return nil
		}
	}
}
func (m *mailObjectHashMap) treeifyBin(tab []*mailParameterHashNode, index int) {
	if len(tab) < 64 {
		m.resize()
		return
	}
	var head, tail *mailParameterHashNode
	for e := tab[index]; e != nil; e = e.next {
		p := &mailParameterHashNode{objectKey: e.objectKey, objectValue: e.objectValue, hash: e.hash, tree: true, prev: tail}
		if tail == nil {
			head = p
		} else {
			tail.next = p
		}
		tail = p
	}
	tab[index] = head
	m.treeify(tab, index)
}
func (m *mailObjectHashMap) treeify(tab []*mailParameterHashNode, index int) {
	var root *mailParameterHashNode
	for x := tab[index]; x != nil; x = x.next {
		x.left = nil
		x.right = nil
		if root == nil {
			x.parent = nil
			x.red = false
			root = x
			continue
		}
		k, h := x.objectKey, x.hash
		var kc *MailActivationClass
		for p := root; ; {
			var d int
			if p.hash > h {
				d = -1
			} else if p.hash < h {
				d = 1
			} else {
				if kc == nil && k != nil {
					kc = k.mailObjectComparableClass()
				}
				if kc != nil {
					d = int(mailObjectCompare(kc, k, p.objectKey))
				}
				if kc == nil || d == 0 {
					d = mailObjectTieBreak(k, p.objectKey)
				}
			}
			xp := p
			if d <= 0 {
				p = p.left
			} else {
				p = p.right
			}
			if p == nil {
				x.parent = xp
				if d <= 0 {
					xp.left = x
				} else {
					xp.right = x
				}
				root = mailParameterBalanceInsertion(root, x)
				break
			}
		}
	}
	m.front(tab, root)
}
func mailObjectUntreeify(head *mailParameterHashNode) *mailParameterHashNode {
	var hd, tl *mailParameterHashNode
	for q := head; q != nil; q = q.next {
		p := &mailParameterHashNode{objectKey: q.objectKey, objectValue: q.objectValue, hash: q.hash}
		if tl == nil {
			hd = p
		} else {
			tl.next = p
		}
		tl = p
	}
	return hd
}
func (m *mailObjectHashMap) resize() []*mailParameterHashNode {
	old := m.table
	capacity := len(old)
	if capacity == 0 {
		m.table = make([]*mailParameterHashNode, 16)
		m.threshold = 12
		return m.table
	}
	if capacity >= 1<<30 {
		m.threshold = 0x7fffffff
		return old
	}
	newCap := capacity * 2
	if newCap < 1<<30 {
		m.threshold = m.threshold * 2
	} else {
		m.threshold = 0x7fffffff
	}
	tab := make([]*mailParameterHashNode, newCap)
	m.table = tab
	for index, e := range old {
		if e == nil {
			continue
		}
		old[index] = nil
		if e.next == nil {
			tab[int(uint32(e.hash)&uint32(newCap-1))] = e
			continue
		}
		tree := e.tree
		var low, lt, high, ht *mailParameterHashNode
		lc, hc := 0, 0
		for e != nil {
			next := e.next
			if uint32(e.hash)&uint32(capacity) == 0 {
				if tree {
					e.prev = lt
				}
				if lt == nil {
					low = e
				} else {
					lt.next = e
				}
				lt = e
				lc++
			} else {
				if tree {
					e.prev = ht
				}
				if ht == nil {
					high = e
				} else {
					ht.next = e
				}
				ht = e
				hc++
			}
			e = next
		}
		if lt != nil {
			lt.next = nil
			tab[index] = low
		}
		if ht != nil {
			ht.next = nil
			tab[index+capacity] = high
		}
		if tree {
			if low != nil {
				if lc <= 6 {
					tab[index] = mailObjectUntreeify(low)
				} else if high != nil {
					m.treeify(tab, index)
				}
			}
			if high != nil {
				if hc <= 6 {
					tab[index+capacity] = mailObjectUntreeify(high)
				} else if low != nil {
					m.treeify(tab, index+capacity)
				}
			}
		}
	}
	return tab
}
func (m *mailObjectHashMap) remove(key mailObjectHashKey) any {
	k := mailObjectKey(key)
	if p := m.removeNode(mailObjectHash(k), k, true); p != nil {
		return p.objectValue
	}
	return nil
}
func (m *mailObjectHashMap) removeNode(h int32, k mailObjectHashKey, movable bool) *mailParameterHashNode {
	tab := m.table
	if len(tab) == 0 {
		return nil
	}
	index := int(uint32(h) & uint32(len(tab)-1))
	p := tab[index]
	if p == nil {
		return nil
	}
	var node *mailParameterHashNode
	if p.hash == h && mailObjectKeysEqual(k, p.objectKey) {
		node = p
	} else if e := p.next; e != nil {
		if p.tree {
			node = mailObjectTreeFind(mailObjectTreeRoot(p), h, k, nil)
		} else {
			for ; e != nil; e = e.next {
				if e.hash == h && mailObjectKeysEqual(k, e.objectKey) {
					node = e
					break
				}
				p = e
			}
		}
	}
	if node == nil {
		return nil
	}
	if node.tree {
		m.removeTree(tab, index, node, movable)
	} else if node == p {
		tab[index] = node.next
	} else {
		p.next = node.next
	}
	m.modCount++
	m.size--
	return node
}
func (m *mailObjectHashMap) clear() {
	m.modCount++
	if len(m.table) > 0 && m.size > 0 {
		m.size = 0
		for i := range m.table {
			m.table[i] = nil
		}
	}
}
func (m *mailObjectHashMap) removeTree(tab []*mailParameterHashNode, index int, p *mailParameterHashNode, movable bool) {
	first := tab[index]
	root := first
	succ, pred := p.next, p.prev
	if pred == nil {
		first = succ
		tab[index] = first
	} else {
		pred.next = succ
	}
	if succ != nil {
		succ.prev = pred
	}
	if first == nil {
		return
	}
	for root.parent != nil {
		root = root.parent
	}
	if movable && (root.right == nil || root.left == nil || root.left.left == nil) {
		tab[index] = mailObjectUntreeify(first)
		return
	}
	pl, pr := p.left, p.right
	var replacement *mailParameterHashNode
	if pl != nil && pr != nil {
		s := pr
		for s.left != nil {
			s = s.left
		}
		s.red, p.red = p.red, s.red
		sr, pp := s.right, p.parent
		if s == pr {
			p.parent = s
			s.right = p
		} else {
			sp := s.parent
			p.parent = sp
			if sp != nil {
				if s == sp.left {
					sp.left = p
				} else {
					sp.right = p
				}
			}
			s.right = pr
			if pr != nil {
				pr.parent = s
			}
		}
		p.left = nil
		p.right = sr
		if sr != nil {
			sr.parent = p
		}
		s.left = pl
		if pl != nil {
			pl.parent = s
		}
		s.parent = pp
		if pp == nil {
			root = s
		} else if p == pp.left {
			pp.left = s
		} else {
			pp.right = s
		}
		if sr != nil {
			replacement = sr
		} else {
			replacement = p
		}
	} else if pl != nil {
		replacement = pl
	} else if pr != nil {
		replacement = pr
	} else {
		replacement = p
	}
	if replacement != p {
		pp := p.parent
		replacement.parent = pp
		if pp == nil {
			root = replacement
			root.red = false
		} else if p == pp.left {
			pp.left = replacement
		} else {
			pp.right = replacement
		}
		p.left = nil
		p.right = nil
		p.parent = nil
	}
	r := root
	if !p.red {
		r = mailParameterBalanceDeletion(root, replacement)
	}
	if replacement == p {
		pp := p.parent
		p.parent = nil
		if pp != nil {
			if p == pp.left {
				pp.left = nil
			} else if p == pp.right {
				pp.right = nil
			}
		}
	}
	if movable {
		m.front(tab, r)
	}
}

type mailObjectHashIterator struct {
	source        *mailObjectHashMap
	next, current *mailParameterHashNode
	index         int
	expected      int32
}

func (m *mailObjectHashMap) iterator() *mailObjectHashIterator {
	i := &mailObjectHashIterator{source: m, expected: m.modCount}
	if len(m.table) > 0 && m.size > 0 {
		i.advance()
	}
	return i
}
func (i *mailObjectHashIterator) advance() {
	for i.index < len(i.source.table) {
		i.next = i.source.table[i.index]
		i.index++
		if i.next != nil {
			return
		}
	}
}
func (i *mailObjectHashIterator) hasNext() bool { return i.next != nil }
func (i *mailObjectHashIterator) nextNode() *mailParameterHashNode {
	if i.source.modCount != i.expected {
		panic(NewConcurrentModificationException())
	}
	e := i.next
	if e == nil {
		panic(NewNoSuchElementException())
	}
	i.current = e
	i.next = e.next
	if i.next == nil {
		i.advance()
	}
	return e
}
func (i *mailObjectHashIterator) remove() {
	p := i.current
	if p == nil {
		panic(NewIllegalStateException())
	}
	if i.source.modCount != i.expected {
		panic(NewConcurrentModificationException())
	}
	i.current = nil
	i.source.removeNode(p.hash, p.objectKey, false)
	i.expected = i.source.modCount
}
