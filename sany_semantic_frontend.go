package tlago

import "github.com/glycerine/tlago/tlc"

// SanyFrontEnd adapts the legacy semantic.FrontEnd utility holder.
type SanyFrontEnd struct{}

// NewSanyFrontEnd constructs the stateless legacy utility holder.
func NewSanyFrontEnd() *SanyFrontEnd { return &SanyFrontEnd{} }

// SanyFrontEndParseAll preserves FrontEnd.parseAll's deliberate null stub.
func SanyFrontEndParseAll(_ *tlc.NamedInputStream, _ int, _ tlc.FilenameToStream) []*SanySyntaxNode {
	return nil
}

// SanyFrontEndParseOne preserves FrontEnd.parseOne's deliberate null stub.
func SanyFrontEndParseOne(_ *tlc.NamedInputStream, _ int) *SanySyntaxNode { return nil }

// SanyFrontEndSemanticAnalysis preserves FrontEnd.semanticAnalysis's null stub.
func SanyFrontEndSemanticAnalysis(_ *SanySyntaxNode, _ int, _ any) any { return nil }

// SanyFrontEndGetToolID returns and advances FrontEnd's process-wide int ID.
func SanyFrontEndGetToolID() int32 { return tlc.GetSemanticToolID() }
