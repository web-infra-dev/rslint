package no_array_concat_in_loop

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/plugins/unicorn/unicornutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

const messageID = "no-array-concat-in-loop"

var message = rule.RuleMessage{
	Id:          messageID,
	Description: "Do not use `Array#concat()` to accumulate an array in a loop.",
}

var NoArrayConcatInLoopRule = rule.Rule{
	Name:   "unicorn/no-array-concat-in-loop",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var variables *scopeVariableFinder
		return rule.RuleListeners{
			ast.KindBinaryExpression: func(node *ast.Node) {
				checkAssignment(ctx, node, &variables)
			},
		}
	},
}

func checkAssignment(ctx rule.RuleContext, node *ast.Node, variables **scopeVariableFinder) {
	binary := node.AsBinaryExpression()
	if binary == nil || binary.OperatorToken == nil ||
		binary.OperatorToken.Kind != ast.KindEqualsToken {
		return
	}

	// ESTree splits `a = b` into `AssignmentExpression` and, inside a
	// destructuring pattern, `AssignmentPattern`; upstream only visits the
	// former. TypeScript parses both as an equals `BinaryExpression`, so a
	// destructuring default such as `[result = result.concat(x)] = source`
	// would otherwise be reported. It assigns once when the destructured
	// value is `undefined` rather than accumulating, so skip any equals
	// expression that is itself an assignment target.
	if node.Parent == nil || ast.IsAssignmentTarget(node) {
		return
	}

	left := ast.SkipParentheses(binary.Left)
	if left == nil || !ast.IsIdentifier(left) {
		return
	}

	right := utils.SkipAssertionsAndParens(binary.Right)
	minimumArguments := 1
	call, ok := unicornutil.MatchDotMethodCall(right, unicornutil.DotMethodCallOptions{
		Method:           "concat",
		MinimumArguments: &minimumArguments,
	})
	if !ok {
		return
	}

	receiver := utils.SkipAssertionsAndParens(call.Object)
	if receiver == nil || !ast.IsIdentifier(receiver) || ctx.Refs == nil {
		return
	}
	// Two differently spelled identifiers cannot denote one lexical variable.
	// Reject them before constructing the scope model for this file.
	if left.Text() != receiver.Text() {
		return
	}

	loop := nearestLoop(node)
	if loop == nil || ctx.SourceFile == nil {
		return
	}

	if *variables == nil {
		*variables = newScopeVariableFinder(ctx.SourceFile)
	}
	variable := (*variables).find(left)
	if !variable.sameBinding((*variables).find(receiver)) ||
		isGlobalScopeVariable(ctx, variable) {
		return
	}

	declaration := mutableEmptyArrayDeclaration(variable)
	if declaration == nil || nodeInside(declaration, loopBody(loop)) {
		return
	}

	ctx.ReportNode(call.Property, message)
}

type scopeVariable struct {
	scope       *scope.Scope
	name        string
	definitions []*scope.Variable
	found       bool
}

func (variable scopeVariable) sameBinding(other scopeVariable) bool {
	return variable.found && other.found && variable.scope == other.scope && variable.name == other.name
}

type scopeVariableFinder struct {
	manager      *scope.Manager
	scopeByBlock map[*ast.Node]*scope.Scope
}

func newScopeVariableFinder(sourceFile *ast.SourceFile) *scopeVariableFinder {
	manager := scope.Build(sourceFile, scope.Options{})
	scopeByBlock := make(map[*ast.Node]*scope.Scope, len(manager.Scopes))
	for _, current := range manager.Scopes {
		if current.Block != nil {
			// A named function expression owns an outer name scope and an inner
			// function scope with the same block. Build creates them in that
			// order, so the later entry is the scope getScope() acquires first.
			scopeByBlock[current.Block] = current
		}
	}
	return &scopeVariableFinder{manager: manager, scopeByBlock: scopeByBlock}
}

func (finder *scopeVariableFinder) find(identifier *ast.Node) scopeVariable {
	if finder == nil || identifier == nil || identifier.Kind != ast.KindIdentifier {
		return scopeVariable{}
	}
	name := identifier.Text()
	for current := finder.acquire(identifier); current != nil; current = current.Parent {
		definitions := current.Declarations(name)
		if hasAuthoredDefinition(definitions) {
			return scopeVariable{scope: current, name: name, definitions: definitions, found: true}
		}
		// eslint-scope creates an `arguments` variable with no definitions in
		// every non-arrow runtime function. An authored declaration of the
		// same name joins that variable and was handled above.
		if name == "arguments" && current.Kind == scope.KindFunction &&
			hasImplicitArguments(current.Block) {
			return scopeVariable{scope: current, name: name, found: true}
		}
	}
	return scopeVariable{}
}

func (finder *scopeVariableFinder) acquire(identifier *ast.Node) *scope.Scope {
	var child *ast.Node
	for current := identifier; current != nil; current = current.Parent {
		if acquired := finder.scopeByBlock[current]; acquired != nil &&
			!outsideRuntimeMethodScope(current, child) {
			return acquired
		}
		child = current
	}
	return finder.manager.Global
}

// TSESTree represents a runtime method as a MethodDefinition wrapping a
// FunctionExpression. Its computed key and member-level decorators sit
// outside that function, while parameters (including their decorators), type
// parameters, return type, and body sit inside it. ts-go combines both parts
// into one method node, so acquisition must skip that node only for the two
// outer positions. TypeScript method signatures are not combined runtime
// methods and deliberately keep their function-type scope here.
func outsideRuntimeMethodScope(method *ast.Node, child *ast.Node) bool {
	if method == nil || child == nil {
		return false
	}
	switch method.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		return child == method.Name() || child.Kind == ast.KindDecorator
	default:
		return false
	}
}

func hasImplicitArguments(block *ast.Node) bool {
	if block == nil {
		return false
	}
	switch block.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindMethodDeclaration, ast.KindGetAccessor,
		ast.KindSetAccessor, ast.KindConstructor:
		return true
	default:
		return false
	}
}

func hasAuthoredDefinition(definitions []*scope.Variable) bool {
	for _, definition := range definitions {
		if definition != nil && definition.DefNode != nil &&
			!utils.IsJSDocSyntaxNode(definition.DefNode) {
			return true
		}
	}
	return false
}

func isGlobalScopeVariable(ctx rule.RuleContext, variable scopeVariable) bool {
	return variable.found && variable.scope != nil &&
		variable.scope.Kind == scope.KindGlobal && !ctx.Refs.HasNonGlobalProgramScope()
}

func mutableEmptyArrayDeclaration(variable scopeVariable) *ast.Node {
	var definition *scope.Variable
	for _, candidate := range variable.definitions {
		if candidate == nil || utils.IsJSDocSyntaxNode(candidate.DefNode) {
			continue
		}
		if definition != nil {
			return nil
		}
		definition = candidate
	}
	if definition == nil || definition.Kind != scope.DefVariable {
		return nil
	}

	declarationNode := definition.DefNode
	if declarationNode == nil || declarationNode.Kind != ast.KindVariableDeclaration {
		return nil
	}
	declaration := declarationNode.AsVariableDeclaration()
	if declaration == nil || declaration.Name() == nil ||
		!ast.IsIdentifier(declaration.Name()) || declaration.Initializer == nil {
		return nil
	}

	declarationList := declarationNode.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList {
		return nil
	}
	kind := utils.GetVarDeclListKind(declarationList)
	if kind != "let" && kind != "var" {
		return nil
	}

	initializer := utils.SkipAssertionsAndParens(declaration.Initializer)
	if initializer == nil || !ast.IsEmptyArrayLiteral(initializer) {
		return nil
	}
	return declarationNode
}

func nearestLoop(node *ast.Node) *ast.Node {
	for ancestor := node.Parent; ancestor != nil; ancestor = ancestor.Parent {
		if isFunctionBoundary(node, ancestor) {
			return nil
		}
		if ast.IsIterationStatement(ancestor, false) {
			return ancestor
		}
	}
	return nil
}

// isFunctionBoundary mirrors unicorn's three ESTree runtime function kinds.
// Bodyless TypeScript declarations and signatures are different ESTree node
// kinds and do not stop the search. ts-go also combines each runtime method's
// outer MethodDefinition positions with its inner FunctionExpression, so the
// computed key and member-level decorators must remain outside the boundary.
func isFunctionBoundary(node *ast.Node, function *ast.Node) bool {
	if function == nil || function.Body() == nil {
		return false
	}

	switch function.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		return true
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor:
		if nodeInside(node, function.Name()) {
			return false
		}
		for current := node.Parent; current != nil && current != function; current = current.Parent {
			if current.Kind == ast.KindDecorator && current.Parent == function {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func loopBody(loop *ast.Node) *ast.Node {
	if loop == nil {
		return nil
	}
	switch loop.Kind {
	case ast.KindForStatement:
		return loop.AsForStatement().Statement
	case ast.KindForInStatement, ast.KindForOfStatement:
		return loop.AsForInOrOfStatement().Statement
	case ast.KindWhileStatement:
		return loop.AsWhileStatement().Statement
	case ast.KindDoStatement:
		return loop.AsDoStatement().Statement
	default:
		return nil
	}
}

func nodeInside(node *ast.Node, ancestor *ast.Node) bool {
	return node != nil && ancestor != nil &&
		node.Pos() >= ancestor.Pos() && node.End() <= ancestor.End()
}
