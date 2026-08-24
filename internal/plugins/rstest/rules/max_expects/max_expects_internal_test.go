package max_expects

import "testing"

func TestFunctionEventStackPairsNestedActions(t *testing.T) {
	events := functionEventStack{}
	events.push(functionPushNone)
	events.push(functionPushTestCallback)
	events.push(functionActivateRegistrationFallback)
	events.push(functionPushDetached)

	want := []functionPushKind{
		functionPushDetached,
		functionActivateRegistrationFallback,
		functionPushTestCallback,
		functionPushNone,
	}
	for index, expected := range want {
		if got := events.pop(); got != expected {
			t.Fatalf("pop %d = %d, want %d", index, got, expected)
		}
	}
	if got := events.pop(); got != functionPushNone {
		t.Fatalf("underflow pop = %d, want functionPushNone", got)
	}
}

func TestCallEventStackPairsRegistrationFrames(t *testing.T) {
	events := callEventStack{}
	events.push(false)
	events.push(false)
	events.markRegistrationFrame()
	events.push(false)

	want := []bool{false, true, false}
	for index, expected := range want {
		if got := events.pop(); got != expected {
			t.Fatalf("pop %d = %t, want %t", index, got, expected)
		}
	}
	if events.pop() {
		t.Fatal("underflow pop must not claim a registration frame")
	}
}

func TestCallEventStackIgnoresMarkWithoutEnter(t *testing.T) {
	events := callEventStack{}
	events.markRegistrationFrame()
	if events.pop() {
		t.Fatal("mark without an enter event must stay a no-op")
	}
}

func TestFunctionEventStackPairsOverflow(t *testing.T) {
	events := functionEventStack{}
	for index := 0; index < inlineFunctionEventCapacity+3; index++ {
		events.push(functionPushKind(index%int(functionActivateRegistrationFallback) + 1))
	}
	for index := inlineFunctionEventCapacity + 2; index >= 0; index-- {
		expected := functionPushKind(index%int(functionActivateRegistrationFallback) + 1)
		if got := events.pop(); got != expected {
			t.Fatalf("overflow pop %d = %d, want %d", index, got, expected)
		}
	}
}

func TestCallEventStackMarksOverflow(t *testing.T) {
	events := callEventStack{}
	for range inlineCallEventCapacity + 2 {
		events.push(false)
	}
	events.markRegistrationFrame()
	if !events.pop() {
		t.Fatal("mark must update the top overflow event")
	}
	for range inlineCallEventCapacity + 1 {
		if events.pop() {
			t.Fatal("mark must not affect an earlier event")
		}
	}
}
