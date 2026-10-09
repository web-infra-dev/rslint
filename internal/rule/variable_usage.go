package rule

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// VariableUsage is a live, read-only view of explicit usage marks for one file
// pass. It does not count ordinary references. Its zero value is empty.
// Retaining the owner lets a view created before the first mark observe later
// writes without allocating usage state for rules that only read it.
type VariableUsage struct {
	cache      *FileCache
	sourceFile *ast.SourceFile
}

// VariableUsage returns the file-pass view shared by native rules. Reads see
// marks made so far; consumers must run after their relevant producers. The
// view and its marks must not be retained beyond this file's lint pass.
func (ctx *RuleContext) VariableUsage() VariableUsage {
	if ctx == nil {
		return VariableUsage{}
	}
	return VariableUsage{cache: ctx.fileCache, sourceFile: ctx.SourceFile}
}

// IsDeclarationUsed reports whether the binding declared by id was explicitly
// marked. Pass the binding identifier, not its enclosing declaration. For a
// class declaration this queries the outer binding; the distinct inner name
// is available through IsScopeBindingUsed. A named class expression has only
// an inner binding, so marking it does mark its declaration identifier.
func (usage VariableUsage) IsDeclarationUsed(id *ast.Node) bool {
	state := usage.state()
	return state != nil && state.declarations[id]
}

// IsScopeBindingUsed queries the exact binding named name in lexical, without
// searching its parents. Scope trees built with different reference options
// for the same source share binding identities. This also distinguishes a
// class's inner name and declaration-less function arguments bindings.
func (usage VariableUsage) IsScopeBindingUsed(lexical *scope.Scope, name string) bool {
	state := usage.state()
	if state == nil || lexical == nil || lexical.Block == nil {
		return false
	}
	if lexical.Kind == scope.KindGlobal && lexical.Block != state.sourceFile.AsNode() {
		return false
	}
	key := state.key(lexical, name)
	binding := state.scopes[key]
	return binding != nil && binding.key == key && binding.used
}

// IsGlobalUsed queries a global binding, including configured and inline
// globals. Marking a shadowing module, CommonJS wrapper, or function binding
// does not mark the global. Script-level declarations belong to the global.
func (usage VariableUsage) IsGlobalUsed(name string) bool {
	state := usage.state()
	if state == nil {
		return false
	}
	binding := state.scopes[usageBindingKey{name: name}]
	return binding != nil && binding.used
}

func (usage VariableUsage) state() *variableUsageState {
	if usage.cache == nil || usage.sourceFile == nil {
		return nil
	}
	state := usage.cache.usage
	if state == nil || state.sourceFile != usage.sourceFile {
		return nil
	}
	return state
}

// MarkVariableAsUsed marks the nearest binding named name in the syntactic
// scope at location, as ESLint's SourceCode.markVariableAsUsed does. This is
// not ordinary reference resolution: for example parameter defaults can mark
// declarations in the function body. The write is immediate and monotonic.
//
// A nil location starts at SourceFile. It returns false for an unknown name,
// foreign node, or context missing SourceFile/FileCache. Manually assembled
// contexts must attach a fresh FileCache per file pass. Native rules for a
// file run serially; this API adds no synchronization or traversal phase.
func (ctx *RuleContext) MarkVariableAsUsed(name string, location *ast.Node) bool {
	if ctx == nil || ctx.SourceFile == nil || ctx.fileCache == nil || name == "" {
		return false
	}
	if location == nil {
		location = ctx.SourceFile.AsNode()
	}
	state := ctx.fileCache.usage
	if state != nil && state.sourceFile != ctx.SourceFile {
		return false
	}
	lookup := usageLookup{location: location, name: name}
	if state != nil {
		if binding, ok := state.locations[lookup]; ok {
			return state.mark(binding)
		}
	}
	if !usageNodeBelongsTo(location, ctx.SourceFile) {
		return false
	}
	if state == nil {
		_, init, _ := ResolveLanguageDefaults(ctx.SourceFile.FileName(), ctx.LanguageOptions)
		if ctx.Refs != nil {
			init = ctx.Refs.init
		}
		state = &variableUsageState{
			sourceFile:     ctx.SourceFile,
			analysis:       ctx.ScopeAnalysis(),
			globals:        ctx.Globals,
			init:           init,
			globalTopLevel: init.globalTopLevelScope || (!init.nonGlobalTopLevelScope && !ast.IsExternalModule(ctx.SourceFile)),
			locations:      make(map[usageLookup]*usageBinding),
			scopes:         make(map[usageBindingKey]*usageBinding),
		}
		ctx.fileCache.usage = state
	}
	binding := state.resolve(name, location)
	state.locations[lookup] = binding // nil caches a missing binding, never a false usage flag.
	return state.mark(binding)
}

type usageLookup struct {
	location *ast.Node
	name     string
}

type usageBindingKey struct {
	block *ast.Node
	kind  scope.Kind
	name  string
}

type usageBinding struct {
	key          usageBindingKey
	declarations []*scope.Variable
	used         bool
}

type variableUsageState struct {
	sourceFile     *ast.SourceFile
	analysis       *scope.Cache
	globals        Globals
	init           RefStoreInit
	globalTopLevel bool
	locations      map[usageLookup]*usageBinding
	scopes         map[usageBindingKey]*usageBinding
	declarations   map[*ast.Node]bool
}

func (state *variableUsageState) resolve(name string, location *ast.Node) *usageBinding {
	manager := state.analysis.PeekDeclarations()
	// HasIdentifier is tsgo's existing lazy source-only name index. Absence
	// rules out authored bindings, but not globals or implicit arguments.
	if manager == nil && name != "arguments" && !state.sourceFile.HasIdentifier(name) {
		return state.global(name)
	}
	if manager == nil {
		manager = state.analysis.Declarations()
	}
	return state.resolveScope(manager.Acquire(location), name)
}

func (state *variableUsageState) resolveScope(lexical *scope.Scope, name string) *usageBinding {
	key := state.key(lexical, name)
	if binding, ok := state.scopes[key]; ok {
		return binding
	}
	var binding *usageBinding
	declarations := lexical.Declarations(name)
	switch {
	case len(declarations) != 0:
		binding = &usageBinding{key: key, declarations: declarations}
	case name == "arguments" && lexical.Kind == scope.KindFunction && lexical.Block.Kind != ast.KindArrowFunction:
		binding = &usageBinding{key: key}
	case lexical.Parent != nil:
		binding = state.resolveScope(lexical.Parent, name)
	case state.init.hasImplicitWrapperBinding(name):
		binding = &usageBinding{key: key}
	default:
		binding = state.global(name)
	}
	state.scopes[key] = binding
	return binding
}

func (state *variableUsageState) global(name string) *usageBinding {
	key := usageBindingKey{name: name}
	if binding, ok := state.scopes[key]; ok {
		return binding
	}
	access := state.globals.Access(name)
	// Parser-provided type bindings survive runtime global overrides. Marking
	// searches both declaration spaces, just like ESLint scope.variables.
	declared := access.IsDeclared() || (!ast.IsInJSFile(state.sourceFile.AsNode()) &&
		IsDefaultTypeScriptTypeGlobal(name))
	var binding *usageBinding
	if declared {
		binding = &usageBinding{key: key}
	}
	state.scopes[key] = binding
	return binding
}

func (state *variableUsageState) key(lexical *scope.Scope, name string) usageBindingKey {
	if lexical.Kind == scope.KindGlobal && state.globalTopLevel {
		return usageBindingKey{name: name}
	}
	return usageBindingKey{block: lexical.Block, kind: lexical.Kind, name: name}
}

func (state *variableUsageState) mark(binding *usageBinding) bool {
	if binding == nil {
		return false
	}
	if binding.used {
		return true
	}
	binding.used = true
	for _, declaration := range binding.declarations {
		if declaration.Anonymous || declaration.ID == nil {
			continue
		}
		if declaration.Kind == scope.DefClassInnerName && declaration.DefNode.Kind != ast.KindClassExpression {
			continue
		}
		if state.declarations == nil {
			state.declarations = make(map[*ast.Node]bool)
		}
		state.declarations[declaration.ID] = true
	}
	return true
}

func usageNodeBelongsTo(node *ast.Node, sourceFile *ast.SourceFile) bool {
	for current := node; current != nil; current = current.Parent {
		if current == sourceFile.AsNode() {
			return true
		}
	}
	return false
}
