package arrow_body_style

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// arrow-body-style enforces or disallows braces around arrow function bodies.
// https://eslint.org/docs/latest/rules/arrow-body-style

const (
	styleAlways   = "always"
	styleAsNeeded = "as-needed"
	styleNever    = "never"
)

type options struct {
	style                         string
	requireReturnForObjectLiteral bool
}

func parseOptions(opts any) options {
	o := options{style: styleAsNeeded}
	if arr, ok := opts.([]interface{}); ok {
		if len(arr) > 0 {
			if s, ok := arr[0].(string); ok && s != "" {
				o.style = s
			}
		}
		if len(arr) > 1 {
			if m, ok := arr[1].(map[string]interface{}); ok {
				if v, ok := m["requireReturnForObjectLiteral"].(bool); ok {
					o.requireReturnForObjectLiteral = v
				}
			}
		}
	} else if s, ok := opts.(string); ok && s != "" {
		o.style = s
	}
	return o
}

func msgExpectedBlock() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectedBlock",
		Description: "Expected block statement surrounding arrow body.",
	}
}

func msgUnexpectedOtherBlock() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedOtherBlock",
		Description: "Unexpected block statement surrounding arrow body.",
	}
}

func msgUnexpectedEmptyBlock() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedEmptyBlock",
		Description: "Unexpected block statement surrounding arrow body; put a value of `undefined` immediately after the `=>`.",
	}
}

func msgUnexpectedObjectBlock() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedObjectBlock",
		Description: "Unexpected block statement surrounding arrow body; parenthesize the returned value and move it immediately after the `=>`.",
	}
}

func msgUnexpectedSingleBlock() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedSingleBlock",
		Description: "Unexpected block statement surrounding arrow body; move the returned value immediately after the `=>`.",
	}
}

// isInsideForLoopInitializer reports whether node sits inside the init clause of
// an enclosing `for` statement. Mirrors upstream's recursive parent walk.
func isInsideForLoopInitializer(node *ast.Node) bool {
	return ast.FindAncestor(node, func(n *ast.Node) bool {
		p := n.Parent
		return p != nil && p.Kind == ast.KindForStatement && p.AsForStatement().Initializer == n
	}) != nil
}

// subtreeHasInOperator reports whether node's subtree contains an `in`
// BinaryExpression. Upstream tracks this with a funcInfo stack that the
// `BinaryExpression[operator='in']` listener flags up the ancestor chain; an
// arrow's flag ends up true exactly when its subtree holds an `in`, so a direct
// subtree walk is equivalent.
func subtreeHasInOperator(node *ast.Node) bool {
	found := false
	var visit func(n *ast.Node)
	visit = func(n *ast.Node) {
		if found || n == nil {
			return
		}
		if n.Kind == ast.KindBinaryExpression {
			if bin := n.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindInKeyword {
				found = true
				return
			}
		}
		n.ForEachChild(func(c *ast.Node) bool {
			visit(c)
			return found
		})
	}
	visit(node)
	return found
}

// hasASIProblemAfter reports whether the first token at/after pos would fuse
// with an unbraced body once the braces are removed. Upstream guards punctuators
// starting with one of `([/+-`; backticks are template tokens (not punctuators),
// so they are excluded.
//
// NOTE: Unlike ESLint, we keep `/` as a hazard. ESLint tokenizes a `/` after the
// arrow block `}` as a regex literal and applies the fix; tsgo tokenizes the same
// `/` as division, so the de-braced expression body (`() => x\n/y/.test(z)`) is a
// syntax error under tsgo. Suppressing the fix keeps the output parseable.
func hasASIProblemAfter(text string, pos int) bool {
	p := scanner.SkipTrivia(text, pos)
	if p < 0 || p >= len(text) {
		return false
	}
	switch text[p] {
	case '[', '(', '+', '-', '/':
		return true
	}
	return false
}

// findClosingParenRange walks up from a parenthesized node to the nearest
// enclosing ParenthesizedExpression and returns its closing `)` range. Mirrors
// upstream's findClosingParen.
func findClosingParenRange(objNode *ast.Node, sf *ast.SourceFile) (int, int, bool) {
	wrapped := ast.FindAncestor(objNode, func(n *ast.Node) bool {
		return n.Parent != nil && n.Parent.Kind == ast.KindParenthesizedExpression
	})
	if wrapped == nil {
		return 0, 0, false
	}
	parenEnd := utils.TrimNodeTextRange(sf, wrapped.Parent).End()
	return parenEnd - 1, parenEnd, true
}

var ArrowBodyStyleRule = rule.Rule{
	Name: "arrow-body-style",
	Run: func(ctx rule.RuleContext, _opts []any) rule.RuleListeners {
		opts := rule.LegacyUnwrapOptions(_opts)
		o := parseOptions(opts)
		always := o.style == styleAlways
		asNeeded := o.style == styleAsNeeded
		never := o.style == styleNever
		sf := ctx.SourceFile
		text := sf.Text()

		sameLine := func(a, b int) bool {
			lineStarts := scanner.GetECMALineStarts(sf)
			return scanner.ComputeLineOfPosition(lineStarts, a) == scanner.ComputeLineOfPosition(lineStarts, b)
		}
		insertAt := func(pos int, s string) rule.RuleFix {
			return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), s)
		}
		removeR := func(start, end int) rule.RuleFix {
			return rule.RuleFixRemoveRange(core.NewTextRange(start, end))
		}

		validate := func(node *ast.Node) {
			arrow := node.AsArrowFunction()
			if arrow == nil || arrow.Body == nil {
				return
			}
			body := arrow.Body

			if body.Kind == ast.KindBlock {
				block := body.AsBlock()
				var stmts []*ast.Node
				if block != nil && block.Statements != nil {
					stmts = block.Statements.Nodes
				}
				blockRange := utils.TrimNodeTextRange(sf, body)

				if len(stmts) != 1 && !never {
					return
				}

				var firstStmt *ast.Node
				if len(stmts) >= 1 {
					firstStmt = stmts[0]
				}

				// as-needed + requireReturnForObjectLiteral: a `return {obj}`
				// keeps its braces, so do not report it.
				if asNeeded && o.requireReturnForObjectLiteral && firstStmt != nil && firstStmt.Kind == ast.KindReturnStatement {
					if ret := firstStmt.AsReturnStatement(); ret.Expression != nil &&
						ast.SkipParentheses(ret.Expression).Kind == ast.KindObjectLiteralExpression {
						return
					}
				}

				blockReportable := never || (asNeeded && firstStmt != nil && firstStmt.Kind == ast.KindReturnStatement)
				if !blockReportable {
					return
				}

				var retExpr *ast.Node
				if firstStmt != nil && firstStmt.Kind == ast.KindReturnStatement {
					retExpr = firstStmt.AsReturnStatement().Expression
				}

				// Determine the messageId, mirroring upstream's branch order.
				var msg rule.RuleMessage
				switch {
				case len(stmts) == 0:
					msg = msgUnexpectedEmptyBlock()
				case len(stmts) > 1 || firstStmt.Kind != ast.KindReturnStatement:
					msg = msgUnexpectedOtherBlock()
				case retExpr == nil:
					msg = msgUnexpectedSingleBlock()
				default:
					fvs := scanner.SkipTrivia(text, retExpr.Pos())
					if fvs < len(text) && text[fvs] == '{' {
						msg = msgUnexpectedObjectBlock()
					} else {
						msg = msgUnexpectedSingleBlock()
					}
				}

				// Build the autofix. It stays empty when upstream would emit no
				// fix (more than one statement, no/empty return argument, or an
				// ASI hazard on the token after the block).
				var fixes []rule.RuleFix
				if len(stmts) == 1 && firstStmt.Kind == ast.KindReturnStatement && retExpr != nil && !hasASIProblemAfter(text, body.End()) {
					argument := ast.SkipParentheses(retExpr)

					braceOpenStart := blockRange.Pos()
					braceCloseEnd := blockRange.End()
					braceCloseStart := braceCloseEnd - 1
					firstValueStart := scanner.SkipTrivia(text, retExpr.Pos())
					valueEnd := utils.TrimNodeTextRange(sf, retExpr).End()

					semiStart, semiEnd := -1, -1
					if p := scanner.SkipTrivia(text, valueEnd); p < braceCloseStart && p < len(text) && text[p] == ';' {
						semiStart, semiEnd = p, p+1
					}
					lastValueEnd := valueEnd
					if semiStart >= 0 {
						lastValueEnd = semiEnd
					}

					// commentsExistBetween(openingBrace, firstValueToken) ||
					// commentsExistBetween(lastValueToken, closingBrace). ForEachComment
					// walks every token's trivia, so it catches a comment sitting after
					// the `return` keyword too (which a position-anchored scan misses).
					commentsExist := false
					utils.ForEachComment(body, func(c *ast.CommentRange) {
						if commentsExist {
							return
						}
						if (c.Pos() >= braceOpenStart+1 && c.End() <= firstValueStart) ||
							(c.Pos() >= lastValueEnd && c.End() <= braceCloseStart) {
							commentsExist = true
						}
					}, sf)

					if commentsExist {
						returnKwStart := scanner.SkipTrivia(text, braceOpenStart+1)
						fixes = append(fixes,
							removeR(braceOpenStart, braceOpenStart+1),
							removeR(braceCloseStart, braceCloseEnd),
							removeR(returnKwStart, returnKwStart+len("return")),
						)
					} else {
						fixes = append(fixes,
							removeR(braceOpenStart, firstValueStart),
							removeR(lastValueEnd, braceCloseEnd),
						)
					}

					isSeq := argument.Kind == ast.KindBinaryExpression &&
						argument.AsBinaryExpression() != nil &&
						argument.AsBinaryExpression().OperatorToken != nil &&
						argument.AsBinaryExpression().OperatorToken.Kind == ast.KindCommaToken
					startsWithBrace := firstValueStart < len(text) && text[firstValueStart] == '{'
					needsParens := startsWithBrace || isSeq ||
						(isInsideForLoopInitializer(node) && subtreeHasInOperator(node))
					if needsParens && retExpr.Kind != ast.KindParenthesizedExpression {
						fixes = append(fixes,
							insertAt(firstValueStart, "("),
							insertAt(lastValueEnd, ")"),
						)
					}

					if semiStart >= 0 {
						fixes = append(fixes, removeR(semiStart, semiEnd))
					}
				}

				if len(fixes) > 0 {
					ctx.ReportRangeWithFixes(blockRange, msg, fixes...)
				} else {
					ctx.ReportRange(blockRange, msg)
				}
				return
			}

			// Expression body.
			argument := ast.SkipParentheses(body)
			exprReportable := always || (asNeeded && o.requireReturnForObjectLiteral && argument.Kind == ast.KindObjectLiteralExpression)
			if !exprReportable {
				return
			}
			reportRange := utils.TrimNodeTextRange(sf, argument)

			arrowToken := arrow.EqualsGreaterThanToken
			if arrowToken == nil {
				ctx.ReportRange(reportRange, msgExpectedBlock())
				return
			}
			firstTokenStart := scanner.SkipTrivia(text, arrowToken.End())
			nodeEnd := utils.TrimNodeTextRange(sf, node).End()

			var fixes []rule.RuleFix
			handled := false
			if firstTokenStart < len(text) && text[firstTokenStart] == '(' {
				braceStart := scanner.SkipTrivia(text, firstTokenStart+1)
				if braceStart < len(text) && text[braceStart] == '{' {
					objNode := ast.GetNodeAtPosition(sf, braceStart, false)
					// A `({a} = b)`-style destructuring target parses as an
					// ObjectLiteralExpression in tsgo but is an ObjectPattern in
					// ESTree, so exclude assignment targets to match ESLint (which
					// keeps the forced parens via its else branch).
					if objNode != nil && objNode.Kind == ast.KindObjectLiteralExpression && !ast.IsAssignmentTarget(objNode) {
						if cpStart, cpEnd, ok := findClosingParenRange(objNode, sf); ok {
							openParenStart := firstTokenStart
							openParenEnd := firstTokenStart + 1
							if sameLine(openParenStart, braceStart) {
								fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(openParenStart, openParenEnd), "{return "))
							} else {
								// Different lines: keep the `return ` next to the
								// object so ASI does not split the statement.
								fixes = append(fixes,
									rule.RuleFixReplaceRange(core.NewTextRange(openParenStart, openParenEnd), "{"),
									insertAt(braceStart, "return "),
								)
							}
							fixes = append(fixes, removeR(cpStart, cpEnd), insertAt(nodeEnd, "}"))
							handled = true
						}
					}
				}
			}
			if !handled {
				fixes = append(fixes,
					insertAt(firstTokenStart, "{return "),
					insertAt(nodeEnd, "}"),
				)
			}

			ctx.ReportRangeWithFixes(reportRange, msgExpectedBlock(), fixes...)
		}

		return rule.RuleListeners{
			ast.KindArrowFunction: validate,
		}
	},
}
