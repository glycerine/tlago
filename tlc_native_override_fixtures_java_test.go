/*******************************************************************************
 * Copyright (c) 2019, 2020 Microsoft Research. All rights reserved.
 *
 * The MIT License (MIT)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
 * of the Software, and to permit persons to whom the Software is furnished to do
 * so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN
 * AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 * WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 ******************************************************************************/
package tlago

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/glycerine/tlago/tlc"
)

// Java fixture implementations, not alternate TLA definitions. The source
// class/annotation metadata travels through the production loader and processor.
func linkJavaNativeOverrideFixtures(t *testing.T) {
	t.Helper()
	path, err := filepath.Abs("tlc/test_vectors/native_overrides")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("CLASSPATH", filepath.Join(path, "UserModuleOverrideFromJar.jar")+string(os.PathListSeparator)+path)
	t.Setenv("tlc2.overrides.TLCOverrides", "tlc2.overrides.TLCOverrides"+string(os.PathListSeparator)+"tlc2.overrides.TLCTestOverrides")
	link := func(class tlc.NativeClass) { t.Cleanup(tlc.RegisterNativeClass(class)) }
	for _, name := range []string{"UserModuleOverride", "UserModuleOverrideFromJar"} {
		// Both original fixture classes: Get returns TRUE, Get2(v1)/Get3 return FALSE.
		link(tlc.NativeClass{Name: name, RequireResource: true, Methods: []tlc.NativeMethod{
			javaNativeBooleanMethod(name, "Get", 0, true, nil),
			javaNativeBooleanMethod(name, "Get2", 1, false, nil),
			javaNativeBooleanMethod(name, "Get3", 0, false, nil),
		}})
	}
	const annotated = "tlc2.overrides.UserModuleOverrideAnnotationImpl"
	annotation := func(identifier, module string) *tlc.NativeOperatorAnnotation {
		return &tlc.NativeOperatorAnnotation{Identifier: identifier, Module: module, Warn: true}
	}
	link(tlc.NativeClass{Name: annotated, RequireResource: true, Methods: []tlc.NativeMethod{
		javaNativeBooleanMethod(annotated, "getNumberOne", 0, true, annotation("Get", "UserModuleOverrideAnnotation")),
		javaNativeBooleanMethod(annotated, "Get2", 0, true, annotation("Get2", "UserModuleOverrideAnnotation")),
		javaNativeBooleanMethod(annotated, "Get2", 1, false, annotation("Get2", "UserModuleOverrideAnnotation")),
		javaNativeBooleanMethod(annotated, "noSuchIdentifier", 0, false, annotation("NoSuchIdentifier", "UserModuleOverrideAnnotation")),
		javaNativeBooleanMethod(annotated, "noSuchModule", 0, false, annotation("Get", "NoSuchModule")),
	}})
	// TLCTestOverrides.get returns these two classes, in this order.
	link(tlc.NativeClass{Name: "tlc2.overrides.TLCTestOverrides", RequireResource: true,
		NewOverrideIndex: func() ([]string, error) { return []string{"tlc2.tool.EvaluatingValueTest", annotated}, nil },
	})
	var actionLock sync.Mutex // original action is synchronized static
	link(tlc.NativeClass{Name: "tlc2.tool.EvaluatingValueTest", Methods: []tlc.NativeMethod{{
		Name: "action", Public: true, Static: true, ParameterCount: 7,
		Signature:  "public static synchronized tlc2.value.impl.Value tlc2.tool.EvaluatingValueTest.action(tlc2.tool.impl.Tool,tla2sany.semantic.ExprOrOpArgNode[],tlc2.util.Context,tlc2.tool.TLCState,tlc2.tool.TLCState,int,tlc2.tool.coverage.CostModel)",
		Evaluation: &tlc.NativeEvaluationAnnotation{Definition: "A", Module: "EvaluatingValueTest", Warn: true, Priority: 100},
		Evaluate: func(tool *tlc.Tool, args []tlc.SemanticNode, c *tlc.Context, s0, s1 *tlc.TLCStateMut, control int, cm tlc.CostModel) (tlc.Value, error) {
			actionLock.Lock()
			defer actionLock.Unlock()
			s1.Bind(tlc.UniqueStringOf("x"), tlc.NewIntValue(42))
			return tlc.BoolTrue, nil
		},
	}}})
}

func javaNativeBooleanMethod(class, name string, arity int, result bool, annotation *tlc.NativeOperatorAnnotation) tlc.NativeMethod {
	parameters := make([]string, arity)
	for i := range parameters {
		parameters[i] = "tlc2.value.impl.Value"
	}
	return tlc.NativeMethod{Name: name, Public: true, Static: true, ParameterCount: arity,
		Signature: "public static tlc2.value.impl.Value " + class + "." + name + "(" + strings.Join(parameters, ",") + ")",
		Operator:  annotation, Eval: func(args []tlc.Value, control int) (tlc.Value, error) {
			if result {
				return tlc.BoolTrue, nil
			}
			return tlc.BoolFalse, nil
		},
	}
}
