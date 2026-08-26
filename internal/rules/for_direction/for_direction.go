package for_direction

import (
	"errors"
	"math"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var incorrectDirectionMessage = rule.RuleMessage{
	Id:          "incorrectDirection",
	Description: "The update clause in this loop moves the variable in the wrong direction.",
}

type staticNumericKind uint8

const (
	staticNumber staticNumericKind = iota
	staticBigInt
	staticBoolean
)

type staticNumeric struct {
	value float64
	kind  staticNumericKind
}

type staticCacheEntry struct {
	value    staticNumeric
	resolved bool
	ok       bool
}

type forDirectionChecker struct {
	ctx         rule.RuleContext
	staticCache map[*ast.Symbol]staticCacheEntry
}

func (checker *forDirectionChecker) check(node *ast.Node) {
	forStatement := node.AsForStatement()
	if forStatement == nil || forStatement.Condition == nil || forStatement.Incrementor == nil {
		return
	}

	condition := forStatement.Condition
	if condition.Kind != ast.KindBinaryExpression {
		return
	}
	binary := condition.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return
	}

	operator := binary.OperatorToken.Kind
	if operator != ast.KindLessThanToken && operator != ast.KindLessThanEqualsToken &&
		operator != ast.KindGreaterThanToken && operator != ast.KindGreaterThanEqualsToken {
		return
	}
	incrementor := skipParentheses(forStatement.Incrementor)
	if incrementor == nil ||
		(incrementor.Kind != ast.KindPostfixUnaryExpression &&
			incrementor.Kind != ast.KindPrefixUnaryExpression &&
			incrementor.Kind != ast.KindBinaryExpression) {
		return
	}

	left := skipParentheses(binary.Left)
	if left != nil && left.Kind == ast.KindIdentifier {
		direction, modifications := checker.modificationDirection(
			incrementor,
			left.AsIdentifier().Text,
		)
		if modifications == 1 && direction == wrongUpdateDirection(operator, true) {
			checker.report(node, forStatement)
			return
		}
	}

	right := skipParentheses(binary.Right)
	if right == nil || right.Kind != ast.KindIdentifier {
		return
	}
	direction, modifications := checker.modificationDirection(
		incrementor,
		right.AsIdentifier().Text,
	)
	if modifications == 1 && direction == wrongUpdateDirection(operator, false) {
		checker.report(node, forStatement)
	}
}

func wrongUpdateDirection(operator ast.Kind, counterOnLeft bool) int {
	switch operator {
	case ast.KindLessThanToken, ast.KindLessThanEqualsToken:
		if counterOnLeft {
			return -1
		}
		return 1
	case ast.KindGreaterThanToken, ast.KindGreaterThanEqualsToken:
		if counterOnLeft {
			return 1
		}
		return -1
	default:
		return 0
	}
}

// modificationDirection returns the direction of the only expression that
// modifies counter, together with the number of modifying expressions found.
// A count other than one makes the loop ambiguous and suppresses a report.
func (checker *forDirectionChecker) modificationDirection(node *ast.Node, counter string) (int, int) {
	if node == nil {
		return 0, 0
	}
	node = skipParentheses(node)

	switch node.Kind {
	case ast.KindPostfixUnaryExpression:
		update := node.AsPostfixUnaryExpression()
		if update == nil || !isIdentifierNamed(update.Operand, counter) {
			return 0, 0
		}
		switch update.Operator {
		case ast.KindPlusPlusToken:
			return 1, 1
		case ast.KindMinusMinusToken:
			return -1, 1
		default:
			return 0, 0
		}

	case ast.KindPrefixUnaryExpression:
		update := node.AsPrefixUnaryExpression()
		if update == nil || !isIdentifierNamed(update.Operand, counter) {
			return 0, 0
		}
		switch update.Operator {
		case ast.KindPlusPlusToken:
			return 1, 1
		case ast.KindMinusMinusToken:
			return -1, 1
		default:
			return 0, 0
		}

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return 0, 0
		}
		operator := binary.OperatorToken.Kind
		if operator == ast.KindCommaToken {
			leftDirection, leftCount := checker.modificationDirection(binary.Left, counter)
			rightDirection, rightCount := checker.modificationDirection(binary.Right, counter)
			count := leftCount + rightCount
			if count != 1 {
				return 0, count
			}
			if leftCount == 1 {
				return leftDirection, 1
			}
			return rightDirection, 1
		}

		if !ast.IsAssignmentOperator(operator) || !isIdentifierNamed(binary.Left, counter) {
			return 0, 0
		}

		direction := 0
		if operator == ast.KindPlusEqualsToken || operator == ast.KindMinusEqualsToken {
			direction = checker.staticDirection(binary.Right)
			if operator == ast.KindMinusEqualsToken {
				direction = -direction
			}
		}
		return direction, 1
	}

	return 0, 0
}

func isIdentifierNamed(node *ast.Node, name string) bool {
	node = skipParentheses(node)
	return node != nil && node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == name
}

func skipParentheses(node *ast.Node) *ast.Node {
	if node != nil && node.Kind == ast.KindParenthesizedExpression {
		return ast.SkipParentheses(node)
	}
	return node
}

func (checker *forDirectionChecker) staticDirection(node *ast.Node) int {
	value, ok := checker.staticNumericValue(node)
	if !ok || value.value == 0 || math.IsNaN(value.value) {
		return 0
	}
	if value.value < 0 {
		return -1
	}
	return 1
}

func (checker *forDirectionChecker) staticNumericValue(node *ast.Node) (staticNumeric, bool) {
	if node == nil {
		return staticNumeric{}, false
	}
	node = ast.SkipOuterExpressions(node, ast.OEKAll)

	switch node.Kind {
	case ast.KindNumericLiteral:
		value, ok := parseNonNegativeNumeric(node.AsNumericLiteral().Text)
		return staticNumeric{value: value, kind: staticNumber}, ok

	case ast.KindBigIntLiteral:
		text := node.AsBigIntLiteral().Text
		if len(text) == 0 || text[len(text)-1] != 'n' {
			return staticNumeric{}, false
		}
		value, ok := parseNonNegativeNumeric(text[:len(text)-1])
		return staticNumeric{value: value, kind: staticBigInt}, ok

	case ast.KindTrueKeyword:
		return staticNumeric{value: 1, kind: staticBoolean}, true

	case ast.KindFalseKeyword:
		return staticNumeric{kind: staticBoolean}, true

	case ast.KindIdentifier:
		return checker.staticIdentifierValue(node)

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix == nil {
			return staticNumeric{}, false
		}
		value, ok := checker.staticNumericValue(prefix.Operand)
		if !ok {
			return staticNumeric{}, false
		}
		switch prefix.Operator {
		case ast.KindPlusToken:
			if value.kind == staticBigInt {
				return staticNumeric{}, false
			}
			value.kind = staticNumber
			return value, true
		case ast.KindMinusToken:
			value.value = -value.value
			if value.kind != staticBigInt {
				value.kind = staticNumber
			}
			return value, true
		case ast.KindTildeToken:
			if value.kind == staticBigInt {
				value.value = -value.value - 1
				return value, true
			}
			return staticNumeric{value: float64(^toInt32(value.value)), kind: staticNumber}, true
		case ast.KindExclamationToken:
			if value.value == 0 || math.IsNaN(value.value) {
				return staticNumeric{value: 1, kind: staticBoolean}, true
			}
			return staticNumeric{kind: staticBoolean}, true
		default:
			return staticNumeric{}, false
		}

	case ast.KindBinaryExpression:
		return checker.staticBinaryValue(node.AsBinaryExpression())

	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		if conditional == nil {
			return staticNumeric{}, false
		}
		condition, ok := checker.staticNumericValue(conditional.Condition)
		if !ok {
			return staticNumeric{}, false
		}
		if condition.value != 0 && !math.IsNaN(condition.value) {
			return checker.staticNumericValue(conditional.WhenTrue)
		}
		return checker.staticNumericValue(conditional.WhenFalse)
	}

	return staticNumeric{}, false
}

func (checker *forDirectionChecker) staticBinaryValue(binary *ast.BinaryExpression) (staticNumeric, bool) {
	if binary == nil || binary.OperatorToken == nil {
		return staticNumeric{}, false
	}
	operator := binary.OperatorToken.Kind
	if operator == ast.KindEqualsToken || operator == ast.KindCommaToken {
		return checker.staticNumericValue(binary.Right)
	}

	left, leftOK := checker.staticNumericValue(binary.Left)
	if !leftOK {
		return staticNumeric{}, false
	}
	if operator == ast.KindAmpersandAmpersandToken {
		if left.value == 0 || math.IsNaN(left.value) {
			return left, true
		}
		return checker.staticNumericValue(binary.Right)
	}
	if operator == ast.KindBarBarToken {
		if left.value != 0 && !math.IsNaN(left.value) {
			return left, true
		}
		return checker.staticNumericValue(binary.Right)
	}
	if operator == ast.KindQuestionQuestionToken {
		return left, true
	}

	right, rightOK := checker.staticNumericValue(binary.Right)
	if !rightOK {
		return staticNumeric{}, false
	}

	boolean := func(value bool) (staticNumeric, bool) {
		if value {
			return staticNumeric{value: 1, kind: staticBoolean}, true
		}
		return staticNumeric{kind: staticBoolean}, true
	}
	const maxSafeInteger = 1<<53 - 1
	leftUnsafeBigInt := left.kind == staticBigInt && math.Abs(left.value) > maxSafeInteger
	rightUnsafeBigInt := right.kind == staticBigInt && math.Abs(right.value) > maxSafeInteger
	if leftUnsafeBigInt || rightUnsafeBigInt {
		switch operator {
		case ast.KindEqualsEqualsEqualsToken:
			if left.kind != right.kind {
				return boolean(false)
			}
		case ast.KindExclamationEqualsEqualsToken:
			if left.kind != right.kind {
				return boolean(true)
			}
		}
		return staticNumeric{}, false
	}
	switch operator {
	case ast.KindEqualsEqualsEqualsToken:
		return boolean(left.kind == right.kind && left.value == right.value)
	case ast.KindExclamationEqualsEqualsToken:
		return boolean(left.kind != right.kind || left.value != right.value)
	case ast.KindEqualsEqualsToken:
		return boolean(left.value == right.value)
	case ast.KindExclamationEqualsToken:
		return boolean(left.value != right.value)
	}

	leftIsBigInt := left.kind == staticBigInt
	rightIsBigInt := right.kind == staticBigInt
	if leftIsBigInt != rightIsBigInt {
		return staticNumeric{}, false
	}
	kind := staticNumber
	if leftIsBigInt {
		kind = staticBigInt
	}

	switch operator {
	case ast.KindPlusToken:
		return staticNumeric{value: left.value + right.value, kind: kind}, true
	case ast.KindMinusToken:
		return staticNumeric{value: left.value - right.value, kind: kind}, true
	case ast.KindAsteriskToken:
		return staticNumeric{value: left.value * right.value, kind: kind}, true
	case ast.KindSlashToken:
		if kind == staticBigInt {
			if right.value == 0 {
				return staticNumeric{}, false
			}
			return staticNumeric{value: math.Trunc(left.value / right.value), kind: kind}, true
		}
		return staticNumeric{value: left.value / right.value, kind: kind}, true
	case ast.KindPercentToken:
		if kind == staticBigInt && right.value == 0 {
			return staticNumeric{}, false
		}
		return staticNumeric{value: math.Mod(left.value, right.value), kind: kind}, true
	case ast.KindAsteriskAsteriskToken:
		if kind == staticBigInt &&
			(right.value < 0 || right.value != math.Trunc(right.value) || math.IsInf(right.value, 0)) {
			return staticNumeric{}, false
		}
		return staticNumeric{value: math.Pow(left.value, right.value), kind: kind}, true
	case ast.KindLessThanToken:
		return boolean(left.value < right.value)
	case ast.KindLessThanEqualsToken:
		return boolean(left.value <= right.value)
	case ast.KindGreaterThanToken:
		return boolean(left.value > right.value)
	case ast.KindGreaterThanEqualsToken:
		return boolean(left.value >= right.value)
	}

	if kind == staticBigInt || math.IsNaN(left.value) || math.IsNaN(right.value) ||
		math.IsInf(left.value, 0) || math.IsInf(right.value, 0) {
		return staticNumeric{}, false
	}

	switch operator {
	case ast.KindAmpersandToken:
		return staticNumeric{value: float64(toInt32(left.value) & toInt32(right.value)), kind: staticNumber}, true
	case ast.KindBarToken:
		return staticNumeric{value: float64(toInt32(left.value) | toInt32(right.value)), kind: staticNumber}, true
	case ast.KindCaretToken:
		return staticNumeric{value: float64(toInt32(left.value) ^ toInt32(right.value)), kind: staticNumber}, true
	case ast.KindLessThanLessThanToken:
		return staticNumeric{value: float64(toInt32(left.value) << (toUint32(right.value) & 31)), kind: staticNumber}, true
	case ast.KindGreaterThanGreaterThanToken:
		return staticNumeric{value: float64(toInt32(left.value) >> (toUint32(right.value) & 31)), kind: staticNumber}, true
	case ast.KindGreaterThanGreaterThanGreaterThanToken:
		return staticNumeric{value: float64(toUint32(left.value) >> (toUint32(right.value) & 31)), kind: staticNumber}, true
	default:
		return staticNumeric{}, false
	}
}

func (checker *forDirectionChecker) staticIdentifierValue(node *ast.Node) (staticNumeric, bool) {
	if checker.ctx.Refs == nil {
		return staticNumeric{}, false
	}
	symbol := checker.ctx.Refs.Resolve(node)
	if symbol == nil {
		return staticNumeric{}, false
	}
	if cached, ok := checker.staticCache[symbol]; ok {
		if !cached.resolved {
			return staticNumeric{}, false
		}
		return cached.value, cached.ok
	}
	if checker.staticCache == nil {
		checker.staticCache = make(map[*ast.Symbol]staticCacheEntry)
	}
	checker.staticCache[symbol] = staticCacheEntry{}

	var declaration *ast.Node
	for _, candidate := range symbol.Declarations {
		if candidate == nil || candidate.Kind != ast.KindVariableDeclaration ||
			candidate.Name() == nil || candidate.Name().Kind != ast.KindIdentifier {
			continue
		}
		if declaration != nil {
			checker.staticCache[symbol] = staticCacheEntry{resolved: true}
			return staticNumeric{}, false
		}
		declaration = candidate
	}
	if declaration == nil {
		checker.staticCache[symbol] = staticCacheEntry{resolved: true}
		return staticNumeric{}, false
	}

	declarationList := declaration.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList {
		checker.staticCache[symbol] = staticCacheEntry{resolved: true}
		return staticNumeric{}, false
	}
	if declarationList.Flags&ast.NodeFlagsConst == 0 {
		for _, reference := range checker.ctx.Refs.References(symbol) {
			if utils.IsWriteReference(reference) {
				checker.staticCache[symbol] = staticCacheEntry{resolved: true}
				return staticNumeric{}, false
			}
		}
	}

	initializer := declaration.AsVariableDeclaration().Initializer
	value, ok := checker.staticNumericValue(initializer)
	checker.staticCache[symbol] = staticCacheEntry{value: value, resolved: true, ok: ok}
	return value, ok
}

func parseNonNegativeNumeric(text string) (float64, bool) {
	if len(text) > 2 && text[0] == '0' {
		base := 0
		switch text[1] {
		case 'b', 'B':
			base = 2
		case 'o', 'O':
			base = 8
		case 'x', 'X':
			base = 16
		}
		if base != 0 {
			value, err := strconv.ParseUint(text[2:], base, 64)
			if err == nil {
				return float64(value), true
			}
			if errors.Is(err, strconv.ErrRange) {
				return math.Inf(1), true
			}
			return 0, false
		}
	}

	value, err := strconv.ParseFloat(text, 64)
	if err == nil || math.IsInf(value, 1) {
		return value, true
	}
	return 0, false
}

func toUint32(value float64) uint32 {
	if value == 0 || math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	value = math.Mod(math.Trunc(value), 1<<32)
	if value < 0 {
		value += 1 << 32
	}
	return uint32(value)
}

func toInt32(value float64) int32 {
	return int32(toUint32(value))
}

func (checker *forDirectionChecker) report(node *ast.Node, forStatement *ast.ForStatement) {
	text := checker.ctx.SourceFile.Text()
	start := scanner.SkipTrivia(text, node.Pos())
	end := forStatement.Incrementor.End()
	closeParen := scanner.SkipTrivia(text, end)
	if closeParen < len(text) && text[closeParen] == ')' {
		end = closeParen + 1
	}
	checker.ctx.ReportRange(core.NewTextRange(start, end), incorrectDirectionMessage)
}

// ForDirectionRule enforces that for loop update clauses move the counter in the right direction.
var ForDirectionRule = rule.Rule{
	Name:   "for-direction",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		checker := &forDirectionChecker{ctx: ctx}
		return rule.RuleListeners{ast.KindForStatement: checker.check}
	},
}
