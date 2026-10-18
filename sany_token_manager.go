package tlago

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/glycerine/tlago/tlc"
)

type SanyTokenManager struct {
	file                                                string
	stream                                              *sanyCharStream
	DebugStream                                         io.Writer
	state                                               SanyLexState
	curChar                                             uint16
	jjrounds                                            [269]int32
	jjstateSet                                          [538]int32
	jjnewStateCnt, jjround, jjmatchedPos, jjmatchedKind int32
	image                                               []uint16
	jjimageLen, lengthOfMatch                           int32
	diags                                               Diagnostics
	lexicalBegin, lastEnd                               Position
}

var sanyCommentBracketCount atomic.Int32

func NewSanyTokenManager(file, input string) *SanyTokenManager {
	tm := NewSanyTokenManagerFromCharStream(NewDefaultSanyCharStream(input))
	tm.file = file
	return tm
}

func NewSanyTokenManagerWithState(file, input string, state SanyLexState) *SanyTokenManager {
	tm := NewSanyTokenManager(file, input)
	tm.SwitchTo(state)
	return tm
}

func NewSanyTokenManagerFromCharStream(stream *SanyCharStream) *SanyTokenManager {
	return &SanyTokenManager{stream: stream, DebugStream: os.Stdout}
}

func NewSanyTokenManagerFromCharStreamWithState(stream *SanyCharStream, state SanyLexState) *SanyTokenManager {
	tm := NewSanyTokenManagerFromCharStream(stream)
	tm.SwitchTo(state)
	return tm
}

func (tm *SanyTokenManager) require() {
	if tm == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (tm *SanyTokenManager) SetDebugStream(stream io.Writer) {
	tm.require()
	tm.DebugStream = stream
}

func (tm *SanyTokenManager) ReInit(file, input string, state ...SanyLexState) {
	tm.require()
	tm.file = file
	tm.ReInitCharStream(NewDefaultSanyCharStream(input), state...)
}

func (tm *SanyTokenManager) ReInitCharStream(stream *SanyCharStream, state ...SanyLexState) {
	tm.require()
	tm.reInit(stream, state...)
}
func (tm *SanyTokenManager) State() SanyLexState {
	if tm == nil {
		return SanyLexDefault
	}
	return tm.state
}
func (tm *SanyTokenManager) SwitchTo(state SanyLexState) {
	tm.require()
	if state < 0 || state >= 6 {
		message := fmt.Sprintf("Error: Ignoring invalid lexical state : %d. State unchanged.", state)
		diagnostic := errorAt(tm.beginPosition(), "E1204", "%s", message)
		diagnostic.SANYParseMessage = message
		failure := NewSanyTokenMgrErrorMessage(message, SanyTokenMgrInvalidLexicalState)
		failure.diagnostic = diagnostic
		panic(failure)
	}
	tm.state = state
}
func (tm *SanyTokenManager) beginPosition() Position {
	return Position{File: tm.file, Line: int(tm.stream.getBeginLine()), Column: int(tm.stream.getBeginColumn())}
}
func (tm *SanyTokenManager) endPosition() Position {
	return Position{File: tm.file, Line: int(tm.stream.getEndLine()), Column: int(tm.stream.getEndColumn())}
}
func (tm *SanyTokenManager) readScanChar() bool {
	c, err := tm.stream.readChar()
	if err != nil {
		return false
	}
	tm.curChar = c
	return true
}
func (tm *SanyTokenManager) beginScanToken() bool {
	c, err := tm.stream.beginToken()
	if err != nil {
		return false
	}
	tm.curChar = c
	return true
}
func (tm *SanyTokenManager) reInitRounds() {
	tm.jjround = -2147483647
	for i := len(tm.jjrounds) - 1; i >= 0; i-- {
		tm.jjrounds[i] = -2147483648
	}
}
func (tm *SanyTokenManager) reInit(stream *sanyCharStream, state ...SanyLexState) {
	tm.jjmatchedPos = 0
	tm.jjnewStateCnt = 0
	tm.state = SanyLexDefault
	tm.stream = stream
	tm.reInitRounds()
	if len(state) != 0 {
		tm.SwitchTo(state[0])
	}
}
func (tm *SanyTokenManager) jjCheckNAdd(state int32) {
	if tm.jjrounds[state] != tm.jjround {
		tm.jjstateSet[tm.jjnewStateCnt] = state
		tm.jjnewStateCnt++
		tm.jjrounds[state] = tm.jjround
	}
}
func (tm *SanyTokenManager) jjAddStates(start, end int32) {
	for {
		tm.jjstateSet[tm.jjnewStateCnt] = sanyScannerNextStates[start]
		tm.jjnewStateCnt++
		old := start
		start++
		if old == end {
			break
		}
	}
}
func (tm *SanyTokenManager) jjCheckNAddTwoStates(state1, state2 int32) {
	tm.jjCheckNAdd(state1)
	tm.jjCheckNAdd(state2)
}
func (tm *SanyTokenManager) jjCheckNAddStates(start int32, ends ...int32) {
	if len(ends) == 0 {
		tm.jjCheckNAdd(sanyScannerNextStates[start])
		tm.jjCheckNAdd(sanyScannerNextStates[start+1])
		return
	}
	end := ends[0]
	for {
		tm.jjCheckNAdd(sanyScannerNextStates[start])
		old := start
		start++
		if old == end {
			break
		}
	}
}
func (tm *SanyTokenManager) fillToken() *SanyToken {
	image := tm.stream.getImage()
	if literal := SanyScannerLiteralImages[tm.jjmatchedKind]; literal != nil {
		image = *literal
	}
	begin, end := tm.beginPosition(), tm.endPosition()
	return &SanyToken{Kind: SanyTokenKind(tm.jjmatchedKind), Image: image,
		BeginLine: begin.Line, BeginColumn: begin.Column, EndLine: end.Line, EndColumn: end.Column,
		SourceFile: begin.File, LexState: tm.state, imageSet: true}
}
func (tm *SanyTokenManager) skipLexicalActions() {
	switch tm.jjmatchedKind {
	case 29, 31, 32:
		if tm.image == nil {
			tm.image = make([]uint16, 0)
		}
		tm.lengthOfMatch = tm.jjmatchedPos + 1
		tm.image = append(tm.image, tm.stream.getSuffix(tm.jjimageLen+tm.lengthOfMatch)...)
		switch tm.jjmatchedKind {
		case 29, 32:
			sanyCommentBracketCount.Add(1)
		case 31:
			if sanyCommentBracketCount.Add(-1) == 0 {
				tm.SwitchTo(SanyLexInComment)
			}
		}
	}
}

func SanyTokenize(file, input string) ([]*SanyToken, Diagnostics) {
	return NewSanyTokenManager(file, input).LexAll()
}

func (tm *SanyTokenManager) LexAll() (tokens []*SanyToken, diags Diagnostics) {
	tm.require()
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
	tm.require()
	var specialToken *SanyToken
	var curPos int32
EOFLoop:
	for {
		if !tm.beginScanToken() {
			tm.jjmatchedKind = 0
			token := tm.fillToken()
			token.SpecialToken = specialToken
			return token
		}
		tm.image = nil
		tm.jjimageLen = 0
		for {
			switch tm.state {
			case SanyLexDefault:
				tm.jjmatchedKind = 2147483647
				tm.jjmatchedPos = 0
				curPos = tm.jjMoveStringLiteralDfa0_0()
				if tm.jjmatchedPos == 0 && tm.jjmatchedKind > 4 {
					tm.jjmatchedKind = 4
				}
			case SanyLexPragma:
				tm.jjmatchedKind = 2147483647
				tm.jjmatchedPos = 0
				curPos = tm.jjMoveStringLiteralDfa0_1()
				if tm.jjmatchedPos == 0 && tm.jjmatchedKind > 22 {
					tm.jjmatchedKind = 22
				}
			case SanyLexSpec:
				tm.stream.backup(0)
				for tm.curChar <= 32 && (uint64(0x100002600)&(uint64(1)<<tm.curChar)) != 0 {
					if !tm.beginScanToken() {
						continue EOFLoop
					}
				}
				tm.jjmatchedKind = 2147483647
				tm.jjmatchedPos = 0
				curPos = tm.jjMoveStringLiteralDfa0_2()
			case SanyLexInComment:
				tm.jjmatchedKind = 2147483647
				tm.jjmatchedPos = 0
				curPos = tm.jjMoveStringLiteralDfa0_3()
				if tm.jjmatchedPos == 0 && tm.jjmatchedKind > 34 {
					tm.jjmatchedKind = 34
				}
			case SanyLexEmbedded:
				tm.jjmatchedKind = 2147483647
				tm.jjmatchedPos = 0
				curPos = tm.jjMoveStringLiteralDfa0_4()
				if tm.jjmatchedPos == 0 && tm.jjmatchedKind > 34 {
					tm.jjmatchedKind = 34
				}
			case SanyLexInEOLComment:
				tm.jjmatchedKind = 2147483647
				tm.jjmatchedPos = 0
				curPos = tm.jjMoveStringLiteralDfa0_5()
				if tm.jjmatchedPos == 0 && tm.jjmatchedKind > 34 {
					tm.jjmatchedKind = 34
				}
			}
			if tm.jjmatchedKind != 2147483647 {
				if tm.jjmatchedPos+1 < curPos {
					tm.stream.backup(curPos - tm.jjmatchedPos - 1)
				}
				bit := uint64(1) << uint32(tm.jjmatchedKind&63)
				word := tm.jjmatchedKind >> 6
				if sanyScannerToToken[word]&bit != 0 {
					token := tm.fillToken()
					token.SpecialToken = specialToken
					if next := SanyScannerNewLexState[tm.jjmatchedKind]; next != -1 {
						tm.state = SanyLexState(next)
					}
					return token
				} else if sanyScannerToSkip[word]&bit != 0 {
					if sanyScannerToSpecial[word]&bit != 0 {
						token := tm.fillToken()
						if specialToken == nil {
							specialToken = token
						} else {
							token.SpecialToken = specialToken
							specialToken.Next = token
							specialToken = token
						}
					}
					tm.skipLexicalActions()
					if next := SanyScannerNewLexState[tm.jjmatchedKind]; next != -1 {
						tm.state = SanyLexState(next)
					}
					continue EOFLoop
				}
				tm.jjimageLen += tm.jjmatchedPos + 1
				if next := SanyScannerNewLexState[tm.jjmatchedKind]; next != -1 {
					tm.state = SanyLexState(next)
				}
				curPos = 0
				tm.jjmatchedKind = 2147483647
				if tm.readScanChar() {
					continue
				}
			}
			errorLine, errorColumn := tm.stream.getEndLine(), tm.stream.getEndColumn()
			var after []uint16
			eof := false
			if _, err := tm.stream.readChar(); err == nil {
				tm.stream.backup(1)
			} else {
				eof = true
				if curPos > 1 {
					after = tm.stream.getImageUnits()
				}
				if tm.curChar == '\n' || tm.curChar == '\r' {
					errorLine++
					errorColumn = 0
				} else {
					errorColumn++
				}
			}
			if !eof {
				tm.stream.backup(1)
				if curPos > 1 {
					after = tm.stream.getImageUnits()
				}
			}
			tm.lexicalBegin = tm.beginPosition()
			tm.lastEnd = tm.endPosition()
			message := sanyLexicalErrorUnits(eof, int(errorLine), int(errorColumn), after, tm.curChar)
			diagnostic := errorAt(tm.lexicalBegin, "E1200", "%s", message)
			diagnostic.SANYParseMessage = message
			failure := NewSanyTokenMgrErrorMessage(message, SanyTokenMgrLexicalError)
			failure.diagnostic = diagnostic
			panic(failure)
		}
	}
}
