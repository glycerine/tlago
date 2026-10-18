package tlc

import (
	"bufio"
	"io"
)

// javaReplacingLineReader is the FileReader/BufferedReader.readLine path used
// by source helpers that decode with the default charset's replacement action.
// It retains the decoder state needed for UTF-16 BOMs and CRLF handling.
type javaReplacingLineReader struct {
	reader        *bufio.Reader
	charset       string
	decodeCharset string
	wide          bool
	littleEndian  bool
	atStart       bool
	skipLF        bool
}

func newJavaReplacingLineReader(input io.Reader, charset string) (*javaReplacingLineReader, error) {
	name, err := javaCharsetName(charset)
	if err != nil {
		return nil, err
	}
	return &javaReplacingLineReader{
		reader:        bufio.NewReaderSize(input, 8192),
		charset:       name,
		decodeCharset: name,
		wide:          name == "UTF-16" || name == "UTF-16BE" || name == "UTF-16LE",
		littleEndian:  name == "UTF-16LE",
		atStart:       true,
	}, nil
}

func (r *javaReplacingLineReader) ReadLine() (*string, error) {
	if r == nil {
		panic(NewNullPointerException())
	}
	if !r.wide {
		return r.readNarrowLine()
	}
	return r.readWideLine()
}

func (r *javaReplacingLineReader) decodeLine(raw []byte) (*string, error) {
	text, err := javaCharsetDecode(raw, r.decodeCharset, true)
	if err != nil {
		return nil, err
	}
	return &text, nil
}

func (r *javaReplacingLineReader) readNarrowLine() (*string, error) {
	var raw []byte
	for {
		value, err := r.reader.ReadByte()
		if err != nil {
			if err == io.EOF {
				if len(raw) == 0 {
					return nil, nil
				}
				return r.decodeLine(raw)
			}
			return nil, err
		}
		if r.skipLF {
			r.skipLF = false
			if value == '\n' {
				continue
			}
		}
		switch value {
		case '\n':
			return r.decodeLine(raw)
		case '\r':
			r.skipLF = true
			return r.decodeLine(raw)
		default:
			raw = append(raw, value)
		}
	}
}

func (r *javaReplacingLineReader) readWideUnit() (raw []byte, unit uint16, eof bool, err error) {
	first, err := r.reader.ReadByte()
	if err != nil {
		if err == io.EOF {
			return nil, 0, true, nil
		}
		return nil, 0, false, err
	}
	second, err := r.reader.ReadByte()
	if err != nil {
		if err == io.EOF {
			return []byte{first}, 0, true, nil
		}
		return nil, 0, false, err
	}
	raw = []byte{first, second}
	if r.littleEndian {
		unit = uint16(first) | uint16(second)<<8
	} else {
		unit = uint16(first)<<8 | uint16(second)
	}
	return raw, unit, false, nil
}

func (r *javaReplacingLineReader) readWideLine() (*string, error) {
	var raw []byte
	for {
		encoded, unit, eof, err := r.readWideUnit()
		if err != nil {
			return nil, err
		}
		if len(encoded) == 0 && eof {
			if len(raw) == 0 {
				return nil, nil
			}
			return r.decodeLine(raw)
		}
		if r.atStart {
			r.atStart = false
			if r.charset == "UTF-16" && len(encoded) == 2 {
				switch {
				case encoded[0] == 0xfe && encoded[1] == 0xff:
					r.decodeCharset = "UTF-16BE"
					if eof {
						return nil, nil
					}
					continue
				case encoded[0] == 0xff && encoded[1] == 0xfe:
					r.littleEndian = true
					r.decodeCharset = "UTF-16LE"
					if eof {
						return nil, nil
					}
					continue
				}
			}
		}
		if len(encoded) != 2 {
			raw = append(raw, encoded...)
			return r.decodeLine(raw)
		}
		if r.skipLF {
			r.skipLF = false
			if unit == '\n' {
				if eof {
					return nil, nil
				}
				continue
			}
		}
		if unit >= 0xd800 && unit <= 0xdbff {
			raw = append(raw, encoded...)
			next, _, nextEOF, nextErr := r.readWideUnit()
			if nextErr != nil {
				return nil, nextErr
			}
			raw = append(raw, next...)
			if nextEOF {
				return r.decodeLine(raw)
			}
			continue
		}
		switch unit {
		case '\n':
			return r.decodeLine(raw)
		case '\r':
			r.skipLF = true
			return r.decodeLine(raw)
		default:
			raw = append(raw, encoded...)
			if eof {
				return r.decodeLine(raw)
			}
		}
	}
}
