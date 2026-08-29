package operator_assignment

import (
	_ "embed"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed operator_assignment.schema.json
var schemaJSON []byte

// shorthandOperator describes one of the 12 operators with an assignment
// shorthand form: its plain binary-operator token (`+`), the corresponding
// compound-assignment token (`+=`), and whether it is commutative. Logical
// assignment (`&&=`, `||=`, `??=`) and plain `=` are intentionally absent.
type shorthandOperator struct {
	plain                 ast.Kind
	compound              ast.Kind
	commutative           bool
	plainText             string
	compoundText          string
	replacedDescription   string
	unexpectedDescription string
}

var shorthandOperators = []shorthandOperator{
	{plain: ast.KindPlusToken, compound: ast.KindPlusEqualsToken},
	{plain: ast.KindMinusToken, compound: ast.KindMinusEqualsToken},
	{plain: ast.KindAsteriskToken, compound: ast.KindAsteriskEqualsToken, commutative: true},
	{plain: ast.KindSlashToken, compound: ast.KindSlashEqualsToken},
	{plain: ast.KindPercentToken, compound: ast.KindPercentEqualsToken},
	{plain: ast.KindAsteriskAsteriskToken, compound: ast.KindAsteriskAsteriskEqualsToken},
	{plain: ast.KindLessThanLessThanToken, compound: ast.KindLessThanLessThanEqualsToken},
	{plain: ast.KindGreaterThanGreaterThanToken, compound: ast.KindGreaterThanGreaterThanEqualsToken},
	{plain: ast.KindGreaterThanGreaterThanGreaterThanToken, compound: ast.KindGreaterThanGreaterThanGreaterThanEqualsToken},
	{plain: ast.KindAmpersandToken, compound: ast.KindAmpersandEqualsToken, commutative: true},
	{plain: ast.KindCaretToken, compound: ast.KindCaretEqualsToken, commutative: true},
	{plain: ast.KindBarToken, compound: ast.KindBarEqualsToken, commutative: true},
}

var (
	shorthandOperatorsByPlain    [ast.KindCaretToken - ast.KindPlusToken + 1]*shorthandOperator
	shorthandOperatorsByCompound [ast.KindCaretEqualsToken - ast.KindPlusEqualsToken + 1]*shorthandOperator
)

func init() {
	for i := range shorthandOperators {
		op := &shorthandOperators[i]
		shorthandOperatorsByPlain[op.plain-ast.KindPlusToken] = op
		shorthandOperatorsByCompound[op.compound-ast.KindPlusEqualsToken] = op
		op.plainText = scanner.TokenToString(op.plain)
		op.compoundText = scanner.TokenToString(op.compound)
		op.replacedDescription = "Assignment (=) can be replaced with operator assignment (" + op.compoundText + ")."
		op.unexpectedDescription = "Unexpected operator assignment (" + op.compoundText + ") shorthand."
	}
}

func (op *shorthandOperator) replacedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "replaced",
		Description: op.replacedDescription,
		Data:        map[string]string{"operator": op.compoundText},
	}
}

func (op *shorthandOperator) unexpectedMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: op.unexpectedDescription,
		Data:        map[string]string{"operator": op.compoundText},
	}
}

func shorthandOperatorByPlain(kind ast.Kind) *shorthandOperator {
	if kind < ast.KindPlusToken || kind > ast.KindCaretToken {
		return nil
	}
	return shorthandOperatorsByPlain[kind-ast.KindPlusToken]
}

func shorthandOperatorByCompound(kind ast.Kind) *shorthandOperator {
	if kind < ast.KindPlusEqualsToken || kind > ast.KindCaretEqualsToken {
		return nil
	}
	return shorthandOperatorsByCompound[kind-ast.KindPlusEqualsToken]
}

// isLiteralPropertyKey reports whether node is a literal usable as a computed
// property key without side effects (matches ESTree's unified "Literal" type:
// string / numeric / bigint / regex literals, plus the `true` / `false` /
// `null` keyword literals). A no-substitution template literal (a backtick
// string with no interpolation) is deliberately excluded: ESTree gives it its
// own "TemplateLiteral" type, not "Literal", so upstream's
// `node.property.type === "Literal"` check — and its isSameReference call,
// which has no case for "TemplateLiteral" either — reject it too. Verified
// against ESLint 10.8.1: a template-literal computed key reports nothing,
// while the equivalent string-literal key does.
func isLiteralPropertyKey(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	}
	return false
}

// canBeFixed reports whether node can be safely rewritten between `x = x op y`
// and `x op= y` without changing how many times a getter/setter or a computed
// key's `toString()` runs. Parentheses and TS-only wrappers (`as`,
// `satisfies`, `!`) are transparent here — they have no runtime effect, and
// whether the fix may drop one is decided separately, by the assertion-syntax
// comparison in reportReplaced. Any node that is part of an optional chain is
// rejected outright: ESLint's own canBeFixed checks the raw ESTree node type, and an
// optional chain is always wrapped in a ChainExpression there, which matches
// neither of its two accepted shapes.
func canBeFixed(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return true
	case ast.KindPropertyAccessExpression:
		if ast.IsOptionalChain(node) {
			return false
		}
		object := utils.SkipAssertionsAndParens(node.AsPropertyAccessExpression().Expression)
		return object != nil && (object.Kind == ast.KindIdentifier || object.Kind == ast.KindThisKeyword)
	case ast.KindElementAccessExpression:
		if ast.IsOptionalChain(node) {
			return false
		}
		elementAccess := node.AsElementAccessExpression()
		object := utils.SkipAssertionsAndParens(elementAccess.Expression)
		if object == nil || (object.Kind != ast.KindIdentifier && object.Kind != ast.KindThisKeyword) {
			return false
		}
		argument := utils.SkipAssertionsAndParens(elementAccess.ArgumentExpression)
		return argument != nil && isLiteralPropertyKey(argument)
	}
	return false
}

// hasSameTypeSyntax takes an allocation-free fast path for the common case
// where both assertions repeat the same source text. The token fallback keeps
// equivalent trivia and outer type parentheses transparent.
func hasSameTypeSyntax(sourceFile *ast.SourceFile, left, right *ast.Node) bool {
	left = ast.SkipParentheses(left)
	right = ast.SkipParentheses(right)
	if left == nil || right == nil || left.Kind != right.Kind {
		return false
	}
	if scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, left, false) ==
		scanner.GetSourceTextOfNodeFromSourceFile(sourceFile, right, false) {
		return true
	}
	// Identifier escapes do not change a TypeScript type reference's meaning.
	// Keep that semantic equivalence here even though the TS parser exposes raw
	// identifier spellings to token-oriented ESLint rules.
	return utils.HasSameTokensWithDecodedIdentifiers(sourceFile, left, right)
}

// hasSameAssertionStructure reports whether two already-matched references
// carry the same TypeScript-only wrappers. Parentheses are transparent, but
// `as`, `satisfies`, angle-bracket assertions, and non-null assertions must
// match at every reference component that the fixer may delete.
//
// Type nodes are compared by tokens rather than generic AST child traversal.
// Some type syntax stores meaning in scalar fields that ForEachChild omits —
// notably TypeOperatorNode.Operator (`keyof` versus `readonly`) and
// ImportTypeNode.IsTypeOf. Token comparison keeps those distinctions while
// ignoring comments and whitespace.
//
// The caller must first establish utils.IsSameReference(left, right, true), so
// base identifier/property/literal equality does not need to be repeated here.
func hasSameAssertionStructure(sourceFile *ast.SourceFile, left, right *ast.Node) bool {
	left = ast.SkipParentheses(left)
	right = ast.SkipParentheses(right)
	if left == nil || right == nil || left.Kind != right.Kind {
		return false
	}
	if ast.IsOptionalChain(left) != ast.IsOptionalChain(right) {
		return false
	}

	switch left.Kind {
	case ast.KindAsExpression:
		leftAs := left.AsAsExpression()
		rightAs := right.AsAsExpression()
		return hasSameTypeSyntax(sourceFile, leftAs.Type, rightAs.Type) &&
			hasSameAssertionStructure(sourceFile, leftAs.Expression, rightAs.Expression)
	case ast.KindSatisfiesExpression:
		leftSatisfies := left.AsSatisfiesExpression()
		rightSatisfies := right.AsSatisfiesExpression()
		return hasSameTypeSyntax(sourceFile, leftSatisfies.Type, rightSatisfies.Type) &&
			hasSameAssertionStructure(sourceFile, leftSatisfies.Expression, rightSatisfies.Expression)
	case ast.KindTypeAssertionExpression:
		leftAssertion := left.AsTypeAssertion()
		rightAssertion := right.AsTypeAssertion()
		return hasSameTypeSyntax(sourceFile, leftAssertion.Type, rightAssertion.Type) &&
			hasSameAssertionStructure(sourceFile, leftAssertion.Expression, rightAssertion.Expression)
	case ast.KindNonNullExpression:
		return hasSameAssertionStructure(
			sourceFile,
			left.AsNonNullExpression().Expression,
			right.AsNonNullExpression().Expression,
		)
	case ast.KindPropertyAccessExpression:
		return hasSameAssertionStructure(
			sourceFile,
			left.AsPropertyAccessExpression().Expression,
			right.AsPropertyAccessExpression().Expression,
		)
	case ast.KindElementAccessExpression:
		leftAccess := left.AsElementAccessExpression()
		rightAccess := right.AsElementAccessExpression()
		return hasSameAssertionStructure(sourceFile, leftAccess.Expression, rightAccess.Expression) &&
			hasSameAssertionStructure(sourceFile, leftAccess.ArgumentExpression, rightAccess.ArgumentExpression)
	default:
		return true
	}
}

// checkAlways implements the default "always" mode: `x = x op y` should be
// written as `x op= y` where possible.
func checkAlways(ctx *rule.RuleContext, node *ast.Node) {
	binExpr := node.AsBinaryExpression()
	if binExpr.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}

	rhs := ast.SkipParentheses(binExpr.Right)
	if rhs == nil || rhs.Kind != ast.KindBinaryExpression {
		return
	}
	rhsExpr := rhs.AsBinaryExpression()
	operatorKind := rhsExpr.OperatorToken.Kind
	op := shorthandOperatorByPlain(operatorKind)
	if op == nil {
		return
	}
	commutative := op.commutative

	left := binExpr.Left

	if utils.IsSameReference(left, rhsExpr.Left, true) {
		reportReplaced(ctx, node, binExpr, left, rhs, rhsExpr, op)
		return
	}
	if commutative && utils.IsSameReference(left, rhsExpr.Right, true) {
		// This case can't be fixed safely: if `a` and `b` both have custom
		// valueOf() behavior, fixing `a = b * a` to `a *= b` would change the
		// order the valueOf() functions run in.
		ctx.ReportNode(node, op.replacedMessage())
	}
}

func reportReplaced(
	ctx *rule.RuleContext,
	node *ast.Node,
	binExpr *ast.BinaryExpression,
	left *ast.Node,
	rhs *ast.Node,
	rhsExpr *ast.BinaryExpression,
	op *shorthandOperator,
) {
	ctx.ReportNodeWithDeferredFixes(node, op.replacedMessage(), func() []rule.RuleFix {
		sourceFile := ctx.SourceFile
		// The fix deletes the right-hand occurrence of the reference and keeps
		// the assignment target verbatim, so the two must be interchangeable at
		// the type level, not just at runtime. Keep this edit-only work deferred:
		// diagnostics-only consumers do not need to know whether a fix is safe.
		if !canBeFixed(left) || !hasSameAssertionStructure(sourceFile, left, rhsExpr.Left) {
			return nil
		}

		eqRange := utils.TrimNodeTextRange(sourceFile, binExpr.OperatorToken)
		opRange := utils.TrimNodeTextRange(sourceFile, rhsExpr.OperatorToken)

		if utils.HasCommentInSpan(ctx.Comments.All(), eqRange.End(), opRange.Pos()) {
			return nil
		}

		text := sourceFile.Text()
		nodeRange := utils.TrimNodeTextRange(sourceFile, node)
		leftText := text[nodeRange.Pos():eqRange.Pos()]
		rightText := text[opRange.End():rhs.End()]
		replacement := leftText + op.compoundText + rightText

		return []rule.RuleFix{rule.RuleFixReplace(sourceFile, node, replacement)}
	})
}

// rightNeedsParens reports whether the right side of a compound assignment has
// to be parenthesized once the expanded `x op ` prefix is written in front of
// it, given the ESLint precedence of the plain operator being written.
func rightNeedsParens(right *ast.Node, newOperatorPrecedence int) bool {
	// A lower- (or equal-) precedence right side needs parentheses to preserve
	// grouping (e.g. `foo *= bar + 1` -> `foo * (bar + 1)`). TS-only right
	// sides (`y!`, `y as T`, `y satisfies T`, `<T>y`) are unknown to ESLint's
	// precedence table and land here too, so they get parenthesized just as
	// upstream does with @typescript-eslint/parser.
	// An already-parenthesized right side reports the maximum precedence here,
	// so this branch is never hit for it — the existing parentheses are
	// instead preserved verbatim by the caller's plain-slice branch.
	if utils.EslintLikePrecedence(right) <= newOperatorPrecedence {
		return true
	}
	return startsWithBareTypeOperator(right)
}

// startsWithBareTypeOperator reports whether the leftmost operand of expr is an
// unparenthesized `as` / `satisfies` expression.
//
// TypeScript parses `a as number * b` as `(a as number) * b`, since the type
// after `as` stops at `number` — a shape plain JavaScript can never produce,
// because a lower-precedence operand cannot sit unparenthesized under a
// higher-precedence one. Looking at the root operator alone (`*`, which binds
// tighter than `+`) would therefore call the right side paren-free, and writing
// `x = x + a as number * b` re-parses as `((x + a) as number) * b` — different
// grouping, different result. Parenthesizing the whole right side keeps it.
func startsWithBareTypeOperator(expr *ast.Node) bool {
	for expr != nil {
		switch expr.Kind {
		case ast.KindAsExpression, ast.KindSatisfiesExpression:
			return true
		case ast.KindBinaryExpression:
			// Parentheses interrupt the walk: a ParenthesizedExpression is not
			// a BinaryExpression, so `(a as number) * b` stops here and is
			// correctly reported as needing no extra parentheses.
			expr = expr.AsBinaryExpression().Left
		default:
			return false
		}
	}
	return false
}

// spellsTypeAngleBrackets reports whether expr writes a TypeScript type
// argument list (`foo<T>`, `f<T>()`, `x as Array<T>`) or type parameter list
// (`<X>(v: X) => X`) anywhere inside it. Angle brackets that belong to a value
// expression — `a < b`, a JSX element, a `<` inside a string or comment — are
// not type syntax and are not reported here.
func spellsTypeAngleBrackets(expr *ast.Node) bool {
	if expr == nil {
		return false
	}
	if hasTypeArgumentList(expr) {
		return true
	}
	if funcLike := expr.FunctionLikeData(); funcLike != nil && funcLike.TypeParameters != nil {
		return true
	}
	return expr.ForEachChild(spellsTypeAngleBrackets)
}

// hasTypeArgumentList reports whether node is one of the shapes that can carry
// an explicit `<...>` type argument list, and does carry one.
func hasTypeArgumentList(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression,
		ast.KindExpressionWithTypeArguments, ast.KindTypeReference, ast.KindTypeQuery,
		ast.KindImportType, ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		return node.TypeArgumentList() != nil
	}
	return false
}

// checkNever implements the "never" mode: any of the 12 shorthand compound
// assignment operators should be written as `x = x op y` instead.
func checkNever(ctx *rule.RuleContext, node *ast.Node) {
	binExpr := node.AsBinaryExpression()
	op := shorthandOperatorByCompound(binExpr.OperatorToken.Kind)
	if op == nil {
		return
	}
	plainOperatorKind := op.plain

	ctx.ReportNodeWithDeferredFixes(node, op.unexpectedMessage(), func() []rule.RuleFix {
		left := binExpr.Left
		if !canBeFixed(left) {
			return nil
		}

		sourceFile := ctx.SourceFile
		nodeRange := utils.TrimNodeTextRange(sourceFile, node)
		opRange := utils.TrimNodeTextRange(sourceFile, binExpr.OperatorToken)

		if utils.HasCommentInSpan(ctx.Comments.All(), nodeRange.Pos(), opRange.Pos()) {
			return nil
		}

		text := sourceFile.Text()
		leftText := text[nodeRange.Pos():opRange.Pos()]
		plainOperatorText := op.plainText

		newOperatorPrecedence := utils.EslintLikeBinaryOperatorPrecedence(plainOperatorKind)

		rightRange := utils.TrimNodeTextRange(sourceFile, binExpr.Right)
		inner := text[rightRange.Pos():rightRange.End()]
		needsParens := rightNeedsParens(binExpr.Right, newOperatorPrecedence)

		// TypeScript re-scans `x << (` as `x <` opening a type argument list,
		// and a right side that spells its own `<...>` supplies the `>` that
		// makes the parser commit to that reading: `x = x << (foo<T>)` is
		// TS1005 instead of a shift. The `(` is the fixer's own when the right
		// side needs parentheses, and the source's own when the right side is
		// already parenthesized — both spellings mis-scan, and no extra layer
		// of parentheses reliably escapes it (`x = x << ((a ? foo<T> : b))`
		// still fails), so this is reported without a fix. Angle brackets that
		// are not type syntax — comparisons, strings, templates, comments —
		// leave no `>` for the mis-scan to close on and are fixed as usual.
		if plainOperatorKind == ast.KindLessThanLessThanToken &&
			(needsParens || strings.HasPrefix(inner, "(")) &&
			spellsTypeAngleBrackets(binExpr.Right) {
			return nil
		}

		var rightText string
		if needsParens {
			between := text[opRange.End():rightRange.Pos()]
			rightText = between + "(" + inner + ")"
		} else {
			rest := text[opRange.End():node.End()]
			prefix := ""
			if firstRune, size := utf8.DecodeRuneInString(rest); firstRune != utf8.RuneError && size > 0 {
				if !utils.CanTokenTextsBeAdjacent(plainOperatorText, rest[:size]) {
					prefix = " "
				}
			}
			rightText = prefix + rest
		}

		replacement := leftText + "= " + leftText + plainOperatorText + rightText
		return []rule.RuleFix{rule.RuleFixReplace(sourceFile, node, replacement)}
	})
}

// https://eslint.org/docs/latest/rules/operator-assignment
var OperatorAssignmentRule = rule.Rule{
	Name:   "operator-assignment",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		never := false
		if len(options) > 0 {
			if s, ok := options[0].(string); ok && s == "never" {
				never = true
			}
		}

		if never {
			return rule.RuleListeners{
				ast.KindBinaryExpression: func(node *ast.Node) {
					checkNever(&ctx, node)
				},
			}
		}
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				checkAlways(&ctx, node)
			},
		}
	},
}
