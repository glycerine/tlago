package tlc

import "strings"

// ValueStringPrinter is Java's buffer-based toString overload. Recursive value
// printing uses this surface without entering Value's public string wrapper.
// Values supplied by Go callers may still implement only fmt.Stringer.
type ValueStringPrinter interface {
	ToString(*strings.Builder, int, bool) *strings.Builder
}

// ValueToString mirrors Value.toStringImpl, including its separate source catch
// and the delimiter appended after the concrete printer returns.
func ValueToString(value Value, delimiter string, checked bool) string {
	defer catchValueFailure(value, nil)
	var buffer strings.Builder
	sb := appendValueString(value, &buffer, 0, checked)
	sb.WriteString(delimiter)
	return sb.String()
}

// ValueToStringUnchecked represents both Java toStringUnchecked overloads.
func ValueToStringUnchecked(value Value, delimiter ...string) string {
	ending := ""
	if len(delimiter) != 0 {
		ending = delimiter[0]
	}
	return ValueToString(value, ending, false)
}

func appendValueString(value Value, sb *strings.Builder, offset int, swallow bool) *strings.Builder {
	if value == nil {
		panic(NewNullPointerException())
	}
	if printer, ok := value.(ValueStringPrinter); ok {
		return printer.ToString(sb, offset, swallow)
	}
	sb.WriteString(value.String())
	return sb
}

// Only the Java catch(Throwable) regions use this helper. In unchecked mode
// the failure is rethrown. The caller's buffer is deliberately never reset.
func tryValueString(swallow bool, action func()) (ok bool) {
	defer func() {
		if failure := recover(); failure != nil {
			if !swallow {
				panic(failure)
			}
			ok = false
		}
	}()
	action()
	return true
}
