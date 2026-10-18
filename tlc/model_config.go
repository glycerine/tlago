package tlc

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
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

// ModelConfigAllKeywords is Java ModelConfig.ALL_KEYWORDS. Like the source
// array, callers receive and can mutate the published backing storage.
var ModelConfigAllKeywords = []string{
	ConfigKeywordConstant,
	ConfigKeywordConstants,
	ConfigKeywordConstraint,
	ConfigKeywordConstraints,
	ConfigKeywordActionConstraint,
	ConfigKeywordActionConstraints,
	ConfigKeywordInvariant,
	ConfigKeywordInvariants,
	ConfigKeywordInit,
	ConfigKeywordNext,
	ConfigKeywordView,
	ConfigKeywordSymmetry,
	ConfigKeywordSpec,
	ConfigKeywordProperty,
	ConfigKeywordProperties,
	ConfigKeywordAlias,
	ConfigKeywordPostCondition,
	ConfigKeywordPostConditions,
	ConfigKeywordPeriodic,
	ConfigKeywordRLReward,
	ConfigKeywordPossible,
	ConfigKeywordCheckDeadlock,
}

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
		return fmt.Sprintf("TLC found an error in the configuration file at line %s\nThe keyword %s was not followed by an identifier.", e.param(0), e.param(1))
	case ECCFGTwiceKeyword:
		return fmt.Sprintf("TLC found an error in the configuration file at line %s\nThe keyword %s appeared twice.", e.param(0), e.param(1))
	case ECCFGExpectID:
		return fmt.Sprintf("TLC found an error in the configuration file at line %s\nExpected an identifier after %s.", e.param(0), e.param(1))
	case ECCFGGeneral:
		return fmt.Sprintf("TLC found an error in the configuration file at line %s\n", e.param(0))
	case ECCFGExpectedSymbol:
		return fmt.Sprintf("TLC found an error in the configuration file at line %s\nIt was expecting %s, but did not find it.", e.param(0), e.param(1))
	case ECCFGErrorReadingFile:
		return fmt.Sprintf("TLC encountered the following error when trying to read the configuration file %s:\n%s", e.param(0), e.param(1))
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
	code = int(int32(code))
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
	mu             sync.Mutex
	configFileName string
	resolver       FilenameToStream

	constants         *ConfigConstants
	modConstants      *InsMap[string, *ConfigConstants]
	overrides         *InsMap[string, string]
	overridesReverse  *InsMap[string, string]
	modOverrides      *InsMap[string, *InsMap[string, string]]
	rawConstants      []string
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

func NewModelConfig(configFileName string, resolvers ...FilenameToStream) *ModelConfig {
	if len(resolvers) > 1 {
		panic(NewIllegalArgumentException("NewModelConfig accepts at most one filename resolver"))
	}
	var resolver FilenameToStream
	if len(resolvers) == 1 {
		resolver = resolvers[0]
	}
	return newModelConfigWithResolver(configFileName, resolver, true)
}

func newModelConfig(configFileName string, resetModelValues bool) *ModelConfig {
	return newModelConfigWithResolver(configFileName, nil, resetModelValues)
}

func newModelConfigWithResolver(configFileName string, resolver FilenameToStream, resetModelValues bool) *ModelConfig {
	if isNil(resolver) {
		// Java constructs the default resolver here, capturing its user and
		// library search paths before a later parse call.
		resolver = NewSimpleFilenameToStream(nil)
	}
	if resetModelValues {
		ModelValueInit()
	}
	return &ModelConfig{
		configFileName:    configFileName,
		resolver:          resolver,
		constants:         NewConfigConstants(),
		modConstants:      NewInsMap[string, *ConfigConstants](),
		overrides:         NewInsMap[string, string](),
		overridesReverse:  NewInsMap[string, string](),
		modOverrides:      NewInsMap[string, *InsMap[string, string]](),
		rawConstants:      make([]string, 0),
		constraints:       make([]string, 0),
		actionConstraints: make([]string, 0),
		invariants:        make([]string, 0),
		properties:        make([]string, 0),
		possible:          make([]string, 0),
		postConditions:    make([]string, 0),
		checkDeadlock:     true,
	}
}

func ParseModelConfigSource(file, source string) (*ModelConfig, error) {
	cfg := NewModelConfig(file)
	if err := cfg.parseSource(source); err != nil {
		return nil, err
	}
	SetModelValues()
	return cfg, nil
}

func (m *ModelConfig) parseSource(source string) error {
	m.require()
	parser := &modelConfigParser{
		cfg:          m,
		lex:          newModelConfigLexer(m.configFileName, source),
		rawConstants: make([]string, 0),
	}
	if err := parser.parse(); err != nil {
		return err
	}
	return nil
}

func ModelConfigPath(configFile string) string {
	if strings.HasSuffix(configFile, ".tla") || strings.HasSuffix(configFile, ".cfg") {
		return configFile
	}
	return configFile + ".cfg"
}

func ParseModelConfigFile(configFile string) (*ModelConfig, error) {
	path := ModelConfigPath(configFile)
	cfg := NewModelConfig(path)
	if err := cfg.Parse(); err != nil {
		return nil, err
	}
	SetModelValues()
	return cfg, nil
}

// MonolithSpecExtractor.config owns its BufferedReader and therefore the
// FileInputStream supplied by ModelConfig. Plain configuration parsing leaves
// its stream open through the token manager, as the source does.
func modelConfigReadMonolith(file *os.File, configName string) (data []byte, err error) {
	defer func() {
		failure := recover()
		closeErr := bufferedRandomAccessFileIOError(file.Close())
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if err == nil {
			if closeErr != nil {
				data = nil
				err = closeErr
			}
		} else {
			javaSuppressCloseError(err, closeErr)
		}
	}()
	reader, err := newJavaReplacingLineReader(file, javaDefaultCharset())
	if err != nil {
		return nil, bufferedRandomAccessFileIOError(err)
	}
	active := false
	var config strings.Builder
	for {
		line, readErr := reader.ReadLine()
		if readErr != nil {
			return nil, bufferedRandomAccessFileIOError(readErr)
		}
		if line == nil {
			break
		}
		if active && monolithEnd(*line) {
			break
		}
		if !active && isMonolithConfigStart(*line, configName) {
			active = true
			continue
		}
		if active {
			config.WriteString(*line)
			config.WriteByte('\n')
		}
	}
	text := strings.TrimFunc(config.String(), func(r rune) bool { return r <= 0x20 })
	data, err = javaCharsetEncode(text, "UTF-8", false)
	return data, err
}

func (m *ModelConfig) Parse() error {
	m.require()
	if isNil(m.resolver) {
		m.resolver = NewSimpleFilenameToStream(nil)
	}
	// ModelConfig receives the name after MonolithSpecExtractor.getConfig has
	// applied any caller-level .cfg default.  Its instance parse method passes
	// that stored name to the resolver unchanged.
	path := m.configFileName
	resolved := m.resolver.Resolve(path, false)
	if resolved == nil {
		return &ConfigFileError{Code: ECCFGErrorReadingFile, Params: []string{path, "File not found."}}
	}
	resolvedPath := resolved.GetPath()
	info, err := os.Stat(resolvedPath)
	if err != nil {
		return &ConfigFileError{Code: ECCFGErrorReadingFile, Params: []string{path, "File not found."}}
	}
	if info.IsDir() {
		fmt.Fprintln(os.Stdout, "***Internal error: Unable to create FileInputStream")
		return &ConfigFileError{Code: ECCFGErrorReadingFile, Params: []string{path, "File not found."}}
	}
	file, err := os.Open(resolvedPath)
	if err != nil {
		fmt.Fprintln(os.Stdout, "***Internal error: Unable to create FileInputStream")
		return &ConfigFileError{Code: ECCFGErrorReadingFile, Params: []string{path, "File not found."}}
	}
	monolith := strings.HasSuffix(path, ".tla")
	var data []byte
	if monolith {
		data, err = modelConfigReadMonolith(file, strings.ReplaceAll(path, ".tla", ""))
	} else {
		data, err = io.ReadAll(file)
	}
	if err != nil {
		return &ConfigFileError{Code: ECCFGErrorReadingFile, Params: []string{path, distributedIOMessage(err)}}
	}
	charset := javaDefaultCharset()
	source, err := javaCharsetDecode(data, charset, true)
	if err != nil {
		return &ConfigFileError{Code: ECCFGErrorReadingFile, Params: []string{path, err.Error()}}
	}
	return m.parseSource(source)
}

func ExtractMonolithConfigSource(source string, configName string) string {
	var out strings.Builder
	active := false
	for _, line := range monolithLines(source) {
		if active && monolithEnd(line) {
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
	return strings.TrimFunc(out.String(), func(r rune) bool { return r <= 0x20 })
}

func isMonolithConfigStart(line string, configName string) bool {
	pattern := `^-{4,}` + monolithWhitespace + `*CONFIG` + monolithWhitespace + `+` + regexp.QuoteMeta(configName) + monolithWhitespace + `*-{4,}$`
	return regexp.MustCompile(pattern).MatchString(line)
}

func (m *ModelConfig) GetRawConstants() []string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.rawConstants
}

func (m *ModelConfig) GetConstantsAsList() [][]string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	constants := constantsAsListFromRaw(m.rawConstants)
	return constants
}

func (m *ModelConfig) GetConstants() *ConfigConstants {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.constants
}

func (m *ModelConfig) GetModConstants() *InsMap[string, *ConfigConstants] {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.modConstants
}

func (m *ModelConfig) GetOverrides() *InsMap[string, string] {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.overrides
}

func (m *ModelConfig) GetOverridenSpecNameForConfigName(configName string) string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	value, found := m.overridesReverse.Get2(configName)
	if !found {
		return ""
	}
	return value
}

func (m *ModelConfig) GetOverridenSpecNameForConfigNameNullable(configName *string) *string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	if configName == nil {
		panic(NewNullPointerException())
	}
	value, found := m.overridesReverse.Get2(*configName)
	if !found {
		return nil
	}
	return javaString(value)
}

func (m *ModelConfig) GetModOverrides() *InsMap[string, *InsMap[string, string]] {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.modOverrides
}

func (m *ModelConfig) GetConstraints() []string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.constraints
}

func (m *ModelConfig) GetActionConstraints() []string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.actionConstraints
}

func (m *ModelConfig) GetInit() string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.init
}

func (m *ModelConfig) GetNext() string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.next
}

func (m *ModelConfig) GetView() string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.view
}

func (m *ModelConfig) ConfigDefinesSpecification() bool {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return strings.TrimFunc(m.spec, func(r rune) bool { return r <= ' ' }) != ""
}

func (m *ModelConfig) GetSymmetry() string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	if value, ok := tlcLookupSystemProperty(modelConfigNoSymmetryProperty); ok && javaBooleanProperty(value) {
		return ""
	}
	return m.symmetry
}

func (m *ModelConfig) GetInvariants() []string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.invariants
}

func (m *ModelConfig) GetPossible() []string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.possible
}

func (m *ModelConfig) GetSpec() string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.spec
}

func (m *ModelConfig) GetProperties() []string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.properties
}

func (m *ModelConfig) GetAlias() string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.alias
}

func (m *ModelConfig) GetPostConditions() []string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.postConditions
}

func (m *ModelConfig) GetPeriodic() string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.periodic
}

func (m *ModelConfig) GetRLReward() string {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.rlReward
}

func (m *ModelConfig) GetCheckDeadlock() bool {
	m.require()
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.checkDeadlock
}

func (m *ModelConfig) require() {
	if m == nil {
		panic(NewNullPointerException())
	}
}

type modelConfigParser struct {
	cfg          *ModelConfig
	lex          *modelConfigLexer
	pushed       *modelConfigToken
	rawConstants []string
}

func (p *modelConfigParser) parse() error {
	for {
		tok := p.next()
		if tok.kind == configTokenEOF {
			p.cfg.rawConstants = p.rawConstants
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
	old := *dst
	*dst = tok.image
	if old != "" {
		return newConfigFileError(ECCFGTwiceKeyword, keyword.line, keyword.image)
	}
	return nil
}

func (p *modelConfigParser) parseList(dst *[]string) {
	for {
		tok := p.next()
		if tok.kind == configTokenEOF {
			// Java retains tt from the list loop, so the outer parse loop sees
			// the same EOF instead of asking the token manager to continue after
			// a recovered lexical error.
			p.push(tok)
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
	parsed := false
	switch tok.image {
	case "TRUE":
		parsed = true
	case "FALSE":
		parsed = false
	default:
		return newConfigFileError(ECCFGExpectedSymbol, tok.line, "TRUE or FALSE")
	}
	wasSet := p.cfg.deadlockSet
	p.cfg.checkDeadlock = parsed
	p.cfg.deadlockSet = true
	if wasSet {
		return newConfigFileError(ECCFGTwiceKeyword, keyword.line, keyword.image)
	}
	return nil
}

func (p *modelConfigParser) parseConstants(keyword modelConfigToken) error {
	var raw strings.Builder
	raw.WriteString(keyword.image)
	for {
		tok := p.next()
		if tok.kind == configTokenEOF {
			p.push(tok)
			break
		}
		if isConfigKeyword(tok.image) {
			p.push(tok)
			break
		}

		raw.WriteByte('\n')
		appendConfigRawToken(&raw, tok)
		name := tok.image
		next := p.nextRaw(&raw)
		for modelConfigTokenImage(next) == "!" {
			part := p.nextRaw(&raw)
			if part.lexicalFailure {
				name += "!null"
			} else {
				name += "!" + part.image
			}
			next = p.nextRaw(&raw)
		}

		args := []Value(nil)
		if modelConfigTokenImage(next) == "(" {
			parsedArgs, afterArgs, err := p.parseConstantArgsRaw(keyword.line, &raw)
			if err != nil {
				return err
			}
			args = parsedArgs
			next = afterArgs
		}

		line := ConfigConstant{Name: name, Args: args}
		if modelConfigTokenImage(next) == "<-" {
			target := p.nextRaw(&raw)
			if modelConfigTokenImage(target) == "[" {
				modName := p.nextRaw(&raw)
				if modName.kind == configTokenEOF {
					return newConfigFileError(ECCFGExpectID, target.line, "<-[")
				}
				close := p.nextRaw(&raw)
				if modelConfigTokenImage(close) != "]" {
					return newConfigFileError(ECCFGExpectedSymbol, close.line, "]")
				}
				replacement := p.nextRaw(&raw)
				if replacement.kind == configTokenEOF {
					return newConfigFileError(ECCFGExpectID, replacement.line, "<-[mod]")
				}
				defs, ok := p.cfg.modOverrides.Get2(modName.image)
				if !ok {
					defs = NewInsMap[string, string]()
					p.cfg.modOverrides.Set(modName.image, defs)
				}
				defs.Set(name, replacement.image)
			} else {
				if target.kind == configTokenEOF {
					return newConfigFileError(ECCFGExpectID, target.line, "<-")
				}
				p.cfg.overrides.Set(name, target.image)
				p.cfg.overridesReverse.Set(target.image, name)
			}
			continue
		}

		if modelConfigTokenImage(next) != "=" {
			return newConfigFileError(ECCFGExpectedSymbol, next.line, "= or <-")
		}
		valueToken := p.nextRaw(&raw)
		if modelConfigTokenImage(valueToken) == "[" {
			modName := p.nextRaw(&raw)
			if modName.kind == configTokenEOF {
				return newConfigFileError(ECCFGExpectID, valueToken.line, "=[")
			}
			close := p.nextRaw(&raw)
			if modelConfigTokenImage(close) != "]" {
				return newConfigFileError(ECCFGExpectedSymbol, close.line, "]")
			}
			valueToken = p.nextRaw(&raw)
			value, err := p.parseValueRaw(valueToken, &raw)
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
			continue
		}
		value, err := p.parseValueRaw(valueToken, &raw)
		if err != nil {
			return err
		}
		line.Value = value
		p.cfg.constants.Add(line)
	}
	p.rawConstants = append(p.rawConstants, raw.String())
	return nil
}

func (p *modelConfigParser) parseConstantArgs(line int) ([]Value, modelConfigToken, error) {
	return p.parseConstantArgsRaw(line, nil)
}

func (p *modelConfigParser) parseConstantArgsRaw(line int, raw *strings.Builder) ([]Value, modelConfigToken, error) {
	tok := p.next()
	if raw != nil {
		appendConfigRawToken(raw, tok)
	}
	args := []Value{}
	for {
		arg, err := p.parseValueRaw(tok, raw)
		if err != nil {
			return nil, modelConfigToken{}, err
		}
		args = append(args, arg)
		tok = p.nextRaw(raw)
		if modelConfigTokenImage(tok) != "," {
			break
		}
		tok = p.nextRaw(raw)
	}
	if modelConfigTokenImage(tok) != ")" {
		return nil, modelConfigToken{}, newConfigFileError(ECCFGGeneral, line)
	}
	return args, p.nextRaw(raw), nil
}

func (p *modelConfigParser) parseValue(tok modelConfigToken) (Value, error) {
	return p.parseValueRaw(tok, nil)
}

func (p *modelConfigParser) parseValueRaw(tok modelConfigToken, raw *strings.Builder) (Value, error) {
	if tok.lexicalFailure {
		// ModelConfig's getNextToken converts TokenMgrError to a freshly
		// allocated EOF token whose null image is dereferenced by parseValue.
		panic(NewNullPointerException())
	}
	switch {
	case tok.kind == configTokenNumber:
		val, err := strconv.ParseInt(tok.image, 10, 32)
		if err != nil {
			panic(NewNumberFormatException(tok.image))
		}
		return NewIntValue(int32(val)), nil
	case tok.kind == configTokenString:
		return NewStringValue(reduceConfigString(tok.image)), nil
	case tok.image == "TRUE":
		return BoolTrue, nil
	case tok.image == "FALSE":
		return BoolFalse, nil
	case tok.image == "{":
		return p.parseSetValueRaw(tok.line, raw)
	case tok.kind != configTokenEOF:
		return MakeModelValue(tok.image), nil
	default:
		return nil, newConfigFileError(ECCFGExpectedSymbol, tok.line, "a value")
	}
}

func (p *modelConfigParser) parseSetValue(line int) (Value, error) {
	return p.parseSetValueRaw(line, nil)
}

func (p *modelConfigParser) parseSetValueRaw(line int, raw *strings.Builder) (Value, error) {
	values := NewValueVec(0)
	tok := p.nextRaw(raw)
	if modelConfigTokenImage(tok) != "}" {
		for {
			value, err := p.parseValueRaw(tok, raw)
			if err != nil {
				return nil, err
			}
			values.Add(value)
			tok = p.nextRaw(raw)
			if modelConfigTokenImage(tok) != "," {
				break
			}
			tok = p.nextRaw(raw)
		}
	}
	if modelConfigTokenImage(tok) != "}" {
		return nil, newConfigFileError(ECCFGExpectedSymbol, tok.line, "}")
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

func (p *modelConfigParser) nextRaw(raw *strings.Builder) modelConfigToken {
	tok := p.next()
	if raw != nil {
		appendConfigRawToken(raw, tok)
	}
	return tok
}

func (p *modelConfigParser) push(tok modelConfigToken) {
	p.pushed = &tok
}

func modelConfigTokenImage(tok modelConfigToken) string {
	if tok.lexicalFailure {
		// getNextToken recovers TokenMgrError with a fresh Java Token whose
		// image is null. String.equals calls on that image fail immediately.
		panic(NewNullPointerException())
	}
	return tok.image
}

func isConfigKeyword(image string) bool {
	_, ok := configKeywords[image]
	return ok
}

func appendConfigRawToken(raw *strings.Builder, tok modelConfigToken) {
	if raw == nil || tok.kind == configTokenEOF {
		return
	}
	raw.WriteString(tok.image)
	raw.WriteByte(' ')
}

func constantsAsListFromRaw(rawConstants []string) [][]string {
	out := make([][]string, 0)
	for _, raw := range rawConstants {
		for _, line := range javaSplitLiteral(raw, "\n") {
			line = strings.TrimFunc(line, func(r rune) bool { return r <= ' ' })
			if line == ConfigKeywordConstant || line == ConfigKeywordConstants {
				continue
			}
			out = append(out, javaSplitLiteral(line, " = "))
		}
	}
	return out
}

func reduceConfigString(image string) string {
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
	kind           configTokenKind
	image          string
	line           int
	col            int
	lexicalFailure bool
}

type modelConfigLexer struct {
	file       string
	input      string
	offset     int
	line       int
	col        int
	previousCR bool
}

func newModelConfigLexer(file, input string) *modelConfigLexer {
	return &modelConfigLexer{file: file, input: input, line: 1, col: 1}
}

func (l *modelConfigLexer) next() modelConfigToken {
	if l.skipWhitespaceAndComments() {
		return modelConfigToken{kind: configTokenEOF, line: l.line, col: l.col, lexicalFailure: true}
	}
	if l.eof() {
		return modelConfigToken{kind: configTokenEOF, line: l.line, col: l.col}
	}
	line, col := l.line, l.col
	rest := l.input[l.offset:]
	ch := l.peekByte()
	baseNumberLength := configBaseNumberLength(rest)
	underscoreIdentifierLength := configUnderscoreIdentifierLength(rest)
	junctionLength := configJunctionLength(rest)
	switch {
	case ch == '"':
		return l.stringToken()
	case strings.HasPrefix(rest, "ℕ") || strings.HasPrefix(rest, "ℤ") || strings.HasPrefix(rest, "ℝ"):
		start := l.offset
		for range len("ℕ") {
			l.advanceByte()
		}
		return modelConfigToken{kind: configTokenIdentifier, image: l.input[start:l.offset], line: line, col: col}
	case strings.HasPrefix(rest, "SF_") || strings.HasPrefix(rest, "WF_"):
		// CASE2 and CASE3 cannot absorb the underscore after SF/WF. The
		// dedicated literal wins and any following identifier is a new token.
		literal := rest[:3]
		for range len(literal) {
			l.advanceByte()
		}
		return modelConfigToken{kind: configTokenSymbol, image: literal, line: line, col: col}
	case junctionLength != 0:
		start := l.offset
		for range junctionLength {
			l.advanceByte()
		}
		return modelConfigToken{kind: configTokenSymbol, image: l.input[start:l.offset], line: line, col: col}
	case baseNumberLength != 0:
		start := l.offset
		for range baseNumberLength {
			l.advanceByte()
		}
		return modelConfigToken{kind: configTokenNumber, image: l.input[start:l.offset], line: line, col: col}
	case isConfigNumberStart(ch):
		start := l.offset
		// SANY identifiers can start with digits when the token contains a
		// letter. JavaCC selects that complete identifier over a numeral.
		end, hasLetter := start, false
		for end < len(l.input) && isConfigIdentPart(l.input[end]) {
			c := l.input[end]
			hasLetter = hasLetter || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z'
			end++
		}
		if hasLetter {
			for l.offset < end {
				l.advanceByte()
			}
			return modelConfigToken{kind: configTokenIdentifier, image: l.input[start:end], line: line, col: col}
		}
		for !l.eof() && isConfigDigit(l.peekByte()) {
			l.advanceByte()
		}
		return modelConfigToken{kind: configTokenNumber, image: l.input[start:l.offset], line: line, col: col}
	case ch == '_' && underscoreIdentifierLength != 0:
		start := l.offset
		for range underscoreIdentifierLength {
			l.advanceByte()
		}
		return modelConfigToken{kind: configTokenIdentifier, image: l.input[start:l.offset], line: line, col: col}
	case isConfigIdentStart(ch):
		start := l.offset
		for !l.eof() && isConfigIdentPart(l.peekByte()) {
			l.advanceByte()
		}
		return modelConfigToken{kind: configTokenIdentifier, image: l.input[start:l.offset], line: line, col: col}
	default:
		// Java ModelConfig reads SANY tokens, so operator names and override
		// targets use the complete longest-match literal vocabulary.
		for _, literal := range modelConfigLiteralTokens {
			if strings.HasPrefix(rest, literal) {
				for range len(literal) {
					l.advanceByte()
				}
				return modelConfigToken{kind: configTokenSymbol, image: literal, line: line, col: col}
			}
		}
		return modelConfigToken{kind: configTokenEOF, line: line, col: col, lexicalFailure: true}
	}
}

func (l *modelConfigLexer) stringToken() modelConfigToken {
	line, col := l.line, l.col
	start := l.offset
	l.advanceByte()
	for !l.eof() {
		ch := l.peekByte()
		if ch == '\n' || ch == '\r' {
			return modelConfigToken{kind: configTokenEOF, line: line, col: col, lexicalFailure: true}
		}
		l.advanceByte()
		if ch == '\\' {
			if l.eof() || !strings.ContainsRune("ntrf\\\"", rune(l.peekByte())) {
				return modelConfigToken{kind: configTokenEOF, line: line, col: col, lexicalFailure: true}
			}
			l.advanceByte()
			continue
		}
		if ch == '"' {
			return modelConfigToken{kind: configTokenString, image: l.input[start:l.offset], line: line, col: col}
		}
	}
	return modelConfigToken{kind: configTokenEOF, line: line, col: col, lexicalFailure: true}
}

func (l *modelConfigLexer) skipWhitespaceAndComments() bool {
	for {
		for !l.eof() {
			ch := l.peekByte()
			if ch != ' ' && ch != '\t' && ch != '\r' && ch != '\n' {
				break
			}
			l.advanceByte()
		}
		if l.eof() {
			return false
		}
		rest := l.input[l.offset:]
		switch {
		case strings.HasPrefix(rest, "\\*"):
			for !l.eof() && l.peekByte() != '\n' && l.peekByte() != '\r' {
				l.advanceByte()
			}
			if l.eof() {
				return true
			}
		case strings.HasPrefix(rest, "(*"):
			if !l.skipBlockComment() {
				return true
			}
		default:
			return false
		}
	}
}

func (l *modelConfigLexer) skipBlockComment() bool {
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
				return true
			}
		default:
			l.advanceByte()
		}
	}
	return false
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
	switch l.input[l.offset] {
	case '\r':
		l.line++
		l.col = 1
		l.previousCR = true
	case '\n':
		if !l.previousCR {
			l.line++
		}
		l.col = 1
		l.previousCR = false
	default:
		l.col++
		l.previousCR = false
	}
	l.offset++
}

func isConfigNumberStart(ch byte) bool {
	return isConfigDigit(ch)
}

func isConfigDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func configBaseNumberLength(input string) int {
	if len(input) < 3 || input[0] != '\\' {
		return 0
	}
	switch input[1] {
	case 'o', 'O', 'b', 'B', 'h', 'H':
	default:
		return 0
	}
	end := 2
	for end < len(input) && isConfigBaseDigit(input[1], input[end]) {
		end++
	}
	if end == 2 {
		return 0
	}
	return end
}

func isConfigBaseDigit(prefix, ch byte) bool {
	switch prefix {
	case 'o', 'O':
		return ch >= '0' && ch <= '7'
	case 'b', 'B':
		return ch == '0' || ch == '1'
	default:
		return ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'f' || ch >= 'A' && ch <= 'F'
	}
}

func isConfigIdentStart(ch byte) bool {
	return ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
}

func isConfigIdentPart(ch byte) bool {
	return ch == '_' || isConfigIdentStart(ch) || isConfigDigit(ch)
}

func configUnderscoreIdentifierLength(input string) int {
	if len(input) == 0 || input[0] != '_' {
		return 0
	}
	end, hasLetter := 0, false
	for end < len(input) && isConfigIdentPart(input[end]) {
		ch := input[end]
		hasLetter = hasLetter || ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z'
		end++
	}
	if !hasLetter {
		return 0
	}
	return end
}

func configJunctionLength(input string) int {
	if len(input) < 4 {
		return 0
	}
	end := 0
	switch {
	case isConfigDigit(input[0]):
		end = 1
	case isConfigIdentStart(input[0]):
		end = 1
		if input[0] == 'S' || input[0] == 'W' {
			if !isConfigASCIIAlphaNumeric(input[1]) || input[1] == 'F' {
				return 0
			}
			end = 2
		}
	default:
		return 0
	}
	for end < len(input) && isConfigASCIIAlphaNumeric(input[end]) {
		end++
	}
	if strings.HasPrefix(input[end:], `.\/`) || strings.HasPrefix(input[end:], `./\`) {
		return end + 3
	}
	return 0
}

func isConfigASCIIAlphaNumeric(ch byte) bool {
	return isConfigIdentStart(ch) || isConfigDigit(ch)
}
