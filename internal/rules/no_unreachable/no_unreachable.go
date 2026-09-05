package no_unreachable

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
)

// https://eslint.org/docs/latest/rules/no-unreachable

// isUnreachable checks if a statement is unreachable using the binder's flow analysis.
// The binder assigns FlowNode only to nodes in [KindFirstStatement, KindLastStatement].
// Declaration nodes like EnumDeclaration, ClassDeclaration, and ModuleDeclaration fall
// outside that range and never receive a FlowNode, so a nil FlowNode alone does not imply
// the node is unreachable. For those nodes we fall back to the NodeFlagsUnreachable flag that the
// binder sets on all potentially-executable unreachable nodes.
func isUnreachable(node *ast.Node) bool {
	flowData := node.FlowNodeData()
	if flowData != nil && flowData.FlowNode != nil {
		return false
	}
	// For nodes in the statement range, the binder always assigns FlowNode;
	// a nil value reliably means unreachable.
	if node.Kind >= ast.KindFirstStatement && node.Kind <= ast.KindLastStatement {
		return true
	}
	// For nodes outside the range, use the explicit flag.
	return node.Flags&ast.NodeFlagsUnreachable != 0
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

// isBooleanLiteralIf reports whether node is an if statement whose condition is
// folded by the TypeScript binder but not by ESLint's code-path analyzer.
func isBooleanLiteralIf(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIfStatement {
		return false
	}
	stmt := node.AsIfStatement()
	return stmt != nil && stmt.Expression != nil &&
		(stmt.Expression.Kind == ast.KindTrueKeyword || stmt.Expression.Kind == ast.KindFalseKeyword)
}

func isCodePathRoot(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindSourceFile ||
		node.Kind == ast.KindClassStaticBlockDeclaration ||
		utils.IsFunctionLikeContainer(node))
}

// flatStatementCompletion reports normal completion for statement shapes that
// do not need a CFG. The second result is false for compound control flow or a
// class evaluation, whose static blocks require full code-path analysis.
func flatStatementCompletion(node *ast.Node, allowBlock bool) (bool, bool) {
	if node == nil {
		return true, true
	}
	if utils.IsFunctionLikeContainer(node) {
		return true, true
	}
	if node.Kind == ast.KindClassDeclaration || node.Kind == ast.KindClassExpression || node.Kind == ast.KindModuleBlock {
		return false, false
	}
	if node.Kind == ast.KindBlock {
		if !allowBlock {
			return false, false
		}
		normal := true
		for _, statement := range node.Statements() {
			if !normal {
				// A branch with its own unreachable tail needs CFG reachability
				// so that tail is still checked when the binder pruned the branch.
				return false, false
			}
			statementNormal, flat := flatStatementCompletion(statement, true)
			if !flat {
				return false, false
			}
			normal = statementNormal
		}
		return normal, true
	}
	if ast.IsStatement(node) {
		switch node.Kind {
		case ast.KindReturnStatement, ast.KindThrowStatement,
			ast.KindBreakStatement, ast.KindContinueStatement:
			return false, true
		case ast.KindEmptyStatement, ast.KindVariableStatement,
			ast.KindExpressionStatement, ast.KindDebuggerStatement:
		default:
			return false, false
		}
	}
	flat := true
	node.ForEachChild(func(child *ast.Node) bool {
		_, childFlat := flatStatementCompletion(child, false)
		if !childFlat {
			flat = false
			return true
		}
		return false
	})
	return true, flat
}

func flatBooleanIfCompletion(node *ast.Node) (normal bool, needsCompensation bool, flat bool) {
	if !isBooleanLiteralIf(node) {
		return false, false, false
	}
	statement := node.AsIfStatement()
	thenNormal, thenFlat := flatStatementCompletion(statement.ThenStatement, true)
	elseNormal, elseFlat := flatStatementCompletion(statement.ElseStatement, true)
	if !thenFlat || !elseFlat {
		return false, false, false
	}
	normal = statement.ElseStatement == nil || thenNormal || elseNormal
	selectedNormal := thenNormal
	if statement.Expression.Kind == ast.KindFalseKeyword {
		selectedNormal = elseNormal
	}
	return normal, !selectedNormal && normal, true
}

func enclosingCodePathRoot(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if isCodePathRoot(current) {
			return current
		}
	}
	return nil
}

func compoundStatementContainsBooleanIf(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindBlock, ast.KindIfStatement, ast.KindDoStatement,
		ast.KindWhileStatement, ast.KindForStatement, ast.KindForInStatement,
		ast.KindForOfStatement, ast.KindWithStatement, ast.KindSwitchStatement,
		ast.KindLabeledStatement, ast.KindTryStatement, ast.KindModuleDeclaration,
		ast.KindModuleBlock:
	default:
		return false
	}
	found := false
	var visit func(*ast.Node) bool
	visit = func(current *ast.Node) bool {
		if current == nil || found || current != node && isCodePathRoot(current) {
			return false
		}
		if isBooleanLiteralIf(current) {
			found = true
			return true
		}
		current.ForEachChild(visit)
		return found
	}
	visit(node)
	return found
}

type reachabilityState struct {
	eslintReachable map[*ast.Node]bool
	cfgRoots        map[*ast.Node]bool
}

func (s *reachabilityState) ensureCFG(root *ast.Node) {
	if root == nil || s.cfgRoots[root] {
		return
	}
	if s.cfgRoots == nil {
		s.cfgRoots = make(map[*ast.Node]bool)
	}
	s.cfgRoots[root] = true
	cfg.Build(root, cfg.Hooks[struct{}]{
		Statement: func(builder *cfg.Builder[struct{}], statement *ast.Node) {
			if !builder.Current().Reachable {
				return
			}
			if s.eslintReachable == nil {
				s.eslintReachable = make(map[*ast.Node]bool)
			}
			s.eslintReachable[statement] = true
		},
	})
}

func (s *reachabilityState) isUnreachable(node *ast.Node) bool {
	return isUnreachable(node) && !s.eslintReachable[node]
}

// NoUnreachableRule disallows unreachable code after return, throw, break, and continue statements.
var NoUnreachableRule = rule.Rule{
	Name:   "no-unreachable",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		msg := rule.RuleMessage{
			Id:          "unreachableCode",
			Description: "Unreachable code.",
		}
		state := reachabilityState{}

		checkStatements := func(statements []*ast.Node) {
			if len(statements) == 0 {
				return
			}
			var root *ast.Node
			rootUsesCFG := false
			if state.cfgRoots != nil {
				root = enclosingCodePathRoot(statements[0])
				rootUsesCFG = state.cfgRoots[root]
			}

			// Unreachable containers are either covered by a parent range or are
			// dead branches that this rule intentionally does not report wholesale.
			if state.isUnreachable(statements[0]) {
				return
			}

			var rangeStart *ast.Node // first unreachable stmt in current consecutive group
			var rangeEnd *ast.Node   // last unreachable stmt in current consecutive group
			synthetic := false
			syntheticReachable := false

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
				normalCompletion := true
				candidateNeedsCompensation := false
				if !rootUsesCFG && (!synthetic || syntheticReachable) {
					needsCFG := false
					if isBooleanLiteralIf(stmt) {
						var flat bool
						normalCompletion, candidateNeedsCompensation, flat = flatBooleanIfCompletion(stmt)
						needsCFG = !flat
					} else if synthetic {
						var flat bool
						normalCompletion, flat = flatStatementCompletion(stmt, false)
						needsCFG = !flat
					} else if compoundStatementContainsBooleanIf(stmt) {
						needsCFG = true
					}
					if needsCFG {
						if root == nil {
							root = enclosingCodePathRoot(stmt)
						}
						state.ensureCFG(root)
						rootUsesCFG = root != nil && state.cfgRoots[root]
						synthetic = false
					}
				}

				reachable := !state.isUnreachable(stmt)
				if synthetic {
					reachable = syntheticReachable
				}

				if !reachable {
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

				if !rootUsesCFG && reachable && candidateNeedsCompensation {
					synthetic = true
					syntheticReachable = true
				} else if synthetic && reachable {
					syntheticReachable = normalCompletion
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
				if state.isUnreachable(node) {
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
