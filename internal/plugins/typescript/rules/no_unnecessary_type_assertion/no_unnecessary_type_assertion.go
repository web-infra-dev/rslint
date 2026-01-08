package no_unnecessary_type_assertion

import (
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildContextuallyUnnecessaryMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "contextuallyUnnecessary",
		Description: "This assertion is unnecessary since the receiver accepts the original type of the expression.",
	}
}
func buildUnnecessaryAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryAssertion",
		Description: "This assertion is unnecessary since it does not change the type of the expression.",
	}
}

type NoUnnecessaryTypeAssertionOptions struct {
	// TODO(port): maybe typeOrValueSpecifier?
	TypesToIgnore               []string
	CheckLiteralConstAssertions bool
}

var NoUnnecessaryTypeAssertionRule = rule.CreateRule(rule.Rule{
	Name: "no-unnecessary-type-assertion",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		var typesToIgnore []string
		var checkLiteralConstAssertions bool

		if opts, ok := options.(NoUnnecessaryTypeAssertionOptions); ok {
			typesToIgnore = opts.TypesToIgnore
			checkLiteralConstAssertions = opts.CheckLiteralConstAssertions
		} else if optsList, ok := options.([]any); ok {
			// handle array of options (ESLint style)
			if len(optsList) > 0 {
				if firstOpt, ok := optsList[0].(map[string]any); ok {
					if v, ok := firstOpt["typesToIgnore"]; ok {
						if arr, ok := v.([]any); ok {
							for _, item := range arr {
								if s, ok := item.(string); ok {
									typesToIgnore = append(typesToIgnore, s)
								}
							}
						}
					}
					if v, ok := firstOpt["checkLiteralConstAssertions"]; ok {
						if b, ok := v.(bool); ok {
							checkLiteralConstAssertions = b
						}
					}
				}
			}
		} else if optsMap, ok := options.(map[string]any); ok {
			// handle single option object passed directly (just in case)
			if v, ok := optsMap["typesToIgnore"]; ok {
				if arr, ok := v.([]any); ok {
					for _, item := range arr {
						if s, ok := item.(string); ok {
							typesToIgnore = append(typesToIgnore, s)
						}
					}
				}
			}
			if v, ok := optsMap["checkLiteralConstAssertions"]; ok {
				if b, ok := v.(bool); ok {
					checkLiteralConstAssertions = b
				}
			}
		}

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		/**
		 * Returns true if there's a chance the variable has been used before a value has been assigned to it
		 */
		isPossiblyUsedBeforeAssigned := func(node *ast.Node) bool {
			declaration := utils.GetDeclaration(ctx.TypeChecker, node)
			if declaration == nil {
				// don't know what the declaration is for some reason, so just assume the worst
				return true
			}
			// non-strict mode doesn't care about used before assigned errors
			if !isStrictNullChecks {
				return false
			}
			// ignore class properties as they are compile time guarded
			// also ignore function arguments as they can't be used before defined
			if !ast.IsVariableDeclaration(declaration) {
				return false
			}

			decl := declaration.AsVariableDeclaration()

			// For var declarations, we need to check whether the node
			// is actually in a descendant of its declaration or not. If not,
			// it may be used before defined.

			// eg
			// if (Math.random() < 0.5) {
			//     var x: number  = 2;
			// } else {
			//     x!.toFixed();
			// }
			if ast.IsVariableDeclarationList(declaration.Parent) &&
				// var
				declaration.Parent.Flags == ast.NodeFlagsNone {
				// If they are not in the same file it will not exist.
				// This situation must not occur using before defined.
				declaratorScope := ast.GetEnclosingBlockScopeContainer(declaration)
				scope := ast.GetEnclosingBlockScopeContainer(node)

				parentScope := declaratorScope
				for {
					parentScope = ast.GetEnclosingBlockScopeContainer(parentScope)
					if parentScope == nil {
						break
					}
					if parentScope == scope {
						return true
					}
				}
			}

			if
			// is it `const x: number`
			decl.Initializer == nil &&
				decl.ExclamationToken == nil &&
				decl.Type != nil {
				// check if the defined variable type has changed since assignment
				declarationType := checker.Checker_getTypeFromTypeNode(ctx.TypeChecker, declaration.Type())
				t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
				if declarationType == t &&
					// `declare`s are never narrowed, so never skip them
					(!ast.IsVariableDeclarationList(declaration.Parent) || !ast.IsVariableStatement(declaration.Parent.Parent) || !utils.IncludesModifier(declaration.Parent.Parent.AsVariableStatement(), ast.KindDeclareKeyword)) {
					// possibly used before assigned, so just skip it
					// better to false negative and skip it, than false positive and fix to compile erroring code
					//
					// no better way to figure this out right now
					// https://github.com/Microsoft/TypeScript/issues/31124
					return true
				}
			}

			return false
		}
		isConstAssertion := func(node *ast.Node) bool {
			if !ast.IsTypeReferenceNode(node) {
				return false
			}
			typeName := node.AsTypeReferenceNode().TypeName
			return ast.IsIdentifier(typeName) && typeName.Text() == "const"
		}

		isImplicitlyNarrowedLiteralDeclaration := func(node *ast.Node) bool {
			expression := node.Expression()
			/**
			 * Even on `const` variable declarations, template literals with expressions can sometimes be widened without a type assertion.
			 * @see https://github.com/typescript-eslint/typescript-eslint/issues/8737
			 */
			if ast.IsTemplateExpression(expression) || expression.Kind == ast.KindNoSubstitutionTemplateLiteral {
				return false
			}

			return (ast.IsVariableDeclaration(node.Parent) && ast.IsVariableDeclarationList(node.Parent.Parent) && node.Parent.Parent.Flags&ast.NodeFlagsConst != 0) ||
				(ast.IsPropertyDeclaration(node.Parent) && node.Parent.ModifierFlags()&ast.ModifierFlagsReadonly != 0)

		}

		isTypeUnchanged := func(uncast, cast *checker.Type) bool {
			if uncast == cast {
				return true
			}

			if compilerOptions.ExactOptionalPropertyTypes.IsFalseOrUnknown() {
				return false
			}

			// if !utils.IsTypeFlagSet(uncast, checker.TypeFlagsUndefined) || !utils.IsTypeFlagSet(cast, checker.TypeFlagsUndefined) || !compilerOptions.ExactOptionalPropertyTypes.IsTrue() {
			// 	return false
			// }

			uncastParts := utils.Set[*checker.Type]{}
			uncastHasUndefined := false
			for _, part := range utils.UnionTypeParts(uncast) {
				if utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
					uncastHasUndefined = true
				} else {
					uncastParts.Add(part)
				}
			}

			if !uncastHasUndefined {
				return false
			}

			uncastPartsCount := uncastParts.Len()

			castPartsCount := 0
			castHasUndefined := false
			for _, part := range utils.UnionTypeParts(cast) {
				if utils.IsTypeFlagSet(part, checker.TypeFlagsUndefined) {
					castHasUndefined = true
				} else {
					if !uncastParts.Has(part) {
						return false
					}
					castPartsCount++
					if castPartsCount > uncastPartsCount {
						return false
					}
				}
			}

			return castHasUndefined && uncastPartsCount == castPartsCount
		}

		checkTypeAssertion := func(node *ast.Node) {
			typeNode := node.Type()
			typeText := strings.TrimSpace(ctx.SourceFile.Text()[typeNode.Pos():typeNode.End()])

			if slices.Contains(typesToIgnore, typeText) {
				return
			}

			castType := ctx.TypeChecker.GetTypeAtLocation(node)

			if !utils.IsTypeFlagSet(castType, checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral|checker.TypeFlagsBooleanLiteral|checker.TypeFlagsEnumLiteral) {
				if isConstAssertion(typeNode) {
					return
				}
			} else {
				if isConstAssertion(typeNode) {
					if !checkLiteralConstAssertions {
						return
					}
				} else {
					if !isImplicitlyNarrowedLiteralDeclaration(node) {
						return
					}
				}
			}

			expression := node.Expression()
			uncastType := ctx.TypeChecker.GetTypeAtLocation(expression)
			if !isTypeUnchanged(uncastType, castType) {
				return
			}

			msg := buildUnnecessaryAssertionMessage()

			if node.Kind == ast.KindAsExpression {
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
				asKeywordRange := s.TokenRange()

				sourceText := ctx.SourceFile.Text()
				startPos := asKeywordRange.Pos()

				if startPos > expression.End() && sourceText[startPos-1] == ' ' {
					if startPos-1 == expression.End() || (startPos-2 >= 0 && sourceText[startPos-2] != ' ') {
						startPos--
					}
				}

				fixRange := asKeywordRange.WithPos(startPos).WithEnd(typeNode.End())
				ctx.ReportNodeWithFixes(node, msg, rule.RuleFixRemoveRange(fixRange))
			} else {
				s := scanner.GetScannerForSourceFile(ctx.SourceFile, node.Pos())
				openingAngleBracket := s.TokenRange()

				fixRange := openingAngleBracket.WithEnd(expression.Pos())
				ctx.ReportNodeWithFixes(node, msg, rule.RuleFixRemoveRange(fixRange))
			}
			// TODO - add contextually unnecessary check for this
		}

		return rule.RuleListeners{
			ast.KindAsExpression:            checkTypeAssertion,
			ast.KindTypeAssertionExpression: checkTypeAssertion,

			ast.KindNonNullExpression: func(node *ast.Node) {
				expression := node.Expression()

				buildRemoveExclamationFix := func() rule.RuleFix {
					s := scanner.GetScannerForSourceFile(ctx.SourceFile, expression.End())
					return rule.RuleFixRemoveRange(s.TokenRange())
				}

				if ast.IsAssignmentExpression(node.Parent, true) {
					if node.Parent.AsBinaryExpression().Left == node {
						ctx.ReportNodeWithFixes(node, buildContextuallyUnnecessaryMessage(), buildRemoveExclamationFix())
					}
					// for all other = assignments we ignore non-null checks
					// this is because non-null assertions can change the type-flow of the code
					// so whilst they might be unnecessary for the assignment - they are necessary
					// for following code
					return
				}

				t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression)

				var tFlags checker.TypeFlags
				for _, part := range utils.UnionTypeParts(t) {
					tFlags |= checker.Type_flags(part)
				}

				if tFlags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|
					checker.TypeFlagsNull|
					checker.TypeFlagsUndefined|
					checker.TypeFlagsVoid) == 0 {
					if ast.IsIdentifier(expression) && isPossiblyUsedBeforeAssigned(expression) {
						return
					}
					ctx.ReportNodeWithFixes(node, buildUnnecessaryAssertionMessage(), buildRemoveExclamationFix())
				} else {
					// we know it's a nullable type
					// so figure out if the variable is used in a place that accepts nullable types
					contextualType := utils.GetContextualType(ctx.TypeChecker, node)
					if contextualType != nil {
						var contextualFlags checker.TypeFlags
						for _, part := range utils.UnionTypeParts(contextualType) {
							contextualFlags |= checker.Type_flags(part)
						}

						if tFlags&checker.TypeFlagsUnknown != 0 && contextualFlags&checker.TypeFlagsUnknown == 0 {
							return
						}

						// make sure that the parent accepts the same types
						// i.e. assigning `string | null | undefined` to `string | undefined` is invalid

						// In some cases (e.g. JSX attributes), the contextual type provided by the checker
						// might include `Null` even if the target property does not explicitly accept it
						// (especially when strictNullChecks is enabled).
						// To avoid false positives (saying assertion is unnecessary when it IS necessary),
						// we try to resolve the "clean" type from the target symbol if possible.
						cleanContextualType := contextualType
						var symbol *ast.Symbol

						if isStrictNullChecks {
							// Try to find the symbol we are assigning to, to get its declared type.
							// We walk up parentheses to find the enclosing context.
							parent := ast.WalkUpParenthesizedExpressions(node.Parent)
							if parent != nil && parent.Kind == ast.KindJsxExpression {
								// JSX Attribute value: <div key={expr} />
								// The parent of JsxExpression is JsxAttribute.
								grandparent := parent.Parent
								if grandparent != nil && grandparent.Kind == ast.KindJsxAttribute {
									attr := grandparent.AsJsxAttribute()
									// Get the symbol of the attribute (e.g. "key").
									// We look up the symbol at the attribute name location.
									if attr.Name() != nil {
										// attr.Name() returns *ast.JsxAttributeName which should have AsNode()
										symbol = ctx.TypeChecker.GetSymbolAtLocation(attr.Name().AsNode())
										if symbol != nil {
											// Get the type of the symbol (this should be the clean declared type).
											cleanContextualType = ctx.TypeChecker.GetTypeOfSymbol(symbol)
										}
									}
								}
							}
						}

						// Use IsTypeAssignableTo to check if the assertion is truly unnecessary.
						// We use cleanContextualType to ensure we don't use the polluted JSX type.
						isUnnecessary := true

						if utils.IsTypeFlagSet(cleanContextualType, checker.TypeFlagsAny) {
							// If context is Any, assertions are usually valid (or at least we don't nag).
							// This matches original behavior.
							isUnnecessary = false
						} else {
							// Check each part of the source type
							parts := utils.UnionTypeParts(t)
							for _, part := range parts {
								partFlags := checker.Type_flags(part)

								if partFlags&checker.TypeFlagsUndefined != 0 {
									// Check if undefined is allowed
									if checker.Checker_isTypeAssignableTo(ctx.TypeChecker, part, cleanContextualType) {
										continue
									}
									// Fallback: If we found a symbol (JSX), GetTypeOfSymbol might have stripped 'undefined' from optional property.
									// But the original contextualType should have it. We trust the original type for Undefined.
									if symbol != nil && checker.Checker_isTypeAssignableTo(ctx.TypeChecker, part, contextualType) {
										continue
									}
									isUnnecessary = false
									break
								}

								// For Null and others, we MUST check against cleanContextualType.
								// We specifically want to reject Null if it was only present in the polluted contextualType.
								if !checker.Checker_isTypeAssignableTo(ctx.TypeChecker, part, cleanContextualType) {
									isUnnecessary = false
									break
								}
							}
						}

						if isUnnecessary {
							ctx.ReportNodeWithFixes(node, buildContextuallyUnnecessaryMessage(), buildRemoveExclamationFix())
						}
					}
				}
			},
		}
	},
})
