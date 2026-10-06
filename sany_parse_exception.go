package tlago

import (
	"fmt"
	"strings"
)

type sanyParseFrame struct {
	name  string
	token *SanyToken
}

type sanyParseException struct{ diagnostic Diagnostic }

// beginProduction/endProduction port the message state in Java's bpa/epa.
// A thrown exception snapshots that state before the Go stack unwinds.
func (p *SanyParser) beginProduction(name string) {
	p.messageStack = append(p.messageStack, sanyParseFrame{name, p.peek()})
	p.expecting = ""
}

func (p *SanyParser) endProduction() {
	p.messageStack = p.messageStack[:len(p.messageStack)-1]
	p.expecting = ""
}

// throwParseException uses the actual expected token sequences, as JavaCC's
// special ParseException constructor does. getShortMessage uses their maximum
// length when rendering the following input, and escapes only the prior token.
func (p *SanyParser) throwParseException(expected [][]SanyTokenKind, nativeMessage string) {
	var message strings.Builder
	maxSize := 0
	for _, sequence := range expected {
		if len(sequence) > maxSize {
			maxSize = len(sequence)
		}
	}
	message.WriteString("Encountered \"")
	for i := 0; i < maxSize; i++ {
		if i > 0 {
			message.WriteByte(' ')
		}
		token := p.tokenAt(i)
		if token.Kind == SanyTokenEOF {
			message.WriteString(token.Kind.JavaImage())
			break
		}
		message.WriteString(token.Image)
	}
	token := p.peek()
	prior := ""
	if previous := p.previous(); previous != nil {
		prior = sanyLexicalEscapes(previous.Image)
	}
	fmt.Fprintf(&message, "\" at line %d, column %d and token \"%s\" ", token.Begin.Line, token.Begin.Column, prior)
	p.throwReportedParseException(message.String(), token.Begin, "E1300", nativeMessage)
}

// A source ParseException created with a message uses getShortMessage's ordinary
// constructor branch, without generated expected tokens or an Encountered line.
func (p *SanyParser) throwReportedParseException(shortMessage string, position Position, code, nativeMessage string) {
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
	panic(&sanyParseException{diagnostic})
}

func (p *SanyParser) throwOperatorStackFailure(failure error, position Position) {
	if source, ok := failure.(*sanyOperatorStackFailure); ok {
		p.throwReportedParseException(source.sourceMessage, position, "E1301", source.nativeMessage)
	}
	panic(failure)
}

func (p *SanyParser) consumeParseToken(kind SanyTokenKind, nativeMessage string) *SanySyntaxNode {
	if !p.check(kind) {
		p.throwParseException([][]SanyTokenKind{{kind}}, nativeMessage)
	}
	return NewSanyTokenNode(p.advance())
}
