package tlago

import (
	"fmt"
	"runtime"
	"strings"
)

type sanyParseFrame struct {
	name  string
	token *SanyToken
}

type sanyParseException struct {
	diagnostic             Diagnostic
	specialConstructor     bool
	message                string
	currentToken           *SanyToken
	expectedTokenSequences [][]SanyTokenKind
}

// beginProduction/endProduction port the message state in Java's bpa/epa.
// A thrown exception snapshots that state before the Go stack unwinds.
func (p *SanyParser) beginProduction(name string) {
	p.log(SanyLogTrace, "Beginning %s", name)
	p.messageStack = append(p.messageStack, sanyParseFrame{name, p.peek()})
	p.expecting = ""
}

func (p *SanyParser) endProduction() {
	// Java does not execute epa when the production throws.
	if failure := recover(); failure != nil {
		panic(failure)
	}
	name := p.messageStack[len(p.messageStack)-1].name
	p.messageStack = p.messageStack[:len(p.messageStack)-1]
	p.log(SanyLogTrace, "Ending %s", name)
	p.expecting = ""
}

// recordDirectChoice retains JavaCC jj_la1's generation at a failed switch.
func (p *SanyParser) recordDirectChoice(site int) {
	p.directChoiceGeneration[site] = p.at
	p.directChoiceRecorded[site] = true
}

// Direct expected tokens precede rescan entries in ascending source token order.
func (p *SanyParser) directExpectedSequences() [][]SanyTokenKind {
	var tokens [295]bool
	if p.expectedKindRecorded && p.expectedKind >= 0 {
		tokens[p.expectedKind] = true
		p.expectedKind = -1
		p.expectedKindRecorded = false
	}
	for site, masks := range sanyDirectChoiceMasks {
		if !p.directChoiceRecorded[site] || p.directChoiceGeneration[site] != p.at {
			continue
		}
		for kind := range tokens {
			if masks[kind/32]&(uint32(1)<<uint(kind%32)) != 0 {
				tokens[kind] = true
			}
		}
	}
	var result [][]SanyTokenKind
	for kind, present := range tokens {
		if present {
			result = append(result, []SanyTokenKind{SanyTokenKind(kind)})
		}
	}
	return result
}

// generateParseException ports JavaCC's generated constructor path without
// consuming tokens or formatting the exception. The failed kind is cleared by
// directExpectedSequences before saved lookahead calls contribute their entries.
func (p *SanyParser) generateParseException() *sanyParseException {
	return &sanyParseException{
		specialConstructor:     true,
		currentToken:           p.currentToken(),
		expectedTokenSequences: p.rescanLookaheads(p.directExpectedSequences()),
	}
}

func (p *SanyParser) throwParseException(kind SanyTokenKind, nativeMessage string) {
	// jj_consume_token obtains the next token before recording the failed kind.
	token := p.peek()
	p.expectedKind, p.expectedKindRecorded = kind, kind >= 0
	failure := p.generateParseException()
	maxSize := 0
	for _, sequence := range failure.expectedTokenSequences {
		if len(sequence) > maxSize {
			maxSize = len(sequence)
		}
	}
	if maxSize > 0 {
		p.tokenAt(maxSize - 1)
	}
	failure.diagnostic = p.reportedParseException(failure.shortMessage(), token.Begin, "E1300", nativeMessage).diagnostic
	panic(failure)
}

// shortMessage ports ParseExceptionExtended.getShortMessage. Only the prior
// token is escaped; following token images are printed literally.
func (e *sanyParseException) shortMessage() string {
	if !e.specialConstructor {
		return e.Error()
	}
	maxSize := 0
	for _, sequence := range e.expectedTokenSequences {
		if len(sequence) > maxSize {
			maxSize = len(sequence)
		}
	}
	var message strings.Builder
	message.WriteString("Encountered \"")
	token := e.currentToken.Next
	for i := 0; i < maxSize; i++ {
		if i != 0 {
			message.WriteByte(' ')
		}
		if token.Kind == SanyTokenEOF {
			message.WriteString(token.Kind.JavaImage())
			break
		}
		message.WriteString(token.Image)
		token = token.Next
	}
	next := e.currentToken.Next
	fmt.Fprintf(&message, "\" at line %d, column %d and token \"%s\" ", next.Begin.Line, next.Begin.Column, sanyLexicalEscapes(e.currentToken.Image))
	return message.String()
}

// Error ports ParseException.getMessage, including the generated constructor's
// complete alternatives and escapes. Ordinary message constructors return the
// supplied message directly.
func (e *sanyParseException) Error() string {
	if !e.specialConstructor {
		return e.message
	}
	eol := "\n"
	if runtime.GOOS == "windows" {
		eol = "\r\n"
	}
	var expected strings.Builder
	maxSize := 0
	for _, sequence := range e.expectedTokenSequences {
		if len(sequence) > maxSize {
			maxSize = len(sequence)
		}
		for _, kind := range sequence {
			expected.WriteString(kind.JavaImage())
			expected.WriteByte(' ')
		}
		if sequence[len(sequence)-1] != SanyTokenEOF {
			expected.WriteString("...")
		}
		expected.WriteString(eol)
		expected.WriteString("    ")
	}
	var message strings.Builder
	message.WriteString("Encountered \"")
	token := e.currentToken.Next
	for i := 0; i < maxSize; i++ {
		if i != 0 {
			message.WriteByte(' ')
		}
		if token.Kind == SanyTokenEOF {
			message.WriteString(token.Kind.JavaImage())
			break
		}
		message.WriteString(sanyLexicalEscapes(token.Image))
		token = token.Next
	}
	next := e.currentToken.Next
	fmt.Fprintf(&message, "\" at line %d, column %d.%s", next.Begin.Line, next.Begin.Column, eol)
	if len(e.expectedTokenSequences) == 1 {
		message.WriteString("Was expecting:")
	} else {
		message.WriteString("Was expecting one of:")
	}
	message.WriteString(eol)
	message.WriteString("    ")
	message.WriteString(expected.String())
	return message.String()
}

// A source ParseException created with a message uses getShortMessage's ordinary
// constructor branch, without generated expected tokens or an Encountered line.
func (p *SanyParser) throwReportedParseException(shortMessage string, position Position, code, nativeMessage string) {
	panic(p.reportedParseException(shortMessage, position, code, nativeMessage))
}

func (p *SanyParser) reportedParseException(shortMessage string, position Position, code, nativeMessage string) *sanyParseException {
	var message strings.Builder
	message.WriteString("***Parse Error***\n")
	if p.expecting != "" {
		message.WriteString("Was expecting \"" + p.expecting + "\"\n")
	}
	message.WriteString(shortMessage)
	message.WriteString("\n\nResidual stack trace follows:\n")
	last := len(p.messageStack) - 5
	if last < 0 {
		last = 0
	}
	for i := len(p.messageStack) - 1; i >= last; i-- {
		frame := p.messageStack[i]
		fmt.Fprintf(&message, "%s starting at line %d, column %d.\n", frame.name, frame.token.Begin.Line, frame.token.Begin.Column)
	}
	diagnostic := errorAt(position, code, "%s", nativeMessage)
	diagnostic.SANYParseMessage = message.String()
	return &sanyParseException{diagnostic: diagnostic, message: shortMessage}
}

func (p *SanyParser) throwOperatorStackFailure(failure error, position Position) {
	if source, ok := failure.(*sanyOperatorStackFailure); ok {
		p.throwReportedParseException(source.sourceMessage, position, "E1301", source.nativeMessage)
	}
	panic(failure)
}

func (p *SanyParser) consumeParseToken(kind SanyTokenKind, nativeMessage string) *SanySyntaxNode {
	if !p.check(kind) {
		p.throwParseException(kind, nativeMessage)
	}
	return NewSanyTokenNode(p.advance())
}
