package tlc

import (
	"fmt"
	"strings"
	"sync"
	"unicode"
)

const valuesWidthProperty = "tlc2.value.Values.width"

var valuesStatics struct {
	once  sync.Once
	width int
}

// InitializeValuesStatics represents a fresh Java classloader. Ordinary calls
// retain the width captured when tlc2.value.Values is first used.
func InitializeValuesStatics() {
	valuesStatics.once = sync.Once{}
	valuesStatics.width = 0
}

const (
	NodeConstant     = 1
	NodeSet          = 2
	NodeSequence     = 3
	NodeRecord       = 4
	NodeRecordPair   = 5
	NodeFunction     = 6
	NodeFunctionPair = 7
	NodeSubset       = 8
)

type ppSource struct {
	text  *string
	units []uint16
}

func newPPSource(text string) *ppSource {
	return newPPSourceNullable(javaString(text))
}

func newPPSourceNullable(text *string) *ppSource {
	return &ppSource{text: text, units: javaStringUTF16(*text)}
}

func (s *ppSource) len() int       { return len(s.units) }
func (s *ppSource) String() string { return *s.text }
func (s *ppSource) at(index int) uint16 {
	index = int(int32(index))
	if index < 0 || index >= len(s.units) {
		panic(NewStringIndexOutOfBoundsException(index, len(s.units)))
	}
	return s.units[index]
}
func (s *ppSource) hasPrefix(index int, prefix string) bool {
	index = int(int32(index))
	units := javaStringUTF16(prefix)
	if index < 0 || index+len(units) > len(s.units) {
		return false
	}
	for i, unit := range units {
		if s.units[index+i] != unit {
			return false
		}
	}
	return true
}
func (s *ppSource) substring(first, end int) string {
	first = int(int32(first))
	end = int(int32(end))
	if first < 0 || end < first || end > len(s.units) {
		panic(NewStringIndexOutOfBoundsException(first, len(s.units)))
	}
	return javaStringFromUTF16(s.units[first:end])
}

type Node struct {
	typ       int
	first     int
	last      int
	children  *Node
	next      *Node
	source    *ppSource
	formatted *string
}

func newPPNode(source *ppSource, first int, typ int) *Node {
	return &Node{source: source, first: int(int32(first)), typ: int(int32(typ))}
}

func newPPNodeRange(source *ppSource, first int, last int, typ int) *Node {
	return &Node{source: source, first: int(int32(first)), last: int(int32(last)), typ: int(int32(typ))}
}

func (n *Node) require() *Node {
	if n == nil {
		panic(NewNullPointerException())
	}
	return n
}

func (n *Node) Type() int       { return n.require().typ }
func (n *Node) First() int      { return n.require().first }
func (n *Node) Last() int       { return n.require().last }
func (n *Node) Children() *Node { return n.require().children }
func (n *Node) Next() *Node     { return n.require().next }
func (n *Node) GetString() *string {
	n.require()
	if n.source == nil {
		return nil
	}
	return n.source.text
}
func (n *Node) Formatted() *string      { return n.require().formatted }
func (n *Node) SetType(value int)       { n.require().typ = int(int32(value)) }
func (n *Node) SetFirst(value int)      { n.require().first = int(int32(value)) }
func (n *Node) SetLast(value int)       { n.require().last = int(int32(value)) }
func (n *Node) SetChildren(value *Node) { n.require().children = value }
func (n *Node) SetNext(value *Node)     { n.require().next = value }
func (n *Node) SetString(value *string) {
	n.require()
	if value == nil {
		n.source = nil
		return
	}
	n.source = newPPSourceNullable(value)
}
func (n *Node) SetFormatted(value *string) { n.require().formatted = value }

func (n *Node) Value() string {
	n.require()
	if n.source == nil {
		panic(NewNullPointerException())
	}
	return n.source.substring(n.first, int(int32(n.last)+1))
}

func (n *Node) Length() int {
	n.require()
	return int(int32(n.last) - int32(n.first) + 1)
}

func (n *Node) AppendChild(child *Node) {
	n.require()
	if n.children == nil {
		n.children = child
		return
	}
	current := n.children
	for current.next != nil {
		current = current.next
	}
	current.next = child
}

func (n *Node) String() string {
	n.require()
	var out strings.Builder
	for child := n.children; child != nil; child = child.next {
		out.WriteString(child.String())
	}
	fmt.Fprintf(&out, "%d %d %s\n", n.first, n.last, n.Value())
	return out.String()
}

func (n *Node) value() string { return n.Value() }
func (n *Node) length() int   { return n.Length() }

func ValuesPPR(value Value) string {
	if isNil(value) {
		return "null"
	}
	return ValuesPPRString(value.String())
}

func ValuesPPRString(value string) string {
	return prettyPrintMypp(value, valuesPPRWidth())
}

func valuesPPRWidth() int {
	valuesStatics.once.Do(func() {
		valuesStatics.width = 80
		if value, ok := tlcLookupSystemProperty(valuesWidthProperty); ok {
			if parsed, ok := javaDecodeIntProperty(value); ok {
				valuesStatics.width = parsed
			}
		}
	})
	return valuesStatics.width
}

func prettyPrintMypp(value string, width int) string {
	tree, err := PPrintParse(value, 0)
	if err != nil {
		return value
	}
	// Java catches the substring failure when a parsed range extends past
	// the input (for example, an operator label with a single closing >).
	if tree.last < tree.source.len()-1 || tree.last >= tree.source.len() {
		return value
	}
	formatted, err := ppFormat(tree, width, 0, "")
	if err != nil {
		return value
	}
	return formatted
}

func prettyPrintPP(value string, width int, padding string) (string, error) {
	tree, err := PPrintParse(value, 0)
	if err != nil {
		return value, err
	}
	return ppFormat(tree, width, 0, padding)
}

func PPrintParse(source string, index int) (*Node, error) {
	return PPrintParseNullable(javaString(source), index)
}

func PPrintParseNullable(source *string, index int) (*Node, error) {
	if source == nil {
		return nil, fmt.Errorf("TLC Bug: while formating output, the formatter was called with a null string for a value")
	}
	index = int(int32(index))
	return ppParseSource(newPPSourceNullable(source), index)
}

func PPrintFormat(value *Node, columnWidth int, trailerWidth int, leftPad string) (string, error) {
	columnWidth = int(int32(columnWidth))
	trailerWidth = int(int32(trailerWidth))
	return ppFormat(value, columnWidth, trailerWidth, leftPad)
}

func ppParseSource(source *ppSource, index int) (*Node, error) {
	if *source.text == "" {
		return nil, fmt.Errorf("TLC Bug: while formating output, the formatter was called with an empty string for a value")
	}
	if index < 0 || index > source.len()-1 {
		return nil, fmt.Errorf("TLC Bug: while formating output, the formatter was called with a string %s for a value and an index %d out of bounds", *source.text, index)
	}
	index, err := ppSkipWhitespace(source, index)
	if err != nil {
		return nil, err
	}
	switch source.at(index) {
	case '{':
		return ppParseSet(source, index)
	case '<':
		return ppParseSequence(source, index)
	case '[':
		return ppParseRecord(source, index)
	case '(':
		return ppParseFunction(source, index)
	}
	if source.hasPrefix(index, "SUBSET") {
		return ppParseSubset(source, index)
	}
	return ppParseConstant(source, index)
}

func ppParseSet(source *ppSource, index int) (*Node, error) {
	first := index
	last := index
	node := newPPNode(source, first, NodeSet)
	var firstElement *Node
	var lastElement *Node
	last++
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	for {
		if last >= source.len() {
			return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a set starting at index %d in the value %s", index, source)
		}
		if source.at(last) == '}' {
			break
		}
		elt, err := ppParseSource(source, last)
		if err != nil {
			return nil, err
		}
		if firstElement == nil {
			firstElement = elt
			lastElement = elt
		} else {
			lastElement.next = elt
			lastElement = elt
		}
		last = elt.last + 1
		last, err = ppSkipWhitespace(source, last)
		if err != nil {
			return nil, err
		}
		if last < source.len() && source.at(last) == ',' {
			last++
		}
		last, err = ppSkipWhitespace(source, last)
		if err != nil {
			return nil, err
		}
	}
	node.children = firstElement
	node.last = last
	return node, nil
}

func ppParseSequence(source *ppSource, index int) (*Node, error) {
	first := index
	last := index
	node := newPPNode(source, first, NodeSequence)
	var firstElement *Node
	var lastElement *Node
	last++
	last++
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	for {
		if last >= source.len() {
			return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a sequence starting at index %d in the value %s", index, source)
		}
		if source.at(last) == '>' {
			break
		}
		elt, err := ppParseSource(source, last)
		if err != nil {
			return nil, err
		}
		if firstElement == nil {
			firstElement = elt
			lastElement = elt
		} else {
			lastElement.next = elt
			lastElement = elt
		}
		last = elt.last + 1
		last, err = ppSkipWhitespace(source, last)
		if err != nil {
			return nil, err
		}
		if last < source.len() && source.at(last) == ',' {
			last++
		}
		last, err = ppSkipWhitespace(source, last)
		if err != nil {
			return nil, err
		}
	}
	last++
	node.children = firstElement
	node.last = last
	return node, nil
}

func ppParseRecord(source *ppSource, index int) (*Node, error) {
	first := index
	last := index
	node := newPPNode(source, first, NodeRecord)
	var firstElement *Node
	var lastElement *Node
	last++
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	for {
		if last >= source.len() {
			return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a record starting at index %d in the value %s", index, source)
		}
		if source.at(last) == ']' {
			break
		}
		elt, err := ppParseRecordPair(source, last)
		if err != nil {
			return nil, err
		}
		if firstElement == nil {
			firstElement = elt
			lastElement = elt
		} else {
			lastElement.next = elt
			lastElement = elt
		}
		last = elt.last + 1
		last, err = ppSkipWhitespace(source, last)
		if err != nil {
			return nil, err
		}
		if last < source.len() && source.at(last) == ',' {
			last++
		}
		last, err = ppSkipWhitespace(source, last)
		if err != nil {
			return nil, err
		}
	}
	node.children = firstElement
	node.last = last
	return node, nil
}

func ppParseRecordPair(source *ppSource, index int) (*Node, error) {
	first := index
	last := index
	node := newPPNode(source, first, NodeRecordPair)
	field, err := ppParseSource(source, last)
	if err != nil {
		return nil, err
	}
	last = field.last + 1
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	if !source.hasPrefix(last, "|->") {
		return nil, fmt.Errorf("TLC Bug: while formating output, the formatter couldn't find the token |-> while parsing a record field/value pair starting at index %d in value %s", index, source)
	}
	last += len("|->")
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	value, err := ppParseSource(source, last)
	if err != nil {
		return nil, err
	}
	field.next = value
	node.children = field
	node.last = value.last
	return node, nil
}

func ppParseFunction(source *ppSource, index int) (*Node, error) {
	first := index
	last := index
	node := newPPNode(source, first, NodeFunction)
	var firstPair *Node
	var lastPair *Node
	last++
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	for {
		if last >= source.len() {
			return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a function starting at index %d in the value %s", index, source)
		}
		if source.at(last) == ')' {
			break
		}
		pair, err := ppParseFunctionPair(source, last)
		if err != nil {
			return nil, err
		}
		if firstPair == nil {
			firstPair = pair
			lastPair = pair
		} else {
			lastPair.next = pair
			lastPair = pair
		}
		last = pair.last + 1
		last, err = ppSkipWhitespace(source, last)
		if err != nil {
			return nil, err
		}
		if source.hasPrefix(last, "@@") {
			last += len("@@")
		}
		last, err = ppSkipWhitespace(source, last)
		if err != nil {
			return nil, err
		}
	}
	node.children = firstPair
	node.last = last
	return node, nil
}

func ppParseSubset(source *ppSource, index int) (*Node, error) {
	first := index
	last := index
	node := newPPNode(source, first, NodeSubset)
	last += len("SUBSET")
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	elt, err := ppParseSource(source, last)
	if err != nil {
		return nil, err
	}
	node.children = elt
	node.last = elt.last
	return node, nil
}

func ppParseFunctionPair(source *ppSource, index int) (*Node, error) {
	first := index
	last := index
	node := newPPNode(source, first, NodeFunctionPair)
	domain, err := ppParseSource(source, last)
	if err != nil {
		return nil, err
	}
	last = domain.last + 1
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	if !source.hasPrefix(last, ":>") {
		return nil, fmt.Errorf("TLC Bug: while formating output, the formatter couldn't find token :> while parsing a function arg/value pair starting at index %d in value %s", index, source)
	}
	last += len(":>")
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	rng, err := ppParseSource(source, last)
	if err != nil {
		return nil, err
	}
	domain.next = rng
	node.children = domain
	node.last = rng.last
	return node, nil
}

func ppParseConstant(source *ppSource, index int) (*Node, error) {
	first, err := ppSkipWhitespace(source, index)
	if err != nil {
		return nil, err
	}
	last := first
	if source.at(first) == '"' {
		for {
			last++
			if last >= source.len() {
				return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a constant starting at index %d in the value %s", index, source)
			}
			if source.at(last) == '"' {
				return newPPNodeRange(source, first, last, NodeConstant), nil
			}
		}
	}
	if isPPDigit(source.at(last)) {
		if node := ppParseInterval(source, index); node != nil {
			return node, nil
		}
	}
	if source.at(last) == '-' {
		last++
		for last < source.len() && isPPDigit(source.at(last)) {
			last++
		}
		last--
		return newPPNodeRange(source, first, last, NodeConstant), nil
	}
	if isPPLetterOrDigit(source.at(last)) {
		for last < source.len() && isPPLetterOrDigit(source.at(last)) {
			last++
		}
		last--
		return newPPNodeRange(source, first, last, NodeConstant), nil
	}
	return nil, fmt.Errorf("TLC Bug: while formating output, the formatter discovered an illegal character while parsing a constant starting at index %d in the value %s", first, source)
}

func ppParseInterval(source *ppSource, index int) *Node {
	first := index
	last := first
	if last >= source.len() || !isPPDigit(source.at(last)) {
		return nil
	}
	for last < source.len() && isPPDigit(source.at(last)) {
		last++
	}
	if !source.hasPrefix(last, "..") {
		return nil
	}
	last += 2
	if last >= source.len() || !isPPDigit(source.at(last)) {
		return nil
	}
	for last < source.len() && isPPDigit(source.at(last)) {
		last++
	}
	last--
	return newPPNodeRange(source, first, last, NodeConstant)
}

func ppSkipWhitespace(source *ppSource, index int) (int, error) {
	i := index
	for {
		if i >= source.len() {
			return 0, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while skipping whitespace starting at index %d in the value %s", index, source)
		}
		if !stringHelperWhitespace(source.at(i)) {
			return i, nil
		}
		i++
	}
}

func isPPDigit(ch uint16) bool {
	return unicode.IsDigit(rune(ch))
}

func isPPLetterOrDigit(ch uint16) bool {
	r := rune(ch)
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func ppFormat(value *Node, columnWidth int, trailerWidth int, leftPad string) (string, error) {
	columnWidth = int(int32(columnWidth))
	trailerWidth = int(int32(trailerWidth))
	if value.length() <= ppIntSub(columnWidth, trailerWidth) {
		return value.value(), nil
	}
	switch value.typ {
	case NodeConstant:
		return value.value(), nil
	case NodeSet:
		return ppFormatSimpleValue(value, columnWidth, trailerWidth, leftPad, "{ ", " }", "  ", ",")
	case NodeSequence:
		return ppFormatSimpleValue(value, columnWidth, trailerWidth, leftPad, "<< ", " >>", "   ", ",")
	case NodeFunction:
		return ppFormatPairValue(value, columnWidth, trailerWidth, leftPad, "( ", " )", "  ", " @@", " :>", "    ")
	case NodeRecord:
		return ppFormatPairValue(value, columnWidth, trailerWidth, leftPad, "[ ", " ]", "  ", ",", " |->", "    ")
	case NodeSubset:
		return ppFormatSimpleValue(value, columnWidth, trailerWidth, leftPad, "SUBSET ", " ", " ", ",")
	default:
		return "", fmt.Errorf("TLC Bug: while formating output, the formatter called with an expression of type %d while formatting %s and this should never happen", value.typ, value.value())
	}
}

func ppFormatSimpleValue(list *Node, columnWidth int, trailerWidth int, leftPad string, open string, close string, pad string, sep string) (string, error) {
	values, err := ppFormatValues(list.children, ppIntSub(columnWidth, maxInt(len(open), len(pad))), ppIntAdd(trailerWidth, len(close)), leftPad+pad, sep)
	if err != nil {
		return "", err
	}
	return open + values + close, nil
}

func ppFormatValues(values *Node, columnWidth int, trailerWidth int, leftPad string, sep string) (string, error) {
	pp := ""
	for value := values; value != nil; value = value.next {
		if value.next != nil {
			formatted, err := ppFormat(value, columnWidth, len(sep), leftPad)
			if err != nil {
				return "", err
			}
			pp += formatted + sep + "\n" + leftPad
		} else {
			formatted, err := ppFormat(value, columnWidth, trailerWidth, leftPad)
			if err != nil {
				return "", err
			}
			pp += formatted
		}
	}
	return pp, nil
}

func ppFormatPairValue(list *Node, columnWidth int, trailerWidth int, leftPad string, open string, close string, pad string, sep string, div string, divPad string) (string, error) {
	pairs, err := ppFormatPairs(list.children, ppIntSub(columnWidth, maxInt(len(open), len(pad))), ppIntAdd(trailerWidth, len(close)), leftPad+pad, sep, div, divPad)
	if err != nil {
		return "", err
	}
	return open + pairs + close, nil
}

func ppFormatPairs(pairs *Node, columnWidth int, trailerWidth int, leftPad string, sep string, div string, divPad string) (string, error) {
	pp := ""
	for pair := pairs; pair != nil; pair = pair.next {
		arg := pair.children
		val := arg.next
		if pair.next != nil {
			if ppIntAdd(arg.length(), len(div), 1, val.length(), len(sep)) <= columnWidth {
				pp += arg.value() + div + " " + val.value() + sep + "\n" + leftPad
				continue
			}
			if ppIntAdd(arg.length(), len(div)) <= columnWidth {
				formatted, err := ppFormat(val, ppIntSub(columnWidth, len(divPad)), len(sep), leftPad+divPad)
				if err != nil {
					return "", err
				}
				pp += arg.value() + div + "\n" + leftPad + divPad + formatted + sep + "\n" + leftPad
				continue
			}
			formattedArg, err := ppFormat(arg, columnWidth, ppIntAdd(trailerWidth, len(div)), leftPad)
			if err != nil {
				return "", err
			}
			formattedVal, err := ppFormat(val, ppIntSub(columnWidth, len(divPad)), len(sep), leftPad+divPad)
			if err != nil {
				return "", err
			}
			pp += formattedArg + div + "\n" + leftPad + divPad + formattedVal + sep + "\n" + leftPad
			continue
		}
		if ppIntAdd(arg.length(), len(div), 1, val.length()) <= ppIntSub(columnWidth, trailerWidth) {
			pp += arg.value() + div + " " + val.value()
			continue
		}
		if ppIntAdd(arg.length(), len(div)) <= columnWidth {
			formatted, err := ppFormat(val, ppIntSub(columnWidth, len(divPad)), trailerWidth, leftPad+divPad)
			if err != nil {
				return "", err
			}
			pp += arg.value() + div + "\n" + leftPad + divPad + formatted
			continue
		}
		formattedArg, err := ppFormat(arg, columnWidth, ppIntAdd(trailerWidth, len(div)), leftPad)
		if err != nil {
			return "", err
		}
		formattedVal, err := ppFormat(val, ppIntSub(columnWidth, len(divPad)), trailerWidth, leftPad+divPad)
		if err != nil {
			return "", err
		}
		pp += formattedArg + div + "\n" + leftPad + divPad + formattedVal
	}
	return pp, nil
}

func ppIntAdd(values ...int) int {
	var sum int32
	for _, value := range values {
		sum += int32(value)
	}
	return int(sum)
}

func ppIntSub(left, right int) int {
	return int(int32(left) - int32(right))
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
