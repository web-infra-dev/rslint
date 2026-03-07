package jsx_wrap_multilines

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Default option values for each context.
var defaultOptions = map[string]string{
	"declaration": "parens",
	"assignment":  "parens",
	"return":      "parens",
	"arrow":       "parens",
	"condition":   "ignore",
	"logical":     "ignore",
	"prop":        "ignore",
}

// JsxWrapMultilinesRule enforces parentheses around multiline JSX.
var JsxWrapMultilinesRule = rule.Rule{
	Name: "react/jsx-wrap-multilines",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Build effective options from defaults
		opts := make(map[string]string)
		for k, v := range defaultOptions {
			opts[k] = v
		}

		optsMap := utils.GetOptionsMap(options)
		if optsMap != nil {
			for key := range defaultOptions {
				if val, ok := optsMap[key]; ok {
					switch v := val.(type) {
					case string:
						opts[key] = v
					case bool:
						if v {
							opts[key] = "parens"
						} else {
							opts[key] = "ignore"
						}
					}
				}
			}
		}

		lineStarts := ctx.SourceFile.ECMALineMap()

		isMultiline := func(node *ast.Node) bool {
			trimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)
			startLine := scanner.ComputeLineOfPosition(lineStarts, trimmed.Pos())
			endLine := scanner.ComputeLineOfPosition(lineStarts, trimmed.End())
			return startLine != endLine
		}

		isJSX := func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
				return true
			}
			return false
		}

		// Unwrap parenthesized expression to get the inner JSX node
		unwrapParens := func(node *ast.Node) *ast.Node {
			n := node
			for n != nil && n.Kind == ast.KindParenthesizedExpression {
				n = n.AsParenthesizedExpression().Expression
			}
			return n
		}

		isWrappedInParens := func(node *ast.Node) bool {
			return node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression
		}

		isParensOnSeparateLines := func(node *ast.Node) bool {
			if node.Parent == nil || node.Parent.Kind != ast.KindParenthesizedExpression {
				return false
			}
			paren := node.Parent
			parenTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, paren)
			innerTrimmed := utils.TrimNodeTextRange(ctx.SourceFile, node)

			// Check that opening paren is followed by newline
			parenStartLine := scanner.ComputeLineOfPosition(lineStarts, parenTrimmed.Pos())
			innerStartLine := scanner.ComputeLineOfPosition(lineStarts, innerTrimmed.Pos())
			if parenStartLine == innerStartLine {
				return false
			}

			// Check that closing paren is preceded by newline
			innerEndLine := scanner.ComputeLineOfPosition(lineStarts, innerTrimmed.End())
			parenEndLine := scanner.ComputeLineOfPosition(lineStarts, parenTrimmed.End())
			return innerEndLine != parenEndLine
		}

		checkJSX := func(jsxNode *ast.Node, setting string) {
			if setting == "ignore" {
				return
			}
			if jsxNode == nil || !isMultiline(jsxNode) {
				return
			}

			wrapped := isWrappedInParens(jsxNode)

			switch setting {
			case "parens":
				if !wrapped {
					ctx.ReportNode(jsxNode, rule.RuleMessage{
						Id:          "missingParens",
						Description: "Missing parentheses around multilines JSX",
					})
				}
			case "parens-new-line":
				if !wrapped {
					ctx.ReportNode(jsxNode, rule.RuleMessage{
						Id:          "missingParens",
						Description: "Missing parentheses around multilines JSX",
					})
				} else if !isParensOnSeparateLines(jsxNode) {
					ctx.ReportNode(jsxNode, rule.RuleMessage{
						Id:          "parensOnNewLines",
						Description: "Parentheses around JSX should be on separate lines",
					})
				}
			case "never":
				if wrapped {
					ctx.ReportNode(jsxNode, rule.RuleMessage{
						Id:          "extraParens",
						Description: "Expected no parentheses around multilines JSX",
					})
				}
			}
		}

		listeners := rule.RuleListeners{}

		// VariableDeclaration → "declaration"
		if opts["declaration"] != "ignore" {
			listeners[ast.KindVariableDeclaration] = func(node *ast.Node) {
				decl := node.AsVariableDeclaration()
				if decl.Initializer == nil {
					return
				}
				init := unwrapParens(decl.Initializer)
				if opts["condition"] == "ignore" && init.Kind == ast.KindConditionalExpression {
					cond := init.AsConditionalExpression()
					consequent := unwrapParens(cond.WhenTrue)
					if isJSX(consequent) {
						checkJSX(consequent, opts["declaration"])
					}
					alternate := unwrapParens(cond.WhenFalse)
					if isJSX(alternate) {
						checkJSX(alternate, opts["declaration"])
					}
					return
				}
				if isJSX(init) {
					checkJSX(init, opts["declaration"])
				}
			}
		}

		// BinaryExpression → "assignment" or "logical"
		if opts["assignment"] != "ignore" || opts["logical"] != "ignore" {
			listeners[ast.KindBinaryExpression] = func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				opKind := bin.OperatorToken.Kind

				// Check for logical operators: &&, ||, ??
				if opKind == ast.KindAmpersandAmpersandToken || opKind == ast.KindBarBarToken || opKind == ast.KindQuestionQuestionToken {
					if opts["logical"] != "ignore" {
						rightUnwrapped := unwrapParens(bin.Right)
						if isJSX(rightUnwrapped) {
							checkJSX(rightUnwrapped, opts["logical"])
						}
					}
					return
				}

				// Check for assignment operators
				if ast.IsAssignmentExpression(node, false) {
					if opts["assignment"] != "ignore" {
						rightUnwrapped := unwrapParens(bin.Right)
						if opts["condition"] == "ignore" && rightUnwrapped.Kind == ast.KindConditionalExpression {
							cond := rightUnwrapped.AsConditionalExpression()
							consequent := unwrapParens(cond.WhenTrue)
							if isJSX(consequent) {
								checkJSX(consequent, opts["assignment"])
							}
							alternate := unwrapParens(cond.WhenFalse)
							if isJSX(alternate) {
								checkJSX(alternate, opts["assignment"])
							}
						} else if isJSX(rightUnwrapped) {
							checkJSX(rightUnwrapped, opts["assignment"])
						}
					}
				}
			}
		}

		// ReturnStatement → "return"
		if opts["return"] != "ignore" {
			listeners[ast.KindReturnStatement] = func(node *ast.Node) {
				ret := node.AsReturnStatement()
				if ret.Expression == nil {
					return
				}
				expr := unwrapParens(ret.Expression)
				if isJSX(expr) {
					checkJSX(expr, opts["return"])
				}
			}
		}

		// ArrowFunction → "arrow"
		if opts["arrow"] != "ignore" {
			listeners[ast.KindArrowFunction] = func(node *ast.Node) {
				body := node.Body()
				if body == nil {
					return
				}
				// Only check expression bodies (not block bodies)
				if body.Kind == ast.KindBlock {
					return
				}
				expr := unwrapParens(body)
				if isJSX(expr) {
					checkJSX(expr, opts["arrow"])
				}
			}
		}

		// ConditionalExpression → "condition"
		if opts["condition"] != "ignore" {
			listeners[ast.KindConditionalExpression] = func(node *ast.Node) {
				cond := node.AsConditionalExpression()
				consequent := unwrapParens(cond.WhenTrue)
				if isJSX(consequent) {
					checkJSX(consequent, opts["condition"])
				}
				alternate := unwrapParens(cond.WhenFalse)
				if isJSX(alternate) {
					checkJSX(alternate, opts["condition"])
				}
			}
		}

		// JsxExpression in JsxAttribute → "prop"
		if opts["prop"] != "ignore" {
			listeners[ast.KindJsxExpression] = func(node *ast.Node) {
				// Only check if parent is a JsxAttribute
				if node.Parent == nil || !ast.IsJsxAttribute(node.Parent) {
					return
				}
				jsxExpr := node.AsJsxExpression()
				if jsxExpr.Expression == nil {
					return
				}
				expr := unwrapParens(jsxExpr.Expression)
				if isJSX(expr) {
					checkJSX(expr, opts["prop"])
				}
			}
		}

		return listeners
	},
}
