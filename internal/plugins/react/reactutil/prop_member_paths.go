package reactutil

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// ComponentScope returns the enclosing block, namespace, or source file.
func ComponentScope(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock:
			return current
		}
	}
	return nil
}

// StaticPropertyName returns a statically known property name.
func StaticPropertyName(node *ast.Node) string {
	if node == nil {
		return ""
	}
	name, _ := utils.GetStaticPropertyName(node)
	return name
}

// PropMemberNames projects a prop usage path, retaining markers for dynamic keys.
func PropMemberNames(node *ast.Node) (*ast.Node, []string, bool) {
	return propMemberNamesWithMode(node, false, false)
}

// PropAssignmentNames projects property.name paths used by upstream propTypes assignments.
func PropAssignmentNames(node *ast.Node) (*ast.Node, []string, bool) {
	return propMemberNamesWithMode(node, true, false)
}

// StaticPropMemberNames resolves only statically named members, retaining cooked
// literal keys and rejecting dynamic accesses. It removes expression wrappers.
func StaticPropMemberNames(node *ast.Node) (*ast.Node, []string, bool) {
	return propMemberNamesWithMode(node, false, true)
}

func propMemberNamesWithMode(node *ast.Node, estreePropertyNames, staticKeys bool) (*ast.Node, []string, bool) {
	var names []string
	current := SkipExpressionWrappers(node)
	var report *ast.Node
	for current != nil {
		switch current.Kind {
		case ast.KindPropertyAccessExpression:
			access := current.AsPropertyAccessExpression()
			name := StaticPropertyName(access.Name())
			if name == "" {
				return nil, nil, false
			}
			names = append(names, name)
			if report == nil {
				report = access.Name()
			}
			current = SkipExpressionWrappers(access.Expression)
		case ast.KindElementAccessExpression:
			access := current.AsElementAccessExpression()
			argument := SkipExpressionWrappers(access.ArgumentExpression)
			name := ""
			if staticKeys {
				if argument == nil || argument.Kind == ast.KindIdentifier {
					return nil, nil, false
				}
				var ok bool
				if argument.Kind == ast.KindNoSubstitutionTemplateLiteral {
					name, ok = argument.Text(), true
				} else {
					name, ok = utils.GetStaticPropertyName(argument)
				}
				if !ok {
					return nil, nil, false
				}
			} else if estreePropertyNames {
				argument = ast.SkipParentheses(access.ArgumentExpression)
				if argument == nil || argument.Kind != ast.KindIdentifier {
					return nil, nil, false
				}
				name = argument.AsIdentifier().Text
			} else {
				name = propElementName(access.ArgumentExpression)
				if argument != nil && argument.Kind == ast.KindNumericLiteral {
					name = "__NUMERIC_PROP__:" + name
				}
				if name == "" {
					if argument != nil && argument.Kind == ast.KindBigIntLiteral {
						return nil, nil, false
					}
					name = "__COMPUTED_PROP__"
				}
			}
			names = append(names, name)
			if report == nil {
				report = access.ArgumentExpression
			}
			current = SkipExpressionWrappers(access.Expression)
		default:
			return current, finishPropMemberNames(names, estreePropertyNames || staticKeys), true
		}
	}
	return report, finishPropMemberNames(names, estreePropertyNames || staticKeys), false
}

func finishPropMemberNames(names []string, preserveNames bool) []string {
	slices.Reverse(names)
	if preserveNames {
		return names
	}
	for index, name := range names {
		if !strings.HasPrefix(name, "__NUMERIC_PROP__:") {
			continue
		}
		if index == 0 {
			names[index] = strings.TrimPrefix(name, "__NUMERIC_PROP__:")
		} else {
			names[index] = "__COMPUTED_PROP__"
		}
	}
	return names
}

func propElementName(node *ast.Node) string {
	node = SkipExpressionWrappers(node)
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindStringLiteral,
		ast.KindNumericLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNullKeyword,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindRegularExpressionLiteral:
		return StaticPropertyName(node)
	}
	return ""
}

// PropMemberIdentifier projects a member using ESTree property.name semantics.
func PropMemberIdentifier(node *ast.Node) (*ast.Node, string, bool) {
	node = SkipExpressionWrappers(node)
	if node == nil {
		return nil, "", false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		member := node.AsPropertyAccessExpression()
		return SkipExpressionWrappers(member.Expression), EsTreeName(member.Name()), true
	case ast.KindElementAccessExpression:
		member := node.AsElementAccessExpression()
		return SkipExpressionWrappers(member.Expression), EsTreeName(ast.SkipParentheses(member.ArgumentExpression)), true
	}
	return nil, "", false
}
