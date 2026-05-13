package alt_text

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jsx_a11y/jsxa11yutil"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Default DOM elements alt-text validates. Mirrors upstream's `DEFAULT_ELEMENTS`.
// The order is preserved because option-driven custom-component lookup falls
// back to `elementOptions.find(...)` which scans in this order.
var defaultElements = []string{
	"img",
	"object",
	"area",
	`input[type="image"]`,
}

// Element name aliases. The upstream rule normalizes the special
// `input[type="image"]` selector to the bare tag `input` when it builds the
// `typesToValidate` set, so the listener can match by tag name. We mirror
// that translation in `domToTagName` and reverse it via `tagNameToDOM` when
// dispatching to the per-element check.
func domToTagName(dom string) string {
	if dom == `input[type="image"]` {
		return "input"
	}
	return dom
}

const (
	msgPreferAlt           = `Prefer alt="" over a presentational role. First rule of aria is to not use aria if it can be achieved via native HTML.`
	msgAriaLabelEmpty      = "The aria-label attribute must have a value. The alt attribute is preferred over aria-label for images."
	msgAriaLabelledByEmpty = "The aria-labelledby attribute must have a value. The alt attribute is preferred over aria-labelledby for images."
	msgObject              = "Embedded <object> elements must have alternative text by providing inner text, aria-label or aria-labelledby props."
	msgArea                = "Each area of an image map must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop."
	msgInputImage          = "<input> elements with type=\"image\" must have a text alternative through the `alt`, `aria-label`, or `aria-labelledby` prop."
)

func missingPropMessage(nodeType string) string {
	return nodeType + " elements must have an alt prop, either with meaningful text, or an empty string for decorative images."
}

func altValueMessage(nodeType string) string {
	return "Invalid alt value for " + nodeType + ". Use alt=\"\" for presentational images."
}

// options holds the parsed shape of the rule's first option object.
type options struct {
	elements         []string
	customComponents map[string][]string // keyed by DOM element name (e.g. "img" / `input[type="image"]`)
}

func parseOptions(raw any) options {
	opts := options{elements: defaultElements}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}

	if rawElements, ok := m["elements"]; ok {
		// Upstream falls through to DEFAULT_ELEMENTS via `||` only when
		// `options.elements` is falsy (undefined / null / "" / []). An
		// explicitly-empty array IS truthy in JS, so it replaces the
		// default and disables ALL element checks. The `ok` guard is the
		// "absent vs present" distinction; StringSliceOption returns a
		// (possibly empty) `[]string` for any present `[]interface{}`.
		if elements := jsxa11yutil.StringSliceOption(rawElements); elements != nil {
			opts.elements = elements
		}
	}

	opts.customComponents = make(map[string][]string)
	for _, element := range opts.elements {
		if list := jsxa11yutil.StringSliceOption(m[element]); len(list) > 0 {
			opts.customComponents[element] = list
		}
	}
	return opts
}

var AltTextRule = rule.Rule{
	Name: "jsx-a11y/alt-text",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// Precompute the set of nodeType strings that should trigger a check.
		// Upstream's `typesToValidate` is the union of customComponents and
		// elementOptions, with `input[type="image"]` translated to `input`
		// because the listener matches by raw tag-name (`<input>`).
		typesToValidate := make(map[string]struct{})
		for _, element := range opts.elements {
			typesToValidate[domToTagName(element)] = struct{}{}
		}
		for _, components := range opts.customComponents {
			for _, c := range components {
				typesToValidate[c] = struct{}{}
			}
		}

		// elementType resolves the effective HTML name for a JSX node, applying
		// jsx-a11y's polymorphicPropName / componentMap transformations.
		elementType := func(node *ast.Node) string {
			return jsxa11yutil.GetElementType(node, ctx.Settings)
		}

		check := func(node *ast.Node) {
			nodeType := elementType(node)
			if _, ok := typesToValidate[nodeType]; !ok {
				return
			}

			// Determine which DOM-element bucket this nodeType maps to so we
			// can dispatch to the right per-element rule. Mirrors upstream:
			//   - `nodeType === "input"` → `input[type="image"]`
			//   - `nodeType in elementOptions` → use it directly
			//   - otherwise it's a custom component → find which DOM element
			//     it was mapped to via `options[element]`
			domElement := nodeType
			if domElement == "input" {
				domElement = `input[type="image"]`
			}
			if !contains(opts.elements, domElement) {
				domElement = ""
				for _, element := range opts.elements {
					if list, ok := opts.customComponents[element]; ok {
						if contains(list, nodeType) {
							domElement = element
							break
						}
					}
				}
				if domElement == "" {
					// Upstream's `find` returns undefined here, then calls
					// `ruleByElement[undefined](...)` which throws. In
					// practice this is unreachable because `typesToValidate`
					// exactly covers the nodeTypes we accept; we add the
					// guard for malformed configs (e.g. `elements: []` with
					// custom components still listed).
					return
				}
			}

			attrs := reactutil.GetJsxElementAttributes(node)
			switch domElement {
			case "img":
				checkImg(ctx, node, attrs, nodeType)
			case "object":
				checkObject(ctx, node, attrs, elementType)
			case "area":
				checkArea(ctx, node, attrs)
			case `input[type="image"]`:
				checkInputImage(ctx, node, attrs, nodeType)
			}
		}

		return rule.RuleListeners{
			ast.KindJsxOpeningElement:     check,
			ast.KindJsxSelfClosingElement: check,
		}
	},
}

// checkImg mirrors `ruleByElement.img`. Branches:
//
//  1. altProp absent:
//     a) presentational role → preferAlt
//     b) aria-label present and value is undefined / "" → ariaLabelEmpty
//     c) aria-labelledby present and value is undefined / "" → ariaLabelledByEmpty
//     d) otherwise → missingProp
//  2. altProp present:
//     a) (truthy && not boolean form) || empty string → valid
//     b) otherwise → altValue
func checkImg(ctx rule.RuleContext, node *ast.Node, attrs []*ast.Node, nodeType string) {
	altProp := jsxa11yutil.FindAttributeByName(attrs, "alt")
	if altProp == nil {
		if jsxa11yutil.IsPresentationRole(attrs) {
			ctx.ReportNode(node, rule.RuleMessage{
				Id:          "preferAlt",
				Description: msgPreferAlt,
			})
			return
		}
		if ariaLabelProp := jsxa11yutil.FindAttributeByName(attrs, "aria-label"); ariaLabelProp != nil {
			if !jsxa11yutil.AriaLabelHasValue(ariaLabelProp) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "ariaLabelValue",
					Description: msgAriaLabelEmpty,
				})
			}
			return
		}
		if ariaLabelledByProp := jsxa11yutil.FindAttributeByName(attrs, "aria-labelledby"); ariaLabelledByProp != nil {
			if !jsxa11yutil.AriaLabelHasValue(ariaLabelledByProp) {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "ariaLabelledByValue",
					Description: msgAriaLabelledByEmpty,
				})
			}
			return
		}
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "missingProp",
			Description: missingPropMessage(nodeType),
		})
		return
	}
	if jsxa11yutil.AltAttributeIsValid(altProp) {
		return
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "altValue",
		Description: altValueMessage(nodeType),
	})
}

// checkObject mirrors `ruleByElement.object`. An <object> is accessible if it
// has aria-label / aria-labelledby with a meaningful value, a literal title,
// OR an accessible child inside its element.
func checkObject(ctx rule.RuleContext, node *ast.Node, attrs []*ast.Node, elementType func(*ast.Node) string) {
	if jsxa11yutil.AriaLabelHasValue(jsxa11yutil.FindAttributeByName(attrs, "aria-label")) {
		return
	}
	if jsxa11yutil.AriaLabelHasValue(jsxa11yutil.FindAttributeByName(attrs, "aria-labelledby")) {
		return
	}
	// Upstream: `hasTitleAttr = !!getLiteralPropValue(...)`. LiteralPropTruthy
	// mirrors that fully — including jsx-ast-utils' quirks like
	// `title={null}` evaluating to the string "null" (truthy) and Identifier
	// references resolving to null (falsy because they're not literal).
	if jsxa11yutil.LiteralPropTruthy(jsxa11yutil.FindAttributeByName(attrs, "title")) {
		return
	}
	// Upstream calls `hasAccessibleChild(node.parent, elementType)`. In
	// ESTree, every JSXOpeningElement (including self-closing) is wrapped
	// in a JSXElement, so `node.parent` always carries the children list
	// AND the opening attributes. tsgo doesn't wrap JsxSelfClosingElement
	// in a JsxElement — the self-closing node carries the attributes
	// directly. We normalize: pass the JsxElement when present, else the
	// self-closing node itself, so HasAccessibleChild sees the same
	// "children + opening attributes" surface upstream does.
	jsxRoot := node
	if node.Kind == ast.KindJsxOpeningElement && node.Parent != nil &&
		(node.Parent.Kind == ast.KindJsxElement || node.Parent.Kind == ast.KindJsxFragment) {
		jsxRoot = node.Parent
	}
	if jsxa11yutil.HasAccessibleChild(jsxRoot, elementType) {
		return
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "object",
		Description: msgObject,
	})
}

// checkArea mirrors `ruleByElement.area`. An <area> is accessible if it has
// aria-label / aria-labelledby with a meaningful value, OR a valid alt prop.
func checkArea(ctx rule.RuleContext, node *ast.Node, attrs []*ast.Node) {
	if jsxa11yutil.AriaLabelHasValue(jsxa11yutil.FindAttributeByName(attrs, "aria-label")) {
		return
	}
	if jsxa11yutil.AriaLabelHasValue(jsxa11yutil.FindAttributeByName(attrs, "aria-labelledby")) {
		return
	}
	altProp := jsxa11yutil.FindAttributeByName(attrs, "alt")
	if altProp == nil {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "area",
			Description: msgArea,
		})
		return
	}
	if jsxa11yutil.AltAttributeIsValid(altProp) {
		return
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "area",
		Description: msgArea,
	})
}

// checkInputImage mirrors `ruleByElement['input[type="image"]']`. For the
// bare `<input>` tag we additionally guard on `type === "image"` — only
// inputs of type "image" require alt text. Custom components (mapped to
// "input[type=\"image\"]" via options) skip this guard, matching upstream.
func checkInputImage(ctx rule.RuleContext, node *ast.Node, attrs []*ast.Node, nodeType string) {
	if nodeType == "input" {
		typeAttr := jsxa11yutil.FindAttributeByName(attrs, "type")
		// Upstream: `getPropValue(...)` (NOT getLiteralPropValue) compared
		// `=== "image"` — case-sensitive, full static evaluation across
		// `+`, `&&` / `||` short-circuit, ternary, etc. Use
		// PropStaticStringValue and exact `==` comparison.
		v, ok := jsxa11yutil.PropStaticStringValue(typeAttr)
		if !ok || v != "image" {
			return
		}
	}
	if jsxa11yutil.AriaLabelHasValue(jsxa11yutil.FindAttributeByName(attrs, "aria-label")) {
		return
	}
	if jsxa11yutil.AriaLabelHasValue(jsxa11yutil.FindAttributeByName(attrs, "aria-labelledby")) {
		return
	}
	altProp := jsxa11yutil.FindAttributeByName(attrs, "alt")
	if altProp == nil {
		ctx.ReportNode(node, rule.RuleMessage{
			Id:          "inputImage",
			Description: msgInputImage,
		})
		return
	}
	if jsxa11yutil.AltAttributeIsValid(altProp) {
		return
	}
	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "inputImage",
		Description: msgInputImage,
	})
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
