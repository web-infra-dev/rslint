// Ported from eslint-plugin-unicorn v77.0.0; see LICENSE.
// cspell:ignore callbag
package no_for_each

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
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
		var calls []unicornutil.DotMethodCall
		staticEvaluator := utils.NewStaticStringEvaluatorWithReferenceResolver(
			ctx.TypeChecker,
			ctx.SourceFile,
			ctx.Refs,
		)

		reportCall := func(call unicornutil.DotMethodCall) {
			optionalObject := ast.IsOptionalChainRoot(call.Callee)
			if optionalObject && staticEvaluator.HasSideEffect(call.Object, false) {
				ctx.ReportNodeWithDeferredSuggestions(call.Property, messageError, func() []rule.RuleSuggestion {
					fixes := buildFixes(ctx, call)
					if len(fixes) == 0 {
						return nil
					}
					return []rule.RuleSuggestion{{Message: messageSuggestion, FixesArr: fixes}}
				})
				return
			}

			ctx.ReportNodeWithDeferredFixes(call.Property, messageError, func() []rule.RuleFix {
				return buildFixes(ctx, call)
			})
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
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
				calls = append(calls, call)
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				for _, call := range calls {
					reportCall(call)
				}
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

func buildFixes(ctx rule.RuleContext, call unicornutil.DotMethodCall) []rule.RuleFix {
	if !unicornutil.IsArray(ctx, call.Object) {
		return nil
	}
	// Optional calls (`array.forEach?.(fn)`) do not have an equivalent direct
	// `for…of` rewrite. Optional member access is handled below.
	if ast.IsOptionalChainRoot(call.Call) || len(call.Call.Arguments()) != 1 {
		return nil
	}

	callback := utils.ESTreeRuntimeExpression(call.Call.Arguments()[0])
	if callback == nil || (!ast.IsArrowFunction(callback) && !ast.IsFunctionExpression(callback)) ||
		ast.HasSyntacticModifier(callback, ast.ModifierFlagsAsync) {
		return nil
	}
	if ast.IsFunctionExpression(callback) {
		function := callback.AsFunctionExpression()
		if function.AsteriskToken != nil {
			return nil
		}
	}

	parameters := callback.Parameters()
	if len(parameters) < 1 || len(parameters) > 2 {
		return nil
	}
	for index, parameter := range parameters {
		if parameter == nil || parameter.Kind != ast.KindParameter {
			return nil
		}
		data := parameter.AsParameterDeclaration()
		if data == nil || data.DotDotDotToken != nil || data.Type != nil ||
			(index == 0 && data.Initializer != nil) ||
			(index == 1 && (data.Initializer != nil || !ast.IsIdentifier(data.Name()))) {
			return nil
		}
	}

	outerCall := outermostRuntimeParentheses(call.Call)
	ancestor := outerCall.Parent
	arrowBody := false
	var replaceRange = utils.TrimNodeTextRange(ctx.SourceFile, outerCall)
	var statement *ast.Node
	spacingNode := ancestor
	if ancestor != nil && ast.IsExpressionStatement(ancestor) {
		statement = ancestor
		replaceRange = utils.TrimNodeTextRange(ctx.SourceFile, ancestor)
	} else if ancestor != nil && ast.IsArrowFunction(ancestor) &&
		ancestor.AsArrowFunction().Body == outerCall {
		arrowBody = true
	} else {
		return nil
	}

	body := functionBody(callback)
	returns, returnsSafe := callbackReturnStatements(callback, body)
	scopes := scopeanalysis.Get(ctx, scope.Options{CollectReferences: true})
	callbackScope := scopes.Acquire(callback)
	callScope := scopes.Acquire(call.Call)
	if body == nil || !returnsSafe || hasUnsafeCallbackSemantics(callback, body) ||
		isNamedFunctionSelfUsed(callback, scopes) ||
		!parametersSafeForFix(call.Call, callbackScope, callScope, parameters, scopes) {
		return nil
	}

	loop, ok := buildLoopText(ctx, call, parameters, body, returns, statement, callbackScope, scopes)
	if !ok {
		return nil
	}
	if arrowBody {
		loop = "{ " + loop + " }"
	}

	fixes := []rule.RuleFix{rule.RuleFixReplaceRange(replaceRange, loop)}
	return append(fixes, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, spacingNode)...)
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

func isNamedFunctionSelfUsed(callback *ast.Node, manager *scope.Manager) bool {
	if !ast.IsFunctionExpression(callback) || callback.Name() == nil ||
		!ast.IsIdentifier(callback.Name()) || manager == nil {
		return false
	}
	name := callback.Name().Text()
	for _, reference := range manager.References {
		if reference.Identifier.Text() != name {
			continue
		}
		for _, declaration := range reference.Declarations {
			if declaration.Kind == scope.DefFnExprName && declaration.DefNode == callback {
				return true
			}
		}
	}
	return false
}

func parametersSafeForFix(
	call *ast.Node,
	callbackScope, callScope *scope.Scope,
	parameters []*ast.Node,
	manager *scope.Manager,
) bool {
	if callbackScope == nil || callbackScope.Kind != scope.KindFunction ||
		callbackScope.Block == nil || manager == nil || callScope == nil {
		return false
	}

	// sourceCode.getDeclaredVariables(callback) returns one Variable per name,
	// with every declaration in that function scope collected in `defs`.
	// Moving the callback parameter into a for-of binding is unsafe whenever
	// any of those variables has merged/repeated definitions (notably a sloppy
	// mode `var` redeclaration of a parameter).
	for _, declarations := range callbackScope.ByName {
		if len(declarations) != 1 {
			return false
		}
	}

	parameterNames := map[string]struct{}{}
	for _, parameter := range parameters {
		utils.CollectBindingNames(parameter.Name(), func(_ *ast.Node, name string) {
			if _, exists := parameterNames[name]; exists {
				return
			}
			parameterNames[name] = struct{}{}
		})
	}
	for name := range parameterNames {
		declarations := callbackScope.Declarations(name)
		if len(declarations) != 1 || declarations[0].Kind != scope.DefParameter {
			return false
		}
	}

	references := make(map[*ast.Node]*scope.Reference, len(manager.References))
	for _, reference := range manager.References {
		references[reference.Identifier] = reference
	}

	unsafe := false
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil || unsafe {
			return
		}
		if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) {
			if _, conflicts := parameterNames[node.Text()]; conflicts {
				reference := references[node]
				if reference == nil || len(reference.Declarations) == 0 {
					unsafe = true
					return
				}
				resolvedScope := reference.Declarations[0].Scope
				if resolvedScope == callScope || isScopeAncestor(resolvedScope, callScope) {
					unsafe = true
					return
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return unsafe
		})
	}
	walk(call)
	return !unsafe
}

func isScopeAncestor(ancestor, child *scope.Scope) bool {
	if ancestor == nil || child == nil || ancestor == child {
		return false
	}
	for current := child.Parent; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}

func buildLoopText(
	ctx rule.RuleContext,
	call unicornutil.DotMethodCall,
	parameters []*ast.Node,
	body *ast.Node,
	returns []*ast.Node,
	statement *ast.Node,
	callbackScope *scope.Scope,
	manager *scope.Manager,
) (string, bool) {
	loopKind := "const"
	if callbackParametersReassigned(callbackScope, parameters, manager) {
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
	if body.Kind != ast.KindBlock && shouldParenthesizeExpressionStatement(ctx.SourceFile, body) {
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
			if shouldParenthesizeExpressionStatement(ctx.SourceFile, data.Expression) {
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

func callbackParametersReassigned(
	callbackScope *scope.Scope,
	parameters []*ast.Node,
	manager *scope.Manager,
) bool {
	if callbackScope == nil || manager == nil {
		return false
	}
	parameterNames := map[string]struct{}{}
	for _, parameter := range parameters {
		utils.CollectBindingNames(parameter.Name(), func(_ *ast.Node, name string) {
			parameterNames[name] = struct{}{}
		})
	}
	for _, reference := range manager.References {
		name := reference.Identifier.Text()
		if _, isParameter := parameterNames[name]; !isParameter ||
			!utils.IsWriteReference(reference.Identifier) {
			continue
		}
		for _, declaration := range reference.Declarations {
			if declaration.Scope == callbackScope && declaration.Kind == scope.DefParameter {
				return true
			}
		}
	}
	return false
}

func startsWithMemberHazard(text string) bool {
	if text == "" {
		return false
	}
	first := text[0]
	return first >= '0' && first <= '9'
}

func shouldParenthesizeExpressionStatement(sourceFile *ast.SourceFile, node *ast.Node) bool {
	if sourceFile == nil || node == nil || ast.IsParenthesizedExpression(node) {
		return false
	}
	first, ok := utils.TokenAtOrAfter(sourceFile, utils.TrimNodeTextRange(sourceFile, node).Pos())
	if !ok {
		return false
	}
	switch first.Kind {
	case ast.KindOpenBraceToken, ast.KindClassKeyword, ast.KindFunctionKeyword:
		return true
	}
	if first.Text != "async" {
		return false
	}
	second, ok := utils.TokenAtOrAfter(sourceFile, first.End)
	return ok && second.Kind == ast.KindFunctionKeyword
}
