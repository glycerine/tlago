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

// Native charset adapters for the JavaMail String.getBytes/String(byte[]) boundary.
// JVM-specific charset inventories and custom providers remain separate.

package tlc

import (
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/ianaindex"
	"strings"
	"unicode/utf8"
)

func mailCharsetName(value string) string {
	if !isMailJavaCharsetName(value) {
		return value
	}
	key := strings.ToLower(value)
	switch key {
	case "utf8", "utf-8", "unicode-1-1-utf-8":
		return "utf-8"
	case "ascii", "us-ascii", "ansi_x3.4-1968", "iso646-us", "default":
		return "us-ascii"
	case "8859_1", "iso8859_1", "iso8859-1", "iso-8859-1", "latin1", "l1", "cp819", "ibm819":
		return "iso-8859-1"
	case "unicode", "utf-16", "utf16":
		return "utf-16"
	case "unicodebigunmarked", "utf-16be", "utf16be":
		return "utf-16be"
	case "unicodelittleunmarked", "utf-16le", "utf16le":
		return "utf-16le"
	case "gb18030":
		return "gb18030"
	case "sjis":
		return "Shift_JIS"
	case "jis", "iso2022jp":
		return "ISO-2022-JP"
	case "eucjis", "euc_jp":
		return "EUC-JP"
	case "ksc5601", "euc_kr":
		return "EUC-KR"
	case "koi8_r":
		return "KOI8-R"
	}
	if strings.HasPrefix(key, "8859_") {
		return "ISO-8859-" + key[5:]
	}
	if strings.HasPrefix(key, "iso8859_") || strings.HasPrefix(key, "iso8859-") {
		return "ISO-8859-" + key[8:]
	}
	return value
}
func mailCharsetEncoding(value string) encoding.Encoding {
	if !isMailJavaCharsetName(value) {
		return nil
	}
	e, err := ianaindex.MIME.Encoding(mailCharsetName(value))
	if err != nil {
		return nil
	}
	return e
}

func isMailJavaCharsetName(value string) bool {
	if value == "" {
		return false
	}
	for i, c := range []byte(value) {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
			continue
		}
		if i > 0 && strings.ContainsRune("-+_.:", rune(c)) {
			continue
		}
		return false
	}
	return true
}
func supportsMailCharset(value string) bool {
	if mailCharsetName(value) == "gb18030" {
		initializeMailGB18030()
	}
	switch mailCharsetName(value) {
	case "utf-8", "us-ascii", "iso-8859-1", "utf-16", "utf-16be", "utf-16le":
		return true
	}
	return mailCharsetEncoding(value) != nil
}
func encodeMailCharset(value, charset string) ([]byte, error) {
	name := mailCharsetName(charset)
	units := mailAddressUTF16(value)
	switch name {
	case "gb18030":
		return encodeMailGB18030(value), nil
	case "utf-8":
		return []byte(mailPrintUTF8String(mailAddressUTF16String(units))), nil
	case "us-ascii", "iso-8859-1":
		limit := uint16(127)
		if name == "iso-8859-1" {
			limit = 255
		}
		data := make([]byte, 0, len(units))
		for i := 0; i < len(units); i++ {
			c := units[i]
			if c > limit {
				data = append(data, '?')
				if c >= 0xd800 && c <= 0xdbff && i+1 < len(units) && units[i+1] >= 0xdc00 && units[i+1] <= 0xdfff {
					i++
				}
			} else {
				data = append(data, byte(c))
			}
		}
		return data, nil
	case "utf-16", "utf-16be", "utf-16le":
		data := make([]byte, 0, 2*len(units)+2)
		if name == "utf-16" && len(units) > 0 {
			data = append(data, 0xfe, 0xff)
		}
		for i := 0; i < len(units); i++ {
			c := units[i]
			if c >= 0xd800 && c <= 0xdbff {
				if i+1 < len(units) && units[i+1] >= 0xdc00 && units[i+1] <= 0xdfff {
					if name == "utf-16le" {
						data = append(data, byte(c), byte(c>>8))
					} else {
						data = append(data, byte(c>>8), byte(c))
					}
					i++
					c = units[i]
				} else {
					c = 0xfffd
				}
			} else if c >= 0xdc00 && c <= 0xdfff {
				c = 0xfffd
			}
			if name == "utf-16le" {
				data = append(data, byte(c), byte(c>>8))
			} else {
				data = append(data, byte(c>>8), byte(c))
			}
		}
		return data, nil
	}
	return nil, mailMissingCharsetProvider(charset)
}
func decodeMailCharset(data []byte, charset string) (string, error) {
	name := mailCharsetName(charset)
	switch name {
	case "utf-8":
		return decodeMailUTF8(data), nil
	case "gb18030":
		return decodeMailGB18030(data), nil
	case "us-ascii", "iso-8859-1":
		units := make([]uint16, len(data))
		for i, c := range data {
			units[i] = uint16(c)
			if name == "us-ascii" && c > 127 {
				units[i] = 0xfffd
			}
		}
		return mailAddressUTF16String(units), nil
	case "utf-16", "utf-16be", "utf-16le":
		little := name == "utf-16le"
		if name == "utf-16" && len(data) >= 2 {
			if data[0] == 0xfe && data[1] == 0xff {
				data = data[2:]
			} else if data[0] == 0xff && data[1] == 0xfe {
				data = data[2:]
				little = true
			}
		}
		read := func(b []byte) uint16 {
			if little {
				return uint16(b[1])<<8 | uint16(b[0])
			}
			return uint16(b[0])<<8 | uint16(b[1])
		}
		out := []uint16{}
		for len(data) >= 2 {
			c := read(data)
			data = data[2:]
			if c >= 0xd800 && c <= 0xdbff {
				if len(data) >= 2 {
					next := read(data)
					data = data[2:]
					if next >= 0xdc00 && next <= 0xdfff {
						out = append(out, c, next)
						continue
					}
				} else {
					data = nil
				}
				c = 0xfffd
			} else if c >= 0xdc00 && c <= 0xdfff {
				c = 0xfffd
			}
			out = append(out, c)
		}
		if len(data) > 0 {
			out = append(out, 0xfffd)
		}
		return mailAddressUTF16String(out), nil
	}
	return "", mailMissingCharsetProvider(charset)
}

// A known but unported JDK codec stays an explicit native boundary. The Go
// IANA codec registry is used for alias availability, never to silently stand
// in for JDK conversion behavior. Full JVM-specific inventories remain pending.
func mailMissingCharsetProvider(charset string) error {
	if supportsMailCharset(charset) {
		return NewUnsupportedOperationException("JDK charset provider is not configured: " + charset)
	}
	return NewUnsupportedEncodingException(charset)
}

// JDK UTF-8 replacement consumes the valid prefix of a malformed sequence;
// surrogate encodings consume all three bytes as one malformed sequence.
func decodeMailUTF8(data []byte) string {
	var out strings.Builder
	for len(data) > 0 {
		c := data[0]
		if c < 128 {
			out.WriteByte(c)
			data = data[1:]
			continue
		}
		width := 0
		switch {
		case c >= 0xc2 && c <= 0xdf:
			width = 2
		case c >= 0xe0 && c <= 0xef:
			width = 3
		case c >= 0xf0 && c <= 0xf4:
			width = 4
		}
		n := 1
		if width > 0 {
			for n < width && n < len(data) {
				next := data[n]
				if next&0xc0 != 0x80 || (n == 1 && ((c == 0xe0 && next < 0xa0) || (c == 0xf0 && next < 0x90) || (c == 0xf4 && next > 0x8f))) {
					break
				}
				n++
			}
		}
		if width == 0 || n < width {
			out.WriteRune(utf8.RuneError)
			data = data[n:]
			continue
		}
		if c == 0xed && data[1] >= 0xa0 {
			out.WriteRune(utf8.RuneError)
		} else {
			r, _ := utf8.DecodeRune(data[:width])
			out.WriteRune(r)
		}
		data = data[width:]
	}
	return out.String()
}
