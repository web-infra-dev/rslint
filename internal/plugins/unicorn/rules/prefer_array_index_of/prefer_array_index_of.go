package prefer_array_index_of

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

type indexReplacement struct {
	method     string
	diagnostic rule.RuleMessage
	suggestion rule.RuleMessage
}

var indexOf = indexReplacement{
	method: "indexOf",
	diagnostic: rule.RuleMessage{
		Id:          "prefer-indexOf-over-findIndex/error",
		Description: "Use `.indexOf()` instead of `.findIndex()` when looking for the index of an item.",
	},
	suggestion: rule.RuleMessage{
		Id:          "prefer-indexOf-over-findIndex/suggestion",
		Description: "Replace `.findIndex()` with `.indexOf()`.",
	},
}

var lastIndexOf = indexReplacement{
	method: "lastIndexOf",
	diagnostic: rule.RuleMessage{
		Id:          "prefer-lastIndexOf-over-findLastIndex/error",
		Description: "Use `.lastIndexOf()` instead of `.findLastIndex()` when looking for the index of an item.",
	},
	suggestion: rule.RuleMessage{
		Id:          "prefer-lastIndexOf-over-findLastIndex/suggestion",
		Description: "Replace `.findLastIndex()` with `.lastIndexOf()`.",
	},
}

var PreferArrayIndexOfRule = rule.Rule{
	Name:   "unicorn/prefer-array-index-of",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		argumentCount := 1
		options := unicornutil.DotMethodCallOptions{
			Methods:             []string{"findIndex", "findLastIndex"},
			ArgumentsLength:     &argumentCount,
			RejectSpreadElement: true,
			AllowOptionalMember: true,
		}
		var evaluator *utils.StaticStringEvaluator
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := unicornutil.MatchDotMethodCall(node, options)
				if !ok {
					return
				}
				argument := node.Arguments()[0]
				callback := utils.ESTreeRuntimeExpression(argument)
				search, ok := matchSearchCallback(callback)
				if !ok || unicornutil.ShouldSkipKnownNonArrayReceiver(ctx, call.Object) {
					return
				}
				// findIndex/findLastIndex visit holes as undefined, but indexOf/lastIndexOf skip them.
				if receiver := utils.SkipAssertionsAndParens(call.Object); receiver != nil && ast.IsArrayLiteralExpression(receiver) &&
					slices.ContainsFunc(receiver.Elements(), ast.IsOmittedExpression) {
					return
				}

				replacement := &indexOf
				if call.Property.Text() == "findLastIndex" {
					replacement = &lastIndexOf
				}
				value := utils.ESTreeRuntimeExpression(search.value)
				if incompatibleSearchValue(ctx, call, replacement.method, value) || search.usesCallbackBindings(ctx, callback) {
					return
				}

				var callbackRange core.TextRange
				var editKind rule.EditDemand
				classified := false
				buildEdits := func(wanted rule.EditDemand) []rule.RuleFix {
					if !classified {
						classified = true
						callbackRange = utils.TrimNodeTextRange(ctx.SourceFile, callback)
						if utils.HasCommentInSpan(ctx.Comments.All(), callbackRange.Pos(), callbackRange.End()) {
							return nil
						}
						// A cast on the callback would become a cast on the search value.
						for wrapper := callback.Parent; wrapper != nil && wrapper != node; wrapper = wrapper.Parent {
							if utils.IsJSDocTypeCastWrapper(wrapper) {
								return nil
							}
						}
						if evaluator == nil {
							evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
						}
						// The replacement evaluates the search value once, even for an empty array.
						editKind = rule.EditDemandAutofix
						if evaluator.HasSideEffect(value, false) {
							editKind = rule.EditDemandSuggestion
						}
					}
					if editKind != wanted {
						return nil
					}
					text := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, value, false)
					// Reuse parentheses already enclosing the callback, if any.
					if ast.IsParenthesizedExpression(search.value) && !ast.IsParenthesizedExpression(argument) {
						text = "(" + text + ")"
					}
					return []rule.RuleFix{
						rule.RuleFixReplace(ctx.SourceFile, call.Property, replacement.method),
						rule.RuleFixReplaceRange(callbackRange, text),
					}
				}
				ctx.ReportNodeWithDeferredFixesAndSuggestions(call.Property, replacement.diagnostic,
					func() []rule.RuleFix { return buildEdits(rule.EditDemandAutofix) },
					func() []rule.RuleSuggestion {
						fixes := buildEdits(rule.EditDemandSuggestion)
						if len(fixes) == 0 {
							return nil
						}
						return []rule.RuleSuggestion{{
							Message:  replacement.suggestion,
							FixesArr: fixes,
						}}
					},
				)
			},
		}
	},
}

type searchCallback struct {
	parameter *ast.Node
	compared  *ast.Node
	value     *ast.Node // Keep the raw operand so the edit can retain grouping.
}

// matchSearchCallback matches a single-parameter callback whose entire result
// is a strict comparison.
func matchSearchCallback(callback *ast.Node) (searchCallback, bool) {
	if callback == nil || (callback.Kind != ast.KindArrowFunction && callback.Kind != ast.KindFunctionExpression) ||
		ast.IsAsyncFunction(callback) {
		return searchCallback{}, false
	}
	if callback.Kind == ast.KindFunctionExpression && callback.AsFunctionExpression().AsteriskToken != nil {
		return searchCallback{}, false
	}
	parameters := callback.Parameters()
	// JSDoc @this prepends a synthetic parameter.
	if len(parameters) > 0 && utils.IsJSDocSyntaxNode(parameters[0]) {
		parameters = utils.ESTreeParameters(callback)
	}
	if len(parameters) != 1 {
		return searchCallback{}, false
	}
	parameter := parameters[0]
	name := unicornutil.PlainParameterIdentifier(parameter)
	if name == nil {
		return searchCallback{}, false
	}
	body := callback.Body()
	if body == nil {
		return searchCallback{}, false
	}
	if body.Kind == ast.KindBlock {
		statements := body.Statements()
		if len(statements) != 1 || statements[0].Kind != ast.KindReturnStatement {
			return searchCallback{}, false
		}
		body = statements[0].AsReturnStatement().Expression
	}
	comparison := utils.ESTreeRuntimeExpression(body)
	if comparison == nil || !ast.IsBinaryExpression(comparison) {
		return searchCallback{}, false
	}
	binary := comparison.AsBinaryExpression()
	if binary.OperatorToken.Kind != ast.KindEqualsEqualsEqualsToken {
		return searchCallback{}, false
	}
	compared, value := utils.ESTreeRuntimeExpression(binary.Left), binary.Right
	if !unicornutil.IsSameIdentifier(compared, name) {
		compared, value = utils.ESTreeRuntimeExpression(binary.Right), binary.Left
		if !unicornutil.IsSameIdentifier(compared, name) {
			return searchCallback{}, false
		}
	}
	return searchCallback{parameter: parameter, compared: compared, value: value}, value != nil
}

func (search searchCallback) usesCallbackBindings(ctx rule.RuleContext, callback *ast.Node) bool {
	if slices.ContainsFunc(ctx.Refs.References(search.parameter.Symbol()), func(ref *ast.Node) bool { return ref != search.compared }) {
		return true
	}
	if callback.Kind == ast.KindFunctionExpression &&
		((callback.Name() != nil && len(ctx.Refs.References(callback.Symbol())) > 0) || usesCallbackContext(ctx, callback, search.value)) {
		return true
	}
	for _, parameter := range utils.ESTreeTypeParameters(callback) {
		for _, ref := range ctx.Refs.References(parameter.Symbol()) {
			if ref.Loc.ContainedBy(search.value.Loc) {
				return true
			}
		}
	}
	return false
}

// Moving a value out of an ordinary function must not rebind its this,
// arguments or new.target, including when a nested arrow captures them.
func usesCallbackContext(ctx rule.RuleContext, callback, value *ast.Node) bool {
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		switch node.Kind {
		case ast.KindThisKeyword:
			return ast.GetThisContainer(node, false, false) == callback
		case ast.KindMetaProperty:
			return node.AsMetaProperty().KeywordToken == ast.KindNewKeyword && ast.GetNewTargetContainer(node) == callback
		case ast.KindIdentifier:
			if node.Text() == "arguments" && scope.IsReferenceIdentifier(node) && ast.GetThisContainer(node, false, false) == callback {
				binding := ctx.Refs.ResolveInFileWithMeaning(node, ast.SymbolFlagsValue)
				return binding == nil || len(binding.Declarations) == 0
			}
		}
		return node.ForEachChild(visit)
	}
	return visit(value)
}

func incompatibleSearchValue(ctx rule.RuleContext, call unicornutil.DotMethodCall, replacement string, value *ast.Node) bool {
	checker := ctx.TypeChecker
	if checker == nil {
		return false
	}
	receiver := checker.GetNonNullableType(checker.GetTypeAtLocation(call.Object))
	method := checker.GetPropertyOfType(receiver, replacement)
	if method == nil {
		return false
	}
	signatures := utils.GetCallSignatures(checker, checker.GetTypeOfSymbolAtLocation(method, call.Call))
	if len(signatures) == 0 || len(signatures[0].Parameters()) == 0 {
		return false
	}
	searchType := checker.GetTypeOfSymbolAtLocation(signatures[0].Parameters()[0], call.Call)
	return !checker.IsTypeAssignableTo(checker.GetTypeAtLocation(value), searchType)
}
