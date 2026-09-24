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
	call             *ast.Node
	callback         *ast.Node
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
	call      *ast.Node
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
				tests         []*testEntry
				testCallbacks = map[*ast.Node]*testEntry{}
				// hookCallbacks maps a hook callback to the hook calls that
				// register it.
				hookCallbacks = map[*ast.Node][]*ast.Node{}
				namedHooks    []namedHook
				// coveringHooks are hook calls that declare assertions for every
				// test they run before.
				coveringHooks []*ast.Node
				staticCalls   []staticEntry
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
						hookCallbacks[function] = append(hookCallbacks[function], hook.call)
					}
				}

				var reports []pendingReport
				for _, entry := range staticCalls {
					if test := testCallbacks[entry.owner]; test != nil {
						if !isFirstStatement(entry.static.Call, entry.owner) {
							continue
						}
						test.satisfied = true
					} else if hookCalls, ok := hookCallbacks[entry.owner]; ok {
						if !runsOnEveryHookCall(entry.static.Call, entry.owner) {
							continue
						}
						coveringHooks = append(coveringHooks, hookCalls...)
					} else {
						continue
					}
					reports = append(reports, checkStaticCall(&ctx, entry.static, opts)...)
				}

				resolver := newSuiteResolver(ctx, func(call *ast.Node) bool {
					return runtime.Classify(call).Kind == RegistrationDescribe
				})
				covered := map[*ast.Node]bool{}
				for _, hook := range coveringHooks {
					for _, suite := range resolver.suitesOf(enclosingFunction(hook)) {
						covered[suite] = true
					}
				}
				// A suite is covered by a hook of its own or of any suite it is
				// nested in. An unknown suite could be any of them.
				suiteCovered := map[*ast.Node]bool{}
				var isSuiteCovered func(suite *ast.Node) bool
				isSuiteCovered = func(suite *ast.Node) bool {
					if suite == unknownSuite {
						return len(covered) > 0
					}
					if covered[suite] {
						return true
					}
					if suite == nil {
						return false
					}
					if result, ok := suiteCovered[suite]; ok {
						return result
					}
					suiteCovered[suite] = false
					for _, parent := range resolver.parentSuites(suite) {
						if isSuiteCovered(parent) {
							suiteCovered[suite] = true
							return true
						}
					}
					return false
				}
				isCovered := func(test *testEntry) bool {
					if covered[unknownSuite] {
						return true
					}
					for _, suite := range resolver.suitesOf(enclosingFunction(test.call)) {
						if isSuiteCovered(suite) {
							return true
						}
					}
					return false
				}

				for _, test := range tests {
					if test.satisfied || isCovered(test) || !shouldCheck(test, opts) {
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
						return
					case RegistrationTest:
						if registration.Callback == nil {
							return
						}
						entry := &testEntry{call: node, callback: registration.Callback}
						tests = append(tests, entry)
						testCallbacks[registration.Callback] = entry
						return
					case RegistrationHook:
						if !slices.Contains(config.CoveringHooks, registration.HookName) || registration.Callback == nil {
							return
						}
						// Type assertions do not change what the hook runs.
						callback := utils.SkipAssertionsAndParens(registration.Callback)
						switch {
						case ast.IsFunctionExpressionOrArrowFunction(callback):
							hookCallbacks[callback] = append(hookCallbacks[callback], node)
						case callback.Kind == ast.KindIdentifier:
							namedHooks = append(namedHooks, namedHook{reference: callback, call: node})
						case runtime.IsHasAssertionsReference != nil && runtime.IsHasAssertionsReference(callback):
							coveringHooks = append(coveringHooks, node)
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
				rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
					finish()
				},
			}
		},
	}
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
// prologue. The call must run whenever that statement runs, so a call nested in
// an inner statement or on one side of `&&` or `?:` does not count.
func isFirstStatement(call *ast.Node, function *ast.Node) bool {
	statement, ok := declarationStatement(call, function)
	if !ok {
		return false
	}
	if statement == nil {
		return true
	}
	for _, candidate := range function.Body().AsBlock().Statements.Nodes {
		if ast.IsPrologueDirective(candidate) {
			continue
		}
		return candidate == statement
	}
	return false
}

// runsOnEveryHookCall reports whether a hook callback makes call each time it
// runs. The call may follow other statements, since setup before the
// declaration still leaves it on every normal path, but it must be a
// top-level statement that runs unconditionally, and no earlier statement may
// return or throw.
func runsOnEveryHookCall(call *ast.Node, function *ast.Node) bool {
	statement, ok := declarationStatement(call, function)
	if !ok {
		return false
	}
	if statement == nil {
		return true
	}
	for _, candidate := range function.Body().AsBlock().Statements.Nodes {
		if candidate == statement {
			return true
		}
		if mayExit(candidate) {
			return false
		}
	}
	return false
}

// declarationStatement returns the statement of function's block body that
// contains call, or nil for a concise arrow body. ok is false when call sits in
// a nested statement or block, or on a path its statement does not always
// evaluate.
func declarationStatement(call *ast.Node, function *ast.Node) (statement *ast.Node, ok bool) {
	body := function.Body()
	if body == nil {
		return nil, false
	}
	child := call
	for current := call.Parent; current != nil; child, current = current, current.Parent {
		if current == body && body.Kind == ast.KindBlock {
			return child, true
		}
		if current == function {
			// Reached from the concise body, not from a parameter default.
			return nil, child == body
		}
		if current.Kind == ast.KindBlock || ast.IsStatement(current) && current.Parent != body {
			return nil, false
		}
		if !evaluatesChild(current, child) {
			return nil, false
		}
	}
	return nil, false
}

// evaluatesChild reports whether evaluating parent always evaluates child. The
// right operand of a short-circuiting operator, the branches of a conditional
// expression, anything after `?.` in an optional chain, and a destructuring
// default may be skipped.
func evaluatesChild(parent *ast.Node, child *ast.Node) bool {
	if ast.IsOptionalChain(parent) {
		// `maybe?.(declare())` and `maybe?.[declare()]` skip everything but the
		// receiver when it is nullish.
		return child == parent.Expression()
	}
	switch parent.Kind {
	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		switch binary.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
			ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return child == binary.Left
		case ast.KindEqualsToken:
			// `[x = declare()] = values` runs the default only for undefined.
			return child == binary.Left || !ast.IsAssignmentTarget(parent)
		}
	case ast.KindBindingElement:
		// `const { x = declare() } = value`
		return child != parent.AsBindingElement().Initializer
	case ast.KindShorthandPropertyAssignment:
		// `({ x = declare() } = value)`
		return child != parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer
	case ast.KindConditionalExpression:
		return child == parent.AsConditionalExpression().Condition
	case ast.KindIfStatement:
		return child == parent.AsIfStatement().Expression
	case ast.KindSwitchStatement:
		return child == parent.AsSwitchStatement().Expression
	case ast.KindExpressionStatement, ast.KindVariableStatement, ast.KindReturnStatement, ast.KindThrowStatement:
		return true
	}
	// Any other statement, such as a loop, may run its expressions zero times
	// or only after its body.
	return !ast.IsStatement(parent)
}

// mayExit reports whether statement contains a return or throw of the
// enclosing function, which could leave a later declaration unexecuted.
func mayExit(statement *ast.Node) bool {
	found := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if found {
			return true
		}
		switch {
		case node.Kind == ast.KindReturnStatement || node.Kind == ast.KindThrowStatement:
			found = true
			return true
		case isFunctionBoundary(node) || ast.IsClassLike(node):
			return false
		}
		return node.ForEachChild(visit)
	}
	visit(statement)
	return found
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
			insertAt, separator := insertionPoint(ctx.SourceFile, body)
			var suggestions []rule.RuleSuggestion
			add := func(message rule.RuleMessage, member string) {
				text := separator + spelling + "." + member + "();"
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
// ordinary expression. separator is the `;` a directive written without one
// needs before the inserted statement; without it the two would run together
// as `'use strict'expect.hasAssertions()`, which does not parse.
func insertionPoint(sourceFile *ast.SourceFile, body *ast.Node) (insertAt int, separator string) {
	insertAt = utils.TrimNodeTextRange(sourceFile, body).Pos() + 1
	for _, statement := range body.AsBlock().Statements.Nodes {
		if !ast.IsPrologueDirective(statement) {
			break
		}
		insertAt = statement.End()
		separator = ""
		if sourceFile.Text()[insertAt-1] != ';' {
			separator = ";"
		}
	}
	return insertAt, separator
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
