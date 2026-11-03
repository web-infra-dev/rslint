package prefer_string_starts_ends_with

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var messages = map[string]string{
	"prefer-starts-with":         "Prefer `String#startsWith()` over a regex with `^`.",
	"prefer-ends-with":           "Prefer `String#endsWith()` over a regex with `$`.",
	"prefer-starts-with-indexOf": "Prefer `String#startsWith()` over `String#indexOf() === 0`.",
	"prefer-starts-with-slice":   "Prefer `String#startsWith()` over comparing a string slice.",
	"prefer-ends-with-slice":     "Prefer `String#endsWith()` over comparing a string slice.",
}

func message(id string) rule.RuleMessage {
	return rule.RuleMessage{Id: id, Description: messages[id]}
}

// PreferStringStartsEndsWithRule follows eslint-plugin-unicorn v79.0.0.
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-string-starts-ends-with.js
var PreferStringStartsEndsWithRule = rule.Rule{
	Name:   "unicorn/prefer-string-starts-ends-with",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		evaluator := utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
		one, two := 1, 2
		match := func(node *ast.Node, method string, minimum, maximum *int) (unicornutil.DotMethodCall, bool) {
			return unicornutil.MatchDotMethodCall(utils.ESTreeRuntimeExpression(node), unicornutil.DotMethodCallOptions{
				Method: method, MinimumArguments: minimum, MaximumArguments: maximum, RejectSpreadElement: true,
			})
		}
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call, ok := match(node, "test", &one, &one)
				if !ok {
					return
				}
				regex := utils.ESTreeRuntimeExpression(call.Object)
				if regex.Kind != ast.KindRegularExpressionLiteral {
					return
				}
				id, pattern := regexProblem(regex.Text())
				if id == "" {
					return
				}
				target := utils.ESTreeRuntimeExpression(node.Arguments()[0])
				if unicornutil.IsKnownNonStringSyntax(ctx, target) {
					return
				}
				knownString := target.Kind == ast.KindTemplateExpression || target.Kind == ast.KindNoSubstitutionTemplateLiteral
				if ast.IsCallExpression(target) {
					callee := utils.ESTreeCallCallee(target.Expression())
					knownString = callee != nil && ast.IsIdentifier(callee) && callee.Text() == "String" && unicornutil.IsGlobalReference(ctx, callee)
				}
				knownNonString := false
				if !knownString {
					if value, known := evaluator.EvalValueIfNoSideEffects(target); known {
						_, knownString = value.(string)
						knownNonString = !knownString
					}
				}
				method := "startsWith"
				if id == "prefer-ends-with" {
					method = "endsWith"
				}
				fix := func(kind string) []rule.RuleFix {
					text := utils.TrimmedNodeText(ctx.SourceFile, utils.OutermostParenthesizedExpression(target))
					regexWrapped := utils.OutermostParenthesizedExpression(regex) != regex
					targetWrapped := utils.OutermostParenthesizedExpression(target) != target
					var edits []rule.RuleFix
					switch kind {
					case "useStringCasting":
						text = "String(" + text + ")"
					case "useNullishCoalescing":
						if !targetWrapped && needsNullishParentheses(target) {
							text = "(" + text + ")"
						}
						text += " ?? ''"
						if !regexWrapped {
							text = "(" + text + ")"
						}
					default:
						if kind == "useOptionalChaining" {
							dot, _ := utils.TokenBeforePosition(ctx.SourceFile, utils.TrimNodeTextRange(ctx.SourceFile, call.Property).Pos())
							edits = append(edits, rule.RuleFixReplaceRange(core.NewTextRange(dot.Start, dot.End), "?."))
						}
						if !regexWrapped && !targetWrapped && needsMemberParentheses(ctx.SourceFile, target, regex) {
							text = "(" + text + ")"
						}
					}
					edits = append(edits, rule.RuleFixReplace(ctx.SourceFile, regex, text))
					edits = append(edits, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, regex)...)
					edits = append(edits, rule.RuleFixReplace(ctx.SourceFile, call.Property, method),
						rule.RuleFixReplace(ctx.SourceFile, utils.OutermostParenthesizedExpression(target), quotePattern(pattern)))
					return edits
				}
				ctx.ReportNodeWithDeferredFixesAndSuggestions(node, message(id), func() []rule.RuleFix {
					if knownNonString {
						return nil
					}
					return fix("")
				}, func() []rule.RuleSuggestion {
					if knownString {
						return nil
					}
					var suggestions []rule.RuleSuggestion
					var canUseString bool
					if ctx.Refs != nil {
						canUseString = ctx.Refs.IsGlobalNameReference(node, "String", ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias)
					} else {
						canUseString = !utils.IsShadowed(node, "String")
					}
					for _, kind := range []string{"useStringCasting", "useOptionalChaining", "useNullishCoalescing"} {
						if kind == "useStringCasting" && !canUseString {
							continue
						}
						if kind == "useOptionalChaining" && node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression && node.Parent.AsTaggedTemplateExpression().Tag == node {
							continue
						}
						description := map[string]string{
							"useStringCasting":     "Convert to string `String(…).%s()`.",
							"useOptionalChaining":  "Use optional chaining `…?.%s()`.",
							"useNullishCoalescing": "Use nullish coalescing `(… ?? '').%s()`.",
						}[kind]
						suggestions = append(suggestions, rule.RuleSuggestion{
							Message:  rule.RuleMessage{Id: kind, Description: fmt.Sprintf(description, method), Data: map[string]string{"method": method}},
							FixesArr: fix(kind),
						})
					}
					return suggestions
				})
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				operator := binary.OperatorToken.Kind
				if !ast.IsEqualityOperator(operator) {
					return
				}
				left, right := utils.ESTreeRuntimeExpression(binary.Left), utils.ESTreeRuntimeExpression(binary.Right)
				call, indexOf := match(left, "indexOf", &one, &one)
				if !indexOf || !isZero(right) {
					call, indexOf = match(right, "indexOf", &one, &one)
					indexOf = indexOf && isZero(left)
				}
				method, id := "startsWith", "prefer-starts-with-indexOf"
				var search *ast.Node
				var removedEvaluation *ast.Node
				reversed := false
				if indexOf {
					search = call.Call.Arguments()[0]
				} else {
					var ok bool
					call, ok = match(left, "slice", &one, &two)
					search = right
					if !ok {
						call, ok = match(right, "slice", &one, &two)
						search, reversed = left, true
					}
					if !ok || !unicornutil.IsString(ctx, search, evaluator) {
						return
					}
					method, removedEvaluation = sliceMethod(call.Call.Arguments(), search, evaluator)
					if method == "" {
						return
					}
					id = "prefer-starts-with-slice"
					if method == "endsWith" {
						id = "prefer-ends-with-slice"
					}
				}
				target := utils.ESTreeRuntimeExpression(call.Object)
				if !unicornutil.IsString(ctx, target, evaluator) {
					return
				}
				negated := operator == ast.KindExclamationEqualsToken || operator == ast.KindExclamationEqualsEqualsToken
				fix := func() []rule.RuleFix {
					if (indexOf && !unicornutil.IsString(ctx, search, evaluator)) || removesComments(ctx, node, target, search) {
						return nil
					}
					if removedEvaluation != nil && evaluator.HasSideEffect(removedEvaluation, true) {
						return nil
					}
					text := utils.TrimmedNodeText(ctx.SourceFile, utils.OutermostParenthesizedExpression(target))
					position := node
					if negated {
						position = target
					}
					if utils.OutermostParenthesizedExpression(target) == target && needsMemberParentheses(ctx.SourceFile, target, position) {
						text = "(" + text + ")"
					}
					if negated {
						text = "!" + text
					}
					text += "." + method + "(" + utils.TrimmedNodeText(ctx.SourceFile, utils.OutermostParenthesizedExpression(search)) + ")"
					if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, text) {
						text = ";" + text
					}
					return append([]rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, text)}, unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)...)
				}
				if reversed && hasReceiverSideEffect(target, evaluator) {
					ctx.ReportNodeWithDeferredSuggestions(node, message(id), func() []rule.RuleSuggestion {
						fixes := fix()
						if len(fixes) == 0 {
							return nil
						}
						return []rule.RuleSuggestion{{Message: rule.RuleMessage{
							Id: "prefer-string-starts-ends-with/slice-suggestion", Description: fmt.Sprintf("Replace with `String#%s()`.", method), Data: map[string]string{"method": method},
						}, FixesArr: fixes}}
					})
				} else {
					ctx.ReportNodeWithDeferredFixes(node, message(id), fix)
				}
			},
		}
	},
}

func regexProblem(literal string) (string, string) {
	slash := strings.LastIndexByte(literal, '/')
	if slash < 1 || strings.ContainsAny(literal[slash+1:], "im") {
		return "", ""
	}
	pattern, flags := literal[1:slash], literal[slash+1:]
	if strings.HasPrefix(pattern, "^") && !strings.ContainsAny(pattern[1:], "^$+[{(\\.?*|") {
		return "prefer-starts-with", pattern[1:]
	}
	if strings.HasSuffix(pattern, "$") && !strings.ContainsRune(flags, 'y') && !strings.ContainsAny(pattern[:len(pattern)-1], "^$+[{(\\.?*|") {
		return "prefer-ends-with", pattern[:len(pattern)-1]
	}
	return "", ""
}

func hasReceiverSideEffect(node *ast.Node, evaluator *utils.StaticStringEvaluator) bool {
	if evaluator.HasSideEffect(node, true) {
		return true
	}
	// Match upstream's conservative treatment of spreads and tagged templates;
	// these may execute custom code the static evaluator has not observed.
	found := false
	utils.VisitDescendants(node, func(child *ast.Node) bool {
		found = found || child.Kind == ast.KindSpreadElement || child.Kind == ast.KindSpreadAssignment || child.Kind == ast.KindTaggedTemplateExpression
		return !found
	})
	return found
}

func isZero(node *ast.Node) bool {
	return node.Kind == ast.KindNumericLiteral && jsnum.FromString(node.Text()) == 0
}

func numericLiteral(node *ast.Node) (jsnum.Number, bool) {
	node = utils.ESTreeRuntimeExpression(node)
	negative := false
	if node.Kind == ast.KindPrefixUnaryExpression && node.AsPrefixUnaryExpression().Operator == ast.KindMinusToken {
		negative, node = true, utils.ESTreeRuntimeExpression(node.AsPrefixUnaryExpression().Operand)
	}
	if node.Kind != ast.KindNumericLiteral {
		return 0, false
	}
	value := jsnum.FromString(node.Text())
	if negative {
		value = -value
	}
	return value, true
}

func matchingLength(node, search *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node.Kind == ast.KindPropertyAccessExpression && !ast.IsOptionalChain(node) &&
		node.Name().Kind == ast.KindIdentifier && node.Name().Text() == "length" &&
		unicornutil.IsSameReference(node.Expression(), search)
}

// The second result is the expression evaluated only by a removed length
// bound. Checking the receiver, rather than its intrinsic string length, keeps
// plain identifier bounds fixable without dropping getter evaluations.
func sliceMethod(args []*ast.Node, search *ast.Node, evaluator *utils.StaticStringEvaluator) (string, *ast.Node) {
	value, known := evaluator.EvalValueIfNoSideEffects(search)
	text, stringValue := value.(string)
	length := jsnum.Number(ecmascript.StringCodeUnitCount(text))
	if len(args) == 2 && isZero(utils.ESTreeRuntimeExpression(args[0])) {
		bound, numeric := numericLiteral(args[1])
		if known && stringValue && numeric && bound == length {
			return "startsWith", nil
		}
		if matchingLength(args[1], search) {
			return "startsWith", utils.ESTreeRuntimeExpression(args[1]).Expression()
		}
	}
	if len(args) == 1 && known && stringValue && length > 0 {
		bound, numeric := numericLiteral(args[0])
		first := utils.ESTreeRuntimeExpression(args[0])
		if numeric && bound == -length {
			return "endsWith", nil
		}
		if first.Kind == ast.KindPrefixUnaryExpression && first.AsPrefixUnaryExpression().Operator == ast.KindMinusToken && matchingLength(first.AsPrefixUnaryExpression().Operand, search) {
			return "endsWith", utils.ESTreeRuntimeExpression(first.AsPrefixUnaryExpression().Operand).Expression()
		}
	}
	return "", nil
}

func removesComments(ctx rule.RuleContext, node, target, search *ast.Node) bool {
	ranges := []core.TextRange{utils.TrimNodeTextRange(ctx.SourceFile, utils.OutermostParenthesizedExpression(target)), utils.TrimNodeTextRange(ctx.SourceFile, utils.OutermostParenthesizedExpression(search))}
	start := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
	comments := ctx.Comments.All()
	first := sort.Search(len(comments), func(index int) bool { return comments[index].Pos() >= start })
	for _, comment := range comments[first:] {
		if comment.Pos() >= node.End() {
			break
		}
		if comment.End() > node.End() {
			continue
		}
		if (comment.Pos() < ranges[0].Pos() || comment.End() > ranges[0].End()) && (comment.Pos() < ranges[1].Pos() || comment.End() > ranges[1].End()) {
			return true
		}
	}
	return false
}

func needsNullishParentheses(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindAwaitExpression, ast.KindConditionalExpression, ast.KindYieldExpression, ast.KindArrowFunction:
		return true
	case ast.KindBinaryExpression:
		op := node.AsBinaryExpression().OperatorToken.Kind
		return op != ast.KindQuestionQuestionToken
	default:
		return false
	}
}

func needsMemberParentheses(source *ast.SourceFile, node, replacement *ast.Node) bool {
	first, _ := utils.TokenAtOrAfter(source, utils.TrimNodeTextRange(source, node).Pos())
	next, _ := utils.TokenAtOrAfter(source, first.End)
	if first.Kind == ast.KindOpenBraceToken || first.Kind == ast.KindFunctionKeyword || first.Kind == ast.KindClassKeyword ||
		(first.Text == "async" && next.Kind == ast.KindFunctionKeyword) ||
		(first.Text == "let" && next.Kind == ast.KindOpenBracketToken) {
		for current := replacement; current.Parent != nil; current = current.Parent {
			parent := current.Parent
			if ast.IsParenthesizedExpression(parent) {
				break
			}
			if parent.Kind == ast.KindExpressionStatement || (parent.Kind == ast.KindArrowFunction && parent.Body() == current && first.Kind == ast.KindOpenBraceToken) {
				return true
			}
			if utils.TrimNodeTextRange(source, parent).Pos() != utils.TrimNodeTextRange(source, current).Pos() {
				break
			}
		}
	}
	switch node.Kind {
	case ast.KindFunctionExpression, ast.KindObjectLiteralExpression:
		return true
	case ast.KindTaggedTemplateExpression, ast.KindNonNullExpression, ast.KindMetaProperty:
		return false
	case ast.KindCallExpression:
		return false
	default:
		return unicornutil.ShouldAddParenthesesToMemberExpressionObject(source, node)
	}
}

// The accepted patterns have no regex escapes; quote their literal characters
// with the same delimiter and control-character escapes as the upstream fixer.
func quotePattern(pattern string) string {
	var text strings.Builder
	text.WriteByte('\'')
	for _, character := range pattern {
		switch character {
		case '\'':
			text.WriteString(`\'`)
		case '\\':
			text.WriteString(`\\`)
		case '\n':
			text.WriteString(`\n`)
		case '\r':
			text.WriteString(`\r`)
		case '\t':
			text.WriteString(`\t`)
		case '\b':
			text.WriteString(`\b`)
		case '\f':
			text.WriteString(`\f`)
		case '\v':
			text.WriteString(`\v`)
		default:
			if character <= 0x1f || character == 0x7f || character == 0x2028 || character == 0x2029 {
				fmt.Fprintf(&text, `\u{%x}`, character)
			} else {
				text.WriteRune(character)
			}
		}
	}
	text.WriteByte('\'')
	return text.String()
}
