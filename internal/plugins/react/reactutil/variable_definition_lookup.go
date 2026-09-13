package reactutil

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

type variableDefinitionOrder uint8

const (
	firstVariableDefinition variableDefinitionOrder = iota
	lastVariableDefinition
)

type variableDefinitionLookupKey struct {
	from  *scope.Scope
	name  string
	order variableDefinitionOrder
}

type variableScopeAnalysis struct {
	manager    *scope.Manager
	firstChild map[*scope.Scope]*scope.Scope
}

type variableScopeAnalysisFileCacheKey struct{}

// VariableDefinitionLookup mirrors eslint-plugin-react's variable utility.
// That utility searches the current scope, its first child, and that child's
// first child before moving to the parent. The returned definitions retain
// eslint-scope's source order rather than TypeScript's merged-symbol order.
type VariableDefinitionLookup struct {
	ctx      rule.RuleContext
	analysis *variableScopeAnalysis
	cache    map[variableDefinitionLookupKey]*scope.Variable
}

func NewVariableDefinitionLookup(ctx rule.RuleContext) *VariableDefinitionLookup {
	return &VariableDefinitionLookup{ctx: ctx}
}

// First returns eslint-plugin-react's variable.defs[0].node for name at ident.
func (l *VariableDefinitionLookup) First(ident *ast.Node, name string) *ast.Node {
	return definitionNode(l.lookup(ident, name, firstVariableDefinition))
}

// Last returns the last definition's node for name at ident.
func (l *VariableDefinitionLookup) Last(ident *ast.Node, name string) *ast.Node {
	return definitionNode(l.lookup(ident, name, lastVariableDefinition))
}

func definitionNode(definition *scope.Variable) *ast.Node {
	if definition == nil {
		return nil
	}
	return definition.DefNode
}

func (l *VariableDefinitionLookup) lookup(ident *ast.Node, name string, order variableDefinitionOrder) *scope.Variable {
	if l == nil || l.ctx.SourceFile == nil || ident == nil || ident.Kind != ast.KindIdentifier || name == "" {
		return nil
	}
	analysis := l.getAnalysis()
	from := analysis.manager.Acquire(ident)
	key := variableDefinitionLookupKey{from: from, name: name, order: order}
	if definition, ok := l.cache[key]; ok {
		return definition
	}
	definition := l.find(from, name, order)
	l.cache[key] = definition
	return definition
}

func (l *VariableDefinitionLookup) getAnalysis() *variableScopeAnalysis {
	if l.analysis != nil {
		return l.analysis
	}
	l.analysis = rule.CachedByFile(l.ctx, variableScopeAnalysisFileCacheKey{}, func() *variableScopeAnalysis {
		manager := scope.Build(l.ctx.SourceFile, scope.Options{})
		firstChild := make(map[*scope.Scope]*scope.Scope)
		for _, current := range manager.Scopes {
			if current.Parent != nil && firstChild[current.Parent] == nil {
				firstChild[current.Parent] = current
			}
		}
		return &variableScopeAnalysis{manager: manager, firstChild: firstChild}
	})
	l.cache = make(map[variableDefinitionLookupKey]*scope.Variable)
	return l.analysis
}

func (l *VariableDefinitionLookup) find(from *scope.Scope, name string, order variableDefinitionOrder) *scope.Variable {
	for current := from; current != nil; current = current.Parent {
		candidate := current
		for range 3 {
			if candidate == nil {
				break
			}
			if declarations := candidate.Declarations(name); len(declarations) != 0 {
				if order == lastVariableDefinition {
					return declarations[len(declarations)-1]
				}
				return declarations[0]
			}
			// ESLint materializes configured globals as variables without defs.
			// Stop here instead of looking through that binding into a child scope.
			if candidate.Kind == scope.KindGlobal &&
				l.ctx.LanguageOptions.EffectiveSourceType() == "script" &&
				l.ctx.Globals.Access(name).IsDeclared() {
				return nil
			}
			candidate = l.analysis.firstChild[candidate]
		}
	}
	return nil
}
