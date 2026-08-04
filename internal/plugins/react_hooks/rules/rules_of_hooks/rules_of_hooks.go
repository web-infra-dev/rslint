package rules_of_hooks

import (
	_ "embed"
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// The schema declares `additionalHooks` exactly as upstream does. Since
// facebook/react#37085 upstream's create() reads it: the rule-level option
// takes precedence, `settings['react-hooks'].additionalEffectHooks` is the
// fallback.
//
//go:embed rules_of_hooks.schema.json
var schemaJSON []byte

// flowSuppressionRegex matches `$FlowFixMe[react-rule-hook]` comments. Used to
// gate `hasFlowSuppression`, which mirrors upstream's same-named helper —
// suppress the diagnostic when an immediately-preceding line carries the
// suppression marker.
var flowSuppressionRegex = regexp.MustCompile(`\$FlowFixMe\[react-rule-hook\]`)

// Aliases over `react_hooksutil` so the call sites in this file stay terse
// while every predicate's authoritative definition lives in the shared
// package. See `react_hooksutil` package docs for semantics.
var (
	isHookCallee    = react_hooksutil.IsHookCallee
	isUseIdentifier = react_hooksutil.IsUseIdentifier
)

// All AST predicates above (stripReactNamespace, isFunctionLikeContainer,
// findEnclosingFunction, getFunctionBody, hasAsyncModifier, ...) live in
// `react_hooksutil`. The aliases below keep the call sites unchanged
// while removing the second copy.
var (
	stripReactNamespace     = react_hooksutil.StripReactNamespace
	isFunctionLikeContainer = react_hooksutil.IsFunctionLikeContainer
	findEnclosingFunction   = react_hooksutil.FindEnclosingFunction
	getFunctionBody         = react_hooksutil.GetFunctionBody
	hasAsyncModifier        = react_hooksutil.HasAsyncModifier
)

// isEffectCalleeName reports whether `node` (post-namespace-strip) names one
// of the built-in effect hooks (`useEffect` / `useLayoutEffect` /
// `useInsertionEffect`) or matches the user-configured `additionalEffectHooks`
// regex from settings. Specific to rules-of-hooks (the additional-effect-hooks
// gate is exclusive to this rule), so it's not promoted to the shared util.
func isEffectCalleeName(node *ast.Node, additional *regexp.Regexp) bool {
	n := stripReactNamespace(node)
	if n == nil || n.Kind != ast.KindIdentifier {
		return false
	}
	name := n.AsIdentifier().Text
	switch name {
	case "useEffect", "useLayoutEffect", "useInsertionEffect":
		return true
	}
	if additional != nil {
		return additional.MatchString(name)
	}
	return false
}

// isUseEffectEventCallee delegates to the shared helper.
func isUseEffectEventCallee(node *ast.Node) bool {
	return react_hooksutil.IsUseEffectEventCallee(node)
}

// isInsideTryCatchOfFunction reports whether `node` is inside a TryStatement
// or CatchClause within `fn` (the enclosing function-like). Mirrors
// upstream's `isInsideTryCatch(hook)`.
func isInsideTryCatchOfFunction(node *ast.Node, fn *ast.Node) bool {
	if node == nil {
		return false
	}
	cur := node.Parent
	for cur != nil && cur != fn {
		switch cur.Kind {
		case ast.KindTryStatement, ast.KindCatchClause:
			return true
		}
		if isFunctionLikeContainer(cur) {
			return false
		}
		cur = cur.Parent
	}
	return false
}

// isInsideLoopOfFunction reports whether `node` is inside any loop construct
// (while / for / for-in / for-of / do/while) within `fn`. Approximates
// upstream's CFG `cycled` set: any segment reachable through a back-edge.
//
// Position-aware for the C-style and for-in/of variants:
//   - `for (init; cond; incr) body`: `init` runs once BEFORE the loop and
//     is NOT cycled; the condition / incrementor / body all are.
//   - `for (var of expr) body`: `expr` is evaluated once before iteration
//     starts (per ECMA-262) and is NOT cycled; `var`'s binding pattern
//     re-binds per iteration and the body is cycled.
//
// Without the position split, hooks placed in `for (let x = useFoo(); ...; )`
// or `for (const x of useArr()) {}` would be misclassified as cycled and
// emit a spurious loopError.
func isInsideLoopOfFunction(node *ast.Node, fn *ast.Node) bool {
	if node == nil {
		return false
	}
	cur := node
	for cur != nil && cur.Parent != nil && cur.Parent != fn {
		parent := cur.Parent
		switch parent.Kind {
		case ast.KindDoStatement, ast.KindWhileStatement:
			return true
		case ast.KindForStatement:
			fs := parent.AsForStatement()
			if cur != fs.Initializer {
				return true
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			fos := parent.AsForInOrOfStatement()
			if cur != fos.Expression {
				return true
			}
		}
		if isFunctionLikeContainer(parent) {
			return false
		}
		cur = parent
	}
	return false
}

// All function-name / forwardRef / classifier predicates live in the
// shared `react_hooksutil` package. The aliases below preserve the
// existing call sites unchanged.
var (
	getFunctionName         = react_hooksutil.GetFunctionName
	isComponentOrHookFn     = react_hooksutil.IsComponentOrHookFn
	isInsideComponentOrHook = react_hooksutil.IsInsideComponentOrHook
	isClassMember           = react_hooksutil.IsClassMember
)

// isConditionalAncestor reports whether the position of `child` within `parent`
// places it on a conditional execution path (only entered on certain
// runtime branches). Used by `isConditional` during the parent-walk.
//
// NOTE: TryStatement is handled by `isInsideTryBlockWithPriorStmt` instead
// — being inside a try block is only conditional if a prior sibling could
// throw before the hook is reached.
func isConditionalAncestor(parent *ast.Node, child *ast.Node) bool {
	switch parent.Kind {
	case ast.KindIfStatement:
		is := parent.AsIfStatement()
		// `if (cond) then else` — `then` and `else` are conditional; the
		// condition expression itself is unconditional and is intentionally
		// not flagged.
		return child == is.ThenStatement || child == is.ElseStatement
	case ast.KindConditionalExpression:
		ce := parent.AsConditionalExpression()
		return child == ce.WhenTrue || child == ce.WhenFalse
	case ast.KindBinaryExpression:
		be := parent.AsBinaryExpression()
		if be.OperatorToken == nil {
			return false
		}
		switch be.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			// Right operand of `&&` / `||` / `??` is conditional; the left
			// is always evaluated.
			return child == be.Right
		}
		return false
	case ast.KindWhileStatement:
		// The condition is always evaluated at least once, so we don't flag
		// `while (useFoo()) {}` here — but the body is conditional. The
		// "in-loop" detection separately reports it as a loop violation,
		// which takes precedence.
		return child == parent.AsWhileStatement().Statement
	case ast.KindForStatement:
		fs := parent.AsForStatement()
		// init runs once before the loop and is unconditional. The body
		// is conditional (and the in-loop detection takes precedence).
		// Condition / incrementor: also conditional in the sense that they
		// re-execute per iteration — but the in-loop detection covers them.
		return child == fs.Statement
	case ast.KindForInStatement, ast.KindForOfStatement:
		fos := parent.AsForInOrOfStatement()
		// `for (const x of expr) body` — `expr` is evaluated once before
		// iteration starts (ECMA-262 step 1 of ForIn/OfBodyEvaluation),
		// so it's NOT conditional. The body / binding pattern run per
		// iteration and are loop-conditional (handled by in-loop detection).
		return child != fos.Expression
	case ast.KindCaseClause, ast.KindDefaultClause:
		// Any descendant of a switch case body is conditional.
		return true
	case ast.KindCatchClause:
		// Catch body only executes on an exception.
		return true
	}
	return false
}

// isInsideTryBlockWithPriorStmt reports whether `node` sits inside a try
// block AND has at least one preceding sibling statement in that block.
// Mirrors upstream's CFG behavior: every statement in a try block has a
// "could throw" exit, so a hook reached only after a prior throwable
// statement is conditionally executed. A hook that is the very first
// statement in a try block (no prior statement) is unconditional in
// upstream's path counting and is NOT flagged here.
func isInsideTryBlockWithPriorStmt(node *ast.Node, fn *ast.Node) bool {
	if node == nil {
		return false
	}
	cur := node
	for cur != nil && cur.Parent != nil && cur.Parent != fn {
		parent := cur.Parent
		if parent.Kind == ast.KindBlock {
			grand := parent.Parent
			if grand != nil && grand.Kind == ast.KindTryStatement {
				ts := grand.AsTryStatement()
				if ts.TryBlock != nil && ts.TryBlock.AsNode() == parent {
					block := parent.AsBlock()
					if block.Statements != nil {
						for _, stmt := range block.Statements.Nodes {
							if stmt == cur {
								return false
							}
							if stmt.Pos() >= cur.Pos() {
								return false
							}
							return true
						}
					}
					return false
				}
			}
		}
		if isFunctionLikeContainer(parent) {
			return false
		}
		cur = parent
	}
	return false
}

// isConditional walks the parent chain from `node` up to (but not crossing)
// `fn`. Returns true once any conditional placement is observed.
// Approximation of upstream's `pathsFromStartToEnd !== allPathsFromStartToEnd`
// signal for the structural cases; sibling-based early return / labeled break
// detection lives in `hasEarlyReturnBefore` and `hasLabeledBreakBefore`.
func isConditional(node *ast.Node, fn *ast.Node) bool {
	cur := node
	for cur != nil && cur.Parent != nil && cur != fn {
		if isConditionalAncestor(cur.Parent, cur) {
			return true
		}
		cur = cur.Parent
	}
	return false
}

// containsEarlyReturn reports whether `node` (recursively, but never crossing
// a nested function-like) contains a ReturnStatement. ThrowStatement is
// intentionally excluded: upstream's CFG models thrown segments as
// `thrownSegments`, which are excluded from path counting — so a hook
// reached only when an exception did NOT throw is treated as unconditional.
// Mirroring that, throw-inside-if does not trigger the "early return" suffix
// here.
func containsEarlyReturn(node *ast.Node) bool {
	if node == nil {
		return false
	}
	found := false
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if found {
			return true
		}
		if n.Kind == ast.KindReturnStatement {
			found = true
			return true
		}
		if isFunctionLikeContainer(n) {
			return false
		}
		n.ForEachChild(visit)
		return false
	}
	visit(node)
	return found
}

// isPrecededByDirectTerminator reports whether `hookNode` lies after an
// unconditional `return` or `throw` statement in the same enclosing block
// (recursively walking up to the function body). Such a hook is unreachable
// in upstream's CFG (`!segment.reachable`) and the rule loop `continue`s
// without checking — we model that by short-circuiting all reports here.
func isPrecededByDirectTerminator(hookNode *ast.Node, fn *ast.Node) bool {
	if hookNode == nil {
		return false
	}
	cur := hookNode
	for cur != nil && cur.Parent != nil && cur.Parent != fn {
		parent := cur.Parent
		if parent.Kind == ast.KindBlock {
			block := parent.AsBlock()
			if block.Statements != nil {
				for _, stmt := range block.Statements.Nodes {
					if stmt == cur {
						break
					}
					if stmt.Pos() >= cur.Pos() {
						break
					}
					switch stmt.Kind {
					case ast.KindReturnStatement, ast.KindThrowStatement:
						return true
					}
				}
			}
		}
		if isFunctionLikeContainer(parent) {
			break
		}
		cur = parent
	}
	return false
}

// hasEarlyReturnBefore walks up from `hookNode` through enclosing Block
// ancestors (stopping at the function body) and, at each level, checks
// preceding siblings for early-return markers (return inside any
// non-function-like child). Triggers the "early return" suffix on
// `conditionalError`.
func hasEarlyReturnBefore(hookNode *ast.Node, fn *ast.Node) bool {
	if hookNode == nil || fn == nil {
		return false
	}
	body := getFunctionBody(fn)
	cur := hookNode
	for cur != nil && cur.Parent != nil && cur.Parent != fn {
		parent := cur.Parent
		if parent.Kind == ast.KindBlock {
			block := parent.AsBlock()
			if block.Statements != nil {
				for _, stmt := range block.Statements.Nodes {
					if stmt == cur {
						break
					}
					if stmt.Pos() >= cur.Pos() {
						break
					}
					if containsEarlyReturn(stmt) {
						return true
					}
				}
			}
		}
		if parent == body {
			break
		}
		cur = parent
	}
	return false
}

// collectEnclosingLabels walks up from `node` and returns a set of
// LabeledStatement labels enclosing `node` within `fn`. Used by
// `hasLabeledBreakBefore` to recognize `break <label>` references.
func collectEnclosingLabels(node *ast.Node, fn *ast.Node) map[string]bool {
	var labels map[string]bool
	p := node.Parent
	for p != nil && p != fn {
		if p.Kind == ast.KindLabeledStatement {
			ls := p.AsLabeledStatement()
			if ls.Label != nil && ls.Label.Kind == ast.KindIdentifier {
				if labels == nil {
					labels = make(map[string]bool)
				}
				labels[ls.Label.AsIdentifier().Text] = true
			}
		}
		if isFunctionLikeContainer(p) {
			break
		}
		p = p.Parent
	}
	return labels
}

// containsLabeledBreak recursively looks for a BreakStatement whose label
// matches one of `labels`, without descending into nested function-likes
// (where the break would refer to its own enclosing label scope).
func containsLabeledBreak(node *ast.Node, labels map[string]bool) bool {
	if node == nil || len(labels) == 0 {
		return false
	}
	found := false
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		if found {
			return true
		}
		if n.Kind == ast.KindBreakStatement {
			bs := n.AsBreakStatement()
			if bs.Label != nil && bs.Label.Kind == ast.KindIdentifier {
				if labels[bs.Label.AsIdentifier().Text] {
					found = true
					return true
				}
			}
		}
		if isFunctionLikeContainer(n) {
			return false
		}
		n.ForEachChild(visit)
		return false
	}
	visit(node)
	return found
}

// hasLabeledBreakBefore reports whether `hookNode` lies after a `break <label>`
// targeting an enclosing LabeledStatement. Used to catch the
// `label: { if (cond) break label; useFoo(); }` family of conditional patterns
// that upstream detects via CFG cycle / segment analysis but are otherwise
// invisible to the parent-walk version of `isConditional`.
func hasLabeledBreakBefore(hookNode *ast.Node, fn *ast.Node) bool {
	if hookNode == nil || fn == nil {
		return false
	}
	labels := collectEnclosingLabels(hookNode, fn)
	if len(labels) == 0 {
		return false
	}
	cur := hookNode
	for cur != nil && cur.Parent != nil && cur.Parent != fn {
		parent := cur.Parent
		if parent.Kind == ast.KindBlock {
			block := parent.AsBlock()
			if block.Statements != nil {
				for _, stmt := range block.Statements.Nodes {
					if stmt == cur {
						break
					}
					if stmt.Pos() >= cur.Pos() {
						break
					}
					if containsLabeledBreak(stmt, labels) {
						return true
					}
				}
			}
		}
		if isFunctionLikeContainer(parent) {
			break
		}
		cur = parent
	}
	return false
}

// nameText returns the source text for a function-name node, or "" when nil.
// Used to format `functionError` messages with the enclosing function's name.
func nameText(node *ast.Node, sf *ast.SourceFile) string {
	if node == nil {
		return ""
	}
	if node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text
	}
	return utils.TrimmedNodeText(sf, node)
}

// isReferenceIdentifier reports whether the Identifier `id` is being used
// as a value reference (as opposed to a declaration name, property key, JSX
// attribute name, etc). Used by the useEffectEvent tracker to ignore the
// declaration site and member-name positions.
func isReferenceIdentifier(id *ast.Node) bool {
	p := id.Parent
	if p == nil {
		return true
	}
	switch p.Kind {
	case ast.KindPropertyAccessExpression:
		// In `obj.prop`, only `obj` is a reference; `prop` is a name.
		return p.AsPropertyAccessExpression().Expression == id
	case ast.KindPropertyAssignment:
		// `{ key: value }` — key is not a reference.
		pa := p.AsPropertyAssignment()
		return pa.Initializer == id
	case ast.KindShorthandPropertyAssignment:
		// `{ x }` — the identifier IS a reference (it stands for both name and value).
		return p.Name() == id
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
		return p.Name() != id
	case ast.KindJsxAttribute:
		// Attribute name (`<Foo bar={...} />`) is not a reference.
		return false
	case ast.KindVariableDeclaration:
		return p.AsVariableDeclaration().Name() != id
	case ast.KindBindingElement:
		return p.AsBindingElement().Name() != id
	case ast.KindParameter:
		return p.AsParameterDeclaration().Name() != id
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindClassDeclaration, ast.KindClassExpression:
		return p.Name() != id
	case ast.KindLabeledStatement:
		return p.AsLabeledStatement().Label != id
	case ast.KindBreakStatement, ast.KindContinueStatement:
		// Label position; not a reference.
		return false
	case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport,
		ast.KindExportSpecifier:
		// Import / export declarations — not value references.
		return false
	}
	return true
}

// isInsideEffectArgument walks up from `idNode` through CallExpression
// ancestors. Returns true once we find a CallExpression whose callee is an
// effect or useEffectEvent hook AND `idNode` (or one of its ancestors on the
// path) is inside the call's arguments (not the callee).
//
// Used to gate useEffectEvent reference reports: a reference to an
// `useEffectEvent` binding is allowed inside `useEffect(...)` /
// `useLayoutEffect(...)` / `useInsertionEffect(...)` / configured effect
// hooks / nested `useEffectEvent(...)` callbacks.
func isInsideEffectArgument(idNode *ast.Node, additional *regexp.Regexp) bool {
	cur := idNode
	for cur != nil && cur.Parent != nil {
		p := cur.Parent
		if p.Kind == ast.KindCallExpression {
			call := p.AsCallExpression()
			// `cur` must be inside the arguments, not the callee.
			if call.Expression != cur {
				if isEffectCalleeName(call.Expression, additional) || isUseEffectEventCallee(call.Expression) {
					return true
				}
			}
		}
		cur = p
	}
	return false
}

// getAdditionalEffectHooks delegates to the shared settings reader.
func getAdditionalEffectHooks(settings map[string]interface{}) *regexp.Regexp {
	return react_hooksutil.AdditionalHooksFromSettings(settings, "additionalEffectHooks")
}

// parseAdditionalHooks resolves the additional-effect-hooks regex: the
// rule-level `additionalHooks` option takes precedence, with
// `settings['react-hooks'].additionalEffectHooks` as the fallback —
// mirroring upstream's create() since facebook/react#37085.
func parseAdditionalHooks(options []any, settings map[string]interface{}) *regexp.Regexp {
	re := getAdditionalEffectHooks(settings)
	if len(options) == 0 {
		return re
	}
	optsMap, _ := options[0].(map[string]interface{})
	// Mirrors upstream's `rawOptions.additionalHooks` truthiness check: a
	// non-empty rule-level pattern replaces the settings fallback even when
	// it fails to compile; absent or empty keeps the settings-derived value.
	if raw, _ := optsMap["additionalHooks"].(string); raw != "" {
		re = nil
		if compiled, err := regexp.Compile(raw); err == nil {
			re = compiled
		}
	}
	return re
}

func sourceMayUseHooks(sourceFile *ast.SourceFile) bool {
	// Parsed identifier names are normalized, including Unicode escapes, and
	// include property names such as the useState in React.useState. The table
	// may also intern literal text, making computed properties a harmless false
	// positive. Stay conservative for direct callers without parser metadata.
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	for name := range sourceFile.Identifiers {
		if hasHookNameShape(name) {
			return true
		}
	}
	return false
}

func hasHookNameShape(name string) bool {
	if name == "use" {
		return true
	}
	if len(name) < 4 || name[0] != 'u' || name[1] != 's' || name[2] != 'e' {
		return false
	}
	next := name[3]
	return next >= 'A' && next <= 'Z' || next >= '0' && next <= '9'
}

func mayBeHookCallee(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return hasHookNameShape(node.AsIdentifier().Text)
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		return name != nil && name.Kind == ast.KindIdentifier && hasHookNameShape(name.AsIdentifier().Text)
	}
	return false
}

func sourceMayUseEffectEvent(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	_, ok := sourceFile.Identifiers["useEffectEvent"]
	return ok
}

// eeRegistry tracks `const X = useEffectEvent(...)` declarations per enclosing
// function-like. Its name set filters the Identifier listener before reference
// resolution; its binding list supports parser-only callers without Refs or a
// TypeChecker.
type eeRegistry struct {
	bindings []eeBinding
	names    map[string]struct{}
}

type eeBinding struct {
	container *ast.Node
	name      string
}

func (r *eeRegistry) record(container *ast.Node, name string) {
	if container == nil || name == "" {
		return
	}
	if r.names == nil {
		r.names = make(map[string]struct{})
	}
	r.bindings = append(r.bindings, eeBinding{container: container, name: name})
	r.names[name] = struct{}{}
}

func (r *eeRegistry) hasName(name string) bool {
	if r == nil {
		return false
	}
	_, ok := r.names[name]
	return ok
}

// resolveContainer returns the closest container with a tracked binding for
// `name` walking up via `findEnclosingFunction`. It is only used when neither
// RefStore nor the TypeChecker can resolve an identifier.
func (r *eeRegistry) resolveContainer(idNode *ast.Node, name string) *ast.Node {
	if r == nil {
		return nil
	}
	cur := findEnclosingFunction(idNode)
	for cur != nil {
		for _, binding := range r.bindings {
			if binding.container == cur && binding.name == name {
				return cur
			}
		}
		cur = findEnclosingFunction(cur)
	}
	return nil
}

// collectEffectEventBindings records every `useEffectEvent(...)` variable
// binding in any enclosing function-like (mirrors upstream's
// `recordAllUseEffectEventFunctions` which fires for every
// FunctionDeclaration / ArrowFunctionExpression). Bound source files use their
// locals-container chain; parser-only callers fall back to an AST walk.
func collectEffectEventBindings(sf *ast.SourceFile) *eeRegistry {
	if sf.IsBound() {
		reg := &eeRegistry{}
		complete := true
		for container := sf.AsNode(); container != nil; {
			data := container.LocalsContainerData()
			if data == nil {
				complete = false
				break
			}
			for _, symbol := range data.Locals {
				for _, declaration := range symbol.Declarations {
					recordEffectEventBinding(reg, declaration)
				}
			}
			container = data.NextContainer
		}
		if complete {
			if len(reg.bindings) == 0 {
				return nil
			}
			return reg
		}
	}

	// Direct parser-only callers do not have binder symbol tables. Preserve a
	// syntax-only fallback for them; normal linter execution takes the cheaper
	// locals-container path above.
	reg := &eeRegistry{}
	var visit func(n *ast.Node) bool
	visit = func(n *ast.Node) bool {
		recordEffectEventBinding(reg, n)
		n.ForEachChild(visit)
		return false
	}
	sf.AsNode().ForEachChild(visit)
	if len(reg.bindings) == 0 {
		return nil
	}
	return reg
}

func recordEffectEventBinding(reg *eeRegistry, node *ast.Node) {
	if !isEffectEventVariableDeclaration(node) && !isEffectEventBindingElement(node) {
		return
	}
	name := node.Name()
	if name != nil && name.Kind == ast.KindIdentifier {
		reg.record(findEnclosingFunction(node), name.AsIdentifier().Text)
	}
}

// isEffectEventVariableDeclaration reports whether `n` is a VariableDeclaration
// whose initializer is a (paren-wrapped) `useEffectEvent(...)` call.
func isEffectEventVariableDeclaration(n *ast.Node) bool {
	if n.Kind != ast.KindVariableDeclaration {
		return false
	}
	vd := n.AsVariableDeclaration()
	return vd.Initializer != nil && isEventInitializer(vd.Initializer)
}

// isEffectEventBindingElement reports whether `n` is a BindingElement whose
// (default) initializer is a `useEffectEvent(...)` call. Covers the
// `const { onClick = useEffectEvent(...) } = obj` shape.
func isEffectEventBindingElement(n *ast.Node) bool {
	if n.Kind != ast.KindBindingElement {
		return false
	}
	be := n.AsBindingElement()
	return be.Initializer != nil && isEventInitializer(be.Initializer)
}

// isEventInitializer reports whether `expr` is a `useEffectEvent(...)` call.
// Strips paren wrappers transparently so `(useEffectEvent(...))` is recognized.
func isEventInitializer(expr *ast.Node) bool {
	e := ast.SkipParentheses(expr)
	if e == nil || e.Kind != ast.KindCallExpression {
		return false
	}
	return isUseEffectEventCallee(e.AsCallExpression().Expression)
}

// isUseEffectEventBindingDeclaration reports whether `decl` is a declaration
// node whose initializer is `useEffectEvent(...)`.
func isUseEffectEventBindingDeclaration(decl *ast.Node) bool {
	if decl == nil {
		return false
	}
	switch decl.Kind {
	case ast.KindVariableDeclaration:
		return decl.AsVariableDeclaration().Initializer != nil &&
			isEventInitializer(decl.AsVariableDeclaration().Initializer)
	case ast.KindBindingElement:
		return decl.AsBindingElement().Initializer != nil &&
			isEventInitializer(decl.AsBindingElement().Initializer)
	}
	return false
}

// resolveBinding uses the shared RefStore first and the TypeChecker only as a
// compatibility fallback. The boolean distinguishes an unresolved identifier
// from one conclusively resolved to a non-Effect-Event shadowing declaration.
func resolveBinding(ctx rule.RuleContext, idNode *ast.Node) (*ast.Node, bool) {
	if idNode == nil {
		return nil, false
	}
	var sym *ast.Symbol
	if ctx.Refs != nil {
		sym = ctx.Refs.Resolve(idNode)
	}
	if sym == nil && ctx.TypeChecker != nil {
		sym = ctx.TypeChecker.GetSymbolAtLocation(idNode)
	}
	if sym == nil {
		return nil, false
	}
	for _, decl := range sym.Declarations {
		if isUseEffectEventBindingDeclaration(decl) {
			return findEnclosingFunction(decl), true
		}
	}
	// A resolved non-Effect-Event symbol proves this occurrence is shadowed.
	// Only unresolved identifiers should use the name-based fallback.
	return nil, true
}

// hasFlowSuppression reports whether the line immediately preceding
// `hookNode` contains a `$FlowFixMe[react-rule-hook]` comment. Mirrors
// upstream's same-named helper observably: the comment must match the
// regex AND its end-line must equal `hookNode.start.line - 1`.
//
// Implementation walks leading comments at `hookNode.Pos()` (tsgo attaches
// preceding trivia to the next-token Pos), as well as the leading comments
// at the hook's enclosing statement (covers the case where the comment is
// attached to the statement-level Pos rather than the inner CallExpression).
func hasFlowSuppression(sf *ast.SourceFile, hookNode *ast.Node) bool {
	if hookNode == nil || sf == nil {
		return false
	}
	// Use the trimmed start (skip leading trivia) so the hook's "start line"
	// is the line of the first significant character of the hook itself,
	// matching upstream's `node.loc.start.line` semantics. Raw `node.Pos()`
	// in tsgo includes leading trivia (whitespace + comments), which would
	// place us on the line of the trivia rather than the token.
	trimmed := utils.TrimNodeTextRange(sf, hookNode)
	hookLine, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, trimmed.Pos())
	text := sf.Text()
	nodeFactory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	check := func(pos int) bool {
		for c := range scanner.GetLeadingCommentRanges(nodeFactory, text, pos) {
			if matchesFlowSuppression(text, c, hookLine, sf) {
				return true
			}
		}
		for c := range scanner.GetTrailingCommentRanges(nodeFactory, text, pos) {
			if matchesFlowSuppression(text, c, hookLine, sf) {
				return true
			}
		}
		return false
	}
	if check(hookNode.Pos()) {
		return true
	}
	// Walk to the enclosing statement (ExpressionStatement, VariableStatement,
	// etc.) and check its leading trivia. The suppression comment is often
	// attached to the statement, not the inner expression.
	cur := hookNode.Parent
	for cur != nil {
		switch cur.Kind {
		case ast.KindExpressionStatement, ast.KindVariableStatement, ast.KindReturnStatement,
			ast.KindIfStatement, ast.KindBlock, ast.KindForStatement, ast.KindWhileStatement,
			ast.KindDoStatement, ast.KindThrowStatement:
			if check(cur.Pos()) {
				return true
			}
		}
		if isFunctionLikeContainer(cur) {
			break
		}
		cur = cur.Parent
	}
	return false
}

// matchesFlowSuppression reports whether `c` carries the suppression marker
// AND its end-line equals `hookLine - 1` (i.e., it sits on the immediately
// preceding line).
func matchesFlowSuppression(text string, c ast.CommentRange, hookLine int, sf *ast.SourceFile) bool {
	commentText := text[c.Pos():c.End()]
	if !flowSuppressionRegex.MatchString(commentText) {
		return false
	}
	endLine, _ := scanner.GetECMALineAndUTF16CharacterOfPosition(sf, c.End())
	return endLine == hookLine-1
}

// makeHookText returns the rendered display text for a hook callee, used in
// every diagnostic message. Identifier hooks render as their bare name;
// member-access hooks render as `Foo.useBar` via the source text of the
// PropertyAccessExpression. Paren wrappers around the callee are peeled
// (mirrors ESTree, which never exposes ParenthesizedExpression at all).
func makeHookText(hookCall *ast.Node, sf *ast.SourceFile) string {
	if hookCall == nil {
		return ""
	}
	n := ast.SkipParentheses(hookCall)
	if n.Kind == ast.KindIdentifier {
		return n.AsIdentifier().Text
	}
	return utils.TrimmedNodeText(sf, n)
}

// useEffectEventErrorMessage formats the diagnostic for a useEffectEvent
// reference outside of the allowed contexts. Mirrors upstream's
// `useEffectEventError(fn, called)` byte-for-byte.
func useEffectEventErrorMessage(fnName string, called bool) string {
	if fnName == "" {
		return `React Hook "useEffectEvent" can only be called at the top level of your component. It cannot be passed down.`
	}
	suffix := ""
	if !called {
		suffix = " It cannot be assigned to a variable or passed down."
	}
	return "`" + fnName + "` is a function created with React Hook \"useEffectEvent\", and can only be called from Effects and Effect Events in the same component." + suffix
}

func reportRange(sourceText string, node *ast.Node) core.TextRange {
	// RuleContext.ReportNode tokenizes from node.Pos() to trim leading trivia.
	// SkipTrivia computes the same start without allocating a scanner, while
	// node.End() remains the report end for every node shape used by this rule.
	return node.Loc.WithPos(scanner.SkipTrivia(sourceText, node.Pos()))
}

// RulesOfHooksRule is the rslint port of upstream `react-hooks/rules-of-hooks`.
//
// Departure from upstream: rslint has no CodePathAnalyzer, so the conditional /
// loop / early-return / cycled-segment detections are AST-shape based rather
// than CFG-based. The implementation matches upstream observably for every
// test in upstream's `valid` and `invalid` suites that doesn't rely on
// BigInt-precise path counting (and we add a `hasLabeledBreakBefore` walk
// for the `label: { if (cond) break label; useFoo(); }` case).
//
// Identifier resolution for `useEffectEvent` references uses the shared Refs
// index first (correct under shadowing), then the TypeChecker compatibility
// path, and finally the per-container registry for parser-only callers.
//
// `$FlowFixMe[react-rule-hook]` suppression on the preceding line is
// honored, mirroring upstream's `hasFlowSuppression` byte-for-byte.
var RulesOfHooksRule = rule.Rule{
	Name:   "react-hooks/rules-of-hooks",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sf := ctx.SourceFile
		if !sourceMayUseHooks(sf) {
			return nil
		}
		additionalRe := parseAdditionalHooks(options, ctx.Settings)
		var registry *eeRegistry
		if sourceMayUseEffectEvent(sf) {
			registry = collectEffectEventBindings(sf)
		}
		sourceText := sf.Text()
		flowSuppressionChecked := false
		mayHaveFlowSuppression := false
		isFlowSuppressed := func(node *ast.Node) bool {
			if !flowSuppressionChecked {
				flowSuppressionChecked = true
				mayHaveFlowSuppression = strings.Contains(sourceText, "$FlowFixMe[react-rule-hook]")
			}
			return mayHaveFlowSuppression && hasFlowSuppression(sf, node)
		}

		report := func(node *ast.Node, msg string) {
			if isFlowSuppressed(node) {
				return
			}
			ctx.ReportRange(reportRange(sourceText, node), rule.RuleMessage{Description: msg})
		}

		listeners := rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				callee := call.Expression
				if !mayBeHookCallee(callee) || !isHookCallee(callee) {
					return
				}

				// Inline-call useEffectEvent that isn't being assigned.
				// Mirrors upstream's standalone check: useEffectEvent must
				// either bind to a variable or be a top-level expression
				// statement. Anything else (`<Child onClick={useEffectEvent(...)} />`)
				// is reported as the "passed down" variant.
				if isUseEffectEventCallee(callee) {
					p := node.Parent
					if p != nil && p.Kind != ast.KindVariableDeclaration && p.Kind != ast.KindExpressionStatement {
						report(node, useEffectEventErrorMessage("", false))
					}
				}

				fn := findEnclosingFunction(node)
				isUse := isUseIdentifier(callee)

				// Skip unreachable hooks (preceded by a direct `return` or
				// `throw` in the same block). Upstream's CFG marks the
				// segment as `!reachable` and the rule loop `continue`s,
				// emitting no diagnostic of any kind.
				if fn != nil && isPrecededByDirectTerminator(node, fn) {
					return
				}

				// Try/catch + use(): hard error, distinct from any other.
				if isUse && isInsideTryCatchOfFunction(node, fn) {
					report(callee, fmt.Sprintf(`React Hook "%s" cannot be called in a try/catch block.`, makeHookText(callee, sf)))
				}

				// Loop check (cycled || do-while). Skipped for use(), which
				// upstream allows in loops.
				inAnyLoop := false
				if !isUse {
					inAnyLoop = isInsideLoopOfFunction(node, fn)
					if inAnyLoop {
						report(callee, fmt.Sprintf(
							`React Hook "%s" may be executed more than once. Possibly because it is called in a loop. React Hooks must be called in the exact same order in every component render.`,
							makeHookText(callee, sf),
						))
					}
				}

				// Container classification.
				if fn == nil {
					report(callee, fmt.Sprintf(
						`React Hook "%s" cannot be called at the top level. React Hooks must be called in a React function component or a custom React Hook function.`,
						makeHookText(callee, sf),
					))
					return
				}

				if isComponentOrHookFn(fn) {
					if hasAsyncModifier(fn) {
						report(callee, fmt.Sprintf(`React Hook "%s" cannot be called in an async function.`, makeHookText(callee, sf)))
						return
					}
					if isUse || inAnyLoop {
						return
					}
					cond := isConditional(node, fn)
					early := hasEarlyReturnBefore(node, fn)
					labeled := hasLabeledBreakBefore(node, fn)
					tryWithPrior := isInsideTryBlockWithPriorStmt(node, fn)
					if cond || early || labeled || tryWithPrior {
						suffix := ""
						if early {
							suffix = " Did you accidentally call a React Hook after an early return?"
						}
						report(callee, fmt.Sprintf(
							`React Hook "%s" is called conditionally. React Hooks must be called in the exact same order in every component render.%s`,
							makeHookText(callee, sf), suffix,
						))
					}
					return
				}

				// Class member.
				if isClassMember(fn) {
					report(callee, fmt.Sprintf(
						`React Hook "%s" cannot be called in a class component. React Hooks must be called in a React function component or a custom React Hook function.`,
						makeHookText(callee, sf),
					))
					return
				}

				// Named function (non-component, non-hook).
				if name := getFunctionName(fn); name != nil {
					report(callee, fmt.Sprintf(
						`React Hook "%s" is called in function "%s" that is neither a React function component nor a custom React Hook function. React component names must start with an uppercase letter. React Hook names must start with the word "use".`,
						makeHookText(callee, sf), nameText(name, sf),
					))
					return
				}

				// Anonymous callback inside (somewhere) a component or hook.
				if !isUse && isInsideComponentOrHook(node) {
					report(callee, fmt.Sprintf(
						`React Hook "%s" cannot be called inside a callback. React Hooks must be called in a React function component or a custom React Hook function.`,
						makeHookText(callee, sf),
					))
				}
			},
		}

		if registry != nil {
			listeners[ast.KindIdentifier] = func(node *ast.Node) {
				name := node.AsIdentifier().Text
				if !registry.hasName(name) || !isReferenceIdentifier(node) {
					return
				}

				// RefStore resolves the common local-binding path without a
				// TypeChecker round trip. A name-based fallback remains for
				// direct callers that provide neither facility.
				container, resolved := resolveBinding(ctx, node)
				if !resolved {
					container = registry.resolveContainer(node, name)
				}
				if container == nil {
					return
				}
				if isInsideEffectArgument(node, additionalRe) {
					return
				}
				called := false
				if p := node.Parent; p != nil && p.Kind == ast.KindCallExpression {
					if p.AsCallExpression().Expression == node {
						called = true
					}
				}
				if isFlowSuppressed(node) {
					return
				}
				ctx.ReportRange(reportRange(sourceText, node), rule.RuleMessage{Description: useEffectEventErrorMessage(name, called)})
			}
		}
		return listeners
	},
}
