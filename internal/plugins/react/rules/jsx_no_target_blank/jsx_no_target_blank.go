package jsx_no_target_blank

import (
	"regexp"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	msgNoreferrer = `Using target="_blank" without rel="noreferrer" (which implies rel="noopener") is a security risk in older browsers: see https://mathiasbynens.github.io/rel-noopener/#recommendations`
	msgNoopener   = `Using target="_blank" without rel="noreferrer" or rel="noopener" (the former implies the latter and is preferred due to wider support) is a security risk: see https://mathiasbynens.github.io/rel-noopener/#recommendations`
)

// externalLinkRe mirrors upstream's `/^(?:\w+:|\/\/)/` — matches absolute URLs
// (`http://`, `mailto:`, …) and protocol-relative URLs (`//example.com`).
var externalLinkRe = regexp.MustCompile(`^(?:\w+:|//)`)

type options struct {
	allowReferrer          bool
	enforceDynamicLinks    string // "always" | "never"
	warnOnSpreadAttributes bool
	links                  bool
	forms                  bool
}

func parseOptions(raw any) options {
	opts := options{
		allowReferrer:          false,
		enforceDynamicLinks:    "always",
		warnOnSpreadAttributes: false,
		links:                  true,
		forms:                  false,
	}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["allowReferrer"].(bool); ok {
		opts.allowReferrer = v
	}
	if v, ok := m["enforceDynamicLinks"].(string); ok && (v == "always" || v == "never") {
		opts.enforceDynamicLinks = v
	}
	if v, ok := m["warnOnSpreadAttributes"].(bool); ok {
		opts.warnOnSpreadAttributes = v
	}
	if v, ok := m["links"].(bool); ok {
		opts.links = v
	}
	if v, ok := m["forms"].(bool); ok {
		opts.forms = v
	}
	return opts
}

// componentMap maps a component tag name (e.g. "a", "Link") to the set of
// attribute names that identify its link target (e.g. ["href"] or ["to"]).
type componentMap map[string][]string

func newDefaultLinkComponents() componentMap {
	return componentMap{"a": {"href"}}
}

func newDefaultFormComponents() componentMap {
	return componentMap{"form": {"action"}}
}

// readComponentsFromSettings extracts a component-name→attribute-list map
// from `settings.<key>`, matching upstream `util/linkComponents`.
//
// Shapes accepted (each entry may appear standalone or as an element of an
// outer array, mirroring upstream's `DEFAULT.concat(settings[key] || [])`):
//
//   - string: "Link"                                    → {Link: [defaultAttr]}
//   - {name, <attrField>}: <attrField> string or []str  → {name: [attr…]}
//
// `attrField` is "linkAttribute" for linkComponents and "formAttribute" for
// formComponents — upstream uses distinct field names for each category
// (`value.linkAttribute` vs `value.formAttribute`), so getting this wrong
// would silently fall back to the default attribute for every custom form
// component the user configures.
func readComponentsFromSettings(settings map[string]interface{}, key, attrField, defaultAttr string, base componentMap) componentMap {
	out := componentMap{}
	for k, v := range base {
		out[k] = slices.Clone(v)
	}
	if settings == nil {
		return out
	}
	raw, ok := settings[key]
	if !ok {
		return out
	}
	addOne := func(entry interface{}) {
		switch e := entry.(type) {
		case string:
			out[e] = appendUnique(out[e], defaultAttr)
		case map[string]interface{}:
			name, _ := e["name"].(string)
			if name == "" {
				return
			}
			var attrs []string
			// Mirrors upstream's `[].concat(value[attrField])` coercion:
			// string → single-element list, array → as-is, missing → empty
			// (which we backfill with the default attribute).
			switch la := e[attrField].(type) {
			case string:
				attrs = []string{la}
			case []interface{}:
				for _, v := range la {
					if s, ok := v.(string); ok {
						attrs = append(attrs, s)
					}
				}
			}
			if len(attrs) == 0 {
				attrs = []string{defaultAttr}
			}
			for _, a := range attrs {
				out[name] = appendUnique(out[name], a)
			}
		}
	}
	// Upstream accepts either a single entry (string/object) or an array of
	// them at `settings[key]`. JS's `[].concat(x)` flattens both into the
	// final list; we mirror that by accepting either shape here.
	switch r := raw.(type) {
	case string:
		addOne(r)
	case map[string]interface{}:
		addOne(r)
	case []interface{}:
		for _, entry := range r {
			addOne(entry)
		}
	}
	return out
}

func appendUnique(list []string, s string) []string {
	if slices.Contains(list, s) {
		return list
	}
	return append(list, s)
}

// jsxExpressionInner unwraps a JsxExpression container and transparently
// skips ParenthesizedExpression wrappers on its payload. tsgo preserves
// parentheses as explicit nodes where ESTree flattens them; every downstream
// `.Kind` / `.As*()` access in this rule goes through this helper so that
// shapes like `{(…)}` and `{( (…) )}` are handled identically to `{…}`.
//
// Returns nil when the container is empty (`{}`) or not a JsxExpression.
func jsxExpressionInner(node *ast.Node) *ast.Node {
	if node == nil || node.Kind != ast.KindJsxExpression {
		return nil
	}
	expr := node.AsJsxExpression().Expression
	if expr == nil {
		return nil
	}
	return ast.SkipParentheses(expr)
}

// stringLiteralText returns the text of a StringLiteral node (the direct
// attribute-value form for `attr="x"`). Parentheses are skipped; every other
// node kind — including NoSubstitutionTemplateLiteral — returns "", false,
// matching upstream's `value.type === 'Literal' && typeof value.value ===
// 'string'` guard for the `target` attribute.
func stringLiteralText(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindStringLiteral {
		return n.AsStringLiteral().Text, true
	}
	return "", false
}

// templateOrStringText extends stringLiteralText by also accepting a
// NoSubstitutionTemplateLiteral — this mirrors upstream's `getStringFromValue`
// branch that reads `TemplateLiteral.quasis[0].value.cooked` for the `rel`
// attribute. Parentheses are skipped.
//
// TemplateExpression (templates with `${}`) is deliberately not handled:
// upstream would read only the first quasi, but rslint treats the value as
// non-literal so the enclosing check falls through to "non-string branch" —
// the same effective outcome for every rel string we care about.
func templateOrStringText(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	n := ast.SkipParentheses(node)
	switch n.Kind {
	case ast.KindStringLiteral:
		return n.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return n.AsNoSubstitutionTemplateLiteral().Text, true
	}
	return "", false
}

// attributeValuePossiblyBlank reports whether a `target` attribute could
// evaluate to "_blank" (case-insensitive). Matches upstream:
//
//   - StringLiteral directly (`target="_blank"`) → check the text
//   - JsxExpression containing StringLiteral (`target={"_blank"}`) → check
//   - JsxExpression containing ConditionalExpression → either branch is a
//     StringLiteral equal to "_blank"
//
// Parentheses are transparently skipped everywhere a child node kind is
// examined. NoSubstitutionTemplateLiteral is deliberately excluded for strict
// upstream parity (upstream's check is `expr.type === 'Literal'`, which
// excludes templates).
func attributeValuePossiblyBlank(attr *ast.Node) bool {
	if attr == nil {
		return false
	}
	init := attr.AsJsxAttribute().Initializer
	if init == nil {
		return false
	}
	// Direct `attr="_blank"` form.
	if s, ok := stringLiteralText(init); ok {
		return strings.EqualFold(s, "_blank")
	}
	// `attr={…}` form — unwrap the JsxExpression and any paren wrappers.
	expr := jsxExpressionInner(init)
	if expr == nil {
		return false
	}
	if s, ok := stringLiteralText(expr); ok {
		return strings.EqualFold(s, "_blank")
	}
	if expr.Kind == ast.KindConditionalExpression {
		cond := expr.AsConditionalExpression()
		if s, ok := stringLiteralText(cond.WhenTrue); ok && strings.EqualFold(s, "_blank") {
			return true
		}
		if s, ok := stringLiteralText(cond.WhenFalse); ok && strings.EqualFold(s, "_blank") {
			return true
		}
	}
	return false
}

// findLastIndex returns the largest i for which pred(attrs[i]) is true, or -1.
func findLastIndex(attrs []*ast.Node, pred func(*ast.Node) bool) int {
	for i := len(attrs) - 1; i >= 0; i-- {
		if pred(attrs[i]) {
			return i
		}
	}
	return -1
}

// attrHasName reports whether an attribute is a JsxAttribute whose name
// display string equals `name`. reactutil.GetJsxPropName returns "spread" for
// JsxSpreadAttribute and "ns:local" for JsxNamespacedName, neither of which
// overlaps with the `target` / `rel` / `href` names we compare against —
// matching upstream's `attr.name && attr.name.name === 'target'` guard which
// implicitly skips spreads and namespaced attributes.
func attrHasName(attr *ast.Node, name string) bool {
	if attr == nil || attr.Kind != ast.KindJsxAttribute {
		return false
	}
	return reactutil.GetJsxPropName(attr) == name
}

func attrNameIsOneOf(attr *ast.Node, names []string) bool {
	if attr == nil || attr.Kind != ast.KindJsxAttribute {
		return false
	}
	return slices.Contains(names, reactutil.GetJsxPropName(attr))
}

// isExternalURL reports whether a string literal value looks like an external
// URL under upstream's `/^(?:\w+:|\/\/)/` regex (absolute `proto:` or
// protocol-relative `//`).
func isExternalURL(s string) bool {
	return externalLinkRe.MatchString(s)
}

func hasExternalLink(attrs []*ast.Node, linkAttrs []string, warnOnSpread bool, spreadIdx int) bool {
	linkIdx := findLastIndex(attrs, func(a *ast.Node) bool {
		return attrNameIsOneOf(a, linkAttrs)
	})
	if linkIdx != -1 {
		init := attrs[linkIdx].AsJsxAttribute().Initializer
		// Upstream guard: `attr.value.type === 'Literal' && typeof value ===
		// 'string' && regex.test(value)`. In tsgo this corresponds to a
		// StringLiteral directly under the attribute (not inside `{…}`).
		if s, ok := stringLiteralText(init); ok && isExternalURL(s) {
			return true
		}
	}
	return warnOnSpread && linkIdx < spreadIdx
}

func hasDynamicLink(attrs []*ast.Node, linkAttrs []string) bool {
	return findLastIndex(attrs, func(a *ast.Node) bool {
		if !attrNameIsOneOf(a, linkAttrs) {
			return false
		}
		init := a.AsJsxAttribute().Initializer
		return init != nil && init.Kind == ast.KindJsxExpression
	}) != -1
}

// relStrings extracts the candidate string values the rel attribute could
// evaluate to. Mirrors upstream's `getStringFromValue` including the
// "matched-test-condition" shortcut: when both `target` and `rel` are
// JsxExpressions wrapping a ConditionalExpression sharing the same test
// identifier, only the rel branch aligned with the `_blank` target branch is
// returned. Otherwise all literal branches of the conditional are returned.
//
// A returned `nil` entry represents a non-string branch — callers treat it as
// "this branch is not a secure rel".
func relStrings(relInit, targetInit *ast.Node) []*string {
	if relInit == nil {
		return nil
	}
	// Direct `rel="…"` form.
	if s, ok := templateOrStringText(relInit); ok {
		sCopy := s
		return []*string{&sCopy}
	}
	expr := jsxExpressionInner(relInit)
	if expr == nil {
		return nil
	}
	// `rel={"…"}` / `rel={\`…\`}` form.
	if s, ok := templateOrStringText(expr); ok {
		sCopy := s
		return []*string{&sCopy}
	}
	if expr.Kind == ast.KindConditionalExpression {
		cond := expr.AsConditionalExpression()
		consequent := stringOrNil(cond.WhenTrue)
		alternate := stringOrNil(cond.WhenFalse)
		// Matched-test shortcut: when both target and rel branch on the same
		// identifier, only the rel branch aligned with the `_blank` target
		// branch matters — the other side is unreachable when target !==
		// "_blank", so rel's value there can't make this usage insecure.
		if targetExpr := jsxExpressionInner(targetInit); targetExpr != nil && targetExpr.Kind == ast.KindConditionalExpression {
			targetCond := targetExpr.AsConditionalExpression()
			if relCondName := identifierName(cond.Condition); relCondName != "" && relCondName == identifierName(targetCond.Condition) {
				tConsequent, _ := stringLiteralText(targetCond.WhenTrue)
				tAlternate, _ := stringLiteralText(targetCond.WhenFalse)
				switch "_blank" {
				case tConsequent:
					return []*string{consequent}
				case tAlternate:
					return []*string{alternate}
				}
			}
		}
		return []*string{consequent, alternate}
	}
	// Non-literal rel expression (`rel={getRel()}`, etc.) — caller should
	// treat as non-secure.
	return []*string{nil}
}

// stringOrNil returns a pointer to the string value of a literal-like node
// (after skipping parentheses), or nil when the node carries a non-string
// value (number, boolean, null, identifier, …).
func stringOrNil(node *ast.Node) *string {
	if s, ok := templateOrStringText(node); ok {
		return &s
	}
	return nil
}

func identifierName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	n := ast.SkipParentheses(node)
	if n.Kind == ast.KindIdentifier {
		return n.AsIdentifier().Text
	}
	return ""
}

func hasSecureRel(attrs []*ast.Node, allowReferrer, warnOnSpread bool, spreadIdx int) bool {
	relIdx := findLastIndex(attrs, func(a *ast.Node) bool { return attrHasName(a, "rel") })
	targetIdx := findLastIndex(attrs, func(a *ast.Node) bool { return attrHasName(a, "target") })
	if relIdx == -1 || (warnOnSpread && relIdx < spreadIdx) {
		return false
	}
	relInit := attrs[relIdx].AsJsxAttribute().Initializer
	var targetInit *ast.Node
	if targetIdx != -1 {
		targetInit = attrs[targetIdx].AsJsxAttribute().Initializer
	}
	values := relStrings(relInit, targetInit)
	if len(values) == 0 {
		return false
	}
	for _, v := range values {
		if v == nil {
			return false
		}
		tags := strings.Split(strings.ToLower(*v), " ")
		if slices.Contains(tags, "noreferrer") {
			continue
		}
		if allowReferrer && slices.Contains(tags, "noopener") {
			continue
		}
		return false
	}
	return true
}

// buildRelFix returns the fix that adds / normalizes the rel attribute so the
// element becomes secure. Mirrors upstream's fixer logic. Returns nil when no
// fix should be emitted (target trailing a spread, or non-string / non-literal
// rel expression the fixer cannot rewrite safely).
func buildRelFix(sf *ast.SourceFile, attrs []*ast.Node, targetIdx, spreadIdx int, relValue string) *rule.RuleFix {
	relIdx := findLastIndex(attrs, func(a *ast.Node) bool { return attrHasName(a, "rel") })
	if targetIdx < spreadIdx || (spreadIdx >= 0 && relIdx == -1) {
		return nil
	}
	if relIdx == -1 {
		// Insert ` rel="…"` after the last attribute.
		lastAttr := attrs[len(attrs)-1]
		fix := rule.RuleFixInsertAfter(lastAttr, ` rel="`+relValue+`"`)
		return &fix
	}
	init := attrs[relIdx].AsJsxAttribute().Initializer
	if init == nil {
		// Boolean shorthand `rel` → insert `="…"` after the attribute.
		fix := rule.RuleFixInsertAfter(attrs[relIdx], `="`+relValue+`"`)
		return &fix
	}
	// `rel="…"` — split on the target token, re-join preserving others.
	if s, ok := stringLiteralText(init); ok {
		parts := splitNonEmpty(s, relValue)
		fix := rule.RuleFixReplace(sf, init, `"`+strings.Join(append(parts, relValue), " ")+`"`)
		return &fix
	}
	// `rel={…}` form.
	expr := jsxExpressionInner(init)
	if expr == nil {
		return nil
	}
	// `rel={"…"}` → rewrite just the inner string literal, preserving the
	// curly braces the user wrote. Mirrors upstream's `replaceText(value.
	// expression, …)`.
	if s, ok := stringLiteralText(expr); ok {
		parts := splitNonEmpty(s, relValue)
		fix := rule.RuleFixReplace(sf, expr, `"`+strings.Join(append(parts, relValue), " ")+`"`)
		return &fix
	}
	// `rel={true}` / `{null}` / `{3}` / `{false}` / `{undefined}` —
	// non-string primitive; upstream collapses the whole `{…}` down to a
	// plain string attribute. We only rewrite shapes we can reason about;
	// arbitrary expressions (identifier, call) are left alone, matching the
	// upstream "return null" branch.
	switch expr.Kind {
	case ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindTrueKeyword,
		ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindUndefinedKeyword:
		fix := rule.RuleFixReplace(sf, init, `"`+relValue+`"`)
		return &fix
	case ast.KindIdentifier:
		if expr.AsIdentifier().Text == "undefined" {
			fix := rule.RuleFixReplace(sf, init, `"`+relValue+`"`)
			return &fix
		}
	}
	return nil
}

// splitNonEmpty splits `s` by `sep` and drops empty segments — matching
// upstream's `value.split(sep).filter(Boolean)`. Used to sanitize rel values
// like "noopenernoreferrer" → ["noopener"] so the fix can re-append the token
// cleanly without duplicating it.
func splitNonEmpty(s, sep string) []string {
	raw := strings.Split(s, sep)
	out := raw[:0]
	for _, p := range raw {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func reportWithOptionalFix(ctx rule.RuleContext, node *ast.Node, messageId, description string, fix *rule.RuleFix) {
	msg := rule.RuleMessage{Id: messageId, Description: description}
	if fix != nil {
		ctx.ReportNodeWithFixes(node, msg, *fix)
		return
	}
	ctx.ReportNode(node, msg)
}

var JsxNoTargetBlankRule = rule.Rule{
	Name: "react/jsx-no-target-blank",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		linkComponents := readComponentsFromSettings(ctx.Settings, "linkComponents", "linkAttribute", "href", newDefaultLinkComponents())
		formComponents := readComponentsFromSettings(ctx.Settings, "formComponents", "formAttribute", "action", newDefaultFormComponents())

		messageId := "noTargetBlankWithoutNoreferrer"
		description := msgNoreferrer
		relValue := "noreferrer"
		if opts.allowReferrer {
			messageId = "noTargetBlankWithoutNoopener"
			description = msgNoopener
			relValue = "noopener"
		}

		check := func(node *ast.Node) {
			// Upstream keys on `node.name.name`, which is only a string for an
			// Identifier tag. PropertyAccessExpression (`<Foo.Bar>`),
			// JsxNamespacedName (`<ns:a>`) and ThisKeyword tags all hash to
			// non-string and fall through — same as our empty-name guard.
			tagName := reactutil.GetJsxTagName(node)
			if tagName == nil || tagName.Kind != ast.KindIdentifier {
				return
			}
			name := tagName.AsIdentifier().Text
			isLink := opts.links && linkComponents[name] != nil
			isForm := opts.forms && formComponents[name] != nil
			if !isLink && !isForm {
				return
			}
			attrs := reactutil.GetJsxElementAttributes(node)
			if len(attrs) == 0 {
				return
			}

			targetIdx := findLastIndex(attrs, func(a *ast.Node) bool { return attrHasName(a, "target") })
			spreadIdx := findLastIndex(attrs, func(a *ast.Node) bool { return a.Kind == ast.KindJsxSpreadAttribute })

			// shouldProceed encodes upstream's entry gate: check the element
			// when target could be "_blank", OR when a spread might inject
			// one and `warnOnSpreadAttributes` is enabled. Upstream's gate is
			// literally equivalent to this condition once the branches of its
			// nested if/else-if are unrolled.
			shouldProceed := func() bool {
				var targetAttr *ast.Node
				if targetIdx != -1 {
					targetAttr = attrs[targetIdx]
				}
				if attributeValuePossiblyBlank(targetAttr) {
					return true
				}
				return opts.warnOnSpreadAttributes && spreadIdx >= 0
			}

			if isLink {
				if shouldProceed() {
					componentAttrs := linkComponents[name]
					dangerous := hasExternalLink(attrs, componentAttrs, opts.warnOnSpreadAttributes, spreadIdx) ||
						(opts.enforceDynamicLinks == "always" && hasDynamicLink(attrs, componentAttrs))
					if dangerous && !hasSecureRel(attrs, opts.allowReferrer, opts.warnOnSpreadAttributes, spreadIdx) {
						reportWithOptionalFix(ctx, node, messageId, description,
							buildRelFix(ctx.SourceFile, attrs, targetIdx, spreadIdx, relValue))
						return
					}
				}
			}
			if isForm {
				// Form branch mirrors upstream: forms never autofix, and the
				// inner `hasSecureRel` / `hasExternalLink` calls are invoked
				// with undefined trailing arguments — equivalent to
				// `(attrs, false, false, -1)` / `(attrs, formAttrs, false, -1)`
				// on our side. Upstream also passes no `allowReferrer` there,
				// so forms with only `rel="noopener"` still report even when
				// the rule-level option is on.
				if !shouldProceed() {
					return
				}
				if hasSecureRel(attrs, false, false, -1) {
					return
				}
				formAttrs := formComponents[name]
				if hasExternalLink(attrs, formAttrs, false, -1) ||
					(opts.enforceDynamicLinks == "always" && hasDynamicLink(attrs, formAttrs)) {
					reportWithOptionalFix(ctx, node, messageId, description, nil)
				}
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}
