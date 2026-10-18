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

// NewRuntimeInvariantTemplate exposes
// ParameterizedSpecObj.RuntimeInvariantTemplate(Set<String>, String). The
// supplied module slice is retained, matching the source Set ownership.
func NewRuntimeInvariantTemplate(modules []string, expression string) RuntimeInvariantTemplate {
	return RuntimeInvariantTemplate{
		Modules: modules, Expression: expression,
		ownerCaptured: true, modulesOwner: modules, expressionOwner: expression,
	}
}

func NewRuntimePostCondition(module string, operator string) RuntimePostCondition {
	return NewRuntimePostConditionWithConstants(module, operator, make(map[string]string))
}

func NewRuntimePostConditionWithConstant(module string, operator string, name string, value string) RuntimePostCondition {
	constants := map[string]string{name: value}
	return newRuntimePostCondition(module, operator, name, value, constants, true)
}

func NewRuntimePostConditionWithConstants(module string, operator string, constants map[string]string) RuntimePostCondition {
	return newRuntimePostCondition(module, operator, "", "", constants, true)
}

func NewRuntimeView(module string, operator string) *RuntimeView {
	return NewRuntimeViewWithConstants(module, operator, make(map[string]string))
}

func NewRuntimeViewWithConstant(module string, operator string, name string, value string) *RuntimeView {
	constants := map[string]string{name: value}
	return newRuntimeView(module, operator, name, value, constants, true)
}

func NewRuntimeViewWithConstants(module string, operator string, constants map[string]string) *RuntimeView {
	return newRuntimeView(module, operator, "", "", constants, true)
}

func NewRuntimeConstraintWithConstant(module string, operator string, name string, value string) RuntimeConstraint {
	constants := map[string]string{name: value}
	return newRuntimeConstraint(module, operator, name, value, constants, true)
}

func NewRuntimeConstraintWithConstants(module string, operator string, constants map[string]string) RuntimeConstraint {
	return newRuntimeConstraint(module, operator, "", "", constants, true)
}

func newRuntimeParameterizedCarrier(module string, operator string, constantName string, fileName string, constants map[string]string, constantsSet bool) runtimeParameterizedCarrier {
	return runtimeParameterizedCarrier{module: module, operator: operator, constantName: constantName, fileName: fileName, constants: constants, constantsSet: constantsSet}
}

func newRuntimePostCondition(module string, operator string, constantName string, fileName string, constants map[string]string, constantsSet bool) RuntimePostCondition {
	owner := newRuntimeParameterizedCarrier(module, operator, constantName, fileName, constants, constantsSet)
	return RuntimePostCondition{Module: module, Operator: operator, ConstantName: constantName, FileName: fileName, Constants: constants, constantsSet: constantsSet, ownerCaptured: true, owner: owner}
}

func newRuntimeView(module string, operator string, constantName string, fileName string, constants map[string]string, constantsSet bool) *RuntimeView {
	owner := newRuntimeParameterizedCarrier(module, operator, constantName, fileName, constants, constantsSet)
	return &RuntimeView{Module: module, Operator: operator, ConstantName: constantName, FileName: fileName, Constants: constants, constantsSet: constantsSet, ownerCaptured: true, owner: owner}
}

func newRuntimeConstraint(module string, operator string, constantName string, fileName string, constants map[string]string, constantsSet bool) RuntimeConstraint {
	owner := newRuntimeParameterizedCarrier(module, operator, constantName, fileName, constants, constantsSet)
	return RuntimeConstraint{Module: module, Operator: operator, ConstantName: constantName, FileName: fileName, Constants: constants, constantsSet: constantsSet, ownerCaptured: true, owner: owner}
}

func (c RuntimeConstraint) ownedCarrier() runtimeParameterizedCarrier {
	if c.ownerCaptured {
		return c.owner
	}
	return newRuntimeParameterizedCarrier(c.Module, c.Operator, c.ConstantName, c.FileName, c.Constants, c.constantsSet)
}

func (c *RuntimeConstraint) captureOwner() {
	if !c.ownerCaptured {
		c.owner = newRuntimeParameterizedCarrier(c.Module, c.Operator, c.ConstantName, c.FileName, c.Constants, c.constantsSet)
		c.ownerCaptured = true
	}
}

func (p RuntimePostCondition) ownedCarrier() runtimeParameterizedCarrier {
	if p.ownerCaptured {
		return p.owner
	}
	return newRuntimeParameterizedCarrier(p.Module, p.Operator, p.ConstantName, p.FileName, p.Constants, p.constantsSet)
}

func (p *RuntimePostCondition) captureOwner() {
	if !p.ownerCaptured {
		p.owner = newRuntimeParameterizedCarrier(p.Module, p.Operator, p.ConstantName, p.FileName, p.Constants, p.constantsSet)
		p.ownerCaptured = true
	}
}

func (v *RuntimeView) ownedCarrier() runtimeParameterizedCarrier {
	if v.ownerCaptured {
		return v.owner
	}
	return newRuntimeParameterizedCarrier(v.Module, v.Operator, v.ConstantName, v.FileName, v.Constants, v.constantsSet)
}

func (v *RuntimeView) captureOwner() {
	if v != nil && !v.ownerCaptured {
		v.owner = newRuntimeParameterizedCarrier(v.Module, v.Operator, v.ConstantName, v.FileName, v.Constants, v.constantsSet)
		v.ownerCaptured = true
	}
}

func (i *RuntimeInvariantTemplate) ownedModules() []string {
	if i.ownerCaptured {
		return i.modulesOwner
	}
	return i.Modules
}

func (i *RuntimeInvariantTemplate) ownedExpression() string {
	if i.ownerCaptured {
		return i.expressionOwner
	}
	return i.Expression
}

func (i *RuntimeInvariantTemplate) captureOwner() {
	if !i.ownerCaptured {
		i.modulesOwner = i.Modules
		i.expressionOwner = i.Expression
		i.ownerCaptured = true
	}
}

func (p RuntimeParameters) captureOwners() RuntimeParameters {
	for i := range p.Invariants {
		p.Invariants[i].captureOwner()
	}
	for i := range p.Constraints {
		p.Constraints[i].captureOwner()
	}
	for i := range p.ActionConstraints {
		p.ActionConstraints[i].captureOwner()
	}
	for i := range p.PostConditions {
		p.PostConditions[i].captureOwner()
	}
	p.View.captureOwner()
	return p
}

func (i *RuntimeInvariantTemplate) require() *RuntimeInvariantTemplate {
	if i == nil {
		panic(NewNullPointerException())
	}
	return i
}

// GetModules is ParameterizedSpecObj.InvariantTemplate.getModules. The returned
// slice shares its backing storage with the template, like Java's returned Set.
func (i *RuntimeInvariantTemplate) GetModules() []string {
	return i.require().ownedModules()
}

// GetAction is ParameterizedSpecObj.RuntimeInvariantTemplate.getAction.
func (i *RuntimeInvariantTemplate) GetAction(spec *SpecProcessor) (*Action, error) {
	i.require()
	spec.require()
	definition, err := spec.RuntimeInvariantCompiler(i.ownedExpression())
	if err != nil {
		return nil, err
	}
	if definition == nil {
		panic(NewNullPointerException())
	}
	return NewActionFromOpDef(definition.GetBody(), EmptyContext, definition, false, true), nil
}

func (p RuntimeParameters) ExtendeeModules() []string {
	modules := make([]string, 0, len(p.PostConditions)+len(p.Constraints)+len(p.ActionConstraints)+1)
	for _, post := range p.PostConditions {
		modules = appendRuntimeModuleNames(modules, post.ownedCarrier().module)
	}
	for _, inv := range p.Invariants {
		modules = appendRuntimeModuleNames(modules, inv.GetModules()...)
	}
	for _, constraint := range p.Constraints {
		modules = appendRuntimeModuleNames(modules, constraint.ownedCarrier().module)
	}
	for _, constraint := range p.ActionConstraints {
		modules = appendRuntimeModuleNames(modules, constraint.ownedCarrier().module)
	}
	if p.View != nil {
		modules = appendRuntimeModuleNames(modules, p.View.ownedCarrier().module)
	}
	return modules
}

func (p RuntimeParameters) StringConstants() []RuntimeStringConstant {
	constants := make([]RuntimeStringConstant, 0, len(p.Constraints)+len(p.ActionConstraints)+len(p.PostConditions)+1)
	for _, constraint := range p.Constraints {
		carrier := constraint.ownedCarrier()
		constants = appendRuntimeStringConstants(constants, carrier.constantName, carrier.fileName, carrier.constants, carrier.constantsSet)
	}
	for _, constraint := range p.ActionConstraints {
		carrier := constraint.ownedCarrier()
		constants = appendRuntimeStringConstants(constants, carrier.constantName, carrier.fileName, carrier.constants, carrier.constantsSet)
	}
	for _, post := range p.PostConditions {
		carrier := post.ownedCarrier()
		constants = appendRuntimeStringConstants(constants, carrier.constantName, carrier.fileName, carrier.constants, carrier.constantsSet)
	}
	if p.View != nil {
		carrier := p.View.ownedCarrier()
		constants = appendRuntimeStringConstants(constants, carrier.constantName, carrier.fileName, carrier.constants, carrier.constantsSet)
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

func appendRuntimeStringConstants(constants []RuntimeStringConstant, legacyName string, legacyValue string, definitions map[string]string, definitionsProvided bool) []RuntimeStringConstant {
	constants = appendRuntimeStringConstant(constants, legacyName, legacyValue)
	if definitionsProvided && definitions == nil {
		panic(NewNullPointerException())
	}
	for name, value := range definitions {
		if name == legacyName {
			continue
		}
		constants = appendRuntimeStringConstant(constants, name, value)
	}
	return constants
}
