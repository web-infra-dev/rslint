package no_unsafe

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// unsafeMeta mirrors upstream's per-method record: the new method to suggest
// in the message body, plus a `details` tail (always the React async-rendering
// blog link in upstream).
type unsafeMeta struct {
	newMethod string
	details   string
}

const asyncRenderingDoc = "See https://reactjs.org/blog/2018/03/27/update-on-async-rendering.html."

// buildUnsafeMap returns the table of unsafe lifecycle method names. The three
// `UNSAFE_`-prefixed names are always present; the unprefixed aliases are
// added only when `checkAliases` is true. Mirrors upstream's `unsafe` object
// constructed inside `create(context)`.
func buildUnsafeMap(checkAliases bool) map[string]unsafeMeta {
	m := map[string]unsafeMeta{
		"UNSAFE_componentWillMount":        {newMethod: "componentDidMount", details: asyncRenderingDoc},
		"UNSAFE_componentWillReceiveProps": {newMethod: "getDerivedStateFromProps", details: asyncRenderingDoc},
		"UNSAFE_componentWillUpdate":       {newMethod: "componentDidUpdate", details: asyncRenderingDoc},
	}
	if checkAliases {
		m["componentWillMount"] = m["UNSAFE_componentWillMount"]
		m["componentWillReceiveProps"] = m["UNSAFE_componentWillReceiveProps"]
		m["componentWillUpdate"] = m["UNSAFE_componentWillUpdate"]
	}
	return m
}

// formatMessage builds the diagnostic string. Mirrors upstream's template:
//
//	{{method}} is unsafe for use in async rendering. Update the component to use {{newMethod}} instead. {{details}}
func formatMessage(method string, m unsafeMeta) string {
	return method + " is unsafe for use in async rendering. Update the component to use " + m.newMethod + " instead. " + m.details
}

// memberKeyName extracts the property-name string for a class member or
// object-literal property — the equivalent of upstream's
// `astUtil.getPropertyName(property)`. That helper returns `nameNode.name`,
// which means **only** Identifier and PrivateIdentifier keys yield a string;
// StringLiteral / NumericLiteral / ComputedPropertyName keys all return
// undefined upstream (Literal nodes have `.value`, not `.name`). So a key
// like `'componentWillMount': fn` does NOT flag in ESLint, and we lock that
// in by relying on `reactutil.EsTreeName` (which only resolves Identifier /
// PrivateIdentifier).
//
// We deliberately do NOT use `utils.GetStaticPropertyName` here: that helper
// statically resolves a computed key like `['componentWillMount']` AND
// resolves StringLiteral keys to their value — both broader than upstream
// would match. `EsTreeName` is the precise port of upstream's read.
func memberKeyName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	return reactutil.EsTreeName(node.Name())
}

// parseOptions extracts `checkAliases` from the rule's options. Upstream
// reads `context.options[0] || {}` and falls back to `false` when the key is
// missing or non-boolean. The CLI-vs-test option-shape unification is
// delegated to `utils.GetOptionsMap` (see PORT_RULE.md "Handling Options").
func parseOptions(options any) bool {
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return false
	}
	v, ok := optsMap["checkAliases"].(bool)
	if !ok {
		return false
	}
	return v
}

var NoUnsafeRule = rule.Rule{
	Name: "react/no-unsafe",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// `testReactVersion(context, '>= 16.3.0')` — at React < 16.3 the
		// `UNSAFE_` lifecycle aliases didn't exist, so the rule disables
		// itself entirely. Default version (no `settings.react.version`)
		// resolves to "latest" → the rule is active.
		if reactutil.ReactVersionLessThan(ctx.Settings, 16, 3, 0) {
			return rule.RuleListeners{}
		}

		checkAliases := parseOptions(options)
		unsafeMap := buildUnsafeMap(checkAliases)
		pragma := reactutil.GetReactPragma(ctx.Settings)
		createClass := reactutil.GetReactCreateClass(ctx.Settings)

		// checkLifeCycleMethods iterates a component's members in source
		// order and reports each one whose name is a known unsafe lifecycle.
		//
		// **Why no in-rule sort**: upstream calls
		// `methods.sort((a,b) => a.localeCompare(b)).forEach(checkUnsafe)`,
		// but ESLint's diagnostic output is re-sorted by `(line, column)` at
		// the reporter layer — so the in-rule `sort()` is dead code as far
		// as user-visible output is concerned. Source-order traversal
		// reproduces ESLint's final ordering exactly without paying the
		// cost of a no-op sort or trying to mirror Node.js's locale-
		// dependent `localeCompare`.
		// Mirror upstream's exact reporting semantics for duplicate-named
		// members (TS overloads, repeated keys). upstream:
		//
		//   methods.sort().forEach(method => {
		//     const propertyNode = members.find(p => getPropertyName(p) === method);
		//     report({ node: propertyNode, ... })
		//   })
		//
		// `methods` is built by `properties.map(getPropertyName)` — it
		// preserves duplicates. For each occurrence of `name`, `find`
		// returns the FIRST member with that name. So N overloads of
		// `UNSAFE_componentWillMount` produce N diagnostics, all anchored
		// to the first overload. We replicate this byte-for-byte: walk
		// members in source order, count duplicates per name, then for
		// each name emit `count` diagnostics on the first matching node.
		checkLifeCycleMethods := func(members []*ast.Node) {
			counts := map[string]int{}
			firstNode := map[string]*ast.Node{}
			// First pass: count occurrences and remember first node per name.
			// Order of first appearance is recorded by the slice below so
			// our output order is deterministic and matches a source-order
			// (== ESLint reporter (line,column)) walk.
			var orderedNames []string
			for _, m := range members {
				if m == nil {
					continue
				}
				name := memberKeyName(m)
				if name == "" {
					continue
				}
				if _, ok := unsafeMap[name]; !ok {
					continue
				}
				if _, seen := firstNode[name]; !seen {
					firstNode[name] = m
					orderedNames = append(orderedNames, name)
				}
				counts[name]++
			}
			// Second pass: emit diagnostics in source-order of first
			// appearance (matches ESLint's reporter-layer (line,column)
			// sort, since `firstNode` is the leftmost occurrence per name).
			for _, name := range orderedNames {
				meta := unsafeMap[name]
				node := firstNode[name]
				for range counts[name] {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "unsafeMethod",
						Description: formatMessage(name, meta),
					})
				}
			}
		}

		// isES5Component mirrors upstream's `componentUtil.isES5Component`
		// directly — looser than `reactutil.IsCreateReactClassObjectArg`,
		// which restricts to the FIRST argument. Upstream's check is
		// `node.parent.callee` only, accepting the ObjectExpression at any
		// argument position. Verified empirically against
		// eslint-plugin-react@latest:
		//
		//   - `createReactClass(other, {...})` — flags (any arg position).
		//   - `new createReactClass({...})` — flags (NewExpression also
		//     exposes a `.callee` in ESTree, so upstream's
		//     `parent.callee` accepts it).
		//   - `createReactClass(({...}))` — flags (ESTree flattens parens).
		//
		// tsgo preserves parens that ESTree flattens, and tsgo's
		// `KindNewExpression` is a separate kind from `KindCallExpression`.
		// We walk past wrapping parens and accept either kind to mirror
		// upstream's flattened view.
		isES5Component := func(obj *ast.Node) bool {
			if obj == nil || obj.Kind != ast.KindObjectLiteralExpression {
				return false
			}
			cur := obj
			for cur.Parent != nil && cur.Parent.Kind == ast.KindParenthesizedExpression {
				cur = cur.Parent
			}
			parent := cur.Parent
			if parent == nil {
				return false
			}
			var calleeExpr *ast.Node
			var args *ast.NodeList
			switch parent.Kind {
			case ast.KindCallExpression:
				ce := parent.AsCallExpression()
				calleeExpr = ce.Expression
				args = ce.Arguments
			case ast.KindNewExpression:
				ne := parent.AsNewExpression()
				calleeExpr = ne.Expression
				args = ne.Arguments
			default:
				return false
			}
			// upstream's `node.parent.callee` accepts any non-callee
			// position; reject only when `obj` IS the callee itself.
			if args == nil {
				return false
			}
			inArgs := false
			for _, arg := range args.Nodes {
				if arg == cur {
					inArgs = true
					break
				}
			}
			if !inArgs {
				return false
			}
			// Reuse `IsCreateClassCall`'s callee matching by synthesizing
			// the equivalent shape — but that helper takes
			// `*CallExpression`, so for NewExpression we inline the same
			// check directly.
			callee := ast.SkipParentheses(calleeExpr)
			switch callee.Kind {
			case ast.KindIdentifier:
				return callee.AsIdentifier().Text == createClass
			case ast.KindPropertyAccessExpression:
				pa := callee.AsPropertyAccessExpression()
				obj := ast.SkipParentheses(pa.Expression)
				if obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != pragma {
					return false
				}
				name := pa.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return false
				}
				return name.AsIdentifier().Text == createClass
			}
			return false
		}

		return rule.RuleListeners{
			ast.KindClassDeclaration: func(node *ast.Node) {
				if !reactutil.ExtendsReactComponent(node, pragma) {
					return
				}
				checkLifeCycleMethods(node.Members())
			},
			ast.KindClassExpression: func(node *ast.Node) {
				if !reactutil.ExtendsReactComponent(node, pragma) {
					return
				}
				checkLifeCycleMethods(node.Members())
			},
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if !isES5Component(node) {
					return
				}
				ol := node.AsObjectLiteralExpression()
				if ol == nil || ol.Properties == nil {
					return
				}
				checkLifeCycleMethods(ol.Properties.Nodes)
			},
		}
	},
}
