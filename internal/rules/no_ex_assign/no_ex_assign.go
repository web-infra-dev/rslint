package no_ex_assign

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// Message builder
func buildExAssignMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpected",
		Description: "Do not assign to the exception parameter.",
	}
}

func collectCatchBindingNamesAndSymbols(name *ast.Node, ctx rule.RuleContext) ([]string, []*ast.Symbol) {
	if name == nil {
		return nil, nil
	}
	if ast.IsIdentifier(name) {
		return []string{name.Text()}, []*ast.Symbol{ctx.TypeChecker.GetSymbolAtLocation(name)}
	}
	if ast.IsBindingPattern(name) {
		var names []string
		var symbols []*ast.Symbol
		for _, elem := range name.Elements() {
			if elem == nil || !ast.IsBindingElement(elem) {
				continue
			}
			be := elem.AsBindingElement()
			if be == nil || be.Name() == nil {
				continue
			}
			names = append(names, be.Name().Text())
			symbols = append(symbols, ctx.TypeChecker.GetSymbolAtLocation(be.Name()))
		}
		return names, symbols
	}
	return nil, nil
}

func isBindingPatternInAssignment(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent

	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}

	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}

	binary := parent.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return false
	}

	switch binary.OperatorToken.Kind {
	case ast.KindEqualsToken:
		return binary.Left == node
	}

	return false
}

func isInDestructuringAssignment(node *ast.Node) bool {
	if node == nil {
		return false
	}

	parent := node.Parent

	for parent != nil {
		if parent.Kind == ast.KindObjectLiteralExpression || parent.Kind == ast.KindArrayLiteralExpression {
			return true
		}
		if parent.Kind == ast.KindBinaryExpression {
			binary := parent.AsBinaryExpression()
			if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
				return binary.Left == node
			}
		}
		if parent.Kind == ast.KindParenthesizedExpression {
			parent = parent.Parent
		} else {
			break
		}
		parent = parent.Parent
	}

	return false
}

func isWriteReference(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}

	parent := node.Parent

	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}

		switch binary.OperatorToken.Kind {
		case ast.KindEqualsToken,
			ast.KindPlusEqualsToken,
			ast.KindMinusEqualsToken,
			ast.KindAsteriskAsteriskEqualsToken,
			ast.KindAsteriskEqualsToken,
			ast.KindSlashEqualsToken,
			ast.KindPercentEqualsToken,
			ast.KindLessThanLessThanEqualsToken,
			ast.KindGreaterThanGreaterThanEqualsToken,
			ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
			ast.KindAmpersandEqualsToken,
			ast.KindBarEqualsToken,
			ast.KindCaretEqualsToken,
			ast.KindBarBarEqualsToken,
			ast.KindAmpersandAmpersandEqualsToken,
			ast.KindQuestionQuestionEqualsToken:
			return binary.Left == node
		}
	case ast.KindPostfixUnaryExpression:
		postfix := parent.AsPostfixUnaryExpression()
		if postfix == nil {
			return false
		}
		switch postfix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return postfix.Operand == node
		}
	case ast.KindPrefixUnaryExpression:
		prefix := parent.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		switch prefix.Operator {
		case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
			return prefix.Operand == node
		}
	case ast.KindObjectBindingPattern:
		return isBindingPatternInAssignment(parent)
	case ast.KindArrayBindingPattern:
		return isBindingPatternInAssignment(parent)
	case ast.KindBindingElement:
		return isWriteReference(parent)
	case ast.KindShorthandPropertyAssignment:
		return isInDestructuringAssignment(parent)
	case ast.KindPropertyAssignment:
		return isInDestructuringAssignment(parent)
	case ast.KindObjectLiteralExpression:
		return isInDestructuringAssignment(parent)
	case ast.KindArrayLiteralExpression:
		return isInDestructuringAssignment(parent)
	case ast.KindParenthesizedExpression:
		return isWriteReference(parent)
	case ast.KindAsExpression, ast.KindTypeAssertionExpression:
		return isWriteReference(parent)
	}

	return false
}

func isNameShadowed(node *ast.Node, symbols []*ast.Symbol, ctx rule.RuleContext) bool {
	if node == nil || ctx.TypeChecker == nil || len(symbols) == 0 {
		return false
	}

	symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
	if symbol == nil {
		return false
	}

	for _, s := range symbols {
		if s == symbol {
			return false
		}
	}
	return true
}

func getIdentifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}
	return node.Text()
}

func checkReassignments(block *ast.Node, names []string, symbols []*ast.Symbol, ctx rule.RuleContext) {
	if block == nil || ctx.TypeChecker == nil || len(names) == 0 || len(symbols) == 0 {
		return
	}

	var walk func(*ast.Node)
	walk = func(block *ast.Node) {
		if block == nil {
			return
		}

		block.ForEachChild(func(child *ast.Node) bool {
			if child == nil {
				return false
			}

			childName := getIdentifierName(child)
			if child.Kind == ast.KindIdentifier && slices.Contains(names, childName) {
				if isWriteReference(child) {
					if !isNameShadowed(child, symbols, ctx) {
						ctx.ReportNode(child, buildExAssignMessage())
					}
				}
			} else {
				walk(child)
			}

			return false
		})
	}

	walk(block)
}

var NoExAssignRule = rule.Rule{
	Name: "no-ex-assign",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCatchClause: func(node *ast.Node) {
				if node.AsCatchClause().VariableDeclaration == nil {
					return
				}

				varDecl := node.AsCatchClause().VariableDeclaration.AsVariableDeclaration()
				if varDecl == nil || varDecl.Name() == nil {
					return
				}

				block := node.AsCatchClause().Block
				if block == nil {
					return
				}

				names, symbols := collectCatchBindingNamesAndSymbols(varDecl.Name(), ctx)
				checkReassignments(block, names, symbols, ctx)
			},
		}
	},
}
