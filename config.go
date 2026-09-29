package tlago

import (
	"bufio"
	"strconv"
	"strings"
)

type ModelConfig struct {
	Specification     string
	Init              string
	Next              string
	Constants         State
	ModelValues       map[string]string
	Invariants        []string
	Properties        []string
	Constraints       []string
	ActionConstraints []string
	Views             []string
	Symmetry          []string
	Aliases           []string
	Postconditions    []string
	CheckDeadlock     bool
}

func ParseConfigSource(file, source string) (ModelConfig, Diagnostics) {
	cfg := ModelConfig{CheckDeadlock: true, Constants: State{}, ModelValues: map[string]string{}}
	var diags Diagnostics
	scanner := bufio.NewScanner(strings.NewReader(source))
	line := 0
	for scanner.Scan() {
		line++
		text := strings.TrimSpace(scanner.Text())
		if text == "" || strings.HasPrefix(text, "\\*") || strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Fields(text)
		if len(fields) == 0 {
			continue
		}
		pos := Position{File: file, Line: line, Column: 1}
		switch fields[0] {
		case "SPECIFICATION":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1400", "SPECIFICATION requires an operator name"))
				continue
			}
			cfg.Specification = fields[1]
		case "INIT":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1401", "INIT requires an operator name"))
				continue
			}
			cfg.Init = fields[1]
		case "NEXT":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1402", "NEXT requires an operator name"))
				continue
			}
			cfg.Next = fields[1]
		case "CONSTANT", "CONSTANTS":
			if len(fields) < 4 {
				diags = append(diags, errorAt(pos, "E1406", "%s requires NAME = value", fields[0]))
				continue
			}
			for i := 1; i < len(fields); {
				if i+2 >= len(fields) || fields[i+1] != "=" {
					diags = append(diags, errorAt(pos, "E1406", "expected NAME = value in %s entry", fields[0]))
					break
				}
				val, err := strconv.Atoi(fields[i+2])
				if err == nil {
					cfg.Constants[fields[i]] = val
					delete(cfg.ModelValues, fields[i])
					i += 3
					continue
				}
				if isConfigModelValue(fields[i+2]) {
					delete(cfg.Constants, fields[i])
					cfg.ModelValues[fields[i]] = fields[i+2]
					i += 3
					continue
				}
				diags = append(diags, errorAt(pos, "E1407", "constant %s must be an integer or model value", fields[i]))
				break
			}
		case "INVARIANT", "INVARIANTS":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1403", "%s requires at least one operator name", fields[0]))
				continue
			}
			cfg.Invariants = append(cfg.Invariants, fields[1:]...)
		case "PROPERTY", "PROPERTIES":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1410", "%s requires at least one operator name", fields[0]))
				continue
			}
			cfg.Properties = append(cfg.Properties, fields[1:]...)
		case "CONSTRAINT", "CONSTRAINTS":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1408", "%s requires at least one operator name", fields[0]))
				continue
			}
			cfg.Constraints = append(cfg.Constraints, fields[1:]...)
		case "ACTION_CONSTRAINT", "ACTION_CONSTRAINTS":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1409", "%s requires at least one operator name", fields[0]))
				continue
			}
			cfg.ActionConstraints = append(cfg.ActionConstraints, fields[1:]...)
		case "VIEW", "VIEWS":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1411", "%s requires at least one operator name", fields[0]))
				continue
			}
			cfg.Views = append(cfg.Views, fields[1:]...)
		case "SYMMETRY", "SYMMETRIES":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1412", "%s requires at least one operator name", fields[0]))
				continue
			}
			cfg.Symmetry = append(cfg.Symmetry, fields[1:]...)
		case "ALIAS", "ALIASES":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1413", "%s requires at least one operator name", fields[0]))
				continue
			}
			cfg.Aliases = append(cfg.Aliases, fields[1:]...)
		case "POSTCONDITION", "POSTCONDITIONS":
			if len(fields) < 2 {
				diags = append(diags, errorAt(pos, "E1414", "%s requires at least one operator name", fields[0]))
				continue
			}
			cfg.Postconditions = append(cfg.Postconditions, fields[1:]...)
		case "CHECK_DEADLOCK":
			if len(fields) >= 2 && fields[1] == "FALSE" {
				cfg.CheckDeadlock = false
			}
		default:
			diags = append(diags, errorAt(pos, "E1404", "unsupported config entry %s", fields[0]))
		}
	}
	if err := scanner.Err(); err != nil {
		diags = append(diags, errorAt(Position{File: file, Line: line, Column: 1}, "E1405", "cannot read config: %v", err))
	}
	return cfg, diags
}

func isConfigModelValue(text string) bool {
	if text == "" || text == "TRUE" || text == "FALSE" {
		return false
	}
	if !isConfigIdentStart(text[0]) {
		return false
	}
	for i := 1; i < len(text); i++ {
		if !isConfigIdentPart(text[i]) {
			return false
		}
	}
	return true
}

func isConfigIdentStart(ch byte) bool {
	return ch == '_' || ('A' <= ch && ch <= 'Z') || ('a' <= ch && ch <= 'z')
}

func isConfigIdentPart(ch byte) bool {
	return isConfigIdentStart(ch) || ('0' <= ch && ch <= '9')
}
