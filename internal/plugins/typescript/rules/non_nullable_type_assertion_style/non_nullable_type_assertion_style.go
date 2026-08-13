package non_nullable_type_assertion_style

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildPreferNonNullAssertionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "preferNonNullAssertion",
		Description: "Use a ! assertion to more succinctly remove null and undefined from the type.",
	}
}

type nonNullableTypeParts struct {
	union  []*checker.Type
	single *checker.Type
}

func (p nonNullableTypeParts) len() int {
	if p.single != nil {
		return 1
	}
	return len(p.union)
}

func (p nonNullableTypeParts) at(index int) *checker.Type {
	if p.single != nil {
		return p.single
	}
	return p.union[index]
}

func isNullableType(t *checker.Type) bool {
	return t != nil && checker.Type_flags(t)&checker.TypeFlagsNullable != 0
}

func splitTypeParts(t *checker.Type, flags checker.TypeFlags) nonNullableTypeParts {
	if flags&checker.TypeFlagsUnion != 0 {
		return nonNullableTypeParts{union: t.Types()}
	}
	return nonNullableTypeParts{single: t}
}

func splitOriginalTypeParts(t *checker.Type, flags checker.TypeFlags) (nonNullableTypeParts, int, bool) {
	types := splitTypeParts(t, flags)
	if types.single != nil {
		if flags&checker.TypeFlagsNullable != 0 {
			return types, 0, true
		}
		return types, 1, false
	}

	nonNullableCount := 0
	for _, part := range types.union {
		if !isNullableType(part) {
			nonNullableCount++
		}
	}
	return types, nonNullableCount, nonNullableCount != len(types.union)
}

// couldBeNullable recursively checks if a type could be null or undefined.
// No explicit cycle detection is needed because TypeScript prevents circular
// type parameter constraints. This matches the upstream implementation.
func couldBeNullable(typeChecker *checker.Checker, t *checker.Type) bool {
	if t == nil {
		return false
	}
	flags := checker.Type_flags(t)
	if flags&checker.TypeFlagsTypeParameter != 0 {
		constraint := checker.Checker_getBaseConstraintOfType(typeChecker, t)
		if constraint == nil {
			return true
		}
		return couldBeNullable(typeChecker, constraint)
	}
	if flags&checker.TypeFlagsUnion != 0 {
		for _, part := range t.Types() {
			if couldBeNullable(typeChecker, part) {
				return true
			}
		}
		return false
	}
	return flags&checker.TypeFlagsNullable != 0
}

func sameTypesWithoutNullable(
	typeChecker *checker.Checker,
	assertedTypes nonNullableTypeParts,
	originalTypes nonNullableTypeParts,
	originalNonNullableTypeCount int,
) bool {
	assertedTypeCount := assertedTypes.len()
	if assertedTypeCount != originalNonNullableTypeCount {
		return false
	}
	for i := range assertedTypeCount {
		if couldBeNullable(typeChecker, assertedTypes.at(i)) {
			return false
		}
	}

	// Small unions dominate in practice, so compare their constituent pointers
	// directly and avoid allocating a map. TypeScript union constituents are
	// unique, so equal counts plus one-way membership proves set equality. Fall
	// back to one set for wide unions to keep adversarial inputs linear.
	const linearTypeComparisonLimit = 8
	originalTypeCount := originalTypes.len()
	if originalTypeCount > linearTypeComparisonLimit || assertedTypeCount > linearTypeComparisonLimit {
		nonNullableOriginalTypes := make(map[*checker.Type]struct{}, originalNonNullableTypeCount)
		for i := range originalTypeCount {
			originalType := originalTypes.at(i)
			if !isNullableType(originalType) {
				nonNullableOriginalTypes[originalType] = struct{}{}
			}
		}

		for i := range assertedTypeCount {
			assertedType := assertedTypes.at(i)
			if _, ok := nonNullableOriginalTypes[assertedType]; !ok {
				return false
			}
		}
		return true
	}

	for i := range assertedTypeCount {
		assertedType := assertedTypes.at(i)
		found := false
		for j := range originalTypeCount {
			if originalTypes.at(j) == assertedType {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func checkAssertion(ctx *rule.RuleContext, node *ast.Node) {
	if ast.IsConstAssertion(node) {
		return
	}

	expression := node.Expression()
	originalType := ctx.TypeChecker.GetTypeAtLocation(expression)
	if originalType == nil {
		return
	}
	originalFlags := checker.Type_flags(originalType)
	if originalFlags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return
	}
	originalTypes, originalNonNullableTypeCount, hasNullableType := splitOriginalTypeParts(originalType, originalFlags)
	if !hasNullableType {
		return
	}

	assertedType := ctx.TypeChecker.GetTypeAtLocation(node.Type())
	if assertedType == nil {
		return
	}
	assertedFlags := checker.Type_flags(assertedType)
	if assertedFlags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
		return
	}
	assertedTypes := splitTypeParts(assertedType, assertedFlags)
	if !sameTypesWithoutNullable(ctx.TypeChecker, assertedTypes, originalTypes, originalNonNullableTypeCount) {
		return
	}

	ctx.ReportNodeWithDeferredFixes(node, buildPreferNonNullAssertionMessage(), func() []rule.RuleFix {
		unwrappedExpression := ast.SkipParentheses(expression)
		expressionText := utils.TrimmedNodeText(ctx.SourceFile, unwrappedExpression)
		if ast.GetExpressionPrecedence(unwrappedExpression) <= ast.OperatorPrecedenceUnary {
			expressionText = "(" + expressionText + ")"
		}
		return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, expressionText+"!")}
	})
}

type nonNullableTypeAssertionStyleState struct {
	ctx rule.RuleContext
}

func (s *nonNullableTypeAssertionStyleState) checkAssertion(node *ast.Node) {
	checkAssertion(&s.ctx, node)
}

func runNonNullableTypeAssertionStyle(ctx rule.RuleContext, _ []any) rule.RuleListeners {
	compilerOptions := ctx.Program.Options()
	if !utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.StrictNullChecks) {
		// With the TypeScript versions supported by the upstream rule, null and
		// undefined are erased when strictNullChecks is not enabled, so the rule
		// cannot find an assertion that only removes those constituents. Keep that
		// behavior even though newer typescript-go versions default strictness on.
		return nil
	}

	state := &nonNullableTypeAssertionStyleState{ctx: ctx}
	listener := state.checkAssertion
	return rule.RuleListeners{
		ast.KindAsExpression:            listener,
		ast.KindTypeAssertionExpression: listener,
	}
}

var NonNullableTypeAssertionStyleRule = rule.CreateRule(rule.Rule{
	Name:             "non-nullable-type-assertion-style",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run:              runNonNullableTypeAssertionStyle,
})
