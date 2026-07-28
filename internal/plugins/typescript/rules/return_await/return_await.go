package return_await

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildDisallowedPromiseAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "disallowedPromiseAwait",
		Description: "Returning an awaited promise is not allowed in this context.",
	}
}
func buildDisallowedPromiseAwaitSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "disallowedPromiseAwaitSuggestion",
		Description: "Remove `await` before the expression. Use caution as this may impact control flow.",
	}
}
func buildNonPromiseAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nonPromiseAwait",
		Description: "Returning an awaited value that is not a promise is not allowed.",
	}
}
func buildRequiredPromiseAwaitMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requiredPromiseAwait",
		Description: "Returning an awaited promise is required in this context.",
	}
}
func buildRequiredPromiseAwaitSuggestionMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "requiredPromiseAwaitSuggestion",
		Description: "Add `await` before the expression. Use caution as this may impact control flow.",
	}
}

type ReturnAwaitOption uint8

const (
	ReturnAwaitOptionAlways ReturnAwaitOption = iota
	ReturnAwaitOptionErrorHandlingCorrectnessOnly
	ReturnAwaitOptionInTryCatch
	ReturnAwaitOptionNever
)

type ReturnAwaitOptions struct {
	Option *ReturnAwaitOption
}

func parseReturnAwaitOption(options []any) ReturnAwaitOption {
	switch opts := rule.LegacyUnwrapOptions(options).(type) {
	case ReturnAwaitOptions:
		if opts.Option != nil {
			return *opts.Option
		}
	case *ReturnAwaitOptions:
		if opts != nil && opts.Option != nil {
			return *opts.Option
		}
	case string:
		switch opts {
		case "always":
			return ReturnAwaitOptionAlways
		case "error-handling-correctness-only":
			return ReturnAwaitOptionErrorHandlingCorrectnessOnly
		case "in-try-catch":
			return ReturnAwaitOptionInTryCatch
		case "never":
			return ReturnAwaitOptionNever
		}
	}

	return ReturnAwaitOptionInTryCatch
}

type scopeInfo struct {
	hasAsync   bool
	owningFunc *ast.Node
}

type containingTryStatementBlock uint8

const (
	containingTryStatementBlockCatch containingTryStatementBlock = iota
	containingTryStatementBlockFinally
	containingTryStatementBlockTry
)

type whetherToAwait uint8

const (
	whetherToAwaitDontCare whetherToAwait = iota
	whetherToAwaitAwait
	whetherToAwaitNoAwait
)

func getWhetherToAwait(affectsErrorHandling bool, option ReturnAwaitOption) whetherToAwait {
	switch option {
	case ReturnAwaitOptionAlways:
		return whetherToAwaitAwait
	case ReturnAwaitOptionErrorHandlingCorrectnessOnly:
		if affectsErrorHandling {
			return whetherToAwaitAwait
		}
		return whetherToAwaitDontCare
	case ReturnAwaitOptionInTryCatch:
		if affectsErrorHandling {
			return whetherToAwaitAwait
		}
		return whetherToAwaitNoAwait
	case ReturnAwaitOptionNever:
		return whetherToAwaitNoAwait
	default:
		panic("unexpected ReturnAwaitOption")
	}
}

func buildRemoveAwaitFix(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	return []rule.RuleFix{
		rule.RuleFixRemoveRange(scanner.GetRangeOfTokenAtPosition(sourceFile, node.Pos())),
	}
}

func buildInsertAwaitFix(sourceFile *ast.SourceFile, node *ast.Node) []rule.RuleFix {
	if utils.IsHigherPrecedenceThanAwait(node) {
		return []rule.RuleFix{
			rule.RuleFixInsertBefore(sourceFile, node, "await "),
		}
	}
	return []rule.RuleFix{
		rule.RuleFixInsertBefore(sourceFile, node, "await ("),
		rule.RuleFixInsertAfter(node, ")"),
	}
}

func reportNodeWithDeferredFixesOrSuggestions(
	ctx *rule.RuleContext,
	node *ast.Node,
	fix bool,
	msg rule.RuleMessage,
	suggestionMsg rule.RuleMessage,
	buildFixes func() []rule.RuleFix,
) {
	if fix {
		ctx.ReportNodeWithDeferredFixes(node, msg, buildFixes)
		return
	}

	ctx.ReportNodeWithDeferredSuggestions(node, msg, func() []rule.RuleSuggestion {
		return []rule.RuleSuggestion{{
			Message:  suggestionMsg,
			FixesArr: buildFixes(),
		}}
	})
}

var ReturnAwaitRule = rule.CreateRule(rule.Rule{
	Name:             "return-await",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		option := parseReturnAwaitOption(_options)
		sourceFile := ctx.SourceFile

		var scopes []scopeInfo

		enterFunction := func(node *ast.Node) {
			scopes = append(scopes, scopeInfo{
				hasAsync:   ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync),
				owningFunc: node,
			})
		}
		exitFunction := func(_ *ast.Node) {
			scopes = scopes[:len(scopes)-1]
		}

		affectsExplicitResourceManagement := func(node *ast.Node) bool {
			currentScope := scopes[len(scopes)-1]
			if !ast.IsBlock(currentScope.owningFunc.Body()) {
				return false
			}

			for declarationScope := ast.GetEnclosingBlockScopeContainer(node); declarationScope != nil; declarationScope = ast.GetEnclosingBlockScopeContainer(declarationScope) {
				locals := declarationScope.Locals()
				for _, local := range locals {
					decl := local.ValueDeclaration
					// if it's a using/await using declaration, and it comes _before_ the
					// node we're checking, it affects control flow for that node.

					if decl != nil && ast.IsVariableDeclaration(decl) && decl.Parent.Flags&ast.NodeFlagsUsing != 0 && decl.Pos() < node.Pos() {
						return true
					}
				}

				if currentScope.owningFunc == declarationScope {
					break
				}
			}

			return false
		}

		/**
		 * A try _statement_ is the whole thing that encompasses try block,
		 * catch clause, and finally block. This function finds the nearest
		 * enclosing try statement (if present) for a given node, and reports which
		 * part of the try statement the node is in.
		 */
		findContainingTryStatement := func(node *ast.Node) (containingTryStatementBlock, *ast.TryStatement, bool) {
			child := node
			ancestor := node.Parent

			for ancestor != nil && !ast.IsFunctionLike(ancestor) {
				if !ast.IsTryStatement(ancestor) {
					child = ancestor
					ancestor = ancestor.Parent
					continue
				}

				statement := ancestor.AsTryStatement()

				var block containingTryStatementBlock
				switch child {
				case statement.TryBlock:
					block = containingTryStatementBlockTry
				case statement.CatchClause:
					block = containingTryStatementBlockCatch
				case statement.FinallyBlock:
					block = containingTryStatementBlockFinally
				}

				return block, statement, true
			}

			return 0, nil, false
		}

		/**
		 * Tests whether a node is inside of an explicit error handling context
		 * (try/catch/finally) in a way that throwing an exception will have an
		 * impact on the program's control flow.
		 */
		affectsExplicitErrorHandling := func(node *ast.Node) bool {
			for {
				// If an error-handling block is followed by another error-handling block,
				// control flow is affected by whether promises in it are awaited or not.
				// Otherwise, we need to check recursively for nested try statements until
				// we get to the top level of a function or the program. If by then,
				// there's no offending error-handling blocks, it doesn't affect control
				// flow.
				containingTryBlock, tryStatement, found := findContainingTryStatement(node)
				if !found {
					return false
				}

				switch containingTryBlock {
				case containingTryStatementBlockCatch:
					// Exceptions thrown in catch blocks followed by a finally block affect
					// control flow.
					if tryStatement.FinallyBlock != nil {
						return true
					}
					// Otherwise recurse.
					node = tryStatement.AsNode()
				case containingTryStatementBlockFinally:
					node = tryStatement.AsNode()
				case containingTryStatementBlockTry:
					// Try blocks are always followed by either a catch or finally,
					// so exceptions thrown here always affect control flow.
					return true
				}
			}
		}

		test := func(node *ast.Node) {
			isAwait := ast.IsAwaitExpression(node)

			var affectsErrorHandling bool
			affectsErrorHandlingKnown := false
			if !isAwait {
				// A non-awaited expression can only be reported when this
				// context requires awaiting. Avoid asking the type checker when
				// the current option/context already makes the return valid.
				switch option {
				case ReturnAwaitOptionNever:
					return
				case ReturnAwaitOptionErrorHandlingCorrectnessOnly, ReturnAwaitOptionInTryCatch:
					affectsErrorHandling = affectsExplicitErrorHandling(node) || affectsExplicitResourceManagement(node)
					affectsErrorHandlingKnown = true
					if !affectsErrorHandling {
						return
					}
				}
			}

			child := node
			if isAwait {
				child = node.Expression()
			}
			t := ctx.TypeChecker.GetTypeAtLocation(child)
			certainty := utils.NeedsToBeAwaited(ctx.TypeChecker, node, t)

			if certainty != utils.TypeAwaitableAlways {
				if isAwait {
					if certainty == utils.TypeAwaitableMay {
						return
					}

					ctx.ReportNodeWithDeferredFixes(node, buildNonPromiseAwaitMessage(), func() []rule.RuleFix {
						return buildRemoveAwaitFix(sourceFile, node)
					})
				}
				return
			}

			// At this point it's definitely a thenable.
			if !affectsErrorHandlingKnown {
				affectsErrorHandling = affectsExplicitErrorHandling(node) || affectsExplicitResourceManagement(node)
			}
			useAutoFix := !affectsErrorHandling

			shouldAwaitInCurrentContext := getWhetherToAwait(affectsErrorHandling, option)

			switch shouldAwaitInCurrentContext {
			case whetherToAwaitAwait:
				if isAwait {
					break
				}
				reportNodeWithDeferredFixesOrSuggestions(
					&ctx,
					node,
					useAutoFix,
					buildRequiredPromiseAwaitMessage(),
					buildRequiredPromiseAwaitSuggestionMessage(),
					func() []rule.RuleFix {
						return buildInsertAwaitFix(sourceFile, node)
					},
				)
			case whetherToAwaitDontCare:
				break
			case whetherToAwaitNoAwait:
				if !isAwait {
					break
				}
				reportNodeWithDeferredFixesOrSuggestions(
					&ctx,
					node,
					useAutoFix,
					buildDisallowedPromiseAwaitMessage(),
					buildDisallowedPromiseAwaitSuggestionMessage(),
					func() []rule.RuleFix {
						return buildRemoveAwaitFix(sourceFile, node)
					},
				)
			}
		}

		var testEachPossiblyReturnedNode func(node *ast.Node)
		testEachPossiblyReturnedNode = func(node *ast.Node) {
			node = ast.SkipParentheses(node)
			if node.Kind == ast.KindConditionalExpression {
				expr := node.AsConditionalExpression()
				testEachPossiblyReturnedNode(expr.WhenFalse)
				testEachPossiblyReturnedNode(expr.WhenTrue)
			} else {
				test(node)
			}
		}

		return rule.RuleListeners{
			ast.KindArrowFunction:       enterFunction,
			ast.KindFunctionDeclaration: enterFunction,
			ast.KindFunctionExpression:  enterFunction,
			ast.KindMethodDeclaration:   enterFunction,
			ast.KindConstructor:         enterFunction,
			ast.KindGetAccessor:         enterFunction,
			ast.KindSetAccessor:         enterFunction,

			rule.ListenerOnExit(ast.KindArrowFunction): func(node *ast.Node) {
				body := node.Body()
				// Expression-bodied arrows have no ReturnStatement listener to
				// apply the async-function guard for us.
				if scopes[len(scopes)-1].hasAsync && !ast.IsBlock(body) {
					testEachPossiblyReturnedNode(body)
				}

				exitFunction(node)
			},
			rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
			rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
			rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,
			rule.ListenerOnExit(ast.KindConstructor):         exitFunction,
			rule.ListenerOnExit(ast.KindGetAccessor):         exitFunction,
			rule.ListenerOnExit(ast.KindSetAccessor):         exitFunction,

			ast.KindReturnStatement: func(node *ast.Node) {
				expr := node.AsReturnStatement().Expression
				if len(scopes) == 0 || !scopes[len(scopes)-1].hasAsync || expr == nil {
					return
				}

				testEachPossiblyReturnedNode(expr)
			},
		}
	},
})
