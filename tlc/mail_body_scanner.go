package tlc

import (
	"bufio"
	"io"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// Scanner(File) uses Channels.newReader and a strict default-charset decoder.
// On malformed UTF-8, Scanner records the IOException and closes its input
// logically. Reader.read(CharBuffer) does not advance the buffer position when
// decoding throws, so characters produced in that failed read are discarded.
// Previously buffered characters remain, and Scanner starts with 1024 UTF-16
// units, compacting or doubling only when readInput needs space.
type mailBodyScanner struct {
	reader             *bufio.Reader
	buf                []uint16
	position, capacity int
	closed             bool
	leftover           uint16
	hasLeftover        bool
}

func mailExtractScannedBody(stream io.Reader) string {
	s := mailBodyScanner{reader: bufio.NewReaderSize(stream, 8192), capacity: 1024}
	var result strings.Builder
	for s.hasNext() {
		line := s.nextLine()
		if !strings.HasPrefix(line, "@!@!@") {
			result.WriteString(line)
			result.WriteByte('\n')
		}
	}
	return result.String()
}
func (s *mailBodyScanner) hasNext() bool {
	for {
		for _, r := range s.buf[s.position:] {
			if !mailJavaWhitespace(rune(r)) {
				return true
			}
		}
		if s.closed {
			return false
		}
		s.readInput()
	}
}
func (s *mailBodyScanner) nextLine() string {
	for {
		for i := s.position; i < len(s.buf); i++ {
			r := s.buf[i]
			if r != '\r' && r != '\n' && r != 0x85 && r != 0x2028 && r != 0x2029 {
				continue
			}
			if r == '\r' && i+1 == len(s.buf) && !s.closed {
				s.readInput()
				goto retry
			}
			line := string(utf16.Decode(s.buf[s.position:i]))
			s.position = i + 1
			if r == '\r' && s.position < len(s.buf) && s.buf[s.position] == '\n' {
				s.position++
			}
			return line
		}
		if s.closed {
			line := string(utf16.Decode(s.buf[s.position:]))
			s.position = len(s.buf)
			return line
		}
		s.readInput()
	retry:
	}
}
func (s *mailBodyScanner) readInput() {
	if len(s.buf) == s.capacity {
		if s.position > 0 {
			s.buf = append(s.buf[:0], s.buf[s.position:]...)
			s.position = 0
		} else {
			s.capacity *= 2
		}
	}
	space := s.capacity - len(s.buf)
	chunk := make([]uint16, 0, space)
	if s.hasLeftover {
		chunk = append(chunk, s.leftover)
		s.hasLeftover = false
	}
	for len(chunk) < space {
		first, err := s.reader.Peek(1)
		if err == io.EOF {
			break
		}
		if err != nil {
			s.readFailure(err)
			return
		}
		width := 1
		switch {
		case first[0] >= 0xc2 && first[0] <= 0xdf:
			width = 2
		case first[0] >= 0xe0 && first[0] <= 0xef:
			width = 3
		case first[0] >= 0xf0 && first[0] <= 0xf4:
			width = 4
		}
		encoded, err := s.reader.Peek(width)
		if err != nil && err != io.EOF {
			s.readFailure(err)
			return
		}
		r, n := utf8.DecodeRune(encoded)
		if r == utf8.RuneError && n == 1 {
			s.closed = true
			return
		}
		_, _ = s.reader.Discard(n)
		if r <= 0xffff {
			chunk = append(chunk, uint16(r))
		} else {
			high, low := utf16.EncodeRune(r)
			chunk = append(chunk, uint16(high))
			if len(chunk) == space {
				s.leftover = uint16(low)
				s.hasLeftover = true
			} else {
				chunk = append(chunk, uint16(low))
			}
		}
	}
	if len(chunk) == 0 {
		s.closed = true
	}
	s.buf = append(s.buf, chunk...)
}

func (s *mailBodyScanner) readFailure(err error) {
	if _, typed := err.(interface{ GetMessage() *string }); typed && !isJavaIOException(err) {
		panic(err)
	}
	s.closed = true
}
