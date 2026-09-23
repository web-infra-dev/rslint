// Package prefer_expect_assertions holds the framework-neutral engine shared by
// jest/prefer-expect-assertions and rstest/prefer-expect-assertions.
//
// A test satisfies the rule when its callback starts with
// `expect.assertions(n)` or `expect.hasAssertions()`, or when an each-hook of an
// enclosing suite makes one of those calls for it. Which hooks run early enough
// to count is a runtime fact that differs between frameworks: Jest verifies the
// assertion count after afterEach hooks, while Rstest verifies it before them.
// Adapters therefore name the covering hooks, and they own every provenance
// question: which calls are test, describe and hook registrations, which
// callback a registration runs, which calls are the framework's expect, and how
// a suggestion may spell expect at the insertion point.
//
// Decisions are taken once the whole file has been seen. A hook covers every
// test of its suite regardless of where it is declared, because frameworks
// collect hooks before running any test of the suite, and a named hook
// callback can be declared anywhere in the file.
package prefer_expect_assertions

import (
	_ "embed"
	"math"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

// Schema accepts the three only* options. A framework that also offers
// disallowHasAssertions supplies its own schema.
//
//go:embed prefer_expect_assertions.schema.json
var Schema []byte

const (
	memberAssertions    = "assertions"
	memberHasAssertions = "hasAssertions"
)

type RegistrationKind uint8

const (
	RegistrationNone RegistrationKind = iota
	RegistrationTest
	RegistrationDescribe
	RegistrationHook
)

// Registration is what an adapter reports about one call.
type Registration struct {
	Kind RegistrationKind
	// HookName is the resolved hook name, such as "beforeEach". It is set only
	// for RegistrationHook.
	HookName string
	// Callback is the function literal a test registration runs, or the raw
	// argument a hook registration runs. A test whose callback is not a function
	// literal written at the registration is not checked: a callback declared
	// elsewhere may be shared by several tests, and a suggestion would edit all
	// of them.
	Callback *ast.Node
}

// StaticCall is a call of `expect.assertions` or `expect.hasAssertions`.
type StaticCall struct {
	Call *ast.Node
	// Member is memberAssertions or memberHasAssertions.
	Member string
	// MemberNode is the node naming the member: an identifier for dot access,
	// a string literal for bracket access.
	MemberNode *ast.Node
}

type Runtime struct {
	Classify func(call *ast.Node) Registration
	// ParseStatic recognizes a static expect call made through the framework's
	// own expect, in any provenance the framework supports.
	ParseStatic func(call *ast.Node) *StaticCall
	// IsExpect recognizes any call of the framework's expect. It feeds the
	// onlyFunctionsWithExpectInLoop and onlyFunctionsWithExpectInCallback
	// options.
	IsExpect func(call *ast.Node) bool
	// IsHasAssertionsReference recognizes a hook argument that is the
	// hasAssertions method itself, as in `beforeEach(expect.hasAssertions)`. A
	// framework whose hasAssertions cannot run detached leaves it nil.
	IsHasAssertionsReference func(node *ast.Node) bool
	// ExpectSpelling returns how a statement inserted at the start of the test
	// callback refers to expect. ok is false when no spelling is known to
	// resolve to the framework's expect there, which withholds the suggestions.
	ExpectSpelling func(test *ast.Node, callback *ast.Node) (spelling string, ok bool)
}

type Config struct {
	Name   string
	Schema []byte
	// CoveringHooks lists the hooks whose static expect calls count for every
	// test in their suite.
	CoveringHooks []string
	// DisallowHasAssertions enables the disallowHasAssertions option.
	DisallowHasAssertions bool
	Prepare               func(ctx rule.RuleContext) Runtime
}

type options struct {
	onlyAsync               bool
	onlyExpectInLoop        bool
	onlyExpectInCallback    bool
	disallowHasAssertions   bool
	anyOnlyFunctionsEnabled bool
}

func parseOptions(raw []any, config Config) options {
	var opts options
	if len(raw) == 0 {
		return opts
	}
	values, _ := raw[0].(map[string]any)
	flag := func(name string) bool {
		value, _ := values[name].(bool)
		return value
	}
	opts.onlyAsync = flag("onlyFunctionsWithAsyncKeyword")
	opts.onlyExpectInLoop = flag("onlyFunctionsWithExpectInLoop")
	opts.onlyExpectInCallback = flag("onlyFunctionsWithExpectInCallback")
	opts.disallowHasAssertions = config.DisallowHasAssertions && flag("disallowHasAssertions")
	opts.anyOnlyFunctionsEnabled = opts.onlyAsync || opts.onlyExpectInLoop || opts.onlyExpectInCallback
	return opts
}

var (
	msgHasAssertionsTakesNoArguments = rule.RuleMessage{
		Id:          "hasAssertionsTakesNoArguments",
		Description: "`expect.hasAssertions` expects no arguments",
	}
	msgAssertionsRequiresOneArgument = rule.RuleMessage{
		Id:          "assertionsRequiresOneArgument",
		Description: "`expect.assertions` expects a single argument of type number",
	}
	msgAssertionsRequiresNumberArgument = rule.RuleMessage{
		Id:          "assertionsRequiresNumberArgument",
		Description: "This argument should be a number",
	}
	msgHaveExpectAssertions = rule.RuleMessage{
		Id:          "haveExpectAssertions",
		Description: "Every test should have either `expect.assertions(<number of assertions>)` or `expect.hasAssertions()` as its first expression",
	}
	msgPreferAssertionsOverHasAssertions = rule.RuleMessage{
		Id:          "preferAssertionsOverHasAssertions",
		Description: "Prefer `expect.assertions(<number of assertions>)` over `expect.hasAssertions()`",
	}
	msgSuggestAddingHasAssertions = rule.RuleMessage{
		Id:          "suggestAddingHasAssertions",
		Description: "Add `expect.hasAssertions()`",
	}
	msgSuggestAddingAssertions = rule.RuleMessage{
		Id:          "suggestAddingAssertions",
		Description: "Add `expect.assertions(<number of assertions>)`",
	}
	msgSuggestReplacingWithAssertions = rule.RuleMessage{
		Id:          "suggestReplacingWithAssertions",
		Description: "Replace with `expect.assertions(<number of assertions>)`",
	}
	msgSuggestRemovingExtraArguments = rule.RuleMessage{
		Id:          "suggestRemovingExtraArguments",
		Description: "Remove extra arguments",
	}
)

type testEntry struct {
	call     *ast.Node
	callback *ast.Node
	// suite is the innermost describe call the test is registered in, or nil at
	// file level.
	suite            *ast.Node
	satisfied        bool
	hasExpectInLoop  bool
	hasExpectInCalls bool
}

type staticEntry struct {
	static *StaticCall
	owner  *ast.Node
}

type namedHook struct {
	reference *ast.Node
	suite     *ast.Node
}

// pendingReport keeps diagnostics in source order: argument diagnostics are
// found before the tests they sit in are decided.
type pendingReport struct {
	pos  int
	emit func()
}

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:   config.Name,
		Schema: rule.NewSchema(config.Schema),
		Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
			opts := parseOptions(rawOptions, config)
			runtime := config.Prepare(ctx)

			var (
				suiteStack    []*ast.Node
				enteredSuites = map[*ast.Node]bool{}
				suiteParents  = map[*ast.Node]*ast.Node{}
				tests         []*testEntry
				testCallbacks = map[*ast.Node]*testEntry{}
				hookCallbacks = map[*ast.Node][]*ast.Node{}
				namedHooks    []namedHook
				coveredSuites = map[*ast.Node]bool{}
				fileCovered   bool
				staticCalls   []staticEntry
				currentSuite  = func() *ast.Node { return lastOrNil(suiteStack) }
				coverSuite    = func(suite *ast.Node) {
					if suite == nil {
						fileCovered = true
						return
					}
					coveredSuites[suite] = true
				}
			)

			recordExpect := func(node *ast.Node) {
				inLoop, inCallback := false, false
				for current := node.Parent; current != nil; current = current.Parent {
					if isLoop(current) {
						inLoop = true
					}
					if !isFunctionBoundary(current) {
						continue
					}
					if entry := testCallbacks[current]; entry != nil {
						entry.hasExpectInLoop = entry.hasExpectInLoop || inLoop
						entry.hasExpectInCalls = entry.hasExpectInCalls || inCallback
						return
					}
					inCallback = true
				}
			}

			finish := func() {
				for _, hook := range namedHooks {
					if function := resolveLocalFunction(ctx, hook.reference); function != nil {
						hookCallbacks[function] = append(hookCallbacks[function], hook.suite)
					}
				}

				var reports []pendingReport
				for _, entry := range staticCalls {
					if test := testCallbacks[entry.owner]; test != nil {
						if !isFirstStatement(entry.static.Call, entry.owner) {
							continue
						}
						test.satisfied = true
					} else if suites, ok := hookCallbacks[entry.owner]; ok {
						for _, suite := range suites {
							coverSuite(suite)
						}
					} else {
						continue
					}
					reports = append(reports, checkStaticCall(&ctx, entry.static, opts)...)
				}

				isCovered := func(suite *ast.Node) bool {
					if fileCovered {
						return true
					}
					for ; suite != nil; suite = suiteParents[suite] {
						if coveredSuites[suite] {
							return true
						}
					}
					return false
				}

				for _, test := range tests {
					if test.satisfied || isCovered(test.suite) || !shouldCheck(test, opts) {
						continue
					}
					reports = append(reports, reportMissingAssertions(&ctx, runtime, test, opts))
				}

				slices.SortStableFunc(reports, func(a, b pendingReport) int { return a.pos - b.pos })
				for _, report := range reports {
					report.emit()
				}
			}

			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					registration := runtime.Classify(node)
					switch registration.Kind {
					case RegistrationDescribe:
						suiteParents[node] = currentSuite()
						suiteStack = append(suiteStack, node)
						enteredSuites[node] = true
						return
					case RegistrationTest:
						if registration.Callback == nil {
							return
						}
						entry := &testEntry{call: node, callback: registration.Callback, suite: currentSuite()}
						tests = append(tests, entry)
						testCallbacks[registration.Callback] = entry
						return
					case RegistrationHook:
						if !slices.Contains(config.CoveringHooks, registration.HookName) || registration.Callback == nil {
							return
						}
						suite := currentSuite()
						// Type assertions do not change what the hook runs.
						callback := utils.SkipAssertionsAndParens(registration.Callback)
						switch {
						case ast.IsFunctionExpressionOrArrowFunction(callback):
							hookCallbacks[callback] = append(hookCallbacks[callback], suite)
						case callback.Kind == ast.KindIdentifier:
							namedHooks = append(namedHooks, namedHook{reference: callback, suite: suite})
						case runtime.IsHasAssertionsReference != nil && runtime.IsHasAssertionsReference(callback):
							coverSuite(suite)
						}
						return
					}

					if static := runtime.ParseStatic(node); static != nil {
						staticCalls = append(staticCalls, staticEntry{static: static, owner: enclosingFunction(node)})
					}
					// Only the loop and callback options read where expect is called, so
					// other configurations skip resolving every call.
					if (opts.onlyExpectInLoop || opts.onlyExpectInCallback) && runtime.IsExpect(node) {
						recordExpect(node)
					}
				},
				rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
					if !enteredSuites[node] {
						return
					}
					delete(enteredSuites, node)
					suiteStack = suiteStack[:len(suiteStack)-1]
				},
				rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
					finish()
				},
			}
		},
	}
}

func lastOrNil(nodes []*ast.Node) *ast.Node {
	if len(nodes) == 0 {
		return nil
	}
	return nodes[len(nodes)-1]
}

// shouldCheck applies the only* options. They are alternatives: a test is
// checked when it matches any enabled one.
func shouldCheck(test *testEntry, opts options) bool {
	if !opts.anyOnlyFunctionsEnabled {
		return true
	}
	return opts.onlyAsync && ast.IsAsyncFunction(test.callback) ||
		opts.onlyExpectInLoop && test.hasExpectInLoop ||
		opts.onlyExpectInCallback && test.hasExpectInCalls
}

// isLoop covers every native loop statement. An expect call only counts as
// inside a loop when the loop sits between it and the test callback, so a test
// registered inside a loop is not mistaken for one that asserts in a loop.
func isLoop(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
		ast.KindWhileStatement, ast.KindDoStatement:
		return true
	}
	return false
}

func isFunctionBoundary(node *ast.Node) bool {
	return ast.IsFunctionLikeOrClassStaticBlockDeclaration(node)
}

func enclosingFunction(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if isFunctionBoundary(current) {
			return current
		}
	}
	return nil
}

// isFirstStatement reports whether call is the test callback's first
// expression: the concise body of an arrow function, or an expression inside
// the first statement of the callback's own block after its directive
// prologue. A call nested in an inner statement, such as the branch of an
// `if`, only runs on some paths and does not count.
func isFirstStatement(call *ast.Node, function *ast.Node) bool {
	body := function.Body()
	if body == nil {
		return false
	}
	if body.Kind != ast.KindBlock {
		return true
	}
	statement := call
	for statement.Parent != body {
		statement = statement.Parent
		if statement == nil || statement.Kind == ast.KindBlock {
			return false
		}
		if statement.Parent != body && ast.IsStatement(statement) {
			return false
		}
	}
	for _, candidate := range body.AsBlock().Statements.Nodes {
		if ast.IsPrologueDirective(candidate) {
			continue
		}
		return candidate == statement
	}
	return false
}

func checkStaticCall(ctx *rule.RuleContext, static *StaticCall, opts options) []pendingReport {
	call := static.Call.AsCallExpression()
	var arguments []*ast.Node
	if call.Arguments != nil {
		arguments = call.Arguments.Nodes
	}
	memberRange := utils.TrimNodeTextRange(ctx.SourceFile, static.MemberNode)

	if static.Member == memberHasAssertions {
		if len(arguments) > 0 {
			return []pendingReport{{pos: memberRange.Pos(), emit: func() {
				ctx.ReportRangeWithDeferredSuggestions(memberRange, msgHasAssertionsTakesNoArguments,
					removeExtraArgumentsSuggestion(ctx, static.Call, 0))
			}}}
		}
		if opts.disallowHasAssertions {
			return []pendingReport{{pos: memberRange.Pos(), emit: func() {
				ctx.ReportRangeWithDeferredSuggestions(memberRange, msgPreferAssertionsOverHasAssertions,
					func() []rule.RuleSuggestion {
						fix, ok := replaceMemberName(ctx.SourceFile, static.MemberNode, memberAssertions)
						if !ok {
							return nil
						}
						return []rule.RuleSuggestion{{Message: msgSuggestReplacingWithAssertions, FixesArr: []rule.RuleFix{fix}}}
					})
			}}}
		}
		return nil
	}

	if len(arguments) != 1 {
		if len(arguments) == 0 {
			return []pendingReport{{pos: memberRange.Pos(), emit: func() {
				ctx.ReportRange(memberRange, msgAssertionsRequiresOneArgument)
			}}}
		}
		extra := utils.TrimNodeTextRange(ctx.SourceFile, arguments[1])
		return []pendingReport{{pos: extra.Pos(), emit: func() {
			ctx.ReportRangeWithDeferredSuggestions(extra, msgAssertionsRequiresOneArgument,
				removeExtraArgumentsSuggestion(ctx, static.Call, 1))
		}}}
	}

	argument := arguments[0]
	if isIntegerLiteral(argument) {
		return nil
	}
	argumentRange := utils.TrimNodeTextRange(ctx.SourceFile, ast.SkipParentheses(argument))
	return []pendingReport{{pos: argumentRange.Pos(), emit: func() {
		ctx.ReportRange(argumentRange, msgAssertionsRequiresNumberArgument)
	}}}
}

// isIntegerLiteral mirrors `Number.isInteger(node.value)` on a numeric
// literal. A bigint literal, a string, a signed number or any other expression
// is not a number literal and is rejected.
func isIntegerLiteral(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindNumericLiteral {
		return false
	}
	value, ok := ecmascript.StringToNumber(node.AsNumericLiteral().Text)
	return ok && isFiniteInteger(value)
}

func isFiniteInteger(value float64) bool {
	return !math.IsInf(value, 0) && !math.IsNaN(value) && math.Trunc(value) == value
}

// removeExtraArgumentsSuggestion deletes arguments from index from through the
// last one, keeping a trailing comma that follows the kept arguments.
func removeExtraArgumentsSuggestion(ctx *rule.RuleContext, call *ast.Node, from int) func() []rule.RuleSuggestion {
	return func() []rule.RuleSuggestion {
		listRange, ok := testFramework.CallArgumentListRange(ctx.SourceFile, call)
		if !ok {
			return nil
		}
		arguments := call.AsCallExpression().Arguments.Nodes
		start := utils.TrimNodeTextRange(ctx.SourceFile, arguments[from]).Pos()
		return []rule.RuleSuggestion{{
			Message:  msgSuggestRemovingExtraArguments,
			FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, listRange.End()))},
		}}
	}
}

// replaceMemberName renames the accessed member, keeping the quotes of a
// bracket access written with a string literal.
func replaceMemberName(sourceFile *ast.SourceFile, member *ast.Node, name string) (rule.RuleFix, bool) {
	textRange := utils.TrimNodeTextRange(sourceFile, member)
	switch member.Kind {
	case ast.KindIdentifier:
		return rule.RuleFixReplaceRange(textRange, name), true
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		if textRange.Len() < 2 {
			return rule.RuleFix{}, false
		}
		return rule.RuleFixReplaceRange(core.NewTextRange(textRange.Pos()+1, textRange.End()-1), name), true
	}
	return rule.RuleFix{}, false
}

func reportMissingAssertions(ctx *rule.RuleContext, runtime Runtime, test *testEntry, opts options) pendingReport {
	callRange := utils.TrimNodeTextRange(ctx.SourceFile, test.call)
	return pendingReport{pos: callRange.Pos(), emit: func() {
		ctx.ReportRangeWithDeferredSuggestions(callRange, msgHaveExpectAssertions, func() []rule.RuleSuggestion {
			body := test.callback.Body()
			if body == nil || body.Kind != ast.KindBlock {
				return nil
			}
			spelling, ok := runtime.ExpectSpelling(test.call, test.callback)
			if !ok {
				return nil
			}
			insertAt := insertionPoint(ctx.SourceFile, body)
			var suggestions []rule.RuleSuggestion
			add := func(message rule.RuleMessage, member string) {
				text := spelling + "." + member + "();"
				suggestions = append(suggestions, rule.RuleSuggestion{
					Message:  message,
					FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(insertAt, insertAt), text)},
				})
			}
			if !opts.disallowHasAssertions {
				add(msgSuggestAddingHasAssertions, memberHasAssertions)
			}
			add(msgSuggestAddingAssertions, memberAssertions)
			return suggestions
		})
	}}
}

// insertionPoint is just inside the opening brace, or after the directive
// prologue so an inserted statement cannot turn `'use strict'` into an
// ordinary expression.
func insertionPoint(sourceFile *ast.SourceFile, body *ast.Node) int {
	insertAt := utils.TrimNodeTextRange(sourceFile, body).Pos() + 1
	for _, statement := range body.AsBlock().Statements.Nodes {
		if !ast.IsPrologueDirective(statement) {
			break
		}
		insertAt = statement.End()
	}
	return insertAt
}

// resolveLocalFunction returns the function a hook argument names when the
// binding is declared once in this file, initialized with a function, and
// never reassigned.
func resolveLocalFunction(ctx rule.RuleContext, reference *ast.Node) *ast.Node {
	if ctx.Refs == nil {
		return nil
	}
	symbol := ctx.Refs.Resolve(reference)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return nil
	}
	declaration := symbol.Declarations[0]
	if declaration == nil || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
		return nil
	}
	for _, use := range ctx.Refs.References(symbol) {
		if utils.IsWriteReference(use) {
			return nil
		}
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return declaration
	case ast.KindVariableDeclaration:
		initializer := declaration.AsVariableDeclaration().Initializer
		if initializer == nil {
			return nil
		}
		initializer = utils.SkipAssertionsAndParens(initializer)
		if ast.IsFunctionExpressionOrArrowFunction(initializer) {
			return initializer
		}
	}
	return nil
}

// ParseStaticMember reports the static member a callee names, returning ok
// only for assertions and hasAssertions accessed with a dot or with a string
// literal key.
func ParseStaticMember(member *ast.Node) (name string, ok bool) {
	if member == nil {
		return "", false
	}
	switch member.Kind {
	case ast.KindIdentifier:
		name = member.AsIdentifier().Text
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		name = member.Text()
	default:
		return "", false
	}
	return name, name == memberAssertions || name == memberHasAssertions
}

// IsStaticMemberName reports whether name is one of the two static members
// this rule inspects.
func IsStaticMemberName(name string) bool {
	return name == memberAssertions || name == memberHasAssertions
}
