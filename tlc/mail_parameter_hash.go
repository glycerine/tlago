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
// Go port of the OpenJDK 21 HashMap bucket/tree ordering used by JavaMail's
// RFC2231 auxiliary HashMap/HashSet and Activation metadata maps. Keys are
// always non-null Java strings.
// Source: https://github.com/openjdk/jdk21u/blob/master/src/java.base/share/classes/java/util/HashMap.java
package tlc

type mailParameterHashNode struct {
	name                            string
	hash                            int32
	next, prev, parent, left, right *mailParameterHashNode
	red, tree                       bool
}
type mailParameterHashOrder struct {
	table []*mailParameterHashNode
	size  int
}

func mailParameterHashDirection(a, b *mailParameterHashNode) int {
	if a.hash < b.hash {
		return -1
	}
	if a.hash > b.hash {
		return 1
	}
	x, y := mailAddressUTF16(a.name), mailAddressUTF16(b.name)
	for i := 0; i < len(x) && i < len(y); i++ {
		if x[i] != y[i] {
			return int(x[i]) - int(y[i])
		}
	}
	return len(x) - len(y)
}
func (m *mailParameterHashOrder) names() []string {
	names := make([]string, 0, m.size)
	for _, head := range m.table {
		for p := head; p != nil; p = p.next {
			names = append(names, p.name)
		}
	}
	return names
}
func (m *mailParameterHashOrder) clear() {
	for i := range m.table {
		m.table[i] = nil
	}
	m.size = 0
}
func (m *mailParameterHashOrder) add(name string) {
	if len(m.table) == 0 {
		m.table = make([]*mailParameterHashNode, 16)
	}
	n := &mailParameterHashNode{name: name, hash: int32(parameterHash(name))}
	index := int(uint32(n.hash) & uint32(len(m.table)-1))
	head := m.table[index]
	if head == nil {
		m.table[index] = n
	} else if head.tree {
		root := head
		for root.parent != nil {
			root = root.parent
		}
		for p := root; ; {
			dir := mailParameterHashDirection(n, p)
			parent := p
			if dir <= 0 {
				p = p.left
			} else {
				p = p.right
			}
			if p == nil {
				n.tree = true
				n.parent = parent
				n.prev = parent
				n.next = parent.next
				if dir <= 0 {
					parent.left = n
				} else {
					parent.right = n
				}
				parent.next = n
				if n.next != nil {
					n.next.prev = n
				}
				m.front(mailParameterBalanceInsertion(root, n))
				break
			}
		}
	} else {
		p, count := head, 1
		for p.next != nil {
			p = p.next
			count++
		}
		p.next = n
		if count >= 8 {
			if len(m.table) < 64 {
				m.resize()
			} else {
				m.treeify(index)
			}
		}
	}
	m.size++
	if m.size > len(m.table)*3/4 {
		m.resize()
	}
}
func (m *mailParameterHashOrder) front(root *mailParameterHashNode) {
	if root == nil {
		return
	}
	index := int(uint32(root.hash) & uint32(len(m.table)-1))
	first := m.table[index]
	if root != first {
		m.table[index] = root
		rp, rn := root.prev, root.next
		if rn != nil {
			rn.prev = rp
		}
		if rp != nil {
			rp.next = rn
		}
		if first != nil {
			first.prev = root
		}
		root.next = first
		root.prev = nil
	}
}
func (m *mailParameterHashOrder) treeify(index int) {
	var root, previous *mailParameterHashNode
	for x := m.table[index]; x != nil; x = x.next {
		x.tree = true
		x.prev = previous
		previous = x
		x.left = nil
		x.right = nil
		if root == nil {
			x.parent = nil
			x.red = false
			root = x
			continue
		}
		for p := root; ; {
			dir := mailParameterHashDirection(x, p)
			parent := p
			if dir <= 0 {
				p = p.left
			} else {
				p = p.right
			}
			if p == nil {
				x.parent = parent
				if dir <= 0 {
					parent.left = x
				} else {
					parent.right = x
				}
				root = mailParameterBalanceInsertion(root, x)
				break
			}
		}
	}
	m.front(root)
}
func mailParameterUntreeify(head *mailParameterHashNode) {
	for p := head; p != nil; p = p.next {
		p.tree = false
		p.prev = nil
		p.parent = nil
		p.left = nil
		p.right = nil
		p.red = false
	}
}
func (m *mailParameterHashOrder) resize() {
	old := m.table
	capacity := len(old)
	m.table = make([]*mailParameterHashNode, capacity*2)
	for i, head := range old {
		if head == nil {
			continue
		}
		tree := head.tree
		var low, lowTail, high, highTail *mailParameterHashNode
		lc, hc := 0, 0
		for e := head; e != nil; {
			next := e.next
			e.next = nil
			if uint32(e.hash)&uint32(capacity) == 0 {
				e.prev = lowTail
				if lowTail == nil {
					low = e
				} else {
					lowTail.next = e
				}
				lowTail = e
				lc++
			} else {
				e.prev = highTail
				if highTail == nil {
					high = e
				} else {
					highTail.next = e
				}
				highTail = e
				hc++
			}
			e = next
		}
		m.table[i] = low
		m.table[i+capacity] = high
		if tree {
			if low != nil {
				if lc <= 6 {
					mailParameterUntreeify(low)
				} else if high != nil {
					m.treeify(i)
				}
			}
			if high != nil {
				if hc <= 6 {
					mailParameterUntreeify(high)
				} else if low != nil {
					m.treeify(i + capacity)
				}
			}
		}
	}
}
func (m *mailParameterHashOrder) remove(name string) {
	if len(m.table) == 0 {
		return
	}
	index := int(parameterHash(name) & uint32(len(m.table)-1))
	var pred *mailParameterHashNode
	node := m.table[index]
	for node != nil && node.name != name {
		pred = node
		node = node.next
	}
	if node == nil {
		return
	}
	if node.tree {
		m.removeTree(index, node)
	} else if pred == nil {
		m.table[index] = node.next
	} else {
		pred.next = node.next
	}
	m.size--
}
func (m *mailParameterHashOrder) removeTree(index int, p *mailParameterHashNode) {
	first := m.table[index]
	root := first
	succ, pred := p.next, p.prev
	if pred == nil {
		first = succ
		m.table[index] = first
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
	if root.right == nil || root.left == nil || root.left.left == nil {
		mailParameterUntreeify(first)
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
	m.front(r)
}
func mailParameterRotateLeft(root, p *mailParameterHashNode) *mailParameterHashNode {
	if p != nil && p.right != nil {
		r := p.right
		p.right = r.left
		if r.left != nil {
			r.left.parent = p
		}
		r.parent = p.parent
		if p.parent == nil {
			root = r
			root.red = false
		} else if p.parent.left == p {
			p.parent.left = r
		} else {
			p.parent.right = r
		}
		r.left = p
		p.parent = r
	}
	return root
}
func mailParameterRotateRight(root, p *mailParameterHashNode) *mailParameterHashNode {
	if p != nil && p.left != nil {
		l := p.left
		p.left = l.right
		if l.right != nil {
			l.right.parent = p
		}
		l.parent = p.parent
		if p.parent == nil {
			root = l
			root.red = false
		} else if p.parent.right == p {
			p.parent.right = l
		} else {
			p.parent.left = l
		}
		l.right = p
		p.parent = l
	}
	return root
}
func mailParameterBalanceInsertion(root, x *mailParameterHashNode) *mailParameterHashNode {
	x.red = true
	for {
		xp := x.parent
		if xp == nil {
			x.red = false
			return x
		}
		xpp := xp.parent
		if !xp.red || xpp == nil {
			return root
		}
		xppl, xppr := xpp.left, xpp.right
		if xp == xppl {
			if xppr != nil && xppr.red {
				xppr.red = false
				xp.red = false
				xpp.red = true
				x = xpp
			} else {
				if x == xp.right {
					x = xp
					root = mailParameterRotateLeft(root, x)
					xp = x.parent
					xpp = nil
					if xp != nil {
						xpp = xp.parent
					}
				}
				if xp != nil {
					xp.red = false
					if xpp != nil {
						xpp.red = true
						root = mailParameterRotateRight(root, xpp)
					}
				}
			}
		} else {
			if xppl != nil && xppl.red {
				xppl.red = false
				xp.red = false
				xpp.red = true
				x = xpp
			} else {
				if x == xp.left {
					x = xp
					root = mailParameterRotateRight(root, x)
					xp = x.parent
					xpp = nil
					if xp != nil {
						xpp = xp.parent
					}
				}
				if xp != nil {
					xp.red = false
					if xpp != nil {
						xpp.red = true
						root = mailParameterRotateLeft(root, xpp)
					}
				}
			}
		}
	}
}
func mailParameterBalanceDeletion(root, x *mailParameterHashNode) *mailParameterHashNode {
	for {
		if x == nil || x == root {
			return root
		}
		xp := x.parent
		if xp == nil {
			x.red = false
			return x
		}
		if x.red {
			x.red = false
			return root
		}
		xpl, xpr := xp.left, xp.right
		if xpl == x {
			if xpr != nil && xpr.red {
				xpr.red = false
				xp.red = true
				root = mailParameterRotateLeft(root, xp)
				xp = x.parent
				xpr = nil
				if xp != nil {
					xpr = xp.right
				}
			}
			if xpr == nil {
				x = xp
			} else {
				sl, sr := xpr.left, xpr.right
				if (sr == nil || !sr.red) && (sl == nil || !sl.red) {
					xpr.red = true
					x = xp
				} else {
					if sr == nil || !sr.red {
						if sl != nil {
							sl.red = false
						}
						xpr.red = true
						root = mailParameterRotateRight(root, xpr)
						xp = x.parent
						xpr = nil
						if xp != nil {
							xpr = xp.right
						}
					}
					if xpr != nil {
						xpr.red = xp != nil && xp.red
						sr = xpr.right
						if sr != nil {
							sr.red = false
						}
					}
					if xp != nil {
						xp.red = false
						root = mailParameterRotateLeft(root, xp)
					}
					x = root
				}
			}
		} else {
			if xpl != nil && xpl.red {
				xpl.red = false
				xp.red = true
				root = mailParameterRotateRight(root, xp)
				xp = x.parent
				xpl = nil
				if xp != nil {
					xpl = xp.left
				}
			}
			if xpl == nil {
				x = xp
			} else {
				sl, sr := xpl.left, xpl.right
				if (sl == nil || !sl.red) && (sr == nil || !sr.red) {
					xpl.red = true
					x = xp
				} else {
					if sl == nil || !sl.red {
						if sr != nil {
							sr.red = false
						}
						xpl.red = true
						root = mailParameterRotateLeft(root, xpl)
						xp = x.parent
						xpl = nil
						if xp != nil {
							xpl = xp.left
						}
					}
					if xpl != nil {
						xpl.red = xp != nil && xp.red
						sl = xpl.left
						if sl != nil {
							sl.red = false
						}
					}
					if xp != nil {
						xp.red = false
						root = mailParameterRotateRight(root, xp)
					}
					x = root
				}
			}
		}
	}
}

// HashMap(Map) and putAll pre-size from the source map's size before visiting
// its entries. Reusing ordinary insertion alone changes observable bucket order.
func copyMailParameterHashTable[V any](source *mailParameterHashTable[V]) *mailParameterHashTable[V] {
	result := newMailParameterHashTable[V]()
	result.putAll(source)
	return result
}
func (m *mailParameterHashTable[V]) putAll(source *mailParameterHashTable[V]) {
	if source == nil {
		panic(NewNullPointerException())
	}
	size := source.items.Len()
	if size == 0 {
		return
	}
	if len(m.order.table) == 0 {
		// OpenJDK 21 uses ceil(size / (double)loadFactor), then tableSizeFor.
		required := (int64(size)*4 + 2) / 3
		capacity := 1
		for int64(capacity) < required && capacity < 1<<30 {
			capacity <<= 1
		}
		m.order.table = make([]*mailParameterHashNode, capacity)
	} else {
		for size > len(m.order.table)*3/4 && len(m.order.table) < 1<<30 {
			m.order.resize()
		}
	}
	for _, name := range source.names() {
		m.set(name, source.items.Get(name))
	}
}
