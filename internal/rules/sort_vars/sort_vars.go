package sort_vars

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

//go:embed sort_vars.schema.json
var schemaJSON []byte

type options struct {
	ignoreCase bool
}

func parseOptions(raw []any) options {
	var opts options
	if len(raw) == 0 {
		return opts
	}
	if object, ok := raw[0].(map[string]any); ok {
		if ignoreCase, ok := object["ignoreCase"].(bool); ok {
			opts.ignoreCase = ignoreCase
		}
	}
	return opts
}

func sortableName(declaration *ast.Node, ignoreCase bool) string {
	name := declaration.AsVariableDeclaration().Name().AsIdentifier().Text
	if ignoreCase {
		return ecmascript.StringToLowerCase(name)
	}
	return name
}

func identifierDeclarations(list *ast.VariableDeclarationList) []*ast.Node {
	if list == nil || list.Declarations == nil {
		return nil
	}
	declarations := make([]*ast.Node, 0, len(list.Declarations.Nodes))
	for _, declaration := range list.Declarations.Nodes {
		variable := declaration.AsVariableDeclaration()
		if variable != nil && variable.Name() != nil && variable.Name().Kind == ast.KindIdentifier {
			declarations = append(declarations, declaration)
		}
	}
	return declarations
}

func buildFix(ctx rule.RuleContext, declarations []*ast.Node, ignoreCase bool) []rule.RuleFix {
	for _, declaration := range declarations {
		initializer := declaration.AsVariableDeclaration().Initializer
		if initializer != nil && !utils.IsESTreeLiteralKind(ast.SkipParentheses(initializer).Kind) {
			return nil
		}
	}

	ordered := append([]*ast.Node(nil), declarations...)
	// Upstream's comparator returns -1 for names that compare equal. V8's
	// stable sort consequently reverses each equal-name group; this insertion
	// sort spells that observable behavior out without violating sort.Interface.
	for i := 1; i < len(ordered); i++ {
		current := ordered[i]
		currentName := sortableName(current, ignoreCase)
		j := i
		for j > 0 && ecmascript.CompareStrings(currentName, sortableName(ordered[j-1], ignoreCase)) <= 0 {
			ordered[j] = ordered[j-1]
			j--
		}
		ordered[j] = current
	}

	text := ctx.SourceFile.Text()
	ranges := make([]core.TextRange, len(declarations))
	for i, declaration := range declarations {
		ranges[i] = utils.TrimNodeTextRange(ctx.SourceFile, declaration)
	}

	replacement := make([]byte, 0, ranges[len(ranges)-1].End()-ranges[0].Pos())
	for i, declaration := range ordered {
		r := utils.TrimNodeTextRange(ctx.SourceFile, declaration)
		replacement = append(replacement, text[r.Pos():r.End()]...)
		if i+1 < len(ranges) {
			replacement = append(replacement, text[ranges[i].End():ranges[i+1].Pos()]...)
		}
	}

	return []rule.RuleFix{rule.RuleFixReplaceRange(
		core.NewTextRange(ranges[0].Pos(), ranges[len(ranges)-1].End()),
		string(replacement),
	)}
}

var sortVarsMessage = rule.RuleMessage{
	Id:          "sortVars",
	Description: "Variables within the same declaration block should be sorted alphabetically.",
}

// https://eslint.org/docs/latest/rules/sort-vars
var SortVarsRule = rule.Rule{
	Name:   "sort-vars",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)
		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				declarations := identifierDeclarations(node.AsVariableDeclarationList())
				if len(declarations) < 2 {
					return
				}

				previous := declarations[0]
				fixed := false
				for _, declaration := range declarations[1:] {
					if ecmascript.CompareStrings(sortableName(declaration, opts.ignoreCase), sortableName(previous, opts.ignoreCase)) < 0 {
						attachFix := !fixed
						ctx.ReportNodeWithDeferredFixes(declaration, sortVarsMessage, func() []rule.RuleFix {
							if !attachFix {
								return nil
							}
							return buildFix(ctx, declarations, opts.ignoreCase)
						})
						fixed = true
						continue
					}
					previous = declaration
				}
			},
		}
	},
}
