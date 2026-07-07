package strict_void_return

import (
	"encoding/json"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

func buildAsyncFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "asyncFunc",
		Description: "Async function used in a context where a void function is expected.",
	}
}

func buildNonVoidFuncMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nonVoidFunc",
		Description: "Value-returning function used in a context where a void function is expected.",
	}
}

func buildNonVoidReturnMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "nonVoidReturn",
		Description: "Value returned in a context where a void return is expected.",
	}
}

type StrictVoidReturnOptions struct {
	AllowReturnAny *bool `json:"allowReturnAny,omitempty"`
}

var StrictVoidReturnRule = rule.CreateRule(rule.Rule{
	Name:             "strict-void-return",
	RequiresTypeInfo: true,
	Run: func(ctx rule.RuleContext, _options []any) rule.RuleListeners {
		options := rule.UnwrapOptions(_options)
		opts, ok := options.(StrictVoidReturnOptions)
		if !ok {
			opts = StrictVoidReturnOptions{}
			if optsMap := utils.GetOptionsMap(options); optsMap != nil {
				if optsJSON, err := json.Marshal(optsMap); err == nil {
					_ = json.Unmarshal(optsJSON, &opts)
				}
			}
		}
		if opts.AllowReturnAny == nil {
			opts.AllowReturnAny = utils.Ref(false)
		}

		allowedReturnTypeFlags := checker.TypeFlagsVoid | checker.TypeFlagsNever | checker.TypeFlagsUndefined
		if *opts.AllowReturnAny {
			allowedReturnTypeFlags |= checker.TypeFlagsAny
		}

		isVoid := func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsVoid)
		}

		// nullishOrAny: VoidLike | Undefined | Null | Any | Never (mirrors upstream).
		isNullishOrAny := func(t *checker.Type) bool {
			return utils.IsTypeFlagSet(t, checker.TypeFlagsVoidLike|checker.TypeFlagsUndefined|checker.TypeFlagsNull|checker.TypeFlagsAny|checker.TypeFlagsNever)
		}

		isVoidReturningFunctionType := func(t *checker.Type) bool {
			var returnTypes []*checker.Type
			for _, sig := range utils.CollectAllCallSignatures(ctx.TypeChecker, t) {
				ret := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig)
				returnTypes = append(returnTypes, utils.UnionTypeParts(ret)...)
			}
			if len(returnTypes) == 0 {
				return false
			}
			for _, rt := range returnTypes {
				if !isVoid(rt) {
					return false
				}
			}
			return true
		}

		// walkStatements yields every non-block, non-control-flow statement in a
		// function body, descending into nested blocks / switch cases / loops /
		// try-catch-finally. Mirrors upstream walkStatements.ts so we visit every
		// `return` reachable from the function head.
		var walkStatements func(stmts []*ast.Node, out *[]*ast.Node)
		walkStatements = func(stmts []*ast.Node, out *[]*ast.Node) {
			for _, s := range stmts {
				switch s.Kind {
				case ast.KindBlock:
					walkStatements(s.AsBlock().Statements.Nodes, out)
				case ast.KindSwitchStatement:
					for _, clause := range s.AsSwitchStatement().CaseBlock.AsCaseBlock().Clauses.Nodes {
						walkStatements(clause.AsCaseOrDefaultClause().Statements.Nodes, out)
					}
				case ast.KindIfStatement:
					ifs := s.AsIfStatement()
					walkStatements([]*ast.Node{ifs.ThenStatement}, out)
					if ifs.ElseStatement != nil {
						walkStatements([]*ast.Node{ifs.ElseStatement}, out)
					}
				case ast.KindWhileStatement, ast.KindDoStatement,
					ast.KindForStatement, ast.KindForInStatement,
					ast.KindForOfStatement, ast.KindWithStatement,
					ast.KindLabeledStatement:
					if body := s.Statement(); body != nil {
						walkStatements([]*ast.Node{body}, out)
					}
				case ast.KindTryStatement:
					try := s.AsTryStatement()
					walkStatements([]*ast.Node{try.TryBlock}, out)
					if try.CatchClause != nil {
						walkStatements([]*ast.Node{try.CatchClause.AsCatchClause().Block}, out)
					}
					if try.FinallyBlock != nil {
						walkStatements([]*ast.Node{try.FinallyBlock}, out)
					}
				default:
					*out = append(*out, s)
				}
			}
		}

		// reportIfNonVoidFunction reports the appropriate diagnostic on funcNode
		// when its actual return type isn't void/never/undefined (or any, when
		// allowReturnAny is on). Mirrors upstream's switch but extends "function
		// literal" to cover tsgo's MethodDeclaration / GetAccessor / SetAccessor
		// / Constructor — upstream maps these onto a synthetic FunctionExpression
		// via MethodDefinition.value, while tsgo represents the method as the
		// function-like itself.
		reportIfNonVoidFunction := func(funcNode *ast.Node) {
			// tsgo preserves ParenthesizedExpression nodes that ESTree flattens.
			// Strip parens so the IsArrowFunction / IsFunctionExpression branch
			// detection (and downstream report-on-body / -head logic) sees the
			// real function-like node — e.g. `((() => 1))` must report at `1`,
			// not at the outer `(`.
			funcNode = ast.SkipParentheses(funcNode)
			actualType := checker.Checker_getApparentType(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(funcNode))

			sigs := utils.CollectAllCallSignatures(ctx.TypeChecker, actualType)
			if len(sigs) == 0 {
				return
			}
			allReturnsAllowed := true
			for _, sig := range sigs {
				ret := checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, sig)
				for _, part := range utils.UnionTypeParts(ret) {
					if !utils.IsTypeFlagSet(part, allowedReturnTypeFlags) {
						allReturnsAllowed = false
						break
					}
				}
				if !allReturnsAllowed {
					break
				}
			}
			if allReturnsAllowed {
				return
			}

			isArrow := ast.IsArrowFunction(funcNode)
			isFuncExpr := ast.IsFunctionExpression(funcNode)
			isMethodLike := funcNode.Kind == ast.KindMethodDeclaration ||
				funcNode.Kind == ast.KindGetAccessor ||
				funcNode.Kind == ast.KindSetAccessor ||
				funcNode.Kind == ast.KindConstructor
			if !isArrow && !isFuncExpr && !isMethodLike {
				ctx.ReportNode(funcNode, buildNonVoidFuncMessage())
				return
			}

			flags := ast.GetFunctionFlags(funcNode)
			if flags&ast.FunctionFlagsGenerator != 0 {
				ctx.ReportRange(utils.GetFunctionHeadLoc(ctx.SourceFile, funcNode), buildNonVoidFuncMessage())
				return
			}
			if flags&ast.FunctionFlagsAsync != 0 {
				ctx.ReportRange(utils.GetFunctionHeadLoc(ctx.SourceFile, funcNode), buildAsyncFuncMessage())
				return
			}

			body := funcNode.Body()
			if isArrow && body != nil && body.Kind != ast.KindBlock {
				// Concise arrow body: strip parens (tsgo preserves; ESTree flattens).
				ctx.ReportNode(ast.SkipParentheses(body), buildNonVoidReturnMessage())
				return
			}

			// Explicit non-void return-type annotation short-circuits the walk.
			if typeAnnotation := funcNode.Type(); typeAnnotation != nil && typeAnnotation.Kind != ast.KindVoidKeyword {
				ctx.ReportNode(typeAnnotation, buildNonVoidFuncMessage())
				return
			}

			if body == nil || body.Kind != ast.KindBlock {
				return
			}

			var stmts []*ast.Node
			walkStatements(body.AsBlock().Statements.Nodes, &stmts)
			for _, stmt := range stmts {
				if stmt.Kind != ast.KindReturnStatement {
					continue
				}
				ret := stmt.AsReturnStatement()
				if ret.Expression == nil {
					continue
				}
				retType := ctx.TypeChecker.GetTypeAtLocation(ret.Expression)
				if utils.IsTypeFlagSet(retType, allowedReturnTypeFlags) {
					continue
				}
				returnKeyword := scanner.GetRangeOfTokenAtPosition(ctx.SourceFile, stmt.Pos())
				ctx.ReportRange(returnKeyword, buildNonVoidReturnMessage())
			}
		}

		// checkExpressionNode is the generic entrypoint used by every non-call
		// context. Returns true if the contextual type was a void-returning
		// function (so callers can suppress fallback checks).
		checkExpressionNode := func(node *ast.Node) bool {
			expected := checker.Checker_getContextualType(ctx.TypeChecker, node, checker.ContextFlagsNone)
			if expected != nil && isVoidReturningFunctionType(expected) {
				reportIfNonVoidFunction(node)
				return true
			}
			return false
		}

		// checkFunctionCallNode mirrors upstream's overload-aware argument check.
		// We iterate every call/construct signature on the callee, build the union
		// of expected return types per argument index, then decide per-argument
		// whether the slot is void-only (after collapsing nullish/any/type-params
		// down to void).
		checkFunctionCallNode := func(callNode *ast.Node) {
			calleeExpr := callNode.Expression()
			funcType := ctx.TypeChecker.GetTypeAtLocation(calleeExpr)

			isCall := ast.IsCallExpression(callNode)
			var allSignatures []*checker.Signature
			for _, t := range utils.UnionTypeParts(funcType) {
				if isCall {
					allSignatures = append(allSignatures, utils.GetCallSignatures(ctx.TypeChecker, t)...)
				} else {
					allSignatures = append(allSignatures, utils.GetConstructSignatures(ctx.TypeChecker, t)...)
				}
			}

			args := callNode.Arguments()
			hasSingleSignature := len(allSignatures) == 1

			for argIdx, argNode := range args {
				if argNode.Kind == ast.KindSpreadElement {
					continue
				}

				// Collect every signature's argIdx-th param call-signature return
				// type so we can detect `(() => void) | (() => any)`-style overloads
				// where contextual type alone would mislead us. Upstream notes that
				// `checker.getResolvedSignature` can't be used here because it
				// prefers an early `() => void` over a later `() => Promise<void>`.
				var argExpectedReturnTypes []*checker.Type
				for _, sig := range allSignatures {
					params := checker.Signature_parameters(sig)
					if argIdx >= len(params) {
						continue
					}
					paramType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(params[argIdx], calleeExpr)
					for _, ps := range utils.CollectAllCallSignatures(ctx.TypeChecker, paramType) {
						argExpectedReturnTypes = append(argExpectedReturnTypes, checker.Checker_getReturnTypeOfSignature(ctx.TypeChecker, ps))
					}
				}

				allSignaturesReturnVoid := true
				for _, t := range argExpectedReturnTypes {
					if isVoid(t) || isNullishOrAny(t) || utils.IsTypeFlagSet(t, checker.TypeFlagsTypeParameter) {
						continue
					}
					allSignaturesReturnVoid = false
					break
				}

				// Only trust contextual type when there is one signature, or every
				// signature returns void for this slot — `getContextualType` picks
				// the first overload's return type even if a later overload matches.
				if hasSingleSignature || allSignaturesReturnVoid {
					if checkExpressionNode(argNode) {
						continue
					}
				}

				// Fallback: treat the slot as void if at least one signature
				// returns void and every other signature returns nullish/any.
				hasVoid := false
				allRestNullishOrAny := true
				for _, t := range argExpectedReturnTypes {
					if isVoid(t) {
						hasVoid = true
					}
					if !isNullishOrAny(t) {
						allRestNullishOrAny = false
					}
				}
				if hasVoid && allRestNullishOrAny {
					reportIfNonVoidFunction(argNode)
				}
			}
		}

		// checkObjectPropertyNode handles each property of an object literal.
		// Method shorthand (`{ cb() {} }`) requires looking up the property
		// symbol on the object's contextual type, since the method itself has
		// no contextual type wired through. Getters/setters in object literals
		// fall through to the same lookup path so a `get cb() { return () => 1 }`
		// inside `{ cb: () => void }` is reported just like in upstream.
		checkObjectPropertyNode := func(propNode *ast.Node) {
			switch propNode.Kind {
			case ast.KindPropertyAssignment:
				if init := propNode.AsPropertyAssignment().Initializer; init != nil {
					checkExpressionNode(init)
				}
			case ast.KindShorthandPropertyAssignment:
				// `{ cb }` — check the identifier against the contextual prop type.
				checkExpressionNode(propNode.Name())
			case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
				name := propNode.Name()
				if name == nil || name.Kind == ast.KindComputedPropertyName {
					return
				}
				objType := checker.Checker_getContextualType(ctx.TypeChecker, propNode.Parent, checker.ContextFlagsNone)
				if objType == nil {
					return
				}
				propSymbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, objType, name.Text())
				if propSymbol == nil {
					return
				}
				propExpectedType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(propSymbol, propNode)
				if isVoidReturningFunctionType(propExpectedType) {
					reportIfNonVoidFunction(propNode)
				}
			}
		}

		// forEachBaseMemberType mirrors upstream's getBaseTypesOfClassMember:
		// for a class member, yield each (base class / implemented interface)'s
		// corresponding member type. Walks heritage clauses (both extends and
		// implements) so behaviour matches upstream regardless of whether the
		// override comes through extension or interface implementation.
		forEachBaseMemberType := func(memberNode *ast.Node, fn func(baseMemberType *checker.Type) bool) {
			name := memberNode.Name()
			if name == nil {
				return
			}
			memberSymbol := ctx.TypeChecker.GetSymbolAtLocation(name)
			if memberSymbol == nil {
				return
			}
			parent := memberNode.Parent
			if parent == nil {
				return
			}
			heritageClauses := utils.GetHeritageClauses(parent)
			if heritageClauses == nil {
				return
			}
			memberName := memberSymbol.Name
			for _, clause := range heritageClauses.Nodes {
				for _, baseTypeNode := range clause.AsHeritageClause().Types.Nodes {
					baseType := ctx.TypeChecker.GetTypeAtLocation(baseTypeNode)
					baseMemberSymbol := checker.Checker_getPropertyOfType(ctx.TypeChecker, baseType, memberName)
					if baseMemberSymbol == nil {
						continue
					}
					baseMemberType := ctx.TypeChecker.GetTypeOfSymbolAtLocation(baseMemberSymbol, memberNode)
					if fn(baseMemberType) {
						return
					}
				}
			}
		}

		// checkClassMember handles methods, getters, setters: only reports when
		// the member overrides a void-returning member of an extended class or
		// implemented interface. The standalone contextual-type case is handled
		// by checkClassPropertyNode for fields.
		checkClassMember := func(memberNode *ast.Node) {
			if memberNode.Body() == nil {
				return
			}
			forEachBaseMemberType(memberNode, func(baseMemberType *checker.Type) bool {
				if isVoidReturningFunctionType(baseMemberType) {
					reportIfNonVoidFunction(memberNode)
					return true // at most one error
				}
				return false
			})
		}

		// classMemberListener wraps a class-member callback with a parent-kind
		// gate so that method shorthand / accessors in object literals get
		// routed through checkObjectPropertyNode instead.
		classMemberListener := func(fn func(*ast.Node)) func(*ast.Node) {
			return func(node *ast.Node) {
				if parent := node.Parent; parent == nil || (!ast.IsClassDeclaration(parent) && !ast.IsClassExpression(parent)) {
					return
				}
				fn(node)
			}
		}

		// checkClassPropertyNode handles class fields: compare both against the
		// base class/interface member and the field's contextual type.
		checkClassPropertyNode := func(propNode *ast.Node) {
			prop := propNode.AsPropertyDeclaration()
			if prop.Initializer == nil {
				return
			}
			reported := false
			forEachBaseMemberType(propNode, func(baseMemberType *checker.Type) bool {
				if isVoidReturningFunctionType(baseMemberType) {
					reportIfNonVoidFunction(prop.Initializer)
					reported = true
					return true
				}
				return false
			})
			if reported {
				return
			}
			checkExpressionNode(prop.Initializer)
		}

		return rule.RuleListeners{
			ast.KindArrayLiteralExpression: func(node *ast.Node) {
				for _, el := range node.AsArrayLiteralExpression().Elements.Nodes {
					if el == nil || el.Kind == ast.KindOmittedExpression || el.Kind == ast.KindSpreadElement {
						continue
					}
					checkExpressionNode(el)
				}
			},
			ast.KindArrowFunction: func(node *ast.Node) {
				body := node.Body()
				if body != nil && body.Kind != ast.KindBlock {
					checkExpressionNode(body)
				}
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
					return
				}
				// Mirrors upstream's AssignmentExpression listener: defer the
				// expected-type lookup to TypeChecker.getContextualType on the
				// RHS. For `+=` / `-=` etc. the contextual type is numeric (not
				// a function), so isVoidReturningFunctionType naturally bails.
				checkExpressionNode(bin.Right)
			},
			ast.KindCallExpression: checkFunctionCallNode,
			ast.KindNewExpression:  checkFunctionCallNode,
			ast.KindJsxAttribute: func(node *ast.Node) {
				init := node.AsJsxAttribute().Initializer
				if init == nil || init.Kind != ast.KindJsxExpression {
					return
				}
				expr := init.AsJsxExpression().Expression
				if expr == nil {
					return
				}
				// Query the JsxExpression wrapper for its contextual type, then
				// dispatch reportIfNonVoidFunction on the inner expression —
				// tsgo's getContextualType on the inner expression doesn't always
				// resolve through the JSX attribute slot.
				expected := checker.Checker_getContextualType(ctx.TypeChecker, init, checker.ContextFlagsNone)
				if expected != nil && isVoidReturningFunctionType(expected) {
					reportIfNonVoidFunction(expr)
				}
			},
			// Class methods / getters / setters / constructor — same handler,
			// the wrapper gates on parent kind so object-literal members are
			// routed through checkObjectPropertyNode instead. Upstream's
			// MethodDefinition listener covers constructor too; in practice
			// it's always a no-op because a base type's constructor is not
			// exposed via getPropertyOfType, but we register it for matrix
			// parity so future heritage-walk changes can't silently drift.
			ast.KindMethodDeclaration: classMemberListener(checkClassMember),
			ast.KindGetAccessor:       classMemberListener(checkClassMember),
			ast.KindSetAccessor:       classMemberListener(checkClassMember),
			ast.KindConstructor:       classMemberListener(checkClassMember),
			ast.KindObjectLiteralExpression: func(node *ast.Node) {
				for _, prop := range node.AsObjectLiteralExpression().Properties.Nodes {
					if prop.Kind == ast.KindSpreadAssignment {
						continue
					}
					checkObjectPropertyNode(prop)
				}
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkClassPropertyNode(node)
			},
			ast.KindReturnStatement: func(node *ast.Node) {
				if expr := node.AsReturnStatement().Expression; expr != nil {
					checkExpressionNode(expr)
				}
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				if init := node.AsVariableDeclaration().Initializer; init != nil {
					checkExpressionNode(init)
				}
			},
		}
	},
})
