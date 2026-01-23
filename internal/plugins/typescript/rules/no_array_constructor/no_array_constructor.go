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
		// Returns empty string if no parentheses found (e.g., "new Array;")
		getArgumentsText := func(node *ast.Node) string {
			text := ctx.SourceFile.Text()

			// Get callee end position to start scanning from
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

			// Start scanning from after callee or after type arguments if present
			scanStart := callee.End()
			if typeArgs != nil && len(typeArgs.Nodes) > 0 {
				scanStart = typeArgs.End()
			}

			// Use scanner to find the opening paren
			s := scanner.GetScannerForSourceFile(ctx.SourceFile, scanStart)
			openParenEnd := -1
			closeParenStart := -1

			for s.TokenStart() < node.End() {
				if s.Token() == ast.KindOpenParenToken {
					openParenEnd = s.TokenEnd()
				} else if s.Token() == ast.KindCloseParenToken {
					closeParenStart = s.TokenStart()
					break
				}
				s.Scan()
			}

			// No parentheses found (e.g., "new Array;")
			if openParenEnd == -1 || closeParenStart == -1 {
				return ""
			}

			// Return the text between open and close parens
			return text[openParenEnd:closeParenStart]
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
