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
						ctx.ReportNode(c.nameNode, rule.RuleMessage{
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
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindSourceFile, ast.KindModuleBlock:
			return current
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return current
		}
		current = current.Parent
	}
	return nil
}

// isPartOfDeclaration checks if an identifier node is part of the variable declaration itself.
func isPartOfDeclaration(identNode *ast.Node, declNode *ast.Node) bool {
	current := identNode
	for current != nil {
		if current == declNode {
			return true
		}
		if current.Kind == ast.KindVariableDeclaration {
			return false
		}
		current = current.Parent
	}
	return false
}

// isInSameScope checks if two nodes share the same enclosing function/module/source scope.
func isInSameScope(a *ast.Node, b *ast.Node) bool {
	return findEnclosingScope(a) == findEnclosingScope(b)
}
