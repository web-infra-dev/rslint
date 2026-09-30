package require_hook

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/require_hook"
)

// executionTimeAPIs are the two Rstest APIs that register work against the
// test that is currently running. They are not lifecycle hooks: calling either
// one while the file is collected throws
// "onTestFinished() can only be called inside a test"
// (packages/core/src/runtime/runner/runner.ts), so telling the user to move
// the call into a hook would name a repair that also fails.
var executionTimeAPIs = map[string]bool{
	"onTestFinished": true,
	"onTestFailed":   true,
}

// executionTimeAPIName returns the execution-time API a call reaches, or "".
//
// The receiver is resolved rather than read as written, so a renamed import
// and a namespace import are both recognized, while a local function that
// happens to be called `onTestFinished` is not.
func executionTimeAPIName(ctx rule.RuleContext, node *ast.Node) string {
	callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
	if callee == nil {
		return ""
	}

	if callee.Kind == ast.KindIdentifier {
		name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbolModules(
			callee.AsIdentifier().Text,
			callee,
			ctx.Refs.Resolve(callee),
			ctx.SourceFile,
			rstestUtils.RstestCoreImportModules,
		)
		if executionTimeAPIs[name] {
			return name
		}
		return ""
	}

	member, ok := internalUtils.AccessExpressionStaticName(callee)
	if !ok || !executionTimeAPIs[member] {
		return ""
	}
	namespace := internalUtils.SkipAssertionsAndParens(callee.Expression())
	if namespace == nil {
		return ""
	}
	if rstestUtils.IsImportMetaRstest(namespace) {
		return member
	}
	if namespace.Kind != ast.KindIdentifier {
		return ""
	}
	if testFramework.IsModuleNamespaceSymbolModules(
		ctx.Refs.Resolve(namespace),
		rstestUtils.RstestCoreImportModules,
	) {
		return member
	}
	return ""
}

// isRootedInUtilitiesObject reports whether a callee chain is read off Rstest's
// utilities object, however many members deep: `rs.mock('./m')`,
// `rstest.spyOn(o, 'm').mockReturnValue(1)`, `import.meta.rstest.rs.fn()`.
//
// This is where the Rstest port deliberately parts company with
// @vitest/eslint-plugin, which spells the same exemption as
// `getNodeName(node)?.startsWith('vi')` — with no separator, so `video.play()`,
// `viewport.set()` and `visit('/home')` are all silently treated as framework
// calls. eslint-plugin-jest has the separator (`'jest.'`) and does not have the
// bug. Reproducing it under Rstest's `rs` prefix would swallow `response.*`,
// `result.*`, `resetDatabase()` and most other setup spellings, so the receiver
// is resolved instead of pattern-matched.
//
// Resolving also buys the spellings a prefix test cannot reach: a renamed
// import, a namespace import and `import.meta.rstest`.
func isRootedInUtilitiesObject(ctx rule.RuleContext, expr *ast.Node) bool {
	for {
		expr = internalUtils.SkipAssertionsAndParens(expr)
		if expr == nil {
			return false
		}
		switch expr.Kind {
		case ast.KindCallExpression:
			expr = expr.AsCallExpression().Expression
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			receiver := expr.Expression()
			if rstestUtils.IsUtilitiesObject(ctx, receiver) {
				return true
			}
			expr = receiver
		default:
			return false
		}
	}
}

// isRstestFnCall reports whether a call is part of Rstest's own API surface,
// which is what the rule takes as "not setup code that belongs in a hook".
//
// The four sources are separate because the plugin models them separately:
// registrations and hooks come from the shared call analysis, expect chains
// have their own parser, and the utilities object is reached either as the
// build rewrites it (by the receiver as written) or through ordinary bindings.
//
// They are asked in increasing cost: a candidate-gated parse, then pure
// syntax, then a receiver resolution, and only last the expect parser, whose
// candidate set is completed by indexing every callback in the file.
func isRstestFnCall(
	node *ast.Node,
	ctx rule.RuleContext,
	analysis *rstestUtils.RstestCallAnalysis,
) bool {
	if analysis.ParseFnCall(node) != nil {
		return true
	}
	// A plugin-managed member is matched by the receiver as written, because
	// the build rewrites it that way even in a file that declares its own `rs`.
	if rstestUtils.ParseRstestPluginManagedCall(node) != nil {
		return true
	}
	if isRootedInUtilitiesObject(ctx, node.AsCallExpression().Expression) {
		return true
	}
	return analysis.IsExpectCall(node)
}

var RequireHookRule = shared.NewRule(shared.Config{
	Name: "rstest/require-hook",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			IsFrameworkCall: func(call *ast.Node) bool {
				return isRstestFnCall(call, ctx, analysis)
			},
			IsDescribe: func(call *ast.Node) bool {
				parsed := analysis.ParseFnCall(call)
				return parsed != nil && parsed.Kind == rstestUtils.RstestFnTypeDescribe
			},
			ExecutionTimeAPI: func(call *ast.Node) string {
				return executionTimeAPIName(ctx, call)
			},
		}
	},
})
