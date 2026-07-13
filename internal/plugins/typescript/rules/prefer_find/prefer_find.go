package prefer_find

import (
	"math"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferFindMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferFind",
		Description: "Prefer .find(...) instead of .filter(...)[0].",
	}
}

func buildPreferFindSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferFindSuggestion",
		Description: "Use .find(...) instead of .filter(...)[0].",
	}
}

// filterExpressionData records a single `<obj>.filter(...)` call that should be
// rewritten to `<obj>.find(...)`. When the rule fires on a ternary, multiple
// distinct filter expressions can need rewriting in one report.
type filterExpressionData struct {
	// filterNode is the AST node that names the method being invoked —
	// either the `filter` identifier (for `.filter(...)`) or the literal /
	// identifier inside the brackets (for `['filter'](...)`, `` [`filter`](...) ``,
	// or `[fltrVar](...)` when fltrVar resolves to "filter").
	filterNode *ast.Node
	// isBracketSyntax is true when the call site uses computed access
	// (`['filter'](...)`). It controls whether the replacement text is
	// `find` (member access) or `"find"` (string literal).
	isBracketSyntax bool
}

// PreferFindRule reports `arr.filter(...)[0]` / `arr.filter(...).at(0)` patterns
// that should be `arr.find(...)`. Mirrors typescript-eslint's prefer-find rule.
var PreferFindRule = rule.CreateRule(rule.Rule{
	Name:             "prefer-find",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		// parseArrayFilterExpressions returns the list of `<obj>.filter(...)`
		// CallExpressions reachable through sequence / ternary / paren wrappers.
		// A non-empty result means the receiver "definitely produces an array
		// from a .filter call" — either directly or in every branch of a
		// ternary.
		var parseArrayFilterExpressions func(expression *ast.Node) []filterExpressionData
		parseArrayFilterExpressions = func(expression *ast.Node) []filterExpressionData {
			// In tsgo, parentheses are explicit nodes and SequenceExpression /
			// AssignmentExpression collapse into BinaryExpression. There is no
			// ChainExpression wrapper, so the upstream `skipChainExpression`
			// has no analogue here.
			node := ast.SkipParentheses(expression)
			if node == nil {
				return nil
			}

			// SequenceExpression: only the last operand matters. `a, b, c` is
			// `BinaryExpression(Op: ',', Left: (a, b), Right: c)` in tsgo, so
			// the "last expression" is the Right child of the outermost comma
			// BinaryExpression.
			if node.Kind == ast.KindBinaryExpression {
				bin := node.AsBinaryExpression()
				if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindCommaToken {
					return parseArrayFilterExpressions(bin.Right)
				}
			}

			if node.Kind == ast.KindConditionalExpression {
				cond := node.AsConditionalExpression()
				consequent := parseArrayFilterExpressions(cond.WhenTrue)
				if len(consequent) == 0 {
					return nil
				}
				alternate := parseArrayFilterExpressions(cond.WhenFalse)
				if len(alternate) == 0 {
					return nil
				}
				return append(consequent, alternate...)
			}

			// CallExpression: `<obj>.filter(...)` (or bracket form). The CALL
			// itself must not be optional (`.filter?.(...)` is excluded);
			// optional-ness of the inner member access (`arr?.filter(...)`)
			// is allowed — upstream behaves the same.
			if node.Kind == ast.KindCallExpression {
				call := node.AsCallExpression()
				if call.QuestionDotToken != nil {
					return nil
				}
				callee := ast.SkipParentheses(call.Expression)
				if callee == nil {
					return nil
				}

				var filterNode, object *ast.Node
				isBracket := false
				switch callee.Kind {
				case ast.KindPropertyAccessExpression:
					pae := callee.AsPropertyAccessExpression()
					name := pae.Name()
					if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "filter" {
						return nil
					}
					filterNode = name
					object = pae.Expression
				case ast.KindElementAccessExpression:
					eae := callee.AsElementAccessExpression()
					arg := eae.ArgumentExpression
					if arg == nil {
						return nil
					}
					strVal, ok := resolveStaticString(ctx, arg)
					if !ok || strVal != "filter" {
						return nil
					}
					filterNode = arg
					isBracket = true
					object = eae.Expression
				default:
					return nil
				}

				if object == nil {
					return nil
				}
				filteredType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, object)
				if !isArrayish(ctx, filteredType) {
					return nil
				}
				return []filterExpressionData{{filterNode: filterNode, isBracketSyntax: isBracket}}
			}

			return nil
		}

		// getObjectIfArrayAtZeroExpression returns the receiver of a
		// `<obj>.at(<treatedAsZero>)` call, or nil if the call doesn't match.
		// Mirrors the upstream helper of the same name.
		getObjectIfArrayAtZero := func(node *ast.Node) *ast.Node {
			call := node.AsCallExpression()
			if call == nil {
				return nil
			}
			// `<obj>.at(arg)` — must have exactly one argument. Unlike the
			// filter call below, we do NOT exclude `at?.(0)` here: upstream
			// only gates on `!callee.optional` (i.e., the member access
			// `?.at` is excluded — see callee.QuestionDotToken below), and
			// happily reports `<obj>.at?.(0)`. The `?.()` form is
			// semantically equivalent to `.find?.()` for our purposes.
			if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
				return nil
			}
			callee := call.Expression
			if callee == nil {
				return nil
			}
			var object *ast.Node
			switch callee.Kind {
			case ast.KindPropertyAccessExpression:
				pae := callee.AsPropertyAccessExpression()
				if pae.QuestionDotToken != nil {
					return nil
				}
				name := pae.Name()
				if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "at" {
					return nil
				}
				object = pae.Expression
			case ast.KindElementAccessExpression:
				eae := callee.AsElementAccessExpression()
				if eae.QuestionDotToken != nil {
					return nil
				}
				arg := eae.ArgumentExpression
				if arg == nil {
					return nil
				}
				strVal, ok := resolveStaticString(ctx, arg)
				if !ok || strVal != "at" {
					return nil
				}
				object = eae.Expression
			default:
				return nil
			}
			if !isTreatedAsZeroByArrayAt(ctx, call.Arguments.Nodes[0]) {
				return nil
			}
			return object
		}

		// isMemberAccessOfZero returns true for `<obj>[0]`, `<obj>['0']`,
		// ``<obj>[`0`]`` and `<obj>[zeroVar]`-style accesses (non-optional).
		// `<obj>?.[0]` is excluded because upstream specifically guards
		// against rewriting the optional form (the autofix would change
		// semantics).
		isMemberAccessOfZero := func(node *ast.Node) bool {
			eae := node.AsElementAccessExpression()
			if eae == nil {
				return false
			}
			if eae.QuestionDotToken != nil {
				return false
			}
			return isTreatedAsZeroByMemberAccess(ctx, eae.ArgumentExpression)
		}

		buildSuggestion := func(object *ast.Node, flagged *ast.Node, filters []filterExpressionData) []rule.RuleFix {
			fixes := make([]rule.RuleFix, 0, len(filters)+1)
			for _, f := range filters {
				replacement := "find"
				if f.isBracketSyntax {
					replacement = `"find"`
				}
				fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, f.filterNode, replacement))
			}
			// Remove the `.at(0)` / `[0]` tail: from the next `.` or `[` token
			// after the array expression up to the end of the flagged expression.
			start := findNextDotOrBracket(ctx.SourceFile, object.End(), flagged.End())
			if start < 0 {
				return nil
			}
			fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(start, flagged.End())))
			return fixes
		}

		report := func(flagged *ast.Node, object *ast.Node, filters []filterExpressionData) {
			fixes := buildSuggestion(object, flagged, filters)
			suggestion := rule.RuleSuggestion{
				Message:  buildPreferFindSuggestionMessage(),
				FixesArr: fixes,
			}
			ctx.ReportNodeWithSuggestions(flagged, buildPreferFindMessage(), suggestion)
		}

		return rule.RuleListeners{
			// `<obj>.at(<treatedAsZero>)` and `<obj>['at'](<treatedAsZero>)`.
			ast.KindCallExpression: func(node *ast.Node) {
				object := getObjectIfArrayAtZero(node)
				if object == nil {
					return
				}
				filters := parseArrayFilterExpressions(object)
				if len(filters) == 0 {
					return
				}
				report(node, object, filters)
			},
			// `<obj>[0]`, `<obj>['0']`, `<obj>[`0`]`, `<obj>[zeroVar]`. tsgo
			// uses ElementAccessExpression for what ESTree calls
			// `MemberExpression[computed=true]`.
			ast.KindElementAccessExpression: func(node *ast.Node) {
				if !isMemberAccessOfZero(node) {
					return
				}
				eae := node.AsElementAccessExpression()
				object := eae.Expression
				if object == nil {
					return
				}
				filters := parseArrayFilterExpressions(object)
				if len(filters) == 0 {
					return
				}
				report(node, object, filters)
			},
		}
	},
})

// isArrayish mirrors upstream's helper of the same name: returns true when
// every non-null/undefined union constituent is an Array or Tuple type (or an
// intersection thereof), and at least one such constituent exists.
// Uses tsgo's built-in `isArrayOrTupleType` (= `isArrayType || isTupleType`),
// matching upstream's `checker.isArrayType(t) || checker.isTupleType(t)`.
func isArrayish(ctx rule.RuleContext, t *checker.Type) bool {
	if t == nil || ctx.TypeChecker == nil {
		return false
	}
	hasArrayPart := false
	for _, unionPart := range utils.UnionTypeParts(t) {
		if utils.IsTypeFlagSet(unionPart, checker.TypeFlagsNull|checker.TypeFlagsUndefined) {
			continue
		}
		intersectionAllArray := true
		for _, ip := range utils.IntersectionTypeParts(unionPart) {
			if !checker.Checker_isArrayOrTupleType(ctx.TypeChecker, ip) {
				intersectionAllArray = false
				break
			}
		}
		if !intersectionAllArray {
			return false
		}
		hasArrayPart = true
	}
	return hasArrayPart
}

// isSymbolLikeType returns true when the type of node is `symbol`,
// `unique symbol`, or a union/intersection that ultimately resolves to one.
// Used as a fast type-info-driven shortcut for upstream's
// `typeof value === 'symbol'` early-return inside isTreatedAsZeroByArrayAt /
// isTreatedAsZeroByMemberAccess — when the type checker says "this is a
// symbol", we can't statically coerce it to a number or to "0", regardless
// of the syntactic shape of the argument (`Symbol('0')`, `Symbol.for('0')`,
// `s` where `s: symbol`, etc.).
func isSymbolLikeType(ctx rule.RuleContext, node *ast.Node) bool {
	if ctx.TypeChecker == nil || node == nil {
		return false
	}
	t := ctx.TypeChecker.GetTypeAtLocation(node)
	return t != nil && utils.IsTypeFlagSet(t, checker.TypeFlagsESSymbolLike)
}

// isTreatedAsZeroByMemberAccess returns true if `String(value) === '0'` for the
// (statically resolved) value of node. Conservatively returns false when the
// value cannot be statically determined — matching upstream, which returns
// undefined from `getStaticValue` and short-circuits to "not zero".
// Symbol-typed arguments short-circuit to false via the type checker, mirroring
// upstream's `typeof value === 'symbol'` guard (which exists so the Number()
// coercion below doesn't throw at runtime).
func isTreatedAsZeroByMemberAccess(ctx rule.RuleContext, node *ast.Node) bool {
	if isSymbolLikeType(ctx, node) {
		return false
	}
	str, ok := resolveStaticString(ctx, node)
	if !ok {
		return false
	}
	return str == "0"
}

// isTreatedAsZeroByArrayAt mirrors upstream's helper: `isNaN(Number(value))` or
// `Math.trunc(Number(value)) === 0`. Symbol-valued arguments are not zero
// (Number(Symbol) throws in JS); we short-circuit via the type checker before
// attempting any static resolution.
func isTreatedAsZeroByArrayAt(ctx rule.RuleContext, node *ast.Node) bool {
	if isSymbolLikeType(ctx, node) {
		return false
	}
	// Try numeric resolution first — handles literal numbers, BigInts, unary
	// minus, the `NaN` / `Infinity` globals, and identifiers that const-resolve
	// to one of those.
	if v, ok := resolveStaticNumber(ctx, node); ok {
		if math.IsNaN(v) {
			return true
		}
		return math.Trunc(v) == 0
	}
	// Fall back to string resolution → JS Number(string) coercion. Covers cases
	// like `arr.filter(...).at('0')` (already tested upstream via the `[0]`
	// form, but symmetrically handled here for robustness).
	if s, ok := resolveStaticString(ctx, node); ok {
		v := jsNumberFromString(s)
		if math.IsNaN(v) {
			return true
		}
		return math.Trunc(v) == 0
	}
	return false
}

// resolveStaticString returns the JS `String(value)` representation of node
// when it can be statically determined. Recurses through const identifier
// initializers using the type checker. Numeric/BigInt literals are normalized
// via the standard helpers so `0x0`, `0o0`, `0.0`, `0n`, `-0n` all canonicalize
// to "0".
func resolveStaticString(ctx rule.RuleContext, node *ast.Node) (string, bool) {
	return resolveStaticStringRec(ctx, node, map[*ast.Symbol]bool{})
}

func resolveStaticStringRec(ctx rule.RuleContext, node *ast.Node, seen map[*ast.Symbol]bool) (string, bool) {
	if node == nil {
		return "", false
	}
	// Transparently unwrap parens, `as T` / `<T>x` / `x satisfies T`, and
	// `x!`. Matches ESLint's `getStaticValue` which transparently descends
	// through all of these wrappers when evaluating a value position.
	node = ast.SkipOuterExpressions(node, ast.OEKAll)
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNumericLiteral,
		ast.KindRegularExpressionLiteral:
		// `utils.GetStaticExpressionValue` normalizes numeric literals
		// (`0x0`/`0o0`/`0.0` → `"0"`) so the `== "0"` test below works for
		// every numeric form ESTree's `String(value)` would.
		return utils.GetStaticExpressionValue(node)
	case ast.KindBigIntLiteral:
		return utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text), true
	case ast.KindNullKeyword:
		return "null", true
	case ast.KindTrueKeyword:
		return "true", true
	case ast.KindFalseKeyword:
		return "false", true
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindMinusToken {
			if v, ok := resolveStaticNumberRec(ctx, node, seen); ok {
				return jsStringFromNumber(v), true
			}
			// `-0n` / `-1n` etc. → string-format the negated bigint.
			operand := ast.SkipParentheses(unary.Operand)
			if operand != nil && operand.Kind == ast.KindBigIntLiteral {
				abs := utils.NormalizeBigIntLiteral(operand.AsBigIntLiteral().Text)
				if abs == "0" {
					return "0", true
				}
				return "-" + abs, true
			}
		}
		if unary.Operator == ast.KindPlusToken {
			if v, ok := resolveStaticNumberRec(ctx, node, seen); ok {
				return jsStringFromNumber(v), true
			}
		}
	case ast.KindIdentifier:
		name := node.AsIdentifier().Text
		if name == "NaN" && !utils.IsShadowed(node, "NaN") {
			if declared, ok := ctx.Globals["NaN"]; ok && !declared {
				return "", false
			}
			return "NaN", true
		}
		if name == "Infinity" && !utils.IsShadowed(node, "Infinity") {
			if declared, ok := ctx.Globals["Infinity"]; ok && !declared {
				return "", false
			}
			return "Infinity", true
		}
		if name == "undefined" && !utils.IsShadowed(node, "undefined") {
			if declared, ok := ctx.Globals["undefined"]; ok && !declared {
				return "", false
			}
			return "undefined", true
		}
		if init := resolveConstInitializer(ctx, node, seen); init != nil {
			return resolveStaticStringRec(ctx, init, seen)
		}
	}
	return "", false
}

// resolveStaticNumber returns the JS `Number(value)` coercion when the value
// can be statically determined. NaN/Infinity are returned as Go's
// math.NaN()/math.Inf(...). Returns ok=false for non-numeric resolvable
// values (strings) — the caller should fall back to resolveStaticString.
func resolveStaticNumber(ctx rule.RuleContext, node *ast.Node) (float64, bool) {
	return resolveStaticNumberRec(ctx, node, map[*ast.Symbol]bool{})
}

func resolveStaticNumberRec(ctx rule.RuleContext, node *ast.Node, seen map[*ast.Symbol]bool) (float64, bool) {
	if node == nil {
		return 0, false
	}
	node = ast.SkipOuterExpressions(node, ast.OEKAll)
	switch node.Kind {
	case ast.KindNumericLiteral:
		return evalNumericLiteralText(node.AsNumericLiteral().Text)
	case ast.KindNullKeyword:
		// Number(null) === 0
		return 0, true
	case ast.KindTrueKeyword:
		return 1, true
	case ast.KindFalseKeyword:
		return 0, true
	case ast.KindBigIntLiteral:
		// Number(bigint) — lossy for large values; tests only use 0n / -0n.
		norm := utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text)
		v, err := strconv.ParseFloat(norm, 64)
		if err != nil {
			return 0, false
		}
		return v, true
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator != ast.KindMinusToken && unary.Operator != ast.KindPlusToken {
			return 0, false
		}
		v, ok := resolveStaticNumberRec(ctx, unary.Operand, seen)
		if !ok {
			return 0, false
		}
		if unary.Operator == ast.KindMinusToken {
			return -v, true
		}
		return v, true
	case ast.KindIdentifier:
		name := node.AsIdentifier().Text
		if name == "NaN" && !utils.IsShadowed(node, "NaN") {
			if declared, ok := ctx.Globals["NaN"]; ok && !declared {
				return 0, false
			}
			return math.NaN(), true
		}
		if name == "Infinity" && !utils.IsShadowed(node, "Infinity") {
			if declared, ok := ctx.Globals["Infinity"]; ok && !declared {
				return 0, false
			}
			return math.Inf(1), true
		}
		if init := resolveConstInitializer(ctx, node, seen); init != nil {
			return resolveStaticNumberRec(ctx, init, seen)
		}
	}
	return 0, false
}

// resolveConstInitializer returns the initializer expression for an identifier
// that resolves (via the type checker) to a `const` variable declaration.
// Returns nil for non-const bindings, declarations without initializers, and
// cycles (e.g. `const a = a`).
func resolveConstInitializer(ctx rule.RuleContext, idNode *ast.Node, seen map[*ast.Symbol]bool) *ast.Node {
	if ctx.TypeChecker == nil {
		return nil
	}
	sym := ctx.TypeChecker.GetSymbolAtLocation(idNode)
	if sym == nil || seen[sym] {
		return nil
	}
	seen[sym] = true
	if sym.Declarations == nil {
		return nil
	}
	for _, decl := range sym.Declarations {
		if decl == nil || !ast.IsVariableDeclaration(decl) {
			continue
		}
		// Only `const` declarations can be statically inlined — `let` and `var`
		// can be reassigned, so the initializer is not authoritative.
		parent := decl.Parent
		if parent == nil || !ast.IsVariableDeclarationList(parent) || parent.Flags&ast.NodeFlagsConst == 0 {
			continue
		}
		init := decl.AsVariableDeclaration().Initializer
		if init != nil {
			return init
		}
	}
	return nil
}

// evalNumericLiteralText returns the float64 value of a NumericLiteral.Text
// (which is always non-negative — unary minus is a separate AST node).
// Handles JS hex / octal / binary / decimal forms via the existing helper.
func evalNumericLiteralText(text string) (float64, bool) {
	norm := utils.NormalizeNumericLiteral(text)
	switch norm {
	case "Infinity":
		return math.Inf(1), true
	case "-Infinity":
		return math.Inf(-1), true
	}
	v, err := strconv.ParseFloat(norm, 64)
	if err != nil {
		return 0, false
	}
	return v, true
}

// jsStringFromNumber formats a float64 the way JavaScript's `String(n)` does.
// Crucially, `String(-0)` is `"0"`, not `"-0"` — `arr[-0]` is the same index
// as `arr[0]` and the rule must match it.
func jsStringFromNumber(v float64) string {
	if math.IsNaN(v) {
		return "NaN"
	}
	if math.IsInf(v, 1) {
		return "Infinity"
	}
	if math.IsInf(v, -1) {
		return "-Infinity"
	}
	if v == 0 {
		return "0"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// jsNumberFromString approximates JavaScript's `Number(str)` coercion: trim
// whitespace, empty → 0, otherwise parse as decimal/scientific/hex/octal/
// binary, NaN on failure. Two JS-vs-Go quirks worth calling out:
//
//   - Go's `strconv.ParseFloat` accepts decimal and scientific forms, plus
//     hex floats with a `p` exponent, but rejects the bare `0x` / `0o` / `0b`
//     integer prefixes that JS `Number()` accepts (`Number('0x1') === 1`).
//     Fall back to `strconv.ParseInt` with base 0 for those.
//   - Go accepts `_` digit separators (`Number('1_000')` in Go = 1000), but
//     JS `Number()` rejects them (`Number('1_000') === NaN`). Reject any
//     string containing `_` up front to match JS.
func jsNumberFromString(s string) float64 {
	trimmed := strings.TrimSpace(s)
	if trimmed == "" {
		return 0
	}
	if strings.ContainsRune(trimmed, '_') {
		return math.NaN()
	}
	if v, err := strconv.ParseFloat(trimmed, 64); err == nil {
		return v
	}
	if len(trimmed) > 2 && trimmed[0] == '0' {
		switch trimmed[1] {
		case 'x', 'X', 'o', 'O', 'b', 'B':
			if v, err := strconv.ParseInt(trimmed, 0, 64); err == nil {
				return float64(v)
			}
		}
	}
	return math.NaN()
}

// findNextDotOrBracket scans forward from `start` until it finds a `.` or `[`
// token (or reaches `end`). Returns the token's start position, or -1 if not
// found. Used to locate the start of the `.at(0)` / `[0]` tail to delete.
func findNextDotOrBracket(sourceFile *ast.SourceFile, start, end int) int {
	s := scanner.GetScannerForSourceFile(sourceFile, start)
	for s.TokenStart() < end {
		tok := s.Token()
		if tok == ast.KindDotToken || tok == ast.KindOpenBracketToken {
			return s.TokenStart()
		}
		s.Scan()
	}
	return -1
}
