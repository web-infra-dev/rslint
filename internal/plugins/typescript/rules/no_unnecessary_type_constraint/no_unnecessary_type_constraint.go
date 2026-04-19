package no_unnecessary_type_constraint

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/web-infra-dev/rslint/internal/rule"
)

func buildUnnecessaryConstraintMessage(name, constraint string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryConstraint",
		Description: fmt.Sprintf("Constraining the generic type `%s` to `%s` does nothing and is unnecessary.", name, constraint),
	}
}

func buildRemoveUnnecessaryConstraintMessage(constraint string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeUnnecessaryConstraint",
		Description: fmt.Sprintf("Remove the unnecessary `%s` constraint.", constraint),
	}
}

var disambiguationExtensions = []string{tspath.ExtensionCts, tspath.ExtensionMts, tspath.ExtensionTsx}

var NoUnnecessaryTypeConstraintRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-constraint",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		needsDisambiguation := tspath.FileExtensionIsOneOf(ctx.SourceFile.FileName(), disambiguationExtensions)
		text := ctx.SourceFile.Text()

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

				typeParam := node.AsTypeParameter()
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

				inArrowFunction := parent.Kind == ast.KindArrowFunction

				addTrailingComma := false
				if inArrowFunction && needsDisambiguation && typeParam.DefaultType == nil {
					if len(parent.TypeParameters()) == 1 {
						nextPos := scanner.SkipTrivia(text, node.End())
						if nextPos >= len(text) || text[nextPos] != ',' {
							addTrailingComma = true
						}
					}
				}

				// Fix replaces ` extends <constraint>` (between name end and constraint end). This
				// assumes source order `name extends constraint`. Bail if the AST ever violates
				// that — better to skip than to emit a destructive fix.
				if nameNode.End() > typeParam.Constraint.End() {
					return
				}

				replacement := ""
				if addTrailingComma {
					replacement = ","
				}

				fixRange := core.NewTextRange(nameNode.End(), typeParam.Constraint.End())

				ctx.ReportNodeWithSuggestions(
					node,
					buildUnnecessaryConstraintMessage(nameNode.Text(), constraintName),
					rule.RuleSuggestion{
						Message:  buildRemoveUnnecessaryConstraintMessage(constraintName),
						FixesArr: []rule.RuleFix{rule.RuleFixReplaceRange(fixRange, replacement)},
					},
				)
			},
		}
	},
})
