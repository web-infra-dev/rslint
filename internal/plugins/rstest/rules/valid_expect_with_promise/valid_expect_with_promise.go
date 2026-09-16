package valid_expect_with_promise

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed valid_expect_with_promise.schema.json
var schema []byte

var ValidExpectWithPromiseRule = rule.Rule{
	Name:             "rstest/valid-expect-with-promise",
	RequiresTypeInfo: true,
	Schema:           rule.NewSchema(schema),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		checkThenables := false
		if len(options) > 0 {
			if option, ok := options[0].(map[string]any); ok {
				checkThenables, _ = option["checkThenables"].(bool)
			}
		}
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		isPromise := func(node *ast.Node, typ *checker.Type) bool {
			return utils.IsPromiseLike(ctx.Program(), ctx.TypeChecker, typ) ||
				(checkThenables && isStrictThenable(ctx.TypeChecker, node, typ))
		}
		return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
			parsed := analysis.ParseExpectCallThroughTransparentExpressions(node)
			if parsed == nil || parsed.Reason != rstestUtils.RstestExpectParseReasonNone ||
				(parsed.Entry != rstestUtils.RstestExpectEntryCall && parsed.Entry != rstestUtils.RstestExpectEntrySoft) {
				return
			}
			arguments := parsed.Head.AsCallExpression().Arguments.Nodes
			if len(arguments) == 0 {
				return
			}
			subject := arguments[0]
			modifier := parsed.PromiseModifierEntry()
			typ := ctx.TypeChecker.GetTypeAtLocation(subject)
			typ, poorlyExpected, known := analyzePromiseSubject(
				ctx.TypeChecker, parsed, modifier, typ, func(typ *checker.Type) bool {
					return isPromise(subject, typ)
				},
			)
			if poorlyExpected {
				ctx.ReportNode(parsed.Expression, rule.RuleMessage{Id: "poorlyExpectedPromise", Description: "Subject is a promise so resolve or reject should be used"})
				return
			}
			if modifier == nil || !known {
				return
			}
			promise := false
			if modifier.Name == "rejects" {
				// Rstest invokes callable subjects before testing their returned value.
				promise = utils.Every(utils.UnionTypeParts(checker.Checker_getApparentType(ctx.TypeChecker, typ)), func(part *checker.Type) bool {
					signatures := checker.Checker_getSignaturesOfType(ctx.TypeChecker, part, checker.SignatureKindCall)
					if len(signatures) == 0 {
						return isPromise(subject, part)
					}
					returned, ok := callableReturnType(ctx.TypeChecker, subject, part)
					return ok && isPromise(subject, returned)
				})
			} else {
				promise = isPromise(subject, typ)
			}
			if !promise {
				ctx.ReportNode(modifier.Node, rule.RuleMessage{Id: "unneededRejectResolve", Description: "Subject is not a promise so " + modifier.Name + " is not needed"})
			}
		}}
	},
}

// analyzePromiseSubject follows the Chai assertion chain once. Matchers inspect
// the current subject before any subject transformation they perform; for
// example, property('value') checks the original object and only then makes the
// selected property the subject of subsequent matchers and modifiers.
func analyzePromiseSubject(
	typeChecker *checker.Checker,
	parsed *rstestUtils.ParsedRstestExpectCall,
	modifier *rstestUtils.ParsedRstestFnMemberEntry,
	typ *checker.Type,
	isPromise func(*checker.Type) bool,
) (*checker.Type, bool, bool) {
	nested := false
	matcherIndex := 0
	subjectUnchecked := true
	for i := range parsed.MemberEntries {
		entry := &parsed.MemberEntries[i]
		if modifier != nil && entry.Node == modifier.Node {
			return typ, false, true
		}
		isMatcher := matcherIndex < len(parsed.Matchers) &&
			parsed.Matchers[matcherIndex].Entry.Node == entry.Node
		if isMatcher {
			matcherIndex++
			if modifier == nil && !isChaiPropertySubjectTransform(entry) && subjectUnchecked {
				subjectUnchecked = false
				if isPromise(typ) {
					return typ, true, true
				}
			}
		}
		switch entry.Name {
		case "nested":
			if rstestUtils.MatcherCall(entry) == nil {
				nested = true
			}
		case "property", "ownProperty", "haveOwnProperty":
			call := rstestUtils.MatcherCall(entry)
			if call == nil || nested || len(call.AsCallExpression().Arguments.Nodes) == 0 {
				return nil, false, false
			}
			name, ok := utils.GetStaticExpressionValue(utils.SkipAssertionsAndParens(call.AsCallExpression().Arguments.Nodes[0]))
			if !ok {
				return nil, false, false
			}
			typ = typeChecker.GetTypeOfPropertyOfType(typ, name)
			if typ == nil {
				return nil, false, false
			}
			subjectUnchecked = true
		case "ownPropertyDescriptor", "haveOwnPropertyDescriptor",
			"throw", "throws", "Throw", "toThrow", "toThrowError", "toContain":
			if rstestUtils.MatcherCall(entry) != nil {
				return nil, false, false
			}
		}
	}
	return typ, false, modifier == nil
}

func isChaiPropertySubjectTransform(entry *rstestUtils.ParsedRstestFnMemberEntry) bool {
	if entry == nil || rstestUtils.MatcherCall(entry) == nil {
		return false
	}
	switch entry.Name {
	case "property", "ownProperty", "haveOwnProperty":
		return true
	default:
		return false
	}
}

func callableReturnType(typeChecker *checker.Checker, subject *ast.Node, typ *checker.Type) (*checker.Type, bool) {
	signatures := checker.Checker_getSignaturesOfType(typeChecker, typ, checker.SignatureKindCall)
	if !slices.ContainsFunc(signatures, func(signature *checker.Signature) bool {
		return checker.Checker_getMinArgumentCount(typeChecker, signature) == 0
	}) {
		return nil, false
	}
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	expression := factory.NewSyntheticExpression(typ, false, nil)
	expression.Loc = subject.Loc
	arguments := factory.NewNodeList(nil)
	arguments.Loc = core.NewTextRange(subject.Pos(), subject.Pos())
	call := factory.NewCallExpression(expression, nil, nil, arguments, ast.NodeFlagsNone)
	call.Loc = subject.Loc
	call.Parent = subject.Parent
	expression.Parent = call
	// Candidate collection suppresses diagnostics for the synthetic zero-argument call.
	var candidates []*checker.Signature
	signature := checker.Checker_getResolvedSignature(typeChecker, call, &candidates, checker.CheckModeNormal)
	if signature == nil || signature.Flags()&checker.SignatureFlagsIsSignatureCandidateForOverloadFailure != 0 {
		return nil, false
	}
	if thisParameter := signature.ThisParameter(); thisParameter != nil {
		thisType := typeChecker.GetTypeOfSymbol(thisParameter)
		if !checker.Checker_isTypeAssignableTo(typeChecker, typeChecker.GetVoidType(), thisType) {
			return nil, false
		}
	}
	return checker.Checker_getReturnTypeOfSignature(typeChecker, signature), true
}

// A single-callback chainable is deliberately excluded, unlike utils.IsThenableType.
func isStrictThenable(typeChecker *checker.Checker, node *ast.Node, typ *checker.Type) bool {
	if utils.IsIntersectionType(typ) {
		return slices.ContainsFunc(typ.Types(), func(part *checker.Type) bool { return isStrictThenable(typeChecker, node, part) })
	}
	if utils.IsUnionType(typ) {
		return utils.Every(typ.Types(), func(part *checker.Type) bool { return isStrictThenable(typeChecker, node, part) })
	}
	if utils.IsTypeParameter(typ) {
		constraint := checker.Checker_getBaseConstraintOfType(typeChecker, typ)
		return constraint != nil && isStrictThenable(typeChecker, node, constraint)
	}
	then := checker.Checker_getPropertyOfType(typeChecker, typ, "then")
	if then == nil {
		return false
	}
	for _, part := range utils.UnionTypeParts(typeChecker.GetTypeOfSymbolAtLocation(then, node)) {
		for _, signature := range checker.Checker_getSignaturesOfType(typeChecker, part, checker.SignatureKindCall) {
			parameters := checker.Signature_parameters(signature)
			if len(parameters) >= 2 && isFunctionParameter(typeChecker, node, parameters[0]) && isFunctionParameter(typeChecker, node, parameters[1]) {
				return true
			}
		}
	}
	return false
}

func isFunctionParameter(typeChecker *checker.Checker, node *ast.Node, parameter *ast.Symbol) bool {
	typ := checker.Checker_getApparentType(typeChecker, typeChecker.GetTypeOfSymbolAtLocation(parameter, node))
	return slices.ContainsFunc(utils.UnionTypeParts(typ), func(part *checker.Type) bool {
		return len(checker.Checker_getSignaturesOfType(typeChecker, part, checker.SignatureKindCall)) > 0
	})
}
