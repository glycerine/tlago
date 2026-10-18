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
	NewOverrideIndex         func() ([]string, error)
	NewOverrideIndexNullable func() ([]*string, error)
}

var linkedNativeClasses = struct {
	sync.RWMutex
	classes map[string]NativeClass
}{classes: make(map[string]NativeClass)}

// RegisterNativeClass links an application's Go port of a Java override class.
// The returned function restores the previous binding, supporting source-style
// classloader isolation without changing the running tool's loaded classes.
func RegisterNativeClass(class NativeClass) func() {
	class = cloneNativeClass(class)
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
	monitor  distributedServerMonitor

	ownersCaptured bool
	packageOwner   string
	resolverOwner  FilenameToStream
}

func NewNativeClassLoader(pkg string, resolver FilenameToStream) *NativeClassLoader {
	return NewNativeClassLoaderNullable(javaString(pkg), resolver)
}

// NewNativeClassLoaderNullable preserves TLAClass's nullable Java String
// constructor boundary.
func NewNativeClassLoaderNullable(pkg *string, resolver FilenameToStream) *NativeClassLoader {
	if pkg == nil {
		panic(NewNullPointerException())
	}
	packageName := *pkg
	if packageName != "" && !strings.HasSuffix(packageName, ".") {
		packageName += "."
	}
	return &NativeClassLoader{
		Package: packageName, Resolver: resolver,
		ownersCaptured: true, packageOwner: packageName, resolverOwner: resolver,
	}
}

func (l *NativeClassLoader) ownedPackage() string {
	if l != nil && l.ownersCaptured {
		return l.packageOwner
	}
	if l == nil {
		return ""
	}
	return l.Package
}

func (l *NativeClassLoader) ownedResolver() FilenameToStream {
	if l != nil && l.ownersCaptured {
		return l.resolverOwner
	}
	if l == nil {
		return nil
	}
	return l.Resolver
}

func (l *NativeClassLoader) LoadClass(name string) (class *NativeClass) {
	return l.LoadClassNullable(javaString(name))
}

// LoadClassNullable preserves the source's nullable class-name behavior. The
// resolver and unqualified Class.forName attempts cannot load a null name; the
// final String concatenation searches the configured package plus "null".
func (l *NativeClassLoader) LoadClassNullable(name *string) (class *NativeClass) {
	// Java dispatch rejects a null TLAClass before entering loadClass, so its
	// catch(Throwable) cannot translate that failure into TLC_ERROR_REPLACING_MODULES.
	if l == nil {
		panic(NewNullPointerException())
	}
	// TLAClass.loadClass is synchronized for the complete three-stage lookup.
	l.monitor.Lock()
	defer l.monitor.Unlock()
	// TLAClass's outer catch(Throwable) follows the three lookup attempts;
	// only the resolver/URLClassLoader attempt has an inner catch(Exception).
	defer func() {
		if failure := recover(); failure != nil {
			className := javaNullableString(name)
			message := fmt.Sprint(failure)
			if throwable, ok := failure.(error); ok {
				message = javaThrowableString(throwable)
				if withMessage, ok := throwable.(interface{ GetMessage() *string }); ok {
					if value := withMessage.GetMessage(); value != nil {
						message = *value
					}
				}
			}
			panic(NewTLCRuntimeException(ECTLCErrorReplacingModules, className, message))
		}
	}()
	className := javaNullableString(name)
	if name == nil {
		// Java still resolves "null.class" before URLClassLoader.loadClass(null)
		// throws an Exception that the inner handler ignores.
		_ = l.resolveClassResource(className)
	} else {
		if class := l.loadLinkedClass(*name); class != nil {
			return class
		}
	}
	return l.loadLinkedClass(l.ownedPackage() + className)
}

func (l *NativeClassLoader) loadLinkedClass(name string) *NativeClass {
	linkedNativeClasses.RLock()
	class, exists := linkedNativeClasses.classes[name]
	linkedNativeClasses.RUnlock()
	if !exists {
		return nil
	}
	class = cloneNativeClass(class)
	resource := strings.ReplaceAll(name, ".", "/") + ".class"
	if resource := l.resolveClassResource(name); resource != nil {
		class.Resource = *resource
		return &class
	}
	// Class.forName searches the classpath rather than the resolver's user paths.
	classpath := &SimpleFilenameToStream{classpath: DefaultFilenameClasspath()}
	if resolver, ok := l.ownedResolver().(*SimpleFilenameToStream); ok && resolver != nil {
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

// cloneNativeClass preserves Class.getDeclaredMethods' fresh array and Java
// annotation immutability at the native linking boundary. Callback functions
// represent executable class members and therefore retain their identities.
func cloneNativeClass(class NativeClass) NativeClass {
	class.Methods = append([]NativeMethod(nil), class.Methods...)
	for i := range class.Methods {
		method := &class.Methods[i]
		if method.Operator != nil {
			annotation := *method.Operator
			method.Operator = &annotation
		}
		if method.Evaluation != nil {
			annotation := *method.Evaluation
			method.Evaluation = &annotation
		}
		if method.Callable != nil {
			annotation := *method.Callable
			method.Callable = &annotation
		}
	}
	return class
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
	resolver := l.ownedResolver()
	if isNil(resolver) {
		return nil
	}
	file := resolver.Resolve(name+".class", false)
	if file == nil || !file.Exists() {
		return nil
	}
	if uri := file.GetLibraryPath(); uri != nil {
		return uri
	}
	return javaString(filenameFileURI(file.GetAbsolutePath()))
}

func (m NativeMethod) methodValue(minLevel int) Value {
	m.requirePublicAccess()
	value := NewMethodValue(m.Signature, minLevel, m.Eval)
	value.ParameterCount = m.ParameterCount
	value.Instance = !m.Static
	value.DeclaringClass = nativeMethodDeclaringClass(m.Signature)
	value.captureMetadata()
	if m.ParameterCount == 0 && m.Final {
		result, err := value.Eval([]Value{}, EvalClear)
		if err != nil {
			panic(err)
		}
		return result
	}
	return value
}

func nativeMethodDeclaringClass(signature string) string {
	if args := strings.IndexByte(signature, '('); args >= 0 {
		signature = signature[:args]
	}
	fields := strings.Fields(signature)
	if len(fields) == 0 {
		return "Object"
	}
	qualifiedMethod := fields[len(fields)-1]
	if dot := strings.LastIndexByte(qualifiedMethod, '.'); dot >= 0 {
		declaringClass := qualifiedMethod[:dot]
		if packageDot := strings.LastIndexByte(declaringClass, '.'); packageDot >= 0 {
			return declaringClass[packageDot+1:]
		}
		return declaringClass
	}
	return "Object"
}

func (m NativeMethod) requirePublicAccess() {
	if !m.Public {
		panic(NewTLCRuntimeException(ECTLCModuleValueJavaMethodOverride, m.Signature, "method is not public"))
	}
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
