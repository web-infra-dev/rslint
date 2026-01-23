package no_array_constructor

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func useLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteral",
		Description: "The array literal notation [] is preferable.",
	}
}

var NoArrayConstructorRule = rule.CreateRule(rule.Rule{
	Name: "no-array-constructor",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// getArgumentsText extracts the text between opening and closing parentheses
		getArgumentsText := func(node *ast.Node) string {
			text := ctx.SourceFile.Text()
			nodeEnd := node.End()

			// Find the last token (should be closing paren)
			lastTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, nodeEnd-1)
			if lastTokenRange.Pos() >= lastTokenRange.End() {
				return ""
			}
			lastToken := text[lastTokenRange.Pos():lastTokenRange.End()]
			if lastToken != ")" {
				return ""
			}

			// Find the opening paren - search forward from callee end
			var callee *ast.Node
			var typeArgs *ast.NodeList
			switch node.Kind {
			case ast.KindCallExpression:
				callExpr := node.AsCallExpression()
				callee = callExpr.Expression
				typeArgs = callExpr.TypeArguments
			case ast.KindNewExpression:
				newExpr := node.AsNewExpression()
				callee = newExpr.Expression
				typeArgs = newExpr.TypeArguments
			default:
				return ""
			}

			// Start searching from after callee or after type arguments if present
			searchStart := callee.End()
			if typeArgs != nil && len(typeArgs.Nodes) > 0 {
				searchStart = typeArgs.End()
			}

			// Find opening paren
			openParenPos := -1
			for i := searchStart; i < nodeEnd; i++ {
				tokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, i)
				if tokenRange.Pos() >= tokenRange.End() {
					continue
				}
				token := text[tokenRange.Pos():tokenRange.End()]
				if token == "(" {
					openParenPos = tokenRange.End()
					break
				}
				// Skip to end of this token
				i = tokenRange.End() - 1
			}

			if openParenPos == -1 {
				return ""
			}

			// Return the text between open and close parens
			return text[openParenPos:lastTokenRange.Pos()]
		}

		check := func(node *ast.Node) {
			var callee *ast.Node
			var args *ast.NodeList
			var typeArgs *ast.NodeList

			switch node.Kind {
			case ast.KindCallExpression:
				callExpr := node.AsCallExpression()
				callee = callExpr.Expression
				args = callExpr.Arguments
				typeArgs = callExpr.TypeArguments
			case ast.KindNewExpression:
				newExpr := node.AsNewExpression()
				callee = newExpr.Expression
				args = newExpr.Arguments
				typeArgs = newExpr.TypeArguments
			default:
				return
			}

			// Check if callee is an Identifier named "Array"
			if callee == nil || callee.Kind != ast.KindIdentifier {
				return
			}
			identifier := callee.AsIdentifier()
			if identifier.Text != "Array" {
				return
			}

			// Skip if there are type arguments (e.g., Array<Foo>())
			if typeArgs != nil && len(typeArgs.Nodes) > 0 {
				return
			}

			// Skip if there's exactly 1 argument (e.g., Array(5))
			argCount := 0
			if args != nil {
				argCount = len(args.Nodes)
			}
			if argCount == 1 {
				return
			}

			// Report with fix
			argsText := getArgumentsText(node)
			replacement := "[" + argsText + "]"

			ctx.ReportNodeWithFixes(
				node,
				useLiteralMessage(),
				rule.RuleFixReplace(ctx.SourceFile, node, replacement),
			)
		}

		return rule.RuleListeners{
			ast.KindCallExpression: check,
			ast.KindNewExpression:  check,
		}
	},
})
