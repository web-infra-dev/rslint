package prefer_expect_assertions

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// unknownSuite stands for a suite the resolver cannot identify, such as the
// place a function stored in an object is eventually called from. A hook
// registered from an unknown place may cover any test, and a test registered
// from one may be covered by any hook, so both answers lean towards not
// reporting.
var unknownSuite = &ast.Node{}

type useKind uint8

const (
	// useCall runs the function in the scope around the call: a direct call,
	// an immediately invoked function, or a callback handed to an ordinary call.
	useCall useKind = iota
	// useDescribe hands the function to a describe call as its body.
	useDescribe
	// useUnknown lets the function escape, so where it runs is not known.
	useUnknown
)

type functionUse struct {
	kind useKind
	call *ast.Node
}

// suiteResolver maps code to the describe blocks it registers hooks and tests
// in. A suite is identified by the function that is its body; nil is the file.
// Hooks and tests register into the suite that is running when they are
// called, so code inside a helper belongs to the suites that call the helper,
// not to the scope the helper is written in.
//
// The two kinds of registration resolve uncertainty in opposite directions,
// because a wrong answer must only ever cost a missed report:
//
//   - For hooks, a function nothing in this file calls registers nothing
//     here. Exporting it only lets other files register hooks in their own
//     suites. Such functions contribute no suites.
//   - For tests, the same functions may still be registered by a caller the
//     resolver cannot see, possibly inside a covered suite, so they resolve to
//     unknownSuite.
type suiteResolver struct {
	ctx        rule.RuleContext
	isDescribe func(call *ast.Node) bool
	forHooks   bool
	suites     map[*ast.Node][]*ast.Node
	parents    map[*ast.Node][]*ast.Node
	inProgress map[*ast.Node]bool
}

func newSuiteResolver(ctx rule.RuleContext, isDescribe func(call *ast.Node) bool, forHooks bool) *suiteResolver {
	return &suiteResolver{
		ctx:        ctx,
		isDescribe: isDescribe,
		forHooks:   forHooks,
		suites:     map[*ast.Node][]*ast.Node{},
		parents:    map[*ast.Node][]*ast.Node{},
		inProgress: map[*ast.Node]bool{},
	}
}

// suitesOf returns the suites that are running while code in scope executes.
// scope is a function, or nil for the file.
func (r *suiteResolver) suitesOf(scope *ast.Node) []*ast.Node {
	if scope == nil {
		return []*ast.Node{nil}
	}
	if suites, ok := r.suites[scope]; ok {
		return suites
	}
	if r.inProgress[scope] {
		// A recursive helper adds no suite beyond those its other callers reach.
		return nil
	}
	r.inProgress[scope] = true
	var suites []*ast.Node
	uses := r.functionUses(scope)
	if len(uses) == 0 && !r.forHooks {
		suites = []*ast.Node{unknownSuite}
	}
	for _, use := range uses {
		switch use.kind {
		case useDescribe:
			suites = appendUnique(suites, scope)
		case useCall:
			for _, suite := range r.suitesOf(enclosingFunction(use.call)) {
				suites = appendUnique(suites, suite)
			}
		default:
			suites = appendUnique(suites, unknownSuite)
		}
	}
	delete(r.inProgress, scope)
	r.suites[scope] = suites
	return suites
}

// parentSuites returns every suite a suite body is registered in.
func (r *suiteResolver) parentSuites(suite *ast.Node) []*ast.Node {
	if parents, ok := r.parents[suite]; ok {
		return parents
	}
	r.parents[suite] = nil
	var parents []*ast.Node
	for _, use := range r.functionUses(suite) {
		if use.kind != useDescribe {
			continue
		}
		for _, parent := range r.suitesOf(enclosingFunction(use.call)) {
			parents = appendUnique(parents, parent)
		}
	}
	r.parents[suite] = parents
	return parents
}

// functionUses lists where a function is run from: the call it is written in,
// or every reference to the name it is declared under.
func (r *suiteResolver) functionUses(function *ast.Node) []functionUse {
	expression := outermostWrapper(function)
	if use, ok := r.classifyUse(expression); ok {
		return []functionUse{use}
	}
	var declaration *ast.Node
	switch {
	case function.Kind == ast.KindFunctionDeclaration:
		declaration = function
	case aliasDeclaration(expression) != nil:
		declaration = expression.Parent
	default:
		return []functionUse{{kind: useUnknown}}
	}
	return r.declarationUses(declaration, map[*ast.Symbol]bool{})
}

// declarationUses lists the uses of the binding a function or variable
// declaration introduces, following `const alias = name` to the alias's uses.
func (r *suiteResolver) declarationUses(declaration *ast.Node, visited map[*ast.Symbol]bool) []functionUse {
	symbol := declaration.Symbol()
	if symbol == nil || r.ctx.Refs == nil {
		return []functionUse{{kind: useUnknown}}
	}
	if visited[symbol] {
		return nil
	}
	visited[symbol] = true

	var uses []functionUse
	if isExportedDeclaration(declaration) && !r.forHooks {
		uses = append(uses, functionUse{kind: useUnknown})
	}
	for _, reference := range r.ctx.Refs.References(symbol) {
		if utils.IsWriteReference(reference) {
			uses = append(uses, functionUse{kind: useUnknown})
			continue
		}
		expression := outermostWrapper(reference)
		if use, ok := r.classifyUse(expression); ok {
			uses = append(uses, use)
			continue
		}
		if alias := aliasDeclaration(expression); alias != nil {
			uses = append(uses, r.declarationUses(alias, visited)...)
			continue
		}
		if r.forHooks && isExportReference(expression) {
			continue
		}
		uses = append(uses, functionUse{kind: useUnknown})
	}
	return uses
}

// classifyUse recognizes an expression that is called, or passed as an
// argument to a call.
func (r *suiteResolver) classifyUse(expression *ast.Node) (functionUse, bool) {
	parent := expression.Parent
	if parent == nil || (parent.Kind != ast.KindCallExpression && parent.Kind != ast.KindNewExpression) {
		return functionUse{}, false
	}
	if parent.Expression() == expression {
		return functionUse{kind: useCall, call: parent}, true
	}
	if parent.Kind == ast.KindCallExpression && r.isDescribe(parent) {
		return functionUse{kind: useDescribe, call: parent}, true
	}
	// A callback handed to any other call is assumed to run there.
	return functionUse{kind: useCall, call: parent}, true
}

// aliasDeclaration returns the variable declaration expression initializes,
// as in `const install = setup`.
func aliasDeclaration(expression *ast.Node) *ast.Node {
	parent := expression.Parent
	if parent == nil || parent.Kind != ast.KindVariableDeclaration ||
		parent.AsVariableDeclaration().Initializer != expression || parent.Name().Kind != ast.KindIdentifier {
		return nil
	}
	return parent
}

func isExportedDeclaration(declaration *ast.Node) bool {
	if declaration.Kind == ast.KindFunctionDeclaration {
		return ast.HasSyntacticModifier(declaration, ast.ModifierFlagsExport)
	}
	if list := declaration.Parent; list != nil {
		if statement := list.Parent; statement != nil && statement.Kind == ast.KindVariableStatement {
			return ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport)
		}
	}
	return false
}

// isExportReference reports `export { name }` and `export default name`,
// which hand the function to other files without calling it here.
func isExportReference(expression *ast.Node) bool {
	parent := expression.Parent
	return parent != nil && (parent.Kind == ast.KindExportSpecifier || parent.Kind == ast.KindExportAssignment)
}

// outermostWrapper climbs from node through parentheses and type assertions,
// which do not change the value an enclosing expression receives.
func outermostWrapper(node *ast.Node) *ast.Node {
	for node.Parent != nil {
		switch node.Parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindAsExpression, ast.KindSatisfiesExpression,
			ast.KindNonNullExpression, ast.KindTypeAssertionExpression:
			node = node.Parent
		default:
			return node
		}
	}
	return node
}

func appendUnique(nodes []*ast.Node, node *ast.Node) []*ast.Node {
	for _, existing := range nodes {
		if existing == node {
			return nodes
		}
	}
	return append(nodes, node)
}
