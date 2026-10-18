package tlc

import (
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func CSVWriteRecord(parameter Value, delim Value, headers Value, absolutePath Value) (Value, error) {
	record := asRecordValue(parameter)
	if record == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "CSVWriteRecord", "record", ValuesPPR(parameter))
	}
	delimiter, err := csvStringArg("second", "CSVWriteRecord", delim)
	if err != nil {
		return nil, err
	}
	headerFlag, ok := headers.(*BoolValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, "third", "CSVWriteRecord", "boolean", ValuesPPR(headers))
	}
	path, err := csvStringArg("fourth", "CSVWriteRecord", absolutePath)
	if err != nil {
		return nil, err
	}

	record.DeepNormalize()
	var lines []string
	if headerFlag.Val {
		names := make([]string, len(record.ownedNames()))
		for i, name := range record.ownedNames() {
			names[i] = name.String()
		}
		lines = append(lines, strings.Join(names, delimiter.RawString()))
	}
	values := make([]string, len(record.ownedValues()))
	for i, value := range record.ownedValues() {
		values[i] = value.String()
	}
	lines = append(lines, strings.Join(values, delimiter.RawString()))

	return BoolTrue, csvAppendLines(path.RawString(), lines)
}

func CSVWrite(template Value, parameter Value, absolutePath Value) (Value, error) {
	format, err := csvStringArg("first", "CSVWrite", template)
	if err != nil {
		return nil, err
	}
	tuple := asTupleValue(parameter)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "CSVWrite", "sequence", ValuesPPR(parameter))
	}
	path, err := csvStringArg("third", "CSVWrite", absolutePath)
	if err != nil {
		return nil, err
	}

	tuple.DeepNormalize()
	params := make([]string, len(tuple.GetElems()))
	for i, value := range tuple.GetElems() {
		params[i] = value.String()
	}
	line, err := JavaFormatStrings(format.RawString(), params...)
	if err != nil {
		return nil, err
	}
	return BoolTrue, csvAppendLines(path.RawString(), []string{javaStringUTF8(line)})
}

func CSVRead(columns Value, delim Value, absolutePath Value) (Value, error) {
	tuple := asTupleValue(columns)
	if tuple == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "CSVRead", "sequence", ValuesPPR(columns))
	}
	delimiter, err := csvStringArg("second", "CSVRead", delim)
	if err != nil {
		return nil, err
	}
	path, err := csvStringArg("third", "CSVRead", absolutePath)
	if err != nil {
		return nil, err
	}
	filePath, err := ioUtilsTXTPath(path.RawString())
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(ioUtilsTXTSystemPath(filePath)); err != nil {
		return EmptyTuple, nil
	}

	names := make([]*UniqueString, len(tuple.GetElems()))
	for i, value := range tuple.GetElems() {
		var str *StringValue
		if !isNil(value) {
			var ok bool
			str, ok = asStringValue(value)
			if !ok {
				return nil, valueStreamClassCast(value, "tlc2.value.impl.StringValue")
			}
		}
		names[i] = str.ownedVal()
	}

	lines, err := csvReadLines(filePath)
	if err != nil {
		return nil, err
	}
	records := make([]Value, len(lines))
	for i, line := range lines {
		parts, err := csvRegexSplit(line, delimiter.RawString())
		if err != nil {
			return nil, err
		}
		values := make([]Value, len(parts))
		for j, part := range parts {
			values[j] = NewStringValue(part)
		}
		records[i] = NewRecordValue(names, values, false)
	}
	return NewTupleValue(records), nil
}

func CSVRecords(absolutePath Value) (Value, error) {
	path, err := csvStringArg("first", "CSVRecords", absolutePath)
	if err != nil {
		return nil, err
	}
	filePath, err := ioUtilsTXTPath(path.RawString())
	if err != nil {
		return nil, err
	}
	if _, err := os.Stat(ioUtilsTXTSystemPath(filePath)); err != nil {
		return IntZero, nil
	}
	count, err := csvCountLines(filePath)
	if err != nil {
		return nil, err
	}
	if count > math.MaxInt32 {
		return nil, newTLCErrorCode(ECTLCModuleOverflow, strconv.Itoa(count))
	}
	return NewIntValue(int32(count)), nil
}

func csvStringArg(position string, operator string, value Value) (*StringValue, error) {
	str, ok := asStringValue(value)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, position, operator, "string", ValuesPPR(value))
	}
	return str, nil
}

func csvAppendLines(path string, lines []string) error {
	for _, line := range lines {
		if err := csvAppendLine(path, line); err != nil {
			return err
		}
	}
	return nil
}

func csvAppendLine(path string, line string) (err error) {
	filePath, err := ioUtilsTXTPath(path)
	if err != nil {
		return err
	}
	data, err := javaCharsetEncode(line+javaFormatLineSeparator(), "UTF-8", false)
	if err != nil {
		return err
	}
	options, err := ioUtilsFileOptionsFromEnums([]string{"CREATE", "APPEND"})
	if err != nil {
		return err
	}
	return ioUtilsTXTWriteFile(filePath, data, options)
}

func csvReadLines(path string) ([]string, error) {
	data, err := ioUtilsTXTReadFile(path)
	if err != nil {
		return nil, err
	}
	text, err := javaCharsetDecode(data, "UTF-8", false)
	if err != nil {
		return nil, err
	}
	return javaBufferedReaderLines(text), nil
}

func csvCountLines(path string) (count int, err error) {
	file, err := ioUtilsTXTOpen(path, ioUtilsFileOptions{flag: os.O_RDONLY})
	if err != nil {
		return 0, err
	}
	defer func() {
		failure := recover()
		closeErr := file.Close()
		var uncheckedClose error
		if closeErr != nil {
			uncheckedClose = NewUncheckedIOExceptionFromCause(ioUtilsTXTChannelError(closeErr))
		}
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, uncheckedClose)
			}
			panic(failure)
		}
		if err == nil {
			err = uncheckedClose
		} else {
			javaSuppressCloseError(err, uncheckedClose)
		}
	}()
	data, err := io.ReadAll(file)
	if err != nil {
		return 0, NewUncheckedIOExceptionFromCause(ioUtilsTXTChannelError(err))
	}
	text, err := javaCharsetDecode(data, "UTF-8", false)
	if err != nil {
		return 0, NewUncheckedIOExceptionFromCause(err)
	}
	return len(javaBufferedReaderLines(text)), nil
}

func csvRegexSplit(line string, delimiter string) ([]string, error) {
	re, err := regexp.Compile(delimiter)
	if err != nil {
		return nil, err
	}
	parts := re.Split(line, -1)
	for len(parts) > 0 && parts[len(parts)-1] == "" {
		parts = parts[:len(parts)-1]
	}
	return parts, nil
}
