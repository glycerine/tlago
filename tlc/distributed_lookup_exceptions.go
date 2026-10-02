package tlc

// NetConnectException is java.net.ConnectException, distinct from the
// java.rmi.ConnectException already used by remote endpoint calls.
type NetConnectException struct{ javaExceptionBase }

func NewNetConnectException(message ...string) *NetConnectException {
	return &NetConnectException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NetConnectException) Error() string { return javaThrowableMessage(e) }

type NotBoundException struct{ javaExceptionBase }

func NewNotBoundException(message ...string) *NotBoundException {
	return &NotBoundException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NotBoundException) Error() string { return javaThrowableMessage(e) }

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

type RMIUnknownHostException struct{ *RemoteException }

func NewRMIUnknownHostException(message string, cause error) *RMIUnknownHostException {
	return &RMIUnknownHostException{NewRemoteException(javaString(message), cause)}
}
func (e *RMIUnknownHostException) Error() string { return javaThrowableMessage(e) }

type ConnectIOException struct{ *RemoteException }

func NewConnectIOException(message string, cause error) *ConnectIOException {
	return &ConnectIOException{NewRemoteException(javaString(message), cause)}
}
func (e *ConnectIOException) Error() string { return javaThrowableMessage(e) }

type NoRouteToHostException struct{ javaExceptionBase }

func NewNoRouteToHostException(message ...string) *NoRouteToHostException {
	return &NoRouteToHostException{newJavaExceptionBase(optionalJavaMessage(message), nil)}
}
func (e *NoRouteToHostException) Error() string { return javaThrowableMessage(e) }
