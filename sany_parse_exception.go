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
	message.WriteString("***Parse Error***\n")
	if p.expecting != "" {
		message.WriteString("Was expecting \"" + p.expecting + "\"\n")
	}
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
	fmt.Fprintf(&message, "\" at line %d, column %d and token \"%s\" \n\nResidual stack trace follows:\n", token.Begin.Line, token.Begin.Column, prior)
	last := len(p.messageStack) - 5
	if last < 0 {
		last = 0
	}
	for i := len(p.messageStack) - 1; i >= last; i-- {
		frame := p.messageStack[i]
		fmt.Fprintf(&message, "%s starting at line %d, column %d.\n", frame.name, frame.token.Begin.Line, frame.token.Begin.Column)
	}
	diagnostic := errorAt(token.Begin, "E1300", "%s", nativeMessage)
	diagnostic.SANYParseMessage = message.String()
	panic(&sanyParseException{diagnostic})
}
