// Package valid_expect_in_promise holds the framework-neutral promise-chain
// collection and consumption analysis shared by Jest and Rstest. Each function
// body and promise-chain call is collected once; assertion reachability then
// flows over the finite graph between handler summaries and their owner chains.
// Framework adapters inject callback discovery, assertion provenance, and async
// assertion sinks through Runtime.
package valid_expect_in_promise

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/cfg"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type Runtime struct {
	TestCallbackFunctions    map[*ast.Node]bool
	IgnoredCallbackFunctions map[*ast.Node]bool
	IsAssertionCall          func(*ast.Node) bool
	IsAsyncAssertionSink     func(call, value *ast.Node) bool
}

type Config struct {
	Name               string
	MessageDescription string
	Prepare            func(rule.RuleContext) Runtime
}

type chainGroup struct {
	root         *ast.Node
	function     *functionSummary
	hasAssertion bool
	safe         bool
	reachable    bool
	binding      *chainBinding
	reportNode   *ast.Node
}

type functionSummary struct {
	function        *ast.Node
	owners          map[*chainGroup]struct{}
	directAssertion bool
	hasAssertion    bool
	collected       bool
}

type chainBinding struct {
	target  *ast.Node
	symbol  *ast.Symbol
	extends bool
	// aggregate marks a binding that holds the chain inside an array literal,
	// so only an element-wise consumption of the binding consumes the chain.
	aggregate bool
}

type flowEventKind uint8

const (
	flowBind flowEventKind = iota
	flowOverwrite
	flowConsume
	flowFailureExit
)

type flowEvent struct {
	kind      flowEventKind
	symbol    *ast.Symbol
	candidate *chainGroup
	extends   bool
	// aggregate marks a consumption that awaits the elements of an iterable
	// rather than the value itself, such as `Promise.all(list)` or
	// `for await (const value of list)`.
	aggregate bool
}

type analyzer struct {
	ctx       rule.RuleContext
	runtime   Runtime
	functions map[*ast.Node]*functionSummary
	groups    map[*ast.Node]*chainGroup
	ordered   []*chainGroup
	message   rule.RuleMessage
}

func buildFloatingPromiseMessage(description string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "expectInFloatingPromise",
		Description: description,
	}
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			analyzer := &analyzer{
				ctx:       ctx,
				runtime:   runtime,
				functions: map[*ast.Node]*functionSummary{},
				groups:    map[*ast.Node]*chainGroup{},
				message:   buildFloatingPromiseMessage(config.MessageDescription),
			}
			for function := range runtime.TestCallbackFunctions {
				if runtime.IgnoredCallbackFunctions[function] {
					continue
				}
				analyzer.collectFunction(function, nil)
			}
			analyzer.resolveConsumption()
			analyzer.propagateAndReport()
			return rule.RuleListeners{}
		},
	}
}

func (a *analyzer) collectFunction(function *ast.Node, owner *chainGroup) {
	if function == nil {
		return
	}

	summary := a.functions[function]
	if summary == nil {
		summary = &functionSummary{
			function: function,
			owners:   map[*chainGroup]struct{}{},
		}
		a.functions[function] = summary
	}
	if owner != nil {
		summary.owners[owner] = struct{}{}
	}
	if summary.collected {
		return
	}
	// Publish the summary before walking the body. Recursive and mutually
	// recursive handlers then add an owner edge to the existing summary instead
	// of re-entering the walk.
	summary.collected = true

	body := function.Body()
	if body == nil {
		return
	}
	a.walk(body, summary)
}

func (a *analyzer) walk(node *ast.Node, currentFunction *functionSummary) {
	if node == nil {
		return
	}
	if node != currentFunction.function && testFramework.IsFunction(node) {
		return
	}
	if node.Kind == ast.KindCallExpression {
		if testFramework.IsPromiseChainCall(node) {
			group := a.groupFor(node, currentFunction)
			call := node.AsCallExpression()
			a.walk(call.Expression, currentFunction)
			if call.Arguments != nil {
				for _, argument := range call.Arguments.Nodes {
					if function := a.resolveFunction(argument); function != nil {
						a.collectFunction(function, group)
						continue
					}
					a.walk(argument, currentFunction)
				}
			}
			return
		}
		if a.isAssertionCall(node) {
			currentFunction.directAssertion = true
		}
	}
	node.ForEachChild(func(child *ast.Node) bool {
		a.walk(child, currentFunction)
		return false
	})
}

func (a *analyzer) groupFor(
	call *ast.Node,
	function *functionSummary,
) *chainGroup {
	root := topMostPromiseChainCall(call)
	if group := a.groups[root]; group != nil {
		return group
	}
	group := &chainGroup{
		root:       root,
		function:   function,
		reportNode: promiseReportNode(root, function.function),
	}
	a.groups[root] = group
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
		case ast.KindCallExpression,
			ast.KindPropertyAccessExpression,
			ast.KindElementAccessExpression:
			// A chain handed to another call belongs to the statement that call
			// sits in, so `list.push(chain);` is reported at the statement.
			current = parent
		default:
			if ast.IsAssignmentExpression(parent, false) {
				operator := parent.AsBinaryExpression().OperatorToken.Kind
				// A logical assignment stores the chain, so the assignment is
				// the place it floats from. A compound arithmetic assignment
				// stores something else and leaves the chain itself floating.
				if operator == ast.KindEqualsToken ||
					ast.IsLogicalOrCoalescingAssignmentOperator(operator) {
					return parent
				}
			}
			return current
		}
	}
	return current
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
		if initializer == nil {
			return nil
		}
		initializer = ast.SkipParentheses(initializer)
		if ast.IsFunctionExpressionOrArrowFunction(initializer) {
			return initializer
		}
	}
	return nil
}

func (a *analyzer) isAssertionCall(node *ast.Node) bool {
	return a.runtime.IsAssertionCall != nil && a.runtime.IsAssertionCall(node)
}

func (a *analyzer) resolveConsumption() {
	groupsByFunction := map[*functionSummary][]*chainGroup{}
	for _, group := range a.ordered {
		groupsByFunction[group.function] = append(groupsByFunction[group.function], group)
	}
	for function, groups := range groupsByFunction {
		a.resolveFunctionConsumption(function, groups)
	}
}

func (a *analyzer) resolveFunctionConsumption(
	function *functionSummary,
	groups []*chainGroup,
) {
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

	graph := cfg.Build(function.function, cfg.Hooks[flowEvent]{
		Expression: func(builder *cfg.Builder[flowEvent], node *ast.Node) {
			if builder.Current().Reachable {
				for _, group := range groupsByRoot[node] {
					group.reachable = true
				}
			}
		},
		Read: func(builder *cfg.Builder[flowEvent], node *ast.Node) {
			safe, aggregate := a.isSafeIdentifierUse(node, function.function)
			if !safe {
				return
			}
			if symbol := a.symbolOf(node); symbol != nil {
				builder.Emit(flowEvent{
					kind:      flowConsume,
					symbol:    symbol,
					aggregate: aggregate,
				})
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
			if a.writePreservesPreviousValue(node) {
				// The store may not happen, or may store what the symbol
				// already holds, so an earlier candidate can survive it.
				return
			}
			builder.Emit(flowEvent{kind: flowOverwrite, symbol: symbol})
		},
		Statement: func(builder *cfg.Builder[flowEvent], node *ast.Node) {
			if node.Kind == ast.KindThrowStatement &&
				testFramework.AbruptCompletionPropagatesFailure(node, function.function) {
				builder.Emit(flowEvent{kind: flowFailureExit})
			}
		},
	})

	a.runDataflow(graph)
	for _, group := range groups {
		if !group.reachable {
			group.safe = true
			continue
		}
		// A binding and a direct consumption are two independent ways of being
		// consumed: `await (pending ||= chain)` stores the chain and awaits the
		// value the assignment produced.
		if !group.safe {
			group.safe = a.isDirectlyConsumed(group.root, function.function)
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
	aggregate := false
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
			return &chainBinding{
				target:    name,
				symbol:    a.symbolOf(name),
				aggregate: aggregate,
			}
		}
		if ast.IsAssignmentExpression(parent, false) {
			binary := parent.AsBinaryExpression()
			operator := binary.OperatorToken.Kind
			// A compound arithmetic assignment produces something that is no
			// longer this promise, so it binds nothing. A logical assignment
			// does store the chain, only conditionally.
			logical := ast.IsLogicalOrCoalescingAssignmentOperator(operator)
			preserving := operator == ast.KindBarBarEqualsToken ||
				operator == ast.KindQuestionQuestionEqualsToken
			if (operator != ast.KindEqualsToken && !logical) ||
				ast.SkipParentheses(binary.Right) != ast.SkipParentheses(current) {
				return nil
			}
			target := ast.SkipParentheses(binary.Left)
			if target == nil || target.Kind != ast.KindIdentifier {
				return nil
			}
			symbol := a.symbolOf(target)
			return &chainBinding{
				target: target,
				symbol: symbol,
				// `p ||= chain` and `p ??= chain` leave a promise already held
				// by p in place, so the bind must not kill it, the same way a
				// chain that continues off p does not. `p &&= chain` overwrites
				// that promise, so it kills like a plain assignment.
				extends: preserving ||
					promiseChainStartsWithSymbol(group.root, symbol, a) ||
					a.assignmentPreservesSymbol(binary.Right, symbol),
				aggregate: aggregate,
			}
		}
		if parent.Kind == ast.KindArrayLiteralExpression {
			// `const list = [chain]` stores the chain in list, but only an
			// element-wise consumption of list reaches it.
			aggregate = true
			current = parent
			continue
		}
		if parent.Kind == ast.KindConditionalExpression {
			conditional := parent.AsConditionalExpression()
			if conditional == nil ||
				(conditional.WhenTrue != current && conditional.WhenFalse != current) {
				return nil
			}
			current = parent
			continue
		}
		// `let p = fallback || chain` stores the chain in p on exactly the
		// paths where the chain is evaluated, so the value reaches the binding
		// the same way isDirectlyConsumed carries it up to an await. An
		// assignment parent is left to the branch above.
		if !ast.IsAssignmentExpression(parent, false) &&
			promiseValueFlowsThroughBinary(parent, current) {
			current = parent
			continue
		}
		return nil
	}
	return nil
}

func promiseChainStartsWithSymbol(
	root *ast.Node,
	symbol *ast.Symbol,
	a *analyzer,
) bool {
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
			if safePromiseAggregatorForExpression(parent, current, function) ||
				forAwaitConsumesExpression(parent, function) {
				return true
			}
			return false
		case ast.KindCallExpression:
			if a.isAsyncAssertionSink(parent, root) ||
				safePromiseWrapperForExpression(parent, current, function) {
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

// isSafeIdentifierUse reports whether this use of an identifier consumes the
// value it holds, and whether it does so element-wise. An element-wise
// consumption reaches the promises inside an array the identifier holds rather
// than the array itself, so it consumes exactly the bindings a plain await does
// not.
func (a *analyzer) isSafeIdentifierUse(identifier, function *ast.Node) (bool, bool) {
	if identifier == nil || identifier.Kind != ast.KindIdentifier {
		return false, false
	}
	current := identifier
	// viaElement records that an element access stood between the identifier
	// and the consumption, so the consumption reaches one element of what the
	// identifier holds rather than the value itself.
	viaElement := false
	for current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			current = parent
		case ast.KindAwaitExpression:
			return testFramework.AbruptCompletionPropagatesFailure(parent, function), viaElement
		case ast.KindReturnStatement:
			return testFramework.AbruptCompletionPropagatesFailure(parent, function), viaElement
		case ast.KindForOfStatement:
			return forAwaitConsumesExpression(current, function), true
		case ast.KindArrayLiteralExpression:
			if safePromiseAggregatorForExpression(parent, current, function) {
				return true, false
			}
			return false, false
		case ast.KindSpreadElement:
			// `[...list]` spreads every element into the literal, so a literal
			// that is consumed element-wise consumes what list holds.
			array := parent.Parent
			if array == nil || array.Kind != ast.KindArrayLiteralExpression {
				return false, false
			}
			if safePromiseAggregatorForExpression(array, array, function) ||
				forAwaitConsumesExpression(array, function) {
				return true, true
			}
			return false, false
		case ast.KindCallExpression:
			if a.isAsyncAssertionSink(parent, identifier) ||
				safePromiseWrapperForExpression(parent, current, function) {
				return true, false
			}
			if safePromiseAggregatorForValue(parent, current, function) {
				return true, true
			}
			return false, false
		case ast.KindConditionalExpression:
			conditional := parent.AsConditionalExpression()
			if conditional == nil ||
				(conditional.WhenTrue != current && conditional.WhenFalse != current) {
				return false, false
			}
			current = parent
		case ast.KindBinaryExpression:
			if !promiseValueFlowsThroughBinary(parent, current) {
				return false, false
			}
			current = parent
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression == current {
				viaElement = true
			}
			current = parent
		case ast.KindPropertyAccessExpression:
			current = parent
		default:
			return false, false
		}
	}
	return false, false
}

// writePreservesPreviousValue reports whether an assignment to node can leave
// the value node already holds in place. A promise a candidate left there is
// truthy and non-nullish, so `p ||= chain` and `p ??= chain` never replace it,
// and neither does the longhand `p = p || chain`. `p &&= chain` and
// `p = p && chain` always do, and are not included.
func (a *analyzer) writePreservesPreviousValue(node *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}
		if !ast.IsAssignmentExpression(parent, false) {
			return false
		}
		binary := parent.AsBinaryExpression()
		if binary == nil ||
			ast.SkipParentheses(binary.Left) != ast.SkipParentheses(current) {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return true
		case ast.KindEqualsToken:
			return a.assignmentPreservesSymbol(binary.Right, a.symbolOf(node))
		}
		return false
	}
	return false
}

// assignmentPreservesSymbol reports whether value, assigned to symbol, can
// evaluate to the value symbol already holds. `p = p || chain` keeps a promise
// p holds, because a promise is truthy, so the write must not kill an earlier
// candidate for p.
func (a *analyzer) assignmentPreservesSymbol(value *ast.Node, symbol *ast.Symbol) bool {
	value = ast.SkipParentheses(value)
	if value == nil || symbol == nil {
		return false
	}
	switch value.Kind {
	case ast.KindIdentifier:
		return a.symbolOf(value) == symbol
	case ast.KindConditionalExpression:
		// Both branches must be able to be the old value. `p = c ? p : other`
		// drops it whenever c is falsy.
		conditional := value.AsConditionalExpression()
		return conditional != nil &&
			a.assignmentPreservesSymbol(conditional.WhenTrue, symbol) &&
			a.assignmentPreservesSymbol(conditional.WhenFalse, symbol)
	case ast.KindBinaryExpression:
		binary := value.AsBinaryExpression()
		if binary == nil {
			return false
		}
		switch binary.OperatorToken.Kind {
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			// Only the left operand survives: a promise is truthy and
			// non-nullish, so `p = p || other` keeps p, while `p = other || p`
			// drops it for any truthy other. `&&` is absent because neither
			// operand survives it.
			return a.assignmentPreservesSymbol(binary.Left, symbol)
		case ast.KindCommaToken:
			return a.assignmentPreservesSymbol(binary.Right, symbol)
		}
	}
	return false
}

// forAwaitConsumesExpression reports whether node is the iterable of a
// `for await` loop, which awaits every element and lets a rejection leave the
// loop the way an await does.
func forAwaitConsumesExpression(node, function *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}
		if parent.Kind != ast.KindForOfStatement {
			return false
		}
		statement := parent.AsForInOrOfStatement()
		return statement != nil && statement.AwaitModifier != nil &&
			statement.Expression == current &&
			testFramework.AbruptCompletionPropagatesFailure(parent, function)
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
		// `await (p ||= chain)` consumes the chain the assignment stored, so
		// the value carries on through the right-hand side of a logical
		// assignment too.
		return ast.IsLogicalOrCoalescingAssignmentOperator(binary.OperatorToken.Kind) &&
			binary.Right == current
	}
}

func safePromiseWrapperForExpression(callNode, value, function *ast.Node) bool {
	if callNode == nil || callNode.Kind != ast.KindCallExpression {
		return false
	}
	call := callNode.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 ||
		ast.SkipParentheses(call.Arguments.Nodes[0]) != ast.SkipParentheses(value) ||
		testFramework.CalleeChainName(call.Expression) != "Promise.resolve" {
		return false
	}
	return expressionIsAwaitedOrReturned(callNode, function)
}

func safePromiseAggregatorForExpression(arrayNode, value, function *ast.Node) bool {
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
	return expressionIsAwaitedOrReturned(callNode, function)
}

// safePromiseAggregatorForValue reports whether callNode is an awaited or
// returned `Promise.all(value)`. Unlike safePromiseAggregatorForExpression the
// argument is not an array literal, so this only reaches promises held inside
// whatever value holds.
func safePromiseAggregatorForValue(callNode, value, function *ast.Node) bool {
	if callNode == nil || callNode.Kind != ast.KindCallExpression {
		return false
	}
	call := callNode.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 ||
		ast.SkipParentheses(call.Arguments.Nodes[0]) != ast.SkipParentheses(value) ||
		testFramework.CalleeChainName(call.Expression) != "Promise.all" {
		return false
	}
	return expressionIsAwaitedOrReturned(callNode, function)
}

func expressionIsAwaitedOrReturned(node, function *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}
		return (parent.Kind == ast.KindAwaitExpression &&
			testFramework.AbruptCompletionPropagatesFailure(parent, function)) ||
			(parent.Kind == ast.KindReturnStatement &&
				testFramework.AbruptCompletionPropagatesFailure(parent, function)) ||
			(parent.Kind == ast.KindArrowFunction &&
				parent.AsArrowFunction().Body == current)
	}
	return false
}

func (a *analyzer) isAsyncAssertionSink(callNode, value *ast.Node) bool {
	return a.runtime.IsAsyncAssertionSink != nil &&
		a.runtime.IsAsyncAssertionSink(callNode, value)
}

func TopMostCallExpressionOnCallee(node *ast.Node) *ast.Node {
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

func LeftMostCallExpression(node *ast.Node) *ast.Node {
	var leftMost *ast.Node
	current := node
	for current != nil {
		current = ast.SkipParentheses(current)
		if current == nil {
			break
		}
		switch current.Kind {
		case ast.KindCallExpression:
			leftMost = current
			current = current.AsCallExpression().Expression
		case ast.KindPropertyAccessExpression:
			current = current.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			current = current.AsElementAccessExpression().Expression
		default:
			return leftMost
		}
	}
	return leftMost
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
			case flowFailureExit:
				current = state{}
			case flowConsume:
				// An element-wise consumption reaches a chain held in an array
				// and nothing else; every other consumption is the reverse.
				candidates := current[event.symbol]
				for candidate := range candidates {
					if candidate.binding != nil &&
						candidate.binding.aggregate != event.aggregate {
						continue
					}
					delete(candidates, candidate)
				}
				if len(candidates) == 0 {
					delete(current, event.symbol)
				}
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
			if in[successor.Index()] == nil {
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
	functionQueue := make([]*functionSummary, 0, len(a.functions))
	groupQueue := make([]*chainGroup, 0, len(a.ordered))
	for _, function := range a.functions {
		if !function.directAssertion {
			continue
		}
		function.hasAssertion = true
		functionQueue = append(functionQueue, function)
	}

	for len(functionQueue) > 0 || len(groupQueue) > 0 {
		if len(functionQueue) > 0 {
			function := functionQueue[0]
			functionQueue = functionQueue[1:]
			for owner := range function.owners {
				if owner.hasAssertion {
					continue
				}
				owner.hasAssertion = true
				groupQueue = append(groupQueue, owner)
			}
			continue
		}

		group := groupQueue[0]
		groupQueue = groupQueue[1:]
		if !group.safe || group.function.hasAssertion {
			continue
		}
		group.function.hasAssertion = true
		functionQueue = append(functionQueue, group.function)
	}

	// Every floating chain is reported once. Two chains handed to the same call
	// share a report node and are two separate floating promises.
	for index := len(a.ordered) - 1; index >= 0; index-- {
		group := a.ordered[index]
		if !group.hasAssertion || group.safe {
			continue
		}
		a.ctx.ReportNode(group.reportNode, a.message)
	}
}
