package forward_ref_uses_ref

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react/reactutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const missingRefParameterText = "forwardRef is used with this component but no ref parameter is set"

func missingRefParameterMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingRefParameter",
		Description: missingRefParameterText,
	}
}

func addRefParameterMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "addRefParameter",
		Description: "Add a ref parameter",
	}
}

func removeForwardRefMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeForwardRef",
		Description: "Remove forwardRef wrapper",
	}
}

func isForwardRefIdentifier(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	return node != nil && node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == "forwardRef"
}

// isForwardRefCall mirrors upstream's narrow textual callee check:
// `forwardRef(...)` or any member call whose property is named `forwardRef`.
// It deliberately ignores imports, the object name, and the argument position,
// so `Other.forwardRef(fn)` and `forwardRef(Component, fn)` match too. Only
// parentheses are transparent; TS-only wrappers stay visible to match upstream.
func isForwardRefCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if isForwardRefIdentifier(callee) {
		return true
	}
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := callee.AsPropertyAccessExpression().Name()
	return isForwardRefIdentifier(name)
}

func parentForwardRefCall(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		parent = parent.Parent
	}
	if !isForwardRefCall(parent) {
		return nil
	}
	return parent
}

func addRefParameterSuggestion(ctx rule.RuleContext, fn *ast.Node, param *ast.Node) rule.RuleSuggestion {
	fixes := []rule.RuleFix{}
	if utils.IsParenlessArrowFunction(fn) {
		fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, param, "("))
		fixes = append(fixes, rule.RuleFixInsertAfter(param, ", ref)"))
	} else {
		fixes = append(fixes, rule.RuleFixInsertAfter(param, ", ref"))
	}
	return rule.RuleSuggestion{
		Message:  addRefParameterMessage(),
		FixesArr: fixes,
	}
}

func removeForwardRefSuggestion(ctx rule.RuleContext, call *ast.Node, fn *ast.Node) rule.RuleSuggestion {
	return rule.RuleSuggestion{
		Message: removeForwardRefMessage(),
		FixesArr: []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, call, utils.TrimmedNodeText(ctx.SourceFile, fn)),
		},
	}
}

// https://github.com/jsx-eslint/eslint-plugin-react/blob/master/docs/rules/forward-ref-uses-ref.md
var ForwardRefUsesRefRule = rule.Rule{
	Name: "react/forward-ref-uses-ref",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkFunction := func(node *ast.Node) {
			call := parentForwardRefCall(node)
			if call == nil {
				return
			}
			params := reactutil.FunctionParameters(node)
			if len(params) != 1 {
				return
			}
			param := params[0]
			ctx.ReportNodeWithSuggestions(
				node,
				missingRefParameterMessage(),
				addRefParameterSuggestion(ctx, node, param),
				removeForwardRefSuggestion(ctx, call, node),
			)
		}

		return rule.RuleListeners{
			ast.KindArrowFunction:      checkFunction,
			ast.KindFunctionExpression: checkFunction,
		}
	},
}
