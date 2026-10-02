package tlc

import (
	"bufio"
	"errors"
	"io"
	"os"
	"unicode/utf16"
	"unicode/utf8"
)

const ModelInJarPath = "/model/"
const ModelCheckFileBasename = "MC"

// ModelInJar uses the same ordered directory/archive/bundled classpath adapter
// as the filename resolvers. Resource lookup does not consult user/library dirs.
type ModelInJar struct {
	classpath []FilenameClasspathEntry
}

func NewModelInJar(classpath ...[]FilenameClasspathEntry) *ModelInJar {
	entries := filenameDefaultClasspath()
	if len(classpath) > 0 {
		entries = classpath[0]
	}
	return &ModelInJar{classpath: append([]FilenameClasspathEntry(nil), entries...)}
}

func (m *ModelInJar) resource(name string) *filenameResource {
	if m == nil {
		panic(NewNullPointerException())
	}
	return (&SimpleFilenameToStream{classpath: m.classpath}).findClasspath("model/" + name)
}

func (m *ModelInJar) HasModel() bool { return m.resource("MC.tla") != nil }
func (m *ModelInJar) HasCfg() bool   { return m.resource("MC.cfg") != nil }

func (m *ModelInJar) Resolver(options ...FilenameResolverOptions) *SimpleFilenameToStream {
	var configuration FilenameResolverOptions
	if len(options) > 0 {
		configuration = options[0]
	}
	configuration.Classpath = append([]FilenameClasspathEntry{}, m.classpath...)
	return NewInJarFilenameToStream(ModelInJarPath, configuration)
}

// resourceStream is Class.getResourceAsStream: checked open failures return
// null. Unchecked lookup/provider failures still propagate.
func (m *ModelInJar) resourceStream(name string) io.ReadCloser {
	resource := m.resource(name)
	if resource == nil {
		return nil
	}
	stream, err := resource.open()
	if err != nil {
		return nil
	}
	return stream
}

func (m *ModelInJar) GetCfg() *TLAFile {
	stream := m.resourceStream("MC.cfg")
	if stream != nil {
		defer stream.Close()
	}
	target, err := os.CreateTemp("", "MC*.cfg")
	if err != nil {
		return NewTLAFile("", false, nil)
	}
	defer target.Close()
	if stream == nil {
		panic(NewNullPointerException()) // Source creates the target first.
	}
	if _, err := io.Copy(target, stream); err != nil {
		return NewTLAFile("", false, nil)
	}
	if err := target.Close(); err != nil {
		return NewTLAFile("", false, nil)
	}
	return NewTLAFile(target.Name(), false, nil)
}

func (m *ModelInJar) LoadProperties() bool {
	stream := m.resourceStream("generated.properties")
	if stream == nil {
		return false
	}
	defer stream.Close()
	properties := NewInsMap[string, string]()
	if err := loadJavaProperties(stream, properties); err != nil {
		printDistributedFileException(NewIOException(distributedIOMessage(err)))
	} else if err := stream.Close(); err != nil {
		printDistributedFileException(NewIOException(distributedIOMessage(err)))
	}
	for key, value := range properties.All() {
		// System.getProperties().setProperty does not mutate already-loaded
		// TLCGlobals fields. Keep it distinct from the runner startup setter.
		tlcSetSystemProperty(key, value)
	}
	return true
}

// loadJavaProperties ports Properties.load(InputStream): bytes are Latin-1,
// logical lines fold continuations, and conversion decodes Java escapes.
func loadJavaProperties(stream io.Reader, properties *InsMap[string, string]) error {
	reader := bufio.NewReaderSize(stream, 8192)
	for {
		line, err := javaPropertyLine(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		keyEnd, valueStart := 0, len(line)
		separator, backslash := false, false
		for keyEnd < len(line) {
			char := line[keyEnd]
			if (char == '=' || char == ':') && !backslash {
				valueStart, separator = keyEnd+1, true
				break
			}
			if javaPropertyWhitespace(char) && !backslash {
				valueStart = keyEnd + 1
				break
			}
			if char == '\\' {
				backslash = !backslash
			} else {
				backslash = false
			}
			keyEnd++
		}
		for valueStart < len(line) {
			char := line[valueStart]
			if !javaPropertyWhitespace(char) {
				if separator || (char != '=' && char != ':') {
					break
				}
				separator = true
			}
			valueStart++
		}
		key := javaPropertyConvert(line[:keyEnd])
		value := javaPropertyConvert(line[valueStart:])
		properties.Set(key, value)
	}
}

func javaPropertyWhitespace(char byte) bool { return char == ' ' || char == '\t' || char == '\f' }

// javaPropertyLine follows the installed JDK Properties.LineReader. Comments
// are recognized whenever the logical line has no characters yet.
func javaPropertyLine(reader *bufio.Reader) ([]byte, error) {
	line := make([]byte, 0, 1024)
	skipWhitespace, appended, backslash := true, false, false
	for {
		char, err := reader.ReadByte()
		if err != nil {
			if errors.Is(err, io.EOF) && len(line) > 0 {
				if backslash {
					line = line[:len(line)-1]
				}
				return line, nil
			}
			return nil, err
		}
		if skipWhitespace {
			if javaPropertyWhitespace(char) || (!appended && (char == '\r' || char == '\n')) {
				continue
			}
			skipWhitespace, appended = false, false
		}
		if len(line) == 0 && (char == '#' || char == '!') {
			for char != '\r' && char != '\n' {
				char, err = reader.ReadByte()
				if err != nil {
					return nil, err
				}
			}
			skipWhitespace = true
			continue
		}
		if char != '\r' && char != '\n' {
			line = append(line, char)
			if char == '\\' {
				backslash = !backslash
			} else {
				backslash = false
			}
			continue
		}
		if len(line) == 0 {
			skipWhitespace = true
			continue
		}
		// LineReader refills the input buffer before returning a line whose
		// terminator falls at the buffer end. Peek retains deferred read errors.
		if reader.Buffered() == 0 {
			if _, err := reader.Peek(1); err != nil {
				if !errors.Is(err, io.EOF) {
					return nil, err
				}
				if backslash {
					line = line[:len(line)-1]
				}
				return line, nil
			}
		}
		if !backslash {
			return line, nil
		}
		line = line[:len(line)-1]
		skipWhitespace, appended, backslash = true, true, false
		if char == '\r' {
			if next, err := reader.Peek(1); err == nil && next[0] == '\n' {
				_, _ = reader.ReadByte()
			}
		}
	}
}

func javaPropertyConvert(input []byte) string {
	chars := make([]uint16, 0, len(input))
	for i := 0; i < len(input); i++ {
		char := input[i]
		if char != '\\' {
			chars = append(chars, uint16(char))
			continue
		}
		i++
		if i >= len(input) {
			panic(NewArrayIndexOutOfBoundsException(i, len(input)))
		}
		char = input[i]
		if char == 'u' {
			if i+4 >= len(input) {
				panic(NewIllegalArgumentException("Malformed \\uxxxx encoding."))
			}
			var value uint16
			for j := 0; j < 4; j++ {
				i++
				digit := input[i]
				var hex byte
				switch {
				case digit >= '0' && digit <= '9':
					hex = digit - '0'
				case digit >= 'a' && digit <= 'f':
					hex = digit - 'a' + 10
				case digit >= 'A' && digit <= 'F':
					hex = digit - 'A' + 10
				default:
					panic(NewIllegalArgumentException("Malformed \\uxxxx encoding."))
				}
				value = value<<4 | uint16(hex)
			}
			chars = append(chars, value)
			continue
		}
		switch char {
		case 't':
			char = '\t'
		case 'r':
			char = '\r'
		case 'n':
			char = '\n'
		case 'f':
			char = '\f'
		}
		chars = append(chars, uint16(char))
	}
	var encoded []byte
	for i := 0; i < len(chars); i++ {
		char := rune(chars[i])
		if char >= 0xd800 && char <= 0xdbff && i+1 < len(chars) && chars[i+1] >= 0xdc00 && chars[i+1] <= 0xdfff {
			char = utf16.DecodeRune(char, rune(chars[i+1]))
			i++
		}
		if char >= 0xd800 && char <= 0xdfff {
			// Retain isolated Java UTF-16 units losslessly as WTF-8 bytes.
			encoded = append(encoded, byte(0xe0|char>>12), byte(0x80|(char>>6)&0x3f), byte(0x80|char&0x3f))
		} else {
			encoded = utf8.AppendRune(encoded, char)
		}
	}
	return string(encoded)
}
