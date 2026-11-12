// Ported from eslint-plugin-unicorn v79.0.0 (MIT).
// cspell:ignore scrollend beforeunload messageerror pagehide pagereveal pageshow pageswap locationbar personalbar
package prefer_global_this

import (
	"cmp"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	"github.com/web-infra-dev/rslint/internal/utils/scopeanalysis"
)

var globalNames = map[string]struct{}{"window": {}, "self": {}, "global": {}}

var PreferGlobalThisRule = rule.Rule{
	Name:   "unicorn/prefer-global-this",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		if !ctx.SourceFile.HasIdentifier("window") && !ctx.SourceFile.HasIdentifier("self") && !ctx.SourceFile.HasIdentifier("global") {
			return nil
		}
		return rule.RuleListeners{rule.ListenerOnExit(ast.KindEndOfFile): func(*ast.Node) {
			manager := scopeanalysis.References(ctx, globalNames)
			globalProgram := !ctx.Refs.HasNonGlobalProgramScope()
			report := func(identifier *ast.Node, write bool) {
				if isTypeofOperand(identifier) || (identifier.Text() == "window" && isWindowException(identifier)) {
					return
				}
				message := rule.RuleMessage{
					Id:          "prefer-global-this/error",
					Description: "Prefer `globalThis` over `" + identifier.Text() + "`.",
					Data:        map[string]string{"replacement": "globalThis", "value": identifier.Text()},
				}
				if write {
					ctx.ReportNode(identifier, message)
					return
				}
				ctx.ReportNodeWithDeferredFixes(identifier, message, func() []rule.RuleFix {
					// Upstream replaces even when the new name is shadowed, and
					// Espree references only the opening name of a paired JSX tag.
					// Keep the diagnostic without changing meaning or breaking JSX.
					if ctx.Refs.IsNameDefinedInFileWithMeaning(identifier, "globalThis", ast.SymbolFlagsValue|ast.SymbolFlagsNamespace|ast.SymbolFlagsAlias) || isPairedJavaScriptJSXName(identifier) {
						return nil
					}
					replacement := "globalThis"
					parent := identifier.Parent
					if parent.Kind == ast.KindShorthandPropertyAssignment && parent.Name() == identifier {
						replacement = identifier.Text() + ": " + replacement
					} else if parent.Kind == ast.KindExportSpecifier && parent.PropertyName() == nil {
						replacement += " as " + identifier.Text()
					}
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, identifier, replacement)}
				})
			}

			// Reuse source-ordered references without copying the shared graph.
			// Script-level initialization writes need their own, sortable slice.
			references := manager.References
			if globalProgram {
				references = slices.Clone(references)
				for _, binding := range manager.Global.Vars {
					if _, tracked := globalNames[binding.Name]; tracked && binding.Kind == scope.DefVariable && utils.IsVariableWriteReference(binding.ID) {
						references = append(references, &scope.Reference{Identifier: binding.ID})
					}
				}
			}
			compare := func(a, b *scope.Reference) int { return cmp.Compare(a.Identifier.Pos(), b.Identifier.Pos()) }
			if !slices.IsSortedFunc(references, compare) {
				if !globalProgram {
					references = slices.Clone(references)
				}
				slices.SortFunc(references, compare)
			}
			for _, ref := range references {
				identifier := ref.Identifier
				if _, tracked := globalNames[identifier.Text()]; !tracked {
					continue // A cached complete graph can contain other names.
				}
				binding := ref.Resolved()
				if binding != nil && (!globalProgram || binding.Scope != manager.Global) {
					continue
				}
				report(identifier, utils.IsVariableWriteReference(identifier) || manager.PatternTargets[identifier] != nil)
			}
		}}
	},
}

func isPairedJavaScriptJSXName(identifier *ast.Node) bool {
	if !ast.IsInJSFile(identifier) || !utils.IsInJsxTagName(identifier) {
		return false
	}
	name := ast.FindAncestor(identifier, ast.IsJsxTagName)
	return name.Parent.Kind == ast.KindJsxOpeningElement
}

func isTypeofOperand(identifier *ast.Node) bool {
	// Parentheses and erased TS wrappers do not change a runtime typeof check.
	expression := identifier
	for parent := expression.Parent; parent != nil &&
		ast.IsOuterExpression(parent, ast.OEKParentheses|ast.OEKAssertions|ast.OEKExpressionsWithTypeArguments) &&
		parent.Expression() == expression; parent = expression.Parent {
		expression = parent
	}
	// Only the queried name is exempt, not its type arguments.
	return expression.Parent != nil && expression.Parent.Kind == ast.KindTypeOfExpression || ast.IsPartOfTypeQuery(identifier)
}

func isWindowException(identifier *ast.Node) bool {
	parent := utils.ESTreeParent(identifier)
	object, property := utils.MemberExpressionParts(parent)
	if utils.ESTreeRuntimeExpression(object) == identifier && !utils.IsInJsxTagName(parent) {
		if parent.Kind == ast.KindElementAccessExpression {
			return true
		}
		if property.Kind == ast.KindIdentifier && isWindowAPI(property.Text()) {
			if isEventMethod(property.Text()) {
				call := utils.ESTreeParent(parent)
				if call != nil && call.Kind == ast.KindCallExpression && utils.ESTreeCallCallee(call.Expression()) == parent {
					arguments := call.Arguments()
					return len(arguments) > 0 && isWindowEvent(utils.GetStaticStringValue(utils.ESTreeRuntimeExpression(arguments[0])))
				}
			}
			return true
		}
	}
	if parent != nil && parent.Kind == ast.KindBinaryExpression {
		binary := parent.AsBinaryExpression()
		return binary.OperatorToken.Kind == ast.KindInKeyword && utils.ESTreeRuntimeExpression(binary.Right) == identifier &&
			isWindowAPI(utils.GetStaticStringValue(utils.ESTreeRuntimeExpression(binary.Left)))
	}
	return false
}

func isEventMethod(name string) bool {
	return name == "addEventListener" || name == "removeEventListener" || name == "dispatchEvent"
}

func isWindowEvent(name string) bool {
	switch name {
	case "resize", "blur", "focus", "load", "scroll", "scrollend", "wheel", "beforeunload",
		"message", "messageerror", "pagehide", "pagereveal", "pageshow", "pageswap", "unload":
		return true
	}
	return false
}

func isWindowAPI(name string) bool {
	if event, ok := strings.CutPrefix(name, "on"); ok && isWindowEvent(event) {
		return true
	}
	switch name {
	case "name", "locationbar", "menubar", "personalbar", "scrollbars", "statusbar", "toolbar", "status",
		"close", "closed", "stop", "focus", "blur", "frames", "length", "top", "opener", "parent", "frameElement",
		"open", "originAgentCluster", "postMessage", "navigation", "addEventListener", "removeEventListener", "dispatchEvent",
		"event", "screen", "visualViewport", "moveTo", "moveBy", "resizeTo", "resizeBy", "innerWidth", "innerHeight",
		"outerWidth", "outerHeight", "scrollX", "pageXOffset", "scrollY", "pageYOffset", "scroll", "scrollTo", "scrollBy",
		"screenX", "screenLeft", "screenY", "screenTop", "screenWidth", "screenHeight", "devicePixelRatio":
		return true
	}
	return false
}
