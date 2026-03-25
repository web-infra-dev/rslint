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
	Name: "prefer-const",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
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

				for _, decl := range declList.Declarations.Nodes {
					varDecl := decl.AsVariableDeclaration()
					if varDecl == nil || varDecl.Name() == nil {
						continue
					}

					hasInit := varDecl.Initializer != nil || isForInOrOf

					// Collect all candidate binding names from this declaration
					candidates := collectBindingNames(varDecl.Name(), hasInit)
					if len(candidates) == 0 {
						continue
					}

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

					// Report the const candidates
					for _, c := range constCandidates {
						name := c.nameNode.Text()
						reportOn := c.nameNode
						if c.reportNode != nil {
							reportOn = c.reportNode
						}
						ctx.ReportNode(reportOn, rule.RuleMessage{
							Id:          "useConst",
							Description: "'" + name + "' is never reassigned. Use 'const' instead.",
						})
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
	if !isReadBeforeFirstAssign(sym, c.nameNode.Text(), declNode, ctx) {
		c.reportNode = writeNode
	}

	// Check ignoreReadBeforeAssign option
	if opts.ignoreReadBeforeAssign {
		if isReadBeforeFirstAssign(sym, c.nameNode.Text(), declNode, ctx) {
			return false
		}
	}
	return true
}

// collectBindingNames collects all identifier nodes from a binding pattern.
func collectBindingNames(nameNode *ast.Node, hasInitializer bool) []candidateInfo {
	var result []candidateInfo

	switch nameNode.Kind {
	case ast.KindIdentifier:
		result = append(result, candidateInfo{
			nameNode:       nameNode,
			hasInitializer: hasInitializer,
		})

	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		nameNode.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				bindingName := child.Name()
				if bindingName != nil {
					result = append(result, collectBindingNames(bindingName, hasInitializer)...)
				}
			}
			return false
		})
	}

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
		if n.Kind == ast.KindShorthandPropertyAssignment && !isPartOfDeclaration(n, declNode) {
			shorthand := n.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.Name() != nil && utils.IsInDestructuringAssignment(n) {
				name := shorthand.Name().Text()
				if name == declName && isInSameScope(n, declNode) {
					count++
					return // Skip children to avoid double-counting the name identifier via Path 1
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
					// Found the first write via shorthand destructuring - stop walking
					return true
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

// findEnclosingScope finds the nearest function/module/source file scope.
func findEnclosingScope(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindSourceFile, ast.KindModuleBlock,
			ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return true
		}
		return false
	})
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
					writeNode = shorthand.Name()
					return
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
	// non-mergeable targets: member expressions (obj.prop), or variables from
	// a different declaration than the candidate. The auto-fix can't produce
	// valid code in these cases.
	if isInUnmergeableDestructuring(writeNode, declNode) {
		return nil
	}

	return writeNode
}

// isInUnmergeableDestructuring checks if a write node is inside a destructuring
// assignment that can't be merged into a const declaration. This includes:
// 1. Destructuring with member expressions (e.g. [obj.prop, v] = foo())
// 2. Destructuring with targets from different VariableDeclarations
func isInUnmergeableDestructuring(writeNode *ast.Node, declNode *ast.Node) bool {
	// Find the enclosing destructuring assignment using the shim's utility
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
// declared in a different VariableDeclarationList than the candidate.
// Uses name-based matching (not symbol resolution) for reliability with shorthand properties.
func hasUnmergeableTarget(lhs *ast.Node, declNode *ast.Node) bool {
	candidateVDL := findVariableDeclarationList(declNode)
	if candidateVDL == nil {
		return true
	}
	vdlNames := collectDeclaratorNames(candidateVDL)
	return hasTargetNotInSet(lhs, vdlNames)
}

// hasTargetNotInSet recursively checks if any target in a destructuring pattern
// is a member expression or an identifier NOT in the given name set.
func hasTargetNotInSet(node *ast.Node, names map[string]bool) bool {
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
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression, ast.KindSpreadElement:
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

// findVariableDeclarationList finds the VariableDeclarationList ancestor of a declaration node.
func findVariableDeclarationList(declNode *ast.Node) *ast.Node {
	return ast.FindAncestor(declNode, func(n *ast.Node) bool {
		return n.Kind == ast.KindVariableDeclarationList
	})
}

// collectDeclaratorNames collects all variable names from a VariableDeclarationList.
func collectDeclaratorNames(vdl *ast.Node) map[string]bool {
	names := make(map[string]bool)
	declList := vdl.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return names
	}
	for _, decl := range declList.Declarations.Nodes {
		varDecl := decl.AsVariableDeclaration()
		if varDecl != nil && varDecl.Name() != nil {
			utils.CollectBindingNames(varDecl.Name(), func(_ *ast.Node, name string) {
				names[name] = true
			})
		}
	}
	return names
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
