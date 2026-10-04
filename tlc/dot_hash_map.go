// Copyright (c) 1997, 2023, Oracle and/or its affiliates. All rights reserved.
// Derived from OpenJDK 21 java.util.HashMap; GPL v2 with Classpath exception.
// See licenses/openjdk-LICENSE and licenses/openjdk-ADDITIONAL_LICENSE_INFO.
package tlc

// dotHashMap implements the HashMap operations used by DotStateWriter. Its
// Integer, Long and String keys are all Comparable; no identity tie-break is
// needed. Bucket lists, resize splitting and red-black trees retain Java's
// iteration order. DOT owns the synchronization and never removes entries.
type dotHashMap[K comparable, V any] struct {
	table           []*dotHashNode[K, V]
	size, threshold int
	hashCode        func(K) int32
	compare         func(K, K) int
}

type dotHashNode[K comparable, V any] struct {
	hash                            uint32
	key                             K
	value                           V
	next, prev, parent, left, right *dotHashNode[K, V]
	tree, red                       bool
}

func newDotHashMap[K comparable, V any](hashCode func(K) int32, compare func(K, K) int) *dotHashMap[K, V] {
	return &dotHashMap[K, V]{hashCode: hashCode, compare: compare}
}

func (m *dotHashMap[K, V]) hash(key K) uint32 {
	h := uint32(m.hashCode(key))
	return h ^ h>>16
}

func (m *dotHashMap[K, V]) Len() int { return m.size }

func (m *dotHashMap[K, V]) Get2(key K) (V, bool) {
	if len(m.table) != 0 {
		h := m.hash(key)
		p := m.table[int(h)&(len(m.table)-1)]
		if p != nil && p.tree {
			for p.parent != nil {
				p = p.parent
			}
			for p != nil {
				if p.hash == h && p.key == key {
					return p.value, true
				}
				if m.direction(h, key, p) <= 0 {
					p = p.left
				} else {
					p = p.right
				}
			}
		}
		for ; p != nil; p = p.next {
			if p.hash == h && p.key == key {
				return p.value, true
			}
		}
	}
	var zero V
	return zero, false
}

func (m *dotHashMap[K, V]) Set(key K, value V) {
	if len(m.table) == 0 {
		m.resize()
	}
	h := m.hash(key)
	i := int(h) & (len(m.table) - 1)
	p := m.table[i]
	if p == nil {
		m.table[i] = &dotHashNode[K, V]{hash: h, key: key, value: value}
	} else if p.tree {
		if old := m.putTree(p, h, key, value); old != nil {
			old.value = value
			return
		}
	} else {
		for count := 0; ; count++ {
			if p.hash == h && p.key == key {
				p.value = value
				return
			}
			if p.next == nil {
				p.next = &dotHashNode[K, V]{hash: h, key: key, value: value}
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
func (m *dotHashMap[K, V]) getOrCreate(key K, create func() V) V {
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
		m.table[i] = &dotHashNode[K, V]{hash: h, key: key, value: value, next: first}
		if count >= 7 {
			m.treeifyBin(i)
		}
	}
	m.size++
	return value
}

func (m *dotHashMap[K, V]) All() func(func(K, V) bool) {
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

func (m *dotHashMap[K, V]) resize() {
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
	m.table = make([]*dotHashNode[K, V], capacity)
	for i, first := range old {
		if first == nil {
			continue
		}
		if first.next == nil {
			m.table[int(first.hash)&(capacity-1)] = first
			continue
		}
		wasTree := first.tree
		var heads, tails [2]*dotHashNode[K, V]
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
					m.table[index] = dotUntreeify(head)
				} else if heads[1-part] != nil {
					m.treeify(index)
				}
			}
		}
	}
}

func dotUntreeify[K comparable, V any](head *dotHashNode[K, V]) *dotHashNode[K, V] {
	var first, last *dotHashNode[K, V]
	for p := head; p != nil; p = p.next {
		node := &dotHashNode[K, V]{hash: p.hash, key: p.key, value: p.value}
		if last == nil {
			first = node
		} else {
			last.next = node
		}
		last = node
	}
	return first
}

func (m *dotHashMap[K, V]) treeifyBin(index int) {
	if len(m.table) < 64 {
		m.resize()
		return
	}
	var previous *dotHashNode[K, V]
	for p := m.table[index]; p != nil; p = p.next {
		p.tree = true
		p.prev = previous
		previous = p
	}
	m.treeify(index)
}

func (m *dotHashMap[K, V]) direction(hash uint32, key K, p *dotHashNode[K, V]) int {
	if int32(hash) < int32(p.hash) {
		return -1
	}
	if int32(hash) > int32(p.hash) {
		return 1
	}
	return m.compare(key, p.key)
}

func (m *dotHashMap[K, V]) treeify(index int) {
	var root *dotHashNode[K, V]
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
				root = dotBalanceInsertion(root, x)
				break
			}
		}
	}
	m.moveRootToFront(root)
}

func (m *dotHashMap[K, V]) putTree(first *dotHashNode[K, V], hash uint32, key K, value V) *dotHashNode[K, V] {
	root := first
	for root.parent != nil {
		root = root.parent
	}
	for p := root; ; {
		if hash == p.hash && key == p.key {
			return p
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
			x := &dotHashNode[K, V]{hash: hash, key: key, value: value, next: next, parent: parent, prev: parent, tree: true}
			if dir <= 0 {
				parent.left = x
			} else {
				parent.right = x
			}
			parent.next = x
			if next != nil {
				next.prev = x
			}
			m.moveRootToFront(dotBalanceInsertion(root, x))
			return nil
		}
	}
}

func (m *dotHashMap[K, V]) moveRootToFront(root *dotHashNode[K, V]) {
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

func dotRotateLeft[K comparable, V any](root, p *dotHashNode[K, V]) *dotHashNode[K, V] {
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

func dotRotateRight[K comparable, V any](root, p *dotHashNode[K, V]) *dotHashNode[K, V] {
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

func dotBalanceInsertion[K comparable, V any](root, x *dotHashNode[K, V]) *dotHashNode[K, V] {
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
				root = dotRotateLeft(root, x)
				parent = x.parent
				grand = parent.parent
			}
			parent.red = false
			if grand != nil {
				grand.red = true
				root = dotRotateRight(root, grand)
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
				root = dotRotateRight(root, x)
				parent = x.parent
				grand = parent.parent
			}
			parent.red = false
			if grand != nil {
				grand.red = true
				root = dotRotateLeft(root, grand)
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

func newDotLongSet() *dotHashMap[uint64, struct{}] {
	return newDotHashMap[uint64, struct{}](func(key uint64) int32 { return int32(key ^ (key >> 32)) }, dotCompareLong)
}

type dotLongSet = dotHashMap[uint64, struct{}]
