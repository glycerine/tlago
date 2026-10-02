/*
 * Copyright (c) 1999, 2022, Oracle and/or its affiliates. All rights reserved.
 * Copyright (c) 2003, 2022, Oracle and/or its affiliates. All rights reserved.
 * Copyright (c) 1996, 2023, Oracle and/or its affiliates. All rights reserved.
 * Copyright (c) 2003, 2012, Oracle and/or its affiliates. All rights reserved.
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

/*
 *
 * (C) Copyright Taligent, Inc. 1996, 1997 - All Rights Reserved
 * (C) Copyright IBM Corp. 1996 - 2002 - All Rights Reserved
 *
 * The original version of this source code and documentation
 * is copyrighted and owned by Taligent, Inc., a wholly-owned
 * subsidiary of IBM. These materials are provided under terms
 * of a License Agreement between Taligent and Sun. This technology
 * is protected by multiple US and International patents.
 *
 * This notice and attribution to Taligent may not be removed.
 * Taligent is a registered trademark of Taligent, Inc.
 */

/*
 * (C) Copyright Taligent, Inc. 1996 - All Rights Reserved
 * (C) Copyright IBM Corp. 1996 - All Rights Reserved
 *
 *   The original version of this source code and documentation is copyrighted
 * and owned by Taligent, Inc., a wholly-owned subsidiary of IBM. These
 * materials are provided under terms of a License Agreement between Taligent
 * and Sun. This technology is protected by multiple US and International
 * patents. This notice and attribution to Taligent may not be removed.
 *   Taligent is a registered trademark of Taligent, Inc.
 *
 */

// English String.toLowerCase conditional mappings used by JavaMail and Activation.
// Ports ConditionalSpecialCasing.isFinalCased/isCased and the RuleBasedBreakIterator
// following/isBoundary/forward/backward paths with CompactByteArray and
// SupplementaryCharacterData lookup. Embedded rules are byte-identical OpenJDK
// 21.0.12.1 WordBreakIteratorData. Full default-locale/locale-provider lifetime is
// separate. Sources: openjdk/jdk21u java.lang.ConditionalSpecialCasing and sun.text.
package tlc

import (
	_ "embed"
	"encoding/binary"
	"strings"
	"unicode"
)

//go:embed resources/openjdk/WordBreakIteratorData
var mailJDKEnglishWordData []byte

type mailJDKWordTables struct {
	states, backwards []uint16
	ends, lookahead   []byte
	bmpIndices        []uint16
	bmp               []byte
	supplementary     []uint32
	categories        int
}

var mailJDKEnglishWordTables = mailJDKReadWordTables(mailJDKEnglishWordData)

func mailJDKReadWordTables(data []byte) mailJDKWordTables {
	pos := 12
	take32 := func() int { v := int(binary.BigEndian.Uint32(data[pos:])); pos += 4; return v }
	ns, nb, ne, nl, nBMP, nSup, nAdd := take32(), take32(), take32(), take32(), take32(), take32(), take32()
	pos += 8 // checksum
	shorts := func(n int) []uint16 {
		v := make([]uint16, n)
		for i := range v {
			v[i] = binary.BigEndian.Uint16(data[pos:])
			pos += 2
		}
		return v
	}
	t := mailJDKWordTables{states: shorts(ns), backwards: shorts(nb)}
	t.ends = data[pos : pos+ne]
	pos += ne
	t.lookahead = data[pos : pos+nl]
	pos += nl
	t.bmpIndices = shorts(512)
	t.bmp = data[pos : pos+nBMP]
	pos += nBMP
	t.supplementary = make([]uint32, nSup)
	for i := range t.supplementary {
		t.supplementary[i] = uint32(take32())
	}
	pos += nAdd
	if pos != len(data) {
		panic("invalid embedded OpenJDK word data")
	}
	t.categories = ns / ne
	return t
}
func (t *mailJDKWordTables) category(c rune) int {
	if c < 0x10000 {
		return int(int8(t.bmp[int(t.bmpIndices[int(c)>>7]&0xffff)+int(c&127)]))
	}
	i, j := 0, len(t.supplementary)-1
	for {
		k := (i + j) / 2
		start, end := t.supplementary[k]>>8, t.supplementary[k+1]>>8
		if uint32(c) < start {
			j = k
		} else if uint32(c) > end-1 {
			i = k
		} else {
			v := int(t.supplementary[k] & 255)
			if v == 255 {
				return -1
			}
			return v
		}
	}
}
func mailJDKCodePointAt(text []uint16, i int) (rune, int) {
	if i == len(text) {
		return 0xffff, 1
	}
	c := text[i]
	if c >= 0xd800 && c <= 0xdbff && i+1 < len(text) && text[i+1] >= 0xdc00 && text[i+1] <= 0xdfff {
		return 0x10000 + rune(c-0xd800)*0x400 + rune(text[i+1]-0xdc00), 2
	}
	return rune(c), 1
}
func mailJDKCodePointBefore(text []uint16, i int) (rune, int) {
	c := text[i-1]
	if c >= 0xdc00 && c <= 0xdfff && i > 1 && text[i-2] >= 0xd800 && text[i-2] <= 0xdbff {
		return 0x10000 + rune(text[i-2]-0xd800)*0x400 + rune(c-0xdc00), 2
	}
	return rune(c), 1
}

type mailJDKWordIterator struct {
	text             []uint16
	position, cached int
}

func (r *mailJDKWordIterator) current() rune {
	c, _ := mailJDKCodePointAt(r.text, r.position)
	return c
}
func (r *mailJDKWordIterator) nextIndex() int {
	_, n := mailJDKCodePointAt(r.text, r.position)
	return min(r.position+n, len(r.text))
}
func (r *mailJDKWordIterator) next() rune {
	next := r.nextIndex()
	if r.position == len(r.text) || next >= len(r.text) {
		return 0xffff
	}
	r.position = next
	return r.current()
}
func (r *mailJDKWordIterator) previous() rune {
	if r.position == 0 {
		return 0xffff
	}
	c, n := mailJDKCodePointBefore(r.text, r.position)
	r.position -= n
	return c
}
func (r *mailJDKWordIterator) handleNext() int {
	if r.position == len(r.text) {
		return -1
	}
	result, lookahead := r.nextIndex(), 0
	state := 1
	c := r.current()
	t := &mailJDKEnglishWordTables
	for c != 0xffff && state != 0 {
		cat := t.category(c)
		if cat != -1 {
			state = int(t.states[state*t.categories+cat])
		}
		if t.lookahead[state] == 1 {
			if t.ends[state] == 1 {
				result = lookahead
			} else {
				lookahead = r.nextIndex()
			}
		} else if t.ends[state] == 1 {
			result = r.nextIndex()
		}
		c = r.next()
	}
	if c == 0xffff && lookahead == len(r.text) {
		result = lookahead
	}
	r.position = result
	return result
}
func (r *mailJDKWordIterator) handlePrevious() int {
	state, category, last := 1, 0, 0
	c := r.current()
	t := &mailJDKEnglishWordTables
	for c != 0xffff && state != 0 {
		last = category
		category = t.category(c)
		if category != -1 {
			state = int(t.backwards[state*t.categories+category])
		}
		c = r.previous()
	}
	if c != 0xffff {
		if last != -1 {
			r.next()
			r.next()
		} else {
			r.next()
		}
	}
	return r.position
}
func (r *mailJDKWordIterator) following(offset int) int {
	r.position = offset
	if offset == 0 {
		r.cached = r.handleNext()
		return r.cached
	}
	result := r.cached
	if result >= offset || result <= -1 {
		result = r.handlePrevious()
	} else {
		r.position = result
	}
	for result != -1 && result <= offset {
		result = r.handleNext()
	}
	r.cached = result
	return result
}
func (r *mailJDKWordIterator) isBoundary(offset int) bool {
	return offset == 0 || r.following(offset-1) == offset
}
func mailJDKIsCased(c rune) bool {
	return unicode.Is(unicode.Lu, c) || unicode.Is(unicode.Ll, c) || unicode.Is(unicode.Lt, c) || c >= 0x2b0 && c <= 0x2b8 || c >= 0x2c0 && c <= 0x2c1 || c >= 0x2e0 && c <= 0x2e4 || c == 0x345 || c == 0x37a || c >= 0x1d2c && c <= 0x1d61 || c >= 0x2160 && c <= 0x217f || c >= 0x24b6 && c <= 0x24e9
}
func mailJDKFinalCased(text []uint16, index int) bool {
	words := mailJDKWordIterator{text: text, cached: -1}
	for i := index; i >= 0 && !words.isBoundary(i); {
		c, n := mailJDKCodePointBefore(text, i)
		if mailJDKIsCased(c) {
			_, n = mailJDKCodePointAt(text, index)
			for i = index + n; i < len(text) && !words.isBoundary(i); {
				c, n = mailJDKCodePointAt(text, i)
				if mailJDKIsCased(c) {
					return false
				}
				i += n
			}
			return true
		}
		i -= n
	}
	return false
}
func mailJDKEnglishLower(value string) string {
	ascii := true
	for i := 0; i < len(value); i++ {
		if value[i] >= 128 {
			ascii = false
			break
		}
	}
	if ascii {
		return strings.ToLower(value)
	}
	text := mailAddressUTF16(value)
	out := make([]uint16, 0, len(text))
	for i := 0; i < len(text); {
		c, n := mailJDKCodePointAt(text, i)
		switch c {
		case 0x130:
			out = append(out, 'i', 0x307)
		case 0x3a3:
			if mailJDKFinalCased(text, i) {
				out = append(out, 0x3c2)
			} else {
				out = append(out, 0x3c3)
			}
		default:
			c = unicode.ToLower(c)
			if c > 0xffff {
				c -= 0x10000
				out = append(out, uint16(0xd800+(c>>10)), uint16(0xdc00+(c&0x3ff)))
			} else {
				out = append(out, uint16(c))
			}
		}
		i += n
	}
	return mailAddressUTF16String(out)
}
