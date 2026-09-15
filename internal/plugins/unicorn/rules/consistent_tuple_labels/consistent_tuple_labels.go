package consistent_tuple_labels

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const messageID = "consistent-tuple-labels"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "This tuple element should have a label, just like the other elements.",
}

func isLabeledElement(element *ast.Node) bool {
	// tsgo represents every labeled tuple element as NamedTupleMember,
	// including optional and rest elements. Unlabeled optional and rest
	// elements instead use OptionalType and RestType nodes.
	return element != nil && element.Kind == ast.KindNamedTupleMember
}

// ConsistentTupleLabelsRule requires every element of a tuple type to have a
// label when any element is labeled.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/consistent-tuple-labels.js
var ConsistentTupleLabelsRule = rule.Rule{
	Name:   "unicorn/consistent-tuple-labels",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindTupleType: func(node *ast.Node) {
				tuple := node.AsTupleTypeNode()
				if tuple == nil || tuple.Elements == nil || len(tuple.Elements.Nodes) < 2 {
					return
				}

				labeledCount := 0
				for _, element := range tuple.Elements.Nodes {
					if isLabeledElement(element) {
						labeledCount++
					}
				}

				if labeledCount == 0 || labeledCount == len(tuple.Elements.Nodes) {
					return
				}

				for _, element := range tuple.Elements.Nodes {
					if !isLabeledElement(element) {
						// TSESTree omits type parentheses from the tuple element node's
						// range. Keep the original node for label classification, but
						// report the innermost type so diagnostic ranges match upstream.
						ctx.ReportNode(ast.SkipTypeParentheses(element), message)
					}
				}
			},
		}
	},
}
