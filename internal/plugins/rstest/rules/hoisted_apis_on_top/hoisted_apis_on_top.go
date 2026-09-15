package hoisted_apis_on_top

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

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
	"hoisted":       "",
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
		// Both of these are needed only to materialize a move suggestion, and
		// the insertion point in particular depends on imports that can sit
		// after the call being reported. They are resolved once, on the first
		// suggestion that is actually built.
		insertPos, insertResolved := 0, false
		insertionPoint := func() int {
			if !insertResolved {
				insertPos = topLevelInsertionPoint(ctx.SourceFile)
				insertResolved = true
			}
			return insertPos
		}
		var topLevelNames map[string]bool
		topLevelBindings := func() map[string]bool {
			if topLevelNames == nil {
				topLevelNames = topLevelBindingNames(ctx.SourceFile)
			}
			return topLevelNames
		}

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				accessor, namespace, api, ok := hoistedAPICall(node)
				if !ok {
					return
				}
				position, lifted := liftedCallPosition(node)
				if !lifted || isWrittenAtTopLevel(position) {
					return
				}

				ctx.ReportNodeWithDeferredSuggestions(node, hoistedApisOnTopMessage, func() []rule.RuleSuggestion {
					suggestions := make([]rule.RuleSuggestion, 0, 2)

					if fixes := moveToTopFixes(ctx, position, insertionPoint(), topLevelBindings); fixes != nil {
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

// hoistedAPICall reports whether node names a module-mock API the Rstest build
// lifts, and returns the API's name node, the receiver's name, and the API's
// name.
func hoistedAPICall(node *ast.Node) (*ast.Node, string, string, bool) {
	utility := rstestUtils.ParseRstestPluginManagedCall(node)
	if utility == nil {
		return nil, "", "", false
	}
	if _, lifted := hoistedAPIs[utility.Member]; !lifted {
		return nil, "", "", false
	}
	return utility.MemberNode, utility.Namespace, utility.Member, true
}

// liftedPosition is where a lifted call is written, as the rewrite sees it:
// exactly one of statement and declaration is set.
type liftedPosition struct {
	statement   *ast.Node
	declaration *ast.Node
}

// liftedCallPosition reports whether the Rstest build actually lifts this call,
// and where it is written.
//
// Only two positions are rewritten, at any nesting depth: the whole expression
// of an expression statement, and the whole initializer of a variable
// declaration, the latter optionally behind an `await`. Anywhere else — an
// argument, an object property, an arrow's expression body, an operand of `&&`,
// the right-hand side of an assignment, an `await` in statement position — the
// call is left alone and throws where it is written. Reporting those would
// describe a lift that does not happen, and the move suggestion would rewrite
// working-or-not code on a false premise.
//
// Parentheses and TypeScript wrappers around the call are transparent here for
// the same reason they are around the callee: `rs.mock('./m') as unknown;` and
// `const v = (await rs.hoisted(f)) as number` are both rewritten.
func liftedCallPosition(call *ast.Node) (liftedPosition, bool) {
	outer := skipTransparentAncestors(call)
	parent := outer.Parent
	if parent == nil {
		return liftedPosition{}, false
	}

	if parent.Kind == ast.KindExpressionStatement {
		if parent.AsExpressionStatement().Expression != outer {
			return liftedPosition{}, false
		}
		return liftedPosition{statement: parent}, true
	}

	// `await rs.mock('./m');` in statement position is NOT rewritten, so the
	// `await` is only followed on the way to a variable declaration.
	if parent.Kind == ast.KindAwaitExpression {
		if parent.AsAwaitExpression().Expression != outer {
			return liftedPosition{}, false
		}
		outer = skipTransparentAncestors(parent)
		parent = outer.Parent
		if parent == nil {
			return liftedPosition{}, false
		}
	}

	if parent.Kind == ast.KindVariableDeclaration &&
		parent.AsVariableDeclaration().Initializer == outer {
		return liftedPosition{declaration: parent}, true
	}
	return liftedPosition{}, false
}

// skipTransparentAncestors walks out of the parentheses and TypeScript wrappers
// that enclose node, stopping at the first ancestor that is neither.
func skipTransparentAncestors(node *ast.Node) *ast.Node {
	for node != nil && node.Parent != nil {
		parent := node.Parent
		var inner *ast.Node
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			inner = parent.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			inner = parent.AsAsExpression().Expression
		case ast.KindSatisfiesExpression:
			inner = parent.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			inner = parent.AsNonNullExpression().Expression
		case ast.KindTypeAssertionExpression:
			inner = parent.AsTypeAssertion().Expression
		default:
			return node
		}
		if inner != node {
			return node
		}
		node = parent
	}
	return node
}

// isWrittenAtTopLevel reports whether the lifted call is already written where
// it runs: in a statement of the file itself.
func isWrittenAtTopLevel(position liftedPosition) bool {
	statement := position.statement
	if statement == nil {
		statement = variableStatementOf(position.declaration)
	}
	return statement != nil &&
		statement.Parent != nil &&
		statement.Parent.Kind == ast.KindSourceFile
}

// moveToTopFixes builds the edits that lift the statement to where it actually
// runs, or nil when it cannot travel.
//
// An expression statement becomes an empty statement rather than being deleted,
// so an `if` written without braces keeps a body. A declaration moves whole,
// because the build lifts only the call and leaves the binding behind: the
// binding would otherwise be assigned when its block runs while the factory had
// already executed. It is withheld when the declaration cannot travel alone — a
// second declarator in the same statement, or a `for` header, which is not a
// statement of its own — and when a name it binds already exists at the top
// level, where re-declaring it would not parse.
func moveToTopFixes(
	ctx rule.RuleContext,
	position liftedPosition,
	insertPos int,
	topLevelBindings func() map[string]bool,
) []rule.RuleFix {
	if position.statement != nil {
		text := terminated(utils.TrimmedNodeText(ctx.SourceFile, position.statement))
		return append(
			insertAtTop(ctx.SourceFile, text, insertPos),
			rule.RuleFixReplace(ctx.SourceFile, position.statement, ";"),
		)
	}

	statement := movableVariableStatement(position.declaration, topLevelBindings)
	if statement == nil {
		return nil
	}
	text := terminated(utils.TrimmedNodeText(ctx.SourceFile, statement))
	return append(
		insertAtTop(ctx.SourceFile, text, insertPos),
		rule.RuleFixRemove(ctx.SourceFile, statement),
	)
}

func terminated(statement string) string {
	if strings.HasSuffix(statement, ";") {
		return statement
	}
	return statement + ";"
}

// insertAtTop places one complete statement on a line of its own at insertPos,
// which is already past anything that has to stay first.
func insertAtTop(sourceFile *ast.SourceFile, statement string, insertPos int) []rule.RuleFix {
	text := "\n" + statement
	if insertPos == 0 || sourceFile.Text()[insertPos-1] == '\n' {
		text = statement + "\n"
	}
	return []rule.RuleFix{
		rule.RuleFixReplaceRange(core.NewTextRange(insertPos, insertPos), text),
	}
}

// topLevelInsertionPoint returns the offset a lifted statement can be written
// at: after the shebang, after the directive prologue, and after the file's last
// import declaration, whichever comes last. Zero means the very start of the
// file.
//
// A shebang has to stay on line one, and a directive only counts as one while it
// is still part of the prologue, so writing before either would change what the
// file means. An import written below the reported call still moves the point
// down: the lifted statement belongs above every import, so the last one in the
// file is the boundary.
func topLevelInsertionPoint(sourceFile *ast.SourceFile) int {
	if sourceFile == nil {
		return 0
	}
	pos := len(scanner.GetShebang(sourceFile.Text()))
	if sourceFile.Statements == nil {
		return pos
	}

	inPrologue := true
	for _, statement := range sourceFile.Statements.Nodes {
		switch statement.Kind {
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			pos = statement.End()
			inPrologue = false
		case ast.KindExpressionStatement:
			if inPrologue && isDirective(statement) {
				pos = statement.End()
				continue
			}
			inPrologue = false
		default:
			inPrologue = false
		}
	}
	return pos
}

// isDirective reports whether a statement is a bare string expression, the shape
// a directive prologue entry takes. Parentheses around the string disqualify it,
// exactly as they do at runtime.
func isDirective(statement *ast.Node) bool {
	expression := statement.AsExpressionStatement().Expression
	return expression != nil && expression.Kind == ast.KindStringLiteral
}

// topLevelBindingNames collects every name the top level of the file already
// declares, in any declaration space, so a moved declaration can be checked
// against it.
func topLevelBindingNames(sourceFile *ast.SourceFile) map[string]bool {
	names := map[string]bool{}
	if sourceFile == nil || sourceFile.Statements == nil {
		return names
	}

	add := func(node *ast.Node) {
		utils.CollectBindingNames(node, func(_ *ast.Node, name string) {
			names[name] = true
		})
	}

	for _, statement := range sourceFile.Statements.Nodes {
		switch statement.Kind {
		case ast.KindVariableStatement:
			list := statement.AsVariableStatement().DeclarationList
			if list == nil || list.AsVariableDeclarationList().Declarations == nil {
				continue
			}
			for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
				add(declaration.Name())
			}
		case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
			addImportedNames(statement, names)
		default:
			if name := statement.Name(); name != nil {
				add(name)
			}
		}
	}
	return names
}

func addImportedNames(statement *ast.Node, names map[string]bool) {
	clause := statement.AsImportDeclaration().ImportClause
	if clause == nil {
		return
	}
	if name := clause.Name(); name != nil && name.Kind == ast.KindIdentifier {
		names[name.AsIdentifier().Text] = true
	}
	bindings := clause.AsImportClause().NamedBindings
	if bindings == nil {
		return
	}
	switch bindings.Kind {
	case ast.KindNamespaceImport:
		if name := bindings.Name(); name != nil && name.Kind == ast.KindIdentifier {
			names[name.AsIdentifier().Text] = true
		}
	case ast.KindNamedImports:
		elements := bindings.AsNamedImports().Elements
		if elements == nil {
			return
		}
		for _, element := range elements.Nodes {
			if name := element.Name(); name != nil && name.Kind == ast.KindIdentifier {
				names[name.AsIdentifier().Text] = true
			}
		}
	}
}

// movableVariableStatement returns the variable statement that can move along
// with declaration, or nil when it cannot travel: a second declarator in the
// same statement would move too, a declaration list that belongs to a `for`
// header has no statement of its own, and a name the top level already declares
// would collide.
func movableVariableStatement(declaration *ast.Node, topLevelBindings func() map[string]bool) *ast.Node {
	list := declaration.Parent
	if list == nil || list.Kind != ast.KindVariableDeclarationList {
		return nil
	}
	declarations := list.AsVariableDeclarationList().Declarations
	if declarations == nil || len(declarations.Nodes) != 1 {
		return nil
	}
	statement := variableStatementOf(declaration)
	if statement == nil {
		return nil
	}

	taken := topLevelBindings()
	collides := false
	utils.CollectBindingNames(declaration.Name(), func(_ *ast.Node, name string) {
		if taken[name] {
			collides = true
		}
	})
	if collides {
		return nil
	}
	return statement
}

// variableStatementOf returns the statement a variable declaration belongs to,
// or nil when the declaration list is a `for` header rather than a statement.
func variableStatementOf(declaration *ast.Node) *ast.Node {
	if declaration == nil {
		return nil
	}
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
