package tlago

import (
	"strconv"
	"sync"

	"github.com/glycerine/tlago/tlc"
)

func init() {
	tlc.SetSemanticBuiltInResolver(SanyContextIsBuiltIn)
}

type sanyBuiltinOperator struct {
	name         string
	arity        int
	level        tlaLevel
	argMaxLevels []tlaLevel
	argWeights   []int
}

// SanyBuiltinOperator adapts BuiltInOperators.BuiltInOperator. Its metadata
// arrays are retained exactly as the source constructor retains them.
type SanyBuiltinOperator struct {
	Name         *tlc.UniqueString
	Arity        int
	OpLevel      int
	ArgMaxLevels []int
	ArgWeights   []int
}

// NewSanyBuiltinOperator constructs a property and retains both metadata arrays.
func NewSanyBuiltinOperator(name string, arity, opLevel int, argMaxLevels, argWeights []int) *SanyBuiltinOperator {
	return &SanyBuiltinOperator{
		Name: tlc.UniqueStringOf(name), Arity: arity, OpLevel: opLevel,
		ArgMaxLevels: argMaxLevels, ArgWeights: argWeights,
	}
}

func (info *SanyBuiltinOperator) internal() sanyBuiltinOperator {
	if info == nil || info.Name == nil {
		panic(tlc.NewNullPointerException())
	}
	return sanyBuiltinOperator{
		name: info.Name.String(), arity: info.Arity, level: info.OpLevel,
		argMaxLevels: info.ArgMaxLevels, argWeights: info.ArgWeights,
	}
}

// SanyBuiltinOperatorProperties exposes BuiltInOperators.Properties. The slice,
// its entries and their metadata arrays are live inputs to context rebuilding.
var SanyBuiltinOperatorProperties = func() []*SanyBuiltinOperator {
	result := make([]*SanyBuiltinOperator, len(sanyBuiltinOperators))
	for i, info := range sanyBuiltinOperators {
		result[i] = NewSanyBuiltinOperator(info.name, info.arity, info.level, info.argMaxLevels, info.argWeights)
	}
	return result
}()

func newSanyInitialContext() *sanyContext {
	ctx := newSanyContext()
	for _, property := range SanyBuiltinOperatorProperties {
		ctx.addSymbol(newSanyBuiltInSymbol(property.internal()))
	}
	return ctx
}

func newSanyBuiltInSymbol(info sanyBuiltinOperator) *sanySemOpDefNode {
	node := &sanySemOpDefNode{
		letInLevel: -1, recursiveSection: -1,
		sanySemSymbolBase: sanySemSymbolBase{
			sanySemanticNode: newSanySemanticNode(sanyBuiltInKind),
			name:             info.name, arity: info.arity,
			originalModuleName: "--TLA+ BUILTINS--",
			pos:                Position{File: "--TLA+ BUILTINS--"},
		},
		defined: true,
	}
	// OpDefNode(BuiltInOperator) creates syntax and phony formals before
	// initializing levels. Operators without both metadata arrays remain unchecked.
	position := Position{File: "--TLA+ BUILTINS--"}
	node.TreeNode = &SanySyntaxNode{Image: info.name, FileName: position.File, Range: SanyRange{Begin: position, End: position},
		Zero: []*SanySyntaxNode{}, One: []*SanySyntaxNode{}, Heirs: []*SanySyntaxNode{}, PreComments: []string{}, ProofLevel: -1, Level: -1}
	node.Location = tlc.NewSourceLocation(position.File, 0, 0, 0, 0)
	if info.arity >= 0 {
		node.formalNodes = make([]*sanyFormalParamNode, info.arity)
		for i := range node.formalNodes {
			node.formalNodes[i] = newSanyFormalParamNode("Formal_"+strconv.Itoa(i), 0, position, nil, nil)
		}
	}
	if info.argMaxLevels != nil && info.argWeights != nil {
		if info.arity == -1 {
			node.argMaxLevels = make([]tlaLevel, 0)
			node.argWeights = make([]int, 0)
			if len(info.argMaxLevels) > 0 {
				node.argMaxLevels = []tlaLevel{info.argMaxLevels[0]}
				node.argWeights = []int{info.argWeights[0]}
			}
		} else {
			node.argMaxLevels = info.argMaxLevels
			node.argWeights = info.argWeights
		}
		node.isLeibniz = true
		node.leibniz = make([]bool, len(info.argWeights))
		for i, weight := range info.argWeights {
			node.leibniz[i] = weight > 0
			node.isLeibniz = node.isLeibniz && node.leibniz[i]
		}
		*node.level = info.level
		*node.levelChecked = 99
	}
	return node
}

// Resolve a source initial-context symbol without XML's variadic Cartesian
// product translation: the initial-context \times symbol has arity two.
func sanyInitialBuiltinOperatorInfo(name string) (sanyBuiltinOperator, bool) {
	name = ResolveSanyOperatorSynonymString(name)
	for _, property := range SanyBuiltinOperatorProperties {
		info := property.internal()
		if info.name == name {
			return info, true
		}
	}
	return sanyBuiltinOperator{}, false
}

func sanyBuiltinOperatorInfo(name string) (sanyBuiltinOperator, bool) {
	name = sanyXMLBuiltinName(name)
	for _, property := range SanyBuiltinOperatorProperties {
		info := property.internal()
		if info.name == name {
			return info, true
		}
	}
	return sanyBuiltinOperator{}, false
}

func sanyBuiltinLeibniz(info sanyBuiltinOperator) []bool {
	if info.arity <= 0 {
		return nil
	}
	out := make([]bool, info.arity)
	for i := range out {
		weight := 1
		if i < len(info.argWeights) {
			weight = info.argWeights[i]
		}
		out[i] = weight == 1
	}
	return out
}

func sanyMaxLevels(vals ...tlaLevel) []tlaLevel {
	return append([]tlaLevel{}, vals...)
}

func sanyWeights(vals ...int) []int {
	return append([]int{}, vals...)
}

var sanyBuiltinOperators = []sanyBuiltinOperator{
	{name: "STRING", arity: 0, level: constantLevel, argMaxLevels: sanyMaxLevels(), argWeights: sanyWeights()},
	{name: "FALSE", arity: 0, level: constantLevel, argMaxLevels: sanyMaxLevels(), argWeights: sanyWeights()},
	{name: "TRUE", arity: 0, level: constantLevel, argMaxLevels: sanyMaxLevels(), argWeights: sanyWeights()},
	{name: "BOOLEAN", arity: 0, level: constantLevel, argMaxLevels: sanyMaxLevels(), argWeights: sanyWeights()},
	{name: "=", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: "/=", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: ".", arity: 2, level: constantLevel},
	{name: "'", arity: 1, level: actionLevel, argMaxLevels: sanyMaxLevels(variableLevel), argWeights: sanyWeights(0)},
	{name: "\\lnot", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "\\neg", arity: 1, level: constantLevel},
	{name: "\\land", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel, temporalLevel), argWeights: sanyWeights(1, 1)},
	{name: "\\lor", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel, temporalLevel), argWeights: sanyWeights(1, 1)},
	{name: "\\equiv", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel, temporalLevel), argWeights: sanyWeights(1, 1)},
	{name: "=>", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel, temporalLevel), argWeights: sanyWeights(1, 1)},
	{name: "SUBSET", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "UNION", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "DOMAIN", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "\\subseteq", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: "\\in", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: "\\notin", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: "\\", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: "\\intersect", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: "\\union", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: "\\times", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: "~>", arity: 2, level: temporalLevel, argMaxLevels: sanyMaxLevels(temporalLevel, temporalLevel), argWeights: sanyWeights(0, 0)},
	{name: "[]", arity: 1, level: temporalLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(0)},
	{name: "<>", arity: 1, level: temporalLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(0)},
	{name: "ENABLED", arity: 1, level: variableLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(0)},
	{name: "UNCHANGED", arity: 1, level: actionLevel, argMaxLevels: sanyMaxLevels(variableLevel), argWeights: sanyWeights(0)},
	{name: "\\cdot", arity: 2, level: actionLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(0, 0)},
	{name: "-+->", arity: 2, level: temporalLevel, argMaxLevels: sanyMaxLevels(temporalLevel, temporalLevel), argWeights: sanyWeights(0, 0)},
	{name: "$AngleAct", arity: 2, level: actionLevel, argMaxLevels: sanyMaxLevels(actionLevel, variableLevel), argWeights: sanyWeights(0, 0)},
	{name: "$BoundedChoose", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$BoundedExists", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$BoundedForall", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$CartesianProd", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$Case", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$ConjList", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$DisjList", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$Except", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$FcnApply", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, actionLevel), argWeights: sanyWeights(1, 1)},
	{name: "$FcnConstructor", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$IfThenElse", arity: 3, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel, temporalLevel, temporalLevel), argWeights: sanyWeights(1, 1, 1)},
	{name: "$NonRecursiveFcnSpec", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$Pair", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel, temporalLevel), argWeights: sanyWeights(1, 1)},
	{name: "$RcdConstructor", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$RcdSelect", arity: 2, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel, constantLevel), argWeights: sanyWeights(1, 1)},
	{name: "$RecursiveFcnSpec", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$Seq", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$SetEnumerate", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$SetOfAll", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$SetOfFcns", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$SetOfRcds", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$SF", arity: 2, level: temporalLevel, argMaxLevels: sanyMaxLevels(variableLevel, actionLevel), argWeights: sanyWeights(0, 0)},
	{name: "$SquareAct", arity: 2, level: actionLevel, argMaxLevels: sanyMaxLevels(actionLevel, variableLevel), argWeights: sanyWeights(0, 0)},
	{name: "$SubsetOf", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$TemporalExists", arity: 1, level: temporalLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(0)},
	{name: "$TemporalForall", arity: 1, level: temporalLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(0)},
	{name: "$TemporalWhile", arity: 2, level: temporalLevel, argMaxLevels: sanyMaxLevels(temporalLevel, temporalLevel), argWeights: sanyWeights(0, 0)},
	{name: "$Tuple", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$UnboundedChoose", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$UnboundedExists", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$UnboundedForall", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$WF", arity: 2, level: temporalLevel, argMaxLevels: sanyMaxLevels(variableLevel, actionLevel), argWeights: sanyWeights(0, 0)},
	{name: "$Nop", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$Qed", arity: 0, level: constantLevel, argMaxLevels: sanyMaxLevels(), argWeights: sanyWeights()},
	{name: "$Pfcase", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$Have", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$Take", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$Pick", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
	{name: "$Witness", arity: -1, level: constantLevel, argMaxLevels: sanyMaxLevels(actionLevel), argWeights: sanyWeights(1)},
	{name: "$Suffices", arity: 1, level: constantLevel, argMaxLevels: sanyMaxLevels(temporalLevel), argWeights: sanyWeights(1)},
}

// Context initializes its global table on first use, then reInit replaces it
// before each full frontend parse. A spec retains that table's builtin nodes.
var sanyInitialContextState struct {
	sync.Mutex
	context        *sanyContext
	runtimeContext *tlc.SemanticContext
}

func sanyGlobalInitialContext(reinitialize bool) *sanyContext {
	sanyInitialContextState.Lock()
	defer sanyInitialContextState.Unlock()
	if sanyInitialContextState.context == nil {
		sanyResetInitialContexts()
	}
	if reinitialize {
		sanyResetInitialContexts()
	}
	return sanyInitialContextState.context
}

func sanyResetInitialContexts() {
	sanyInitialContextState.context = newSanyInitialContext()
	runtime := tlc.NewSemanticContext(nil)
	for _, symbol := range sanyInitialContextState.context.orderedSymbols() {
		source := symbol.(*sanySemOpDefNode)
		runtime.AddSymbolToContext(
			tlc.SemanticContextKey{Name: tlc.UniqueStringOf(source.semName())},
			sanyRuntimeBuiltinDefinition(source),
		)
	}
	sanyInitialContextState.runtimeContext = runtime
}

func sanyRuntimeBuiltinDefinition(source *sanySemOpDefNode) *tlc.OpDefNode {
	symbol := tlc.NewSymbolNode(source.semName())
	node := tlc.NewOpDefNodeForSymbolWithBase(symbol, nil, nil, source.SemanticNodeBase)
	symbol.Kind, symbol.Arity = tlc.SymbolBuiltIn, source.semArity()
	node.SetNumberOfArgs(source.semArity())
	transferCanonicalOpDefLevelMetadata(node, source)
	transferCanonicalOpDefRenderMetadata(node, source)
	if source.formalNodes == nil {
		node.Params = nil
	} else {
		node.Params = make([]*tlc.SymbolNode, len(source.formalNodes))
		for i, parameter := range source.formalNodes {
			// This is a public view of the already allocated SANY formal. Do not
			// consume and discard another semantic UID before attaching its base.
			node.Params[i] = &tlc.SymbolNode{
				SemanticBase: parameter.SemanticNodeBase,
				Name:         tlc.UniqueStringOf(parameter.semName()),
				Arity:        parameter.semArity(),
				Kind:         tlc.SymbolFormalParam,
				Location:     parameter.Location,
				TreeNode:     parameter.TreeNode,
			}
		}
	}
	bindSanyRuntimeLevelNode(source.sanyLevelData, source, func(candidate sanySemSymbol) *tlc.SymbolNode {
		candidate = sanyLevelSymbolReference(candidate)
		if candidate == source {
			return node.Symbol
		}
		for i, formal := range source.formalNodes {
			if candidate == formal {
				return node.Params[i]
			}
		}
		return nil
	})
	return node
}

// SanyContext adapts tla2sany.semantic.Context. Its table, pair history,
// enumeration, duplication and typed accessors are implemented by the shared
// runtime semantic context rather than by a second public symbol model.
type SanyContext = tlc.SemanticContext

type SanyContextKey = tlc.SemanticContextKey

type SanyContextSymbolEnumeration = tlc.SemanticContextSymbolEnumeration

// NewSanyContext constructs an empty semantic context owned by table.
func NewSanyContext(table *tlc.ExternalModuleTable) *SanyContext {
	return tlc.NewSemanticContext(table)
}

// SanyContextReInit replaces the process-wide built-in context and its nodes.
func SanyContextReInit() { sanyGlobalInitialContext(true) }

// SanyContextGetGlobalContext returns the current process-wide built-in context.
func SanyContextGetGlobalContext() *SanyContext {
	sanyInitialContextState.Lock()
	defer sanyInitialContextState.Unlock()
	if sanyInitialContextState.context == nil {
		sanyResetInitialContexts()
	}
	return sanyInitialContextState.runtimeContext
}

// SanyContextIsBuiltIn uses source semantic identity in the current global
// context. Split Go views of one Java node share its class, kind and UID.
func SanyContextIsBuiltIn(node tlc.SemanticNode) bool {
	for _, symbol := range SanyContextGetGlobalContext().Content() {
		if tlc.SemanticEquals(symbol, node) {
			return true
		}
	}
	return false
}

// SanySemanticNodeIsBuiltIn adapts SemanticNode.isBuiltIn(). The source asks
// the current global Context by identity; a matching kind or name is not enough.
func SanySemanticNodeIsBuiltIn(node tlc.SemanticNode) bool {
	if sanyExploreNull(node) {
		panic(tlc.NewNullPointerException())
	}
	return SanyContextIsBuiltIn(node)
}

func (g *sanyExpressionGeneration) initialBuiltin(name string) *sanySemOpDefNode {
	var context *sanyContext
	if g.spec != nil && g.spec.initialContext != nil {
		context = g.spec.initialContext
	} else {
		context = sanyGlobalInitialContext(false)
	}
	symbol, _ := context.getSymbol(ResolveSanyOperatorSynonymString(name)).(*sanySemOpDefNode)
	return symbol
}
