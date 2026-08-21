package valid_expect_in_promise

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/valid_expect_in_promise"
)

// sourceMayContainPromiseChain reports whether the file contains a call accepted
// by testFramework.IsPromiseChainCall, which recognizes `then`, `catch` and
// `finally` reached either as a property access or as a bracket access keyed by
// a string or no-substitution template literal.
//
// The two forms are gated differently. A property access spells the member as
// an identifier — a keyword-named property is interned like any other — so the
// source file's identifier table answers that form exactly and in O(1). It also
// excludes the `catch` of a `try`/`catch`, which stays a keyword token and never
// reaches the table; gating on the raw source text instead let every file with a
// `try`/`catch`, and every file with a backslash, through to the AST walk this
// function exists to avoid.
//
// A bracket access keys the member off a literal that never reaches the
// identifier table, so it keeps a text scan. That scan stays narrow by first
// requiring a quote or backtick directly after `[`. The backslash check remains
// for escaped keys such as promise["\x74hen"](), whose source text does not
// spell the member name.
func sourceMayContainPromiseChain(sourceFile *ast.SourceFile) bool {
	if sourceFile == nil {
		return true
	}
	// Reading a nil identifier table is well defined; a file with no identifiers
	// has no property-access chain either way.
	for _, name := range promiseChainMemberNames {
		if _, ok := sourceFile.Identifiers[name]; ok {
			return nodeContainsPromiseChain(sourceFile.AsNode())
		}
	}

	text := sourceFile.Text()
	if !strings.Contains(text, `["`) &&
		!strings.Contains(text, `['`) &&
		!strings.Contains(text, "[`") {
		return false
	}
	if !strings.Contains(text, "then") &&
		!strings.Contains(text, "catch") &&
		!strings.Contains(text, "finally") &&
		!strings.Contains(text, `\`) {
		return false
	}
	return nodeContainsPromiseChain(sourceFile.AsNode())
}

var promiseChainMemberNames = [...]string{"then", "catch", "finally"}

func nodeContainsPromiseChain(node *ast.Node) bool {
	if node == nil {
		return false
	}
	if testFramework.IsPromiseChainCall(node) {
		return true
	}
	// ForEachChild reports whether any visit returned true, so passing the
	// recursion itself keeps the walk free of a capturing closure.
	return node.ForEachChild(nodeContainsPromiseChain)
}

var ValidExpectInPromiseRule = shared.NewRule(shared.Config{
	Name:               "rstest/valid-expect-in-promise",
	MessageDescription: "This promise should either be returned or awaited to ensure the assertions in its chain are called",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		if !sourceMayContainPromiseChain(ctx.SourceFile) {
			return shared.Runtime{}
		}
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			TestCallbackFunctions: analysis.Callbacks().Functions,
			IsAssertionCall: func(node *ast.Node) bool {
				return analysis.IsExpectCall(node) || isRstestAssertCall(node, ctx)
			},
			IsAsyncAssertionSink: func(callNode, value *ast.Node) bool {
				if callNode == nil || callNode.Kind != ast.KindCallExpression {
					return false
				}
				top := shared.TopMostCallExpressionOnCallee(callNode)
				parsed := analysis.ParseExpectCall(top)
				if parsed == nil ||
					(!slices.Contains(parsed.Modifiers, "resolves") &&
						!slices.Contains(parsed.Modifiers, "rejects")) ||
					parsed.Head == nil || parsed.Head.AsCallExpression().Arguments == nil ||
					len(parsed.Head.AsCallExpression().Arguments.Nodes) == 0 {
					return false
				}
				return ast.SkipParentheses(parsed.Head.AsCallExpression().Arguments.Nodes[0]) ==
					ast.SkipParentheses(value)
			},
		}
	},
})

func isRstestAssertCall(node *ast.Node, ctx rule.RuleContext) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	entries := testFramework.GetMemberEntries(node)
	if len(entries) == 0 {
		return false
	}
	if isDirectImportMetaRstestAssert(node, entries) {
		return true
	}
	root := entries[0].Node
	if root == nil || root.Kind != ast.KindIdentifier {
		return false
	}
	localName := root.Text()
	var symbol *ast.Symbol
	if ctx.TypeChecker != nil {
		symbol = ctx.TypeChecker.GetSymbolAtLocation(root)
	}
	if name, ok := importMetaRstestBindingName(symbol); ok {
		return name == "assert"
	}
	if isImportMetaRstestAlias(symbol) {
		return len(entries) > 1 && entries[1].Name == "assert"
	}
	name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
		localName,
		root,
		symbol,
		ctx.SourceFile,
		rstestUtils.RstestImportModule,
	)
	if name == "assert" {
		return true
	}
	return testFramework.IsModuleNamespaceSymbol(symbol, rstestUtils.RstestImportModule) &&
		len(entries) > 1 && entries[1].Name == "assert"
}

func isDirectImportMetaRstestAssert(
	node *ast.Node,
	entries []testFramework.MemberEntry,
) bool {
	call := node.AsCallExpression()
	if call == nil || len(entries) < 2 {
		return false
	}
	root := testFramework.ResolveFirstIdentifier(call.Expression)
	if root != nil {
		return false
	}
	return entries[0].Name == "rstest" && entries[1].Name == "assert" &&
		containsImportMeta(call.Expression)
}

func containsImportMeta(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindMetaProperty {
		meta := node.AsMetaProperty()
		return meta != nil && meta.Name() != nil && meta.Name().Text() == "meta"
	}
	switch node.Kind {
	case ast.KindCallExpression:
		return containsImportMeta(node.AsCallExpression().Expression)
	case ast.KindPropertyAccessExpression:
		return containsImportMeta(node.AsPropertyAccessExpression().Expression)
	case ast.KindElementAccessExpression:
		return containsImportMeta(node.AsElementAccessExpression().Expression)
	}
	return false
}

func importMetaRstestBindingName(symbol *ast.Symbol) (string, bool) {
	if symbol == nil {
		return "", false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindBindingElement {
			continue
		}
		variable := internalUtils.EnclosingVariableDeclarationOfBindingElement(declaration)
		if variable == nil || !isImportMetaRstest(variable.AsVariableDeclaration().Initializer) {
			continue
		}
		binding := declaration.AsBindingElement()
		if binding.PropertyName != nil {
			if name, ok := internalUtils.GetStaticStringLiteralValue(binding.PropertyName); ok {
				return name, true
			}
			if binding.PropertyName.Kind == ast.KindIdentifier {
				return binding.PropertyName.Text(), true
			}
		}
		if binding.Name() != nil && binding.Name().Kind == ast.KindIdentifier {
			return binding.Name().Text(), true
		}
	}
	return "", false
}

func isImportMetaRstestAlias(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
			continue
		}
		if isImportMetaRstest(declaration.AsVariableDeclaration().Initializer) {
			return true
		}
	}
	return false
}

func isImportMetaRstest(node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}
	var expression *ast.Node
	var name string
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		expression = property.Expression
		if property.Name() != nil {
			name = property.Name().Text()
		}
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		expression = element.Expression
		name, _ = internalUtils.GetStaticStringLiteralValue(
			ast.SkipParentheses(element.ArgumentExpression),
		)
	default:
		return false
	}
	return name == "rstest" && containsImportMeta(expression)
}
