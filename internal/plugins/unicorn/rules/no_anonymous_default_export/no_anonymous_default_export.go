// Ported from eslint-plugin-unicorn v76.0.0 (MIT).
package no_anonymous_default_export

import (
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

var NoAnonymousDefaultExportRule = rule.Rule{
	Name:   "unicorn/no-anonymous-default-export",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		report := func(node, statement *ast.Node) {
			node = utils.ESTreeRuntimeExpression(node)
			if node == nil {
				return
			}
			var location core.TextRange
			var description string
			switch node.Kind {
			case ast.KindClassDeclaration, ast.KindClassExpression:
				if node.Name() != nil {
					return
				}
				location = classHeadRange(ctx.SourceFile, node)
				description = "class"
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
				if node.Name() != nil || node.Body() == nil {
					return
				}
				location = utils.GetFunctionHeadLoc(ctx.SourceFile, node)
				description, _, _ = strings.Cut(utils.GetFunctionNameWithKind(node), " '")
			default:
				return
			}
			ctx.ReportRangeWithDeferredSuggestions(location, rule.RuleMessage{
				Id:          "no-anonymous-default-export/error",
				Description: "The " + description + " should be named.",
			}, func() []rule.RuleSuggestion {
				name := suggestionName(ctx, node)
				if name == "" {
					return nil
				}
				fixes := addName(ctx.SourceFile, node, statement, name)
				if len(fixes) == 0 {
					return nil
				}
				return []rule.RuleSuggestion{{
					Message:  rule.RuleMessage{Id: "no-anonymous-default-export/suggestion", Description: "Name it as `" + name + "`."},
					FixesArr: fixes,
				}}
			})
		}
		declaration := func(node *ast.Node) {
			if ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) && ast.HasSyntacticModifier(node, ast.ModifierFlagsDefault) {
				report(node, node)
			}
		}
		return rule.RuleListeners{
			ast.KindClassDeclaration:    declaration,
			ast.KindFunctionDeclaration: declaration,
			ast.KindExportAssignment: func(node *ast.Node) {
				if !node.AsExportAssignment().IsExportEquals {
					report(node.Expression(), node)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				if !ast.IsAssignmentExpression(node, false) {
					return
				}
				statement := utils.ESTreeParent(node)
				if statement == nil || statement.Kind != ast.KindExpressionStatement {
					return
				}
				assignment := node.AsBinaryExpression()
				left := utils.ESTreeRuntimeExpression(assignment.Left)
				if left.Kind == ast.KindIdentifier && left.Text() == "exports" {
					report(assignment.Right, statement)
					return
				}
				if left.Kind != ast.KindPropertyAccessExpression || ast.IsOptionalChain(left) {
					return
				}
				member := left.AsPropertyAccessExpression()
				object := utils.ESTreeRuntimeExpression(member.Expression)
				if object.Kind == ast.KindIdentifier && object.Text() == "module" && member.Name().Kind == ast.KindIdentifier && member.Name().Text() == "exports" {
					report(assignment.Right, statement)
				}
			},
		}
	},
}

func classHeadRange(source *ast.SourceFile, node *ast.Node) core.TextRange {
	start := utils.FindFunctionKeywordPos(source, node)
	// Members starts immediately after the class body's opening brace, even
	// when the heritage expression contains a nested class or object literal.
	brace, _ := utils.TokenBeforePosition(source, node.ClassLikeData().Members.Pos())
	last, _ := utils.TokenBeforePosition(source, brace.Start)
	return core.NewTextRange(start, last.End)
}

func suggestionName(ctx rule.RuleContext, node *ast.Node) string {
	filename := ctx.SourceFile.FileName()
	if filename == "<input>" || filename == "<text>" {
		return ""
	}
	base, _, _ := strings.Cut(tspath.GetBaseFileName(filename), ".")
	name := unicornutil.CamelCase(base)
	if !scanner.IsIdentifierText(name, core.LanguageVariantStandard) {
		return ""
	}
	if ast.IsClassLike(node) {
		first, size := utf8.DecodeRuneInString(name)
		// Upstream upperFirst reads one UTF-16 code unit, leaving astral letters intact.
		if first <= 0xFFFF {
			name = ecmascript.StringToUpperCase(name[:size]) + name[size:]
		}
	}
	if isReservedName(name) {
		name += "_"
	}
	// No scope can conflict with a name absent from both the file and globals.
	if !ctx.SourceFile.HasIdentifier(name) {
		access := ctx.Globals.Access(name)
		if access != utils.GlobalAccessReadonly && access != utils.GlobalAccessWritable {
			return name
		}
	}

	manager := scopeanalysis.References(ctx, nil)
	root := manager.Acquire(node)
	used := map[string]bool{}
	for current := root; current != nil; current = current.Parent {
		for name := range current.ByName {
			used[name] = true
		}
	}
	// Only the exported value's descendants participate; sibling scopes do not.
	inside := map[*scope.Scope]bool{root: true}
	for _, current := range manager.Scopes {
		if current != root && !inside[current.Parent] {
			continue
		}
		inside[current] = true
		for name := range current.ByName {
			used[name] = true
		}
		for _, reference := range current.References {
			if reference.Resolved() == nil {
				used[reference.Identifier.Text()] = true
			}
		}
	}
	for used[name] || ctx.Globals.Access(name) == utils.GlobalAccessReadonly || ctx.Globals.Access(name) == utils.GlobalAccessWritable {
		name += "_"
	}
	return name
}

// Unicorn also reserves several TypeScript contextual words for suggestions.
func isReservedName(name string) bool {
	switch name {
	case "await", "break", "case", "catch", "class", "const", "continue", "debugger", "default", "delete", "do", "else", "enum", "export", "extends", "false", "finally", "for", "function", "if", "import", "in", "instanceof", "new", "null", "return", "super", "switch", "this", "throw", "true", "try", "typeof", "var", "void", "while", "with", "yield",
		"implements", "interface", "let", "package", "private", "protected", "public", "static", "arguments", "eval", "globalThis", "Infinity", "NaN", "undefined",
		"as", "any", "boolean", "constructor", "declare", "get", "module", "require", "number", "set", "string", "symbol", "type", "from", "of":
		return true
	}
	return false
}

func addName(source *ast.SourceFile, node, statement *ast.Node, name string) []rule.RuleFix {
	if ast.IsClassLike(node) {
		start := node.Pos()
		if modifiers := node.Modifiers(); modifiers != nil {
			start = modifiers.End()
		}
		keyword := scanner.GetRangeOfTokenAtPosition(source, start)
		return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(keyword.End(), keyword.End()), " "+name)}
	}
	if node.Kind != ast.KindArrowFunction {
		pos := utils.GetFunctionHeadLoc(source, node).End()
		suffix := " "
		if parameters := node.TypeParameterList(); parameters != nil {
			// Upstream inserts before the first `(`, which produces invalid
			// TypeScript for generic functions. Name them before `<T>` instead.
			var ok bool
			pos, _, ok = utils.RangeEnclosingDelimiters(source.Text(), parameters.Pos(), parameters.End(), '<', '>')
			if !ok {
				return nil
			}
			suffix = ""
		}
		prefix := ""
		if source.Text()[pos-1] != ' ' {
			prefix = " "
		}
		return []rule.RuleFix{rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), prefix+name+suffix)}
	}

	switch statement.Parent.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock, ast.KindCaseClause, ast.KindDefaultClause:
		// These positions accept the added const declaration and export statement.
	default:
		// Unlike upstream, do not suggest invalid declarations in unbraced
		// conditional, loop or labeled statement bodies. Adding braces first
		// makes the normal suggestion available.
		return nil
	}

	outer := node
	for outer.Parent != nil && (outer.Parent.Kind == ast.KindParenthesizedExpression || utils.IsJSDocTypeCastWrapper(outer.Parent)) {
		outer = outer.Parent
	}
	arrowRange := utils.TrimNodeTextRange(source, outer)
	statementRange := utils.TrimNodeTextRange(source, statement)
	text := source.Text()
	linebreak := "\n"
	for pos, char := range text {
		if ecmascript.IsLineTerminator(char) {
			linebreak = ecmascript.LineTerminatorSequenceAt(text, pos)
			break
		}
	}
	before := linebreak + text[statementRange.Pos():arrowRange.Pos()]
	last, _ := utf8.DecodeLastRuneInString(before)
	if !ecmascript.IsWhiteSpaceOrLineTerminator(last) {
		before += " "
	}
	after := text[arrowRange.End():statementRange.End()]
	if !strings.HasSuffix(after, ";") {
		after += ";"
	}
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(statementRange.Pos(), arrowRange.Pos()), "const "+name+" = "),
		rule.RuleFixReplaceRange(core.NewTextRange(arrowRange.End(), statementRange.End()), ";"),
		rule.RuleFixReplaceRange(core.NewTextRange(statementRange.End(), statementRange.End()), before+name+after),
	}
}
