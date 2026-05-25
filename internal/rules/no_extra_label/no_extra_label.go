package no_extra_label

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// scopeInfo tracks nested breakable / labeled scopes as a linked-list stack.
// A scope is pushed for every breakable statement (loops + switch) and for
// every LabeledStatement whose body is NOT itself a breakable statement
// (labeled breakables fold their label into the breakable's scope — matching
// ESLint's enterBreakableStatement / enterLabeledStatement handling).
type scopeInfo struct {
	label     string // "" when the breakable has no enclosing label
	breakable bool
	upper     *scopeInfo
}

// hasBreakableBody mirrors astUtils.isBreakableStatement in ESLint:
// iteration statements and switch statements are "breakable".
func hasBreakableBody(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}
	if stmt.Kind == ast.KindSwitchStatement {
		return true
	}
	return ast.IsIterationStatement(stmt, false)
}

// https://eslint.org/docs/latest/rules/no-extra-label
var NoExtraLabelRule = rule.Rule{
	Name: "no-extra-label",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var scope *scopeInfo

		enterBreakable := func(node *ast.Node) {
			label := ""
			if parent := node.Parent; parent != nil && parent.Kind == ast.KindLabeledStatement {
				label = parent.AsLabeledStatement().Label.Text()
			}
			scope = &scopeInfo{
				label:     label,
				breakable: true,
				upper:     scope,
			}
		}

		exitBreakable := func(node *ast.Node) {
			if scope != nil {
				scope = scope.upper
			}
		}

		enterLabeled := func(node *ast.Node) {
			ls := node.AsLabeledStatement()
			if hasBreakableBody(ls.Statement) {
				return
			}
			scope = &scopeInfo{
				label:     ls.Label.Text(),
				breakable: false,
				upper:     scope,
			}
		}

		exitLabeled := func(node *ast.Node) {
			ls := node.AsLabeledStatement()
			if hasBreakableBody(ls.Statement) {
				return
			}
			if scope != nil {
				scope = scope.upper
			}
		}

		reportIfUnnecessary := func(node *ast.Node, labelNode *ast.Node) {
			if labelNode == nil {
				return
			}
			targetName := labelNode.Text()

			for info := scope; info != nil; info = info.upper {
				if info.breakable || info.label == targetName {
					if info.breakable && info.label == targetName {
						msg := rule.RuleMessage{
							Id:          "unexpected",
							Description: "This label '" + targetName + "' is unnecessary.",
						}

						// End of the `break` / `continue` keyword (scanner
						// skips leading trivia on node.Pos()).
						firstTokenRange := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, node.Pos())
						keywordEnd := firstTokenRange.End()

						// First non-trivia position of the label identifier.
						sourceText := ctx.SourceFile.Text()
						labelTrimmedStart := scanner.SkipTrivia(sourceText, labelNode.Pos())

						// Suppress the autofix if there is a comment between
						// the keyword and the label (matching ESLint's
						// commentsExistBetween guard).
						if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(keywordEnd, labelTrimmedStart)) {
							ctx.ReportNode(labelNode, msg)
						} else {
							fix := rule.RuleFixRemoveRange(core.NewTextRange(keywordEnd, labelNode.End()))
							ctx.ReportNodeWithFixes(labelNode, msg, fix)
						}
					}
					return
				}
			}
		}

		return rule.RuleListeners{
			ast.KindWhileStatement:                        enterBreakable,
			rule.ListenerOnExit(ast.KindWhileStatement):   exitBreakable,
			ast.KindDoStatement:                           enterBreakable,
			rule.ListenerOnExit(ast.KindDoStatement):      exitBreakable,
			ast.KindForStatement:                          enterBreakable,
			rule.ListenerOnExit(ast.KindForStatement):     exitBreakable,
			ast.KindForInStatement:                        enterBreakable,
			rule.ListenerOnExit(ast.KindForInStatement):   exitBreakable,
			ast.KindForOfStatement:                        enterBreakable,
			rule.ListenerOnExit(ast.KindForOfStatement):   exitBreakable,
			ast.KindSwitchStatement:                       enterBreakable,
			rule.ListenerOnExit(ast.KindSwitchStatement):  exitBreakable,
			ast.KindLabeledStatement:                      enterLabeled,
			rule.ListenerOnExit(ast.KindLabeledStatement): exitLabeled,
			ast.KindBreakStatement: func(node *ast.Node) {
				bs := node.AsBreakStatement()
				if bs.Label == nil {
					return
				}
				reportIfUnnecessary(node, bs.Label)
			},
			ast.KindContinueStatement: func(node *ast.Node) {
				cs := node.AsContinueStatement()
				if cs.Label == nil {
					return
				}
				reportIfUnnecessary(node, cs.Label)
			},
		}
	},
}
