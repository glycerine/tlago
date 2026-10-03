// TXT argument and exception boundaries from CommunityModules IOUtils.
// The upstream MIT license is retained in test_vectors/CommunityModules/LICENSE.
package tlc

// The TXT overrides use Java casts, not the value conversion operations.
func ioUtilsTXTString(value Value) (*StringValue, error) {
	if value == nil {
		return nil, nil
	}
	switch value := value.(type) {
	case *StringValue:
		return value, nil
	case *DebuggerValue:
		return value.StringValue, nil
	default:
		return nil, valueStreamClassCast(value, "tlc2.value.impl.StringValue")
	}
}

func ioUtilsTXTRecord(value Value) (*RecordValue, error) {
	if value == nil {
		return nil, nil
	}
	if record, ok := value.(*RecordValue); ok {
		return record, nil
	}
	return nil, valueStreamClassCast(value, "tlc2.value.impl.RecordValue")
}

func ioUtilsTXTTuple(value Value) (*TupleValue, error) {
	if value == nil {
		return nil, nil
	}
	if tuple, ok := value.(*TupleValue); ok {
		return tuple, nil
	}
	return nil, valueStreamClassCast(value, "tlc2.value.impl.TupleValue")
}

func ioUtilsTXTField(record *RecordValue, name string) (Value, error) {
	if record == nil {
		return nil, NewNullPointerException()
	}
	return record.Apply(NewStringValue(name))
}

func ioUtilsTXTStringField(record *RecordValue, name string) (*StringValue, error) {
	value, err := ioUtilsTXTField(record, name)
	if err != nil {
		return nil, err
	}
	return ioUtilsTXTString(value)
}

func ioUtilsTXTFormat(record *RecordValue) (string, error) {
	value, err := ioUtilsTXTStringField(record, "format")
	if err != nil {
		return "", err
	}
	if value == nil {
		return "", NewNullPointerException()
	}
	return value.RawString(), nil
}

func ioUtilsTXTFailure(operation, stage string, err error) Value {
	if isJavaError(err) {
		panic(err)
	}
	return ioUtilsResult(1, "", operation+" error "+stage+": "+javaThrowableString(err))
}

// Only the Java catch(Exception) regions consume evaluation failures. Error
// remains outside those catches even when a Go evaluator returns it as error.
func ioUtilsTXTEval(tool *Tool, node SemanticNode, con *Context, state, pstate *TLCStateMut, control int, cm CostModel) (value Value, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			if exception, ok := failure.(error); ok && !isJavaError(exception) {
				value, err = nil, exception
			} else {
				panic(failure)
			}
		}
	}()
	value, err = tool.Eval(node, con, state, pstate, control, cm)
	if err != nil && isJavaError(err) {
		panic(err)
	}
	return
}

func ioUtilsTXTOptions(record *RecordValue) ([]*StringValue, *StringValue, error) {
	value, err := ioUtilsTXTField(record, "openOptions")
	if err != nil {
		return nil, nil, err
	}
	tuple, err := ioUtilsTXTTuple(value)
	if err != nil {
		return nil, nil, err
	}
	charset, err := ioUtilsTXTStringField(record, "charset")
	if err != nil {
		return nil, nil, err
	}
	if tuple == nil {
		return nil, nil, NewNullPointerException()
	}
	options := make([]*StringValue, len(tuple.Elems))
	for i, value := range tuple.Elems {
		options[i], err = ioUtilsTXTString(value)
		if err != nil {
			return nil, nil, err
		}
	}
	return options, charset, nil
}

func ioUtilsTXTEnums(values []*StringValue) ([]string, error) {
	options := make([]string, len(values))
	for i, value := range values {
		if value == nil {
			return nil, NewNullPointerException()
		}
		name := value.RawString()
		switch name {
		case "READ", "WRITE", "APPEND", "TRUNCATE_EXISTING", "CREATE", "CREATE_NEW", "DELETE_ON_CLOSE", "SPARSE", "SYNC", "DSYNC":
			options[i] = name
		default:
			return nil, NewIllegalArgumentException("No enum constant java.nio.file.StandardOpenOption." + name)
		}
	}
	return options, nil
}
