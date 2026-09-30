package tlc

import (
	"fmt"
	"strings"
	"unicode"
)

const valuesWidthProperty = "tlc2.value.Values.width"

const (
	ppNodeConstant     = 1
	ppNodeSet          = 2
	ppNodeSequence     = 3
	ppNodeRecord       = 4
	ppNodeRecordPair   = 5
	ppNodeFunction     = 6
	ppNodeFunctionPair = 7
	ppNodeSubset       = 8
)

type ppNode struct {
	typ       int
	first     int
	last      int
	children  *ppNode
	next      *ppNode
	source    string
	formatted string
}

func newPPNode(source string, first int, typ int) *ppNode {
	return &ppNode{source: source, first: first, typ: typ}
}

func newPPNodeRange(source string, first int, last int, typ int) *ppNode {
	return &ppNode{source: source, first: first, last: last, typ: typ}
}

func (n *ppNode) value() string {
	if n == nil || n.first < 0 || n.last < n.first || n.last >= len(n.source) {
		return ""
	}
	return n.source[n.first : n.last+1]
}

func (n *ppNode) length() int {
	if n == nil {
		return 0
	}
	return n.last - n.first + 1
}

func ValuesPPR(value Value) string {
	if value == nil {
		return "null"
	}
	return ValuesPPRString(value.String())
}

func ValuesPPRString(value string) string {
	return prettyPrintMypp(value, valuesPPRWidth())
}

func valuesPPRWidth() int {
	if value, ok := tlcLookupSystemProperty(valuesWidthProperty); ok {
		if parsed, ok := javaIntProperty(value); ok {
			return parsed
		}
	}
	return 80
}

func prettyPrintMypp(value string, width int) string {
	tree, err := ppParse(value, 0)
	if err != nil {
		return value
	}
	if tree.last < len(value)-1 {
		return value
	}
	formatted, err := ppFormat(tree, width, 0, "")
	if err != nil {
		return value
	}
	return formatted
}

func prettyPrintPP(value string, width int, padding string) (string, error) {
	tree, err := ppParse(value, 0)
	if err != nil {
		return value, err
	}
	return ppFormat(tree, width, 0, padding)
}

func ppParse(source string, index int) (*ppNode, error) {
	if source == "" {
		return nil, fmt.Errorf("TLC Bug: while formating output, the formatter was called with an empty string for a value")
	}
	if index < 0 || index > len(source)-1 {
		return nil, fmt.Errorf("TLC Bug: while formating output, the formatter was called with a string %s for a value and an index %d out of bounds", source, index)
	}
	index, err := ppSkipWhitespace(source, index)
	if err != nil {
		return nil, err
	}
	switch source[index] {
	case '{':
		return ppParseSet(source, index)
	case '<':
		return ppParseSequence(source, index)
	case '[':
		return ppParseRecord(source, index)
	case '(':
		return ppParseFunction(source, index)
	}
	if strings.HasPrefix(source[index:], "SUBSET") {
		return ppParseSubset(source, index)
	}
	return ppParseConstant(source, index)
}

func ppParseSet(source string, index int) (*ppNode, error) {
	first := index
	last := index
	node := newPPNode(source, first, ppNodeSet)
	var firstElement *ppNode
	var lastElement *ppNode
	last++
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	for {
		if last >= len(source) {
			return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a set starting at index %d in the value %s", index, source)
		}
		if source[last] == '}' {
			break
		}
		elt, err := ppParse(source, last)
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
		if last < len(source) && source[last] == ',' {
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

func ppParseSequence(source string, index int) (*ppNode, error) {
	first := index
	last := index
	node := newPPNode(source, first, ppNodeSequence)
	var firstElement *ppNode
	var lastElement *ppNode
	last++
	last++
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	for {
		if last >= len(source) {
			return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a sequence starting at index %d in the value %s", index, source)
		}
		if source[last] == '>' {
			break
		}
		elt, err := ppParse(source, last)
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
		if last < len(source) && source[last] == ',' {
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

func ppParseRecord(source string, index int) (*ppNode, error) {
	first := index
	last := index
	node := newPPNode(source, first, ppNodeRecord)
	var firstElement *ppNode
	var lastElement *ppNode
	last++
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	for {
		if last >= len(source) {
			return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a record starting at index %d in the value %s", index, source)
		}
		if source[last] == ']' {
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
		if last < len(source) && source[last] == ',' {
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

func ppParseRecordPair(source string, index int) (*ppNode, error) {
	first := index
	last := index
	node := newPPNode(source, first, ppNodeRecordPair)
	field, err := ppParse(source, last)
	if err != nil {
		return nil, err
	}
	last = field.last + 1
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(source[last:], "|->") {
		return nil, fmt.Errorf("TLC Bug: while formating output, the formatter couldn't find the token |-> while parsing a record field/value pair starting at index %d in value %s", index, source)
	}
	last += len("|->")
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	value, err := ppParse(source, last)
	if err != nil {
		return nil, err
	}
	field.next = value
	node.children = field
	node.last = value.last
	return node, nil
}

func ppParseFunction(source string, index int) (*ppNode, error) {
	first := index
	last := index
	node := newPPNode(source, first, ppNodeFunction)
	var firstPair *ppNode
	var lastPair *ppNode
	last++
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	for {
		if last >= len(source) {
			return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a function starting at index %d in the value %s", index, source)
		}
		if source[last] == ')' {
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
		if strings.HasPrefix(source[last:], "@@") {
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

func ppParseSubset(source string, index int) (*ppNode, error) {
	first := index
	last := index
	node := newPPNode(source, first, ppNodeSubset)
	last += len("SUBSET")
	var err error
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	elt, err := ppParse(source, last)
	if err != nil {
		return nil, err
	}
	node.children = elt
	node.last = elt.last
	return node, nil
}

func ppParseFunctionPair(source string, index int) (*ppNode, error) {
	first := index
	last := index
	node := newPPNode(source, first, ppNodeFunctionPair)
	domain, err := ppParse(source, last)
	if err != nil {
		return nil, err
	}
	last = domain.last + 1
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(source[last:], ":>") {
		return nil, fmt.Errorf("TLC Bug: while formating output, the formatter couldn't find token :> while parsing a function arg/value pair starting at index %d in value %s", index, source)
	}
	last += len(":>")
	last, err = ppSkipWhitespace(source, last)
	if err != nil {
		return nil, err
	}
	rng, err := ppParse(source, last)
	if err != nil {
		return nil, err
	}
	domain.next = rng
	node.children = domain
	node.last = rng.last
	return node, nil
}

func ppParseConstant(source string, index int) (*ppNode, error) {
	first, err := ppSkipWhitespace(source, index)
	if err != nil {
		return nil, err
	}
	last := first
	if source[first] == '"' {
		for {
			last++
			if last >= len(source) {
				return nil, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while parsing a constant starting at index %d in the value %s", index, source)
			}
			if source[last] == '"' {
				return newPPNodeRange(source, first, last, ppNodeConstant), nil
			}
		}
	}
	if isPPDigit(source[last]) {
		if node := ppParseInterval(source, index); node != nil {
			return node, nil
		}
	}
	if source[last] == '-' {
		last++
		for last < len(source) && isPPDigit(source[last]) {
			last++
		}
		last--
		return newPPNodeRange(source, first, last, ppNodeConstant), nil
	}
	if isPPLetterOrDigit(source[last]) {
		for last < len(source) && isPPLetterOrDigit(source[last]) {
			last++
		}
		last--
		return newPPNodeRange(source, first, last, ppNodeConstant), nil
	}
	return nil, fmt.Errorf("TLC Bug: while formating output, the formatter discovered an illegal character while parsing a constant starting at index %d in the value %s", first, source)
}

func ppParseInterval(source string, index int) *ppNode {
	first := index
	last := first
	if last >= len(source) || !isPPDigit(source[last]) {
		return nil
	}
	for last < len(source) && isPPDigit(source[last]) {
		last++
	}
	if !strings.HasPrefix(source[last:], "..") {
		return nil
	}
	last += 2
	if last >= len(source) || !isPPDigit(source[last]) {
		return nil
	}
	for last < len(source) && isPPDigit(source[last]) {
		last++
	}
	last--
	return newPPNodeRange(source, first, last, ppNodeConstant)
}

func ppSkipWhitespace(source string, index int) (int, error) {
	i := index
	for {
		if i >= len(source) {
			return 0, fmt.Errorf("TLC Bug: while formating output, the formatter ran off the end of the string while skipping whitespace starting at index %d in the value %s", index, source)
		}
		if !unicode.IsSpace(rune(source[i])) {
			return i, nil
		}
		i++
	}
}

func isPPDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isPPLetterOrDigit(ch byte) bool {
	return (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z') || isPPDigit(ch)
}

func ppFormat(value *ppNode, columnWidth int, trailerWidth int, leftPad string) (string, error) {
	if value.length() <= columnWidth-trailerWidth {
		return value.value(), nil
	}
	switch value.typ {
	case ppNodeConstant:
		return value.value(), nil
	case ppNodeSet:
		return ppFormatSimpleValue(value, columnWidth, trailerWidth, leftPad, "{ ", " }", "  ", ",")
	case ppNodeSequence:
		return ppFormatSimpleValue(value, columnWidth, trailerWidth, leftPad, "<< ", " >>", "   ", ",")
	case ppNodeFunction:
		return ppFormatPairValue(value, columnWidth, trailerWidth, leftPad, "( ", " )", "  ", " @@", " :>", "    ")
	case ppNodeRecord:
		return ppFormatPairValue(value, columnWidth, trailerWidth, leftPad, "[ ", " ]", "  ", ",", " |->", "    ")
	case ppNodeSubset:
		return ppFormatSimpleValue(value, columnWidth, trailerWidth, leftPad, "SUBSET ", " ", " ", ",")
	default:
		return "", fmt.Errorf("TLC Bug: while formating output, the formatter called with an expression of type %d while formatting %s and this should never happen", value.typ, value.value())
	}
}

func ppFormatSimpleValue(list *ppNode, columnWidth int, trailerWidth int, leftPad string, open string, close string, pad string, sep string) (string, error) {
	values, err := ppFormatValues(list.children, columnWidth-maxInt(len(open), len(pad)), trailerWidth+len(close), leftPad+pad, sep)
	if err != nil {
		return "", err
	}
	return open + values + close, nil
}

func ppFormatValues(values *ppNode, columnWidth int, trailerWidth int, leftPad string, sep string) (string, error) {
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

func ppFormatPairValue(list *ppNode, columnWidth int, trailerWidth int, leftPad string, open string, close string, pad string, sep string, div string, divPad string) (string, error) {
	pairs, err := ppFormatPairs(list.children, columnWidth-maxInt(len(open), len(pad)), trailerWidth+len(close), leftPad+pad, sep, div, divPad)
	if err != nil {
		return "", err
	}
	return open + pairs + close, nil
}

func ppFormatPairs(pairs *ppNode, columnWidth int, trailerWidth int, leftPad string, sep string, div string, divPad string) (string, error) {
	pp := ""
	for pair := pairs; pair != nil; pair = pair.next {
		arg := pair.children
		val := arg.next
		if pair.next != nil {
			if arg.length()+len(div)+1+val.length()+len(sep) <= columnWidth {
				pp += arg.value() + div + " " + val.value() + sep + "\n" + leftPad
				continue
			}
			if arg.length()+len(div) <= columnWidth {
				formatted, err := ppFormat(val, columnWidth-len(divPad), len(sep), leftPad+divPad)
				if err != nil {
					return "", err
				}
				pp += arg.value() + div + "\n" + leftPad + divPad + formatted + sep + "\n" + leftPad
				continue
			}
			formattedArg, err := ppFormat(arg, columnWidth, trailerWidth+len(div), leftPad)
			if err != nil {
				return "", err
			}
			formattedVal, err := ppFormat(val, columnWidth-len(divPad), len(sep), leftPad+divPad)
			if err != nil {
				return "", err
			}
			pp += formattedArg + div + "\n" + leftPad + divPad + formattedVal + sep + "\n" + leftPad
			continue
		}
		if arg.length()+len(div)+1+val.length() <= columnWidth-trailerWidth {
			pp += arg.value() + div + " " + val.value()
			continue
		}
		if arg.length()+len(div) <= columnWidth {
			formatted, err := ppFormat(val, columnWidth-len(divPad), trailerWidth, leftPad+divPad)
			if err != nil {
				return "", err
			}
			pp += arg.value() + div + "\n" + leftPad + divPad + formatted
			continue
		}
		formattedArg, err := ppFormat(arg, columnWidth, trailerWidth+len(div), leftPad)
		if err != nil {
			return "", err
		}
		formattedVal, err := ppFormat(val, columnWidth-len(divPad), trailerWidth, leftPad+divPad)
		if err != nil {
			return "", err
		}
		pp += formattedArg + div + "\n" + leftPad + divPad + formattedVal
	}
	return pp, nil
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}
