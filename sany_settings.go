package tlago

import (
	"sort"
	"strconv"

	"github.com/glycerine/tlago/tlc"
)

// SanyErrorCodeSet is the immutable native counterpart of the copied Java
// ErrorCode sets retained by SanySettings.
type SanyErrorCodeSet struct {
	values map[int]SanyErrorCode
}

func NewSanyErrorCodeSet(codes ...SanyErrorCode) SanyErrorCodeSet {
	values := make(map[int]SanyErrorCode, len(codes))
	for _, code := range codes {
		values[code.Value] = code
	}
	return SanyErrorCodeSet{values: values}
}

func (set SanyErrorCodeSet) Contains(code SanyErrorCode) bool {
	stored, ok := set.values[code.Value]
	return ok && stored == code
}

func (set SanyErrorCodeSet) Len() int { return len(set.values) }

func (set SanyErrorCodeSet) All() []SanyErrorCode {
	result := make([]SanyErrorCode, 0, len(set.values))
	for _, code := range set.values {
		result = append(result, code)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Value < result[j].Value })
	return result
}

type SanySettings struct {
	DoSemanticAnalysis      bool
	DoLevelChecking         bool
	DoLinting               bool
	ValidatePCalTranslation bool
	SuppressedCodes         SanyErrorCodeSet
	MessagesAsErrorCodes    SanyErrorCodeSet
	doStrictErrorCodes      bool
}

func DefaultSanySettings() *SanySettings {
	return newSanySettings(true, true, true, true, true, nil, nil)
}

func ValidASTSanySettings() *SanySettings {
	return newSanySettings(true, true, true, false, false, nil, nil)
}

func NewSanySettings(semantic, levels, lint, validatePCal bool) *SanySettings {
	return newSanySettings(true, semantic, levels, lint, validatePCal, nil, nil)
}

func NewSanySettingsWithCodes(semantic, levels, lint, validatePCal bool, suppressed, elevated []SanyErrorCode) *SanySettings {
	return newSanySettings(true, semantic, levels, lint, validatePCal, suppressed, elevated)
}

func SanySettingsForExternalCaller(suppressed, elevated []SanyErrorCode) *SanySettings {
	return newSanySettings(true, true, true, true, true, suppressed, elevated)
}

func newSanySettings(strict, semantic, levels, lint, validatePCal bool, suppressed, elevated []SanyErrorCode) *SanySettings {
	return &SanySettings{
		DoSemanticAnalysis:      semantic,
		DoLevelChecking:         levels,
		DoLinting:               lint,
		ValidatePCalTranslation: validatePCal,
		SuppressedCodes:         NewSanyErrorCodeSet(suppressed...),
		MessagesAsErrorCodes:    NewSanyErrorCodeSet(elevated...),
		doStrictErrorCodes:      strict,
	}
}

func (settings *SanySettings) driverSettings() sanyDriverSettings {
	if settings == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyDriverSettings{
		strict:       settings.doStrictErrorCodes,
		semantic:     settings.DoSemanticAnalysis,
		levels:       settings.DoLevelChecking,
		lint:         settings.DoLinting,
		validatePCal: settings.ValidatePCalTranslation,
		messages: diagnosticCLIOptions{
			suppressed: sanySettingsDiagnosticCodes(settings.SuppressedCodes),
			elevated:   sanySettingsDiagnosticCodes(settings.MessagesAsErrorCodes),
		},
	}
}

func sanySettingsDiagnosticCodes(codes SanyErrorCodeSet) map[string]bool {
	if codes.Len() == 0 {
		return nil
	}
	result := make(map[string]bool, codes.Len())
	for _, code := range codes.All() {
		result[strconv.Itoa(code.Value)] = true
	}
	return result
}

// ParseSanyWithSettings exposes the settings-controlled SANY driver path.
func ParseSanyWithSettings(file string, opts LoadOptions, settings *SanySettings, output func(Severity, string)) (*Spec, SanyExitCode) {
	return parseSanyWithSettings(file, opts, settings.driverSettings(), output)
}
