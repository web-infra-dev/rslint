// Ported from eslint-plugin-unicorn v79.0.0; see LICENSE.
// cspell:words jsxfrag
package prefer_export_from

import (
	"cmp"
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

//go:embed prefer_export_from.schema.json
var schemaJSON []byte

var PreferExportFromRule = rule.Rule{
	Name:   "unicorn/prefer-export-from",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkUsedVariables := true
		if len(options) != 0 {
			settings, _ := options[0].(map[string]any)
			if value, ok := settings["checkUsedVariables"].(bool); ok {
				checkUsedVariables = value
			}
		}

		var imports []*ast.ImportClause
		var directExports []*ast.ExportDeclaration
		var localExports map[*ast.Node]localExport
		var jsxNodes []*ast.Node
		var hasDirectEval bool
		addExport := func(local *ast.Node, exported localExport) {
			if local == nil || local.Kind != ast.KindIdentifier {
				return
			}
			if localExports == nil {
				localExports = make(map[*ast.Node]localExport)
			}
			localExports[local] = exported
		}

		listeners := rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				decl := node.AsImportDeclaration()
				if decl.ImportClause != nil && decl.ModuleSpecifier != nil && ast.IsStringLiteral(decl.ModuleSpecifier) {
					imports = append(imports, decl.ImportClause.AsImportClause())
				}
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				decl := node.AsExportDeclaration()
				if decl.ExportClause == nil || decl.ExportClause.Kind != ast.KindNamedExports {
					return
				}
				if decl.ModuleSpecifier != nil {
					if ast.IsStringLiteral(decl.ModuleSpecifier) {
						directExports = append(directExports, decl)
					}
					return
				}
				for _, specifier := range decl.ExportClause.Elements() {
					addExport(specifier.PropertyNameOrName(), localExport{specifier, specifier.Name(), ast.IsTypeOnlyImportOrExportDeclaration(specifier)})
				}
			},
			ast.KindExportAssignment: func(node *ast.Node) {
				decl := node.AsExportAssignment()
				if !decl.IsExportEquals {
					addExport(utils.ESTreeRuntimeExpression(decl.Expression), localExport{node: node})
				}
			},
			ast.KindVariableStatement: func(node *ast.Node) {
				list := node.AsVariableStatement().DeclarationList
				if !ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) || !ast.IsVarConst(list) {
					return
				}
				declarations := list.AsVariableDeclarationList().Declarations.Nodes
				if len(declarations) == 1 {
					decl := declarations[0].AsVariableDeclaration()
					if decl.Type == nil && ast.IsIdentifier(decl.Name()) {
						addExport(utils.ESTreeRuntimeExpression(decl.Initializer), localExport{node: node, name: decl.Name()})
					}
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				if hasDirectEval || node.AsCallExpression().QuestionDotToken != nil {
					return
				}
				callee := ast.SkipOuterExpressions(node.Expression(), ast.OEKAll)
				hasDirectEval = ast.IsIdentifier(callee) && callee.Text() == "eval"
			},
			// The linter traverses the source file's children, ending at EOF.
			ast.KindEndOfFile: func(_ *ast.Node) {
				if len(imports) == 0 || len(localExports) == 0 {
					return
				}
				if names := reexportNames(imports, localExports); names != nil {
					for _, match := range findReexports(ctx, localExports, jsxNodes, names, checkUsedVariables) {
						reportReexport(ctx, match, directExports, hasDirectEval)
					}
				}
			},
		}
		if ctx.SourceFile.LanguageVariant == core.LanguageVariantJSX {
			collectJSX := func(node *ast.Node) { jsxNodes = append(jsxNodes, node) }
			listeners[ast.KindJsxOpeningElement] = collectJSX
			listeners[ast.KindJsxSelfClosingElement] = collectJSX
			listeners[ast.KindJsxFragment] = collectJSX
		}
		return listeners
	},
}

type localExport struct {
	node     *ast.Node
	name     *ast.Node // nil for `export default expression`
	typeOnly bool
}

func (exported localExport) statement() *ast.Node {
	if exported.node.Kind == ast.KindExportSpecifier {
		return exported.node.Parent.Parent
	}
	return exported.node
}

type reexport struct {
	localExport
	imported     *scope.Variable
	importCount  int
	removeImport bool
	suggest      bool
}

// Include every import name because checkUsedVariables considers the whole
// declaration. An exported const alias also needs its own usage checked.
func reexportNames(imports []*ast.ImportClause, localExports map[*ast.Node]localExport) map[string]struct{} {
	names := make(map[string]struct{})
	add := func(name *ast.Node) {
		if name != nil {
			names[name.Text()] = struct{}{}
		}
	}
	for _, clause := range imports {
		add(clause.Name())
		if bindings := clause.NamedBindings; bindings != nil {
			if bindings.Kind == ast.KindNamespaceImport {
				add(bindings.Name())
			} else {
				for _, specifier := range bindings.Elements() {
					add(specifier.Name())
				}
			}
		}
	}
	hasCandidate := false
	var aliases []*ast.Node
	for local, exported := range localExports {
		if _, imported := names[local.Text()]; imported {
			hasCandidate = true
			if exported.node.Kind == ast.KindVariableStatement {
				aliases = append(aliases, exported.name)
			}
		}
	}
	if !hasCandidate {
		return nil
	}
	for _, alias := range aliases {
		add(alias)
	}
	return names
}

func findReexports(ctx rule.RuleContext, localExports map[*ast.Node]localExport, jsxNodes []*ast.Node, names map[string]struct{}, checkUsedVariables bool) []reexport {
	manager := scopeanalysis.References(ctx, names)
	references := make(map[*scope.Variable][]*ast.Node)
	for _, reference := range manager.References {
		binding := reference.Resolved()
		if binding == nil {
			continue
		}
		if _, relevant := names[binding.Name]; relevant {
			references[binding] = append(references[binding], reference.Identifier)
		}
	}
	if len(jsxNodes) != 0 {
		addJSXFactoryUses(ctx, manager, jsxNodes, references)
	}

	var matches []reexport
	for _, lexicalScope := range manager.Scopes {
		// Scope records the bindings from one import together, in declaration order.
		for bindings := lexicalScope.Vars; len(bindings) != 0; {
			first := bindings[0]
			if first.Kind != scope.DefImport || !ast.IsImportDeclaration(first.DefNode) {
				bindings = bindings[1:]
				continue
			}
			end := 1
			for end < len(bindings) && bindings[end].DefNode == first.DefNode {
				end++
			}
			source := first.DefNode.AsImportDeclaration().ModuleSpecifier
			if source != nil && ast.IsStringLiteral(source) {
				matches = appendReexports(matches, bindings[:end], localExports, manager, references, checkUsedVariables)
			}
			bindings = bindings[end:]
		}
	}
	slices.SortFunc(matches, func(left, right reexport) int { return cmp.Compare(left.node.Pos(), right.node.Pos()) })
	return matches
}

func appendReexports(matches []reexport, bindings []*scope.Variable, localExports map[*ast.Node]localExport, manager *scope.Manager, references map[*scope.Variable][]*ast.Node, checkUsedVariables bool) []reexport {
	start := len(matches)
	var referenceCount int
	var hasUnusedBinding bool
	for _, imported := range bindings {
		if len(imported.Scope.Declarations(imported.Name)) != 1 {
			return matches[:start]
		}
		uses := references[imported]
		referenceCount += len(uses)
		hasUnusedBinding = hasUnusedBinding || len(uses) == 0
		for _, use := range uses {
			exported, ok := localExports[use]
			if !ok || imported.ID.Parent.Kind == ast.KindNamespaceImport && isDefaultName(exported.name) {
				continue
			}
			if exported.node.Kind == ast.KindVariableStatement {
				alias := manager.Acquire(exported.node).Declarations(exported.name.Text())
				if len(alias) != 1 || alias[0].ID != exported.name || len(references[alias[0]]) != 0 {
					continue
				}
			}
			matches = append(matches, reexport{
				localExport:  exported,
				imported:     imported,
				importCount:  len(bindings),
				removeImport: len(uses) == 1 && (!exported.typeOnly || imported.IsTypeOnlyImport),
			})
		}
	}
	if !checkUsedVariables {
		if len(matches)-start != referenceCount {
			return matches[:start]
		}
		for i := start; i < len(matches); i++ {
			matches[i].suggest = hasUnusedBinding
		}
	}
	return matches
}

// JSX can use a local factory without spelling its name at the use site.
// Resolve each site so an earlier shadowed factory cannot hide a later use.
func addJSXFactoryUses(ctx rule.RuleContext, manager *scope.Manager, nodes []*ast.Node, references map[*scope.Variable][]*ast.Node) {
	factory, fragment := "React", ""
	if program := ctx.Program(); program != nil {
		options := program.Options()
		if ctx.SourceFile.ScriptKind != core.ScriptKindTSX && ast.GetJSXImplicitImportBase(options, ctx.SourceFile) != "" {
			return
		}
		if options.JsxFactory != "" {
			factory = options.JsxFactory
		}
		fragment = options.JsxFragmentFactory
	}
	factory = jsxFactoryName(ctx.SourceFile, "jsx", factory)
	fragment = jsxFactoryName(ctx.SourceFile, "jsxfrag", fragment)
	mark := func(node *ast.Node, name string) {
		for lexical := manager.Acquire(node); lexical != nil; lexical = lexical.Parent {
			for _, binding := range lexical.Declarations(name) {
				if binding.IsValueBinding {
					references[binding] = append(references[binding], node)
					return
				}
			}
		}
	}
	for _, node := range nodes {
		if factory != "" {
			mark(node, factory)
		}
		if fragment != "" && node.Kind == ast.KindJsxFragment {
			mark(node, fragment)
		}
	}
}

func jsxFactoryName(file *ast.SourceFile, pragma, fallback string) string {
	name := ast.GetPragmaArgument(ast.GetPragmaFromSourceFile(file, pragma), "factory")
	if name == "" {
		name = fallback
	}
	root, _, _ := strings.Cut(name, ".")
	return ecmascript.StringTrim(root)
}

func isDefaultName(node *ast.Node) bool {
	return node == nil || node.Kind == ast.KindIdentifier && node.Text() == "default"
}

func nameText(sourceFile *ast.SourceFile, name *ast.Node) string {
	if name == nil {
		return "default"
	}
	return utils.TrimmedNodeText(sourceFile, name)
}

func reportReexport(ctx rule.RuleContext, match reexport, directExports []*ast.ExportDeclaration, hasDirectEval bool) {
	if match.node.Kind == ast.KindVariableStatement && hasAttachedJSDoc(ctx, match.node) {
		return
	}
	name := nameText(ctx.SourceFile, match.name)
	message := rule.RuleMessage{
		Id:          "error",
		Description: "Use `export…from` to re-export `" + name + "`.",
		Data:        map[string]string{"exported": name},
	}
	// Direct eval may read a local binding without an identifier reference in
	// the source. Keep diagnostics, but don't remove that binding.
	if hasDirectEval && (match.removeImport || match.node.Kind == ast.KindVariableStatement) {
		ctx.ReportNode(match.node, message)
		return
	}
	if match.suggest {
		ctx.ReportNodeWithDeferredSuggestions(match.node, message, func() []rule.RuleSuggestion {
			fixes := reexportFixes(ctx, match, directExports, name)
			if len(fixes) == 0 {
				return nil
			}
			return []rule.RuleSuggestion{{
				Message:  rule.RuleMessage{Id: "suggestion", Description: "Switch to `export…from`."},
				FixesArr: fixes,
			}}
		})
		return
	}
	ctx.ReportNodeWithDeferredFixes(match.node, message, func() []rule.RuleFix {
		return reexportFixes(ctx, match, directExports, name)
	})
}

func hasAttachedJSDoc(ctx rule.RuleContext, node *ast.Node) bool {
	sourceFile := ctx.SourceFile
	start := utils.TrimNodeTextRange(sourceFile, node).Pos()
	comments := utils.CommentsInSpan(ctx.Comments.All(), node.Pos(), start)
	if len(comments) == 0 {
		return false
	}
	comment := comments[len(comments)-1]
	if comment.Kind != ast.KindMultiLineCommentTrivia {
		return false
	}
	body := utils.CommentValue(sourceFile.Text(), comment)
	if !strings.HasPrefix(body, "*") || strings.HasPrefix(body, "**") {
		return false
	}
	if previous, ok := utils.TokenBeforePosition(sourceFile, comment.Pos()); ok && utils.IsSameLine(sourceFile, previous.End, comment.Pos()) {
		return false
	}
	lines := sourceFile.ECMALineMap()
	return scanner.ComputeLineOfPosition(lines, start)-scanner.ComputeLineOfPosition(lines, comment.End()) <= 1
}
