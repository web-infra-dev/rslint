package preserve_manual_memoization

import (
	"fmt"
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/react_hooks/react_hooksutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const preserveManualMemoizationReason = "Existing memoization could not be preserved"

type dependencyPathPart struct {
	property string
	optional bool
}

type memoDependency struct {
	root string
	path []dependencyPathPart
	node *ast.Node
}

func (dep memoDependency) String() string {
	var builder strings.Builder
	builder.WriteString(dep.root)
	for _, part := range dep.path {
		if part.optional {
			builder.WriteString("?.")
		} else {
			builder.WriteString(".")
		}
		builder.WriteString(part.property)
	}
	return builder.String()
}

func (dep memoDependency) hasRefCurrentAccess() bool {
	for _, part := range dep.path {
		if part.property == "current" {
			return true
		}
	}
	return false
}

type compareDependencyResult int

const (
	compareDependencyOk compareDependencyResult = iota
	compareDependencyRootDifference
	compareDependencyPathDifference
	compareDependencySubpath
	compareDependencyRefAccessDifference
)

type preserveManualMemoizationState struct {
	checked map[*ast.Node]bool
}

type rootMemoContext struct {
	reactNames     map[string]bool
	memoNames      map[string]string
	moduleBindings map[string]bool
	stableBindings map[string]bool
}

// PreserveManualMemoizationRule is the rslint port of upstream
// `react-hooks/preserve-manual-memoization`.
//
// The upstream ESLint rule is backed by React Compiler diagnostics. This port
// implements the user-facing dependency-preservation contract locally: manual
// useMemo/useCallback dependency arrays must cover the dependencies inferred
// from the memo callback, while stable React values are ignored.
var PreserveManualMemoizationRule = rule.Rule{
	Name: "react-hooks/preserve-manual-memoization",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		state := &preserveManualMemoizationState{checked: map[*ast.Node]bool{}}
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

func (state *preserveManualMemoizationState) checkFunction(ctx rule.RuleContext, fn *ast.Node) {
	if fn == nil || state.checked[fn] {
		return
	}
	state.checked[fn] = true
	if react_hooksutil.GetCompilerReactFunctionType(fn) == "" {
		return
	}

	memoCtx := buildRootMemoContext(ctx.SourceFile, fn)
	react_hooksutil.WalkFunctionBody(fn, func(node *ast.Node) bool {
		if node.Kind != ast.KindCallExpression {
			return false
		}
		state.checkMemoCall(ctx, fn, node, memoCtx)
		return false
	})
}

func (state *preserveManualMemoizationState) checkMemoCall(ctx rule.RuleContext, rootFn, callNode *ast.Node, memoCtx rootMemoContext) {
	call := callNode.AsCallExpression()
	if call == nil || call.Arguments == nil || len(call.Arguments.Nodes) < 2 {
		return
	}
	kind, ok := manualMemoCalleeKind(call.Expression, memoCtx)
	if !ok || (kind != "useMemo" && kind != "useCallback") {
		return
	}

	callback := utils.SkipAssertionsAndParens(call.Arguments.Nodes[0])
	if !react_hooksutil.IsCompilerFunctionKind(callback) {
		return
	}
	sourceDeps, validDeps := collectSourceDependencies(call.Arguments.Nodes[1])
	if !validDeps {
		return
	}

	collector := newDependencyCollector(callback, memoCtx)
	inferredDeps := collector.collect()
	for _, inferred := range inferredDeps {
		result := compareDependencySet(inferred, sourceDeps)
		if result == compareDependencyOk {
			continue
		}
		ctx.ReportNode(callback, buildPreservedManualMemoizationDependencyMessage(inferred, sourceDeps, result))
	}

	for _, sourceDep := range sourceDeps {
		if dependencyMutatedAfterCall(rootFn, callNode, sourceDep) {
			reportNode := sourceDep.node
			if reportNode == nil {
				reportNode = callback
			}
			ctx.ReportNode(reportNode, buildPreservedManualMemoizationMutationMessage())
		}
	}
}

func buildPreservedManualMemoizationDependencyMessage(dep memoDependency, sourceDeps []memoDependency, result compareDependencyResult) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "preserveManualMemoization",
		Description: fmt.Sprintf(
			"Compilation Skipped: %s\n\nReact Compiler has skipped optimizing this component because the existing manual memoization could not be preserved. The inferred dependencies did not match the manually specified dependencies, which could cause the value to change more or less frequently than expected. The inferred dependency was `%s`, but the source dependencies were [%s]. %s",
			preserveManualMemoizationReason,
			dep.String(),
			printDependencyList(sourceDeps),
			compareDependencyResultDescription(result),
		),
		Data: map[string]string{
			"dependency": dep.String(),
		},
	}
}

func buildPreservedManualMemoizationMutationMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id: "preserveManualMemoization",
		Description: fmt.Sprintf(
			"Compilation Skipped: %s\n\nReact Compiler has skipped optimizing this component because the existing manual memoization could not be preserved. This dependency may be mutated later, which could cause the value to change unexpectedly.",
			preserveManualMemoizationReason,
		),
	}
}

func compareDependencyResultDescription(result compareDependencyResult) string {
	switch result {
	case compareDependencyRootDifference:
		return "Inferred dependency not present in source."
	case compareDependencyPathDifference:
		return "Inferred different dependency than source."
	case compareDependencySubpath:
		return "Inferred less specific property than source."
	case compareDependencyRefAccessDifference:
		return "Differences in ref.current access."
	default:
		return "Dependencies equal."
	}
}

func compareDependencySet(inferred memoDependency, sourceDeps []memoDependency) compareDependencyResult {
	if len(sourceDeps) == 0 {
		return compareDependencyRootDifference
	}
	var best compareDependencyResult
	for i, source := range sourceDeps {
		result := compareDependencies(inferred, source)
		if result == compareDependencyOk {
			return compareDependencyOk
		}
		if i == 0 || result > best {
			best = result
		}
	}
	return best
}

func compareDependencies(inferred, source memoDependency) compareDependencyResult {
	if inferred.root != source.root {
		return compareDependencyRootDifference
	}

	isSubpath := true
	limit := min(len(inferred.path), len(source.path))
	for i := range limit {
		if inferred.path[i].property != source.path[i].property {
			isSubpath = false
			break
		}
		if inferred.path[i].optional && !source.path[i].optional {
			return compareDependencyPathDifference
		}
	}

	if isSubpath && (len(source.path) == len(inferred.path) || (len(inferred.path) >= len(source.path) && !inferred.hasRefCurrentAccess())) {
		return compareDependencyOk
	}
	if isSubpath {
		if source.hasRefCurrentAccess() || inferred.hasRefCurrentAccess() {
			return compareDependencyRefAccessDifference
		}
		return compareDependencySubpath
	}
	return compareDependencyPathDifference
}

func printDependencyList(deps []memoDependency) string {
	parts := make([]string, 0, len(deps))
	for _, dep := range deps {
		parts = append(parts, dep.String())
	}
	return strings.Join(parts, ", ")
}

func buildRootMemoContext(sourceFile *ast.SourceFile, rootFn *ast.Node) rootMemoContext {
	ctx := rootMemoContext{
		reactNames:     map[string]bool{"React": true},
		memoNames:      map[string]string{"useMemo": "useMemo", "useCallback": "useCallback"},
		moduleBindings: map[string]bool{},
		stableBindings: map[string]bool{},
	}
	for name := range builtinStableNames {
		ctx.stableBindings[name] = true
	}
	collectModuleBindings(sourceFile, &ctx)
	collectStableRootBindings(rootFn, &ctx)
	return ctx
}

var builtinStableNames = map[string]bool{
	"Array":           true,
	"Boolean":         true,
	"Date":            true,
	"Error":           true,
	"Intl":            true,
	"JSON":            true,
	"Math":            true,
	"Number":          true,
	"Object":          true,
	"Promise":         true,
	"RegExp":          true,
	"Set":             true,
	"String":          true,
	"Symbol":          true,
	"URL":             true,
	"URLSearchParams": true,
	"console":         true,
	"document":        true,
	"globalThis":      true,
	"window":          true,
}

func collectModuleBindings(sourceFile *ast.SourceFile, memoCtx *rootMemoContext) {
	if sourceFile == nil || sourceFile.Statements == nil {
		return
	}
	for _, stmt := range sourceFile.Statements.Nodes {
		switch stmt.Kind {
		case ast.KindImportDeclaration:
			collectImportBindings(stmt, memoCtx)
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if name := stmt.Name(); name != nil && name.Kind == ast.KindIdentifier {
				memoCtx.moduleBindings[name.AsIdentifier().Text] = true
			}
		case ast.KindVariableStatement:
			stmt.ForEachChild(func(child *ast.Node) bool {
				if child.Kind == ast.KindVariableDeclaration {
					utils.CollectBindingNames(child.AsVariableDeclaration().Name(), func(_ *ast.Node, name string) {
						memoCtx.moduleBindings[name] = true
					})
				}
				return false
			})
		}
	}
}

func collectImportBindings(node *ast.Node, memoCtx *rootMemoContext) {
	importDecl := node.AsImportDeclaration()
	if importDecl == nil || importDecl.ModuleSpecifier == nil || importDecl.ImportClause == nil {
		return
	}
	source := utils.GetStaticStringValue(importDecl.ModuleSpecifier)
	clause := importDecl.ImportClause.AsImportClause()
	if clause == nil {
		return
	}
	if clause.Name() != nil && clause.Name().Kind == ast.KindIdentifier {
		local := clause.Name().AsIdentifier().Text
		memoCtx.moduleBindings[local] = true
		if source == "react" {
			memoCtx.reactNames[local] = true
		}
	}
	if clause.NamedBindings == nil {
		return
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		name := clause.NamedBindings.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			local := name.AsIdentifier().Text
			memoCtx.moduleBindings[local] = true
			if source == "react" {
				memoCtx.reactNames[local] = true
			}
		}
	case ast.KindNamedImports:
		named := clause.NamedBindings.AsNamedImports()
		if named == nil || named.Elements == nil {
			return
		}
		for _, elem := range named.Elements.Nodes {
			spec := elem.AsImportSpecifier()
			if spec == nil || spec.Name() == nil || spec.Name().Kind != ast.KindIdentifier {
				continue
			}
			local := spec.Name().AsIdentifier().Text
			memoCtx.moduleBindings[local] = true
			imported := local
			if spec.PropertyName != nil {
				imported = spec.PropertyName.Text()
			}
			if source == "react" {
				switch imported {
				case "useMemo", "useCallback":
					memoCtx.memoNames[local] = imported
				}
			}
		}
	}
}

func collectStableRootBindings(rootFn *ast.Node, memoCtx *rootMemoContext) {
	react_hooksutil.WalkFunctionBody(rootFn, func(node *ast.Node) bool {
		if node.Kind != ast.KindVariableDeclaration {
			return false
		}
		decl := node.AsVariableDeclaration()
		if decl == nil || decl.Initializer == nil {
			return false
		}
		init := utils.SkipAssertionsAndParens(decl.Initializer)
		if kind, ok := manualMemoCalleeKind(init, *memoCtx); ok {
			utils.CollectBindingNames(decl.Name(), func(_ *ast.Node, name string) {
				memoCtx.memoNames[name] = kind
			})
			return false
		}
		if init == nil || init.Kind != ast.KindCallExpression {
			return false
		}
		hookName := hookCalleeName(init.AsCallExpression().Expression, *memoCtx)
		switch hookName {
		case "useState", "useReducer", "useTransition", "useActionState", "useOptimistic":
			addSecondArrayBindingName(decl.Name(), memoCtx.stableBindings)
		case "useRef":
			utils.CollectBindingNames(decl.Name(), func(_ *ast.Node, name string) {
				memoCtx.stableBindings[name] = true
			})
		}
		return false
	})
}

func addSecondArrayBindingName(nameNode *ast.Node, stable map[string]bool) {
	nameNode = utils.SkipAssertionsAndParens(nameNode)
	if nameNode == nil || nameNode.Kind != ast.KindArrayBindingPattern {
		return
	}
	pattern := nameNode.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil || len(pattern.Elements.Nodes) < 2 {
		return
	}
	second := pattern.Elements.Nodes[1]
	if second == nil || second.Kind != ast.KindBindingElement {
		return
	}
	utils.CollectBindingNames(second.AsBindingElement().Name(), func(_ *ast.Node, name string) {
		stable[name] = true
	})
}

func manualMemoCalleeKind(callee *ast.Node, memoCtx rootMemoContext) (string, bool) {
	callee = utils.SkipAssertionsAndParens(callee)
	if callee == nil {
		return "", false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		kind, ok := memoCtx.memoNames[callee.AsIdentifier().Text]
		return kind, ok
	case ast.KindPropertyAccessExpression:
		prop := callee.AsPropertyAccessExpression().Name()
		if prop == nil || prop.Kind != ast.KindIdentifier {
			return "", false
		}
		name := prop.AsIdentifier().Text
		if name != "useMemo" && name != "useCallback" {
			return "", false
		}
		obj := utils.SkipAssertionsAndParens(callee.AsPropertyAccessExpression().Expression)
		if obj != nil && obj.Kind == ast.KindIdentifier && memoCtx.reactNames[obj.AsIdentifier().Text] {
			return name, true
		}
	}
	return "", false
}

func hookCalleeName(callee *ast.Node, memoCtx rootMemoContext) string {
	callee = utils.SkipAssertionsAndParens(callee)
	if callee == nil {
		return ""
	}
	if callee.Kind == ast.KindIdentifier {
		return callee.AsIdentifier().Text
	}
	if callee.Kind == ast.KindPropertyAccessExpression {
		prop := callee.AsPropertyAccessExpression().Name()
		obj := utils.SkipAssertionsAndParens(callee.AsPropertyAccessExpression().Expression)
		if prop != nil && prop.Kind == ast.KindIdentifier &&
			obj != nil && obj.Kind == ast.KindIdentifier && memoCtx.reactNames[obj.AsIdentifier().Text] {
			return prop.AsIdentifier().Text
		}
	}
	return ""
}

func collectSourceDependencies(node *ast.Node) ([]memoDependency, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindArrayLiteralExpression {
		return nil, false
	}
	array := node.AsArrayLiteralExpression()
	if array == nil || array.Elements == nil {
		return nil, true
	}
	deps := make([]memoDependency, 0, len(array.Elements.Nodes))
	seen := map[string]bool{}
	for _, elem := range array.Elements.Nodes {
		if elem == nil || elem.Kind == ast.KindSpreadElement {
			return nil, false
		}
		dep, ok := parseMemoDependency(elem)
		if !ok {
			return nil, false
		}
		key := dep.String()
		if seen[key] {
			continue
		}
		seen[key] = true
		deps = append(deps, dep)
	}
	return deps, true
}

func parseMemoDependency(node *ast.Node) (memoDependency, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return memoDependency{}, false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return memoDependency{root: node.AsIdentifier().Text, node: node}, node.AsIdentifier().Text != ""
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		base, ok := parseMemoDependency(access.Expression)
		if !ok {
			return memoDependency{}, false
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return memoDependency{}, false
		}
		base.path = append(base.path, dependencyPathPart{
			property: name.AsIdentifier().Text,
			optional: access.QuestionDotToken != nil,
		})
		base.node = node
		return base, true
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		base, ok := parseMemoDependency(access.Expression)
		if !ok {
			return memoDependency{}, false
		}
		property, ok := utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(access.ArgumentExpression))
		if !ok {
			return memoDependency{}, false
		}
		base.path = append(base.path, dependencyPathPart{
			property: property,
			optional: access.QuestionDotToken != nil,
		})
		base.node = node
		return base, true
	}
	return memoDependency{}, false
}

type dependencyUse struct {
	dep  memoDependency
	node *ast.Node
}

type dependencyCollector struct {
	callback      *ast.Node
	memoCtx       rootMemoContext
	localBindings map[string]bool
	seen          map[string]bool
	dependencies  []dependencyUse
}

func newDependencyCollector(callback *ast.Node, memoCtx rootMemoContext) *dependencyCollector {
	collector := &dependencyCollector{
		callback:      callback,
		memoCtx:       memoCtx,
		localBindings: collectFunctionLocalBindings(callback),
		seen:          map[string]bool{},
	}
	return collector
}

func (collector *dependencyCollector) collect() []memoDependency {
	body := react_hooksutil.GetFunctionBody(collector.callback)
	if body != nil {
		collector.walk(body)
	}
	sort.SliceStable(collector.dependencies, func(i, j int) bool {
		return collector.dependencies[i].node.Pos() < collector.dependencies[j].node.Pos()
	})
	out := make([]memoDependency, 0, len(collector.dependencies))
	for _, use := range collector.dependencies {
		out = append(out, use.dep)
	}
	return out
}

func collectFunctionLocalBindings(fn *ast.Node) map[string]bool {
	locals := map[string]bool{}
	for _, param := range fn.Parameters() {
		utils.CollectBindingNames(param.Name(), func(_ *ast.Node, name string) {
			locals[name] = true
		})
	}
	if name := fn.Name(); name != nil && name.Kind == ast.KindIdentifier {
		locals[name.AsIdentifier().Text] = true
	}
	body := react_hooksutil.GetFunctionBody(fn)
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node != fn && react_hooksutil.IsFunctionLikeContainer(node) {
			if node.Kind == ast.KindFunctionDeclaration {
				if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
					locals[name.AsIdentifier().Text] = true
				}
			}
			return
		}
		switch node.Kind {
		case ast.KindVariableDeclaration:
			utils.CollectBindingNames(node.AsVariableDeclaration().Name(), func(_ *ast.Node, name string) {
				locals[name] = true
			})
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if name := node.Name(); name != nil && name.Kind == ast.KindIdentifier {
				locals[name.AsIdentifier().Text] = true
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(body)
	return locals
}

func (collector *dependencyCollector) walk(node *ast.Node) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || ast.IsTypeNode(node) {
		return
	}
	if node != collector.callback && react_hooksutil.IsFunctionLikeContainer(node) {
		return
	}
	if collector.collectCallExpressionDependencies(node) {
		return
	}
	if dep, ok := collector.dependencyFromAccess(node); ok {
		collector.add(dep, node)
		collector.walkDynamicElementArgument(node)
		return
	}
	if node.Kind == ast.KindIdentifier {
		if collector.isExternalReference(node) {
			collector.add(memoDependency{root: node.AsIdentifier().Text, node: node}, node)
		}
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		collector.walk(child)
		return false
	})
}

func (collector *dependencyCollector) collectCallExpressionDependencies(node *ast.Node) bool {
	if node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	if call == nil {
		return false
	}
	callee := utils.SkipAssertionsAndParens(call.Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return false
	}
	// React Compiler treats a member call as depending on the receiver, not on
	// the method property itself. This keeps `prop.x()` aligned with an
	// inferred `prop` dependency instead of `prop.x`.
	receiver := utils.SkipAssertionsAndParens(utils.AccessExpressionObject(callee))
	if dep, ok := parseAccessDependency(receiver); ok && collector.shouldKeepDependency(dep.root) {
		collector.add(dep, receiver)
	} else {
		collector.walk(receiver)
	}
	collector.walkDynamicElementArgument(callee)
	if call.Arguments != nil {
		for _, arg := range call.Arguments.Nodes {
			collector.walk(arg)
		}
	}
	return true
}

func (collector *dependencyCollector) walkDynamicElementArgument(node *ast.Node) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return
	}
	if node.Kind == ast.KindElementAccessExpression {
		elem := node.AsElementAccessExpression()
		if _, ok := utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(elem.ArgumentExpression)); !ok {
			collector.walk(elem.ArgumentExpression)
		}
		collector.walkDynamicElementArgument(elem.Expression)
		return
	}
	if node.Kind == ast.KindPropertyAccessExpression {
		collector.walkDynamicElementArgument(node.AsPropertyAccessExpression().Expression)
	}
}

func (collector *dependencyCollector) dependencyFromAccess(node *ast.Node) (memoDependency, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil || !ast.IsAccessExpression(node) {
		return memoDependency{}, false
	}
	dep, ok := parseAccessDependency(node)
	if !ok || !collector.shouldKeepDependency(dep.root) {
		return memoDependency{}, false
	}
	return dep, true
}

func parseAccessDependency(node *ast.Node) (memoDependency, bool) {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return memoDependency{}, false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return memoDependency{root: node.AsIdentifier().Text, node: node}, node.AsIdentifier().Text != ""
	case ast.KindPropertyAccessExpression:
		access := node.AsPropertyAccessExpression()
		base, ok := parseAccessDependency(access.Expression)
		if !ok {
			return memoDependency{}, false
		}
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return memoDependency{}, false
		}
		base.path = append(base.path, dependencyPathPart{
			property: name.AsIdentifier().Text,
			optional: access.QuestionDotToken != nil,
		})
		base.node = node
		return base, true
	case ast.KindElementAccessExpression:
		access := node.AsElementAccessExpression()
		base, ok := parseAccessDependency(access.Expression)
		if !ok {
			return memoDependency{}, false
		}
		if property, ok := utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(access.ArgumentExpression)); ok {
			base.path = append(base.path, dependencyPathPart{
				property: property,
				optional: access.QuestionDotToken != nil,
			})
		}
		base.node = node
		return base, true
	}
	return memoDependency{}, false
}

func (collector *dependencyCollector) isExternalReference(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier || utils.IsNonReferenceIdentifier(node) {
		return false
	}
	return collector.shouldKeepDependency(node.AsIdentifier().Text)
}

func (collector *dependencyCollector) shouldKeepDependency(name string) bool {
	if name == "" || collector.localBindings[name] || collector.memoCtx.stableBindings[name] || collector.memoCtx.moduleBindings[name] {
		return false
	}
	return true
}

func (collector *dependencyCollector) add(dep memoDependency, node *ast.Node) {
	key := dep.String()
	if collector.seen[key] {
		return
	}
	for _, existing := range collector.dependencies {
		if compareDependencies(dep, existing.dep) == compareDependencyOk {
			return
		}
	}
	dependencies := collector.dependencies[:0]
	for _, existing := range collector.dependencies {
		if compareDependencies(existing.dep, dep) == compareDependencyOk {
			delete(collector.seen, existing.dep.String())
			continue
		}
		dependencies = append(dependencies, existing)
	}
	collector.dependencies = dependencies
	collector.seen[key] = true
	collector.dependencies = append(collector.dependencies, dependencyUse{dep: dep, node: node})
}

func dependencyMutatedAfterCall(rootFn, callNode *ast.Node, dep memoDependency) bool {
	found := false
	react_hooksutil.WalkFunctionBody(rootFn, func(node *ast.Node) bool {
		if found || node.Pos() <= callNode.End() {
			return false
		}
		if isMutationOfDependency(node, dep) {
			found = true
			return true
		}
		return false
	})
	return found
}

func isMutationOfDependency(node *ast.Node, dep memoDependency) bool {
	stripped := utils.SkipAssertionsAndParens(node)
	if stripped != nil && (stripped.Kind == ast.KindIdentifier || ast.IsAccessExpression(stripped)) &&
		utils.IsWriteReference(stripped) && targetMatchesDependency(stripped, dep) {
		return true
	}

	switch node.Kind {
	case ast.KindBinaryExpression:
		if !ast.IsAssignmentExpression(node, false) {
			return false
		}
		left := node.AsBinaryExpression().Left
		return targetMatchesDependency(left, dep)
	case ast.KindPrefixUnaryExpression:
		expr := node.AsPrefixUnaryExpression()
		if expr == nil || (expr.Operator != ast.KindPlusPlusToken && expr.Operator != ast.KindMinusMinusToken) {
			return false
		}
		return targetMatchesDependency(expr.Operand, dep)
	case ast.KindPostfixUnaryExpression:
		expr := node.AsPostfixUnaryExpression()
		if expr == nil {
			return false
		}
		return targetMatchesDependency(expr.Operand, dep)
	case ast.KindDeleteExpression:
		return targetMatchesDependency(node.AsDeleteExpression().Expression, dep)
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		if call == nil {
			return false
		}
		callee := utils.SkipAssertionsAndParens(call.Expression)
		if callee == nil || !ast.IsAccessExpression(callee) {
			return false
		}
		name, ok := utils.AccessExpressionStaticName(callee)
		if !ok || !mutatingMethodNames[name] {
			return false
		}
		return targetMatchesDependency(utils.AccessExpressionObject(callee), dep)
	}
	return false
}

var mutatingMethodNames = map[string]bool{
	"copyWithin": true,
	"fill":       true,
	"pop":        true,
	"push":       true,
	"reverse":    true,
	"shift":      true,
	"sort":       true,
	"splice":     true,
	"unshift":    true,
	"set":        true,
	"add":        true,
	"delete":     true,
	"clear":      true,
}

func targetMatchesDependency(target *ast.Node, dep memoDependency) bool {
	targetDep, ok := parseAccessDependency(target)
	if !ok {
		return false
	}
	if targetDep.root != dep.root {
		return false
	}
	if len(targetDep.path) < len(dep.path) {
		return false
	}
	for i, part := range dep.path {
		if targetDep.path[i].property != part.property {
			return false
		}
	}
	return true
}
