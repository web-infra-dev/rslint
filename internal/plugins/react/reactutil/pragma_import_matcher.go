package reactutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// newPragmaImportMatcher mirrors isDestructuredFromPragmaImport's name lookup.
// Upstream searches the current scope, its first child, and that child's first
// child before moving to the parent. It reads the last definition regardless
// of whether that definition declares a value. Reference resolution alone
// therefore cannot answer this particular React question.
func newPragmaImportMatcher(ctx rule.RuleContext, pragma, name string) func(*ast.Node) bool {
	sourceFile := ctx.SourceFile
	// Only a script puts configured globals in the same scope as its code.
	// Modules and CommonJS wrappers search their own children first.
	hasGlobalBinding := ctx.LanguageOptions.EffectiveSourceType() == "script" && ctx.Globals.Access(name).IsDeclared()
	if pragma == "" {
		pragma = DefaultReactPragma
	}
	var pragmaLower string
	var manager *scope.Manager
	var firstChild map[*scope.Scope]*scope.Scope
	var matches map[*scope.Scope]bool
	return func(ident *ast.Node) bool {
		if sourceFile == nil || ident == nil || ident.Kind != ast.KindIdentifier || ident.Text() != name {
			return false
		}
		if manager == nil {
			pragmaLower = ecmascript.StringToLowerCase(pragma)
			manager = scope.Build(sourceFile, scope.Options{})
			firstChild = make(map[*scope.Scope]*scope.Scope)
			for _, current := range manager.Scopes {
				if current.Parent != nil && firstChild[current.Parent] == nil {
					firstChild[current.Parent] = current
				}
			}
			matches = make(map[*scope.Scope]bool)
		}
		from := manager.Acquire(ident)
		if matched, ok := matches[from]; ok {
			return matched
		}
		definition := latestPragmaDefinition(from, name, firstChild, hasGlobalBinding)
		matched := isDestructuredFromPragmaDeclaration(definition, pragma, pragmaLower)
		matches[from] = matched
		return matched
	}
}

func latestPragmaDefinition(from *scope.Scope, name string, firstChild map[*scope.Scope]*scope.Scope, hasGlobalBinding bool) *ast.Node {
	for current := from; current != nil; current = current.Parent {
		candidate := current
		for range 3 {
			if candidate == nil {
				break
			}
			if declarations := candidate.Declarations(name); len(declarations) != 0 {
				return declarations[len(declarations)-1].DefNode
			}
			if candidate.Kind == scope.KindGlobal && hasGlobalBinding {
				return nil
			}
			candidate = firstChild[candidate]
		}
	}
	return nil
}
