package no_async_promise_executor

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/binder"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

var asyncMessage = rule.RuleMessage{
	Id:          "async",
	Description: "Promise executor functions should not be async.",
}

func sourceMayUsePromise(sourceFile *ast.SourceFile) bool {
	// Parsed identifier names are normalized, including Unicode escapes. Every
	// async executor also spells the contextual keyword contiguously; matches in
	// comments and strings are harmless false positives. Stay conservative for
	// direct callers that do not provide parser metadata.
	if sourceFile == nil || sourceFile.Identifiers == nil {
		return true
	}
	_, ok := sourceFile.Identifiers["Promise"]
	return ok && strings.Contains(sourceFile.Text(), "async")
}

func enclosingPromiseScope(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsLocalsContainer(current) {
			return current
		}
	}
	return nil
}

func sourceHasPromiseBinding(ctx rule.RuleContext) bool {
	if ctx.SourceFile == nil || !ctx.SourceFile.IsBound() {
		return true
	}

	// Binder keeps every locals container in one source-ordered chain. A full
	// miss proves all Promise candidates in the file are global, so subsequent
	// scope changes need no individual resolution.
	for container := ctx.SourceFile.AsNode(); container != nil; {
		data := container.LocalsContainerData()
		if data == nil {
			return true
		}
		if len(data.Locals) != 0 && data.Locals["Promise"] != nil {
			return true
		}
		if (container.Kind == ast.KindFunctionExpression || container.Kind == ast.KindClassExpression) &&
			container.Name() != nil && container.Name().Text() == "Promise" {
			return true
		}
		container = data.NextContainer
	}
	return false
}

func isGlobalPromiseReference(ctx rule.RuleContext, identifier *ast.Node) bool {
	if ctx.Program != nil && ctx.SourceFile != nil && ctx.SourceFile.IsBound() {
		if ast.IsGlobalSourceFile(ctx.SourceFile.AsNode()) && ctx.SourceFile.Locals["Promise"] != nil {
			// ESLint stores user declarations and configured globals in one
			// global-scope variable. Any source definition, including a type-only
			// one, makes isGlobalReference return false for that name.
			return false
		}
		// RefStore.Resolve falls back to the TypeChecker for globals. This rule
		// only needs to disprove a global by finding a local binding, so use the
		// same binder scope walk without paying for that global fallback.
		resolver := binder.NameResolver{CompilerOptions: ctx.Program.Options()}
		if ast.IsGlobalSourceFile(ctx.SourceFile.AsNode()) {
			resolver.Globals = ctx.SourceFile.Locals
		}
		return resolver.Resolve(
			identifier,
			"Promise",
			ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias,
			nil,
			true,  // isUse
			false, // excludeGlobals
		) == nil
	}
	return !utils.IsShadowed(identifier, "Promise")
}

type noAsyncPromiseExecutorState struct {
	ctx                   rule.RuleContext
	cachedScope           *ast.Node
	cachedGlobal          bool
	hasCachedScope        bool
	checkedFileBindings   bool
	fileHasPromiseBinding bool
}

func (state *noAsyncPromiseExecutorState) check(node *ast.Node) {
	newExpression := node.AsNewExpression()
	callee := newExpression.Expression
	if callee != nil && callee.Kind == ast.KindParenthesizedExpression {
		callee = ast.SkipParentheses(callee)
	}
	if callee == nil || callee.Kind != ast.KindIdentifier ||
		callee.AsIdentifier().Text != "Promise" {
		return
	}

	arguments := newExpression.Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return
	}

	executor := arguments.Nodes[0]
	if executor == nil {
		return
	}
	if executor.Kind == ast.KindParenthesizedExpression {
		executor = ast.SkipParentheses(executor)
	}
	if executor == nil ||
		(executor.Kind != ast.KindFunctionExpression && executor.Kind != ast.KindArrowFunction) {
		return
	}

	modifiers := executor.Modifiers()
	if modifiers == nil || modifiers.ModifierFlags&ast.ModifierFlagsAsync == 0 {
		return
	}

	scope := enclosingPromiseScope(callee)
	if !state.hasCachedScope || state.cachedScope != scope {
		state.cachedScope = scope
		if state.hasCachedScope && !state.checkedFileBindings {
			state.fileHasPromiseBinding = sourceHasPromiseBinding(state.ctx)
			state.checkedFileBindings = true
		}
		if state.checkedFileBindings && !state.fileHasPromiseBinding {
			state.cachedGlobal = true
		} else {
			state.cachedGlobal = isGlobalPromiseReference(state.ctx, callee)
		}
		state.hasCachedScope = true
	}
	if !state.cachedGlobal {
		return
	}

	for _, modifier := range modifiers.Nodes {
		if modifier != nil && modifier.Kind == ast.KindAsyncKeyword {
			// The modifier end is already the exact token end. Skip only its
			// leading trivia so dense reports do not allocate a scanner per token.
			state.ctx.ReportRange(core.NewTextRange(
				scanner.SkipTrivia(state.ctx.SourceFile.Text(), modifier.Pos()),
				modifier.End(),
			), asyncMessage)
			return
		}
	}
}

func noAsyncPromiseExecutorListeners(ctx rule.RuleContext) rule.RuleListeners {
	state := &noAsyncPromiseExecutorState{ctx: ctx}
	return rule.RuleListeners{ast.KindNewExpression: state.check}
}

// NoAsyncPromiseExecutorRule disallows using an async function as a Promise executor.
var NoAsyncPromiseExecutorRule = rule.Rule{
	Name:   "no-async-promise-executor",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		if !sourceMayUsePromise(ctx.SourceFile) {
			return nil
		}
		if !ctx.Globals.Access("Promise").IsDeclared() {
			return nil
		}
		return noAsyncPromiseExecutorListeners(ctx)
	},
}
