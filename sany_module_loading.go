package tlago

import "strings"

// A ParseUnit is a source file, not each inner module in that file. Relationships
// are recorded as names resolve; their order drives calculateDependencies.
type sanyLoadUnit struct {
	root       *Module
	filename   string
	extendees  []*sanyLoadUnit
	instancees []*sanyLoadUnit
}

type sanyLoadImport struct {
	name      string
	owner     *Module
	extension bool
}

func (l *sanyLoader) registerLoadUnit(root *Module, filename string) {
	unit := &sanyLoadUnit{root: root, filename: filename}
	l.parseUnits[root.Name] = unit
	var register func(*Module, *Module)
	register = func(mod, parent *Module) {
		l.moduleUnits[mod] = unit
		l.moduleParents[mod] = parent
		l.moduleBindings[mod] = map[string]*Module{}
		for _, nested := range mod.Nested {
			register(nested, mod)
		}
	}
	register(root, nil)
}

func sanyLoadingInstances(mod *Module) []Instance {
	if mod.Syntax == nil {
		return mod.Instances
	}
	var result []Instance
	var walk func(*SanySyntaxNode)
	walk = func(node *SanySyntaxNode) {
		if node == nil || (node != mod.Syntax && node.Kind.JavaName() == "N_Module") {
			return
		}
		if node.Kind.JavaName() == "N_NonLocalInstance" {
			heirs := node.GetHeirs()
			for i, child := range heirs {
				if child.Kind.JavaName() == "INSTANCE" && i+1 < len(heirs) {
					result = append(result, Instance{Module: heirs[i+1].Image, Pos: sanyNodePosition(heirs[i+1])})
					break
				}
			}
		}
		for _, child := range node.GetHeirs() {
			walk(child)
		}
	}
	walk(mod.Syntax)
	return result
}

func (l *sanyLoader) resolveLoadingModule(mod *Module, name string, instance bool) *Module {
	visited := map[*Module]bool{}
	var resolve func(*Module) *Module
	resolve = func(current *Module) *Module {
		if current == nil || visited[current] {
			return nil
		}
		visited[current] = true
		if target := l.moduleBindings[current][name]; target != nil {
			return target
		}
		// Inner modules are visible after their declaration. INSTANCE does not
		// export a module's inner modules; EXTENDS does.
		if instance && current == mod {
			for _, inst := range sanyLoadingInstances(current) {
				if inst.Module != name {
					continue
				}
				for _, inner := range current.Nested {
					if inner.Name == name && positionBefore(inner.Pos, inst.Pos) {
						return inner
					}
				}
				break
			}
		}
		for _, extendee := range current.Extends {
			if extended := l.moduleBindings[current][extendee]; extended != nil {
				if target := l.resolveLoadingInner(extended, name, map[*Module]bool{}); target != nil {
					return target
				}
			}
		}
		if parent := l.moduleParents[current]; parent != nil {
			for _, sibling := range parent.Nested {
				if sibling.Name == name && positionBefore(sibling.Pos, current.Pos) {
					return sibling
				}
			}
			return resolve(parent)
		}
		return nil
	}
	return resolve(mod)
}

// EXTENDS exposes inner modules of its parse units, not arbitrary bindings
// created while resolving those units' own external module names.
func (l *sanyLoader) resolveLoadingInner(mod *Module, name string, visited map[*Module]bool) *Module {
	if mod == nil || visited[mod] {
		return nil
	}
	visited[mod] = true
	for _, inner := range mod.Nested {
		if inner.Name == name {
			return inner
		}
	}
	for _, extended := range mod.Extends {
		if target := l.resolveLoadingInner(l.moduleBindings[mod][extended], name, visited); target != nil {
			return target
		}
	}
	return nil
}

// SpecObj restarts the search at the root after each binding. Exhaust all
// unresolved EXTENDS before looking for an unresolved INSTANCE.
func (l *sanyLoader) nextLoadingImport(root *Module, extension bool) *sanyLoadImport {
	visited := map[*Module]bool{}
	var search func(*Module) *sanyLoadImport
	search = func(mod *Module) *sanyLoadImport {
		if mod == nil || visited[mod] {
			return nil
		}
		visited[mod] = true
		instances := sanyLoadingInstances(mod)
		if extension {
			for _, name := range mod.Extends {
				if l.resolveLoadingModule(mod, name, false) == nil {
					return &sanyLoadImport{name, mod, true}
				}
			}
		} else {
			for _, inst := range instances {
				if l.resolveLoadingModule(mod, inst.Module, true) == nil {
					return &sanyLoadImport{inst.Module, mod, false}
				}
			}
		}
		for _, name := range mod.Extends {
			if found := search(l.resolveLoadingModule(mod, name, false)); found != nil {
				return found
			}
		}
		for _, inst := range instances {
			if found := search(l.resolveLoadingModule(mod, inst.Module, true)); found != nil {
				return found
			}
		}
		for _, nested := range mod.Nested {
			if found := search(nested); found != nil {
				return found
			}
		}
		return nil
	}
	return search(root)
}

func appendSanyLoadUnit(units []*sanyLoadUnit, unit *sanyLoadUnit) []*sanyLoadUnit {
	for _, existing := range units {
		if existing == unit {
			return units
		}
	}
	return append(units, unit)
}

func (l *sanyLoader) checkLoadingCycle(start *sanyLoadUnit) {
	visited := map[*sanyLoadUnit]bool{}
	var search func(*sanyLoadUnit, []*sanyLoadUnit)
	search = func(candidate *sanyLoadUnit, path []*sanyLoadUnit) {
		if visited[candidate] {
			return
		}
		visited[candidate] = true
		// Source Vector.appendNoRepeats mutates getExtendees(). Preserve that
		// observable relationship order for calculateDependencies as well.
		for _, unit := range candidate.instancees {
			candidate.extendees = appendSanyLoadUnit(candidate.extendees, unit)
		}
		references := candidate.extendees
		for _, target := range references {
			if target == start {
				var names []string
				for _, unit := range path {
					names = append(names, unit.filename)
				}
				names = append(names, start.filename)
				cycle := strings.Join(names, " --> ")
				diagnostic := errorAt(Position{}, "E4222", "Circular dependency among .tla files; dependency cycle is:\n\n  %s", cycle)
				diagnostic.SANYParameters = []any{cycle}
				l.abortParse(diagnostic)
			}
			search(target, append(path, target))
		}
	}
	search(start, []*sanyLoadUnit{start})
}

func (l *sanyLoader) loadDependencies(root *Module) {
	// Source's short-circuited search retains instantiationFound when a
	// subsequent EXTENDS search succeeds; both relationships are then added.
	instantiationFound := false
	for {
		dependency := l.nextLoadingImport(root, true)
		if dependency == nil {
			dependency = l.nextLoadingImport(root, false)
			instantiationFound = dependency != nil
		}
		if dependency == nil {
			break
		}
		target := l.loadModule(dependency.name, dependency.owner)
		if target == nil || l.diags.HasErrors() {
			return
		}
		ownerUnit, targetUnit := l.moduleUnits[dependency.owner], l.moduleUnits[target]
		if dependency.extension {
			ownerUnit.extendees = appendSanyLoadUnit(ownerUnit.extendees, targetUnit)
		}
		if instantiationFound {
			ownerUnit.instancees = appendSanyLoadUnit(ownerUnit.instancees, targetUnit)
		}
		l.checkLoadingCycle(targetUnit)
		l.moduleBindings[dependency.owner][dependency.name] = target
	}
	seen := map[*sanyLoadUnit]bool{}
	var order func(*sanyLoadUnit)
	order = func(unit *sanyLoadUnit) {
		if seen[unit] {
			return
		}
		for _, extendee := range unit.extendees {
			order(extendee)
		}
		for _, instancee := range unit.instancees {
			order(instancee)
		}
		seen[unit] = true
		l.semanticOrder = append(l.semanticOrder, unit.root.Name)
	}
	order(l.moduleUnits[root])
}
