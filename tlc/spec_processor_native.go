package tlc

import (
	"os"
	"strings"
)

// ProcessModuleOverrides ports SpecProcessor.processModuleOverrides for linked
// native classes. Conventional module classes are processed before the ordered
// ITLCOverrides indexes, just as on the Java side.
func (p *SpecProcessor) ProcessModuleOverrides(tool *Tool, loader *NativeClassLoader) bool {
	p.require()
	return p.ProcessModuleOverridesForModule(p.ModuleTbl.GetRootModule(), p.ModuleTbl, tool, loader)
}

// ProcessModuleOverridesForModule ports the public runtime-module overload of
// SpecProcessor.processModuleOverrides. The supplied module owns the root
// definitions used for conventional override arity checks, while moduleTbl
// supplies the complete module set used by annotated overrides.
func (p *SpecProcessor) ProcessModuleOverridesForModule(module *ModuleNode, moduleTbl *ExternalModuleTable, tool *Tool, loader *NativeClassLoader) bool {
	p.require()
	modules := moduleTbl.GetModuleNodes()
	moduleSet := make(map[string]*ModuleNode, len(modules))
	for _, candidate := range modules {
		moduleSet[candidate.GetName().String()] = candidate
	}
	return p.ProcessModuleOverridesWithDefinitions(modules, module.GetOpDefs(), moduleSet, tool, loader)
}

// ProcessModuleOverridesWithDefinitions ports the public array/map overload of
// SpecProcessor.processModuleOverrides. Java's final Hashtable parameter is
// unused by the source body and therefore has no native counterpart.
func (p *SpecProcessor) ProcessModuleOverridesWithDefinitions(modules []*ModuleNode, rootDefs []*OpDefNode, moduleSet map[string]*ModuleNode, tool *Tool, loader *NativeClassLoader) bool {
	p.require()
	if modules == nil {
		panic(NewNullPointerException())
	}
	for _, module := range modules {
		moduleName := module.GetName()
		class := loader.LoadClass(moduleName.String())
		if class == nil {
			continue
		}
		arities := make(map[*UniqueString]int)
		if !class.BuiltIn {
			if rootDefs == nil {
				panic(NewNullPointerException())
			}
			for _, def := range rootDefs {
				origin := def.GetOriginallyDefinedInModuleNode()
				if origin == nil {
					panic(NewNullPointerException())
				}
				originName := origin.GetName()
				if originName == nil {
					panic(NewNullPointerException())
				}
				if originName == moduleName {
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
			if value == nil {
				panic(NewNullPointerException())
			}
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
		if index == nil || (index.NewOverrideIndex == nil && index.NewOverrideIndexNullable == nil) {
			continue
		}
		names, err := nativeOverrideIndexNames(index)
		if err != nil {
			switch err.(type) {
			case *InstantiationException, *IllegalAccessException:
				panic(NewTLCRuntimeException(ECGeneral))
			}
			panic(err)
		}
		if names == nil {
			panic(NewNullPointerException())
		}
		for _, className := range names {
			if className == nil {
				panic(NewNullPointerException())
			}
			class := loader.LoadClass(*className)
			if class == nil {
				panic(NewNoClassDefFoundError(strings.ReplaceAll(*className, ".", "/")))
			}
			for _, method := range class.Methods {
				display := "<Java Method: " + method.Signature + ">"
				if annotation := method.Evaluation; annotation != nil {
					module := nativeOverrideModule(moduleSet, annotation.Module)
					name := annotation.Module + "!" + annotation.Definition
					if module == nil {
						if annotation.Warn {
							PrintMessage(ECTLCModuleValueJavaMethodOverrideModuleMismatch, name, class.Resource, display)
						}
						continue
					}
					def := module.GetOpDefByName(annotation.Definition)
					if def == nil {
						if annotation.Warn {
							PrintMessage(ECTLCModuleValueJavaMethodOverrideIdentifierMismatch, name, class.Resource, display)
						}
						continue
					}
					method.requirePublicAccess()
					value := NewEvaluatingValue(method.Signature, annotation.MinLevel, annotation.Priority, def, method.Evaluate)
					value.Instance = !method.Static
					value.DeclaringClass = nativeMethodDeclaringClass(method.Signature)
					value.captureMetadata()
					var native Value = value
					if body, ok := def.GetBody().(interface{ GetToolObjectAt(int32) any }); ok {
						switch previous := body.GetToolObjectAt(p.ownedToolID()).(type) {
						case *EvaluatingValue:
							if previous != nil {
								native = NewPriorityEvaluatingValue(value, previous)
							}
						case *PriorityEvaluatingValue:
							if previous != nil {
								previous.Add(value)
								native = previous
							}
						}
					}
					p.installNativeDefinition(tool, def, native)
					if !annotation.Silent {
						PrintMessage(ECTLCModuleValueJavaMethodOverrideLoaded, name, class.Resource, display)
					}
					continue
				}
				if annotation := method.Callable; annotation != nil {
					module := nativeOverrideModule(moduleSet, annotation.Module)
					name := annotation.Module + "!" + annotation.Definition
					if module == nil {
						if annotation.Warn {
							PrintMessage(ECTLCModuleValueJavaMethodOverrideModuleMismatch, name, class.Resource, display)
						}
						continue
					}
					def := module.GetOpDefByName(annotation.Definition)
					if def == nil {
						if annotation.Warn {
							PrintMessage(ECTLCModuleValueJavaMethodOverrideIdentifierMismatch, name, class.Resource, display)
						}
						continue
					}
					if !method.Public {
						panic(NewTLCRuntimeException(ECGeneral))
					}
					value := NewCallableValue(method.Signature, annotation.MinLevel, def, method.Call)
					value.Instance = !method.Static
					value.DeclaringClass = nativeMethodDeclaringClass(method.Signature)
					value.captureMetadata()
					p.installNativeDefinition(tool, def, value)
					hasCallableValue = true
					PrintMessage(ECTLCModuleValueJavaMethodOverrideLoaded, name, class.Resource, value.String())
					continue
				}
				if annotation := method.Operator; annotation != nil {
					module := nativeOverrideModule(moduleSet, annotation.Module)
					if module == nil {
						if annotation.Warn {
							PrintWarning(ECTLCModuleValueJavaMethodOverrideModuleMismatch, annotation.Identifier, annotation.Module, method.Signature)
						}
						continue
					}
					def := module.GetOpDefByName(annotation.Identifier)
					if def == nil {
						if annotation.Warn {
							PrintWarning(ECTLCModuleValueJavaMethodOverrideIdentifierMismatch, annotation.Identifier, annotation.Module, method.Signature)
						}
						continue
					}
					value := method.methodValue(annotation.MinLevel)
					if def.Arity() != method.ParameterCount {
						if annotation.Warn {
							if value == nil {
								panic(NewNullPointerException())
							}
							PrintWarning(ECTLCModuleValueJavaMethodOverrideMismatch, def.Name.String(), class.Name, value.String())
						}
						continue
					}
					if value == nil {
						panic(NewNullPointerException())
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

func nativeOverrideModule(moduleSet map[string]*ModuleNode, name string) *ModuleNode {
	if moduleSet == nil {
		panic(NewNullPointerException())
	}
	return moduleSet[name]
}

func nativeOverrideIndexNames(index *NativeClass) ([]*string, error) {
	if index.NewOverrideIndexNullable != nil {
		return index.NewOverrideIndexNullable()
	}
	names, err := index.NewOverrideIndex()
	if names == nil || err != nil {
		return nil, err
	}
	result := make([]*string, len(names))
	for i := range names {
		result[i] = javaString(names[i])
	}
	return result, nil
}

func (p *SpecProcessor) installNativeDefinition(tool *Tool, def *OpDefNode, value Value) {
	SetSemanticToolObjectForToolID(p.ownedToolID(), def.GetBody(), value)
	p.ownedDefns().Put(def.Name, value)
	tool.Define(def.Symbol, value)
	tool.DefnsByName[def.Name] = value
}
