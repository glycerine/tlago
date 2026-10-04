// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.

package tlc

import (
	"os"
	"strings"
)

// ProcessModuleOverrides ports SpecProcessor.processModuleOverrides for linked
// native classes. Conventional module classes are processed before the ordered
// ITLCOverrides indexes, just as on the Java side.
func (p *SpecProcessor) ProcessModuleOverrides(tool *Tool, loader *NativeClassLoader) bool {
	modules := p.ModuleTbl.GetModuleNodes()
	rootDefs := p.ModuleTbl.GetRootModule().GetOpDefs()
	for _, module := range modules {
		class := loader.LoadClass(module.Name.String())
		if class == nil {
			continue
		}
		arities := make(map[*UniqueString]int)
		if !class.BuiltIn {
			for _, def := range rootDefs {
				if origin := def.GetOriginallyDefinedInModuleNode(); origin != nil && origin.Name == module.Name {
					arities[def.Name] = def.Arity()
				}
			}
		}
		javaDefs := make(map[*UniqueString]Value)
		for _, method := range class.Methods {
			if !method.Public || !method.Static {
				continue
			}
			name := UniqueStringOf(TLARegistryMapName(method.Name))
			if method.Operator != nil || method.Evaluation != nil {
				continue
			}
			value := method.methodValue(0)
			if !class.BuiltIn {
				arity, found := arities[name]
				if !found || arity != method.ParameterCount {
					PrintWarning(ECTLCModuleValueJavaMethodOverrideMismatch, name.String(), class.Resource, value.String())
					continue
				}
				PrintMessage(ECTLCModuleValueJavaMethodOverrideLoaded, name.String(), class.Resource, value.String())
			}
			javaDefs[name] = value
		}
		for _, def := range module.GetOpDefs() {
			if value := javaDefs[def.Name]; value != nil {
				p.installNativeDefinition(tool, def, value)
			}
		}
	}
	hasCallableValue := false
	indexes := "tlc2.module.TLCBuiltInOverrides" + string(os.PathListSeparator) + tlcGetSystemProperty("tlc2.overrides.TLCOverrides", "tlc2.overrides.TLCOverrides")
	for _, indexName := range filenameSplitPaths(indexes) {
		index := loader.LoadClass(indexName)
		if index == nil || index.NewOverrideIndex == nil {
			continue
		}
		names, err := index.NewOverrideIndex()
		if err != nil {
			switch err.(type) {
			case *InstantiationException, *IllegalAccessException:
				panic(NewTLCRuntimeException(ECGeneral))
			}
			panic(err)
		}
		for _, className := range names {
			class := loader.LoadClass(className)
			if class == nil {
				panic(NewNoClassDefFoundError(strings.ReplaceAll(className, ".", "/")))
			}
			for _, method := range class.Methods {
				display := "<Java Method: " + method.Signature + ">"
				if annotation := method.Evaluation; annotation != nil {
					module := p.ModuleTbl.GetModuleNode(UniqueStringOf(annotation.Module))
					name := annotation.Module + "!" + annotation.Definition
					if module == nil {
						if annotation.Warn {
							PrintMessage(ECTLCModuleValueJavaMethodOverrideModuleMismatch, name, class.Resource, display)
						}
						continue
					}
					def := module.GetOpDef(UniqueStringOf(annotation.Definition))
					if def == nil {
						if annotation.Warn {
							PrintMessage(ECTLCModuleValueJavaMethodOverrideIdentifierMismatch, name, class.Resource, display)
						}
						continue
					}
					value := NewEvaluatingValue(method.Signature, annotation.MinLevel, annotation.Priority, def, method.Evaluate)
					var native Value = value
					if body, ok := def.Body.(interface{ GetToolObject() any }); ok {
						switch previous := body.GetToolObject().(type) {
						case *EvaluatingValue:
							native = NewPriorityEvaluatingValue(value, previous)
						case *PriorityEvaluatingValue:
							previous.Add(value)
							native = previous
						}
					}
					p.installNativeDefinition(tool, def, native)
					if !annotation.Silent {
						PrintMessage(ECTLCModuleValueJavaMethodOverrideLoaded, name, class.Resource, display)
					}
					continue
				}
				if annotation := method.Callable; annotation != nil {
					module := p.ModuleTbl.GetModuleNode(UniqueStringOf(annotation.Module))
					name := annotation.Module + "!" + annotation.Definition
					if module == nil {
						if annotation.Warn {
							PrintMessage(ECTLCModuleValueJavaMethodOverrideModuleMismatch, name, class.Resource, display)
						}
						continue
					}
					def := module.GetOpDef(UniqueStringOf(annotation.Definition))
					if def == nil {
						if annotation.Warn {
							PrintMessage(ECTLCModuleValueJavaMethodOverrideIdentifierMismatch, name, class.Resource, display)
						}
						continue
					}
					value := NewCallableValue(method.Signature, annotation.MinLevel, def, method.Call)
					p.installNativeDefinition(tool, def, value)
					hasCallableValue = true
					PrintMessage(ECTLCModuleValueJavaMethodOverrideLoaded, name, class.Resource, value.String())
					continue
				}
				if annotation := method.Operator; annotation != nil {
					module := p.ModuleTbl.GetModuleNode(UniqueStringOf(annotation.Module))
					if module == nil {
						if annotation.Warn {
							PrintWarning(ECTLCModuleValueJavaMethodOverrideModuleMismatch, annotation.Identifier, annotation.Module, method.Signature)
						}
						continue
					}
					def := module.GetOpDef(UniqueStringOf(annotation.Identifier))
					if def == nil {
						if annotation.Warn {
							PrintWarning(ECTLCModuleValueJavaMethodOverrideIdentifierMismatch, annotation.Identifier, annotation.Module, method.Signature)
						}
						continue
					}
					value := method.methodValue(annotation.MinLevel)
					if def.Arity() != method.ParameterCount {
						if annotation.Warn {
							PrintWarning(ECTLCModuleValueJavaMethodOverrideMismatch, def.Name.String(), class.Name, value.String())
						}
						continue
					}
					if annotation.Warn {
						PrintMessage(ECTLCModuleValueJavaMethodOverrideLoaded, def.Name.String(), class.Name, nativeMethodDisplay(value))
					}
					p.installNativeDefinition(tool, def, value)
				}
			}
		}
	}
	return hasCallableValue
}

func (p *SpecProcessor) installNativeDefinition(tool *Tool, def *OpDefNode, value Value) {
	if body, ok := def.Body.(interface{ SetToolObject(any) }); ok {
		body.SetToolObject(value)
	}
	p.Defns.Put(def.Name, value)
	tool.Define(def.Symbol, value)
	tool.DefnsByName[def.Name] = value
}
