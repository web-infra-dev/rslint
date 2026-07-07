// cspell:ignore mdash
package jsx_no_literals

import (
	"html"
	"math"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// reOverridableElement mirrors upstream's `^[A-Z][\w.]*$` — overrides apply
// only to user-component names (capitalized identifiers, optionally
// dot-chained like `Foo.Bar`). The regex intentionally rejects DOM tag names
// like `div` so a `div` override silently no-ops, matching upstream's
// "HTML element tag names are not supported" docs note. Tests still rely on
// the validation gate skipping `div`-cased overrides without crashing.
var reOverridableElement = regexp.MustCompile(`^[A-Z][\w.]*$`)

const (
	configTypeElement  = "element"
	configTypeOverride = "override"
)

const (
	msgInvalidPropValue                   = "invalidPropValue"
	msgInvalidPropValueInElement          = "invalidPropValueInElement"
	msgNoStringsInAttributes              = "noStringsInAttributes"
	msgNoStringsInAttributesInElement     = "noStringsInAttributesInElement"
	msgNoStringsInJSX                     = "noStringsInJSX"
	msgNoStringsInJSXInElement            = "noStringsInJSXInElement"
	msgLiteralNotInJSXExpression          = "literalNotInJSXExpression"
	msgLiteralNotInJSXExpressionInElement = "literalNotInJSXExpressionInElement"
	msgRestrictedAttributeString          = "restrictedAttributeString"
	msgRestrictedAttributeStringInElement = "restrictedAttributeStringInElement"
)

type elementConfig struct {
	configType            string
	name                  string
	noStrings             bool
	allowedStrings        map[string]struct{}
	ignoreProps           bool
	noAttributeStrings    bool
	restrictedAttributes  map[string]struct{}
	allowElement          bool
	applyToNestedElements bool
}

type ruleConfig struct {
	base             elementConfig
	elementOverrides map[string]*elementConfig
}

func (c *ruleConfig) hasElementOverrides() bool {
	return len(c.elementOverrides) > 0
}

func parseConfig(options any) ruleConfig {
	cfg := ruleConfig{
		base: elementConfig{
			configType:           configTypeElement,
			allowedStrings:       map[string]struct{}{},
			restrictedAttributes: map[string]struct{}{},
		},
		elementOverrides: map[string]*elementConfig{},
	}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return cfg
	}
	populateElementConfig(&cfg.base, optsMap)
	rawOverrides, _ := optsMap["elementOverrides"].(map[string]interface{})
	for elementName, raw := range rawOverrides {
		if !reOverridableElement.MatchString(elementName) {
			continue
		}
		childMap, _ := raw.(map[string]interface{})
		if childMap == nil {
			childMap = map[string]interface{}{}
		}
		child := &elementConfig{
			configType:            configTypeOverride,
			name:                  elementName,
			allowedStrings:        map[string]struct{}{},
			restrictedAttributes:  map[string]struct{}{},
			applyToNestedElements: true,
		}
		populateElementConfig(child, childMap)
		if v, ok := childMap["allowElement"].(bool); ok {
			child.allowElement = v
		}
		// applyToNestedElements defaults to true; only an explicit falsy
		// value flips it. Upstream's `typeof v === 'undefined' || !!v`
		// collapses to: present and JS-falsy → false; otherwise → true.
		if v, ok := childMap["applyToNestedElements"]; ok {
			child.applyToNestedElements = truthyBool(v)
		}
		cfg.elementOverrides[elementName] = child
	}
	return cfg
}

func populateElementConfig(c *elementConfig, m map[string]interface{}) {
	if v, ok := m["noStrings"].(bool); ok {
		c.noStrings = v
	}
	if v, ok := m["ignoreProps"].(bool); ok {
		c.ignoreProps = v
	}
	if v, ok := m["noAttributeStrings"].(bool); ok {
		c.noAttributeStrings = v
	}
	populateStringSetFromMap(m, "allowedStrings", c.allowedStrings)
	populateStringSetFromMap(m, "restrictedAttributes", c.restrictedAttributes)
}

// populateStringSetFromMap reads the array-of-strings option at `key` from
// `source`, trims each entry, and inserts it into `target`. Mirrors upstream's
// `new Set(map(iterFrom(config.<key>), trimIfString))` shape.
func populateStringSetFromMap(source map[string]interface{}, key string, target map[string]struct{}) {
	arr, ok := source[key].([]interface{})
	if !ok {
		return
	}
	for _, s := range arr {
		if str, ok := s.(string); ok {
			target[strings.TrimSpace(str)] = struct{}{}
		}
	}
}

// truthyBool emulates JS `!!v` for option values. JSON-decoded numbers
// arrive as float64 (with `0` and `NaN` falsy); strings are truthy unless
// empty; objects / arrays are always truthy. Used for `applyToNestedElements`
// where upstream writes `typeof v === 'undefined' || !!v`.
func truthyBool(v interface{}) bool {
	switch x := v.(type) {
	case bool:
		return x
	case string:
		return x != ""
	case nil:
		return false
	case float64:
		return x != 0 && !math.IsNaN(x)
	case float32:
		return x != 0 && !math.IsNaN(float64(x))
	case int:
		return x != 0
	case int32:
		return x != 0
	case int64:
		return x != 0
	default:
		return true
	}
}

// skipBinaryAndParens walks up while the parent is a `BinaryExpression` or a
// `ParenthesizedExpression`. Upstream skips only `BinaryExpression` because
// ESTree drops parens at parse time; tsgo keeps parens as explicit nodes, so
// we additionally peel them to match the same observable behavior on inputs
// like `<div>{('foo')}</div>` and `<div>{('foo' + 'bar')}</div>`.
func skipBinaryAndParens(node *ast.Node) *ast.Node {
	for node != nil && node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindBinaryExpression, ast.KindParenthesizedExpression:
			node = node.Parent
			continue
		}
		return node
	}
	return node
}

// parentIgnoringBinaryAndParens returns the first ancestor of `node` that is
// NOT a `BinaryExpression` / `ParenthesizedExpression`. Mirrors upstream's
// `getParentIgnoringBinaryExpressions`, but also peels parens (see
// skipBinaryAndParens).
func parentIgnoringBinaryAndParens(node *ast.Node) *ast.Node {
	current := skipBinaryAndParens(node)
	if current == nil {
		return nil
	}
	return current.Parent
}

// jsxContentParentKind reports whether `kind` is a JSX content container
// (JsxElement / JsxFragment). JsxSelfClosingElement is excluded because it
// has no children, so a literal can never appear inside one.
func jsxContentParentKind(kind ast.Kind) bool {
	return kind == ast.KindJsxElement || kind == ast.KindJsxFragment
}

// hasJSXContentParentOrGrandParent mirrors upstream's
// `hasJSXElementParentOrGrandParent` — true when the literal lives directly
// in JSX content (not in a JSX attribute).
func hasJSXContentParentOrGrandParent(node *ast.Node) bool {
	parent := parentIgnoringBinaryAndParens(node)
	if parent == nil {
		return false
	}
	if jsxContentParentKind(parent.Kind) {
		return true
	}
	grand := parent.Parent
	// Peel an extra ParenthesizedExpression layer between parent and the JSX
	// container — `<div>{('foo')}</div>` parses as
	// StringLiteral > ParenthesizedExpression > JsxExpression > JsxElement
	// in tsgo, so grandParent reads as JsxExpression and the JsxElement is
	// one level higher. Walking up here keeps parity with the no-parens
	// ESTree shape upstream's logic targets.
	for grand != nil && grand.Kind == ast.KindParenthesizedExpression {
		grand = grand.Parent
	}
	return grand != nil && jsxContentParentKind(grand.Kind)
}

// isStandardJSXNode is the inner branch of upstream's `isViableTextNode` that
// classifies whether the parent shape qualifies as a JSX literal site.
// `text` is the cooked literal value used to check whitespace-only short-circuit.
func isStandardJSXNode(text string, parentKind ast.Kind, cfg *elementConfig) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	switch parentKind {
	case ast.KindJsxAttribute, ast.KindJsxElement, ast.KindJsxExpression, ast.KindJsxFragment:
	default:
		return false
	}
	if cfg.noAttributeStrings {
		return parentKind == ast.KindJsxAttribute || parentKind == ast.KindJsxElement
	}
	return parentKind != ast.KindJsxAttribute
}

func isViableTextNode(rawText, cookedText string, parentKind ast.Kind, cfg *elementConfig) bool {
	rawTrim := strings.TrimSpace(rawText)
	cookedTrim := strings.TrimSpace(cookedText)
	if _, ok := cfg.allowedStrings[rawTrim]; ok {
		return false
	}
	if _, ok := cfg.allowedStrings[cookedTrim]; ok {
		return false
	}
	standard := isStandardJSXNode(cookedText, parentKind, cfg)
	if cfg.noStrings {
		return standard
	}
	return standard && parentKind != ast.KindJsxExpression
}

// elementNameInfo carries the simple + dotted display name of a JSX element's
// tag, with renamed-import resolution already applied to the leftmost segment.
type elementNameInfo struct {
	name         string
	compoundName string
}

// jsxElementName mirrors upstream's `getJSXElementName`. Peels paired
// JsxElement → its OpeningElement so the same code-path handles both
// `<Foo>...</Foo>` and `<Foo />`, then delegates to
// `reactutil.GetJsxElementTypeString` for the dotted source string and
// applies renamed-import resolution to the root segment. For a member-chain
// tag (`Foo.Bar.Baz`), returns the last segment as `name` and the dotted
// form as `compoundName` (after rename). Namespaced (`<a:b>`) and `<this.X>`
// shapes pass through and are filtered later by the upstream-validated
// `reOverridableElement` regex.
func jsxElementName(element *ast.Node, renamedImports map[string]string) (elementNameInfo, bool) {
	target := element
	if target != nil && target.Kind == ast.KindJsxElement {
		target = target.AsJsxElement().OpeningElement
	}
	typeStr := reactutil.GetJsxElementTypeString(target)
	if typeStr == "" {
		return elementNameInfo{}, false
	}
	parts := strings.Split(typeStr, ".")
	if mapped, ok := renamedImports[parts[0]]; ok {
		parts[0] = mapped
	}
	info := elementNameInfo{name: parts[len(parts)-1]}
	if len(parts) > 1 {
		info.compoundName = strings.Join(parts, ".")
	}
	return info, true
}

// jsxElementAncestors returns all enclosing JsxElement ancestors, innermost
// first. JsxFragment ancestors are skipped — fragments have no name and
// cannot match an override.
func jsxElementAncestors(node *ast.Node) []*ast.Node {
	var out []*ast.Node
	for cur := node; cur != nil; cur = cur.Parent {
		if cur.Kind == ast.KindJsxElement || cur.Kind == ast.KindJsxSelfClosingElement {
			out = append(out, cur)
		}
	}
	return out
}

func resolveOverride(node *ast.Node, cfg *ruleConfig, renamedImports map[string]string) *elementConfig {
	if !cfg.hasElementOverrides() {
		return nil
	}
	ancestors := jsxElementAncestors(node)
	if len(ancestors) == 0 {
		return nil
	}
	for i, ancestor := range ancestors {
		isClosest := i == 0
		info, ok := jsxElementName(ancestor, renamedImports)
		if !ok || info.name == "" {
			continue
		}
		base := cfg.elementOverrides[info.name]
		var candidate *elementConfig
		if info.compoundName != "" {
			if c, ok := cfg.elementOverrides[info.compoundName]; ok {
				candidate = c
			} else {
				candidate = base
			}
		} else {
			candidate = base
		}
		if candidate == nil {
			continue
		}
		if isClosest || candidate.applyToNestedElements {
			return candidate
		}
	}
	return nil
}

func shouldAllowElement(cfg *elementConfig) bool {
	return cfg.configType == configTypeOverride && cfg.allowElement
}

// defaultMessageId mirrors upstream's `defaultMessageId(ancestorIsJSXElement, resolvedConfig)`.
// `ancestorIsJSXElement` follows upstream's parameter name: true when the
// literal lives inside JSX content (vs. inside a JSX attribute).
func defaultMessageId(ancestorIsJSXElement bool, cfg *elementConfig) string {
	if cfg.noAttributeStrings && !ancestorIsJSXElement {
		if cfg.configType == configTypeOverride {
			return msgNoStringsInAttributesInElement
		}
		return msgNoStringsInAttributes
	}
	if cfg.noStrings {
		if cfg.configType == configTypeOverride {
			return msgNoStringsInJSXInElement
		}
		return msgNoStringsInJSX
	}
	if cfg.configType == configTypeOverride {
		return msgLiteralNotInJSXExpressionInElement
	}
	return msgLiteralNotInJSXExpression
}

func formatMessage(messageId string, text, attribute, element string) string {
	switch messageId {
	case msgInvalidPropValue:
		return "Invalid prop value: \"" + text + "\""
	case msgInvalidPropValueInElement:
		return "Invalid prop value: \"" + text + "\" in " + element
	case msgNoStringsInAttributes:
		return "Strings not allowed in attributes: \"" + text + "\""
	case msgNoStringsInAttributesInElement:
		return "Strings not allowed in attributes: \"" + text + "\" in " + element
	case msgNoStringsInJSX:
		return "Strings not allowed in JSX files: \"" + text + "\""
	case msgNoStringsInJSXInElement:
		return "Strings not allowed in JSX files: \"" + text + "\" in " + element
	case msgLiteralNotInJSXExpression:
		return "Missing JSX expression container around literal string: \"" + text + "\""
	case msgLiteralNotInJSXExpressionInElement:
		return "Missing JSX expression container around literal string: \"" + text + "\" in " + element
	case msgRestrictedAttributeString:
		return "Restricted attribute string: \"" + text + "\" in " + attribute
	case msgRestrictedAttributeStringInElement:
		return "Restricted attribute string: \"" + text + "\" in " + attribute + " of " + element
	}
	return ""
}

func reportLiteralNode(ctx rule.RuleContext, node *ast.Node, messageId string, cfg *elementConfig, text string) {
	element := ""
	if cfg.configType == configTypeOverride {
		element = cfg.name
	}
	desc := formatMessage(messageId, text, "", element)
	data := map[string]string{"text": text}
	if element != "" {
		data["element"] = element
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          messageId,
		Description: desc,
		Data:        data,
	})
}

func reportRestrictedAttribute(ctx rule.RuleContext, attrNode *ast.Node, valueText, attribute string, cfg *elementConfig) {
	messageId := msgRestrictedAttributeString
	element := ""
	if cfg.configType == configTypeOverride {
		messageId = msgRestrictedAttributeStringInElement
		element = cfg.name
	}
	desc := formatMessage(messageId, valueText, attribute, element)
	data := map[string]string{"text": valueText, "attribute": attribute}
	if element != "" {
		data["element"] = element
	}
	ctx.ReportNode(attrNode, rule.RuleMessage{
		Id:          messageId,
		Description: desc,
		Data:        data,
	})
}

// isRequireCall mirrors upstream's `isRequireStatement` — peels nested
// PropertyAccess to find a `require(...)` CallExpression. Used to gate the
// `const { T: U } = require(...)` renamed-binding harvest. Skips
// ParenthesizedExpression at every read site (tsgo retains them as nodes;
// ESTree drops at parse time) so `(require('foo')).Foo` is detected just
// like the unparenthesized form upstream sees.
func isRequireCall(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindCallExpression:
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		if callee != nil && callee.Kind == ast.KindIdentifier {
			return callee.AsIdentifier().Text == "require"
		}
		return false
	case ast.KindPropertyAccessExpression:
		return isRequireCall(node.AsPropertyAccessExpression().Expression)
	}
	return false
}

func collectRenamedImports(sf *ast.SourceFile) map[string]string {
	out := map[string]string{}
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil {
			return
		}
		switch n.Kind {
		case ast.KindImportDeclaration:
			id := n.AsImportDeclaration()
			if id.ImportClause != nil {
				ic := id.ImportClause.AsImportClause()
				if ic.NamedBindings != nil && ic.NamedBindings.Kind == ast.KindNamedImports {
					ni := ic.NamedBindings.AsNamedImports()
					if ni.Elements != nil {
						for _, spec := range ni.Elements.Nodes {
							s := spec.AsImportSpecifier()
							local := s.Name()
							if local == nil || local.Kind != ast.KindIdentifier {
								continue
							}
							imported := local.AsIdentifier().Text
							if s.PropertyName != nil && s.PropertyName.Kind == ast.KindIdentifier {
								imported = s.PropertyName.AsIdentifier().Text
							}
							out[local.AsIdentifier().Text] = imported
						}
					}
				}
			}
		case ast.KindVariableStatement:
			vs := n.AsVariableStatement()
			if vs.DeclarationList == nil {
				return
			}
			decls := vs.DeclarationList.AsVariableDeclarationList().Declarations
			if decls == nil {
				return
			}
			for _, d := range decls.Nodes {
				vd := d.AsVariableDeclaration()
				if vd.Initializer == nil || !isRequireCall(vd.Initializer) {
					continue
				}
				name := vd.Name()
				if name == nil || name.Kind != ast.KindObjectBindingPattern {
					continue
				}
				bp := name.AsBindingPattern()
				if bp.Elements == nil {
					continue
				}
				for _, el := range bp.Elements.Nodes {
					be := el.AsBindingElement()
					// Upstream's filter is `property.type === 'Property'`,
					// which excludes RestElement (`{ ...rest }`). tsgo
					// represents rest patterns as a BindingElement with a
					// non-nil DotDotDotToken; skip them.
					if be.DotDotDotToken != nil {
						continue
					}
					if be.Name() == nil || be.Name().Kind != ast.KindIdentifier {
						continue
					}
					// Upstream maps `property.value.name` (local) → `property.key.name`
					// (imported). `{ T: U }` → PropertyName="T", name="U" → U → T.
					// Shorthand `{ T }` → PropertyName=nil, name="T" → T → T (no-op).
					localName := be.Name().AsIdentifier().Text
					importedName := localName
					if be.PropertyName != nil && be.PropertyName.Kind == ast.KindIdentifier {
						importedName = be.PropertyName.AsIdentifier().Text
					}
					out[localName] = importedName
				}
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sf.AsNode())
	return out
}

// trimmedSource returns `node`'s raw source text with leading/trailing whitespace stripped.
func trimmedSource(ctx rule.RuleContext, node *ast.Node) string {
	return strings.TrimSpace(utils.TrimmedNodeText(ctx.SourceFile, node))
}

// stringLiteralValue extracts the cooked text of a StringLiteral node.
func stringLiteralValue(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	}
	return ""
}

func handleJsxAttribute(ctx rule.RuleContext, node *ast.Node, cfg *ruleConfig, renamedImports map[string]string) {
	attr := node.AsJsxAttribute()
	if attr.Initializer == nil {
		return
	}
	init := attr.Initializer
	if init.Kind != ast.KindStringLiteral {
		return
	}
	resolved := resolveOverride(node, cfg, renamedImports)
	if resolved == nil {
		resolved = &cfg.base
	}

	// Upstream gates the restrictedAttributes branch on
	// `node.name.type === 'JSXIdentifier'`, which excludes JsxNamespacedName
	// (`xlink:href`). The noStrings fall-through still applies to
	// namespaced attrs — match that asymmetry.
	nameNode := attr.Name()
	isSimpleIdentifier := nameNode != nil && nameNode.Kind == ast.KindIdentifier
	if len(resolved.restrictedAttributes) > 0 && isSimpleIdentifier {
		attrName := nameNode.AsIdentifier().Text
		if _, restricted := resolved.restrictedAttributes[attrName]; restricted {
			cooked := stringLiteralValue(init)
			if _, allowed := resolved.allowedStrings[strings.TrimSpace(cooked)]; !allowed {
				reportRestrictedAttribute(ctx, node, trimmedSource(ctx, init), attrName, resolved)
			}
			return
		}
	}

	if !resolved.noStrings || resolved.ignoreProps {
		return
	}
	cooked := stringLiteralValue(init)
	if _, allowed := resolved.allowedStrings[cooked]; allowed {
		return
	}
	messageId := msgInvalidPropValue
	if resolved.configType == configTypeOverride {
		messageId = msgInvalidPropValueInElement
	}
	reportLiteralNode(ctx, node, messageId, resolved, trimmedSource(ctx, node))
}

// handleStringLiteral handles upstream's Literal listener for string-typed values.
// In tsgo a StringLiteral nested under a JsxAttribute is also routed here, but
// the upstream `Literal` handler intentionally bails on attribute-direct
// literals (the JSXAttribute handler owns that report); replicated here via
// the parent-kind check inside `isStandardJSXNode`.
func handleStringLiteral(ctx rule.RuleContext, node *ast.Node, cfg *ruleConfig, renamedImports map[string]string) {
	resolved := resolveOverride(node, cfg, renamedImports)
	if resolved == nil {
		resolved = &cfg.base
	}
	hasJSXContent := hasJSXContentParentOrGrandParent(node)
	if hasJSXContent && shouldAllowElement(resolved) {
		return
	}
	parent := parentIgnoringBinaryAndParens(node)
	if parent == nil {
		return
	}
	cooked := node.AsStringLiteral().Text
	rawSource := trimmedSource(ctx, node)
	if !isViableTextNode(rawSource, cooked, parent.Kind, resolved) {
		return
	}
	// Upstream gates the report on `hasJSXParentOrGrandParent || !config.ignoreProps`
	// where `config` is the BASE config — not the resolved override. Preserved
	// here so an override can't bypass the base `ignoreProps` for non-content
	// literals (e.g. attribute-context literals when the base says ignoreProps).
	if !hasJSXContent && cfg.base.ignoreProps {
		return
	}
	reportLiteralNode(ctx, node, defaultMessageId(hasJSXContent, resolved), resolved, rawSource)
}

func handleJsxText(ctx rule.RuleContext, node *ast.Node, cfg *ruleConfig, renamedImports map[string]string) {
	resolved := resolveOverride(node, cfg, renamedImports)
	if resolved == nil {
		resolved = &cfg.base
	}
	if shouldAllowElement(resolved) {
		return
	}
	jt := node.AsJsxText()
	if jt.ContainsOnlyTriviaWhiteSpaces {
		return
	}
	// tsgo's JsxText.Text is the raw source text — HTML entities like
	// `&mdash;` stay encoded (decoding happens later, during JSX transform).
	// ESLint's espree+acorn-jsx feeds the rule a decoded `node.value`, so to
	// keep `allowedStrings` lookups byte-equivalent we decode here. The raw
	// form is still checked separately via `rawSource` below.
	cooked := html.UnescapeString(jt.Text)
	rawSource := trimmedSource(ctx, node)
	parent := node.Parent
	if parent == nil {
		return
	}
	if !isViableTextNode(rawSource, cooked, parent.Kind, resolved) {
		return
	}
	hasJSXContent := hasJSXContentParentOrGrandParent(node)
	reportLiteralNode(ctx, node, defaultMessageId(hasJSXContent, resolved), resolved, rawSource)
}

// handleTemplate handles upstream's TemplateLiteral listener for both
// no-substitution and substitution shapes. tsgo splits these into
// KindNoSubstitutionTemplateLiteral and KindTemplateExpression respectively;
// upstream's ESTree groups both under TemplateLiteral.
func handleTemplate(ctx rule.RuleContext, node *ast.Node, cfg *ruleConfig, renamedImports map[string]string) {
	parent := parentIgnoringBinaryAndParens(node)
	if parent == nil {
		return
	}
	// Upstream guards `isParentJSXExpressionCont` — only fire when the
	// container is a JsxExpression. NoSubstitutionTemplateLiteral can also
	// appear directly as a JsxAttribute initializer in some grammars; tsgo
	// rejects that form, so the JsxExpression-only gate is sufficient.
	if parent.Kind != ast.KindJsxExpression {
		return
	}
	grand := parent.Parent
	for grand != nil && grand.Kind == ast.KindParenthesizedExpression {
		grand = grand.Parent
	}
	isParentJSXElement := grand != nil && jsxContentParentKind(grand.Kind)

	resolved := resolveOverride(node, cfg, renamedImports)
	if resolved == nil {
		resolved = &cfg.base
	}
	if !resolved.noStrings {
		return
	}
	if !isParentJSXElement && resolved.ignoreProps {
		return
	}
	reportLiteralNode(ctx, node, defaultMessageId(isParentJSXElement, resolved), resolved, trimmedSource(ctx, node))
}

var JsxNoLiteralsRule = rule.Rule{
	Name: "react/jsx-no-literals",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.UnwrapOptions(_options)
		cfg := parseConfig(options)
		var renamedImports map[string]string
		if cfg.hasElementOverrides() {
			renamedImports = collectRenamedImports(ctx.SourceFile)
		} else {
			renamedImports = map[string]string{}
		}

		return rule.RuleListeners{
			ast.KindStringLiteral: func(node *ast.Node) {
				handleStringLiteral(ctx, node, &cfg, renamedImports)
			},
			ast.KindJsxText: func(node *ast.Node) {
				handleJsxText(ctx, node, &cfg, renamedImports)
			},
			ast.KindNoSubstitutionTemplateLiteral: func(node *ast.Node) {
				handleTemplate(ctx, node, &cfg, renamedImports)
			},
			ast.KindTemplateExpression: func(node *ast.Node) {
				handleTemplate(ctx, node, &cfg, renamedImports)
			},
			ast.KindJsxAttribute: func(node *ast.Node) {
				handleJsxAttribute(ctx, node, &cfg, renamedImports)
			},
		}
	},
}
