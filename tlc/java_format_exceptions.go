package tlc

import "strconv"

type IllegalFormatException struct{ *IllegalArgumentException }

func newIllegalFormatException() *IllegalFormatException {
	return &IllegalFormatException{NewIllegalArgumentException()}
}

func (e *IllegalFormatException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalFormatException) GetMessage() *string {
	e.require()
	return e.IllegalArgumentException.GetMessage()
}
func (e *IllegalFormatException) GetCause() error {
	e.require()
	return e.IllegalArgumentException.GetCause()
}
func (e *IllegalFormatException) require() {
	if e == nil || e.IllegalArgumentException == nil {
		panic(NewNullPointerException())
	}
}

type DuplicateFormatFlagsException struct {
	*IllegalFormatException
	Flags string
}

func NewDuplicateFormatFlagsException(flags string) *DuplicateFormatFlagsException {
	return NewDuplicateFormatFlagsExceptionNullable(javaString(flags))
}
func NewDuplicateFormatFlagsExceptionNullable(flags *string) *DuplicateFormatFlagsException {
	return &DuplicateFormatFlagsException{IllegalFormatException: newIllegalFormatException(), Flags: requireFormatString(flags)}
}
func newDuplicateFormatFlagsException(flags string) *DuplicateFormatFlagsException {
	return NewDuplicateFormatFlagsException(flags)
}
func (e *DuplicateFormatFlagsException) GetFlags() string {
	e.require()
	return e.Flags
}
func (e *DuplicateFormatFlagsException) GetMessage() *string {
	e.require()
	return javaString("Flags = '" + e.Flags + "'")
}
func (e *DuplicateFormatFlagsException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *DuplicateFormatFlagsException) Error() string { return javaThrowableMessage(e) }
func (e *DuplicateFormatFlagsException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

type FormatFlagsConversionMismatchException struct {
	*IllegalFormatException
	Flags      string
	Conversion uint16
}

func NewFormatFlagsConversionMismatchException(flags string, conversion uint16) *FormatFlagsConversionMismatchException {
	return NewFormatFlagsConversionMismatchExceptionNullable(javaString(flags), conversion)
}
func NewFormatFlagsConversionMismatchExceptionNullable(flags *string, conversion uint16) *FormatFlagsConversionMismatchException {
	return &FormatFlagsConversionMismatchException{IllegalFormatException: newIllegalFormatException(), Flags: requireFormatString(flags), Conversion: conversion}
}
func newFormatFlagsConversionMismatchException(flags string, conversion byte) *FormatFlagsConversionMismatchException {
	return NewFormatFlagsConversionMismatchException(flags, uint16(conversion))
}
func (e *FormatFlagsConversionMismatchException) GetFlags() string {
	e.require()
	return e.Flags
}
func (e *FormatFlagsConversionMismatchException) GetConversion() uint16 {
	e.require()
	return e.Conversion
}
func (e *FormatFlagsConversionMismatchException) GetMessage() *string {
	e.require()
	return javaString("Conversion = " + formatExceptionCharacter(e.Conversion) + ", Flags = " + e.Flags)
}
func (e *FormatFlagsConversionMismatchException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *FormatFlagsConversionMismatchException) Error() string { return javaThrowableMessage(e) }
func (e *FormatFlagsConversionMismatchException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

type IllegalFormatConversionException struct {
	*IllegalFormatException
	Conversion    uint16
	ArgumentClass string
}

func NewIllegalFormatConversionException(conversion uint16, argumentClass string) *IllegalFormatConversionException {
	return NewIllegalFormatConversionExceptionNullable(conversion, javaString(argumentClass))
}
func NewIllegalFormatConversionExceptionNullable(conversion uint16, argumentClass *string) *IllegalFormatConversionException {
	return &IllegalFormatConversionException{IllegalFormatException: newIllegalFormatException(), Conversion: conversion, ArgumentClass: requireFormatString(argumentClass)}
}
func newIllegalFormatConversionException(conversion byte) *IllegalFormatConversionException {
	return NewIllegalFormatConversionException(uint16(conversion), "java.lang.String")
}
func (e *IllegalFormatConversionException) GetConversion() uint16 {
	e.require()
	return e.Conversion
}
func (e *IllegalFormatConversionException) GetArgumentClass() string {
	e.require()
	return e.ArgumentClass
}
func (e *IllegalFormatConversionException) GetMessage() *string {
	e.require()
	return javaString(formatExceptionCharacter(e.Conversion) + " != " + e.ArgumentClass)
}
func (e *IllegalFormatConversionException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *IllegalFormatConversionException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalFormatConversionException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

type IllegalFormatFlagsException struct {
	*IllegalFormatException
	Flags string
}

func NewIllegalFormatFlagsException(flags string) *IllegalFormatFlagsException {
	return NewIllegalFormatFlagsExceptionNullable(javaString(flags))
}
func NewIllegalFormatFlagsExceptionNullable(flags *string) *IllegalFormatFlagsException {
	return &IllegalFormatFlagsException{IllegalFormatException: newIllegalFormatException(), Flags: requireFormatString(flags)}
}
func newIllegalFormatFlagsException(flags string) *IllegalFormatFlagsException {
	return NewIllegalFormatFlagsException(flags)
}
func (e *IllegalFormatFlagsException) GetFlags() string {
	e.require()
	return e.Flags
}
func (e *IllegalFormatFlagsException) GetMessage() *string {
	e.require()
	return javaString("Flags = '" + e.Flags + "'")
}
func (e *IllegalFormatFlagsException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *IllegalFormatFlagsException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalFormatFlagsException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

type IllegalFormatPrecisionException struct {
	*IllegalFormatException
	Precision int
}

func NewIllegalFormatPrecisionException(precision int) *IllegalFormatPrecisionException {
	return &IllegalFormatPrecisionException{IllegalFormatException: newIllegalFormatException(), Precision: int(int32(precision))}
}
func newIllegalFormatPrecisionException(precision int) *IllegalFormatPrecisionException {
	return NewIllegalFormatPrecisionException(precision)
}
func (e *IllegalFormatPrecisionException) GetPrecision() int {
	e.require()
	return e.Precision
}
func (e *IllegalFormatPrecisionException) GetMessage() *string {
	e.require()
	return javaString(strconv.Itoa(e.Precision))
}
func (e *IllegalFormatPrecisionException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *IllegalFormatPrecisionException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalFormatPrecisionException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

type IllegalFormatWidthException struct {
	*IllegalFormatException
	Width int
}

func NewIllegalFormatWidthException(width int) *IllegalFormatWidthException {
	return &IllegalFormatWidthException{IllegalFormatException: newIllegalFormatException(), Width: int(int32(width))}
}
func newIllegalFormatWidthException(width int) *IllegalFormatWidthException {
	return NewIllegalFormatWidthException(width)
}
func (e *IllegalFormatWidthException) GetWidth() int {
	e.require()
	return e.Width
}
func (e *IllegalFormatWidthException) GetMessage() *string {
	e.require()
	return javaString(strconv.Itoa(e.Width))
}
func (e *IllegalFormatWidthException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *IllegalFormatWidthException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalFormatWidthException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

type MissingFormatArgumentException struct {
	*IllegalFormatException
	Specifier string
}

func NewMissingFormatArgumentException(specifier string) *MissingFormatArgumentException {
	return NewMissingFormatArgumentExceptionNullable(javaString(specifier))
}
func NewMissingFormatArgumentExceptionNullable(specifier *string) *MissingFormatArgumentException {
	return &MissingFormatArgumentException{IllegalFormatException: newIllegalFormatException(), Specifier: requireFormatString(specifier)}
}
func newMissingFormatArgumentException(specifier string) *MissingFormatArgumentException {
	return NewMissingFormatArgumentException(specifier)
}
func (e *MissingFormatArgumentException) GetFormatSpecifier() string {
	e.require()
	return e.Specifier
}
func (e *MissingFormatArgumentException) GetMessage() *string {
	e.require()
	return javaString("Format specifier '" + e.Specifier + "'")
}
func (e *MissingFormatArgumentException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *MissingFormatArgumentException) Error() string { return javaThrowableMessage(e) }
func (e *MissingFormatArgumentException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

type MissingFormatWidthException struct {
	*IllegalFormatException
	Specifier string
}

func NewMissingFormatWidthException(specifier string) *MissingFormatWidthException {
	return NewMissingFormatWidthExceptionNullable(javaString(specifier))
}
func NewMissingFormatWidthExceptionNullable(specifier *string) *MissingFormatWidthException {
	return &MissingFormatWidthException{IllegalFormatException: newIllegalFormatException(), Specifier: requireFormatString(specifier)}
}
func newMissingFormatWidthException(specifier string) *MissingFormatWidthException {
	return NewMissingFormatWidthException(specifier)
}
func (e *MissingFormatWidthException) GetFormatSpecifier() string {
	e.require()
	return e.Specifier
}
func (e *MissingFormatWidthException) GetMessage() *string {
	e.require()
	return javaString(e.Specifier)
}
func (e *MissingFormatWidthException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *MissingFormatWidthException) Error() string { return javaThrowableMessage(e) }
func (e *MissingFormatWidthException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

type UnknownFormatConversionException struct {
	*IllegalFormatException
	Conversion string
}

func NewUnknownFormatConversionException(conversion string) *UnknownFormatConversionException {
	return NewUnknownFormatConversionExceptionNullable(javaString(conversion))
}
func NewUnknownFormatConversionExceptionNullable(conversion *string) *UnknownFormatConversionException {
	return &UnknownFormatConversionException{IllegalFormatException: newIllegalFormatException(), Conversion: requireFormatString(conversion)}
}
func newUnknownFormatConversionException(conversion string) *UnknownFormatConversionException {
	return NewUnknownFormatConversionException(conversion)
}
func (e *UnknownFormatConversionException) GetConversion() string {
	e.require()
	return e.Conversion
}
func (e *UnknownFormatConversionException) GetMessage() *string {
	e.require()
	return javaString("Conversion = '" + e.Conversion + "'")
}
func (e *UnknownFormatConversionException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *UnknownFormatConversionException) Error() string { return javaThrowableMessage(e) }
func (e *UnknownFormatConversionException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

type IllegalFormatArgumentIndexException struct {
	*IllegalFormatException
	Index int
}

func newIllegalFormatArgumentIndexException(index int) *IllegalFormatArgumentIndexException {
	return &IllegalFormatArgumentIndexException{newIllegalFormatException(), int(int32(index))}
}
func (e *IllegalFormatArgumentIndexException) GetIndex() int {
	e.require()
	return e.Index
}
func (e *IllegalFormatArgumentIndexException) GetMessage() *string {
	e.require()
	if e.Index == -1<<31 {
		return javaString("Format argument index: (not representable as int)")
	}
	return javaString("Illegal format argument index = " + strconv.Itoa(e.Index))
}
func (e *IllegalFormatArgumentIndexException) GetCause() error {
	e.require()
	return e.IllegalFormatException.GetCause()
}
func (e *IllegalFormatArgumentIndexException) Error() string { return javaThrowableMessage(e) }
func (e *IllegalFormatArgumentIndexException) require() {
	if e == nil || e.IllegalFormatException == nil {
		panic(NewNullPointerException())
	}
	e.IllegalFormatException.require()
}

func requireFormatString(value *string) string {
	if value == nil {
		panic(NewNullPointerException())
	}
	return *value
}

func formatExceptionCharacter(value uint16) string {
	return javaStringFromUTF16([]uint16{value})
}
