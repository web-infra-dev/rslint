package no_unreachable

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-unreachable

// isUnreachable checks if a statement is unreachable using the binder's flow analysis.
// The binder sets FlowNode on each reachable statement; unreachable statements have nil FlowNode.
func isUnreachable(node *ast.Node) bool {
	flowData := node.FlowNodeData()
	return flowData == nil || flowData.FlowNode == nil
}

// isHoistedOrEmpty returns true if the statement is safe to appear
// after a terminal statement because it is hoisted or has no runtime effect.
// - FunctionDeclaration: hoisted
// - ClassDeclaration: NOT hoisted (has temporal dead zone), should be reported
// - EmptyStatement: no effect
// - var declarations without initializers: the declaration is hoisted
// - TypeAliasDeclaration, InterfaceDeclaration: type-only, erased at compile time
func isHoistedOrEmpty(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return true
	case ast.KindEmptyStatement:
		return true
	case ast.KindTypeAliasDeclaration,
		ast.KindInterfaceDeclaration:
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

		checkStatements := func(statements []*ast.Node) {
			if len(statements) == 0 {
				return
			}

			// If the first statement is already unreachable, this container is
			// either inside unreachable code reported by a parent, or a dead
			// branch from a constant condition (e.g. else of if(true)). In both
			// cases, skip to avoid noise and double-reporting.
			if isUnreachable(statements[0]) {
				return
			}

			var rangeStart *ast.Node // first unreachable stmt in current consecutive group
			var rangeEnd *ast.Node   // last unreachable stmt in current consecutive group

			flush := func() {
				if rangeStart != nil {
					// Trim leading trivia on the start node (same as ReportNode)
					startRange := utils.TrimNodeTextRange(ctx.SourceFile, rangeStart)
					ctx.ReportRange(
						core.NewTextRange(startRange.Pos(), rangeEnd.End()),
						msg,
					)
					rangeStart = nil
					rangeEnd = nil
				}
			}

			for _, stmt := range statements {
				if stmt == nil {
					continue
				}
				if isUnreachable(stmt) {
					if isHoistedOrEmpty(stmt) {
						// Hoisted/empty statements break the consecutive chain
						// but are not reported themselves
						flush()
					} else {
						if rangeStart == nil {
							rangeStart = stmt
						}
						rangeEnd = stmt
					}
				} else {
					flush()
				}
			}
			flush()
		}

		// Check SourceFile top-level statements directly, since the linter
		// visits SourceFile's children (not the SourceFile node itself),
		// so a KindSourceFile listener would never fire.
		if sf := ctx.SourceFile; sf != nil && sf.Statements != nil {
			checkStatements(sf.Statements.Nodes)
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
			ast.KindTryStatement: func(node *ast.Node) {
				ts := node.AsTryStatement()
				if ts == nil || ts.CatchClause == nil || ts.TryBlock == nil {
					return
				}
				// If the try block is itself unreachable, skip — the parent
				// already reported it.
				if isUnreachable(node) {
					return
				}
				// If the try block cannot throw before reaching a terminal,
				// the catch clause is unreachable.
				if !utils.CanBlockThrow(ts.TryBlock) {
					cc := ts.CatchClause.AsCatchClause()
					if cc != nil && cc.Block != nil {
						startRange := utils.TrimNodeTextRange(ctx.SourceFile, ts.CatchClause)
						ctx.ReportRange(
							core.NewTextRange(startRange.Pos(), ts.CatchClause.End()),
							msg,
						)
					}
				}
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
