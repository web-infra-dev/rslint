package code_path_analysis

import (
	"math"

	"github.com/microsoft/typescript-go/shim/ast"
)

type CodePathAnalyzer struct {
	currentNode *ast.Node
	codePath    *CodePath
	idGenerator *IdGenerator

	// Maintain code segment path stack as we traverse.
	onCodePathSegmentStart func(segment *CodePathSegment, node *ast.Node)
	onCodePathSegmentEnd   func(segment *CodePathSegment, node *ast.Node)

	// Maintain code path stack as we traverse.
	onCodePathStart func(codePath *CodePath, node *ast.Node)
	onCodePathEnd   func(codePath *CodePath, node *ast.Node)

	onCodePathSegmentLoop func(fromSegment *CodePathSegment, toSegment *CodePathSegment, node *ast.Node)
}

func NewCodePathAnalyzer(
	onCodePathSegmentStart func(segment *CodePathSegment, node *ast.Node),
	onCodePathSegmentEnd func(segment *CodePathSegment, node *ast.Node),
	onCodePathStart func(codePath *CodePath, node *ast.Node),
	onCodePathEnd func(codePath *CodePath, node *ast.Node),
	onCodePathSegmentLoop func(fromSegment *CodePathSegment, toSegment *CodePathSegment, node *ast.Node),
) *CodePathAnalyzer {
	return &CodePathAnalyzer{
		currentNode:            nil,
		codePath:               nil,
		idGenerator:            NewIdGenerator("s"),
		onCodePathSegmentStart: onCodePathSegmentStart,
		onCodePathSegmentEnd:   onCodePathSegmentEnd,
		onCodePathStart:        onCodePathStart,
		onCodePathEnd:          onCodePathEnd,
		onCodePathSegmentLoop:  onCodePathSegmentLoop,
	}
}

func (analyzer *CodePathAnalyzer) State() *CodePathState {
	codePath := analyzer.codePath
	var state *CodePathState
	if codePath != nil {
		state = codePath.State()
	}
	return state
}

// Does the process to enter a given AST node.
// This updates state of analysis and calls `enterNode` of the wrapped.
func (analyzer *CodePathAnalyzer) EnterNode(node *ast.Node) {
	analyzer.currentNode = node

	// Updates the code path due to node's position in its parent node.
	if node.Parent != nil {
		analyzer.preprocess(node)
	}

	// Updates the code path.
	// And emits onCodePathStart/onCodePathSegmentStart events.
	analyzer.processCodePathToEnter(node)

	analyzer.currentNode = nil
}

// Does the process to leave a given AST node.
// This updates state of analysis and calls `leaveNode` of the wrapped.
func (analyzer *CodePathAnalyzer) LeaveNode(node *ast.Node) {
	analyzer.currentNode = node

	analyzer.processCodePathToExit(node)

	analyzer.postprocess(node)

	analyzer.currentNode = nil
}

// Updates the code path due to the position of a given node in the parent node thereof.
//
// For example, if the node is `parent.consequent`, this creates a fork from the current path.
func (analyzer *CodePathAnalyzer) preprocess(node *ast.Node) {
	state := analyzer.State()
	parent := node.Parent

	switch parent.Kind {
	// The `arguments.length == 0` case is in `postprocess` function.
	case ast.KindCallExpression:
		if ast.IsOptionalChain(parent) && len(parent.Arguments()) >= 1 && parent.Arguments()[0] == node {
			state.MakeOptionalRight()
		}

	case ast.KindPropertyAccessExpression:
		// Corresponds to ESLint's MemberExpression
		expr := parent.AsPropertyAccessExpression()
		if ast.IsOptionalChain(parent) && len(expr.Properties()) > 0 && expr.Properties()[0] == node {
			state.MakeOptionalRight()
		}

	case ast.KindBinaryExpression:
		// Handle LogicalExpression (&&, ||, ??)
		binExpr := parent.AsBinaryExpression()
		if binExpr.Right == node && isHandledLogicalOperator(binExpr.OperatorToken.Kind) {
			state.MakeLogicalRight()
		}

	case ast.KindConditionalExpression:
		// Handle ternary operator: condition ? consequent : alternate
		condExpr := parent.AsConditionalExpression()
		if condExpr.WhenTrue == node {
			state.MakeIfConsequent()
		} else if condExpr.WhenFalse == node {
			state.MakeIfAlternate()
		}

	case ast.KindIfStatement:
		// Handle if-else statements
		ifStmt := parent.AsIfStatement()
		if ifStmt.ThenStatement == node {
			state.MakeIfConsequent()
		} else if ifStmt.ElseStatement == node {
			state.MakeIfAlternate()
		}

	case ast.KindCaseClause:
		// Handle switch case body
		caseStmts := parent.AsCaseOrDefaultClause().Statements.Nodes
		if caseStmts[0] == node {
			state.MakeSwitchCaseBody(false, false)
		}

	case ast.KindDefaultClause:
		defaultStmts := parent.AsCaseOrDefaultClause().Statements.Nodes
		if defaultStmts[0] == node {
			state.MakeSwitchCaseBody(false, true)
		}

	case ast.KindTryStatement:
		// Handle try-catch-finally
		tryStmt := parent.AsTryStatement()
		if tryStmt.CatchClause == node {
			state.MakeCatchBlock()
		} else if tryStmt.FinallyBlock == node {
			state.MakeFinallyBlock()
		}

	case ast.KindWhileStatement:
		// Handle while loops
		whileStmt := parent.AsWhileStatement()
		if whileStmt.Expression == node {
			state.MakeWhileTest(getBooleanValueIfSimpleConstant(node))
		} else if whileStmt.Statement == node {
			state.MakeWhileBody()
		}

	case ast.KindDoStatement:
		// Handle do-while loops
		doStmt := parent.AsDoStatement()
		if doStmt.Statement == node {
			state.MakeDoWhileBody()
		} else if doStmt.Expression == node {
			state.MakeDoWhileTest(getBooleanValueIfSimpleConstant(node))
		}

	case ast.KindForStatement:
		// Handle for loops
		forStmt := parent.AsForStatement()
		if forStmt.Condition == node {
			state.MakeForTest(getBooleanValueIfSimpleConstant(node))
		} else if forStmt.Incrementor == node {
			state.MakeForUpdate()
		} else if forStmt.Statement == node {
			state.MakeForBody()
		}

	case ast.KindForInStatement:
		// Handle for-in loops
		forInStmt := parent.AsForInOrOfStatement()
		if forInStmt.Initializer == node {
			state.MakeForInOfLeft()
		} else if forInStmt.Expression == node {
			state.MakeForInOfRight()
		} else if forInStmt.Statement == node {
			state.MakeForInOfBody()
		}

	case ast.KindForOfStatement:
		// Handle for-of loops
		forOfStmt := parent.AsForInOrOfStatement()
		if forOfStmt.Initializer == node {
			state.MakeForInOfLeft()
		} else if forOfStmt.Expression == node {
			state.MakeForInOfRight()
		} else if forOfStmt.Statement == node {
			state.MakeForInOfBody()
		}

	case ast.KindBindingElement:
		// Handle assignment patterns (destructuring with defaults)
		bindingElem := parent.AsBindingElement()
		if bindingElem.Initializer == node {
			state.PushForkContext(nil)
			state.ForkBypassPath()
			state.ForkPath()
		}
	}
}

func (analyzer *CodePathAnalyzer) processCodePathToEnter(node *ast.Node) {
	// Special case: The right side of class field initializer is considered
	// to be its own function, so we need to start a new code path in this case.
	if isPropertyDefinitionValue(node) {
		analyzer.startCodePath("class-field-initializer", node)

		/*
		 * Intentional fall through because `node` needs to also be
		 * processed by the code below. For example, if we have:
		 *
		 * class Foo {
		 *     a = () => {}
		 * }
		 *
		 * In this case, we also need start a second code path.
		 */
	}

	state := analyzer.State()
	parent := node.Parent

	switch node.Kind {
	case ast.KindSourceFile:
		analyzer.startCodePath("program", node)

	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		analyzer.startCodePath("function", node)

	case ast.KindClassStaticBlockDeclaration:
		analyzer.startCodePath("class-static-block", node)

	case ast.KindCallExpression:
		if ast.IsOptionalChain(node) {
			state.MakeOptionalNode()
		}

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		if ast.IsOptionalChain(node) {
			state.MakeOptionalNode()
		}

	case ast.KindBinaryExpression:
		// Handle LogicalExpression (&&, ||, ??)
		binExpr := node.AsBinaryExpression()
		if isHandledLogicalOperator(binExpr.OperatorToken.Kind) {
			state.PushChoiceContext(tokenToText[binExpr.OperatorToken.Kind], isForkingByTrueOrFalse(node))
		} else if isLogicalAssignmentOperator(binExpr.OperatorToken.Kind) {
			text := tokenToText[binExpr.OperatorToken.Kind]
			// removes `=` from the end
			text = text[:len(text)-1]
			state.PushChoiceContext(text, isForkingByTrueOrFalse(node))
		}

	case ast.KindConditionalExpression, ast.KindIfStatement:
		state.PushChoiceContext("test", false)

	case ast.KindSwitchStatement:
		switchStmt := node.AsSwitchStatement()
		hasDefaultCase := false
		for _, clause := range switchStmt.CaseBlock.AsCaseBlock().Clauses.Nodes {
			if ast.IsDefaultClause(clause) {
				hasDefaultCase = true
				break
			}
		}
		label := getLabel(node)
		state.PushSwitchContext(hasDefaultCase, label)

	case ast.KindTryStatement:
		tryStmt := node.AsTryStatement()
		hasFinalizer := tryStmt.FinallyBlock != nil
		state.PushTryContext(hasFinalizer)

	case ast.KindCaseClause:
		// Fork if this node is after the 1st node in `cases`.
		if parent != nil && parent.Kind == ast.KindSwitchStatement {
			state.ForkPath()
		}

	case ast.KindWhileStatement:
		label := getLabel(node)
		state.PushLoopContext(WhileStatement, label)
	case ast.KindDoStatement:
		label := getLabel(node)
		state.PushLoopContext(DoWhileStatement, label)
	case ast.KindForStatement:
		label := getLabel(node)
		state.PushLoopContext(ForStatement, label)
	case ast.KindForInStatement:
		label := getLabel(node)
		state.PushLoopContext(ForInStatement, label)
	case ast.KindForOfStatement:
		label := getLabel(node)
		state.PushLoopContext(ForOfStatement, label)
	case ast.KindLabeledStatement:
		if !isBreakableType(node.Body().Kind) {
			state.PushBreakContext(false, node.Label().Text())
		}
	default:
		// No special handling needed
	}

	// Emits onCodePathSegmentStart events if updated.
	analyzer.forwardCurrentToHead(node)
}

// Updates the code path due to the type of a given node in leaving.
func (analyzer *CodePathAnalyzer) processCodePathToExit(node *ast.Node) {
	state := analyzer.State()
	if state == nil {
		return
	}

	dontForward := false

	switch node.Kind {
	// !!! ChainExpression
	case ast.KindIfStatement, ast.KindConditionalExpression:
		state.PopChoiceContext()

	case ast.KindBinaryExpression:
		// Handle LogicalExpression (&&, ||, ??)
		binExpr := node.AsBinaryExpression()
		if isHandledLogicalOperator(binExpr.OperatorToken.Kind) ||
			isLogicalAssignmentOperator(binExpr.OperatorToken.Kind) {
			state.PopBreakContext()
		}

	case ast.KindSwitchStatement:
		state.PopSwitchContext()

	case ast.KindCaseClause:
		// This is the same as the process at the 1st `consequent` node in preprocess function.
		// Must do if this `consequent` is empty.
		caseClause := node.AsCaseOrDefaultClause()
		if len(caseClause.Statements.Nodes) == 0 {
			isDefault := caseClause.Expression == nil
			state.MakeSwitchCaseBody(true, isDefault)
		}
		if state.forkContext.IsReachable() {
			dontForward = true
		}

	case ast.KindTryStatement:
		state.PopTryContext()

	case ast.KindBreakStatement:
		analyzer.forwardCurrentToHead(node)
		breakStmt := node.AsBreakStatement()
		label := ""
		if breakStmt.Label != nil {
			label = breakStmt.Label.Text()
		}
		state.MakeBreak(label)
		dontForward = true

	case ast.KindContinueStatement:
		analyzer.forwardCurrentToHead(node)
		continueStmt := node.AsContinueStatement()
		label := ""
		if continueStmt.Label != nil {
			label = continueStmt.Label.Text()
		}
		state.MakeContinue(label)
		dontForward = true

	case ast.KindReturnStatement:
		analyzer.forwardCurrentToHead(node)
		state.MakeReturn()
		dontForward = true

	case ast.KindThrowStatement:
		analyzer.forwardCurrentToHead(node)
		state.MakeThrow()
		dontForward = true

	case ast.KindIdentifier:
		// TODO: Implement isIdentifierReference check
		// if analyzer.isIdentifierReference(node) {
		//     state.MakeFirstThrowablePathInTryBlock()
		//     dontForward = true
		// }

	case ast.KindCallExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindNewExpression, ast.KindYieldExpression:
		state.MakeFirstThrowablePathInTryBlock()

	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		state.PopLoopContext()

	case ast.KindBindingElement:
		state.PopForkContext()

	case ast.KindLabeledStatement:
		labeledStmt := node.AsLabeledStatement()
		if !isBreakableType(labeledStmt.Body().Kind) {
			state.PopBreakContext()
		}

	default:
		// No special handling needed
	}

	// Emits onCodePathSegmentStart events if updated.
	if !dontForward {
		analyzer.forwardCurrentToHead(node)
	}
}

func (analyzer *CodePathAnalyzer) postprocess(node *ast.Node) {
	switch node.Kind {
	case ast.KindSourceFile,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindArrowFunction,
		ast.KindClassStaticBlockDeclaration:
		analyzer.endCodePath(node)

	// The `arguments.length >= 1` case is in `preprocess` function.
	case ast.KindCallExpression:
		callExpr := node.AsCallExpression()
		if ast.IsOptionalChain(node) && len(callExpr.Arguments.Nodes) == 0 {
			if analyzer.codePath != nil {
				analyzer.codePath.state.MakeOptionalRight()
			}
		}

	default:
		// No special handling needed
	}

	// Special case: The right side of class field initializer is considered
	// to be its own function, so we need to end a code path in this case.
	if isPropertyDefinitionValue(node) {
		analyzer.endCodePath(node)
	}
}

func (analyzer *CodePathAnalyzer) startCodePath(origin string, node *ast.Node) {
	codePath := analyzer.codePath
	if codePath != nil {
		// Emits onCodePathSegmentStart events if updated.
		analyzer.forwardCurrentToHead(node)
	}

	// Create the code path of this scope.
	analyzer.codePath = NewCodePath(
		analyzer.idGenerator.Next(),
		origin,
		codePath,
		analyzer.onLooped,
	)

	if analyzer.onCodePathStart != nil {
		analyzer.onCodePathStart(codePath, node)
	}
}

// Ends the code path for the current node.
func (analyzer *CodePathAnalyzer) endCodePath(node *ast.Node) {
	codePath := analyzer.codePath
	if codePath == nil {
		return
	}

	// Mark the current path as the final node.
	codePath.state.MakeFinal()

	// Emits onCodePathSegmentEnd event of the current segments.
	analyzer.leaveFromCurrentSegment(node)

	// Emits onCodePathEnd event of this code path.
	if analyzer.onCodePathEnd != nil {
		analyzer.onCodePathEnd(codePath, node)
	}

	analyzer.codePath = codePath.upper
}

func (analyzer *CodePathAnalyzer) onLooped(fromSegment *CodePathSegment, toSegment *CodePathSegment) {
	if fromSegment.reachable && toSegment.reachable {
		if analyzer.onCodePathSegmentLoop != nil {
			analyzer.onCodePathSegmentLoop(fromSegment, toSegment, analyzer.currentNode)
		}
	}
}

// Updates the current segment with the head segment.
// This is similar to local branches and tracking branches of git.
//
// To separate the current and the head is in order to not make useless segments.
//
// In this process, both "onCodePathSegmentStart" and "onCodePathSegmentEnd"
// events are fired.
func (analyzer *CodePathAnalyzer) forwardCurrentToHead(node *ast.Node) {
	state := analyzer.State()
	currentSegments := state.currentSegments
	headSegments := state.HeadSegments()
	end := int(math.Max(float64(len(currentSegments)), float64(len(headSegments))))

	if analyzer.onCodePathSegmentEnd != nil {
		for i := range end {
			var currentSegment *CodePathSegment
			var headSegment *CodePathSegment

			if i < len(currentSegments) {
				currentSegment = currentSegments[i]
			}
			if i < len(headSegments) {
				headSegment = headSegments[i]
			}

			if currentSegment != headSegment && currentSegment != nil {
				if currentSegment.reachable {
					analyzer.onCodePathSegmentEnd(currentSegment, node)
				}
			}
		}
	}

	// Update state.
	state.currentSegments = headSegments

	if analyzer.onCodePathSegmentStart != nil {
		for i := range end {
			var currentSegment *CodePathSegment
			var headSegment *CodePathSegment

			if i < len(currentSegments) {
				currentSegment = currentSegments[i]
			}
			if i < len(headSegments) {
				headSegment = headSegments[i]
			}

			if currentSegment != headSegment && headSegment != nil {
				markUsed(headSegment)
				if headSegment.reachable {
					analyzer.onCodePathSegmentStart(headSegment, node)
				}
			}
		}
	}
}

// Updates the current segment with empty.
// This is called at the last of functions or the program.
func (analyzer *CodePathAnalyzer) leaveFromCurrentSegment(node *ast.Node) {
	state := analyzer.State()
	currentSegments := state.currentSegments

	for _, currentSegment := range currentSegments {
		if currentSegment.reachable {
			analyzer.onCodePathSegmentEnd(currentSegment, node)
		}
	}

	state.currentSegments = make([]*CodePathSegment, 0)
}

// Checks if a given node appears as the value of a PropertyDefinition node.
func isPropertyDefinitionValue(node *ast.Node) bool {
	parent := node.Parent

	return parent != nil && ast.IsPropertyDeclaration(parent) && parent.AsPropertyDeclaration().Initializer == node
}

// Checks whether the given logical operator is taken into account for the code path analysis.
func isHandledLogicalOperator(operatorKind ast.Kind) bool {
	return operatorKind == ast.KindBarBarToken || operatorKind == ast.KindAmpersandAmpersandToken || operatorKind == ast.KindQuestionQuestionToken
}

// Checks whether the given assignment operator is a logical assignment operator.
// Logical assignments are taken into account for the code path analysis
// because of their short-circuiting semantics.
func isLogicalAssignmentOperator(operatorKind ast.Kind) bool {
	return operatorKind == ast.KindAmpersandAmpersandEqualsToken || operatorKind == ast.KindBarBarEqualsToken || operatorKind == ast.KindQuestionQuestionEqualsToken
}

// Checks whether or not a given logical expression node goes different path
// between the `true` case and the `false` case.
func isForkingByTrueOrFalse(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindConditionalExpression:
		condExpr := parent.AsConditionalExpression()
		return condExpr.Condition == node

	case ast.KindIfStatement:
		ifStmt := parent.AsIfStatement()
		return ifStmt.Expression == node

	case ast.KindWhileStatement:
		whileStmt := parent.AsWhileStatement()
		return whileStmt.Expression == node

	case ast.KindDoStatement:
		doStmt := parent.AsDoStatement()
		return doStmt.Expression == node

	case ast.KindForStatement:
		forStmt := parent.AsForStatement()
		return forStmt.Condition == node

	case ast.KindBinaryExpression:
		binExpr := parent.AsBinaryExpression()
		return isHandledLogicalOperator(binExpr.OperatorToken.Kind) || isLogicalAssignmentOperator(binExpr.OperatorToken.Kind)

	default:
		return false
	}
}

// Gets the boolean value of a given literal node.
//
// This is used to detect infinity loops (e.g. `while (true) {}`).
// Statements preceded by an infinity loop are unreachable if the loop didn't
// have any `break` statement.
func getBooleanValueIfSimpleConstant(node *ast.Node) bool {
	if node.Kind == ast.KindTrueKeyword {
		return true
	}
	if node.Kind == ast.KindFalseKeyword {
		return false
	}
	if node.Kind == ast.KindNumericLiteral {
		numLiteral := node.AsNumericLiteral()
		// In JavaScript, any non-zero number is truthy, zero is falsy
		if numLiteral.Text == "0" {
			return false
		} else {
			return true
		}
	}
	if node.Kind == ast.KindStringLiteral {
		strLiteral := node.AsStringLiteral()
		// In JavaScript, empty string is falsy, non-empty string is truthy
		if strLiteral.Text == `""` || strLiteral.Text == `''` {
			return false
		} else {
			return true
		}
	}
	if node.Kind == ast.KindNullKeyword {
		return false
	}
	// Return nil for non-literal nodes or literals we can't determine
	return false
}

// Gets the label if the parent node of a given node is a LabeledStatement.
func getLabel(node *ast.Node) string {
	if ast.IsLabeledStatement(node.Parent) {
		return node.Parent.Label().Text()
	}
	return ""
}

func isBreakableType(kind ast.Kind) bool {
	switch kind {
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindSwitchStatement:
		return true
	default:
		return false
	}
}

var tokenToText = map[ast.Kind]string{
	ast.KindAmpersandAmpersandToken:       "&&",
	ast.KindBarBarToken:                   "||",
	ast.KindQuestionQuestionToken:         "??",
	ast.KindAmpersandAmpersandEqualsToken: "&&=",
	ast.KindBarBarEqualsToken:             "||=",
	ast.KindQuestionQuestionEqualsToken:   "??=",
}
