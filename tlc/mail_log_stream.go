package tlc

import (
	"fmt"
	"io"
	"sync"
	"unicode/utf16"
	"unicode/utf8"
)

// MailLogPrintStream ports MailSender's LogPrintStream/ErrLogPrintStream.
// Only PrintlnString echoes to System.out/err. Ordinary byte writes and other
// PrintStream overloads write just the log; this is deliberately not a tee.
// Checked I/O sets the PrintStream error flag instead of escaping to callers.
type MailLogPrintStream struct {
	mu      sync.Mutex
	file    io.WriteCloser
	echo    io.Writer
	closed  bool
	closing bool
	trouble bool
}

func NewMailLogPrintStream(file io.WriteCloser, echo io.Writer) *MailLogPrintStream {
	if file == nil {
		panic(NewNullPointerException())
	}
	return &MailLogPrintStream{file: file, echo: echo}
}

func (s *MailLogPrintStream) Write(data []byte) (int, error) {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		s.trouble = true
		return len(data), nil
	}
	n, err := s.file.Write(data)
	s.observe(err)
	if n != len(data) {
		s.trouble = true
	}
	return len(data), nil
}

func (s *MailLogPrintStream) PrintlnString(value *string) {
	if s == nil {
		panic(NewNullPointerException())
	}
	// Source does this before super.println acquires this stream's monitor.
	if s.echo == nil {
		panic(NewNullPointerException())
	}
	encoded := mailPrintUTF8String(javaNullableString(value))
	_, echoErr := fmt.Fprintln(s.echo, encoded)
	if _, typed := echoErr.(interface{ GetMessage() *string }); typed && !isJavaIOException(echoErr) {
		panic(echoErr)
	}
	_, _ = s.Write([]byte(encoded + "\n"))
}

// Properties and other Java string adapters retain isolated UTF-16 units in
// WTF-8. The default PrintStream UTF-8 encoder writes '?' for an unpaired unit,
// while two adjacent surrogate units form a supplementary character.
func mailPrintUTF8String(text string) string {
	var data []byte
	unit := func(value string) (uint16, bool) {
		if len(value) >= 3 && value[0] == 0xed && value[1] >= 0xa0 && value[1] <= 0xbf && value[2]&0xc0 == 0x80 {
			return uint16(value[0]&15)<<12 | uint16(value[1]&63)<<6 | uint16(value[2]&63), true
		}
		return 0, false
	}
	for len(text) > 0 {
		if high, ok := unit(text); ok {
			text = text[3:]
			if high <= 0xdbff {
				if low, ok := unit(text); ok && low >= 0xdc00 {
					data = utf8.AppendRune(data, utf16.DecodeRune(rune(high), rune(low)))
					text = text[3:]
					continue
				}
			}
			data = append(data, '?')
			continue
		}
		r, n := utf8.DecodeRuneInString(text)
		data = utf8.AppendRune(data, r)
		text = text[n:]
	}
	return string(data)
}

func (s *MailLogPrintStream) Flush() error {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		s.trouble = true
		return nil
	}
	if flusher, ok := s.file.(interface{ Flush() error }); ok {
		s.observe(flusher.Flush())
	}
	return nil
}
func (s *MailLogPrintStream) Close() error {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.closing {
		s.closing = true
		if flusher, ok := s.file.(interface{ Flush() error }); ok {
			s.observe(flusher.Flush())
		}
		s.observe(s.file.Close())
		s.closed = true
	}
	return nil
}
func (s *MailLogPrintStream) CheckError() bool {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.mu.Lock()
	if s.closed {
		trouble := s.trouble
		s.mu.Unlock()
		return trouble
	}
	s.mu.Unlock()
	_ = s.Flush()
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.trouble
}
func (s *MailLogPrintStream) observe(err error) {
	if err == nil {
		return
	}
	if _, typed := err.(interface{ GetMessage() *string }); typed && !isJavaIOException(err) {
		panic(err)
	}
	s.trouble = true
}
