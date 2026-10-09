package dynamic_import_chunkname

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// containsContextError reports syntax that is an error because of where it is,
// which the TypeScript parser leaves to later passes: a `return` or `arguments`
// in a class static block, `arguments` in a field initializer, an `await`
// outside an async function, and a `break` or `continue` without a target.
func containsContextError(root *ast.Node) bool {
	checker := &contextChecker{}
	root.ForEachChild(checker.visit)
	return checker.invalid
}

// contextState is what the code around a node allows. It is saved and replaced
// at every function boundary.
type contextState struct {
	argumentsForbidden bool
	returnForbidden    bool
	awaitAllowed       bool
	// breakable counts the enclosing loops and switches, loops the loops alone.
	breakable, loops int
	labels           []jumpLabel
}

type jumpLabel struct {
	name   string
	isLoop bool
}

type contextChecker struct {
	contextState
	invalid bool
}

func (c *contextChecker) fail() bool {
	c.invalid = true
	return true
}

func (c *contextChecker) visit(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionExpression, ast.KindFunctionDeclaration, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return c.visitInContext(node, contextState{awaitAllowed: hasModifier(node, ast.KindAsyncKeyword)})
	case ast.KindArrowFunction:
		// An arrow function has no `arguments` of its own.
		return c.visitInContext(node, contextState{
			argumentsForbidden: c.argumentsForbidden,
			awaitAllowed:       hasModifier(node, ast.KindAsyncKeyword),
		})
	case ast.KindClassStaticBlockDeclaration:
		return c.visitInContext(node, contextState{argumentsForbidden: true, returnForbidden: true})
	case ast.KindPropertyDeclaration:
		property := node.AsPropertyDeclaration()
		if name := property.Name(); name != nil && c.visit(name) {
			return true
		}
		if property.Initializer == nil {
			return false
		}
		saved := c.contextState
		c.contextState = contextState{argumentsForbidden: true}
		stop := c.visit(property.Initializer)
		c.contextState = saved
		return stop
	case ast.KindReturnStatement:
		if c.returnForbidden {
			return c.fail()
		}
	case ast.KindAwaitExpression:
		if !c.awaitAllowed {
			return c.fail()
		}
	case ast.KindIdentifier:
		if c.argumentsForbidden && node.Text() == "arguments" && !utils.IsNonReferenceIdentifier(node) {
			return c.fail()
		}
	case ast.KindLabeledStatement:
		labeled := node.AsLabeledStatement()
		body := labeled.Statement
		for body.Kind == ast.KindLabeledStatement {
			body = body.AsLabeledStatement().Statement
		}
		c.labels = append(c.labels, jumpLabel{name: labeled.Label.Text(), isLoop: isLoop(body)})
		stop := node.ForEachChild(c.visit)
		c.labels = c.labels[:len(c.labels)-1]
		return stop
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindWhileStatement, ast.KindDoStatement:
		c.breakable++
		c.loops++
		stop := node.ForEachChild(c.visit)
		c.breakable--
		c.loops--
		return stop
	case ast.KindSwitchStatement:
		c.breakable++
		stop := node.ForEachChild(c.visit)
		c.breakable--
		return stop
	case ast.KindBreakStatement:
		if label := node.AsBreakStatement().Label; label != nil {
			if !c.hasLabel(label.Text(), false) {
				return c.fail()
			}
		} else if c.breakable == 0 {
			return c.fail()
		}
	case ast.KindContinueStatement:
		if label := node.AsContinueStatement().Label; label != nil {
			if !c.hasLabel(label.Text(), true) {
				return c.fail()
			}
		} else if c.loops == 0 {
			return c.fail()
		}
	}
	return node.ForEachChild(c.visit)
}

// visitInContext visits node's children with a context that replaces the
// current one.
func (c *contextChecker) visitInContext(node *ast.Node, state contextState) bool {
	saved := c.contextState
	c.contextState = state
	stop := node.ForEachChild(c.visit)
	c.contextState = saved
	return stop
}

// hasLabel reports an enclosing label with the name; for continue the label
// must name a loop.
func (c *contextChecker) hasLabel(name string, loopOnly bool) bool {
	for i := len(c.labels) - 1; i >= 0; i-- {
		if c.labels[i].name == name {
			return !loopOnly || c.labels[i].isLoop
		}
	}
	return false
}

func isLoop(statement *ast.Node) bool {
	switch statement.Kind {
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindWhileStatement, ast.KindDoStatement:
		return true
	}
	return false
}

// hasModifier reports whether node has the modifier.
func hasModifier(node *ast.Node, kind ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier.Kind == kind {
			return true
		}
	}
	return false
}
