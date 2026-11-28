package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type RstestCallAnalysis struct {
	ctx         rule.RuleContext
	candidates  map[string]rstestCandidateKind
	fnCalls     map[*ast.Node]*ParsedRstestFnCall
	expectCalls map[*ast.Node]*ParsedRstestExpectCall
	isExpect    map[*ast.Node]bool
	expectRoots map[*ast.Symbol]rstestExpectRoot
	// isAssert and assertRoots mirror isExpect and expectRoots for Rstest's
	// Chai `assert` global, which rules that treat it as an assertion must
	// resolve through the same bindings as `expect`.
	isAssert    map[*ast.Node]bool
	assertRoots map[*ast.Symbol]rstestExpectRootKind
	calls       []*ast.Node
	// globalExpectWritten records whether the file assigns to the unresolved
	// global expect binding. Source-only programs cannot resolve that binding
	// to a symbol, so the symbol-keyed write check in classifyRstestExpectRoot
	// cannot protect it. This fact is collected during the analysis's existing
	// file walk so every expect consumer shares one linear-time check.
	globalExpectWritten bool
	// globalAssertWritten is the same fact for the unresolved global assert.
	globalAssertWritten bool
	// Expect customization is computed lazily because only rules whose safety
	// depends on built-in matcher implementations need it.
	overriddenExpectMatchers    map[string]bool
	allExpectMatchersOverridden bool
	hasCustomEqualityTesters    bool
	expectCustomizationOK       bool
	expectConfigAliases         map[*ast.Symbol]expectConfigKind
	expectConfigExprs           map[*ast.Node]expectConfigKind
	sccIndex                    int
	sccIndices                  map[*ast.Symbol]int
	sccLowLink                  map[*ast.Symbol]int
	sccStack                    []*ast.Symbol
	sccOnStack                  map[*ast.Symbol]bool
	sccCurrent                  *ast.Symbol
	sccHitActive                bool
	sccQueueActive              bool
	sccQueueRecording           bool
	sccQueueEdge                sccRevEdge
	sccQueueMembers             map[*ast.Symbol]bool
	sccQueueRevEdges            map[*ast.Symbol][]sccRevEdge
	sccQueueRevEdgeSeen         map[*ast.Symbol]map[sccRevEdge]bool
	safeMatcherChecks           []safeExpectMatcherCheck
	shallowCopyKilledProps      map[*ast.Symbol]map[string]bool
	stmtIndices                 map[*ast.Node]map[*ast.Node]int
	staticStrings               *internalUtils.StaticStringEvaluator
	hasImportMetaRstestWrites   bool
	// functions indexes named function declarations by name so a callback
	// passed by an identifier the checker could not resolve still has a
	// candidate. The index is file-wide and carries no scope information, so
	// entries record whether the name was declared more than once; see
	// walkRstestCallbackRegistrations for the guards that keeps it honest.
	functions map[string]rstestFunctionEntry
	// callbackInfos memoizes the overload-aware callback resolution, which
	// costs a symbol lookup per registration and is asked for by both the test
	// context collector and the callback ownership index.
	callbackInfos map[*ast.Node]rstestCallbackInfo
	// callbackBindings memoizes the function a same-file callback binding
	// denotes. Negative results are cached too: checking whether a binding is
	// written requires asking the per-file reference index for every use.
	callbackBindings        map[*ast.Symbol]rstestCallbackInfo
	callbacks               RstestTestCallbacks
	callbacksOK             bool
	ownership               map[*ast.Node][]rstestCallbackRegistration
	ownershipOK             bool
	registrationCallbacks   map[*ast.Node]bool
	registrationCallbacksOK bool
	hookCallbacks           map[*ast.Node]bool
	hookCallbacksOK         bool
	hasTests                bool
}

type rstestCallAnalysisFileCacheKey struct{}

// GetRstestCallAnalysis returns the analysis shared by every Rstest rule that
// lints the same file. Manually-constructed contexts without a file cache keep
// the standalone behavior used by rule and parser tests.
func GetRstestCallAnalysis(ctx rule.RuleContext) *RstestCallAnalysis {
	// Refs travels with SourceFile and TypeChecker because it is file-scoped
	// like both: the parser uses it to resolve a registration's root binding
	// when no checker is available (see resolveRstestRootSymbol).
	analysisCtx := rule.RuleContext{
		SourceFile:  ctx.SourceFile,
		TypeChecker: ctx.TypeChecker,
		Refs:        ctx.Refs,
	}
	return rule.CachedByFile(
		ctx,
		rstestCallAnalysisFileCacheKey{},
		func() *RstestCallAnalysis { return newRstestCallAnalysis(analysisCtx) },
	)
}

func newRstestCallAnalysis(ctx rule.RuleContext) *RstestCallAnalysis {
	analysis := &RstestCallAnalysis{
		ctx:              ctx,
		candidates:       cloneRstestCandidateSeeds(),
		fnCalls:          map[*ast.Node]*ParsedRstestFnCall{},
		expectCalls:      map[*ast.Node]*ParsedRstestExpectCall{},
		isExpect:         map[*ast.Node]bool{},
		expectRoots:      map[*ast.Symbol]rstestExpectRoot{},
		isAssert:         map[*ast.Node]bool{},
		assertRoots:      map[*ast.Symbol]rstestExpectRootKind{},
		functions:        map[string]rstestFunctionEntry{},
		callbackInfos:    map[*ast.Node]rstestCallbackInfo{},
		callbackBindings: map[*ast.Symbol]rstestCallbackInfo{},
	}
	analysis.indexSourceFile()
	return analysis
}

// ParseFnCall parses any Rstest registration, reusing a cached result and
// computing misses instead of depending on a previous collector traversal.
func (analysis *RstestCallAnalysis) ParseFnCall(node *ast.Node) *ParsedRstestFnCall {
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil
	}
	if !analysis.isFnCallCandidate(node) {
		return nil
	}
	return analysis.parseFnCallCandidate(node)
}

func (analysis *RstestCallAnalysis) parseFnCallCandidate(
	node *ast.Node,
) *ParsedRstestFnCall {
	if parsed, ok := analysis.fnCalls[node]; ok {
		return parsed
	}
	parsed := parseRstestFnCall(node, analysis.ctx)
	analysis.fnCalls[node] = parsed
	return parsed
}

// ParseTestCall is the test-only fast path used by callback collectors.
func (analysis *RstestCallAnalysis) ParseTestCall(node *ast.Node) *ParsedRstestFnCall {
	if !analysis.isTestCandidate(node) {
		return nil
	}
	parsed := analysis.parseFnCallCandidate(node)
	if parsed == nil || parsed.Kind != RstestFnTypeTest {
		return nil
	}
	return parsed
}

// TestCallback returns the callback associated with a final Rstest test
// registration. It reuses the analysis parser and the overload-aware callback
// resolver without scanning the source file or retaining rule-specific state.
func (analysis *RstestCallAnalysis) TestCallback(node *ast.Node) (*ast.Node, string) {
	if analysis.ParseTestCall(node) == nil {
		return nil, ""
	}
	info := analysis.callbackInfo(node)
	return info.functionNode, info.name
}

// callbackInfo resolves the callback argument of a registration call once per
// file. Both callback collectors ask about the same calls, so resolving twice
// would double the declaration lookups they perform.
func (analysis *RstestCallAnalysis) callbackInfo(node *ast.Node) rstestCallbackInfo {
	if info, ok := analysis.callbackInfos[node]; ok {
		return info
	}
	info := resolveRstestTestCallback(analysis, node.AsCallExpression())
	analysis.callbackInfos[node] = info
	return info
}

// parseRegistrationCall keeps the registrations that own a callback body,
// which are the only ones an execution mode can be inherited through.
func (analysis *RstestCallAnalysis) parseRegistrationCall(
	node *ast.Node,
) *ParsedRstestFnCall {
	parsed := analysis.ParseFnCall(node)
	if parsed == nil ||
		(parsed.Kind != RstestFnTypeTest && parsed.Kind != RstestFnTypeDescribe) {
		return nil
	}
	return parsed
}

func (analysis *RstestCallAnalysis) IsExpectCall(node *ast.Node) bool {
	// Test-context expect names are discovered while collecting callbacks, so
	// callbacks must be complete before the candidate gate runs. Reordering
	// these calls would silently reject ctx.expect and destructured aliases.
	analysis.Callbacks()
	if !analysis.isExpectCandidate(node) {
		return false
	}
	if result, ok := analysis.isExpect[node]; ok {
		return result
	}
	result := isRstestExpectCall(node, analysis)
	analysis.isExpect[node] = result
	return result
}

// ParseExpectCall keeps a separate full-parse cache from the cheap identity
// cache used by IsExpectCall. The full parse resolves matcher and modifier
// chains; the identity path deliberately avoids those allocations.
func (analysis *RstestCallAnalysis) ParseExpectCall(
	node *ast.Node,
) *ParsedRstestExpectCall {
	// See IsExpectCall: callback collection widens the expect candidate set.
	analysis.Callbacks()
	if !analysis.isExpectCandidate(node) {
		return nil
	}
	if parsed, ok := analysis.expectCalls[node]; ok {
		return parsed
	}
	parsed := parseRstestExpectCall(node, analysis)
	analysis.expectCalls[node] = parsed
	analysis.isExpect[node] = parsed != nil
	return parsed
}

// ParseExpectCallThroughTransparentExpressions parses an expect chain through
// TypeScript expression wrappers that preserve its JavaScript runtime value.
// It is opt-in so existing parser consumers retain their established wrapper
// boundaries.
func (analysis *RstestCallAnalysis) ParseExpectCallThroughTransparentExpressions(node *ast.Node) *ParsedRstestExpectCall {
	analysis.Callbacks()
	if node == nil || node.Kind != ast.KindCallExpression {
		return nil
	}
	root := testFramework.ResolveFirstIdentifierThroughTransparentExpressions(node.AsCallExpression().Expression)
	if root != nil && root.Kind == ast.KindIdentifier && analysis.candidates[root.AsIdentifier().Text]&rstestCandidateExpect == 0 {
		return nil
	}
	return parseRstestExpectCallThroughTransparentExpressions(node, analysis)
}

// IsExpectMatcherOverridden reports whether an Rstest expect.extend call in
// this file may replace matcher. Dynamic matcher objects, spreads, and escaped
// references to expect or expect.extend conservatively answer true so rules
// never emit unsafe rewrites or skip overridden matchers.
func (analysis *RstestCallAnalysis) IsExpectMatcherOverridden(matcher string) bool {
	analysis.collectExpectCustomization()
	return analysis.allExpectMatchersOverridden || analysis.overriddenExpectMatchers[matcher]
}

// HasCustomEqualityTesters reports whether this file may install Rstest
// equality testers, which may execute user code from equality-based built-in
// matchers.
func (analysis *RstestCallAnalysis) HasCustomEqualityTesters() bool {
	analysis.collectExpectCustomization()
	return analysis.hasCustomEqualityTesters
}

func (analysis *RstestCallAnalysis) staticStringEvaluator() *internalUtils.StaticStringEvaluator {
	if analysis.staticStrings == nil {
		// Avoid passing a typed-nil `*rule.RefStore` into the interface
		// parameter, which would disable StaticStringEvaluator's nil-resolver
		// fallbacks when ctx.Refs is nil.
		var refResolver internalUtils.StaticReferenceResolver
		if analysis.ctx.Refs != nil {
			refResolver = analysis.ctx.Refs
		}
		analysis.staticStrings = internalUtils.NewStaticStringEvaluatorWithCallEscapes(
			analysis.ctx.TypeChecker,
			analysis.ctx.SourceFile,
			refResolver,
		)
	}
	return analysis.staticStrings
}

type expectConfigKind uint16

type safeExpectMatcherCheck struct {
	matcherName string
	argKind     expectConfigKind
}

type sccRevEdge struct {
	dst         *ast.Symbol
	expr        *ast.Node
	bindingElem *ast.Node
}

const (
	expectConfigNone expectConfigKind = 0
	// expectConfigGlobalObject denotes an unshadowed global object
	// (`globalThis`, `window`, `global`, `self`). It is non-escaping on its own
	// and only resolves `.expect` / `{ expect }`.
	expectConfigGlobalObject expectConfigKind = 1 << iota
	// expectConfigImportMeta denotes `import.meta`. It is non-escaping on its
	// own and only resolves `.rstest` / `{ rstest }`.
	expectConfigImportMeta
	// expectConfigReceiver denotes an Rstest namespace/receiver
	// (`import * as r from '@rstest/core'`, `require('@rstest/core')`,
	// `import.meta.rstest`, `await import('@rstest/core')`, or a test callback
	// context parameter `ctx`).
	expectConfigReceiver
	// expectConfigReceiverArray denotes a rest parameter capturing the test
	// callback context at index 0 (`test('t', (...args) => ...)`).
	expectConfigReceiverArray
	// expectConfigExpect denotes the Rstest `expect` function/object.
	expectConfigExpect
	// expectConfigExtend denotes `expect.extend`.
	expectConfigExtend
	// expectConfigEquality denotes `expect.addEqualityTesters`.
	expectConfigEquality
)

// collectExpectCustomization scans the source file once to record matcher
// overrides (`expect.extend`) and custom equality testers
// (`expect.addEqualityTesters`).
//
// Direct calls and clean local aliases (`const check = expect`,
// `const { extend } = expect`, `const ext = expect.extend`, static folded
// bracket keys, shallow copies) are tracked precisely. Any non-trivial escape
// of an Rstest receiver, `expect`, `expect.extend`, or `expect.addEqualityTesters`
// (for example storing it in an object/array, assigning it to a property inside
// a getter or setter, calling `.bind`/`.call`/`.apply`, returning it, or passing
// it to an external helper) conservatively falls back to marking customization
// active rather than attempting to symbolically execute JavaScript object models.
func (analysis *RstestCallAnalysis) collectExpectCustomization() {
	if analysis.expectCustomizationOK {
		return
	}
	analysis.expectCustomizationOK = true
	if analysis.ctx.SourceFile == nil {
		return
	}
	analysis.callbacksRef()

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || ast.IsTypeNode(node) {
			return
		}
		if analysis.allExpectMatchersOverridden && analysis.hasCustomEqualityTesters {
			return
		}
		switch node.Kind {
		case ast.KindCallExpression:
			callee := node.AsCallExpression().Expression
			kind := analysis.expectConfigMember(callee)
			if kind&(expectConfigExtend|expectConfigEquality) == 0 {
				if parsed := analysis.ParseExpectCall(node); parsed != nil &&
					parsed.Entry == RstestExpectEntryStatic &&
					testFramework.DirectMemberCall(parsed.MatcherEntry) == node {
					switch parsed.Matcher {
					case "extend":
						kind |= expectConfigExtend
					case "addEqualityTesters":
						kind |= expectConfigEquality
					}
				}
			}
			if kind&expectConfigExtend != 0 {
				analysis.recordExpectExtendCall(node)
			}
			if kind&expectConfigEquality != 0 {
				analysis.recordExpectEqualityCall(node)
			}
			if analysis.isDynamicRstestImportCall(node) {
				outer := internalUtils.OutermostTransparentExpression(node)
				if outer == nil || outer.Parent == nil ||
					outer.Parent.Kind != ast.KindAwaitExpression ||
					outer.Parent.AsAwaitExpression().Expression != outer {
					analysis.allExpectMatchersOverridden = true
					analysis.hasCustomEqualityTesters = true
				}
			} else if resultKind := analysis.expectConfigMember(node); resultKind != expectConfigNone {
				analysis.checkExpectConfigEscape(node, resultKind)
			}
		case ast.KindAwaitExpression:
			if kind := analysis.expectConfigMember(node); kind != expectConfigNone {
				analysis.checkExpectConfigEscape(node, kind)
			}
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindMetaProperty:
			if kind := analysis.expectConfigMember(node); kind != expectConfigNone {
				analysis.checkExpectConfigEscape(node, kind)
			}
		case ast.KindIdentifier:
			if scope.IsReferenceIdentifier(node) {
				if kind := analysis.expectConfigMember(node); kind != expectConfigNone {
					analysis.checkExpectConfigEscape(node, kind)
				}
			}
		case ast.KindBindingElement:
			be := node.AsBindingElement()
			if be != nil && be.Name() != nil {
				if be.Name().Kind != ast.KindIdentifier {
					// Nested destructuring of `extend` or `addEqualityTesters` (e.g.
					// `const { extend: { call } } = expect`) extracts methods off the
					// function object and escapes conservatively.
					kind := analysis.destructuredExpectConfigMember(node)
					if be.Initializer != nil {
						kind |= analysis.expectConfigMember(be.Initializer)
					}
					if kind&expectConfigExtend != 0 {
						analysis.allExpectMatchersOverridden = true
					}
					if kind&expectConfigEquality != 0 {
						analysis.hasCustomEqualityTesters = true
					}
				} else if isExportedVariableDeclaration(internalUtils.EnclosingVariableDeclarationOfBindingElement(node)) {
					// Exported destructuring (`export const { expect } = r`,
					// `export const { ...api } = r`, `export const { extend } = r.expect`)
					// escapes only the bindings that actually extract expect-config values,
					// so `export const { test, describe } = r` remains non-escaping.
					kind := analysis.destructuredExpectConfigMember(node)
					if be.Initializer != nil {
						kind |= analysis.expectConfigMember(be.Initializer)
					}
					if kind != expectConfigNone {
						analysis.markExpectConfigEscaped(kind)
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return analysis.allExpectMatchersOverridden && analysis.hasCustomEqualityTesters
		})
	}
	visit(analysis.ctx.SourceFile.AsNode())

	// If any `expect(expect.extend).<safeMatcher>()` call relied on a matcher
	// that the file itself overrode via `expect.extend`, that custom matcher
	// receives the argument and escapes it conservatively. Re-check once if
	// an escape flips `allExpectMatchersOverridden` so earlier safe-matcher
	// checks (e.g. `expect(expect).toBeTruthy()`) also escalate.
	for _, check := range analysis.safeMatcherChecks {
		if analysis.allExpectMatchersOverridden || analysis.overriddenExpectMatchers[check.matcherName] {
			analysis.markExpectConfigEscaped(check.argKind)
		}
	}
	if analysis.allExpectMatchersOverridden && !analysis.hasCustomEqualityTesters {
		for _, check := range analysis.safeMatcherChecks {
			analysis.markExpectConfigEscaped(check.argKind)
		}
	}
	analysis.safeMatcherChecks = nil
}

func (analysis *RstestCallAnalysis) markExpectConfigEscaped(kind expectConfigKind) {
	if kind&(expectConfigExtend|expectConfigEquality) != 0 {
		if kind&expectConfigExtend != 0 {
			analysis.allExpectMatchersOverridden = true
		}
		if kind&expectConfigEquality != 0 {
			analysis.hasCustomEqualityTesters = true
		}
		return
	}
	if kind&(expectConfigExpect|expectConfigReceiver|expectConfigReceiverArray) != 0 {
		analysis.allExpectMatchersOverridden = true
		analysis.hasCustomEqualityTesters = true
	}
}

func (analysis *RstestCallAnalysis) recordExpectExtendCall(call *ast.Node) {
	arguments := call.Arguments()
	if len(arguments) == 0 {
		return
	}
	if arguments[0] == nil || arguments[0].Kind == ast.KindSpreadElement {
		analysis.allExpectMatchersOverridden = true
		return
	}
	matchers := internalUtils.SkipAssertionsAndParens(arguments[0])
	if matchers == nil || matchers.Kind != ast.KindObjectLiteralExpression || matchers.AsObjectLiteralExpression().Properties == nil {
		analysis.allExpectMatchersOverridden = true
		return
	}
	for _, property := range matchers.AsObjectLiteralExpression().Properties.Nodes {
		if property == nil || property.Kind == ast.KindSpreadAssignment || property.Name() == nil {
			analysis.allExpectMatchersOverridden = true
			return
		}
		matcher, known, staticKey := analysis.evalPropertyNameKey(property.Name())
		if !known {
			if !staticKey {
				analysis.allExpectMatchersOverridden = true
				return
			}
			// Static Symbol keys cannot override named string matchers.
			continue
		}
		if analysis.overriddenExpectMatchers == nil {
			analysis.overriddenExpectMatchers = map[string]bool{}
		}
		analysis.overriddenExpectMatchers[matcher] = true
	}
}

func (analysis *RstestCallAnalysis) recordExpectEqualityCall(call *ast.Node) {
	arguments := call.Arguments()
	if len(arguments) == 0 {
		return
	}
	if arguments[0] != nil && arguments[0].Kind != ast.KindSpreadElement {
		testers := internalUtils.SkipAssertionsAndParens(arguments[0])
		if testers != nil && testers.Kind == ast.KindArrayLiteralExpression &&
			(testers.AsArrayLiteralExpression().Elements == nil || len(testers.AsArrayLiteralExpression().Elements.Nodes) == 0) {
			return
		}
	}
	analysis.hasCustomEqualityTesters = true
}

func isUnshadowedGlobalConstructor(node *ast.Node, analysis *RstestCallAnalysis) bool {
	u := internalUtils.SkipAssertionsAndParens(node)
	if u == nil || u.Kind != ast.KindIdentifier {
		return false
	}
	name := u.AsIdentifier().Text
	if name != "Function" && name != "Object" {
		return false
	}
	sym := resolveRstestRootSymbol(analysis.ctx, u)
	if hasSameFileRuntimeValueDecl(sym, analysis.ctx.SourceFile) {
		return false
	}
	return sym != nil || analysis.ctx.Refs != nil || !internalUtils.IsShadowed(u, name)
}

func safeReadOnlyExpectMatcherName(call *ast.Node, analysis *RstestCallAnalysis) (string, bool) {
	top := FindTopMostCallExpression(call)
	parsed := analysis.ParseExpectCall(top)
	if parsed == nil || parsed.Head != call || len(parsed.Matchers) != 1 || parsed.Expression == nil {
		return "", false
	}
	exprOuter := internalUtils.OutermostTransparentExpression(parsed.Expression)
	if exprOuter == nil || exprOuter.Parent == nil {
		return "", false
	}
	inSafeParent := (exprOuter.Parent.Kind == ast.KindExpressionStatement && exprOuter.Parent.AsExpressionStatement().Expression == exprOuter) ||
		(exprOuter.Parent.Kind == ast.KindArrowFunction && exprOuter.Parent.AsArrowFunction().Body == exprOuter)
	if !inSafeParent {
		return "", false
	}
	matcher := parsed.Matchers[0]
	switch matcher.Name {
	case "toBeDefined", "toBeUndefined", "toBeNull", "toBeTruthy", "toBeFalsy", "toBeNaN":
		return matcher.Name, true
	case "toBeInstanceOf":
		matcherCall := testFramework.DirectMemberCall(&matcher.Entry)
		if matcherCall == nil || matcherCall.Kind != ast.KindCallExpression {
			return "", false
		}
		args := matcherCall.AsCallExpression().Arguments
		if args == nil || len(args.Nodes) != 1 || !isUnshadowedGlobalConstructor(args.Nodes[0], analysis) {
			return "", false
		}
		return matcher.Name, true
	case "toBeTypeOf", "toBe":
		matcherCall := testFramework.DirectMemberCall(&matcher.Entry)
		if matcherCall == nil || matcherCall.Kind != ast.KindCallExpression {
			return "", false
		}
		if args := matcherCall.AsCallExpression().Arguments; args != nil {
			for _, a := range args.Nodes {
				u := internalUtils.SkipAssertionsAndParens(a)
				if u == nil {
					return "", false
				}
				switch u.Kind {
				case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
					ast.KindNumericLiteral, ast.KindBigIntLiteral,
					ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword,
					ast.KindUndefinedKeyword, ast.KindIdentifier:
				default:
					return "", false
				}
			}
		}
		return matcher.Name, true
	}
	return "", false
}

func isUnshadowedGlobalIdent(node *ast.Node, expected string, analysis *RstestCallAnalysis) bool {
	u := internalUtils.SkipAssertionsAndParens(node)
	if u == nil || u.Kind != ast.KindIdentifier || u.AsIdentifier().Text != expected {
		return false
	}
	sym := resolveRstestRootSymbol(analysis.ctx, u)
	if hasSameFileRuntimeValueDecl(sym, analysis.ctx.SourceFile) {
		return false
	}
	return sym != nil || analysis.ctx.Refs != nil || !internalUtils.IsShadowed(u, expected)
}

func isReadOnlyBuiltinInspectionCall(call *ast.Node, arg *ast.Node, analysis *RstestCallAnalysis) bool {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	args := call.AsCallExpression().Arguments
	if args == nil || len(args.Nodes) == 0 {
		return false
	}
	callee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	if callee == nil {
		return false
	}
	// `Boolean(x)` coerces to boolean without invoking `x` or leaking it.
	if isUnshadowedGlobalIdent(callee, "Boolean", analysis) {
		return len(args.Nodes) == 1 && args.Nodes[0] == arg
	}
	if !ast.IsAccessExpression(callee) {
		return false
	}
	if !isUnshadowedGlobalIdent(internalUtils.AccessExpressionObject(callee), "Object", analysis) {
		return false
	}
	name, known, _ := analysis.evalExpectConfigAccessName(callee)
	if !known {
		return false
	}
	switch name {
	case "isFrozen", "isSealed", "isExtensible", "keys", "getOwnPropertyNames", "getOwnPropertySymbols":
		return len(args.Nodes) == 1 && args.Nodes[0] == arg
	case "hasOwn":
		return len(args.Nodes) >= 2 && args.Nodes[0] == arg
	case "is":
		return len(args.Nodes) == 2 && (args.Nodes[0] == arg || args.Nodes[1] == arg)
	case "freeze", "seal", "preventExtensions":
		// These return their argument, so they are only safe when discarded.
		return len(args.Nodes) == 1 && args.Nodes[0] == arg && isStandaloneExpressionStmt(call)
	}
	return false
}

func isSafeReadOnlyExpectUse(outer *ast.Node, kind expectConfigKind, analysis *RstestCallAnalysis) bool {
	if outer == nil || outer.Parent == nil {
		return true
	}
	p := outer.Parent
	switch p.Kind {
	case ast.KindExpressionStatement:
		return p.AsExpressionStatement().Expression == outer
	case ast.KindIfStatement:
		return p.AsIfStatement().Expression == outer
	case ast.KindWhileStatement:
		return p.AsWhileStatement().Expression == outer
	case ast.KindDoStatement:
		return p.AsDoStatement().Expression == outer
	case ast.KindForStatement:
		return p.AsForStatement().Condition == outer
	case ast.KindConditionalExpression:
		return p.AsConditionalExpression().Condition == outer
	case ast.KindTypeOfExpression, ast.KindVoidExpression:
		return true
	case ast.KindPrefixUnaryExpression:
		return p.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindBinaryExpression:
		bin := p.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindEqualsEqualsToken, ast.KindExclamationEqualsToken,
			ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return true
		case ast.KindInKeyword:
			return bin.Right == outer
		case ast.KindInstanceOfKeyword:
			// Arbitrary RHS can define `[Symbol.hasInstance](lhs)` and invoke `lhs`.
			return bin.Left == outer && isUnshadowedGlobalConstructor(bin.Right, analysis)
		case ast.KindAmpersandAmpersandToken, ast.KindCommaToken:
			return bin.Left == outer
		}
	case ast.KindCallExpression:
		if p.AsCallExpression().Expression != outer {
			if matcherName, ok := safeReadOnlyExpectMatcherName(p, analysis); ok {
				if analysis.allExpectMatchersOverridden || analysis.overriddenExpectMatchers[matcherName] {
					return false
				}
				analysis.safeMatcherChecks = append(analysis.safeMatcherChecks, safeExpectMatcherCheck{
					matcherName: matcherName,
					argKind:     kind,
				})
				return true
			}
			if isReadOnlyBuiltinInspectionCall(p, outer, analysis) {
				return true
			}
		}
	}
	return false
}

func isStandaloneExpressionStmt(node *ast.Node) bool {
	outer := internalUtils.OutermostTransparentExpression(node)
	return outer != nil && outer.Parent != nil &&
		outer.Parent.Kind == ast.KindExpressionStatement &&
		outer.Parent.AsExpressionStatement().Expression == outer
}

func (analysis *RstestCallAnalysis) isTrackedLocalAssignment(bin *ast.BinaryExpression, outer *ast.Node) bool {
	if bin == nil || bin.OperatorToken == nil {
		return false
	}
	if internalUtils.IsDefaultValueInDestructuringAssignment(bin.AsNode()) {
		return false
	}
	lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
	if lhs == nil || lhs.Kind != ast.KindIdentifier {
		return false
	}
	sym := resolveRstestRootSymbol(analysis.ctx, lhs)
	if !hasSameFileRuntimeValueDecl(sym, analysis.ctx.SourceFile) {
		return false
	}
	if bin.Left == outer {
		if bin.OperatorToken.Kind == ast.KindEqualsToken {
			// Plain reassignment target of a local variable (`check = other`,
			// `ext = other`) overwrites the binding without leaking its old value.
			return true
		}
		switch bin.OperatorToken.Kind {
		case ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken:
			return isStandaloneExpressionStmt(bin.AsNode())
		}
	}
	if bin.Right == outer {
		if hasExportedDeclaration(sym) {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindEqualsToken, ast.KindBarBarEqualsToken,
			ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken:
			// Only non-escaping when the assignment expression's result value is
			// discarded (standalone statement); `(tmp = expect.extend)` used as a
			// nested expression leaks the RHS value to its parent.
			return isStandaloneExpressionStmt(bin.AsNode())
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) objectLiteralDestructuresKey(obj *ast.Node, target string) bool {
	if obj == nil || obj.Kind != ast.KindObjectLiteralExpression || obj.AsObjectLiteralExpression().Properties == nil {
		return false
	}
	for _, prop := range obj.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		if prop.Kind == ast.KindSpreadAssignment {
			return true
		}
		if prop.Kind == ast.KindShorthandPropertyAssignment {
			if prop.Name() != nil && prop.Name().Kind == ast.KindIdentifier && prop.Name().AsIdentifier().Text == target {
				return true
			}
			continue
		}
		if prop.Name() != nil {
			name, known, staticKey := analysis.evalPropertyNameKey(prop.Name())
			if !staticKey || (known && name == target) {
				return true
			}
		}
	}
	return false
}

func isExportedVariableDeclaration(vd *ast.Node) bool {
	return vd != nil &&
		vd.Parent != nil &&
		vd.Parent.Kind == ast.KindVariableDeclarationList &&
		vd.Parent.Parent != nil &&
		vd.Parent.Parent.Kind == ast.KindVariableStatement &&
		ast.HasSyntacticModifier(vd.Parent.Parent, ast.ModifierFlagsExport)
}

func hasExportedDeclaration(sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	for _, decl := range sym.Declarations {
		if decl == nil {
			continue
		}
		if decl.Kind == ast.KindBindingElement {
			decl = internalUtils.EnclosingVariableDeclarationOfBindingElement(decl)
		}
		if isExportedVariableDeclaration(decl) || ast.HasSyntacticModifier(decl, ast.ModifierFlagsExport) {
			return true
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) checkExpectConfigEscape(node *ast.Node, kind expectConfigKind) {
	outer := internalUtils.OutermostTransparentExpression(node)
	if outer == nil || outer.Parent == nil {
		return
	}
	p := outer.Parent
	if isSafeReadOnlyExpectUse(outer, kind, analysis) {
		return
	}

	if kind&(expectConfigExtend|expectConfigEquality) != 0 {
		// Safe uses of `expect.extend` / `expect.addEqualityTesters`:
		// 1. Direct callee: `expect.extend(...)` / `ext(...)`
		if p.Kind == ast.KindCallExpression && p.AsCallExpression().Expression == outer {
			return
		}
		// 2. Simple non-exported local variable initializer: `const ext = expect.extend`
		if p.Kind == ast.KindVariableDeclaration && p.AsVariableDeclaration().Initializer == outer &&
			!isExportedVariableDeclaration(p) {
			name := p.AsVariableDeclaration().Name()
			if name != nil && name.Kind == ast.KindIdentifier {
				return
			}
		}
		// 3. Plain local variable overwrite/assignment: `ext = other`, `fn = expect.extend`
		if p.Kind == ast.KindBinaryExpression && analysis.isTrackedLocalAssignment(p.AsBinaryExpression(), outer) {
			return
		}
		if kind&expectConfigExtend != 0 {
			analysis.allExpectMatchersOverridden = true
		}
		if kind&expectConfigEquality != 0 {
			analysis.hasCustomEqualityTesters = true
		}
		return
	}

	if kind&expectConfigExpect != 0 {
		// Safe uses of `expect`:
		// 1. Direct assertion call: `expect(...)`
		if p.Kind == ast.KindCallExpression && p.AsCallExpression().Expression == outer {
			return
		}
		// 2. Receiver of member access: `expect.toBe`, `expect.extend`, `expect[k]`
		if ast.IsAccessExpression(p) && p.Expression() == outer {
			return
		}
		// 3. Initializer of a local VariableDeclaration, or an exported
		//    destructuring declaration whose individual bindings are checked
		//    in the KindBindingElement visitor:
		//    `const check = expect`, `const { extend } = expect`, `const { ...copy } = expect`
		if p.Kind == ast.KindVariableDeclaration && p.AsVariableDeclaration().Initializer == outer &&
			(!isExportedVariableDeclaration(p) || (p.AsVariableDeclaration().Name() != nil && p.AsVariableDeclaration().Name().Kind != ast.KindIdentifier)) {
			return
		}
		// 4. Plain local variable overwrite/assignment: `check = other`, `c = expect`
		if p.Kind == ast.KindBinaryExpression && analysis.isTrackedLocalAssignment(p.AsBinaryExpression(), outer) {
			return
		}
		analysis.allExpectMatchersOverridden = true
		analysis.hasCustomEqualityTesters = true
		return
	}

	if kind&(expectConfigReceiver|expectConfigReceiverArray) != 0 {
		// Safe uses of an Rstest namespace/receiver or rest context array:
		// 1. Receiver of member access: `r.test`, `r.expect`, `ctx.expect`, `args[0]`
		if ast.IsAccessExpression(p) && p.Expression() == outer {
			return
		}
		// 2. Initializer of a local VariableDeclaration, or an exported
		//    destructuring declaration whose selected bindings are checked in
		//    the KindBindingElement visitor:
		//    `const { expect } = r`, `export const { test, describe } = r`, `const core = r`, `const [ctx] = args`
		if p.Kind == ast.KindVariableDeclaration && p.AsVariableDeclaration().Initializer == outer &&
			(!isExportedVariableDeclaration(p) || (p.AsVariableDeclaration().Name() != nil && p.AsVariableDeclaration().Name().Kind != ast.KindIdentifier)) {
			return
		}
		// 3. Plain local variable overwrite/assignment: `core = other`, `c = r`, `c ??= r`
		if p.Kind == ast.KindBinaryExpression && analysis.isTrackedLocalAssignment(p.AsBinaryExpression(), outer) {
			return
		}
		// Any other use (storing in an object/array, spreading `{ ...r }`,
		// passing to a helper / `Object.assign` / `Reflect.get`, returning,
		// destructuring assignment `({ expect } = r)`, default initializer, etc.)
		// conservatively escapes.
		analysis.allExpectMatchersOverridden = true
		analysis.hasCustomEqualityTesters = true
		return
	}

	if kind&(expectConfigGlobalObject|expectConfigImportMeta) != 0 {
		// `globalThis` / `window` / `import.meta` only escape when destructuring-assigned
		// directly into a pattern that may extract `expect` / `rstest`.
		if p.Kind == ast.KindBinaryExpression {
			bin := p.AsBinaryExpression()
			if bin != nil && bin.Right == outer && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
				lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
				if lhs != nil && lhs.Kind == ast.KindObjectLiteralExpression {
					if (kind&expectConfigGlobalObject != 0 && analysis.objectLiteralDestructuresKey(lhs, "expect")) ||
						(kind&expectConfigImportMeta != 0 && analysis.objectLiteralDestructuresKey(lhs, "rstest")) {
						analysis.allExpectMatchersOverridden = true
						analysis.hasCustomEqualityTesters = true
					}
				}
			}
		}
	}
}

func hasSameFileRuntimeValueDecl(symbol *ast.Symbol, sf *ast.SourceFile) bool {
	if symbol == nil {
		return false
	}
	for _, decl := range symbol.Declarations {
		if isSameFileRuntimeValueDecl(decl, sf) {
			return true
		}
	}
	return false
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

func (analysis *RstestCallAnalysis) isDynamicRstestImportCall(node *ast.Node) bool {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	call := node.AsCallExpression()
	if call == nil || call.Expression == nil || call.Expression.Kind != ast.KindImportKeyword ||
		call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return false
	}
	arg := internalUtils.SkipAssertionsAndParens(call.Arguments.Nodes[0])
	if arg == nil {
		return false
	}
	if s, ok := internalUtils.GetStaticStringLiteralValue(arg); ok {
		return isRstestImportModule(s)
	}
	if s, ok := analysis.staticStringEvaluator().EvalToString(arg); ok {
		return isRstestImportModule(s)
	}
	return false
}

func (analysis *RstestCallAnalysis) expectConfigMember(node *ast.Node) expectConfigKind {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return expectConfigNone
	}
	if isImportMetaOptions(node, true) {
		return expectConfigImportMeta
	}
	if isRstestRequireCall(node) {
		return expectConfigReceiver
	}
	if node.Kind == ast.KindAwaitExpression {
		if analysis.isDynamicRstestImportCall(node.AsAwaitExpression().Expression) {
			return expectConfigReceiver
		}
		return expectConfigNone
	}
	if ast.IsAccessExpression(node) {
		canCache := len(analysis.sccStack) == 0 && !analysis.sccQueueActive
		if canCache && analysis.expectConfigExprs != nil {
			if cached, ok := analysis.expectConfigExprs[node]; ok {
				return cached
			}
		}
		recvExpr := internalUtils.SkipAssertionsAndParens(node.Expression())
		recvKind := analysis.expectConfigMember(recvExpr)
		var out expectConfigKind
		if recvKind != expectConfigNone {
			member, known, staticKey := analysis.evalExpectConfigAccessName(node)
			if recvKind&expectConfigGlobalObject != 0 {
				if known && member == "expect" {
					out |= expectConfigExpect
				}
			}
			if recvKind&expectConfigImportMeta != 0 {
				if known && member == "rstest" {
					out |= expectConfigReceiver
				}
			}
			if recvKind&expectConfigReceiverArray != 0 {
				if (known && member == "0") || !staticKey {
					out |= expectConfigReceiver
				}
			}
			if recvKind&expectConfigReceiver != 0 {
				if known {
					if member == "expect" {
						out |= expectConfigExpect
					}
				} else if !staticKey {
					out |= expectConfigExpect
				}
			}
			if recvKind&expectConfigExpect != 0 {
				if known {
					switch member {
					case "extend":
						if !analysis.isShallowCopyPropKilled(recvExpr, "extend") {
							out |= expectConfigExtend
						}
					case "addEqualityTesters":
						if !analysis.isShallowCopyPropKilled(recvExpr, "addEqualityTesters") {
							out |= expectConfigEquality
						}
					}
				} else if !staticKey {
					out |= expectConfigExtend | expectConfigEquality
				}
			}
		}
		if canCache && len(analysis.sccStack) == 0 && !analysis.sccQueueActive {
			if analysis.expectConfigExprs == nil {
				analysis.expectConfigExprs = map[*ast.Node]expectConfigKind{}
			}
			analysis.expectConfigExprs[node] = out
		}
		return out
	}
	if node.Kind != ast.KindIdentifier {
		return expectConfigNone
	}
	symbol := resolveRstestRootSymbol(analysis.ctx, node)
	if symbol == nil || !hasSameFileRuntimeValueDecl(symbol, analysis.ctx.SourceFile) {
		name := node.AsIdentifier().Text
		if symbol == nil && analysis.ctx.Refs == nil && internalUtils.IsShadowed(node, name) {
			return expectConfigNone
		}
		switch name {
		case "expect":
			return expectConfigExpect
		case "globalThis", "window", "global", "self":
			return expectConfigGlobalObject
		}
		return expectConfigNone
	}
	return analysis.classifyExpectConfigSymbol(symbol)
}

// evaluateExpectConfigSymbolOnce computes one monotone step of `symbol`'s
// expectConfigKind from its declarations and direct/logical assignment writes.
// When `sccQueueRecording` is active, it records each individual incoming edge
// on the referenced SCC members so the subsequent queue propagates edge-by-edge
// without rescanning a hub symbol's full reference list.
func (analysis *RstestCallAnalysis) evaluateExpectConfigSymbolOnce(symbol *ast.Symbol) expectConfigKind {
	prevEdge := analysis.sccQueueEdge
	defer func() {
		analysis.sccQueueEdge = prevEdge
	}()

	callbacks := analysis.callbacksRef()
	var out expectConfigKind
	if callbacks.ContextReceivers[symbol] {
		out |= expectConfigReceiver
	}
	if callbacks.ContextExpectNames[symbol] {
		out |= expectConfigExpect
	}

	for _, decl := range symbol.Declarations {
		if decl == nil {
			continue
		}
		switch decl.Kind {
		case ast.KindParameter:
			pd := decl.AsParameterDeclaration()
			if callbacks.ContextParameters[decl] {
				if pd != nil && pd.DotDotDotToken != nil {
					out |= expectConfigReceiverArray
				} else {
					out |= expectConfigReceiver
				}
			}
			if pd != nil && pd.Initializer != nil {
				analysis.sccQueueEdge = sccRevEdge{dst: symbol, expr: pd.Initializer}
				out |= analysis.expectConfigMember(pd.Initializer)
			}
		case ast.KindImportSpecifier:
			imp := testFramework.FindImportDeclaration(decl)
			if imp != nil && imp.ModuleSpecifier != nil && isRstestImportModule(imp.ModuleSpecifier.Text()) &&
				!ast.IsTypeOnlyImportDeclaration(decl) {
				if testFramework.ImportedSpecifierName(decl) == "expect" {
					out |= expectConfigExpect
				}
			}
		case ast.KindNamespaceImport:
			imp := testFramework.FindImportDeclaration(decl)
			if imp != nil && imp.ModuleSpecifier != nil && isRstestImportModule(imp.ModuleSpecifier.Text()) &&
				!ast.IsTypeOnlyImportDeclaration(decl) {
				out |= expectConfigReceiver
			}
		case ast.KindImportEqualsDeclaration:
			if testFramework.IsImportEqualsOfModules(decl, RstestAllImportModules) &&
				!decl.AsImportEqualsDeclaration().IsTypeOnly {
				out |= expectConfigReceiver
			}
		case ast.KindVariableDeclaration:
			vd := decl.AsVariableDeclaration()
			if vd == nil || vd.Initializer == nil {
				continue
			}
			analysis.sccQueueEdge = sccRevEdge{dst: symbol, expr: vd.Initializer}
			k := analysis.expectConfigMember(vd.Initializer)
			if k == expectConfigNone {
				continue
			}
			if (!ast.IsVarConst(decl) || ast.IsVarAwaitUsing(decl)) &&
				analysis.isUnconditionallyKilledBeforeRead(symbol, decl) {
				continue
			}
			out |= k
		case ast.KindBindingElement:
			analysis.sccQueueEdge = sccRevEdge{dst: symbol, bindingElem: decl}
			k := analysis.destructuredExpectConfigMember(decl)
			if be := decl.AsBindingElement(); be != nil && be.Initializer != nil {
				analysis.sccQueueEdge = sccRevEdge{dst: symbol, expr: be.Initializer}
				k |= analysis.expectConfigMember(be.Initializer)
			}
			if k == expectConfigNone {
				continue
			}
			if !isConstBindingElement(decl) &&
				analysis.isUnconditionallyKilledBeforeRead(symbol, decl) {
				continue
			}
			out |= k
		}
	}

	// Include direct and logical assignment writes (`let fn; fn = expect.extend;`, `c ??= r;`).
	if analysis.ctx.Refs != nil {
		for _, ref := range analysis.ctx.Refs.References(symbol) {
			if ref == nil {
				continue
			}
			outer := internalUtils.OutermostTransparentExpression(ref)
			if outer == nil || outer.Parent == nil || outer.Parent.Kind != ast.KindBinaryExpression {
				continue
			}
			bin := outer.Parent.AsBinaryExpression()
			if bin == nil || bin.OperatorToken == nil || bin.Left != outer {
				continue
			}
			switch bin.OperatorToken.Kind {
			case ast.KindEqualsToken, ast.KindBarBarEqualsToken,
				ast.KindAmpersandAmpersandEqualsToken, ast.KindQuestionQuestionEqualsToken:
				if !internalUtils.IsDefaultValueInDestructuringAssignment(outer.Parent) {
					analysis.sccQueueEdge = sccRevEdge{dst: symbol, expr: bin.Right}
					out |= analysis.expectConfigMember(bin.Right)
				}
			}
		}
	}
	return out
}

// classifyExpectConfigSymbol uses Tarjan SCC discovery plus an edge-level
// monotone fixed-point queue over each cyclic SCC so arbitrary alias cycles
// (including `a = a.extend` member-access cycles and hub-and-chain branching
// cycles) converge cleanly in O(V + E) time without rescanning reference lists.
func (analysis *RstestCallAnalysis) classifyExpectConfigSymbol(symbol *ast.Symbol) expectConfigKind {
	if symbol == nil {
		return expectConfigNone
	}
	if analysis.sccQueueRecording && analysis.sccQueueEdge.dst != nil && analysis.sccQueueMembers[symbol] {
		seen := analysis.sccQueueRevEdgeSeen[symbol]
		if seen == nil {
			seen = map[sccRevEdge]bool{}
			analysis.sccQueueRevEdgeSeen[symbol] = seen
		}
		if !seen[analysis.sccQueueEdge] {
			seen[analysis.sccQueueEdge] = true
			analysis.sccQueueRevEdges[symbol] = append(analysis.sccQueueRevEdges[symbol], analysis.sccQueueEdge)
		}
	}
	if cached, ok := analysis.expectConfigAliases[symbol]; ok {
		return cached
	}
	if analysis.sccOnStack[symbol] {
		analysis.sccHitActive = true
		if caller := analysis.sccCurrent; caller != nil {
			if idx := analysis.sccIndices[symbol]; idx < analysis.sccLowLink[caller] {
				analysis.sccLowLink[caller] = idx
			}
		}
		return expectConfigNone
	}
	if analysis.sccIndices == nil {
		analysis.sccIndices = map[*ast.Symbol]int{}
		analysis.sccLowLink = map[*ast.Symbol]int{}
		analysis.sccOnStack = map[*ast.Symbol]bool{}
	}

	caller := analysis.sccCurrent
	analysis.sccIndex++
	myIdx := analysis.sccIndex
	analysis.sccIndices[symbol] = myIdx
	analysis.sccLowLink[symbol] = myIdx
	analysis.sccStack = append(analysis.sccStack, symbol)
	analysis.sccOnStack[symbol] = true

	prevHit := analysis.sccHitActive
	prevRecording := analysis.sccQueueRecording
	analysis.sccHitActive = false
	analysis.sccQueueRecording = false
	analysis.sccCurrent = symbol
	out := analysis.evaluateExpectConfigSymbolOnce(symbol)
	analysis.sccCurrent = caller
	analysis.sccQueueRecording = prevRecording
	hadSelfOrCycleHit := analysis.sccHitActive
	analysis.sccHitActive = prevHit || hadSelfOrCycleHit

	if caller != nil && analysis.sccLowLink[symbol] < analysis.sccLowLink[caller] {
		analysis.sccLowLink[caller] = analysis.sccLowLink[symbol]
	}

	if analysis.sccLowLink[symbol] != myIdx {
		// Non-root member of an open SCC: stays on sccStack until the SCC root
		// pops the whole component and runs the monotone fixed point.
		return out
	}

	if analysis.expectConfigAliases == nil {
		analysis.expectConfigAliases = map[*ast.Symbol]expectConfigKind{}
	}

	// Root of a strongly connected component: pop all members.
	var component []*ast.Symbol
	for len(analysis.sccStack) > 0 {
		top := analysis.sccStack[len(analysis.sccStack)-1]
		analysis.sccStack = analysis.sccStack[:len(analysis.sccStack)-1]
		analysis.sccOnStack[top] = false
		component = append(component, top)
		if top == symbol {
			break
		}
	}
	if len(component) == 1 && !hadSelfOrCycleHit {
		analysis.expectConfigAliases[symbol] = out
		return out
	}
	// Seed every SCC member in the cache, run one recording pass to collect
	// exact reverse edges, then propagate monotonically edge-by-edge. Because
	// each symbol can gain at most 7 bits, every intra-SCC edge is evaluated
	// O(1) times and reference lists are never scanned again (strictly O(V + E)).
	members := make(map[*ast.Symbol]bool, len(component))
	revEdges := make(map[*ast.Symbol][]sccRevEdge, len(component))
	revEdgeSeen := make(map[*ast.Symbol]map[sccRevEdge]bool, len(component))
	for _, member := range component {
		analysis.expectConfigAliases[member] = expectConfigNone
		members[member] = true
	}
	analysis.expectConfigAliases[symbol] = out

	prevActive := analysis.sccQueueActive
	prevRec := analysis.sccQueueRecording
	prevEdge := analysis.sccQueueEdge
	prevMembers := analysis.sccQueueMembers
	prevRevEdges := analysis.sccQueueRevEdges
	prevRevEdgeSeen := analysis.sccQueueRevEdgeSeen
	analysis.sccQueueActive = true
	analysis.sccQueueRecording = true
	analysis.sccQueueEdge = sccRevEdge{}
	analysis.sccQueueMembers = members
	analysis.sccQueueRevEdges = revEdges
	analysis.sccQueueRevEdgeSeen = revEdgeSeen
	defer func() {
		analysis.sccQueueActive = prevActive
		analysis.sccQueueRecording = prevRec
		analysis.sccQueueEdge = prevEdge
		analysis.sccQueueMembers = prevMembers
		analysis.sccQueueRevEdges = prevRevEdges
		analysis.sccQueueRevEdgeSeen = prevRevEdgeSeen
	}()

	queue := make([]*ast.Symbol, 0, len(component))
	inQueue := make(map[*ast.Symbol]bool, len(component))
	for _, member := range component {
		next := analysis.expectConfigAliases[member] | analysis.evaluateExpectConfigSymbolOnce(member)
		analysis.expectConfigAliases[member] = next
		if next != expectConfigNone && !inQueue[member] {
			inQueue[member] = true
			queue = append(queue, member)
		}
	}
	analysis.sccQueueRecording = false
	analysis.sccQueueEdge = sccRevEdge{}

	for len(queue) > 0 {
		changedSym := queue[0]
		queue = queue[1:]
		inQueue[changedSym] = false
		for _, edge := range revEdges[changedSym] {
			var edgeKind expectConfigKind
			if edge.bindingElem != nil {
				edgeKind = analysis.destructuredExpectConfigMember(edge.bindingElem)
			} else if edge.expr != nil {
				edgeKind = analysis.expectConfigMember(edge.expr)
			}
			next := analysis.expectConfigAliases[edge.dst] | edgeKind
			if next != analysis.expectConfigAliases[edge.dst] {
				analysis.expectConfigAliases[edge.dst] = next
				if !inQueue[edge.dst] {
					inQueue[edge.dst] = true
					queue = append(queue, edge.dst)
				}
			}
		}
	}
	return analysis.expectConfigAliases[symbol]
}

func isConstBindingElement(decl *ast.Node) bool {
	vd := internalUtils.EnclosingVariableDeclarationOfBindingElement(decl)
	return vd != nil && vd.Kind == ast.KindVariableDeclaration &&
		ast.IsVarConst(vd) && !ast.IsVarAwaitUsing(vd)
}

func (analysis *RstestCallAnalysis) statementIndexInParent(parent *ast.Node, stmt *ast.Node) ([]*ast.Node, int) {
	stmts := statementListOf(parent)
	if len(stmts) == 0 {
		return nil, -1
	}
	if analysis.stmtIndices == nil {
		analysis.stmtIndices = map[*ast.Node]map[*ast.Node]int{}
	}
	idxMap, ok := analysis.stmtIndices[parent]
	if !ok {
		idxMap = make(map[*ast.Node]int, len(stmts))
		for i, s := range stmts {
			idxMap[s] = i
		}
		analysis.stmtIndices[parent] = idxMap
	}
	idx, found := idxMap[stmt]
	if !found {
		return nil, -1
	}
	return stmts, idx
}

func isPureNonCallingKillRHS(expr *ast.Node) bool {
	u := internalUtils.SkipAssertionsAndParens(expr)
	if u == nil {
		return false
	}
	switch u.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword, ast.KindUndefinedKeyword,
		ast.KindArrowFunction, ast.KindFunctionExpression:
		return true
	}
	return false
}

func isInsideTryBlock(node *ast.Node) bool {
	for cur := node; cur != nil; cur = cur.Parent {
		if testFramework.IsFunction(cur) || cur.Kind == ast.KindClassStaticBlockDeclaration {
			return false
		}
		if cur.Parent != nil && cur.Parent.Kind == ast.KindTryStatement {
			ts := cur.Parent.AsTryStatement()
			if ts != nil && ts.TryBlock == cur && (ts.CatchClause != nil || ts.FinallyBlock != nil) {
				return true
			}
		}
	}
	return false
}

// isUnconditionallyKilledBeforeRead reports whether a non-const declaration
// (`let check = expect; check = other; check.extend(...)`) is immediately and
// unconditionally overwritten by a pure non-config assignment in the next
// statement before any reference can read it.
func (analysis *RstestCallAnalysis) isUnconditionallyKilledBeforeRead(symbol *ast.Symbol, decl *ast.Node) bool {
	if analysis.ctx.Refs == nil || symbol == nil || decl == nil || analysis.sccQueueActive {
		return false
	}
	vd := decl
	if decl.Kind == ast.KindBindingElement {
		be := decl.AsBindingElement()
		if be == nil || (be.PropertyName != nil && be.PropertyName.Kind == ast.KindComputedPropertyName) || be.Initializer != nil {
			return false
		}
		vd = internalUtils.EnclosingVariableDeclarationOfBindingElement(decl)
		if vd == nil || decl.Parent != vd.Name() || decl.Parent == nil {
			return false
		}
		bp := decl.Parent.AsBindingPattern()
		if bp == nil || bp.Elements == nil || len(bp.Elements.Nodes) == 0 ||
			bp.Elements.Nodes[len(bp.Elements.Nodes)-1] != decl {
			return false
		}
	}
	if vd == nil || vd.Kind != ast.KindVariableDeclaration || vd.Parent == nil || vd.Parent.Kind != ast.KindVariableDeclarationList {
		return false
	}
	declList := vd.Parent.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil || len(declList.Declarations.Nodes) == 0 ||
		declList.Declarations.Nodes[len(declList.Declarations.Nodes)-1] != vd {
		return false
	}
	varStmt := vd.Parent.Parent
	if varStmt == nil || varStmt.Kind != ast.KindVariableStatement {
		return false
	}
	// A hoisted `var` inside `try` can still be observed after `catch` if the
	// kill statement throws (e.g. ReferenceError on an undeclared identifier).
	if vd.Parent.Flags&ast.NodeFlagsBlockScoped == 0 && isInsideTryBlock(varStmt) {
		return false
	}
	stmts, varIdx := analysis.statementIndexInParent(varStmt.Parent, varStmt)
	if varIdx < 0 || varIdx+1 >= len(stmts) {
		return false
	}
	killStmt := stmts[varIdx+1]
	if killStmt == nil || killStmt.Kind != ast.KindExpressionStatement {
		return false
	}
	expr := ast.SkipParentheses(killStmt.AsExpressionStatement().Expression)
	if expr == nil || expr.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := expr.AsBinaryExpression()
	if bin == nil || bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken {
		return false
	}
	lhs := ast.SkipParentheses(bin.Left)
	if lhs == nil || lhs.Kind != ast.KindIdentifier || lhs.AsIdentifier().Text != symbol.Name ||
		resolveRstestRootSymbol(analysis.ctx, lhs) != symbol {
		return false
	}
	if !isPureNonCallingKillRHS(bin.Right) {
		return false
	}
	prevHit := analysis.sccHitActive
	analysis.sccHitActive = false
	rhsKind := analysis.expectConfigMember(bin.Right)
	rhsHitActive := analysis.sccHitActive
	analysis.sccHitActive = prevHit || rhsHitActive
	if rhsKind != expectConfigNone || rhsHitActive {
		return false
	}
	for _, ref := range analysis.ctx.Refs.References(symbol) {
		if ref == nil {
			continue
		}
		if ref == lhs {
			continue
		}
		if ref.Pos() < killStmt.End() {
			return false
		}
	}
	return true
}

func statementListOf(parent *ast.Node) []*ast.Node {
	if parent == nil {
		return nil
	}
	switch parent.Kind {
	case ast.KindBlock:
		if b := parent.AsBlock(); b != nil && b.Statements != nil {
			return b.Statements.Nodes
		}
	case ast.KindSourceFile:
		if sf := parent.AsSourceFile(); sf != nil && sf.Statements != nil {
			return sf.Statements.Nodes
		}
	case ast.KindModuleBlock:
		if mb := parent.AsModuleBlock(); mb != nil && mb.Statements != nil {
			return mb.Statements.Nodes
		}
	}
	return nil
}

// isShallowCopyPropKilled reports whether `recvExpr` is a `const { ...copy } = expect`
// shallow copy whose `propName` property is immediately overwritten to a pure
// non-config value in the next statement before any other use of `copy`.
func (analysis *RstestCallAnalysis) isShallowCopyPropKilled(recvExpr *ast.Node, propName string) bool {
	if analysis.ctx.Refs == nil || recvExpr == nil || recvExpr.Kind != ast.KindIdentifier || analysis.sccQueueActive {
		return false
	}
	sym := resolveRstestRootSymbol(analysis.ctx, recvExpr)
	if sym == nil || len(sym.Declarations) != 1 {
		return false
	}
	if cached, ok := analysis.shallowCopyKilledProps[sym]; ok {
		return cached[propName]
	}
	killed := map[string]bool{}
	if analysis.shallowCopyKilledProps == nil {
		analysis.shallowCopyKilledProps = map[*ast.Symbol]map[string]bool{}
	}
	analysis.shallowCopyKilledProps[sym] = killed

	decl := sym.Declarations[0]
	if decl == nil || decl.Kind != ast.KindBindingElement || decl.AsBindingElement().DotDotDotToken == nil || !isConstBindingElement(decl) {
		return false
	}
	vd := internalUtils.EnclosingVariableDeclarationOfBindingElement(decl)
	if vd == nil || decl.Parent != vd.Name() || vd.Parent == nil || vd.Parent.Kind != ast.KindVariableDeclarationList ||
		vd.Parent.Parent == nil || vd.Parent.Parent.Kind != ast.KindVariableStatement {
		return false
	}
	declList := vd.Parent.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil || len(declList.Declarations.Nodes) == 0 ||
		declList.Declarations.Nodes[len(declList.Declarations.Nodes)-1] != vd {
		return false
	}
	varStmt := vd.Parent.Parent
	stmts, varIdx := analysis.statementIndexInParent(varStmt.Parent, varStmt)
	if varIdx < 0 || varIdx+1 >= len(stmts) {
		return false
	}
	killStmt := stmts[varIdx+1]
	if killStmt == nil || killStmt.Kind != ast.KindExpressionStatement {
		return false
	}
	expr := ast.SkipParentheses(killStmt.AsExpressionStatement().Expression)
	if expr == nil || expr.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := expr.AsBinaryExpression()
	if bin == nil || bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken {
		return false
	}
	lhs := ast.SkipParentheses(bin.Left)
	if lhs == nil || !ast.IsAccessExpression(lhs) {
		return false
	}
	obj := ast.SkipParentheses(lhs.Expression())
	if obj == nil || obj.Kind != ast.KindIdentifier || obj.AsIdentifier().Text != sym.Name ||
		resolveRstestRootSymbol(analysis.ctx, obj) != sym {
		return false
	}
	pName, pKnown := internalUtils.AccessExpressionStaticName(lhs)
	if !pKnown || (pName != "extend" && pName != "addEqualityTesters") {
		return false
	}
	if !isPureNonCallingKillRHS(bin.Right) {
		return false
	}
	prevHit := analysis.sccHitActive
	analysis.sccHitActive = false
	rhsKind := analysis.expectConfigMember(bin.Right)
	rhsHitActive := analysis.sccHitActive
	analysis.sccHitActive = prevHit || rhsHitActive
	if rhsKind != expectConfigNone || rhsHitActive {
		return false
	}
	for _, ref := range analysis.ctx.Refs.References(sym) {
		if ref == nil {
			continue
		}
		if ref.Parent == lhs {
			continue
		}
		if ref.Pos() < killStmt.End() {
			return false
		}
		if ref.Parent != nil && ast.IsAccessExpression(ref.Parent) && ref.Parent.Expression() == ref {
			if internalUtils.IsWriteReference(ref.Parent) {
				return false
			}
		} else {
			return false
		}
	}
	killed[pName] = true
	return killed[propName]
}

func (analysis *RstestCallAnalysis) destructuredExpectConfigMember(declaration *ast.Node) expectConfigKind {
	binding := declaration.AsBindingElement()
	if binding == nil || binding.Name() == nil {
		return expectConfigNone
	}
	pattern := declaration.Parent
	if pattern == nil {
		return expectConfigNone
	}
	if pattern.Kind == ast.KindArrayBindingPattern {
		container := analysis.expectConfigBindingPatternKind(pattern)
		if container&expectConfigReceiverArray == 0 {
			return expectConfigNone
		}
		if binding.DotDotDotToken != nil {
			return expectConfigReceiverArray
		}
		if bp := pattern.AsBindingPattern(); bp != nil && bp.Elements != nil && len(bp.Elements.Nodes) > 0 && bp.Elements.Nodes[0] == declaration {
			return expectConfigReceiver
		}
		return expectConfigNone
	}
	if pattern.Kind != ast.KindObjectBindingPattern {
		return expectConfigNone
	}
	container := analysis.expectConfigBindingPatternKind(pattern)
	if container == expectConfigNone {
		return expectConfigNone
	}
	if binding.DotDotDotToken != nil {
		// Rest copy preserves container capabilities (`{ ...copy } = r`,
		// `{ ...copy } = expect`, `{ ...copy } = args`, `{ ...m } = import.meta`,
		// `{ ...g } = globalThis`).
		return container & (expectConfigReceiver | expectConfigReceiverArray | expectConfigExpect | expectConfigImportMeta | expectConfigGlobalObject)
	}
	member, known, staticKey := analysis.evalBindingElementPropertyName(binding)
	var out expectConfigKind
	if container&expectConfigGlobalObject != 0 {
		if known && member == "expect" {
			out |= expectConfigExpect
		}
	}
	if container&expectConfigImportMeta != 0 {
		if known && member == "rstest" {
			out |= expectConfigReceiver
		}
	}
	if container&expectConfigReceiverArray != 0 {
		if (known && member == "0") || !staticKey {
			out |= expectConfigReceiver
		}
	}
	if container&expectConfigReceiver != 0 {
		if known {
			if member == "expect" {
				out |= expectConfigExpect
			}
		} else if !staticKey {
			out |= expectConfigExpect
		}
	}
	if container&expectConfigExpect != 0 {
		if known {
			switch member {
			case "extend":
				out |= expectConfigExtend
			case "addEqualityTesters":
				out |= expectConfigEquality
			}
		} else if !staticKey {
			out |= expectConfigExtend | expectConfigEquality
		}
	}
	return out
}

func (analysis *RstestCallAnalysis) expectConfigBindingPatternKind(pattern *ast.Node) expectConfigKind {
	if pattern == nil || pattern.Parent == nil {
		return expectConfigNone
	}
	switch pattern.Parent.Kind {
	case ast.KindVariableDeclaration:
		return analysis.expectConfigMember(pattern.Parent.AsVariableDeclaration().Initializer)
	case ast.KindBindingElement:
		k := analysis.destructuredExpectConfigMember(pattern.Parent)
		if be := pattern.Parent.AsBindingElement(); be != nil && be.Initializer != nil {
			k |= analysis.expectConfigMember(be.Initializer)
		}
		return k
	case ast.KindParameter:
		var out expectConfigKind
		pd := pattern.Parent.AsParameterDeclaration()
		if analysis.callbacksRef().ContextParameters[pattern.Parent] {
			if pd != nil && pd.DotDotDotToken != nil {
				out |= expectConfigReceiverArray
			} else {
				out |= expectConfigReceiver
			}
		}
		if pd != nil && pd.Initializer != nil {
			out |= analysis.expectConfigMember(pd.Initializer)
		}
		return out
	default:
		return expectConfigNone
	}
}

func (analysis *RstestCallAnalysis) evalExpectConfigAccessName(access *ast.Node) (name string, known bool, staticKey bool) {
	if access == nil {
		return "", false, false
	}
	if staticName, ok := internalUtils.AccessExpressionStaticName(access); ok {
		return staticName, true, true
	}
	if access.Kind == ast.KindElementAccessExpression {
		arg := access.AsElementAccessExpression().ArgumentExpression
		evaluator := analysis.staticStringEvaluator()
		if folded, ok := evaluator.EvalToString(arg); ok {
			return folded, true, true
		}
		return "", false, evaluator.HasStaticAccessExpressionKey(access)
	}
	return "", false, false
}

func (analysis *RstestCallAnalysis) evalBindingElementPropertyName(binding *ast.BindingElement) (string, bool, bool) {
	if binding.PropertyName == nil {
		if binding.Name().Kind == ast.KindIdentifier {
			return binding.Name().AsIdentifier().Text, true, true
		}
		return "", false, false
	}
	return analysis.evalPropertyNameKey(binding.PropertyName)
}

func (analysis *RstestCallAnalysis) evalPropertyNameKey(propName *ast.Node) (string, bool, bool) {
	if name, ok := internalUtils.GetStaticPropertyName(propName); ok {
		return name, true, true
	}
	if propName.Kind == ast.KindComputedPropertyName {
		expression := propName.AsComputedPropertyName().Expression
		evaluator := analysis.staticStringEvaluator()
		if name, ok := evaluator.EvalToString(expression); ok {
			return name, true, true
		}
		return "", false, evaluator.HasStaticComputedPropertyKey(propName)
	}
	return "", false, false
}

func (analysis *RstestCallAnalysis) HasImportMetaRstestWrites() bool {
	return analysis.hasImportMetaRstestWrites
}

func (analysis *RstestCallAnalysis) Callbacks() RstestTestCallbacks {
	return *analysis.callbacksRef()
}

// HookCallbacks returns every function that a lifecycle hook registration
// runs as its callback, including one passed by name.
func (analysis *RstestCallAnalysis) HookCallbacks() map[*ast.Node]bool {
	if !analysis.hookCallbacksOK {
		analysis.hookCallbacks = collectRstestHookCallbacks(analysis)
		analysis.hookCallbacksOK = true
	}
	return analysis.hookCallbacks
}

func (analysis *RstestCallAnalysis) callbacksRef() *RstestTestCallbacks {
	if !analysis.callbacksOK {
		if analysis.hasTests {
			analysis.callbacks = collectRstestTestCallbacks(analysis)
		} else {
			analysis.callbacks = newRstestTestCallbacks()
		}
		analysis.callbacksOK = true
	}
	return &analysis.callbacks
}

func (analysis *RstestCallAnalysis) callbackOwnership() map[*ast.Node][]rstestCallbackRegistration {
	if !analysis.ownershipOK {
		analysis.ownership = collectRstestCallbackOwnership(analysis)
		analysis.ownershipOK = true
	}
	return analysis.ownership
}

// RegistrationCallbacks returns every function that a test or describe
// registration invokes as its callback. The ownership index is already cached
// by file; callers receive only the function identities they need.
func (analysis *RstestCallAnalysis) RegistrationCallbacks() map[*ast.Node]bool {
	if analysis.registrationCallbacksOK {
		return analysis.registrationCallbacks
	}
	callbacks := make(map[*ast.Node]bool, len(analysis.callbackOwnership()))
	for function := range analysis.callbackOwnership() {
		callbacks[function] = true
	}
	analysis.registrationCallbacks = callbacks
	analysis.registrationCallbacksOK = true
	return analysis.registrationCallbacks
}

// isFnCallCandidate reports whether syntax and local aliases permit any Rstest
// registration kind at node.
func (analysis *RstestCallAnalysis) isFnCallCandidate(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	root := testFramework.ResolveFirstIdentifier(node.AsCallExpression().Expression)
	return root == nil ||
		root.Kind != ast.KindIdentifier ||
		analysis.candidates[root.AsIdentifier().Text]&rstestCandidateFn != 0
}

// isTestCandidate reports whether syntax and local aliases permit a Rstest test
// registration at node.
func (analysis *RstestCallAnalysis) isTestCandidate(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	root := testFramework.ResolveFirstIdentifier(node.AsCallExpression().Expression)
	return root == nil ||
		root.Kind != ast.KindIdentifier ||
		analysis.candidates[root.AsIdentifier().Text]&rstestCandidateTest != 0
}

// isExpectCandidate reports whether syntax and local aliases permit a Rstest
// expect root at node.
func (analysis *RstestCallAnalysis) isExpectCandidate(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindCallExpression {
		return false
	}
	root := testFramework.ResolveFirstIdentifier(node.AsCallExpression().Expression)
	return root == nil ||
		root.Kind != ast.KindIdentifier ||
		analysis.candidates[root.AsIdentifier().Text]&rstestCandidateExpect != 0
}

func (analysis *RstestCallAnalysis) addExpectRootName(name string) {
	if name != "" {
		analysis.candidates[name] |= rstestCandidateExpect
	}
}

type rstestAliasCandidate struct {
	localName string
	rootName  string
}

type rstestCandidateKind uint8

const (
	rstestCandidateFn rstestCandidateKind = 1 << iota
	rstestCandidateTest
	rstestCandidateExpect
	rstestCandidateAssert
)

const rstestCandidateAll = rstestCandidateFn | rstestCandidateTest |
	rstestCandidateExpect | rstestCandidateAssert

func cloneRstestCandidateSeeds() map[string]rstestCandidateKind {
	candidates := make(
		map[string]rstestCandidateKind,
		len(rstestDirectAPIStates)+1,
	)
	for name, states := range rstestDirectAPIStates {
		kind := rstestCandidateFn
		for _, state := range states {
			switch state {
			case rstestAPITest, rstestAPITestWithExtend, rstestAPIParameterizedTest:
				kind |= rstestCandidateTest
			}
		}
		candidates[name] = kind
	}
	candidates["expect"] = rstestCandidateExpect
	candidates[rstestAssertAPIName] = rstestCandidateAssert
	return candidates
}

func (analysis *RstestCallAnalysis) indexSourceFile() {
	if analysis.ctx.SourceFile == nil {
		return
	}
	aliases := make([]rstestAliasCandidate, 0)
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node.Kind == ast.KindIdentifier && internalUtils.IsWriteReference(node) {
			if name := node.AsIdentifier().Text; name == "expect" || name == rstestAssertAPIName {
				written := &analysis.globalExpectWritten
				if name == rstestAssertAPIName {
					written = &analysis.globalAssertWritten
				}
				if !*written {
					if analysis.ctx.Refs != nil {
						*written = analysis.ctx.Refs.IsGlobalReference(node)
					} else {
						// Preserve standalone parser-test behavior for manually
						// assembled contexts. Normal lint runs always provide
						// RefStore.
						*written = !internalUtils.IsShadowed(node, name)
					}
				}
			}
		}
		switch node.Kind {
		case ast.KindImportDeclaration:
			analysis.collectImportCandidates(node.AsImportDeclaration())
		case ast.KindImportEqualsDeclaration:
			analysis.collectImportEqualsCandidates(node)
		case ast.KindVariableDeclaration:
			analysis.collectVariableCandidates(node, &aliases)
			analysis.recordFunction(node)
		case ast.KindFunctionDeclaration:
			analysis.recordFunction(node)
		case ast.KindCallExpression:
			analysis.calls = append(analysis.calls, node)
		case ast.KindMetaProperty:
			if !analysis.hasImportMetaRstestWrites && isImportMeta(node) {
				reference := node
				for reference.Parent != nil && internalUtils.SkipAssertionsAndParens(reference.Parent) == node {
					reference = reference.Parent
				}
				access := reference.Parent
				if access != nil && ast.IsAccessExpression(access) && access.Expression() == reference {
					name, known := internalUtils.AccessExpressionStaticName(access)
					if !known || name == "rstest" {
						reference = access
						for reference.Parent != nil && (internalUtils.SkipAssertionsAndParens(reference.Parent) == internalUtils.SkipAssertionsAndParens(reference) ||
							(ast.IsAccessExpression(reference.Parent) && reference.Parent.Expression() == reference)) {
							reference = reference.Parent
						}
						analysis.hasImportMetaRstestWrites = internalUtils.IsWriteReference(reference) ||
							(reference.Parent != nil && reference.Parent.Kind == ast.KindDeleteExpression)
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(analysis.ctx.SourceFile.Node.AsNode())

	for changed := true; changed; {
		changed = false
		for _, alias := range aliases {
			kind := analysis.candidates[alias.localName] |
				analysis.candidates[alias.rootName]
			if kind != analysis.candidates[alias.localName] {
				analysis.candidates[alias.localName] = kind
				changed = true
			}
		}
	}
	for _, call := range analysis.calls {
		if analysis.isTestCandidate(call) {
			analysis.hasTests = true
			break
		}
	}
}

func (analysis *RstestCallAnalysis) collectImportCandidates(
	declaration *ast.ImportDeclaration,
) {
	if declaration == nil ||
		declaration.ModuleSpecifier == nil ||
		!isRstestImportModule(declaration.ModuleSpecifier.Text()) ||
		declaration.ImportClause == nil ||
		declaration.ImportClause.IsTypeOnly() {
		return
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil || clause.NamedBindings == nil {
		return
	}
	switch clause.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		namespace := clause.NamedBindings.AsNamespaceImport()
		if namespace != nil && namespace.Name() != nil {
			analysis.candidates[namespace.Name().Text()] |= rstestCandidateAll
		}
	case ast.KindNamedImports:
		named := clause.NamedBindings.AsNamedImports()
		if named == nil || named.Elements == nil {
			return
		}
		for _, element := range named.Elements.Nodes {
			specifier := element.AsImportSpecifier()
			if specifier == nil || specifier.IsTypeOnly || specifier.Name() == nil {
				continue
			}
			importedName := specifier.Name().Text()
			if specifier.PropertyName != nil {
				importedName = specifier.PropertyName.Text()
			}
			analysis.candidates[specifier.Name().Text()] |=
				rstestImportedCandidateKind(importedName)
		}
	}
}

func (analysis *RstestCallAnalysis) collectImportEqualsCandidates(node *ast.Node) {
	if node == nil || node.Kind != ast.KindImportEqualsDeclaration {
		return
	}
	declaration := node.AsImportEqualsDeclaration()
	if declaration == nil || declaration.IsTypeOnly || declaration.Name() == nil {
		return
	}
	if testFramework.IsImportEqualsOfModules(node, RstestAllImportModules) {
		analysis.candidates[declaration.Name().Text()] |= rstestCandidateAll
	}
}

func (analysis *RstestCallAnalysis) collectVariableCandidates(
	node *ast.Node,
	aliases *[]rstestAliasCandidate,
) {
	if node == nil || node.Kind != ast.KindVariableDeclaration {
		return
	}
	declaration := node.AsVariableDeclaration()
	if declaration == nil || declaration.Name() == nil || declaration.Initializer == nil {
		return
	}
	name := declaration.Name()
	initializer := internalUtils.SkipAssertionsAndParens(declaration.Initializer)
	if name.Kind == ast.KindObjectBindingPattern &&
		(isRstestRequireCall(initializer) || isImportMetaRstest(initializer)) {
		pattern := name.AsBindingPattern()
		if pattern == nil || pattern.Elements == nil {
			return
		}
		for _, element := range pattern.Elements.Nodes {
			binding := element.AsBindingElement()
			if binding == nil || binding.Name() == nil || binding.Name().Kind != ast.KindIdentifier {
				continue
			}
			importedName := binding.Name().Text()
			if binding.PropertyName != nil {
				// A computed key carries no text of its own, so the name has to
				// be folded from the expression. `{ ['expect']: check }` names
				// the same API as `{ expect: check }`.
				resolved, known := internalUtils.GetStaticPropertyName(binding.PropertyName)
				if !known {
					// A genuinely dynamic key binds one Rstest API without
					// saying which. Narrowing it to no candidate would make
					// every rule treat the local as foreign — an assertion
					// through it would read as no assertion at all. This is the
					// same unknown a namespace import leaves behind, so it gets
					// the same answer.
					analysis.candidates[binding.Name().Text()] |= rstestCandidateAll
					continue
				}
				importedName = resolved
			}
			analysis.candidates[binding.Name().Text()] |=
				rstestImportedCandidateKind(importedName)
		}
		return
	}
	if name.Kind != ast.KindIdentifier {
		return
	}
	localName := name.AsIdentifier().Text
	isConst := node.Parent != nil &&
		node.Parent.Kind == ast.KindVariableDeclarationList &&
		node.Parent.Flags&ast.NodeFlagsConst != 0
	if isRstestRequireCall(initializer) || isImportMetaRstest(initializer) {
		if isConst {
			analysis.candidates[localName] |= rstestCandidateAll
		} else {
			analysis.candidates[localName] |= rstestCandidateExpect | rstestCandidateAssert
		}
		return
	}
	if !isConst {
		return
	}
	root := testFramework.ResolveFirstIdentifier(initializer)
	if root != nil && root.Kind == ast.KindIdentifier {
		*aliases = append(*aliases, rstestAliasCandidate{
			localName: localName,
			rootName:  root.AsIdentifier().Text,
		})
	}
}

func (analysis *RstestCallAnalysis) recordFunction(node *ast.Node) {
	name := ""
	var function *ast.Node
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		declaration := node.AsFunctionDeclaration()
		if declaration != nil && declaration.Name() != nil {
			name = declaration.Name().Text()
			function = node
		}
	case ast.KindVariableDeclaration:
		declaration := node.AsVariableDeclaration()
		if declaration != nil &&
			declaration.Name() != nil &&
			declaration.Name().Kind == ast.KindIdentifier &&
			declaration.Initializer != nil {
			initializer := internalUtils.SkipAssertionsAndParens(declaration.Initializer)
			if ast.IsFunctionExpressionOrArrowFunction(initializer) {
				name = declaration.Name().AsIdentifier().Text
				function = initializer
			}
		}
	}
	if function == nil {
		return
	}
	entry, seen := analysis.functions[name]
	if !seen {
		analysis.functions[name] = rstestFunctionEntry{node: function}
		return
	}
	// A redeclared name has no single answer, and the index cannot tell which
	// declaration a call site meant. Recording the collision lets the lookup
	// give up rather than guess the first one.
	entry.ambiguous = true
	analysis.functions[name] = entry
}

func isRstestRequireCall(node *ast.Node) bool {
	return testFramework.IsModuleRequireCallModules(node, RstestAllImportModules)
}

func isRstestImportModule(name string) bool {
	return IsRstestCoreImportModule(name) || name == RstestPlaywrightImportModule
}

func isRstestRegistrationName(name string) bool {
	return directRstestAPIState(rstestProfileCore, name) != rstestAPIInvalid ||
		directRstestAPIState(rstestProfilePlaywright, name) != rstestAPIInvalid
}

func rstestImportedCandidateKind(name string) rstestCandidateKind {
	kind := rstestCandidateKind(0)
	if isRstestRegistrationName(name) {
		kind |= rstestCandidateFn
	}
	if name == "test" || name == "it" {
		kind |= rstestCandidateTest
	}
	if name == "expect" {
		kind |= rstestCandidateExpect
	}
	if name == rstestAssertAPIName {
		kind |= rstestCandidateAssert
	}
	return kind
}
