package tlago

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

var SANYLocationMatcher = regexp.MustCompile(`line ([0-9]+), col ([0-9]+) to line ([0-9]+), col ([0-9]+) of module ([A-Za-z_0-9]+)`)
var SANYLocationMatcher2 = regexp.MustCompile(`In module ([A-Za-z_0-9]+)`)
var SANYLocationMatcher3 = regexp.MustCompile(`<[A-Za-z_0-9]+ line ([0-9]+), col ([0-9]+) to line ([0-9]+), col ([0-9]+) of module ([A-Za-z_0-9]+)>`)
var SANYLocationMatcher4 = regexp.MustCompile(`Line ([0-9]+), column ([0-9]+) to line ([0-9]+), column ([0-9]+) in ([A-Za-z_0-9]+)`)

var SANYAllLocationPatterns = []*regexp.Regexp{
	SANYLocationMatcher,
	SANYLocationMatcher2,
	SANYLocationMatcher3,
	SANYLocationMatcher4,
}

var SANYNullLocation = NewSANYLocation("--unknown--", 0, 0, 0, 0)

func NewSANYLocation(source string, beginLine, beginColumn, endLine, endColumn int) Position {
	if source != "" {
		tlc.UniqueStringOf(source)
	}
	return Position{File: source, Line: beginLine, Column: beginColumn, EndLine: endLine, EndColumn: endColumn}
}

func NewSANYLocationFromCoordinates(coordinates []int) Position {
	if len(coordinates) != 4 {
		panic(tlc.NewTLCRuntimeExceptionMessage("Illegal coordinates found."))
	}
	position := Position{Line: coordinates[0], Column: coordinates[1], EndLine: coordinates[2], EndColumn: coordinates[3]}
	if position.EndLine < position.Line {
		position.EndLine = position.Line
	}
	if position.EndColumn < position.Column {
		position.EndColumn = position.Column
	}
	return position
}

func SANYModuleLocation(module string) Position {
	return NewSANYLocation(module, 0, 0, 0, 0)
}

func ParseSANYCoordinates(source, coordinates string) Position {
	parts := strings.Split(coordinates, " ")
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	values := [4]int{}
	for i := range values {
		if i >= len(parts) {
			panic(tlc.NewArrayIndexOutOfBoundsException(i, len(parts)))
		}
		value, err := strconv.ParseInt(parts[i], 10, 32)
		if err != nil {
			panic(tlc.NewNumberFormatException(parts[i]))
		}
		values[i] = int(value)
	}
	return NewSANYLocation(source, values[0], values[1], values[2], values[3])
}

type SANYPCalLocation struct {
	Line   int
	Column int
}

type SANYRegion struct {
	Begin SANYPCalLocation
	End   SANYPCalLocation
}

func (p Position) Includes(other Position) bool {
	if p.File != other.File {
		return false
	}
	if compareLineColumn(p.Line, p.Column, other.Line, other.Column) > 0 {
		return false
	}
	return compareLineColumn(p.EndLine, p.EndColumn, other.EndLine, other.EndColumn) >= 0
}

func (p Position) Compare(other Position) int {
	if p.File != other.File {
		if p.File != "" && other.File != "" {
			return tlc.UniqueStringOf(p.File).Compare(tlc.UniqueStringOf(other.File))
		}
		return strings.Compare(p.File, other.File)
	}
	if cmp := compareLineColumn(p.Line, p.Column, other.Line, other.Column); cmp != 0 {
		return cmp
	}
	if cmp := compareLineColumn(p.EndLine, p.EndColumn, other.EndLine, other.EndColumn); cmp != 0 {
		return cmp
	}
	return 0
}

func (p Position) SANYCoordinates() []int {
	return []int{p.Line, p.Column, p.EndLine, p.EndColumn}
}

func (p Position) SANYSourceAsUniqueString() *tlc.UniqueString {
	if p.File == "" {
		return nil
	}
	return tlc.UniqueStringOf(p.File)
}

func (p Position) SANYHashCode() int32 {
	result := int32(1)
	result = 31*result + int32(p.Column)
	result = 31*result + int32(p.Line)
	result = 31*result + int32(p.EndColumn)
	result = 31*result + int32(p.EndLine)
	if source := p.SANYSourceAsUniqueString(); source != nil {
		result = 31*result + source.HashCode()
	} else {
		result *= 31
	}
	return result
}

func (p Position) SANYToRegion() SANYRegion {
	return SANYRegion{
		Begin: SANYPCalLocation{Line: p.Line - 1, Column: p.Column - 1},
		End:   SANYPCalLocation{Line: p.EndLine - 1, Column: p.EndColumn - 1},
	}
}

func (p Position) SANYContainsLine(line int) bool {
	return line >= p.Line && line <= p.EndLine
}

func (p Position) SANYLocationString() string {
	if p == SANYNullLocation {
		return "Unknown location"
	}
	if p.File != "" && p.File != "--unknown--" && p.Line == 0 && p.Column == 0 && p.EndLine == 0 && p.EndColumn == 0 {
		return "In module " + p.File
	}
	module := p.File
	if module == "" {
		module = "null"
	}
	return "line " + strconv.Itoa(p.Line) + ", col " + strconv.Itoa(p.Column) + " to line " + strconv.Itoa(p.EndLine) + ", col " + strconv.Itoa(p.EndColumn) + " of module " + module
}

func (p Position) SANYLinesAndColumns() string {
	text := p.SANYLocationString()
	if index := strings.Index(text, " of module "); index >= 0 {
		return text[:index]
	}
	return text
}

func ParseSANYLocation(text string) (Position, bool) {
	if text == "" || text == "Unknown location" {
		return Position{}, false
	}
	patterns := [...]*regexp.Regexp{SANYLocationMatcher, SANYLocationMatcher2, SANYLocationMatcher3, SANYLocationMatcher4}
	for index, pattern := range patterns {
		candidate := text
		if index >= 2 {
			candidate = strings.TrimFunc(candidate, func(r rune) bool { return r <= 0x20 })
		}
		match := pattern.FindStringSubmatch(candidate)
		if match == nil || match[0] != candidate {
			continue
		}
		if index == 1 {
			return SANYModuleLocation(match[1]), true
		}
		coordinates := [4]int{}
		valid := true
		for i := range coordinates {
			value, err := strconv.ParseInt(match[i+1], 10, 32)
			if err != nil {
				valid = false
				break
			}
			coordinates[i] = int(value)
		}
		if !valid {
			return Position{}, false
		}
		return NewSANYLocation(match[5], coordinates[0], coordinates[1], coordinates[2], coordinates[3]), true
	}
	return Position{}, false
}

func ParseSANYLocations(text string) []Position {
	var out []Position
	for _, pattern := range SANYAllLocationPatterns {
		for _, bounds := range pattern.FindAllStringIndex(text, -1) {
			if loc, ok := ParseSANYLocation(text[bounds[0]:bounds[1]]); ok {
				out = append(out, loc)
			}
		}
	}
	return out
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
