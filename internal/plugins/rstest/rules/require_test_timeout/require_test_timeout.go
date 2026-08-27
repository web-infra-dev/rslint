package require_test_timeout

import (
	"strconv"

	"github.com/microsoft/typescript-go/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

var missingTimeoutMessage = rule.RuleMessage{
	Id:          "missingTimeout",
	Description: "Test is missing a timeout. Add an explicit timeout.",
}

// timeoutFinding is what a static read of one registration argument, one
// options object or one runtime-config call was able to conclude.
//
// The rule only reports on timeoutAbsent and timeoutNegative: everything it
// cannot read is timeoutUnknown, which exempts. Upstream instead reports the
// unreadable shapes through its `Number.NaN` branch; reporting a timeout that
// is merely written somewhere unreadable is a false positive, and TypeScript
// already rejects a `timeout` that is not a number.
type timeoutFinding uint8

const (
	// timeoutAbsent means the shape was fully readable and declares no timeout.
	timeoutAbsent timeoutFinding = iota
	// timeoutDeclared means a timeout of zero or more was written down.
	timeoutDeclared
	// timeoutNegative means a timeout below zero was written down. Rstest has
	// no meaning for it, and `0` already spells "no limit", so the author
	// still owes the test a real timeout.
	timeoutNegative
	// timeoutUnknown means the shape could not be read statically.
	timeoutUnknown
)

var RequireTestTimeoutRule = rule.Rule{
	Name:   "rstest/require-test-timeout",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		runtimeConfig := runtimeConfigTracker{ctx: ctx}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// Every `setConfig` / `resetConfig` call has to be seen before
				// the tests that follow it, which the source-order walk already
				// guarantees; the tracker itself allocates nothing until one of
				// those calls actually appears in the file.
				runtimeConfig.observe(node)

				parsed := analysis.ParseTestCall(node)
				if parsed == nil || parsed.Todo || parsed.Skipped {
					return
				}
				// A registration with no callback declares no test body to time
				// out. Reporting it belongs to `rstest/prefer-todo`.
				if callback, name := analysis.TestCallback(node); callback == nil && name == "" {
					return
				}
				if runtimeConfig.exemptsAt(node.Pos()) {
					return
				}
				if inheritsSuiteTimeout(ctx, analysis, node) {
					return
				}
				switch registrationTimeout(ctx, node) {
				case timeoutDeclared, timeoutUnknown:
					return
				}

				ctx.ReportNode(node, missingTimeoutMessage)
			},
		}
	},
}

// registrationTimeout reads the timeout a `test` or `describe` registration
// declares for itself.
//
// Rstest has exactly two overloads — `(name, fn?, timeout?)` and
// `(name, options, fn?)` (packages/core/src/types/api.ts:91-94) — so only the
// second and third arguments can carry a timeout, and a call with more
// arguments than that matches neither overload and is left alone.
func registrationTimeout(ctx rule.RuleContext, node *ast.Node) timeoutFinding {
	call := node.AsCallExpression()
	if call == nil || call.Arguments == nil {
		return timeoutAbsent
	}
	arguments := call.Arguments.Nodes
	if len(arguments) > 3 {
		return timeoutUnknown
	}

	finding := timeoutAbsent
	for index := 1; index < len(arguments) && index <= 2; index++ {
		switch argument := argumentTimeout(ctx, arguments[index]); argument {
		case timeoutNegative:
			return timeoutNegative
		case timeoutDeclared, timeoutUnknown:
			finding = argument
		}
	}
	return finding
}

// argumentTimeout reads one registration argument as a timeout carrier. The
// callback itself carries nothing, so it reads as absent rather than unknown.
func argumentTimeout(ctx rule.RuleContext, node *ast.Node) timeoutFinding {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return timeoutUnknown
	}
	if ast.IsFunctionExpressionOrArrowFunction(node) {
		return timeoutAbsent
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		return objectTimeout(ctx, node, "timeout")
	}
	if finding, ok := numericTimeout(node); ok {
		return finding
	}
	if node.Kind == ast.KindIdentifier {
		// A callback passed by name occupies an argument slot the other
		// overload reads as options or as the timeout, so it has to be
		// recognized as the callback before it is read as either.
		if resolvesToFunction(ctx, node) {
			return timeoutAbsent
		}
		initializer := constInitializer(ctx, node)
		if initializer == nil {
			return timeoutUnknown
		}
		if initializer.Kind == ast.KindObjectLiteralExpression {
			return objectTimeout(ctx, initializer, "timeout")
		}
		if finding, ok := numericTimeout(initializer); ok {
			return finding
		}
	}
	return timeoutUnknown
}

// resolvesToFunction reports whether an identifier names a function declared
// in this file, by either spelling. It matches how the Rstest parser resolves
// a callback passed by name, so that the same argument is not read as a
// callback there and as a timeout here.
func resolvesToFunction(ctx rule.RuleContext, node *ast.Node) bool {
	declaration := internalUtils.GetDeclaration(ctx.TypeChecker, node)
	if declaration == nil {
		return false
	}
	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		return true
	case ast.KindVariableDeclaration:
		initializer := declaration.AsVariableDeclaration().Initializer
		if initializer == nil {
			return false
		}
		return ast.IsFunctionExpressionOrArrowFunction(
			internalUtils.SkipAssertionsAndParens(initializer),
		)
	}
	return false
}

// objectTimeout reads property from an options object literal. A spread or a
// key that is not statically known could carry the property, so either one
// makes the whole object unreadable rather than absent.
func objectTimeout(ctx rule.RuleContext, object *ast.Node, property string) timeoutFinding {
	literal := object.AsObjectLiteralExpression()
	if literal == nil || literal.Properties == nil {
		return timeoutAbsent
	}
	for _, member := range literal.Properties.Nodes {
		if member == nil {
			continue
		}
		if member.Kind == ast.KindSpreadAssignment {
			return timeoutUnknown
		}
		name := member.Name()
		if name == nil {
			return timeoutUnknown
		}
		key, ok := internalUtils.GetStaticPropertyName(name)
		if !ok {
			return timeoutUnknown
		}
		if key != property {
			continue
		}
		switch member.Kind {
		case ast.KindPropertyAssignment:
			return scalarTimeout(ctx, member.AsPropertyAssignment().Initializer)
		case ast.KindShorthandPropertyAssignment:
			return scalarTimeout(ctx, name)
		default:
			// A method or accessor named `timeout` is not a number, and the
			// rule stays silent on everything it cannot read as one.
			return timeoutUnknown
		}
	}
	return timeoutAbsent
}

// scalarTimeout reads a value written in a timeout position: a number, or a
// same-file `const` bound to one. Upstream also resolves `let` bindings and
// reports the ones it fails to resolve; this port keeps the `const` half and
// stays silent on the rest.
func scalarTimeout(ctx rule.RuleContext, node *ast.Node) timeoutFinding {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return timeoutUnknown
	}
	if finding, ok := numericTimeout(node); ok {
		return finding
	}
	if node.Kind == ast.KindIdentifier {
		if initializer := constInitializer(ctx, node); initializer != nil {
			if finding, ok := numericTimeout(initializer); ok {
				return finding
			}
		}
	}
	return timeoutUnknown
}

// numericTimeout classifies a numeric literal, including the unary forms the
// parser keeps separate from the literal itself. `0` disables the timeout in
// Rstest (website/docs/zh/config/test/test-timeout.mdx) and so counts as
// declared; a BigInt is not a number and is left unread.
func numericTimeout(node *ast.Node) (timeoutFinding, bool) {
	negated := false
	if node.Kind == ast.KindPrefixUnaryExpression {
		unary := node.AsPrefixUnaryExpression()
		if unary == nil {
			return timeoutUnknown, false
		}
		switch unary.Operator {
		case ast.KindMinusToken:
			negated = true
		case ast.KindPlusToken:
		default:
			return timeoutUnknown, false
		}
		node = ast.SkipParentheses(unary.Operand)
		if node == nil {
			return timeoutUnknown, false
		}
	}
	if node.Kind != ast.KindNumericLiteral {
		return timeoutUnknown, false
	}
	// tsgo normalizes numeric literal text at parse time, so `5e3` and `0x10`
	// both arrive as their decimal form.
	value, err := strconv.ParseFloat(node.AsNumericLiteral().Text, 64)
	if err != nil {
		return timeoutUnknown, false
	}
	if negated && value > 0 {
		return timeoutNegative, true
	}
	return timeoutDeclared, true
}

// constInitializer returns the initializer of the same-file `const` an
// identifier is bound to. Anything else — `let`, a parameter, an import, a
// function — has no initializer this rule may read.
func constInitializer(ctx rule.RuleContext, node *ast.Node) *ast.Node {
	declaration := internalUtils.GetDeclaration(ctx.TypeChecker, node)
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
		return nil
	}
	list := declaration.Parent
	if list == nil ||
		list.Kind != ast.KindVariableDeclarationList ||
		list.Flags&ast.NodeFlagsConst == 0 {
		return nil
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	if initializer == nil {
		return nil
	}
	return internalUtils.SkipAssertionsAndParens(initializer)
}

// inheritsSuiteTimeout reports whether a lexically enclosing `describe`
// already gives the test a timeout.
//
// Rstest's runner applies suite options as defaults —
// `test.timeout ??= current.timeout` in
// packages/core/src/runtime/runner/runtime.ts — so a test inside a timed suite
// runs with a definite timeout and is not missing one. Upstream models no such
// inheritance because it never looks past the registration itself.
//
// Ownership is read lexically. A function this rule cannot tie back to a call
// argument — a declaration handed to `describe` by name, for instance — could
// belong to any suite, so the walk gives up and exempts the test rather than
// guess. Attributing those needs the callback ownership index that this rule
// deliberately does not build.
func inheritsSuiteTimeout(
	ctx rule.RuleContext,
	analysis *rstestUtils.RstestCallAnalysis,
	testCall *ast.Node,
) bool {
	for node := testCall.Parent; node != nil; node = node.Parent {
		if !isFunctionBoundary(node) {
			continue
		}
		call := enclosingCallOfArgument(node)
		if call == nil {
			return true
		}
		parsed := analysis.ParseFnCall(call)
		if parsed != nil &&
			parsed.Kind == rstestUtils.RstestFnTypeDescribe &&
			registrationTimeout(ctx, call) != timeoutAbsent {
			return true
		}
		// Anything else — another registration, or a plain callback such as
		// `forEach` — still registers the test wherever it is called from, so
		// the walk continues from the call site.
		node = call
	}
	return false
}

func isFunctionBoundary(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindFunctionDeclaration,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindConstructor:
		return true
	}
	return false
}

// enclosingCallOfArgument returns the call node is passed to as an argument,
// looking through the parentheses and TypeScript assertions a callback may be
// wrapped in. A function in callee position, or in no call at all, has no
// enclosing call for this purpose.
func enclosingCallOfArgument(node *ast.Node) *ast.Node {
	child := node
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindAsExpression,
			ast.KindSatisfiesExpression,
			ast.KindNonNullExpression,
			ast.KindTypeAssertionExpression:
			child = parent
			continue
		case ast.KindCallExpression:
			call := parent.AsCallExpression()
			if call == nil || call.Arguments == nil {
				return nil
			}
			for _, argument := range call.Arguments.Nodes {
				if argument == child {
					return parent
				}
			}
		}
		return nil
	}
	return nil
}

// runtimeConfigEvent records one `setConfig` that declared a test timeout, or
// one `resetConfig` that took it back, at the offset the call ends on.
type runtimeConfigEvent struct {
	root   string
	offset int
	sets   bool
}

// runtimeConfigTracker follows the runtime timeout configured through the
// Rstest utility object. Positions are compared rather than scopes because
// `setConfig` takes effect from the moment it runs, which for a file of
// top-level registrations is exactly "from this point in the source down".
type runtimeConfigTracker struct {
	ctx    rule.RuleContext
	events []runtimeConfigEvent
}

func (tracker *runtimeConfigTracker) observe(node *ast.Node) {
	call := node.AsCallExpression()
	if call == nil {
		return
	}
	method, receiver := configMethodAndReceiver(call.Expression)
	if method != "setConfig" && method != "resetConfig" {
		return
	}
	root, ok := runtimeObjectRoot(tracker.ctx, receiver)
	if !ok {
		return
	}
	if method == "resetConfig" {
		tracker.events = append(tracker.events, runtimeConfigEvent{
			root:   root,
			offset: node.End(),
		})
		return
	}
	if !setsTestTimeout(tracker.ctx, call) {
		return
	}
	tracker.events = append(tracker.events, runtimeConfigEvent{
		root:   root,
		offset: node.End(),
		sets:   true,
	})
}

// exemptsAt reports whether a test starting at offset runs under a configured
// test timeout. Events are recorded in source order by the rule's own walk, so
// replaying the ones that end before the test reproduces the state the runtime
// would be in when the test is registered.
func (tracker *runtimeConfigTracker) exemptsAt(offset int) bool {
	if len(tracker.events) == 0 {
		return false
	}
	configured := map[string]bool{}
	for _, event := range tracker.events {
		if event.offset > offset {
			continue
		}
		configured[event.root] = event.sets
	}
	for _, sets := range configured {
		if sets {
			return true
		}
	}
	return false
}

// setsTestTimeout reads the `RuntimeOptions` argument of a `setConfig` call.
// An object this rule cannot read could set `testTimeout`, so it counts as
// setting it; `setConfig({})` and a negative `testTimeout` do not.
func setsTestTimeout(ctx rule.RuleContext, call *ast.CallExpression) bool {
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	argument := internalUtils.SkipAssertionsAndParens(call.Arguments.Nodes[0])
	if argument == nil {
		return true
	}
	if argument.Kind == ast.KindIdentifier {
		if initializer := constInitializer(ctx, argument); initializer != nil {
			argument = initializer
		}
	}
	if argument.Kind != ast.KindObjectLiteralExpression {
		return true
	}
	switch objectTimeout(ctx, argument, "testTimeout") {
	case timeoutDeclared, timeoutUnknown:
		return true
	}
	return false
}

// configMethodAndReceiver splits `<receiver>.setConfig` into its accessor name
// and the object it is read from, in both the dotted and the indexed form.
func configMethodAndReceiver(callee *ast.Node) (string, *ast.Node) {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return "", nil
	}
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "", nil
		}
		return name.AsIdentifier().Text, access.Expression
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		name, ok := internalUtils.GetStaticStringLiteralValue(
			ast.SkipParentheses(access.ArgumentExpression),
		)
		if !ok {
			return "", nil
		}
		return name, access.Expression
	}
	return "", nil
}

// runtimeObjectRoot recognizes the Rstest utility object and returns a key
// that identifies the binding it was reached through, so that a `resetConfig`
// is only paired with a `setConfig` on the same one.
//
// This recognizer is deliberately private to the rule. Modeling the utility
// object in general — its mocking, timer and spy surface — is a separate piece
// of work, and a half-built shared version would invite rules to depend on it.
// It is also deliberately narrow: failing to recognize a root only costs the
// exemption it would have granted.
func runtimeObjectRoot(ctx rule.RuleContext, receiver *ast.Node) (string, bool) {
	receiver = internalUtils.SkipAssertionsAndParens(receiver)
	if receiver == nil {
		return "", false
	}
	if isImportMetaRstest(receiver) {
		return "import.meta.rstest", true
	}
	if receiver.Kind == ast.KindIdentifier {
		local := receiver.AsIdentifier().Text
		// `rs` and `rstest` are the same object under two names
		// (packages/core/src/runtime/api/public.ts:63-64), and either may be
		// imported under a further name of its own.
		name, _, _ := testFramework.ResolveFunctionIdentifierReference(
			local,
			receiver,
			ctx.TypeChecker,
			ctx.SourceFile,
			rstestUtils.RstestImportModule,
		)
		if name != "rs" && name != "rstest" {
			return "", false
		}
		return "identifier:" + local, true
	}

	// A namespace import or whole-module require reaches the same object
	// through a member: `core.rs.setConfig(...)`.
	member, namespace := configMethodAndReceiver(receiver)
	if member != "rs" && member != "rstest" {
		return "", false
	}
	namespace = internalUtils.SkipAssertionsAndParens(namespace)
	if namespace == nil || namespace.Kind != ast.KindIdentifier || ctx.TypeChecker == nil {
		return "", false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(namespace)
	if !testFramework.IsModuleNamespaceSymbol(symbol, rstestUtils.RstestImportModule) {
		return "", false
	}
	return "namespace:" + namespace.AsIdentifier().Text + "." + member, true
}

// isImportMetaRstest reports whether node is `import.meta.rstest`, the third
// spelling of the utility object.
func isImportMetaRstest(node *ast.Node) bool {
	name, receiver := configMethodAndReceiver(node)
	if name != "rstest" || receiver == nil {
		return false
	}
	receiver = ast.SkipParentheses(receiver)
	return receiver != nil && receiver.Kind == ast.KindMetaProperty &&
		receiver.AsMetaProperty().KeywordToken == ast.KindImportKeyword
}
