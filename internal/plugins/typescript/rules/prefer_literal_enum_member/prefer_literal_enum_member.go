package prefer_literal_enum_member

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
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

type PreferLiteralEnumMemberOptions struct {
	AllowBitwiseExpressions *bool `json:"allowBitwiseExpressions"`
}

var defaultOptions = PreferLiteralEnumMemberOptions{
	AllowBitwiseExpressions: utils.Ref(false),
}

var PreferLiteralEnumMemberRule = rule.CreateRule(rule.Rule{
	Name: "prefer-literal-enum-member",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := defaultOptions

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
					opts.AllowBitwiseExpressions = utils.Ref(allowBitwiseExpressions)
				}
			}
		}

		allowBitwise := opts.AllowBitwiseExpressions != nil && *opts.AllowBitwiseExpressions

		return rule.RuleListeners{
			ast.KindEnumDeclaration: func(node *ast.Node) {
				enumDecl := node.AsEnumDeclaration()
				if enumDecl == nil || enumDecl.Members == nil {
					return
				}

				// Collect member names for self-referencing check
				memberNames := make(map[string]bool)
				for _, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil {
						continue
					}
					name := getMemberName(member)
					if name != "" {
						memberNames[name] = true
					}
				}

				enumName := ""
				if enumDecl.Name() != nil {
					enumName = enumDecl.Name().Text()
				}

				for _, memberNode := range enumDecl.Members.Nodes {
					member := memberNode.AsEnumMember()
					if member == nil || member.Initializer == nil {
						continue
					}

					if !isAllowedInitializer(member.Initializer, allowBitwise, false, enumName, memberNames) {
						msg := buildNotLiteralMessage()
						if allowBitwise {
							msg = buildNotLiteralOrBitwiseExpressionMessage()
						}
						ctx.ReportNode(member.Name(), msg)
					}
				}
			},
		}
	},
})

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

func isAllowedInitializer(expr *ast.Node, allowBitwise bool, partOfBitwiseComputation bool, enumName string, memberNames map[string]bool) bool {
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
			return isAllowedInitializer(unary.Operand, allowBitwise, partOfBitwiseComputation, enumName, memberNames)
		case ast.KindTildeToken:
			if !allowBitwise {
				return false
			}
			return isAllowedInitializer(unary.Operand, allowBitwise, true, enumName, memberNames)
		}
		return false

	case ast.KindBinaryExpression:
		if !allowBitwise {
			return false
		}
		binary := expr.AsBinaryExpression()
		if binary == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandToken, ast.KindBarToken, ast.KindCaretToken,
			ast.KindLessThanLessThanToken, ast.KindGreaterThanGreaterThanToken, ast.KindGreaterThanGreaterThanGreaterThanToken:
			return isAllowedInitializer(binary.Left, allowBitwise, true, enumName, memberNames) &&
				isAllowedInitializer(binary.Right, allowBitwise, true, enumName, memberNames)
		}
		return false

	case ast.KindIdentifier:
		// Self-referencing enum member is only allowed in bitwise context
		if partOfBitwiseComputation {
			name := expr.Text()
			return memberNames[name]
		}
		return false

	case ast.KindPropertyAccessExpression:
		// Foo.A or Foo['A'] style self-references, only in bitwise context
		if partOfBitwiseComputation {
			propAccess := expr.AsPropertyAccessExpression()
			if propAccess == nil {
				return false
			}
			obj := propAccess.Expression
			if obj == nil || obj.Kind != ast.KindIdentifier || obj.Text() != enumName {
				return false
			}
			memberName := propAccess.Name().Text()
			return memberNames[memberName]
		}
		return false

	case ast.KindElementAccessExpression:
		// Foo['A'] style self-references, only in bitwise context
		if partOfBitwiseComputation {
			elemAccess := expr.AsElementAccessExpression()
			if elemAccess == nil {
				return false
			}
			obj := elemAccess.Expression
			if obj == nil || obj.Kind != ast.KindIdentifier || obj.Text() != enumName {
				return false
			}
			arg := elemAccess.ArgumentExpression
			if arg == nil || arg.Kind != ast.KindStringLiteral {
				return false
			}
			memberName := arg.AsStringLiteral().Text
			return memberNames[memberName]
		}
		return false

	case ast.KindParenthesizedExpression:
		paren := expr.AsParenthesizedExpression()
		if paren == nil {
			return false
		}
		return isAllowedInitializer(paren.Expression, allowBitwise, partOfBitwiseComputation, enumName, memberNames)
	}

	return false
}
