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

// isShorthandPropertyRead reports whether the identifier is the value side of
// a shorthand property such as `{ expect }`. The AST also treats it as the
// property's declaration name, so the name filters below would drop it even
// though it reads the global. A shorthand that is written to, as in
// `({ expect } = value)` or a `for` destructuring head, is not a read.
func isShorthandPropertyRead(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil || parent.Kind != ast.KindShorthandPropertyAssignment || parent.Name() != node ||
		parent.AsShorthandPropertyAssignment().ObjectAssignmentInitializer != nil {
		return false
	}
	object := parent.Parent
	if object == nil || object.Parent == nil {
		return false
	}
	switch owner := object.Parent; owner.Kind {
	case ast.KindBinaryExpression:
		binary := owner.AsBinaryExpression()
		return binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken || binary.Left != object
	case ast.KindForInStatement, ast.KindForOfStatement:
		return owner.AsForInOrOfStatement().Initializer != object
	}
	return true
}

func shouldReportIdentifier(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier || !rstestUtils.IsRstestGlobal(node.AsIdentifier().Text) {
		return false
	}
	if !isShorthandPropertyRead(node) && (ast.IsIdentifierName(node) || ast.IsDeclarationNameOrImportPropertyName(node)) {
		return false
	}
	// A type position never touches the value at runtime, so `const value:
	// expect = input;` does not need the API imported.
	if ast.IsPartOfTypeNode(node) || ast.IsPartOfTypeQuery(node) {
		return false
	}
	// A write such as `expect = value;` assigns to the global. Importing the
	// name would turn the target into a read-only module binding, so the write
	// is not a use that the import can satisfy.
	if internalUtils.IsWriteReference(node) {
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

// directiveText returns the text of a directive statement. Only a bare
// string-literal expression statement belongs to the directive prologue, so a
// template literal or a parenthesized string is an ordinary statement and must
// not shift the insertion point.
func directiveText(statement *ast.Node) (string, bool) {
	if statement == nil || statement.Kind != ast.KindExpressionStatement {
		return "", false
	}
	expression := statement.AsExpressionStatement().Expression
	if expression == nil || expression.Kind != ast.KindStringLiteral {
		return "", false
	}
	return expression.AsStringLiteral().Text, true
}

// strictDirectivePrologue returns the last statement of the directive prologue
// when the prologue turns the file strict. Inserting anywhere inside or before
// the prologue would end it early and silently drop strict mode, so the whole
// prologue has to be scanned rather than only its first statement.
func strictDirectivePrologue(statements []*ast.Node) *ast.Node {
	var last *ast.Node
	strict := false
	for _, statement := range statements {
		text, ok := directiveText(statement)
		if !ok {
			break
		}
		if text == "use strict" {
			strict = true
		}
		last = statement
	}
	if !strict {
		return nil
	}
	return last
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

// mergeIntoRequire adds the missing names to an existing destructured require
// the same way mergeIntoImport does for an import. Rewriting the statement as
// an import instead would have to translate `{ it: testCase }` into import
// syntax, and would drop whatever the generated text cannot express.
func mergeIntoRequire(sourceFile *ast.SourceFile, binding *ast.Node, names *nameSet) []rule.RuleFix {
	if binding == nil || binding.Kind != ast.KindObjectBindingPattern {
		return nil
	}
	pattern := binding.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil || len(pattern.Elements.Nodes) == 0 {
		return nil
	}
	present := make(map[string]struct{}, len(pattern.Elements.Nodes))
	var last *ast.Node
	var rest *ast.Node
	for _, element := range pattern.Elements.Nodes {
		be := element.AsBindingElement()
		if be == nil {
			return nil
		}
		if be.DotDotDotToken != nil {
			rest = element
			continue
		}
		if be.Name() != nil && be.Name().Kind == ast.KindIdentifier {
			present[be.Name().AsIdentifier().Text] = struct{}{}
		}
		last = element
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
	// A rest element has to stay last, so names go in front of it.
	if last == nil {
		if rest == nil {
			return nil
		}
		return []rule.RuleFix{rule.RuleFixInsertBefore(sourceFile, rest, additions.sortedJoined()+", ")}
	}
	return []rule.RuleFix{rule.RuleFixInsertAfter(last, ", "+additions.sortedJoined())}
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

// findRequire returns the object binding pattern of an existing destructured
// require of an rstest module.
func findRequire(statements []*ast.Node) *ast.Node {
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
		if _, ok := rstestUtils.RstestCoreModuleFromRequireCall(declaration.Initializer); ok {
			return declaration.Name()
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
	if prologue := strictDirectivePrologue(statements); prologue != nil {
		return []rule.RuleFix{rule.RuleFixInsertAfter(prologue, "\n"+lineIndent(ctx, prologue)+createImport(isModule, defaultModule, names.sortedJoined()))}
	}
	if existing := findImport(statements); existing != nil {
		return mergeIntoImport(existing, names)
	}
	if existing := findRequire(statements); existing != nil {
		return mergeIntoRequire(ctx.SourceFile, existing, names)
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
