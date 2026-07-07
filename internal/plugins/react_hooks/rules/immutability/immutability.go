package immutability

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const (
	immutabilityReason = "This value cannot be modified"
)

type immutableKind int

const (
	immutablePropsOrHookArgs immutableKind = iota
	immutableUseState
	immutableUseReducer
	immutableHookReturn
	immutableFrozenHookArgument
	immutableJSXValue
)

type immutableInfo struct {
	kind     immutableKind
	hookName string
}

type immutabilityState struct {
	checked map[*ast.Node]bool
}

type immutabilityScope struct {
	ctx  rule.RuleContext
	root *ast.Node

	symbols map[*ast.Symbol]immutableInfo
	names   map[string]immutableInfo
	reports map[*ast.Node]bool
}

type mutationTarget struct {
	node *ast.Node
	info immutableInfo
}

// ImmutabilityRule is the rslint port of upstream
// `react-hooks/immutability`.
//
// The upstream rule is emitted from React Compiler diagnostics. This port keeps
// the user-facing lint contract local to rslint: React-like functions treat
// component props, hook arguments, hook return values, and values passed to
// hooks as immutable, then report later direct writes and known mutating method
// calls.
var ImmutabilityRule = rule.Rule{
	Name: "react-hooks/immutability",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		state := &immutabilityState{checked: map[*ast.Node]bool{}}
		check := func(node *ast.Node) {
			state.checkFunction(ctx, node)
		}
		return rule.RuleListeners{
			ast.KindFunctionDeclaration: check,
			ast.KindFunctionExpression:  check,
			ast.KindArrowFunction:       check,
		}
	},
}

func (state *immutabilityState) checkFunction(ctx rule.RuleContext, fn *ast.Node) {
	if fn == nil || state.checked[fn] {
		return
	}
	state.checked[fn] = true

	fnTyp := getImmutabilityFunctionType(fn)
	if fnTyp == "" {
		return
	}

	scope := &immutabilityScope{
		ctx:     ctx,
		root:    fn,
		symbols: map[*ast.Symbol]immutableInfo{},
		names:   map[string]immutableInfo{},
		reports: map[*ast.Node]bool{},
	}
	scope.seedParameters()
	scope.scan(fn)
}

func getImmutabilityFunctionType(fn *ast.Node) react_hooksutil.CompilerReactFunctionType {
	if fn == nil || react_hooksutil.IsClassMember(fn) {
		return ""
	}
	name := react_hooksutil.GetFunctionName(fn)
	if name != nil {
		n := utils.SkipAssertionsAndParens(name)
		switch {
		case n != nil && n.Kind == ast.KindIdentifier:
			text := n.AsIdentifier().Text
			if react_hooksutil.IsComponentNameStr(text) && react_hooksutil.CallsHooksOrCreatesJsx(fn) {
				return react_hooksutil.CompilerReactFunctionComponent
			}
			if react_hooksutil.IsCompilerHookName(text) {
				return react_hooksutil.CompilerReactFunctionHook
			}
		case n != nil && react_hooksutil.IsCompilerHookCallee(n):
			return react_hooksutil.CompilerReactFunctionHook
		}
	}
	if isDefaultExportedAnonymousFunction(fn) && react_hooksutil.CallsHooksOrCreatesJsx(fn) {
		return react_hooksutil.CompilerReactFunctionComponent
	}
	if ast.IsFunctionExpressionOrArrowFunction(fn) {
		if react_hooksutil.IsForwardRefOrMemoCallback(fn, "forwardRef") ||
			react_hooksutil.IsForwardRefOrMemoCallback(fn, "memo") {
			return react_hooksutil.CompilerReactFunctionComponent
		}
	}
	return ""
}

func isDefaultExportedAnonymousFunction(fn *ast.Node) bool {
	if fn == nil || react_hooksutil.GetFunctionName(fn) != nil {
		return false
	}
	if fn.Kind == ast.KindFunctionDeclaration {
		return ast.GetCombinedModifierFlags(fn)&ast.ModifierFlagsDefault != 0
	}
	child := fn
	for parent := fn.Parent; parent != nil; parent = parent.Parent {
		if parent.Kind == ast.KindParenthesizedExpression {
			child = parent
			continue
		}
		if parent.Kind != ast.KindExportAssignment {
			return false
		}
		assignment := parent.AsExportAssignment()
		return assignment != nil && !assignment.IsExportEquals && assignment.Expression == child
	}
	return false
}

func (scope *immutabilityScope) seedParameters() {
	info := immutableInfo{kind: immutablePropsOrHookArgs}
	for _, param := range scope.root.Parameters() {
		if param == nil {
			continue
		}
		scope.addBinding(param.Name(), info)
	}
}

func (scope *immutabilityScope) scan(node *ast.Node) {
	if node == nil {
		return
	}

	switch node.Kind {
	case ast.KindVariableDeclaration:
		scope.processVariableDeclaration(node)
	case ast.KindBinaryExpression:
		scope.checkAssignment(node)
	case ast.KindPrefixUnaryExpression:
		scope.checkPrefixUpdate(node)
	case ast.KindPostfixUnaryExpression:
		scope.checkPostfixUpdate(node)
	case ast.KindDeleteExpression:
		scope.checkDelete(node)
	case ast.KindCallExpression:
		scope.checkMutatingCall(node)
		scope.freezeHookArguments(node)
		scope.freezeCreateElementArguments(node)
	case ast.KindJsxExpression:
		scope.freezeJSXExpression(node)
	case ast.KindJsxSpreadAttribute:
		scope.freezeJSXSpreadAttribute(node)
	}

	node.ForEachChild(func(child *ast.Node) bool {
		scope.scan(child)
		return false
	})
}

func (scope *immutabilityScope) processVariableDeclaration(node *ast.Node) {
	if isForInOrOfVariableDeclaration(node) {
		return
	}
	decl := node.AsVariableDeclaration()
	if decl == nil || decl.Initializer == nil {
		return
	}
	init := utils.SkipAssertionsAndParens(decl.Initializer)
	if init == nil {
		return
	}
	if hookName, ok := hookCallName(init); ok {
		scope.addHookReturnBinding(decl.Name(), hookName)
		return
	}
	if info, ok := scope.infoForExpression(init); ok {
		scope.addBinding(decl.Name(), info)
	}
}

func isForInOrOfVariableDeclaration(node *ast.Node) bool {
	return node != nil &&
		node.Kind == ast.KindVariableDeclaration &&
		node.Parent != nil &&
		node.Parent.Kind == ast.KindVariableDeclarationList &&
		node.Parent.Parent != nil &&
		ast.IsForInOrOfStatement(node.Parent.Parent)
}

func (scope *immutabilityScope) addHookReturnBinding(name *ast.Node, hookName string) {
	switch hookName {
	case "useState":
		scope.addFirstArrayBinding(name, immutableInfo{kind: immutableUseState, hookName: hookName})
	case "useReducer":
		scope.addFirstArrayBinding(name, immutableInfo{kind: immutableUseReducer, hookName: hookName})
	case "useOptimistic":
		scope.addFirstArrayBinding(name, immutableInfo{kind: immutableHookReturn, hookName: hookName})
	case "useRef":
		return
	default:
		scope.addBinding(name, immutableInfo{kind: immutableHookReturn, hookName: hookName})
	}
}

func (scope *immutabilityScope) addHookReturnAssignmentBinding(name *ast.Node, hookName string) {
	switch hookName {
	case "useState":
		scope.addFirstArrayAssignmentBinding(name, immutableInfo{kind: immutableUseState, hookName: hookName})
	case "useReducer":
		scope.addFirstArrayAssignmentBinding(name, immutableInfo{kind: immutableUseReducer, hookName: hookName})
	case "useOptimistic":
		scope.addFirstArrayAssignmentBinding(name, immutableInfo{kind: immutableHookReturn, hookName: hookName})
	case "useRef":
		return
	default:
		scope.addAssignmentBinding(name, immutableInfo{kind: immutableHookReturn, hookName: hookName})
	}
}

func (scope *immutabilityScope) addFirstArrayBinding(name *ast.Node, info immutableInfo) {
	if name == nil {
		return
	}
	if name.Kind != ast.KindArrayBindingPattern {
		scope.addBinding(name, info)
		return
	}
	found := false
	name.ForEachChild(func(child *ast.Node) bool {
		if found {
			return true
		}
		if child == nil || child.Kind != ast.KindBindingElement {
			return false
		}
		binding := child.AsBindingElement()
		if binding == nil || binding.Name() == nil {
			return false
		}
		scope.addBinding(binding.Name(), info)
		found = true
		return true
	})
}

func (scope *immutabilityScope) addFirstArrayAssignmentBinding(name *ast.Node, info immutableInfo) {
	name = utils.SkipAssertionsAndParens(name)
	if name == nil {
		return
	}
	if name.Kind != ast.KindArrayLiteralExpression {
		scope.addAssignmentBinding(name, info)
		return
	}
	found := false
	name.ForEachChild(func(child *ast.Node) bool {
		if found {
			return true
		}
		if child == nil || child.Kind == ast.KindOmittedExpression {
			return false
		}
		scope.addAssignmentBinding(child, info)
		found = true
		return true
	})
}

func (scope *immutabilityScope) addBinding(nameNode *ast.Node, info immutableInfo) {
	utils.CollectBindingNames(nameNode, func(ident *ast.Node, name string) {
		if ident == nil || name == "" {
			return
		}
		if scope.ctx.TypeChecker != nil {
			if sym := scope.ctx.TypeChecker.GetSymbolAtLocation(ident); sym != nil {
				scope.addSymbol(sym, info)
				return
			}
		}
		if _, exists := scope.names[name]; !exists {
			scope.names[name] = info
		}
	})
}

func (scope *immutabilityScope) addSymbol(sym *ast.Symbol, info immutableInfo) {
	if sym != nil {
		if _, exists := scope.symbols[sym]; exists {
			return
		}
		scope.symbols[sym] = info
	}
}

func (scope *immutabilityScope) infoForExpression(node *ast.Node) (immutableInfo, bool) {
	root := react_hooksutil.AccessChainRootIdentifier(node)
	if root == nil {
		return immutableInfo{}, false
	}
	return scope.infoForIdentifier(root)
}

func (scope *immutabilityScope) infoForIdentifier(id *ast.Node) (immutableInfo, bool) {
	if id == nil || id.Kind != ast.KindIdentifier {
		return immutableInfo{}, false
	}
	if scope.ctx.TypeChecker != nil {
		if sym := utils.GetReferenceSymbol(id, scope.ctx.TypeChecker); sym != nil {
			info, ok := scope.symbols[sym]
			return info, ok
		}
	}
	if info, ok := scope.names[id.AsIdentifier().Text]; ok {
		return info, true
	}
	return immutableInfo{}, false
}

func (scope *immutabilityScope) checkAssignment(node *ast.Node) {
	if !ast.IsAssignmentExpression(node, false) || utils.IsDefaultValueInDestructuringAssignment(node) {
		return
	}
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil {
		return
	}
	targets := scope.collectMutationTargets(binary.Left)
	if binary.OperatorToken.Kind == ast.KindEqualsToken && isBareIdentifierTarget(binary.Left) {
		// `x = value` rebinds a local slot; it does not mutate the value that
		// `x` previously pointed at. Property writes and compound assignments
		// are still reported through the normal mutation-target path.
		targets = nil
	}
	for _, target := range targets {
		scope.report(target)
	}
	if binary.OperatorToken.Kind == ast.KindEqualsToken {
		if len(targets) == 0 {
			if hookName, ok := hookCallName(binary.Right); ok {
				scope.addHookReturnAssignmentBinding(binary.Left, hookName)
			} else if info, ok := scope.infoForExpression(binary.Right); ok {
				scope.addAssignmentBinding(binary.Left, info)
			}
		}
	}
}

func isBareIdentifierTarget(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	return node != nil && node.Kind == ast.KindIdentifier
}

func (scope *immutabilityScope) checkPrefixUpdate(node *ast.Node) {
	prefix := node.AsPrefixUnaryExpression()
	if prefix == nil || (prefix.Operator != ast.KindPlusPlusToken && prefix.Operator != ast.KindMinusMinusToken) {
		return
	}
	for _, target := range scope.collectMutationTargets(prefix.Operand) {
		scope.report(target)
	}
}

func (scope *immutabilityScope) checkPostfixUpdate(node *ast.Node) {
	postfix := node.AsPostfixUnaryExpression()
	if postfix == nil || (postfix.Operator != ast.KindPlusPlusToken && postfix.Operator != ast.KindMinusMinusToken) {
		return
	}
	for _, target := range scope.collectMutationTargets(postfix.Operand) {
		scope.report(target)
	}
}

func (scope *immutabilityScope) checkDelete(node *ast.Node) {
	del := node.AsDeleteExpression()
	if del == nil {
		return
	}
	for _, target := range scope.collectMutationTargets(del.Expression) {
		scope.report(target)
	}
}

func (scope *immutabilityScope) checkMutatingCall(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil {
		return
	}
	callee := utils.SkipAssertionsAndParens(call.Expression)
	if callee == nil {
		return
	}
	if utils.IsSpecificMemberAccess(callee, "Object", "assign") && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
		for _, target := range scope.collectMutationTargets(call.Arguments.Nodes[0]) {
			if isMutatingCallReportable(target.info) {
				scope.report(target)
			}
		}
		return
	}
	if ast.IsAccessExpression(callee) {
		name, ok := utils.AccessExpressionStaticName(callee)
		if ok && isKnownMutatingMethod(name) {
			receiver := utils.AccessExpressionObject(callee)
			for _, target := range scope.collectMutationTargets(receiver) {
				if isMutatingCallReportable(target.info) {
					scope.report(target)
				}
			}
		}
		return
	}
}

func (scope *immutabilityScope) freezeHookArguments(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil || call.Arguments == nil {
		return
	}
	if _, ok := hookCallName(node); !ok {
		return
	}
	info := immutableInfo{kind: immutableFrozenHookArgument}
	for _, arg := range call.Arguments.Nodes {
		if arg == nil {
			continue
		}
		if id := react_hooksutil.AccessChainRootIdentifier(arg); id != nil {
			if sym := scope.symbolForIdentifier(id); sym != nil {
				scope.addSymbol(sym, info)
			}
		}
		if react_hooksutil.IsCompilerFunctionKind(utils.SkipAssertionsAndParens(arg)) {
			scope.freezeCapturedValues(arg, info)
		}
	}
}

func (scope *immutabilityScope) freezeCreateElementArguments(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return
	}
	callee := utils.SkipAssertionsAndParens(call.Expression)
	if !isCreateElementCallee(callee) {
		return
	}
	info := immutableInfo{kind: immutableJSXValue}
	for i, arg := range call.Arguments.Nodes {
		if i == 0 {
			continue
		}
		scope.freezeExpressionReferences(arg, info)
	}
}

func (scope *immutabilityScope) freezeJSXExpression(node *ast.Node) {
	expr := node.AsJsxExpression()
	if expr == nil {
		return
	}
	scope.freezeExpressionReferences(expr.Expression, immutableInfo{kind: immutableJSXValue})
}

func (scope *immutabilityScope) freezeJSXSpreadAttribute(node *ast.Node) {
	attr := node.AsJsxSpreadAttribute()
	if attr == nil {
		return
	}
	scope.freezeExpressionReferences(attr.Expression, immutableInfo{kind: immutableJSXValue})
}

func (scope *immutabilityScope) freezeExpressionReferences(node *ast.Node, info immutableInfo) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return
	}
	if react_hooksutil.IsCompilerFunctionKind(node) {
		// Function bodies are separate execution points. Hook callbacks get
		// their captured values frozen through freezeCapturedValues.
		return
	}
	if id := react_hooksutil.AccessChainRootIdentifier(node); id != nil && !utils.IsNonReferenceIdentifier(id) {
		scope.freezeIdentifier(id, info)
		if ast.IsAccessExpression(node) {
			return
		}
	}
	node.ForEachChild(func(child *ast.Node) bool {
		scope.freezeExpressionReferences(child, info)
		return false
	})
}

func isCreateElementCallee(callee *ast.Node) bool {
	if utils.IsSpecificMemberAccess(callee, "React", "createElement") {
		return true
	}
	return callee != nil && callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "createElement"
}

func (scope *immutabilityScope) freezeIdentifier(id *ast.Node, info immutableInfo) {
	if id == nil || id.Kind != ast.KindIdentifier || scope.ctx.TypeChecker == nil {
		return
	}
	sym := utils.GetReferenceSymbol(id, scope.ctx.TypeChecker)
	if sym == nil || !scope.symbolDeclaredInRootFunction(sym) {
		return
	}
	scope.addSymbol(sym, info)
}

func (scope *immutabilityScope) freezeCapturedValues(fn *ast.Node, info immutableInfo) {
	fn = utils.SkipAssertionsAndParens(fn)
	if fn == nil || !react_hooksutil.IsCompilerFunctionKind(fn) {
		return
	}
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node != fn && react_hooksutil.IsCompilerFunctionKind(node) {
			return
		}
		if node.Kind == ast.KindIdentifier && !utils.IsNonReferenceIdentifier(node) {
			if sym := scope.symbolForIdentifier(node); sym != nil && scope.symbolBelongsToRoot(sym, fn) {
				scope.addSymbol(sym, info)
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(fn)
}

func (scope *immutabilityScope) symbolForIdentifier(id *ast.Node) *ast.Symbol {
	if id == nil || id.Kind != ast.KindIdentifier || scope.ctx.TypeChecker == nil {
		return nil
	}
	return utils.GetReferenceSymbol(id, scope.ctx.TypeChecker)
}

func (scope *immutabilityScope) symbolDeclaredInRootFunction(sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	for _, decl := range sym.Declarations {
		if decl == nil || ast.GetSourceFileOfNode(decl) != scope.ctx.SourceFile {
			continue
		}
		if react_hooksutil.ContainsNode(scope.root, decl) && react_hooksutil.FindEnclosingFunction(decl) == scope.root {
			return true
		}
	}
	return false
}

func (scope *immutabilityScope) symbolBelongsToRoot(sym *ast.Symbol, callback *ast.Node) bool {
	if sym == nil {
		return false
	}
	for _, decl := range sym.Declarations {
		if decl == nil || ast.GetSourceFileOfNode(decl) != scope.ctx.SourceFile {
			continue
		}
		if react_hooksutil.ContainsNode(scope.root, decl) && !react_hooksutil.ContainsNode(callback, decl) {
			return true
		}
	}
	return false
}

func (scope *immutabilityScope) collectMutationTargets(node *ast.Node) []mutationTarget {
	var targets []mutationTarget
	var collect func(*ast.Node)
	collect = func(cur *ast.Node) {
		cur = utils.SkipAssertionsAndParens(cur)
		if cur == nil {
			return
		}
		if id := react_hooksutil.AccessChainRootIdentifier(cur); id != nil {
			if info, ok := scope.infoForIdentifier(id); ok {
				targets = append(targets, mutationTarget{node: id, info: info})
				return
			}
		}
		switch cur.Kind {
		case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
			cur.ForEachChild(func(child *ast.Node) bool {
				collect(child)
				return false
			})
		case ast.KindPropertyAssignment:
			assignment := cur.AsPropertyAssignment()
			if assignment != nil {
				collect(assignment.Initializer)
			}
		case ast.KindShorthandPropertyAssignment:
			name := cur.Name()
			if name != nil {
				collect(name)
			}
			shorthand := cur.AsShorthandPropertyAssignment()
			if shorthand != nil {
				collect(shorthand.ObjectAssignmentInitializer)
			}
		case ast.KindSpreadAssignment, ast.KindSpreadElement:
			cur.ForEachChild(func(child *ast.Node) bool {
				collect(child)
				return false
			})
		case ast.KindBinaryExpression:
			binary := cur.AsBinaryExpression()
			if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
				collect(binary.Left)
			}
		}
	}
	collect(node)
	return targets
}

func (scope *immutabilityScope) addAssignmentBinding(node *ast.Node, info immutableInfo) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return
	}
	switch node.Kind {
	case ast.KindVariableDeclarationList:
		list := node.AsVariableDeclarationList()
		if list == nil || list.Declarations == nil {
			return
		}
		for _, decl := range list.Declarations.Nodes {
			if decl != nil && decl.Kind == ast.KindVariableDeclaration {
				scope.addBinding(decl.Name(), info)
			}
		}
	case ast.KindIdentifier, ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		scope.addBinding(node, info)
	case ast.KindObjectLiteralExpression:
		obj := node.AsObjectLiteralExpression()
		if obj == nil || obj.Properties == nil {
			return
		}
		for _, prop := range obj.Properties.Nodes {
			switch prop.Kind {
			case ast.KindShorthandPropertyAssignment:
				scope.addBinding(prop.Name(), info)
				shorthand := prop.AsShorthandPropertyAssignment()
				if shorthand != nil {
					scope.addAssignmentBinding(shorthand.ObjectAssignmentInitializer, info)
				}
			case ast.KindPropertyAssignment:
				assignment := prop.AsPropertyAssignment()
				if assignment != nil {
					scope.addAssignmentBinding(assignment.Initializer, info)
				}
			case ast.KindSpreadAssignment:
				prop.ForEachChild(func(child *ast.Node) bool {
					scope.addAssignmentBinding(child, info)
					return false
				})
			}
		}
	case ast.KindArrayLiteralExpression:
		node.ForEachChild(func(child *ast.Node) bool {
			scope.addAssignmentBinding(child, info)
			return false
		})
	case ast.KindSpreadElement:
		node.ForEachChild(func(child *ast.Node) bool {
			scope.addAssignmentBinding(child, info)
			return false
		})
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary != nil && binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken {
			scope.addAssignmentBinding(binary.Left, info)
		}
	}
}

func (scope *immutabilityScope) report(target mutationTarget) {
	if target.node == nil || scope.reports[target.node] {
		return
	}
	scope.reports[target.node] = true
	scope.ctx.ReportNode(target.node, buildImmutabilityMessage(target.info))
}

func buildImmutabilityMessage(info immutableInfo) rule.RuleMessage {
	description := ""
	switch info.kind {
	case immutablePropsOrHookArgs:
		description = "Modifying component props or hook arguments is not allowed. Consider using a local variable instead."
	case immutableUseState:
		description = "Modifying a value returned from 'useState()', which should not be modified directly. Use the setter function to update instead."
	case immutableUseReducer:
		description = "Modifying a value returned from 'useReducer()', which should not be modified directly. Use the dispatch function to update instead."
	case immutableHookReturn:
		if info.hookName != "" {
			description = fmt.Sprintf("Modifying a value returned from '%s()' is not allowed.", info.hookName)
			if info.hookName == "useContext" {
				description += "."
			}
		} else {
			description = "Modifying a value returned from a hook is not allowed. Consider moving the modification into the hook where the value is constructed."
		}
	case immutableFrozenHookArgument:
		description = "Modifying a value previously passed as an argument to a hook is not allowed. Consider moving the modification before calling the hook."
	case immutableJSXValue:
		description = "Modifying a value used previously in JSX is not allowed. Consider moving the modification before the JSX."
	}
	return rule.RuleMessage{
		Id:          "immutableMutation",
		Description: fmt.Sprintf("Error: %s\n\n%s", immutabilityReason, description),
	}
}

func hookCallName(node *ast.Node) (string, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindCallExpression {
		return "", false
	}
	callee := utils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
	if !react_hooksutil.IsCompilerHookCallee(callee) {
		return "", false
	}
	name := hookCalleeName(callee)
	if name == "" {
		return "", false
	}
	return name, true
}

func hookCalleeName(callee *ast.Node) string {
	callee = utils.SkipAssertionsAndParens(callee)
	if callee == nil {
		return ""
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.AsIdentifier().Text
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.AsIdentifier().Text
		}
	}
	return ""
}

func isKnownMutatingMethod(name string) bool {
	switch name {
	case "clear", "delete", "pop", "push", "set":
		return true
	}
	return false
}

func isMutatingCallReportable(info immutableInfo) bool {
	switch info.kind {
	case immutableFrozenHookArgument, immutableJSXValue:
		return true
	default:
		return false
	}
}
