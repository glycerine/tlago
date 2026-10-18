package tlago

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"unicode"

	"github.com/glycerine/tlago/tlc"
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
	Kind          SanyNodeKind
	Image         string
	tokenImage    *tlc.UniqueString
	fileNameUS    *tlc.UniqueString
	OriginalImage string
	Range         SanyRange
	Zero          []*SanySyntaxNode
	One           []*SanySyntaxNode
	Heirs         []*SanySyntaxNode
	Token         *SanyToken
	FileName      string
	PreComments   []string
	JunctionList  bool
	ProofLevel    int
	Level         int
	Parent        *SanySyntaxNode
}

var sanyNoPreComments = make([]string, 0)

var SANYNullSyntaxNode = NewSanyBuiltinSyntaxNode("***I do not exist***")

func NewEmptySanySyntaxNode() *SanySyntaxNode {
	return &SanySyntaxNode{
		Zero:        make([]*SanySyntaxNode, 0),
		One:         make([]*SanySyntaxNode, 0),
		PreComments: sanyNoPreComments,
		ProofLevel:  -1,
		Level:       -1,
	}
}

func NewSanyBuiltinSyntaxNode(image string) *SanySyntaxNode {
	imageUS := tlc.UniqueStringOf(image)
	filename := tlc.UniqueStringOf("--TLA+ BUILTINS--")
	return &SanySyntaxNode{
		Image:       image,
		tokenImage:  imageUS,
		FileName:    filename.String(),
		fileNameUS:  filename,
		Zero:        make([]*SanySyntaxNode, 0),
		One:         make([]*SanySyntaxNode, 0),
		PreComments: sanyNoPreComments,
		ProofLevel:  -1,
		Level:       -1,
	}
}

// NewSanyIgnoredSyntaxNode ports the constructor used by Generator for the
// synthetic @ node in EXCEPT. Java deliberately skips image, filename and
// location initialization for this shape and retains the supplied child array.
func NewSanyIgnoredSyntaxNode(kind SanyNodeKind, heirs []*SanySyntaxNode) *SanySyntaxNode {
	return &SanySyntaxNode{
		Kind:        kind,
		Zero:        heirs,
		Heirs:       heirs,
		PreComments: sanyNoPreComments,
		ProofLevel:  -1,
		Level:       -1,
	}
}

func NewSanyNode(kind SanyNodeKind, heirs ...*SanySyntaxNode) *SanySyntaxNode {
	image, imageUS := sanySyntaxNodeImageFields(kind)
	node := &SanySyntaxNode{Kind: kind, Image: image, tokenImage: imageUS, Zero: compactSanyHeirs(heirs), PreComments: sanyNoPreComments, ProofLevel: -1, Level: -1}
	node.refreshHeirsAndRange()
	return node
}

func NewSanySplitNode(kind SanyNodeKind, zero, one []*SanySyntaxNode) *SanySyntaxNode {
	image, imageUS := sanySyntaxNodeImageFields(kind)
	node := &SanySyntaxNode{Kind: kind, Image: image, tokenImage: imageUS, Zero: compactSanyHeirs(zero), One: compactSanyHeirs(one), PreComments: sanyNoPreComments, ProofLevel: -1, Level: -1}
	node.refreshHeirsAndRange()
	return node
}

// Java token-node constructors intern the raw token image during parsing.
// Retain that identity even when a parser production later normalizes Image.
func NewSanyTokenNode(tok *SanyToken) *SanySyntaxNode {
	if tok == nil {
		return nil
	}
	node := &SanySyntaxNode{
		Kind:        SanyNodeKind(tok.Kind),
		Image:       tok.Image,
		tokenImage:  tlc.UniqueStringOf(tok.Image),
		Range:       tok.Range(),
		FileName:    tok.SourceFile,
		fileNameUS:  tlc.UniqueStringOf(tok.SourceFile),
		Token:       tok,
		PreComments: sanyNoPreComments,
		ProofLevel:  -1,
		Level:       -1,
	}
	var comments []*SanyToken
	for special := tok.SpecialToken; special != nil; special = special.SpecialToken {
		comments = append(comments, special)
	}
	for i := len(comments) - 1; i >= 0; i-- {
		node.PreComments = append(node.PreComments, comments[i].Image)
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
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.Kind == kind
}

func (n *SanySyntaxNode) IsGenOp() bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	switch n.Kind.JavaName() {
	case "N_GenPrefixOp", "N_GenNonExpPrefixOp", "N_GenInfixOp", "N_GenPostfixOp":
		return true
	default:
		return false
	}
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
		panic(tlc.NewNullPointerException())
	}
	return append(append(make([]*SanySyntaxNode, 0, len(n.Zero)+len(n.One)), n.Zero...), n.One...)
}

func (n *SanySyntaxNode) GetFilename() string {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	filename := n.GetFN()
	if filename == nil {
		panic(tlc.NewNullPointerException())
	}
	return filename.String()
}

func (n *SanySyntaxNode) GetFN() *tlc.UniqueString {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.fileNameUS != nil {
		return n.fileNameUS
	}
	if n.FileName == "" {
		return nil
	}
	return tlc.UniqueStringOf(n.FileName)
}

func (n *SanySyntaxNode) GetLocation() tlc.SourceLocation {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanySyntaxLocation(n)
}

func (n *SanySyntaxNode) GetImage() string {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	image := n.GetUS()
	if image == nil {
		panic(tlc.NewNullPointerException())
	}
	return image.String()
}

func (n *SanySyntaxNode) GetUS() *tlc.UniqueString {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.Image == "" && n.tokenImage == nil {
		return nil
	}
	return tlc.UniqueStringOf(n.Image)
}

// Java returns the backing array rather than a defensive copy.
func (n *SanySyntaxNode) GetPreComments() []string {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.PreComments == nil {
		return sanyNoPreComments
	}
	return n.PreComments
}

func (n *SanySyntaxNode) First() *SanySyntaxNode {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.Zero != nil {
		if len(n.Zero) == 0 {
			panic(tlc.NewArrayIndexOutOfBoundsException(0, 0))
		}
		return n.Zero[0]
	}
	if len(n.One) == 0 {
		panic(tlc.NewArrayIndexOutOfBoundsException(0, 0))
	}
	return n.One[0]
}

func (n *SanySyntaxNode) Local() bool {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.Zero != nil
}

// PrintST ports SyntaxTreeNode.printST's direct System.out diagnostic tree.
func (n *SanySyntaxNode) PrintST(indentLevel int) {
	n.PrintSTTo(os.Stdout, indentLevel)
}

// PrintSTTo retains PrintST's formatting while allowing an owned writer at the
// Go boundary. As with PrintStream, write failures are not returned.
func (n *SanySyntaxNode) PrintSTTo(out io.Writer, indentLevel int) {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	heirs := n.GetHeirs()
	operator := ""
	if n.Image == "N_OperatorDefinition" {
		if heirs[0].Image == "N_IdentLHS" {
			operator = "*" + heirs[0].GetHeirs()[0].Image
		}
		if heirs[1].Image == "N_IdentLHS" {
			operator = heirs[1].GetHeirs()[0].Image
		}
		if heirs[0].Image == "N_InfixLHS" {
			operator = heirs[0].GetHeirs()[1].Image
		}
	}
	for i := 0; i < indentLevel; i++ {
		_, _ = io.WriteString(out, "| ")
	}
	image := n.Image
	if image == "" && n.tokenImage == nil {
		image = "(" + sanySyntaxNodeImage(n.Kind) + ")"
	}
	_, _ = fmt.Fprintf(out, "%s\t", image)
	if operator != "" {
		_, _ = fmt.Fprintf(out, "%s\t", operator)
	}
	_, _ = fmt.Fprintf(out, "  #heirs: %d\t  kind:   %d%s\n", len(heirs), int(n.Kind), SanyPreCommentToString(n.PreComments))
	for _, heir := range heirs {
		if heir != nil {
			heir.PrintSTTo(out, indentLevel+1)
			continue
		}
		for j := 0; j <= indentLevel; j++ {
			_, _ = io.WriteString(out, "| ")
		}
		_, _ = io.WriteString(out, "<null>\n")
	}
}

// GetAttachedComments follows the leftmost heir to the token, as in SANY.
func (n *SanySyntaxNode) GetAttachedComments() []string {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	if n.Kind < SanyNodeNullId {
		return n.GetPreComments()
	}
	heirs := n.GetHeirs()
	if len(heirs) == 0 {
		return []string{}
	}
	return heirs[0].GetAttachedComments()
}

// Expose the child images separately across the parser/TLC package boundary.
// OpDefNode prints spaces after one children; LNStateEnabled.toDotViz joins
// zero and one children without spaces.
func (n *SanySyntaxNode) GetZeroHumanReadableImages() []string {
	if n.Zero == nil {
		return nil
	}
	images := make([]string, len(n.Zero))
	for i, child := range n.Zero {
		images[i] = child.GetHumanReadableImage()
	}
	return images
}

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
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
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
	image := n.GetImage()
	for _, prefix := range []string{"N_", "Not a node", "Token"} {
		if strings.HasPrefix(image, prefix) {
			return ""
		}
	}
	return image
}

func (n *SanySyntaxNode) GetProofLevel() int {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.ProofLevel
}

// SetLevel and SetParent mirror ParseUnit's post-parse tree initialization.
func (n *SanySyntaxNode) SetLevel(level int) {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	n.Level = level
	for _, child := range n.GetHeirs() {
		child.SetLevel(level + 1)
	}
}

func (n *SanySyntaxNode) GetLevel() int {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.Level
}

func (n *SanySyntaxNode) GetParent() *SanySyntaxNode {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	return n.Parent
}

func (n *SanySyntaxNode) SetParent() {
	if n == nil {
		panic(tlc.NewNullPointerException())
	}
	for _, child := range n.GetHeirs() {
		if child == nil {
			panic(tlc.NewNullPointerException())
		}
		child.Parent = n
		child.SetParent()
	}
}

func (n *SanySyntaxNode) GetOperatorDefinition() any {
	return SanySyntaxNodeGetOperatorDefinition(n)
}

func (n *SanySyntaxNode) String() string {
	return n.GetHumanReadableImage()
}

func SanySyntaxNodeGetOperatorDefinition(node *SanySyntaxNode) *SanySyntaxNode {
	for node != nil {
		if node.Kind == SanySyntaxNodeKindByName["N_OperatorDefinition"] {
			return node
		}
		node = node.Parent
	}
	return nil
}

func SanyPreCommentToString(comments []string) string {
	if len(comments) == 0 {
		return ""
	}
	var out strings.Builder
	out.WriteString("\n preComment: ")
	for i, comment := range comments {
		if i != 0 {
			out.WriteString("\n             ")
		}
		fmt.Fprintf(&out, "%d %s", i, comment)
	}
	return out.String()
}

var sanyCommentLinePattern = regexp.MustCompile(`\r?\n`)

func sanyJavaCommentLines(text string) []string {
	lines := sanyCommentLinePattern.Split(text, -1)
	if text != "" {
		for len(lines) > 0 && lines[len(lines)-1] == "" {
			lines = lines[:len(lines)-1]
		}
	}
	return lines
}

func sanyJavaCharacterWhitespace(char rune) bool {
	if char >= 0x1c && char <= 0x1f {
		return true
	}
	if char == 0x85 || char == 0xa0 || char == 0x2007 || char == 0x202f {
		return false
	}
	return unicode.IsSpace(char)
}

func SanyUnboxComment(text string) string {
	return *SanyUnboxCommentNullable(&text)
}

func SanyUnboxCommentNullable(text *string) *string {
	if text == nil {
		return nil
	}
	result := sanyUnboxComment(*text)
	return &result
}

func sanyUnboxComment(text string) string {
	start, end := strings.Index(text, "(**"), strings.LastIndex(text, "**)")
	if text == "" || start < 0 || end < 0 || end <= start {
		return text
	}
	lines := sanyJavaCommentLines(text)
	beginLine, endLine := -1, -1
	for i, line := range lines {
		if beginLine < 0 && strings.Contains(line, "(**") {
			beginLine = i
		}
		if strings.Contains(line, "**)") {
			endLine = i
		}
	}
	if beginLine < 0 || endLine < 0 || beginLine >= endLine {
		return text
	}
	var out strings.Builder
	for i := beginLine + 1; i < endLine; i++ {
		line := lines[i]
		open, close := strings.Index(line, "(*"), strings.Index(line, "*)")
		if open >= 0 && close > open {
			contentStart := open + 3
			if contentStart <= close {
				out.WriteString(strings.TrimRightFunc(line[contentStart:close], sanyJavaCharacterWhitespace))
			}
		} else {
			out.WriteString(line)
		}
		if i < endLine-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
}

func SanyUnboxBackslashStarComment(text string) string {
	return *SanyUnboxBackslashStarCommentNullable(&text)
}

func SanyUnboxBackslashStarCommentNullable(text *string) *string {
	if text == nil {
		return nil
	}
	result := sanyUnboxBackslashStarComment(*text)
	return &result
}

func sanyUnboxBackslashStarComment(text string) string {
	if text == "" {
		return text
	}
	lines := sanyJavaCommentLines(text)
	count := 0
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimFunc(line, func(r rune) bool { return r <= 0x20 }), `\*`) {
			count++
		}
	}
	if count < len(lines)/2 {
		return text
	}
	var out strings.Builder
	for i, line := range lines {
		trimmed := strings.TrimFunc(line, func(r rune) bool { return r <= 0x20 })
		if strings.HasPrefix(trimmed, `\*`) {
			if pos := strings.Index(line, `\*`); pos >= 0 {
				start := pos + 2
				if start < len(line) && line[start] == ' ' {
					start++
				}
				if start < len(line) {
					out.WriteString(line[start:])
				}
			}
		} else {
			out.WriteString(line)
		}
		if i < len(lines)-1 {
			out.WriteByte('\n')
		}
	}
	return out.String()
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
		if n.fileNameUS == nil && n.FileName == "" {
			n.FileName = heir.FileName
			n.fileNameUS = heir.GetFN()
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
	if heirs == nil {
		return nil
	}
	out := make([]*SanySyntaxNode, 0, len(heirs))
	for _, heir := range heirs {
		if heir == nil {
			panic(tlc.NewNullPointerException())
		}
		out = append(out, heir)
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
