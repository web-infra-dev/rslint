package prefer_numeric_literals

import (
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type radixLiteralInfo struct {
	radix         int
	system        string
	literalPrefix string
}

var radixLiteralInfos = map[int]radixLiteralInfo{
	2:  {radix: 2, system: "binary", literalPrefix: "0b"},
	8:  {radix: 8, system: "octal", literalPrefix: "0o"},
	16: {radix: 16, system: "hexadecimal", literalPrefix: "0x"},
}

func useLiteralMessage(system string, functionName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useLiteral",
		Description: "Use " + system + " literals instead of " + functionName + "().",
		Data: map[string]string{
			"system":       system,
			"functionName": functionName,
		},
	}
}

// https://eslint.org/docs/latest/rules/prefer-numeric-literals
var PreferNumericLiteralsRule = rule.Rule{
	Name: "prefer-numeric-literals",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) != 2 {
					return
				}

				args := call.Arguments.Nodes
				// ESTree treats parentheses as transparent here, but TS-only
				// assertion wrappers must remain opaque to match ESLint.
				str, ok := utils.GetStaticStringLiteralValue(ast.SkipParentheses(args[0]))
				if !ok {
					return
				}

				radixInfo, ok := numericRadixInfo(args[1])
				if !ok {
					return
				}

				if !utils.IsGlobalParseIntCallee(call.Expression, ctx.Globals) {
					return
				}

				functionName := utils.TrimmedNodeText(ctx.SourceFile, ast.SkipParentheses(call.Expression))
				msg := useLiteralMessage(radixInfo.system, functionName)
				if fix := numericLiteralFix(ctx, node, str, radixInfo); fix != nil {
					ctx.ReportNodeWithFixes(node, msg, *fix)
					return
				}
				ctx.ReportNode(node, msg)
			},
		}
	},
}

func numericRadixInfo(node *ast.Node) (radixLiteralInfo, bool) {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return radixLiteralInfo{}, false
	}

	text := utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text)
	value, err := strconv.Atoi(text)
	if err != nil {
		return radixLiteralInfo{}, false
	}

	info, ok := radixLiteralInfos[value]
	return info, ok
}

func numericLiteralFix(ctx rule.RuleContext, node *ast.Node, str string, radixInfo radixLiteralInfo) *rule.RuleFix {
	if utils.HasCommentInsideNode(ctx.SourceFile, node) {
		return nil
	}
	if !canRepresentAsNumericLiteral(str, radixInfo.radix) {
		return nil
	}

	replacement := radixInfo.literalPrefix + str
	fix := rule.RuleFixReplace(
		ctx.SourceFile,
		node,
		utils.SafeReplacementText(ctx.SourceFile, node, replacement),
	)
	return &fix
}

// parseInt accepts prefixes, signs, separators, whitespace, and then stops at
// the first invalid character. Numeric literals need the whole string to be a
// valid digit sequence for the target radix, so this safety check stays
// rule-specific rather than using a general static-value helper.
func canRepresentAsNumericLiteral(str string, radix int) bool {
	if str == "" {
		return false
	}
	for _, r := range str {
		switch {
		case r >= '0' && r <= '9':
			if int(r-'0') >= radix {
				return false
			}
		case radix == 16 && r >= 'a' && r <= 'f':
		case radix == 16 && r >= 'A' && r <= 'F':
		default:
			return false
		}
	}
	return true
}
