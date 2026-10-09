package tlc

// Generic source diagnostic carriers are retained independently of transport.
// Native endpoint adapters use Go operation traits for discovery and retries.
type NetConnectException struct{ javaExceptionBase }

func NewNetConnectException(message ...string) *NetConnectException {
	return &NetConnectException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NetConnectException) Error() string { return javaThrowableMessage(e) }

type MalformedURLException struct{ javaExceptionBase }

func NewMalformedURLException(message ...string) *MalformedURLException {
	return &MalformedURLException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *MalformedURLException) Error() string { return javaThrowableMessage(e) }

type InterruptedException struct{ javaExceptionBase }

func NewInterruptedException(message ...string) *InterruptedException {
	return &InterruptedException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *InterruptedException) Error() string { return javaThrowableMessage(e) }

type NoRouteToHostException struct{ javaExceptionBase }

func NewNoRouteToHostException(message ...string) *NoRouteToHostException {
	return &NoRouteToHostException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NoRouteToHostException) Error() string { return javaThrowableMessage(e) }

type NetBindException struct{ javaExceptionBase }

func NewNetBindException(message ...string) *NetBindException {
	return &NetBindException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NetBindException) Error() string { return javaThrowableMessage(e) }
