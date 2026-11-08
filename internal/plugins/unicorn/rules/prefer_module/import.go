package prefer_module

import (
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func importFixes(ctx rule.RuleContext, identifier *ast.Node) []rule.RuleFix {
	call := utils.ESTreeParent(identifier)
	if call == nil || call.Kind != ast.KindCallExpression || call.AsCallExpression().QuestionDotToken != nil ||
		utils.ESTreeCallCallee(call.Expression()) != identifier || len(call.Arguments()) != 1 {
		return nil
	}
	// An import cannot retain a call's type arguments. Upstream leaves them in
	// the suggestion, producing invalid syntax; keep the diagnostic without it.
	if call.AsCallExpression().TypeArguments != nil {
		return nil
	}
	source := utils.ESTreeRuntimeExpression(call.Arguments()[0])
	value, ok := utils.GetStaticStringLiteralValue(source)
	if !ok {
		return nil
	}
	text := utils.TrimmedNodeText(ctx.SourceFile, source)
	if source.Kind != ast.KindStringLiteral {
		text = quoteModuleSource(value)
	}
	attributes := ""
	json := strings.HasSuffix(value, ".json")
	if json {
		quote := text[:1]
		attributes = " with {type: " + quote + "json" + quote + "}"
	}
	parent := utils.ESTreeParent(call)
	sideEffect := parent != nil && parent.Kind == ast.KindExpressionStatement && parent.Parent.Kind == ast.KindSourceFile
	var declaration, binding *ast.Node
	if !sideEffect {
		if parent == nil || parent.Kind != ast.KindVariableDeclaration {
			return nil
		}
		declarator := parent.AsVariableDeclaration()
		binding = declarator.Name()
		if utils.ESTreeRuntimeExpression(declarator.Initializer) != call || declarator.Type != nil || !importBinding(binding, json) {
			return nil
		}
		list := parent.Parent
		if list.Kind != ast.KindVariableDeclarationList || list.Flags&ast.NodeFlagsConst == 0 || len(list.AsVariableDeclarationList().Declarations.Nodes) != 1 {
			return nil
		}
		declaration = list.Parent
		if declaration.Kind != ast.KindVariableStatement || declaration.Parent.Kind != ast.KindSourceFile || ast.HasSyntacticModifier(declaration, ast.ModifierFlagsExport) {
			return nil
		}
	}

	// Edit individual tokens so comments and the source's original spacing survive.
	opening, ok := utils.TokenAtOrAfter(ctx.SourceFile, call.Expression().End())
	if !ok || opening.Kind != ast.KindOpenParenToken {
		return nil
	}
	closing, ok := utils.TokenBeforePosition(ctx.SourceFile, call.End())
	if !ok || closing.Kind != ast.KindCloseParenToken {
		return nil
	}
	// Likewise, a trailing argument comma would survive as an invalid import.
	if previous, ok := utils.TokenBeforePosition(ctx.SourceFile, closing.Start); ok && previous.Kind == ast.KindCommaToken {
		return nil
	}
	fixes := []rule.RuleFix{
		rule.RuleFixReplace(ctx.SourceFile, source, text),
		rule.RuleFixReplaceRange(closing.Range(), attributes),
	}
	if sideEffect {
		fixes = append(fixes, rule.RuleFixReplace(ctx.SourceFile, identifier, "import"), rule.RuleFixReplaceRange(opening.Range(), " "))
	} else {
		keyword, _ := utils.TokenAtOrAfter(ctx.SourceFile, declaration.Pos())
		equal, _ := utils.TokenAtOrAfter(ctx.SourceFile, binding.End())
		fixes = append(fixes,
			rule.RuleFixReplaceRange(keyword.Range(), "import"),
			rule.RuleFixRemoveRange(core.NewTextRange(binding.End(), spacesEnd(ctx.SourceFile, binding.End()))),
			rule.RuleFixReplaceRange(core.NewTextRange(equal.Start, spacesEnd(ctx.SourceFile, equal.End)), " from "),
			rule.RuleFixRemove(ctx.SourceFile, identifier),
			rule.RuleFixRemoveRange(opening.Range()),
		)
		if binding.Kind == ast.KindObjectBindingPattern {
			for _, element := range binding.AsBindingPattern().Elements.Nodes {
				key := element.AsBindingElement().PropertyName
				if key == nil {
					continue
				}
				colon, _ := utils.TokenAtOrAfter(ctx.SourceFile, key.End())
				fixes = append(fixes,
					rule.RuleFixRemoveRange(core.NewTextRange(key.End(), spacesEnd(ctx.SourceFile, key.End()))),
					rule.RuleFixReplaceRange(core.NewTextRange(colon.Start, spacesEnd(ctx.SourceFile, colon.End)), " as "),
				)
			}
		}
	}
	for _, node := range []*ast.Node{identifier, call, source} {
		fixes = append(fixes, removeParentheses(ctx.SourceFile, node)...)
	}
	return fixes
}

func importBinding(binding *ast.Node, json bool) bool {
	if binding.Kind == ast.KindIdentifier {
		return true
	}
	if json || binding.Kind != ast.KindObjectBindingPattern {
		return false
	}
	for _, node := range binding.AsBindingPattern().Elements.Nodes {
		element := node.AsBindingElement()
		if element.DotDotDotToken != nil || element.Initializer != nil || element.Name().Kind != ast.KindIdentifier ||
			(element.PropertyName != nil && element.PropertyName.Kind != ast.KindIdentifier) {
			return false
		}
	}
	return true
}

// Static templates have no raw string token to reuse. Match Unicorn's
// quote-js-string spelling, including control characters and lone surrogates.
// The tsgo printer uses different escapes for NUL, DEL, and U+0085; a local
// adapter keeps suggestion output identical without introducing a printer API.
func quoteModuleSource(value string) string {
	var result strings.Builder
	result.WriteByte('\'')
	for len(value) > 0 {
		character, size := ecmascript.DecodeStringRune(value)
		value = value[size:]
		switch character {
		case '\\', '\'':
			result.WriteByte('\\')
			result.WriteRune(character)
		case '\n':
			result.WriteString(`\n`)
		case '\r':
			result.WriteString(`\r`)
		case '\t':
			result.WriteString(`\t`)
		case '\b':
			result.WriteString(`\b`)
		case '\f':
			result.WriteString(`\f`)
		case '\v':
			result.WriteString(`\v`)
		default:
			if character < 0x20 || character == 0x7f || character == 0x2028 || character == 0x2029 || (character >= 0xd800 && character <= 0xdfff) {
				result.WriteString(`\u{` + strconv.FormatUint(uint64(character), 16) + `}`)
			} else {
				result.WriteRune(character)
			}
		}
	}
	result.WriteByte('\'')
	return result.String()
}
