package max_expects

import (
	_ "embed"
	"fmt"
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed max_expects.schema.json
var schemaJSON []byte

const defaultMax = 5

type options struct {
	Max int
}

type frameKind uint8

const (
	frameInactive frameKind = iota
	frameTestCallback
	frameDetachedFunction
	frameRegistrationFallback
)

type countFrame struct {
	kind                   frameKind
	count                  int
	active                 bool
	registration           *ast.Node
	callbackCandidateRoots []*ast.Node
}

type frameStack struct {
	frames []countFrame
}

type functionPushKind uint8

const (
	functionPushNone functionPushKind = iota
	functionPushTestCallback
	functionPushDetached
	functionActivateRegistrationFallback
)

func parseOptions(rawOptions []any) options {
	opts := options{Max: defaultMax}
	if len(rawOptions) == 0 {
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]any)
	if maxVal, ok := internalUtils.CoerceIntegral(optsMap["max"]); ok && maxVal >= 1 {
		opts.Max = maxVal
	}

	return opts
}

func buildExceededMaxAssertionMessage(count, maxAllowed int) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "exceededMaxAssertion",
		Description: fmt.Sprintf("Too many assertion calls (%d) - maximum allowed is %d", count, maxAllowed),
		Data: map[string]string{
			"count": strconv.Itoa(count),
			"max":   strconv.Itoa(maxAllowed),
		},
	}
}

func newFrameStack() frameStack {
	return frameStack{
		frames: []countFrame{{
			kind:   frameInactive,
			active: false,
		}},
	}
}

func (stack *frameStack) top() *countFrame {
	return &stack.frames[len(stack.frames)-1]
}

func (stack *frameStack) push(kind frameKind, active bool, registration *ast.Node, callbackCandidateRoots []*ast.Node) {
	stack.frames = append(stack.frames, countFrame{
		kind:                   kind,
		count:                  0,
		active:                 active,
		registration:           registration,
		callbackCandidateRoots: callbackCandidateRoots,
	})
}

func (stack *frameStack) pop() {
	if len(stack.frames) <= 1 {
		return
	}
	stack.frames = stack.frames[:len(stack.frames)-1]
}

func (stack *frameStack) increment() int {
	top := stack.top()
	if !top.active {
		return 0
	}
	top.count++
	return top.count
}

func isCountedExpectCall(parsed *rstestUtils.ParsedRstestExpectCall) bool {
	if parsed == nil ||
		parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
		parsed.MatcherEntry == nil ||
		parsed.Head == nil {
		return false
	}

	switch parsed.Entry {
	case rstestUtils.RstestExpectEntryCall,
		rstestUtils.RstestExpectEntrySoft,
		rstestUtils.RstestExpectEntryPoll,
		rstestUtils.RstestExpectEntryElement:
		return true
	default:
		return false
	}
}

func functionPushForNode(node *ast.Node, callbacks map[*ast.Node]bool) functionPushKind {
	if node == nil || node.Body() == nil {
		return functionPushNone
	}
	if callbacks[node] {
		return functionPushTestCallback
	}

	switch node.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		parent := node.Parent
		for parent != nil && ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions) {
			parent = parent.Parent
		}
		if parent != nil && parent.Kind == ast.KindCallExpression {
			return functionPushNone
		}
		return functionPushDetached
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return functionPushDetached
	case ast.KindFunctionDeclaration:
		return functionPushNone
	default:
		return functionPushNone
	}
}

func isObjectLiteralLike(node *ast.Node) bool {
	node = internalUtils.SkipAssertionsAndParens(node)
	return node != nil && node.Kind == ast.KindObjectLiteralExpression
}

func containsFallbackFunction(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if ast.IsFunctionExpressionOrArrowFunction(node) && node.Body() != nil {
		return true
	}
	found := false
	node.ForEachChild(func(child *ast.Node) bool {
		if containsFallbackFunction(child) {
			found = true
			return true
		}
		return false
	})
	return found
}

func registrationFallbackCandidateRoots(call *ast.CallExpression) []*ast.Node {
	if call == nil || call.Arguments == nil {
		return nil
	}

	args := call.Arguments.Nodes
	if len(args) < 2 || len(args) > 3 {
		return nil
	}

	candidateRoot := func(node *ast.Node) *ast.Node {
		node = internalUtils.SkipAssertionsAndParens(node)
		if node == nil || isObjectLiteralLike(node) {
			return nil
		}
		return node
	}

	if len(args) == 2 {
		if root := candidateRoot(args[1]); root != nil {
			return []*ast.Node{root}
		}
		return nil
	}

	if isObjectLiteralLike(args[1]) {
		if root := candidateRoot(args[2]); root != nil {
			return []*ast.Node{root}
		}
		return nil
	}

	// Prefer a function nested in the third argument: it identifies `(name,
	// options, callback)`, including when options is a variable or arbitrary
	// expression. Otherwise a function nested in the second argument identifies
	// a wrapped callback-first registration whose third argument is a timeout.
	// With no inline function in either position, keeping the overload-shaped
	// third root is harmless because it cannot activate the fallback frame.
	callbackIndex := 2
	if !containsFallbackFunction(args[2]) && containsFallbackFunction(args[1]) {
		callbackIndex = 1
	}
	if root := candidateRoot(args[callbackIndex]); root != nil {
		return []*ast.Node{root}
	}
	return nil
}

func nodeWithinSubtree(node *ast.Node, root *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == root {
			return true
		}
	}
	return false
}

func canActivateRegistrationFallback(frame *countFrame, node *ast.Node) bool {
	if frame == nil || frame.kind != frameRegistrationFallback || frame.active || frame.registration == nil {
		return false
	}
	for _, root := range frame.callbackCandidateRoots {
		if root != nil && nodeWithinSubtree(node, root) {
			return true
		}
	}
	return false
}

var MaxExpectsRule = rule.Rule{
	Name:   "rstest/max-expects",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		callbacks := analysis.Callbacks().Functions

		stack := newFrameStack()
		pushedFunctions := map[*ast.Node]functionPushKind{}
		pushedRegistrations := map[*ast.Node]bool{}

		enterFunction := func(node *ast.Node) {
			kind := functionPushForNode(node, callbacks)
			if kind == functionPushNone &&
				node != nil &&
				node.Body() != nil &&
				canActivateRegistrationFallback(stack.top(), node) {
				kind = functionActivateRegistrationFallback
			}

			switch kind {
			case functionPushTestCallback:
				stack.push(frameTestCallback, true, nil, nil)
				pushedFunctions[node] = functionPushTestCallback
			case functionPushDetached:
				stack.push(frameDetachedFunction, stack.top().active, nil, nil)
				pushedFunctions[node] = functionPushDetached
			case functionActivateRegistrationFallback:
				stack.top().active = true
				pushedFunctions[node] = functionActivateRegistrationFallback
			}
		}

		exitFunction := func(node *ast.Node) {
			switch pushedFunctions[node] {
			case functionPushNone:
				return
			case functionActivateRegistrationFallback:
				delete(pushedFunctions, node)
				stack.top().active = false
				return
			}
			delete(pushedFunctions, node)
			stack.pop()
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                      enterFunction,
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			ast.KindFunctionExpression:                       enterFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			ast.KindArrowFunction:                            enterFunction,
			rule.ListenerOnExit(ast.KindArrowFunction):       exitFunction,
			ast.KindMethodDeclaration:                        enterFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,
			ast.KindGetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):         exitFunction,
			ast.KindSetAccessor:                              enterFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):         exitFunction,
			ast.KindConstructor:                              enterFunction,
			rule.ListenerOnExit(ast.KindConstructor):         exitFunction,
			ast.KindCallExpression: func(node *ast.Node) {
				if analysis.ParseTestCall(node) != nil {
					stack.push(
						frameRegistrationFallback,
						false,
						node,
						registrationFallbackCandidateRoots(node.AsCallExpression()),
					)
					pushedRegistrations[node] = true
					return
				}

				parsed := analysis.ParseExpectCall(node)
				if !isCountedExpectCall(parsed) {
					return
				}

				count := stack.increment()
				if count > opts.Max {
					ctx.ReportNode(node, buildExceededMaxAssertionMessage(count, opts.Max))
				}
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if !pushedRegistrations[node] {
					return
				}
				delete(pushedRegistrations, node)
				stack.pop()
			},
		}
	},
}
