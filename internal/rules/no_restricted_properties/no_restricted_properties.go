package no_restricted_properties

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_restricted_properties.schema.json
var schemaJSON []byte

// pairRestriction is a specific `{ object, property }` restriction: it never
// carries an allow-list (the schema forbids combining `allowProperties` with
// `property` and `allowObjects` with `object`).
type pairRestriction struct {
	message string
}

// objectRestriction is an `{ object, allowProperties? }` restriction: every
// property on the named object is restricted except those in allowProperties.
type objectRestriction struct {
	allowProperties    []string
	hasAllowProperties bool
	message            string
}

// propertyOnlyRestriction is a `{ property, allowObjects? }` restriction: the
// named property is restricted on every object except those in allowObjects.
type propertyOnlyRestriction struct {
	allowObjects    []string
	hasAllowObjects bool
	message         string
}

type restrictions struct {
	// pairs[objectName][propertyName] holds a specific object+property restriction.
	pairs map[string]map[string]pairRestriction
	// objects[objectName] holds a whole-object restriction (no `property`).
	objects map[string]objectRestriction
	// properties[propertyName] holds a whole-property restriction (no `object`).
	properties map[string]propertyOnlyRestriction
}

func parseOptions(options []any) restrictions {
	r := restrictions{
		pairs:      map[string]map[string]pairRestriction{},
		objects:    map[string]objectRestriction{},
		properties: map[string]propertyOnlyRestriction{},
	}

	for _, raw := range options {
		m, ok := raw.(map[string]any)
		if !ok {
			continue
		}

		objectName, hasObject := m["object"].(string)
		propertyName, hasProperty := m["property"].(string)
		message, _ := m["message"].(string)

		switch {
		case !hasObject:
			allowObjects, hasAllowObjects := toStringSlice(m["allowObjects"])
			r.properties[propertyName] = propertyOnlyRestriction{
				allowObjects:    allowObjects,
				hasAllowObjects: hasAllowObjects,
				message:         message,
			}
		case !hasProperty:
			allowProperties, hasAllowProperties := toStringSlice(m["allowProperties"])
			r.objects[objectName] = objectRestriction{
				allowProperties:    allowProperties,
				hasAllowProperties: hasAllowProperties,
				message:            message,
			}
		default:
			if r.pairs[objectName] == nil {
				r.pairs[objectName] = map[string]pairRestriction{}
			}
			r.pairs[objectName][propertyName] = pairRestriction{
				message: message,
			}
		}
	}

	return r
}

func toStringSlice(v any) ([]string, bool) {
	arr, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out, true
}

func isAllowedName(name string, allowedList []string) bool {
	for _, allowed := range allowedList {
		if allowed == name {
			return true
		}
	}
	return false
}

func reportRange(ctx rule.RuleContext, node *ast.Node) core.TextRange {
	textRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	if node.Kind != ast.KindObjectBindingPattern || node.Parent == nil {
		return textRange
	}

	var name, typeNode *ast.Node
	switch node.Parent.Kind {
	case ast.KindVariableDeclaration:
		declaration := node.Parent.AsVariableDeclaration()
		name, typeNode = declaration.Name(), declaration.Type
	case ast.KindParameter:
		parameter := node.Parent.AsParameterDeclaration()
		name, typeNode = parameter.Name(), parameter.Type
	}
	if name == node && typeNode != nil && typeNode.Pos() >= node.End() {
		return core.NewTextRange(textRange.Pos(), typeNode.End())
	}
	return textRange
}

// matchedProperty is the object-side match found for a given (objectName,
// propertyName) pair: either a specific pair restriction (no allow-list) or a
// whole-object restriction (with an optional allowProperties list).
type matchedProperty struct {
	found        bool
	allowList    []string
	hasAllowList bool
	message      string
}

// checkPropertyAccess mirrors upstream's checkPropertyAccess(node, objectName,
// propertyName): it reports at most one diagnostic on node, preferring a
// specific object+property (or whole-object) match over a whole-property
// match — even when the object-side match turns out to be allowed.
func checkPropertyAccess(ctx rule.RuleContext, r restrictions, node *ast.Node, objectName string, hasObjectName bool, propertyName string, hasPropertyName bool) {
	if !hasPropertyName {
		return
	}

	var matched matchedProperty
	if hasObjectName {
		if propMap, ok := r.pairs[objectName]; ok {
			// A specific pair restriction on this object shadows any
			// whole-object restriction: a mismatched property here does NOT
			// fall back to the whole-object entry, even if one exists.
			if pr, ok2 := propMap[propertyName]; ok2 {
				matched = matchedProperty{found: true, message: pr.message}
			}
		} else if obj, ok2 := r.objects[objectName]; ok2 {
			matched = matchedProperty{
				found:        true,
				allowList:    obj.allowProperties,
				hasAllowList: obj.hasAllowProperties,
				message:      obj.message,
			}
		}
	}

	if matched.found && (!matched.hasAllowList || !isAllowedName(propertyName, matched.allowList)) {
		allowedPropertiesMessage := ""
		if matched.hasAllowList {
			allowedPropertiesMessage = " Only these properties are allowed: " + strings.Join(matched.allowList, ", ") + "."
		}
		messageSuffix := ""
		if matched.message != "" {
			messageSuffix = " " + matched.message
		}
		ctx.ReportRange(reportRange(ctx, node), rule.RuleMessage{
			Id:          "restrictedObjectProperty",
			Description: fmt.Sprintf("'%s.%s' is restricted from being used.%s%s", objectName, propertyName, allowedPropertiesMessage, messageSuffix),
		})
		return
	}

	globalProp, ok := r.properties[propertyName]
	if !ok {
		return
	}
	if hasObjectName && globalProp.hasAllowObjects && isAllowedName(objectName, globalProp.allowObjects) {
		return
	}

	allowedObjectsMessage := ""
	if globalProp.hasAllowObjects {
		allowedObjectsMessage = fmt.Sprintf(" Property '%s' is only allowed on these objects: %s.", propertyName, strings.Join(globalProp.allowObjects, ", "))
	}
	messageSuffix := ""
	if globalProp.message != "" {
		messageSuffix = " " + globalProp.message
	}
	ctx.ReportRange(reportRange(ctx, node), rule.RuleMessage{
		Id:          "restrictedProperty",
		Description: fmt.Sprintf("'%s' is restricted from being used.%s%s", propertyName, allowedObjectsMessage, messageSuffix),
	})
}

// staticObjectName returns the identifier text of expr, or ("", false) if
// expr is not (after unwrapping parentheses) a plain identifier. Matches
// upstream's `node.object && node.object.name`: ESTree never wraps an
// identifier receiver in parentheses, so tsgo's explicit ParenthesizedExpression
// must be unwrapped to match. Non-null assertions (`x!`) and type-expression
// wrappers (`x as T`) are deliberately NOT unwrapped, since ESTree (even under
// typescript-eslint) never collapses those back to a bare Identifier either.
func staticObjectName(expr *ast.Node) (string, bool) {
	if expr == nil {
		return "", false
	}
	expr = ast.SkipParentheses(expr)
	if expr != nil && expr.Kind == ast.KindIdentifier {
		return expr.AsIdentifier().Text, true
	}
	return "", false
}

// staticMemberPropertyName returns the statically-known property name of a
// PropertyAccessExpression or ElementAccessExpression, or ("", false) if the
// name can't be determined at compile time. A private-name receiver
// (`this.#foo`) is deliberately never static: it's a distinct name-space from
// string-keyed properties, matching upstream's astUtils.getStaticPropertyName
// (which only special-cases plain Identifier member names).
func staticMemberPropertyName(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return utils.GetStaticPropertyName(node.AsPropertyAccessExpression().Name())
	case ast.KindQualifiedName:
		return utils.GetStaticPropertyName(node.AsQualifiedName().Right)
	case ast.KindElementAccessExpression:
		arg := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		return utils.GetStaticExpressionValue(arg)
	}
	return "", false
}

// staticPatternPropertyName returns the statically-known property name of a
// binding-pattern element (BindingElement) or object-literal-pattern property
// (PropertyAssignment / ShorthandPropertyAssignment), or ("", false) when it
// can't be determined or when el is a rest element (`...rest` never names a
// property to restrict, matching upstream: getStaticPropertyName has no case
// for RestElement/SpreadElement).
func staticPatternPropertyName(el *ast.Node) (string, bool) {
	switch el.Kind {
	case ast.KindBindingElement:
		be := el.AsBindingElement()
		if be.DotDotDotToken != nil {
			return "", false
		}
		name := el.PropertyNameOrName()
		if name == nil {
			return "", false
		}
		return utils.GetStaticPropertyName(name)
	case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
		name := el.Name()
		if name == nil {
			return "", false
		}
		return utils.GetStaticPropertyName(name)
	}
	return "", false
}

// patternObjectName computes the object name a destructuring pattern is
// pulled from, by looking only at the pattern's *immediate* parent — matching
// upstream, which checks node.parent.type without walking further up.
//
// tsgo has no dedicated node for ESLint's AssignmentPattern (the `= default`
// wrapper for a pattern element with a default value): a default is just
// ordinary `=` syntax, so it surfaces as a BinaryExpression in expression
// position and as an Initializer field in declaration/parameter/binding-element
// position. That collapses upstream's three separate parent-type checks
// (VariableDeclarator.init / AssignmentExpression.right / AssignmentPattern.right)
// into one shape per tsgo parent kind.
func patternObjectName(pattern *ast.Node) (string, bool) {
	parent := pattern.Parent
	if parent == nil {
		return "", false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		return staticObjectName(parent.AsVariableDeclaration().Initializer)
	case ast.KindParameter:
		return staticObjectName(parent.AsParameterDeclaration().Initializer)
	case ast.KindBindingElement:
		return staticObjectName(parent.AsBindingElement().Initializer)
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			return staticObjectName(binary.Right)
		}
	}
	return "", false
}

// checkPattern replicates upstream's ObjectPattern listener for one pattern
// node: it resolves the pattern's own object name (if any) and checks each of
// its immediate elements — nested patterns are visited independently when the
// traversal reaches them as their own listener call, exactly as upstream's
// ObjectPattern visitor fires once per ObjectPattern node in the tree.
func checkPattern(ctx rule.RuleContext, r restrictions, pattern *ast.Node) {
	objectName, hasObjectName := patternObjectName(pattern)

	var elements []*ast.Node
	switch pattern.Kind {
	case ast.KindObjectBindingPattern:
		bp := pattern.AsBindingPattern()
		if bp.Elements != nil {
			elements = bp.Elements.Nodes
		}
	case ast.KindObjectLiteralExpression:
		props := pattern.AsObjectLiteralExpression().Properties
		if props != nil {
			elements = props.Nodes
		}
	}

	for _, el := range elements {
		propertyName, hasPropertyName := staticPatternPropertyName(el)
		checkPropertyAccess(ctx, r, pattern, objectName, hasObjectName, propertyName, hasPropertyName)
	}
}

func isObjectLiteralDestructuringPattern(node *ast.Node) bool {
	if ast.IsArrayLiteralOrObjectLiteralDestructuringPattern(node) {
		return true
	}
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}

	current := node
	for current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindForInStatement, ast.KindForOfStatement:
			statement := parent.AsForInOrOfStatement()
			return statement != nil && statement.Initializer == current
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			return binary != nil && binary.Left == current && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
			current = parent
		case ast.KindSpreadElement:
			spread := parent.AsSpreadElement()
			if spread == nil || spread.Expression != current || parent.Parent == nil || parent.Parent.Kind != ast.KindArrayLiteralExpression {
				return false
			}
			current = parent
		case ast.KindPropertyAssignment:
			property := parent.AsPropertyAssignment()
			if property == nil || property.Initializer != current || parent.Parent == nil {
				return false
			}
			current = parent.Parent
		default:
			return false
		}
	}
	return false
}

// https://eslint.org/docs/latest/rules/no-restricted-properties
var NoRestrictedPropertiesRule = rule.Rule{
	Name:   "no-restricted-properties",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		if len(options) == 0 {
			return nil
		}
		r := parseOptions(options)

		checkMemberAccess := func(node *ast.Node) {
			if utils.IsInJsxTagName(node) {
				return
			}

			objectExpr, property := utils.MemberExpressionParts(node)
			if objectExpr == nil || property == nil {
				return
			}
			objectName, hasObjectName := staticObjectName(objectExpr)
			propertyName, hasPropertyName := staticMemberPropertyName(node)
			checkPropertyAccess(ctx, r, node, objectName, hasObjectName, propertyName, hasPropertyName)
		}

		return rule.RuleListeners{
			ast.KindPropertyAccessExpression: checkMemberAccess,
			ast.KindElementAccessExpression:  checkMemberAccess,
			ast.KindQualifiedName:            checkMemberAccess,
			ast.KindObjectBindingPattern: func(node *ast.Node) {
				checkPattern(ctx, r, node)
			},
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				if !isObjectLiteralDestructuringPattern(node) {
					return
				}
				checkPattern(ctx, r, node)
			},
		}
	},
}
