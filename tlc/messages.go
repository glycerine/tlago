package tlc

import (
	"bytes"
	_ "embed"
)

//go:embed messages.properties
var messagesProperties []byte

var messages = func() *InsMap[string, string] {
	values := NewInsMap[string, string]()
	if err := loadJavaProperties(bytes.NewReader(messagesProperties), values); err != nil {
		panic(err)
	}
	return values
}()

// MessagesGetString ports tlc2.output.Messages.getString.
func MessagesGetString(key string) string {
	if value, ok := messages.Get2(key); ok {
		return value
	}
	return "!" + key + "!"
}
