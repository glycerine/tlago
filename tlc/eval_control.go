package tlc

const (
	EvalKeepLazy = 1
	EvalPrimed   = 1 << 1
	EvalEnabled  = 1 << 2
	EvalInit     = 1 << 3
	EvalConst    = 1 << 4
	EvalClear    = 0
)

type PartialBoolean int

const (
	PartialMaybe PartialBoolean = iota
	PartialYes
	PartialNo
)

func (p PartialBoolean) String() string {
	switch p {
	case PartialYes:
		return "YES"
	case PartialNo:
		return "NO"
	case PartialMaybe:
		return "MAYBE"
	default:
		return "UNKNOWN"
	}
}

func (p PartialBoolean) IsDefinitely(value bool) bool {
	switch p {
	case PartialYes:
		return value
	case PartialNo:
		return !value
	default:
		return false
	}
}

func (p PartialBoolean) CouldBe(value bool) bool {
	return !p.IsDefinitely(!value)
}

func evalControlIsSet(control int, constant int) bool {
	return control&constant > 0
}

func EvalIsKeepLazy(control int) bool { return evalControlIsSet(control, EvalKeepLazy) }
func EvalSetKeepLazy(control int) int { return control | EvalKeepLazy }
func EvalIsPrimed(control int) bool   { return evalControlIsSet(control, EvalPrimed) }
func EvalSetPrimed(control int) int   { return control | EvalPrimed }
func EvalIsEnabled(control int) bool  { return evalControlIsSet(control, EvalEnabled) }
func EvalSetEnabled(control int) int  { return control | EvalEnabled }
func EvalIsInit(control int) bool     { return evalControlIsSet(control, EvalInit) }
func EvalIsConst(control int) bool    { return evalControlIsSet(control, EvalConst) }

func EvalSetPrimedIfEnabled(control int) int {
	if EvalIsEnabled(control) {
		return EvalSetPrimed(control)
	}
	return control
}

func EvalSemanticallyEquivalent(control1, control2 int) PartialBoolean {
	flagsThatCanBeSafelyIgnored := EvalKeepLazy | EvalInit | EvalConst
	mask := ^flagsThatCanBeSafelyIgnored
	if control1&mask == control2&mask {
		return PartialYes
	}
	return PartialMaybe
}
