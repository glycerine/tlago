package tlc

const (
	ParameterizedSpecActionConstraints = "ACTION_CONSTRAINT"
	ParameterizedSpecConstraints       = "CONSTRAINT"
	ParameterizedSpecPostConditions    = "POST_CONDITIONS"
	ParameterizedSpecInvariant         = "INVARIANT"
	ParameterizedSpecView              = "VIEW"
)

type RuntimeStringConstant struct {
	Name  string
	Value string
}

func (p RuntimeParameters) ExtendeeModules() []string {
	modules := make([]string, 0, len(p.PostConditions)+len(p.Constraints)+len(p.ActionConstraints)+1)
	for _, post := range p.PostConditions {
		modules = appendRuntimeModuleNames(modules, post.Module)
	}
	for _, inv := range p.Invariants {
		modules = appendRuntimeModuleNames(modules, inv.Modules...)
	}
	for _, constraint := range p.Constraints {
		modules = appendRuntimeModuleNames(modules, constraint.Module)
	}
	for _, constraint := range p.ActionConstraints {
		modules = appendRuntimeModuleNames(modules, constraint.Module)
	}
	if p.View != nil {
		modules = appendRuntimeModuleNames(modules, p.View.Module)
	}
	return modules
}

func (p RuntimeParameters) StringConstants() []RuntimeStringConstant {
	constants := make([]RuntimeStringConstant, 0, len(p.Constraints)+len(p.ActionConstraints)+len(p.PostConditions)+1)
	for _, constraint := range p.Constraints {
		constants = appendRuntimeStringConstant(constants, constraint.ConstantName, constraint.FileName)
	}
	for _, constraint := range p.ActionConstraints {
		constants = appendRuntimeStringConstant(constants, constraint.ConstantName, constraint.FileName)
	}
	for _, post := range p.PostConditions {
		constants = appendRuntimeStringConstant(constants, post.ConstantName, post.FileName)
	}
	if p.View != nil {
		constants = appendRuntimeStringConstant(constants, p.View.ConstantName, p.View.FileName)
	}
	return constants
}

func appendRuntimeModuleNames(modules []string, names ...string) []string {
	for _, name := range names {
		if name == "" {
			continue
		}
		seen := false
		for _, existing := range modules {
			if existing == name {
				seen = true
				break
			}
		}
		if !seen {
			modules = append(modules, name)
		}
	}
	return modules
}

func appendRuntimeStringConstant(constants []RuntimeStringConstant, name string, value string) []RuntimeStringConstant {
	if name == "" {
		return constants
	}
	return append(constants, RuntimeStringConstant{Name: name, Value: value})
}
