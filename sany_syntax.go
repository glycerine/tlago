package tlago

import (
	"fmt"
	"strings"
)

type SanyRange struct {
	Begin Position
	End   Position
}

type SanySyntaxNodeDefinition struct {
	Kind SanyNodeKind
	Name string
}

type SanySyntaxNode struct {
	Kind         SanyNodeKind
	Image        string
	Original     string
	Range        SanyRange
	Zero         []*SanySyntaxNode
	One          []*SanySyntaxNode
	Heirs        []*SanySyntaxNode
	Token        *SanyToken
	FileName     string
	PreComments  []string
	JunctionList bool
	ProofLevel   int
	Level        int
	Parent       *SanySyntaxNode
}

func NewSanyNode(kind SanyNodeKind, heirs ...*SanySyntaxNode) *SanySyntaxNode {
	node := &SanySyntaxNode{Kind: kind, Image: sanySyntaxNodeImage(kind), Zero: compactSanyHeirs(heirs), ProofLevel: -1, Level: -1}
	node.refreshHeirsAndRange()
	return node
}

func NewSanySplitNode(kind SanyNodeKind, zero, one []*SanySyntaxNode) *SanySyntaxNode {
	node := &SanySyntaxNode{Kind: kind, Image: sanySyntaxNodeImage(kind), Zero: compactSanyHeirs(zero), One: compactSanyHeirs(one), ProofLevel: -1, Level: -1}
	node.refreshHeirsAndRange()
	return node
}

func NewSanyTokenNode(tok *SanyToken) *SanySyntaxNode {
	if tok == nil {
		return nil
	}
	node := &SanySyntaxNode{
		Kind:       SanyNodeKind(tok.Kind),
		Image:      tok.Image,
		Range:      tok.Range(),
		Token:      tok,
		ProofLevel: -1,
		Level:      -1,
	}
	for special := tok.Special; special != nil; special = special.Next {
		node.PreComments = append(node.PreComments, special.Image)
	}
	return node
}

func (k SanyNodeKind) JavaName() string {
	for _, def := range SanySyntaxNodeKinds {
		if def.Kind == k {
			return def.Name
		}
	}
	tok := SanyTokenKind(k)
	if int(tok) >= 0 && int(tok) < len(SanyTokenImages) {
		return tok.JavaName()
	}
	return fmt.Sprintf("node(%d)", int(k))
}

func (n *SanySyntaxNode) IsKind(kind SanyNodeKind) bool {
	return n != nil && n.Kind == kind
}

func (n *SanySyntaxNode) AddHeir(heir *SanySyntaxNode) {
	if n == nil || heir == nil {
		return
	}
	n.Zero = append(n.Zero, heir)
	n.refreshHeirsAndRange()
}

func (n *SanySyntaxNode) GetHeirs() []*SanySyntaxNode {
	if n == nil {
		return nil
	}
	return append([]*SanySyntaxNode(nil), n.Heirs...)
}

// GetAttachedComments follows the leftmost heir to the token, as in SANY.
func (n *SanySyntaxNode) GetAttachedComments() []string {
	if n.Kind < SanyNodeNullId {
		return n.PreComments
	}
	if len(n.Heirs) == 0 {
		return nil
	}
	return n.Heirs[0].GetAttachedComments()
}

// Expose the one-child images across the parser/TLC package boundary without
// merging them with zero: OpDefNode prints a space after each one child.
func (n *SanySyntaxNode) GetOneHumanReadableImages() []string {
	if n.One == nil {
		return nil
	}
	images := make([]string, len(n.One))
	for i, child := range n.One {
		images[i] = child.GetHumanReadableImage()
	}
	return images
}

// GetHumanReadableImage mirrors SyntaxTreeNode, including its zero-child
// check: a node with only one children returns its own image.
func (n *SanySyntaxNode) GetHumanReadableImage() string {
	if len(n.Zero) > 0 {
		var out strings.Builder
		for _, child := range n.Zero {
			out.WriteString(child.GetHumanReadableImage())
		}
		for _, child := range n.One {
			out.WriteString(child.GetHumanReadableImage())
		}
		return out.String()
	}
	for _, prefix := range []string{"N_", "Not a node", "Token"} {
		if strings.HasPrefix(n.Image, prefix) {
			return ""
		}
	}
	return n.Image
}

func (n *SanySyntaxNode) GetProofLevel() int {
	if n == nil {
		return -1
	}
	return n.ProofLevel
}

// SetLevel and SetParent mirror ParseUnit's post-parse tree initialization.
func (n *SanySyntaxNode) SetLevel(level int) {
	if n == nil {
		return
	}
	n.Level = level
	for _, child := range n.Heirs {
		child.SetLevel(level + 1)
	}
}

func (n *SanySyntaxNode) GetLevel() int              { return n.Level }
func (n *SanySyntaxNode) GetParent() *SanySyntaxNode { return n.Parent }

func (n *SanySyntaxNode) SetParent() {
	if n == nil {
		return
	}
	for _, child := range n.Heirs {
		child.Parent = n
		child.SetParent()
	}
}

func (n *SanySyntaxNode) GetOperatorDefinition() any {
	for node := n; node != nil; node = node.Parent {
		if node.Kind == SanySyntaxNodeKindByName["N_OperatorDefinition"] {
			return node
		}
	}
	return nil
}

func (n *SanySyntaxNode) refreshHeirsAndRange() {
	if n == nil {
		return
	}
	n.Heirs = append(append([]*SanySyntaxNode(nil), n.Zero...), n.One...)
	// SyntaxTreeNode.updateLocation retains these Java int extrema when no
	// heir supplies a location. Recompute from both child arrays each time.
	n.Range = SanyRange{
		Begin: Position{Line: 2147483647, Column: 2147483647},
		End:   Position{Line: -2147483648, Column: -2147483648},
	}
	for _, heir := range n.Heirs {
		if heir == nil {
			continue
		}
		if n.FileName == "" {
			n.FileName = heir.FileName
		}
		begin, end := heir.Range.Begin, heir.Range.End
		if begin.Line < n.Range.Begin.Line || begin.Line == n.Range.Begin.Line && begin.Column < n.Range.Begin.Column {
			n.Range.Begin = begin
		}
		if end.Line > n.Range.End.Line || end.Line == n.Range.End.Line && end.Column > n.Range.End.Column {
			n.Range.End = end
		}
	}
}

func compactSanyHeirs(heirs []*SanySyntaxNode) []*SanySyntaxNode {
	if len(heirs) == 0 {
		return nil
	}
	out := make([]*SanySyntaxNode, 0, len(heirs))
	for _, heir := range heirs {
		if heir != nil {
			out = append(out, heir)
		}
	}
	return out
}

func positionBefore(a, b Position) bool {
	if a.Line == 0 {
		return false
	}
	if b.Line == 0 {
		return true
	}
	if a.Line != b.Line {
		return a.Line < b.Line
	}
	return a.Column < b.Column
}
