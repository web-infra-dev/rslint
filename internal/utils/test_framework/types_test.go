package test_framework

import "testing"

func TestHooksOrder(t *testing.T) {
	if len(HooksOrder) != 4 {
		t.Fatalf("expected 4 hooks in order table, got %d", len(HooksOrder))
	}

	for i, name := range []string{"beforeAll", "beforeEach", "afterEach", "afterAll"} {
		if got := HookOrderIndex(name); got != i {
			t.Errorf("HookOrderIndex(%q) = %d, want %d", name, got, i)
		}
		if !IsHookName(name) {
			t.Errorf("IsHookName(%q) = false, want true", name)
		}
	}

	// onTestFinished / onTestFailed are Rstest execution-time APIs, and
	// test / describe are registrations; none of them is a hook.
	for _, name := range []string{"onTestFinished", "onTestFailed", "test", "describe", ""} {
		if got := HookOrderIndex(name); got != -1 {
			t.Errorf("HookOrderIndex(%q) = %d, want -1", name, got)
		}
		if IsHookName(name) {
			t.Errorf("IsHookName(%q) = true, want false", name)
		}
	}
}

func TestIsCallOfKind(t *testing.T) {
	test := &ParsedCall{Kind: FnKindTest}

	tests := []struct {
		name   string
		parsed *ParsedCall
		kinds  []FnKind
		want   bool
	}{
		{"matches the only kind", test, []FnKind{FnKindTest}, true},
		{"matches one of several", test, []FnKind{FnKindDescribe, FnKindTest}, true},
		{"rejects other kinds", test, []FnKind{FnKindDescribe, FnKindHook}, false},
		// Both guards matter: callers pass the result of a parse that may have
		// failed, and an empty kinds list must not read as "any kind".
		{"nil parse is false", nil, []FnKind{FnKindTest}, false},
		{"no kinds is false", test, nil, false},
		{"nil parse and no kinds is false", nil, nil, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := IsCallOfKind(test.parsed, test.kinds...); got != test.want {
				t.Errorf("IsCallOfKind = %v, want %v", got, test.want)
			}
		})
	}
}
