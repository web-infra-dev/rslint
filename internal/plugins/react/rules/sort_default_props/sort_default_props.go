package sort_default_props

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
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
		// NOTE: Unlike ESLint's React variable helper, resolve lexical value bindings.
		// Source definitions preserve order even when the binder replaces or
		// merges symbols (var/function and type/value declarations). Collect the
		// queried identifiers, then reuse the shared lexical reference graph once.
		var targets []*ast.Node
		names := make(map[string]struct{})
		identifiers := make(map[*ast.Node]*ast.Node)
		collect := func(node *ast.Node) {
			node = utils.ESTreeRuntimeExpression(node)
			if node == nil {
				return
			}
			switch node.Kind {
			case ast.KindIdentifier:
				names[node.Text()] = struct{}{}
				identifiers[node] = nil
			case ast.KindObjectLiteralExpression:
			default:
				return
			}
			targets = append(targets, node)
		}
		checkObject := func(node *ast.Node) {
			node = utils.ESTreeRuntimeExpression(node)
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
		return rule.RuleListeners{
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				if node.Name() == nil || ast.HasSyntacticModifier(node, ast.ModifierFlagsAccessor) {
					return
				}
				if name, ok := utils.GetStaticPropertyName(node.Name()); ok && isDefaultPropsName(name) {
					collect(node.Initializer())
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				binary := node.AsBinaryExpression()
				// Only direct and logical assignments can store this RHS object.
				// Reads, comparisons and iteration sources are not defaults.
				switch binary.OperatorToken.Kind {
				case ast.KindEqualsToken, ast.KindBarBarEqualsToken,
					ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken:
				default:
					return
				}
				left := utils.ESTreeRuntimeExpression(binary.Left)
				if left == nil || ast.IsOptionalChain(left) {
					return
				}
				if name, ok := utils.AccessExpressionStaticName(left); ok && isDefaultPropsName(name) {
					collect(binary.Right)
				}
			},
			ast.KindEndOfFile: func(_ *ast.Node) {
				if len(identifiers) != 0 {
					manager := scopeanalysis.References(ctx, names)
					for _, reference := range manager.References {
						if _, needed := identifiers[reference.Identifier]; !needed {
							continue
						}
						// Separate namespace declarations share exported values in TypeScript.
						// The ESLint scope graph keeps those declarations separate.
						if initializer, found := mergedNamespaceInitializer(ctx, reference); found {
							identifiers[reference.Identifier] = initializer
							continue
						}
						for _, declaration := range reference.Declarations {
							// Type declarations do not mask a runtime variable.
							if !declaration.IsValueBinding {
								continue
							}
							node := declaration.DefNode
							// Destructuring binds one value, not its entire container.
							// Stop at parameters/imports/functions and uninitialized vars.
							if node != nil && node.Kind == ast.KindVariableDeclaration &&
								node.Name() != nil && node.Name().Kind == ast.KindIdentifier {
								identifiers[reference.Identifier] = node.Initializer()
							}
							break
						}
					}
				}
				for _, target := range targets {
					if target.Kind == ast.KindIdentifier {
						checkObject(identifiers[target])
					} else {
						checkObject(target)
					}
				}
			},
		}
	},
}

func isDefaultPropsName(name string) bool {
	return name == "defaultProps" || name == "getDefaultProps"
}

// mergedNamespaceInitializer checks exports of an intervening namespace before
// falling back to an outer lexical binding. Stop at the scope that already binds
// the reference, so parameters and locals still shadow namespace exports.
func mergedNamespaceInitializer(ctx rule.RuleContext, reference *scope.Reference) (*ast.Node, bool) {
	var boundScope *scope.Scope
	if len(reference.Declarations) != 0 {
		boundScope = reference.Declarations[0].Scope
	}
	for current := reference.From; current != nil && current != boundScope; current = current.Parent {
		if current.Kind != scope.KindModule || current.Block == nil {
			continue
		}
		owner := current.Block.Symbol()
		if owner == nil {
			continue
		}
		symbol := owner.Exports[reference.Identifier.Text()]
		if symbol == nil || symbol.Flags&(ast.SymbolFlagsValue|ast.SymbolFlagsAlias) == 0 {
			continue
		}
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindVariableDeclaration &&
				declaration.Name() != nil && declaration.Name().Kind == ast.KindIdentifier &&
				ast.GetSourceFileOfNode(declaration) == ctx.SourceFile {
				return declaration.Initializer(), true
			}
		}
		return nil, true
	}
	return nil, false
}
