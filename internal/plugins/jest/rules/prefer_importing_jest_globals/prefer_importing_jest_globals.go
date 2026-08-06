package prefer_importing_jest_globals

import (
	_ "embed"
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslintUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed prefer_importing_jest_globals.schema.json
var schemaJSON []byte

const jestGlobalsModule = "@jest/globals"

var allJestFnTypes = []utils.JestFnType{
	utils.JestFnTypeHook,
	utils.JestFnTypeDescribe,
	utils.JestFnTypeTest,
	utils.JestFnTypeExpect,
	utils.JestFnTypeJest,
	utils.JestFnTypeUnknown,
}

func buildPreferImportingJestGlobalMessage(jestFunctions string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferImportingJestGlobal",
		Description: "Import the following Jest functions from '@jest/globals': " + jestFunctions,
		Data: map[string]string{
			"jestFunctions": jestFunctions,
		},
	}
}

type options struct {
	types []utils.JestFnType
}

func parseOptions(rawOptions []any) options {
	opts := options{types: slices.Clone(allJestFnTypes)}
	if len(rawOptions) == 0 {
		return opts
	}

	optsMap, _ := rawOptions[0].(map[string]interface{})
	rawTypes, ok := optsMap["types"].([]interface{})
	if !ok {
		return opts
	}

	parsed := make([]utils.JestFnType, 0, len(rawTypes))
	for _, item := range rawTypes {
		s, ok := item.(string)
		if !ok {
			continue
		}
		kind := utils.JestFnType(s)
		if slices.Contains(allJestFnTypes, kind) {
			parsed = append(parsed, kind)
		}
	}
	opts.types = parsed
	return opts
}

func createFixerImports(isModule bool, names []string) string {
	formatted := strings.Join(names, ", ")
	if isModule {
		return fmt.Sprintf("import { %s } from '%s';", formatted, jestGlobalsModule)
	}
	return fmt.Sprintf("const { %s } = require('%s');", formatted, jestGlobalsModule)
}

func isSupportedAccessor(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindIdentifier,
		ast.KindStringLiteral,
		ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNumericLiteral:
		return true
	default:
		return false
	}
}

func getAccessorValue(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text
	case ast.KindNoSubstitutionTemplateLiteral:
		return node.AsNoSubstitutionTemplateLiteral().Text
	case ast.KindNumericLiteral:
		return node.AsNumericLiteral().Text
	default:
		return ""
	}
}

func isUseStrictDirective(stmt *ast.Node) bool {
	if stmt == nil || stmt.Kind != ast.KindExpressionStatement {
		return false
	}
	expr := ast.SkipParentheses(stmt.AsExpressionStatement().Expression)
	return rslintUtils.GetStaticStringValue(expr) == "use strict"
}

func importNameFromSpecifier(spec *ast.ImportSpecifier) string {
	if spec == nil {
		return ""
	}

	localNode := spec.Name()
	local := getAccessorValue(localNode)

	if spec.PropertyName != nil {
		imported := spec.PropertyName
		switch imported.Kind {
		case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
			importName := ""
			if local != importName {
				importName = " as " + local
			}
			return "'" + getAccessorValue(imported) + "'" + importName
		default:
			importName := getAccessorValue(imported)
			if local != importName {
				importName = importName + " as " + local
			}
			return importName
		}
	}

	return local
}

func collectExistingImportNames(importDecl *ast.ImportDeclaration, into map[string]struct{}, order *[]string) {
	if importDecl == nil || importDecl.ImportClause == nil {
		return
	}

	add := func(name string) {
		if name == "" {
			return
		}
		if _, exists := into[name]; exists {
			return
		}
		into[name] = struct{}{}
		*order = append(*order, name)
	}

	clause := importDecl.ImportClause.AsImportClause()
	if clause.Name() != nil {
		add(clause.Name().Text())
	}
	if clause.NamedBindings == nil {
		return
	}
	if clause.NamedBindings.Kind != ast.KindNamedImports {
		return
	}
	named := clause.NamedBindings.AsNamedImports()
	if named.Elements == nil {
		return
	}
	for _, el := range named.Elements.Nodes {
		if el == nil || el.Kind != ast.KindImportSpecifier {
			continue
		}
		add(importNameFromSpecifier(el.AsImportSpecifier()))
	}
}

func collectExistingRequireNames(binding *ast.Node, into map[string]struct{}, order *[]string) {
	if binding == nil || binding.Kind != ast.KindObjectBindingPattern {
		return
	}

	add := func(name string) {
		if name == "" {
			return
		}
		if _, exists := into[name]; exists {
			return
		}
		into[name] = struct{}{}
		*order = append(*order, name)
	}

	pattern := binding.AsBindingPattern()
	if pattern.Elements == nil {
		return
	}
	for _, el := range pattern.Elements.Nodes {
		if el == nil || el.Kind != ast.KindBindingElement {
			continue
		}
		be := el.AsBindingElement()
		if be.DotDotDotToken != nil {
			continue
		}

		keyNode := be.PropertyName
		if keyNode == nil {
			keyNode = be.Name()
		}
		if !isSupportedAccessor(keyNode) {
			continue
		}

		importName := getAccessorValue(keyNode)
		valueNode := be.Name()
		if isSupportedAccessor(valueNode) {
			local := getAccessorValue(valueNode)
			if importName != local {
				importName += ": " + local
			}
		}
		add(importName)
	}
}

func findJestGlobalsImport(statements []*ast.Node) *ast.Node {
	for _, stmt := range statements {
		if stmt == nil || stmt.Kind != ast.KindImportDeclaration {
			continue
		}
		decl := stmt.AsImportDeclaration()
		if decl.ModuleSpecifier != nil &&
			rslintUtils.GetStaticStringValue(decl.ModuleSpecifier) == jestGlobalsModule {
			return stmt
		}
	}
	return nil
}

func findJestGlobalsRequire(statements []*ast.Node) *ast.Node {
	for _, stmt := range statements {
		if stmt == nil || stmt.Kind != ast.KindVariableStatement {
			continue
		}
		vs := stmt.AsVariableStatement()
		if vs.DeclarationList == nil {
			continue
		}
		list := vs.DeclarationList.AsVariableDeclarationList()
		if list.Declarations == nil {
			continue
		}
		for _, declNode := range list.Declarations.Nodes {
			if declNode == nil {
				continue
			}
			decl := declNode.AsVariableDeclaration()
			if decl == nil || decl.Initializer == nil {
				continue
			}
			name := decl.Name()
			if name == nil ||
				(name.Kind != ast.KindIdentifier && name.Kind != ast.KindObjectBindingPattern) {
				continue
			}
			if testFramework.IsModuleRequireCall(decl.Initializer, jestGlobalsModule) {
				return stmt
			}
		}
	}
	return nil
}

func sortedNames(order []string) []string {
	sorted := slices.Clone(order)
	slices.Sort(sorted)
	return sorted
}

// lineIndentOfToken returns the whitespace between the start of the line and
// the first token of node. Used so inserted imports keep the surrounding
// indentation of indented fixtures (upstream tests use dedent / column 0).
func lineIndentOfToken(ctx rule.RuleContext, node *ast.Node) string {
	if ctx.SourceFile == nil || node == nil {
		return ""
	}
	text := ctx.SourceFile.Text()
	trimmed := rslintUtils.TrimNodeTextRange(ctx.SourceFile, node)
	pos := trimmed.Pos()
	if pos < 0 || pos > len(text) {
		return ""
	}
	lineStart := strings.LastIndex(text[:pos], "\n") + 1
	indent := text[lineStart:pos]
	for _, r := range indent {
		if r != ' ' && r != '\t' {
			return ""
		}
	}
	return indent
}

func buildAutofix(
	ctx rule.RuleContext,
	namesOrder []string,
) []rule.RuleFix {
	if ctx.SourceFile == nil || ctx.SourceFile.Statements == nil || len(ctx.SourceFile.Statements.Nodes) == 0 {
		return nil
	}

	statements := ctx.SourceFile.Statements.Nodes
	firstNode := statements[0]

	into := make(map[string]struct{}, len(namesOrder))
	order := slices.Clone(namesOrder)
	for _, name := range namesOrder {
		into[name] = struct{}{}
	}

	isModule := ast.IsExternalModule(ctx.SourceFile)
	imports := createFixerImports(isModule, sortedNames(order))
	indent := lineIndentOfToken(ctx, firstNode)

	if isUseStrictDirective(firstNode) {
		return []rule.RuleFix{
			rule.RuleFixInsertAfter(firstNode, "\n"+indent+imports),
		}
	}

	if importNode := findJestGlobalsImport(statements); importNode != nil {
		collectExistingImportNames(importNode.AsImportDeclaration(), into, &order)
		replacement := createFixerImports(isModule, sortedNames(order))
		return []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, importNode, replacement),
		}
	}

	if requireNode := findJestGlobalsRequire(statements); requireNode != nil {
		vs := requireNode.AsVariableStatement()
		if vs.DeclarationList != nil {
			list := vs.DeclarationList.AsVariableDeclarationList()
			if list.Declarations != nil && len(list.Declarations.Nodes) > 0 {
				decl := list.Declarations.Nodes[0].AsVariableDeclaration()
				if decl != nil && decl.Name() != nil && decl.Name().Kind == ast.KindObjectBindingPattern {
					collectExistingRequireNames(decl.Name(), into, &order)
				}
			}
		}
		replacement := createFixerImports(isModule, sortedNames(order))
		return []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, requireNode, replacement),
		}
	}

	// Keep the following statement's indentation after the inserted newline.
	return []rule.RuleFix{
		rule.RuleFixInsertBefore(ctx.SourceFile, firstNode, imports+"\n"+indent),
	}
}

var PreferImportingJestGlobalsRule = rule.Rule{
	Name:   "jest/prefer-importing-jest-globals",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		functionsToImport := make(map[string]struct{})
		var importOrder []string
		var reportingNode *ast.Node

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil {
					return
				}
				if jestFnCall.Head.Type == utils.JEST_IMPORT_MODE {
					return
				}
				if !slices.Contains(opts.types, jestFnCall.Kind) {
					return
				}

				name := jestFnCall.Name
				if _, exists := functionsToImport[name]; !exists {
					functionsToImport[name] = struct{}{}
					importOrder = append(importOrder, name)
				}
				if reportingNode == nil {
					reportingNode = jestFnCall.Head.Local.Node
				}
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(node *ast.Node) {
				_ = node
				if reportingNode == nil || len(importOrder) == 0 {
					return
				}

				namesForMessage := slices.Clone(importOrder)
				msg := buildPreferImportingJestGlobalMessage(strings.Join(namesForMessage, ", "))
				namesForFix := slices.Clone(importOrder)

				ctx.ReportNodeWithDeferredFixes(reportingNode, msg, func() []rule.RuleFix {
					return buildAutofix(ctx, namesForFix)
				})
			},
		}
	},
}
