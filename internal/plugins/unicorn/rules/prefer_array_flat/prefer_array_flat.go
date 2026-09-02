package prefer_array_flat

import (
	_ "embed"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/unicode17"
)

const messageID = "prefer-array-flat"

var (
	//go:embed prefer_array_flat.schema.json
	schemaJSON []byte

	lodashFlattenFunctions = []flattenFunction{
		{name: "_.flatten", leaf: "flatten"},
		{name: "lodash.flatten", leaf: "flatten"},
		{name: "underscore.flatten", leaf: "flatten"},
	}
)

type flattenMatch struct {
	array       *ast.Node
	description string
	optional    bool
}

type flattenFunction struct {
	name string
	leaf string
}

func message(description string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          messageID,
		Description: fmt.Sprintf("Prefer `Array#flat()` over `%s` to flatten an array.", description),
		Data:        map[string]string{"description": description},
	}
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
	callback := utils.ESTreeRuntimeExpression(arguments[0])
	if callback == nil || !ast.IsArrowFunction(callback) || ast.IsAsyncFunction(callback) {
		return flattenMatch{}, false
	}

	arrow := callback.AsArrowFunction()
	if arrow == nil || arrow.Parameters == nil || len(arrow.Parameters.Nodes) != 1 ||
		arrow.Body == nil {
		return flattenMatch{}, false
	}
	parameter := unicornutil.PlainParameterIdentifier(arrow.Parameters.Nodes[0])
	if parameter == nil ||
		!unicornutil.IsSameIdentifier(parameter, utils.ESTreeRuntimeExpression(arrow.Body)) ||
		isObviouslyNonArrayFlatMapReceiver(utils.ESTreeRuntimeExpression(call.Object), ctx) {
		return flattenMatch{}, false
	}

	return flattenMatch{
		array:       call.Object,
		description: "Array#flatMap()",
		optional:    ast.IsOptionalChainRoot(call.Callee),
	}, true
}

func matchArrayReduce(node *ast.Node, ctx rule.RuleContext) (flattenMatch, bool) {
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
		!ast.IsEmptyArrayLiteral(utils.ESTreeRuntimeExpression(arguments[1])) {
		return flattenMatch{}, false
	}

	callback := utils.ESTreeRuntimeExpression(arguments[0])
	if callback == nil || !ast.IsArrowFunction(callback) || ast.IsAsyncFunction(callback) {
		return flattenMatch{}, false
	}
	arrow := callback.AsArrowFunction()
	if arrow == nil || arrow.Parameters == nil || len(arrow.Parameters.Nodes) != 2 ||
		arrow.Body == nil {
		return flattenMatch{}, false
	}

	firstParameter := unicornutil.PlainParameterIdentifier(arrow.Parameters.Nodes[0])
	secondParameter := unicornutil.PlainParameterIdentifier(arrow.Parameters.Nodes[1])
	if firstParameter == nil || secondParameter == nil {
		return flattenMatch{}, false
	}

	body := utils.ESTreeRuntimeExpression(arrow.Body)
	if !matchesConcatReducer(body, firstParameter, secondParameter) &&
		!matchesSpreadReducer(body, firstParameter, secondParameter) {
		return flattenMatch{}, false
	}
	if unicornutil.IsKnownNonArray(ctx, utils.ESTreeRuntimeExpression(call.Object)) {
		return flattenMatch{}, false
	}

	return flattenMatch{
		array:       call.Object,
		description: "Array#reduce()",
		optional:    ast.IsOptionalChainRoot(call.Callee),
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
		unicornutil.IsSameIdentifier(
			firstParameter,
			utils.ESTreeRuntimeExpression(call.Object),
		) &&
		unicornutil.IsSameIdentifier(
			secondParameter,
			utils.ESTreeRuntimeExpression(arguments[0]),
		)
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
		if !unicornutil.IsSameIdentifier(
			parameters[index],
			utils.ESTreeRuntimeExpression(element.AsSpreadElement().Expression),
		) {
			return false
		}
	}
	return true
}

func matchEmptyArrayConcat(node *ast.Node) (flattenMatch, bool) {
	oneArgument := 1
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Method:          "concat",
		ArgumentsLength: &oneArgument,
	})
	if !ok {
		return flattenMatch{}, false
	}

	object := utils.ESTreeRuntimeExpression(call.Object)
	if !ast.IsEmptyArrayLiteral(object) {
		return flattenMatch{}, false
	}

	argument := node.Arguments()[0]
	if !ast.IsSpreadElement(argument) {
		return flattenMatch{}, false
	}
	return flattenMatch{
		array:       argument.AsSpreadElement().Expression,
		description: "[].concat()",
	}, true
}

func matchArrayPrototypeConcat(node *ast.Node) (flattenMatch, bool) {
	twoArguments := 2
	call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
		Methods:         []string{"apply", "call"},
		ArgumentsLength: &twoArguments,
	})
	if !ok || !unicornutil.IsArrayPrototypeProperty(call.Object, "concat") {
		return flattenMatch{}, false
	}

	arguments := node.Arguments()
	if len(arguments) != 2 || ast.IsSpreadElement(arguments[0]) ||
		!ast.IsEmptyArrayLiteral(utils.ESTreeRuntimeExpression(arguments[0])) {
		return flattenMatch{}, false
	}

	method := call.Property.AsIdentifier().Text
	secondArgument := arguments[1]
	isSpread := ast.IsSpreadElement(secondArgument)
	matchesApply := method == "apply" && !isSpread
	matchesCall := method == "call" && isSpread
	if !matchesApply && !matchesCall {
		return flattenMatch{}, false
	}
	if isSpread {
		secondArgument = secondArgument.AsSpreadElement().Expression
	}

	return flattenMatch{
		array:       secondArgument,
		description: "Array.prototype.concat()",
	}, true
}

func matchFlattenFunction(node *ast.Node, functions []flattenFunction) (flattenMatch, bool) {
	if node == nil || !ast.IsCallExpression(node) {
		return flattenMatch{}, false
	}
	call := node.AsCallExpression()
	arguments := node.Arguments()
	if ast.IsOptionalChainRoot(node) || len(arguments) != 1 ||
		ast.IsSpreadElement(arguments[0]) {
		return flattenMatch{}, false
	}

	callee := utils.ESTreeRuntimeExpression(call.Expression)
	leaf := nodePathLeaf(callee)
	if leaf == "" {
		return flattenMatch{}, false
	}
	for _, function := range functions {
		if leaf == function.leaf && unicornutil.NodeMatchesPath(callee, function.name) {
			return flattenMatch{
				array:       arguments[0],
				description: function.name + "()",
			}, true
		}
	}
	return flattenMatch{}, false
}

func nodePathLeaf(node *ast.Node) string {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return ""
	}
	if ast.IsIdentifier(node) {
		return node.AsIdentifier().Text
	}
	switch node.Kind {
	case ast.KindThisKeyword:
		return "this"
	case ast.KindSuperKeyword:
		return "super"
	}
	if ast.IsPropertyAccessExpression(node) {
		name := node.AsPropertyAccessExpression().Name()
		if name != nil && ast.IsIdentifier(name) {
			return name.AsIdentifier().Text
		}
		return ""
	}
	if node.Kind == ast.KindMetaProperty {
		name := node.AsMetaProperty().Name()
		if name != nil {
			return name.Text()
		}
	}
	return ""
}

// Upstream checks only whether the first Unicode rune is an uppercase letter;
// this is not intended to validate the rest of a PascalCase name.
func isPascalCaseIdentifier(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || !ast.IsIdentifier(node) {
		return false
	}
	first, _ := utf8.DecodeRuneInString(node.AsIdentifier().Text)
	return first != utf8.RuneError && unicode17.IsUpper(first)
}

func isDefinitelyArrayExpression(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return false
	}
	if ast.IsArrayLiteralExpression(node) {
		return true
	}
	if !ast.IsNewExpression(node) {
		return false
	}
	callee := utils.ESTreeRuntimeExpression(node.AsNewExpression().Expression)
	return callee != nil && ast.IsIdentifier(callee) && callee.AsIdentifier().Text == "Array"
}

func isDefinitelyNonArrayExpression(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
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
		callee := utils.ESTreeRuntimeExpression(node.AsNewExpression().Expression)
		return callee != nil && ast.IsIdentifier(callee) &&
			callee.AsIdentifier().Text != "Array"
	default:
		return false
	}
}

func isObviouslyNonArrayFlatMapReceiver(node *ast.Node, ctx rule.RuleContext) bool {
	// Deliberately keep Unicorn's syntactic classification. Following only a
	// single const initializer avoids broadening reports based on inferred
	// TypeScript types.
	initializer := utils.GetConstVariableInitializer(node, ctx.TypeChecker)
	isConstArray := initializer != nil && isDefinitelyArrayExpression(initializer)
	isConstNonArray := initializer != nil && isDefinitelyNonArrayExpression(initializer)
	return (isPascalCaseIdentifier(node) && !isConstArray) || isConstNonArray
}

func canFix(node *ast.Node, array *ast.Node, ctx rule.RuleContext) bool {
	nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	arrayRange := utils.TrimNodeTextRange(ctx.SourceFile, utils.ESTreeRuntimeExpression(array))
	comments := ctx.Comments.All()
	// Upstream fixes only when every comment in the matched call belongs to the
	// selected array expression. Since array is nested in node, check the two
	// surrounding spans. Skip its parentheses because ESTree drops them.
	return !utils.HasCommentInSpan(comments, nodeRange.Pos(), arrayRange.Pos()) &&
		!utils.HasCommentInSpan(comments, arrayRange.End(), nodeRange.End())
}

func replacementText(sourceFile *ast.SourceFile, match flattenMatch) string {
	arrayText := utils.TrimmedNodeText(sourceFile, match.array)
	fixed := arrayText
	if match.array.Kind != ast.KindParenthesizedExpression &&
		unicornutil.ShouldAddParenthesesToMemberExpressionObject(sourceFile, match.array) {
		fixed = "(" + fixed + ")"
	}
	if match.optional {
		fixed += "?"
	}
	return fixed + ".flat()"
}

func buildFixes(node *ast.Node, match flattenMatch, ctx rule.RuleContext) []rule.RuleFix {
	fixed := replacementText(ctx.SourceFile, match)
	if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, fixed) {
		fixed = ";" + fixed
	}

	fixes := []rule.RuleFix{
		rule.RuleFixReplaceRange(utils.TrimNodeTextRange(ctx.SourceFile, node), fixed),
	}
	return append(fixes, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)...)
}

func parseFunctions(options []any) []flattenFunction {
	var configured []string
	if len(options) > 0 {
		optionsMap, _ := options[0].(map[string]any)
		configured = utils.ToStringSlice(optionsMap["functions"])
	}
	if len(configured) == 0 {
		return lodashFlattenFunctions
	}

	functions := make([]flattenFunction, 0, len(configured)+len(lodashFlattenFunctions))
	for _, function := range configured {
		name := ecmascript.StringTrim(function)
		leafStart := strings.LastIndexByte(name, '.') + 1
		functions = append(functions, flattenFunction{
			name: name,
			leaf: name[leafStart:],
		})
	}
	return append(functions, lodashFlattenFunctions...)
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/docs/rules/prefer-array-flat.md
var PreferArrayFlatRule = rule.Rule{
	Name:   "unicorn/prefer-array-flat",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		functions := parseFunctions(options)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if match, ok := matchArrayFlatMap(node, ctx); ok {
					reportMatch(node, match, ctx)
				}
				if match, ok := matchArrayReduce(node, ctx); ok {
					reportMatch(node, match, ctx)
				}
				if match, ok := matchEmptyArrayConcat(node); ok {
					reportMatch(node, match, ctx)
				}
				if match, ok := matchArrayPrototypeConcat(node); ok {
					reportMatch(node, match, ctx)
				}
				if match, ok := matchFlattenFunction(node, functions); ok {
					reportMatch(node, match, ctx)
				}
			},
		}
	},
}

func reportMatch(node *ast.Node, match flattenMatch, ctx rule.RuleContext) {
	ruleMessage := message(match.description)
	ctx.ReportNodeWithDeferredFixes(node, ruleMessage, func() []rule.RuleFix {
		if !canFix(node, match.array, ctx) {
			return nil
		}
		return buildFixes(node, match, ctx)
	})
}
