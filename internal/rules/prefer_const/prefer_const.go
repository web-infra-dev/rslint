package prefer_const

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type preferConstOptions struct {
	destructuring          string // "any" or "all", default "any"
	ignoreReadBeforeAssign bool   // default false
}

func parseOptions(opts any) preferConstOptions {
	result := preferConstOptions{
		destructuring:          "any",
		ignoreReadBeforeAssign: false,
	}

	optsMap := utils.GetOptionsMap(opts)
	if optsMap != nil {
		if d, ok := optsMap["destructuring"].(string); ok && (d == "any" || d == "all") {
			result.destructuring = d
		}
		if v, ok := optsMap["ignoreReadBeforeAssign"].(bool); ok {
			result.ignoreReadBeforeAssign = v
		}
	}

	return result
}

// candidateInfo holds information about a single binding name candidate.
type candidateInfo struct {
	nameNode       *ast.Node
	hasInitializer bool
	reportNode     *ast.Node // node to report on (may differ from nameNode for uninitialized vars)
}

// https://eslint.org/docs/latest/rules/prefer-const
var PreferConstRule = rule.Rule{
	Name:             "prefer-const",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Defense-in-depth: RequiresTypeInfo: true filters this rule out for
		// gap files / inferred-project files, but if a future caller bypasses
		// the filter we still want to no-op rather than nil-deref.
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}

		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				declList := node.AsVariableDeclarationList()
				if declList == nil || node.Flags&ast.NodeFlagsLet == 0 || declList.Declarations == nil {
					return
				}

				// ESLint does not report prefer-const for regular for-loop initializer variables.
				// Only for-in and for-of loop variables are checked.
				if isInForStatement(node) {
					return
				}

				isForInOrOf := isInForInOrOf(node)

				// Collect candidates across ALL declarators in the VDL to determine
				// if the entire VDL can be auto-fixed (let → const).
				var allConstCandidates []*candidateInfo
				totalBindings := 0
				totalConstBindings := 0
				allHaveInit := true

				for _, decl := range declList.Declarations.Nodes {
					varDecl := decl.AsVariableDeclaration()
					if varDecl == nil || varDecl.Name() == nil {
						continue
					}

					hasInit := varDecl.Initializer != nil || isForInOrOf
					if !hasInit {
						allHaveInit = false
					}

					// Collect all candidate binding names from this declaration
					candidates := collectBindingNames(varDecl.Name(), hasInit)
					if len(candidates) == 0 {
						continue
					}

					totalBindings += len(candidates)

					// Check each candidate
					var constCandidates []*candidateInfo
					for i := range candidates {
						c := &candidates[i]
						if shouldReport(c, decl, &ctx, opts) {
							constCandidates = append(constCandidates, c)
						}
					}

					// Apply destructuring option
					isDestructuring := varDecl.Name().Kind == ast.KindObjectBindingPattern ||
						varDecl.Name().Kind == ast.KindArrayBindingPattern
					if isDestructuring && opts.destructuring == "all" {
						// Only report if ALL candidates in the destructuring can be const
						if len(constCandidates) != len(candidates) {
							continue
						}
					}

					// For destructuring: "all", also suppress uninitialized candidates
					// whose write is in a destructuring assignment where not all targets
					// can be const. ESLint groups by the destructuring write, not just
					// the declaration pattern.
					if opts.destructuring == "all" {
						var filtered []*candidateInfo
						for _, c := range constCandidates {
							if c.reportNode != nil && !allDestructuringWriteTargetsConst(c.reportNode, &ctx) {
								continue
							}
							filtered = append(filtered, c)
						}
						constCandidates = filtered
					}

					totalConstBindings += len(constCandidates)
					allConstCandidates = append(allConstCandidates, constCandidates...)
				}

				// Determine if auto-fix is possible: ALL bindings in the VDL must be
				// const-eligible AND all declarators must have initializers.
				// ESLint only auto-fixes when the entire `let` can become `const`.
				canFix := allHaveInit && totalBindings > 0 && totalConstBindings == totalBindings

				// Additionally, suppress fix for uninitialized candidates whose write
				// is in a destructuring with non-let targets (var/const in same scope,
				// member expressions, or cross-scope identifiers).
				if canFix {
					firstDecl := declList.Declarations.Nodes[0]
					for _, c := range allConstCandidates {
						if c.reportNode != nil {
							if isInUnfixableDestructuring(c.reportNode, firstDecl) {
								canFix = false
								break
							}
						}
					}
				}

				// Report the const candidates
				for _, c := range allConstCandidates {
					name := c.nameNode.Text()
					reportOn := c.nameNode
					if c.reportNode != nil {
						reportOn = c.reportNode
					}
					msg := rule.RuleMessage{
						Id:          "useConst",
						Description: "'" + name + "' is never reassigned. Use 'const' instead.",
					}
					if canFix {
						letRange := utils.GetVarKeywordRange(node, ctx.SourceFile)
						ctx.ReportNodeWithFixes(reportOn, msg,
							rule.RuleFixReplaceRange(letRange, "const"))
					} else {
						ctx.ReportNode(reportOn, msg)
					}
				}
			},
		}
	},
}

// shouldReport determines whether a candidate should be reported as "use const".
func shouldReport(c *candidateInfo, declNode *ast.Node, ctx *rule.RuleContext, opts preferConstOptions) bool {
	sym := ctx.TypeChecker.GetSymbolAtLocation(c.nameNode)
	if sym == nil {
		return false
	}

	if c.hasInitializer {
		// For initialized candidates: report if 0 writes after declaration (never reassigned)
		return !isReassigned(sym, c.nameNode.Text(), declNode, ctx)
	}

	// For uninitialized candidates (let x;):
	// Count write references (excluding declaration)
	writeCount := countWriteReferences(sym, c.nameNode.Text(), declNode, ctx)
	if writeCount != 1 {
		// 0 writes: never assigned, "let x;" alone is fine - don't report
		// 2+ writes: truly reassigned - don't report
		return false
	}

	// Exactly 1 write: single assignment, could be "const x = ..."
	// But only if the write is at the same block level as the declaration.
	// If the write is inside a nested block (if, for, try, function, etc.),
	// we can't safely convert to "const x = ..." because it would change semantics.
	writeNode := findWriteInSameBlock(sym, c.nameNode.Text(), declNode, ctx)
	if writeNode == nil {
		return false
	}
	// ESLint reports uninitialized variables at the write location when there's no read
	// between declaration and write. If there IS a read before write, report at the declaration.
	readBeforeAssign := isReadBeforeFirstAssign(sym, c.nameNode.Text(), declNode, ctx)
	if !readBeforeAssign {
		c.reportNode = writeNode
	}

	// Check ignoreReadBeforeAssign option
	if opts.ignoreReadBeforeAssign && readBeforeAssign {
		return false
	}
	return true
}

// collectBindingNames collects all identifier nodes from a binding pattern
// using the public utils.CollectBindingNames utility.
func collectBindingNames(nameNode *ast.Node, hasInitializer bool) []candidateInfo {
	var result []candidateInfo
	utils.CollectBindingNames(nameNode, func(ident *ast.Node, _ string) {
		result = append(result, candidateInfo{
			nameNode:       ident,
			hasInitializer: hasInitializer,
		})
	})
	return result
}

// isInForStatement checks if a VariableDeclarationList is the initializer of a regular for statement.
func isInForStatement(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	return node.Parent.Kind == ast.KindForStatement
}

// isInForInOrOf checks if a VariableDeclarationList is the initializer of a for-in or for-of statement.
func isInForInOrOf(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	return node.Parent.Kind == ast.KindForInStatement || node.Parent.Kind == ast.KindForOfStatement
}

// isReassigned checks if a symbol is ever assigned to after its declaration.
func isReassigned(sym *ast.Symbol, declName string, declNode *ast.Node, ctx *rule.RuleContext) bool {
	return countWriteReferences(sym, declName, declNode, ctx) > 0
}

// countWriteReferences counts the number of write references to a symbol after its declaration.
func countWriteReferences(sym *ast.Symbol, declName string, declNode *ast.Node, ctx *rule.RuleContext) int {
	// Find enclosing scope to limit the walk
	scope := findEnclosingScope(declNode)
	if scope == nil {
		scope = ctx.SourceFile.AsNode()
	}

	count := 0
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}

		if n.Kind == ast.KindIdentifier && !isPartOfDeclaration(n, declNode) {
			refSym := ctx.TypeChecker.GetSymbolAtLocation(n)
			if refSym == sym && utils.IsWriteReference(n) {
				count++
			}
		}

		// Also check ShorthandPropertyAssignment - in ({x} = {x: 2}), the TypeChecker
		// resolves the shorthand name to the property symbol, not the variable symbol.
		// Use name-based matching combined with scope check for this case.
		// To avoid false-matching shadowed variables, verify that the shorthand's
		// value symbol (obtained via the generated identifier) matches the target.
		if n.Kind == ast.KindShorthandPropertyAssignment && !isPartOfDeclaration(n, declNode) {
			shorthand := n.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil && utils.IsInDestructuringAssignment(n) {
				name := shorthand.Name().Text()
				if name == declName && isInSameScope(n, declNode) {
					// Guard against shadowed variables: if the TypeChecker can resolve
					// the shorthand's value to a different symbol, skip it.
					valSym := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(n)
					if valSym == nil || valSym == sym {
						count++
						return // Skip children to avoid double-counting the name identifier via Path 1
					}
				}
			}
		}

		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(scope)
	return count
}

// isReadBeforeFirstAssign checks if a variable is read between its declaration and first assignment.
// This is used for uninitialized variables (let x;) when ignoreReadBeforeAssign is true.
func isReadBeforeFirstAssign(sym *ast.Symbol, declName string, declNode *ast.Node, ctx *rule.RuleContext) bool {
	scope := findEnclosingScope(declNode)
	if scope == nil {
		scope = ctx.SourceFile.AsNode()
	}

	// Walk the scope in source order; track whether we passed the declaration and found the first write
	pastDecl := false
	foundRead := false

	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n == nil {
			return false
		}

		// Check if we reached the declaration
		if n == declNode {
			pastDecl = true
			return false
		}

		if !pastDecl {
			done := false
			n.ForEachChild(func(child *ast.Node) bool {
				if walk(child) {
					done = true
					return true
				}
				return false
			})
			return done
		}

		// After declaration, check for reads and writes
		if n.Kind == ast.KindIdentifier {
			refSym := ctx.TypeChecker.GetSymbolAtLocation(n)
			if refSym == sym {
				if utils.IsWriteReference(n) {
					// Found the first write - stop walking
					return true
				}
				// It is a read reference before the first write
				foundRead = true
			}
		}

		// Also detect writes through ShorthandPropertyAssignment in destructuring.
		// The TypeChecker may resolve shorthand names to property symbols instead of
		// variable symbols, so we use name-based matching (same as countWriteReferences).
		if n.Kind == ast.KindShorthandPropertyAssignment && !isPartOfDeclaration(n, declNode) {
			shorthand := n.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil && utils.IsInDestructuringAssignment(n) {
				if shorthand.Name().Text() == declName && isInSameScope(n, declNode) {
					valSym := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(n)
					if valSym == nil || valSym == sym {
						// Found the first write via shorthand destructuring - stop walking
						return true
					}
				}
			}
		}

		done := false
		n.ForEachChild(func(child *ast.Node) bool {
			if walk(child) {
				done = true
				return true
			}
			return false
		})
		return done
	}
	walk(scope)
	return foundRead
}

// findEnclosingScope delegates to the public utils.FindEnclosingScope.
func findEnclosingScope(node *ast.Node) *ast.Node {
	return utils.FindEnclosingScope(node)
}

// isPartOfDeclaration checks if an identifier node is part of the variable declaration's
// binding name (not its initializer). This ensures that writes inside the initializer
// (e.g. `let x = (x = 1)`) are still counted as reassignments.
func isPartOfDeclaration(identNode *ast.Node, declNode *ast.Node) bool {
	varDecl := declNode.AsVariableDeclaration()
	if varDecl == nil || varDecl.Name() == nil {
		return false
	}
	// Check if the identifier is a descendant of the binding name node
	nameNode := varDecl.Name()
	return ast.FindAncestorOrQuit(identNode, func(n *ast.Node) ast.FindAncestorResult {
		if n == nameNode {
			return ast.FindAncestorTrue
		}
		// Stop if we've left the declaration entirely
		if n == declNode {
			return ast.FindAncestorQuit
		}
		return ast.FindAncestorFalse
	}) != nil
}

// isInSameScope checks if two nodes share the same enclosing function/module/source scope.
func isInSameScope(a *ast.Node, b *ast.Node) bool {
	return findEnclosingScope(a) == findEnclosingScope(b)
}

// findContainingBlock finds the nearest Block, SourceFile, ModuleBlock, CaseClause, or DefaultClause ancestor.
func findContainingBlock(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock,
			ast.KindCaseClause, ast.KindDefaultClause:
			return true
		}
		return false
	})
}

// isDirectChildOfBlock checks if a write node is a direct statement within the
// given block. Walks from the write node to the block, ensuring there are no
// intervening control flow statements (if, while, for, etc.) or nested blocks.
// This handles both braced (`if (c) { x=1; }`) and brace-less (`if (c) x=1;`) forms.
func isDirectChildOfBlock(writeNode *ast.Node, declBlock *ast.Node) bool {
	current := writeNode.Parent
	for current != nil {
		if current == declBlock {
			return true
		}
		switch current.Kind {
		// Any of these between the write and the block means the write is nested
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock,
			ast.KindCaseClause, ast.KindDefaultClause,
			ast.KindIfStatement, ast.KindWhileStatement, ast.KindDoStatement,
			ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindWithStatement, ast.KindLabeledStatement, ast.KindSwitchStatement:
			return false
		}
		current = current.Parent
	}
	return false
}

// findWriteInSameBlock finds the single write reference to an uninitialized variable
// if it is at the same block nesting level as its declaration. Returns the write node,
// or nil if the write is in a nested block or not found.
// ESLint only suggests const for uninitialized variables when the write can be merged
// into the declaration (i.e., same block level). Writes inside nested blocks (if, for,
// try, function bodies, etc.) cannot be safely merged.
func findWriteInSameBlock(sym *ast.Symbol, declName string, declNode *ast.Node, ctx *rule.RuleContext) *ast.Node {
	declBlock := findContainingBlock(declNode)
	if declBlock == nil {
		return nil
	}

	scope := findEnclosingScope(declNode)
	if scope == nil {
		scope = ctx.SourceFile.AsNode()
	}

	var writeNode *ast.Node
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || writeNode != nil {
			return
		}

		if n.Kind == ast.KindIdentifier && !isPartOfDeclaration(n, declNode) {
			refSym := ctx.TypeChecker.GetSymbolAtLocation(n)
			if refSym == sym && utils.IsWriteReference(n) {
				writeNode = n
				return
			}
		}

		if n.Kind == ast.KindShorthandPropertyAssignment && !isPartOfDeclaration(n, declNode) {
			shorthand := n.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil && utils.IsInDestructuringAssignment(n) {
				name := shorthand.Name().Text()
				if name == declName && isInSameScope(n, declNode) {
					valSym := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(n)
					if valSym == nil || valSym == sym {
						writeNode = shorthand.Name()
						return
					}
				}
			}
		}

		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(scope)

	if writeNode == nil {
		return nil
	}

	// The write must be a direct statement in the declaration's containing block.
	// Writes inside nested blocks, control flow bodies (if/while/for without braces),
	// labeled statements, etc. cannot be safely merged into the declaration.
	if !isDirectChildOfBlock(writeNode, declBlock) {
		return nil
	}

	// ESLint only flags uninitialized variables when the write is a standalone
	// assignment statement (ExpressionStatement > AssignmentExpression),
	// not when embedded in conditions, chained assignments, or other expressions.
	if !isStandaloneAssignment(writeNode) {
		return nil
	}

	// ESLint doesn't report variables in destructuring assignments that contain
	// member expressions (obj.prop) or identifiers from a different block scope.
	// Same-scope var/const targets don't suppress reporting (only suppress fix).
	if hasNonReportableDestructuringTarget(writeNode, declNode, ctx) {
		return nil
	}

	return writeNode
}

// allDestructuringWriteTargetsConst checks whether all identifier targets in the
// destructuring assignment containing writeNode have at most 1 write reference.
// Used with destructuring: "all" to suppress reporting when not all variables in
// the destructuring write group can be const.
// Uses name-based matching because GetSymbolAtLocation on shorthand property names
// in destructuring assignments resolves to the property symbol, not the variable.
func allDestructuringWriteTargetsConst(writeNode *ast.Node, ctx *rule.RuleContext) bool {
	assignExpr := ast.FindAncestor(writeNode, func(n *ast.Node) bool {
		return ast.IsDestructuringAssignment(n)
	})
	if assignExpr == nil {
		return true // not in a destructuring, no group constraint
	}

	scope := findEnclosingScope(writeNode)
	if scope == nil {
		scope = ctx.SourceFile.AsNode()
	}

	left := assignExpr.AsBinaryExpression().Left
	allConst := true
	utils.VisitDestructuringIdentifiers(left, func(ident *ast.Node) {
		if !allConst {
			return
		}
		// Resolve the symbol directly from the identifier node provided by
		// VisitDestructuringIdentifiers. For shorthand properties, the parent
		// is a ShorthandPropertyAssignment and GetSymbolAtLocation returns the
		// property symbol, so use GetShorthandAssignmentValueSymbol instead.
		var sym *ast.Symbol
		if ident.Parent != nil && ident.Parent.Kind == ast.KindShorthandPropertyAssignment {
			sym = ctx.TypeChecker.GetShorthandAssignmentValueSymbol(ident.Parent)
		} else {
			sym = ctx.TypeChecker.GetSymbolAtLocation(ident)
		}
		if sym == nil {
			return
		}
		if countWritesBySym(sym, ident.Text(), scope, ctx) > 1 {
			allConst = false
		}
	})
	return allConst
}

// countWritesBySym counts write references to a specific symbol within a scope.
// Uses symbol comparison for identifiers and GetShorthandAssignmentValueSymbol
// for shorthand property assignments.
func countWritesBySym(sym *ast.Symbol, name string, scope *ast.Node, ctx *rule.RuleContext) int {
	count := 0
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindIdentifier && n.Text() == name && utils.IsWriteReference(n) {
			refSym := ctx.TypeChecker.GetSymbolAtLocation(n)
			if refSym == sym {
				count++
			}
		}
		if n.Kind == ast.KindShorthandPropertyAssignment {
			shorthand := n.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil &&
				shorthand.Name().Text() == name && utils.IsInDestructuringAssignment(n) {
				valSym := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(n)
				if valSym == nil || valSym == sym {
					count++
					return // avoid double-counting
				}
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(scope)
	return count
}

// hasNonReportableDestructuringTarget checks if a write node is inside a destructuring
// assignment that should suppress REPORTING. This is limited to:
// 1. Member expressions (obj.prop, arr[i]) — can never be const declarations
// 2. Identifiers whose declaration is in a different block scope — can't safely merge
// Same-scope identifiers (var, const, import, param, etc.) do NOT suppress reporting.
// Uses TypeChecker symbol resolution instead of name-set collection to correctly
// handle all declaration types (imports, function/class declarations, parameters, etc.).
func hasNonReportableDestructuringTarget(writeNode *ast.Node, declNode *ast.Node, ctx *rule.RuleContext) bool {
	assignExpr := ast.FindAncestor(writeNode, func(n *ast.Node) bool {
		return ast.IsDestructuringAssignment(n)
	})
	if assignExpr == nil {
		return false
	}

	left := assignExpr.AsBinaryExpression().Left
	declBlock := findContainingBlock(declNode)
	if declBlock == nil {
		return true
	}
	return hasNonReportableTarget(left, declBlock, ctx)
}

// hasNonReportableTarget checks if a destructuring pattern contains targets that
// should suppress reporting: member expressions, or identifiers declared in a
// different block scope. Uses TypeChecker to resolve each identifier's declaration
// rather than pre-collecting names, so it correctly handles imports, parameters,
// function declarations, class declarations, etc.
func hasNonReportableTarget(node *ast.Node, declBlock *ast.Node, ctx *rule.RuleContext) bool {
	if node.Kind == ast.KindIdentifier {
		sym := ctx.TypeChecker.GetSymbolAtLocation(node)
		if sym == nil || len(sym.Declarations) == 0 {
			// Can't resolve — treat as non-reportable (conservative)
			return true
		}
		targetDeclBlock := findContainingBlock(sym.Declarations[0])
		return targetDeclBlock != declBlock
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if found {
			return true
		}
		switch child.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			found = true
			return true
		case ast.KindIdentifier:
			sym := ctx.TypeChecker.GetSymbolAtLocation(child)
			if sym == nil || len(sym.Declarations) == 0 {
				found = true
				return true
			}
			if findContainingBlock(sym.Declarations[0]) != declBlock {
				found = true
				return true
			}
		case ast.KindShorthandPropertyAssignment:
			shorthand := child.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil {
				valSym := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(child)
				if valSym == nil || len(valSym.Declarations) == 0 {
					found = true
					return true
				}
				if findContainingBlock(valSym.Declarations[0]) != declBlock {
					found = true
					return true
				}
			}
		case ast.KindPropertyAssignment:
			pa := child.AsPropertyAssignment()
			if pa != nil && pa.Initializer != nil {
				if hasNonReportableTarget(pa.Initializer, declBlock, ctx) {
					found = true
					return true
				}
			}
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindSpreadElement, ast.KindSpreadAssignment:
			if hasNonReportableTarget(child, declBlock, ctx) {
				found = true
				return true
			}
		case ast.KindBinaryExpression:
			be := child.AsBinaryExpression()
			if be != nil && be.Left != nil {
				if hasNonReportableTarget(be.Left, declBlock, ctx) {
					found = true
					return true
				}
			}
		}
		return false
	})
	return found
}

// isInUnfixableDestructuring checks if a write node is inside a destructuring
// assignment that should suppress AUTO-FIX. This is stricter than the reporting
// check — it also rejects same-scope var/const targets (can't change `let` to
// `const` if the destructuring also writes to a non-let variable).
func isInUnfixableDestructuring(writeNode *ast.Node, declNode *ast.Node) bool {
	assignExpr := ast.FindAncestor(writeNode, func(n *ast.Node) bool {
		return ast.IsDestructuringAssignment(n)
	})
	if assignExpr == nil {
		return false
	}

	left := assignExpr.AsBinaryExpression().Left
	return hasUnmergeableTarget(left, declNode)
}

// hasUnmergeableTarget checks if a destructuring target contains elements that
// prevent merging into a const declaration: member expressions, or identifiers
// declared in a different block scope than the candidate.
// ESLint reports cross-declaration destructuring when all targets are in the same
// block scope (with output: null for no auto-fix), but skips when targets span
// different scopes.
func hasUnmergeableTarget(lhs *ast.Node, declNode *ast.Node) bool {
	declBlock := findContainingBlock(declNode)
	if declBlock == nil {
		return true
	}
	blockNames := collectBlockLetNames(declBlock)
	return hasTargetNotInSet(lhs, blockNames)
}

// collectBlockLetNames collects all variable names from let declarations that are
// direct children of the given block. Used for the auto-fix check (only let
// declarations can be converted to const).
func collectBlockLetNames(block *ast.Node) map[string]bool {
	names := make(map[string]bool)
	block.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == ast.KindVariableStatement {
			child.ForEachChild(func(grandchild *ast.Node) bool {
				if grandchild.Kind == ast.KindVariableDeclarationList && grandchild.Flags&ast.NodeFlagsLet != 0 {
					declList := grandchild.AsVariableDeclarationList()
					if declList != nil && declList.Declarations != nil {
						for _, decl := range declList.Declarations.Nodes {
							varDecl := decl.AsVariableDeclaration()
							if varDecl != nil && varDecl.Name() != nil {
								utils.CollectBindingNames(varDecl.Name(), func(_ *ast.Node, name string) {
									names[name] = true
								})
							}
						}
					}
				}
				return false
			})
		}
		return false
	})
	return names
}

// hasTargetNotInSet recursively checks if any target in a destructuring pattern
// is a member expression or an identifier NOT in the given name set.
func hasTargetNotInSet(node *ast.Node, names map[string]bool) bool {
	// Handle leaf Identifier nodes directly (reached via PropertyAssignment or
	// BinaryExpression default-value recursion).
	if node.Kind == ast.KindIdentifier {
		return !names[node.Text()]
	}

	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if found {
			return true
		}
		switch child.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			found = true
			return true
		case ast.KindIdentifier:
			if !names[child.Text()] {
				found = true
				return true
			}
		case ast.KindShorthandPropertyAssignment:
			shorthand := child.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil {
				if !names[shorthand.Name().Text()] {
					found = true
					return true
				}
			}
		case ast.KindPropertyAssignment:
			// Only check the value (target), not the key
			pa := child.AsPropertyAssignment()
			if pa != nil && pa.Initializer != nil {
				if hasTargetNotInSet(pa.Initializer, names) {
					found = true
					return true
				}
			}
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindSpreadElement, ast.KindSpreadAssignment:
			if hasTargetNotInSet(child, names) {
				found = true
				return true
			}
		case ast.KindBinaryExpression:
			// Default value: [x = 5] → check left side only
			be := child.AsBinaryExpression()
			if be != nil && be.Left != nil {
				if hasTargetNotInSet(be.Left, names) {
					found = true
					return true
				}
			}
		}
		return false
	})
	return found
}

// isStandaloneAssignment checks if a write reference identifier is part of an
// assignment expression that is directly inside an ExpressionStatement.
// Uses ast.GetAssignmentTarget from the TypeScript shim for robust pattern walking
// through destructuring, shorthand properties, parentheses, etc.
// Returns false for ++/--, for-in/of targets, conditions, chained assignments,
// and other non-statement expressions.
func isStandaloneAssignment(identNode *ast.Node) bool {
	target := ast.GetAssignmentTarget(identNode)
	if target == nil {
		return false
	}

	// ++/-- (PrefixUnary/PostfixUnary) can't be converted to const for uninitialized variables.
	// for-in/of targets may execute multiple times.
	switch target.Kind {
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression,
		ast.KindForInStatement, ast.KindForOfStatement:
		return false
	}

	// GetAssignmentTarget may return a default value's BinaryExpression
	// (e.g. [x = 5] returns x=5, not [x=5]=[1]). If so, find the outer destructuring.
	if target.Kind == ast.KindBinaryExpression && isDefaultValueInDestructuring(target) {
		target = ast.FindAncestor(target.Parent, func(n *ast.Node) bool {
			return ast.IsDestructuringAssignment(n)
		})
		if target == nil {
			return false
		}
	}

	// The assignment must be directly inside an ExpressionStatement (possibly through parens).
	parent := target.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	return parent != nil && parent.Kind == ast.KindExpressionStatement
}

// isDefaultValueInDestructuring checks if a BinaryExpression(=) node is a default
// value inside a destructuring assignment target (e.g., x = 5 in [x = 5] = [1]
// or val: x = 5 in ({val: x = 5} = {val: 1})).
func isDefaultValueInDestructuring(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindArrayLiteralExpression:
		// [x = 5] — default in array destructuring target
		return utils.IsInDestructuringAssignment(parent)
	case ast.KindPropertyAssignment:
		// {val: x = 5} — default in object destructuring rename
		pa := parent.AsPropertyAssignment()
		if pa != nil && pa.Initializer == node {
			return utils.IsInDestructuringAssignment(parent)
		}
	case ast.KindSpreadElement:
		// [...x = 5] — unlikely but handle for completeness
		return utils.IsInDestructuringAssignment(parent)
	}
	return false
}
