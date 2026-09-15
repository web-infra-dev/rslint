package warn_todo

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildWarnTodoMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "warnTodo",
		Description: "The use of `.todo` is not recommended.",
	}
}

var WarnTodoRule = rule.Rule{
	Name:   "rstest/warn-todo",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil || !parsed.Todo {
					return
				}
				// Hooks take no modifiers, so `parsed.Todo` cannot be set for
				// one; the kind check states the contract anyway, so a later
				// parser change cannot silently widen this rule.
				if parsed.Kind != rstestUtils.RstestFnTypeTest &&
					parsed.Kind != rstestUtils.RstestFnTypeDescribe {
					return
				}

				ctx.ReportNode(todoReportNode(node, parsed), buildWarnTodoMessage())
			},
		}
	},
}

// todoReportNode picks the node that best explains the diagnostic. `.todo`
// written at the call site is the accessor the reader can delete, so it wins.
// When `.todo` arrived through an alias (`const t = test.todo; t('x')`) the
// call site has no such accessor — its own members are whatever was chained
// onto the alias — and the identifier that resolves to the todo registration is
// the honest anchor instead.
func todoReportNode(node *ast.Node, parsed *rstestUtils.ParsedRstestFnCall) *ast.Node {
	for _, entry := range parsed.MemberEntries {
		if entry.Name == "todo" && entry.Node != nil {
			return entry.Node
		}
	}
	if parsed.Head.Local.Node != nil {
		return parsed.Head.Local.Node
	}
	return node
}
