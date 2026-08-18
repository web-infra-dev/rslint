package max_statements

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed max_statements.schema.json
var schemaJSON []byte

const defaultMax = 10

// MaxStatementsRule enforces a maximum number of statements allowed in
// function blocks.
// https://eslint.org/docs/latest/rules/max-statements
var MaxStatementsRule = rule.Rule{
	Name:   "max-statements",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		// stack mirrors upstream's functionStack: one counter per open
		// function-like / class-static-block code path, incremented by every
		// nested KindBlock (ESTree BlockStatement) it directly or indirectly
		// contains.
		var stack []int

		type topLevelEntry struct {
			node  *ast.Node
			count int
		}
		var topLevelFunctions []topLevelEntry
		headLocator := functionHeadLocator{sourceFile: ctx.SourceFile}

		report := func(node *ast.Node, count int) {
			if count <= opts.max {
				return
			}
			name := utils.UpperCaseFirstASCII(utils.GetFunctionNameWithKindCore(node))
			ctx.ReportRange(headLocator.loc(node), rule.RuleMessage{
				Id: "exceed",
				Description: fmt.Sprintf(
					"%s has too many statements (%d). Maximum allowed is %d.",
					name, count, opts.max,
				),
			})
		}

		// ESLint listens for FunctionDeclaration / FunctionExpression /
		// ArrowFunctionExpression, which between them cover methods,
		// constructors, getters, and setters through ESTree's MethodDefinition
		// wrapper. tsgo represents those function-likes directly, so each kind
		// is wired individually.
		startFunc := func(node *ast.Node) {
			// Overload signatures and abstract/declare members reach this
			// listener with no body. ESLint never visits them as
			// FunctionDeclaration/FunctionExpression (they parse as
			// TSDeclareFunction / body-less MethodDefinition instead), so skip
			// pushing a frame for parity — and skip popping it in endFunc.
			if node.Body() == nil {
				return
			}
			stack = append(stack, 0)
		}
		endFunc := func(node *ast.Node) {
			if node.Body() == nil {
				return
			}
			n := len(stack) - 1
			count := stack[n]
			stack = stack[:n]

			if opts.ignoreTopLevelFunctions && len(stack) == 0 {
				topLevelFunctions = append(topLevelFunctions, topLevelEntry{node, count})
			} else {
				report(node, count)
			}
		}

		// tsgo keeps a class member's decorators and computed name inside the
		// member node, so ForEachChild walks them while the member's frame is
		// open. ESTree hangs both off MethodDefinition — a sibling of the
		// FunctionExpression this rule listens for — so a function written
		// there is not nested in the member. Hide the member's frame while
		// traversing that metadata so nesting depth, and with it the
		// `ignoreTopLevelFunctions` bookkeeping, matches upstream.
		var hiddenFrames []int
		hideMemberFrame := func(node *ast.Node) {
			if !isMemberMetadata(node) {
				return
			}
			n := len(stack) - 1
			hiddenFrames = append(hiddenFrames, stack[n])
			stack = stack[:n]
		}
		restoreMemberFrame := func(node *ast.Node) {
			if !isMemberMetadata(node) {
				return
			}
			n := len(hiddenFrames) - 1
			stack = append(stack, hiddenFrames[n])
			hiddenFrames = hiddenFrames[:n]
		}

		// A class static block always has a body and never reports — it is
		// tracked only so statements inside it do not leak into the enclosing
		// function's count.
		startStaticBlock := func(*ast.Node) {
			stack = append(stack, 0)
		}
		endStaticBlock := func(*ast.Node) {
			stack = stack[:len(stack)-1]
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration:                              startFunc,
			rule.ListenerOnExit(ast.KindFunctionDeclaration):         endFunc,
			ast.KindFunctionExpression:                               startFunc,
			rule.ListenerOnExit(ast.KindFunctionExpression):          endFunc,
			ast.KindArrowFunction:                                    startFunc,
			rule.ListenerOnExit(ast.KindArrowFunction):               endFunc,
			ast.KindMethodDeclaration:                                startFunc,
			rule.ListenerOnExit(ast.KindMethodDeclaration):           endFunc,
			ast.KindGetAccessor:                                      startFunc,
			rule.ListenerOnExit(ast.KindGetAccessor):                 endFunc,
			ast.KindSetAccessor:                                      startFunc,
			rule.ListenerOnExit(ast.KindSetAccessor):                 endFunc,
			ast.KindConstructor:                                      startFunc,
			rule.ListenerOnExit(ast.KindConstructor):                 endFunc,
			ast.KindClassStaticBlockDeclaration:                      startStaticBlock,
			rule.ListenerOnExit(ast.KindClassStaticBlockDeclaration): endStaticBlock,

			ast.KindDecorator:                                 hideMemberFrame,
			rule.ListenerOnExit(ast.KindDecorator):            restoreMemberFrame,
			ast.KindComputedPropertyName:                      hideMemberFrame,
			rule.ListenerOnExit(ast.KindComputedPropertyName): restoreMemberFrame,

			// tsgo wraps a class static block's body in a real Block node (not
			// flattened the way ESTree flattens StaticBlock.body), so counting
			// every KindBlock's direct statements uniformly reproduces
			// upstream's countStatements without a StaticBlock special case.
			ast.KindBlock: func(node *ast.Node) {
				if len(stack) == 0 {
					return
				}
				block := node.AsBlock()
				if block == nil || block.Statements == nil {
					return
				}
				stack[len(stack)-1] += len(block.Statements.Nodes)
			},

			// SourceFile.ForEachChild visits each statement followed by the
			// EndOfFileToken — using its listener as the "Program:exit"
			// equivalent guarantees every top-level function has been counted
			// before we decide whether to report deferred entries.
			ast.KindEndOfFile: func(*ast.Node) {
				if len(topLevelFunctions) == 1 {
					return
				}
				for _, entry := range topLevelFunctions {
					report(entry.node, entry.count)
				}
			},
		}
	},
}

// isClassMember reports whether node is one of the function-like kinds that
// ESTree models as a `MethodDefinition` (or object-literal `Property`) wrapping
// a `FunctionExpression`, i.e. the kinds whose decorators and computed name
// live outside the function in upstream's AST.
func isClassMember(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return true
	default:
		return false
	}
}

// isMemberMetadata reports whether node is a decorator or computed name that
// belongs to a member this rule opened a frame for.
func isMemberMetadata(node *ast.Node) bool {
	parent := node.Parent
	return parent != nil && isClassMember(parent) && parent.Body() != nil
}

// hasDecorator reports whether node carries at least one decorator.
func hasDecorator(node *ast.Node) bool {
	mods := node.Modifiers()
	if mods == nil {
		return false
	}
	for _, mod := range mods.Nodes {
		if mod.Kind == ast.KindDecorator {
			return true
		}
	}
	return false
}

// functionHeadLocator mirrors core ESLint's getFunctionHeadLoc, which starts
// a class member's report at the member's own start — decorators included. The
// shared helper follows typescript-eslint's variant, which deliberately skips
// them, so the locator restores the member start for decorated members.
type functionHeadLocator struct {
	sourceFile *ast.SourceFile
	tokens     *scanner.Scanner
}

func (l *functionHeadLocator) loc(node *ast.Node) core.TextRange {
	// GetFunctionHeadLoc has to reproduce SourceCode.getTokenBefore() for a
	// single-parameter field arrow. Its general implementation scans from the
	// start of the source file, which becomes quadratic when this rule reports
	// many such fields in one file. The owning field gives us a narrow, local
	// token boundary while preserving the same core-ESLint range.
	if loc, ok := l.singleParameterFieldArrowHeadLoc(node); ok {
		return loc
	}

	loc := utils.GetFunctionHeadLoc(l.sourceFile, node)

	member := node
	switch {
	case isClassMember(node):
	case node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionExpression:
		// The field owning a parenthesized initializer sits above the
		// ParenthesizedExpression tsgo interposes; an auto-accessor never owns
		// the report range in the first place, so it keeps the shared helper's.
		owner := ast.WalkUpParenthesizedExpressions(node.Parent)
		if owner == nil || owner.Kind != ast.KindPropertyDeclaration ||
			ast.HasSyntacticModifier(owner, ast.ModifierFlagsAccessor) {
			return loc
		}
		member = owner
	default:
		return loc
	}

	if !hasDecorator(member) {
		return loc
	}
	return core.NewTextRange(utils.TrimNodeTextRange(l.sourceFile, member).Pos(), loc.End())
}

// singleParameterFieldArrowHeadLoc returns core ESLint's report range for an
// arrow held directly by an object property or ordinary class field. Only
// ParenthesizedExpression wrappers are transparent; an auto-accessor or any
// other expression wrapper deliberately falls back to GetFunctionHeadLoc.
func (l *functionHeadLocator) singleParameterFieldArrowHeadLoc(node *ast.Node) (core.TextRange, bool) {
	if node.Kind != ast.KindArrowFunction {
		return core.TextRange{}, false
	}
	arrow := node.AsArrowFunction()
	if arrow.Parameters == nil || len(arrow.Parameters.Nodes) != 1 {
		return core.TextRange{}, false
	}

	owner := ast.WalkUpParenthesizedExpressions(node.Parent)
	if owner == nil {
		return core.TextRange{}, false
	}
	switch owner.Kind {
	case ast.KindPropertyAssignment:
	case ast.KindPropertyDeclaration:
		if ast.HasSyntacticModifier(owner, ast.ModifierFlagsAccessor) {
			return core.TextRange{}, false
		}
	default:
		return core.TextRange{}, false
	}

	parameterStart := utils.TrimNodeTextRange(l.sourceFile, arrow.Parameters.Nodes[0]).Pos()
	scanStart := utils.TrimNodeTextRange(l.sourceFile, node).Pos()
	if parent := node.Parent; parent != nil && parent.Kind == ast.KindParenthesizedExpression {
		// A parenless arrow can itself be wrapped: `field = (value => {})`.
		// ESLint sees no wrapper node, so that `(` is the token before `value`.
		scanStart = utils.TrimNodeTextRange(l.sourceFile, parent).Pos()
	}

	end := parameterStart
	var previousKind ast.Kind
	previousStart := 0
	foundPrevious := false
	if l.tokens == nil {
		l.tokens = scanner.NewScanner()
	}
	// Reuse one scanner per file: constructing one for every report would trade
	// the quadratic scan for avoidable per-diagnostic allocations.
	l.tokens.SetText(l.sourceFile.Text())
	l.tokens.SetLanguageVariant(l.sourceFile.LanguageVariant)
	l.tokens.ResetTokenState(scanStart)
	l.tokens.Scan()
	tokens := l.tokens
	for tokens.Token() != ast.KindEndOfFile && tokens.TokenStart() < parameterStart {
		if tokens.TokenEnd() <= parameterStart {
			previousKind = tokens.Token()
			previousStart = tokens.TokenStart()
			foundPrevious = true
		}
		tokens.Scan()
	}
	if foundPrevious && previousKind == ast.KindOpenParenToken {
		end = previousStart
	}

	return core.NewTextRange(utils.TrimNodeTextRange(l.sourceFile, owner).Pos(), end), true
}

type ruleOptions struct {
	max                     int
	ignoreTopLevelFunctions bool
}

func parseOptions(options []any) ruleOptions {
	opts := ruleOptions{max: defaultMax}
	if len(options) == 0 {
		return opts
	}
	opts.max = utils.ResolveLegacyMaxOption(options[0], defaultMax)
	if len(options) < 2 {
		return opts
	}
	m, _ := options[1].(map[string]interface{})
	if v, ok := m["ignoreTopLevelFunctions"].(bool); ok {
		opts.ignoreTopLevelFunctions = v
	}
	return opts
}
