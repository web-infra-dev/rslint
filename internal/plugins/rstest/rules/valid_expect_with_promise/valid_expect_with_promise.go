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
			parsed := analysis.ParseExpectCall(node)
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
			promise := false
			if modifier != nil && modifier.Name == "rejects" {
				// Rstest invokes callable subjects before testing their returned value.
				promise = utils.Every(utils.UnionTypeParts(checker.Checker_getApparentType(ctx.TypeChecker, typ)), func(part *checker.Type) bool {
					signatures := checker.Checker_getSignaturesOfType(ctx.TypeChecker, part, checker.SignatureKindCall)
					if len(signatures) == 0 {
						return isPromise(subject, part)
					}
					return isPromise(subject, callableReturnType(ctx.TypeChecker, subject, part))
				})
			} else {
				promise = isPromise(subject, typ)
			}
			if promise && modifier == nil {
				ctx.ReportNode(parsed.Expression, rule.RuleMessage{Id: "poorlyExpectedPromise", Description: "Subject is a promise so resolve or reject should be used"})
			} else if !promise && modifier != nil {
				ctx.ReportNode(modifier.Node, rule.RuleMessage{Id: "unneededRejectResolve", Description: "Subject is not a promise so " + modifier.Name + " is not needed"})
			}
		}}
	},
}

func callableReturnType(typeChecker *checker.Checker, subject *ast.Node, typ *checker.Type) *checker.Type {
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
	return checker.Checker_getReturnTypeOfSignature(typeChecker, signature)
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
