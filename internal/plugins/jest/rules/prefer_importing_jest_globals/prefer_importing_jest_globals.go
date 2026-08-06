package prefer_importing_jest_globals

import (
	_ "embed"
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

// nameSet preserves insertion order while de-duplicating import/require names,
// matching upstream's Set semantics for both the diagnostic message and autofix.
type nameSet struct {
	seen  map[string]struct{}
	order []string
}

func newNameSet(capacity int) *nameSet {
	return &nameSet{seen: make(map[string]struct{}, capacity)}
}

func newNameSetFrom(names []string) *nameSet {
	s := newNameSet(len(names))
	for _, name := range names {
		s.add(name)
	}
	return s
}

func (s *nameSet) add(name string) {
	if name == "" {
		return
	}
	if _, exists := s.seen[name]; exists {
		return
	}
	s.seen[name] = struct{}{}
	s.order = append(s.order, name)
}

func (s *nameSet) joined() string {
	return strings.Join(s.order, ", ")
}

func (s *nameSet) sortedJoined() string {
	sorted := slices.Clone(s.order)
	slices.Sort(sorted)
	return strings.Join(sorted, ", ")
}

type options struct {
	types map[utils.JestFnType]struct{}
}

func parseOptions(rawOptions []any) options {
	if len(rawOptions) == 0 {
		return options{types: defaultTypesSet()}
	}

	optsMap, _ := rawOptions[0].(map[string]interface{})
	rawTypes, ok := optsMap["types"].([]interface{})
	if !ok {
		return options{types: defaultTypesSet()}
	}

	parsed := make(map[utils.JestFnType]struct{}, len(rawTypes))
	for _, item := range rawTypes {
		s, ok := item.(string)
		if !ok {
			continue
		}
		kind := utils.JestFnType(s)
		if slices.Contains(allJestFnTypes, kind) {
			parsed[kind] = struct{}{}
		}
	}
	return options{types: parsed}
}

func defaultTypesSet() map[utils.JestFnType]struct{} {
	types := make(map[utils.JestFnType]struct{}, len(allJestFnTypes))
	for _, kind := range allJestFnTypes {
		types[kind] = struct{}{}
	}
	return types
}

func (o options) allows(kind utils.JestFnType) bool {
	_, ok := o.types[kind]
	return ok
}

func createFixerImports(isModule bool, names string) string {
	if isModule {
		return "import { " + names + " } from '" + jestGlobalsModule + "';"
	}
	return "const { " + names + " } = require('" + jestGlobalsModule + "');"
}

// preferModuleImport mirrors upstream's
// `parserOptions/languageOptions.sourceType === 'module'` check, with a
// structural ESM fallback when sourceType is unset (rslint default).
func preferModuleImport(ctx rule.RuleContext) bool {
	switch ctx.SourceType {
	case "module":
		return true
	case "script", "commonjs":
		return false
	default:
		return ctx.SourceFile != nil && ast.IsExternalModule(ctx.SourceFile)
	}
}

// accessorValue mirrors eslint-plugin-jest's isSupportedAccessor + getAccessorValue.
func accessorValue(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text, true
	case ast.KindNumericLiteral:
		return node.AsNumericLiteral().Text, true
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return rslintUtils.GetStaticStringValue(node), true
	default:
		return "", false
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

	local, _ := accessorValue(spec.Name())
	if spec.PropertyName == nil {
		return local
	}

	imported := spec.PropertyName
	switch imported.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		importedName := rslintUtils.GetStaticStringValue(imported)
		if local == "" {
			return "'" + importedName + "'"
		}
		return "'" + importedName + "' as " + local
	default:
		importName, _ := accessorValue(imported)
		if local != "" && local != importName {
			return importName + " as " + local
		}
		return importName
	}
}

func collectExistingImportNames(importDecl *ast.ImportDeclaration, names *nameSet) {
	if importDecl == nil || importDecl.ImportClause == nil {
		return
	}

	clause := importDecl.ImportClause.AsImportClause()
	if clause.Name() != nil {
		names.add(clause.Name().Text())
	}
	if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
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
		names.add(importNameFromSpecifier(el.AsImportSpecifier()))
	}
}

func collectExistingRequireNames(binding *ast.Node, names *nameSet) {
	if binding == nil || binding.Kind != ast.KindObjectBindingPattern {
		return
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
		importName, ok := accessorValue(keyNode)
		if !ok {
			continue
		}
		if local, ok := accessorValue(be.Name()); ok && importName != local {
			importName += ": " + local
		}
		names.add(importName)
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

type jestGlobalsRequire struct {
	stmt *ast.Node
	decl *ast.VariableDeclaration
}

func findJestGlobalsRequire(statements []*ast.Node) *jestGlobalsRequire {
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
				return &jestGlobalsRequire{stmt: stmt, decl: decl}
			}
		}
	}
	return nil
}

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

func buildAutofix(ctx rule.RuleContext, collected []string) []rule.RuleFix {
	if ctx.SourceFile == nil || ctx.SourceFile.Statements == nil || len(ctx.SourceFile.Statements.Nodes) == 0 {
		return nil
	}

	statements := ctx.SourceFile.Statements.Nodes
	firstNode := statements[0]
	names := newNameSetFrom(collected)
	isModule := preferModuleImport(ctx)

	if isUseStrictDirective(firstNode) {
		indent := lineIndentOfToken(ctx, firstNode)
		return []rule.RuleFix{
			rule.RuleFixInsertAfter(firstNode, "\n"+indent+createFixerImports(isModule, names.sortedJoined())),
		}
	}

	if importNode := findJestGlobalsImport(statements); importNode != nil {
		collectExistingImportNames(importNode.AsImportDeclaration(), names)
		return []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, importNode, createFixerImports(isModule, names.sortedJoined())),
		}
	}

	if req := findJestGlobalsRequire(statements); req != nil {
		if req.decl.Name() != nil && req.decl.Name().Kind == ast.KindObjectBindingPattern {
			collectExistingRequireNames(req.decl.Name(), names)
		}
		return []rule.RuleFix{
			rule.RuleFixReplace(ctx.SourceFile, req.stmt, createFixerImports(isModule, names.sortedJoined())),
		}
	}

	// Keep the following statement's indentation after the inserted newline.
	indent := lineIndentOfToken(ctx, firstNode)
	return []rule.RuleFix{
		rule.RuleFixInsertBefore(ctx.SourceFile, firstNode, createFixerImports(isModule, names.sortedJoined())+"\n"+indent),
	}
}

var PreferImportingJestGlobalsRule = rule.Rule{
	Name:   "jest/prefer-importing-jest-globals",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		names := newNameSet(4)
		var reportingNode *ast.Node

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				jestFnCall := utils.ParseJestFnCall(node, ctx)
				if jestFnCall == nil ||
					jestFnCall.Head.Type == utils.JEST_IMPORT_MODE ||
					!opts.allows(jestFnCall.Kind) {
					return
				}

				names.add(jestFnCall.Name)
				if reportingNode == nil {
					reportingNode = jestFnCall.Head.Local.Node
				}
			},
			rule.ListenerOnExit(ast.KindEndOfFile): func(node *ast.Node) {
				_ = node
				if reportingNode == nil || len(names.order) == 0 {
					return
				}

				// Snapshot order for the deferred fixer; collectExisting* mutates a copy.
				collected := slices.Clone(names.order)
				msg := buildPreferImportingJestGlobalMessage(names.joined())
				ctx.ReportNodeWithDeferredFixes(reportingNode, msg, func() []rule.RuleFix {
					return buildAutofix(ctx, collected)
				})
			},
		}
	},
}
