package tlago

import "github.com/glycerine/tlago/tlc"

type SanyParseError struct {
	reported string
	backup   string
}

func NewSanyParseError(reported string, backup ...string) *SanyParseError {
	defaultError := ""
	if len(backup) != 0 {
		defaultError = backup[0]
	}
	return &SanyParseError{reported: reported, backup: defaultError}
}

func (e *SanyParseError) require() {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (e *SanyParseError) ReportedError() string {
	e.require()
	return e.reported
}

func (e *SanyParseError) DefaultError() string {
	e.require()
	return e.backup
}

type SanyParseErrors struct {
	errors []*SanyParseError
}

func NewSanyParseErrors() *SanyParseErrors {
	return &SanyParseErrors{errors: make([]*SanyParseError, 0)}
}

func (e *SanyParseErrors) require() {
	if e == nil {
		panic(tlc.NewNullPointerException())
	}
}

func (e *SanyParseErrors) empty() bool {
	e.require()
	return len(e.errors) == 0
}

func (e *SanyParseErrors) push(parseError *SanyParseError) {
	e.require()
	e.errors = append(e.errors, parseError)
}

func (e *SanyParseErrors) Errors() []*SanyParseError {
	e.require()
	result := make([]*SanyParseError, len(e.errors))
	copy(result, e.errors)
	return result
}
