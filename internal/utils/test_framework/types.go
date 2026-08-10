package test_framework

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
)

// FnKind classifies a parsed test-framework call without encoding one
// framework's root names or aliases. Framework parsers may define additional
// values for assertion and namespace APIs.
type FnKind string

const (
	FnKindDescribe FnKind = "describe"
	FnKindHook     FnKind = "hook"
	FnKindTest     FnKind = "test"
	FnKindUnknown  FnKind = "unknown"
)

// ReferenceMode records whether a test API was resolved as a global or as an
// import/require binding from a framework module.
type ReferenceMode string

const (
	ReferenceModeGlobal ReferenceMode = "global"
	ReferenceModeImport ReferenceMode = "import"
)

// MemberEntry describes one segment in a test API member chain.
type MemberEntry struct {
	Name string
	Node *ast.Node
	Call *ast.Node
}

// ParsedCall is the framework-neutral portion of a parsed test API call.
// Framework parsers own root/member validation and may embed this type to add
// matcher or framework-specific metadata.
type ParsedCall struct {
	Name          string
	LocalName     string
	Kind          FnKind
	Members       []string
	MemberEntries []MemberEntry
	Head          ParsedCallHead
}

// HooksOrder is the order setup/teardown hooks are expected to be declared in,
// used by prefer-hooks-in-order. Jest and Rstest expose the same four hooks in
// the same order; frameworks that differ should not reuse this.
//
// Rstest source: rstest c4b67c72 packages/core/src/types/api.ts:244-255
// (RunnerAPI) and packages/core/globals.d.ts; @rstest/playwright re-exports the
// same four (packages/playwright/src/index.ts:2-9). Rstest's onTestFinished and
// onTestFailed are deliberately absent: they are execution-time APIs called
// inside a test body, so they must not parse as hooks.
var HooksOrder = []string{"beforeAll", "beforeEach", "afterEach", "afterAll"}

// IsHookName reports whether name is one of HooksOrder.
func IsHookName(name string) bool {
	return slices.Contains(HooksOrder, name)
}

// HookOrderIndex returns the expected declaration order index for a hook name,
// or -1 if it is not a hook.
func HookOrderIndex(name string) int {
	return slices.Index(HooksOrder, name)
}

// IsCallOfKind reports whether parsed resolved to one of kinds. It is the
// shared body of jest's IsTypeOfJestFnCall and rstest's IsTypeOfRstestFnCall:
// each plugin owns which parse entry point to call, this owns what the answer
// means. An empty kinds list is false rather than "any kind", so a caller that
// forgets its arguments does not silently match everything.
func IsCallOfKind(parsed *ParsedCall, kinds ...FnKind) bool {
	if parsed == nil || len(kinds) == 0 {
		return false
	}
	return slices.Contains(kinds, parsed.Kind)
}

type ParsedCallHead struct {
	Type     ReferenceMode
	Local    ParsedCallHeadEntry
	Original ParsedCallHeadEntry
}

type ParsedCallHeadEntry struct {
	Value string
	Node  *ast.Node
}
