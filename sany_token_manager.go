package tlago

import (
	"strings"
	"unicode/utf8"
)

type SanyTokenManager struct {
	file               string
	input              string
	offset             int
	line               int
	column             int
	state              SanyLexState
	lastEnd            Position
	diags              Diagnostics
	pendingSpecialHead *SanyToken
	pendingSpecialTail *SanyToken
}

type sanyLexCandidate struct {
	kind     SanyTokenKind
	n        int
	priority int
	diagCode string
	diagMsg  string
}

func NewSanyTokenManager(file, input string) *SanyTokenManager {
	return &SanyTokenManager{
		file:   file,
		input:  input,
		line:   1,
		column: 1,
		state:  SanyLexDefault,
	}
}

func (tm *SanyTokenManager) State() SanyLexState {
	if tm == nil {
		return SanyLexDefault
	}
	return tm.state
}

func (tm *SanyTokenManager) SwitchTo(state SanyLexState) {
	if tm == nil {
		return
	}
	tm.state = state
}

func SanyTokenize(file, input string) ([]*SanyToken, Diagnostics) {
	return NewSanyTokenManager(file, input).LexAll()
}

func (tm *SanyTokenManager) LexAll() (tokens []*SanyToken, diags Diagnostics) {
	defer func() {
		if failure := recover(); failure != nil {
			if lexical, ok := failure.(*sanyTokenMgrError); ok {
				diags = append(tm.diags, lexical.diagnostic)
			} else {
				panic(failure)
			}
		}
	}()
	var previous *SanyToken
	for {
		tok := tm.NextToken()
		if previous != nil {
			previous.Next = tok
		}
		tokens = append(tokens, tok)
		previous = tok
		if tok.Kind == SanyTokenEOF {
			break
		}
	}
	return tokens, tm.diags
}

func (tm *SanyTokenManager) NextToken() *SanyToken {
	for {
		switch tm.state {
		case SanyLexDefault:
			return tm.nextDefaultToken()
		case SanyLexPragma:
			return tm.nextPragmaToken()
		case SanyLexInEOLComment:
			tm.consumeStartedLineComment()
		case SanyLexInComment, SanyLexEmbedded:
			tm.consumeStartedBlockComment()
		default:
			return tm.nextSpecToken()
		}
	}
}

func (tm *SanyTokenManager) nextDefaultToken() *SanyToken {
	for !tm.eof() {
		if strings.HasPrefix(tm.rest(), "--->") {
			return tm.consumeToken(SanyTokenBeginPragma, len("--->"), SanyLexPragma)
		}
		if n, ok := matchSanyBeginModule(tm.rest()); ok {
			return tm.consumeToken(SanyTokenBm1, n, SanyLexSpec)
		}
		tm.advance()
	}
	return tm.eofToken()
}

func (tm *SanyTokenManager) nextPragmaToken() *SanyToken {
	for !tm.eof() {
		if n, ok := matchSanyBeginModule(tm.rest()); ok {
			return tm.consumeToken(SanyTokenBm2, n, SanyLexSpec)
		}
		if isHorizontalOrVerticalWhitespace(tm.peek()) {
			tm.advance()
			continue
		}
		if n := scanSanyDecimalNumber(tm.rest()); n > 0 {
			return tm.consumeToken(SanyTokenNumber, n, SanyLexPragma)
		}
		if ident := tm.identifierCandidate(); ident.n > 0 {
			return tm.consumeToken(SanyTokenIdentifier, ident.n, SanyLexPragma)
		}
		tm.advance()
	}
	return tm.eofToken()
}

func (tm *SanyTokenManager) nextSpecToken() *SanyToken {
	for tm.skipSpecWhitespaceOrSpecial() {
	}
	if tm.eof() {
		return tm.eofToken()
	}
	if n, ok := matchSanyBeginModule(tm.rest()); ok {
		return tm.consumeToken(SanyTokenBm0, n, SanyLexSpec)
	}
	if n := matchSanyRepeated(tm.rest(), "====", '='); n > 0 {
		return tm.consumeToken(SanyTokenEndModule, n, SanyLexSpec)
	}
	if n := matchSanyRepeated(tm.rest(), "----", '-'); n > 0 {
		return tm.consumeToken(SanyTokenSeparator, n, SanyLexSpec)
	}

	candidate := tm.bestSpecCandidate()
	if candidate.n > 0 {
		if candidate.diagCode != "" {
			begin, start := tm.pos(), tm.offset
			tm.consumeBytes(candidate.n)
			position, after := tm.pos(), tm.input[start:tm.offset]
			character := tm.peek()
			if candidate.diagCode == "E1203" {
				position = tm.lastEnd
				_, size := utf8.DecodeLastRuneInString(after)
				character, _ = utf8.DecodeLastRuneInString(after)
				if character > 0xffff {
					position.Column--
				}
				if !tm.eof() || character > 0xffff {
					after = after[:len(after)-size]
				}
			}
			eof := tm.eof() && character <= 0xffff
			if eof {
				position = tm.lexicalEOFPosition()
			}
			tm.lexicalFailure(begin, candidate.diagCode, candidate.diagMsg, position, after, character, eof)
		}
		return tm.consumeToken(candidate.kind, candidate.n, SanyLexSpec)
	}

	begin := tm.pos()
	character := tm.advance()
	position := begin
	// A supplementary rune still has a second Java code unit, even when
	// its UTF-8 bytes end the native input.
	eof := tm.eof() && character <= 0xffff
	if eof {
		position = tm.lexicalEOFPosition()
	}
	tm.lexicalFailure(begin, "E1200", "unexpected character in SANY token stream", position, "", character, eof)
	return nil
}

func (tm *SanyTokenManager) skipSpecWhitespaceOrSpecial() bool {
	if tm.eof() {
		return false
	}
	if isHorizontalOrVerticalWhitespace(tm.peek()) {
		tm.advance()
		return true
	}
	if strings.HasPrefix(tm.rest(), "\\*") {
		tm.consumeLineSpecial()
		return true
	}
	if strings.HasPrefix(tm.rest(), "(*") {
		tm.consumeBlockSpecial()
		return true
	}
	return false
}

func (tm *SanyTokenManager) consumeLineSpecial() {
	begin := tm.pos()
	start := tm.offset
	lexState := SanyLexInEOLComment
	tm.state = lexState
	for !tm.eof() {
		r := tm.peek()
		tm.advance()
		if r == '\n' || r == '\r' {
			break
		}
	}
	tm.state = SanyLexSpec
	tm.appendSpecial(tm.emitDetachedToken(SanyTokenEOLComment, begin, tm.lastEnd, tm.input[start:tm.offset], lexState))
}

func (tm *SanyTokenManager) consumeStartedLineComment() {
	for !tm.eof() {
		r := tm.peek()
		tm.advance()
		if r == '\n' || r == '\r' {
			break
		}
	}
	tm.state = SanyLexSpec
}

func (tm *SanyTokenManager) consumeBlockSpecial() {
	begin := tm.pos()
	start := tm.offset
	lexState := SanyLexInComment
	tm.state = lexState
	depth := 0
	for !tm.eof() {
		if strings.HasPrefix(tm.rest(), "(*") {
			depth++
			tm.consumeBytes(2)
			if depth > 1 {
				tm.state = SanyLexEmbedded
			}
			continue
		}
		if strings.HasPrefix(tm.rest(), "*)") {
			tm.consumeBytes(2)
			depth--
			if depth == 0 {
				tm.state = SanyLexSpec
				tm.appendSpecial(tm.emitDetachedToken(SanyTokenBlockComment, begin, tm.lastEnd, tm.input[start:tm.offset], lexState))
				return
			}
			continue
		}
		tm.advance()
	}
	tm.state = SanyLexSpec
	tm.lexicalFailure(begin, "E1201", "unterminated block comment", tm.lexicalEOFPosition(), "", 0, true)
}

func (tm *SanyTokenManager) consumeStartedBlockComment() {
	depth := 1
	for !tm.eof() {
		if strings.HasPrefix(tm.rest(), "(*") {
			depth++
			tm.consumeBytes(2)
			tm.state = SanyLexEmbedded
			continue
		}
		if strings.HasPrefix(tm.rest(), "*)") {
			tm.consumeBytes(2)
			depth--
			if depth == 0 {
				break
			}
			continue
		}
		tm.advance()
	}
	tm.state = SanyLexSpec
}

func (tm *SanyTokenManager) bestSpecCandidate() sanyLexCandidate {
	best := sanyLexCandidate{}
	for _, candidate := range []sanyLexCandidate{
		tm.proofCandidate(),
		tm.junctionCandidate(),
		tm.literalCandidate(),
		tm.stringCandidate(),
		tm.numberLiteralCandidate(),
		tm.identifierCandidate(),
	} {
		if candidate.n == 0 {
			continue
		}
		if candidate.n > best.n || (candidate.n == best.n && candidate.priority < best.priority) {
			best = candidate
		}
	}
	return best
}

func (tm *SanyTokenManager) literalCandidate() sanyLexCandidate {
	rest := tm.rest()
	for _, lit := range SanyLiteralTokens {
		if strings.HasPrefix(rest, lit.Literal) {
			return sanyLexCandidate{kind: lit.Kind, n: len(lit.Literal), priority: 20}
		}
	}
	return sanyLexCandidate{}
}

func (tm *SanyTokenManager) numberLiteralCandidate() sanyLexCandidate {
	rest := tm.rest()
	if len(rest) >= 3 && rest[0] == '\\' {
		switch rest[1] {
		case 'o', 'O':
			if n := scanSanyRadixDigits(rest[2:], isSanyOctalDigit); n > 0 {
				return sanyLexCandidate{kind: SanyTokenNumberLiteral, n: 2 + n, priority: 40}
			}
		case 'b', 'B':
			if n := scanSanyRadixDigits(rest[2:], isSanyBinaryDigit); n > 0 {
				return sanyLexCandidate{kind: SanyTokenNumberLiteral, n: 2 + n, priority: 40}
			}
		case 'h', 'H':
			if n := scanSanyRadixDigits(rest[2:], isSanyHexDigit); n > 0 {
				return sanyLexCandidate{kind: SanyTokenNumberLiteral, n: 2 + n, priority: 40}
			}
		}
	}
	if n := scanSanyDecimalNumber(rest); n > 0 {
		return sanyLexCandidate{kind: SanyTokenNumberLiteral, n: n, priority: 40}
	}
	return sanyLexCandidate{}
}

func (tm *SanyTokenManager) stringCandidate() sanyLexCandidate {
	rest := tm.rest()
	if !strings.HasPrefix(rest, "\"") {
		return sanyLexCandidate{}
	}
	for i := 1; i < len(rest); {
		r, size := utf8.DecodeRuneInString(rest[i:])
		if r == utf8.RuneError && size == 0 {
			break
		}
		if r == '"' {
			return sanyLexCandidate{kind: SanyTokenStringLiteral, n: i + size, priority: 30}
		}
		if r == '\n' || r == '\r' {
			return sanyLexCandidate{kind: SanyTokenStringLiteral, n: i, priority: 30, diagCode: "E1202", diagMsg: "unterminated string literal"}
		}
		if r == '\\' {
			i += size
			if i >= len(rest) {
				return sanyLexCandidate{kind: SanyTokenStringLiteral, n: len(rest), priority: 30, diagCode: "E1202", diagMsg: "unterminated string literal"}
			}
			esc, escSize := utf8.DecodeRuneInString(rest[i:])
			if !isSanyStringEscape(esc) {
				return sanyLexCandidate{kind: SanyTokenStringLiteral, n: i + escSize, priority: 30, diagCode: "E1203", diagMsg: "invalid string escape"}
			}
			i += escSize
			continue
		}
		i += size
	}
	return sanyLexCandidate{kind: SanyTokenStringLiteral, n: len(rest), priority: 30, diagCode: "E1202", diagMsg: "unterminated string literal"}
}

func isSanyStringEscape(r rune) bool {
	switch r {
	case 'n', 't', 'r', 'f', '\\', '"':
		return true
	default:
		return false
	}
}

func (tm *SanyTokenManager) identifierCandidate() sanyLexCandidate {
	rest := tm.rest()
	if rest == "" {
		return sanyLexCandidate{}
	}
	if strings.HasPrefix(rest, "@") {
		return sanyLexCandidate{kind: SanyTokenIdentifier, n: 1, priority: 50}
	}
	if strings.HasPrefix(rest, "ℕ") || strings.HasPrefix(rest, "ℤ") || strings.HasPrefix(rest, "ℝ") {
		_, size := utf8.DecodeRuneInString(rest)
		return sanyLexCandidate{kind: SanyTokenIdentifier, n: size, priority: 50}
	}
	i := 0
	hasLetter := false
	for i < len(rest) {
		r, size := utf8.DecodeRuneInString(rest[i:])
		if !isSanyIdentifierPart(r) {
			break
		}
		if isSanyLetter(r) {
			hasLetter = true
		}
		i += size
	}
	if i == 0 || !hasLetter {
		return sanyLexCandidate{}
	}
	return sanyLexCandidate{kind: SanyTokenIdentifier, n: i, priority: 50}
}

func (tm *SanyTokenManager) junctionCandidate() sanyLexCandidate {
	rest := tm.rest()
	i := 0
	for i < len(rest) {
		r, size := utf8.DecodeRuneInString(rest[i:])
		if !isSanyIdentifierPart(r) {
			break
		}
		i += size
	}
	if i == 0 {
		return sanyLexCandidate{}
	}
	switch {
	case strings.HasPrefix(rest[i:], `./\`):
		return sanyLexCandidate{kind: SanyTokenBand, n: i + len(`./\`), priority: 10}
	case strings.HasPrefix(rest[i:], `.\/`):
		return sanyLexCandidate{kind: SanyTokenBor, n: i + len(`.\/`), priority: 10}
	default:
		return sanyLexCandidate{}
	}
}

func (tm *SanyTokenManager) proofCandidate() sanyLexCandidate {
	rest := tm.rest()
	if !strings.HasPrefix(rest, "<") {
		return sanyLexCandidate{}
	}
	i := 1
	implicit := false
	if i < len(rest) && (rest[i] == '+' || rest[i] == '*') {
		i++
		implicit = true
	} else {
		start := i
		for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
			i++
		}
		if i == start {
			return sanyLexCandidate{}
		}
	}
	if i >= len(rest) || rest[i] != '>' {
		return sanyLexCandidate{}
	}
	i++
	nameStart := i
	for i < len(rest) {
		r, size := utf8.DecodeRuneInString(rest[i:])
		if !(isSanyLetter(r) || isSanyDigit(r) || r == '_') {
			break
		}
		i += size
	}
	if i > nameStart {
		dotStart := i
		for i < len(rest) && rest[i] == '.' {
			i++
		}
		if i > dotStart {
			return sanyLexCandidate{kind: SanyTokenProofstepdotlexeme, n: i, priority: 5}
		}
		if implicit {
			return sanyLexCandidate{kind: SanyTokenProofimplicitsteplexeme, n: i, priority: 5}
		}
		return sanyLexCandidate{kind: SanyTokenProofsteplexeme, n: i, priority: 5}
	}
	if i < len(rest) && rest[i] == '*' {
		i++
		for i < len(rest) && rest[i] == '.' {
			i++
		}
		return sanyLexCandidate{kind: SanyTokenUnnumberedsteplexeme, n: i, priority: 5}
	}
	if i < len(rest) && rest[i] == '-' {
		i++
		for i < len(rest) && rest[i] == '.' {
			i++
		}
		return sanyLexCandidate{kind: SanyTokenUnnumberedsteplexeme, n: i, priority: 5}
	}
	dotStart := i
	for i < len(rest) && rest[i] == '.' {
		i++
	}
	if i > dotStart {
		return sanyLexCandidate{kind: SanyTokenUnnumberedsteplexeme, n: i, priority: 5}
	}
	return sanyLexCandidate{kind: SanyTokenBarelevellexeme, n: i, priority: 5}
}

func (tm *SanyTokenManager) consumeToken(kind SanyTokenKind, n int, nextState SanyLexState) *SanyToken {
	begin := tm.pos()
	start := tm.offset
	lexState := tm.state
	tm.consumeBytes(n)
	tm.state = nextState
	return tm.emitToken(kind, begin, tm.lastEnd, tm.input[start:tm.offset], lexState)
}

func (tm *SanyTokenManager) emitToken(kind SanyTokenKind, begin, end Position, image string, lexState SanyLexState) *SanyToken {
	tok := tm.emitDetachedToken(kind, begin, end, image, lexState)
	if tm.pendingSpecialHead != nil {
		tok.Special = tm.pendingSpecialHead
		tm.pendingSpecialHead = nil
		tm.pendingSpecialTail = nil
	}
	return tok
}

func (tm *SanyTokenManager) emitDetachedToken(kind SanyTokenKind, begin, end Position, image string, lexState SanyLexState) *SanyToken {
	return &SanyToken{Kind: kind, Image: image, Begin: begin, End: end, LexState: lexState}
}

func (tm *SanyTokenManager) eofToken() *SanyToken {
	// BeginToken's failed read leaves SimpleCharStream at the last character.
	// Its position arrays remain zero for an empty input stream.
	pos := tm.lastEnd
	pos.File = tm.file
	return tm.emitToken(SanyTokenEOF, pos, pos, "", tm.state)
}

func (tm *SanyTokenManager) appendSpecial(tok *SanyToken) {
	if tok == nil {
		return
	}
	if tm.pendingSpecialTail == nil {
		tm.pendingSpecialHead = tok
		tm.pendingSpecialTail = tok
		return
	}
	tm.pendingSpecialTail.Next = tok
	tm.pendingSpecialTail = tok
}

func (tm *SanyTokenManager) consumeBytes(n int) {
	target := tm.offset + n
	for tm.offset < target && !tm.eof() {
		tm.advance()
	}
}

func (tm *SanyTokenManager) advance() rune {
	if tm.eof() {
		return 0
	}
	begin := tm.pos()
	r, size := utf8.DecodeRuneInString(tm.rest())
	if r == utf8.RuneError && size == 0 {
		return 0
	}
	tm.offset += size
	tm.lastEnd = begin
	if r == '\r' {
		if strings.HasPrefix(tm.rest(), "\n") {
			tm.offset++
			// Java consumes CR and LF separately, on the same physical line.
			tm.lastEnd.Column++
		}
		tm.line++
		tm.column = 1
	} else if r == '\n' {
		tm.line++
		tm.column = 1
	} else if r == '\t' {
		tm.column = nextSanyTabColumn(tm.column)
		tm.lastEnd.Column = tm.column - 1
	} else {
		tm.column++
		if r > 0xffff {
			// SimpleCharStream reads Java UTF-16 code units.
			tm.column++
			tm.lastEnd.Column++
		}
	}
	return r
}

func (tm *SanyTokenManager) pos() Position {
	return Position{File: tm.file, Line: tm.line, Column: tm.column}
}

func nextSanyTabColumn(column int) int {
	if column <= 0 {
		return 1
	}
	return column + (8 - ((column - 1) % 8))
}

func (tm *SanyTokenManager) eof() bool {
	return tm.offset >= len(tm.input)
}

func (tm *SanyTokenManager) rest() string {
	return tm.input[tm.offset:]
}

func (tm *SanyTokenManager) peek() rune {
	if tm.eof() {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(tm.rest())
	return r
}

func matchSanyBeginModule(rest string) (int, bool) {
	if !strings.HasPrefix(rest, "----") {
		return 0, false
	}
	i := len("----")
	for i < len(rest) && rest[i] == '-' {
		i++
	}
	for i < len(rest) && rest[i] == ' ' {
		i++
	}
	if !strings.HasPrefix(rest[i:], "MODULE") {
		return 0, false
	}
	return i + len("MODULE"), true
}

func matchSanyRepeated(rest, prefix string, repeated byte) int {
	if !strings.HasPrefix(rest, prefix) {
		return 0
	}
	i := len(prefix)
	for i < len(rest) && rest[i] == repeated {
		i++
	}
	return i
}

func scanSanyDecimalNumber(rest string) int {
	i := 0
	for i < len(rest) && rest[i] >= '0' && rest[i] <= '9' {
		i++
	}
	return i
}

func scanSanyRadixDigits(rest string, ok func(byte) bool) int {
	i := 0
	for i < len(rest) && ok(rest[i]) {
		i++
	}
	return i
}

func isSanyBinaryDigit(ch byte) bool {
	return ch == '0' || ch == '1'
}

func isSanyOctalDigit(ch byte) bool {
	return ch >= '0' && ch <= '7'
}

func isSanyHexDigit(ch byte) bool {
	return (ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func isHorizontalOrVerticalWhitespace(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r'
}

func isSanyIdentifierPart(r rune) bool {
	return isSanyLetter(r) || isSanyDigit(r) || r == '_'
}

func isSanyLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func isSanyDigit(r rune) bool {
	return r >= '0' && r <= '9'
}
