package hoisted_apis_on_top

import (
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// hoistedFactory is the one hoisted API that produces a value. Every other
// hoisted API returns undefined, which is what lets the value-position
// suggestion below leave `undefined` behind.
const hoistedFactory = "hoisted"

// hoistedAPIs maps each module-mock API Rstest lifts to the top of the module
// onto the counterpart that runs where it is written, or "" when the API has
// none.
//
// The Rstest build rewrites `mock`, `mockRequire`, `unmock`, `unmockRequire`
// and `hoisted`; `doMock`, `doMockRequire`, `doUnmock` and `doUnmockRequire`
// are the same four module operations without the lift. `hoisted` has no
// non-hoisted twin: its whole purpose is to run early.
var hoistedAPIs = map[string]string{
	"mock":          "doMock",
	"mockRequire":   "doMockRequire",
	"unmock":        "doUnmock",
	"unmockRequire": "doUnmockRequire",
	hoistedFactory:  "",
}

// namespaceNames are the two spellings of the utilities object the rewrite
// recognizes.
//
// NOTE: no import or scope analysis backs this set, and that is deliberate
// rather than an omission. The rewrite matches the receiver by the name written
// at the call site: `rs.mock('./m')` is lifted in a file whose only `rs` is a
// local `const rs = {...}`, and in a file that imports `rs` from another
// package, while the same call written through a renamed binding
// (`import { rs as r }`, then `r.mock('./m')`) is not rewritten at all and
// throws when it runs. Resolving the receiver would therefore report calls the
// build leaves alone and stay silent on calls it lifts.
var namespaceNames = map[string]bool{
	"rs":     true,
	"rstest": true,
}

var hoistedApisOnTopMessage = rule.RuleMessage{
	Id:          "hoistedApisOnTop",
	Description: "Hoisted API is used in a runtime location in this file, but it is actually executed before this file is loaded.",
}

var suggestMoveHoistedApiToTopMessage = rule.RuleMessage{
	Id:          "suggestMoveHoistedApiToTop",
	Description: "Move this hoisted API to the top of the file to better reflect its behavior.",
}

var HoistedApisOnTopRule = rule.Rule{
	Name:   "rstest/hoisted-apis-on-top",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		// The insertion point is the end of the last import in the file, which
		// can sit after the call being reported. It is resolved once, on the
		// first suggestion that is actually materialized.
		insertPos := -1
		insertResolved := false
		insertionPoint := func() int {
			if !insertResolved {
				insertPos = lastImportEnd(ctx.SourceFile)
				insertResolved = true
			}
			return insertPos
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				accessor, namespace, api, ok := hoistedAPICall(node)
				if !ok || isWrittenAtTopLevel(node, api) {
					return
				}

				ctx.ReportNodeWithDeferredSuggestions(node, hoistedApisOnTopMessage, func() []rule.RuleSuggestion {
					suggestions := make([]rule.RuleSuggestion, 0, 2)

					if fixes := moveToTopFixes(ctx, node, api, insertionPoint()); fixes != nil {
						suggestions = append(suggestions, rule.RuleSuggestion{
							Message:  suggestMoveHoistedApiToTopMessage,
							FixesArr: fixes,
						})
					}

					if nonHoisted := hoistedAPIs[api]; nonHoisted != "" {
						suggestions = append(suggestions, rule.RuleSuggestion{
							Message: useNonHoistedAPIMessage(namespace, api, nonHoisted),
							FixesArr: []rule.RuleFix{
								rule.RuleFixReplace(ctx.SourceFile, accessor, nonHoisted),
							},
						})
					}

					if len(suggestions) == 0 {
						return nil
					}
					return suggestions
				})
			},
		}
	},
}

func useNonHoistedAPIMessage(namespace, api, replacement string) rule.RuleMessage {
	hoisted := namespace + "." + api
	nonHoisted := namespace + "." + replacement
	return rule.RuleMessage{
		Id:          "suggestUseNonHoistedApi",
		Description: "Replace '" + hoisted + "()' with '" + nonHoisted + "()', which is not hoisted.",
		Data: map[string]string{
			"hoisted":    hoisted,
			"nonHoisted": nonHoisted,
		},
	}
}

// hoistedAPICall reports whether node is a call the Rstest build lifts, and
// returns the API's name node, the receiver's name, and the API's name.
//
// The shapes accepted here are the ones the rewrite accepts. The receiver may
// carry parentheses and TypeScript wrappers — `(rs).mock()`, `rs!.mock()`,
// `(rs as any).mock()`, `(rs satisfies RstestUtilities).mock()` are all lifted,
// because those are erased before the module-mock rewrite runs. The API name
// must be written as a plain dotted member: `rs['mock']()` and a call reached
// through `import.meta.rstest` are left as ordinary runtime calls and throw,
// and so is anything on an optional chain, so none of them are reported here.
func hoistedAPICall(node *ast.Node) (*ast.Node, string, string, bool) {
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil ||
		call.Expression.Kind != ast.KindPropertyAccessExpression ||
		ast.IsOptionalChain(node) || ast.IsOptionalChain(call.Expression) {
		return nil, "", "", false
	}

	access := call.Expression.AsPropertyAccessExpression()
	accessor := access.Name()
	if accessor == nil || accessor.Kind != ast.KindIdentifier {
		return nil, "", "", false
	}
	api := accessor.AsIdentifier().Text
	if _, lifted := hoistedAPIs[api]; !lifted {
		return nil, "", "", false
	}

	receiver := utils.SkipAssertionsAndParens(access.Expression)
	if receiver == nil || receiver.Kind != ast.KindIdentifier {
		return nil, "", "", false
	}
	namespace := receiver.AsIdentifier().Text
	if !namespaceNames[namespace] {
		return nil, "", "", false
	}

	return accessor, namespace, api, true
}

// isWrittenAtTopLevel reports whether the call is already written where it
// runs: as a statement of the file itself.
//
// `hoisted` is also accepted as the initializer of a top-level variable, with
// or without an `await`, because that is the shape the API is designed for —
// its return value is what the rest of the file uses. The other hoisted APIs
// evaluate to undefined, so a top-level statement is the only shape that reads
// as intended.
func isWrittenAtTopLevel(call *ast.Node, api string) bool {
	outer := utils.OutermostParenthesizedExpression(call)
	parent := outer.Parent

	if api == hoistedFactory {
		if parent != nil && parent.Kind == ast.KindAwaitExpression {
			outer = utils.OutermostParenthesizedExpression(parent)
			parent = outer.Parent
		}
		if parent != nil && parent.Kind == ast.KindVariableDeclaration {
			statement := variableStatementOf(parent)
			return statement != nil &&
				statement.Parent != nil &&
				statement.Parent.Kind == ast.KindSourceFile
		}
	}

	return parent != nil &&
		parent.Kind == ast.KindExpressionStatement &&
		parent.Parent != nil &&
		parent.Parent.Kind == ast.KindSourceFile
}

// moveToTopFixes builds the edits that lift the call to where it actually runs,
// or nil when the call cannot be lifted without changing what the surrounding
// expression evaluates to.
//
// Three shapes are handled, and the third is why the first two are separate:
//
//   - A statement of its own becomes an empty statement. Deleting the statement
//     outright would leave `if (condition)` without a body when it was written
//     without braces, so `;` takes its place.
//   - `hoisted` bound by a single-declarator variable statement moves whole.
//     The Rstest build lifts only the call, so the binding is assigned when its
//     block runs while the factory has already executed; moving the declaration
//     with the call is what makes the file read the way it runs.
//   - Any other position of a void API leaves `undefined`, which is exactly
//     what the call evaluated to.
//
// `hoisted` in any other position produces no edit: its value is used, and no
// text can stand in for it.
func moveToTopFixes(ctx rule.RuleContext, call *ast.Node, api string, insertPos int) []rule.RuleFix {
	outer := utils.OutermostParenthesizedExpression(call)

	if parent := outer.Parent; parent != nil && parent.Kind == ast.KindExpressionStatement {
		return append(
			insertAtTop(utils.TrimmedNodeText(ctx.SourceFile, call)+";", insertPos),
			rule.RuleFixReplace(ctx.SourceFile, parent, ";"),
		)
	}

	if api == hoistedFactory {
		statement := movableVariableStatement(outer)
		if statement == nil {
			return nil
		}
		text := utils.TrimmedNodeText(ctx.SourceFile, statement)
		if !strings.HasSuffix(text, ";") {
			text += ";"
		}
		return append(
			insertAtTop(text, insertPos),
			rule.RuleFixRemove(ctx.SourceFile, statement),
		)
	}

	return append(
		insertAtTop(utils.TrimmedNodeText(ctx.SourceFile, call)+";", insertPos),
		rule.RuleFixReplace(ctx.SourceFile, call, "undefined"),
	)
}

// insertAtTop places one complete statement after the file's last import, or at
// the very start of a file that has none.
func insertAtTop(statement string, insertPos int) []rule.RuleFix {
	if insertPos < 0 {
		return []rule.RuleFix{
			rule.RuleFixReplaceRange(core.NewTextRange(0, 0), statement+"\n"),
		}
	}
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(insertPos, insertPos), "\n"+statement),
	}
}

// lastImportEnd returns the end of the file's last import declaration, or -1
// when the file has none. An import written below the reported call still
// counts: the lifted call belongs above every import, so the last one in the
// file is the boundary.
func lastImportEnd(sourceFile *ast.SourceFile) int {
	end := -1
	if sourceFile == nil || sourceFile.Statements == nil {
		return end
	}
	for _, statement := range sourceFile.Statements.Nodes {
		switch statement.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			end = statement.End()
		}
	}
	return end
}

// movableVariableStatement returns the variable statement that can move along
// with initializer, or nil when the binding cannot travel with it: a second
// declarator in the same statement would move too, and a declaration list that
// belongs to a `for` header has no statement of its own.
func movableVariableStatement(initializer *ast.Node) *ast.Node {
	node := initializer
	if node.Parent != nil && node.Parent.Kind == ast.KindAwaitExpression {
		node = utils.OutermostParenthesizedExpression(node.Parent)
	}

	declaration := node.Parent
	if declaration == nil ||
		declaration.Kind != ast.KindVariableDeclaration ||
		declaration.AsVariableDeclaration().Initializer != node {
		return nil
	}

	list := declaration.Parent.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil || len(list.Declarations.Nodes) != 1 {
		return nil
	}
	return variableStatementOf(declaration)
}

// variableStatementOf returns the statement a variable declaration belongs to,
// or nil when the declaration list is a `for` header rather than a statement.
func variableStatementOf(declaration *ast.Node) *ast.Node {
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return nil
	}
	statement := list.Parent
	if statement == nil || statement.Kind != ast.KindVariableStatement {
		return nil
	}
	return statement
}
