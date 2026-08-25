package max_expects

import (
	"testing"

	"github.com/microsoft/typescript-go/shim/ast"
)

// popIfOwned and deactivateIfOwned mirror the exit listeners, which pair enter
// with exit purely through node ownership.
func popIfOwned(stack *frameStack, node *ast.Node) bool {
	if stack.top().owner != node {
		return false
	}
	stack.pop()
	return true
}

func deactivateIfOwned(stack *frameStack, node *ast.Node) bool {
	top := stack.top()
	if top.activationOwner != node {
		return false
	}
	top.active = false
	top.activationOwner = nil
	return true
}

func TestOwnershipPairsNestedFrames(t *testing.T) {
	registration, callback, detached := &ast.Node{}, &ast.Node{}, &ast.Node{}

	stack := newFrameStack()
	stack.push(frameRegistrationFallback, false, registration, nil)
	stack.push(frameTestCallback, true, callback, nil)
	stack.push(frameDetachedFunction, true, detached, nil)

	for _, owner := range []*ast.Node{detached, callback, registration} {
		if !popIfOwned(&stack, owner) {
			t.Fatalf("owner %p did not pop the frame it pushed", owner)
		}
	}
	if len(stack.frames) != 1 {
		t.Fatalf("frames after unwinding = %d, want only the sentinel", len(stack.frames))
	}
}

func TestOwnershipIgnoresNodesThatPushedNothing(t *testing.T) {
	registration, ordinary := &ast.Node{}, &ast.Node{}

	stack := newFrameStack()
	stack.push(frameRegistrationFallback, false, registration, nil)

	if popIfOwned(&stack, ordinary) {
		t.Fatal("a node that pushed no frame must not pop one")
	}
	if deactivateIfOwned(&stack, ordinary) {
		t.Fatal("a node that activated nothing must not deactivate a frame")
	}
	if len(stack.frames) != 2 {
		t.Fatalf("frames = %d, want the registration frame untouched", len(stack.frames))
	}
}

func TestOwnershipCannotUnwindTheSentinel(t *testing.T) {
	stack := newFrameStack()
	if stack.top().owner != nil {
		t.Fatal("the sentinel frame must have no owner")
	}
	if popIfOwned(&stack, &ast.Node{}) {
		t.Fatal("an exit without a matching enter must not unwind the sentinel")
	}
}

func TestActivationDeactivatesWithoutPopping(t *testing.T) {
	registration, callback := &ast.Node{}, &ast.Node{}
	root := &ast.Node{}

	stack := newFrameStack()
	stack.push(frameRegistrationFallback, false, registration, []*ast.Node{root})
	callback.Parent = root

	if !canActivateRegistrationFallback(stack.top(), callback) {
		t.Fatal("a function inside the candidate root must activate the frame")
	}
	stack.top().active = true
	stack.top().activationOwner = callback

	if canActivateRegistrationFallback(stack.top(), callback) {
		t.Fatal("an already active frame must not activate a second time")
	}
	if popIfOwned(&stack, callback) {
		t.Fatal("the activating function must not own the frame it activated")
	}
	if !deactivateIfOwned(&stack, callback) {
		t.Fatal("the activating function must deactivate the frame on exit")
	}
	if len(stack.frames) != 2 || stack.top().owner != registration {
		t.Fatal("deactivation must leave the frame and its ownership in place")
	}

	// A later sibling in the same candidate root activates the frame again.
	sibling := &ast.Node{Parent: root}
	if !canActivateRegistrationFallback(stack.top(), sibling) {
		t.Fatal("a deactivated frame must be activatable again")
	}
}
