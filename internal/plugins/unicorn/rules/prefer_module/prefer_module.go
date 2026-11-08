package prefer_module

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

func identifierMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{Id: "error/identifier", Description: "Do not use \"" + name + "\".", Data: map[string]string{"name": name}}
}

// https://github.com/sindresorhus/eslint-plugin-unicorn/blob/v79.0.0/rules/prefer-module.js
var PreferModuleRule = rule.Rule{
	Name:   "unicorn/prefer-module",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		filename := ecmascript.StringToLowerCase(ctx.SourceFile.FileName())
		if strings.HasSuffix(filename, ".cjs") {
			return nil
		}
		isMJS := strings.HasSuffix(filename, ".mjs")
		return rule.RuleListeners{
			ast.KindExpressionStatement: func(node *ast.Node) {
				// Espree does not mark static-block strings as directives, while
				// typescript-eslint does. The shared helper also covers namespaces.
				isDirective := utils.IsDirectivePrologueStatement(node)
				if !ast.IsInJSFile(node) {
					isDirective = utils.IsDirectivePrologueStatementIncludingClassStaticBlocks(node)
				}
				if !isDirective {
					return
				}
				// ESTree's directive value is the raw text between the quotes.
				text := utils.TrimmedNodeText(ctx.SourceFile, node.Expression())
				if text != "\"use strict\"" && text != "'use strict'" {
					return
				}
				message := rule.RuleMessage{Id: "error/use-strict-directive", Description: "Do not use \"use strict\" directive."}
				fix := func() []rule.RuleFix {
					return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(
						utils.TrimNodeTextRange(ctx.SourceFile, node).Pos(), spacesEnd(ctx.SourceFile, node.End()),
					))}
				}
				if isMJS {
					ctx.ReportNodeWithDeferredFixes(node, message, fix)
				} else {
					ctx.ReportNodeWithDeferredSuggestions(node, message, func() []rule.RuleSuggestion {
						return []rule.RuleSuggestion{{Message: rule.RuleMessage{Id: "suggestion/use-strict-directive", Description: "Remove \"use strict\" directive."}, FixesArr: fix()}}
					})
				}
			},
			ast.KindReturnStatement: func(node *ast.Node) {
				if ast.GetContainingFunction(node) == nil {
					token, ok := utils.TokenAtOrAfter(ctx.SourceFile, node.Pos())
					if ok {
						ctx.ReportRange(token.Range(), rule.RuleMessage{Id: "error/global-return", Description: "\"return\" should be used inside a function."})
					}
				}
			},
			ast.KindIdentifier: func(node *ast.Node) {
				name := node.Text()
				switch name {
				case "exports", "require", "module", "__filename", "__dirname":
				default:
					return
				}
				// JSX names are JSXIdentifier nodes in ESTree, including both
				// parts of a namespace such as <module:exports />.
				if utils.IsInJsxTagName(node) || ast.IsJsxNamespacedName(node.Parent) ||
					!ctx.Globals.Access(name).IsDeclared() || !ctx.Refs.IsGlobalReference(node) {
					return
				}
				ctx.ReportNodeWithDeferredSuggestions(node, identifierMessage(name), func() []rule.RuleSuggestion {
					switch name {
					case "__filename", "__dirname":
						return pathSuggestions(ctx, node)
					case "require":
						if fixes := importFixes(ctx, node); len(fixes) > 0 {
							return []rule.RuleSuggestion{{Message: rule.RuleMessage{Id: "suggestion/import", Description: "Switch to `import`."}, FixesArr: fixes}}
						}
					case "exports", "module":
						if fixes := exportFixes(ctx, node); len(fixes) > 0 {
							return []rule.RuleSuggestion{{Message: rule.RuleMessage{Id: "suggestion/export", Description: "Switch to `export`."}, FixesArr: fixes}}
						}
					}
					return nil
				})
			},
		}
	},
}

func spacesEnd(file *ast.SourceFile, position int) int {
	return ecmascript.SkipLeadingWhitespace(file.Text(), position, len(file.Text()))
}

// Parentheses are absent from ESTree. Remove their tokens individually to
// preserve whitespace and comments around the expression, as upstream does.
func removeParentheses(file *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	var fixes []rule.RuleFix
	for node.Parent != nil {
		parent := node.Parent
		if !ast.IsParenthesizedExpression(parent) {
			if utils.IsJSDocTypeCastWrapper(parent) {
				node = parent
				continue
			}
			break
		}
		node = parent
		opening, _ := utils.TokenAtOrAfter(file, node.Pos())
		closing, _ := utils.TokenBeforePosition(file, node.End())
		fixes = append(fixes, rule.RuleFixRemoveRange(opening.Range()), rule.RuleFixRemoveRange(closing.Range()))
	}
	return fixes
}

func pathSuggestions(ctx rule.RuleContext, node *ast.Node) []rule.RuleSuggestion {
	// Replacing a type name or type query with import.meta is not valid syntax.
	if scope.InTypePosition(node) || ast.IsPartOfTypeQuery(node) {
		return nil
	}
	kind := strings.TrimPrefix(node.Text(), "__")
	member := "import.meta." + kind
	call := "url.fileURLToPath(import.meta.url)"
	if kind == "dirname" {
		call = "path.dirname(" + call + ")"
	}
	result := make([]rule.RuleSuggestion, 0, 2)
	for index, replacement := range []string{member, call} {
		// A call expression cannot be an assignment or update target.
		if index == 1 && utils.IsWriteReference(node) {
			continue
		}
		id := "suggestion/import-meta-" + kind
		description := "Replace `" + node.Text() + "` with `" + member + "`."
		if index == 1 {
			id = "suggestion/import-meta-url-to-" + kind
			description = "Replace `" + node.Text() + "` with `…(import.meta.url)`."
		}
		if (index == 1 || !ast.IsInJSFile(node)) && inConstructorCallee(node) {
			replacement = "(" + replacement + ")"
		}
		parent := node.Parent
		if parent != nil && parent.Kind == ast.KindShorthandPropertyAssignment {
			replacement = node.Text() + ": " + replacement
		} else if parent != nil && parent.Kind == ast.KindExportSpecifier && parent.PropertyName() == nil {
			replacement += " as " + node.Text()
		}
		result = append(result, rule.RuleSuggestion{
			Message:  rule.RuleMessage{Id: id, Description: description},
			FixesArr: []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, replacement)},
		})
	}
	return result
}

func inConstructorCallee(node *ast.Node) bool {
	for node.Parent != nil {
		parent := node.Parent
		switch parent.Kind {
		case ast.KindNewExpression:
			return parent.Expression() == node
		case ast.KindTaggedTemplateExpression:
			if parent.AsTaggedTemplateExpression().Tag != node {
				return false
			}
			node = parent
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
			ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindTypeAssertionExpression, ast.KindNonNullExpression,
			ast.KindExpressionWithTypeArguments:
			if parent.Expression() != node {
				return false
			}
			node = parent
		default:
			return false
		}
	}
	return false
}

func topLevelAssignment(node *ast.Node) *ast.Node {
	parent := utils.ESTreeParent(node)
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return nil
	}
	assignment := parent.AsBinaryExpression()
	statement := utils.ESTreeParent(parent)
	if assignment.OperatorToken.Kind != ast.KindEqualsToken || utils.ESTreeRuntimeExpression(assignment.Left) != node ||
		statement == nil || statement.Kind != ast.KindExpressionStatement || statement.Parent.Kind != ast.KindSourceFile {
		return nil
	}
	return parent
}

func exportFixes(ctx rule.RuleContext, node *ast.Node) []rule.RuleFix {
	if node.Text() == "module" {
		parent := utils.ESTreeParent(node)
		if !dotMember(parent, node) || parent.Name().Text() != "exports" {
			return nil
		}
		node = parent
	}
	if assignment := topLevelAssignment(node); assignment != nil {
		// Replacing module.exports also replaces comments between its tokens.
		// Upstream drops those comments; retain the diagnostic without an edit.
		if commentsInside(ctx, node) {
			return nil
		}
		equal := assignment.AsBinaryExpression().OperatorToken
		fixes := []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(utils.TrimNodeTextRange(ctx.SourceFile, node).Pos(), spacesEnd(ctx.SourceFile, node.End())), "export default "),
			rule.RuleFixRemoveRange(core.NewTextRange(utils.TrimNodeTextRange(ctx.SourceFile, equal).Pos(), spacesEnd(ctx.SourceFile, equal.End()))),
		}
		fixes = append(fixes, removeParentheses(ctx.SourceFile, node)...)
		return append(fixes, removeParentheses(ctx.SourceFile, assignment)...)
	}
	member := utils.ESTreeParent(node)
	if !dotMember(member, node) {
		return nil
	}
	assignment := topLevelAssignment(member)
	if assignment == nil {
		return nil
	}
	local := utils.ESTreeRuntimeExpression(assignment.AsBinaryExpression().Right)
	if local.Kind != ast.KindIdentifier {
		return nil
	}
	if commentsInside(ctx, assignment) {
		return nil
	}
	fixes := []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, assignment, "export {"+local.Text()+" as "+member.Name().Text()+"}")}
	return append(fixes, removeParentheses(ctx.SourceFile, assignment)...)
}

func commentsInside(ctx rule.RuleContext, node *ast.Node) bool {
	span := utils.TrimNodeTextRange(ctx.SourceFile, node)
	return utils.HasCommentInSpan(ctx.Comments.All(), span.Pos(), span.End())
}

func dotMember(node, object *ast.Node) bool {
	return node != nil && node.Kind == ast.KindPropertyAccessExpression && node.AsPropertyAccessExpression().QuestionDotToken == nil &&
		node.Name().Kind == ast.KindIdentifier && utils.ESTreeRuntimeExpression(node.Expression()) == object
}
