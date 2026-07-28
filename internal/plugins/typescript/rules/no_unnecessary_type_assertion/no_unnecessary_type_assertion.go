package no_unnecessary_type_assertion

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

const nullableTypeFlags = checker.TypeFlagsAny |
	checker.TypeFlagsUnknown |
	checker.TypeFlagsNull |
	checker.TypeFlagsUndefined |
	checker.TypeFlagsVoid

func getUnionTypeFlags(t *checker.Type) checker.TypeFlags {
	var flags checker.TypeFlags
	for _, part := range utils.UnionTypeParts(t) {
		flags |= checker.Type_flags(part)
	}
	return flags
}

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
	TypesToIgnore []string `json:"typesToIgnore"`
	// Whether to check const assertions on literal values
	// When true, reports cases like `const foo = 'bar' as const` where the assertion is unnecessary
	CheckLiteralConstAssertions bool `json:"checkLiteralConstAssertions"`
}

var NoUnnecessaryTypeAssertionRule = rule.CreateRule(rule.Rule{
	Name:             "no-unnecessary-type-assertion",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.LegacyUnwrapOptions(_options)
		opts := NoUnnecessaryTypeAssertionOptions{}
		if options != nil {
			// Try direct type assertion first (for Go tests)
			if directOpts, ok := options.(NoUnnecessaryTypeAssertionOptions); ok {
				opts = directOpts
			} else {
				// For IPC mode, options come as map[string]interface{}, convert via JSON
				if jsonBytes, err := json.Marshal(options); err == nil {
					_ = json.Unmarshal(jsonBytes, &opts)
				}
			}
		}
		if opts.TypesToIgnore == nil {
			opts.TypesToIgnore = []string{}
		}

		sourceText := ctx.SourceFile.Text()
		var fixScanner *scanner.Scanner
		getTokenRange := func(pos int) core.TextRange {
			if fixScanner == nil {
				fixScanner = scanner.NewScanner()
			} else {
				fixScanner.Reset()
			}
			fixScanner.SetText(sourceText)
			fixScanner.SetLanguageVariant(ctx.SourceFile.LanguageVariant)
			fixScanner.ResetPos(pos)
			fixScanner.Scan()
			return fixScanner.TokenRange()
		}

		compilerOptions := ctx.Program.Options()
		isStrictNullChecks := utils.IsStrictCompilerOptionEnabled(
			compilerOptions,
			compilerOptions.StrictNullChecks,
		)

		/**
		 * Returns true if there's a chance the variable has been used before a value has been assigned to it
		 */
		isPossiblyUsedBeforeAssigned := func(
			node *ast.Node,
			declaration *ast.Declaration,
			constrainedType *checker.Type,
		) bool {
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
				if constrainedType == nil {
					constrainedType = utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, node)
				}
				if declarationType == constrainedType &&
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

		type identifierInfo struct {
			declaration     *ast.Declaration
			constrainedType *checker.Type
			typeFlags       checker.TypeFlags
		}
		identifierInfoCache := make(map[*ast.Symbol]identifierInfo)
		getIdentifierInfo := func(node *ast.Node) (identifierInfo, bool) {
			symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
			if symbol == nil {
				return identifierInfo{}, false
			}
			if info, ok := identifierInfoCache[symbol]; ok {
				return info, true
			}

			info := identifierInfo{}
			if len(symbol.Declarations) > 0 {
				info.declaration = symbol.Declarations[0]
			}
			declaredType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(symbol, nil)
			if declaredType != nil {
				info.constrainedType = declaredType
				if constraint := checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, declaredType); constraint != nil {
					info.constrainedType = constraint
				}
				info.typeFlags = getUnionTypeFlags(info.constrainedType)
			}
			identifierInfoCache[symbol] = info
			return info, true
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
			if ast.IsTemplateExpression(expression) {
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
			if slices.Contains(opts.TypesToIgnore, strings.TrimSpace(sourceText[typeNode.Pos():typeNode.End()])) {
				return
			}

			typeAnnotationIsConstAssertion := isConstAssertion(typeNode)
			if typeAnnotationIsConstAssertion && !opts.CheckLiteralConstAssertions {
				return
			}

			castType := ctx.TypeChecker.GetTypeAtLocation(node)

			if utils.IsTypeFlagSet(castType, checker.TypeFlagsStringLiteral|checker.TypeFlagsNumberLiteral|checker.TypeFlagsBigIntLiteral) {
				// For literal types, only check if it's an implicitly narrowed declaration
				// (e.g., const variable or readonly property)
				// OR if checkLiteralConstAssertions is enabled for explicit const assertions
				if !isImplicitlyNarrowedLiteralDeclaration(node) &&
					(!opts.CheckLiteralConstAssertions || !typeAnnotationIsConstAssertion) {
					return
				}
			}

			expression := node.Expression()
			uncastType := ctx.TypeChecker.GetTypeAtLocation(expression)
			if !isTypeUnchanged(uncastType, castType) {
				return
			}

			msg := buildUnnecessaryAssertionMessage()

			if node.Kind == ast.KindAsExpression {
				asKeywordRange := getTokenRange(expression.End())

				startPos := asKeywordRange.Pos()

				if startPos > expression.End() && sourceText[startPos-1] == ' ' {
					if startPos-1 == expression.End() || (startPos-2 >= 0 && sourceText[startPos-2] != ' ') {
						startPos--
					}
				}

				fixRange := asKeywordRange.WithPos(startPos).WithEnd(typeNode.End())
				ctx.ReportNodeWithFixes(node, msg, rule.RuleFixRemoveRange(fixRange))
			} else {
				openingAngleBracket := getTokenRange(node.Pos())

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
					return rule.RuleFixRemoveRange(getTokenRange(expression.End()))
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

				var (
					expressionIdentifierInfo identifierInfo
					hasIdentifierInfo        bool
				)
				if ast.IsIdentifier(expression) {
					expressionIdentifierInfo, hasIdentifierInfo = getIdentifierInfo(expression)
					// The checker assumes an identifier below a non-null expression is
					// initialized, and flow analysis can only preserve or narrow its
					// declared constrained type. A declared non-nullish type therefore
					// cannot become nullish at this location.
					if hasIdentifierInfo &&
						expressionIdentifierInfo.declaration != nil &&
						expressionIdentifierInfo.constrainedType != nil &&
						expressionIdentifierInfo.typeFlags&nullableTypeFlags == 0 {
						if isPossiblyUsedBeforeAssigned(expression, expressionIdentifierInfo.declaration, nil) {
							return
						}
						ctx.ReportNodeWithFixes(node, buildUnnecessaryAssertionMessage(), buildRemoveExclamationFix())
						return
					}
				}

				t := utils.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression)
				tFlags := getUnionTypeFlags(t)

				if tFlags&nullableTypeFlags == 0 {
					if ast.IsIdentifier(expression) {
						declaration := expressionIdentifierInfo.declaration
						if !hasIdentifierInfo {
							declaration = utils.GetDeclaration(ctx.TypeChecker, expression)
						}
						if isPossiblyUsedBeforeAssigned(expression, declaration, t) {
							return
						}
					}
					ctx.ReportNodeWithFixes(node, buildUnnecessaryAssertionMessage(), buildRemoveExclamationFix())
				} else {
					// we know it's a nullable type
					// so figure out if the variable is used in a place that accepts nullable types
					contextualType := utils.GetContextualType(ctx.TypeChecker, node)
					if contextualType != nil {
						contextualFlags := getUnionTypeFlags(contextualType)

						if tFlags&checker.TypeFlagsUnknown != 0 && contextualFlags&checker.TypeFlagsUnknown == 0 {
							return
						}

						// in strict mode you can't assign null to undefined, so we have to make sure that
						// the two types share a nullable type
						typeIncludesUndefined := tFlags&checker.TypeFlagsUndefined != 0
						typeIncludesNull := tFlags&checker.TypeFlagsNull != 0
						typeIncludesVoid := tFlags&checker.TypeFlagsVoid != 0

						contextualTypeIncludesUndefined := contextualFlags&checker.TypeFlagsUndefined != 0
						contextualTypeIncludesNull := contextualFlags&checker.TypeFlagsNull != 0
						contextualTypeIncludesVoid := contextualFlags&checker.TypeFlagsVoid != 0

						// make sure that the parent accepts the same types
						// i.e. assigning `string | null | undefined` to `string | undefined` is invalid
						isValidUndefined := !typeIncludesUndefined || contextualTypeIncludesUndefined
						isValidNull := !typeIncludesNull || contextualTypeIncludesNull
						isValidVoid := !typeIncludesVoid || contextualTypeIncludesVoid

						if isValidUndefined && isValidNull && isValidVoid {
							ctx.ReportNodeWithFixes(node, buildContextuallyUnnecessaryMessage(), buildRemoveExclamationFix())
						}
					}
				}
			},
		}
	},
})
