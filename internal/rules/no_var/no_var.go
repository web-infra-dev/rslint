package no_var

import (
	"github.com/microsoft/typescript-go/shim/ast"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-var
var NoVarRule = rule.Rule{
	Name: "no-var",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				// BlockScoped = Let | Const | Using | AwaitUsing
				// If none of those flags are set, it's a var declaration.
				if node.Flags&ast.NodeFlagsBlockScoped != 0 {
					return
				}

				// Skip var inside `declare global { var ... }` (TypeScript ambient context)
				if isInDeclareGlobal(node) {
					return
				}

				// Report on the VariableStatement parent to include `declare` in the range,
				// but not when exported (ESLint excludes the `export` keyword from the range).
				// For for-loop initializers (no VariableStatement parent), report on self.
				reportNode := node
				if node.Parent != nil && node.Parent.Kind == ast.KindVariableStatement &&
					!ast.HasSyntacticModifier(node.Parent, ast.ModifierFlagsExport) {
					reportNode = node.Parent
				}

				msg := rule.RuleMessage{
					Id:          "unexpectedVar",
					Description: "Unexpected var, use let or const instead.",
				}

				if ctx.TypeChecker != nil && canFix(node, &ctx) {
					varRange := utils.GetVarKeywordRange(node, ctx.SourceFile)
					ctx.ReportNodeWithFixes(reportNode, msg,
						rule.RuleFixReplaceRange(varRange, "let"))
				} else {
					ctx.ReportNode(reportNode, msg)
				}
			},
		}
	},
}

// isInDeclareGlobal checks if a node is inside a `declare global { }` block.
func isInDeclareGlobal(node *ast.Node) bool {
	return ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		return ast.IsGlobalScopeAugmentation(n)
	}) != nil
}

// ---------- canFix: determines if var→let is safe ----------

// canFix checks all 10 ESLint conditions to determine if var→let is safe.
func canFix(node *ast.Node, ctx *rule.RuleContext) bool {
	declList := node.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return false
	}

	// The statement node (VariableStatement for standalone, ForStatement etc. for loops)
	stmtNode := node.Parent

	// Condition 1: var inside switch case — `let` in case without braces is confusing
	if stmtNode != nil && stmtNode.Parent != nil &&
		(stmtNode.Parent.Kind == ast.KindCaseClause || stmtNode.Parent.Kind == ast.KindDefaultClause) {
		return false
	}

	// Condition 10: var in statement position where `let` is syntactically invalid
	// e.g. `if (foo) var x = 1;` — `let` is not allowed as a bare statement
	if !isInValidLetPosition(node) {
		return false
	}

	// Collect all variable names declared in this var statement
	var vars []varInfo
	for _, decl := range declList.Declarations.Nodes {
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil || varDecl.Name() == nil {
			continue
		}
		utils.CollectBindingNames(varDecl.Name(), func(ident *ast.Node, _ string) {
			sym := ctx.TypeChecker.GetSymbolAtLocation(ident)
			if sym != nil {
				vars = append(vars, varInfo{nameNode: ident, sym: sym})
			}
		})
	}

	if len(vars) == 0 {
		return false
	}

	// Collect all references in the enclosing scope
	scope := findEnclosingScope(node)
	if scope == nil {
		scope = ctx.SourceFile.AsNode()
	}
	refs := collectReferences(scope, vars, ctx)

	// Condition 2: self-reference or forward-reference in TDZ
	// Uses positional range checks (matching ESLint's approach).
	if hasTDZIssue(declList, vars, refs) {
		return false
	}

	for _, v := range vars {
		// Condition 3: global scope variable (script mode, not module)
		if isGlobalVar(v.nameNode, ctx) {
			return false
		}

		// Condition 4: redeclared variable
		if isRedeclared(v.sym) {
			return false
		}

		// Condition 6: variable name is `let`
		if v.nameNode.Text() == "let" {
			return false
		}

		varRefs := refs[v.sym]

		// Condition 5: used from outside the block scope
		if isUsedFromOutsideScope(node, varRefs) {
			return false
		}

		// Condition 7: referenced before declaration (hoisting)
		if hasReferenceBeforeDeclaration(v.nameNode, varRefs) {
			return false
		}
	}

	// Conditions 8 & 9: loop-specific checks
	if isInLoop(node) {
		for _, v := range vars {
			// Condition 8: referenced in a closure within the loop
			if isReferencedInClosure(v.sym, node, refs[v.sym]) {
				return false
			}
		}

		// Condition 9: uninitialized declaration in a loop
		if !isLoopAssignee(node) && !isDeclarationFullyInitialized(declList) {
			return false
		}
	}

	return true
}

// ---------- Condition helpers ----------

// Condition 2: hasTDZIssue uses positional range checks (matching ESLint) to detect
// references that would cause a Temporal Dead Zone error with `let`.
func hasTDZIssue(declList *ast.VariableDeclarationList, vars []varInfo, refs map[*ast.Symbol][]*ast.Node) bool {
	if declList.Declarations == nil {
		return false
	}

	for _, decl := range declList.Declarations.Nodes {
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil {
			continue
		}

		// Check initializer: if init is NOT a function expression/arrow, any reference
		// to a same-declaration binding within the init range is a TDZ issue.
		// This catches `var a = a`, `var foo = (function(){ foo(); })()`, etc.
		if varDecl.Initializer != nil && !isFunctionNode(varDecl.Initializer) {
			initStart := varDecl.Initializer.Pos()
			initEnd := varDecl.Initializer.End()
			for _, v := range vars {
				for _, ref := range refs[v.sym] {
					refPos := ref.Pos()
					if refPos >= initStart && refPos < initEnd {
						return true
					}
				}
			}
		}
	}

	// Check default values: a reference to a binding within its OWN default value
	// is a TDZ issue. e.g. `var {a = a} = {}`.
	for _, v := range vars {
		defaultRange := getDefaultValueRange(v.nameNode)
		if defaultRange == nil {
			continue
		}
		defStart := defaultRange.Pos()
		defEnd := defaultRange.End()
		for _, ref := range refs[v.sym] {
			refPos := ref.Pos()
			if refPos >= defStart && refPos < defEnd {
				return true
			}
		}
	}

	return false
}

// getDefaultValueRange returns the default value initializer node for a binding
// inside a destructuring pattern, or nil if there is no default.
func getDefaultValueRange(nameNode *ast.Node) *ast.Node {
	if nameNode.Parent == nil {
		return nil
	}
	if nameNode.Parent.Kind == ast.KindBindingElement {
		be := nameNode.Parent.AsBindingElement()
		if be != nil && be.Initializer != nil {
			return be.Initializer
		}
	}
	return nil
}

// isFunctionNode checks if the node is a function/arrow expression (deferred execution).
func isFunctionNode(node *ast.Node) bool {
	return node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction
}

// Condition 3: isGlobalVar checks if a variable is in the global scope (script mode).
func isGlobalVar(nameNode *ast.Node, ctx *rule.RuleContext) bool {
	// If the source file is a module, top-level vars are module-scoped, not global
	if ctx.SourceFile.ExternalModuleIndicator != nil {
		return false
	}
	// Check if the declaration is at the source file top level (no enclosing function)
	enclosing := findEnclosingScope(nameNode)
	return enclosing != nil && enclosing.Kind == ast.KindSourceFile
}

// Condition 4: isRedeclared checks if a symbol has multiple declarations.
func isRedeclared(sym *ast.Symbol) bool {
	return len(sym.Declarations) >= 2
}

// Condition 5: isUsedFromOutsideScope checks if any reference is positionally
// outside the block scope containing the declaration.
func isUsedFromOutsideScope(declListNode *ast.Node, refs []*ast.Node) bool {
	scopeNode := findBlockScope(declListNode)
	if scopeNode == nil {
		return false
	}
	scopeStart := scopeNode.Pos()
	scopeEnd := scopeNode.End()
	for _, ref := range refs {
		refPos := ref.Pos()
		if refPos < scopeStart || refPos >= scopeEnd {
			return true
		}
	}
	return false
}

// Condition 7: hasReferenceBeforeDeclaration checks if any reference appears
// before the variable's declaration position (relies on var hoisting).
func hasReferenceBeforeDeclaration(nameNode *ast.Node, refs []*ast.Node) bool {
	declPos := nameNode.Pos()
	for _, ref := range refs {
		if ref.Pos() < declPos && ref != nameNode {
			return true
		}
	}
	return false
}

// Condition 8: isReferencedInClosure checks if any reference is from a different
// function scope than the variable (closure captures in a loop).
func isReferencedInClosure(sym *ast.Symbol, declNode *ast.Node, refs []*ast.Node) bool {
	declFuncScope := findEnclosingScope(declNode)
	for _, ref := range refs {
		refFuncScope := findEnclosingScope(ref)
		if refFuncScope != declFuncScope {
			return true
		}
	}
	return false
}

// Condition 9: isDeclarationFullyInitialized checks if every declarator has an initializer.
func isDeclarationFullyInitialized(declList *ast.VariableDeclarationList) bool {
	if declList.Declarations == nil {
		return false
	}
	for _, decl := range declList.Declarations.Nodes {
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil || varDecl.Initializer == nil {
			return false
		}
	}
	return true
}

// Condition 10: isInValidLetPosition checks if the declaration is in a position
// where `let` is syntactically valid.
func isInValidLetPosition(node *ast.Node) bool {
	if isLoopAssignee(node) {
		return true
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	// For-loop initializer
	if parent.Kind == ast.KindForStatement {
		return true
	}
	// VariableStatement — check its parent
	if parent.Kind == ast.KindVariableStatement {
		grandparent := parent.Parent
		if grandparent == nil {
			return false
		}
		switch grandparent.Kind {
		case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock,
			ast.KindCaseClause, ast.KindDefaultClause:
			return true
		}
		return false
	}
	return false
}

// ---------- Shared helpers ----------

// isInLoop checks if a node is inside a loop body.
func isInLoop(node *ast.Node) bool {
	current := node.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindWhileStatement, ast.KindDoStatement:
			return true
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindSourceFile:
			return false
		}
		current = current.Parent
	}
	return false
}

// isLoopAssignee checks if a VariableDeclarationList is the left side of for-in/for-of.
func isLoopAssignee(node *ast.Node) bool {
	if node.Parent == nil {
		return false
	}
	return node.Parent.Kind == ast.KindForInStatement || node.Parent.Kind == ast.KindForOfStatement
}

// findEnclosingScope delegates to the public utils.FindEnclosingScope.
func findEnclosingScope(node *ast.Node) *ast.Node {
	return utils.FindEnclosingScope(node)
}

// findBlockScope finds the nearest block-level scope that contains the declaration.
// This is the scope that `let` would be confined to.
func findBlockScope(node *ast.Node) *ast.Node {
	return ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
		switch n.Kind {
		case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock,
			ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindSwitchStatement:
			return true
		}
		return false
	})
}

// collectReferences walks the enclosing scope and collects all identifier references
// to any of the given variables, grouped by symbol.
func collectReferences(scope *ast.Node, vars []varInfo, ctx *rule.RuleContext) map[*ast.Symbol][]*ast.Node {
	symSet := make(map[*ast.Symbol]bool, len(vars))
	for _, v := range vars {
		symSet[v.sym] = true
	}

	result := make(map[*ast.Symbol][]*ast.Node)
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindIdentifier {
			sym := ctx.TypeChecker.GetSymbolAtLocation(n)
			if sym != nil && symSet[sym] {
				// Skip if this is the declaration name itself
				if !ast.IsDeclarationName(n) {
					result[sym] = append(result[sym], n)
				}
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(scope)
	return result
}

type varInfo struct {
	nameNode *ast.Node
	sym      *ast.Symbol
}
