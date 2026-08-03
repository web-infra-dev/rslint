package prefer_literal_enum_member

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildNotLiteralMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notLiteral",
		Description: "Explicit enum value must only be a literal value (string or number).",
	}
}

func buildNotLiteralOrBitwiseExpressionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "notLiteralOrBitwiseExpression",
		Description: "Explicit enum value must only be a literal value (string or number) or a bitwise expression.",
	}
}

var PreferLiteralEnumMemberRule = rule.CreateRule(rule.Rule{
	Name: "prefer-literal-enum-member",
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		allowBitwise := false

		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if allowBitwiseExpressions, ok := optsMap["allowBitwiseExpressions"].(bool); ok {
					allowBitwise = allowBitwiseExpressions
				}
			}
		}

		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Members == nil {
					return
				}

				// Most enums contain only literal expressions. Delay indexing member
				// names until a bitwise expression actually references one.
				var enumMembers *enumMemberNames
				if allowBitwise {
					enumName := ""
					if name := enumDecl.Name(); name != nil {
						enumName = name.Text()
					}
					enumMembers = &enumMemberNames{
						enumName: enumName,
						members:  enumDecl.Members.Nodes,
					}
				}

				for _, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil || member.Initializer == nil {
						continue
					}

					if !isAllowedInitializer(member.Initializer, false, enumMembers) {
						message := buildNotLiteralMessage()
						if allowBitwise {
							message = buildNotLiteralOrBitwiseExpressionMessage()
						}
						ctx.ReportNode(member.Name(), message)
					}
				}
			},
		}
	},
})

type enumMemberNames struct {
	enumName    string
	members     []*ast.Node
	names       map[string]struct{}
	initialized bool
}

func (e *enumMemberNames) contains(name string) bool {
	if !e.initialized {
		e.initialized = true
		e.names = make(map[string]struct{}, len(e.members))
		for _, memberNode := range e.members {
			member := memberNode.AsEnumMember()
			if member == nil {
				continue
			}
			if memberName := getMemberName(member); memberName != "" {
				e.names[memberName] = struct{}{}
			}
		}
	}
	_, ok := e.names[name]
	return ok
}

func getMemberName(member *ast.EnumMember) string {
	name := member.Name()
	if name == nil {
		return ""
	}
	switch name.Kind {
	case ast.KindIdentifier:
		return name.Text()
	case ast.KindStringLiteral:
		return name.AsStringLiteral().Text
	case ast.KindComputedPropertyName:
		// For computed properties like ['a'], check if it's a string literal inside
		computed := name.AsComputedPropertyName()
		if computed != nil && computed.Expression != nil && computed.Expression.Kind == ast.KindStringLiteral {
			return computed.Expression.AsStringLiteral().Text
		}
	}
	return ""
}

func isAllowedInitializer(expr *ast.Node, partOfBitwiseComputation bool, enumMembers *enumMemberNames) bool {
	if expr == nil {
		return false
	}

	switch expr.Kind {
	case ast.KindNumericLiteral, ast.KindStringLiteral, ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindRegularExpressionLiteral:
		return true

	case ast.KindNoSubstitutionTemplateLiteral:
		// Plain template literal with no expressions is allowed
		return true

	case ast.KindTemplateExpression:
		// Template literal with expressions is NOT allowed
		return false

	case ast.KindPrefixUnaryExpression:
		unary := expr.AsPrefixUnaryExpression()
		if unary == nil {
			return false
		}
		switch unary.Operator {
		case ast.KindMinusToken, ast.KindPlusToken:
			return isAllowedInitializer(unary.Operand, partOfBitwiseComputation, enumMembers)
		case ast.KindTildeToken:
			if enumMembers == nil {
				return false
			}
			return isAllowedInitializer(unary.Operand, true, enumMembers)
		}
		return false

	case ast.KindBinaryExpression:
		if enumMembers == nil {
			return false
		}
		binary := expr.AsBinaryExpression()
		if binary == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandToken, ast.KindBarToken, ast.KindCaretToken,
			ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken, ast.KindGreaterThanGreaterThanGreaterThanToken:
			return isAllowedInitializer(binary.Left, true, enumMembers) &&
				isAllowedInitializer(binary.Right, true, enumMembers)
		}
		return false

	case ast.KindIdentifier:
		// Self-referencing enum member is only allowed in bitwise context
		if partOfBitwiseComputation && enumMembers != nil {
			name := expr.Text()
			return enumMembers.contains(name)
		}
		return false

	case ast.KindPropertyAccessExpression:
		// Foo.A or Foo['A'] style self-references, only in bitwise context
		if partOfBitwiseComputation && enumMembers != nil {
			propAccess := expr.AsPropertyAccessExpression()
			if propAccess == nil {
				return false
			}
			obj := propAccess.Expression
			if obj == nil || obj.Kind != ast.KindIdentifier || obj.Text() != enumMembers.enumName {
				return false
			}
			memberName := propAccess.Name().Text()
			return enumMembers.contains(memberName)
		}
		return false

	case ast.KindElementAccessExpression:
		// Foo['A'] style self-references, only in bitwise context
		if partOfBitwiseComputation && enumMembers != nil {
			elemAccess := expr.AsElementAccessExpression()
			if elemAccess == nil {
				return false
			}
			obj := elemAccess.Expression
			if obj == nil || obj.Kind != ast.KindIdentifier || obj.Text() != enumMembers.enumName {
				return false
			}
			arg := elemAccess.ArgumentExpression
			if arg == nil || arg.Kind != ast.KindStringLiteral {
				return false
			}
			memberName := arg.AsStringLiteral().Text
			return enumMembers.contains(memberName)
		}
		return false

	case ast.KindParenthesizedExpression:
		paren := expr.AsParenthesizedExpression()
		if paren == nil {
			return false
		}
		return isAllowedInitializer(paren.Expression, partOfBitwiseComputation, enumMembers)
	}

	return false
}
