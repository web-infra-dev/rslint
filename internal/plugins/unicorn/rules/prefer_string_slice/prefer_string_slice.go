package prefer_string_slice

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var (
	substrMessage    = rule.RuleMessage{Id: "substr", Description: "Prefer `String#slice()` over `String#substr()`."}
	substringMessage = rule.RuleMessage{Id: "substring", Description: "Prefer `String#slice()` over `String#substring()`."}
)

// PreferStringSliceRule follows eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-string-slice.js
var PreferStringSliceRule = rule.Rule{
	Name:   "unicorn/prefer-string-slice",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
				Methods: []string{"substr", "substring"}, AllowOptionalCall: true, AllowOptionalMember: true,
			})
			if !ok || isKnownNonStringReceiver(ctx, call.Object) {
				return
			}
			method := call.Property.Text()
			if method == "substring" && unicornutil.SingleCharacterSubstringIndex(node.Arguments()) != nil {
				return
			}
			message := substrMessage
			if method == "substring" {
				message = substringMessage
			}
			ctx.ReportNodeWithDeferredFixes(node, message, func() []rule.RuleFix {
				fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, call.Property, "slice")}
				args := node.Arguments()
				if len(args) == 0 {
					return fixes
				}
				if len(args) > 2 {
					return nil
				}
				for _, arg := range args {
					if arg.Kind == ast.KindSpreadElement {
						return nil
					}
				}
				if evaluator == nil {
					evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
					evaluator.GlobalAccess = ctx.Globals.Access
				}

				// Replacing an argument includes its parentheses; abort the whole
				// fix if that replacement would discard a comment.
				aborted := false
				replace := func(arg *ast.Node, text string) {
					// Numeric serialization can introduce an identifier even when
					// the original argument was a literal such as 1e400.
					global := ""
					switch text {
					case "Infinity", "-Infinity":
						global = "Infinity"
					case "NaN":
						global = "NaN"
					}
					if global != "" && !availableGlobals(ctx, node, global) {
						aborted = true
						return
					}
					span := utils.TrimNodeTextRange(ctx.SourceFile, arg)
					if utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
						aborted = true
						return
					}
					fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, arg, text))
				}
				if method == "substr" {
					if !fixSubstrArguments(ctx, call, evaluator, replace, &fixes) {
						return nil
					}
				} else if !fixSubstringArguments(ctx, call, evaluator, replace, &fixes) {
					return nil
				}
				if aborted {
					return nil
				}
				return fixes
			})
		}}
	},
}

func fixSubstrArguments(ctx rule.RuleContext, call unicornutil.DotMethodCall, evaluator *utils.StaticStringEvaluator, replace func(*ast.Node, string), fixes *[]rule.RuleFix) bool {
	args := call.Call.Arguments()
	if len(args) == 1 {
		return true
	}
	first, second := utils.ESTreeRuntimeExpression(args[0]), utils.ESTreeRuntimeExpression(args[1])
	if value, known := evaluator.EvalValueIfNoSideEffects(first); known && value == jsnum.Number(0) {
		if negative := negativeIndex(ctx, args[1], call.Object); negative != "" {
			replace(args[1], negative)
			return true
		}
		if second.Kind == ast.KindNumericLiteral || isLengthProperty(second) {
			return true
		}
		if number, known := numericValue(second); known {
			replace(args[1], ecmascript.NumberToString(math.Max(0, number)))
			return true
		}
		text, ok := coalescedEndText(ctx, args[1], evaluator)
		if !ok {
			return false
		}
		// The original expression, including its comments, is retained twice.
		*fixes = append(*fixes, rule.RuleFixReplace(ctx.SourceFile, args[1], "Math.max(0, "+text+")"))
		return true
	}
	if first.Kind != ast.KindNumericLiteral || second.Kind != ast.KindNumericLiteral {
		return false
	}
	// substr truncates each argument before adding the start and length.
	start, _ := numericValue(first)
	length, _ := numericValue(second)
	replace(args[1], ecmascript.NumberToString(math.Trunc(start)+math.Trunc(length)))
	return true
}

func fixSubstringArguments(ctx rule.RuleContext, call unicornutil.DotMethodCall, evaluator *utils.StaticStringEvaluator, replace func(*ast.Node, string), fixes *[]rule.RuleFix) bool {
	args := call.Call.Arguments()
	first := utils.ESTreeRuntimeExpression(args[0])
	start, knownStart := numericValue(first)
	if len(args) == 1 {
		if isLengthProperty(first) {
			return true
		}
		if knownStart {
			replace(args[0], ecmascript.NumberToString(math.Max(0, start)))
		} else {
			if !availableGlobals(ctx, call.Call, "Math") {
				return false
			}
			// Wrapping preserves comments inside the existing parentheses.
			*fixes = append(*fixes, rule.RuleFixReplace(ctx.SourceFile, args[0], "Math.max(0, "+utils.TrimmedNodeText(ctx.SourceFile, args[0])+")"))
		}
		return true
	}
	end, knownEnd := numericValue(utils.ESTreeRuntimeExpression(args[1]))
	if knownStart && knownEnd {
		newStart, newEnd := math.Max(0, start), math.Max(0, end)
		if start > end {
			newStart, newEnd = newEnd, newStart
		}
		if newStart != start {
			replace(args[0], ecmascript.NumberToString(newStart))
		}
		if newEnd != end {
			replace(args[1], ecmascript.NumberToString(newEnd))
		}
		return true
	}
	if (knownStart && start == 0) || (knownEnd && end == 0) {
		if !availableGlobals(ctx, call.Call, "Math") {
			return false
		}
		var text string
		if knownEnd && end == 0 {
			// A moved start is evaluated once and keeps undefined -> 0.
			text = utils.TrimmedNodeText(ctx.SourceFile, args[0])
		} else {
			var ok bool
			text, ok = coalescedEndText(ctx, args[1], evaluator)
			if !ok {
				return false
			}
		}
		replace(args[0], "0")
		replace(args[1], "Math.max(0, "+text+")")
		return true
	}
	return false
}

func coalescedEndText(ctx rule.RuleContext, arg *ast.Node, evaluator *utils.StaticStringEvaluator) (string, bool) {
	if !availableGlobals(ctx, arg.Parent, "Math", "undefined", "Infinity") ||
		evaluator.HasSideEffect(arg, true) || containsSpreadOrTag(arg) || !repeatableEnd(arg, evaluator) {
		return "", false
	}
	text := utils.TrimmedNodeText(ctx.SourceFile, arg)
	node := utils.ESTreeRuntimeExpression(arg)
	if arg == node && ast.GetExpressionPrecedence(node) < ast.OperatorPrecedenceAdditive {
		text = "(" + text + ")"
	}
	// Missing/undefined ends mean "to the end"; null still means zero.
	return text + " === undefined ? Infinity : " + text, true
}

func availableGlobals(ctx rule.RuleContext, location *ast.Node, names ...string) bool {
	for _, name := range names {
		if !ctx.Globals.Access(name).IsDeclared() ||
			(ctx.Refs != nil && ctx.Refs.IsNameDefinedInFile(location, name)) ||
			(ctx.Refs == nil && utils.IsShadowed(location, name)) {
			return false
		}
	}
	return true
}

// The conditional reads its operand again only on the non-undefined branch.
// Avoid getters and coercions such as +value that could execute user code on
// both reads. Plain values are coerced only once, by the outer Math.max call.
func repeatableEnd(node *ast.Node, evaluator *utils.StaticStringEvaluator) bool {
	node = ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments)
	switch node.Kind {
	case ast.KindIdentifier, ast.KindThisKeyword, ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	case ast.KindTypeOfExpression:
		return repeatableEnd(node.Expression(), evaluator)
	case ast.KindPrefixUnaryExpression:
		if node.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken {
			return repeatableEnd(node.AsPrefixUnaryExpression().Operand, evaluator)
		}
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken,
			ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken, ast.KindCommaToken:
			return repeatableEnd(binary.Left, evaluator) && repeatableEnd(binary.Right, evaluator)
		}
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return repeatableEnd(conditional.Condition, evaluator) && repeatableEnd(conditional.WhenTrue, evaluator) && repeatableEnd(conditional.WhenFalse, evaluator)
	}
	return unicornutil.IsStaticPrimitiveArgument(evaluator, node)
}

func isKnownNonStringReceiver(ctx rule.RuleContext, node *ast.Node) bool {
	if unicornutil.IsKnownNonStringSyntax(ctx, node) ||
		(ctx.TypeChecker != nil && unicornutil.IsKnownNonString(ctx, node)) {
		return true
	}
	node = utils.ESTreeRuntimeExpression(node)
	node = ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKNonNullAssertions)
	if ast.IsCallExpression(node) && ast.IsOptionalChain(node) {
		// Optional invocation does not change an authored return annotation.
		// Classify a clone so the source tree and its chain boundaries stay intact.
		ordinary := node.Clone(ast.NewNodeFactory(ast.NodeFactoryHooks{}))
		ordinary.Flags &^= ast.NodeFlagsOptionalChain
		return unicornutil.IsKnownNonStringSyntax(ctx, ordinary)
	}
	if !ast.IsIdentifier(node) || ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind == ast.KindVariableDeclaration {
		value := utils.ESTreeRuntimeExpression(declaration.AsVariableDeclaration().Initializer)
		for value != nil && (value.Kind == ast.KindNonNullExpression || value.Kind == ast.KindSatisfiesExpression) {
			value = utils.ESTreeRuntimeExpression(value.Expression())
		}
		if value != nil && (value.Kind == ast.KindAsExpression || value.Kind == ast.KindTypeAssertionExpression) {
			return false
		}
	}
	annotation, _ := unicornutil.ValueTypeAnnotation(ctx, declaration)
	return annotation != nil && utils.IsJSDocSyntaxNode(annotation) && nonStringAnnotation(ctx.SourceFile, annotation)
}

func nonStringAnnotation(sourceFile *ast.SourceFile, node *ast.Node) bool {
	switch node.Kind {
	case ast.KindParenthesizedType, ast.KindJSDocNonNullableType, ast.KindJSDocNullableType, ast.KindJSDocOptionalType:
		return nonStringAnnotation(sourceFile, node.Type())
	case ast.KindUnionType, ast.KindIntersectionType:
		var parts []*ast.Node
		if node.Kind == ast.KindIntersectionType {
			parts = node.AsIntersectionTypeNode().Types.Nodes
		} else {
			parts = node.AsUnionTypeNode().Types.Nodes
		}
		for _, part := range parts {
			if !nonStringAnnotation(sourceFile, part) {
				return false
			}
		}
		return true
	case ast.KindLiteralType:
		return !ast.IsStringLiteralLike(node.AsLiteralTypeNode().Literal)
	case ast.KindTypeReference:
		// tsgo retains Closure-style function syntax as a reference to "function".
		name := node.AsTypeReferenceNode().TypeName
		return ast.IsIdentifier(name) && name.Text() == "function" &&
			scanner.ScanTokenAtPosition(sourceFile, node.End()) == ast.KindOpenParenToken
	case ast.KindNumberKeyword, ast.KindBooleanKeyword, ast.KindBigIntKeyword, ast.KindSymbolKeyword,
		ast.KindObjectKeyword, ast.KindVoidKeyword, ast.KindUndefinedKeyword, ast.KindNullKeyword,
		ast.KindArrayType, ast.KindTupleType, ast.KindTypeLiteral, ast.KindJSDocTypeLiteral,
		ast.KindFunctionType, ast.KindConstructorType:
		return true
	}
	return false
}

func containsSpreadOrTag(node *ast.Node) bool {
	if node.Kind == ast.KindSpreadElement || node.Kind == ast.KindSpreadAssignment || node.Kind == ast.KindTaggedTemplateExpression {
		return true
	}
	return node.ForEachChild(containsSpreadOrTag)
}

func numericValue(node *ast.Node) (float64, bool) {
	node = utils.ESTreeRuntimeExpression(node)
	if node.Kind == ast.KindNumericLiteral {
		return float64(jsnum.FromString(node.Text())), true
	}
	if node.Kind == ast.KindPrefixUnaryExpression && node.AsPrefixUnaryExpression().Operator == ast.KindMinusToken {
		value, known := numericValue(node.AsPrefixUnaryExpression().Operand)
		return -value, known
	}
	return 0, false
}

func isLengthProperty(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return ast.IsPropertyAccessExpression(node) && !ast.IsOptionalChain(node) &&
		ast.IsIdentifier(node.Name()) && node.Name().Text() == "length"
}

func negativeIndex(ctx rule.RuleContext, arg, receiver *ast.Node) string {
	node := utils.ESTreeRuntimeExpression(arg)
	receiver = utils.ESTreeRuntimeExpression(receiver)
	if node.Kind != ast.KindBinaryExpression {
		return ""
	}
	binary := node.AsBinaryExpression()
	right := utils.ESTreeRuntimeExpression(binary.Right)
	value, isInteger := unicornutil.SafeIntegerLiteral(right)
	if binary.OperatorToken.Kind != ast.KindMinusToken || !isInteger || value == 0 || !isLengthProperty(binary.Left) {
		return ""
	}
	_, staticString := utils.GetStaticStringLiteralValue(receiver)
	if (!ast.IsIdentifier(receiver) && !staticString) || !unicornutil.IsSameComparisonOperand(utils.ESTreeRuntimeExpression(binary.Left).Expression(), receiver) {
		return ""
	}
	// arg.Pos starts after the previous token, before its leading comments.
	if utils.HasCommentInSpan(ctx.Comments.All(), arg.Pos(), arg.End()) {
		return ""
	}
	return "-" + utils.TrimmedNodeText(ctx.SourceFile, right)
}
