// Core Json.textSerialize's Files.newBufferedWriter path, following OpenJDK
// BufferedWriter/StreamEncoder. GPLv2 with Classpath Exception notices are in
// tlc/licenses/; the TLC module's MIT notice is retained in test_vectors/.
package tlc

import (
	"bytes"
	"os"
	"unicode/utf8"
)

func jsonWriteNDJSONFile(path *StringValue, payload *TupleValue, options []*StringValue, charset *StringValue) (err error) {
	// Only construction/writing/closing lies in Json's catch(Exception).
	defer func() {
		if failure := recover(); failure != nil {
			if exception, ok := failure.(error); ok && !isJavaError(exception) {
				err = exception
			} else {
				panic(failure)
			}
		}
		if err != nil && !isJavaError(err) {
			err = NewRuntimeExceptionFromCause(err)
		}
	}()
	if path == nil {
		return NewNullPointerException()
	}
	filePath, err := ioUtilsTXTPath(path.RawString())
	if err != nil {
		return err
	}
	if charset == nil {
		return NewNullPointerException()
	}
	name, err := javaCharsetName(charset.RawString())
	if err != nil {
		return err
	}
	enums, err := ioUtilsTXTEnums(options)
	if err != nil {
		return err
	}
	flags, err := ioUtilsFileOptionsFromEnums(enums)
	if err != nil {
		return err
	}
	file, err := ioUtilsTXTOpen(filePath, flags)
	if err != nil {
		return err
	}
	writer := newJSONFileWriter(file, name)
	defer func() {
		failure := recover()
		closeErr := writer.close()
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if closeErr != nil {
			if err == nil {
				err = closeErr
			} else {
				javaSuppressCloseError(err, closeErr)
			}
		}
		ioUtilsTXTDeleteAfterClose(filePath, flags)
	}()
	for _, elem := range payload.Elems {
		var node bytes.Buffer
		if err := jsonWriteValue(&node, elem); err != nil {
			return err
		}
		node.WriteByte('\n')
		if err := writer.write(node.String()); err != nil {
			return err
		}
	}
	return nil
}

// BufferedWriter keeps 8192 UTF-16 characters on TLC's platform threads.
// StreamEncoder separately keeps 8192 bytes; a failed write retains each
// buffer's source position/limit, which affects close and partial file contents.
type jsonFileWriter struct {
	file             *os.File
	charset          string
	chars            []uint16
	byteBuffer       [8192]byte
	position, limit  int
	leftover         uint16
	hasLeftover, bom bool
}

func newJSONFileWriter(file *os.File, charset string) *jsonFileWriter {
	return &jsonFileWriter{file: file, charset: charset, chars: make([]uint16, 0, 8192), limit: 8192, bom: charset == "UTF-16"}
}

func (w *jsonFileWriter) write(text string) error {
	units := javaStringUTF16(text)
	for len(units) > 0 {
		n := min(len(units), 8192-len(w.chars))
		w.chars = append(w.chars, units[:n]...)
		units = units[n:]
		if len(w.chars) == 8192 {
			if err := w.flushChars(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *jsonFileWriter) flushChars() error {
	if len(w.chars) == 0 {
		return nil
	}
	if err := w.encode(w.chars); err != nil {
		return err
	}
	w.chars = w.chars[:0]
	return nil
}

func (w *jsonFileWriter) encode(units []uint16) error {
	if w.hasLeftover {
		if len(units) == 0 {
			return nil
		}
		// flushLeftoverChar consumes one input character even when encoding it
		// with the retained character fails. The retained flag clears on success.
		if err := w.encodeUnits([]uint16{w.leftover, units[0]}); err != nil {
			return err
		}
		w.hasLeftover = false
		units = units[1:]
	}
	return w.encodeUnits(units)
}

func (w *jsonFileWriter) encodeUnits(units []uint16) error {
	if len(units) > 0 && w.bom {
		if err := w.put([]byte{0xfe, 0xff}); err != nil {
			return err
		}
		w.bom = false
	}
	for i := 0; i < len(units); i++ {
		unit := units[i]
		count := 1
		if unit >= 0xd800 && unit <= 0xdbff {
			if i+1 == len(units) {
				w.leftover, w.hasLeftover = unit, true
				return nil
			}
			if units[i+1] < 0xdc00 || units[i+1] > 0xdfff {
				return NewMalformedInputException(1)
			}
			count = 2
		} else if unit >= 0xdc00 && unit <= 0xdfff {
			return NewMalformedInputException(1)
		}
		var encoded []byte
		switch w.charset {
		case "US-ASCII", "ISO-8859-1":
			limit := uint16(127)
			if w.charset == "ISO-8859-1" {
				limit = 255
			}
			if unit > limit {
				return NewUnmappableCharacterException(count)
			}
			encoded = []byte{byte(unit)}
		case "UTF-8":
			codepoint := rune(unit)
			if count == 2 {
				codepoint = 0x10000 + (rune(unit)-0xd800)*1024 + rune(units[i+1]) - 0xdc00
			}
			var buffer [4]byte
			n := utf8.EncodeRune(buffer[:], codepoint)
			encoded = buffer[:n]
		default:
			encoded = javaCharsetUTF16Bytes(units[i:i+count], false, w.charset == "UTF-16LE")
		}
		if err := w.put(encoded); err != nil {
			return err
		}
		i += count - 1
	}
	return nil
}

func (w *jsonFileWriter) put(data []byte) error {
	if len(data) > w.limit-w.position {
		if err := w.writeBytes(); err != nil {
			return err
		}
	}
	copy(w.byteBuffer[w.position:], data)
	w.position += len(data)
	return nil
}

func (w *jsonFileWriter) writeBytes() error {
	w.limit, w.position = w.position, 0 // ByteBuffer.flip precedes native I/O.
	if w.limit > 0 {
		if _, err := w.file.Write(w.byteBuffer[:w.limit]); err != nil {
			return ioUtilsTXTChannelError(err)
		}
	}
	w.limit = len(w.byteBuffer) // ByteBuffer.clear only follows a successful write.
	return nil
}

func (w *jsonFileWriter) close() (err error) {
	err = w.flushChars()
	// BufferedWriter closes StreamEncoder even if flushing characters fails.
	closeErr := w.closeEncoder()
	if err == nil {
		err = closeErr
	} else {
		javaSuppressCloseError(err, closeErr)
	}
	return err
}

func (w *jsonFileWriter) closeEncoder() (err error) {
	defer func() {
		if closeErr := w.file.Close(); closeErr != nil {
			if err == nil {
				err = ioUtilsTXTChannelError(closeErr)
			} else {
				javaSuppressCloseError(err, ioUtilsTXTChannelError(closeErr))
			}
		}
	}()
	if w.hasLeftover {
		return NewMalformedInputException(1)
	}
	if w.position > 0 {
		return w.writeBytes()
	}
	return nil
}
