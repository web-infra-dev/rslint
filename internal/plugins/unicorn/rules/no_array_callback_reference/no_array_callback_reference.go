// Ported from eslint-plugin-unicorn v76.0.0; see LICENSE.
// cspell:ignore metadatum sses ches zzes
package no_array_callback_reference

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_array_callback_reference.schema.json
var schemaJSON []byte

var defaultIgnoredCallees = []string{
	"Promise", "React.Children", "Children", "lodash", "underscore", "_",
	"Async", "async", "this", "$", "jQuery",
}

var iteratorMethods = []string{
	"every", "filter", "find", "findLast", "findIndex", "findLastIndex",
	"flatMap", "forEach", "map", "reduce", "reduceRight", "some",
}

var NoArrayCallbackReferenceRule = rule.Rule{
	Name:   "unicorn/no-array-callback-reference",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ignored := slices.Clone(defaultIgnoredCallees)
		if len(options) > 0 {
			if object, ok := options[0].(map[string]any); ok {
				if names, ok := object["ignore"].([]any); ok {
					for _, name := range names {
						if name, ok := name.(string); ok {
							ignored = append(ignored, name)
						}
					}
				}
			}
		}
		minimum, maximum := 1, 2
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, unicornutil.DotMethodCallOptions{
					Methods: iteratorMethods, MinimumArguments: &minimum, MaximumArguments: &maximum,
					RejectSpreadElement: true, AllowOptionalMember: true,
				})
				if !ok {
					return
				}
				method := call.Property.Text()
				if parent := utils.ESTreeParent(node); !isReduce(method) && parent != nil && parent.Kind == ast.KindAwaitExpression && !ast.IsOptionalChain(node) {
					return
				}
				object := utils.ESTreeRuntimeExpression(call.Object)
				// Parentheses around an optional factory call terminate its chain;
				// within the same chain ESTree still exposes a CallExpression.
				isFactoryCall := ast.IsCallExpression(object) && (!ast.IsOptionalChain(object) || call.Object == object)
				for _, name := range ignored {
					if unicornutil.NodeMatchesPath(object, name) ||
						(isFactoryCall && unicornutil.NodeMatchesPath(object.Expression(), name)) {
						return
					}
				}
				if ast.IsIdentifier(object) && ((method == "filter" && object.Text() == "Vue") || (method == "map" && object.Text() == "types")) {
					return
				}

				var callbacks []*ast.Node
				var collect func(*ast.Node)
				collect = func(callback *ast.Node) {
					callback = utils.ESTreeRuntimeExpression(callback)
					if ast.IsConditionalExpression(callback) {
						collect(callback.AsConditionalExpression().WhenTrue)
						collect(callback.AsConditionalExpression().WhenFalse)
					} else if !ignoreCallback(ctx, callback, method) {
						callbacks = append(callbacks, callback)
					}
				}
				collect(node.Arguments()[0])
				if len(callbacks) == 0 || unicornutil.IsKnownNonIndexedCollectionWithOptions(ctx, object, unicornutil.ArrayReceiverOptions{
					AllowNullishInMixedUnion: true, CheckClassHeritage: true,
					CheckClassSyntax: true, TreatMixedUnionAsNonTarget: true,
				}) {
					return
				}
				for _, callback := range callbacks {
					name := ""
					message := rule.RuleMessage{Id: "error-without-name", Description: fmt.Sprintf("Do not pass function directly to `.%s(…)`.", method)}
					if ast.IsIdentifier(callback) {
						name = callback.Text()
						message = rule.RuleMessage{Id: "error-with-name", Description: fmt.Sprintf("Do not pass function `%s` directly to `.%s(…)`.", name, method)}
					}
					message.Data = map[string]string{"name": name, "method": method}
					if callback.Kind == ast.KindYieldExpression || callback.Kind == ast.KindAwaitExpression {
						ctx.ReportNode(callback, message)
						continue
					}
					ctx.ReportNodeWithDeferredSuggestions(callback, message, func() []rule.RuleSuggestion {
						return callbackSuggestions(ctx, callback, object, method, name)
					})
				}
			},
		}
	},
}

func isReduce(method string) bool {
	return method == "reduce" || method == "reduceRight"
}

func ignoreCallback(ctx rule.RuleContext, node *ast.Node, method string) bool {
	if ast.IsFunctionExpression(node) || ast.IsArrowFunction(node) ||
		(ast.IsCallExpression(node) && !ast.IsOptionalChain(node) && !ast.IsImportCall(node)) ||
		unicornutil.IsNodeValueNotFunction(node) || isConstNonFunction(ctx, node) {
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	name := node.Text()
	if method == "map" && slices.Contains([]string{"String", "Number", "BigInt", "Boolean", "Symbol"}, name) {
		return true
	}
	if name == "Boolean" && slices.Contains([]string{"every", "filter", "find", "findLast", "findIndex", "findLastIndex", "some"}, method) {
		return true
	}
	return slices.Contains([]string{"every", "filter", "find", "findLast"}, method) && isTypePredicateCallback(ctx, node)
}

// Follow only const identifier aliases. A mutable binding or a factory result
// cannot establish that a reference will never hold a function.
func isConstNonFunction(ctx rule.RuleContext, node *ast.Node) bool {
	visited := map[*ast.Symbol]bool{}
	for node != nil {
		node = utils.SkipAssertionsAndParens(node)
		if utils.IsESTreeLiteralKind(node.Kind) {
			return true
		}
		switch node.Kind {
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindClassExpression,
			ast.KindTemplateExpression, ast.KindNoSubstitutionTemplateLiteral,
			ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
			ast.KindVoidExpression, ast.KindTypeOfExpression, ast.KindDeleteExpression:
			return true
		case ast.KindBinaryExpression:
			op := node.AsBinaryExpression().OperatorToken.Kind
			return !ast.IsAssignmentOperator(op) && op != ast.KindCommaToken &&
				op != ast.KindAmpersandAmpersandToken && op != ast.KindBarBarToken && op != ast.KindQuestionQuestionToken
		}
		if utils.IsUndefinedIdentifier(node) {
			return true
		}
		if !ast.IsIdentifier(node) || ctx.Refs == nil {
			return false
		}
		symbol := ctx.Refs.ResolveInFile(node)
		if symbol == nil || visited[symbol] || len(symbol.Declarations) != 1 {
			return false
		}
		visited[symbol] = true
		declaration := symbol.Declarations[0]
		if !ast.IsVariableDeclaration(declaration) || !ast.IsIdentifier(declaration.Name()) ||
			declaration.Name().Text() != node.Text() || !ast.IsVarConst(declaration) {
			return false
		}
		node = declaration.AsVariableDeclaration().Initializer
	}
	return false
}

func isTypePredicateCallback(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.Refs == nil {
		return false
	}
	symbol := ctx.Refs.ResolveInFile(node)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	// Merged symbols can list a function before an earlier interface/namespace.
	// ESLint's first scope definition follows source order.
	declaration := slices.MinFunc(symbol.Declarations, func(a, b *ast.Node) int { return a.Pos() - b.Pos() })
	switch declaration.Kind {
	case ast.KindImportClause, ast.KindImportSpecifier, ast.KindNamespaceImport, ast.KindImportEqualsDeclaration:
		return true
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		return declaration.Type() != nil && declaration.Type().Kind == ast.KindTypePredicate
	case ast.KindVariableDeclaration, ast.KindParameter:
		annotation := declaration.Type()
		if annotation != nil {
			annotation = ast.SkipTypeParentheses(annotation)
		}
		if annotation != nil && annotation.Kind == ast.KindFunctionType &&
			annotation.Type() != nil && annotation.Type().Kind == ast.KindTypePredicate {
			return true
		}
		if ast.IsVariableDeclaration(declaration) {
			initializer := utils.ESTreeRuntimeExpression(declaration.AsVariableDeclaration().Initializer)
			return initializer != nil && (ast.IsArrowFunction(initializer) || ast.IsFunctionExpression(initializer)) &&
				initializer.Type() != nil && initializer.Type().Kind == ast.KindTypePredicate
		}
	}
	return false
}

func callbackSuggestions(ctx rule.RuleContext, callback, object *ast.Node, method, name string) []rule.RuleSuggestion {
	parameters := []string{"element", "index", "array"}
	minimum := 1
	if isReduce(method) {
		parameters = append([]string{"accumulator"}, parameters...)
		minimum = 2
	}
	if name != "" {
		parameters = suggestionParameters(parameters, object, name)
	}
	outer := utils.OutermostParenthesizedExpression(callback)
	for outer.Parent != nil && utils.IsJSDocTypeCastWrapper(outer.Parent) {
		outer = utils.OutermostParenthesizedExpression(outer.Parent)
	}
	text := utils.TrimmedNodeText(ctx.SourceFile, outer)
	// Instantiation expressions (fn<T>) already form a call's callee. For the
	// other reportable shapes, ESLint precedence also handles optional chains
	// ending in a non-null assertion without introducing extra parentheses.
	if outer == callback && callback.Kind != ast.KindExpressionWithTypeArguments && utils.EslintLikePrecedence(callback) < 18 {
		text = "(" + text + ")"
	}
	var suggestions []rule.RuleSuggestion
	for count := minimum; count <= len(parameters); count++ {
		args := strings.Join(parameters[:count], ", ")
		replacement := "(" + args + ") => " + text + "(" + args + ")"
		if method == "forEach" {
			replacement = "(" + args + ") => { " + text + "(" + args + "); }"
		}
		message := rule.RuleMessage{Id: "replace-without-name", Description: fmt.Sprintf("Replace function with `… => …(%s)`.", args)}
		if name != "" {
			message = rule.RuleMessage{Id: "replace-with-name", Description: fmt.Sprintf("Replace function `%s` with `… => %s(%s)`.", name, name, args)}
		}
		message.Data = map[string]string{"name": name, "parameters": args}
		suggestions = append(suggestions, rule.RuleSuggestion{Message: message, FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, outer, replacement)}})
	}
	return suggestions
}

func suggestionParameters(parameters []string, object *ast.Node, callback string) []string {
	result := slices.Clone(parameters)
	if ast.IsIdentifier(object) {
		if element := singular(object.Text()); element != "" {
			for index, parameter := range parameters {
				replacement := ""
				switch parameter {
				case "element":
					replacement = element
				case "array":
					replacement = object.Text()
				}
				if replacement != "" && replacement != callback && !slices.Contains(parameters, replacement) &&
					isParameterName(replacement) {
					result[index] = replacement
				}
			}
		}
	}
	for index, parameter := range result {
		if parameter == callback {
			result[index] += "_"
		}
	}
	return result
}

func isParameterName(name string) bool {
	if !scanner.IsValidIdentifier(name) {
		return false
	}
	keyword := scanner.StringToToken(name)
	if keyword >= ast.KindFirstReservedWord && keyword <= ast.KindLastFutureReservedWord {
		return false
	}
	switch name {
	case "await", "let", "static", "yield", "arguments", "eval", "globalThis", "Infinity", "NaN", "undefined":
		return false
	}
	return true
}

// Parameter names are cosmetic. Keep common English inflections here rather
// than importing a general natural-language inflection engine.
func singular(name string) string {
	switch name {
	case "people":
		return "person"
	case "children":
		return "child"
	case "men":
		return "man"
	case "women":
		return "woman"
	case "mice":
		return "mouse"
	case "geese":
		return "goose"
	case "teeth":
		return "tooth"
	case "feet":
		return "foot"
	case "indices":
		return "index"
	case "matrices":
		return "matrix"
	case "vertices":
		return "vertex"
	case "data":
		return "datum"
	case "metadata":
		return "metadatum"
	case "movies":
		return "movie"
	case "media", "news", "series", "species", "status", "alias", "bus", "analysis", "basis", "axis":
		return ""
	}
	if strings.HasSuffix(name, "ies") && len(name) > 3 {
		return strings.TrimSuffix(name, "ies") + "y"
	}
	for _, suffix := range []string{"sses", "shes", "ches", "xes", "zzes", "statuses", "aliases", "buses"} {
		if strings.HasSuffix(name, suffix) {
			return strings.TrimSuffix(name, "es")
		}
	}
	if strings.HasSuffix(name, "s") && !strings.HasSuffix(name, "ss") && len(name) > 1 {
		return strings.TrimSuffix(name, "s")
	}
	return ""
}
