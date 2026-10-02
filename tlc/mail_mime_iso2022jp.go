/*
 * Copyright (c) 2002, 2021, Oracle and/or its affiliates. All rights reserved.
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS FILE HEADER.
 *
 * This code is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License version 2 only, as
 * published by the Free Software Foundation.  Oracle designates this
 * particular file as subject to the "Classpath" exception as provided
 * by Oracle in the LICENSE file that accompanied this code.
 *
 * This code is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License
 * version 2 for more details (a copy is included in the LICENSE file that
 * accompanied this code).
 *
 * You should have received a copy of the GNU General Public License version
 * 2 along with this work; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin St, Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * Please contact Oracle, 500 Oracle Parkway, Redwood Shores, CA 94065 USA
 * or visit www.oracle.com if you need additional information or have any
 * questions.
 */

// Native Go port of OpenJDK 21 ISO2022_JP.Decoder used by JavaMail.
// Source: https://github.com/openjdk/jdk21u/blob/master/src/jdk.charsets/share/classes/sun/nio/cs/ext/ISO2022_JP.java
// The JIS_X_0208 mapping uses pinned x/text EUC-JP with the source mapping
// differences below, checked against all 65,536 OpenJDK decodeDouble entries.
// Encoding and Microsoft/JIS0212 variants remain separate provider boundaries.
package tlc

import (
	"golang.org/x/text/encoding/japanese"
	"unicode/utf8"
)

func decodeMailISO2022JP(data []byte) string {
	const (
		ascii = iota
		roman
		jis1978
		jis1983
		kana
		shiftOut
	)
	state, previous := ascii, ascii
	units := make([]uint16, 0, len(data))
	decoder := japanese.EUCJP.NewDecoder()
	for i := 0; i < len(data); {
		b := data[i]
		size := 1
		c := uint16(0xfffd)
		if b >= 0x80 {
			units = append(units, c)
			i++
			continue
		}
		switch b {
		case 0x1b:
			if len(data)-i < 3 {
				units = append(units, c)
				return mailAddressUTF16String(units)
			}
			size = 2
			valid := false
			switch data[i+1] {
			case '(':
				size = 3
				switch data[i+2] {
				case 'B':
					state = ascii
					valid = true
				case 'J':
					state = roman
					valid = true
				case 'I':
					state = kana
					valid = true
				}
			case '$':
				size = 3
				switch data[i+2] {
				case '@':
					state = jis1978
					valid = true
				case 'B':
					state = jis1983
					valid = true
				}
			}
			i += size
			if !valid {
				units = append(units, c)
			}
			continue
		case 0x0e:
			previous = state
			state = shiftOut
			i++
			continue
		case 0x0f:
			state = previous
			i++
			continue
		}
		switch state {
		case ascii:
			c = uint16(b)
		case roman:
			c = uint16(b)
			if b == 0x5c {
				c = 0x00a5
			} else if b == 0x7e {
				c = 0x203e
			}
		case jis1978, jis1983:
			if len(data)-i < 2 {
				units = append(units, c)
				return mailAddressUTF16String(units)
			}
			size = 2
			b2 := data[i+1]
			if b >= 0x21 && b <= 0x7e && b2 >= 0x21 && b2 <= 0x7e {
				key := uint16(b)<<8 | uint16(b2)
				overridden := false
				for _, r := range mailJIS0208DecodeDeltas {
					if key >= r.first && key <= r.last {
						c = r.value
						overridden = true
						break
					}
				}
				if !overridden {
					decoded, _ := decoder.Bytes([]byte{b + 0x80, b2 + 0x80})
					r, n := utf8.DecodeRune(decoded)
					if n == len(decoded) {
						c = uint16(r)
					}
				}
			}
		case kana, shiftOut:
			if b <= 0x5f {
				c = uint16(b) + 0xff40
			}
		}
		units = append(units, c)
		i += size
	}
	return mailAddressUTF16String(units)
}

type mailJISDecodeRange struct{ first, last, value uint16 }

var mailJIS0208DecodeDeltas = [...]mailJISDecodeRange{
	{0x213d, 0x213d, 0x2014},
	{0x2141, 0x2141, 0x301c},
	{0x2142, 0x2142, 0x2016},
	{0x215d, 0x215d, 0x2212},
	{0x2171, 0x2171, 0x00a2},
	{0x2172, 0x2172, 0x00a3},
	{0x224c, 0x224c, 0x00ac},
	{0x2d21, 0x2d3e, 0xfffd},
	{0x2d40, 0x2d56, 0xfffd},
	{0x2d5f, 0x2d7c, 0xfffd},
	{0x7921, 0x797e, 0xfffd},
	{0x7a21, 0x7a7e, 0xfffd},
	{0x7b21, 0x7b7e, 0xfffd},
	{0x7c21, 0x7c6e, 0xfffd},
	{0x7c71, 0x7c7e, 0xfffd},
}
