// Portions Copyright (c) 2025, Oracle and/or its affiliates.
package tlc

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

var mpConsole = struct {
	sync.Mutex
	warningHistory map[string]bool
}{warningHistory: make(map[string]bool)}

type mpDecimalSymbols struct {
	zero                                  rune
	group, negativePrefix, negativeSuffix string
}

var mpNumberSymbolsOnce sync.Once
var mpNumberSymbols mpDecimalSymbols
var mpNumberLocaleKey string

// Java's explicit ###,###.### DecimalFormat groups by three, independently of
// the locale's usual grouping pattern. Formatting retains all signed long bits.
func MessageNumberFormat(value int64) string {
	mpNumberSymbolsOnce.Do(initializeMessageNumberSymbols)
	text := strconv.FormatInt(value, 10)
	negative := value < 0
	if negative {
		text = text[1:]
	}
	var out strings.Builder
	if negative {
		out.WriteString(mpNumberSymbols.negativePrefix)
	}
	for i, c := range text {
		if i > 0 && (len(text)-i)%3 == 0 {
			out.WriteString(mpNumberSymbols.group)
		}
		out.WriteRune(mpNumberSymbols.zero + rune(c-'0'))
	}
	if negative {
		out.WriteString(mpNumberSymbols.negativeSuffix)
	}
	return out.String()
}

func initializeMessageNumberSymbols() {
	lang := tlcGetSystemProperty("user.language.format", tlcGetSystemProperty("user.language", "en"))
	script := tlcGetSystemProperty("user.script.format", tlcGetSystemProperty("user.script", ""))
	country := tlcGetSystemProperty("user.country.format", tlcGetSystemProperty("user.country", ""))
	variant := tlcGetSystemProperty("user.variant.format", tlcGetSystemProperty("user.variant", ""))
	extensions := tlcGetSystemProperty("user.extensions.format", tlcGetSystemProperty("user.extensions", ""))
	switch lang {
	case "iw":
		lang = "he"
	case "ji":
		lang = "yi"
	case "in":
		lang = "id"
	}
	parts := []string{strings.ToLower(lang)}
	if script != "" {
		parts = append(parts, strings.ToUpper(script[:1])+strings.ToLower(script[1:]))
	}
	if country != "" {
		parts = append(parts, strings.ToUpper(country))
	}
	if variant != "" {
		parts = append(parts, variant)
	}
	// Locale's compatibility forms retain legacy Japanese/Thai variants.
	if lang == "no" && country == "NO" && variant == "NY" {
		parts = []string{"nn", "NO"}
	}
	if lang == "th" && country == "TH" && variant == "TH" && extensions == "" {
		extensions = "u-nu-thai"
	}
	mpNumberSymbols, mpNumberLocaleKey = selectMessageNumberSymbols(parts, extensions)
	cacheNumberSpecialSymbols = selectCacheDecimalSymbols(parts, extensions)
}

func selectMessageNumberSymbols(parts []string, extensions string) (mpDecimalSymbols, string) {
	set := javaMPDecimalSymbolSets[javaMPDecimalSymbols["und"]]
	key := "und"
	for len(parts) > 0 {
		if index, ok := javaMPDecimalSymbols[strings.Join(parts, "-")]; ok {
			key = strings.Join(parts, "-")
			set = javaMPDecimalSymbolSets[index]
			break
		}
		parts = parts[:len(parts)-1]
	}
	symbols := set["default"]
	if selected, ok := set[mpNumberingSystem(extensions)]; ok {
		symbols = selected
	}
	return symbols, key
}

// Only a valid Unicode nu keyword selects a numbering system. Private-use
// subtags and unknown/multipart numbering names retain the locale default.
func mpNumberingSystem(extensions string) string {
	parts := strings.Split(strings.ToLower(extensions), "-")
	var unicodeParts []string
	for i := 0; i < len(parts); {
		key := parts[i]
		if len(key) != 1 || !mpLocaleSubtag(key) {
			return ""
		}
		i++
		start := i
		min := 2
		if key == "x" {
			min = 1
		}
		for i < len(parts) && (key == "x" || len(parts[i]) != 1) {
			if len(parts[i]) < min || len(parts[i]) > 8 || !mpLocaleSubtag(parts[i]) {
				return ""
			}
			i++
		}
		if start == i {
			return ""
		}
		if key == "u" && unicodeParts == nil {
			unicodeParts = parts[start:i]
		}
	}
	for i := 0; i < len(unicodeParts); {
		if len(unicodeParts[i]) != 2 {
			i++
			continue
		}
		key := unicodeParts[i]
		i++
		start := i
		for i < len(unicodeParts) && len(unicodeParts[i]) > 2 {
			i++
		}
		if key == "nu" {
			return strings.Join(unicodeParts[start:i], "-")
		}
	}
	return ""
}
func mpLocaleSubtag(s string) bool {
	for _, c := range s {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9') {
			return false
		}
	}
	return s != ""
}

func printConsoleMessage(code int, severity Severity, text string, visible bool) string {
	Globals.Lock()
	tool := Globals.Tool
	Globals.Unlock()
	mpConsole.Lock()
	defer mpConsole.Unlock()
	if severity == SeverityWarning {
		if tool {
			text = consoleMessageEnvelope(code, severity, text)
		} else {
			text = "Warning: " + text
			if len(mpConsole.warningHistory) == 0 {
				text += "\n(Use the -nowarning option to disable this warning.)"
			}
		}
		DebugPrintMessage("Leaving getMessage()")
		if mpConsole.warningHistory[text] {
			return text
		}
		mpConsole.warningHistory[text] = true
		// Java checks individual warning suppression after formatting and history.
		suppressed, _, _ := messageControlFor(code)
		visible = !suppressed
	} else if tool {
		text = consoleMessageEnvelope(code, severity, text)
	} else {
		switch severity {
		case SeverityError:
			text = "Error: " + text
		case SeverityTLCBug:
			text = "TLC Bug: " + text
		case SeverityState:
			text = "State " + text
		}
	}
	if severity != SeverityWarning {
		DebugPrintMessage("Leaving getMessage()")
	}
	if visible {
		ToolIOPrintln(text)
	}
	return text
}

func consoleMessageEnvelope(code int, severity Severity, text string) string {
	return fmt.Sprintf("@!@!@STARTMSG %d:%d @!@!@\n%s\n@!@!@ENDMSG %d @!@!@", code, severity, text, code)
}
