package prefer_expect_assertions

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// unknownSuite stands for a suite the resolver cannot identify, such as the
// callers of an exported function. A hook registered from an unknown place may
// cover any test, and a test registered from one may be covered by any hook,
// so both answers lean towards not reporting.
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
// called, so a hook inside a helper belongs to the suites that call the helper,
// not to the scope the helper is written in.
type suiteResolver struct {
	ctx        rule.RuleContext
	isDescribe func(call *ast.Node) bool
	suites     map[*ast.Node][]*ast.Node
	parents    map[*ast.Node][]*ast.Node
	inProgress map[*ast.Node]bool
}

func newSuiteResolver(ctx rule.RuleContext, isDescribe func(call *ast.Node) bool) *suiteResolver {
	return &suiteResolver{
		ctx:        ctx,
		isDescribe: isDescribe,
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
	if len(uses) == 0 {
		// Never referenced in this file: dead code, or reached some way the
		// resolver does not see.
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

// parentSuites returns the suites a suite body is nested in.
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

	var symbol *ast.Symbol
	exported := false
	switch {
	case function.Kind == ast.KindFunctionDeclaration:
		symbol = function.Symbol()
		exported = ast.HasSyntacticModifier(function, ast.ModifierFlagsExport)
	case expression.Parent != nil && expression.Parent.Kind == ast.KindVariableDeclaration &&
		expression.Parent.AsVariableDeclaration().Initializer == expression &&
		expression.Parent.Name().Kind == ast.KindIdentifier:
		declaration := expression.Parent
		symbol = declaration.Symbol()
		if statement := declaration.Parent.Parent; statement != nil && statement.Kind == ast.KindVariableStatement {
			exported = ast.HasSyntacticModifier(statement, ast.ModifierFlagsExport)
		}
	}
	if symbol == nil || r.ctx.Refs == nil {
		return []functionUse{{kind: useUnknown}}
	}

	var uses []functionUse
	if exported {
		uses = append(uses, functionUse{kind: useUnknown})
	}
	for _, reference := range r.ctx.Refs.References(symbol) {
		if utils.IsWriteReference(reference) {
			uses = append(uses, functionUse{kind: useUnknown})
			continue
		}
		if use, ok := r.classifyUse(outermostWrapper(reference)); ok {
			uses = append(uses, use)
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
	// A callback handed to any other call is assumed to run there. That can
	// only widen what a hook covers, or where a test looks for coverage.
	return functionUse{kind: useCall, call: parent}, true
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
