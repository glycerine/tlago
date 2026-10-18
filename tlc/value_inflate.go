package tlc

import "io"

type valueInflateNode struct {
	child  [2]int
	symbol int
}
type valueInflateTree struct {
	nodes [576]valueInflateNode
	count int
}

func (t *valueInflateTree) build(lengths []int, codes bool) bool {
	var counts [16]int
	maxBits := 0
	for _, n := range lengths {
		if n < 0 || n > 15 {
			return false
		}
		if n > 0 {
			counts[n]++
			if n > maxBits {
				maxBits = n
			}
		}
	}
	left := 1
	for n := 1; n <= 15; n++ {
		left = left*2 - counts[n]
		if left < 0 {
			return false
		}
	}
	if maxBits != 0 && left > 0 && (codes || maxBits != 1) {
		return false
	}
	t.count = 1
	t.nodes[0] = valueInflateNode{child: [2]int{-1, -1}, symbol: -1}
	var next [16]int
	code := 0
	for n := 1; n <= 15; n++ {
		code = (code + counts[n-1]) << 1
		next[n] = code
	}
	for symbol, n := range lengths {
		if n == 0 {
			continue
		}
		code = next[n]
		next[n]++
		node := 0
		for bit := n - 1; bit >= 0; bit-- {
			direction := (code >> bit) & 1
			child := t.nodes[node].child[direction]
			if child < 0 {
				child = t.count
				t.count++
				t.nodes[child] = valueInflateNode{child: [2]int{-1, -1}, symbol: -1}
				t.nodes[node].child[direction] = child
			}
			node = child
		}
		t.nodes[node].symbol = symbol
	}
	return true
}

const (
	valueInflateHeader = iota
	valueInflateStoredLength
	valueInflateStoredCopy
	valueInflateDynamicHeader
	valueInflateCodeLengths
	valueInflateLengths
	valueInflateRepeat
	valueInflateLiteral
	valueInflateLengthExtra
	valueInflateDistance
	valueInflateDistanceExtra
	valueInflateCopy
	valueInflateDone
)

var valueInflateCodeOrder = [19]int{16, 17, 18, 0, 8, 7, 9, 6, 10, 5, 11, 4, 12, 3, 13, 2, 14, 1, 15}
var valueInflateLengthBase = [29]int{3, 4, 5, 6, 7, 8, 9, 10, 11, 13, 15, 17, 19, 23, 27, 31, 35, 43, 51, 59, 67, 83, 99, 115, 131, 163, 195, 227, 258}
var valueInflateLengthExtraBits = [29]uint{0, 0, 0, 0, 0, 0, 0, 0, 1, 1, 1, 1, 2, 2, 2, 2, 3, 3, 3, 3, 4, 4, 4, 4, 5, 5, 5, 5, 0}
var valueInflateDistanceBase = [30]int{1, 2, 3, 4, 5, 7, 9, 13, 17, 25, 33, 49, 65, 97, 129, 193, 257, 385, 513, 769, 1025, 1537, 2049, 3073, 4097, 6145, 8193, 12289, 16385, 24577}
var valueInflateDistanceExtraBits = [30]uint{0, 0, 0, 0, 1, 1, 2, 2, 3, 3, 4, 4, 5, 5, 6, 6, 7, 7, 8, 8, 9, 9, 10, 10, 11, 11, 12, 12, 13, 13}

type valueInflater struct {
	source                    *valueGzipSource
	bits                      uint64
	nbits                     uint
	state                     int
	final                     bool
	err                       error
	literal, distance, codes  valueInflateTree
	huffNode                  int
	codeLengths               [19]int
	lengths                   [320]int
	nlit, ndist, ncode, index int
	repeat, repeatBase        int
	extra                     uint
	length, dist              int
	window                    [32768]byte
	position, history         int
}

func newValueInflater(source *valueGzipSource) *valueInflater { return &valueInflater{source: source} }
func (f *valueInflater) Reset(source *valueGzipSource) error {
	*f = valueInflater{source: source}
	return nil
}
func (f *valueInflater) Close() error { return nil }

// take never refills the source: exhausting a 512-byte chunk must return any
// output already produced before the caller supplies more compressed input.
func (f *valueInflater) take(n uint) (int, bool) {
	for f.nbits < n {
		if f.source.curr == f.source.length {
			return 0, false
		}
		f.bits |= uint64(f.source.buff[f.source.curr]) << f.nbits
		f.source.curr++
		f.nbits += 8
	}
	v := int(f.bits & ((uint64(1) << n) - 1))
	f.bits >>= n
	f.nbits -= n
	return v, true
}
func (f *valueInflater) symbol(t *valueInflateTree, message string) (int, bool, error) {
	for {
		bit, ok := f.take(1)
		if !ok {
			return 0, false, nil
		}
		node := t.nodes[f.huffNode].child[bit]
		if node < 0 {
			return 0, false, NewZipException(message)
		}
		f.huffNode = node
		if s := t.nodes[node].symbol; s >= 0 {
			f.huffNode = 0
			return s, true, nil
		}
	}
}
func (f *valueInflater) fixed() {
	var lengths [288]int
	for i := range lengths {
		switch {
		case i < 144:
			lengths[i] = 8
		case i < 256:
			lengths[i] = 9
		case i < 280:
			lengths[i] = 7
		default:
			lengths[i] = 8
		}
	}
	f.literal.build(lengths[:], false)
	var distances [32]int
	for i := range distances {
		distances[i] = 5
	}
	f.distance.build(distances[:], false)
}
func (f *valueInflater) endBlock() {
	if f.final {
		f.state = valueInflateDone
		f.bits, f.nbits = 0, 0
	} else {
		f.state = valueInflateHeader
	}
}
func (f *valueInflater) put(p []byte, n *int, b byte) {
	p[*n] = b
	*n += 1
	f.window[f.position] = b
	f.position = (f.position + 1) & 32767
	if f.history < 32768 {
		f.history++
	}
}

func (f *valueInflater) inflate(p []byte) (n int, err error) {
	for n < len(p) {
		switch f.state {
		case valueInflateHeader:
			v, ok := f.take(3)
			if !ok {
				return
			}
			f.final = v&1 != 0
			switch v >> 1 {
			case 0:
				f.bits >>= f.nbits % 8
				f.nbits -= f.nbits % 8
				f.state = valueInflateStoredLength
			case 1:
				f.fixed()
				f.state = valueInflateLiteral
			case 2:
				f.state = valueInflateDynamicHeader
			default:
				return n, NewZipException("invalid block type")
			}
		case valueInflateStoredLength:
			v, ok := f.take(32)
			if !ok {
				return
			}
			if (v&65535)^(v>>16) != 65535 {
				return n, NewZipException("invalid stored block lengths")
			}
			f.length = v & 65535
			f.state = valueInflateStoredCopy
		case valueInflateStoredCopy:
			if f.length == 0 {
				f.endBlock()
				continue
			}
			v, ok := f.take(8)
			if !ok {
				return
			}
			f.put(p, &n, byte(v))
			f.length--
		case valueInflateDynamicHeader:
			v, ok := f.take(14)
			if !ok {
				return
			}
			f.nlit = (v & 31) + 257
			f.ndist = ((v >> 5) & 31) + 1
			f.ncode = (v >> 10) + 4
			if f.nlit > 286 || f.ndist > 30 {
				return n, NewZipException("too many length or distance symbols")
			}
			clear(f.codeLengths[:])
			clear(f.lengths[:])
			f.index = 0
			f.state = valueInflateCodeLengths
		case valueInflateCodeLengths:
			if f.index == f.ncode {
				if !f.codes.build(f.codeLengths[:], true) {
					return n, NewZipException("invalid code lengths set")
				}
				f.index = 0
				f.state = valueInflateLengths
				continue
			}
			v, ok := f.take(3)
			if !ok {
				return
			}
			f.codeLengths[valueInflateCodeOrder[f.index]] = v
			f.index++
		case valueInflateLengths:
			if f.index == f.nlit+f.ndist {
				if f.lengths[256] == 0 {
					return n, NewZipException("invalid code -- missing end-of-block")
				}
				if !f.literal.build(f.lengths[:f.nlit], false) {
					return n, NewZipException("invalid literal/lengths set")
				}
				if !f.distance.build(f.lengths[f.nlit:f.nlit+f.ndist], false) {
					return n, NewZipException("invalid distances set")
				}
				f.state = valueInflateLiteral
				continue
			}
			v, ok, e := f.symbol(&f.codes, "invalid code lengths set")
			if e != nil {
				return n, e
			}
			if !ok {
				return
			}
			if v < 16 {
				f.lengths[f.index] = v
				f.index++
				continue
			}
			f.repeat = 0
			switch v {
			case 16:
				if f.index == 0 {
					return n, NewZipException("invalid bit length repeat")
				}
				f.repeat = f.lengths[f.index-1]
				f.repeatBase = 3
				f.extra = 2
			case 17:
				f.repeatBase = 3
				f.extra = 3
			case 18:
				f.repeatBase = 11
				f.extra = 7
			}
			f.state = valueInflateRepeat
		case valueInflateRepeat:
			v, ok := f.take(f.extra)
			if !ok {
				return
			}
			count := f.repeatBase + v
			if f.index+count > f.nlit+f.ndist {
				return n, NewZipException("invalid bit length repeat")
			}
			for i := 0; i < count; i++ {
				f.lengths[f.index] = f.repeat
				f.index++
			}
			f.state = valueInflateLengths
		case valueInflateLiteral:
			v, ok, e := f.symbol(&f.literal, "invalid literal/length code")
			if e != nil {
				return n, e
			}
			if !ok {
				return
			}
			if v < 256 {
				f.put(p, &n, byte(v))
				continue
			}
			if v == 256 {
				f.endBlock()
				continue
			}
			if v > 285 {
				return n, NewZipException("invalid literal/length code")
			}
			f.length = valueInflateLengthBase[v-257]
			f.extra = valueInflateLengthExtraBits[v-257]
			f.state = valueInflateLengthExtra
		case valueInflateLengthExtra:
			v, ok := f.take(f.extra)
			if !ok {
				return
			}
			f.length += v
			f.state = valueInflateDistance
		case valueInflateDistance:
			v, ok, e := f.symbol(&f.distance, "invalid distance code")
			if e != nil {
				return n, e
			}
			if !ok {
				return
			}
			if v > 29 {
				return n, NewZipException("invalid distance code")
			}
			f.dist = valueInflateDistanceBase[v]
			f.extra = valueInflateDistanceExtraBits[v]
			f.state = valueInflateDistanceExtra
		case valueInflateDistanceExtra:
			v, ok := f.take(f.extra)
			if !ok {
				return
			}
			f.dist += v
			if f.dist > f.history {
				return n, NewZipException("invalid distance too far back")
			}
			f.state = valueInflateCopy
		case valueInflateCopy:
			if f.length == 0 {
				f.state = valueInflateLiteral
				continue
			}
			f.put(p, &n, f.window[(f.position-f.dist)&32767])
			f.length--
		case valueInflateDone:
			return
		}
	}
	return
}
func (f *valueInflater) Read(p []byte) (int, error) {
	if f.err != nil {
		return 0, f.err
	}
	if len(p) == 0 {
		return 0, nil
	}
	for {
		n, err := f.inflate(p)
		if err != nil {
			f.err = err
			return 0, err
		}
		if f.state == valueInflateDone {
			if n > 0 {
				return n, io.EOF
			}
			return 0, io.EOF
		}
		if n > 0 {
			return n, nil
		}
		if err := f.source.fill(); err != nil {
			if bufferedInputEOF(err) {
				err = io.ErrUnexpectedEOF
			}
			f.err = err
			return 0, err
		}
	}
}
