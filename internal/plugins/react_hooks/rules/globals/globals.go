package globals

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	globalReassignmentReason = "Cannot reassign variables declared outside of the component/hook"
)

type activeFunctionKind int

const (
	activeFunctionNone activeFunctionKind = iota
	activeFunctionRoot
	activeFunctionRenderCallback
	activeFunctionHelper
)

type globalsState struct {
	activeFunctions map[*ast.Node]activeFunctionKind
}

// GlobalsRule is the rslint port of upstream `react-hooks/globals`.
var GlobalsRule = rule.Rule{
	Name: "react-hooks/globals",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		state := &globalsState{
			activeFunctions: make(map[*ast.Node]activeFunctionKind),
		}
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				state.checkAssignment(ctx, node)
			},
		}
	},
}

func (state *globalsState) checkAssignment(ctx rule.RuleContext, node *ast.Node) {
	if node == nil || !ast.IsAssignmentExpression(node, false) {
		return
	}
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return
	}
	if ast.IsLogicalOrCoalescingAssignmentExpression(node) {
		return
	}
	if utils.IsDefaultValueInDestructuringAssignment(node) {
		return
	}
	targets := collectAssignmentTargets(binary.Left)
	for _, target := range targets {
		fn := react_hooksutil.FindEnclosingFunction(target.node)
		if fn == nil || state.activeFunctionKind(fn) == activeFunctionNone {
			continue
		}
		if state.isLocalToFunction(fn, target.node, target.name) {
			continue
		}
		reportNode := target.node
		if binary.OperatorToken.Kind != ast.KindEqualsToken {
			// Upstream reports compound assignments over the whole assignment
			// expression, while plain assignments point at the written binding.
			reportNode = node
		}
		ctx.ReportNode(reportNode, buildGlobalReassignmentMessage(target.name))
	}
}

func buildGlobalReassignmentMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "globalReassignment",
		Description: fmt.Sprintf(
			"Error: %s\n\nVariable `%s` is declared outside of the component/hook. Reassigning this value during render is a form of side effect, which can cause unpredictable behavior depending on when the component happens to re-render. If this variable is used in rendering, use useState instead. Otherwise, consider updating it in an effect. (https://react.dev/reference/rules/components-and-hooks-must-be-pure#side-effects-must-run-outside-of-render).",
			globalReassignmentReason,
			name,
		),
		Data: map[string]string{
			"name": name,
		},
	}
}

type assignmentTarget struct {
	node *ast.Node
	name string
}

func collectAssignmentTargets(node *ast.Node) []assignmentTarget {
	var targets []assignmentTarget
	collectAssignmentTargetsInto(node, &targets)
	return targets
}

func collectAssignmentTargetsInto(node *ast.Node, targets *[]assignmentTarget) {
	if node == nil {
		return
	}
	node = ast.SkipParentheses(node)
	switch node.Kind {
	case ast.KindIdentifier:
		name := node.AsIdentifier().Text
		if name != "" {
			*targets = append(*targets, assignmentTarget{node: node, name: name})
		}
	case ast.KindObjectLiteralExpression:
		obj := node.AsObjectLiteralExpression()
		if obj == nil || obj.Properties == nil {
			return
		}
		for _, prop := range obj.Properties.Nodes {
			switch prop.Kind {
			case ast.KindShorthandPropertyAssignment:
				shorthand := prop.AsShorthandPropertyAssignment()
				name := prop.Name()
				if shorthand != nil && shorthand.ObjectAssignmentInitializer != nil && name != nil && name.Kind == ast.KindIdentifier {
					*targets = append(*targets, assignmentTarget{node: prop, name: name.AsIdentifier().Text})
					continue
				}
				collectAssignmentTargetsInto(name, targets)
			case ast.KindPropertyAssignment:
				assignment := prop.AsPropertyAssignment()
				if assignment != nil {
					collectAssignmentTargetsInto(assignment.Initializer, targets)
				}
			case ast.KindSpreadAssignment:
				prop.ForEachChild(func(child *ast.Node) bool {
					collectAssignmentTargetsInto(child, targets)
					return false
				})
			}
		}
	case ast.KindArrayLiteralExpression:
		node.ForEachChild(func(child *ast.Node) bool {
			collectAssignmentTargetsInto(child, targets)
			return false
		})
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			left := ast.SkipParentheses(binary.Left)
			if left != nil && left.Kind == ast.KindIdentifier {
				name := left.AsIdentifier().Text
				if name != "" {
					*targets = append(*targets, assignmentTarget{node: node, name: name})
				}
				return
			}
			collectAssignmentTargetsInto(left, targets)
		}
	case ast.KindSpreadElement:
		node.ForEachChild(func(child *ast.Node) bool {
			collectAssignmentTargetsInto(child, targets)
			return false
		})
	}
}

func (state *globalsState) activeFunctionKind(fn *ast.Node) activeFunctionKind {
	if fn == nil {
		return activeFunctionNone
	}
	if kind, ok := state.activeFunctions[fn]; ok {
		return kind
	}

	kind := activeFunctionNone
	if !isNonRenderCallback(fn) {
		parentFn := react_hooksutil.FindEnclosingFunction(fn)
		if parentFn == nil {
			if isCompilerRenderFunction(fn) {
				kind = activeFunctionRoot
			}
		} else {
			parentKind := state.activeFunctionKind(parentFn)
			if parentKind == activeFunctionRoot || parentKind == activeFunctionRenderCallback {
				switch {
				case isUseMemoCallback(fn):
					kind = activeFunctionRenderCallback
				case isImmediatelyCalled(fn):
					kind = activeFunctionHelper
				case state.isObjectHelperCalled(fn, parentFn):
					kind = activeFunctionHelper
				case state.isDirectlyCalledByName(fn, parentFn):
					kind = activeFunctionHelper
				case state.isReferencedAsJsxChild(fn, parentFn):
					kind = activeFunctionHelper
				case functionReturnsJsx(fn) && state.isPassedAsJsxProp(fn, parentFn):
					kind = activeFunctionHelper
				}
			}
		}
	}

	state.activeFunctions[fn] = kind
	return kind
}

func isCompilerRenderFunction(fn *ast.Node) bool {
	return react_hooksutil.GetCompilerReactFunctionType(fn, react_hooksutil.CompilerFunctionOptions{}) != ""
}

func isNonRenderCallback(fn *ast.Node) bool {
	return isCallbackArgumentOf(fn, "useEffect", "useLayoutEffect", "useInsertionEffect", "useCallback")
}

func isUseMemoCallback(fn *ast.Node) bool {
	return isCallbackArgumentOf(fn, "useMemo")
}

func isCallbackArgumentOf(fn *ast.Node, names ...string) bool {
	for _, name := range names {
		if react_hooksutil.GetReactCallbackCall(fn, name) != nil {
			return true
		}
	}
	return false
}

func isImmediatelyCalled(fn *ast.Node) bool {
	return ast.GetImmediatelyInvokedFunctionExpression(fn) != nil
}

func (state *globalsState) isObjectHelperCalled(fn *ast.Node, parentFn *ast.Node) bool {
	name := objectLiteralRootBindingName(fn)
	if name == "" {
		return false
	}
	found := false
	walkRenderParentBody(parentFn, func(node *ast.Node) bool {
		if node.Kind != ast.KindCallExpression {
			return false
		}
		call := node.AsCallExpression()
		if memberExpressionRootIdentifier(call.Expression) == name {
			found = true
			return true
		}
		return false
	})
	return found
}

func objectLiteralRootBindingName(fn *ast.Node) string {
	child := fn
	for parent := fn.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			child = parent
		case ast.KindPropertyAssignment:
			assignment := parent.AsPropertyAssignment()
			if assignment == nil || assignment.Initializer != child {
				return ""
			}
			child = parent
		case ast.KindObjectLiteralExpression:
			child = parent
		case ast.KindVariableDeclaration:
			decl := parent.AsVariableDeclaration()
			if decl == nil || decl.Initializer != child {
				return ""
			}
			return identifierName(parent.Name())
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken || binary.Right != child {
				return ""
			}
			return identifierName(ast.SkipParentheses(binary.Left))
		default:
			return ""
		}
	}
	return ""
}

func memberExpressionRootIdentifier(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsAccessExpression(node) {
		return ""
	}
	for node != nil && ast.IsAccessExpression(node) {
		node = utils.AccessExpressionObject(node)
		node = ast.SkipParentheses(node)
	}
	if node != nil && node.Kind == ast.KindIdentifier {
		return node.AsIdentifier().Text
	}
	return ""
}

func identifierName(node *ast.Node) string {
	if node == nil || node.Kind != ast.KindIdentifier {
		return ""
	}
	return node.AsIdentifier().Text
}

func (state *globalsState) isDirectlyCalledByName(fn *ast.Node, parentFn *ast.Node) bool {
	name := functionIdentifierName(fn)
	if name == "" {
		return false
	}
	found := false
	walkRenderParentBody(parentFn, func(node *ast.Node) bool {
		if node.Kind != ast.KindCallExpression {
			return false
		}
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee != nil && callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == name {
			found = true
			return true
		}
		return false
	})
	return found
}

func (state *globalsState) isReferencedAsJsxChild(fn *ast.Node, parentFn *ast.Node) bool {
	name := functionIdentifierName(fn)
	if name == "" {
		return false
	}
	found := false
	walkRenderParentBody(parentFn, func(node *ast.Node) bool {
		if node.Kind != ast.KindJsxExpression {
			return false
		}
		if node.Parent != nil && ast.IsJsxAttributeLike(node.Parent) {
			return false
		}
		expr := ast.SkipParentheses(node.AsJsxExpression().Expression)
		if expr != nil && expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text == name {
			found = true
			return true
		}
		return false
	})
	return found
}

func (state *globalsState) isPassedAsJsxProp(fn *ast.Node, parentFn *ast.Node) bool {
	name := functionIdentifierName(fn)
	if name == "" {
		return false
	}
	found := false
	walkRenderParentBody(parentFn, func(node *ast.Node) bool {
		if node.Kind != ast.KindJsxExpression || node.Parent == nil || !ast.IsJsxAttributeLike(node.Parent) {
			return false
		}
		expr := ast.SkipParentheses(node.AsJsxExpression().Expression)
		if expr != nil && expr.Kind == ast.KindIdentifier && expr.AsIdentifier().Text == name {
			found = true
			return true
		}
		return false
	})
	return found
}

func functionIdentifierName(fn *ast.Node) string {
	name := react_hooksutil.GetFunctionName(fn)
	if name == nil {
		return ""
	}
	name = ast.SkipParentheses(name)
	if name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

func walkRenderParentBody(fn *ast.Node, visit func(*ast.Node) bool) {
	body := react_hooksutil.GetFunctionBody(fn)
	if body == nil {
		return
	}
	var walk func(*ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node != body && utils.IsFunctionLikeContainer(node) {
			return false
		}
		if visit(node) {
			return true
		}
		stop := false
		node.ForEachChild(func(child *ast.Node) bool {
			if walk(child) {
				stop = true
				return true
			}
			return false
		})
		return stop
	}
	walk(body)
}

func functionReturnsJsx(fn *ast.Node) bool {
	body := react_hooksutil.GetFunctionBody(fn)
	if body == nil {
		return false
	}
	if body.Kind != ast.KindBlock {
		return containsJsx(body)
	}
	found := false
	var walk func(*ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if node != body && utils.IsFunctionLikeContainer(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			ret := node.AsReturnStatement()
			if ret != nil && ret.Expression != nil && containsJsx(ret.Expression) {
				found = true
				return true
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			return walk(child)
		})
		return found
	}
	walk(body)
	return found
}

func containsJsx(node *ast.Node) bool {
	if node == nil {
		return false
	}
	found := false
	var walk func(*ast.Node) bool
	walk = func(n *ast.Node) bool {
		if n == nil {
			return false
		}
		switch n.Kind {
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement, ast.KindJsxFragment:
			found = true
			return true
		}
		if n != node && utils.IsFunctionLikeContainer(n) {
			return false
		}
		n.ForEachChild(func(child *ast.Node) bool {
			return walk(child)
		})
		return found
	}
	walk(node)
	return found
}

func (state *globalsState) isLocalToFunction(fn *ast.Node, target *ast.Node, name string) bool {
	if fn == nil || target == nil || name == "" {
		return false
	}
	if fn.Kind == ast.KindFunctionExpression {
		if fnName := fn.Name(); fnName != nil && fnName.Kind == ast.KindIdentifier && fnName.AsIdentifier().Text == name {
			return true
		}
	}
	if utils.HasShadowingParameter(fn, name) {
		return true
	}
	body := react_hooksutil.GetFunctionBody(fn)
	if body != nil && utils.HasHoistedVarDeclaration(body, name) {
		return true
	}
	return utils.IsNameShadowedBetween(target, fn, name)
}
