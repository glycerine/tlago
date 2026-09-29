package tlago

import (
	"regexp"
	"strconv"
	"strings"
)

var sanyLocationRE = regexp.MustCompile(`^line ([0-9]+), col ([0-9]+) to line ([0-9]+), col ([0-9]+) of module (.+)$`)

func (p Position) Includes(other Position) bool {
	if p.File != "" && other.File != "" && p.File != other.File {
		return false
	}
	if compareLineColumn(p.Line, p.Column, other.Line, other.Column) > 0 {
		return false
	}
	pEndLine, pEndColumn := p.endLineColumn()
	otherEndLine, otherEndColumn := other.endLineColumn()
	return compareLineColumn(pEndLine, pEndColumn, otherEndLine, otherEndColumn) >= 0
}

func (p Position) Compare(other Position) int {
	if cmp := compareLineColumn(p.Line, p.Column, other.Line, other.Column); cmp != 0 {
		return cmp
	}
	pEndLine, pEndColumn := p.endLineColumn()
	otherEndLine, otherEndColumn := other.endLineColumn()
	if cmp := compareLineColumn(pEndLine, pEndColumn, otherEndLine, otherEndColumn); cmp != 0 {
		return cmp
	}
	return strings.Compare(p.File, other.File)
}

func ParseSANYLocation(text string) (Position, bool) {
	match := sanyLocationRE.FindStringSubmatch(strings.TrimSpace(text))
	if match == nil {
		return Position{}, false
	}
	line, err := strconv.Atoi(match[1])
	if err != nil {
		return Position{}, false
	}
	column, err := strconv.Atoi(match[2])
	if err != nil {
		return Position{}, false
	}
	endLine, err := strconv.Atoi(match[3])
	if err != nil {
		return Position{}, false
	}
	endColumn, err := strconv.Atoi(match[4])
	if err != nil {
		return Position{}, false
	}
	return Position{
		File:      match[5],
		Line:      line,
		Column:    column,
		EndLine:   endLine,
		EndColumn: endColumn,
	}, true
}

func ParseSANYLocations(text string) []Position {
	var out []Position
	for _, line := range strings.Split(text, "\n") {
		if loc, ok := ParseSANYLocation(line); ok {
			out = append(out, loc)
		}
	}
	return out
}

func (p Position) endLineColumn() (int, int) {
	if p.EndLine != 0 || p.EndColumn != 0 {
		return p.EndLine, p.EndColumn
	}
	return p.Line, p.Column
}

func compareLineColumn(leftLine, leftColumn, rightLine, rightColumn int) int {
	if leftLine < rightLine {
		return -1
	}
	if leftLine > rightLine {
		return 1
	}
	if leftColumn < rightColumn {
		return -1
	}
	if leftColumn > rightColumn {
		return 1
	}
	return 0
}
