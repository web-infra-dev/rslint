package prefer_string_raw

import (
	"iter"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

var preferRaw = rule.RuleMessage{
	Id:          "prefer-string-raw",
	Description: "`String.raw` should be used to avoid escaping `\\`.",
}

var useRawSuggestion = rule.RuleMessage{
	Id:          "prefer-string-raw/suggestion",
	Description: "Use `String.raw`. This may widen the string literal type to `string`.",
}

var unnecessaryRaw = rule.RuleMessage{
	Id:          "unnecessary-string-raw",
	Description: "Using `String.raw` is unnecessary as the string does not contain any `\\`.",
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-string-raw.js
var PreferStringRawRule = rule.Rule{
	Name:   "unicorn/prefer-string-raw",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		checkTemplate := func(node *ast.Node) { checkTemplateLiteral(ctx, node) }
		return rule.RuleListeners{
			ast.KindStringLiteral:                 func(node *ast.Node) { checkStringLiteral(ctx, node) },
			ast.KindNoSubstitutionTemplateLiteral: checkTemplate,
			ast.KindTemplateExpression:            checkTemplate,
			ast.KindTaggedTemplateExpression:      func(node *ast.Node) { checkRawTag(ctx, node) },
		}
	},
}

func checkStringLiteral(ctx rule.RuleContext, node *ast.Node) {
	raw := scanner.GetSourceTextOfNodeFromSourceFile(ctx.SourceFile, node, false)
	if len(raw) < 2 || raw[len(raw)-2] == '\\' || !strings.Contains(raw, `\\`) ||
		strings.ContainsRune(raw, '`') || strings.Contains(raw, "${") || strings.ContainsFunc(raw, ecmascript.IsLineTerminator) ||
		!onlyBackslashAndQuoteEscapes(raw[1:len(raw)-1], raw[0]) {
		return
	}
	reportConversion(ctx, node, func() []rule.RuleFix {
		fixes := unicornutil.SpaceAroundKeywordFixes(ctx.SourceFile, node)
		return append(fixes, rule.RuleFixReplace(ctx.SourceFile, node, "String.raw`"+node.Text()+"`"))
	})
}

func checkTemplateLiteral(ctx rule.RuleContext, node *ast.Node) {
	if ast.IsTaggedTemplateExpression(node.Parent) && node.Parent.AsTaggedTemplateExpression().Template == node {
		return
	}
	hasEscapes := false
	for part := range templateParts(node) {
		raw := templatePartText(ctx.SourceFile, part)
		if !strings.Contains(raw, `\`) {
			continue
		}
		if strings.HasSuffix(raw, `\`) || !onlyBackslashAndQuoteEscapes(raw, 0) {
			return
		}
		hasEscapes = true
	}
	if hasEscapes {
		reportConversion(ctx, node, func() []rule.RuleFix { return addRawTagFixes(ctx.SourceFile, node) })
	}
}

func checkRawTag(ctx rule.RuleContext, node *ast.Node) {
	tagged := node.AsTaggedTemplateExpression()
	tag := utils.ESTreeRuntimeExpression(tagged.Tag)
	if !ast.IsPropertyAccessExpression(tag) || ast.IsOptionalChain(tag) || ast.IsOptionalChain(node) ||
		!ast.IsIdentifier(tag.Name()) || tag.Name().Text() != "raw" {
		return
	}
	object := utils.ESTreeRuntimeExpression(tag.Expression())
	if !ast.IsIdentifier(object) || object.Text() != "String" || object.Flags&ast.NodeFlagsInWithStatement != 0 {
		return
	}
	for part := range templateParts(tagged.Template) {
		if strings.Contains(templatePartText(ctx.SourceFile, part), `\`) {
			return
		}
	}
	if !ctx.Globals.Access("String").IsDeclared() || !unicornutil.IsGlobalReference(ctx, object) {
		return
	}
	ctx.ReportNodeWithDeferredFixes(tag, unnecessaryRaw, func() []rule.RuleFix {
		return removeRawTagFixes(ctx, node, tag)
	})
}

func templateParts(node *ast.Node) iter.Seq[*ast.Node] {
	return func(yield func(*ast.Node) bool) {
		if node.Kind == ast.KindNoSubstitutionTemplateLiteral {
			yield(node)
			return
		}
		template := node.AsTemplateExpression()
		if yield(template.Head) {
			for _, span := range template.TemplateSpans.Nodes {
				if !yield(span.AsTemplateSpan().Literal) {
					return
				}
			}
		}
	}
}

func templatePartText(source *ast.SourceFile, part *ast.Node) string {
	if part.Kind != ast.KindNoSubstitutionTemplateLiteral {
		return part.RawText()
	}
	textRange := templatePartRange(source, part)
	return source.Text()[textRange.Pos():textRange.End()]
}

func templatePartRange(source *ast.SourceFile, part *ast.Node) core.TextRange {
	// The opening delimiter is a backtick or `}`. The closing delimiter is
	// either `${` (head and middle) or a backtick (tail and simple template).
	end := part.End() - 1
	if part.Kind == ast.KindTemplateHead || part.Kind == ast.KindTemplateMiddle {
		end--
	}
	return core.NewTextRange(scanner.GetTokenPosOfNode(part, source, false)+1, end)
}

// For valid literals, any other escape changes the cooked value when moved
// into a raw template. A template part has no quote escape to preserve.
func onlyBackslashAndQuoteEscapes(raw string, quote byte) bool {
	for {
		_, rest, escaped := strings.Cut(raw, `\`)
		if !escaped {
			return true
		}
		if len(rest) == 0 || (rest[0] != '\\' && (quote == 0 || rest[0] != quote)) {
			return false
		}
		raw = rest[1:]
	}
}

func reportConversion(ctx rule.RuleContext, node *ast.Node, build func() []rule.RuleFix) {
	parent := utils.ESTreeParent(node)
	if requiresStringLiteral(node, parent) || isInlineSnapshot(node, parent) {
		return
	}
	if ast.IsAssertionExpression(parent) && ast.IsConstTypeReference(ast.SkipTypeParentheses(parent.Type())) {
		return
	}
	canFix := true
	if ctx.SourceFile.ScriptKind == core.ScriptKindTS || ctx.SourceFile.ScriptKind == core.ScriptKindTSX {
		var binding *ast.Node
		if ast.IsVariableDeclaration(parent) && ast.IsIdentifier(parent.Name()) &&
			utils.ESTreeRuntimeExpression(parent.AsVariableDeclaration().Initializer) == node {
			binding = parent.Name()
		}
		if checker := ctx.TypeChecker; checker != nil {
			stringType := checker.GetStringType()
			if checker.GetTypeAtLocation(node) != stringType &&
				(binding == nil || checker.GetNonNullableType(checker.GetTypeAtLocation(binding)) != stringType) {
				return
			}
		} else {
			canFix = binding != nil && isWidenedBinding(parent)
		}
	}

	edits := func() []rule.RuleFix {
		if ctx.LanguageOptions.EffectiveECMAVersion() < 2015 || node.Flags&ast.NodeFlagsInWithStatement != 0 ||
			!ctx.Globals.Access("String").IsDeclared() || !ctx.Refs.IsGlobalNameReference(node, "String", ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) {
			return nil
		}
		return build()
	}
	if canFix {
		ctx.ReportNodeWithDeferredFixes(node, preferRaw, edits)
		return
	}
	ctx.ReportNodeWithDeferredSuggestions(node, preferRaw, func() []rule.RuleSuggestion {
		fixes := edits()
		if len(fixes) == 0 {
			return nil
		}
		return []rule.RuleSuggestion{{Message: useRawSuggestion, FixesArr: fixes}}
	})
}

func isWidenedBinding(declaration *ast.Node) bool {
	if annotation := declaration.Type(); annotation != nil {
		return ast.SkipTypeParentheses(annotation).Kind == ast.KindStringKeyword
	}
	kind := utils.GetVarDeclListKind(declaration.Parent)
	return kind == "let" || kind == "var"
}

func requiresStringLiteral(node, parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindExpressionStatement:
		if ast.IsInJSFile(node) {
			return utils.IsDirectivePrologueStatement(parent)
		}
		return utils.IsDirectivePrologueStatementIncludingClassStaticBlocks(parent)
	case ast.KindImportDeclaration, ast.KindExportDeclaration, ast.KindImportAttribute,
		ast.KindImportSpecifier, ast.KindExportSpecifier, ast.KindNamespaceExport,
		ast.KindJsxAttribute, ast.KindEnumMember, ast.KindModuleDeclaration,
		ast.KindExternalModuleReference, ast.KindLiteralType:
		return true
	case ast.KindComputedPropertyName:
		member := parent.Parent
		return member.Kind == ast.KindPropertySignature || member.Kind == ast.KindMethodSignature || ast.HasAbstractModifier(member)
	default:
		return parent.Name() == node || parent.PropertyName() == node
	}
}

func isInlineSnapshot(node, parent *ast.Node) bool {
	one := 1
	match, ok := unicornutil.MatchDotMethodCall(parent, unicornutil.DotMethodCallOptions{
		Method: "toMatchInlineSnapshot", ArgumentsLength: &one, RejectSpreadElement: true,
	})
	if !ok || utils.ESTreeRuntimeExpression(parent.Arguments()[0]) != node {
		return false
	}
	expectCall := utils.ESTreeRuntimeExpression(match.Object)
	if !ast.IsCallExpression(expectCall) || ast.IsOptionalChain(expectCall) {
		return false
	}
	callee := utils.ESTreeCallCallee(expectCall.Expression())
	args := expectCall.Arguments()
	return ast.IsIdentifier(callee) && callee.Text() == "expect" && len(args) == 1 && args[0].Kind != ast.KindSpreadElement
}
