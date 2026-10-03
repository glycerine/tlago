package tlc

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
)

var ioUtilsExecNames = []*UniqueString{
	UniqueStringOf("exitValue"),
	UniqueStringOf("stdout"),
	UniqueStringOf("stderr"),
}

var ioUtilsEnvSnapshot = ioUtilsBuildEnv()

func IOUtilsIOSerialize(value Value, absolutePath *StringValue, compress *BoolValue) (Value, error) {
	if absolutePath == nil {
		return nil, newTLCError(ECGeneral, "IOSerialize expected a string path")
	}
	file, err := os.Create(absolutePath.RawString())
	if err != nil {
		return nil, err
	}
	out := NewValueOutputStreamWithCompression(file, compress != nil && compress.Val)
	if err := out.WriteExternal(value); err != nil {
		_ = out.Close()
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
	in, err := NewValueInputStreamWithCompression(file, compress != nil && compress.Val)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	defer in.Close()
	return in.ReadExternal()
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
	opts, err := ioUtilsTXTRecord(options)
	if err != nil {
		return ioUtilsTXTFailure("Deserialize", "invalid parameters", err), nil
	}
	format, err := ioUtilsTXTFormat(opts)
	if err != nil {
		return nil, err
	}
	if format != "TXT" {
		return ValUndef, nil
	}
	path, err := ioUtilsTXTString(src)
	if err != nil {
		return ioUtilsTXTFailure("Deserialize", "invalid parameters", err), nil
	}
	charset, err := ioUtilsTXTStringField(opts, "charset")
	if err != nil {
		return ioUtilsTXTFailure("Deserialize", "invalid parameters", err), nil
	}
	if path == nil || charset == nil {
		return ioUtilsTXTFailure("Deserialize", "reading from the file", NewNullPointerException()), nil
	}
	if strings.ContainsRune(path.RawString(), 0) {
		return ioUtilsTXTFailure("Deserialize", "reading from the file", NewInvalidPathException(path.RawString(), "Nul character not allowed")), nil
	}
	name, err := javaCharsetName(charset.RawString())
	if err != nil {
		return ioUtilsTXTFailure("Deserialize", "reading from the file", err), nil
	}
	data, err := os.ReadFile(path.RawString())
	if err != nil {
		return ioUtilsTXTFailure("Deserialize", "reading from the file", err), nil
	}
	text, err := ioUtilsDecodeString(data, name)
	if err != nil {
		return ioUtilsTXTFailure("Deserialize", "reading from the file", err), nil
	}
	return ioUtilsResult(0, text, ""), nil
}

func IOUtilsIOEnv() Value {
	return ioUtilsEnvSnapshot
}

func ioUtilsBuildEnv() Value {
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
	if value == nil {
		// Java reaches v.toString() while building the argument-error parameters.
		return nil, NewNullPointerException()
	}
	var str *StringValue
	switch value := value.(type) {
	case *StringValue:
		str = value
	case *DebuggerValue:
		str = value.StringValue
	default:
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "atoi", "string", ValuesPPR(value))
	}
	i, ok := javaParseDecimalInt(str.RawString())
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "atoi", "string", ValuesPPR(value))
	}
	return NewIntValue(i), nil
}

func IOUtilsIOExec(command Value) (Value, error) {
	argv, err := ioUtilsTupleStrings("IOExec", command)
	if err != nil {
		return nil, err
	}
	return ioUtilsRunProcess(nil, argv)
}

func IOUtilsIOEnvExec(env Value, command Value) (Value, error) {
	environment, err := ioUtilsRequireEnvRecord("IOEnvExec", env)
	if err != nil {
		return nil, err
	}
	argv, err := ioUtilsTupleStrings("IOEnvExec", command)
	if err != nil {
		return nil, err
	}
	return ioUtilsRunProcess(ioUtilsEnvRecord(environment), argv)
}

func IOUtilsIOExecTemplate(commandTemplate Value, parameters Value) (Value, error) {
	argv, err := ioUtilsTemplateStrings("IOExecTemplate", commandTemplate, parameters)
	if err != nil {
		return nil, err
	}
	return ioUtilsRunProcess(nil, argv)
}

func IOUtilsIOEnvExecTemplate(env Value, commandTemplate Value, parameters Value) (Value, error) {
	environment, err := ioUtilsRequireEnvRecord("IOEnvExecTemplate", env)
	if err != nil {
		return nil, err
	}
	argv, err := ioUtilsTemplateStrings("IOEnvExecTemplate", commandTemplate, parameters)
	if err != nil {
		return nil, err
	}
	return ioUtilsRunProcess(ioUtilsEnvRecord(environment), argv)
}

func ioUtilsSerializeTXT(payload Value, dest Value, opts *RecordValue) Value {
	text, err := ioUtilsTXTString(payload)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "invalid parameters", err)
	}
	path, err := ioUtilsTXTString(dest)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "invalid parameters", err)
	}
	options, charset, err := ioUtilsTXTOptions(opts)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "invalid parameters", err)
	}
	if path == nil {
		return ioUtilsTXTFailure("Serialize", "writing to the file", NewNullPointerException())
	}
	if strings.ContainsRune(path.RawString(), 0) {
		return ioUtilsTXTFailure("Serialize", "writing to the file", NewInvalidPathException(path.RawString(), "Nul character not allowed"))
	}
	if text == nil || charset == nil {
		return ioUtilsTXTFailure("Serialize", "writing to the file", NewNullPointerException())
	}
	name, err := javaCharsetName(charset.RawString())
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "writing to the file", err)
	}
	enums, err := ioUtilsTXTEnums(options)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "writing to the file", err)
	}
	data, err := javaCharsetEncode(text.RawString(), name, true)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "writing to the file", err)
	}
	fileOptions, err := ioUtilsFileOptionsFromEnums(enums)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "writing to the file", err)
	}
	filePath := path.RawString()
	file, err := os.OpenFile(filePath, fileOptions.flag, 0o666)
	if err != nil {
		return ioUtilsTXTFailure("Serialize", "writing to the file", err)
	}
	defer file.Close()
	if fileOptions.deleteOnClose {
		defer os.Remove(filePath)
	}
	if _, err := file.Write(data); err != nil {
		return ioUtilsTXTFailure("Serialize", "writing to the file", err)
	}
	return ioUtilsResult(0, "Finish writing to the file with success!", "")
}

type ioUtilsFileOptions struct {
	flag          int
	deleteOnClose bool
}

func ioUtilsOpenFileOptions(opts *RecordValue) (ioUtilsFileOptions, error) {
	value, err := opts.Apply(NewStringValue("openOptions"))
	if err != nil {
		return ioUtilsFileOptions{}, err
	}
	tuple, err := ioUtilsTXTTuple(value)
	if err != nil {
		return ioUtilsFileOptions{}, err
	}
	if tuple == nil {
		return ioUtilsFileOptions{}, NewNullPointerException()
	}
	strings := make([]*StringValue, len(tuple.Elems))
	for i, value := range tuple.Elems {
		strings[i], err = ioUtilsTXTString(value)
		if err != nil {
			return ioUtilsFileOptions{}, err
		}
	}
	enums, err := ioUtilsTXTEnums(strings)
	if err != nil {
		return ioUtilsFileOptions{}, err
	}
	return ioUtilsFileOptionsFromEnums(enums)
}

func ioUtilsFileOptionsFromEnums(enums []string) (ioUtilsFileOptions, error) {
	options := ioUtilsFileOptions{flag: os.O_WRONLY}
	if len(enums) == 0 {
		options.flag |= os.O_CREATE | os.O_TRUNC
		return options, nil
	}
	appendMode, truncate := false, false
	for _, name := range enums {
		switch name {
		case "READ":
			return ioUtilsFileOptions{}, NewIllegalArgumentException("READ not allowed")
		case "CREATE":
			options.flag |= os.O_CREATE
		case "CREATE_NEW":
			options.flag |= os.O_CREATE | os.O_EXCL
		case "TRUNCATE_EXISTING":
			options.flag |= os.O_TRUNC
			truncate = true
		case "APPEND":
			options.flag |= os.O_APPEND
			appendMode = true
		case "DELETE_ON_CLOSE":
			options.deleteOnClose = true
		case "SYNC":
			options.flag |= os.O_SYNC
		case "DSYNC":
			options.flag |= ioUtilsDataSyncFlag()
		}
	}
	if appendMode && truncate {
		return ioUtilsFileOptions{}, NewIllegalArgumentException("APPEND + TRUNCATE_EXISTING not allowed")
	}
	return options, nil
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

func ioUtilsRecordRequiredString(record *RecordValue, key string) (string, error) {
	value, err := record.Apply(NewStringValue(key))
	if err != nil {
		return "", err
	}
	str, ok := value.(*StringValue)
	if !ok {
		return "", fmt.Errorf("%s is not a string", key)
	}
	return str.RawString(), nil
}

func ioUtilsRequireTuple(name string, value Value) (*TupleValue, error) {
	// IOUtils checks instanceof TupleValue rather than converting a function
	// to a tuple. Both template tuple checks precede element conversion.
	tuple, ok := value.(*TupleValue)
	if !ok {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, name, "sequence", ValuesPPR(value))
	}
	return tuple, nil
}

func ioUtilsTupleStrings(name string, value Value) ([]string, error) {
	tuple, err := ioUtilsRequireTuple(name, value)
	if err != nil {
		return nil, err
	}
	return ioUtilsConvertStrings(tuple)
}

func ioUtilsConvertStrings(tuple *TupleValue) ([]string, error) {
	out := make([]string, len(tuple.Elems))
	for i, elem := range tuple.Elems {
		var str *StringValue
		switch value := elem.(type) {
		case *StringValue:
			str = value
		case *DebuggerValue:
			str = value.StringValue
		default:
			return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, "IOExec", "sequence", ValuesPPR(elem))
		}
		out[i] = str.RawString()
	}
	return out, nil
}

func ioUtilsTemplateStrings(name string, commandTemplate Value, parameters Value) ([]string, error) {
	commandTuple, err := ioUtilsRequireTuple(name, commandTemplate)
	if err != nil {
		return nil, err
	}
	parameterTuple, err := ioUtilsRequireTuple(name, parameters)
	if err != nil {
		return nil, err
	}
	argv, err := ioUtilsConvertStrings(commandTuple)
	if err != nil {
		return nil, err
	}
	params, err := ioUtilsConvertStrings(parameterTuple)
	if err != nil {
		return nil, err
	}
	for i := range argv {
		argv[i], err = JavaFormatStrings(argv[i], params...)
		if err != nil {
			return nil, err
		}
	}
	return argv, nil
}

func ioUtilsRequireEnvRecord(name string, value Value) (*RecordValue, error) {
	record := asRecordValue(value)
	if record == nil {
		return nil, newTLCErrorCode(ECTLCModuleOneArgumentError, name, "record", ValuesPPR(value))
	}
	return record, nil
}

func ioUtilsEnvRecord(record *RecordValue) map[string]string {
	env := make(map[string]string, len(record.Names))
	for i, name := range record.Names {
		env[name.String()] = ioUtilsValueString(record.Values[i])
	}
	return env
}

func ioUtilsValueString(value Value) string {
	switch value := value.(type) {
	case *StringValue:
		return value.UnquotedString()
	case *DebuggerValue:
		return value.UnquotedString()
	default:
		return value.String()
	}
}

func ioUtilsRunProcess(env map[string]string, argv []string) (Value, error) {
	if len(argv) == 0 {
		return nil, newTLCError(ECGeneral, "java.lang.ArrayIndexOutOfBoundsException: Index 0 out of bounds for length 0")
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
			return nil, err
		}
	}
	return ioUtilsResult(exit, stdout.String(), stderr.String()), nil
}

// JSON's stream writer replaces unencodable input; Files.writeString in the
// TXT handler instead calls javaCharsetEncode with no replacement.
func ioUtilsEncodeString(text string, charset string) ([]byte, error) {
	return javaCharsetEncode(text, charset, false)
}

func ioUtilsDecodeString(data []byte, charset string) (string, error) {
	return javaCharsetDecode(data, charset, false)
}
