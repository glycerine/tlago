/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 *  http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
// Go port of the bundled Geronimo CommandMap and its Activation interface
// provider boundaries. Source: geronimo-activation_1.1_spec/src/main/java/javax/activation.
package tlc

import (
	"io"
	"sync"
	"sync/atomic"
)

// DataFlavor objects remain opaque JVM/AWT objects at this boundary. These
// callbacks carry all four DataContentHandler operations; native flavor classes
// and concrete bundled handler algorithms are separate required ports.
type MailDataContentHandler struct {
	TransferDataFlavorsFunc func() []any
	TransferDataFunc        func(any, *MailDataSource) (any, error)
	ContentFunc             func(*MailDataSource) (any, error)
	WriteToFunc             func(any, *string, io.Writer) error
}

func (h *MailDataContentHandler) GetTransferDataFlavors() []any {
	if h == nil {
		panic(NewNullPointerException())
	}
	if h.TransferDataFlavorsFunc == nil {
		panic(NewUnsupportedOperationException("DataContentHandler DataFlavor provider required"))
	}
	return h.TransferDataFlavorsFunc()
}
func (h *MailDataContentHandler) GetTransferData(flavor any, source *MailDataSource) (any, error) {
	return mailInvoke(func() (any, error) {
		if h == nil {
			return nil, NewNullPointerException()
		}
		if h.TransferDataFunc == nil {
			return nil, NewUnsupportedOperationException("DataContentHandler transfer provider required")
		}
		return h.TransferDataFunc(flavor, source)
	})
}
func (h *MailDataContentHandler) GetContent(source *MailDataSource) (any, error) {
	return mailInvoke(func() (any, error) {
		if h == nil {
			return nil, NewNullPointerException()
		}
		if h.ContentFunc == nil {
			return nil, NewUnsupportedOperationException("DataContentHandler content provider required")
		}
		return h.ContentFunc(source)
	})
}
func (h *MailDataContentHandler) WriteTo(value any, mime *string, out io.Writer) error {
	_, err := mailInvoke(func() (bool, error) {
		if h == nil {
			return false, NewNullPointerException()
		}
		if h.WriteToFunc == nil {
			return false, NewUnsupportedOperationException("DataContentHandler writeTo provider required")
		}
		return false, h.WriteToFunc(value, mime, out)
	})
	return err
}
func (h *MailDataContentHandler) AsMailDataContentHandler() *MailDataContentHandler { return h }

// Concrete carrier for CommandMap subclass overrides. The DataSource overloads
// delegate to the ordinary methods unless a subclass provides its own overload.
type MailCommandMap struct {
	PreferredFunc           func(*string) []*MailCommandInfo
	AllFunc                 func(*string) []*MailCommandInfo
	CommandFunc             func(*string, *string) *MailCommandInfo
	CreateHandlerFunc       func(*string) (*MailDataContentHandler, error)
	MIMETypesFunc           func() []string
	PreferredSourceFunc     func(*string, *MailDataSource) []*MailCommandInfo
	AllSourceFunc           func(*string, *MailDataSource) []*MailCommandInfo
	CommandSourceFunc       func(*string, *string, *MailDataSource) *MailCommandInfo
	CreateHandlerSourceFunc func(*string, *MailDataSource) (*MailDataContentHandler, error)
}

func NewMailCommandMap() *MailCommandMap {
	initializeMailCommandMap()
	return &MailCommandMap{}
}
func (m *MailCommandMap) GetPreferredCommands(mime *string, source ...*MailDataSource) []*MailCommandInfo {
	if m == nil {
		panic(NewNullPointerException())
	}
	if len(source) > 0 && m.PreferredSourceFunc != nil {
		return m.PreferredSourceFunc(mime, source[0])
	}
	return m.PreferredFunc(mime)
}
func (m *MailCommandMap) GetAllCommands(mime *string, source ...*MailDataSource) []*MailCommandInfo {
	if m == nil {
		panic(NewNullPointerException())
	}
	if len(source) > 0 && m.AllSourceFunc != nil {
		return m.AllSourceFunc(mime, source[0])
	}
	return m.AllFunc(mime)
}
func (m *MailCommandMap) GetCommand(mime, command *string, source ...*MailDataSource) *MailCommandInfo {
	if m == nil {
		panic(NewNullPointerException())
	}
	if len(source) > 0 && m.CommandSourceFunc != nil {
		return m.CommandSourceFunc(mime, command, source[0])
	}
	return m.CommandFunc(mime, command)
}
func (m *MailCommandMap) CreateDataContentHandler(mime *string, source ...*MailDataSource) (*MailDataContentHandler, error) {
	return mailInvoke(func() (*MailDataContentHandler, error) {
		if m == nil {
			return nil, NewNullPointerException()
		}
		if len(source) > 0 && m.CreateHandlerSourceFunc != nil {
			return m.CreateHandlerSourceFunc(mime, source[0])
		}
		return m.CreateHandlerFunc(mime)
	})
}
func (m *MailCommandMap) GetMIMETypes() []string {
	if m == nil {
		panic(NewNullPointerException())
	}
	if m.MIMETypesFunc == nil {
		return nil
	}
	return m.MIMETypesFunc()
}

var mailDefaultCommandMap atomic.Pointer[MailCommandMap]
var mailCommandMapInitialization, mailMailcapInitialization mailActivationClassInitialization

func initializeMailCommandMap() {
	mailCommandMapInitialization.initialize("javax.activation.CommandMap", nil, func() {
		mailDefaultCommandMap.Store(NewMailMailcapCommandMap().AsCommandMap())
	})
}
func initializeMailMailcapCommandMap() {
	mailMailcapInitialization.initialize("javax.activation.MailcapCommandMap", initializeMailCommandMap, nil)
}
func GetDefaultMailCommandMap() *MailCommandMap {
	initializeMailCommandMap()
	return mailDefaultCommandMap.Load()
}
func SetDefaultMailCommandMap(value *MailCommandMap) error {
	_, err := mailInvoke(func() (bool, error) {
		initializeMailCommandMap()
		if check := DefaultMailActivationEnvironment.CheckSetFactory; check != nil {
			if err := check(); err != nil {
				return false, err
			}
		}
		if value == nil {
			value = NewMailMailcapCommandMap().AsCommandMap()
		}
		// The source setter/getter are unsynchronized. Publication is independent
		// of construction; reentrant setters are allowed during the initializer.
		mailDefaultCommandMap.Store(value)
		return false, nil
	})
	return err
}

// JLS 12.4.2 class initialization for this superclass/subclass cycle. Each class
// keeps its own initializing owner, completion and erroneous state. Same-owner
// recursion returns normally; other owners wait with the initialization lock
// released. Superclass failure propagates unchanged; initializer Exception is
// wrapped once and Error propagates unchanged. These are not sync.Once semantics.
// https://docs.oracle.com/javase/specs/jls/se21/html/jls-12.html#jls-12.4.2
// Full JVM linking, assertion status and low-memory throwable allocation remain
// provider boundaries, as do JVM-generated stack frames and native thread names.
type mailActivationClassInitialization struct {
	mu      sync.Mutex
	ready   *sync.Cond
	state   uint8
	owner   uint64
	failure *ExceptionInInitializerError
}

func (c *mailActivationClassInitialization) initialize(name string, parent, body func()) {
	id := currentGoroutineID()
	c.mu.Lock()
	if c.ready == nil {
		c.ready = sync.NewCond(&c.mu)
	}
	for c.state == 1 && c.owner != id {
		c.ready.Wait()
	}
	if c.state == 1 || c.state == 2 {
		c.mu.Unlock()
		return
	}
	if c.state == 3 {
		cause := c.failure
		c.mu.Unlock()
		err := NewNoClassDefFoundError("Could not initialize class " + name)
		err.Cause = cause
		panic(err)
	}
	c.state, c.owner = 1, id
	c.mu.Unlock()
	_, original := mailInvoke(func() (bool, error) {
		if parent != nil {
			parent()
		}
		return false, nil
	})
	failure := original
	if original == nil {
		_, original = mailInvoke(func() (bool, error) {
			if body != nil {
				body()
			}
			return false, nil
		})
		failure = original
		if failure != nil && !isJavaError(failure) {
			failure = NewExceptionInInitializerError(failure)
		}
	}
	var snapshot *ExceptionInInitializerError
	if original != nil {
		thread := "main"
		if f := DefaultMailActivationEnvironment.ThreadName; f != nil {
			thread = f()
		}
		snapshot = NewExceptionInInitializerErrorMessage(javaString("Exception " + javaThrowableString(original) + " [in thread \"" + thread + "\"]"))
	}
	c.mu.Lock()
	c.owner = 0
	c.state = 2
	if failure != nil {
		c.state, c.failure = 3, snapshot
	}
	c.ready.Broadcast()
	c.mu.Unlock()
	if failure != nil {
		panic(failure)
	}
}
