package no_null

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	messageIDError   = "error"
	messageIDReplace = "replace"
	messageIDRemove  = "remove"
)

var (
	errorMessage = rule.RuleMessage{
		Id:          messageIDError,
		Description: "Use `undefined` instead of `null`.",
	}
	replaceMessage = rule.RuleMessage{
		Id:          messageIDReplace,
		Description: "Replace `null` with `undefined`.",
	}
	removeMessage = rule.RuleMessage{
		Id:          messageIDRemove,
		Description: "Remove `null`.",
	}
)

//go:embed no_null.schema.json
var schemaJSON []byte

type options struct {
	checkArguments      bool
	checkStrictEquality bool
}

func parseOptions(raw []any) options {
	result := options{checkArguments: true}
	if len(raw) == 0 {
		return result
	}

	values, _ := raw[0].(map[string]any)
	if checkArguments, ok := values["checkArguments"].(bool); ok {
		result.checkArguments = checkArguments
	}
	if checkStrictEquality, ok := values["checkStrictEquality"].(bool); ok {
		result.checkStrictEquality = checkStrictEquality
	}
	return result
}

func equalityKind(node *ast.Node) ast.Kind {
	parent := utils.ESTreeParent(node)
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return ast.KindUnknown
	}
	return parent.AsBinaryExpression().OperatorToken.Kind
}

func directCallOrNewArgumentParent(node *ast.Node) *ast.Node {
	parent := utils.ESTreeParent(node)
	if parent == nil || ast.IsImportCall(parent) ||
		(!ast.IsCallExpression(parent) && !ast.IsNewExpression(parent)) {
		return nil
	}
	for _, argument := range parent.Arguments() {
		if utils.ESTreeRuntimeExpression(argument) == node {
			return parent
		}
	}
	return nil
}

func isIdentifierNamed(node *ast.Node, name string) bool {
	node = utils.ESTreeRuntimeExpression(node)
	return node != nil && ast.IsIdentifier(node) && node.AsIdentifier().Text == name
}

func isExemptObjectCreate(call *ast.Node, nullNode *ast.Node) bool {
	minimumArguments := 1
	maximumArguments := 2
	match, ok := unicornutil.MatchDotMethodCall(call, unicornutil.DotMethodCallOptions{
		Method:              "create",
		MinimumArguments:    &minimumArguments,
		MaximumArguments:    &maximumArguments,
		RejectSpreadElement: true,
		AllowOptionalCall:   false,
		AllowOptionalMember: false,
	})
	return ok && isIdentifierNamed(match.Object, "Object") &&
		utils.ESTreeRuntimeExpression(call.Arguments()[0]) == nullNode
}

func isExemptUseRef(call *ast.Node, nullNode *ast.Node) bool {
	if !ast.IsCallExpression(call) || len(call.Arguments()) != 1 ||
		utils.ESTreeRuntimeExpression(call.Arguments()[0]) != nullNode {
		return false
	}

	// Upstream's identifier-call check does not consume its optionalCall and
	// optionalMember fields, so `useRef?.(null)` receives the same exception.
	if isIdentifierNamed(call.AsCallExpression().Expression, "useRef") {
		return true
	}

	argumentsLength := 1
	match, ok := unicornutil.MatchDotMethodCall(call, unicornutil.DotMethodCallOptions{
		Method:              "useRef",
		ArgumentsLength:     &argumentsLength,
		RejectSpreadElement: true,
		AllowOptionalCall:   false,
		AllowOptionalMember: false,
	})
	return ok && isIdentifierNamed(match.Object, "React")
}

func isExemptInsertBefore(call *ast.Node, nullNode *ast.Node) bool {
	argumentsLength := 2
	match, ok := unicornutil.MatchDotMethodCall(call, unicornutil.DotMethodCallOptions{
		Method:              "insertBefore",
		ArgumentsLength:     &argumentsLength,
		RejectSpreadElement: true,
		AllowOptionalCall:   false,
		AllowOptionalMember: true,
	})
	return ok && utils.ESTreeRuntimeExpression(match.Call.Arguments()[1]) == nullNode
}

func replacementSuggestion(ctx rule.RuleContext, node *ast.Node) rule.RuleSuggestion {
	return rule.RuleSuggestion{
		Message:  replaceMessage,
		FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "undefined")},
	}
}

func variableIDEnd(ctx rule.RuleContext, declaration *ast.VariableDeclaration) int {
	name := declaration.Name()
	if name == nil {
		return 0
	}
	if ast.IsIdentifier(name) {
		return utils.GetESTreeBindingIdentifierRange(ctx.SourceFile, name).End()
	}
	if declaration.Type != nil {
		return declaration.Type.End()
	}
	return utils.TrimNodeTextRange(ctx.SourceFile, name).End()
}

func reportNull(ctx rule.RuleContext, node *ast.Node, opts options) {
	// TSESTree represents `null` in a type position as TSNullKeyword rather
	// than the runtime Literal node that this rule listens for upstream.
	if node.Parent != nil && node.Parent.Kind == ast.KindLiteralType {
		return
	}

	operator := equalityKind(node)
	if !opts.checkStrictEquality &&
		(operator == ast.KindEqualsEqualsEqualsToken || operator == ast.KindExclamationEqualsEqualsToken) {
		return
	}

	argumentParent := directCallOrNewArgumentParent(node)
	if argumentParent != nil {
		if !opts.checkArguments ||
			isExemptObjectCreate(argumentParent, node) ||
			isExemptUseRef(argumentParent, node) ||
			isExemptInsertBefore(argumentParent, node) {
			return
		}
	}

	if operator == ast.KindEqualsEqualsToken || operator == ast.KindExclamationEqualsToken {
		ctx.ReportNodeWithDeferredFixes(node, errorMessage, func() []rule.RuleFix {
			return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, "undefined")}
		})
		return
	}

	parent := utils.ESTreeParent(node)
	if parent != nil && parent.Kind == ast.KindReturnStatement &&
		utils.ESTreeRuntimeExpression(parent.AsReturnStatement().Expression) == node {
		ctx.ReportNodeWithDeferredSuggestions(node, errorMessage, func() []rule.RuleSuggestion {
			return []rule.RuleSuggestion{
				{
					Message:  removeMessage,
					FixesArr: []rule.RuleFix{rule.RuleFixRemove(ctx.SourceFile, node)},
				},
				replacementSuggestion(ctx, node),
			}
		})
		return
	}

	if parent != nil && parent.Kind == ast.KindVariableDeclaration {
		declaration := parent.AsVariableDeclaration()
		declarationList := parent.Parent
		if declaration != nil && declarationList != nil &&
			declarationList.Kind == ast.KindVariableDeclarationList &&
			!ast.IsVarConst(declarationList) &&
			utils.ESTreeRuntimeExpression(declaration.Initializer) == node {
			start := variableIDEnd(ctx, declaration)
			if start > 0 {
				ctx.ReportNodeWithDeferredSuggestions(node, errorMessage, func() []rule.RuleSuggestion {
					return []rule.RuleSuggestion{
						{
							Message: removeMessage,
							FixesArr: []rule.RuleFix{
								rule.RuleFixRemoveRange(core.NewTextRange(start, node.End())),
							},
						},
						replacementSuggestion(ctx, node),
					}
				})
				return
			}
		}
	}

	ctx.ReportNodeWithDeferredSuggestions(node, errorMessage, func() []rule.RuleSuggestion {
		return []rule.RuleSuggestion{replacementSuggestion(ctx, node)}
	})
}

// NoNullRule disallows the null literal, with narrowly scoped exceptions for
// APIs that require null and configurable direct-argument/strict-equality checks.
//
// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v74.0.0/rules/no-null.js
var NoNullRule = rule.Rule{
	Name:   "unicorn/no-null",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		return rule.RuleListeners{
			ast.KindNullKeyword: func(node *ast.Node) {
				reportNull(ctx, node, opts)
			},
		}
	},
}
