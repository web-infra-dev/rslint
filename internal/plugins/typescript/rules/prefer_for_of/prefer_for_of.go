package prefer_for_of

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var PreferForOfRule = rule.CreateRule(rule.Rule{
	Name:             "prefer-for-of",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindForStatement: func(node *ast.Node) {
				forStmt := node.AsForStatement()
				if forStmt == nil {
					return
				}

				// Step 1: init must be a single variable declaration (let or var, not const),
				// initialized to 0.
				init := forStmt.Initializer
				if init == nil || init.Kind != ast.KindVariableDeclarationList {
					return
				}
				declList := init.AsVariableDeclarationList()
				if declList == nil || declList.Declarations == nil {
					return
				}
				// Must not be const
				if init.Flags&ast.NodeFlagsConst != 0 {
					return
				}
				decls := declList.Declarations.Nodes
				if len(decls) != 1 {
					return
				}
				declarator := decls[0].AsVariableDeclaration()
				if declarator == nil {
					return
				}
				nameNode := declarator.Name()
				if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
					return
				}
				// Must be initialized to 0
				if declarator.Initializer == nil {
					return
				}
				initExpr := ast.SkipParentheses(declarator.Initializer)
				if initExpr.Kind != ast.KindNumericLiteral || initExpr.Text() != "0" {
					return
				}

				indexName := nameNode.AsIdentifier().Text

				// Step 2: test must be `i < arr.length`
				arrayExpression := isLessThanLengthExpression(forStmt.Condition, indexName)
				if arrayExpression == nil {
					return
				}

				// Step 3: update must be an increment pattern
				if !isIncrement(forStmt.Incrementor, indexName) {
					return
				}

				// Step 4: Get symbol for the index variable
				sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
				if sym == nil {
					return
				}

				// Step 5: Check all references of the index variable in the body
				body := forStmt.Statement
				if body == nil {
					return
				}

				arrayText := utils.TrimmedNodeText(ctx.SourceFile, arrayExpression)

				if isIndexOnlyUsedWithArray(ctx, body, sym, arrayText) {
					ctx.ReportNode(node, rule.RuleMessage{
						Id:          "preferForOf",
						Description: "Expected a `for-of` loop instead of a `for` loop with this simple iteration.",
					})
				}
			},
		}
	},
})

// isLessThanLengthExpression checks if node is `i < arr.length` and returns the array expression.
func isLessThanLengthExpression(node *ast.Node, name string) *ast.Node {
	if node == nil || node.Kind != ast.KindBinaryExpression {
		return nil
	}
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return nil
	}
	if binary.OperatorToken.Kind != ast.KindLessThanToken {
		return nil
	}

	// Left must be the index identifier
	left := ast.SkipParentheses(binary.Left)
	if !isMatchingIdentifier(left, name) {
		return nil
	}

	// Right must be arr.length (PropertyAccessExpression with name "length")
	right := ast.SkipParentheses(binary.Right)
	if right.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	propAccess := right.AsPropertyAccessExpression()
	if propAccess == nil {
		return nil
	}
	// Reject optional chaining: arr?.length should not match.
	// In ESTree, optional chains are wrapped in ChainExpression, which causes the
	// pattern match to fail naturally. In tsgo, optional chains are represented by
	// a QuestionDotToken flag on the node, so we must explicitly reject them.
	// Semantically, arr?.length implies arr may be null/undefined, and converting
	// to for-of would throw at runtime in that case.
	if propAccess.QuestionDotToken != nil {
		return nil
	}
	propName := propAccess.Name()
	if propName == nil || propName.Kind != ast.KindIdentifier {
		return nil
	}
	if propName.AsIdentifier().Text != "length" {
		return nil
	}

	return propAccess.Expression
}

// isMatchingIdentifier checks if node is an Identifier with the given name.
func isMatchingIdentifier(node *ast.Node, name string) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	return node.AsIdentifier().Text == name
}

// isLiteralValue checks if a node is a numeric literal with the given value text.
func isLiteralValue(node *ast.Node, valueText string) bool {
	if node == nil {
		return false
	}
	n := ast.SkipParentheses(node)
	return n.Kind == ast.KindNumericLiteral && n.Text() == valueText
}

// isIncrement checks if the update expression is an increment of the named variable.
// Supports: i++, ++i, i += 1, i = i + 1, i = 1 + i
func isIncrement(node *ast.Node, name string) bool {
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindPostfixUnaryExpression:
		postfix := node.AsPostfixUnaryExpression()
		if postfix == nil {
			return false
		}
		// i++
		return postfix.Operator == ast.KindPlusPlusToken && isMatchingIdentifier(ast.SkipParentheses(postfix.Operand), name)

	case ast.KindPrefixUnaryExpression:
		prefix := node.AsPrefixUnaryExpression()
		if prefix == nil {
			return false
		}
		// ++i
		return prefix.Operator == ast.KindPlusPlusToken && isMatchingIdentifier(ast.SkipParentheses(prefix.Operand), name)

	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil {
			return false
		}

		left := ast.SkipParentheses(binary.Left)
		if !isMatchingIdentifier(left, name) {
			return false
		}

		switch binary.OperatorToken.Kind {
		case ast.KindPlusEqualsToken:
			// i += 1
			return isLiteralValue(binary.Right, "1")
		case ast.KindEqualsToken:
			// i = i + 1 or i = 1 + i
			right := ast.SkipParentheses(binary.Right)
			if right.Kind != ast.KindBinaryExpression {
				return false
			}
			addExpr := right.AsBinaryExpression()
			if addExpr == nil || addExpr.OperatorToken == nil || addExpr.OperatorToken.Kind != ast.KindPlusToken {
				return false
			}
			addLeft := ast.SkipParentheses(addExpr.Left)
			addRight := ast.SkipParentheses(addExpr.Right)
			return (isMatchingIdentifier(addLeft, name) && isLiteralValue(addExpr.Right, "1")) ||
				(isLiteralValue(addExpr.Left, "1") && isMatchingIdentifier(addRight, name))
		}
	}

	return false
}

// isIndexOnlyUsedWithArray checks that all references to the index symbol within the
// body are only used as arr[i] (read-only element access on the matching array).
func isIndexOnlyUsedWithArray(ctx rule.RuleContext, body *ast.Node, indexSym *ast.Symbol, arrayText string) bool {
	result := true
	var walk func(n *ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || !result {
			return
		}
		if n.Kind == ast.KindIdentifier {
			sym := ctx.TypeChecker.GetSymbolAtLocation(n)
			if sym == indexSym {
				// This identifier references the index variable.
				// It must be used as arr[i] with matching array text and not as assignee.
				if !isValidIndexUsage(ctx, n, arrayText) {
					result = false
					return
				}
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(body)
	return result
}

// isValidIndexUsage checks if an index variable reference is a valid arr[i] usage:
// - the identifier (possibly wrapped in parentheses) is the ArgumentExpression of an ElementAccessExpression
// - the object's text matches the array text
// - the object is not `this`
// - the ElementAccessExpression is not an assignee
func isValidIndexUsage(ctx rule.RuleContext, id *ast.Node, arrayText string) bool {
	// Walk up through ParenthesizedExpression wrappers.
	// In tsgo, arr[(i)] has ArgumentExpression → ParenthesizedExpression → Identifier,
	// whereas ESTree strips parentheses so the identifier is the direct argument.
	current := id
	for current.Parent != nil && current.Parent.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	parent := current.Parent
	if parent == nil || parent.Kind != ast.KindElementAccessExpression {
		return false
	}
	elemAccess := parent.AsElementAccessExpression()
	if elemAccess == nil {
		return false
	}
	// The argument must be the (possibly parenthesized) identifier
	if elemAccess.ArgumentExpression != current {
		return false
	}
	// The object must not be `this`
	obj := elemAccess.Expression
	if obj == nil {
		return false
	}
	objSkipped := ast.SkipParentheses(obj)
	if objSkipped.Kind == ast.KindThisKeyword {
		return false
	}
	// The object's text must match the array expression text
	objText := utils.TrimmedNodeText(ctx.SourceFile, obj)
	if objText != arrayText {
		return false
	}
	// The element access must not be an assignee
	if isAssignee(parent) {
		return false
	}
	return true
}

// isAssignee checks if a node is used as an assignment target.
// This matches the upstream typescript-eslint isAssignee utility.
// NOTE: This is similar to utils.IsWriteReference but NOT identical:
// - IsWriteReference doesn't handle DeleteExpression or SatisfiesExpression
// - IsWriteReference is designed for identifiers; isAssignee targets member expressions
// - delete semantically removes a property (not a "write"), so adding it to
//   IsWriteReference would change its contract for other rules.
func isAssignee(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent

	switch parent.Kind {
	case ast.KindBinaryExpression:
		// a[i] = 1, a[i] += 1, etc.
		binary := parent.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil {
			switch binary.OperatorToken.Kind {
			case ast.KindEqualsToken,
				ast.KindPlusEqualsToken,
				ast.KindMinusEqualsToken,
				ast.KindAsteriskEqualsToken,
				ast.KindSlashEqualsToken,
				ast.KindPercentEqualsToken,
				ast.KindAsteriskAsteriskEqualsToken,
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
		}

	case ast.KindDeleteExpression:
		// delete a[i]
		del := parent.AsDeleteExpression()
		if del != nil {
			return del.Expression == node
		}

	case ast.KindPostfixUnaryExpression:
		// a[i]++
		postfix := parent.AsPostfixUnaryExpression()
		if postfix != nil && (postfix.Operator == ast.KindPlusPlusToken || postfix.Operator == ast.KindMinusMinusToken) {
			return postfix.Operand == node
		}

	case ast.KindPrefixUnaryExpression:
		// ++a[i]
		prefix := parent.AsPrefixUnaryExpression()
		if prefix != nil && (prefix.Operator == ast.KindPlusPlusToken || prefix.Operator == ast.KindMinusMinusToken) {
			return prefix.Operand == node
		}

	case ast.KindArrayLiteralExpression:
		// [a[i]] = [0]
		return isAssignee(parent)

	case ast.KindSpreadElement:
		// [...a[i]] = [0]
		return isAssignee(parent)

	case ast.KindPropertyAssignment:
		// ({ foo: a[i] }) = { foo: 0 }
		prop := parent.AsPropertyAssignment()
		if prop != nil && prop.Initializer == node {
			gp := parent.Parent
			if gp != nil && gp.Kind == ast.KindObjectLiteralExpression {
				return isAssignee(gp)
			}
		}

	case ast.KindNonNullExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindSatisfiesExpression:
		// (a[i] as number)++, (a[i]! satisfies number)++, etc.
		return isAssignee(parent)

	case ast.KindParenthesizedExpression:
		// (a[i])++ etc.
		return isAssignee(parent)
	}

	return false
}
