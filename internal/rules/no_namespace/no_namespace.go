// Package no_namespace 提供了 TypeScript 代码检查规则
// 注意：当前文件内容似乎与文件名不匹配，实际实现的是 promise-function-async 规则
package no_namespace

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// buildMissingAsyncMessage 构建缺失 async 关键字的错误消息
func buildMissingAsyncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingAsync",
		Description: "Functions that return promises must be async.",
	}
}

// PromiseFunctionAsyncOptions 定义了 Promise 函数异步化规则的配置选项
type PromiseFunctionAsyncOptions struct {
	AllowAny *bool // 是否允许 any 类型
	// TODO(port): TypeOrValueSpecifier
	AllowedPromiseNames       []string // 允许的 Promise 类型名称列表
	CheckArrowFunctions       *bool    // 是否检查箭头函数
	CheckFunctionDeclarations *bool    // 是否检查函数声明
	CheckFunctionExpressions  *bool    // 是否检查函数表达式
	CheckMethodDeclarations   *bool    // 是否检查方法声明
}

// PromiseFunctionAsyncRule 是主要的规则实例
// 该规则检查返回 Promise 的函数是否使用了 async 关键字
var PromiseFunctionAsyncRule = rule.CreateRule(rule.Rule{
	Name: "promise-function-async",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		// 解析和设置默认选项
		opts, ok := options.(PromiseFunctionAsyncOptions)
		if !ok {
			opts = PromiseFunctionAsyncOptions{}
		}
		if opts.AllowAny == nil {
			opts.AllowAny = utils.Ref(true)
		}
		if opts.AllowedPromiseNames == nil {
			opts.AllowedPromiseNames = []string{}
		}
		if opts.CheckArrowFunctions == nil {
			opts.CheckArrowFunctions = utils.Ref(true)
		}
		if opts.CheckFunctionDeclarations == nil {
			opts.CheckFunctionDeclarations = utils.Ref(true)
		}
		if opts.CheckFunctionExpressions == nil {
			opts.CheckFunctionExpressions = utils.Ref(true)
		}
		if opts.CheckMethodDeclarations == nil {
			opts.CheckMethodDeclarations = utils.Ref(true)
		}

		// 构建允许的 Promise 类型名称集合
		allAllowedPromiseNames := utils.NewSetWithSizeHint[string](len(opts.AllowedPromiseNames))
		allAllowedPromiseNames.Add("Promise")
		for _, name := range opts.AllowedPromiseNames {
			allAllowedPromiseNames.Add(name)
		}

		// containsAllTypesByName 检查类型是否包含指定的 Promise 类型名称
		var containsAllTypesByName func(t *checker.Type, matchAnyInstead bool) bool
		containsAllTypesByName = func(t *checker.Type, matchAnyInstead bool) bool {
			// 跳过 any 或 unknown 类型
			if utils.IsTypeFlagSet(t, checker.TypeFlagsAnyOrUnknown) {
				return false
			}

			// 处理引用类型
			if utils.IsTypeFlagSet(t, checker.TypeFlagsObject) && checker.Type_objectFlags(t)&checker.ObjectFlagsReference != 0 {
				t = t.Target()
			}

			// 检查符号名称是否匹配允许的 Promise 名称
			symbol := checker.Type_symbol(t)
			if symbol != nil && allAllowedPromiseNames.Has(symbol.Name) {
				return true
			}

			predicate := func(t *checker.Type) bool {
				return containsAllTypesByName(t, matchAnyInstead)
			}

			// 处理联合类型和交叉类型
			if utils.IsUnionType(t) || utils.IsIntersectionType(t) {
				if matchAnyInstead {
					return utils.Every(t.Types(), predicate)
				}
				return utils.Some(t.Types(), predicate)
			}

			// 处理类或接口类型
			if checker.Type_objectFlags(t)&checker.ObjectFlagsClassOrInterface == 0 {
				return false
			}

			bases := checker.Checker_getBaseTypes(ctx.TypeChecker, t)
			if matchAnyInstead {
				return utils.Some(bases, predicate)
			}
			return len(bases) > 0 && utils.Every(bases, predicate)
		}

		// 创建规则监听器
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
