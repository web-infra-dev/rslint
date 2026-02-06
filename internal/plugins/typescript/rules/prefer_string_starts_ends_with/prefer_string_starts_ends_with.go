package prefer_string_starts_ends_with

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferStartsWithMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferStartsWith",
		Description: "Use 'String#startsWith' method instead.",
	}
}

func buildPreferEndsWithMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferEndsWith",
		Description: "Use the 'String#endsWith' method instead.",
	}
}

type Options struct {
	AllowSingleElementEquality *string `json:"allowSingleElementEquality"`
}

var defaultOpts = Options{
	AllowSingleElementEquality: utils.Ref("never"),
}

// callInfo holds parsed call expression info
type callInfo struct {
	object     *ast.Node
	methodName string
	args       []*ast.Node
	isOptional bool
}

// regexpInfo holds parsed regex info
type regexpInfo struct {
	text    string
	isStart bool
}

// ruleHelper encapsulates all helper functions that depend on RuleContext
type ruleHelper struct {
	ctx                        rule.RuleContext
	allowSingleElementEquality bool
}

func (h *ruleHelper) isStringType(node *ast.Node) bool {
	t := utils.GetConstrainedTypeAtLocation(h.ctx.TypeChecker, node)
	return utils.GetTypeName(h.ctx.TypeChecker, t) == "string"
}

func (h *ruleHelper) getNodeText(node *ast.Node) string {
	r := utils.TrimNodeTextRange(h.ctx.SourceFile, node)
	return h.ctx.SourceFile.Text()[r.Pos():r.End()]
}

func (h *ruleHelper) isSameNode(a, b *ast.Node) bool {
	if a == nil || b == nil {
		return false
	}
	return h.getNodeText(a) == h.getNodeText(b)
}

func isNumber(node *ast.Node, value int) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindNumericLiteral:
		n := node.AsNumericLiteral()
		if n == nil {
			return false
		}
		v, err := strconv.Atoi(n.Text)
		return err == nil && v == value
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary == nil {
			return false
		}
		if unary.Operator == ast.KindMinusToken {
			return isNumber(unary.Operand, -value)
		}
	}
	return false
}

func isNullLiteral(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindNullKeyword
}

func isNegativeOp(op ast.Kind) bool {
	return op == ast.KindExclamationEqualsToken || op == ast.KindExclamationEqualsEqualsToken
}

func isCharacter(node *ast.Node) bool {
	if node == nil {
		return false
	}
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindStringLiteral {
		lit := n.AsStringLiteral()
		if lit == nil {
			return false
		}
		// Check JavaScript string length (UTF-16 code units), not Unicode code points.
		// Characters above U+FFFF need 2 UTF-16 code units (surrogate pairs).
		return jsStringLength(lit.Text) == 1
	}
	return false
}

func jsStringLength(s string) int {
	n := 0
	for _, r := range s {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

func isStringLiteral(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindStringLiteral
}

func getStringLength(node *ast.Node) int {
	if node == nil {
		return -1
	}
	if node.Kind == ast.KindStringLiteral {
		lit := node.AsStringLiteral()
		if lit == nil {
			return -1
		}
		return len(lit.Text)
	}
	return -1
}

func (h *ruleHelper) isLengthProperty(node *ast.Node, obj *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		propAccess := node.AsPropertyAccessExpression()
		if propAccess == nil {
			return false
		}
		return propAccess.Name().Text() == "length" && h.isSameNode(propAccess.Expression, obj)
	}
	return false
}

func (h *ruleHelper) isLengthExpression(node *ast.Node, search *ast.Node) bool {
	strLen := getStringLength(search)
	if strLen >= 0 && isNumber(node, strLen) {
		return true
	}
	return h.isLengthProperty(node, search)
}

func (h *ruleHelper) isLastIndexExpression(node *ast.Node, obj *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary == nil {
			return false
		}
		return binary.OperatorToken.Kind == ast.KindMinusToken &&
			h.isLengthProperty(binary.Left, obj) &&
			isNumber(binary.Right, 1)
	}
	return false
}

func (h *ruleHelper) isLengthAheadOfEnd(node *ast.Node, search *ast.Node, parent *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindPrefixUnaryExpression {
		unary := node.AsPrefixUnaryExpression()
		if unary != nil && unary.Operator == ast.KindMinusToken {
			return h.isLengthExpression(unary.Operand, search)
		}
	}
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken.Kind == ast.KindMinusToken {
			return h.isLengthProperty(binary.Left, parent) &&
				h.isLengthExpression(binary.Right, search)
		}
	}
	return false
}

func getCallInfo(node *ast.Node) *callInfo {
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil
	}
	callExpr := node.AsCallExpression()
	if callExpr == nil {
		return nil
	}

	expr := callExpr.Expression
	if expr == nil {
		return nil
	}

	var args []*ast.Node
	if callExpr.Arguments != nil {
		args = callExpr.Arguments.Nodes
	}

	if expr.Kind == ast.KindPropertyAccessExpression {
		propAccess := expr.AsPropertyAccessExpression()
		if propAccess == nil {
			return nil
		}
		return &callInfo{
			object:     propAccess.Expression,
			methodName: propAccess.Name().Text(),
			args:       args,
			isOptional: callExpr.QuestionDotToken != nil || propAccess.QuestionDotToken != nil,
		}
	}

	return nil
}

func parseRegExpPattern(pattern string, flags string) (string, bool, bool) {
	if strings.Contains(flags, "i") || strings.Contains(flags, "m") {
		return "", false, false
	}

	isStart := strings.HasPrefix(pattern, "^")
	isEnd := strings.HasSuffix(pattern, "$")

	if isStart == isEnd {
		return "", false, false
	}

	var inner string
	if isStart {
		inner = pattern[1:]
	} else {
		inner = pattern[:len(pattern)-1]
	}

	text, ok := extractStaticString(inner)
	if !ok {
		return "", false, false
	}

	return text, isStart, true
}

func getRegExpInfo(node *ast.Node) *regexpInfo {
	if node == nil {
		return nil
	}

	var pattern, flags string

	switch node.Kind {
	case ast.KindRegularExpressionLiteral:
		regexLit := node.AsRegularExpressionLiteral()
		if regexLit == nil {
			return nil
		}
		text := regexLit.Text
		lastSlash := strings.LastIndex(text, "/")
		if lastSlash <= 0 {
			return nil
		}
		pattern = text[1:lastSlash]
		flags = text[lastSlash+1:]
	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		if newExpr == nil {
			return nil
		}
		if newExpr.Expression == nil || newExpr.Expression.Kind != ast.KindIdentifier {
			return nil
		}
		if newExpr.Expression.Text() != "RegExp" {
			return nil
		}
		if newExpr.Arguments == nil || len(newExpr.Arguments.Nodes) < 1 {
			return nil
		}
		arg0 := newExpr.Arguments.Nodes[0]
		if arg0.Kind != ast.KindStringLiteral {
			return nil
		}
		pattern = arg0.AsStringLiteral().Text
		if len(newExpr.Arguments.Nodes) >= 2 {
			arg1 := newExpr.Arguments.Nodes[1]
			if arg1.Kind == ast.KindStringLiteral {
				flags = arg1.AsStringLiteral().Text
			}
		}
	default:
		return nil
	}

	text, isStart, ok := parseRegExpPattern(pattern, flags)
	if !ok {
		return nil
	}

	return &regexpInfo{text: text, isStart: isStart}
}

func (h *ruleHelper) resolveRegExpFromNode(node *ast.Node) *regexpInfo {
	info := getRegExpInfo(node)
	if info != nil {
		return info
	}

	if node.Kind == ast.KindIdentifier {
		sym := h.ctx.TypeChecker.GetSymbolAtLocation(node)
		if sym != nil {
			decls := sym.Declarations
			for _, decl := range decls {
				if decl.Kind == ast.KindVariableDeclaration {
					varDecl := decl.AsVariableDeclaration()
					if varDecl != nil && varDecl.Initializer != nil {
						info = getRegExpInfo(varDecl.Initializer)
						if info != nil {
							return info
						}
					}
				}
			}
		}
	}

	return nil
}

func escapeStringForJs(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

func extractStaticString(pattern string) (string, bool) {
	if pattern == "" {
		return "", true
	}

	var result strings.Builder
	i := 0
	for i < len(pattern) {
		ch := pattern[i]

		switch ch {
		case '.', '*', '+', '?', '|', '(', ')', '[', ']', '{', '}', '^', '$':
			return "", false
		case '\\':
			if i+1 >= len(pattern) {
				return "", false
			}
			next := pattern[i+1]
			switch next {
			case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B', 'p', 'P', 'k':
				return "", false
			case 'n':
				result.WriteByte('\n')
			case 'r':
				result.WriteByte('\r')
			case 't':
				result.WriteByte('\t')
			default:
				result.WriteByte(next)
			}
			i += 2
		default:
			result.WriteByte(ch)
			i++
		}
	}

	return result.String(), true
}

func needsParentheses(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression, ast.KindCallExpression, ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression:
		return false
	}
	return true
}

var PreferStringStartsEndsWithRule = rule.CreateRule(rule.Rule{
	Name: "prefer-string-starts-ends-with",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := defaultOpts

		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if allowSingleElementEquality, ok := optsMap["allowSingleElementEquality"].(string); ok {
					opts.AllowSingleElementEquality = utils.Ref(allowSingleElementEquality)
				}
			}
		}

		h := &ruleHelper{
			ctx:                        ctx,
			allowSingleElementEquality: opts.AllowSingleElementEquality != nil && *opts.AllowSingleElementEquality == "always",
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil {
					return
				}

				op := binary.OperatorToken.Kind

				switch op {
				case ast.KindEqualsEqualsToken, ast.KindEqualsEqualsEqualsToken,
					ast.KindExclamationEqualsToken, ast.KindExclamationEqualsEqualsToken:
				default:
					return
				}

				left := binary.Left
				right := binary.Right

				h.checkCharacterAccess(left, right, op)
				h.checkIndexOf(left, right, op)
				h.checkLastIndexOf(left, right, op)
				h.checkMatch(left, right, op)
				h.checkSliceSubstring(left, right, op)
			},

			ast.KindCallExpression: func(node *ast.Node) {
				h.checkRegExpTest(node)
			},
		}
	},
})

func (h *ruleHelper) checkCharacterAccess(left, right *ast.Node, op ast.Kind) {
	if h.allowSingleElementEquality {
		return
	}

	// Skip parentheses on the right side for fix text
	innerRight := ast.SkipParentheses(right)
	rightText := h.getNodeText(innerRight)
	canFix := isCharacter(right)

	if left.Kind == ast.KindElementAccessExpression {
		ea := left.AsElementAccessExpression()
		if ea != nil && ea.Expression != nil && ea.ArgumentExpression != nil {
			obj := ea.Expression
			arg := ea.ArgumentExpression
			isOpt := ea.QuestionDotToken != nil

			if h.isStringType(obj) {
				objText := h.getNodeText(obj)
				optChain := ""
				if isOpt {
					optChain = "?"
				}

				if isNumber(arg, 0) {
					h.reportStartsWith(left.Parent, objText, optChain, rightText, op, canFix)
					return
				}
				if h.isLastIndexExpression(arg, obj) {
					h.reportEndsWith(left.Parent, objText, optChain, rightText, op, canFix)
					return
				}
			}
		}
	}

	if left.Kind == ast.KindCallExpression {
		ci := getCallInfo(left)
		if ci != nil && ci.methodName == "charAt" && len(ci.args) == 1 {
			obj := ci.object
			arg := ci.args[0]

			if h.isStringType(obj) {
				objText := h.getNodeText(obj)
				optChain := ""
				if ci.isOptional {
					optChain = "?"
				}

				if isNumber(arg, 0) {
					h.reportStartsWith(left.Parent, objText, optChain, rightText, op, canFix)
					return
				}
				if h.isLastIndexExpression(arg, obj) {
					h.reportEndsWith(left.Parent, objText, optChain, rightText, op, canFix)
					return
				}
			}
		}
	}
}

func (h *ruleHelper) checkIndexOf(left, right *ast.Node, op ast.Kind) {
	ci := getCallInfo(left)
	if ci == nil || ci.methodName != "indexOf" || len(ci.args) != 1 {
		return
	}

	if !isNumber(right, 0) {
		return
	}

	if !h.isStringType(ci.object) {
		return
	}

	objText := h.getNodeText(ci.object)
	optChain := ""
	if ci.isOptional {
		optChain = "?"
	}
	argText := h.getNodeText(ci.args[0])

	fixText := h.buildFixText(objText, optChain, "startsWith", argText, op)
	nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, left.Parent)
	h.ctx.ReportNodeWithFixes(left.Parent, buildPreferStartsWithMessage(),
		rule.RuleFixReplaceRange(nodeRange, fixText))
}

func (h *ruleHelper) checkLastIndexOf(left, right *ast.Node, op ast.Kind) {
	ci := getCallInfo(left)
	if ci == nil || ci.methodName != "lastIndexOf" || len(ci.args) != 1 {
		return
	}

	if !h.isStringType(ci.object) {
		return
	}

	if right.Kind != ast.KindBinaryExpression {
		return
	}
	rightBinary := right.AsBinaryExpression()
	if rightBinary == nil || rightBinary.OperatorToken.Kind != ast.KindMinusToken {
		return
	}

	if !h.isLengthProperty(rightBinary.Left, ci.object) {
		return
	}

	if !h.isLengthExpression(rightBinary.Right, ci.args[0]) {
		return
	}

	objText := h.getNodeText(ci.object)
	optChain := ""
	if ci.isOptional {
		optChain = "?"
	}
	argText := h.getNodeText(ci.args[0])

	fixText := h.buildFixText(objText, optChain, "endsWith", argText, op)
	nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, left.Parent)
	h.ctx.ReportNodeWithFixes(left.Parent, buildPreferEndsWithMessage(),
		rule.RuleFixReplaceRange(nodeRange, fixText))
}

func (h *ruleHelper) checkMatch(left, right *ast.Node, op ast.Kind) {
	ci := getCallInfo(left)
	if ci == nil || ci.methodName != "match" || len(ci.args) != 1 {
		return
	}

	if !isNullLiteral(right) {
		return
	}

	if !h.isStringType(ci.object) {
		return
	}

	info := h.resolveRegExpFromNode(ci.args[0])
	if info == nil {
		return
	}

	objText := h.getNodeText(ci.object)
	optChain := ""
	if ci.isOptional {
		optChain = "?"
	}

	var fixText string
	var msg rule.RuleMessage
	// === null means NOT matching, so we negate
	if info.isStart {
		msg = buildPreferStartsWithMessage()
		if isNegativeOp(op) {
			fixText = fmt.Sprintf("%s%s.startsWith(%s)", objText, optChain, escapeStringForJs(info.text))
		} else {
			fixText = fmt.Sprintf("!%s%s.startsWith(%s)", objText, optChain, escapeStringForJs(info.text))
		}
	} else {
		msg = buildPreferEndsWithMessage()
		if isNegativeOp(op) {
			fixText = fmt.Sprintf("%s%s.endsWith(%s)", objText, optChain, escapeStringForJs(info.text))
		} else {
			fixText = fmt.Sprintf("!%s%s.endsWith(%s)", objText, optChain, escapeStringForJs(info.text))
		}
	}

	nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, left.Parent)
	h.ctx.ReportNodeWithFixes(left.Parent, msg,
		rule.RuleFixReplaceRange(nodeRange, fixText))
}

func (h *ruleHelper) checkSliceSubstring(left, right *ast.Node, op ast.Kind) {
	ci := getCallInfo(left)
	if ci == nil {
		return
	}

	isSlice := ci.methodName == "slice"
	isSubstring := ci.methodName == "substring"
	if !isSlice && !isSubstring {
		return
	}

	if !h.isStringType(ci.object) {
		return
	}

	objText := h.getNodeText(ci.object)
	optChain := ""
	if ci.isOptional {
		optChain = "?"
	}
	rightText := h.getNodeText(right)

	isLooseEquality := op == ast.KindEqualsEqualsToken || op == ast.KindExclamationEqualsToken

	// startsWith: s.slice(0, N) === 'bar' or s.substring(0, N) === 'bar'
	if len(ci.args) == 2 && isNumber(ci.args[0], 0) {
		secondArg := ci.args[1]
		if !h.isLengthExpression(secondArg, right) {
			return
		}

		canFix := true
		if isLooseEquality && !isStringLiteral(right) {
			canFix = false
		}

		fixText := h.buildFixText(objText, optChain, "startsWith", rightText, op)

		if canFix {
			nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, left.Parent)
			h.ctx.ReportNodeWithFixes(left.Parent, buildPreferStartsWithMessage(),
				rule.RuleFixReplaceRange(nodeRange, fixText))
		} else {
			h.ctx.ReportNode(left.Parent, buildPreferStartsWithMessage())
		}
		return
	}

	// endsWith patterns
	if len(ci.args) >= 1 {
		firstArg := ci.args[0]

		// s.slice(-3) === 'bar'
		if isSlice && len(ci.args) == 1 && h.isLengthAheadOfEnd(firstArg, right, ci.object) {
			fixText := h.buildFixText(objText, optChain, "endsWith", rightText, op)
			nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, left.Parent)
			h.ctx.ReportNodeWithFixes(left.Parent, buildPreferEndsWithMessage(),
				rule.RuleFixReplaceRange(nodeRange, fixText))
			return
		}

		// s.slice(s.length - 3, s.length) === 'bar'
		if isSlice && len(ci.args) == 2 {
			secondArg := ci.args[1]
			if h.isLengthAheadOfEnd(firstArg, right, ci.object) && h.isLengthProperty(secondArg, ci.object) {
				fixText := h.buildFixText(objText, optChain, "endsWith", rightText, op)
				nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, left.Parent)
				h.ctx.ReportNodeWithFixes(left.Parent, buildPreferEndsWithMessage(),
					rule.RuleFixReplaceRange(nodeRange, fixText))
				return
			}
		}

		// s.substring(-3) === 'bar' (probable mistake, report but don't fix)
		if isSubstring && len(ci.args) == 1 && h.isLengthAheadOfEnd(firstArg, right, ci.object) {
			h.ctx.ReportNode(left.Parent, buildPreferEndsWithMessage())
			return
		}

		// s.substring(s.length - 3, s.length) === 'bar'
		if isSubstring && len(ci.args) == 2 {
			secondArg := ci.args[1]
			if h.isLengthAheadOfEnd(firstArg, right, ci.object) && h.isLengthProperty(secondArg, ci.object) {
				fixText := h.buildFixText(objText, optChain, "endsWith", rightText, op)
				nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, left.Parent)
				h.ctx.ReportNodeWithFixes(left.Parent, buildPreferEndsWithMessage(),
					rule.RuleFixReplaceRange(nodeRange, fixText))
				return
			}
		}
	}
}

func (h *ruleHelper) checkRegExpTest(node *ast.Node) {
	callExpr := node.AsCallExpression()
	if callExpr == nil {
		return
	}

	expr := callExpr.Expression
	if expr == nil || expr.Kind != ast.KindPropertyAccessExpression {
		return
	}
	propAccess := expr.AsPropertyAccessExpression()
	if propAccess == nil || propAccess.Name().Text() != "test" {
		return
	}

	if callExpr.Arguments == nil || len(callExpr.Arguments.Nodes) != 1 {
		return
	}

	regexNode := propAccess.Expression
	testArg := callExpr.Arguments.Nodes[0]

	info := h.resolveRegExpFromNode(regexNode)
	if info == nil {
		return
	}

	isOptional := callExpr.QuestionDotToken != nil || propAccess.QuestionDotToken != nil
	argText := h.getNodeText(testArg)

	if needsParentheses(testArg) {
		argText = "(" + argText + ")"
	}

	optChain := ""
	if isOptional {
		optChain = "?"
	}

	var fixText string
	var msg rule.RuleMessage
	if info.isStart {
		msg = buildPreferStartsWithMessage()
		fixText = fmt.Sprintf("%s%s.startsWith(%s)", argText, optChain, escapeStringForJs(info.text))
	} else {
		msg = buildPreferEndsWithMessage()
		fixText = fmt.Sprintf("%s%s.endsWith(%s)", argText, optChain, escapeStringForJs(info.text))
	}

	nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, node)
	h.ctx.ReportNodeWithFixes(node, msg,
		rule.RuleFixReplaceRange(nodeRange, fixText))
}

func (h *ruleHelper) buildFixText(objText, optChain, method, argText string, op ast.Kind) string {
	if isNegativeOp(op) {
		return fmt.Sprintf("!%s%s.%s(%s)", objText, optChain, method, argText)
	}
	return fmt.Sprintf("%s%s.%s(%s)", objText, optChain, method, argText)
}

func (h *ruleHelper) reportStartsWith(reportNode *ast.Node, objText, optChain, rightText string, op ast.Kind, canFix bool) {
	fixText := h.buildFixText(objText, optChain, "startsWith", rightText, op)
	if canFix {
		nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, reportNode)
		h.ctx.ReportNodeWithFixes(reportNode, buildPreferStartsWithMessage(),
			rule.RuleFixReplaceRange(nodeRange, fixText))
	} else {
		h.ctx.ReportNode(reportNode, buildPreferStartsWithMessage())
	}
}

func (h *ruleHelper) reportEndsWith(reportNode *ast.Node, objText, optChain, rightText string, op ast.Kind, canFix bool) {
	fixText := h.buildFixText(objText, optChain, "endsWith", rightText, op)
	if canFix {
		nodeRange := utils.TrimNodeTextRange(h.ctx.SourceFile, reportNode)
		h.ctx.ReportNodeWithFixes(reportNode, buildPreferEndsWithMessage(),
			rule.RuleFixReplaceRange(nodeRange, fixText))
	} else {
		h.ctx.ReportNode(reportNode, buildPreferEndsWithMessage())
	}
}
