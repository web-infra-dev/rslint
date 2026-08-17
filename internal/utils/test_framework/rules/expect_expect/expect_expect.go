// Package expect_expect holds the framework-neutral traversal shared by
// jest/expect-expect and rstest/expect-expect.
//
// The rule keeps a ledger of test registrations that have not yet been seen to
// contain an assertion (`unchecked`). Entering a test registration pushes it;
// encountering a call whose callee chain matches one of the configured
// assertion patterns walks the ancestor chain and removes the nearest enclosing
// test from the ledger. Assertions reached through a named function callback
// (`test('x', run); function run() { expect(...) }`) are matched by resolving
// the callback to its declaration or name. At end of file every test still in
// the ledger is reported.
//
// The account-keeping is deliberately identical to eslint-plugin-jest's
// expect-expect; framework differences live entirely in the injected Config:
// which calls are test registrations (and whether they are `.todo`), how a
// named-function callback resolves, and the default assertFunctionNames.
package expect_expect

import (
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed expect_expect.schema.json
var schemaJSON []byte

// TestClassification is what Config.ClassifyTest reports about a call.
type TestClassification struct {
	// IsTest is true when the call is a final test registration whose callback
	// must contain an assertion.
	IsTest bool
	// IsTodo is true when the registration is a `.todo` form; such tests have no
	// callback and are exempt from the assertion requirement.
	IsTodo bool
}

// NamedCallback identifies a test callback passed by reference, e.g.
// `test('x', run)`. Frameworks resolve it however their parser allows; the
// shared body only needs the declaration node (when statically known) and the
// name (for the second-pass fallback used when the declaration is not resolved
// through the type checker).
type NamedCallback struct {
	DeclarationNode *ast.Node
	Name            string
}

type Runtime struct {
	ClassifyTest         func(node *ast.Node) TestClassification
	ResolveNamedCallback func(callNode *ast.Node) NamedCallback
	// IsAssertion recognizes assertions the configured assertFunctionNames
	// patterns cannot express, because they match callee text only: a framework
	// whose `expect` reaches the call site through the test context, a namespace
	// import or an import alias produces a callee chain (`ctx.expect`,
	// `rstest.expect`, `check`) that no pattern for `expect` matches. It is
	// consulted only after the patterns miss, so the common bare `expect(...)`
	// never pays for it. Frameworks whose `expect` is always a bare global —
	// jest — leave it nil. Assertions it recognizes always count, including when
	// assertFunctionNames is configured without `expect`: the option names extra
	// asserting calls, it does not hide the framework's own.
	IsAssertion func(node *ast.Node) bool
}

type Config struct {
	Name string
	// DefaultAssertFunctionNames is used when the rule receives no
	// assertFunctionNames option. jest uses ["expect"]; rstest uses
	// ["expect", "assert"].
	DefaultAssertFunctionNames []string
	// Prepare creates the framework adapter once per file.
	Prepare func(ctx rule.RuleContext) Runtime
}

func buildErrorNoAssertionsMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noAssertions",
		Description: "Test has no assertions",
	}
}

func parseOptions(options []any, defaultAssertNames []string) ([]string, []string) {
	assertNames := slices.Clone(defaultAssertNames)
	additional := []string{}

	if len(options) == 0 {
		return assertNames, additional
	}
	optsMap, _ := options[0].(map[string]interface{})

	if arr, ok := optsMap["assertFunctionNames"].([]interface{}); ok {
		assertNames = stringList(arr)
	}
	if arr, ok := optsMap["additionalTestBlockFunctions"].([]interface{}); ok {
		additional = stringList(arr)
	}

	return assertNames, additional
}

func stringList(raw []interface{}) []string {
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

func compileAssertPatterns(patterns []string) []*esregexp.RegExp {
	out := make([]*esregexp.RegExp, 0, len(patterns))
	for _, p := range patterns {
		out = append(out, compileAssertPattern(p))
	}
	return out
}

func compileAssertPattern(pattern string) *esregexp.RegExp {
	segs := strings.Split(pattern, ".")
	parts := make([]string, 0, len(segs))
	for _, s := range segs {
		if s == "**" {
			parts = append(parts, `[a-zA-Z0-9.]*`)
		} else {
			parts = append(parts, strings.ReplaceAll(s, "*", `[a-zA-Z0-9]*`))
		}
	}
	joined := strings.Join(parts, `\.`)
	// Segments follow eslint-plugin-jest: only `*` is expanded; other characters
	// (e.g. `\$`) are copied into the Regexp source like JavaScript's RegExp
	// constructor. A malformed pattern compiles to nil and never matches.
	re, err := esregexp.Compile(`^(?:`+joined+`)(?:\.|$)`, "iu")
	if err != nil {
		return nil
	}
	return re
}

func matchesAssertName(name string, compiled []*esregexp.RegExp) bool {
	if name == "" {
		return false
	}
	for _, re := range compiled {
		if re != nil && re.Test(name) {
			return true
		}
	}
	return false
}

func indexUnchecked(unchecked []*ast.Node, call *ast.Node) int {
	for i, c := range unchecked {
		if c == call {
			return i
		}
	}
	return -1
}

func removeUncheckedCall(unchecked *[]*ast.Node, call *ast.Node) bool {
	if idx := indexUnchecked(*unchecked, call); idx >= 0 {
		*unchecked = slices.Delete(*unchecked, idx, idx+1)
		return true
	}
	return false
}

func clearUncheckedCalls(unchecked *[]*ast.Node, calls []*ast.Node) {
	for _, call := range calls {
		if call == nil || call.Kind != ast.KindCallExpression {
			continue
		}
		removeUncheckedCall(unchecked, call)
	}
}

// markAsserted records that the callback identified by declNode/fnName has been
// seen to assert, then clears any registrations already queued against that
// declaration or name from the ledger. Both the FunctionDeclaration and
// VariableDeclaration branches of checkCallExpressionUsed share this sequence;
// they differ only in which node stands in for the declaration.
func markAsserted(
	declNode *ast.Node,
	fnName string,
	unchecked *[]*ast.Node,
	uncheckedByDecl map[*ast.Node][]*ast.Node,
	uncheckedByName map[string][]*ast.Node,
	assertedByDecl map[*ast.Node]bool,
	assertedByName map[string]bool,
) {
	assertedByDecl[declNode] = true
	assertedByName[fnName] = true
	clearUncheckedCalls(unchecked, uncheckedByDecl[declNode])
	delete(uncheckedByDecl, declNode)
	clearUncheckedCalls(unchecked, uncheckedByName[fnName])
	delete(uncheckedByName, fnName)
}

// checkCallExpressionUsed walks up from an asserting call and clears the
// registrations it covers: the nearest enclosing test call is dequeued, and any
// enclosing named callback declaration is marked asserted so registrations that
// reference it by declaration or by name are cleared too.
//
// Known limitation (deliberate, in favor of under-reporting): the walk credits
// every declaration it crosses, without checking whether the assertion can
// actually run when that declaration is used as the test callback. Both of the
// following go unreported, while eslint-plugin-jest reports them:
//
//	const outerCb = () => { test("inner", () => { expect(v).toBe(1) }) }
//	test("outer", outerCb)   // asserts only in the inner test's callback
//
//	const makeCb = () => () => { expect(v).toBe(1) }
//	test("x", makeCb)        // returns an asserting closure, asserts nothing
//
// Bounding the walk correctly would mean stopping once it crosses a function
// that is itself registered as another test's callback, and refusing to credit
// across a returned inner function. Both cost more than they buy here: the
// failure mode of getting them wrong is a false "Test has no assertions" on
// valid code, which is worse than missing these shapes.
func checkCallExpressionUsed(
	assertNode *ast.Node,
	unchecked *[]*ast.Node,
	uncheckedByDecl map[*ast.Node][]*ast.Node,
	uncheckedByName map[string][]*ast.Node,
	assertedByDecl map[*ast.Node]bool,
	assertedByName map[string]bool,
) {
	var ancestors []*ast.Node
	for n := assertNode.Parent; n != nil; n = n.Parent {
		ancestors = append(ancestors, n)
	}

	for i := len(ancestors) - 1; i >= 0; i-- {
		n := ancestors[i]
		switch n.Kind {
		case ast.KindFunctionDeclaration:
			decl := n.AsFunctionDeclaration()
			if decl != nil && decl.Name() != nil {
				markAsserted(decl.AsNode(), decl.Name().Text(),
					unchecked, uncheckedByDecl, uncheckedByName, assertedByDecl, assertedByName)
			}
		case ast.KindVariableDeclaration:
			declaration := n.AsVariableDeclaration()
			if declaration != nil && declaration.Name() != nil &&
				declaration.Name().Kind == ast.KindIdentifier &&
				declaration.Initializer != nil {
				initializer := ast.SkipParentheses(declaration.Initializer)
				if ast.IsFunctionExpressionOrArrowFunction(initializer) {
					markAsserted(initializer, declaration.Name().Text(),
						unchecked, uncheckedByDecl, uncheckedByName, assertedByDecl, assertedByName)
				}
			}
		}
		if n.Kind == ast.KindCallExpression {
			if removeUncheckedCall(unchecked, n) {
				return
			}
		}
	}
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			assertNames, additionalTestBlocks := parseOptions(options, config.DefaultAssertFunctionNames)
			compiled := compileAssertPatterns(assertNames)
			var unchecked []*ast.Node
			uncheckedByDecl := map[*ast.Node][]*ast.Node{}
			uncheckedByName := map[string][]*ast.Node{}
			assertedByDecl := map[*ast.Node]bool{}
			assertedByName := map[string]bool{}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					callExpr := node.AsCallExpression()
					if callExpr == nil {
						return
					}

					calleeName := testFramework.CalleeChainName(callExpr.Expression)

					classification := TestClassification{}
					if runtime.ClassifyTest != nil {
						classification = runtime.ClassifyTest(node)
					}
					isTest := classification.IsTest
					isExtraBlock := calleeName != "" && slices.Contains(additionalTestBlocks, calleeName)

					if isTest || isExtraBlock {
						if classification.IsTodo || strings.HasSuffix(calleeName, ".todo") {
							return
						}
						if isTest && runtime.ResolveNamedCallback != nil {
							named := runtime.ResolveNamedCallback(node)
							if named.DeclarationNode != nil && assertedByDecl[named.DeclarationNode] ||
								named.DeclarationNode == nil && named.Name != "" && assertedByName[named.Name] {
								return
							}
							switch {
							case named.DeclarationNode != nil:
								uncheckedByDecl[named.DeclarationNode] = append(uncheckedByDecl[named.DeclarationNode], node)
							case named.Name != "":
								uncheckedByName[named.Name] = append(uncheckedByName[named.Name], node)
							}
						}
						unchecked = append(unchecked, node)
						return
					}

					// Assertions are the union of the configured name patterns and
					// whatever the framework can resolve; patterns run first because
					// they answer the overwhelmingly common bare `expect(...)`
					// without touching the framework analysis.
					if !matchesAssertName(calleeName, compiled) &&
						(runtime.IsAssertion == nil || !runtime.IsAssertion(node)) {
						return
					}
					checkCallExpressionUsed(
						node,
						&unchecked,
						uncheckedByDecl,
						uncheckedByName,
						assertedByDecl,
						assertedByName,
					)
				},
				rule.ListenerOnExit(ast.KindEndOfFile): func(node *ast.Node) {
					_ = node
					for _, call := range unchecked {
						ce := call.AsCallExpression()
						if ce != nil && ce.Expression != nil {
							ctx.ReportNode(ce.Expression, buildErrorNoAssertionsMessage())
						}
					}
				},
			}
		},
	}
}
