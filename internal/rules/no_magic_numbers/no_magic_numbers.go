package no_magic_numbers

import (
	_ "embed"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_magic_numbers.schema.json
var schemaJSON []byte

// Maximum array length by the ECMAScript Specification.
const maxArrayLength = 1<<32 - 1

type options struct {
	detectObjects                 bool
	enforceConst                  bool
	ignore                        map[string]bool // normalized values as strings
	ignoreArrayIndexes            bool
	ignoreDefaultValues           bool
	ignoreClassFieldInitialValues bool
	ignoreEnums                   bool
	ignoreNumericLiteralTypes     bool
	ignoreReadonlyClassProperties bool
	ignoreTypeIndexes             bool
}

func parseOptions(rawOpts []any) options {
	opts := options{}
	if len(rawOpts) == 0 {
		return opts
	}
	optsMap, _ := rawOpts[0].(map[string]interface{})
	if v, ok := optsMap["detectObjects"].(bool); ok {
		opts.detectObjects = v
	}
	if v, ok := optsMap["enforceConst"].(bool); ok {
		opts.enforceConst = v
	}
	if v, ok := optsMap["ignoreArrayIndexes"].(bool); ok {
		opts.ignoreArrayIndexes = v
	}
	if v, ok := optsMap["ignoreDefaultValues"].(bool); ok {
		opts.ignoreDefaultValues = v
	}
	if v, ok := optsMap["ignoreClassFieldInitialValues"].(bool); ok {
		opts.ignoreClassFieldInitialValues = v
	}
	if v, ok := optsMap["ignoreEnums"].(bool); ok {
		opts.ignoreEnums = v
	}
	if v, ok := optsMap["ignoreNumericLiteralTypes"].(bool); ok {
		opts.ignoreNumericLiteralTypes = v
	}
	if v, ok := optsMap["ignoreReadonlyClassProperties"].(bool); ok {
		opts.ignoreReadonlyClassProperties = v
	}
	if v, ok := optsMap["ignoreTypeIndexes"].(bool); ok {
		opts.ignoreTypeIndexes = v
	}
	if arr, ok := optsMap["ignore"].([]interface{}); ok {
		opts.ignore = make(map[string]bool, len(arr))
		for _, item := range arr {
			switch v := item.(type) {
			case float64:
				opts.ignore[normalizeFloatKey(v)] = true
			case string:
				// BigInt string like "100n" or "-100n"
				opts.ignore[normalizeIgnoreBigIntString(v)] = true
			}
		}
	}
	return opts
}

// normalizeFloatKey converts a float64 to a canonical lookup key. Upstream
// looks values up in a JavaScript Set, whose SameValueZero comparison treats
// -0 and 0 as the same value, so signed zero collapses to "0".
func normalizeFloatKey(v float64) string {
	if v == 0 {
		return "0"
	}
	return strconv.FormatFloat(v, 'f', -1, 64)
}

// normalizeIgnoreBigIntString converts a bigint ignore string like "100n" or "-100n"
// to a canonical key: "bigint:100" or "bigint:-100".
func normalizeIgnoreBigIntString(s string) string {
	negative := false
	text := s
	if strings.HasPrefix(text, "-") {
		negative = true
		text = text[1:]
	} else if strings.HasPrefix(text, "+") {
		text = text[1:]
	}
	text = strings.TrimSuffix(text, "n")
	n, ok := new(big.Int).SetString(text, 10)
	if !ok {
		return "bigint:" + s
	}
	if negative {
		n.Neg(n)
	}
	return "bigint:" + n.String()
}

// parseRawNumericValue parses a raw numeric literal string (which may be hex, octal, binary)
// into a float64. Unlike strconv.ParseFloat, this handles all JS numeric prefixes.
func parseRawNumericValue(raw string) (float64, bool) {
	if len(raw) > 2 && raw[0] == '0' {
		switch raw[1] {
		case 'x', 'X', 'o', 'O', 'b', 'B':
			n, ok := new(big.Int).SetString(raw, 0)
			if !ok {
				return 0, false
			}
			f, _ := new(big.Float).SetInt(n).Float64()
			return f, true
		}
	}
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil && !math.IsInf(f, 0) {
		return 0, false
	}
	return f, true
}

// skipParensUp walks from a node's parent upward through ParenthesizedExpression
// nodes and returns the first non-parenthesized ancestor.
// ESTree has no ParenthesizedExpression nodes; tsgo does. This bridges the gap.
func skipParensUp(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

// findUnaryParent checks if the numeric literal's effective parent (after skipping
// parentheses) is a PrefixUnaryExpression with +/-. Returns the PrefixUnaryExpression
// node and operator kind, or nil if not found.
func findUnaryParent(node *ast.Node) (*ast.Node, ast.Kind) {
	p := skipParensUp(node.Parent)
	if p != nil && p.Kind == ast.KindPrefixUnaryExpression {
		pref := p.AsPrefixUnaryExpression()
		if pref.Operator == ast.KindMinusToken || pref.Operator == ast.KindPlusToken {
			return p, pref.Operator
		}
	}
	return nil, 0
}

var noMagicMessage = rule.RuleMessage{
	Id:          "noMagic",
	Description: "No magic number: {{raw}}.",
}

var useConstMessage = rule.RuleMessage{
	Id:          "useConst",
	Description: "Number constants declarations must use 'const'.",
}

func noMagicNumberMessage(raw string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          noMagicMessage.Id,
		Description: "No magic number: " + raw + ".",
	}
}

// NoMagicNumbersRule implements the ESLint core no-magic-numbers rule.
// https://eslint.org/docs/latest/rules/no-magic-numbers
var NoMagicNumbersRule = rule.Rule{
	Name:   "no-magic-numbers",
	Schema: rule.NewSchema(schemaJSON),
	Run:    RunTSESLint,
}

// RunTSESLint implements no-magic-numbers. It is exported so that
// @typescript-eslint/no-magic-numbers (internal/plugins/typescript/rules/no_magic_numbers)
// can share this implementation: as of the ESLint version this rule is
// ported from, the core rule's schema and behavior already fully subsume
// the typescript-eslint extension (ignoreEnums, ignoreNumericLiteralTypes,
// ignoreReadonlyClassProperties, ignoreTypeIndexes), so there is no
// TypeScript-specific variant to select between.
func RunTSESLint(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
	opts := parseOptions(rawOptions)
	hasIgnoredValues := len(opts.ignore) != 0

	handleNumericNode := func(node *ast.Node, isBigInt bool) {
		nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
		ownRaw := ctx.SourceFile.Text()[nodeRange.Pos():nodeRange.End()]

		// Promote to the enclosing unary +/- (through parentheses), if
		// any, so that both the reported text and the ignore-list value
		// account for the sign. Matches upstream: `raw` and `value` are
		// computed once, up front, and reused by every check below.
		// The reported text concatenates the operator with the literal's
		// own raw text, so parentheses and whitespace between the two
		// (`-(1)`, `- 1`) still report as `-1`, while the reported range
		// still spans the whole unary expression.
		fullNumberNode := node
		fullNumberRange := nodeRange
		fullRaw := ownRaw
		isNegative := false
		if unary, op := findUnaryParent(node); unary != nil {
			fullNumberNode = unary
			fullNumberRange = utils.TrimNodeTextRange(ctx.SourceFile, unary)
			isNegative = op == ast.KindMinusToken
			if isNegative {
				fullRaw = "-" + ownRaw
			} else {
				fullRaw = "+" + ownRaw
			}
		}

		parent := skipParensUp(fullNumberNode.Parent)
		if parent == nil {
			return
		}

		var numericValue float64
		var bigintValue *big.Int
		if hasIgnoredValues || opts.ignoreArrayIndexes {
			if isBigInt {
				text := strings.TrimSuffix(ownRaw, "n")
				bigintValue, _ = new(big.Int).SetString(text, 0)
				if bigintValue == nil {
					bigintValue = new(big.Int)
				}
				if isNegative {
					bigintValue.Neg(bigintValue)
				}
			} else {
				numericValue, _ = parseRawNumericValue(ownRaw)
				if isNegative {
					numericValue = -numericValue
				}
			}
		}

		if hasIgnoredValues {
			var valueKey string
			if isBigInt {
				valueKey = "bigint:" + bigintValue.String()
			} else {
				valueKey = normalizeFloatKey(numericValue)
			}
			if opts.ignore[valueKey] {
				return
			}
		}

		if opts.ignoreDefaultValues && isDefaultValue(fullNumberNode, parent) {
			return
		}
		if opts.ignoreClassFieldInitialValues && isClassFieldInitialValue(fullNumberNode, parent) {
			return
		}
		// TypeScript-aware checks operate on the original literal node,
		// which re-derives its own unwrapped parent.
		if opts.ignoreEnums && isParentTSEnumDeclaration(node) {
			return
		}
		if opts.ignoreNumericLiteralTypes && isTSNumericLiteralType(node) {
			return
		}
		if opts.ignoreTypeIndexes && isAncestorTSIndexedAccessType(node) {
			return
		}
		if opts.ignoreReadonlyClassProperties && isParentTSReadonlyPropertyDefinition(node) {
			return
		}
		if isParseIntRadix(fullNumberNode, parent) {
			return
		}
		if isJSXNumber(parent) {
			return
		}
		if opts.ignoreArrayIndexes && isArrayIndex(fullNumberNode, parent, numericValue, bigintValue, isBigInt) {
			return
		}

		if parent.Kind == ast.KindVariableDeclaration {
			if opts.enforceConst {
				declList := parent.Parent
				if declList != nil && declList.Kind == ast.KindVariableDeclarationList && !ast.IsVarConst(declList) {
					ctx.ReportRange(fullNumberRange, useConstMessage)
				}
			}
		} else if !isOkParent(parent, opts.detectObjects) {
			ctx.ReportRange(fullNumberRange, noMagicNumberMessage(fullRaw))
		}
	}

	return rule.RuleListeners{
		ast.KindNumericLiteral: func(node *ast.Node) {
			handleNumericNode(node, false)
		},
		ast.KindBigIntLiteral: func(node *ast.Node) {
			handleNumericNode(node, true)
		},
	}
}

// isOkParent checks if the parent node type is one that allows numbers without reporting.
func isOkParent(parent *ast.Node, detectObjects bool) bool {
	if detectObjects {
		return false
	}
	switch parent.Kind {
	case ast.KindObjectLiteralExpression:
		return true
	case ast.KindPropertyAssignment:
		return true
	case ast.KindShorthandPropertyAssignment:
		spa := parent.AsShorthandPropertyAssignment()
		if spa.ObjectAssignmentInitializer != nil {
			return false
		}
		return true
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return isObjectLiteralMember(parent)
	case ast.KindComputedPropertyName:
		gp := parent.Parent
		if gp != nil && (gp.Kind == ast.KindPropertyAssignment || gp.Kind == ast.KindShorthandPropertyAssignment || isObjectLiteralMember(gp)) {
			return true
		}
		return false
	case ast.KindBinaryExpression:
		op := parent.AsBinaryExpression().OperatorToken.Kind
		if isAssignmentOperator(op) {
			left := parent.AsBinaryExpression().Left
			if left != nil && ast.SkipParentheses(left).Kind == ast.KindIdentifier {
				return false
			}
			return true
		}
	}
	return false
}

// isObjectLiteralMember reports whether a method or accessor belongs to an
// object literal rather than a class. ESTree represents object-literal methods
// and accessors as Property, which upstream allows, while class members are
// MethodDefinition and stay reportable.
func isObjectLiteralMember(node *ast.Node) bool {
	return node.Parent != nil && node.Parent.Kind == ast.KindObjectLiteralExpression
}

func isAssignmentOperator(kind ast.Kind) bool {
	switch kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindBarBarEqualsToken,
		ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// getLiteralParent returns the "logical parent" of a numeric literal node,
// skipping ParenthesizedExpression and unary +/- prefix.
func getLiteralParent(node *ast.Node) *ast.Node {
	p := skipParensUp(node.Parent)
	if p != nil && p.Kind == ast.KindPrefixUnaryExpression {
		pref := p.AsPrefixUnaryExpression()
		if pref.Operator == ast.KindMinusToken || pref.Operator == ast.KindPlusToken {
			return skipParensUp(p.Parent)
		}
	}
	return p
}

func isParentTSEnumDeclaration(node *ast.Node) bool {
	parent := getLiteralParent(node)
	return parent != nil && parent.Kind == ast.KindEnumMember
}

func isTSNumericLiteralType(node *ast.Node) bool {
	// For negative numbers, step up past parentheses and unary minus
	current := node
	p := skipParensUp(current.Parent)
	if p != nil && p.Kind == ast.KindPrefixUnaryExpression {
		pref := p.AsPrefixUnaryExpression()
		if pref.Operator == ast.KindMinusToken {
			current = p
		}
	}

	if current.Parent == nil || current.Parent.Kind != ast.KindLiteralType {
		return false
	}

	gp := current.Parent.Parent
	if gp == nil {
		return false
	}
	for gp != nil && gp.Kind == ast.KindParenthesizedType {
		gp = gp.Parent
	}
	if gp != nil && gp.Kind == ast.KindTypeAliasDeclaration {
		return true
	}

	if gp != nil && gp.Kind == ast.KindUnionType {
		ancestor := gp.Parent
		for ancestor != nil && (ancestor.Kind == ast.KindUnionType || ancestor.Kind == ast.KindParenthesizedType) {
			ancestor = ancestor.Parent
		}
		return ancestor != nil && ancestor.Kind == ast.KindTypeAliasDeclaration
	}

	return false
}

func isParentTSReadonlyPropertyDefinition(node *ast.Node) bool {
	parent := getLiteralParent(node)
	if parent == nil || parent.Kind != ast.KindPropertyDeclaration {
		return false
	}
	return ast.HasSyntacticModifier(parent, ast.ModifierFlagsReadonly)
}

func isAncestorTSIndexedAccessType(node *ast.Node) bool {
	ancestor := getLiteralParent(node)

	for ancestor != nil && ancestor.Parent != nil &&
		(ancestor.Parent.Kind == ast.KindUnionType ||
			ancestor.Parent.Kind == ast.KindIntersectionType ||
			ancestor.Parent.Kind == ast.KindParenthesizedType) {
		ancestor = ancestor.Parent
	}

	return ancestor != nil && ancestor.Parent != nil && ancestor.Parent.Kind == ast.KindIndexedAccessType
}

// isDefaultValue checks if the fullNumberNode is a default value assignment.
// parent is the already-resolved logical parent (parens skipped).
func isDefaultValue(fullNumberNode *ast.Node, parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindParameter {
		init := parent.AsParameterDeclaration().Initializer
		return init != nil && ast.SkipParentheses(init) == fullNumberNode
	}
	if parent.Kind == ast.KindBindingElement {
		init := parent.AsBindingElement().Initializer
		return init != nil && ast.SkipParentheses(init) == fullNumberNode
	}
	if parent.Kind == ast.KindShorthandPropertyAssignment {
		spa := parent.AsShorthandPropertyAssignment()
		return spa.ObjectAssignmentInitializer != nil &&
			ast.SkipParentheses(spa.ObjectAssignmentInitializer) == fullNumberNode
	}
	if parent.Kind == ast.KindBinaryExpression {
		binExpr := parent.AsBinaryExpression()
		if binExpr.OperatorToken.Kind == ast.KindEqualsToken && ast.SkipParentheses(binExpr.Right) == fullNumberNode {
			return isInsideDestructuringAssignment(parent)
		}
	}
	return false
}

func isInsideDestructuringAssignment(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
			parent = parent.Parent
			continue
		case ast.KindSpreadElement, ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
			parent = parent.Parent
			continue
		case ast.KindParenthesizedExpression:
			parent = parent.Parent
			continue
		case ast.KindBinaryExpression:
			binExpr := parent.AsBinaryExpression()
			if binExpr.OperatorToken.Kind == ast.KindEqualsToken {
				left := binExpr.Left
				return left != nil && (left.Kind == ast.KindArrayLiteralExpression || left.Kind == ast.KindObjectLiteralExpression)
			}
			return false
		case ast.KindForOfStatement, ast.KindForInStatement:
			return true
		default:
			return false
		}
	}
	return false
}

// isClassFieldInitialValue checks if the fullNumberNode is the direct initializer
// of a class field. parent is the already-resolved logical parent (parens skipped).
func isClassFieldInitialValue(fullNumberNode *ast.Node, parent *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindPropertyDeclaration {
		return false
	}
	init := parent.AsPropertyDeclaration().Initializer
	return init != nil && ast.SkipParentheses(init) == fullNumberNode
}

// isParseIntRadix checks if the fullNumberNode is used as the radix argument in
// parseInt() or Number.parseInt(). parent is the already-resolved logical parent.
func isParseIntRadix(fullNumberNode *ast.Node, parent *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	args := call.Arguments.Nodes
	if len(args) < 2 || ast.SkipParentheses(args[1]) != fullNumberNode {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil {
		return false
	}
	if callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "parseInt" {
		return true
	}
	return utils.IsSpecificMemberAccess(call.Expression, "Number", "parseInt")
}

// isJSXNumber checks if the parent is a JSX node.
func isJSXNumber(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindJsxExpression, ast.KindJsxAttribute, ast.KindJsxElement,
		ast.KindJsxSelfClosingElement, ast.KindJsxOpeningElement,
		ast.KindJsxFragment, ast.KindJsxSpreadAttribute:
		return true
	}
	return false
}

// isArrayIndex checks if the fullNumberNode is used as a valid array index.
// parent is the already-resolved logical parent (parens skipped).
func isArrayIndex(fullNumberNode *ast.Node, parent *ast.Node, numericValue float64, bigintValue *big.Int, isBigInt bool) bool {
	if parent == nil || parent.Kind != ast.KindElementAccessExpression {
		return false
	}
	elemAccess := parent.AsElementAccessExpression()
	if ast.SkipParentheses(elemAccess.ArgumentExpression) != fullNumberNode {
		return false
	}

	if isBigInt {
		if bigintValue == nil {
			return false
		}
		if bigintValue.Sign() < 0 {
			return false
		}
		maxIdx := new(big.Int).SetUint64(maxArrayLength)
		return bigintValue.Cmp(maxIdx) < 0
	}

	if !isIntegerValue(numericValue) {
		return false
	}
	return numericValue >= 0 && numericValue < maxArrayLength
}

func isIntegerValue(v float64) bool {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return false
	}
	return v == math.Floor(v)
}
