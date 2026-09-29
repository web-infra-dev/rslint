package prefer_exact_props

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const propTypesMessage = "Component propTypes should be exact by using %s."

var PreferExactPropsRule = rule.Rule{
	Name:   "react/prefer-exact-props",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		exactWrappers := getExactPropWrappers(ctx.Settings)
		if len(exactWrappers) == 0 {
			return rule.RuleListeners{}
		}

		definitions := reactutil.NewVariableDefinitionLookup(ctx)
		message := rule.RuleMessage{
			Id:          "propTypes",
			Description: fmt.Sprintf(propTypesMessage, formatExactPropWrappers(exactWrappers)),
		}

		reportIfNonExact := func(reportNode, value *ast.Node, resolveIdentifier bool) {
			if isNonExactPropTypesValue(ctx, definitions, value, exactWrappers, resolveIdentifier) {
				ctx.ReportNode(reportNode, message)
			}
		}

		return rule.RuleListeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				property := node.AsPropertyDeclaration()
				if property == nil || !isPropTypesPropertyDeclaration(property) {
					return
				}
				// The upstream class-property listener only inspects direct object
				// and call initializers; identifier resolution belongs solely to
				// its MemberExpression assignment path.
				reportIfNonExact(node, property.Initializer, false)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				if binary == nil {
					return
				}
				left := utils.ESTreeRuntimeExpression(binary.Left)
				if !isPropTypesMember(left) {
					return
				}
				// Optional members sit inside an ESTree ChainExpression rather than
				// directly under the binary expression inspected upstream.
				if ast.IsOptionalChain(left) {
					return
				}
				reportIfNonExact(left, binary.Right, true)
			},
		}
	},
}

func getExactPropWrappers(settings map[string]interface{}) []reactutil.PropWrapperEntry {
	wrappers := reactutil.GetPropWrapperFunctions(settings)
	exact := make([]reactutil.PropWrapperEntry, 0, len(wrappers))
	for _, wrapper := range wrappers {
		if wrapper.Exact {
			exact = append(exact, wrapper)
		}
	}
	return exact
}

func formatExactPropWrappers(wrappers []reactutil.PropWrapperEntry) string {
	formatted := make([]string, 0, len(wrappers))
	for _, wrapper := range wrappers {
		name := wrapper.Property
		if wrapper.Object != "" {
			name = wrapper.Object + "." + wrapper.Property
		}
		formatted = append(formatted, "'"+name+"'")
	}
	joined := strings.Join(formatted, ", ")
	if len(formatted) > 1 {
		return "one of " + joined
	}
	return joined
}

func isNonExactPropTypesValue(
	ctx rule.RuleContext,
	definitions *reactutil.VariableDefinitionLookup,
	value *ast.Node,
	exactWrappers []reactutil.PropWrapperEntry,
	resolveIdentifier bool,
) bool {
	value = utils.ESTreeRuntimeExpression(value)
	if value == nil {
		return false
	}
	switch value.Kind {
	case ast.KindObjectLiteralExpression:
		return len(value.AsObjectLiteralExpression().Properties.Nodes) > 0
	case ast.KindCallExpression:
		// ESLint exposes any optional call/property chain as a ChainExpression,
		// so it never reaches the upstream CallExpression predicate.
		if ast.IsOptionalChain(value) {
			return false
		}
		return !isExactPropWrapperCall(ctx.SourceFile, value, exactWrappers)
	case ast.KindIdentifier:
		if !resolveIdentifier {
			return false
		}
		definition := definitions.First(value, value.AsIdentifier().Text)
		return isNonExactPropTypesDefinition(ctx, definition, exactWrappers)
	default:
		return false
	}
}

func isNonExactPropTypesDefinition(
	ctx rule.RuleContext,
	definition *ast.Node,
	exactWrappers []reactutil.PropWrapperEntry,
) bool {
	if definition == nil {
		return false
	}
	var initializer *ast.Node
	switch definition.Kind {
	case ast.KindVariableDeclaration:
		initializer = definition.AsVariableDeclaration().Initializer
	case ast.KindBindingElement:
		root := ast.GetRootDeclaration(definition)
		if root != nil && root.Kind == ast.KindVariableDeclaration {
			initializer = root.AsVariableDeclaration().Initializer
		}
	}
	initializer = utils.ESTreeRuntimeExpression(initializer)
	if initializer == nil {
		return false
	}
	switch initializer.Kind {
	case ast.KindObjectLiteralExpression:
		return len(initializer.AsObjectLiteralExpression().Properties.Nodes) > 0
	case ast.KindCallExpression:
		if ast.IsOptionalChain(initializer) {
			return false
		}
		return !isExactPropWrapperCall(ctx.SourceFile, initializer, exactWrappers)
	default:
		return false
	}
}

func isExactPropWrapperCall(sourceFile *ast.SourceFile, node *ast.Node, wrappers []reactutil.PropWrapperEntry) bool {
	if sourceFile == nil || node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := utils.ESTreeCallCallee(node.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	name := utils.TrimmedNodeText(sourceFile, callee)
	for _, wrapper := range wrappers {
		if name == wrapper.Property {
			return true
		}
		if strings.Count(name, ".") == 1 && wrapper.Object != "" && name == wrapper.Object+"."+wrapper.Property {
			return true
		}
	}
	return false
}

func isPropTypesPropertyDeclaration(property *ast.PropertyDeclaration) bool {
	if property == nil {
		return false
	}
	if isAuthoredPropertyName(property.Name(), "propTypes") {
		return true
	}
	// propsUtil.isPropTypesDeclaration treats an annotated class field named
	// `props` as a props declaration before checking the annotation language.
	return property.Type != nil && isAuthoredPropertyName(property.Name(), "props")
}

func isAuthoredPropertyName(name *ast.Node, expected string) bool {
	if name == nil {
		return false
	}
	if name.Kind == ast.KindComputedPropertyName {
		name = utils.ESTreeRuntimeExpression(name.AsComputedPropertyName().Expression)
	}
	return reactutil.IdentifierOrPrivateName(name) == expected
}

func isPropTypesMember(node *ast.Node) bool {
	if node == nil {
		return false
	}
	_, property := utils.MemberExpressionParts(node)
	property = utils.ESTreeRuntimeExpression(property)
	return reactutil.IdentifierOrPrivateName(property) == "propTypes"
}
