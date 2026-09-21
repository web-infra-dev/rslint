// Package prefer_expect_resolves shares await detection and conservative edits
// while adapters retain ownership of framework provenance and matcher semantics.
package prefer_expect_resolves

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	framework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type ExpectCall struct {
	Head, Matcher *ast.Node
	AddResolves   bool
	// Editable excludes matchers whose promise mode or return value differs.
	Editable bool
}

type Config struct {
	Name              string
	Message           rule.RuleMessage
	RequirePromise    bool
	ConservativeEdits bool
	Prepare           func(rule.RuleContext) func(*ast.Node) *ExpectCall
}

var suggestion = rule.RuleMessage{Id: "suggestExpectResolves", Description: "Replace with an awaited resolves assertion"}

func NewRule(config Config) rule.Rule {
	return rule.Rule{Name: config.Name, Schema: rule.EmptyArraySchema,
		Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
			parse := config.Prepare(ctx)
			return rule.RuleListeners{ast.KindCallExpression: func(node *ast.Node) {
				call := parse(node)
				if call == nil || call.Head == nil || call.Matcher == nil || len(call.Head.Arguments()) == 0 {
					return
				}
				awaited := ast.SkipParentheses(call.Head.Arguments()[0])
				if awaited == nil || awaited.Kind != ast.KindAwaitExpression {
					return
				}
				subject := awaited.AsAwaitExpression().Expression
				if subject == nil {
					return
				}
				promise := false
				if config.RequirePromise {
					known := false
					known, promise = promiseType(ctx, subject)
					if known && !promise {
						return
					}
				}
				build := func() []rule.RuleFix {
					return buildFixes(ctx, call, awaited, config.ConservativeEdits)
				}
				ctx.ReportNodeWithDeferredFixesAndSuggestions(awaited, config.Message,
					func() []rule.RuleFix {
						if config.RequirePromise && !promise {
							return nil
						}
						return build()
					},
					func() []rule.RuleSuggestion {
						if !config.RequirePromise || promise {
							return nil
						}
						fixes := build()
						if len(fixes) == 0 {
							return nil
						}
						return []rule.RuleSuggestion{{Message: suggestion, FixesArr: fixes}}
					})
			}}
		}}
}

// A union must be entirely thenable. Await accepts scalars, but resolves does
// not. Unknown/error types retain the syntax diagnostic without an autofix.
func promiseType(ctx rule.RuleContext, subject *ast.Node) (known, promise bool) {
	if ctx.TypeChecker == nil {
		return false, false
	}
	typ := ctx.TypeChecker.GetTypeAtLocation(subject)
	if typ == nil {
		return false, false
	}
	if utils.IsTypeParameter(typ) {
		typ = checker.Checker_getBaseConstraintOfType(ctx.TypeChecker, typ)
		if typ == nil {
			return false, false
		}
	}
	for _, part := range utils.UnionTypeParts(typ) {
		if part.Flags()&checker.TypeFlagsNever != 0 {
			return true, false
		}
		if part.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
			return false, false
		}
	}
	for _, part := range utils.UnionTypeParts(typ) {
		if !utils.IsThenableType(ctx.TypeChecker, subject, part) {
			return true, false
		}
	}
	return true, true
}

// Literal arguments cannot observe the promise settling or execute user code.
// Even an identifier can change while suspended, so it is not safe to move.
func literal(node *ast.Node) bool {
	node = utils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
		ast.KindNoSubstitutionTemplateLiteral:
		return true
	default:
		return false
	}
}

func buildFixes(ctx rule.RuleContext, call *ExpectCall, awaited *ast.Node, conservative bool) []rule.RuleFix {
	if !call.Editable {
		return nil
	}
	if conservative {
		if call.Head.AsCallExpression().TypeArguments != nil {
			return nil
		}
		for _, arg := range call.Head.Arguments()[1:] {
			if !literal(arg) {
				return nil
			}
		}
		for _, arg := range call.Matcher.Arguments() {
			if !literal(arg) {
				return nil
			}
		}
		// Guard every link, not just the final call: a missing optional receiver
		// otherwise loses its original await suspension and short-circuit behavior.
		for current := call.Matcher; current != nil; {
			if current.Kind == ast.KindCallExpression && current.AsCallExpression().QuestionDotToken != nil {
				return nil
			}
			if framework.AccessorQuestionDotToken(current) != nil {
				return nil
			}
			switch current.Kind {
			case ast.KindCallExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
				ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression, ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
				current = current.Expression()
			default:
				current = nil
			}
		}
		// Only a discarded result or an already-awaited/returned assertion is
		// rewritten. Chai matcher return values must not change inside expressions.
		outer := call.Matcher
		for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
			outer = outer.Parent
		}
		if outer.Parent == nil {
			return nil
		}
		switch outer.Parent.Kind {
		case ast.KindExpressionStatement, ast.KindReturnStatement, ast.KindAwaitExpression:
		default:
			return nil
		}
	}
	start := utils.TrimNodeTextRange(ctx.SourceFile, awaited).Pos()
	// Remove the keyword alone. Comments and parentheses after it stay intact.
	fixes := []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, start+len("await")))}
	// Remove one plain space for conventional output; preserve line breaks/trivia.
	text := ctx.SourceFile.Text()
	if start+5 < len(text) && text[start+5] == ' ' {
		fixes[0] = rule.RuleFixRemoveRange(core.NewTextRange(start, start+6))
	}
	if call.AddResolves {
		fixes = append(fixes, rule.RuleFixInsertAfter(call.Head, ".resolves"))
	}
	outer := call.Matcher
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent.Kind != ast.KindAwaitExpression {
		fixes = append(fixes, rule.RuleFixInsertBefore(ctx.SourceFile, outer, "await "))
	}
	return fixes
}

// BuiltinValueMatcher lists the common value matchers whose promise mode only
// unwraps the subject. Throw/snapshot/custom matchers need separate semantics.
func BuiltinValueMatcher(name string) bool {
	switch name {
	case "toBe", "toEqual", "toStrictEqual", "toBeDefined", "toBeUndefined",
		"toBeNull", "toBeTruthy", "toBeFalsy", "toBeNaN", "toBeGreaterThan",
		"toBeGreaterThanOrEqual", "toBeLessThan", "toBeLessThanOrEqual",
		"toBeCloseTo", "toContain", "toContainEqual", "toHaveLength", "toMatch",
		"toMatchObject", "toHaveProperty", "toBeInstanceOf":
		return true
	default:
		return false
	}
}
