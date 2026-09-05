package prefer_importing_rstest_globals

import (
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
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

type identifierUse int

const (
	// useIrrelevant marks an identifier that only happens to be spelled like
	// an Rstest global: a property key, a label, a declaration name, a type
	// position, or a name the file declares itself.
	useIrrelevant identifierUse = iota
	// useWrite marks an assignment to the global, including a compound
	// assignment, an update, and a destructuring target.
	useWrite
	// useRead marks a runtime read of the global, the only use an import can
	// stand in for.
	useRead
)

// classifyIdentifier decides what an identifier does with the Rstest global of
// the same name. IsReferenceIdentifier matches the positions Refs.Resolve
// accepts, including JSX names. IsReadReference excludes type-only uses, and
// IsWriteReference distinguishes runtime reads from assignments.
func classifyIdentifier(ctx rule.RuleContext, node *ast.Node) identifierUse {
	if node == nil || node.Kind != ast.KindIdentifier || !rstestUtils.IsRstestGlobal(node.AsIdentifier().Text) {
		return useIrrelevant
	}
	if !scope.IsReferenceIdentifier(node) || !internalUtils.IsReadReference(node) {
		return useIrrelevant
	}
	for parent := node.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport,
			ast.KindExportSpecifier, ast.KindExportDeclaration:
			return useIrrelevant
		}
		if parent.Kind == ast.KindSourceFile || parent.Kind == ast.KindBlock || ast.IsExpression(parent) {
			break
		}
	}
	if ctx.Refs != nil {
		if symbol := ctx.Refs.Resolve(node); symbol != nil {
			name := node.AsIdentifier().Text
			resolved, _, mode := testFramework.ResolveFunctionIdentifierReferenceFromSymbolModules(
				name, node, symbol, ctx.SourceFile, rstestUtils.RstestCoreImportModules,
			)
			if mode == testFramework.ReferenceModeImport && resolved != "" {
				return useIrrelevant
			}
			if internalUtils.IsRuntimeValueSymbolDeclaredInFile(symbol, ctx.SourceFile) {
				return useIrrelevant
			}
		}
	}
	// A write such as `expect = value;` assigns to the global. An import
	// binding is read-only, so the write is not a use an import can satisfy —
	// and, when the same name is read elsewhere, importing it would break the
	// write that is already there.
	if internalUtils.IsWriteReference(node) {
		return useWrite
	}
	return useRead
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
func mergeIntoRequire(binding *ast.Node, names *nameSet) []rule.RuleFix {
	if binding == nil || binding.Kind != ast.KindObjectBindingPattern {
		return nil
	}
	pattern := binding.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil || len(pattern.Elements.Nodes) == 0 {
		return nil
	}
	present := make(map[string]struct{}, len(pattern.Elements.Nodes))
	var last *ast.Node
	for _, element := range pattern.Elements.Nodes {
		be := element.AsBindingElement()
		if be == nil {
			return nil
		}
		// A rest element collects every property the earlier bindings leave
		// behind, so naming one more of them here would quietly take it out of
		// the rest object and break the code that reads it off there. Putting
		// the name after the rest element is not syntax, so the fix is dropped.
		if be.DotDotDotToken != nil {
			return nil
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
	if last == nil {
		return nil
	}
	return []rule.RuleFix{rule.RuleFixInsertAfter(last, ", "+additions.sortedJoined())}
}

// bindsNameAsTypeOnly reports whether the file already binds one of names as
// a type-only import. Such a binding provides no value, so the rule still asks
// for the import, but every fix it could write would declare the name a second
// time: adding a value specifier beside the type-only one, or a second
// declaration next to a whole `import type` clause, is a duplicate identifier
// either way. Turning the existing binding into a value import is the edit a
// reader would make, and it is not one this rule is in a position to choose,
// so the diagnostic is reported without a fix.
//
// Every module and every import form counts, because the collision is with the
// name in this file's scope, not with what the name was imported from. A
// binding that does survive to runtime never gets here: the identifier
// resolves to it and is not reported at all.
func bindsNameAsTypeOnly(statements []*ast.Node, names *nameSet) bool {
	collides := func(name *ast.Node) bool {
		if name == nil || name.Kind != ast.KindIdentifier {
			return false
		}
		_, exists := names.seen[name.AsIdentifier().Text]
		return exists
	}
	for _, statement := range statements {
		if statement == nil {
			continue
		}
		switch statement.Kind {
		case ast.KindImportEqualsDeclaration:
			declaration := statement.AsImportEqualsDeclaration()
			if declaration != nil && declaration.IsTypeOnly && collides(declaration.Name()) {
				return true
			}
		case ast.KindImportDeclaration:
			declaration := statement.AsImportDeclaration()
			if declaration == nil || declaration.ImportClause == nil {
				continue
			}
			clause := declaration.ImportClause.AsImportClause()
			if clause == nil {
				continue
			}
			clauseIsTypeOnly := declaration.ImportClause.IsTypeOnly()
			if clauseIsTypeOnly && collides(clause.Name()) {
				return true
			}
			bindings := clause.NamedBindings
			if bindings == nil {
				continue
			}
			if bindings.Kind == ast.KindNamespaceImport {
				if clauseIsTypeOnly && collides(bindings.Name()) {
					return true
				}
				continue
			}
			named := bindings.AsNamedImports()
			if named == nil || named.Elements == nil {
				continue
			}
			for _, element := range named.Elements.Nodes {
				specifier := element.AsImportSpecifier()
				if specifier == nil || (!clauseIsTypeOnly && !specifier.IsTypeOnly) {
					continue
				}
				if collides(specifier.Name()) {
					return true
				}
			}
		}
	}
	return false
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

// buildAutofix writes the import the diagnostic asks for. written carries the
// names the file also assigns to: importing such a name would turn the target
// of that assignment into a read-only module binding, so the whole fix is
// withheld and the diagnostic stands on its own. Skipping only the write
// reference is not enough — a single read of the same name would otherwise
// still pull in the import that breaks the write.
func buildAutofix(ctx rule.RuleContext, collected []string, written map[string]struct{}) []rule.RuleFix {
	names := newNameSet(len(collected))
	for _, name := range collected {
		if _, assigned := written[name]; assigned {
			return nil
		}
		names.add(name)
	}
	isModule := ctx.LanguageOptions.EffectiveSourceType() == "module"
	if ctx.SourceFile == nil || ctx.SourceFile.Statements == nil || len(ctx.SourceFile.Statements.Nodes) == 0 {
		return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(0, 0), createImport(isModule, defaultModule, names.sortedJoined()))}
	}
	statements := ctx.SourceFile.Statements.Nodes
	if bindsNameAsTypeOnly(statements, names) {
		return nil
	}
	first := statements[0]
	if prologue := strictDirectivePrologue(statements); prologue != nil {
		return []rule.RuleFix{rule.RuleFixInsertAfter(prologue, "\n"+lineIndent(ctx, prologue)+createImport(isModule, defaultModule, names.sortedJoined()))}
	}
	if existing := findImport(statements); existing != nil {
		return mergeIntoImport(existing, names)
	}
	if existing := findRequire(statements); existing != nil {
		return mergeIntoRequire(existing, names)
	}
	return []rule.RuleFix{rule.RuleFixInsertBefore(ctx.SourceFile, first, createImport(isModule, defaultModule, names.sortedJoined())+"\n"+lineIndent(ctx, first))}
}

var PreferImportingRstestGlobalsRule = rule.Rule{
	Name:   "rstest/prefer-importing-rstest-globals",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		names := newNameSet(4)
		written := map[string]struct{}{}
		var reportingNode *ast.Node
		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				switch classifyIdentifier(ctx, node) {
				case useWrite:
					written[node.AsIdentifier().Text] = struct{}{}
				case useRead:
					names.add(node.AsIdentifier().Text)
					if reportingNode == nil {
						reportingNode = node
					}
				}
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(_ *ast.Node) {
				if reportingNode == nil || len(names.order) == 0 {
					return
				}
				collected := slices.Clone(names.order)
				ctx.ReportNodeWithDeferredFixes(reportingNode, preferMessage(names.joined()), func() []rule.RuleFix {
					return buildAutofix(ctx, collected, written)
				})
			},
		}
	},
}
