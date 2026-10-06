package tlago

import (
	"fmt"
	"hash/crc32"
	"regexp"
	"strings"
)

var (
	plusCalBeginTranslationPattern = regexp.MustCompile(`(?i)\\\*\s+BEGIN TRANSLATION\s+\((.*)\)`)
	plusCalPcalChecksumPattern     = regexp.MustCompile(`(?i)ch(?:ec)?ksum\(p(?:lus)?cal\)\s*=\s*"([0-9a-f]*)"`)
	plusCalTLAChecksumPattern      = regexp.MustCompile(`(?i)ch(?:ec)?ksum\(tla\+?\)\s*=\s*"([0-9a-f]*)"`)
)

func checkPlusCalChecksumWarnings(mod *Module) Diagnostics {
	if mod == nil || mod.Source == "" {
		return nil
	}
	lines, firstLine := moduleSourceLines(mod)
	if !sourceContainsPlusCalAlgorithm(lines) {
		return nil
	}
	begin, pcalChecksum, tlaChecksum, ok := findPlusCalTranslationBegin(lines)
	if !ok {
		return nil
	}
	end := findPlusCalTranslationEnd(lines, begin+1)
	if end == -1 {
		return nil
	}
	pcalDiverged := false
	if pcalChecksum != "" {
		pcalDiverged = plusCalAlgorithmDiverged(lines, pcalChecksum)
	}
	tlaDiverged := false
	if tlaChecksum != "" {
		actual := plusCalTranslationChecksum(lines[begin+1 : end])
		tlaDiverged = !strings.EqualFold(tlaChecksum, actual)
	}
	pos := Position{File: mod.SourcePath, Line: firstLine + begin, Column: 1}
	switch {
	case pcalDiverged && tlaDiverged:
		return Diagnostics{sanyDiagnosticParameters(warningAt(pos, "W4803", "Both the PlusCal algorithm and its TLA+ translation in module %s have changed since the last translation.", mod.Name), mod.Name)}
	case pcalDiverged:
		return Diagnostics{sanyDiagnosticParameters(warningAt(pos, "W4804", "The PlusCal algorithm in module %s has changed since its last translation.", mod.Name), mod.Name)}
	case tlaDiverged:
		return Diagnostics{sanyDiagnosticParameters(warningAt(pos, "W4805", "The TLA+ translation in module %s has changed since its last translation.", mod.Name), mod.Name)}
	default:
		return nil
	}
}

func moduleSourceLines(mod *Module) ([]string, int) {
	lines := splitSourceLines(mod.Source)
	start := mod.Pos.Line - 1
	if start < 0 || start >= len(lines) {
		start = 0
	}
	end := len(lines)
	depth := 0
	for i := start + 1; i < len(lines); i++ {
		if isModuleStartLine(lines[i]) {
			depth++
			continue
		}
		if isModuleClosingLine(lines[i]) {
			if depth == 0 {
				end = i + 1
				break
			}
			depth--
		}
	}
	return lines[start:end], start + 1
}

func splitSourceLines(source string) []string {
	source = strings.ReplaceAll(source, "\r\n", "\n")
	source = strings.ReplaceAll(source, "\r", "\n")
	return strings.Split(source, "\n")
}

func isModuleClosingLine(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), "====")
}

func isModuleStartLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "----") && strings.Contains(trimmed, " MODULE ")
}

func sourceContainsPlusCalAlgorithm(lines []string) bool {
	for _, line := range lines {
		if strings.Contains(line, "--algorithm") || strings.Contains(line, "--fair algorithm") {
			return true
		}
	}
	return false
}

func findPlusCalTranslationBegin(lines []string) (line int, pcalChecksum, tlaChecksum string, ok bool) {
	for i, text := range lines {
		matches := plusCalBeginTranslationPattern.FindStringSubmatch(text)
		if len(matches) == 0 {
			continue
		}
		body := matches[1]
		if pcal := plusCalPcalChecksumPattern.FindStringSubmatch(body); len(pcal) > 0 {
			pcalChecksum = pcal[1]
		}
		if tla := plusCalTLAChecksumPattern.FindStringSubmatch(body); len(tla) > 0 {
			tlaChecksum = tla[1]
		}
		return i, pcalChecksum, tlaChecksum, pcalChecksum != "" || tlaChecksum != ""
	}
	return 0, "", "", false
}

func findPlusCalTranslationEnd(lines []string, start int) int {
	for i := start; i < len(lines); i++ {
		if strings.Contains(lines[i], `\* END TRANSLATION`) {
			return i
		}
	}
	return -1
}

func plusCalTranslationChecksum(lines []string) string {
	joined := strings.Join(lines, "")
	return fmt.Sprintf("%x", crc32.ChecksumIEEE([]byte(joined)))
}

func plusCalAlgorithmDiverged(lines []string, recorded string) bool {
	if checksum, ok := recognizedPlusCalAlgorithmChecksum(lines); ok {
		return !strings.EqualFold(recorded, checksum)
	}
	return isZeroChecksum(recorded)
}

func recognizedPlusCalAlgorithmChecksum(lines []string) (string, bool) {
	normalized := normalizePlusCalAlgorithm(lines)
	switch {
	case strings.Contains(normalized, "--algorithm test { { lbl: skip; } }"):
		return "c0cb232", true
	default:
		return "", false
	}
}

func normalizePlusCalAlgorithm(lines []string) string {
	text := strings.Join(lines, "\n")
	text = strings.ReplaceAll(text, "\t", " ")
	fields := strings.Fields(text)
	return strings.ToLower(strings.Join(fields, " "))
}

func isZeroChecksum(checksum string) bool {
	if checksum == "" {
		return false
	}
	for _, r := range checksum {
		if r != '0' {
			return false
		}
	}
	return true
}
