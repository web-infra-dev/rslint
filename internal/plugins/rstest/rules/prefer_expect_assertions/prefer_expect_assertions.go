package prefer_expect_assertions

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstest "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/prefer_expect_assertions"
)

//go:embed prefer_expect_assertions.schema.json
var schemaJSON []byte

// PreferExpectAssertionsRule counts assertions set up in beforeEach only.
// Rstest resets the assertion state before beforeEach hooks run and checks it
// when the test callback settles, before any afterEach hook runs, so an
// afterEach or beforeAll hook cannot set a requirement for the test.
var PreferExpectAssertionsRule = shared.NewRule(shared.Config{
	Name:                  "rstest/prefer-expect-assertions",
	Schema:                schemaJSON,
	CoveringHooks:         []string{"beforeEach"},
	DisallowHasAssertions: true,
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstest.GetRstestCallAnalysis(ctx)
		var spellings *moduleSpellings
		return shared.Runtime{
			Classify: func(node *ast.Node) shared.Registration {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return shared.Registration{}
				}
				switch parsed.Kind {
				case rstest.RstestFnTypeDescribe:
					return shared.Registration{Kind: shared.RegistrationDescribe}
				case rstest.RstestFnTypeHook:
					registration := shared.Registration{Kind: shared.RegistrationHook, HookName: parsed.Name}
					if arguments := node.AsCallExpression().Arguments.Nodes; len(arguments) > 0 {
						registration.Callback = arguments[0]
					}
					return registration
				case rstest.RstestFnTypeTest:
					registration := shared.Registration{Kind: shared.RegistrationTest}
					// A callback passed by name is resolved by TestCallback too, but
					// only a function literal written at this registration is checked.
					if callback, name := analysis.TestCallback(node); callback != nil && name == "" {
						registration.Callback = callback
					}
					return registration
				}
				return shared.Registration{}
			},
			ParseStatic: func(node *ast.Node) *shared.StaticCall {
				member, memberName, ok := staticCallee(node)
				if !ok {
					return nil
				}
				if parsed := analysis.ParseExpectCall(node); parsed != nil {
					if parsed.Entry != rstest.RstestExpectEntryStatic || len(parsed.MemberEntries) != 1 ||
						parsed.MemberEntries[0].Node != member {
						return nil
					}
					return &shared.StaticCall{Call: node, Member: memberName, MemberNode: member}
				}
				if isHookContextExpect(ctx, analysis, node) {
					return &shared.StaticCall{Call: node, Member: memberName, MemberNode: member}
				}
				return nil
			},
			IsExpect: analysis.IsExpectCall,
			ExpectSpelling: func(test *ast.Node, callback *ast.Node) (string, bool) {
				if spellings == nil {
					spellings = collectModuleSpellings(ctx)
				}
				return expectSpelling(analysis, spellings, test, callback)
			},
		}
	},
})

// staticCallee returns the member node of a call written as
// `<receiver>.assertions(...)` or `<receiver>.hasAssertions(...)`, including
// bracket access with a string key.
func staticCallee(node *ast.Node) (*ast.Node, string, bool) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	var member *ast.Node
	switch {
	case ast.IsPropertyAccessExpression(callee):
		member = callee.Name()
	case ast.IsElementAccessExpression(callee):
		member = callee.AsElementAccessExpression().ArgumentExpression
	default:
		return nil, "", false
	}
	name, ok := shared.ParseStaticMember(member)
	return member, name, ok
}

// isHookContextExpect recognizes `expect` taken from the TestContext a hook
// receives, as in `beforeEach(({ expect }) => expect.hasAssertions())` or
// `beforeEach((context) => context.expect.hasAssertions())`. Rstest passes the
// running test's context to beforeEach, and assertion requirements set on its
// expect apply to that test. The call analysis tracks TestContext only for
// test callbacks, so the hook form is recognized here. The engine only counts
// the call when its function is a hook callback, which keeps the first
// parameter of any other function from being mistaken for a TestContext.
func isHookContextExpect(ctx rule.RuleContext, analysis *rstest.RstestCallAnalysis, node *ast.Node) bool {
	if ctx.Refs == nil {
		return false
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	receiver := ast.SkipParentheses(callee.Expression())
	function := enclosingFunction(node)
	if function == nil || analysis.Callbacks().Functions[function] {
		return false
	}
	parameters := function.Parameters()
	if len(parameters) == 0 {
		return false
	}
	context := parameters[0].Name()

	switch {
	case receiver.Kind == ast.KindIdentifier:
		// `({ expect }) => expect.hasAssertions()`
		declaration := resolveUnwrittenDeclaration(ctx, receiver)
		if declaration == nil || declaration.Kind != ast.KindBindingElement ||
			declaration.Parent != context || context.Kind != ast.KindObjectBindingPattern {
			return false
		}
		return rstest.RequireBindingImportedName(declaration) == "expect"
	case ast.IsPropertyAccessExpression(receiver):
		// `(context) => context.expect.hasAssertions()`
		object := ast.SkipParentheses(receiver.Expression())
		if receiver.Name().Text() != "expect" || object.Kind != ast.KindIdentifier || context.Kind != ast.KindIdentifier {
			return false
		}
		declaration := resolveUnwrittenDeclaration(ctx, object)
		return declaration != nil && declaration == parameters[0]
	}
	return false
}

func resolveUnwrittenDeclaration(ctx rule.RuleContext, identifier *ast.Node) *ast.Node {
	symbol := ctx.Refs.Resolve(identifier)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	for _, reference := range ctx.Refs.References(symbol) {
		if internalUtils.IsWriteReference(reference) {
			return nil
		}
	}
	return symbol.Declarations[0]
}

func enclosingFunction(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(current) {
			return current
		}
	}
	return nil
}

// moduleSpellings records how the file reaches Rstest's expect at module
// level.
type moduleSpellings struct {
	// expect is the spelling of an import or require of expect, such as
	// `expect`, `check` or `rstest.expect`; empty when there is none.
	expect string
	// importsRstest reports whether the file imports or requires any Rstest
	// module. A file that does not is assumed to run with globals enabled.
	importsRstest bool
	// declaresExpect reports whether the file binds the name expect at module
	// level to something other than Rstest, such as Chai's expect.
	declaresExpect bool
}

func collectModuleSpellings(ctx rule.RuleContext) *moduleSpellings {
	result := &moduleSpellings{}
	for _, statement := range ctx.SourceFile.Statements.Nodes {
		switch statement.Kind {
		case ast.KindImportDeclaration:
			collectImportSpelling(result, statement.AsImportDeclaration())
		case ast.KindVariableStatement:
			for _, declaration := range statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
				collectRequireSpelling(ctx, result, declaration)
			}
		case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			if name := statement.Name(); name != nil && name.Text() == "expect" {
				result.declaresExpect = true
			}
		}
	}
	return result
}

func collectImportSpelling(result *moduleSpellings, declaration *ast.ImportDeclaration) {
	clause := declaration.ImportClause
	if clause == nil || clause.IsTypeOnly() {
		return
	}
	fromRstest := slices.Contains(rstest.RstestAllImportModules, internalUtils.GetStaticStringValue(declaration.ModuleSpecifier))
	if fromRstest {
		result.importsRstest = true
	}
	importClause := clause.AsImportClause()
	if name := importClause.Name(); name != nil && name.Text() == "expect" {
		result.declaresExpect = true
	}
	bindings := importClause.NamedBindings
	if bindings == nil {
		return
	}
	if bindings.Kind == ast.KindNamespaceImport {
		local := bindings.Name().Text()
		if fromRstest && result.expect == "" {
			result.expect = local + ".expect"
		} else if local == "expect" {
			result.declaresExpect = true
		}
		return
	}
	for _, element := range rstest.NamedImportElements(declaration) {
		local := element.Name().Text()
		if fromRstest && rstest.ImportedSpecifierName(element) == "expect" {
			if result.expect == "" {
				result.expect = local
			}
			continue
		}
		if local == "expect" {
			result.declaresExpect = true
		}
	}
}

// collectRequireSpelling records the spelling of a require of Rstest. Unlike
// an import binding, a `let` or `var` binding can be reassigned after the
// require, so a binding with any write is not used: after
// `check = chai.expect`, an inserted `check.hasAssertions()` would call Chai.
func collectRequireSpelling(ctx rule.RuleContext, result *moduleSpellings, declaration *ast.Node) {
	name := declaration.Name()
	if _, ok := rstest.RstestCoreModuleFromRequireCall(declaration.AsVariableDeclaration().Initializer); ok {
		result.importsRstest = true
		switch name.Kind {
		case ast.KindIdentifier:
			if result.expect == "" && !isRebound(ctx, declaration) {
				result.expect = name.Text() + ".expect"
			}
		case ast.KindObjectBindingPattern:
			for _, element := range name.AsBindingPattern().Elements.Nodes {
				local := element.Name()
				if local == nil || local.Kind != ast.KindIdentifier {
					continue
				}
				if rstest.RequireBindingImportedName(element) == "expect" && result.expect == "" && !isRebound(ctx, element) {
					result.expect = local.Text()
				}
			}
		}
		return
	}
	if name.Kind == ast.KindIdentifier && name.Text() == "expect" {
		result.declaresExpect = true
	}
}

// isRebound reports whether the binding a declaration introduces is ever
// written. A binding whose symbol cannot be found counts as written, which
// only withholds a suggestion.
func isRebound(ctx rule.RuleContext, declaration *ast.Node) bool {
	symbol := declaration.Symbol()
	if symbol == nil || ctx.Refs == nil {
		return true
	}
	for _, reference := range ctx.Refs.References(symbol) {
		if internalUtils.IsWriteReference(reference) {
			return true
		}
	}
	return false
}

// expectSpelling picks how an inserted `expect.hasAssertions()` refers to
// expect. A TestContext expect the callback destructures is used first, since
// it is what the callback's own assertions use. A concurrent test prefers its
// TestContext because Rstest tracks the assertions of concurrent tests on the
// local expect. Otherwise the file's own import is used, then the TestContext
// parameter, then the global for files that import nothing from Rstest. With
// none of these, or when the name is shadowed on the way to the callback, no
// suggestion is offered: inserting a bare `expect` into a file that runs
// without globals would throw a ReferenceError.
func expectSpelling(
	analysis *rstest.RstestCallAnalysis,
	spellings *moduleSpellings,
	test *ast.Node,
	callback *ast.Node,
) (string, bool) {
	parsed := analysis.ParseTestCall(test)
	if parsed == nil {
		return "", false
	}
	contextName, destructuredExpect := testContextBinding(parsed, callback)
	if destructuredExpect != "" {
		return destructuredExpect, !declaresInBody(callback, destructuredExpect)
	}
	contextSpelling := ""
	if contextName != "" && !declaresInBody(callback, contextName) {
		contextSpelling = contextName + ".expect"
	}
	if contextSpelling != "" && parsed.ExecutionMode == rstest.RstestExecutionConcurrent {
		return contextSpelling, true
	}
	if spellings.expect != "" && !isShadowed(callback, rootName(spellings.expect)) {
		return spellings.expect, true
	}
	if contextSpelling != "" {
		return contextSpelling, true
	}
	if !spellings.importsRstest && !spellings.declaresExpect && !isShadowed(callback, "expect") {
		return "expect", true
	}
	return "", false
}

// testContextBinding returns the TestContext parameter's name when it is an
// identifier, or the local name of its destructured expect. `.each` spreads
// row values into the parameters and passes no TestContext; `.for` passes the
// row first and the TestContext second.
func testContextBinding(parsed *rstest.ParsedRstestFnCall, callback *ast.Node) (contextName string, destructuredExpect string) {
	index := 0
	switch parsed.ParameterizedKind {
	case rstest.RstestParameterizedEach:
		return "", ""
	case rstest.RstestParameterizedFor:
		index = 1
	}
	parameters := callback.Parameters()
	if index >= len(parameters) {
		return "", ""
	}
	name := parameters[index].Name()
	switch name.Kind {
	case ast.KindIdentifier:
		return name.Text(), ""
	case ast.KindObjectBindingPattern:
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			local := element.Name()
			if local != nil && local.Kind == ast.KindIdentifier && rstest.RequireBindingImportedName(element) == "expect" {
				return "", local.Text()
			}
		}
	}
	return "", ""
}

func rootName(spelling string) string {
	for i := range len(spelling) {
		if spelling[i] == '.' {
			return spelling[:i]
		}
	}
	return spelling
}

// isShadowed conservatively reports whether a function between the callback
// and the module declares name, either as a parameter or anywhere in its own
// body. Declarations in sibling blocks are counted too, which only withholds
// a suggestion.
func isShadowed(callback *ast.Node, name string) bool {
	for current := callback; current != nil && current.Kind != ast.KindSourceFile; current = current.Parent {
		if !ast.IsFunctionLikeOrClassStaticBlockDeclaration(current) {
			continue
		}
		for _, parameter := range current.Parameters() {
			if bindsName(parameter.Name(), name) {
				return true
			}
		}
		if declaresInBody(current, name) {
			return true
		}
	}
	return false
}

// declaresInBody reports whether function declares name in its body outside
// nested functions. A nested function declaration's own name still counts.
func declaresInBody(function *ast.Node, name string) bool {
	body := function.Body()
	if body == nil {
		return false
	}
	found := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if found {
			return true
		}
		switch node.Kind {
		case ast.KindVariableDeclaration, ast.KindClassDeclaration, ast.KindFunctionDeclaration,
			ast.KindClassExpression, ast.KindEnumDeclaration, ast.KindModuleDeclaration:
			if node.Kind != ast.KindClassExpression && bindsName(node.Name(), name) {
				found = true
				return true
			}
		case ast.KindCatchClause:
			if declaration := node.AsCatchClause().VariableDeclaration; declaration != nil && bindsName(declaration.Name(), name) {
				found = true
				return true
			}
		}
		if ast.IsFunctionLikeOrClassStaticBlockDeclaration(node) || ast.IsClassLike(node) {
			return false
		}
		return node.ForEachChild(visit)
	}
	body.ForEachChild(visit)
	return found
}

func bindsName(binding *ast.Node, name string) bool {
	if binding == nil {
		return false
	}
	switch binding.Kind {
	case ast.KindIdentifier:
		return binding.Text() == name
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		for _, element := range binding.AsBindingPattern().Elements.Nodes {
			if element.Kind == ast.KindBindingElement && bindsName(element.Name(), name) {
				return true
			}
		}
	}
	return false
}
