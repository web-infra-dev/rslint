package no_path_concat

import (
	"cmp"
	"os"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-path-concat.js
var NoPathConcatRule = rule.Rule{
	Name:   "node/no-path-concat",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		type report struct {
			node    *ast.Node
			message rule.RuleMessage
		}
		var reports []report
		var globals [2]struct {
			references []*ast.Node
			modified   bool
		}
		var separators map[*ast.Node]bool
		var evaluator *utils.StaticStringEvaluator
		var startsWithSeparator func(*ast.Node) bool

		templateElementStartsWithSeparator := func(literal, nextExpression *ast.Node) bool {
			// Tagged templates can have a null cooked value after an invalid escape.
			if literal.TemplateLiteralLikeData().TemplateFlags&ast.TokenFlagsContainsInvalidEscape != 0 {
				return false
			}
			if text := literal.Text(); text != "" {
				return isPathSeparator(text)
			}
			return startsWithSeparator(nextExpression)
		}
		startsWithSeparator = func(node *ast.Node) bool {
			node = utils.ESTreeRuntimeExpression(node)
			if node == nil {
				return false
			}
			switch node.Kind {
			case ast.KindBinaryExpression:
				binary := node.AsBinaryExpression()
				switch binary.OperatorToken.Kind {
				case ast.KindCommaToken:
					return startsWithSeparator(binary.Right)
				case ast.KindBarBarToken, ast.KindAmpersandAmpersandToken, ast.KindQuestionQuestionToken:
					return startsWithSeparator(binary.Left) || startsWithSeparator(binary.Right)
				default:
					if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
						return startsWithSeparator(binary.Right)
					}
					return startsWithSeparator(binary.Left)
				}
			case ast.KindConditionalExpression:
				conditional := node.AsConditionalExpression()
				return startsWithSeparator(conditional.WhenTrue) || startsWithSeparator(conditional.WhenFalse)
			case ast.KindTemplateExpression:
				template := node.AsTemplateExpression()
				return templateElementStartsWithSeparator(template.Head, template.TemplateSpans.Nodes[0].AsTemplateSpan().Expression)
			case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				if separators == nil {
					separators = nodeutil.CollectModulePropertyReads(ctx, "path", "sep")
				}
				// An optional member is wrapped in ESTree's ChainExpression,
				// so it does not match a tracked MemberExpression here.
				if !ast.IsOptionalChain(node) && separators[node] {
					return true
				}
			}
			if evaluator == nil {
				evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(nil, ctx.SourceFile, ctx.Refs)
			}
			// Reuse the shared static evaluator; uncommon calls such as
			// String.fromCodePoint remain unknown (see Differences from upstream).
			text, ok := evaluator.EvalToString(node)
			return ok && isPathSeparator(text)
		}

		check := func(node *ast.Node, message rule.RuleMessage) {
			parent := utils.ESTreeParent(node)
			if parent == nil {
				return
			}
			if parent.Kind == ast.KindBinaryExpression {
				binary := parent.AsBinaryExpression()
				if binary.OperatorToken.Kind == ast.KindPlusToken && utils.ESTreeRuntimeExpression(binary.Left) == node && startsWithSeparator(binary.Right) {
					reports = append(reports, report{parent, message})
				}
			} else if parent.Kind == ast.KindTemplateSpan {
				span := parent.AsTemplateSpan()
				if utils.ESTreeRuntimeExpression(span.Expression) != node {
					return
				}
				template := parent.Parent
				var nextExpression *ast.Node
				if span.Literal.Text() == "" {
					spans := template.AsTemplateExpression().TemplateSpans.Nodes
					for i, current := range spans {
						if current == parent && i+1 < len(spans) {
							nextExpression = spans[i+1].AsTemplateSpan().Expression
							break
						}
					}
				}
				if templateElementStartsWithSeparator(span.Literal, nextExpression) {
					reports = append(reports, report{template, message})
				}
			}
		}
		pathMessage := rule.RuleMessage{Id: "usePathFunctions", Description: "Use path.join() or path.resolve() instead of string concatenation."}
		checkImportMeta := func(node *ast.Node) {
			if ast.IsOptionalChain(node) || !ast.IsImportMeta(utils.ESTreeRuntimeExpression(utils.AccessExpressionObject(node))) {
				return
			}
			var property string
			if node.Kind == ast.KindPropertyAccessExpression {
				name := node.Name()
				if name.Kind == ast.KindIdentifier {
					property = name.Text()
				}
			} else {
				key := utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
				if ast.IsStringLiteralLike(key) {
					property = key.Text()
				}
			}
			switch property {
			case "dirname", "filename":
				check(node, pathMessage)
			case "url":
				check(node, rule.RuleMessage{Id: "useUrl", Description: "Use new URL() instead of string concatenation."})
			}
		}
		listeners := rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				index := 0
				switch node.Text() {
				case "__dirname":
				case "__filename":
					index = 1
				default:
					return
				}
				global := &globals[index]
				if global.modified || !ctx.Globals.Access(node.Text()).IsDeclared() || !ctx.Refs.IsGlobalReference(node) {
					return
				}
				if utils.IsWriteReference(node) {
					global.modified = true
					global.references = nil
					return
				}
				global.references = append(global.references, node)
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
				for _, global := range globals {
					for _, node := range global.references {
						check(node, pathMessage)
					}
				}
				// Global references are checked on exit; expose all diagnostics in
				// source order, preserving duplicate reports on the same template.
				slices.SortStableFunc(reports, func(a, b report) int { return cmp.Compare(a.node.Pos(), b.node.Pos()) })
				for _, report := range reports {
					ctx.ReportNode(report.node, report.message)
				}
			},
		}
		if ctx.SourceFile.Flags&ast.NodeFlagsPossiblyContainsImportMeta != 0 {
			listeners[ast.KindPropertyAccessExpression] = checkImportMeta
			listeners[ast.KindElementAccessExpression] = checkImportMeta
		}
		return listeners
	},
}

func isPathSeparator(text string) bool {
	return text != "" && (text[0] == '/' || text[0] == os.PathSeparator)
}
