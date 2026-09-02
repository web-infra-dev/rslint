package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

const RstestImportModule = "@rstest/core"
const RstestPlaywrightImportModule = "@rstest/playwright"

// RstackTestImportModule is the Rstack CLI's re-export of the Rstest core API.
// `rstack/test` is a fixed subpath export whose module body is
// `export * from '@rstest/core'`, so a binding imported from it is the same API
// under a second specifier, not a separate framework.
const RstackTestImportModule = "rstack/test"

// The module lists below are package-level so that resolving a call site does
// not allocate a slice per lookup.
var (
	// RstestCoreImportModules holds every specifier that reaches the Rstest
	// core API surface.
	RstestCoreImportModules = []string{RstestImportModule, RstackTestImportModule}
	// RstestPlaywrightImportModules holds every specifier that reaches the
	// Rstest Playwright API surface.
	RstestPlaywrightImportModules = []string{RstestPlaywrightImportModule}
	// RstestAllImportModules holds every specifier of either surface.
	RstestAllImportModules = []string{
		RstestImportModule,
		RstackTestImportModule,
		RstestPlaywrightImportModule,
	}
)

// IsRstestCoreImportModule reports whether specifier reaches the Rstest core
// API surface.
func IsRstestCoreImportModule(specifier string) bool {
	return specifier == RstestImportModule || specifier == RstackTestImportModule
}

// IsImportMetaRstest reports whether node is the Rstest module namespace
// exposed through import.meta. Parentheses around either link are transparent.
func IsImportMetaRstest(node *ast.Node) bool {
	return isImportMetaRstest(node)
}

type RstestFnType = testFramework.FnKind

type RstestImportMode = testFramework.ReferenceMode

type RstestParameterizedKind string

const (
	RstestParameterizedNone RstestParameterizedKind = ""
	RstestParameterizedEach RstestParameterizedKind = "each"
	RstestParameterizedFor  RstestParameterizedKind = "for"
)

// RstestExecutionMode records the mode explicitly applied to a registration.
// Default registrations inherit from an enclosing describe callback; explicit
// concurrent or sequential registrations override that inherited mode.
type RstestExecutionMode uint8

const (
	RstestExecutionDefault RstestExecutionMode = iota
	RstestExecutionConcurrent
	RstestExecutionSequential
)

type ParsedRstestFnCall struct {
	testFramework.ParsedCall
	ParameterizedKind RstestParameterizedKind
	ExecutionMode     RstestExecutionMode
	// Skipped and Todo report whether the call carries `.skip` / `.todo`.
	// Like ParameterizedKind these are semantic conclusions that survive alias
	// resolution, so prefer them over scanning Members, which is call-site only.
	Skipped bool
	Todo    bool
	// focus is allocated only when the resolved registration carries `.only`.
	// Keeping rare provenance behind one pointer avoids increasing every parsed
	// registration allocation in files without focused tests.
	focus *rstestFocus
}

type rstestFocus struct {
	// Entries holds every `.only` source accessor that contributed to the
	// focused registration. Unlike MemberEntries, alias-internal entries live
	// here because they have a concrete source node useful for diagnostics and
	// suggestions.
	entries []ParsedRstestFnMemberEntry
}

func (parsed *ParsedRstestFnCall) IsFocused() bool {
	return parsed != nil && parsed.focus != nil
}

func (parsed *ParsedRstestFnCall) FocusEntries() []ParsedRstestFnMemberEntry {
	if parsed == nil || parsed.focus == nil {
		return nil
	}
	return parsed.focus.entries
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

func RstestFnMemberEntriesRange(sourceFile *ast.SourceFile, entries []ParsedRstestFnMemberEntry) (core.TextRange, bool) {
	return testFramework.MemberEntriesRange(sourceFile, entries)
}

func ResolveFirstIdentifier(node *ast.Node) *ast.Node {
	return testFramework.ResolveFirstIdentifier(node)
}

func ResolveFunctionReferenceForModules(
	node *ast.Node,
	localName string,
	localNode *ast.Node,
	ctx rule.RuleContext,
	importModules []string,
) (string, *ast.Node, RstestImportMode) {
	return testFramework.ResolveFunctionReferenceForModules(
		node,
		localName,
		localNode,
		ctx.TypeChecker,
		ctx.SourceFile,
		importModules,
	)
}

func ResolveRstestFunctionReference(
	node *ast.Node,
	localName string,
	localNode *ast.Node,
	ctx rule.RuleContext,
) (string, *ast.Node, RstestImportMode) {
	return ResolveFunctionReferenceForModules(node, localName, localNode, ctx, RstestCoreImportModules)
}
