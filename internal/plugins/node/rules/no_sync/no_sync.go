package no_sync

import (
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_sync.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-sync.js
var NoSyncRule = rule.Rule{
	Name:   "node/no-sync",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		allowAtRootLevel := false
		var ignores []any
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok {
				allowAtRootLevel, _ = option["allowAtRootLevel"].(bool)
				ignores, _ = option["ignores"].([]any)
			}
		}
		// Declaration origins are stable for a type throughout this file's run.
		// Keep name matching outside the cache: intrinsic types may be shared by
		// identifiers with different names.
		var declarationMatches []map[*checker.Type]bool
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if !strings.HasSuffix(node.Text(), "Sync") || !matchesCall(node, allowAtRootLevel) {
					return
				}
				name := node.Text()
				var typ *checker.Type
				for index, ignore := range ignores {
					if text, ok := ignore.(string); ok {
						if text == node.Text() {
							return
						}
						continue
					}
					if ctx.TypeChecker == nil {
						continue
					}
					if typ == nil {
						typ = utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
						if fullName := fullTypeName(typ); fullName != "" {
							name = fullName
						}
					}
					specifier, ok := utils.ParseTypeOrValueSpecifier(ignore)
					fields, _ := ignore.(map[string]any)
					if !ok || fields["from"] == nil || fields["name"] != nil && !slices.Contains(specifier.Name, name) {
						continue
					}
					if declarationMatches == nil {
						declarationMatches = make([]map[*checker.Type]bool, len(ignores))
					}
					matches, cached := declarationMatches[index][typ]
					if !cached {
						matches = utils.TypeMatchesDeclarationSpecifier(typ, specifier, ctx.Program())
						if declarationMatches[index] == nil {
							declarationMatches[index] = make(map[*checker.Type]bool)
						}
						declarationMatches[index][typ] = matches
					}
					if matches {
						return
					}
				}
				parent := utils.ESTreeParent(node)
				message := rule.RuleMessage{
					Id:          "noSync",
					Description: "Unexpected sync method: '" + name + "'.",
				}
				ctx.ReportRange(parentRange(ctx.SourceFile, node), message)
				// ESTree visits a shorthand property's key and value separately.
				if parent.Name() == node && (parent.Kind == ast.KindShorthandPropertyAssignment ||
					parent.Kind == ast.KindBindingElement && parent.Parent.Kind == ast.KindObjectBindingPattern &&
						parent.AsBindingElement().PropertyName == nil && parent.AsBindingElement().DotDotDotToken == nil) {
					ctx.ReportNode(parent, message)
				}
			},
		}
	},
}

// Upstream selects direct identifier children of calls (including arguments),
// and every identifier below a member-expression callee, not just its property.
func matchesCall(node *ast.Node, allowAtRootLevel bool) bool {
	if utils.IsInJsxTagName(node) || node.Parent.Kind == ast.KindJsxAttribute || node.Parent.Kind == ast.KindJsxNamespacedName ||
		node.Parent.Kind == ast.KindPropertyAccessExpression && utils.IsInJsxTagName(node.Parent) {
		return false
	}
	parent := utils.ESTreeParent(node)
	matched := false
	for child, ancestor := node, parent; ancestor != nil; child, ancestor = ancestor, utils.ESTreeParent(ancestor) {
		if utils.IsJSDocSyntaxNode(ancestor) {
			return false
		}
		if ancestor.Kind == ast.KindCallExpression {
			object, _ := utils.MemberExpressionParts(child)
			if ancestor == parent || object != nil && utils.ESTreeCallCallee(ancestor.Expression()) == child {
				matched = true
				if !allowAtRootLevel {
					return true
				}
			}
		}
		// A method's computed key is outside its ESTree function value.
		if matched && utils.IsFunctionLikeContainer(ancestor) && child != ancestor.Name() && child.Kind != ast.KindDecorator && ancestor.Body() != nil {
			return true
		}
	}
	return false
}

// Parameters and binding elements are wrappers in tsgo, while ESTree attaches
// their identifiers directly to a function, pattern, or default assignment.
func parentRange(source *ast.SourceFile, node *ast.Node) core.TextRange {
	parent := utils.ESTreeParent(node)
	switch parent.Kind {
	case ast.KindComputedPropertyName:
		return utils.TrimNodeTextRange(source, parent.Parent)
	case ast.KindParameter:
		parameter := parent.AsParameterDeclaration()
		if parameter.Initializer != nil {
			return core.NewTextRange(utils.TrimNodeTextRange(source, parent.Name()).Pos(), parent.End())
		}
		if parameter.DotDotDotToken == nil && parent.Modifiers() == nil {
			return utils.ESTreeFunctionRange(source, parent.Parent)
		}
	case ast.KindBindingElement:
		binding := parent.AsBindingElement()
		if binding.Initializer != nil && binding.PropertyName != node {
			return core.NewTextRange(utils.TrimNodeTextRange(source, parent.Name()).Pos(), parent.End())
		}
		if binding.DotDotDotToken == nil && parent.Parent.Kind == ast.KindArrayBindingPattern {
			return utils.TrimNodeTextRange(source, parent.Parent)
		}
	}
	return utils.TrimNodeTextRange(source, parent)
}

func fullTypeName(typ *checker.Type) string {
	var names []string
	for symbol := checker.Type_symbol(typ); symbol != nil; symbol = symbol.Parent {
		if declaration := symbol.ValueDeclaration; declaration != nil &&
			(declaration.Kind == ast.KindSourceFile || declaration.Kind == ast.KindModuleDeclaration) {
			break
		}
		names = append(names, ast.EscapeInternalSymbolName(ast.SymbolName(symbol)))
	}
	slices.Reverse(names)
	return strings.Join(names, ".")
}
