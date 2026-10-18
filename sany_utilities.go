package tlago

import (
	"strings"

	"github.com/glycerine/tlago/tlc"
)

type SanyAssert struct{}

func NewSanyAssert() *SanyAssert { return &SanyAssert{} }

func SanyAssertion(bool) {}

func SanyAssertFail(string) {}

type SanyIntWrapper struct {
	value int32
}

func NewSanyIntWrapper(initialValues ...int) *SanyIntWrapper {
	if len(initialValues) > 1 {
		panic(tlc.NewIllegalArgumentException("NewSanyIntWrapper accepts at most one initial value"))
	}
	var value int32
	if len(initialValues) == 1 {
		value = int32(initialValues[0])
	}
	return &SanyIntWrapper{value: value}
}

func (w *SanyIntWrapper) require() {
	if w == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (w *SanyIntWrapper) Inc(increments ...int) {
	w.require()
	if len(increments) > 1 {
		panic(tlc.NewIllegalArgumentException("Inc accepts at most one increment"))
	}
	increment := int32(1)
	if len(increments) == 1 {
		increment = int32(increments[0])
	}
	w.value += increment
}

func (w *SanyIntWrapper) Value() int {
	w.require()
	return int(w.value)
}

func (w *SanyIntWrapper) Set(value int) {
	w.require()
	w.value = int32(value)
}

var SanyStringBlanks = [9]string{"", " ", "| ", "   ", "| | ", "     ", "| | | ", "       ", "| | | | "}

func SanyIndent(count int, value string) string {
	return sanyIndent(count, value)
}

func SanyIndentSB(count int, value string) string {
	return sanyIndent(count, value)
}

func sanyIndent(count int, value string) string {
	if count >= len(SanyStringBlanks) {
		count = len(SanyStringBlanks) - 1
	}
	var result strings.Builder
	result.Grow(len(value) * 2)
	for index := 0; index < len(value); index++ {
		character := value[index]
		result.WriteByte(character)
		if character == '\n' && index != len(value)-1 {
			if count < 0 || count >= len(SanyStringBlanks) {
				panic(tlc.NewArrayIndexOutOfBoundsExceptionNoMessage())
			}
			result.WriteString(SanyStringBlanks[count])
		}
	}
	return result.String()
}
