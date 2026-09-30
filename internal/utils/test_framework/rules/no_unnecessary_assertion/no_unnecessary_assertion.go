// Package no_unnecessary_assertion holds the rule shell shared by
// jest/no-unnecessary-assertion and rstest/no-unnecessary-assertion: the
// strictNullChecks precondition, the checked matchers and their diagnostics.
// Framework adapters own expect-call parsing and decide, from the subject's
// type, whether an assertion can ever observe the checked value.
package no_unnecessary_assertion

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Matcher describes the value a checked matcher looks for. Flags selects the
// subject types that can hold that value and Thing names it in the message.
type Matcher struct {
	Flags checker.TypeFlags
	Thing string
}

var (
	nullMatcher      = Matcher{Flags: checker.TypeFlagsNull, Thing: "null"}
	undefinedMatcher = Matcher{Flags: checker.TypeFlagsUndefined, Thing: "undefined"}
	nanMatcher       = Matcher{Flags: checker.TypeFlagsNumberLike, Thing: "a number"}
)

var callMatchers = map[string]Matcher{
	"toBeNull":      nullMatcher,
	"toBeUndefined": undefinedMatcher,
	"toBeDefined":   undefinedMatcher,
	"toBeNaN":       nanMatcher,
}

var propertyMatchers = map[string]Matcher{
	"null":      nullMatcher,
	"undefined": undefinedMatcher,
	"NaN":       nanMatcher,
}

// CallMatcher looks up a checked matcher that is invoked, such as toBeNull().
func CallMatcher(name string) (Matcher, bool) {
	matcher, ok := callMatchers[name]
	return matcher, ok
}

// PropertyMatcher looks up a checked Chai property assertion, such as .null.
func PropertyMatcher(name string) (Matcher, bool) {
	matcher, ok := propertyMatchers[name]
	return matcher, ok
}

// OnlyNotModifiers reports whether every modifier is `not`. Any other modifier
// (resolves, rejects, ...) changes the value the matcher sees.
func OnlyNotModifiers(modifiers []string) bool {
	for _, modifier := range modifiers {
		if modifier != "not" {
			return false
		}
	}
	return true
}

// Assertion is an assertion the adapter found unnecessary.
type Assertion struct {
	Node    *ast.Node
	Matcher Matcher
}

type Runtime struct {
	// Check returns the assertion to report for node, or nil.
	Check func(*ast.Node) *Assertion
}

type Config struct {
	Name    string
	Prepare func(rule.RuleContext) Runtime
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

func NewRule(config Config) rule.Rule {
	return rule.Rule{
		Name:             config.Name,
		Schema:           rule.EmptyArraySchema,
		RequiresTypeInfo: true,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			compilerOptions := ctx.Program().Options()
			if !utils.IsStrictCompilerOptionEnabled(compilerOptions, compilerOptions.StrictNullChecks) {
				ctx.ReportRange(core.NewTextRange(0, 0), noStrictNullCheckMessage())
			}

			runtime := config.Prepare(ctx)
			return rule.RuleListeners{
				ast.KindCallExpression: func(node *ast.Node) {
					assertion := runtime.Check(node)
					if assertion == nil {
						return
					}
					ctx.ReportNode(assertion.Node, unnecessaryAssertionMessage(assertion.Matcher.Thing))
				},
			}
		},
	}
}
