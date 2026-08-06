package preserve_caught_error

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed preserve_caught_error.schema.json
var schemaJSON []byte

// Global error constructors that accept an options bag with a `cause`.
var builtInErrorTypes = map[string]bool{
	"Error":          true,
	"EvalError":      true,
	"RangeError":     true,
	"ReferenceError": true,
	"SyntaxError":    true,
	"TypeError":      true,
	"URIError":       true,
	"AggregateError": true,
}

// defaultArgumentPosition is the 1-based options position assumed for an
// `errorClassNames` entry given as a bare string.
const defaultArgumentPosition = 2

type Options struct {
	RequireCatchParameter bool
	// ErrorClassNames maps a configured error class name to the 1-based
	// position of its error-options argument.
	ErrorClassNames map[string]int
}

func parseOptions(options []any) Options {
	opts := Options{ErrorClassNames: map[string]int{}}
	optsMap := utils.GetOptionsMap(options)
	if optsMap == nil {
		return opts
	}
	opts.RequireCatchParameter, _ = optsMap["requireCatchParameter"].(bool)
	items, _ := optsMap["errorClassNames"].([]interface{})
	for _, item := range items {
		switch entry := item.(type) {
		case string:
			opts.ErrorClassNames[entry] = defaultArgumentPosition
		case map[string]interface{}:
			name, ok := entry["name"].(string)
			if !ok {
				continue
			}
			position, ok := utils.CoerceIntegral(entry["argumentPosition"])
			if !ok {
				continue
			}
			opts.ErrorClassNames[name] = position
		}
	}
	return opts
}

func messageMissingCause() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingCause",
		Description: "There is no `cause` attached to the symptom error being thrown.",
	}
}

func messageIncorrectCause() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "incorrectCause",
		Description: "The symptom error is being thrown with an incorrect `cause`.",
	}
}

func messageIncludeCause() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "includeCause",
		Description: "Include the original caught error as the `cause` of the symptom error.",
	}
}

func messageMissingCatchErrorParam() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "missingCatchErrorParam",
		Description: "The caught error is not accessible because the catch clause lacks the error parameter. Start referencing the caught error using the catch parameter.",
	}
}

func messagePartiallyLostError() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "partiallyLostError",
		Description: "Re-throws cannot preserve the caught error as a part of it is being lost due to destructuring.",
	}
}

func messageCaughtErrorShadowed() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "caughtErrorShadowed",
		Description: "The caught error is being attached as `cause`, but is shadowed by a closer scoped redeclaration.",
	}
}

// findParentCatch returns the `catch` clause the node re-throws from, or nil
// when the node is outside a catch clause or separated from it by a function
// or class static block — the caught error is not directly related to a throw
// nested in one of those.
func findParentCatch(node *ast.Node) *ast.Node {
	for current := node; current != nil; current = current.Parent {
		if current.Kind == ast.KindCatchClause {
			return current
		}
		switch current.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor,
			ast.KindClassStaticBlockDeclaration,
			ast.KindSourceFile:
			return nil
		}
	}
	return nil
}

// isBuiltInGlobalError reports whether the callee refers to one of the global
// error constructors. A local declaration or a disabled global un-declares the
// name, so the built-in signature can no longer be assumed.
func isBuiltInGlobalError(ctx rule.RuleContext, callee *ast.Node) bool {
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return false
	}
	name := callee.AsIdentifier().Text
	if !builtInErrorTypes[name] || utils.IsShadowed(callee, name) {
		return false
	}
	return ctx.Globals.Access(name).IsDeclared()
}

// thrownError describes a `throw` of a newly constructed error the rule checks.
type thrownError struct {
	// expression is the CallExpression / NewExpression being thrown, with any
	// wrapping parentheses removed.
	expression *ast.Node
	// callee is the constructor expression as written, parentheses included.
	callee    *ast.Node
	arguments *ast.NodeList
	className string
	builtIn   bool
	// optionsIndex is the 0-based argument position of the error options bag.
	optionsIndex int
}

func (t thrownError) args() []*ast.Node {
	if t.arguments == nil {
		return nil
	}
	return t.arguments.Nodes
}

// classifyThrownError inspects the thrown expression and reports whether it
// constructs an error whose `cause` this rule checks.
func classifyThrownError(ctx rule.RuleContext, expression *ast.Node, opts Options) (thrownError, bool) {
	var info thrownError

	switch expression.Kind {
	case ast.KindNewExpression:
		newExpr := expression.AsNewExpression()
		info.callee, info.arguments = newExpr.Expression, newExpr.Arguments
	case ast.KindCallExpression:
		callExpr := expression.AsCallExpression()
		info.callee, info.arguments = callExpr.Expression, callExpr.Arguments
	default:
		return info, false
	}
	// ESLint sees an optional chain as a `ChainExpression`, which is neither a
	// `CallExpression` nor a `NewExpression`, so those never reach the rule.
	if info.callee == nil || ast.IsOptionalChain(expression) {
		return info, false
	}
	info.expression = expression

	callee := ast.SkipParentheses(info.callee)
	// Parentheses end an optional chain, so a chain used as the callee stays a
	// `ChainExpression` in ESLint's AST — `(ns?.AppError)('m')` matches neither
	// a global error name nor a configured class.
	if ast.IsOptionalChain(callee) {
		return info, false
	}
	if isBuiltInGlobalError(ctx, callee) {
		info.builtIn = true
		info.className = callee.AsIdentifier().Text
		info.optionsIndex = 1
		if info.className == "AggregateError" {
			info.optionsIndex = 2
		}
		return info, true
	}

	target := callee
	if callee.Kind == ast.KindPropertyAccessExpression {
		target = callee.AsPropertyAccessExpression().Name()
	}
	if target == nil || target.Kind != ast.KindIdentifier {
		return info, false
	}
	name := target.AsIdentifier().Text
	position, ok := opts.ErrorClassNames[name]
	if !ok {
		return info, false
	}
	info.className = name
	info.optionsIndex = position - 1
	return info, true
}

type causeStatus int

const (
	// causeUnknown means the error options are too complicated to analyze.
	causeUnknown causeStatus = iota
	causeMissing
	causeFound
)

type causeInfo struct {
	status causeStatus
	// member is the object-literal member declaring `cause`.
	member *ast.Node
	// value is the `cause` value, with wrapping parentheses removed. It is nil
	// when `cause` is declared as a method or accessor.
	value               *ast.Node
	multipleDefinitions bool
}

// getErrorCause finds the `cause` declared in the thrown error's options bag.
func getErrorCause(info thrownError) causeInfo {
	args := info.args()

	// A spread at or before the options position shifts the effective argument
	// order, so where the error options actually land is no longer known.
	for index, arg := range args {
		if arg.Kind == ast.KindSpreadElement {
			if index <= info.optionsIndex {
				return causeInfo{status: causeUnknown}
			}
			break
		}
	}

	if info.optionsIndex < 0 || info.optionsIndex >= len(args) {
		return causeInfo{status: causeMissing}
	}

	errorOptions := ast.SkipParentheses(args[info.optionsIndex])
	if errorOptions.Kind != ast.KindObjectLiteralExpression {
		return causeInfo{status: causeUnknown}
	}
	properties := errorOptions.AsObjectLiteralExpression().Properties
	if properties == nil {
		return causeInfo{status: causeMissing}
	}

	var causeMember *ast.Node
	causeCount := 0
	for _, member := range properties.Nodes {
		if member.Kind == ast.KindSpreadAssignment {
			// A spread can contribute a `cause` that is impossible to verify.
			return causeInfo{status: causeUnknown}
		}
		name := member.Name()
		if name == nil {
			continue
		}
		if propertyName, ok := utils.GetStaticPropertyName(name); ok && propertyName == "cause" {
			causeMember = member
			causeCount++
		}
	}
	if causeMember == nil {
		return causeInfo{status: causeMissing}
	}
	return causeInfo{
		status:              causeFound,
		member:              causeMember,
		value:               causeValue(causeMember),
		multipleDefinitions: causeCount > 1,
	}
}

// causeValue returns the value node of an object-literal member, or nil for
// methods and accessors, which hold their body instead of a value expression.
func causeValue(member *ast.Node) *ast.Node {
	switch member.Kind {
	case ast.KindPropertyAssignment:
		initializer := member.AsPropertyAssignment().Initializer
		if initializer == nil {
			return nil
		}
		return ast.SkipParentheses(initializer)
	case ast.KindShorthandPropertyAssignment:
		return member.Name()
	default:
		return nil
	}
}

// causeReportRange is the range ESLint reports an incorrect `cause` on. For a
// method or accessor that is the function ESLint stores as the member's value,
// which starts at the parameter list and runs to the end of the body.
func causeReportRange(ctx rule.RuleContext, cause causeInfo) core.TextRange {
	if cause.value != nil {
		return utils.TrimNodeTextRange(ctx.SourceFile, cause.value)
	}
	start := cause.member.End()
	if name := cause.member.Name(); name != nil {
		start = scanner.SkipTrivia(ctx.SourceFile.Text(), name.End())
	}
	return core.NewTextRange(start, cause.member.End())
}

func insertAt(pos int, text string) rule.RuleFix {
	return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), text)
}

// appendArguments inserts text after the last argument as written, so the
// insertion lands outside any parentheses wrapping that argument.
func appendArguments(info thrownError, text string) []rule.RuleFix {
	args := info.args()
	return []rule.RuleFix{rule.RuleFixInsertAfter(args[len(args)-1], text)}
}

// addArgumentsToEmptyCall inserts text as the whole argument list, whether the
// call already has empty parentheses (`new Error()`) or none (`new Error`).
func addArgumentsToEmptyCall(info thrownError, text string) []rule.RuleFix {
	if info.arguments != nil {
		return []rule.RuleFix{insertAt(info.arguments.Pos(), text)}
	}
	return []rule.RuleFix{rule.RuleFixInsertAfter(info.expression, "("+text+")")}
}

// insertCauseIntoOptions adds `cause` to an existing inline options object.
func insertCauseIntoOptions(optionsArg *ast.Node, caughtName string) []rule.RuleFix {
	options := ast.SkipParentheses(optionsArg)
	if options.Kind != ast.KindObjectLiteralExpression {
		return nil
	}
	properties := options.AsObjectLiteralExpression().Properties
	if properties == nil {
		return nil
	}
	if len(properties.Nodes) == 0 {
		return []rule.RuleFix{insertAt(properties.Pos(), "cause: "+caughtName)}
	}
	lastProperty := properties.Nodes[len(properties.Nodes)-1]
	return []rule.RuleFix{rule.RuleFixInsertAfter(lastProperty, ", cause: "+caughtName)}
}

// buildMissingCauseFixes attaches the caught error as the `cause` of an error
// thrown without one, filling in any positional arguments it has to skip over.
func buildMissingCauseFixes(info thrownError, caughtName string) []rule.RuleFix {
	args := info.args()
	options := "{ cause: " + caughtName + " }"

	if info.builtIn && info.className == "AggregateError" {
		switch len(args) {
		case 0:
			return addArgumentsToEmptyCall(info, `[], "", `+options)
		case 1:
			return appendArguments(info, `, "", `+options)
		case 2:
			return appendArguments(info, ", "+options)
		default:
			return insertCauseIntoOptions(args[2], caughtName)
		}
	}

	if info.builtIn {
		switch len(args) {
		case 0:
			return addArgumentsToEmptyCall(info, `"", `+options)
		case 1:
			return appendArguments(info, ", "+options)
		default:
			return insertCauseIntoOptions(args[1], caughtName)
		}
	}

	// A custom error's signature is unknown, so skip the suggestion rather than
	// synthesize placeholder values for missing positional arguments.
	if len(args) < info.optionsIndex {
		return nil
	}
	if info.optionsIndex >= len(args) {
		if len(args) > 0 {
			return appendArguments(info, ", "+options)
		}
		return addArgumentsToEmptyCall(info, options)
	}
	return insertCauseIntoOptions(args[info.optionsIndex], caughtName)
}

// buildIncorrectCauseFixes replaces the wrong `cause` with the caught error.
// Shorthand, method, and accessor members carry no plain value to replace, so
// the whole member becomes a `cause: <caught>` property.
func buildIncorrectCauseFixes(ctx rule.RuleContext, cause causeInfo, caughtName string) []rule.RuleFix {
	switch cause.member.Kind {
	case ast.KindShorthandPropertyAssignment,
		ast.KindMethodDeclaration,
		ast.KindGetAccessor,
		ast.KindSetAccessor:
		return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, cause.member, "cause: "+caughtName)}
	}
	if cause.value == nil {
		return nil
	}
	return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, cause.value, caughtName)}
}

// isCaughtErrorShadowed reports whether the identifier attached as `cause`
// resolves to a binding other than the catch parameter it is spelled like. A
// `var` sharing the name is hoisted out of the catch clause and still resolves
// to the parameter, so only a closer block-scoped binding shadows it.
func isCaughtErrorShadowed(ctx rule.RuleContext, causeValue *ast.Node, param *ast.Node) bool {
	caught := param.Symbol()
	if caught == nil {
		return false
	}
	target := ctx.Refs.Resolve(causeValue)
	return target != nil && target != caught
}

func suggestIncludeCause(fixes []rule.RuleFix) []rule.RuleSuggestion {
	if len(fixes) == 0 {
		return nil
	}
	return []rule.RuleSuggestion{{Message: messageIncludeCause(), FixesArr: fixes}}
}

// https://eslint.org/docs/latest/rules/preserve-caught-error
var PreserveCaughtErrorRule = rule.Rule{
	Name:   "preserve-caught-error",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindThrowStatement: func(node *ast.Node) {
				throwStatement := node.AsThrowStatement()
				if throwStatement.Expression == nil {
					return
				}
				parentCatch := findParentCatch(node)
				if parentCatch == nil {
					return
				}
				info, ok := classifyThrownError(ctx, ast.SkipParentheses(throwStatement.Expression), opts)
				if !ok {
					return
				}

				param := parentCatch.AsCatchClause().VariableDeclaration
				if param != nil && param.Name().Kind != ast.KindIdentifier {
					// Destructuring the caught error at the parameter level
					// loses everything the pattern doesn't name.
					ctx.ReportNode(parentCatch, messagePartiallyLostError())
					return
				}
				if param == nil {
					if opts.RequireCatchParameter {
						ctx.ReportNode(node, messageMissingCatchErrorParam())
					}
					return
				}
				caughtName := param.Name().AsIdentifier().Text

				cause := getErrorCause(info)
				switch cause.status {
				case causeUnknown:
					return
				case causeMissing:
					ctx.ReportNodeWithDeferredSuggestions(node, messageMissingCause(), func() []rule.RuleSuggestion {
						return suggestIncludeCause(buildMissingCauseFixes(info, caughtName))
					})
					return
				}

				if cause.value == nil || cause.value.Kind != ast.KindIdentifier ||
					cause.value.AsIdentifier().Text != caughtName {
					ctx.ReportRangeWithDeferredSuggestions(causeReportRange(ctx, cause), messageIncorrectCause(), func() []rule.RuleSuggestion {
						if cause.multipleDefinitions {
							// With several `cause` definitions a suggestion
							// would be more confusing than helpful.
							return nil
						}
						return suggestIncludeCause(buildIncorrectCauseFixes(ctx, cause, caughtName))
					})
					return
				}

				if isCaughtErrorShadowed(ctx, cause.value, param) {
					ctx.ReportNode(node, messageCaughtErrorShadowed())
				}
			},
		}
	},
}
