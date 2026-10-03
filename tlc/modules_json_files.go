package tlc

import "bytes"

// The ordinary Json operators use FileWriter, whose encoder replaces coding
// errors. Their try-with-resources does not wrap the primary exception.
func jsonWriteOrdinaryFile(path *StringValue, payload []Value, newline bool) (err error) {
	if path == nil {
		return NewNullPointerException()
	}
	file, err := jsonOpenFileWriter(path.RawString())
	if err != nil {
		return err
	}
	writer := newJSONFileWriter(file, javaDefaultCharset())
	writer.replace = true
	defer func() {
		failure := recover()
		closeErr := writer.close()
		if failure != nil {
			if primary, ok := failure.(error); ok {
				javaSuppressCloseError(primary, closeErr)
			}
			panic(failure)
		}
		if err == nil {
			err = closeErr
		} else {
			javaSuppressCloseError(err, closeErr)
		}
	}()
	for _, value := range payload {
		var node bytes.Buffer
		if err := jsonWriteValue(&node, value); err != nil {
			return err
		}
		if newline {
			node.WriteByte('\n')
		}
		if err := writer.write(node.String()); err != nil {
			return err
		}
	}
	return nil
}
