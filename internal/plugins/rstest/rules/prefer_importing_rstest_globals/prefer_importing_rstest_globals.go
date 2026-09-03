package prefer_importing_rstest_globals

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

const defaultModule = "@rstest/core"

type nameSet struct {
	seen  map[string]struct{}
	order []string
}

func newNameSet(capacity int) *nameSet {
	return &nameSet{seen: make(map[string]struct{}, capacity)}
}

func (set *nameSet) add(name string) {
	if name == "" {
		return
	}
	if _, exists := set.seen[name]; exists {
		return
	}
	set.seen[name] = struct{}{}
	set.order = append(set.order, name)
}

func (set *nameSet) joined() string { return strings.Join(set.order, ", ") }

func (set *nameSet) sortedJoined() string {
	names := slices.Clone(set.order)
	slices.SortFunc(names, ecmascript.CompareStrings)
	return strings.Join(names, ", ")
}

func preferMessage(names string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferImportingRstestGlobals",
		Description: "Import `" + names + "` from `@rstest/core`.",
		Data:        map[string]string{"names": names},
	}
}

func shouldReportIdentifier(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier || !rstestUtils.IsRstestGlobal(node.AsIdentifier().Text) ||
		ast.IsIdentifierName(node) || ast.IsDeclarationNameOrImportPropertyName(node) {
		return false
	}
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport,
			ast.KindExportSpecifier, ast.KindExportDeclaration:
			return false
		}
		if parent.Kind == ast.KindSourceFile || parent.Kind == ast.KindBlock || ast.IsExpression(parent) {
			break
		}
	}
	if ctx.Refs == nil {
		return true
	}
	symbol := ctx.Refs.Resolve(node)
	if symbol == nil {
		return true
	}
	name := node.AsIdentifier().Text
	resolved, _, mode := testFramework.ResolveFunctionIdentifierReferenceFromSymbolModules(
		name, node, symbol, ctx.SourceFile, rstestUtils.RstestCoreImportModules,
	)
	if mode == testFramework.ReferenceModeImport && resolved != "" {
		return false
	}
	return !internalUtils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile)
}

func createImport(isModule bool, module, names string) string {
	if isModule {
		return "import { " + names + " } from '" + module + "';"
	}
	return "const { " + names + " } = require('" + module + "');"
}

// isUseStrictDirective reports whether the statement is a `use strict`
// directive. Only a bare string-literal expression statement belongs to the
// directive prologue, so a template literal or a parenthesized string is an
// ordinary statement and must not shift the insertion point.
func isUseStrictDirective(statement *ast.Node) bool {
	if statement == nil || statement.Kind != ast.KindExpressionStatement {
		return false
	}
	expression := statement.AsExpressionStatement().Expression
	if expression == nil || expression.Kind != ast.KindStringLiteral {
		return false
	}
	return expression.AsStringLiteral().Text == "use strict"
}

// mergeIntoImport appends the missing names to an existing rstest import
// instead of rewriting it. Regenerating the declaration would drop everything
// the generated text cannot express, such as a default binding or a comment
// between the specifiers, so the fix only inserts after the last specifier.
func mergeIntoImport(declaration *ast.ImportDeclaration, names *nameSet) []rule.RuleFix {
	elements := rstestUtils.NamedImportElements(declaration)
	if len(elements) == 0 {
		return nil
	}
	present := make(map[string]struct{}, len(elements))
	for _, element := range elements {
		specifier := element.AsImportSpecifier()
		if specifier != nil && specifier.Name() != nil {
			present[specifier.Name().Text()] = struct{}{}
		}
	}
	additions := newNameSet(len(names.order))
	for _, name := range names.order {
		if _, exists := present[name]; !exists {
			additions.add(name)
		}
	}
	if len(additions.order) == 0 {
		return nil
	}
	last := elements[len(elements)-1]
	return []rule.RuleFix{rule.RuleFixInsertAfter(last, ", "+additions.sortedJoined())}
}

func collectRequireNames(binding *ast.Node, names *nameSet) {
	if binding == nil || binding.Kind != ast.KindObjectBindingPattern {
		return
	}
	pattern := binding.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil {
		return
	}
	for _, element := range pattern.Elements.Nodes {
		be := element.AsBindingElement()
		if be == nil || be.DotDotDotToken != nil || be.Initializer != nil {
			continue
		}
		imported := rstestUtils.RequireBindingImportedName(element)
		if imported == "" {
			continue
		}
		local := ""
		if be.Name() != nil && be.Name().Kind == ast.KindIdentifier {
			local = be.Name().AsIdentifier().Text
		}
		if local != "" && local != imported {
			imported += ": " + local
		}
		names.add(imported)
	}
}

func findImport(statements []*ast.Node) *ast.ImportDeclaration {
	for _, statement := range statements {
		if statement == nil || statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if declaration == nil || declaration.ModuleSpecifier == nil || !rstestUtils.IsRstestCoreImportModule(declaration.ModuleSpecifier.Text()) {
			continue
		}
		if len(rstestUtils.NamedImportElements(declaration)) > 0 {
			return declaration
		}
	}
	return nil
}

type existingRequire struct {
	node        *ast.Node
	declaration *ast.VariableDeclaration
	module      string
}

func findRequire(statements []*ast.Node) *existingRequire {
	for _, statement := range statements {
		if statement == nil || statement.Kind != ast.KindVariableStatement {
			continue
		}
		list := statement.AsVariableStatement().DeclarationList.AsVariableDeclarationList()
		if list == nil || list.Declarations == nil || len(list.Declarations.Nodes) != 1 {
			continue
		}
		declaration := list.Declarations.Nodes[0].AsVariableDeclaration()
		if declaration == nil || declaration.Name() == nil || declaration.Name().Kind != ast.KindObjectBindingPattern {
			continue
		}
		module, ok := rstestUtils.RstestCoreModuleFromRequireCall(declaration.Initializer)
		if ok {
			return &existingRequire{node: statement, declaration: declaration, module: module}
		}
	}
	return nil
}

func lineIndent(ctx rule.RuleContext, node *ast.Node) string {
	if ctx.SourceFile == nil || node == nil {
		return ""
	}
	text := ctx.SourceFile.Text()
	position := internalUtils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
	lineStart := strings.LastIndex(text[:position], "\n") + 1
	indent := text[lineStart:position]
	if strings.Trim(indent, " \t") != "" {
		return ""
	}
	return indent
}

func buildAutofix(ctx rule.RuleContext, collected []string) []rule.RuleFix {
	names := newNameSet(len(collected))
	for _, name := range collected {
		names.add(name)
	}
	isModule := ctx.LanguageOptions.EffectiveSourceType() == "module"
	if ctx.SourceFile == nil || ctx.SourceFile.Statements == nil || len(ctx.SourceFile.Statements.Nodes) == 0 {
		return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(0, 0), createImport(isModule, defaultModule, names.sortedJoined()))}
	}
	statements := ctx.SourceFile.Statements.Nodes
	first := statements[0]
	if isUseStrictDirective(first) {
		return []rule.RuleFix{rule.RuleFixInsertAfter(first, "\n"+lineIndent(ctx, first)+createImport(isModule, defaultModule, names.sortedJoined()))}
	}
	if existing := findImport(statements); existing != nil {
		return mergeIntoImport(existing, names)
	}
	if existing := findRequire(statements); existing != nil {
		collectRequireNames(existing.declaration.Name(), names)
		return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, existing.node, createImport(isModule, existing.module, names.sortedJoined()))}
	}
	return []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, first, createImport(isModule, defaultModule, names.sortedJoined())+"\n"+lineIndent(ctx, first))}
}

var PreferImportingRstestGlobalsRule = rule.Rule{
	Name:   "rstest/prefer-importing-rstest-globals",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		names := newNameSet(4)
		var reportingNode *ast.Node
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if !shouldReportIdentifier(ctx, node) {
					return
				}
				names.add(node.AsIdentifier().Text)
				if reportingNode == nil {
					reportingNode = node
				}
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(_ *ast.Node) {
				if reportingNode == nil || len(names.order) == 0 {
					return
				}
				collected := slices.Clone(names.order)
				ctx.ReportNodeWithDeferredFixes(reportingNode, preferMessage(names.joined()), func() []rule.RuleFix {
					return buildAutofix(ctx, collected)
				})
			},
		}
	},
}
