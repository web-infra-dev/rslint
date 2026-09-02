package no_var

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// https://eslint.org/docs/latest/rules/no-var
var NoVarRule = rule.Rule{
	Name:   "no-var",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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

				msg := rule.RuleMessage{
					Id:          "unexpectedVar",
					Description: "Unexpected var, use let or const instead.",
				}

				ctx.ReportRangeWithDeferredFixes(noVarReportRange(node, ctx.SourceFile), msg, func() []rule.RuleFix {
					if !canFix(node, &ctx) {
						return nil
					}
					varRange := utils.GetVarKeywordRange(node, ctx.SourceFile)
					return []rule.RuleFix{rule.RuleFixReplaceRange(varRange, "let")}
				})
			},
		}
	},
}

// noVarReportRange mirrors the range of ESLint's VariableDeclaration node.
// tsgo stores export/declare as VariableStatement modifiers and excludes the
// trailing semicolon from the nested VariableDeclarationList, while ESTree
// wraps only `export` and keeps every other modifier plus the semicolon on the
// declaration itself.
func noVarReportRange(node *ast.Node, sourceFile *ast.SourceFile) core.TextRange {
	statement := node.Parent
	if statement == nil || statement.Kind != ast.KindVariableStatement {
		return utils.TrimNodeTextRange(sourceFile, node)
	}

	modifiers := statement.Modifiers()
	if modifiers == nil || !ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport) {
		return utils.TrimNodeTextRange(sourceFile, statement)
	}

	start := utils.TrimNodeTextRange(sourceFile, node).Pos()
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind != ast.KindExportKeyword {
			start = utils.TrimNodeTextRange(sourceFile, modifier).Pos()
			break
		}
	}
	return core.NewTextRange(start, statement.End())
}

// isInDeclareGlobal checks whether the declaration is directly contained by a
// `declare global { }` block. Nested namespaces are separate TSModuleBlocks and
// are reported by upstream.
func isInDeclareGlobal(node *ast.Node) bool {
	statement := node.Parent
	if statement == nil || statement.Kind != ast.KindVariableStatement {
		return false
	}
	moduleBlock := statement.Parent
	return moduleBlock != nil && moduleBlock.Kind == ast.KindModuleBlock &&
		moduleBlock.Parent != nil && ast.IsGlobalScopeAugmentation(moduleBlock.Parent)
}

// ---------- canFix: determines if var→let is safe ----------

// canFix applies ESLint's safety conditions plus TypeScript/runtime cases where
// upstream's scope approximation would change behavior.
func canFix(node *ast.Node, ctx *rule.RuleContext) bool {
	if ctx.Refs == nil {
		return false
	}
	declList := node.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return false
	}

	// The statement node (VariableStatement for standalone, ForStatement etc. for loops)
	stmtNode := node.Parent
	// ESLint sees an exported declaration through an ExportNamedDeclaration
	// wrapper, which is not a statement-list parent and therefore is never
	// autofixed by this rule.
	if stmtNode != nil && stmtNode.Kind == ast.KindVariableStatement &&
		ast.HasSyntacticModifier(stmtNode, ast.ModifierFlagsExport) {
		return false
	}

	// Condition 1: var inside switch case — `let` in case without braces is confusing
	if stmtNode != nil && stmtNode.Kind == ast.KindVariableStatement && stmtNode.Parent != nil &&
		(stmtNode.Parent.Kind == ast.KindCaseClause || stmtNode.Parent.Kind == ast.KindDefaultClause) {
		return false
	}

	// Condition 10: var in statement position where `let` is syntactically invalid
	// e.g. `if (foo) var x = 1;` — `let` is not allowed as a bare statement
	if !isInValidLetPosition(node) {
		return false
	}

	// Collect all variable names declared in this var statement.
	var vars []varInfo
	hasUnresolvedBinding := false
	for _, decl := range declList.Declarations.Nodes {
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil || varDecl.Name() == nil {
			continue
		}
		utils.CollectBindingNames(varDecl.Name(), func(ident *ast.Node, _ string) {
			sym := utils.BindingNameSymbol(ident)
			if sym == nil {
				hasUnresolvedBinding = true
				return
			}
			vars = append(vars, varInfo{
				nameNode: ident,
				sym:      sym,
			})
		})
	}
	if hasUnresolvedBinding {
		return false
	}

	for _, v := range vars {
		varScope := findEnclosingScope(v.nameNode)
		// Condition 3: global scope variable (script mode, not module)
		if isGlobalVar(varScope, ctx) {
			return false
		}

		// Condition 4: redeclared variable
		if isRedeclared(v.sym) || isRedeclaredByFunction(v.nameNode, v.sym, varScope) {
			return false
		}

		// Condition 6: variable name is `let`
		if v.nameNode.Text() == "let" {
			return false
		}

		// A var may reuse a simple catch parameter, but let may not. In a
		// nested block let would instead create a different binding.
		if isShadowedByCatchParameter(v.nameNode, varScope) {
			return false
		}
	}

	inLoop := isInLoop(node)
	for _, v := range vars {
		// Collect references only after the cheap blockers above. A reference
		// can only resolve to this symbol from within its variable scope, so the
		// file-wide list is identical to a bounded scope walk.
		refs := filterSafeNamespaceExportReferences(ctx.Refs.References(v.sym))
		refs = filterSafeAmbientNamespaceReferences(v.nameNode, refs)

		// Condition 2: self-reference in TDZ
		if hasTDZIssue(v.nameNode, refs) {
			return false
		}

		// Condition 5: used from outside the block scope
		if isUsedFromOutsideScope(node, refs) {
			return false
		}

		// Condition 7: referenced before declaration (hoisting)
		if hasReferenceBeforeDeclaration(v.nameNode, refs) {
			return false
		}

		// A hoisted function declaration can execute a later var reference
		// before let would be initialized, even when the reference itself is
		// textually after the declaration.
		if hasUnsafeHoistedFunctionReference(v.nameNode, refs, ctx) {
			return false
		}

		// Condition 8: referenced in a closure within the loop
		if inLoop && isReferencedInClosure(node, refs) {
			return false
		}
	}

	// Condition 9: uninitialized declaration in a loop
	if inLoop {
		if !isLoopAssignee(node) && !isDeclarationFullyInitialized(declList) {
			return false
		}
	}

	return true
}

// ---------- Condition helpers ----------

// Condition 2: hasTDZIssue reuses the shared initializer walk to detect
// references that would enter the temporal dead zone with `let`.
func hasTDZIssue(nameNode *ast.Node, refs []*ast.Node) bool {
	var declaration *ast.VariableDeclaration
	for _, ref := range refs {
		if isTypeOnlyReferenceLocation(ref) ||
			!scope.IsInsideOwnInitializer(nameNode, ref.End()) {
			continue
		}

		// A function expression used directly as the declarator's value is
		// evaluated only after the binding has been initialized. Defaults in
		// the binding pattern itself are still immediate and stay unsafe.
		if declaration == nil {
			root := ast.GetRootDeclaration(nameNode.Parent)
			if root != nil {
				declaration = root.AsVariableDeclaration()
			}
		}
		if isInDirectFunctionInitializer(declaration, ref) {
			continue
		}
		return true
	}

	return false
}

func isInDirectFunctionInitializer(declaration *ast.VariableDeclaration, reference *ast.Node) bool {
	return declaration != nil && declaration.Initializer != nil && isFunctionNode(declaration.Initializer) &&
		reference.Pos() >= declaration.Initializer.Pos() && reference.End() <= declaration.Initializer.End()
}

// isFunctionNode checks if the runtime expression is a function/arrow
// expression (deferred execution). Parentheses and TypeScript assertions do
// not change when the function body runs.
func isFunctionNode(node *ast.Node) bool {
	node = ast.SkipOuterExpressions(node, ast.OEKParentheses|ast.OEKAssertions)
	if node == nil {
		return false
	}
	return node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction
}

// Condition 3: isGlobalVar checks if a variable is in the global scope (script mode).
func isGlobalVar(varScope *ast.Node, ctx *rule.RuleContext) bool {
	// A syntactic module or a file with module/CommonJS language defaults has a
	// non-global top-level scope. Omitted sourceType gets ESLint's module default
	// even when the file contains no import/export marker.
	if ctx.Refs.HasNonGlobalProgramScope() {
		return false
	}
	return varScope != nil && varScope.Kind == ast.KindSourceFile
}

// Condition 4: isRedeclared checks if a symbol has multiple declarations.
func isRedeclared(sym *ast.Symbol) bool {
	return sym != nil && len(sym.Declarations) >= 2
}

// isRedeclaredByFunction covers the one legal value-space redeclaration that
// tsgo's binder keeps as a separate symbol: a var sharing its variable scope
// with a function declaration. Replacing the var with let would either be a
// syntax error or introduce a new binding. The binder-owned scope tables let
// us detect it without reconstructing the file's scopes.
func isRedeclaredByFunction(nameNode *ast.Node, ownSymbol *ast.Symbol, scopeNode *ast.Node) bool {
	if nameNode == nil {
		return false
	}
	if scopeNode == nil {
		return false
	}
	last := scopeNode
	if scopeNode.Kind == ast.KindModuleBlock && scopeNode.Parent != nil &&
		scopeNode.Parent.Kind == ast.KindModuleDeclaration {
		last = scopeNode.Parent
	}

	name := nameNode.Text()
	for current := nameNode.Parent; current != nil; current = current.Parent {
		if symbolDeclaresFunction(current.Locals()[name], ownSymbol) {
			return true
		}
		if owner := current.Symbol(); owner != nil &&
			symbolDeclaresFunction(owner.Exports[name], ownSymbol) {
			return true
		}
		if current == last {
			break
		}
	}
	return false
}

func symbolDeclaresFunction(symbol *ast.Symbol, ownSymbol *ast.Symbol) bool {
	if symbol == nil || symbol == ownSymbol {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind == ast.KindFunctionDeclaration {
			return true
		}
	}
	return false
}

// isShadowedByCatchParameter checks only as far as the var's own variable
// scope. A nested function or class static block owns a different var, so a
// catch parameter outside that boundary cannot conflict with the replacement.
func isShadowedByCatchParameter(nameNode *ast.Node, boundary *ast.Node) bool {
	if nameNode == nil {
		return false
	}
	name := nameNode.Text()
	for current := nameNode.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindCatchClause {
			catchClause := current.AsCatchClause()
			if catchClause != nil && catchClause.VariableDeclaration != nil {
				declaration := catchClause.VariableDeclaration.AsVariableDeclaration()
				if declaration != nil && bindingContainsName(declaration.Name(), name) {
					return true
				}
			}
		}
		if current == boundary {
			break
		}
	}
	return false
}

func bindingContainsName(binding *ast.Node, target string) bool {
	found := false
	utils.CollectBindingNames(binding, func(_ *ast.Node, name string) {
		if name == target {
			found = true
		}
	})
	return found
}

// filterSafeNamespaceExportReferences removes `export as namespace Name`.
// It names the UMD module in the global namespace but does not read the local
// value at runtime, so it cannot make a var-to-let replacement unsafe.
func filterSafeNamespaceExportReferences(refs []*ast.Node) []*ast.Node {
	return slices.DeleteFunc(slices.Clone(refs), func(ref *ast.Node) bool {
		return ref != nil && ref.Parent != nil && ref.Parent.Kind == ast.KindNamespaceExportDeclaration
	})
}

// filterSafeAmbientNamespaceReferences removes type-only reads contributed by
// another declaration that reopens the same ambient namespace. Those reads
// remain valid after var becomes let, while keeping them would incorrectly
// look like uses outside the declaration's ModuleBlock. Duplicate ambient var
// declarations are still rejected earlier by the binder symbol check.
func filterSafeAmbientNamespaceReferences(nameNode *ast.Node, refs []*ast.Node) []*ast.Node {
	if len(refs) == 0 || !utils.IsInAmbientContext(nameNode) {
		return refs
	}
	var filtered []*ast.Node
	for i, ref := range refs {
		if !isTypeOnlyReferenceLocation(ref) || !isInReopenedNamespace(nameNode, ref) {
			if filtered != nil {
				filtered = append(filtered, ref)
			}
			continue
		}
		if filtered == nil {
			filtered = make([]*ast.Node, 0, len(refs)-1)
			filtered = append(filtered, refs[:i]...)
		}
	}
	if filtered == nil {
		return refs
	}
	return filtered
}

func isInReopenedNamespace(declaration *ast.Node, reference *ast.Node) bool {
	for declarationAncestor := declaration.Parent; declarationAncestor != nil; declarationAncestor = declarationAncestor.Parent {
		if declarationAncestor.Kind != ast.KindModuleDeclaration || declarationAncestor.Symbol() == nil {
			continue
		}
		for referenceAncestor := reference.Parent; referenceAncestor != nil; referenceAncestor = referenceAncestor.Parent {
			if referenceAncestor.Kind == ast.KindModuleDeclaration &&
				referenceAncestor != declarationAncestor &&
				referenceAncestor.Symbol() == declarationAncestor.Symbol() {
				return true
			}
		}
	}
	return false
}

// Condition 5: isUsedFromOutsideScope checks if any reference is positionally
// outside the block scope containing the declaration.
func isUsedFromOutsideScope(declListNode *ast.Node, refs []*ast.Node) bool {
	if len(refs) == 0 {
		return false
	}
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
		if ref.Pos() < declPos && ref != nameNode && !isTypeOnlyReferenceLocation(ref) {
			return true
		}
	}
	return false
}

// hasUnsafeHoistedFunctionReference detects a later textual reference that is
// evaluated by a hoisted function declaration called before the var. With let,
// that call would enter the temporal dead zone. Calls through parentheses and
// TypeScript expression wrappers, constructor calls, and tagged templates all
// execute the declaration and are therefore unsafe.
func hasUnsafeHoistedFunctionReference(nameNode *ast.Node, refs []*ast.Node, ctx *rule.RuleContext) bool {
	if len(refs) == 0 {
		return false
	}
	declarationScope := findEnclosingExecutionScope(nameNode)
	declarationStart := nameNode.Pos()
	for _, ref := range refs {
		if ref.Pos() < declarationStart || isTypeOnlyReferenceLocation(ref) {
			continue
		}
		for currentScope := findEnclosingExecutionScope(ref); currentScope != nil && currentScope != declarationScope; currentScope = findEnclosingExecutionScope(currentScope) {
			if currentScope.Kind != ast.KindFunctionDeclaration || currentScope.Name() == nil {
				continue
			}
			functionSymbol := utils.BindingNameSymbol(currentScope.Name())
			for _, functionRef := range ctx.Refs.References(functionSymbol) {
				if isImmediateInvocation(functionRef) &&
					(functionRef.Pos() < declarationStart || isInvocationInOwnInitializer(nameNode, functionRef)) {
					return true
				}
			}
		}
	}
	return false
}

// isInvocationInOwnInitializer catches calls such as
// `var value = readValue(); function readValue() { return value; }`. The call
// is textually after the binding name but still runs before let initializes
// it. A function/arrow used directly as the initializer stays deferred.
func isInvocationInOwnInitializer(nameNode *ast.Node, invocation *ast.Node) bool {
	if !scope.IsInsideOwnInitializer(nameNode, invocation.End()) {
		return false
	}
	root := ast.GetRootDeclaration(nameNode.Parent)
	return root == nil || !isInDirectFunctionInitializer(root.AsVariableDeclaration(), invocation)
}

func isImmediateInvocation(reference *ast.Node) bool {
	current := invocationResultExpression(reference)
	if utils.IsCallee(current) || isTaggedTemplateTag(current) {
		return true
	}

	// Function.prototype.call/apply invoke their receiver immediately. Reuse
	// the shared access helpers so dotted, computed, optional, and asserted
	// member expressions agree with the rest of the rules.
	parent := current.Parent
	if parent == nil || !ast.IsAccessExpression(parent) ||
		utils.AccessExpressionObject(parent) != current {
		return false
	}
	name, ok := utils.AccessExpressionStaticName(parent)
	return ok && (name == "call" || name == "apply") && utils.IsCallee(parent)
}

// invocationResultExpression walks only expressions whose result can be the
// referenced function itself. This covers the right-most operand of a comma
// expression and the value-producing arms of logical and conditional
// expressions without treating an ordinary member access as a function call.
func invocationResultExpression(reference *ast.Node) *ast.Node {
	current := reference
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions) {
			current = parent
			continue
		}
		if parent.Kind == ast.KindBinaryExpression {
			binary := parent.AsBinaryExpression()
			if binary != nil && binary.OperatorToken != nil &&
				((binary.OperatorToken.Kind == ast.KindCommaToken && binary.Right == current) ||
					ast.IsLogicalOrCoalescingBinaryOperator(binary.OperatorToken.Kind)) {
				current = parent
				continue
			}
		}
		if parent.Kind == ast.KindConditionalExpression {
			conditional := parent.AsConditionalExpression()
			if conditional != nil && (conditional.WhenTrue == current || conditional.WhenFalse == current) {
				current = parent
				continue
			}
		}
		break
	}
	return current
}

func isTaggedTemplateTag(node *ast.Node) bool {
	current := node
	parent := current.Parent
	for parent != nil && ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions) {
		current = parent
		parent = current.Parent
	}
	if parent == nil || parent.Kind != ast.KindTaggedTemplateExpression {
		return false
	}
	tagged := parent.AsTaggedTemplateExpression()
	return tagged != nil && tagged.Tag == current
}

// Condition 8: isReferencedInClosure checks if any reference is from a different
// function scope than the variable (closure captures in a loop).
func isReferencedInClosure(declNode *ast.Node, refs []*ast.Node) bool {
	if len(refs) == 0 {
		return false
	}
	declFuncScope := findEnclosingExecutionScope(declNode)
	for _, ref := range refs {
		// Type syntax is erased and cannot observe per-iteration bindings.
		if isTypeOnlyReferenceLocation(ref) {
			continue
		}
		refFuncScope := findEnclosingExecutionScope(ref)
		if refFuncScope != declFuncScope {
			return true
		}
	}
	return false
}

func isTypeOnlyReferenceLocation(node *ast.Node) bool {
	return ast.IsPartOfTypeNode(node) || ast.IsPartOfTypeQuery(node)
}

// findEnclosingExecutionScope uses tsgo's container walk, which already keeps
// computed keys and member/parameter decorators outside their method body.
// Static blocks and static fields run while their class is evaluated and are
// folded into the surrounding execution context. Instance field initializers
// run later, when an instance is constructed, and remain a boundary.
func findEnclosingExecutionScope(node *ast.Node) *ast.Node {
	for node != nil && node.Parent != nil {
		container := ast.GetThisContainer(node, true, false)
		if container.Kind == ast.KindSourceFile || utils.IsFunctionLikeContainer(container) {
			return container
		}
		if container.Kind == ast.KindPropertyDeclaration && !ast.HasStaticModifier(container) {
			return container
		}
		node = container
	}
	return nil
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
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsIterationStatement(current, false) {
			return true
		}
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(current) || current.Kind == ast.KindSourceFile {
			return false
		}
	}
	return false
}

// isLoopAssignee checks if a VariableDeclarationList is the left side of for-in/for-of.
func isLoopAssignee(node *ast.Node) bool {
	return node.Parent != nil && ast.IsForInOrOfStatement(node.Parent)
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

type varInfo struct {
	nameNode *ast.Node
	sym      *ast.Symbol
}
