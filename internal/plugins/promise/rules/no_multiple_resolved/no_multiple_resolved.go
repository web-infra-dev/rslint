package no_multiple_resolved

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// pathState tracks the resolution state along a code path.
// firstCall is the CallExpression node of the first resolver call on this path (nil if none).
// certain is true only when ALL paths reaching here have called a resolver.
// dead is true when this path always terminates (return/throw/break/continue).
type pathState struct {
	firstCall *ast.Node
	certain   bool
	dead      bool
}

type controlTarget struct {
	node      *ast.Node
	label     string
	exitState *pathState
	isLoop    bool
}

// ruleCtx holds shared analysis context.
// In TS mode (TypeChecker available), resolverSymbols is populated and used for
// scope-accurate resolver detection. In JS mode, resolverSymbols is nil and
// resolverNames provides a string-based fallback.
type ruleCtx struct {
	resolverNames   map[string]bool      // always populated; used as JS fallback + empty check
	resolverSymbols map[*ast.Symbol]bool // non-nil in TS mode only
	ctx             rule.RuleContext
	reported        map[*ast.Node]bool
	targets         []controlTarget
}

func buildAlreadyResolvedMessage(line int) rule.RuleMessage {
	return rule.RuleMessage{
		Description: fmt.Sprintf("Promise should not be resolved multiple times. Promise is already resolved on line %d.", line),
	}
}

func buildPotentiallyResolvedMessage(line int) rule.RuleMessage {
	return rule.RuleMessage{
		Description: fmt.Sprintf("Promise should not be resolved multiple times. Promise is potentially resolved on line %d.", line),
	}
}

func lineOf(node *ast.Node, ctx rule.RuleContext) int {
	lineMap := ctx.SourceFile.ECMALineMap()
	pos := utils.TrimNodeTextRange(ctx.SourceFile, node).Pos()
	return scanner.ComputeLineOfPosition(lineMap, pos) + 1
}

// mergeStates combines two branch states (e.g. if-then and if-else).
func mergeStates(a, b pathState) pathState {
	if a.dead && b.dead {
		return pathState{dead: true}
	}
	if a.dead {
		return b
	}
	if b.dead {
		return a
	}
	if a.firstCall == nil && b.firstCall == nil {
		return pathState{}
	}
	if a.firstCall != nil && b.firstCall != nil {
		return pathState{firstCall: a.firstCall, certain: a.certain && b.certain}
	}
	if a.firstCall != nil {
		return pathState{firstCall: a.firstCall, certain: false}
	}
	return pathState{firstCall: b.firstCall, certain: false}
}

// onResolverCall handles a resolver call in the current path state.
// Reports a violation if already resolved, then updates state.
func onResolverCall(callExpr *ast.Node, state pathState, rCtx *ruleCtx) pathState {
	if !rCtx.reported[callExpr] && state.firstCall != nil {
		line := lineOf(state.firstCall, rCtx.ctx)
		var msg rule.RuleMessage
		if state.certain {
			msg = buildAlreadyResolvedMessage(line)
		} else {
			msg = buildPotentiallyResolvedMessage(line)
		}
		rCtx.ctx.ReportNode(callExpr, msg)
		rCtx.reported[callExpr] = true
	}
	certain := state.certain || (callExpr.AsCallExpression().QuestionDotToken == nil)
	return pathState{firstCall: callExpr, certain: certain}
}

// isResolverIdent reports whether ident refers to a tracked resolver parameter.
// In TS mode uses TypeChecker symbol comparison (handles shadowing / re-declarations
// correctly). Falls back to name matching when TypeChecker is unavailable or when
// symbol resolution returns nil.
func isResolverIdent(ident *ast.Node, rCtx *ruleCtx) bool {
	if rCtx.resolverSymbols != nil {
		sym := rCtx.ctx.TypeChecker.GetSymbolAtLocation(ident)
		if sym != nil {
			return rCtx.resolverSymbols[sym]
		}
		// TypeChecker couldn't resolve (shouldn't happen in valid code); fall back.
	}
	return rCtx.resolverNames[ident.AsIdentifier().Text]
}

// isDirectResolverCall returns the CallExpression if expr is a direct call to a resolver.
func isDirectResolverCall(expr *ast.Node, rCtx *ruleCtx) *ast.Node {
	if expr == nil || !ast.IsCallExpression(expr) {
		return nil
	}
	callee := ast.SkipParentheses(expr.AsCallExpression().Expression)
	if callee == nil || !ast.IsIdentifier(callee) {
		return nil
	}
	if isResolverIdent(callee, rCtx) {
		return expr
	}
	return nil
}

// walkNestedFunctions finds nested function expressions and analyzes each in the
// context of the outer resolver set. Nested functions that rebind a resolver name
// as a parameter are handled differently per mode:
//   - TS mode: pass the same rCtx — symbol identity distinguishes the inner parameter
//     from the outer resolver automatically, so no manual shadowing is needed.
//   - JS mode: build an inner rCtx with parameter-shadowed names removed.
func walkNestedFunctions(node *ast.Node, rCtx *ruleCtx) {
	if node == nil {
		return
	}
	if ast.IsFunctionLike(node) {
		if rCtx.resolverSymbols != nil {
			// TS mode: symbol comparison handles shadowing; reuse same rCtx.
			analyzeFunctionBody(node, rCtx)
		} else {
			// JS mode: strip any names re-declared as parameters.
			innerNames := shadowedResolverNames(rCtx.resolverNames, node)
			innerCtx := &ruleCtx{
				resolverNames: innerNames,
				ctx:           rCtx.ctx,
				reported:      rCtx.reported,
			}
			analyzeFunctionBody(node, innerCtx)
		}
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		walkNestedFunctions(child, rCtx)
		return false
	})
}

// shadowedResolverNames returns resolver names with any parameter-shadowed names removed.
func shadowedResolverNames(names map[string]bool, funcNode *ast.Node) map[string]bool {
	params := funcNode.Parameters()
	if len(params) == 0 {
		return names
	}
	result := make(map[string]bool, len(names))
	for k, v := range names {
		result[k] = v
	}
	for _, param := range params {
		if param == nil || !ast.IsParameterDeclaration(param) {
			continue
		}
		nameNode := param.AsParameterDeclaration().Name()
		if nameNode != nil && ast.IsIdentifier(nameNode) {
			delete(result, nameNode.AsIdentifier().Text)
		}
	}
	return result
}

// analyzeFunctionBody analyzes a FunctionExpression or ArrowFunction body from empty state.
func analyzeFunctionBody(funcNode *ast.Node, rCtx *ruleCtx) {
	if funcNode == nil || len(rCtx.resolverNames) == 0 {
		return
	}
	body := funcNode.Body()
	if body == nil {
		return
	}
	if body.Kind == ast.KindBlock {
		analyzeStmts(body.Statements(), pathState{}, rCtx)
	} else {
		// Arrow function with expression body
		analyzeExprForResolvers(body, pathState{}, rCtx)
	}
}

// analyzeStmts processes statements sequentially, threading state through.
func analyzeStmts(stmts []*ast.Node, state pathState, rCtx *ruleCtx) pathState {
	for _, stmt := range stmts {
		if state.dead {
			break
		}
		state = analyzeStmt(stmt, state, rCtx)
	}
	return state
}

// analyzeStmt processes a single statement under the default empty label.
func analyzeStmt(stmt *ast.Node, state pathState, rCtx *ruleCtx) pathState {
	return analyzeStmtWithLabel(stmt, state, rCtx, "")
}

// analyzeStmtWithLabel processes a single statement, taking into account any enclosing label.
func analyzeStmtWithLabel(stmt *ast.Node, state pathState, rCtx *ruleCtx, label string) pathState {
	if stmt == nil {
		return state
	}

	if label != "" && !isLoopOrSwitch(stmt) {
		blockExitState := pathState{dead: true}
		rCtx.targets = append(rCtx.targets, controlTarget{
			node:      stmt,
			label:     label,
			exitState: &blockExitState,
			isLoop:    false,
		})
		bodyState := analyzeStmtWithLabel(stmt, state, rCtx, "")
		rCtx.targets = rCtx.targets[:len(rCtx.targets)-1]
		return mergeStates(blockExitState, bodyState)
	}

	switch stmt.Kind {
	case ast.KindExpressionStatement:
		expr := stmt.AsExpressionStatement().Expression
		return analyzeExprForResolvers(expr, state, rCtx)

	case ast.KindBlock:
		return analyzeStmts(stmt.Statements(), state, rCtx)

	case ast.KindIfStatement:
		return analyzeIfStmt(stmt, state, rCtx)

	case ast.KindFunctionDeclaration:
		walkNestedFunctions(stmt, rCtx)
		return state

	case ast.KindReturnStatement:
		rs := stmt.AsReturnStatement()
		if rs.Expression != nil {
			state = analyzeExprForResolvers(rs.Expression, state, rCtx)
		}
		return pathState{firstCall: state.firstCall, certain: state.certain, dead: true}

	case ast.KindThrowStatement:
		ts := stmt.AsThrowStatement()
		if ts.Expression != nil {
			state = analyzeExprForResolvers(ts.Expression, state, rCtx)
		}
		return pathState{firstCall: state.firstCall, certain: state.certain, dead: true}

	case ast.KindBreakStatement:
		bs := stmt.AsBreakStatement()
		if len(rCtx.targets) > 0 {
			if bs.Label == nil {
				for i := len(rCtx.targets) - 1; i >= 0; i-- {
					t := rCtx.targets[i]
					if isLoopOrSwitch(t.node) {
						if t.exitState != nil {
							*t.exitState = mergeStates(*t.exitState, state)
						}
						break
					}
				}
			} else {
				labelName := bs.Label.Text()
				for i := len(rCtx.targets) - 1; i >= 0; i-- {
					if rCtx.targets[i].label == labelName {
						if rCtx.targets[i].exitState != nil {
							*rCtx.targets[i].exitState = mergeStates(*rCtx.targets[i].exitState, state)
						}
						break
					}
				}
			}
		}
		return pathState{firstCall: state.firstCall, certain: state.certain, dead: true}

	case ast.KindContinueStatement:
		cs := stmt.AsContinueStatement()
		if len(rCtx.targets) > 0 {
			if cs.Label == nil {
				for i := len(rCtx.targets) - 1; i >= 0; i-- {
					if rCtx.targets[i].isLoop {
						if rCtx.targets[i].exitState != nil {
							*rCtx.targets[i].exitState = mergeStates(*rCtx.targets[i].exitState, state)
						}
						break
					}
				}
			} else {
				labelName := cs.Label.Text()
				for i := len(rCtx.targets) - 1; i >= 0; i-- {
					if rCtx.targets[i].label == labelName && rCtx.targets[i].isLoop {
						if rCtx.targets[i].exitState != nil {
							*rCtx.targets[i].exitState = mergeStates(*rCtx.targets[i].exitState, state)
						}
						break
					}
				}
			}
		}
		return pathState{firstCall: state.firstCall, certain: state.certain, dead: true}

	case ast.KindWhileStatement:
		ws := stmt.AsWhileStatement()
		state = analyzeExprForResolvers(ws.Expression, state, rCtx)
		loopExitState := pathState{dead: true}
		rCtx.targets = append(rCtx.targets, controlTarget{
			node:      stmt,
			label:     label,
			exitState: &loopExitState,
			isLoop:    true,
		})
		bodyState := analyzeStmt(ws.Statement, state, rCtx)
		rCtx.targets = rCtx.targets[:len(rCtx.targets)-1]
		return mergeStates(loopExitState, mergeStates(state, bodyState))

	case ast.KindDoStatement:
		ds := stmt.AsDoStatement()
		loopExitState := pathState{dead: true}
		rCtx.targets = append(rCtx.targets, controlTarget{
			node:      stmt,
			label:     label,
			exitState: &loopExitState,
			isLoop:    true,
		})
		bodyState := analyzeStmt(ds.Statement, state, rCtx)
		bodyState = analyzeExprForResolvers(ds.Expression, bodyState, rCtx)
		rCtx.targets = rCtx.targets[:len(rCtx.targets)-1]
		return mergeStates(loopExitState, bodyState)

	case ast.KindForStatement:
		fs := stmt.AsForStatement()
		if fs.Initializer != nil {
			state = analyzeExprForResolvers(fs.Initializer, state, rCtx)
		}
		if fs.Condition != nil {
			state = analyzeExprForResolvers(fs.Condition, state, rCtx)
		}
		loopExitState := pathState{dead: true}
		rCtx.targets = append(rCtx.targets, controlTarget{
			node:      stmt,
			label:     label,
			exitState: &loopExitState,
			isLoop:    true,
		})
		bodyState := analyzeStmt(fs.Statement, state, rCtx)
		if fs.Incrementor != nil {
			bodyState = analyzeExprForResolvers(fs.Incrementor, bodyState, rCtx)
		}
		rCtx.targets = rCtx.targets[:len(rCtx.targets)-1]
		return mergeStates(loopExitState, mergeStates(state, bodyState))

	case ast.KindForInStatement, ast.KindForOfStatement:
		fio := stmt.AsForInOrOfStatement()
		state = analyzeExprForResolvers(fio.Expression, state, rCtx)
		loopExitState := pathState{dead: true}
		rCtx.targets = append(rCtx.targets, controlTarget{
			node:      stmt,
			label:     label,
			exitState: &loopExitState,
			isLoop:    true,
		})
		bodyState := analyzeStmt(fio.Statement, state, rCtx)
		rCtx.targets = rCtx.targets[:len(rCtx.targets)-1]
		return mergeStates(loopExitState, mergeStates(state, bodyState))

	case ast.KindTryStatement:
		return analyzeTryStmt(stmt, state, rCtx)

	case ast.KindSwitchStatement:
		return analyzeSwitchStmt(stmt, state, rCtx, label)

	case ast.KindLabeledStatement:
		ls := stmt.AsLabeledStatement()
		labelName := ""
		if ls.Label != nil {
			labelName = ls.Label.Text()
		}
		if ls.Statement != nil {
			return analyzeStmtWithLabel(ls.Statement, state, rCtx, labelName)
		}
		return state

	case ast.KindVariableStatement:
		return analyzeVarStmt(stmt, state, rCtx)

	case ast.KindWithStatement:
		ws := stmt.AsWithStatement()
		state = analyzeExprForResolvers(ws.Expression, state, rCtx)
		return analyzeStmt(ws.Statement, state, rCtx)

	default:
		stmt.ForEachChild(func(child *ast.Node) bool {
			walkNestedFunctions(child, rCtx)
			return false
		})
		return state
	}
}

func isLoopOrSwitch(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement,
		ast.KindForInStatement, ast.KindForOfStatement, ast.KindSwitchStatement:
		return true
	}
	return false
}

// analyzeExprForResolvers checks if expr is a direct resolver call and updates state.
// For non-resolver expressions, it walks for nested functions without changing state.
func analyzeExprForResolvers(expr *ast.Node, state pathState, rCtx *ruleCtx) pathState {
	if expr == nil {
		return state
	}
	expr = ast.SkipParentheses(expr)
	if call := isDirectResolverCall(expr, rCtx); call != nil {
		return onResolverCall(call, state, rCtx)
	}
	if ast.IsFunctionLike(expr) {
		walkNestedFunctions(expr, rCtx)
		return state
	}

	switch expr.Kind {
	case ast.KindConditionalExpression:
		ce := expr.AsConditionalExpression()
		state = analyzeExprForResolvers(ce.Condition, state, rCtx)
		thenState := analyzeExprForResolvers(ce.WhenTrue, state, rCtx)
		elseState := analyzeExprForResolvers(ce.WhenFalse, state, rCtx)
		return mergeStates(thenState, elseState)

	case ast.KindBinaryExpression:
		be := expr.AsBinaryExpression()
		switch op := be.OperatorToken.Kind; op {
		case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
			state = analyzeExprForResolvers(be.Left, state, rCtx)
			rightState := analyzeExprForResolvers(be.Right, state, rCtx)
			return mergeStates(rightState, state)
		case ast.KindCommaToken:
			state = analyzeExprForResolvers(be.Left, state, rCtx)
			return analyzeExprForResolvers(be.Right, state, rCtx)
		default:
			state = analyzeExprForResolvers(be.Left, state, rCtx)
			return analyzeExprForResolvers(be.Right, state, rCtx)
		}

	default:
		expr.ForEachChild(func(child *ast.Node) bool {
			state = analyzeExprForResolvers(child, state, rCtx)
			return false
		})
		return state
	}
}

func analyzeIfStmt(stmt *ast.Node, state pathState, rCtx *ruleCtx) pathState {
	ifStmt := stmt.AsIfStatement()
	state = analyzeExprForResolvers(ifStmt.Expression, state, rCtx)

	thenState := analyzeStmt(ifStmt.ThenStatement, state, rCtx)

	var elseState pathState
	if ifStmt.ElseStatement != nil {
		elseState = analyzeStmt(ifStmt.ElseStatement, state, rCtx)
	} else {
		// No else: one path has no resolver effect
		elseState = state
	}

	return mergeStates(thenState, elseState)
}

func analyzeTryStmt(stmt *ast.Node, state pathState, rCtx *ruleCtx) pathState {
	ts := stmt.AsTryStatement()

	// Analyze try block normally
	var tryStmts []*ast.Node
	if ts.TryBlock != nil {
		tryStmts = ts.TryBlock.Statements()
	}
	tryState := analyzeStmts(tryStmts, state, rCtx)

	if ts.CatchClause != nil {
		// Determine what state the catch clause begins with.
		// Catch is entered when an exception is thrown inside try.
		// If the last throwable expression in the try block is the resolver call itself,
		// the catch can only be reached before that resolve call.
		catchEntryState := computeCatchEntryState(tryStmts, state, tryState, rCtx)

		cc := ts.CatchClause.AsCatchClause()
		var catchStmts []*ast.Node
		if cc.Block != nil {
			catchStmts = cc.Block.Statements()
		}
		catchState := analyzeStmts(catchStmts, catchEntryState, rCtx)

		// After try-catch: normal completion of try or catch completion
		tryState = mergeStates(tryState, catchState)
	}

	if ts.FinallyBlock != nil {
		// Finally always runs after try (and catch, if present).
		finallyStmts := ts.FinallyBlock.Statements()
		tryState = analyzeStmts(finallyStmts, tryState, rCtx)
	}

	return tryState
}

// computeCatchEntryState determines the pathState with which catch begins.
// If the last throwable expression in the try block is a resolver call, the catch
// is only reachable before that resolve, so it inherits the initial state.
// Otherwise the catch may run after a resolve, so it inherits the try state.
func computeCatchEntryState(tryStmts []*ast.Node, initialState, tryState pathState, rCtx *ruleCtx) pathState {
	if tryState.firstCall == nil {
		// No resolver was called in the try block
		return initialState
	}

	// Find the last throwable expression in the try block
	lastIsResolver := findLastThrowableIsResolver(tryStmts, rCtx)
	if lastIsResolver {
		// Last throwable is the resolver itself → catch runs before the resolve
		return initialState
	}

	// A throwable expression appears after the resolver → catch may run after resolve
	return pathState{firstCall: tryState.firstCall, certain: false}
}

// findLastThrowableIsResolver scans try statements in order and reports whether
// the last "throwable expression" (Call, PropertyAccess, ElementAccess, New, Yield)
// is a resolver call. This mirrors ESLint's lastThrowableExpression tracking.
func findLastThrowableIsResolver(stmts []*ast.Node, rCtx *ruleCtx) bool {
	var lastIsResolver bool
	for _, stmt := range stmts {
		walkThrowables(stmt, rCtx, func(isResolver bool) {
			lastIsResolver = isResolver
		})
	}
	return lastIsResolver
}

// walkThrowables walks a node (skipping nested functions) and calls onThrowable
// in post-order for every throwable expression encountered. The last call to
// onThrowable corresponds to the last throwable expression in execution order.
func walkThrowables(node *ast.Node, rCtx *ruleCtx, onThrowable func(isResolver bool)) {
	if node == nil {
		return
	}
	// Don't descend into nested functions
	if ast.IsFunctionLike(node) {
		return
	}

	switch node.Kind {
	case ast.KindCallExpression:
		// First recurse (children are evaluated before the call itself)
		node.ForEachChild(func(child *ast.Node) bool {
			walkThrowables(child, rCtx, onThrowable)
			return false
		})
		callee := ast.SkipParentheses(node.AsCallExpression().Expression)
		isResolver := callee != nil && ast.IsIdentifier(callee) && isResolverIdent(callee, rCtx)
		onThrowable(isResolver)

	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindNewExpression,
		ast.KindYieldExpression:
		node.ForEachChild(func(child *ast.Node) bool {
			walkThrowables(child, rCtx, onThrowable)
			return false
		})
		onThrowable(false)

	default:
		node.ForEachChild(func(child *ast.Node) bool {
			walkThrowables(child, rCtx, onThrowable)
			return false
		})
	}
}

func analyzeSwitchStmt(stmt *ast.Node, state pathState, rCtx *ruleCtx, label string) pathState {
	ss := stmt.AsSwitchStatement()
	state = analyzeExprForResolvers(ss.Expression, state, rCtx)

	if ss.CaseBlock == nil {
		return state
	}
	cb := ss.CaseBlock.AsCaseBlock()
	if cb.Clauses == nil || len(cb.Clauses.Nodes) == 0 {
		return state
	}

	hasDefault := false
	for _, clause := range cb.Clauses.Nodes {
		if clause != nil && clause.AsCaseOrDefaultClause().Expression == nil {
			hasDefault = true
			break
		}
	}

	// Create and push the switch exit state onto targets
	switchExitState := pathState{dead: true}
	rCtx.targets = append(rCtx.targets, controlTarget{
		node:      stmt,
		label:     label,
		exitState: &switchExitState,
		isLoop:    false,
	})

	// fallthroughState carries the state falling through from the previous case.
	fallthroughState := pathState{dead: true}

	for _, clause := range cb.Clauses.Nodes {
		if clause == nil {
			continue
		}
		cc := clause.AsCaseOrDefaultClause()
		if cc.Expression != nil {
			state = analyzeExprForResolvers(cc.Expression, state, rCtx)
		}

		caseEntryState := mergeStates(state, fallthroughState)

		caseExitState := caseEntryState
		for _, caseStmt := range cc.Statements.Nodes {
			if caseExitState.dead {
				break
			}
			caseExitState = analyzeStmt(caseStmt, caseExitState, rCtx)
		}

		fallthroughState = caseExitState
	}

	// Pop the switch exit state
	rCtx.targets = rCtx.targets[:len(rCtx.targets)-1]

	finalExitState := mergeStates(switchExitState, fallthroughState)
	if !hasDefault {
		finalExitState = mergeStates(finalExitState, state)
	}

	return finalExitState
}

func analyzeVarStmt(stmt *ast.Node, state pathState, rCtx *ruleCtx) pathState {
	vs := stmt.AsVariableStatement()
	if vs.DeclarationList == nil {
		return state
	}
	dl := vs.DeclarationList.AsVariableDeclarationList()
	if dl.Declarations == nil {
		return state
	}
	for _, decl := range dl.Declarations.Nodes {
		if decl == nil || decl.Kind != ast.KindVariableDeclaration {
			continue
		}
		vd := decl.AsVariableDeclaration()
		if vd.Initializer != nil {
			state = analyzeExprForResolvers(vd.Initializer, state, rCtx)
		}
	}
	return state
}

// isPromiseExecutorFunction checks whether node is `new Promise(executor)` and
// returns the executor function if so.
func isPromiseExecutorFunction(node *ast.Node) *ast.Node {
	ne := node.AsNewExpression()
	if ne == nil {
		return nil
	}
	callee := ast.SkipParentheses(ne.Expression)
	if callee == nil || !ast.IsIdentifier(callee) || callee.AsIdentifier().Text != "Promise" {
		return nil
	}
	args := node.Arguments()
	if len(args) != 1 {
		return nil
	}
	arg := ast.SkipParentheses(args[0])
	if arg == nil {
		return nil
	}
	if arg.Kind == ast.KindFunctionExpression || arg.Kind == ast.KindArrowFunction {
		return arg
	}
	return nil
}

var NoMultipleResolvedRule = rule.Rule{
	Name: "promise/no-multiple-resolved",
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		reported := make(map[*ast.Node]bool)

		return rule.RuleListeners{
			ast.KindNewExpression: func(node *ast.Node) {
				executor := isPromiseExecutorFunction(node)
				if executor == nil {
					return
				}

				// Collect resolver parameter names (always) and symbols (TS mode only).
				resolverNames := make(map[string]bool)
				var resolverSymbols map[*ast.Symbol]bool
				if ctx.TypeChecker != nil {
					resolverSymbols = make(map[*ast.Symbol]bool)
				}
				for _, param := range executor.Parameters() {
					if param == nil || !ast.IsParameterDeclaration(param) {
						continue
					}
					// Skip TypeScript's implicit `this` parameter
					if ast.IsThisParameter(param) {
						continue
					}
					nameNode := param.AsParameterDeclaration().Name()
					if nameNode != nil && ast.IsIdentifier(nameNode) {
						resolverNames[nameNode.AsIdentifier().Text] = true
						if resolverSymbols != nil {
							if sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode); sym != nil {
								resolverSymbols[sym] = true
							}
						}
					}
				}
				if len(resolverNames) == 0 {
					return
				}

				rCtx := &ruleCtx{
					resolverNames:   resolverNames,
					resolverSymbols: resolverSymbols,
					ctx:             ctx,
					reported:        reported,
				}
				analyzeFunctionBody(executor, rCtx)
			},
		}
	},
}
