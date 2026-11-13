package consistent_json_file_read

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Folding describes the original properties. Before relying on them, check
// that the objects used to build the options have no potentially escaping use.
func optionsStayLocal(ctx *rule.RuleContext, evaluator *utils.StaticStringEvaluator, node *ast.Node) bool {
	refs := optionReferences{ctx: ctx, evaluator: evaluator}
	if !refs.collectInputs(node) {
		return false
	}
	for symbol := range refs.bindings {
		if !refs.checkUses(symbol) {
			return false
		}
	}
	return true
}

type optionReferences struct {
	ctx       *rule.RuleContext
	evaluator *utils.StaticStringEvaluator
	inputs    map[*ast.Node]bool
	bindings  map[*ast.Symbol]bool
}

func (refs *optionReferences) collectInputs(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || ast.IsTypeNode(node) {
		return true
	}
	if !ast.IsIdentifier(node) {
		return !node.ForEachChild(func(child *ast.Node) bool { return !refs.collectInputs(child) })
	}
	if !utils.IsReadReference(node) {
		return true
	}
	initializer, ok := refs.evaluator.ResolveIdentifierInitializer(node)
	if !ok {
		return true
	}
	if value, known := refs.evaluator.EvalValue(node); known && utils.IsPrimitiveStaticValue(value) {
		return true
	}
	symbol := refs.ctx.Refs.ResolveInFile(node)
	if symbol == nil {
		return false
	}
	if refs.inputs == nil {
		refs.inputs = make(map[*ast.Node]bool)
		refs.bindings = make(map[*ast.Symbol]bool)
	}
	refs.inputs[node] = true
	if _, seen := refs.bindings[symbol]; seen {
		return true
	}
	refs.bindings[symbol] = false
	return refs.collectInputs(initializer)
}

func (refs *optionReferences) checkUses(symbol *ast.Symbol) bool {
	if refs.bindings[symbol] {
		return true
	}
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if !ast.IsVariableDeclaration(declaration) || !ast.IsIdentifier(declaration.Name()) ||
		ast.IsVarUsing(declaration) || ast.IsVarAwaitUsing(declaration) ||
		ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 ||
		refs.ctx.IsExportedGlobalBinding(utils.BindingNameSymbol(declaration.Name()), declaration.Name().Text()) {
		return false
	}
	refs.bindings[symbol] = true
	for _, reference := range refs.ctx.Refs.References(symbol) {
		if !refs.inputs[reference] && utils.IsReadReference(reference) && !refs.localUse(reference) {
			return false
		}
	}
	return true
}

func (refs *optionReferences) localUse(node *ast.Node) bool {
	for parent := node.Parent; parent != nil; parent = node.Parent {
		if ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments) && parent.Expression() == node {
			node = parent
			continue
		}
		switch parent.Kind {
		case ast.KindVariableDeclaration:
			return parent.Initializer() == node && refs.checkUses(utils.BindingNameSymbol(parent.Name()))
		case ast.KindExpressionStatement:
			return true
		case ast.KindCallExpression:
			switch refs.objectMethod(parent, node) {
			case "isFrozen", "isSealed", "isExtensible":
				return true
			case "freeze", "seal", "preventExtensions":
				// The returned object can be assigned or passed on as an alias.
				node = parent
				continue
			}
		}
		break
	}
	return false
}

func (refs *optionReferences) objectMethod(call, argument *ast.Node) string {
	one := 1
	match, ok := unicornutil.MatchDotMethodCall(call, unicornutil.DotMethodCallOptions{ArgumentsLength: &one})
	if !ok || call.Arguments()[0] != argument {
		return ""
	}
	object := utils.ESTreeRuntimeExpression(match.Object)
	if !ast.IsIdentifier(object) || object.Text() != "Object" ||
		!refs.ctx.Globals.Access("Object").IsDeclared() || !refs.ctx.Refs.IsGlobalReference(object) {
		return ""
	}
	return match.Property.Text()
}
