package no_async_mock_factory

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	lintprogram "github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// Rstest installs the second argument of a module mock as the module's own
// factory and calls it while the module graph is being built, so it has to hand
// back the module object there and then. The generated wrapper checks the
// result with `instanceof Promise` and throws
// "[Rstest] An async mock factory is not supported." when it matches
// (packages/core/src/core/plugins/mockRuntimeCode.js). `mock`, `doMock`,
// `mockRequire` and `doMockRequire` are all produced by the same
// `getMockImplementation`, so the four behave identically here.
//
// Nothing in the type system stops it: the factory parameter is
// `MockFactory<T> = () => Partial<T>` (packages/core/src/types/mock.ts), a
// string module path leaves `T` as `unknown`, and every promise is assignable
// to `Partial<unknown>`. The throw surfaces only once the mocked module is
// really loaded, which is why it is worth catching in the editor.
//
// The gate is "returns a Promise", not "is declared async": a plain function
// returning `Promise.resolve(…)` throws just the same, and an async generator
// hands back an AsyncGenerator rather than a promise and does not throw. A
// hand-written thenable is not caught by `instanceof Promise` either; it breaks
// differently and is out of scope.

var mockAPIs = map[string]bool{
	"mock":          true,
	"doMock":        true,
	"mockRequire":   true,
	"doMockRequire": true,
}

// asyncModuleLoaders are the plugin-managed members that load a module
// asynchronously, so a factory built on either can only hand back a promise.
// Their `require` twins are synchronous and are not listed.
//
// The two reach their failure by different routes on @rstest/core 0.11.8.
// `() => rs.importActual('./m')` runs and hits the runtime's promise check.
// `() => rs.importMock('./m')` throws "importMock() was not transformed by
// Rstest" instead, because that call site is not rewritten inside a factory the
// build hoists. Both are reported under the one message: the factory is
// asynchronous either way and the mock cannot take effect.
var asyncModuleLoaders = map[string]bool{
	"importActual": true,
	"importMock":   true,
}

// promiseStatics are the `Promise` statics whose result is a promise. `Promise`
// also carries `withResolvers`, whose result is an object holding one, so it is
// deliberately absent.
var promiseStatics = map[string]bool{
	"resolve":    true,
	"reject":     true,
	"all":        true,
	"allSettled": true,
	"race":       true,
	"any":        true,
}

func buildAsyncMockFactoryMessage(member string) rule.RuleMessage {
	return rule.RuleMessage{
		Id: "asyncMockFactory",
		Description: "Mock factory must return the module object synchronously; '" + member +
			"()' throws when the factory returns a Promise. To keep part of the original module, " +
			"import it with `with { rstest: 'importActual' }` and spread it in.",
		Data: map[string]string{"member": member},
	}
}

var removeAsyncMessage = rule.RuleMessage{
	Id:          "suggestRemoveAsync",
	Description: "Remove 'async' so the factory returns the module object directly.",
}

var unwrapPromiseResolveMessage = rule.RuleMessage{
	Id:          "suggestUnwrapPromiseResolve",
	Description: "Return the module object directly instead of wrapping it in 'Promise.resolve()'.",
}

// verdict is what the layers below conclude about a factory argument.
type verdict int

const (
	// verdictUnknown means the layer could not decide and the next one has to
	// look. It is the answer that is reported on only when a TypeChecker
	// settles it, so a source-only program stays silent rather than guessing.
	verdictUnknown verdict = iota
	// verdictPromise means the factory certainly hands back a promise.
	verdictPromise
	// verdictSync means it certainly does not.
	verdictSync
)

var NoAsyncMockFactoryRule = rule.Rule{
	Name:   "rstest/no-async-mock-factory",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				argument, member := mockFactoryArgument(node)
				if argument == nil {
					return
				}

				factory := utils.SkipAssertionsAndParens(argument)
				if factory == nil {
					return
				}

				switch classify(ctx, factory) {
				case verdictSync, verdictUnknown:
					return
				}

				ctx.ReportNodeWithDeferredSuggestions(
					argument,
					buildAsyncMockFactoryMessage(member),
					func() []rule.RuleSuggestion { return suggestions(ctx, factory) },
				)
			},
		}
	},
}

// mockFactoryArgument returns the factory argument of a module mock call the
// build rewrites, together with the member as it is written. It is nil for
// everything else.
func mockFactoryArgument(node *ast.Node) (*ast.Node, string) {
	utility := rstestUtils.ParseRstestPluginManagedCall(node)
	if utility == nil || !mockAPIs[utility.Member] || !rstestUtils.IsTransformablePosition(node) {
		return nil, ""
	}

	call := node.AsCallExpression()
	if call.Arguments == nil {
		return nil, ""
	}
	// The transform reads the path and the factory positionally. A third
	// argument fails the build, and a spread's elements are only known at run
	// time, so neither shape reaches a factory this rule can judge. What the
	// path itself is written as does not matter: a string, a dynamic import
	// and a variable all install the same factory. An explicit type argument
	// only names the mocked module's shape and is no exemption either.
	arguments := call.Arguments.Nodes
	if len(arguments) != 2 {
		return nil, ""
	}
	for _, argument := range arguments {
		if argument == nil || argument.Kind == ast.KindSpreadElement {
			return nil, ""
		}
	}
	return arguments[1], utility.Member
}

// classify walks the three layers in order: what the syntax settles on its own,
// what a declaration in this file settles, and — only for what is left — what
// the types say.
func classify(ctx rule.RuleContext, factory *ast.Node) verdict {
	if v := classifySyntactically(ctx, factory); v != verdictUnknown {
		return v
	}
	if v := classifyThroughDeclaration(ctx, factory); v != verdictUnknown {
		return v
	}
	return classifyByType(ctx, factory)
}

// classifySyntactically answers from the argument alone, without resolving a
// binding and without types.
func classifySyntactically(ctx rule.RuleContext, factory *ast.Node) verdict {
	// An object literal in this position is the mock options bag, not a
	// factory.
	if factory.Kind == ast.KindObjectLiteralExpression {
		return verdictSync
	}
	if !isFunctionExpressionLike(factory) {
		return verdictUnknown
	}
	return classifyFunction(ctx, factory)
}

func classifyFunction(ctx rule.RuleContext, fn *ast.Node) verdict {
	flags := ast.GetFunctionFlags(fn)
	// A generator hands back a Generator and an async generator an
	// AsyncGenerator. Neither is a promise, so neither trips the runtime gate.
	if flags&ast.FunctionFlagsGenerator != 0 {
		return verdictSync
	}
	if flags&ast.FunctionFlagsAsync != 0 {
		return verdictPromise
	}

	// A path that runs off the end of a block body hands back `undefined`,
	// which the runtime's `instanceof Promise` check lets through: such a
	// factory only sometimes returns a promise, and this rule reports the ones
	// that certainly do. `async` is settled above, where the implicit
	// `undefined` is wrapped like every other value the body hands back.
	if body := fn.Body(); body != nil && body.Kind == ast.KindBlock &&
		utils.IsFunctionEndReachable(fn) {
		return verdictUnknown
	}

	returned, ok := returnedExpressions(fn)
	if !ok || len(returned) == 0 {
		return verdictUnknown
	}
	for _, expression := range returned {
		if !isPromiseProducingExpression(ctx, expression) {
			return verdictUnknown
		}
	}
	return verdictPromise
}

// classifyThroughDeclaration follows an identifier to a declaration in this
// file. A name that resolves to more than one declaration, to another file, to
// anything but a `const` function or a function declaration, or to a function
// declaration that is written to, is left to the type layer, where a
// reassignment or a re-export is accounted for.
func classifyThroughDeclaration(ctx rule.RuleContext, factory *ast.Node) verdict {
	if factory.Kind != ast.KindIdentifier || ctx.Refs == nil {
		return verdictUnknown
	}
	symbol := ctx.Refs.Resolve(factory)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return verdictUnknown
	}
	declaration := symbol.Declarations[0]
	if declaration == nil || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
		return verdictUnknown
	}

	switch declaration.Kind {
	case ast.KindFunctionDeclaration:
		// A function declaration's name is a writable binding, and an
		// assignment to it adds no declaration of its own: the body below is
		// the installed factory only while nothing writes to the name.
		if isWrittenTo(ctx, symbol) {
			return verdictUnknown
		}
		return classifyFunction(ctx, declaration)
	case ast.KindVariableDeclaration:
		if !ast.IsVarConst(declaration) {
			return verdictUnknown
		}
		initializer := utils.SkipAssertionsAndParens(declaration.Initializer())
		if initializer == nil || !isFunctionExpressionLike(initializer) {
			return verdictUnknown
		}
		return classifyFunction(ctx, initializer)
	}
	return verdictUnknown
}

// isWrittenTo reports whether anything in this file assigns to the symbol,
// which would replace the value its declaration describes.
func isWrittenTo(ctx rule.RuleContext, symbol *ast.Symbol) bool {
	for _, reference := range ctx.Refs.References(symbol) {
		if utils.IsWriteReference(reference) {
			return true
		}
	}
	return false
}

// classifyByType asks the TypeChecker, and only ever answers verdictPromise or
// verdictUnknown: it exists to decide the cases syntax could not, never to
// overrule it.
//
// A source-only program has no TypeChecker, and the rule then reports what the
// first two layers found and nothing more.
func classifyByType(ctx rule.RuleContext, factory *ast.Node) verdict {
	if ctx.TypeChecker == nil {
		return verdictUnknown
	}
	if alwaysReturnsPromise(ctx.Program(), ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(factory)) {
		return verdictPromise
	}
	return verdictUnknown
}

// alwaysReturnsPromise reports whether calling a value of type t can only
// produce a promise. Every union member has to be callable and every one of
// their signatures has to return a promise: a type that mixes an async factory
// with a synchronous one leaves the outcome to run time, and this rule stays
// silent rather than report a call that may well work.
func alwaysReturnsPromise(sourceProgram *lintprogram.Program, typeChecker *checker.Checker, t *checker.Type) bool {
	if t == nil {
		return false
	}
	if utils.IsUnionType(t) {
		parts := t.Types()
		if len(parts) == 0 {
			return false
		}
		for _, part := range parts {
			if !alwaysReturnsPromise(sourceProgram, typeChecker, part) {
				return false
			}
		}
		return true
	}

	signatures := utils.GetCallSignatures(typeChecker, t)
	if len(signatures) == 0 {
		return false
	}
	for _, signature := range signatures {
		if !isPromiseInstanceType(sourceProgram, typeChecker, checker.Checker_getReturnTypeOfSignature(typeChecker, signature), 0) {
			return false
		}
	}
	return true
}

// isPromiseInstanceType reports whether a value of type t is an instance of
// the global `Promise`. The name alone does not settle it: a file is free to
// declare its own `Promise`, and a value of that type is not what the
// runtime's `instanceof Promise` looks for, so the symbol has to come from the
// default library.
//
// utils.IsThenableType is deliberately not used. It answers "has a `then` whose
// first parameter is a callback", which is wider than the `instanceof Promise`
// the runtime performs, and the gap is reachable: mocking a module that itself
// exports `then` gives the factory a thenable return type while the runtime
// accepts it happily.
func isPromiseInstanceType(sourceProgram *lintprogram.Program, typeChecker *checker.Checker, t *checker.Type, depth int) bool {
	if t == nil || depth > 8 {
		return false
	}
	if utils.IsTypeFlagSet(t, checker.TypeFlagsAnyOrUnknown) {
		return false
	}
	if utils.IsUnionType(t) {
		parts := t.Types()
		if len(parts) == 0 {
			return false
		}
		for _, part := range parts {
			if !isPromiseInstanceType(sourceProgram, typeChecker, part, depth+1) {
				return false
			}
		}
		return true
	}
	if utils.IsIntersectionType(t) {
		// `Promise<T> & Branded` is still a promise at run time.
		for _, part := range t.Types() {
			if isPromiseInstanceType(sourceProgram, typeChecker, part, depth+1) {
				return true
			}
		}
		return false
	}

	resolved := t
	if utils.IsTypeFlagSet(resolved, checker.TypeFlagsObject) &&
		checker.Type_objectFlags(resolved)&checker.ObjectFlagsReference != 0 {
		resolved = resolved.Target()
	}
	if symbol := checker.Type_symbol(resolved); symbol != nil && symbol.Name == "Promise" &&
		utils.IsSymbolFromDefaultLibrary(sourceProgram, symbol) {
		return true
	}
	if checker.Type_objectFlags(resolved)&checker.ObjectFlagsClassOrInterface == 0 {
		return false
	}
	// A subclass of Promise is an instance of Promise.
	for _, base := range checker.Checker_getBaseTypes(typeChecker, resolved) {
		if isPromiseInstanceType(sourceProgram, typeChecker, base, depth+1) {
			return true
		}
	}
	return false
}

// isPromiseProducingExpression reports whether the expression can only evaluate
// to a promise, judged by how it is written.
func isPromiseProducingExpression(ctx rule.RuleContext, node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindNewExpression:
		callee := utils.SkipAssertionsAndParens(node.AsNewExpression().Expression)
		return isUnshadowedPromiseIdentifier(callee)
	case ast.KindCallExpression:
	default:
		return false
	}

	call := node.AsCallExpression()
	// A dynamic import always yields a promise, whatever it names.
	if call.Expression != nil && call.Expression.Kind == ast.KindImportKeyword {
		return true
	}
	if call.QuestionDotToken != nil {
		return false
	}

	// `rs.importActual` / `rs.importMock` load a module asynchronously. The
	// shapes the build rewrites differ per member, and `importActual` honours
	// a local declaration of the receiver, so both questions go through the
	// shared table rather than being assumed here.
	if utility := rstestUtils.ParseRstestPluginManagedCall(node); utility != nil {
		if !asyncModuleLoaders[utility.Member] {
			return false
		}
		return !utility.API.ResolvesReceiver ||
			!rstestUtils.ReceiverIsLocallyDeclared(ctx, utility.NamespaceNode)
	}

	callee := utils.SkipAssertionsAndParens(call.Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil {
		return false
	}
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || !promiseStatics[name.AsIdentifier().Text] {
		return false
	}
	return isUnshadowedPromiseIdentifier(utils.SkipAssertionsAndParens(access.Expression))
}

// isUnshadowedPromiseIdentifier reports whether node is the global `Promise`.
// A file that declares its own `Promise` is left to the type layer.
func isUnshadowedPromiseIdentifier(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindIdentifier &&
		node.AsIdentifier().Text == "Promise" &&
		!utils.IsShadowed(node, "Promise")
}

func isFunctionExpressionLike(node *ast.Node) bool {
	return node != nil &&
		(node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionExpression)
}

// returnedExpressions collects every expression the function can hand back. The
// second result is false when the body is missing or holds a bare `return;`,
// which makes the set incomplete and the function undecidable here.
func returnedExpressions(fn *ast.Node) ([]*ast.Node, bool) {
	body := fn.Body()
	if body == nil {
		return nil, false
	}
	if body.Kind != ast.KindBlock {
		return []*ast.Node{body}, true
	}

	returned := make([]*ast.Node, 0, 2)
	complete := true
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		// A nested function's returns belong to that function.
		if ast.IsFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			expression := node.AsReturnStatement().Expression
			if expression == nil {
				complete = false
				return false
			}
			returned = append(returned, expression)
			return false
		}
		return node.ForEachChild(walk)
	}
	body.ForEachChild(walk)
	return returned, complete
}

// suggestions builds the two narrow rewrites that are safe to apply on their
// own. Neither is offered as a fix: the common shape awaits the real module
// inside the factory, and turning that into a top-level
// `import … with { rstest: 'importActual' }` moves code across statements.
func suggestions(ctx rule.RuleContext, factory *ast.Node) []rule.RuleSuggestion {
	if !isFunctionExpressionLike(factory) {
		return nil
	}
	// Both rewrites take the promise out of the value while leaving the
	// factory's own return type as written, so an annotated
	// `(): Promise<T> => …` would be left annotating a `T`. Rewriting the
	// annotation as well is a second edit with its own shapes to get right —
	// `PromiseLike<T>`, a union, an alias — so an annotated factory is simply
	// left alone.
	if factory.Type() != nil {
		return nil
	}
	returned, complete := returnedExpressions(factory)
	if !complete {
		return nil
	}

	if ast.GetFunctionFlags(factory)&ast.FunctionFlagsAsync != 0 {
		return removeAsyncSuggestion(ctx, factory, returned)
	}
	return unwrapPromiseResolveSuggestion(ctx, factory, returned)
}

// removeAsyncSuggestion drops the `async` keyword, which is equivalent only
// when the body neither awaits nor already hands back a promise of its own.
func removeAsyncSuggestion(ctx rule.RuleContext, factory *ast.Node, returned []*ast.Node) []rule.RuleSuggestion {
	for _, expression := range returned {
		if !isDefinitelyNotPromise(expression) {
			return nil
		}
	}
	if bodyAwaits(factory) {
		return nil
	}

	modifiers := factory.Modifiers()
	if modifiers == nil {
		return nil
	}
	for _, modifier := range modifiers.Nodes {
		if modifier == nil || modifier.Kind != ast.KindAsyncKeyword {
			continue
		}
		keyword := utils.TrimNodeTextRange(ctx.SourceFile, modifier)
		text := ctx.SourceFile.Text()
		removal := core.NewTextRange(
			keyword.Pos(),
			ecmascript.SkipLeadingWhitespace(text, keyword.End(), len(text)),
		)
		if utils.HasCommentInSpan(ctx.Comments.All(), removal.Pos(), removal.End()) {
			return nil
		}
		return []rule.RuleSuggestion{{
			Message:  removeAsyncMessage,
			FixesArr: []rule.RuleFix{rule.RuleFixRemoveRange(removal)},
		}}
	}
	return nil
}

// unwrapPromiseResolveSuggestion replaces a lone `Promise.resolve(x)` with `x`.
func unwrapPromiseResolveSuggestion(ctx rule.RuleContext, factory *ast.Node, returned []*ast.Node) []rule.RuleSuggestion {
	if len(returned) != 1 {
		return nil
	}
	wrapper := utils.SkipAssertionsAndParens(returned[0])
	if wrapper == nil || wrapper.Kind != ast.KindCallExpression {
		return nil
	}
	call := wrapper.AsCallExpression()
	callee := utils.SkipAssertionsAndParens(call.Expression)
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return nil
	}
	access := callee.AsPropertyAccessExpression()
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.AsIdentifier().Text != "resolve" ||
		!isUnshadowedPromiseIdentifier(utils.SkipAssertionsAndParens(access.Expression)) {
		return nil
	}
	if call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return nil
	}
	resolved := call.Arguments.Nodes[0]
	if resolved == nil || resolved.Kind == ast.KindSpreadElement || !isDefinitelyNotPromise(resolved) {
		return nil
	}

	// The whole call is replaced by the resolved value as written, so anything
	// in the space that goes — a comment around `Promise.resolve` or around the
	// argument — would be deleted with it.
	wrapperRange := utils.TrimNodeTextRange(ctx.SourceFile, wrapper)
	resolvedRange := utils.TrimNodeTextRange(ctx.SourceFile, resolved)
	comments := ctx.Comments.All()
	if utils.HasCommentInSpan(comments, wrapperRange.Pos(), resolvedRange.Pos()) ||
		utils.HasCommentInSpan(comments, resolvedRange.End(), wrapperRange.End()) {
		return nil
	}

	replacement := ctx.SourceFile.Text()[resolvedRange.Pos():resolvedRange.End()]
	// A concise arrow body starting with `{` would parse as a block.
	if factory.Body() == returned[0] && len(replacement) > 0 && replacement[0] == '{' {
		replacement = "(" + replacement + ")"
	}
	return []rule.RuleSuggestion{{
		Message:  unwrapPromiseResolveMessage,
		FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(wrapperRange, replacement)},
	}}
}

// isDefinitelyNotPromise reports whether the expression's value can be read off
// the syntax and is certainly not a promise. It gates the suggestions, so
// anything it cannot vouch for is a "no".
func isDefinitelyNotPromise(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindObjectLiteralExpression,
		ast.KindArrayLiteralExpression,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindTemplateExpression,
		ast.KindNumericLiteral,
		ast.KindBigIntLiteral,
		ast.KindRegularExpressionLiteral,
		ast.KindTrueKeyword,
		ast.KindFalseKeyword,
		ast.KindNullKeyword,
		ast.KindArrowFunction,
		ast.KindFunctionExpression,
		ast.KindClassExpression:
		return true
	case ast.KindIdentifier:
		return node.AsIdentifier().Text == "undefined" && !utils.IsShadowed(node, "undefined")
	}
	return false
}

// bodyAwaits reports whether the function's own body suspends: an `await`
// expression, `for await`, or an `await using` declaration. A nested function's
// body does not count — but everything else written on that nested function
// does: a computed member name and a decorator are evaluated where the function
// appears, which is this body.
func bodyAwaits(fn *ast.Node) bool {
	body := fn.Body()
	if body == nil {
		return false
	}

	found := false
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil || found {
			return found
		}
		if ast.IsFunctionLike(node) {
			nestedBody := node.Body()
			return node.ForEachChild(func(child *ast.Node) bool {
				if child == nestedBody {
					return false
				}
				return walk(child)
			})
		}
		switch node.Kind {
		case ast.KindAwaitExpression:
			found = true
			return true
		case ast.KindForOfStatement:
			if node.AsForInOrOfStatement().AwaitModifier != nil {
				found = true
				return true
			}
		case ast.KindVariableDeclarationList:
			if ast.IsVarAwaitUsing(node) {
				found = true
				return true
			}
		}
		return node.ForEachChild(walk)
	}
	body.ForEachChild(walk)
	return found
}
