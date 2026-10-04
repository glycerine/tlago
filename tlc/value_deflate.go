/* -*-mode:java; c-basic-offset:2; -*- */
/*
Copyright (c) 2000-2011 ymnk, JCraft,Inc. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:

  1. Redistributions of source code must retain the above copyright notice,
     this list of conditions and the following disclaimer.

  2. Redistributions in binary form must reproduce the above copyright
     notice, this list of conditions and the following disclaimer in
     the documentation and/or other materials provided with the distribution.

  3. The names of the authors may not be used to endorse or promote products
     derived from this software without specific prior written permission.

THIS SOFTWARE IS PROVIDED ``AS IS'' AND ANY EXPRESSED OR IMPLIED WARRANTIES,
INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND
FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL JCRAFT,
INC. OR ANY CONTRIBUTORS TO THIS SOFTWARE BE LIABLE FOR ANY DIRECT, INDIRECT,
INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT
LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA,
OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE,
EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/
/*
 * This program is based on zlib-1.1.3, so all credit should go authors
 * Jean-loup Gailly(jloup@gzip.org) and Mark Adler(madler@alumni.caltech.edu)
 * and contributors of zlib.
 */

// Algorithm source: JZlib Deflate.java and Tree.java at
// https://github.com/ymnk/jzlib/tree/a21be20213d66eff15904d925e9b721956a01ef7
// This port exposes the default raw-deflate path used by TLC gzip streams.
package tlc

import (
	"encoding/binary"
	"hash/crc32"
	"io"
)

// zlibGzipWriter uses the default zlib lazy matcher and Huffman construction
// ported from JZlib. Go's compress/flate terminates with an additional stored
// block, unlike Java's GZIPOutputStream. The input window and token buffers are
// bounded independently of the size of the serialized value.
type zlibGzipWriter struct {
	out    io.Writer
	def    zlibDeflater
	crc    uint32
	size   uint32
	err    error
	closed bool
}

func newZlibGzipWriter(out io.Writer) *zlibGzipWriter {
	w := &zlibGzipWriter{out: out}
	w.def.init(w.emit)
	// java.util.zip.GZIPOutputStream.writeHeader: deflate, no flags, zero
	// modification time, no extra flags, unknown operating system.
	w.emit([]byte{0x1f, 0x8b, 8, 0, 0, 0, 0, 0, 0, 255})
	return w
}
func (w *zlibGzipWriter) emit(p []byte) {
	if w.err != nil {
		return
	}
	n, err := w.out.Write(p)
	if err == nil && n != len(p) {
		err = io.ErrShortWrite
	}
	w.err = err
}
func (w *zlibGzipWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	if w.closed {
		return 0, io.ErrClosedPipe
	}
	w.crc = crc32.Update(w.crc, crc32.IEEETable, p)
	w.size += uint32(len(p))
	w.def.input = p
	w.def.slow(false)
	return len(p), w.err
}
func (w *zlibGzipWriter) Close() error {
	if w.closed {
		return w.err
	}
	w.closed = true
	if w.err != nil {
		return w.err
	}
	w.def.slow(true)
	var trailer [8]byte
	binary.LittleEndian.PutUint32(trailer[:4], w.crc)
	binary.LittleEndian.PutUint32(trailer[4:], w.size)
	w.emit(trailer[:])
	return w.err
}

const (
	zlibWindow        = 32768
	zlibMinMatch      = 3
	zlibMaxMatch      = 258
	zlibMinLookahead  = zlibMaxMatch + zlibMinMatch + 1
	zlibLCodes        = 286
	zlibDCodes        = 30
	zlibBLCodes       = 19
	zlibHeapSize      = 2*zlibLCodes + 1
	zlibLiteralBuffer = 16384
)

type zlibTree struct {
	tree                            []int16
	static                          []int
	extra                           []int
	base, elems, maxLength, maxCode int
}
type zlibDeflater struct {
	window                                                         [2 * zlibWindow]byte
	prev, head                                                     [zlibWindow]uint16
	input                                                          []byte
	strstart, blockStart, lookahead, insH                          int
	matchLength, prevLength, matchStart, prevMatch, matchAvailable int
	literals                                                       [zlibLiteralBuffer]byte
	distances                                                      [zlibLiteralBuffer]uint16
	lastLit, matches                                               int
	ltree                                                          [2 * zlibHeapSize]int16
	dtree                                                          [2 * (2*zlibDCodes + 1)]int16
	bltree                                                         [2 * (2*zlibBLCodes + 1)]int16
	ldesc, ddesc, bldesc                                           zlibTree
	heap                                                           [zlibHeapSize]int
	depth                                                          [zlibHeapSize]byte
	heapLen, heapMax                                               int
	blCount, nextCode                                              [16]int16
	optLen, staticLen                                              int
	bitBuffer                                                      uint16
	bitValid                                                       int
	pending                                                        []byte
	emit                                                           func([]byte)
}

func (s *zlibDeflater) init(emit func([]byte)) {
	s.emit = emit
	s.pending = make([]byte, 0, 65536)
	s.matchLength = 2
	s.prevLength = 2
	s.ldesc = zlibTree{tree: s.ltree[:], static: zlibStaticLtree[:], extra: zlibExtraLbits[:], base: 257, elems: zlibLCodes, maxLength: 15}
	s.ddesc = zlibTree{tree: s.dtree[:], static: zlibStaticDtree[:], extra: zlibExtraDbits[:], elems: zlibDCodes, maxLength: 15}
	s.bldesc = zlibTree{tree: s.bltree[:], extra: zlibExtraBlbits[:], elems: zlibBLCodes, maxLength: 7}
	s.initBlock()
}
func (s *zlibDeflater) initBlock() {
	for i := 0; i < zlibLCodes; i++ {
		s.ltree[2*i] = 0
	}
	for i := 0; i < zlibDCodes; i++ {
		s.dtree[2*i] = 0
	}
	for i := 0; i < zlibBLCodes; i++ {
		s.bltree[2*i] = 0
	}
	s.ltree[512] = 1
	s.optLen = 0
	s.staticLen = 0
	s.lastLit = 0
	s.matches = 0
}
func (s *zlibDeflater) insert() int {
	s.insH = ((s.insH << 5) ^ int(s.window[s.strstart+2])) & (zlibWindow - 1)
	h := s.head[s.insH]
	s.prev[s.strstart&(zlibWindow-1)] = h
	s.head[s.insH] = uint16(s.strstart)
	return int(h)
}
func (s *zlibDeflater) fillWindow() {
	for {
		more := len(s.window) - s.lookahead - s.strstart
		if s.strstart >= 2*zlibWindow-zlibMinLookahead {
			copy(s.window[:zlibWindow], s.window[zlibWindow:])
			s.matchStart -= zlibWindow
			s.strstart -= zlibWindow
			s.blockStart -= zlibWindow
			for i, m := range s.head {
				if int(m) >= zlibWindow {
					s.head[i] = m - zlibWindow
				} else {
					s.head[i] = 0
				}
			}
			for i, m := range s.prev {
				if int(m) >= zlibWindow {
					s.prev[i] = m - zlibWindow
				} else {
					s.prev[i] = 0
				}
			}
			more += zlibWindow
		}
		if len(s.input) == 0 {
			return
		}
		n := min(more, len(s.input))
		copy(s.window[s.strstart+s.lookahead:], s.input[:n])
		s.input = s.input[n:]
		s.lookahead += n
		if s.lookahead >= zlibMinMatch {
			s.insH = ((int(s.window[s.strstart]) << 5) ^ int(s.window[s.strstart+1])) & (zlibWindow - 1)
		}
		if s.lookahead >= zlibMinLookahead || len(s.input) == 0 {
			return
		}
	}
}
func (s *zlibDeflater) longestMatch(curMatch int) int {
	chain := 128
	scan := s.strstart
	best := s.prevLength
	limit := max(0, s.strstart-(zlibWindow-zlibMinLookahead))
	nice := min(128, s.lookahead)
	if s.prevLength >= 8 {
		chain >>= 2
	}
	end := s.strstart + zlibMaxMatch
	end1, endByte := s.window[scan+best-1], s.window[scan+best]
	for {
		m := curMatch
		if s.window[m+best] == endByte && s.window[m+best-1] == end1 && s.window[m] == s.window[scan] && s.window[m+1] == s.window[scan+1] {
			scan += 2
			m += 2
			// Preserve the source's eight comparisons before its strend check.
			for {
				equal := true
				for i := 0; i < 8; i++ {
					scan++
					m++
					if s.window[scan] != s.window[m] {
						equal = false
						break
					}
				}
				if !equal || scan >= end {
					break
				}
			}
			length := zlibMaxMatch - (end - scan)
			scan = s.strstart
			if length > best {
				s.matchStart = curMatch
				best = length
				if length >= nice {
					break
				}
				end1 = s.window[scan+best-1]
				endByte = s.window[scan+best]
			}
		}
		curMatch = int(s.prev[curMatch&(zlibWindow-1)])
		chain--
		if curMatch <= limit || chain == 0 {
			break
		}
	}
	return min(best, s.lookahead)
}
func (s *zlibDeflater) slow(finish bool) {
	for {
		if s.lookahead < zlibMinLookahead {
			s.fillWindow()
			if s.lookahead < zlibMinLookahead && !finish {
				return
			}
			if s.lookahead == 0 {
				break
			}
		}
		hashHead := 0
		if s.lookahead >= zlibMinMatch {
			hashHead = s.insert()
		}
		s.prevLength = s.matchLength
		s.prevMatch = s.matchStart
		s.matchLength = 2
		if hashHead != 0 && s.prevLength < 16 && ((s.strstart-hashHead)&0xffff) <= zlibWindow-zlibMinLookahead {
			s.matchLength = s.longestMatch(hashHead)
			if s.matchLength == 3 && s.strstart-s.matchStart > 4096 {
				s.matchLength = 2
			}
		}
		if s.prevLength >= 3 && s.matchLength <= s.prevLength {
			maxInsert := s.strstart + s.lookahead - 3
			flush := s.tally(s.strstart-1-s.prevMatch, s.prevLength-3)
			s.lookahead -= s.prevLength - 1
			s.prevLength -= 2
			for {
				s.strstart++
				if s.strstart <= maxInsert {
					s.insert()
				}
				s.prevLength--
				if s.prevLength == 0 {
					break
				}
			}
			s.matchAvailable = 0
			s.matchLength = 2
			s.strstart++
			if flush {
				s.flushBlock(false)
			}
		} else if s.matchAvailable != 0 {
			flush := s.tally(0, int(s.window[s.strstart-1]))
			if flush {
				s.flushBlock(false)
			}
			s.strstart++
			s.lookahead--
		} else {
			s.matchAvailable = 1
			s.strstart++
			s.lookahead--
		}
	}
	if s.matchAvailable != 0 {
		s.tally(0, int(s.window[s.strstart-1]))
		s.matchAvailable = 0
	}
	s.flushBlock(finish)
}
func zlibDistanceCode(dist int) int {
	if dist < 256 {
		return zlibDistCode[dist]
	}
	return zlibDistCode[256+(dist>>7)]
}
func (s *zlibDeflater) tally(dist, lc int) bool {
	s.distances[s.lastLit] = uint16(dist)
	s.literals[s.lastLit] = byte(lc)
	s.lastLit++
	if dist == 0 {
		s.ltree[lc*2]++
	} else {
		s.matches++
		s.ltree[(zlibLengthCode[lc]+257)*2]++
		s.dtree[zlibDistanceCode(dist-1)*2]++
	}
	if s.lastLit&0x1fff == 0 {
		outLength := s.lastLit * 8
		for code := 0; code < 30; code++ {
			outLength += int(s.dtree[code*2]) * (5 + zlibExtraDbits[code])
		}
		outLength >>= 3
		if s.matches < s.lastLit/2 && outLength < (s.strstart-s.blockStart)/2 {
			return true
		}
	}
	return s.lastLit == zlibLiteralBuffer-1
}
func (s *zlibDeflater) putShort(v uint16) { s.pending = append(s.pending, byte(v), byte(v>>8)) }
func (s *zlibDeflater) sendBits(value, length int) {
	if s.bitValid > 16-length {
		s.bitBuffer |= uint16(value << s.bitValid)
		s.putShort(s.bitBuffer)
		s.bitBuffer = uint16(uint32(value) >> uint(16-s.bitValid))
		s.bitValid += length - 16
	} else {
		s.bitBuffer |= uint16(value << s.bitValid)
		s.bitValid += length
	}
}
func (s *zlibDeflater) windup() {
	if s.bitValid > 8 {
		s.putShort(s.bitBuffer)
	} else if s.bitValid > 0 {
		s.pending = append(s.pending, byte(s.bitBuffer))
	}
	s.bitBuffer = 0
	s.bitValid = 0
}
func (s *zlibDeflater) sendCode(code int, tree []int16) {
	s.sendBits(int(uint16(tree[code*2])), int(uint16(tree[code*2+1])))
}
func (s *zlibDeflater) compressBlock(ltree, dtree []int16) {
	for i := 0; i < s.lastLit; i++ {
		dist, lc := int(s.distances[i]), int(s.literals[i])
		if dist == 0 {
			s.sendCode(lc, ltree)
		} else {
			code := zlibLengthCode[lc]
			s.sendCode(code+257, ltree)
			if extra := zlibExtraLbits[code]; extra != 0 {
				s.sendBits(lc-zlibBaseLength[code], extra)
			}
			dist--
			code = zlibDistanceCode(dist)
			s.sendCode(code, dtree)
			if extra := zlibExtraDbits[code]; extra != 0 {
				s.sendBits(dist-zlibBaseDist[code], extra)
			}
		}
	}
	s.sendCode(256, ltree)
}
func (s *zlibDeflater) flushBlock(eof bool) {
	s.ldesc.build(s)
	s.ddesc.build(s)
	s.scanTree(s.ltree[:], s.ldesc.maxCode)
	s.scanTree(s.dtree[:], s.ddesc.maxCode)
	s.bldesc.build(s)
	maxBL := 18
	for maxBL >= 3 && s.bltree[zlibBlOrder[maxBL]*2+1] == 0 {
		maxBL--
	}
	s.optLen += 3*(maxBL+1) + 5 + 5 + 4
	optBytes := (s.optLen + 3 + 7) >> 3
	staticBytes := (s.staticLen + 3 + 7) >> 3
	optBytes = min(optBytes, staticBytes)
	storedLength := s.strstart - s.blockStart
	last := 0
	if eof {
		last = 1
	}
	if storedLength+4 <= optBytes && s.blockStart >= 0 {
		s.sendBits(last, 3)
		s.windup()
		s.putShort(uint16(storedLength))
		s.putShort(^uint16(storedLength))
		s.pending = append(s.pending, s.window[s.blockStart:s.blockStart+storedLength]...)
	} else if staticBytes == optBytes {
		s.sendBits(2+last, 3)
		// Static tables are immutable; the source stores code/length as shorts.
		var lt [576]int16
		var dt [60]int16
		for i, x := range zlibStaticLtree {
			lt[i] = int16(x)
		}
		for i, x := range zlibStaticDtree {
			dt[i] = int16(x)
		}
		s.compressBlock(lt[:], dt[:])
	} else {
		s.sendBits(4+last, 3)
		s.sendBits(s.ldesc.maxCode+1-257, 5)
		s.sendBits(s.ddesc.maxCode, 5)
		s.sendBits(maxBL+1-4, 4)
		for rank := 0; rank <= maxBL; rank++ {
			s.sendBits(int(s.bltree[zlibBlOrder[rank]*2+1]), 3)
		}
		s.sendTree(s.ltree[:], s.ldesc.maxCode)
		s.sendTree(s.dtree[:], s.ddesc.maxCode)
		s.compressBlock(s.ltree[:], s.dtree[:])
	}
	s.initBlock()
	if eof {
		s.windup()
	}
	s.blockStart = s.strstart
	s.emit(s.pending)
	s.pending = s.pending[:0]
}

// scanTree and sendTree retain zlib's run thresholds and guard sentinel.
func (s *zlibDeflater) scanTree(tree []int16, maxCode int) {
	prev, next, count, maxCount, minCount := -1, int(tree[1]), 0, 7, 4
	if next == 0 {
		maxCount = 138
		minCount = 3
	}
	tree[(maxCode+1)*2+1] = -1
	for n := 0; n <= maxCode; n++ {
		cur := next
		next = int(tree[(n+1)*2+1])
		count++
		if count < maxCount && cur == next {
			continue
		}
		if count < minCount {
			s.bltree[cur*2] += int16(count)
		} else if cur != 0 {
			if cur != prev {
				s.bltree[cur*2]++
			}
			s.bltree[32]++
		} else if count <= 10 {
			s.bltree[34]++
		} else {
			s.bltree[36]++
		}
		count = 0
		prev = cur
		if next == 0 {
			maxCount = 138
			minCount = 3
		} else if cur == next {
			maxCount = 6
			minCount = 3
		} else {
			maxCount = 7
			minCount = 4
		}
	}
}
func (s *zlibDeflater) sendTree(tree []int16, maxCode int) {
	prev, next, count, maxCount, minCount := -1, int(tree[1]), 0, 7, 4
	if next == 0 {
		maxCount = 138
		minCount = 3
	}
	for n := 0; n <= maxCode; n++ {
		cur := next
		next = int(tree[(n+1)*2+1])
		count++
		if count < maxCount && cur == next {
			continue
		}
		if count < minCount {
			for ; count > 0; count-- {
				s.sendCode(cur, s.bltree[:])
			}
		} else if cur != 0 {
			if cur != prev {
				s.sendCode(cur, s.bltree[:])
				count--
			}
			s.sendCode(16, s.bltree[:])
			s.sendBits(count-3, 2)
		} else if count <= 10 {
			s.sendCode(17, s.bltree[:])
			s.sendBits(count-3, 3)
		} else {
			s.sendCode(18, s.bltree[:])
			s.sendBits(count-11, 7)
		}
		count = 0
		prev = cur
		if next == 0 {
			maxCount = 138
			minCount = 3
		} else if cur == next {
			maxCount = 6
			minCount = 3
		} else {
			maxCount = 7
			minCount = 4
		}
	}
}
func (s *zlibDeflater) smaller(tree []int16, n, m int) bool {
	return tree[n*2] < tree[m*2] || (tree[n*2] == tree[m*2] && s.depth[n] <= s.depth[m])
}
func (s *zlibDeflater) downHeap(tree []int16, k int) {
	v := s.heap[k]
	for j := k * 2; j <= s.heapLen; j = k * 2 {
		if j < s.heapLen && s.smaller(tree, s.heap[j+1], s.heap[j]) {
			j++
		}
		if s.smaller(tree, v, s.heap[j]) {
			break
		}
		s.heap[k] = s.heap[j]
		k = j
	}
	s.heap[k] = v
}
func (t *zlibTree) build(s *zlibDeflater) {
	tree := t.tree
	maxCode := -1
	s.heapLen = 0
	s.heapMax = zlibHeapSize
	for n := 0; n < t.elems; n++ {
		if tree[n*2] != 0 {
			s.heapLen++
			s.heap[s.heapLen] = n
			maxCode = n
			s.depth[n] = 0
		} else {
			tree[n*2+1] = 0
		}
	}
	for s.heapLen < 2 {
		node := 0
		if maxCode < 2 {
			maxCode++
			node = maxCode
		}
		s.heapLen++
		s.heap[s.heapLen] = node
		tree[node*2] = 1
		s.depth[node] = 0
		s.optLen--
		if t.static != nil {
			s.staticLen -= t.static[node*2+1]
		}
	}
	t.maxCode = maxCode
	for n := s.heapLen / 2; n >= 1; n-- {
		s.downHeap(tree, n)
	}
	node := t.elems
	for {
		n := s.heap[1]
		s.heap[1] = s.heap[s.heapLen]
		s.heapLen--
		s.downHeap(tree, 1)
		m := s.heap[1]
		s.heapMax--
		s.heap[s.heapMax] = n
		s.heapMax--
		s.heap[s.heapMax] = m
		tree[node*2] = tree[n*2] + tree[m*2]
		s.depth[node] = max(s.depth[n], s.depth[m]) + 1
		tree[n*2+1] = int16(node)
		tree[m*2+1] = int16(node)
		s.heap[1] = node
		node++
		s.downHeap(tree, 1)
		if s.heapLen < 2 {
			break
		}
	}
	s.heapMax--
	s.heap[s.heapMax] = s.heap[1]
	t.genBitlen(s)
	var code int16
	s.nextCode[0] = 0
	for bits := 1; bits <= 15; bits++ {
		code = (code + s.blCount[bits-1]) << 1
		s.nextCode[bits] = code
	}
	for n := 0; n <= maxCode; n++ {
		length := int(tree[n*2+1])
		if length == 0 {
			continue
		}
		tree[n*2] = int16(zlibReverse(uint16(s.nextCode[length]), length))
		s.nextCode[length]++
	}
}
func zlibReverse(code uint16, length int) uint16 {
	var res uint16
	for i := 0; i < length; i++ {
		res = (res << 1) | (code & 1)
		code >>= 1
	}
	return res
}
func (t *zlibTree) genBitlen(s *zlibDeflater) {
	tree := t.tree
	for i := range s.blCount {
		s.blCount[i] = 0
	}
	overflow := 0
	tree[s.heap[s.heapMax]*2+1] = 0
	h := s.heapMax + 1
	for ; h < zlibHeapSize; h++ {
		n := s.heap[h]
		bits := int(tree[int(tree[n*2+1])*2+1]) + 1
		if bits > t.maxLength {
			bits = t.maxLength
			overflow++
		}
		tree[n*2+1] = int16(bits)
		if n > t.maxCode {
			continue
		}
		s.blCount[bits]++
		extra := 0
		if n >= t.base {
			extra = t.extra[n-t.base]
		}
		freq := int(tree[n*2])
		s.optLen += freq * (bits + extra)
		if t.static != nil {
			s.staticLen += freq * (t.static[n*2+1] + extra)
		}
	}
	if overflow == 0 {
		return
	}
	for {
		bits := t.maxLength - 1
		for s.blCount[bits] == 0 {
			bits--
		}
		s.blCount[bits]--
		s.blCount[bits+1] += 2
		s.blCount[t.maxLength]--
		overflow -= 2
		if overflow <= 0 {
			break
		}
	}
	for bits := t.maxLength; bits != 0; bits-- {
		n := int(s.blCount[bits])
		for n != 0 {
			h--
			m := s.heap[h]
			if m > t.maxCode {
				continue
			}
			if int(tree[m*2+1]) != bits {
				s.optLen += (bits - int(tree[m*2+1])) * int(tree[m*2])
				tree[m*2+1] = int16(bits)
			}
			n--
		}
	}
}
