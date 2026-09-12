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
				if head == nil || len(head.Arguments.Nodes) == 0 {
					return
				}
				checked, canBeDesiredType := subjectCanBeDesiredType(
					ctx.TypeChecker,
					head.Arguments.Nodes[0],
					parsed.Entry,
					matcher.flags,
				)
				if !checked || canBeDesiredType {
					return
				}

				ctx.ReportNode(parsed.Expression, unnecessaryAssertionMessage(matcher.thing))
			},
		}
	},
}

func subjectCanBeDesiredType(
	typeChecker *checker.Checker,
	subject *ast.Node,
	entry rstestUtils.RstestExpectEntry,
	desiredFlags checker.TypeFlags,
) (checked bool, canBeDesiredType bool) {
	subjectType := typeChecker.GetTypeAtLocation(subject)
	if entry != rstestUtils.RstestExpectEntryPoll {
		return true, typeCanBeDesiredType(typeChecker, subjectType, desiredFlags, nil)
	}

	signatures := utils.GetCallSignatures(typeChecker, subjectType)
	if len(signatures) == 0 {
		return false, false
	}
	for _, signature := range signatures {
		returnType := checker.Checker_getReturnTypeOfSignature(typeChecker, signature)
		awaitedType := checker.Checker_getAwaitedType(typeChecker, returnType)
		if awaitedType == nil {
			return false, false
		}
		if typeCanBeDesiredType(typeChecker, awaitedType, desiredFlags, nil) {
			return true, true
		}
	}
	return true, false
}

func typeCanBeDesiredType(
	typeChecker *checker.Checker,
	typeValue *checker.Type,
	desiredFlags checker.TypeFlags,
	seen map[*checker.Type]bool,
) bool {
	if typeValue == nil {
		return true
	}
	if seen == nil {
		seen = make(map[*checker.Type]bool)
	}
	if seen[typeValue] {
		return true
	}
	seen[typeValue] = true

	for _, part := range utils.UnionTypeParts(typeValue) {
		flags := checker.Type_flags(part)
		// TypeScript permits a value-returning function to be assigned to a
		// void-returning signature. A call whose static type is void can therefore
		// produce any runtime value, including the values checked by this rule.
		if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown|checker.TypeFlagsVoid|desiredFlags) != 0 {
			return true
		}
		if flags&checker.TypeFlagsTypeParameter != 0 {
			constraint, _ := utils.GetConstraintInfo(typeChecker, part)
			if constraint == nil || typeCanBeDesiredType(typeChecker, constraint, desiredFlags, seen) {
				return true
			}
			continue
		}
		// Conditional, substitution, and indexed-access types may resolve to a
		// matching constituent after generic instantiation. Treat them as unknown
		// instead of claiming the assertion is impossible.
		if flags&checker.TypeFlagsInstantiableNonPrimitive != 0 ||
			(flags&checker.TypeFlagsIndex != 0 && desiredFlags&checker.TypeFlagsNumberLike != 0) {
			return true
		}
		if utils.IsIntersectionType(part) && utils.Some(
			utils.IntersectionTypeParts(part),
			func(constituent *checker.Type) bool {
				return typeCanBeDesiredType(typeChecker, constituent, desiredFlags, seen)
			},
		) {
			return true
		}
	}
	return false
}
