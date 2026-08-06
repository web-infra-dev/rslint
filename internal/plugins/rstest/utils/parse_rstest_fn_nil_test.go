package utils

import "testing"

// ast.SkipParentheses dereferences its argument, so every helper that may be
// handed a nil node — an initializer, an optional expression — has to nil-check
// before calling it. These are white-box tests because the call sites that
// would reach them are currently all guarded; the point is to keep the helpers
// safe if a future caller is not.
func TestParseRstestChainAcceptsNil(t *testing.T) {
	root, parts, rootInvoked, ok := parseRstestChain(nil)
	if root != nil || parts != nil || rootInvoked || ok {
		t.Fatalf("expected zero values, got (%v, %v, %v, %v)", root, parts, rootInvoked, ok)
	}
}

func TestParseImportMetaRstestChainAcceptsNil(t *testing.T) {
	root, parts, rootInvoked, ok := parseImportMetaRstestChain(nil)
	if root != nil || parts != nil || rootInvoked || ok {
		t.Fatalf("expected zero values, got (%v, %v, %v, %v)", root, parts, rootInvoked, ok)
	}
}

func TestIsImportMetaAcceptsNil(t *testing.T) {
	if isImportMeta(nil) {
		t.Fatal("expected false")
	}
}

func TestIsImportMetaRstestAcceptsNil(t *testing.T) {
	if isImportMetaRstest(nil) {
		t.Fatal("expected false")
	}
}
