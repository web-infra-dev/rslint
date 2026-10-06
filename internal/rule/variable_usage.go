package rule

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

type variableUsage struct {
	scopes      *scope.Manager
	identifiers map[*ast.Node]struct{}
	globals     map[string]struct{}
}

// MarkVariableAsUsed records an implicit use in the ESLint scope at location.
// It is shared by native rules for this lint pass only; it never changes the
// Program's symbols or the reference index. Consumers must check the marks
// after ordinary AST listeners have run. Disable comments suppress reports,
// not this side effect, matching ESLint's markVariableAsUsed.
func (ctx RuleContext) MarkVariableAsUsed(name string, location *ast.Node) bool {
	if ctx.fileCache == nil || ctx.SourceFile == nil || location == nil || name == "" {
		return false
	}
	usage := ctx.fileCache.variableUsage
	if usage == nil {
		// Acquire models getScope rather than reference resolution: even a
		// parameter initializer acquires the function's body declarations.
		usage = &variableUsage{scopes: scope.Build(ctx.SourceFile, scope.Options{})}
		ctx.fileCache.variableUsage = usage
	}
	for current := usage.scopes.Acquire(location); current != nil; current = current.Parent {
		declarations := current.Declarations(name)
		if len(declarations) == 0 {
			continue
		}
		for _, declaration := range declarations {
			// A class's inner name does not keep its outer declaration alive.
			if declaration.Kind == scope.DefClassInnerName {
				continue
			}
			if usage.identifiers == nil {
				usage.identifiers = make(map[*ast.Node]struct{})
			}
			// Binding identifiers distinguish import specifiers and destructured
			// parameters, and bridge separate local/export symbols.
			usage.identifiers[declaration.ID] = struct{}{}
		}
		return true
	}
	if ctx.Refs != nil && ctx.Refs.HasImplicitWrapperBinding(name) {
		return true
	}
	if !ctx.Globals.Access(name).IsDeclared() {
		return false
	}
	if usage.globals == nil {
		usage.globals = make(map[string]struct{})
	}
	usage.globals[name] = struct{}{}
	return true
}

// IsVariableMarkedAsUsed reports an implicit use of this binder-owned binding.
func (ctx RuleContext) IsVariableMarkedAsUsed(symbol *ast.Symbol) bool {
	if ctx.fileCache == nil || ctx.fileCache.variableUsage == nil || symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if _, used := ctx.fileCache.variableUsage.identifiers[declaration.Name()]; used {
			return true
		}
	}
	return false
}

// IsGlobalMarkedAsUsed reports an implicit use of a declaration-less global.
// A mark on a local binding with the same name does not mark the global.
func (ctx RuleContext) IsGlobalMarkedAsUsed(name string) bool {
	if ctx.fileCache == nil || ctx.fileCache.variableUsage == nil {
		return false
	}
	_, used := ctx.fileCache.variableUsage.globals[name]
	return used
}
