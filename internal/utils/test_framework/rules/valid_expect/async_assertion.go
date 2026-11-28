// Package valid_expect holds the framework-neutral async-assertion machinery
// shared by jest/valid-expect and rstest/valid-expect: walking out of a chained
// .then/.catch, spotting a Promise.all([...]) wrapper, deciding whether the
// result is awaited or returned, and the async/await fixes.
//
// Only the parts whose correctness does not depend on the framework live here.
// The rules keep their own bodies: how a broken chain's reason is derived, what
// counts as an excess expect argument and which matcher shapes exist all differ
// between jest and rstest, and merging those would mean framework branches
// inside a shared body.
package valid_expect

import (
	"fmt"
	"math/big"
	"strconv"
	"unicode"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type blockStatementIndex struct {
	stmts      []*ast.Node
	stmtIndex  map[*ast.Node]int
	exitPrefix []int
}

type blockExitSummary struct {
	unconditionalExit bool
	unlabeledBreak    bool
	unlabeledContinue bool
	escapedLabels     []string
}

type promiseTryArgPos struct {
	runtimeIdx       int  // 0-based runtime argument index in Promise.try (0 = callback, >= 1 = callback arg)
	hasUnknownSpread bool // true if a non-statically-sized spread precedes this arg
}

type promiseTryCallIndex struct {
	cb                    *ast.Node // resolved runtime arg 0 callback if statically known
	cbUnknown             bool      // true if runtime arg 0 may come from an unknown spread
	cbMissing             bool      // true if runtime arg 0 is omitted (e.g. `...[, cb]`)
	firstUnknownSpreadIdx int       // first 1-based callback arg index at/after an unknown spread (0 if none)
	args                  map[*ast.Node]promiseTryArgPos
	argExprByIdx          map[int]*ast.Node // 1-based callback arg index -> expr when no unknown spread precedes it
}

type inlineCallbackSummary struct {
	usesArguments     bool
	bodyCanConsume    bool
	hasDefaultParam   bool
	restParamIdx      int // -1 if no rest parameter
	params            []*ast.Node
	restWholeConsumed map[*ast.Node]bool
}

type scopeDepKind uint8

const (
	scopeDepVarDecl scopeDepKind = iota
	scopeDepAssign
	scopeDepContainerPush
	scopeDepContainerIndexWrite
	scopeDepMapSet
)

type scopeDepEdge struct {
	kind      scopeDepKind
	target    *ast.Node
	rhs       *ast.Node
	pos       int
	topLevel  bool
	outerLoop *ast.Node
}

type containerRemoval struct {
	startPos  int
	pos       int
	outerLoop *ast.Node
}

type mapKeySetEvent struct {
	key       string
	startPos  int
	pos       int
	outerLoop *ast.Node
}

type scopeConsumeKind uint8

const (
	scopeConsumeAwait scopeConsumeKind = iota
	scopeConsumeReturn
	scopeConsumeForAwait
)

type scopeConsumeSite struct {
	kind scopeConsumeKind
	node *ast.Node
	expr *ast.Node
}

// scopeBodyIndex pre-indexes local alias/container edges and await/return/for-await
// sites once per callback or catch block so per-argument queries only touch
// statements referencing tracked bindings (O(S + A) instead of O(A * S)).
type scopeBodyIndex struct {
	deps                  []scopeDepEdge
	depsBySym             map[*ast.Symbol][]int
	depsByName            map[string][]int
	consumes              []scopeConsumeSite
	consumesBySym         map[*ast.Symbol][]int
	consumesByName        map[string][]int
	removedContainers     map[*ast.Symbol][]containerRemoval
	removedContainerNames map[string][]containerRemoval
	shiftedContainers     map[*ast.Symbol][]containerRemoval
	shiftedContainerNames map[string][]containerRemoval
	pushedContainers      map[*ast.Symbol][]containerRemoval
	pushedContainerNames  map[string][]containerRemoval
	mapKeySets            map[*ast.Symbol][]mapKeySetEvent
	mapNameKeySets        map[string][]mapKeySetEvent
	reassignedContainers  map[*ast.Symbol]bool
	reassignedNames       map[string]bool
	topLevelReassigns     map[*ast.Symbol][]int
	topLevelReassignNames map[string][]int
}

type arrayLiteralSlotIndex struct {
	bySym               map[*ast.Symbol][]int
	byName              map[string][]int
	nonIdentifierSlices []int
	firstSpreadIdx      int // -1 if no spread element
	elemCount           int
}

type arrayLiteralIndex struct {
	elemIndex          map[*ast.Node]int
	firstSpreadIdx     int  // -1 if no spread element
	lastSideEffectInit bool // lazily computed only when autofix checks siblings
	lastSideEffectIdx  int  // -1 if no side-effecting element
}

type callArgsIndex struct {
	argIndex          map[*ast.Node]int
	lastSideEffectIdx int // -1 if no side-effecting argument
}

type nodeListSideEffectIndex struct {
	nodeIndex         map[*ast.Node]int
	lastSideEffectIdx int // -1 if no side-effecting item
}

// AsyncAssertionResolver caches per-file shadowing and static-string lookups
// for valid-expect rules so repeated `Promise.all` assertions stay linear-time,
// `const Promise = globalThis.Promise` aliases resolve accurately, and computed
// `Promise[key]` method names fold through `StaticStringEvaluator`.
type AsyncAssertionResolver struct {
	program                        *program.Program
	sourceFile                     *ast.SourceFile
	typeChecker                    *checker.Checker
	refs                           *rule.RefStore
	shadowCache                    *internalUtils.ShadowCache
	staticStrings                  *internalUtils.StaticStringEvaluator
	blockIndices                   map[*ast.Node]*blockStatementIndex
	blockExits                     map[*ast.Node]blockExitSummary
	bindingObserved                map[*ast.Symbol]bool
	tryCallIndices                 map[*ast.Node]*promiseTryCallIndex
	callbackInfos                  map[*ast.Node]*inlineCallbackSummary
	scopeBodyIndices               map[*ast.Node]*scopeBodyIndex
	arrayElemIndices               map[*ast.Node]*arrayLiteralIndex
	arraySlotIndices               map[*ast.Node]*arrayLiteralSlotIndex
	pureLiteralExpectArrays        map[*ast.Node]bool
	callArgIndices                 map[*ast.Node]*callArgsIndex
	exprMutates                    map[*ast.Node]bool
	listSideEffectIndices          map[*ast.Node]*nodeListSideEffectIndex
	fnMutatesState                 map[*ast.Node]bool
	arrayRestIndices               map[*ast.Node]int
	bindingWritten                 map[*ast.Symbol]bool
	objMutated                     map[*ast.Symbol]bool
	disallowBuiltinToResolveReject bool
	checkedFileExtend              bool
	fileHasExtend                  bool
}

// NewAsyncAssertionResolver creates a per-file resolver backed by ctx.
func NewAsyncAssertionResolver(ctx rule.RuleContext) *AsyncAssertionResolver {
	return &AsyncAssertionResolver{
		program:     ctx.Program(),
		sourceFile:  ctx.SourceFile,
		typeChecker: ctx.TypeChecker,
		refs:        ctx.Refs,
	}
}

// DisableBuiltinToResolveReject marks `toResolve`/`toReject` as non-builtin
// matchers (e.g. in Rstest) so autofix never treats them as side-effect-free.
func (r *AsyncAssertionResolver) DisableBuiltinToResolveReject() {
	if r != nil {
		r.disallowBuiltinToResolveReject = true
	}
}

func (r *AsyncAssertionResolver) isShadowed(node *ast.Node, name string) bool {
	if r != nil && r.sourceFile != nil {
		if r.shadowCache == nil {
			r.shadowCache = internalUtils.NewShadowCache(r.sourceFile)
		}
		return r.shadowCache.IsShadowed(node, name)
	}
	return internalUtils.IsShadowed(node, name)
}

func (r *AsyncAssertionResolver) staticStringEvaluator() *internalUtils.StaticStringEvaluator {
	if r == nil {
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
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return nil
	}
	if r != nil && r.refs != nil {
		if sym := r.refs.Resolve(ident); sym != nil {
			return sym
		}
		if r.typeChecker != nil {
			if sym := internalUtils.GetReferenceSymbol(ident, r.typeChecker); sym != nil {
				return sym
			}
		}
		return nil
	}
	if r != nil && r.typeChecker != nil {
		if sym := internalUtils.GetReferenceSymbol(ident, r.typeChecker); sym != nil {
			return sym
		}
		return nil
	}
	// Fallback ONLY for standalone contexts without Refs or TypeChecker:
	// consult binder-owned Locals while skipping parameter/computed-key/decorator
	// scopes that cannot observe the enclosing member's body locals.
	name := ident.AsIdentifier().Text
	prev := ident
	for current := ident.Parent; current != nil; prev, current = current, current.Parent {
		if prev != nil && (prev.Kind == ast.KindDecorator || prev == current.Name()) {
			continue
		}
		if prev != nil && prev.Kind == ast.KindParameter && testFramework.IsFunction(current) {
			// A parameter default initializer cannot see function-body `var` locals.
			continue
		}
		if current.Kind == ast.KindFunctionExpression || current.Kind == ast.KindClassExpression {
			if exprName := current.Name(); exprName != nil && exprName.Kind == ast.KindIdentifier && exprName.AsIdentifier().Text == name {
				return nil
			}
		}
		if !ast.IsLocalsContainer(current) {
			continue
		}
		if local := ast.GetLocals(current)[name]; local != nil &&
			local.Flags&(ast.SymbolFlagsValue|ast.SymbolFlagsAlias) != 0 &&
			hasValueDeclaration(local) {
			return local
		}
	}
	return nil
}

func (r *AsyncAssertionResolver) hasSameFileValueDeclaration(sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	for _, decl := range sym.Declarations {
		if !isValueDeclarationNode(decl) {
			continue
		}
		if r != nil && r.sourceFile != nil && ast.GetSourceFileOfNode(decl) != r.sourceFile {
			continue
		}
		return true
	}
	return false
}

func (r *AsyncAssertionResolver) isLocallyValueShadowed(node *ast.Node, name string) bool {
	if r != nil && r.refs != nil {
		sym := r.refs.Resolve(node)
		if sym == nil {
			return false
		}
		return r.hasSameFileValueDeclaration(sym)
	}
	return r.isShadowed(node, name)
}

func isValueDeclarationNode(decl *ast.Node) bool {
	if decl == nil {
		return false
	}
	switch decl.Kind {
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindTypeParameter:
		return false
	case ast.KindFunctionDeclaration:
		return decl.AsFunctionDeclaration().Body != nil
	case ast.KindImportEqualsDeclaration:
		return !decl.AsImportEqualsDeclaration().IsTypeOnly
	case ast.KindImportClause:
		return !decl.AsImportClause().IsTypeOnly()
	case ast.KindImportSpecifier:
		if decl.AsImportSpecifier().IsTypeOnly {
			return false
		}
		if decl.Parent != nil && decl.Parent.Parent != nil && decl.Parent.Parent.Kind == ast.KindImportClause {
			return !decl.Parent.Parent.AsImportClause().IsTypeOnly()
		}
		return true
	case ast.KindNamespaceImport:
		if decl.Parent != nil && decl.Parent.Kind == ast.KindImportClause {
			return !decl.Parent.AsImportClause().IsTypeOnly()
		}
		return true
	default:
		return true
	}
}

func hasValueDeclaration(sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	if len(sym.Declarations) == 0 {
		return sym.Flags&ast.SymbolFlagsValue != 0
	}
	for _, decl := range sym.Declarations {
		if isValueDeclarationNode(decl) {
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) singleSameFileValueDeclaration(sym *ast.Symbol) *ast.Node {
	if sym == nil {
		return nil
	}
	var valueDecl *ast.Node
	for _, decl := range sym.Declarations {
		if !isValueDeclarationNode(decl) {
			continue
		}
		if valueDecl != nil {
			return nil
		}
		valueDecl = decl
	}
	if valueDecl == nil {
		return nil
	}
	if r != nil && r.sourceFile != nil && ast.GetSourceFileOfNode(valueDecl) != r.sourceFile {
		return nil
	}
	return valueDecl
}

func nodeContains(ancestor, descendant *ast.Node) bool {
	if ancestor == nil || descendant == nil {
		return false
	}
	return descendant.Pos() >= ancestor.Pos() && descendant.End() <= ancestor.End()
}

func (r *AsyncAssertionResolver) isGlobalThisObject(node *ast.Node, depth int) bool {
	if depth > 4 {
		return false
	}
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	if node.AsIdentifier().Text == "globalThis" && !r.isLocallyValueShadowed(node, "globalThis") {
		return true
	}
	decl := r.singleSameFileValueDeclaration(r.resolveSymbol(node))
	if decl != nil && decl.Kind == ast.KindVariableDeclaration && ast.IsVarConst(decl) {
		init := decl.AsVariableDeclaration().Initializer
		if init != nil && !nodeContains(init, node) {
			return r.isGlobalThisObject(init, depth+1)
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isGlobalPromiseObject(node *ast.Node, depth int) bool {
	if depth > 4 {
		return false
	}
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindIdentifier {
		if node.AsIdentifier().Text == "Promise" && !r.isLocallyValueShadowed(node, "Promise") {
			return true
		}
		decl := r.singleSameFileValueDeclaration(r.resolveSymbol(node))
		if decl == nil {
			return false
		}
		switch decl.Kind {
		case ast.KindVariableDeclaration:
			if !ast.IsVarConst(decl) {
				return false
			}
			init := decl.AsVariableDeclaration().Initializer
			if init != nil && !nodeContains(init, node) {
				return r.isGlobalPromiseObject(init, depth+1)
			}
		case ast.KindBindingElement:
			b := decl.AsBindingElement()
			if b == nil || b.DotDotDotToken != nil {
				return false
			}
			vd := internalUtils.EnclosingVariableDeclarationOfBindingElement(decl)
			if vd == nil || !ast.IsVarConst(vd) ||
				decl.Parent == nil || decl.Parent.Kind != ast.KindObjectBindingPattern ||
				decl.Parent != vd.Name() {
				return false
			}
			propName := ""
			if b.PropertyName != nil {
				if name, ok := internalUtils.GetStaticPropertyName(b.PropertyName); ok {
					propName = name
				}
			} else if b.Name() != nil && b.Name().Kind == ast.KindIdentifier {
				propName = b.Name().AsIdentifier().Text
			}
			if propName == "Promise" {
				init := vd.AsVariableDeclaration().Initializer
				if init != nil && !nodeContains(init, node) {
					return r.isGlobalThisObject(init, depth+1)
				}
			}
		}
		return false
	}
	if ast.IsAccessExpression(node) {
		name, ok := r.accessStaticNameThroughAssertions(node)
		if !ok || name != "Promise" {
			return false
		}
		return r.isGlobalThisObject(internalUtils.AccessExpressionObject(node), depth+1)
	}
	return false
}

func accessStaticNameThroughAssertions(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	if name, ok := internalUtils.AccessExpressionStaticName(node); ok {
		return name, true
	}
	if node.Kind == ast.KindElementAccessExpression {
		arg := internalUtils.SkipAssertionsAndParens(node.AsElementAccessExpression().ArgumentExpression)
		return internalUtils.GetStaticStringLiteralValue(arg)
	}
	return "", false
}

func (r *AsyncAssertionResolver) accessStaticNameThroughAssertions(node *ast.Node) (string, bool) {
	if name, ok := accessStaticNameThroughAssertions(node); ok {
		return name, true
	}
	if node != nil && node.Kind == ast.KindElementAccessExpression {
		if eval := r.staticStringEvaluator(); eval != nil {
			arg := internalUtils.SkipAssertionsAndParens(node.AsElementAccessExpression().ArgumentExpression)
			if name, ok := eval.EvalToString(arg); ok {
				return name, true
			}
		}
	}
	return "", false
}

func (r *AsyncAssertionResolver) promiseMethodName(node *ast.Node) (string, bool) {
	if node == nil || node.Kind != ast.KindCallExpression {
		return "", false
	}

	callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
	if !ast.IsAccessExpression(callee) {
		return "", false
	}

	if !r.isGlobalPromiseObject(internalUtils.AccessExpressionObject(callee), 0) {
		return "", false
	}

	// Fold statically known method names (dot access, string/template literals,
	// or static string constants such as `const resolve = 'all'` vs
	// `const resolve = 'resolve'`). Unresolvable dynamic keys return ("", true).
	method, _ := r.accessStaticNameThroughAssertions(callee)
	return method, true
}

func runtimeParameters(cb *ast.Node) []*ast.Node {
	if cb == nil {
		return nil
	}
	params := cb.Parameters()
	if len(params) > 0 && params[0] != nil && ast.IsThisParameter(params[0]) {
		return params[1:]
	}
	return params
}

func (r *AsyncAssertionResolver) getPromiseTryCallIndex(call *ast.Node) *promiseTryCallIndex {
	if r != nil && r.tryCallIndices != nil {
		if idx, ok := r.tryCallIndices[call]; ok {
			return idx
		}
	}
	args := call.AsCallExpression().Arguments
	idx := &promiseTryCallIndex{}
	if args != nil && len(args.Nodes) > 0 {
		idx.args = make(map[*ast.Node]promiseTryArgPos, len(args.Nodes))
		idx.argExprByIdx = make(map[int]*ast.Node, len(args.Nodes))
		runtimeIdx := 0
		hasUnknownSpread := false
		var walkArgList func(nodes []*ast.Node, depth int)
		walkArgList = func(nodes []*ast.Node, depth int) {
			for _, a := range nodes {
				if a == nil || a.Kind == ast.KindOmittedExpression {
					if !hasUnknownSpread && runtimeIdx == 0 {
						idx.cbMissing = true
					}
					runtimeIdx++
					continue
				}
				if a.Kind == ast.KindSpreadElement {
					spreadExpr := internalUtils.SkipAssertionsAndParens(a.AsSpreadElement().Expression)
					if depth < 4 && spreadExpr != nil && spreadExpr.Kind == ast.KindArrayLiteralExpression {
						if elems := spreadExpr.AsArrayLiteralExpression().Elements; elems != nil {
							walkArgList(elems.Nodes, depth+1)
						}
						continue
					}
					if !hasUnknownSpread {
						if runtimeIdx == 0 && idx.cb == nil && !idx.cbMissing {
							idx.cbUnknown = true
						}
						if runtimeIdx >= 1 {
							idx.firstUnknownSpreadIdx = runtimeIdx
						} else {
							idx.firstUnknownSpreadIdx = 1
						}
						hasUnknownSpread = true
					}
					continue
				}
				if !hasUnknownSpread {
					if runtimeIdx == 0 {
						idx.cb = internalUtils.SkipAssertionsAndParens(a)
					} else {
						idx.argExprByIdx[runtimeIdx] = a
					}
				}
				idx.args[a] = promiseTryArgPos{
					runtimeIdx:       runtimeIdx,
					hasUnknownSpread: hasUnknownSpread,
				}
				runtimeIdx++
			}
		}
		walkArgList(args.Nodes, 0)
	}
	if r != nil {
		if r.tryCallIndices == nil {
			r.tryCallIndices = map[*ast.Node]*promiseTryCallIndex{}
		}
		r.tryCallIndices[call] = idx
	}
	return idx
}

func (r *AsyncAssertionResolver) getInlineCallbackSummary(cb *ast.Node) *inlineCallbackSummary {
	if r != nil && r.callbackInfos != nil {
		if info, ok := r.callbackInfos[cb]; ok {
			return info
		}
	}
	params := runtimeParameters(cb)
	restParamIdx := -1
	hasDefaultParam := false
	for i, param := range params {
		if param == nil {
			continue
		}
		if p := param.AsParameterDeclaration(); p != nil {
			if p.DotDotDotToken != nil && restParamIdx < 0 {
				restParamIdx = i
			}
			if p.Initializer != nil || (p.Name() != nil && (p.Name().Kind == ast.KindObjectBindingPattern || p.Name().Kind == ast.KindArrayBindingPattern)) {
				hasDefaultParam = true
			}
		}
	}
	info := &inlineCallbackSummary{
		usesArguments:   cb != nil && cb.Kind == ast.KindFunctionExpression && functionBodyUsesArguments(cb),
		bodyCanConsume:  inlineCallbackBodyCanConsume(cb),
		hasDefaultParam: hasDefaultParam,
		restParamIdx:    restParamIdx,
		params:          params,
	}
	if r != nil {
		if r.callbackInfos == nil {
			r.callbackInfos = map[*ast.Node]*inlineCallbackSummary{}
		}
		r.callbackInfos[cb] = info
	}
	return info
}

func isIdentifierValueReference(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindIdentifier {
		return false
	}
	if ast.IsPartOfTypeNode(node) || ast.IsPartOfTypeQuery(node) {
		return false
	}
	parent := node.Parent
	if parent == nil {
		return true
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		return parent.AsPropertyAccessExpression().Name() != node
	case ast.KindPropertyAssignment, ast.KindMethodDeclaration, ast.KindPropertyDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor, ast.KindMethodSignature,
		ast.KindPropertySignature, ast.KindEnumMember, ast.KindNamedTupleMember,
		ast.KindJsxAttribute:
		return parent.Name() != node
	case ast.KindBindingElement:
		be := parent.AsBindingElement()
		return be == nil || (be.PropertyName != node && be.Name() != node)
	case ast.KindVariableDeclaration, ast.KindParameter,
		ast.KindFunctionDeclaration, ast.KindFunctionExpression,
		ast.KindClassDeclaration, ast.KindClassExpression,
		ast.KindEnumDeclaration, ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration, ast.KindTypeParameter,
		ast.KindModuleDeclaration, ast.KindImportClause,
		ast.KindImportSpecifier, ast.KindNamespaceImport:
		return parent.Name() != node
	case ast.KindLabeledStatement:
		return parent.AsLabeledStatement().Label != node
	case ast.KindBreakStatement:
		return parent.AsBreakStatement().Label != node
	case ast.KindContinueStatement:
		return parent.AsContinueStatement().Label != node
	case ast.KindExportSpecifier:
		es := parent.AsExportSpecifier()
		return es == nil || (es.PropertyName != node && es.Name() != node)
	case ast.KindQualifiedName, ast.KindTypeReference, ast.KindMetaProperty:
		return false
	}
	return true
}

func functionBodyUsesArguments(fn *ast.Node) bool {
	found := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || found {
			return
		}
		if ast.IsTypeNode(node) {
			return
		}
		if node != fn {
			switch node.Kind {
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindClassStaticBlockDeclaration:
				return
			case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
				// Decorators and computed property keys evaluate in `fn`'s scope.
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
		if node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == "arguments" && isIdentifierValueReference(node) {
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

func inlineCallbackBodyCanConsume(cb *ast.Node) bool {
	if cb == nil || functionAsteriskToken(cb) != nil {
		// Generator callbacks (`function*` / `async function*`) return a
		// suspended generator without running their body.
		return false
	}
	body := cb.Body()
	if body == nil {
		return false
	}
	if body.Kind != ast.KindBlock {
		// Concise arrow expression body implicitly returns its result.
		return true
	}
	consumes := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || consumes {
			return
		}
		if node != body {
			switch node.Kind {
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
				ast.KindClassStaticBlockDeclaration:
				return
			case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
				// Decorators and computed property keys evaluate in `cb`'s own scope.
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
			}
		}
		switch node.Kind {
		case ast.KindAwaitExpression:
			consumes = true
			return
		case ast.KindForOfStatement:
			if forOf := node.AsForInOrOfStatement(); forOf != nil && forOf.AwaitModifier != nil {
				consumes = true
				return
			}
		case ast.KindReturnStatement:
			if ret := node.AsReturnStatement(); ret != nil && ret.Expression != nil {
				consumes = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return consumes
		})
	}
	visit(body)
	return consumes
}

func collectBindingIdentifiers(pattern *ast.Node) []*ast.Node {
	if pattern == nil {
		return nil
	}
	var out []*ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		switch node.Kind {
		case ast.KindIdentifier:
			out = append(out, node)
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			if bp := node.AsBindingPattern(); bp != nil && bp.Elements != nil {
				for _, el := range bp.Elements.Nodes {
					if el != nil && el.Kind == ast.KindBindingElement {
						visit(el.Name())
					}
				}
			}
		}
	}
	visit(pattern)
	return out
}

type containerWrite struct {
	scope          *ast.Node
	pos            int
	endPos         int
	topLevel       bool
	outerLoop      *ast.Node
	isMutation     bool
	isLastPushElem bool
	mapStaticKeys  map[string]bool
}

type containerPeerLink struct {
	sym      *ast.Symbol
	scope    *ast.Node
	pos      int
	topLevel bool
	isDest   bool
}

type containerNamePeerLink struct {
	name     string
	scope    *ast.Node
	pos      int
	topLevel bool
	isDest   bool
}

type trackedBindings struct {
	symbols               map[*ast.Symbol]bool
	names                 map[string]bool
	restSymbols           map[*ast.Symbol]bool
	restNames             map[string]bool
	containerSymbols      map[*ast.Symbol]bool
	containerNames        map[string]bool
	mapSymbols            map[*ast.Symbol]bool
	mapNames              map[string]bool
	arraySlotSymbols      map[*ast.Symbol]map[int]bool
	arraySlotNames        map[string]map[int]bool
	arraySlotDeclPos      map[*ast.Symbol]int
	arraySlotDeclNamePos  map[string]int
	arraySlotLen          map[*ast.Symbol]int
	arraySlotNameLen      map[string]int
	containerScopes       []*ast.Node
	containerWrites       map[*ast.Symbol][]containerWrite
	containerNameWrites   map[string][]containerWrite
	initialContainers     map[*ast.Symbol]bool
	initialContainerNames map[string]bool
	containerPeers        map[*ast.Symbol][]containerPeerLink
	containerNamePeers    map[string][]containerNamePeerLink
	argumentsAliasSymbols map[*ast.Symbol]bool
	argumentsAliasNames   map[string]bool
	restOffset            int // >= 0 exact expected 0-based index into rest/`arguments`, -1 if unknown
	trackArguments        bool
}

func (tb *trackedBindings) isEmpty() bool {
	return len(tb.symbols) == 0 && len(tb.names) == 0 &&
		len(tb.restSymbols) == 0 && len(tb.restNames) == 0 &&
		len(tb.containerSymbols) == 0 && len(tb.containerNames) == 0 &&
		len(tb.mapSymbols) == 0 && len(tb.mapNames) == 0 &&
		!tb.trackArguments
}

type trackedBindingKind uint8

const (
	trackedBindingDirect trackedBindingKind = iota
	trackedBindingRest
	trackedBindingContainer
	trackedBindingMap
)

func (r *AsyncAssertionResolver) addTrackedIdentKind(
	tb *trackedBindings,
	ident *ast.Node,
	kind trackedBindingKind,
) (sym *ast.Symbol, name string, added bool) {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return nil, "", false
	}
	name = ident.AsIdentifier().Text
	if ident.Parent != nil && ident.Parent.Symbol() != nil {
		sym = ident.Parent.Symbol()
	} else if r != nil {
		sym = r.resolveSymbol(ident)
	}
	if r != nil && r.refs != nil && sym != nil {
		switch kind {
		case trackedBindingRest:
			if tb.restSymbols[sym] {
				return sym, name, false
			}
			if tb.restSymbols == nil {
				tb.restSymbols = map[*ast.Symbol]bool{}
			}
			tb.restSymbols[sym] = true
			return sym, name, true
		case trackedBindingContainer:
			if tb.containerSymbols[sym] {
				return sym, name, false
			}
			if tb.containerSymbols == nil {
				tb.containerSymbols = map[*ast.Symbol]bool{}
			}
			tb.containerSymbols[sym] = true
			return sym, name, true
		case trackedBindingMap:
			if tb.mapSymbols[sym] {
				return sym, name, false
			}
			if tb.mapSymbols == nil {
				tb.mapSymbols = map[*ast.Symbol]bool{}
			}
			tb.mapSymbols[sym] = true
			return sym, name, true
		default:
			if tb.symbols[sym] {
				return sym, name, false
			}
			if tb.symbols == nil {
				tb.symbols = map[*ast.Symbol]bool{}
			}
			tb.symbols[sym] = true
			return sym, name, true
		}
	}
	switch kind {
	case trackedBindingRest:
		if tb.restNames[name] {
			return nil, name, false
		}
		if tb.restNames == nil {
			tb.restNames = map[string]bool{}
		}
		tb.restNames[name] = true
		return nil, name, true
	case trackedBindingContainer:
		if tb.containerNames[name] {
			return nil, name, false
		}
		if tb.containerNames == nil {
			tb.containerNames = map[string]bool{}
		}
		tb.containerNames[name] = true
		return nil, name, true
	case trackedBindingMap:
		if tb.mapNames[name] {
			return nil, name, false
		}
		if tb.mapNames == nil {
			tb.mapNames = map[string]bool{}
		}
		tb.mapNames[name] = true
		return nil, name, true
	default:
		if tb.names[name] {
			return nil, name, false
		}
		if tb.names == nil {
			tb.names = map[string]bool{}
		}
		tb.names[name] = true
		return nil, name, true
	}
}

func (r *AsyncAssertionResolver) addTrackedIdent(tb *trackedBindings, ident *ast.Node, isRest bool) {
	if isRest {
		r.addTrackedIdentKind(tb, ident, trackedBindingRest)
		return
	}
	r.addTrackedIdentKind(tb, ident, trackedBindingDirect)
}

func (r *AsyncAssertionResolver) matchesTrackedDirectIdent(node *ast.Node, tb *trackedBindings, shadowed map[string]bool) bool {
	if node == nil || node.Kind != ast.KindIdentifier || !isIdentifierValueReference(node) {
		return false
	}
	name := node.AsIdentifier().Text
	if r != nil && r.refs != nil && len(tb.symbols) > 0 {
		if sym := r.resolveSymbol(node); sym != nil {
			return tb.symbols[sym]
		}
		if len(tb.names) > 0 && tb.names[name] && !shadowed[name] {
			return true
		}
		return false
	}
	return tb.names[name] && !shadowed[name]
}

type peerSymbolState struct {
	sym      *ast.Symbol
	tPos     int  // position at which `sym` holds the same object as `start` (-1 = unknown)
	minPos   int  // earliest position at which `sym` holds it (-1 = unbounded before tPos)
	fromDest bool // true once reached via a downstream (`src -> dst`) alias edge
}

type peerNameState struct {
	name     string
	tPos     int
	minPos   int
	fromDest bool
}

func hasTopLevelReassignBetween(positions []int, afterPos, beforePos int) bool {
	for _, rPos := range positions {
		if rPos > afterPos && rPos < beforePos {
			return true
		}
	}
	return false
}

func peerCanTransition(tPos, minPos, linkPos int, isDest bool, reassigns []int) bool {
	if tPos < 0 {
		return true
	}
	if !isDest {
		// Moving `dst -> src` across `dst = src` at `linkPos`: `dst` held `src`'s value from `linkPos` onward.
		if tPos < linkPos {
			return false
		}
		return !hasTopLevelReassignBetween(reassigns, linkPos, tPos)
	}
	// Moving `src -> dst` across `dst = src` at `linkPos`: `src` was read at `linkPos`.
	if minPos >= 0 && linkPos < minPos {
		return false
	}
	lo, hi := tPos, linkPos
	if lo > hi {
		lo, hi = hi, lo
	}
	return !hasTopLevelReassignBetween(reassigns, lo, hi)
}

func peerStateHoldsAt(tPos, minPos, eventPos int, reassigns []int) bool {
	if tPos < 0 {
		return true
	}
	if minPos >= 0 && eventPos < minPos {
		return false
	}
	if eventPos > tPos {
		return !hasTopLevelReassignBetween(reassigns, tPos, eventPos)
	}
	if eventPos < tPos {
		return !hasTopLevelReassignBetween(reassigns, eventPos, tPos)
	}
	return true
}

func peerStateWriteMatches(tPos, minPos int, fromDest bool, w containerWrite, reassigns []int) bool {
	if !w.isMutation && fromDest {
		return false
	}
	if tPos < 0 {
		return true
	}
	if !w.isMutation {
		if tPos < w.pos {
			return false
		}
		return !hasTopLevelReassignBetween(reassigns, w.pos, tPos)
	}
	return peerStateHoldsAt(tPos, minPos, w.pos, reassigns)
}

func eventInSameLoopAsUse(outerLoop *ast.Node, useNode *ast.Node) bool {
	return outerLoop != nil && useNode != nil && nodeContains(outerLoop, useNode)
}

func collectActivePeerSymbols(
	start *ast.Symbol,
	scope *ast.Node,
	usePos int,
	peers map[*ast.Symbol][]containerPeerLink,
	reassigns map[*ast.Symbol][]int,
) []peerSymbolState {
	if start == nil {
		return nil
	}
	initState := peerSymbolState{sym: start, tPos: usePos, minPos: -1}
	out := []peerSymbolState{initState}
	if len(peers) == 0 {
		return out
	}
	seen := map[peerSymbolState]bool{initState: true}
	const maxStates = 32
	for i := 0; i < len(out) && len(out) < maxStates; i++ {
		cur := out[i]
		for _, link := range peers[cur.sym] {
			if link.sym == nil {
				continue
			}
			// Never zigzag down (`src -> dst`) and then back up (`dst -> otherSrc`),
			// which would conflate two distinct upstream sources assigned to `dst`.
			if cur.fromDest && !link.isDest {
				continue
			}
			next := peerSymbolState{sym: link.sym, tPos: -1, minPos: -1, fromDest: cur.fromDest || link.isDest}
			if link.scope == scope && link.topLevel && cur.tPos >= 0 {
				if !peerCanTransition(cur.tPos, cur.minPos, link.pos, link.isDest, reassigns[cur.sym]) {
					continue
				}
				next.tPos = link.pos
				if link.isDest {
					next.minPos = link.pos
				}
			}
			if !seen[next] {
				seen[next] = true
				out = append(out, next)
				if len(out) >= maxStates {
					break
				}
			}
		}
	}
	return out
}

func collectActivePeerNames(
	start string,
	scope *ast.Node,
	usePos int,
	peers map[string][]containerNamePeerLink,
	reassigns map[string][]int,
) []peerNameState {
	if start == "" {
		return nil
	}
	initState := peerNameState{name: start, tPos: usePos, minPos: -1}
	out := []peerNameState{initState}
	if len(peers) == 0 {
		return out
	}
	seen := map[peerNameState]bool{initState: true}
	const maxStates = 32
	for i := 0; i < len(out) && len(out) < maxStates; i++ {
		cur := out[i]
		for _, link := range peers[cur.name] {
			if link.name == "" {
				continue
			}
			if cur.fromDest && !link.isDest {
				continue
			}
			next := peerNameState{name: link.name, tPos: -1, minPos: -1, fromDest: cur.fromDest || link.isDest}
			if link.scope == scope && link.topLevel && cur.tPos >= 0 {
				if !peerCanTransition(cur.tPos, cur.minPos, link.pos, link.isDest, reassigns[cur.name]) {
					continue
				}
				next.tPos = link.pos
				if link.isDest {
					next.minPos = link.pos
				}
			}
			if !seen[next] {
				seen[next] = true
				out = append(out, next)
				if len(out) >= maxStates {
					break
				}
			}
		}
	}
	return out
}

func hasRemovedInScope(scope *ast.Node, useNode *ast.Node, usePos int, rems []containerRemoval, writes []containerWrite, isInitial bool) bool {
	if len(rems) == 0 {
		return false
	}
	readyPos := -1
	if isInitial {
		readyPos = 0
	}
	for _, w := range writes {
		if w.scope == scope && w.pos < usePos {
			if w.topLevel && w.pos > readyPos {
				readyPos = w.pos
			} else if !w.topLevel && readyPos < 0 {
				readyPos = 0
			}
		}
	}
	if readyPos < 0 {
		readyPos = 0
	}
	for _, rem := range rems {
		if eventInSameLoopAsUse(rem.outerLoop, useNode) || (rem.pos > readyPos && rem.pos < usePos) {
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isContainerInvalidatedAt(node *ast.Node, tb *trackedBindings) bool {
	if node == nil || node.Kind != ast.KindIdentifier || tb == nil || len(tb.containerScopes) == 0 {
		return false
	}
	name := node.AsIdentifier().Text
	nodeFn := ast.GetContainingFunction(node)
	for _, scope := range tb.containerScopes {
		idx := r.getScopeBodyIndex(scope)
		if idx == nil || (len(idx.removedContainers) == 0 && len(idx.removedContainerNames) == 0) {
			continue
		}
		usePos := scope.End()
		sameFn := nodeFn == ast.GetContainingFunction(scope)
		if sameFn {
			usePos = node.Pos()
		}
		var useNode *ast.Node
		if sameFn {
			useNode = node
		}
		if name == "arguments" && tb.trackArguments {
			isImplicitArgs := r == nil || r.refs == nil
			if !isImplicitArgs {
				if sym := r.resolveSymbol(node); sym == nil || len(sym.Declarations) == 0 {
					isImplicitArgs = true
				}
			}
			if isImplicitArgs {
				if hasRemovedInScope(scope, useNode, usePos, idx.removedContainerNames["arguments"], nil, true) {
					return true
				}
				for aliasSym := range tb.argumentsAliasSymbols {
					for _, st := range collectActivePeerSymbols(aliasSym, scope, -1, tb.containerPeers, idx.topLevelReassigns) {
						for _, rem := range idx.removedContainers[st.sym] {
							if (eventInSameLoopAsUse(rem.outerLoop, useNode) || rem.pos < usePos) &&
								peerStateHoldsAt(st.tPos, st.minPos, rem.pos, idx.topLevelReassigns[st.sym]) {
								return true
							}
						}
					}
				}
				for aliasName := range tb.argumentsAliasNames {
					for _, st := range collectActivePeerNames(aliasName, scope, -1, tb.containerNamePeers, idx.topLevelReassignNames) {
						for _, rem := range idx.removedContainerNames[st.name] {
							if (eventInSameLoopAsUse(rem.outerLoop, useNode) || rem.pos < usePos) &&
								peerStateHoldsAt(st.tPos, st.minPos, rem.pos, idx.topLevelReassignNames[st.name]) {
								return true
							}
						}
					}
				}
			}
		}
		if r != nil && r.refs != nil {
			if sym := r.resolveSymbol(node); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
				states := collectActivePeerSymbols(sym, scope, usePos, tb.containerPeers, idx.topLevelReassigns)
				readyPos := -1
				for _, st := range states {
					if tb.initialContainers[st.sym] && readyPos < 0 {
						readyPos = 0
					}
					for _, w := range tb.containerWrites[st.sym] {
						if w.scope == scope && w.pos < usePos &&
							peerStateWriteMatches(st.tPos, st.minPos, st.fromDest, w, idx.topLevelReassigns[st.sym]) {
							if w.topLevel && w.pos > readyPos {
								readyPos = w.pos
							} else if !w.topLevel && readyPos < 0 {
								readyPos = 0
							}
						}
					}
				}
				if readyPos < 0 {
					readyPos = 0
				}
				for _, st := range states {
					if tb.argumentsAliasSymbols[st.sym] &&
						hasRemovedInScope(scope, useNode, usePos, idx.removedContainerNames["arguments"], nil, true) {
						return true
					}
					for _, rem := range idx.removedContainers[st.sym] {
						if !peerStateHoldsAt(st.tPos, st.minPos, rem.pos, idx.topLevelReassigns[st.sym]) {
							continue
						}
						if eventInSameLoopAsUse(rem.outerLoop, useNode) || (rem.pos > readyPos && rem.pos < usePos) {
							return true
						}
					}
				}
				continue
			}
		}
		states := collectActivePeerNames(name, scope, usePos, tb.containerNamePeers, idx.topLevelReassignNames)
		readyPos := -1
		for _, st := range states {
			if tb.initialContainerNames[st.name] && readyPos < 0 {
				readyPos = 0
			}
			for _, w := range tb.containerNameWrites[st.name] {
				if w.scope == scope && w.pos < usePos &&
					peerStateWriteMatches(st.tPos, st.minPos, st.fromDest, w, idx.topLevelReassignNames[st.name]) {
					if w.topLevel && w.pos > readyPos {
						readyPos = w.pos
					} else if !w.topLevel && readyPos < 0 {
						readyPos = 0
					}
				}
			}
		}
		if readyPos < 0 {
			readyPos = 0
		}
		for _, st := range states {
			if tb.argumentsAliasNames[st.name] &&
				hasRemovedInScope(scope, useNode, usePos, idx.removedContainerNames["arguments"], nil, true) {
				return true
			}
			for _, rem := range idx.removedContainerNames[st.name] {
				if !peerStateHoldsAt(st.tPos, st.minPos, rem.pos, idx.topLevelReassignNames[st.name]) {
					continue
				}
				if eventInSameLoopAsUse(rem.outerLoop, useNode) || (rem.pos > readyPos && rem.pos < usePos) {
					return true
				}
			}
		}
	}
	return false
}

// survivingContainerWritesAt reports whether the container or map named by `node`
// already holds a tracked Promise at `node`'s evaluation point (preventing a
// later write such as `a.push(check)` from crediting an earlier empty
// `Promise.all(a)` snapshot, and rejecting a map write whose static key(s) were
// later overwritten via `.set(sameKey, other)` before `node`), together with
// the union of surviving static Map keys (or `mapHasDynamic = true` when any
// surviving write is non-map or has dynamic/unknown keys).
func (r *AsyncAssertionResolver) survivingContainerWritesAt(node *ast.Node, tb *trackedBindings) (populated bool, mapStaticKeys map[string]bool, mapHasDynamic bool) {
	if node == nil || node.Kind != ast.KindIdentifier || tb == nil || len(tb.containerScopes) == 0 {
		return true, nil, true
	}
	name := node.AsIdentifier().Text
	nodeFn := ast.GetContainingFunction(node)
	var activeScope *ast.Node
	for _, scope := range tb.containerScopes {
		if nodeFn == ast.GetContainingFunction(scope) {
			activeScope = scope
			break
		}
	}
	var activeIdx *scopeBodyIndex
	if activeScope != nil {
		activeIdx = r.getScopeBodyIndex(activeScope)
	}
	enclosingScope := activeScope
	if enclosingScope == nil {
		for i := len(tb.containerScopes) - 1; i >= 0; i-- {
			if nodeContains(tb.containerScopes[i], node) {
				enclosingScope = tb.containerScopes[i]
				break
			}
		}
	}
	usePos := node.Pos()
	addSurviving := func(keys map[string]bool, dyn bool) {
		populated = true
		if dyn || len(keys) == 0 {
			mapHasDynamic = true
			return
		}
		if mapStaticKeys == nil {
			mapStaticKeys = map[string]bool{}
		}
		for k := range keys {
			mapStaticKeys[k] = true
		}
	}
	if r != nil && r.refs != nil {
		if sym := r.resolveSymbol(node); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
			var initReassigns map[*ast.Symbol][]int
			initTPos := -1
			if enclosingScope != nil {
				if eIdx := r.getScopeBodyIndex(enclosingScope); eIdx != nil {
					initReassigns = eIdx.topLevelReassigns
				}
				initTPos = usePos
			}
			states := collectActivePeerSymbols(sym, enclosingScope, initTPos, tb.containerPeers, initReassigns)
			scopeStatesCache := map[*ast.Node][]peerSymbolState{}
			if enclosingScope != nil {
				scopeStatesCache[enclosingScope] = states
			}
			statesInScope := func(scope *ast.Node, reassigns map[*ast.Symbol][]int) []peerSymbolState {
				cached, ok := scopeStatesCache[scope]
				if !ok {
					cached = collectActivePeerSymbols(sym, scope, usePos, tb.containerPeers, reassigns)
					scopeStatesCache[scope] = cached
				}
				return cached
			}
			var activeReassigns map[*ast.Symbol][]int
			if activeIdx != nil {
				activeReassigns = activeIdx.topLevelReassigns
			}
			for _, st := range states {
				if tb.argumentsAliasSymbols[st.sym] {
					return true, nil, true
				}
				if tb.initialContainers[st.sym] &&
					(st.tPos < 0 || !hasTopLevelReassignBetween(initReassigns[st.sym], -1, st.tPos)) {
					return true, nil, true
				}
				for _, w := range tb.containerWrites[st.sym] {
					if w.scope == nil || !nodeContains(w.scope, node) {
						addSurviving(w.mapStaticKeys, len(w.mapStaticKeys) == 0)
						if mapHasDynamic {
							return true, nil, true
						}
						continue
					}
					wScope := w.scope
					wIdx := r.getScopeBodyIndex(wScope)
					var wReassigns map[*ast.Symbol][]int
					if wIdx != nil {
						wReassigns = wIdx.topLevelReassigns
					}
					sameFn := wScope == activeScope
					var wUseNode *ast.Node
					if sameFn {
						wUseNode = node
					}
					if sameFn {
						if w.pos >= usePos && !eventInSameLoopAsUse(w.outerLoop, wUseNode) {
							continue
						}
						if !peerStateWriteMatches(st.tPos, st.minPos, st.fromDest, w, wReassigns[st.sym]) {
							continue
						}
					} else if w.pos < usePos {
						wStates := statesInScope(wScope, wReassigns)
						matchedPeer := false
						for _, cand := range wStates {
							if cand.sym == st.sym && peerStateWriteMatches(cand.tPos, cand.minPos, cand.fromDest, w, wReassigns[st.sym]) {
								matchedPeer = true
								break
							}
						}
						if !matchedPeer {
							continue
						}
					}
					if len(w.mapStaticKeys) == 0 {
						return true, nil, true
					}
					wEnd := w.endPos
					if wEnd == 0 {
						wEnd = w.pos
					}
					var wStates []peerSymbolState
					if wIdx != nil && len(wIdx.mapKeySets) > 0 && (sameFn || w.pos < usePos) {
						wStates = statesInScope(wScope, wReassigns)
					}
					var surviving map[string]bool
					for sk := range w.mapStaticKeys {
						overwritten := false
						if len(wStates) > 0 {
							for _, st2 := range wStates {
								for _, ev := range wIdx.mapKeySets[st2.sym] {
									if ev.key != sk || !peerStateHoldsAt(st2.tPos, st2.minPos, ev.pos, wReassigns[st2.sym]) {
										continue
									}
									if w.pos < usePos {
										if ev.startPos >= wEnd && ev.pos < usePos {
											overwritten = true
											break
										}
									} else if eventInSameLoopAsUse(ev.outerLoop, wUseNode) && (ev.startPos >= wEnd || ev.pos < usePos) {
										overwritten = true
										break
									}
								}
								if overwritten {
									break
								}
							}
						}
						if !overwritten && activeScope != nil && activeScope != wScope && activeIdx != nil && len(activeIdx.mapKeySets) > 0 {
							for _, st2 := range states {
								for _, ev := range activeIdx.mapKeySets[st2.sym] {
									if ev.key != sk || !peerStateHoldsAt(st2.tPos, st2.minPos, ev.pos, activeReassigns[st2.sym]) {
										continue
									}
									if ev.pos < usePos {
										overwritten = true
										break
									}
								}
								if overwritten {
									break
								}
							}
						}
						if !overwritten {
							if surviving == nil {
								surviving = map[string]bool{}
							}
							surviving[sk] = true
						}
					}
					if len(surviving) > 0 {
						addSurviving(surviving, false)
					}
				}
			}
			return populated, mapStaticKeys, mapHasDynamic
		}
	}
	var initReassigns map[string][]int
	initTPos := -1
	if enclosingScope != nil {
		if eIdx := r.getScopeBodyIndex(enclosingScope); eIdx != nil {
			initReassigns = eIdx.topLevelReassignNames
		}
		initTPos = usePos
	}
	states := collectActivePeerNames(name, enclosingScope, initTPos, tb.containerNamePeers, initReassigns)
	scopeStatesCache := map[*ast.Node][]peerNameState{}
	if enclosingScope != nil {
		scopeStatesCache[enclosingScope] = states
	}
	statesInScope := func(scope *ast.Node, reassigns map[string][]int) []peerNameState {
		cached, ok := scopeStatesCache[scope]
		if !ok {
			cached = collectActivePeerNames(name, scope, usePos, tb.containerNamePeers, reassigns)
			scopeStatesCache[scope] = cached
		}
		return cached
	}
	var activeReassigns map[string][]int
	if activeIdx != nil {
		activeReassigns = activeIdx.topLevelReassignNames
	}
	for _, st := range states {
		if tb.argumentsAliasNames[st.name] {
			return true, nil, true
		}
		if tb.initialContainerNames[st.name] &&
			(st.tPos < 0 || !hasTopLevelReassignBetween(initReassigns[st.name], -1, st.tPos)) {
			return true, nil, true
		}
		for _, w := range tb.containerNameWrites[st.name] {
			if w.scope == nil || !nodeContains(w.scope, node) {
				addSurviving(w.mapStaticKeys, len(w.mapStaticKeys) == 0)
				if mapHasDynamic {
					return true, nil, true
				}
				continue
			}
			wScope := w.scope
			wIdx := r.getScopeBodyIndex(wScope)
			var wReassigns map[string][]int
			if wIdx != nil {
				wReassigns = wIdx.topLevelReassignNames
			}
			sameFn := wScope == activeScope
			var wUseNode *ast.Node
			if sameFn {
				wUseNode = node
			}
			if sameFn {
				if w.pos >= usePos && !eventInSameLoopAsUse(w.outerLoop, wUseNode) {
					continue
				}
				if !peerStateWriteMatches(st.tPos, st.minPos, st.fromDest, w, wReassigns[st.name]) {
					continue
				}
			} else if w.pos < usePos {
				wStates := statesInScope(wScope, wReassigns)
				matchedPeer := false
				for _, cand := range wStates {
					if cand.name == st.name && peerStateWriteMatches(cand.tPos, cand.minPos, cand.fromDest, w, wReassigns[st.name]) {
						matchedPeer = true
						break
					}
				}
				if !matchedPeer {
					continue
				}
			}
			if len(w.mapStaticKeys) == 0 {
				return true, nil, true
			}
			wEnd := w.endPos
			if wEnd == 0 {
				wEnd = w.pos
			}
			var wStates []peerNameState
			if wIdx != nil && len(wIdx.mapNameKeySets) > 0 && (sameFn || w.pos < usePos) {
				wStates = statesInScope(wScope, wReassigns)
			}
			var surviving map[string]bool
			for sk := range w.mapStaticKeys {
				overwritten := false
				if len(wStates) > 0 {
					for _, st2 := range wStates {
						for _, ev := range wIdx.mapNameKeySets[st2.name] {
							if ev.key != sk || !peerStateHoldsAt(st2.tPos, st2.minPos, ev.pos, wReassigns[st2.name]) {
								continue
							}
							if w.pos < usePos {
								if ev.startPos >= wEnd && ev.pos < usePos {
									overwritten = true
									break
								}
							} else if eventInSameLoopAsUse(ev.outerLoop, wUseNode) && (ev.startPos >= wEnd || ev.pos < usePos) {
								overwritten = true
								break
							}
						}
						if overwritten {
							break
						}
					}
				}
				if !overwritten && activeScope != nil && activeScope != wScope && activeIdx != nil && len(activeIdx.mapNameKeySets) > 0 {
					for _, st2 := range states {
						for _, ev := range activeIdx.mapNameKeySets[st2.name] {
							if ev.key != sk || !peerStateHoldsAt(st2.tPos, st2.minPos, ev.pos, activeReassigns[st2.name]) {
								continue
							}
							if ev.pos < usePos {
								overwritten = true
								break
							}
						}
						if overwritten {
							break
						}
					}
				}
				if !overwritten {
					if surviving == nil {
						surviving = map[string]bool{}
					}
					surviving[sk] = true
				}
			}
			if len(surviving) > 0 {
				addSurviving(surviving, false)
			}
		}
	}
	return populated, mapStaticKeys, mapHasDynamic
}

func (r *AsyncAssertionResolver) isContainerPopulatedAt(node *ast.Node, tb *trackedBindings) bool {
	populated, _, _ := r.survivingContainerWritesAt(node, tb)
	return populated
}

func (r *AsyncAssertionResolver) matchesTrackedRestIdent(node *ast.Node, tb *trackedBindings, shadowed map[string]bool) bool {
	if node == nil || node.Kind != ast.KindIdentifier || !isIdentifierValueReference(node) {
		return false
	}
	if r.isContainerInvalidatedAt(node, tb) {
		return false
	}
	name := node.AsIdentifier().Text
	if tb.trackArguments && name == "arguments" && !shadowed["arguments"] {
		if r == nil || r.refs == nil {
			return true
		}
		if sym := r.resolveSymbol(node); sym == nil || len(sym.Declarations) == 0 {
			return true
		}
	}
	if r != nil && r.refs != nil && len(tb.restSymbols) > 0 {
		if sym := r.resolveSymbol(node); sym != nil {
			return tb.restSymbols[sym]
		}
		if len(tb.restNames) > 0 && tb.restNames[name] && !shadowed[name] {
			return true
		}
		return false
	}
	return tb.restNames[name] && !shadowed[name]
}

func (r *AsyncAssertionResolver) matchesTrackedContainerIdent(node *ast.Node, tb *trackedBindings, shadowed map[string]bool, requirePopulated bool) bool {
	if node == nil || node.Kind != ast.KindIdentifier || !isIdentifierValueReference(node) {
		return false
	}
	if r.isContainerInvalidatedAt(node, tb) {
		return false
	}
	if requirePopulated && !r.isContainerPopulatedAt(node, tb) {
		return false
	}
	name := node.AsIdentifier().Text
	if r != nil && r.refs != nil && len(tb.containerSymbols) > 0 {
		if sym := r.resolveSymbol(node); sym != nil {
			return tb.containerSymbols[sym]
		}
		if len(tb.containerNames) > 0 && tb.containerNames[name] && !shadowed[name] {
			return true
		}
		return false
	}
	return tb.containerNames[name] && !shadowed[name]
}

func (r *AsyncAssertionResolver) matchesTrackedMapIdent(node *ast.Node, tb *trackedBindings, shadowed map[string]bool, requirePopulated bool) bool {
	if node == nil || node.Kind != ast.KindIdentifier || !isIdentifierValueReference(node) {
		return false
	}
	if r.isContainerInvalidatedAt(node, tb) {
		return false
	}
	if requirePopulated && !r.isContainerPopulatedAt(node, tb) {
		return false
	}
	name := node.AsIdentifier().Text
	if r != nil && r.refs != nil && len(tb.mapSymbols) > 0 {
		if sym := r.resolveSymbol(node); sym != nil {
			return tb.mapSymbols[sym]
		}
		if len(tb.mapNames) > 0 && tb.mapNames[name] && !shadowed[name] {
			return true
		}
		return false
	}
	return tb.mapNames[name] && !shadowed[name]
}

func (r *AsyncAssertionResolver) matchesTrackedIdent(node *ast.Node, tb *trackedBindings, shadowed map[string]bool) bool {
	return r.matchesTrackedDirectIdent(node, tb, shadowed) ||
		r.matchesTrackedRestIdent(node, tb, shadowed) ||
		r.matchesTrackedContainerIdent(node, tb, shadowed, false) ||
		r.matchesTrackedMapIdent(node, tb, shadowed, false)
}

func scopeShadowsNames(node *ast.Node, tb *trackedBindings, shadowed map[string]bool) map[string]bool {
	if node == nil {
		return shadowed
	}
	var added []string
	checkName := func(n string) {
		if (tb.names[n] || tb.restNames[n] || tb.containerNames[n] || tb.mapNames[n] || (tb.trackArguments && n == "arguments")) && !shadowed[n] {
			added = append(added, n)
		}
	}
	if ast.IsLocalsContainer(node) {
		locals := ast.GetLocals(node)
		for n, sym := range locals {
			if sym != nil && sym.Flags&(ast.SymbolFlagsValue|ast.SymbolFlagsAlias) != 0 && hasValueDeclaration(sym) {
				checkName(n)
			}
		}
	}
	if testFramework.IsFunction(node) {
		for _, p := range node.Parameters() {
			if p != nil {
				for _, ident := range collectBindingIdentifiers(p.Name()) {
					checkName(ident.AsIdentifier().Text)
				}
			}
		}
		if node.Kind == ast.KindFunctionExpression {
			if fnName := node.Name(); fnName != nil && fnName.Kind == ast.KindIdentifier {
				checkName(fnName.AsIdentifier().Text)
			}
		}
		// Non-arrow nested functions bind their own `arguments`.
		if node.Kind != ast.KindArrowFunction {
			checkName("arguments")
		}
	}
	if node.Kind == ast.KindClassExpression {
		if cn := node.Name(); cn != nil && cn.Kind == ast.KindIdentifier {
			checkName(cn.AsIdentifier().Text)
		}
	}
	if len(added) == 0 {
		return shadowed
	}
	next := make(map[string]bool, len(shadowed)+len(added))
	for k, v := range shadowed {
		next[k] = v
	}
	for _, n := range added {
		next[n] = true
	}
	return next
}

// exprReferencesTracked reports whether `root` contains a value reference to
// any binding in `tb` (respecting shadowing and stopping at non-arrow
// functions for `arguments`).
func (r *AsyncAssertionResolver) exprReferencesTracked(root *ast.Node, tb *trackedBindings, enterNestedFunctions bool) bool {
	if root == nil {
		return false
	}
	found := false
	var visit func(*ast.Node, map[string]bool)
	visit = func(node *ast.Node, shadowed map[string]bool) {
		if node == nil || found || ast.IsTypeNode(node) || ast.IsPartOfTypeQuery(node) {
			return
		}
		if node != root {
			switch node.Kind {
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
				ast.KindClassStaticBlockDeclaration:
				if !enterNestedFunctions {
					return
				}
				shadowed = scopeShadowsNames(node, tb, shadowed)
			case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
				if mods := node.Modifiers(); mods != nil {
					for _, mod := range mods.Nodes {
						if mod != nil && mod.Kind == ast.KindDecorator {
							visit(mod, shadowed)
						}
					}
				}
				if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					visit(name, shadowed)
				}
				if !enterNestedFunctions {
					return
				}
				shadowed = scopeShadowsNames(node, tb, shadowed)
			default:
				if ast.IsLocalsContainer(node) || node.Kind == ast.KindClassExpression {
					shadowed = scopeShadowsNames(node, tb, shadowed)
				}
			}
		}
		if node.Kind == ast.KindIdentifier && r.matchesTrackedIdent(node, tb, shadowed) {
			found = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child, shadowed)
			return found
		})
	}
	visit(root, nil)
	return found
}

func isConsoleCall(call *ast.Node) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	callee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return false
	}
	recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
	return recv != nil && recv.Kind == ast.KindIdentifier && recv.AsIdentifier().Text == "console"
}

func (r *AsyncAssertionResolver) isRestElementAccess(expr *ast.Node, tb *trackedBindings) bool {
	if expr == nil || expr.Kind != ast.KindElementAccessExpression {
		return false
	}
	recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(expr))
	if !r.matchesTrackedRestIdent(recv, tb, nil) {
		return false
	}
	if name, ok := r.accessStaticNameThroughAssertions(expr); ok {
		return tb.restOffset >= 0 && name == strconv.Itoa(tb.restOffset)
	}
	return true
}

func (r *AsyncAssertionResolver) resolveValidArraySlotsAt(recv *ast.Node, tb *trackedBindings) (slots map[int]bool, arrLen int, ok bool) {
	if recv == nil || recv.Kind != ast.KindIdentifier || !isIdentifierValueReference(recv) || tb == nil ||
		(len(tb.arraySlotSymbols) == 0 && len(tb.arraySlotNames) == 0) {
		return nil, -1, false
	}
	name := recv.AsIdentifier().Text
	var sym *ast.Symbol
	arrLen = -1
	if r != nil && r.refs != nil && len(tb.arraySlotSymbols) > 0 {
		sym = r.resolveSymbol(recv)
		if sym != nil {
			slots = tb.arraySlotSymbols[sym]
			if l, hasLen := tb.arraySlotLen[sym]; hasLen {
				arrLen = l
			}
		} else {
			slots = tb.arraySlotNames[name]
			if l, hasLen := tb.arraySlotNameLen[name]; hasLen {
				arrLen = l
			}
		}
	} else {
		slots = tb.arraySlotNames[name]
		if l, hasLen := tb.arraySlotNameLen[name]; hasLen {
			arrLen = l
		}
	}
	if len(slots) == 0 {
		return nil, -1, false
	}
	// Array literal slot indices remain valid only if neither `recv` nor any
	// active alias peer has been shifted, overwritten, or truncated since the
	// literal was created (later `.push()` calls never restore invalidated slot indices).
	nodeFn := ast.GetContainingFunction(recv)
	for _, scope := range tb.containerScopes {
		idx := r.getScopeBodyIndex(scope)
		if idx == nil {
			continue
		}
		usePos := scope.End()
		sameFn := nodeFn == ast.GetContainingFunction(scope)
		var useNode *ast.Node
		if sameFn {
			usePos = recv.Pos()
			useNode = recv
		}
		if sym != nil {
			declPos := tb.arraySlotDeclPos[sym]
			if declPos < scope.Pos() || declPos >= scope.End() {
				declPos = 0
			}
			for _, st := range collectActivePeerSymbols(sym, scope, usePos, tb.containerPeers, idx.topLevelReassigns) {
				if tb.argumentsAliasSymbols[st.sym] &&
					hasRemovedInScope(scope, useNode, usePos, idx.removedContainerNames["arguments"], nil, true) {
					return nil, -1, false
				}
				for _, sh := range idx.shiftedContainers[st.sym] {
					if !peerStateHoldsAt(st.tPos, st.minPos, sh.pos, idx.topLevelReassigns[st.sym]) {
						continue
					}
					if eventInSameLoopAsUse(sh.outerLoop, useNode) || (sh.pos > declPos && sh.pos < usePos) {
						return nil, -1, false
					}
				}
				for _, rem := range idx.removedContainers[st.sym] {
					if !peerStateHoldsAt(st.tPos, st.minPos, rem.pos, idx.topLevelReassigns[st.sym]) {
						continue
					}
					if eventInSameLoopAsUse(rem.outerLoop, useNode) || (rem.pos > declPos && rem.pos < usePos) {
						return nil, -1, false
					}
				}
			}
		} else {
			declPos := tb.arraySlotDeclNamePos[name]
			if declPos < scope.Pos() || declPos >= scope.End() {
				declPos = 0
			}
			for _, st := range collectActivePeerNames(name, scope, usePos, tb.containerNamePeers, idx.topLevelReassignNames) {
				if tb.argumentsAliasNames[st.name] &&
					hasRemovedInScope(scope, useNode, usePos, idx.removedContainerNames["arguments"], nil, true) {
					return nil, -1, false
				}
				for _, sh := range idx.shiftedContainerNames[st.name] {
					if !peerStateHoldsAt(st.tPos, st.minPos, sh.pos, idx.topLevelReassignNames[st.name]) {
						continue
					}
					if eventInSameLoopAsUse(sh.outerLoop, useNode) || (sh.pos > declPos && sh.pos < usePos) {
						return nil, -1, false
					}
				}
				for _, rem := range idx.removedContainerNames[st.name] {
					if !peerStateHoldsAt(st.tPos, st.minPos, rem.pos, idx.topLevelReassignNames[st.name]) {
						continue
					}
					if eventInSameLoopAsUse(rem.outerLoop, useNode) || (rem.pos > declPos && rem.pos < usePos) {
						return nil, -1, false
					}
				}
			}
		}
	}
	return slots, arrLen, true
}

func (r *AsyncAssertionResolver) hasContainerPushSinceDecl(recv *ast.Node, tb *trackedBindings) bool {
	if recv == nil || recv.Kind != ast.KindIdentifier || tb == nil {
		return false
	}
	name := recv.AsIdentifier().Text
	nodeFn := ast.GetContainingFunction(recv)
	for _, scope := range tb.containerScopes {
		idx := r.getScopeBodyIndex(scope)
		if idx == nil || (len(idx.pushedContainers) == 0 && len(idx.pushedContainerNames) == 0) {
			continue
		}
		usePos := scope.End()
		sameFn := nodeFn == ast.GetContainingFunction(scope)
		var useNode *ast.Node
		if sameFn {
			usePos = recv.Pos()
			useNode = recv
		}
		if r != nil && r.refs != nil {
			if sym := r.resolveSymbol(recv); sym != nil {
				declPos := tb.arraySlotDeclPos[sym]
				if declPos < scope.Pos() || declPos >= scope.End() {
					declPos = 0
				}
				for _, st := range collectActivePeerSymbols(sym, scope, usePos, tb.containerPeers, idx.topLevelReassigns) {
					for _, p := range idx.pushedContainers[st.sym] {
						if !peerStateHoldsAt(st.tPos, st.minPos, p.pos, idx.topLevelReassigns[st.sym]) {
							continue
						}
						if eventInSameLoopAsUse(p.outerLoop, useNode) || (p.pos > declPos && p.pos < usePos) {
							return true
						}
					}
				}
				continue
			}
		}
		declPos := tb.arraySlotDeclNamePos[name]
		if declPos < scope.Pos() || declPos >= scope.End() {
			declPos = 0
		}
		for _, st := range collectActivePeerNames(name, scope, usePos, tb.containerNamePeers, idx.topLevelReassignNames) {
			for _, p := range idx.pushedContainerNames[st.name] {
				if !peerStateHoldsAt(st.tPos, st.minPos, p.pos, idx.topLevelReassignNames[st.name]) {
					continue
				}
				if eventInSameLoopAsUse(p.outerLoop, useNode) || (p.pos > declPos && p.pos < usePos) {
					return true
				}
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) hasAnyPushBeforeUse(recv *ast.Node, tb *trackedBindings) bool {
	if recv == nil || recv.Kind != ast.KindIdentifier || tb == nil {
		return false
	}
	name := recv.AsIdentifier().Text
	nodeFn := ast.GetContainingFunction(recv)
	for _, scope := range tb.containerScopes {
		idx := r.getScopeBodyIndex(scope)
		if idx == nil || (len(idx.pushedContainers) == 0 && len(idx.pushedContainerNames) == 0) {
			continue
		}
		usePos := scope.End()
		sameFn := nodeFn == ast.GetContainingFunction(scope)
		var useNode *ast.Node
		if sameFn {
			usePos = recv.Pos()
			useNode = recv
		}
		if r != nil && r.refs != nil {
			if sym := r.resolveSymbol(recv); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
				for _, st := range collectActivePeerSymbols(sym, scope, usePos, tb.containerPeers, idx.topLevelReassigns) {
					for _, p := range idx.pushedContainers[st.sym] {
						if !peerStateHoldsAt(st.tPos, st.minPos, p.pos, idx.topLevelReassigns[st.sym]) {
							continue
						}
						if eventInSameLoopAsUse(p.outerLoop, useNode) || p.pos < usePos {
							return true
						}
					}
				}
				continue
			}
		}
		for _, st := range collectActivePeerNames(name, scope, usePos, tb.containerNamePeers, idx.topLevelReassignNames) {
			for _, p := range idx.pushedContainerNames[st.name] {
				if !peerStateHoldsAt(st.tPos, st.minPos, p.pos, idx.topLevelReassignNames[st.name]) {
					continue
				}
				if eventInSameLoopAsUse(p.outerLoop, useNode) || p.pos < usePos {
					return true
				}
			}
		}
	}
	return false
}

// lastPushBeforeUseIsTracked reports whether the latest `.push(...)` call on
// `recv` (or an active alias peer) before `recv` ends with a tracked Promise
// element (e.g. `const a = []; a.push(check); return a.pop();`), returning
// `false` if an untracked `.push(other)` happened later or the push ended with
// a non-tracked argument (e.g. `a.push(check, Promise.resolve(0)); return a.pop();`).
func (r *AsyncAssertionResolver) lastPushBeforeUseIsTracked(recv *ast.Node, tb *trackedBindings) bool {
	if recv == nil || recv.Kind != ast.KindIdentifier || tb == nil || len(tb.containerScopes) == 0 {
		return false
	}
	name := recv.AsIdentifier().Text
	nodeFn := ast.GetContainingFunction(recv)
	var activeScope *ast.Node
	for _, scope := range tb.containerScopes {
		if nodeFn == ast.GetContainingFunction(scope) {
			activeScope = scope
			break
		}
	}
	if activeScope == nil {
		return true
	}
	idx := r.getScopeBodyIndex(activeScope)
	if idx == nil {
		return false
	}
	usePos := recv.Pos()
	checkStates := func(
		latestTrackedPos int,
		latestAnyPushEnd int,
		latestAnyPushStart int,
		hasOtherLoopPush bool,
		hasShiftAfterTracked bool,
	) bool {
		if latestTrackedPos < 0 || hasShiftAfterTracked || hasOtherLoopPush {
			return false
		}
		return latestAnyPushStart <= latestTrackedPos && latestAnyPushEnd > latestTrackedPos
	}
	if r != nil && r.refs != nil {
		if sym := r.resolveSymbol(recv); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
			latestTrackedPos := -1
			latestAnyPushEnd := -1
			latestAnyPushStart := -1
			states := collectActivePeerSymbols(sym, activeScope, usePos, tb.containerPeers, idx.topLevelReassigns)
			for _, st := range states {
				for _, w := range tb.containerWrites[st.sym] {
					if w.scope != activeScope || !w.isLastPushElem {
						continue
					}
					if !peerStateWriteMatches(st.tPos, st.minPos, st.fromDest, w, idx.topLevelReassigns[st.sym]) {
						continue
					}
					if w.pos < usePos && w.pos > latestTrackedPos {
						latestTrackedPos = w.pos
					}
				}
				for _, p := range idx.pushedContainers[st.sym] {
					if !peerStateHoldsAt(st.tPos, st.minPos, p.pos, idx.topLevelReassigns[st.sym]) {
						continue
					}
					if p.pos < usePos && p.pos > latestAnyPushEnd {
						latestAnyPushEnd = p.pos
						latestAnyPushStart = p.startPos
					}
				}
			}
			hasOtherLoopPush := false
			hasShift := false
			if latestTrackedPos >= 0 {
				for _, st := range states {
					for _, p := range idx.pushedContainers[st.sym] {
						if !peerStateHoldsAt(st.tPos, st.minPos, p.pos, idx.topLevelReassigns[st.sym]) {
							continue
						}
						if eventInSameLoopAsUse(p.outerLoop, recv) && (p.startPos > latestTrackedPos || p.pos <= latestTrackedPos) {
							hasOtherLoopPush = true
							break
						}
					}
					for _, sh := range idx.shiftedContainers[st.sym] {
						if !peerStateHoldsAt(st.tPos, st.minPos, sh.pos, idx.topLevelReassigns[st.sym]) {
							continue
						}
						if eventInSameLoopAsUse(sh.outerLoop, recv) || (sh.startPos > latestTrackedPos && sh.pos < usePos) {
							hasShift = true
							break
						}
					}
				}
			}
			return checkStates(latestTrackedPos, latestAnyPushEnd, latestAnyPushStart, hasOtherLoopPush, hasShift)
		}
	}
	latestTrackedPos := -1
	latestAnyPushEnd := -1
	latestAnyPushStart := -1
	states := collectActivePeerNames(name, activeScope, usePos, tb.containerNamePeers, idx.topLevelReassignNames)
	for _, st := range states {
		for _, w := range tb.containerNameWrites[st.name] {
			if w.scope != activeScope || !w.isLastPushElem {
				continue
			}
			if !peerStateWriteMatches(st.tPos, st.minPos, st.fromDest, w, idx.topLevelReassignNames[st.name]) {
				continue
			}
			if w.pos < usePos && w.pos > latestTrackedPos {
				latestTrackedPos = w.pos
			}
		}
		for _, p := range idx.pushedContainerNames[st.name] {
			if !peerStateHoldsAt(st.tPos, st.minPos, p.pos, idx.topLevelReassignNames[st.name]) {
				continue
			}
			if p.pos < usePos && p.pos > latestAnyPushEnd {
				latestAnyPushEnd = p.pos
				latestAnyPushStart = p.startPos
			}
		}
	}
	hasOtherLoopPush := false
	hasShift := false
	if latestTrackedPos >= 0 {
		for _, st := range states {
			for _, p := range idx.pushedContainerNames[st.name] {
				if !peerStateHoldsAt(st.tPos, st.minPos, p.pos, idx.topLevelReassignNames[st.name]) {
					continue
				}
				if eventInSameLoopAsUse(p.outerLoop, recv) && (p.startPos > latestTrackedPos || p.pos <= latestTrackedPos) {
					hasOtherLoopPush = true
					break
				}
			}
			for _, sh := range idx.shiftedContainerNames[st.name] {
				if !peerStateHoldsAt(st.tPos, st.minPos, sh.pos, idx.topLevelReassignNames[st.name]) {
					continue
				}
				if eventInSameLoopAsUse(sh.outerLoop, recv) || (sh.startPos > latestTrackedPos && sh.pos < usePos) {
					hasShift = true
					break
				}
			}
		}
	}
	return checkStates(latestTrackedPos, latestAnyPushEnd, latestAnyPushStart, hasOtherLoopPush, hasShift)
}

func (r *AsyncAssertionResolver) isTrackedArraySlotAccess(expr *ast.Node, tb *trackedBindings) bool {
	if expr == nil || expr.Kind != ast.KindElementAccessExpression {
		return false
	}
	recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(expr))
	slots, _, ok := r.resolveValidArraySlotsAt(recv, tb)
	if !ok {
		return false
	}
	propName, ok := r.accessStaticNameThroughAssertions(expr)
	if !ok {
		return false
	}
	idx, err := strconv.Atoi(propName)
	if err != nil || idx < 0 || strconv.Itoa(idx) != propName {
		return false
	}
	return slots[idx]
}

// exprEvaluatesToTrackedPromise reports whether `expr` can evaluate directly to
// a tracked Promise itself (used for `throw <expr>` rethrow checks so
// `throw new Error(String(err))` is not mistaken for rethrowing `err`).
func (r *AsyncAssertionResolver) exprEvaluatesToTrackedPromise(expr *ast.Node, tb *trackedBindings) bool {
	expr = internalUtils.SkipAssertionsAndParens(expr)
	for expr != nil && internalUtils.IsCommaOperator(expr) {
		expr = internalUtils.SkipAssertionsAndParens(expr.AsBinaryExpression().Right)
	}
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		return r.matchesTrackedDirectIdent(expr, tb, nil)
	case ast.KindElementAccessExpression:
		return r.isRestElementAccess(expr, tb) || r.isTrackedArraySlotAccess(expr, tb)
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		if cond == nil {
			return false
		}
		return r.exprEvaluatesToTrackedPromise(cond.WhenTrue, tb) ||
			r.exprEvaluatesToTrackedPromise(cond.WhenFalse, tb)
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
			return r.exprEvaluatesToTrackedPromise(bin.Right, tb)
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken,
			ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return r.exprEvaluatesToTrackedPromise(bin.Left, tb) ||
				r.exprEvaluatesToTrackedPromise(bin.Right, tb)
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isKnownBuiltinCollectionType(t *checker.Type) bool {
	if r == nil || r.typeChecker == nil || t == nil {
		return false
	}
	parts := internalUtils.UnionTypeParts(t)
	if len(parts) == 0 {
		return false
	}
	for _, part := range parts {
		if part == nil {
			return false
		}
		if part.Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
			return false
		}
		if checker.Checker_isArrayType(r.typeChecker, part) || checker.IsTupleType(part) {
			continue
		}
		sym := part.Symbol()
		if sym == nil || r.program == nil || !internalUtils.IsSymbolFromDefaultLibrary(r.program, sym) {
			return false
		}
		switch sym.Name {
		case "Set", "ReadonlySet", "Map", "ReadonlyMap", "WeakSet", "WeakMap":
			continue
		default:
			return false
		}
	}
	return true
}

func (r *AsyncAssertionResolver) fnBodyCanReturnPromise(fn *ast.Node, depth int) bool {
	if fn == nil || functionAsteriskToken(fn) != nil {
		return false
	}
	if IsAsyncFunctionLike(fn) {
		return true
	}
	body := fn.Body()
	if body == nil {
		return true
	}
	exprCanBePromise := func(expr *ast.Node) bool {
		u := internalUtils.SkipAssertionsAndParens(expr)
		if u == nil || u.Kind == ast.KindVoidExpression {
			return false
		}
		if r != nil && r.typeChecker != nil {
			if retType := r.typeChecker.GetTypeAtLocation(u); retType != nil &&
				r.isKnownNonPromiseReturnTypeDepth(retType, u, depth+1) {
				return false
			}
		}
		return true
	}
	if body.Kind != ast.KindBlock {
		return exprCanBePromise(body)
	}
	hasReturnValue := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || hasReturnValue || ast.IsTypeNode(node) {
			return
		}
		if node != body {
			switch node.Kind {
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
				ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
				ast.KindClassStaticBlockDeclaration:
				return
			}
		}
		if node.Kind == ast.KindReturnStatement {
			if ret := node.AsReturnStatement(); ret != nil && ret.Expression != nil {
				if exprCanBePromise(ret.Expression) {
					hasReturnValue = true
					return
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return hasReturnValue
		})
	}
	visit(body)
	return hasReturnValue
}

// voidCallCanReturnValue reports whether a call whose TypeScript return type is
// `void` can still return a Promise at runtime because of TypeScript's contextual
// `() => void` assignability (e.g. `const dispatch: (p: Promise<unknown>) => void = p => p`),
// while rejecting void-typed helpers whose bodies return non-Promise values (e.g. `p => queue.push(p)`).
func (r *AsyncAssertionResolver) voidCallCanReturnValue(callExpr *ast.Node, depth int) bool {
	if depth > 4 {
		return true
	}
	if callExpr == nil || callExpr.Kind != ast.KindCallExpression {
		return false
	}
	callee := internalUtils.SkipAssertionsAndParens(callExpr.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	resolveExprToFn := func(expr *ast.Node) (fn *ast.Node, unknown bool) {
		for d := 0; d < 4 && expr != nil; d++ {
			expr = internalUtils.SkipAssertionsAndParens(expr)
			if expr == nil {
				return nil, false
			}
			switch expr.Kind {
			case ast.KindArrowFunction, ast.KindFunctionExpression:
				return expr, false
			case ast.KindIdentifier:
				sym := r.resolveSymbol(expr)
				if sym == nil {
					return nil, false
				}
				decl := r.singleSameFileValueDeclaration(sym)
				if decl == nil {
					return nil, false
				}
				switch decl.Kind {
				case ast.KindFunctionDeclaration:
					return decl, false
				case ast.KindVariableDeclaration:
					if !ast.IsVarConst(decl) && r.hasSymbolWrite(sym) {
						return nil, true
					}
					init := decl.AsVariableDeclaration().Initializer
					if init == nil || nodeContains(init, expr) {
						return nil, false
					}
					expr = init
					continue
				default:
					return nil, false
				}
			default:
				return nil, false
			}
		}
		return nil, false
	}
	if fn, unknown := resolveExprToFn(callee); unknown {
		return true
	} else if fn != nil {
		return r.fnBodyCanReturnPromise(fn, depth)
	}
	if ast.IsAccessExpression(callee) {
		propName, ok := r.accessStaticNameThroughAssertions(callee)
		if !ok {
			return false
		}
		recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
		if recv != nil && recv.Kind == ast.KindIdentifier {
			if sym := r.resolveSymbol(recv); sym != nil {
				if decl := r.singleSameFileValueDeclaration(sym); decl != nil &&
					decl.Kind == ast.KindVariableDeclaration &&
					(ast.IsVarConst(decl) || !r.hasSymbolWrite(sym)) {
					init := internalUtils.SkipAssertionsAndParens(decl.AsVariableDeclaration().Initializer)
					if init != nil && init.Kind == ast.KindObjectLiteralExpression && init.AsObjectLiteralExpression().Properties != nil {
						var matchedProp *ast.Node
						for _, prop := range init.AsObjectLiteralExpression().Properties.Nodes {
							if prop == nil || prop.Kind == ast.KindSpreadAssignment || prop.Name() == nil {
								continue
							}
							name, nameOK := internalUtils.GetStaticPropertyName(prop.Name())
							if nameOK && name == propName {
								matchedProp = prop
							}
						}
						if matchedProp != nil {
							switch matchedProp.Kind {
							case ast.KindMethodDeclaration:
								return r.fnBodyCanReturnPromise(matchedProp, depth)
							case ast.KindPropertyAssignment:
								if fn, unknown := resolveExprToFn(matchedProp.AsPropertyAssignment().Initializer); unknown {
									return true
								} else if fn != nil {
									return r.fnBodyCanReturnPromise(fn, depth)
								}
							case ast.KindShorthandPropertyAssignment:
								if fn, unknown := resolveExprToFn(matchedProp.Name()); unknown {
									return true
								} else if fn != nil {
									return r.fnBodyCanReturnPromise(fn, depth)
								}
							}
						}
					}
				}
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isKnownNonPromiseReturnType(t *checker.Type, callExpr *ast.Node) bool {
	return r.isKnownNonPromiseReturnTypeDepth(t, callExpr, 0)
}

func (r *AsyncAssertionResolver) isKnownNonPromiseReturnTypeDepth(t *checker.Type, callExpr *ast.Node, depth int) bool {
	if depth > 4 || r == nil || r.typeChecker == nil || t == nil {
		return false
	}
	parts := internalUtils.UnionTypeParts(t)
	if len(parts) == 0 {
		return false
	}
	const nonPromisePrimitiveFlags = checker.TypeFlagsStringLike |
		checker.TypeFlagsNumberLike |
		checker.TypeFlagsBooleanLike |
		checker.TypeFlagsBigIntLike |
		checker.TypeFlagsESSymbolLike |
		checker.TypeFlagsUndefined |
		checker.TypeFlagsNull |
		checker.TypeFlagsNever
	checkedVoid := false
	voidReturnsValue := false
	for _, part := range parts {
		if part == nil {
			return false
		}
		flags := part.Flags()
		if flags&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
			return false
		}
		if flags&nonPromisePrimitiveFlags != 0 {
			continue
		}
		if flags&checker.TypeFlagsVoid != 0 {
			if callExpr == nil || callExpr.Kind != ast.KindCallExpression {
				return false
			}
			if !checkedVoid {
				checkedVoid = true
				voidReturnsValue = r.voidCallCanReturnValue(callExpr, depth)
			}
			if voidReturnsValue {
				return false
			}
			continue
		}
		if r.isKnownBuiltinCollectionType(part) {
			continue
		}
		return false
	}
	return true
}

func (r *AsyncAssertionResolver) isLocallyTypeShadowed(node *ast.Node, name string) bool {
	const typeFlags = ast.SymbolFlagsType | ast.SymbolFlagsAlias
	for cur := node; cur != nil; cur = cur.Parent {
		if cur.Kind == ast.KindClassExpression {
			if cn := cur.Name(); cn != nil && cn.Kind == ast.KindIdentifier && cn.AsIdentifier().Text == name {
				return true
			}
		}
		switch cur.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindClassDeclaration, ast.KindClassExpression,
			ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration,
			ast.KindFunctionType, ast.KindConstructorType,
			ast.KindCallSignature, ast.KindConstructSignature, ast.KindMethodSignature:
			for _, tp := range cur.TypeParameters() {
				if tp != nil && tp.Name() != nil && tp.Name().Kind == ast.KindIdentifier && tp.Name().AsIdentifier().Text == name {
					return true
				}
			}
		}
		if !ast.IsLocalsContainer(cur) {
			continue
		}
		if local := ast.GetLocals(cur)[name]; local != nil {
			if local.Flags&typeFlags != 0 {
				return true
			}
			if local.ExportSymbol != nil && local.ExportSymbol.Flags&typeFlags != 0 {
				return true
			}
		}
		if sym := cur.Symbol(); sym != nil && sym.Exports != nil {
			if exp := sym.Exports[name]; exp != nil && exp.Flags&typeFlags != 0 {
				return true
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isSpecificBuiltinTypeNode(typeNode *ast.Node, depth int, allowed ...string) bool {
	if depth > 6 || typeNode == nil {
		return false
	}
	switch typeNode.Kind {
	case ast.KindParenthesizedType:
		return r.isSpecificBuiltinTypeNode(typeNode.AsParenthesizedTypeNode().Type, depth+1, allowed...)
	case ast.KindUnionType:
		ut := typeNode.AsUnionTypeNode()
		if ut == nil || ut.Types == nil || len(ut.Types.Nodes) == 0 {
			return false
		}
		for _, member := range ut.Types.Nodes {
			if !r.isSpecificBuiltinTypeNode(member, depth+1, allowed...) {
				return false
			}
		}
		return true
	case ast.KindTypeReference:
		tr := typeNode.AsTypeReferenceNode()
		if tr == nil || tr.TypeName == nil || tr.TypeName.Kind != ast.KindIdentifier {
			return false
		}
		name := tr.TypeName.AsIdentifier().Text
		for _, want := range allowed {
			if name == want {
				return !r.isLocallyTypeShadowed(tr.TypeName, name)
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isBuiltinCollectionTypeNode(typeNode *ast.Node, depth int) bool {
	if depth > 6 || typeNode == nil {
		return false
	}
	switch typeNode.Kind {
	case ast.KindParenthesizedType:
		return r.isBuiltinCollectionTypeNode(typeNode.AsParenthesizedTypeNode().Type, depth+1)
	case ast.KindTypeOperator:
		op := typeNode.AsTypeOperatorNode()
		return op != nil && op.Operator == ast.KindReadonlyKeyword &&
			r.isBuiltinCollectionTypeNode(op.Type, depth+1)
	case ast.KindArrayType, ast.KindTupleType:
		return !r.isLocallyTypeShadowed(typeNode, "Array")
	case ast.KindUnionType:
		ut := typeNode.AsUnionTypeNode()
		if ut == nil || ut.Types == nil || len(ut.Types.Nodes) == 0 {
			return false
		}
		for _, member := range ut.Types.Nodes {
			if !r.isBuiltinCollectionTypeNode(member, depth+1) {
				return false
			}
		}
		return true
	case ast.KindTypeReference:
		tr := typeNode.AsTypeReferenceNode()
		if tr == nil || tr.TypeName == nil || tr.TypeName.Kind != ast.KindIdentifier {
			return false
		}
		name := tr.TypeName.AsIdentifier().Text
		switch name {
		case "Array", "ReadonlyArray", "Set", "ReadonlySet", "Map", "ReadonlyMap", "WeakSet", "WeakMap":
			return !r.isLocallyTypeShadowed(tr.TypeName, name)
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isKnownBuiltinCollectionExpr(expr *ast.Node, tb *trackedBindings) bool {
	return r.isKnownBuiltinCollectionExprDepth(expr, tb, 0)
}

func (r *AsyncAssertionResolver) isKnownBuiltinCollectionExprDepth(expr *ast.Node, tb *trackedBindings, depth int) bool {
	if depth > 4 {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindArrayLiteralExpression:
		return true
	case ast.KindIdentifier:
		if r.matchesTrackedRestIdent(expr, tb, nil) {
			return true
		}
		if sym := r.resolveSymbol(expr); sym != nil {
			if decl := r.singleSameFileValueDeclaration(sym); decl != nil {
				if decl.Kind == ast.KindParameter && !r.hasSymbolWrite(sym) {
					pd := decl.AsParameterDeclaration()
					if pd.DotDotDotToken != nil || ((r == nil || r.typeChecker == nil) && r.isBuiltinCollectionTypeNode(pd.Type, 0)) {
						return true
					}
				}
				if decl.Kind == ast.KindVariableDeclaration &&
					(ast.IsVarConst(decl) || !r.hasSymbolWrite(sym)) {
					vd := decl.AsVariableDeclaration()
					if vd.Initializer != nil &&
						!nodeContains(vd.Initializer, expr) &&
						r.isKnownBuiltinCollectionExprDepth(vd.Initializer, tb, depth+1) {
						return true
					}
					if (r == nil || r.typeChecker == nil) && r.isBuiltinCollectionTypeNode(vd.Type, 0) {
						return true
					}
				}
			}
		}
	case ast.KindNewExpression:
		if r.isBuiltinNonPromiseContainerCtor(expr.AsNewExpression().Expression) {
			return true
		}
	case ast.KindCallExpression:
		callee := internalUtils.SkipAssertionsAndParens(expr.AsCallExpression().Expression)
		if callee != nil && callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "Array" &&
			!r.isLocallyValueShadowed(callee, "Array") {
			return true
		}
		if callee != nil && ast.IsAccessExpression(callee) {
			recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
			if mName, ok := r.accessStaticNameThroughAssertions(callee); ok {
				if recv != nil && recv.Kind == ast.KindIdentifier && recv.AsIdentifier().Text == "Array" &&
					!r.isLocallyValueShadowed(recv, "Array") && (mName == "from" || mName == "of") {
					return true
				}
				switch mName {
				case "add", "set", "slice", "concat", "filter", "map", "flat", "flatMap":
					if r.isKnownBuiltinCollectionExprDepth(recv, tb, depth+1) {
						return true
					}
				}
			}
		}
	}
	if r != nil && r.typeChecker != nil {
		return r.isKnownBuiltinCollectionType(r.typeChecker.GetTypeAtLocation(expr))
	}
	return false
}

func (r *AsyncAssertionResolver) isSetExpr(expr *ast.Node) bool {
	return r.isSetExprDepth(expr, 0)
}

func (r *AsyncAssertionResolver) isSetExprDepth(expr *ast.Node, depth int) bool {
	if depth > 4 {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindNewExpression:
		ctor := internalUtils.SkipAssertionsAndParens(expr.AsNewExpression().Expression)
		return ctor != nil && ctor.Kind == ast.KindIdentifier && ctor.AsIdentifier().Text == "Set" &&
			!r.isLocallyValueShadowed(ctor, "Set")
	case ast.KindCallExpression:
		callee := internalUtils.SkipAssertionsAndParens(expr.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if mName, ok := r.accessStaticNameThroughAssertions(callee); ok && mName == "add" {
				return r.isSetExprDepth(internalUtils.AccessExpressionObject(callee), depth+1)
			}
		}
	case ast.KindIdentifier:
		if sym := r.resolveSymbol(expr); sym != nil {
			if decl := r.singleSameFileValueDeclaration(sym); decl != nil {
				if decl.Kind == ast.KindParameter && !r.hasSymbolWrite(sym) &&
					(r == nil || r.typeChecker == nil) && r.isSpecificBuiltinTypeNode(decl.AsParameterDeclaration().Type, 0, "Set", "ReadonlySet") {
					return true
				}
				if decl.Kind == ast.KindVariableDeclaration && (ast.IsVarConst(decl) || !r.hasSymbolWrite(sym)) {
					vd := decl.AsVariableDeclaration()
					if vd.Initializer != nil && !nodeContains(vd.Initializer, expr) && r.isSetExprDepth(vd.Initializer, depth+1) {
						return true
					}
					if (r == nil || r.typeChecker == nil) && r.isSpecificBuiltinTypeNode(vd.Type, 0, "Set", "ReadonlySet") {
						return true
					}
				}
			}
		}
	}
	if r != nil && r.typeChecker != nil && r.program != nil {
		t := r.typeChecker.GetTypeAtLocation(expr)
		parts := internalUtils.UnionTypeParts(t)
		if len(parts) > 0 {
			for _, part := range parts {
				if part == nil || part.Symbol() == nil ||
					(part.Symbol().Name != "Set" && part.Symbol().Name != "ReadonlySet") ||
					!internalUtils.IsSymbolFromDefaultLibrary(r.program, part.Symbol()) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isWeakSetExpr(expr *ast.Node) bool {
	return r.isWeakSetExprDepth(expr, 0)
}

func (r *AsyncAssertionResolver) isWeakSetExprDepth(expr *ast.Node, depth int) bool {
	if depth > 4 {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindNewExpression:
		ctor := internalUtils.SkipAssertionsAndParens(expr.AsNewExpression().Expression)
		return ctor != nil && ctor.Kind == ast.KindIdentifier && ctor.AsIdentifier().Text == "WeakSet" &&
			!r.isLocallyValueShadowed(ctor, "WeakSet")
	case ast.KindCallExpression:
		callee := internalUtils.SkipAssertionsAndParens(expr.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if mName, ok := r.accessStaticNameThroughAssertions(callee); ok && mName == "add" {
				return r.isWeakSetExprDepth(internalUtils.AccessExpressionObject(callee), depth+1)
			}
		}
	case ast.KindIdentifier:
		if sym := r.resolveSymbol(expr); sym != nil {
			if decl := r.singleSameFileValueDeclaration(sym); decl != nil {
				if decl.Kind == ast.KindParameter && !r.hasSymbolWrite(sym) &&
					(r == nil || r.typeChecker == nil) && r.isSpecificBuiltinTypeNode(decl.AsParameterDeclaration().Type, 0, "WeakSet") {
					return true
				}
				if decl.Kind == ast.KindVariableDeclaration &&
					(ast.IsVarConst(decl) || !r.hasSymbolWrite(sym)) {
					vd := decl.AsVariableDeclaration()
					if vd.Initializer != nil &&
						!nodeContains(vd.Initializer, expr) &&
						r.isWeakSetExprDepth(vd.Initializer, depth+1) {
						return true
					}
					if (r == nil || r.typeChecker == nil) && r.isSpecificBuiltinTypeNode(vd.Type, 0, "WeakSet") {
						return true
					}
				}
			}
		}
	}
	if r != nil && r.typeChecker != nil && r.program != nil {
		t := r.typeChecker.GetTypeAtLocation(expr)
		parts := internalUtils.UnionTypeParts(t)
		if len(parts) > 0 {
			for _, part := range parts {
				if part == nil || part.Symbol() == nil || part.Symbol().Name != "WeakSet" ||
					!internalUtils.IsSymbolFromDefaultLibrary(r.program, part.Symbol()) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isWeakMapExpr(expr *ast.Node) bool {
	return r.isWeakMapExprDepth(expr, 0)
}

func (r *AsyncAssertionResolver) isWeakMapExprDepth(expr *ast.Node, depth int) bool {
	if depth > 4 {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindNewExpression:
		ctor := internalUtils.SkipAssertionsAndParens(expr.AsNewExpression().Expression)
		return ctor != nil && ctor.Kind == ast.KindIdentifier && ctor.AsIdentifier().Text == "WeakMap" &&
			!r.isLocallyValueShadowed(ctor, "WeakMap")
	case ast.KindCallExpression:
		callee := internalUtils.SkipAssertionsAndParens(expr.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if mName, ok := r.accessStaticNameThroughAssertions(callee); ok && mName == "set" {
				return r.isWeakMapExprDepth(internalUtils.AccessExpressionObject(callee), depth+1)
			}
		}
	case ast.KindIdentifier:
		if sym := r.resolveSymbol(expr); sym != nil {
			if decl := r.singleSameFileValueDeclaration(sym); decl != nil {
				if decl.Kind == ast.KindParameter && !r.hasSymbolWrite(sym) &&
					(r == nil || r.typeChecker == nil) && r.isSpecificBuiltinTypeNode(decl.AsParameterDeclaration().Type, 0, "WeakMap") {
					return true
				}
				if decl.Kind == ast.KindVariableDeclaration &&
					(ast.IsVarConst(decl) || !r.hasSymbolWrite(sym)) {
					vd := decl.AsVariableDeclaration()
					if vd.Initializer != nil &&
						!nodeContains(vd.Initializer, expr) &&
						r.isWeakMapExprDepth(vd.Initializer, depth+1) {
						return true
					}
					if (r == nil || r.typeChecker == nil) && r.isSpecificBuiltinTypeNode(vd.Type, 0, "WeakMap") {
						return true
					}
				}
			}
		}
	}
	if r != nil && r.typeChecker != nil && r.program != nil {
		t := r.typeChecker.GetTypeAtLocation(expr)
		parts := internalUtils.UnionTypeParts(t)
		if len(parts) > 0 {
			for _, part := range parts {
				if part == nil || part.Symbol() == nil || part.Symbol().Name != "WeakMap" ||
					!internalUtils.IsSymbolFromDefaultLibrary(r.program, part.Symbol()) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) getArrayLiteralSlotIndex(arr *ast.Node) *arrayLiteralSlotIndex {
	if arr == nil || arr.Kind != ast.KindArrayLiteralExpression {
		return nil
	}
	if r != nil && r.arraySlotIndices != nil {
		if idx, ok := r.arraySlotIndices[arr]; ok {
			return idx
		}
	}
	a := arr.AsArrayLiteralExpression()
	idx := &arrayLiteralSlotIndex{firstSpreadIdx: -1}
	if a != nil && a.Elements != nil {
		idx.elemCount = len(a.Elements.Nodes)
		for i, el := range a.Elements.Nodes {
			if el == nil || el.Kind == ast.KindOmittedExpression {
				continue
			}
			if el.Kind == ast.KindSpreadElement {
				if idx.firstSpreadIdx < 0 {
					idx.firstSpreadIdx = i
				}
				idx.nonIdentifierSlices = append(idx.nonIdentifierSlices, i)
				continue
			}
			unwrapped := internalUtils.SkipAssertionsAndParens(el)
			if unwrapped != nil && unwrapped.Kind == ast.KindIdentifier && isIdentifierValueReference(unwrapped) {
				name := unwrapped.AsIdentifier().Text
				if r != nil && r.refs != nil {
					if sym := r.resolveSymbol(unwrapped); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
						if idx.bySym == nil {
							idx.bySym = map[*ast.Symbol][]int{}
						}
						idx.bySym[sym] = append(idx.bySym[sym], i)
						continue
					}
				}
				if idx.byName == nil {
					idx.byName = map[string][]int{}
				}
				idx.byName[name] = append(idx.byName[name], i)
				continue
			}
			idx.nonIdentifierSlices = append(idx.nonIdentifierSlices, i)
		}
	}
	if r != nil {
		if r.arraySlotIndices == nil {
			r.arraySlotIndices = map[*ast.Node]*arrayLiteralSlotIndex{}
		}
		r.arraySlotIndices[arr] = idx
	}
	return idx
}

func (r *AsyncAssertionResolver) isSymbolIteratorPropertyName(name *ast.Node) bool {
	if name == nil || name.Kind != ast.KindComputedPropertyName {
		return false
	}
	expr := internalUtils.SkipAssertionsAndParens(name.AsComputedPropertyName().Expression)
	if expr == nil || !ast.IsAccessExpression(expr) {
		return false
	}
	propName, ok := r.accessStaticNameThroughAssertions(expr)
	if !ok || propName != "iterator" {
		return false
	}
	obj := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(expr))
	return obj != nil && obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == "Symbol" &&
		!r.isLocallyValueShadowed(obj, "Symbol")
}

func (r *AsyncAssertionResolver) isNullOrUndefinedExpr(expr *ast.Node) bool {
	u := internalUtils.SkipAssertionsAndParens(expr)
	if u == nil {
		return false
	}
	return u.Kind == ast.KindNullKeyword || r.isDefinitelyUndefinedExpr(u)
}

func (r *AsyncAssertionResolver) symbolIteratorPropYieldsTracked(prop *ast.Node, tb *trackedBindings, depth int) bool {
	if prop == nil {
		return false
	}
	var fn *ast.Node
	switch prop.Kind {
	case ast.KindMethodDeclaration:
		fn = prop
	case ast.KindPropertyAssignment:
		init := internalUtils.SkipAssertionsAndParens(prop.AsPropertyAssignment().Initializer)
		if r.isNullOrUndefinedExpr(init) {
			return false
		}
		if init != nil && (init.Kind == ast.KindFunctionExpression || init.Kind == ast.KindArrowFunction) {
			fn = init
		}
	}
	if fn != nil && functionAsteriskToken(fn) != nil {
		body := fn.Body()
		if body == nil {
			return false
		}
		innerTB := cloneTrackedBindings(tb, true)
		if !innerTB.isEmpty() {
			r.propagateLocalAliasesInScope(body, innerTB)
		}
		foundYield := false
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			if node == nil || foundYield || ast.IsTypeNode(node) {
				return
			}
			if node != body && testFramework.IsFunction(node) {
				return
			}
			if node.Kind == ast.KindYieldExpression {
				ye := node.AsYieldExpression()
				if ye != nil && ye.Expression != nil {
					if ye.AsteriskToken != nil {
						if r.exprCanBeTrackedContainerDepth(ye.Expression, innerTB, depth+1, true) {
							foundYield = true
							return
						}
					} else if r.exprCanReturnTrackedPromiseDepth(ye.Expression, innerTB, depth+1) {
						foundYield = true
						return
					}
				}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				visit(child)
				return foundYield
			})
		}
		visit(body)
		return foundYield
	}
	return r.exprReferencesTracked(prop, tb, true)
}

func (r *AsyncAssertionResolver) isArrayLikeObjectLiteralWithTracked(expr *ast.Node, tb *trackedBindings, depth int) bool {
	if depth > 6 {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil || expr.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	obj := expr.AsObjectLiteralExpression()
	if obj == nil || obj.Properties == nil {
		return false
	}
	hasLength := false
	lengthVal := -1
	hasSpread := false
	hasSymbolIterator := false
	var slotTracked map[int]bool
	for _, prop := range obj.Properties.Nodes {
		if prop == nil {
			continue
		}
		if prop.Kind == ast.KindSpreadAssignment {
			hasSpread = true
			continue
		}
		pName := prop.Name()
		if pName == nil {
			continue
		}
		if r.isSymbolIteratorPropertyName(pName) {
			if prop.Kind == ast.KindPropertyAssignment && r.isNullOrUndefinedExpr(prop.AsPropertyAssignment().Initializer) {
				hasSymbolIterator = false
			} else {
				hasSymbolIterator = true
			}
			continue
		}
		key, ok := internalUtils.GetStaticPropertyName(pName)
		if !ok && r != nil && r.staticStringEvaluator() != nil {
			key, ok = r.staticStringEvaluator().EvalPropertyName(pName)
		}
		if !ok {
			continue
		}
		if key == "length" {
			hasLength = true
			lengthVal = -1
			if prop.Kind == ast.KindPropertyAssignment {
				val := internalUtils.SkipAssertionsAndParens(prop.AsPropertyAssignment().Initializer)
				if val != nil && val.Kind == ast.KindNumericLiteral {
					if n, err := strconv.Atoi(val.AsNumericLiteral().Text); err == nil && n >= 0 {
						lengthVal = n
					}
				} else if val != nil && val.Kind == ast.KindPrefixUnaryExpression &&
					val.AsPrefixUnaryExpression().Operator == ast.KindMinusToken {
					op := internalUtils.SkipAssertionsAndParens(val.AsPrefixUnaryExpression().Operand)
					if op != nil && op.Kind == ast.KindNumericLiteral {
						lengthVal = 0
					}
				}
			}
			continue
		}
		idx, err := strconv.Atoi(key)
		if err != nil || idx < 0 || strconv.Itoa(idx) != key {
			continue
		}
		matched := false
		switch prop.Kind {
		case ast.KindPropertyAssignment:
			matched = r.exprCanReturnTrackedPromiseDepth(prop.AsPropertyAssignment().Initializer, tb, depth+1)
		case ast.KindShorthandPropertyAssignment:
			matched = r.exprCanReturnTrackedPromiseDepth(prop.Name(), tb, depth+1)
		case ast.KindGetAccessor:
			matched = r.exprReferencesTracked(prop.Body(), tb, true)
		}
		if slotTracked == nil {
			slotTracked = map[int]bool{}
		}
		slotTracked[idx] = matched
	}
	// In JS, `Array.from` always uses `[Symbol.iterator]` when present and ignores numeric slots.
	if hasSymbolIterator {
		return false
	}
	if !hasLength && !hasSpread {
		return false
	}
	for idx, ok := range slotTracked {
		if ok && (lengthVal < 0 || idx < lengthVal) {
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) evalMapStaticKey(expr *ast.Node) (string, bool) {
	return r.evalMapStaticKeyDepth(expr, 0)
}

func (r *AsyncAssertionResolver) evalMapStaticKeyDepth(expr *ast.Node, depth int) (string, bool) {
	if depth > 4 {
		return "", false
	}
	u := internalUtils.SkipAssertionsAndParens(expr)
	if u == nil {
		return "", false
	}
	switch u.Kind {
	case ast.KindStringLiteral:
		return "s:" + u.AsStringLiteral().Text, true
	case ast.KindNoSubstitutionTemplateLiteral:
		return "s:" + u.Text(), true
	case ast.KindNumericLiteral:
		return "n:" + internalUtils.NormalizeNumericLiteral(u.AsNumericLiteral().Text), true
	case ast.KindBigIntLiteral:
		return "bi:" + internalUtils.NormalizeBigIntLiteral(u.Text()), true
	case ast.KindPrefixUnaryExpression:
		pref := u.AsPrefixUnaryExpression()
		if pref != nil && (pref.Operator == ast.KindMinusToken || pref.Operator == ast.KindPlusToken) {
			op := internalUtils.SkipAssertionsAndParens(pref.Operand)
			if op != nil {
				switch op.Kind {
				case ast.KindNumericLiteral:
					norm := internalUtils.NormalizeNumericLiteral(op.AsNumericLiteral().Text)
					if pref.Operator == ast.KindMinusToken && norm != "0" {
						return "n:-" + norm, true
					}
					return "n:" + norm, true
				case ast.KindBigIntLiteral:
					if pref.Operator == ast.KindMinusToken {
						norm := internalUtils.NormalizeBigIntLiteral(op.Text())
						if norm != "0" {
							return "bi:-" + norm, true
						}
						return "bi:0", true
					}
				case ast.KindIdentifier:
					if op.AsIdentifier().Text == "Infinity" && !r.isLocallyValueShadowed(op, "Infinity") {
						if pref.Operator == ast.KindMinusToken {
							return "n:-Infinity", true
						}
						return "n:Infinity", true
					}
				}
			}
		}
	case ast.KindTrueKeyword:
		return "b:true", true
	case ast.KindFalseKeyword:
		return "b:false", true
	case ast.KindNullKeyword:
		return "null", true
	case ast.KindUndefinedKeyword, ast.KindVoidExpression:
		return "undef", true
	case ast.KindIdentifier:
		name := u.AsIdentifier().Text
		if name == "undefined" && !r.isLocallyValueShadowed(u, "undefined") {
			return "undef", true
		}
		if name == "NaN" && !r.isLocallyValueShadowed(u, "NaN") {
			return "n:NaN", true
		}
		if name == "Infinity" && !r.isLocallyValueShadowed(u, "Infinity") {
			return "n:Infinity", true
		}
		if r != nil {
			if sym := r.resolveSymbol(u); sym != nil {
				if decl := r.singleSameFileValueDeclaration(sym); decl != nil &&
					decl.Kind == ast.KindVariableDeclaration &&
					(ast.IsVarConst(decl) || !r.hasSymbolWrite(sym)) {
					if init := decl.AsVariableDeclaration().Initializer; init != nil && !nodeContains(init, u) {
						if key, ok := r.evalMapStaticKeyDepth(init, depth+1); ok {
							return key, true
						}
					}
				}
			}
		}
	}
	if r != nil && r.staticStringEvaluator() != nil {
		if val, ok := r.staticStringEvaluator().EvalValue(u); ok {
			switch v := val.(type) {
			case string:
				return "s:" + v, true
			case bool:
				if v {
					return "b:true", true
				}
				return "b:false", true
			case *big.Int:
				return "bi:" + v.String(), true
			}
		}
	}
	return "", false
}

// extractTrackedMapKeys analyzes `expr` as a Map expression and returns the
// tracked static keys (when statically enumerable), whether any dynamic/unknown
// key holds a tracked Promise, and whether `expr` currently holds a tracked Promise.
func (r *AsyncAssertionResolver) extractTrackedMapKeys(
	expr *ast.Node,
	tb *trackedBindings,
	depth int,
	requirePopulated bool,
) (staticKeys map[string]bool, hasDynamic bool, matched bool) {
	if depth > 6 {
		return nil, false, false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	for expr != nil && internalUtils.IsCommaOperator(expr) {
		expr = internalUtils.SkipAssertionsAndParens(expr.AsBinaryExpression().Right)
	}
	if expr == nil {
		return nil, false, false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		if !r.matchesTrackedMapIdent(expr, tb, nil, false) {
			return nil, false, false
		}
		populated, keys, dyn := r.survivingContainerWritesAt(expr, tb)
		if requirePopulated && !populated {
			return nil, false, false
		}
		if !populated || dyn {
			return nil, true, true
		}
		return keys, false, len(keys) > 0
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		if cond == nil {
			return nil, false, false
		}
		k1, d1, m1 := r.extractTrackedMapKeys(cond.WhenTrue, tb, depth+1, requirePopulated)
		k2, d2, m2 := r.extractTrackedMapKeys(cond.WhenFalse, tb, depth+1, requirePopulated)
		if !m1 && !m2 {
			return nil, false, false
		}
		if d1 || d2 {
			return nil, true, true
		}
		merged := map[string]bool{}
		for k := range k1 {
			merged[k] = true
		}
		for k := range k2 {
			merged[k] = true
		}
		return merged, false, len(merged) > 0
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return nil, false, false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
			return r.extractTrackedMapKeys(bin.Right, tb, depth+1, requirePopulated)
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken,
			ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			k1, d1, m1 := r.extractTrackedMapKeys(bin.Left, tb, depth+1, requirePopulated)
			k2, d2, m2 := r.extractTrackedMapKeys(bin.Right, tb, depth+1, requirePopulated)
			if !m1 && !m2 {
				return nil, false, false
			}
			if d1 || d2 {
				return nil, true, true
			}
			merged := map[string]bool{}
			for k := range k1 {
				merged[k] = true
			}
			for k := range k2 {
				merged[k] = true
			}
			return merged, false, len(merged) > 0
		}
		return nil, false, false
	case ast.KindNewExpression:
		ctor := internalUtils.SkipAssertionsAndParens(expr.AsNewExpression().Expression)
		if ctor == nil || ctor.Kind != ast.KindIdentifier || ctor.AsIdentifier().Text != "Map" ||
			r.isLocallyValueShadowed(ctor, "Map") {
			return nil, false, false
		}
		arg0, argOK := staticCallArgAt(expr, 0)
		if !argOK {
			if r.exprReferencesTracked(expr, tb, true) {
				return nil, true, true
			}
			return nil, false, false
		}
		if arg0 == nil {
			return nil, false, false
		}
		arg0 = internalUtils.SkipAssertionsAndParens(arg0)
		if arg0 == nil {
			return nil, false, false
		}
		if arg0.Kind == ast.KindArrayLiteralExpression {
			arr := arg0.AsArrayLiteralExpression()
			if arr == nil || arr.Elements == nil {
				return nil, false, false
			}
			keys := map[string]bool{}
			dyn := false
			for _, el := range arr.Elements.Nodes {
				if el == nil || el.Kind == ast.KindOmittedExpression {
					continue
				}
				if el.Kind == ast.KindSpreadElement {
					if r.exprReferencesTracked(el.AsSpreadElement().Expression, tb, true) {
						dyn = true
					}
					continue
				}
				entry := internalUtils.SkipAssertionsAndParens(el)
				if entry == nil {
					continue
				}
				if entry.Kind == ast.KindArrayLiteralExpression {
					pair := entry.AsArrayLiteralExpression()
					if pair == nil || pair.Elements == nil || len(pair.Elements.Nodes) == 0 {
						continue
					}
					elems := pair.Elements.Nodes
					if elems[0] != nil && elems[0].Kind == ast.KindSpreadElement {
						if r.exprReferencesTracked(entry, tb, true) {
							dyn = true
						}
						continue
					}
					sk, hasKey := r.evalMapStaticKey(elems[0])
					valTracked := false
					if len(elems) >= 2 && elems[1] != nil && elems[1].Kind != ast.KindOmittedExpression {
						if elems[1].Kind == ast.KindSpreadElement {
							valTracked = r.exprReferencesTracked(elems[1], tb, true)
						} else {
							valTracked = r.exprCanReturnTrackedPromiseDepth(elems[1], tb, depth+1)
						}
					}
					if hasKey {
						if valTracked {
							keys[sk] = true
						} else {
							delete(keys, sk)
						}
					} else if valTracked {
						dyn = true
					}
					continue
				}
				if !r.matchesTrackedDirectIdent(entry, tb, nil) && r.exprReferencesTracked(entry, tb, true) {
					dyn = true
				}
			}
			if dyn {
				return nil, true, true
			}
			if len(keys) > 0 {
				return keys, false, true
			}
			return nil, false, false
		}
		if k, d, m := r.extractTrackedMapKeys(arg0, tb, depth+1, requirePopulated); m {
			return k, d, true
		}
		if arg0.Kind == ast.KindCallExpression {
			callee := internalUtils.SkipAssertionsAndParens(arg0.AsCallExpression().Expression)
			if callee != nil && ast.IsAccessExpression(callee) {
				if mName, ok := r.accessStaticNameThroughAssertions(callee); ok && mName == "entries" {
					return r.extractTrackedMapKeys(internalUtils.AccessExpressionObject(callee), tb, depth+1, requirePopulated)
				}
			}
		}
		return nil, false, false
	case ast.KindCallExpression:
		callee := internalUtils.SkipAssertionsAndParens(expr.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if mName, ok := r.accessStaticNameThroughAssertions(callee); ok && mName == "set" {
				recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
				if r.isWeakMapExpr(recv) {
					return nil, false, false
				}
				if r.isKnownBuiltinCollectionExpr(recv, tb) {
					recvKeys, recvDyn, _ := r.extractTrackedMapKeys(recv, tb, depth+1, requirePopulated)
					keys := map[string]bool{}
					for k := range recvKeys {
						keys[k] = true
					}
					dyn := recvDyn
					arg0, arg0OK := staticCallArgAt(expr, 0)
					arg1, arg1OK := staticCallArgAt(expr, 1)
					if !arg0OK || !arg1OK {
						if dyn || r.exprReferencesTracked(expr, tb, true) {
							return nil, true, true
						}
						return keys, false, len(keys) > 0
					}
					valTracked := arg1 != nil && r.exprCanReturnTrackedPromiseDepth(arg1, tb, depth+1)
					if sk, hasKey := r.evalMapStaticKey(arg0); hasKey {
						if valTracked {
							keys[sk] = true
						} else {
							delete(keys, sk)
						}
					} else if valTracked {
						dyn = true
					}
					if dyn {
						return nil, true, true
					}
					if len(keys) > 0 {
						return keys, false, true
					}
					return nil, false, false
				}
			}
		}
	}
	return nil, false, false
}

func (r *AsyncAssertionResolver) exprCanBeTrackedMapDepth(expr *ast.Node, tb *trackedBindings, depth int, requirePopulated bool) bool {
	_, _, matched := r.extractTrackedMapKeys(expr, tb, depth, requirePopulated)
	return matched
}

func (r *AsyncAssertionResolver) exprCanBeTrackedContainer(expr *ast.Node, tb *trackedBindings) bool {
	return r.exprCanBeTrackedContainerDepth(expr, tb, 0, true)
}

// exprCanBeTrackedContainerDepth reports whether `expr` evaluates to an array
// or iterable holding a tracked Promise (e.g. `[p]`, `saved`, `...checks`,
// `[...saved]`, `new Set([p])`, `new Set().add(p)`, `map.values()`,
// `Array.from({ 0: p, length: 1 })`, `{ *[Symbol.iterator]() { yield p; } }`)
// that `Promise.all`/`allSettled`/`race`/`any` or `for await` awaits.
func (r *AsyncAssertionResolver) exprCanBeTrackedContainerDepth(expr *ast.Node, tb *trackedBindings, depth int, requirePopulated bool) bool {
	if depth > 6 {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	for expr != nil && internalUtils.IsCommaOperator(expr) {
		expr = internalUtils.SkipAssertionsAndParens(expr.AsBinaryExpression().Right)
	}
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		return r.matchesTrackedRestIdent(expr, tb, nil) || r.matchesTrackedContainerIdent(expr, tb, nil, requirePopulated)
	case ast.KindArrayLiteralExpression:
		slotIdx := r.getArrayLiteralSlotIndex(expr)
		if slotIdx == nil {
			return false
		}
		for sym := range tb.symbols {
			if len(slotIdx.bySym[sym]) > 0 {
				return true
			}
		}
		for name := range tb.names {
			if len(slotIdx.byName[name]) > 0 {
				return true
			}
		}
		if len(slotIdx.nonIdentifierSlices) == 0 {
			return false
		}
		arr := expr.AsArrayLiteralExpression()
		if arr == nil || arr.Elements == nil {
			return false
		}
		for _, pos := range slotIdx.nonIdentifierSlices {
			el := arr.Elements.Nodes[pos]
			if el == nil || el.Kind == ast.KindOmittedExpression {
				continue
			}
			if el.Kind == ast.KindSpreadElement {
				if r.exprCanBeTrackedContainerDepth(el.AsSpreadElement().Expression, tb, depth+1, requirePopulated) {
					return true
				}
				continue
			}
			if r.exprCanReturnTrackedPromiseDepth(el, tb, depth+1) {
				return true
			}
		}
		return false
	case ast.KindObjectLiteralExpression:
		obj := expr.AsObjectLiteralExpression()
		if obj == nil || obj.Properties == nil {
			return false
		}
		var lastIterProp *ast.Node
		for _, prop := range obj.Properties.Nodes {
			if prop == nil || prop.Kind == ast.KindSpreadAssignment {
				continue
			}
			if r.isSymbolIteratorPropertyName(prop.Name()) {
				lastIterProp = prop
			}
		}
		return lastIterProp != nil && r.symbolIteratorPropYieldsTracked(lastIterProp, tb, depth)
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		if cond == nil {
			return false
		}
		return r.exprCanBeTrackedContainerDepth(cond.WhenTrue, tb, depth+1, requirePopulated) ||
			r.exprCanBeTrackedContainerDepth(cond.WhenFalse, tb, depth+1, requirePopulated)
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
			return r.exprCanBeTrackedContainerDepth(bin.Right, tb, depth+1, requirePopulated)
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken,
			ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return r.exprCanBeTrackedContainerDepth(bin.Left, tb, depth+1, requirePopulated) ||
				r.exprCanBeTrackedContainerDepth(bin.Right, tb, depth+1, requirePopulated)
		}
		return false
	case ast.KindCallExpression:
		if isConsoleCall(expr) {
			return false
		}
		if _, ok := r.promiseMethodName(expr); ok {
			return false
		}
		callee := internalUtils.SkipAssertionsAndParens(expr.AsCallExpression().Expression)
		if callee != nil && callee.Kind == ast.KindIdentifier && callee.AsIdentifier().Text == "Array" &&
			!r.isLocallyValueShadowed(callee, "Array") {
			return r.argsCanProduceTrackedContainer(expr.AsCallExpression().Arguments, tb, depth+1)
		}
		if callee != nil && ast.IsAccessExpression(callee) {
			recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
			if mName, ok := r.accessStaticNameThroughAssertions(callee); ok {
				if recv != nil && recv.Kind == ast.KindIdentifier && recv.AsIdentifier().Text == "Array" &&
					!r.isLocallyValueShadowed(recv, "Array") {
					switch mName {
					case "of":
						return r.argsCanProduceTrackedContainer(expr.AsCallExpression().Arguments, tb, depth+1)
					case "from":
						if args := expr.AsCallExpression().Arguments; args != nil && len(args.Nodes) == 1 && args.Nodes[0] != nil && args.Nodes[0].Kind != ast.KindSpreadElement {
							return r.exprCanBeTrackedContainerDepth(args.Nodes[0], tb, depth+1, requirePopulated) ||
								r.isArrayLikeObjectLiteralWithTracked(args.Nodes[0], tb, depth+1)
						}
					}
				}
				switch mName {
				case "then", "catch", "finally":
					return false
				case "values":
					if r.isWeakMapExpr(recv) || r.isWeakSetExpr(recv) {
						return false
					}
					if r.exprCanBeTrackedMapDepth(recv, tb, depth+1, requirePopulated) ||
						r.exprCanBeTrackedContainerDepth(recv, tb, depth+1, requirePopulated) {
						return true
					}
					if r.isKnownBuiltinCollectionExpr(recv, tb) {
						return false
					}
				case "keys":
					if r.isSetExpr(recv) {
						return r.exprCanBeTrackedContainerDepth(recv, tb, depth+1, requirePopulated)
					}
					if r.isKnownBuiltinCollectionExpr(recv, tb) {
						return false
					}
				case "add":
					if r.isWeakSetExpr(recv) {
						return false
					}
					if r.isKnownBuiltinCollectionExpr(recv, tb) {
						if r.exprCanBeTrackedContainerDepth(recv, tb, depth+1, requirePopulated) {
							return true
						}
						arg0, argOK := staticCallArgAt(expr, 0)
						if !argOK {
							return r.exprReferencesTracked(expr, tb, true)
						}
						return arg0 != nil && r.exprCanReturnTrackedPromiseDepth(arg0, tb, depth+1)
					}
				case "set", "push", "unshift", "forEach", "includes", "indexOf", "lastIndexOf",
					"join", "toString", "every", "some", "has", "delete", "clear", "entries":
					if r.isKnownBuiltinCollectionExpr(recv, tb) {
						return false
					}
				}
			}
		}
		return r.exprReferencesTracked(expr, tb, true)
	case ast.KindNewExpression:
		ctor := internalUtils.SkipAssertionsAndParens(expr.AsNewExpression().Expression)
		if r.isBuiltinNonPromiseContainerCtor(ctor) {
			switch ctor.AsIdentifier().Text {
			case "Set":
				arg0, argOK := staticCallArgAt(expr, 0)
				if !argOK {
					return r.exprReferencesTracked(expr, tb, true)
				}
				return arg0 != nil && r.exprCanBeTrackedContainerDepth(arg0, tb, depth+1, requirePopulated)
			case "Array":
				return r.argsCanProduceTrackedContainer(expr.AsNewExpression().Arguments, tb, depth+1)
			case "Map", "WeakSet", "WeakMap":
				return false
			}
		}
		return r.exprReferencesTracked(expr, tb, true)
	}
	return false
}

func (r *AsyncAssertionResolver) argsCanProduceTrackedContainer(args *ast.NodeList, tb *trackedBindings, depth int) bool {
	if args == nil {
		return false
	}
	for _, a := range args.Nodes {
		if a == nil {
			continue
		}
		if a.Kind == ast.KindSpreadElement {
			if r.exprCanBeTrackedContainerDepth(a.AsSpreadElement().Expression, tb, depth, true) {
				return true
			}
			continue
		}
		if r.exprCanReturnTrackedPromiseDepth(a, tb, depth) {
			return true
		}
	}
	return false
}

// exprCanReturnTrackedPromise reports whether evaluating `expr` can produce a
// Promise/thenable derived from a binding in `tb` (rejecting primitive-only
// expressions such as `void p`, `!p`, `p ? 1 : 2`, `p.message`, `args.length`,
// `console.log(p)`, uncalled closures `() => p`, array literals `[p]`,
// `Promise.resolve([p])`, and non-thenable objects `{ p }` / `{ ...p }`).
func (r *AsyncAssertionResolver) exprCanReturnTrackedPromise(expr *ast.Node, tb *trackedBindings) bool {
	return r.exprCanReturnTrackedPromiseDepth(expr, tb, 0)
}

func cloneTrackedBindings(tb *trackedBindings, clearArguments bool) *trackedBindings {
	if tb == nil {
		return &trackedBindings{restOffset: -1}
	}
	out := &trackedBindings{
		restOffset:     tb.restOffset,
		trackArguments: tb.trackArguments && !clearArguments,
	}
	if len(tb.symbols) > 0 {
		out.symbols = make(map[*ast.Symbol]bool, len(tb.symbols))
		for k, v := range tb.symbols {
			out.symbols[k] = v
		}
	}
	if len(tb.names) > 0 {
		out.names = make(map[string]bool, len(tb.names))
		for k, v := range tb.names {
			out.names[k] = v
		}
	}
	if len(tb.restSymbols) > 0 {
		out.restSymbols = make(map[*ast.Symbol]bool, len(tb.restSymbols))
		for k, v := range tb.restSymbols {
			out.restSymbols[k] = v
		}
	}
	if len(tb.restNames) > 0 {
		out.restNames = make(map[string]bool, len(tb.restNames))
		for k, v := range tb.restNames {
			out.restNames[k] = v
		}
	}
	if len(tb.containerSymbols) > 0 {
		out.containerSymbols = make(map[*ast.Symbol]bool, len(tb.containerSymbols))
		for k, v := range tb.containerSymbols {
			out.containerSymbols[k] = v
		}
	}
	if len(tb.containerNames) > 0 {
		out.containerNames = make(map[string]bool, len(tb.containerNames))
		for k, v := range tb.containerNames {
			out.containerNames[k] = v
		}
	}
	if len(tb.mapSymbols) > 0 {
		out.mapSymbols = make(map[*ast.Symbol]bool, len(tb.mapSymbols))
		for k, v := range tb.mapSymbols {
			out.mapSymbols[k] = v
		}
	}
	if len(tb.mapNames) > 0 {
		out.mapNames = make(map[string]bool, len(tb.mapNames))
		for k, v := range tb.mapNames {
			out.mapNames[k] = v
		}
	}
	if len(tb.arraySlotSymbols) > 0 {
		out.arraySlotSymbols = make(map[*ast.Symbol]map[int]bool, len(tb.arraySlotSymbols))
		for sym, slots := range tb.arraySlotSymbols {
			cp := make(map[int]bool, len(slots))
			for idx, v := range slots {
				cp[idx] = v
			}
			out.arraySlotSymbols[sym] = cp
		}
	}
	if len(tb.arraySlotNames) > 0 {
		out.arraySlotNames = make(map[string]map[int]bool, len(tb.arraySlotNames))
		for name, slots := range tb.arraySlotNames {
			cp := make(map[int]bool, len(slots))
			for idx, v := range slots {
				cp[idx] = v
			}
			out.arraySlotNames[name] = cp
		}
	}
	if len(tb.arraySlotDeclPos) > 0 {
		out.arraySlotDeclPos = make(map[*ast.Symbol]int, len(tb.arraySlotDeclPos))
		for k, v := range tb.arraySlotDeclPos {
			out.arraySlotDeclPos[k] = v
		}
	}
	if len(tb.arraySlotDeclNamePos) > 0 {
		out.arraySlotDeclNamePos = make(map[string]int, len(tb.arraySlotDeclNamePos))
		for k, v := range tb.arraySlotDeclNamePos {
			out.arraySlotDeclNamePos[k] = v
		}
	}
	if len(tb.arraySlotLen) > 0 {
		out.arraySlotLen = make(map[*ast.Symbol]int, len(tb.arraySlotLen))
		for k, v := range tb.arraySlotLen {
			out.arraySlotLen[k] = v
		}
	}
	if len(tb.arraySlotNameLen) > 0 {
		out.arraySlotNameLen = make(map[string]int, len(tb.arraySlotNameLen))
		for k, v := range tb.arraySlotNameLen {
			out.arraySlotNameLen[k] = v
		}
	}
	if len(tb.containerScopes) > 0 {
		out.containerScopes = append([]*ast.Node(nil), tb.containerScopes...)
	}
	if len(tb.containerWrites) > 0 {
		out.containerWrites = make(map[*ast.Symbol][]containerWrite, len(tb.containerWrites))
		for k, v := range tb.containerWrites {
			out.containerWrites[k] = append([]containerWrite(nil), v...)
		}
	}
	if len(tb.containerNameWrites) > 0 {
		out.containerNameWrites = make(map[string][]containerWrite, len(tb.containerNameWrites))
		for k, v := range tb.containerNameWrites {
			out.containerNameWrites[k] = append([]containerWrite(nil), v...)
		}
	}
	if len(tb.initialContainers) > 0 {
		out.initialContainers = make(map[*ast.Symbol]bool, len(tb.initialContainers))
		for k, v := range tb.initialContainers {
			out.initialContainers[k] = v
		}
	}
	if len(tb.initialContainerNames) > 0 {
		out.initialContainerNames = make(map[string]bool, len(tb.initialContainerNames))
		for k, v := range tb.initialContainerNames {
			out.initialContainerNames[k] = v
		}
	}
	if len(tb.containerPeers) > 0 {
		out.containerPeers = make(map[*ast.Symbol][]containerPeerLink, len(tb.containerPeers))
		for k, v := range tb.containerPeers {
			out.containerPeers[k] = append([]containerPeerLink(nil), v...)
		}
	}
	if len(tb.containerNamePeers) > 0 {
		out.containerNamePeers = make(map[string][]containerNamePeerLink, len(tb.containerNamePeers))
		for k, v := range tb.containerNamePeers {
			out.containerNamePeers[k] = append([]containerNamePeerLink(nil), v...)
		}
	}
	if len(tb.argumentsAliasSymbols) > 0 {
		out.argumentsAliasSymbols = make(map[*ast.Symbol]bool, len(tb.argumentsAliasSymbols))
		for k, v := range tb.argumentsAliasSymbols {
			out.argumentsAliasSymbols[k] = v
		}
	}
	if len(tb.argumentsAliasNames) > 0 {
		out.argumentsAliasNames = make(map[string]bool, len(tb.argumentsAliasNames))
		for k, v := range tb.argumentsAliasNames {
			out.argumentsAliasNames[k] = v
		}
	}
	return out
}

func (r *AsyncAssertionResolver) isBuiltinNonPromiseContainerCtor(expr *ast.Node) bool {
	ctor := internalUtils.SkipAssertionsAndParens(expr)
	if ctor == nil || ctor.Kind != ast.KindIdentifier {
		return false
	}
	name := ctor.AsIdentifier().Text
	switch name {
	case "Set", "Map", "WeakSet", "WeakMap", "Array":
		return !r.isLocallyValueShadowed(ctor, name)
	}
	return false
}

func (r *AsyncAssertionResolver) exprCanReturnTrackedPromiseDepth(expr *ast.Node, tb *trackedBindings, depth int) bool {
	if depth > 6 {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	for expr != nil && internalUtils.IsCommaOperator(expr) {
		expr = internalUtils.SkipAssertionsAndParens(expr.AsBinaryExpression().Right)
	}
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		return r.matchesTrackedDirectIdent(expr, tb, nil)
	case ast.KindAwaitExpression:
		return r.exprCanReturnTrackedPromiseDepth(expr.AsAwaitExpression().Expression, tb, depth+1)
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		if cond == nil {
			return false
		}
		return r.exprCanReturnTrackedPromiseDepth(cond.WhenTrue, tb, depth+1) ||
			r.exprCanReturnTrackedPromiseDepth(cond.WhenFalse, tb, depth+1)
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
			return r.exprCanReturnTrackedPromiseDepth(bin.Right, tb, depth+1)
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken,
			ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return r.exprCanReturnTrackedPromiseDepth(bin.Left, tb, depth+1) ||
				r.exprCanReturnTrackedPromiseDepth(bin.Right, tb, depth+1)
		}
		return false
	case ast.KindElementAccessExpression:
		return r.isRestElementAccess(expr, tb) || r.isTrackedArraySlotAccess(expr, tb)
	case ast.KindCallExpression:
		if isConsoleCall(expr) {
			return false
		}
		if method, ok := r.promiseMethodName(expr); ok {
			switch method {
			case "reject", "withResolvers":
				return false
			case "resolve":
				arg0, argOK := staticCallArgAt(expr, 0)
				if !argOK {
					return r.exprReferencesTracked(expr, tb, true)
				}
				return arg0 != nil && r.exprCanReturnTrackedPromiseDepth(arg0, tb, depth+1)
			case "all", "allSettled", "race", "any":
				arg0, argOK := staticCallArgAt(expr, 0)
				if !argOK {
					return r.exprReferencesTracked(expr, tb, true)
				}
				return arg0 != nil && r.exprCanBeTrackedContainerDepth(arg0, tb, depth+1, true)
			}
		}
		callee := internalUtils.SkipAssertionsAndParens(expr.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if mName, ok := r.accessStaticNameThroughAssertions(callee); ok {
				switch mName {
				case "then", "catch", "finally":
					recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
					if r.exprCanReturnTrackedPromiseDepth(recv, tb, depth+1) {
						return true
					}
					// If the receiver is a non-tracked Promise (e.g. `Promise.resolve([p])`),
					// only treat the chain as returning a tracked Promise when a handler
					// consumes it (either by capturing `tb` or by unwrapping a resolved container).
					if args := expr.AsCallExpression().Arguments; args != nil {
						var resolvedContainerArg *ast.Node
						if mName == "then" && recv != nil && recv.Kind == ast.KindCallExpression {
							if pMethod, pOK := r.promiseMethodName(recv); pOK && pMethod == "resolve" {
								if a0, aOK := staticCallArgAt(recv, 0); aOK && a0 != nil && r.exprCanBeTrackedContainerDepth(a0, tb, depth+1, true) {
									resolvedContainerArg = internalUtils.SkipAssertionsAndParens(a0)
								}
							}
						}
						for i, a := range args.Nodes {
							if a == nil {
								continue
							}
							h := internalUtils.SkipAssertionsAndParens(a)
							if h == nil || r.isNonCallablePromiseHandler(h) {
								continue
							}
							if h.Kind == ast.KindArrowFunction || h.Kind == ast.KindFunctionExpression {
								if functionAsteriskToken(h) != nil {
									continue
								}
								outerTB := cloneTrackedBindings(tb, h.Kind == ast.KindFunctionExpression)
								if !outerTB.isEmpty() {
									r.propagateLocalAliasesInScope(h.Body(), outerTB)
									if r.scopeConsumesTracked(h.Body(), outerTB, true) {
										return true
									}
								}
								if i == 0 && resolvedContainerArg != nil {
									params := runtimeParameters(h)
									if len(params) > 0 && params[0] != nil {
										if p := params[0].AsParameterDeclaration(); p != nil && p.DotDotDotToken == nil && p.Name() != nil {
											var matchedIndices []int
											if resolvedContainerArg.Kind == ast.KindArrayLiteralExpression {
												if arr := resolvedContainerArg.AsArrayLiteralExpression(); arr != nil && arr.Elements != nil {
													for idx, el := range arr.Elements.Nodes {
														if el == nil || el.Kind == ast.KindOmittedExpression {
															continue
														}
														if el.Kind == ast.KindSpreadElement {
															break
														}
														if r.exprCanReturnTrackedPromiseDepth(el, tb, depth+1) {
															matchedIndices = append(matchedIndices, idx)
														}
													}
												}
											} else if r.matchesTrackedRestIdent(resolvedContainerArg, tb, nil) && tb.restOffset >= 0 {
												matchedIndices = []int{tb.restOffset}
											}
											checkInner := func(innerTB *trackedBindings) bool {
												if innerTB.isEmpty() {
													return false
												}
												r.propagateLocalAliasesInScope(h.Body(), innerTB)
												return r.scopeConsumesTracked(h.Body(), innerTB, true)
											}
											if len(matchedIndices) > 0 {
												for _, elemIdx := range matchedIndices {
													innerTB := &trackedBindings{restOffset: -1}
													if r.addTrackedRestParam(innerTB, p.Name(), elemIdx) && checkInner(innerTB) {
														return true
													}
												}
											} else if p.Name().Kind == ast.KindIdentifier {
												innerTB := &trackedBindings{restOffset: -1}
												kind := trackedBindingContainer
												if r.isKnownBuiltinCollectionExpr(resolvedContainerArg, tb) {
													kind = trackedBindingRest
												}
												r.addTrackedIdentKind(innerTB, p.Name(), kind)
												if checkInner(innerTB) {
													return true
												}
											} else if p.Name().Kind == ast.KindArrayBindingPattern {
												if bp := p.Name().AsBindingPattern(); bp != nil && bp.Elements != nil && len(bp.Elements.Nodes) == 1 {
													el := bp.Elements.Nodes[0]
													if el != nil && el.Kind == ast.KindBindingElement {
														be := el.AsBindingElement()
														if be.DotDotDotToken != nil && be.Name() != nil && be.Name().Kind == ast.KindIdentifier {
															innerTB := &trackedBindings{restOffset: -1}
															r.addTrackedIdentKind(innerTB, be.Name(), trackedBindingRest)
															if checkInner(innerTB) {
																return true
															}
														}
													}
												}
											}
										}
									}
								}
								continue
							}
							if r.exprReferencesTracked(h, tb, true) {
								return true
							}
						}
					}
					return false
				case "from", "of":
					recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
					if recv != nil && recv.Kind == ast.KindIdentifier && recv.AsIdentifier().Text == "Array" &&
						!r.isLocallyValueShadowed(recv, "Array") {
						return false
					}
				case "pop", "shift":
					recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
					if r.isKnownBuiltinCollectionExpr(recv, tb) {
						if recv != nil && recv.Kind == ast.KindArrayLiteralExpression {
							if arr := recv.AsArrayLiteralExpression(); arr != nil && arr.Elements != nil {
								elems := arr.Elements.Nodes
								hasSpread := false
								for _, el := range elems {
									if el != nil && el.Kind == ast.KindSpreadElement {
										hasSpread = true
										break
									}
								}
								if !hasSpread {
									if len(elems) == 0 {
										return false
									}
									target := elems[0]
									if mName == "pop" {
										target = elems[len(elems)-1]
									}
									return target != nil && target.Kind != ast.KindOmittedExpression &&
										r.exprCanReturnTrackedPromiseDepth(target, tb, depth+1)
								}
							}
						}
						if recv != nil && recv.Kind == ast.KindIdentifier {
							if r.isContainerInvalidatedAt(recv, tb) || !r.isContainerPopulatedAt(recv, tb) {
								return false
							}
							if slots, arrLen, ok := r.resolveValidArraySlotsAt(recv, tb); ok {
								if mName == "shift" {
									return slots[0]
								}
								if arrLen > 0 && !r.hasContainerPushSinceDecl(recv, tb) {
									return slots[arrLen-1]
								}
								return r.lastPushBeforeUseIsTracked(recv, tb)
							}
							if mName == "shift" && r.matchesTrackedRestIdent(recv, tb, nil) && tb.restOffset >= 0 {
								return tb.restOffset == 0
							}
							if mName == "pop" && r.hasAnyPushBeforeUse(recv, tb) {
								return r.lastPushBeforeUseIsTracked(recv, tb)
							}
						}
					}
				case "push", "unshift", "forEach", "includes", "indexOf", "lastIndexOf",
					"join", "toString", "slice", "concat", "filter", "map", "flat", "flatMap",
					"splice", "reverse", "sort", "fill", "copyWithin", "every", "some",
					"keys", "values", "entries", "has", "add", "set", "delete", "clear":
					if r.isKnownBuiltinCollectionExpr(internalUtils.AccessExpressionObject(callee), tb) {
						return false
					}
				}
			}
		}
		if r != nil && r.typeChecker != nil && r.isKnownNonPromiseReturnType(r.typeChecker.GetTypeAtLocation(expr), expr) {
			return false
		}
		return r.exprReferencesTracked(expr, tb, true)
	case ast.KindNewExpression:
		if r.isBuiltinNonPromiseContainerCtor(expr.AsNewExpression().Expression) {
			return false
		}
		return r.exprReferencesTracked(expr, tb, true)
	case ast.KindTaggedTemplateExpression:
		return r.exprReferencesTracked(expr, tb, true)
	case ast.KindObjectLiteralExpression:
		// Plain objects (including `{ ...promise }`, since `Promise.prototype.then`
		// is non-enumerable) are not Promises unless they define an own `then`
		// property/method referencing `tb`.
		obj := expr.AsObjectLiteralExpression()
		if obj == nil || obj.Properties == nil {
			return false
		}
		for _, prop := range obj.Properties.Nodes {
			if prop == nil {
				continue
			}
			if pName := prop.Name(); pName != nil {
				if name, ok := internalUtils.GetStaticPropertyName(pName); ok && name == "then" {
					if r.exprReferencesTracked(prop, tb, true) {
						return true
					}
				}
			}
		}
		return false
	default:
		return false
	}
}

func (r *AsyncAssertionResolver) isDefinitelyUndefinedExpr(expr *ast.Node) bool {
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return true
	}
	switch expr.Kind {
	case ast.KindUndefinedKeyword, ast.KindVoidExpression, ast.KindOmittedExpression:
		return true
	case ast.KindIdentifier:
		return expr.AsIdentifier().Text == "undefined" && !r.isLocallyValueShadowed(expr, "undefined")
	}
	return false
}

func (r *AsyncAssertionResolver) propagateParamDefaults(params []*ast.Node, startIdx int, tb *trackedBindings, passedArgs map[int]*ast.Node, firstUnknownSpreadIdx int) {
	var visitPattern func(pattern *ast.Node, initExpr *ast.Node)
	visitPattern = func(pattern *ast.Node, initExpr *ast.Node) {
		if pattern == nil || initExpr == nil {
			return
		}
		initExpr = internalUtils.SkipAssertionsAndParens(initExpr)
		if initExpr == nil {
			return
		}
		switch pattern.Kind {
		case ast.KindObjectBindingPattern:
			if initExpr.Kind != ast.KindObjectLiteralExpression {
				return
			}
			obj := initExpr.AsObjectLiteralExpression()
			presentProps := map[string]bool{}
			if obj != nil && obj.Properties != nil {
				for _, prop := range obj.Properties.Nodes {
					if prop == nil {
						continue
					}
					if prop.Kind == ast.KindSpreadAssignment || prop.Name() == nil {
						return
					}
					pName, ok := internalUtils.GetStaticPropertyName(prop.Name())
					if !ok && r != nil && r.staticStringEvaluator() != nil {
						pName, ok = r.staticStringEvaluator().EvalPropertyName(prop.Name())
					}
					if !ok {
						return
					}
					presentProps[pName] = true
				}
			}
			bp := pattern.AsBindingPattern()
			if bp == nil || bp.Elements == nil {
				return
			}
			for _, el := range bp.Elements.Nodes {
				if el == nil || el.Kind != ast.KindBindingElement {
					continue
				}
				be := el.AsBindingElement()
				if be.DotDotDotToken != nil {
					continue
				}
				var keyName string
				var keyOK bool
				if be.PropertyName != nil {
					keyName, keyOK = internalUtils.GetStaticPropertyName(be.PropertyName)
					if !keyOK && r != nil && r.staticStringEvaluator() != nil {
						keyName, keyOK = r.staticStringEvaluator().EvalPropertyName(be.PropertyName)
					}
				} else if be.Name() != nil && be.Name().Kind == ast.KindIdentifier {
					keyName, keyOK = be.Name().AsIdentifier().Text, true
				}
				if !keyOK || presentProps[keyName] {
					continue
				}
				if be.Initializer != nil && r.exprCanReturnTrackedPromise(be.Initializer, tb) {
					if be.Name() != nil && be.Name().Kind == ast.KindIdentifier {
						r.addTrackedIdent(tb, be.Name(), false)
					}
				}
				visitPattern(be.Name(), be.Initializer)
			}
		case ast.KindArrayBindingPattern:
			if initExpr.Kind != ast.KindArrayLiteralExpression {
				return
			}
			arr := initExpr.AsArrayLiteralExpression()
			var elems []*ast.Node
			if arr != nil && arr.Elements != nil {
				elems = arr.Elements.Nodes
			}
			bp := pattern.AsBindingPattern()
			if bp == nil || bp.Elements == nil {
				return
			}
			for i, el := range bp.Elements.Nodes {
				if i < len(elems) && elems[i] != nil && elems[i].Kind == ast.KindSpreadElement {
					return
				}
				if el == nil || el.Kind != ast.KindBindingElement {
					continue
				}
				be := el.AsBindingElement()
				if be.DotDotDotToken != nil {
					return
				}
				if i < len(elems) && !r.isDefinitelyUndefinedExpr(elems[i]) {
					continue
				}
				if be.Initializer != nil && r.exprCanReturnTrackedPromise(be.Initializer, tb) {
					if be.Name() != nil && be.Name().Kind == ast.KindIdentifier {
						r.addTrackedIdent(tb, be.Name(), false)
					}
				}
				visitPattern(be.Name(), be.Initializer)
			}
		}
	}
	for i := startIdx; i < len(params); i++ {
		if params[i] == nil {
			continue
		}
		p := params[i].AsParameterDeclaration()
		if p == nil {
			continue
		}
		paramMayBeNonUndefined := false
		if firstUnknownSpreadIdx > 0 && i+1 >= firstUnknownSpreadIdx {
			paramMayBeNonUndefined = true
		} else if passedArgs != nil {
			if argExpr, ok := passedArgs[i+1]; ok && !r.isDefinitelyUndefinedExpr(argExpr) {
				paramMayBeNonUndefined = true
			}
		}
		if !paramMayBeNonUndefined {
			if p.Initializer != nil && r.exprCanReturnTrackedPromise(p.Initializer, tb) {
				if p.Name() != nil && p.Name().Kind == ast.KindIdentifier {
					r.addTrackedIdent(tb, p.Name(), false)
				}
			}
			visitPattern(p.Name(), p.Initializer)
		}
	}
}

func (r *AsyncAssertionResolver) indexExprValueRefs(
	root *ast.Node,
	enterNestedFunctions bool,
	bySym map[*ast.Symbol][]int,
	byName map[string][]int,
	idx int,
) {
	if root == nil {
		return
	}
	seenSym := map[*ast.Symbol]bool{}
	seenName := map[string]bool{}
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || ast.IsTypeNode(node) || ast.IsPartOfTypeQuery(node) {
			return
		}
		if node != root {
			switch node.Kind {
			case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
				ast.KindClassStaticBlockDeclaration:
				if !enterNestedFunctions {
					return
				}
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
				if !enterNestedFunctions {
					return
				}
			}
		}
		if node.Kind == ast.KindIdentifier && isIdentifierValueReference(node) {
			name := node.AsIdentifier().Text
			if r != nil && r.refs != nil {
				if sym := r.resolveSymbol(node); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
					if !seenSym[sym] {
						seenSym[sym] = true
						bySym[sym] = append(bySym[sym], idx)
					}
					return
				}
			}
			if !seenName[name] {
				seenName[name] = true
				byName[name] = append(byName[name], idx)
			}
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(root)
}

func (r *AsyncAssertionResolver) getScopeBodyIndex(root *ast.Node) *scopeBodyIndex {
	if root == nil {
		return nil
	}
	if r != nil && r.scopeBodyIndices != nil {
		if idx, ok := r.scopeBodyIndices[root]; ok {
			return idx
		}
	}
	idx := &scopeBodyIndex{
		depsBySym:      map[*ast.Symbol][]int{},
		depsByName:     map[string][]int{},
		consumesBySym:  map[*ast.Symbol][]int{},
		consumesByName: map[string][]int{},
	}
	addDep := func(edge scopeDepEdge, exprToIndex *ast.Node) {
		edgeIdx := len(idx.deps)
		idx.deps = append(idx.deps, edge)
		r.indexExprValueRefs(exprToIndex, true, idx.depsBySym, idx.depsByName, edgeIdx)
	}
	addConsume := func(site scopeConsumeSite, enterNested bool) {
		siteIdx := len(idx.consumes)
		idx.consumes = append(idx.consumes, site)
		r.indexExprValueRefs(site.expr, enterNested, idx.consumesBySym, idx.consumesByName, siteIdx)
	}
	markContainerRemoved := func(ident *ast.Node, startPos, pos int, outerLoop *ast.Node) {
		if ident == nil || ident.Kind != ast.KindIdentifier {
			return
		}
		rem := containerRemoval{startPos: startPos, pos: pos, outerLoop: outerLoop}
		name := ident.AsIdentifier().Text
		if r != nil && r.refs != nil {
			if sym := r.resolveSymbol(ident); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
				if idx.removedContainers == nil {
					idx.removedContainers = map[*ast.Symbol][]containerRemoval{}
				}
				idx.removedContainers[sym] = append(idx.removedContainers[sym], rem)
				return
			}
		}
		if idx.removedContainerNames == nil {
			idx.removedContainerNames = map[string][]containerRemoval{}
		}
		idx.removedContainerNames[name] = append(idx.removedContainerNames[name], rem)
	}
	markContainerShifted := func(ident *ast.Node, startPos, pos int, outerLoop *ast.Node) {
		if ident == nil || ident.Kind != ast.KindIdentifier {
			return
		}
		sh := containerRemoval{startPos: startPos, pos: pos, outerLoop: outerLoop}
		name := ident.AsIdentifier().Text
		if r != nil && r.refs != nil {
			if sym := r.resolveSymbol(ident); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
				if idx.shiftedContainers == nil {
					idx.shiftedContainers = map[*ast.Symbol][]containerRemoval{}
				}
				idx.shiftedContainers[sym] = append(idx.shiftedContainers[sym], sh)
				return
			}
		}
		if idx.shiftedContainerNames == nil {
			idx.shiftedContainerNames = map[string][]containerRemoval{}
		}
		idx.shiftedContainerNames[name] = append(idx.shiftedContainerNames[name], sh)
	}
	markMapKeySet := func(ident *ast.Node, key string, startPos, pos int, outerLoop *ast.Node) {
		if ident == nil || ident.Kind != ast.KindIdentifier || key == "" {
			return
		}
		ev := mapKeySetEvent{key: key, startPos: startPos, pos: pos, outerLoop: outerLoop}
		name := ident.AsIdentifier().Text
		if r != nil && r.refs != nil {
			if sym := r.resolveSymbol(ident); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
				if idx.mapKeySets == nil {
					idx.mapKeySets = map[*ast.Symbol][]mapKeySetEvent{}
				}
				idx.mapKeySets[sym] = append(idx.mapKeySets[sym], ev)
				return
			}
		}
		if idx.mapNameKeySets == nil {
			idx.mapNameKeySets = map[string][]mapKeySetEvent{}
		}
		idx.mapNameKeySets[name] = append(idx.mapNameKeySets[name], ev)
	}
	markContainerPushed := func(ident *ast.Node, startPos, pos int, outerLoop *ast.Node) {
		if ident == nil || ident.Kind != ast.KindIdentifier {
			return
		}
		p := containerRemoval{startPos: startPos, pos: pos, outerLoop: outerLoop}
		name := ident.AsIdentifier().Text
		if r != nil && r.refs != nil {
			if sym := r.resolveSymbol(ident); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
				if idx.pushedContainers == nil {
					idx.pushedContainers = map[*ast.Symbol][]containerRemoval{}
				}
				idx.pushedContainers[sym] = append(idx.pushedContainers[sym], p)
				return
			}
		}
		if idx.pushedContainerNames == nil {
			idx.pushedContainerNames = map[string][]containerRemoval{}
		}
		idx.pushedContainerNames[name] = append(idx.pushedContainerNames[name], p)
	}
	markContainerReassigned := func(ident *ast.Node, pos int, topLevelEquals bool) {
		if ident == nil || ident.Kind != ast.KindIdentifier {
			return
		}
		name := ident.AsIdentifier().Text
		if r != nil && r.refs != nil {
			if sym := r.resolveSymbol(ident); sym != nil && (name != "arguments" || len(sym.Declarations) > 0) {
				if idx.reassignedContainers == nil {
					idx.reassignedContainers = map[*ast.Symbol]bool{}
				}
				idx.reassignedContainers[sym] = true
				if topLevelEquals {
					if idx.topLevelReassigns == nil {
						idx.topLevelReassigns = map[*ast.Symbol][]int{}
					}
					idx.topLevelReassigns[sym] = append(idx.topLevelReassigns[sym], pos)
				}
				return
			}
		}
		if idx.reassignedNames == nil {
			idx.reassignedNames = map[string]bool{}
		}
		idx.reassignedNames[name] = true
		if topLevelEquals {
			if idx.topLevelReassignNames == nil {
				idx.topLevelReassignNames = map[string][]int{}
			}
			idx.topLevelReassignNames[name] = append(idx.topLevelReassignNames[name], pos)
		}
	}
	markLengthOperandRemoved := func(operand *ast.Node, startPos, pos int, outerLoop *ast.Node) {
		lhs := internalUtils.SkipAssertionsAndParens(operand)
		if lhs != nil && ast.IsAccessExpression(lhs) {
			if mName, ok := r.accessStaticNameThroughAssertions(lhs); ok && mName == "length" {
				if recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(lhs)); recv != nil &&
					recv.Kind == ast.KindIdentifier {
					markContainerRemoved(recv, startPos, pos, outerLoop)
				}
			}
		}
	}
	var outerLoop *ast.Node
	branchDepth := 0
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindClassStaticBlockDeclaration:
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
		}
		isLoop := ast.IsIterationStatement(node, false)
		isBranch := isLoop || node.Kind == ast.KindIfStatement || node.Kind == ast.KindSwitchStatement ||
			node.Kind == ast.KindConditionalExpression || node.Kind == ast.KindCatchClause ||
			(node.Kind == ast.KindTryStatement && node.AsTryStatement().CatchClause != nil) ||
			ast.IsOptionalChain(node)
		if !isBranch && node.Kind == ast.KindBinaryExpression {
			if bin := node.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil {
				switch bin.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
					ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
					isBranch = true
				}
			}
		}
		edgeTopLevel := !isBranch && branchDepth == 0
		switch node.Kind {
		case ast.KindVariableDeclaration:
			vd := node.AsVariableDeclaration()
			if vd != nil && vd.Initializer != nil && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
				addDep(scopeDepEdge{kind: scopeDepVarDecl, target: vd.Name(), rhs: vd.Initializer, pos: node.Pos(), topLevel: edgeTopLevel, outerLoop: outerLoop}, vd.Initializer)
			}
		case ast.KindElementAccessExpression:
			if internalUtils.IsWriteReference(node) {
				if recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(node)); recv != nil &&
					recv.Kind == ast.KindIdentifier {
					markContainerShifted(recv, node.Pos(), node.End(), outerLoop)
				}
			}
		case ast.KindBinaryExpression:
			bin := node.AsBinaryExpression()
			if bin != nil && bin.OperatorToken != nil {
				op := bin.OperatorToken.Kind
				if ast.IsAssignmentOperator(op) {
					lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
					if lhs != nil && ast.IsAccessExpression(lhs) {
						if mName, ok := r.accessStaticNameThroughAssertions(lhs); ok && mName == "length" {
							if recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(lhs)); recv != nil &&
								recv.Kind == ast.KindIdentifier {
								markContainerRemoved(recv, node.Pos(), node.End(), outerLoop)
							}
						}
					}
				}
				switch op {
				case ast.KindEqualsToken, ast.KindQuestionQuestionEqualsToken,
					ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
					lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
					if lhs != nil && lhs.Kind == ast.KindIdentifier {
						markContainerReassigned(lhs, node.Pos(), edgeTopLevel && op == ast.KindEqualsToken)
						addDep(scopeDepEdge{kind: scopeDepAssign, target: lhs, rhs: bin.Right, pos: node.Pos(), topLevel: edgeTopLevel, outerLoop: outerLoop}, bin.Right)
					} else if op == ast.KindEqualsToken &&
						lhs != nil && lhs.Kind == ast.KindElementAccessExpression {
						if recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(lhs)); recv != nil &&
							recv.Kind == ast.KindIdentifier {
							markContainerShifted(recv, node.Pos(), node.End(), outerLoop)
							addDep(scopeDepEdge{kind: scopeDepContainerIndexWrite, target: recv, rhs: bin.Right, pos: node.Pos(), topLevel: edgeTopLevel, outerLoop: outerLoop}, bin.Right)
						}
					}
				}
			}
		case ast.KindPrefixUnaryExpression:
			if op := node.AsPrefixUnaryExpression().Operator; op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken {
				markLengthOperandRemoved(node.AsPrefixUnaryExpression().Operand, node.Pos(), node.End(), outerLoop)
			}
		case ast.KindPostfixUnaryExpression:
			markLengthOperandRemoved(node.AsPostfixUnaryExpression().Operand, node.Pos(), node.End(), outerLoop)
		case ast.KindDeleteExpression:
			if target := internalUtils.SkipAssertionsAndParens(node.AsDeleteExpression().Expression); target != nil &&
				target.Kind == ast.KindElementAccessExpression {
				if recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(target)); recv != nil &&
					recv.Kind == ast.KindIdentifier {
					markContainerRemoved(recv, node.Pos(), node.End(), outerLoop)
				}
			}
		case ast.KindCallExpression:
			callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
			if callee != nil && ast.IsAccessExpression(callee) {
				if mName, ok := r.accessStaticNameThroughAssertions(callee); ok {
					recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
					if mName == "set" {
						if rootRecv := r.rootChainedMapSetIdent(recv); rootRecv != nil {
							if arg0, arg0OK := staticCallArgAt(node, 0); arg0OK && arg0 != nil {
								if sk, hasKey := r.evalMapStaticKey(arg0); hasKey {
									markMapKeySet(rootRecv, sk, recv.End(), node.End(), outerLoop)
								}
							}
							if args := node.AsCallExpression().Arguments; args != nil && len(args.Nodes) > 0 {
								edgeIdx := len(idx.deps)
								idx.deps = append(idx.deps, scopeDepEdge{kind: scopeDepMapSet, target: rootRecv, rhs: node, pos: node.Pos(), topLevel: edgeTopLevel, outerLoop: outerLoop})
								for _, a := range args.Nodes {
									r.indexExprValueRefs(a, true, idx.depsBySym, idx.depsByName, edgeIdx)
								}
							}
						}
					} else if recv != nil && recv.Kind == ast.KindIdentifier {
						switch mName {
						case "push", "unshift", "add":
							switch mName {
							case "push":
								markContainerPushed(recv, node.Pos(), node.End(), outerLoop)
							case "unshift":
								markContainerShifted(recv, node.Pos(), node.End(), outerLoop)
							}
							if args := node.AsCallExpression().Arguments; args != nil && len(args.Nodes) > 0 {
								edgeIdx := len(idx.deps)
								idx.deps = append(idx.deps, scopeDepEdge{kind: scopeDepContainerPush, target: recv, rhs: node, pos: node.Pos(), topLevel: edgeTopLevel, outerLoop: outerLoop})
								for _, a := range args.Nodes {
									r.indexExprValueRefs(a, true, idx.depsBySym, idx.depsByName, edgeIdx)
								}
							}
						case "reverse", "sort":
							markContainerShifted(recv, node.Pos(), node.End(), outerLoop)
						case "pop", "shift", "splice", "clear", "delete", "fill", "copyWithin":
							markContainerRemoved(recv, node.Pos(), node.End(), outerLoop)
						}
					}
				}
			}
		case ast.KindAwaitExpression:
			addConsume(scopeConsumeSite{kind: scopeConsumeAwait, node: node, expr: node.AsAwaitExpression().Expression}, true)
		case ast.KindForOfStatement:
			if forOf := node.AsForInOrOfStatement(); forOf != nil && forOf.AwaitModifier != nil {
				addConsume(scopeConsumeSite{kind: scopeConsumeForAwait, node: node, expr: forOf.Expression}, true)
			}
		case ast.KindReturnStatement:
			if ret := node.AsReturnStatement(); ret != nil && ret.Expression != nil {
				addConsume(scopeConsumeSite{kind: scopeConsumeReturn, node: node, expr: ret.Expression}, true)
			}
		}
		prevOuterLoop := outerLoop
		if isLoop && outerLoop == nil {
			outerLoop = node
		}
		if isBranch {
			branchDepth++
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
		if isBranch {
			branchDepth--
		}
		outerLoop = prevOuterLoop
	}
	visit(root)
	if r != nil {
		if r.scopeBodyIndices == nil {
			r.scopeBodyIndices = map[*ast.Node]*scopeBodyIndex{}
		}
		r.scopeBodyIndices[root] = idx
	}
	return idx
}

func isPureAliasExpr(expr *ast.Node, depth int) bool {
	if depth > 6 || expr == nil {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	for expr != nil && internalUtils.IsCommaOperator(expr) {
		expr = internalUtils.SkipAssertionsAndParens(expr.AsBinaryExpression().Right)
	}
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindIdentifier:
		return true
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		return cond != nil && isPureAliasExpr(cond.WhenTrue, depth+1) && isPureAliasExpr(cond.WhenFalse, depth+1)
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindAmpersandAmpersandToken, ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
			return isPureAliasExpr(bin.Right, depth+1)
		case ast.KindBarBarToken, ast.KindQuestionQuestionToken,
			ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return isPureAliasExpr(bin.Left, depth+1) && isPureAliasExpr(bin.Right, depth+1)
		}
	}
	return false
}

func (r *AsyncAssertionResolver) propagateLocalAliasesInScope(root *ast.Node, tb *trackedBindings) {
	if root == nil || root.Kind != ast.KindBlock || tb == nil || tb.isEmpty() {
		return
	}
	idx := r.getScopeBodyIndex(root)
	if idx == nil {
		return
	}
	isInitialScope := len(tb.containerScopes) == 0
	hasScope := false
	for _, s := range tb.containerScopes {
		if s == root {
			hasScope = true
			break
		}
	}
	if !hasScope {
		tb.containerScopes = append(tb.containerScopes, root)
	}
	if isInitialScope {
		markInitialSym := func(sym *ast.Symbol) {
			if tb.initialContainers == nil {
				tb.initialContainers = map[*ast.Symbol]bool{}
			}
			tb.initialContainers[sym] = true
		}
		for sym := range tb.restSymbols {
			markInitialSym(sym)
		}
		for sym := range tb.containerSymbols {
			markInitialSym(sym)
		}
		for sym := range tb.mapSymbols {
			markInitialSym(sym)
		}
		markInitialName := func(name string) {
			if tb.initialContainerNames == nil {
				tb.initialContainerNames = map[string]bool{}
			}
			tb.initialContainerNames[name] = true
		}
		for name := range tb.restNames {
			markInitialName(name)
		}
		for name := range tb.containerNames {
			markInitialName(name)
		}
		for name := range tb.mapNames {
			markInitialName(name)
		}
	}
	if len(idx.deps) == 0 || tb.isEmpty() {
		return
	}
	type workItem struct {
		sym  *ast.Symbol
		name string
		kind trackedBindingKind
	}
	var queue []workItem
	queued := map[workItem]bool{}
	enqueue := func(sym *ast.Symbol, name string, kind trackedBindingKind) {
		if sym != nil {
			w := workItem{sym: sym, kind: kind}
			if !queued[w] {
				queued[w] = true
				queue = append(queue, w)
			}
		} else if name != "" {
			w := workItem{name: name, kind: kind}
			if !queued[w] {
				queued[w] = true
				queue = append(queue, w)
			}
		}
	}
	for sym := range tb.symbols {
		enqueue(sym, "", trackedBindingDirect)
	}
	for sym := range tb.restSymbols {
		enqueue(sym, "", trackedBindingRest)
	}
	for sym := range tb.containerSymbols {
		enqueue(sym, "", trackedBindingContainer)
	}
	for sym := range tb.mapSymbols {
		enqueue(sym, "", trackedBindingMap)
	}
	for name := range tb.names {
		enqueue(nil, name, trackedBindingDirect)
	}
	for name := range tb.restNames {
		enqueue(nil, name, trackedBindingRest)
	}
	for name := range tb.containerNames {
		enqueue(nil, name, trackedBindingContainer)
	}
	for name := range tb.mapNames {
		enqueue(nil, name, trackedBindingMap)
	}
	if tb.trackArguments {
		enqueue(nil, "arguments", trackedBindingRest)
	}

	recordContainerWrite := func(sym *ast.Symbol, name string, pos, endPos int, topLevel bool, outerLoop *ast.Node, isMutation, isLastPushElem bool, mapKeys map[string]bool) {
		w := containerWrite{scope: root, pos: pos, endPos: endPos, topLevel: topLevel, outerLoop: outerLoop, isMutation: isMutation, isLastPushElem: isLastPushElem, mapStaticKeys: mapKeys}
		if sym != nil {
			if tb.containerWrites == nil {
				tb.containerWrites = map[*ast.Symbol][]containerWrite{}
			}
			tb.containerWrites[sym] = append(tb.containerWrites[sym], w)
			return
		}
		if name != "" {
			if tb.containerNameWrites == nil {
				tb.containerNameWrites = map[string][]containerWrite{}
			}
			tb.containerNameWrites[name] = append(tb.containerNameWrites[name], w)
		}
	}
	linkPeerSymbols := func(dst, src *ast.Symbol, pos int, topLevel bool) {
		if dst == nil || src == nil || dst == src {
			return
		}
		if tb.containerPeers == nil {
			tb.containerPeers = map[*ast.Symbol][]containerPeerLink{}
		}
		addEdge := func(from *ast.Symbol, link containerPeerLink) {
			for _, existing := range tb.containerPeers[from] {
				if existing == link {
					return
				}
			}
			tb.containerPeers[from] = append(tb.containerPeers[from], link)
		}
		addEdge(dst, containerPeerLink{sym: src, scope: root, pos: pos, topLevel: topLevel, isDest: false})
		addEdge(src, containerPeerLink{sym: dst, scope: root, pos: pos, topLevel: topLevel, isDest: true})
	}
	linkPeerNames := func(dst, src string, pos int, topLevel bool) {
		if dst == "" || src == "" || dst == src {
			return
		}
		if tb.containerNamePeers == nil {
			tb.containerNamePeers = map[string][]containerNamePeerLink{}
		}
		addEdge := func(from string, link containerNamePeerLink) {
			for _, existing := range tb.containerNamePeers[from] {
				if existing == link {
					return
				}
			}
			tb.containerNamePeers[from] = append(tb.containerNamePeers[from], link)
		}
		addEdge(dst, containerNamePeerLink{name: src, scope: root, pos: pos, topLevel: topLevel, isDest: false})
		addEdge(src, containerNamePeerLink{name: dst, scope: root, pos: pos, topLevel: topLevel, isDest: true})
	}
	var collectAliasIdentifiers func(expr *ast.Node, depth int, out *[]*ast.Node)
	collectAliasIdentifiers = func(expr *ast.Node, depth int, out *[]*ast.Node) {
		if depth > 6 || expr == nil {
			return
		}
		expr = internalUtils.SkipAssertionsAndParens(expr)
		for expr != nil && internalUtils.IsCommaOperator(expr) {
			expr = internalUtils.SkipAssertionsAndParens(expr.AsBinaryExpression().Right)
		}
		if expr == nil {
			return
		}
		switch expr.Kind {
		case ast.KindIdentifier:
			*out = append(*out, expr)
		case ast.KindConditionalExpression:
			if cond := expr.AsConditionalExpression(); cond != nil {
				collectAliasIdentifiers(cond.WhenTrue, depth+1, out)
				collectAliasIdentifiers(cond.WhenFalse, depth+1, out)
			}
		case ast.KindBinaryExpression:
			if bin := expr.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil {
				switch bin.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken, ast.KindEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
					collectAliasIdentifiers(bin.Right, depth+1, out)
				case ast.KindBarBarToken, ast.KindQuestionQuestionToken,
					ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
					collectAliasIdentifiers(bin.Left, depth+1, out)
					collectAliasIdentifiers(bin.Right, depth+1, out)
				}
			}
		}
	}
	linkContainerPeersFromExpr := func(dstSym *ast.Symbol, dstName string, rhs *ast.Node, pos int, topLevel bool) {
		var srcIdentifiers []*ast.Node
		collectAliasIdentifiers(rhs, 0, &srcIdentifiers)
		for _, srcIdent := range srcIdentifiers {
			if srcIdent == nil || srcIdent.Kind != ast.KindIdentifier {
				continue
			}
			srcName := srcIdent.AsIdentifier().Text
			isImplicitArgs := false
			if srcName == "arguments" && tb.trackArguments {
				if r == nil || r.refs == nil {
					isImplicitArgs = true
				} else if s := r.resolveSymbol(srcIdent); s == nil || len(s.Declarations) == 0 {
					isImplicitArgs = true
				}
			}
			if isImplicitArgs {
				if dstSym != nil {
					if tb.argumentsAliasSymbols == nil {
						tb.argumentsAliasSymbols = map[*ast.Symbol]bool{}
					}
					tb.argumentsAliasSymbols[dstSym] = true
				}
				if dstName != "" {
					if tb.argumentsAliasNames == nil {
						tb.argumentsAliasNames = map[string]bool{}
					}
					tb.argumentsAliasNames[dstName] = true
				}
				continue
			}
			if dstSym != nil && r != nil && r.refs != nil {
				if srcSym := r.resolveSymbol(srcIdent); srcSym != nil && (srcName != "arguments" || len(srcSym.Declarations) > 0) {
					linkPeerSymbols(dstSym, srcSym, pos, topLevel)
					continue
				}
			}
			linkPeerNames(dstName, srcName, pos, topLevel)
		}
	}
	isContainerReassigned := func(sym *ast.Symbol, name string) bool {
		if sym != nil {
			if idx.reassignedContainers[sym] {
				return true
			}
			if r.hasSymbolWrite(sym) || r.singleSameFileValueDeclaration(sym) == nil {
				return true
			}
			return false
		}
		return idx.reassignedNames[name]
	}
	recordArraySlots := func(sym *ast.Symbol, name string, declPos int, slots map[int]bool, arrLen int) {
		if len(slots) == 0 {
			return
		}
		if sym != nil {
			if tb.arraySlotSymbols == nil {
				tb.arraySlotSymbols = map[*ast.Symbol]map[int]bool{}
			}
			if tb.arraySlotSymbols[sym] == nil {
				tb.arraySlotSymbols[sym] = map[int]bool{}
			}
			for slot := range slots {
				tb.arraySlotSymbols[sym][slot] = true
			}
			if tb.arraySlotDeclPos == nil {
				tb.arraySlotDeclPos = map[*ast.Symbol]int{}
			}
			tb.arraySlotDeclPos[sym] = declPos
			if arrLen >= 0 {
				if tb.arraySlotLen == nil {
					tb.arraySlotLen = map[*ast.Symbol]int{}
				}
				tb.arraySlotLen[sym] = arrLen
			}
			return
		}
		if name != "" {
			if tb.arraySlotNames == nil {
				tb.arraySlotNames = map[string]map[int]bool{}
			}
			if tb.arraySlotNames[name] == nil {
				tb.arraySlotNames[name] = map[int]bool{}
			}
			for slot := range slots {
				tb.arraySlotNames[name][slot] = true
			}
			if tb.arraySlotDeclNamePos == nil {
				tb.arraySlotDeclNamePos = map[string]int{}
			}
			tb.arraySlotDeclNamePos[name] = declPos
			if arrLen >= 0 {
				if tb.arraySlotNameLen == nil {
					tb.arraySlotNameLen = map[string]int{}
				}
				tb.arraySlotNameLen[name] = arrLen
			}
		}
	}

	const maxWorkSteps = 256
	steps := 0
	for len(queue) > 0 && steps < maxWorkSteps {
		item := queue[0]
		queue = queue[1:]
		steps++
		var edgeIndices []int
		if item.sym != nil {
			edgeIndices = idx.depsBySym[item.sym]
		} else {
			edgeIndices = idx.depsByName[item.name]
		}
		for _, edgeIdx := range edgeIndices {
			edge := idx.deps[edgeIdx]
			switch edge.kind {
			case scopeDepVarDecl, scopeDepAssign:
				rhsUnwrapped := internalUtils.SkipAssertionsAndParens(edge.rhs)
				for rhsUnwrapped != nil && internalUtils.IsCommaOperator(rhsUnwrapped) {
					rhsUnwrapped = internalUtils.SkipAssertionsAndParens(rhsUnwrapped.AsBinaryExpression().Right)
				}
				if r.matchesTrackedRestIdent(rhsUnwrapped, tb, nil) {
					sym, name, added := r.addTrackedIdentKind(tb, edge.target, trackedBindingRest)
					linkContainerPeersFromExpr(sym, name, rhsUnwrapped, edge.pos, edge.topLevel)
					if added {
						enqueue(sym, name, trackedBindingRest)
					}
					continue
				}
				if r.exprCanReturnTrackedPromise(edge.rhs, tb) {
					if sym, name, added := r.addTrackedIdentKind(tb, edge.target, trackedBindingDirect); added {
						enqueue(sym, name, trackedBindingDirect)
					}
				}
				pureAlias := isPureAliasExpr(edge.rhs, 0)
				if r.exprCanBeTrackedContainerDepth(edge.rhs, tb, 0, !pureAlias) {
					sym, name, added := r.addTrackedIdentKind(tb, edge.target, trackedBindingContainer)
					if !pureAlias {
						recordContainerWrite(sym, name, edge.pos, edge.rhs.End(), edge.topLevel, edge.outerLoop, false, false, nil)
					}
					linkContainerPeersFromExpr(sym, name, edge.rhs, edge.pos, edge.topLevel)
					if edge.kind == scopeDepVarDecl && !isContainerReassigned(sym, name) {
						if rhsUnwrapped != nil && rhsUnwrapped.Kind == ast.KindArrayLiteralExpression {
							slotIdx := r.getArrayLiteralSlotIndex(rhsUnwrapped)
							if slotIdx != nil {
								var matched map[int]bool
								addSlot := func(pos int) {
									if slotIdx.firstSpreadIdx >= 0 && pos >= slotIdx.firstSpreadIdx {
										return
									}
									if matched == nil {
										matched = map[int]bool{}
									}
									matched[pos] = true
								}
								for s := range tb.symbols {
									for _, pos := range slotIdx.bySym[s] {
										addSlot(pos)
									}
								}
								for n := range tb.names {
									for _, pos := range slotIdx.byName[n] {
										addSlot(pos)
									}
								}
								if len(slotIdx.nonIdentifierSlices) > 0 {
									if arr := rhsUnwrapped.AsArrayLiteralExpression(); arr != nil && arr.Elements != nil {
										for _, pos := range slotIdx.nonIdentifierSlices {
											if slotIdx.firstSpreadIdx >= 0 && pos >= slotIdx.firstSpreadIdx {
												break
											}
											el := arr.Elements.Nodes[pos]
											if el != nil && r.exprCanReturnTrackedPromise(el, tb) {
												addSlot(pos)
											}
										}
									}
								}
								arrLen := -1
								if slotIdx.firstSpreadIdx < 0 {
									arrLen = slotIdx.elemCount
								}
								recordArraySlots(sym, name, edge.pos, matched, arrLen)
							}
						} else if rhsUnwrapped != nil && rhsUnwrapped.Kind == ast.KindIdentifier {
							var srcSlots map[int]bool
							srcDeclPos := edge.pos
							srcLen := -1
							if r != nil && r.refs != nil && len(tb.arraySlotSymbols) > 0 {
								if srcSym := r.resolveSymbol(rhsUnwrapped); srcSym != nil {
									srcSlots = tb.arraySlotSymbols[srcSym]
									if p, ok := tb.arraySlotDeclPos[srcSym]; ok {
										srcDeclPos = p
									}
									if l, ok := tb.arraySlotLen[srcSym]; ok {
										srcLen = l
									}
								} else {
									srcName := rhsUnwrapped.AsIdentifier().Text
									srcSlots = tb.arraySlotNames[srcName]
									if p, ok := tb.arraySlotDeclNamePos[srcName]; ok {
										srcDeclPos = p
									}
									if l, ok := tb.arraySlotNameLen[srcName]; ok {
										srcLen = l
									}
								}
							} else {
								srcName := rhsUnwrapped.AsIdentifier().Text
								srcSlots = tb.arraySlotNames[srcName]
								if p, ok := tb.arraySlotDeclNamePos[srcName]; ok {
									srcDeclPos = p
								}
								if l, ok := tb.arraySlotNameLen[srcName]; ok {
									srcLen = l
								}
							}
							recordArraySlots(sym, name, srcDeclPos, srcSlots, srcLen)
						}
					}
					if added {
						enqueue(sym, name, trackedBindingContainer)
					}
				}
				if mapKeys, mapDyn, mapMatched := r.extractTrackedMapKeys(edge.rhs, tb, 0, !pureAlias); mapMatched {
					sym, name, added := r.addTrackedIdentKind(tb, edge.target, trackedBindingMap)
					if !pureAlias {
						var staticKeys map[string]bool
						if !mapDyn && len(mapKeys) > 0 {
							staticKeys = mapKeys
						}
						recordContainerWrite(sym, name, edge.pos, edge.rhs.End(), edge.topLevel, edge.outerLoop, false, false, staticKeys)
					}
					linkContainerPeersFromExpr(sym, name, edge.rhs, edge.pos, edge.topLevel)
					if added {
						enqueue(sym, name, trackedBindingMap)
					}
				}
			case scopeDepContainerPush:
				if args := edge.rhs.AsCallExpression().Arguments; args != nil {
					matched := false
					for _, a := range args.Nodes {
						if a == nil {
							continue
						}
						if a.Kind == ast.KindSpreadElement {
							if r.exprCanBeTrackedContainer(a.AsSpreadElement().Expression, tb) {
								matched = true
								break
							}
							continue
						}
						if r.exprCanReturnTrackedPromise(a, tb) {
							matched = true
							break
						}
					}
					if matched {
						isLastPushElem := false
						if callCallee := internalUtils.SkipAssertionsAndParens(edge.rhs.AsCallExpression().Expression); callCallee != nil && ast.IsAccessExpression(callCallee) {
							if mName, ok := r.accessStaticNameThroughAssertions(callCallee); ok && mName == "push" && len(args.Nodes) > 0 {
								lastArg := args.Nodes[len(args.Nodes)-1]
								if lastArg != nil && lastArg.Kind != ast.KindSpreadElement && r.exprCanReturnTrackedPromise(lastArg, tb) {
									isLastPushElem = true
								}
							}
						}
						sym, name, added := r.addTrackedIdentKind(tb, edge.target, trackedBindingContainer)
						recordContainerWrite(sym, name, edge.pos, edge.rhs.End(), edge.topLevel, edge.outerLoop, true, isLastPushElem, nil)
						if added {
							enqueue(sym, name, trackedBindingContainer)
						}
					}
				}
			case scopeDepMapSet:
				arg1, argOK := staticCallArgAt(edge.rhs, 1)
				matched := false
				if !argOK {
					matched = r.exprReferencesTracked(edge.rhs, tb, true)
				} else if arg1 != nil {
					matched = r.exprCanReturnTrackedPromise(arg1, tb)
				}
				if matched {
					var staticKeys map[string]bool
					if arg0, arg0OK := staticCallArgAt(edge.rhs, 0); arg0OK && arg0 != nil {
						if sk, hasKey := r.evalMapStaticKey(arg0); hasKey {
							staticKeys = map[string]bool{sk: true}
						}
					}
					sym, name, added := r.addTrackedIdentKind(tb, edge.target, trackedBindingMap)
					recordContainerWrite(sym, name, edge.pos, edge.rhs.End(), edge.topLevel, edge.outerLoop, true, false, staticKeys)
					if added {
						enqueue(sym, name, trackedBindingMap)
					}
				}
			case scopeDepContainerIndexWrite:
				if r.exprCanReturnTrackedPromise(edge.rhs, tb) {
					sym, name, added := r.addTrackedIdentKind(tb, edge.target, trackedBindingContainer)
					recordContainerWrite(sym, name, edge.pos, edge.rhs.End(), edge.topLevel, edge.outerLoop, true, false, nil)
					if added {
						enqueue(sym, name, trackedBindingContainer)
					}
				}
			}
		}
	}
}

func (r *AsyncAssertionResolver) forEachTrackedConsumeSite(
	idx *scopeBodyIndex,
	tb *trackedBindings,
	fn func(scopeConsumeSite) bool,
) {
	if idx == nil || len(idx.consumes) == 0 || tb == nil || tb.isEmpty() {
		return
	}
	checked := map[int]bool{}
	visitIndices := func(indices []int) bool {
		for _, siteIdx := range indices {
			if checked[siteIdx] {
				continue
			}
			checked[siteIdx] = true
			if fn(idx.consumes[siteIdx]) {
				return true
			}
		}
		return false
	}
	for sym := range tb.symbols {
		if visitIndices(idx.consumesBySym[sym]) {
			return
		}
	}
	for sym := range tb.restSymbols {
		if visitIndices(idx.consumesBySym[sym]) {
			return
		}
	}
	for sym := range tb.containerSymbols {
		if visitIndices(idx.consumesBySym[sym]) {
			return
		}
	}
	for sym := range tb.mapSymbols {
		if visitIndices(idx.consumesBySym[sym]) {
			return
		}
	}
	for name := range tb.names {
		if visitIndices(idx.consumesByName[name]) {
			return
		}
	}
	for name := range tb.restNames {
		if visitIndices(idx.consumesByName[name]) {
			return
		}
	}
	for name := range tb.containerNames {
		if visitIndices(idx.consumesByName[name]) {
			return
		}
	}
	for name := range tb.mapNames {
		if visitIndices(idx.consumesByName[name]) {
			return
		}
	}
	if tb.trackArguments {
		visitIndices(idx.consumesByName["arguments"])
	}
}

// scopeConsumesTracked reports whether `root` (a callback body or catch block)
// awaits or returns a Promise derived from `tb` in its own synchronous scope.
func (r *AsyncAssertionResolver) scopeConsumesTracked(root *ast.Node, tb *trackedBindings, allowReturn bool) bool {
	if root == nil || tb == nil || tb.isEmpty() {
		return false
	}
	if root.Kind != ast.KindBlock {
		if allowReturn && r.exprCanReturnTrackedPromise(root, tb) {
			return true
		}
	}
	consumed := false
	r.forEachTrackedConsumeSite(r.getScopeBodyIndex(root), tb, func(site scopeConsumeSite) bool {
		switch site.kind {
		case scopeConsumeAwait:
			if r.exprCanReturnTrackedPromise(site.expr, tb) {
				consumed = true
				return true
			}
		case scopeConsumeForAwait:
			if r.exprCanBeTrackedContainer(site.expr, tb) {
				consumed = true
				return true
			}
		case scopeConsumeReturn:
			if allowReturn && r.exprCanReturnTrackedPromise(site.expr, tb) {
				consumed = true
				return true
			}
		}
		return false
	})
	return consumed
}

// addTrackedRestParam adds bindings from a rest parameter (`...rest` or
// `...[first, second, ...tail]`) corresponding to 0-based `restOffset` into the
// rest slice (`restOffset < 0` means unknown offset due to an earlier spread).
func (r *AsyncAssertionResolver) addTrackedRestParam(tb *trackedBindings, pattern *ast.Node, restOffset int) bool {
	if pattern == nil {
		return false
	}
	if pattern.Kind == ast.KindIdentifier {
		tb.restOffset = restOffset
		r.addTrackedIdent(tb, pattern, true)
		return true
	}
	if pattern.Kind != ast.KindArrayBindingPattern {
		return false
	}
	bp := pattern.AsBindingPattern()
	if bp == nil || bp.Elements == nil {
		return false
	}
	if restOffset < 0 {
		// Unknown offset: handled by caller checking every positional element up to a nested rest.
		return false
	}
	for i, el := range bp.Elements.Nodes {
		if el == nil || el.Kind != ast.KindBindingElement {
			continue
		}
		be := el.AsBindingElement()
		if be.DotDotDotToken != nil {
			if i <= restOffset && be.Name() != nil && be.Name().Kind == ast.KindIdentifier {
				tb.restOffset = restOffset - i
				r.addTrackedIdent(tb, be.Name(), true)
				return true
			}
			return false
		}
		if i == restOffset {
			if be.Name() != nil && be.Name().Kind == ast.KindIdentifier {
				r.addTrackedIdent(tb, be.Name(), false)
				return true
			}
			return false
		}
	}
	return false
}

func (r *AsyncAssertionResolver) buildCallbackTrackedBindings(
	cb *ast.Node,
	targetParam *ast.Node,
	paramIdx int,
	restOffset int,
	includeArguments bool,
	passedArgs map[int]*ast.Node,
	firstUnknownSpreadIdx int,
) *trackedBindings {
	tb := &trackedBindings{
		trackArguments: includeArguments,
		restOffset:     -1,
	}
	if includeArguments {
		tb.restOffset = restOffset
	}
	if targetParam != nil {
		p := targetParam.AsParameterDeclaration()
		if p != nil && p.DotDotDotToken != nil {
			r.addTrackedRestParam(tb, p.Name(), restOffset)
		} else if targetParam.Name() != nil && targetParam.Name().Kind == ast.KindIdentifier {
			r.addTrackedIdent(tb, targetParam.Name(), false)
		}
		if !tb.isEmpty() {
			summary := r.getInlineCallbackSummary(cb)
			if summary.hasDefaultParam {
				r.propagateParamDefaults(summary.params, paramIdx+1, tb, passedArgs, firstUnknownSpreadIdx)
			}
		}
	}
	if cb != nil && !tb.isEmpty() {
		r.propagateLocalAliasesInScope(cb.Body(), tb)
	}
	return tb
}

func (r *AsyncAssertionResolver) isParamConsumedByCallback(cb *ast.Node, targetParam *ast.Node, paramIdx int, restOffset int, passedArgs map[int]*ast.Node, firstUnknownSpreadIdx int) bool {
	if cb == nil || targetParam == nil || cb.Body() == nil {
		return false
	}
	tb := r.buildCallbackTrackedBindings(cb, targetParam, paramIdx, restOffset, false, passedArgs, firstUnknownSpreadIdx)
	if tb.isEmpty() {
		return false
	}
	return r.scopeConsumesTracked(cb.Body(), tb, true)
}

func (r *AsyncAssertionResolver) isRestArrayPatternConsumedForSpread(cb *ast.Node, restParam *ast.Node, restParamIdx int, minRestOffset int) bool {
	if cb == nil || restParam == nil || restParam.Name() == nil || restParam.Name().Kind != ast.KindArrayBindingPattern {
		return false
	}
	bp := restParam.Name().AsBindingPattern()
	if bp == nil || bp.Elements == nil {
		return false
	}
	if minRestOffset < 0 {
		minRestOffset = 0
	}
	for i, el := range bp.Elements.Nodes {
		if el == nil || el.Kind != ast.KindBindingElement {
			continue
		}
		be := el.AsBindingElement()
		if be.DotDotDotToken != nil {
			if be.Name() == nil || be.Name().Kind != ast.KindIdentifier {
				return false
			}
			tb := &trackedBindings{restOffset: -1}
			r.addTrackedIdent(tb, be.Name(), true)
			r.propagateLocalAliasesInScope(cb.Body(), tb)
			if !r.scopeConsumesTracked(cb.Body(), tb, true) {
				return false
			}
			for j := minRestOffset; j < i; j++ {
				if j >= len(bp.Elements.Nodes) {
					return false
				}
				posEl := bp.Elements.Nodes[j]
				if posEl == nil || posEl.Kind != ast.KindBindingElement || posEl.Name() == nil || posEl.Name().Kind != ast.KindIdentifier {
					return false
				}
				posTB := &trackedBindings{restOffset: -1}
				r.addTrackedIdent(posTB, posEl.Name(), false)
				r.propagateLocalAliasesInScope(cb.Body(), posTB)
				if !r.scopeConsumesTracked(cb.Body(), posTB, true) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func (r *AsyncAssertionResolver) statementAlwaysThrows(stmt *ast.Node, tb *trackedBindings) bool {
	if stmt == nil {
		return false
	}
	switch stmt.Kind {
	case ast.KindThrowStatement:
		if tb == nil {
			return true
		}
		return r.exprEvaluatesToTrackedPromise(stmt.AsThrowStatement().Expression, tb)
	case ast.KindBlock:
		block := stmt.AsBlock()
		if block == nil || block.Statements == nil {
			return false
		}
		for _, s := range block.Statements.Nodes {
			if s == nil {
				continue
			}
			if r.statementAlwaysThrows(s, tb) {
				return true
			}
			if r.statementMayExitEarly(s) {
				return false
			}
		}
		return false
	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		if ifStmt == nil || ifStmt.ElseStatement == nil {
			return false
		}
		return r.statementAlwaysThrows(ifStmt.ThenStatement, tb) &&
			r.statementAlwaysThrows(ifStmt.ElseStatement, tb)
	}
	return false
}

func (r *AsyncAssertionResolver) exprReturnsRejectedTrackedPromise(expr *ast.Node, tb *trackedBindings, isAsync bool, depth int) int {
	if depth > 6 {
		return 0
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	for expr != nil && internalUtils.IsCommaOperator(expr) {
		expr = internalUtils.SkipAssertionsAndParens(expr.AsBinaryExpression().Right)
	}
	if expr == nil {
		return 0
	}
	switch expr.Kind {
	case ast.KindAwaitExpression:
		if isAsync {
			return r.exprReturnsRejectedTrackedPromise(expr.AsAwaitExpression().Expression, tb, true, depth+1)
		}
	case ast.KindCallExpression:
		if method, ok := r.promiseMethodName(expr); ok && method == "reject" {
			firstArg, argOK := staticCallArgAt(expr, 0)
			if !argOK || firstArg == nil {
				return 0
			}
			firstArgUnwrapped := internalUtils.SkipAssertionsAndParens(firstArg)
			if isAsync && firstArgUnwrapped != nil && firstArgUnwrapped.Kind == ast.KindAwaitExpression {
				// `await` throws before `Promise.reject` is invoked.
				return r.exprReturnsRejectedTrackedPromise(firstArgUnwrapped, tb, true, depth+1)
			}
			if nested := r.exprReturnsRejectedTrackedPromise(firstArg, tb, isAsync, depth+1); nested > 0 {
				return 1 + nested
			}
			if r.exprEvaluatesToTrackedPromise(firstArg, tb) {
				return 1
			}
		}
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		if cond != nil {
			a := r.exprReturnsRejectedTrackedPromise(cond.WhenTrue, tb, isAsync, depth+1)
			b := r.exprReturnsRejectedTrackedPromise(cond.WhenFalse, tb, isAsync, depth+1)
			if a > 0 && b > 0 {
				return min(a, b)
			}
		}
	case ast.KindIdentifier:
		if decl := r.singleSameFileValueDeclaration(r.resolveSymbol(expr)); decl != nil &&
			decl.Kind == ast.KindVariableDeclaration && ast.IsVarConst(decl) &&
			decl.AsVariableDeclaration().Initializer != nil &&
			!nodeContains(decl.AsVariableDeclaration().Initializer, expr) {
			return r.exprReturnsRejectedTrackedPromise(decl.AsVariableDeclaration().Initializer, tb, isAsync, depth+1)
		}
	}
	return 0
}

func (r *AsyncAssertionResolver) statementAlwaysRejectsTracked(stmt *ast.Node, tb *trackedBindings, isAsync bool) int {
	if stmt == nil {
		return 0
	}
	switch stmt.Kind {
	case ast.KindThrowStatement:
		if r.exprEvaluatesToTrackedPromise(stmt.AsThrowStatement().Expression, tb) {
			return 1
		}
		return 0
	case ast.KindReturnStatement:
		return r.exprReturnsRejectedTrackedPromise(stmt.AsReturnStatement().Expression, tb, isAsync, 0)
	case ast.KindExpressionStatement:
		if isAsync {
			if expr := internalUtils.SkipAssertionsAndParens(stmt.AsExpressionStatement().Expression); expr != nil &&
				expr.Kind == ast.KindAwaitExpression {
				return r.exprReturnsRejectedTrackedPromise(expr.AsAwaitExpression().Expression, tb, true, 0)
			}
		}
		return 0
	case ast.KindBlock:
		block := stmt.AsBlock()
		if block == nil || block.Statements == nil {
			return 0
		}
		for _, s := range block.Statements.Nodes {
			if s == nil {
				continue
			}
			if d := r.statementAlwaysRejectsTracked(s, tb, isAsync); d > 0 {
				return d
			}
			if s.Kind == ast.KindIfStatement {
				ifStmt := s.AsIfStatement()
				if ifStmt != nil && ifStmt.ElseStatement == nil &&
					r.statementAlwaysRejectsTracked(ifStmt.ThenStatement, tb, isAsync) > 0 {
					continue
				}
			}
			if r.statementMayExitEarly(s) {
				return 0
			}
		}
		return 0
	case ast.KindIfStatement:
		ifStmt := stmt.AsIfStatement()
		if ifStmt == nil || ifStmt.ElseStatement == nil {
			return 0
		}
		a := r.statementAlwaysRejectsTracked(ifStmt.ThenStatement, tb, isAsync)
		b := r.statementAlwaysRejectsTracked(ifStmt.ElseStatement, tb, isAsync)
		if a > 0 && b > 0 {
			return min(a, b)
		}
		return 0
	}
	return 0
}

// inlineCallbackPropagatesRejection reports whether an inline rejection handler
// unconditionally re-rejects with the tracked parameter (`throw e`,
// `return Promise.reject(e)`, `(e) => Promise.reject(e)`, or `async (e) => { await Promise.reject(e); }`).
// Returns the number of rejection layers produced (0 if not unconditionally propagating).
func (r *AsyncAssertionResolver) inlineCallbackPropagatesRejection(cb *ast.Node, argIdx int) int {
	if cb == nil || argIdx < 1 || functionAsteriskToken(cb) != nil {
		return 0
	}
	body := cb.Body()
	if body == nil {
		return 0
	}
	isAsync := IsAsyncFunctionLike(cb)
	checkBody := func(tb *trackedBindings) int {
		if body.Kind != ast.KindBlock {
			return r.exprReturnsRejectedTrackedPromise(body, tb, isAsync, 0)
		}
		return r.statementAlwaysRejectsTracked(body, tb, isAsync)
	}
	summary := r.getInlineCallbackSummary(cb)
	if summary.usesArguments {
		tb := r.buildCallbackTrackedBindings(cb, nil, 0, argIdx-1, true, nil, 0)
		if d := checkBody(tb); d > 0 {
			return d
		}
	}
	params := runtimeParameters(cb)
	for i, param := range params {
		if param == nil {
			continue
		}
		if p := param.AsParameterDeclaration(); p != nil && p.DotDotDotToken != nil && i < argIdx {
			tb := r.buildCallbackTrackedBindings(cb, param, i, argIdx-1-i, false, nil, 0)
			if !tb.isEmpty() {
				return checkBody(tb)
			}
			return 0
		}
	}
	if len(params) >= argIdx && params[argIdx-1] != nil {
		tb := r.buildCallbackTrackedBindings(cb, params[argIdx-1], argIdx-1, -1, false, nil, 0)
		if !tb.isEmpty() {
			return checkBody(tb)
		}
	}
	return 0
}

// inlineCallbackCanConsumePromiseArg reports whether an inline `Promise.try`
// or `.catch` callback can actually await or return the forwarded argument at
// `argIdx`. Callbacks that ignore the argument (e.g. `() => {}`,
// `(first, second) => first`), generator callbacks, or bodies that never await
// or return a Promise derived from `argIdx` cannot keep the assertion pending.
func (r *AsyncAssertionResolver) inlineCallbackCanConsumePromiseArg(cb *ast.Node, argIdx int, passedArgs map[int]*ast.Node, firstUnknownSpreadIdx int) bool {
	if cb == nil || argIdx < 1 {
		return false
	}
	summary := r.getInlineCallbackSummary(cb)
	if !summary.bodyCanConsume {
		return false
	}
	if summary.usesArguments {
		tb := r.buildCallbackTrackedBindings(cb, nil, 0, argIdx-1, true, passedArgs, firstUnknownSpreadIdx)
		if r.scopeConsumesTracked(cb.Body(), tb, true) {
			return true
		}
	}
	params := summary.params
	if summary.restParamIdx >= 0 && summary.restParamIdx < argIdx {
		param := params[summary.restParamIdx]
		p := param.AsParameterDeclaration()
		if p != nil && p.Name() != nil && p.Name().Kind == ast.KindIdentifier {
			whole, cached := summary.restWholeConsumed[param]
			if !cached {
				whole = r.isParamConsumedByCallback(cb, param, summary.restParamIdx, -1, nil, 1)
				if summary.restWholeConsumed == nil {
					summary.restWholeConsumed = map[*ast.Node]bool{}
				}
				summary.restWholeConsumed[param] = whole
			}
			if whole {
				return true
			}
		}
		return r.isParamConsumedByCallback(cb, param, summary.restParamIdx, argIdx-1-summary.restParamIdx, passedArgs, firstUnknownSpreadIdx)
	}
	if len(params) >= argIdx {
		return r.isParamConsumedByCallback(cb, params[argIdx-1], argIdx-1, -1, passedArgs, firstUnknownSpreadIdx)
	}
	return false
}

// inlineCallbackCanConsumeSpreadArg reports whether `cb` is guaranteed to
// consume an argument whose exact runtime position is >= `minRuntimeIdx` due
// to a preceding unknown-length spread. It requires either `arguments` usage
// or a consumed rest parameter plus consumption of every earlier positional
// parameter from `minRuntimeIdx-1` up to the rest parameter.
func (r *AsyncAssertionResolver) inlineCallbackCanConsumeSpreadArg(cb *ast.Node, minRuntimeIdx int) bool {
	if cb == nil || minRuntimeIdx < 1 {
		return false
	}
	summary := r.getInlineCallbackSummary(cb)
	if !summary.bodyCanConsume {
		return false
	}
	if summary.usesArguments {
		tb := r.buildCallbackTrackedBindings(cb, nil, 0, -1, true, nil, 1)
		if r.scopeConsumesTracked(cb.Body(), tb, true) {
			return true
		}
	}
	params := summary.params
	if summary.restParamIdx >= 0 {
		i := summary.restParamIdx
		param := params[i]
		p := param.AsParameterDeclaration()
		if p != nil && p.Name() != nil && p.Name().Kind == ast.KindArrayBindingPattern {
			if !r.isRestArrayPatternConsumedForSpread(cb, param, i, minRuntimeIdx-1-i) {
				return false
			}
		} else if !r.isParamConsumedByCallback(cb, param, i, -1, nil, 1) {
			return false
		}
		for j := minRuntimeIdx - 1; j < i; j++ {
			if !r.isParamConsumedByCallback(cb, params[j], j, -1, nil, 1) {
				return false
			}
		}
		return true
	}
	return false
}

// outermostTransparentExpression walks upward across parentheses and
// TypeScript assertion wrappers (`as`, `<T>`, `satisfies`, `!`) that are erased
// at runtime and preserve the expression's Promise value.
func outermostTransparentExpression(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if ast.IsParenthesizedExpression(parent) {
			if paren := parent.AsParenthesizedExpression(); paren != nil && paren.Expression == current {
				current = parent
				continue
			}
		} else if expr, ok := internalUtils.TransparentExpression(parent); ok && expr == current {
			current = parent
			continue
		}
		break
	}
	return current
}

func outermostChainedBase(node *ast.Node) *ast.Node {
	outer := outermostTransparentExpression(node)
	for outer != nil && outer.Parent != nil {
		parent := outer.Parent
		if parent.Kind == ast.KindConditionalExpression {
			cond := parent.AsConditionalExpression()
			if cond != nil && (cond.WhenTrue == outer || cond.WhenFalse == outer) {
				outer = outermostTransparentExpression(parent)
				continue
			}
			return nil
		}
		if internalUtils.IsCommaOperator(parent) && parent.AsBinaryExpression().Right == outer {
			outer = outermostTransparentExpression(parent)
			continue
		}
		break
	}
	return outer
}

func (r *AsyncAssertionResolver) isNonCallablePromiseHandler(node *ast.Node) bool {
	if node == nil {
		return true
	}
	switch node.Kind {
	case ast.KindNullKeyword, ast.KindUndefinedKeyword, ast.KindVoidExpression,
		ast.KindTrueKeyword, ast.KindFalseKeyword,
		ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression,
		ast.KindRegularExpressionLiteral,
		ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression, ast.KindTypeOfExpression:
		return true
	case ast.KindIdentifier:
		return node.AsIdentifier().Text == "undefined" && !r.isLocallyValueShadowed(node, "undefined")
	}
	return false
}

// staticCallArgAt resolves the 0-based argument `targetIdx` of `call`,
// unpacking static array literal spreads (`...[]`, `...[undefined, fn]`).
// Returns `(nil, true)` when the call statically passes fewer than `targetIdx+1`
// arguments (so the slot is `undefined`), and `(nil, false)` when an unknown
// spread prevents determining the argument statically.
func staticCallArgAt(call *ast.Node, targetIdx int) (*ast.Node, bool) {
	if call == nil || targetIdx < 0 {
		return nil, false
	}
	var args *ast.NodeList
	switch call.Kind {
	case ast.KindCallExpression:
		args = call.AsCallExpression().Arguments
	case ast.KindNewExpression:
		args = call.AsNewExpression().Arguments
	default:
		return nil, false
	}
	if args == nil {
		return nil, true
	}
	curIdx := 0
	var result *ast.Node
	found := false
	unknown := false
	var walk func(nodes []*ast.Node, depth int)
	walk = func(nodes []*ast.Node, depth int) {
		for _, a := range nodes {
			if found || unknown {
				return
			}
			if a == nil || a.Kind == ast.KindOmittedExpression {
				if curIdx == targetIdx {
					found = true
					return
				}
				curIdx++
				continue
			}
			if a.Kind == ast.KindSpreadElement {
				spreadExpr := internalUtils.SkipAssertionsAndParens(a.AsSpreadElement().Expression)
				if depth < 4 && spreadExpr != nil && spreadExpr.Kind == ast.KindArrayLiteralExpression {
					if elems := spreadExpr.AsArrayLiteralExpression().Elements; elems != nil {
						walk(elems.Nodes, depth+1)
					}
					continue
				}
				unknown = true
				return
			}
			if curIdx == targetIdx {
				result = a
				found = true
				return
			}
			curIdx++
		}
	}
	walk(args.Nodes, 0)
	if unknown {
		return nil, false
	}
	return result, true
}

type rejectChainOutcome uint8

const (
	rejectPropagated rejectChainOutcome = iota
	rejectRecovered
	rejectDropped
)

// walkPromiseRejectChain walks `.then`/`.catch`/`.finally` chains starting from
// a rejected Promise whose rejection reason is an assertion Promise. Unlike
// `Promise.resolve(assertion)`, `Promise.reject(assertion)` rejects immediately
// with the still-pending `assertion` Promise as its reason:
//   - a `.catch(cb)` or 2-arg `.then(_, cb)` whose inline handler returns or
//     awaits the reason recovers it (`rejectRecovered`);
//   - a handler that unconditionally rethrows the reason, a non-callable
//     handler, or a non-throwing `.finally` propagates the rejection
//     (`rejectPropagated`);
//   - a handler that swallows the reason or a `.finally` that unconditionally
//     throws drops the assertion (`rejectDropped`).
func (r *AsyncAssertionResolver) walkPromiseRejectChain(start *ast.Node) (*ast.Node, rejectChainOutcome, int) {
	cur := start
	extraRejects := 0
	for cur != nil {
		outer := outermostChainedBase(cur)
		if outer == nil || outer.Parent == nil ||
			!ast.IsAccessExpression(outer.Parent) ||
			internalUtils.AccessExpressionObject(outer.Parent) != outer {
			return cur, rejectPropagated, extraRejects
		}
		memberOuter := outermostTransparentExpression(outer.Parent)
		if memberOuter == nil || memberOuter.Parent == nil ||
			memberOuter.Parent.Kind != ast.KindCallExpression ||
			memberOuter.Parent.AsCallExpression().Expression != memberOuter {
			return cur, rejectPropagated, extraRejects
		}
		grandParent := memberOuter.Parent
		method, ok := r.accessStaticNameThroughAssertions(outer.Parent)
		if !ok {
			return cur, rejectPropagated, extraRejects
		}
		var onRejected *ast.Node
		var argOK bool
		switch method {
		case "finally":
			if fnArg, ok := staticCallArgAt(grandParent, 0); ok && fnArg != nil {
				fnArg = internalUtils.SkipAssertionsAndParens(fnArg)
				if (fnArg.Kind == ast.KindArrowFunction || fnArg.Kind == ast.KindFunctionExpression) &&
					functionAsteriskToken(fnArg) == nil && r.statementAlwaysThrows(fnArg.Body(), nil) {
					return grandParent, rejectDropped, extraRejects
				}
			}
			cur = grandParent
			continue
		case "then":
			onRejected, argOK = staticCallArgAt(grandParent, 1)
		case "catch":
			onRejected, argOK = staticCallArgAt(grandParent, 0)
		default:
			return cur, rejectPropagated, extraRejects
		}
		if !argOK {
			// Unknown spread in `.then`/`.catch`: stop conservatively.
			return grandParent, rejectRecovered, extraRejects
		}
		onRejected = internalUtils.SkipAssertionsAndParens(onRejected)
		if r.isNonCallablePromiseHandler(onRejected) {
			cur = grandParent
			continue
		}
		if onRejected.Kind == ast.KindArrowFunction || onRejected.Kind == ast.KindFunctionExpression {
			// Check unconditional rejection propagation (`throw e`, `(e) => Promise.reject(e)`)
			// BEFORE general consumption so `Promise.reject(e)` is not mistaken for recovery.
			if rejDepth := r.inlineCallbackPropagatesRejection(onRejected, 1); rejDepth > 0 {
				extraRejects += rejDepth - 1
				cur = grandParent
				continue
			}
			if r.inlineCallbackCanConsumePromiseArg(onRejected, 1, nil, 0) {
				return grandParent, rejectRecovered, extraRejects
			}
			return grandParent, rejectDropped, extraRejects
		}
		// Non-inline rejection handler (e.g. identifier): conservatively assume it handles the reason.
		return grandParent, rejectRecovered, extraRejects
	}
	return cur, rejectPropagated, extraRejects
}

func (r *AsyncAssertionResolver) getPromiseCallExpressionNode(node *ast.Node, directArg *ast.Node) (*ast.Node, string) {
	if node == nil {
		return nil, ""
	}

	if node.Kind == ast.KindArrayLiteralExpression {
		outer := outermostTransparentExpression(node)
		if outer == nil || outer.Parent == nil || outer.Parent.Kind != ast.KindCallExpression {
			return nil, ""
		}
		call := outer.Parent
		firstArg, argOK := staticCallArgAt(call, 0)
		if !argOK || firstArg != outer {
			return nil, ""
		}
		method, ok := r.promiseMethodName(call)
		// `Promise.resolve([...])`, `Promise.reject([...])`,
		// `Promise.withResolvers([...])`, and `Promise.try([...])` do not await
		// array elements, unlike `Promise.all` / `allSettled` / `race` / `any`
		// (or custom combinators such as `Promise.x`).
		if !ok || method == "" || method == "resolve" || method == "reject" || method == "withResolvers" || method == "try" {
			return nil, ""
		}
		return call, method
	}

	method, ok := r.promiseMethodName(node)
	if !ok || method == "" || method == "withResolvers" {
		return nil, ""
	}
	if directArg != nil {
		args := node.AsCallExpression().Arguments
		switch method {
		case "resolve", "reject", "all", "allSettled", "race", "any":
			firstArg, argOK := staticCallArgAt(node, 0)
			if !argOK || firstArg != directArg {
				return nil, ""
			}
		case "try":
			if args == nil || len(args.Nodes) == 0 {
				return nil, ""
			}
			tryIdx := r.getPromiseTryCallIndex(node)
			pos, ok := tryIdx.args[directArg]
			if !ok || tryIdx.cbUnknown || tryIdx.cbMissing || pos.runtimeIdx < 1 {
				return nil, ""
			}
			cb := tryIdx.cb
			if cb == nil || r.isNonCallablePromiseHandler(cb) {
				return nil, ""
			}
			if cb.Kind == ast.KindArrowFunction || cb.Kind == ast.KindFunctionExpression {
				if pos.hasUnknownSpread {
					if !r.inlineCallbackCanConsumeSpreadArg(cb, pos.runtimeIdx) {
						return nil, ""
					}
				} else if !r.inlineCallbackCanConsumePromiseArg(cb, pos.runtimeIdx, tryIdx.argExprByIdx, tryIdx.firstUnknownSpreadIdx) {
					return nil, ""
				}
			}
		}
	}
	return node, method
}

func (r *AsyncAssertionResolver) findPromiseCallExpressionNode(node *ast.Node) (*ast.Node, string) {
	current := outermostTransparentExpression(node)
	for current != nil && current.Parent != nil {
		parent := current.Parent
		// A conditional branch or comma-operator RHS inside an array/call
		// (`Promise.all([cond ? expect(...) : ...])`, `Promise.all([(log(), expect(...))])`)
		// forwards the assertion promise to the enclosing container.
		if parent.Kind == ast.KindConditionalExpression {
			cond := parent.AsConditionalExpression()
			if cond != nil && (cond.WhenTrue == current || cond.WhenFalse == current) {
				current = outermostTransparentExpression(parent)
				continue
			}
			return nil, ""
		}
		if internalUtils.IsCommaOperator(parent) && parent.AsBinaryExpression().Right == current {
			current = outermostTransparentExpression(parent)
			continue
		}
		if parent.Kind == ast.KindCallExpression {
			if parent.AsCallExpression().Expression == current {
				return nil, ""
			}
			return r.getPromiseCallExpressionNode(parent, current)
		}
		if parent.Kind == ast.KindArrayLiteralExpression {
			arr := parent
			for depth := 0; depth < 4 && arr != nil; depth++ {
				arrOuter := outermostTransparentExpression(arr)
				if arrOuter == nil || arrOuter.Parent == nil {
					return nil, ""
				}
				if arrOuter.Parent.Kind == ast.KindCallExpression {
					return r.getPromiseCallExpressionNode(arr, nil)
				}
				if arrOuter.Parent.Kind == ast.KindSpreadElement && arrOuter.Parent.AsSpreadElement().Expression == arrOuter {
					sp := arrOuter.Parent
					if sp.Parent == nil {
						return nil, ""
					}
					if sp.Parent.Kind == ast.KindCallExpression {
						// Spread-wrapped argument to `Promise.try(cb, ...[expect(...)])`.
						return r.getPromiseCallExpressionNode(sp.Parent, current)
					}
					if sp.Parent.Kind == ast.KindArrayLiteralExpression {
						arr = sp.Parent
						continue
					}
				}
				return nil, ""
			}
			return nil, ""
		}
		return nil, ""
	}
	return nil, ""
}

func getParentIfPromiseChained(node *ast.Node) *ast.Node {
	outer := outermostChainedBase(node)
	if outer == nil || outer.Parent == nil {
		return node
	}

	member := outer.Parent
	if !ast.IsAccessExpression(member) || internalUtils.AccessExpressionObject(member) != outer {
		return node
	}

	outerMember := outermostTransparentExpression(member)
	grandParent := outerMember.Parent
	if grandParent == nil ||
		grandParent.Kind != ast.KindCallExpression ||
		grandParent.AsCallExpression().Expression != outerMember {
		return node
	}

	last, ok := accessStaticNameThroughAssertions(member)
	if ok && (last == "then" || last == "catch" || last == "finally") {
		return getParentIfPromiseChained(grandParent)
	}

	return node
}

func isPromiseChainMethodCall(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return false
	}
	name, ok := accessStaticNameThroughAssertions(callee)
	return ok && (name == "then" || name == "catch" || name == "finally")
}

func isAcceptableReturnNode(child, parent *ast.Node, allowReturn bool) bool {
	if parent == nil {
		return false
	}

	if allowReturn && parent.Kind == ast.KindReturnStatement {
		return parent.AsReturnStatement().Expression == child
	}
	if parent.Kind == ast.KindConditionalExpression {
		cond := parent.AsConditionalExpression()
		// Only a value-producing branch can forward the assertion promise;
		// testing the promise in the condition discards it as always-truthy.
		if cond == nil || (cond.WhenTrue != child && cond.WhenFalse != child) {
			return false
		}
		outer := outermostTransparentExpression(parent)
		return isAcceptableReturnNode(outer, outer.Parent, allowReturn)
	}
	if internalUtils.IsCommaOperator(parent) && parent.AsBinaryExpression().Right == child {
		outer := outermostTransparentExpression(parent)
		return isAcceptableReturnNode(outer, outer.Parent, allowReturn)
	}
	if parent.Kind == ast.KindArrowFunction {
		return parent.AsArrowFunction().Body == child
	}
	if parent.Kind == ast.KindAwaitExpression {
		return parent.AsAwaitExpression().Expression == child
	}

	return false
}

func isTerminalFunctionReturn(ret *ast.Node) bool {
	if ret == nil || ret.Kind != ast.KindReturnStatement || ret.Parent == nil || ret.Parent.Kind != ast.KindBlock {
		return false
	}
	block := ret.Parent
	fn := block.Parent
	if fn == nil || !testFramework.IsFunction(fn) {
		return false
	}
	if block.AsBlock() == nil || block.AsBlock().Statements == nil {
		return false
	}
	stmts := block.AsBlock().Statements.Nodes
	return len(stmts) > 0 && stmts[len(stmts)-1] == ret
}

// isSafeTerminalReturnStrip reports whether `return expr` can safely be
// rewritten to `await expr` (dropping `return`) rather than `return await expr`.
// Only anonymous callbacks passed directly as call arguments (such as
// `test("...", () => { return expect(...); })`) returning a plain assertion
// promise strip `return`. Value-returning helpers, assigned functions,
// object/class methods (e.g. `expect.extend`), `.then()`/`Promise.*` wrappers
// that may resolve to a non-undefined value, and functions with explicit return
// type annotations keep `return await`.
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
	fnOuter := outermostTransparentExpression(fn)
	if fnOuter == nil || fnOuter.Parent == nil ||
		fnOuter.Parent.Kind != ast.KindCallExpression ||
		fnOuter.Parent.AsCallExpression().Expression == fnOuter {
		return false
	}
	unwrapped := internalUtils.SkipAssertionsAndParens(expr)
	if _, isPromiseWrapper := r.promiseMethodName(unwrapped); isPromiseWrapper {
		return false
	}
	if isPromiseChainMethodCall(unwrapped) {
		return false
	}
	return true
}

// FindFixableContainingFunction returns the enclosing function that can safely
// receive `await` (and `async` if not already async) for node. It returns
// (nil, false) when node sits inside a scope where `await` or `async` is
// illegal or would break callers (constructors, getters, setters, default
// parameters, class field initializers, class static blocks, sync generators,
// Annex B single-statement function declarations, directly constructed
// `new (function () {})()`, functions using `await` as an identifier, or
// non-async functions with explicit non-Promise return type annotations).
func FindFixableContainingFunction(node *ast.Node) (*ast.Node, bool) {
	return FindFixableContainingFunctionCached(node, nil)
}

// FindFixableContainingFunctionCached is the memoized form of
// FindFixableContainingFunction. Passing a per-file cache map avoids rescanning
// a large synchronous function body for `await` identifiers on every diagnostic.
func FindFixableContainingFunctionCached(node *ast.Node, fnAllowed map[*ast.Node]bool) (*ast.Node, bool) {
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
			// `await` is a syntax error inside parameter default initializers
			// even when the function itself is async.
			return nil, false
		case ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindMethodSignature, ast.KindCallSignature, ast.KindConstructSignature,
			ast.KindFunctionType, ast.KindConstructorType:
			// Decorators and computed property keys evaluate in the enclosing
			// scope rather than inside this member's body.
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
				return current, true
			}
			allowed := isFunctionSafelyAsyncFixable(current)
			if fnAllowed != nil {
				fnAllowed[current] = allowed
			}
			if !allowed {
				return nil, false
			}
			return current, true
		}
	}
	return nil, false
}

func isFunctionSafelyAsyncFixable(fn *ast.Node) bool {
	if ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync) {
		return true
	}
	if functionAsteriskToken(fn) != nil {
		// Converting a synchronous generator into an async generator changes
		// its runtime return protocol.
		return false
	}
	if hasExplicitNonPromiseReturnType(fn) {
		return false
	}
	if fn.Kind == ast.KindFunctionDeclaration && !isValidAsyncFunctionDeclarationParent(fn.Parent) {
		return false
	}
	if fn.Kind == ast.KindFunctionExpression {
		fnOuter := outermostTransparentExpression(fn)
		if fnOuter != nil && fnOuter.Parent != nil &&
			fnOuter.Parent.Kind == ast.KindNewExpression &&
			fnOuter.Parent.AsNewExpression().Expression == fnOuter {
			return false
		}
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

// containsIdentifierNamedAwait reports whether a non-async function uses
// `await` as its own name, an inner declaration name, a parameter, a binding,
// or an expression identifier in its own scope, where adding `async` would
// turn `await` into a reserved keyword syntax error.
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
				// The function's name is declared in `fn`'s own scope, while
				// its parameters and body form a nested non-async scope.
				if n := node.Name(); n != nil && n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == "await" {
					found = true
				}
				return
			case ast.KindFunctionExpression, ast.KindArrowFunction:
				return
			case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
				// Decorators and computed property keys of inner methods
				// evaluate in `fn`'s own scope.
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
					ast.KindGetAccessor, ast.KindSetAccessor, ast.KindBindingElement:
					if parent.Name() == node && parent.Kind != ast.KindBindingElement {
						return
					}
					if parent.Kind == ast.KindBindingElement && parent.AsBindingElement().PropertyName == node {
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

// IsAsyncFunctionLike reports whether fn already carries the `async` modifier
// (including `async function*` and `async *m()`, unlike `ast.IsAsyncFunction`).
func IsAsyncFunctionLike(fn *ast.Node) bool {
	return fn != nil && ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync)
}

func hasExplicitNonPromiseReturnType(fn *ast.Node) bool {
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

// PromiseArrayExceptionKey identifies a Promise.all([...]) call by source range,
// so the assertions inside its array report once between them.
func PromiseArrayExceptionKey(sourceFile *ast.SourceFile, node *ast.Node) string {
	if sourceFile == nil || node == nil {
		return ""
	}
	r := internalUtils.TrimNodeTextRange(sourceFile, node)
	return fmt.Sprintf("%d:%d", r.Pos(), r.End())
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
		// Insert `async ` after all modifiers/decorators (e.g. `static`,
		// `public`, `override`), right before `*` or the method name.
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

// CanSafelyAwaitFix reports whether inserting `await` at node preserves the
// test's control flow, syntax validity, and downstream value expectations.
func CanSafelyAwaitFix(node *ast.Node) bool {
	return (*AsyncAssertionResolver)(nil).CanSafelyAwaitFix(node)
}

// CanSafelyAwaitFix is the resolver-backed form of CanSafelyAwaitFix, reusing
// per-file Promise alias and static-string caches.
func (r *AsyncAssertionResolver) CanSafelyAwaitFix(node *ast.Node) bool {
	if node == nil {
		return false
	}
	outer := outermostTransparentExpression(node)
	if outer == nil || outer.Parent == nil {
		return false
	}

	// Inserting `await` directly in front of an unhandled `Promise.reject(assertion)`
	// (including wrapped/chained forms such as `Promise.resolve(Promise.reject(...))`
	// or `Promise.reject(...).finally(() => {})`) is only safe at a top-level
	// statement/return outside `try/catch` or an abruptly exiting `finally` block
	// (never inside `Promise.allSettled([...])`, `Promise.any([...])`, or `try/catch`).
	if r.unhandledPromiseRejectDepth(outer, 0) > 0 {
		if r.isInsideTryWithCatchOrAbruptFinally(outer) {
			return false
		}
		switch outer.Parent.Kind {
		case ast.KindExpressionStatement:
			return outer.Parent.AsExpressionStatement().Expression == outer
		case ast.KindReturnStatement:
			return outer.Parent.AsReturnStatement().Expression == outer
		case ast.KindArrowFunction:
			return outer.Parent.AsArrowFunction().Body == outer
		default:
			return false
		}
	}

	hasExplicitNonAnyType := func(typeNode *ast.Node) bool {
		return typeNode != nil && typeNode.Kind != ast.KindAnyKeyword && typeNode.Kind != ast.KindUnknownKeyword
	}

	isSafeVarDecl := func(vdNode *ast.Node, initExpr *ast.Node) bool {
		vd := vdNode.AsVariableDeclaration()
		if vd == nil || vd.Initializer != initExpr || vd.Name() == nil || vd.Name().Kind != ast.KindIdentifier {
			return false
		}
		if hasExplicitNonAnyType(vd.Type) {
			return false
		}
		if r.isBindingReadOrExported(vdNode, vd.Name()) {
			return false
		}
		if vdNode.Parent == nil || vdNode.Parent.Kind != ast.KindVariableDeclarationList {
			return false
		}
		declList := vdNode.Parent
		if declList.Parent == nil || declList.Parent.Kind != ast.KindVariableStatement {
			return false
		}
		if r.hasFollowingSideEffectDecl(declList, vdNode) {
			return false
		}
		return true
	}

	isStatementLevelPromiseCall := func(call *ast.Node, method string) bool {
		callOuter := outermostTransparentExpression(call)
		if callOuter != nil && callOuter.Parent != nil && callOuter.Parent.Kind == ast.KindAwaitExpression &&
			callOuter.Parent.AsAwaitExpression().Expression == callOuter {
			callOuter = outermostTransparentExpression(callOuter.Parent)
		}
		if callOuter == nil || callOuter.Parent == nil || callOuter.Parent.Kind != ast.KindExpressionStatement ||
			callOuter.Parent.AsExpressionStatement().Expression != callOuter {
			return false
		}
		if method == "reject" && r.isInsideTryWithCatchOrAbruptFinally(callOuter.Parent) {
			return false
		}
		return true
	}

	// Walk all the way up to the enclosing statement/function boundary through
	// an explicit allowlist of safe containers so inserting `await` can never
	// delay sibling setup mutations, mutating callees/constructors, `Promise.try`
	// callbacks, returned containers, or observed properties/bindings across a
	// microtask yield.
	cur := outer
	fromComputedProp := false
	inContainer := false
	for cur != nil && cur.Parent != nil {
		parent := cur.Parent
		switch parent.Kind {
		case ast.KindExpressionStatement:
			return parent.AsExpressionStatement().Expression == cur
		case ast.KindReturnStatement:
			if inContainer {
				return false
			}
			return parent.AsReturnStatement().Expression == cur
		case ast.KindArrowFunction:
			if inContainer {
				return false
			}
			return parent.AsArrowFunction().Body == cur
		case ast.KindVariableDeclaration:
			return isSafeVarDecl(parent, cur)
		case ast.KindAwaitExpression:
			if cur == outer || parent.AsAwaitExpression().Expression != cur {
				return false
			}
			cur = outermostTransparentExpression(parent)
			continue
		case ast.KindConditionalExpression:
			cond := parent.AsConditionalExpression()
			if cond == nil || (cond.WhenTrue != cur && cond.WhenFalse != cur) {
				return false
			}
			cur = outermostTransparentExpression(parent)
			continue
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil {
				return false
			}
			switch bin.OperatorToken.Kind {
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				if bin.Right != cur {
					return false
				}
				cur = outermostTransparentExpression(parent)
				continue
			case ast.KindCommaToken:
				if bin.Right == cur {
					cur = outermostTransparentExpression(parent)
					continue
				}
				if bin.Left == cur && !r.exprMayMutateState(bin.Right) {
					cur = outermostTransparentExpression(parent)
					continue
				}
				return false
			case ast.KindEqualsToken, ast.KindQuestionQuestionEqualsToken,
				ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
				if bin.Right != cur {
					return false
				}
				lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
				if r.isAssignmentTargetObserved(lhs) {
					return false
				}
				if lhs != nil && lhs.Kind == ast.KindIdentifier {
					if sym := r.resolveSymbol(lhs); sym != nil {
						for _, decl := range sym.Declarations {
							if decl == nil {
								continue
							}
							if decl.Kind != ast.KindVariableDeclaration ||
								decl.Parent == nil || decl.Parent.Kind != ast.KindVariableDeclarationList ||
								decl.Parent.Parent == nil || decl.Parent.Parent.Kind != ast.KindVariableStatement {
								return false
							}
							vd := decl.AsVariableDeclaration()
							if vd == nil {
								return false
							}
							if vd.Type != nil {
								if hasExplicitNonAnyType(vd.Type) {
									return false
								}
							} else if vd.Initializer != nil {
								return false
							}
						}
					}
				}
				cur = outermostTransparentExpression(parent)
				continue
			}
			return false
		case ast.KindCallExpression:
			if parent.AsCallExpression().Expression == cur {
				return false
			}
			if r.isPromiseTryOrRejectCall(parent) {
				return false
			}
			if r.calleeMayMutateState(parent.AsCallExpression().Expression, 0) {
				return false
			}
			if r.hasFollowingSideEffectArg(parent, cur) {
				return false
			}
			if method, ok := r.promiseMethodName(parent); ok && (method == "resolve" || method == "reject") {
				if !isStatementLevelPromiseCall(parent, method) {
					return false
				}
			}
			inContainer = true
			cur = outermostTransparentExpression(parent)
			continue
		case ast.KindNewExpression:
			if parent.AsNewExpression().Expression == cur {
				return false
			}
			if r.calleeMayMutateState(parent.AsNewExpression().Expression, 0) {
				return false
			}
			if r.hasFollowingSideEffectArg(parent, cur) {
				return false
			}
			inContainer = true
			cur = outermostTransparentExpression(parent)
			continue
		case ast.KindArrayLiteralExpression:
			if r.hasFollowingSideEffectElem(parent, cur) {
				return false
			}
			arrOuter := outermostTransparentExpression(parent)
			if arrOuter != nil && arrOuter.Parent != nil && arrOuter.Parent.Kind == ast.KindCallExpression {
				call := arrOuter.Parent
				if method, ok := r.promiseMethodName(call); ok && (method == "resolve" || method == "reject") {
					if !isStatementLevelPromiseCall(call, method) {
						return false
					}
				}
			}
			inContainer = true
			cur = outermostTransparentExpression(parent)
			continue
		case ast.KindSpreadElement:
			if parent.AsSpreadElement().Expression != cur {
				return false
			}
			inContainer = true
			cur = outermostTransparentExpression(parent)
			continue
		case ast.KindComputedPropertyName:
			if parent.AsComputedPropertyName().Expression != cur {
				return false
			}
			fromComputedProp = true
			inContainer = true
			cur = parent
			continue
		case ast.KindMethodDeclaration:
			if !fromComputedProp || parent.Name() != cur {
				return false
			}
			cur = parent
			continue
		case ast.KindObjectLiteralExpression:
			if !fromComputedProp {
				return false
			}
			fromComputedProp = false
			if r.hasFollowingSideEffectObjProp(parent, cur) {
				return false
			}
			inContainer = true
			cur = outermostTransparentExpression(parent)
			continue
		default:
			return false
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isPromiseTryOrRejectCallee(callee *ast.Node, depth int) bool {
	if depth > 4 || callee == nil {
		return false
	}
	callee = internalUtils.SkipAssertionsAndParens(callee)
	if callee == nil {
		return false
	}
	if ast.IsAccessExpression(callee) {
		obj := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
		method, ok := r.accessStaticNameThroughAssertions(callee)
		if r.isGlobalPromiseObject(obj, 0) && (!ok || method == "try" || method == "reject") {
			return true
		}
		// `Promise.try.call(...)`, `Promise.try.apply(...)`, `Promise.try.bind(...)`.
		if ok && (method == "call" || method == "apply" || method == "bind") {
			return r.isPromiseTryOrRejectCallee(obj, depth+1)
		}
		return false
	}
	if callee.Kind == ast.KindIdentifier {
		decl := r.singleSameFileValueDeclaration(r.resolveSymbol(callee))
		if decl == nil {
			return false
		}
		switch decl.Kind {
		case ast.KindVariableDeclaration:
			if ast.IsVarConst(decl) {
				init := decl.AsVariableDeclaration().Initializer
				if init != nil && !nodeContains(init, callee) {
					return r.isPromiseTryOrRejectCallee(init, depth+1)
				}
			}
		case ast.KindBindingElement:
			b := decl.AsBindingElement()
			if b == nil || b.DotDotDotToken != nil || decl.Parent == nil || decl.Parent.Kind != ast.KindObjectBindingPattern {
				return false
			}
			vd := internalUtils.EnclosingVariableDeclarationOfBindingElement(decl)
			if vd == nil || !ast.IsVarConst(vd) || vd.AsVariableDeclaration().Initializer == nil ||
				!r.isGlobalPromiseObject(vd.AsVariableDeclaration().Initializer, 0) {
				return false
			}
			var propName string
			var ok bool
			if b.PropertyName != nil {
				propName, ok = internalUtils.GetStaticPropertyName(b.PropertyName)
				if !ok && r.staticStringEvaluator() != nil {
					propName, ok = r.staticStringEvaluator().EvalPropertyName(b.PropertyName)
				}
			} else if b.Name() != nil && b.Name().Kind == ast.KindIdentifier {
				propName, ok = b.Name().AsIdentifier().Text, true
			}
			return ok && (propName == "try" || propName == "reject")
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isPromiseTryOrRejectCall(call *ast.Node) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	if method, ok := r.promiseMethodName(call); ok && (method == "try" || method == "reject") {
		return true
	}
	callee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	if r.isPromiseTryOrRejectCallee(callee, 0) {
		return true
	}
	// `Reflect.apply(Promise.try, ...)`
	if callee != nil && ast.IsAccessExpression(callee) {
		if m, ok := r.accessStaticNameThroughAssertions(callee); ok && m == "apply" {
			obj := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
			if obj != nil && obj.Kind == ast.KindIdentifier && obj.AsIdentifier().Text == "Reflect" &&
				!r.isLocallyValueShadowed(obj, "Reflect") {
				if args := call.AsCallExpression().Arguments; args != nil && len(args.Nodes) > 0 {
					return r.isPromiseTryOrRejectCallee(args.Nodes[0], 0)
				}
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isExpectCallWithoutCallback(call *ast.Node) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	callee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	if callee == nil || callee.Kind != ast.KindIdentifier || callee.AsIdentifier().Text != "expect" {
		return false
	}
	// Reject locally shadowed `const expect = () => { started = true; }` (allow imports).
	if sym := r.resolveSymbol(callee); sym != nil {
		for _, decl := range sym.Declarations {
			if !isValueDeclarationNode(decl) {
				continue
			}
			if r != nil && r.sourceFile != nil && ast.GetSourceFileOfNode(decl) != r.sourceFile {
				continue
			}
			switch decl.Kind {
			case ast.KindImportClause, ast.KindImportSpecifier, ast.KindNamespaceImport, ast.KindImportEqualsDeclaration:
				continue
			default:
				return false
			}
		}
	}
	// Reject `expect(() => { started = true; })` since `.toThrow()` invokes the callback synchronously.
	if args := call.AsCallExpression().Arguments; args != nil {
		for _, arg := range args.Nodes {
			unwrapped := internalUtils.SkipAssertionsAndParens(arg)
			if unwrapped != nil && (unwrapped.Kind == ast.KindArrowFunction || unwrapped.Kind == ast.KindFunctionExpression) {
				return false
			}
		}
	}
	return true
}

func (r *AsyncAssertionResolver) rootChainedMapSetIdent(recv *ast.Node) *ast.Node {
	for depth := 0; depth < 16 && recv != nil; depth++ {
		u := internalUtils.SkipAssertionsAndParens(recv)
		if u == nil {
			return nil
		}
		if u.Kind == ast.KindIdentifier {
			return u
		}
		if u.Kind != ast.KindCallExpression {
			return nil
		}
		callee := internalUtils.SkipAssertionsAndParens(u.AsCallExpression().Expression)
		if callee == nil || !ast.IsAccessExpression(callee) {
			return nil
		}
		mName, ok := r.accessStaticNameThroughAssertions(callee)
		if !ok || mName != "set" {
			return nil
		}
		recv = internalUtils.AccessExpressionObject(callee)
	}
	return nil
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
	isExtendOrEqualityKey := func(s string) bool {
		return s == "extend" || s == "addEqualityTesters"
	}
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || found || ast.IsTypeNode(node) {
			return
		}
		if node.Kind == ast.KindIdentifier && isExtendOrEqualityKey(node.AsIdentifier().Text) {
			found = true
			return
		}
		if (node.Kind == ast.KindStringLiteral && isExtendOrEqualityKey(node.AsStringLiteral().Text)) ||
			(node.Kind == ast.KindNoSubstitutionTemplateLiteral && isExtendOrEqualityKey(node.Text())) {
			found = true
			return
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

func (r *AsyncAssertionResolver) isPurePromiseLiteralArg(arg *ast.Node) bool {
	return r.isPurePromiseLiteralArgDepth(arg, 0)
}

func (r *AsyncAssertionResolver) isPurePromiseLiteralArgDepth(arg *ast.Node, depth int) bool {
	if depth > 4 {
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
			if !r.isPurePrimitiveLiteral(a) && !r.isPurePromiseLiteralArgDepth(a, depth+1) {
				return false
			}
		}
	}
	return true
}

func (r *AsyncAssertionResolver) isPureLiteralExpectPromiseCall(call *ast.Node) bool {
	u := internalUtils.SkipAssertionsAndParens(call)
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
	outerMatcherName, hasOuterMatcher := r.accessStaticNameThroughAssertions(outerCallee)
	isBareAsyncMatcher := hasOuterMatcher && (outerMatcherName == "toResolve" || outerMatcherName == "toReject") && (args == nil || len(args.Nodes) == 0)
	hasAsyncModifier := false
	cur := outerCallee
	for cur != nil {
		if ast.IsAccessExpression(cur) {
			if cur.Kind == ast.KindElementAccessExpression && !r.isPurePrimitiveLiteral(cur.AsElementAccessExpression().ArgumentExpression) {
				return false
			}
			if cur != outerCallee {
				mName, ok := r.accessStaticNameThroughAssertions(cur)
				if !ok {
					return false
				}
				switch mName {
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
		if cur.Kind == ast.KindCallExpression && r.isExpectCallWithoutCallback(cur) {
			if !isBareAsyncMatcher && (!hasAsyncModifier || !hasOuterMatcher || !isPureBuiltinAsyncMatcherName(outerMatcherName)) {
				return false
			}
			headArgs := cur.AsCallExpression().Arguments
			return headArgs != nil && len(headArgs.Nodes) == 1 && r.isPurePromiseLiteralArg(headArgs.Nodes[0])
		}
		return false
	}
	return false
}

func (r *AsyncAssertionResolver) arrayLiteralAllPurePromiseExpect(arr *ast.Node) bool {
	if arr == nil || arr.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	if r != nil && r.pureLiteralExpectArrays != nil {
		if cached, ok := r.pureLiteralExpectArrays[arr]; ok {
			return cached
		}
	}
	ok := false
	if a := arr.AsArrayLiteralExpression(); a != nil && a.Elements != nil && len(a.Elements.Nodes) > 0 {
		ok = true
		for _, el := range a.Elements.Nodes {
			if !r.isPureLiteralExpectPromiseCall(el) {
				ok = false
				break
			}
		}
	}
	if r != nil {
		if r.pureLiteralExpectArrays == nil {
			r.pureLiteralExpectArrays = map[*ast.Node]bool{}
		}
		r.pureLiteralExpectArrays[arr] = ok
	}
	return ok
}

// isStandaloneBareAsyncMatcherElem reports whether a bare `toResolve()` /
// `toReject()` call (without `.resolves`/`.rejects`) is safe to treat as
// side-effect-free for sibling autofix checks. Because Jest `setupFilesAfterEnv`
// can supply a custom `toResolve`/`toReject` without any `expect.extend` in the
// current file, only accept zero-arg `expect(Promise.resolve/reject(<primitive>))`
// elements inside an unread first-statement array literal whose elements are
// all pure `expect(Promise.resolve/reject(<primitive>))` calls.
func (r *AsyncAssertionResolver) isStandaloneBareAsyncMatcherElem(call *ast.Node, expectHead *ast.Node) bool {
	if r == nil || r.disallowBuiltinToResolveReject || r.fileMayExtendExpect() {
		return false
	}
	if args := call.AsCallExpression().Arguments; args != nil && len(args.Nodes) > 0 {
		return false
	}
	headArgs := expectHead.AsCallExpression().Arguments
	if headArgs == nil || len(headArgs.Nodes) != 1 || !r.isPurePromiseLiteralArg(headArgs.Nodes[0]) {
		return false
	}
	arr := outermostTransparentExpression(call).Parent
	if arr == nil || arr.Kind != ast.KindArrayLiteralExpression {
		return false
	}
	outerArr := outermostTransparentExpression(arr)
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
		declList.AsVariableDeclarationList().Declarations == nil ||
		len(declList.AsVariableDeclarationList().Declarations.Nodes) != 1 ||
		declList.Parent == nil || declList.Parent.Kind != ast.KindVariableStatement {
		return false
	}
	varStmt := declList.Parent
	block := varStmt.Parent
	if block == nil || block.Kind != ast.KindBlock ||
		block.Parent == nil || !testFramework.IsFunction(block.Parent) ||
		block.AsBlock().Statements == nil || len(block.AsBlock().Statements.Nodes) != 1 ||
		block.AsBlock().Statements.Nodes[0] != varStmt {
		return false
	}
	return r.arrayLiteralAllPurePromiseExpect(arr)
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

func (r *AsyncAssertionResolver) isPureExpectOrPromiseWrapperCall(call *ast.Node) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	if method, ok := r.promiseMethodName(call); ok && (method == "resolve" || method == "reject") {
		return true
	}
	// Only accept side-effect-free built-in async matcher chains
	// (`expect(...).resolves.<builtin>` / `expect(...).rejects.<builtin>` when
	// the file does not extend `expect`), or standalone `toResolve`/`toReject`
	// elements inside a single-statement unused array literal. Custom matchers
	// (even behind `.resolves`/`.rejects`) can resolve deferred gates needed by
	// earlier assertions, so awaiting an earlier element would deadlock or reorder.
	outerCallee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	if outerCallee == nil || !ast.IsAccessExpression(outerCallee) {
		return false
	}
	cur := outerCallee
	hasAsyncModifier := false
	outerMatcherName, hasOuterMatcherName := r.accessStaticNameThroughAssertions(outerCallee)
	isOuterAsyncMatcher := hasOuterMatcherName && (outerMatcherName == "toResolve" || outerMatcherName == "toReject")
	for cur != nil {
		if ast.IsAccessExpression(cur) {
			if cur != outerCallee {
				mName, ok := r.accessStaticNameThroughAssertions(cur)
				if !ok {
					return false
				}
				switch mName {
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
			if r.isExpectCallWithoutCallback(cur) {
				if hasAsyncModifier {
					if r.fileMayExtendExpect() || !hasOuterMatcherName {
						return false
					}
					if isPureBuiltinAsyncMatcherName(outerMatcherName) {
						return true
					}
					return isOuterAsyncMatcher && r.isStandaloneBareAsyncMatcherElem(call, cur)
				}
				return isOuterAsyncMatcher && r.isStandaloneBareAsyncMatcherElem(call, cur)
			}
			if hasAsyncModifier || isOuterAsyncMatcher {
				return false
			}
			cur = internalUtils.SkipAssertionsAndParens(cur.AsCallExpression().Expression)
			continue
		}
		return false
	}
	return false
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
		return node.AsIdentifier().Text == "undefined" && !r.isLocallyValueShadowed(node, "undefined")
	}
	return false
}

func (r *AsyncAssertionResolver) pureExpectOrPromiseCallMayMutate(call *ast.Node, compute func(*ast.Node) bool) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return true
	}
	if method, ok := r.promiseMethodName(call); ok && (method == "resolve" || method == "reject") {
		callee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
		if callee != nil && callee.Kind == ast.KindElementAccessExpression {
			argExpr := callee.AsElementAccessExpression().ArgumentExpression
			if !r.isPurePrimitiveLiteral(argExpr) || compute(argExpr) {
				return true
			}
		}
		if callee != nil && ast.IsAccessExpression(callee) {
			if recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee)); recv != nil &&
				recv.Kind == ast.KindElementAccessExpression {
				recvArg := recv.AsElementAccessExpression().ArgumentExpression
				if !r.isPurePrimitiveLiteral(recvArg) || compute(recvArg) {
					return true
				}
			}
		}
		if args := call.AsCallExpression().Arguments; args != nil {
			for i, arg := range args.Nodes {
				if compute(arg) {
					return true
				}
				// `Promise.resolve(x)` synchronously reads `x.then` on objects.
				if method == "resolve" && i == 0 {
					u := internalUtils.SkipAssertionsAndParens(arg)
					if u != nil && !r.isPurePrimitiveLiteral(u) &&
						u.Kind != ast.KindArrayLiteralExpression &&
						u.Kind != ast.KindArrowFunction && u.Kind != ast.KindFunctionExpression &&
						!r.isPureExpectOrPromiseWrapperCall(u) {
						return true
					}
				}
			}
		}
		return false
	}
	cur := call
	for cur != nil {
		switch cur.Kind {
		case ast.KindCallExpression:
			isExpectHead := r.isExpectCallWithoutCallback(cur)
			if args := cur.AsCallExpression().Arguments; args != nil {
				for _, arg := range args.Nodes {
					if compute(arg) {
						return true
					}
					// `expect(t).resolves...` / `toResolve()` synchronously reads `t.then`
					// and matchers inspect the resolved/rejected payload (which may be a
					// function or object with getters/asymmetricMatch). Only treat
					// primitive literals and `Promise.resolve/reject(<primitive>)` payloads
					// and primitive matcher arguments as guaranteed side-effect-free.
					if isExpectHead {
						if arg != nil && !r.isPurePrimitiveLiteral(arg) && !r.isPurePromiseLiteralArg(arg) {
							return true
						}
					} else if cur == call && arg != nil && !r.isPurePrimitiveLiteral(arg) {
						return true
					}
				}
			}
			cur = internalUtils.SkipAssertionsAndParens(cur.AsCallExpression().Expression)
		case ast.KindPropertyAccessExpression:
			cur = internalUtils.SkipAssertionsAndParens(cur.AsPropertyAccessExpression().Expression)
		case ast.KindElementAccessExpression:
			argExpr := cur.AsElementAccessExpression().ArgumentExpression
			if !r.isPurePrimitiveLiteral(argExpr) || compute(argExpr) {
				return true
			}
			cur = internalUtils.SkipAssertionsAndParens(cur.AsElementAccessExpression().Expression)
		case ast.KindIdentifier:
			return false
		default:
			return true
		}
	}
	return true
}

// exprMayMutateState reports whether synchronously evaluating `root` may mutate
// state or invoke arbitrary user code (calls other than pure `expect(...)` /
// `Promise.resolve` / `Promise.reject` chains, property/element reads that can
// trigger getters or Proxy traps, non-array spreads/iterators, destructuring,
// object coercions, constructors, tagged templates, assignments, updates,
// `delete`, `yield`, `await`, class static blocks, or method computed
// keys/decorators). Results are memoized per node in `r.exprMutates` so nested
// expressions stay O(N) total.
func (r *AsyncAssertionResolver) exprMayMutateState(root *ast.Node) bool {
	if root == nil || ast.IsTypeNode(root) {
		return false
	}
	if r != nil && r.exprMutates != nil {
		if cached, ok := r.exprMutates[root]; ok {
			return cached
		}
	}
	var compute func(node *ast.Node) bool
	compute = func(node *ast.Node) bool {
		if node == nil || ast.IsTypeNode(node) {
			return false
		}
		if r != nil && r.exprMutates != nil {
			if cached, ok := r.exprMutates[node]; ok {
				return cached
			}
		}
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
			return false
		case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			mutates := false
			if mods := node.Modifiers(); mods != nil {
				for _, mod := range mods.Nodes {
					if mod != nil && mod.Kind == ast.KindDecorator && compute(mod) {
						mutates = true
						break
					}
				}
			}
			if !mutates {
				if name := node.Name(); name != nil && name.Kind == ast.KindComputedPropertyName {
					mutates = compute(name)
				}
			}
			if r != nil {
				if r.exprMutates == nil {
					r.exprMutates = map[*ast.Node]bool{}
				}
				r.exprMutates[node] = mutates
			}
			return mutates
		}
		mutates := false
		skipChildren := false
		switch node.Kind {
		case ast.KindDeleteExpression, ast.KindYieldExpression, ast.KindAwaitExpression,
			ast.KindPostfixUnaryExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression,
			ast.KindThrowStatement,
			ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
			ast.KindSpreadAssignment,
			ast.KindObjectBindingPattern, ast.KindArrayBindingPattern,
			ast.KindForOfStatement, ast.KindForInStatement, ast.KindWithStatement,
			ast.KindClassStaticBlockDeclaration, ast.KindHeritageClause,
			ast.KindTemplateExpression:
			mutates = true
		case ast.KindSpreadElement:
			spExpr := internalUtils.SkipAssertionsAndParens(node.AsSpreadElement().Expression)
			if spExpr == nil || (spExpr.Kind != ast.KindArrayLiteralExpression &&
				spExpr.Kind != ast.KindStringLiteral && spExpr.Kind != ast.KindNoSubstitutionTemplateLiteral) {
				mutates = true
			}
		case ast.KindComputedPropertyName:
			if !r.isPurePrimitiveLiteral(node.AsComputedPropertyName().Expression) {
				mutates = true
			}
		case ast.KindPrefixUnaryExpression:
			p := node.AsPrefixUnaryExpression()
			switch p.Operator {
			case ast.KindPlusPlusToken, ast.KindMinusMinusToken:
				mutates = true
			case ast.KindPlusToken, ast.KindMinusToken, ast.KindTildeToken:
				if !r.isPurePrimitiveLiteral(p.Operand) {
					mutates = true
				}
			}
		case ast.KindBinaryExpression:
			if bin := node.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil {
				op := bin.OperatorToken.Kind
				if ast.IsAssignmentOperator(op) {
					mutates = true
				} else {
					switch op {
					case ast.KindCommaToken, ast.KindAmpersandAmpersandToken, ast.KindBarBarToken,
						ast.KindQuestionQuestionToken, ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
					case ast.KindInKeyword, ast.KindInstanceOfKeyword:
						mutates = true
					default:
						if !r.isPurePrimitiveLiteral(bin.Left) || !r.isPurePrimitiveLiteral(bin.Right) {
							mutates = true
						}
					}
				}
			}
		case ast.KindCallExpression:
			if !r.isPureExpectOrPromiseWrapperCall(node) {
				mutates = true
			} else {
				mutates = r.pureExpectOrPromiseCallMayMutate(node, compute)
				skipChildren = true
			}
		}
		if !mutates && !skipChildren {
			node.ForEachChild(func(child *ast.Node) bool {
				if compute(child) {
					mutates = true
					return true
				}
				return false
			})
		}
		if r != nil {
			if r.exprMutates == nil {
				r.exprMutates = map[*ast.Node]bool{}
			}
			r.exprMutates[node] = mutates
		}
		return mutates
	}
	return compute(root)
}

func hasExplicitNonAnyTypeNode(typeNode *ast.Node) bool {
	return typeNode != nil && typeNode.Kind != ast.KindAnyKeyword && typeNode.Kind != ast.KindUnknownKeyword
}

func fnBodyInspectsOrBranchesOnArgs(body *ast.Node) bool {
	if body == nil {
		return false
	}
	unsafe := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || unsafe || ast.IsTypeNode(node) {
			return
		}
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
			ast.KindClassStaticBlockDeclaration:
			return
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
			ast.KindSpreadElement, ast.KindSpreadAssignment,
			ast.KindObjectBindingPattern, ast.KindArrayBindingPattern,
			ast.KindIfStatement, ast.KindConditionalExpression, ast.KindSwitchStatement,
			ast.KindForInStatement, ast.KindForOfStatement, ast.KindWithStatement,
			ast.KindBinaryExpression, ast.KindPrefixUnaryExpression, ast.KindThrowStatement:
			unsafe = true
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return unsafe
		})
	}
	visit(body)
	return unsafe
}

func (r *AsyncAssertionResolver) fnBodyMayMutateState(fn *ast.Node) bool {
	if fn == nil {
		return true
	}
	if r != nil && r.fnMutatesState != nil {
		if cached, ok := r.fnMutatesState[fn]; ok {
			return cached
		}
	}
	mutates := r.exprMayMutateState(fn.Body()) || fnBodyInspectsOrBranchesOnArgs(fn.Body())
	if !mutates {
		for _, p := range fn.Parameters() {
			if p == nil {
				continue
			}
			pd := p.AsParameterDeclaration()
			// Reject destructuring parameters, default initializers, explicit
			// non-any parameter types, or mutating decorators because awaiting
			// changes a Promise argument into `undefined`.
			if pd == nil || p.Name() == nil || p.Name().Kind != ast.KindIdentifier ||
				pd.Initializer != nil || hasExplicitNonAnyTypeNode(pd.Type) || r.exprMayMutateState(p) {
				mutates = true
				break
			}
		}
	}
	if r != nil {
		if r.fnMutatesState == nil {
			r.fnMutatesState = map[*ast.Node]bool{}
		}
		r.fnMutatesState[fn] = mutates
	}
	return mutates
}

func (r *AsyncAssertionResolver) hasSymbolWrite(sym *ast.Symbol) bool {
	if sym == nil || r == nil || r.refs == nil {
		return true
	}
	if written, cached := r.bindingWritten[sym]; cached {
		return written
	}
	written := false
	for _, ref := range r.refs.References(sym) {
		if ref != nil && internalUtils.IsWriteReference(ref) {
			written = true
			break
		}
	}
	if r.bindingWritten == nil {
		r.bindingWritten = map[*ast.Symbol]bool{}
	}
	r.bindingWritten[sym] = written
	return written
}

func (r *AsyncAssertionResolver) isKnownPromiseExpression(expr *ast.Node, depth int) bool {
	if depth > 8 || expr == nil {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindNewExpression:
		return r.isGlobalPromiseObject(expr.AsNewExpression().Expression, 0)
	case ast.KindCallExpression:
		if method, ok := r.promiseMethodName(expr); ok &&
			(method == "resolve" || method == "reject" || method == "all" ||
				method == "allSettled" || method == "race" || method == "any" || method == "try") {
			return true
		}
		callee := internalUtils.SkipAssertionsAndParens(expr.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if m, ok := r.accessStaticNameThroughAssertions(callee); ok &&
				(m == "then" || m == "catch" || m == "finally") {
				return r.isKnownPromiseExpression(internalUtils.AccessExpressionObject(callee), depth+1)
			}
		}
		return false
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		return cond != nil &&
			r.isKnownPromiseExpression(cond.WhenTrue, depth+1) &&
			r.isKnownPromiseExpression(cond.WhenFalse, depth+1)
	}
	return false
}

func hasNonAnyTypeAssertionWrapper(expr *ast.Node) bool {
	cur := expr
	for cur != nil {
		switch cur.Kind {
		case ast.KindParenthesizedExpression:
			cur = cur.AsParenthesizedExpression().Expression
		case ast.KindAsExpression:
			if hasExplicitNonAnyTypeNode(cur.AsAsExpression().Type) {
				return true
			}
			cur = cur.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			if hasExplicitNonAnyTypeNode(cur.AsTypeAssertion().Type) {
				return true
			}
			cur = cur.AsTypeAssertion().Expression
		case ast.KindSatisfiesExpression:
			if hasExplicitNonAnyTypeNode(cur.AsSatisfiesExpression().Type) {
				return true
			}
			cur = cur.AsSatisfiesExpression().Expression
		case ast.KindNonNullExpression:
			cur = cur.AsNonNullExpression().Expression
		default:
			return false
		}
	}
	return false
}

func (r *AsyncAssertionResolver) isPureSimpleCallableExpr(expr *ast.Node, depth int) bool {
	if depth > 4 || expr == nil || hasNonAnyTypeAssertionWrapper(expr) {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return !r.fnBodyMayMutateState(expr)
	case ast.KindClassExpression:
		return !r.classConstructionMayMutateState(expr)
	case ast.KindIdentifier:
		sym := r.resolveSymbol(expr)
		if sym == nil {
			return false
		}
		decl := r.singleSameFileValueDeclaration(sym)
		if decl == nil {
			return false
		}
		switch decl.Kind {
		case ast.KindFunctionDeclaration:
			return !r.hasSymbolWrite(sym) && !r.fnBodyMayMutateState(decl)
		case ast.KindClassDeclaration:
			return !r.hasSymbolWrite(sym) && !r.classConstructionMayMutateState(decl)
		case ast.KindVariableDeclaration:
			vd := decl.AsVariableDeclaration()
			if ast.IsVarConst(decl) && vd != nil && !hasExplicitNonAnyTypeNode(vd.Type) &&
				vd.Initializer != nil && !nodeContains(vd.Initializer, expr) {
				return r.isPureSimpleCallableExpr(vd.Initializer, depth+1)
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) hasObjectBindingMutationOrEscape(objSym *ast.Symbol, decl *ast.Node, initObj *ast.Node) bool {
	if r == nil || r.refs == nil || objSym == nil || decl == nil || initObj == nil ||
		initObj.Kind != ast.KindObjectLiteralExpression || initObj.AsObjectLiteralExpression().Properties == nil {
		return true
	}
	if mutated, cached := r.objMutated[objSym]; cached {
		return mutated
	}
	mutated := isDeclarationExportedOrAmbient(decl)
	var propByName map[string]*ast.Node
	if !mutated {
		props := initObj.AsObjectLiteralExpression().Properties.Nodes
		propByName = make(map[string]*ast.Node, len(props))
		for _, prop := range props {
			if prop == nil {
				continue
			}
			if prop.Kind == ast.KindSpreadAssignment || prop.Kind == ast.KindGetAccessor ||
				prop.Kind == ast.KindSetAccessor || prop.Name() == nil {
				mutated = true
				break
			}
			name, ok := internalUtils.GetStaticPropertyName(prop.Name())
			if !ok {
				mutated = true
				break
			}
			propByName[name] = prop
		}
	}
	if !mutated {
		for _, ref := range r.refs.References(objSym) {
			if ref == nil || ref == decl.Name() {
				continue
			}
			if internalUtils.IsWriteReference(ref) {
				mutated = true
				break
			}
			outer := outermostTransparentExpression(ref)
			if outer == nil || outer.Parent == nil ||
				!ast.IsAccessExpression(outer.Parent) ||
				internalUtils.AccessExpressionObject(outer.Parent) != outer {
				mutated = true
				break
			}
			member := outer.Parent
			memberOuter := outermostTransparentExpression(member)
			if internalUtils.IsWriteReference(member) || internalUtils.IsWriteReference(memberOuter) {
				mutated = true
				break
			}
			if memberOuter != nil && memberOuter.Parent != nil {
				if memberOuter.Parent.Kind == ast.KindDeleteExpression {
					mutated = true
					break
				}
				invoked := false
				switch memberOuter.Parent.Kind {
				case ast.KindCallExpression:
					invoked = memberOuter.Parent.AsCallExpression().Expression == memberOuter
				case ast.KindNewExpression:
					invoked = memberOuter.Parent.AsNewExpression().Expression == memberOuter
				case ast.KindTaggedTemplateExpression:
					invoked = memberOuter.Parent.AsTaggedTemplateExpression().Tag == memberOuter
				}
				if invoked {
					mName, ok := r.accessStaticNameThroughAssertions(member)
					if !ok {
						mutated = true
						break
					}
					prop := propByName[mName]
					if prop == nil {
						mutated = true
						break
					}
					pure := false
					switch prop.Kind {
					case ast.KindMethodDeclaration:
						pure = !r.fnBodyMayMutateState(prop)
					case ast.KindPropertyAssignment:
						pure = r.isPureSimpleCallableExpr(prop.AsPropertyAssignment().Initializer, 0)
					case ast.KindShorthandPropertyAssignment:
						pure = r.isPureSimpleCallableExpr(prop.AsShorthandPropertyAssignment().Name(), 0)
					}
					if !pure {
						mutated = true
						break
					}
				}
			}
		}
	}
	if r.objMutated == nil {
		r.objMutated = map[*ast.Symbol]bool{}
	}
	r.objMutated[objSym] = mutated
	return mutated
}

// calleeMayMutateState fails closed (returns `true`) unless `callee` is proven
// non-mutating: global `Promise.*` (non-`try`/`reject`), `.then`/`.catch`/
// `.finally` Promise chains on known Promise expressions, unresolved ambient
// helpers (preserving upstream Jest tests like `Promise.all([foo(expect(...))])`),
// or same-file `const`/never-reassigned declared pure functions, non-extending
// pure classes, and non-mutated object literal methods/properties whose bodies
// have no state mutations.
func (r *AsyncAssertionResolver) calleeMayMutateState(callee *ast.Node, depth int) bool {
	if depth > 4 || callee == nil || hasNonAnyTypeAssertionWrapper(callee) {
		return true
	}
	callee = internalUtils.SkipAssertionsAndParens(callee)
	if callee == nil {
		return true
	}
	switch callee.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return r.fnBodyMayMutateState(callee)
	case ast.KindClassExpression:
		return r.classConstructionMayMutateState(callee)
	case ast.KindIdentifier:
		sym := r.resolveSymbol(callee)
		if sym == nil {
			// Unresolved ambient identifier (e.g. upstream `foo(expect(...))`).
			return false
		}
		decl := r.singleSameFileValueDeclaration(sym)
		if decl == nil {
			// Imported or multi-declaration symbol may mutate state or expect a Promise.
			return true
		}
		switch decl.Kind {
		case ast.KindFunctionDeclaration:
			return r.hasSymbolWrite(sym) || r.fnBodyMayMutateState(decl)
		case ast.KindClassDeclaration:
			return r.hasSymbolWrite(sym) || r.classConstructionMayMutateState(decl)
		case ast.KindVariableDeclaration:
			vd := decl.AsVariableDeclaration()
			if ast.IsVarConst(decl) && vd != nil && !hasExplicitNonAnyTypeNode(vd.Type) &&
				vd.Initializer != nil && !nodeContains(vd.Initializer, callee) {
				return r.calleeMayMutateState(vd.Initializer, depth+1)
			}
		}
		return true
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		propName, ok := r.accessStaticNameThroughAssertions(callee)
		rawObj := internalUtils.AccessExpressionObject(callee)
		obj := internalUtils.SkipAssertionsAndParens(rawObj)
		if obj == nil {
			return true
		}
		// Global `Promise.*` (excluding `try`/`reject` and dynamic keys)
		// and custom `Promise.all.x(...)` (upstream Jest test).
		if r.isGlobalPromiseObject(obj, 0) {
			return !ok || propName == "try" || propName == "reject"
		}
		if ast.IsAccessExpression(obj) && r.isGlobalPromiseObject(internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(obj)), 0) {
			objMethod, objOK := r.accessStaticNameThroughAssertions(obj)
			return !objOK || objMethod == "try" || objMethod == "reject"
		}
		// Chained `.then`/`.catch`/`.finally` on a known Promise expression.
		if ok && (propName == "then" || propName == "catch" || propName == "finally") &&
			r.isKnownPromiseExpression(obj, 0) {
			return false
		}
		if !ok || hasNonAnyTypeAssertionWrapper(rawObj) {
			return true
		}
		if obj.Kind == ast.KindIdentifier {
			objSym := r.resolveSymbol(obj)
			decl := r.singleSameFileValueDeclaration(objSym)
			if decl == nil || decl.Kind != ast.KindVariableDeclaration || !ast.IsVarConst(decl) ||
				hasExplicitNonAnyTypeNode(decl.AsVariableDeclaration().Type) ||
				hasNonAnyTypeAssertionWrapper(decl.AsVariableDeclaration().Initializer) {
				return true
			}
			initObj := internalUtils.SkipAssertionsAndParens(decl.AsVariableDeclaration().Initializer)
			if r.hasObjectBindingMutationOrEscape(objSym, decl, initObj) {
				return true
			}
			obj = initObj
		}
		if obj != nil && obj.Kind == ast.KindObjectLiteralExpression && obj.AsObjectLiteralExpression().Properties != nil {
			// Last matching property wins in JS; any trailing spread or dynamic
			// key invalidates static resolution.
			var matched *ast.Node
			for _, prop := range obj.AsObjectLiteralExpression().Properties.Nodes {
				if prop == nil {
					continue
				}
				if prop.Kind == ast.KindSpreadAssignment || prop.Name() == nil {
					matched = nil
					continue
				}
				name, nameOK := internalUtils.GetStaticPropertyName(prop.Name())
				if !nameOK {
					if prop.Name().Kind == ast.KindComputedPropertyName {
						matched = nil
					}
					continue
				}
				if name == propName {
					matched = prop
				}
			}
			if matched == nil {
				return true
			}
			switch matched.Kind {
			case ast.KindMethodDeclaration:
				return r.fnBodyMayMutateState(matched)
			case ast.KindPropertyAssignment:
				return r.calleeMayMutateState(matched.AsPropertyAssignment().Initializer, depth+1)
			case ast.KindShorthandPropertyAssignment:
				return r.calleeMayMutateState(matched.AsShorthandPropertyAssignment().Name(), depth+1)
			}
		}
		return true
	}
	return true
}

func (r *AsyncAssertionResolver) classConstructionMayMutateState(cls *ast.Node) bool {
	if cls == nil {
		return true
	}
	if r != nil && r.fnMutatesState != nil {
		if cached, ok := r.fnMutatesState[cls]; ok {
			return cached
		}
	}
	mutates := internalUtils.GetHeritageClauses(cls) != nil
	if !mutates {
		for _, member := range cls.Members() {
			if member == nil {
				continue
			}
			switch member.Kind {
			case ast.KindConstructor:
				if r.fnBodyMayMutateState(member) {
					mutates = true
				}
			case ast.KindPropertyDeclaration:
				if !ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) &&
					r.exprMayMutateState(member.AsPropertyDeclaration().Initializer) {
					mutates = true
				}
			default:
				if ast.IsAutoAccessorPropertyDeclaration(member) &&
					!ast.HasSyntacticModifier(member, ast.ModifierFlagsStatic) &&
					r.exprMayMutateState(member.Initializer()) {
					mutates = true
				}
			}
			if mutates {
				break
			}
		}
	}
	if r != nil {
		if r.fnMutatesState == nil {
			r.fnMutatesState = map[*ast.Node]bool{}
		}
		r.fnMutatesState[cls] = mutates
	}
	return mutates
}

func (r *AsyncAssertionResolver) getCallArgsIndex(call *ast.Node) *callArgsIndex {
	if r != nil && r.callArgIndices != nil {
		if idx, ok := r.callArgIndices[call]; ok {
			return idx
		}
	}
	var nodes []*ast.Node
	switch call.Kind {
	case ast.KindCallExpression:
		if args := call.AsCallExpression().Arguments; args != nil {
			nodes = args.Nodes
		}
	case ast.KindNewExpression:
		if args := call.AsNewExpression().Arguments; args != nil {
			nodes = args.Nodes
		}
	}
	idx := &callArgsIndex{
		argIndex:          make(map[*ast.Node]int, len(nodes)),
		lastSideEffectIdx: -1,
	}
	for i, n := range nodes {
		if n != nil {
			idx.argIndex[n] = i
		}
	}
	for i := len(nodes) - 1; i >= 0; i-- {
		if r.exprMayMutateState(nodes[i]) {
			idx.lastSideEffectIdx = i
			break
		}
	}
	if r != nil {
		if r.callArgIndices == nil {
			r.callArgIndices = map[*ast.Node]*callArgsIndex{}
		}
		r.callArgIndices[call] = idx
	}
	return idx
}

func (r *AsyncAssertionResolver) hasFollowingSideEffectArg(call *ast.Node, cur *ast.Node) bool {
	idx := r.getCallArgsIndex(call)
	pos, ok := idx.argIndex[cur]
	return ok && pos < idx.lastSideEffectIdx
}

func (r *AsyncAssertionResolver) hasFollowingSideEffectElem(arrNode *ast.Node, cur *ast.Node) bool {
	idx := r.getArrayLiteralIndex(arrNode)
	if !idx.lastSideEffectInit {
		idx.lastSideEffectInit = true
		idx.lastSideEffectIdx = -1
		if arr := arrNode.AsArrayLiteralExpression(); arr != nil && arr.Elements != nil {
			nodes := arr.Elements.Nodes
			for i := len(nodes) - 1; i >= 0; i-- {
				if r.exprMayMutateState(nodes[i]) {
					idx.lastSideEffectIdx = i
					break
				}
			}
		}
	}
	pos, ok := idx.elemIndex[cur]
	return ok && pos < idx.lastSideEffectIdx
}

func (r *AsyncAssertionResolver) getNodeListSideEffectIndex(
	container *ast.Node,
	nodes []*ast.Node,
	extractExpr func(*ast.Node) *ast.Node,
) *nodeListSideEffectIndex {
	if r != nil && r.listSideEffectIndices != nil {
		if idx, ok := r.listSideEffectIndices[container]; ok {
			return idx
		}
	}
	idx := &nodeListSideEffectIndex{
		nodeIndex:         make(map[*ast.Node]int, len(nodes)),
		lastSideEffectIdx: -1,
	}
	for i, n := range nodes {
		if n != nil {
			idx.nodeIndex[n] = i
		}
	}
	for i := len(nodes) - 1; i >= 0; i-- {
		if nodes[i] != nil && r.exprMayMutateState(extractExpr(nodes[i])) {
			idx.lastSideEffectIdx = i
			break
		}
	}
	if r != nil {
		if r.listSideEffectIndices == nil {
			r.listSideEffectIndices = map[*ast.Node]*nodeListSideEffectIndex{}
		}
		r.listSideEffectIndices[container] = idx
	}
	return idx
}

func (r *AsyncAssertionResolver) hasFollowingSideEffectDecl(declList *ast.Node, cur *ast.Node) bool {
	decls := declList.AsVariableDeclarationList().Declarations
	if decls == nil || len(decls.Nodes) <= 1 {
		return false
	}
	idx := r.getNodeListSideEffectIndex(declList, decls.Nodes, func(n *ast.Node) *ast.Node {
		return n
	})
	pos, ok := idx.nodeIndex[cur]
	return ok && pos < idx.lastSideEffectIdx
}

func (r *AsyncAssertionResolver) hasFollowingSideEffectObjProp(objNode *ast.Node, cur *ast.Node) bool {
	props := objNode.AsObjectLiteralExpression().Properties
	if props == nil || len(props.Nodes) <= 1 {
		return false
	}
	idx := r.getNodeListSideEffectIndex(objNode, props.Nodes, func(n *ast.Node) *ast.Node {
		return n
	})
	pos, ok := idx.nodeIndex[cur]
	return ok && pos < idx.lastSideEffectIdx
}

func (r *AsyncAssertionResolver) isAssignmentTargetObserved(lhs *ast.Node) bool {
	if lhs == nil {
		return false
	}
	switch lhs.Kind {
	case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return true
	case ast.KindIdentifier:
		return r.isBindingReadOrExported(nil, lhs)
	default:
		return false
	}
}

func isPureWriteReference(ref *ast.Node) bool {
	if !internalUtils.IsWriteReference(ref) {
		return false
	}
	outer := outermostTransparentExpression(ref)
	if outer == nil || outer.Parent == nil {
		return true
	}
	switch outer.Parent.Kind {
	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
		return false
	case ast.KindBinaryExpression:
		bin := outer.Parent.AsBinaryExpression()
		if bin != nil && bin.Left == outer && bin.OperatorToken != nil && bin.OperatorToken.Kind != ast.KindEqualsToken {
			return false
		}
	}
	return true
}

func isDeclarationExportedOrAmbient(decl *ast.Node) bool {
	if decl == nil {
		return false
	}
	if internalUtils.IsInAmbientContext(decl) || ast.HasSyntacticModifier(decl, ast.ModifierFlagsAmbient) {
		return true
	}
	if decl.Kind == ast.KindBindingElement {
		decl = internalUtils.EnclosingVariableDeclarationOfBindingElement(decl)
		if decl == nil {
			return false
		}
	}
	if decl.Kind == ast.KindVariableDeclaration &&
		decl.Parent != nil && decl.Parent.Parent != nil &&
		ast.HasSyntacticModifier(decl.Parent.Parent, ast.ModifierFlagsExport) {
		return true
	}
	return ast.HasSyntacticModifier(decl, ast.ModifierFlagsExport)
}

func (r *AsyncAssertionResolver) isBindingReadOrExported(declNode *ast.Node, ident *ast.Node) bool {
	if ident == nil || ident.Kind != ast.KindIdentifier {
		return false
	}
	if isDeclarationExportedOrAmbient(declNode) {
		return true
	}
	if r == nil || r.refs == nil {
		// Fail closed without RefStore: we cannot prove the binding is unread.
		return true
	}
	var sym *ast.Symbol
	if declNode != nil && declNode.Symbol() != nil {
		sym = declNode.Symbol()
	} else {
		sym = r.resolveSymbol(ident)
	}
	if sym == nil {
		// Undeclared/global binding may be observed outside this file.
		return true
	}
	if observed, cached := r.bindingObserved[sym]; cached {
		return observed
	}
	observed := len(sym.Declarations) == 0
	if !observed {
		for _, d := range sym.Declarations {
			if d == nil {
				continue
			}
			if (r.sourceFile != nil && ast.GetSourceFileOfNode(d) != r.sourceFile) || isDeclarationExportedOrAmbient(d) {
				observed = true
				break
			}
		}
	}
	if !observed {
		for _, ref := range r.refs.References(sym) {
			if ref == nil {
				continue
			}
			if parent := ref.Parent; parent != nil &&
				((parent.Kind == ast.KindVariableDeclaration && parent.Name() == ref) ||
					(parent.Kind == ast.KindBindingElement && parent.Name() == ref)) {
				continue
			}
			if isPureWriteReference(ref) {
				continue
			}
			observed = true
			break
		}
	}
	if r.bindingObserved == nil {
		r.bindingObserved = map[*ast.Symbol]bool{}
	}
	r.bindingObserved[sym] = observed
	return observed
}

func (r *AsyncAssertionResolver) unhandledPromiseRejectDepth(node *ast.Node, depth int) int {
	if depth > 8 || node == nil {
		return 0
	}
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return 0
	}
	switch node.Kind {
	case ast.KindConditionalExpression:
		cond := node.AsConditionalExpression()
		if cond == nil {
			return 0
		}
		return max(
			r.unhandledPromiseRejectDepth(cond.WhenTrue, depth+1),
			r.unhandledPromiseRejectDepth(cond.WhenFalse, depth+1),
		)
	case ast.KindBinaryExpression:
		if internalUtils.IsCommaOperator(node) {
			return r.unhandledPromiseRejectDepth(node.AsBinaryExpression().Right, depth+1)
		}
		return 0
	case ast.KindArrayLiteralExpression:
		arr := node.AsArrayLiteralExpression()
		if arr == nil || arr.Elements == nil {
			return 0
		}
		maxDepth := 0
		for _, el := range arr.Elements.Nodes {
			if el != nil && el.Kind == ast.KindSpreadElement {
				el = el.AsSpreadElement().Expression
			}
			if d := r.unhandledPromiseRejectDepth(el, depth+1); d > maxDepth {
				maxDepth = d
			}
		}
		return maxDepth
	case ast.KindCallExpression:
		if method, ok := r.promiseMethodName(node); ok {
			if method == "reject" {
				inner := 0
				if firstArg, argOK := staticCallArgAt(node, 0); argOK && firstArg != nil {
					inner = r.unhandledPromiseRejectDepth(firstArg, depth+1)
				}
				return 1 + inner
			}
			if method == "allSettled" || method == "any" || method == "withResolvers" {
				return 0
			}
			maxDepth := 0
			if args := node.AsCallExpression().Arguments; args != nil {
				for _, arg := range args.Nodes {
					if arg != nil && arg.Kind == ast.KindSpreadElement {
						arg = arg.AsSpreadElement().Expression
					}
					if d := r.unhandledPromiseRejectDepth(arg, depth+1); d > maxDepth {
						maxDepth = d
					}
				}
			}
			return maxDepth
		}
		callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if m, ok := r.accessStaticNameThroughAssertions(callee); ok {
				recv := internalUtils.AccessExpressionObject(callee)
				switch m {
				case "finally":
					return r.unhandledPromiseRejectDepth(recv, depth+1)
				case "then", "catch":
					recvDepth := r.unhandledPromiseRejectDepth(recv, depth+1)
					if recvDepth == 0 {
						return 0
					}
					argIdx := 0
					if m == "then" {
						argIdx = 1
					}
					onRejected, argOK := staticCallArgAt(node, argIdx)
					if !argOK {
						return max(0, recvDepth-1)
					}
					onRejected = internalUtils.SkipAssertionsAndParens(onRejected)
					if r.isNonCallablePromiseHandler(onRejected) {
						return recvDepth
					}
					if onRejected.Kind == ast.KindArrowFunction || onRejected.Kind == ast.KindFunctionExpression {
						if rejDepth := r.inlineCallbackPropagatesRejection(onRejected, 1); rejDepth > 0 {
							return recvDepth + rejDepth - 1
						}
					}
					return max(0, recvDepth-1)
				}
			}
		}
	}
	return 0
}

func (r *AsyncAssertionResolver) isInsideTryWithCatchOrAbruptFinally(node *ast.Node) bool {
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
				if tryStmt.FinallyBlock != nil && r.statementMayExitEarly(tryStmt.FinallyBlock) {
					return true
				}
			} else if tryStmt.CatchClause == current {
				if tryStmt.FinallyBlock != nil && r.statementMayExitEarly(tryStmt.FinallyBlock) {
					return true
				}
			}
		}
	}
	return false
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
func AwaitFix(sourceFile *ast.SourceFile, node *ast.Node, alwaysAwait bool) rule.RuleFix {
	return (*AsyncAssertionResolver)(nil).AwaitFix(sourceFile, node, alwaysAwait)
}

// AwaitFix is the resolver-backed form of AwaitFix, reusing per-file Promise
// alias and static-string caches when deciding whether a terminal `return` may
// be stripped.
func (r *AsyncAssertionResolver) AwaitFix(sourceFile *ast.SourceFile, node *ast.Node, alwaysAwait bool) rule.RuleFix {
	outer := outermostTransparentExpression(node)
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
	// If the assertion is wrapped in a TypeScript assertion (`as`, `<T>`,
	// `satisfies`, `!`), await the outermost wrapper so `await` binds to the
	// cast Promise rather than splitting `expect(...)` from `as Promise<void>`.
	// Plain parentheses keep inserting before `node` to preserve `(await ...)`.
	if outer != internalUtils.OutermostParenthesizedExpression(node) {
		if outer.Kind == ast.KindAsExpression || outer.Kind == ast.KindSatisfiesExpression {
			return wrapAwaitWithParens(sourceFile, outer)
		}
		return insertAwaitBefore(sourceFile, outer)
	}
	return insertAwaitBefore(sourceFile, node)
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
func ResolveAsyncAssertionReportNode(
	assertionNode *ast.Node,
	alwaysAwait bool,
) (reportNode *ast.Node, promiseWrapped bool, insideAssertionArray bool, shouldReport bool) {
	return (*AsyncAssertionResolver)(nil).ResolveAsyncAssertionReportNode(assertionNode, alwaysAwait)
}

// ResolveAsyncAssertionReportNode is the resolver-backed form of
// ResolveAsyncAssertionReportNode, reusing per-file scope and static-string
// caches.
func (r *AsyncAssertionResolver) ResolveAsyncAssertionReportNode(
	assertionNode *ast.Node,
	alwaysAwait bool,
) (reportNode *ast.Node, promiseWrapped bool, insideAssertionArray bool, shouldReport bool) {
	return r.resolveAsyncAssertionReportNode(assertionNode, alwaysAwait, 0, 0)
}

func (r *AsyncAssertionResolver) enclosingAwaitOfValue(node *ast.Node) *ast.Node {
	outer := outermostChainedBase(node)
	if outer != nil && outer.Parent != nil && outer.Parent.Kind == ast.KindAwaitExpression &&
		outer.Parent.AsAwaitExpression().Expression == outer {
		return outer.Parent
	}
	return nil
}

func (r *AsyncAssertionResolver) hasEnclosingAbruptFinally(start *ast.Node) bool {
	for current := start; current != nil; current = current.Parent {
		if testFramework.IsFunction(current) || current.Kind == ast.KindClassStaticBlockDeclaration {
			return false
		}
		if current.Parent != nil && current.Parent.Kind == ast.KindTryStatement {
			tryStmt := current.Parent.AsTryStatement()
			if tryStmt != nil && (tryStmt.TryBlock == current || tryStmt.CatchClause == current) &&
				tryStmt.FinallyBlock != nil && r.statementMayExitEarly(tryStmt.FinallyBlock) {
				return true
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) findTrackedConsumablesInScope(root *ast.Node, tb *trackedBindings) (awaits []*ast.Node, returns []*ast.Node, hasForAwaitOf bool) {
	if root == nil || tb == nil || tb.isEmpty() {
		return nil, nil, false
	}
	r.forEachTrackedConsumeSite(r.getScopeBodyIndex(root), tb, func(site scopeConsumeSite) bool {
		switch site.kind {
		case scopeConsumeAwait:
			if r.exprCanReturnTrackedPromise(site.expr, tb) {
				awaits = append(awaits, site.node)
			}
		case scopeConsumeForAwait:
			if r.exprCanBeTrackedContainer(site.expr, tb) {
				hasForAwaitOf = true
			}
		case scopeConsumeReturn:
			if r.exprCanReturnTrackedPromise(site.expr, tb) {
				returns = append(returns, site.node)
			}
		}
		return false
	})
	return awaits, returns, hasForAwaitOf
}

// tryCatchDropsRejection reports whether `awaitExpr` (awaiting a
// `Promise.reject(assertion)` whose rejection reason is an assertion Promise
// nested `rejectDepth` levels deep) is caught by an enclosing `try { ... } catch`
// block that swallows the caught reason without awaiting, returning (when
// `!alwaysAwait`), or rethrowing it, or overridden by an abruptly exiting
// `finally` block.
func (r *AsyncAssertionResolver) tryCatchDropsRejection(awaitExpr *ast.Node, alwaysAwait bool, rejectDepth int) bool {
	depth := rejectDepth
	for current := awaitExpr; current != nil; current = current.Parent {
		if testFramework.IsFunction(current) || current.Kind == ast.KindClassStaticBlockDeclaration {
			return false
		}
		if current.Parent != nil && current.Parent.Kind == ast.KindTryStatement {
			tryStmt := current.Parent.AsTryStatement()
			if tryStmt == nil {
				continue
			}
			if tryStmt.CatchClause == current {
				// Throwing from `catch` enters `finally`; an abrupt `finally`
				// overrides the thrown rejection reason.
				if tryStmt.FinallyBlock != nil && r.statementMayExitEarly(tryStmt.FinallyBlock) {
					return true
				}
				continue
			}
			if tryStmt.TryBlock == current {
				if tryStmt.CatchClause != nil {
					cc := tryStmt.CatchClause.AsCatchClause()
					if cc == nil || cc.VariableDeclaration == nil ||
						cc.VariableDeclaration.Name() == nil || cc.VariableDeclaration.Name().Kind != ast.KindIdentifier {
						return true
					}
					tb := &trackedBindings{restOffset: -1}
					r.addTrackedIdent(tb, cc.VariableDeclaration.Name(), false)
					r.propagateLocalAliasesInScope(cc.Block, tb)
					// Check `catch` before `finally`: `await e` runs before
					// `finally` and awaits the caught Promise directly. For
					// `depth > 1`, recurse from the specific `await` node so
					// inner `try/catch`/`finally` inside `cc.Block` and outer
					// enclosing `try` blocks are both checked accurately.
					awaits, returns, hasForAwaitOf := r.findTrackedConsumablesInScope(cc.Block, tb)
					if depth == 1 && (len(awaits) > 0 || hasForAwaitOf) {
						return false
					}
					if depth > 1 {
						for _, aw := range awaits {
							if !r.tryCatchDropsRejection(aw, alwaysAwait, depth-1) {
								return false
							}
						}
					}
					// `return e` in `catch` is overridden if ANY enclosing `try`
					// (including an inner `try` inside `cc.Block` or an outer
					// `try`) has an abruptly exiting `finally` block.
					if !alwaysAwait {
						for _, ret := range returns {
							if !r.hasEnclosingAbruptFinally(ret) {
								return false
							}
						}
					}
					if r.statementAlwaysThrows(cc.Block, tb) {
						if tryStmt.FinallyBlock != nil && r.statementMayExitEarly(tryStmt.FinallyBlock) {
							return true
						}
						continue
					}
					return true
				}
				// No `catch` clause: an abrupt `finally` overrides the rejection.
				if tryStmt.FinallyBlock != nil && r.statementMayExitEarly(tryStmt.FinallyBlock) {
					return true
				}
			}
		}
	}
	return false
}

func (r *AsyncAssertionResolver) resolveAsyncAssertionReportNode(
	assertionNode *ast.Node,
	alwaysAwait bool,
	rejectDepth int,
	depth int,
) (reportNode *ast.Node, promiseWrapped bool, insideAssertionArray bool, shouldReport bool) {
	if assertionNode == nil || depth > 2 {
		return nil, false, false, false
	}

	reportNode = assertionNode
	// Walk outward to a fixed point across `.then`/`.catch`/`.finally` chains
	// and enclosing `Promise.*` wrappers (including `Promise.all([...]).then(...)`
	// and nested `Promise.all([Promise.resolve(...)])`).
	for {
		var chained *ast.Node
		if rejectDepth > 0 {
			var outcome rejectChainOutcome
			var extraRejects int
			chained, outcome, extraRejects = r.walkPromiseRejectChain(reportNode)
			rejectDepth += extraRejects
			if outcome == rejectDropped {
				return reportNode, promiseWrapped, insideAssertionArray, true
			}
			if outcome == rejectRecovered {
				reportNode = chained
				rejectDepth--
				continue
			}
		} else {
			chained = getParentIfPromiseChained(reportNode)
		}
		outerChained := outermostTransparentExpression(chained)
		if promiseCallNode, method := r.findPromiseCallExpressionNode(chained); promiseCallNode != nil {
			if rejectDepth > 0 && (method == "allSettled" || method == "any") {
				// `Promise.allSettled` and `Promise.any` absorb the rejected
				// `Promise.reject(assertion)` without awaiting `assertion`.
				return reportNode, promiseWrapped, insideAssertionArray, true
			}
			if method == "reject" {
				if _, outcome, _ := r.walkPromiseRejectChain(promiseCallNode); outcome == rejectDropped {
					return reportNode, promiseWrapped, insideAssertionArray, true
				}
				rejectDepth++
			}
			reportNode = promiseCallNode
			promiseWrapped = true
			if outerChained != nil && outerChained.Parent != nil && outerChained.Parent.Kind == ast.KindArrayLiteralExpression {
				insideAssertionArray = true
			}
			continue
		}
		if chained != reportNode {
			reportNode = chained
			continue
		}
		break
	}

	// Parentheses and TypeScript assertion wrappers around the assertion are
	// erased at runtime, so the await or return that consumes the Promise may
	// sit above one or more wrappers. reportNode itself stays the report range.
	outermost := outermostTransparentExpression(reportNode)
	if rejectDepth > 0 {
		if awaitExpr := r.enclosingAwaitOfValue(outermost); awaitExpr != nil && r.tryCatchDropsRejection(awaitExpr, alwaysAwait, rejectDepth) {
			return reportNode, promiseWrapped, insideAssertionArray, true
		}
	}
	if outermost.Parent == nil || isAcceptableReturnNode(outermost, outermost.Parent, !alwaysAwait) {
		return reportNode, promiseWrapped, insideAssertionArray, false
	}

	if depth < 2 && r.isBoundToAwaitedConst(outermost, alwaysAwait, rejectDepth, depth) {
		return reportNode, promiseWrapped, insideAssertionArray, false
	}

	return reportNode, promiseWrapped, insideAssertionArray, true
}

func (r *AsyncAssertionResolver) getBlockIndex(block *ast.Node) *blockStatementIndex {
	if r != nil && r.blockIndices != nil {
		if idx, ok := r.blockIndices[block]; ok {
			return idx
		}
	}
	stmts := blockStatements(block)
	stmtIndex := make(map[*ast.Node]int, len(stmts))
	exitPrefix := make([]int, len(stmts)+1)
	for i, s := range stmts {
		stmtIndex[s] = i
		exitPrefix[i+1] = exitPrefix[i]
		if r.statementMayExitEarly(s) {
			exitPrefix[i+1]++
		}
	}
	idx := &blockStatementIndex{
		stmts:      stmts,
		stmtIndex:  stmtIndex,
		exitPrefix: exitPrefix,
	}
	if r != nil {
		if r.blockIndices == nil {
			r.blockIndices = map[*ast.Node]*blockStatementIndex{}
		}
		r.blockIndices[block] = idx
	}
	return idx
}

// isBoundToAwaitedConst reports whether outermost is assigned to an immutable
// `const` binding — directly (`const check = expect(...)`) or via positional
// array destructuring (`const [check] = [expect(...)]` or
// `const [check] = await Promise.resolve([expect(...)] as const)`) — and that
// binding is unconditionally awaited (or returned when !alwaysAwait) later in
// the same straight-line block with no intervening early exit.
func (r *AsyncAssertionResolver) isBoundToAwaitedConst(outermost *ast.Node, alwaysAwait bool, rejectDepth int, depth int) bool {
	bindingIdent, declNode := r.extractConstPromiseBinding(outermost)
	if bindingIdent == nil || declNode == nil {
		return false
	}
	declStmt, declBlock := enclosingStatementAndBlock(declNode)
	if declStmt == nil || declBlock == nil {
		return false
	}
	blockIdx := r.getBlockIndex(declBlock)
	if len(blockIdx.stmts) == 0 {
		return false
	}
	declIdx, ok := blockIdx.stmtIndex[declStmt]
	if !ok {
		return false
	}

	var refs []*ast.Node
	var sym *ast.Symbol
	if declNode.Symbol() != nil {
		sym = declNode.Symbol()
	} else if r != nil && r.typeChecker != nil {
		sym = r.typeChecker.GetSymbolAtLocation(bindingIdent)
	}
	if r != nil && r.refs != nil && sym != nil {
		refs = r.refs.References(sym)
	} else {
		// Fallback when RefStore is unavailable: scan subsequent statements in
		// the same block for matching value references while skipping nested
		// functions and shadowing scopes.
		name := bindingIdent.AsIdentifier().Text
		var collect func(*ast.Node)
		collect = func(n *ast.Node) {
			if n == nil {
				return
			}
			if testFramework.IsFunction(n) || n.Kind == ast.KindClassStaticBlockDeclaration {
				return
			}
			if n.Kind == ast.KindClassExpression {
				if cn := n.Name(); cn != nil && cn.Kind == ast.KindIdentifier && cn.AsIdentifier().Text == name {
					return
				}
			}
			if n != declBlock && ast.IsLocalsContainer(n) {
				if local := ast.GetLocals(n)[name]; local != nil &&
					local.Flags&(ast.SymbolFlagsValue|ast.SymbolFlagsAlias) != 0 &&
					hasValueDeclaration(local) {
					return
				}
			}
			if n.Kind == ast.KindIdentifier && n.AsIdentifier().Text == name && isIdentifierValueReference(n) {
				refs = append(refs, n)
			}
			n.ForEachChild(func(child *ast.Node) bool {
				collect(child)
				return false
			})
		}
		for _, s := range blockIdx.stmts[declIdx+1:] {
			collect(s)
		}
	}

	for _, ref := range refs {
		if ref == nil || ref == bindingIdent || ref.Pos() <= declStmt.End() {
			continue
		}
		refStmt, refBlock, unconditional := readStatementInfoInBlock(ref)
		if !unconditional || refBlock != declBlock || refStmt == nil {
			continue
		}
		refIdx, ok := blockIdx.stmtIndex[refStmt]
		if !ok || refIdx <= declIdx {
			continue
		}
		// Ensure no statement between the declaration and the await/return can
		// exit the block early (e.g. `if (early) return;`, `throw ...`, or `yield`).
		if blockIdx.exitPrefix[refIdx] != blockIdx.exitPrefix[declIdx+1] {
			continue
		}
		// Also reject a `yield` inside `declStmt` at/after `outermost` starts
		// (e.g. `const c = Promise.all([expect(...), yield 1])`, `const c = expect(...), _ = yield 1`),
		// or inside `refStmt` before the awaited/returned wrapper completes
		// (e.g. `use(yield 1, await c)`, `await Promise.all([c, yield 1])`, `await c.then(yield 1)`).
		// Guarded by O(1) `exitPrefix` checks so statements without exits pay zero cost.
		if blockIdx.exitPrefix[declIdx+1] != blockIdx.exitPrefix[declIdx] {
			if hasYieldInRange(declStmt, outermost.Pos(), declStmt.End()) {
				continue
			}
			if declNode.Kind == ast.KindBindingElement {
				if vd := internalUtils.EnclosingVariableDeclarationOfBindingElement(declNode); vd != nil &&
					vd.Name() != nil && hasYieldInRange(vd.Name(), vd.Name().Pos(), vd.Name().End()) {
					continue
				}
			}
		}
		refReportNode, _, _, shouldReport := r.resolveAsyncAssertionReportNode(ref, alwaysAwait, rejectDepth, depth+1)
		if shouldReport {
			continue
		}
		if blockIdx.exitPrefix[refIdx+1] != blockIdx.exitPrefix[refIdx] {
			yieldMaxPos := ref.End()
			if refReportNode != nil {
				if outerRef := outermostTransparentExpression(refReportNode); outerRef != nil && outerRef.End() > yieldMaxPos {
					yieldMaxPos = outerRef.End()
				}
			}
			if hasYieldInRange(refStmt, refStmt.Pos(), yieldMaxPos) {
				continue
			}
		}
		return true
	}
	return false
}

func (r *AsyncAssertionResolver) getArrayLiteralIndex(arrNode *ast.Node) *arrayLiteralIndex {
	if r != nil && r.arrayElemIndices != nil {
		if idx, ok := r.arrayElemIndices[arrNode]; ok {
			return idx
		}
	}
	arr := arrNode.AsArrayLiteralExpression()
	idx := &arrayLiteralIndex{firstSpreadIdx: -1}
	if arr != nil && arr.Elements != nil {
		idx.elemIndex = make(map[*ast.Node]int, len(arr.Elements.Nodes))
		for i, elem := range arr.Elements.Nodes {
			if elem == nil {
				continue
			}
			idx.elemIndex[elem] = i
			if idx.firstSpreadIdx < 0 && elem.Kind == ast.KindSpreadElement {
				idx.firstSpreadIdx = i
			}
		}
	}
	if r != nil {
		if r.arrayElemIndices == nil {
			r.arrayElemIndices = map[*ast.Node]*arrayLiteralIndex{}
		}
		r.arrayElemIndices[arrNode] = idx
	}
	return idx
}

func (r *AsyncAssertionResolver) getArrayBindingFirstRestIdx(bpNode *ast.Node) int {
	if r != nil && r.arrayRestIndices != nil {
		if idx, ok := r.arrayRestIndices[bpNode]; ok {
			return idx
		}
	}
	firstRest := -1
	if bp := bpNode.AsBindingPattern(); bp != nil && bp.Elements != nil {
		for i, el := range bp.Elements.Nodes {
			if el != nil && el.Kind == ast.KindBindingElement && el.AsBindingElement().DotDotDotToken != nil {
				firstRest = i
				break
			}
		}
	}
	if r != nil {
		if r.arrayRestIndices == nil {
			r.arrayRestIndices = map[*ast.Node]int{}
		}
		r.arrayRestIndices[bpNode] = firstRest
	}
	return firstRest
}

func hasYieldInRange(stmt *ast.Node, minPos, maxPos int) bool {
	if stmt == nil || maxPos <= minPos {
		return false
	}
	found := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || found || node.End() <= minPos || node.Pos() >= maxPos {
			return
		}
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindClassStaticBlockDeclaration:
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
		case ast.KindYieldExpression:
			if node.Pos() >= minPos && node.End() <= maxPos {
				found = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return found
		})
	}
	visit(stmt)
	return found
}

func (r *AsyncAssertionResolver) extractConstPromiseBinding(outermost *ast.Node) (ident *ast.Node, declNode *ast.Node) {
	cur := outermost
	for cur != nil && cur.Parent != nil {
		parent := cur.Parent
		if parent.Kind == ast.KindConditionalExpression {
			cond := parent.AsConditionalExpression()
			if cond != nil && (cond.WhenTrue == cur || cond.WhenFalse == cur) {
				cur = outermostTransparentExpression(parent)
				continue
			}
			return nil, nil
		}
		if internalUtils.IsCommaOperator(parent) && parent.AsBinaryExpression().Right == cur {
			cur = outermostTransparentExpression(parent)
			continue
		}
		switch parent.Kind {
		case ast.KindVariableDeclaration:
			vd := parent.AsVariableDeclaration()
			if vd != nil && vd.Initializer == cur && ast.IsVarConst(parent) &&
				vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
				return vd.Name(), parent
			}
			return nil, nil
		case ast.KindArrayLiteralExpression:
			arrIdx := r.getArrayLiteralIndex(parent)
			elemIdx, ok := arrIdx.elemIndex[cur]
			if !ok || elemIdx < 0 || (arrIdx.firstSpreadIdx >= 0 && arrIdx.firstSpreadIdx < elemIdx) {
				return nil, nil
			}
			arrOuter := outermostTransparentExpression(parent)
			if arrOuter == nil || arrOuter.Parent == nil {
				return nil, nil
			}
			// Unwrap `await Promise.resolve([...])`.
			if arrOuter.Parent.Kind == ast.KindCallExpression {
				call := arrOuter.Parent
				args := call.AsCallExpression().Arguments
				if args == nil || len(args.Nodes) == 0 || args.Nodes[0] != arrOuter {
					return nil, nil
				}
				method, ok := r.promiseMethodName(call)
				if !ok || method != "resolve" {
					return nil, nil
				}
				callOuter := outermostTransparentExpression(call)
				if callOuter == nil || callOuter.Parent == nil || callOuter.Parent.Kind != ast.KindAwaitExpression {
					return nil, nil
				}
				arrOuter = outermostTransparentExpression(callOuter.Parent)
				if arrOuter == nil || arrOuter.Parent == nil {
					return nil, nil
				}
			}
			if arrOuter.Parent.Kind != ast.KindVariableDeclaration {
				return nil, nil
			}
			vd := arrOuter.Parent.AsVariableDeclaration()
			if vd == nil || vd.Initializer != arrOuter || !ast.IsVarConst(arrOuter.Parent) ||
				vd.Name() == nil || vd.Name().Kind != ast.KindArrayBindingPattern {
				return nil, nil
			}
			bp := vd.Name().AsBindingPattern()
			if bp == nil || bp.Elements == nil || elemIdx >= len(bp.Elements.Nodes) {
				return nil, nil
			}
			if firstRest := r.getArrayBindingFirstRestIdx(vd.Name()); firstRest >= 0 && firstRest < elemIdx {
				return nil, nil
			}
			target := bp.Elements.Nodes[elemIdx]
			if target == nil || target.Kind != ast.KindBindingElement {
				return nil, nil
			}
			be := target.AsBindingElement()
			if be == nil || be.DotDotDotToken != nil || be.Name() == nil || be.Name().Kind != ast.KindIdentifier {
				return nil, nil
			}
			return be.Name(), target
		default:
			return nil, nil
		}
	}
	return nil, nil
}

func enclosingStatementAndBlock(node *ast.Node) (*ast.Node, *ast.Node) {
	for current := node; current != nil && current.Parent != nil; current = current.Parent {
		switch current.Parent.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock, ast.KindCaseClause, ast.KindDefaultClause:
			return current, current.Parent
		}
	}
	return nil, nil
}

func blockStatements(block *ast.Node) []*ast.Node {
	if block == nil {
		return nil
	}
	switch block.Kind {
	case ast.KindBlock:
		if b := block.AsBlock(); b != nil && b.Statements != nil {
			return b.Statements.Nodes
		}
	case ast.KindSourceFile:
		if sf := block.AsSourceFile(); sf != nil && sf.Statements != nil {
			return sf.Statements.Nodes
		}
	case ast.KindModuleBlock:
		if mb := block.AsModuleBlock(); mb != nil && mb.Statements != nil {
			return mb.Statements.Nodes
		}
	case ast.KindCaseClause, ast.KindDefaultClause:
		if cc := block.AsCaseOrDefaultClause(); cc != nil && cc.Statements != nil {
			return cc.Statements.Nodes
		}
	}
	return nil
}

func readStatementInfoInBlock(node *ast.Node) (stmt *ast.Node, block *ast.Node, unconditional bool) {
	stmt, block = enclosingStatementAndBlock(node)
	if stmt == nil || block == nil {
		return stmt, block, false
	}
	prev := node
	for current := node.Parent; current != nil && current != block; prev, current = current, current.Parent {
		switch current.Kind {
		case ast.KindConditionalExpression, ast.KindIfStatement, ast.KindSwitchStatement,
			ast.KindTryStatement, ast.KindCatchClause, ast.KindClassStaticBlockDeclaration:
			return stmt, block, false
		case ast.KindBindingElement:
			if be := current.AsBindingElement(); be != nil && be.Initializer == prev {
				return stmt, block, false
			}
		case ast.KindShorthandPropertyAssignment:
			if sp := current.AsShorthandPropertyAssignment(); sp != nil && sp.ObjectAssignmentInitializer == prev {
				return stmt, block, false
			}
		case ast.KindBinaryExpression:
			if bin := current.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil {
				switch bin.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
					ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
					if bin.Right == prev {
						return stmt, block, false
					}
				case ast.KindEqualsToken:
					if bin.Right == prev && internalUtils.IsDefaultValueInDestructuringAssignment(current) {
						return stmt, block, false
					}
				}
			}
		case ast.KindCallExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			if ast.IsOptionalChain(current) && current.Expression() != prev {
				return stmt, block, false
			}
		default:
			if ast.IsIterationStatement(current, false) || testFramework.IsFunction(current) {
				return stmt, block, false
			}
		}
	}
	return stmt, block, true
}

func (r *AsyncAssertionResolver) statementMayExitEarly(stmt *ast.Node) bool {
	if stmt == nil {
		return false
	}
	ex := r.computeExitSummary(stmt)
	return ex.unconditionalExit || ex.unlabeledBreak || ex.unlabeledContinue || len(ex.escapedLabels) > 0
}

func (r *AsyncAssertionResolver) computeExitSummary(root *ast.Node) blockExitSummary {
	if root == nil {
		return blockExitSummary{}
	}
	if root.Kind == ast.KindBlock && r != nil && r.blockExits != nil {
		if cached, ok := r.blockExits[root]; ok {
			return cached
		}
	}
	var res blockExitSummary
	var escapedSet map[string]bool
	addLabel := func(name string) {
		if escapedSet == nil {
			escapedSet = map[string]bool{}
		}
		if !escapedSet[name] {
			escapedSet[name] = true
			res.escapedLabels = append(res.escapedLabels, name)
		}
	}
	loopDepth := 0
	switchDepth := 0
	var innerLabels map[string]int
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || res.unconditionalExit {
			return
		}
		if node != root && node.Kind == ast.KindBlock {
			sub := r.computeExitSummary(node)
			if sub.unconditionalExit {
				res.unconditionalExit = true
				return
			}
			if sub.unlabeledBreak && loopDepth == 0 && switchDepth == 0 {
				res.unlabeledBreak = true
			}
			if sub.unlabeledContinue && loopDepth == 0 {
				res.unlabeledContinue = true
			}
			for _, lbl := range sub.escapedLabels {
				if innerLabels[lbl] == 0 {
					addLabel(lbl)
				}
			}
			return
		}
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
			ast.KindClassStaticBlockDeclaration:
			return
		case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			// Decorators and computed property names evaluate in the enclosing
			// scope (e.g. `const o = { [yield 1]() {} }`).
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
		case ast.KindReturnStatement, ast.KindThrowStatement, ast.KindYieldExpression:
			res.unconditionalExit = true
			return
		case ast.KindBreakStatement:
			label := node.AsBreakStatement().Label
			if label != nil {
				if innerLabels[label.Text()] == 0 {
					addLabel(label.Text())
				}
			} else if loopDepth == 0 && switchDepth == 0 {
				res.unlabeledBreak = true
			}
			return
		case ast.KindContinueStatement:
			label := node.AsContinueStatement().Label
			if label != nil {
				if innerLabels[label.Text()] == 0 {
					addLabel(label.Text())
				}
			} else if loopDepth == 0 {
				res.unlabeledContinue = true
			}
			return
		case ast.KindLabeledStatement:
			ls := node.AsLabeledStatement()
			if ls != nil && ls.Label != nil {
				name := ls.Label.Text()
				if innerLabels == nil {
					innerLabels = map[string]int{}
				}
				innerLabels[name]++
				visit(ls.Statement)
				innerLabels[name]--
				return
			}
		case ast.KindSwitchStatement:
			sw := node.AsSwitchStatement()
			if sw != nil {
				visit(sw.Expression)
				switchDepth++
				visit(sw.CaseBlock)
				switchDepth--
				return
			}
		}
		if ast.IsIterationStatement(node, false) {
			loopDepth++
			node.ForEachChild(func(child *ast.Node) bool {
				visit(child)
				return res.unconditionalExit
			})
			loopDepth--
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return res.unconditionalExit
		})
	}
	visit(root)
	if root.Kind == ast.KindBlock && r != nil {
		if r.blockExits == nil {
			r.blockExits = map[*ast.Node]blockExitSummary{}
		}
		r.blockExits[root] = res
	}
	return res
}
