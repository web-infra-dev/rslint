package sort_default_props

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed sort_default_props.schema.json
var schemaJSON []byte

var SortDefaultPropsRule = rule.Rule{
	Name:   "react/sort-default-props",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		ignoreCase := false
		if len(options) != 0 {
			config, _ := options[0].(map[string]any)
			ignoreCase, _ = config["ignoreCase"].(bool)
		}
		checkNode := func(node *ast.Node) {
			node = utils.ESTreeRuntimeExpression(node)
			if node == nil {
				return
			}
			if node.Kind == ast.KindIdentifier {
				// NOTE: Unlike ESLint's React variable helper, resolve only lexical
				// bindings; searching child scopes can select an unrelated object.
				symbol := ctx.Refs.ResolveInFile(node)
				if symbol == nil || len(symbol.Declarations) == 0 {
					return
				}
				declaration := symbol.Declarations[0]
				if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
					return
				}
				node = utils.ESTreeRuntimeExpression(declaration.AsVariableDeclaration().Initializer)
			}
			if node == nil || node.Kind != ast.KindObjectLiteralExpression {
				return
			}
			var previous string
			hasPrevious := false
			for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
				if property.Kind == ast.KindSpreadAssignment {
					hasPrevious = false
					continue
				}
				key := property.Name()
				if key == nil {
					continue
				}
				if key.Kind == ast.KindComputedPropertyName {
					key = utils.ESTreeRuntimeExpression(key.AsComputedPropertyName().Expression)
				}
				if key == nil {
					continue
				}
				// Compare authored keys, including quotes and escapes, as upstream does.
				current := utils.TrimmedNodeText(ctx.SourceFile, key)
				if ignoreCase {
					current = ecmascript.StringToLowerCase(current)
				}
				if hasPrevious && ecmascript.CompareStrings(current, previous) < 0 {
					// Keep the largest key in this spread-delimited group.
					ctx.ReportNode(property, rule.RuleMessage{
						Id:          "propsNotSorted",
						Description: "Default prop types declarations should be sorted alphabetically",
					})
				} else {
					previous = current
					hasPrevious = true
				}
			}
		}
		checkMember := func(node *ast.Node) {
			var name *ast.Node
			if node.Kind == ast.KindPropertyAccessExpression {
				name = node.Name()
			} else {
				name = utils.ESTreeRuntimeExpression(node.AsElementAccessExpression().ArgumentExpression)
			}
			if !isDefaultPropsName(name) || ast.IsOptionalChain(node) {
				return
			}
			parent := utils.ESTreeParent(node)
			// Upstream checks any parent with a right operand, including comparisons.
			if parent != nil && parent.Kind == ast.KindBinaryExpression && parent.AsBinaryExpression().OperatorToken.Kind != ast.KindCommaToken {
				checkNode(parent.AsBinaryExpression().Right)
			}
		}
		return rule.RuleListeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				// TypeScript auto-accessors are AccessorProperty nodes upstream.
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsAccessor) && isDefaultPropsName(node.Name()) {
					checkNode(node.AsPropertyDeclaration().Initializer)
				}
			},
			ast.KindPropertyAccessExpression: checkMember,
			ast.KindElementAccessExpression:  checkMember,
		}
	},
}

func isDefaultPropsName(name *ast.Node) bool {
	return reactutil.IsAuthoredPropertyName(name, "defaultProps") || reactutil.IsAuthoredPropertyName(name, "getDefaultProps")
}
