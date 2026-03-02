package no_ex_assign

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildExAssignMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Do not assign to the exception parameter.",
	}
}

func collectCatchBindingNames(name *ast.Node) []string {
	if name == nil {
		return nil
	}
	if ast.IsIdentifier(name) {
		return []string{name.Text()}
	}
	if ast.IsBindingPattern(name) {
		var names []string
		for _, elem := range name.Elements() {
			if elem == nil || !ast.IsBindingElement(elem) {
				continue
			}
			be := elem.AsBindingElement()
			if be == nil || be.Name() == nil {
				continue
			}
			names = append(names, collectCatchBindingNames(be.Name())...)
		}
		return names
	}
	return nil
}

func checkExpressionForExAssign(expression *ast.Node, namesSet map[string]bool, ctx rule.RuleContext) {
	if expression == nil {
		return
	}

	if ast.IsBinaryExpression(expression) {
		binary := expression.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
			return
		}

		left := binary.Left
		if left == nil {
			return
		}

		if ast.IsIdentifier(left) {
			if namesSet[left.Text()] {
				ctx.ReportNode(expression, buildExAssignMessage())
			}
		} else if ast.IsArrayLiteralExpression(left) {
			for _, elem := range left.Elements() {
				if elem == nil || !ast.IsIdentifier(elem) {
					continue
				}
				if namesSet[elem.Text()] {
					ctx.ReportNode(expression, buildExAssignMessage())
				}
			}
		} else if ast.IsObjectLiteralExpression(left) {
			for _, elem := range left.PropertyList().Nodes {
				if elem == nil || !ast.IsPropertyAssignment(elem) {
					continue
				}
				checkExpressionForExAssign(elem.AsPropertyAssignment().Initializer, namesSet, ctx)
			}
		}
	} else if ast.IsParenthesizedExpression(expression) {
		checkExpressionForExAssign(expression.AsParenthesizedExpression().Expression, namesSet, ctx)
	}
}

var NoExAssignRule = rule.Rule{
	Name: "no-ex-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCatchClause: func(node *ast.Node) {
				if node.AsCatchClause().VariableDeclaration == nil {
					return
				}

				varDecl := node.AsCatchClause().VariableDeclaration.AsVariableDeclaration()
				if varDecl == nil || varDecl.Name() == nil {
					return
				}

				names := collectCatchBindingNames(varDecl.Name())
				if len(names) == 0 {
					return
				}

				namesSet := make(map[string]bool)
				for _, n := range names {
					namesSet[n] = true
				}

				block := node.AsCatchClause().Block.AsBlock()
				if block == nil {
					return
				}

				for _, stmt := range block.Statements.Nodes {
					if stmt == nil {
						continue
					}

					if ast.IsExpressionStatement(stmt) {
						checkExpressionForExAssign(stmt.AsExpressionStatement().Expression, namesSet, ctx)
					} else if ast.IsParenthesizedExpression(stmt) {
						checkExpressionForExAssign(stmt.AsParenthesizedExpression().Expression, namesSet, ctx)
					}
				}
			},
		}
	},
}
