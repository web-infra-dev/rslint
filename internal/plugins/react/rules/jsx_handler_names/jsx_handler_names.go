package jsx_handler_names

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	msgBadHandlerName = "Handler function for {{propKey}} prop key must be a camelCase name beginning with '{{handlerPrefix}}' only"
	msgBadPropKey     = "Prop key for {{propValue}} must begin with '{{handlerPropPrefix}}'"
)

type options struct {
	eventHandlerPrefix     string
	eventHandlerPropPrefix string
	checkLocalVariables    bool
	checkInlineFunction    bool
	ignoreComponentNames   []string

	// nil when the corresponding prefix is disabled (`false`); upstream's
	// `EVENT_HANDLER_REGEX` / `PROP_EVENT_HANDLER_REGEX` are also `null` in
	// that case and serve as a sentinel "this side is disabled" gate.
	eventHandlerRegex     *regexp.Regexp
	propEventHandlerRegex *regexp.Regexp
}

func parseOptions(raw any) options {
	opts := options{
		eventHandlerPrefix:     "handle",
		eventHandlerPropPrefix: "on",
	}
	prefixDisabled := false
	propPrefixDisabled := false

	m := utils.GetOptionsMap(raw)
	if m != nil {
		// Mirror upstream's `configuration.eventHandlerPrefix || 'handle'`:
		// `false` disables the side; any other falsy value (empty string,
		// missing key) falls back to the default. Only a non-empty string
		// overrides the default.
		if v, ok := m["eventHandlerPrefix"]; ok {
			switch val := v.(type) {
			case bool:
				if !val {
					prefixDisabled = true
				}
			case string:
				if val != "" {
					opts.eventHandlerPrefix = val
				}
			}
		}
		if v, ok := m["eventHandlerPropPrefix"]; ok {
			switch val := v.(type) {
			case bool:
				if !val {
					propPrefixDisabled = true
				}
			case string:
				if val != "" {
					opts.eventHandlerPropPrefix = val
				}
			}
		}
		if v, ok := m["checkLocalVariables"].(bool); ok {
			opts.checkLocalVariables = v
		}
		if v, ok := m["checkInlineFunction"].(bool); ok {
			opts.checkInlineFunction = v
		}
		if v, ok := m["ignoreComponentNames"].([]interface{}); ok {
			for _, p := range v {
				if s, ok := p.(string); ok {
					opts.ignoreComponentNames = append(opts.ignoreComponentNames, s)
				}
			}
		}
	}

	// Mirror upstream: when the corresponding prefix is `false`, the regex is
	// `null` and acts as a one-sided shut-off — disabling either prefix
	// disables the matching half of the rule entirely.
	//
	// Use `regexp.Compile` (not `MustCompile`) so a user prefix containing
	// unbalanced regex metacharacters — e.g. `(`, `[` — fails gracefully
	// instead of panicking. Upstream's `new RegExp(...)` throws a
	// SyntaxError that ESLint surfaces as a rule-loading error; we mirror
	// the "rule effectively disabled" outcome by leaving the regex `nil`,
	// which the `regex != nil` gate below treats the same as a `false`
	// prefix. Note: prefixes that are *valid* regex (`.+`, `[a-z]`, etc.)
	// are still interpreted as regex metacharacters, exactly as upstream
	// does — escaping them with `regexp.QuoteMeta` would diverge from
	// upstream's input/output semantics on those inputs.
	if !prefixDisabled {
		propAlt := opts.eventHandlerPropPrefix
		if propPrefixDisabled {
			propAlt = ""
		}
		if re, err := regexp.Compile(
			`^((props\.` + propAlt + `)|((.*\.)?` + opts.eventHandlerPrefix + `))[0-9]*[A-Z].*$`,
		); err == nil {
			opts.eventHandlerRegex = re
		}
	}
	if !propPrefixDisabled {
		if re, err := regexp.Compile(
			`^(` + opts.eventHandlerPropPrefix + `[A-Z].*|ref)$`,
		); err == nil {
			opts.propEventHandlerRegex = re
		}
	}

	if prefixDisabled {
		opts.eventHandlerPrefix = ""
	}
	if propPrefixDisabled {
		opts.eventHandlerPropPrefix = ""
	}
	return opts
}

// stripWhitespace removes every character that JS's `\s` regex matches,
// mirroring upstream's `propValue.replace(/\s*/g, ”)`. The set is JS
// WhiteSpace + LineTerminator per ECMA-262: ASCII whitespace, U+00A0
// NBSP, U+FEFF ZWNBSP, U+2028 / U+2029 line/paragraph separators, and
// Unicode "Space_Separator" (Zs) code points.
//
// Go's `unicode.IsSpace` differs from JS `\s` in two chars only:
//
//   - U+0085 NEL — Go matches, JS does not.
//   - U+FEFF BOM — Go does not match, JS does.
//
// We patch those two so the strip behavior is byte-for-byte aligned with
// JS — otherwise an exotic-whitespace member-access chain could produce
// different regex outcomes here vs. upstream.
func stripWhitespace(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isJSWhitespace(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func isJSWhitespace(r rune) bool {
	switch r {
	case '\u0085':
		// NEL: Go's unicode.IsSpace matches it, JS \s does not.
		return false
	case '\uFEFF':
		// BOM: JS \s matches it (legacy WhiteSpace category), Go's
		// unicode.IsSpace does not.
		return true
	}
	return unicode.IsSpace(r)
}

// stripThisOrBindBase mirrors upstream's `.replace(/^this\.|.*::/, ”)`
// (single, non-global). Alternatives are tried at the same position in
// declaration order: an anchored leading `this.` wins over a `::` rewrite,
// matching JS's left-to-right alternation semantics.
//
// For inputs like `this::handleChange` the leading `this.` does NOT match
// (literal dot vs. `:`), so the second alternative kicks in and strips
// through the last `::`. For `this.props.handleChange` the leading `this.`
// strips to `props.handleChange` and any later `::` is left alone.
func stripThisOrBindBase(s string) string {
	if strings.HasPrefix(s, "this.") {
		return s[len("this."):]
	}
	if idx := strings.LastIndex(s, "::"); idx != -1 {
		return s[idx+len("::"):]
	}
	return s
}

// arrowBodyCallExprCallee returns the callee of an arrow-function inline
// handler whose body is a CallExpression — the only inline shape upstream
// recognises as having a `body.callee`. Returns nil for arrow functions whose
// body is a Block, a literal, an Identifier, or any non-CallExpression form;
// callers treat that as "no recognisable inline call" and skip the inline
// check (mirrors upstream's `!body.callee` short-circuit).
//
// Parentheses are unwrapped at every step (body, then callee) so shapes like
// `() => (this.handleChange())` and `() => (this.handleChange)()` are handled
// identically to their flat ESTree equivalents.
func arrowBodyCallExprCallee(arrow *ast.Node) *ast.Node {
	if arrow == nil || !ast.IsArrowFunction(arrow) {
		return nil
	}
	body := ast.SkipParentheses(arrow.AsArrowFunction().Body)
	if body == nil || !ast.IsCallExpression(body) {
		return nil
	}
	return ast.SkipParentheses(body.AsCallExpression().Expression)
}

// isPlainMemberAccess reports whether a node is a non-optional member access —
// the shape upstream's `expression.object` truthiness check matches. tsgo
// encodes optional chains (`x?.y`, `x?.[y]`) as PropertyAccessExpression /
// ElementAccessExpression with the OptionalChain flag set; modern @typescript-
// eslint instead wraps those in `ChainExpression`, whose top-level node has
// no `.object` and so fails the gate in ESLint. Excluding optional-chain
// nodes here mirrors that gate, keeping `<X onChange={this?.foo} />` and
// `<X onChange={this?.props?.onChange} />` skipped under default options
// (matching ESLint v6+/typescript-eslint v6+ behavior).
func isPlainMemberAccess(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if !ast.IsPropertyAccessExpression(node) && !ast.IsElementAccessExpression(node) {
		return false
	}
	return !ast.IsOptionalChain(node)
}

var JsxHandlerNamesRule = rule.Rule{
	Name: "react/jsx-handler-names",
	Run: func(ctx rule.RuleContext, _rawOptions []any) rule.RuleListeners {
		rawOptions := rule.LegacyUnwrapOptions(_rawOptions)
		opts := parseOptions(rawOptions)

		return rule.RuleListeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				attr := node.AsJsxAttribute()

				// Component name — `<Foo>`, `<A.B>`, `<ns:Local>`. Upstream's
				// `getComponentName` returns the same dotted/colon string;
				// `GetJsxElementTypeString` is the canonical port.
				parent := reactutil.GetJsxParentElement(node)
				componentName := reactutil.GetJsxElementTypeString(parent)
				for _, pattern := range opts.ignoreComponentNames {
					if reactutil.MatchGlob(componentName, pattern) {
						return
					}
				}

				// Skip `attr` (no value), `attr="..."` (StringLiteral), and
				// `attr={}` (empty JsxExpression). Only `attr={expr}` proceeds.
				init := attr.Initializer
				if init == nil || !ast.IsJsxExpression(init) {
					return
				}
				rawExpr := init.AsJsxExpression().Expression
				if rawExpr == nil {
					return
				}
				// SkipParentheses mirrors ESLint's flat-AST view: `(x)` and
				// `((x))` both behave as `x` for the gating checks below and
				// for the propValue text. TS-only wrappers (`as`, `!`,
				// `satisfies`) are intentionally NOT peeled — upstream's
				// gating reads `.object` on the raw expression, which is
				// undefined for those wrappers, and we want the same
				// "treated as non-member" outcome.
				expression := ast.SkipParentheses(rawExpr)
				if expression == nil {
					return
				}

				isInline := ast.IsArrowFunction(expression)

				// `checkInlineFunction: false` (default) → ignore arrow inline
				// handlers entirely. Matches upstream's first guard.
				if !opts.checkInlineFunction && isInline {
					return
				}

				// `checkLocalVariables: false` (default) → restrict to handlers
				// reached through a member access. For inline handlers the
				// member access lives inside the arrow body (its CallExpression
				// callee); for non-inline it's the expression itself.
				if !opts.checkLocalVariables {
					if isInline {
						callee := arrowBodyCallExprCallee(expression)
						if !isPlainMemberAccess(callee) {
							return
						}
					} else {
						if !isPlainMemberAccess(expression) {
							return
						}
					}
				}

				// propKey — only Identifier-named attributes participate.
				// JsxNamespacedName attributes (`<X ns:attr={...} />`) fall out
				// here, matching upstream's effective behavior: ESLint sets
				// `propKey` to a node object for namespaced names, which then
				// fails every regex/string compare and produces no diagnostic.
				nameNode := attr.Name()
				if nameNode == nil || !ast.IsIdentifier(nameNode) {
					return
				}
				propKey := nameNode.AsIdentifier().Text
				if propKey == "ref" {
					return
				}

				// propValue text — for inline `checkInlineFunction` handlers,
				// upstream reads `getText(expression.body.callee)`; otherwise
				// `getText(expression)`. SkipParentheses on the inner pieces
				// keeps tsgo aligned with ESLint's flat view.
				var textNode *ast.Node
				if opts.checkInlineFunction && isInline {
					callee := arrowBodyCallExprCallee(expression)
					if callee == nil {
						// Reachable only when checkLocalVariables is true and
						// the body is not a CallExpression — upstream's
						// `expression.body.callee` is then `undefined` and
						// `getText(undefined)` returns "" in eslint-utils.
						// Mirror that: an empty propValue can't match either
						// regex, so it's a no-op.
						textNode = nil
					} else {
						textNode = callee
					}
				} else {
					textNode = expression
				}

				var propValue string
				if textNode != nil {
					propValue = stripWhitespace(utils.TrimmedNodeText(ctx.SourceFile, textNode))
					propValue = stripThisOrBindBase(propValue)
				}

				propIsEventHandler := opts.propEventHandlerRegex != nil && opts.propEventHandlerRegex.MatchString(propKey)
				propFnIsNamedCorrectly := opts.eventHandlerRegex != nil && opts.eventHandlerRegex.MatchString(propValue)

				switch {
				case propIsEventHandler && opts.eventHandlerRegex != nil && !propFnIsNamedCorrectly:
					data := map[string]string{
						"propKey":       propKey,
						"handlerPrefix": opts.eventHandlerPrefix,
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "badHandlerName",
						Description: reactutil.ApplyData(msgBadHandlerName, data),
						Data:        data,
					})
				case propFnIsNamedCorrectly && opts.propEventHandlerRegex != nil && !propIsEventHandler:
					data := map[string]string{
						"propValue":         propValue,
						"handlerPropPrefix": opts.eventHandlerPropPrefix,
					}
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "badPropKey",
						Description: reactutil.ApplyData(msgBadPropKey, data),
						Data:        data,
					})
				}
			},
		}
	},
}
