package tlc

// Generic source diagnostic carriers are retained independently of transport.
// Native endpoint adapters use Go operation traits for discovery and retries.
type NetConnectException struct{ javaExceptionBase }

func NewNetConnectException(message ...string) *NetConnectException {
	return NewNetConnectExceptionNullable(optionalJavaMessage(message))
}
func NewNetConnectExceptionNullable(message *string) *NetConnectException {
	return &NetConnectException{newJavaExceptionBase(message, nil)}
}
func (e *NetConnectException) Error() string { return javaThrowableMessage(e) }
func (e *NetConnectException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *NetConnectException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *NetConnectException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type MalformedURLException struct{ javaExceptionBase }

func NewMalformedURLException(message ...string) *MalformedURLException {
	return NewMalformedURLExceptionNullable(optionalJavaMessage(message))
}
func NewMalformedURLExceptionNullable(message *string) *MalformedURLException {
	return &MalformedURLException{newJavaExceptionBase(message, nil)}
}
func (e *MalformedURLException) Error() string { return javaThrowableMessage(e) }
func (e *MalformedURLException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *MalformedURLException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *MalformedURLException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type InterruptedException struct{ javaExceptionBase }

func NewInterruptedException(message ...string) *InterruptedException {
	return NewInterruptedExceptionNullable(optionalJavaMessage(message))
}
func NewInterruptedExceptionNullable(message *string) *InterruptedException {
	return &InterruptedException{newJavaExceptionBase(message, nil)}
}
func (e *InterruptedException) Error() string { return javaThrowableMessage(e) }
func (e *InterruptedException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *InterruptedException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *InterruptedException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type NoRouteToHostException struct{ javaExceptionBase }

func NewNoRouteToHostException(message ...string) *NoRouteToHostException {
	return NewNoRouteToHostExceptionNullable(optionalJavaMessage(message))
}
func NewNoRouteToHostExceptionNullable(message *string) *NoRouteToHostException {
	return &NoRouteToHostException{newJavaExceptionBase(message, nil)}
}
func (e *NoRouteToHostException) Error() string { return javaThrowableMessage(e) }
func (e *NoRouteToHostException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *NoRouteToHostException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *NoRouteToHostException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}

type NetBindException struct{ javaExceptionBase }

func NewNetBindException(message ...string) *NetBindException {
	return NewNetBindExceptionNullable(optionalJavaMessage(message))
}
func NewNetBindExceptionNullable(message *string) *NetBindException {
	return &NetBindException{newJavaExceptionBase(message, nil)}
}
func (e *NetBindException) Error() string { return javaThrowableMessage(e) }
func (e *NetBindException) GetMessage() *string {
	e.require()
	return e.javaExceptionBase.GetMessage()
}
func (e *NetBindException) GetCause() error {
	e.require()
	return e.javaExceptionBase.GetCause()
}
func (e *NetBindException) require() {
	if e == nil {
		panic(NewNullPointerException())
	}
}
