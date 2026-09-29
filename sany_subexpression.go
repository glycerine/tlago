package tlago

import (
	"strconv"
	"strings"
)

func sanySubexpressionPath(name string) (string, []int, bool) {
	parts := strings.Split(name, "!")
	if len(parts) < 2 || parts[0] == "" {
		return "", nil, false
	}
	selectors := make([]int, 0, len(parts)-1)
	for _, part := range parts[1:] {
		if part == "" {
			return "", nil, false
		}
		n, err := strconv.Atoi(part)
		if err != nil || n <= 0 {
			return "", nil, false
		}
		selectors = append(selectors, n)
	}
	return parts[0], selectors, true
}

func sanySelectSubexpression(expr Expr, selectors []int) Expr {
	cur := expr
	for _, selector := range selectors {
		children := sanySubexpressionChildren(cur)
		if selector <= 0 || selector > len(children) {
			return nil
		}
		cur = children[selector-1]
	}
	return cur
}

func sanySubexpressionChildren(expr Expr) []Expr {
	switch e := expr.(type) {
	case *UnaryExpr:
		return []Expr{e.Expr}
	case *BinaryExpr:
		if e.JunctionList {
			var out []Expr
			var collect func(Expr)
			collect = func(item Expr) {
				if nested, ok := item.(*BinaryExpr); ok && sanySameJunctionFrame(e, nested) {
					collect(nested.Left)
					collect(nested.Right)
					return
				}
				out = append(out, item)
			}
			collect(e.Left)
			collect(e.Right)
			return out
		}
		return []Expr{e.Left, e.Right}
	case *CallExpr:
		children := make([]Expr, 0, 1+len(e.Args))
		children = append(children, e.Callee)
		children = append(children, e.Args...)
		return children
	case *IfExpr:
		return []Expr{e.Cond, e.Then, e.Else}
	case *LetExpr:
		return []Expr{e.Body}
	case *QuantifierExpr:
		if e.Set != nil {
			return []Expr{e.Set, e.Body}
		}
		return []Expr{e.Body}
	case *CaseExpr:
		children := make([]Expr, 0, len(e.Arms)*2+1)
		for _, arm := range e.Arms {
			children = append(children, arm.Test, arm.Value)
		}
		if e.Other != nil {
			children = append(children, e.Other)
		}
		return children
	case *ChooseExpr:
		return []Expr{e.Set, e.Body}
	case *TupleExpr:
		return e.Elems
	case *SetExpr:
		return e.Elems
	case *RecordExpr:
		children := make([]Expr, 0, len(e.Fields))
		for _, field := range e.Fields {
			children = append(children, field.Value)
		}
		return children
	case *RecordComponentExpr:
		return []Expr{e.Record}
	case *RecordSetExpr:
		children := make([]Expr, 0, len(e.Fields))
		for _, field := range e.Fields {
			children = append(children, field.Set)
		}
		return children
	case *FunctionExpr:
		children := make([]Expr, 0, len(e.Bounds)+1)
		for _, bound := range e.Bounds {
			children = append(children, bound.Set)
		}
		children = append(children, e.Body)
		return children
	case *FunctionAppExpr:
		children := make([]Expr, 0, 1+len(e.Args))
		children = append(children, e.Function)
		children = append(children, e.Args...)
		return children
	case *ExceptExpr:
		children := []Expr{e.Base}
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				children = append(children, component.Indices...)
			}
			children = append(children, spec.Value)
		}
		return children
	case *LabelExpr:
		return []Expr{e.Body}
	case *ActionExpr:
		return []Expr{e.Action, e.Subscript}
	case *FairnessExpr:
		return []Expr{e.Action, e.Subscript}
	case *FunctionSetExpr:
		return []Expr{e.Domain, e.Range}
	case *SetComprehensionExpr:
		children := make([]Expr, 0, len(e.Bounds)+2)
		for _, bound := range e.Bounds {
			children = append(children, bound.Set)
		}
		children = append(children, e.Element)
		if e.Predicate != nil {
			children = append(children, e.Predicate)
		}
		return children
	default:
		return nil
	}
}
