package no_class_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Message builder
func buildClassReassignmentMessage(className string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "classReassignment",
		Description: "'" + className + "' is a class.",
	}
}

// checkClassReassignments reports every write-reference to classNode's own
// symbol. RefStore resolution is scope-correct by construction, so a local
// binding that shadows the class name is never returned as a reference here.
//
// name comes from the declaration's name node rather than the symbol: a
// default-exported class is bound to an export symbol named "default",
// while the reported name must be the one written in the source.
func checkClassReassignments(classNode *ast.Node, nameNode *ast.Node, ctx *rule.RuleContext) {
	symbol := classNode.Symbol()
	if symbol == nil {
		return
	}

	var message rule.RuleMessage
	hasMessage := false
	for _, reference := range ctx.Refs.References(symbol) {
		if !utils.IsWriteReference(reference) {
			continue
		}
		if !hasMessage {
			message = buildClassReassignmentMessage(nameNode.Text())
			hasMessage = true
		}
		ctx.ReportNode(reference, message)
	}
}

// NoClassAssignRule disallows reassigning class declarations
var NoClassAssignRule = rule.Rule{
	Name: "no-class-assign",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkClass := func(node *ast.Node) {
			nameNode := node.Name()
			if nameNode == nil {
				return
			}
			checkClassReassignments(node, nameNode, &ctx)
		}
		return rule.RuleListeners{
			ast.KindClassDeclaration: checkClass,
			ast.KindClassExpression:  checkClass,
		}
	},
}
