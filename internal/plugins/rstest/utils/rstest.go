package utils

import (
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
