package incompatible_library

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const incompatibleLibraryDescription = "This API returns functions which cannot be memoized without leading to stale UI. " +
	"To prevent this, by default React Compiler will skip memoizing this component/hook. " +
	"However, you may see issues if values from this API are passed to other components/hooks that are memoized."

const (
	reactHookFormWatchMessage = "React Hook Form's `useForm()` API returns a `watch()` function which cannot be memoized safely."
	tanStackTableMessage      = "TanStack Table's `useReactTable()` API returns functions that cannot be memoized safely"
	tanStackVirtualMessage    = "TanStack Virtual's `useVirtualizer()` API returns functions that cannot be memoized safely"
)

type bindingKind int

const (
	bindingUnknown bindingKind = iota
	bindingReactHookFormNamespace
	bindingTanStackTableNamespace
	bindingTanStackVirtualNamespace
	bindingUseForm
	bindingFormReturn
	bindingWatch
	bindingUseReactTable
	bindingUseVirtualizer
)

type lexicalScope map[string]bindingKind

type incompatibleLibraryState struct {
	ctx    rule.RuleContext
	scopes []lexicalScope
}

// IncompatibleLibraryRule is the rslint port of upstream
// `react-hooks/incompatible-library`.
//
// Upstream emits this rule from React Compiler's IncompatibleLibrary
// diagnostic category. This port keeps the published built-in module model
// local to rslint: React Hook Form's `useForm().watch`, TanStack Table's
// `useReactTable`, and TanStack Virtual's `useVirtualizer`.
var IncompatibleLibraryRule = rule.Rule{
	Name: "react-hooks/incompatible-library",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		state := &incompatibleLibraryState{ctx: ctx}
		state.pushScope()
		return rule.RuleListeners{
			ast.KindImportDeclaration:   state.processImportDeclaration,
			ast.KindFunctionDeclaration: state.enterFunctionDeclaration,
			ast.KindFunctionExpression:  state.enterFunctionExpression,
			ast.KindArrowFunction:       state.enterFunctionExpression,
			ast.KindBlock:               state.enterBlock,
			ast.KindModuleBlock:         state.enterBlock,
			ast.KindCatchClause:         state.enterCatchClause,
			ast.KindVariableDeclaration: state.processVariableDeclaration,
			ast.KindBinaryExpression:    state.processBinaryExpression,
			ast.KindCallExpression:      state.processCallExpression,

			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(node *ast.Node) { state.popScope() },
			rule.ListenerOnExit(ast.KindFunctionExpression):  func(node *ast.Node) { state.popScope() },
			rule.ListenerOnExit(ast.KindArrowFunction):       func(node *ast.Node) { state.popScope() },
			rule.ListenerOnExit(ast.KindBlock):               func(node *ast.Node) { state.popScope() },
			rule.ListenerOnExit(ast.KindModuleBlock):         func(node *ast.Node) { state.popScope() },
			rule.ListenerOnExit(ast.KindCatchClause):         func(node *ast.Node) { state.popScope() },
		}
	},
}

func (state *incompatibleLibraryState) processImportDeclaration(node *ast.Node) {
	if ast.IsTypeOnlyImportDeclaration(node) {
		return
	}
	decl := node.AsImportDeclaration()
	if decl == nil || decl.ImportClause == nil {
		return
	}
	moduleName, ok := utils.GetStaticStringLiteralValue(utils.SkipAssertionsAndParens(decl.ModuleSpecifier))
	if !ok {
		return
	}
	clause := decl.ImportClause.AsImportClause()
	if clause == nil {
		return
	}
	if clause.Name() != nil {
		state.declareNode(clause.Name(), bindingUnknown)
	}
	if clause.NamedBindings == nil {
		return
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		state.processNamespaceImport(moduleName, clause.NamedBindings)
	case ast.KindNamedImports:
		state.processNamedImports(moduleName, clause.NamedBindings)
	}
}

func (state *incompatibleLibraryState) processNamespaceImport(moduleName string, node *ast.Node) {
	namespace := node.AsNamespaceImport()
	if namespace == nil || namespace.Name() == nil {
		return
	}
	switch moduleName {
	case "react-hook-form":
		state.declareNode(namespace.Name(), bindingReactHookFormNamespace)
	case "@tanstack/react-table":
		state.declareNode(namespace.Name(), bindingTanStackTableNamespace)
	case "@tanstack/react-virtual":
		state.declareNode(namespace.Name(), bindingTanStackVirtualNamespace)
	default:
		state.declareNode(namespace.Name(), bindingUnknown)
	}
}

func (state *incompatibleLibraryState) processNamedImports(moduleName string, node *ast.Node) {
	named := node.AsNamedImports()
	if named == nil || named.Elements == nil {
		return
	}
	for _, elem := range named.Elements.Nodes {
		spec := elem.AsImportSpecifier()
		if spec == nil || spec.Name() == nil || spec.IsTypeOnly {
			continue
		}
		importedName := importSpecifierImportedName(spec)
		localName := spec.Name()
		switch {
		case moduleName == "react-hook-form" && importedName == "useForm":
			state.declareNode(localName, bindingUseForm)
		case moduleName == "@tanstack/react-table" && importedName == "useReactTable":
			state.declareNode(localName, bindingUseReactTable)
		case moduleName == "@tanstack/react-virtual" && importedName == "useVirtualizer":
			state.declareNode(localName, bindingUseVirtualizer)
		default:
			state.declareNode(localName, bindingUnknown)
		}
	}
}

func (state *incompatibleLibraryState) enterFunctionDeclaration(node *ast.Node) {
	if name := node.Name(); name != nil {
		state.declareNode(name, bindingUnknown)
	}
	state.pushScope()
	if name := node.Name(); name != nil {
		state.declareNode(name, bindingUnknown)
	}
	state.declareParameters(node)
	state.predeclareHoistedVarNames(node.Body())
}

func (state *incompatibleLibraryState) enterFunctionExpression(node *ast.Node) {
	state.pushScope()
	if name := node.Name(); name != nil {
		state.declareNode(name, bindingUnknown)
	}
	state.declareParameters(node)
	state.predeclareHoistedVarNames(node.Body())
}

func (state *incompatibleLibraryState) enterBlock(node *ast.Node) {
	state.pushScope()
	state.predeclareBlockDeclarations(node)
}

func (state *incompatibleLibraryState) enterCatchClause(node *ast.Node) {
	state.pushScope()
	catchClause := node.AsCatchClause()
	if catchClause == nil || catchClause.VariableDeclaration == nil {
		return
	}
	decl := catchClause.VariableDeclaration.AsVariableDeclaration()
	if decl != nil {
		state.declareBindingName(decl.Name(), bindingUnknown)
	}
}

func (state *incompatibleLibraryState) processVariableDeclaration(node *ast.Node) {
	decl := node.AsVariableDeclaration()
	if decl == nil {
		return
	}
	name := decl.Name()
	if name == nil {
		return
	}
	initKind := state.expressionBindingKind(decl.Initializer)
	if initKind != bindingUnknown && !isInsideIncompatibleLibraryTarget(node) {
		state.declareBindingName(name, bindingUnknown)
		return
	}
	if initKind == bindingFormReturn {
		state.declareFormReturnBinding(name)
		return
	}
	if initKind == bindingReactHookFormNamespace || initKind == bindingTanStackTableNamespace || initKind == bindingTanStackVirtualNamespace {
		state.declareModuleObjectPattern(name, initKind)
		return
	}
	state.declareBindingName(name, initKind)
}

func (state *incompatibleLibraryState) processBinaryExpression(node *ast.Node) {
	if !ast.IsAssignmentExpression(node, false) {
		return
	}
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}
	rightKind := state.expressionBindingKind(binary.Right)
	if rightKind != bindingUnknown && !isInsideIncompatibleLibraryTarget(node) {
		state.assignBindingName(binary.Left, bindingUnknown)
		return
	}
	switch rightKind {
	case bindingFormReturn:
		state.assignFormReturnBinding(binary.Left)
	case bindingReactHookFormNamespace, bindingTanStackTableNamespace, bindingTanStackVirtualNamespace:
		state.assignModuleObjectPattern(binary.Left, rightKind)
	default:
		state.assignBindingName(binary.Left, rightKind)
	}
}

func (state *incompatibleLibraryState) processCallExpression(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil {
		return
	}
	if !isInsideIncompatibleLibraryTarget(node) {
		return
	}
	callee := utils.SkipAssertionsAndParens(call.Expression)
	kind, reportNode := state.calleeBindingKindAndReportNode(callee)
	switch kind {
	case bindingUseReactTable:
		state.report(reportNode, tanStackTableMessage)
	case bindingUseVirtualizer:
		state.report(reportNode, tanStackVirtualMessage)
	case bindingWatch:
		state.report(reportNode, reactHookFormWatchMessage)
	}
}

// This diagnostic runs only inside functions selected as React Compiler roots.
// Nested callbacks inside those roots are checked, but a top-level memo or
// forwardRef callback is not a root for this rule in upstream.
func isInsideIncompatibleLibraryTarget(node *ast.Node) bool {
	for fn := react_hooksutil.FindEnclosingFunction(node); fn != nil; fn = react_hooksutil.FindEnclosingFunction(fn) {
		if isIncompatibleLibraryRootFunction(fn) {
			return true
		}
	}
	return false
}

func isIncompatibleLibraryRootFunction(fn *ast.Node) bool {
	if fn == nil || isImmediateCallArgument(fn) {
		return false
	}
	name := react_hooksutil.GetFunctionName(fn)
	if name != nil && name.Kind == ast.KindIdentifier && react_hooksutil.IsComponentNameStr(name.AsIdentifier().Text) {
		return react_hooksutil.CallsHooksOrCreatesJsx(fn) &&
			react_hooksutil.IsValidCompilerComponentParams(fn) &&
			!react_hooksutil.ReturnsCompilerNonNode(fn)
	}
	if name != nil && react_hooksutil.IsCompilerHookCallee(name) {
		return react_hooksutil.CallsHooksOrCreatesJsx(fn)
	}
	return false
}

func isImmediateCallArgument(fn *ast.Node) bool {
	child := fn
	parent := fn.Parent
	for parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		child = parent
		parent = parent.Parent
	}
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if call == nil || call.Arguments == nil {
		return false
	}
	for _, arg := range call.Arguments.Nodes {
		if arg == child {
			return true
		}
	}
	return false
}

func (state *incompatibleLibraryState) declareParameters(fn *ast.Node) {
	for _, param := range fn.Parameters() {
		if param == nil {
			continue
		}
		state.declareBindingName(param.Name(), bindingUnknown)
	}
}

func (state *incompatibleLibraryState) predeclareBlockDeclarations(block *ast.Node) {
	var statements []*ast.Node
	switch block.Kind {
	case ast.KindBlock:
		if b := block.AsBlock(); b != nil && b.Statements != nil {
			statements = b.Statements.Nodes
		}
	case ast.KindModuleBlock:
		if b := block.AsModuleBlock(); b != nil && b.Statements != nil {
			statements = b.Statements.Nodes
		}
	}
	for _, stmt := range statements {
		state.predeclareStatement(stmt)
	}
}

func (state *incompatibleLibraryState) predeclareStatement(stmt *ast.Node) {
	if stmt == nil {
		return
	}
	switch stmt.Kind {
	case ast.KindVariableStatement:
		varStmt := stmt.AsVariableStatement()
		if varStmt == nil || varStmt.DeclarationList == nil {
			return
		}
		state.declareVarListBindings(varStmt.DeclarationList)
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration, ast.KindEnumDeclaration, ast.KindModuleDeclaration:
		if name := stmt.Name(); name != nil && name.Kind == ast.KindIdentifier {
			state.declareNode(name, bindingUnknown)
		}
	}
}

func (state *incompatibleLibraryState) predeclareHoistedVarNames(root *ast.Node) {
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node != root && ast.IsFunctionLikeOrClassStaticBlockDeclaration(node) {
			return
		}
		switch node.Kind {
		case ast.KindVariableStatement:
			varStmt := node.AsVariableStatement()
			if varStmt != nil && varStmt.DeclarationList != nil && utils.IsVarKeyword(varStmt.DeclarationList) {
				state.declareVarListBindings(varStmt.DeclarationList)
			}
		case ast.KindForStatement:
			forStmt := node.AsForStatement()
			if forStmt != nil && forStmt.Initializer != nil &&
				forStmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				utils.IsVarKeyword(forStmt.Initializer) {
				state.declareVarListBindings(forStmt.Initializer)
			}
		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := node.AsForInOrOfStatement()
			if stmt != nil && stmt.Initializer != nil &&
				stmt.Initializer.Kind == ast.KindVariableDeclarationList &&
				utils.IsVarKeyword(stmt.Initializer) {
				state.declareVarListBindings(stmt.Initializer)
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(root)
}

func (state *incompatibleLibraryState) declareVarListBindings(node *ast.Node) {
	declList := node.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return
	}
	for _, decl := range declList.Declarations.Nodes {
		if decl != nil && decl.Kind == ast.KindVariableDeclaration {
			state.declareBindingName(decl.AsVariableDeclaration().Name(), bindingUnknown)
		}
	}
}

func (state *incompatibleLibraryState) declareFormReturnBinding(name *ast.Node) {
	if name == nil {
		return
	}
	if name.Kind == ast.KindIdentifier {
		state.declareNode(name, bindingFormReturn)
		return
	}
	if name.Kind == ast.KindObjectBindingPattern {
		state.declareObjectPattern(name, func(prop string) bindingKind {
			if prop == "watch" {
				return bindingWatch
			}
			return bindingUnknown
		})
		return
	}
	state.declareBindingName(name, bindingUnknown)
}

func (state *incompatibleLibraryState) assignFormReturnBinding(name *ast.Node) {
	if name == nil {
		return
	}
	name = utils.SkipAssertionsAndParens(name)
	if name.Kind == ast.KindIdentifier {
		state.assignNode(name, bindingFormReturn)
		return
	}
	if name.Kind == ast.KindObjectLiteralExpression {
		state.assignObjectPattern(name, func(prop string) bindingKind {
			if prop == "watch" {
				return bindingWatch
			}
			return bindingUnknown
		})
		return
	}
	state.assignBindingName(name, bindingUnknown)
}

func (state *incompatibleLibraryState) declareModuleObjectPattern(name *ast.Node, moduleKind bindingKind) {
	if name == nil {
		return
	}
	if name.Kind == ast.KindIdentifier {
		state.declareNode(name, moduleKind)
		return
	}
	if name.Kind == ast.KindObjectBindingPattern {
		state.declareObjectPattern(name, func(prop string) bindingKind {
			switch moduleKind {
			case bindingReactHookFormNamespace:
				if prop == "useForm" {
					return bindingUseForm
				}
			case bindingTanStackTableNamespace:
				if prop == "useReactTable" {
					return bindingUseReactTable
				}
			case bindingTanStackVirtualNamespace:
				if prop == "useVirtualizer" {
					return bindingUseVirtualizer
				}
			}
			return bindingUnknown
		})
		return
	}
	state.declareBindingName(name, bindingUnknown)
}

func (state *incompatibleLibraryState) assignModuleObjectPattern(name *ast.Node, moduleKind bindingKind) {
	if name == nil {
		return
	}
	name = utils.SkipAssertionsAndParens(name)
	if name.Kind == ast.KindIdentifier {
		state.assignNode(name, moduleKind)
		return
	}
	if name.Kind == ast.KindObjectLiteralExpression {
		state.assignObjectPattern(name, func(prop string) bindingKind {
			switch moduleKind {
			case bindingReactHookFormNamespace:
				if prop == "useForm" {
					return bindingUseForm
				}
			case bindingTanStackTableNamespace:
				if prop == "useReactTable" {
					return bindingUseReactTable
				}
			case bindingTanStackVirtualNamespace:
				if prop == "useVirtualizer" {
					return bindingUseVirtualizer
				}
			}
			return bindingUnknown
		})
		return
	}
	state.assignBindingName(name, bindingUnknown)
}

func (state *incompatibleLibraryState) declareBindingName(name *ast.Node, kind bindingKind) {
	if name == nil {
		return
	}
	switch name.Kind {
	case ast.KindIdentifier:
		state.declareNode(name, kind)
	case ast.KindObjectBindingPattern:
		state.declareObjectPattern(name, func(string) bindingKind { return kind })
	case ast.KindArrayBindingPattern:
		name.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				state.declareBindingName(child.AsBindingElement().Name(), kind)
			}
			return false
		})
	}
}

func (state *incompatibleLibraryState) assignBindingName(name *ast.Node, kind bindingKind) {
	if name == nil {
		return
	}
	name = utils.SkipAssertionsAndParens(name)
	switch name.Kind {
	case ast.KindIdentifier:
		state.assignNode(name, kind)
	case ast.KindObjectLiteralExpression:
		state.assignObjectPattern(name, func(string) bindingKind { return kind })
	case ast.KindArrayLiteralExpression:
		name.ForEachChild(func(child *ast.Node) bool {
			if child.Kind != ast.KindSpreadElement {
				state.assignBindingName(child, kind)
			}
			return false
		})
	}
}

func (state *incompatibleLibraryState) declareObjectPattern(pattern *ast.Node, kindForProperty func(string) bindingKind) {
	pattern.ForEachChild(func(child *ast.Node) bool {
		if child.Kind != ast.KindBindingElement {
			return false
		}
		binding := child.AsBindingElement()
		if binding == nil {
			return false
		}
		propName := bindingElementPropertyName(child)
		childKind := kindForProperty(propName)
		if childKind != bindingUnknown && binding.Name() != nil && binding.Name().Kind != ast.KindIdentifier {
			state.declareBindingName(binding.Name(), bindingUnknown)
			return false
		}
		state.declareBindingName(binding.Name(), childKind)
		return false
	})
}

func (state *incompatibleLibraryState) assignObjectPattern(pattern *ast.Node, kindForProperty func(string) bindingKind) {
	pattern.ForEachChild(func(child *ast.Node) bool {
		switch child.Kind {
		case ast.KindPropertyAssignment:
			prop := child.AsPropertyAssignment()
			if prop == nil {
				return false
			}
			childKind := kindForProperty(bindingElementPropertyName(child))
			if childKind != bindingUnknown && prop.Initializer != nil && prop.Initializer.Kind != ast.KindIdentifier {
				state.assignBindingName(prop.Initializer, bindingUnknown)
				return false
			}
			state.assignBindingName(prop.Initializer, childKind)
		case ast.KindShorthandPropertyAssignment:
			name := child.Name()
			if name != nil {
				state.assignBindingName(name, kindForProperty(name.Text()))
			}
		case ast.KindSpreadAssignment:
			if name := child.Name(); name != nil {
				state.assignBindingName(name, bindingUnknown)
			}
		}
		return false
	})
}

func (state *incompatibleLibraryState) expressionBindingKind(expr *ast.Node) bindingKind {
	expr = utils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return bindingUnknown
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		return state.lookup(expr.AsIdentifier().Text)
	case ast.KindCallExpression:
		if state.expressionBindingKind(expr.AsCallExpression().Expression) == bindingUseForm {
			return bindingFormReturn
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return state.accessExpressionBindingKind(expr)
	}
	return bindingUnknown
}

func (state *incompatibleLibraryState) calleeBindingKindAndReportNode(callee *ast.Node) (bindingKind, *ast.Node) {
	kind := state.expressionBindingKind(callee)
	if kind == bindingUnknown {
		return kind, callee
	}
	if ast.IsAccessExpression(callee) {
		if receiver := reportableAccessReceiver(callee); receiver != nil {
			return kind, receiver
		}
	}
	return kind, callee
}

// Upstream reports member-call diagnostics on the receiver expression:
// `form.watch()` reports `form`, and `(Table).useReactTable()` reports `Table`.
func reportableAccessReceiver(expr *ast.Node) *ast.Node {
	object := utils.AccessExpressionObject(expr)
	for object != nil {
		object = ast.SkipParentheses(object)
		if object == nil || object.Kind != ast.KindNonNullExpression {
			return object
		}
		object = object.AsNonNullExpression().Expression
	}
	return expr
}

func (state *incompatibleLibraryState) accessExpressionBindingKind(expr *ast.Node) bindingKind {
	propName, ok := utils.AccessExpressionStaticName(expr)
	if !ok {
		return bindingUnknown
	}
	object := utils.SkipAssertionsAndParens(utils.AccessExpressionObject(expr))
	objectKind := state.expressionBindingKind(object)
	switch objectKind {
	case bindingReactHookFormNamespace:
		if propName == "useForm" {
			return bindingUseForm
		}
	case bindingTanStackTableNamespace:
		if propName == "useReactTable" {
			return bindingUseReactTable
		}
	case bindingTanStackVirtualNamespace:
		if propName == "useVirtualizer" {
			return bindingUseVirtualizer
		}
	case bindingFormReturn:
		if propName == "watch" {
			return bindingWatch
		}
	}
	return bindingUnknown
}

func (state *incompatibleLibraryState) report(node *ast.Node, detail string) {
	if node == nil {
		return
	}
	state.ctx.ReportNode(node, buildIncompatibleLibraryMessage(detail))
}

func buildIncompatibleLibraryMessage(detail string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "incompatibleLibrary",
		Description: fmt.Sprintf(
			"Compilation Skipped: Use of incompatible library\n\n%s\n\n%s",
			incompatibleLibraryDescription,
			detail,
		),
		Data: map[string]string{"detail": detail},
	}
}

func (state *incompatibleLibraryState) pushScope() {
	state.scopes = append(state.scopes, lexicalScope{})
}

func (state *incompatibleLibraryState) popScope() {
	state.scopes = state.scopes[:len(state.scopes)-1]
}

func (state *incompatibleLibraryState) declareNode(node *ast.Node, kind bindingKind) {
	if node == nil || node.Kind != ast.KindIdentifier || len(state.scopes) == 0 {
		return
	}
	state.scopes[len(state.scopes)-1][node.AsIdentifier().Text] = kind
}

func (state *incompatibleLibraryState) assignNode(node *ast.Node, kind bindingKind) {
	if node == nil || node.Kind != ast.KindIdentifier {
		return
	}
	name := node.AsIdentifier().Text
	for i := len(state.scopes) - 1; i >= 0; i-- {
		if _, ok := state.scopes[i][name]; ok {
			state.scopes[i][name] = kind
			return
		}
	}
	state.declareNode(node, kind)
}

func (state *incompatibleLibraryState) lookup(name string) bindingKind {
	for i := len(state.scopes) - 1; i >= 0; i-- {
		if kind, ok := state.scopes[i][name]; ok {
			return kind
		}
	}
	return bindingUnknown
}

func importSpecifierImportedName(spec *ast.ImportSpecifier) string {
	if spec.PropertyName != nil {
		return moduleExportNameText(spec.PropertyName)
	}
	if spec.Name() != nil {
		return spec.Name().Text()
	}
	return ""
}

func moduleExportNameText(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral:
		return node.Text()
	}
	return ""
}

func bindingElementPropertyName(node *ast.Node) string {
	if propertyName := ast.TryGetPropertyNameOfBindingOrAssignmentElement(node); propertyName != nil {
		if prop, ok := utils.GetStaticPropertyName(propertyName); ok {
			return prop
		}
	}
	return ""
}
