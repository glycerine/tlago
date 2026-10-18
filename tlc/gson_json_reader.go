package tlc

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf16"
)

const gsonMalformedJSONHelp = "\nSee https://github.com/google/gson/blob/main/Troubleshooting.md#malformed-json"

type gsonJSONReader struct {
	input      []uint16
	pos        int
	initialBOM bool
}

type gsonNumberNode string
type gsonJSONObject struct {
	names  []string
	values []any
}

func parseGsonJSON(text string) (Value, error) {
	r := &gsonJSONReader{input: utf16.Encode([]rune(text))}
	if !r.eof() && r.peek() == '\ufeff' {
		r.pos++
		r.initialBOM = true
	}
	if err := r.skipSpace(true, "$"); err != nil {
		return nil, NewJsonSyntaxExceptionFromCause(err)
	}
	r.consumeNonExecutePrefix()
	if err := r.skipSpace(true, "$"); err != nil {
		return nil, NewJsonSyntaxExceptionFromCause(err)
	}
	if r.eof() {
		return nil, nil
	}
	node, err := r.readValue("$")
	if err != nil {
		return nil, wrapGsonSyntaxError(err)
	}
	// JsonParser.parseReader deliberately omits its strict end-of-document
	// check for JsonNull before Json.getValue rejects the unsupported node.
	if node == nil {
		return gsonJSONNodeToValue(node)
	}
	if err := r.skipSpace(false, "$"); err != nil {
		return nil, NewJsonSyntaxExceptionFromCause(err)
	}
	if !r.eof() {
		r.consumeTrailingPeek()
		return nil, NewJsonSyntaxExceptionFromCause(r.malformed(
			"Use JsonReader.setStrictness(Strictness.LENIENT) to accept malformed JSON", "$"))
	}
	return gsonJSONNodeToValue(node)
}

func wrapGsonSyntaxError(err error) error {
	switch err.(type) {
	case *MalformedJsonException, *EOFException:
		return NewJsonSyntaxExceptionFromCause(err)
	default:
		return err
	}
}

func (r *gsonJSONReader) readValue(path string) (any, error) {
	if err := r.skipSpace(true, path); err != nil {
		return nil, err
	}
	if r.eof() {
		return nil, r.endOfInput(path)
	}
	switch r.peek() {
	case '{':
		return r.readObject(path)
	case '[':
		return r.readArray(path)
	case '"', '\'':
		value, err := r.readQuoted(r.peek(), path)
		if err != nil {
			return nil, err
		}
		return value, nil
	default:
		literal := r.readUnquoted()
		if literal == "" {
			return nil, r.malformed("Expected value", path)
		}
		switch strings.ToLower(literal) {
		case "true":
			return true, nil
		case "false":
			return false, nil
		case "null":
			return nil, nil
		}
		if isGsonJSONNumber(literal) {
			return gsonNumberNode(literal), nil
		}
		return literal, nil
	}
}

func (r *gsonJSONReader) readArray(path string) (any, error) {
	r.pos++
	values := make([]any, 0)
	afterSeparator := false
	for {
		nextPath := gsonArrayPath(path, len(values))
		if err := r.skipSpace(true, nextPath); err != nil {
			return nil, err
		}
		if r.eof() {
			return nil, r.endOfInput(nextPath)
		}
		if r.peek() == ']' {
			r.pos++
			if afterSeparator {
				values = append(values, nil)
			}
			return values, nil
		}
		if r.peek() == ',' || r.peek() == ';' {
			r.pos++
			values = append(values, nil)
			afterSeparator = true
			continue
		}
		value, err := r.readValue(nextPath)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		afterSeparator = false
		nextPath = gsonArrayPath(path, len(values))
		if err := r.skipSpace(true, nextPath); err != nil {
			return nil, err
		}
		if r.eof() {
			return nil, r.endOfInput(nextPath)
		}
		switch r.peek() {
		case ']':
			r.pos++
			return values, nil
		case ',', ';':
			r.pos++
			afterSeparator = true
		default:
			r.pos++
			return nil, r.malformed("Unterminated array", nextPath)
		}
	}
}

func (r *gsonJSONReader) readObject(path string) (any, error) {
	r.pos++
	names := make([]string, 0)
	values := make([]any, 0)
	indices := make(map[string]int)
	lastPath := path
	first := true
	for {
		if err := r.skipSpace(true, lastPath); err != nil {
			return nil, err
		}
		if r.eof() {
			return nil, r.endOfInput(lastPath)
		}
		if r.peek() == '}' {
			r.pos++
			if !first {
				return nil, r.malformed("Expected name", lastPath)
			}
			return gsonJSONObject{names: names, values: values}, nil
		}
		var name string
		var err error
		if r.peek() == '"' || r.peek() == '\'' {
			name, err = r.readQuoted(r.peek(), lastPath)
		} else {
			name = r.readUnquoted()
			if name == "" {
				err = r.malformed("Expected name", lastPath)
			}
		}
		if err != nil {
			return nil, err
		}
		lastPath = gsonObjectPath(path, name)
		if err := r.skipSpace(true, lastPath); err != nil {
			return nil, err
		}
		if r.eof() {
			return nil, r.endOfInput(lastPath)
		}
		if r.peek() == ':' {
			r.pos++
		} else if r.peek() == '=' {
			r.pos++
			if !r.eof() && r.peek() == '>' {
				r.pos++
			}
		} else {
			r.pos++
			return nil, r.malformed("Expected ':'", lastPath)
		}
		if err := r.skipSpace(true, lastPath); err != nil {
			return nil, err
		}
		if r.eof() {
			return nil, r.endOfInput(lastPath)
		}
		if r.peek() == '}' || r.peek() == ',' || r.peek() == ';' {
			return nil, r.malformed("Expected value", lastPath)
		}
		value, err := r.readValue(lastPath)
		if err != nil {
			return nil, err
		}
		if index, found := indices[name]; found {
			values[index] = value
		} else {
			indices[name] = len(names)
			names = append(names, name)
			values = append(values, value)
		}
		first = false
		if err := r.skipSpace(true, lastPath); err != nil {
			return nil, err
		}
		if r.eof() {
			return nil, r.endOfInput(lastPath)
		}
		switch r.peek() {
		case '}':
			r.pos++
			return gsonJSONObject{names: names, values: values}, nil
		case ',', ';':
			r.pos++
		default:
			r.pos++
			return nil, r.malformed("Unterminated object", lastPath)
		}
	}
}

func gsonJSONNodeToValue(node any) (Value, error) {
	switch value := node.(type) {
	case nil:
		return nil, NewIOException(jsonUnsupportedNullMessage)
	case bool:
		return NewBoolValue(value), nil
	case string:
		return NewStringValue(value), nil
	case gsonNumberNode:
		integer, err := jsonNumberToJavaInt(string(value))
		if err != nil {
			return nil, err
		}
		return NewIntValue(integer), nil
	case []any:
		values := make([]Value, len(value))
		for i := range value {
			converted, err := gsonJSONNodeToValue(value[i])
			if err != nil {
				return nil, err
			}
			values[i] = converted
		}
		return NewTupleValue(values), nil
	case gsonJSONObject:
		names := make([]*UniqueString, len(value.names))
		values := make([]Value, len(value.values))
		for i := range value.names {
			names[i] = UniqueStringOf(value.names[i])
			converted, err := gsonJSONNodeToValue(value.values[i])
			if err != nil {
				return nil, err
			}
			values[i] = converted
		}
		return NewRecordValue(names, values, false), nil
	default:
		return nil, NewIOException(fmt.Sprintf("Cannot convert value: unsupported JSON value %v", node))
	}
}

func (r *gsonJSONReader) readQuoted(quote uint16, path string) (string, error) {
	r.pos++
	units := make([]uint16, 0)
	for !r.eof() {
		ch := r.peek()
		r.pos++
		if ch == quote {
			return javaStringFromUTF16(units), nil
		}
		if ch != '\\' {
			units = append(units, ch)
			continue
		}
		if r.eof() {
			return "", r.malformed("Unterminated escape sequence", path)
		}
		escaped := r.peek()
		r.pos++
		switch escaped {
		case 'u':
			if r.pos+4 > len(r.input) {
				return "", r.malformed("Unterminated escape sequence", path)
			}
			digits := javaStringFromUTF16(r.input[r.pos : r.pos+4])
			value, err := strconv.ParseUint(digits, 16, 16)
			if err != nil {
				return "", r.malformed("Malformed Unicode escape \\u"+digits, path)
			}
			units = append(units, uint16(value))
			r.pos += 4
		case 't':
			units = append(units, '\t')
		case 'b':
			units = append(units, '\b')
		case 'n':
			units = append(units, '\n')
		case 'r':
			units = append(units, '\r')
		case 'f':
			units = append(units, '\f')
		case '\n':
			// Lenient JsonReader preserves an escaped physical line feed.
			units = append(units, '\n')
		case '\'', '"', '\\', '/':
			units = append(units, escaped)
		default:
			return "", r.malformed("Invalid escape sequence", path)
		}
	}
	return "", r.malformed("Unterminated string", path)
}

func (r *gsonJSONReader) skipSpace(lenient bool, path string) error {
	for !r.eof() {
		switch r.peek() {
		case ' ', '\t', '\r', '\n':
			r.pos++
			continue
		case '#':
			if !lenient {
				r.pos++
				return r.malformed("Use JsonReader.setStrictness(Strictness.LENIENT) to accept malformed JSON", path)
			}
			r.skipToLineEnd()
			continue
		case '/':
			if r.pos+1 >= len(r.input) || (r.input[r.pos+1] != '/' && r.input[r.pos+1] != '*') {
				return nil
			}
			if !lenient {
				r.pos++
				return r.malformed("Use JsonReader.setStrictness(Strictness.LENIENT) to accept malformed JSON", path)
			}
			if r.input[r.pos+1] == '/' {
				r.pos += 2
				r.skipToLineEnd()
				continue
			}
			r.pos += 2
			for r.pos+1 < len(r.input) && !(r.input[r.pos] == '*' && r.input[r.pos+1] == '/') {
				r.pos++
			}
			if r.pos+1 >= len(r.input) {
				return r.malformed("Unterminated comment", path)
			}
			r.pos += 2
			continue
		}
		return nil
	}
	return nil
}

func (r *gsonJSONReader) readUnquoted() string {
	start := r.pos
	for !r.eof() && !gsonLiteralDelimiter(r.peek()) {
		r.pos++
	}
	return javaStringFromUTF16(r.input[start:r.pos])
}

func gsonLiteralDelimiter(ch uint16) bool {
	switch ch {
	case '/', '\\', ';', '#', '=', '{', '}', '[', ']', ':', ',', ' ', '\t', '\f', '\r', '\n':
		return true
	}
	return false
}

func isGsonJSONNumber(text string) bool {
	if text == "" {
		return false
	}
	i := 0
	if text[i] == '-' {
		i++
		if i == len(text) {
			return false
		}
	}
	if text[i] == '0' {
		i++
		if i < len(text) && text[i] >= '0' && text[i] <= '9' {
			return false
		}
	} else if text[i] >= '1' && text[i] <= '9' {
		for i < len(text) && text[i] >= '0' && text[i] <= '9' {
			i++
		}
	} else {
		return false
	}
	if i < len(text) && text[i] == '.' {
		i++
		start := i
		for i < len(text) && text[i] >= '0' && text[i] <= '9' {
			i++
		}
		if i == start {
			return false
		}
	}
	if i < len(text) && (text[i] == 'e' || text[i] == 'E') {
		i++
		if i < len(text) && (text[i] == '+' || text[i] == '-') {
			i++
		}
		start := i
		for i < len(text) && text[i] >= '0' && text[i] <= '9' {
			i++
		}
		if i == start {
			return false
		}
	}
	return i == len(text)
}

func (r *gsonJSONReader) malformed(message, path string) error {
	line, column := r.position()
	return NewMalformedJsonException(fmt.Sprintf("%s at line %d column %d path %s%s", message, line, column, path, gsonMalformedJSONHelp))
}

func (r *gsonJSONReader) endOfInput(path string) error {
	line, column := r.position()
	return NewEOFException(fmt.Sprintf("End of input at line %d column %d path %s", line, column, path))
}

func (r *gsonJSONReader) position() (int, int) {
	line, lineStart := 1, 0
	if r.initialBOM {
		lineStart = 1
	}
	for i := 0; i < r.pos && i < len(r.input); i++ {
		if r.input[i] == '\n' {
			line++
			lineStart = i + 1
		}
	}
	return line, r.pos - lineStart + 1
}

func (r *gsonJSONReader) consumeTrailingPeek() {
	if r.eof() {
		return
	}
	remaining := javaStringFromUTF16(r.input[r.pos:])
	if isGsonJSONNumber(remaining) {
		r.pos = len(r.input)
		return
	}
	r.pos++
}

func (r *gsonJSONReader) consumeNonExecutePrefix() {
	prefix := []uint16{')', ']', '}', '\'', '\n'}
	if len(r.input)-r.pos < len(prefix) {
		return
	}
	for i := range prefix {
		if r.input[r.pos+i] != prefix[i] {
			return
		}
	}
	r.pos += len(prefix)
}

func (r *gsonJSONReader) skipToLineEnd() {
	for !r.eof() {
		ch := r.peek()
		r.pos++
		if ch == '\n' || ch == '\r' {
			return
		}
	}
}

func gsonArrayPath(path string, index int) string {
	return fmt.Sprintf("%s[%d]", path, index)
}

func gsonObjectPath(path, name string) string { return path + "." + name }

func (r *gsonJSONReader) eof() bool    { return r.pos >= len(r.input) }
func (r *gsonJSONReader) peek() uint16 { return r.input[r.pos] }
