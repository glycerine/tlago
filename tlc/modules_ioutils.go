package tlc

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

var ioUtilsExecNames = []*UniqueString{
	UniqueStringOf("exitValue"),
	UniqueStringOf("stdout"),
	UniqueStringOf("stderr"),
}

func IOUtilsIOSerialize(value Value, absolutePath *StringValue, compress *BoolValue) (Value, error) {
	if absolutePath == nil {
		return nil, newTLCError(ECGeneral, "IOSerialize expected a string path")
	}
	path := absolutePath.RawString()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		return nil, err
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if value != nil {
		value.FingerPrint(0)
	}
	var writer io.Writer = file
	var gzipWriter *gzip.Writer
	if compress != nil && compress.Val {
		gzipWriter = gzip.NewWriter(file)
		writer = gzipWriter
	}
	out := NewValueOutputStream(writer)
	if err := out.WriteExternal(value); err != nil {
		return nil, err
	}
	if err := out.Close(); err != nil {
		return nil, err
	}
	return BoolTrue, nil
}

func IOUtilsIODeserialize(absolutePath *StringValue, compress *BoolValue) (Value, error) {
	if absolutePath == nil {
		return nil, newTLCError(ECGeneral, "IODeserialize expected a string path")
	}
	file, err := os.Open(absolutePath.RawString())
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var reader io.Reader = file
	var gzipReader *gzip.Reader
	if compress != nil && compress.Val {
		gzipReader, err = gzip.NewReader(file)
		if err != nil {
			return nil, err
		}
		defer gzipReader.Close()
		reader = gzipReader
	}
	return NewValueInputStream(reader).ReadExternal()
}

func IOUtilsSerialize(payload Value, dest Value, options Value) (Value, error) {
	opts := asRecordValue(options)
	if opts == nil {
		return ioUtilsResult(1, "", "Serialize error invalid parameters: options is not a record"), nil
	}
	format := ioUtilsRecordString(opts, "format")
	switch format {
	case "TXT":
		return ioUtilsSerializeTXT(payload, dest, opts), nil
	case "NDJSON":
		path, ok := dest.(*StringValue)
		if !ok {
			return nil, newTLCErrorCode(ECTLCModuleArgumentError, "second", "ndJsonSerialize", "sequence", ValuesPPR(dest))
		}
		return JsonTextSerialize(path, payload, options)
	default:
		return ValUndef, nil
	}
}

func IOUtilsDeserialize(src Value, options Value) (Value, error) {
	opts := asRecordValue(options)
	if opts == nil {
		return ioUtilsResult(1, "", "Deserialize error invalid parameters: options is not a record"), nil
	}
	if ioUtilsRecordString(opts, "format") != "TXT" {
		return ValUndef, nil
	}
	path, ok := src.(*StringValue)
	if !ok {
		return ioUtilsResult(1, "", "Deserialize error invalid parameters: source is not a string"), nil
	}
	data, err := os.ReadFile(path.RawString())
	if err != nil {
		return ioUtilsResult(1, "", "Deserialize error reading from the file: "+err.Error()), nil
	}
	return ioUtilsResult(0, string(data), ""), nil
}

func IOUtilsIOEnv() Value {
	env := os.Environ()
	sort.Strings(env)
	names := make([]*UniqueString, 0, len(env))
	values := make([]Value, 0, len(env))
	for _, entry := range env {
		key, value, ok := strings.Cut(entry, "=")
		if !ok {
			continue
		}
		names = append(names, UniqueStringOf(key))
		values = append(values, NewStringValue(value))
	}
	return NewRecordValue(names, values, false).Normalize()
}

func IOUtilsAtoi(value Value) (Value, error) {
	str, ok := value.(*StringValue)
	if !ok {
		return nil, newTLCError(ECGeneral, "atoi expected a string, got %s", value)
	}
	i, err := strconv.ParseInt(str.RawString(), 10, 32)
	if err != nil {
		return nil, newTLCError(ECGeneral, "atoi expected a string containing an integer, got %s", value)
	}
	return NewIntValue(int32(i)), nil
}

func IOUtilsIOExec(command Value) (Value, error) {
	argv, err := ioUtilsTupleStrings("IOExec", command)
	if err != nil {
		return nil, err
	}
	return ioUtilsRunProcess(nil, argv)
}

func IOUtilsIOEnvExec(env Value, command Value) (Value, error) {
	argv, err := ioUtilsTupleStrings("IOEnvExec", command)
	if err != nil {
		return nil, err
	}
	envMap, err := ioUtilsEnvRecord(env)
	if err != nil {
		return nil, err
	}
	return ioUtilsRunProcess(envMap, argv)
}

func IOUtilsIOExecTemplate(commandTemplate Value, parameters Value) (Value, error) {
	argv, err := ioUtilsTupleStrings("IOExecTemplate", commandTemplate)
	if err != nil {
		return nil, err
	}
	params, err := ioUtilsTupleStrings("IOExecTemplate", parameters)
	if err != nil {
		return nil, err
	}
	for i := range argv {
		argv[i] = ioUtilsJavaSprintf(argv[i], params)
	}
	return ioUtilsRunProcess(nil, argv)
}

func IOUtilsIOEnvExecTemplate(env Value, commandTemplate Value, parameters Value) (Value, error) {
	argv, err := ioUtilsTupleStrings("IOEnvExecTemplate", commandTemplate)
	if err != nil {
		return nil, err
	}
	params, err := ioUtilsTupleStrings("IOEnvExecTemplate", parameters)
	if err != nil {
		return nil, err
	}
	envMap, err := ioUtilsEnvRecord(env)
	if err != nil {
		return nil, err
	}
	for i := range argv {
		argv[i] = ioUtilsJavaSprintf(argv[i], params)
	}
	return ioUtilsRunProcess(envMap, argv)
}

func ioUtilsSerializeTXT(payload Value, dest Value, opts *RecordValue) Value {
	path, ok := dest.(*StringValue)
	if !ok {
		return ioUtilsResult(1, "", "Serialize error invalid parameters: destination is not a string")
	}
	text, ok := payload.(*StringValue)
	if !ok {
		return ioUtilsResult(1, "", "Serialize error invalid parameters: payload is not a string")
	}
	flag := ioUtilsOpenFileFlag(opts)
	filePath := path.RawString()
	if flag&os.O_CREATE != 0 {
		if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil && filepath.Dir(filePath) != "." {
			return ioUtilsResult(1, "", "Serialize error writing to the file: "+err.Error())
		}
	}
	file, err := os.OpenFile(filePath, flag, 0o644)
	if err != nil {
		return ioUtilsResult(1, "", "Serialize error writing to the file: "+err.Error())
	}
	defer file.Close()
	if _, err := file.WriteString(text.RawString()); err != nil {
		return ioUtilsResult(1, "", "Serialize error writing to the file: "+err.Error())
	}
	return ioUtilsResult(0, "Finish writing to the file with success!", "")
}

func ioUtilsOpenFileFlag(opts *RecordValue) int {
	openOptions := ioUtilsRecordTupleStrings(opts, "openOptions")
	flag := 0
	for _, opt := range openOptions {
		switch opt {
		case "WRITE":
			flag |= os.O_WRONLY
		case "CREATE":
			flag |= os.O_CREATE
		case "CREATE_NEW":
			flag |= os.O_CREATE | os.O_EXCL
		case "TRUNCATE_EXISTING":
			flag |= os.O_TRUNC
		case "APPEND":
			flag |= os.O_APPEND | os.O_WRONLY
		}
	}
	if flag == 0 {
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	}
	return flag
}

func ioUtilsResult(exitValue int32, stdout string, stderr string) Value {
	return NewRecordValue(ioUtilsExecNames, []Value{
		NewIntValue(exitValue),
		NewStringValue(stdout),
		NewStringValue(stderr),
	}, false)
}

func ioUtilsRecordString(record *RecordValue, key string) string {
	value, err := record.Select(NewStringValue(key))
	if err != nil {
		return ""
	}
	if str, ok := value.(*StringValue); ok {
		return str.RawString()
	}
	return ""
}

func ioUtilsRecordTupleStrings(record *RecordValue, key string) []string {
	value, err := record.Select(NewStringValue(key))
	if err != nil {
		return nil
	}
	tuple := asTupleValue(value)
	if tuple == nil {
		return nil
	}
	out := make([]string, 0, len(tuple.Elems))
	for _, elem := range tuple.Elems {
		str, ok := elem.(*StringValue)
		if !ok {
			continue
		}
		out = append(out, str.RawString())
	}
	return out
}

func ioUtilsTupleStrings(name string, value Value) ([]string, error) {
	tuple := asTupleValue(value)
	if tuple == nil {
		return nil, newTLCError(ECGeneral, "%s expected a sequence, got %s", name, value)
	}
	out := make([]string, 0, len(tuple.Elems))
	for _, elem := range tuple.Elems {
		str, ok := elem.(*StringValue)
		if !ok {
			return nil, newTLCError(ECGeneral, "%s expected a sequence of strings, got %s", name, elem)
		}
		out = append(out, str.RawString())
	}
	return out, nil
}

func ioUtilsEnvRecord(value Value) (map[string]string, error) {
	record := asRecordValue(value)
	if record == nil {
		return nil, newTLCError(ECGeneral, "IOEnvExec expected a record, got %s", value)
	}
	env := make(map[string]string, len(record.Names))
	for i, name := range record.Names {
		env[name.String()] = ioUtilsValueString(record.Values[i])
	}
	return env, nil
}

func ioUtilsValueString(value Value) string {
	if str, ok := value.(*StringValue); ok {
		return str.RawString()
	}
	return value.String()
}

func ioUtilsRunProcess(env map[string]string, argv []string) (Value, error) {
	if len(argv) == 0 {
		return ioUtilsResult(1, "", "empty command"), nil
	}
	cmd := exec.Command(argv[0], argv[1:]...)
	if env != nil {
		cmd.Env = os.Environ()
		keys := make([]string, 0, len(env))
		for key := range env {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			cmd.Env = append(cmd.Env, key+"="+env[key])
		}
	}
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exit := int32(0)
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exit = int32(exitErr.ExitCode())
		} else {
			return ioUtilsResult(1, stdout.String(), err.Error()), nil
		}
	}
	return ioUtilsResult(exit, stdout.String(), stderr.String()), nil
}

func ioUtilsJavaSprintf(format string, args []string) string {
	converted := format
	for i := range args {
		old := "%" + strconv.Itoa(i+1) + "$s"
		newFmt := "%[" + strconv.Itoa(i+1) + "]s"
		converted = strings.ReplaceAll(converted, old, newFmt)
	}
	values := make([]any, len(args))
	for i, arg := range args {
		values[i] = arg
	}
	return fmt.Sprintf(converted, values...)
}
