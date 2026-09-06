package require_to_throw_message

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
	shared "github.com/web-infra-dev/rslint/internal/utils/test_framework/rules/require_to_throw_message"
)

var RequireToThrowMessageRule = shared.NewRule(shared.Config{
	Name: "rstest/require-to-throw-message",
	Prepare: func(ctx rule.RuleContext) shared.Runtime {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return shared.Runtime{
			ParseExpectCall: func(node *ast.Node) *shared.ExpectCall {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil ||
					parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
					parsed.Head == nil ||
					parsed.MatcherEntry == nil ||
					len(parsed.Matchers) == 0 ||
					parsed.Matchers[0].Kind != rstestUtils.RstestExpectMatcherCall {
					return nil
				}
				if isSourceOnlyLocalExpect(node, parsed, ctx) {
					return nil
				}

				matcherCall := rstestUtils.MatcherCall(parsed.MatcherEntry)
				if matcherCall == nil {
					return nil
				}

				return &shared.ExpectCall{
					Matcher:      parsed.Matcher,
					MatcherEntry: parsed.MatcherEntry,
					Modifiers:    parsed.Modifiers,
					MatcherArgs:  matcherCall.Arguments(),
				}
			},
		}
	},
})

// isSourceOnlyLocalExpect closes the one provenance gap in the shared Rstest
// parser that matters to this rule. Without a checker the parser deliberately
// treats a bare `expect` as the global, but the binder can still prove that a
// call resolves to a local value. Rstest imports and import.meta destructuring
// remain accepted; this only rejects an actual local shadow.
func isSourceOnlyLocalExpect(
	node *ast.Node,
	parsed *rstestUtils.ParsedRstestExpectCall,
	ctx rule.RuleContext,
) bool {
	if ctx.TypeChecker != nil || ctx.Refs == nil || parsed.FromTestContext {
		return false
	}
	root := testFramework.ResolveFirstIdentifier(node.AsCallExpression().Expression)
	if root == nil || root.Kind != ast.KindIdentifier {
		return false
	}
	symbol := ctx.Refs.Resolve(root)
	if symbol == nil {
		return false
	}
	name, _, _ := testFramework.ResolveFunctionIdentifierReferenceFromSymbolModules(
		root.AsIdentifier().Text,
		root,
		symbol,
		ctx.SourceFile,
		rstestUtils.RstestAllImportModules,
	)
	if name == "expect" {
		return false
	}
	if testFramework.IsModuleNamespaceSymbolModules(symbol, rstestUtils.RstestAllImportModules) {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		variable := declaration
		if declaration.Kind == ast.KindBindingElement {
			variable = internalUtils.EnclosingVariableDeclarationOfBindingElement(declaration)
		}
		if variable != nil &&
			variable.Kind == ast.KindVariableDeclaration &&
			rstestUtils.IsImportMetaRstest(variable.AsVariableDeclaration().Initializer) {
			return false
		}
	}
	return internalUtils.IsRuntimeValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
}
