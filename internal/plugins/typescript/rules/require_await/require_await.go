package require_await

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func buildMissingAwaitMessage(node *ast.Node) rule.RuleMessage {
	name := utils.UpperCaseFirstASCII(functionNameWithKind(node))
	return rule.RuleMessage{
		Id:          "missingAwait",
		Description: name + " has no 'await' expression.",
		Data:        map[string]string{"name": name},
	}
}

// functionNameWithKind mirrors @eslint-community/eslint-utils for tsgo's
// function-shaped nodes. Unlike ESLint core's helper, it recovers names from
// outer bindings and folds computed property names without a scope.
func functionNameWithKind(node *ast.Node) string {
	if node.Kind == ast.KindConstructor {
		return utils.GetFunctionNameWithKindCore(node)
	}

	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	if parent == nil {
		return ""
	}
	owner := functionPropertyOwner(node, parent)
	tokens := make([]string, 0, 5)

	if isClassFunctionOwner(node, owner) {
		if ast.HasSyntacticModifier(owner, ast.ModifierFlagsStatic) {
			tokens = append(tokens, "static")
		}
		if name := owner.Name(); name != nil && name.Kind == ast.KindPrivateIdentifier {
			tokens = append(tokens, "private")
		}
	}
	flags := ast.GetFunctionFlags(node)
	if flags&ast.FunctionFlagsAsync != 0 {
		tokens = append(tokens, "async")
	}
	if flags&ast.FunctionFlagsGenerator != 0 {
		tokens = append(tokens, "generator")
	}

	switch {
	case node.Kind == ast.KindGetAccessor:
		tokens = append(tokens, "getter")
	case node.Kind == ast.KindSetAccessor:
		tokens = append(tokens, "setter")
	case owner != nil:
		tokens = append(tokens, "method")
	case node.Kind == ast.KindArrowFunction:
		tokens = append(tokens, "arrow", "function")
	default:
		tokens = append(tokens, "function")
	}

	if owner != nil {
		if nameNode := owner.Name(); nameNode != nil && nameNode.Kind == ast.KindPrivateIdentifier {
			tokens = append(tokens, nameNode.AsPrivateIdentifier().Text)
		} else if name, ok := functionPropertyName(owner); ok && name != "" {
			tokens = append(tokens, "'"+name+"'")
		}
	} else if name := functionOwnName(node); name != "" {
		tokens = append(tokens, "'"+name+"'")
	} else if name := functionOuterName(node); name != "" {
		tokens = append(tokens, "'"+name+"'")
	}

	return strings.Join(tokens, " ")
}

func functionPropertyOwner(node *ast.Node, parent *ast.Node) *ast.Node {
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return node
	}

	switch parent.Kind {
	case ast.KindPropertyAssignment:
		if ast.SkipParentheses(parent.AsPropertyAssignment().Initializer) == node {
			return parent
		}
	case ast.KindPropertyDeclaration:
		declaration := parent.AsPropertyDeclaration()
		if !ast.HasSyntacticModifier(parent, ast.ModifierFlagsAccessor) &&
			declaration.Initializer != nil &&
			ast.SkipParentheses(declaration.Initializer) == node {
			return parent
		}
	}
	return nil
}

func isClassFunctionOwner(node *ast.Node, owner *ast.Node) bool {
	if owner == nil {
		return false
	}
	if owner == node {
		parent := node.Parent
		return parent != nil && (parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression)
	}
	parent := owner.Parent
	return parent != nil && (parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression)
}

func functionPropertyName(owner *ast.Node) (string, bool) {
	nameNode := owner.Name()
	if nameNode == nil {
		return "", false
	}
	if name, ok := utils.GetStaticPropertyName(nameNode); ok {
		return name, true
	}
	if nameNode.Kind != ast.KindComputedPropertyName {
		return "", false
	}
	expression := ast.SkipParentheses(nameNode.AsComputedPropertyName().Expression)
	if expression == nil || expression.Kind == ast.KindIdentifier {
		return "", false
	}
	return utils.NewStaticStringEvaluatorWithoutScope().EvalToString(expression)
}

func functionOwnName(node *ast.Node) string {
	var name *ast.Node
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		name = node.AsFunctionDeclaration().Name()
	case ast.KindFunctionExpression:
		name = node.AsFunctionExpression().Name()
	}
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

func functionOuterName(node *ast.Node) string {
	if node.Name() != nil {
		return ""
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsExportDefault) {
		return "default"
	}
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	if parent == nil {
		return ""
	}

	var binding *ast.Node
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		declaration := parent.AsVariableDeclaration()
		if declaration.Initializer != nil && ast.SkipParentheses(declaration.Initializer) == node {
			binding = parent.Name()
		}
	case ast.KindParameter:
		parameter := parent.AsParameterDeclaration()
		if parameter.Initializer != nil && ast.SkipParentheses(parameter.Initializer) == node {
			binding = parent.Name()
		}
	case ast.KindBindingElement:
		element := parent.AsBindingElement()
		if element.Initializer != nil && ast.SkipParentheses(element.Initializer) == node {
			binding = parent.Name()
		}
	case ast.KindShorthandPropertyAssignment:
		property := parent.AsShorthandPropertyAssignment()
		if property.ObjectAssignmentInitializer != nil && ast.SkipParentheses(property.ObjectAssignmentInitializer) == node {
			binding = parent.Name()
		}
	case ast.KindBinaryExpression:
		assignment := parent.AsBinaryExpression()
		if ast.IsAssignmentOperator(assignment.OperatorToken.Kind) && ast.SkipParentheses(assignment.Right) == node {
			binding = ast.SkipParentheses(assignment.Left)
		}
	case ast.KindExportAssignment:
		assignment := parent.AsExportAssignment()
		if !assignment.IsExportEquals && ast.SkipParentheses(assignment.Expression) == node {
			return "default"
		}
	}

	if binding != nil && binding.Kind == ast.KindIdentifier {
		return binding.AsIdentifier().Text
	}
	return ""
}

func buildRemoveAsyncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeAsync",
		Description: "Remove 'async'.",
	}
}

type asyncKeywordInfo struct {
	tokenRange        core.TextRange
	removeRange       core.TextRange
	hasAuthoredPrefix bool
}

func findAsyncKeyword(sourceFile *ast.SourceFile, node *ast.Node) (asyncKeywordInfo, bool) {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return asyncKeywordInfo{}, false
	}
	hasAuthoredPrefix := false
	for _, modifier := range modifiers.Nodes {
		if modifier == nil {
			continue
		}
		if modifier.Kind != ast.KindAsyncKeyword {
			if !utils.IsJSDocSyntaxNode(modifier) {
				hasAuthoredPrefix = true
			}
			continue
		}
		tokenRange := utils.TrimNodeTextRange(sourceFile, modifier)
		removeEnd := ecmascript.SkipLeadingWhitespace(sourceFile.Text(), tokenRange.End(), len(sourceFile.Text()))
		return asyncKeywordInfo{
			tokenRange:        tokenRange,
			removeRange:       core.NewTextRange(tokenRange.Pos(), removeEnd),
			hasAuthoredPrefix: hasAuthoredPrefix,
		}, true
	}
	return asyncKeywordInfo{}, false
}

func needsLeadingSemicolon(sourceFile *ast.SourceFile, node *ast.Node, asyncInfo asyncKeywordInfo, nextToken utils.SourceToken) bool {
	if node.Kind == ast.KindMethodDeclaration {
		// Removing async cannot expose nextToken as the member's first token
		// when a decorator or another modifier precedes it.
		if asyncInfo.hasAuthoredPrefix {
			return false
		}
		return utils.NeedsClassMemberLeadingSemicolon(
			sourceFile,
			ast.GetContainingClass(node),
			node,
			nextToken,
			utils.ClassMemberLeadingSemicolonOptions{
				IncludePropertiesWithoutInitializers: true,
				IncludePostfixInitializers:           true,
			},
		)
	}
	return (nextToken.Kind == ast.KindOpenBracketToken || nextToken.Kind == ast.KindOpenParenToken) &&
		utils.IsStartOfExpressionStatement(sourceFile, node) &&
		utils.NeedsPrecedingSemicolon(sourceFile, node)
}

func mayContainJSDocComment(text string, start int, end int) bool {
	return start >= 0 && start < end && end <= len(text) && strings.Contains(text[start:end], "/**")
}

func hasJSDocCommentBetween(sourceFile *ast.SourceFile, text string, start int, end int) bool {
	if start < 0 || start >= end || end > len(text) {
		return false
	}

	for comment := range utils.GetCommentsInRange(sourceFile, core.NewTextRange(start, end)) {
		if comment.Kind == ast.KindMultiLineCommentTrivia &&
			comment.Pos() >= start &&
			comment.End() <= end &&
			comment.Pos()+3 < comment.End() &&
			text[comment.Pos():comment.Pos()+3] == "/**" &&
			text[comment.Pos()+3] != '/' {
			return true
		}
	}
	return false
}

func appendReturnTypeSuggestionFixes(sourceFile *ast.SourceFile, node *ast.Node, isGenerator bool, fixes []rule.RuleFix) []rule.RuleFix {
	returnType := node.Type()
	if returnType == nil {
		return fixes
	}
	returnType = ast.SkipTypeParentheses(returnType)
	if returnType.Kind != ast.KindTypeReference {
		return fixes
	}

	typeReference := returnType.AsTypeReferenceNode()
	typeName := typeReference.TypeName
	if typeName == nil || typeName.Kind != ast.KindIdentifier {
		return fixes
	}

	typeNameText := typeName.AsIdentifier().Text
	if isGenerator {
		if typeNameText == "AsyncGenerator" {
			fixes = append(fixes, rule.RuleFixReplace(sourceFile, typeName, "Generator"))
		}
		return fixes
	}
	if typeNameText != "Promise" || typeReference.TypeArguments == nil {
		return fixes
	}

	typeNameRange := utils.TrimNodeTextRange(sourceFile, typeName)
	returnTypeRange := utils.TrimNodeTextRange(sourceFile, returnType)
	openAngleEnd := typeReference.TypeArguments.Pos()
	closeAngleStart := returnTypeRange.End() - 1
	text := sourceFile.Text()
	if openAngleEnd <= typeNameRange.End() ||
		openAngleEnd > len(text) ||
		text[openAngleEnd-1] != '<' ||
		closeAngleStart < openAngleEnd ||
		closeAngleStart >= len(text) ||
		text[closeAngleStart] != '>' {
		return fixes
	}

	return append(
		fixes,
		rule.RuleFixRemoveRange(core.NewTextRange(closeAngleStart, closeAngleStart+1)),
		rule.RuleFixRemoveRange(core.NewTextRange(typeNameRange.Pos(), openAngleEnd)),
	)
}

func buildRemoveAsyncSuggestion(sourceFile *ast.SourceFile, node *ast.Node, isGenerator bool) []rule.RuleSuggestion {
	asyncInfo, ok := findAsyncKeyword(sourceFile, node)
	if !ok {
		return nil
	}

	nextToken, hasNextToken := utils.TokenAtOrAfter(sourceFile, asyncInfo.tokenRange.End())
	text := sourceFile.Text()
	if hasNextToken &&
		mayContainJSDocComment(text, asyncInfo.tokenRange.End(), nextToken.Start) &&
		hasJSDocCommentBetween(sourceFile, text, asyncInfo.tokenRange.End(), nextToken.Start) {
		// Removing async can turn a JSDoc-style comment inside the function
		// header into leading JSDoc and silently add declaration metadata. There
		// is no syntax-preserving separator that works in every function context,
		// so omit the suggestion instead of changing program semantics.
		return nil
	}

	needsSemicolon := hasNextToken && needsLeadingSemicolon(sourceFile, node, asyncInfo, nextToken)
	fixes := make([]rule.RuleFix, 0, 1)
	if node.Kind == ast.KindMethodDeclaration && needsSemicolon {
		memberStart := node.Pos()
		asyncStart := asyncInfo.removeRange.Pos()
		if memberStart >= 0 && memberStart < asyncStart &&
			ecmascript.SkipLeadingWhitespace(text, memberStart, asyncStart) != asyncStart {
			// Put the separator before leading comments so JSDoc stays attached
			// to the method. Keep the insertion before the removal so the raw
			// fixes also form legal LSP edits.
			fixes = append(fixes,
				rule.RuleFixReplaceRange(core.NewTextRange(memberStart, memberStart), ";"),
				rule.RuleFixRemoveRange(asyncInfo.removeRange),
			)
		} else {
			// This single edit preserves the established `;member` output and
			// cannot conflict with the async removal at the same position.
			fixes = append(fixes, rule.RuleFixReplaceRange(asyncInfo.removeRange, ";"))
		}
	} else {
		replacement := ""
		if needsSemicolon {
			replacement = ";"
		}
		fixes = append(fixes, rule.RuleFixReplaceRange(asyncInfo.removeRange, replacement))
	}
	fixes = appendReturnTypeSuggestionFixes(sourceFile, node, isGenerator, fixes)
	return []rule.RuleSuggestion{{
		Message:  buildRemoveAsyncMessage(),
		FixesArr: fixes,
	}}
}

type scopeInfo struct {
	hasAwait      bool
	isAsyncYield  bool
	functionFlags ast.FunctionFlags
}

func hasCallSignature(typeChecker *checker.Checker, t *checker.Type) bool {
	if utils.IsUnionType(t) {
		for _, typePart := range t.Types() {
			if len(utils.GetCallSignatures(typeChecker, typePart)) != 0 {
				return true
			}
		}
		return false
	}
	return len(utils.GetCallSignatures(typeChecker, t)) != 0
}

func isCallbackParameter(typeChecker *checker.Checker, param *ast.Symbol, node *ast.Node) bool {
	t := checker.Checker_getApparentType(typeChecker, typeChecker.GetTypeOfSymbolAtLocation(param, node))
	if param.ValueDeclaration != nil &&
		ast.IsParameterDeclaration(param.ValueDeclaration) &&
		param.ValueDeclaration.AsParameterDeclaration().DotDotDotToken != nil {
		t = checker.Checker_getIndexTypeOfType(typeChecker, t, checker.Checker_numberType(typeChecker))
		if t == nil {
			return false
		}
	}
	return hasCallSignature(typeChecker, t)
}

func hasThenCallbackSignature(typeChecker *checker.Checker, node *ast.Node, t *checker.Type) bool {
	if utils.IsUnionType(t) {
		for _, typePart := range t.Types() {
			if hasThenCallbackSignature(typeChecker, node, typePart) {
				return true
			}
		}
		return false
	}
	for _, signature := range utils.GetCallSignatures(typeChecker, t) {
		parameters := checker.Signature_parameters(signature)
		if len(parameters) != 0 && isCallbackParameter(typeChecker, parameters[0], node) {
			return true
		}
	}
	return false
}

func isThenableTypePart(typeChecker *checker.Checker, node *ast.Node, t *checker.Type) bool {
	then := checker.Checker_getPropertyOfType(typeChecker, t, "then")
	if then == nil {
		return false
	}
	return hasThenCallbackSignature(typeChecker, node, typeChecker.GetTypeOfSymbolAtLocation(then, node))
}

// isThenableType mirrors utils.IsThenableType without materializing singleton
// slices for the overwhelmingly common non-union path.
func isThenableType(typeChecker *checker.Checker, node *ast.Node, t *checker.Type) bool {
	if t == nil {
		t = typeChecker.GetTypeAtLocation(node)
	}
	t = checker.Checker_getApparentType(typeChecker, t)
	if utils.IsUnionType(t) {
		for _, typePart := range t.Types() {
			if isThenableTypePart(typeChecker, node, typePart) {
				return true
			}
		}
		return false
	}
	return isThenableTypePart(typeChecker, node, t)
}

func hasAsyncIterator(typeChecker *checker.Checker, t *checker.Type) bool {
	if utils.IsTypeFlagSet(t, checker.TypeFlagsUnionOrIntersection) {
		for _, typePart := range t.Types() {
			if hasAsyncIterator(typeChecker, typePart) {
				return true
			}
		}
		return false
	}
	return utils.GetWellKnownSymbolPropertyOfType(t, "asyncIterator", typeChecker) != nil
}

var RequireAwaitRule = rule.CreateRule(rule.Rule{
	Name:             "require-await",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var scopes []scopeInfo

		enterFunction := func(node *ast.Node) {
			scope := scopeInfo{
				hasAwait:      false,
				isAsyncYield:  false,
				functionFlags: ast.FunctionFlagsNormal,
			}
			if utils.HasNonEmptyFunctionBody(node) {
				scope.functionFlags = ast.GetFunctionFlags(node)
			}
			scopes = append(scopes, scope)
		}

		exitFunction := func(node *ast.Node) {
			scope := &scopes[len(scopes)-1]
			if scope.functionFlags&ast.FunctionFlagsAsync != 0 && !scope.hasAwait && (scope.functionFlags&ast.FunctionFlagsGenerator == 0 || !scope.isAsyncYield) {
				isGenerator := scope.functionFlags&ast.FunctionFlagsGenerator != 0
				ctx.ReportRangeWithDeferredSuggestions(
					utils.GetFunctionHeadLoc(ctx.SourceFile, node),
					buildMissingAwaitMessage(node),
					func() []rule.RuleSuggestion {
						return buildRemoveAsyncSuggestion(ctx.SourceFile, node, isGenerator)
					},
				)
			}

			scopes = scopes[:len(scopes)-1]
		}

		markAsHasAwait := func() {
			if len(scopes) != 0 {
				scopes[len(scopes)-1].hasAwait = true
			}
		}

		exitArrowFunction := func(node *ast.Node) {
			scope := &scopes[len(scopes)-1]
			if scope.functionFlags&ast.FunctionFlagsAsync != 0 && !scope.hasAwait {
				body := ast.SkipParentheses(node.Body())
				if !ast.IsBlock(body) &&
					!ast.IsAwaitExpression(body) &&
					isThenableType(ctx.TypeChecker, body, ctx.TypeChecker.GetTypeAtLocation(body)) {
					scope.hasAwait = true
				}
			}
			exitFunction(node)
		}

		return rule.RuleListeners{
			// from isFunctionLikeDeclarationKind
			ast.KindFunctionDeclaration:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			ast.KindMethodDeclaration:                        enterFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,
			ast.KindConstructor:                              enterFunction,
			rule.ListenerOnExit(ast.KindConstructor):         exitFunction,
			ast.KindGetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):         exitFunction,
			ast.KindSetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):         exitFunction,
			ast.KindFunctionExpression:                       enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			ast.KindArrowFunction:                            enterFunction,
			rule.ListenerOnExit(ast.KindArrowFunction):       exitArrowFunction,

			ast.KindAwaitExpression: func(node *ast.Node) { markAsHasAwait() },
			ast.KindForOfStatement: func(node *ast.Node) {
				if node.AsForInOrOfStatement().AwaitModifier != nil {
					markAsHasAwait()
				}
			},
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if ast.IsVarAwaitUsing(node) {
					markAsHasAwait()
				}
			},
			/**
			 * Mark `scopeInfo.isAsyncYield` to `true` if it
			 *  1) delegates async generator function
			 *    or
			 *  2) yields thenable type
			 */
			rule.ListenerOnExit(ast.KindYieldExpression): func(node *ast.Node) {
				if len(scopes) == 0 {
					return
				}
				scope := &scopes[len(scopes)-1]
				argument := node.Expression()
				if scope.hasAwait ||
					scope.isAsyncYield ||
					scope.functionFlags&ast.FunctionFlagsAsyncGenerator != ast.FunctionFlagsAsyncGenerator ||
					argument == nil {
					return
				}

				if ast.IsLiteralExpression(argument) {
					// ignoring this as for literals we don't need to check the definition
					// eg : async function* run() { yield* 1 }
					return
				}

				if node.AsYieldExpression().AsteriskToken == nil {
					if isThenableType(ctx.TypeChecker, argument, ctx.TypeChecker.GetTypeAtLocation(argument)) {
						scope.isAsyncYield = true
					}
					return
				}

				t := ctx.TypeChecker.GetTypeAtLocation(argument)
				if hasAsyncIterator(ctx.TypeChecker, t) {
					scope.isAsyncYield = true
				}
			},
			rule.ListenerOnExit(ast.KindReturnStatement): func(node *ast.Node) {
				if len(scopes) == 0 {
					return
				}
				scope := &scopes[len(scopes)-1]
				if scope.hasAwait || scope.functionFlags&ast.FunctionFlagsAsync == 0 {
					return
				}

				expr := node.Expression()
				if expr != nil && isThenableType(ctx.TypeChecker, expr, ctx.TypeChecker.GetTypeAtLocation(expr)) {
					scope.hasAwait = true
				}
			},
		}
	},
})
