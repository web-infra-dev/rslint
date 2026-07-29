package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

const RstestImportModule = "@rstest/core"

type RstestFnType = testFramework.FnKind

type RstestImportMode = testFramework.ReferenceMode

type ParsedRstestFnCall = testFramework.ParsedCall

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
