package prefer_includes

import (
	"strconv"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferIncludesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferIncludes",
		Description: "Use 'includes()' method instead.",
	}
}

func buildPreferStringIncludesMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferStringIncludes",
		Description: "Use `String#includes()` method with a string instead.",
	}
}

func needsParentheses(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression,
		ast.KindIdentifier,
		ast.KindPropertyAccessExpression,
		ast.KindElementAccessExpression,
		ast.KindCallExpression,
		ast.KindParenthesizedExpression:
		return false
	default:
		return true
	}
}

func escapeForSingleQuotedString(str string) string {
	var b strings.Builder
	b.WriteByte('\'')
	for _, r := range str {
		switch r {
		case 0:
			b.WriteString(`\0`)
		case '\t':
			b.WriteString(`\t`)
		case '\n':
			b.WriteString(`\n`)
		case '\v':
			b.WriteString(`\v`)
		case '\f':
			b.WriteString(`\f`)
		case '\r':
			b.WriteString(`\r`)
		case '\'':
			b.WriteString(`\'`)
		case '\\':
			b.WriteString(`\\`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('\'')
	return b.String()
}

func isStringLikeFixableType(ctx rule.RuleContext, t *checker.Type) bool {
	if t == nil || ctx.TypeChecker == nil {
		return false
	}
	if !utils.IsUnionType(t) {
		return utils.IsTypeFlagSet(t, checker.TypeFlagsStringLike)
	}
	parts := utils.UnionTypeParts(t)
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if !utils.IsTypeFlagSet(part, checker.TypeFlagsStringLike) {
			return false
		}
	}
	return true
}

func hasOptionalChain(node *ast.Node) bool {
	for n := node; n != nil; n = n.Parent {
		if ast.IsOptionalChain(n) {
			return true
		}
		if n.Kind == ast.KindBinaryExpression || n.Kind == ast.KindExpressionStatement || n.Kind == ast.KindBlock {
			break
		}
	}
	return false
}

func stripParens(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

func extractRegexStaticString(pattern string) (string, bool) {
	if pattern == "" {
		return "", true
	}
	var out strings.Builder
	for i := 0; i < len(pattern); i++ {
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
			case 'd', 'D', 'w', 'W', 's', 'S', 'b', 'B', 'p', 'P', 'k', 'x', 'u':
				return "", false
			case '0':
				out.WriteByte(0)
			case 'n':
				out.WriteByte('\n')
			case 'r':
				out.WriteByte('\r')
			case 't':
				out.WriteByte('\t')
			case 'v':
				out.WriteByte('\v')
			case 'f':
				out.WriteByte('\f')
			default:
				out.WriteByte(next)
			}
			i++
		default:
			out.WriteByte(ch)
		}
	}
	return out.String(), true
}

func parseRegexPattern(pattern, flags string) (string, bool) {
	if flags != "" {
		return "", false
	}
	return extractRegexStaticString(pattern)
}

// PreferIncludesRule checks for indexOf comparisons that can use includes instead.
var PreferIncludesRule = rule.CreateRule(rule.Rule{
	Name: "prefer-includes",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		getNodeText := func(n *ast.Node) string {
			if n == nil {
				return ""
			}
			sf := ast.GetSourceFileOfNode(n)
			if sf == nil {
				return ""
			}
			rng := utils.TrimNodeTextRange(sf, n)
			return sf.Text()[rng.Pos():rng.End()]
		}

		isNumber := func(node *ast.Node, value int) bool {
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindPrefixUnaryExpression:
				pref := node.AsPrefixUnaryExpression()
				if pref != nil && pref.Operator == ast.KindMinusToken {
					if ast.IsNumericLiteral(pref.Operand) {
						v, err := strconv.Atoi(pref.Operand.AsNumericLiteral().Text)
						return err == nil && -v == value
					}
				}
			case ast.KindNumericLiteral:
				v, err := strconv.Atoi(node.AsNumericLiteral().Text)
				return err == nil && v == value
			}
			return false
		}

		isPositiveCheck := func(node *ast.Node) bool {
			if node == nil || node.Kind != ast.KindBinaryExpression {
				return false
			}
			bin := node.AsBinaryExpression()
			switch bin.OperatorToken.Kind {
			case ast.KindExclamationEqualsEqualsToken, ast.KindExclamationEqualsToken, ast.KindGreaterThanToken:
				return isNumber(bin.Right, -1)
			case ast.KindGreaterThanEqualsToken:
				return isNumber(bin.Right, 0)
			}
			return false
		}

		isNegativeCheck := func(node *ast.Node) bool {
			if node == nil || node.Kind != ast.KindBinaryExpression {
				return false
			}
			bin := node.AsBinaryExpression()
			switch bin.OperatorToken.Kind {
			case ast.KindEqualsEqualsEqualsToken, ast.KindEqualsEqualsToken, ast.KindLessThanEqualsToken:
				return isNumber(bin.Right, -1)
			case ast.KindLessThanToken:
				return isNumber(bin.Right, 0)
			}
			return false
		}

		hasSameParameters := func(nodeA, nodeB *ast.Node) bool {
			if nodeA == nil || nodeB == nil || !ast.IsFunctionLike(nodeA) || !ast.IsFunctionLike(nodeB) {
				return false
			}
			paramsA := nodeA.Parameters()
			paramsB := nodeB.Parameters()
			if paramsA == nil || paramsB == nil || len(paramsA) != len(paramsB) {
				return false
			}
			for i := range paramsA {
				if getNodeText(paramsA[i]) != getNodeText(paramsB[i]) {
					return false
				}
			}
			return true
		}

		var parseRegExpObject func(node *ast.Node, seen map[string]bool) (string, bool)
		parseRegExpObject = func(node *ast.Node, seen map[string]bool) (string, bool) {
			if node == nil {
				return "", false
			}
			switch node.Kind {
			case ast.KindRegularExpressionLiteral:
				regex := node.AsRegularExpressionLiteral()
				if regex == nil {
					return "", false
				}
				text := regex.Text
				lastSlash := strings.LastIndex(text, "/")
				if !strings.HasPrefix(text, "/") || lastSlash <= 0 {
					return "", false
				}
				return parseRegexPattern(text[1:lastSlash], text[lastSlash+1:])
			case ast.KindNewExpression:
				newExpr := node.AsNewExpression()
				if newExpr == nil || newExpr.Expression == nil || newExpr.Expression.Kind != ast.KindIdentifier || newExpr.Expression.AsIdentifier().Text != "RegExp" {
					return "", false
				}
				if newExpr.Arguments == nil || len(newExpr.Arguments.Nodes) < 1 {
					return "", false
				}
				first := newExpr.Arguments.Nodes[0]
				if first == nil || first.Kind != ast.KindStringLiteral {
					return "", false
				}
				pattern := first.AsStringLiteral().Text
				flags := ""
				if len(newExpr.Arguments.Nodes) >= 2 {
					second := newExpr.Arguments.Nodes[1]
					if second == nil || second.Kind != ast.KindStringLiteral {
						return "", false
					}
					flags = second.AsStringLiteral().Text
				}
				return parseRegexPattern(pattern, flags)
			case ast.KindIdentifier:
				if ctx.TypeChecker == nil {
					return "", false
				}
				name := node.AsIdentifier().Text
				if seen[name] {
					return "", false
				}
				seen[name] = true
				sym := ctx.TypeChecker.GetSymbolAtLocation(node)
				if sym == nil || sym.Declarations == nil {
					return "", false
				}
				for _, decl := range sym.Declarations {
					if decl == nil || decl.Kind != ast.KindVariableDeclaration {
						continue
					}
					varDecl := decl.AsVariableDeclaration()
					if varDecl == nil || varDecl.Initializer == nil {
						continue
					}
					if value, ok := parseRegExpObject(varDecl.Initializer, seen); ok {
						return value, true
					}
				}
			}
			return "", false
		}

		checkArrayIndexOf := func(prop *ast.Node, allowFix bool, bin *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}
			if prop == nil || prop.Kind != ast.KindPropertyAccessExpression {
				return
			}
			pae := prop.AsPropertyAccessExpression()
			nameNode := pae.Name()
			if pae == nil || nameNode == nil || nameNode.Text() != "indexOf" {
				return
			}
			call := prop.Parent
			if call == nil || call.Kind != ast.KindCallExpression {
				return
			}
			negative := isNegativeCheck(bin)
			if !negative && !isPositiveCheck(bin) {
				return
			}
			name := nameNode
			sym := ctx.TypeChecker.GetSymbolAtLocation(name)
			if sym == nil || sym.Declarations == nil || len(sym.Declarations) == 0 {
				return
			}
			for _, decl := range sym.Declarations {
				typeDecl := decl.Parent
				t := ctx.TypeChecker.GetTypeAtLocation(typeDecl)
				includesSym := checker.Checker_getPropertyOfType(ctx.TypeChecker, t, "includes")
				if includesSym == nil || includesSym.Declarations == nil {
					return
				}
				ok := false
				for _, incDecl := range includesSym.Declarations {
					if hasSameParameters(incDecl, decl) {
						ok = true
						break
					}
				}
				if !ok {
					return
				}
			}
			fixes := []rule.RuleFix{}
			if allowFix {
				if negative {
					fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, call, "!"))
				}
				fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, name, "includes"))
				callRange := utils.TrimNodeTextRange(ctx.SourceFile, call)
				binRange := utils.TrimNodeTextRange(ctx.SourceFile, bin)
				if callRange.End() > binRange.End() {
					fixes = nil
				} else {
					leftRange := utils.TrimNodeTextRange(ctx.SourceFile, bin.AsBinaryExpression().Left)
					removeStart := callRange.End()
					if leftRange.End() > removeStart {
						removeStart = leftRange.End()
					}
					fixes = append(fixes, rule.RuleFixRemoveRange(core.NewTextRange(removeStart, binRange.End())))
				}
			}
			ctx.ReportNodeWithFixes(bin, buildPreferIncludesMessage(), fixes...)
		}

		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				left := stripParens(bin.Left)
				if left == nil || left.Kind != ast.KindCallExpression {
					return
				}
				call := left.AsCallExpression()
				expr := stripParens(call.Expression)
				if expr == nil || expr.Kind != ast.KindPropertyAccessExpression {
					return
				}
				prop := expr
				allowFix := expr.AsPropertyAccessExpression().QuestionDotToken == nil
				if allowFix && hasOptionalChain(call.Expression) {
					allowFix = false
				}
				checkArrayIndexOf(prop, allowFix, node)
			},
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil || call.Expression == nil || call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
					return
				}
				if call.Expression.Kind != ast.KindPropertyAccessExpression {
					return
				}
				access := call.Expression.AsPropertyAccessExpression()
				if access == nil || access.Name() == nil || access.Name().Text() != "test" || access.Expression == nil {
					return
				}
				text, ok := parseRegExpObject(access.Expression, map[string]bool{})
				if !ok {
					return
				}
				arg := call.Arguments.Nodes[0]
				if arg == nil || ctx.TypeChecker == nil {
					return
				}
				argType := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, arg)
				includesSym := checker.Checker_getPropertyOfType(ctx.TypeChecker, argType, "includes")
				if includesSym == nil {
					return
				}

				callRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
				argRange := utils.TrimNodeTextRange(ctx.SourceFile, arg)
				if callRange.Pos() >= argRange.Pos() || argRange.End() > callRange.End() {
					ctx.ReportNode(node, buildPreferStringIncludesMessage())
					return
				}

				fixes := []rule.RuleFix{}
				allowFix := isStringLikeFixableType(ctx, argType) &&
					access.QuestionDotToken == nil &&
					!hasOptionalChain(access.Expression)
				if allowFix {
					fixes = append(fixes,
						rule.RuleFixRemoveRange(core.NewTextRange(callRange.Pos(), argRange.Pos())),
						rule.RuleFixRemoveRange(core.NewTextRange(argRange.End(), callRange.End())),
						rule.RuleFixInsertAfter(arg, ".includes("+escapeForSingleQuotedString(text)+")"),
					)
					if needsParentheses(arg) {
						fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, arg, "("))
						fixes = append(fixes, rule.RuleFixInsertAfter(arg, ")"))
					}
				}

				ctx.ReportNodeWithFixes(node, buildPreferStringIncludesMessage(), fixes...)
			},
		}
	},
})
