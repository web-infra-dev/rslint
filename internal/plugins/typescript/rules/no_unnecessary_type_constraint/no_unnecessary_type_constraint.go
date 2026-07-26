package no_unnecessary_type_constraint

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnnecessaryConstraintMessage(name, constraint string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryConstraint",
		Description: "Constraining the generic type `" + name + "` to `" + constraint + "` does nothing and is unnecessary.",
	}
}

func buildRemoveUnnecessaryConstraintMessage(constraint string) rule.RuleMessage {
	var description string
	switch constraint {
	case "any":
		description = "Remove the unnecessary `any` constraint."
	case "unknown":
		description = "Remove the unnecessary `unknown` constraint."
	default:
		description = "Remove the unnecessary `" + constraint + "` constraint."
	}
	return rule.RuleMessage{
		Id:          "removeUnnecessaryConstraint",
		Description: description,
	}
}

var disambiguationExtensions = []string{tspath.ExtensionCts, tspath.ExtensionMts, tspath.ExtensionTsx}

var NoUnnecessaryTypeConstraintRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-constraint",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		needsDisambiguationResolved := false
		needsDisambiguation := false

		return rule.RuleListeners{
			ast.KindTypeParameter: func(node *ast.Node) {
				// Match typescript-eslint's selector:
				// `TSTypeParameterDeclaration > TSTypeParameter[constraint]`.
				// In tsgo, `infer U`, mapped-type `[P in K]`, and JSDoc `@template` also
				// surface as KindTypeParameter but have no TSTypeParameterDeclaration
				// analog, so upstream doesn't report them.
				parent := node.Parent
				if parent == nil {
					return
				}
				switch parent.Kind {
				case ast.KindInferType, ast.KindMappedType, ast.KindJSDocTemplateTag:
					return
				}

				typeParam := node.AsTypeParameterDeclaration()
				if typeParam == nil || typeParam.Constraint == nil {
					return
				}

				var constraintName string
				switch typeParam.Constraint.Kind {
				case ast.KindAnyKeyword:
					constraintName = "any"
				case ast.KindUnknownKeyword:
					constraintName = "unknown"
				default:
					return
				}

				nameNode := typeParam.Name()
				if nameNode == nil {
					return
				}

				reportRange := core.NewTextRange(
					scanner.GetTokenPosOfNode(node, ctx.SourceFile, false),
					node.End(),
				)
				ctx.ReportRangeWithDeferredSuggestions(
					reportRange,
					buildUnnecessaryConstraintMessage(nameNode.Text(), constraintName),
					func() []rule.RuleSuggestion {
						// The edit replaces ` extends <constraint>` between the name and
						// constraint. Keep the diagnostic if a malformed AST violates that
						// source order, but withhold its unsafe suggestion.
						if nameNode.End() > typeParam.Constraint.End() {
							return nil
						}

						replacement := ""
						if parent.Kind == ast.KindArrowFunction &&
							typeParam.DefaultType == nil &&
							len(parent.TypeParameters()) == 1 {
							if !needsDisambiguationResolved {
								needsDisambiguation = tspath.FileExtensionIsOneOf(
									ctx.SourceFile.FileName(),
									disambiguationExtensions,
								)
								needsDisambiguationResolved = true
							}
							if needsDisambiguation {
								text := ctx.SourceFile.Text()
								nextPos := scanner.SkipTrivia(text, node.End())
								if nextPos >= len(text) || text[nextPos] != ',' {
									replacement = ","
								}
							}
						}

						return []rule.RuleSuggestion{{
							Message: buildRemoveUnnecessaryConstraintMessage(constraintName),
							FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(
								core.NewTextRange(nameNode.End(), typeParam.Constraint.End()),
								replacement,
							)},
						}}
					},
				)
			},
		}
	},
})
