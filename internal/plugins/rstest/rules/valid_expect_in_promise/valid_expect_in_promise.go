package valid_expect_in_promise

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type chainGroup struct {
	root            *ast.Node
	function        *ast.Node
	owner           *chainGroup
	directAssertion bool
	safe            bool
	reachable       bool
	binding         *chainBinding
	reportNode      *ast.Node
}

type chainBinding struct {
	target  *ast.Node
	symbol  *ast.Symbol
	extends bool
}

type flowEventKind uint8

const (
	flowBind flowEventKind = iota
	flowOverwrite
	flowConsume
)

type flowEvent struct {
	kind      flowEventKind
	symbol    *ast.Symbol
	candidate *chainGroup
	extends   bool
}

type analyzer struct {
	ctx       rule.RuleContext
	analysis  *rstestUtils.RstestCallAnalysis
	groups    map[groupKey]*chainGroup
	ordered   []*chainGroup
	analyzing map[functionOwner]bool
	message   rule.RuleMessage
}

type groupKey struct {
	root  *ast.Node
	owner *chainGroup
}

type functionOwner struct {
	function *ast.Node
	owner    *chainGroup
}

func buildFloatingPromiseMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectInFloatingPromise",
		Description: "This promise should either be returned or awaited to ensure the assertions in its chain are called",
	}
}

var ValidExpectInPromiseRule = rule.Rule{
	Name:   "rstest/valid-expect-in-promise",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		analyzer := &analyzer{
			ctx:       ctx,
			analysis:  analysis,
			groups:    map[groupKey]*chainGroup{},
			analyzing: map[functionOwner]bool{},
			message:   buildFloatingPromiseMessage(),
		}
		for function := range analysis.Callbacks().Functions {
			analyzer.collectFunction(function, nil)
		}
		analyzer.resolveConsumption()
		analyzer.propagateAndReport()
		return rule.RuleListeners{}
	},
}

func (a *analyzer) collectFunction(function *ast.Node, owner *chainGroup) {
	if function == nil {
		return
	}
	key := functionOwner{function: function, owner: owner}
	if a.analyzing[key] {
		return
	}
	a.analyzing[key] = true
	defer delete(a.analyzing, key)

	body := function.Body()
	if body == nil {
		return
	}
	a.walk(body, function, owner)
}

func (a *analyzer) walk(node, currentFunction *ast.Node, owner *chainGroup) {
	if node == nil {
		return
	}
	if node != currentFunction && testFramework.IsFunction(node) {
		return
	}
	if node.Kind == ast.KindCallExpression {
		if testFramework.IsPromiseChainCall(node) {
			group := a.groupFor(node, currentFunction, owner)
			call := node.AsCallExpression()
			a.walk(call.Expression, currentFunction, owner)
			if call.Arguments != nil {
				for _, argument := range call.Arguments.Nodes {
					if function := a.resolveFunction(argument); function != nil {
						a.collectFunction(function, group)
						continue
					}
					a.walk(argument, currentFunction, owner)
				}
			}
			return
		}
		if owner != nil && a.isAssertionCall(node) {
			owner.directAssertion = true
		}
	}
	node.ForEachChild(func(child *ast.Node) bool {
		a.walk(child, currentFunction, owner)
		return false
	})
}

func (a *analyzer) groupFor(call, function *ast.Node, owner *chainGroup) *chainGroup {
	root := topMostPromiseChainCall(call)
	key := groupKey{root: root, owner: owner}
	if group := a.groups[key]; group != nil {
		return group
	}
	group := &chainGroup{
		root:       root,
		function:   function,
		owner:      owner,
		reportNode: promiseReportNode(root, function),
	}
	a.groups[key] = group
	a.ordered = append(a.ordered, group)
	return group
}

func topMostPromiseChainCall(node *ast.Node) *ast.Node {
	top := node
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}
		if parent.Kind != ast.KindPropertyAccessExpression &&
			parent.Kind != ast.KindElementAccessExpression {
			break
		}
		var receiver *ast.Node
		if parent.Kind == ast.KindPropertyAccessExpression {
			receiver = parent.AsPropertyAccessExpression().Expression
		} else {
			receiver = parent.AsElementAccessExpression().Expression
		}
		if receiver != current || parent.Parent == nil ||
			parent.Parent.Kind != ast.KindCallExpression ||
			parent.Parent.AsCallExpression().Expression != parent ||
			!testFramework.IsPromiseChainCall(parent.Parent) {
			break
		}
		top = parent.Parent
		current = parent.Parent
	}
	return top
}

func promiseReportNode(root, function *ast.Node) *ast.Node {
	current := root
	for current != nil && current.Parent != nil && current.Parent != function {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent
		case ast.KindExpressionStatement, ast.KindVariableDeclaration:
			return parent
		default:
			if ast.IsAssignmentExpression(parent, true) {
				return parent
			}
			return current
		}
	}
	return root
}

func (a *analyzer) resolveFunction(node *ast.Node) *ast.Node {
	node = ast.SkipParentheses(node)
	if node == nil {
		return nil
	}
	if ast.IsFunctionExpressionOrArrowFunction(node) {
		return node
	}
	if node.Kind != ast.KindIdentifier {
		return nil
	}
	declaration := internalUtils.GetDeclaration(a.ctx.TypeChecker, node)
	if declaration == nil {
		return nil
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return declaration
	case ast.KindVariableDeclaration:
		initializer := declaration.AsVariableDeclaration().Initializer
		initializer = ast.SkipParentheses(initializer)
		if ast.IsFunctionExpressionOrArrowFunction(initializer) {
			return initializer
		}
	}
	return nil
}

func (a *analyzer) isAssertionCall(node *ast.Node) bool {
	return a.analysis.IsExpectCall(node) ||
		isRstestAssertCall(node, a.ctx)
}

func (a *analyzer) resolveConsumption() {
	groupsByFunction := map[*ast.Node][]*chainGroup{}
	for _, group := range a.ordered {
		groupsByFunction[group.function] = append(groupsByFunction[group.function], group)
	}
	for function, groups := range groupsByFunction {
		a.resolveFunctionConsumption(function, groups)
	}
}

func (a *analyzer) resolveFunctionConsumption(function *ast.Node, groups []*chainGroup) {
	groupsByRoot := map[*ast.Node][]*chainGroup{}
	groupsByTarget := map[*ast.Node][]*chainGroup{}
	for _, group := range groups {
		groupsByRoot[group.root] = append(groupsByRoot[group.root], group)
		group.binding = a.bindingFor(group)
		if group.binding != nil && group.binding.symbol != nil {
			groupsByTarget[group.binding.target] = append(
				groupsByTarget[group.binding.target],
				group,
			)
			group.safe = true
		} else {
			group.binding = nil
		}
	}

	graph := cfg.Build(function, cfg.Hooks[flowEvent]{
		Expression: func(builder *cfg.Builder[flowEvent], node *ast.Node) {
			if builder.Current().Reachable {
				for _, group := range groupsByRoot[node] {
					group.reachable = true
				}
			}
		},
		Read: func(builder *cfg.Builder[flowEvent], node *ast.Node) {
			if !a.isSafeIdentifierUse(node) {
				return
			}
			if symbol := a.symbolOf(node); symbol != nil {
				builder.Emit(flowEvent{kind: flowConsume, symbol: symbol})
			}
		},
		Write: func(builder *cfg.Builder[flowEvent], node *ast.Node) {
			symbol := a.symbolOf(node)
			if symbol == nil {
				return
			}
			if boundGroups := groupsByTarget[node]; len(boundGroups) > 0 {
				for _, group := range boundGroups {
					builder.Emit(flowEvent{
						kind:      flowBind,
						symbol:    symbol,
						candidate: group,
						extends:   group.binding.extends,
					})
				}
				return
			}
			builder.Emit(flowEvent{kind: flowOverwrite, symbol: symbol})
		},
	})

	a.runDataflow(graph)
	for _, group := range groups {
		if !group.reachable {
			group.safe = true
			continue
		}
		if group.binding == nil {
			group.safe = a.isDirectlyConsumed(group.root, function)
		}
	}
}

func (a *analyzer) bindingFor(group *chainGroup) *chainBinding {
	if target := testFramework.StaticBindingForValue(group.root); target != nil {
		return &chainBinding{
			target: target,
			symbol: a.symbolOf(target),
		}
	}
	current := group.root
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}
		if parent.Kind == ast.KindVariableDeclaration {
			declaration := parent.AsVariableDeclaration()
			if ast.SkipParentheses(declaration.Initializer) != ast.SkipParentheses(current) {
				return nil
			}
			name := ast.SkipParentheses(declaration.Name())
			if name == nil || name.Kind != ast.KindIdentifier {
				return nil
			}
			return &chainBinding{target: name, symbol: a.symbolOf(name)}
		}
		if ast.IsAssignmentExpression(parent, true) {
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken.Kind != ast.KindEqualsToken ||
				ast.SkipParentheses(binary.Right) != ast.SkipParentheses(current) {
				return nil
			}
			target := ast.SkipParentheses(binary.Left)
			if target == nil || target.Kind != ast.KindIdentifier {
				return nil
			}
			symbol := a.symbolOf(target)
			return &chainBinding{
				target:  target,
				symbol:  symbol,
				extends: promiseChainStartsWithSymbol(group.root, symbol, a),
			}
		}
		return nil
	}
	return nil
}

func promiseChainStartsWithSymbol(root *ast.Node, symbol *ast.Symbol, a *analyzer) bool {
	if root == nil || symbol == nil {
		return false
	}
	identifier := testFramework.ResolveFirstIdentifier(root.AsCallExpression().Expression)
	return identifier != nil && a.symbolOf(identifier) == symbol
}

func (a *analyzer) symbolOf(node *ast.Node) *ast.Symbol {
	if node == nil {
		return nil
	}
	if symbol := node.Symbol(); symbol != nil {
		return symbol
	}
	if a.ctx.Refs != nil {
		if symbol := a.ctx.Refs.Resolve(node); symbol != nil {
			return symbol
		}
	}
	if a.ctx.TypeChecker != nil {
		return a.ctx.TypeChecker.GetSymbolAtLocation(node)
	}
	return nil
}

func (a *analyzer) isDirectlyConsumed(root, function *ast.Node) bool {
	if function != nil && function.Kind == ast.KindArrowFunction &&
		ast.SkipParentheses(function.AsArrowFunction().Body) == ast.SkipParentheses(root) {
		return true
	}
	current := root
	for current != nil && current.Parent != nil && current.Parent != function {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent
		case ast.KindAwaitExpression:
			return testFramework.AbruptCompletionPropagatesFailure(parent, function)
		case ast.KindReturnStatement:
			return testFramework.AbruptCompletionPropagatesFailure(parent, function)
		case ast.KindArrowFunction:
			return parent.AsArrowFunction().Body == current
		case ast.KindArrayLiteralExpression:
			if safePromiseAggregatorForExpression(parent, current) {
				return true
			}
			return false
		case ast.KindCallExpression:
			if a.isAsyncExpectSink(parent, root) ||
				safePromiseWrapperForExpression(parent, current) {
				return true
			}
			return false
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional == nil ||
				(conditional.WhenTrue != current && conditional.WhenFalse != current) {
				return false
			}
			current = parent
		case ast.KindBinaryExpression:
			if !promiseValueFlowsThroughBinary(parent, current) {
				return false
			}
			current = parent
		default:
			return false
		}
	}
	if function != nil && function.Kind == ast.KindArrowFunction &&
		ast.SkipParentheses(function.AsArrowFunction().Body) == ast.SkipParentheses(current) {
		return true
	}
	return false
}

func (a *analyzer) isSafeIdentifierUse(identifier *ast.Node) bool {
	if identifier == nil || identifier.Kind != ast.KindIdentifier {
		return false
	}
	current := identifier
	for current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent
		case ast.KindAwaitExpression:
			return testFramework.AbruptCompletionPropagatesFailure(parent, nil)
		case ast.KindReturnStatement:
			return testFramework.AbruptCompletionPropagatesFailure(parent, nil)
		case ast.KindArrayLiteralExpression:
			if safePromiseAggregatorForExpression(parent, current) {
				return true
			}
			return false
		case ast.KindCallExpression:
			if a.isAsyncExpectSink(parent, identifier) ||
				safePromiseWrapperForExpression(parent, current) {
				return true
			}
			return false
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional == nil ||
				(conditional.WhenTrue != current && conditional.WhenFalse != current) {
				return false
			}
			current = parent
		case ast.KindBinaryExpression:
			if !promiseValueFlowsThroughBinary(parent, current) {
				return false
			}
			current = parent
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			current = parent
		default:
			return false
		}
	}
	return false
}

func promiseValueFlowsThroughBinary(parent, current *ast.Node) bool {
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return false
	}
	binary := parent.AsBinaryExpression()
	if binary == nil {
		return false
	}
	switch binary.OperatorToken.Kind {
	case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
		return binary.Left == current || binary.Right == current
	case ast.KindAmpersandAmpersandToken:
		return binary.Right == current
	case ast.KindCommaToken:
		return binary.Right == current
	default:
		return false
	}
}

func safePromiseWrapperForExpression(callNode, value *ast.Node) bool {
	if callNode == nil || callNode.Kind != ast.KindCallExpression {
		return false
	}
	call := callNode.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 ||
		ast.SkipParentheses(call.Arguments.Nodes[0]) != ast.SkipParentheses(value) ||
		testFramework.CalleeChainName(call.Expression) != "Promise.resolve" {
		return false
	}
	return expressionIsAwaitedOrReturned(callNode)
}

func safePromiseAggregatorForExpression(arrayNode, value *ast.Node) bool {
	if arrayNode == nil || arrayNode.Kind != ast.KindArrayLiteralExpression ||
		arrayNode.Parent == nil || arrayNode.Parent.Kind != ast.KindCallExpression {
		return false
	}
	callNode := arrayNode.Parent
	call := callNode.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 ||
		ast.SkipParentheses(call.Arguments.Nodes[0]) != ast.SkipParentheses(arrayNode) {
		return false
	}
	name := testFramework.CalleeChainName(call.Expression)
	if name != "Promise.all" {
		return false
	}
	return expressionIsAwaitedOrReturned(callNode)
}

func expressionIsAwaitedOrReturned(node *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}
		return (parent.Kind == ast.KindAwaitExpression &&
			testFramework.AbruptCompletionPropagatesFailure(parent, nil)) ||
			(parent.Kind == ast.KindReturnStatement &&
				testFramework.AbruptCompletionPropagatesFailure(parent, nil)) ||
			(parent.Kind == ast.KindArrowFunction &&
				parent.AsArrowFunction().Body == current)
	}
	return false
}

func (a *analyzer) isAsyncExpectSink(callNode, identifier *ast.Node) bool {
	if callNode == nil || callNode.Kind != ast.KindCallExpression {
		return false
	}
	top := topMostCallExpressionOnCallee(callNode)
	parsed := a.analysis.ParseExpectCall(top)
	if parsed == nil ||
		(!slices.Contains(parsed.Modifiers, "resolves") &&
			!slices.Contains(parsed.Modifiers, "rejects")) ||
		parsed.Head == nil || parsed.Head.AsCallExpression().Arguments == nil ||
		len(parsed.Head.AsCallExpression().Arguments.Nodes) == 0 {
		return false
	}
	return ast.SkipParentheses(parsed.Head.AsCallExpression().Arguments.Nodes[0]) ==
		ast.SkipParentheses(identifier)
}

func topMostCallExpressionOnCallee(node *ast.Node) *ast.Node {
	top := node
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression != current {
				return top
			}
			current = parent
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression != current {
				return top
			}
			current = parent
		case ast.KindCallExpression:
			if parent.AsCallExpression().Expression != current {
				return top
			}
			top = parent
			current = parent
		default:
			return top
		}
	}
	return top
}

func (a *analyzer) runDataflow(graph *cfg.Graph[flowEvent]) {
	type state map[*ast.Symbol]map[*chainGroup]bool
	in := make([]state, len(graph.Blocks))
	queued := make([]bool, len(graph.Blocks))
	queue := []*cfg.Block[flowEvent]{graph.Blocks[0]}
	in[0] = state{}
	queued[0] = true

	clone := func(input state) state {
		result := state{}
		for symbol, candidates := range input {
			copyCandidates := map[*chainGroup]bool{}
			for candidate := range candidates {
				copyCandidates[candidate] = true
			}
			result[symbol] = copyCandidates
		}
		return result
	}
	merge := func(target state, source state) bool {
		changed := false
		for symbol, candidates := range source {
			if target[symbol] == nil {
				target[symbol] = map[*chainGroup]bool{}
			}
			for candidate := range candidates {
				if !target[symbol][candidate] {
					target[symbol][candidate] = true
					changed = true
				}
			}
		}
		return changed
	}

	for len(queue) > 0 {
		block := queue[0]
		queue = queue[1:]
		queued[block.Index()] = false
		if !block.Reachable {
			continue
		}
		current := clone(in[block.Index()])
		for _, event := range block.Events {
			switch event.kind {
			case flowConsume:
				delete(current, event.symbol)
			case flowOverwrite:
				for candidate := range current[event.symbol] {
					candidate.safe = false
				}
				delete(current, event.symbol)
			case flowBind:
				if !event.extends {
					for candidate := range current[event.symbol] {
						candidate.safe = false
					}
					delete(current, event.symbol)
				}
				if current[event.symbol] == nil {
					current[event.symbol] = map[*chainGroup]bool{}
				}
				current[event.symbol][event.candidate] = true
			}
		}

		reachableSuccessor := false
		for _, successor := range block.Successors {
			if !successor.Reachable {
				continue
			}
			reachableSuccessor = true
			firstVisit := in[successor.Index()] == nil
			if firstVisit {
				in[successor.Index()] = state{}
			}
			changed := merge(in[successor.Index()], current)
			if (firstVisit || changed) && !queued[successor.Index()] {
				queue = append(queue, successor)
				queued[successor.Index()] = true
			}
		}
		if !reachableSuccessor {
			for _, candidates := range current {
				for candidate := range candidates {
					candidate.safe = false
				}
			}
		}
	}
}

func (a *analyzer) propagateAndReport() {
	reported := map[*ast.Node]bool{}
	for index := len(a.ordered) - 1; index >= 0; index-- {
		group := a.ordered[index]
		if !group.directAssertion {
			continue
		}
		if group.safe {
			if group.owner != nil {
				group.owner.directAssertion = true
			}
			continue
		}
		if reported[group.reportNode] {
			continue
		}
		reported[group.reportNode] = true
		a.ctx.ReportNode(group.reportNode, a.message)
	}
}

func isRstestAssertCall(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	entries := testFramework.GetMemberEntries(node)
	if len(entries) == 0 {
		return false
	}
	if isDirectImportMetaRstestAssert(node, entries) {
		return true
	}
	root := entries[0].Node
	if root == nil || root.Kind != ast.KindIdentifier {
		return false
	}
	localName := root.Text()
	var symbol *ast.Symbol
	if ctx.TypeChecker != nil {
		symbol = ctx.TypeChecker.GetSymbolAtLocation(root)
	}
	if name, ok := importMetaRstestBindingName(symbol); ok {
		return name == "assert"
	}
	if isImportMetaRstestAlias(symbol) {
		return len(entries) > 1 && entries[1].Name == "assert"
	}
	name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
		localName,
		root,
		symbol,
		ctx.SourceFile,
		rstestUtils.RstestImportModule,
	)
	if name == "assert" {
		return true
	}
	return testFramework.IsModuleNamespaceSymbol(symbol, rstestUtils.RstestImportModule) &&
		len(entries) > 1 && entries[1].Name == "assert"
}

func isDirectImportMetaRstestAssert(
	node *ast.Node,
	entries []testFramework.MemberEntry,
) bool {
	call := node.AsCallExpression()
	if call == nil || len(entries) < 2 {
		return false
	}
	root := testFramework.ResolveFirstIdentifier(call.Expression)
	if root != nil {
		return false
	}
	return entries[0].Name == "rstest" && entries[1].Name == "assert" &&
		containsImportMeta(call.Expression)
}

func containsImportMeta(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindMetaProperty {
		meta := node.AsMetaProperty()
		return meta != nil && meta.Name() != nil && meta.Name().Text() == "meta"
	}
	switch node.Kind {
	case ast.KindCallExpression:
		return containsImportMeta(node.AsCallExpression().Expression)
	case ast.KindPropertyAccessExpression:
		return containsImportMeta(node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		return containsImportMeta(node.AsElementAccessExpression().Expression)
	}
	return false
}

func importMetaRstestBindingName(symbol *ast.Symbol) (string, bool) {
	if symbol == nil {
		return "", false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindBindingElement {
			continue
		}
		variable := internalUtils.EnclosingVariableDeclarationOfBindingElement(declaration)
		if variable == nil || !isImportMetaRstest(variable.AsVariableDeclaration().Initializer) {
			continue
		}
		binding := declaration.AsBindingElement()
		if binding.PropertyName != nil {
			if name, ok := internalUtils.GetStaticStringLiteralValue(binding.PropertyName); ok {
				return name, true
			}
			if binding.PropertyName.Kind == ast.KindIdentifier {
				return binding.PropertyName.Text(), true
			}
		}
		if binding.Name() != nil && binding.Name().Kind == ast.KindIdentifier {
			return binding.Name().Text(), true
		}
	}
	return "", false
}

func isImportMetaRstestAlias(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
			continue
		}
		if isImportMetaRstest(declaration.AsVariableDeclaration().Initializer) {
			return true
		}
	}
	return false
}

func isImportMetaRstest(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	var expression *ast.Node
	var name string
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		expression = property.Expression
		if property.Name() != nil {
			name = property.Name().Text()
		}
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		expression = element.Expression
		name, _ = internalUtils.GetStaticStringLiteralValue(
			ast.SkipParentheses(element.ArgumentExpression),
		)
	default:
		return false
	}
	return name == "rstest" && containsImportMeta(expression)
}
