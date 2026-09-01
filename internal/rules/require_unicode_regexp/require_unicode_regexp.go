package require_unicode_regexp

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed require_unicode_regexp.schema.json
var schemaJSON []byte

func parseOptions(options []any) string {
	if len(options) == 0 {
		return ""
	}
	optsMap, _ := options[0].(map[string]any)
	if requireFlag, ok := optsMap["requireFlag"].(string); ok {
		return requireFlag
	}
	return ""
}

// checkFlags reports whether flags is missing the flag the option requires:
// exactly "u" or "v" when requireFlag pins one, otherwise either.
func checkFlags(requireFlag string, flags string) bool {
	switch requireFlag {
	case "v":
		return !strings.Contains(flags, "v")
	case "u":
		return !strings.Contains(flags, "u")
	default:
		return !strings.Contains(flags, "u") && !strings.Contains(flags, "v")
	}
}

func requireMessage(requireFlag string) rule.RuleMessage {
	if requireFlag == "v" {
		return rule.RuleMessage{Id: "requireVFlag", Description: "Use the 'v' flag."}
	}
	return rule.RuleMessage{Id: "requireUFlag", Description: "Use the 'u' flag."}
}

func addMessage(requireFlag string) rule.RuleMessage {
	if requireFlag == "v" {
		return rule.RuleMessage{Id: "addVFlag", Description: "Add the 'v' flag."}
	}
	return rule.RuleMessage{Id: "addUFlag", Description: "Add the 'u' flag."}
}

// isValidWithUnicodeFlag reports whether pattern would still be a valid
// RegExp pattern under the u/v flag requireFlag resolves to (defaulting to
// "u" when unset), the way JavaScript itself would parse it — not whether it
// currently matches anything.
func isValidWithUnicodeFlag(ecmaVersion int, pattern string, requireFlag string) bool {
	if requireFlag == "v" {
		if ecmaVersion <= 2023 {
			return false
		}
		return utils.IsValidRegexPatternForECMAVersion(pattern, utils.RegexFlags{UnicodeSets: true}, ecmaVersion)
	}
	if ecmaVersion <= 5 {
		return false
	}
	return utils.IsValidRegexPatternForECMAVersion(pattern, utils.RegexFlags{Unicode: true}, ecmaVersion)
}

// buildLiteralFix mirrors upstream's Literal[regex] fixer: it either swaps an
// existing opposite-flavor flag for the required one, or appends the required
// flag after the whole literal.
func buildLiteralFix(ctx rule.RuleContext, node *ast.Node, requireFlag string) (rule.RuleFix, bool) {
	replaceFlag := "u"
	if requireFlag == "v" {
		replaceFlag = "v"
	}

	regex := node.Text()
	slashPos := strings.LastIndex(regex, "/")
	if slashPos == -1 {
		return rule.RuleFix{}, false
	}

	if requireFlag != "" {
		flag := "v"
		if requireFlag == "v" {
			flag = "u"
		}
		if strings.Contains(regex[slashPos:], flag) {
			replacedFlags := strings.Replace(regex[slashPos:], flag, requireFlag, 1)
			return rule.RuleFixReplace(ctx.SourceFile, node, regex[:slashPos]+replacedFlags), true
		}
	}

	return rule.RuleFixInsertAfter(node, replaceFlag), true
}

// flagsNodeShape classifies flagsNode the way upstream's fixer does: a plain
// string literal or template literal is fixable (possibly with substitutions,
// which forces the conservative refusal below); anything else — an
// identifier, a member access, an assignment — is not.
func flagsNodeShape(node *ast.Node) (fixable bool, hasSubstitutions bool) {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return true, false
	case ast.KindTemplateExpression:
		return true, true
	default:
		return false, false
	}
}

// buildCallFix mirrors upstream's RegExp()/new RegExp() fixer. flags is the
// statically-evaluated (cooked) flags value; flagsNode is nil when the call
// has no second argument.
func buildCallFix(ctx rule.RuleContext, refNode *ast.Node, flagsNode *ast.Node, flags string, requireFlag string) (rule.RuleFix, bool) {
	replaceFlag := "u"
	if requireFlag == "v" {
		replaceFlag = "v"
	}

	if flagsNode != nil {
		// ESTree has no parenthesis node, so upstream classifies `('gi')` by
		// the literal inside and its fix keeps the parens in place.
		flagsNode = ast.SkipParentheses(flagsNode)
		fixable, hasSubstitutions := flagsNodeShape(flagsNode)
		if !fixable {
			// We intentionally don't suggest concatenating + "u"/"v" to non-literals.
			return rule.RuleFix{}, false
		}

		flagsNodeText := utils.TrimmedNodeText(ctx.SourceFile, flagsNode)
		if len(flagsNodeText) < 2 {
			return rule.RuleFix{}, false
		}

		flag := "u"
		if requireFlag == "u" {
			flag = "v"
		}
		if strings.Contains(flags, flag) {
			// Avoid replacing "u"/"v" inside an escape like `g`, or a
			// template whose interpolated parts we can't safely rewrite.
			if hasSubstitutions || strings.Contains(flagsNodeText, `\`) {
				return rule.RuleFix{}, false
			}
			newText := strings.Replace(flagsNodeText, flag, replaceFlag, 1)
			return rule.RuleFixReplace(ctx.SourceFile, flagsNode, newText), true
		}

		newText := flagsNodeText[:len(flagsNodeText)-1] + replaceFlag + flagsNodeText[len(flagsNodeText)-1:]
		return rule.RuleFixReplace(ctx.SourceFile, flagsNode, newText), true
	}

	// No second argument: insert after the token preceding the closing paren
	// (skipping over it), the way `sourceCode.getLastToken(refNode, {skip: 1})`
	// does — landing on a trailing comma when one is present, or the last
	// argument otherwise.
	closingParenPos := refNode.End() - 1
	tok, ok := utils.PreviousTokenBefore(ctx.SourceFile, refNode, closingParenPos)
	if !ok {
		return rule.RuleFix{}, false
	}
	insertText := `, "` + replaceFlag + `"`
	if tok.Kind == ast.KindCommaToken {
		insertText = ` "` + replaceFlag + `",`
	}
	return rule.RuleFixReplaceRange(core.NewTextRange(tok.End, tok.End), insertText), true
}

// https://eslint.org/docs/latest/rules/require-unicode-regexp
var RequireUnicodeRegexpRule = rule.Rule{
	Name:   "require-unicode-regexp",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		requireFlag := parseOptions(options)
		listeners := rule.RuleListeners{
			ast.KindRegularExpressionLiteral: func(node *ast.Node) {
				pattern, flags := utils.ExtractRegexPatternAndFlags(node.Text())
				if !checkFlags(requireFlag, flags) {
					return
				}
				ctx.ReportNodeWithDeferredSuggestions(node, requireMessage(requireFlag), func() []rule.RuleSuggestion {
					if !isValidWithUnicodeFlag(ctx.LanguageOptions.EffectiveECMAVersion(), pattern, requireFlag) {
						return nil
					}
					fix, ok := buildLiteralFix(ctx, node, requireFlag)
					if !ok {
						return nil
					}
					return []rule.RuleSuggestion{{Message: addMessage(requireFlag), FixesArr: []rule.RuleFix{fix}}}
				})
			},
		}
		if !sourceMayUseRegexpConstructor(ctx.SourceFile) {
			return listeners
		}

		callTracker := newRegexpCallTracker(ctx)
		var evaluator *utils.StaticStringEvaluator
		getEvaluator := func() *utils.StaticStringEvaluator {
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(ctx.TypeChecker, ctx.SourceFile, ctx.Refs)
			}
			return evaluator
		}

		checkCall := func(node *ast.Node, argsList *ast.NodeList) {
			if !callTracker.isCall(node) {
				return
			}

			var args []*ast.Node
			if argsList != nil {
				args = argsList.Nodes
			}

			var patternNode, flagsNode *ast.Node
			if len(args) > 0 {
				patternNode = args[0]
				if patternNode.Kind == ast.KindSpreadElement {
					return
				}
			}
			if len(args) > 1 {
				flagsNode = args[1]
			}

			// Mirrors upstream's getStringIfConstant: any statically-evaluable
			// value (not just an already-string one) is coerced to its String()
			// form before the flag check, so e.g. a literal `false` or `1`
			// flags argument is compared as the text "false" / "1".
			flagsValue, flagsOk := "", false
			if flagsNode != nil {
				flagsValue, flagsOk = getEvaluator().EvalToString(flagsNode)
			}
			missingFlag := flagsNode == nil
			if flagsOk {
				missingFlag = checkFlags(requireFlag, flagsValue)
			}
			if !missingFlag {
				return
			}

			ctx.ReportNodeWithDeferredSuggestions(node, requireMessage(requireFlag), func() []rule.RuleSuggestion {
				patternValue, patternOk := getEvaluator().EvalToString(patternNode)
				if !patternOk {
					return nil
				}
				if !isValidWithUnicodeFlag(ctx.LanguageOptions.EffectiveECMAVersion(), patternValue, requireFlag) {
					return nil
				}
				fix, ok := buildCallFix(ctx, node, flagsNode, flagsValue, requireFlag)
				if !ok {
					return nil
				}
				return []rule.RuleSuggestion{{Message: addMessage(requireFlag), FixesArr: []rule.RuleFix{fix}}}
			})
		}

		listeners[ast.KindCallExpression] = func(node *ast.Node) {
			call := node.AsCallExpression()
			checkCall(node, call.Arguments)
		}
		listeners[ast.KindNewExpression] = func(node *ast.Node) {
			newExpr := node.AsNewExpression()
			checkCall(node, newExpr.Arguments)
		}
		return listeners
	},
}
