package tlc

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	ConfigKeywordConstant          = "CONSTANT"
	ConfigKeywordConstants         = "CONSTANTS"
	ConfigKeywordConstraint        = "CONSTRAINT"
	ConfigKeywordConstraints       = "CONSTRAINTS"
	ConfigKeywordActionConstraint  = "ACTION_CONSTRAINT"
	ConfigKeywordActionConstraints = "ACTION_CONSTRAINTS"
	ConfigKeywordInvariant         = "INVARIANT"
	ConfigKeywordInvariants        = "INVARIANTS"
	ConfigKeywordInit              = "INIT"
	ConfigKeywordNext              = "NEXT"
	ConfigKeywordView              = "VIEW"
	ConfigKeywordSymmetry          = "SYMMETRY"
	ConfigKeywordSpec              = "SPECIFICATION"
	ConfigKeywordProperty          = "PROPERTY"
	ConfigKeywordProperties        = "PROPERTIES"
	ConfigKeywordAlias             = "ALIAS"
	ConfigKeywordPostCondition     = "POSTCONDITION"
	ConfigKeywordPostConditions    = "POSTCONDITIONS"
	ConfigKeywordPeriodic          = "_PERIODIC"
	ConfigKeywordRLReward          = "_RL_REWARD"
	ConfigKeywordPossible          = "_POSSIBLE"
	ConfigKeywordCheckDeadlock     = "CHECK_DEADLOCK"
)

const modelConfigNoSymmetryProperty = "tlc2.tool.impl.ModelConfig.nosymmetry"

var configKeywords = map[string]struct{}{
	ConfigKeywordConstant:          {},
	ConfigKeywordConstants:         {},
	ConfigKeywordConstraint:        {},
	ConfigKeywordConstraints:       {},
	ConfigKeywordActionConstraint:  {},
	ConfigKeywordActionConstraints: {},
	ConfigKeywordInvariant:         {},
	ConfigKeywordInvariants:        {},
	ConfigKeywordInit:              {},
	ConfigKeywordNext:              {},
	ConfigKeywordView:              {},
	ConfigKeywordSymmetry:          {},
	ConfigKeywordSpec:              {},
	ConfigKeywordProperty:          {},
	ConfigKeywordProperties:        {},
	ConfigKeywordAlias:             {},
	ConfigKeywordPostCondition:     {},
	ConfigKeywordPostConditions:    {},
	ConfigKeywordPeriodic:          {},
	ConfigKeywordRLReward:          {},
	ConfigKeywordPossible:          {},
	ConfigKeywordCheckDeadlock:     {},
}

type ConfigFileError struct {
	Code   int
	Params []string
}

func (e *ConfigFileError) Error() string {
	if e == nil {
		return ""
	}
	switch e.Code {
	case ECCFGMissingID:
		return fmt.Sprintf("config line %s: missing identifier after %s", e.param(0), e.param(1))
	case ECCFGTwiceKeyword:
		return fmt.Sprintf("config line %s: keyword %s occurs more than once", e.param(0), e.param(1))
	case ECCFGExpectID:
		return fmt.Sprintf("config line %s: expected identifier after %s", e.param(0), e.param(1))
	case ECCFGExpectedSymbol:
		return fmt.Sprintf("config line %s: expected %s", e.param(0), e.param(1))
	case ECCFGErrorReadingFile:
		return fmt.Sprintf("cannot read config file %s: %s", e.param(0), e.param(1))
	default:
		if len(e.Params) != 0 {
			return fmt.Sprintf("config error %d: %s", e.Code, strings.Join(e.Params, ", "))
		}
		return fmt.Sprintf("config error %d", e.Code)
	}
}

func (e *ConfigFileError) param(i int) string {
	if i >= 0 && i < len(e.Params) {
		return e.Params[i]
	}
	return ""
}

func newConfigFileError(code int, line int, params ...string) *ConfigFileError {
	all := make([]string, 0, len(params)+1)
	all = append(all, strconv.Itoa(line))
	all = append(all, params...)
	return &ConfigFileError{Code: code, Params: all}
}

type ConfigConstant struct {
	Name  string
	Args  []Value
	Value Value
}

type ConfigConstants struct {
	data []ConfigConstant
}

func NewConfigConstants() *ConfigConstants {
	return &ConfigConstants{}
}

func (c *ConfigConstants) Add(constant ConfigConstant) {
	c.data = append(c.data, constant)
}

func (c *ConfigConstants) Len() int {
	if c == nil {
		return 0
	}
	return len(c.data)
}

func (c *ConfigConstants) At(i int) ConfigConstant {
	return c.data[i]
}

func (c *ConfigConstants) All() []ConfigConstant {
	if c == nil || len(c.data) == 0 {
		return nil
	}
	out := make([]ConfigConstant, len(c.data))
	copy(out, c.data)
	return out
}

type ModelConfig struct {
	configFileName string

	constants         *ConfigConstants
	modConstants      *InsMap[string, *ConfigConstants]
	overrides         *InsMap[string, string]
	overridesReverse  *InsMap[string, string]
	modOverrides      *InsMap[string, *InsMap[string, string]]
	rawConstants      []string
	constantsAsList   [][]string
	constraints       []string
	actionConstraints []string
	invariants        []string
	properties        []string
	possible          []string
	postConditions    []string

	init          string
	next          string
	view          string
	symmetry      string
	spec          string
	alias         string
	periodic      string
	rlReward      string
	checkDeadlock bool
	deadlockSet   bool
}

func NewModelConfig(configFileName string) *ModelConfig {
	return newModelConfig(configFileName, true)
}

func newModelConfig(configFileName string, resetModelValues bool) *ModelConfig {
	if resetModelValues {
		ModelValueInit()
	}
	return &ModelConfig{
		configFileName:   configFileName,
		constants:        NewConfigConstants(),
		modConstants:     NewInsMap[string, *ConfigConstants](),
		overrides:        NewInsMap[string, string](),
		overridesReverse: NewInsMap[string, string](),
		modOverrides:     NewInsMap[string, *InsMap[string, string]](),
		checkDeadlock:    true,
	}
}

func ParseModelConfigSource(file, source string) (*ModelConfig, error) {
	cfg := NewModelConfig(file)
	parser := &modelConfigParser{
		cfg: cfg,
		lex: newModelConfigLexer(file, source),
	}
	if err := parser.parse(); err != nil {
		return nil, err
	}
	SetModelValues()
	return cfg, nil
}

func ModelConfigPath(configFile string) string {
	if strings.HasSuffix(configFile, ".tla") || strings.HasSuffix(configFile, ".cfg") {
		return configFile
	}
	return configFile + ".cfg"
}

func ParseModelConfigFile(configFile string) (*ModelConfig, error) {
	path := ModelConfigPath(configFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &ConfigFileError{Code: ECCFGErrorReadingFile, Params: []string{path, err.Error()}}
	}
	source := string(data)
	if strings.HasSuffix(path, ".tla") {
		name := strings.TrimSuffix(filepath.Base(path), ".tla")
		source = ExtractMonolithConfigSource(source, name)
	}
	return ParseModelConfigSource(path, source)
}

func ExtractMonolithConfigSource(source string, configName string) string {
	var out strings.Builder
	active := false
	for _, rawLine := range strings.Split(source, "\n") {
		line := strings.TrimSuffix(rawLine, "\r")
		if active && strings.HasPrefix(line, "====") {
			break
		}
		if !active && isMonolithConfigStart(line, configName) {
			active = true
			continue
		}
		if active {
			out.WriteString(line)
			out.WriteByte('\n')
		}
	}
	return strings.TrimSpace(out.String())
}

func isMonolithConfigStart(line string, configName string) bool {
	text := strings.TrimSpace(line)
	if !strings.HasPrefix(text, "----") {
		return false
	}
	text = strings.TrimLeft(text, "-")
	text = strings.TrimSpace(text)
	fields := strings.Fields(text)
	if len(fields) < 2 || fields[0] != "CONFIG" || fields[1] != configName {
		return false
	}
	for _, field := range fields[2:] {
		if strings.Trim(field, "-") != "" {
			return false
		}
	}
	return true
}

func (m *ModelConfig) GetRawConstants() []string {
	if len(m.rawConstants) == 0 {
		return nil
	}
	out := make([]string, len(m.rawConstants))
	copy(out, m.rawConstants)
	return out
}

func (m *ModelConfig) GetConstantsAsList() [][]string {
	if len(m.constantsAsList) == 0 {
		return nil
	}
	out := make([][]string, len(m.constantsAsList))
	for i := range m.constantsAsList {
		out[i] = append([]string(nil), m.constantsAsList[i]...)
	}
	return out
}

func (m *ModelConfig) GetConstants() *ConfigConstants {
	return m.constants
}

func (m *ModelConfig) GetModConstants() *InsMap[string, *ConfigConstants] {
	return m.modConstants
}

func (m *ModelConfig) GetOverrides() *InsMap[string, string] {
	return m.overrides
}

func (m *ModelConfig) GetOverridenSpecNameForConfigName(configName string) string {
	return m.overridesReverse.Get(configName)
}

func (m *ModelConfig) GetModOverrides() *InsMap[string, *InsMap[string, string]] {
	return m.modOverrides
}

func (m *ModelConfig) GetConstraints() []string {
	return append([]string(nil), m.constraints...)
}

func (m *ModelConfig) GetActionConstraints() []string {
	return append([]string(nil), m.actionConstraints...)
}

func (m *ModelConfig) GetInit() string {
	return m.init
}

func (m *ModelConfig) GetNext() string {
	return m.next
}

func (m *ModelConfig) GetView() string {
	return m.view
}

func (m *ModelConfig) ConfigDefinesSpecification() bool {
	return strings.TrimSpace(m.spec) != ""
}

func (m *ModelConfig) GetSymmetry() string {
	if value, ok := tlcLookupSystemProperty(modelConfigNoSymmetryProperty); ok && javaBooleanProperty(value) {
		return ""
	}
	return m.symmetry
}

func (m *ModelConfig) GetInvariants() []string {
	return append([]string(nil), m.invariants...)
}

func (m *ModelConfig) GetPossible() []string {
	return append([]string(nil), m.possible...)
}

func (m *ModelConfig) GetSpec() string {
	return m.spec
}

func (m *ModelConfig) GetProperties() []string {
	return append([]string(nil), m.properties...)
}

func (m *ModelConfig) GetAlias() string {
	return m.alias
}

func (m *ModelConfig) GetPostConditions() []string {
	return append([]string(nil), m.postConditions...)
}

func (m *ModelConfig) GetPeriodic() string {
	return m.periodic
}

func (m *ModelConfig) GetRLReward() string {
	return m.rlReward
}

func (m *ModelConfig) GetCheckDeadlock() bool {
	return m.checkDeadlock
}

type modelConfigParser struct {
	cfg    *ModelConfig
	lex    *modelConfigLexer
	pushed *modelConfigToken
}

func (p *modelConfigParser) parse() error {
	for {
		tok := p.next()
		if tok.kind == configTokenEOF {
			return nil
		}
		switch tok.image {
		case ConfigKeywordInit:
			if err := p.parseSingleton(tok, &p.cfg.init); err != nil {
				return err
			}
		case ConfigKeywordNext:
			if err := p.parseSingleton(tok, &p.cfg.next); err != nil {
				return err
			}
		case ConfigKeywordSpec:
			if err := p.parseSingleton(tok, &p.cfg.spec); err != nil {
				return err
			}
		case ConfigKeywordView:
			if err := p.parseSingleton(tok, &p.cfg.view); err != nil {
				return err
			}
		case ConfigKeywordSymmetry:
			if err := p.parseSingleton(tok, &p.cfg.symmetry); err != nil {
				return err
			}
		case ConfigKeywordAlias:
			if err := p.parseSingleton(tok, &p.cfg.alias); err != nil {
				return err
			}
		case ConfigKeywordPeriodic:
			if err := p.parseSingleton(tok, &p.cfg.periodic); err != nil {
				return err
			}
		case ConfigKeywordRLReward:
			if err := p.parseSingleton(tok, &p.cfg.rlReward); err != nil {
				return err
			}
		case ConfigKeywordPostCondition, ConfigKeywordPostConditions:
			p.parseList(&p.cfg.postConditions)
		case ConfigKeywordConstant, ConfigKeywordConstants:
			if err := p.parseConstants(tok); err != nil {
				return err
			}
		case ConfigKeywordInvariant, ConfigKeywordInvariants:
			p.parseList(&p.cfg.invariants)
		case ConfigKeywordProperty, ConfigKeywordProperties:
			p.parseList(&p.cfg.properties)
		case ConfigKeywordConstraint, ConfigKeywordConstraints:
			p.parseList(&p.cfg.constraints)
		case ConfigKeywordActionConstraint, ConfigKeywordActionConstraints:
			p.parseList(&p.cfg.actionConstraints)
		case ConfigKeywordPossible:
			p.parseList(&p.cfg.possible)
		case ConfigKeywordCheckDeadlock:
			if err := p.parseCheckDeadlock(tok); err != nil {
				return err
			}
		default:
			return newConfigFileError(ECCFGExpectedSymbol, tok.line, "a keyword")
		}
	}
}

func (p *modelConfigParser) parseSingleton(keyword modelConfigToken, dst *string) error {
	tok := p.next()
	if tok.kind == configTokenEOF {
		return newConfigFileError(ECCFGMissingID, keyword.line, keyword.image)
	}
	if *dst != "" {
		return newConfigFileError(ECCFGTwiceKeyword, keyword.line, keyword.image)
	}
	*dst = tok.image
	return nil
}

func (p *modelConfigParser) parseList(dst *[]string) {
	for {
		tok := p.next()
		if tok.kind == configTokenEOF {
			return
		}
		if isConfigKeyword(tok.image) {
			p.push(tok)
			return
		}
		*dst = append(*dst, tok.image)
	}
}

func (p *modelConfigParser) parseCheckDeadlock(keyword modelConfigToken) error {
	tok := p.next()
	if tok.kind == configTokenEOF {
		return newConfigFileError(ECCFGMissingID, keyword.line, keyword.image)
	}
	if p.cfg.deadlockSet {
		return newConfigFileError(ECCFGTwiceKeyword, keyword.line, keyword.image)
	}
	switch tok.image {
	case "TRUE":
		p.cfg.checkDeadlock = true
	case "FALSE":
		p.cfg.checkDeadlock = false
	default:
		return newConfigFileError(ECCFGExpectedSymbol, tok.line, "TRUE or FALSE")
	}
	p.cfg.deadlockSet = true
	return nil
}

func (p *modelConfigParser) parseConstants(keyword modelConfigToken) error {
	rawLines := []string{}
	for {
		tok := p.next()
		if tok.kind == configTokenEOF {
			break
		}
		if isConfigKeyword(tok.image) {
			p.push(tok)
			break
		}

		name := tok.image
		next := p.next()
		for next.image == "!" {
			part := p.next()
			if part.kind == configTokenEOF {
				return newConfigFileError(ECCFGExpectID, part.line, "!")
			}
			name += "!" + part.image
			next = p.next()
		}

		args := []Value(nil)
		if next.image == "(" {
			parsedArgs, afterArgs, err := p.parseConstantArgs(next.line)
			if err != nil {
				return err
			}
			args = parsedArgs
			next = afterArgs
		}

		lhs := formatConfigConstantLHS(name, args)
		line := ConfigConstant{Name: name, Args: args}
		if next.image == "<-" {
			target := p.next()
			if target.image == "[" {
				modName := p.next()
				if modName.kind == configTokenEOF {
					return newConfigFileError(ECCFGExpectID, target.line, "<-[")
				}
				close := p.next()
				if close.image != "]" {
					return newConfigFileError(ECCFGExpectedSymbol, close.line, "]")
				}
				replacement := p.next()
				if replacement.kind == configTokenEOF {
					return newConfigFileError(ECCFGExpectID, replacement.line, "<-[mod]")
				}
				defs, ok := p.cfg.modOverrides.Get2(modName.image)
				if !ok {
					defs = NewInsMap[string, string]()
					p.cfg.modOverrides.Set(modName.image, defs)
				}
				defs.Set(name, replacement.image)
				rawLines = append(rawLines, lhs+" <- ["+modName.image+"] "+replacement.image)
			} else {
				if target.kind == configTokenEOF {
					return newConfigFileError(ECCFGExpectID, target.line, "<-")
				}
				p.cfg.overrides.Set(name, target.image)
				p.cfg.overridesReverse.Set(target.image, name)
				entry := lhs + " <- " + target.image
				p.cfg.constantsAsList = append(p.cfg.constantsAsList, []string{entry})
				rawLines = append(rawLines, entry)
			}
			continue
		}

		if next.image != "=" {
			return newConfigFileError(ECCFGExpectedSymbol, next.line, "= or <-")
		}
		valueToken := p.next()
		if valueToken.image == "[" {
			modName := p.next()
			if modName.kind == configTokenEOF {
				return newConfigFileError(ECCFGExpectID, valueToken.line, "=[")
			}
			close := p.next()
			if close.image != "]" {
				return newConfigFileError(ECCFGExpectedSymbol, close.line, "]")
			}
			valueToken = p.next()
			value, err := p.parseValue(valueToken)
			if err != nil {
				return err
			}
			line.Value = value
			constants, ok := p.cfg.modConstants.Get2(modName.image)
			if !ok {
				constants = NewConfigConstants()
				p.cfg.modConstants.Set(modName.image, constants)
			}
			constants.Add(line)
			rawLines = append(rawLines, lhs+" = ["+modName.image+"] "+value.String())
			continue
		}
		value, err := p.parseValue(valueToken)
		if err != nil {
			return err
		}
		line.Value = value
		p.cfg.constants.Add(line)
		entry := []string{lhs, value.String()}
		p.cfg.constantsAsList = append(p.cfg.constantsAsList, entry)
		rawLines = append(rawLines, lhs+" = "+value.String())
	}
	if len(rawLines) == 0 {
		p.cfg.rawConstants = append(p.cfg.rawConstants, keyword.image)
		return nil
	}
	p.cfg.rawConstants = append(p.cfg.rawConstants, keyword.image+"\n"+strings.Join(rawLines, "\n"))
	return nil
}

func (p *modelConfigParser) parseConstantArgs(line int) ([]Value, modelConfigToken, error) {
	tok := p.next()
	if tok.image == ")" {
		return nil, p.next(), nil
	}
	args := []Value{}
	for {
		arg, err := p.parseValue(tok)
		if err != nil {
			return nil, modelConfigToken{}, err
		}
		args = append(args, arg)
		tok = p.next()
		if tok.image != "," {
			break
		}
		tok = p.next()
	}
	if tok.image != ")" {
		return nil, modelConfigToken{}, newConfigFileError(ECCFGGeneral, line)
	}
	return args, p.next(), nil
}

func (p *modelConfigParser) parseValue(tok modelConfigToken) (Value, error) {
	switch {
	case tok.kind == configTokenNumber:
		val, err := strconv.ParseInt(tok.image, 10, 32)
		if err != nil {
			return nil, newConfigFileError(ECCFGExpectedSymbol, tok.line, "a value")
		}
		return NewIntValue(int32(val)), nil
	case tok.kind == configTokenString:
		return NewStringValue(reduceConfigString(tok.image)), nil
	case tok.image == "TRUE":
		return BoolTrue, nil
	case tok.image == "FALSE":
		return BoolFalse, nil
	case tok.image == "{":
		return p.parseSetValue(tok.line)
	case tok.kind != configTokenEOF:
		return MakeModelValue(tok.image), nil
	default:
		return nil, newConfigFileError(ECCFGExpectedSymbol, tok.line, "a value")
	}
}

func (p *modelConfigParser) parseSetValue(line int) (Value, error) {
	values := NewValueVec(0)
	tok := p.next()
	if tok.image != "}" {
		for {
			value, err := p.parseValue(tok)
			if err != nil {
				return nil, err
			}
			values.Add(value)
			tok = p.next()
			if tok.image != "," {
				break
			}
			tok = p.next()
		}
	}
	if tok.image != "}" {
		return nil, newConfigFileError(ECCFGExpectedSymbol, line, "}")
	}
	return NewSetEnumValueVec(values, false), nil
}

func (p *modelConfigParser) next() modelConfigToken {
	if p.pushed != nil {
		tok := *p.pushed
		p.pushed = nil
		return tok
	}
	return p.lex.next()
}

func (p *modelConfigParser) push(tok modelConfigToken) {
	p.pushed = &tok
}

func isConfigKeyword(image string) bool {
	_, ok := configKeywords[image]
	return ok
}

func formatConfigConstantLHS(name string, args []Value) string {
	if len(args) == 0 {
		return name
	}
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = compactConfigValueString(arg)
	}
	return name + "(" + strings.Join(parts, ",") + ")"
}

func compactConfigValueString(value Value) string {
	if value == nil {
		return ""
	}
	return strings.ReplaceAll(value.String(), ", ", ",")
}

func reduceConfigString(image string) string {
	value, err := strconv.Unquote(image)
	if err == nil {
		return value
	}
	if len(image) >= 2 && image[0] == '"' && image[len(image)-1] == '"' {
		return image[1 : len(image)-1]
	}
	return image
}

type configTokenKind byte

const (
	configTokenEOF configTokenKind = iota
	configTokenIdentifier
	configTokenNumber
	configTokenString
	configTokenSymbol
)

type modelConfigToken struct {
	kind  configTokenKind
	image string
	line  int
	col   int
}

type modelConfigLexer struct {
	file   string
	input  string
	offset int
	line   int
	col    int
}

func newModelConfigLexer(file, input string) *modelConfigLexer {
	return &modelConfigLexer{file: file, input: input, line: 1, col: 1}
}

func (l *modelConfigLexer) next() modelConfigToken {
	l.skipWhitespaceAndComments()
	if l.eof() {
		return modelConfigToken{kind: configTokenEOF, line: l.line, col: l.col}
	}
	line, col := l.line, l.col
	rest := l.input[l.offset:]
	if strings.HasPrefix(rest, "<-") {
		l.advanceByte()
		l.advanceByte()
		return modelConfigToken{kind: configTokenSymbol, image: "<-", line: line, col: col}
	}
	ch := l.peekByte()
	switch {
	case ch == '"':
		return l.stringToken()
	case isConfigNumberStart(ch):
		start := l.offset
		for !l.eof() && isConfigDigit(l.peekByte()) {
			l.advanceByte()
		}
		return modelConfigToken{kind: configTokenNumber, image: l.input[start:l.offset], line: line, col: col}
	case isConfigIdentStart(ch):
		start := l.offset
		for !l.eof() && isConfigIdentPart(l.peekByte()) {
			l.advanceByte()
		}
		return modelConfigToken{kind: configTokenIdentifier, image: l.input[start:l.offset], line: line, col: col}
	default:
		l.advanceByte()
		return modelConfigToken{kind: configTokenSymbol, image: string(ch), line: line, col: col}
	}
}

func (l *modelConfigLexer) stringToken() modelConfigToken {
	line, col := l.line, l.col
	start := l.offset
	l.advanceByte()
	for !l.eof() {
		ch := l.peekByte()
		l.advanceByte()
		if ch == '\\' && !l.eof() {
			l.advanceByte()
			continue
		}
		if ch == '"' {
			break
		}
	}
	return modelConfigToken{kind: configTokenString, image: l.input[start:l.offset], line: line, col: col}
}

func (l *modelConfigLexer) skipWhitespaceAndComments() {
	for {
		for !l.eof() {
			ch := l.peekByte()
			if ch != ' ' && ch != '\t' && ch != '\r' && ch != '\n' && ch != '\f' {
				break
			}
			l.advanceByte()
		}
		if l.eof() {
			return
		}
		rest := l.input[l.offset:]
		switch {
		case strings.HasPrefix(rest, "\\*"):
			for !l.eof() && l.peekByte() != '\n' {
				l.advanceByte()
			}
		case strings.HasPrefix(rest, "(*"):
			l.skipBlockComment()
		default:
			return
		}
	}
}

func (l *modelConfigLexer) skipBlockComment() {
	depth := 0
	for !l.eof() {
		rest := l.input[l.offset:]
		switch {
		case strings.HasPrefix(rest, "(*"):
			depth++
			l.advanceByte()
			l.advanceByte()
		case strings.HasPrefix(rest, "*)"):
			depth--
			l.advanceByte()
			l.advanceByte()
			if depth == 0 {
				return
			}
		default:
			l.advanceByte()
		}
	}
}

func (l *modelConfigLexer) eof() bool {
	return l.offset >= len(l.input)
}

func (l *modelConfigLexer) peekByte() byte {
	return l.input[l.offset]
}

func (l *modelConfigLexer) advanceByte() {
	if l.eof() {
		return
	}
	if l.input[l.offset] == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	l.offset++
}

func isConfigNumberStart(ch byte) bool {
	return isConfigDigit(ch)
}

func isConfigDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isConfigIdentStart(ch byte) bool {
	return ch == '_' || ch == '$' || (ch >= 'A' && ch <= 'Z') || (ch >= 'a' && ch <= 'z')
}

func isConfigIdentPart(ch byte) bool {
	return isConfigIdentStart(ch) || isConfigDigit(ch)
}
