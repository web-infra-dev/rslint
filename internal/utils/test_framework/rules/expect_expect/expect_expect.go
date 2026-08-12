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
	"regexp"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
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

func compileAssertPatterns(patterns []string) []*regexp.Regexp {
	out := make([]*regexp.Regexp, 0, len(patterns))
	for _, p := range patterns {
		out = append(out, compileAssertPattern(p))
	}
	return out
}

func compileAssertPattern(pattern string) *regexp.Regexp {
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
	re, err := regexp.Compile(`(?i)^(?:` + joined + `)(?:\.|$)`)
	if err != nil {
		return nil
	}
	return re
}

func matchesAssertName(name string, compiled []*regexp.Regexp) bool {
	if name == "" {
		return false
	}
	for _, re := range compiled {
		if re != nil && re.MatchString(name) {
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

					if !matchesAssertName(calleeName, compiled) {
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
