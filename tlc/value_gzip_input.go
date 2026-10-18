package tlc

import (
	"encoding/binary"
	"errors"
	"hash/crc32"
	"io"
	"os"
)

// The framing follows GZIPInputStream. Incremental decompression retains zlib
// input/output boundaries.
// Keeping framing separate preserves checked header/trailer failure boundaries
// and does not impose gzip.Reader's fixed limit on optional header strings.
type valueGzipSource struct {
	io.Reader
	buff         [512]byte
	curr, length int
	pending      error
	buffered     bool
}

func (s *valueGzipSource) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if s.curr < s.length {
		n := copy(p, s.buff[s.curr:s.length])
		s.curr += n
		return n, nil
	}
	if s.pending != nil {
		err := s.pending
		s.pending = nil
		return 0, err
	}
	if !s.buffered {
		return s.Reader.Read(p)
	}
	if err := s.fill(); err != nil {
		return 0, err
	}
	n := copy(p, s.buff[:s.length])
	s.curr = n
	return n, nil
}

func (s *valueGzipSource) fill() error {
	if s.pending != nil {
		err := s.pending
		s.pending = nil
		return err
	}
	n, err := s.Reader.Read(s.buff[:])
	s.curr, s.length = 0, n
	if n == 0 {
		return err
	}
	s.pending = err
	return nil
}

func (s *valueGzipSource) ReadByte() (byte, error) {
	var b [1]byte
	_, err := io.ReadFull(s, b[:])
	return b[0], err
}

type valueGzipReader struct {
	source      *valueGzipSource
	deflate     *valueInflater
	pending     error
	crc, size   uint32
	closed, eos bool
}

func newValueGzipReader(in io.Reader) (*valueGzipReader, error) {
	if isNil(in) {
		panic(NewNullPointerException())
	}
	r := &valueGzipReader{source: &valueGzipSource{Reader: in}}
	if err := r.readHeader(); err != nil {
		return nil, err
	}
	r.source.buffered = true
	r.deflate = newValueInflater(r.source)
	return r, nil
}

func (r *valueGzipReader) framingByte() (byte, error) {
	b, err := r.source.ReadByte()
	if _, represented := err.(interface{ GetMessage() *string }); represented {
		return b, err
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return 0, NewEOFException()
	}
	return b, bufferedRandomAccessFileIOError(err)
}

func (r *valueGzipReader) readHeader() error {
	var crc uint32
	read := func() (byte, error) {
		b, err := r.framingByte()
		if err == nil {
			crc = crc32.Update(crc, crc32.IEEETable, []byte{b})
		}
		return b, err
	}
	ushort := func() (uint16, error) {
		lo, err := read()
		if err != nil {
			return 0, err
		}
		hi, err := read()
		return uint16(lo) | uint16(hi)<<8, err
	}
	magic, err := ushort()
	if err != nil {
		return err
	}
	if magic != 0x8b1f {
		return NewZipException("Not in GZIP format")
	}
	method, err := read()
	if err != nil {
		return err
	}
	if method != 8 {
		return NewZipException("Unsupported compression method")
	}
	flags, err := read()
	if err != nil {
		return err
	}
	for i := 0; i < 6; i++ {
		if _, err := read(); err != nil {
			return err
		}
	}
	if flags&4 != 0 {
		n, err := ushort()
		if err != nil {
			return err
		}
		for i := 0; i < int(n); i++ {
			if _, err := read(); err != nil {
				return err
			}
		}
	}
	for _, flag := range []byte{8, 16} {
		if flags&flag == 0 {
			continue
		}
		for {
			b, err := read()
			if err != nil {
				return err
			}
			if b == 0 {
				break
			}
		}
	}
	if flags&2 != 0 {
		expected := uint16(crc)
		actual, err := ushort()
		if err != nil {
			return err
		}
		if actual != expected {
			return NewZipException("Corrupt GZIP header")
		}
	}
	r.crc, r.size = 0, 0
	return nil
}

func (r *valueGzipReader) trailerUint() (uint32, error) {
	var b [4]byte
	for i := range b {
		v, err := r.framingByte()
		if err != nil {
			return 0, err
		}
		b[i] = v
	}
	return binary.LittleEndian.Uint32(b[:]), nil
}

func (r *valueGzipReader) readTrailer() error {
	crc, err := r.trailerUint()
	if err != nil {
		return err
	}
	if crc != r.crc {
		return NewZipException("Corrupt GZIP trailer")
	}
	size, err := r.trailerUint()
	if err != nil {
		return err
	}
	if size != r.size {
		return NewZipException("Corrupt GZIP trailer")
	}
	// GZIPInputStream accepts trailing bytes that do not form another header.
	if err := r.readHeader(); err != nil {
		if !isJavaIOException(err) {
			return err
		}
		r.eos = true
		return nil
	}
	r.source.buffered = true
	return r.deflate.Reset(r.source)
}

func (r *valueGzipReader) Read(p []byte) (int, error) {
	if r.closed {
		return 0, NewIOException("Stream closed")
	}
	for {
		if r.eos {
			return 0, io.EOF
		}
		if len(p) == 0 {
			return 0, nil
		}
		if r.pending == io.EOF {
			r.source.buffered = false
			if err := r.readTrailer(); err != nil {
				return 0, err
			}
			r.pending = nil
			continue
		}
		if r.pending != nil {
			if errors.Is(r.pending, io.ErrUnexpectedEOF) {
				return 0, NewEOFException("Unexpected end of ZLIB input stream")
			}
			return 0, bufferedRandomAccessFileIOError(r.pending)
		}
		n, err := r.deflate.Read(p)
		r.pending = err
		if n > 0 {
			r.crc = crc32.Update(r.crc, crc32.IEEETable, p[:n])
			r.size += uint32(n)
			return n, nil
		}
		if err == nil {
			return 0, nil
		}
	}
}

func (r *valueGzipReader) Close() error {
	if r.closed {
		return nil
	}
	// Inflater.end releases decoder state independently of a prior read failure.
	_ = r.deflate.Close()
	if closer, ok := r.source.Reader.(io.Closer); ok {
		if err := closer.Close(); err != nil {
			_, file := r.source.Reader.(*os.File)
			if !file || !errors.Is(err, os.ErrClosed) {
				return bufferedRandomAccessFileIOError(err)
			}
		}
	}
	r.closed, r.eos = true, true
	return nil
}
