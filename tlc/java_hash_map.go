// Copyright (c) 1997, 2023, Oracle and/or its affiliates. All rights reserved.
// Derived from OpenJDK 21 java.util.HashMap; GPL v2 with Classpath exception.
// See licenses/openjdk-LICENSE and licenses/openjdk-ADDITIONAL_LICENSE_INFO.
package tlc

// javaHashMap implements the HashMap operations used by DOT, state values and TLCCache.
// Bucket lists, resizing and red-black trees retain source iteration order.
// Comparable keys use compare; non-Comparable keys supply an identity tie-break.
// The optional equality callback ports object equals beyond native identity.
// Callers own synchronization. Semantic UID sets also remove entries.
type javaHashMap[K comparable, V any] struct {
	table           []*javaHashNode[K, V]
	size, threshold int
	hashCode        func(K) int32
	equal           func(K, K) bool
	compare         func(K, K) int
	tieBreak        func(K, K) int
}

type javaHashNode[K comparable, V any] struct {
	hash                            uint32
	key                             K
	value                           V
	next, prev, parent, left, right *javaHashNode[K, V]
	tree, red                       bool
}

func newJavaHashMap[K comparable, V any](hashCode func(K) int32, compare func(K, K) int) *javaHashMap[K, V] {
	return &javaHashMap[K, V]{hashCode: hashCode, compare: compare}
}

// HashMap invokes the lookup key's equals only after hashes match.
func (m *javaHashMap[K, V]) keysEqual(key, stored K) bool {
	return key == stored || (m.equal != nil && m.equal(key, stored))
}

func (m *javaHashMap[K, V]) hash(key K) uint32 {
	h := uint32(m.hashCode(key))
	return h ^ h>>16
}

func (m *javaHashMap[K, V]) Len() int { return m.size }

func (m *javaHashMap[K, V]) Get2(key K) (V, bool) {
	if len(m.table) != 0 {
		h := m.hash(key)
		p := m.table[int(h)&(len(m.table)-1)]
		if p != nil && p.tree {
			for p.parent != nil {
				p = p.parent
			}
			if found := m.findTree(p, h, key); found != nil {
				return found.value, true
			}
			var zero V
			return zero, false
		}
		for ; p != nil; p = p.next {
			if p.hash == h && m.keysEqual(key, p.key) {
				return p.value, true
			}
		}
	}
	var zero V
	return zero, false
}

// TreeNode.find searches both subtrees for equal-hash non-Comparable keys.
// Identity hashes guide insertion but cannot guide equality lookup: they may collide.
func (m *javaHashMap[K, V]) findTree(p *javaHashNode[K, V], hash uint32, key K) *javaHashNode[K, V] {
	for p != nil {
		left, right := p.left, p.right
		switch {
		case int32(p.hash) > int32(hash):
			p = left
		case int32(p.hash) < int32(hash):
			p = right
		case m.keysEqual(key, p.key):
			return p
		case left == nil:
			p = right
		case right == nil:
			p = left
		default:
			if m.compare != nil {
				if dir := m.compare(key, p.key); dir != 0 {
					if dir < 0 {
						p = left
					} else {
						p = right
					}
					continue
				}
			}
			if found := m.findTree(right, hash, key); found != nil {
				return found
			}
			p = left
		}
	}
	return nil
}

func (m *javaHashMap[K, V]) Set(key K, value V) {
	// HashMap.put evaluates hash(key) before putVal can allocate its table.
	h := m.hash(key)
	if len(m.table) == 0 {
		m.resize()
	}
	i := int(h) & (len(m.table) - 1)
	p := m.table[i]
	if p == nil {
		m.table[i] = &javaHashNode[K, V]{hash: h, key: key, value: value}
	} else if p.tree {
		if old := m.putTree(p, h, key, value); old != nil {
			old.value = value
			return
		}
	} else {
		for count := 0; ; count++ {
			if p.hash == h && m.keysEqual(key, p.key) {
				p.value = value
				return
			}
			if p.next == nil {
				p.next = &javaHashNode[K, V]{hash: h, key: key, value: value}
				if count >= 7 {
					m.treeifyBin(i)
				}
				break
			}
			p = p.next
		}
	}
	m.size++
	if m.size > m.threshold {
		m.resize()
	}
}

// DotStateWriter.computeIfAbsent always constructs a non-null HashSet. Java
// prepends this mapping, and checks the resize threshold before insertion.
func (m *javaHashMap[K, V]) getOrCreate(key K, create func() V) V {
	if m.size > m.threshold || len(m.table) == 0 {
		m.resize()
	}
	if value, ok := m.Get2(key); ok {
		return value
	}
	h := m.hash(key)
	i := int(h) & (len(m.table) - 1)
	first := m.table[i]
	value := create()
	if first != nil && first.tree {
		m.putTree(first, h, key, value)
	} else {
		count := 0
		for p := first; p != nil; p = p.next {
			count++
		}
		m.table[i] = &javaHashNode[K, V]{hash: h, key: key, value: value, next: first}
		if count >= 7 {
			m.treeifyBin(i)
		}
	}
	m.size++
	return value
}

// mergeNonNull implements HashMap.merge for non-null values and a combiner
// that neither removes entries nor modifies this map. _Possible sums integers.
// Like computeIfAbsent, merge prepends new list nodes and resizes before lookup.
func (m *javaHashMap[K, V]) mergeNonNull(key K, value V, combine func(V, V) V) V {
	h := m.hash(key)
	if m.size > m.threshold || len(m.table) == 0 {
		m.resize()
	}
	i := int(h) & (len(m.table) - 1)
	first := m.table[i]
	var old *javaHashNode[K, V]
	count := 0
	if first != nil && first.tree {
		root := first
		for root.parent != nil {
			root = root.parent
		}
		old = m.findTree(root, h, key)
	} else {
		for p := first; p != nil; p = p.next {
			if p.hash == h && m.keysEqual(key, p.key) {
				old = p
				break
			}
			count++
		}
	}
	if old != nil {
		old.value = combine(old.value, value)
		return old.value
	}
	if first != nil && first.tree {
		m.putTree(first, h, key, value)
	} else {
		m.table[i] = &javaHashNode[K, V]{hash: h, key: key, value: value, next: first}
		if count >= 7 {
			m.treeifyBin(i)
		}
	}
	m.size++
	return value
}

func (m *javaHashMap[K, V]) All() func(func(K, V) bool) {
	return func(yield func(K, V) bool) {
		if m == nil {
			return
		}
		for _, first := range m.table {
			for p := first; p != nil; p = p.next {
				if !yield(p.key, p.value) {
					return
				}
			}
		}
	}
}

func (m *javaHashMap[K, V]) resize() {
	old := m.table
	n := len(old)
	if n >= 1<<30 {
		m.threshold = 1<<31 - 1
		return
	}
	capacity := 16
	if n != 0 {
		capacity = n * 2
	} else if m.threshold > 0 {
		// HashMap(Map) stores its initial capacity in threshold until allocation.
		capacity = m.threshold
	}
	if capacity < 1<<30 {
		m.threshold = capacity * 3 / 4
	} else {
		m.threshold = 1<<31 - 1
	}
	m.table = make([]*javaHashNode[K, V], capacity)
	for i, first := range old {
		if first == nil {
			continue
		}
		if first.next == nil {
			m.table[int(first.hash)&(capacity-1)] = first
			continue
		}
		wasTree := first.tree
		var heads, tails [2]*javaHashNode[K, V]
		var counts [2]int
		for p := first; p != nil; {
			next := p.next
			part := 0
			if p.hash&uint32(n) != 0 {
				part = 1
			}
			p.next = nil
			if wasTree {
				p.prev = tails[part]
			}
			if tails[part] == nil {
				heads[part] = p
			} else {
				tails[part].next = p
			}
			tails[part] = p
			counts[part]++
			p = next
		}
		for part, head := range heads {
			if head == nil {
				continue
			}
			index := i + part*n
			m.table[index] = head
			if wasTree {
				if counts[part] <= 6 {
					m.table[index] = javaHashUntreeify(head)
				} else if heads[1-part] != nil {
					m.treeify(index)
				}
			}
		}
	}
}

func javaHashUntreeify[K comparable, V any](head *javaHashNode[K, V]) *javaHashNode[K, V] {
	var first, last *javaHashNode[K, V]
	for p := head; p != nil; p = p.next {
		node := &javaHashNode[K, V]{hash: p.hash, key: p.key, value: p.value}
		if last == nil {
			first = node
		} else {
			last.next = node
		}
		last = node
	}
	return first
}

func (m *javaHashMap[K, V]) treeifyBin(index int) {
	if len(m.table) < 64 {
		m.resize()
		return
	}
	var previous *javaHashNode[K, V]
	for p := m.table[index]; p != nil; p = p.next {
		p.tree = true
		p.prev = previous
		previous = p
	}
	m.treeify(index)
}

func (m *javaHashMap[K, V]) direction(hash uint32, key K, p *javaHashNode[K, V]) int {
	if int32(hash) < int32(p.hash) {
		return -1
	}
	if int32(hash) > int32(p.hash) {
		return 1
	}
	if m.compare != nil {
		if dir := m.compare(key, p.key); dir != 0 {
			return dir
		}
	}
	return m.tieBreak(key, p.key)
}

func (m *javaHashMap[K, V]) treeify(index int) {
	var root *javaHashNode[K, V]
	for x := m.table[index]; x != nil; x = x.next {
		x.left, x.right = nil, nil
		if root == nil {
			x.parent = nil
			x.red = false
			root = x
			continue
		}
		for p := root; ; {
			dir := m.direction(x.hash, x.key, p)
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
				root = javaHashBalanceInsertion(root, x)
				break
			}
		}
	}
	m.moveRootToFront(root)
}

func (m *javaHashMap[K, V]) putTree(first *javaHashNode[K, V], hash uint32, key K, value V) *javaHashNode[K, V] {
	root := first
	for root.parent != nil {
		root = root.parent
	}
	searched := false
	for p := root; ; {
		if hash == p.hash && m.keysEqual(key, p.key) {
			return p
		}
		if hash == p.hash && (m.compare == nil || m.compare(key, p.key) == 0) && !searched {
			searched = true
			if found := m.findTree(p.left, hash, key); found != nil {
				return found
			}
			if found := m.findTree(p.right, hash, key); found != nil {
				return found
			}
		}
		dir := m.direction(hash, key, p)
		parent := p
		if dir <= 0 {
			p = p.left
		} else {
			p = p.right
		}
		if p == nil {
			next := parent.next
			x := &javaHashNode[K, V]{hash: hash, key: key, value: value, next: next, parent: parent, prev: parent, tree: true}
			if dir <= 0 {
				parent.left = x
			} else {
				parent.right = x
			}
			parent.next = x
			if next != nil {
				next.prev = x
			}
			m.moveRootToFront(javaHashBalanceInsertion(root, x))
			return nil
		}
	}
}

func (m *javaHashMap[K, V]) moveRootToFront(root *javaHashNode[K, V]) {
	index := int(root.hash) & (len(m.table) - 1)
	first := m.table[index]
	if first == root {
		return
	}
	m.table[index] = root
	previous, next := root.prev, root.next
	if next != nil {
		next.prev = previous
	}
	if previous != nil {
		previous.next = next
	}
	if first != nil {
		first.prev = root
	}
	root.next, root.prev = first, nil
}

func javaHashRotateLeft[K comparable, V any](root, p *javaHashNode[K, V]) *javaHashNode[K, V] {
	if p != nil && p.right != nil {
		r := p.right
		p.right = r.left
		if r.left != nil {
			r.left.parent = p
		}
		r.parent = p.parent
		if p.parent == nil {
			root = r
			r.red = false
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

func javaHashRotateRight[K comparable, V any](root, p *javaHashNode[K, V]) *javaHashNode[K, V] {
	if p != nil && p.left != nil {
		l := p.left
		p.left = l.right
		if l.right != nil {
			l.right.parent = p
		}
		l.parent = p.parent
		if p.parent == nil {
			root = l
			l.red = false
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

func javaHashBalanceInsertion[K comparable, V any](root, x *javaHashNode[K, V]) *javaHashNode[K, V] {
	x.red = true
	for {
		parent := x.parent
		if parent == nil {
			x.red = false
			return x
		}
		grand := parent.parent
		if !parent.red || grand == nil {
			return root
		}
		if parent == grand.left {
			uncle := grand.right
			if uncle != nil && uncle.red {
				uncle.red = false
				parent.red = false
				grand.red = true
				x = grand
				continue
			}
			if x == parent.right {
				x = parent
				root = javaHashRotateLeft(root, x)
				parent = x.parent
				grand = parent.parent
			}
			parent.red = false
			if grand != nil {
				grand.red = true
				root = javaHashRotateRight(root, grand)
			}
		} else {
			uncle := grand.left
			if uncle != nil && uncle.red {
				uncle.red = false
				parent.red = false
				grand.red = true
				x = grand
				continue
			}
			if x == parent.left {
				x = parent
				root = javaHashRotateRight(root, x)
				parent = x.parent
				grand = parent.parent
			}
			parent.red = false
			if grand != nil {
				grand.red = true
				root = javaHashRotateLeft(root, grand)
			}
		}
	}
}

func dotCompareInt(a, b int) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

func dotCompareLong(a, b uint64) int {
	if int64(a) < int64(b) {
		return -1
	}
	if int64(a) > int64(b) {
		return 1
	}
	return 0
}

func dotCompareString(a, b string) int {
	aa, bb := javaStringUTF16(a), javaStringUTF16(b)
	for i := 0; i < len(aa) && i < len(bb); i++ {
		if aa[i] != bb[i] {
			return int(aa[i]) - int(bb[i])
		}
	}
	return len(aa) - len(bb)
}

func newDotLongSet() *javaHashMap[uint64, struct{}] {
	return newJavaHashMap[uint64, struct{}](func(key uint64) int32 { return int32(key ^ (key >> 32)) }, dotCompareLong)
}

type dotLongSet = javaHashMap[uint64, struct{}]

// Remove ports removeNode(hash, key, null, false, true). These callers do
// not request value matching or iterator removal with an immovable tree root.
func (m *javaHashMap[K, V]) Remove(key K) bool {
	hash := m.hash(key)
	if len(m.table) == 0 {
		return false
	}
	index := int(hash) & (len(m.table) - 1)
	p := m.table[index]
	if p == nil {
		return false
	}
	var node *javaHashNode[K, V]
	if p.hash == hash && m.keysEqual(key, p.key) {
		node = p
	} else if p.next != nil {
		if p.tree {
			root := p
			for root.parent != nil {
				root = root.parent
			}
			node = m.findTree(root, hash, key)
		} else {
			for e := p.next; e != nil; e = e.next {
				if e.hash == hash && m.keysEqual(key, e.key) {
					node = e
					break
				}
				p = e
			}
		}
	}
	if node == nil {
		return false
	}
	if node.tree {
		m.removeTreeNode(node)
	} else if node == p {
		m.table[index] = node.next
	} else {
		p.next = node.next
	}
	m.size--
	return true
}

func (m *javaHashMap[K, V]) removeTreeNode(p *javaHashNode[K, V]) {
	index := int(p.hash) & (len(m.table) - 1)
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
		m.table[index] = javaHashUntreeify(first)
		return
	}
	pl, pr := p.left, p.right
	var replacement *javaHashNode[K, V]
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
		replacement.parent = p.parent
		pp := p.parent
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
		r = javaHashBalanceDeletion(root, replacement)
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
	m.moveRootToFront(r)
}

func javaHashBalanceDeletion[K comparable, V any](root, x *javaHashNode[K, V]) *javaHashNode[K, V] {
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
				root = javaHashRotateLeft(root, xp)
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
						root = javaHashRotateRight(root, xpr)
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
						root = javaHashRotateLeft(root, xp)
					}
					x = root
				}
			}
		} else {
			if xpl != nil && xpl.red {
				xpl.red = false
				xp.red = true
				root = javaHashRotateRight(root, xp)
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
						root = javaHashRotateLeft(root, xpl)
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
						root = javaHashRotateRight(root, xp)
					}
					x = root
				}
			}
		}
	}
}
