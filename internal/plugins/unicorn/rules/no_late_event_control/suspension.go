package no_late_event_control

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var (
	afterSuspension = rule.RuleMessage{
		Id:          "after-suspension",
		Description: "has no effect after the handler suspends. Call it before `await` or `yield` while the event is still being dispatched.",
	}
	inGenerator = rule.RuleMessage{
		Id:          "in-generator",
		Description: "is not called during event dispatch inside a generator function. Use a normal function handler instead.",
	}
	inNestedFunction = rule.RuleMessage{
		Id:          "in-nested-function",
		Description: "may run after synchronous event dispatch has finished. Call it in the outer handler before deferring work.",
	}
)

type handlerTiming struct {
	suspended map[*ast.Node]bool
	loops     map[*ast.Node]bool
}

func (timing *handlerTiming) markSuspended(node *ast.Node) {
	if handler := enclosingFunction(node); handler != nil {
		if timing.suspended == nil {
			timing.suspended = make(map[*ast.Node]bool)
		}
		timing.suspended[handler] = true
	}
}

func (timing *handlerTiming) lateCall(node, handler *ast.Node) rule.RuleMessage {
	if enclosingFunction(node) != handler {
		return inNestedFunction
	}
	flags := ast.GetFunctionFlags(handler)
	if flags&ast.FunctionFlagsGenerator != 0 {
		return inGenerator
	}
	if timing.suspended[handler] || flags&ast.FunctionFlagsAsync != 0 && timing.isInsideSuspendingLoop(node, handler) {
		return afterSuspension
	}
	return rule.RuleMessage{}
}

func enclosingFunction(node *ast.Node) *ast.Node {
	for child, parent := node, node.Parent; parent != nil; child, parent = parent, parent.Parent {
		if ast.IsFunctionLikeDeclaration(parent) && !isEvaluatedOutsideFunction(parent, child) {
			return parent
		}
	}
	return nil
}

// A method's computed name and decorators are evaluated in the enclosing code.
func isEvaluatedOutsideFunction(function, child *ast.Node) bool {
	return child == function.Name() || ast.IsDecorator(child)
}

func (timing *handlerTiming) isInsideSuspendingLoop(node, handler *ast.Node) bool {
	for child, parent := node, node.Parent; parent != nil && parent != handler; child, parent = parent, parent.Parent {
		if !isRepeatedLoopPart(parent, child) {
			continue
		}
		suspends, known := timing.loops[parent]
		if !known {
			suspends = containsSuspension(parent)
			if timing.loops == nil {
				timing.loops = make(map[*ast.Node]bool)
			}
			timing.loops[parent] = suspends
		}
		if suspends {
			return true
		}
	}
	return false
}

func isRepeatedLoopPart(loop, child *ast.Node) bool {
	switch loop.Kind {
	case ast.KindForStatement:
		return child != loop.Initializer()
	case ast.KindForInStatement, ast.KindForOfStatement:
		return child != loop.Expression()
	case ast.KindWhileStatement, ast.KindDoStatement:
		return true
	}
	return false
}

func containsSuspension(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindAwaitExpression, ast.KindYieldExpression:
		return true
	case ast.KindForOfStatement:
		if isAsyncIteration(node) {
			return true
		}
	}
	if ast.IsFunctionLikeDeclaration(node) {
		return node.ForEachChild(func(child *ast.Node) bool {
			return isEvaluatedOutsideFunction(node, child) && containsSuspension(child)
		})
	}
	return node.ForEachChild(containsSuspension)
}

func isAsyncIteration(node *ast.Node) bool {
	return node.AsForInOrOfStatement().AwaitModifier != nil
}
