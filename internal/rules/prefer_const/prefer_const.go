package prefer_const

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
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
	nameNode   *ast.Node
	reportNode *ast.Node // node to report on (may differ from nameNode for uninitialized vars)
}

// https://eslint.org/docs/latest/rules/prefer-const
var PreferConstRule = rule.Rule{
	Name: "prefer-const",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := parseOptions(options)
		destructuringAll := opts.destructuring == "all"
		sourceText := ctx.SourceFile.Text()

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
				// Keep the common small declaration list on the stack.
				var candidateBuffer [4]candidateInfo
				allConstCandidates := candidateBuffer[:0]
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

					// Collect reportable candidates directly into the declaration-list
					// result. The overwhelmingly common identifier case avoids the
					// temporary binding and reportable-candidate slices entirely.
					nameNode := varDecl.Name()
					candidateStart := len(allConstCandidates)
					candidateCount := 0
					if nameNode.Kind == ast.KindIdentifier {
						if nameNode.Text() == "" {
							continue
						}
						candidateCount = 1
						candidate := candidateInfo{nameNode: nameNode}
						if shouldReport(&candidate, decl, &ctx, opts, hasInit) {
							allConstCandidates = append(allConstCandidates, candidate)
						}
					} else {
						utils.CollectBindingNames(nameNode, func(ident *ast.Node, _ string) {
							candidateCount++
							candidate := candidateInfo{nameNode: ident}
							if shouldReport(&candidate, decl, &ctx, opts, hasInit) {
								allConstCandidates = append(allConstCandidates, candidate)
							}
						})
					}
					if candidateCount == 0 {
						continue
					}

					totalBindings += candidateCount
					constCandidateCount := len(allConstCandidates) - candidateStart

					// Apply destructuring option
					isDestructuring := nameNode.Kind == ast.KindObjectBindingPattern ||
						nameNode.Kind == ast.KindArrayBindingPattern
					if isDestructuring && destructuringAll {
						// Only report if ALL candidates in the destructuring can be const
						if constCandidateCount != candidateCount {
							allConstCandidates = allConstCandidates[:candidateStart]
							continue
						}
					}

					// For destructuring: "all", also suppress uninitialized candidates
					// whose write is in a destructuring assignment where not all targets
					// can be const. ESLint groups by the destructuring write, not just
					// the declaration pattern.
					if destructuringAll {
						writeIndex := candidateStart
						for _, c := range allConstCandidates[candidateStart:] {
							if c.reportNode != nil && !allDestructuringWriteTargetsConst(c.reportNode, &ctx) {
								continue
							}
							allConstCandidates[writeIndex] = c
							writeIndex++
						}
						allConstCandidates = allConstCandidates[:writeIndex]
					}

					totalConstBindings += len(allConstCandidates) - candidateStart
				}

				// Determine if auto-fix is possible: ALL bindings in the VDL must be
				// const-eligible AND all declarators must have initializers.
				// ESLint only auto-fixes when the entire `let` can become `const`.
				// reportNode is set only for an uninitialized binding, so a candidate
				// that would need its later assignment merged can never reach canFix.
				canFix := allHaveInit && totalBindings > 0 && totalConstBindings == totalBindings

				// Report the const candidates
				var buildFix func() []rule.RuleFix
				if canFix {
					letStart := -1
					buildFix = func() []rule.RuleFix {
						if letStart < 0 {
							letStart = scanner.SkipTrivia(sourceText, node.Pos())
						}
						letRange := node.Loc.WithPos(letStart).WithEnd(letStart + len("let"))
						return []rule.RuleFix{rule.RuleFixReplaceRange(letRange, "const")}
					}
				}
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
					reportRange := reportOn.Loc.WithPos(scanner.SkipTrivia(sourceText, reportOn.Pos()))
					if canFix {
						ctx.ReportRangeWithDeferredFixes(reportRange, msg, buildFix)
					} else {
						ctx.ReportRange(reportRange, msg)
					}
				}
			},
		}
	},
}

// shouldReport determines whether a candidate should be reported as "use const".
func shouldReport(c *candidateInfo, declNode *ast.Node, ctx *rule.RuleContext, opts preferConstOptions, hasInitializer bool) bool {
	// The binder attaches the symbol to the enclosing declaration
	// (VariableDeclaration or BindingElement). ctx.Refs is keyed by that
	// binder symbol, not by checker.GetSymbolAtLocation results.
	decl := c.nameNode.Parent
	if decl == nil || decl.Name() != c.nameNode {
		return false
	}
	sym := decl.Symbol()
	if sym == nil {
		return false
	}
	refs := ctx.Refs.References(sym)

	if hasInitializer {
		// Initialized candidates only need to know whether any later write
		// exists, so stop at the first one.
		return !hasWriteReference(refs)
	}

	// Uninitialized candidates need the write count, first write, and whether
	// a read precedes the first post-declaration write. Derive all three in a
	// single pass over the source-ordered RefStore result.
	references := summarizeReferences(refs, declNode.Pos())
	if references.writeCount != 1 {
		// 0 writes: never assigned, "let x;" alone is fine - don't report
		// 2+ writes: truly reassigned - don't report
		return false
	}

	// Exactly 1 write: single assignment, could be "const x = ..."
	// But only if the write is at the same block level as the declaration.
	// If the write is inside a nested block (if, for, try, function, etc.),
	// we can't safely convert to "const x = ..." because it would change semantics.
	writeNode := findWriteInSameBlock(references.firstWrite, declNode, ctx)
	if writeNode == nil {
		return false
	}
	// ESLint reports uninitialized variables at the write location when there's no read
	// between declaration and write. If there IS a read before write, report at the declaration.
	readBeforeAssign := references.readBeforeFirstAssign
	if !readBeforeAssign {
		c.reportNode = writeNode
	}

	// Check ignoreReadBeforeAssign option
	if opts.ignoreReadBeforeAssign && readBeforeAssign {
		return false
	}
	return true
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

// hasWriteReference reports whether RefStore found any write after excluding
// declaration names. Shorthand destructuring names remain genuine writes.
func hasWriteReference(refs []*ast.Node) bool {
	for _, ref := range refs {
		if utils.IsWriteReference(ref) {
			return true
		}
	}
	return false
}

func hasMultipleWriteReferences(refs []*ast.Node) bool {
	foundWrite := false
	for _, ref := range refs {
		if utils.IsWriteReference(ref) {
			if foundWrite {
				return true
			}
			foundWrite = true
		}
	}
	return false
}

type referenceSummary struct {
	firstWrite            *ast.Node
	writeCount            int
	readBeforeFirstAssign bool
}

// summarizeReferences classifies source-ordered references to an
// uninitialized binding. Declaration names are absent from RefStore; shorthand
// destructuring assignment names remain because they are genuine writes.
func summarizeReferences(refs []*ast.Node, declPos int) referenceSummary {
	var result referenceSummary
	readAfterDeclaration := false
	foundPostDeclWrite := false
	for _, ref := range refs {
		if utils.IsWriteReference(ref) {
			result.writeCount++
			if result.firstWrite == nil {
				result.firstWrite = ref
			}
			if result.writeCount > 1 {
				return result
			}
			if ref.Pos() >= declPos && !foundPostDeclWrite {
				result.readBeforeFirstAssign = readAfterDeclaration
				foundPostDeclWrite = true
			}
			continue
		}
		if ref.Pos() >= declPos && !foundPostDeclWrite {
			readAfterDeclaration = true
		}
	}
	if !foundPostDeclWrite {
		result.readBeforeFirstAssign = readAfterDeclaration
	}
	return result
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
func findWriteInSameBlock(writeNode *ast.Node, declNode *ast.Node, ctx *rule.RuleContext) *ast.Node {
	declBlock := findContainingBlock(declNode)
	if declBlock == nil || writeNode == nil {
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
func allDestructuringWriteTargetsConst(writeNode *ast.Node, ctx *rule.RuleContext) bool {
	assignExpr := ast.FindAncestor(writeNode, func(n *ast.Node) bool {
		return ast.IsDestructuringAssignment(n)
	})
	if assignExpr == nil {
		return true // not in a destructuring, no group constraint
	}

	left := assignExpr.AsBinaryExpression().Left
	allConst := true
	utils.VisitDestructuringIdentifiers(left, func(ident *ast.Node) {
		if !allConst {
			return
		}
		// ctx.Refs.Resolve handles shorthand property names the same as
		// plain identifiers (unlike checker.GetSymbolAtLocation, which
		// resolves a shorthand name to the property symbol rather than the
		// variable it reads/writes).
		sym := ctx.Refs.Resolve(ident)
		if sym == nil {
			return
		}
		if hasMultipleWriteReferences(ctx.Refs.References(sym)) {
			allConst = false
		}
	})
	return allConst
}

// hasNonReportableDestructuringTarget checks if a write node is inside a destructuring
// assignment that should suppress REPORTING. This is limited to:
// 1. Member expressions (obj.prop, arr[i]) — can never be const declarations
// 2. Identifiers whose declaration is in a different block scope — can't safely merge
// Same-scope identifiers (var, const, import, param, etc.) do NOT suppress reporting.
// Uses ctx.Refs symbol resolution instead of name-set collection to correctly
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
// different block scope. Uses ctx.Refs to resolve each identifier's declaration
// rather than pre-collecting names, so it correctly handles imports, parameters,
// function declarations, class declarations, etc.
func hasNonReportableTarget(node *ast.Node, declBlock *ast.Node, ctx *rule.RuleContext) bool {
	if node.Kind == ast.KindIdentifier {
		sym := ctx.Refs.Resolve(node)
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
			sym := ctx.Refs.Resolve(child)
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
				valSym := ctx.Refs.Resolve(shorthand.Name())
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
	if utils.IsDefaultValueInDestructuringAssignment(target) {
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
