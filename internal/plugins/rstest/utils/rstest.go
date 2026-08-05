package utils

import (
	"slices"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

const RstestImportModule = "@rstest/core"
const RstestPlaywrightImportModule = "@rstest/playwright"

type RstestFnType = testFramework.FnKind

type RstestImportMode = testFramework.ReferenceMode

type RstestParameterizedKind string

const (
	RstestParameterizedNone RstestParameterizedKind = ""
	RstestParameterizedEach RstestParameterizedKind = "each"
	RstestParameterizedFor  RstestParameterizedKind = "for"
)

type ParsedRstestFnCall struct {
	testFramework.ParsedCall
	ParameterizedKind RstestParameterizedKind
	// Skipped and Todo report whether the call carries `.skip` / `.todo`.
	// Like ParameterizedKind these are semantic conclusions that survive alias
	// resolution, so prefer them over scanning Members, which is call-site only.
	Skipped bool
	Todo    bool
}

// IsParameterized reports whether the call was registered through `.each` or
// `.for`. This survives alias resolution, unlike the call site's Members.
func (parsed *ParsedRstestFnCall) IsParameterized() bool {
	return parsed.ParameterizedKind != RstestParameterizedNone
}

type ParsedRstestFnCallHead = testFramework.ParsedCallHead

type ParsedRstestFnCallHeadEntry = testFramework.ParsedCallHeadEntry

type ParsedRstestFnMemberEntry = testFramework.MemberEntry

const (
	RstestFnTypeDescribe = testFramework.FnKindDescribe
	RstestFnTypeHook     = testFramework.FnKindHook
	RstestFnTypeTest     = testFramework.FnKindTest
)

// RSTEST_HOOK_NAMES lists the setup/teardown hooks. onTestFinished and
// onTestFailed are deliberately absent: they are execution-time APIs called
// inside a test body (RunnerAPI exposes them next to the hooks, but their
// handlers run per test), so they must not parse as hooks.
// Source: rstest c4b67c72 packages/core/src/types/api.ts:244-255 (RunnerAPI)
// and packages/core/globals.d.ts; @rstest/playwright re-exports the same four
// (packages/playwright/src/index.ts:2-9).
var RSTEST_HOOK_NAMES = map[string]bool{
	"afterAll":   true,
	"afterEach":  true,
	"beforeAll":  true,
	"beforeEach": true,
}

// RSTEST_HOOKS_ORDER is the expected declaration order used by
// prefer-hooks-in-order, mirroring jest's JEST_HOOKS_ORDER.
var RSTEST_HOOKS_ORDER = []string{"beforeAll", "beforeEach", "afterEach", "afterAll"}

// RstestHookOrderIndex returns the expected declaration order index for a
// Rstest hook name, or -1 if unknown. Mirrors jest's JestHookOrderIndex.
func RstestHookOrderIndex(name string) int {
	return slices.Index(RSTEST_HOOKS_ORDER, name)
}

const (
	RSTEST_GLOBAL_MODE = testFramework.ReferenceModeGlobal
	RSTEST_IMPORT_MODE = testFramework.ReferenceModeImport
)

func GetRstestFnMemberEntries(node *ast.Node) []ParsedRstestFnMemberEntry {
	return testFramework.GetMemberEntries(node)
}

func JoinRstestFnMemberEntries(entries []ParsedRstestFnMemberEntry) string {
	return testFramework.JoinMemberEntries(entries)
}

func RstestFnMemberEntriesRange(entries []ParsedRstestFnMemberEntry) (core.TextRange, bool) {
	return testFramework.MemberEntriesRange(entries)
}

func ResolveFirstIdentifier(node *ast.Node) *ast.Node {
	return testFramework.ResolveFirstIdentifier(node)
}

func ResolveFunctionReferenceForModule(
	node *ast.Node,
	localName string,
	localNode *ast.Node,
	ctx rule.RuleContext,
	importModule string,
) (string, *ast.Node, RstestImportMode) {
	return testFramework.ResolveFunctionReferenceForModule(
		node,
		localName,
		localNode,
		ctx.TypeChecker,
		ctx.SourceFile,
		importModule,
	)
}

func ResolveRstestFunctionReference(
	node *ast.Node,
	localName string,
	localNode *ast.Node,
	ctx rule.RuleContext,
) (string, *ast.Node, RstestImportMode) {
	return ResolveFunctionReferenceForModule(node, localName, localNode, ctx, RstestImportModule)
}
