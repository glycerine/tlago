package tlago

import "github.com/glycerine/tlago/tlc"

// SanyLevelException adapts tla2sany.semantic.LevelException, the checked
// exception carrier used by semantic level checking.
type SanyLevelException struct {
	tlc.JavaExceptionBase
}

// NewSanyLevelException constructs a level exception with a non-null message.
func NewSanyLevelException(message string) *SanyLevelException {
	return NewSanyLevelExceptionNullable(&message)
}

// NewSanyLevelExceptionNullable preserves Java's nullable String constructor.
func NewSanyLevelExceptionNullable(message *string) *SanyLevelException {
	return &SanyLevelException{JavaExceptionBase: tlc.NewJavaExceptionBase(message, nil)}
}

func (e *SanyLevelException) require() {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (e *SanyLevelException) Error() string {
	if e == nil {
		return "tla2sany.semantic.LevelException"
	}
	result := "tla2sany.semantic.LevelException"
	if message := e.GetMessage(); message != nil {
		result += ": " + *message
	}
	return result
}

func (e *SanyLevelException) JavaThrowableString() string { return e.Error() }

func (e *SanyLevelException) GetMessage() *string {
	e.require()
	return e.JavaExceptionBase.GetMessage()
}

func (e *SanyLevelException) GetCause() error {
	e.require()
	return e.JavaExceptionBase.GetCause()
}
