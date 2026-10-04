// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.

package tlc

import (
	"fmt"
	"strings"
	"sync"
)

// NativeMethod retains the source reflection metadata alongside a linked Go
// implementation. Declared methods, including ineligible and annotated methods,
// remain visible to SpecProcessor's registration checks.
type NativeMethod struct {
	Name, Signature       string
	Public, Static, Final bool
	ParameterCount        int
	Operator              *NativeOperatorAnnotation
	Evaluation            *NativeEvaluationAnnotation
	Callable              *NativeCallableAnnotation
	Eval                  OperatorEvalFunc
	Evaluate              EvaluatingEvalFunc
	Call                  CallableEvalFunc
}

type NativeOperatorAnnotation struct {
	Identifier, Module string
	MinLevel           int
	Warn               bool
}

type NativeEvaluationAnnotation struct {
	Definition, Module string
	MinLevel, Priority int
	Warn, Silent       bool
}

type NativeCallableAnnotation struct {
	Definition, Module string
	MinLevel           int
	Warn               bool
}

// NativeClass is the native counterpart of a loaded Java override class. Go
// callbacks must be linked by the application; Java class bytes are resources,
// not executable Go code. RequireResource preserves resolver/classpath discovery
// for adapters whose source provenance is a class file or archive.
type NativeClass struct {
	Name, Resource           string
	BuiltIn, RequireResource bool
	Methods                  []NativeMethod
	// NewOverrideIndex corresponds to instantiation and ITLCOverrides.get().
	// A nil callback denotes a class that does not implement ITLCOverrides.
	NewOverrideIndex func() ([]string, error)
}

var linkedNativeClasses = struct {
	sync.RWMutex
	classes map[string]NativeClass
}{classes: make(map[string]NativeClass)}

// RegisterNativeClass links an application's Go port of a Java override class.
// The returned function restores the previous binding, supporting source-style
// classloader isolation without changing the running tool's loaded classes.
func RegisterNativeClass(class NativeClass) func() {
	class.Methods = append([]NativeMethod(nil), class.Methods...)
	linkedNativeClasses.Lock()
	previous, existed := linkedNativeClasses.classes[class.Name]
	linkedNativeClasses.classes[class.Name] = class
	linkedNativeClasses.Unlock()
	return func() {
		linkedNativeClasses.Lock()
		defer linkedNativeClasses.Unlock()
		if existed {
			linkedNativeClasses.classes[class.Name] = previous
		} else {
			delete(linkedNativeClasses.classes, class.Name)
		}
	}
}

// NativeClassLoader ports TLAClass's resolver, unqualified classpath, and package
// fallback order. Its linked implementations are the native equivalent of the
// VM's class definitions; resource lookup retains original archive provenance.
type NativeClassLoader struct {
	Package  string
	Resolver FilenameToStream
}

func NewNativeClassLoader(pkg string, resolver FilenameToStream) *NativeClassLoader {
	if pkg != "" && !strings.HasSuffix(pkg, ".") {
		pkg += "."
	}
	return &NativeClassLoader{Package: pkg, Resolver: resolver}
}

func (l *NativeClassLoader) LoadClass(name string) (class *NativeClass) {
	// TLAClass's outer catch(Throwable) follows the three lookup attempts;
	// only the resolver/URLClassLoader attempt has an inner catch(Exception).
	defer func() {
		if failure := recover(); failure != nil {
			message := fmt.Sprint(failure)
			if throwable, ok := failure.(error); ok {
				message = javaThrowableString(throwable)
				if withMessage, ok := throwable.(interface{ GetMessage() *string }); ok {
					if value := withMessage.GetMessage(); value != nil {
						message = *value
					}
				}
			}
			panic(NewTLCRuntimeException(ECTLCErrorReplacingModules, name, message))
		}
	}()
	if class := l.loadLinkedClass(name); class != nil {
		return class
	}
	return l.loadLinkedClass(l.Package + name)
}

func (l *NativeClassLoader) loadLinkedClass(name string) *NativeClass {
	linkedNativeClasses.RLock()
	class, exists := linkedNativeClasses.classes[name]
	linkedNativeClasses.RUnlock()
	if !exists {
		return nil
	}
	class.Methods = append([]NativeMethod(nil), class.Methods...)
	resource := strings.ReplaceAll(name, ".", "/") + ".class"
	if resource := l.resolveClassResource(name); resource != nil {
		class.Resource = *resource
		return &class
	}
	// Class.forName searches the classpath rather than the resolver's user paths.
	classpath := &SimpleFilenameToStream{classpath: DefaultFilenameClasspath()}
	if resolver, ok := l.Resolver.(*SimpleFilenameToStream); ok {
		classpath.classpath = resolver.classpath
	}
	if found := classpath.findClasspath(resource); found != nil {
		class.Resource = found.location
		return &class
	}
	if class.RequireResource {
		return nil
	}
	if class.Resource == "" {
		class.Resource = "go:" + name
	}
	return &class
}

func (l *NativeClassLoader) resolveClassResource(name string) (resource *string) {
	defer func() {
		if failure := recover(); failure != nil {
			if exception, ok := failure.(error); ok && !isJavaError(exception) {
				resource = nil
				return
			}
			panic(failure)
		}
	}()
	if l.Resolver == nil {
		return nil
	}
	file := l.Resolver.Resolve(name+".class", false)
	if file == nil || !file.Exists() {
		return nil
	}
	if uri := file.GetLibraryPath(); uri != nil {
		return uri
	}
	return javaString(filenameFileURI(file.GetAbsolutePath()))
}

func (m NativeMethod) methodValue(minLevel int) Value {
	if !m.Public {
		panic(NewTLCRuntimeException(ECTLCModuleValueJavaMethodOverride, m.Signature, "method is not public"))
	}
	value := NewMethodValue(m.Signature, minLevel, m.Eval)
	value.ParameterCount = m.ParameterCount
	if m.ParameterCount == 0 && m.Final {
		result, err := value.Eval(nil, EvalClear)
		if err != nil {
			panic(err)
		}
		return result
	}
	return value
}

func nativeMethodDisplay(value Value) string {
	if _, ok := value.(*MethodValue); ok {
		return value.String()
	}
	// Java's annotated loader avoids potentially expensive value printing.
	name := fmt.Sprintf("%T", value)
	name = strings.TrimPrefix(name, "*tlc.")
	return "tlc2.value.impl." + name
}
