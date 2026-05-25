package require_optimization

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Options mirrors upstream's schema:
//
//	[{
//	  type: 'object',
//	  properties: {
//	    allowDecorators: {
//	      type: 'array',
//	      items: { type: 'string' },
//	    },
//	  },
//	  additionalProperties: false,
//	}]
type Options struct {
	AllowDecorators []string
}

func parseOptions(options any) Options {
	opts := Options{}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	if v, ok := optsMap["allowDecorators"].([]interface{}); ok {
		for _, name := range v {
			if s, ok := name.(string); ok {
				opts.AllowDecorators = append(opts.AllowDecorators, s)
			}
		}
	}
	return opts
}

// hasPureRenderDecorator mirrors upstream's `hasPureRenderDecorator`: the
// class carries a `@reactMixin.decorate(PureRenderMixin)` decorator. The
// decorator expression must be a non-optional-chain CallExpression whose
// callee is `reactMixin.decorate` and whose first argument is the Identifier
// `PureRenderMixin`. Anything else (optional chain on the call or its
// receiver, wrong member, missing args) doesn't match — same as upstream's
// chained property accesses silently bailing on undefined.
//
// Parens are transparent on the call expression and its receiver (ESTree
// flattens; tsgo preserves) so `(reactMixin.decorate)(...)` and
// `(reactMixin).decorate(...)` still match.
func hasPureRenderDecorator(classNode *ast.Node) bool {
	mods := classNode.Modifiers()
	if mods == nil {
		return false
	}
	for _, mod := range mods.Nodes {
		if mod.Kind != ast.KindDecorator {
			continue
		}
		expr := ast.SkipParentheses(mod.AsDecorator().Expression)
		if expr == nil || expr.Kind != ast.KindCallExpression {
			continue
		}
		if ast.IsOptionalChain(expr) {
			continue
		}
		call := expr.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
			continue
		}
		if ast.IsOptionalChain(callee) {
			continue
		}
		pa := callee.AsPropertyAccessExpression()
		obj := ast.SkipParentheses(pa.Expression)
		if obj == nil || obj.Kind != ast.KindIdentifier {
			continue
		}
		if obj.AsIdentifier().Text != "reactMixin" {
			continue
		}
		propName := pa.Name()
		if propName == nil || propName.Kind != ast.KindIdentifier {
			continue
		}
		if propName.AsIdentifier().Text != "decorate" {
			continue
		}
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			continue
		}
		firstArg := ast.SkipParentheses(call.Arguments.Nodes[0])
		if firstArg == nil || firstArg.Kind != ast.KindIdentifier {
			continue
		}
		if firstArg.AsIdentifier().Text == "PureRenderMixin" {
			return true
		}
	}
	return false
}

// hasCustomDecorator mirrors upstream's `hasCustomDecorator`: any of the
// class's decorators is a bare Identifier whose name appears in
// `allowDecorators`. Decorators of any other shape (CallExpression,
// PropertyAccess, etc.) do not match — matching upstream's `expression.name`
// lookup, which is undefined on non-Identifier expressions.
func hasCustomDecorator(classNode *ast.Node, allowDecorators []string) bool {
	if len(allowDecorators) == 0 {
		return false
	}
	mods := classNode.Modifiers()
	if mods == nil {
		return false
	}
	for _, mod := range mods.Nodes {
		if mod.Kind != ast.KindDecorator {
			continue
		}
		expr := ast.SkipParentheses(mod.AsDecorator().Expression)
		if expr == nil || expr.Kind != ast.KindIdentifier {
			continue
		}
		name := expr.AsIdentifier().Text
		for _, allow := range allowDecorators {
			if name == allow {
				return true
			}
		}
	}
	return false
}

// isPureRenderMixinsProperty mirrors upstream's `isPureRenderDeclared`: the
// property is a `mixins: [..., PureRenderMixin, ...]` PropertyAssignment whose
// initializer is an ArrayLiteralExpression containing at least one Identifier
// element named `PureRenderMixin`. Other property forms (shorthand,
// MethodDeclaration, accessors, spread) cannot satisfy this shape —
// MethodDeclaration's body isn't an array, ShorthandPropertyAssignment has no
// initializer, SpreadAssignment has no name. Upstream's `node.value.elements`
// lookup naturally short-circuits the same way (undefined on
// non-ArrayExpression values).
func isPureRenderMixinsProperty(prop *ast.Node) bool {
	if prop == nil || prop.Kind != ast.KindPropertyAssignment {
		return false
	}
	pa := prop.AsPropertyAssignment()
	if reactutil.IdentifierOrPrivateName(pa.Name()) != "mixins" {
		return false
	}
	init := pa.Initializer
	if init == nil || init.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	for _, el := range init.AsArrayLiteralExpression().Elements.Nodes {
		el = ast.SkipParentheses(el)
		if el == nil {
			continue
		}
		if el.Kind == ast.KindIdentifier && el.AsIdentifier().Text == "PureRenderMixin" {
			return true
		}
	}
	return false
}

// objectDeclaresSCUOrMixin mirrors upstream's ObjectExpression listener gate:
// any property whose key (Identifier / PrivateIdentifier) is
// `shouldComponentUpdate`, OR a `mixins: [..., PureRenderMixin, ...]`
// PropertyAssignment, marks the object as having SCU.
func objectDeclaresSCUOrMixin(obj *ast.Node) bool {
	for _, prop := range obj.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		if reactutil.IdentifierOrPrivateName(prop.Name()) == "shouldComponentUpdate" {
			return true
		}
		if isPureRenderMixinsProperty(prop) {
			return true
		}
	}
	return false
}

// isInClassDeclarationMethodBody mirrors upstream's `isFunctionInClass`:
// walks the parent chain looking for a ClassDeclaration scope. The function
// must be reached via a body that ESLint's scope manager opens as a
// function-like / class-static-block scope, anchored within a
// ClassDeclaration (not a ClassExpression — upstream's `isFunctionInClass`
// only matches `ClassDeclaration`).
//
// Note: class-field initializer arrows (PropertyDeclaration → ArrowFunction)
// are NOT pre-rejected here. Their exclusion from SFC classification is
// handled by `reactutil.IsStatelessReactComponent`'s
// `isInAllowedPositionForComponent` gate (PropertyDeclaration is not in
// the allowed-parent set), which keeps the boundary in one place. An
// anonymous arrow nested INSIDE a class-field initializer (e.g.
// `build = () => () => <div/>`) does live in a ClassDeclaration scope
// upstream — its scope walk goes inner-arrow → outer-arrow → class — and
// must be allowed to reach the ClassDeclaration here.
//
// Verified empirically against eslint-plugin-react:
//
//	class C extends React.Component {
//	  build() { function Inner() { return <div />; } }   // Inner reported
//	  build = () => () => <div/>;                         // inner arrow reported
//	  build = () => <div />;                              // arrow not reported (small `build`)
//	  Inner = (p) => <div />;                             // arrow not reported (PropertyDeclaration parent reject)
//	  static { function Inner() { return <div />; } }    // Inner reported
//	}
//	const X = class extends React.Component {
//	  build() { function Inner() { return <div />; } }   // Inner NOT reported
//	};
func isInClassDeclarationMethodBody(fn *ast.Node) bool {
	sawMethodBody := false
	for p := fn.Parent; p != nil; p = p.Parent {
		switch p.Kind {
		case ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor,
			ast.KindClassStaticBlockDeclaration:
			sawMethodBody = true
		case ast.KindPropertyDeclaration:
			// A class-field initializer is also a scope under ESLint's
			// scope manager whose enclosing scope is the class. Treat
			// it as a method-like boundary so nested functions inside
			// it can still reach the ClassDeclaration.
			sawMethodBody = true
		case ast.KindClassDeclaration:
			return sawMethodBody
		case ast.KindClassExpression:
			// Enclosing class is not a ClassDeclaration —
			// `isFunctionInClass` only matches `ClassDeclaration`.
			return false
		}
	}
	return false
}

// methodNameIsSCU returns true for a class member whose Kind is one of the
// MethodDefinition equivalents (regular method, get/set accessor, constructor)
// AND whose key resolves to "shouldComponentUpdate".
//
// Abstract methods are excluded: TSESTree maps `abstract foo()` to
// `TSAbstractMethodDefinition`, NOT `MethodDefinition`, so upstream's
// `MethodDefinition(node)` listener never fires on them. We mirror that by
// rejecting members carrying the `abstract` modifier.
func methodNameIsSCU(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor:
	default:
		return false
	}
	if mods := node.Modifiers(); mods != nil {
		for _, m := range mods.Nodes {
			if m.Kind == ast.KindAbstractKeyword {
				return false
			}
		}
	}
	return reactutil.IdentifierOrPrivateName(node.Name()) == "shouldComponentUpdate"
}

var RequireOptimizationRule = rule.Rule{
	Name: "react/require-optimization",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := parseOptions(options)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)

		msg := rule.RuleMessage{
			Id:          "noShouldComponentUpdate",
			Description: "Component is not optimized. Please add a shouldComponentUpdate method.",
		}

		// isDetectedComponent reports whether `node` is a node that upstream's
		// `Components.detect` would register and that this rule would report
		// when missing SCU. Three categories:
		//
		//  - ClassDeclaration / ClassExpression extending React.Component
		//  - ObjectLiteralExpression that's the arg of a createReactClass call
		//  - Stateless functional component inside a ClassDeclaration's
		//    method body. Top-level / module-level SFCs are NOT detected
		//    here because upstream's `mark-SCU-as-declared(self)` always
		//    silences them (`isFunctionInClass` returns false), so they'd
		//    be a no-op. SFCs inside a ClassExpression's methods are also
		//    excluded — upstream's `isFunctionInClass` only matches
		//    `ClassDeclaration`.
		//
		// SFC classification is delegated to
		// `reactutil.IsStatelessReactComponentWithChecker` which mirrors
		// upstream's full `Components.detect` SFC heuristic (binding-name
		// resolution across VariableDeclaration / PropertyAssignment /
		// AssignmentExpression / wrapper calls, returns-JSX-via-binding
		// resolution, anonymous-arrow handling, …). The TypeChecker-aware
		// variant resolves Identifier returns through symbol→declaration;
		// it falls back to a local-block scan when `ctx.TypeChecker` is
		// nil.
		//
		// Results are memoized per node: the same node is queried during
		// the top-level `collect` walk AND during every enclosing
		// component's `markedFromSubtree` skip-check, and SFC detection
		// internally walks the function body looking for JSX returns —
		// without memoization the rule degrades to O(N²) on large files.
		detectionCache := make(map[*ast.Node]bool)
		isDetectedComponent := func(node *ast.Node) bool {
			if cached, ok := detectionCache[node]; ok {
				return cached
			}
			var result bool
			switch node.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				result = reactutil.ExtendsReactComponent(node, pragma)
			case ast.KindObjectLiteralExpression:
				result = reactutil.IsCreateReactClassObjectArg(node, pragma, createClass)
			case ast.KindFunctionDeclaration,
				ast.KindFunctionExpression,
				ast.KindArrowFunction:
				if isInClassDeclarationMethodBody(node) {
					result = reactutil.IsStatelessReactComponentWithChecker(node, pragma, ctx.TypeChecker)
				}
			}
			detectionCache[node] = result
			return result
		}

		// classSelfMarked covers upstream's `ClassDeclaration(node)` listener
		// — it fires only on ClassDeclaration (NOT ClassExpression) and
		// `mark-SCU-as-declared(node)` resolves to the class itself when the
		// class is a detected component. So `extends PureComponent` /
		// PureRender / custom-allow decorators silence the rule for
		// ClassDeclaration only.
		classSelfMarked := func(c *ast.Node) bool {
			if c.Kind != ast.KindClassDeclaration {
				return false
			}
			return hasPureRenderDecorator(c) ||
				hasCustomDecorator(c, opts.AllowDecorators) ||
				reactutil.ExtendsReactPureComponent(c, pragma)
		}

		// markedFromSubtree implements upstream's `mark-SCU-as-declared` walk-up
		// resolution: any markSCU source node (ObjectExpression with
		// SCU/PureRenderMixin, MethodDefinition with name SCU, ClassDeclaration
		// with PureRender / custom-allow decorator) inside `self`'s subtree
		// flows up the parent chain until it hits the nearest detected
		// component. We mirror that with a pre-order subtree scan that
		// PRUNES nested detected components — they would absorb the
		// markSCU themselves before it reaches `self`.
		//
		// Verified empirically against eslint-plugin-react@latest:
		//
		//   class C extends React.Component {
		//     init() {
		//       const cfg = { shouldComponentUpdate() {} };  // marks C
		//     }
		//   }                                                  // not reported
		//
		//   class Outer extends React.Component {
		//     build() {
		//       class Inner extends React.Component { scu() {} }  // marks Inner
		//     }
		//   }                                                       // Outer reported
		var markedFromSubtree func(self *ast.Node, current *ast.Node) bool
		markedFromSubtree = func(self *ast.Node, current *ast.Node) bool {
			// Skip the subtree of any OTHER detected component — its own
			// markSCU events resolve to itself, not bubbling up to `self`.
			if current != self && isDetectedComponent(current) {
				return false
			}

			// markSCU source: any ObjectExpression with SCU/PureRenderMixin.
			if current.Kind == ast.KindObjectLiteralExpression && objectDeclaresSCUOrMixin(current) {
				return true
			}
			// markSCU source: any class member (MethodDef/Get/Set/Constructor)
			// named shouldComponentUpdate.
			if methodNameIsSCU(current) {
				return true
			}
			// markSCU source: nested ClassDeclaration carrying PureRender or
			// custom-allow decorator. Upstream's listener fires only on
			// ClassDeclaration, so a nested ClassExpression with the same
			// shape is NOT a markSCU source. Nested ClassDeclaration that
			// itself extends PureComponent is already handled by the
			// "skip detected component" guard above (it's a detected
			// component, absorbs its own markSCU).
			if current != self && current.Kind == ast.KindClassDeclaration {
				if hasPureRenderDecorator(current) ||
					hasCustomDecorator(current, opts.AllowDecorators) {
					return true
				}
			}

			found := false
			current.ForEachChild(func(child *ast.Node) bool {
				if markedFromSubtree(self, child) {
					found = true
					return true // early exit
				}
				return false
			})
			return found
		}

		report := func(c *ast.Node) {
			switch c.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				// Class report anchors at `class` keyword (or
				// abstract/decorator/declare-equivalent), trimming
				// `export` / `default` modifiers.
				start := reactutil.ClassKeywordStart(ctx.SourceFile.Text(), c)
				ctx.ReportRange(core.NewTextRange(start, c.End()), msg)
			default:
				// ObjectLiteralExpression / FunctionDeclaration /
				// FunctionExpression / ArrowFunction — anchor at
				// the node's start (function keyword, `{`, or
				// `(` for parenless arrow).
				ctx.ReportNode(c, msg)
			}
		}

		// Single-pass collect-and-resolve. Done eagerly here (instead of via
		// listeners + a SourceFile-exit hook) because rslint's visitor enters
		// the SourceFile's children directly — the SourceFile node itself is
		// never visited, so a `ListenerOnExit(KindSourceFile)` registration
		// would silently never fire.
		var collect func(*ast.Node)
		collect = func(n *ast.Node) {
			if n == nil {
				return
			}
			if isDetectedComponent(n) {
				if !classSelfMarked(n) && !markedFromSubtree(n, n) {
					report(n)
				}
				// Continue walking — a detected component's subtree may
				// contain further detected components (nested class /
				// nested createReactClass) that need their own
				// independent evaluation.
			}
			n.ForEachChild(func(c *ast.Node) bool {
				collect(c)
				return false
			})
		}
		ctx.SourceFile.Node.ForEachChild(func(n *ast.Node) bool {
			collect(n)
			return false
		})

		return rule.RuleListeners{}
	},
}
