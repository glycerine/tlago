// Copyright (c) 2024, Oracle and/or its affiliates.

package tlc

import (
	"fmt"
	"io"
	"math/bits"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

const javaBRAFFuzzBound = BufferedRandomAccessFileBuffSz * 2

// AbstractFileState uses independent java.util.BitSet semantics: exclusive
// range ends, union with prior writes, and undefined holes even on a zero-filling OS.
type javaBRAFFuzzState struct {
	cursor, length int
	written        []uint64
}

func newJavaBRAFFuzzState() *javaBRAFFuzzState {
	return &javaBRAFFuzzState{written: make([]uint64, 1024/64)}
}
func (s *javaBRAFFuzzState) writeBytes(count int) {
	end := s.cursor + count
	if end > s.cursor {
		last := (end - 1) / 64
		if last >= len(s.written) {
			next := make([]uint64, max(2*len(s.written), last+1))
			copy(next, s.written)
			s.written = next
		}
		first := s.cursor / 64
		for word := first; word <= last; word++ {
			mask := ^uint64(0)
			if word == first {
				mask &= ^uint64(0) << uint(s.cursor%64)
			}
			if word == last {
				mask &= ^uint64(0) >> uint(63-(end-1)%64)
			}
			s.written[word] |= mask
		}
	}
	s.cursor = end
	s.length = max(s.length, s.cursor)
}
func (s *javaBRAFFuzzState) readBytes(count int) {
	s.cursor = max(min(s.cursor+count, s.length), s.cursor)
}
func (s *javaBRAFFuzzState) nextClearBit(from int) int {
	word := from / 64
	shift := uint(from % 64)
	if word >= len(s.written) {
		return from
	}
	available := ^s.written[word] & (^uint64(0) << shift)
	for {
		if available != 0 {
			return word*64 + bits.TrailingZeros64(available)
		}
		word++
		if word >= len(s.written) {
			return word * 64
		}
		available = ^s.written[word]
	}
}
func (s *javaBRAFFuzzState) readWouldBeWellDefined(count int) bool {
	return s.nextClearBit(s.cursor) >= min(s.cursor+count, s.length)
}
func (s *javaBRAFFuzzState) setLength(length int) {
	word := length / 64
	shift := uint(length % 64)
	if word < len(s.written) {
		s.written[word] &= (uint64(1) << shift) - 1
		for i := word + 1; i < len(s.written); i++ {
			s.written[i] = 0
		}
	}
	s.cursor = min(s.cursor, length)
	s.length = length
}
func (s *javaBRAFFuzzState) seek(pos int) { s.cursor = pos }
func (s *javaBRAFFuzzState) cardinality() int {
	count := 0
	for _, word := range s.written {
		count += bits.OnesCount64(word)
	}
	return count
}
func (s *javaBRAFFuzzState) String() string {
	return fmt.Sprintf("AbstractFileState[cursor=%d, length=%d, |writtenPositions|=%d]", s.cursor, s.length, s.cardinality())
}

// The oracle is a separate native, unbuffered RandomAccessFile adapter. Java
// setLength constrains its underlying file pointer even when extending a file.
type javaBRAFFuzzNative struct{ file *os.File }

func (f *javaBRAFFuzzNative) close() { javaBRAFThrow(bufferedRandomAccessFileIOError(f.file.Close())) }
func (f *javaBRAFFuzzNative) read() int64 {
	var b [1]byte
	n, e := f.file.Read(b[:])
	if n == 0 && e == io.EOF {
		return -1
	}
	javaBRAFThrow(bufferedRandomAccessFileIOError(e))
	return int64(b[0])
}
func (f *javaBRAFFuzzNative) readArray(b []byte, offset, length int) int64 {
	n, e := f.file.Read(b[offset : offset+length])
	if n == 0 && e == io.EOF {
		return -1
	}
	if e != io.EOF {
		javaBRAFThrow(bufferedRandomAccessFileIOError(e))
	}
	return int64(n)
}
func (f *javaBRAFFuzzNative) write(value int) {
	_, e := f.file.Write([]byte{byte(value)})
	javaBRAFThrow(bufferedRandomAccessFileIOError(e))
}
func (f *javaBRAFFuzzNative) writeArray(b []byte, offset, length int) {
	_, e := f.file.Write(b[offset : offset+length])
	javaBRAFThrow(bufferedRandomAccessFileIOError(e))
}
func (f *javaBRAFFuzzNative) seek(pos int64) {
	_, e := f.file.Seek(pos, io.SeekStart)
	javaBRAFThrow(bufferedRandomAccessFileIOError(e))
}
func (f *javaBRAFFuzzNative) pointer() int64 {
	pos, e := f.file.Seek(0, io.SeekCurrent)
	javaBRAFThrow(bufferedRandomAccessFileIOError(e))
	return pos
}
func (f *javaBRAFFuzzNative) length() int64 {
	info, e := f.file.Stat()
	javaBRAFThrow(bufferedRandomAccessFileIOError(e))
	return info.Size()
}
func (f *javaBRAFFuzzNative) setLength(length int64) {
	javaBRAFThrow(bufferedRandomAccessFileIOError(f.file.Truncate(length)))
	if f.pointer() > length {
		f.seek(length)
	}
}

type javaBRAFFuzzFile interface {
	read() int64
	readArray([]byte, int, int) int64
	write(int)
	writeArray([]byte, int, int)
	seek(int64)
	pointer() int64
	length() int64
	setLength(int64)
}

// Discriminants are the source arbitraryRandomOperation switch's eight cases.
type javaBRAFFuzzOp struct {
	kind           int
	value          int8
	buffer         []byte
	offset, length int
}

func (op javaBRAFFuzzOp) execute(f javaBRAFFuzzFile) any {
	switch op.kind {
	case 0:
		f.write(int(op.value))
		return nil
	case 1:
		f.writeArray(op.buffer, op.offset, op.length)
		return nil
	case 2:
		return f.read()
	case 3:
		var result []int8
		// Preserve null vs empty List, the first zero-length read, the unchanged
		// offset on every partial read, and the source smoothing loop exactly.
		for result == nil || len(result) < op.length {
			n := int(f.readArray(op.buffer, op.offset, op.length-len(result)))
			if n == -1 {
				break
			}
			if result == nil {
				result = make([]int8, 0, op.length)
			}
			for i := 0; i < n; i++ {
				result = append(result, int8(op.buffer[op.offset+i]))
			}
		}
		if result == nil {
			return nil
		}
		return result
	case 4:
		f.seek(int64(op.offset))
		return nil
	case 5:
		return f.pointer()
	case 6:
		return f.length()
	case 7:
		f.setLength(int64(op.length))
		return nil
	default:
		panic(NewIllegalStateException("Unexpected random number"))
	}
}
func (op javaBRAFFuzzOp) simulateIfWellDefined(s *javaBRAFFuzzState) bool {
	switch op.kind {
	case 0:
		s.writeBytes(1)
	case 1:
		s.writeBytes(op.length)
	case 2:
		if !s.readWouldBeWellDefined(1) {
			return false
		}
		s.readBytes(1)
	case 3:
		if !s.readWouldBeWellDefined(op.length) {
			return false
		}
		s.readBytes(op.length)
	case 4:
		s.seek(op.offset)
	case 7:
		s.setLength(op.length)
	}
	return true
}
func (op javaBRAFFuzzOp) String() string {
	switch op.kind {
	case 0:
		return fmt.Sprintf("write(%d)", op.value)
	case 1:
		values := make([]string, len(op.buffer))
		for i, v := range op.buffer {
			values[i] = fmt.Sprint(int8(v))
		}
		return fmt.Sprintf("write([%s], %d, %d)", strings.Join(values, ", "), op.offset, op.length)
	case 2:
		return "read()"
	case 3:
		return fmt.Sprintf("read(buffer, %d, %d)", op.offset, op.length)
	case 4:
		return fmt.Sprintf("seek(%d)", op.offset)
	case 5:
		return "getFilePointer()"
	case 6:
		return "length()"
	case 7:
		return fmt.Sprintf("setLength(%d)", op.length)
	}
	panic(NewIllegalStateException("Unexpected random number"))
}
func javaBRAFFuzzWellDefined(ops []javaBRAFFuzzOp) bool {
	s := newJavaBRAFFuzzState()
	for _, op := range ops {
		if !op.simulateIfWellDefined(s) {
			return false
		}
	}
	return true
}
func javaBRAFFuzzRandomBytes(r *JavaRandom) []byte {
	b := make([]byte, int(r.NextIntN(javaBRAFFuzzBound))+1)
	r.NextBytes(b)
	return b
}
func javaBRAFFuzzArbitrary(r *JavaRandom) javaBRAFFuzzOp {
	op := javaBRAFFuzzOp{kind: int(r.NextIntN(8))}
	switch op.kind {
	case 0:
		op.value = int8(r.NextIntN(0xff))
	case 1, 3:
		op.buffer = javaBRAFFuzzRandomBytes(r)
		op.offset = int(r.NextIntN(int32(len(op.buffer))))
		op.length = int(r.NextIntN(int32(len(op.buffer) - op.offset)))
	case 4:
		op.offset = int(r.NextIntN(javaBRAFFuzzBound))
	case 7:
		op.length = int(r.NextIntN(javaBRAFFuzzBound))
	}
	return op
}
func javaBRAFFuzzGenerate(r *JavaRandom, n int) []javaBRAFFuzzOp {
	ops := make([]javaBRAFFuzzOp, 0, n)
	state := newJavaBRAFFuzzState()
	for i := 0; i < n; i++ {
		for {
			op := javaBRAFFuzzArbitrary(r)
			if op.simulateIfWellDefined(state) {
				ops = append(ops, op)
				break
			}
		}
	}
	if !javaBRAFFuzzWellDefined(ops) {
		panic(NewIllegalStateException(fmt.Sprint("generated not-well-defined ops: ", ops)))
	}
	return ops
}

// Render the source Objects/List/Class diagnostic forms without asserting
// nondeterministic native identities or weakening any comparison.
func javaBRAFFuzzObjectString(value any) string {
	if value == nil {
		return "null"
	}
	if list, ok := value.([]int8); ok {
		parts := make([]string, len(list))
		for i, v := range list {
			parts[i] = fmt.Sprint(v)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	}
	if kind, ok := value.(reflect.Type); ok {
		return "class " + kind.String()
	}
	return fmt.Sprint(value)
}

type javaBRAFFuzzResult struct {
	ops              []javaBRAFFuzzOp
	successes        int
	expected, actual any
}

func (r *javaBRAFFuzzResult) ok() bool { return reflect.DeepEqual(r.expected, r.actual) }
func (r *javaBRAFFuzzResult) equal(other *javaBRAFFuzzResult) bool {
	return other != nil && r.successes == other.successes && reflect.DeepEqual(r.expected, other.expected) && reflect.DeepEqual(r.actual, other.actual)
}
func (r *javaBRAFFuzzResult) String() string {
	return fmt.Sprintf("RunResult{opsExecuted=%d, expected=%v, actual=%v}", r.successes, javaBRAFFuzzObjectString(r.expected), javaBRAFFuzzObjectString(r.actual))
}
func javaBRAFFuzzCheck(dir string, ops []javaBRAFFuzzOp) *javaBRAFFuzzResult {
	temp1, e := os.CreateTemp(dir, "tmp*.bin")
	javaBRAFThrow(e)
	name1 := temp1.Name()
	javaBRAFThrow(temp1.Close())
	temp2, e := os.CreateTemp(dir, "tmp*.bin")
	javaBRAFThrow(e)
	name2 := temp2.Name()
	javaBRAFThrow(temp2.Close())
	native, e := os.OpenFile(name1, os.O_RDWR, 0600)
	javaBRAFThrow(e)
	f1 := &javaBRAFFuzzNative{native}
	var f2 *javaBRAF
	// Java try-with-resources closes in reverse order and preserves the first
	// thrown failure, including failures while initializing the second resource.
	defer func() {
		primary := recover()
		closeResource := func(close func()) {
			defer func() {
				if secondary := recover(); secondary != nil && primary == nil {
					primary = secondary
				}
			}()
			close()
		}
		if f2 != nil {
			closeResource(f2.close)
		}
		closeResource(f1.close)
		if primary != nil {
			panic(primary)
		}
	}()
	f2 = javaBRAFOpen(name2, "rw")
	for i, op := range ops {
		expected := op.execute(f1)
		var actual any
		caught := false
		func() {
			defer func() {
				if failure := recover(); failure != nil {
					if e, ok := failure.(error); ok && isJavaError(e) {
						panic(failure)
					}
					caught = true
					actual = reflect.TypeOf(failure)
				}
			}()
			actual = op.execute(f2)
		}()
		if caught {
			return &javaBRAFFuzzResult{ops, i, nil, actual}
		}
		if !reflect.DeepEqual(expected, actual) {
			return &javaBRAFFuzzResult{ops, i, expected, actual}
		}
	}
	return &javaBRAFFuzzResult{ops, len(ops), nil, nil}
}
func javaBRAFFuzzAppend(a, b []javaBRAFFuzzOp) []javaBRAFFuzzOp {
	if len(a) == 0 {
		return b
	}
	if len(b) == 0 {
		return a
	}
	result := make([]javaBRAFFuzzOp, 0, len(a)+len(b))
	result = append(result, a...)
	return append(result, b...)
}
func javaBRAFFuzzMinimize(list []javaBRAFFuzzOp, test func([]javaBRAFFuzzOp) bool) []javaBRAFFuzzOp {
	stride := len(list) / 2
	for stride > 0 {
		modified := false
		for i := 0; i < len(list); i += stride {
			end := min(i+stride, len(list))
			candidate := javaBRAFFuzzAppend(list[:i], list[end:])
			if test(candidate) {
				list = candidate
				i -= stride
				modified = true
			}
		}
		if !modified {
			stride /= 2
		}
	}
	return list
}

func TestJavaBufferedRandomAccessFileFuzz(t *testing.T) {
	t.Run("fuzz", func(t *testing.T) {
		dir := t.TempDir() // all source deleteOnExit files retained until this cleanup
		var next atomic.Int32
		next.Store(1)
		var running atomic.Bool
		running.Store(true)
		var bad atomic.Pointer[javaBRAFFuzzResult]
		type thrown struct{ value any }
		var failure atomic.Pointer[thrown]
		var output sync.Mutex
		var workers sync.WaitGroup
		for threadID := 0; threadID < runtime.NumCPU(); threadID++ {
			workers.Add(1)
			go func() {
				defer workers.Done()
				rng := NewJavaRandom(int64(threadID))
				for running.Load() {
					runID := next.Add(1) - 1
					if runID > 10000 {
						return
					}
					output.Lock()
					fmt.Println("---- starting run", runID)
					output.Unlock()
					escaped := false
					func() {
						defer func() {
							if e := recover(); e != nil {
								failure.Store(&thrown{e})
								running.Store(false)
								escaped = true
							}
						}()
						result := javaBRAFFuzzCheck(dir, javaBRAFFuzzGenerate(rng, 50))
						if !result.ok() {
							bad.Store(result)
							running.Store(false)
						}
					}()
					if escaped {
						return
					}
				}
			}()
		}
		workers.Wait()
		if result := bad.Load(); result != nil && !result.ok() {
			fmt.Printf("Violation found on op # %d; minimizing...\n", result.successes)
			minimal := javaBRAFFuzzMinimize(result.ops[:result.successes+1], func(subset []javaBRAFFuzzOp) bool {
				return javaBRAFFuzzWellDefined(subset) && !javaBRAFFuzzCheck(dir, subset).ok()
			})
			fmt.Println("Minimal trace:")
			for _, op := range minimal {
				fmt.Println(" -->", op)
			}
			minResult := javaBRAFFuzzCheck(dir, minimal)
			panic(NewRuntimeException("got " + javaBRAFFuzzObjectString(minResult.actual) + " but expected " + javaBRAFFuzzObjectString(minResult.expected)))
		}
		if e := failure.Load(); e != nil {
			cause, ok := e.value.(error)
			if !ok {
				cause = fmt.Errorf("%v", e.value)
			}
			panic(NewRuntimeExceptionFromCause(cause))
		}
	})
	t.Run("testWellDefined", func(t *testing.T) {
		ops := []javaBRAFFuzzOp{{kind: 2}, {kind: 4, offset: 12637}, {kind: 2}, {kind: 0, value: -39}, {kind: 4, offset: 4953}, {kind: 3, buffer: make([]byte, javaBRAFFuzzBound), offset: 2121, length: 4449}}
		if javaBRAFFuzzWellDefined(ops) {
			t.Fatal("assertFalse(wellDefined(ops))")
		}
	})
}
