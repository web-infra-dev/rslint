// Package valid_expect holds the framework-neutral async-assertion machinery
// shared by jest/valid-expect and rstest/valid-expect: walking out of a chained
// .then/.catch/.finally, spotting a Promise.all([...]) wrapper, deciding
// whether the result is awaited or returned, and generating safe async/await
// autofixes.
//
// Only the parts whose correctness does not depend on the framework live here.
// The rules keep their own bodies: how a broken chain's reason is derived, what
// counts as an excess expect argument and which matcher shapes exist all differ
// between jest and rstest, and merging those would mean framework branches
// inside a shared body.
package valid_expect

import (
	"fmt"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type consumeFlags uint8

const (
	consumeViaReturn  consumeFlags = 1 << iota // bare `return x` (not `return await x`)
	consumeNonFinally                          // chained via `.then`/`.catch` or `Promise.allSettled([x])`
)

type consumeSite struct {
	stmtIdx int
	flags   consumeFlags
}

type blockAwaitSummary struct {
	stmtIndex     map[*ast.Node]int
	nextEarlyExit []int
	bySym         map[*ast.Symbol][]consumeSite
	byName        map[string][]consumeSite
}

type arrayPureSuffixInfo struct {
	elemIndex           map[*ast.Node]int
	minPureIdx          int
	allowStandaloneBare bool
}

type syncFnReturnSummary struct {
	count int
	only  *ast.Node
}

// AsyncAssertionResolver caches per-file lookup state for resolving whether an
// async assertion is awaited/returned and whether an autofix is safe.
type AsyncAssertionResolver struct {
	sourceFile                     *ast.SourceFile
	typeChecker                    *checker.Checker
	refs                           *rule.RefStore
	staticStrings                  *internalUtils.StaticStringEvaluator
	disallowBuiltinToResolveReject bool
	checkedFileExtend              bool
	fileHasExtend                  bool
	blockAwaitCache                map[*ast.Block]*blockAwaitSummary
	arrayPureSuffixCache           map[*ast.Node]*arrayPureSuffixInfo
	syncFnReturnCache              map[*ast.Node]syncFnReturnSummary
}

// NewAsyncAssertionResolver creates a per-file resolver backed by ctx.
func NewAsyncAssertionResolver(ctx rule.RuleContext) *AsyncAssertionResolver {
	return &AsyncAssertionResolver{
		sourceFile:  ctx.SourceFile,
		typeChecker: ctx.TypeChecker,
		refs:        ctx.Refs,
	}
}

// DisableBuiltinToResolveReject tells the resolver that bare `toResolve()` /
// `toReject()` matchers are not built-in in this framework (e.g. Rstest).
func (r *AsyncAssertionResolver) DisableBuiltinToResolveReject() {
	if r != nil {
		r.disallowBuiltinToResolveReject = true
	}
}

func (r *AsyncAssertionResolver) staticStringEvaluator() *internalUtils.StaticStringEvaluator {
	if r == nil || r.sourceFile == nil {
		return nil
	}
	if r.staticStrings == nil {
		var refResolver internalUtils.StaticReferenceResolver
		if r.refs != nil {
			refResolver = r.refs
		}
		r.staticStrings = internalUtils.NewStaticStringEvaluatorWithCallEscapes(
			r.typeChecker,
			r.sourceFile,
			refResolver,
		)
	}
	return r.staticStrings
}

func (r *AsyncAssertionResolver) resolveSymbol(ident *ast.Node) *ast.Symbol {
	if r == nil || ident == nil || ident.Kind != ast.KindIdentifier {
		return nil
	}
	if r.refs != nil {
		if sym := r.refs.Resolve(ident); sym != nil {
			return sym
		}
	}
	if r.typeChecker != nil {
		return r.typeChecker.GetSymbolAtLocation(ident)
	}
	return nil
}

func isSameFileRuntimeValueDecl(decl *ast.Node, sf *ast.SourceFile) bool {
	if decl == nil || decl.Kind == ast.KindSourceFile {
		return false
	}
	if sf != nil && ast.GetSourceFileOfNode(decl) != sf {
		return false
	}
	if decl.Flags&ast.NodeFlagsAmbient != 0 || ast.HasSyntacticModifier(decl, ast.ModifierFlagsAmbient) {
		return false
	}
	switch decl.Kind {
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindTypeParameter:
		return false
	case ast.KindImportSpecifier, ast.KindNamespaceImport, ast.KindImportClause, ast.KindImportDeclaration:
		return !ast.IsTypeOnlyImportDeclaration(decl)
	case ast.KindImportEqualsDeclaration:
		return !decl.AsImportEqualsDeclaration().IsTypeOnly
	case ast.KindModuleDeclaration:
		return ast.GetModuleInstanceState(decl) == ast.ModuleInstanceStateInstantiated
	case ast.KindVariableDeclaration:
		if decl.Parent != nil && decl.Parent.Parent != nil &&
			ast.HasSyntacticModifier(decl.Parent.Parent, ast.ModifierFlagsAmbient) {
			return false
		}
	}
	return true
}

func (r *AsyncAssertionResolver) hasSameFileRuntimeValueDecl(sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	var sf *ast.SourceFile
	if r != nil {
		sf = r.sourceFile
	}
	for _, decl := range sym.Declarations {
		if isSameFileRuntimeValueDecl(decl, sf) {
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) singleSameFileValueDecl(sym *ast.Symbol) *ast.Node {
	if sym == nil {
		return nil
	}
	var sf *ast.SourceFile
	if r != nil {
		sf = r.sourceFile
	}
	var found *ast.Node
	for _, decl := range sym.Declarations {
		if !isSameFileRuntimeValueDecl(decl, sf) {
			continue
		}
		if found != nil {
			return nil
		}
		found = decl
	}
	return found
}

func (r *AsyncAssertionResolver) isLocallyShadowed(node *ast.Node, name string) bool {
	if node == nil || name == "" {
		return false
	}
	if r != nil && r.refs != nil && node.Kind == ast.KindIdentifier {
		if r.refs.IsGlobalReference(node) {
			return false
		}
		// RefStore.Resolve searches value/alias spaces only; when it returns
		// nil (e.g. script-level `interface Promise` or non-instantiated
		// `namespace Promise`), there is no runtime value binding in this file.
		sym := r.refs.Resolve(node)
		if sym == nil {
			return false
		}
		return r.hasSameFileRuntimeValueDecl(sym)
	}
	return internalUtils.IsShadowed(node, name)
}

func (r *AsyncAssertionResolver) accessStaticNameThroughAssertions(node *ast.Node) (string, bool) {
	if node == nil || !ast.IsAccessExpression(node) {
		return "", false
	}
	if name, ok := internalUtils.AccessExpressionStaticName(node); ok {
		return name, true
	}
	if node.Kind == ast.KindElementAccessExpression {
		arg := internalUtils.SkipAssertionsAndParens(node.AsElementAccessExpression().ArgumentExpression)
		if arg == nil {
			return "", false
		}
		if name, ok := internalUtils.GetStaticStringLiteralValue(arg); ok {
			return name, true
		}
		if ev := r.staticStringEvaluator(); ev != nil {
			return ev.EvalToString(arg)
		}
	}
	return "", false
}

func (r *AsyncAssertionResolver) isGlobalThisObject(node *ast.Node) bool {
	u := internalUtils.SkipAssertionsAndParens(node)
	return u != nil && u.Kind == ast.KindIdentifier &&
		u.AsIdentifier().Text == "globalThis" &&
		!r.isLocallyShadowed(u, "globalThis")
}

func (r *AsyncAssertionResolver) isGlobalPromiseObject(node *ast.Node) bool {
	return r.isGlobalPromiseObjectDepth(node, 0)
}

func (r *AsyncAssertionResolver) isGlobalPromiseObjectDepth(node *ast.Node, depth int) bool {
	if depth > 2 {
		return false
	}
	u := internalUtils.SkipAssertionsAndParens(node)
	if u == nil {
		return false
	}
	if u.Kind == ast.KindIdentifier {
		if u.AsIdentifier().Text == "Promise" && !r.isLocallyShadowed(u, "Promise") {
			return true
		}
		if sym := r.resolveSymbol(u); sym != nil {
			if decl := r.singleSameFileValueDecl(sym); decl != nil {
				switch decl.Kind {
				case ast.KindVariableDeclaration:
					if ast.IsVarConst(decl) && !ast.IsVarAwaitUsing(decl) {
						init := decl.AsVariableDeclaration().Initializer
						if init != nil && !nodeContains(init, u) {
							return r.isGlobalPromiseObjectDepth(init, depth+1)
						}
					}
				case ast.KindBindingElement:
					be := decl.AsBindingElement()
					if be != nil && be.DotDotDotToken == nil &&
						decl.Parent != nil && decl.Parent.Kind == ast.KindObjectBindingPattern {
						// Direct destructuring only (`const { Promise: P } = globalThis`),
						// never nested patterns like `const { fake: { Promise: P } } = globalThis`.
						vd := decl.Parent.Parent
						if vd != nil && vd.Kind == ast.KindVariableDeclaration &&
							ast.IsVarConst(vd) && !ast.IsVarAwaitUsing(vd) &&
							r.isGlobalThisObject(vd.AsVariableDeclaration().Initializer) {
							propName := ""
							if be.PropertyName != nil {
								propName, _ = internalUtils.GetStaticPropertyName(be.PropertyName)
							} else if be.Name() != nil && be.Name().Kind == ast.KindIdentifier {
								propName = be.Name().AsIdentifier().Text
							}
							return propName == "Promise"
						}
					}
				}
			}
		}
		return false
	}
	if ast.IsAccessExpression(u) {
		name, ok := r.accessStaticNameThroughAssertions(u)
		return ok && name == "Promise" && r.isGlobalThisObject(internalUtils.AccessExpressionObject(u))
	}
	return false
}

func nodeContains(ancestor, target *ast.Node) bool {
	return ancestor != nil && target != nil && target.Pos() >= ancestor.Pos() && target.End() <= ancestor.End()
}

func (r *AsyncAssertionResolver) promiseMethodName(call *ast.Node) (string, bool) {
	if call == nil || call.Kind != ast.KindCallExpression {
		return "", false
	}
	callee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return "", false
	}
	if !r.isGlobalPromiseObject(internalUtils.AccessExpressionObject(callee)) {
		return "", false
	}
	return r.accessStaticNameThroughAssertions(callee)
}

func isNonPromiseWrapperMethod(method string) bool {
	switch method {
	case "try", "withResolvers",
		"bind", "call", "apply",
		"toString", "valueOf", "toLocaleString",
		"hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable",
		"constructor", "prototype":
		return true
	}
	return false
}

func isNonCombinatorPromiseMethod(method string) bool {
	return method == "resolve" || method == "reject" || isNonPromiseWrapperMethod(method)
}

func (r *AsyncAssertionResolver) promiseChainMethodName(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindCallExpression {
		return "", false
	}
	callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return "", false
	}
	name, ok := r.accessStaticNameThroughAssertions(callee)
	if !ok || (name != "then" && name != "catch" && name != "finally") {
		return "", false
	}
	return name, true
}

func (r *AsyncAssertionResolver) getParentIfPromiseChained(node *ast.Node, onlyFinally bool) *ast.Node {
	for node != nil {
		outer := internalUtils.OutermostTransparentExpression(node)
		if outer == nil || outer.Parent == nil || !ast.IsAccessExpression(outer.Parent) || outer.Parent.Expression() != outer {
			return node
		}
		member := outer.Parent
		name, ok := r.accessStaticNameThroughAssertions(member)
		if !ok {
			return node
		}
		if onlyFinally {
			if name != "finally" {
				return node
			}
		} else if name != "then" && name != "catch" && name != "finally" {
			return node
		}
		memberOuter := internalUtils.OutermostTransparentExpression(member)
		if memberOuter == nil || memberOuter.Parent == nil ||
			memberOuter.Parent.Kind != ast.KindCallExpression ||
			memberOuter.Parent.AsCallExpression().Expression != memberOuter {
			return node
		}
		node = memberOuter.Parent
	}
	return node
}

func isAcceptableReturnNode(child *ast.Node, parent *ast.Node, allowReturn bool) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindAwaitExpression:
		return parent.AsAwaitExpression().Expression == child
	case ast.KindReturnStatement:
		return allowReturn && parent.AsReturnStatement().Expression == child
	case ast.KindArrowFunction:
		return parent.AsArrowFunction().Body == child
	case ast.KindConditionalExpression:
		cond := parent.AsConditionalExpression()
		if cond == nil || (cond.WhenTrue != child && cond.WhenFalse != child) {
			return false
		}
		outer := internalUtils.OutermostTransparentExpression(parent)
		return isAcceptableReturnNode(outer, outer.Parent, allowReturn)
	default:
		return false
	}
}

func isDirectFirstCallArg(call *ast.Node, arg *ast.Node) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	args := call.AsCallExpression().Arguments
	return args != nil && len(args.Nodes) > 0 && args.Nodes[0] == arg
}

// findPromiseWrapperCall reports whether `node` is directly wrapped in a
// `Promise.<combinator>` call (e.g. `Promise.all([node])`, `Promise.race([node])`,
// `Promise.allSettled([node])`, `Promise.any([node])`, `Promise.resolve(node)`,
// or `Promise.reject(node)`). `Promise.resolve([node])`, `Promise.reject([node])`,
// `Promise.try(...)`, and non-combinator methods like `Promise.withResolvers(...)`
// do not await array elements or arguments and are rejected.
func (r *AsyncAssertionResolver) findPromiseWrapperCall(node *ast.Node) (wrapper *ast.Node, insideArray bool, isReject bool) {
	outer := internalUtils.OutermostTransparentExpression(node)
	if outer == nil || outer.Parent == nil {
		return nil, false, false
	}
	if outer.Parent.Kind == ast.KindArrayLiteralExpression {
		arrOuter := internalUtils.OutermostTransparentExpression(outer.Parent)
		if arrOuter != nil && arrOuter.Parent != nil &&
			isDirectFirstCallArg(arrOuter.Parent, arrOuter) {
			call := arrOuter.Parent
			if method, ok := r.promiseMethodName(call); ok && !isNonCombinatorPromiseMethod(method) {
				return call, true, false
			}
		}
		return nil, true, false
	}
	if isDirectFirstCallArg(outer.Parent, outer) {
		call := outer.Parent
		if method, ok := r.promiseMethodName(call); ok {
			switch method {
			case "all", "allSettled", "race", "any":
				return nil, false, false
			case "reject":
				return call, false, true
			default:
				if isNonPromiseWrapperMethod(method) {
					return nil, false, false
				}
				return call, false, false
			}
		}
	}
	return nil, false, false
}

func (r *AsyncAssertionResolver) getBlockAwaitSummary(block *ast.Block) *blockAwaitSummary {
	if r != nil && r.blockAwaitCache != nil {
		if cached, ok := r.blockAwaitCache[block]; ok {
			return cached
		}
	}
	stmts := block.Statements.Nodes
	n := len(stmts)
	summary := &blockAwaitSummary{
		stmtIndex:     make(map[*ast.Node]int, n),
		nextEarlyExit: make([]int, n+1),
		bySym:         map[*ast.Symbol][]consumeSite{},
		byName:        map[string][]consumeSite{},
	}
	summary.nextEarlyExit[n] = n
	for i := n - 1; i >= 0; i-- {
		summary.stmtIndex[stmts[i]] = i
		if statementMayExitEarly(stmts[i]) {
			summary.nextEarlyExit[i] = i
		} else {
			summary.nextEarlyExit[i] = summary.nextEarlyExit[i+1]
		}
	}

	recordIdent := func(expr *ast.Node, stmtIdx int, flags consumeFlags) {
		u := internalUtils.SkipAssertionsAndParens(expr)
		if u == nil || u.Kind != ast.KindIdentifier {
			return
		}
		site := consumeSite{stmtIdx: stmtIdx, flags: flags}
		if sym := r.resolveSymbol(u); sym != nil {
			summary.bySym[sym] = append(summary.bySym[sym], site)
		} else {
			name := u.AsIdentifier().Text
			summary.byName[name] = append(summary.byName[name], site)
		}
	}

	var recordConsumedExpr func(expr *ast.Node, stmtIdx int, baseFlags consumeFlags, depth int)
	recordConsumedExpr = func(expr *ast.Node, stmtIdx int, baseFlags consumeFlags, depth int) {
		if depth > 8 {
			return
		}
		u := internalUtils.SkipAssertionsAndParens(expr)
		if u == nil {
			return
		}
		flags := baseFlags
		for {
			m, ok := r.promiseChainMethodName(u)
			if !ok {
				break
			}
			if m != "finally" {
				flags |= consumeNonFinally
			}
			callee := internalUtils.SkipAssertionsAndParens(u.AsCallExpression().Expression)
			u = internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
			if u == nil {
				return
			}
		}
		if u.Kind == ast.KindIdentifier {
			recordIdent(u, stmtIdx, flags)
			return
		}
		if method, ok := r.promiseMethodName(u); ok {
			args := u.AsCallExpression().Arguments
			if args == nil || len(args.Nodes) == 0 {
				return
			}
			arg0 := internalUtils.SkipAssertionsAndParens(args.Nodes[0])
			switch method {
			case "resolve":
				recordConsumedExpr(arg0, stmtIdx, flags, depth+1)
			case "all", "allSettled", "race", "any":
				if method == "allSettled" || method == "any" {
					flags |= consumeNonFinally
				}
				if arg0 != nil && arg0.Kind == ast.KindArrayLiteralExpression && arg0.AsArrayLiteralExpression().Elements != nil {
					for _, el := range arg0.AsArrayLiteralExpression().Elements.Nodes {
						recordConsumedExpr(el, stmtIdx, flags, depth+1)
					}
				}
			}
		}
	}

	for i, stmt := range stmts {
		if stmt == nil {
			continue
		}
		switch stmt.Kind {
		case ast.KindExpressionStatement:
			expr := internalUtils.SkipAssertionsAndParens(stmt.AsExpressionStatement().Expression)
			if expr != nil && expr.Kind == ast.KindAwaitExpression {
				recordConsumedExpr(expr.AsAwaitExpression().Expression, i, 0, 0)
			} else if expr != nil && expr.Kind == ast.KindBinaryExpression {
				if bin := expr.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil &&
					bin.OperatorToken.Kind == ast.KindEqualsToken {
					if rhs := internalUtils.SkipAssertionsAndParens(bin.Right); rhs != nil && rhs.Kind == ast.KindAwaitExpression {
						recordConsumedExpr(rhs.AsAwaitExpression().Expression, i, 0, 0)
					}
				}
			}
		case ast.KindVariableStatement:
			if dl := stmt.AsVariableStatement().DeclarationList; dl != nil && dl.AsVariableDeclarationList().Declarations != nil {
				for _, d := range dl.AsVariableDeclarationList().Declarations.Nodes {
					if d == nil || d.AsVariableDeclaration().Initializer == nil {
						continue
					}
					init := internalUtils.SkipAssertionsAndParens(d.AsVariableDeclaration().Initializer)
					if init != nil && init.Kind == ast.KindAwaitExpression {
						recordConsumedExpr(init.AsAwaitExpression().Expression, i, 0, 0)
					}
				}
			}
		case ast.KindReturnStatement:
			ret := stmt.AsReturnStatement()
			if ret != nil && ret.Expression != nil {
				expr := internalUtils.SkipAssertionsAndParens(ret.Expression)
				if expr != nil && expr.Kind == ast.KindAwaitExpression {
					recordConsumedExpr(expr.AsAwaitExpression().Expression, i, 0, 0)
				} else {
					recordConsumedExpr(expr, i, consumeViaReturn, 0)
				}
			}
		}
	}

	if r != nil {
		if r.blockAwaitCache == nil {
			r.blockAwaitCache = map[*ast.Block]*blockAwaitSummary{}
		}
		r.blockAwaitCache[block] = summary
	}
	return summary
}

// isBoundToDirectAwaitedConst reports whether `outermost` is the initializer of
// a `const p = ...` declaration that is directly awaited (or returned when
// !alwaysAwait, or passed into an awaited `Promise.all/race/allSettled/any([p])`)
// in a subsequent top-level statement of the same block with no intervening
// early exit.
func (r *AsyncAssertionResolver) isBoundToDirectAwaitedConst(outermost *ast.Node, alwaysAwait bool, isReject bool) bool {
	if outermost == nil || outermost.Parent == nil || outermost.Parent.Kind != ast.KindVariableDeclaration {
		return false
	}
	vd := outermost.Parent
	if vd.AsVariableDeclaration().Initializer != outermost ||
		!ast.IsVarConst(vd) || ast.IsVarAwaitUsing(vd) {
		return false
	}
	nameNode := vd.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return false
	}
	if vd.Parent == nil || vd.Parent.Kind != ast.KindVariableDeclarationList ||
		vd.Parent.Parent == nil || vd.Parent.Parent.Kind != ast.KindVariableStatement ||
		vd.Parent.Parent.Parent == nil || vd.Parent.Parent.Parent.Kind != ast.KindBlock {
		return false
	}
	varStmt := vd.Parent.Parent
	block := varStmt.Parent.AsBlock()
	if block == nil || block.Statements == nil {
		return false
	}
	summary := r.getBlockAwaitSummary(block)
	declIdx, ok := summary.stmtIndex[varStmt]
	if !ok {
		return false
	}
	limit := summary.nextEarlyExit[declIdx+1]
	var sites []consumeSite
	if r != nil && (r.refs != nil || r.typeChecker != nil) {
		if declSym := internalUtils.GetVariableDeclarationSymbol(vd, r.typeChecker); declSym != nil {
			sites = summary.bySym[declSym]
		} else {
			sites = summary.byName[nameNode.AsIdentifier().Text]
		}
	} else {
		sites = summary.byName[nameNode.AsIdentifier().Text]
	}
	for _, site := range sites {
		if site.stmtIdx <= declIdx || site.stmtIdx > limit {
			continue
		}
		if alwaysAwait && site.flags&consumeViaReturn != 0 {
			continue
		}
		if isReject && site.flags&consumeNonFinally != 0 {
			continue
		}
		return true
	}
	return false
}

func statementMayExitEarly(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}
	found := false
	var visit func(node *ast.Node, inLoop, inSwitch bool)
	visit = func(node *ast.Node, inLoop, inSwitch bool) {
		if node == nil || found || ast.IsTypeNode(node) {
			return
		}
		if node != stmt && (testFramework.IsFunction(node) || node.Kind == ast.KindClassStaticBlockDeclaration) {
			return
		}
		switch node.Kind {
		case ast.KindReturnStatement, ast.KindThrowStatement:
			found = true
			return
		case ast.KindBreakStatement:
			if node.AsBreakStatement().Label != nil || (!inLoop && !inSwitch) {
				found = true
				return
			}
		case ast.KindContinueStatement:
			if node.AsContinueStatement().Label != nil || !inLoop {
				found = true
				return
			}
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement, ast.KindWhileStatement, ast.KindDoStatement:
			inLoop = true
		case ast.KindSwitchStatement:
			inSwitch = true
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, inLoop, inSwitch)
			return found
		})
	}
	visit(stmt, false, false)
	return found
}

func (r *AsyncAssertionResolver) unwrapPromiseWrapperForConsumption(expr *ast.Node, isReject bool) (*ast.Node, bool, bool) {
	wrap, _, wrapReject := r.findPromiseWrapperCall(expr)
	if wrap == nil {
		return nil, isReject, false
	}
	if isReject {
		// `Promise.allSettled` and `Promise.any` swallow rejected inputs
		// without awaiting the inner assertion promise inside `Promise.reject(...)`.
		if m, ok := r.promiseMethodName(wrap); ok && (m == "allSettled" || m == "any") {
			return nil, isReject, false
		}
	}
	nextReject := isReject || wrapReject
	return r.getParentIfPromiseChained(wrap, nextReject), nextReject, true
}

func (r *AsyncAssertionResolver) isExpressionConsumed(node *ast.Node, alwaysAwait bool, isReject bool) bool {
	cur := node
	for depth := 0; depth < 8 && cur != nil; depth++ {
		outer := internalUtils.OutermostTransparentExpression(cur)
		if outer == nil || outer.Parent == nil {
			return false
		}
		if isAcceptableReturnNode(outer, outer.Parent, !alwaysAwait) ||
			r.isBoundToDirectAwaitedConst(outer, alwaysAwait, isReject) {
			return true
		}
		if next, nextReject, ok := r.unwrapPromiseWrapperForConsumption(outer, isReject); ok {
			isReject = nextReject
			cur = next
			continue
		}
		p := outer.Parent
		if p.Kind == ast.KindConditionalExpression {
			cond := p.AsConditionalExpression()
			if cond != nil && (cond.WhenTrue == outer || cond.WhenFalse == outer) {
				next := r.getParentIfPromiseChained(p, isReject)
				if wrapped, nextReject, ok := r.unwrapPromiseWrapperForConsumption(next, isReject); ok {
					isReject = nextReject
					next = wrapped
				}
				cur = next
				continue
			}
		}
		if p.Kind == ast.KindBinaryExpression && internalUtils.IsCommaOperator(p) &&
			p.AsBinaryExpression().Right == outer {
			next := r.getParentIfPromiseChained(p, isReject)
			if wrapped, nextReject, ok := r.unwrapPromiseWrapperForConsumption(next, isReject); ok {
				isReject = nextReject
				next = wrapped
			}
			cur = next
			continue
		}
		return false
	}
	return false
}

// ResolveAsyncAssertionReportNode decides where an async assertion is reported
// and whether it has to be reported at all. assertionNode must be the outermost
// expression of the assertion chain: for a Chai chain carrying several matchers,
// such as expect(p).resolves.to.be.a("string").that.contains("x"), an inner
// matcher's call would be inspected against the wrong parent.
//
// reportNode is the node the report and the await fix apply to. promiseWrapped
// tells the caller to use the Promise-specific message, and insideAssertionArray
// marks a Promise.all([...]) array whose assertions report only once.
func (r *AsyncAssertionResolver) ResolveAsyncAssertionReportNode(
	assertionNode *ast.Node,
	alwaysAwait bool,
) (reportNode *ast.Node, promiseWrapped bool, insideAssertionArray bool, shouldReport bool) {
	if assertionNode == nil {
		return nil, false, false, false
	}

	promiseChainedAssertionNode := r.getParentIfPromiseChained(assertionNode, false)
	reportNode = promiseChainedAssertionNode
	isReject := false
	if promiseCallNode, inArr, rej := r.findPromiseWrapperCall(promiseChainedAssertionNode); promiseCallNode != nil {
		reportNode = r.getParentIfPromiseChained(promiseCallNode, rej)
		promiseWrapped = true
		insideAssertionArray = inArr
		isReject = rej
	} else {
		insideAssertionArray = inArr
	}

	outermost := internalUtils.OutermostTransparentExpression(reportNode)
	if outermost == nil || outermost.Parent == nil || r.isExpressionConsumed(reportNode, alwaysAwait, isReject) {
		return reportNode, promiseWrapped, insideAssertionArray, false
	}

	return reportNode, promiseWrapped, insideAssertionArray, true
}

// ReportAsyncDescriptor reports a floating async assertion and attaches a
// self-contained, atomic autofix when safe. For a synchronous enclosing
// function, every fixable diagnostic carries both `async` and `await` so
// cross-rule fix conflicts can never emit a bare `await` into a sync function.
func (r *AsyncAssertionResolver) ReportAsyncDescriptor(
	ctx rule.RuleContext,
	reportNode *ast.Node,
	msg rule.RuleMessage,
	alwaysAwait bool,
	fnFixable map[*ast.Node]bool,
) {
	ctx.ReportNodeWithDeferredFixes(reportNode, msg, func() []rule.RuleFix {
		fn, ok := r.FindFixableContainingFunctionCached(reportNode, fnFixable)
		if !ok || !r.CanSafelyAwaitFix(reportNode) {
			return nil
		}
		var fixes []rule.RuleFix
		if !IsAsyncFunctionLike(fn) {
			fixes = append(fixes, AsyncInsertFix(ctx.SourceFile, fn))
		}
		fixes = append(fixes, r.AwaitFix(ctx.SourceFile, reportNode, alwaysAwait))
		return fixes
	})
}

// PromiseArrayExceptionKey identifies a Promise.all([...]) call by source range,
// so the assertions inside its array report once between them.
func PromiseArrayExceptionKey(sourceFile *ast.SourceFile, node *ast.Node) string {
	if sourceFile == nil || node == nil {
		return ""
	}
	r := internalUtils.TrimNodeTextRange(sourceFile, node)
	return fmt.Sprintf("%d:%d", r.Pos(), r.End())
}

// IsAsyncFunctionLike reports whether fn already carries the `async` modifier
// (including `async function*` and `async *m()`, unlike `ast.IsAsyncFunction`).
func IsAsyncFunctionLike(fn *ast.Node) bool {
	return fn != nil && ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync)
}

func functionAsteriskToken(fn *ast.Node) *ast.Node {
	if fn == nil {
		return nil
	}
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		return fn.AsFunctionDeclaration().AsteriskToken
	case ast.KindFunctionExpression:
		return fn.AsFunctionExpression().AsteriskToken
	case ast.KindMethodDeclaration:
		return fn.AsMethodDeclaration().AsteriskToken
	default:
		return nil
	}
}

// FindFixableContainingFunctionCached returns the enclosing function that can
// safely receive `await` (and `async` if not already async) for node, memoizing
// per-function fixability in fnAllowed.
func (r *AsyncAssertionResolver) FindFixableContainingFunctionCached(node *ast.Node, fnAllowed map[*ast.Node]bool) (*ast.Node, bool) {
	if node == nil {
		return nil, false
	}
	prev := node
	for current := node.Parent; current != nil; prev, current = current, current.Parent {
		switch current.Kind {
		case ast.KindClassStaticBlockDeclaration:
			return nil, false
		case ast.KindPropertyDeclaration:
			if prop := current.AsPropertyDeclaration(); prop != nil && prop.Initializer == prev {
				return nil, false
			}
		case ast.KindParameter:
			return nil, false
		case ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindMethodSignature, ast.KindCallSignature, ast.KindConstructSignature,
			ast.KindFunctionType, ast.KindConstructorType:
			if prev.Kind == ast.KindDecorator || prev == current.Name() {
				continue
			}
			return nil, false
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration:
			if prev.Kind == ast.KindDecorator || prev == current.Name() {
				continue
			}
			if current.Body() != prev {
				return nil, false
			}
			if allowed, cached := fnAllowed[current]; cached {
				if !allowed {
					return nil, false
				}
				if !IsAsyncFunctionLike(current) && r.hasUnrelatedSyncReturn(current, node) {
					return nil, false
				}
				return current, true
			}
			allowed := r.isFunctionSafelyAsyncFixable(current)
			if fnAllowed != nil {
				fnAllowed[current] = allowed
			}
			if !allowed {
				return nil, false
			}
			if !IsAsyncFunctionLike(current) && r.hasUnrelatedSyncReturn(current, node) {
				return nil, false
			}
			return current, true
		}
	}
	return nil, false
}

func isTestOrHookName(name string) bool {
	switch name {
	case "test", "it", "fit", "xit", "beforeEach", "afterEach", "beforeAll", "afterAll":
		return true
	}
	return false
}

func isTestOrHookCallbackArg(arg *ast.Node) bool {
	argOuter := internalUtils.OutermostTransparentExpression(arg)
	if argOuter == nil || argOuter.Parent == nil || argOuter.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := argOuter.Parent.AsCallExpression()
	if call == nil || call.Expression == argOuter {
		return false
	}
	cur := internalUtils.SkipAssertionsAndParens(call.Expression)
	for cur != nil {
		switch {
		case cur.Kind == ast.KindIdentifier:
			return isTestOrHookName(cur.AsIdentifier().Text)
		case ast.IsAccessExpression(cur):
			if name, ok := internalUtils.AccessExpressionStaticName(cur); ok && isTestOrHookName(name) {
				return true
			}
			cur = internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(cur))
		case cur.Kind == ast.KindCallExpression:
			cur = internalUtils.SkipAssertionsAndParens(cur.AsCallExpression().Expression)
		default:
			return false
		}
	}
	return false
}

func (r *AsyncAssertionResolver) functionSymbol(fn *ast.Node) *ast.Symbol {
	if r == nil || fn == nil {
		return nil
	}
	if fn.Kind == ast.KindFunctionDeclaration {
		if sym := fn.Symbol(); sym != nil {
			return sym
		}
		if name := fn.Name(); name != nil && name.Kind == ast.KindIdentifier {
			return r.resolveSymbol(name)
		}
		return nil
	}
	fnOuter := internalUtils.OutermostTransparentExpression(fn)
	if fnOuter == nil || fnOuter.Parent == nil {
		return nil
	}
	switch fnOuter.Parent.Kind {
	case ast.KindVariableDeclaration:
		if fnOuter.Parent.AsVariableDeclaration().Initializer == fnOuter {
			return internalUtils.GetVariableDeclarationSymbol(fnOuter.Parent, r.typeChecker)
		}
	case ast.KindBinaryExpression:
		bin := fnOuter.Parent.AsBinaryExpression()
		if bin != nil && bin.Right == fnOuter && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
			if lhs := internalUtils.SkipAssertionsAndParens(bin.Left); lhs != nil && lhs.Kind == ast.KindIdentifier {
				return r.resolveSymbol(lhs)
			}
		}
	}
	return nil
}

// walkFunctionAliasRefs visits references to `sym` and up to `maxDepth` hops of
// local variable/assignment aliases (`const Alias = Worker; Alias = Worker;`),
// stopping early if `visit` returns true.
func (r *AsyncAssertionResolver) walkFunctionAliasRefs(sym *ast.Symbol, maxDepth int, visit func(ref *ast.Node) bool) bool {
	if r == nil || r.refs == nil || sym == nil {
		return false
	}
	type queueItem struct {
		sym   *ast.Symbol
		depth int
	}
	queue := []queueItem{{sym: sym, depth: 0}}
	visited := map[*ast.Symbol]bool{sym: true}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for _, ref := range r.refs.References(cur.sym) {
			if ref == nil {
				continue
			}
			if visit(ref) {
				return true
			}
			if cur.depth >= maxDepth {
				continue
			}
			outer := internalUtils.OutermostTransparentExpression(ref)
			if outer == nil || outer.Parent == nil {
				continue
			}
			var nextSym *ast.Symbol
			switch outer.Parent.Kind {
			case ast.KindVariableDeclaration:
				vd := outer.Parent.AsVariableDeclaration()
				if vd != nil && vd.Initializer == outer && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
					nextSym = internalUtils.GetVariableDeclarationSymbol(outer.Parent, r.typeChecker)
				}
			case ast.KindBinaryExpression:
				bin := outer.Parent.AsBinaryExpression()
				if bin != nil && bin.Right == outer && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
					if lhs := internalUtils.SkipAssertionsAndParens(bin.Left); lhs != nil && lhs.Kind == ast.KindIdentifier {
						nextSym = r.resolveSymbol(lhs)
					}
				}
			}
			if nextSym != nil && !visited[nextSym] {
				visited[nextSym] = true
				queue = append(queue, queueItem{sym: nextSym, depth: cur.depth + 1})
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isParamUsedAsCallback(param *ast.Node) bool {
	if r == nil || r.refs == nil || param == nil || param.Name() == nil {
		return false
	}
	sym := internalUtils.BindingNameSymbol(param.Name())
	if sym == nil {
		return false
	}
	for _, ref := range r.refs.References(sym) {
		if ref == nil {
			continue
		}
		refOuter := internalUtils.OutermostTransparentExpression(ref)
		if refOuter != nil && refOuter.Parent != nil && refOuter.Parent.Kind == ast.KindCallExpression {
			if refOuter.Parent.AsCallExpression().Expression == refOuter {
				return true
			}
			if _, isChain := r.promiseChainMethodName(refOuter.Parent); isChain {
				return true
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isDoneCallbackFunction(fn *ast.Node) bool {
	if fn == nil || (fn.Kind != ast.KindArrowFunction && fn.Kind != ast.KindFunctionExpression && fn.Kind != ast.KindFunctionDeclaration) {
		return false
	}
	params := fn.Parameters()
	if len(params) == 0 {
		return false
	}
	isCallArg := false
	isTestOrHook := false
	if fn.Kind != ast.KindFunctionDeclaration {
		fnOuter := internalUtils.OutermostTransparentExpression(fn)
		if fnOuter != nil && fnOuter.Parent != nil && fnOuter.Parent.Kind == ast.KindCallExpression &&
			fnOuter.Parent.AsCallExpression().Expression != fnOuter {
			isCallArg = true
			isTestOrHook = isTestOrHookCallbackArg(fn)
		}
	}
	if !isTestOrHook {
		if sym := r.functionSymbol(fn); sym != nil {
			r.walkFunctionAliasRefs(sym, 3, func(ref *ast.Node) bool {
				refOuter := internalUtils.OutermostTransparentExpression(ref)
				if refOuter != nil && refOuter.Parent != nil && refOuter.Parent.Kind == ast.KindCallExpression &&
					refOuter.Parent.AsCallExpression().Expression != refOuter {
					isCallArg = true
					if isTestOrHookCallbackArg(ref) {
						isTestOrHook = true
						return true
					}
				}
				return false
			})
		}
	}
	for _, param := range params {
		if param == nil || param.Name() == nil || param.Name().Kind != ast.KindIdentifier {
			continue
		}
		isNamedDone := param.Name().AsIdentifier().Text == "done"
		if isNamedDone && isCallArg {
			return true
		}
		if (isTestOrHook || isNamedDone) && r.isParamUsedAsCallback(param) {
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isConstructedOrExtendedFunction(fn *ast.Node) bool {
	if fn == nil || (fn.Kind != ast.KindFunctionDeclaration && fn.Kind != ast.KindFunctionExpression) {
		return false
	}
	if fn.Kind == ast.KindFunctionExpression {
		fnOuter := internalUtils.OutermostTransparentExpression(fn)
		if fnOuter != nil && fnOuter.Parent != nil {
			switch fnOuter.Parent.Kind {
			case ast.KindNewExpression:
				if fnOuter.Parent.AsNewExpression().Expression == fnOuter {
					return true
				}
			case ast.KindExpressionWithTypeArguments:
				if fnOuter.Parent.AsExpressionWithTypeArguments().Expression == fnOuter {
					return true
				}
			}
		}
	}
	sym := r.functionSymbol(fn)
	if sym == nil {
		return false
	}
	return r.walkFunctionAliasRefs(sym, 4, func(ref *ast.Node) bool {
		outer := internalUtils.OutermostTransparentExpression(ref)
		if outer == nil || outer.Parent == nil {
			return false
		}
		switch outer.Parent.Kind {
		case ast.KindNewExpression:
			return outer.Parent.AsNewExpression().Expression == outer
		case ast.KindExpressionWithTypeArguments:
			return outer.Parent.AsExpressionWithTypeArguments().Expression == outer
		}
		return false
	})
}

func isNonVoidNonPromiseType(tc *checker.Checker, node *ast.Node, t *checker.Type) bool {
	if tc == nil || t == nil {
		return false
	}
	const safeFlags = checker.TypeFlagsVoid | checker.TypeFlagsUndefined | checker.TypeFlagsNever |
		checker.TypeFlagsAny | checker.TypeFlagsUnknown
	for _, part := range internalUtils.UnionTypeParts(t) {
		if part != nil && part.Flags()&safeFlags == 0 && !internalUtils.IsThenableType(tc, node, part) {
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isLikelyAsyncReturnExpr(expr *ast.Node) bool {
	u := internalUtils.SkipAssertionsAndParens(expr)
	if u == nil {
		return true
	}
	if u.Kind == ast.KindAwaitExpression {
		return true
	}
	if u.Kind == ast.KindConditionalExpression {
		cond := u.AsConditionalExpression()
		return cond != nil && r.isLikelyAsyncReturnExpr(cond.WhenTrue) && r.isLikelyAsyncReturnExpr(cond.WhenFalse)
	}
	if m, ok := r.promiseMethodName(u); ok && !isNonPromiseWrapperMethod(m) {
		return true
	}
	if _, ok := r.promiseChainMethodName(u); ok {
		return true
	}
	if r != nil && r.typeChecker != nil {
		if t := r.typeChecker.GetTypeAtLocation(u); t != nil && internalUtils.IsThenableType(r.typeChecker, u, t) {
			return true
		}
	}
	if u.Kind == ast.KindCallExpression {
		cur := internalUtils.SkipAssertionsAndParens(u.AsCallExpression().Expression)
		allowBuiltinToResolve := r == nil || !r.disallowBuiltinToResolveReject
		hasExpectRoot := false
		hasExpectAsyncProp := false
		for cur != nil {
			if ast.IsAccessExpression(cur) {
				if name, ok := r.accessStaticNameThroughAssertions(cur); ok {
					switch name {
					case "resolves", "rejects":
						return true
					case "poll", "element":
						hasExpectAsyncProp = true
					case "toResolve", "toReject":
						if allowBuiltinToResolve {
							hasExpectAsyncProp = true
						}
					case "expect":
						hasExpectRoot = true
					}
				}
				cur = internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(cur))
				continue
			}
			if cur.Kind == ast.KindCallExpression {
				cur = internalUtils.SkipAssertionsAndParens(cur.AsCallExpression().Expression)
				continue
			}
			if cur.Kind == ast.KindIdentifier && cur.AsIdentifier().Text == "expect" {
				hasExpectRoot = true
			}
			break
		}
		if hasExpectRoot && hasExpectAsyncProp {
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) hasNonPromiseSyncReturnType(fn *ast.Node) bool {
	if fn == nil {
		return false
	}
	// Check explicit typed variable annotation: `const cb: () => number = () => { ... }`.
	fnOuter := internalUtils.OutermostTransparentExpression(fn)
	if fnOuter != nil && fnOuter.Parent != nil && fnOuter.Parent.Kind == ast.KindVariableDeclaration &&
		fnOuter.Parent.AsVariableDeclaration().Initializer == fnOuter {
		if vdType := fnOuter.Parent.AsVariableDeclaration().Type; vdType != nil && vdType.Kind == ast.KindFunctionType {
			if hasNonPromiseReturnTypeAnnotation(vdType) {
				return true
			}
		}
	}
	// Check inferred or contextual return type when TypeChecker is available.
	if r != nil && r.typeChecker != nil {
		if fnType := r.typeChecker.GetTypeAtLocation(fn); fnType != nil {
			for _, sig := range internalUtils.GetCallSignatures(r.typeChecker, fnType) {
				retType := checker.Checker_getReturnTypeOfSignature(r.typeChecker, sig)
				if isNonVoidNonPromiseType(r.typeChecker, fn, retType) {
					return true
				}
			}
		}
	}
	return false
}

// hasUnrelatedSyncReturn reports whether a synchronous function returns a
// non-async value other than `reportNode` itself (e.g. `return 1` or
// `return Promise.withResolvers()`), which adding `async` would wrap in a Promise.
func (r *AsyncAssertionResolver) hasUnrelatedSyncReturn(fn *ast.Node, reportNode *ast.Node) bool {
	if fn == nil {
		return false
	}
	if r != nil && r.syncFnReturnCache != nil {
		if cached, ok := r.syncFnReturnCache[fn]; ok {
			return cached.count > 1 || (cached.count == 1 && internalUtils.SkipAssertionsAndParens(cached.only) != internalUtils.SkipAssertionsAndParens(reportNode))
		}
	}
	var summary syncFnReturnSummary
	body := fn.Body()
	if body != nil && body.Kind == ast.KindBlock {
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			if node == nil || summary.count > 1 || ast.IsTypeNode(node) {
				return
			}
			if node != body && (testFramework.IsFunction(node) || node.Kind == ast.KindClassStaticBlockDeclaration) {
				return
			}
			if node.Kind == ast.KindReturnStatement {
				if ret := node.AsReturnStatement(); ret != nil && ret.Expression != nil && !r.isLikelyAsyncReturnExpr(ret.Expression) {
					summary.count++
					if summary.count == 1 {
						summary.only = ret.Expression
					}
					return
				}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				visit(child)
				return summary.count > 1
			})
		}
		visit(body)
	}
	if r != nil {
		if r.syncFnReturnCache == nil {
			r.syncFnReturnCache = map[*ast.Node]syncFnReturnSummary{}
		}
		r.syncFnReturnCache[fn] = summary
	}
	return summary.count > 1 || (summary.count == 1 && internalUtils.SkipAssertionsAndParens(summary.only) != internalUtils.SkipAssertionsAndParens(reportNode))
}

func (r *AsyncAssertionResolver) isFunctionSafelyAsyncFixable(fn *ast.Node) bool {
	if ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync) {
		return true
	}
	if functionAsteriskToken(fn) != nil {
		return false
	}
	if hasExplicitNonPromiseReturnType(fn) {
		return false
	}
	if fn.Kind == ast.KindFunctionDeclaration && !isValidAsyncFunctionDeclarationParent(fn.Parent) {
		return false
	}
	if r.isConstructedOrExtendedFunction(fn) {
		return false
	}
	if r.isDoneCallbackFunction(fn) {
		return false
	}
	if r.hasNonPromiseSyncReturnType(fn) {
		return false
	}
	if containsIdentifierNamedAwait(fn) {
		return false
	}
	return true
}

func isValidAsyncFunctionDeclarationParent(parent *ast.Node) bool {
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindSourceFile, ast.KindBlock, ast.KindModuleBlock, ast.KindCaseClause, ast.KindDefaultClause:
		return true
	default:
		return false
	}
}

func containsIdentifierNamedAwait(fn *ast.Node) bool {
	if fn.Kind == ast.KindFunctionExpression || fn.Kind == ast.KindFunctionDeclaration {
		if n := fn.Name(); n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == "await" {
			return true
		}
	}
	found := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		if node != fn {
			switch node.Kind {
			case ast.KindFunctionDeclaration:
				if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == "await" {
					found = true
				}
				return
			case ast.KindFunctionExpression:
				return
			case ast.KindArrowFunction:
				// ArrowParameters inherit [+Await] from the enclosing function,
				// so `(x = await) => x` becomes a syntax error if the outer
				// function becomes `async`.
				for _, param := range node.Parameters() {
					visit(param)
				}
				return
			case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
				if mods := node.Modifiers(); mods != nil {
					for _, mod := range mods.Nodes {
						if mod != nil && mod.Kind == ast.KindDecorator {
							visit(mod)
						}
					}
				}
				if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					visit(name)
				}
				return
			case ast.KindTypeReference, ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration:
				return
			}
		}
		if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == "await" {
			parent := node.Parent
			if parent != nil {
				switch parent.Kind {
				case ast.KindPropertyAccessExpression:
					if parent.AsPropertyAccessExpression().Name() == node {
						return
					}
				case ast.KindPropertyAssignment, ast.KindMethodDeclaration, ast.KindPropertyDeclaration,
					ast.KindGetAccessor, ast.KindSetAccessor:
					if parent.Name() == node {
						return
					}
				case ast.KindBindingElement:
					if parent.AsBindingElement().PropertyName == node {
						return
					}
				}
			}
			found = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	for _, param := range fn.Parameters() {
		visit(param)
	}
	visit(fn.Body())
	return found
}

func hasNonPromiseReturnTypeAnnotation(fn *ast.Node) bool {
	if fn == nil {
		return false
	}
	retType := fn.Type()
	if retType == nil {
		return false
	}
	if retType.Kind == ast.KindTypeReference {
		ref := retType.AsTypeReferenceNode()
		if ref != nil && ref.TypeName != nil {
			if ref.TypeName.Kind == ast.KindIdentifier && ref.TypeName.AsIdentifier().Text == "Promise" {
				return false
			}
			if ref.TypeName.Kind == ast.KindQualifiedName {
				qn := ref.TypeName.AsQualifiedName()
				if qn != nil && qn.Left != nil && qn.Left.Kind == ast.KindIdentifier &&
					qn.Left.AsIdentifier().Text == "globalThis" &&
					qn.Right != nil && qn.Right.Text() == "Promise" {
					return false
				}
			}
		}
	}
	return true
}

func hasExplicitNonPromiseReturnType(fn *ast.Node) bool {
	if hasNonPromiseReturnTypeAnnotation(fn) {
		return true
	}
	// Check overload signatures on the same symbol so adding `async` never
	// breaks `function f(x: number): void; function f(x: number) { ... }`.
	if sym := fn.Symbol(); sym != nil && len(sym.Declarations) > 1 {
		for _, decl := range sym.Declarations {
			if decl != nil && decl != fn && hasNonPromiseReturnTypeAnnotation(decl) {
				return true
			}
		}
	}
	return false
}

func needsLeadingSpace(text string, pos int) bool {
	if pos <= 0 || pos > len(text) {
		return false
	}
	r, _ := utf8.DecodeLastRuneInString(text[:pos])
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$'
}

// AsyncInsertFix marks the function containing an async assertion as async.
func AsyncInsertFix(sourceFile *ast.SourceFile, fn *ast.Node) rule.RuleFix {
	text := sourceFile.Text()
	var pos int
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		pos = internalUtils.FindFunctionKeywordPos(sourceFile, fn)
	case ast.KindMethodDeclaration:
		if asterisk := functionAsteriskToken(fn); asterisk != nil {
			pos = scanner.SkipTrivia(text, asterisk.Pos())
		} else if fn.Name() != nil {
			pos = scanner.SkipTrivia(text, fn.Name().Pos())
		} else {
			pos = internalUtils.GetFunctionHeadLoc(sourceFile, fn).Pos()
		}
	default:
		pos = internalUtils.TrimNodeTextRange(sourceFile, fn).Pos()
	}
	insertion := "async "
	if needsLeadingSpace(text, pos) {
		insertion = " async "
	}
	return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), insertion)
}

func (r *AsyncAssertionResolver) fileMayExtendExpect() bool {
	if r == nil || r.sourceFile == nil {
		return false
	}
	if r.checkedFileExtend {
		return r.fileHasExtend
	}
	r.checkedFileExtend = true
	found := false
	isExtendOrEquality := func(s string) bool {
		return s == "extend" || s == "addEqualityTesters"
	}
	ev := r.staticStringEvaluator()
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || found || ast.IsTypeNode(node) {
			return
		}
		switch node.Kind {
		case ast.KindIdentifier:
			if isExtendOrEquality(node.AsIdentifier().Text) {
				found = true
				return
			}
		case ast.KindStringLiteral:
			if isExtendOrEquality(node.AsStringLiteral().Text) {
				found = true
				return
			}
		case ast.KindNoSubstitutionTemplateLiteral:
			if isExtendOrEquality(node.Text()) {
				found = true
				return
			}
		case ast.KindElementAccessExpression:
			if arg := node.AsElementAccessExpression().ArgumentExpression; arg != nil && ev != nil {
				if folded, ok := ev.EvalToString(arg); ok && isExtendOrEquality(folded) {
					found = true
					return
				}
			}
		case ast.KindComputedPropertyName:
			if expr := node.AsComputedPropertyName().Expression; expr != nil && ev != nil {
				if folded, ok := ev.EvalToString(expr); ok && isExtendOrEquality(folded) {
					found = true
					return
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(r.sourceFile.AsNode())
	r.fileHasExtend = found
	return found
}

func (r *AsyncAssertionResolver) isPurePrimitiveLiteral(node *ast.Node) bool {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindUndefinedKeyword:
		return true
	case ast.KindIdentifier:
		return node.AsIdentifier().Text == "undefined" && !r.isLocallyShadowed(node, "undefined")
	}
	return false
}

func (r *AsyncAssertionResolver) isPurePromiseLiteralArg(arg *ast.Node, depth int) bool {
	if depth > 3 {
		return false
	}
	u := internalUtils.SkipAssertionsAndParens(arg)
	if u == nil || u.Kind != ast.KindCallExpression {
		return false
	}
	if m, ok := r.promiseMethodName(u); !ok || (m != "resolve" && m != "reject") {
		return false
	}
	if args := u.AsCallExpression().Arguments; args != nil {
		for _, a := range args.Nodes {
			if !r.isPurePrimitiveLiteral(a) && !r.isPurePromiseLiteralArg(a, depth+1) {
				return false
			}
		}
	}
	return true
}

func isPureBuiltinAsyncMatcherName(name string) bool {
	switch name {
	case "toBe", "toEqual", "toStrictEqual",
		"toBeDefined", "toBeUndefined", "toBeNull",
		"toBeTruthy", "toBeFalsy", "toBeNaN",
		"toBeCloseTo", "toBeGreaterThan", "toBeGreaterThanOrEqual",
		"toBeLessThan", "toBeLessThanOrEqual",
		"toContain", "toContainEqual", "toHaveLength", "toMatch",
		"toMatchObject", "toHaveProperty", "toBeInstanceOf",
		"toHaveBeenCalled", "toHaveBeenCalledTimes", "toHaveBeenCalledWith",
		"toHaveBeenLastCalledWith", "toHaveBeenNthCalledWith",
		"toHaveReturned", "toHaveReturnedTimes", "toHaveReturnedWith",
		"toHaveLastReturnedWith", "toHaveNthReturnedWith":
		return true
	}
	return false
}

func (r *AsyncAssertionResolver) isFrameworkExpectImport(ident *ast.Node) bool {
	sym := r.resolveSymbol(ident)
	if sym == nil {
		return false
	}
	for _, decl := range sym.Declarations {
		if decl == nil || decl.Kind != ast.KindImportSpecifier || ast.IsTypeOnlyImportDeclaration(decl) {
			continue
		}
		imp := testFramework.FindImportDeclaration(decl)
		if imp == nil || imp.ModuleSpecifier == nil {
			continue
		}
		switch imp.ModuleSpecifier.Text() {
		case "@jest/globals", "@rstest/core", "rstack/test":
			if testFramework.ImportedSpecifierName(decl) == "expect" {
				return true
			}
		}
	}
	return false
}

// isPureFollowingArrayExpect reports whether an array element evaluated after
// an async assertion is a trivially pure built-in literal `expect(Promise.resolve/reject(<primitive>))`
// assertion (or, for Jest's standalone unread single-statement `const assertions = [...]`
// upstream fixtures, zero-arg `toResolve()`/`toReject()`). This keeps autofix
// safe without trying to prove arbitrary sibling expressions side-effect-free.
func (r *AsyncAssertionResolver) isPureFollowingArrayExpect(elem *ast.Node, allowStandaloneBare bool) bool {
	if r.fileMayExtendExpect() {
		return false
	}
	u := internalUtils.SkipAssertionsAndParens(elem)
	if u == nil || u.Kind != ast.KindCallExpression {
		return false
	}
	args := u.AsCallExpression().Arguments
	if args != nil {
		for _, a := range args.Nodes {
			if !r.isPurePrimitiveLiteral(a) {
				return false
			}
		}
	}
	outerCallee := internalUtils.SkipAssertionsAndParens(u.AsCallExpression().Expression)
	if outerCallee == nil || !ast.IsAccessExpression(outerCallee) {
		return false
	}
	matcherName, hasMatcher := r.accessStaticNameThroughAssertions(outerCallee)
	if !hasMatcher {
		return false
	}
	isBareAsync := (matcherName == "toResolve" || matcherName == "toReject") && (args == nil || len(args.Nodes) == 0)
	hasAsyncModifier := false
	cur := outerCallee
	for cur != nil {
		if ast.IsAccessExpression(cur) {
			if cur.Kind == ast.KindElementAccessExpression && !r.isPurePrimitiveLiteral(cur.AsElementAccessExpression().ArgumentExpression) {
				return false
			}
			if cur != outerCallee {
				modName, ok := r.accessStaticNameThroughAssertions(cur)
				if !ok {
					return false
				}
				switch modName {
				case "resolves", "rejects":
					hasAsyncModifier = true
				case "not":
				default:
					return false
				}
			}
			cur = internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(cur))
			continue
		}
		if cur.Kind == ast.KindCallExpression {
			headCallee := internalUtils.SkipAssertionsAndParens(cur.AsCallExpression().Expression)
			if headCallee == nil || headCallee.Kind != ast.KindIdentifier || headCallee.AsIdentifier().Text != "expect" {
				return false
			}
			if r.isLocallyShadowed(headCallee, "expect") && !r.isFrameworkExpectImport(headCallee) {
				return false
			}
			headArgs := cur.AsCallExpression().Arguments
			if headArgs == nil || len(headArgs.Nodes) != 1 || !r.isPurePromiseLiteralArg(headArgs.Nodes[0], 0) {
				return false
			}
			if hasAsyncModifier && isPureBuiltinAsyncMatcherName(matcherName) {
				return true
			}
			return allowStandaloneBare && isBareAsync
		}
		return false
	}
	return false
}

func (r *AsyncAssertionResolver) isStandaloneUnreadSingleStmtArray(arr *ast.Node) bool {
	if r == nil || r.disallowBuiltinToResolveReject || arr == nil || arr.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	outerArr := internalUtils.OutermostTransparentExpression(arr)
	vd := outerArr.Parent
	if vd == nil || vd.Kind != ast.KindVariableDeclaration || vd.AsVariableDeclaration().Initializer != outerArr {
		return false
	}
	name := vd.Name()
	if name == nil || name.Kind != ast.KindIdentifier || r.isBindingReadOrExported(vd, name) {
		return false
	}
	declList := vd.Parent
	if declList == nil || declList.Kind != ast.KindVariableDeclarationList ||
		ast.IsVarUsing(declList) || ast.IsVarAwaitUsing(declList) ||
		declList.AsVariableDeclarationList().Declarations == nil ||
		len(declList.AsVariableDeclarationList().Declarations.Nodes) != 1 ||
		declList.Parent == nil || declList.Parent.Kind != ast.KindVariableStatement {
		return false
	}
	varStmt := declList.Parent
	block := varStmt.Parent
	return block != nil && block.Kind == ast.KindBlock &&
		block.Parent != nil && testFramework.IsFunction(block.Parent) &&
		block.AsBlock().Statements != nil && len(block.AsBlock().Statements.Nodes) == 1 &&
		block.AsBlock().Statements.Nodes[0] == varStmt
}

func (r *AsyncAssertionResolver) getArrayPureSuffixInfo(arr *ast.Node) *arrayPureSuffixInfo {
	if r != nil && r.arrayPureSuffixCache != nil {
		if cached, ok := r.arrayPureSuffixCache[arr]; ok {
			return cached
		}
	}
	elems := arr.AsArrayLiteralExpression().Elements.Nodes
	allowStandaloneBare := r.isStandaloneUnreadSingleStmtArray(arr)
	elemIndex := make(map[*ast.Node]int, len(elems))
	for i, el := range elems {
		elemIndex[el] = i
	}
	minPureIdx := len(elems)
	for j := len(elems) - 1; j >= 0; j-- {
		if !r.isPureFollowingArrayExpect(elems[j], allowStandaloneBare) {
			break
		}
		minPureIdx = j
	}
	info := &arrayPureSuffixInfo{
		elemIndex:           elemIndex,
		minPureIdx:          minPureIdx,
		allowStandaloneBare: allowStandaloneBare,
	}
	if r != nil {
		if r.arrayPureSuffixCache == nil {
			r.arrayPureSuffixCache = map[*ast.Node]*arrayPureSuffixInfo{}
		}
		r.arrayPureSuffixCache[arr] = info
	}
	return info
}

func isVarShadowedByEnclosingCatch(vd *ast.Node, name string) bool {
	if vd == nil || vd.Parent == nil || vd.Parent.Flags&ast.NodeFlagsBlockScoped != 0 {
		return false
	}
	for cur := vd.Parent; cur != nil; cur = cur.Parent {
		if testFramework.IsFunction(cur) || cur.Kind == ast.KindClassStaticBlockDeclaration {
			return false
		}
		if cur.Kind == ast.KindCatchClause {
			cc := cur.AsCatchClause()
			if cc != nil && cc.VariableDeclaration != nil {
				matched := false
				internalUtils.CollectBindingNames(cc.VariableDeclaration.Name(), func(_ *ast.Node, n string) {
					if n == name {
						matched = true
					}
				})
				if matched {
					return true
				}
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isBindingReadOrExported(vd *ast.Node, ident *ast.Node) bool {
	if vd == nil || ident == nil || ident.Kind != ast.KindIdentifier {
		return true
	}
	if vd.Parent != nil && vd.Parent.Parent != nil && vd.Parent.Parent.Kind == ast.KindVariableStatement &&
		ast.HasSyntacticModifier(vd.Parent.Parent, ast.ModifierFlagsExport) {
		return true
	}
	name := ident.AsIdentifier().Text
	// Annex B `catch (e) { var e = ...; e.then(...) }`: references inside the
	// catch block resolve to the catch binding rather than the hoisted `var`.
	if isVarShadowedByEnclosingCatch(vd, name) {
		return true
	}
	if r != nil && r.refs != nil {
		if sym := internalUtils.GetVariableDeclarationSymbol(vd, r.typeChecker); sym != nil {
			for _, ref := range r.refs.References(sym) {
				if ref != nil && ref != ident {
					return true
				}
			}
			return false
		}
	}
	// Fallback when RefStore is unavailable: scan the enclosing scope for any
	// other identifier with the same name (including `typeof name` type queries).
	scope := ast.GetContainingFunction(vd)
	if scope == nil {
		if r != nil && r.sourceFile != nil {
			scope = r.sourceFile.AsNode()
		} else {
			return true
		}
	}
	read := false
	var visit func(*ast.Node)
	visit = func(n *ast.Node) {
		if n == nil || read {
			return
		}
		if ast.IsTypeNode(n) && n.Kind != ast.KindTypeQuery {
			return
		}
		if n.Kind == ast.KindIdentifier && n != ident && n.AsIdentifier().Text == name {
			if n.Parent != nil && n.Parent.Kind == ast.KindPropertyAccessExpression && n.Parent.AsPropertyAccessExpression().Name() == n {
				return
			}
			read = true
			return
		}
		n.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return read
		})
	}
	visit(scope)
	return read
}

func isInsideTryWithCatchOrAbruptFinally(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if testFramework.IsFunction(current) || current.Kind == ast.KindClassStaticBlockDeclaration {
			return false
		}
		if current.Parent != nil && current.Parent.Kind == ast.KindTryStatement {
			tryStmt := current.Parent.AsTryStatement()
			if tryStmt == nil {
				continue
			}
			if tryStmt.TryBlock == current {
				if tryStmt.CatchClause != nil {
					return true
				}
				if tryStmt.FinallyBlock != nil && statementMayExitEarly(tryStmt.FinallyBlock) {
					return true
				}
			} else if tryStmt.CatchClause == current {
				if tryStmt.FinallyBlock != nil && statementMayExitEarly(tryStmt.FinallyBlock) {
					return true
				}
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isGlobalPromiseMemberChainCall(call *ast.Node) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	cur := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	depth := 0
	for cur != nil && ast.IsAccessExpression(cur) && depth < 4 {
		seg, ok := r.accessStaticNameThroughAssertions(cur)
		if !ok || isNonCombinatorPromiseMethod(seg) {
			return false
		}
		cur = internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(cur))
		depth++
		if r.isGlobalPromiseObject(cur) {
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isLocallyDeclaredCallee(call *ast.Node) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return true
	}
	callee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier {
		return true
	}
	// Any resolved symbol (including ambient `declare function wrap(...)` or
	// local/imported helpers) may enforce Promise parameter types or invoke
	// `.then()` on its argument; only allow completely undeclared identifiers
	// (such as `foo(...)` in upstream Jest fixtures).
	if r.resolveSymbol(callee) != nil {
		return true
	}
	return internalUtils.IsShadowed(callee, callee.AsIdentifier().Text)
}

func isSingleElementArrayLiteral(node *ast.Node) bool {
	u := internalUtils.SkipAssertionsAndParens(node)
	return u != nil && u.Kind == ast.KindArrayLiteralExpression &&
		u.AsArrayLiteralExpression().Elements != nil &&
		len(u.AsArrayLiteralExpression().Elements.Nodes) == 1
}

func isArrayLiteralExpr(node *ast.Node) bool {
	u := internalUtils.SkipAssertionsAndParens(node)
	return u != nil && u.Kind == ast.KindArrayLiteralExpression
}

func (r *AsyncAssertionResolver) isInsideGlobalPromiseCombinatorArray(call *ast.Node) bool {
	callOuter := internalUtils.OutermostTransparentExpression(call)
	if callOuter == nil || callOuter.Parent == nil || callOuter.Parent.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	arr := callOuter.Parent
	if arr.AsArrayLiteralExpression().Elements == nil || len(arr.AsArrayLiteralExpression().Elements.Nodes) != 1 {
		return false
	}
	arrOuter := internalUtils.OutermostTransparentExpression(arr)
	if arrOuter == nil || arrOuter.Parent == nil || !isDirectFirstCallArg(arrOuter.Parent, arrOuter) {
		return false
	}
	args := arrOuter.Parent.AsCallExpression().Arguments
	if args == nil || len(args.Nodes) != 1 {
		return false
	}
	m, ok := r.promiseMethodName(arrOuter.Parent)
	return ok && (m == "all" || m == "allSettled" || m == "race" || m == "any")
}

// CanSafelyAwaitFix reports whether inserting `await` at node preserves the
// test's control flow, syntax validity, and sibling evaluation order.
//
// Rather than attempting to prove arbitrary sibling expressions side-effect-free,
// autofix is restricted to structurally safe positions (direct statements,
// direct returns, conditional branches, unread sole/last non-using variable
// declarations, and final array elements — plus pure literal built-in async
// expect arrays and global Promise combinators for upstream compatibility). All
// other positions still report the diagnostic without emitting an autofix.
func (r *AsyncAssertionResolver) CanSafelyAwaitFix(node *ast.Node) bool {
	if node == nil {
		return false
	}
	outer := internalUtils.OutermostTransparentExpression(node)
	if outer == nil || outer.Parent == nil {
		return false
	}
	if isInsideTryWithCatchOrAbruptFinally(outer) {
		return false
	}
	// An unhandled `Promise.reject(...)` is only safe to await directly as a
	// standalone statement or return value.
	if u := internalUtils.SkipAssertionsAndParens(outer); u != nil {
		if m, ok := r.promiseMethodName(u); ok && m == "reject" {
			return (outer.Parent.Kind == ast.KindExpressionStatement && outer.Parent.AsExpressionStatement().Expression == outer) ||
				(outer.Parent.Kind == ast.KindReturnStatement && outer.Parent.AsReturnStatement().Expression == outer)
		}
	}
	return r.isSafeAwaitPosition(outer, outer, false)
}

func (r *AsyncAssertionResolver) isSafeAwaitPosition(reportOuter *ast.Node, cur *ast.Node, wrappedValue bool) bool {
	if cur == nil || cur.Parent == nil {
		return false
	}
	parent := cur.Parent
	switch parent.Kind {
	case ast.KindExpressionStatement:
		return parent.AsExpressionStatement().Expression == cur
	case ast.KindReturnStatement:
		return !wrappedValue && parent.AsReturnStatement().Expression == cur
	case ast.KindAwaitExpression:
		return parent.AsAwaitExpression().Expression == cur &&
			r.isSafeAwaitPosition(reportOuter, internalUtils.OutermostTransparentExpression(parent), wrappedValue)
	case ast.KindConditionalExpression:
		cond := parent.AsConditionalExpression()
		if cond == nil || (cond.WhenTrue != cur && cond.WhenFalse != cur) {
			return false
		}
		return r.isSafeAwaitPosition(reportOuter, internalUtils.OutermostTransparentExpression(parent), wrappedValue)
	case ast.KindVariableDeclaration:
		vd := parent.AsVariableDeclaration()
		if vd == nil || vd.Initializer != cur || vd.Type != nil {
			return false
		}
		name := vd.Name()
		if name == nil || name.Kind != ast.KindIdentifier || r.isBindingReadOrExported(parent, name) {
			return false
		}
		declList := parent.Parent
		if declList == nil || declList.Kind != ast.KindVariableDeclarationList ||
			ast.IsVarUsing(declList) || ast.IsVarAwaitUsing(declList) ||
			declList.AsVariableDeclarationList().Declarations == nil {
			return false
		}
		decls := declList.AsVariableDeclarationList().Declarations.Nodes
		if len(decls) == 0 || decls[len(decls)-1] != parent {
			return false
		}
		return declList.Parent != nil && declList.Parent.Kind == ast.KindVariableStatement
	case ast.KindCallExpression:
		args := parent.AsCallExpression().Arguments
		if args == nil || len(args.Nodes) != 1 || args.Nodes[0] != cur {
			return false
		}
		// Inserting `await` inside a call argument delays invoking the callee
		// and changes the argument value from a Promise to undefined. Restrict
		// to single-argument global `Promise.resolve([...])`, single-element
		// pure literal `Promise.<combinator>([foo(...)])`, and upstream Jest's
		// `Promise.all.x([...])` literal-expect arrays.
		if m, isPromise := r.promiseMethodName(parent); isPromise {
			switch m {
			case "resolve":
				if !isArrayLiteralExpr(cur) {
					return false
				}
				return r.isSafeAwaitPosition(reportOuter, internalUtils.OutermostTransparentExpression(parent), true)
			case "all", "allSettled", "race", "any":
				if !isSingleElementArrayLiteral(cur) || !r.isPureFollowingArrayExpect(reportOuter, false) {
					return false
				}
				return r.isSafeAwaitPosition(reportOuter, internalUtils.OutermostTransparentExpression(parent), true)
			default:
				if isNonCombinatorPromiseMethod(m) || !isArrayLiteralExpr(cur) || !r.isPureFollowingArrayExpect(reportOuter, false) {
					return false
				}
				return r.isSafeAwaitPosition(reportOuter, internalUtils.OutermostTransparentExpression(parent), true)
			}
		}
		if !r.isPureFollowingArrayExpect(reportOuter, false) {
			return false
		}
		if r.isGlobalPromiseMemberChainCall(parent) && isArrayLiteralExpr(cur) {
			return r.isSafeAwaitPosition(reportOuter, internalUtils.OutermostTransparentExpression(parent), true)
		}
		if cur == reportOuter && !r.isLocallyDeclaredCallee(parent) && r.isInsideGlobalPromiseCombinatorArray(parent) {
			return r.isSafeAwaitPosition(reportOuter, internalUtils.OutermostTransparentExpression(parent), true)
		}
		return false
	case ast.KindArrayLiteralExpression:
		arr := parent.AsArrayLiteralExpression()
		if arr == nil || arr.Elements == nil {
			return false
		}
		info := r.getArrayPureSuffixInfo(parent)
		idx, ok := info.elemIndex[cur]
		if !ok {
			return false
		}
		if idx < len(arr.Elements.Nodes)-1 {
			if cur == reportOuter {
				if idx < info.minPureIdx {
					return false
				}
			} else if idx+1 < info.minPureIdx || !r.isPureFollowingArrayExpect(reportOuter, info.allowStandaloneBare) {
				return false
			}
		}
		return r.isSafeAwaitPosition(reportOuter, internalUtils.OutermostTransparentExpression(parent), true)
	default:
		return false
	}
}

func isTerminalFunctionReturn(ret *ast.Node) bool {
	if ret == nil || ret.Kind != ast.KindReturnStatement || ret.Parent == nil || ret.Parent.Kind != ast.KindBlock {
		return false
	}
	block := ret.Parent
	if block.Parent == nil || !testFramework.IsFunction(block.Parent) {
		return false
	}
	stmts := block.AsBlock().Statements
	return stmts != nil && len(stmts.Nodes) > 0 && stmts.Nodes[len(stmts.Nodes)-1] == ret
}

func (r *AsyncAssertionResolver) isSafeTerminalReturnStrip(ret *ast.Node, expr *ast.Node) bool {
	if !isTerminalFunctionReturn(ret) {
		return false
	}
	fn := ret.Parent.Parent
	if fn == nil || fn.Type() != nil {
		return false
	}
	if fn.Kind != ast.KindArrowFunction && fn.Kind != ast.KindFunctionExpression {
		return false
	}
	fnOuter := internalUtils.OutermostTransparentExpression(fn)
	if fnOuter == nil || fnOuter.Parent == nil ||
		fnOuter.Parent.Kind != ast.KindCallExpression ||
		fnOuter.Parent.AsCallExpression().Expression == fnOuter {
		return false
	}
	// Preserve explicit type assertions on returned expressions (`return expect(...) as Promise<T>`)
	// so stripping `return` cannot alter the callback's inferred return type.
	for cur := expr; cur != nil; {
		if cur.Kind == ast.KindParenthesizedExpression {
			cur = cur.AsParenthesizedExpression().Expression
			continue
		}
		if cur.Kind == ast.KindAsExpression || cur.Kind == ast.KindTypeAssertionExpression {
			return false
		}
		if inner, ok := internalUtils.TransparentExpression(cur); ok {
			cur = inner
			continue
		}
		break
	}
	unwrapped := internalUtils.SkipAssertionsAndParens(expr)
	if _, isPromiseWrapper := r.promiseMethodName(unwrapped); isPromiseWrapper {
		return false
	}
	if _, isChain := r.promiseChainMethodName(unwrapped); isChain {
		return false
	}
	return true
}

func insertAwaitBefore(sourceFile *ast.SourceFile, node *ast.Node) rule.RuleFix {
	text := sourceFile.Text()
	pos := internalUtils.TrimNodeTextRange(sourceFile, node).Pos()
	insertion := "await "
	if needsLeadingSpace(text, pos) {
		insertion = " await "
	}
	return rule.RuleFixReplaceRange(core.NewTextRange(pos, pos), insertion)
}

func wrapAwaitWithParens(sourceFile *ast.SourceFile, node *ast.Node) rule.RuleFix {
	text := sourceFile.Text()
	r := internalUtils.TrimNodeTextRange(sourceFile, node)
	prefix := "await ("
	if needsLeadingSpace(text, r.Pos()) {
		prefix = " await ("
	}
	return rule.RuleFixReplaceRange(r, prefix+text[r.Pos():r.End()]+")")
}

// AwaitFix awaits an async assertion, replacing a terminal `return` in a void
// callback with `await` when returning it is not accepted. Early returns and
// value-returning helpers keep `return await` to preserve control flow and
// return values.
func (r *AsyncAssertionResolver) AwaitFix(sourceFile *ast.SourceFile, node *ast.Node, alwaysAwait bool) rule.RuleFix {
	outer := internalUtils.OutermostTransparentExpression(node)
	if alwaysAwait && outer.Parent != nil && outer.Parent.Kind == ast.KindReturnStatement {
		ret := outer.Parent
		retRange := internalUtils.TrimNodeTextRange(sourceFile, ret)
		outerRange := internalUtils.TrimNodeTextRange(sourceFile, outer)
		if r.isSafeTerminalReturnStrip(ret, outer) {
			text := sourceFile.Text()
			retKeywordEnd := retRange.Pos() + len("return")
			between := ""
			if retKeywordEnd <= outerRange.Pos() && outerRange.Pos() <= len(text) {
				between = text[retKeywordEnd:outerRange.Pos()]
			}
			if between == "" {
				between = " "
			}
			if outer.Kind == ast.KindAsExpression || outer.Kind == ast.KindSatisfiesExpression {
				return rule.RuleFixReplaceRange(
					core.NewTextRange(retRange.Pos(), outerRange.End()),
					"await"+between+"("+text[outerRange.Pos():outerRange.End()]+")",
				)
			}
			return rule.RuleFixReplaceRange(core.NewTextRange(retRange.Pos(), outerRange.Pos()), "await"+between)
		}
		if outer.Kind == ast.KindAsExpression || outer.Kind == ast.KindSatisfiesExpression {
			return wrapAwaitWithParens(sourceFile, outer)
		}
		return insertAwaitBefore(sourceFile, outer)
	}
	if outer != internalUtils.OutermostParenthesizedExpression(node) {
		if outer.Kind == ast.KindAsExpression || outer.Kind == ast.KindSatisfiesExpression {
			return wrapAwaitWithParens(sourceFile, outer)
		}
		return insertAwaitBefore(sourceFile, outer)
	}
	return insertAwaitBefore(sourceFile, node)
}
