package prefer_string_raw

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func addRawTagFixes(source *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	fixes := unicornutil.SpaceAroundKeywordFixes(source, node)
	start := scanner.GetTokenPosOfNode(node, source, false)
	fixes = append(fixes, rule.RuleFixReplaceRange(core.NewTextRange(start, start+1), "String.raw`"))
	for part := range templateParts(node) {
		if strings.Contains(templatePartText(source, part), `\`) {
			fixes = append(fixes, rule.RuleFixReplaceRange(templatePartRange(source, part), part.Text()))
		}
	}
	return fixes
}

func removeRawTagFixes(ctx rule.RuleContext, node, tag *ast.Node) []rule.RuleFix {
	source := ctx.SourceFile
	tagged := node.AsTaggedTemplateExpression()
	tagRange := utils.TrimNodeTextRange(source, tag)
	if tagged.TypeArguments != nil || utils.HasCommentInSpan(ctx.Comments.All(), tagRange.Pos(), tagRange.End()) {
		return nil
	}

	template := tagged.Template
	templateRange := utils.TrimNodeTextRange(source, template)
	replacement := source.Text()[templateRange.Pos():templateRange.End()]
	if template.Kind == ast.KindNoSubstitutionTemplateLiteral && !strings.ContainsFunc(replacement, ecmascript.IsLineTerminator) &&
		!wouldIntroduceDirective(ctx, node) {
		replacement = "'" + strings.ReplaceAll(template.Text(), "'", `\'`) + "'"
	}

	fixes := keywordArgumentFixes(source, node, templateRange.Pos())
	if unicornutil.NeedsSemicolonBefore(source, node, replacement) {
		fixes = append(fixes, rule.RuleFixInsertBefore(source, node, ";"))
	}
	// Remove only the tag and its parentheses, retaining trivia between them.
	for wrapper := tag.Parent; wrapper != node; wrapper = wrapper.Parent {
		if ast.IsParenthesizedExpression(wrapper) {
			textRange := utils.TrimNodeTextRange(source, wrapper)
			fixes = append(fixes,
				rule.RuleFixRemoveRange(core.NewTextRange(textRange.Pos(), textRange.Pos()+1)),
				rule.RuleFixRemoveRange(core.NewTextRange(textRange.End()-1, textRange.End())),
			)
		}
	}
	return append(fixes, rule.RuleFixRemoveRange(tagRange), rule.RuleFixReplaceRange(templateRange, replacement))
}

func wouldIntroduceDirective(ctx rule.RuleContext, node *ast.Node) bool {
	statement := node.Parent
	if !ast.IsExpressionStatement(statement) {
		return false
	}
	container := statement.Parent
	switch container.Kind {
	case ast.KindSourceFile:
		if ctx.LanguageOptions.EffectiveSourceType() == "module" {
			return false
		}
	case ast.KindBlock:
		if !ast.IsFunctionLike(container.Parent) {
			return false
		}
	case ast.KindModuleBlock:
	default:
		return false
	}
	for _, preceding := range container.Statements() {
		if preceding == statement {
			return true
		}
		if !ast.IsPrologueDirective(preceding) {
			return false
		}
	}
	return false
}

// Removing the leading tag can expose a line terminator after return, throw,
// or yield, even when the tagged template begins a larger expression.
func keywordArgumentFixes(source *ast.SourceFile, node *ast.Node, templateStart int) []rule.RuleFix {
	start := utils.TrimNodeTextRange(source, node).Pos()
	if utils.IsSameLine(source, start, templateStart) {
		return nil
	}
	for expr := node; expr.Parent != nil; expr = expr.Parent {
		parent := expr.Parent
		switch parent.Kind {
		case ast.KindReturnStatement, ast.KindThrowStatement:
			return unicornutil.ReturnOrThrowParenthesesFixes(source, expr, templateStart)
		case ast.KindYieldExpression:
			if parent.AsYieldExpression().AsteriskToken == nil {
				return []rule.RuleFix{rule.RuleFixInsertBefore(source, node, "("), rule.RuleFixInsertAfter(node, ")")}
			}
			return nil
		}
		if utils.TrimNodeTextRange(source, parent).Pos() != start {
			return nil
		}
	}
	return nil
}
