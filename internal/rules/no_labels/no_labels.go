package no_labels

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// labelScope tracks nested labeled statements as a linked list (stack).
// Each entry records the label name and the kind of its body ("loop", "switch", or "other"),
// used to determine whether the label is allowed by the allowLoop/allowSwitch options.
type labelScope struct {
	label string
	kind  string // "loop", "switch", or "other"
	upper *labelScope
}

// getBodyKind categorizes a labeled statement's body for option matching.
func getBodyKind(node *ast.Node) string {
	switch {
	case ast.IsIterationStatement(node, false):
		return "loop"
	case node.Kind == ast.KindSwitchStatement:
		return "switch"
	default:
		return "other"
	}
}

// https://eslint.org/docs/latest/rules/no-labels
var NoLabelsRule = rule.Rule{
	Name: "no-labels",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		allowLoop := false
		allowSwitch := false
		optsMap := utils.GetOptionsMap(options)
		if optsMap != nil {
			if v, ok := optsMap["allowLoop"].(bool); ok {
				allowLoop = v
			}
			if v, ok := optsMap["allowSwitch"].(bool); ok {
				allowSwitch = v
			}
		}

		isAllowed := func(kind string) bool {
			switch kind {
			case "loop":
				return allowLoop
			case "switch":
				return allowSwitch
			default:
				return false
			}
		}

		var scopeInfo *labelScope

		getKind := func(label string) string {
			info := scopeInfo
			for info != nil {
				if info.label == label {
					return info.kind
				}
				info = info.upper
			}
			return "other"
		}

		return rule.RuleListeners{
			ast.KindLabeledStatement: func(node *ast.Node) {
				ls := node.AsLabeledStatement()
				scopeInfo = &labelScope{
					label: ls.Label.Text(),
					kind:  getBodyKind(ls.Statement),
					upper: scopeInfo,
				}
			},
			rule.ListenerOnExit(ast.KindLabeledStatement): func(node *ast.Node) {
				if !isAllowed(scopeInfo.kind) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedLabel",
						Description: "Unexpected labeled statement.",
					})
				}
				scopeInfo = scopeInfo.upper
			},
			ast.KindBreakStatement: func(node *ast.Node) {
				bs := node.AsBreakStatement()
				if bs.Label != nil && !isAllowed(getKind(bs.Label.Text())) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedLabelInBreak",
						Description: "Unexpected label in break statement.",
					})
				}
			},
			ast.KindContinueStatement: func(node *ast.Node) {
				cs := node.AsContinueStatement()
				if cs.Label != nil && !isAllowed(getKind(cs.Label.Text())) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unexpectedLabelInContinue",
						Description: "Unexpected label in continue statement.",
					})
				}
			},
		}
	},
}
