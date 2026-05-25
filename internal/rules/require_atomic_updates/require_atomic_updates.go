package require_atomic_updates

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func messageNonAtomicUpdate(value string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nonAtomicUpdate",
		Description: fmt.Sprintf("Possible race condition: `%s` might be reassigned based on an outdated value of `%s`.", value, value),
	}
}

func messageNonAtomicObjectUpdate(value, object string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nonAtomicObjectUpdate",
		Description: fmt.Sprintf("Possible race condition: `%s` might be assigned based on an outdated state of `%s`.", value, object),
	}
}

// analysisState tracks which symbols have been read before/after an await/yield.
//
// Mirrors ESLint's per-segment freshReadVariables / outdatedReadVariables sets,
// keyed by *ast.Symbol (not name) so shadowing, imports, namespace aliasing,
// catch bindings, block-scoped let/const, and TS type wrappers are handled
// naturally — distinct symbols never collide in state regardless of their name.
type analysisState struct {
	freshReads    map[*ast.Symbol]bool
	outdatedReads map[*ast.Symbol]bool
}

func newAnalysisState() *analysisState {
	return &analysisState{
		freshReads:    make(map[*ast.Symbol]bool),
		outdatedReads: make(map[*ast.Symbol]bool),
	}
}

func (s *analysisState) clone() *analysisState {
	fresh := make(map[*ast.Symbol]bool, len(s.freshReads))
	outdated := make(map[*ast.Symbol]bool, len(s.outdatedReads))
	for k := range s.freshReads {
		fresh[k] = true
	}
	for k := range s.outdatedReads {
		outdated[k] = true
	}
	return &analysisState{freshReads: fresh, outdatedReads: outdated}
}

// merge unions another branch's state into this one (used after if/else, etc.).
func (s *analysisState) merge(other *analysisState) {
	for k := range other.freshReads {
		s.freshReads[k] = true
	}
	for k := range other.outdatedReads {
		s.outdatedReads[k] = true
	}
}

func (s *analysisState) markRead(sym *ast.Symbol) {
	if sym == nil {
		return
	}
	s.freshReads[sym] = true
	// A re-read after await captures the current value, clearing outdated.
	delete(s.outdatedReads, sym)
}

// makeOutdated moves all fresh reads to outdated (called on await/yield).
func (s *analysisState) makeOutdated() {
	for k := range s.freshReads {
		s.outdatedReads[k] = true
	}
	s.freshReads = make(map[*ast.Symbol]bool)
}

func (s *analysisState) isOutdated(sym *ast.Symbol) bool {
	return sym != nil && s.outdatedReads[sym]
}

// breakTarget captures state snapshots produced by `break` statements that
// target this loop or switch (with optional label).
type breakTarget struct {
	label  string
	states []*analysisState
}

// analyzer performs the require-atomic-updates analysis within one resumable
// (async or generator) function.
type analyzer struct {
	ctx             rule.RuleContext
	funcNode        *ast.Node
	allowProperties bool

	// Pre-computed once per function:
	//   declaredInFunc : symbols whose declaration site is inside funcNode's
	//                    subtree (including params, locals, nested block decls).
	//   parameterSymbols  : subset that are parameters of funcNode itself
	//                    (for the isMember+param escape quirk).
	//   escapedSymbols : subset referenced inside a nested function/arrow/
	//                    class-static-block — they can be observed by a
	//                    concurrent context, so writes to them are race-prone.
	declaredInFunc map[*ast.Symbol]bool
	parameterSymbols  map[*ast.Symbol]bool
	escapedSymbols map[*ast.Symbol]bool

	// breakTargets is the stack of currently-enclosing loops/switches that
	// can receive a `break` (optionally labeled). Innermost at the end.
	breakTargets []*breakTarget
}

func newAnalyzer(ctx rule.RuleContext, funcNode *ast.Node, allowProperties bool) *analyzer {
	a := &analyzer{
		ctx:             ctx,
		funcNode:        funcNode,
		allowProperties: allowProperties,
		declaredInFunc:  make(map[*ast.Symbol]bool),
		parameterSymbols:   make(map[*ast.Symbol]bool),
		escapedSymbols:  make(map[*ast.Symbol]bool),
	}
	a.collectDeclarations()
	a.collectEscapes()
	return a
}

// symbolOf returns the *ast.Symbol that `node` resolves to (or nil). The rule
// is type-aware (RequiresTypeInfo is set), so TypeChecker is guaranteed non-nil
// at runtime. Shorthand property identifiers resolve to their VALUE symbol
// (the outer binding of the same name), and symbols augmented into the global
// scope via `declare global { … }` are treated as unresolved — ESLint's
// scope analyzer leaves those references unresolved too.
func (a *analyzer) symbolOf(node *ast.Node) *ast.Symbol {
	if node == nil || a.ctx.TypeChecker == nil {
		return nil
	}
	var sym *ast.Symbol
	if node.Kind == ast.KindIdentifier && node.Parent != nil &&
		node.Parent.Kind == ast.KindShorthandPropertyAssignment &&
		node.Parent.AsShorthandPropertyAssignment().Name() == node {
		sym = a.ctx.TypeChecker.GetShorthandAssignmentValueSymbol(node.Parent)
	}
	if sym == nil {
		sym = a.ctx.TypeChecker.GetSymbolAtLocation(node)
	}
	if sym == nil {
		return nil
	}
	if symbolIsDeclareGlobal(sym) {
		return nil
	}
	return sym
}

// symbolIsDeclareGlobal reports whether every declaration of `sym` is nested
// inside a `declare global { … }` block. Such globals augment the ambient
// scope and are not visible to ESLint's scope analyzer, so the rule must
// treat references to them like unresolved identifiers.
func symbolIsDeclareGlobal(sym *ast.Symbol) bool {
	if sym == nil || len(sym.Declarations) == 0 {
		return false
	}
	for _, decl := range sym.Declarations {
		if !isInsideDeclareGlobal(decl) {
			return false
		}
	}
	return true
}

func isInsideDeclareGlobal(node *ast.Node) bool {
	for cur := node; cur != nil; cur = cur.Parent {
		if cur.Kind == ast.KindModuleDeclaration {
			name := cur.Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "global" {
				return true
			}
		}
	}
	return false
}

// collectDeclarations walks funcNode (not descending into nested functions)
// and records the symbol of every declaration name it finds.
func (a *analyzer) collectDeclarations() {
	if params := a.funcNode.Parameters(); params != nil {
		for _, param := range params {
			a.collectBindingSymbols(param.Name(), true)
		}
	}
	if body := a.funcNode.Body(); body != nil {
		a.collectDeclsIn(body)
	}
}

// collectBindingSymbols walks a binding target (Identifier or
// ObjectBindingPattern/ArrayBindingPattern) and registers each contained name's
// symbol. isParam=true only when collecting parameters of funcNode itself.
func (a *analyzer) collectBindingSymbols(nameNode *ast.Node, isParam bool) {
	if nameNode == nil {
		return
	}
	utils.CollectBindingNames(nameNode, func(ident *ast.Node, _ string) {
		if sym := a.symbolOf(ident); sym != nil {
			a.declaredInFunc[sym] = true
			if isParam {
				a.parameterSymbols[sym] = true
			}
		}
	})
}

// collectDeclsIn descends into funcNode's body skipping nested functions and
// registers every declaration's symbol. Shadowing is handled naturally by
// distinct symbols, so there is no need to track which declarations belong
// to which block.
func (a *analyzer) collectDeclsIn(node *ast.Node) {
	if node == nil {
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		if ast.IsFunctionLikeDeclaration(child) {
			return false
		}
		switch child.Kind {
		case ast.KindVariableDeclaration:
			vd := child.AsVariableDeclaration()
			if vd.Name() != nil {
				a.collectBindingSymbols(vd.Name(), false)
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration,
			ast.KindEnumDeclaration, ast.KindModuleDeclaration:
			if n := child.Name(); n != nil && n.Kind == ast.KindIdentifier {
				if sym := a.symbolOf(n); sym != nil {
					a.declaredInFunc[sym] = true
				}
			}
		case ast.KindCatchClause:
			cc := child.AsCatchClause()
			if cc.VariableDeclaration != nil {
				if vd := cc.VariableDeclaration.AsVariableDeclaration(); vd != nil && vd.Name() != nil {
					a.collectBindingSymbols(vd.Name(), false)
				}
			}
		}
		a.collectDeclsIn(child)
		return false
	})
}

// collectEscapes walks funcNode's body and marks any symbol in declaredInFunc
// that is referenced inside a nested function as "escaped" — writes to such a
// symbol can be observed by the nested function's body, so race analysis must
// treat it like a non-local variable.
func (a *analyzer) collectEscapes() {
	a.walkForEscape(a.funcNode.Body(), false)
}

func (a *analyzer) walkForEscape(node *ast.Node, inNested bool) {
	if node == nil {
		return
	}

	// Check this node if it's a value-reference identifier inside a nested
	// function (arrow expression body, method body, static block, …).
	if inNested && node.Kind == ast.KindIdentifier && !utils.IsNonReferenceIdentifier(node) {
		if sym := a.symbolOf(node); sym != nil && a.declaredInFunc[sym] {
			a.escapedSymbols[sym] = true
		}
	}

	node.ForEachChild(func(child *ast.Node) bool {
		if ast.IsFunctionLikeDeclaration(child) {
			// Walk the full function-like subtree with inNested=true so
			// every identifier inside (body expression, block, parameter
			// defaults, etc.) is scanned.
			a.walkForEscape(child, true)
			return false
		}
		a.walkForEscape(child, inNested)
		return false
	})
}

// isLocalWithoutEscape reports whether writes to `sym` can only be observed by
// code inside funcNode — i.e. the symbol is declared in funcNode and never
// referenced from a nested function scope. For member-access writes on a
// parameter, the underlying object comes from outside, so we treat the param
// as not-local (matches ESLint's isLocalVariableWithoutEscape param shortcut).
func (a *analyzer) isLocalWithoutEscape(sym *ast.Symbol, isMember bool) bool {
	if sym == nil {
		return false
	}
	if !a.declaredInFunc[sym] {
		return false
	}
	if a.escapedSymbols[sym] {
		return false
	}
	if isMember && a.parameterSymbols[sym] {
		return false
	}
	return true
}

// getBaseIdentifierNode walks down a member expression chain (foo.bar[baz].qux)
// and returns the base identifier node. Only parentheses are transparent —
// type assertions break the chain, matching ESLint's getWriteExpr.
func getBaseIdentifierNode(node *ast.Node) *ast.Node {
	n := ast.SkipParentheses(node)
	for n != nil {
		switch n.Kind {
		case ast.KindIdentifier:
			return n
		case ast.KindPropertyAccessExpression:
			n = ast.SkipParentheses(n.AsPropertyAccessExpression().Expression)
		case ast.KindElementAccessExpression:
			n = ast.SkipParentheses(n.AsElementAccessExpression().Expression)
		default:
			return nil
		}
	}
	return nil
}

// normalizedNodeText returns the source text of a node with internal whitespace collapsed.
func normalizedNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	return strings.Join(strings.Fields(utils.TrimmedNodeText(sourceFile, node)), "")
}

// run performs the analysis on parameter defaults and the function body.
func (a *analyzer) run() {
	body := a.funcNode.Body()
	if body == nil {
		return
	}
	state := newAnalysisState()
	if params := a.funcNode.Parameters(); params != nil {
		for _, param := range params {
			p := param.AsParameterDeclaration()
			if p != nil && p.Initializer != nil {
				a.walkNode(p.Initializer, state)
			}
		}
	}
	a.walkNode(body, state)
}

// walkNode walks a node in evaluation order, updating the analysis state.
func (a *analyzer) walkNode(node *ast.Node, state *analysisState) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return

	case ast.KindClassDeclaration, ast.KindClassExpression:
		return

	case ast.KindBlock:
		// Shadowing is handled naturally by distinct symbols; no mask needed.
		a.walkStatements(node.AsBlock().Statements.Nodes, state)

	case ast.KindCatchClause:
		cc := node.AsCatchClause()
		if cc.Block != nil {
			a.walkNode(cc.Block, state)
		}

	case ast.KindIfStatement:
		a.walkIfStatement(node, state)

	case ast.KindConditionalExpression:
		a.walkConditionalExpression(node, state)

	case ast.KindSwitchStatement:
		a.walkSwitchStatement(node, state)

	case ast.KindLabeledStatement:
		a.walkLabeledStatement(node, state)

	case ast.KindBreakStatement:
		bs := node.AsBreakStatement()
		label := ""
		if bs.Label != nil && bs.Label.Kind == ast.KindIdentifier {
			label = bs.Label.AsIdentifier().Text
		}
		a.recordBreak(label, state)

	case ast.KindTryStatement:
		a.walkTryStatement(node, state)

	case ast.KindWhileStatement:
		whileStmt := node.AsWhileStatement()
		a.walkNode(whileStmt.Expression, state)
		bodyState := state.clone()
		tgt := a.pushBreakTarget("")
		a.walkNode(whileStmt.Statement, bodyState)
		a.popBreakTargetInto(tgt, state)

	case ast.KindDoStatement:
		doStmt := node.AsDoStatement()
		bodyState := state.clone()
		tgt := a.pushBreakTarget("")
		a.walkNode(doStmt.Statement, bodyState)
		a.popBreakTargetInto(tgt, state)
		a.walkNode(doStmt.Expression, state)

	case ast.KindForStatement:
		a.walkForStatement(node, state)

	case ast.KindForInStatement, ast.KindForOfStatement:
		stmt := node.AsForInOrOfStatement()
		a.walkNode(stmt.Expression, state)
		// ESLint quirk: for-of / for-in / for-await-of bodies get a FRESH
		// entry state (pre-loop reads are NOT inherited), but the body's
		// END state IS union-merged back into post-loop — post-loop is
		// reachable from the body's tail via the iterator-done edge, so
		// reads that became outdated inside the body propagate.
		bodyState := newAnalysisState()
		if stmt.AwaitModifier != nil {
			bodyState.makeOutdated()
		}
		a.walkForInOfInitializer(stmt.Initializer, bodyState)
		tgt := a.pushBreakTarget("")
		a.walkNode(stmt.Statement, bodyState)
		state.merge(bodyState)
		a.popBreakTargetInto(tgt, state)

	case ast.KindBinaryExpression:
		a.walkBinaryExpression(node, state)

	case ast.KindAwaitExpression:
		if node.Expression() != nil {
			a.walkNode(node.Expression(), state)
		}
		state.makeOutdated()

	case ast.KindYieldExpression:
		if node.Expression() != nil {
			a.walkNode(node.Expression(), state)
		}
		state.makeOutdated()

	case ast.KindIdentifier:
		if !utils.IsNonReferenceIdentifier(node) {
			if sym := a.symbolOf(node); sym != nil {
				state.markRead(sym)
			}
		}

	case ast.KindVariableDeclaration:
		vd := node.AsVariableDeclaration()
		a.walkBindingDefaults(vd.Name(), state)
		if vd.Initializer != nil {
			a.walkNode(vd.Initializer, state)
		}

	case ast.KindExpressionStatement:
		a.walkNode(node.Expression(), state)

	case ast.KindReturnStatement, ast.KindThrowStatement:
		if node.Expression() != nil {
			a.walkNode(node.Expression(), state)
		}

	case ast.KindCallExpression:
		call := node.AsCallExpression()
		a.walkNode(call.Expression, state)
		if call.Arguments != nil {
			for _, arg := range call.Arguments.Nodes {
				a.walkNode(arg, state)
			}
		}

	case ast.KindNewExpression:
		newExpr := node.AsNewExpression()
		a.walkNode(newExpr.Expression, state)
		if newExpr.Arguments != nil {
			for _, arg := range newExpr.Arguments.Nodes {
				a.walkNode(arg, state)
			}
		}

	case ast.KindTaggedTemplateExpression:
		tagged := node.AsTaggedTemplateExpression()
		a.walkNode(tagged.Tag, state)
		a.walkNode(tagged.Template, state)

	case ast.KindParenthesizedExpression:
		a.walkNode(node.AsParenthesizedExpression().Expression, state)

	// TypeScript type wrappers: walk the runtime expression, skip the type annotation.
	case ast.KindAsExpression:
		a.walkNode(node.AsAsExpression().Expression, state)
	case ast.KindTypeAssertionExpression:
		a.walkNode(node.AsTypeAssertion().Expression, state)
	case ast.KindSatisfiesExpression:
		a.walkNode(node.AsSatisfiesExpression().Expression, state)
	case ast.KindNonNullExpression:
		a.walkNode(node.AsNonNullExpression().Expression, state)

	case ast.KindPropertyAccessExpression:
		a.walkNode(node.AsPropertyAccessExpression().Expression, state)

	case ast.KindElementAccessExpression:
		elem := node.AsElementAccessExpression()
		a.walkNode(elem.Expression, state)
		a.walkNode(elem.ArgumentExpression, state)

	case ast.KindArrayLiteralExpression:
		for _, elem := range node.AsArrayLiteralExpression().Elements.Nodes {
			a.walkNode(elem, state)
		}

	case ast.KindObjectLiteralExpression:
		for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
			a.walkNode(prop, state)
		}

	case ast.KindSpreadElement:
		a.walkNode(node.Expression(), state)

	default:
		a.walkChildren(node, state)
	}
}

// walkBindingDefaults walks only the default-value initializers inside a
// destructuring BINDING pattern (var/let/const declarations). Each default is
// conditional (runs only if the key is undefined), so it's walked in a
// cloned state and union-merged back — an await in one default outdates the
// outer state, but a re-read in a later default doesn't clear it.
func (a *analyzer) walkBindingDefaults(nameNode *ast.Node, state *analysisState) {
	if nameNode == nil {
		return
	}
	switch nameNode.Kind {
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		nameNode.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				be := child.AsBindingElement()
				if be.Initializer != nil {
					branch := state.clone()
					a.walkNode(be.Initializer, branch)
					state.merge(branch)
				}
				if be.Name() != nil {
					a.walkBindingDefaults(be.Name(), state)
				}
			}
			return false
		})
	}
}

func (a *analyzer) walkStatements(statements []*ast.Node, state *analysisState) {
	for _, stmt := range statements {
		a.walkNode(stmt, state)
	}
}

func (a *analyzer) walkChildren(node *ast.Node, state *analysisState) {
	node.ForEachChild(func(child *ast.Node) bool {
		a.walkNode(child, state)
		return false
	})
}

// walkBinaryExpression handles assignments and regular binary operations.
func (a *analyzer) walkBinaryExpression(node *ast.Node, state *analysisState) {
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		a.walkChildren(node, state)
		return
	}

	opKind := binary.OperatorToken.Kind

	if !ast.IsAssignmentOperator(opKind) {
		a.walkNode(binary.Left, state)
		a.walkNode(binary.Right, state)
		return
	}

	// --- Assignment expression ---
	left := ast.SkipOuterExpressions(binary.Left, ast.OEKParentheses|ast.OEKAssertions)
	isCompound := ast.IsCompoundAssignment(opKind)
	isMember := ast.IsAccessExpression(left)

	var targetSym *ast.Symbol
	var targetIdentNode *ast.Node
	if isMember {
		targetIdentNode = getBaseIdentifierNode(left)
	} else if left.Kind == ast.KindIdentifier {
		targetIdentNode = left
	}
	if targetIdentNode != nil {
		targetSym = a.symbolOf(targetIdentNode)
	}

	if targetIdentNode == nil {
		// Destructuring assignment: `({a, b} = src)` / `[a, b] = src`.
		if !isCompound && isDestructuringAssignmentTarget(left) {
			a.walkDestructuringAssignment(node, left, binary.Right, state)
			return
		}
		a.walkNode(binary.Left, state)
		a.walkNode(binary.Right, state)
		return
	}

	// Walk the LHS (side-reads always register; the write-target chain itself
	// is skipped for simple assignments to match ESLint's getWriteExpr chain).
	if isCompound {
		a.walkNode(binary.Left, state)
	} else {
		a.walkSimpleAssignLHS(binary.Left, state)
	}

	a.walkNode(binary.Right, state)

	// Undeclared globals (sym == nil) and local-without-escape vars: no race.
	if targetSym == nil {
		return
	}
	if a.isLocalWithoutEscape(targetSym, isMember) {
		return
	}

	// ESLint quirk: RHS that is a function/arrow expression silently skips
	// the check (the :expression:exit fires inside the inner function's
	// CodePath whose referenceMap is null).
	if isFunctionLikeRHS(binary.Right) {
		return
	}

	if state.isOutdated(targetSym) {
		// TS type-wrapped simple target (e.g. `(foo as any) = 1`): ESLint
		// reports nonAtomicObjectUpdate with the wrapper's source text as the
		// value, and allowProperties: true suppresses it. Detect by comparing
		// the raw LHS to its assertion-stripped form.
		wrapped := !isMember && binary.Left != left
		if wrapped {
			a.report(node, binary.Left, targetSym, true)
		} else {
			a.report(node, left, targetSym, isMember)
		}
	}
}

func (a *analyzer) report(assignmentNode, left *ast.Node, sym *ast.Symbol, isMemberAccess bool) {
	if isMemberAccess {
		if a.allowProperties {
			return
		}
		leftText := normalizedNodeText(a.ctx.SourceFile, left)
		a.ctx.ReportNode(assignmentNode, messageNonAtomicObjectUpdate(leftText, sym.Name))
	} else {
		a.ctx.ReportNode(assignmentNode, messageNonAtomicUpdate(sym.Name))
	}
}

// isFunctionLikeRHS reports whether an assignment's RHS is a function-like
// expression. ESLint's :expression:exit handler skips outdated checks when
// the RHS starts a new CodePath with a non-async/non-generator scope.
func isFunctionLikeRHS(node *ast.Node) bool {
	if node == nil {
		return false
	}
	n := ast.SkipParentheses(node)
	switch n.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	}
	return false
}

// isDestructuringAssignmentTarget returns true if `node` is an ObjectLiteral
// or ArrayLiteral being reinterpreted as a destructuring pattern.
func isDestructuringAssignmentTarget(node *ast.Node) bool {
	n := ast.SkipParentheses(node)
	return n.Kind == ast.KindObjectLiteralExpression || n.Kind == ast.KindArrayLiteralExpression
}

// destructuringTarget records a single variable-write target collected from
// a destructuring pattern.
type destructuringTarget struct {
	symbol *ast.Symbol
	name   string
}

// walkDestructuringAssignment handles `({a, b = def} = src)` / `[a, b] = src`.
func (a *analyzer) walkDestructuringAssignment(assignNode, pattern, rhs *ast.Node, state *analysisState) {
	if rhs != nil {
		a.walkNode(rhs, state)
	}
	var targets []destructuringTarget
	a.collectDestructuringTargets(pattern, state, &targets)

	if isFunctionLikeRHS(rhs) {
		return
	}
	leftText := normalizedNodeText(a.ctx.SourceFile, pattern)
	for _, t := range targets {
		if t.symbol == nil {
			continue
		}
		if a.isLocalWithoutEscape(t.symbol, false) {
			continue
		}
		if state.isOutdated(t.symbol) {
			a.ctx.ReportNode(assignNode, messageNonAtomicObjectUpdate(leftText, t.name))
		}
	}
}

// collectDestructuringTargets walks a destructuring pattern. Default
// expressions are walked (in cloned branches) so awaits in defaults outdate
// without letting re-reads clear the outer state. Each variable-target leaf is
// appended to `targets`.
func (a *analyzer) collectDestructuringTargets(node *ast.Node, state *analysisState, targets *[]destructuringTarget) {
	if node == nil {
		return
	}
	n := ast.SkipParentheses(node)
	switch n.Kind {
	case ast.KindObjectLiteralExpression:
		for _, prop := range n.AsObjectLiteralExpression().Properties.Nodes {
			a.collectDestructuringProperty(prop, state, targets)
		}
	case ast.KindArrayLiteralExpression:
		for _, elem := range n.AsArrayLiteralExpression().Elements.Nodes {
			a.collectDestructuringElement(elem, state, targets)
		}
	case ast.KindIdentifier:
		if sym := a.symbolOf(n); sym != nil {
			*targets = append(*targets, destructuringTarget{symbol: sym, name: sym.Name})
		}
	default:
		// Member target or other — walk as a normal expression so any side
		// reads register, but don't record as a tracked target.
		a.walkNode(n, state)
	}
}

func (a *analyzer) collectDestructuringProperty(prop *ast.Node, state *analysisState, targets *[]destructuringTarget) {
	switch prop.Kind {
	case ast.KindShorthandPropertyAssignment:
		sp := prop.AsShorthandPropertyAssignment()
		if sp.ObjectAssignmentInitializer != nil {
			branch := state.clone()
			a.walkNode(sp.ObjectAssignmentInitializer, branch)
			state.merge(branch)
		}
		name := sp.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			if sym := a.symbolOf(name); sym != nil {
				*targets = append(*targets, destructuringTarget{symbol: sym, name: sym.Name})
			}
		}
	case ast.KindPropertyAssignment:
		pa := prop.AsPropertyAssignment()
		if pa.Name() != nil && pa.Name().Kind == ast.KindComputedPropertyName {
			a.walkNode(pa.Name(), state)
		}
		a.collectDestructuringValue(pa.Initializer, state, targets)
	case ast.KindSpreadAssignment:
		a.collectDestructuringValue(prop.AsSpreadAssignment().Expression, state, targets)
	default:
		a.walkNode(prop, state)
	}
}

func (a *analyzer) collectDestructuringElement(elem *ast.Node, state *analysisState, targets *[]destructuringTarget) {
	if elem == nil || elem.Kind == ast.KindOmittedExpression {
		return
	}
	if elem.Kind == ast.KindSpreadElement {
		a.collectDestructuringValue(elem.Expression(), state, targets)
		return
	}
	a.collectDestructuringValue(elem, state, targets)
}

func (a *analyzer) collectDestructuringValue(node *ast.Node, state *analysisState, targets *[]destructuringTarget) {
	if node == nil {
		return
	}
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindBinaryExpression {
		bin := n.AsBinaryExpression()
		if bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
			branch := state.clone()
			a.walkNode(bin.Right, branch)
			state.merge(branch)
			a.collectDestructuringValue(bin.Left, state, targets)
			return
		}
	}
	a.collectDestructuringTargets(n, state, targets)
}

// walkLabeledStatement handles `label: stmt`. Loops/switches inside a labeled
// statement become reachable via `break <label>` from nested constructs.
func (a *analyzer) walkLabeledStatement(node *ast.Node, state *analysisState) {
	ls := node.AsLabeledStatement()
	label := ""
	if ls.Label != nil && ls.Label.Kind == ast.KindIdentifier {
		label = ls.Label.AsIdentifier().Text
	}
	body := ls.Statement
	switch body.Kind {
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement, ast.KindSwitchStatement,
		ast.KindBlock:
		t := a.pushBreakTarget(label)
		a.walkNode(body, state)
		a.popBreakTargetInto(t, state)
	default:
		a.walkNode(body, state)
	}
}

func (a *analyzer) pushBreakTarget(label string) *breakTarget {
	t := &breakTarget{label: label}
	a.breakTargets = append(a.breakTargets, t)
	return t
}

func (a *analyzer) popBreakTargetInto(t *breakTarget, post *analysisState) {
	if n := len(a.breakTargets); n > 0 && a.breakTargets[n-1] == t {
		a.breakTargets = a.breakTargets[:n-1]
	}
	for _, s := range t.states {
		post.merge(s)
	}
}

func (a *analyzer) recordBreak(label string, state *analysisState) {
	for i := len(a.breakTargets) - 1; i >= 0; i-- {
		t := a.breakTargets[i]
		if label == "" || t.label == label {
			t.states = append(t.states, state.clone())
			return
		}
	}
}

// walkForInOfInitializer walks a for-in / for-of statement's Initializer in
// bodyState so awaits inside destructuring defaults correctly outdate per
// iteration.
func (a *analyzer) walkForInOfInitializer(init *ast.Node, state *analysisState) {
	if init == nil {
		return
	}
	if init.Kind == ast.KindVariableDeclarationList {
		a.walkNode(init, state)
		return
	}
	n := ast.SkipParentheses(init)
	switch n.Kind {
	case ast.KindIdentifier:
		// `for (foo of gen)` — foo is a pure write target; no read.
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
		var discard []destructuringTarget
		a.collectDestructuringTargets(n, state, &discard)
	default:
		a.walkNode(n, state)
	}
}

// walkSimpleAssignLHS walks the LHS of a simple `=` assignment. The
// write-target chain (the base identifier reached via PropertyAccess.Expression
// / ElementAccess.Expression / TS wrappers) is NOT marked as a read. Side
// expressions (computed indices, etc.) are walked normally.
func (a *analyzer) walkSimpleAssignLHS(node *ast.Node, state *analysisState) {
	if node == nil {
		return
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		a.walkSimpleAssignLHS(node.AsPropertyAccessExpression().Expression, state)
	case ast.KindElementAccessExpression:
		elem := node.AsElementAccessExpression()
		a.walkSimpleAssignLHS(elem.Expression, state)
		a.walkNode(elem.ArgumentExpression, state)
	case ast.KindAsExpression:
		a.walkSimpleAssignLHS(node.AsAsExpression().Expression, state)
	case ast.KindTypeAssertionExpression:
		a.walkSimpleAssignLHS(node.AsTypeAssertion().Expression, state)
	case ast.KindSatisfiesExpression:
		a.walkSimpleAssignLHS(node.AsSatisfiesExpression().Expression, state)
	case ast.KindNonNullExpression:
		a.walkSimpleAssignLHS(node.AsNonNullExpression().Expression, state)
	case ast.KindIdentifier:
		// Base of the write-target chain: do not mark as a read.
	default:
		a.walkNode(node, state)
	}
}

// stmtAlwaysTerminates returns true if a statement always exits via
// return/throw.
func stmtAlwaysTerminates(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return true
	case ast.KindBlock:
		for _, stmt := range node.AsBlock().Statements.Nodes {
			if stmtAlwaysTerminates(stmt) {
				return true
			}
		}
		return false
	case ast.KindIfStatement:
		ifStmt := node.AsIfStatement()
		if ifStmt.ElseStatement == nil {
			return false
		}
		return stmtAlwaysTerminates(ifStmt.ThenStatement) && stmtAlwaysTerminates(ifStmt.ElseStatement)
	}
	return false
}

func (a *analyzer) walkIfStatement(node *ast.Node, state *analysisState) {
	ifStmt := node.AsIfStatement()

	a.walkNode(ifStmt.Expression, state)

	thenState := state.clone()
	a.walkNode(ifStmt.ThenStatement, thenState)
	thenExits := stmtAlwaysTerminates(ifStmt.ThenStatement)

	if ifStmt.ElseStatement != nil {
		elseState := state.clone()
		a.walkNode(ifStmt.ElseStatement, elseState)
		elseExits := stmtAlwaysTerminates(ifStmt.ElseStatement)

		switch {
		case thenExits && elseExits:
			*state = *thenState
		case thenExits:
			*state = *elseState
		case elseExits:
			*state = *thenState
		default:
			thenState.merge(elseState)
			*state = *thenState
		}
	} else {
		if !thenExits {
			state.merge(thenState)
		}
	}
}

func (a *analyzer) walkConditionalExpression(node *ast.Node, state *analysisState) {
	cond := node.AsConditionalExpression()

	a.walkNode(cond.Condition, state)

	consequentState := state.clone()
	a.walkNode(cond.WhenTrue, consequentState)

	alternateState := state.clone()
	a.walkNode(cond.WhenFalse, alternateState)

	consequentState.merge(alternateState)
	*state = *consequentState
}

// clauseEndKind classifies how a switch clause's statements end.
type clauseEndKind int

const (
	clauseEndFallthrough clauseEndKind = iota
	clauseEndBreak
	clauseEndExit
)

func classifyClauseEnd(statements []*ast.Node) clauseEndKind {
	for _, stmt := range statements {
		if stmtAlwaysTerminates(stmt) {
			return clauseEndExit
		}
		if stmt.Kind == ast.KindBreakStatement {
			return clauseEndBreak
		}
		if stmt.Kind == ast.KindContinueStatement {
			return clauseEndExit
		}
	}
	return clauseEndFallthrough
}

func (a *analyzer) walkSwitchStatement(node *ast.Node, state *analysisState) {
	switchStmt := node.AsSwitchStatement()

	a.walkNode(switchStmt.Expression, state)

	caseBlock := switchStmt.CaseBlock.AsCaseBlock()
	if caseBlock == nil {
		return
	}

	swTgt := a.pushBreakTarget("")

	mergedState := state.clone()
	var fallthroughState *analysisState
	clauses := caseBlock.Clauses.Nodes

	for i, clause := range clauses {
		caseOrDefault := clause.AsCaseOrDefaultClause()
		if caseOrDefault == nil {
			continue
		}

		clauseState := state.clone()
		if fallthroughState != nil {
			clauseState.merge(fallthroughState)
		}

		if clause.Kind == ast.KindCaseClause && caseOrDefault.Expression != nil {
			a.walkNode(caseOrDefault.Expression, clauseState)
		}
		var clauseStatements []*ast.Node
		if caseOrDefault.Statements != nil {
			clauseStatements = caseOrDefault.Statements.Nodes
			a.walkStatements(clauseStatements, clauseState)
		}

		switch classifyClauseEnd(clauseStatements) {
		case clauseEndBreak, clauseEndExit:
			fallthroughState = nil
		case clauseEndFallthrough:
			if i == len(clauses)-1 {
				mergedState.merge(clauseState)
			}
			fallthroughState = clauseState
		}
	}

	a.popBreakTargetInto(swTgt, mergedState)
	*state = *mergedState
}

func (a *analyzer) walkTryStatement(node *ast.Node, state *analysisState) {
	tryStmt := node.AsTryStatement()

	tryEnd := state.clone()
	if tryStmt.TryBlock != nil {
		a.walkNode(tryStmt.TryBlock, tryEnd)
	}

	tryTerminates := tryStmt.TryBlock != nil && stmtAlwaysTerminates(tryStmt.TryBlock)

	var catchEnd *analysisState
	if tryStmt.CatchClause != nil {
		if tryTerminates {
			catchEnd = state.clone()
		} else {
			catchEnd = tryEnd.clone()
		}
		a.walkNode(tryStmt.CatchClause, catchEnd)
	}

	var catchTerminates bool
	if tryStmt.CatchClause != nil {
		cc := tryStmt.CatchClause.AsCatchClause()
		if cc.Block != nil {
			catchTerminates = stmtAlwaysTerminates(cc.Block)
		}
	}

	post := newAnalysisState()
	contributions := 0
	if !tryTerminates {
		post.merge(tryEnd)
		contributions++
	}
	if tryStmt.CatchClause != nil && !catchTerminates {
		post.merge(catchEnd)
		contributions++
	}
	if contributions == 0 {
		*state = *state.clone()
	} else {
		*state = *post
	}

	if tryStmt.FinallyBlock != nil {
		a.walkNode(tryStmt.FinallyBlock, state)
	}
}

func (a *analyzer) walkForStatement(node *ast.Node, state *analysisState) {
	forStmt := node.AsForStatement()
	if forStmt.Initializer != nil {
		a.walkNode(forStmt.Initializer, state)
	}
	if forStmt.Condition != nil {
		a.walkNode(forStmt.Condition, state)
	}
	bodyState := state.clone()
	tgt := a.pushBreakTarget("")
	a.walkNode(forStmt.Statement, bodyState)
	a.popBreakTargetInto(tgt, state)
	// Incrementor runs in its own segment whose entry state is fresh.
	if forStmt.Incrementor != nil {
		a.walkNode(forStmt.Incrementor, newAnalysisState())
	}
}

// RequireAtomicUpdatesRule disallows assignments that can lead to race conditions
// due to usage of `await` or `yield`.
var RequireAtomicUpdatesRule = rule.Rule{
	Name:             "require-atomic-updates",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		allowProperties := false
		optsMap := utils.GetOptionsMap(options)
		if optsMap != nil {
			if v, ok := optsMap["allowProperties"]; ok {
				if b, ok := v.(bool); ok {
					allowProperties = b
				}
			}
		}

		enterFunction := func(node *ast.Node) {
			flags := ast.GetFunctionFlags(node)
			if flags&(ast.FunctionFlagsAsync|ast.FunctionFlagsGenerator) == 0 {
				return
			}
			a := newAnalyzer(ctx, node, allowProperties)
			a.run()
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: enterFunction,
			ast.KindFunctionExpression:  enterFunction,
			ast.KindArrowFunction:       enterFunction,
			ast.KindMethodDeclaration:   enterFunction,
			ast.KindGetAccessor:         enterFunction,
			ast.KindSetAccessor:         enterFunction,
		}
	},
}
