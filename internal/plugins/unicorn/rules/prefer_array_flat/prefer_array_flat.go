package prefer_array_flat

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const messageID = "prefer-array-flat"

var (
	//go:embed prefer_array_flat.schema.json
	schemaJSON []byte

	decimalIntegerPattern = regexp.MustCompile(`^(?:0|0[0-7]*[89][0-9]*|[1-9](?:_?[0-9])*)$`)

	lodashFlattenFunctions = []string{
		"_.flatten",
		"lodash.flatten",
		"underscore.flatten",
	}
)

type flattenMatch struct {
	array         *ast.Node
	description   string
	switchToArray bool
	optional      bool
}

func message(description string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: fmt.Sprintf("Prefer `Array#flat()` over `%s` to flatten an array.", description),
		Data:        map[string]string{"description": description},
	}
}

func simpleParameterIdentifier(parameter *ast.Node) *ast.Node {
	if parameter == nil || parameter.Kind != ast.KindParameter {
		return nil
	}
	declaration := parameter.AsParameterDeclaration()
	if declaration == nil || declaration.DotDotDotToken != nil || declaration.Initializer != nil {
		return nil
	}
	name := declaration.Name()
	if name == nil || !ast.IsIdentifier(name) {
		return nil
	}
	return name
}

func isSameIdentifier(left *ast.Node, right *ast.Node) bool {
	left = ast.SkipParentheses(left)
	right = ast.SkipParentheses(right)
	return left != nil && right != nil &&
		ast.IsIdentifier(left) && ast.IsIdentifier(right) &&
		left.AsIdentifier().Text == right.AsIdentifier().Text
}

func isAsync(node *ast.Node) bool {
	return utils.IncludesModifier(node, ast.KindAsyncKeyword)
}

func matchArrayFlatMap(node *ast.Node, ctx rule.RuleContext) (flattenMatch, bool) {
	oneArgument := 1
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Method:              "flatMap",
		ArgumentsLength:     &oneArgument,
		AllowOptionalMember: true,
	})
	if !ok {
		return flattenMatch{}, false
	}

	arguments := node.Arguments()
	if len(arguments) != 1 || ast.IsSpreadElement(arguments[0]) {
		return flattenMatch{}, false
	}
	callback := ast.SkipParentheses(arguments[0])
	if callback == nil || callback.Kind != ast.KindArrowFunction || isAsync(callback) {
		return flattenMatch{}, false
	}

	arrow := callback.AsArrowFunction()
	if arrow == nil || arrow.Parameters == nil || len(arrow.Parameters.Nodes) != 1 ||
		arrow.Body == nil {
		return flattenMatch{}, false
	}
	parameter := simpleParameterIdentifier(arrow.Parameters.Nodes[0])
	if parameter == nil || !isSameIdentifier(parameter, arrow.Body) ||
		isObviouslyNonArrayFlatMapReceiver(call.Object, ctx) {
		return flattenMatch{}, false
	}

	return flattenMatch{
		array:       call.Object,
		description: "Array#flatMap()",
		optional:    call.Callee.AsPropertyAccessExpression().QuestionDotToken != nil,
	}, true
}

func matchArrayReduce(node *ast.Node) (flattenMatch, bool) {
	twoArguments := 2
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Method:              "reduce",
		ArgumentsLength:     &twoArguments,
		AllowOptionalMember: true,
	})
	if !ok {
		return flattenMatch{}, false
	}

	arguments := node.Arguments()
	if len(arguments) != 2 || ast.IsSpreadElement(arguments[0]) ||
		ast.IsSpreadElement(arguments[1]) ||
		!ast.IsEmptyArrayLiteral(ast.SkipParentheses(arguments[1])) {
		return flattenMatch{}, false
	}

	callback := ast.SkipParentheses(arguments[0])
	if callback == nil || callback.Kind != ast.KindArrowFunction || isAsync(callback) {
		return flattenMatch{}, false
	}
	arrow := callback.AsArrowFunction()
	if arrow == nil || arrow.Parameters == nil || len(arrow.Parameters.Nodes) != 2 ||
		arrow.Body == nil {
		return flattenMatch{}, false
	}

	firstParameter := simpleParameterIdentifier(arrow.Parameters.Nodes[0])
	secondParameter := simpleParameterIdentifier(arrow.Parameters.Nodes[1])
	if firstParameter == nil || secondParameter == nil {
		return flattenMatch{}, false
	}

	body := ast.SkipParentheses(arrow.Body)
	if !matchesConcatReducer(body, firstParameter, secondParameter) &&
		!matchesSpreadReducer(body, firstParameter, secondParameter) {
		return flattenMatch{}, false
	}

	return flattenMatch{
		array:       call.Object,
		description: "Array#reduce()",
		optional:    call.Callee.AsPropertyAccessExpression().QuestionDotToken != nil,
	}, true
}

func matchesConcatReducer(body *ast.Node, firstParameter *ast.Node, secondParameter *ast.Node) bool {
	oneArgument := 1
	call, ok := unicornutil.MatchDotMethodCall(body, unicornutil.DotMethodCallOptions{
		Method:          "concat",
		ArgumentsLength: &oneArgument,
	})
	if !ok {
		return false
	}
	arguments := body.Arguments()
	return len(arguments) == 1 &&
		!ast.IsSpreadElement(arguments[0]) &&
		isSameIdentifier(firstParameter, call.Object) &&
		isSameIdentifier(secondParameter, arguments[0])
}

func matchesSpreadReducer(body *ast.Node, firstParameter *ast.Node, secondParameter *ast.Node) bool {
	if body == nil || body.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	elements := body.AsArrayLiteralExpression().Elements
	if elements == nil || len(elements.Nodes) != 2 {
		return false
	}

	parameters := []*ast.Node{firstParameter, secondParameter}
	for index, element := range elements.Nodes {
		if element == nil || element.Kind != ast.KindSpreadElement {
			return false
		}
		if !isSameIdentifier(parameters[index], element.AsSpreadElement().Expression) {
			return false
		}
	}
	return true
}

func matchEmptyArrayConcat(node *ast.Node) (flattenMatch, bool) {
	oneArgument := 1
	_, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Method:          "concat",
		ArgumentsLength: &oneArgument,
	})
	if !ok {
		return flattenMatch{}, false
	}

	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	object := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	if !ast.IsEmptyArrayLiteral(object) {
		return flattenMatch{}, false
	}

	argument := node.Arguments()[0]
	if ast.IsSpreadElement(argument) {
		return flattenMatch{
			array:       argument.AsSpreadElement().Expression,
			description: "[].concat()",
		}, true
	}
	return flattenMatch{
		array:         argument,
		description:   "[].concat()",
		switchToArray: true,
	}, true
}

func matchArrayPrototypeConcat(node *ast.Node) (flattenMatch, bool) {
	twoArguments := 2
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Method:          "apply",
		ArgumentsLength: &twoArguments,
	})
	method := "apply"
	if !ok {
		call, ok = unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
			Method:          "call",
			ArgumentsLength: &twoArguments,
		})
		method = "call"
	}
	if !ok || !unicornutil.IsArrayPrototypeProperty(call.Object, "concat") {
		return flattenMatch{}, false
	}

	arguments := node.Arguments()
	if len(arguments) != 2 || ast.IsSpreadElement(arguments[0]) ||
		!ast.IsEmptyArrayLiteral(ast.SkipParentheses(arguments[0])) {
		return flattenMatch{}, false
	}

	secondArgument := arguments[1]
	isSpread := ast.IsSpreadElement(secondArgument)
	if method == "apply" && isSpread {
		return flattenMatch{}, false
	}
	if isSpread {
		secondArgument = secondArgument.AsSpreadElement().Expression
	}

	return flattenMatch{
		array:         secondArgument,
		description:   "Array.prototype.concat()",
		switchToArray: method == "call" && !isSpread,
	}, true
}

func matchFlattenFunction(node *ast.Node, functions []string) (flattenMatch, bool) {
	if node == nil || !ast.IsCallExpression(node) {
		return flattenMatch{}, false
	}
	call := node.AsCallExpression()
	arguments := node.Arguments()
	if call.QuestionDotToken != nil || len(arguments) != 1 || ast.IsSpreadElement(arguments[0]) {
		return flattenMatch{}, false
	}

	callee := ast.SkipParentheses(call.Expression)
	for _, function := range functions {
		if unicornutil.NodeMatchesPath(callee, function) {
			return flattenMatch{
				array:       arguments[0],
				description: strings.TrimSpace(function) + "()",
			}, true
		}
	}
	return flattenMatch{}, false
}

func getConstVariableInitializer(node *ast.Node, ctx rule.RuleContext) *ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsIdentifier(node) || ctx.TypeChecker == nil {
		return nil
	}

	symbol := utils.GetReferenceSymbol(node, ctx.TypeChecker)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration ||
		declaration.Parent == nil ||
		utils.GetVarDeclListKind(declaration.Parent) != "const" {
		return nil
	}
	return declaration.AsVariableDeclaration().Initializer
}

func isPascalCaseIdentifier(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsIdentifier(node) {
		return false
	}
	first, _ := utf8.DecodeRuneInString(node.AsIdentifier().Text)
	return first != utf8.RuneError && unicode.IsUpper(first)
}

func isDefinitelyArrayExpression(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindArrayLiteralExpression {
		return true
	}
	if node.Kind != ast.KindNewExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsNewExpression().Expression)
	return callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "Array"
}

func isDefinitelyNonArrayExpression(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindObjectLiteralExpression,
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindClassExpression:
		return true
	case ast.KindNewExpression:
		callee := ast.SkipParentheses(node.AsNewExpression().Expression)
		return callee != nil && ast.IsIdentifier(callee) &&
			callee.AsIdentifier().Text != "Array"
	default:
		return false
	}
}

func isObviouslyNonArrayFlatMapReceiver(node *ast.Node, ctx rule.RuleContext) bool {
	initializer := getConstVariableInitializer(node, ctx)
	isConstArray := initializer != nil && isDefinitelyArrayExpression(initializer)
	isConstNonArray := initializer != nil && isDefinitelyNonArrayExpression(initializer)
	return (isPascalCaseIdentifier(node) && !isConstArray) || isConstNonArray
}

func countCommentsInside(comments []*ast.CommentRange, textRange core.TextRange) int {
	count := 0
	for _, comment := range comments {
		if comment.Pos() >= textRange.Pos() && comment.End() <= textRange.End() {
			count++
		}
	}
	return count
}

func canFix(node *ast.Node, array *ast.Node, ctx rule.RuleContext) bool {
	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	// ESTree drops parentheses, so comments inside a parenthesized argument but
	// outside the argument expression suppress the upstream fix.
	arrayRange := utils.TrimNodeTextRange(ctx.SourceFile, ast.SkipParentheses(array))
	comments := ctx.Comments.All()
	return countCommentsInside(comments, nodeRange) ==
		countCommentsInside(comments, arrayRange)
}

func shouldParenthesizeMemberObject(node *ast.Node, sourceText string) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindCallExpression,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression,
		ast.KindThisKeyword,
		ast.KindArrayLiteralExpression,
		ast.KindFunctionExpression,
		ast.KindStringLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword:
		return false
	case ast.KindNewExpression:
		return node.AsNewExpression().Arguments == nil
	case ast.KindNumericLiteral:
		return decimalIntegerPattern.MatchString(sourceText)
	default:
		return true
	}
}

func replacementText(sourceFile *ast.SourceFile, match flattenMatch) string {
	arrayText := utils.TrimmedNodeText(sourceFile, match.array)
	fixed := arrayText
	if match.switchToArray {
		fixed = "[" + fixed + "]"
	} else if match.array.Kind != ast.KindParenthesizedExpression &&
		shouldParenthesizeMemberObject(match.array, arrayText) {
		fixed = "(" + fixed + ")"
	}
	if match.optional {
		fixed += "?"
	}
	return fixed + ".flat()"
}

func outerParenthesizedNode(node *ast.Node) *ast.Node {
	outer := node
	for outer != nil && outer.Parent != nil &&
		outer.Parent.Kind == ast.KindParenthesizedExpression {
		parent := outer.Parent.AsParenthesizedExpression()
		if parent == nil || parent.Expression != outer {
			break
		}
		outer = outer.Parent
	}
	return outer
}

func adjacentWordBefore(source string, position int) string {
	if position <= 0 || position > len(source) ||
		source[position-1] < 'a' || source[position-1] > 'z' {
		return ""
	}
	start := position - 1
	for start > 0 && source[start-1] >= 'a' && source[start-1] <= 'z' {
		start--
	}
	return source[start:position]
}

func adjacentWordAfter(source string, position int) string {
	if position < 0 || position >= len(source) ||
		source[position] < 'a' || source[position] > 'z' {
		return ""
	}
	end := position + 1
	for end < len(source) && source[end] >= 'a' && source[end] <= 'z' {
		end++
	}
	return source[position:end]
}

func isProblematicKeyword(word string) bool {
	if word == "" {
		return false
	}
	kind := scanner.StringToToken(word)
	return (kind >= ast.KindFirstKeyword && kind <= ast.KindLastKeyword) ||
		word == "of" || word == "await"
}

func keywordSpacingFixes(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	outer := outerParenthesizedNode(node)
	textRange := utils.TrimNodeTextRange(sourceFile, outer)
	source := sourceFile.Text()
	var fixes []rule.RuleFix
	if isProblematicKeyword(adjacentWordBefore(source, textRange.Pos())) {
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(textRange.Pos(), textRange.Pos()),
			" ",
		))
	}
	if isProblematicKeyword(adjacentWordAfter(source, textRange.End())) {
		fixes = append(fixes, rule.RuleFixReplaceRange(
			core.NewTextRange(textRange.End(), textRange.End()),
			" ",
		))
	}
	return fixes
}

func startsWithSemicolonHazard(text string) bool {
	if text == "" {
		return false
	}
	return strings.ContainsRune("[(/`+-*,.", rune(text[0]))
}

func previousTokenKind(sourceFile *ast.SourceFile, position int) ast.Kind {
	token := ast.KindUnknown
	s := scanner.GetScannerForSourceFile(sourceFile, 0)
	for s.Token() != ast.KindEndOfFile && s.TokenStart() < position {
		if s.TokenEnd() <= position {
			token = s.Token()
		}
		s.Scan()
	}
	return token
}

func isEmbeddedStatement(statement *ast.Node) bool {
	if statement == nil || statement.Parent == nil {
		return false
	}
	parent := statement.Parent
	switch parent.Kind {
	case ast.KindIfStatement:
		ifStatement := parent.AsIfStatement()
		return ifStatement.ThenStatement == statement || ifStatement.ElseStatement == statement
	case ast.KindForStatement:
		return parent.AsForStatement().Statement == statement
	case ast.KindForInStatement, ast.KindForOfStatement:
		return parent.AsForInOrOfStatement().Statement == statement
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Statement == statement
	case ast.KindDoStatement:
		return parent.AsDoStatement().Statement == statement
	case ast.KindWithStatement:
		return parent.AsWithStatement().Statement == statement
	default:
		return false
	}
}

func needsSemicolonBefore(sourceFile *ast.SourceFile, node *ast.Node, fixed string) bool {
	if !startsWithSemicolonHazard(fixed) || outerParenthesizedNode(node) != node ||
		node.Parent == nil || node.Parent.Kind != ast.KindExpressionStatement ||
		isEmbeddedStatement(node.Parent) {
		return false
	}

	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	switch previousTokenKind(sourceFile, nodeRange.Pos()) {
	case ast.KindCloseBracketToken,
		ast.KindCloseParenToken,
		ast.KindIdentifier,
		ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateTail,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword:
		return true
	default:
		return false
	}
}

func buildFixes(node *ast.Node, match flattenMatch, ctx rule.RuleContext) []rule.RuleFix {
	fixed := replacementText(ctx.SourceFile, match)
	if needsSemicolonBefore(ctx.SourceFile, node, fixed) {
		fixed = ";" + fixed
	}

	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(utils.TrimNodeTextRange(ctx.SourceFile, node), fixed),
	}
	return append(fixes, keywordSpacingFixes(ctx.SourceFile, node)...)
}

func parseFunctions(options []any) []string {
	functions := make([]string, 0, len(lodashFlattenFunctions))
	if optionsMap := utils.GetOptionsMap(options); optionsMap != nil {
		functions = append(functions, utils.ToStringSlice(optionsMap["functions"])...)
	}
	return append(functions, lodashFlattenFunctions...)
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v64.0.0/docs/rules/prefer-array-flat.md
var PreferArrayFlatRule = rule.Rule{
	Name:   "unicorn/prefer-array-flat",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		functions := parseFunctions(options)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				var matches []flattenMatch
				if match, ok := matchArrayFlatMap(node, ctx); ok {
					matches = append(matches, match)
				}
				if match, ok := matchArrayReduce(node); ok {
					matches = append(matches, match)
				}
				if match, ok := matchEmptyArrayConcat(node); ok {
					matches = append(matches, match)
				}
				if match, ok := matchArrayPrototypeConcat(node); ok {
					matches = append(matches, match)
				}
				if match, ok := matchFlattenFunction(node, functions); ok {
					matches = append(matches, match)
				}

				for _, match := range matches {
					ruleMessage := message(match.description)
					if canFix(node, match.array, ctx) {
						ctx.ReportNodeWithFixes(node, ruleMessage, buildFixes(node, match, ctx)...)
					} else {
						ctx.ReportNode(node, ruleMessage)
					}
				}
			},
		}
	},
}
