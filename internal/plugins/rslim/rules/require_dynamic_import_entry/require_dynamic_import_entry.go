package require_dynamic_import_entry

import (
	"fmt"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var RequireDynamicImportEntryRule = rule.Rule{
	Name:             "rslim/require-dynamic-import-entry",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Expression.Kind != ast.KindImportKeyword || call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				specifier := call.Arguments.Nodes[0]
				if !ast.IsStringLiteralLike(specifier) {
					return
				}
				module := ctx.TypeChecker.GetSymbolAtLocation(specifier)
				if module == nil {
					return
				}
				used := exportUsage{ctx: &ctx, names: map[string]bool{}, seen: map[*ast.Node]bool{}}
				used.expression(node, true)
				var missing []string
				for _, exported := range ctx.TypeChecker.GetExportsOfModule(module) {
					if !used.all && !used.names[exported.Name] {
						continue
					}
					symbol := exported
					if symbol.Flags&ast.SymbolFlagsAlias != 0 {
						if ctx.TypeChecker.GetTypeOnlyAliasDeclaration(symbol) != nil {
							continue
						}
						symbol = ctx.TypeChecker.GetAliasedSymbol(symbol)
					}
					if symbol == nil || symbol.Flags&ast.SymbolFlagsValue == 0 {
						continue
					}
					if needsEntry(ctx, symbol) {
						missing = append(missing, exported.Name)
					}
				}
				if len(missing) > 0 {
					slices.Sort(missing)
					ctx.ReportNode(specifier, rule.RuleMessage{
						Id:          "missingEntry",
						Description: fmt.Sprintf("Add @entry to the declarations of these dynamically imported exports from %q: %s.", specifier.Text(), strings.Join(missing, ", ")),
					})
				}
			},
		}
	},
}

// Follow local bindings so a known property selection does not require marking
// unrelated exports. Escaping namespaces and computed keys require every export.
// The analysis follows uses within this file, without tracing function calls.
type exportUsage struct {
	ctx   *rule.RuleContext
	names map[string]bool
	seen  map[*ast.Node]bool
	all   bool
}

func (u *exportUsage) binding(declaration *ast.Node, promise bool) {
	name := declaration.Name()
	if name == nil {
		u.all = true
		return
	}
	if name.Kind == ast.KindObjectBindingPattern && !promise {
		for _, element := range name.AsBindingPattern().Elements.Nodes {
			binding := element.AsBindingElement()
			key := binding.PropertyName
			if key == nil {
				key = binding.Name()
			}
			if binding.DotDotDotToken != nil {
				u.all = true
			} else if text, ok := utils.GetStaticPropertyName(key); ok {
				u.names[text] = true
			} else {
				u.all = true
			}
		}
		return
	}
	if ast.GetCombinedModifierFlags(declaration)&ast.ModifierFlagsExport != 0 {
		u.all = true
		return
	}
	if name.Kind != ast.KindIdentifier || declaration.Symbol() == nil || u.ctx.Refs == nil {
		u.all = true
		return
	}
	for _, reference := range u.ctx.Refs.References(declaration.Symbol()) {
		u.expression(reference, promise)
	}
}

func (u *exportUsage) expression(node *ast.Node, promise bool) {
	if u.seen[node] {
		return
	}
	u.seen[node] = true
	parent := node.Parent
	if parent == nil {
		u.all = true
		return
	}
	switch parent.Kind {
	case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression, ast.KindSatisfiesExpression:
		u.expression(parent, promise)
		return
	case ast.KindAwaitExpression:
		u.expression(parent, false)
		return
	case ast.KindVariableDeclaration:
		if parent.AsVariableDeclaration().Initializer == node {
			u.binding(parent, promise)
			return
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if parent.Expression() != node {
			break
		}
		name, known := utils.AccessExpressionStaticName(parent)
		if !promise {
			if known {
				u.names[name] = true
			} else {
				u.all = true
			}
			return
		}
		if known && name == "then" && parent.Parent != nil && parent.Parent.Kind == ast.KindCallExpression {
			call := parent.Parent.AsCallExpression()
			if call.Expression == parent && call.Arguments != nil && len(call.Arguments.Nodes) > 0 {
				callback := ast.SkipOuterExpressions(call.Arguments.Nodes[0], ast.OEKAll)
				if ast.IsArrowFunction(callback) || ast.IsFunctionExpression(callback) {
					parameters := callback.Parameters()
					if len(parameters) > 0 {
						u.binding(parameters[0], false)
					}
					return
				}
			}
		}
	case ast.KindExpressionStatement, ast.KindVoidExpression:
		// A discarded import result only requests module side effects.
		return
	}
	u.all = true
}

func needsEntry(ctx rule.RuleContext, symbol *ast.Symbol) bool {
	return needsEntrySeen(ctx, symbol, map[*ast.Symbol]bool{})
}

func needsEntrySeen(ctx rule.RuleContext, symbol *ast.Symbol, seen map[*ast.Symbol]bool) bool {
	if seen[symbol] {
		return false
	}
	seen[symbol] = true
	hasSource := false
	for _, declaration := range symbol.Declarations {
		file := ast.GetSourceFileOfNode(declaration)
		if file == nil || file.IsDeclarationFile || utils.IsInAmbientContext(declaration) || ctx.Program().IsSourceFileFromExternalLibrary(file) {
			continue
		}
		// `export * as ns` aliases a source-file symbol. Check its exports;
		// placing @entry on a source file cannot preserve that namespace.
		if declaration.Kind == ast.KindSourceFile {
			for _, exported := range ctx.TypeChecker.GetExportsOfModule(symbol) {
				if exported.Flags&ast.SymbolFlagsAlias != 0 {
					if ctx.TypeChecker.GetTypeOnlyAliasDeclaration(exported) != nil {
						continue
					}
					exported = ctx.TypeChecker.GetAliasedSymbol(exported)
				}
				if exported != nil && exported.Flags&ast.SymbolFlagsValue != 0 && needsEntrySeen(ctx, exported, seen) {
					return true
				}
			}
			continue
		}
		hasSource = true
		if hasEntry(file, declaration) {
			return false
		}
		if declaration.Kind == ast.KindVariableDeclaration && declaration.Parent != nil && declaration.Parent.Parent != nil && declaration.Parent.Parent.Kind == ast.KindVariableStatement {
			if hasEntry(file, declaration.Parent.Parent) {
				return false
			}
		}
	}
	return hasSource
}

// Rslim accepts an annotation only when it occupies a complete comment line.
// Match Rslim's leading-trivia lookup followed by its trailing-trivia fallback
// at the declaration's full start.
func hasEntry(file *ast.SourceFile, node *ast.Node) bool {
	factory := &ast.NodeFactory{}
	text := file.Text()
	var comments []ast.CommentRange
	for comment := range scanner.GetLeadingCommentRanges(factory, text, node.Pos()) {
		comments = append(comments, comment)
	}
	if len(comments) == 0 {
		for comment := range scanner.GetTrailingCommentRanges(factory, text, node.Pos()) {
			if comment.End() <= node.End() {
				comments = append(comments, comment)
			}
		}
	}
	if len(comments) == 0 {
		return false
	}
	// Preserve the entire comment group: two block comments on the same line
	// are not two separate annotation lines in Rslim.
	for _, line := range strings.Split(text[comments[0].Pos():comments[len(comments)-1].End()], "\n") {
		line = ecmascript.StringTrim(line)
		line = strings.TrimPrefix(line, "/*")
		line = strings.TrimSuffix(line, "*/")
		line = strings.Trim(ecmascript.StringTrim(line), "*")
		line = strings.Trim(line, "/")
		if strings.TrimLeftFunc(line, ecmascript.IsWhiteSpaceOrLineTerminator) == "@entry" {
			return true
		}
	}
	return false
}
