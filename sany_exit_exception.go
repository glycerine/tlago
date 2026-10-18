package tlago

import "github.com/glycerine/tlago/tlc"

// SanyExitException is tla2sany.drivers.SANYExitException. It carries the
// enumerated SANY status at boundaries where the Java driver would exit.
type SanyExitException struct {
	tlc.JavaExceptionBase
	exitCode *SanyExitCode
}

func NewSanyExitException(exitCode SanyExitCode, message string) *SanyExitException {
	return NewSanyExitExceptionNullable(&exitCode, &message)
}

func NewSanyExitExceptionNullable(exitCode *SanyExitCode, message *string) *SanyExitException {
	return &SanyExitException{
		JavaExceptionBase: tlc.NewJavaExceptionBase(message, nil),
		exitCode:          exitCode,
	}
}

func (e *SanyExitException) require() {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (e *SanyExitException) Error() string {
	if e == nil {
		return "tla2sany.drivers.SANYExitException"
	}
	result := "tla2sany.drivers.SANYExitException"
	if message := e.GetMessage(); message != nil {
		result += ": " + *message
	}
	return result
}

func (e *SanyExitException) GetMessage() *string {
	e.require()
	return e.JavaExceptionBase.GetMessage()
}

func (e *SanyExitException) GetCause() error {
	e.require()
	return e.JavaExceptionBase.GetCause()
}

func (e *SanyExitException) GetExitCode() int {
	e.require()
	if e.exitCode == nil {
		panic(tlc.NewNullPointerException())
	}
	return e.exitCode.Code()
}

func (e *SanyExitException) GetEnumeratedExitCode() *SanyExitCode {
	e.require()
	return e.exitCode
}
