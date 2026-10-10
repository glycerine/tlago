package tlc

import (
	"bufio"
	"io"
	"os"
	"strings"
	"sync"
	"unicode/utf8"
)

// TLCExt owns one console scanner, retaining its input, charset and read-ahead
// across calls. Its class monitor serializes all interactive access.
var tlcExtConsole struct {
	once  sync.Once
	input *tlcExtConsoleInput
}

func ensureTLCExtConsole() *tlcExtConsoleInput {
	tlcExtConsole.once.Do(func() {
		tlcExtConsole.input = &tlcExtConsoleInput{reader: bufio.NewReader(os.Stdin), charset: javaDefaultCharset()}
	})
	return tlcExtConsole.input
}

type tlcExtConsoleInput struct {
	reader     *bufio.Reader
	charset    string
	started    bool
	closed     bool
	pending    rune
	hasPending bool
}

func (s *tlcExtConsoleInput) readByte() (byte, error) {
	if s.closed {
		return 0, io.EOF
	}
	b, err := s.reader.ReadByte()
	if err != nil {
		// Scanner records an underlying IOException and treats it as end of input.
		s.closed = true
		return 0, io.EOF
	}
	return b, nil
}

func (s *tlcExtConsoleInput) readRune() (rune, error) {
	if s.hasPending {
		s.hasPending = false
		return s.pending, nil
	}
	first, err := s.readByte()
	if err != nil {
		return 0, err
	}
	data := []byte{first}
	if s.charset == "UTF-8" {
		if first < 0x80 {
			return rune(first), nil
		}
		width := 1
		switch {
		case first >= 0xc2 && first <= 0xdf:
			width = 2
		case first >= 0xe0 && first <= 0xef:
			width = 3
		case first >= 0xf0 && first <= 0xf4:
			width = 4
		}
		for i := 1; i < width; i++ {
			b, err := s.readByte()
			if err != nil {
				break
			}
			if b < 0x80 || b > 0xbf || i == 1 && (first == 0xe0 && b < 0xa0 || first == 0xf0 && b < 0x90 || first == 0xf4 && b > 0x8f) {
				_ = s.reader.UnreadByte()
				break
			}
			data = append(data, b)
		}
		text, _ := javaCharsetDecodeUTF8(data, true)
		r, _ := utf8.DecodeRuneInString(text)
		return r, nil
	}
	if s.charset == "UTF-16" || s.charset == "UTF-16BE" || s.charset == "UTF-16LE" {
		if b, err := s.readByte(); err == nil {
			data = append(data, b)
		}
		if !s.started {
			s.started = true
			if s.charset == "UTF-16" {
				s.charset = "UTF-16BE"
				if len(data) == 2 && (data[0] == 0xfe && data[1] == 0xff || data[0] == 0xff && data[1] == 0xfe) {
					if data[0] == 0xff {
						s.charset = "UTF-16LE"
					}
					return s.readRune()
				}
			}
		}
		if len(data) == 2 {
			unit := uint16(data[0])<<8 | uint16(data[1])
			if s.charset == "UTF-16LE" {
				unit = uint16(data[1])<<8 | uint16(data[0])
			}
			if unit >= 0xd800 && unit <= 0xdbff {
				for i := 0; i < 2; i++ {
					b, err := s.readByte()
					if err != nil {
						break
					}
					data = append(data, b)
				}
			}
		}
	}
	text, _ := javaCharsetDecode(data, s.charset, true)
	r, _ := utf8.DecodeRuneInString(text)
	return r, nil
}

func (s *tlcExtConsoleInput) nextLine() string {
	var line strings.Builder
	for {
		r, err := s.readRune()
		if err != nil {
			if line.Len() == 0 {
				panic(NewNoSuchElementException("No line found"))
			}
			return line.String()
		}
		switch r {
		case '\n', '\u0085', '\u2028', '\u2029':
			return line.String()
		case '\r':
			if next, err := s.readRune(); err == nil && next != '\n' {
				s.pending, s.hasPending = next, true
			}
			return line.String()
		default:
			line.WriteRune(r)
		}
	}
}

func tlcExtConsoleTrim(text string) string {
	return strings.TrimFunc(text, func(r rune) bool { return r <= 0x20 })
}
