package tlc

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

type statePoolInputStream interface {
	stateInputStream
	Close() error
}

type statePoolOutputStream interface {
	stateOutputStream
	Close() error
}

// Capture the stream implementation before starting pool workers. Java's
// verification harness supplies its original modeled IValue streams; normal
// constructors retain the file-backed implementation.
type statePoolStreamFactory interface {
	openInput(string) (statePoolInputStream, error)
	openOutput(string) (statePoolOutputStream, error)
}

type fileStatePoolStreams struct{}

func (fileStatePoolStreams) openInput(name string) (statePoolInputStream, error) {
	file, err := openPoolFile(name, false)
	if err != nil {
		return nil, err
	}
	return NewValueInputStreamWithGlobalCompression(file)
}

func (fileStatePoolStreams) openOutput(name string) (statePoolOutputStream, error) {
	file, err := openPoolFile(name, true)
	if err != nil {
		return nil, err
	}
	return newValueOutputStreamWithCompression(file, UseGZIP())
}

func poolStreams(factory statePoolStreamFactory) statePoolStreamFactory {
	if factory == nil {
		return fileStatePoolStreams{}
	}
	return factory
}

type StatePoolReader struct {
	streams  statePoolStreamFactory
	mu       sync.Mutex
	cond     *sync.Cond
	buf      []*TLCStateMut
	poolFile string
	isFull   bool
	canRead  bool
	finished bool
	started  atomic.Bool
	done     chan struct{}
}

func NewStatePoolReader(bufSize int, file string) *StatePoolReader {
	return newStatePoolReaderWithStreams(bufSize, file, nil)
}

func newStatePoolReaderWithStreams(bufSize int, file string, streams statePoolStreamFactory) *StatePoolReader {
	bufSize = int(int32(bufSize))
	if bufSize < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(bufSize)))
	}
	r := &StatePoolReader{
		streams:  streams,
		buf:      make([]*TLCStateMut, bufSize),
		poolFile: file,
		done:     make(chan struct{}),
	}
	r.cond = sync.NewCond(&r.mu)
	return r
}

func (r *StatePoolReader) requireStatePoolReader() {
	if r == nil {
		panic(NewNullPointerException())
	}
}

func (r *StatePoolReader) Start() {
	r.requireStatePoolReader()
	if !r.started.CompareAndSwap(false, true) {
		panic(NewIllegalThreadStateException())
	}
	go r.run()
}

func (r *StatePoolReader) Wakeup() {
	r.requireStatePoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.canRead = true
	r.cond.Signal()
}

func (r *StatePoolReader) Restart(file string, canRead bool) {
	r.requireStatePoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.poolFile = file
	r.isFull = false
	r.canRead = canRead
	r.cond.Signal()
}

func (r *StatePoolReader) DoWork(deqBuf []*TLCStateMut, file string) ([]*TLCStateMut, error) {
	r.requireStatePoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isFull {
		if r.poolFile != "" {
			panic(NewAssertionErrorInt(int32(ECSystemFileNull)))
		}
		out := r.buf
		r.buf = deqBuf
		r.poolFile = file
		r.isFull = false
		r.canRead = true
		r.cond.Signal()
		return out, nil
	}
	if r.poolFile != "" {
		if err := readStatePoolFileWithStreams(r.poolFile, deqBuf, r.streams); err != nil {
			return nil, err
		}
		r.poolFile = file
		r.canRead = true
		r.cond.Signal()
		return deqBuf, nil
	}
	if err := readStatePoolFileWithStreams(file, deqBuf, r.streams); err != nil {
		return nil, err
	}
	return deqBuf, nil
}

func (r *StatePoolReader) GetCache(deqBuf []*TLCStateMut, file string) ([]*TLCStateMut, error) {
	r.requireStatePoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.isFull {
		if r.poolFile != "" {
			panic(NewAssertionErrorInt(int32(ECSystemFileNull)))
		}
		out := r.buf
		r.buf = deqBuf
		r.poolFile = file
		r.isFull = false
		r.canRead = false
		return out, nil
	}
	if r.poolFile != "" && r.canRead {
		if err := readStatePoolFileWithStreams(r.poolFile, deqBuf, r.streams); err != nil {
			return nil, err
		}
		r.poolFile = file
		r.canRead = false
		return deqBuf, nil
	}
	return nil, nil
}

func (r *StatePoolReader) SetFinished() {
	r.requireStatePoolReader()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finished = true
	r.cond.Broadcast()
}

func (r *StatePoolReader) run() {
	defer close(r.done)
	defer func() {
		if failure := recover(); failure != nil {
			r.mu.Lock()
			name := r.poolFile
			r.mu.Unlock()
			statePoolFailureExit(ECSystemErrorReadingPool, name, failure)
		}
	}()
	r.mu.Lock()
	defer r.mu.Unlock()
	for {
		for r.poolFile == "" || r.isFull || !r.canRead {
			if r.finished {
				return
			}
			r.cond.Wait()
			if r.finished {
				return
			}
		}
		if err := readStatePoolFileWithStreams(r.poolFile, r.buf, r.streams); err != nil {
			panic(err)
		}
		r.poolFile = ""
		r.isFull = true
	}
}

type StatePoolWriter struct {
	streams  statePoolStreamFactory
	mu       sync.Mutex
	cond     *sync.Cond
	buf      []*TLCStateMut
	poolFile string
	reader   *StatePoolReader
	finished bool
	started  atomic.Bool
	done     chan struct{}
}

func NewStatePoolWriter(bufSize int, reader *StatePoolReader) *StatePoolWriter {
	return newStatePoolWriterWithStreams(bufSize, reader, nil)
}

func newStatePoolWriterWithStreams(bufSize int, reader *StatePoolReader, streams statePoolStreamFactory) *StatePoolWriter {
	bufSize = int(int32(bufSize))
	if bufSize < 0 {
		panic(NewNegativeArraySizeException(fmt.Sprint(bufSize)))
	}
	w := &StatePoolWriter{
		streams: streams,
		buf:     make([]*TLCStateMut, bufSize),
		reader:  reader,
		done:    make(chan struct{}),
	}
	w.cond = sync.NewCond(&w.mu)
	return w
}

func (w *StatePoolWriter) requireStatePoolWriter() {
	if w == nil {
		panic(NewNullPointerException())
	}
}

func (w *StatePoolWriter) Start() {
	w.requireStatePoolWriter()
	if !w.started.CompareAndSwap(false, true) {
		panic(NewIllegalThreadStateException())
	}
	go w.run()
}

func (w *StatePoolWriter) DoWork(enqBuf []*TLCStateMut, file string) ([]*TLCStateMut, error) {
	w.requireStatePoolWriter()
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.poolFile != "" {
		if err := writeStatePoolFileWithStreams(w.poolFile, w.buf, w.streams); err != nil {
			return nil, err
		}
	}
	out := w.buf
	w.buf = enqBuf
	w.poolFile = file
	w.cond.Signal()
	return out, nil
}

func (w *StatePoolWriter) EnsureWritten() error {
	w.requireStatePoolWriter()
	w.mu.Lock()
	defer w.mu.Unlock()
	for w.poolFile != "" {
		w.cond.Wait()
	}
	return nil
}

func (w *StatePoolWriter) SetFinished() {
	w.requireStatePoolWriter()
	w.mu.Lock()
	defer w.mu.Unlock()
	w.finished = true
	w.cond.Broadcast()
}

func (w *StatePoolWriter) run() {
	defer close(w.done)
	defer func() {
		if failure := recover(); failure != nil {
			statePoolFailureExit(ECSystemErrorWritingPool, "", failure)
		}
	}()
	w.mu.Lock()
	defer w.mu.Unlock()
	for {
		for w.poolFile == "" {
			if w.finished {
				return
			}
			w.cond.Wait()
		}
		if err := writeStatePoolFileWithStreams(w.poolFile, w.buf, w.streams); err != nil {
			panic(err)
		}
		w.poolFile = ""
		w.cond.Broadcast()
		if w.reader != nil {
			w.reader.Wakeup()
		}
	}
}

// Source run catches ordinary failures, reports their cause, and exits the
// process. A stored error or a stopped goroutine would leave live queue waiters.
func statePoolFailureExit(code int, name string, failure any) {
	err := panicValueAsError(failure)
	if isJavaError(err) {
		panic(failure)
	}
	params := []*string{javaThrowableDetailMessage(err)}
	if code == ECSystemErrorReadingPool && name != "" {
		basename := filepath.Base(name)
		params = append(params, &basename)
	}
	PrintErrorParametersWithThrowable(code, params, err)
	os.Exit(1)
}

func readStatePoolFile(name string, states []*TLCStateMut) error {
	return readStatePoolFileWithStreams(name, states, nil)
}

func readStatePoolFileWithStreams(name string, states []*TLCStateMut, streams statePoolStreamFactory) error {
	in, err := poolStreams(streams).openInput(name)
	if err != nil {
		return err
	}
	if states == nil {
		panic(NewNullPointerException())
	}
	for i := range states {
		state := newStateFromEmptyFactory()
		states[i] = state
		if err := state.readStateStream(in); err != nil {
			if isJavaIOException(err) {
				return err
			}
			panic(err)
		}
	}
	return in.Close()
}

func writeStatePoolFile(name string, states []*TLCStateMut) error {
	return writeStatePoolFileWithStreams(name, states, nil)
}

func writeStatePoolFileWithStreams(name string, states []*TLCStateMut, streams statePoolStreamFactory) error {
	out, err := poolStreams(streams).openOutput(name)
	if err != nil {
		return err
	}
	if states == nil {
		panic(NewNullPointerException())
	}
	for _, state := range states {
		if state == nil {
			panic(NewNullPointerException())
		}
		if err := state.writeStateStream(out); err != nil {
			if isJavaIOException(err) {
				return err
			}
			panic(err)
		}
	}
	return out.Close()
}

func openPoolFile(name string, write bool) (*os.File, error) {
	if write {
		return openFileOutputStream(name)
	}
	return openFileInputStream(name)
}
