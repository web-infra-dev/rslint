// Ported from eslint-plugin-unicorn v79.0.0. See LICENSE.
package prefer_bigint_literals

import (
	"math"
	"math/big"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/jsnum"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var preferLiteral = rule.RuleMessage{
	Id:          "prefer-bigint-literals/error",
	Description: "Prefer using bigint literal over `BigInt(…)`.",
}

var PreferBigIntLiteralsRule = rule.Rule{
	Name:   "unicorn/prefer-bigint-literals",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		if ctx.LanguageOptions.EffectiveECMAVersion() < 2020 || !ctx.Globals.Access("BigInt").IsDeclared() {
			return nil
		}
		if !ast.IsInJSFile(ctx.SourceFile.AsNode()) {
			if options := ctx.Program().Options(); options != nil && options.GetEmitScriptTarget() < core.ScriptTargetES2020 {
				return nil
			}
		}
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				checkCall(&ctx, node)
			},
		}
	},
}

func checkCall(ctx *rule.RuleContext, node *ast.Node) {
	call := node.AsCallExpression()
	arguments := node.Arguments()
	if call.QuestionDotToken != nil || len(arguments) != 1 || ast.IsSpreadElement(arguments[0]) ||
		len(node.TypeArguments()) != 0 {
		return
	}
	callee := utils.ESTreeCallCallee(call.Expression)
	// A local binding or a with object can provide a different BigInt function.
	if callee == nil || !ast.IsIdentifier(callee) || callee.Text() != "BigInt" ||
		callee.Flags&ast.NodeFlagsInWithStatement != 0 || !unicornutil.IsGlobalReference(*ctx, callee) {
		return
	}

	literal, suggest := bigintLiteral(ctx.SourceFile, utils.ESTreeRuntimeExpression(arguments[0]))
	if literal == "" {
		return
	}
	span := utils.TrimNodeTextRange(ctx.SourceFile, node)
	buildFix := func() []rule.RuleFix {
		text := literal
		if strings.HasPrefix(text, "-") {
			parent := node.Parent
			for utils.IsJSDocTypeCastWrapper(parent) && !ast.IsParenthesizedExpression(parent) {
				parent = parent.Parent
			}
			if parent == nil || !ast.IsParenthesizedExpression(parent) {
				text = "(" + text + ")"
				if unicornutil.NeedsSemicolonBefore(ctx.SourceFile, node, text) {
					text = ";" + text
				}
			}
		}
		text = utils.SafeReplacementText(ctx.SourceFile, node, text)
		return []rule.RuleFix{rule.RuleFixReplaceRange(span, text)}
	}
	if suggest || utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End()) {
		ctx.ReportRangeWithDeferredSuggestions(span, preferLiteral, func() []rule.RuleSuggestion {
			replacement := "a bigint literal"
			if len(literal) < 20 {
				replacement = "`" + literal + "`"
			}
			return []rule.RuleSuggestion{{
				Message: rule.RuleMessage{
					Id:          "prefer-bigint-literals/suggestion",
					Description: "Replace with " + replacement + ".",
					Data:        map[string]string{"replacement": replacement},
				},
				FixesArr: buildFix(),
			}}
		})
		return
	}
	ctx.ReportRangeWithDeferredFixes(span, preferLiteral, buildFix)
}

func bigintLiteral(source *ast.SourceFile, node *ast.Node) (string, bool) {
	if value, ok := utils.GetStaticStringLiteralValue(node); ok {
		integer, ok := ecmascript.StringToBigInt(value)
		if !ok {
			return "", false
		}
		text, hasPlus := strings.CutPrefix(ecmascript.StringTrim(value), "+")
		if !matchesPrefixedInteger(text, integer) {
			text = integer.String()
		}
		return text + "n", hasPlus
	}

	var negated, suggest bool
	for node != nil && node.Kind == ast.KindPrefixUnaryExpression {
		unary := node.AsPrefixUnaryExpression()
		switch unary.Operator {
		case ast.KindPlusToken:
			suggest = true
		case ast.KindMinusToken:
			negated = !negated
		default:
			return "", false
		}
		node = utils.ESTreeRuntimeExpression(unary.Operand)
	}
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return "", false
	}
	return numericLiteral(source, node, negated, suggest)
}

func numericLiteral(source *ast.SourceFile, node *ast.Node, negated, suggest bool) (string, bool) {
	raw := utils.TrimmedNodeText(source, node)
	digits := strings.ReplaceAll(raw, "_", "")
	// tsgo already normalizes numeric literals to their JavaScript Number text.
	numberText := node.Text()
	if node.AsNumericLiteral().TokenFlags&ast.TokenFlagsOctal != 0 {
		// The scanner's legacy-octal conversion is limited to int64.
		numberText = "0o" + digits
	}
	number, ok := ecmascript.StringToNumber(numberText)
	if !ok || math.IsInf(number, 0) || math.Trunc(number) != number {
		return "", false
	}
	text := raw
	if digits != numberText || number > float64(jsnum.MaxSafeInteger) {
		// Beyond the safe range, Number's shortest spelling can differ from
		// the exact integer BigInt receives. Check the integer before reusing it.
		integer, _ := new(big.Float).SetFloat64(number).Int(nil)
		if !matchesPrefixedInteger(digits, integer) {
			decimal := integer.String()
			if digits != numberText || digits != decimal {
				text, suggest = decimal, true
			}
		}
	}
	if negated && number != 0 {
		text = "-" + text
	}
	return text + "n", suggest
}

func matchesPrefixedInteger(text string, integer *big.Int) bool {
	if len(text) < 3 || text[0] != '0' {
		return false
	}
	var base int
	switch text[1] {
	case 'b', 'B':
		base = 2
	case 'o', 'O':
		base = 8
	case 'x', 'X':
		base = 16
	default:
		return false
	}
	return ecmascript.EqualsWhenLowercased(text[2:], integer.Text(base))
}
