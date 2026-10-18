package tlc

import (
	"os"
	"sync"
)

// Charset.defaultCharset reads StaticProperty.fileEncoding, not later changes
// to System.getProperties. The Go command's startup -D handling supplies that
// initial value before any default encoder is constructed.
var javaDefaultCharsetState = struct {
	sync.Mutex
	encoding string
	resolved string
}{encoding: javaInitialFileEncoding()}

func javaInitialFileEncoding() string {
	if value, present := os.LookupEnv("file.encoding"); present {
		return value
	}
	return "UTF-8"
}

func javaSetStartupFileEncoding(value string) {
	javaDefaultCharsetState.Lock()
	defer javaDefaultCharsetState.Unlock()
	if javaDefaultCharsetState.resolved == "" {
		javaDefaultCharsetState.encoding = value
	}
}

func javaDefaultCharset() string {
	javaDefaultCharsetState.Lock()
	defer javaDefaultCharsetState.Unlock()
	if javaDefaultCharsetState.resolved == "" {
		name, err := javaCharsetName(javaDefaultCharsetState.encoding)
		if err != nil {
			// The standard provider returns null for an unknown initial name,
			// and Charset.defaultCharset falls back to UTF-8. Extended provider
			// codecs and COMPAT/native locale discovery remain source work.
			name = "UTF-8"
		}
		javaDefaultCharsetState.resolved = name
	}
	return javaDefaultCharsetState.resolved
}
