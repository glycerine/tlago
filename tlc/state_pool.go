package tlc

import (
	"fmt"
	"os"
	"sync"
)

type StatePoolReader struct {
	mu       sync.Mutex
	cond     *sync.Cond
	buf      []*TLCStateMut
	poolFile string
	isFull   bool
	canRead  bool
	finished bool
	err      error
	done     chan struct{}
}

func NewStatePoolReader(bufSize int, file string) *StatePoolReader {
	r := &StatePoolReader{
		buf:      make([]*TLCStateMut, bufSize),
		poolFile: file,
		done:     make(chan struct{}),
	}
	r.cond = sync.NewCond(&r.mu)
	return r
}

func (r *StatePoolReader) Start() {
	go r.run()
}

func (r *StatePoolReader) Wakeup() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.canRead = true
	r.cond.Signal()
}

func (r *StatePoolReader) Restart(file string, canRead bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.poolFile = file
	r.isFull = false
	r.canRead = canRead
	r.cond.Signal()
}

func (r *StatePoolReader) DoWork(deqBuf []*TLCStateMut, file string) ([]*TLCStateMut, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return nil, r.err
	}
	if r.isFull {
		out := r.buf
		r.buf = deqBuf
		r.poolFile = file
		r.isFull = false
		r.canRead = true
		r.cond.Signal()
		return out, nil
	}
	if r.poolFile != "" {
		if err := readStatePoolFile(r.poolFile, deqBuf); err != nil {
			return nil, err
		}
		r.poolFile = file
		r.canRead = true
		r.cond.Signal()
		return deqBuf, nil
	}
	if err := readStatePoolFile(file, deqBuf); err != nil {
		return nil, err
	}
	return deqBuf, nil
}

func (r *StatePoolReader) GetCache(deqBuf []*TLCStateMut, file string) ([]*TLCStateMut, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.err != nil {
		return nil, r.err
	}
	if r.isFull {
		out := r.buf
		r.buf = deqBuf
		r.poolFile = file
		r.isFull = false
		r.canRead = false
		return out, nil
	}
	if r.poolFile != "" && r.canRead {
		if err := readStatePoolFile(r.poolFile, deqBuf); err != nil {
			return nil, err
		}
		r.poolFile = file
		r.canRead = false
		return deqBuf, nil
	}
	return nil, nil
}

func (r *StatePoolReader) SetFinished() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finished = true
	r.cond.Broadcast()
}

func (r *StatePoolReader) run() {
	defer close(r.done)
	r.mu.Lock()
	defer r.mu.Unlock()
	for {
		for r.poolFile == "" || r.isFull || !r.canRead {
			if r.finished {
				return
			}
			r.cond.Wait()
		}
		if err := readStatePoolFile(r.poolFile, r.buf); err != nil {
			r.err = err
			r.cond.Broadcast()
			return
		}
		r.poolFile = ""
		r.isFull = true
	}
}

type StatePoolWriter struct {
	mu       sync.Mutex
	cond     *sync.Cond
	buf      []*TLCStateMut
	poolFile string
	reader   *StatePoolReader
	finished bool
	err      error
	done     chan struct{}
}

func NewStatePoolWriter(bufSize int, reader *StatePoolReader) *StatePoolWriter {
	w := &StatePoolWriter{
		buf:    make([]*TLCStateMut, bufSize),
		reader: reader,
		done:   make(chan struct{}),
	}
	w.cond = sync.NewCond(&w.mu)
	return w
}

func (w *StatePoolWriter) Start() {
	go w.run()
}

func (w *StatePoolWriter) DoWork(enqBuf []*TLCStateMut, file string) ([]*TLCStateMut, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.err != nil {
		return nil, w.err
	}
	if w.poolFile != "" {
		if err := writeStatePoolFile(w.poolFile, w.buf); err != nil {
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
	w.mu.Lock()
	defer w.mu.Unlock()
	for w.poolFile != "" {
		w.cond.Wait()
	}
	return w.err
}

func (w *StatePoolWriter) SetFinished() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.finished = true
	w.cond.Broadcast()
}

func (w *StatePoolWriter) run() {
	defer close(w.done)
	w.mu.Lock()
	defer w.mu.Unlock()
	for {
		for w.poolFile == "" {
			if w.finished {
				return
			}
			w.cond.Wait()
		}
		if err := writeStatePoolFile(w.poolFile, w.buf); err != nil {
			w.err = err
			w.cond.Broadcast()
			return
		}
		w.poolFile = ""
		w.cond.Broadcast()
		if w.reader != nil {
			w.reader.Wakeup()
		}
	}
}

func readStatePoolFile(name string, states []*TLCStateMut) error {
	file, err := os.Open(name)
	if err != nil {
		return err
	}
	in, err := NewValueInputStreamWithGlobalCompression(file)
	if err != nil {
		_ = file.Close()
		return err
	}
	defer in.Close()
	for i := range states {
		state := NewEmptyState()
		if err := state.Read(in); err != nil {
			return err
		}
		states[i] = state
	}
	return nil
}

func writeStatePoolFile(name string, states []*TLCStateMut) error {
	file, err := os.Create(name)
	if err != nil {
		return err
	}
	out := NewValueOutputStreamWithGlobalCompression(file)
	for i, state := range states {
		if state == nil {
			_ = out.Close()
			return fmt.Errorf("state pool write encountered nil state at slot %d", i)
		}
		if err := state.Write(out); err != nil {
			_ = out.Close()
			return err
		}
	}
	return out.Close()
}
