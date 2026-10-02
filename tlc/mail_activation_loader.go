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
// Bundled Geronimo MailcapCommandMap's ClassLoader/URL provider operations.
// Unlike current upstream trunk, the bundled jar calls loadClass/newInstance
// directly, with no ProviderLocator and no fallback after loading fails.
package tlc

import "sync/atomic"

// These carriers represent JVM class identity, loadClass/newInstance and resource
// enumeration separately. In particular, enumerations remain lazy across opens.
type MailActivationClass struct {
	NewInstance          func() (any, error)
	Name                 string
	Parents              []*MailActivationClass
	Primitive            bool
	IsAssignableFromFunc func(*MailActivationClass) bool
	HashCodeFunc         func() int32
	identityHash         atomic.Int32
}
type MailActivationResource struct {
	Location   string
	OpenStream func() (*MailInputStream, error)
}
type MailActivationResourceEnumeration struct {
	HasMoreElements func() (bool, error)
	NextElement     func() (*MailActivationResource, error)
}
type MailActivationClassLoader struct {
	GetResources func(string) (*MailActivationResourceEnumeration, error)
	LoadClass    func(*string) (*MailActivationClass, error)
}

func mailActivationResources(openers []func() (*MailInputStream, error)) *MailActivationResourceEnumeration {
	index := 0
	return &MailActivationResourceEnumeration{
		HasMoreElements: func() (bool, error) { return index < len(openers), nil },
		NextElement: func() (*MailActivationResource, error) {
			if index >= len(openers) {
				return nil, NewNoSuchElementException()
			}
			open := openers[index]
			index++
			return &MailActivationResource{OpenStream: open}, nil
		},
	}
}
func mailActivationDefaultLoader(e MailActivationEnvironment) *MailActivationClassLoader {
	return &MailActivationClassLoader{
		GetResources: func(name string) (*MailActivationResourceEnumeration, error) {
			openers, err := e.Resources(name)
			if err != nil {
				return nil, err
			}
			return mailActivationResources(openers), nil
		},
		LoadClass: func(name *string) (*MailActivationClass, error) {
			if name == nil {
				return nil, NewNullPointerException()
			}
			if class := mailNativeContentHandlerClass(*name); class != nil {
				return class, nil
			}
			// Other known bundled classes still need their native ports or the
			// full class-loader provider; do not fabricate class-not-found.
			return nil, NewUnsupportedOperationException("JVM ClassLoader provider required for " + *name)
		},
	}
}
func (m *MailMailcapCommandMap) CreateDataContentHandler(mimeType *string, source ...*MailDataSource) (*MailDataContentHandler, error) {
	return mailInvoke(func() (*MailDataContentHandler, error) {
		m.mu.Lock()
		defer m.mu.Unlock()
		// getCommand reenters the same Java monitor; loader and constructor
		// callbacks may themselves reenter command lookup or handler creation.
		info := m.GetCommand(mimeType, javaString("content-handler"))
		if info == nil {
			return nil, nil
		}
		loader, err := m.environment.ContextClassLoader()
		if err != nil {
			return nil, err
		}
		if loader == nil {
			loader, err = m.environment.DefiningClassLoader()
			if err != nil {
				return nil, err
			}
		}
		handler, err := mailInvoke(func() (*MailDataContentHandler, error) {
			if loader == nil {
				return nil, NewNullPointerException()
			}
			class, err := loader.LoadClass(info.GetCommandClass())
			if err != nil {
				return nil, err
			}
			if class == nil {
				return nil, NewNullPointerException()
			}
			value, err := class.NewInstance()
			if err != nil {
				return nil, err
			}
			if value == nil {
				return nil, nil
			}
			if handler, ok := value.(interface {
				AsMailDataContentHandler() *MailDataContentHandler
			}); ok {
				return handler.AsMailDataContentHandler(), nil
			}
			return nil, NewClassCastException()
		})
		switch err.(type) {
		case interface{ mailClassNotFoundException() }, interface{ mailIllegalAccessException() }, interface{ mailInstantiationException() }:
			return nil, nil
		}
		return handler, err
	})
}
