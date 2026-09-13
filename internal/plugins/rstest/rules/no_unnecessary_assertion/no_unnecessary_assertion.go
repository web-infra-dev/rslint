package no_unnecessary_assertion

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type checkedMatcher struct {
	flags checker.TypeFlags
	thing string
	kind  rstestUtils.RstestExpectMatcherKind
}

var checkedMatchers = map[string]checkedMatcher{
	"toBeNull":      {flags: checker.TypeFlagsNull, thing: "null", kind: rstestUtils.RstestExpectMatcherCall},
	"toBeUndefined": {flags: checker.TypeFlagsUndefined, thing: "undefined", kind: rstestUtils.RstestExpectMatcherCall},
	"toBeDefined":   {flags: checker.TypeFlagsUndefined, thing: "undefined", kind: rstestUtils.RstestExpectMatcherCall},
	"toBeNaN":       {flags: checker.TypeFlagsNumberLike, thing: "a number", kind: rstestUtils.RstestExpectMatcherCall},
	"null":          {flags: checker.TypeFlagsNull, thing: "null", kind: rstestUtils.RstestExpectMatcherProperty},
	"undefined":     {flags: checker.TypeFlagsUndefined, thing: "undefined", kind: rstestUtils.RstestExpectMatcherProperty},
	"NaN":           {flags: checker.TypeFlagsNumberLike, thing: "a number", kind: rstestUtils.RstestExpectMatcherProperty},
}

func unnecessaryAssertionMessage(thing string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unnecessaryAssertion",
		Description: "Unnecessary assertion, subject cannot be " + thing,
	}
}

func noStrictNullCheckMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noStrictNullCheck",
		Description: "This rule requires the `strictNullChecks` compiler option to be turned on to function correctly.",
	}
}

var NoUnnecessaryAssertionRule = rule.Rule{
	Name:             "rstest/no-unnecessary-assertion",
	Schema:           rule.EmptyArraySchema,
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		compilerOptions := ctx.Program().Options()
		if !utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.StrictNullChecks) {
			ctx.ReportRange(core.NewTextRange(0, 0), noStrictNullCheckMessage())
		}

		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseExpectCall(node)
				if parsed == nil ||
					parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
					parsed.Head == nil ||
					parsed.MatcherEntry == nil ||
					len(parsed.Matchers) == 0 ||
					parsed.Entry == rstestUtils.RstestExpectEntryElement ||
					parsed.Entry == rstestUtils.RstestExpectEntryStatic {
					return
				}

				matcher, ok := checkedMatchers[parsed.Matcher]
				if !ok || parsed.Matchers[0].Kind != matcher.kind {
					return
				}
				// Rstest's poll proxy only defers function matchers. Chai property
				// getters execute immediately against the proxy's placeholder null
				// subject, before the callback is invoked, so its awaited return type
				// does not describe those assertions.
				if parsed.Entry == rstestUtils.RstestExpectEntryPoll &&
					matcher.kind == rstestUtils.RstestExpectMatcherProperty {
					return
				}
				for _, modifier := range parsed.Modifiers {
					if modifier != "not" {
						return
					}
				}

				head := parsed.Head.AsCallExpression()
				if head == nil {
					return
				}
				possibility := runtimeSubjectPossibility(
					ctx.TypeChecker,
					head.Arguments.Nodes,
					parsed.Entry,
					matcher.flags,
				)
				if possibility != typePossibilityImpossible {
					return
				}

				ctx.ReportNode(parsed.Expression, unnecessaryAssertionMessage(matcher.thing))
			},
		}
	},
}

type typePossibility uint8

const (
	typePossibilityImpossible typePossibility = iota
	typePossibilityUnknown
	typePossibilityPossible
)

func mergePossibilities(left, right typePossibility) typePossibility {
	if left == typePossibilityPossible || right == typePossibilityPossible {
		return typePossibilityPossible
	}
	if left == typePossibilityUnknown || right == typePossibilityUnknown {
		return typePossibilityUnknown
	}
	return typePossibilityImpossible
}

type runtimeSubjects struct {
	types          []*checker.Type
	mayBeUndefined bool
	unknown        bool
}

func runtimeSubjectPossibility(
	typeChecker *checker.Checker,
	arguments []*ast.Node,
	entry rstestUtils.RstestExpectEntry,
	desiredFlags checker.TypeFlags,
) typePossibility {
	subjects := resolveRuntimeSubjects(typeChecker, arguments)
	result := typePossibilityImpossible
	if subjects.unknown {
		result = typePossibilityUnknown
	}
	if subjects.mayBeUndefined {
		if entry == rstestUtils.RstestExpectEntryPoll {
			result = mergePossibilities(result, typePossibilityUnknown)
		} else if desiredFlags&checker.TypeFlagsUndefined != 0 {
			return typePossibilityPossible
		}
	}

	analyzer := newTypePossibilityAnalyzer(typeChecker, desiredFlags)
	for _, subjectType := range subjects.types {
		var candidate typePossibility
		if entry == rstestUtils.RstestExpectEntryPoll {
			candidate = pollCallbackPossibility(analyzer, subjectType)
		} else {
			candidate = analyzer.analyze(subjectType)
		}
		result = mergePossibilities(result, candidate)
		if result == typePossibilityPossible {
			return result
		}
	}
	if len(subjects.types) == 0 && !subjects.mayBeUndefined && !subjects.unknown {
		return typePossibilityUnknown
	}
	return result
}

func resolveRuntimeSubjects(
	typeChecker *checker.Checker,
	arguments []*ast.Node,
) runtimeSubjects {
	result := runtimeSubjects{}
	for _, argument := range arguments {
		if argument == nil {
			result.unknown = true
			return result
		}
		if !ast.IsSpreadElement(argument) {
			result.types = append(result.types, typeChecker.GetTypeAtLocation(argument))
			return result
		}

		spread := resolveSpreadRuntimeSubjects(
			typeChecker,
			typeChecker.GetTypeAtLocation(argument.AsSpreadElement().Expression),
			make(map[*checker.Type]bool),
		)
		result.types = append(result.types, spread.types...)
		if spread.unknown {
			result.unknown = true
			return result
		}
		if !spread.mayBeUndefined {
			return result
		}
	}
	result.mayBeUndefined = true
	return result
}

func resolveSpreadRuntimeSubjects(
	typeChecker *checker.Checker,
	typeValue *checker.Type,
	active map[*checker.Type]bool,
) runtimeSubjects {
	if typeValue == nil || active[typeValue] {
		return runtimeSubjects{unknown: true}
	}
	active[typeValue] = true
	defer delete(active, typeValue)

	if utils.IsTypeParameter(typeValue) {
		constraint, _ := utils.GetConstraintInfo(typeChecker, typeValue)
		if constraint == nil {
			return runtimeSubjects{unknown: true}
		}
		return resolveSpreadRuntimeSubjects(typeChecker, constraint, active)
	}

	if utils.IsUnionType(typeValue) {
		result := runtimeSubjects{}
		for _, part := range utils.UnionTypeParts(typeValue) {
			candidate := resolveSpreadRuntimeSubjects(typeChecker, part, active)
			result.types = append(result.types, candidate.types...)
			result.mayBeUndefined = result.mayBeUndefined || candidate.mayBeUndefined
			result.unknown = result.unknown || candidate.unknown
		}
		return result
	}

	if checker.IsTupleType(typeValue) {
		typeArguments := checker.Checker_getTypeArguments(typeChecker, typeValue)
		elementFlags := typeValue.TargetTupleType().ElementFlags()
		if len(typeArguments) == 0 || len(elementFlags) == 0 {
			return runtimeSubjects{mayBeUndefined: true}
		}
		firstFlags := elementFlags[0]
		switch {
		case firstFlags&checker.ElementFlagsRequired != 0:
			return runtimeSubjects{types: []*checker.Type{typeArguments[0]}}
		case firstFlags&checker.ElementFlagsOptional != 0:
			return runtimeSubjects{
				types:          []*checker.Type{typeArguments[0]},
				mayBeUndefined: true,
			}
		case firstFlags&checker.ElementFlagsRest != 0:
			elementType := spreadElementType(typeChecker, typeValue)
			if elementType == nil {
				return runtimeSubjects{unknown: true}
			}
			return runtimeSubjects{
				types:          []*checker.Type{elementType},
				mayBeUndefined: true,
			}
		default:
			// A variadic generic tuple can expand to either an empty or non-empty
			// sequence, and its first concrete element is not known here.
			return runtimeSubjects{unknown: true}
		}
	}

	elementType := spreadElementType(typeChecker, typeValue)
	if elementType == nil {
		return runtimeSubjects{unknown: true}
	}
	return runtimeSubjects{
		types:          []*checker.Type{elementType},
		mayBeUndefined: true,
	}
}

func spreadElementType(
	typeChecker *checker.Checker,
	typeValue *checker.Type,
) *checker.Type {
	elementType := checker.Checker_getIterationTypeOfIterable(
		typeChecker,
		checker.IterationUseSpread,
		checker.IterationTypeKindYield,
		typeValue,
		nil,
	)
	if elementType != nil {
		return elementType
	}
	constraint := checker.Checker_getBaseConstraintOfType(typeChecker, typeValue)
	if constraint == nil {
		constraint = typeValue
	}
	return utils.GetNumberIndexType(typeChecker, constraint)
}

func pollCallbackPossibility(
	analyzer *typePossibilityAnalyzer,
	subjectType *checker.Type,
) typePossibility {
	signatures := utils.GetCallSignatures(analyzer.typeChecker, subjectType)
	result := typePossibilityImpossible
	checkedZeroArgumentSignature := false
	for _, signature := range signatures {
		if checker.Checker_getMinArgumentCount(analyzer.typeChecker, signature) != 0 {
			continue
		}
		checkedZeroArgumentSignature = true
		returnType := checker.Checker_getReturnTypeOfSignature(analyzer.typeChecker, signature)
		awaitedType := checker.Checker_getAwaitedType(analyzer.typeChecker, returnType)
		if awaitedType == nil {
			result = mergePossibilities(result, typePossibilityUnknown)
			continue
		}
		result = mergePossibilities(result, analyzer.analyze(awaitedType))
		if result == typePossibilityPossible {
			return result
		}
	}
	if !checkedZeroArgumentSignature {
		return typePossibilityUnknown
	}
	return result
}

type typePossibilityAnalyzer struct {
	typeChecker  *checker.Checker
	desiredFlags checker.TypeFlags
	active       map[*checker.Type]bool
	memo         map[*checker.Type]typePossibility
}

func newTypePossibilityAnalyzer(
	typeChecker *checker.Checker,
	desiredFlags checker.TypeFlags,
) *typePossibilityAnalyzer {
	return &typePossibilityAnalyzer{
		typeChecker:  typeChecker,
		desiredFlags: desiredFlags,
		active:       make(map[*checker.Type]bool),
		memo:         make(map[*checker.Type]typePossibility),
	}
}

func (analyzer *typePossibilityAnalyzer) analyze(typeValue *checker.Type) typePossibility {
	if typeValue == nil {
		return typePossibilityUnknown
	}
	if result, ok := analyzer.memo[typeValue]; ok {
		return result
	}
	if analyzer.active[typeValue] {
		return typePossibilityUnknown
	}
	analyzer.active[typeValue] = true
	result := analyzer.analyzeUncached(typeValue)
	delete(analyzer.active, typeValue)
	analyzer.memo[typeValue] = result
	return result
}

func (analyzer *typePossibilityAnalyzer) analyzeUncached(typeValue *checker.Type) typePossibility {
	result := typePossibilityImpossible
	for _, part := range utils.UnionTypeParts(typeValue) {
		flags := checker.Type_flags(part)
		// TypeScript permits a value-returning function to be assigned to a
		// void-returning signature. A call whose static type is void can therefore
		// produce any runtime value, including the values checked by this rule.
		if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsVoid|analyzer.desiredFlags) != 0 {
			return typePossibilityPossible
		}
		if flags&checker.TypeFlagsTypeParameter != 0 {
			constraint, _ := utils.GetConstraintInfo(analyzer.typeChecker, part)
			if constraint == nil {
				result = mergePossibilities(result, typePossibilityUnknown)
			} else {
				result = mergePossibilities(result, analyzer.analyze(constraint))
			}
			continue
		}
		// Conditional, substitution, and indexed-access types may resolve to a
		// matching constituent after generic instantiation. Treat them as unknown
		// instead of claiming the assertion is impossible.
		if flags&checker.TypeFlagsInstantiableNonPrimitive != 0 ||
			(flags&checker.TypeFlagsIndex != 0 && analyzer.desiredFlags&checker.TypeFlagsNumberLike != 0) {
			result = mergePossibilities(result, typePossibilityUnknown)
			continue
		}
		if analyzer.desiredFlags&checker.TypeFlagsNumberLike != 0 &&
			checker.Checker_isTypeAssignableTo(
				analyzer.typeChecker,
				checker.Checker_numberType(analyzer.typeChecker),
				part,
			) {
			return typePossibilityPossible
		}
		if utils.IsIntersectionType(part) {
			intersectionResult := typePossibilityImpossible
			for _, constituent := range utils.IntersectionTypeParts(part) {
				intersectionResult = mergePossibilities(intersectionResult, analyzer.analyze(constituent))
				if intersectionResult == typePossibilityPossible {
					return intersectionResult
				}
			}
			result = mergePossibilities(result, intersectionResult)
		}
	}
	return result
}
