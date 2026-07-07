package jsx_max_depth

import (
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Bound on the identifier-resolution chain length. Upstream uses ESLint's
// `containDuplicates` check on the scope reference list; we approximate it
// with a (symbol|name)-keyed visited set plus this fixed cap so any
// pathological aliasing — even one our visited set fails to fingerprint —
// terminates.
const maxResolveDepth = 32

const msgWrongDepth = "Expected the depth of nested jsx elements to be <= {{needed}}, but found {{found}}."

const defaultMaxDepth = 2

type options struct {
	max int
}

func parseOptions(raw any) options {
	opts := options{max: defaultMaxDepth}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["max"]; ok {
		// JS-config / CLI numbers come through as float64 (json.Unmarshal default);
		// Go-test maps may pass int / int64 directly. Reject negatives — upstream's
		// JSON schema declares `minimum: 0` and the option only makes sense for
		// non-negative depths.
		switch val := v.(type) {
		case float64:
			if val >= 0 {
				opts.max = int(val)
			}
		case int:
			if val >= 0 {
				opts.max = val
			}
		case int64:
			if val >= 0 {
				opts.max = int(val)
			}
		}
	}
	return opts
}

// isJsxExpressionNode is a nil-safe wrapper over `ast.IsJsxExpression`. tsgo's
// shim predicate dereferences `node.Kind` directly, so callers that may pass
// nil need this guard.
func isJsxExpressionNode(node *ast.Node) bool {
	return node != nil && ast.IsJsxExpression(node)
}

func jsxExpressionInner(node *ast.Node) *ast.Node {
	if !isJsxExpressionNode(node) {
		return nil
	}
	return node.AsJsxExpression().Expression
}

// hasJSX mirrors upstream's
// `jsxUtil.isJSX(node) || (isExpression(node) && jsxUtil.isJSX(node.expression))`:
// a JSX element/fragment, OR a `{...}` whose direct expression is JSX.
// Used to decide whether a child of a JSX container counts as "JSX-bearing".
func hasJSX(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if reactutil.IsJsxLike(node) {
		return true
	}
	if isJsxExpressionNode(node) {
		return reactutil.IsJsxLike(jsxExpressionInner(node))
	}
	return false
}

// isLeaf mirrors upstream's
// `!children || children.length === 0 || !children.some(hasJSX)`. Note that
// `reactutil.GetJsxChildren` returns nil for any non-element/fragment kind
// (JsxSelfClosingElement, JsxExpression, …), so those uniformly count as
// leaves — which is exactly what upstream's `node.children` undefined check
// does on a JSXExpressionContainer or JSXText sibling.
func isLeaf(node *ast.Node) bool {
	children := reactutil.GetJsxChildren(node)
	if len(children) == 0 {
		return true
	}
	for _, c := range children {
		if hasJSX(c) {
			return false
		}
	}
	return true
}

// getDepth counts JSX element/fragment ancestors of `node`, walking through
// JsxExpression containers (which don't add depth themselves but do bridge
// a JSX value across an interpolation boundary). Mirrors upstream's
// `while (jsxUtil.isJSX(node.parent) || isExpression(node.parent))`.
func getDepth(node *ast.Node) int {
	count := 0
	cur := node
	for cur != nil && cur.Parent != nil {
		parent := cur.Parent
		if !reactutil.IsJsxLike(parent) && !isJsxExpressionNode(parent) {
			break
		}
		cur = parent
		if reactutil.IsJsxLike(cur) {
			count++
		}
	}
	return count
}

func report(ctx rule.RuleContext, node *ast.Node, found, needed int) {
	data := map[string]string{
		"found":  strconv.Itoa(found),
		"needed": strconv.Itoa(needed),
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "wrongDepth",
		Description: reactutil.ApplyData(msgWrongDepth, data),
		Data:        data,
	})
}

// findJsxElementOrFragment mirrors upstream's `findJSXElementOrFragment`:
// resolves an Identifier reference to its underlying JSX value by following
// the binding chain, taking the **most recent write** (declaration init OR
// reassignment) in source order at every step.
//
// Cycle detection uses a (symbol-pointer | name)-keyed visited set plus a
// hard depth cap so a `let x = y; let y = x;` pair can't loop. Parens are
// unwrapped because tsgo preserves them as explicit nodes; TS type-expression
// wrappers (`as`, `satisfies`, non-null `!`) are intentionally NOT peeled to
// keep upstream's literal Identifier-only matching.
func findJsxElementOrFragment(ident *ast.Node, tc *checker.Checker, visited map[string]bool, depth int) *ast.Node {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return nil
	}
	if depth >= maxResolveDepth {
		return nil
	}
	name := ident.AsIdentifier().Text
	var key string
	if tc != nil {
		if sym := tc.GetSymbolAtLocation(ident); sym != nil {
			key = fmt.Sprintf("s:%p", sym)
		}
	}
	if key == "" {
		key = "n:" + name
	}
	if visited[key] {
		return nil
	}
	visited[key] = true

	write := resolveLatestWrite(ident, tc)
	if write == nil {
		return nil
	}
	write = ast.SkipParentheses(write)
	if write == nil {
		return nil
	}
	if reactutil.IsJsxLike(write) {
		return write
	}
	if write.Kind == ast.KindIdentifier {
		return findJsxElementOrFragment(write, tc, visited, depth+1)
	}
	return nil
}

// resolveLatestWrite returns the source-order latest write expression for the
// binding referenced by `ident`. Mirrors ESLint scope-manager's
// reverse-iteration over `variable.references`: declaration init counts as
// a write, every direct `name = expr` reassignment counts as a write, and
// the LAST one in source order wins regardless of where the use site sits.
//
// Inner functions / methods are descended into when they DON'T shadow `name`
// (parameter or local declaration with the same identifier) — that captures
// closure writes the way ESLint's scope manager does. Functions that DO
// shadow `name` are skipped because their writes refer to a different binding.
//
// Handles three declaration shapes:
//
//   - VariableDeclaration (`let x = …` / `const x = …` / `var x = …`).
//     Scope is the enclosing Block / SourceFile / ForStatement / etc.
//
//   - Parameter (`function f(x = <jsx>)`). Scope is the function body;
//     the parameter's `Initializer` (default value) seeds the write list
//     so `<jsx>` can be picked up even though it lives outside the body.
//
//   - BindingElement (destructured parameter or destructured variable
//     declarator: `function f({ x = <jsx> })`, `const { x = <jsx> } = …`).
//     Scope follows the enclosing Parameter or VariableDeclaration; the
//     element's `Initializer` (default value) seeds the write list.
//
// Returns nil when:
//   - The binding can't be resolved (free identifier, function declaration,
//     class declaration, import, type-only binding).
//   - The declaration is bound to a non-Identifier name (e.g. a top-level
//     ObjectBindingPattern from `const {x} = …` is followed via its inner
//     BindingElement, not its outer VariableDeclaration).
func resolveLatestWrite(ident *ast.Node, tc *checker.Checker) *ast.Node {
	name := ident.AsIdentifier().Text
	decl := resolveDeclarationOf(ident, name, tc)
	if decl == nil {
		return nil
	}

	var (
		scope          *ast.Node
		seedWrite      *ast.Node
		seedWritePos   = -1
		fallbackResult *ast.Node
	)
	switch decl.Kind {
	case ast.KindVariableDeclaration:
		scope = enclosingScopeContainer(decl)
		fallbackResult = decl.AsVariableDeclaration().Initializer
	case ast.KindParameter:
		scope = parameterScope(decl)
		if init := decl.AsParameterDeclaration().Initializer; init != nil {
			seedWrite = init
			seedWritePos = decl.Pos()
			fallbackResult = init
		}
	case ast.KindBindingElement:
		// Two seed strategies, decided by the BindingElement's enclosing
		// declaration kind:
		//
		//   - Inside a Parameter (`function f({ x = <jsx> })`): ESLint's
		//     scope manager surfaces the BindingElement's default as the
		//     binding's writeExpr. Seed with the default.
		//
		//   - Inside a VariableDeclaration (`const { x = <jsx> } = obj`):
		//     ESLint's scope manager surfaces the VariableDeclaration's
		//     `init` (the destructure RHS), NOT the BindingElement default.
		//     Seed with the RHS, which then resolves through the normal
		//     Identifier-recurse path or bails on a non-Identifier value.
		//     Verified empirically against upstream.
		container := enclosingDeclOrParameter(decl)
		if container == nil {
			scope = bindingElementScope(decl)
			break
		}
		switch container.Kind {
		case ast.KindParameter:
			scope = parameterScope(container)
			if init := decl.AsBindingElement().Initializer; init != nil {
				seedWrite = init
				seedWritePos = decl.Pos()
				fallbackResult = init
			}
		case ast.KindVariableDeclaration:
			scope = enclosingScopeContainer(container)
			if init := container.AsVariableDeclaration().Initializer; init != nil {
				seedWrite = init
				seedWritePos = container.Pos()
				fallbackResult = init
			}
		}
	default:
		return nil
	}
	if scope == nil {
		return fallbackResult
	}

	latest := seedWrite
	latestPos := seedWritePos

	var visit func(n *ast.Node)
	visit = func(n *ast.Node) {
		if n == nil {
			return
		}
		// At a function-like boundary other than the scope itself, descend
		// only when the inner function doesn't shadow `name` (parameter or
		// local declaration). Shadowing means writes inside refer to a
		// different binding and must not bleed into our resolution.
		if n != scope && isFunctionLikeContainer(n) {
			if functionLikeShadowsName(n, name) {
				return
			}
		}
		// Block-level shadowing: `let` / `const` introduce fresh bindings
		// scoped to the enclosing Block (or CaseBlock). When an inner block
		// redeclares `name` with `let` / `const`, every write inside refers
		// to the inner binding and must not bleed into our resolution.
		// `var` declarations are function-scoped and don't shadow at block
		// level, so they're intentionally excluded from this check.
		if n != scope && (n.Kind == ast.KindBlock || n.Kind == ast.KindCaseBlock) {
			if blockShadowsName(n, name) {
				return
			}
		}
		// For-loop init shadowing: `for (let|const x = …; …; …)`,
		// `for (let|const x of …)`, `for (let|const x in …)` introduce a
		// fresh `name` binding scoped to the entire ForStatement. Writes
		// inside the for refer to the inner binding. `var` is excluded
		// (function-hoisted, not block-scoped).
		if n != scope && (n.Kind == ast.KindForStatement ||
			n.Kind == ast.KindForInStatement ||
			n.Kind == ast.KindForOfStatement) {
			if forInitShadowsName(n, name) {
				return
			}
		}
		// Catch-clause parameter shadowing: `try { } catch (name) { … }`
		// binds the caught error to a fresh `name` scoped to the catch.
		// Writes inside the catch refer to the parameter, not an outer
		// same-name binding.
		if n != scope && n.Kind == ast.KindCatchClause {
			if catchClauseShadowsName(n, name) {
				return
			}
		}
		switch n.Kind {
		case ast.KindVariableDeclaration:
			vd := n.AsVariableDeclaration()
			if vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier &&
				vd.Name().AsIdentifier().Text == name && vd.Initializer != nil {
				if pos := n.Pos(); pos > latestPos {
					latestPos = pos
					latest = vd.Initializer
				}
			}
		case ast.KindBinaryExpression:
			bin := n.AsBinaryExpression()
			if bin.OperatorToken != nil &&
				bin.OperatorToken.Kind == ast.KindEqualsToken &&
				bin.Left != nil &&
				bin.Left.Kind == ast.KindIdentifier &&
				bin.Left.AsIdentifier().Text == name {
				if pos := n.Pos(); pos > latestPos {
					latestPos = pos
					latest = bin.Right
				}
			}
		}
		n.ForEachChild(func(c *ast.Node) bool {
			visit(c)
			return false
		})
	}
	visit(scope)
	return latest
}

// resolveDeclarationOf returns the declaration node that defines the
// binding referenced by `ident`. Tries the TypeChecker first (the only
// path that resolves cross-module imports correctly) and falls back to a
// local scan over enclosing scopes. Recognizes three declaration kinds:
//
//   - VariableDeclaration (`var/let/const name = …`)
//   - Parameter (`function f(name = …)`)
//   - BindingElement (destructured: `function f({ name = … })` or
//     `const { name } = …`)
//
// Returns nil when the binding can't be resolved or the declaration kind
// isn't one of those three (e.g. function declaration, class declaration,
// import specifier).
func resolveDeclarationOf(ident *ast.Node, name string, tc *checker.Checker) *ast.Node {
	if tc != nil {
		if sym := tc.GetSymbolAtLocation(ident); sym != nil {
			d := sym.ValueDeclaration
			if d == nil && len(sym.Declarations) > 0 {
				d = sym.Declarations[0]
			}
			if d != nil {
				switch d.Kind {
				case ast.KindVariableDeclaration:
					// tsgo's TypeChecker returns the outer VariableDeclaration
					// even for destructured bindings (`const { x } = obj` →
					// ValueDeclaration is the whole `const`). Drill into the
					// binding pattern when the declaration's name doesn't
					// match `name` directly.
					if found := matchVarDeclBinding(d, name); found != nil {
						return found
					}
				case ast.KindParameter:
					if pn := d.AsParameterDeclaration().Name(); pn != nil &&
						pn.Kind == ast.KindIdentifier &&
						pn.AsIdentifier().Text == name {
						return d
					}
					// Destructured parameter: drill in.
					if pn := d.AsParameterDeclaration().Name(); pn != nil {
						if be := findBindingElementByName(pn, name); be != nil {
							return be
						}
					}
				case ast.KindBindingElement:
					return d
				}
			}
		}
	}
	for cur := ident.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindCaseBlock, ast.KindModuleBlock:
			if d := findVarDeclByName(cur, name); d != nil {
				return d
			}
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
			// `for (let x = ...; ...; ...)`, `for (let x of …)`, and
			// `for (let x in …)` all bind `x` to the loop's scope, not
			// the enclosing Block. The init list lives on the loop's
			// Initializer slot — check it directly.
			if d := findVarDeclInForInit(cur, name); d != nil {
				return d
			}
		}
		// Function-like containers introduce parameter bindings scoped to
		// the function body. The closest such container that binds `name`
		// wins (innermost shadow rule).
		if isFunctionLikeContainer(cur) {
			if d := findParamByName(cur, name); d != nil {
				return d
			}
		}
	}
	return nil
}

// findParamByName searches the parameter list of `funcLike` for a binding
// named `name`. Returns the Parameter when bound directly as an Identifier,
// or the BindingElement when bound via an Object / Array destructure
// pattern. Returns nil when no parameter binds `name`.
func findParamByName(funcLike *ast.Node, name string) *ast.Node {
	for _, p := range functionParameters(funcLike) {
		if p == nil || p.Kind != ast.KindParameter {
			continue
		}
		pn := p.AsParameterDeclaration().Name()
		if pn == nil {
			continue
		}
		if pn.Kind == ast.KindIdentifier {
			if pn.AsIdentifier().Text == name {
				return p
			}
			continue
		}
		if be := findBindingElementByName(pn, name); be != nil {
			return be
		}
	}
	return nil
}

// findBindingElementByName recursively searches an Object / Array binding
// pattern for a BindingElement bound to a bare Identifier matching `name`.
// Nested patterns are followed (e.g. `{ a: { name } }`). Rest elements
// (`{ ...name }`) are also matched when their target is an Identifier.
func findBindingElementByName(pattern *ast.Node, name string) *ast.Node {
	if pattern == nil {
		return nil
	}
	switch pattern.Kind {
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
	default:
		return nil
	}
	bp := pattern.AsBindingPattern()
	if bp == nil || bp.Elements == nil {
		return nil
	}
	for _, el := range bp.Elements.Nodes {
		if el == nil || el.Kind != ast.KindBindingElement {
			continue
		}
		be := el.AsBindingElement()
		if be == nil {
			continue
		}
		beName := be.Name()
		if beName == nil {
			continue
		}
		if beName.Kind == ast.KindIdentifier {
			if beName.AsIdentifier().Text == name {
				return el
			}
			continue
		}
		if nested := findBindingElementByName(beName, name); nested != nil {
			return nested
		}
	}
	return nil
}

// parameterScope returns the function body Block that scopes the parameter's
// binding. Returns nil when the Parameter isn't actually inside a
// function-like container (shouldn't happen in well-formed AST).
func parameterScope(parameter *ast.Node) *ast.Node {
	for cur := parameter.Parent; cur != nil; cur = cur.Parent {
		if isFunctionLikeContainer(cur) {
			return functionBody(cur)
		}
	}
	return nil
}

// bindingElementScope returns the scope of a BindingElement based on whether
// it sits inside a Parameter or a VariableDeclaration. Walks up until either
// is found.
func bindingElementScope(be *ast.Node) *ast.Node {
	if c := enclosingDeclOrParameter(be); c != nil {
		switch c.Kind {
		case ast.KindParameter:
			return parameterScope(c)
		case ast.KindVariableDeclaration:
			return enclosingScopeContainer(c)
		}
	}
	return nil
}

// enclosingDeclOrParameter walks up from a BindingElement to find the nearest
// Parameter or VariableDeclaration that anchors its scope and seed value.
// Used by both `bindingElementScope` and the seed-strategy switch in
// `resolveLatestWrite`.
func enclosingDeclOrParameter(be *ast.Node) *ast.Node {
	for cur := be.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindParameter, ast.KindVariableDeclaration:
			return cur
		}
	}
	return nil
}

// findVarDeclInForInit returns the binding for `name` declared in a for-loop
// initializer (`for (let|const name = …; …; …)`, `for (let|const name of …)`,
// `for (let|const { name } of …)`). Returns the VariableDeclaration for bare
// Identifier bindings or the inner BindingElement for destructure patterns.
func findVarDeclInForInit(forNode *ast.Node, name string) *ast.Node {
	var initList *ast.Node
	switch forNode.Kind {
	case ast.KindForStatement:
		initList = forNode.AsForStatement().Initializer
	case ast.KindForInStatement:
		initList = forNode.AsForInOrOfStatement().Initializer
	case ast.KindForOfStatement:
		initList = forNode.AsForInOrOfStatement().Initializer
	}
	if initList == nil || initList.Kind != ast.KindVariableDeclarationList {
		return nil
	}
	decls := initList.AsVariableDeclarationList()
	if decls == nil || decls.Declarations == nil {
		return nil
	}
	for _, d := range decls.Declarations.Nodes {
		if found := matchVarDeclBinding(d, name); found != nil {
			return found
		}
	}
	return nil
}

// findVarDeclByName scans `scope`'s direct VariableStatement children for a
// VariableDeclaration with an Identifier-named binding equal to `name`.
// For CaseBlock the scan descends into each clause's Statements (per
// ES2015 `let` / `const` are scoped to the entire CaseBlock, so a binding
// in `case 'a':` is visible to all sibling cases).
//
// For other scope kinds, does NOT descend — callers do their own walk if
// recursive lookup is needed.
func findVarDeclByName(scope *ast.Node, name string) *ast.Node {
	var found *ast.Node
	if scope != nil && scope.Kind == ast.KindCaseBlock {
		scope.ForEachChild(func(clause *ast.Node) bool {
			if found != nil || clause == nil {
				return found != nil
			}
			clause.ForEachChild(func(stmt *ast.Node) bool {
				if found != nil {
					return true
				}
				if d := findVarDeclByNameInStmt(stmt, name); d != nil {
					found = d
					return true
				}
				return false
			})
			return found != nil
		})
		return found
	}
	scope.ForEachChild(func(stmt *ast.Node) bool {
		if found != nil {
			return true
		}
		if stmt == nil {
			return false
		}
		if d := findVarDeclByNameInStmt(stmt, name); d != nil {
			found = d
			return true
		}
		return false
	})
	return found
}

// findVarDeclByNameInStmt extracts the per-statement check used by
// findVarDeclByName (and its CaseBlock descend). Returns:
//   - the VariableDeclaration when `stmt` declares `name` via a bare
//     Identifier binding (`let name = …`)
//   - the BindingElement when `stmt` declares `name` via a destructure
//     pattern (`let { name = … } = …`, `let [name] = …`, including nested)
//   - nil otherwise.
func findVarDeclByNameInStmt(stmt *ast.Node, name string) *ast.Node {
	if stmt == nil || stmt.Kind != ast.KindVariableStatement {
		return nil
	}
	list := stmt.AsVariableStatement().DeclarationList
	if list == nil {
		return nil
	}
	decls := list.AsVariableDeclarationList()
	if decls == nil || decls.Declarations == nil {
		return nil
	}
	for _, d := range decls.Declarations.Nodes {
		if d == nil || d.Kind != ast.KindVariableDeclaration {
			continue
		}
		if found := matchVarDeclBinding(d, name); found != nil {
			return found
		}
	}
	return nil
}

// matchVarDeclBinding returns:
//   - `vd` itself when the declaration's name is the bare Identifier `name`.
//   - the inner BindingElement when the declaration uses a destructure
//     pattern that binds `name` (recursive search via findBindingElementByName).
//   - nil otherwise.
func matchVarDeclBinding(vd *ast.Node, name string) *ast.Node {
	if vd == nil || vd.Kind != ast.KindVariableDeclaration {
		return nil
	}
	n := vd.AsVariableDeclaration().Name()
	if n == nil {
		return nil
	}
	if n.Kind == ast.KindIdentifier {
		if n.AsIdentifier().Text == name {
			return vd
		}
		return nil
	}
	return findBindingElementByName(n, name)
}

// enclosingScopeContainer returns the AST node that owns `node`'s binding's
// lexical scope. Walks up the parent chain looking for the nearest
// scope-introducing kind:
//
//   - For a VariableDeclaration declared in a for-loop initializer
//     (`for (let x = …; …; …)`, `for (let x of …)`, `for (let x in …)`),
//     the loop statement itself is the scope — `let` / `const` introduced
//     in for-init are scoped to the entire ForStatement (init + condition
//   - incrementor + body), not the enclosing Block. Returning the outer
//     Block would let us see same-name writes that don't refer to this
//     binding and miss reassignments in the loop body.
//   - Otherwise, the nearest enclosing Block / SourceFile / ModuleBlock /
//     CaseBlock.
func enclosingScopeContainer(node *ast.Node) *ast.Node {
	if node != nil && node.Parent != nil &&
		node.Parent.Kind == ast.KindVariableDeclarationList {
		if gp := node.Parent.Parent; gp != nil {
			switch gp.Kind {
			case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
				return gp
			}
		}
	}
	for cur := node.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock, ast.KindCaseBlock:
			return cur
		}
	}
	return nil
}

// isFunctionLikeContainer reports whether `n` introduces its own variable
// scope. Class static blocks are included because their bodies declare
// fresh `let`/`const` bindings independent of the enclosing class.
func isFunctionLikeContainer(n *ast.Node) bool {
	switch n.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor,
		ast.KindClassStaticBlockDeclaration:
		return true
	}
	return false
}

// forInitShadowsName reports whether the `for` / `for-in` / `for-of`
// statement's initializer declares a `let` / `const` binding for `name`.
// Such bindings are scoped to the whole ForStatement, so any write to
// `name` inside the loop refers to the inner binding — not an outer
// same-name. `var` initializers are skipped because they hoist to the
// enclosing function scope.
func forInitShadowsName(forNode *ast.Node, name string) bool {
	var initList *ast.Node
	switch forNode.Kind {
	case ast.KindForStatement:
		initList = forNode.AsForStatement().Initializer
	case ast.KindForInStatement, ast.KindForOfStatement:
		initList = forNode.AsForInOrOfStatement().Initializer
	}
	if initList == nil || initList.Kind != ast.KindVariableDeclarationList {
		return false
	}
	decls := initList.AsVariableDeclarationList()
	if decls == nil || decls.Declarations == nil {
		return false
	}
	if decls.Flags&ast.NodeFlagsBlockScoped == 0 {
		return false
	}
	for _, d := range decls.Declarations.Nodes {
		if d == nil || d.Kind != ast.KindVariableDeclaration {
			continue
		}
		vd := d.AsVariableDeclaration()
		if vd.Name() != nil && bindingNameContains(vd.Name(), name) {
			return true
		}
	}
	return false
}

// catchClauseShadowsName reports whether the `catch (…)` clause binds
// `name` (either as a bare Identifier `catch (name)` or as a destructure
// pattern `catch ({ name })`). Bare `catch {}` (no binding) returns false.
func catchClauseShadowsName(catchClause *ast.Node, name string) bool {
	if catchClause == nil || catchClause.Kind != ast.KindCatchClause {
		return false
	}
	cc := catchClause.AsCatchClause()
	if cc == nil || cc.VariableDeclaration == nil {
		return false
	}
	vd := cc.VariableDeclaration
	if vd.Kind != ast.KindVariableDeclaration {
		return false
	}
	bn := vd.AsVariableDeclaration().Name()
	if bn == nil {
		return false
	}
	return bindingNameContains(bn, name)
}

// blockShadowsName reports whether `block` (a Block / CaseBlock) introduces
// a `let` / `const` binding named `name`. `var` declarations are skipped
// because they hoist to the enclosing function scope and don't shadow at
// block level.
func blockShadowsName(block *ast.Node, name string) bool {
	var found bool
	visitVariableStatements := func(stmt *ast.Node) bool {
		if stmt == nil || stmt.Kind != ast.KindVariableStatement {
			return false
		}
		list := stmt.AsVariableStatement().DeclarationList
		if list == nil {
			return false
		}
		// Only block-scoped (`let` / `const`) declarations shadow.
		if list.Flags&ast.NodeFlagsBlockScoped == 0 {
			return false
		}
		decls := list.AsVariableDeclarationList()
		if decls == nil || decls.Declarations == nil {
			return false
		}
		for _, d := range decls.Declarations.Nodes {
			if d == nil || d.Kind != ast.KindVariableDeclaration {
				continue
			}
			vd := d.AsVariableDeclaration()
			if vd.Name() != nil && bindingNameContains(vd.Name(), name) {
				found = true
				return true
			}
		}
		return false
	}
	switch block.Kind {
	case ast.KindBlock:
		block.ForEachChild(func(stmt *ast.Node) bool {
			if found {
				return true
			}
			return visitVariableStatements(stmt)
		})
	case ast.KindCaseBlock:
		// CaseBlock contains CaseClause / DefaultClause; each holds a
		// Statements list. Treat all clauses as one shared scope, matching
		// JS semantics for `case` falls-through.
		block.ForEachChild(func(clause *ast.Node) bool {
			if found || clause == nil {
				return found
			}
			clause.ForEachChild(func(stmt *ast.Node) bool {
				if found {
					return true
				}
				return visitVariableStatements(stmt)
			})
			return found
		})
	}
	return found
}

// functionLikeShadowsName reports whether `funcLike` introduces a binding
// named `name` that would shadow an outer binding. Checks parameters and
// top-level `var/let/const` declarations in the function body — covers the
// cases where descending into the body would attribute writes to the wrong
// binding.
func functionLikeShadowsName(funcLike *ast.Node, name string) bool {
	for _, p := range functionParameters(funcLike) {
		if parameterBindsName(p, name) {
			return true
		}
	}
	body := functionBody(funcLike)
	if body != nil && findVarDeclByName(body, name) != nil {
		return true
	}
	return false
}

func functionParameters(funcLike *ast.Node) []*ast.Node {
	switch funcLike.Kind {
	case ast.KindFunctionDeclaration:
		if fd := funcLike.AsFunctionDeclaration(); fd.Parameters != nil {
			return fd.Parameters.Nodes
		}
	case ast.KindFunctionExpression:
		if fe := funcLike.AsFunctionExpression(); fe.Parameters != nil {
			return fe.Parameters.Nodes
		}
	case ast.KindArrowFunction:
		if af := funcLike.AsArrowFunction(); af.Parameters != nil {
			return af.Parameters.Nodes
		}
	case ast.KindMethodDeclaration:
		if md := funcLike.AsMethodDeclaration(); md.Parameters != nil {
			return md.Parameters.Nodes
		}
	case ast.KindGetAccessor:
		if ga := funcLike.AsGetAccessorDeclaration(); ga.Parameters != nil {
			return ga.Parameters.Nodes
		}
	case ast.KindSetAccessor:
		if sa := funcLike.AsSetAccessorDeclaration(); sa.Parameters != nil {
			return sa.Parameters.Nodes
		}
	case ast.KindConstructor:
		if cd := funcLike.AsConstructorDeclaration(); cd.Parameters != nil {
			return cd.Parameters.Nodes
		}
	}
	return nil
}

// parameterBindsName recognises the common shapes — bare Identifier and
// `{ name }` / `[ name ]` destructuring pattern — and returns true when any
// of them binds `name`. Rest parameters (`...rest`) and assignment patterns
// (`name = default`) are also covered because they wrap an Identifier or
// pattern that we descend through.
func parameterBindsName(p *ast.Node, name string) bool {
	if p == nil || p.Kind != ast.KindParameter {
		return false
	}
	pn := p.AsParameterDeclaration().Name()
	return bindingNameContains(pn, name)
}

func bindingNameContains(n *ast.Node, name string) bool {
	if n == nil {
		return false
	}
	switch n.Kind {
	case ast.KindIdentifier:
		return n.AsIdentifier().Text == name
	case ast.KindObjectBindingPattern:
		obp := n.AsBindingPattern()
		if obp == nil || obp.Elements == nil {
			return false
		}
		for _, el := range obp.Elements.Nodes {
			if el == nil {
				continue
			}
			be := el.AsBindingElement()
			if be == nil {
				continue
			}
			if bindingNameContains(be.Name(), name) {
				return true
			}
		}
	case ast.KindArrayBindingPattern:
		abp := n.AsBindingPattern()
		if abp == nil || abp.Elements == nil {
			return false
		}
		for _, el := range abp.Elements.Nodes {
			if el == nil || el.Kind != ast.KindBindingElement {
				continue
			}
			be := el.AsBindingElement()
			if be == nil {
				continue
			}
			if bindingNameContains(be.Name(), name) {
				return true
			}
		}
	}
	return false
}

func functionBody(funcLike *ast.Node) *ast.Node {
	switch funcLike.Kind {
	case ast.KindFunctionDeclaration:
		return funcLike.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		return funcLike.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		body := funcLike.AsArrowFunction().Body
		if body != nil && body.Kind == ast.KindBlock {
			return body
		}
	case ast.KindMethodDeclaration:
		return funcLike.AsMethodDeclaration().Body
	case ast.KindGetAccessor:
		return funcLike.AsGetAccessorDeclaration().Body
	case ast.KindSetAccessor:
		return funcLike.AsSetAccessorDeclaration().Body
	case ast.KindConstructor:
		return funcLike.AsConstructorDeclaration().Body
	case ast.KindClassStaticBlockDeclaration:
		return funcLike.AsClassStaticBlockDeclaration().Body
	}
	return nil
}

// checkDescendant mirrors upstream verbatim: walk JSX-bearing children, report
// each one whose effective depth (baseDepth + 1, then recursively) exceeds
// the configured max, and recurse only into JsxElement / JsxFragment children
// that are non-leaves. JsxExpressionContainer children are treated as leaves
// (their `.children` is undefined in ESTree), so we do NOT unwrap them and
// keep walking — doing so would silently fire on shapes upstream never
// reports.
func checkDescendant(ctx rule.RuleContext, baseDepth, maxDepth int, children []*ast.Node) {
	baseDepth++
	for _, child := range children {
		if !hasJSX(child) {
			continue
		}
		if baseDepth > maxDepth {
			report(ctx, child, baseDepth, maxDepth)
			continue
		}
		// `isLeaf` returns true for any non-element/fragment kind (including
		// JsxExpressionContainer with JSX inner), so the recursion stops at
		// the same boundaries upstream's `node.children` does.
		if !isLeaf(child) {
			checkDescendant(ctx, baseDepth, maxDepth, reactutil.GetJsxChildren(child))
		}
	}
}

var JsxMaxDepthRule = rule.Rule{
	Name: "react/jsx-max-depth",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.UnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)

		handleJSX := func(node *ast.Node) {
			// Only the OUTERMOST leaf in a leaf chain reports here; non-leaf
			// containers are handled implicitly by their descendants firing
			// the same listener. This mirrors upstream's `if (!isLeaf(node)) return`
			// and avoids double-counting any container's depth.
			if !isLeaf(node) {
				return
			}
			d := getDepth(node)
			if d > opts.max {
				report(ctx, node, d, opts.max)
			}
		}

		return rule.RuleListeners{
			ast.KindJsxElement:            handleJSX,
			ast.KindJsxSelfClosingElement: handleJSX,
			ast.KindJsxFragment:           handleJSX,
			ast.KindJsxExpression: func(node *ast.Node) {
				// Upstream gates on `node.expression.type === 'Identifier'`.
				// tsgo preserves Parens as explicit nodes — peel them so
				// `{(x)}` resolves identically to `{x}` (espree flattens parens
				// at the AST level so upstream sees them the same way). TS-only
				// wrappers (`x as T`, `x!`, `x satisfies T`) are intentionally
				// NOT peeled to keep upstream's literal Identifier-only match.
				expr := jsxExpressionInner(node)
				if expr == nil {
					return
				}
				expr = ast.SkipParentheses(expr)
				if expr == nil || expr.Kind != ast.KindIdentifier {
					return
				}
				element := findJsxElementOrFragment(expr, ctx.TypeChecker, map[string]bool{}, 0)
				if element == nil {
					return
				}
				baseDepth := getDepth(node)
				checkDescendant(ctx, baseDepth, opts.max, reactutil.GetJsxChildren(element))
			},
		}
	},
}
