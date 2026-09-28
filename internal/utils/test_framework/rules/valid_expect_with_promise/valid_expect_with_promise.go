// Package valid_expect_with_promise shares the options, diagnostics, and
// strict thenable check of valid-expect-with-promise. Adapters retain
// ownership of expect-chain parsing and of what counts as a Promise subject.
package valid_expect_with_promise

import (
	_ "embed"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed valid_expect_with_promise.schema.json
var schemaJSON []byte

var Schema = rule.NewSchema(schemaJSON)

type Options struct {
	CheckThenables bool
}

func ParseOptions(options []any) Options {
	var opts Options
	if len(options) > 0 {
		if option, ok := options[0].(map[string]any); ok {
			opts.CheckThenables, _ = option["checkThenables"].(bool)
		}
	}
	return opts
}

var PoorlyExpectedPromiseMessage = rule.RuleMessage{
	Id:          "poorlyExpectedPromise",
	Description: "Subject is a promise so resolve or reject should be used",
}

func UnneededRejectResolveMessage(modifier string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unneededRejectResolve",
		Description: "Subject is not a promise so " + modifier + " is not needed",
		Data:        map[string]string{"modifier": modifier},
	}
}

// IsStrictThenable reports whether typ has a `then` method accepting both a
// fulfillment and a rejection callback. A single-callback chainable is
// deliberately excluded, unlike utils.IsThenableType.
func IsStrictThenable(typeChecker *checker.Checker, node *ast.Node, typ *checker.Type) bool {
	if utils.IsIntersectionType(typ) {
		return slices.ContainsFunc(typ.Types(), func(part *checker.Type) bool { return IsStrictThenable(typeChecker, node, part) })
	}
	if utils.IsUnionType(typ) {
		return utils.Every(typ.Types(), func(part *checker.Type) bool { return IsStrictThenable(typeChecker, node, part) })
	}
	if utils.IsTypeParameter(typ) {
		constraint := checker.Checker_getBaseConstraintOfType(typeChecker, typ)
		return constraint != nil && IsStrictThenable(typeChecker, node, constraint)
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
