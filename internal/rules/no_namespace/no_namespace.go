package no_namespace

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// build the message for no-namespace rule
func buildNoNamespaceMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noNamespace",
		Description: "Namespace is not allowed.",
	}
}

// rule options
type NoNamespaceOptions struct {
	AllowDeclarations    *bool `json:"allowDeclarations"`
	AllowDefinitionFiles *bool `json:"allowDefinitionFiles"`
}

// default options
var defaultNoNamespaceOptions = NoNamespaceOptions{
	AllowDeclarations:    utils.Ref(false),
	AllowDefinitionFiles: utils.Ref(true),
}

// rule instance
// check if the namespace is used
var NoNamespaceRule = rule.CreateRule(rule.Rule{
	Name: "no-namespace",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts, ok := options.(NoNamespaceOptions)
		if !ok {
			opts = NoNamespaceOptions{}
		}

		// set default options
		if opts.AllowDeclarations == nil {
			opts.AllowDeclarations = defaultNoNamespaceOptions.AllowDeclarations
		}
		if opts.AllowDefinitionFiles == nil {
			opts.AllowDefinitionFiles = defaultNoNamespaceOptions.AllowDefinitionFiles
		}

		// create listeners
		listeners := make(rule.RuleListeners, 3)

		// validateNode 验证节点是否需要 async 关键字
		validateNode := func(node *ast.Node) {

			// 如果已经是 async 函数或没有函数体，则跳过
			if utils.IncludesModifier(node, ast.KindAsyncKeyword) || node.Body() == nil {
				return
			}

			// 获取函数类型和调用签名
			t := ctx.TypeChecker.GetTypeAtLocation(node)
			signatures := utils.GetCallSignatures(ctx.TypeChecker, t)
			if len(signatures) == 0 {
				return
			}

			// 检查所有签名是否都返回 Promise
			everySignatureReturnsPromise := true
			for _, signature := range signatures {
				returnType := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, signature)
				if !*opts.AllowAny && utils.IsTypeFlagSet(returnType, checker.TypeFlagsAnyOrUnknown) {
					// 如果返回类型是未知的且不允许 any，则报告错误但不自动修复
					// TODO(port): getFunctionHeadLoc
					ctx.ReportNode(node, buildMissingAsyncMessage())
					return
				}

				// 要求所有潜在的返回类型都是 promise/any/unknown
				everySignatureReturnsPromise = everySignatureReturnsPromise && containsAllTypesByName(
					returnType,
					// 如果没有显式设置返回类型，我们检查返回类型的任何部分是否匹配 Promise（而不是要求全部匹配）
					node.Type() != nil,
				)
			}

			if !everySignatureReturnsPromise {
				return
			}

			// 确定插入 async 关键字的位置
			insertAsyncBeforeNode := node
			if ast.IsMethodDeclaration(node) {
				insertAsyncBeforeNode = node.Name()
			}
			// TODO(port): getFunctionHeadLoc
			ctx.ReportNodeWithFixes(node, buildMissingAsyncMessage(), rule.RuleFixInsertBefore(ctx.SourceFile, insertAsyncBeforeNode, " async "))
		}

		// 根据配置添加相应的监听器
		if *opts.CheckArrowFunctions {
			listeners[ast.KindArrowFunction] = validateNode
		}

		if *opts.CheckFunctionDeclarations {
			listeners[ast.KindFunctionDeclaration] = validateNode
		}

		if *opts.CheckFunctionExpressions {
			listeners[ast.KindFunctionExpression] = validateNode
		}

		if *opts.CheckMethodDeclarations {
			listeners[ast.KindMethodDeclaration] = func(node *ast.Node) {
				// 抽象方法不能是 async
				if utils.IncludesModifier(node, ast.KindAbstractKeyword) {
					return
				}
				validateNode(node)
			}
		}

		return listeners
	},
})
