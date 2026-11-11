package no_useless_recursion

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

// A declaration directly in a function body can share its name with a
// parameter. The body binding contains the function, even if tsgo returns
// only the parameter symbol.
func reusesEnclosingParameter(function *ast.Node, symbol *ast.Symbol) bool {
	if function.Kind != ast.KindFunctionDeclaration {
		return false
	}
	enclosing := ast.GetContainingFunction(function)
	return enclosing != nil && function.Parent == enclosing.Body() &&
		slices.ContainsFunc(symbol.Declarations, func(decl *ast.Node) bool {
			return ast.IsPartOfParameterDeclaration(decl) && ast.GetContainingFunction(decl) == enclosing
		})
}

type scopeBinding struct {
	scope *scope.Scope
	name  string
}

type functionBindings struct {
	scopes      *scope.Manager
	reassigned  map[*ast.Node]struct{}
	blockCopies map[scopeBinding]struct{}
}

func (bindings *functionBindings) mayBeReplaced(function *ast.Node) bool {
	if _, ok := bindings.reassigned[function]; ok {
		return true
	}
	if len(bindings.blockCopies) == 0 {
		return false
	}
	_, ok := bindings.blockCopies[scopeBinding{bindings.scopes.Acquire(function), function.Name().Text()}]
	return ok
}

func analyzeFunctionBindings(ctx rule.RuleContext) *functionBindings {
	manager := scopeanalysis.References(ctx, nil)
	bindings := &functionBindings{scopes: manager, reassigned: make(map[*ast.Node]struct{})}
	sourceType := ctx.LanguageOptions.EffectiveSourceType()
	for _, lexicalScope := range manager.Scopes {
		for _, binding := range lexicalScope.Vars {
			if binding.Kind != scope.DefFunctionName {
				continue
			}
			if slices.ContainsFunc(lexicalScope.Declarations(binding.Name), func(declaration *scope.Variable) bool {
				return declaration.Kind == scope.DefVariable && utils.IsVariableWriteReference(declaration.ID)
			}) {
				bindings.reassigned[binding.DefNode] = struct{}{}
			}
			if sourceType != "module" {
				bindings.collectBlockCopy(ctx, sourceType, binding)
			}
		}
	}

	var propertyNames *utils.StaticStringEvaluator
	for _, reference := range manager.References {
		if !reference.IsValueReference() {
			continue
		}
		id := reference.Identifier
		if utils.IsWriteReference(id) {
			for _, binding := range reference.Declarations {
				if binding.Kind == scope.DefFunctionName && !isBeforeFunctionBody(id, binding.Scope) &&
					!bindings.crossesBlockCopy(reference, binding.Scope) {
					bindings.reassigned[binding.DefNode] = struct{}{}
				}
			}
		}
		if sourceType != "script" || id.Text() != "globalThis" {
			continue
		}
		access := writtenProperty(id)
		if access == nil || !ctx.Refs.IsGlobalReference(id) || bindings.crossesBlockCopy(reference, manager.Global) {
			continue
		}
		if propertyNames == nil {
			propertyNames = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
			propertyNames.GlobalAccess = ctx.Globals.Access
		}
		if name, ok := propertyNames.EvalAccessExpressionName(access); ok {
			bindings.markReassigned(manager.Global.Declarations(name))
		}
	}
	return bindings
}

func (bindings *functionBindings) markReassigned(declarations []*scope.Variable) {
	for _, binding := range declarations {
		if binding.Kind == scope.DefFunctionName {
			bindings.reassigned[binding.DefNode] = struct{}{}
		}
	}
}

func writtenProperty(id *ast.Node) *ast.Node {
	object := id
	for object.Parent != nil && utils.SkipAssertionsAndParens(object.Parent) == id {
		object = object.Parent
	}
	access := object.Parent
	if access != nil && utils.AccessExpressionObject(access) == object && utils.IsWriteReference(access) {
		return access
	}
	return nil
}

// Parameter defaults run before the body's declarations are visible.
func isBeforeFunctionBody(id *ast.Node, lexicalScope *scope.Scope) bool {
	if lexicalScope.Kind != scope.KindFunction {
		return false
	}
	body := lexicalScope.Block.Body()
	return body != nil && id.Pos() < body.Pos()
}

func (bindings *functionBindings) crossesBlockCopy(ref *scope.Reference, stop *scope.Scope) bool {
	if len(bindings.blockCopies) == 0 {
		return false
	}
	for current := ref.From; current != nil && current != stop; current = current.Parent {
		if _, ok := bindings.blockCopies[scopeBinding{current, ref.Identifier.Text()}]; ok &&
			!isBeforeFunctionBody(ref.Identifier, current) {
			return true
		}
	}
	return false
}

// Annex B can copy a block function into the surrounding variable scope.
// The declaration inside the block still has a separate lexical binding.
func (bindings *functionBindings) collectBlockCopy(ctx rule.RuleContext, sourceType string, binding *scope.Variable) {
	function := binding.DefNode
	if binding.Scope.Kind != scope.KindBlock || function.Body() == nil ||
		ast.GetFunctionFlags(function)&(ast.FunctionFlagsAsync|ast.FunctionFlagsGenerator) != 0 ||
		utils.IsInStrictModeWithSourceType(function, ctx.SourceFile, sourceType) {
		return
	}
	target := binding.Scope.VariableScope()
	if target == nil || (target.Kind != scope.KindGlobal && target.Kind != scope.KindFunction) {
		return
	}
	for current := binding.Scope.Parent; current != nil; current = current.Parent {
		if slices.ContainsFunc(current.Declarations(binding.Name), func(decl *scope.Variable) bool {
			return blocksAnnexBCopy(decl, current == target)
		}) {
			return
		}
		if current == target {
			if bindings.blockCopies == nil {
				bindings.blockCopies = make(map[scopeBinding]struct{})
			}
			bindings.blockCopies[scopeBinding{target, binding.Name}] = struct{}{}
			bindings.markReassigned(target.Declarations(binding.Name))
			return
		}
	}
}

func blocksAnnexBCopy(decl *scope.Variable, isTargetScope bool) bool {
	if !decl.IsValueBinding || decl.DeclareModifier {
		return false
	}
	switch decl.Kind {
	case scope.DefFunctionName:
		return !isTargetScope
	case scope.DefVariable:
		return !utils.IsVarKeyword(utils.GetDeclListForSymbolDecl(decl.DefNode))
	case scope.DefCatch:
		// A simple catch parameter permits a same-name var declaration.
		catch := decl.Scope.Block.AsCatchClause()
		return !ast.IsIdentifier(catch.VariableDeclaration.Name())
	default:
		return true
	}
}
