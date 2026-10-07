// Package no_useless_spread ports eslint-plugin-unicorn v77.0.0.
package no_useless_spread

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var cloneMethods = []string{"flat", "slice", "splice", "toReversed", "toSorted", "toSpliced", "with"}

var NoUselessSpreadRule = rule.Rule{
	Name:   "unicorn/no-useless-spread",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				checkLiteralSpread(ctx, node)
				if argument := singleSpreadArgument(node); argument != nil {
					checkIterableConversion(ctx, node)
					checkArrayClone(ctx, node, argument)
				}
			},
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				checkLiteralSpread(ctx, node)
				checkObjectAssign(ctx, node)
			},
			ast.KindNewExpression: func(node *ast.Node) {
				name := collectionConstructor(node)
				if name == "" || node.Arguments()[0].Kind != ast.KindSpreadElement {
					return
				}
				spread := node.Arguments()[0]
				if utils.ESTreeRuntimeExpression(spread.Expression()).Kind == ast.KindArrayLiteralExpression {
					return
				}
				constructor := "new " + name + "(…)"
				ctx.ReportNode(spread, rule.RuleMessage{
					Id:          "spread-in-collection-constructor",
					Description: "`" + constructor + "` accepts a single iterable argument, spreading is misleading.",
					Data:        map[string]string{"constructorName": constructor},
				})
			},
		}
	},
}

func singleSpreadArgument(node *ast.Node) *ast.Node {
	if node.Kind != ast.KindArrayLiteralExpression {
		return nil
	}
	elements := node.AsArrayLiteralExpression().Elements.Nodes
	if len(elements) != 1 || elements[0].Kind != ast.KindSpreadElement {
		return nil
	}
	return utils.ESTreeRuntimeExpression(elements[0].Expression())
}

func constructorName(node *ast.Node) string {
	if node == nil || !ast.IsNewExpression(node) {
		return ""
	}
	callee := utils.ESTreeRuntimeExpression(node.Expression())
	if callee != nil && ast.IsIdentifier(callee) {
		return callee.Text()
	}
	return ""
}

func collectionConstructor(node *ast.Node) string {
	name := constructorName(node)
	if name != "" && len(node.Arguments()) == 1 {
		switch name {
		case "Map", "WeakMap", "Set", "WeakSet":
			return name
		}
	}
	return ""
}

func checkLiteralSpread(ctx rule.RuleContext, node *ast.Node) {
	spread := utils.ESTreeParent(node)
	if spread == nil || (spread.Kind != ast.KindSpreadElement && spread.Kind != ast.KindSpreadAssignment) {
		return
	}
	parent := spread.Parent
	argumentType, parentDescription := "array", ""
	switch parent.Kind {
	case ast.KindArrayLiteralExpression:
		parentDescription = "array literal"
	case ast.KindObjectLiteralExpression:
		if node.Kind != ast.KindObjectLiteralExpression || !canFlattenObject(node) {
			return
		}
		argumentType, parentDescription = "object", "object literal"
	case ast.KindCallExpression, ast.KindNewExpression:
		parentDescription = "arguments"
	default:
		return
	}
	if argumentType == "array" && node.Kind != ast.KindArrayLiteralExpression {
		return
	}
	spreadRange := utils.TrimNodeTextRange(ctx.SourceFile, spread)
	ctx.ReportRangeWithDeferredFixes(core.NewTextRange(spreadRange.Pos(), spreadRange.Pos()+3), rule.RuleMessage{
		Id:          "spread-in-list",
		Description: "Spread an " + argumentType + " literal in " + parentDescription + " is unnecessary.",
		Data:        map[string]string{"argumentType": argumentType, "parentDescription": parentDescription},
	}, func() []rule.RuleFix {
		// In new Map(...[...iterable]), neither removing only the inner array
		// nor guessing the intended collection input is a safe automatic edit.
		if collectionConstructor(parent) != "" && singleSpreadArgument(node) != nil {
			return nil
		}
		return flattenLiteral(ctx, node, spread)
	})
}

func canFlattenObject(node *ast.Node) bool {
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		switch property.Kind {
		case ast.KindGetAccessor, ast.KindSetAccessor:
			// Spread reads accessors and creates data properties on the target.
			return false
		case ast.KindPropertyAssignment:
			name := property.Name()
			if name.Kind != ast.KindComputedPropertyName && ast.GetPropertyNameForPropertyNameNode(name) == "__proto__" {
				return false
			}
		case ast.KindMethodDeclaration:
			// Moving a method changes its home object, including super captured
			// by arrows or used in parameter initializers.
			if property.SubtreeFacts()&ast.SubtreeContainsLexicalSuper != 0 {
				return false
			}
		}
	}
	return true
}

func checkObjectAssign(ctx rule.RuleContext, node *ast.Node) {
	properties := node.AsObjectLiteralExpression().Properties.Nodes
	if len(properties) == 0 {
		return
	}
	for _, property := range properties {
		if property.Kind != ast.KindSpreadAssignment ||
			utils.ESTreeRuntimeExpression(property.Expression()).Kind == ast.KindObjectLiteralExpression {
			return
		}
	}
	parent := utils.ESTreeParent(node)
	call, ok := unicornutil.MatchDotMethodCall(parent, unicornutil.DotMethodCallOptions{Method: "assign"})
	if !ok || !isIdentifier(call.Object, "Object") || len(parent.Arguments()) < 2 {
		return
	}
	hasTarget := false
	replacement := node
	for _, argument := range parent.Arguments() {
		if utils.ESTreeRuntimeExpression(argument) == node {
			// Multiple sources must become separate arguments, not a grouped
			// comma expression. Single sources can retain their parentheses.
			if len(properties) > 1 {
				replacement = argument
			}
			break
		}
		hasTarget = hasTarget || argument.Kind != ast.KindSpreadElement
	}
	if !hasTarget {
		return
	}
	start := utils.TrimNodeTextRange(ctx.SourceFile, properties[0]).Pos()
	ctx.ReportRangeWithDeferredSuggestions(core.NewTextRange(start, start+3), rule.RuleMessage{
		Id:          "spread-in-object-assign",
		Description: "`Object.assign(…)` source object with only spread properties is unnecessary.",
	}, func() []rule.RuleSuggestion {
		objectRange := utils.TrimNodeTextRange(ctx.SourceFile, replacement)
		if utils.HasCommentInSpan(ctx.Comments.All(), objectRange.Pos(), objectRange.End()) {
			return nil
		}
		arguments := make([]string, 0, len(properties))
		for _, property := range properties {
			arguments = append(arguments, utils.TrimmedNodeText(ctx.SourceFile, property.Expression()))
		}
		return []rule.RuleSuggestion{{
			Message:  rule.RuleMessage{Id: "suggestion/remove-object-assign-spread", Description: "Remove the object literal wrapper."},
			FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, replacement, strings.Join(arguments, ", "))},
		}}
	})
}

func isIdentifier(node *ast.Node, name string) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && ast.IsIdentifier(node) && node.Text() == name
}

func checkIterableConversion(ctx rule.RuleContext, node *ast.Node) {
	parent := utils.ESTreeParent(node)
	if parent == nil {
		return
	}
	message := rule.RuleMessage{}
	switch {
	case parent.Kind == ast.KindForOfStatement && utils.ESTreeRuntimeExpression(parent.AsForInOrOfStatement().Expression) == node:
		message.Id = "iterable-to-array-in-for-of"
		message.Description = "`for…of` can iterate directly when an array snapshot is not needed."
	case parent.Kind == ast.KindYieldExpression && parent.AsYieldExpression().AsteriskToken != nil:
		message.Id = "iterable-to-array-in-yield-star"
		message.Description = "`yield*` can delegate directly when materializing the iterable is unnecessary."
	default:
		description := iterableConsumer(parent)
		if description == "" || utils.ESTreeRuntimeExpression(parent.Arguments()[0]) != node {
			return
		}
		message.Id = "iterable-to-array"
		message.Description = "`" + description + "` accepts an iterable as an argument, it's unnecessary to convert to an array."
		message.Data = map[string]string{"parentDescription": description}
	}
	if parent.Kind == ast.KindForOfStatement || parent.Kind == ast.KindYieldExpression {
		// Iterating directly can change snapshots, side-effect timing and the
		// generator protocol. Leave that choice to an explicit suggestion.
		ctx.ReportNodeWithDeferredSuggestions(node, message, func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{{
				Message: rule.RuleMessage{
					Id: "suggestion/remove-iterable-to-array", Description: "Use the iterable directly.",
				},
				FixesArr: unwrapArray(ctx, node),
			}}
		})
		return
	}
	ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
		return unwrapArray(ctx, node)
	})
}

func iterableConsumer(node *ast.Node) string {
	if name := collectionConstructor(node); name != "" && node.Arguments()[0].Kind != ast.KindSpreadElement {
		return "new " + name + "(…)"
	}
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		ArgumentsLength: new(1), RejectSpreadElement: true,
	})
	if !ok {
		return ""
	}
	object := utils.ESTreeRuntimeExpression(call.Object)
	if !ast.IsIdentifier(object) {
		return ""
	}
	name, method := object.Text(), call.Property.Text()
	if (name == "Promise" && slices.Contains([]string{"all", "allSettled", "any", "race"}, method)) ||
		(name == "Object" && method == "fromEntries") ||
		(method == "from" && (name == "Array" || unicornutil.IsTypedArrayName(name))) {
		return name + "." + method + "(…)"
	}
	return ""
}

func checkArrayClone(ctx rule.RuleContext, array, argument *ast.Node) {
	// ESTree exposes a complete optional chain as a ChainExpression. tsgo
	// marks every continuation; parentheses ending the chain clear that flag.
	if ast.IsOptionalChain(argument) {
		return
	}
	call, isMethod := unicornutil.MatchDotMethodCall(argument, unicornutil.DotMethodCallOptions{})
	method := ""
	if isMethod {
		method = call.Property.Text()
	}
	cloneMethod := slices.Contains(cloneMethods, method)
	knownClone := isMethod && (cloneMethod || method == "filter" || method == "flatMap" || method == "map") &&
		unicornutil.IsArray(ctx, call.Object) && unicornutil.IsArray(ctx, argument)
	if !knownClone && !isHeuristicClone(ctx, argument, call, isMethod, cloneMethod) {
		return
	}
	ctx.ReportNodeWithDeferredFixes(array, rule.RuleMessage{
		Id: "clone-array", Description: "Unnecessarily cloning an array.",
	}, func() []rule.RuleFix {
		// Array constructors can create holes. A slice of an unknown receiver
		// might still need conversion from a string or a typed array.
		if constructorName(argument) == "Array" || (method == "slice" && !knownClone) {
			return nil
		}
		return unwrapArray(ctx, array)
	})
}

func isHeuristicClone(ctx rule.RuleContext, node *ast.Node, call unicornutil.DotMethodCall, isMethod, cloneMethod bool) bool {
	matched := false
	if isMethod {
		method := call.Property.Text()
		switch {
		case cloneMethod || method == "concat":
			matched = (method != "concat" || !isIdentifier(call.Object, "Iterator")) &&
				!unicornutil.IsKnownNonArray(ctx, call.Object)
		case method == "split" || (isIdentifier(call.Object, "Array") && (method == "from" || method == "of")):
			matched = true
		case isIdentifier(call.Object, "Object") && (method == "keys" || method == "values"):
			matched = len(node.Arguments()) == 1 && node.Arguments()[0].Kind != ast.KindSpreadElement
		}
	} else if node.Kind == ast.KindAwaitExpression {
		call, ok := unicornutil.MatchDotMethodCall(utils.ESTreeRuntimeExpression(node.Expression()), unicornutil.DotMethodCallOptions{
			Methods: []string{"all", "allSettled"}, ArgumentsLength: new(1), RejectSpreadElement: true,
		})
		matched = ok && isIdentifier(call.Object, "Promise")
	} else {
		matched = constructorName(node) == "Array"
	}
	return matched && !unicornutil.IsKnownNonArray(ctx, node)
}

func flattenLiteral(ctx rule.RuleContext, node, spread *ast.Node) []rule.RuleFix {
	source := ctx.SourceFile
	start := utils.TrimNodeTextRange(source, spread).Pos()
	fixes := []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, start+3))}
	for outer := node; outer.Parent != nil && utils.ESTreeRuntimeExpression(outer.Parent) == node; outer = outer.Parent {
		if ast.IsParenthesizedExpression(outer.Parent) {
			r := utils.TrimNodeTextRange(source, outer.Parent)
			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(r.Pos(), r.Pos()+1)),
				rule.RuleFixRemoveRange(core.NewTextRange(r.End()-1, r.End())))
		}
	}
	r := utils.TrimNodeTextRange(source, node)
	fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(r.Pos(), r.Pos()+1)),
		rule.RuleFixRemoveRange(core.NewTextRange(r.End()-1, r.End())))
	last, _ := utils.TokenBeforePosition(source, r.End()-1)
	trailingComma := last.Kind == ast.KindCommaToken
	empty := false
	if node.Kind == ast.KindArrayLiteralExpression {
		elements := node.AsArrayLiteralExpression().Elements.Nodes
		empty = len(elements) == 0
		if ast.IsCallExpression(spread.Parent) || ast.IsNewExpression(spread.Parent) {
			for _, element := range elements {
				if element.Kind != ast.KindOmittedExpression {
					continue
				}
				comma, _ := utils.TokenAtOrAfter(source, element.Pos())
				if trailingComma && comma.Start == last.Start {
					fixes = append(fixes, rule.RuleFixReplaceRange(comma.Range(), "undefined"))
					trailingComma = false
				} else {
					fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(comma.Start, comma.Start), "undefined"))
				}
			}
		}
	} else {
		empty = len(node.AsObjectLiteralExpression().Properties.Nodes) == 0
	}
	if trailingComma {
		fixes = append(fixes, rule.RuleFixRemoveRange(last.Range()))
	}
	if empty {
		if next, ok := utils.TokenAtOrAfter(source, spread.End()); ok && next.Kind == ast.KindCommaToken {
			fixes = append(fixes, rule.RuleFixRemoveRange(next.Range()))
		}
	}
	return fixes
}

func unwrapArray(ctx rule.RuleContext, node *ast.Node) []rule.RuleFix {
	source := ctx.SourceFile
	r := utils.TrimNodeTextRange(source, node)
	spread := node.AsArrayLiteralExpression().Elements.Nodes[0]
	spreadStart := utils.TrimNodeTextRange(source, spread).Pos()
	argument := spread.Expression()
	opening, closing := "", ""
	if !strongPrecedence(argument) && !strongPrecedence(utils.SkipAssertionsAndParens(argument)) {
		opening, closing = "(", ")"
	}
	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(r.Pos(), r.Pos()+1), opening),
		rule.RuleFixRemoveRange(core.NewTextRange(spreadStart, spreadStart+3)),
		rule.RuleFixReplaceRange(core.NewTextRange(r.End()-1, r.End()), closing),
	}
	if last, ok := utils.TokenBeforePosition(source, r.End()-1); ok && last.Kind == ast.KindCommaToken {
		fixes = append(fixes, rule.RuleFixRemoveRange(last.Range()))
	}
	parent := utils.ESTreeParent(node)
	third, _ := utils.TokenAtOrAfter(source, spreadStart+3)
	if parent != nil && (parent.Kind == ast.KindReturnStatement || parent.Kind == ast.KindThrowStatement) &&
		utils.OutermostParenthesizedExpression(node) == node &&
		strings.ContainsAny(source.Text()[r.Pos()+1:third.Start], "\r\n\u2028\u2029") {
		// Preserve ASI when deleting the opening bracket exposes a line break.
		keyword, _ := utils.TokenAtOrAfter(source, parent.Pos())
		fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(keyword.End, keyword.End), " ("))
		last, _ := utils.TokenBeforePosition(source, parent.End())
		end := last.End
		if last.Kind == ast.KindSemicolonToken {
			end = last.Start
		}
		return append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(end, end), ")"))
	}
	return append(fixes, unicornutil.SpaceAroundKeywordFixes(source, node)...)
}

func strongPrecedence(node *ast.Node) bool {
	// Match Unicorn's narrower precedence set while reusing the common cases.
	// In particular, object literals and TS instantiations retain parentheses.
	if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindExpressionWithTypeArguments {
		return false
	}
	switch node.Kind {
	case ast.KindNullKeyword, ast.KindThisKeyword, ast.KindSuperKeyword,
		ast.KindTemplateExpression, ast.KindNonNullExpression:
		return true
	default:
		return utils.IsStrongPrecedenceNode(node)
	}
}
