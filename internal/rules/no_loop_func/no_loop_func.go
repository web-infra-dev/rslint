package no_loop_func

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func BuildUnsafeRefsMessage(varNames string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unsafeRefs",
		Description: fmt.Sprintf("Function declared in a loop contains unsafe references to variable(s) %s.", varNames),
	}
}

type RunState struct {
	Ctx          rule.RuleContext
	SkippedIIFEs map[*ast.Node]bool
	// refIndex buckets every value-position identifier (and destructuring
	// shorthand write target) in the source file by resolved symbol. Populated
	// lazily on the first forEachReference call and reused for all subsequent
	// lookups — amortizes what would otherwise be a full-file walk per symbol
	// into a single pass per rule invocation.
	refIndex map[*ast.Symbol][]*ast.Node
}

// NewRunState constructs a RunState ready for one source-file pass.
func NewRunState(ctx rule.RuleContext) *RunState {
	return &RunState{
		Ctx:          ctx,
		SkippedIIFEs: map[*ast.Node]bool{},
	}
}

// getContainingLoopNode walks up from `node` and returns the nearest enclosing
// loop statement. Returns nil if no loop is encountered before a non-skipped
// function-like ancestor is reached.
func (s *RunState) GetContainingLoopNode(node *ast.Node) *ast.Node {
	for current := node; current.Parent != nil; current = current.Parent {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindWhileStatement, ast.KindDoStatement:
			return parent
		case ast.KindForStatement:
			forStmt := parent.AsForStatement()
			if forStmt != nil && forStmt.Initializer != current {
				return parent
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := parent.AsForInOrOfStatement()
			if stmt != nil && stmt.Expression != current {
				return parent
			}
		case ast.KindArrowFunction, ast.KindFunctionExpression, ast.KindFunctionDeclaration,
			ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindConstructor:
			// NOTE: ClassStaticBlockDeclaration is intentionally NOT a boundary —
			// ESLint walks through static blocks to the outer loop, because a
			// static block runs once per class instantiation and each iteration
			// creates a new class, so functions defined inside a static block
			// inside a loop still exhibit the "closure per iteration" problem.
			if s.SkippedIIFEs[parent] {
				continue
			}
			return nil
		}
	}
	return nil
}

// getTopLoopNode walks up through containing loops until it finds the outermost
// loop whose start is not before `excludedNode`'s end. If `excludedNode` is nil,
// walks to the outermost loop.
func (s *RunState) GetTopLoopNode(node *ast.Node, excludedNode *ast.Node) *ast.Node {
	border := 0
	if excludedNode != nil {
		border = utils.TrimNodeTextRange(s.Ctx.SourceFile, excludedNode).End()
	}
	outermost := node
	containing := node
	for containing != nil {
		pos := utils.TrimNodeTextRange(s.Ctx.SourceFile, containing).Pos()
		if pos < border {
			break
		}
		outermost = containing
		containing = s.GetContainingLoopNode(containing)
	}
	return outermost
}

// isIIFE reports whether `node` is a function directly invoked as the callee
// of a call expression (e.g. `(function () {})()`). The function may be
// wrapped in one or more parenthesized expressions.
func IsIIFE(node *ast.Node) bool {
	if node.Kind != ast.KindFunctionExpression && node.Kind != ast.KindArrowFunction {
		return false
	}
	outer := ast.WalkUpParenthesizedExpressions(node.Parent)
	if outer == nil || outer.Kind != ast.KindCallExpression {
		return false
	}
	// WalkUpParenthesizedExpressions ascends to the first non-paren ancestor,
	// so `outer` is the CallExpression. The function is the callee iff the
	// CallExpression's Expression resolves back to `node` (possibly through
	// ParenthesizedExpression wrappers).
	call := outer.AsCallExpression()
	if call == nil {
		return false
	}
	return ast.SkipParentheses(call.Expression) == node
}

// isAsyncOrGenerator reports whether a function-like node is declared `async`
// or a generator (`function*` / `*m()`).
func IsAsyncOrGenerator(node *ast.Node) bool {
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync) {
		return true
	}
	switch node.Kind {
	case ast.KindFunctionExpression:
		if fn := node.AsFunctionExpression(); fn != nil && fn.AsteriskToken != nil {
			return true
		}
	case ast.KindFunctionDeclaration:
		if fn := node.AsFunctionDeclaration(); fn != nil && fn.AsteriskToken != nil {
			return true
		}
	case ast.KindMethodDeclaration:
		if m := node.AsMethodDeclaration(); m != nil && m.AsteriskToken != nil {
			return true
		}
	}
	return false
}

// getFunctionBodyRoots returns the subtrees of a function-like node to scan
// for identifier references — parameters and body. Excludes the function's
// own name node so self-references inside the body are picked up normally.
func getFunctionBodyRoots(node *ast.Node) []*ast.Node {
	var roots []*ast.Node
	for _, param := range node.Parameters() {
		if param != nil {
			roots = append(roots, param)
		}
	}
	if body := node.Body(); body != nil {
		roots = append(roots, body)
	}
	return roots
}

// ReferenceEntry captures a single identifier occurrence and its resolved
// variable symbol, preserved in source order for first-seen deduplication.
type ReferenceEntry struct {
	name   string
	node   *ast.Node
	symbol *ast.Symbol
}

// Name returns the textual name of the identifier reference.
func (r ReferenceEntry) Name() string { return r.name }

// Node returns the identifier AST node.
func (r ReferenceEntry) Node() *ast.Node { return r.node }

// Symbol returns the resolved symbol for the reference, or nil if unresolved.
func (r ReferenceEntry) Symbol() *ast.Symbol { return r.symbol }

// collectThroughReferences walks the function-like node's parameters and body
// and returns all identifier references whose resolved symbol has at least one
// declaration outside the function subtree. The function's own name node is
// excluded from the walk (it's the declaration, not a reference).
func (s *RunState) CollectThroughReferences(funcNode *ast.Node) []ReferenceEntry {
	var refs []ReferenceEntry
	nameNode := funcNode.Name()

	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		// Skip the function's own name node.
		if n == nameNode {
			return
		}
		// Skip type annotation subtrees entirely.
		if ast.IsPartOfTypeNode(n) {
			return
		}
		// Skip type-only children (type parameters, interface bodies, etc.).
		switch n.Kind {
		case ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration,
			ast.KindTypeParameter, ast.KindTypeReference:
			return
		}

		if n.Kind == ast.KindIdentifier && isValueReferencePosition(n) {
			// Shorthand property assignments dual-purpose the identifier as
			// the property name AND a value reference to a same-named
			// variable. GetSymbolAtLocation returns the property symbol;
			// GetShorthandAssignmentValueSymbol returns the variable symbol.
			// This applies in BOTH destructuring (`({port} = obj)`, where
			// the variable is being written) and object literal value
			// position (`f({port})`, where the variable is being read).
			var sym *ast.Symbol
			if n.Parent != nil && n.Parent.Kind == ast.KindShorthandPropertyAssignment {
				sym = s.Ctx.TypeChecker.GetShorthandAssignmentValueSymbol(n.Parent)
			} else {
				sym = s.Ctx.TypeChecker.GetSymbolAtLocation(n)
			}
			if sym != nil && (sym.Flags&ast.SymbolFlagsValue) != 0 {
				if !isSymbolDeclaredInside(sym, funcNode) {
					refs = append(refs, ReferenceEntry{
						name:   n.Text(),
						node:   n,
						symbol: sym,
					})
				}
			}
		}

		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}

	for _, root := range getFunctionBodyRoots(funcNode) {
		walk(root)
	}
	return refs
}

// isValueReferencePosition reports whether an Identifier node is used as a
// value reference (readable or writable), as opposed to a property name,
// object literal key, or member access right-hand side.
func isValueReferencePosition(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return true
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// foo.bar — `bar` is a property name, not a variable.
		pa := parent.AsPropertyAccessExpression()
		if pa != nil && pa.Name() == node {
			return false
		}
	case ast.KindQualifiedName:
		qn := parent.AsQualifiedName()
		if qn != nil && qn.Right == node {
			return false
		}
	case ast.KindMetaProperty:
		return false
	case ast.KindPropertyAssignment:
		// { key: value } — the `key` side is a property name.
		pa := parent.AsPropertyAssignment()
		if pa != nil && pa.Name() == node {
			return false
		}
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindPropertyDeclaration, ast.KindPropertySignature,
		ast.KindMethodSignature, ast.KindEnumMember:
		// Member/enum key position.
		if parent.Name() == node {
			return false
		}
	case ast.KindImportSpecifier:
		is := parent.AsImportSpecifier()
		if is != nil && is.PropertyName == node {
			return false
		}
	case ast.KindExportSpecifier:
		es := parent.AsExportSpecifier()
		if es != nil && es.PropertyName == node {
			return false
		}
	case ast.KindLabeledStatement:
		if parent.Name() == node {
			return false
		}
	case ast.KindBreakStatement, ast.KindContinueStatement:
		// break/continue label — not a variable.
		return false
	}
	return true
}

// isSymbolDeclaredInside reports whether any declaration of `sym` resolves to
// a parameter or body binding of `funcNode`. The function's own name node is
// intentionally NOT treated as "inside": ESLint models a named function
// expression's name in an outer name scope, so self-references like
// `function f() { f }` leak through the function's scope. The TypeChecker
// records the "name binding" either as the name identifier or the function
// node itself, depending on the declaration kind, so both are excluded.
func isSymbolDeclaredInside(sym *ast.Symbol, funcNode *ast.Node) bool {
	if sym == nil {
		return false
	}
	nameNode := funcNode.Name()
	for _, decl := range sym.Declarations {
		if decl == funcNode || (nameNode != nil && decl == nameNode) {
			continue
		}
		for n := decl; n != nil; n = n.Parent {
			if n == funcNode {
				return true
			}
		}
	}
	return false
}

// GetVarDeclListKind returns the kind of a VariableDeclarationList: one of
// "const", "let", "var", "using", "await using", or "" for anything else.
//
// IMPORTANT: tsgo encodes `await using` as `NodeFlagsConst | NodeFlagsUsing`
// (see the node-flags definitions in typescript-go's ast package), so the aggregate flag
// `NodeFlagsAwaitUsing` is NOT a single bit. We must inspect the individual
// `NodeFlagsUsing` and `NodeFlagsConst` bits and only report `"await using"`
// when BOTH are set, otherwise `flags & NodeFlagsAwaitUsing != 0` would
// collapse plain `const` and plain `using` into `"await using"`.
func GetVarDeclListKind(declList *ast.Node) string {
	if declList == nil || declList.Kind != ast.KindVariableDeclarationList {
		return ""
	}
	flags := declList.Flags
	hasUsing := flags&ast.NodeFlagsUsing != 0
	hasConst := flags&ast.NodeFlagsConst != 0
	switch {
	case hasUsing && hasConst:
		return "await using"
	case hasUsing:
		return "using"
	case hasConst:
		return "const"
	case flags&ast.NodeFlagsLet != 0:
		return "let"
	default:
		return "var"
	}
}

// enclosingVarDeclOfBindingElement walks up through nested BindingElement /
// BindingPattern layers to the containing VariableDeclaration, or returns nil
// if the binding does not ultimately belong to a VariableDeclaration.
func enclosingVarDeclOfBindingElement(bindingElement *ast.Node) *ast.Node {
	if bindingElement == nil || bindingElement.Kind != ast.KindBindingElement {
		return nil
	}
	parent := ast.WalkUpBindingElementsAndPatterns(bindingElement)
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return nil
	}
	return parent
}

// isWriteRef reports whether an identifier participates in a write to its
// variable. Extends utils.IsWriteReference with the cases ESLint's scope
// manager marks as writes but we don't otherwise detect: the binding names of
// `var/let/const` declarations with initializers, the bindings introduced
// by `for (var/let/const ... in/of ...)` iteration, and catch-clause
// parameters (bound anew per thrown exception).
func IsWriteRef(node *ast.Node) bool {
	if utils.IsWriteReference(node) {
		return true
	}
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		varDecl := parent.AsVariableDeclaration()
		if varDecl == nil || varDecl.Name() != node {
			return false
		}
		return varDeclIntroducesWrite(parent)
	case ast.KindBindingElement:
		be := parent.AsBindingElement()
		if be == nil || be.Name() != node {
			return false
		}
		varDecl := enclosingVarDeclOfBindingElement(parent)
		if varDecl == nil {
			return false
		}
		return varDeclIntroducesWrite(varDecl)
	}
	return false
}

// varDeclIntroducesWrite reports whether a VariableDeclaration's binding is
// written at its introduction: it has an initializer, is the target of a
// for-in/for-of iteration, or is a catch-clause parameter.
func varDeclIntroducesWrite(varDecl *ast.Node) bool {
	vd := varDecl.AsVariableDeclaration()
	if vd == nil {
		return false
	}
	if vd.Initializer != nil {
		return true
	}
	if isVarDeclInForInOrOf(varDecl) {
		return true
	}
	// `catch (e) {...}` binds `e` per thrown exception.
	return varDecl.Parent != nil && varDecl.Parent.Kind == ast.KindCatchClause
}

// isVarDeclInForInOrOf reports whether a VariableDeclaration sits directly
// inside a for-in/for-of initializer.
func isVarDeclInForInOrOf(varDecl *ast.Node) bool {
	if varDecl == nil || varDecl.Parent == nil {
		return false
	}
	declList := varDecl.Parent
	if declList.Kind != ast.KindVariableDeclarationList || declList.Parent == nil {
		return false
	}
	outer := declList.Parent
	return outer.Kind == ast.KindForInStatement || outer.Kind == ast.KindForOfStatement
}

// isSafe reports whether a through-reference `ref` to a symbol `sym` is safe
// with respect to a loop node `loopNode`. Safe means: no write to `sym` can
// modify the function's closed-over view of it during successive iterations.
func (s *RunState) isSafeCore(loopNode *ast.Node, ref ReferenceEntry) bool {
	sym := ref.symbol
	if sym == nil || len(sym.Declarations) == 0 {
		return true
	}
	decl := sym.Declarations[0]

	// Look up the enclosing VariableDeclarationList (for var/let/const/using).
	declList := GetDeclListForSymbolDecl(decl)
	kind := GetVarDeclListKind(declList)

	// Constant bindings never get rewritten, so they're safe.
	if kind == "const" || kind == "using" || kind == "await using" {
		return true
	}

	sf := s.Ctx.SourceFile
	loopRange := utils.TrimNodeTextRange(sf, loopNode)

	// `let` declared inside the loop gets a fresh binding per iteration.
	if kind == "let" && declList != nil {
		declRange := utils.TrimNodeTextRange(sf, declList)
		if declRange.Pos() > loopRange.Pos() && declRange.End() < loopRange.End() {
			return true
		}
	}

	var excluded *ast.Node
	if kind == "let" {
		excluded = declList
	}
	top := s.GetTopLoopNode(loopNode, excluded)
	border := utils.TrimNodeTextRange(sf, top).Pos()

	// The variable's "variable scope" — the nearest function-like scope of its
	// declaration. Used to tell whether a write happens in the same function
	// or inside a nested function.
	varScope := utils.FindEnclosingScope(decl)

	safe := true
	s.ForEachReference(sym, func(refNode *ast.Node) bool {
		if !IsWriteRef(refNode) {
			return false
		}
		refScope := utils.FindEnclosingScope(refNode)
		if refScope == varScope {
			refPos := utils.TrimNodeTextRange(sf, refNode).Pos()
			if refPos < border {
				return false
			}
		}
		safe = false
		return true
	})
	return safe
}

// getDeclListForSymbolDecl returns the VariableDeclarationList associated with
// a declaration node, or nil if the declaration is not a variable-like binding.
func GetDeclListForSymbolDecl(decl *ast.Node) *ast.Node {
	if decl == nil {
		return nil
	}
	current := decl
	for current != nil {
		if current.Kind == ast.KindVariableDeclarationList {
			return current
		}
		if current.Kind == ast.KindVariableDeclaration ||
			current.Kind == ast.KindBindingElement ||
			current.Kind == ast.KindObjectBindingPattern ||
			current.Kind == ast.KindArrayBindingPattern {
			current = current.Parent
			continue
		}
		return nil
	}
	return nil
}

// buildRefIndex performs a single pass over the source file and groups every
// value-position identifier by its resolved symbol. ShorthandPropertyAssignment
// nodes need special handling because TypeChecker.GetSymbolAtLocation resolves
// the shorthand key to the property symbol, not the variable symbol — we use
// GetShorthandAssignmentValueSymbol instead and key by the variable symbol.
// This is needed BOTH for destructuring shorthand (which is a write) and for
// object-literal shorthand (which is a read).
func (s *RunState) buildRefIndex() {
	if s.refIndex != nil {
		return
	}
	s.refIndex = map[*ast.Symbol][]*ast.Node{}
	tc := s.Ctx.TypeChecker
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		if n.Kind == ast.KindIdentifier && isValueReferencePosition(n) {
			var sym *ast.Symbol
			if n.Parent != nil && n.Parent.Kind == ast.KindShorthandPropertyAssignment {
				sym = tc.GetShorthandAssignmentValueSymbol(n.Parent)
			} else {
				sym = tc.GetSymbolAtLocation(n)
			}
			if sym != nil {
				s.refIndex[sym] = append(s.refIndex[sym], n)
			}
		} else if n.Kind == ast.KindShorthandPropertyAssignment && utils.IsInDestructuringAssignment(n) {
			// Already handled above via the Identifier branch in most cases,
			// but keep this explicit indexing for the destructuring write
			// path where the parent walk may not visit the Identifier directly.
			if sym := tc.GetShorthandAssignmentValueSymbol(n); sym != nil {
				if nameNode := n.AsShorthandPropertyAssignment().Name(); nameNode != nil {
					s.refIndex[sym] = append(s.refIndex[sym], nameNode)
				}
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(s.Ctx.SourceFile.AsNode())
}

// forEachReference invokes `cb` for every identifier node resolving to `sym`,
// in source order. Returns early when `cb` returns true. The underlying index
// is built on first use via buildRefIndex so subsequent calls are O(refs).
func (s *RunState) ForEachReference(sym *ast.Symbol, cb func(*ast.Node) bool) {
	s.buildRefIndex()
	for _, node := range s.refIndex[sym] {
		if cb(node) {
			return
		}
	}
}

// checkForLoops processes a function-like node: if it is inside a loop and
// has unsafe through-references, it is reported.
func (s *RunState) checkForLoopsCore(node *ast.Node) {
	loopNode := s.GetContainingLoopNode(node)
	if loopNode == nil {
		return
	}

	refs := s.CollectThroughReferences(node)

	// IIFE handling — matches ESLint: non-async, non-generator IIFEs that are
	// not self-referenced (either anonymous or whose name isn't used inside the
	// function body) are skipped. Skipping marks them so nested functions can
	// walk through them to find the enclosing loop.
	if !IsAsyncOrGenerator(node) && IsIIFE(node) {
		isFunctionExpression := node.Kind == ast.KindFunctionExpression
		name := node.Name()
		isFunctionReferenced := false
		if isFunctionExpression && name != nil {
			refName := name.Text()
			for _, r := range refs {
				if r.name == refName {
					isFunctionReferenced = true
					break
				}
			}
		}
		if !isFunctionReferenced {
			s.SkippedIIFEs[node] = true
			return
		}
	}

	seen := map[string]bool{}
	var names []string
	for _, r := range refs {
		if r.symbol == nil {
			continue
		}
		if seen[r.name] {
			continue
		}
		if s.isSafeCore(loopNode, r) {
			continue
		}
		seen[r.name] = true
		names = append(names, r.name)
	}

	if len(names) == 0 {
		return
	}

	varNames := "'" + strings.Join(names, "', '") + "'"
	s.Ctx.ReportNode(node, BuildUnsafeRefsMessage(varNames))
}

// NoLoopFuncRule disallows function declarations that contain unsafe
// references to variable(s) inside loop statements.
var NoLoopFuncRule = rule.Rule{
	Name:             "no-loop-func",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// Defense-in-depth: RequiresTypeInfo: true filters this rule out for
		// gap files / inferred-project files, but if a future caller bypasses
		// the filter we still want to no-op rather than nil-deref.
		if ctx.TypeChecker == nil {
			return rule.RuleListeners{}
		}
		s := &RunState{
			Ctx:          ctx,
			SkippedIIFEs: map[*ast.Node]bool{},
		}
		check := func(node *ast.Node) {
			s.checkForLoopsCore(node)
		}
		return rule.RuleListeners{
			ast.KindArrowFunction:       check,
			ast.KindFunctionExpression:  check,
			ast.KindFunctionDeclaration: check,
			// NOTE: ESLint's ESTree represents class/object methods, getters,
			// setters, and constructors as a FunctionExpression value under the
			// property/method node — so its FunctionExpression listener catches
			// them. In tsgo these are distinct kinds; listen explicitly so we
			// match ESLint's behavior.
			ast.KindMethodDeclaration: check,
			ast.KindGetAccessor:       check,
			ast.KindSetAccessor:       check,
			ast.KindConstructor:       check,
		}
	},
}
