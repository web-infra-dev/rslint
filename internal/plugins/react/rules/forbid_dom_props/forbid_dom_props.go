package forbid_dom_props

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	msgPropIsForbidden          = `Prop "{{prop}}" is forbidden on DOM Nodes`
	msgPropIsForbiddenWithValue = `Prop "{{prop}}" with value "{{propValue}}" is forbidden on DOM Nodes`
)

// forbidEntry mirrors a single value of upstream's `forbid` Map. `propName` is
// the literal attribute name we're matching (no glob support — unlike
// forbid-component-props, this rule has no `propNamePattern`).
//
// `disallowList` / `disallowedValues` are nil when the user did not supply the
// option (string entries in `forbid` produce nil lists). A non-nil empty slice
// is preserved separately and means "user supplied an explicit empty array",
// which mirrors upstream's truthy-check on the array reference and disables
// matching for that conjunct entirely.
type forbidEntry struct {
	propName            string
	hasDisallowList     bool
	disallowList        []string
	hasDisallowedValues bool
	disallowedValues    []string
	message             string
}

type forbidConfig struct {
	byProp map[string]*forbidEntry
}

func parseOptions(options any) *forbidConfig {
	cfg := &forbidConfig{byProp: map[string]*forbidEntry{}}

	var forbidList []interface{}
	if optsMap := utils.GetOptionsMap(options); optsMap != nil {
		if raw, ok := optsMap["forbid"].([]interface{}); ok {
			forbidList = raw
		}
	}
	// Upstream defaults to `DEFAULTS = []` — empty list. With `forbid: []`
	// (or no options at all) the rule is effectively a no-op; both paths land
	// here with `forbidList == nil` and we return an empty config.
	if forbidList == nil {
		return cfg
	}

	for _, raw := range forbidList {
		switch v := raw.(type) {
		case string:
			if v == "" {
				continue
			}
			cfg.byProp[v] = &forbidEntry{propName: v}
		case map[string]interface{}:
			propName, _ := v["propName"].(string)
			if propName == "" {
				// Upstream creates an entry keyed by `undefined` here; in
				// practice the listener never emits the prop name `undefined`,
				// so the entry is unreachable. Skip silently.
				continue
			}
			e := &forbidEntry{propName: propName}
			if l, ok := v["disallowedFor"].([]interface{}); ok {
				e.disallowList = stringSlice(l)
				e.hasDisallowList = true
			}
			if l, ok := v["disallowedValues"].([]interface{}); ok {
				e.disallowedValues = stringSlice(l)
				e.hasDisallowedValues = true
			}
			if msg, ok := v["message"].(string); ok && msg != "" {
				e.message = msg
			}
			cfg.byProp[propName] = e
		}
	}
	return cfg
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

// isForbidden mirrors upstream `isForbidden`:
//
//	(!options.disallowList || options.disallowList.indexOf(tagName) !== -1)
//	  && (!options.disallowedValues || options.disallowedValues.indexOf(propValue) !== -1)
//
// In Go terms: each conjunct is "no list configured OR list contains the
// tag/value". An explicit empty `disallowedFor: []` keeps `hasDisallowList`
// true with an empty slice — `slices.Contains` returns false → the whole
// conjunct fails → not forbidden, matching upstream's truthy-on-array-ref
// check.
//
// `hasValue` distinguishes string-literal initializers (where `propValue` is
// the cooked string) from missing or expression initializers (where upstream
// reads `node.value.value` as `undefined`). When `hasValue` is false and
// `disallowedValues` is configured, the rule does not fire — `[].indexOf(undef)`
// in upstream is `-1`.
func (c *forbidConfig) isForbidden(prop, tag, propValue string, hasValue bool) (bool, *forbidEntry) {
	e, ok := c.byProp[prop]
	if !ok {
		return false, nil
	}
	if e.hasDisallowList && !slices.Contains(e.disallowList, tag) {
		return false, nil
	}
	if e.hasDisallowedValues {
		if !hasValue {
			return false, nil
		}
		if !slices.Contains(e.disallowedValues, propValue) {
			return false, nil
		}
	}
	return true, e
}

// stringLiteralValue extracts the cooked-string value from a JsxAttribute
// initializer, matching upstream's `node.value.value` access. Returns
// (text, true) only for `KindStringLiteral` initializers; expression
// containers and missing initializers (boolean shorthand `<div hidden />`)
// produce (empty, false).
//
// Note: upstream throws a TypeError on boolean shorthand because
// `node.value` is null and `null.value` is invalid; we degrade gracefully.
// This is benign — without `disallowedValues` configured the rule still fires
// (matching the propName-only path); with `disallowedValues` configured the
// upstream behavior was a crash, so any defined output is an improvement.
func stringLiteralValue(attr *ast.Node) (string, bool) {
	if attr == nil || attr.Kind != ast.KindJsxAttribute {
		return "", false
	}
	init := attr.AsJsxAttribute().Initializer
	if init == nil {
		return "", false
	}
	if init.Kind == ast.KindStringLiteral {
		return init.AsStringLiteral().Text, true
	}
	return "", false
}

var ForbidDomPropsRule = rule.Rule{
	Name: "react/forbid-dom-props",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		cfg := parseOptions(options)
		// Empty config short-circuits the whole listener — matches the
		// no-options / `forbid: []` path where upstream's Map iteration
		// finds nothing.
		if len(cfg.byProp) == 0 {
			return rule.RuleListeners{}
		}
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
				// Upstream:
				//   const tag = node.parent.name.name;
				//   if (!(tag && typeof tag === 'string'
				//         && tag[0] !== tag[0].toUpperCase())) return;
				// `node.parent.name.name` is the inner string only when the
				// tag-name is a plain JSXIdentifier. Member expressions
				// (`<Foo.Bar>`, `<this.x>`) yield an undefined name; namespaced
				// names (`<fbt:param>`) yield a JSXIdentifier object (truthy,
				// not a string). Both fail the `typeof === 'string'` test and
				// are skipped by upstream — mirror that here.
				if tagName.Kind != ast.KindIdentifier {
					return
				}
				tag := tagName.AsIdentifier().Text
				if !reactutil.IsCasedLowercaseFirstLetter(tag) {
					return
				}
				// Upstream: `const prop = node.name.name`. For a
				// `JSXNamespacedName` attribute name (e.g. `xlink:href`),
				// `node.name.name` is the inner JSXIdentifier object — not a
				// string — so `forbid.get(<object>)` never matches and the
				// listener silently exits. Mirror that by rejecting any
				// non-Identifier attribute name here. (`reactutil.GetJsxPropName`
				// would otherwise stringify it as `"xlink:href"` and produce a
				// false positive when a user configures `forbid: ['xlink:href']`.)
				nameNode := node.AsJsxAttribute().Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				prop := nameNode.AsIdentifier().Text
				if prop == "" {
					return
				}
				// Upstream reads `node.value.value` unconditionally — boolean
				// shorthand (`<div hidden />`) has `node.value === null`, so
				// `null.value` throws TypeError and the rule produces no
				// diagnostic for that JsxAttribute. To match the observable
				// "no diagnostic" outcome instead of recovering with a
				// best-effort report, exit here when the initializer is
				// absent. Net effect: 1:1 with upstream (no diagnostic) on
				// boolean shorthand, without the crash.
				attr := node.AsJsxAttribute()
				if attr.Initializer == nil {
					return
				}
				propValue, hasValue := stringLiteralValue(node)
				forbidden, e := cfg.isForbidden(prop, tag, propValue, hasValue)
				if !forbidden {
					return
				}
				// Upstream `report(context, message, messageId, { data })`
				// runs ESLint's `{{prop}}` / `{{propValue}}` interpolation
				// only on the messages-table path (resolved via messageId).
				// When `customMessage` is supplied directly as the `message`
				// arg, ESLint emits it verbatim — placeholders are NOT
				// substituted. Mirror exactly: custom message stays literal;
				// only the `propIsForbidden` / `propIsForbiddenWithValue`
				// templates interpolate.
				if e.message != "" {
					ctx.ReportNode(node, rule.RuleMessage{
						Description: e.message,
						Data:        map[string]string{"prop": prop, "propValue": propValue},
					})
					return
				}
				if e.hasDisallowedValues {
					desc := strings.ReplaceAll(msgPropIsForbiddenWithValue, "{{prop}}", prop)
					desc = strings.ReplaceAll(desc, "{{propValue}}", propValue)
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "propIsForbiddenWithValue",
						Description: desc,
						Data:        map[string]string{"prop": prop, "propValue": propValue},
					})
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "propIsForbidden",
					Description: strings.ReplaceAll(msgPropIsForbidden, "{{prop}}", prop),
					Data:        map[string]string{"prop": prop, "propValue": propValue},
				})
			},
		}
	},
}
