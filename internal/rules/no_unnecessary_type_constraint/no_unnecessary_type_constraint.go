package no_unnecessary_type_constraint

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
)

var unnecessaryConstraints = map[ast.Kind]string{
	ast.KindAnyKeyword:     "any",
	ast.KindUnknownKeyword: "unknown",
}

func checkRequiresGenericDeclarationDisambiguation(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".cts", ".mts", ".tsx":
		return true
	default:
		return false
	}
}

func shouldAddTrailingComma(node *ast.Node, inArrowFunction bool, requiresDisambiguation bool, ctx rule.RuleContext) bool {
	if !inArrowFunction || !requiresDisambiguation {
		return false
	}

	// Only <T>() => {} would need trailing comma
	typeParam := node.AsTypeParameter()
	if typeParam.Parent == nil {
		return false
	}

	// Walk up to find the parent declaration that contains type parameters
	current := typeParam.Parent
	for current != nil {
		// Check if current node has TypeParameters method
		switch current.Kind {
		case ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindMethodDeclaration, ast.KindClassDeclaration, ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration, ast.KindCallSignature, ast.KindConstructSignature,
			ast.KindMethodSignature, ast.KindConstructorType, ast.KindFunctionType:

			typeParams := current.TypeParameters()
			if typeParams != nil && len(typeParams) == 1 {
				// Check if there's already a trailing comma
				nodeEnd := typeParam.End()
				// Find the next token after the type parameter
				for i := nodeEnd; i < len(ctx.SourceFile.Text()); i++ {
					char := ctx.SourceFile.Text()[i]
					if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
						if char == ',' {
							return false // Already has trailing comma
						}
						break
					}
				}
				return true
			}
			return false
		}
		current = current.Parent
	}

	return false
}

var NoUnnecessaryTypeConstraintRule = rule.Rule{
	Name: "no-unnecessary-type-constraint",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		requiresGenericDeclarationDisambiguation := checkRequiresGenericDeclarationDisambiguation(ctx.SourceFile.FileName())

		checkNode := func(node *ast.Node, inArrowFunction bool) {
			typeParam := node.AsTypeParameter()

			// At this point we know there's a constraint and it's unnecessary
			constraint, found := unnecessaryConstraints[typeParam.Constraint.Kind]
			if !found {
				// This should not happen since we already checked in the listener
				return
			}

			name := typeParam.Name()
			typeName := name.Text()

			// Create the fix
			var fixText string
			if shouldAddTrailingComma(node, inArrowFunction, requiresGenericDeclarationDisambiguation, ctx) {
				fixText = ","
			} else {
				fixText = ""
			}

			// Calculate the range to replace (from after the name to the end of the constraint)
			nameEnd := name.End()
			constraintEnd := typeParam.Constraint.End()
			fixRange := core.NewTextRange(nameEnd, constraintEnd)

			message := rule.RuleMessage{
				Id:          "unnecessaryConstraint",
				Description: fmt.Sprintf("Constraining the generic type `%s` to `%s` does nothing and is unnecessary.", typeName, constraint),
			}

			fix := rule.RuleFix{
				Range: fixRange,
				Text:  fixText,
			}

			ctx.ReportNodeWithFixes(node, message, fix)
		}

		return rule.RuleListeners{
			ast.KindTypeParameter: func(node *ast.Node) {
				typeParam := node.AsTypeParameter()

				// Only check type parameters that have constraints
				if typeParam.Constraint == nil {
					return
				}

				// Only check for unnecessary constraints (any or unknown)
				_, isUnnecessary := unnecessaryConstraints[typeParam.Constraint.Kind]
				if !isUnnecessary {
					return
				}

				// Check if this is in an arrow function
				parent := node.Parent
				inArrowFunction := false

				// Walk up the tree to find if we're in an arrow function
				current := parent
				for current != nil {
					if current.Kind == ast.KindArrowFunction {
						inArrowFunction = true
						break
					}
					// Stop if we hit a non-arrow function declaration
					if current.Kind == ast.KindFunctionDeclaration ||
						current.Kind == ast.KindFunctionExpression ||
						current.Kind == ast.KindMethodDeclaration ||
						current.Kind == ast.KindConstructor ||
						current.Kind == ast.KindGetAccessor ||
						current.Kind == ast.KindSetAccessor {
						break
					}
					current = current.Parent
				}

				checkNode(node, inArrowFunction)
			},
		}
	},
}
