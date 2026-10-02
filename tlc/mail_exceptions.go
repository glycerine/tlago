package tlc

import (
	"sync"
	"sync/atomic"
)

// These are the JavaMail/JNDI families caught by util.MailSender. The mail
// protocol and address providers return these concrete carriers at their
// checked exception boundaries.
type NamingException struct{ javaExceptionBase }

func NewNamingException(message ...string) *NamingException {
	return &NamingException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NamingException) Error() string { return javaThrowableMessage(e) }

type mailNextException struct{ failure error }
type MessagingException struct {
	javaExceptionBase
	mu   sync.Mutex
	next atomic.Pointer[mailNextException]
}

func NewMessagingException(message *string, next error) *MessagingException {
	e := &MessagingException{javaExceptionBase: newJavaExceptionBase(message, nil)}
	if next != nil {
		e.next.Store(&mailNextException{next})
	}
	return e
}
func (e *MessagingException) Error() string           { return javaThrowableMessage(e) }
func (e *MessagingException) GetNextException() error { return e.GetCause() }
func (e *MessagingException) GetCause() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.nextException()
}
func (e *MessagingException) Unwrap() error { return e.GetCause() }
func (e *MessagingException) nextException() error {
	if next := e.next.Load(); next != nil {
		return next.failure
	}
	return nil
}

// MessagingException.getCause is its next-exception field. setNextException
// walks only MessagingException links and appends at their first empty link.
func (e *MessagingException) SetNextException(next error) bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	for current := e; ; {
		failure := current.nextException()
		if failure == nil {
			if next != nil {
				current.next.Store(&mailNextException{next})
			} else {
				current.next.Store(nil)
			}
			return true
		}
		current = javaMessagingException(failure)
		if current == nil {
			return false
		}
	}
}

type SendFailedException struct{ *MessagingException }

type MailParseException struct{ *MessagingException }

func NewMailParseException(message string) *MailParseException {
	return &MailParseException{NewMessagingException(javaString(message), nil)}
}
func (e *MailParseException) Error() string { return javaThrowableMessage(e) }

type UnsupportedEncodingException struct{ *IOException }

func NewUnsupportedEncodingException(message string) *UnsupportedEncodingException {
	return &UnsupportedEncodingException{NewIOException(message)}
}
func (e *UnsupportedEncodingException) Error() string { return javaThrowableMessage(e) }

func NewSendFailedException(message *string, next error) *SendFailedException {
	return &SendFailedException{NewMessagingException(message, next)}
}
func (e *SendFailedException) Error() string { return javaThrowableMessage(e) }

type MailAddressException struct {
	*MessagingException
	Ref      *string
	Position int
}

func NewMailAddressException(message, ref *string, position int) *MailAddressException {
	return &MailAddressException{NewMessagingException(message, nil), copyJavaMessage(ref), position}
}
func (e *MailAddressException) Error() string   { return javaThrowableMessage(e) }
func (e *MailAddressException) GetRef() *string { return copyJavaMessage(e.Ref) }
func (e *MailAddressException) GetPos() int     { return e.Position }

func javaMessagingException(err error) *MessagingException {
	switch e := err.(type) {
	case *MessagingException:
		return e
	case *SendFailedException:
		return e.MessagingException
	case *MailAddressException:
		return e.MessagingException
	case *MailParseException:
		return e.MessagingException
	}
	return nil
}

type NumberFormatException struct{ *IllegalArgumentException }

func NewNumberFormatException(value string) *NumberFormatException {
	return &NumberFormatException{NewIllegalArgumentException("For input string: \"" + value + "\"")}
}
func (e *NumberFormatException) Error() string { return javaThrowableMessage(e) }
