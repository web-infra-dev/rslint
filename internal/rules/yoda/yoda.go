package yoda

import (
	_ "embed"
	"fmt"
	"math"
	"math/big"
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed yoda.schema.json
var schemaJSON []byte

type yodaOptions struct {
	always       bool
	exceptRange  bool
	onlyEquality bool
}

func parseOptions(options []any) yodaOptions {
	opts := yodaOptions{}
	if len(options) == 0 {
		return opts
	}
	if when, ok := options[0].(string); ok {
		opts.always = when == "always"
	}
	if len(options) < 2 {
		return opts
	}
	optsMap, ok := options[1].(map[string]any)
	if !ok {
		return opts
	}
	if v, ok := optsMap["exceptRange"].(bool); ok {
		opts.exceptRange = v
	}
	if v, ok := optsMap["onlyEquality"].(bool); ok {
		opts.onlyEquality = v
	}
	return opts
}

// isComparisonOperator matches ESLint's isComparisonOperator: the operators
// yoda cares about. `in`/`instanceof` share the BinaryExpression kind in tsgo
// but are excluded, as they are there.
func isComparisonOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken,
		ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken,
		ast.KindLessThanToken, ast.KindGreaterThanToken,
		ast.KindLessThanEqualsToken, ast.KindGreaterThanEqualsToken:
		return true
	}
	return false
}

// isEqualityOperator matches ESLint's isEqualityOperator: only `==`/`===`.
// `!=`/`!==` are comparison operators but not equality operators, so
// onlyEquality does not exempt them.
func isEqualityOperator(kind ast.Kind) bool {
	return kind == ast.KindEqualsEqualsToken || kind == ast.KindEqualsEqualsEqualsToken
}

// isRangeTestOperator matches ESLint's isRangeTestOperator: only `<`/`<=`
// build "ascending" range tests (`LOW <= x && x < HIGH`); `>`/`>=` do not.
func isRangeTestOperator(kind ast.Kind) bool {
	return kind == ast.KindLessThanToken || kind == ast.KindLessThanEqualsToken
}

// flipOperatorKind mirrors ESLint's OPERATOR_FLIP_MAP: relational operators
// swap direction; equality operators are symmetric and stay the same.
func flipOperatorKind(kind ast.Kind) ast.Kind {
	switch kind {
	case ast.KindLessThanToken:
		return ast.KindGreaterThanToken
	case ast.KindGreaterThanToken:
		return ast.KindLessThanToken
	case ast.KindLessThanEqualsToken:
		return ast.KindGreaterThanEqualsToken
	case ast.KindGreaterThanEqualsToken:
		return ast.KindLessThanEqualsToken
	default:
		return kind
	}
}

// isLiteralKind reports whether kind is one of ESTree's `Literal` kinds
// (string, number, bigint, boolean, null, regex), which tsgo splits across
// several node kinds.
func isLiteralKind(kind ast.Kind) bool {
	switch kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindRegularExpressionLiteral:
		return true
	}
	return false
}

// isNegativeNumericLike reports whether node is `-<number>` / `-<bigint>`,
// matching ESLint's isNegativeNumericLiteral.
func isNegativeNumericLike(node *ast.Node) bool {
	if node.Kind != ast.KindPrefixUnaryExpression {
		return false
	}
	unary := node.AsPrefixUnaryExpression()
	if unary.Operator != ast.KindMinusToken {
		return false
	}
	operand := ast.SkipParentheses(unary.Operand)
	return operand.Kind == ast.KindNumericLiteral || operand.Kind == ast.KindBigIntLiteral
}

// isLiteralLike reports whether node (already parenthesis-unwrapped) should
// be treated as a literal for yoda's purposes: a real literal, a negative
// numeric/bigint literal (`-1`, `-1n`), or a static template literal without
// substitutions (“ `red` “). Mirrors ESLint's looksLikeLiteral +
// getNormalizedLiteral combination.
func isLiteralLike(node *ast.Node) bool {
	return isLiteralKind(node.Kind) ||
		isNegativeNumericLike(node) ||
		node.Kind == ast.KindNoSubstitutionTemplateLiteral
}

// literalValue is a comparable range-test bound, mirroring ESLint's
// getNormalizedLiteral: it only needs to support `<=` comparison, not the
// full ESTree Literal shape.
type literalValue struct {
	kind string // "number" | "string" | "bigint"
	num  float64
	str  string
	big  *big.Int
}

// lessOrEqual reports whether v <= other, matching JS's abstract relational
// comparison: two strings compare lexicographically, everything else
// compares numerically (ToNumber-like coercion). A non-numeric string
// bound (e.g. `'a'` compared against a numeric bound) coerces to NaN, and
// any comparison against NaN is false — Go's float64 already has that
// behavior built in, so no special case is needed here.
func (v literalValue) lessOrEqual(other literalValue) bool {
	if v.kind == "string" && other.kind == "string" {
		return ecmascript.CompareStrings(v.str, other.str) <= 0
	}
	if v.kind == "bigint" && other.kind == "bigint" {
		return v.big.Cmp(other.big) <= 0
	}
	// A BigInt/Number pair must compare exactly: JS's abstract relational
	// comparison never rounds the BigInt through a float64 (unlike
	// numeric()'s general fallback below), so a bound beyond 2^53 doesn't
	// silently lose precision against its counterpart.
	if v.kind == "bigint" && other.kind == "number" {
		cmp, ok := compareBigIntFloat(v.big, other.num)
		return ok && cmp <= 0
	}
	if v.kind == "number" && other.kind == "bigint" {
		cmp, ok := compareBigIntFloat(other.big, v.num)
		return ok && cmp >= 0
	}
	// A BigInt/String pair also compares exactly: JS converts the String
	// operand via StringToBigInt for an exact BigInt comparison rather than
	// parsing it as a Number.
	if v.kind == "bigint" && other.kind == "string" {
		n, ok := ecmascript.StringToBigInt(other.str)
		return ok && v.big.Cmp(n) <= 0
	}
	if v.kind == "string" && other.kind == "bigint" {
		n, ok := ecmascript.StringToBigInt(v.str)
		return ok && n.Cmp(other.big) <= 0
	}
	return v.numeric() <= other.numeric()
}

// compareBigIntFloat compares b against f using exact arithmetic, matching
// JS's BigInt-vs-Number abstract relational comparison (which compares
// mathematical values, not float64-rounded ones): -1/0/1 as b is
// less/equal/greater than f. ok is false when f is NaN, mirroring JS's
// "any comparison against NaN is false".
func compareBigIntFloat(b *big.Int, f float64) (cmp int, ok bool) {
	if math.IsNaN(f) {
		return 0, false
	}
	if math.IsInf(f, 1) {
		return -1, true
	}
	if math.IsInf(f, -1) {
		return 1, true
	}
	return new(big.Rat).SetInt(b).Cmp(new(big.Rat).SetFloat64(f)), true
}

// numeric converts v to a float64 the way JS's ToNumber coerces a bound:
// numbers convert directly and strings read as a JS numeric literal, which is
// a wider set of spellings than Go's own float parsing accepts and a narrower
// one in places. lessOrEqual never calls this for a bigint bound — every
// bigint/other-kind pairing has its own exact comparison above.
func (v literalValue) numeric() float64 {
	switch v.kind {
	case "number":
		return v.num
	case "string":
		if number, ok := ecmascript.StringToNumber(v.str); ok {
			return number
		}
	}
	return math.NaN()
}

// normalizedLiteralValue extracts a comparable bound from a range-test
// operand, mirroring ESLint's getNormalizedLiteral. Every literal is a bound,
// including a boolean, null or regex one, because ESLint compares bounds with
// `<=`, which coerces whatever value the literal carries. ok is false for
// anything that isn't a literal (e.g. a variable or a call).
func normalizedLiteralValue(node *ast.Node) (literalValue, bool) {
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindNumericLiteral:
		return numericLiteralValue(node, false)
	case ast.KindBigIntLiteral:
		return bigintValue(node.AsBigIntLiteral().Text, false)
	case ast.KindStringLiteral:
		return literalValue{kind: "string", str: node.AsStringLiteral().Text}, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return literalValue{kind: "string", str: node.AsNoSubstitutionTemplateLiteral().Text}, true
	case ast.KindTrueKeyword:
		return literalValue{kind: "number", num: 1}, true
	case ast.KindFalseKeyword:
		return literalValue{kind: "number", num: 0}, true
	case ast.KindNullKeyword:
		return literalValue{kind: "number", num: 0}, true
	case ast.KindRegularExpressionLiteral:
		// ESLint's getNormalizedLiteral returns any `Literal` node verbatim,
		// including regex literals — it only checks node.type, not the
		// runtime type of node.value. JS's abstract relational comparison
		// then coerces a RegExp operand via ToPrimitive, which falls back to
		// RegExp.prototype.toString() (there is no valueOf override), including
		// canonical flag order. Treat it as a string bound for the same reason.
		return literalValue{kind: "string", str: utils.RegExpLiteralStringValue(node.Text())}, true
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator != ast.KindMinusToken {
			return literalValue{}, false
		}
		operand := ast.SkipParentheses(unary.Operand)
		switch operand.Kind {
		case ast.KindNumericLiteral:
			return numericLiteralValue(operand, true)
		case ast.KindBigIntLiteral:
			return bigintValue(operand.AsBigIntLiteral().Text, true)
		}
	}
	return literalValue{}, false
}

func numericLiteralValue(node *ast.Node, negate bool) (literalValue, bool) {
	text, ok := utils.GetStaticExpressionValue(node)
	if !ok {
		return literalValue{}, false
	}
	return numericValue(text, negate)
}

func numericValue(text string, negate bool) (literalValue, bool) {
	f, err := strconv.ParseFloat(utils.NormalizeNumericLiteral(text), 64)
	if err != nil {
		return literalValue{}, false
	}
	if negate {
		f = -f
	}
	return literalValue{kind: "number", num: f}, true
}

func bigintValue(text string, negate bool) (literalValue, bool) {
	n, ok := new(big.Int).SetString(utils.NormalizeBigIntLiteral(text), 10)
	if !ok {
		return literalValue{}, false
	}
	if negate {
		n.Neg(n)
	}
	return literalValue{kind: "bigint", big: n}, true
}

// isBetweenTest reports whether leftBin/rightBin form a "between" range test
// like `0 <= x && x < 1`: the shared operand is leftBin.Right/rightBin.Left,
// and the literal bounds (when both present) must be ascending.
func isBetweenTest(leftBin, rightBin *ast.BinaryExpression) bool {
	if !utils.IsSameReference(leftBin.Right, rightBin.Left, false) {
		return false
	}
	return isAscendingRange(leftBin.Left, rightBin.Right)
}

// isOutsideTest reports whether leftBin/rightBin form an "outside" range
// test like `x < 0 || 1 <= x`: the shared operand is leftBin.Left/
// rightBin.Right, and the literal bounds (when both present) must be
// ascending.
func isOutsideTest(leftBin, rightBin *ast.BinaryExpression) bool {
	if !utils.IsSameReference(leftBin.Left, rightBin.Right, false) {
		return false
	}
	return isAscendingRange(leftBin.Right, rightBin.Left)
}

func isAscendingRange(lowNode, highNode *ast.Node) bool {
	low, lowOK := normalizedLiteralValue(lowNode)
	high, highOK := normalizedLiteralValue(highNode)
	if !lowOK && !highOK {
		return false
	}
	if !lowOK || !highOK {
		return true
	}
	return low.lessOrEqual(high)
}

// isParenWrapped reports whether node is immediately preceded by `(` and
// followed by `)`, matching ESLint's token-based astUtils.isParenthesised.
// This intentionally does not rely on tsgo's ParenthesizedExpression node:
// the parens around an `if`/`while` condition are mandatory grammar, not an
// optional wrapper, so they never produce one.
func isParenWrapped(sf *ast.SourceFile, node *ast.Node) bool {
	nodeRange := utils.TrimNodeTextRange(sf, node)
	before, ok := utils.TokenBeforePosition(sf, nodeRange.Pos())
	if !ok || before.Text != "(" {
		return false
	}
	after, ok := utils.TokenAtOrAfter(sf, nodeRange.End())
	if !ok || after.Text != ")" {
		return false
	}
	return true
}

// isRangeTest reports whether node is one side of a range test like
// `(0 <= x && x < 1)` or `(x < 0 || 1 <= x)`, mirroring ESLint's isRangeTest
// (called there as `isRangeTest(node.parent)`). node's logical parent is
// found by walking up through any redundant parens around node itself, since
// tsgo (unlike ESTree) keeps those as explicit ParenthesizedExpression nodes.
func isRangeTest(sf *ast.SourceFile, node *ast.Node) bool {
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	logical := parent.AsBinaryExpression()
	opKind := logical.OperatorToken.Kind
	if opKind != ast.KindAmpersandAmpersandToken && opKind != ast.KindBarBarToken {
		return false
	}

	left := ast.SkipParentheses(logical.Left)
	right := ast.SkipParentheses(logical.Right)
	if left.Kind != ast.KindBinaryExpression || right.Kind != ast.KindBinaryExpression {
		return false
	}
	leftBin := left.AsBinaryExpression()
	rightBin := right.AsBinaryExpression()
	if !isRangeTestOperator(leftBin.OperatorToken.Kind) || !isRangeTestOperator(rightBin.OperatorToken.Kind) {
		return false
	}

	var isValidRange bool
	if opKind == ast.KindAmpersandAmpersandToken {
		isValidRange = isBetweenTest(leftBin, rightBin)
	} else {
		isValidRange = isOutsideTest(leftBin, rightBin)
	}
	if !isValidRange {
		return false
	}

	return isParenWrapped(sf, parent)
}

// buildFlippedText returns node's source text with Left/Right swapped and
// the operator flipped, preserving whitespace/comments around the operator
// and guarding against the replacement fusing with an adjacent token.
// Mirrors ESLint's getFlippedString.
func buildFlippedText(sf *ast.SourceFile, node *ast.Node) string {
	bin := node.AsBinaryExpression()
	text := sf.Text()

	leftRange := utils.TrimNodeTextRange(sf, bin.Left)
	opRange := utils.TrimNodeTextRange(sf, bin.OperatorToken)
	rightRange := utils.TrimNodeTextRange(sf, bin.Right)

	leftText := text[leftRange.Pos():leftRange.End()]
	rightText := text[rightRange.Pos():rightRange.End()]
	textBeforeOperator := text[leftRange.End():opRange.Pos()]
	textAfterOperator := text[opRange.End():rightRange.Pos()]
	flippedOperator := scanner.TokenToString(flipOperatorKind(bin.OperatorToken.Kind))

	flipped := rightText + textBeforeOperator + flippedOperator + textAfterOperator + leftText
	return utils.SafeReplacementText(sf, node, flipped)
}

// YodaRule requires or disallows "Yoda" conditions.
// https://eslint.org/docs/latest/rules/yoda
var YodaRule = rule.Rule{
	Name:   "yoda",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if bin == nil || bin.OperatorToken == nil {
					return
				}
				opKind := bin.OperatorToken.Kind
				if !isComparisonOperator(opKind) {
					return
				}
				if opts.onlyEquality && !isEqualityOperator(opKind) {
					return
				}

				left := ast.SkipParentheses(bin.Left)
				right := ast.SkipParentheses(bin.Right)

				var expectedLiteral, expectedNonLiteral *ast.Node
				expectedSide := "right"
				if opts.always {
					expectedLiteral, expectedNonLiteral, expectedSide = left, right, "left"
				} else {
					expectedLiteral, expectedNonLiteral = right, left
				}

				if !isLiteralLike(expectedNonLiteral) || isLiteralLike(expectedLiteral) {
					return
				}
				if opts.exceptRange && isRangeTest(ctx.SourceFile, node) {
					return
				}

				operatorText := scanner.TokenToString(opKind)
				msg := rule.RuleMessage{
					Id:          "expected",
					Description: fmt.Sprintf("Expected literal to be on the %s side of %s.", expectedSide, operatorText),
					Data: map[string]string{
						"expectedSide": expectedSide,
						"operator":     operatorText,
					},
				}

				ctx.ReportNodeWithDeferredFixes(node, msg, func() []rule.RuleFix {
					return []rule.RuleFix{
						rule.RuleFixReplace(ctx.SourceFile, node, buildFlippedText(ctx.SourceFile, node)),
					}
				})
			},
		}
	},
}
