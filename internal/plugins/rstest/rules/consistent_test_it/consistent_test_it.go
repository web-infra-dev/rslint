package consistent_test_it

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed consistent_test_it.schema.json
var schemaJSON []byte

func parseOptions(options []any) (outside, inside string) {
	outside, inside = "test", "it"
	if len(options) == 0 {
		return
	}
	config, _ := options[0].(map[string]any)
	if value, ok := config["fn"].(string); ok && (value == "test" || value == "it") {
		outside, inside = value, value
	}
	if value, ok := config["withinDescribe"].(string); ok && (value == "test" || value == "it") {
		inside = value
	}
	return
}

// ConsistentTestItRule follows the option and call-diagnostic contract of
// eslint-plugin-jest v29.16.1 and @vitest/eslint-plugin v1.6.27. Imports are
// dependencies of call fixes, not standalone violations: unused imports and
// references outside registrations do not express a test naming convention.
var ConsistentTestItRule = rule.Rule{
	Name:   "rstest/consistent-test-it",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		outside, inside := parseOptions(options)
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		depth := 0
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil {
					return
				}
				if parsed.Kind == rstestUtils.RstestFnTypeDescribe {
					depth++
					return
				}
				if parsed.Kind != rstestUtils.RstestFnTypeTest || parsed.IsPlaywright {
					return
				}
				preferred := outside
				id, key, suffix := "consistentMethod", "testKeyword", ""
				if depth > 0 {
					preferred = inside
					id, key, suffix = "consistentMethodWithinDescribe", "testKeywordWithinDescribe", " within describe"
				}
				actual := parsed.Name
				root, original := parsed.Head.Local.Node, parsed.Head.Original.Node
				if root != nil && (parsed.LocalName == "test" || parsed.LocalName == "it") &&
					(original == root || original == nil || original.Pos() < node.Pos() || original.End() > node.End()) {
					// Fixture idioms such as `const it = test.extend(...)` already
					// choose a spelling. Do not report a no-op rename (Vitest #956).
					actual = parsed.LocalName
				}
				if actual == preferred {
					return
				}
				message := rule.RuleMessage{
					Id:          id,
					Description: fmt.Sprintf("Prefer using '%s' instead of '%s'%s", preferred, actual, suffix),
					Data:        map[string]string{key: preferred, "oppositeTestKeyword": actual},
				}
				ctx.ReportNodeWithDeferredFixes(node.AsCallExpression().Expression, message, func() []rule.RuleFix {
					return callFixes(ctx, node, parsed, preferred)
				})
			},
			rule.ListenerOnExit(ast.KindCallExpression): func(node *ast.Node) {
				if parsed := analysis.ParseFnCall(node); parsed != nil && parsed.Kind == rstestUtils.RstestFnTypeDescribe {
					depth--
				}
			},
		}
	},
}

func callFixes(ctx rule.RuleContext, node *ast.Node, parsed *rstestUtils.ParsedRstestFnCall, preferred string) []rule.RuleFix {
	root, original := parsed.Head.Local.Node, parsed.Head.Original.Node
	// Namespace accessors are at this call site. An alias's original node is
	// in its declaration; rewriting it would affect other registrations too.
	if original != nil && original != root && original.Pos() >= node.Pos() && original.End() <= node.End() {
		if root.Kind == ast.KindIdentifier {
			symbol := ctx.Refs.Resolve(root)
			if symbol == nil || len(symbol.Declarations) != 1 || symbol.Declarations[0].Kind != ast.KindNamespaceImport {
				// CommonJS namespace objects and local aliases can be mutated.
				return nil
			}
		}
		rangeToReplace, text, ok := testFramework.AccessorReplacement(ctx.SourceFile, original, preferred)
		if ok {
			return []rule.RuleFix{rule.RuleFixReplaceRange(rangeToReplace, text)}
		}
		return nil
	}
	if root == nil || root.Kind != ast.KindIdentifier || root.Text() != parsed.Name {
		return nil
	}
	// Only a direct import/global can be replaced by the other base API.
	// `const test = base.extend(...)` must retain its fixtures.
	symbol := ctx.Refs.Resolve(root)
	name, _, mode := testFramework.ResolveFunctionIdentifierReferenceFromSymbolModules(
		root.Text(), root, symbol, ctx.SourceFile, rstestUtils.RstestCoreImportModules,
	)
	if name != parsed.Name {
		return nil
	}
	fix := rule.RuleFixReplace(ctx.SourceFile, root, preferred)
	if utils.IsNameShadowedBetween(root, ctx.SourceFile.AsNode(), preferred) {
		return nil
	}
	for ancestor := root.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if ancestor.Kind == ast.KindWithStatement {
			return nil
		}
		if ancestor.Kind == ast.KindModuleBlock &&
			(utils.HasLocalDeclarationInStatements(ancestor.AsModuleBlock().Statements.Nodes, preferred) ||
				utils.HasHoistedVarDeclaration(ancestor, preferred)) {
			return nil
		}
	}
	// Reuse a plain, value import. Aliased imports do not bind the spelling
	// being introduced, even if they import the preferred export.
	for _, statement := range ctx.SourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if !rstestUtils.IsRstestCoreImportModule(declaration.ModuleSpecifier.Text()) {
			continue
		}
		for _, element := range rstestUtils.NamedImportElements(declaration) {
			if rstestUtils.ImportedSpecifierName(element) == preferred && element.Name().Text() == preferred {
				return []rule.RuleFix{fix}
			}
		}
	}
	if utils.IsShadowed(root, preferred) {
		return nil
	}
	if mode == rstestUtils.RSTEST_GLOBAL_MODE {
		return []rule.RuleFix{fix}
	}
	for _, declaration := range symbol.Declarations {
		if declaration.Kind != ast.KindImportSpecifier || ast.IsTypeOnlyImportDeclaration(declaration) {
			continue
		}
		// Keep the original import for exports, fixture factories, and calls
		// governed by the other option. Insertion preserves comments/trivia.
		return []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, declaration, preferred+", "), fix}
	}
	return nil
}
