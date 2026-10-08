// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
// cspell:ignore callbag
package no_for_each

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const (
	messageIDError      = "no-for-each/error"
	messageIDSuggestion = "no-for-each/suggestion"
)

var messageError = rule.RuleMessage{
	Id:          messageIDError,
	Description: "Use `for…of` instead of `.forEach(…)`.",
}

var messageSuggestion = rule.RuleMessage{
	Id:          messageIDSuggestion,
	Description: "Switch to `for…of`.",
}

var ignoredObjects = []string{
	"React.Children",
	"Children",
	"R",
	"pIteration",
	"Effect",
}

var NoForEachRule = rule.Rule{
	Name:   "unicorn/no-for-each",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		staticEvaluator := utils.NewStaticStringEvaluatorWithReferenceResolver(
			ctx.TypeChecker,
			ctx.SourceFile,
			ctx.Refs,
		)

		return rule.RuleListeners{
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Method:              "forEach",
					AllowOptionalCall:   true,
					AllowOptionalMember: true,
				})
				if !ok || isIgnoredObject(call.Object) ||
					isStrictCallbagBasicsNamespace(ctx, call.Object) ||
					unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, call.Object) {
					return
				}

				if !unicornutil.IsArray(ctx, call.Object) {
					ctx.ReportNode(call.Property, messageError)
					return
				}

				plan, ok := buildFixPlan(ctx, call)
				if !ok {
					ctx.ReportNode(call.Property, messageError)
					return
				}

				optionalObject := ast.IsOptionalChainRoot(call.Callee)
				if optionalObject && staticEvaluator.HasSideEffect(call.Object, false) {
					ctx.ReportNodeWithDeferredSuggestions(call.Property, messageError, func() []rule.RuleSuggestion {
						fixes := plan.fixes()
						if len(fixes) == 0 {
							return nil
						}
						return []rule.RuleSuggestion{{Message: messageSuggestion, FixesArr: fixes}}
					})
					return
				}

				ctx.ReportNodeWithDeferredFixes(call.Property, messageError, plan.fixes)
			},
		}
	},
}

func isIgnoredObject(node *ast.Node) bool {
	for _, object := range ignoredObjects {
		if unicornutil.NodeMatchesPath(node, object) {
			return true
		}
	}
	return false
}

func isStrictCallbagBasicsNamespace(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !ast.IsIdentifier(node) || ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindNamespaceImport ||
			declaration.Parent == nil || declaration.Parent.Parent == nil ||
			declaration.Parent.Parent.Kind != ast.KindImportDeclaration {
			continue
		}
		importDeclaration := declaration.Parent.Parent.AsImportDeclaration()
		if importDeclaration.ModuleSpecifier != nil &&
			importDeclaration.ModuleSpecifier.Text() == "strict-callbag-basics" {
			return true
		}
	}
	return false
}

type fixPlan struct {
	replaceRange core.TextRange
	replacement  string
}

func (plan fixPlan) fixes() []rule.RuleFix {
	if plan.replacement == "" {
		return nil
	}
	return []rule.RuleFix{rule.RuleFixReplaceRange(plan.replaceRange, plan.replacement)}
}

func buildFixPlan(ctx rule.RuleContext, call unicornutil.DotMethodCall) (fixPlan, bool) {
	// Optional calls (`array.forEach?.(fn)`) do not have an equivalent direct
	// `for…of` rewrite. Optional member access is handled below.
	if ast.IsOptionalChainRoot(call.Call) || len(call.Call.Arguments()) != 1 {
		return fixPlan{}, false
	}

	callback := utils.ESTreeRuntimeExpression(call.Call.Arguments()[0])
	if callback == nil || (!ast.IsArrowFunction(callback) && !ast.IsFunctionExpression(callback)) ||
		ast.HasSyntacticModifier(callback, ast.ModifierFlagsAsync) {
		return fixPlan{}, false
	}
	if ast.IsFunctionExpression(callback) {
		function := callback.AsFunctionExpression()
		if function.AsteriskToken != nil || function.Name() != nil {
			return fixPlan{}, false
		}
	}

	parameters := callback.Parameters()
	if len(parameters) < 1 || len(parameters) > 2 {
		return fixPlan{}, false
	}
	for index, parameter := range parameters {
		if parameter == nil || parameter.Kind != ast.KindParameter {
			return fixPlan{}, false
		}
		data := parameter.AsParameterDeclaration()
		if data == nil || data.DotDotDotToken != nil || data.Type != nil ||
			(index == 0 && data.Initializer != nil) ||
			(index == 1 && (data.Initializer != nil || !ast.IsIdentifier(data.Name()))) {
			return fixPlan{}, false
		}
	}

	outerCall := outermostRuntimeParentheses(call.Call)
	ancestor := outerCall.Parent
	arrowBody := false
	var replaceRange core.TextRange
	var statement *ast.Node
	if ancestor != nil && ast.IsExpressionStatement(ancestor) {
		statement = ancestor
		replaceRange = utils.TrimNodeTextRange(ctx.SourceFile, ancestor)
	} else if ancestor != nil && ast.IsArrowFunction(ancestor) &&
		ancestor.AsArrowFunction().Body == outerCall {
		arrowBody = true
		replaceRange = utils.TrimNodeTextRange(ctx.SourceFile, outerCall)
	} else {
		return fixPlan{}, false
	}

	body := functionBody(callback)
	returns, returnsSafe := callbackReturnStatements(callback, body)
	if body == nil || !returnsSafe || hasUnsafeCallbackSemantics(callback, body) ||
		!parametersSafeForFix(ctx, call.Call, callback, parameters) {
		return fixPlan{}, false
	}

	loop, ok := buildLoopText(ctx, call, callback, parameters, body, returns, statement)
	if !ok {
		return fixPlan{}, false
	}
	if arrowBody {
		loop = "{ " + loop + " }"
	}

	return fixPlan{replaceRange: replaceRange, replacement: loop}, true
}

func outermostRuntimeParentheses(node *ast.Node) *ast.Node {
	outer := utils.OutermostParenthesizedExpression(node)
	for outer != nil && outer.Parent != nil && utils.IsJSDocTypeCastWrapper(outer.Parent) {
		outer = utils.OutermostParenthesizedExpression(outer.Parent)
	}
	return outer
}

func functionBody(function *ast.Node) *ast.Node {
	switch function.Kind {
	case ast.KindArrowFunction:
		return function.AsArrowFunction().Body
	case ast.KindFunctionExpression:
		return function.AsFunctionExpression().Body
	}
	return nil
}

func hasUnsafeCallbackSemantics(callback, body *ast.Node) bool {
	unsafe := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || unsafe {
			return
		}
		// Arrow functions capture the callback function's `this` and
		// `arguments`; ordinary nested functions create their own bindings.
		if node != callback && ast.IsFunctionLike(node) && !ast.IsArrowFunction(node) {
			return
		}
		if ast.IsFunctionExpression(callback) {
			if node.Kind == ast.KindThisKeyword || node.Kind == ast.KindMetaProperty ||
				(ast.IsIdentifier(node) && node.Text() == "arguments" && !utils.IsNonReferenceIdentifier(node)) {
				unsafe = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return unsafe
		})
	}
	walk(body)
	return unsafe
}

func callbackReturnStatements(callback, body *ast.Node) ([]*ast.Node, bool) {
	if body == nil {
		return nil, false
	}
	var returns []*ast.Node
	safe := true
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || !safe {
			return
		}
		if node != callback && ast.IsFunctionLike(node) {
			return
		}
		if node.Kind == ast.KindReturnStatement {
			for ancestor := node.Parent; ancestor != nil && ancestor != callback; ancestor = ancestor.Parent {
				switch ancestor.Kind {
				case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
					ast.KindForInStatement, ast.KindForOfStatement:
					safe = false
					return
				}
			}
			returns = append(returns, node)
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return !safe
		})
	}
	walk(body)
	return returns, safe
}

func parametersSafeForFix(ctx rule.RuleContext, call, callback *ast.Node, parameters []*ast.Node) bool {
	parameterNames := map[string]struct{}{}
	parameterSymbols := map[*ast.Symbol]struct{}{}
	for _, parameter := range parameters {
		duplicate := false
		utils.CollectBindingNames(parameter.Name(), func(identifier *ast.Node, name string) {
			if _, exists := parameterNames[name]; exists {
				duplicate = true
			}
			parameterNames[name] = struct{}{}
			if ctx.Refs != nil {
				symbol := identifier.Symbol()
				if symbol == nil {
					symbol = ctx.Refs.ResolveInFile(identifier)
				}
				if symbol != nil {
					parameterSymbols[symbol] = struct{}{}
				}
			}
		})
		if duplicate {
			return false
		}
	}

	// A parameter name used by the receiver or another callback parameter
	// would resolve to a different binding after the callback boundary is
	// removed. The callback body itself is safe because its references move
	// together with the declaration into the loop.
	unsafe := false
	var walkOutsideCallback func(*ast.Node)
	walkOutsideCallback = func(node *ast.Node) {
		if node == nil || unsafe || node == callback {
			return
		}
		if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) {
			if _, conflicts := parameterNames[node.Text()]; conflicts {
				unsafe = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walkOutsideCallback(child)
			return unsafe
		})
	}
	walkOutsideCallback(call)
	if unsafe {
		return false
	}

	// Duplicate/sloppy bindings and unresolved parameter symbols are unsafe to
	// move. When references are available, every parameter binding must be the
	// only declaration represented by its symbol.
	for symbol := range parameterSymbols {
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
	}
	return true
}

func buildLoopText(
	ctx rule.RuleContext,
	call unicornutil.DotMethodCall,
	callback *ast.Node,
	parameters []*ast.Node,
	body *ast.Node,
	returns []*ast.Node,
	statement *ast.Node,
) (string, bool) {
	loopKind := "const"
	if callbackParametersReassigned(ctx, callback, parameters) {
		loopKind = "let"
	}

	parameterTexts := make([]string, len(parameters))
	for index, parameter := range parameters {
		parameterTexts[index] = utils.TrimmedNodeText(ctx.SourceFile, parameter.Name())
	}
	binding := parameterTexts[0]
	if len(parameterTexts) == 2 {
		binding = "[" + parameterTexts[1] + ", " + parameterTexts[0] + "]"
	}

	objectText := utils.TrimmedNodeText(ctx.SourceFile, call.Object)
	iterableText := objectText
	if len(parameterTexts) == 2 {
		if ast.IsParenthesizedExpression(call.Object) ||
			(ast.IsOptionalChainRoot(call.Callee) && startsWithMemberHazard(objectText)) {
			iterableText = "(" + objectText + ")"
		}
		iterableText += ".entries()"
	}

	bodyText := utils.TrimmedNodeText(ctx.SourceFile, body)
	if len(returns) > 0 {
		bodyText = rewriteCallbackReturns(ctx, body, bodyText, returns)
	}
	if body.Kind != ast.KindBlock && shouldParenthesizeExpressionStatement(bodyText) {
		bodyText = "(" + bodyText + ")"
	}

	optionalObject := ast.IsOptionalChainRoot(call.Callee)
	loop := "for (" + loopKind + " " + binding + " of " + iterableText + ") " + bodyText
	if body.Kind != ast.KindBlock && statement != nil &&
		strings.HasSuffix(utils.TrimmedNodeText(ctx.SourceFile, statement), ";") {
		loop += ";"
	}
	if optionalObject {
		loop = "if (" + objectText + ") " + loop
	}

	// Rebuilding the loop head would otherwise drop comments between the call
	// start and the callback body.
	callRange := utils.TrimNodeTextRange(ctx.SourceFile, call.Call)
	bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, body)
	if utils.HasCommentInSpan(ctx.Comments.All(), callRange.Pos(), bodyRange.Pos()) {
		return "", false
	}
	return loop, true
}

func rewriteCallbackReturns(ctx rule.RuleContext, body *ast.Node, bodyText string, returns []*ast.Node) string {
	bodyRange := utils.TrimNodeTextRange(ctx.SourceFile, body)
	for index := len(returns) - 1; index >= 0; index-- {
		returnStatement := returns[index]
		returnRange := utils.TrimNodeTextRange(ctx.SourceFile, returnStatement)
		start := returnRange.Pos() - bodyRange.Pos()
		end := returnRange.End() - bodyRange.Pos()
		if start < 0 || end < start || end > len(bodyText) {
			continue
		}

		data := returnStatement.AsReturnStatement()
		replacement := "continue"
		completeBlockReplacement := false
		if data.Expression != nil {
			expressionRange := utils.TrimNodeTextRange(ctx.SourceFile, data.Expression)
			gapStart := returnRange.Pos() + len("return")
			gap := ""
			if gapStart <= expressionRange.Pos() {
				gap = ctx.SourceFile.Text()[gapStart:expressionRange.Pos()]
			}
			expressionText := utils.TrimmedNodeText(ctx.SourceFile, data.Expression)
			if shouldParenthesizeExpressionStatement(expressionText) {
				expressionText = "(" + expressionText + ")"
			}
			semicolon := ""
			if previous, ok := utils.TokenBeforePosition(ctx.SourceFile, returnRange.Pos()); ok &&
				unicornutil.NeedsSemicolonAfter(ctx.SourceFile, previous, expressionText, false) {
				semicolon = ";"
			}
			replacement = gap + semicolon + expressionText + "; continue"
			if returnNeedsBlock(returnStatement) {
				replacement = gap + "{ " + expressionText + "; continue; }"
				completeBlockReplacement = true
			}
		}
		if !completeBlockReplacement &&
			(strings.HasSuffix(utils.TrimmedNodeText(ctx.SourceFile, returnStatement), ";") || data.Expression != nil) {
			replacement += ";"
		}
		bodyText = bodyText[:start] + replacement + bodyText[end:]
	}
	return bodyText
}

func returnNeedsBlock(returnStatement *ast.Node) bool {
	if returnStatement == nil || returnStatement.Parent == nil {
		return false
	}
	parent := returnStatement.Parent
	switch parent.Kind {
	case ast.KindIfStatement:
		ifStatement := parent.AsIfStatement()
		return ifStatement.ThenStatement == returnStatement || ifStatement.ElseStatement == returnStatement
	case ast.KindWithStatement:
		return parent.AsWithStatement().Statement == returnStatement
	default:
		return false
	}
}

func callbackParametersReassigned(ctx rule.RuleContext, callback *ast.Node, parameters []*ast.Node) bool {
	type binding struct {
		declaration *ast.Node
		name        string
		symbol      *ast.Symbol
	}
	var bindings []binding
	for _, parameter := range parameters {
		utils.CollectBindingNames(parameter.Name(), func(identifier *ast.Node, name string) {
			var symbol *ast.Symbol
			if ctx.TypeChecker != nil {
				symbol = ctx.TypeChecker.GetSymbolAtLocation(identifier)
			}
			bindings = append(bindings, binding{
				declaration: identifier.Parent,
				name:        name,
				symbol:      symbol,
			})
		})
	}

	reassigned := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || reassigned {
			return
		}
		if ast.IsIdentifier(node) && utils.IsWriteReference(node) {
			for _, binding := range bindings {
				if node.Text() != binding.name {
					continue
				}
				if ctx.TypeChecker == nil {
					if !utils.IsNameShadowedBetween(node, callback, binding.name) {
						reassigned = true
					}
					break
				}
				referenceSymbol := utils.GetReferenceSymbol(node, ctx.TypeChecker)
				if referenceSymbol == binding.symbol ||
					(referenceSymbol != nil && referenceSymbol.ValueDeclaration == binding.declaration) {
					reassigned = true
					break
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return reassigned
		})
	}
	walk(functionBody(callback))

	return reassigned
}

func startsWithMemberHazard(text string) bool {
	if text == "" {
		return false
	}
	first := text[0]
	return first >= '0' && first <= '9'
}

func shouldParenthesizeExpressionStatement(text string) bool {
	trimmed := ecmascript.StringTrim(text)
	if trimmed == "" || strings.HasPrefix(trimmed, "(") {
		return false
	}
	return strings.HasPrefix(trimmed, "class") ||
		strings.HasPrefix(trimmed, "function") ||
		strings.HasPrefix(trimmed, "async function") ||
		strings.HasPrefix(trimmed, "{")
}
