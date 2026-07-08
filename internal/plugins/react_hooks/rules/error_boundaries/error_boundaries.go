package error_boundaries

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
)

const errorBoundariesMessage = "Error: Avoid constructing JSX within try/catch\n\n" +
	"React does not immediately render components when JSX is rendered, so any errors from this component will not be caught by the try/catch. " +
	"To catch errors in rendering a given component, wrap that component in an error boundary. " +
	"(https://react.dev/reference/react/Component#catching-rendering-errors-with-an-error-boundary)."

// ErrorBoundariesRule is the rslint port of upstream
// `react-hooks/error-boundaries`.
var ErrorBoundariesRule = rule.Rule{
	Name: "react-hooks/error-boundaries",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		if !mayContainReactCode(ctx.SourceFile) {
			return rule.RuleListeners{}
		}
		check := func(node *ast.Node) {
			fn := react_hooksutil.FindEnclosingFunction(node)
			if fn == nil || !isCompilerLintFunction(fn) {
				return
			}
			if jsxIsInsideTryBlock(node, fn) {
				ctx.ReportNode(node, rule.RuleMessage{Description: errorBoundariesMessage})
			}
		}
		return rule.RuleListeners{
			ast.KindJsxElement:            check,
			ast.KindJsxSelfClosingElement: check,
			ast.KindJsxFragment:           check,
		}
	},
}

func mayContainReactCode(sf *ast.SourceFile) bool {
	if sf == nil || sf.Statements == nil {
		return false
	}
	for _, stmt := range sf.Statements.Nodes {
		if checkTopLevelNode(stmt) {
			return true
		}
	}
	return false
}

func checkTopLevelNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		if ast.GetFunctionFlags(node)&ast.FunctionFlagsGenerator != 0 {
			return false
		}
		if name := node.Name(); isReactComponentOrHookName(name) {
			return true
		}
		return node.Name() == nil && ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault)
	case ast.KindVariableStatement:
		vs := node.AsVariableStatement()
		if vs.DeclarationList == nil {
			return false
		}
		decls := vs.DeclarationList.AsVariableDeclarationList().Declarations
		if decls == nil {
			return false
		}
		for _, decl := range decls.Nodes {
			if topLevelVariableDeclMayContainReactCode(decl) {
				return true
			}
		}
	case ast.KindExportAssignment:
		ea := node.AsExportAssignment()
		if ea.IsExportEquals || ea.Expression == nil {
			return false
		}
		expr := ast.SkipParentheses(ea.Expression)
		return isFunctionExpressionLike(expr)
	}
	return false
}

func topLevelVariableDeclMayContainReactCode(decl *ast.Node) bool {
	if decl == nil || decl.Kind != ast.KindVariableDeclaration {
		return false
	}
	vd := decl.AsVariableDeclaration()
	if !isReactComponentOrHookName(vd.Name()) || vd.Initializer == nil {
		return false
	}
	return isFunctionExpressionLike(ast.SkipParentheses(vd.Initializer))
}

func isReactComponentOrHookName(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	name := node.AsIdentifier().Text
	return react_hooksutil.IsComponentNameStr(name) || isCompilerHookNameStr(name)
}

func isHookNameNode(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindIdentifier && isCompilerHookNameStr(node.AsIdentifier().Text)
}

func isComponentNameNode(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindIdentifier && react_hooksutil.IsComponentNameStr(node.AsIdentifier().Text)
}

func isCompilerHookNameStr(name string) bool {
	return name != "use" && react_hooksutil.IsHookName(name)
}

func isFunctionExpressionLike(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindFunctionExpression && ast.GetFunctionFlags(node)&ast.FunctionFlagsGenerator != 0 {
		return false
	}
	return node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindArrowFunction
}

func isCompilerLintFunction(fn *ast.Node) bool {
	if fn == nil {
		return false
	}
	if ast.GetFunctionFlags(fn)&ast.FunctionFlagsGenerator != 0 {
		return false
	}
	// Keep this narrower than react_hooksutil.IsComponentOrHookFn: the
	// compiler-backed lint pre-pass only checks top-level components, any hook,
	// and memo/forwardRef callbacks once the file may contain React code.
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		name := fn.Name()
		if isHookNameNode(name) {
			return true
		}
		return isComponentNameNode(name) && isTopLevelStatement(fn)
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		if name, topLevel := variableNameForFunction(fn); name != nil {
			if isHookNameNode(name) {
				return true
			}
			return topLevel && isComponentNameNode(name)
		}
		return isMemoOrForwardRefCallback(fn)
	}
	return false
}

func variableNameForFunction(fn *ast.Node) (*ast.Node, bool) {
	child, parent := walkUpParens(fn)
	if parent == nil || parent.Kind != ast.KindVariableDeclaration {
		return nil, false
	}
	vd := parent.AsVariableDeclaration()
	if vd.Initializer != child {
		return nil, false
	}
	return vd.Name(), isTopLevelVariableDeclaration(parent)
}

func isMemoOrForwardRefCallback(fn *ast.Node) bool {
	call := react_hooksutil.GetForwardRefOrMemoCallbackCall(fn, "memo")
	if call == nil {
		call = react_hooksutil.GetForwardRefOrMemoCallbackCall(fn, "forwardRef")
	}
	return call != nil && !callOrCalleeIsOptionalChain(call)
}

func callOrCalleeIsOptionalChain(callNode *ast.Node) bool {
	if ast.IsOptionalChain(callNode) {
		return true
	}
	call := callNode.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	return callee != nil && ast.IsOptionalChain(callee)
}

func walkUpParens(node *ast.Node) (*ast.Node, *ast.Node) {
	child := node
	parent := node.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child = parent
		parent = parent.Parent
	}
	return child, parent
}

func isTopLevelVariableDeclaration(decl *ast.Node) bool {
	if decl == nil || decl.Parent == nil || decl.Parent.Parent == nil {
		return false
	}
	return decl.Parent.Kind == ast.KindVariableDeclarationList &&
		decl.Parent.Parent.Kind == ast.KindVariableStatement &&
		isTopLevelStatement(decl.Parent.Parent)
}

func isTopLevelStatement(node *ast.Node) bool {
	return node != nil && node.Parent != nil && node.Parent.Kind == ast.KindSourceFile
}

func jsxIsInsideTryBlock(node *ast.Node, fn *ast.Node) bool {
	insideTryBlock := false
	child := node
	for parent := node.Parent; parent != nil && parent != fn; parent = parent.Parent {
		if react_hooksutil.IsFunctionLikeContainer(parent) {
			return false
		}
		if parent.Kind == ast.KindTryStatement {
			tryStmt := parent.AsTryStatement()
			if tryStmt.FinallyBlock != nil && tryStmt.FinallyBlock.AsNode() == child {
				return false
			}
			if tryStmt.TryBlock != nil && tryStmt.TryBlock.AsNode() == child {
				insideTryBlock = true
			}
		}
		child = parent
	}
	return insideTryBlock
}
