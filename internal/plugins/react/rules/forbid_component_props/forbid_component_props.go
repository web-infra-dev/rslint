package forbid_component_props

import (
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const msgPropIsForbidden = `Prop "{{prop}}" is forbidden on Components`

// forbidEntry mirrors a single value of upstream's `forbid` Map. The raw key
// is the literal `propName` (or `propNamePattern`) the user supplied — pattern
// entries are matched via minimatch against the JSX attribute name; non-pattern
// entries are looked up by exact string equality first.
type forbidEntry struct {
	key                 string
	isPattern           bool
	allowList           []string
	allowPatternList    []string
	disallowList        []string
	disallowPatternList []string
	message             string
}

type forbidConfig struct {
	byKey   map[string]*forbidEntry
	ordered []*forbidEntry
}

func parseOptions(options any) *forbidConfig {
	cfg := &forbidConfig{byKey: map[string]*forbidEntry{}}

	var forbidList []interface{}
	if optsMap := utils.GetOptionsMap(options); optsMap != nil {
		if raw, ok := optsMap["forbid"].([]interface{}); ok {
			forbidList = raw
		}
	}
	// Upstream falls back to DEFAULTS = ['className', 'style'] when `forbid`
	// is absent. An explicit empty array (`forbid: []`) keeps an empty list,
	// matching upstream behavior.
	if forbidList == nil {
		forbidList = []interface{}{"className", "style"}
	}

	for _, raw := range forbidList {
		switch v := raw.(type) {
		case string:
			if v == "" {
				continue
			}
			cfg.set(v, &forbidEntry{key: v})
		case map[string]interface{}:
			propName, _ := v["propName"].(string)
			propPattern, _ := v["propNamePattern"].(string)
			key := propName
			if key == "" {
				key = propPattern
			}
			if key == "" {
				continue
			}
			e := &forbidEntry{key: key, isPattern: propPattern != ""}
			if l, ok := v["allowedFor"].([]interface{}); ok {
				e.allowList = stringSlice(l)
			}
			if l, ok := v["allowedForPatterns"].([]interface{}); ok {
				e.allowPatternList = stringSlice(l)
			}
			if l, ok := v["disallowedFor"].([]interface{}); ok {
				e.disallowList = stringSlice(l)
			}
			if l, ok := v["disallowedForPatterns"].([]interface{}); ok {
				e.disallowPatternList = stringSlice(l)
			}
			if msg, ok := v["message"].(string); ok && msg != "" {
				e.message = msg
			}
			cfg.set(key, e)
		}
	}
	return cfg
}

// set mirrors `Map.prototype.set`: re-setting an existing key replaces the
// value but preserves the original insertion position. Pattern iteration in
// `getPropOptions` walks `ordered` so the order is observable.
//
// The replace path scans `ordered` linearly (O(n)). `forbid` configs are
// expected to be small (typically ≤10 entries in real configs), so a linear
// scan is preferred over carrying a parallel index map purely for replace.
func (c *forbidConfig) set(key string, e *forbidEntry) {
	if _, exists := c.byKey[key]; exists {
		for i, prev := range c.ordered {
			if prev.key == key {
				c.ordered[i] = e
				break
			}
		}
	} else {
		c.ordered = append(c.ordered, e)
	}
	c.byKey[key] = e
}

func stringSlice(raw []interface{}) []string {
	out := make([]string, 0, len(raw))
	for _, x := range raw {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// getPropOptions mirrors upstream `getPropOptions`: direct string-key lookup
// wins; otherwise the first pattern entry whose key matches the prop wins.
func (c *forbidConfig) getPropOptions(prop string) *forbidEntry {
	if e, ok := c.byKey[prop]; ok {
		return e
	}
	for _, e := range c.ordered {
		if !e.isPattern {
			continue
		}
		if reactutil.MatchGlob(prop, e.key) {
			return e
		}
	}
	return nil
}

// isForbidden mirrors upstream `isForbidden`. The disallow path takes
// precedence when at least one disallow list is configured; otherwise the
// allow path applies.
func (c *forbidConfig) isForbidden(prop, tag string) (forbidden bool, opts *forbidEntry) {
	opts = c.getPropOptions(prop)
	if opts == nil {
		return false, nil
	}

	hasDisallowOptions := len(opts.disallowList) > 0 || len(opts.disallowPatternList) > 0

	var tagForbidden bool
	if hasDisallowOptions {
		tagForbidden = slices.Contains(opts.disallowList, tag)
		if !tagForbidden {
			for _, p := range opts.disallowPatternList {
				if reactutil.MatchGlob(tag, p) {
					tagForbidden = true
					break
				}
			}
		}
	} else {
		if slices.Contains(opts.allowList, tag) {
			tagForbidden = false
		} else if len(opts.allowPatternList) == 0 {
			tagForbidden = true
		} else {
			tagForbidden = true
			for _, p := range opts.allowPatternList {
				if reactutil.MatchGlob(tag, p) {
					tagForbidden = false
					break
				}
			}
		}
	}

	// Upstream: `typeof tagName === 'undefined' || isTagForbidden`. tag == ""
	// in the Go port stands in for the upstream undefined case (an unknown /
	// unrepresentable tag shape) — treat it as forbidden.
	return tag == "" || tagForbidden, opts
}

// tagShape carries the data the rule needs out of a JSX tag-name node: the
// `tag` string (matched against allow/disallow lists), the `componentName`
// string (used for the DOM check), and `skipDomCheck` to mirror upstream's
// quirk that namespaced tags bypass the lowercase-rejection branch.
type tagShape struct {
	tag           string
	componentName string
	skipDomCheck  bool
}

// getTagShape mirrors upstream's tag/componentName extraction:
//
//	const tag = parentName.name || `${parentName.object.name}.${parentName.property.name}`;
//	const componentName = parentName.name || parentName.property.name;
//
// JSX tag-name grammar (tsgo `parseJsxElementName`, parser.go:4948) restricts
// the receiver to one of:
//   - `Identifier`            — `<Foo>`
//   - `ThisKeyword`           — `<this>` / `<this.X>`
//   - `JsxNamespacedName`     — `<ns:Name>`
//   - `PropertyAccessExpression` whose own `.Expression` is recursively one
//     of the above — `<Foo.Bar>`, `<Foo.Bar.Baz>`
//
// Expression wrappers like `(X)`, `X as T`, `X!`, `X satisfies T`, `<T>x` are
// REJECTED by the parser at JSX-tag-name position (they're parsed as ordinary
// expressions only outside of JSX tag names). Therefore no `SkipParentheses`
// / `SkipExpressionWrappers` pass is needed when walking `pa.Expression` — a
// wrapper kind is unreachable here.
//
// Two upstream behaviors are mirrored exactly so allow/disallow lists match
// byte-for-byte:
//
//  1. Member expressions deeper than two segments (`<Foo.Bar.Baz />`) produce
//     the literal string `"undefined.Baz"` because `parentName.object.name` is
//     undefined when `object` is itself a member expression.
//  2. JSX namespaced names (`<fbt:param />`) — upstream's `tag` becomes a
//     JSXIdentifier object (truthy but not a string), so allow/disallow
//     `indexOf` never matches and `componentName[0]` is undefined, which
//     bypasses the lowercase-rejection branch. We model this with a synthetic
//     tag string that allow/disallow lists can still match (more useful for
//     real configs) and `skipDomCheck = true` so namespaced lowercase tags
//     are NOT silently dropped as DOM intrinsics.
func getTagShape(tagName *ast.Node) (tagShape, bool) {
	if tagName == nil {
		return tagShape{}, false
	}
	switch tagName.Kind {
	case ast.KindIdentifier:
		id := tagName.AsIdentifier()
		if id == nil {
			return tagShape{}, false
		}
		text := id.Text
		return tagShape{tag: text, componentName: text}, true
	case ast.KindThisKeyword:
		// Bare `<this>` is not legal JSX; defensive branch only.
		return tagShape{tag: "this", componentName: "this"}, true
	case ast.KindPropertyAccessExpression:
		pa := tagName.AsPropertyAccessExpression()
		if pa == nil || pa.Expression == nil {
			return tagShape{}, false
		}
		propName := pa.Name()
		if propName == nil || propName.Kind != ast.KindIdentifier {
			return tagShape{}, false
		}
		propText := propName.AsIdentifier().Text
		var baseText string
		switch pa.Expression.Kind {
		case ast.KindIdentifier:
			baseText = pa.Expression.AsIdentifier().Text
		case ast.KindThisKeyword:
			baseText = "this"
		default:
			// Per the grammar note above, the only remaining shape here is
			// `pa.Expression` being itself a `PropertyAccessExpression` —
			// i.e. chains deeper than two segments (`<Foo.Bar.Baz />`).
			// Upstream produces `"undefined.<prop>"` for that case because
			// `parentName.object.name` is undefined when `object` is itself
			// a member expression; mirror it byte-for-byte.
			baseText = "undefined"
		}
		return tagShape{tag: baseText + "." + propText, componentName: propText}, true
	case ast.KindJsxNamespacedName:
		ns := tagName.AsJsxNamespacedName()
		if ns == nil || ns.Namespace == nil || ns.Name() == nil {
			return tagShape{}, false
		}
		if ns.Namespace.Kind != ast.KindIdentifier || ns.Name().Kind != ast.KindIdentifier {
			return tagShape{}, false
		}
		nsText := ns.Namespace.AsIdentifier().Text
		nameText := ns.Name().AsIdentifier().Text
		return tagShape{
			tag:           nsText + ":" + nameText,
			componentName: nameText,
			// Mirror upstream: namespaced names bypass the DOM lowercase-skip
			// path entirely, so `<fbt:param>` is treated as a Component.
			skipDomCheck: true,
		}, true
	}
	return tagShape{}, false
}

// isLowercaseStart mirrors upstream's
// `componentName[0] !== componentName[0].toUpperCase()` test: returns true iff
// the first rune is a cased letter in its lowercase form. Digits, `_`, `$`,
// and uppercase letters all return false (the rule then continues, matching
// upstream).
func isLowercaseStart(s string) bool {
	if s == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.ToLower(r) == r && unicode.ToUpper(r) != r
}

var ForbidComponentPropsRule = rule.Rule{
	Name: "react/forbid-component-props",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		cfg := parseOptions(options)
		return rule.RuleListeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				parent := reactutil.GetJsxParentElement(node)
				if parent == nil {
					return
				}
				tagName := reactutil.GetJsxTagName(parent)
				if tagName == nil {
					return
				}
				shape, ok := getTagShape(tagName)
				if !ok {
					return
				}
				if !shape.skipDomCheck && isLowercaseStart(shape.componentName) {
					return
				}
				prop := reactutil.GetJsxPropName(node)
				if prop == "" {
					return
				}
				forbidden, opts := cfg.isForbidden(prop, shape.tag)
				if !forbidden {
					return
				}
				if opts != nil && opts.message != "" {
					ctx.ReportNode(node, rule.RuleMessage{
						Description: opts.message,
					})
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "propIsForbidden",
					Description: strings.ReplaceAll(msgPropIsForbidden, "{{prop}}", prop),
					Data:        map[string]string{"prop": prop},
				})
			},
		}
	},
}
