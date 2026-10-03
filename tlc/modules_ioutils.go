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
	charset, err := ioUtilsRecordRequiredString(opts, "charset")
	if err != nil {
		return ioUtilsResult(1, "", "Deserialize error invalid parameters: "+err.Error()), nil
	}
	// Java evaluates Paths.get and Charset.forName before Files.readString
	// opens the file. In particular, a missing file must not hide a bad charset.
	if strings.ContainsRune(path.RawString(), 0) {
		err := NewInvalidPathException(path.RawString(), "Nul character not allowed")
		return ioUtilsResult(1, "", "Deserialize error reading from the file: "+javaThrowableString(err)), nil
	}
	charset, err = javaCharsetName(charset)
	if err != nil {
		return ioUtilsResult(1, "", "Deserialize error reading from the file: "+javaThrowableString(err)), nil
	}
	data, err := os.ReadFile(path.RawString())
	if err != nil {
		return ioUtilsResult(1, "", "Deserialize error reading from the file: "+err.Error()), nil
	}
	text, err := ioUtilsDecodeString(data, charset)
	if err != nil {
		return ioUtilsResult(1, "", "Deserialize error reading from the file: "+javaThrowableString(err)), nil
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
	path, ok := dest.(*StringValue)
	if !ok {
		return ioUtilsResult(1, "", "Serialize error invalid parameters: destination is not a string")
	}
	text, ok := payload.(*StringValue)
	if !ok {
		return ioUtilsResult(1, "", "Serialize error invalid parameters: payload is not a string")
	}
	fileOptions, err := ioUtilsOpenFileOptions(opts)
	if err != nil {
		return ioUtilsResult(1, "", "Serialize error invalid parameters: "+err.Error())
	}
	charset, err := ioUtilsRecordRequiredString(opts, "charset")
	if err != nil {
		return ioUtilsResult(1, "", "Serialize error invalid parameters: "+err.Error())
	}
	if strings.ContainsRune(path.RawString(), 0) {
		err := NewInvalidPathException(path.RawString(), "Nul character not allowed")
		return ioUtilsResult(1, "", "Serialize error writing to the file: "+javaThrowableString(err))
	}
	data, err := javaCharsetEncode(text.RawString(), charset, true)
	if err != nil {
		return ioUtilsResult(1, "", "Serialize error writing to the file: "+javaThrowableString(err))
	}
	filePath := path.RawString()
	file, err := os.OpenFile(filePath, fileOptions.flag, 0o644)
	if err != nil {
		return ioUtilsResult(1, "", "Serialize error writing to the file: "+err.Error())
	}
	defer file.Close()
	if fileOptions.deleteOnClose {
		defer os.Remove(filePath)
	}
	if _, err := file.Write(data); err != nil {
		return ioUtilsResult(1, "", "Serialize error writing to the file: "+err.Error())
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
	tuple := asTupleValue(value)
	if tuple == nil {
		return ioUtilsFileOptions{}, fmt.Errorf("openOptions is not a sequence")
	}
	options := ioUtilsFileOptions{flag: os.O_WRONLY}
	if len(tuple.Elems) == 0 {
		options.flag |= os.O_CREATE | os.O_TRUNC
		return options, nil
	}
	sawWriteOrAppend := false
	sawCreate := false
	for _, opt := range tuple.Elems {
		str, ok := opt.(*StringValue)
		if !ok {
			return ioUtilsFileOptions{}, fmt.Errorf("openOptions contains a non-string value")
		}
		switch str.RawString() {
		case "WRITE":
			sawWriteOrAppend = true
		case "CREATE":
			options.flag |= os.O_CREATE
			sawCreate = true
		case "CREATE_NEW":
			options.flag |= os.O_CREATE | os.O_EXCL
			sawCreate = true
		case "TRUNCATE_EXISTING":
			options.flag |= os.O_TRUNC
		case "APPEND":
			options.flag |= os.O_APPEND
			sawWriteOrAppend = true
		case "DELETE_ON_CLOSE":
			options.deleteOnClose = true
		case "SPARSE", "SYNC", "DSYNC":
			// Java accepts these StandardOpenOption values. Their durability and
			// allocation hints are not visible at the TLA+ value level.
		case "READ":
			return ioUtilsFileOptions{}, fmt.Errorf("READ not allowed for writing")
		default:
			return ioUtilsFileOptions{}, fmt.Errorf("No enum constant java.nio.file.StandardOpenOption.%s", str.RawString())
		}
	}
	if !sawWriteOrAppend && !sawCreate && options.flag == os.O_WRONLY {
		options.flag |= os.O_CREATE | os.O_TRUNC
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
