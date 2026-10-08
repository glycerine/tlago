// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
// Port of JavaCC's tla2sany.parser.SimpleCharStream buffering and position rules.
package tlago

import (
	"fmt"
	"io"
	"unicode/utf16"

	"github.com/glycerine/tlago/tlc"
)

// Readers provide Java UTF-16 code units; decoding bytes belongs to the reader,
// rather than to SimpleCharStream. The generated scanner consumes one unit.
type sanyCharReader interface {
	Read([]uint16) (int, error)
	Close() error
}
type sanyStringCharReader struct {
	units  []uint16
	at     int
	closed bool
}

func newSanyStringCharReader(input string) *sanyStringCharReader {
	return &sanyStringCharReader{units: utf16.Encode([]rune(input))}
}
func (r *sanyStringCharReader) Read(dst []uint16) (int, error) {
	if r.closed {
		return 0, tlc.NewIOException("Stream closed")
	}
	if len(dst) == 0 {
		return 0, nil
	}
	if r.at >= len(r.units) {
		return 0, io.EOF
	}
	n := copy(dst, r.units[r.at:])
	r.at += n
	return n, nil
}
func (r *sanyStringCharReader) Close() error { r.closed = true; return nil }

type sanyCharStream struct {
	bufsize, available, tokenBegin, bufpos int32
	bufline, bufcolumn                     []int32
	column, line                           int32
	prevCharIsCR, prevCharIsLF             bool
	inputStream                            sanyCharReader
	buffer                                 []uint16
	maxNextCharInd, inBuf, tabSize         int32
}

func newSanyCharStream(reader sanyCharReader, startline, startcolumn, buffersize int32) *sanyCharStream {
	if buffersize < 0 {
		panic(tlc.NewNegativeArraySizeException(fmt.Sprint(buffersize)))
	}
	return &sanyCharStream{inputStream: reader, line: startline, column: startcolumn - 1,
		available: buffersize, bufsize: buffersize, bufpos: -1, tabSize: 8,
		buffer: make([]uint16, buffersize), bufline: make([]int32, buffersize), bufcolumn: make([]int32, buffersize)}
}
func (s *sanyCharStream) expandBuff(wrapAround bool) {
	newbuffer := make([]uint16, s.bufsize+2048)
	newbufline := make([]int32, s.bufsize+2048)
	newbufcolumn := make([]int32, s.bufsize+2048)
	func() {
		defer func() {
			if failure := recover(); failure != nil {
				if throwable, ok := failure.(interface{ GetMessage() *string }); ok {
					if message := throwable.GetMessage(); message != nil {
						panic(tlc.NewJavaError(*message))
					}
					panic(tlc.NewJavaError())
				}
				panic(tlc.NewJavaError(fmt.Sprint(failure)))
			}
		}()
		if wrapAround {
			copy(newbuffer, s.buffer[s.tokenBegin:s.bufsize])
			copy(newbuffer[s.bufsize-s.tokenBegin:], s.buffer[:s.bufpos])
			s.buffer = newbuffer
			copy(newbufline, s.bufline[s.tokenBegin:s.bufsize])
			copy(newbufline[s.bufsize-s.tokenBegin:], s.bufline[:s.bufpos])
			s.bufline = newbufline
			copy(newbufcolumn, s.bufcolumn[s.tokenBegin:s.bufsize])
			copy(newbufcolumn[s.bufsize-s.tokenBegin:], s.bufcolumn[:s.bufpos])
			s.bufcolumn = newbufcolumn
			s.bufpos += s.bufsize - s.tokenBegin
			s.maxNextCharInd = s.bufpos
		} else {
			copy(newbuffer, s.buffer[s.tokenBegin:s.bufsize])
			s.buffer = newbuffer
			copy(newbufline, s.bufline[s.tokenBegin:s.bufsize])
			s.bufline = newbufline
			copy(newbufcolumn, s.bufcolumn[s.tokenBegin:s.bufsize])
			s.bufcolumn = newbufcolumn
			s.bufpos -= s.tokenBegin
			s.maxNextCharInd = s.bufpos
		}
	}()
	s.bufsize += 2048
	s.available = s.bufsize
	s.tokenBegin = 0
}
func (s *sanyCharStream) fillBuff() error {
	if s.maxNextCharInd == s.available {
		if s.available == s.bufsize {
			if s.tokenBegin > 2048 {
				s.bufpos = 0
				s.maxNextCharInd = 0
				s.available = s.tokenBegin
			} else if s.tokenBegin < 0 {
				s.bufpos = 0
				s.maxNextCharInd = 0
			} else {
				s.expandBuff(false)
			}
		} else if s.available > s.tokenBegin {
			s.available = s.bufsize
		} else if s.tokenBegin-s.available < 2048 {
			s.expandBuff(true)
		} else {
			s.available = s.tokenBegin
		}
	}
	if s.inputStream == nil {
		panic(tlc.NewNullPointerException())
	}
	n, err := s.inputStream.Read(s.buffer[s.maxNextCharInd:s.available])
	if err == nil {
		s.maxNextCharInd += int32(n)
		return nil
	}
	if err == io.EOF {
		if closeErr := s.inputStream.Close(); closeErr != nil {
			err = closeErr
		}
	}
	s.bufpos--
	s.backup(0)
	if s.tokenBegin == -1 {
		s.tokenBegin = s.bufpos
	}
	return err
}
func (s *sanyCharStream) beginToken() (uint16, error) {
	s.tokenBegin = -1
	c, err := s.readChar()
	if err != nil {
		return 0, err
	}
	s.tokenBegin = s.bufpos
	return c, nil
}
func (s *sanyCharStream) updateLineColumn(c uint16) {
	s.column++
	if s.prevCharIsLF {
		s.prevCharIsLF = false
		s.column = 1
		s.line += s.column
	} else if s.prevCharIsCR {
		s.prevCharIsCR = false
		if c == '\n' {
			s.prevCharIsLF = true
		} else {
			s.column = 1
			s.line += s.column
		}
	}
	switch c {
	case '\r':
		s.prevCharIsCR = true
	case '\n':
		s.prevCharIsLF = true
	case '\t':
		s.column--
		s.column += s.tabSize - s.column%s.tabSize
	}
	s.bufline[s.bufpos] = s.line
	s.bufcolumn[s.bufpos] = s.column
}
func (s *sanyCharStream) readChar() (uint16, error) {
	if s.inBuf > 0 {
		s.inBuf--
		s.bufpos++
		if s.bufpos == s.bufsize {
			s.bufpos = 0
		}
		return s.buffer[s.bufpos], nil
	}
	s.bufpos++
	if s.bufpos >= s.maxNextCharInd {
		if err := s.fillBuff(); err != nil {
			return 0, err
		}
	}
	c := s.buffer[s.bufpos]
	s.updateLineColumn(c)
	return c, nil
}
func (s *sanyCharStream) getEndColumn() int32   { return s.bufcolumn[s.bufpos] }
func (s *sanyCharStream) getEndLine() int32     { return s.bufline[s.bufpos] }
func (s *sanyCharStream) getBeginColumn() int32 { return s.bufcolumn[s.tokenBegin] }
func (s *sanyCharStream) getBeginLine() int32   { return s.bufline[s.tokenBegin] }
func (s *sanyCharStream) backup(amount int32) {
	s.inBuf += amount
	s.bufpos -= amount
	if s.bufpos < 0 {
		s.bufpos += s.bufsize
	}
}
func (s *sanyCharStream) reInit(reader sanyCharReader, startline, startcolumn, buffersize int32) {
	s.inputStream = reader
	s.line = startline
	s.column = startcolumn - 1
	if s.buffer == nil || buffersize != int32(len(s.buffer)) {
		if buffersize < 0 {
			panic(tlc.NewNegativeArraySizeException(fmt.Sprint(buffersize)))
		}
		s.available = buffersize
		s.bufsize = buffersize
		s.buffer = make([]uint16, buffersize)
		s.bufline = make([]int32, buffersize)
		s.bufcolumn = make([]int32, buffersize)
	}
	s.prevCharIsLF = false
	s.prevCharIsCR = false
	s.tokenBegin = 0
	s.inBuf = 0
	s.maxNextCharInd = 0
	s.bufpos = -1
}
func (s *sanyCharStream) getImageUnits() []uint16 {
	if s.bufpos >= s.tokenBegin {
		return append([]uint16(nil), s.buffer[s.tokenBegin:s.bufpos+1]...)
	}
	result := append([]uint16(nil), s.buffer[s.tokenBegin:s.bufsize]...)
	return append(result, s.buffer[:s.bufpos+1]...)
}
func (s *sanyCharStream) getImage() string { return string(utf16.Decode(s.getImageUnits())) }
func (s *sanyCharStream) getSuffix(length int32) []uint16 {
	if length < 0 {
		panic(tlc.NewNegativeArraySizeException(fmt.Sprint(length)))
	}
	result := make([]uint16, length)
	if s.bufpos+1 >= length {
		copy(result, s.buffer[s.bufpos-length+1:s.bufpos+1])
	} else {
		n := length - s.bufpos - 1
		copy(result, s.buffer[s.bufsize-n:s.bufsize])
		copy(result[n:], s.buffer[:s.bufpos+1])
	}
	return result
}
func (s *sanyCharStream) done() { s.buffer = nil; s.bufline = nil; s.bufcolumn = nil }
func (s *sanyCharStream) adjustBeginLineColumn(newLine, newCol int32) {
	start := s.tokenBegin
	var length int32
	if s.bufpos >= s.tokenBegin {
		length = s.bufpos - s.tokenBegin + s.inBuf + 1
	} else {
		length = s.bufsize - s.tokenBegin + s.bufpos + 1 + s.inBuf
	}
	var i, j, k int32
	var nextColDiff, columnDiff int32
	for i < length {
		j = start % s.bufsize
		start++
		k = start % s.bufsize
		if s.bufline[j] != s.bufline[k] {
			break
		}
		s.bufline[j] = newLine
		nextColDiff = columnDiff + s.bufcolumn[k] - s.bufcolumn[j]
		s.bufcolumn[j] = newCol + columnDiff
		columnDiff = nextColDiff
		i++
	}
	if i < length {
		s.bufline[j] = newLine
		newLine++
		s.bufcolumn[j] = newCol + columnDiff
		for {
			previous := i
			i++
			if previous >= length {
				break
			}
			j = start % s.bufsize
			start++
			if s.bufline[j] != s.bufline[start%s.bufsize] {
				s.bufline[j] = newLine
				newLine++
			} else {
				s.bufline[j] = newLine
			}
		}
	}
	s.line = s.bufline[j]
	s.column = s.bufcolumn[j]
}
