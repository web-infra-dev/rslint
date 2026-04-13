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

// analysisState tracks which variables have been read before/after an await/yield.
//
// Algorithm (mirrors ESLint's code-path segment approach):
//   - When a variable is read, it is added to freshReads.
//   - When an await/yield is encountered, all freshReads move to outdatedReads.
//   - When an assignment targets an outdated variable, the rule reports.
//   - Re-reading a variable removes it from outdated (the new read is current).
type analysisState struct {
	freshReads    map[string]bool
	outdatedReads map[string]bool
}

func newAnalysisState() *analysisState {
	return &analysisState{
		freshReads:    make(map[string]bool),
		outdatedReads: make(map[string]bool),
	}
}

func (s *analysisState) clone() *analysisState {
	fresh := make(map[string]bool, len(s.freshReads))
	outdated := make(map[string]bool, len(s.outdatedReads))
	for k := range s.freshReads {
		fresh[k] = true
	}
	for k := range s.outdatedReads {
		outdated[k] = true
	}
	return &analysisState{freshReads: fresh, outdatedReads: outdated}
}

// merge unions another branch's state into this one (used after if/else, ternary, etc.).
func (s *analysisState) merge(other *analysisState) {
	for k := range other.freshReads {
		s.freshReads[k] = true
	}
	for k := range other.outdatedReads {
		s.outdatedReads[k] = true
	}
}

func (s *analysisState) markRead(varName string) {
	s.freshReads[varName] = true
	// A re-read after await captures the current value, so the variable is no longer outdated.
	delete(s.outdatedReads, varName)
}

// makeOutdated moves all fresh reads to outdated (called on await/yield).
func (s *analysisState) makeOutdated() {
	for k := range s.freshReads {
		s.outdatedReads[k] = true
	}
	s.freshReads = make(map[string]bool)
}

func (s *analysisState) isOutdated(varName string) bool {
	return s.outdatedReads[varName]
}

// getIdentifierName returns the name if the node (after stripping parentheses) is an Identifier.
func getIdentifierName(node *ast.Node) string {
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindIdentifier {
		return n.AsIdentifier().Text
	}
	return ""
}

// getBaseIdentifierNode walks down a member expression chain (foo.bar[baz].qux)
// and returns the base identifier node and its name. Unlike ast.GetFirstIdentifier,
// this handles ElementAccessExpression and returns nil instead of panicking.
func getBaseIdentifierNode(node *ast.Node) (*ast.Node, string) {
	n := ast.SkipParentheses(node)
	for n != nil {
		switch n.Kind {
		case ast.KindIdentifier:
			return n, n.AsIdentifier().Text
		case ast.KindPropertyAccessExpression:
			n = ast.SkipParentheses(n.AsPropertyAccessExpression().Expression)
		case ast.KindElementAccessExpression:
			n = ast.SkipParentheses(n.AsElementAccessExpression().Expression)
		default:
			return nil, ""
		}
	}
	return nil, ""
}

// normalizedNodeText returns the source text of a node with internal whitespace collapsed.
// For example, "foo . bar" → "foo.bar". Used in diagnostic messages.
func normalizedNodeText(sourceFile *ast.SourceFile, node *ast.Node) string {
	return strings.Join(strings.Fields(utils.TrimmedNodeText(sourceFile, node)), "")
}

// analyzer performs the require-atomic-updates analysis within one resumable
// (async or generator) function. It walks the function body in evaluation order,
// tracking variable reads relative to await/yield, and reports assignments that
// may use stale values.
type analyzer struct {
	ctx             rule.RuleContext
	funcNode        *ast.Node
	allowProperties bool

	// localVarsSafe: true = declared in this function and never referenced in a closure.
	localVarsSafe map[string]bool
	paramNames    map[string]bool
	// declaredOuterVars: variables declared in outer scopes. Together with
	// localVarsSafe, this is used to distinguish declared variables from
	// undeclared globals (which are not tracked per ESLint semantics).
	declaredOuterVars map[string]bool
}

func newAnalyzer(ctx rule.RuleContext, funcNode *ast.Node, allowProperties bool) *analyzer {
	a := &analyzer{
		ctx:               ctx,
		funcNode:          funcNode,
		allowProperties:   allowProperties,
		localVarsSafe:     make(map[string]bool),
		paramNames:        make(map[string]bool),
		declaredOuterVars: make(map[string]bool),
	}
	a.collectLocalDeclarations()
	a.collectOuterDeclarations()
	return a
}

// collectLocalDeclarations gathers all declarations local to this function
// and determines which ones escape to closures.
func (a *analyzer) collectLocalDeclarations() {
	allLocals := make(map[string]bool)
	escapedLocals := make(map[string]bool)

	if params := a.funcNode.Parameters(); params != nil {
		for _, param := range params {
			utils.CollectBindingNames(param.Name(), func(_ *ast.Node, name string) {
				allLocals[name] = true
				a.paramNames[name] = true
			})
		}
	}

	body := a.funcNode.Body()
	if body == nil {
		return
	}
	a.collectDeclsInBlock(body, allLocals)
	a.findEscapedVars(body, allLocals, escapedLocals, false, nil)

	for name := range allLocals {
		a.localVarsSafe[name] = !escapedLocals[name]
	}
}

// collectOuterDeclarations walks up from the function to the source file,
// collecting variable declarations in outer scopes. Undeclared globals are
// not tracked (ESLint skips unresolved references).
func (a *analyzer) collectOuterDeclarations() {
	current := a.funcNode.Parent
	for current != nil {
		switch current.Kind {
		case ast.KindSourceFile:
			for _, stmt := range current.AsSourceFile().Statements.Nodes {
				a.collectOuterDeclsFromStmt(stmt)
			}
		case ast.KindBlock:
			for _, stmt := range current.AsBlock().Statements.Nodes {
				a.collectOuterDeclsFromStmt(stmt)
			}
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindArrowFunction, ast.KindMethodDeclaration:
			if params := current.Parameters(); params != nil {
				for _, param := range params {
					utils.CollectBindingNames(param.Name(), func(_ *ast.Node, name string) {
						a.declaredOuterVars[name] = true
					})
				}
			}
		case ast.KindCatchClause:
			cc := current.AsCatchClause()
			if cc.VariableDeclaration != nil {
				utils.CollectBindingNames(cc.VariableDeclaration.Name(), func(_ *ast.Node, name string) {
					a.declaredOuterVars[name] = true
				})
			}
		case ast.KindForStatement:
			forStmt := current.AsForStatement()
			if forStmt.Initializer != nil && forStmt.Initializer.Kind == ast.KindVariableDeclarationList {
				a.collectDeclsFromVarList(forStmt.Initializer, a.declaredOuterVars)
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := current.AsForInOrOfStatement()
			if stmt.Initializer != nil && stmt.Initializer.Kind == ast.KindVariableDeclarationList {
				a.collectDeclsFromVarList(stmt.Initializer, a.declaredOuterVars)
			}
		}
		current = current.Parent
	}
}

func (a *analyzer) collectOuterDeclsFromStmt(stmt *ast.Node) {
	switch stmt.Kind {
	case ast.KindVariableStatement:
		stmt.ForEachChild(func(declList *ast.Node) bool {
			if declList.Kind == ast.KindVariableDeclarationList {
				a.collectDeclsFromVarList(declList, a.declaredOuterVars)
			}
			return false
		})
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
		if stmt.Name() != nil && stmt.Name().Kind == ast.KindIdentifier {
			a.declaredOuterVars[stmt.Name().AsIdentifier().Text] = true
		}
	}
}

// collectDeclsFromVarList extracts variable names from a VariableDeclarationList.
func (a *analyzer) collectDeclsFromVarList(declList *ast.Node, target map[string]bool) {
	declList.ForEachChild(func(decl *ast.Node) bool {
		if decl.Kind == ast.KindVariableDeclaration {
			utils.CollectBindingNames(decl.Name(), func(_ *ast.Node, name string) {
				target[name] = true
			})
		}
		return false
	})
}

// collectDeclsInBlock collects variable declarations from a block without
// recursing into nested functions. The topLevel flag controls scoping:
//   - topLevel=true  (function body): collect let/const/var/function/class
//   - topLevel=false (nested block):  collect only var (it hoists to function scope)
func (a *analyzer) collectDeclsInBlock(node *ast.Node, locals map[string]bool) {
	a.collectDeclsInBlockImpl(node, locals, true)
}

func (a *analyzer) collectDeclsInBlockImpl(node *ast.Node, locals map[string]bool, topLevel bool) {
	if node == nil {
		return
	}

	node.ForEachChild(func(child *ast.Node) bool {
		switch child.Kind {
		case ast.KindVariableStatement:
			child.ForEachChild(func(declList *ast.Node) bool {
				if declList.Kind == ast.KindVariableDeclarationList {
					// In nested blocks, only collect var (hoisted); skip let/const (block-scoped).
					if topLevel || utils.IsVarKeyword(declList) {
						a.collectDeclsFromVarList(declList, locals)
					}
				}
				return false
			})
		case ast.KindFunctionDeclaration:
			if child.Name() != nil && child.Name().Kind == ast.KindIdentifier {
				locals[child.Name().AsIdentifier().Text] = true
			}
			// Don't recurse into the function body
		case ast.KindClassDeclaration:
			if child.Name() != nil && child.Name().Kind == ast.KindIdentifier {
				locals[child.Name().AsIdentifier().Text] = true
			}
		default:
			if !ast.IsFunctionLikeDeclaration(child) {
				a.collectDeclsInBlockImpl(child, locals, false)
			}
		}
		return false
	})
}

// findEscapedVars finds local variables referenced inside nested functions.
// shadowed tracks variables redeclared in inner scopes to avoid false positives
// (e.g., inner `let foo` should not cause outer `foo` to be marked escaped).
func (a *analyzer) findEscapedVars(node *ast.Node, locals map[string]bool, escaped map[string]bool, inNestedFunc bool, shadowed map[string]bool) {
	if node == nil {
		return
	}

	node.ForEachChild(func(child *ast.Node) bool {
		if ast.IsFunctionLikeDeclaration(child) {
			// Collect declarations in the nested function to detect shadowing
			innerLocals := make(map[string]bool)
			a.collectNestedFuncDecls(child, innerLocals)

			innerShadowed := make(map[string]bool, len(shadowed)+len(innerLocals))
			for k := range shadowed {
				innerShadowed[k] = true
			}
			for k := range innerLocals {
				innerShadowed[k] = true
			}

			a.findEscapedVars(child, locals, escaped, true, innerShadowed)
			return false
		}

		if child.Kind == ast.KindIdentifier && !utils.IsNonReferenceIdentifier(child) {
			name := child.AsIdentifier().Text
			if inNestedFunc && locals[name] && !shadowed[name] {
				escaped[name] = true
			}
		}

		a.findEscapedVars(child, locals, escaped, inNestedFunc, shadowed)
		return false
	})
}

func (a *analyzer) collectNestedFuncDecls(funcNode *ast.Node, decls map[string]bool) {
	if params := funcNode.Parameters(); params != nil {
		for _, param := range params {
			utils.CollectBindingNames(param.Name(), func(_ *ast.Node, name string) {
				decls[name] = true
			})
		}
	}
	if body := funcNode.Body(); body != nil {
		a.collectDeclsInBlock(body, decls)
	}
}

// isDeclaredVariable returns true if the variable is declared in any scope
// (local or outer). Falls back to TypeChecker to resolve globals from type
// definitions (e.g., `process` from @types/node). Truly undeclared names return false.
func (a *analyzer) isDeclaredVariable(name string, identNode *ast.Node) bool {
	_, isLocal := a.localVarsSafe[name]
	if isLocal || a.declaredOuterVars[name] {
		return true
	}
	// TypeChecker can resolve globals from lib/types (e.g., process, console).
	if a.ctx.TypeChecker != nil && identNode != nil {
		sym := a.ctx.TypeChecker.GetSymbolAtLocation(identNode)
		if sym == nil {
			return false
		}
		// If the symbol is declared inside the current function, it's local (safe).
		// This covers block-scoped variables (for-loop let, catch var) that our
		// AST-based collection may have missed.
		if sym.ValueDeclaration != nil && a.isNodeInsideFunc(sym.ValueDeclaration) {
			a.localVarsSafe[name] = true
			return true
		}
		return true
	}
	return false
}

// isNodeInsideFunc checks if a node is a descendant of the current function.
func (a *analyzer) isNodeInsideFunc(node *ast.Node) bool {
	return ast.FindAncestor(node, func(n *ast.Node) bool {
		return n == a.funcNode
	}) != nil
}

// isLocalVariableWithoutEscape returns true if the variable is local and
// never referenced in a closure (i.e., no other concurrent code can observe it).
// For member access on parameters, returns false because the object comes from outside.
func (a *analyzer) isLocalVariableWithoutEscape(name string, isMemberAccess bool) bool {
	if isMemberAccess && a.paramNames[name] {
		return false
	}
	safe, exists := a.localVarsSafe[name]
	return exists && safe
}

// run performs the analysis on parameter defaults and the function body.
func (a *analyzer) run() {
	body := a.funcNode.Body()
	if body == nil {
		return
	}
	state := newAnalysisState()
	// Parameter default values are evaluated at call time and may contain await.
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
		// Function boundaries: don't recurse (inner functions are separate scopes)
		return

	case ast.KindClassDeclaration, ast.KindClassExpression:
		// Class bodies contain methods which are separate function scopes
		return

	case ast.KindBlock:
		a.walkStatements(node.AsBlock().Statements.Nodes, state)

	case ast.KindIfStatement:
		a.walkIfStatement(node, state)

	case ast.KindConditionalExpression:
		a.walkConditionalExpression(node, state)

	case ast.KindSwitchStatement:
		a.walkSwitchStatement(node, state)

	case ast.KindTryStatement:
		a.walkTryStatement(node, state)

	case ast.KindWhileStatement:
		whileStmt := node.AsWhileStatement()
		a.walkNode(whileStmt.Expression, state)
		a.walkNode(whileStmt.Statement, state)

	case ast.KindDoStatement:
		doStmt := node.AsDoStatement()
		a.walkNode(doStmt.Statement, state)
		a.walkNode(doStmt.Expression, state)

	case ast.KindForStatement:
		a.walkForStatement(node, state)

	case ast.KindForInStatement, ast.KindForOfStatement:
		stmt := node.AsForInOrOfStatement()
		a.walkNode(stmt.Expression, state)
		// for-await-of implicitly awaits on each iteration
		if stmt.AwaitModifier != nil {
			state.makeOutdated()
		}
		a.walkNode(stmt.Statement, state)

	case ast.KindBinaryExpression:
		a.walkBinaryExpression(node, state)

	case ast.KindAwaitExpression:
		// The operand is evaluated BEFORE the await pauses
		if node.Expression() != nil {
			a.walkNode(node.Expression(), state)
		}
		state.makeOutdated()

	case ast.KindYieldExpression:
		// The operand is evaluated BEFORE the yield pauses
		if node.Expression() != nil {
			a.walkNode(node.Expression(), state)
		}
		state.makeOutdated()

	case ast.KindIdentifier:
		if !utils.IsNonReferenceIdentifier(node) {
			if name := node.AsIdentifier().Text; name != "" {
				state.markRead(name)
			}
		}

	case ast.KindVariableDeclaration:
		vd := node.AsVariableDeclaration()
		// Walk binding pattern defaults (e.g., `const {a = await bar} = obj`)
		// Declaration identifiers are skipped by IsNonReferenceIdentifier.
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
	// Without this, type reference identifiers (e.g., MyType in `x as MyType`) would
	// be incorrectly marked as variable reads.
	case ast.KindAsExpression:
		a.walkNode(node.AsAsExpression().Expression, state)
	case ast.KindTypeAssertionExpression:
		a.walkNode(node.AsTypeAssertion().Expression, state)
	case ast.KindSatisfiesExpression:
		a.walkNode(node.AsSatisfiesExpression().Expression, state)
	case ast.KindNonNullExpression:
		a.walkNode(node.AsNonNullExpression().Expression, state)

	case ast.KindPropertyAccessExpression:
		// Walk only the object expression, not the property name
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
// destructuring binding pattern (e.g., `{a = await bar}` or `[x = await y]`).
// This is needed because default values are evaluated at runtime and may contain await.
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
					a.walkNode(be.Initializer, state)
				}
				// Recurse into nested patterns (e.g., `{a: {b = await c}} = obj`)
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
	// Skip parentheses and TS type assertions (as T, <T>, !, satisfies T) on LHS.
	left := ast.SkipOuterExpressions(binary.Left, ast.OEKParentheses|ast.OEKAssertions)
	isCompound := ast.IsCompoundAssignment(opKind)
	isMember := ast.IsAccessExpression(left)

	// Determine the target variable name and its identifier node
	var targetName string
	var targetIdentNode *ast.Node
	if isMember {
		targetIdentNode, targetName = getBaseIdentifierNode(left)
	} else {
		targetName = getIdentifierName(left)
		if left.Kind == ast.KindIdentifier {
			targetIdentNode = left
		}
	}

	if targetName == "" {
		a.walkNode(binary.Left, state)
		a.walkNode(binary.Right, state)
		return
	}

	// Skip undeclared globals (ESLint skips unresolved references)
	if !a.isDeclaredVariable(targetName, targetIdentNode) {
		if isCompound {
			a.walkNode(binary.Left, state)
		}
		a.walkNode(binary.Right, state)
		return
	}

	// Local variables that don't escape can't have race conditions
	if a.isLocalVariableWithoutEscape(targetName, isMember) {
		if isCompound {
			a.walkNode(binary.Left, state)
		}
		a.walkNode(binary.Right, state)
		return
	}

	// For compound assignments (+=, -=, etc.), the LHS is read first.
	// For simple assignment (=), the LHS is NOT marked as a fresh read
	// (matches ESLint behavior: only the write target is registered).
	if isCompound {
		a.walkNode(binary.Left, state)
	}

	a.walkNode(binary.Right, state)

	// Check if the target variable was outdated when assigned
	if state.isOutdated(targetName) {
		a.report(node, left, targetName, isMember)
	}
}

func (a *analyzer) report(assignmentNode *ast.Node, left *ast.Node, targetName string, isMemberAccess bool) {
	if isMemberAccess {
		if a.allowProperties {
			return
		}
		leftText := normalizedNodeText(a.ctx.SourceFile, left)
		a.ctx.ReportNode(assignmentNode, messageNonAtomicObjectUpdate(leftText, targetName))
	} else {
		a.ctx.ReportNode(assignmentNode, messageNonAtomicUpdate(targetName))
	}
}

// stmtAlwaysTerminates returns true if a statement always exits via
// return/throw (so the state from this branch should NOT merge into
// the continuation).
func stmtAlwaysTerminates(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindReturnStatement, ast.KindThrowStatement:
		return true
	case ast.KindBlock:
		statements := node.AsBlock().Statements.Nodes
		if len(statements) == 0 {
			return false
		}
		return stmtAlwaysTerminates(statements[len(statements)-1])
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
			// Both branches exit — code after if is unreachable.
			// Use either state (doesn't matter).
			*state = *thenState
		case thenExits:
			// Only then exits — continuation only from else path.
			*state = *elseState
		case elseExits:
			// Only else exits — continuation only from then path.
			*state = *thenState
		default:
			// Neither exits — merge both paths.
			thenState.merge(elseState)
			*state = *thenState
		}
	} else {
		if thenExits {
			// Then-branch always exits — continuation only from the non-taken path.
			// state already represents the pre-if state (condition was walked).
		} else {
			*state = *thenState
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

func (a *analyzer) walkSwitchStatement(node *ast.Node, state *analysisState) {
	switchStmt := node.AsSwitchStatement()

	a.walkNode(switchStmt.Expression, state)

	caseBlock := switchStmt.CaseBlock.AsCaseBlock()
	if caseBlock == nil {
		return
	}

	mergedState := state.clone()
	for _, clause := range caseBlock.Clauses.Nodes {
		clauseState := state.clone()
		caseOrDefault := clause.AsCaseOrDefaultClause()
		if caseOrDefault == nil {
			continue
		}
		if clause.Kind == ast.KindCaseClause && caseOrDefault.Expression != nil {
			a.walkNode(caseOrDefault.Expression, clauseState)
		}
		if caseOrDefault.Statements != nil {
			a.walkStatements(caseOrDefault.Statements.Nodes, clauseState)
		}
		mergedState.merge(clauseState)
	}
	*state = *mergedState
}

func (a *analyzer) walkTryStatement(node *ast.Node, state *analysisState) {
	tryStmt := node.AsTryStatement()

	tryState := state.clone()
	if tryStmt.TryBlock != nil {
		a.walkNode(tryStmt.TryBlock, tryState)
	}

	// The catch block may be entered at any point during try execution
	// (an exception can occur after an await). Use the post-try state so
	// the catch block sees all reads/outdating from the try block.
	catchState := tryState.clone()
	if tryStmt.CatchClause != nil {
		a.walkNode(tryStmt.CatchClause, catchState)
	}

	tryState.merge(catchState)
	*state = *tryState

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
	a.walkNode(forStmt.Statement, state)
	if forStmt.Incrementor != nil {
		a.walkNode(forStmt.Incrementor, state)
	}
}

// RequireAtomicUpdatesRule disallows assignments that can lead to race conditions
// due to usage of `await` or `yield`.
var RequireAtomicUpdatesRule = rule.Rule{
	Name: "require-atomic-updates",
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
