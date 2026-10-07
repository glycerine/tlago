package tlago

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/glycerine/tlago/tlc"
)

// SanyFrontEndMain is the legacy SANY.frontEndMain boundary with silent output.
// It runs parsing and semantic analysis, retaining ordinary semantic diagnostics
// on the spec. Like Java's default doStrictErrorCodes=false, only a parsing-phase
// failure returns ERROR(-1); ordinary semantic diagnostics do not change OK(0).
// Unexpected runtime failures propagate, as FrontEndException does in Java.
func SanyFrontEndMain(file string, opts LoadOptions) (*Spec, int) {
	spec, _, _, parseFailed := runSanyFrontEnd(file, opts, nil)
	if parseFailed {
		return spec, -1
	}
	return spec, 0
}

func CheckSanySource(file, source string) (*Spec, Diagnostics) {
	return CheckSanySourceWithOptions(file, source, LoadOptions{})
}

func CheckSanySourceWithOptions(file, source string, opts LoadOptions) (*Spec, Diagnostics) {
	spec, diags := ParseSanySpecSource(file, source, opts)
	if !diags.HasErrors() {
		diags = append(diags, CheckSpec(spec)...)
	}
	spec.Diags = diags
	return spec, diags
}

// ParseSanySpecSource parses source and resolves dependencies without semantic
// generation or level checking. Callers may inspect the distinct phases.
func ParseSanySpecSource(file, source string, opts LoadOptions) (*Spec, Diagnostics) {
	initialContext := sanyGlobalInitialContext(true)
	mod, diags := ParseSanyModuleSource(file, source)
	loader := newSanyLoader(opts)
	loader.rootDir = "."
	if file != "" {
		loader.moduleFiles = append(loader.moduleFiles, file)
		rootPath := file
		if abs, err := filepath.Abs(rootPath); err == nil {
			rootPath = abs
		}
		loader.rootDir = filepath.Dir(rootPath)
	}
	loader.registerModuleRecursive(mod)
	if mod != nil {
		loader.registerLoadUnit(mod, file)
	}
	if diags.HasErrors() {
		spec := &Spec{initialContext: initialContext, LibraryPaths: append([]string(nil), opts.LibraryPaths...), Root: mod, Modules: loader.modules, SemanticOrder: append([]string(nil), loader.semanticOrder...), ModuleFiles: append([]string(nil), loader.moduleFiles...), Diags: diags}
		return spec, diags
	}
	if mod != nil {
		func() {
			defer func() {
				if failure := recover(); failure != nil {
					if _, ok := failure.(*sanyParseAbort); !ok {
						panic(failure)
					}
				}
			}()
			loader.loadDependencies(mod)
		}()
		diags = append(diags, loader.diags...)
	}
	spec := &Spec{initialContext: initialContext, LibraryPaths: append([]string(nil), opts.LibraryPaths...), Root: mod, Modules: loader.modules, SemanticOrder: append([]string(nil), loader.semanticOrder...), ModuleFiles: append([]string(nil), loader.moduleFiles...), Diags: diags}
	if diags.HasErrors() {
		spec.Diags = diags
		return spec, diags
	}
	spec.Diags = diags
	return spec, diags
}

func ModelCheckSanySource(specFile, specSource, cfgSource string, opts ModelCheckOptions) (ModelCheckResult, Diagnostics) {
	spec, diags := CheckSanySource(specFile, specSource)
	if diags.HasErrors() {
		return ModelCheckResult{}, diags
	}
	cfg, cfgDiags := ParseConfigSource(specFile+".cfg", cfgSource)
	if cfgDiags.HasErrors() {
		return ModelCheckResult{}, cfgDiags
	}
	result, runDiags := ModelCheck(spec, cfg, opts)
	return result, runDiags
}

type sanyLoader struct {
	initialContext  *sanyContext
	opts            LoadOptions
	modules         map[string]*Module
	diags           Diagnostics
	parseUnits      map[string]*sanyLoadUnit
	moduleUnits     map[*Module]*sanyLoadUnit
	moduleParents   map[*Module]*Module
	moduleBindings  map[*Module]map[string]*Module
	semanticOrder   []string
	moduleFiles     []string
	rootDir         string
	rootPath        string
	rootModule      *Module
	monolithTempDir string
}

// sanyParseAbort is Errors.addMessage's unrecoverable load failure. The TLC
// front end catches it at Java's frontEndParse boundary; the native API returns
// the same accumulated diagnostics to its caller.
type sanyParseAbort struct {
	diagnostic Diagnostic
}

func (failure *sanyParseAbort) Error() string {
	return sanyJavaErrorDetails(failure.diagnostic)
}

func (l *sanyLoader) abortParse(diagnostic Diagnostic) {
	l.diags = appendSanyDiagnostics(l.diags, diagnostic)
	panic(&sanyParseAbort{diagnostic: diagnostic})
}

func (l *sanyLoader) missingModule(name string, importer *Module) {
	diagnostic := errorAt(Position{}, "E4220", "Cannot find source file for module %s", name)
	diagnostic.SANYParameters = []any{name}
	if importer != nil {
		diagnostic = errorAt(Position{File: importer.Name}, "E4220", "Cannot find source file for module %s imported in module %s.", name, importer.Name)
		diagnostic.SANYParameters = []any{name, importer.Name}
	}
	diagnostic.SANYMessage = diagnostic.Message
	l.abortParse(diagnostic)
}

func LoadSanySpec(root string, opts LoadOptions) (spec *Spec, diags Diagnostics) {
	loader := newSanyLoader(opts)
	defer func() {
		if failure := recover(); failure != nil {
			if _, ok := failure.(*sanyParseAbort); !ok {
				panic(failure)
			}
			spec = loader.snapshot(nil)
			diags = loader.diags
		}
	}()
	return loader.loadSpec(root)
}

func newSanyLoader(opts LoadOptions) *sanyLoader {
	return &sanyLoader{
		opts:           opts,
		modules:        map[string]*Module{},
		parseUnits:     map[string]*sanyLoadUnit{},
		moduleUnits:    map[*Module]*sanyLoadUnit{},
		moduleParents:  map[*Module]*Module{},
		moduleBindings: map[*Module]map[string]*Module{},
	}
}

func (l *sanyLoader) loadSpec(root string) (*Spec, Diagnostics) {
	opts := l.opts
	rootPath := root
	if opts.FilenameResolver != nil {
		rootPath = tlc.ModuleFilename(rootPath)
	} else if filepath.Ext(rootPath) == "" {
		rootPath += ".tla"
	}
	rootFilename := rootPath
	if opts.FilenameResolver != nil {
		file := opts.FilenameResolver.Resolve(rootPath, true)
		if !file.Exists() {
			l.reportMissingFile(file)
			l.missingModule(strings.TrimSuffix(filepath.Base(rootPath), ".tla"), nil)
		}
		rootPath = file.GetPath()
	}
	if abs, err := filepath.Abs(rootPath); err == nil {
		rootPath = abs
	}
	l.rootDir = filepath.Dir(rootPath)
	l.rootPath = rootPath
	if _, err := os.Stat(rootPath); os.IsNotExist(err) {
		l.missingModule(strings.TrimSuffix(filepath.Base(rootPath), ".tla"), nil)
	}
	rootMod := l.loadPath(rootPath, false, rootFilename)
	l.rootModule = rootMod
	if rootMod != nil && !l.diags.HasErrors() {
		sourceExtendsLen := len(rootMod.Extends)
		rootMod.Extends = appendModuleNames(rootMod.Extends, l.opts.ExtraModules...)
		rootMod.ImplicitExtends = append([]string(nil), rootMod.Extends[sourceExtendsLen:]...)
		l.loadDependencies(rootMod)
	}
	return l.snapshot(rootMod), l.diags
}

func (l *sanyLoader) snapshot(root *Module) *Spec {
	if root == nil {
		root = l.rootModule
	}
	return &Spec{initialContext: l.initialContext, FilenameResolver: l.opts.FilenameResolver, LibraryPaths: append([]string(nil), l.opts.LibraryPaths...), Root: root, Modules: l.modules, SemanticOrder: append([]string(nil), l.semanticOrder...), ModuleFiles: append([]string(nil), l.moduleFiles...), Diags: l.diags}
}

func appendModuleNames(names []string, extra ...string) []string {
	seen := map[string]bool{}
	for _, name := range names {
		if name != "" {
			seen[name] = true
		}
	}
	for _, name := range extra {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		names = append(names, name)
	}
	return names
}

func (l *sanyLoader) loadModule(name string, importer *Module) *Module {
	if unit := l.parseUnits[name]; unit != nil {
		return unit.root
	}
	if l.opts.FilenameResolver != nil {
		file := l.opts.FilenameResolver.Resolve(name+".tla", true)
		if !file.Exists() {
			if mod := l.loadMonolithModule(name); mod != nil {
				return mod
			}
			l.missingModule(name, importer)
		}
		provenance := ""
		if path := file.GetLibraryPath(); path != nil {
			provenance = *path
		}
		return l.loadPath(file.GetPath(), file.IsLibraryModule(), name+".tla", provenance)
	}
	if l.opts.PreferLibraryModules {
		if mod := l.loadLibraryModule(name); mod != nil {
			return mod
		}
	}
	if src, ok := standardModules[name]; ok {
		l.moduleFiles = append(l.moduleFiles, name+".tla")
		mod, diags := ParseSanyModuleSource(name+".tla", src)
		l.diags = append(l.diags, diags...)
		if mod.Name != "" {
			setModuleLibraryRecursive(mod, true)
			l.modules[mod.Name] = mod
			l.registerLoadUnit(mod, name+".tla")
		}
		return mod
	}
	if !l.opts.PreferLibraryModules {
		if mod := l.loadLibraryModule(name); mod != nil {
			return mod
		}
	}
	if mod := l.loadMonolithModule(name); mod != nil {
		return mod
	}
	l.missingModule(name, importer)
	return nil
}

func (l *sanyLoader) reportMissingFile(file *tlc.TLAFile) {
	if l.opts.ResolutionError != nil {
		l.opts.ResolutionError(fmt.Sprintf("File does not exist: %s while looking in these directories: %s", file.GetAbsolutePath(), l.opts.FilenameResolver.GetFullPath()))
	}
}

func (l *sanyLoader) loadLibraryModule(name string) *Module {
	for i, dir := range append([]string{l.rootDir}, l.opts.LibraryPaths...) {
		path := filepath.Join(dir, name+".tla")
		if _, err := os.Stat(path); err == nil {
			mod := l.loadPath(path, false, name+".tla")
			if i > 0 {
				setModuleLibraryRecursive(mod, true)
			}
			return mod
		}
	}
	return nil
}

func (l *sanyLoader) loadPath(path string, standard bool, logicalFilename string, provenance ...string) *Module {
	data, err := os.ReadFile(path)
	if err != nil {
		l.diags = append(l.diags, errorAt(Position{File: path, Line: 1, Column: 1}, "E1202", "cannot read %s: %v", path, err))
		return nil
	}
	l.moduleFiles = append(l.moduleFiles, logicalFilename)
	l.reportParsing(path, provenance...)
	extracted := l.monolithTempDir != "" && filepath.Dir(path) == l.monolithTempDir
	mod, _, diags := parseSanyModuleSourceWithDependencies(path, string(data), extracted)
	parserFailure := false
	for _, diagnostic := range diags {
		if diagnostic.SANYParseMessage == "" {
			continue
		}
		parserFailure = true
		if l.opts.ParsingProgress != nil {
			l.opts.ParsingProgress(diagnostic.SANYParseMessage)
		}
	}
	if parserFailure {
		name := strings.TrimSuffix(filepath.Base(logicalFilename), ".tla")
		failure := errorAt(Position{File: name}, "E4003", "Could not parse module %s from file %s", name, logicalFilename)
		failure.SANYParameters = []any{name, logicalFilename}
		l.abortParse(failure)
	}
	l.diags = append(l.diags, diags...)
	if standard {
		setModuleLibraryRecursive(mod, true)
	}
	if mod.Name == "" {
		return mod
	}
	if !standard {
		fileMod := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
		if fileMod != mod.Name {
			diagnostic := errorAt(Position{}, "E4221", "File name '%s' does not match the name '%s' of the top level module it contains.", fileMod, mod.Name)
			diagnostic.SANYParameters = []any{fileMod, mod.Name}
			diagnostic.SANYMessage = diagnostic.Message
			l.abortParse(diagnostic)
		}
	}
	l.registerModuleRecursive(mod)
	l.registerLoadUnit(mod, logicalFilename)
	return mod
}

// FileUtil.createNamedInputStream falls back to the root monolith only after
// ordinary resolution fails. MonolithSpecExtractor keeps the module delimiters
// and writes an actual temporary file, whose provenance is the root file.
func (l *sanyLoader) loadMonolithModule(name string) *Module {
	if l.rootPath == "" {
		return nil
	}
	data, err := os.ReadFile(l.rootPath)
	if err != nil {
		return nil
	}
	text, active := tlc.ExtractMonolithModuleSource(string(data), name)
	if !active {
		return nil
	}
	if l.monolithTempDir == "" {
		l.monolithTempDir, err = os.MkdirTemp("", "tlago-monolith-")
		if err != nil {
			return nil
		}
	}
	path := filepath.Join(l.monolithTempDir, filepath.Base(name)+".tla")
	if err = os.WriteFile(path, []byte(text), 0600); err != nil {
		return nil
	}
	return l.loadPath(path, false, name+".tla", l.rootPath)
}

func (l *sanyLoader) reportParsing(path string, provenance ...string) {
	if l.opts.ParsingProgress == nil {
		return
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		resolved = path
	}
	if absolute, err := filepath.Abs(resolved); err == nil {
		resolved = absolute
	}
	origin := ""
	if len(provenance) > 0 && provenance[0] != "" {
		original := provenance[0]
		originalPath := original
		if uri, err := url.Parse(original); err == nil && uri.Scheme != "" {
			originalPath = uri.Path
		}
		if originalPath != resolved {
			origin = " (" + original + ")"
		}
	}
	l.opts.ParsingProgress(fmt.Sprintf("Parsing file %s%s", resolved, origin))
}

func (l *sanyLoader) registerModuleRecursive(mod *Module) {
	if mod == nil || mod.Name == "" {
		return
	}
	l.modules[mod.Name] = mod
	for _, nested := range mod.Nested {
		l.registerModuleRecursive(nested)
	}
}

func ParseSanyModuleSource(file, source string) (*Module, Diagnostics) {
	module, _, diags := parseSanyModuleSourceWithDependencies(file, source)
	return module, diags
}

func parseSanyModuleSourceWithDependencies(file, source string, extracted ...bool) (*Module, []string, Diagnostics) {
	root, dependencies, diags := parseSanySyntaxWithDependencies(file, source)
	if len(extracted) > 0 && extracted[0] && root != nil {
		// An extracted sibling has its own module-relative source locations;
		// its physical temporary path remains the module's SourcePath.
		rebaseSameFileSiblingModuleSyntax(root)
	}
	if diags.HasErrors() || root == nil {
		mod := &Module{SourcePath: file, Source: source}
		if root != nil {
			mod.Name = SanyModuleName(root)
			mod.Pos = sanyNodePosition(root)
		}
		return mod, dependencies, diags
	}
	mod, modDiags := sanyModuleFromSyntax(file, root)
	setModuleSourceRecursive(mod, source)
	diags = append(diags, modDiags...)
	return mod, dependencies, diags
}

func parseSanyModuleSources(file, source string) ([]*Module, Diagnostics) {
	roots, diags := ParseSanySyntaxModules(file, source)
	if len(roots) == 0 {
		return nil, diags
	}
	mods := make([]*Module, 0, len(roots))
	for i, root := range roots {
		if root == nil {
			continue
		}
		if i > 0 {
			rebaseSameFileSiblingModuleSyntax(root)
		}
		mod, modDiags := sanyModuleFromSyntax(file, root)
		setModuleSourceRecursive(mod, source)
		diags = append(diags, modDiags...)
		mods = append(mods, mod)
	}
	return mods, diags
}

func rebaseSameFileSiblingModuleSyntax(root *SanySyntaxNode) {
	if root == nil {
		return
	}
	moduleName := SanyModuleName(root)
	if moduleName == "" {
		return
	}
	lineOffset := root.Range.Begin.Line - 1
	rebaseSanySyntaxNodePositions(root, moduleName, lineOffset)
}

func rebaseSanySyntaxNodePositions(node *SanySyntaxNode, file string, lineOffset int) {
	if node == nil {
		return
	}
	node.FileName = file
	rebaseSanyRange(&node.Range, file, lineOffset)
	if node.Token != nil {
		rebaseSanyPosition(&node.Token.Begin, file, lineOffset)
		rebaseSanyPosition(&node.Token.End, file, lineOffset)
	}
	for _, child := range node.Zero {
		rebaseSanySyntaxNodePositions(child, file, lineOffset)
	}
	for _, child := range node.One {
		rebaseSanySyntaxNodePositions(child, file, lineOffset)
	}
}

func rebaseSanyRange(rng *SanyRange, file string, lineOffset int) {
	if rng == nil {
		return
	}
	rebaseSanyPosition(&rng.Begin, file, lineOffset)
	rebaseSanyPosition(&rng.End, file, lineOffset)
}

func rebaseSanyPosition(pos *Position, file string, lineOffset int) {
	if pos == nil {
		return
	}
	pos.File = file
	if pos.Line > 0 {
		pos.Line -= lineOffset
	}
	if pos.EndLine > 0 {
		pos.EndLine -= lineOffset
	}
}

func setModuleSourceRecursive(mod *Module, source string) {
	if mod == nil {
		return
	}
	mod.Source = source
	for _, nested := range mod.Nested {
		setModuleSourceRecursive(nested, source)
	}
}

func setModuleLibraryRecursive(mod *Module, library bool) {
	if mod == nil {
		return
	}
	mod.Library = library
	for _, nested := range mod.Nested {
		setModuleLibraryRecursive(nested, library)
	}
}

func sanyNodePosition(node *SanySyntaxNode) Position {
	if node == nil {
		return Position{}
	}
	pos := node.Range.Begin
	if node.Range.End.Line > 0 && node.Range.End.Column > 0 {
		pos.EndLine = node.Range.End.Line
		pos.EndColumn = node.Range.End.Column
	}
	return pos
}

func sanyModuleFromSyntax(file string, root *SanySyntaxNode) (*Module, Diagnostics) {
	mod := &Module{Name: SanyModuleName(root), SourcePath: file, Syntax: root}
	if root != nil {
		mod.Pos = sanyNodePosition(root)
	}
	var diags Diagnostics
	if root == nil {
		return mod, diags
	}
	heirs := root.GetHeirs()
	if len(heirs) >= 2 {
		for _, child := range heirs[1].GetHeirs() {
			if child != nil && child.Kind.JavaName() == "IDENTIFIER" {
				mod.Extends = append(mod.Extends, child.Image)
			}
		}
	}
	if len(heirs) < 3 {
		return mod, diags
	}
	for _, item := range heirs[2].GetHeirs() {
		switch item.Kind.JavaName() {
		case "N_VariableDeclaration":
			mod.Declarations = append(mod.Declarations, sanyDeclaration(item, VariableDecl))
		case "N_ParamDeclaration":
			mod.Declarations = append(mod.Declarations, sanyDeclaration(item, ConstantDecl))
		case "N_Recursive":
			mod.Recursives = append(mod.Recursives, sanyDeclaration(item, RecursiveDecl))
		case "N_OperatorDefinition":
			def, defDiags := sanyDefinition(item)
			diags = append(diags, defDiags...)
			mod.Definitions = append(mod.Definitions, def)
		case "N_FunctionDefinition":
			def, defDiags := sanyFunctionDefinition(item)
			diags = append(diags, defDiags...)
			mod.Definitions = append(mod.Definitions, def)
		case "N_ModuleDefinition":
			inst, instDiags := sanyModuleDefinition(item)
			diags = append(diags, instDiags...)
			if inst.Module != "" {
				mod.Instances = append(mod.Instances, inst)
			}
		case "N_Instance":
			inst, instDiags := sanyInstance(item)
			diags = append(diags, instDiags...)
			if inst.Module != "" {
				mod.Instances = append(mod.Instances, inst)
			}
		case "N_Module":
			nested, nestedDiags := sanyModuleFromSyntax(file, item)
			diags = append(diags, nestedDiags...)
			if nested != nil && nested.Name != "" {
				mod.Nested = append(mod.Nested, nested)
			}
		case "N_Assumption":
			var named NamedExpr
			if def, ok, defDiags := sanyAssumptionDefinition(item); ok {
				diags = append(diags, defDiags...)
				named = namedExprFromDefinition(def, item)
			}
			if named.Expr == nil {
				var expr Expr
				var exprDiags Diagnostics
				expr, named.AssumeProveBody, exprDiags = sanyExprWithAssumeProveBody(lastSanyExpression(item))
				diags = append(diags, exprDiags...)
				named.Expr = expr
				named.AssumeProve = named.AssumeProveBody != nil
				named.Pos = sanyNodePosition(item)
				named.Source = sanyNodePosition(item)
				named.Syntax = item
			}
			if named.Expr != nil {
				mod.Assumptions = append(mod.Assumptions, named)
			}
		case "N_Theorem":
			var named NamedExpr
			if def, ok, defDiags := sanyTheoremDefinition(item); ok {
				diags = append(diags, defDiags...)
				mod.Definitions = append(mod.Definitions, def)
				named = namedExprFromDefinition(def, item)
			}
			if named.Expr == nil {
				var expr Expr
				var exprDiags Diagnostics
				expr, named.AssumeProveBody, exprDiags = sanyExprWithAssumeProveBody(lastSanyExpression(item))
				diags = append(diags, exprDiags...)
				named.Expr = expr
				named.AssumeProve = named.AssumeProveBody != nil
				named.Pos = sanyNodePosition(item)
				named.Source = sanyNodePosition(item)
				named.Syntax = item
			}
			if named.Expr != nil {
				mod.Theorems = append(mod.Theorems, named)
				if proof := sanyTheoremProof(item, named.Expr); len(proof.Steps) > 0 || len(proof.LeafRefs) > 0 || sanyLeafProofSyntax(item) != nil {
					mod.Proofs = append(mod.Proofs, proof)
				}
			}
		case "N_UseOrHide":
			mod.ProofRefs = append(mod.ProofRefs, sanyUseOrHideRefs(item)...)
			mod.ProofRefNodes = append(mod.ProofRefNodes, item)
		}
	}
	return mod, diags
}

func namedExprFromDefinition(def Definition, syntax *SanySyntaxNode) NamedExpr {
	pos := sanyNodePosition(syntax)
	return NamedExpr{
		Name:            def.Name,
		Expr:            def.Expr,
		AssumeProve:     def.AssumeProve,
		AssumeProveBody: def.AssumeProveBody,
		PreComments:     append([]string(nil), def.PreComments...),
		Pos:             pos,
		Source:          pos,
		Syntax:          syntax,
	}
}

func sanyExprWithAssumeProveBody(node *SanySyntaxNode) (Expr, *AssumeProve, Diagnostics) {
	if node == nil || node.Kind.JavaName() != "N_AssumeProve" {
		expr, diags := sanyExpr(node)
		return expr, nil, diags
	}
	body, diags := sanyAssumeProveBody(node)
	if body == nil || body.Prove == nil {
		expr, exprDiags := unsupportedSanyExpr(node)
		diags = append(diags, exprDiags...)
		return expr, body, diags
	}
	return assumeProveExpr(body, sanyNodePosition(node)), body, diags
}

func moduleImports(mod *Module) []string {
	if mod == nil {
		return nil
	}
	imports := make([]string, 0, len(mod.Extends)+len(mod.Instances))
	seen := map[string]bool{}
	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		imports = append(imports, name)
	}
	for _, name := range mod.Extends {
		add(name)
	}
	for _, inst := range mod.Instances {
		add(inst.Module)
		moduleImportsFromInstance(inst, add)
	}
	for _, def := range mod.Definitions {
		moduleImportsFromExpr(def.Expr, add)
	}
	for _, assume := range mod.Assumptions {
		moduleImportsFromExpr(assume.Expr, add)
	}
	for _, theorem := range mod.Theorems {
		moduleImportsFromExpr(theorem.Expr, add)
	}
	for _, nested := range mod.Nested {
		for _, name := range moduleImports(nested) {
			add(name)
		}
	}
	return imports
}

func moduleSemanticImports(mod *Module) []string {
	if mod == nil {
		return nil
	}
	imports := make([]string, 0, len(mod.Extends)+len(mod.Instances))
	seen := map[string]bool{}
	add := func(name string) {
		if name == "" || seen[name] {
			return
		}
		seen[name] = true
		imports = append(imports, name)
	}
	for _, name := range mod.Extends {
		add(name)
	}
	for _, inst := range mod.Instances {
		moduleImportsFromInstance(inst, add)
	}
	for _, def := range mod.Definitions {
		moduleImportsFromExpr(def.Expr, add)
	}
	for _, assume := range mod.Assumptions {
		moduleImportsFromExpr(assume.Expr, add)
	}
	for _, theorem := range mod.Theorems {
		moduleImportsFromExpr(theorem.Expr, add)
	}
	for _, nested := range mod.Nested {
		for _, name := range moduleSemanticImports(nested) {
			add(name)
		}
	}
	return imports
}

func moduleImportsFromInstance(inst Instance, add func(string)) {
	add(inst.Module)
	for _, subst := range instanceSubstitutions(inst) {
		moduleImportsFromExpr(subst.Expr, add)
	}
}

func moduleImportsFromExpr(expr Expr, add func(string)) {
	if expr == nil {
		return
	}
	switch e := expr.(type) {
	case *IdentExpr, *LiteralExpr:
	case *UnaryExpr:
		moduleImportsFromExpr(e.Expr, add)
	case *BinaryExpr:
		moduleImportsFromExpr(e.Left, add)
		moduleImportsFromExpr(e.Right, add)
	case *CallExpr:
		moduleImportsFromExpr(e.Callee, add)
		for _, arg := range e.Args {
			moduleImportsFromExpr(arg, add)
		}
	case *IfExpr:
		moduleImportsFromExpr(e.Cond, add)
		moduleImportsFromExpr(e.Then, add)
		moduleImportsFromExpr(e.Else, add)
	case *LetExpr:
		for _, inst := range e.Instances {
			moduleImportsFromInstance(inst, add)
		}
		for _, def := range e.Definitions {
			moduleImportsFromExpr(def.Expr, add)
		}
		moduleImportsFromExpr(e.Body, add)
	case *QuantifierExpr:
		moduleImportsFromExpr(e.Set, add)
		moduleImportsFromExpr(e.Body, add)
	case *CaseExpr:
		for _, arm := range e.Arms {
			moduleImportsFromExpr(arm.Test, add)
			moduleImportsFromExpr(arm.Value, add)
		}
		moduleImportsFromExpr(e.Other, add)
	case *ChooseExpr:
		moduleImportsFromExpr(e.Set, add)
		moduleImportsFromExpr(e.Body, add)
	case *TupleExpr:
		for _, elem := range e.Elems {
			moduleImportsFromExpr(elem, add)
		}
	case *SetExpr:
		for _, elem := range e.Elems {
			moduleImportsFromExpr(elem, add)
		}
	case *RecordExpr:
		for _, field := range e.Fields {
			moduleImportsFromExpr(field.Value, add)
		}
	case *RecordComponentExpr:
		moduleImportsFromExpr(e.Record, add)
	case *RecordSetExpr:
		for _, field := range e.Fields {
			moduleImportsFromExpr(field.Set, add)
		}
	case *FunctionExpr:
		for _, bound := range e.Bounds {
			moduleImportsFromExpr(bound.Set, add)
		}
		moduleImportsFromExpr(e.Body, add)
	case *FunctionAppExpr:
		moduleImportsFromExpr(e.Function, add)
		for _, arg := range e.Args {
			moduleImportsFromExpr(arg, add)
		}
	case *ExceptExpr:
		moduleImportsFromExpr(e.Base, add)
		for _, spec := range e.Specs {
			for _, component := range spec.Components {
				for _, index := range component.Indices {
					moduleImportsFromExpr(index, add)
				}
			}
			moduleImportsFromExpr(spec.Value, add)
		}
	case *LabelExpr:
		moduleImportsFromExpr(e.Body, add)
	case *ActionExpr:
		moduleImportsFromExpr(e.Action, add)
		moduleImportsFromExpr(e.Subscript, add)
	case *FairnessExpr:
		moduleImportsFromExpr(e.Subscript, add)
		moduleImportsFromExpr(e.Action, add)
	case *FunctionSetExpr:
		moduleImportsFromExpr(e.Domain, add)
		moduleImportsFromExpr(e.Range, add)
	case *SetComprehensionExpr:
		for _, bound := range e.Bounds {
			moduleImportsFromExpr(bound.Set, add)
		}
		moduleImportsFromExpr(e.Element, add)
		moduleImportsFromExpr(e.Predicate, add)
	}
}

func sanyModuleDefinition(node *SanySyntaxNode) (Instance, Diagnostics) {
	inst, diags := sanyInstance(node)
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() != "N_IdentLHS" {
			continue
		}
		inst.LHSPos = sanyNodePosition(child)
		if id := firstSanyIdentifier(child); id != nil {
			inst.Name = id.Image
		}
		for _, lhsChild := range child.GetHeirs() {
			name, arity, pos := "", 0, Position{}
			switch lhsChild.Kind.JavaName() {
			case "N_IdentDecl":
				if id := firstSanyIdentifier(lhsChild); id != nil {
					name, pos = id.Image, sanyNodePosition(id)
					arity = countDirectSanyChildren(lhsChild, "US")
					if arity > 0 {
						pos = sanyNodePosition(lhsChild)
					}
				}
			case "N_PrefixDecl", "N_PostfixDecl":
				name, arity, pos = sanyFixDeclOperatorName(lhsChild), 1, sanyFixDeclOperatorPosition(lhsChild)
			case "N_InfixDecl":
				name, arity, pos = sanyFixDeclOperatorName(lhsChild), 2, sanyNodePosition(lhsChild)
			}
			if name != "" {
				inst.Params = append(inst.Params, name)
				if inst.ParamPositions == nil {
					inst.ParamPositions = map[string]Position{}
				}
				inst.ParamPositions[name] = pos
				if arity > 0 {
					if inst.ParamArities == nil {
						inst.ParamArities = map[string]int{}
					}
					inst.ParamArities[name] = arity
				}
			}
		}
		break
	}
	return inst, diags
}

func sanyInstance(node *SanySyntaxNode) (Instance, Diagnostics) {
	var search func(*SanySyntaxNode) *SanySyntaxNode
	search = func(cur *SanySyntaxNode) *SanySyntaxNode {
		if cur == nil {
			return nil
		}
		if cur.Kind.JavaName() == "N_NonLocalInstance" {
			return cur
		}
		for _, child := range cur.GetHeirs() {
			if found := search(child); found != nil {
				return found
			}
		}
		return nil
	}
	nonLocal := search(node)
	inst := Instance{Syntax: node, Pos: sanyNodePosition(node), Source: sanyNodePosition(node), Substitutions: map[string]Expr{}, Local: sanyHasLocalPrefix(node), PreComments: sanyLeadingPreComments(node)}
	if nonLocal == nil {
		return inst, nil
	}
	var diags Diagnostics
	heirs := nonLocal.GetHeirs()
	for i, child := range heirs {
		switch child.Kind.JavaName() {
		case "INSTANCE":
			if i+1 < len(heirs) && heirs[i+1].Kind.JavaName() == "IDENTIFIER" {
				inst.Module = heirs[i+1].Image
				inst.Pos = sanyNodePosition(heirs[i+1])
			}
		case "N_Substitution":
			subst, substDiags := sanySubstitution(child)
			diags = append(diags, substDiags...)
			if subst.Name != "" && subst.Expr != nil {
				inst.SubstitutionList = append(inst.SubstitutionList, subst)
				inst.Substitutions[subst.Name] = subst.Expr
			}
		}
	}
	return inst, diags
}

func sanySubstitution(node *SanySyntaxNode) (Substitution, Diagnostics) {
	subst := Substitution{Pos: sanyNodePosition(node)}
	heirs := node.GetHeirs()
	if len(heirs) > 0 {
		if heirs[0].Kind.JavaName() == "IDENTIFIER" {
			subst.Name = heirs[0].Image
			subst.Pos = sanyNodePosition(heirs[0])
		} else {
			subst.Name = sanyOperatorImage(heirs[0])
			subst.Pos = sanyNodePosition(heirs[0])
		}
	}
	expr, diags := sanyExpr(lastSanyExpression(node))
	subst.Expr = expr
	return subst, diags
}

func sanyUseOrHideRefs(node *SanySyntaxNode) []ProofRef {
	var refs []ProofRef
	mode := ""
	inDefs := false
	for _, child := range node.GetHeirs() {
		if child.Token != nil && (child.Token.Kind == SanyTokenUse || child.Token.Kind == SanyTokenHide || child.Token.Kind == SanyTokenBy) {
			mode = child.Image
			break
		}
	}
	heirs := node.GetHeirs()
	for i := 0; i < len(heirs); i++ {
		child := heirs[i]
		if child.Token != nil && child.Token.Kind == SanyTokenModule {
			i++
			if i < len(heirs) {
				refs = append(refs, ProofRef{Syntax: node, Module: heirs[i].Image, Mode: mode, Defs: inDefs, Pos: sanyNodePosition(heirs[i])})
			}
			continue
		}
		if child.Token != nil && child.Token.Kind == SanyTokenDF {
			inDefs = true
			continue
		}
		if isSanyExpressionNode(child) && child.Kind.JavaName() != "N_ModuleRef" && child.Kind.JavaName() != "N_ModuleDefinition" {
			expr, _ := sanyExpr(child)
			ref := ProofRef{Syntax: node, Expr: expr, Mode: mode, Defs: inDefs, Pos: sanyNodePosition(child)}
			if id, ok := expr.(*IdentExpr); ok {
				ref.Name = id.Name
			} else if literal, ok := expr.(*LiteralExpr); ok {
				ref.Name = literal.Value
			} else if call, ok := expr.(*CallExpr); ok {
				if id, ok := call.Callee.(*IdentExpr); ok {
					ref.Name = id.Name
				}
			} else if child.Kind.JavaName() == "IDENTIFIER" {
				ref.Name = child.Image
			}
			refs = append(refs, ref)
			continue
		}
		if child.Token != nil && isSanyProofStepStartKind(child.Token.Kind) {
			expr, _ := sanyExpr(child)
			refs = append(refs, ProofRef{Syntax: node, Expr: expr, Name: sanyXMLProofStepNameImage(child.Image), Mode: mode, Defs: inDefs, Pos: sanyNodePosition(child)})
			continue
		}
		switch child.Kind.JavaName() {
		case "N_ModuleRef", "N_ModuleDefinition":
			heirs := child.GetHeirs()
			if len(heirs) > 1 {
				refs = append(refs, ProofRef{Syntax: node, Module: sanyFirstTokenImage(heirs[1]), Mode: mode, Defs: inDefs, Pos: sanyNodePosition(heirs[1])})
			}
		case "IDENTIFIER":
			refs = append(refs, ProofRef{Syntax: node, Name: child.Image, Mode: mode, Defs: inDefs, Pos: sanyNodePosition(child)})
		case "N_GeneralId":
			if name := sanyGeneralIDName(child); name != "" {
				expr, _ := sanyExpr(child)
				refs = append(refs, ProofRef{Syntax: node, Name: name, Expr: expr, Mode: mode, Defs: inDefs, Pos: sanyNodePosition(child)})
			}
		}
	}
	return refs
}

func sanyTheoremProof(node *SanySyntaxNode, goal Expr) ProofSummary {
	proof := ProofSummary{Syntax: node, Goal: goal, Pos: sanyNodePosition(node), LeafRefs: sanyLeafProofReferences(node)}
	proof.Facts = sanyProofFactsFromRefs(proof.LeafRefs)
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() == "N_Proof" {
			proof.Steps = append(proof.Steps, sanyProofSteps(child)...)
		}
	}
	return proof
}

func sanyProofSteps(node *SanySyntaxNode) []ProofStep {
	return sanyProofStepsAt(node, 0)
}

func sanyProofStepsAt(node *SanySyntaxNode, depth int) []ProofStep {
	var steps []ProofStep
	if node == nil {
		return steps
	}
	if node.Kind.JavaName() == "N_ProofStep" {
		if step, ok := sanyProofStep(node); ok {
			step.Depth = depth
			steps = append(steps, step)
		}
		for _, child := range node.GetHeirs() {
			if child.Kind.JavaName() == "N_Proof" {
				steps = append(steps, sanyProofStepsAt(child, depth+1)...)
			}
		}
		return steps
	}
	for _, child := range node.GetHeirs() {
		steps = append(steps, sanyProofStepsAt(child, depth)...)
	}
	return steps
}

func sanyProofStep(node *SanySyntaxNode) (ProofStep, bool) {
	step := ProofStep{Syntax: node, Pos: sanyNodePosition(node), LeafRefs: sanyLeafProofReferences(node)}
	step.Facts = sanyProofFactsFromRefs(step.LeafRefs)
	for _, child := range node.GetHeirs() {
		if child.Token != nil && isSanyProofStepStartKind(child.Token.Kind) {
			step.Name = sanyProofStepName(child.Image)
			step.QualifiedName = sanyXMLProofStepNameImage(child.Image)
			step.Implicit = child.Original != "" && child.Original != child.Image &&
				(strings.HasPrefix(child.Original, "<+>") || strings.HasPrefix(child.Original, "<*>"))
			step.Pos = sanyNodePosition(child)
			break
		}
	}
	for _, child := range node.GetHeirs() {
		if strings.HasSuffix(child.Kind.JavaName(), "Step") {
			step.Statement = sanyNodePosition(child)
		}
		switch child.Kind.JavaName() {
		case "N_DefStep":
			step.Kind = "DEFINE"
			for _, definition := range child.GetHeirs() {
				switch definition.Kind.JavaName() {
				case "N_OperatorDefinition":
					def, _ := sanyDefinition(definition)
					step.Definitions = append(step.Definitions, def)
				case "N_FunctionDefinition":
					def, _ := sanyFunctionDefinition(definition)
					step.Definitions = append(step.Definitions, def)
				case "N_ModuleDefinition":
					instance, _ := sanyModuleDefinition(definition)
					step.Instances = append(step.Instances, instance)
				}
			}
		case "N_NonLocalInstance":
			step.Kind = "INSTANCE"
			instance, _ := sanyInstance(child)
			step.Instances = append(step.Instances, instance)
		case "N_UseOrHide":
			if heirs := child.GetHeirs(); len(heirs) > 0 {
				step.Kind = heirs[0].Image
			}
			step.UseHideRefs = sanyUseOrHideRefs(child)
			for _, ref := range step.UseHideRefs {
				if step.Kind == "" {
					step.Kind = ref.Mode
				}
				if !ref.Defs && ref.Expr != nil {
					_, direct := ref.Expr.(*IdentExpr)
					step.Facts = append(step.Facts, ProofFact{Expr: ref.Expr, Direct: direct})
				}
			}
		case "N_HaveStep":
			step.Kind = "HAVE"
			step.Expr, _ = sanyExpr(lastSanyExpression(child))
			step.Refs = append(step.Refs, sanyProofStepRefs(child)...)
		case "N_TakeStep":
			step.Kind = "TAKE"
			step.Bounds = sanyProofStepBounds(child)
			step.Refs = append(step.Refs, sanyProofStepRefs(child)...)
		case "N_WitnessStep":
			step.Kind = "WITNESS"
			step.Exprs = sanyProofStepExprs(child)
			step.Refs = append(step.Refs, sanyProofStepRefs(child)...)
		case "N_PickStep":
			step.Kind = "PICK"
			step.Bounds = sanyProofStepBounds(child)
			step.Expr, _ = sanyExpr(lastSanyExpression(child))
			step.Refs = append(step.Refs, sanyProofStepRefs(child)...)
		case "N_CaseStep":
			step.Kind = "CASE"
			step.Expr, _ = sanyExpr(lastSanyExpression(child))
			step.Refs = append(step.Refs, sanyProofStepRefs(child)...)
		case "N_AssertStep":
			step.Kind = "ASSERT"
			step.Expr, step.AssumeProveBody, _ = sanyExprWithAssumeProveBody(lastSanyExpression(child))
			for _, token := range child.GetHeirs() {
				if token.Token != nil && token.Token.Kind == SanyTokenSuffices {
					step.Suffices = true
				}
			}
			step.Refs = append(step.Refs, sanyProofStepRefs(child)...)
		case "N_QEDStep":
			step.Kind = "QED"
			step.Refs = append(step.Refs, sanyProofStepRefs(child)...)
		}
	}
	return step, step.Kind != ""
}

func sanyProofStepName(image string) string {
	end := strings.Index(image, ">")
	if end < 0 || end+1 >= len(image) {
		return ""
	}
	return strings.TrimRight(image[end+1:], ".")
}

func sanyProofStepBounds(node *SanySyntaxNode) []BoundVar {
	var bounds []BoundVar
	for _, child := range node.GetHeirs() {
		if child.Token != nil && child.Token.Kind == SanyTokenColon {
			break
		}
		switch child.Kind.JavaName() {
		case "N_QuantBound":
			bound, _ := sanyQuantBoundVars(child)
			bounds = append(bounds, bound...)
		case "IDENTIFIER":
			bounds = append(bounds, BoundVar{Name: child.Image, Pos: sanyNodePosition(child)})
		case "N_IdentDecl":
			if id := firstSanyIdentifier(child); id != nil {
				bounds = append(bounds, BoundVar{Name: id.Image, Pos: sanyNodePosition(id)})
			}
		}
	}
	return bounds
}

func sanyProofStepExprs(node *SanySyntaxNode) []Expr {
	var exprs []Expr
	for _, child := range expressionChildren(node) {
		expr, _ := sanyExpr(child)
		if expr != nil {
			exprs = append(exprs, expr)
		}
	}
	return exprs
}

func sanyProofStepRefs(node *SanySyntaxNode) []string {
	var refs []string
	if node == nil {
		return refs
	}
	if node.Token != nil && isSanyProofStepStartKind(node.Token.Kind) {
		if name := sanyProofStepName(node.Image); name != "" {
			refs = append(refs, name)
		}
	}
	for _, child := range node.GetHeirs() {
		refs = append(refs, sanyProofStepRefs(child)...)
	}
	return refs
}

func sanyDeclaration(node *SanySyntaxNode, kind DeclarationKind) Declaration {
	decl := Declaration{Syntax: node, Kind: kind, Pos: sanyNodePosition(node), Arities: map[string]int{}, NamePositions: map[string]Position{}, NamePreComments: map[string][]string{}}
	addName := func(name string, arity int, pos Position, comments []string) {
		decl.Names = append(decl.Names, name)
		decl.Arities[name] = arity
		decl.NamePositions[name] = pos
		if len(comments) > 0 {
			decl.NamePreComments[name] = append([]string(nil), comments...)
		}
	}
	for _, child := range node.GetHeirs() {
		switch child.Kind.JavaName() {
		case "IDENTIFIER":
			addName(child.Image, 0, sanyNodePosition(child), sanyLeadingPreComments(child))
		case "N_IdentDecl":
			if id := firstSanyIdentifier(child); id != nil {
				arity := countDirectSanyChildren(child, "US")
				pos := sanyNodePosition(id)
				if arity > 0 {
					pos = sanyNodePosition(child)
				}
				addName(id.Image, arity, pos, sanyLeadingPreComments(child))
			}
		case "N_PrefixDecl", "N_PostfixDecl":
			name := sanyFixDeclOperatorName(child)
			if name != "" {
				addName(name, 1, sanyFixDeclOperatorPosition(child), sanyLeadingPreComments(child))
			}
		case "N_InfixDecl":
			name := sanyFixDeclOperatorName(child)
			if name != "" {
				addName(name, 2, sanyFixDeclOperatorPosition(child), sanyLeadingPreComments(child))
			}
		}
	}
	return decl
}

func sanyFixDeclOperatorName(node *SanySyntaxNode) string {
	heirs := node.GetHeirs()
	switch node.Kind.JavaName() {
	case "N_PrefixDecl":
		if len(heirs) > 0 {
			return sanyOperatorImage(heirs[0])
		}
	case "N_InfixDecl", "N_PostfixDecl":
		if len(heirs) > 1 {
			return sanyOperatorImage(heirs[1])
		}
	}
	return sanyOperatorImage(node)
}

func sanyFixDeclOperatorPosition(node *SanySyntaxNode) Position {
	heirs := node.GetHeirs()
	switch node.Kind.JavaName() {
	case "N_PrefixDecl":
		if len(heirs) > 0 {
			return sanyNodePosition(heirs[0])
		}
	case "N_InfixDecl", "N_PostfixDecl":
		if len(heirs) > 1 {
			return sanyNodePosition(heirs[1])
		}
	}
	return sanyNodePosition(node)
}

func countDirectSanyChildren(node *SanySyntaxNode, kind string) int {
	if node == nil {
		return 0
	}
	count := 0
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() == kind {
			count++
		}
	}
	return count
}

func sanyDefinitionHeirs(node *SanySyntaxNode) []*SanySyntaxNode {
	if node == nil {
		return nil
	}
	if len(node.One) > 0 {
		return append([]*SanySyntaxNode(nil), node.One...)
	}
	return node.GetHeirs()
}

func sanyHasLocalPrefix(node *SanySyntaxNode) bool {
	if node == nil {
		return false
	}
	for _, child := range node.Zero {
		if child != nil && child.Token != nil && child.Token.Kind == SanyTokenLocal {
			return true
		}
	}
	return false
}

func sanyDefinition(node *SanySyntaxNode) (Definition, Diagnostics) {
	def := Definition{Syntax: node, Pos: sanyNodePosition(node), Source: sanyNodePosition(node), Local: sanyHasLocalPrefix(node), ParamPositions: map[string]Position{}, PreComments: sanyLeadingPreComments(node)}
	heirs := sanyDefinitionHeirs(node)
	if len(heirs) == 0 {
		return def, nil
	}
	lhs := heirs[0]
	def.LHSPos = sanyNodePosition(lhs)
	switch lhs.Kind.JavaName() {
	case "N_IdentLHS":
		lhsHeirs := lhs.GetHeirs()
		if len(lhsHeirs) > 0 && lhsHeirs[0].Kind.JavaName() == "IDENTIFIER" {
			def.Name = lhsHeirs[0].Image
			def.Pos = sanyNodePosition(lhsHeirs[0])
		}
		for _, child := range lhsHeirs[1:] {
			switch child.Kind.JavaName() {
			case "N_IdentDecl":
				if id := firstSanyIdentifier(child); id != nil {
					def.Params = append(def.Params, id.Image)
					if arity := countDirectSanyChildren(child, "US"); arity > 0 {
						if def.ParamArities == nil {
							def.ParamArities = map[string]int{}
						}
						def.ParamArities[id.Image] = arity
						def.ParamPositions[id.Image] = sanyNodePosition(child)
					} else {
						def.ParamPositions[id.Image] = sanyNodePosition(id)
					}
				}
			case "N_PrefixDecl", "N_PostfixDecl":
				if name := sanyFixDeclOperatorName(child); name != "" {
					def.Params = append(def.Params, name)
					if def.ParamArities == nil {
						def.ParamArities = map[string]int{}
					}
					def.ParamArities[name] = 1
					def.ParamPositions[name] = sanyFixDeclOperatorPosition(child)
				}
			case "N_InfixDecl":
				if name := sanyFixDeclOperatorName(child); name != "" {
					def.Params = append(def.Params, name)
					if def.ParamArities == nil {
						def.ParamArities = map[string]int{}
					}
					def.ParamArities[name] = 2
					def.ParamPositions[name] = sanyNodePosition(child)
				}
			}
		}
	case "N_InfixLHS":
		lhsHeirs := lhs.GetHeirs()
		if len(lhsHeirs) >= 3 {
			def.Name = sanyOperatorImage(lhsHeirs[1])
			def.Pos = sanyNodePosition(lhsHeirs[1])
			if lhsHeirs[0].Kind.JavaName() == "IDENTIFIER" {
				def.Params = append(def.Params, lhsHeirs[0].Image)
				def.ParamPositions[lhsHeirs[0].Image] = sanyNodePosition(lhsHeirs[0])
			}
			if lhsHeirs[2].Kind.JavaName() == "IDENTIFIER" {
				def.Params = append(def.Params, lhsHeirs[2].Image)
				def.ParamPositions[lhsHeirs[2].Image] = sanyNodePosition(lhsHeirs[2])
			}
		}
	case "N_PrefixLHS":
		lhsHeirs := lhs.GetHeirs()
		if len(lhsHeirs) >= 2 {
			def.Name = sanyOperatorImage(lhsHeirs[0])
			def.Pos = sanyNodePosition(lhsHeirs[0])
			if lhsHeirs[1].Kind.JavaName() == "IDENTIFIER" {
				def.Params = append(def.Params, lhsHeirs[1].Image)
				def.ParamPositions[lhsHeirs[1].Image] = sanyNodePosition(lhsHeirs[1])
			}
		}
	case "N_PostfixLHS":
		lhsHeirs := lhs.GetHeirs()
		if len(lhsHeirs) >= 2 {
			def.Name = sanyOperatorImage(lhsHeirs[1])
			def.Pos = sanyNodePosition(lhsHeirs[1])
			if lhsHeirs[0].Kind.JavaName() == "IDENTIFIER" {
				def.Params = append(def.Params, lhsHeirs[0].Image)
				def.ParamPositions[lhsHeirs[0].Image] = sanyNodePosition(lhsHeirs[0])
			}
		}
	}
	expr, diags := sanyExpr(lastSanyExpression(node))
	def.Expr = expr
	return def, diags
}

func sanyAssumptionDefinition(node *SanySyntaxNode) (Definition, bool, Diagnostics) {
	def, ok, diags := sanyNamedBodyDefinition(node)
	def.TheoremLike = ok
	if ok {
		def.FactKind = "assume"
	}
	return def, ok, diags
}

func sanyTheoremDefinition(node *SanySyntaxNode) (Definition, bool, Diagnostics) {
	def, ok, diags := sanyNamedBodyDefinition(node)
	def.TheoremLike = ok
	if ok {
		def.FactKind = "theorem"
		def.FactKeyword = "theorem"
		if heirs := node.GetHeirs(); len(heirs) > 0 && heirs[0].Token != nil && heirs[0].Token.Kind == SanyTokenProposition {
			def.FactKeyword = "lemma"
		}
	}
	return def, ok, diags
}

func sanyNamedBodyDefinition(node *SanySyntaxNode) (Definition, bool, Diagnostics) {
	heirs := node.GetHeirs()
	for i := 0; i+1 < len(heirs); i++ {
		if heirs[i].Kind.JavaName() != "IDENTIFIER" || heirs[i+1].Token == nil || heirs[i+1].Token.Kind != SanyTokenDef {
			continue
		}
		body := firstSanyBodyExpressionAfter(heirs, i+2)
		expr, diags := sanyExpr(body)
		var assumeProveBody *AssumeProve
		if body != nil && body.Kind.JavaName() == "N_AssumeProve" {
			ap, apDiags := sanyAssumeProveBody(body)
			diags = append(diags, apDiags...)
			assumeProveBody = ap
		}
		return Definition{
			Syntax:          node,
			Name:            heirs[i].Image,
			Expr:            expr,
			AssumeProve:     body != nil && body.Kind.JavaName() == "N_AssumeProve",
			AssumeProveBody: assumeProveBody,
			Pos:             sanyNodePosition(heirs[i]),
			Source:          sanyNodePosition(node),
			LHSPos:          sanyNodePosition(heirs[i]),
			PreComments:     sanyLeadingPreComments(node),
		}, true, diags
	}
	return Definition{}, false, nil
}

func firstSanyBodyExpressionAfter(heirs []*SanySyntaxNode, start int) *SanySyntaxNode {
	for i := start; i < len(heirs); i++ {
		if heirs[i].Kind.JavaName() == "N_Proof" || heirs[i].Kind.JavaName() == "N_TerminalProof" {
			return nil
		}
		if isSanyExpressionNode(heirs[i]) {
			return heirs[i]
		}
	}
	return nil
}

func sanyFunctionDefinition(node *SanySyntaxNode) (Definition, Diagnostics) {
	def := Definition{Syntax: node, Pos: sanyNodePosition(node), Source: sanyNodePosition(node), Local: sanyHasLocalPrefix(node), FunctionDef: true, ParamPositions: map[string]Position{}, PreComments: sanyLeadingPreComments(node)}
	if id := firstSanyIdentifier(node); id != nil {
		def.Name = id.Image
		def.Pos = sanyNodePosition(id)
		def.LHSPos = sanyNodePosition(node)
	}
	fcn := &FunctionExpr{Pos: sanyNodePosition(node)}
	var diags Diagnostics
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() != "N_QuantBound" {
			continue
		}
		bounds, boundDiags := sanyQuantBoundVars(child)
		diags = append(diags, boundDiags...)
		fcn.Bounds = append(fcn.Bounds, bounds...)
	}
	bodyNode := lastSanyExpression(node)
	if bodyNode != nil {
		fcn.Pos = sanyNodePosition(bodyNode)
	}
	body, bodyDiags := sanyExpr(bodyNode)
	diags = append(diags, bodyDiags...)
	fcn.Body = body
	def.Expr = fcn
	return def, diags
}

func sanyLeadingPreComments(node *SanySyntaxNode) []string {
	if node == nil {
		return nil
	}
	if len(node.PreComments) > 0 {
		return append([]string(nil), node.PreComments...)
	}
	for _, child := range node.GetHeirs() {
		if child == nil {
			continue
		}
		return sanyLeadingPreComments(child)
	}
	return nil
}

func sanyExpr(node *SanySyntaxNode) (expr Expr, diags Diagnostics) {
	defer func() {
		if node != nil && node.Kind.JavaName() != "N_ParenExpr" {
			if source, ok := expr.(interface{ SetSyntaxNode(*SanySyntaxNode) }); ok {
				source.SetSyntaxNode(node)
			}
		}
	}()
	return sanyExprImpl(node)
}

func sanyExprImpl(node *SanySyntaxNode) (Expr, Diagnostics) {
	if node == nil {
		return nil, nil
	}
	if node.Token != nil && isSanyProofStepStartKind(node.Token.Kind) {
		return &IdentExpr{Name: sanyXMLProofStepNameImage(node.Image), Pos: sanyNodePosition(node)}, nil
	}
	switch node.Kind.JavaName() {
	case "IDENTIFIER":
		if node.Image == "TRUE" || node.Image == "FALSE" {
			return &LiteralExpr{Kind: "bool", Value: node.Image, Pos: sanyNodePosition(node)}, nil
		}
		return &IdentExpr{Name: node.Image, Pos: sanyNodePosition(node)}, nil
	case "N_GeneralId":
		// A primitive identifier is wrapped in GeneralId by Java's parser.
		// Preserve literal translation for its unqualified built-in booleans.
		if heirs := node.GetHeirs(); len(heirs) == 2 && len(heirs[0].GetHeirs()) == 0 &&
			heirs[1].Kind.JavaName() == "IDENTIFIER" && (heirs[1].Image == "TRUE" || heirs[1].Image == "FALSE") {
			return sanyExpr(heirs[1])
		}
		if call, ok, diags := sanyGeneralIDCall(node); ok {
			return call, diags
		}
		return &IdentExpr{Name: sanyGeneralIDName(node), Pos: sanyNodePosition(node)}, nil
	case "N_GenInfixOp", "N_GenPrefixOp", "N_GenPostfixOp", "N_GenNonExpPrefixOp":
		return sanyOperatorReferenceExpr(node)
	case "N_Number":
		return &LiteralExpr{Kind: "number", Value: sanyFirstTokenImage(node), Pos: sanyNodePosition(node)}, nil
	case "N_Real":
		return &LiteralExpr{Kind: "number", Value: sanyRealLiteralImage(node), Pos: sanyNodePosition(node)}, nil
	case "N_String":
		// Generator's StringNode(treeNode, true) strips the retained quotes.
		value := node.Image[1 : len(node.Image)-1]
		return &LiteralExpr{Kind: "string", Value: value, Pos: sanyNodePosition(node)}, nil
	case "N_ParenExpr":
		return sanyExpr(firstSanyExpression(node))
	case "N_ConjList":
		return sanyJunctionListExpr(node, "/\\")
	case "N_DisjList":
		return sanyJunctionListExpr(node, "\\/")
	case "N_InfixExpr":
		heirs := node.GetHeirs()
		if len(heirs) < 3 {
			return unsupportedSanyExpr(node)
		}
		left, leftDiags := sanyExpr(heirs[0])
		right, rightDiags := sanyExpr(heirs[2])
		diags := append(leftDiags, rightDiags...)
		return &BinaryExpr{Op: sanyASTOperator(sanyOperatorImage(heirs[1])), Left: left, Right: right, Pos: sanyNodePosition(node), JunctionList: node.JunctionList}, diags
	case "N_Times":
		return sanyBinaryExpression(node, sanyNaryOperator(node, "\\times"))
	case "N_PrefixExpr":
		heirs := node.GetHeirs()
		if len(heirs) < 2 {
			return unsupportedSanyExpr(node)
		}
		expr, diags := sanyExpr(heirs[1])
		op := sanyASTOperator(sanyOperatorImage(heirs[0]))
		if op == "-" {
			op = "-."
		}
		return &UnaryExpr{Op: op, Expr: expr, Pos: sanyNodePosition(node)}, diags
	case "N_PostfixExpr":
		heirs := node.GetHeirs()
		if len(heirs) < 2 {
			return unsupportedSanyExpr(node)
		}
		expr, diags := sanyExpr(heirs[0])
		return &UnaryExpr{Op: sanyASTOperator(sanyOperatorImage(heirs[1])), Expr: expr, Pos: sanyNodePosition(node)}, diags
	case "N_IfThenElse":
		heirs := node.GetHeirs()
		if len(heirs) < 6 {
			return unsupportedSanyExpr(node)
		}
		cond, condDiags := sanyExpr(heirs[1])
		thenExpr, thenDiags := sanyExpr(heirs[3])
		elseExpr, elseDiags := sanyExpr(heirs[5])
		diags := append(append(condDiags, thenDiags...), elseDiags...)
		return &IfExpr{Cond: cond, Then: thenExpr, Else: elseExpr, Pos: sanyNodePosition(node)}, diags
	case "N_Tuple":
		tuple := &TupleExpr{Pos: sanyNodePosition(node)}
		var diags Diagnostics
		for _, child := range expressionChildren(node) {
			expr, exprDiags := sanyExpr(child)
			diags = append(diags, exprDiags...)
			tuple.Elems = append(tuple.Elems, expr)
		}
		return tuple, diags
	case "N_SetEnumerate":
		set := &SetExpr{Pos: sanyNodePosition(node)}
		var diags Diagnostics
		for _, child := range expressionChildren(node) {
			expr, exprDiags := sanyExpr(child)
			diags = append(diags, exprDiags...)
			set.Elems = append(set.Elems, expr)
		}
		return set, diags
	case "N_SubsetOf":
		return sanySubsetOf(node)
	case "N_SetOfAll":
		return sanySetOfAll(node)
	case "N_RcdConstructor":
		return sanyRecordConstructor(node)
	case "N_RecordComponent":
		return sanyRecordComponent(node)
	case "N_SetOfRcds":
		return sanyRecordSet(node)
	case "N_FcnConst":
		return sanyFunctionConstructor(node)
	case "N_FcnAppl":
		return sanyFunctionApplication(node)
	case "N_Except":
		return sanyExcept(node)
	case "N_Label":
		return sanyLabel(node)
	case "N_ActionExpr":
		return sanyAction(node)
	case "N_FairnessExpr":
		return sanyFairness(node)
	case "N_SetOfFcns":
		return sanyFunctionSet(node)
	case "N_OpApplication":
		heirs := node.GetHeirs()
		if len(heirs) < 2 {
			return unsupportedSanyExpr(node)
		}
		callee, calleeDiags := sanyExpr(heirs[0])
		call, flatten := callee.(*CallExpr)
		if flatten {
			call.Pos = sanyNodePosition(node)
		} else {
			call = &CallExpr{Callee: callee, Pos: sanyNodePosition(node)}
		}
		diags := calleeDiags
		for _, argNode := range expressionChildren(heirs[1]) {
			arg, argDiags := sanyExpr(argNode)
			diags = append(diags, argDiags...)
			call.Args = append(call.Args, arg)
		}
		return call, diags
	case "N_BoundQuant":
		return sanyBoundQuantifier(node)
	case "N_UnboundQuant":
		return sanyUnboundQuantifier(node)
	case "N_LetIn":
		return sanyLetIn(node)
	case "N_Case":
		return sanyCase(node)
	case "N_UnboundOrBoundChoose":
		return sanyChoose(node)
	case "N_Lambda":
		return sanyLambda(node)
	case "N_AssumeProve":
		return sanyAssumeProve(node)
	default:
		if node.Token != nil && node.Kind.JavaName() == "NUMBER_LITERAL" {
			return &LiteralExpr{Kind: "number", Value: node.Image, Pos: sanyNodePosition(node)}, nil
		}
		return unsupportedSanyExpr(node)
	}
}

func sanyJunctionListExpr(node *SanySyntaxNode, op string) (Expr, Diagnostics) {
	items := node.GetHeirs()
	if len(items) == 0 {
		return unsupportedSanyExpr(node)
	}
	left, diags := sanyExpr(sanyJunctionItemExpression(items[0]))
	if len(items) == 1 {
		return &UnaryExpr{Op: op, Expr: left, Pos: sanyNodePosition(node)}, diags
	}
	for _, item := range items[1:] {
		right, rightDiags := sanyExpr(sanyJunctionItemExpression(item))
		diags = append(diags, rightDiags...)
		left = &BinaryExpr{Op: op, Left: left, Right: right, Pos: sanyNodePosition(node), JunctionList: true}
	}
	return left, diags
}

func sanyJunctionItemExpression(item *SanySyntaxNode) *SanySyntaxNode {
	heirs := item.GetHeirs()
	if len(heirs) == 0 {
		return nil
	}
	return heirs[len(heirs)-1]
}

func sanyOperatorReferenceExpr(node *SanySyntaxNode) (Expr, Diagnostics) {
	name := sanyASTOperator(sanyOperatorImage(node))
	ident := &IdentExpr{Name: name, Pos: sanyNodePosition(node)}
	var opArgs *SanySyntaxNode
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() == "N_OpArgs" {
			opArgs = child
			break
		}
	}
	if opArgs == nil {
		return ident, nil
	}
	call := &CallExpr{Callee: ident, Pos: sanyNodePosition(node)}
	var diags Diagnostics
	for _, argNode := range expressionChildren(opArgs) {
		arg, argDiags := sanyExpr(argNode)
		diags = append(diags, argDiags...)
		call.Args = append(call.Args, arg)
	}
	return call, diags
}

func sanyBinaryExpression(node *SanySyntaxNode, op string) (Expr, Diagnostics) {
	heirs := node.GetHeirs()
	if len(heirs) == 3 {
		left, leftDiags := sanyExpr(heirs[0])
		right, rightDiags := sanyExpr(heirs[2])
		diags := append(leftDiags, rightDiags...)
		return &BinaryExpr{Op: op, Left: left, Right: right, Pos: sanyNodePosition(node)}, diags
	}
	exprs := sanyBinaryOperandChildren(node, op)
	if len(exprs) < 2 {
		return unsupportedSanyExpr(node)
	}
	left, diags := sanyExpr(exprs[0])
	for _, exprNode := range exprs[1:] {
		right, rightDiags := sanyExpr(exprNode)
		diags = append(diags, rightDiags...)
		left = &BinaryExpr{Op: op, Left: left, Right: right, Pos: sanyNodePosition(node), SanyNary: true}
	}
	return left, diags
}

func sanyNaryOperator(node *SanySyntaxNode, fallback string) string {
	for _, child := range node.GetHeirs() {
		if child != nil && child.Kind.JavaName() == "N_GenInfixOp" {
			if op := sanyASTOperator(sanyOperatorImage(child)); op != "" {
				return op
			}
		}
	}
	return fallback
}

func sanyBinaryOperandChildren(node *SanySyntaxNode, op string) []*SanySyntaxNode {
	var out []*SanySyntaxNode
	for _, child := range node.GetHeirs() {
		if child.Token != nil && isSanyProofStepStartKind(child.Token.Kind) {
			out = append(out, child)
			continue
		}
		if !isSanyExpressionNode(child) {
			continue
		}
		if child.Kind.JavaName() == "N_GenInfixOp" && sanyASTOperator(sanyOperatorImage(child)) == op {
			continue
		}
		out = append(out, child)
	}
	return out
}

func sanySubsetOf(node *SanySyntaxNode) (Expr, Diagnostics) {
	exprs := expressionChildren(node)
	if len(exprs) < 3 {
		return unsupportedSanyExpr(node)
	}
	intro := firstSanyQuantBoundIntro(node)
	introVars := sanyBoundIntroVars(intro)
	if len(introVars) == 0 {
		return unsupportedSanyExpr(node)
	}
	set, setDiags := sanyExpr(exprs[1])
	pred, predDiags := sanyExpr(exprs[2])
	diags := append(setDiags, predDiags...)
	bounds := make([]BoundVar, 0, len(introVars))
	for _, bound := range introVars {
		bound.Set = set
		bounds = append(bounds, bound)
	}
	return &SetComprehensionExpr{
		Element:   sanyBoundElement(bounds, sanyNodePosition(node)),
		Bounds:    bounds,
		Predicate: pred,
		Pos:       sanyNodePosition(node),
	}, diags
}

func sanySetOfAll(node *SanySyntaxNode) (Expr, Diagnostics) {
	set := &SetComprehensionExpr{Pos: sanyNodePosition(node)}
	var elemNode *SanySyntaxNode
	var diags Diagnostics
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() == "N_QuantBound" {
			bounds, boundDiags := sanyQuantBoundVars(child)
			diags = append(diags, boundDiags...)
			set.Bounds = append(set.Bounds, bounds...)
			continue
		}
		if elemNode == nil && isSanyExpressionNode(child) {
			elemNode = child
		}
	}
	elem, elemDiags := sanyExpr(elemNode)
	diags = append(diags, elemDiags...)
	set.Element = elem
	return set, diags
}

func sanyRecordConstructor(node *SanySyntaxNode) (Expr, Diagnostics) {
	record := &RecordExpr{Pos: sanyNodePosition(node)}
	var diags Diagnostics
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() != "N_FieldVal" {
			continue
		}
		field := RecordField{Pos: sanyNodePosition(child), Source: sanyNodePosition(child)}
		if id := firstSanyIdentifier(child); id != nil {
			field.Name = id.Image
			field.Pos = sanyNodePosition(id)
		}
		value, valueDiags := sanyExpr(lastSanyExpression(child))
		diags = append(diags, valueDiags...)
		field.Value = value
		record.Fields = append(record.Fields, field)
	}
	return record, diags
}

func sanyRecordComponent(node *SanySyntaxNode) (Expr, Diagnostics) {
	heirs := node.GetHeirs()
	if len(heirs) < 3 {
		return unsupportedSanyExpr(node)
	}
	record, diags := sanyExpr(heirs[0])
	field := ""
	fieldPos := sanyNodePosition(heirs[2])
	if id := firstSanyIdentifier(heirs[2]); id != nil {
		field = id.Image
		fieldPos = sanyNodePosition(id)
	} else if heirs[2] != nil {
		field = heirs[2].Image
	}
	return &RecordComponentExpr{Record: record, Field: field, FieldPos: fieldPos, Pos: sanyNodePosition(node)}, diags
}

func sanyRecordSet(node *SanySyntaxNode) (Expr, Diagnostics) {
	record := &RecordSetExpr{Pos: sanyNodePosition(node)}
	var diags Diagnostics
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() != "N_FieldSet" {
			continue
		}
		field := RecordSetField{Pos: sanyNodePosition(child), Source: sanyNodePosition(child)}
		if id := firstSanyIdentifier(child); id != nil {
			field.Name = id.Image
			field.Pos = sanyNodePosition(id)
		}
		set, setDiags := sanyExpr(lastSanyExpression(child))
		diags = append(diags, setDiags...)
		field.Set = set
		record.Fields = append(record.Fields, field)
	}
	return record, diags
}

func sanyFunctionConstructor(node *SanySyntaxNode) (Expr, Diagnostics) {
	fcn := &FunctionExpr{Pos: sanyNodePosition(node)}
	var diags Diagnostics
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() != "N_QuantBound" {
			continue
		}
		bounds, boundDiags := sanyQuantBoundVars(child)
		diags = append(diags, boundDiags...)
		fcn.Bounds = append(fcn.Bounds, bounds...)
	}
	body, bodyDiags := sanyExpr(lastSanyExpression(node))
	diags = append(diags, bodyDiags...)
	fcn.Body = body
	return fcn, diags
}

func sanyFunctionApplication(node *SanySyntaxNode) (Expr, Diagnostics) {
	exprs := expressionChildren(node)
	if len(exprs) == 0 {
		return unsupportedSanyExpr(node)
	}
	fn, fnDiags := sanyExpr(exprs[0])
	app := &FunctionAppExpr{Function: fn, Pos: sanyNodePosition(node)}
	diags := fnDiags
	argNodes := exprs[1:]
	for _, argNode := range argNodes {
		arg, argDiags := sanyExpr(argNode)
		diags = append(diags, argDiags...)
		app.Args = append(app.Args, arg)
	}
	return app, diags
}

func sanyExcept(node *SanySyntaxNode) (Expr, Diagnostics) {
	except := &ExceptExpr{Pos: sanyNodePosition(node)}
	var diags Diagnostics
	var baseNode *SanySyntaxNode
	for _, child := range node.GetHeirs() {
		switch child.Kind.JavaName() {
		case "N_ExceptSpec":
			spec, specDiags := sanyExceptSpec(child)
			diags = append(diags, specDiags...)
			except.Specs = append(except.Specs, spec)
		default:
			if baseNode == nil && isSanyExpressionNode(child) {
				baseNode = child
			}
		}
	}
	base, baseDiags := sanyExpr(baseNode)
	diags = append(baseDiags, diags...)
	except.Base = base
	return except, diags
}

func sanyExceptSpec(node *SanySyntaxNode) (ExceptSpec, Diagnostics) {
	spec := ExceptSpec{Pos: sanyNodePosition(node)}
	var diags Diagnostics
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() != "N_ExceptComponent" {
			continue
		}
		component, componentDiags := sanyExceptComponent(child)
		diags = append(diags, componentDiags...)
		spec.Components = append(spec.Components, component)
	}
	value, valueDiags := sanyExpr(lastSanyExpression(node))
	diags = append(diags, valueDiags...)
	spec.Value = value
	return spec, diags
}

func sanyExceptComponent(node *SanySyntaxNode) (ExceptComponent, Diagnostics) {
	component := ExceptComponent{Pos: sanyNodePosition(node)}
	heirs := node.GetHeirs()
	if len(heirs) > 0 && heirs[0].Token != nil && heirs[0].Token.Kind == SanyTokenDot {
		if id := firstSanyIdentifier(node); id != nil {
			component.Field = id.Image
			component.FieldPos = sanyNodePosition(id)
		}
		return component, nil
	}
	var diags Diagnostics
	for _, child := range expressionChildren(node) {
		index, indexDiags := sanyExpr(child)
		diags = append(diags, indexDiags...)
		component.Indices = append(component.Indices, index)
	}
	return component, diags
}

func sanyLabel(node *SanySyntaxNode) (Expr, Diagnostics) {
	label := &LabelExpr{Pos: sanyNodePosition(node)}
	heirs := node.GetHeirs()
	if len(heirs) > 0 {
		name := heirs[0]
		if name.Kind.JavaName() == "N_OpApplication" {
			application := name.GetHeirs()
			label.Params = sanyLabelArgumentNames(name)
			name = application[0]
		}
		if name.Kind.JavaName() == "N_GeneralId" {
			label.Name = sanyFirstTokenImage(name.GetHeirs()[1])
		}
	}
	body, diags := sanyExpr(lastSanyExpression(node))
	label.Body = body
	return label, diags
}

func sanyLabelArgumentNames(application *SanySyntaxNode) []string {
	var params []string
	args := application.GetHeirs()[1].GetHeirs()
	for i := 1; i < len(args); i += 2 {
		params = append(params, sanyFirstTokenImage(args[i].GetHeirs()[1]))
	}
	return params
}

func sanyAction(node *SanySyntaxNode) (Expr, Diagnostics) {
	exprs := expressionChildren(node)
	if len(exprs) < 2 {
		return unsupportedSanyExpr(node)
	}
	kind := "square"
	heirs := node.GetHeirs()
	if len(heirs) > 0 && heirs[0].Token != nil && heirs[0].Token.Kind == SanyTokenLab {
		kind = "angle"
	}
	action, actionDiags := sanyExpr(exprs[0])
	subscript, subscriptDiags := sanyExpr(exprs[1])
	diags := append(actionDiags, subscriptDiags...)
	return &ActionExpr{Kind: kind, Action: action, Subscript: subscript, Pos: sanyNodePosition(node)}, diags
}

func sanyFairness(node *SanySyntaxNode) (Expr, Diagnostics) {
	heirs := node.GetHeirs()
	if len(heirs) != 5 {
		return unsupportedSanyExpr(node)
	}
	kind := "WF"
	if heirs[0].Token != nil && heirs[0].Token.Kind == SanyTokenSF {
		kind = "SF"
	}
	subscript, subscriptDiags := sanyExpr(heirs[1])
	action, actionDiags := sanyExpr(heirs[3])
	diags := append(subscriptDiags, actionDiags...)
	return &FairnessExpr{Kind: kind, Subscript: subscript, Action: action, Pos: sanyNodePosition(node)}, diags
}

func sanyFunctionSet(node *SanySyntaxNode) (Expr, Diagnostics) {
	exprs := expressionChildren(node)
	if len(exprs) < 2 {
		return unsupportedSanyExpr(node)
	}
	domain, domainDiags := sanyExpr(exprs[0])
	rangeExpr, rangeDiags := sanyExpr(exprs[1])
	diags := append(domainDiags, rangeDiags...)
	return &FunctionSetExpr{Domain: domain, Range: rangeExpr, Pos: sanyNodePosition(node)}, diags
}

func sanyQuantBoundVars(node *SanySyntaxNode) ([]BoundVar, Diagnostics) {
	set, setDiags := sanyExpr(lastSanyExpression(node))
	var vars []BoundVar
	for _, child := range node.GetHeirs() {
		if child.Token != nil && child.Token.Kind == SanyTokenIN {
			break
		}
		for _, bound := range sanyBoundIntroVars(child) {
			bound.Set = set
			vars = append(vars, bound)
		}
	}
	return vars, setDiags
}

func firstSanyQuantBoundIntro(node *SanySyntaxNode) *SanySyntaxNode {
	for _, child := range node.GetHeirs() {
		if child.Token != nil && child.Token.Kind == SanyTokenIN {
			return nil
		}
		switch child.Kind.JavaName() {
		case "IDENTIFIER", "N_IdentifierTuple":
			return child
		}
	}
	return nil
}

func sanyBoundIntroVars(node *SanySyntaxNode) []BoundVar {
	if node == nil {
		return nil
	}
	switch node.Kind.JavaName() {
	case "IDENTIFIER":
		return []BoundVar{{Name: node.Image, Pos: sanyNodePosition(node)}}
	case "N_IdentifierTuple":
		// Source processQuantBoundArgs includes delimiter heirs in its count.
		heirs := node.GetHeirs()
		vars := make([]BoundVar, len(heirs)/2)
		for i := range vars {
			identifier := heirs[2*i+1]
			vars[i] = BoundVar{Name: identifier.Image, Pos: sanyNodePosition(identifier), TupleBound: true}
		}
		return vars
	default:
		return nil
	}
}

func sanyBoundElement(bounds []BoundVar, pos Position) Expr {
	if len(bounds) == 1 {
		return &IdentExpr{Name: bounds[0].Name, Pos: bounds[0].Pos}
	}
	tuple := &TupleExpr{Pos: pos}
	for _, bound := range bounds {
		tuple.Elems = append(tuple.Elems, &IdentExpr{Name: bound.Name, Pos: bound.Pos})
	}
	return tuple
}

func directSanyIdentifiers(node *SanySyntaxNode) []*SanySyntaxNode {
	var ids []*SanySyntaxNode
	if node == nil {
		return ids
	}
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() == "IDENTIFIER" {
			ids = append(ids, child)
		}
	}
	return ids
}

func sanyBoundQuantifier(node *SanySyntaxNode) (Expr, Diagnostics) {
	heirs := node.GetHeirs()
	if len(heirs) < 4 {
		return unsupportedSanyExpr(node)
	}
	var bounds []BoundVar
	var diags Diagnostics
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() != "N_QuantBound" {
			continue
		}
		boundVars, boundDiags := sanyQuantBoundVars(child)
		diags = append(diags, boundDiags...)
		bounds = append(bounds, boundVars...)
	}
	if len(bounds) == 0 {
		return unsupportedSanyExpr(node)
	}
	body, bodyDiags := sanyExpr(heirs[len(heirs)-1])
	diags = append(diags, bodyDiags...)
	return wrapQuantifierExprs(sanyQuantifierKind(heirs[0]), bounds, body, sanyNodePosition(node), node), diags
}

func sanyUnboundQuantifier(node *SanySyntaxNode) (Expr, Diagnostics) {
	heirs := node.GetHeirs()
	if len(heirs) < 4 {
		return unsupportedSanyExpr(node)
	}
	body, diags := sanyExpr(heirs[len(heirs)-1])
	var vars []BoundVar
	for _, child := range heirs[1 : len(heirs)-1] {
		if child.Kind.JavaName() != "IDENTIFIER" {
			continue
		}
		vars = append(vars, BoundVar{Name: child.Image, Pos: sanyNodePosition(child)})
	}
	if len(vars) == 0 {
		return unsupportedSanyExpr(node)
	}
	return wrapQuantifierExprs(sanyQuantifierKind(heirs[0]), vars, body, sanyNodePosition(node), node), diags
}

// Generator chooses quantifier semantics by token kind, including aliases.
func sanyQuantifierKind(node *SanySyntaxNode) string {
	switch node.Kind.JavaName() {
	case "EXISTS":
		return "\\E"
	case "FORALL":
		return "\\A"
	case "T_EXISTS":
		return "\\EE"
	case "T_FORALL":
		return "\\AA"
	default:
		return sanyOperatorImage(node)
	}
}

func wrapQuantifierExprs(kind string, vars []BoundVar, body Expr, pos Position, syntax *SanySyntaxNode) Expr {
	out := body
	for i := len(vars) - 1; i >= 0; i-- {
		out = &QuantifierExpr{
			SanyExprSource:   SanyExprSource{Syntax: syntax},
			Kind:             kind,
			Var:              vars[i].Name,
			VarPos:           vars[i].Pos,
			Set:              vars[i].Set,
			Body:             out,
			OperatorArity:    vars[i].OperatorArity,
			HasOperatorArity: vars[i].HasOperatorArity,
			TupleBound:       vars[i].TupleBound,
			LevelKnown:       vars[i].LevelKnown,
			Level:            vars[i].Level,
			Pos:              pos,
		}
	}
	return out
}

func sanyLetIn(node *SanySyntaxNode) (Expr, Diagnostics) {
	heirs := node.GetHeirs()
	if len(heirs) < 4 {
		return unsupportedSanyExpr(node)
	}
	let := &LetExpr{SanyExprSource: SanyExprSource{Syntax: node}, Pos: sanyNodePosition(node)}
	var diags Diagnostics
	for _, child := range heirs[1].GetHeirs() {
		switch child.Kind.JavaName() {
		case "N_Recursive":
			let.Recursives = append(let.Recursives, sanyDeclaration(child, RecursiveDecl))
		case "N_OperatorDefinition":
			def, defDiags := sanyDefinition(child)
			diags = append(diags, defDiags...)
			let.Definitions = append(let.Definitions, def)
		case "N_FunctionDefinition":
			def, defDiags := sanyFunctionDefinition(child)
			diags = append(diags, defDiags...)
			let.Definitions = append(let.Definitions, def)
		case "N_ModuleDefinition":
			inst, instDiags := sanyModuleDefinition(child)
			diags = append(diags, instDiags...)
			if inst.Name != "" && inst.Module != "" {
				let.Instances = append(let.Instances, inst)
			}
		}
	}
	body, bodyDiags := sanyExpr(heirs[3])
	diags = append(diags, bodyDiags...)
	let.Body = body
	return let, diags
}

func sanyCase(node *SanySyntaxNode) (Expr, Diagnostics) {
	caseExpr := &CaseExpr{Pos: sanyNodePosition(node)}
	var diags Diagnostics
	for _, child := range node.GetHeirs() {
		switch child.Kind.JavaName() {
		case "N_CaseArm":
			arm, armDiags := sanyCaseArm(child)
			diags = append(diags, armDiags...)
			caseExpr.Arms = append(caseExpr.Arms, arm)
		case "N_OtherArm":
			other, otherDiags := sanyExpr(lastSanyExpression(child))
			diags = append(diags, otherDiags...)
			caseExpr.Other = other
			caseExpr.OtherPos = sanyNodePosition(child)
		}
	}
	return caseExpr, diags
}

func sanyCaseArm(node *SanySyntaxNode) (CaseArm, Diagnostics) {
	exprs := expressionChildren(node)
	if len(exprs) < 2 {
		_, diags := unsupportedSanyExpr(node)
		return CaseArm{Pos: sanyNodePosition(node)}, diags
	}
	test, testDiags := sanyExpr(exprs[0])
	value, valueDiags := sanyExpr(exprs[1])
	diags := append(testDiags, valueDiags...)
	return CaseArm{Test: test, Value: value, Pos: sanyNodePosition(node)}, diags
}

func sanyChoose(node *SanySyntaxNode) (Expr, Diagnostics) {
	choose := &ChooseExpr{Pos: sanyNodePosition(node)}
	if tuple := firstSanyChildKind(node, "N_IdentifierTuple"); tuple != nil {
		heirs := tuple.GetHeirs()
		// Source processChoose sizes formals from all tuple heirs, including
		// delimiters. For <<>>, this retains the closing token as a formal.
		choose.TupleVars = make([]BoundVar, len(heirs)/2)
		for i := range choose.TupleVars {
			identifier := heirs[2*i+1]
			choose.TupleVars[i] = BoundVar{Name: identifier.Image, Pos: sanyNodePosition(identifier), TupleBound: true}
		}
	} else if id := firstSanyIdentifier(node); id != nil {
		choose.Var, choose.VarPos = id.Image, sanyNodePosition(id)
	} else {
		return unsupportedSanyExpr(node)
	}
	var diags Diagnostics
	if maybe := firstSanyChildKind(node, "N_MaybeBound"); maybe != nil {
		if setNode := lastSanyExpression(maybe); setNode != nil {
			set, setDiags := sanyExpr(setNode)
			diags = append(diags, setDiags...)
			choose.Set = set
		}
	}
	if choose.Set == nil && choose.TupleVars == nil {
		if chooseToken := firstSanyChildKind(node, "CHOOSE"); chooseToken != nil {
			choose.VarPos = sanyNodePosition(chooseToken)
		}
	}
	body, bodyDiags := sanyExpr(lastSanyExpression(node))
	diags = append(diags, bodyDiags...)
	choose.Body = body
	return choose, diags
}

func sanyLambda(node *SanySyntaxNode) (Expr, Diagnostics) {
	lambda := &FunctionExpr{Pos: sanyNodePosition(node), IsLambda: true, PreComments: sanyLeadingPreComments(node)}
	heirs := node.GetHeirs()
	// Lambda's source heirs are LAMBDA, identifier (, identifier)*, ':', body.
	for i := 1; i < len(heirs)-2; i += 2 {
		identifier := heirs[i]
		lambda.Bounds = append(lambda.Bounds, BoundVar{Name: identifier.Image, Pos: sanyNodePosition(identifier)})
	}
	body, diags := sanyExpr(lastSanyExpression(node))
	lambda.Body = body
	return lambda, diags
}

func sanyAssumeProve(node *SanySyntaxNode) (Expr, Diagnostics) {
	body, diags := sanyAssumeProveBody(node)
	if body == nil || body.Prove == nil {
		return unsupportedSanyExpr(node)
	}
	return assumeProveExpr(body, sanyNodePosition(node)), diags
}

func assumeProveExpr(body *AssumeProve, pos Position) Expr {
	var assumptions []Expr
	var newBounds []BoundVar
	for _, item := range body.Assumptions {
		if item.NewSymbol != nil {
			bound := BoundVar{Name: item.NewSymbol.Name, Set: item.NewSymbol.Domain, Pos: item.NewSymbol.Pos, LevelKnown: true, Level: int(item.NewSymbol.Level)}
			if item.NewSymbol.Arity > 0 {
				bound.OperatorArity = item.NewSymbol.Arity
				bound.HasOperatorArity = true
			}
			newBounds = append(newBounds, bound)
			continue
		}
		if item.Nested != nil {
			assumptions = append(assumptions, assumeProveExpr(item.Nested, item.Nested.Pos))
			continue
		}
		if item.Expr != nil {
			assumptions = append(assumptions, item.Expr)
		}
	}
	expr := body.Prove
	if len(assumptions) > 0 {
		assumption := assumptions[0]
		for _, next := range assumptions[1:] {
			assumption = &BinaryExpr{Op: "/\\", Left: assumption, Right: next, Pos: pos}
		}
		expr = &BinaryExpr{Op: "=>", Left: assumption, Right: body.Prove, Pos: pos}
	}
	return wrapQuantifierExprs("\\A", newBounds, expr, pos, nil)
}

func sanyAssumeProveBody(node *SanySyntaxNode) (*AssumeProve, Diagnostics) {
	body := &AssumeProve{Pos: sanyNodePosition(node)}
	var diags Diagnostics
	inProve := false
	for _, child := range node.GetHeirs() {
		if isSanyAssumeProveProveToken(child) {
			inProve = true
			continue
		}
		if child.Kind.JavaName() == "N_NewSymb" {
			sym, symDiags, ok := sanyNewSymbol(child)
			diags = append(diags, symDiags...)
			if ok {
				body.Assumptions = append(body.Assumptions, AssumeProveItem{NewSymbol: &sym})
			}
			continue
		}
		if child.Kind.JavaName() == "N_AssumeProve" {
			nested, nestedDiags := sanyAssumeProveBody(child)
			diags = append(diags, nestedDiags...)
			if inProve {
				body.Prove = assumeProveExpr(nested, nested.Pos)
			} else {
				body.Assumptions = append(body.Assumptions, AssumeProveItem{Nested: nested})
			}
			continue
		}
		if !isSanyExpressionNode(child) {
			continue
		}
		expr, exprDiags := sanyExpr(child)
		diags = append(diags, exprDiags...)
		if inProve {
			body.Prove = expr
		} else {
			body.Assumptions = append(body.Assumptions, AssumeProveItem{Expr: expr})
		}
	}
	return body, diags
}

func sanyNewSymbol(node *SanySyntaxNode) (NewSymbol, Diagnostics, bool) {
	id := sanyNewSymbIdentifier(node)
	sym := NewSymbol{
		Kind:   24,
		Level:  constantLevel,
		Pos:    sanyNodePosition(node),
		Source: sanyNodePosition(node),
	}
	if id != nil {
		sym.Name = id.Image
		sym.Pos = sanyNodePosition(id)
	}
	for _, child := range node.GetHeirs() {
		if child.Token == nil {
			continue
		}
		switch child.Token.Kind {
		case SanyTokenVariable:
			sym.Kind = 25
			sym.Level = variableLevel
		case SanyTokenState:
			sym.Kind = 26
			sym.Level = variableLevel
		case SanyTokenAction:
			sym.Kind = 27
			sym.Level = actionLevel
		case SanyTokenTemporal:
			sym.Kind = 28
			sym.Level = temporalLevel
		case SanyTokenConstant:
			sym.Kind = 24
			sym.Level = constantLevel
		}
	}
	for _, child := range node.GetHeirs() {
		switch child.Kind.JavaName() {
		case "N_IdentDecl":
			if id := firstSanyIdentifier(child); id != nil {
				sym.Name = id.Image
				sym.Arity = countDirectSanyChildren(child, "US")
				if sym.Arity > 0 {
					sym.Pos = sanyNodePosition(child)
				} else {
					sym.Pos = sanyNodePosition(id)
				}
			}
		case "N_PrefixDecl", "N_PostfixDecl":
			if name := sanyFixDeclOperatorName(child); name != "" {
				sym.Name = name
				sym.Pos = sanyFixDeclOperatorPosition(child)
				sym.Arity = 1
			}
		case "N_InfixDecl":
			if name := sanyFixDeclOperatorName(child); name != "" {
				sym.Name = name
				sym.Pos = sanyFixDeclOperatorPosition(child)
				sym.Arity = 2
			}
		}
	}
	if sym.Name == "" {
		return NewSymbol{}, nil, false
	}
	var setNode *SanySyntaxNode
	seenIn := false
	for _, child := range node.GetHeirs() {
		if child.Token != nil && child.Token.Kind == SanyTokenIN {
			seenIn = true
			continue
		}
		if seenIn && isSanyExpressionNode(child) {
			setNode = child
		}
	}
	if setNode == nil {
		return sym, nil, true
	}
	set, diags := sanyExpr(setNode)
	sym.Domain = set
	return sym, diags, true
}

func sanyNewSymbBound(node *SanySyntaxNode) (BoundVar, Diagnostics, bool) {
	id := sanyNewSymbIdentifier(node)
	if id == nil {
		return BoundVar{Pos: sanyNodePosition(node)}, nil, false
	}
	bound := BoundVar{Name: id.Image, Pos: sanyNodePosition(id)}
	var setNode *SanySyntaxNode
	seenIn := false
	for _, child := range node.GetHeirs() {
		if child.Token != nil && child.Token.Kind == SanyTokenIN {
			seenIn = true
			continue
		}
		if seenIn && isSanyExpressionNode(child) {
			setNode = child
		}
	}
	if setNode == nil {
		return bound, nil, true
	}
	set, diags := sanyExpr(setNode)
	bound.Set = set
	return bound, diags, true
}

func sanyNewSymbIdentifier(node *SanySyntaxNode) *SanySyntaxNode {
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() == "N_IdentDecl" {
			return firstSanyIdentifier(child)
		}
	}
	return nil
}

func isSanyAssumeProveProveToken(node *SanySyntaxNode) bool {
	return node != nil && node.Token != nil && (node.Token.Kind == SanyTokenProve || node.Token.Kind == SanyTokenBoxprove)
}

func expressionChildren(node *SanySyntaxNode) []*SanySyntaxNode {
	var out []*SanySyntaxNode
	for _, child := range node.GetHeirs() {
		if !isSanyExpressionNode(child) {
			continue
		}
		out = append(out, child)
	}
	return out
}

func firstSanyExpression(node *SanySyntaxNode) *SanySyntaxNode {
	for _, child := range node.GetHeirs() {
		if isSanyExpressionNode(child) {
			return child
		}
	}
	return nil
}

func lastSanyExpression(node *SanySyntaxNode) *SanySyntaxNode {
	heirs := node.GetHeirs()
	for i := len(heirs) - 1; i >= 0; i-- {
		if isSanyExpressionNode(heirs[i]) {
			return heirs[i]
		}
	}
	return nil
}

func isSanyExpressionNode(node *SanySyntaxNode) bool {
	if node == nil {
		return false
	}
	if isSanyProofSyntaxNode(node) {
		return false
	}
	if node.Token == nil {
		return true
	}
	switch node.Kind.JavaName() {
	case "IDENTIFIER", "NUMBER_LITERAL", "STRING_LITERAL", "N_String":
		return true
	default:
		return false
	}
}

func isSanyProofSyntaxNode(node *SanySyntaxNode) bool {
	switch node.Kind.JavaName() {
	case "N_Proof", "N_TerminalProof", "N_ProofStep", "N_QEDStep", "N_DefStep", "N_UseOrHide", "N_HaveStep", "N_TakeStep", "N_WitnessStep", "N_PickStep", "N_CaseStep", "N_AssertStep":
		return true
	default:
		return false
	}
}

func firstSanyIdentifier(node *SanySyntaxNode) *SanySyntaxNode {
	return firstSanyChildKind(node, "IDENTIFIER")
}

func firstSanyChildKind(node *SanySyntaxNode, kind string) *SanySyntaxNode {
	for _, child := range node.GetHeirs() {
		if child != nil && child.Kind.JavaName() == kind {
			return child
		}
	}
	return nil
}

func sanyGeneralIDName(node *SanySyntaxNode) string {
	heirs := node.GetHeirs()
	if len(heirs) < 2 {
		return ""
	}
	var parts []string
	for _, elem := range heirs[0].GetHeirs() {
		elemHeirs := elem.GetHeirs()
		if len(elemHeirs) == 0 {
			continue
		}
		if elemHeirs[0].Kind.JavaName() == "N_OpArgs" {
			continue
		}
		if name := sanySelectorName(elemHeirs[0]); name != "" {
			parts = append(parts, name)
		}
	}
	if name := sanySelectorName(heirs[1]); name != "" {
		parts = append(parts, name)
	}
	return strings.Join(parts, "!")
}

func sanyGeneralIDCall(node *SanySyntaxNode) (Expr, bool, Diagnostics) {
	heirs := node.GetHeirs()
	if len(heirs) < 2 {
		return nil, false, nil
	}
	var parts []string
	var args []Expr
	var diags Diagnostics
	collectArgs := func(argRoot *SanySyntaxNode) {
		for _, argNode := range expressionChildren(argRoot) {
			arg, argDiags := sanyExpr(argNode)
			diags = append(diags, argDiags...)
			args = append(args, arg)
		}
	}
	for _, elem := range heirs[0].GetHeirs() {
		elemHeirs := elem.GetHeirs()
		if len(elemHeirs) == 0 {
			continue
		}
		if elemHeirs[0].Kind.JavaName() == "N_OpArgs" {
			collectArgs(elemHeirs[0])
			continue
		}
		if name := sanySelectorName(elemHeirs[0]); name != "" {
			parts = append(parts, name)
		}
		if len(elemHeirs) > 1 && elemHeirs[1].Kind.JavaName() == "N_OpArgs" {
			collectArgs(elemHeirs[1])
		}
	}
	if heirs[1].Kind.JavaName() == "N_OpArgs" {
		collectArgs(heirs[1])
	} else if name := sanySelectorName(heirs[1]); name != "" {
		parts = append(parts, name)
	}
	if len(parts) == 0 || len(args) == 0 {
		return nil, false, nil
	}
	call := &CallExpr{Callee: &IdentExpr{Name: strings.Join(parts, "!"), Pos: sanyNodePosition(heirs[0])}, Args: args, Pos: sanyNodePosition(node)}
	return call, true, diags
}

func sanySelectorName(node *SanySyntaxNode) string {
	if node == nil {
		return ""
	}
	if node.Kind.JavaName() == "IDENTIFIER" {
		return node.Image
	}
	if node.Kind.JavaName() == "N_OpArgs" {
		return ""
	}
	if image := sanyOperatorImage(node); image != "" {
		return image
	}
	return sanyFirstTokenImage(node)
}

func sanyOperatorImage(node *SanySyntaxNode) string {
	if node == nil {
		return ""
	}
	if node.Token != nil {
		return sanyCanonicalOperatorImage(node.Image)
	}
	for _, child := range node.GetHeirs() {
		if image := sanyOperatorImage(child); image != "" {
			return image
		}
	}
	return ""
}

func sanyCanonicalOperatorImage(image string) string {
	if canonical := SanyOperatorSynonymCanonical[image]; canonical != "" {
		return canonical
	}
	return image
}

func sanyASTOperator(op string) string {
	if op == "\\X" {
		return "\\X"
	}
	canonical := ResolveSanyOperatorSynonym(op)
	switch canonical {
	case "\\land":
		return "/\\"
	case "\\lor":
		return "\\/"
	case "\\neg":
		return "\\lnot"
	case "/=":
		return "/="
	case "\\leq":
		return "\\leq"
	case "\\geq":
		return "\\geq"
	case "\\intersect":
		return "\\intersect"
	case "\\union":
		return "\\union"
	case "\\times":
		return "\\times"
	default:
		return canonical
	}
}

func sanyFirstTokenImage(node *SanySyntaxNode) string {
	if node == nil {
		return ""
	}
	if node.Token != nil {
		return node.Image
	}
	for _, child := range node.GetHeirs() {
		if image := sanyFirstTokenImage(child); image != "" {
			return image
		}
	}
	return ""
}

func sanyRealLiteralImage(node *SanySyntaxNode) string {
	heirs := node.GetHeirs()
	if len(heirs) >= 3 {
		return heirs[0].Image + heirs[1].Image + heirs[2].Image
	}
	return sanyFirstTokenImage(node)
}

func unsupportedSanyExpr(node *SanySyntaxNode) (Expr, Diagnostics) {
	name := "<nil>"
	pos := Position{}
	if node != nil {
		name = node.Kind.JavaName()
		pos = sanyNodePosition(node)
	}
	return &IdentExpr{Pos: pos}, Diagnostics{errorAt(pos, "E1400", "unsupported SANY expression %s", name)}
}

// BY has USE/HIDE's fact syntax. Only a bare GeneralId is generated in fact
// mode; references nested inside expressions remain expression selections.
func sanyLeafProofFacts(node *SanySyntaxNode) []ProofFact {
	return sanyProofFactsFromRefs(sanyLeafProofReferences(node))
}

func sanyProofFactsFromRefs(refs []ProofRef) []ProofFact {
	var facts []ProofFact
	for _, ref := range refs {
		if ref.Defs || ref.Expr == nil {
			continue
		}
		direct := false
		if source, ok := ref.Expr.(interface{ GetSyntaxNode() *SanySyntaxNode }); ok && source.GetSyntaxNode() != nil {
			direct = source.GetSyntaxNode().Kind.JavaName() == "N_GeneralId"
		}
		facts = append(facts, ProofFact{Expr: ref.Expr, Direct: direct})
	}
	return facts
}

func sanyLeafProofSyntax(node *SanySyntaxNode) *SanySyntaxNode {
	if node == nil {
		return nil
	}
	for _, child := range node.GetHeirs() {
		if child.Token != nil && child.Token.Kind == SanyTokenBy {
			return node
		}
	}
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() != "N_ProofStep" {
			if command := sanyLeafProofSyntax(child); command != nil {
				return command
			}
		}
	}
	return nil
}

func sanyLeafProofReferences(node *SanySyntaxNode) []ProofRef {
	if node == nil {
		return nil
	}
	for _, child := range node.GetHeirs() {
		if child.Token != nil && child.Token.Kind == SanyTokenBy {
			return sanyUseOrHideRefs(node)
		}
	}
	var refs []ProofRef
	for _, child := range node.GetHeirs() {
		if child.Kind.JavaName() != "N_ProofStep" {
			refs = append(refs, sanyLeafProofReferences(child)...)
		}
	}
	return refs
}
