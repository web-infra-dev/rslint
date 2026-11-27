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

type promiseTryCallIndex struct {
	argIndex       map[*ast.Node]int
	firstSpreadIdx int // -1 if no spread argument at index >= 1
}

type inlineCallbackSummary struct {
	usesArguments  bool
	bodyCanConsume bool
}

type arrayLiteralIndex struct {
	elemIndex      map[*ast.Node]int
	firstSpreadIdx int // -1 if no spread element
}

// AsyncAssertionResolver caches per-file shadowing and static-string lookups
// for valid-expect rules so repeated `Promise.all` assertions stay linear-time,
// `const Promise = globalThis.Promise` aliases resolve accurately, and computed
// `Promise[key]` method names fold through `StaticStringEvaluator`.
type AsyncAssertionResolver struct {
	sourceFile       *ast.SourceFile
	typeChecker      *checker.Checker
	refs             *rule.RefStore
	shadowCache      *internalUtils.ShadowCache
	staticStrings    *internalUtils.StaticStringEvaluator
	blockIndices     map[*ast.Node]*blockStatementIndex
	blockExits       map[*ast.Node]blockExitSummary
	bindingObserved  map[*ast.Symbol]bool
	tryCallIndices   map[*ast.Node]*promiseTryCallIndex
	callbackInfos    map[*ast.Node]inlineCallbackSummary
	arrayElemIndices map[*ast.Node]*arrayLiteralIndex
	arrayRestIndices map[*ast.Node]int
}

// NewAsyncAssertionResolver creates a per-file resolver backed by ctx.
func NewAsyncAssertionResolver(ctx rule.RuleContext) *AsyncAssertionResolver {
	return &AsyncAssertionResolver{
		sourceFile:  ctx.SourceFile,
		typeChecker: ctx.TypeChecker,
		refs:        ctx.Refs,
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
	idx := &promiseTryCallIndex{firstSpreadIdx: -1}
	if args != nil {
		idx.argIndex = make(map[*ast.Node]int, len(args.Nodes))
		for i, a := range args.Nodes {
			if a == nil {
				continue
			}
			idx.argIndex[a] = i
			if i >= 1 && idx.firstSpreadIdx < 0 && a.Kind == ast.KindSpreadElement {
				idx.firstSpreadIdx = i
			}
		}
	}
	if r != nil {
		if r.tryCallIndices == nil {
			r.tryCallIndices = map[*ast.Node]*promiseTryCallIndex{}
		}
		r.tryCallIndices[call] = idx
	}
	return idx
}

func (r *AsyncAssertionResolver) getInlineCallbackSummary(cb *ast.Node) inlineCallbackSummary {
	if r != nil && r.callbackInfos != nil {
		if info, ok := r.callbackInfos[cb]; ok {
			return info
		}
	}
	info := inlineCallbackSummary{
		usesArguments:  cb != nil && cb.Kind == ast.KindFunctionExpression && functionBodyUsesArguments(cb),
		bodyCanConsume: inlineCallbackBodyCanConsume(cb),
	}
	if r != nil {
		if r.callbackInfos == nil {
			r.callbackInfos = map[*ast.Node]inlineCallbackSummary{}
		}
		r.callbackInfos[cb] = info
	}
	return info
}

// inlineCallbackAcceptsArg reports whether an inline function literal (`cb`)
// can read argument `argIdx` (1-based for `Promise.try(cb, arg1, ...)`).
func (r *AsyncAssertionResolver) inlineCallbackAcceptsArg(cb *ast.Node, argIdx int) bool {
	if cb == nil || argIdx < 1 {
		return false
	}
	params := runtimeParameters(cb)
	for i, param := range params {
		if param == nil {
			continue
		}
		if p := param.AsParameterDeclaration(); p != nil && p.DotDotDotToken != nil && i < argIdx {
			return true
		}
	}
	if len(params) >= argIdx {
		return true
	}
	return r.getInlineCallbackSummary(cb).usesArguments
}

func (r *AsyncAssertionResolver) inlineCallbackAcceptsAnyArg(cb *ast.Node) bool {
	if cb == nil {
		return false
	}
	if len(runtimeParameters(cb)) > 0 {
		return true
	}
	return r.getInlineCallbackSummary(cb).usesArguments
}

func (r *AsyncAssertionResolver) inlineCallbackAcceptsArbitraryTrailingArg(cb *ast.Node) bool {
	if cb == nil {
		return false
	}
	for _, param := range runtimeParameters(cb) {
		if param == nil {
			continue
		}
		if p := param.AsParameterDeclaration(); p != nil && p.DotDotDotToken != nil {
			return true
		}
	}
	return r.getInlineCallbackSummary(cb).usesArguments
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

// inlineCallbackCanConsumePromiseArg reports whether an inline `Promise.try`
// callback can actually await or return the forwarded argument at `argIdx`.
// Callbacks that ignore the argument (e.g. `() => {}`), generator callbacks,
// or block bodies with neither `return <expr>` nor `await`/`for await` cannot
// keep the assertion pending.
func (r *AsyncAssertionResolver) inlineCallbackCanConsumePromiseArg(cb *ast.Node, argIdx int) bool {
	if !r.inlineCallbackAcceptsArg(cb, argIdx) {
		return false
	}
	return r.getInlineCallbackSummary(cb).bodyCanConsume
}

func (r *AsyncAssertionResolver) inlineCallbackCanConsumeSpreadArg(cb *ast.Node) bool {
	if !r.inlineCallbackAcceptsArbitraryTrailingArg(cb) {
		return false
	}
	return r.getInlineCallbackSummary(cb).bodyCanConsume
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

func (r *AsyncAssertionResolver) getPromiseCallExpressionNode(node *ast.Node, directArg *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}

	if node.Kind == ast.KindArrayLiteralExpression {
		outer := outermostTransparentExpression(node)
		if outer == nil || outer.Parent == nil || outer.Parent.Kind != ast.KindCallExpression {
			return nil
		}
		call := outer.Parent
		args := call.AsCallExpression().Arguments
		if args == nil || len(args.Nodes) == 0 || args.Nodes[0] != outer {
			return nil
		}
		method, ok := r.promiseMethodName(call)
		// `Promise.resolve([...])`, `Promise.reject([...])`,
		// `Promise.withResolvers([...])`, and `Promise.try([...])` do not await
		// array elements, unlike `Promise.all` / `allSettled` / `race` / `any`
		// (or custom combinators such as `Promise.x`).
		if !ok || method == "resolve" || method == "reject" || method == "withResolvers" || method == "try" {
			return nil
		}
		return call
	}

	method, ok := r.promiseMethodName(node)
	if !ok || method == "withResolvers" {
		return nil
	}
	if directArg != nil {
		args := node.AsCallExpression().Arguments
		switch method {
		case "resolve", "reject", "all", "allSettled", "race", "any":
			if args == nil || len(args.Nodes) == 0 || args.Nodes[0] != directArg {
				return nil
			}
		case "try":
			if args == nil || len(args.Nodes) < 2 {
				return nil
			}
			tryIdx := r.getPromiseTryCallIndex(node)
			argIdx, ok := tryIdx.argIndex[directArg]
			if !ok || argIdx < 1 {
				return nil
			}
			hasPrecedingSpread := tryIdx.firstSpreadIdx >= 1 && tryIdx.firstSpreadIdx < argIdx
			cb := internalUtils.SkipAssertionsAndParens(args.Nodes[0])
			if cb != nil && (cb.Kind == ast.KindArrowFunction || cb.Kind == ast.KindFunctionExpression) {
				if hasPrecedingSpread {
					if !r.inlineCallbackCanConsumeSpreadArg(cb) {
						return nil
					}
				} else if !r.inlineCallbackCanConsumePromiseArg(cb, argIdx) {
					return nil
				}
			}
		}
	}
	return node
}

func (r *AsyncAssertionResolver) findPromiseCallExpressionNode(node *ast.Node) *ast.Node {
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
			return nil
		}
		if internalUtils.IsCommaOperator(parent) && parent.AsBinaryExpression().Right == current {
			current = outermostTransparentExpression(parent)
			continue
		}
		if parent.Kind == ast.KindCallExpression {
			if parent.AsCallExpression().Expression == current {
				return nil
			}
			return r.getPromiseCallExpressionNode(parent, current)
		}
		if parent.Kind == ast.KindArrayLiteralExpression {
			return r.getPromiseCallExpressionNode(parent, nil)
		}
		return nil
	}
	return nil
}

func getParentIfPromiseChained(node *ast.Node) *ast.Node {
	outer := outermostTransparentExpression(node)
	for outer != nil && outer.Parent != nil {
		parent := outer.Parent
		if parent.Kind == ast.KindConditionalExpression {
			cond := parent.AsConditionalExpression()
			if cond != nil && (cond.WhenTrue == outer || cond.WhenFalse == outer) {
				outer = outermostTransparentExpression(parent)
				continue
			}
			return node
		}
		if internalUtils.IsCommaOperator(parent) && parent.AsBinaryExpression().Right == outer {
			outer = outermostTransparentExpression(parent)
			continue
		}
		break
	}
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
		return true
	}

	// Walk upward through value-propagating expressions to detect member/call/
	// syntax-forbidden positions, destructuring, comparisons, truthiness tests,
	// switch discriminants, or short-circuiting operators where turning a
	// Promise into its resolved value (`undefined`) breaks syntax or semantics.
	cur := outer
	for cur != nil && cur.Parent != nil {
		parent := cur.Parent
		switch parent.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			if internalUtils.AccessExpressionObject(parent) == cur {
				return false
			}
			return true
		case ast.KindCallExpression:
			if parent.AsCallExpression().Expression == cur {
				return false
			}
			// If `Promise.try(cb, expect(...))` has an inline callback that
			// accepts this argument (e.g. `(check) => { check.then(...); }`
			// without `return`/`await`), awaiting `expect(...)` would turn the
			// parameter into `undefined` and crash the callback.
			if r.promiseTryCallbackAcceptsArg(parent, cur) {
				return false
			}
			return true
		case ast.KindNewExpression:
			if parent.AsNewExpression().Expression == cur {
				return false
			}
			return true
		case ast.KindTaggedTemplateExpression:
			if parent.AsTaggedTemplateExpression().Tag == cur {
				return false
			}
			return true
		case ast.KindExpressionWithTypeArguments, ast.KindDecorator,
			ast.KindPostfixUnaryExpression, ast.KindTypeOfExpression:
			return false
		case ast.KindVariableDeclaration:
			vd := parent.AsVariableDeclaration()
			if vd != nil && vd.Initializer == cur && vd.Name() != nil {
				switch vd.Name().Kind {
				case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
					return false
				case ast.KindIdentifier:
					if r.isBindingReadOrExported(parent, vd.Name()) {
						return false
					}
				}
			}
			return true
		case ast.KindConditionalExpression:
			cond := parent.AsConditionalExpression()
			if cond == nil {
				return true
			}
			if cond.Condition == cur {
				return false
			}
			if cond.WhenTrue == cur || cond.WhenFalse == cur {
				cur = outermostTransparentExpression(parent)
				continue
			}
			return true
		case ast.KindIfStatement:
			if stmt := parent.AsIfStatement(); stmt != nil && stmt.Expression == cur {
				return false
			}
			return true
		case ast.KindWhileStatement:
			if stmt := parent.AsWhileStatement(); stmt != nil && stmt.Expression == cur {
				return false
			}
			return true
		case ast.KindDoStatement:
			if stmt := parent.AsDoStatement(); stmt != nil && stmt.Expression == cur {
				return false
			}
			return true
		case ast.KindForStatement:
			if stmt := parent.AsForStatement(); stmt != nil && stmt.Condition == cur {
				return false
			}
			return true
		case ast.KindSwitchStatement:
			if stmt := parent.AsSwitchStatement(); stmt != nil && stmt.Expression == cur {
				return false
			}
			return true
		case ast.KindCaseClause:
			if clause := parent.AsCaseOrDefaultClause(); clause != nil && clause.Expression == cur {
				return false
			}
			return true
		case ast.KindPrefixUnaryExpression:
			if unary := parent.AsPrefixUnaryExpression(); unary != nil && unary.Operator == ast.KindExclamationToken && unary.Operand == cur {
				return false
			}
			return true
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil {
				return true
			}
			switch bin.OperatorToken.Kind {
			case ast.KindAsteriskAsteriskToken:
				if bin.Left == cur {
					return false
				}
			case ast.KindInKeyword, ast.KindInstanceOfKeyword,
				ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
				ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken,
				ast.KindLessThanToken, ast.KindGreaterThanToken,
				ast.KindLessThanEqualsToken, ast.KindGreaterThanEqualsToken:
				return false
			case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				if bin.Left == cur {
					return false
				}
				if bin.Right == cur {
					cur = outermostTransparentExpression(parent)
					continue
				}
			case ast.KindEqualsToken:
				if bin.Right == cur {
					lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
					if r.isAssignmentTargetObserved(lhs) {
						return false
					}
					cur = outermostTransparentExpression(parent)
					continue
				}
			case ast.KindCommaToken:
				if bin.Right == cur {
					cur = outermostTransparentExpression(parent)
					continue
				}
			case ast.KindQuestionQuestionEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
				if bin.Right == cur {
					lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
					if r.isAssignmentTargetObserved(lhs) {
						return false
					}
					cur = outermostTransparentExpression(parent)
					continue
				}
			}
			return true
		case ast.KindArrayLiteralExpression:
			arrOuter := outermostTransparentExpression(parent)
			if arrOuter != nil && arrOuter.Parent != nil {
				switch arrOuter.Parent.Kind {
				case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
					if internalUtils.AccessExpressionObject(arrOuter.Parent) == arrOuter {
						return false
					}
				case ast.KindVariableDeclaration:
					vd := arrOuter.Parent.AsVariableDeclaration()
					if vd != nil && vd.Initializer == arrOuter && vd.Name() != nil {
						switch vd.Name().Kind {
						case ast.KindArrayBindingPattern, ast.KindObjectBindingPattern:
							return false
						case ast.KindIdentifier:
							if r.isBindingReadOrExported(arrOuter.Parent, vd.Name()) {
								return false
							}
						}
					}
				case ast.KindBinaryExpression:
					bin := arrOuter.Parent.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.Right == arrOuter {
						switch bin.OperatorToken.Kind {
						case ast.KindEqualsToken, ast.KindQuestionQuestionEqualsToken,
							ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
							lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
							if r.isAssignmentTargetObserved(lhs) {
								return false
							}
							cur = outermostTransparentExpression(arrOuter.Parent)
							continue
						case ast.KindCommaToken, ast.KindAmpersandAmpersandToken,
							ast.KindBarBarToken, ast.KindQuestionQuestionToken:
							cur = outermostTransparentExpression(arrOuter.Parent)
							continue
						}
					}
				case ast.KindConditionalExpression:
					cond := arrOuter.Parent.AsConditionalExpression()
					if cond != nil && (cond.WhenTrue == arrOuter || cond.WhenFalse == arrOuter) {
						cur = outermostTransparentExpression(arrOuter.Parent)
						continue
					}
				case ast.KindCallExpression:
					call := arrOuter.Parent
					if r.promiseTryCallbackAcceptsArg(call, arrOuter) {
						return false
					}
					if method, ok := r.promiseMethodName(call); ok && (method == "resolve" || method == "reject") {
						callOuter := outermostTransparentExpression(call)
						if callOuter != nil && callOuter.Parent != nil && callOuter.Parent.Kind == ast.KindAwaitExpression {
							callOuter = outermostTransparentExpression(callOuter.Parent)
						}
						if callOuter == nil || callOuter.Parent == nil || callOuter.Parent.Kind != ast.KindExpressionStatement {
							return false
						}
						if method == "reject" && isInsideTryWithCatch(callOuter.Parent) {
							return false
						}
					}
				}
			}
			return true
		default:
			return true
		}
	}
	return true
}

func (r *AsyncAssertionResolver) promiseTryCallbackAcceptsArg(call *ast.Node, arg *ast.Node) bool {
	method, ok := r.promiseMethodName(call)
	if !ok || method != "try" {
		return false
	}
	args := call.AsCallExpression().Arguments
	if args == nil || len(args.Nodes) < 2 {
		return false
	}
	tryIdx := r.getPromiseTryCallIndex(call)
	argIdx, ok := tryIdx.argIndex[arg]
	if !ok || argIdx < 1 {
		return false
	}
	cb := internalUtils.SkipAssertionsAndParens(args.Nodes[0])
	if cb == nil || (cb.Kind != ast.KindArrowFunction && cb.Kind != ast.KindFunctionExpression) {
		return false
	}
	if tryIdx.firstSpreadIdx >= 1 && tryIdx.firstSpreadIdx < argIdx {
		return r.inlineCallbackAcceptsAnyArg(cb)
	}
	return r.inlineCallbackAcceptsArg(cb, argIdx)
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

func isInsideTryWithCatch(node *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if testFramework.IsFunction(current) {
			return false
		}
		if current.Parent != nil && current.Parent.Kind == ast.KindTryStatement {
			tryStmt := current.Parent.AsTryStatement()
			if tryStmt != nil && tryStmt.TryBlock == current && tryStmt.CatchClause != nil {
				return true
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
	return r.resolveAsyncAssertionReportNode(assertionNode, alwaysAwait, 0)
}

func (r *AsyncAssertionResolver) resolveAsyncAssertionReportNode(
	assertionNode *ast.Node,
	alwaysAwait bool,
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
		chained := getParentIfPromiseChained(reportNode)
		outerChained := outermostTransparentExpression(chained)
		if promiseCallNode := r.findPromiseCallExpressionNode(chained); promiseCallNode != nil {
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
	if outermost.Parent == nil || isAcceptableReturnNode(outermost, outermost.Parent, !alwaysAwait) {
		return reportNode, promiseWrapped, insideAssertionArray, false
	}

	if depth < 2 && r.isBoundToAwaitedConst(outermost, alwaysAwait, depth) {
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
func (r *AsyncAssertionResolver) isBoundToAwaitedConst(outermost *ast.Node, alwaysAwait bool, depth int) bool {
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
		if blockIdx.exitPrefix[declIdx+1] != blockIdx.exitPrefix[declIdx] &&
			hasYieldInRange(declStmt, outermost.Pos(), declStmt.End()) {
			continue
		}
		refReportNode, _, _, shouldReport := r.resolveAsyncAssertionReportNode(ref, alwaysAwait, depth+1)
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
