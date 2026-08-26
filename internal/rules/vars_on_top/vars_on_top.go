package vars_on_top

import "github.com/microsoft/typescript-go/shim/ast"

import "github.com/web-infra-dev/rslint/internal/rule"

var varsOnTopMessage = rule.RuleMessage{
	Id:          "top",
	Description: "All 'var' declarations must be at the top of the function scope.",
}

// https://eslint.org/docs/latest/rules/vars-on-top
var VarsOnTopRule = rule.Rule{
	Name:   "vars-on-top",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				if node == nil || node.Flags&ast.NodeFlagsBlockScoped != 0 {
					return
				}

				statement := node
				if node.Parent != nil && node.Parent.Kind == ast.KindVariableStatement {
					statement = node.Parent
				}

				if !isAtTopOfAllowedScope(statement) {
					ctx.ReportNode(statement, varsOnTopMessage)
				}
			},
		}
	},
}

func isAtTopOfAllowedScope(statement *ast.Node) bool {
	if statement == nil || statement.Parent == nil {
		return false
	}

	container := statement.Parent
	switch container.Kind {
	case ast.KindSourceFile:
		return isVarOnTop(statement, container.AsSourceFile().Statements.Nodes, true)
	case ast.KindBlock:
		block := container.AsBlock()
		if block == nil || block.Statements == nil {
			return false
		}
		parent := container.Parent
		if parent != nil && ast.IsFunctionLike(parent) {
			return isVarOnTop(statement, block.Statements.Nodes, true)
		}
		if parent != nil && parent.Kind == ast.KindClassStaticBlockDeclaration {
			return isVarOnTop(statement, block.Statements.Nodes, false)
		}
	}
	return false
}

func isVarOnTop(statement *ast.Node, statements []*ast.Node, skipDirectivesAndImports bool) bool {
	start := 0
	if skipDirectivesAndImports {
		for start < len(statements) && (looksLikeDirective(statements[start]) || looksLikeImport(statements[start])) {
			start++
		}
	}

	for _, candidate := range statements[start:] {
		if candidate == statement {
			return true
		}
		if candidate.Kind != ast.KindVariableStatement {
			return false
		}
	}
	return false
}

func looksLikeDirective(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindExpressionStatement {
		return false
	}
	expression := node.AsExpressionStatement().Expression
	expression = ast.SkipParentheses(expression)
	return expression != nil && expression.Kind == ast.KindStringLiteral
}

func looksLikeImport(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindImportDeclaration
}
