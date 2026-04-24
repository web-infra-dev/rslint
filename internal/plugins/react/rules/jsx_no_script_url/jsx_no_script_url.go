package jsx_no_script_url

import (
	"regexp"
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// isJavaScriptProtocol matches the `javascript:` scheme with optional control
// characters (\u0000-\u001F) and spaces before the scheme, and optional \r \n
// \t characters between the letters. Case-insensitive. Mirrors React's
// sanitizeURL check:
// https://github.com/facebook/react/blob/d0ebde77f6d1232cefc0da184d731943d78e86f2/packages/react-dom/src/shared/sanitizeURL.js#L30
var isJavaScriptProtocol = regexp.MustCompile(
	`(?i)^[\x00-\x1f ]*j[\r\n\t]*a[\r\n\t]*v[\r\n\t]*a[\r\n\t]*s[\r\n\t]*c[\r\n\t]*r[\r\n\t]*i[\r\n\t]*p[\r\n\t]*t[\r\n\t]*:`)

const noScriptURLMessage = "A future version of React will block javascript: URLs as a security precaution. " +
	"Use event handlers instead if you can. If you need to generate unsafe HTML, try using dangerouslySetInnerHTML instead."

var JsxNoScriptUrlRule = rule.Rule{
	Name: "react/jsx-no-script-url",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		config := parseConfig(options, ctx.Settings)

		return rule.RuleListeners{
			ast.KindJsxAttribute: func(node *ast.Node) {
				if !shouldVerifyProp(node, config) {
					return
				}
				if !hasJavaScriptProtocol(node) {
					return
				}
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "noScriptURL",
					Description: noScriptURLMessage,
				})
			},
		}
	},
}

// hasJavaScriptProtocol reports whether the attribute value is a string literal
// that starts with the `javascript:` protocol. Only direct string literals are
// checked — expression containers like `href={"javascript:"}` are not flagged,
// matching upstream behavior.
func hasJavaScriptProtocol(attr *ast.Node) bool {
	init := attr.AsJsxAttribute().Initializer
	if init == nil {
		return false
	}
	if init.Kind != ast.KindStringLiteral {
		return false
	}
	return isJavaScriptProtocol.MatchString(init.AsStringLiteral().Text)
}

// shouldVerifyProp reports whether a JsxAttribute should be checked for
// javascript: URLs based on the component/prop configuration. Upstream uses
// `node.parent.name.name` which resolves to:
//   - Identifier tag: the element name (e.g. "a", "Link")
//   - JsxNamespacedName tag: the local part (e.g. "a" in "ns:a")
//   - MemberExpression tag: undefined (not matched)
func shouldVerifyProp(node *ast.Node, config reactutil.ComponentMap) bool {
	attrName := reactutil.GetJsxPropName(node)
	if attrName == "" {
		return false
	}
	parent := reactutil.GetJsxParentElement(node)
	if parent == nil {
		return false
	}
	tagName := reactutil.GetJsxTagName(parent)
	if tagName == nil {
		return false
	}
	var parentName string
	switch tagName.Kind {
	case ast.KindIdentifier:
		parentName = tagName.AsIdentifier().Text
	case ast.KindJsxNamespacedName:
		// Upstream: node.parent.name.name on a JSXNamespacedName returns
		// the local part (e.g., "a" for <ns:a>).
		ns := tagName.AsJsxNamespacedName()
		nameNode := ns.Name()
		if nameNode != nil && nameNode.Kind == ast.KindIdentifier {
			parentName = nameNode.AsIdentifier().Text
		}
	default:
		return false
	}
	if parentName == "" {
		return false
	}
	props, ok := config[parentName]
	if !ok {
		return false
	}
	return slices.Contains(props, attrName)
}

// parseConfig builds the component→props map from rule options and settings.
//
// Options format (after config.go unwrapping):
//   - nil → default {"a": ["href"]}
//   - map{"includeFromSettings": true} → merge settings
//   - [{name, props}, ...] → legacy custom components
//   - [[{name, props}, ...], {includeFromSettings: true}] → both
func parseConfig(options any, settings map[string]interface{}) reactutil.ComponentMap {
	var legacyOptions []interface{}
	includeFromSettings := false

	switch opts := options.(type) {
	case nil:
		// no options
	case map[string]interface{}:
		// Object option only: {includeFromSettings: true}
		if v, ok := opts["includeFromSettings"].(bool); ok {
			includeFromSettings = v
		}
	case []interface{}:
		if len(opts) > 0 {
			if inner, ok := opts[0].([]interface{}); ok {
				// Shape: [[{name, props}, ...], {includeFromSettings}]
				legacyOptions = inner
				if len(opts) > 1 {
					if objOpt, ok := opts[1].(map[string]interface{}); ok {
						if v, ok := objOpt["includeFromSettings"].(bool); ok {
							includeFromSettings = v
						}
					}
				}
			} else {
				// Shape: [{name, props}, ...] — config.go unwrapped single-element
				legacyOptions = opts
			}
		}
	}

	// Start with defaults, optionally merging settings.
	var config reactutil.ComponentMap
	if includeFromSettings {
		config = reactutil.ReadComponentsFromSettings(
			settings, "linkComponents", "linkAttribute", "href",
			reactutil.DefaultLinkComponents())
	} else {
		config = reactutil.DefaultLinkComponents()
	}

	// Merge legacy option components. Upstream uses config.set(name, props)
	// which REPLACES the entry — if a legacy option redefines "a", the
	// default ["href"] is dropped entirely.
	for _, opt := range legacyOptions {
		item, ok := opt.(map[string]interface{})
		if !ok {
			continue
		}
		name, _ := item["name"].(string)
		if name == "" {
			continue
		}
		propsRaw, ok := item["props"].([]interface{})
		if !ok {
			continue
		}
		var props []string
		for _, p := range propsRaw {
			if s, ok := p.(string); ok {
				props = append(props, s)
			}
		}
		config[name] = props
	}

	return config
}
