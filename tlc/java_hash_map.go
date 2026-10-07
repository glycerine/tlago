// Copyright (c) 1997, 2023, Oracle and/or its affiliates. All rights reserved.
// Derived from OpenJDK 21 java.util.HashMap; GPL v2 with Classpath exception.
// See licenses/openjdk-LICENSE and licenses/openjdk-ADDITIONAL_LICENSE_INFO.
package tlc

// javaHashMap implements the HashMap operations used by DOT, state values and TLCCache.
// Bucket lists, resizing and red-black trees retain source iteration order.
// Comparable keys use compare; non-Comparable keys supply an identity tie-break.
// The optional equality callback ports object equals beyond native identity.
// Callers own synchronization. These uses never remove entries.
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
	if len(m.table) == 0 {
		m.resize()
	}
	h := m.hash(key)
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
