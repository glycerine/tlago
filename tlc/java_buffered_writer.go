package tlc

import (
	"io"
	"unicode/utf16"
)

// JavaBufferedWriter models BufferedWriter(OutputStreamWriter): its outer
// buffer counts UTF-16 code units and its encoder owns a separate byte buffer.
// It is exported for root-package adapters that port Java character writers.
type JavaBufferedWriter struct {
	out         io.Writer
	charset     string
	chars       []uint16
	bytes       []byte
	pendingHigh uint16
}

// NewJavaBufferedWriter captures the default charset at construction like
// OutputStreamWriter and allocates Java's two default-sized buffers.
func NewJavaBufferedWriter(out io.Writer) *JavaBufferedWriter {
	if isNil(out) {
		panic(NewNullPointerException())
	}
	return &JavaBufferedWriter{
		out:     out,
		charset: javaDefaultCharset(),
		chars:   make([]uint16, 0, 8192),
		bytes:   make([]byte, 0, 8192),
	}
}

func (w *JavaBufferedWriter) Write(p []byte) (int, error) {
	if w == nil {
		panic(NewNullPointerException())
	}
	units := utf16.Encode([]rune(string(p)))
	for len(units) > 0 {
		count := min(8192-len(w.chars), len(units))
		w.chars = append(w.chars, units[:count]...)
		units = units[count:]
		if len(w.chars) == 8192 {
			if err := w.flushChars(false); err != nil {
				return 0, err
			}
		}
	}
	return len(p), nil
}

func (w *JavaBufferedWriter) WriteString(text string) (int, error) {
	if w == nil {
		panic(NewNullPointerException())
	}
	return w.Write([]byte(text))
}

func (w *JavaBufferedWriter) Flush() error {
	if w == nil {
		panic(NewNullPointerException())
	}
	flushErr := w.flushChars(true)
	if err := w.flushBytes(); flushErr == nil {
		flushErr = err
	}
	return flushErr
}

func (w *JavaBufferedWriter) flushChars(final bool) error {
	if len(w.chars) == 0 && w.pendingHigh == 0 {
		return nil
	}
	units := w.chars
	w.chars = w.chars[:0]
	if w.pendingHigh != 0 {
		units = append([]uint16{w.pendingHigh}, units...)
		w.pendingHigh = 0
	}
	if !final && len(units) > 0 && units[len(units)-1] >= 0xD800 && units[len(units)-1] <= 0xDBFF {
		w.pendingHigh = units[len(units)-1]
		units = units[:len(units)-1]
	}
	encoded, err := javaCharsetEncode(string(utf16.Decode(units)), w.charset, false)
	if err != nil {
		return err
	}
	return w.writeEncoded(encoded)
}

func (w *JavaBufferedWriter) writeEncoded(data []byte) error {
	for len(data) > 0 {
		if len(w.bytes) == 8192 {
			if err := w.flushBytes(); err != nil {
				return err
			}
		}
		count := min(8192-len(w.bytes), len(data))
		w.bytes = append(w.bytes, data[:count]...)
		data = data[count:]
	}
	return nil
}

func (w *JavaBufferedWriter) flushBytes() error {
	if len(w.bytes) == 0 {
		return nil
	}
	offset := 0
	for offset < len(w.bytes) {
		written, err := w.out.Write(w.bytes[offset:])
		offset += written
		if err != nil {
			copy(w.bytes, w.bytes[offset:])
			w.bytes = w.bytes[:len(w.bytes)-offset]
			return err
		}
		if written == 0 {
			copy(w.bytes, w.bytes[offset:])
			w.bytes = w.bytes[:len(w.bytes)-offset]
			return io.ErrShortWrite
		}
	}
	w.bytes = w.bytes[:0]
	return nil
}

// JavaLineSeparator exposes System.lineSeparator for root-package adapters
// that port Java BufferedWriter.newLine calls.
func JavaLineSeparator() string { return javaFormatLineSeparator() }

// WriteJavaBufferedFile models an explicit BufferedWriter(FileWriter) write and
// close sequence. A body write failure leaves the writer open; close always
// closes FileWriter after its flush and suppresses a secondary close failure.
func WriteJavaBufferedFile(path, content string) error {
	file, err := openFileOutputStream(path)
	if err != nil {
		return err
	}
	writer := NewJavaBufferedWriter(file)
	if _, err := writer.WriteString(content); err != nil {
		return bufferedRandomAccessFileIOError(err)
	}
	flushErr := bufferedRandomAccessFileIOError(writer.Flush())
	closeErr := bufferedRandomAccessFileIOError(file.Close())
	if flushErr != nil {
		javaSuppressCloseError(flushErr, closeErr)
		return flushErr
	}
	return closeErr
}
