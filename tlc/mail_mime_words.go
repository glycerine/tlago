/*
 * Copyright (c) 1997, 2018 Oracle and/or its affiliates. All rights reserved.
 *
 * This program and the accompanying materials are made available under the
 * terms of the Eclipse Public License v. 2.0, which is available at
 * http://www.eclipse.org/legal/epl-2.0.
 *
 * This Source Code may also be made available under the following Secondary
 * Licenses when the conditions for such availability set forth in the
 * Eclipse Public License v. 2.0 are satisfied: GNU General Public License,
 * version 2 with the GNU Classpath Exception, which is available at
 * https://www.gnu.org/software/classpath/license.html.
 *
 * SPDX-License-Identifier: EPL-2.0 OR GPL-2.0 WITH Classpath-exception-2.0
 */

// Go port of JavaMail 1.6.8 MimeUtility encoded-word methods and their
// QEncoderStream/QDecoderStream/BASE64DecoderStream byte-array dependencies.
// Source: https://github.com/eclipse-ee4j/mail/blob/1.6.8/mail/src/main/java/javax/mail/internet/MimeUtility.java

package tlc

import (
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
)

// MailMIMECodec separates JavaMail's algorithms from the JDK charset provider.
// Nil callbacks use the native provider. Callbacks must retain String.getBytes
// and String(byte[]) replacement behavior and checked charset failures.
type MailMIMECodec struct {
	EncodeCharset   func(string, string) ([]byte, error)
	DecodeCharset   func([]byte, string) (string, error)
	SupportsCharset func(string) bool
}

// DefaultMailMIMECodec is the process JDK charset boundary used by address
// objects and MIME convenience functions. Configure providers before use.
// Its native conversions cover UTF-8, ASCII, Latin-1, UTF-16 and GB18030;
// other known codecs remain explicit boundaries until their Java port.
var DefaultMailMIMECodec MailMIMECodec

func EncodeMailText(value string, charset, encoding *string) (string, error) {
	return DefaultMailMIMECodec.EncodeText(value, charset, encoding)
}
func EncodeMailWord(value string, charset, encoding *string) (string, error) {
	return DefaultMailMIMECodec.EncodeWord(value, charset, encoding)
}
func DecodeMailText(value string) (string, error) { return DefaultMailMIMECodec.DecodeText(value) }
func DecodeMailWord(value string) (string, error) { return DefaultMailMIMECodec.DecodeWord(value) }

// Locale.ENGLISH lowercasing keeps dotted-I expansion and the JDK word-boundary
// rule for final sigma. Activation separately supports a default-locale provider.
func mailCharsetLower(value string) string {
	return mailJDKEnglishLower(value)
}

// QuoteMailWord is MimeUtility.quote; unlike phrase quoting it escapes CR/LF
// and quotes empty or null words. A CRLF pair escapes only its CR.
func QuoteMailWord(value *string, specials string) string {
	initializeMailMIMEProperties()
	if value == nil || *value == "" {
		return "\"\""
	}
	chars := mailAddressUTF16(*value)
	quote := false
	for i, c := range chars {
		if c == '"' || c == '\\' || c == '\r' || c == '\n' {
			out := append([]uint16{'"'}, chars[:i]...)
			var last uint16
			for _, cc := range chars[i:] {
				if (cc == '"' || cc == '\\' || cc == '\r' || cc == '\n') && !(cc == '\n' && last == '\r') {
					out = append(out, '\\')
				}
				out = append(out, cc)
				last = cc
			}
			out = append(out, '"')
			return mailAddressUTF16String(out)
		}
		if c < 0x20 || c >= 0x7f && !mailMIMEClass.allowUTF8 || strings.ContainsRune(specials, rune(c)) {
			quote = true
		}
	}
	if quote {
		return "\"" + *value + "\""
	}
	return *value
}

func (m MailMIMECodec) EncodeText(value string, charset, encoding *string) (string, error) {
	return m.encodeWord(value, charset, encoding, false)
}
func (m MailMIMECodec) EncodeWord(value string, charset, encoding *string) (string, error) {
	return m.encodeWord(value, charset, encoding, true)
}
func (m MailMIMECodec) encodeWord(value string, charset, encoding *string, word bool) (string, error) {
	initializeMailMIMEProperties()
	ascii := checkMailASCII(value)
	if ascii == 1 {
		return value, nil
	}
	var jcharset, label string
	if charset == nil {
		jcharset = m.defaultJavaCharset()
		label = m.defaultMIMECharset()
	} else {
		label = *charset
		jcharset = m.JavaCharset(label)
	}
	transfer := ""
	if encoding != nil {
		transfer = *encoding
	} else if ascii != 3 {
		transfer = "Q"
	} else {
		transfer = "B"
	}
	b64 := mailAddressEqualsIgnoreCase(transfer, "B")
	if !b64 && !mailAddressEqualsIgnoreCase(transfer, "Q") {
		return "", NewUnsupportedEncodingException("Unknown transfer encoding: " + transfer)
	}
	var b strings.Builder
	err := m.doEncode(mailAddressUTF16(value), b64, jcharset, 75-7-len(mailAddressUTF16(label)), "=?"+label+"?"+transfer+"?", true, word, &b)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}
func (m MailMIMECodec) doEncode(value []uint16, b64 bool, charset string, avail int, prefix string, first, word bool, out *strings.Builder) error {
	encoder := m.EncodeCharset
	if encoder == nil {
		encoder = encodeMailCharset
	}
	data, err := encoder(mailAddressUTF16String(value), charset)
	if err != nil {
		return err
	}
	encoded := ""
	if b64 {
		encoded = base64.StdEncoding.EncodeToString(data)
	} else {
		encoded = encodeMailQ(data, word)
	}
	if len(encoded) > avail && len(value) > 1 {
		split := len(value) / 2
		if value[split-1] >= 0xd800 && value[split-1] <= 0xdbff {
			split--
		}
		if split > 0 {
			if err = m.doEncode(value[:split], b64, charset, avail, prefix, first, word, out); err != nil {
				return err
			}
		}
		// A pair exceeding the available space gives split=0 in Java and recurses
		// without shrinking, eventually raising StackOverflowError.
		if split == 0 {
			return NewStackOverflowError()
		}
		return m.doEncode(value[split:], b64, charset, avail, prefix, false, word, out)
	}
	if !first {
		if mailMIMEClass.foldEncodedWords {
			out.WriteString("\r\n ")
		} else {
			out.WriteByte(' ')
		}
	}
	out.WriteString(prefix)
	out.WriteString(encoded)
	out.WriteString("?=")
	return nil
}
func checkMailASCII(value string) int {
	ascii, nonascii := 0, 0
	for _, c := range mailAddressUTF16(value) {
		if c >= 0x7f || (c < 0x20 && c != '\r' && c != '\n' && c != '\t') {
			nonascii++
		} else {
			ascii++
		}
	}
	if nonascii == 0 {
		return 1
	}
	if ascii > nonascii {
		return 2
	}
	return 3
}
func encodeMailQ(data []byte, word bool) string {
	specials := "=_?"
	if word {
		specials = "=_?\"#$%&'(),.:;<>@[\\]^`{|}~"
	}
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for _, c := range data {
		switch {
		case c == ' ':
			b.WriteByte('_')
		case c < 0x20 || c >= 0x7f || strings.ContainsRune(specials, rune(c)):
			b.WriteByte('=')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&15])
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}
func (m MailMIMECodec) DecodeText(value string) (string, error) {
	initializeMailMIMEProperties()
	if !strings.Contains(value, "=?") {
		return value, nil
	}
	var result, white strings.Builder
	previousEncoded := false
	for len(value) > 0 {
		if strings.ContainsRune(" \t\r\n", rune(value[0])) {
			white.WriteByte(value[0])
			value = value[1:]
			continue
		}
		end := strings.IndexAny(value, " \t\r\n")
		if end < 0 {
			end = len(value)
		}
		token := value[:end]
		value = value[end:]
		word, err := m.DecodeWord(token)
		if err == nil {
			if !previousEncoded {
				result.WriteString(white.String())
			}
			previousEncoded = true
		} else if _, ok := err.(*MailParseException); ok {
			word = token
			if !mailMIMEClass.decodeStrict {
				var changed bool
				word, changed, err = m.decodeInnerWords(token)
				if err != nil {
					return "", err
				}
				if changed {
					if !(previousEncoded && strings.HasPrefix(token, "=?")) {
						result.WriteString(white.String())
					}
					previousEncoded = strings.HasSuffix(token, "?=")
				} else {
					result.WriteString(white.String())
					previousEncoded = false
				}
			} else {
				result.WriteString(white.String())
				previousEncoded = false
			}
		} else {
			return "", err
		}
		result.WriteString(word)
		white.Reset()
	}
	result.WriteString(white.String())
	return result.String(), nil
}
func (m MailMIMECodec) DecodeWord(value string) (string, error) {
	initializeMailMIMEProperties()
	parseFailure := func(message string) (string, error) { return "", NewMailParseException(message + value) }
	if !strings.HasPrefix(value, "=?") {
		return parseFailure("encoded word does not start with \"=?\": ")
	}
	pos := strings.IndexByte(value[2:], '?')
	if pos < 0 {
		return parseFailure("encoded word does not include charset: ")
	}
	pos += 2
	charset := value[2:pos]
	if i := strings.IndexByte(charset, '*'); i >= 0 {
		charset = charset[:i]
	}
	charset = m.JavaCharset(charset)
	start := pos + 1
	pos = strings.IndexByte(value[start:], '?')
	if pos < 0 {
		return parseFailure("encoded word does not include encoding: ")
	}
	pos += start
	transfer := value[start:pos]
	start = pos + 1
	pos = strings.Index(value[start:], "?=")
	if pos < 0 {
		return parseFailure("encoded word does not end with \"?=\": ")
	}
	pos += start
	word := value[start:pos]
	decoded := ""
	if word != "" {
		units := mailAddressUTF16(word)
		data := make([]byte, len(units))
		for i, c := range units {
			data[i] = byte(c)
		}
		var raw []byte
		var err error
		if mailAddressEqualsIgnoreCase(transfer, "B") {
			raw, err = decodeMailBase64(data)
		} else if mailAddressEqualsIgnoreCase(transfer, "Q") {
			raw, err = decodeMailQ(data)
		} else {
			return "", NewUnsupportedEncodingException("unknown encoding: " + transfer)
		}
		if err != nil {
			return "", err
		}
		if len(raw) > 0 {
			decoder := m.DecodeCharset
			if decoder == nil {
				decoder = decodeMailCharset
			}
			decoded, err = decoder(raw, charset)
			if err != nil {
				return "", err
			}
		}
	}
	if pos+2 < len(value) {
		rest := value[pos+2:]
		if !mailMIMEClass.decodeStrict {
			var err error
			rest, _, err = m.decodeInnerWords(rest)
			if err != nil {
				return "", err
			}
		}
		decoded += rest
	}
	return decoded, nil
}

// changed represents Java's String object identity test; processing even an
// invalid but delimited encoded word returns a different String object.
func (m MailMIMECodec) decodeInnerWords(value string) (string, bool, error) {
	start := 0
	var b strings.Builder
	for {
		i := strings.Index(value[start:], "=?")
		if i < 0 {
			break
		}
		i += start
		b.WriteString(value[start:i])
		end := strings.IndexByte(value[i+2:], '?')
		if end < 0 {
			break
		}
		end += i + 2
		next := strings.IndexByte(value[end+1:], '?')
		if next < 0 {
			break
		}
		end += 1 + next
		next = strings.Index(value[end+1:], "?=")
		if next < 0 {
			break
		}
		end += 1 + next
		word := value[i : end+2]
		decoded, err := m.DecodeWord(word)
		if err != nil {
			if _, ok := err.(*MailParseException); !ok {
				return "", false, err
			}
		} else {
			word = decoded
		}
		b.WriteString(word)
		start = end + 2
	}
	if start == 0 {
		return value, false, nil
	}
	b.WriteString(value[start:])
	return b.String(), true, nil
}
func decodeMailQ(data []byte) ([]byte, error) {
	out := make([]byte, 0, len(data))
	digit := func(c byte) int {
		if c >= '0' && c <= '9' {
			return int(c - '0')
		}
		if c >= 'a' && c <= 'f' {
			return int(c-'a') + 10
		}
		if c >= 'A' && c <= 'F' {
			return int(c-'A') + 10
		}
		return -1
	}
	for i := 0; i < len(data); i++ {
		c := data[i]
		if c == '_' {
			c = ' '
		} else if c == '=' {
			a, b := byte(255), byte(255)
			if i+1 < len(data) {
				a = data[i+1]
			}
			if i+2 < len(data) {
				b = data[i+2]
			}
			i += 2
			first := a
			negative := a == '-'
			if negative {
				first = b
			}
			d := digit(first)
			detail := ""
			if d < 0 {
				detail = "illegal number: " + string([]rune{rune(a), rune(b)})
			} else if !negative && digit(b) < 0 {
				detail = "illegal number"
			}
			if detail != "" {
				return nil, NewMailParseException("com.sun.mail.util.DecodingException: QDecoder: Error in QP stream " + detail)
			}
			n := d
			if negative {
				n = -n
			} else {
				n = n*16 + digit(b)
			}
			if n == -1 {
				break
			}
			c = byte(n)
		}
		out = append(out, c)
	}
	return out, nil
}
func decodeMailBase64(data []byte) ([]byte, error) {
	ignore := mailBooleanProperty("mail.mime.base64.ignoreerrors", false)
	position := 0
	getByte := func() int {
		for position < len(data) {
			c := data[position]
			position++
			if c == '=' {
				return -2
			}
			switch {
			case c >= 'A' && c <= 'Z':
				return int(c - 'A')
			case c >= 'a' && c <= 'z':
				return int(c-'a') + 26
			case c >= '0' && c <= '9':
				return int(c-'0') + 52
			case c == '+':
				return 62
			case c == '/':
				return 63
			case c == 255:
				return 0
			}
		}
		return -1
	}
	failure := func(detail string) error {
		// getByte fills 78*105 bytes at a time. The diagnostic examines only the
		// current buffer's last ten consumed bytes, including skipped bytes.
		bufferStart := 0
		if position > 0 {
			bufferStart = ((position - 1) / 8190) * 8190
		}
		count := position - bufferStart
		if count > 10 {
			count = 10
		}
		recent := ""
		if count > 0 {
			var b strings.Builder
			fmt.Fprintf(&b, ", the %d most recent characters were: \"", count)
			for _, c := range data[position-count : position] {
				switch c {
				case '\r':
					b.WriteString("\\r")
				case '\n':
					b.WriteString("\\n")
				case '\t':
					b.WriteString("\\t")
				default:
					if c >= ' ' && c < 0x7f {
						b.WriteByte(c)
					} else {
						fmt.Fprintf(&b, "\\%d", c)
					}
				}
			}
			b.WriteByte('"')
			recent = b.String()
		}
		return NewMailParseException("com.sun.mail.util.DecodingException: BASE64Decoder: Error in encoded stream: " + detail + recent)
	}
	out := make([]byte, 0, len(data))
	for {
		got, val := 0, 0
		for got < 4 {
			c := getByte()
			if c == -1 || c == -2 {
				eof := c == -1
				if eof {
					if got == 0 {
						return out, nil
					}
					if !ignore {
						return nil, failure(fmt.Sprintf("needed 4 valid base64 characters but only got %d before EOF", got))
					}
				} else {
					if got < 2 && !ignore {
						return nil, failure(fmt.Sprintf("needed at least 2 valid base64 characters, but only got %d before padding character (=)", got))
					}
					if got == 0 {
						return out, nil
					}
				}
				size := got - 1
				if size == 0 {
					size = 1
				}
				got++
				val <<= 6
				for got < 4 {
					if !eof {
						c = getByte()
						if c == -1 && !ignore {
							return nil, failure("hit EOF while looking for padding characters (=)")
						}
						if c != -1 && c != -2 && !ignore {
							return nil, failure("found valid base64 character after a padding character (=)")
						}
					}
					val <<= 6
					got++
				}
				out = append(out, byte(val>>16))
				if size == 2 {
					out = append(out, byte(val>>8))
				}
				return out, nil
			}
			val = (val << 6) | c
			got++
		}
		out = append(out, byte(val>>16), byte(val>>8), byte(val))
	}
}

var mailDefaultCharsets struct {
	sync.Mutex
	java, mime *string
}

func (m MailMIMECodec) defaultJavaCharset() string {
	mailDefaultCharsets.Lock()
	defer mailDefaultCharsets.Unlock()
	return m.defaultJavaCharsetLocked()
}
func (m MailMIMECodec) defaultJavaCharsetLocked() string {
	if mailDefaultCharsets.java == nil {
		if cs, ok := tlcLookupSystemProperty("mail.mime.charset"); ok && cs != "" {
			mailDefaultCharsets.java = javaString(m.JavaCharset(cs))
		} else {
			mailDefaultCharsets.java = javaString(tlcGetSystemProperty("file.encoding", "8859_1"))
		}
	}
	return *mailDefaultCharsets.java
}
func (m MailMIMECodec) defaultMIMECharset() string {
	mailDefaultCharsets.Lock()
	defer mailDefaultCharsets.Unlock()
	if mailDefaultCharsets.mime == nil {
		if cs, ok := tlcLookupSystemProperty("mail.mime.charset"); ok {
			mailDefaultCharsets.mime = javaString(cs)
		} else {
			mailDefaultCharsets.mime = javaString(m.MIMECharset(m.defaultJavaCharsetLocked()))
		}
	}
	return *mailDefaultCharsets.mime
}
func (m MailMIMECodec) JavaCharset(value string) string {
	initializeMailMIMEProperties()
	aliases := map[string]string{"iso-2022-cn": "ISO2022CN", "iso-2022-kr": "ISO2022KR", "utf-8": "UTF8", "utf8": "UTF8", "ja_jp.iso2022-7": "ISO2022JP", "ja_jp.eucjp": "EUCJIS", "euc-kr": "KSC5601", "euckr": "KSC5601", "us-ascii": "ISO-8859-1", "x-us-ascii": "ISO-8859-1", "gb2312": "GB18030", "cp936": "GB18030", "ms936": "GB18030", "gbk": "GB18030"}
	if alias, ok := aliases[mailCharsetLower(value)]; ok {
		supported := m.SupportsCharset
		if supported == nil {
			supported = supportsMailCharset
		}
		if supported(alias) {
			return alias
		}
	}
	return value
}
func (m MailMIMECodec) MIMECharset(value string) string {
	initializeMailMIMEProperties()
	key := mailCharsetLower(value)
	for i := 1; i <= 9; i++ {
		if key == fmt.Sprintf("8859_%d", i) || key == fmt.Sprintf("iso8859_%d", i) || key == fmt.Sprintf("iso8859-%d", i) {
			return fmt.Sprintf("ISO-8859-%d", i)
		}
	}
	aliases := map[string]string{"sjis": "Shift_JIS", "jis": "ISO-2022-JP", "iso2022jp": "ISO-2022-JP", "euc_jp": "euc-jp", "koi8_r": "koi8-r", "euc_cn": "euc-cn", "euc_tw": "euc-tw", "euc_kr": "euc-kr"}
	if alias, ok := aliases[key]; ok {
		return alias
	}
	return value
}
