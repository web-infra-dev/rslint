// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
package prefer_object_from_entries

import (
	"cmp"
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed prefer_object_from_entries.schema.json
var schemaJSON []byte

var PreferObjectFromEntriesRule = rule.Rule{
	Name:   "unicorn/prefer-object-from-entries",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var functions []string
		if len(options) > 0 {
			settings, _ := options[0].(map[string]any)
			configured, _ := settings["functions"].([]any)
			for _, value := range configured {
				if name, ok := value.(string); ok {
					functions = append(functions, ecmascript.StringTrim(name))
				}
			}
		}
		functions = append(functions, "_.fromPairs", "lodash.fromPairs")
		return rule.RuleListeners{
			ast.KindForOfStatement: func(node *ast.Node) { checkForOfLoop(ctx, node) },
			ast.KindCallExpression: func(node *ast.Node) {
				if call, callback, property := reduceProperty(ctx, node); property != nil {
					ctx.ReportNodeWithDeferredFixes(call.Property, rule.RuleMessage{
						Id: "reduce", Description: "Prefer `Object.fromEntries()` over `Array#reduce()`.",
					}, func() []rule.RuleFix { return reduceFixes(ctx, call, callback, property) })
				}

				if ast.IsOptionalChainRoot(node) || len(node.Arguments()) != 1 || node.Arguments()[0].Kind == ast.KindSpreadElement {
					return
				}
				callee := utils.ESTreeCallCallee(node.Expression())
				for _, name := range functions {
					if unicornutil.NodeMatchesPath(callee, name) {
						ctx.ReportNodeWithDeferredFixes(callee, rule.RuleMessage{
							Id: "function", Description: "Prefer `Object.fromEntries()` over `" + name + "()`.",
							Data: map[string]string{"functionName": name},
						}, func() []rule.RuleFix {
							return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, callee, "Object.fromEntries")}
						})
					}
				}
			},
		}
	},
}

func isEmptyObject(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && ast.IsEmptyObjectLiteral(node)
}

func sameIdentifier(left, right *ast.Node) bool {
	return unicornutil.IsSameIdentifier(utils.ESTreeRuntimeExpression(left), utils.ESTreeRuntimeExpression(right))
}

// reduceProperty recognizes only the two upstream callback shapes. Reference
// counting belongs to RefStore; the rule only decides whether removing the
// accumulator parameter is valid for this transformation.
func reduceProperty(ctx rule.RuleContext, node *ast.Node) (unicornutil.DotMethodCall, *ast.Node, *ast.Node) {
	two := 2
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Method: "reduce", ArgumentsLength: &two, RejectSpreadElement: true,
	})
	if !ok || !isEmptyObject(node.Arguments()[1]) {
		return call, nil, nil
	}
	callback := utils.ESTreeRuntimeExpression(node.Arguments()[0])
	if callback.Kind != ast.KindArrowFunction || ast.HasSyntacticModifier(callback, ast.ModifierFlagsAsync) {
		return call, nil, nil
	}
	parameters := callback.AsArrowFunction().Parameters.Nodes
	if len(parameters) == 0 {
		return call, nil, nil
	}
	accumulator := unicornutil.PlainParameterIdentifier(parameters[0])
	if accumulator == nil {
		return call, nil, nil
	}
	body := utils.ESTreeRuntimeExpression(callback.Body())
	var property *ast.Node
	if assign, ok := unicornutil.MatchDotMethodCall(body, unicornutil.DotMethodCallOptions{
		Method: "assign", ArgumentsLength: &two, RejectSpreadElement: true,
	}); ok && unicornutil.NodeMatchesPath(assign.Object, "Object") && sameIdentifier(accumulator, body.Arguments()[0]) {
		object := utils.ESTreeRuntimeExpression(body.Arguments()[1])
		if object.Kind == ast.KindObjectLiteralExpression && len(object.AsObjectLiteralExpression().Properties.Nodes) == 1 {
			property = object.AsObjectLiteralExpression().Properties.Nodes[0]
		}
	} else if body.Kind == ast.KindObjectLiteralExpression {
		properties := body.AsObjectLiteralExpression().Properties.Nodes
		if len(properties) == 2 && properties[0].Kind == ast.KindSpreadAssignment &&
			sameIdentifier(accumulator, properties[0].AsSpreadAssignment().Expression) {
			property = properties[1]
		}
	}
	name, _, dataProperty := unicornutil.ObjectDataProperty(property)
	if !dataProperty || (name.Kind != ast.KindComputedPropertyName && name.Text() == "__proto__") ||
		unicornutil.IsKnownNonArray(ctx, call.Object) || len(ctx.Refs.References(parameters[0].Symbol())) != 1 {
		return call, nil, nil
	}
	return call, callback, property
}

func reduceFixes(ctx rule.RuleContext, call unicornutil.DotMethodCall, callback, property *ast.Node) []rule.RuleFix {
	if len(call.Call.TypeArguments()) > 0 || callback.Type() != nil {
		return nil
	}
	sf := ctx.SourceFile
	callbackRange := utils.TrimNodeTextRange(sf, callback)
	if utils.HasCommentInSpan(ctx.Comments.All(), callbackRange.Pos(), callbackRange.End()) {
		return nil
	}
	initialRange := utils.TrimNodeTextRange(sf, call.Call.Arguments()[1])
	comma, ok := utils.TokenBeforePosition(sf, initialRange.Pos())
	if !ok || comma.Kind != ast.KindCommaToken || utils.HasCommentInSpan(ctx.Comments.All(), comma.Start, initialRange.End()) {
		return nil
	}
	fixes := []rule.RuleFix{
		rule.RuleFixInsertBefore(sf, call.Call, "Object.fromEntries("),
		rule.RuleFixInsertAfter(call.Call, ")"),
		rule.RuleFixReplace(sf, call.Property, "flatMap"),
		rule.RuleFixRemoveRange(core.NewTextRange(comma.Start, initialRange.End())),
	}
	first := callback.AsArrowFunction().Parameters.Nodes[0]
	after, _ := utils.TokenAtOrAfter(sf, first.End())
	if after.Kind == ast.KindCommaToken {
		next, _ := utils.TokenAtOrAfter(sf, after.End)
		fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(after.Start, next.Start)))
	}
	replacement := ""
	if len(callback.AsArrowFunction().Parameters.Nodes) == 1 && after.Kind == ast.KindEqualsGreaterThanToken {
		replacement = "()"
	}
	fixes = append(fixes, rule.RuleFixReplace(sf, first, replacement))

	key, value, _ := unicornutil.ObjectDataProperty(property)
	computed := key.Kind == ast.KindComputedPropertyName
	if computed {
		key = key.AsComputedPropertyName().Expression
	}
	keyText := utils.TrimmedNodeText(sf, key)
	if !computed && key.Kind == ast.KindIdentifier {
		keyText = "'" + keyText + "'"
	}
	body := utils.ESTreeRuntimeExpression(callback.Body())
	// Emit one entry per callback invocation. flatMap skips holes and passes
	// the original indices to the callback.
	fixes = append(fixes, rule.RuleFixReplace(sf, body, "[["+keyText+", "+utils.TrimmedNodeText(sf, value)+"]]"))
	// Remove only the surrounding parenthesis tokens, retaining their spacing.
	for wrapper := callback.Body(); wrapper != body; wrapper = wrapper.Expression() {
		wrapperRange := utils.TrimNodeTextRange(sf, wrapper)
		fixes = append(fixes,
			rule.RuleFixRemoveRange(core.NewTextRange(wrapperRange.Pos(), wrapperRange.Pos()+1)),
			rule.RuleFixRemoveRange(core.NewTextRange(wrapperRange.End()-1, wrapperRange.End())),
		)
	}
	return fixes
}

func checkForOfLoop(ctx rule.RuleContext, loop *ast.Node) {
	if loop.Parent == nil || !loop.Parent.CanHaveStatements() {
		return
	}
	// Statement positions are ordered, so a wide block needs no linear scan
	// for each loop. Declarations without a following loop need no listener.
	statements := loop.Parent.Statements()
	index, found := slices.BinarySearchFunc(statements, loop, func(left, right *ast.Node) int {
		return cmp.Compare(left.Pos(), right.Pos())
	})
	if !found || index == 0 {
		return
	}
	declaration := statements[index-1]
	// ESTree places an exported declaration inside ExportNamedDeclaration,
	// where it has no next sibling.
	if declaration.Kind != ast.KindVariableStatement || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsExport) {
		return
	}
	declarations := declaration.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes
	if len(declarations) != 1 {
		return
	}
	declarator := declarations[0].AsVariableDeclaration()
	if declarator.Name().Kind != ast.KindIdentifier || !isEmptyObject(declarator.Initializer) {
		return
	}
	forOf := loop.AsForInOrOfStatement()
	if forOf.AwaitModifier != nil || forOf.Initializer.Kind != ast.KindVariableDeclarationList {
		return
	}
	list := forOf.Initializer.AsVariableDeclarationList()
	if forOf.Initializer.Flags&(ast.NodeFlagsLet|ast.NodeFlagsConst) == 0 || len(list.Declarations.Nodes) != 1 {
		return
	}
	pattern := list.Declarations.Nodes[0].Name()
	if pattern.Kind != ast.KindArrayBindingPattern || len(pattern.AsBindingPattern().Elements.Nodes) != 2 {
		return
	}
	var pair [2]*ast.Node
	for i, element := range pattern.AsBindingPattern().Elements.Nodes {
		if element.Kind != ast.KindBindingElement {
			return
		}
		binding := element.AsBindingElement()
		if binding.DotDotDotToken != nil || binding.Initializer != nil || binding.Name() == nil || binding.Name().Kind != ast.KindIdentifier || sameIdentifier(binding.Name(), declarator.Name()) {
			return
		}
		pair[i] = binding.Name()
	}
	for _, reference := range ctx.Refs.References(declarations[0].Symbol()) {
		if reference.Pos() >= forOf.Expression.Pos() && reference.End() <= forOf.Expression.End() {
			return
		}
	}
	body := forOf.Statement
	if body.Kind == ast.KindBlock {
		if len(body.AsBlock().Statements.Nodes) != 1 {
			return
		}
		body = body.AsBlock().Statements.Nodes[0]
	}
	if body.Kind != ast.KindExpressionStatement {
		return
	}
	assignment := utils.ESTreeRuntimeExpression(body.Expression())
	if assignment.Kind != ast.KindBinaryExpression || assignment.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
		return
	}
	binary := assignment.AsBinaryExpression()
	left := utils.ESTreeRuntimeExpression(binary.Left)
	if left.Kind != ast.KindElementAccessExpression || !sameIdentifier(left.Expression(), declarator.Name()) ||
		!sameIdentifier(left.AsElementAccessExpression().ArgumentExpression, pair[0]) || !sameIdentifier(binary.Right, pair[1]) {
		return
	}
	ctx.ReportNodeWithDeferredFixes(loop, rule.RuleMessage{
		Id: "loop", Description: "Prefer `Object.fromEntries()` over a `for-of` loop.",
	}, func() []rule.RuleFix {
		sf := ctx.SourceFile
		end := len(sf.Text())
		if next, ok := utils.TokenAtOrAfter(sf, loop.End()); ok {
			end = next.Start
		}
		if utils.HasCommentInSpan(ctx.Comments.All(), utils.TrimNodeTextRange(sf, declaration).Pos(), end) {
			return nil
		}
		loopRange := utils.TrimNodeTextRange(sf, loop)
		removal := rule.RuleFixRemoveRange(unicornutil.StatementRemovalRange(sf, loopRange))
		// Upstream preserves this statement boundary independently of the next
		// token, so simultaneous fixes cannot expose an ASI continuation.
		if previous, ok := utils.TokenBeforePosition(sf, loopRange.Pos()); ok && previous.Kind != ast.KindSemicolonToken {
			removal = rule.RuleFixReplaceRange(loopRange, ";")
		}
		return []rule.RuleFix{
			rule.RuleFixReplace(sf, utils.ESTreeRuntimeExpression(declarator.Initializer), "Object.fromEntries("+utils.TrimmedNodeText(sf, forOf.Expression)+")"),
			removal,
		}
	})
}
