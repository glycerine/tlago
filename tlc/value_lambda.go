package tlc

import "strings"

type FcnParams struct {
	Formals  [][]*SymbolNode
	IsTuples []bool
	Domains  []Value
	ArgLen   int
}

func NewFcnParams(formals [][]*SymbolNode, isTuples []bool, domains []Value) *FcnParams {
	outFormals := make([][]*SymbolNode, len(formals))
	for i, ids := range formals {
		outFormals[i] = make([]*SymbolNode, len(ids))
		copy(outFormals[i], ids)
	}
	outIsTuples := make([]bool, len(isTuples))
	copy(outIsTuples, isTuples)
	outDomains := make([]Value, len(domains))
	copy(outDomains, domains)
	params := &FcnParams{
		Formals:  outFormals,
		IsTuples: outIsTuples,
		Domains:  outDomains,
	}
	for i := range params.Formals {
		if i < len(params.IsTuples) && params.IsTuples[i] {
			params.ArgLen++
		} else {
			params.ArgLen += len(params.Formals[i])
		}
	}
	return params
}

func NewSingleFcnParam(formal *SymbolNode, domain Value) *FcnParams {
	return NewFcnParams([][]*SymbolNode{{formal}}, []bool{false}, []Value{domain})
}

func NewTupleFcnParam(formals []*SymbolNode, domain Value) *FcnParams {
	return NewFcnParams([][]*SymbolNode{formals}, []bool{true}, []Value{domain})
}

func (p *FcnParams) Length() int {
	if p == nil {
		return 0
	}
	return p.ArgLen
}

func (p *FcnParams) Size() (int, error) {
	if p == nil {
		return 0, nil
	}
	size := int64(1)
	for i, domain := range p.Domains {
		domainSize, err := domain.Size()
		if err != nil {
			return 0, err
		}
		repeat := 1
		if i >= len(p.IsTuples) || !p.IsTuples[i] {
			repeat = len(p.Formals[i])
		}
		for j := 0; j < repeat; j++ {
			size *= int64(domainSize)
			if size < -2147483648 || size > 2147483647 {
				return 0, newTLCError(ECGeneral, "overflow when computing the number of elements in:\n%s", p)
			}
		}
	}
	return int(size), nil
}

func (p *FcnParams) Elements() ValueEnumeration {
	if p == nil || p.ArgLen == 0 {
		return EmptySet.Elements()
	}
	if p.ArgLen == 1 {
		enum, ok := asEnumerable(p.Domains[0])
		if !ok {
			return newErrorEnumeration(newTLCError(ECGeneral, "the domains of formal parameters must be enumerable"))
		}
		return enum.Elements()
	}
	return newFcnParamsEnumeration(p)
}

func (p *FcnParams) String() string {
	if p == nil || len(p.Domains) == 0 {
		return ""
	}
	parts := make([]string, len(p.Domains))
	for i, domain := range p.Domains {
		ids := p.Formals[i]
		var lhs string
		if i < len(p.IsTuples) && p.IsTuples[i] {
			names := make([]string, len(ids))
			for j, id := range ids {
				names[j] = id.String()
			}
			lhs = "<<" + strings.Join(names, ", ") + ">>"
		} else if len(ids) != 0 {
			lhs = ids[0].String()
		}
		parts[i] = lhs + " \\in " + domain.String()
	}
	return strings.Join(parts, ", ")
}

type fcnParamsEnumeration struct {
	params       *FcnParams
	enums        []ValueEnumeration
	currentElems []Value
	done         bool
	err          error
}

func newFcnParamsEnumeration(params *FcnParams) *fcnParamsEnumeration {
	out := &fcnParamsEnumeration{
		params:       params,
		enums:        make([]ValueEnumeration, params.ArgLen),
		currentElems: make([]Value, params.ArgLen),
	}
	idx := 0
	for i, domain := range params.Domains {
		enumDomain, ok := asEnumerable(domain)
		if !ok {
			out.err = newTLCError(ECGeneral, "the domains of the parameters must be enumerable")
			out.done = true
			return out
		}
		if params.IsTuples[i] {
			out.enums[idx] = enumDomain.Elements()
			out.currentElems[idx] = out.enums[idx].NextElement()
			if err := out.enums[idx].Err(); err != nil {
				out.err = err
				out.done = true
				return out
			}
			if out.currentElems[idx] == nil {
				out.done = true
				return out
			}
			idx++
			continue
		}
		for j := 0; j < len(params.Formals[i]); j++ {
			out.enums[idx] = enumDomain.Elements()
			out.currentElems[idx] = out.enums[idx].NextElement()
			if err := out.enums[idx].Err(); err != nil {
				out.err = err
				out.done = true
				return out
			}
			if out.currentElems[idx] == nil {
				out.done = true
				return out
			}
			idx++
		}
	}
	return out
}

func (e *fcnParamsEnumeration) Reset() {
	if e.err != nil {
		return
	}
	for i := range e.enums {
		e.enums[i].Reset()
		e.currentElems[i] = e.enums[i].NextElement()
		if err := e.enums[i].Err(); err != nil {
			e.err = err
			e.done = true
			return
		}
		if e.currentElems[i] == nil {
			e.done = true
			return
		}
	}
	e.done = false
}

func (e *fcnParamsEnumeration) NextElement() Value {
	if e.done || e.err != nil {
		return nil
	}
	elems := make([]Value, len(e.currentElems))
	copy(elems, e.currentElems)
	for i := range e.currentElems {
		e.currentElems[i] = e.enums[i].NextElement()
		if err := e.enums[i].Err(); err != nil {
			e.err = err
			e.done = true
			break
		}
		if e.currentElems[i] != nil {
			break
		}
		if i == len(e.currentElems)-1 {
			e.done = true
			break
		}
		e.enums[i].Reset()
		e.currentElems[i] = e.enums[i].NextElement()
		if err := e.enums[i].Err(); err != nil {
			e.err = err
			e.done = true
			break
		}
	}
	return NewTupleValue(elems)
}

func (e *fcnParamsEnumeration) Err() error { return e.err }

type FcnLambdaValue struct {
	BaseValue
	Params  *FcnParams
	Body    SemanticNode
	Excepts []ValueExcept
	Tool    *Tool
	Con     *Context
	State   *TLCStateMut
	PState  *TLCStateMut
	Control int
	CM      CostModel
	FcnRcd  *FcnRcdValue
}

func NewFcnLambdaValue(params *FcnParams, body SemanticNode, tool *Tool, con *Context, state *TLCStateMut, pstate *TLCStateMut, control int, cms ...CostModel) *FcnLambdaValue {
	if con == nil {
		con = EmptyContext
	}
	cm := DoNotRecordCostModel
	if len(cms) > 0 {
		cm = cms[0]
	}
	return &FcnLambdaValue{
		Params:  params,
		Body:    body,
		Tool:    tool,
		Con:     con,
		State:   copyTLCStateForLambda(state),
		PState:  copyTLCStateForLambda(pstate),
		Control: control,
		CM:      cm,
	}
}

func NewFcnLambdaValueFrom(other *FcnLambdaValue, tool *Tool) *FcnLambdaValue {
	if other == nil {
		return nil
	}
	if tool == nil {
		tool = other.Tool
	}
	return &FcnLambdaValue{
		Params:  other.Params,
		Body:    other.Body,
		Excepts: copyValueExcepts(other.Excepts),
		Tool:    tool,
		Con:     other.Con,
		State:   other.State,
		PState:  other.PState,
		Control: other.Control,
		FcnRcd:  other.FcnRcd,
	}
}

func copyTLCStateForLambda(state *TLCStateMut) *TLCStateMut {
	if state == nil {
		return nil
	}
	return state.Copy()
}

func (v *FcnLambdaValue) Kind() ValueKind    { return FcnLambdaValueKind }
func (v *FcnLambdaValue) KindString() string { return v.KindStringFor(v.Kind()) }

func (v *FcnLambdaValue) MakeRecursive(fname *SymbolNode) {
	if v.Con == nil {
		v.Con = EmptyContext
	}
	v.Con = v.Con.Cons(fname, v)
	v.Control = EvalSetKeepLazy(v.Control)
}

func (v *FcnLambdaValue) Compare(other Value) (int, error) {
	fcn, err := v.materializeFcnRcd()
	if err != nil {
		return 0, err
	}
	return fcn.Compare(other)
}

func (v *FcnLambdaValue) Equal(other Value) (bool, error) {
	fcn, err := v.materializeFcnRcd()
	if err != nil {
		return false, err
	}
	return fcn.Equal(other)
}

func (v *FcnLambdaValue) Member(elem Value) (bool, error) {
	return false, v.unsupported("attempted to check if the value:\n%s\nis an element of the function %s", elem, v)
}

func (v *FcnLambdaValue) IsFinite() (bool, error) {
	return false, v.unsupported("attempted to check if the function:\n%s\nis a finite set", v)
}

func (v *FcnLambdaValue) Apply(arg Value) (Value, error) {
	return v.ApplyWithControl(arg, EvalClear)
}

func (v *FcnLambdaValue) ApplyWithControl(arg Value, control int) (Value, error) {
	if v.FcnRcd != nil {
		return v.FcnRcd.Apply(arg)
	}
	res, matches, matchedTerminal, err := v.matchExcepts(arg)
	if err != nil {
		return nil, err
	}
	if !matchedTerminal {
		ctx, returnsNull, err := v.bindArgumentForApply(arg)
		if err != nil || returnsNull {
			return nil, err
		}
		res, err = v.evalBody(ctx, control)
		if err != nil {
			return nil, err
		}
	}
	return takeMatchedExcepts(res, matches)
}

func (v *FcnLambdaValue) ApplyArgs(args []Value, control int) (Value, error) {
	if len(args) == 1 {
		return v.ApplyWithControl(args[0], control)
	}
	return v.ApplyWithControl(NewTupleValue(args), control)
}

func (v *FcnLambdaValue) bindArgumentForApply(arg Value) (*Context, bool, error) {
	if v.Params == nil {
		return v.Con, true, nil
	}
	ctx := v.Con
	if ctx == nil {
		ctx = EmptyContext
	}
	formals := v.Params.Formals
	domains := v.Params.Domains
	isTuples := v.Params.IsTuples
	if v.Params.Length() == 1 {
		in, err := domains[0].Member(arg)
		if err != nil {
			return nil, false, err
		}
		if !in {
			return nil, false, v.unsupported("in applying the function\n%s,\nthe first argument is:\n%s\nwhich is not in its domain", v, arg)
		}
		if isTuples[0] {
			ids := formals[0]
			argTuple := asTupleValue(arg)
			if argTuple == nil {
				return nil, false, v.unsupported("in applying the function\n%s,\nthe first argument is:\n%s\nwhich does not match its formal parameter", v, arg)
			}
			if len(argTuple.Elems) != len(ids) {
				return nil, true, nil
			}
			for i, id := range ids {
				ctx = ctx.Cons(id, argTuple.Elems[i])
			}
			return ctx, false, nil
		}
		if len(formals[0]) != 0 {
			ctx = ctx.Cons(formals[0][0], arg)
		}
		return ctx, false, nil
	}
	argTuple := asTupleValue(arg)
	if argTuple == nil {
		return nil, false, v.unsupported("in applying the function\n%s,\nthe argument list is:\n%s\nwhich does not match its formal parameter", v, arg)
	}
	elems := argTuple.Elems
	argn := 0
	for i, ids := range formals {
		domain := domains[i]
		if isTuples[i] {
			if argn >= len(elems) {
				return nil, true, nil
			}
			in, err := domain.Member(elems[argn])
			if err != nil {
				return nil, false, err
			}
			if !in {
				return nil, false, v.unsupported("in applying the function\n%s,\nthe argument number %d is:\n%s\nwhich is not in its domain", v, argn+1, elems[argn])
			}
			tv := asTupleValue(elems[argn])
			argn++
			if tv == nil || len(tv.Elems) != len(ids) {
				return nil, false, v.unsupported("in applying the function\n%s,\nthe argument number %d is:\n%s\nwhich does not match its formal parameter", v, argn, elems[argn-1])
			}
			for j, id := range ids {
				ctx = ctx.Cons(id, tv.Elems[j])
			}
			continue
		}
		for _, id := range ids {
			if argn >= len(elems) {
				return nil, true, nil
			}
			in, err := domain.Member(elems[argn])
			if err != nil {
				return nil, false, err
			}
			if !in {
				domainValue, derr := v.GetDomain()
				if derr != nil {
					return nil, false, derr
				}
				return nil, false, v.unsupported("in applying the function\n%s,\nthe argument number %d is:\n%s\nwhich is not in the function's domain %s", v, argn+1, elems[argn], domainValue)
			}
			ctx = ctx.Cons(id, elems[argn])
			argn++
		}
	}
	return ctx, false, nil
}

func (v *FcnLambdaValue) Select(arg Value) (Value, error) {
	if v.FcnRcd != nil {
		return v.FcnRcd.Select(arg)
	}
	res, matches, matchedTerminal, err := v.matchExcepts(arg)
	if err != nil {
		return nil, err
	}
	if !matchedTerminal {
		ctx, ok, err := v.bindArgument(arg)
		if err != nil || !ok {
			return nil, err
		}
		res, err = v.evalBody(ctx, v.Control)
		if err != nil {
			return nil, err
		}
	}
	return takeMatchedExcepts(res, matches)
}

func (v *FcnLambdaValue) evalBody(ctx *Context, control int) (Value, error) {
	if ctx == nil {
		ctx = EmptyContext
	}
	if v.Tool == nil {
		return ValUndef, nil
	}
	return v.Tool.Eval(v.Body, ctx, v.State, v.PState, control, v.CM)
}

func (v *FcnLambdaValue) bindArgument(arg Value) (*Context, bool, error) {
	if v.Params == nil {
		return v.Con, false, nil
	}
	ctx := v.Con
	if ctx == nil {
		ctx = EmptyContext
	}
	formals := v.Params.Formals
	domains := v.Params.Domains
	isTuples := v.Params.IsTuples
	if v.Params.Length() == 1 {
		in, err := domains[0].Member(arg)
		if err != nil || !in {
			return ctx, false, err
		}
		if isTuples[0] {
			ids := formals[0]
			argTuple := asTupleValue(arg)
			if argTuple == nil {
				return ctx, false, v.unsupported("in applying the function\n%s,\nthe first argument is:\n%s\nwhich does not match its formal parameter", v, arg)
			}
			if len(argTuple.Elems) != len(ids) {
				return ctx, false, nil
			}
			for i, id := range ids {
				ctx = ctx.Cons(id, argTuple.Elems[i])
			}
			return ctx, true, nil
		}
		if len(formals[0]) == 0 {
			return ctx, true, nil
		}
		return ctx.Cons(formals[0][0], arg), true, nil
	}
	argTuple := asTupleValue(arg)
	if argTuple == nil {
		return ctx, false, v.unsupported("in applying the function\n%s,\nthe argument list is:\n%s\nwhich does not match its formal parameter", v, arg)
	}
	elems := argTuple.Elems
	argn := 0
	for i, ids := range formals {
		domain := domains[i]
		if isTuples[i] {
			if argn >= len(elems) {
				return ctx, false, nil
			}
			in, err := domain.Member(elems[argn])
			if err != nil || !in {
				return ctx, false, err
			}
			tv := asTupleValue(elems[argn])
			argn++
			if tv == nil {
				return ctx, false, v.unsupported("in applying the function\n%s,\nthe argument number %d is:\n%s\nwhich does not match its formal parameter", v, argn, elems[argn-1])
			}
			if len(tv.Elems) != len(ids) {
				return ctx, false, nil
			}
			for j, id := range ids {
				ctx = ctx.Cons(id, tv.Elems[j])
			}
			continue
		}
		for _, id := range ids {
			if argn >= len(elems) {
				return ctx, false, nil
			}
			in, err := domain.Member(elems[argn])
			if err != nil || !in {
				return ctx, false, err
			}
			ctx = ctx.Cons(id, elems[argn])
			argn++
		}
	}
	return ctx, true, nil
}

func (v *FcnLambdaValue) bindEnumeratedArgument(arg Value) (*Context, error) {
	ctx := v.Con
	if ctx == nil {
		ctx = EmptyContext
	}
	if v.Params.Length() == 1 {
		if v.Params.IsTuples[0] {
			tuple := asTupleValue(arg)
			if tuple == nil {
				return nil, v.unsupported("in applying the function\n%s,\nthe first argument is:\n%s\nwhich does not match its formal parameter", v, arg)
			}
			for i, id := range v.Params.Formals[0] {
				ctx = ctx.Cons(id, tuple.Elems[i])
			}
			return ctx, nil
		}
		if len(v.Params.Formals[0]) != 0 {
			ctx = ctx.Cons(v.Params.Formals[0][0], arg)
		}
		return ctx, nil
	}
	argTuple := asTupleValue(arg)
	if argTuple == nil {
		return nil, v.unsupported("in applying the function\n%s,\nthe argument list is:\n%s\nwhich does not match its formal parameter", v, arg)
	}
	argn := 0
	for i, ids := range v.Params.Formals {
		if v.Params.IsTuples[i] {
			tv := asTupleValue(argTuple.Elems[argn])
			argn++
			if tv == nil {
				return nil, v.unsupported("in applying the function\n%s,\nthe argument number %d is:\n%s\nwhich does not match its formal parameter", v, argn, argTuple.Elems[argn-1])
			}
			for j, id := range ids {
				ctx = ctx.Cons(id, tv.Elems[j])
			}
			continue
		}
		for _, id := range ids {
			ctx = ctx.Cons(id, argTuple.Elems[argn])
			argn++
		}
	}
	return ctx, nil
}

func (v *FcnLambdaValue) matchExcepts(arg Value) (Value, []ValueExcept, bool, error) {
	var res Value
	matchedTerminal := false
	var matches []ValueExcept
	for i := len(v.Excepts) - 1; i >= 0; i-- {
		ex := v.Excepts[i]
		cur := ex.Current()
		if cur == nil {
			continue
		}
		eq, err := cur.Equal(arg)
		if err != nil {
			return nil, nil, false, err
		}
		if !eq {
			continue
		}
		if ex.IsLast() {
			res = ex.Value
			matchedTerminal = true
			break
		}
		matches = append(matches, ex.Advanced())
	}
	for i, j := 0, len(matches)-1; i < j; i, j = i+1, j-1 {
		matches[i], matches[j] = matches[j], matches[i]
	}
	return res, matches, matchedTerminal, nil
}

func takeMatchedExcepts(value Value, matches []ValueExcept) (Value, error) {
	if len(matches) == 0 || value == nil {
		return value, nil
	}
	return value.TakeExcepts(matches)
}

func (v *FcnLambdaValue) TakeExcept(ex ValueExcept) (Value, error) {
	if ex.Index >= len(ex.Path) {
		return ex.Value, nil
	}
	if v.FcnRcd != nil {
		return v.FcnRcd.TakeExcept(ex)
	}
	fcn := NewFcnLambdaValueFrom(v, v.Tool)
	fcn.Excepts = append(copyValueExcepts(v.Excepts), copyValueExcept(ex))
	return fcn, nil
}

func (v *FcnLambdaValue) TakeExcepts(exs []ValueExcept) (Value, error) {
	if v.FcnRcd != nil {
		return v.FcnRcd.TakeExcepts(exs)
	}
	fcn := NewFcnLambdaValueFrom(v, v.Tool)
	if len(exs) == 0 {
		return fcn, nil
	}
	lastComplete := -1
	for i := len(exs) - 1; i >= 0; i-- {
		if exs[i].Index >= len(exs[i].Path) {
			lastComplete = i
			break
		}
	}
	if lastComplete >= 0 {
		fcn.Excepts = copyValueExcepts(exs[lastComplete+1:])
		return fcn, nil
	}
	fcn.Excepts = append(copyValueExcepts(v.Excepts), copyValueExcepts(exs)...)
	return fcn, nil
}

func (v *FcnLambdaValue) GetDomain() (Value, error) {
	if v.FcnRcd != nil {
		return v.FcnRcd.DomainValue(), nil
	}
	if v.Params == nil {
		return EmptySet, nil
	}
	if v.Params.Length() == 1 {
		return v.Params.Domains[0], nil
	}
	sets := make([]Value, 0, v.Params.Length())
	for i, domain := range v.Params.Domains {
		if v.Params.IsTuples[i] {
			sets = append(sets, domain)
			continue
		}
		for range v.Params.Formals[i] {
			sets = append(sets, domain)
		}
	}
	return NewSetOfTuplesValue(sets), nil
}

func (v *FcnLambdaValue) Size() (int, error) {
	if v.FcnRcd != nil {
		return v.FcnRcd.Size()
	}
	if v.Params == nil {
		return 0, nil
	}
	return v.Params.Size()
}

func (v *FcnLambdaValue) IsDefined() bool { return true }

func (v *FcnLambdaValue) DeepCopy() Value {
	fcn := NewFcnLambdaValueFrom(v, v.Tool)
	if v.FcnRcd != nil {
		if copied, ok := v.FcnRcd.DeepCopy().(*FcnRcdValue); ok {
			fcn.FcnRcd = copied
		}
	}
	return fcn
}

func (v *FcnLambdaValue) IsNormalized() bool {
	return v.FcnRcd != nil && v.FcnRcd.IsNormalized()
}

func (v *FcnLambdaValue) Normalize() Value {
	if v.FcnRcd != nil {
		v.FcnRcd.Normalize()
	}
	return v
}

func (v *FcnLambdaValue) DeepNormalize() {
	if v.FcnRcd != nil {
		v.FcnRcd.DeepNormalize()
		return
	}
	for i := range v.Excepts {
		v.Excepts[i].Value.DeepNormalize()
		for _, path := range v.Excepts[i].Path {
			path.DeepNormalize()
		}
	}
	if v.Params != nil {
		for _, domain := range v.Params.Domains {
			domain.DeepNormalize()
		}
	}
}

func (v *FcnLambdaValue) ToTuple() *TupleValue {
	if v.Params == nil || v.Params.Length() != 1 {
		return nil
	}
	domain := v.Params.Domains[0]
	if intv, ok := domain.(*IntervalValue); ok {
		size, err := intv.Size()
		if err != nil {
			return nil
		}
		if intv.Low != 1 && size != 0 {
			return nil
		}
		elems := make([]Value, size)
		for i := 0; i < size; i++ {
			elem, err := v.Select(NewIntValue(int32(i + 1)))
			if err != nil || elem == nil {
				return nil
			}
			elems[i] = elem
		}
		return NewTupleValue(elems)
	}
	set, err := toSetEnumValue(domain)
	if err != nil {
		return nil
	}
	set.Normalize()
	elems := make([]Value, set.Elems.Len())
	for i := 0; i < set.Elems.Len(); i++ {
		arg := set.Elems.At(i)
		iv, ok := arg.(*IntValue)
		if !ok || iv.Val != int32(i+1) {
			return nil
		}
		elem, err := v.Select(arg)
		if err != nil || elem == nil {
			return nil
		}
		elems[i] = elem
	}
	return NewTupleValue(elems)
}

func (v *FcnLambdaValue) ToRecord() *RecordValue {
	fcn, err := v.materializeFcnRcd()
	if err != nil || fcn == nil {
		return nil
	}
	return fcn.ToRecord()
}

func (v *FcnLambdaValue) ToFcnRcd() *FcnRcdValue {
	fcn, err := v.materializeFcnRcd()
	if err != nil {
		return nil
	}
	return fcn
}

func (v *FcnLambdaValue) materializeFcnRcd() (*FcnRcdValue, error) {
	if v.FcnRcd != nil {
		return v.FcnRcd, nil
	}
	if v.Params == nil {
		v.FcnRcd = EmptyFcn
		return v.FcnRcd, nil
	}
	size, err := v.Params.Size()
	if err != nil {
		return nil, err
	}
	domain := make([]Value, size)
	values := make([]Value, size)
	idx := 0
	enum := v.Params.Elements()
	for arg := enum.NextElement(); arg != nil; arg = enum.NextElement() {
		if idx >= size {
			return nil, v.unsupported("function parameter enumeration exceeded declared size for %s", v.Params)
		}
		domain[idx] = arg
		ctx, err := v.bindEnumeratedArgument(arg)
		if err != nil {
			return nil, err
		}
		values[idx], err = v.evalBody(ctx, v.Control)
		if err != nil {
			return nil, err
		}
		idx++
	}
	if err := enum.Err(); err != nil {
		return nil, err
	}
	values = values[:idx]
	domain = domain[:idx]
	if v.Params.Length() == 1 {
		if intv, ok := v.Params.Domains[0].(*IntervalValue); ok {
			v.FcnRcd = NewFcnRcdIntervalValue(intv, values)
		} else {
			v.FcnRcd = NewFcnRcdValue(domain, values, false)
		}
	} else {
		v.FcnRcd = NewFcnRcdValue(domain, values, false)
	}
	if len(v.Excepts) != 0 {
		taken, err := v.FcnRcd.TakeExcepts(copyValueExcepts(v.Excepts))
		if err != nil {
			return nil, err
		}
		fcn, ok := taken.(*FcnRcdValue)
		if !ok {
			return nil, v.unsupported("EXCEPT conversion of function lambda produced %T", taken)
		}
		v.FcnRcd = fcn
	}
	return v.FcnRcd, nil
}

func (v *FcnLambdaValue) FingerPrint(fp uint64) uint64 {
	fcn, err := v.materializeFcnRcd()
	if err != nil {
		panic(err)
	}
	return fcn.FingerPrint(fp)
}

func (v *FcnLambdaValue) Permute(perm *MVPerm) Value {
	fcn, err := v.materializeFcnRcd()
	if err != nil {
		return v
	}
	return fcn.Permute(perm)
}

func (v *FcnLambdaValue) String() string {
	if Globals.Expand || v.Params == nil {
		if fcn, err := v.materializeFcnRcd(); err == nil && fcn != nil {
			return fcn.String()
		}
	}
	return "[" + v.Params.String() + " |-> <expression " + toContextString(v.Body) + ">]"
}

func (ex ValueExcept) Current() Value {
	if ex.Index < 0 || ex.Index >= len(ex.Path) {
		return nil
	}
	return ex.Path[ex.Index]
}

func (ex ValueExcept) IsLast() bool {
	return ex.Index == len(ex.Path)-1
}

func (ex ValueExcept) Advanced() ValueExcept {
	next := copyValueExcept(ex)
	next.Index++
	return next
}

func copyValueExcept(ex ValueExcept) ValueExcept {
	path := make([]Value, len(ex.Path))
	copy(path, ex.Path)
	ex.Path = path
	return ex
}

func copyValueExcepts(exs []ValueExcept) []ValueExcept {
	if len(exs) == 0 {
		return nil
	}
	out := make([]ValueExcept, len(exs))
	for i, ex := range exs {
		out[i] = copyValueExcept(ex)
	}
	return out
}
