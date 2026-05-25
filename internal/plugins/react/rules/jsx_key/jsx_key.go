package jsx_key

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type options struct {
	checkFragmentShorthand   bool
	checkKeyMustBeforeSpread bool
	warnOnDuplicates         bool
}

func parseOptions(raw any) options {
	opts := options{}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["checkFragmentShorthand"].(bool); ok {
		opts.checkFragmentShorthand = v
	}
	if v, ok := m["checkKeyMustBeforeSpread"].(bool); ok {
		opts.checkKeyMustBeforeSpread = v
	}
	if v, ok := m["warnOnDuplicates"].(bool); ok {
		opts.warnOnDuplicates = v
	}
	return opts
}

var JsxKeyRule = rule.Rule{
	Name: "react/jsx-key",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		reactPragma := reactutil.GetReactPragma(ctx.Settings)
		fragmentPragma := reactutil.GetReactFragmentPragma(ctx.Settings)

		missingIterKeyUsePragDesc := `Missing "key" prop for element in iterator. Shorthand fragment syntax does not support providing keys. Use ` + reactPragma + `.` + fragmentPragma + ` instead`
		missingArrayKeyUsePragDesc := `Missing "key" prop for element in array. Shorthand fragment syntax does not support providing keys. Use ` + reactPragma + `.` + fragmentPragma + ` instead`

		// isWithinChildrenToArray mirrors eslint-plugin-react's flag of the
		// same name: a plain boolean flipped on enter / off on exit of a
		// `<pragma>.Children.toArray(...)` or `Children.toArray(...)` call.
		//
		// Intentionally a boolean, not a depth counter — upstream's
		// implementation has an observable quirk for nested Children.toArray:
		// an inner call's exit resets the flag even while the outer call
		// still encloses the code. Real input
		//   React.Children.toArray([React.Children.toArray(a), xs.map(x=><A/>)])
		// reports `missingIterKey` on `<A/>` under ESLint, because the inner
		// exit clobbered the flag before `xs.map` was visited. A depth counter
		// would silently diverge here; mirroring the boolean keeps 1:1 parity
		// with eslint-plugin-react.
		isWithinChildrenToArray := false

		reportMissingIterKey := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingIterKey",
				Description: `Missing "key" prop for element in iterator`,
			})
		}
		reportMissingIterKeyUsePrag := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingIterKeyUsePrag",
				Description: missingIterKeyUsePragDesc,
			})
		}
		reportMissingArrayKey := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingArrayKey",
				Description: `Missing "key" prop for element in array`,
			})
		}
		reportMissingArrayKeyUsePrag := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "missingArrayKeyUsePrag",
				Description: missingArrayKeyUsePragDesc,
			})
		}
		reportKeyBeforeSpread := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "keyBeforeSpread",
				Description: "`key` prop must be placed before any `{...spread}, to avoid conflicting with React\u2019s new JSX transform: https://reactjs.org/blog/2020/09/22/introducing-the-new-jsx-transform.html`",
			})
		}
		reportNonUniqueKeys := func(node *ast.Node) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "nonUniqueKeys",
				Description: "`key` prop must be unique",
			})
		}

		// checkIteratorElement handles a node that was identified as the
		// "JSX return" of a map/from callback. Non-JSX nodes are no-ops,
		// matching eslint-plugin-react's helper of the same name.
		checkIteratorElement := func(node *ast.Node) {
			if node == nil {
				return
			}
			if isJsxElementLike(node) {
				attrs := getJsxAttributeProps(node)
				if !hasKeyAttribute(attrs) {
					reportMissingIterKey(node)
					return
				}
				if opts.checkKeyMustBeforeSpread && isKeyAfterSpread(attrs) {
					reportKeyBeforeSpread(node)
				}
				return
			}
			if opts.checkFragmentShorthand && ast.IsJsxFragment(node) {
				reportMissingIterKeyUsePrag(node)
			}
		}

		// peelToJsxCandidates peels ONE level of ternary / logical off the
		// given expression and calls checkIteratorElement on every JSX leaf
		// it exposes. Intentionally shallow: upstream only inspects
		// consequent/alternate of a single ternary, or .right of a single
		// logical — it does NOT recurse. Deeply nested ternaries intentionally
		// slip through (matches eslint-plugin-react).
		peelToJsxCandidates := func(expr *ast.Node) {
			expr = ast.SkipParentheses(expr)
			if expr == nil {
				return
			}
			if ast.IsConditionalExpression(expr) {
				ce := expr.AsConditionalExpression()
				if ce.WhenTrue != nil {
					t := ast.SkipParentheses(ce.WhenTrue)
					if isJsxNode(t) {
						checkIteratorElement(t)
					}
				}
				if ce.WhenFalse != nil {
					f := ast.SkipParentheses(ce.WhenFalse)
					if isJsxNode(f) {
						checkIteratorElement(f)
					}
				}
				return
			}
			if ast.IsLogicalOrCoalescingBinaryExpression(expr) {
				right := expr.AsBinaryExpression().Right
				if right != nil {
					r := ast.SkipParentheses(right)
					if isJsxNode(r) {
						checkIteratorElement(r)
					}
				}
				return
			}
			checkIteratorElement(expr)
		}

		// checkArrowBodyJSX mirrors upstream checkArrowFunctionWithJSX.
		// Only arrows with expression bodies go through this path; block
		// bodies are handled by checkFunctionsBlockStatement below.
		checkArrowBodyJSX := func(fn *ast.Node) {
			if !ast.IsArrowFunction(fn) {
				return
			}
			body := fn.AsArrowFunction().Body
			if body == nil {
				return
			}
			peelToJsxCandidates(body)
		}

		// checkFunctionsBlockStatement mirrors upstream's helper: walk the
		// top-level `return` statements of a block body (descending only
		// through `if` branches) and dispatch each argument.
		checkFunctionsBlockStatement := func(fn *ast.Node) {
			if !ast.IsFunctionExpressionOrArrowFunction(fn) {
				return
			}
			var body *ast.Node
			if ast.IsArrowFunction(fn) {
				body = fn.AsArrowFunction().Body
			} else {
				body = fn.AsFunctionExpression().Body
			}
			if body == nil || !ast.IsBlock(body) {
				return
			}
			var returns []*ast.Node
			returns = collectReturnStatements(body, returns)
			for _, rs := range returns {
				arg := rs.AsReturnStatement().Expression
				if arg == nil {
					continue
				}
				peelToJsxCandidates(arg)
			}
		}

		processMapOrFromCallback := func(fn *ast.Node) {
			fn = ast.SkipParentheses(fn)
			if !ast.IsFunctionExpressionOrArrowFunction(fn) {
				return
			}
			checkArrowBodyJSX(fn)
			checkFunctionsBlockStatement(fn)
		}

		processArrayLikeSiblings := func(elements []*ast.Node, parentNode *ast.Node, inArray bool) {
			if len(elements) == 0 {
				return
			}
			var jsxEls []*ast.Node
			for _, el := range elements {
				if isJsxElementLike(el) {
					jsxEls = append(jsxEls, el)
				}
			}
			if len(jsxEls) == 0 {
				return
			}
			var keysByText map[string][]*ast.Node
			if opts.warnOnDuplicates {
				keysByText = map[string][]*ast.Node{}
			}
			for _, el := range jsxEls {
				attrs := getJsxAttributeProps(el)
				keyAttrs := getKeyAttributes(attrs)
				if len(keyAttrs) == 0 {
					if inArray {
						reportMissingArrayKey(el)
					}
					continue
				}
				for _, k := range keyAttrs {
					if opts.warnOnDuplicates {
						text := keyValueText(ctx.SourceFile, k)
						keysByText[text] = append(keysByText[text], k)
					}
					if opts.checkKeyMustBeforeSpread && isKeyAfterSpread(attrs) {
						// Upstream reports on the container (the array or
						// outer JSX parent), not on the offending element.
						reportKeyBeforeSpread(parentNode)
					}
				}
			}
			for _, group := range keysByText {
				if len(group) <= 1 {
					continue
				}
				for _, attr := range group {
					reportNonUniqueKeys(attr)
				}
			}
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if isChildrenToArrayCall(call, reactPragma) {
					isWithinChildrenToArray = true
					return
				}
				if isWithinChildrenToArray {
					return
				}
				switch calleePropertyName(call) {
				case "map":
					if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
						return
					}
					processMapOrFromCallback(call.Arguments.Nodes[0])
				case "from":
					if call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
						return
					}
					processMapOrFromCallback(call.Arguments.Nodes[1])
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if isChildrenToArrayCall(node.AsCallExpression(), reactPragma) {
					isWithinChildrenToArray = false
				}
			},
			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				if isWithinChildrenToArray {
					return
				}
				arr := node.AsArrayLiteralExpression()
				if arr.Elements == nil {
					return
				}
				processArrayLikeSiblings(arr.Elements.Nodes, node, true)
			},
			ast.KindJsxElement: func(node *ast.Node) {
				if isWithinChildrenToArray {
					return
				}
				jsxEl := node.AsJsxElement()
				if jsxEl.Children == nil {
					return
				}
				processArrayLikeSiblings(jsxEl.Children.Nodes, node, false)
			},
			ast.KindJsxFragment: func(node *ast.Node) {
				if isWithinChildrenToArray || !opts.checkFragmentShorthand {
					return
				}
				parent := node.Parent
				for parent != nil && ast.IsParenthesizedExpression(parent) {
					parent = parent.Parent
				}
				if parent != nil && ast.IsArrayLiteralExpression(parent) {
					reportMissingArrayKeyUsePrag(node)
				}
			},
		}
	},
}

// isJsxElementLike / isJsxNode shadow reactutil.IsJsxElementLike /
// reactutil.IsJsxLike — kept as local aliases so call sites read tightly.
var (
	isJsxElementLike = reactutil.IsJsxElementLike
	isJsxNode        = reactutil.IsJsxLike
)

// getJsxAttributeProps returns the JsxAttributes.Properties list for a
// JsxElement or JsxSelfClosingElement, or nil otherwise.
func getJsxAttributeProps(node *ast.Node) []*ast.Node {
	if node == nil {
		return nil
	}
	var attrs *ast.Node
	switch {
	case ast.IsJsxElement(node):
		opening := node.AsJsxElement().OpeningElement
		if opening == nil {
			return nil
		}
		attrs = opening.AsJsxOpeningElement().Attributes
	case ast.IsJsxSelfClosingElement(node):
		attrs = node.AsJsxSelfClosingElement().Attributes
	default:
		return nil
	}
	if attrs == nil {
		return nil
	}
	props := attrs.AsJsxAttributes().Properties
	if props == nil {
		return nil
	}
	return props.Nodes
}

// getKeyAttributes returns the JsxAttribute nodes in `attrs` whose attribute
// name is the identifier `key`. Mirrors jsx-ast-utils' `hasProp` behavior:
// JsxSpreadAttribute is opaque (a `{...{key: x}}` spread is NOT a match).
func getKeyAttributes(attrs []*ast.Node) []*ast.Node {
	var out []*ast.Node
	for _, a := range attrs {
		if !ast.IsJsxAttribute(a) {
			continue
		}
		name := a.AsJsxAttribute().Name()
		if !ast.IsIdentifier(name) {
			continue
		}
		if name.AsIdentifier().Text == "key" {
			out = append(out, a)
		}
	}
	return out
}

func hasKeyAttribute(attrs []*ast.Node) bool {
	return len(getKeyAttributes(attrs)) > 0
}

// isKeyAfterSpread reports whether a `key` JsxAttribute appears after a
// JsxSpreadAttribute in the attribute list.
func isKeyAfterSpread(attrs []*ast.Node) bool {
	sawSpread := false
	for _, a := range attrs {
		if ast.IsJsxSpreadAttribute(a) {
			sawSpread = true
			continue
		}
		if !ast.IsJsxAttribute(a) || !sawSpread {
			continue
		}
		name := a.AsJsxAttribute().Name()
		if ast.IsIdentifier(name) && name.AsIdentifier().Text == "key" {
			return true
		}
	}
	return false
}

// calleePropertyName returns the property-access name for a call's callee,
// e.g. `"map"` for `xs.map(...)` / `xs?.map(...)` / `xs.map?.(...)`. Returns
// the empty string for bracket access or non-property callees.
func calleePropertyName(call *ast.CallExpression) string {
	if call == nil || call.Expression == nil {
		return ""
	}
	callee := ast.SkipParentheses(call.Expression)
	if !ast.IsPropertyAccessExpression(callee) {
		return ""
	}
	name := callee.AsPropertyAccessExpression().Name()
	if !ast.IsIdentifier(name) {
		return ""
	}
	return name.AsIdentifier().Text
}

// isChildrenToArrayCall returns true when `call`'s callee matches either
// `<pragma>.Children.toArray` or `Children.toArray`. Parens are skipped on
// both the callee and its sub-expressions.
func isChildrenToArrayCall(call *ast.CallExpression, pragma string) bool {
	if call == nil || call.Expression == nil {
		return false
	}
	if pragma == "" {
		pragma = reactutil.DefaultReactPragma
	}
	callee := ast.SkipParentheses(call.Expression)
	if !ast.IsPropertyAccessExpression(callee) {
		return false
	}
	pa := callee.AsPropertyAccessExpression()
	name := pa.Name()
	if !ast.IsIdentifier(name) || name.AsIdentifier().Text != "toArray" {
		return false
	}
	obj := ast.SkipParentheses(pa.Expression)
	switch {
	case ast.IsIdentifier(obj):
		return obj.AsIdentifier().Text == "Children"
	case ast.IsPropertyAccessExpression(obj):
		innerPa := obj.AsPropertyAccessExpression()
		innerName := innerPa.Name()
		if !ast.IsIdentifier(innerName) || innerName.AsIdentifier().Text != "Children" {
			return false
		}
		inner2 := ast.SkipParentheses(innerPa.Expression)
		return ast.IsIdentifier(inner2) && inner2.AsIdentifier().Text == pragma
	}
	return false
}

// collectReturnStatements mirrors eslint-plugin-react's getReturnStatements.
// From a function body BlockStatement it collects top-level ReturnStatements,
// descending only into IfStatement branches. Nested function bodies,
// switch/for/while/try bodies, etc. are intentionally NOT traversed — this
// matches the upstream semantics exactly.
func collectReturnStatements(node *ast.Node, out []*ast.Node) []*ast.Node {
	if node == nil {
		return out
	}
	switch {
	case ast.IsIfStatement(node):
		ifs := node.AsIfStatement()
		if ifs.ThenStatement != nil {
			out = collectReturnStatements(ifs.ThenStatement, out)
		}
		if ifs.ElseStatement != nil {
			out = collectReturnStatements(ifs.ElseStatement, out)
		}
	case ast.IsReturnStatement(node):
		out = append(out, node)
	case ast.IsBlock(node):
		stmts := node.AsBlock().Statements
		if stmts == nil {
			return out
		}
		for _, stmt := range stmts.Nodes {
			switch {
			case ast.IsIfStatement(stmt):
				out = collectReturnStatements(stmt, out)
			case ast.IsReturnStatement(stmt):
				out = append(out, stmt)
			}
		}
	}
	return out
}

// keyValueText returns the raw source text of a JsxAttribute's value — used
// to group identical keys under `warnOnDuplicates`. Matches upstream's
// `SourceCode.getText(attr.value)`: two keys are duplicates iff their
// initializer source text is byte-identical. `<Foo key />` (no initializer,
// = `key={true}` shorthand) yields the empty string; multiple such elements
// do cluster together under warnOnDuplicates, same as upstream treating
// `undefined` consistently.
func keyValueText(sf *ast.SourceFile, attr *ast.Node) string {
	if attr == nil || sf == nil {
		return ""
	}
	init := attr.AsJsxAttribute().Initializer
	if init == nil {
		return ""
	}
	return utils.TrimmedNodeText(sf, init)
}
