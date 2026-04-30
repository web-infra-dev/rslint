package strict

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/strict
//
// NOTE: Unlike ESLint, rslint does not expose languageOptions.parserOptions
// (ecmaFeatures.impliedStrict / ecmaFeatures.globalReturn) or
// sourceType === "commonjs". rslint detects ES modules via
// ast.IsExternalModule (presence of import/export) and treats every other
// file as a script. Consequences:
//   - "safe" resolves to "function" on script files and "module" on ES modules
//     (ESLint's commonjs-aware "global" fallback has no rslint analogue).
//   - "implied" and "globalReturn" code paths are unreachable.
// See strict.md "Differences from ESLint" for the user-visible contract.

type strictMode int

const (
	modeNever strictMode = iota
	modeGlobal
	modeFunction
	modeModule
)

func (m strictMode) messageId() string {
	switch m {
	case modeNever:
		return "never"
	case modeGlobal:
		return "global"
	case modeFunction:
		return "function"
	case modeModule:
		return "module"
	}
	return ""
}

// shouldFix mirrors ESLint's shouldFix — true when the reported diagnostic
// corresponds to a redundant directive that should simply be removed.
func shouldFix(errorType string) bool {
	switch errorType {
	case "multiple", "unnecessary", "module", "unnecessaryInClasses":
		return true
	}
	return false
}

// messageDescriptionFor returns the static description text for a given
// messageId (matches ESLint's meta.messages entries verbatim so Go tests can
// assert on Message if desired).
func messageDescriptionFor(msgId string) string {
	switch msgId {
	case "function":
		return "Use the function form of 'use strict'."
	case "global":
		return "Use the global form of 'use strict'."
	case "multiple":
		return "Multiple 'use strict' directives."
	case "never":
		return "Strict mode is not permitted."
	case "unnecessary":
		return "Unnecessary 'use strict' directive."
	case "module":
		return "'use strict' is unnecessary inside of modules."
	case "unnecessaryInClasses":
		return "'use strict' is unnecessary inside of classes."
	case "nonSimpleParameterList":
		return "'use strict' directive inside a function with non-simple parameter list throws a syntax error since ES2016."
	}
	return ""
}

// getUseStrictDirectives collects all "use strict" directives at the start of
// a statement list, stopping at the first non-matching statement. This
// mirrors ESLint's helper of the same name exactly — any leading non-matching
// ExpressionStatement terminates the scan, even if it is itself a string
// literal prologue directive (e.g. "use asm").
func getUseStrictDirectives(statements []*ast.Node) []*ast.Node {
	var directives []*ast.Node
	for _, stmt := range statements {
		if !ast.IsPrologueDirective(stmt) {
			break
		}
		expr := stmt.Expression()
		if expr == nil || expr.Text() != "use strict" {
			break
		}
		directives = append(directives, stmt)
	}
	return directives
}

// isSimpleParameter is true when the parameter is a plain identifier with no
// default value, no rest, no destructuring, and no type annotation-only
// binding patterns. Matches ESLint's `node.type === "Identifier"` check on
// ESTree parameters.
func isSimpleParameter(param *ast.Node) bool {
	if param == nil || param.Kind != ast.KindParameter {
		return false
	}
	p := param.AsParameterDeclaration()
	if p == nil {
		return false
	}
	if p.DotDotDotToken != nil || p.Initializer != nil {
		return false
	}
	name := p.Name()
	return name != nil && name.Kind == ast.KindIdentifier
}

func isSimpleParameterList(params []*ast.Node) bool {
	for _, p := range params {
		if !isSimpleParameter(p) {
			return false
		}
	}
	return true
}

// getFunctionBodyStatements returns the statements of a function's block
// body, or nil for arrow functions with expression bodies (no directive
// prologue is possible).
func getFunctionBodyStatements(node *ast.Node) []*ast.Node {
	body := node.Body()
	if body == nil || body.Kind != ast.KindBlock {
		return nil
	}
	block := body.AsBlock()
	if block == nil || block.Statements == nil {
		return nil
	}
	return block.Statements.Nodes
}


func reportDirectives(ctx rule.RuleContext, directives []*ast.Node, msgId string, fix bool) {
	desc := messageDescriptionFor(msgId)
	for _, d := range directives {
		msg := rule.RuleMessage{Id: msgId, Description: desc}
		if fix {
			ctx.ReportNodeWithFixes(d, msg, rule.RuleFixRemove(ctx.SourceFile, d))
		} else {
			ctx.ReportNode(d, msg)
		}
	}
}

// handleProgram processes the source-file level prologue once at rule setup.
// The linter does not fire a KindSourceFile listener, so this runs eagerly.
func handleProgram(ctx rule.RuleContext, m strictMode) {
	if ctx.SourceFile.Statements == nil {
		return
	}
	body := ctx.SourceFile.Statements.Nodes
	directives := getUseStrictDirectives(body)

	if m == modeGlobal {
		if len(body) > 0 && len(directives) == 0 {
			// ESLint v9 reports the range from the first body statement to
			// the last body statement.
			startRange := utils.TrimNodeTextRange(ctx.SourceFile, body[0])
			endRange := utils.TrimNodeTextRange(ctx.SourceFile, body[len(body)-1])
			rng := core.NewTextRange(startRange.Pos(), endRange.End())
			ctx.ReportRange(rng, rule.RuleMessage{
				Id:          "global",
				Description: messageDescriptionFor("global"),
			})
		}
		if len(directives) > 1 {
			reportDirectives(ctx, directives[1:], "multiple", true)
		}
		return
	}

	// never / function / module: report every directive with the mode's
	// messageId. "function" at global scope is unfixable (shouldFix=false).
	msgId := m.messageId()
	reportDirectives(ctx, directives, msgId, shouldFix(msgId))
}

// buildSimpleListeners is the listener set for every mode except "function".
// Each function body's directive prologue is reported as a whole; no scope
// tracking is needed because the mode is fixed across the file.
func buildSimpleListeners(ctx rule.RuleContext, m strictMode) rule.RuleListeners {
	handle := func(node *ast.Node) {
		directives := getUseStrictDirectives(getFunctionBodyStatements(node))
		if len(directives) == 0 {
			return
		}
		if isSimpleParameterList(node.Parameters()) {
			msgId := m.messageId()
			reportDirectives(ctx, directives, msgId, shouldFix(msgId))
			return
		}
		ctx.ReportNode(directives[0], rule.RuleMessage{
			Id:          "nonSimpleParameterList",
			Description: messageDescriptionFor("nonSimpleParameterList"),
		})
		if len(directives) > 1 {
			reportDirectives(ctx, directives[1:], "multiple", true)
		}
	}
	return rule.RuleListeners{
		ast.KindFunctionDeclaration: handle,
		ast.KindFunctionExpression:  handle,
		ast.KindArrowFunction:       handle,
		ast.KindMethodDeclaration:   handle,
		ast.KindConstructor:         handle,
		ast.KindGetAccessor:         handle,
		ast.KindSetAccessor:         handle,
	}
}

// isClassBodyMember returns true when the given node sits in a class body
// member slot (method / accessor / constructor / field / static block). Used
// to distinguish class-body descendants from decorators, heritage clauses,
// type parameters, and other ClassDeclaration children that are visited
// outside the `{ … }` body in ESLint's ClassBody scope.
func isClassBodyMember(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindPropertyDeclaration,
		ast.KindMethodDeclaration,
		ast.KindConstructor,
		ast.KindGetAccessor,
		ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration,
		ast.KindSemicolonClassElement:
		return true
	}
	return false
}

// isInClassBody matches ESLint's `classScopes.length > 0` check. It is true
// when `node` is a descendant of some enclosing class body member. Functions
// that appear in a class's decorators, heritage clause, or modifiers are
// NOT considered in a class body (ESLint only pushes classScopes on
// ClassBody, not ClassDeclaration).
func isInClassBody(node *ast.Node) bool {
	child, parent := node, node.Parent
	for parent != nil {
		if parent.Kind == ast.KindClassDeclaration || parent.Kind == ast.KindClassExpression {
			if isClassBodyMember(child) {
				return true
			}
			// Reached this class via a non-body slot (decorator / heritage /
			// etc.) — keep walking; an outer class may still contain us.
		}
		child, parent = parent, parent.Parent
	}
	return false
}

// buildFunctionModeListeners implements ESLint's "function" mode bookkeeping.
// `scopes` tracks whether each enclosing function body is in strict mode
// (self-declared or inherited from a strict parent). Class-body membership is
// resolved per-visit via ancestor walking so that functions appearing in a
// class's heritage / decorators are correctly treated as outside the class.
func buildFunctionModeListeners(ctx rule.RuleContext) rule.RuleListeners {
	var scopes []bool

	enterFunction := func(node *ast.Node) {
		// Skip ambient / bodyless declarations (TypeScript `declare function`,
		// abstract methods, overload signatures) — no runtime body where a
		// directive prologue applies.
		if node.Body() == nil {
			return
		}
		directives := getUseStrictDirectives(getFunctionBodyStatements(node))
		isInClass := isInClassBody(node)
		isParentGlobal := len(scopes) == 0 && !isInClass
		isParentStrict := len(scopes) > 0 && scopes[len(scopes)-1]
		isStrict := len(directives) > 0
		simpleParams := isSimpleParameterList(node.Parameters())

		if isStrict {
			first := directives[0]
			switch {
			case !simpleParams:
				ctx.ReportNode(first, rule.RuleMessage{
					Id:          "nonSimpleParameterList",
					Description: messageDescriptionFor("nonSimpleParameterList"),
				})
			case isParentStrict:
				ctx.ReportNodeWithFixes(first, rule.RuleMessage{
					Id:          "unnecessary",
					Description: messageDescriptionFor("unnecessary"),
				}, rule.RuleFixRemove(ctx.SourceFile, first))
			case isInClass:
				ctx.ReportNodeWithFixes(first, rule.RuleMessage{
					Id:          "unnecessaryInClasses",
					Description: messageDescriptionFor("unnecessaryInClasses"),
				}, rule.RuleFixRemove(ctx.SourceFile, first))
			}
			if len(directives) > 1 {
				reportDirectives(ctx, directives[1:], "multiple", true)
			}
		} else if isParentGlobal {
			if simpleParams {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "function",
					Description: messageDescriptionFor("function"),
				})
			} else {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "wrap",
					Description: fmt.Sprintf("Wrap %s in a function with 'use strict' directive.", utils.GetFunctionNameWithKind(node)),
				})
			}
		}

		scopes = append(scopes, isParentStrict || isStrict)
	}

	exitFunction := func(node *ast.Node) {
		if node.Body() == nil {
			return
		}
		if len(scopes) > 0 {
			scopes = scopes[:len(scopes)-1]
		}
	}

	return rule.RuleListeners{
		ast.KindFunctionDeclaration:                      enterFunction,
		ast.KindFunctionExpression:                       enterFunction,
		ast.KindArrowFunction:                            enterFunction,
		ast.KindMethodDeclaration:                        enterFunction,
		ast.KindConstructor:                              enterFunction,
		ast.KindGetAccessor:                              enterFunction,
		ast.KindSetAccessor:                              enterFunction,
		rule.ListenerOnExit(ast.KindFunctionDeclaration): exitFunction,
		rule.ListenerOnExit(ast.KindFunctionExpression):  exitFunction,
		rule.ListenerOnExit(ast.KindArrowFunction):       exitFunction,
		rule.ListenerOnExit(ast.KindMethodDeclaration):   exitFunction,
		rule.ListenerOnExit(ast.KindConstructor):         exitFunction,
		rule.ListenerOnExit(ast.KindGetAccessor):         exitFunction,
		rule.ListenerOnExit(ast.KindSetAccessor):         exitFunction,
	}
}

var StrictRule = rule.Rule{
	Name: "strict",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		optStr := utils.GetOptionsString(options)
		if optStr == "" {
			optStr = "safe"
		}

		var m strictMode
		switch optStr {
		case "never":
			m = modeNever
		case "global":
			m = modeGlobal
		case "function":
			m = modeFunction
		case "safe":
			// rslint cannot detect commonjs / globalReturn; "safe" on a
			// script file therefore always maps to "function" (matching
			// ESLint's non-commonjs, non-globalReturn behavior).
			m = modeFunction
		default:
			m = modeFunction
		}

		if ast.IsExternalModule(ctx.SourceFile) {
			m = modeModule
		}

		handleProgram(ctx, m)

		if m == modeFunction {
			return buildFunctionModeListeners(ctx)
		}
		return buildSimpleListeners(ctx, m)
	},
}
