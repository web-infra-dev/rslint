// Package prefer_ending_with_an_expect holds the framework-neutral traversal
// shared by jest/prefer-ending-with-an-expect and
// rstest/prefer-ending-with-an-expect.
//
// The rule looks at one statement per test: the last one the callback executes.
// A block body contributes its final statement (unwrapped when it is an
// expression statement), a concise arrow body contributes its expression, and a
// leading `await` is skipped. That statement must be a call the rule recognizes
// as an assertion, either because it matches one of the configured
// assertFunctionNames patterns or because the framework resolves it as its own
// `expect`.
//
// Everything a framework decides lives in the injected Config: which calls
// register a test, whether such a registration is exempt, where its callback
// sits among the arguments, which calls are native assertions, and the default
// assertFunctionNames.
package prefer_ending_with_an_expect

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed prefer_ending_with_an_expect.schema.json
var schemaJSON []byte

// TestBlock is what Config.ClassifyTest reports about a call.
type TestBlock struct {
	// IsTest is true when the call registers a test whose callback the rule
	// should inspect.
	IsTest bool
	// IsExempt is true when the registration declares a body that never runs, so
	// requiring it to end with an assertion would be a false positive.
	IsExempt bool
}

type Runtime struct {
	// ClassifyTest recognizes the framework's own test registrations. Calls named
	// by the additionalTestBlockFunctions option are added by the shared body and
	// never reach it.
	ClassifyTest func(node *ast.Node) TestBlock
	// Callback returns the inline function whose last statement is checked, or
	// nil when the registration passes no function literal. Frameworks differ in
	// where that argument sits, so each one locates it itself.
	Callback func(call *ast.CallExpression) *ast.Node
	// IsAssertion recognizes assertions the configured assertFunctionNames
	// patterns cannot express, because those match callee text only: a framework
	// whose `expect` reaches the call site through the test context, a namespace
	// import or an import alias produces a callee chain (`ctx.expect`,
	// `rstest.expect`, `check`) that no pattern for `expect` matches. It is
	// consulted only after the patterns miss. Frameworks whose `expect` is always
	// a bare global leave it nil.
	IsAssertion func(node *ast.Node) bool
	// IsPropertyAssertion recognizes an assertion that is not a call at all.
	// Chai exposes its truthiness assertions as property getters, so
	// `expect(value).to.be.true` asserts without invoking anything and the
	// statement the test ends with is a property access.
	//
	// A framework whose assertions are always calls leaves this nil, which
	// keeps the call-only shape for every caller that came before it. It is
	// consulted only for a statement that is not a call, so it can neither
	// widen nor narrow what the assertFunctionNames patterns already match.
	IsPropertyAssertion func(node *ast.Node) bool
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

func buildMustEndWithExpectMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "mustEndWithExpect",
		Description: "Tests should end with an assertion",
	}
}

// lastFunctionStatement returns the statement a callback ends with: the final
// statement of a block body, unwrapped when it is an expression statement, or
// the expression of a concise arrow body.
func lastFunctionStatement(fn *ast.Node) *ast.Node {
	if fn == nil {
		return nil
	}
	body := fn.Body()
	if body == nil {
		return nil
	}
	if body.Kind != ast.KindBlock {
		return ast.SkipParentheses(body)
	}

	block := body.AsBlock()
	if block == nil || block.Statements == nil || len(block.Statements.Nodes) == 0 {
		return nil
	}
	last := block.Statements.Nodes[len(block.Statements.Nodes)-1]
	if last.Kind == ast.KindExpressionStatement {
		last = last.AsExpressionStatement().Expression
	}
	return ast.SkipParentheses(last)
}

func isAssertionCall(
	node *ast.Node,
	runtime Runtime,
	patterns []*esregexp.RegExp,
) bool {
	if node == nil {
		return false
	}
	if node.Kind == ast.KindAwaitExpression {
		node = ast.SkipParentheses(node.AsAwaitExpression().Expression)
	}
	if node == nil {
		return false
	}
	if node.Kind != ast.KindCallExpression {
		return runtime.IsPropertyAssertion != nil && runtime.IsPropertyAssertion(node)
	}
	if testFramework.MatchesAssertName(
		testFramework.CalleeChainName(node.AsCallExpression().Expression),
		patterns,
	) {
		return true
	}
	return runtime.IsAssertion != nil && runtime.IsAssertion(node)
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(schemaJSON),
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			runtime := config.Prepare(ctx)
			parsedOptions := testFramework.ParseAssertionFunctionOptions(options, config.DefaultAssertFunctionNames)
			patterns := testFramework.CompileAssertPatterns(parsedOptions.AssertFunctionNames)
			additional := parsedOptions.AdditionalTestBlockFunctions

			isTestBlock := func(node *ast.Node, call *ast.CallExpression) bool {
				if runtime.ClassifyTest != nil {
					classification := runtime.ClassifyTest(node)
					if classification.IsTest {
						return !classification.IsExempt
					}
				}
				if len(additional) == 0 {
					return false
				}
				// CalleeChainName returns "" for every callee it cannot name
				// (`arr[i]()`, `obj[key]()`, `(a ?? b)()`). The option's items carry
				// no minLength, so a configured "" would otherwise turn each of
				// those ordinary calls into a reported test block.
				name := testFramework.CalleeChainName(call.Expression)
				return name != "" && slices.Contains(additional, name)
			}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					call := node.AsCallExpression()
					if call == nil || !isTestBlock(node, call) {
						return
					}
					callback := runtime.Callback(call)
					if callback == nil ||
						isAssertionCall(lastFunctionStatement(callback), runtime, patterns) {
						return
					}
					ctx.ReportNode(call.Expression, buildMustEndWithExpectMessage())
				},
			}
		},
	}
}
