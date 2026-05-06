package no_array_index_key

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const noArrayIndexMessage = "Do not use Array index in keys"

// indexParamPositions maps tracked iterator method names to the 0-based
// position of the index parameter inside the callback. Mirrors upstream
// `iteratorFunctionsToIndexParamPosition`.
var indexParamPositions = map[string]int{
	"every":       1,
	"filter":      1,
	"find":        1,
	"findIndex":   1,
	"flatMap":     1,
	"forEach":     1,
	"map":         1,
	"reduce":      2,
	"reduceRight": 2,
	"some":        1,
}

// isImportSpecifierFromReact reports whether `ident` resolves to a binding
// introduced by `import { <anything> } from 'react'`. Mirrors upstream
// `eslint-plugin-react`'s `isCreateCloneElement` Identifier branch
// byte-for-byte:
//
//   - The binding's declaration must be an `ImportSpecifier` (the
//     `import { x } from 'pkg'` shape). Destructuring like
//     `const { cloneElement } = React`, aliasing like
//     `const cloneElement = React.cloneElement`, and
//     `const { cloneElement } = require('react')` are NOT recognized
//     — they're VariableDeclaration / BindingElement, not ImportSpecifier.
//
//   - The module specifier must be the LITERAL string `'react'`. The
//     `settings.react.pragma` configuration does NOT influence this
//     check upstream — `pragma` only governs the dotted-form receiver.
//
//   - The identifier's TEXT is intentionally NOT validated. Upstream
//     accepts any react-imported name (e.g. `import { foo } from 'react'`
//     also passes); this is an upstream quirk we mirror for 1:1 parity.
//
// Without a TypeChecker, falls back to a SourceFile-level scan for any
// matching ImportDeclaration — strictly less precise than scope-based
// resolution but sufficient for the canonical top-level patterns.
func isImportSpecifierFromReact(ctx rule.RuleContext, ident *ast.Node) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}

	if ctx.TypeChecker != nil {
		symbol := ctx.TypeChecker.GetSymbolAtLocation(ident)
		if symbol == nil {
			return false
		}
		var decl *ast.Node
		if symbol.ValueDeclaration != nil {
			decl = symbol.ValueDeclaration
		} else if len(symbol.Declarations) > 0 {
			decl = symbol.Declarations[0]
		}
		if decl == nil || decl.Kind != ast.KindImportSpecifier {
			return false
		}
		for p := decl.Parent; p != nil; p = p.Parent {
			if p.Kind != ast.KindImportDeclaration {
				continue
			}
			ms := p.AsImportDeclaration().ModuleSpecifier
			return ms != nil && ms.Kind == ast.KindStringLiteral && ms.Text() == "react"
		}
		return false
	}

	sf := ast.GetSourceFileOfNode(ident)
	if sf == nil {
		return false
	}
	name := ident.AsIdentifier().Text
	found := false
	var visit func(n *ast.Node)
	visit = func(n *ast.Node) {
		if found || n == nil {
			return
		}
		if n.Kind == ast.KindImportDeclaration {
			id := n.AsImportDeclaration()
			if id.ModuleSpecifier != nil &&
				id.ModuleSpecifier.Kind == ast.KindStringLiteral &&
				id.ModuleSpecifier.Text() == "react" &&
				id.ImportClause != nil {
				ic := id.ImportClause.AsImportClause()
				if ic.NamedBindings != nil && ic.NamedBindings.Kind == ast.KindNamedImports {
					ni := ic.NamedBindings.AsNamedImports()
					if ni.Elements != nil {
						for _, spec := range ni.Elements.Nodes {
							local := spec.Name()
							if local != nil && local.Kind == ast.KindIdentifier && local.AsIdentifier().Text == name {
								found = true
								return
							}
						}
					}
				}
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(sf.AsNode())
	return found
}

var NoArrayIndexKeyRule = rule.Rule{
	Name: "react/no-array-index-key",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		pragma := reactutil.GetReactPragma(ctx.Settings)

		// indexParamNames is a stack of identifier names introduced as the
		// index parameter of an enclosing tracked iterator callback. An
		// identifier reference whose text matches any name on the stack is
		// treated as the array index. Mirrors upstream's flat array — pushed
		// on iterator-call enter, popped on exit. The stack is intentionally
		// a flat name list (not a depth counter): nested iterators with the
		// SAME index name (`foo.map((a,i)=>bar.map((c,i)=>...))`) push the
		// name twice, and inner functions defined inside the callback do
		// NOT shadow it — both behaviors mirror upstream byte-for-byte.
		var indexParamNames []string

		isArrayIndex := func(node *ast.Node) bool {
			if node == nil {
				return false
			}
			node = ast.SkipParentheses(node)
			if node == nil || node.Kind != ast.KindIdentifier {
				return false
			}
			name := node.AsIdentifier().Text
			for _, n := range indexParamNames {
				if n == name {
					return true
				}
			}
			return false
		}

		// isUsingReactChildren mirrors upstream's `isUsingReactChildren`.
		// True when the callee is `Children.<m>` or `<pragma>.<X>.<m>`
		// for `<m>` ∈ {map, forEach}. Note: upstream's check on the outer
		// `<pragma>.X` form does NOT validate that the inner property is
		// literally `Children` — preserved here for parity even though it
		// is lenient (e.g. `React.Foo.map(...)` would also be treated as
		// Children-shape, but the missing args[1] callback then fails the
		// next gate, so it's benign in practice).
		isUsingReactChildren := func(call *ast.CallExpression) bool {
			if call == nil || call.Expression == nil {
				return false
			}
			callee := ast.SkipParentheses(call.Expression)
			if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
				return false
			}
			pa := callee.AsPropertyAccessExpression()
			name := pa.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				return false
			}
			method := name.AsIdentifier().Text
			if method != "map" && method != "forEach" {
				return false
			}
			obj := ast.SkipParentheses(pa.Expression)
			if obj == nil {
				return false
			}
			if obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == "Children" {
				return true
			}
			if obj.Kind == ast.KindPropertyAccessExpression {
				inner := obj.AsPropertyAccessExpression()
				innerObj := ast.SkipParentheses(inner.Expression)
				return innerObj != nil && innerObj.Kind == ast.KindIdentifier && innerObj.AsIdentifier().Text == pragma
			}
			return false
		}

		// getMapIndexParamName mirrors upstream's `getMapIndexParamName`.
		// Returns the parameter identifier name for the iterator's index
		// position, or "" if the call isn't a tracked iterator with a
		// function-like callback whose param at the expected position is
		// a plain (or default-valued) identifier. Destructured / rest
		// params yield "" — same observable behavior as upstream pushing
		// `undefined` (which never matches a real identifier reference).
		getMapIndexParamName := func(call *ast.CallExpression) string {
			if call == nil || call.Expression == nil {
				return ""
			}
			callee := ast.SkipParentheses(call.Expression)
			if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
				return ""
			}
			pa := callee.AsPropertyAccessExpression()
			name := pa.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				return ""
			}
			pos, ok := indexParamPositions[name.AsIdentifier().Text]
			if !ok {
				return ""
			}
			if call.Arguments == nil {
				return ""
			}
			args := call.Arguments.Nodes
			argIdx := 0
			if isUsingReactChildren(call) {
				argIdx = 1
			}
			if argIdx >= len(args) {
				return ""
			}
			callback := ast.SkipParentheses(args[argIdx])
			if callback == nil || !ast.IsFunctionExpressionOrArrowFunction(callback) {
				return ""
			}
			params := callback.Parameters()
			if pos >= len(params) {
				return ""
			}
			param := params[pos]
			if param == nil || param.Kind != ast.KindParameter {
				return ""
			}
			pd := param.AsParameterDeclaration()
			// Rest parameter (`...rest`) — upstream RestElement has no
			// `.name`; treat as un-named.
			if pd.DotDotDotToken != nil {
				return ""
			}
			// Default-valued parameter (`i = 0`) — upstream's parser
			// wraps the binding in an AssignmentPattern whose `.name`
			// is undefined, so the index is NOT pushed and references
			// to `i` inside the body are NOT reported. Mirror that:
			// when an Initializer is present, refuse to push, even
			// though tsgo's ParameterDeclaration exposes the inner
			// Identifier directly. Keeps input/output parity with
			// `eslint-plugin-react`.
			if pd.Initializer != nil {
				return ""
			}
			paramName := pd.Name()
			// Destructured patterns (`[i]` / `{i}`) and other binding
			// shapes — upstream `params[pos].name` returns undefined.
			if paramName == nil || paramName.Kind != ast.KindIdentifier {
				return ""
			}
			return paramName.AsIdentifier().Text
		}

		report := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "noArrayIndex",
				Description: noArrayIndexMessage,
			})
		}

		// collectIdentifiersFromBinary recursively flattens a chain of
		// BinaryExpressions into the bare identifiers they reference.
		// Mirrors upstream's `getIdentifiersFromBinaryExpression`: only
		// Identifier and BinaryExpression are walked — every other shape
		// (ConditionalExpression, CallExpression, MemberExpression, …)
		// is opaque, matching upstream's selective recursion.
		var collectIdentifiersFromBinary func(node *ast.Node) []*ast.Node
		collectIdentifiersFromBinary = func(node *ast.Node) []*ast.Node {
			if node == nil {
				return nil
			}
			node = ast.SkipParentheses(node)
			if node == nil {
				return nil
			}
			if node.Kind == ast.KindIdentifier {
				return []*ast.Node{node}
			}
			if node.Kind == ast.KindBinaryExpression {
				be := node.AsBinaryExpression()
				return append(collectIdentifiersFromBinary(be.Left), collectIdentifiersFromBinary(be.Right)...)
			}
			return nil
		}

		// checkPropValue inspects the value expression of a `key` JSX
		// attribute or a `key` property in createElement / cloneElement
		// props and reports any reference to a tracked array index.
		// Mirrors upstream's `checkPropValue` byte-for-byte:
		//   - Direct Identifier        → report on the value
		//   - TemplateExpression       → report on the value, once per
		//                                index-typed substitution (so
		//                                `\`${i}-${i}\`` reports twice)
		//   - BinaryExpression chain   → report on the value, once per
		//                                index-typed leaf identifier
		//   - `index.toString()`       → report on the call expression
		//   - `String(index)`          → report on the index argument
		// All other shapes (ConditionalExpression, ObjectLiteral, function
		// calls that aren't `String` / `.toString()`, etc.) are opaque.
		checkPropValue := func(node *ast.Node) {
			if node == nil {
				return
			}
			node = ast.SkipParentheses(node)
			if node == nil {
				return
			}

			if isArrayIndex(node) {
				report(node)
				return
			}

			if node.Kind == ast.KindTemplateExpression {
				te := node.AsTemplateExpression()
				if te.TemplateSpans != nil {
					for _, span := range te.TemplateSpans.Nodes {
						if span.Kind != ast.KindTemplateSpan {
							continue
						}
						if isArrayIndex(span.AsTemplateSpan().Expression) {
							report(node)
						}
					}
				}
				return
			}

			if node.Kind == ast.KindBinaryExpression {
				for _, id := range collectIdentifiersFromBinary(node) {
					if isArrayIndex(id) {
						report(node)
					}
				}
				return
			}

			if node.Kind == ast.KindCallExpression {
				// Upstream's `astUtil.isCallExpression` accepts only
				// `CallExpression`, NOT `OptionalCallExpression`. tsgo
				// collapses both into `KindCallExpression`, so we must
				// gate on `IsOptionalChain` to keep parity:
				// `key={i?.toString()}` / `key={String?.(i)}` are NOT
				// reported upstream and must NOT report here.
				if ast.IsOptionalChain(node) {
					return
				}
				call := node.AsCallExpression()
				ce := ast.SkipParentheses(call.Expression)

				if ce != nil && ce.Kind == ast.KindPropertyAccessExpression {
					// Upstream `node.callee.type === 'MemberExpression'`
					// rejects `OptionalMemberExpression`, so an optional
					// `i?.toString` chain on the callee is opaque too.
					if ast.IsOptionalChain(ce) {
						return
					}
					pa := ce.AsPropertyAccessExpression()
					prop := pa.Name()
					if isArrayIndex(pa.Expression) && prop != nil && prop.Kind == ast.KindIdentifier && prop.AsIdentifier().Text == "toString" {
						report(node)
						return
					}
				}

				if ce != nil && ce.Kind == ast.KindIdentifier && ce.AsIdentifier().Text == "String" {
					if call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
						firstArg := call.Arguments.Nodes[0]
						if isArrayIndex(firstArg) {
							// Report on the unwrapped Identifier — ESTree
							// flattens parens, so upstream's `node.arguments[0]`
							// is the bare Identifier; tsgo preserves the
							// wrapping ParenthesizedExpression, so reporting
							// `firstArg` directly would widen the position
							// range relative to upstream. Skip parens to
							// match upstream's report range exactly.
							report(ast.SkipParentheses(firstArg))
						}
					}
				}
			}
		}

		// checkObjectKeyProp walks the props object passed to
		// createElement / cloneElement and runs `checkPropValue` on the
		// initializer of any property named `key`. Computed /
		// string-literal / numeric keys and spread elements are skipped
		// — matches upstream's `prop.key.name === 'key'` guard which
		// only fires for Identifier keys. SHORTHAND properties are
		// included: upstream sees `prop.key` as Identifier `key` and
		// `prop.value` as the same Identifier, so `{ key }` does match
		// when the iterator's index parameter happens to be named `key`
		// (e.g. `foo.map((bar, key) => React.cloneElement(c, { key }))`).
		// keyMatchesUpstreamIdentifier mirrors upstream's
		// `prop.key.name === 'key'` test in ESTree:
		//   - `key: ...`              (Identifier key)               → name is 'key'
		//   - `{ key }`               (Shorthand)                    → name is 'key'
		//   - `[key]: ...`            (Computed, inner is Identifier `key`) → name is 'key'
		//
		// Other key shapes (StringLiteral / NumericLiteral / `[expr]`
		// where expr is not an Identifier named `key`, PrivateIdentifier)
		// have either `prop.key.name === undefined` or a different
		// `.name`, and upstream rejects them.
		keyMatchesUpstreamIdentifier := func(keyName *ast.Node) bool {
			if keyName == nil {
				return false
			}
			// ComputedPropertyName: ESTree puts the inner expression
			// directly as `prop.key` (with `computed: true`), so an
			// Identifier `key` inside `[key]` reads `prop.key.name === 'key'`.
			if keyName.Kind == ast.KindComputedPropertyName {
				inner := keyName.AsComputedPropertyName().Expression
				return inner != nil && inner.Kind == ast.KindIdentifier && inner.AsIdentifier().Text == "key"
			}
			return keyName.Kind == ast.KindIdentifier && keyName.AsIdentifier().Text == "key"
		}

		checkObjectKeyProp := func(props *ast.Node) {
			props = ast.SkipParentheses(props)
			if props == nil || props.Kind != ast.KindObjectLiteralExpression {
				return
			}
			obj := props.AsObjectLiteralExpression()
			if obj.Properties == nil {
				return
			}
			for _, prop := range obj.Properties.Nodes {
				switch prop.Kind {
				case ast.KindPropertyAssignment:
					pa := prop.AsPropertyAssignment()
					if !keyMatchesUpstreamIdentifier(pa.Name()) {
						continue
					}
					if pa.Initializer != nil {
						checkPropValue(pa.Initializer)
					}
				case ast.KindShorthandPropertyAssignment:
					// `{ key }` — both the property's name AND the
					// referenced binding are the Identifier `key`.
					// Upstream's `checkPropValue(prop.value)` runs on
					// that Identifier, which matches the index stack
					// only when the callback's index parameter is
					// itself named `key`.
					spa := prop.AsShorthandPropertyAssignment()
					nameNode := spa.Name()
					if !keyMatchesUpstreamIdentifier(nameNode) {
						continue
					}
					checkPropValue(nameNode)
				}
			}
		}

		// isCreateCloneElement mirrors upstream's `isCreateCloneElement`
		// byte-for-byte:
		//
		//   - Member-access (or optional-member-access) callee:
		//     `<pragma>.createElement` / `<pragma>.cloneElement`. Parens
		//     on the pragma sub-expression are skipped (ESTree
		//     flattens), TS expression wrappers are NOT (ESLint's JS
		//     parser cannot produce them).
		//
		//   - Bare Identifier callee: ANY identifier whose binding
		//     resolves to an `import { <name> } from 'react'`
		//     ImportSpecifier (the literal string 'react' is hardcoded
		//     upstream — `pragma` setting does NOT change it). Crucially,
		//     upstream does NOT verify that the imported name is
		//     `createElement` / `cloneElement`: any react-imported name
		//     is accepted. This is technically an upstream bug, but we
		//     mirror it exactly for 1:1 parity.
		//
		// Constructions like `const { cloneElement } = React`,
		// `const cloneElement = React.cloneElement`, and
		// `const { cloneElement } = require('react')` are NOT
		// recognized — upstream's `findVariableByName` returns the
		// VariableDeclaration, whose `.type` is not `'ImportSpecifier'`.
		isCreateCloneElement := func(callee *ast.Node) bool {
			if callee == nil {
				return false
			}
			callee = ast.SkipParentheses(callee)
			if callee == nil {
				return false
			}

			if callee.Kind == ast.KindPropertyAccessExpression {
				pa := callee.AsPropertyAccessExpression()
				name := pa.Name()
				if name == nil || name.Kind != ast.KindIdentifier {
					return false
				}
				txt := name.AsIdentifier().Text
				if txt != "createElement" && txt != "cloneElement" {
					return false
				}
				obj := ast.SkipParentheses(pa.Expression)
				return obj != nil && obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == pragma
			}

			if callee.Kind == ast.KindIdentifier {
				return isImportSpecifierFromReact(ctx, callee)
			}

			return false
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()

				if isCreateCloneElement(call.Expression) && call.Arguments != nil && len(call.Arguments.Nodes) > 1 {
					if len(indexParamNames) == 0 {
						return
					}
					checkObjectKeyProp(call.Arguments.Nodes[1])
					return
				}

				if name := getMapIndexParamName(call); name != "" {
					indexParamNames = append(indexParamNames, name)
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if name := getMapIndexParamName(node.AsCallExpression()); name != "" && len(indexParamNames) > 0 {
					indexParamNames = indexParamNames[:len(indexParamNames)-1]
				}
			},
			ast.KindJsxAttribute: func(node *ast.Node) {
				attr := node.AsJsxAttribute()
				name := attr.Name()
				if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "key" {
					return
				}
				if len(indexParamNames) == 0 {
					return
				}
				init := attr.Initializer
				if init == nil || init.Kind != ast.KindJsxExpression {
					return
				}
				expr := init.AsJsxExpression().Expression
				if expr == nil {
					return
				}
				checkPropValue(expr)
			},
		}
	},
}
