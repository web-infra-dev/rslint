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
		scan := &shadowScan{sourceFile: ctx.SourceFile}
		// Both listeners resolve registrations through the same filter so a
		// shadowed `describe` cannot unbalance the suite depth counter.
		registration := func(node *ast.Node) *rstestUtils.ParsedRstestFnCall {
			parsed := analysis.ParseFnCall(node)
			if parsed == nil || scan.isShadowedRegistration(ctx, parsed) {
				return nil
			}
			return parsed
		}
		depth := 0
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := registration(node)
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
				if parsed := registration(node); parsed != nil && parsed.Kind == rstestUtils.RstestFnTypeDescribe {
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
	// A `let`/`var` CommonJS binding can be reassigned after its declaration
	// (`it = it.extend({ account: {} })`), so the destructuring pattern no
	// longer proves which API the call receives. Renaming it to the other base
	// API would silently drop the fixtures.
	if isMutableRequireBinding(symbol) {
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
		// A class static block is a variable environment of its own, and is not
		// function-like, so a `var` nested inside it is reported by neither the
		// walk above nor `IsNameShadowedBetween`.
		if ancestor.Kind == ast.KindClassStaticBlockDeclaration {
			if body := ancestor.AsClassStaticBlockDeclaration().Body; body != nil &&
				utils.HasHoistedVarDeclaration(body, preferred) {
				return nil
			}
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

// isMutableRequireBinding reports whether symbol is bound by a destructured
// `require` declaration that is not `const`. ESM import bindings cannot be
// reassigned, so only CommonJS reaches this shape.
func isMutableRequireBinding(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindBindingElement {
			continue
		}
		if declarationList := utils.GetDeclListForSymbolDecl(declaration); declarationList != nil &&
			!ast.IsVarConst(declarationList) {
			return true
		}
	}
	return false
}

// shadowScan filters registrations whose root is really a local binding.
//
// The source-only binder does not resolve every value-producing TypeScript
// declaration (notably namespaces), and can resolve through a declaration in a
// nearer module block to an outer import, so the shared parser accepts a few
// shadowed spellings as Rstest APIs. Correcting that inside the parser would
// put a full scope walk on the path of every registration in every file, and
// so on every rule that uses the parser; this rule pays for its own strictness
// instead.
type shadowScan struct {
	sourceFile *ast.SourceFile
	names      map[string]bool
}

// declaresName reports whether the file declares name anywhere, from a single
// indexing pass. Every binding has a declaration name, so this is a superset of
// what the scope walk can find: a miss rules the walk out, and a hit only costs
// the walk. Test files rarely declare `test`, `it`, or `describe` themselves,
// which keeps the walk off the common path entirely.
func (scan *shadowScan) declaresName(name string) bool {
	if scan.names == nil {
		scan.names = map[string]bool{}
		if scan.sourceFile != nil {
			var visit func(*ast.Node)
			visit = func(node *ast.Node) {
				if node.Kind == ast.KindIdentifier && ast.IsDeclarationName(node) {
					scan.names[node.Text()] = true
				}
				node.ForEachChild(func(child *ast.Node) bool {
					visit(child)
					return false
				})
			}
			visit(scan.sourceFile.AsNode())
		}
	}
	return scan.names[name]
}

func (scan *shadowScan) isShadowedRegistration(ctx rule.RuleContext, parsed *rstestUtils.ParsedRstestFnCall) bool {
	root := parsed.Head.Local.Node
	if root == nil || root.Kind != ast.KindIdentifier {
		return false
	}
	name := root.Text()
	symbol := ctx.Refs.Resolve(root)
	if isShadowedByModuleBlock(root, ctx.SourceFile, name, symbol) {
		return true
	}
	return symbol == nil && scan.declaresName(name) && utils.IsShadowed(root, name)
}

// isShadowedByModuleBlock reports whether a `namespace`/`module` body between
// the call and the file declares name. A resolved symbol is trusted only when
// its declaration lives inside that body.
func isShadowedByModuleBlock(root *ast.Node, sourceFile *ast.SourceFile, name string, symbol *ast.Symbol) bool {
	if sourceFile == nil {
		return false
	}
	for ancestor := root.Parent; ancestor != nil && ancestor != sourceFile.AsNode(); ancestor = ancestor.Parent {
		if ancestor.Kind != ast.KindModuleBlock {
			continue
		}
		statements := ancestor.AsModuleBlock().Statements
		if statements == nil || !utils.HasLocalDeclarationInStatements(statements.Nodes, name) {
			continue
		}
		if symbol == nil {
			return true
		}
		for _, declaration := range symbol.Declarations {
			if ast.IsNodeDescendantOf(declaration, ancestor) {
				return false
			}
		}
		return true
	}
	return false
}
