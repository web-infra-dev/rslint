package no_unreachable

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// https://eslint.org/docs/latest/rules/no-unreachable

// isTerminal returns true if the statement is a terminal statement
// that prevents subsequent statements from being reached.
func isTerminal(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindReturnStatement,
		ast.KindThrowStatement,
		ast.KindBreakStatement,
		ast.KindContinueStatement:
		return true
	}
	return false
}

// isHoistedOrEmpty returns true if the statement is safe to appear
// after a terminal statement because it is hoisted or has no runtime effect.
// - FunctionDeclaration: hoisted
// - ClassDeclaration: NOT hoisted (has temporal dead zone), should be reported
// - EmptyStatement: no effect
// - var declarations without initializers: the declaration is hoisted
func isHoistedOrEmpty(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return true
	case ast.KindEmptyStatement:
		return true
	case ast.KindVariableStatement:
		return isVarWithoutInitializer(node)
	}
	return false
}

// isVarWithoutInitializer checks if a VariableStatement is a `var` declaration
// where none of the declarators have initializers. `let` and `const` are not
// hoisted in the same way, so they are always reported.
func isVarWithoutInitializer(node *ast.Node) bool {
	varStmt := node.AsVariableStatement()
	if varStmt == nil || varStmt.DeclarationList == nil {
		return false
	}

	declList := varStmt.DeclarationList.AsVariableDeclarationList()
	if declList == nil {
		return false
	}

	// If it's let, const, or using, it's not hoisted like var
	flags := varStmt.DeclarationList.Flags
	if flags&ast.NodeFlagsLet != 0 || flags&ast.NodeFlagsConst != 0 || flags&ast.NodeFlagsUsing != 0 {
		return false
	}

	// Check that all declarations have no initializer
	if declList.Declarations == nil {
		return true
	}
	for _, decl := range declList.Declarations.Nodes {
		if decl.Kind != ast.KindVariableDeclaration {
			continue
		}
		varDecl := decl.AsVariableDeclaration()
		if varDecl != nil && varDecl.Initializer != nil {
			return false
		}
	}
	return true
}

// NoUnreachableRule disallows unreachable code after return, throw, break, and continue statements.
var NoUnreachableRule = rule.Rule{
	Name: "no-unreachable",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		msg := rule.RuleMessage{
			Id:          "unreachableCode",
			Description: "Unreachable code.",
		}

		checkStatements := func(stmts []*ast.Node) {
			foundTerminal := false
			for _, stmt := range stmts {
				if stmt == nil {
					continue
				}
				if foundTerminal {
					if !isHoistedOrEmpty(stmt) {
						ctx.ReportNode(stmt, msg)
					}
				} else if isTerminal(stmt) {
					foundTerminal = true
				}
			}
		}

		return rule.RuleListeners{
			ast.KindBlock: func(node *ast.Node) {
				block := node.AsBlock()
				if block == nil || block.Statements == nil {
					return
				}
				checkStatements(block.Statements.Nodes)
			},
			ast.KindCaseClause: func(node *ast.Node) {
				clause := node.AsCaseOrDefaultClause()
				if clause == nil || clause.Statements == nil {
					return
				}
				checkStatements(clause.Statements.Nodes)
			},
			ast.KindDefaultClause: func(node *ast.Node) {
				clause := node.AsCaseOrDefaultClause()
				if clause == nil || clause.Statements == nil {
					return
				}
				checkStatements(clause.Statements.Nodes)
			},
			ast.KindSourceFile: func(node *ast.Node) {
				sf := node.AsSourceFile()
				if sf == nil || sf.Statements == nil {
					return
				}
				checkStatements(sf.Statements.Nodes)
			},
			ast.KindModuleBlock: func(node *ast.Node) {
				mb := node.AsModuleBlock()
				if mb == nil || mb.Statements == nil {
					return
				}
				checkStatements(mb.Statements.Nodes)
			},
		}
	},
}
