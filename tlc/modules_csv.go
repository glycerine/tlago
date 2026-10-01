package tlc

import (
	"bufio"
	"fmt"
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
		names := make([]string, len(record.Names))
		for i, name := range record.Names {
			names[i] = name.String()
		}
		lines = append(lines, strings.Join(names, delimiter.RawString()))
	}
	values := make([]string, len(record.Values))
	for i, value := range record.Values {
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
	params := make([]any, len(tuple.Elems))
	for i, value := range tuple.Elems {
		params[i] = value.String()
	}
	line, err := csvJavaStringFormat(format.RawString(), params)
	if err != nil {
		return nil, err
	}
	return BoolTrue, csvAppendLines(path.RawString(), []string{line})
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
	if _, err := os.Stat(path.RawString()); os.IsNotExist(err) {
		return EmptyTuple, nil
	} else if err != nil {
		return nil, err
	}

	names := make([]*UniqueString, len(tuple.Elems))
	for i, value := range tuple.Elems {
		str, ok := value.(*StringValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "first", "CSVRead", "sequence of strings", ValuesPPR(columns))
		}
		names[i] = str.Val
	}

	lines, err := csvReadLines(path.RawString())
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
	if _, err := os.Stat(path.RawString()); os.IsNotExist(err) {
		return IntZero, nil
	} else if err != nil {
		return nil, err
	}
	lines, err := csvReadLines(path.RawString())
	if err != nil {
		return nil, err
	}
	if len(lines) > math.MaxInt32 {
		return nil, newTLCErrorCode(ECTLCModuleOverflow, strconv.Itoa(len(lines)))
	}
	return NewIntValue(int32(len(lines))), nil
}

func csvStringArg(position string, operator string, value Value) (*StringValue, error) {
	str, ok := value.(*StringValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleArgumentError, position, operator, "string", ValuesPPR(value))
	}
	return str, nil
}

func csvAppendLines(path string, lines []string) error {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, line := range lines {
		if _, err := file.WriteString(line + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func csvReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var lines []string
	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadString('\n')
		if err != nil && err != io.EOF {
			return nil, err
		}
		if len(line) != 0 {
			line = strings.TrimSuffix(line, "\n")
			line = strings.TrimSuffix(line, "\r")
			lines = append(lines, line)
		}
		if err == io.EOF {
			break
		}
	}
	return lines, nil
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

func csvJavaStringFormat(format string, params []any) (string, error) {
	var converted strings.Builder
	for i := 0; i < len(format); i++ {
		if format[i] != '%' || i+1 >= len(format) {
			converted.WriteByte(format[i])
			continue
		}
		if format[i+1] == '%' {
			converted.WriteString("%%")
			i++
			continue
		}
		j := i + 1
		for j < len(format) && format[j] >= '0' && format[j] <= '9' {
			j++
		}
		if j > i+1 && j+1 < len(format) && format[j] == '$' && format[j+1] == 's' {
			converted.WriteString("%[")
			converted.WriteString(format[i+1 : j])
			converted.WriteString("]s")
			i = j + 1
			continue
		}
		converted.WriteByte(format[i])
	}
	return fmt.Sprintf(converted.String(), params...), nil
}
