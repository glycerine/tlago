// OpenJDK formatter exception families reached by TLC's string arguments.
// GPLv2 with Classpath Exception; see tlc/licenses/openjdk-*.
package tlc

import "strconv"

type IllegalFormatException struct{ *IllegalArgumentException }

func newIllegalFormatException(message string) *IllegalFormatException {
	return &IllegalFormatException{NewIllegalArgumentException(message)}
}

func (e *IllegalFormatException) Error() string { return javaThrowableMessage(e) }

type DuplicateFormatFlagsException struct {
	*IllegalFormatException
	Flags string
}

func newDuplicateFormatFlagsException(flags string) *DuplicateFormatFlagsException {
	return &DuplicateFormatFlagsException{IllegalFormatException: newIllegalFormatException("Flags = '" + flags + "'"), Flags: flags}
}
func (e *DuplicateFormatFlagsException) GetFlags() string { return e.Flags }
func (e *DuplicateFormatFlagsException) Error() string    { return javaThrowableMessage(e) }

type FormatFlagsConversionMismatchException struct {
	*IllegalFormatException
	Flags      string
	Conversion byte
}

func newFormatFlagsConversionMismatchException(flags string, conversion byte) *FormatFlagsConversionMismatchException {
	return &FormatFlagsConversionMismatchException{IllegalFormatException: newIllegalFormatException("Conversion = " + string(conversion) + ", Flags = " + flags), Flags: flags, Conversion: conversion}
}
func (e *FormatFlagsConversionMismatchException) GetFlags() string    { return e.Flags }
func (e *FormatFlagsConversionMismatchException) GetConversion() byte { return e.Conversion }
func (e *FormatFlagsConversionMismatchException) Error() string       { return javaThrowableMessage(e) }

type IllegalFormatConversionException struct {
	*IllegalFormatException
	Conversion    byte
	ArgumentClass string
}

func newIllegalFormatConversionException(conversion byte) *IllegalFormatConversionException {
	return &IllegalFormatConversionException{IllegalFormatException: newIllegalFormatException(string(conversion) + " != java.lang.String"), Conversion: conversion, ArgumentClass: "java.lang.String"}
}
func (e *IllegalFormatConversionException) GetConversion() byte      { return e.Conversion }
func (e *IllegalFormatConversionException) GetArgumentClass() string { return e.ArgumentClass }
func (e *IllegalFormatConversionException) Error() string            { return javaThrowableMessage(e) }

type IllegalFormatFlagsException struct {
	*IllegalFormatException
	Flags string
}

func newIllegalFormatFlagsException(flags string) *IllegalFormatFlagsException {
	return &IllegalFormatFlagsException{IllegalFormatException: newIllegalFormatException("Flags = '" + flags + "'"), Flags: flags}
}
func (e *IllegalFormatFlagsException) GetFlags() string { return e.Flags }
func (e *IllegalFormatFlagsException) Error() string    { return javaThrowableMessage(e) }

type IllegalFormatPrecisionException struct {
	*IllegalFormatException
	Precision int
}

func newIllegalFormatPrecisionException(precision int) *IllegalFormatPrecisionException {
	return &IllegalFormatPrecisionException{IllegalFormatException: newIllegalFormatException(strconv.Itoa(precision)), Precision: precision}
}
func (e *IllegalFormatPrecisionException) GetPrecision() int { return e.Precision }
func (e *IllegalFormatPrecisionException) Error() string     { return javaThrowableMessage(e) }

type IllegalFormatWidthException struct {
	*IllegalFormatException
	Width int
}

func newIllegalFormatWidthException(width int) *IllegalFormatWidthException {
	return &IllegalFormatWidthException{IllegalFormatException: newIllegalFormatException(strconv.Itoa(width)), Width: width}
}
func (e *IllegalFormatWidthException) GetWidth() int { return e.Width }
func (e *IllegalFormatWidthException) Error() string { return javaThrowableMessage(e) }

type MissingFormatArgumentException struct {
	*IllegalFormatException
	Specifier string
}

func newMissingFormatArgumentException(specifier string) *MissingFormatArgumentException {
	return &MissingFormatArgumentException{IllegalFormatException: newIllegalFormatException("Format specifier '" + specifier + "'"), Specifier: specifier}
}
func (e *MissingFormatArgumentException) GetFormatSpecifier() string { return e.Specifier }
func (e *MissingFormatArgumentException) Error() string              { return javaThrowableMessage(e) }

type MissingFormatWidthException struct {
	*IllegalFormatException
	Specifier string
}

func newMissingFormatWidthException(specifier string) *MissingFormatWidthException {
	return &MissingFormatWidthException{IllegalFormatException: newIllegalFormatException(specifier), Specifier: specifier}
}
func (e *MissingFormatWidthException) GetFormatSpecifier() string { return e.Specifier }
func (e *MissingFormatWidthException) Error() string              { return javaThrowableMessage(e) }

type UnknownFormatConversionException struct {
	*IllegalFormatException
	Conversion string
}

func newUnknownFormatConversionException(conversion string) *UnknownFormatConversionException {
	return &UnknownFormatConversionException{IllegalFormatException: newIllegalFormatException("Conversion = '" + conversion + "'"), Conversion: conversion}
}
func (e *UnknownFormatConversionException) GetConversion() string { return e.Conversion }
func (e *UnknownFormatConversionException) Error() string         { return javaThrowableMessage(e) }

type IllegalFormatArgumentIndexException struct {
	*IllegalFormatException
	Index int
}

func newIllegalFormatArgumentIndexException(index int) *IllegalFormatArgumentIndexException {
	message := "Illegal format argument index = " + strconv.Itoa(index)
	if index == -1<<31 {
		message = "Format argument index: (not representable as int)"
	}
	return &IllegalFormatArgumentIndexException{newIllegalFormatException(message), index}
}
func (e *IllegalFormatArgumentIndexException) GetIndex() int { return e.Index }
func (e *IllegalFormatArgumentIndexException) Error() string { return javaThrowableMessage(e) }
