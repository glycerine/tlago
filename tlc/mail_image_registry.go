/*
 * Copyright (c) 2000, 2022, Oracle and/or its affiliates. All rights reserved.
 * DO NOT ALTER OR REMOVE COPYRIGHT NOTICES OR THIS FILE HEADER.
 *
 * This code is free software; you can redistribute it and/or modify it
 * under the terms of the GNU General Public License version 2 only, as
 * published by the Free Software Foundation.  Oracle designates this
 * particular file as subject to the "Classpath" exception as provided
 * by Oracle in the LICENSE file that accompanied this code.
 *
 * This code is distributed in the hope that it will be useful, but WITHOUT
 * ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
 * FITNESS FOR A PARTICULAR PURPOSE.  See the GNU General Public License
 * version 2 for more details (a copy is included in the LICENSE file that
 * accompanied this code).
 *
 * You should have received a copy of the GNU General Public License version
 * 2 along with this work; if not, write to the Free Software Foundation,
 * Inc., 51 Franklin St, Fifth Floor, Boston, MA 02110-1301 USA.
 *
 * Please contact Oracle, 500 Oracle Parkway, Redwood Shores, CA 94065 USA
 * or visit www.oracle.com if you need additional information or have any
 * questions.
 */

// OpenJDK ServiceRegistry, SubRegistry and FilterIterator. Java objects that
// participate in this registry are represented by MailImageSPI carriers with
// explicit leaf-Class identities; metadata loading and VM security remain
// native runtime boundaries.
package tlc

import (
	"fmt"
	"io"
	"os"
)

// These hooks represent VM operations, not registry policy. Configure them
// before registry use. The native default runs without a SecurityManager;
// secured capture/privileged execution requires its runtime provider.
type MailImageRegistryEnvironment struct {
	CaptureAccessContext   func() (any, error)
	SecurityManagerPresent func() bool
	DoPrivileged           func(action func() error, context any) error
	SystemErr              io.Writer
	PrintStackTrace        func(error) error
	// contextLoader distinguishes ServiceLoader.load(Class) from load(Class,Loader),
	// including an explicitly null loader. Discovery itself remains lazy.
	LookupProviders func(*MailActivationClass, *MailActivationClassLoader, bool) (*MailImageSPIIterator, error)
}

var DefaultMailImageRegistryEnvironment MailImageRegistryEnvironment

// A captured native unsecured context has presence and identity. Protection
// domains and stack permission checks belong to the secured runtime provider.
type mailImageUnsecuredContext struct{ marker byte }

func mailImageRegistryCall(call func() error) error {
	_, err := mailInvoke(func() (bool, error) { return false, call() })
	return err
}
func mailNewImageSPIClass(name string, parents []*MailActivationClass) *MailActivationClass {
	return &MailActivationClass{Name: name, Parents: parents, NewInstance: func() (any, error) {
		return nil, NewUnsupportedOperationException("JVM Image SPI class construction provider required for " + name)
	}}
}
func mailImageRegistryClass(name string) *MailActivationClass {
	for _, c := range []*MailActivationClass{mailImageRegisterableServiceClass, mailImageIIOServiceProviderClass, mailImageReaderWriterSPIClass, mailImageReaderSPIClass, mailImageWriterSPIClass, mailImageTranscoderSPIClass, mailImageInputStreamSPIClass, mailImageOutputStreamSPIClass} {
		if c.Name == name {
			return c
		}
	}
	for _, list := range [][]*MailActivationClass{mailStandardImageReaderClasses, mailStandardImageWriterClasses} {
		for _, c := range list {
			if c.Name == name {
				return c
			}
		}
	}
	return nil
}
func mailImageCheckClassAllowed(c *MailActivationClass) error {
	if c == nil {
		return NewIllegalArgumentException("class must not be null")
	}
	if c != mailImageInputStreamSPIClass && c != mailImageOutputStreamSPIClass && c != mailImageReaderSPIClass && c != mailImageTranscoderSPIClass && c != mailImageWriterSPIClass {
		return NewIllegalArgumentException(c.GetName() + " is not an ImageIO SPI class")
	}
	return nil
}

// An iterator of Java Class identities. Registry category iterators are live
// HashMap key iterators, including removal without provider deregistration.
type MailImageClassIterator struct {
	HasNextFunc func() (bool, error)
	NextFunc    func() (*MailActivationClass, error)
	RemoveFunc  func() error
}

func (i *MailImageClassIterator) HasNext() (bool, error) {
	return mailInvoke(func() (bool, error) {
		if i == nil {
			return false, NewNullPointerException()
		}
		return i.HasNextFunc()
	})
}
func (i *MailImageClassIterator) Next() (*MailActivationClass, error) {
	return mailInvoke(func() (*MailActivationClass, error) {
		if i == nil {
			return nil, NewNullPointerException()
		}
		return i.NextFunc()
	})
}
func (i *MailImageClassIterator) Remove() error {
	return mailImageRegistryCall(func() error {
		if i == nil {
			return NewNullPointerException()
		}
		if i.RemoveFunc == nil {
			return NewUnsupportedOperationException()
		}
		return i.RemoveFunc()
	})
}
func mailImageClassList(classes []*MailActivationClass) *MailImageClassIterator {
	index := 0
	return &MailImageClassIterator{HasNextFunc: func() (bool, error) { return index < len(classes), nil }, NextFunc: func() (*MailActivationClass, error) {
		if index >= len(classes) {
			return nil, NewNoSuchElementException()
		}
		c := classes[index]
		index++
		return c, nil
	}}
}

// Java synchronizes each SubRegistry, allowing callback reentry. The native
// category guard and per-operation iterator guards also prevent Go data races;
// they retain live nodes, source modCount checks and callback boundaries.
type MailImageServiceRegistry struct {
	categoryMonitor distributedServerMonitor
	categories      mailObjectHashMap
}
type mailImageSubRegistry struct {
	monitor             distributedServerMonitor
	registry            *MailImageServiceRegistry
	category            *MailActivationClass
	poset               mailImagePartiallyOrderedSet
	providers, contexts mailObjectHashMap
}

func NewMailImageServiceRegistry(categories *MailImageClassIterator) (*MailImageServiceRegistry, error) {
	return mailInvoke(func() (*MailImageServiceRegistry, error) {
		if categories == nil {
			return nil, NewIllegalArgumentException("categories == null!")
		}
		r := &MailImageServiceRegistry{}
		for {
			more, err := categories.HasNext()
			if err != nil {
				return nil, err
			}
			if !more {
				return r, nil
			}
			c, err := categories.Next()
			if err != nil {
				return nil, err
			}
			if err = mailImageCheckClassAllowed(c); err != nil {
				return nil, err
			}
			r.categories.put(c, &mailImageSubRegistry{registry: r, category: c})
		}
	})
}
func (r *MailImageServiceRegistry) category(c *MailActivationClass) *mailImageSubRegistry {
	if r == nil {
		panic(NewNullPointerException())
	}
	r.categoryMonitor.Lock()
	defer r.categoryMonitor.Unlock()
	v := r.categories.get(c)
	if v == nil {
		return nil
	}
	return v.(*mailImageSubRegistry)
}
func (r *MailImageServiceRegistry) GetCategories() (*MailImageClassIterator, error) {
	return mailInvoke(func() (*MailImageClassIterator, error) {
		if r == nil {
			return nil, NewNullPointerException()
		}
		r.categoryMonitor.Lock()
		iter := r.categories.iterator()
		r.categoryMonitor.Unlock()
		return &MailImageClassIterator{
			HasNextFunc: func() (bool, error) {
				r.categoryMonitor.Lock()
				defer r.categoryMonitor.Unlock()
				return iter.hasNext(), nil
			},
			NextFunc: func() (*MailActivationClass, error) {
				r.categoryMonitor.Lock()
				defer r.categoryMonitor.Unlock()
				return iter.nextNode().objectKey.(*MailActivationClass), nil
			},
			RemoveFunc: func() error { r.categoryMonitor.Lock(); defer r.categoryMonitor.Unlock(); iter.remove(); return nil },
		}, nil
	})
}
func mailImageProviderClass(p *MailImageSPI) *MailActivationClass {
	if p == nil {
		panic(NewNullPointerException())
	}
	return p.mailObjectClass()
}
func mailImageProviderIsInstance(c *MailActivationClass, p *MailImageSPI) bool {
	return p != nil && c.IsAssignableFrom(mailImageProviderClass(p))
}
func (r *MailImageServiceRegistry) subRegistries(p *MailImageSPI) ([]*mailImageSubRegistry, error) {
	iter, err := r.GetCategories()
	if err != nil {
		return nil, err
	}
	var list []*mailImageSubRegistry
	for {
		more, err := iter.HasNext()
		if err != nil {
			return nil, err
		}
		if !more {
			return list, nil
		}
		c, err := iter.Next()
		if err != nil {
			return nil, err
		}
		if mailImageProviderIsInstance(c, p) {
			list = append(list, r.category(c))
		}
	}
}
func (r *MailImageServiceRegistry) RegisterServiceProviderInCategory(p *MailImageSPI, c *MailActivationClass) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if r == nil {
			return false, NewNullPointerException()
		}
		if p == nil {
			return false, NewIllegalArgumentException("provider == null!")
		}
		sub := r.category(c)
		if sub == nil {
			return false, NewIllegalArgumentException("category unknown!")
		}
		if !mailImageProviderIsInstance(c, p) {
			return false, NewClassCastException()
		}
		return sub.register(p)
	})
}
func (r *MailImageServiceRegistry) RegisterServiceProvider(p *MailImageSPI) error {
	return mailImageRegistryCall(func() error {
		if r == nil {
			return NewNullPointerException()
		}
		if p == nil {
			return NewIllegalArgumentException("provider == null!")
		}
		list, err := r.subRegistries(p)
		if err != nil {
			return err
		}
		for _, sub := range list {
			if _, err := sub.register(p); err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *MailImageServiceRegistry) RegisterServiceProviders(providers *MailImageSPIIterator) error {
	return mailImageRegistryCall(func() error {
		if r == nil {
			return NewNullPointerException()
		}
		if providers == nil {
			return NewIllegalArgumentException("provider == null!")
		}
		for {
			more, err := providers.HasNext()
			if err != nil {
				return err
			}
			if !more {
				return nil
			}
			p, err := providers.Next()
			if err != nil {
				return err
			}
			if err = r.RegisterServiceProvider(p); err != nil {
				return err
			}
		}
	})
}
func (r *MailImageServiceRegistry) DeregisterServiceProviderInCategory(p *MailImageSPI, c *MailActivationClass) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if r == nil {
			return false, NewNullPointerException()
		}
		if p == nil {
			return false, NewIllegalArgumentException("provider == null!")
		}
		sub := r.category(c)
		if sub == nil {
			return false, NewIllegalArgumentException("category unknown!")
		}
		if !mailImageProviderIsInstance(c, p) {
			return false, NewClassCastException()
		}
		return sub.deregister(p)
	})
}
func (r *MailImageServiceRegistry) DeregisterServiceProvider(p *MailImageSPI) error {
	return mailImageRegistryCall(func() error {
		if r == nil {
			return NewNullPointerException()
		}
		if p == nil {
			return NewIllegalArgumentException("provider == null!")
		}
		list, err := r.subRegistries(p)
		if err != nil {
			return err
		}
		for _, sub := range list {
			if _, err := sub.deregister(p); err != nil {
				return err
			}
		}
		return nil
	})
}
func (r *MailImageServiceRegistry) Contains(p *MailImageSPI) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if r == nil {
			return false, NewNullPointerException()
		}
		if p == nil {
			return false, NewIllegalArgumentException("provider == null!")
		}
		list, err := r.subRegistries(p)
		if err != nil {
			return false, err
		}
		for _, sub := range list {
			if sub.contains(p) {
				return true, nil
			}
		}
		return false, nil
	})
}
func (r *MailImageServiceRegistry) GetServiceProviders(c *MailActivationClass, ordered bool) (*MailImageSPIIterator, error) {
	return mailInvoke(func() (*MailImageSPIIterator, error) {
		sub := r.category(c)
		if sub == nil {
			return nil, NewIllegalArgumentException("category unknown!")
		}
		return sub.iterator(ordered), nil
	})
}
func (r *MailImageServiceRegistry) GetServiceProvidersFiltered(c *MailActivationClass, filter func(*MailImageSPI) (bool, error), ordered bool) (*MailImageSPIIterator, error) {
	return mailInvoke(func() (*MailImageSPIIterator, error) {
		if r.category(c) == nil {
			return nil, NewIllegalArgumentException("category unknown!")
		}
		source, err := r.GetServiceProviders(c, ordered)
		if err != nil {
			return nil, err
		}
		f := &mailImageRegistryFilter{source: source, filter: filter}
		if err = f.advance(); err != nil {
			return nil, err
		}
		return &MailImageSPIIterator{HasNextFunc: func() (bool, error) { return f.next != nil, nil }, NextFunc: f.nextProvider, RemoveFunc: func() error { return NewUnsupportedOperationException() }}, nil
	})
}
func (r *MailImageServiceRegistry) GetServiceProviderByClass(c *MailActivationClass) (*MailImageSPI, error) {
	return mailInvoke(func() (*MailImageSPI, error) {
		if r == nil {
			return nil, NewNullPointerException()
		}
		if c == nil {
			return nil, NewIllegalArgumentException("providerClass == null!")
		}
		iter, err := r.GetCategories()
		if err != nil {
			return nil, err
		}
		for {
			more, err := iter.HasNext()
			if err != nil {
				return nil, err
			}
			if !more {
				return nil, nil
			}
			category, err := iter.Next()
			if err != nil {
				return nil, err
			}
			if category.IsAssignableFrom(c) {
				sub := r.category(category)
				p := sub.byClass(c)
				if p != nil {
					return p, nil
				}
			}
		}
	})
}
func (r *MailImageServiceRegistry) ordering(c *MailActivationClass, first, second *MailImageSPI, unset bool) (bool, error) {
	return mailInvoke(func() (bool, error) {
		if r == nil {
			return false, NewNullPointerException()
		}
		if first == nil || second == nil {
			return false, NewIllegalArgumentException("provider is null!")
		}
		if first == second {
			return false, NewIllegalArgumentException("providers are the same!")
		}
		sub := r.category(c)
		if sub == nil {
			return false, NewIllegalArgumentException("category unknown!")
		}
		if sub.contains(first) && sub.contains(second) {
			sub.monitor.Lock()
			defer sub.monitor.Unlock()
			if unset {
				return sub.poset.unsetOrdering(first, second), nil
			}
			return sub.poset.setOrdering(first, second), nil
		}
		return false, nil
	})
}
func (r *MailImageServiceRegistry) SetOrdering(c *MailActivationClass, first, second *MailImageSPI) (bool, error) {
	return r.ordering(c, first, second, false)
}
func (r *MailImageServiceRegistry) UnsetOrdering(c *MailActivationClass, first, second *MailImageSPI) (bool, error) {
	return r.ordering(c, first, second, true)
}
func (r *MailImageServiceRegistry) DeregisterAllInCategory(c *MailActivationClass) error {
	return mailImageRegistryCall(func() error {
		sub := r.category(c)
		if sub == nil {
			return NewIllegalArgumentException("category unknown!")
		}
		return sub.clear()
	})
}
func (r *MailImageServiceRegistry) DeregisterAll() error {
	return mailImageRegistryCall(func() error {
		if r == nil {
			return NewNullPointerException()
		}
		r.categoryMonitor.Lock()
		iter := r.categories.iterator()
		r.categoryMonitor.Unlock()
		for {
			r.categoryMonitor.Lock()
			more := iter.hasNext()
			r.categoryMonitor.Unlock()
			if !more {
				return nil
			}
			sub, err := mailInvoke(func() (*mailImageSubRegistry, error) {
				r.categoryMonitor.Lock()
				defer r.categoryMonitor.Unlock()
				return iter.nextNode().objectValue.(*mailImageSubRegistry), nil
			})
			if err != nil {
				return err
			}
			if err = sub.clear(); err != nil {
				return err
			}
		}
	})
}

// JVM finalization scheduling remains with the process runtime. This is the
// source finalizer body, without an unsolicited Go GC finalizer.
func (r *MailImageServiceRegistry) Finalize() error { return r.DeregisterAll() }

func (s *mailImageSubRegistry) register(p *MailImageSPI) (bool, error) {
	if s == nil {
		return false, NewNullPointerException()
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	old, _ := s.providers.get(mailImageProviderClass(p)).(*MailImageSPI)
	present := old != nil
	if present {
		if _, err := s.deregister(old); err != nil {
			return false, err
		}
	}
	s.providers.put(mailImageProviderClass(p), p)
	c := mailImageProviderClass(p) // evaluate key before AccessController.getContext
	context, err := mailImageCaptureAccessContext()
	if err != nil {
		return false, err
	}
	s.contexts.put(c, context)
	s.poset.add(p)
	if mailImageRegisterableServiceClass.IsAssignableFrom(mailImageProviderClass(p)) {
		failure := mailImageRegistryCall(func() error {
			if p.OnRegistrationFunc != nil {
				return p.OnRegistrationFunc(s.registry, s.category)
			}
			return nil
		})
		if failure != nil {
			if err := mailImagePrintRegistrationFailure(failure); err != nil {
				return false, err
			}
		}
	}
	return !present, nil
}
func (s *mailImageSubRegistry) deregister(p *MailImageSPI) (bool, error) {
	if s == nil {
		return false, NewNullPointerException()
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	old, _ := s.providers.get(mailImageProviderClass(p)).(*MailImageSPI)
	if p == old {
		s.providers.remove(mailImageProviderClass(p))
		s.contexts.remove(mailImageProviderClass(p))
		s.poset.remove(p)
		if mailImageRegisterableServiceClass.IsAssignableFrom(mailImageProviderClass(p)) && p.OnDeregistrationFunc != nil {
			if err := p.OnDeregistrationFunc(s.registry, s.category); err != nil {
				return false, err
			}
		}
		return true, nil
	}
	return false, nil
}
func (s *mailImageSubRegistry) contains(p *MailImageSPI) bool {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	old, _ := s.providers.get(mailImageProviderClass(p)).(*MailImageSPI)
	return old == p
}
func (s *mailImageSubRegistry) byClass(c *MailActivationClass) *MailImageSPI {
	if s == nil {
		panic(NewNullPointerException())
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	p, _ := s.providers.get(c).(*MailImageSPI)
	return p
}
func (s *mailImageSubRegistry) iterator(ordered bool) *MailImageSPIIterator {
	s.monitor.Lock()
	defer s.monitor.Unlock()
	if ordered {
		i := s.poset.iterator()
		return &MailImageSPIIterator{
			HasNextFunc: func() (bool, error) { s.monitor.Lock(); defer s.monitor.Unlock(); return i.hasNext(), nil },
			NextFunc:    func() (*MailImageSPI, error) { s.monitor.Lock(); defer s.monitor.Unlock(); return i.next(), nil },
			RemoveFunc:  func() error { return NewUnsupportedOperationException() },
		}
	}
	i := s.providers.iterator()
	return &MailImageSPIIterator{
		HasNextFunc: func() (bool, error) { s.monitor.Lock(); defer s.monitor.Unlock(); return i.hasNext(), nil },
		NextFunc: func() (*MailImageSPI, error) {
			s.monitor.Lock()
			defer s.monitor.Unlock()
			return i.nextNode().objectValue.(*MailImageSPI), nil
		},
		RemoveFunc: func() error { s.monitor.Lock(); defer s.monitor.Unlock(); i.remove(); return nil },
	}
}
func (s *mailImageSubRegistry) clear() error {
	if s == nil {
		return NewNullPointerException()
	}
	s.monitor.Lock()
	defer s.monitor.Unlock()
	i := s.providers.iterator()
	for i.hasNext() {
		p := i.nextNode().objectValue.(*MailImageSPI)
		i.remove()
		if mailImageRegisterableServiceClass.IsAssignableFrom(mailImageProviderClass(p)) {
			context := s.contexts.get(mailImageProviderClass(p))
			if !mailHandlerNull(context) || !mailImageSecurityManagerPresent() {
				action := func() error {
					if p.OnDeregistrationFunc != nil {
						return p.OnDeregistrationFunc(s.registry, s.category)
					}
					return nil
				}
				env := DefaultMailImageRegistryEnvironment
				if env.DoPrivileged != nil {
					if err := env.DoPrivileged(action, context); err != nil {
						return err
					}
				} else {
					if mailImageSecurityManagerPresent() {
						return NewUnsupportedOperationException("JVM secured AccessController.doPrivileged provider required")
					}
					if err := action(); err != nil {
						return err
					}
				}
			}
		}
	}
	s.poset.clear()
	s.contexts.clear()
	return nil
}
func (s *mailImageSubRegistry) finalize() error {
	return mailImageRegistryCall(func() error {
		if s == nil {
			return NewNullPointerException()
		}
		s.monitor.Lock()
		defer s.monitor.Unlock()
		return s.clear()
	})
}
func mailImageSecurityManagerPresent() bool {
	cb := DefaultMailImageRegistryEnvironment.SecurityManagerPresent
	return cb != nil && cb()
}
func mailImageCaptureAccessContext() (any, error) {
	if cb := DefaultMailImageRegistryEnvironment.CaptureAccessContext; cb != nil {
		return cb()
	}
	if mailImageSecurityManagerPresent() {
		return nil, NewUnsupportedOperationException("JVM secured AccessController.getContext provider required")
	}
	return &mailImageUnsecuredContext{}, nil
}
func mailImagePrintRegistrationFailure(failure error) error {
	env := DefaultMailImageRegistryEnvironment
	out := env.SystemErr
	if out == nil {
		out = os.Stderr
	}
	// PrintStream suppresses checked/native I/O failures, but typed Java
	// unchecked failures and native writer panics escape the callback catch.
	_, err := fmt.Fprintln(out, "Caught and handled this exception :")
	if _, typed := err.(interface{ GetMessage() *string }); typed && !isJavaIOException(err) {
		return err
	}
	if env.PrintStackTrace != nil {
		return env.PrintStackTrace(failure)
	}
	_, err = fmt.Fprint(out, javaThrowableStackTrace(failure))
	if _, typed := err.(interface{ GetMessage() *string }); typed && !isJavaIOException(err) {
		return err
	}
	return nil
}

type mailImageRegistryFilter struct {
	source *MailImageSPIIterator
	filter func(*MailImageSPI) (bool, error)
	next   *MailImageSPI
}

func (f *mailImageRegistryFilter) advance() error {
	for {
		more, err := f.source.HasNext()
		if err != nil {
			return err
		}
		if !more {
			f.next = nil
			return nil
		}
		p, err := f.source.Next()
		if err != nil {
			return err
		}
		if f.filter == nil {
			return NewNullPointerException()
		}
		found, err := f.filter(p)
		if err != nil {
			return err
		}
		if found {
			f.next = p
			return nil
		}
	}
}
func (f *mailImageRegistryFilter) nextProvider() (*MailImageSPI, error) {
	if f.next == nil {
		return nil, NewNoSuchElementException()
	}
	p := f.next
	if err := f.advance(); err != nil {
		return nil, err
	}
	return p, nil
}
func LookupMailImageProviders(c *MailActivationClass) (*MailImageSPIIterator, error) {
	return mailImageLookupProviders(c, nil, true)
}
func LookupMailImageProvidersWithLoader(c *MailActivationClass, loader *MailActivationClassLoader) (*MailImageSPIIterator, error) {
	return mailImageLookupProviders(c, loader, false)
}
func mailImageLookupProviders(c *MailActivationClass, loader *MailActivationClassLoader, context bool) (*MailImageSPIIterator, error) {
	return mailInvoke(func() (*MailImageSPIIterator, error) {
		if c == nil {
			return nil, NewIllegalArgumentException("providerClass == null!")
		}
		if err := mailImageCheckClassAllowed(c); err != nil {
			return nil, err
		}
		if cb := DefaultMailImageRegistryEnvironment.LookupProviders; cb != nil {
			return cb(c, loader, context)
		}
		// ServiceLoader.iterator() is lazy. Do not report no providers merely
		// because full module/classpath discovery is not supplied by the VM adapter.
		return &MailImageSPIIterator{
			HasNextFunc: func() (bool, error) {
				return false, NewUnsupportedOperationException("JVM ServiceLoader discovery provider required")
			},
			NextFunc: func() (*MailImageSPI, error) {
				return nil, NewUnsupportedOperationException("JVM ServiceLoader discovery provider required")
			},
		}, nil
	})
}
