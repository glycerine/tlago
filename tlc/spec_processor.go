package tlc

import (
	"path/filepath"
	"strings"
)

type Defns struct {
	values *InsMap[string, any]
}

func NewDefns() *Defns {
	return &Defns{values: NewInsMap[string, any]()}
}

func (d *Defns) Put(name string, value any) {
	if d != nil {
		d.values.Set(name, value)
	}
}

func (d *Defns) Get(name string) any {
	if d == nil {
		return nil
	}
	return d.values.Get(name)
}

func (d *Defns) Snapshot() *Defns {
	out := NewDefns()
	if d == nil {
		return out
	}
	for name, value := range d.values.All() {
		out.Put(name, value)
	}
	return out
}

type TLAClass struct {
	Package string
	Modules *InsMap[string, any]
}

func NewTLAClass(pkg string) *TLAClass {
	return &TLAClass{Package: pkg, Modules: NewInsMap[string, any]()}
}

type Spec struct {
	SpecDir         string
	RootFile        string
	ConfigFile      string
	ProcessedDefs   *InsMap[string, struct{}]
	Defns           *Defns
	UnprocessedDefn *Defns
	TLAClass        *TLAClass
	Config          *ModelConfig
	Processor       *SpecProcessor
}

func NewSpec(rootFile string, config *ModelConfig) *Spec {
	if config == nil {
		config = newModelConfig("", false)
	}
	specDir := filepath.Dir(rootFile)
	if specDir == "." {
		specDir = ""
	}
	spec := &Spec{
		SpecDir:         specDir,
		RootFile:        rootFile,
		ConfigFile:      config.configFileName,
		ProcessedDefs:   NewInsMap[string, struct{}](),
		Defns:           NewDefns(),
		UnprocessedDefn: NewDefns(),
		TLAClass:        NewTLAClass("tlc2.module"),
		Config:          config,
	}
	spec.Processor = NewSpecProcessor(spec.RootName(), spec.Defns, config)
	spec.UnprocessedDefn = spec.Processor.UnprocessedDefns
	return spec
}

func (s *Spec) RootName() string {
	if s == nil || s.RootFile == "" {
		return "Spec"
	}
	base := filepath.Base(s.RootFile)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	if base == "" {
		return "Spec"
	}
	return base
}

func (s *Spec) GetModelConfig() *ModelConfig {
	if s == nil {
		return nil
	}
	return s.Config
}

func (s *Spec) GetSpecProcessor() *SpecProcessor {
	if s == nil {
		return nil
	}
	return s.Processor
}

func (s *Spec) ApplyToTool(tool *Tool) *Tool {
	if tool == nil {
		tool = NewTool()
	}
	if s == nil || s.Processor == nil {
		return tool
	}
	s.Processor.ApplyToTool(tool)
	tool.RootName = s.RootName()
	tool.RootFile = s.RootFile
	tool.ConfigFile = s.ConfigFile
	tool.SpecDir = s.SpecDir
	tool.ModelConfig = s.Config
	return tool
}

type SpecProcessor struct {
	RootFile string
	ToolID   int64
	Defns    *Defns
	Config   *ModelConfig

	Variables         []*UniqueString
	ProcessedDefs     *InsMap[string, struct{}]
	UnprocessedDefns  *Defns
	ConstantDefns     *InsMap[string, Value]
	InitPred          []*Action
	NextPred          *Action
	Temporals         []*Action
	TemporalNames     []string
	ImpliedTemporals  []*Action
	ImpliedTempNames  []string
	Invariants        []*Action
	InvariantNames    []string
	ImpliedInits      []*Action
	ImpliedInitNames  []string
	ImpliedActions    []*Action
	ImpliedActNames   []string
	ModelConstraints  []SemanticNode
	ActionConstraints []SemanticNode
	PossiblePostConds []*Action
	Assumptions       []SemanticNode
	AssumptionIsAxiom []bool
	RLReward          SemanticNode
	Periodic          SemanticNode
	ViewSpec          SemanticNode
	SymmetrySpec      string
	AliasSpec         string
	SpecificationName string
}

func NewSpecProcessor(rootFile string, defns *Defns, config *ModelConfig) *SpecProcessor {
	if defns == nil {
		defns = NewDefns()
	}
	if config == nil {
		config = newModelConfig("", false)
	}
	p := &SpecProcessor{
		RootFile:         rootFile,
		ToolID:           nextToolID.Add(1),
		Defns:            defns,
		Config:           config,
		ProcessedDefs:    NewInsMap[string, struct{}](),
		UnprocessedDefns: NewDefns(),
		ConstantDefns:    NewInsMap[string, Value](),
	}
	p.ProcessConfig()
	return p
}

func (p *SpecProcessor) SetVariables(names []string) {
	p.Variables = make([]*UniqueString, len(names))
	for i, name := range names {
		p.Variables[i] = UniqueStringOf(name)
	}
}

func (p *SpecProcessor) ProcessConfig() {
	if p == nil || p.Config == nil {
		return
	}
	p.SpecificationName = p.Config.GetSpec()
	p.SymmetrySpec = p.Config.GetSymmetry()
	p.AliasSpec = p.Config.GetAlias()
	p.ViewSpec = p.Config.GetView()
	p.RLReward = p.Config.GetRLReward()
	p.Periodic = p.Config.GetPeriodic()

	p.InitPred = nil
	if initName := p.Config.GetInit(); initName != "" {
		p.InitPred = append(p.InitPred, &Action{Name: initName, IsInitPred: true})
	}
	if nextName := p.Config.GetNext(); nextName != "" {
		p.NextPred = &Action{Name: nextName}
	}
	p.Invariants, p.InvariantNames = actionsFromNames(p.Config.GetInvariants(), false)
	p.Temporals, p.TemporalNames = actionsFromNames(p.Config.GetProperties(), false)
	p.PossiblePostConds, _ = actionsFromNames(p.Config.GetPostConditions(), false)
	p.ModelConstraints = semanticNodesFromNames(p.Config.GetConstraints())
	p.ActionConstraints = semanticNodesFromNames(p.Config.GetActionConstraints())
}

func (p *SpecProcessor) ApplyToTool(tool *Tool) {
	if p == nil || tool == nil {
		return
	}
	names := make([]string, len(p.Variables))
	for i, variable := range p.Variables {
		names[i] = variable.String()
	}
	if len(names) != 0 {
		SetStateVariables(names)
	}
	tool.ModelConfig = p.Config
	tool.InitStateSpec = append([]*Action(nil), p.InitPred...)
	tool.NextStateSpec = p.NextPred
	tool.Temporals = append([]*Action(nil), p.Temporals...)
	tool.TemporalNames = append([]string(nil), p.TemporalNames...)
	tool.ImpliedTemporals = append([]*Action(nil), p.ImpliedTemporals...)
	tool.ImpliedTempNames = append([]string(nil), p.ImpliedTempNames...)
	tool.Invariants = append([]*Action(nil), p.Invariants...)
	tool.InvariantNames = append([]string(nil), p.InvariantNames...)
	tool.ImpliedInits = append([]*Action(nil), p.ImpliedInits...)
	tool.ImpliedInitNames = append([]string(nil), p.ImpliedInitNames...)
	tool.ImpliedActions = append([]*Action(nil), p.ImpliedActions...)
	tool.ImpliedActNames = append([]string(nil), p.ImpliedActNames...)
	tool.ModelConstraints = append([]SemanticNode(nil), p.ModelConstraints...)
	tool.ActionConstraints = append([]SemanticNode(nil), p.ActionConstraints...)
	tool.PostConditionSpecs = append([]*Action(nil), p.PossiblePostConds...)
	tool.ViewSpec = p.ViewSpec
	tool.AssignActionIDs()
}

func actionsFromNames(names []string, init bool) ([]*Action, []string) {
	actions := make([]*Action, 0, len(names))
	labels := make([]string, 0, len(names))
	for _, name := range names {
		action := &Action{Name: name, IsInitPred: init}
		actions = append(actions, action)
		labels = append(labels, name)
	}
	return actions, labels
}

func semanticNodesFromNames(names []string) []SemanticNode {
	nodes := make([]SemanticNode, len(names))
	for i, name := range names {
		nodes[i] = name
	}
	return nodes
}
