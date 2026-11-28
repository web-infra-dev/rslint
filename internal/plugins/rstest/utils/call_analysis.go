package utils

import (
	"slices"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

type boundCallPostKey struct {
	sym         *ast.Symbol
	block       *ast.Node
	lastKillPos int
}

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
	// Expect customization is computed lazily from calls because only rules
	// whose safety depends on the built-in matcher implementations need it. The
	// source-file call index is already shared by every Rstest rule, so this never
	// adds another AST walk.
	overriddenExpectMatchers       map[string]bool
	recordedMatcherLiterals        map[*ast.Node]bool
	allExpectMatchersOverridden    bool
	hasCustomEqualityTesters       bool
	expectCustomizationOK          bool
	expectConfigAliases            map[*ast.Symbol]expectConfigKind
	expectConfigSymbolInfo         map[*ast.Symbol]*expectConfigSymbolSummary
	expectConfigFnReach            map[*ast.Symbol]*expectConfigFnSummary
	expectConfigMustOverwrite      map[*ast.Symbol]*expectConfigOverwriteSummary
	expectConfigOverwriteVisit     map[*ast.Symbol]bool
	expectConfigOverwriteCycleHits int
	expectConfigVisiting           map[*ast.Symbol]bool
	expectConfigPartial            map[*ast.Symbol]expectConfigKind
	expectConfigSelfHit            map[*ast.Symbol]bool
	expectConfigOtherCycleHits     map[*ast.Symbol]int
	expectConfigCycleHits          int
	boundCallExprCache             map[*ast.Node]boundCallResult
	boundCallConstCache            map[*ast.Symbol]boundCallResult
	boundCallPostAllWritesCache    map[boundCallPostKey]boundCallResult
	objectLiteralMutations         map[*ast.Symbol]*objectLiteralMutationSummary
	objectLiteralMutationVisiting  map[*ast.Symbol]bool
	objectLiteralMutationCycleHits int
	symbolNoExternalWrites         map[*ast.Symbol]bool
	symbolDefinePropertyAccessor   map[*ast.Symbol]bool
	objectLiteralGetterSummaries   map[*ast.Node]*objGetterSummary
	boundCycleHits                 int
	staticStrings                  *internalUtils.StaticStringEvaluator
	hasImportMetaRstestWrites      bool
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
// this file may replace matcher. A dynamic matcher object or spread may replace
// any name, so it conservatively answers true for every matcher.
func (analysis *RstestCallAnalysis) IsExpectMatcherOverridden(matcher string) bool {
	analysis.collectExpectCustomization()
	return analysis.allExpectMatchersOverridden || analysis.overriddenExpectMatchers[matcher]
}

// HasCustomEqualityTesters reports whether this file installs Rstest equality
// testers, which may execute user code from equality-based built-in matchers.
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

type expectConfigKind uint8

const (
	expectConfigNone               expectConfigKind = 0
	expectConfigReceiver           expectConfigKind = 1 << iota // e.g. `core`, `import.meta.rstest`
	expectConfigExpectExtend                                    // `expect` with intact `.extend`
	expectConfigExpectEquality                                  // `expect` with intact `.addEqualityTesters`
	expectConfigExtend                                          // `expect.extend` function
	expectConfigAddEqualityTesters                              // `expect.addEqualityTesters` function
	expectConfigDynamicMember                                   // unknown dynamic property on `expect`
)

const expectConfigExpect = expectConfigExpectExtend | expectConfigExpectEquality

func (analysis *RstestCallAnalysis) collectExpectCustomization() {
	if analysis.expectCustomizationOK {
		return
	}
	analysis.expectCustomizationOK = true
	if analysis.ctx.SourceFile == nil || len(analysis.calls) == 0 {
		return
	}
	analysis.Callbacks()
	for _, call := range analysis.calls {
		parsed := analysis.ParseExpectCall(call)
		kind := expectConfigNone
		var configArgs []*ast.Node
		if parsed != nil && parsed.Entry == RstestExpectEntryStatic {
			switch parsed.Matcher {
			case "extend":
				kind = expectConfigExtend
			case "addEqualityTesters":
				kind = expectConfigAddEqualityTesters
			}
			if kind != expectConfigNone {
				directCall := testFramework.DirectMemberCall(parsed.MatcherEntry)
				if directCall == nil {
					directCall = call
				}
				if args := directCall.Arguments(); len(args) > 0 {
					if args[0] != nil && args[0].Kind == ast.KindSpreadElement {
						configArgs = append(configArgs, nil)
					} else {
						configArgs = append(configArgs, internalUtils.SkipAssertionsAndParens(args[0]))
					}
				}
			}
		}
		if kind == expectConfigNone {
			kind, configArgs = analysis.expectConfigCallTarget(call)
		}
		if len(configArgs) == 0 || kind == expectConfigNone {
			continue
		}
		for _, arg := range configArgs {
			analysis.recordExpectConfigArg(kind, arg)
		}
	}
}

func (analysis *RstestCallAnalysis) recordExpectConfigArg(kind expectConfigKind, arg *ast.Node) {
	if kind&expectConfigAddEqualityTesters != 0 {
		if arg == nil || arg.Kind != ast.KindArrayLiteralExpression ||
			arg.AsArrayLiteralExpression().Elements == nil ||
			len(arg.AsArrayLiteralExpression().Elements.Nodes) > 0 {
			analysis.hasCustomEqualityTesters = true
		}
	}
	if kind&expectConfigExtend != 0 {
		analysis.recordExtendedMatchers(arg)
	}
	if kind&expectConfigDynamicMember != 0 {
		if arg == nil {
			analysis.allExpectMatchersOverridden = true
			analysis.hasCustomEqualityTesters = true
			return
		}
		switch arg.Kind {
		case ast.KindObjectLiteralExpression:
			analysis.recordExtendedMatchers(arg)
		case ast.KindArrayLiteralExpression:
			if arg.AsArrayLiteralExpression().Elements != nil && len(arg.AsArrayLiteralExpression().Elements.Nodes) > 0 {
				analysis.hasCustomEqualityTesters = true
			}
		case ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindStringLiteral,
			ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression,
			ast.KindRegularExpressionLiteral, ast.KindTrueKeyword, ast.KindFalseKeyword,
			ast.KindNullKeyword, ast.KindUndefinedKeyword,
			ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindClassExpression:
			return
		default:
			if _, isString := analysis.staticStringEvaluator().Eval(arg); isString {
				return
			}
			analysis.allExpectMatchersOverridden = true
			analysis.hasCustomEqualityTesters = true
		}
	}
}

func (analysis *RstestCallAnalysis) isGlobalBuiltinMemberCall(callee *ast.Node, objectName, methodName string) bool {
	if callee == nil || !ast.IsAccessExpression(callee) {
		return false
	}
	receiver := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
	if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.AsIdentifier().Text != objectName {
		return false
	}
	if analysis.ctx.Refs != nil {
		if !analysis.ctx.Refs.IsGlobalReference(receiver) {
			return false
		}
	} else if internalUtils.IsShadowed(receiver, objectName) {
		return false
	}
	method, ok := analysis.evalExpectConfigAccessName(callee)
	return ok && method == methodName
}

func firstArgFromApplyArray(argArr *ast.Node) (*ast.Node, bool) {
	argArr = internalUtils.SkipAssertionsAndParens(argArr)
	if argArr != nil && argArr.Kind == ast.KindArrayLiteralExpression {
		if argArr.AsArrayLiteralExpression().Elements == nil {
			return nil, false
		}
		elems := argArr.AsArrayLiteralExpression().Elements.Nodes
		if len(elems) == 0 {
			return nil, false
		}
		if elems[0] != nil && elems[0].Kind == ast.KindSpreadElement {
			return nil, true
		}
		return internalUtils.SkipAssertionsAndParens(elems[0]), true
	}
	return nil, true
}

func hasSpreadBeforeIndex(args []*ast.Node, limit int) bool {
	for i := 0; i < len(args) && i <= limit; i++ {
		if args[i] != nil && args[i].Kind == ast.KindSpreadElement {
			return true
		}
	}
	return false
}

type boundCallResult struct {
	args        []*ast.Node
	argSet      map[*ast.Node]struct{}
	hasOpaque   bool
	alwaysBound bool
	hasCallable bool
}

func (r *boundCallResult) addArg(arg *ast.Node) {
	if arg == nil {
		r.hasOpaque = true
		r.args = nil
		r.argSet = nil
		return
	}
	if r.hasOpaque {
		return
	}
	if r.argSet != nil {
		if _, exists := r.argSet[arg]; exists {
			return
		}
		r.argSet[arg] = struct{}{}
		r.args = append(r.args, arg)
		return
	}
	if slices.Contains(r.args, arg) {
		return
	}
	r.args = append(r.args, arg)
	if len(r.args) >= 8 {
		r.argSet = make(map[*ast.Node]struct{}, len(r.args)*2)
		for _, a := range r.args {
			r.argSet[a] = struct{}{}
		}
	}
}

func (r *boundCallResult) merge(other boundCallResult) {
	if !other.hasCallable {
		return
	}
	if !r.hasCallable {
		r.hasCallable = true
		r.alwaysBound = other.alwaysBound
		r.hasOpaque = other.hasOpaque
		if !other.hasOpaque && len(other.args) > 0 {
			r.args = append([]*ast.Node(nil), other.args...)
			if len(r.args) >= 8 {
				r.argSet = make(map[*ast.Node]struct{}, len(r.args)*2)
				for _, a := range r.args {
					r.argSet[a] = struct{}{}
				}
			}
		}
		return
	}
	r.alwaysBound = r.alwaysBound && other.alwaysBound
	if other.hasOpaque {
		r.hasOpaque = true
		r.args = nil
		r.argSet = nil
		return
	}
	for _, a := range other.args {
		r.addArg(a)
	}
}

func (r *boundCallResult) toArgs() []*ast.Node {
	if r.hasOpaque {
		return []*ast.Node{nil}
	}
	return r.args
}

func (analysis *RstestCallAnalysis) extractBoundCallFirstArgs(
	node *ast.Node,
	depth int,
	visitedWrites map[*expectConfigWrite]bool,
) boundCallResult {
	if node == nil {
		return boundCallResult{}
	}
	if depth > 32 {
		return boundCallResult{hasOpaque: true, alwaysBound: true, hasCallable: true}
	}
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return boundCallResult{}
	}
	if node.Kind != ast.KindIdentifier && len(visitedWrites) == 0 {
		if cached, ok := analysis.boundCallExprCache[node]; ok {
			return cached
		}
	}
	cycleBefore := analysis.boundCycleHits + analysis.expectConfigCycleHits + analysis.expectConfigOverwriteCycleHits + analysis.objectLiteralMutationCycleHits
	var res boundCallResult
	switch node.Kind {
	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin != nil && bin.OperatorToken != nil {
			switch bin.OperatorToken.Kind {
			case ast.KindCommaToken, ast.KindAmpersandAmpersandToken:
				res = analysis.extractBoundCallFirstArgs(bin.Right, depth+1, visitedWrites)
			case ast.KindQuestionQuestionToken, ast.KindBarBarToken:
				res = analysis.combineBoundCallBranches(bin.Left, bin.Right, depth+1, visitedWrites)
			}
		}
	case ast.KindConditionalExpression:
		if cond := node.AsConditionalExpression(); cond != nil {
			res = analysis.combineBoundCallBranches(cond.WhenTrue, cond.WhenFalse, depth+1, visitedWrites)
		}
	case ast.KindCallExpression:
		callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if method, ok := analysis.evalExpectConfigAccessName(callee); ok && method == "bind" {
				// Chained `.bind` prepends inner bound args before outer ones,
				// so whenever the inner callable already binds the first arg,
				// that inner bound arg wins at runtime.
				inner := analysis.extractBoundCallFirstArgs(callee.Expression(), depth+1, visitedWrites)
				if inner.hasCallable && inner.alwaysBound {
					res = inner
					break
				}
				bArgs := node.Arguments()
				res = inner
				res.hasCallable = true
				if hasSpreadBeforeIndex(bArgs, 1) {
					res.addArg(nil)
					res.alwaysBound = true
				} else if len(bArgs) >= 2 {
					res.addArg(internalUtils.SkipAssertionsAndParens(bArgs[1]))
					res.alwaysBound = true
				} else {
					res.alwaysBound = false
				}
			}
		}
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		res = analysis.extractBoundMemberAccessArgs(node, depth+1, visitedWrites)
	case ast.KindIdentifier:
		return analysis.extractBoundIdentifierArgs(node, depth+1, visitedWrites)
	}
	if node.Kind != ast.KindIdentifier && len(visitedWrites) == 0 &&
		analysis.boundCycleHits+analysis.expectConfigCycleHits+analysis.expectConfigOverwriteCycleHits+analysis.objectLiteralMutationCycleHits == cycleBefore {
		if analysis.boundCallExprCache == nil {
			analysis.boundCallExprCache = map[*ast.Node]boundCallResult{}
		}
		analysis.boundCallExprCache[node] = res
	}
	return res
}

func (analysis *RstestCallAnalysis) extractBoundMemberAccessArgs(
	node *ast.Node,
	depth int,
	visitedWrites map[*expectConfigWrite]bool,
) boundCallResult {
	member, ok := analysis.evalExpectConfigAccessName(node)
	if !ok {
		if analysis.expectConfigCallMember(node) != expectConfigNone {
			return boundCallResult{hasCallable: true, alwaysBound: false}
		}
		return boundCallResult{}
	}
	recv := internalUtils.SkipAssertionsAndParens(node.Expression())
	objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(recv)
	if objExpr == nil {
		if analysis.expectConfigCallMember(node) != expectConfigNone {
			return boundCallResult{hasCallable: true, alwaysBound: false}
		}
		return boundCallResult{}
	}
	if mutableObj && recv != nil && recv.Kind == ast.KindIdentifier {
		if sym := resolveRstestRootSymbol(analysis.ctx, recv); sym != nil {
			if replacement, replaced := analysis.unconditionalPropertyReplacementBefore(sym, objExpr, member, node); replaced {
				if analysis.expectConfigCallMember(replacement) == expectConfigNone {
					return boundCallResult{}
				}
				sub := analysis.extractBoundCallFirstArgs(replacement, depth, visitedWrites)
				if !sub.hasCallable {
					return boundCallResult{hasCallable: true, alwaysBound: false}
				}
				return sub
			}
		}
	}
	lookup := analysis.lookupObjectLiteralProperty(objExpr, member, 0)
	var res boundCallResult
	for _, expr := range lookup.possibleExprs {
		if analysis.expectConfigCallMember(expr) == expectConfigNone {
			continue
		}
		sub := analysis.extractBoundCallFirstArgs(expr, depth, visitedWrites)
		if !sub.hasCallable {
			sub = boundCallResult{hasCallable: true, alwaysBound: false}
		}
		res.merge(sub)
	}
	if mutableObj || lookup.mayHaveUnknownSource || !lookup.foundExplicit {
		if analysis.expectConfigCallMember(node) != expectConfigNone {
			res.merge(boundCallResult{hasCallable: true, alwaysBound: false})
		}
	}
	return res
}

type objReplacementKey struct {
	block  *ast.Node
	member string
}

type objPropReplacementEvent struct {
	pos        int
	stmtPos    int
	stmtEnd    int
	node       *ast.Node
	rhs        *ast.Node
	viaMethod  bool
	methodName string
}

type objMutationEvent struct {
	pos     int
	stmtPos int
	stmtEnd int
	fn      *ast.Node
}

type ownMethodAssignKey struct {
	methodName string
	targetProp string
}

type ownMethodAssignResult struct {
	rhs *ast.Node
	ok  bool
}

// objectLiteralMutationSummary caches a single-pass O(R) index of mutations,
// escapes, and unconditional property replacements for a single-write `const`
// object literal binding so repeated member lookups stay O(log R).
type objectLiteralMutationSummary struct {
	escaped                 bool
	hasAnyMutation          bool
	hasAnyGetter            bool
	hasAnyAccessor          bool
	pureExtend              bool
	pureEquality            bool
	replacements            map[objReplacementKey][]objPropReplacementEvent
	mutationsByFn           map[*ast.Node][]objMutationEvent
	methodInvalidationsByFn map[*ast.Node][]objMutationEvent
	propWriteEventsByProp   map[string]map[*ast.Node][]objMutationEvent
	ownMethodCache          map[ownMethodAssignKey]ownMethodAssignResult
	restoredConfigMask      expectConfigKind
}

func (analysis *RstestCallAnalysis) objectLiteralHasAccessor(
	objExpr *ast.Node,
	propName string,
	known bool,
	includeSetters bool,
	depth int,
) bool {
	if depth > 4 || objExpr == nil || objExpr.Kind != ast.KindObjectLiteralExpression ||
		objExpr.AsObjectLiteralExpression().Properties == nil {
		return depth > 4
	}
	var protoExpr *ast.Node
	for _, prop := range objExpr.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		if prop.Kind == ast.KindPropertyAssignment && prop.Name() != nil && prop.Name().Kind != ast.KindComputedPropertyName {
			if name, ok := internalUtils.GetStaticPropertyName(prop.Name()); ok && name == "__proto__" {
				protoExpr = prop.AsPropertyAssignment().Initializer
				continue
			}
		}
		isAccessor := prop.Kind == ast.KindGetAccessor || (includeSetters && prop.Kind == ast.KindSetAccessor)
		if !isAccessor {
			continue
		}
		if !known || prop.Name() == nil {
			return true
		}
		name, ok, staticKey := analysis.evalPropertyNameKey(prop.Name())
		if !ok && !staticKey {
			return true
		}
		if ok && name == propName {
			return true
		}
	}
	if protoExpr != nil {
		unwrapped := internalUtils.SkipAssertionsAndParens(protoExpr)
		if unwrapped != nil && unwrapped.Kind != ast.KindNullKeyword {
			protoObj, mutableProto := analysis.resolveObjectLiteralForBoundLookup(protoExpr)
			if protoObj == nil || mutableProto {
				return true
			}
			if analysis.objectLiteralHasAccessor(protoObj, propName, known, includeSetters, depth+1) {
				return true
			}
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) isPureRstestExpectObject(expr *ast.Node, depth int) bool {
	if depth > 4 || expr == nil {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	if ast.IsAccessExpression(expr) {
		name, ok := analysis.evalExpectConfigAccessName(expr)
		if !ok || name != "expect" {
			return false
		}
		recv := internalUtils.SkipAssertionsAndParens(expr.Expression())
		if recv == nil {
			return false
		}
		if (!analysis.hasImportMetaRstestWrites && isImportMetaRstest(recv)) || isRstestRequireCall(recv) {
			return true
		}
		if recv.Kind == ast.KindIdentifier {
			if sym := resolveRstestRootSymbol(analysis.ctx, recv); sym != nil {
				rootInfo, ok := analysis.expectRoots[sym]
				if !ok {
					rootInfo = classifyRstestExpectRoot(recv.AsIdentifier().Text, recv, sym, analysis)
					analysis.expectRoots[sym] = rootInfo
				}
				return rootInfo.Kind == rstestExpectRootReceiver
			}
		}
		return false
	}
	if expr.Kind == ast.KindIdentifier {
		sym := resolveRstestRootSymbol(analysis.ctx, expr)
		if sym == nil {
			return expr.AsIdentifier().Text == "expect" && !analysis.globalExpectWritten
		}
		rootInfo, ok := analysis.expectRoots[sym]
		if !ok {
			rootInfo = classifyRstestExpectRoot(expr.AsIdentifier().Text, expr, sym, analysis)
			analysis.expectRoots[sym] = rootInfo
		}
		if rootInfo.Kind == rstestExpectRootDirect {
			return true
		}
		summary := analysis.getExpectConfigSymbolSummary(sym)
		if len(summary.writes) == 1 && summary.writes[0].isConstDecl && summary.writes[0].rhs != nil &&
			!nodeContains(summary.writes[0].rhs, expr) {
			return analysis.isPureRstestExpectObject(summary.writes[0].rhs, depth+1)
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) isPureRstestExpectCallable(expr *ast.Node, depth int) bool {
	if depth > 4 || expr == nil {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	if internalUtils.IsCommaOperator(expr) {
		return analysis.isPureRstestExpectCallable(expr.AsBinaryExpression().Right, depth+1)
	}
	if expr.Kind == ast.KindCallExpression {
		callee := internalUtils.SkipAssertionsAndParens(expr.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if method, ok := analysis.evalExpectConfigAccessName(callee); ok && method == "bind" {
				return analysis.isPureRstestExpectCallable(callee.Expression(), depth+1)
			}
		}
		return false
	}
	if ast.IsAccessExpression(expr) {
		member, ok := analysis.evalExpectConfigAccessName(expr)
		if !ok || (member != "extend" && member != "addEqualityTesters") {
			return false
		}
		return analysis.isPureRstestExpectObject(expr.Expression(), depth+1)
	}
	if expr.Kind == ast.KindIdentifier {
		sym := resolveRstestRootSymbol(analysis.ctx, expr)
		if sym == nil {
			return false
		}
		summary := analysis.getExpectConfigSymbolSummary(sym)
		if len(summary.writes) == 1 && summary.writes[0].isConstDecl && summary.writes[0].rhs != nil &&
			!nodeContains(summary.writes[0].rhs, expr) {
			return analysis.isPureRstestExpectCallable(summary.writes[0].rhs, depth+1)
		}
	}
	return false
}

func isSideEffectFreeNoOpFn(expr *ast.Node) bool {
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil || (expr.Kind != ast.KindArrowFunction && expr.Kind != ast.KindFunctionExpression && expr.Kind != ast.KindMethodDeclaration) {
		return false
	}
	for _, p := range expr.Parameters() {
		if p == nil {
			continue
		}
		pd := p.AsParameterDeclaration()
		if pd == nil || pd.Initializer != nil || pd.Name() == nil || pd.Name().Kind != ast.KindIdentifier {
			return false
		}
	}
	body := expr.Body()
	if body == nil || body.Kind != ast.KindBlock {
		return false
	}
	stmts := body.AsBlock().Statements
	return stmts == nil || len(stmts.Nodes) == 0
}

func (analysis *RstestCallAnalysis) isPureOwnRstestExpectCallableProp(objExpr *ast.Node, propName string) bool {
	if propName != "extend" && propName != "addEqualityTesters" {
		return false
	}
	lookup := analysis.lookupObjectLiteralProperty(objExpr, propName, 0)
	return lookup.foundExplicit && !lookup.mayHaveUnknownSource && lookup.exactExpr != nil &&
		(analysis.isPureRstestExpectCallable(lookup.exactExpr, 0) || isSideEffectFreeNoOpFn(lookup.exactExpr))
}

// isPureThisArgInvocation reports whether `refOuter` is passed solely as the
// `thisArg` of `Reflect.apply(fn, sym, args)` (which invokes `[[Call]]`
// directly without property lookup on `fn`), where `fn` is either a pure expect
// callable/no-op (`isPureDirect`) or a static member access on `sym` itself
// (`isMemberOnSym`, `memberName`).
func (analysis *RstestCallAnalysis) isPureThisArgInvocation(
	call *ast.Node,
	refOuter *ast.Node,
	sym *ast.Symbol,
) (isPureDirect bool, isMemberOnSym bool, memberName string) {
	if call == nil || call.Kind != ast.KindCallExpression || refOuter == nil || sym == nil {
		return false, false, ""
	}
	args := call.Arguments()
	callee := internalUtils.SkipAssertionsAndParens(call.Expression())
	if callee == nil || !analysis.isGlobalBuiltinMemberCall(callee, "Reflect", "apply") {
		return false, false, ""
	}
	if len(args) < 2 || args[0] == nil || args[0].Kind == ast.KindSpreadElement || args[1] != refOuter {
		return false, false, ""
	}
	targetFn := internalUtils.SkipAssertionsAndParens(args[0])
	if targetFn == nil {
		return false, false, ""
	}
	if analysis.isPureRstestExpectCallable(targetFn, 0) || isSideEffectFreeNoOpFn(targetFn) {
		return true, false, ""
	}
	if ast.IsAccessExpression(targetFn) {
		recv := internalUtils.SkipAssertionsAndParens(targetFn.Expression())
		if recv != nil && recv.Kind == ast.KindIdentifier && resolveRstestRootSymbol(analysis.ctx, recv) == sym {
			if mName, ok := analysis.evalExpectConfigAccessName(targetFn); ok {
				return false, true, mName
			}
		}
	}
	return false, false, ""
}

func (analysis *RstestCallAnalysis) exprDependsOnMethodScope(expr *ast.Node, fnNode *ast.Node) bool {
	if expr == nil || fnNode == nil {
		return true
	}
	paramNames := map[string]bool{}
	var collectParamNames func(*ast.Node)
	collectParamNames = func(n *ast.Node) {
		if n == nil {
			return
		}
		switch n.Kind {
		case ast.KindIdentifier:
			paramNames[n.AsIdentifier().Text] = true
		case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
			if bp := n.AsBindingPattern(); bp != nil && bp.Elements != nil {
				for _, el := range bp.Elements.Nodes {
					if el != nil && el.Kind == ast.KindBindingElement {
						collectParamNames(el.Name())
					}
				}
			}
		}
	}
	for _, p := range fnNode.Parameters() {
		if p != nil {
			collectParamNames(p.Name())
		}
	}

	depends := false
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || depends || ast.IsTypeNode(node) || ast.IsPartOfTypeQuery(node) {
			return
		}
		switch node.Kind {
		case ast.KindThisKeyword, ast.KindSuperKeyword:
			depends = true
			return
		case ast.KindIdentifier:
			// Ignore property-access member names (`r.expect`), property keys, etc.
			if parent := node.Parent; parent != nil {
				switch parent.Kind {
				case ast.KindPropertyAccessExpression:
					if parent.AsPropertyAccessExpression().Name() == node {
						return
					}
				case ast.KindPropertyAssignment, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
					if parent.Name() == node {
						return
					}
				}
			}
			text := node.AsIdentifier().Text
			if text == "arguments" || text == "eval" {
				depends = true
				return
			}
			sym := resolveRstestRootSymbol(analysis.ctx, node)
			if sym != nil {
				for _, d := range sym.Declarations {
					if d != nil && nodeContains(fnNode, d) {
						depends = true
						return
					}
				}
			} else if paramNames[text] {
				depends = true
				return
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return depends
		})
	}
	visit(expr)
	return depends
}

func functionNodeAsteriskToken(fn *ast.Node) *ast.Node {
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

func isPureSimpleAccessKey(lhs *ast.Node) bool {
	if lhs == nil {
		return false
	}
	if lhs.Kind == ast.KindPropertyAccessExpression {
		return true
	}
	if lhs.Kind == ast.KindElementAccessExpression {
		arg := internalUtils.SkipAssertionsAndParens(lhs.AsElementAccessExpression().ArgumentExpression)
		if arg == nil {
			return false
		}
		switch arg.Kind {
		case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
			ast.KindNumericLiteral, ast.KindIdentifier:
			return true
		}
	}
	return false
}

// ownMethodUnconditionalThisAssign reports whether `objExpr` defines a
// synchronous non-generator own method `methodName` whose parameters have no
// defaults/destructuring and whose entire body consists solely of straight-line
// `this.<targetProp> = <rhs>` assignments with independent `<rhs>` and pure keys.
func (analysis *RstestCallAnalysis) ownMethodUnconditionalThisAssign(objExpr *ast.Node, methodName, targetProp string) (*ast.Node, bool) {
	if objExpr == nil || objExpr.Kind != ast.KindObjectLiteralExpression || objExpr.AsObjectLiteralExpression().Properties == nil {
		return nil, false
	}
	var fnNode *ast.Node
	for _, prop := range objExpr.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		if prop.Kind == ast.KindSpreadAssignment || prop.Name() == nil {
			fnNode = nil
			continue
		}
		name, ok, staticKey := analysis.evalPropertyNameKey(prop.Name())
		if !ok {
			if !staticKey {
				fnNode = nil
			}
			continue
		}
		// Non-computed `__proto__` sets the prototype rather than an own method.
		if prop.Kind == ast.KindPropertyAssignment && name == "__proto__" && prop.Name().Kind != ast.KindComputedPropertyName {
			continue
		}
		if name != methodName {
			continue
		}
		switch prop.Kind {
		case ast.KindMethodDeclaration:
			fnNode = prop
		case ast.KindPropertyAssignment:
			init := internalUtils.SkipAssertionsAndParens(prop.AsPropertyAssignment().Initializer)
			if init != nil && init.Kind == ast.KindFunctionExpression {
				fnNode = init
			} else {
				fnNode = nil
			}
		default:
			fnNode = nil
		}
	}
	if fnNode == nil || fnNode.Body() == nil || fnNode.Body().Kind != ast.KindBlock {
		return nil, false
	}
	if functionNodeAsteriskToken(fnNode) != nil || ast.HasSyntacticModifier(fnNode, ast.ModifierFlagsAsync) {
		return nil, false
	}
	// Reject methods with parameter defaults or destructuring patterns because
	// parameter evaluation runs synchronously with `this === receiver`.
	for _, param := range fnNode.Parameters() {
		if param == nil {
			continue
		}
		p := param.AsParameterDeclaration()
		if p == nil || p.Initializer != nil || p.Name() == nil || p.Name().Kind != ast.KindIdentifier {
			return nil, false
		}
	}
	block := fnNode.Body().AsBlock()
	if block == nil || block.Statements == nil {
		return nil, false
	}
	var assignedRHS *ast.Node
	for _, stmt := range block.Statements.Nodes {
		if stmt == nil {
			continue
		}
		if stmt.Kind != ast.KindExpressionStatement {
			return nil, false
		}
		expr := ast.SkipParentheses(stmt.AsExpressionStatement().Expression)
		if expr == nil || expr.Kind != ast.KindBinaryExpression {
			return nil, false
		}
		bin := expr.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken ||
			internalUtils.IsDefaultValueInDestructuringAssignment(expr) {
			return nil, false
		}
		lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
		if lhs == nil || !ast.IsAccessExpression(lhs) || !isPureSimpleAccessKey(lhs) {
			return nil, false
		}
		if lhs.Kind == ast.KindElementAccessExpression &&
			analysis.exprDependsOnMethodScope(lhs.AsElementAccessExpression().ArgumentExpression, fnNode) {
			return nil, false
		}
		recv := internalUtils.SkipAssertionsAndParens(lhs.Expression())
		if recv == nil || recv.Kind != ast.KindThisKeyword {
			return nil, false
		}
		pName, ok := analysis.evalExpectConfigAccessName(lhs)
		if !ok || pName != targetProp {
			return nil, false
		}
		if analysis.exprDependsOnMethodScope(bin.Right, fnNode) {
			return nil, false
		}
		assignedRHS = bin.Right
	}
	if assignedRHS == nil {
		return nil, false
	}
	return assignedRHS, true
}

// collectThisAssignedConfigBits walks a method or getter body on an object
// literal and collects any expect-config bits that may be assigned onto `this`
// (including conditional assignments, compound/destructuring writes, local
// `const self = this` aliases, `Object.assign(this, ...)`,
// `this.__defineGetter__`, and nested functions assigned onto `this.*`).
func (analysis *RstestCallAnalysis) collectThisAssignedConfigBits(fnNode *ast.Node) expectConfigKind {
	return analysis.collectThisAssignedConfigBitsDepth(fnNode, 0, nil)
}

func (analysis *RstestCallAnalysis) collectThisAssignedConfigBitsDepth(
	fnNode *ast.Node,
	depth int,
	aliasVisited map[*ast.Symbol]bool,
) expectConfigKind {
	if depth > 4 || fnNode == nil || fnNode.Body() == nil {
		return expectConfigNone
	}
	var bits expectConfigKind
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node != fnNode && testFramework.IsFunction(node) && node.Kind != ast.KindArrowFunction {
			return
		}
		if node.Kind == ast.KindThisKeyword {
			outer := outermostExpressionWrapper(node)
			if outer != nil && outer.Parent != nil {
				switch outer.Parent.Kind {
				case ast.KindVariableDeclaration:
					vd := outer.Parent.AsVariableDeclaration()
					if vd != nil && vd.Initializer == outer && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
						selfSym := outer.Parent.Symbol()
						if selfSym == nil {
							selfSym = resolveRstestRootSymbol(analysis.ctx, vd.Name())
						}
						if selfSym != nil {
							if aliasVisited == nil {
								aliasVisited = map[*ast.Symbol]bool{}
							}
							bits |= analysis.collectAliasRestoredConfigBits(selfSym, aliasVisited)
						}
					}
				case ast.KindBinaryExpression:
					bin := outer.Parent.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == outer {
						if lhs := internalUtils.SkipAssertionsAndParens(bin.Left); lhs != nil && lhs.Kind == ast.KindIdentifier {
							if selfSym := resolveRstestRootSymbol(analysis.ctx, lhs); selfSym != nil {
								if aliasVisited == nil {
									aliasVisited = map[*ast.Symbol]bool{}
								}
								bits |= analysis.collectAliasRestoredConfigBits(selfSym, aliasVisited)
							}
						}
					}
				case ast.KindCallExpression:
					for bit, clears := range analysis.mutatingBuiltinOverwriteChanges(outer.Parent, outer) {
						if !clears {
							bits |= bit
						}
					}
				}
			}
		} else if ast.IsAccessExpression(node) {
			if recv := internalUtils.SkipAssertionsAndParens(node.Expression()); recv != nil && recv.Kind == ast.KindThisKeyword {
				if internalUtils.IsWriteReference(node) {
					pName, ok := analysis.evalExpectConfigAccessName(node)
					outer := outermostExpressionWrapper(node)
					if ok {
						if bit := expectConfigOverwriteBit(pName); bit != expectConfigNone {
							if outer != nil && outer.Parent != nil && outer.Parent.Kind == ast.KindBinaryExpression {
								bin := outer.Parent.AsBinaryExpression()
								if bin != nil && bin.OperatorToken != nil && bin.Left == outer {
									switch bin.OperatorToken.Kind {
									case ast.KindEqualsToken, ast.KindQuestionQuestionEqualsToken,
										ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
										if analysis.overwritePreservesConfigBit(bit, bin.Right) {
											bits |= bit
										}
									default:
										bits |= bit
									}
								}
							} else {
								bits |= bit
							}
						}
					} else if outer != nil && outer.Parent != nil && outer.Parent.Kind == ast.KindBinaryExpression {
						bin := outer.Parent.AsBinaryExpression()
						if bin != nil && bin.Left == outer {
							bits |= analysis.configBitsFromDynamicWriteRHS(bin.Right)
						}
					}
					// A function/helper assigned onto `this.<prop>` (e.g.
					// `this.restore = function () { this.extend = r.expect.extend; }; this.restore();`
					// or `this.restore = restoreFn; this.restore();`) executes with
					// `this` bound to the same object.
					if outer != nil && outer.Parent != nil && outer.Parent.Kind == ast.KindBinaryExpression {
						bin := outer.Parent.AsBinaryExpression()
						if bin != nil && bin.Left == outer && bin.Right != nil {
							bits |= analysis.collectThisBitsFromExpr(bin.Right, depth+1, nil)
						}
					}
				} else if mName, ok := analysis.evalExpectConfigAccessName(node); ok && mName == "__defineGetter__" {
					accessOuter := outermostExpressionWrapper(node)
					if accessOuter != nil && accessOuter.Parent != nil &&
						accessOuter.Parent.Kind == ast.KindCallExpression &&
						accessOuter.Parent.AsCallExpression().Expression == accessOuter {
						args := accessOuter.Parent.Arguments()
						if len(args) >= 2 {
							if propStr, pOK := analysis.staticStringEvaluator().EvalToString(internalUtils.SkipAssertionsAndParens(args[0])); pOK {
								if bit := expectConfigOverwriteBit(propStr); bit != expectConfigNone &&
									analysis.getterPreservesConfigBit(bit, args[1]) {
									bits |= bit
								}
							} else {
								k := analysis.getterExpectConfigKind(internalUtils.SkipAssertionsAndParens(args[1]))
								if k&(expectConfigExtend|expectConfigDynamicMember) != 0 {
									bits |= expectConfigExpectExtend
								}
								if k&(expectConfigAddEqualityTesters|expectConfigDynamicMember) != 0 {
									bits |= expectConfigExpectEquality
								}
								if k&expectConfigExpect != 0 {
									bits |= expectConfigReceiver
								}
							}
						}
					}
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	for _, p := range fnNode.Parameters() {
		visit(p)
	}
	visit(fnNode.Body())
	return bits
}

func (analysis *RstestCallAnalysis) configBitsFromDynamicWriteRHS(rhs *ast.Node) expectConfigKind {
	if rhs == nil {
		return expectConfigNone
	}
	k := analysis.expectConfigMember(rhs)
	var bits expectConfigKind
	if k&expectConfigExtend != 0 {
		bits |= expectConfigExpectExtend
	}
	if k&expectConfigAddEqualityTesters != 0 {
		bits |= expectConfigExpectEquality
	}
	if k&expectConfigDynamicMember != 0 {
		bits |= expectConfigExpect
	}
	if k&expectConfigExpect != 0 {
		bits |= expectConfigReceiver
	}
	return bits
}

func (analysis *RstestCallAnalysis) collectThisBitsFromSymbolPropWrites(
	sym *ast.Symbol,
	propName string,
	propKnown bool,
	depth int,
	visited map[*ast.Symbol]bool,
) expectConfigKind {
	if depth > 4 || sym == nil {
		return expectConfigNone
	}
	var bits expectConfigKind
	// Inspect object literals written to `sym` (covers `let helpers = { install() { ... } }`
	// and `let proto = { restore() { ... } }`, not just `const` literals).
	symSummary := analysis.getExpectConfigSymbolSummary(sym)
	for i := range symSummary.writes {
		if rhs := internalUtils.SkipAssertionsAndParens(symSummary.writes[i].rhs); rhs != nil &&
			rhs.Kind == ast.KindObjectLiteralExpression {
			bits |= analysis.ownObjectLiteralPropThisAssignedBits(rhs, propName, propKnown, depth+1, visited)
		}
	}
	if analysis.ctx.Refs == nil {
		return bits
	}
	for _, ref := range analysis.ctx.Refs.References(sym) {
		if ref == nil {
			continue
		}
		outer := outermostExpressionWrapper(ref)
		if outer == nil || outer.Parent == nil || !ast.IsAccessExpression(outer.Parent) || outer.Parent.Expression() != outer {
			continue
		}
		if !internalUtils.IsWriteReference(outer.Parent) {
			continue
		}
		pName, pKnown := analysis.evalExpectConfigAccessName(outer.Parent)
		if propKnown && pKnown && pName != propName {
			continue
		}
		accessOuter := outermostExpressionWrapper(outer.Parent)
		if accessOuter != nil && accessOuter.Parent != nil && accessOuter.Parent.Kind == ast.KindBinaryExpression {
			bin := accessOuter.Parent.AsBinaryExpression()
			if bin != nil && bin.Left == accessOuter && bin.Right != nil {
				bits |= analysis.collectThisBitsFromExpr(bin.Right, depth+1, visited)
			}
		}
	}
	return bits
}

func (analysis *RstestCallAnalysis) collectThisBitsFromExpr(
	expr *ast.Node,
	depth int,
	visited map[*ast.Symbol]bool,
) expectConfigKind {
	if depth > 4 || expr == nil {
		return expectConfigNone
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return expectConfigNone
	}
	switch expr.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration,
		ast.KindGetAccessor, ast.KindSetAccessor:
		return analysis.collectThisAssignedConfigBits(expr)
	case ast.KindIdentifier:
		sym := resolveRstestRootSymbol(analysis.ctx, expr)
		if sym == nil || visited[sym] {
			return expectConfigNone
		}
		if visited == nil {
			visited = map[*ast.Symbol]bool{}
		}
		visited[sym] = true
		defer delete(visited, sym)
		var bits expectConfigKind
		for _, decl := range sym.Declarations {
			if decl != nil && decl.Kind == ast.KindFunctionDeclaration && decl.AsFunctionDeclaration().Body != nil &&
				analysis.ctx.SourceFile != nil && ast.GetSourceFileOfNode(decl) == analysis.ctx.SourceFile {
				bits |= analysis.collectThisAssignedConfigBits(decl)
			}
		}
		symSummary := analysis.getExpectConfigSymbolSummary(sym)
		for i := range symSummary.writes {
			if rhs := symSummary.writes[i].rhs; rhs != nil && !nodeContains(rhs, expr) {
				bits |= analysis.collectThisBitsFromExpr(rhs, depth+1, visited)
			}
		}
		return bits
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		pName, pKnown := analysis.evalExpectConfigAccessName(expr)
		recv := internalUtils.SkipAssertionsAndParens(expr.Expression())
		if recv == nil {
			return expectConfigNone
		}
		var bits expectConfigKind
		if recvObj, mutable := analysis.resolveObjectLiteralForBoundLookup(recv); recvObj != nil {
			bits |= analysis.ownObjectLiteralPropThisAssignedBits(recvObj, pName, pKnown, depth+1, visited)
			if mutable && recv.Kind == ast.KindIdentifier {
				if recvSym := resolveRstestRootSymbol(analysis.ctx, recv); recvSym != nil && !visited[recvSym] {
					if visited == nil {
						visited = map[*ast.Symbol]bool{}
					}
					visited[recvSym] = true
					bits |= analysis.collectThisBitsFromSymbolPropWrites(recvSym, pName, pKnown, depth+1, visited)
					delete(visited, recvSym)
				}
			}
		} else if recv.Kind == ast.KindIdentifier {
			if recvSym := resolveRstestRootSymbol(analysis.ctx, recv); recvSym != nil && !visited[recvSym] {
				if visited == nil {
					visited = map[*ast.Symbol]bool{}
				}
				visited[recvSym] = true
				bits |= analysis.collectThisBitsFromSymbolPropWrites(recvSym, pName, pKnown, depth+1, visited)
				delete(visited, recvSym)
			}
		}
		return bits
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		if cond != nil {
			return analysis.collectThisBitsFromExpr(cond.WhenTrue, depth+1, visited) |
				analysis.collectThisBitsFromExpr(cond.WhenFalse, depth+1, visited)
		}
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin != nil && bin.OperatorToken != nil {
			switch bin.OperatorToken.Kind {
			case ast.KindCommaToken, ast.KindAmpersandAmpersandToken, ast.KindEqualsToken:
				return analysis.collectThisBitsFromExpr(bin.Right, depth+1, visited)
			case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				return analysis.collectThisBitsFromExpr(bin.Left, depth+1, visited) |
					analysis.collectThisBitsFromExpr(bin.Right, depth+1, visited)
			}
		}
	}
	return expectConfigNone
}

func (analysis *RstestCallAnalysis) ownObjectLiteralPropThisAssignedBits(
	objExpr *ast.Node,
	propName string,
	propKnown bool,
	depth int,
	visited map[*ast.Symbol]bool,
) expectConfigKind {
	if depth > 4 || objExpr == nil || objExpr.Kind != ast.KindObjectLiteralExpression || objExpr.AsObjectLiteralExpression().Properties == nil {
		return expectConfigNone
	}
	var bits expectConfigKind
	for _, prop := range objExpr.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		switch prop.Kind {
		case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
			if propKnown && prop.Name() != nil {
				if name, ok := analysis.staticStringEvaluator().EvalPropertyName(prop.Name()); ok && name != propName {
					continue
				}
			}
			bits |= analysis.collectThisAssignedConfigBits(prop)
		case ast.KindPropertyAssignment:
			if pName := prop.Name(); pName != nil && pName.Kind != ast.KindComputedPropertyName {
				if name, ok := internalUtils.GetStaticPropertyName(pName); ok && name == "__proto__" {
					protoExpr := internalUtils.SkipAssertionsAndParens(prop.AsPropertyAssignment().Initializer)
					if protoObj, mutableProto := analysis.resolveObjectLiteralForBoundLookup(protoExpr); protoObj != nil {
						bits |= analysis.ownObjectLiteralPropThisAssignedBits(protoObj, propName, propKnown, depth+1, visited)
						if mutableProto && protoExpr != nil && protoExpr.Kind == ast.KindIdentifier {
							if protoSym := resolveRstestRootSymbol(analysis.ctx, protoExpr); protoSym != nil && !visited[protoSym] {
								if visited == nil {
									visited = map[*ast.Symbol]bool{}
								}
								visited[protoSym] = true
								bits |= analysis.collectThisBitsFromSymbolPropWrites(protoSym, propName, propKnown, depth+1, visited)
								delete(visited, protoSym)
							}
						}
					} else if protoExpr != nil && protoExpr.Kind == ast.KindIdentifier {
						if protoSym := resolveRstestRootSymbol(analysis.ctx, protoExpr); protoSym != nil && !visited[protoSym] {
							if visited == nil {
								visited = map[*ast.Symbol]bool{}
							}
							visited[protoSym] = true
							bits |= analysis.collectThisBitsFromSymbolPropWrites(protoSym, propName, propKnown, depth+1, visited)
							delete(visited, protoSym)
						}
					}
					continue
				}
			}
			if propKnown && prop.Name() != nil {
				if name, ok := analysis.staticStringEvaluator().EvalPropertyName(prop.Name()); ok && name != propName {
					continue
				}
			}
			bits |= analysis.collectThisBitsFromExpr(prop.AsPropertyAssignment().Initializer, depth+1, visited)
		case ast.KindShorthandPropertyAssignment:
			ident := prop.AsShorthandPropertyAssignment().Name()
			if propKnown && ident != nil && ident.Kind == ast.KindIdentifier && ident.AsIdentifier().Text != propName {
				continue
			}
			bits |= analysis.collectThisBitsFromExpr(ident, depth+1, visited)
		case ast.KindSpreadAssignment:
			spreadExpr := internalUtils.SkipAssertionsAndParens(prop.AsSpreadAssignment().Expression)
			if spreadObj, mutableSpread := analysis.resolveObjectLiteralForBoundLookup(spreadExpr); spreadObj != nil {
				bits |= analysis.ownObjectLiteralPropThisAssignedBits(spreadObj, propName, propKnown, depth+1, visited)
				if mutableSpread && spreadExpr != nil && spreadExpr.Kind == ast.KindIdentifier {
					if spreadSym := resolveRstestRootSymbol(analysis.ctx, spreadExpr); spreadSym != nil && !visited[spreadSym] {
						if visited == nil {
							visited = map[*ast.Symbol]bool{}
						}
						visited[spreadSym] = true
						bits |= analysis.collectThisBitsFromSymbolPropWrites(spreadSym, propName, propKnown, depth+1, visited)
						delete(visited, spreadSym)
					}
				}
			} else if spreadExpr != nil && spreadExpr.Kind == ast.KindIdentifier {
				if spreadSym := resolveRstestRootSymbol(analysis.ctx, spreadExpr); spreadSym != nil && !visited[spreadSym] {
					if visited == nil {
						visited = map[*ast.Symbol]bool{}
					}
					visited[spreadSym] = true
					bits |= analysis.collectThisBitsFromSymbolPropWrites(spreadSym, propName, propKnown, depth+1, visited)
					delete(visited, spreadSym)
				}
			}
		}
	}
	return bits
}

func (analysis *RstestCallAnalysis) collectAliasRestoredConfigBits(sym *ast.Symbol, visited map[*ast.Symbol]bool) expectConfigKind {
	if analysis.ctx.Refs == nil || sym == nil || visited[sym] {
		return expectConfigNone
	}
	visited[sym] = true
	var bits expectConfigKind
	for _, ref := range analysis.ctx.Refs.References(sym) {
		if ref == nil {
			continue
		}
		if parent := ref.Parent; parent != nil &&
			((parent.Kind == ast.KindVariableDeclaration && parent.Name() == ref) ||
				(parent.Kind == ast.KindBindingElement && parent.Name() == ref)) {
			continue
		}
		outer := outermostExpressionWrapper(ref)
		if outer == nil || outer.Parent == nil {
			continue
		}
		switch outer.Parent.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			if outer.Parent.Expression() != outer || !internalUtils.IsWriteReference(outer.Parent) {
				continue
			}
			propName, propKnown := analysis.evalExpectConfigAccessName(outer.Parent)
			accessOuter := outermostExpressionWrapper(outer.Parent)
			if !propKnown {
				if accessOuter != nil && accessOuter.Parent != nil && accessOuter.Parent.Kind == ast.KindBinaryExpression {
					bin := accessOuter.Parent.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.Left == accessOuter {
						bits |= analysis.configBitsFromDynamicWriteRHS(bin.Right)
						continue
					}
				}
				bits |= expectConfigExpect | expectConfigReceiver
				continue
			}
			if bit := expectConfigOverwriteBit(propName); bit != expectConfigNone {
				if accessOuter != nil && accessOuter.Parent != nil && accessOuter.Parent.Kind == ast.KindBinaryExpression {
					bin := accessOuter.Parent.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.Left == accessOuter {
						switch bin.OperatorToken.Kind {
						case ast.KindEqualsToken, ast.KindQuestionQuestionEqualsToken,
							ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
							if analysis.overwritePreservesConfigBit(bit, bin.Right) {
								bits |= bit
							}
						default:
							bits |= bit
						}
						continue
					}
				}
				bits |= bit
			}
		case ast.KindCallExpression:
			for bit, clears := range analysis.mutatingBuiltinOverwriteChanges(outer.Parent, outer) {
				if !clears {
					bits |= bit
				}
			}
		case ast.KindVariableDeclaration:
			vd := outer.Parent.AsVariableDeclaration()
			if vd != nil && vd.Initializer == outer && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
				aliasSym := outer.Parent.Symbol()
				if aliasSym == nil {
					aliasSym = resolveRstestRootSymbol(analysis.ctx, vd.Name())
				}
				bits |= analysis.collectAliasRestoredConfigBits(aliasSym, visited)
			}
		case ast.KindBinaryExpression:
			bin := outer.Parent.AsBinaryExpression()
			if bin != nil && bin.Right == outer && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
				if lhs := internalUtils.SkipAssertionsAndParens(bin.Left); lhs != nil && lhs.Kind == ast.KindIdentifier {
					bits |= analysis.collectAliasRestoredConfigBits(resolveRstestRootSymbol(analysis.ctx, lhs), visited)
				}
			}
		}
	}
	return bits
}

func (analysis *RstestCallAnalysis) cachedOwnMethodThisAssign(
	summary *objectLiteralMutationSummary,
	objExpr *ast.Node,
	methodName, targetProp string,
) (*ast.Node, bool) {
	key := ownMethodAssignKey{methodName: methodName, targetProp: targetProp}
	if summary != nil && summary.ownMethodCache != nil {
		if res, ok := summary.ownMethodCache[key]; ok {
			return res.rhs, res.ok
		}
	}
	rhs, ok := analysis.ownMethodUnconditionalThisAssign(objExpr, methodName, targetProp)
	if summary != nil {
		if summary.ownMethodCache == nil {
			summary.ownMethodCache = map[ownMethodAssignKey]ownMethodAssignResult{}
		}
		summary.ownMethodCache[key] = ownMethodAssignResult{rhs: rhs, ok: ok}
	}
	return rhs, ok
}

func (analysis *RstestCallAnalysis) getObjectLiteralMutationSummary(
	sym *ast.Symbol,
	objExpr *ast.Node,
) *objectLiteralMutationSummary {
	if analysis.ctx.Refs == nil || sym == nil || objExpr == nil {
		return &objectLiteralMutationSummary{escaped: true, hasAnyMutation: true, hasAnyGetter: true, hasAnyAccessor: true}
	}
	if cached, ok := analysis.objectLiteralMutations[sym]; ok {
		return cached
	}
	if analysis.objectLiteralMutationVisiting[sym] {
		analysis.objectLiteralMutationCycleHits++
		return &objectLiteralMutationSummary{escaped: true, hasAnyMutation: true, hasAnyGetter: true, hasAnyAccessor: true}
	}
	if analysis.objectLiteralMutationVisiting == nil {
		analysis.objectLiteralMutationVisiting = map[*ast.Symbol]bool{}
	}
	analysis.objectLiteralMutationVisiting[sym] = true
	defer delete(analysis.objectLiteralMutationVisiting, sym)

	cycleBefore := analysis.objectLiteralMutationCycleHits
	expectCycleBefore := analysis.expectConfigCycleHits
	overwriteCycleBefore := analysis.expectConfigOverwriteCycleHits
	refs := analysis.ctx.Refs.References(sym)
	summary := &objectLiteralMutationSummary{}

	initialPureExtend := analysis.isPureOwnRstestExpectCallableProp(objExpr, "extend")
	initialPureEquality := analysis.isPureOwnRstestExpectCallableProp(objExpr, "addEqualityTesters")

	// Pass 1: track which properties are ever written or deleted, which writes
	// could install a non-pure function on `extend`/`addEqualityTesters`,
	// whether the object escapes, and all invoked static property names so
	// `hasUntrustedCall` is evaluated AFTER all writes/deletes are known.
	writtenOrDeletedProps := map[string]bool{}
	nonPureWriteProps := map[string]bool{}
	hasDynamicWriteOrDelete := false
	hasPass1Escape := false
	hasUntrustedCall := false
	var calledStaticProps []string
	seenCalledProp := map[string]bool{}
	for _, ref := range refs {
		if ref == nil {
			continue
		}
		if parent := ref.Parent; parent != nil &&
			((parent.Kind == ast.KindVariableDeclaration && parent.Name() == ref) ||
				(parent.Kind == ast.KindBindingElement && parent.Name() == ref)) {
			continue
		}
		outer := outermostExpressionWrapper(ref)
		if outer == nil || outer.Parent == nil {
			hasPass1Escape = true
			continue
		}
		if !ast.IsAccessExpression(outer.Parent) || outer.Parent.Expression() != outer {
			// Plain object-pattern destructuring and passing `sym` solely as
			// `thisArg` to `Reflect.apply` of a pure callable (or `sym`'s own
			// method) do not unconditionally escape `sym`.
			isSafeUse := false
			switch outer.Parent.Kind {
			case ast.KindVariableDeclaration:
				vd := outer.Parent.AsVariableDeclaration()
				isSafeUse = vd != nil && vd.Initializer == outer && vd.Name() != nil && vd.Name().Kind == ast.KindObjectBindingPattern
			case ast.KindBinaryExpression:
				bin := outer.Parent.AsBinaryExpression()
				if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == outer {
					lhs := ast.SkipParentheses(bin.Left)
					isSafeUse = lhs != nil && lhs.Kind == ast.KindObjectLiteralExpression
				}
			case ast.KindCallExpression:
				isPureDirect, isMemberOnSym, mName := analysis.isPureThisArgInvocation(outer.Parent, outer, sym)
				if isPureDirect {
					isSafeUse = true
				} else if isMemberOnSym {
					isSafeUse = true
					if !seenCalledProp[mName] {
						seenCalledProp[mName] = true
						calledStaticProps = append(calledStaticProps, mName)
					}
				}
			}
			if !isSafeUse {
				hasPass1Escape = true
			}
			continue
		}
		accessOuter := outermostExpressionWrapper(outer.Parent)
		isWrite := internalUtils.IsWriteReference(outer.Parent)
		isDelete := accessOuter != nil && accessOuter.Parent != nil && accessOuter.Parent.Kind == ast.KindDeleteExpression
		if isWrite || isDelete {
			if pName, pKnown := analysis.evalExpectConfigAccessName(outer.Parent); pKnown {
				writtenOrDeletedProps[pName] = true
				isPureAssign := false
				if isWrite && accessOuter != nil && accessOuter.Parent != nil && accessOuter.Parent.Kind == ast.KindBinaryExpression {
					bin := accessOuter.Parent.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken &&
						bin.Left == accessOuter && !internalUtils.IsDefaultValueInDestructuringAssignment(accessOuter.Parent) &&
						(analysis.isPureRstestExpectCallable(bin.Right, 0) || isSideEffectFreeNoOpFn(bin.Right)) {
						isPureAssign = true
					}
				}
				if !isPureAssign {
					nonPureWriteProps[pName] = true
				}
			} else {
				hasDynamicWriteOrDelete = true
			}
			continue
		}
		isCall := accessOuter != nil && accessOuter.Parent != nil &&
			accessOuter.Parent.Kind == ast.KindCallExpression &&
			accessOuter.Parent.AsCallExpression().Expression == accessOuter
		isTagged := accessOuter != nil && accessOuter.Parent != nil &&
			accessOuter.Parent.Kind == ast.KindTaggedTemplateExpression &&
			accessOuter.Parent.AsTaggedTemplateExpression().Tag == accessOuter
		if isCall || isTagged {
			if isCall && !ast.IsOptionalChain(accessOuter.Parent) {
				if pName, pKnown := analysis.evalExpectConfigAccessName(outer.Parent); pKnown {
					if !seenCalledProp[pName] {
						seenCalledProp[pName] = true
						calledStaticProps = append(calledStaticProps, pName)
					}
					continue
				}
			}
			hasUntrustedCall = true
		}
	}
	if !hasUntrustedCall {
		for _, pName := range calledStaticProps {
			if pName == "extend" && initialPureExtend && !nonPureWriteProps["extend"] {
				continue
			}
			if pName == "addEqualityTesters" && initialPureEquality && !nonPureWriteProps["addEqualityTesters"] {
				continue
			}
			trustedReset := false
			if !writtenOrDeletedProps[pName] {
				for _, targetProp := range []string{"extend", "addEqualityTesters", "expect"} {
					if rhs, ok := analysis.cachedOwnMethodThisAssign(summary, objExpr, pName, targetProp); ok &&
						(analysis.isPureRstestExpectCallable(rhs, 0) || isSideEffectFreeNoOpFn(rhs)) {
						trustedReset = true
						break
					}
				}
			}
			if !trustedReset {
				hasUntrustedCall = true
				break
			}
		}
	}

	hasAnyGetter := analysis.objectLiteralHasAccessor(objExpr, "", false, false, 0)
	hasDefineProp := analysis.symbolHasDefineProperty(sym, nil)
	hasAnyAccessor := hasAnyGetter || hasDefineProp || analysis.objectLiteralHasAccessor(objExpr, "", false, true, 0)
	pureExtend := !hasAnyAccessor && !hasPass1Escape && !hasUntrustedCall && !hasDynamicWriteOrDelete && !nonPureWriteProps["extend"] && initialPureExtend
	pureEquality := !hasAnyAccessor && !hasPass1Escape && !hasUntrustedCall && !hasDynamicWriteOrDelete && !nonPureWriteProps["addEqualityTesters"] && initialPureEquality
	summary.hasAnyGetter = hasAnyGetter || hasDefineProp
	summary.hasAnyAccessor = hasAnyAccessor
	summary.pureExtend = pureExtend
	summary.pureEquality = pureEquality

	var allOwnMemberBitsComputed bool
	var allOwnMemberBits expectConfigKind
	addPropMemberBits := func(propName string, propKnown bool) {
		if !propKnown {
			if !allOwnMemberBitsComputed {
				allOwnMemberBitsComputed = true
				allOwnMemberBits = analysis.ownObjectLiteralPropThisAssignedBits(objExpr, "", false, 0, nil)
			}
			summary.restoredConfigMask |= allOwnMemberBits
			return
		}
		summary.restoredConfigMask |= analysis.ownObjectLiteralPropThisAssignedBits(objExpr, propName, true, 0, nil)
	}

	addMutation := func(ref, stmt *ast.Node, refFn *ast.Node, wildcardInvalidation bool, writtenProp string) {
		summary.hasAnyMutation = true
		if stmt == nil {
			summary.escaped = true
			return
		}
		ev := objMutationEvent{
			pos:     ref.Pos(),
			stmtPos: stmt.Pos(),
			stmtEnd: stmt.End(),
			fn:      refFn,
		}
		if summary.mutationsByFn == nil {
			summary.mutationsByFn = map[*ast.Node][]objMutationEvent{}
		}
		summary.mutationsByFn[refFn] = append(summary.mutationsByFn[refFn], ev)
		if wildcardInvalidation {
			if summary.methodInvalidationsByFn == nil {
				summary.methodInvalidationsByFn = map[*ast.Node][]objMutationEvent{}
			}
			summary.methodInvalidationsByFn[refFn] = append(summary.methodInvalidationsByFn[refFn], ev)
		} else if writtenProp != "" {
			if summary.propWriteEventsByProp == nil {
				summary.propWriteEventsByProp = map[string]map[*ast.Node][]objMutationEvent{}
			}
			byFn := summary.propWriteEventsByProp[writtenProp]
			if byFn == nil {
				byFn = map[*ast.Node][]objMutationEvent{}
				summary.propWriteEventsByProp[writtenProp] = byFn
			}
			byFn[refFn] = append(byFn[refFn], ev)
		}
	}

	addReplacement := func(block *ast.Node, member string, ev objPropReplacementEvent) {
		if summary.replacements == nil {
			summary.replacements = map[objReplacementKey][]objPropReplacementEvent{}
		}
		key := objReplacementKey{block: block, member: member}
		summary.replacements[key] = append(summary.replacements[key], ev)
	}

	aliasVisited := map[*ast.Symbol]bool{sym: true}

	for _, ref := range refs {
		if ref == nil {
			continue
		}
		if parent := ref.Parent; parent != nil &&
			((parent.Kind == ast.KindVariableDeclaration && parent.Name() == ref) ||
				(parent.Kind == ast.KindBindingElement && parent.Name() == ref)) {
			continue
		}
		outer := outermostExpressionWrapper(ref)
		if outer == nil || outer.Parent == nil {
			summary.escaped = true
			summary.hasAnyMutation = true
			continue
		}
		stmt, block, unconditional := writeStatementInfo(ref)
		refFn := ast.GetContainingFunction(ref)

		switch outer.Parent.Kind {
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			if outer.Parent.Expression() != outer {
				summary.escaped = true
				summary.hasAnyMutation = true
				continue
			}
			accessOuter := outermostExpressionWrapper(outer.Parent)
			propName, propKnown := analysis.evalExpectConfigAccessName(outer.Parent)
			if internalUtils.IsWriteReference(outer.Parent) {
				hasAccessorOnProp := hasAnyAccessor && (hasDefineProp || analysis.objectLiteralHasAccessor(objExpr, propName, propKnown, true, 0))
				if hasAccessorOnProp {
					// Writing through a setter executes user code with `this === sym`,
					// which can escape `this` into an outer alias or assign `this.extend`.
					summary.escaped = true
					addPropMemberBits(propName, propKnown)
				}
				wildcardInv := !propKnown || hasAccessorOnProp
				writtenProp := ""
				if propKnown {
					writtenProp = propName
				}
				addMutation(ref, stmt, refFn, wildcardInv, writtenProp)
				handledBinWrite := false
				if accessOuter != nil && accessOuter.Parent != nil && accessOuter.Parent.Kind == ast.KindBinaryExpression {
					bin := accessOuter.Parent.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.Left == accessOuter {
						handledBinWrite = true
						switch bin.OperatorToken.Kind {
						case ast.KindEqualsToken:
							if !internalUtils.IsDefaultValueInDestructuringAssignment(accessOuter.Parent) {
								if propKnown {
									if bit := expectConfigOverwriteBit(propName); bit != expectConfigNone &&
										analysis.overwritePreservesConfigBit(bit, bin.Right) {
										summary.restoredConfigMask |= bit
									}
								} else {
									summary.restoredConfigMask |= analysis.configBitsFromDynamicWriteRHS(bin.Right)
								}
								if unconditional && propKnown && stmt != nil && block != nil &&
									stmt.Kind == ast.KindExpressionStatement &&
									ast.SkipParentheses(stmt.AsExpressionStatement().Expression) == accessOuter.Parent &&
									!hasAccessorOnProp {
									addReplacement(block, propName, objPropReplacementEvent{
										pos:       ref.Pos(),
										stmtPos:   stmt.Pos(),
										stmtEnd:   stmt.End(),
										node:      ref,
										rhs:       bin.Right,
										viaMethod: false,
									})
								}
							} else if propKnown {
								if bit := expectConfigOverwriteBit(propName); bit != expectConfigNone {
									summary.restoredConfigMask |= bit
								}
							} else {
								summary.restoredConfigMask |= expectConfigExpect | expectConfigReceiver
							}
						case ast.KindQuestionQuestionEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
							if propKnown {
								if bit := expectConfigOverwriteBit(propName); bit != expectConfigNone &&
									analysis.overwritePreservesConfigBit(bit, bin.Right) {
									summary.restoredConfigMask |= bit
								}
							} else {
								summary.restoredConfigMask |= analysis.configBitsFromDynamicWriteRHS(bin.Right)
							}
						default:
							if propKnown {
								if bit := expectConfigOverwriteBit(propName); bit != expectConfigNone {
									summary.restoredConfigMask |= bit
								}
							} else {
								summary.restoredConfigMask |= expectConfigExpect | expectConfigReceiver
							}
						}
					}
				}
				if !handledBinWrite {
					if propKnown {
						if bit := expectConfigOverwriteBit(propName); bit != expectConfigNone {
							summary.restoredConfigMask |= bit
						}
					} else {
						summary.restoredConfigMask |= expectConfigExpect | expectConfigReceiver
					}
				}
				continue
			}
			if accessOuter != nil && accessOuter.Parent != nil && accessOuter.Parent.Kind == ast.KindDeleteExpression {
				addMutation(ref, stmt, refFn, true, "")
				continue
			}
			isCall := accessOuter != nil && accessOuter.Parent != nil &&
				accessOuter.Parent.Kind == ast.KindCallExpression &&
				accessOuter.Parent.AsCallExpression().Expression == accessOuter
			isTagged := accessOuter != nil && accessOuter.Parent != nil &&
				accessOuter.Parent.Kind == ast.KindTaggedTemplateExpression &&
				accessOuter.Parent.AsTaggedTemplateExpression().Tag == accessOuter
			if isCall || isTagged {
				if isCall && propKnown && ((propName == "extend" && pureExtend) || (propName == "addEqualityTesters" && pureEquality)) {
					continue
				}
				isPureSingleTargetReset := false
				writtenTargetProp := ""
				if isCall && !ast.IsOptionalChain(accessOuter.Parent) && propKnown {
					for _, targetProp := range []string{"extend", "addEqualityTesters", "expect"} {
						rhs, ok := analysis.cachedOwnMethodThisAssign(summary, objExpr, propName, targetProp)
						if !ok {
							continue
						}
						isPureSingleTargetReset = true
						writtenTargetProp = targetProp
						if bit := expectConfigOverwriteBit(targetProp); bit != expectConfigNone &&
							analysis.overwritePreservesConfigBit(bit, rhs) {
							summary.restoredConfigMask |= bit
						}
						hasTargetAccessor := hasAnyAccessor && (hasDefineProp || analysis.objectLiteralHasAccessor(objExpr, targetProp, true, true, 0))
						if hasTargetAccessor {
							summary.escaped = true
						}
						if unconditional && stmt != nil && block != nil &&
							stmt.Kind == ast.KindExpressionStatement &&
							ast.SkipParentheses(stmt.AsExpressionStatement().Expression) == accessOuter.Parent &&
							!hasTargetAccessor {
							addReplacement(block, targetProp, objPropReplacementEvent{
								pos:        ref.Pos(),
								stmtPos:    stmt.Pos(),
								stmtEnd:    stmt.End(),
								node:       ref,
								rhs:        rhs,
								viaMethod:  true,
								methodName: propName,
							})
						}
					}
				}
				// Any non-pure method/tagged-template call — or a call to a method
				// property that was written/deleted elsewhere — can execute user code
				// with `this === sym` and escape `this` into an alias.
				if !isPureSingleTargetReset || hasDynamicWriteOrDelete || writtenOrDeletedProps[propName] {
					summary.escaped = true
					addPropMemberBits(propName, propKnown)
					if hasDynamicWriteOrDelete || !propKnown || writtenOrDeletedProps[propName] {
						summary.restoredConfigMask |= analysis.collectThisBitsFromSymbolPropWrites(sym, propName, propKnown, 0, nil)
					}
					if callOuter := outermostExpressionWrapper(accessOuter.Parent); callOuter != nil && callOuter.Parent != nil {
						switch callOuter.Parent.Kind {
						case ast.KindVariableDeclaration:
							vd := callOuter.Parent.AsVariableDeclaration()
							if vd != nil && vd.Initializer == callOuter && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
								aliasSym := callOuter.Parent.Symbol()
								if aliasSym == nil {
									aliasSym = resolveRstestRootSymbol(analysis.ctx, vd.Name())
								}
								summary.restoredConfigMask |= analysis.collectAliasRestoredConfigBits(aliasSym, aliasVisited)
							}
						case ast.KindBinaryExpression:
							bin := callOuter.Parent.AsBinaryExpression()
							if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == callOuter {
								if lhs := internalUtils.SkipAssertionsAndParens(bin.Left); lhs != nil && lhs.Kind == ast.KindIdentifier {
									summary.restoredConfigMask |= analysis.collectAliasRestoredConfigBits(resolveRstestRootSymbol(analysis.ctx, lhs), aliasVisited)
								}
							}
						}
					}
				}
				addMutation(ref, stmt, refFn, !isPureSingleTargetReset, writtenTargetProp)
				continue
			}
			if hasAnyGetter && analysis.objectLiteralHasAccessor(objExpr, propName, propKnown, false, 0) {
				summary.escaped = true
				addPropMemberBits(propName, propKnown)
				addMutation(ref, stmt, refFn, true, "")
			}
		case ast.KindVariableDeclaration:
			vd := outer.Parent.AsVariableDeclaration()
			if vd != nil && vd.Initializer == outer && vd.Name() != nil && vd.Name().Kind == ast.KindObjectBindingPattern {
				hasObjMutation := hasDynamicWriteOrDelete || len(writtenOrDeletedProps) > 0
				if hasDefineProp || (hasAnyGetter && analysis.destructuringPatternMayInvokeGetter(vd.Name(), objExpr, hasObjMutation)) {
					summary.escaped = true
					addPropMemberBits("", false)
					addMutation(ref, stmt, refFn, true, "")
				}
				continue
			}
			summary.escaped = true
			summary.hasAnyMutation = true
			if vd != nil && vd.Initializer == outer && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
				aliasSym := outer.Parent.Symbol()
				if aliasSym == nil {
					aliasSym = resolveRstestRootSymbol(analysis.ctx, vd.Name())
				}
				summary.restoredConfigMask |= analysis.collectAliasRestoredConfigBits(aliasSym, aliasVisited)
			}
		case ast.KindBinaryExpression:
			bin := outer.Parent.AsBinaryExpression()
			if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == outer {
				lhs := ast.SkipParentheses(bin.Left)
				if lhs != nil && lhs.Kind == ast.KindObjectLiteralExpression {
					hasObjMutation := hasDynamicWriteOrDelete || len(writtenOrDeletedProps) > 0
					if hasDefineProp || (hasAnyGetter && analysis.destructuringPatternMayInvokeGetter(lhs, objExpr, hasObjMutation)) {
						summary.escaped = true
						addPropMemberBits("", false)
						addMutation(ref, stmt, refFn, true, "")
					}
					continue
				}
				if lhs != nil && lhs.Kind == ast.KindIdentifier {
					summary.restoredConfigMask |= analysis.collectAliasRestoredConfigBits(resolveRstestRootSymbol(analysis.ctx, lhs), aliasVisited)
				}
			}
			summary.escaped = true
			summary.hasAnyMutation = true
		case ast.KindCallExpression:
			isPureDirect, isMemberOnSym, mName := analysis.isPureThisArgInvocation(outer.Parent, outer, sym)
			if isPureDirect || (isMemberOnSym && ((mName == "extend" && pureExtend) || (mName == "addEqualityTesters" && pureEquality))) {
				continue
			}
			if isMemberOnSym {
				addPropMemberBits(mName, true)
				if hasDynamicWriteOrDelete || writtenOrDeletedProps[mName] {
					summary.restoredConfigMask |= analysis.collectThisBitsFromSymbolPropWrites(sym, mName, true, 0, nil)
				}
			}
			for bit, clears := range analysis.mutatingBuiltinOverwriteChanges(outer.Parent, outer) {
				if !clears {
					summary.restoredConfigMask |= bit
				}
			}
			summary.escaped = true
			summary.hasAnyMutation = true
		default:
			summary.escaped = true
			summary.hasAnyMutation = true
		}
	}

	sortMutEvents := func(evs []objMutationEvent) {
		sort.Slice(evs, func(i, j int) bool {
			if evs[i].pos != evs[j].pos {
				return evs[i].pos < evs[j].pos
			}
			return evs[i].stmtEnd < evs[j].stmtEnd
		})
	}
	for _, evs := range summary.mutationsByFn {
		sortMutEvents(evs)
	}
	for _, evs := range summary.methodInvalidationsByFn {
		sortMutEvents(evs)
	}
	for _, byFn := range summary.propWriteEventsByProp {
		for _, evs := range byFn {
			sortMutEvents(evs)
		}
	}
	for _, reps := range summary.replacements {
		sort.Slice(reps, func(i, j int) bool {
			if reps[i].stmtEnd != reps[j].stmtEnd {
				return reps[i].stmtEnd < reps[j].stmtEnd
			}
			return reps[i].pos < reps[j].pos
		})
	}

	if analysis.objectLiteralMutationCycleHits == cycleBefore &&
		analysis.expectConfigCycleHits == expectCycleBefore &&
		analysis.expectConfigOverwriteCycleHits == overwriteCycleBefore {
		if analysis.objectLiteralMutations == nil {
			analysis.objectLiteralMutations = map[*ast.Symbol]*objectLiteralMutationSummary{}
		}
		analysis.objectLiteralMutations[sym] = summary
	}
	return summary
}

func (analysis *RstestCallAnalysis) hasInvalidatingEventForCall(
	sym *ast.Symbol,
	evsByFn map[*ast.Node][]objMutationEvent,
	refFn *ast.Node,
	callRef *ast.Node,
	callStmtEnd int,
) bool {
	for fn := range evsByFn {
		if fn == refFn {
			continue
		}
		if fn == nil || analysis.functionCanExecuteBefore(fn, callRef, sym, nil) {
			return true
		}
	}
	sameFnEvs := evsByFn[refFn]
	if len(sameFnEvs) == 0 {
		return false
	}
	// Reject any invalidation before `callStmtEnd` (including earlier statements
	// and same-statement computed-key mutations like `holder[(holder.reset = ..., 'reset')]()`),
	// skipping only the exact callRef position itself.
	for _, ev := range sameFnEvs {
		if ev.pos >= callStmtEnd {
			break
		}
		if ev.pos != callRef.Pos() {
			return true
		}
	}
	if last := sameFnEvs[len(sameFnEvs)-1]; last.pos >= callStmtEnd {
		idx := sort.Search(len(sameFnEvs), func(i int) bool {
			return sameFnEvs[i].pos >= callStmtEnd
		})
		if idx < len(sameFnEvs) && !directRangeCannotExecuteBefore(sameFnEvs[idx].pos, last.stmtEnd, callRef, sym) {
			return true
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) hasMethodInvalidatingMutationForCall(
	sym *ast.Symbol,
	summary *objectLiteralMutationSummary,
	callRef *ast.Node,
	methodName string,
) bool {
	if summary == nil || summary.escaped || callRef == nil {
		return true
	}
	callStmt, _ := enclosingStatementAndList(callRef)
	if callStmt == nil {
		return true
	}
	callStmtEnd := callStmt.End()
	refFn := ast.GetContainingFunction(callRef)
	if analysis.hasInvalidatingEventForCall(sym, summary.methodInvalidationsByFn, refFn, callRef, callStmtEnd) {
		return true
	}
	if byFn := summary.propWriteEventsByProp[methodName]; len(byFn) > 0 {
		if analysis.hasInvalidatingEventForCall(sym, byFn, refFn, callRef, callStmtEnd) {
			return true
		}
	}
	return false
}

// unconditionalPropertyReplacementBefore reports whether `sym` (a single-write
// `const` object literal `objExpr`) unconditionally replaced its own property
// `member` strictly before `useNode` in the same block — either directly
// (`holder.extend = rhs;`) or via an own method call (`holder.reset();`) —
// with no intervening, concurrent, or loop-re-executed mutation or escape.
func (analysis *RstestCallAnalysis) unconditionalPropertyReplacementBefore(
	sym *ast.Symbol,
	objExpr *ast.Node,
	member string,
	useNode *ast.Node,
) (*ast.Node, bool) {
	if analysis.ctx.Refs == nil || sym == nil || objExpr == nil || useNode == nil {
		return nil, false
	}
	summary := analysis.getObjectLiteralMutationSummary(sym, objExpr)
	if summary.escaped {
		return nil, false
	}
	if useNode.Kind == ast.KindIdentifier && analysis.symbolHasDefineProperty(sym, nil) {
		return nil, false
	}
	useStmt, useBlock := enclosingStatementAndList(useNode)
	if useStmt == nil || useBlock == nil {
		return nil, false
	}
	reps := summary.replacements[objReplacementKey{block: useBlock, member: member}]
	if len(reps) == 0 {
		return nil, false
	}
	useStmtPos := useStmt.Pos()
	useStmtEnd := useStmt.End()
	idx := sort.Search(len(reps), func(i int) bool {
		return reps[i].stmtEnd > useStmtPos
	})
	if idx == 0 {
		return nil, false
	}
	rep := reps[idx-1]

	useFn := ast.GetContainingFunction(useNode)
	for fn := range summary.mutationsByFn {
		if fn == useFn {
			continue
		}
		if fn == nil || analysis.functionCanExecuteBefore(fn, useNode, sym, nil) {
			return nil, false
		}
	}

	sameFnMuts := summary.mutationsByFn[useFn]
	if len(sameFnMuts) > 0 {
		// Reject any mutation strictly after `rep.stmtEnd` and before `useStmtEnd`
		// (including same-statement mutations such as `(holder.extend = ..., holder.extend())`,
		// computed-key mutations `holder[(holder.extend = ..., 'extend')]()`, and
		// sibling destructuring mutations), skipping only the exact self-call at `useNode`.
		afterIdx := sort.Search(len(sameFnMuts), func(i int) bool {
			return sameFnMuts[i].pos >= rep.stmtEnd
		})
		for i := afterIdx; i < len(sameFnMuts) && sameFnMuts[i].pos < useStmtEnd; i++ {
			if ast.IsAccessExpression(useNode) && sameFnMuts[i].pos == useNode.Pos() {
				continue
			}
			return nil, false
		}
		// Reject any mutation in a strictly later statement (`>= useStmtEnd`) that
		// can run before `useNode` on a subsequent loop iteration or re-invocation.
		if last := sameFnMuts[len(sameFnMuts)-1]; last.pos >= useStmtEnd {
			postIdx := sort.Search(len(sameFnMuts), func(i int) bool {
				return sameFnMuts[i].pos >= useStmtEnd
			})
			if postIdx < len(sameFnMuts) && !directRangeCannotExecuteBefore(sameFnMuts[postIdx].pos, last.stmtEnd, useNode, sym) {
				return nil, false
			}
		}
	}

	if rep.viaMethod && analysis.hasMethodInvalidatingMutationForCall(sym, summary, rep.node, rep.methodName) {
		return nil, false
	}
	return rep.rhs, true
}

func (analysis *RstestCallAnalysis) resolveObjectLiteralForBoundLookup(expr *ast.Node) (objExpr *ast.Node, mutableObj bool) {
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return nil, false
	}
	if expr.Kind == ast.KindObjectLiteralExpression {
		return expr, false
	}
	if expr.Kind == ast.KindIdentifier {
		sym := resolveRstestRootSymbol(analysis.ctx, expr)
		if sym == nil {
			return nil, false
		}
		summary := analysis.getExpectConfigSymbolSummary(sym)
		if len(summary.writes) == 1 && summary.writes[0].isConstDecl && summary.writes[0].rhs != nil {
			if initObj := internalUtils.SkipAssertionsAndParens(summary.writes[0].rhs); initObj != nil &&
				initObj.Kind == ast.KindObjectLiteralExpression {
				return initObj, analysis.symbolHasAnyPropertyMutation(sym, initObj)
			}
		}
	}
	return nil, false
}

func (analysis *RstestCallAnalysis) symbolHasAnyPropertyMutation(sym *ast.Symbol, objExpr *ast.Node) bool {
	summary := analysis.getObjectLiteralMutationSummary(sym, objExpr)
	return summary.escaped || summary.hasAnyMutation
}

func (analysis *RstestCallAnalysis) extractBoundIdentifierArgs(
	node *ast.Node,
	depth int,
	visitedWrites map[*expectConfigWrite]bool,
) boundCallResult {
	sym := resolveRstestRootSymbol(analysis.ctx, node)
	if sym == nil {
		return boundCallResult{}
	}
	if len(visitedWrites) == 0 {
		if cached, ok := analysis.boundCallConstCache[sym]; ok {
			return cached
		}
	}
	// Ensure `summary.evaluatedSources` and `configWritesByFn` are populated
	// so multi-write reaching queries use the O(log W) kill index.
	if analysis.expectConfigCallMember(node) == expectConfigNone {
		return boundCallResult{}
	}
	summary := analysis.getExpectConfigSymbolSummary(sym)
	if len(summary.writes) == 0 {
		return boundCallResult{}
	}
	readStmt, readBlock := enclosingStatementAndList(node)
	readFn := ast.GetContainingFunction(node)
	lastKillPos := lastUnconditionalKillBefore(summary.blockKills[readBlock], readStmt)

	postAllWrites := len(visitedWrites) == 0 &&
		summary.evaluatedSources &&
		summary.singleEnclosingFnOK &&
		summary.singleEnclosingFn == readFn &&
		node.Pos() > summary.maxWriteEnd
	postKey := boundCallPostKey{sym: sym, block: readBlock, lastKillPos: lastKillPos}
	if postAllWrites {
		if cached, ok := analysis.boundCallPostAllWritesCache[postKey]; ok {
			return cached
		}
	}

	cycleBefore := analysis.boundCycleHits + analysis.expectConfigCycleHits + analysis.expectConfigOverwriteCycleHits + analysis.objectLiteralMutationCycleHits
	var res boundCallResult
	evalWrite := func(w *expectConfigWrite) {
		if visitedWrites[w] {
			analysis.boundCycleHits++
			return
		}
		if w.boundCached {
			res.merge(w.boundRes)
			return
		}
		if visitedWrites == nil {
			visitedWrites = map[*expectConfigWrite]bool{}
		}
		if len(visitedWrites) >= 256 {
			res.merge(boundCallResult{hasOpaque: true, alwaysBound: true, hasCallable: true})
			return
		}
		visitedWrites[w] = true
		wCycleBefore := analysis.boundCycleHits + analysis.expectConfigCycleHits + analysis.expectConfigOverwriteCycleHits + analysis.objectLiteralMutationCycleHits
		wRes := analysis.evalWriteBoundArgs(w, 0, visitedWrites)
		delete(visitedWrites, w)
		if analysis.boundCycleHits+analysis.expectConfigCycleHits+analysis.expectConfigOverwriteCycleHits+analysis.objectLiteralMutationCycleHits == wCycleBefore {
			w.boundCached = true
			w.boundRes = wRes
		}
		res.merge(wRes)
	}
	saturated := func() bool {
		return res.hasCallable && res.hasOpaque && !res.alwaysBound
	}

	if summary.evaluatedSources {
		if sameFnIndices := summary.configWritesByFn[readFn]; len(sameFnIndices) > 0 {
			start := 0
			if lastKillPos >= 0 {
				start = sort.Search(len(sameFnIndices), func(i int) bool {
					return summary.writes[sameFnIndices[i]].exprEnd > lastKillPos
				})
			}
			for _, idx := range sameFnIndices[start:] {
				w := &summary.writes[idx]
				if w.evaluatedMember&expectConfigCallableMask == 0 {
					continue
				}
				if !analysis.writeCanReachRead(sym, w, node, readStmt, readBlock, readFn, summary) {
					continue
				}
				evalWrite(w)
				if saturated() {
					break
				}
			}
		}
		if !saturated() {
			for fn, indices := range summary.configWritesByFn {
				if fn == readFn {
					continue
				}
				for _, idx := range indices {
					w := &summary.writes[idx]
					if w.evaluatedMember&expectConfigCallableMask == 0 {
						continue
					}
					if !analysis.writeCanReachRead(sym, w, node, readStmt, readBlock, readFn, summary) {
						continue
					}
					evalWrite(w)
					if saturated() {
						break
					}
				}
				if saturated() {
					break
				}
			}
		}
	} else {
		for i := range summary.writes {
			w := &summary.writes[i]
			wKind := analysis.evalExpectConfigWriteStatic(w)
			if wKind&expectConfigCallableMask == 0 {
				continue
			}
			if !analysis.writeCanReachRead(sym, w, node, readStmt, readBlock, readFn, summary) {
				continue
			}
			evalWrite(w)
			if saturated() {
				break
			}
		}
	}

	if len(visitedWrites) == 0 &&
		analysis.boundCycleHits+analysis.expectConfigCycleHits+analysis.expectConfigOverwriteCycleHits+analysis.objectLiteralMutationCycleHits == cycleBefore {
		if len(summary.writes) == 1 && summary.writes[0].isConstDecl {
			if analysis.boundCallConstCache == nil {
				analysis.boundCallConstCache = map[*ast.Symbol]boundCallResult{}
			}
			analysis.boundCallConstCache[sym] = res
		} else if postAllWrites {
			if analysis.boundCallPostAllWritesCache == nil {
				analysis.boundCallPostAllWritesCache = map[boundCallPostKey]boundCallResult{}
			}
			analysis.boundCallPostAllWritesCache[postKey] = res
		}
	}
	return res
}

func (analysis *RstestCallAnalysis) evalWriteBoundArgs(
	w *expectConfigWrite,
	depth int,
	visitedWrites map[*expectConfigWrite]bool,
) boundCallResult {
	exprs, hasUnbound := analysis.expectConfigWriteBoundSources(w)
	var res boundCallResult
	if hasUnbound {
		res.merge(boundCallResult{hasCallable: true, alwaysBound: false})
	}
	for _, expr := range exprs {
		if analysis.expectConfigCallMember(expr) == expectConfigNone {
			continue
		}
		sub := analysis.extractBoundCallFirstArgs(expr, depth, visitedWrites)
		if !sub.hasCallable {
			sub = boundCallResult{hasCallable: true, alwaysBound: false}
		}
		res.merge(sub)
	}
	return res
}

func (analysis *RstestCallAnalysis) combineBoundCallBranches(
	left, right *ast.Node,
	depth int,
	visitedWrites map[*expectConfigWrite]bool,
) boundCallResult {
	var res boundCallResult
	if analysis.expectConfigCallMember(left) != expectConfigNone {
		sub := analysis.extractBoundCallFirstArgs(left, depth, visitedWrites)
		if !sub.hasCallable {
			sub = boundCallResult{hasCallable: true, alwaysBound: false}
		}
		res.merge(sub)
	}
	if analysis.expectConfigCallMember(right) != expectConfigNone {
		sub := analysis.extractBoundCallFirstArgs(right, depth, visitedWrites)
		if !sub.hasCallable {
			sub = boundCallResult{hasCallable: true, alwaysBound: false}
		}
		res.merge(sub)
	}
	return res
}

func (analysis *RstestCallAnalysis) expectConfigWriteBoundSources(w *expectConfigWrite) (exprs []*ast.Node, hasUnbound bool) {
	if w.initialKind&expectConfigCallableMask != 0 {
		hasUnbound = true
	}
	if w.rhs != nil {
		exprs = append(exprs, w.rhs)
	}
	if w.bindingElem != nil {
		bExprs, bUnbound := analysis.destructuredBoundSources(w.bindingElem)
		exprs = append(exprs, bExprs...)
		if bUnbound {
			hasUnbound = true
		}
	}
	if w.assignDestruct != nil {
		aExprs, aUnbound := analysis.assignmentDestructBoundSources(w.assignDestruct, w.assignDefault)
		exprs = append(exprs, aExprs...)
		if aUnbound {
			hasUnbound = true
		}
	}
	return exprs, hasUnbound
}

func (analysis *RstestCallAnalysis) destructuredBoundSources(declaration *ast.Node) (exprs []*ast.Node, hasUnbound bool) {
	binding := declaration.AsBindingElement()
	if binding == nil || binding.Name() == nil {
		return nil, false
	}
	pattern := declaration.Parent
	if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		if binding.Initializer != nil {
			exprs = append(exprs, binding.Initializer)
		}
		return exprs, false
	}
	if binding.DotDotDotToken != nil {
		if binding.Initializer != nil {
			exprs = append(exprs, binding.Initializer)
		}
		return exprs, false
	}
	member, known, staticKey := analysis.evalBindingElementPropertyName(binding)
	defaultSkipped := false
	exactFromLiteral := false
	sources, definiteSrc, sourcesMayBeMutable, sourcesUnresolvedGetter := analysis.bindingPatternSourceExprs(pattern)
	if sourcesMayBeMutable || sourcesUnresolvedGetter {
		hasUnbound = true
	}
	if known {
		if definiteSrc != nil {
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(definiteSrc); objExpr != nil && !mutableObj {
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(pattern, declaration, objExpr, false)
				defLookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter, 0)
				defaultSkipped = defLookup.definitelyDefined
			}
		}
		allExact := len(sources) > 0 && !sourcesMayBeMutable && !sourcesUnresolvedGetter
		for _, srcExpr := range sources {
			objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(srcExpr)
			if objExpr == nil {
				allExact = false
				continue
			}
			priorGetter := analysis.precedingDestructuringMayInvokeGetter(pattern, declaration, objExpr, mutableObj)
			if mutableObj && !sourcesMayBeMutable && !priorGetter && len(sources) == 1 {
				if unwrapped := internalUtils.SkipAssertionsAndParens(srcExpr); unwrapped != nil && unwrapped.Kind == ast.KindIdentifier {
					if sym := resolveRstestRootSymbol(analysis.ctx, unwrapped); sym != nil {
						if replacement, replaced := analysis.unconditionalPropertyReplacementBefore(sym, objExpr, member, unwrapped); replaced {
							exprs = append(exprs, replacement)
							defaultSkipped = analysis.isDefinitelyDefinedExpr(replacement, 0, nil)
							continue
						}
					}
				}
			}
			lookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter || mutableObj || sourcesMayBeMutable, 0)
			exprs = append(exprs, lookup.possibleExprs...)
			useAllGetters := sourcesMayBeMutable || (mutableObj && analysis.srcExprObjEscaped(srcExpr, objExpr))
			if priorGetter || useAllGetters || mutableObj {
				gExprs, gContainers, gUnresolved := analysis.collectPrecedingGetterPropWriteExprs(pattern, declaration, objExpr, member, useAllGetters, mutableObj)
				exprs = append(exprs, gExprs...)
				if len(gContainers) > 0 || gUnresolved {
					hasUnbound = true
					allExact = false
				}
			}
			if mutableObj || lookup.mayHaveUnknownSource {
				hasUnbound = true
				allExact = false
			} else if !lookup.foundExplicit {
				allExact = false
			}
		}
		exactFromLiteral = allExact && !hasUnbound
	} else if !staticKey {
		for _, srcExpr := range sources {
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(srcExpr); objExpr != nil {
				exprs = append(exprs, analysis.allObjectLiteralPropertyExprs(objExpr)...)
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(pattern, declaration, objExpr, mutableObj)
				useAllGetters := sourcesMayBeMutable || (mutableObj && analysis.srcExprObjEscaped(srcExpr, objExpr))
				if priorGetter || useAllGetters || mutableObj {
					gExprs, _, _ := analysis.collectPrecedingGetterPropWriteExprs(pattern, declaration, objExpr, "", useAllGetters, mutableObj)
					exprs = append(exprs, gExprs...)
				}
			}
		}
		hasUnbound = true
	}
	if binding.Initializer != nil && !defaultSkipped {
		exprs = append(exprs, binding.Initializer)
	}
	if !exactFromLiteral {
		if container := analysis.expectConfigBindingPatternKind(pattern); container != expectConfigNone {
			if classifyDestructuredExpectConfigProperty(container, member, known, staticKey)&expectConfigCallableMask != 0 {
				hasUnbound = true
			}
		}
	}
	return exprs, hasUnbound
}

func (analysis *RstestCallAnalysis) assignmentDestructBoundSources(prop, defaultExpr *ast.Node) (exprs []*ast.Node, hasUnbound bool) {
	if prop == nil || prop.Parent == nil || prop.Parent.Kind != ast.KindObjectLiteralExpression || prop.Kind == ast.KindSpreadAssignment {
		return nil, false
	}
	propName := prop.Name()
	if propName == nil {
		return nil, false
	}
	obj := prop.Parent
	member, known, staticKey := analysis.evalPropertyNameKey(propName)
	defaultSkipped := false
	exactFromLiteral := false
	sources, definiteSrc, sourcesMayBeMutable, sourcesUnresolvedGetter := analysis.assignmentObjectSourceExprs(obj)
	if sourcesMayBeMutable || sourcesUnresolvedGetter {
		hasUnbound = true
	}
	if known {
		if definiteSrc != nil {
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(definiteSrc); objExpr != nil && !mutableObj {
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(obj, prop, objExpr, false)
				defLookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter, 0)
				defaultSkipped = defLookup.definitelyDefined
			}
		}
		allExact := len(sources) > 0 && !sourcesMayBeMutable && !sourcesUnresolvedGetter
		for _, srcExpr := range sources {
			objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(srcExpr)
			if objExpr == nil {
				allExact = false
				continue
			}
			priorGetter := analysis.precedingDestructuringMayInvokeGetter(obj, prop, objExpr, mutableObj)
			if mutableObj && !sourcesMayBeMutable && !priorGetter && len(sources) == 1 {
				if unwrapped := internalUtils.SkipAssertionsAndParens(srcExpr); unwrapped != nil && unwrapped.Kind == ast.KindIdentifier {
					if sym := resolveRstestRootSymbol(analysis.ctx, unwrapped); sym != nil {
						if replacement, replaced := analysis.unconditionalPropertyReplacementBefore(sym, objExpr, member, unwrapped); replaced {
							exprs = append(exprs, replacement)
							defaultSkipped = analysis.isDefinitelyDefinedExpr(replacement, 0, nil)
							continue
						}
					}
				}
			}
			lookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter || mutableObj || sourcesMayBeMutable, 0)
			exprs = append(exprs, lookup.possibleExprs...)
			useAllGetters := sourcesMayBeMutable || (mutableObj && analysis.srcExprObjEscaped(srcExpr, objExpr))
			if priorGetter || useAllGetters || mutableObj {
				gExprs, gContainers, gUnresolved := analysis.collectPrecedingGetterPropWriteExprs(obj, prop, objExpr, member, useAllGetters, mutableObj)
				exprs = append(exprs, gExprs...)
				if len(gContainers) > 0 || gUnresolved {
					hasUnbound = true
					allExact = false
				}
			}
			if mutableObj || lookup.mayHaveUnknownSource {
				hasUnbound = true
				allExact = false
			} else if !lookup.foundExplicit {
				allExact = false
			}
		}
		exactFromLiteral = allExact && !hasUnbound
	} else if !staticKey {
		for _, srcExpr := range sources {
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(srcExpr); objExpr != nil {
				exprs = append(exprs, analysis.allObjectLiteralPropertyExprs(objExpr)...)
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(obj, prop, objExpr, mutableObj)
				useAllGetters := sourcesMayBeMutable || (mutableObj && analysis.srcExprObjEscaped(srcExpr, objExpr))
				if priorGetter || useAllGetters || mutableObj {
					gExprs, _, _ := analysis.collectPrecedingGetterPropWriteExprs(obj, prop, objExpr, "", useAllGetters, mutableObj)
					exprs = append(exprs, gExprs...)
				}
			}
		}
		hasUnbound = true
	}
	if defaultExpr != nil && !defaultSkipped {
		exprs = append(exprs, defaultExpr)
	}
	if !exactFromLiteral {
		if container := analysis.expectConfigAssignmentObjectKind(obj); container != expectConfigNone {
			if classifyDestructuredExpectConfigProperty(container, member, known, staticKey)&expectConfigCallableMask != 0 {
				hasUnbound = true
			}
		}
	}
	return exprs, hasUnbound
}

func (analysis *RstestCallAnalysis) expectConfigCallTarget(call *ast.Node) (kind expectConfigKind, outArgs []*ast.Node) {
	if call == nil || call.Kind != ast.KindCallExpression {
		return expectConfigNone, nil
	}
	args := call.Arguments()
	callee := internalUtils.SkipAssertionsAndParens(call.Expression())
	if callee == nil {
		return expectConfigNone, nil
	}
	// Handle `Reflect.apply(expect.extend, expect, [matchers])`.
	if analysis.isGlobalBuiltinMemberCall(callee, "Reflect", "apply") {
		if len(args) > 0 && args[0] != nil && args[0].Kind != ast.KindSpreadElement {
			target := analysis.expectConfigCallMember(args[0])
			if target != expectConfigNone {
				// `Reflect.apply` throws a TypeError before calling `target`
				// unless a third `argumentsList` argument is provided.
				if !hasSpreadBeforeIndex(args, 2) && len(args) < 3 {
					return expectConfigNone, nil
				}
				bound := analysis.extractBoundCallFirstArgs(args[0], 0, nil)
				outArgs = append(outArgs, bound.toArgs()...)
				if !bound.alwaysBound {
					if hasSpreadBeforeIndex(args, 2) {
						outArgs = append(outArgs, nil)
					} else if argNode, ok := firstArgFromApplyArray(args[2]); ok {
						outArgs = append(outArgs, argNode)
					}
				}
				return target, outArgs
			}
		}
	}
	// Handle `expect.extend.call(expect, matchers)` and `expect.extend.apply(expect, [matchers])`.
	if ast.IsAccessExpression(callee) {
		if method, ok := analysis.evalExpectConfigAccessName(callee); ok && (method == "call" || method == "apply") {
			target := analysis.expectConfigCallMember(callee.Expression())
			if target != expectConfigNone {
				bound := analysis.extractBoundCallFirstArgs(callee.Expression(), 0, nil)
				outArgs = append(outArgs, bound.toArgs()...)
				if !bound.alwaysBound {
					if hasSpreadBeforeIndex(args, 1) {
						outArgs = append(outArgs, nil)
					} else if len(args) >= 2 {
						if method == "call" {
							outArgs = append(outArgs, internalUtils.SkipAssertionsAndParens(args[1]))
						} else if argNode, ok := firstArgFromApplyArray(args[1]); ok {
							outArgs = append(outArgs, argNode)
						}
					}
				}
				return target, outArgs
			}
		}
	}
	kind = analysis.expectConfigCallMember(callee)
	if kind == expectConfigNone {
		return expectConfigNone, nil
	}
	bound := analysis.extractBoundCallFirstArgs(callee, 0, nil)
	outArgs = append(outArgs, bound.toArgs()...)
	if !bound.alwaysBound && len(args) > 0 {
		if args[0] != nil && args[0].Kind == ast.KindSpreadElement {
			outArgs = append(outArgs, nil)
		} else {
			outArgs = append(outArgs, internalUtils.SkipAssertionsAndParens(args[0]))
		}
	}
	return kind, outArgs
}

func (analysis *RstestCallAnalysis) recordExtendedMatchers(matchers *ast.Node) {
	if matchers == nil || matchers.Kind != ast.KindObjectLiteralExpression ||
		matchers.AsObjectLiteralExpression().Properties == nil {
		analysis.allExpectMatchersOverridden = true
		return
	}
	if analysis.recordedMatcherLiterals[matchers] {
		return
	}
	if analysis.recordedMatcherLiterals == nil {
		analysis.recordedMatcherLiterals = map[*ast.Node]bool{}
	}
	analysis.recordedMatcherLiterals[matchers] = true
	for _, property := range matchers.AsObjectLiteralExpression().Properties.Nodes {
		if property == nil {
			continue
		}
		name := property.Name()
		if name == nil {
			analysis.allExpectMatchersOverridden = true
			continue
		}
		matcher, ok := internalUtils.GetStaticPropertyName(name)
		if !ok {
			matcher, ok = analysis.staticStringEvaluator().EvalPropertyName(name)
		}
		if !ok {
			analysis.allExpectMatchersOverridden = true
			continue
		}
		if analysis.overriddenExpectMatchers == nil {
			analysis.overriddenExpectMatchers = map[string]bool{}
		}
		analysis.overriddenExpectMatchers[matcher] = true
	}
}

const expectConfigCallableMask = expectConfigExtend | expectConfigAddEqualityTesters | expectConfigDynamicMember

func (analysis *RstestCallAnalysis) expectConfigCallMember(node *ast.Node) expectConfigKind {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return expectConfigNone
	}
	// `(0, expect.extend)(...)`
	if internalUtils.IsCommaOperator(node) {
		return analysis.expectConfigCallMember(node.AsBinaryExpression().Right)
	}
	// `expect.extend.bind(expect)(...)`
	if node.Kind == ast.KindCallExpression {
		callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if method, ok := analysis.evalExpectConfigAccessName(callee); ok && method == "bind" {
				return analysis.expectConfigCallMember(callee.Expression())
			}
		}
		return expectConfigNone
	}
	return analysis.expectConfigMember(node) & expectConfigCallableMask
}

func (analysis *RstestCallAnalysis) evalExpectConfigAccessName(node *ast.Node) (string, bool) {
	if name, ok := internalUtils.AccessExpressionStaticName(node); ok {
		return name, true
	}
	if node.Kind == ast.KindElementAccessExpression {
		return analysis.staticStringEvaluator().EvalAccessExpressionName(node)
	}
	return "", false
}

func (analysis *RstestCallAnalysis) expectConfigMember(node *ast.Node) expectConfigKind {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return expectConfigNone
	}
	if isImportMetaRstest(node) || isRstestRequireCall(node) {
		return expectConfigReceiver
	}
	// Handle wrapper expressions that forward an expect/receiver value:
	// comma `(0, expect)`, logical `&&` (RHS only), `??`/`||` (union),
	// conditional `cond ? a : b` (union), `.bind(...)`, `Object.assign({}, ...)`,
	// and object spread `{ ...expect }`.
	switch node.Kind {
	case ast.KindBinaryExpression:
		bin := node.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return expectConfigNone
		}
		switch bin.OperatorToken.Kind {
		case ast.KindCommaToken, ast.KindAmpersandAmpersandToken:
			return analysis.expectConfigMember(bin.Right)
		case ast.KindQuestionQuestionToken, ast.KindBarBarToken:
			return analysis.expectConfigMember(bin.Left) | analysis.expectConfigMember(bin.Right)
		}
		return expectConfigNone
	case ast.KindConditionalExpression:
		cond := node.AsConditionalExpression()
		if cond == nil {
			return expectConfigNone
		}
		return analysis.expectConfigMember(cond.WhenTrue) | analysis.expectConfigMember(cond.WhenFalse)
	case ast.KindCallExpression:
		callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if method, ok := analysis.evalExpectConfigAccessName(callee); ok && method == "bind" {
				return analysis.expectConfigMember(callee.Expression())
			}
			if analysis.isGlobalBuiltinMemberCall(callee, "Object", "assign") {
				return analysis.expectConfigObjectAssignMember(node.Arguments())
			}
		}
		return expectConfigNone
	case ast.KindObjectLiteralExpression:
		return analysis.expectConfigObjectLiteralSpreadMember(node)
	}
	if ast.IsAccessExpression(node) {
		member, ok := analysis.evalExpectConfigAccessName(node)
		parentKind := analysis.expectConfigMember(node.Expression())
		if parentKind == expectConfigNone {
			return expectConfigNone
		}
		var out expectConfigKind
		if parentKind&expectConfigReceiver != 0 {
			if ok {
				if member == "expect" {
					out |= expectConfigExpect
				}
			} else if !analysis.staticStringEvaluator().HasStaticAccessExpressionKey(node) {
				// A genuinely dynamic key on an Rstest receiver (e.g.
				// `r[key].extend(...)`) may select `expect` at runtime.
				out |= expectConfigExpect
			}
		}
		if parentKind&expectConfigExpect != 0 {
			if ok {
				if member == "extend" && parentKind&expectConfigExpectExtend != 0 {
					out |= expectConfigExtend
				}
				if member == "addEqualityTesters" && parentKind&expectConfigExpectEquality != 0 {
					out |= expectConfigAddEqualityTesters
				}
			} else if !analysis.staticStringEvaluator().HasStaticAccessExpressionKey(node) {
				out |= expectConfigDynamicMember
			}
		}
		return out
	}
	if ast.IsIdentifier(node) {
		return analysis.expectConfigIdentifierMember(node)
	}
	return expectConfigNone
}

func (analysis *RstestCallAnalysis) applyObjectLiteralPropertiesToKind(kind expectConfigKind, objNode *ast.Node) expectConfigKind {
	return analysis.applyObjectLiteralPropertiesWithTargetAccessors(kind, objNode, expectConfigNone)
}

func (analysis *RstestCallAnalysis) applyObjectLiteralPropertiesWithTargetAccessors(
	kind expectConfigKind,
	objNode *ast.Node,
	targetAccessorMask expectConfigKind,
) expectConfigKind {
	obj := objNode.AsObjectLiteralExpression()
	if obj == nil || obj.Properties == nil {
		return kind
	}
	var protoExpr *ast.Node
	var ownExplicitMask expectConfigKind
	var hasGetterMask expectConfigKind
	for _, prop := range obj.Properties.Nodes {
		if prop == nil {
			continue
		}
		if prop.Kind == ast.KindSpreadAssignment {
			spreadKind := analysis.expectConfigMember(prop.AsSpreadAssignment().Expression)
			kind |= spreadKind & (expectConfigReceiver | expectConfigExpect)
			continue
		}
		propName := prop.Name()
		if propName == nil {
			continue
		}
		name, ok, staticKey := analysis.evalPropertyNameKey(propName)
		if !ok {
			// A dynamically computed key (`{ [key]: r.expect.extend }`) may
			// define `"extend"`, `"addEqualityTesters"`, or `"expect"` at runtime.
			if !staticKey {
				var dynKind expectConfigKind
				switch prop.Kind {
				case ast.KindPropertyAssignment:
					dynKind = analysis.expectConfigMember(prop.AsPropertyAssignment().Initializer)
				case ast.KindShorthandPropertyAssignment:
					dynKind = analysis.expectConfigIdentifierMember(prop.AsShorthandPropertyAssignment().Name())
				case ast.KindGetAccessor:
					dynKind = analysis.getterExpectConfigKind(prop)
				}
				if dynKind&expectConfigExtend != 0 {
					kind |= expectConfigExpectExtend
				}
				if dynKind&expectConfigAddEqualityTesters != 0 {
					kind |= expectConfigExpectEquality
				}
				if dynKind&expectConfigDynamicMember != 0 {
					kind |= expectConfigExpect
				}
				if dynKind&expectConfigExpect != 0 {
					kind |= expectConfigReceiver
				}
			}
			continue
		}
		if name == "__proto__" && prop.Kind == ast.KindPropertyAssignment && propName.Kind != ast.KindComputedPropertyName {
			protoExpr = prop.AsPropertyAssignment().Initializer
			continue
		}
		var valKind expectConfigKind
		switch prop.Kind {
		case ast.KindPropertyAssignment:
			valKind = analysis.expectConfigMember(prop.AsPropertyAssignment().Initializer)
		case ast.KindShorthandPropertyAssignment:
			valKind = analysis.expectConfigIdentifierMember(prop.AsShorthandPropertyAssignment().Name())
		case ast.KindGetAccessor:
			valKind = analysis.getterExpectConfigKind(prop)
		}
		applySlot := func(slotBit, matchBit expectConfigKind) {
			ownExplicitMask |= slotBit
			// A setter (`set extend(v) {}`) paired with an earlier getter keeps
			// the getter readable in JS rather than wiping it out.
			if prop.Kind == ast.KindSetAccessor {
				if hasGetterMask&slotBit == 0 && targetAccessorMask&slotBit == 0 {
					kind &^= slotBit
				}
				return
			}
			if prop.Kind == ast.KindGetAccessor {
				hasGetterMask |= slotBit
			} else {
				hasGetterMask &^= slotBit
			}
			if targetAccessorMask&slotBit == 0 {
				kind &^= slotBit
			}
			if valKind&matchBit != 0 {
				kind |= slotBit
			}
		}
		switch name {
		case "extend":
			applySlot(expectConfigExpectExtend, expectConfigExtend|expectConfigDynamicMember)
		case "addEqualityTesters":
			applySlot(expectConfigExpectEquality, expectConfigAddEqualityTesters|expectConfigDynamicMember)
		case "expect":
			applySlot(expectConfigReceiver, expectConfigExpect)
		}
	}
	// In JS, own properties (even set to `undefined`) always shadow `__proto__`
	// regardless of whether `__proto__` appears before or after them in the literal.
	if protoExpr != nil {
		protoKind := analysis.expectConfigMember(protoExpr) & (expectConfigReceiver | expectConfigExpect)
		kind |= protoKind &^ ownExplicitMask
	}
	return kind
}

// expectConfigObjectLiteralSpreadMember resolves shallow-copy literals such as
// `{ ...expect }` or `{ ...r }`, clearing only the specific overwritten member
// bits unless the property assigns back a genuine expect/extend/tester value.
func (analysis *RstestCallAnalysis) expectConfigObjectLiteralSpreadMember(node *ast.Node) expectConfigKind {
	return analysis.applyObjectLiteralPropertiesToKind(expectConfigNone, node)
}

func (analysis *RstestCallAnalysis) objectAccessorConfigMask(expr *ast.Node) expectConfigKind {
	var mask expectConfigKind
	if analysis.exprMayHaveAccessorOnProp(expr, "extend", true, 0) {
		mask |= expectConfigExpectExtend
	}
	if analysis.exprMayHaveAccessorOnProp(expr, "addEqualityTesters", true, 0) {
		mask |= expectConfigExpectEquality
	}
	if analysis.exprMayHaveAccessorOnProp(expr, "expect", true, 0) {
		mask |= expectConfigReceiver
	}
	return mask
}

func (analysis *RstestCallAnalysis) expectConfigObjectAssignMember(args []*ast.Node) expectConfigKind {
	if len(args) == 0 {
		return expectConfigNone
	}
	targetAccessorMask := analysis.objectAccessorConfigMask(args[0])
	var kind expectConfigKind
	for i, arg := range args {
		a := internalUtils.SkipAssertionsAndParens(arg)
		if a == nil || a.Kind == ast.KindSpreadElement {
			continue
		}
		if a.Kind == ast.KindObjectLiteralExpression {
			accMask := expectConfigNone
			if i > 0 {
				accMask = targetAccessorMask
			}
			kind = analysis.applyObjectLiteralPropertiesWithTargetAccessors(kind, a, accMask)
			continue
		}
		kind |= analysis.expectConfigMember(a) & (expectConfigReceiver | expectConfigExpect)
	}
	return kind
}

type expectConfigWrite struct {
	pos             int
	exprEnd         int
	stmt            *ast.Node
	block           *ast.Node
	enclosingFn     *ast.Node
	unconditional   bool
	isConstDecl     bool
	isShallowCopy   bool
	bindingElem     *ast.Node
	assignDestruct  *ast.Node
	assignDefault   *ast.Node
	rhs             *ast.Node
	initialKind     expectConfigKind
	evaluatedMember expectConfigKind
	boundCached     bool
	boundRes        boundCallResult
}

type expectConfigBlockKill struct {
	pos     int
	stmtEnd int
}

type expectConfigSymbolSummary struct {
	writes                   []expectConfigWrite
	blockKills               map[*ast.Node][]expectConfigBlockKill
	configWritesByFn         map[*ast.Node][]int // indices into writes sorted by exprEnd
	allConfigMask            expectConfigKind
	initialObjectLiteralKind expectConfigKind
	hasAnyConfigSource       bool
	evaluatedSources         bool
	shallowAliasTarget       *ast.Symbol
	isDirectShallowCopy      bool
	maxWriteEnd              int
	singleEnclosingFnOK      bool
	singleEnclosingFn        *ast.Node
}

type expectConfigOverwriteKey struct {
	block *ast.Node
	bit   expectConfigKind
}

type expectConfigOverwriteEvent struct {
	pos          int
	stmtEnd      int
	clears       bool
	postCallNode *ast.Node // non-nil when this restore runs only after evaluating/calling `postCallNode`
}

// expectConfigOverwriteSummary stores, for each (block, config bit), the
// sorted sequence of property overwrite/delete (`clears: true`) and restore
// (`clears: false`) events in that block, plus any functions that may restore
// a bit. Queries binary-search the latest event before `useStmt` (and after
// the reaching write), running in O(log E) with sound MUST-overwrite semantics.
type expectConfigOverwriteSummary struct {
	events        map[expectConfigOverwriteKey][]expectConfigOverwriteEvent
	maxRestoreEnd map[expectConfigOverwriteKey]int
	fnRestores    map[expectConfigKind][]*ast.Node
	escaped       bool
}

type expectConfigPosRange struct {
	minPos int
	maxEnd int
}

type expectConfigFnSummary struct {
	alwaysExecutable bool
	minDirectPos     int
	maxDirectEnd     int
	hasDirect        bool
	callerFns        []*ast.Node
	callerBounds     map[*ast.Node]expectConfigPosRange
}

func (analysis *RstestCallAnalysis) expectConfigIdentifierMember(node *ast.Node) expectConfigKind {
	if node == nil || node.Kind != ast.KindIdentifier {
		return expectConfigNone
	}
	symbol := resolveRstestRootSymbol(analysis.ctx, node)
	if symbol == nil {
		if node.AsIdentifier().Text == "expect" && !analysis.globalExpectWritten {
			return expectConfigExpect
		}
		return expectConfigNone
	}
	rootInfo, ok := analysis.expectRoots[symbol]
	if !ok {
		rootInfo = classifyRstestExpectRoot(node.AsIdentifier().Text, node, symbol, analysis)
		analysis.expectRoots[symbol] = rootInfo
	}
	switch rootInfo.Kind {
	case rstestExpectRootDirect:
		return expectConfigExpect
	case rstestExpectRootReceiver:
		return expectConfigReceiver
	}
	// Consult the permanent per-symbol alias cache BEFORE inspecting writes so
	// repeated calls to ordinary helper functions and alias DAGs stay O(1).
	if member, ok := analysis.expectConfigAliases[symbol]; ok {
		return member
	}
	if analysis.expectConfigVisiting[symbol] {
		analysis.expectConfigCycleHits++
		if analysis.expectConfigSelfHit == nil {
			analysis.expectConfigSelfHit = map[*ast.Symbol]bool{}
		}
		analysis.expectConfigSelfHit[symbol] = true
		for active := range analysis.expectConfigVisiting {
			if active != symbol {
				if analysis.expectConfigOtherCycleHits == nil {
					analysis.expectConfigOtherCycleHits = map[*ast.Symbol]int{}
				}
				analysis.expectConfigOtherCycleHits[active]++
			}
		}
		return analysis.expectConfigPartial[symbol]
	}
	if analysis.expectConfigVisiting == nil {
		analysis.expectConfigVisiting = map[*ast.Symbol]bool{}
	}
	if analysis.expectConfigPartial == nil {
		analysis.expectConfigPartial = map[*ast.Symbol]expectConfigKind{}
	}
	analysis.expectConfigVisiting[symbol] = true
	analysis.expectConfigPartial[symbol] = expectConfigNone
	delete(analysis.expectConfigSelfHit, symbol)
	defer func() {
		delete(analysis.expectConfigVisiting, symbol)
		delete(analysis.expectConfigPartial, symbol)
		delete(analysis.expectConfigSelfHit, symbol)
	}()

	summary := analysis.getExpectConfigSymbolSummary(symbol)
	if len(summary.writes) == 0 {
		analysis.cacheExpectConfigAlias(symbol, expectConfigNone)
		return expectConfigNone
	}

	// Evaluate every write's underlying config kind ONCE per symbol — running
	// a small monotone fixed-point when `symbol` references itself (e.g.
	// `fn = fn.bind(...)` or `ex = { ...ex }`), and refusing to poison the
	// permanent cache only when a strictly outer ancestor cycle was hit.
	var localEvaluated []expectConfigKind
	var allMask expectConfigKind
	if !summary.evaluatedSources {
		localEvaluated = make([]expectConfigKind, len(summary.writes))
		var initObj *ast.Node
		if len(summary.writes) == 1 && summary.writes[0].isConstDecl && summary.writes[0].rhs != nil {
			if obj := internalUtils.SkipAssertionsAndParens(summary.writes[0].rhs); obj != nil &&
				obj.Kind == ast.KindObjectLiteralExpression {
				initObj = obj
			}
		}
		var origObjKind expectConfigKind
		otherCycleHit := false
		for range 6 {
			otherBefore := analysis.expectConfigOtherCycleHits[symbol]
			mutCycleBefore := analysis.objectLiteralMutationCycleHits
			overwriteCycleBefore := analysis.expectConfigOverwriteCycleHits
			delete(analysis.expectConfigSelfHit, symbol)
			allMask = expectConfigNone
			for i := range summary.writes {
				k := analysis.evalExpectConfigWriteStatic(&summary.writes[i])
				localEvaluated[i] = k
				allMask |= k
			}
			// A single-write `const holder = { extend: () => {}, reset() { this.extend = r.expect.extend; } }`
			// starts with `expectConfigNone` on its initializer but can acquire
			// `expect.extend` / `expect.addEqualityTesters` / `expect` via direct
			// writes, aliases, mutating builtins, or own-method `this.<prop> = ...`.
			// Running this inside the fixed-point loop ensures mutual cycles are
			// observed by `otherCycleHit` and self-cycles converge cleanly.
			if initObj != nil {
				origObjKind = localEvaluated[0]
				mutSummary := analysis.getObjectLiteralMutationSummary(symbol, initObj)
				if restored := mutSummary.restoredConfigMask; restored != expectConfigNone {
					localEvaluated[0] |= restored
					allMask |= restored
				}
			}
			otherCycleHit = analysis.expectConfigOtherCycleHits[symbol] != otherBefore ||
				analysis.objectLiteralMutationCycleHits != mutCycleBefore ||
				analysis.expectConfigOverwriteCycleHits != overwriteCycleBefore
			if !analysis.expectConfigSelfHit[symbol] || allMask == analysis.expectConfigPartial[symbol] {
				break
			}
			analysis.expectConfigPartial[symbol] = allMask
		}
		if !otherCycleHit {
			if initObj != nil {
				summary.initialObjectLiteralKind = origObjKind
				if localEvaluated[0] != origObjKind {
					summary.writes[0].isShallowCopy = true
				}
			}
			summary.evaluatedSources = true
			summary.allConfigMask = allMask
			summary.hasAnyConfigSource = allMask != expectConfigNone
			for i, k := range localEvaluated {
				summary.writes[i].evaluatedMember = k
				if k != expectConfigNone {
					if summary.configWritesByFn == nil {
						summary.configWritesByFn = map[*ast.Node][]int{}
					}
					fn := summary.writes[i].enclosingFn
					summary.configWritesByFn[fn] = append(summary.configWritesByFn[fn], i)
				}
			}
			for _, indices := range summary.configWritesByFn {
				sort.Slice(indices, func(i, j int) bool {
					return summary.writes[indices[i]].exprEnd < summary.writes[indices[j]].exprEnd
				})
			}
			// Track `const alias = mock` only when `mock` itself is a single-write
			// immutable `const` shallow copy (or chain of such aliases), so a
			// reassigned `let mock` cannot leak stale overwrites onto `alias`.
			if len(summary.writes) == 1 && summary.writes[0].isConstDecl && !summary.writes[0].isShallowCopy && summary.writes[0].rhs != nil {
				if rhsIdent := internalUtils.SkipAssertionsAndParens(summary.writes[0].rhs); rhsIdent != nil && rhsIdent.Kind == ast.KindIdentifier {
					if targetSym := resolveRstestRootSymbol(analysis.ctx, rhsIdent); targetSym != nil && targetSym != symbol {
						if targetInfo, ok := analysis.expectConfigSymbolInfo[targetSym]; ok &&
							len(targetInfo.writes) == 1 && targetInfo.writes[0].isConstDecl &&
							(targetInfo.writes[0].isShallowCopy || targetInfo.shallowAliasTarget != nil) {
							summary.shallowAliasTarget = targetSym
						}
					}
				}
			}
			localEvaluated = nil
		}
	} else {
		allMask = summary.allConfigMask
	}

	if allMask == expectConfigNone {
		if summary.evaluatedSources {
			analysis.cacheExpectConfigAlias(symbol, expectConfigNone)
		}
		return expectConfigNone
	}

	// Fast path: a single immutable `const` binding that is neither a shallow
	// copy nor an alias of one has the same classification at every read.
	if summary.evaluatedSources && len(summary.writes) == 1 && summary.writes[0].isConstDecl &&
		!summary.writes[0].isShallowCopy && summary.shallowAliasTarget == nil {
		member := summary.writes[0].evaluatedMember
		analysis.cacheExpectConfigAlias(symbol, member)
		return member
	}

	readStmt, readBlock := enclosingStatementAndList(node)
	readFn := ast.GetContainingFunction(node)

	// Fast O(log W) path when summary is latched: check same-function writes
	// starting strictly after the last unconditional kill in readBlock, then
	// check any other-function writes only if bits remain unset.
	if summary.evaluatedSources {
		var best expectConfigKind
		lastKillPos := lastUnconditionalKillBefore(summary.blockKills[readBlock], readStmt)
		if sameFnIndices := summary.configWritesByFn[readFn]; len(sameFnIndices) > 0 {
			start := 0
			if lastKillPos >= 0 {
				start = sort.Search(len(sameFnIndices), func(i int) bool {
					return summary.writes[sameFnIndices[i]].exprEnd > lastKillPos
				})
			}
			for _, idx := range sameFnIndices[start:] {
				w := &summary.writes[idx]
				if !analysis.writeCanReachRead(symbol, w, node, readStmt, readBlock, readFn, summary) {
					continue
				}
				k := w.evaluatedMember
				if w.isShallowCopy || summary.shallowAliasTarget != nil {
					k &^= analysis.expectConfigMustOverwrittenMask(symbol, node, w)
				}
				best |= k
				if best == allMask {
					return best
				}
			}
		}
		for fn, indices := range summary.configWritesByFn {
			if fn == readFn {
				continue
			}
			for _, idx := range indices {
				w := &summary.writes[idx]
				if !analysis.writeCanReachRead(symbol, w, node, readStmt, readBlock, readFn, summary) {
					continue
				}
				k := w.evaluatedMember
				if w.isShallowCopy || summary.shallowAliasTarget != nil {
					k &^= analysis.expectConfigMustOverwrittenMask(symbol, node, w)
				}
				best |= k
				if best == allMask {
					return best
				}
			}
		}
		return best
	}

	// Uncached fallback (only during an active cycle).
	var best expectConfigKind
	for i := range summary.writes {
		k := localEvaluated[i]
		if k == expectConfigNone {
			continue
		}
		w := &summary.writes[i]
		if !analysis.writeCanReachRead(symbol, w, node, readStmt, readBlock, readFn, summary) {
			continue
		}
		if w.isShallowCopy || summary.shallowAliasTarget != nil {
			k &^= analysis.expectConfigMustOverwrittenMask(symbol, node, w)
		}
		best |= k
		if best == allMask {
			return best
		}
	}
	return best
}

func lastUnconditionalKillBefore(kills []expectConfigBlockKill, readStmt *ast.Node) int {
	if len(kills) == 0 || readStmt == nil {
		return -1
	}
	readPos := readStmt.Pos()
	// `kills` is sorted by `pos` (monotone with `stmtEnd` across distinct
	// statements). Find the rightmost kill with `stmtEnd <= readPos`.
	idx := sort.Search(len(kills), func(i int) bool {
		return kills[i].pos >= readPos
	})
	for i := idx - 1; i >= 0; i-- {
		if kills[i].stmtEnd <= readPos {
			return kills[i].pos
		}
	}
	return -1
}

func (analysis *RstestCallAnalysis) cacheExpectConfigAlias(symbol *ast.Symbol, member expectConfigKind) {
	if analysis.expectConfigAliases == nil {
		analysis.expectConfigAliases = map[*ast.Symbol]expectConfigKind{}
	}
	analysis.expectConfigAliases[symbol] = member
}

func (analysis *RstestCallAnalysis) evalExpectConfigWriteStatic(write *expectConfigWrite) expectConfigKind {
	var kind expectConfigKind
	if write.initialKind != expectConfigNone {
		kind |= write.initialKind
	}
	if write.bindingElem != nil {
		kind |= analysis.destructuredExpectConfigMember(write.bindingElem)
	}
	if write.assignDestruct != nil {
		kind |= analysis.assignmentDestructuredExpectConfigMember(write.assignDestruct, write.assignDefault)
	}
	if write.rhs != nil {
		kind |= analysis.expectConfigMember(write.rhs)
	}
	return kind
}

func isRestBindingElement(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindBindingElement && node.AsBindingElement().DotDotDotToken != nil
}

func isObjectLiteralSpread(node *ast.Node) bool {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		return true
	}
	// `Object.assign({}, expect)` allocates a fresh shallow copy ONLY when
	// arg 0 is a fresh accessor-free object literal (never an existing shared
	// target or a target literal with getters/setters/spread/`__proto__` that
	// could execute user code during `Object.assign`).
	if node.Kind == ast.KindCallExpression {
		callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if name, ok := internalUtils.AccessExpressionStaticName(callee); ok && name == "assign" {
				if recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee)); recv != nil &&
					recv.Kind == ast.KindIdentifier && recv.AsIdentifier().Text == "Object" {
					args := node.Arguments()
					if len(args) > 0 {
						if first := internalUtils.SkipAssertionsAndParens(args[0]); first != nil &&
							first.Kind == ast.KindObjectLiteralExpression {
							if props := first.AsObjectLiteralExpression().Properties; props != nil {
								for _, prop := range props.Nodes {
									if prop == nil {
										continue
									}
									if prop.Kind == ast.KindGetAccessor || prop.Kind == ast.KindSetAccessor ||
										prop.Kind == ast.KindSpreadAssignment {
										return false
									}
									if prop.Kind == ast.KindPropertyAssignment && prop.Name() != nil &&
										prop.Name().Kind != ast.KindComputedPropertyName {
										if pName, pOK := internalUtils.GetStaticPropertyName(prop.Name()); pOK && pName == "__proto__" {
											return false
										}
									}
								}
							}
							return true
						}
					}
				}
			}
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) getExpectConfigSymbolSummary(symbol *ast.Symbol) *expectConfigSymbolSummary {
	if summary, ok := analysis.expectConfigSymbolInfo[symbol]; ok {
		return summary
	}
	summary := &expectConfigSymbolSummary{}
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		switch declaration.Kind {
		case ast.KindVariableDeclaration:
			init := declaration.AsVariableDeclaration().Initializer
			if init == nil {
				continue
			}
			isConst := declaration.Parent != nil &&
				declaration.Parent.Kind == ast.KindVariableDeclarationList &&
				declaration.Parent.Flags&ast.NodeFlagsConst != 0
			shallow := isObjectLiteralSpread(init)
			if shallow {
				summary.isDirectShallowCopy = true
			}
			stmt, block, unconditional := writeStatementInfo(declaration)
			summary.writes = append(summary.writes, expectConfigWrite{
				pos:           declaration.Pos(),
				exprEnd:       declaration.End(),
				stmt:          stmt,
				block:         block,
				enclosingFn:   ast.GetContainingFunction(declaration),
				unconditional: unconditional,
				isConstDecl:   isConst,
				isShallowCopy: shallow,
				rhs:           init,
			})
		case ast.KindBindingElement:
			isConst := false
			exprEnd := declaration.End()
			if variable := internalUtils.EnclosingVariableDeclarationOfBindingElement(declaration); variable != nil {
				// Use the enclosing VariableDeclaration's End so writes inside
				// the RHS (`var { extend: fn } = ((fn = () => {}), expect)`) do
				// not appear after the destructuring binding's exprEnd.
				exprEnd = variable.End()
				if variable.Parent != nil &&
					variable.Parent.Kind == ast.KindVariableDeclarationList &&
					variable.Parent.Flags&ast.NodeFlagsConst != 0 {
					isConst = true
				}
			}
			shallow := isRestBindingElement(declaration)
			if !shallow {
				if be := declaration.AsBindingElement(); be != nil && be.Initializer != nil && isObjectLiteralSpread(be.Initializer) {
					shallow = true
				}
			}
			if shallow {
				summary.isDirectShallowCopy = true
			}
			stmt, block, unconditional := writeStatementInfo(declaration)
			summary.writes = append(summary.writes, expectConfigWrite{
				pos:           declaration.Pos(),
				exprEnd:       exprEnd,
				stmt:          stmt,
				block:         block,
				enclosingFn:   ast.GetContainingFunction(declaration),
				unconditional: unconditional,
				isConstDecl:   isConst,
				isShallowCopy: shallow,
				bindingElem:   declaration,
			})
		case ast.KindParameter:
			param := declaration.AsParameterDeclaration()
			var rhs *ast.Node
			if param != nil && param.Initializer != nil {
				rhs = param.Initializer
			}
			var initKind expectConfigKind
			if analysis.callbacksRef().ContextParameters[declaration] {
				initKind = expectConfigReceiver
			}
			stmt, block, unconditional := writeStatementInfo(declaration)
			summary.writes = append(summary.writes, expectConfigWrite{
				pos:           declaration.Pos(),
				exprEnd:       declaration.End(),
				stmt:          stmt,
				block:         block,
				enclosingFn:   ast.GetContainingFunction(declaration),
				unconditional: unconditional,
				rhs:           rhs,
				initialKind:   initKind,
			})
		case ast.KindImportSpecifier, ast.KindImportClause, ast.KindNamespaceImport, ast.KindFunctionDeclaration, ast.KindClassDeclaration:
			stmt, block, unconditional := writeStatementInfo(declaration)
			summary.writes = append(summary.writes, expectConfigWrite{
				pos:           declaration.Pos(),
				exprEnd:       declaration.End(),
				stmt:          stmt,
				block:         block,
				enclosingFn:   ast.GetContainingFunction(declaration),
				unconditional: unconditional,
			})
		}
	}
	if analysis.ctx.Refs != nil {
		for _, reference := range analysis.ctx.Refs.References(symbol) {
			if !internalUtils.IsWriteReference(reference) {
				continue
			}
			if parent := reference.Parent; parent != nil &&
				((parent.Kind == ast.KindVariableDeclaration && parent.Name() == reference) ||
					(parent.Kind == ast.KindBindingElement && parent.Name() == reference)) {
				continue
			}
			var rhs *ast.Node
			var assignDestruct *ast.Node
			var assignDefault *ast.Node
			exprEnd := reference.End()
			outer := internalUtils.OutermostParenthesizedExpression(reference)
			if outer != nil && outer.Parent != nil {
				switch outer.Parent.Kind {
				case ast.KindBinaryExpression:
					bin := outer.Parent.AsBinaryExpression()
					if bin != nil && bin.OperatorToken != nil && bin.Left == outer {
						switch bin.OperatorToken.Kind {
						case ast.KindEqualsToken, ast.KindQuestionQuestionEqualsToken,
							ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
							rhs = bin.Right
							exprEnd = outer.Parent.End()
							// Handle destructuring assignment with default:
							// `({ extend: fn = () => {} } = expect)`
							if bin.OperatorToken.Kind == ast.KindEqualsToken &&
								outer.Parent.Parent != nil &&
								outer.Parent.Parent.Kind == ast.KindPropertyAssignment &&
								outer.Parent.Parent.AsPropertyAssignment().Initializer == outer.Parent {
								assignDestruct = outer.Parent.Parent
								assignDefault = bin.Right
								rhs = nil
								exprEnd = outer.Parent.Parent.End()
							}
						}
					}
				case ast.KindShorthandPropertyAssignment:
					assignDestruct = outer.Parent
					exprEnd = enclosingDestructuringAssignmentEnd(outer.Parent)
					if sp := outer.Parent.AsShorthandPropertyAssignment(); sp != nil && sp.ObjectAssignmentInitializer != nil {
						assignDefault = sp.ObjectAssignmentInitializer
					}
				case ast.KindPropertyAssignment, ast.KindSpreadAssignment:
					assignDestruct = outer.Parent
					exprEnd = enclosingDestructuringAssignmentEnd(outer.Parent)
				}
			}
			if assignDestruct != nil && exprEnd < assignDestruct.End() {
				exprEnd = assignDestruct.End()
			}
			if assignDestruct != nil {
				if outerEnd := enclosingDestructuringAssignmentEnd(assignDestruct); outerEnd > exprEnd {
					exprEnd = outerEnd
				}
			}
			shallow := isObjectLiteralSpread(rhs) || isObjectLiteralSpread(assignDefault) || (assignDestruct != nil && assignDestruct.Kind == ast.KindSpreadAssignment)
			if shallow {
				summary.isDirectShallowCopy = true
			}
			stmt, block, unconditional := writeStatementInfo(reference)
			summary.writes = append(summary.writes, expectConfigWrite{
				pos:            reference.Pos(),
				exprEnd:        exprEnd,
				stmt:           stmt,
				block:          block,
				enclosingFn:    ast.GetContainingFunction(reference),
				unconditional:  unconditional,
				isShallowCopy:  shallow,
				assignDestruct: assignDestruct,
				assignDefault:  assignDefault,
				rhs:            rhs,
			})
		}
	}
	// Index unconditional writes per block in sorted order so kill checks run in O(log W),
	// and record write span/enclosing-function metadata for post-write memoization.
	summary.singleEnclosingFnOK = len(summary.writes) > 0
	for i := range summary.writes {
		w := &summary.writes[i]
		if w.exprEnd > summary.maxWriteEnd {
			summary.maxWriteEnd = w.exprEnd
		}
		if i == 0 {
			summary.singleEnclosingFn = w.enclosingFn
		} else if w.enclosingFn != summary.singleEnclosingFn {
			summary.singleEnclosingFnOK = false
		}
		if !w.unconditional || w.block == nil || w.stmt == nil {
			continue
		}
		if summary.blockKills == nil {
			summary.blockKills = map[*ast.Node][]expectConfigBlockKill{}
		}
		summary.blockKills[w.block] = append(summary.blockKills[w.block], expectConfigBlockKill{
			pos:     w.pos,
			stmtEnd: w.stmt.End(),
		})
	}
	for _, kills := range summary.blockKills {
		sort.Slice(kills, func(i, j int) bool {
			return kills[i].pos < kills[j].pos
		})
	}
	if analysis.expectConfigSymbolInfo == nil {
		analysis.expectConfigSymbolInfo = map[*ast.Symbol]*expectConfigSymbolSummary{}
	}
	analysis.expectConfigSymbolInfo[symbol] = summary
	return summary
}

// enclosingDestructuringAssignmentEnd walks up from a destructuring target
// property/element to its enclosing assignment `BinaryExpression(=)` and
// returns the assignment's `End()`, so writes inside the assignment's RHS
// (evaluated before the destructuring target is written) never kill it.
func enclosingDestructuringAssignmentEnd(node *ast.Node) int {
	if node == nil {
		return 0
	}
	cur := node
	for cur != nil && cur.Parent != nil {
		parent := cur.Parent
		switch parent.Kind {
		case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
			ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
			ast.KindSpreadAssignment, ast.KindSpreadElement, ast.KindParenthesizedExpression:
			cur = parent
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Left == cur {
				if internalUtils.IsDefaultValueInDestructuringAssignment(parent) {
					cur = parent
					continue
				}
				return parent.End()
			}
			return node.End()
		default:
			return node.End()
		}
	}
	return node.End()
}

// expectConfigMustOverwrittenMask returns the bitmask of config properties
// (`expectConfigExpectExtend`, `expectConfigExpectEquality`,
// `expectConfigReceiver`) that were UNCONDITIONALLY overwritten or deleted on
// `symbol` (or its immutable aliased shallow-copy target) after `write` and
// strictly before `useNode` without being restored in between.
func (analysis *RstestCallAnalysis) expectConfigMustOverwrittenMask(
	symbol *ast.Symbol,
	useNode *ast.Node,
	write *expectConfigWrite,
) expectConfigKind {
	if analysis.ctx.Refs == nil || symbol == nil || useNode == nil {
		return expectConfigNone
	}
	var chain []*ast.Symbol
	visited := map[*ast.Symbol]bool{}
	for cur := symbol; cur != nil && !visited[cur]; {
		visited[cur] = true
		chain = append(chain, cur)
		if info := analysis.expectConfigSymbolInfo[cur]; info != nil {
			cur = info.shallowAliasTarget
		} else {
			break
		}
	}
	useFn := ast.GetContainingFunction(useNode)
	var mask expectConfigKind
	for _, bit := range []expectConfigKind{expectConfigExpectExtend, expectConfigExpectEquality, expectConfigReceiver} {
		restoredByFn := false
		hasLatest := false
		var latestEv expectConfigOverwriteEvent
		for i, cur := range chain {
			summary := analysis.getExpectConfigOverwriteSummary(cur)
			for _, fn := range summary.fnRestores[bit] {
				if fn == useFn || nodeContains(fn, useNode) {
					continue
				}
				if analysis.functionCanExecuteBefore(fn, useNode, cur, nil) {
					restoredByFn = true
					break
				}
			}
			if restoredByFn {
				break
			}
			var minBlock *ast.Node
			minPos := -1
			if i == 0 && write != nil {
				minBlock = write.block
				minPos = write.exprEnd
			}
			ev, found := analysis.latestSymbolOverwriteEvent(summary, bit, useNode, cur, minBlock, minPos)
			if !found {
				continue
			}
			if !hasLatest || isLaterOverwriteEvent(ev, latestEv) {
				latestEv = ev
				hasLatest = true
			}
		}
		if !restoredByFn && hasLatest && latestEv.clears {
			mask |= bit
		}
	}
	return mask
}

func isLaterOverwriteEvent(a, b expectConfigOverwriteEvent) bool {
	if a.stmtEnd != b.stmtEnd {
		return a.stmtEnd > b.stmtEnd
	}
	// Within the same statement, a restore (`!clears`) always dominates an overwrite (`clears`).
	if a.clears != b.clears {
		return !a.clears
	}
	return a.pos > b.pos
}

func (analysis *RstestCallAnalysis) latestSymbolOverwriteEvent(
	summary *expectConfigOverwriteSummary,
	bit expectConfigKind,
	useNode *ast.Node,
	symbol *ast.Symbol,
	writeBlock *ast.Node,
	writeExprEnd int,
) (expectConfigOverwriteEvent, bool) {
	if summary == nil || len(summary.events) == 0 {
		return expectConfigOverwriteEvent{}, false
	}
	// Walk upward from useNode across enclosing statements/blocks within the
	// same function (never crossing a function boundary into an outer scope).
	cur := useNode
	for cur != nil {
		stmt, block := enclosingStatementAndList(cur)
		if stmt == nil || block == nil {
			break
		}
		stmtPos := stmt.Pos()
		stmtEnd := stmt.End()
		key := expectConfigOverwriteKey{block: block, bit: bit}
		events := summary.events[key]
		if len(events) > 0 {
			// Find the latest event whose statement finished before `stmtPos`.
			idx := sort.Search(len(events), func(i int) bool {
				return events[i].stmtEnd > stmtPos
			})
			if idx > 0 {
				ev := events[idx-1]
				if block == writeBlock && ev.pos < writeExprEnd {
					// Overwrite happened before this reaching write (e.g. before
					// `mock = { ...expect }` replaced the object).
					return expectConfigOverwriteEvent{}, false
				}
				if ev.clears {
					// A restore in the SAME statement (e.g. computed key
					// `holder[(holder.extend = r.expect.extend, 'extend')]()`)
					// can execute before `useNode` unconditionally. Because
					// `events` sorts `!clears` last within each `stmtEnd`,
					// the last event with `stmtEnd <= stmt.End()` reveals any
					// same-statement restore in O(log E).
					sameStmtEndIdx := sort.Search(len(events), func(i int) bool {
						return events[i].stmtEnd > stmtEnd
					})
					allowSkipSelfPostCall := cur == useNode && !ast.IsIterationStatement(stmt, false)
					for k := sameStmtEndIdx - 1; k >= idx && events[k].stmtEnd == stmtEnd && !events[k].clears; k-- {
						if allowSkipSelfPostCall && events[k].postCallNode == useNode {
							continue
						}
						return events[k], true
					}
					// A restore in a strictly later statement cancels the clear
					// whenever `useNode` sits inside a loop or capturing function
					// that can re-execute after that restore.
					if maxRestoreEnd := summary.maxRestoreEnd[key]; maxRestoreEnd > stmtEnd &&
						!directRangeCannotExecuteBefore(stmtEnd, maxRestoreEnd, useNode, symbol) {
						return expectConfigOverwriteEvent{pos: maxRestoreEnd, stmtEnd: maxRestoreEnd, clears: false}, true
					}
				}
				return ev, true
			}
		}
		if block == writeBlock || block.Parent == nil || testFramework.IsFunction(block.Parent) ||
			block.Kind == ast.KindSourceFile || block.Kind == ast.KindModuleBlock {
			break
		}
		cur = block
	}
	return expectConfigOverwriteEvent{}, false
}

type expectConfigOverwriteEventDedupKey struct {
	block        *ast.Node
	bit          expectConfigKind
	pos          int
	stmtEnd      int
	clears       bool
	postCallNode *ast.Node
}

func (analysis *RstestCallAnalysis) isAccessorDefiningBuiltinCall(call *ast.Node) (isAssign bool, isDefineOrCreate bool) {
	if call == nil || call.Kind != ast.KindCallExpression {
		return false, false
	}
	callee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return false, false
	}
	method, ok := analysis.evalExpectConfigAccessName(callee)
	if !ok {
		return false, false
	}
	obj := internalUtils.SkipAssertionsAndParens(callee.Expression())
	if obj == nil || obj.Kind != ast.KindIdentifier {
		return false, false
	}
	objName := obj.AsIdentifier().Text
	if objName != "Object" && objName != "Reflect" {
		return false, false
	}
	if analysis.ctx.Refs != nil {
		if !analysis.ctx.Refs.IsGlobalReference(obj) {
			return false, false
		}
	} else if internalUtils.IsShadowed(obj, objName) {
		return false, false
	}
	if objName == "Object" {
		switch method {
		case "assign":
			return true, false
		case "defineProperty", "defineProperties", "create", "setPrototypeOf":
			return false, true
		}
	}
	if objName == "Reflect" {
		if method == "defineProperty" || method == "setPrototypeOf" {
			return false, true
		}
	}
	return false, false
}

// symbolHasDefineProperty reports whether `symbol` (or a tracked alias of it)
// is passed as the target of `Object.defineProperty`, `Object.defineProperties`,
// `Reflect.defineProperty`, `Object.setPrototypeOf`, `Reflect.setPrototypeOf`,
// `__defineGetter__`, or `__defineSetter__`.
func (analysis *RstestCallAnalysis) symbolHasDefineProperty(symbol *ast.Symbol, visited map[*ast.Symbol]bool) bool {
	if symbol == nil || analysis.ctx.Refs == nil {
		return false
	}
	if cached, ok := analysis.symbolDefinePropertyAccessor[symbol]; ok {
		return cached
	}
	if visited[symbol] {
		return false
	}
	isRoot := visited == nil
	if visited == nil {
		visited = map[*ast.Symbol]bool{}
	}
	visited[symbol] = true

	found := false
	for _, ref := range analysis.ctx.Refs.References(symbol) {
		if ref == nil {
			continue
		}
		if parent := ref.Parent; parent != nil &&
			((parent.Kind == ast.KindVariableDeclaration && parent.Name() == ref) ||
				(parent.Kind == ast.KindBindingElement && parent.Name() == ref)) {
			continue
		}
		outer := outermostExpressionWrapper(ref)
		if outer == nil || outer.Parent == nil {
			continue
		}
		switch outer.Parent.Kind {
		case ast.KindCallExpression:
			args := outer.Parent.Arguments()
			if len(args) > 0 && internalUtils.SkipAssertionsAndParens(args[0]) == internalUtils.SkipAssertionsAndParens(outer) {
				if _, isDef := analysis.isAccessorDefiningBuiltinCall(outer.Parent); isDef {
					found = true
				}
			}
		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			if outer.Parent.Expression() == outer {
				if mName, ok := analysis.evalExpectConfigAccessName(outer.Parent); ok &&
					(mName == "__defineGetter__" || mName == "__defineSetter__" ||
						(mName == "__proto__" && internalUtils.IsWriteReference(outer.Parent))) {
					found = true
				}
			}
		case ast.KindVariableDeclaration:
			vd := outer.Parent.AsVariableDeclaration()
			if vd != nil && vd.Initializer == outer && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
				aliasSym := outer.Parent.Symbol()
				if aliasSym == nil {
					aliasSym = resolveRstestRootSymbol(analysis.ctx, vd.Name())
				}
				if aliasSym != nil && aliasSym != symbol && analysis.symbolHasDefineProperty(aliasSym, visited) {
					found = true
				}
			}
		case ast.KindBinaryExpression:
			bin := outer.Parent.AsBinaryExpression()
			if bin != nil && bin.Right == outer && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
				if lhs := internalUtils.SkipAssertionsAndParens(bin.Left); lhs != nil && lhs.Kind == ast.KindIdentifier {
					if aliasSym := resolveRstestRootSymbol(analysis.ctx, lhs); aliasSym != nil && aliasSym != symbol &&
						analysis.symbolHasDefineProperty(aliasSym, visited) {
						found = true
					}
				}
			}
		}
		if found {
			break
		}
	}
	if isRoot {
		if analysis.symbolDefinePropertyAccessor == nil {
			analysis.symbolDefinePropertyAccessor = map[*ast.Symbol]bool{}
		}
		analysis.symbolDefinePropertyAccessor[symbol] = found
	}
	return found
}

func (analysis *RstestCallAnalysis) exprMayHaveAccessorOnProp(
	expr *ast.Node,
	propName string,
	propStatic bool,
	depth int,
) bool {
	if depth > 4 {
		return true
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindObjectLiteralExpression:
		return analysis.objectLiteralHasAccessor(expr, propName, propStatic, true, 0)
	case ast.KindIdentifier:
		if targetSym := resolveRstestRootSymbol(analysis.ctx, expr); targetSym != nil {
			return analysis.symbolMayHaveAccessorOnProp(targetSym, propName, propStatic, depth+1)
		}
	case ast.KindCallExpression:
		isAssign, isDef := analysis.isAccessorDefiningBuiltinCall(expr)
		if isDef {
			return true
		}
		if isAssign {
			args := expr.Arguments()
			if len(args) > 0 {
				return analysis.exprMayHaveAccessorOnProp(args[0], propName, propStatic, depth+1)
			}
		}
	case ast.KindConditionalExpression:
		cond := expr.AsConditionalExpression()
		if cond != nil {
			return analysis.exprMayHaveAccessorOnProp(cond.WhenTrue, propName, propStatic, depth+1) ||
				analysis.exprMayHaveAccessorOnProp(cond.WhenFalse, propName, propStatic, depth+1)
		}
	case ast.KindBinaryExpression:
		bin := expr.AsBinaryExpression()
		if bin != nil && bin.OperatorToken != nil {
			switch bin.OperatorToken.Kind {
			case ast.KindCommaToken, ast.KindAmpersandAmpersandToken, ast.KindEqualsToken:
				return analysis.exprMayHaveAccessorOnProp(bin.Right, propName, propStatic, depth+1)
			case ast.KindBarBarToken, ast.KindQuestionQuestionToken:
				return analysis.exprMayHaveAccessorOnProp(bin.Left, propName, propStatic, depth+1) ||
					analysis.exprMayHaveAccessorOnProp(bin.Right, propName, propStatic, depth+1)
			}
		}
	}
	return false
}

// symbolMayHaveAccessorOnProp reports whether any write, alias, `Object.assign`,
// or `Object.defineProperty` call on `symbol` may define a getter/setter accessor
// for `propName`. Covers `let` bindings, aliases, and dynamic descriptor installs.
func (analysis *RstestCallAnalysis) symbolMayHaveAccessorOnProp(
	symbol *ast.Symbol,
	propName string,
	propStatic bool,
	depth int,
) bool {
	if depth > 4 || symbol == nil {
		return depth > 4
	}
	if analysis.symbolHasDefineProperty(symbol, nil) {
		return true
	}
	symSummary := analysis.getExpectConfigSymbolSummary(symbol)
	for i := range symSummary.writes {
		rhs := internalUtils.SkipAssertionsAndParens(symSummary.writes[i].rhs)
		if rhs == nil {
			continue
		}
		if rhs.Kind == ast.KindIdentifier {
			if targetSym := resolveRstestRootSymbol(analysis.ctx, rhs); targetSym == symbol {
				continue
			}
		}
		if analysis.exprMayHaveAccessorOnProp(rhs, propName, propStatic, depth+1) {
			return true
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) isNonEscapingSymbolUse(parent, current *ast.Node, symHasAnyAccessor bool) bool {
	if parent == nil {
		return true
	}
	if internalUtils.IsWriteReference(current) {
		return true
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration:
		vd := parent.AsVariableDeclaration()
		return vd != nil && vd.Initializer == current && vd.Name() != nil &&
			(vd.Name().Kind == ast.KindIdentifier || vd.Name().Kind == ast.KindObjectBindingPattern)
	case ast.KindBinaryExpression:
		bin := parent.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil {
			return false
		}
		switch bin.OperatorToken.Kind {
		case ast.KindEqualsToken:
			if bin.Right != current {
				return false
			}
			// Only discarded top-level assignment statements do not leak the
			// expression value (`(a = holder).extend = ...`, `b = a = holder`).
			binOuter := outermostExpressionWrapper(parent)
			if binOuter == nil || binOuter.Parent == nil || binOuter.Parent.Kind != ast.KindExpressionStatement {
				return false
			}
			lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
			if lhs == nil {
				return false
			}
			if lhs.Kind == ast.KindObjectLiteralExpression {
				return true
			}
			if lhs.Kind == ast.KindIdentifier {
				return resolveRstestRootSymbol(analysis.ctx, lhs) != nil
			}
			return false
		case ast.KindEqualsEqualsEqualsToken, ast.KindExclamationEqualsEqualsToken:
			return true
		case ast.KindInKeyword:
			// `'k' in holder` only checks property presence; `holder in {}`
			// coerces `holder` via `ToPropertyKey` (`toString`/`valueOf`).
			return bin.Right == current
		}
		return false
	case ast.KindTypeOfExpression, ast.KindVoidExpression, ast.KindExpressionStatement:
		return true
	case ast.KindPrefixUnaryExpression:
		return parent.AsPrefixUnaryExpression().Operator == ast.KindExclamationToken
	case ast.KindIfStatement:
		return parent.AsIfStatement().Expression == current
	case ast.KindWhileStatement:
		return parent.AsWhileStatement().Expression == current
	case ast.KindDoStatement:
		return parent.AsDoStatement().Expression == current
	case ast.KindConditionalExpression:
		return parent.AsConditionalExpression().Condition == current
	case ast.KindSpreadAssignment:
		return !symHasAnyAccessor
	}
	return false
}

func (analysis *RstestCallAnalysis) objectLiteralHasNonPureProp(objNode *ast.Node, propName string) bool {
	if objNode == nil || objNode.Kind != ast.KindObjectLiteralExpression || objNode.AsObjectLiteralExpression().Properties == nil {
		return false
	}
	for _, prop := range objNode.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		if prop.Kind == ast.KindSpreadAssignment {
			spreadExpr := internalUtils.SkipAssertionsAndParens(prop.AsSpreadAssignment().Expression)
			if analysis.isPureRstestExpectObject(spreadExpr, 0) {
				continue
			}
			if spreadExpr != nil && spreadExpr.Kind == ast.KindObjectLiteralExpression {
				if analysis.objectLiteralHasNonPureProp(spreadExpr, propName) {
					return true
				}
				continue
			}
			return true
		}
		if prop.Name() != nil {
			n, nOK := analysis.staticStringEvaluator().EvalPropertyName(prop.Name())
			if nOK && n != propName {
				continue
			}
			switch prop.Kind {
			case ast.KindMethodDeclaration:
				if !isSideEffectFreeNoOpFn(prop) {
					return true
				}
			case ast.KindPropertyAssignment:
				rhs := prop.AsPropertyAssignment().Initializer
				if !analysis.isPureRstestExpectCallable(rhs, 0) && !isSideEffectFreeNoOpFn(rhs) {
					return true
				}
			case ast.KindShorthandPropertyAssignment:
				ident := prop.AsShorthandPropertyAssignment().Name()
				if !analysis.isPureRstestExpectCallable(ident, 0) {
					return true
				}
			default:
				return true
			}
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) symbolHasNonPurePropWrite(
	symbol *ast.Symbol,
	propName string,
	depth int,
	visited map[*ast.Symbol]bool,
) bool {
	if depth > 4 || symbol == nil || analysis.ctx.Refs == nil {
		return depth > 4
	}
	if visited[symbol] {
		return false
	}
	if visited == nil {
		visited = map[*ast.Symbol]bool{}
	}
	visited[symbol] = true

	symSummary := analysis.getExpectConfigSymbolSummary(symbol)
	for i := range symSummary.writes {
		if rhs := internalUtils.SkipAssertionsAndParens(symSummary.writes[i].rhs); rhs != nil &&
			rhs.Kind == ast.KindObjectLiteralExpression {
			if analysis.objectLiteralHasNonPureProp(rhs, propName) {
				return true
			}
		}
	}

	for _, ref := range analysis.ctx.Refs.References(symbol) {
		if ref == nil {
			continue
		}
		outer := outermostExpressionWrapper(ref)
		if outer == nil || outer.Parent == nil {
			continue
		}
		if ast.IsAccessExpression(outer.Parent) && outer.Parent.Expression() == outer {
			if !internalUtils.IsWriteReference(outer.Parent) {
				continue
			}
			pName, pKnown := analysis.evalExpectConfigAccessName(outer.Parent)
			if pKnown && pName != propName {
				continue
			}
			accessOuter := outermostExpressionWrapper(outer.Parent)
			if accessOuter != nil && accessOuter.Parent != nil && accessOuter.Parent.Kind == ast.KindBinaryExpression {
				bin := accessOuter.Parent.AsBinaryExpression()
				if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken &&
					bin.Left == accessOuter && !internalUtils.IsDefaultValueInDestructuringAssignment(accessOuter.Parent) &&
					(analysis.isPureRstestExpectCallable(bin.Right, 0) || isSideEffectFreeNoOpFn(bin.Right)) {
					continue
				}
			}
			return true
		}
		switch outer.Parent.Kind {
		case ast.KindCallExpression:
			args := outer.Parent.Arguments()
			if len(args) > 0 && internalUtils.SkipAssertionsAndParens(args[0]) == internalUtils.SkipAssertionsAndParens(outer) {
				callee := internalUtils.SkipAssertionsAndParens(outer.Parent.AsCallExpression().Expression)
				isAssign, _ := analysis.isAccessorDefiningBuiltinCall(outer.Parent)
				if isAssign {
					for _, src := range args[1:] {
						s := internalUtils.SkipAssertionsAndParens(src)
						if s == nil {
							continue
						}
						if s.Kind == ast.KindObjectLiteralExpression {
							if analysis.objectLiteralHasNonPureProp(s, propName) {
								return true
							}
						} else if !analysis.isPureRstestExpectObject(s, 0) {
							return true
						}
					}
				} else if analysis.isGlobalBuiltinMemberCall(callee, "Reflect", "set") {
					if len(args) < 3 {
						return true
					}
					kName, kOK := analysis.staticStringEvaluator().EvalToString(internalUtils.SkipAssertionsAndParens(args[1]))
					if !kOK || kName == propName {
						if !analysis.isPureRstestExpectCallable(args[2], 0) && !isSideEffectFreeNoOpFn(args[2]) {
							return true
						}
					}
				}
			}
		case ast.KindVariableDeclaration:
			vd := outer.Parent.AsVariableDeclaration()
			if vd != nil && vd.Initializer == outer && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
				aliasSym := outer.Parent.Symbol()
				if aliasSym == nil {
					aliasSym = resolveRstestRootSymbol(analysis.ctx, vd.Name())
				}
				if aliasSym != nil && analysis.symbolHasNonPurePropWrite(aliasSym, propName, depth+1, visited) {
					return true
				}
			}
		case ast.KindBinaryExpression:
			bin := outer.Parent.AsBinaryExpression()
			if bin != nil && bin.Right == outer && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
				if lhs := internalUtils.SkipAssertionsAndParens(bin.Left); lhs != nil && lhs.Kind == ast.KindIdentifier {
					if aliasSym := resolveRstestRootSymbol(analysis.ctx, lhs); aliasSym != nil &&
						analysis.symbolHasNonPurePropWrite(aliasSym, propName, depth+1, visited) {
						return true
					}
				}
			}
		}
	}
	if symSummary.shallowAliasTarget != nil && symSummary.shallowAliasTarget != symbol {
		if analysis.symbolHasNonPurePropWrite(symSummary.shallowAliasTarget, propName, depth+1, visited) {
			return true
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) getExpectConfigOverwriteSummary(symbol *ast.Symbol) *expectConfigOverwriteSummary {
	if summary, ok := analysis.expectConfigMustOverwrite[symbol]; ok {
		return summary
	}
	if analysis.expectConfigOverwriteVisit[symbol] {
		analysis.expectConfigOverwriteCycleHits++
		return &expectConfigOverwriteSummary{}
	}
	if analysis.expectConfigOverwriteVisit == nil {
		analysis.expectConfigOverwriteVisit = map[*ast.Symbol]bool{}
	}
	analysis.expectConfigOverwriteVisit[symbol] = true
	defer delete(analysis.expectConfigOverwriteVisit, symbol)

	cycleBefore := analysis.expectConfigOverwriteCycleHits
	expectCycleBefore := analysis.expectConfigCycleHits
	mutCycleBefore := analysis.objectLiteralMutationCycleHits
	summary := &expectConfigOverwriteSummary{}
	seenEvents := map[expectConfigOverwriteEventDedupKey]bool{}
	seenAliases := map[*ast.Symbol]bool{}
	addEvent := func(block *ast.Node, pos, stmtEnd int, bit expectConfigKind, clears bool, postCallNode *ast.Node) {
		if block == nil || bit == expectConfigNone {
			return
		}
		dk := expectConfigOverwriteEventDedupKey{
			block:        block,
			bit:          bit,
			pos:          pos,
			stmtEnd:      stmtEnd,
			clears:       clears,
			postCallNode: postCallNode,
		}
		if seenEvents[dk] {
			return
		}
		seenEvents[dk] = true
		if summary.events == nil {
			summary.events = map[expectConfigOverwriteKey][]expectConfigOverwriteEvent{}
		}
		key := expectConfigOverwriteKey{block: block, bit: bit}
		summary.events[key] = append(summary.events[key], expectConfigOverwriteEvent{
			pos:          pos,
			stmtEnd:      stmtEnd,
			clears:       clears,
			postCallNode: postCallNode,
		})
		if !clears {
			if summary.maxRestoreEnd == nil {
				summary.maxRestoreEnd = map[expectConfigOverwriteKey]int{}
			}
			if stmtEnd > summary.maxRestoreEnd[key] {
				summary.maxRestoreEnd[key] = stmtEnd
			}
		}
	}
	addFnRestore := func(fn *ast.Node, bit expectConfigKind) {
		if fn == nil || bit == expectConfigNone {
			return
		}
		if summary.fnRestores == nil {
			summary.fnRestores = map[expectConfigKind][]*ast.Node{}
		}
		if !slices.Contains(summary.fnRestores[bit], fn) {
			summary.fnRestores[bit] = append(summary.fnRestores[bit], fn)
		}
	}
	recordWithPostCall := func(node, stmt, block *ast.Node, bit expectConfigKind, clears bool, postCallNode *ast.Node) {
		if block == nil || stmt == nil || bit == expectConfigNone {
			return
		}
		if clears {
			addEvent(block, stmt.Pos(), stmt.End(), bit, true, nil)
			return
		}
		// A restore inside a nested block (e.g. `if (cond) { mock.extend = ... }`)
		// must also invalidate earlier overwrites in enclosing blocks of the
		// same function; a restore inside a function is tracked separately so
		// cross-function callers see it too.
		curBlock := block
		curStmt := stmt
		for curBlock != nil && curStmt != nil {
			addEvent(curBlock, curStmt.Pos(), curStmt.End(), bit, false, postCallNode)
			if curBlock.Parent == nil || testFramework.IsFunction(curBlock.Parent) ||
				curBlock.Kind == ast.KindSourceFile || curBlock.Kind == ast.KindModuleBlock {
				break
			}
			curStmt, curBlock = enclosingStatementAndList(curBlock)
		}
		if fn := ast.GetContainingFunction(node); fn != nil {
			addFnRestore(fn, bit)
		}
	}
	record := func(node, stmt, block *ast.Node, bit expectConfigKind, clears bool) {
		recordWithPostCall(node, stmt, block, bit, clears, nil)
	}
	recordAllRestores := func(node, stmt, block *ast.Node) {
		recordWithPostCall(node, stmt, block, expectConfigExpectExtend, false, nil)
		recordWithPostCall(node, stmt, block, expectConfigExpectEquality, false, nil)
		recordWithPostCall(node, stmt, block, expectConfigReceiver, false, nil)
	}
	recordAllPostCallRestores := func(node, stmt, block *ast.Node) {
		recordWithPostCall(node, stmt, block, expectConfigExpectExtend, false, node)
		recordWithPostCall(node, stmt, block, expectConfigExpectEquality, false, node)
		recordWithPostCall(node, stmt, block, expectConfigReceiver, false, node)
	}
	var constInitObj *ast.Node
	var constMutSummary *objectLiteralMutationSummary
	symSummary := analysis.getExpectConfigSymbolSummary(symbol)
	if len(symSummary.writes) == 1 && symSummary.writes[0].isConstDecl && symSummary.writes[0].rhs != nil {
		if initObj := internalUtils.SkipAssertionsAndParens(symSummary.writes[0].rhs); initObj != nil &&
			initObj.Kind == ast.KindObjectLiteralExpression {
			constInitObj = initObj
			constMutSummary = analysis.getObjectLiteralMutationSummary(symbol, initObj)
			// Only emit synthetic initial clears once `symSummary.evaluatedSources`
			// is latched so an active cycle with zeroed `initialObjectLiteralKind`
			// can never fabricate a bogus clear.
			if symSummary.evaluatedSources && !constMutSummary.escaped &&
				symSummary.writes[0].block != nil && symSummary.writes[0].stmt != nil {
				if missing := constMutSummary.restoredConfigMask &^ symSummary.initialObjectLiteralKind; missing != expectConfigNone {
					for _, bit := range []expectConfigKind{expectConfigExpectExtend, expectConfigExpectEquality, expectConfigReceiver} {
						if missing&bit != 0 {
							addEvent(symSummary.writes[0].block, symSummary.writes[0].exprEnd, symSummary.writes[0].stmt.End(), bit, true, nil)
						}
					}
				}
			}
		}
	}
	symHasAnyAccessor := (constMutSummary != nil && constMutSummary.hasAnyAccessor) ||
		analysis.symbolMayHaveAccessorOnProp(symbol, "", false, 0)
	if analysis.ctx.Refs != nil {
		for _, reference := range analysis.ctx.Refs.References(symbol) {
			if reference == nil {
				continue
			}
			if parent := reference.Parent; parent != nil &&
				((parent.Kind == ast.KindVariableDeclaration && parent.Name() == reference) ||
					(parent.Kind == ast.KindBindingElement && parent.Name() == reference)) {
				continue
			}
			stmt, block, unconditional := writeStatementInfo(reference)
			if stmt == nil || block == nil {
				continue
			}
			// If `symbol` is aliased to another binding (`const alias = mock`
			// or `alias = mock`), merge any restores performed through `alias`
			// so `alias.extend = r.expect.extend` cancels overwrites on `mock`.
			if refOuter := outermostExpressionWrapper(reference); refOuter != nil && refOuter.Parent != nil {
				var aliasSym *ast.Symbol
				switch refOuter.Parent.Kind {
				case ast.KindVariableDeclaration:
					vd := refOuter.Parent.AsVariableDeclaration()
					if vd != nil && vd.Initializer == refOuter && vd.Name() != nil {
						if vd.Name().Kind == ast.KindIdentifier {
							aliasSym = refOuter.Parent.Symbol()
							if aliasSym == nil {
								aliasSym = resolveRstestRootSymbol(analysis.ctx, vd.Name())
							}
						} else if vd.Name().Kind == ast.KindObjectBindingPattern && symHasAnyAccessor {
							if constInitObj == nil || analysis.symbolHasDefineProperty(symbol, nil) ||
								analysis.destructuringPatternMayInvokeGetter(vd.Name(), constInitObj, true) {
								summary.escaped = true
								recordAllRestores(reference, stmt, block)
							}
						}
					}
				case ast.KindBinaryExpression:
					bin := refOuter.Parent.AsBinaryExpression()
					if bin != nil && bin.Right == refOuter && bin.OperatorToken != nil &&
						bin.OperatorToken.Kind == ast.KindEqualsToken {
						if lhs := internalUtils.SkipAssertionsAndParens(bin.Left); lhs != nil {
							if lhs.Kind == ast.KindIdentifier {
								aliasSym = resolveRstestRootSymbol(analysis.ctx, lhs)
							} else if lhs.Kind == ast.KindObjectLiteralExpression && symHasAnyAccessor {
								if constInitObj == nil || analysis.symbolHasDefineProperty(symbol, nil) ||
									analysis.destructuringPatternMayInvokeGetter(lhs, constInitObj, true) {
									summary.escaped = true
									recordAllRestores(reference, stmt, block)
								}
							}
						}
					}
				}
				if aliasSym != nil && aliasSym != symbol && !seenAliases[aliasSym] {
					seenAliases[aliasSym] = true
					aliasSummary := analysis.getExpectConfigOverwriteSummary(aliasSym)
					if aliasSummary.escaped {
						summary.escaped = true
					}
					for key, evs := range aliasSummary.events {
						for _, ev := range evs {
							if !ev.clears {
								addEvent(key.block, ev.pos, ev.stmtEnd, key.bit, false, ev.postCallNode)
							}
						}
					}
					for bit, fns := range aliasSummary.fnRestores {
						for _, fn := range fns {
							addFnRestore(fn, bit)
						}
					}
				}
			}
			current := reference
			var firstAccess *ast.Node
			for current.Parent != nil {
				parent := current.Parent
				if ast.IsParenthesizedExpression(parent) {
					current = parent
					continue
				}
				if expr, ok := internalUtils.TransparentExpression(parent); ok && expr == current {
					current = parent
					continue
				}
				if ast.IsAccessExpression(parent) && parent.Expression() == current {
					if firstAccess == nil {
						firstAccess = parent
					}
					current = parent
					outerAccess := outermostExpressionWrapper(current)
					isDelete := outerAccess != nil && outerAccess.Parent != nil && outerAccess.Parent.Kind == ast.KindDeleteExpression
					isWrite := internalUtils.IsWriteReference(current)
					if isWrite || isDelete {
						// Only direct property overwrites on the shallow copy
						// itself (`mock.extend = ...`, `delete mock.extend`,
						// `api.expect = ...`) affect its bits; nested writes
						// like `api.other.extend = ...` or `api.expect.tag = ...`
						// never overwrite `api.expect` itself.
						if current != firstAccess {
							break
						}
						propName, propStatic := analysis.evalExpectConfigAccessName(firstAccess)
						hasSetterOrAccessor := symHasAnyAccessor &&
							((constInitObj != nil && analysis.objectLiteralHasAccessor(constInitObj, propName, propStatic, true, 0)) ||
								analysis.symbolMayHaveAccessorOnProp(symbol, propName, propStatic, 0))
						if isWrite && hasSetterOrAccessor {
							// Writing through a setter does not overwrite the
							// getter and may execute user code that restores `this.*`.
							summary.escaped = true
							recordAllRestores(reference, stmt, block)
							break
						}
						if !propStatic {
							// Dynamic-key write `(holder as any)[key] = r.expect.extend`
							// can restore any config property at runtime.
							if isWrite {
								recordAllRestores(reference, stmt, block)
							}
							break
						}
						bit := expectConfigOverwriteBit(propName)
						if bit == expectConfigNone {
							break
						}
						if isDelete {
							if unconditional {
								record(reference, stmt, block, bit, true)
							}
							break
						}
						if outerAccess != nil && outerAccess.Parent != nil && outerAccess.Parent.Kind == ast.KindBinaryExpression {
							if bin := outerAccess.Parent.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil && bin.Left == outerAccess {
								switch bin.OperatorToken.Kind {
								case ast.KindEqualsToken:
									if !internalUtils.IsDefaultValueInDestructuringAssignment(outerAccess.Parent) {
										if analysis.overwritePreservesConfigBit(bit, bin.Right) {
											record(reference, stmt, block, bit, false)
										} else if unconditional {
											record(reference, stmt, block, bit, true)
										}
										break
									}
									// Destructuring assignment default: could restore.
									record(reference, stmt, block, bit, false)
								case ast.KindQuestionQuestionEqualsToken, ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken:
									if analysis.overwritePreservesConfigBit(bit, bin.Right) {
										record(reference, stmt, block, bit, false)
									}
								default:
									record(reference, stmt, block, bit, false)
								}
								break
							}
						}
						// Any other write form (e.g. destructuring assignment
						// `({ extend: mock.extend } = r.expect)`) may restore
						// the property rather than unconditionally clearing it.
						record(reference, stmt, block, bit, false)
						break
					}
					// Own method/tagged-template calls and getter reads on any
					// object binding (`const`, `let`, or alias) can clear or
					// restore `this.*` bits or leak `this`.
					if current == firstAccess {
						isCall := outerAccess != nil && outerAccess.Parent != nil &&
							outerAccess.Parent.Kind == ast.KindCallExpression &&
							outerAccess.Parent.AsCallExpression().Expression == outerAccess
						isTagged := outerAccess != nil && outerAccess.Parent != nil &&
							outerAccess.Parent.Kind == ast.KindTaggedTemplateExpression &&
							outerAccess.Parent.AsTaggedTemplateExpression().Tag == outerAccess
						propName, propStatic := analysis.evalExpectConfigAccessName(firstAccess)
						if isCall || isTagged {
							if constInitObj != nil && constMutSummary != nil {
								if isCall && propStatic && !constMutSummary.escaped &&
									((propName == "extend" && constMutSummary.pureExtend) ||
										(propName == "addEqualityTesters" && constMutSummary.pureEquality)) {
									break
								}
								handledStraightLine := false
								if isCall && !ast.IsOptionalChain(outerAccess.Parent) && propStatic &&
									!analysis.hasMethodInvalidatingMutationForCall(symbol, constMutSummary, reference, propName) {
									for _, targetProp := range []string{"extend", "addEqualityTesters", "expect"} {
										rhs, ok := analysis.cachedOwnMethodThisAssign(constMutSummary, constInitObj, propName, targetProp)
										if !ok {
											continue
										}
										handledStraightLine = true
										bit := expectConfigOverwriteBit(targetProp)
										if bit == expectConfigNone {
											continue
										}
										hasTargetAccessor := symHasAnyAccessor &&
											(analysis.objectLiteralHasAccessor(constInitObj, targetProp, true, true, 0) ||
												analysis.symbolMayHaveAccessorOnProp(symbol, targetProp, true, 0))
										if analysis.overwritePreservesConfigBit(bit, rhs) || hasTargetAccessor {
											record(reference, stmt, block, bit, false)
										} else if unconditional && !constMutSummary.escaped &&
											stmt.Kind == ast.KindExpressionStatement &&
											ast.SkipParentheses(stmt.AsExpressionStatement().Expression) == outerAccess.Parent {
											record(reference, stmt, block, bit, true)
										}
									}
								}
								if !handledStraightLine {
									// Any non-straight-line, optional, tagged, or
									// potentially overwritten method call may restore `this.*` or escape `this`.
									summary.escaped = true
									recordAllRestores(reference, stmt, block)
								}
								break
							}
							// Non-const or aliased object: calling `extend`/`addEqualityTesters`
							// still records restores for subsequent statements without
							// retroactively cancelling an immediately preceding clear, while
							// any other method/tagged call may restore or leak `this`.
							if isCall && propStatic && (propName == "extend" || propName == "addEqualityTesters") &&
								!analysis.symbolHasNonPurePropWrite(symbol, propName, 0, nil) {
								recordAllPostCallRestores(reference, stmt, block)
							} else {
								summary.escaped = true
								recordAllRestores(reference, stmt, block)
							}
							break
						}
						if symHasAnyAccessor &&
							((constInitObj != nil && analysis.objectLiteralHasAccessor(constInitObj, propName, propStatic, false, 0)) ||
								analysis.symbolMayHaveAccessorOnProp(symbol, propName, propStatic, 0)) {
							summary.escaped = true
							recordAllRestores(reference, stmt, block)
							break
						}
					}
					continue
				}
				if firstAccess == nil {
					if parent.Kind == ast.KindCallExpression {
						isPureDirect, isMemberOnSym, mName := analysis.isPureThisArgInvocation(parent, current, symbol)
						if isPureDirect || (isMemberOnSym && constInitObj != nil && constMutSummary != nil && !constMutSummary.escaped &&
							((mName == "extend" && constMutSummary.pureExtend) ||
								(mName == "addEqualityTesters" && constMutSummary.pureEquality))) {
							break
						}
						isAssignCall, _ := analysis.isAccessorDefiningBuiltinCall(parent)
						isReflectSet := analysis.isGlobalBuiltinMemberCall(
							internalUtils.SkipAssertionsAndParens(parent.AsCallExpression().Expression),
							"Reflect",
							"set",
						)
						changes := analysis.mutatingBuiltinOverwriteChanges(parent, current)
						if len(changes) > 0 && ((!isAssignCall && !isReflectSet) || !symHasAnyAccessor) {
							for bit, clears := range changes {
								if !clears {
									record(reference, stmt, block, bit, false)
								} else if unconditional {
									bitProp := ""
									switch bit {
									case expectConfigExpectExtend:
										bitProp = "extend"
									case expectConfigExpectEquality:
										bitProp = "addEqualityTesters"
									case expectConfigReceiver:
										bitProp = "expect"
									}
									hasAcc := symHasAnyAccessor && bitProp != "" &&
										analysis.symbolMayHaveAccessorOnProp(symbol, bitProp, true, 0)
									if hasAcc {
										summary.escaped = true
										record(reference, stmt, block, bit, false)
									} else {
										record(reference, stmt, block, bit, true)
									}
								}
							}
						} else {
							// Passing `symbol` into an opaque helper call (`restore(holder)`,
							// `Object.defineProperty(holder, ...)`) or `Object.assign` on an
							// object with accessors (which can invoke a setter on any property)
							// may mutate/escape `symbol`.
							summary.escaped = true
							recordAllRestores(reference, stmt, block)
						}
					} else if !analysis.isNonEscapingSymbolUse(parent, current, symHasAnyAccessor) {
						// Any non-call escape (`box.h = holder`, `new Mutator(holder)`,
						// `[holder]`, `{ h: holder }`, `return holder`, `(a = holder).x`,
						// `holder == 1`, etc.) can mutate `symbol` indirectly later.
						summary.escaped = true
						recordAllRestores(reference, stmt, block)
					}
				}
				break
			}
		}
	}
	if summary.escaped {
		// Once `symbol` has an untracked escape (e.g. `box.h = holder`, `capture(holder)`,
		// `let leaked; set trap(_) { leaked = this; }`), indirect mutations via the
		// escaped reference can restore properties without mentioning `symbol`, so
		// MUST-overwrite `clears: true` events are no longer sound.
		for key, events := range summary.events {
			kept := events[:0]
			for _, ev := range events {
				if !ev.clears {
					kept = append(kept, ev)
				}
			}
			if len(kept) == 0 {
				delete(summary.events, key)
			} else {
				summary.events[key] = kept
			}
		}
	}
	for _, events := range summary.events {
		sort.Slice(events, func(i, j int) bool {
			if events[i].stmtEnd != events[j].stmtEnd {
				return events[i].stmtEnd < events[j].stmtEnd
			}
			// Put clears before restores within the same statement so `events[idx-1]`
			// always picks any same-statement restore over a same-statement clear.
			if events[i].clears != events[j].clears {
				return events[i].clears
			}
			return events[i].pos < events[j].pos
		})
	}
	if analysis.expectConfigOverwriteCycleHits == cycleBefore &&
		analysis.expectConfigCycleHits == expectCycleBefore &&
		analysis.objectLiteralMutationCycleHits == mutCycleBefore &&
		(constInitObj == nil || symSummary.evaluatedSources) {
		if analysis.expectConfigMustOverwrite == nil {
			analysis.expectConfigMustOverwrite = map[*ast.Symbol]*expectConfigOverwriteSummary{}
		}
		analysis.expectConfigMustOverwrite[symbol] = summary
	}
	return summary
}

func expectConfigOverwriteBit(propName string) expectConfigKind {
	switch propName {
	case "extend":
		return expectConfigExpectExtend
	case "addEqualityTesters":
		return expectConfigExpectEquality
	case "expect":
		return expectConfigReceiver
	default:
		return expectConfigNone
	}
}

func (analysis *RstestCallAnalysis) overwritePreservesConfigBit(bit expectConfigKind, rhs *ast.Node) bool {
	if rhs == nil {
		return false
	}
	k := analysis.expectConfigMember(rhs)
	switch bit {
	case expectConfigExpectExtend:
		return k&(expectConfigExtend|expectConfigDynamicMember) != 0
	case expectConfigExpectEquality:
		return k&(expectConfigAddEqualityTesters|expectConfigDynamicMember) != 0
	case expectConfigReceiver:
		return k&expectConfigExpect != 0
	default:
		return false
	}
}

func outermostExpressionWrapper(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		if ast.IsParenthesizedExpression(parent) {
			current = parent
			continue
		}
		if expr, ok := internalUtils.TransparentExpression(parent); ok && expr == current {
			current = parent
			continue
		}
		break
	}
	return current
}

func (analysis *RstestCallAnalysis) getterExpectConfigKind(fn *ast.Node) expectConfigKind {
	fn = internalUtils.SkipAssertionsAndParens(fn)
	if fn == nil {
		return expectConfigNone
	}
	body := fn.Body()
	if body == nil {
		return expectConfigNone
	}
	if body.Kind != ast.KindBlock {
		return analysis.expectConfigMember(body)
	}
	var out expectConfigKind
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node != body && testFramework.IsFunction(node) {
			return
		}
		if node.Kind == ast.KindReturnStatement {
			if ret := node.AsReturnStatement(); ret != nil && ret.Expression != nil {
				out |= analysis.expectConfigMember(ret.Expression)
			}
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(body)
	return out
}

func (analysis *RstestCallAnalysis) getterPreservesConfigBit(bit expectConfigKind, getterExpr *ast.Node) bool {
	getterExpr = internalUtils.SkipAssertionsAndParens(getterExpr)
	if getterExpr == nil {
		return true
	}
	switch getterExpr.Kind {
	case ast.KindGetAccessor, ast.KindMethodDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction:
		kind := analysis.getterExpectConfigKind(getterExpr)
		switch bit {
		case expectConfigExpectExtend:
			return kind&(expectConfigExtend|expectConfigDynamicMember) != 0
		case expectConfigExpectEquality:
			return kind&(expectConfigAddEqualityTesters|expectConfigDynamicMember) != 0
		case expectConfigReceiver:
			return kind&expectConfigExpect != 0
		default:
			return false
		}
	default:
		// Non-literal getter function may return the original API.
		return true
	}
}

// propertyDescriptorChange classifies an `Object.defineProperty` /
// `Reflect.defineProperty` / `Object.defineProperties` descriptor object.
// Attribute-only descriptors such as `{ enumerable: false }` keep the existing
// property value (ok = false). Descriptors specifying `value` or `get` return
// `(clears, true)` where `clears` is false when `value`/`get` may return the real API.
func (analysis *RstestCallAnalysis) propertyDescriptorChange(descNode *ast.Node, bit expectConfigKind) (clears bool, ok bool) {
	desc := internalUtils.SkipAssertionsAndParens(descNode)
	if desc == nil || desc.Kind != ast.KindObjectLiteralExpression || desc.AsObjectLiteralExpression().Properties == nil {
		return false, false
	}
	hasValue := false
	valuePreserves := false
	hasGetter := false
	getterPreserves := false
	for _, prop := range desc.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil || prop.Name() == nil {
			continue
		}
		name, static := internalUtils.GetStaticPropertyName(prop.Name())
		if !static {
			name, static = analysis.staticStringEvaluator().EvalPropertyName(prop.Name())
		}
		if !static {
			continue
		}
		switch name {
		case "value":
			hasValue = true
			switch prop.Kind {
			case ast.KindPropertyAssignment:
				valuePreserves = analysis.overwritePreservesConfigBit(bit, prop.AsPropertyAssignment().Initializer)
			case ast.KindShorthandPropertyAssignment:
				valuePreserves = analysis.overwritePreservesConfigBit(bit, prop.AsShorthandPropertyAssignment().Name())
			default:
				valuePreserves = false
			}
		case "get":
			hasGetter = true
			switch prop.Kind {
			case ast.KindPropertyAssignment:
				getterPreserves = analysis.getterPreservesConfigBit(bit, prop.AsPropertyAssignment().Initializer)
			case ast.KindMethodDeclaration:
				getterPreserves = analysis.getterPreservesConfigBit(bit, prop)
			default:
				getterPreserves = true
			}
		}
	}
	if hasValue {
		return !valuePreserves, true
	}
	if hasGetter {
		return !getterPreserves, true
	}
	return false, false
}

// mutatingBuiltinOverwriteChanges returns the net map of `bit -> clears`
// (`true` = overwritten, `false` = restored) produced by a standard
// property-mutating builtin call targeting `argNode`.
func (analysis *RstestCallAnalysis) mutatingBuiltinOverwriteChanges(
	call *ast.Node,
	argNode *ast.Node,
) map[expectConfigKind]bool {
	args := call.AsCallExpression().Arguments
	if args == nil || len(args.Nodes) < 2 || args.Nodes[0] != argNode {
		return nil
	}
	callee := internalUtils.SkipAssertionsAndParens(call.AsCallExpression().Expression)
	if callee == nil || !ast.IsAccessExpression(callee) {
		return nil
	}
	receiver := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee))
	if receiver == nil || receiver.Kind != ast.KindIdentifier {
		return nil
	}
	receiverName := receiver.AsIdentifier().Text
	if receiverName != "Object" && receiverName != "Reflect" {
		return nil
	}
	if analysis.ctx.Refs != nil {
		if !analysis.ctx.Refs.IsGlobalReference(receiver) {
			return nil
		}
	} else if internalUtils.IsShadowed(receiver, receiverName) {
		return nil
	}
	method, staticMethod := analysis.evalExpectConfigAccessName(callee)
	if !staticMethod {
		return nil
	}
	changes := map[expectConfigKind]bool{}
	switch receiverName + "." + method {
	case "Object.assign":
		for _, srcArg := range args.Nodes[1:] {
			src := internalUtils.SkipAssertionsAndParens(srcArg)
			if src == nil {
				continue
			}
			if src.Kind == ast.KindSpreadElement {
				// A spread source list may include `expect` or objects that
				// restore any previously overwritten property.
				changes[expectConfigExpectExtend] = false
				changes[expectConfigExpectEquality] = false
				changes[expectConfigReceiver] = false
				continue
			}
			if src.Kind != ast.KindObjectLiteralExpression {
				// A non-literal source such as `Object.assign(mock, { extend() {} }, expect)`
				// restores the bits carried by that source.
				srcKind := analysis.expectConfigMember(src)
				if srcKind&expectConfigExpectExtend != 0 {
					changes[expectConfigExpectExtend] = false
				}
				if srcKind&expectConfigExpectEquality != 0 {
					changes[expectConfigExpectEquality] = false
				}
				if srcKind&expectConfigReceiver != 0 {
					changes[expectConfigReceiver] = false
				}
				continue
			}
			if src.AsObjectLiteralExpression().Properties == nil {
				continue
			}
			for _, prop := range src.AsObjectLiteralExpression().Properties.Nodes {
				if prop == nil {
					continue
				}
				if prop.Kind == ast.KindSpreadAssignment {
					spreadKind := analysis.expectConfigMember(prop.AsSpreadAssignment().Expression)
					if spreadKind&expectConfigExpectExtend != 0 {
						changes[expectConfigExpectExtend] = false
					}
					if spreadKind&expectConfigExpectEquality != 0 {
						changes[expectConfigExpectEquality] = false
					}
					if spreadKind&expectConfigReceiver != 0 {
						changes[expectConfigReceiver] = false
					}
					continue
				}
				if prop.Name() == nil {
					continue
				}
				name, ok, staticKey := analysis.evalPropertyNameKey(prop.Name())
				if !ok {
					if !staticKey {
						changes[expectConfigExpectExtend] = false
						changes[expectConfigExpectEquality] = false
						changes[expectConfigReceiver] = false
					}
					continue
				}
				bit := expectConfigOverwriteBit(name)
				if bit == expectConfigNone {
					continue
				}
				preserves := false
				switch prop.Kind {
				case ast.KindPropertyAssignment:
					preserves = analysis.overwritePreservesConfigBit(bit, prop.AsPropertyAssignment().Initializer)
				case ast.KindShorthandPropertyAssignment:
					preserves = analysis.overwritePreservesConfigBit(bit, prop.AsShorthandPropertyAssignment().Name())
				case ast.KindGetAccessor:
					preserves = analysis.getterPreservesConfigBit(bit, prop)
				}
				changes[bit] = !preserves
			}
		}
		return changes
	case "Object.defineProperties":
		src := internalUtils.SkipAssertionsAndParens(args.Nodes[1])
		if src == nil || src.Kind != ast.KindObjectLiteralExpression || src.AsObjectLiteralExpression().Properties == nil {
			return nil
		}
		for _, prop := range src.AsObjectLiteralExpression().Properties.Nodes {
			if prop == nil || prop.Name() == nil || prop.Kind != ast.KindPropertyAssignment {
				continue
			}
			name, ok := internalUtils.GetStaticPropertyName(prop.Name())
			if !ok {
				name, ok = analysis.staticStringEvaluator().EvalPropertyName(prop.Name())
			}
			if !ok {
				continue
			}
			bit := expectConfigOverwriteBit(name)
			if bit == expectConfigNone {
				continue
			}
			if clears, ok := analysis.propertyDescriptorChange(prop.AsPropertyAssignment().Initializer, bit); ok {
				changes[bit] = clears
			}
		}
		return changes
	case "Object.defineProperty", "Reflect.defineProperty":
		if len(args.Nodes) < 3 {
			return nil
		}
		if name, known := analysis.staticStringEvaluator().EvalToString(args.Nodes[1]); known {
			if bit := expectConfigOverwriteBit(name); bit != expectConfigNone {
				if clears, ok := analysis.propertyDescriptorChange(args.Nodes[2], bit); ok {
					changes[bit] = clears
				}
			}
		}
		return changes
	case "Reflect.set":
		if name, known := analysis.staticStringEvaluator().EvalToString(args.Nodes[1]); known {
			if bit := expectConfigOverwriteBit(name); bit != expectConfigNone {
				preserves := len(args.Nodes) >= 3 && analysis.overwritePreservesConfigBit(bit, args.Nodes[2])
				changes[bit] = !preserves
			}
		}
		return changes
	case "Reflect.deleteProperty":
		if name, known := analysis.staticStringEvaluator().EvalToString(args.Nodes[1]); known {
			if bit := expectConfigOverwriteBit(name); bit != expectConfigNone {
				changes[bit] = true
			}
		}
		return changes
	default:
		return nil
	}
}

func (analysis *RstestCallAnalysis) writeCanReachRead(
	symbol *ast.Symbol,
	write *expectConfigWrite,
	readNode *ast.Node,
	readStmt *ast.Node,
	readBlock *ast.Node,
	readFn *ast.Node,
	summary *expectConfigSymbolSummary,
) bool {
	// If the write lives inside a function that does not enclose readNode
	// (including a hoisted function declared below readNode), check whether
	// that function could have been invoked before readNode.
	if write.enclosingFn != nil && !nodeContains(write.enclosingFn, readNode) {
		return analysis.functionCanExecuteBefore(write.enclosingFn, readNode, symbol, nil)
	}
	if write.pos > readNode.Pos() && directRangeCannotExecuteBefore(write.pos, write.exprEnd, readNode, symbol) {
		return false
	}
	// An intervening unconditional write in readBlock kills any write in the
	// same function whose full expression completed before the kill.
	if readStmt == nil || readBlock == nil || write.enclosingFn != readFn {
		return true
	}
	kills := summary.blockKills[readBlock]
	if len(kills) == 0 {
		return true
	}
	// Binary search for the first unconditional kill strictly after write's
	// full expression (so `fn = (fn = () => {}, expect.extend)` does not let
	// the inner RHS write kill the outer assignment).
	idx := sort.Search(len(kills), func(i int) bool {
		return kills[i].pos >= write.exprEnd
	})
	if idx < len(kills) && kills[idx].stmtEnd <= readStmt.Pos() {
		return false
	}
	return true
}

func nodeContains(ancestor, descendant *ast.Node) bool {
	if ancestor == nil || descendant == nil {
		return false
	}
	return descendant.Pos() >= ancestor.Pos() && descendant.End() <= ancestor.End()
}

func symbolDeclaredInside(symbol *ast.Symbol, fn *ast.Node) bool {
	if symbol == nil || fn == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, decl := range symbol.Declarations {
		if decl == nil || !nodeContains(fn, decl) {
			return false
		}
	}
	return true
}

// directRangeCannotExecuteBefore reports whether a non-function-wrapped node
// spanning [minPos, maxEnd] cannot run before readNode because it occurs
// strictly after readNode and readNode is not inside a loop, non-static class
// field initializer, or function that could be re-invoked while observing an
// outer variable `symbol`.
func directRangeCannotExecuteBefore(minPos, maxEnd int, readNode *ast.Node, symbol *ast.Symbol) bool {
	if readNode == nil || minPos <= readNode.Pos() {
		return false
	}
	prev := readNode
	for current := readNode.Parent; current != nil; prev, current = current, current.Parent {
		if ast.IsIterationStatement(current, false) {
			return false
		}
		if current.Kind == ast.KindPropertyDeclaration {
			if prop := current.AsPropertyDeclaration(); prop != nil && prop.Initializer == prev &&
				!ast.HasSyntacticModifier(current, ast.ModifierFlagsStatic) {
				return false
			}
		}
		if current.Kind == ast.KindBindingElement {
			if be := current.AsBindingElement(); be != nil && be.Initializer == prev {
				// Destructuring RHS executes before the default initializer.
				return false
			}
		}
		if current.Kind == ast.KindShorthandPropertyAssignment {
			if sp := current.AsShorthandPropertyAssignment(); sp != nil && sp.ObjectAssignmentInitializer == prev {
				return false
			}
		}
		if current.Kind == ast.KindBinaryExpression {
			if bin := current.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil &&
				bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Right == prev &&
				internalUtils.IsDefaultValueInDestructuringAssignment(current) {
				return false
			}
		}
		if testFramework.IsFunction(current) {
			// A write occurring later inside a function can still execute
			// before readNode on a subsequent invocation of that function
			// whenever `symbol` lives outside `current` (captured variable).
			if minPos < current.Pos() || maxEnd > current.End() || !symbolDeclaredInside(symbol, current) {
				return false
			}
		}
	}
	return true
}

func (analysis *RstestCallAnalysis) functionCanExecuteBefore(
	fn *ast.Node,
	readNode *ast.Node,
	symbol *ast.Symbol,
	visited map[*ast.Symbol]bool,
) bool {
	if fn == nil || readNode == nil || analysis.ctx.Refs == nil {
		return true
	}
	fnSym := analysis.declarationSymbolOfFunction(fn)
	if fnSym == nil {
		return true
	}
	if visited[fnSym] {
		return false
	}
	summary := analysis.getExpectConfigFnSummary(fn, fnSym)
	if summary.alwaysExecutable {
		return true
	}
	if summary.hasDirect {
		if summary.minDirectPos < readNode.Pos() ||
			!directRangeCannotExecuteBefore(summary.minDirectPos, summary.maxDirectEnd, readNode, symbol) {
			return true
		}
	}
	if len(summary.callerFns) == 0 {
		return false
	}
	if visited == nil {
		visited = map[*ast.Symbol]bool{}
	}
	visited[fnSym] = true
	readFn := ast.GetContainingFunction(readNode)
	for _, callerFn := range summary.callerFns {
		if callerFn == readFn || nodeContains(callerFn, readNode) {
			bounds, ok := summary.callerBounds[callerFn]
			if !ok || bounds.minPos < readNode.Pos() ||
				!directRangeCannotExecuteBefore(bounds.minPos, bounds.maxEnd, readNode, symbol) {
				return true
			}
			continue
		}
		if analysis.functionCanExecuteBefore(callerFn, readNode, symbol, visited) {
			return true
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) declarationSymbolOfFunction(fn *ast.Node) *ast.Symbol {
	var declNode *ast.Node
	var nameNode *ast.Node
	switch fn.Kind {
	case ast.KindFunctionDeclaration:
		declNode = fn
		nameNode = fn.Name()
	case ast.KindFunctionExpression, ast.KindArrowFunction:
		outer := internalUtils.OutermostParenthesizedExpression(fn)
		if outer != nil && outer.Parent != nil && outer.Parent.Kind == ast.KindVariableDeclaration &&
			outer.Parent.AsVariableDeclaration().Initializer == outer {
			declNode = outer.Parent
			nameNode = declNode.Name()
		}
	}
	if declNode == nil || nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return nil
	}
	if sym := declNode.Symbol(); sym != nil {
		return sym
	}
	if analysis.ctx.TypeChecker != nil {
		return analysis.ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	}
	return nil
}

func (analysis *RstestCallAnalysis) getExpectConfigFnSummary(fn *ast.Node, fnSym *ast.Symbol) *expectConfigFnSummary {
	if summary, ok := analysis.expectConfigFnReach[fnSym]; ok {
		return summary
	}
	summary := &expectConfigFnSummary{}
	if fn.Kind == ast.KindFunctionDeclaration && ast.HasSyntacticModifier(fn, ast.ModifierFlagsExport) {
		summary.alwaysExecutable = true
	} else if outer := internalUtils.OutermostParenthesizedExpression(fn); outer != nil &&
		outer.Parent != nil && outer.Parent.Kind == ast.KindVariableDeclaration {
		decl := outer.Parent
		if decl.Parent != nil && decl.Parent.Parent != nil &&
			ast.HasSyntacticModifier(decl.Parent.Parent, ast.ModifierFlagsExport) {
			summary.alwaysExecutable = true
		}
	}
	if !summary.alwaysExecutable && analysis.ctx.Refs != nil {
		for _, ref := range analysis.ctx.Refs.References(fnSym) {
			if ref == nil {
				continue
			}
			if ast.IsDeclarationName(ref) {
				continue
			}
			if ref.Parent != nil && (ref.Parent.Kind == ast.KindExportSpecifier || ref.Parent.Kind == ast.KindExportAssignment) {
				summary.alwaysExecutable = true
				break
			}
			callerFn := ast.GetContainingFunction(ref)
			if callerFn == fn {
				continue
			}
			if callerFn != nil {
				if summary.callerBounds == nil {
					summary.callerBounds = map[*ast.Node]expectConfigPosRange{}
				}
				if bounds, exists := summary.callerBounds[callerFn]; !exists {
					summary.callerFns = append(summary.callerFns, callerFn)
					summary.callerBounds[callerFn] = expectConfigPosRange{minPos: ref.Pos(), maxEnd: ref.End()}
				} else {
					if ref.Pos() < bounds.minPos {
						bounds.minPos = ref.Pos()
					}
					if ref.End() > bounds.maxEnd {
						bounds.maxEnd = ref.End()
					}
					summary.callerBounds[callerFn] = bounds
				}
				continue
			}
			if !summary.hasDirect {
				summary.hasDirect = true
				summary.minDirectPos = ref.Pos()
				summary.maxDirectEnd = ref.End()
			} else {
				if ref.Pos() < summary.minDirectPos {
					summary.minDirectPos = ref.Pos()
				}
				if ref.End() > summary.maxDirectEnd {
					summary.maxDirectEnd = ref.End()
				}
			}
		}
	}
	if analysis.expectConfigFnReach == nil {
		analysis.expectConfigFnReach = map[*ast.Symbol]*expectConfigFnSummary{}
	}
	analysis.expectConfigFnReach[fnSym] = summary
	return summary
}

func enclosingStatementAndList(node *ast.Node) (*ast.Node, *ast.Node) {
	var stmt *ast.Node
	for current := node; current != nil; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			break
		}
		switch parent.Kind {
		case ast.KindBlock, ast.KindSourceFile, ast.KindModuleBlock, ast.KindCaseClause, ast.KindDefaultClause:
			return current, parent
		}
		stmt = current
	}
	return stmt, nil
}

func writeStatementInfo(node *ast.Node) (stmt *ast.Node, block *ast.Node, unconditional bool) {
	stmt, block = enclosingStatementAndList(node)
	if stmt == nil || block == nil {
		return stmt, block, false
	}
	prev := node
	for current := node.Parent; current != nil && current != block; prev, current = current, current.Parent {
		switch current.Kind {
		case ast.KindConditionalExpression, ast.KindIfStatement, ast.KindSwitchStatement,
			ast.KindTryStatement, ast.KindCatchClause, ast.KindClassStaticBlockDeclaration:
			return stmt, block, false
		case ast.KindPropertyDeclaration:
			if prop := current.AsPropertyDeclaration(); prop != nil && prop.Initializer == prev {
				return stmt, block, false
			}
		case ast.KindBindingElement:
			if be := current.AsBindingElement(); be != nil && be.Initializer == prev {
				return stmt, block, false
			}
		case ast.KindShorthandPropertyAssignment:
			if sp := current.AsShorthandPropertyAssignment(); sp != nil && sp.ObjectAssignmentInitializer == prev {
				return stmt, block, false
			}
		case ast.KindBinaryExpression:
			bin := current.AsBinaryExpression()
			if bin != nil && bin.OperatorToken != nil {
				switch bin.OperatorToken.Kind {
				case ast.KindAmpersandAmpersandToken, ast.KindBarBarToken, ast.KindQuestionQuestionToken,
					ast.KindAmpersandAmpersandEqualsToken, ast.KindBarBarEqualsToken, ast.KindQuestionQuestionEqualsToken:
					return stmt, block, false
				case ast.KindEqualsToken:
					if bin.Right == prev && internalUtils.IsDefaultValueInDestructuringAssignment(current) {
						return stmt, block, false
					}
				}
			}
		case ast.KindCallExpression:
			if ce := current.AsCallExpression(); ce != nil && ast.IsOptionalChain(current) && ce.Expression != prev {
				return stmt, block, false
			}
		case ast.KindPropertyAccessExpression:
			if pa := current.AsPropertyAccessExpression(); pa != nil && ast.IsOptionalChain(current) && pa.Expression != prev {
				return stmt, block, false
			}
		case ast.KindElementAccessExpression:
			if ea := current.AsElementAccessExpression(); ea != nil && ast.IsOptionalChain(current) && ea.Expression != prev {
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

func (analysis *RstestCallAnalysis) destructuredExpectConfigMember(declaration *ast.Node) expectConfigKind {
	binding := declaration.AsBindingElement()
	if binding == nil || binding.Name() == nil {
		return expectConfigNone
	}
	var out expectConfigKind
	pattern := declaration.Parent
	if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		if binding.Initializer != nil {
			out |= analysis.expectConfigMember(binding.Initializer)
		}
		return out
	}
	if binding.DotDotDotToken != nil {
		if binding.Initializer != nil {
			out |= analysis.expectConfigMember(binding.Initializer)
		}
		container := analysis.expectConfigBindingPatternKind(pattern)
		return out | (container & (expectConfigReceiver | expectConfigExpect))
	}
	member, known, staticKey := analysis.evalBindingElementPropertyName(binding)
	defaultSkipped := false
	sources, definiteSrc, sourcesMayBeMutable, sourcesUnresolvedGetter := analysis.bindingPatternSourceExprs(pattern)
	if sourcesUnresolvedGetter {
		out |= expectConfigDynamicMember
	}
	if known {
		if definiteSrc != nil {
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(definiteSrc); objExpr != nil && !mutableObj {
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(pattern, declaration, objExpr, false)
				defLookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter, 0)
				defaultSkipped = defLookup.definitelyDefined
			}
		}
		for _, srcExpr := range sources {
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(srcExpr); objExpr != nil {
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(pattern, declaration, objExpr, mutableObj)
				if mutableObj && !sourcesMayBeMutable && !priorGetter && len(sources) == 1 {
					if unwrapped := internalUtils.SkipAssertionsAndParens(srcExpr); unwrapped != nil && unwrapped.Kind == ast.KindIdentifier {
						if sym := resolveRstestRootSymbol(analysis.ctx, unwrapped); sym != nil {
							if replacement, replaced := analysis.unconditionalPropertyReplacementBefore(sym, objExpr, member, unwrapped); replaced {
								out |= analysis.expectConfigMember(replacement)
								defaultSkipped = analysis.isDefinitelyDefinedExpr(replacement, 0, nil)
								continue
							}
						}
					}
				}
				lookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter || mutableObj || sourcesMayBeMutable, 0)
				for _, expr := range lookup.possibleExprs {
					out |= analysis.expectConfigMember(expr)
				}
				useAllGetters := sourcesMayBeMutable || (mutableObj && analysis.srcExprObjEscaped(srcExpr, objExpr))
				if priorGetter || useAllGetters || mutableObj {
					gExprs, gContainers, gUnresolved := analysis.collectPrecedingGetterPropWriteExprs(pattern, declaration, objExpr, member, useAllGetters, mutableObj)
					for _, expr := range gExprs {
						out |= analysis.expectConfigMember(expr)
					}
					for _, cExpr := range gContainers {
						out |= classifyDestructuredExpectConfigProperty(analysis.expectConfigMember(cExpr), member, true, true)
					}
					if gUnresolved {
						out |= expectConfigDynamicMember
					}
				}
				for _, pExpr := range lookup.exposedProtoExprs {
					out |= classifyDestructuredExpectConfigProperty(analysis.expectConfigMember(pExpr), member, true, true)
				}
			}
		}
	} else if !staticKey {
		for _, srcExpr := range sources {
			for _, expr := range analysis.allObjectLiteralPropertyExprs(srcExpr) {
				out |= analysis.expectConfigMember(expr)
			}
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(srcExpr); objExpr != nil {
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(pattern, declaration, objExpr, mutableObj)
				useAllGetters := sourcesMayBeMutable || (mutableObj && analysis.srcExprObjEscaped(srcExpr, objExpr))
				if priorGetter || useAllGetters || mutableObj {
					gExprs, gContainers, gUnresolved := analysis.collectPrecedingGetterPropWriteExprs(pattern, declaration, objExpr, "", useAllGetters, mutableObj)
					for _, expr := range gExprs {
						out |= analysis.expectConfigMember(expr)
					}
					for _, cExpr := range gContainers {
						out |= classifyDestructuredExpectConfigProperty(analysis.expectConfigMember(cExpr), "", false, false)
					}
					if gUnresolved {
						out |= expectConfigDynamicMember
					}
				}
			}
		}
	}
	if binding.Initializer != nil && !defaultSkipped {
		out |= analysis.expectConfigMember(binding.Initializer)
	}
	container := analysis.expectConfigBindingPatternKind(pattern)
	if container == expectConfigNone {
		return out
	}
	return out | classifyDestructuredExpectConfigProperty(container, member, known, staticKey)
}

func (analysis *RstestCallAnalysis) bindingPatternSourceExprs(pattern *ast.Node) (sources []*ast.Node, definiteSource *ast.Node, sourcesMayBeMutable bool, unresolvedGetter bool) {
	if pattern == nil || pattern.Parent == nil {
		return nil, nil, false, false
	}
	switch pattern.Parent.Kind {
	case ast.KindVariableDeclaration:
		init := pattern.Parent.AsVariableDeclaration().Initializer
		if init != nil {
			return []*ast.Node{init}, init, false, false
		}
	case ast.KindBindingElement:
		be := pattern.Parent.AsBindingElement()
		if be == nil || be.DotDotDotToken != nil {
			return nil, nil, false, false
		}
		parentPattern := pattern.Parent.Parent
		if parentPattern == nil {
			return nil, nil, false, false
		}
		if parentPattern.Kind != ast.KindObjectBindingPattern {
			if be.Initializer != nil {
				return []*ast.Node{be.Initializer}, nil, false, false
			}
			return nil, nil, false, false
		}
		member, known, _ := analysis.evalBindingElementPropertyName(be)
		if !known {
			if be.Initializer != nil {
				return []*ast.Node{be.Initializer}, nil, false, false
			}
			return nil, nil, false, false
		}
		parentSources, parentDefinite, parentMutable, parentUnresolved := analysis.bindingPatternSourceExprs(parentPattern)
		sourcesMayBeMutable = parentMutable
		unresolvedGetter = parentUnresolved
		parentDefDefined := false
		// Only step `definiteSource` across a direct fresh `ObjectLiteralExpression`
		// (not a `const` alias `cfg` whose nested sub-object `cfg.nested` could
		// have been mutated before destructuring).
		if parentDefinite != nil {
			if directObj := internalUtils.SkipAssertionsAndParens(parentDefinite); directObj != nil &&
				directObj.Kind == ast.KindObjectLiteralExpression {
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(parentPattern, pattern.Parent, directObj, false)
				defLookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(directObj, member, priorGetter, 0)
				if defLookup.definitelyDefined && defLookup.exactExpr != nil &&
					!defLookup.fromNonFreshSource && !defLookup.mayHaveUnknownSource {
					parentDefDefined = true
					definiteSource = defLookup.exactExpr
				}
			}
		}
		for _, pSrc := range parentSources {
			unwrapped := internalUtils.SkipAssertionsAndParens(pSrc)
			if unwrapped == nil || unwrapped.Kind != ast.KindObjectLiteralExpression {
				sourcesMayBeMutable = true
			}
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(pSrc); objExpr != nil {
				if mutableObj {
					sourcesMayBeMutable = true
				}
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(parentPattern, pattern.Parent, objExpr, mutableObj)
				lookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter || mutableObj || sourcesMayBeMutable, 0)
				if lookup.mayHaveUnknownSource || lookup.fromNonFreshSource {
					sourcesMayBeMutable = true
				}
				sources = append(sources, lookup.possibleExprs...)
				useAllGetters := parentMutable || (mutableObj && analysis.srcExprObjEscaped(pSrc, objExpr))
				if priorGetter || useAllGetters || mutableObj {
					gExprs, _, gUnresolved := analysis.collectPrecedingGetterPropWriteExprs(parentPattern, pattern.Parent, objExpr, member, useAllGetters, mutableObj)
					sources = append(sources, gExprs...)
					if gUnresolved {
						unresolvedGetter = true
					}
				}
			}
		}
		if be.Initializer != nil && !parentDefDefined {
			sources = append(sources, be.Initializer)
		}
		return sources, definiteSource, sourcesMayBeMutable, unresolvedGetter
	}
	return nil, nil, false, false
}

func (analysis *RstestCallAnalysis) assignmentObjectSourceExprs(obj *ast.Node) (sources []*ast.Node, definiteSource *ast.Node, sourcesMayBeMutable bool, unresolvedGetter bool) {
	if obj == nil || obj.Parent == nil {
		return nil, nil, false, false
	}
	outer := internalUtils.OutermostParenthesizedExpression(obj)
	if outer == nil || outer.Parent == nil {
		return nil, nil, false, false
	}
	switch outer.Parent.Kind {
	case ast.KindBinaryExpression:
		bin := outer.Parent.AsBinaryExpression()
		if bin == nil || bin.OperatorToken == nil || bin.OperatorToken.Kind != ast.KindEqualsToken || bin.Left != outer {
			return nil, nil, false, false
		}
		if !internalUtils.IsDefaultValueInDestructuringAssignment(outer.Parent) {
			if bin.Right != nil {
				return []*ast.Node{bin.Right}, bin.Right, false, false
			}
			return nil, nil, false, false
		}
		// Nested destructuring assignment default, e.g.
		// `({ a: { register = r.expect.extend } = fallback } = { a: {} })` or
		// `([{ register } = fallback] = [])`: collect any parent property
		// values AND include `fallback` when the parent property is not
		// guaranteed defined (never mistaking `fallback` for `definiteSource`).
		if outer.Parent.Parent != nil && outer.Parent.Parent.Kind == ast.KindPropertyAssignment &&
			outer.Parent.Parent.AsPropertyAssignment().Initializer == outer.Parent {
			pSources, pDef, pMut, pUnresolved := analysis.assignmentPropertySourceExprs(outer.Parent.Parent)
			sources = append(sources, pSources...)
			if pDef != nil {
				return sources, pDef, pMut, pUnresolved
			}
			if bin.Right != nil {
				sources = append(sources, bin.Right)
			}
			return sources, nil, pMut, pUnresolved
		}
		if bin.Right != nil {
			return []*ast.Node{bin.Right}, nil, false, false
		}
	case ast.KindPropertyAssignment:
		return analysis.assignmentPropertySourceExprs(outer.Parent)
	}
	return nil, nil, false, false
}

func (analysis *RstestCallAnalysis) assignmentPropertySourceExprs(prop *ast.Node) (sources []*ast.Node, definiteSource *ast.Node, sourcesMayBeMutable bool, unresolvedGetter bool) {
	if prop == nil || prop.Parent == nil || prop.Parent.Kind != ast.KindObjectLiteralExpression || prop.Name() == nil {
		return nil, nil, false, false
	}
	member, known, _ := analysis.evalPropertyNameKey(prop.Name())
	if !known {
		return nil, nil, false, false
	}
	parentSources, parentDefinite, parentMutable, parentUnresolved := analysis.assignmentObjectSourceExprs(prop.Parent)
	sourcesMayBeMutable = parentMutable
	unresolvedGetter = parentUnresolved
	if parentDefinite != nil {
		if directObj := internalUtils.SkipAssertionsAndParens(parentDefinite); directObj != nil &&
			directObj.Kind == ast.KindObjectLiteralExpression {
			priorGetter := analysis.precedingDestructuringMayInvokeGetter(prop.Parent, prop, directObj, false)
			defLookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(directObj, member, priorGetter, 0)
			if defLookup.definitelyDefined && defLookup.exactExpr != nil &&
				!defLookup.fromNonFreshSource && !defLookup.mayHaveUnknownSource {
				definiteSource = defLookup.exactExpr
			}
		}
	}
	for _, pSrc := range parentSources {
		unwrapped := internalUtils.SkipAssertionsAndParens(pSrc)
		if unwrapped == nil || unwrapped.Kind != ast.KindObjectLiteralExpression {
			sourcesMayBeMutable = true
		}
		if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(pSrc); objExpr != nil {
			if mutableObj {
				sourcesMayBeMutable = true
			}
			priorGetter := analysis.precedingDestructuringMayInvokeGetter(prop.Parent, prop, objExpr, mutableObj)
			lookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter || mutableObj || sourcesMayBeMutable, 0)
			if lookup.mayHaveUnknownSource || lookup.fromNonFreshSource {
				sourcesMayBeMutable = true
			}
			sources = append(sources, lookup.possibleExprs...)
			useAllGetters := parentMutable || (mutableObj && analysis.srcExprObjEscaped(pSrc, objExpr))
			if priorGetter || useAllGetters || mutableObj {
				gExprs, _, gUnresolved := analysis.collectPrecedingGetterPropWriteExprs(prop.Parent, prop, objExpr, member, useAllGetters, mutableObj)
				sources = append(sources, gExprs...)
				if gUnresolved {
					unresolvedGetter = true
				}
			}
		}
	}
	return sources, definiteSource, sourcesMayBeMutable, unresolvedGetter
}

type objectLiteralPropertyLookup struct {
	exactExpr            *ast.Node
	possibleExprs        []*ast.Node
	exposedProtoExprs    []*ast.Node
	definitelyDefined    bool
	foundExplicit        bool
	mayHaveUnknownSource bool
	fromNonFreshSource   bool
}

func collectGetterReturnExprs(getter *ast.Node) []*ast.Node {
	if getter == nil {
		return nil
	}
	body := getter.Body()
	if body == nil {
		return nil
	}
	if body.Kind != ast.KindBlock {
		return []*ast.Node{body}
	}
	var out []*ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if node != body && testFramework.IsFunction(node) {
			return
		}
		if node.Kind == ast.KindReturnStatement {
			if ret := node.AsReturnStatement(); ret != nil && ret.Expression != nil {
				out = append(out, ret.Expression)
			}
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(body)
	return out
}

func (analysis *RstestCallAnalysis) allObjectLiteralPropertyExprs(sourceExpr *ast.Node) []*ast.Node {
	return analysis.allObjectLiteralPropertyExprsDepth(sourceExpr, 0)
}

func (analysis *RstestCallAnalysis) allObjectLiteralPropertyExprsDepth(sourceExpr *ast.Node, depth int) []*ast.Node {
	if depth > 4 {
		return nil
	}
	obj, _ := analysis.resolveObjectLiteralForBoundLookup(sourceExpr)
	if obj == nil || obj.AsObjectLiteralExpression().Properties == nil {
		return nil
	}
	var out []*ast.Node
	for _, prop := range obj.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		switch prop.Kind {
		case ast.KindSpreadAssignment:
			out = append(out, analysis.allObjectLiteralPropertyExprsDepth(prop.AsSpreadAssignment().Expression, depth+1)...)
		case ast.KindPropertyAssignment:
			if init := prop.AsPropertyAssignment().Initializer; init != nil {
				if pName := prop.Name(); pName != nil && pName.Kind != ast.KindComputedPropertyName {
					if name, ok := internalUtils.GetStaticPropertyName(pName); ok && name == "__proto__" {
						out = append(out, analysis.allObjectLiteralPropertyExprsDepth(init, depth+1)...)
						continue
					}
				}
				out = append(out, init)
			}
		case ast.KindShorthandPropertyAssignment:
			out = append(out, prop.AsShorthandPropertyAssignment().Name())
		case ast.KindGetAccessor:
			out = append(out, collectGetterReturnExprs(prop)...)
		}
	}
	return out
}

type destructuringPropKey struct {
	name  string
	known bool
}

func (analysis *RstestCallAnalysis) collectDestructuringKeys(pattern, stopBefore *ast.Node) (keys []destructuringPropKey, foundStop bool) {
	if pattern == nil {
		return nil, false
	}
	switch pattern.Kind {
	case ast.KindObjectBindingPattern:
		bp := pattern.AsBindingPattern()
		if bp == nil || bp.Elements == nil {
			return nil, false
		}
		for _, el := range bp.Elements.Nodes {
			if el == stopBefore {
				return keys, true
			}
			if el == nil || el.Kind != ast.KindBindingElement {
				continue
			}
			be := el.AsBindingElement()
			if be.DotDotDotToken != nil {
				keys = append(keys, destructuringPropKey{known: false})
				continue
			}
			name, known, _ := analysis.evalBindingElementPropertyName(be)
			keys = append(keys, destructuringPropKey{name: name, known: known})
		}
	case ast.KindObjectLiteralExpression:
		obj := pattern.AsObjectLiteralExpression()
		if obj == nil || obj.Properties == nil {
			return nil, false
		}
		for _, prop := range obj.Properties.Nodes {
			if prop == stopBefore {
				return keys, true
			}
			if prop == nil {
				continue
			}
			if prop.Kind == ast.KindSpreadAssignment || prop.Name() == nil {
				keys = append(keys, destructuringPropKey{known: false})
				continue
			}
			name, known, _ := analysis.evalPropertyNameKey(prop.Name())
			keys = append(keys, destructuringPropKey{name: name, known: known})
		}
	}
	return keys, false
}

type objGetterSummary struct {
	hasAny                  bool
	hasOwnDynamic           bool
	hasProtoDynamic         bool
	hasOpaqueProto          bool
	ownShadowedProps        map[string]bool
	staticGetters           map[string][]*ast.Node
	allGetters              []*ast.Node
	protoDynamicGetters     []*ast.Node
	shadowedProtoGetters    map[string][]*ast.Node
	allShadowedProtoGetters []*ast.Node
}

func (analysis *RstestCallAnalysis) getObjectLiteralGetterSummary(objExpr *ast.Node) *objGetterSummary {
	if objExpr == nil || objExpr.Kind != ast.KindObjectLiteralExpression {
		return &objGetterSummary{}
	}
	if analysis.objectLiteralGetterSummaries != nil {
		if cached, ok := analysis.objectLiteralGetterSummaries[objExpr]; ok {
			return cached
		}
	}
	summary := &objGetterSummary{}
	visited := map[*ast.Node]bool{}
	shadowed := map[string]bool{}
	var collect func(cur *ast.Node, depth int)
	collect = func(cur *ast.Node, depth int) {
		if depth > 4 {
			summary.hasAny = true
			summary.hasProtoDynamic = true
			summary.hasOpaqueProto = true
			return
		}
		if cur == nil || cur.Kind != ast.KindObjectLiteralExpression || cur.AsObjectLiteralExpression().Properties == nil || visited[cur] {
			return
		}
		visited[cur] = true
		var protoExpr *ast.Node
		curStaticGetters := map[string]*ast.Node{}
		curStaticData := map[string]bool{}
		curStaticSetters := map[string]bool{}
		var curStaticGetterOrder []string
		var curDynamicGetters []*ast.Node
		var curSymbolGetters []*ast.Node
		for _, prop := range cur.AsObjectLiteralExpression().Properties.Nodes {
			if prop == nil {
				continue
			}
			if prop.Kind == ast.KindPropertyAssignment && prop.Name() != nil && prop.Name().Kind != ast.KindComputedPropertyName {
				if name, ok := internalUtils.GetStaticPropertyName(prop.Name()); ok && name == "__proto__" {
					protoExpr = prop.AsPropertyAssignment().Initializer
					continue
				}
			}
			switch prop.Kind {
			case ast.KindGetAccessor:
				if prop.Name() == nil {
					if depth == 0 {
						summary.hasOwnDynamic = true
					} else {
						summary.hasProtoDynamic = true
						summary.protoDynamicGetters = append(summary.protoDynamicGetters, prop)
					}
					curDynamicGetters = append(curDynamicGetters, prop)
					continue
				}
				name, ok, staticKey := analysis.evalPropertyNameKey(prop.Name())
				if !ok && !staticKey {
					if depth == 0 {
						summary.hasOwnDynamic = true
					} else {
						summary.hasProtoDynamic = true
						summary.protoDynamicGetters = append(summary.protoDynamicGetters, prop)
					}
					curDynamicGetters = append(curDynamicGetters, prop)
					continue
				}
				if !ok && staticKey {
					curSymbolGetters = append(curSymbolGetters, prop)
					continue
				}
				if ok {
					if shadowed[name] {
						if depth > 0 {
							if summary.shadowedProtoGetters == nil {
								summary.shadowedProtoGetters = map[string][]*ast.Node{}
							}
							summary.shadowedProtoGetters[name] = append(summary.shadowedProtoGetters[name], prop)
							summary.allShadowedProtoGetters = append(summary.allShadowedProtoGetters, prop)
						}
						continue
					}
					delete(curStaticData, name)
					if _, exists := curStaticGetters[name]; !exists {
						curStaticGetterOrder = append(curStaticGetterOrder, name)
					}
					curStaticGetters[name] = prop
				}
			case ast.KindSetAccessor:
				if prop.Name() != nil {
					if name, ok, _ := analysis.evalPropertyNameKey(prop.Name()); ok {
						delete(curStaticData, name)
						curStaticSetters[name] = true
					}
				}
			case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment, ast.KindMethodDeclaration:
				if prop.Name() != nil {
					if name, ok, _ := analysis.evalPropertyNameKey(prop.Name()); ok {
						delete(curStaticGetters, name)
						delete(curStaticSetters, name)
						curStaticData[name] = true
					}
				}
			}
		}
		if depth == 0 && (len(curStaticData) > 0 || len(curStaticGetters) > 0 || len(curStaticSetters) > 0) {
			summary.ownShadowedProps = make(map[string]bool, len(curStaticData)+len(curStaticGetters)+len(curStaticSetters))
			for name := range curStaticData {
				summary.ownShadowedProps[name] = true
			}
			for name := range curStaticGetters {
				summary.ownShadowedProps[name] = true
			}
			for name := range curStaticSetters {
				summary.ownShadowedProps[name] = true
			}
		}
		for _, prop := range curDynamicGetters {
			summary.hasAny = true
			summary.allGetters = append(summary.allGetters, prop)
		}
		for _, prop := range curSymbolGetters {
			summary.hasAny = true
			summary.allGetters = append(summary.allGetters, prop)
		}
		for _, name := range curStaticGetterOrder {
			if shadowed[name] {
				continue
			}
			prop, ok := curStaticGetters[name]
			if !ok {
				continue
			}
			summary.hasAny = true
			summary.allGetters = append(summary.allGetters, prop)
			if summary.staticGetters == nil {
				summary.staticGetters = map[string][]*ast.Node{}
			}
			summary.staticGetters[name] = append(summary.staticGetters[name], prop)
			shadowed[name] = true
		}
		for name := range curStaticData {
			shadowed[name] = true
		}
		for name := range curStaticSetters {
			shadowed[name] = true
		}
		if protoExpr != nil {
			unwrapped := internalUtils.SkipAssertionsAndParens(protoExpr)
			if unwrapped != nil && unwrapped.Kind != ast.KindNullKeyword {
				protoObj, mutableProto := analysis.resolveObjectLiteralForBoundLookup(protoExpr)
				if protoObj == nil {
					summary.hasAny = true
					summary.hasProtoDynamic = true
					summary.hasOpaqueProto = true
				} else if mutableProto {
					summary.hasAny = true
					summary.hasProtoDynamic = true
					if analysis.srcExprObjEscaped(protoExpr, protoObj) {
						summary.hasOpaqueProto = true
					}
				}
				if protoObj != nil {
					collect(protoObj, depth+1)
				}
			}
		}
	}
	collect(objExpr, 0)
	if analysis.objectLiteralGetterSummaries == nil {
		analysis.objectLiteralGetterSummaries = map[*ast.Node]*objGetterSummary{}
	}
	analysis.objectLiteralGetterSummaries[objExpr] = summary
	return summary
}

func (analysis *RstestCallAnalysis) srcExprObjEscaped(srcExpr, objExpr *ast.Node) bool {
	unwrapped := internalUtils.SkipAssertionsAndParens(srcExpr)
	if unwrapped == nil || unwrapped.Kind != ast.KindIdentifier {
		return false
	}
	sym := resolveRstestRootSymbol(analysis.ctx, unwrapped)
	if sym == nil {
		return false
	}
	return analysis.getObjectLiteralMutationSummary(sym, objExpr).escaped
}

func (analysis *RstestCallAnalysis) collectDescriptorValueAndGetExprs(descObj *ast.Node) (out []*ast.Node, complete bool) {
	if descObj == nil || descObj.Kind != ast.KindObjectLiteralExpression || descObj.AsObjectLiteralExpression().Properties == nil {
		return nil, false
	}
	complete = true
	for _, prop := range descObj.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		if prop.Kind == ast.KindSpreadAssignment || prop.Name() == nil {
			complete = false
			continue
		}
		if prop.Kind == ast.KindPropertyAssignment && prop.Name().Kind != ast.KindComputedPropertyName {
			if name, ok := internalUtils.GetStaticPropertyName(prop.Name()); ok && name == "__proto__" {
				init := internalUtils.SkipAssertionsAndParens(prop.AsPropertyAssignment().Initializer)
				if init == nil || init.Kind != ast.KindNullKeyword {
					complete = false
				}
				continue
			}
		}
		name, ok, _ := analysis.evalPropertyNameKey(prop.Name())
		if !ok {
			complete = false
			continue
		}
		switch name {
		case "value":
			switch prop.Kind {
			case ast.KindPropertyAssignment:
				out = append(out, prop.AsPropertyAssignment().Initializer)
			case ast.KindShorthandPropertyAssignment:
				out = append(out, prop.AsShorthandPropertyAssignment().Name())
			case ast.KindMethodDeclaration:
				out = append(out, prop)
			case ast.KindGetAccessor:
				out = append(out, collectGetterReturnExprs(prop)...)
			default:
				complete = false
			}
		case "get":
			switch prop.Kind {
			case ast.KindMethodDeclaration:
				out = append(out, collectGetterReturnExprs(prop)...)
			case ast.KindPropertyAssignment:
				init := internalUtils.SkipAssertionsAndParens(prop.AsPropertyAssignment().Initializer)
				if init != nil && (init.Kind == ast.KindArrowFunction || init.Kind == ast.KindFunctionExpression) {
					out = append(out, collectGetterReturnExprs(init)...)
				} else {
					complete = false
				}
			default:
				complete = false
			}
		}
	}
	return out, complete
}

func (gs *objGetterSummary) keyMayInvokeGetter(k destructuringPropKey, mutable bool) bool {
	if !k.known {
		return gs.hasAny || (mutable && len(gs.allShadowedProtoGetters) > 0)
	}
	if len(gs.staticGetters[k.name]) > 0 {
		return true
	}
	if mutable && len(gs.shadowedProtoGetters[k.name]) > 0 {
		return true
	}
	if gs.hasOwnDynamic {
		return true
	}
	return gs.hasProtoDynamic && (mutable || !gs.ownShadowedProps[k.name])
}

func (analysis *RstestCallAnalysis) destructuringPatternMayInvokeGetter(pattern, objExpr *ast.Node, mutable bool) bool {
	if pattern == nil || objExpr == nil {
		return false
	}
	gs := analysis.getObjectLiteralGetterSummary(objExpr)
	if !gs.hasAny && (!mutable || len(gs.allShadowedProtoGetters) == 0) {
		return false
	}
	keys, _ := analysis.collectDestructuringKeys(pattern, nil)
	for _, k := range keys {
		if gs.keyMayInvokeGetter(k, mutable) {
			return true
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) precedingDestructuringMayInvokeGetter(pattern, targetElem, objExpr *ast.Node, mutable bool) bool {
	if pattern == nil || targetElem == nil || objExpr == nil {
		return false
	}
	gs := analysis.getObjectLiteralGetterSummary(objExpr)
	if !gs.hasAny && (!mutable || len(gs.allShadowedProtoGetters) == 0) {
		return false
	}
	keys, found := analysis.collectDestructuringKeys(pattern, targetElem)
	if !found {
		return true
	}
	for _, k := range keys {
		if gs.keyMayInvokeGetter(k, mutable) {
			return true
		}
	}
	return false
}

func isImmediatelyInvokedFunction(node *ast.Node) bool {
	if node == nil || (node.Kind != ast.KindArrowFunction && node.Kind != ast.KindFunctionExpression) {
		return false
	}
	outer := outermostExpressionWrapper(node)
	if outer == nil || outer.Parent == nil {
		return false
	}
	if outer.Parent.Kind == ast.KindCallExpression {
		if outer.Parent.AsCallExpression().Expression == outer {
			return true
		}
		// Inline arrow passed directly as a callback argument (e.g. `[1].forEach(() => { this.register = ... })`).
		if node.Kind == ast.KindArrowFunction {
			return true
		}
	}
	if ast.IsAccessExpression(outer.Parent) && outer.Parent.Expression() == outer {
		accOuter := outermostExpressionWrapper(outer.Parent)
		if accOuter != nil && accOuter.Parent != nil && accOuter.Parent.Kind == ast.KindCallExpression &&
			accOuter.Parent.AsCallExpression().Expression == accOuter {
			if name, ok := internalUtils.AccessExpressionStaticName(outer.Parent); ok && (name == "call" || name == "apply") {
				return true
			}
		}
	}
	return false
}

func isSimplePrimitiveLiteralExpr(node *ast.Node) bool {
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral,
		ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	}
	return false
}

func thisAccessRootProp(node *ast.Node, isThis func(*ast.Node) bool) (rootAccess *ast.Node, isDirect bool, ok bool) {
	cur := internalUtils.SkipAssertionsAndParens(node)
	if cur == nil || !ast.IsAccessExpression(cur) {
		return nil, false, false
	}
	direct := true
	for cur != nil && ast.IsAccessExpression(cur) {
		recv := internalUtils.SkipAssertionsAndParens(cur.Expression())
		if isThis(recv) {
			return cur, direct, true
		}
		cur = recv
		direct = false
	}
	return nil, false, false
}

func (analysis *RstestCallAnalysis) lookupOwnLiteralInlineMethod(objExpr *ast.Node, methodName string) *ast.Node {
	if objExpr == nil || objExpr.Kind != ast.KindObjectLiteralExpression || objExpr.AsObjectLiteralExpression().Properties == nil {
		return nil
	}
	var found *ast.Node
	for _, prop := range objExpr.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		if prop.Kind == ast.KindSpreadAssignment {
			found = nil
			continue
		}
		if prop.Name() == nil {
			found = nil
			continue
		}
		name, ok, staticKey := analysis.evalPropertyNameKey(prop.Name())
		if !ok && !staticKey {
			found = nil
			continue
		}
		if !ok || name != methodName {
			continue
		}
		switch prop.Kind {
		case ast.KindMethodDeclaration:
			found = prop
		case ast.KindPropertyAssignment:
			init := internalUtils.SkipAssertionsAndParens(prop.AsPropertyAssignment().Initializer)
			if init != nil && init.Kind == ast.KindFunctionExpression {
				found = init
			} else {
				found = nil
			}
		default:
			found = nil
		}
	}
	return found
}

// collectPrecedingGetterPropWriteExprs collects RHS expressions (and non-literal
// `Object.assign(this, src)` containers) written to `this[targetProp]` (or any
// property when `targetProp == ""`) inside getters that can execute before
// `targetElem` during object destructuring. It sets `unresolved` when an
// executed getter mutates `this` / `this[targetProp]` in a way whose installed
// value cannot be statically extracted (e.g. nested writes `this.nested.x = ...`,
// destructuring writes, unresolved descriptors, or external method/helper calls).
func (analysis *RstestCallAnalysis) collectPrecedingGetterPropWriteExprs(
	pattern, targetElem, objExpr *ast.Node,
	targetProp string,
	allGetters bool,
	mutableObj bool,
) (exprs []*ast.Node, containers []*ast.Node, unresolved bool) {
	if pattern == nil || targetElem == nil || objExpr == nil {
		return nil, nil, false
	}
	gs := analysis.getObjectLiteralGetterSummary(objExpr)
	if len(gs.allGetters) == 0 && len(gs.allShadowedProtoGetters) == 0 && !gs.hasOpaqueProto {
		return nil, nil, false
	}
	var getters []*ast.Node
	seen := map[*ast.Node]bool{}
	appendGetters := func(list []*ast.Node) {
		for _, g := range list {
			if g != nil && !seen[g] {
				seen[g] = true
				getters = append(getters, g)
			}
		}
	}
	keys, found := analysis.collectDestructuringKeys(pattern, targetElem)
	precedingKeySet := map[string]bool{}
	hasDynamicPrecedingKey := !found
	for _, k := range keys {
		if k.known {
			precedingKeySet[k.name] = true
		} else {
			hasDynamicPrecedingKey = true
		}
	}
	if allGetters {
		appendGetters(gs.allGetters)
		appendGetters(gs.allShadowedProtoGetters)
		if gs.hasOpaqueProto && (len(keys) > 0 || !found) {
			unresolved = true
		}
	} else if gs.hasOwnDynamic || !found {
		appendGetters(gs.allGetters)
		if mutableObj {
			appendGetters(gs.allShadowedProtoGetters)
		}
		if gs.hasProtoDynamic {
			appendGetters(gs.protoDynamicGetters)
			if gs.hasOpaqueProto && (len(keys) > 0 || !found) {
				unresolved = true
			}
		}
	} else {
		for _, k := range keys {
			if !k.known {
				appendGetters(gs.allGetters)
				if mutableObj {
					appendGetters(gs.allShadowedProtoGetters)
				}
				if gs.hasOpaqueProto {
					unresolved = true
				}
				continue
			}
			appendGetters(gs.staticGetters[k.name])
			if mutableObj {
				appendGetters(gs.shadowedProtoGetters[k.name])
			}
			if gs.hasProtoDynamic && (!gs.ownShadowedProps[k.name] || mutableObj) {
				appendGetters(gs.protoDynamicGetters)
				if gs.hasOpaqueProto {
					unresolved = true
				}
			}
		}
	}
	if len(getters) == 0 {
		return nil, nil, unresolved
	}
	matchesTarget := func(pName string, pKnown bool) bool {
		return targetProp == "" || !pKnown || pName == targetProp
	}
	thisPropWrittenInGetters := map[string]bool{}
	gettersWriteDynamicThisProp := false
	scanThisPropWrites := func(body *ast.Node, isThis func(*ast.Node) bool) {
		var scan func(*ast.Node)
		scan = func(node *ast.Node) {
			if node == nil {
				return
			}
			if node != body && testFramework.IsFunction(node) && !isImmediatelyInvokedFunction(node) {
				return
			}
			if node.Kind == ast.KindBinaryExpression {
				if bin := node.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil && ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
					if rootAccess, isDirect, ok := thisAccessRootProp(bin.Left, isThis); ok && isDirect {
						if pName, pKnown := analysis.evalExpectConfigAccessName(rootAccess); pKnown {
							thisPropWrittenInGetters[pName] = true
						} else {
							gettersWriteDynamicThisProp = true
						}
					}
				}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				scan(child)
				return false
			})
		}
		scan(body)
	}
	for _, g := range getters {
		if g != nil && g.Body() != nil {
			scanThisPropWrites(g.Body(), func(e *ast.Node) bool {
				e = internalUtils.SkipAssertionsAndParens(e)
				return e != nil && e.Kind == ast.KindThisKeyword
			})
		}
	}
	exprReferencesParams := func(expr *ast.Node, paramSymbols map[*ast.Symbol]bool) bool {
		if expr == nil || len(paramSymbols) == 0 {
			return false
		}
		foundParam := false
		var walk func(*ast.Node)
		walk = func(node *ast.Node) {
			if node == nil || foundParam || ast.IsTypeNode(node) {
				return
			}
			if node.Kind == ast.KindIdentifier {
				if node.AsIdentifier().Text == "arguments" {
					foundParam = true
					return
				}
				if sym := resolveRstestRootSymbol(analysis.ctx, node); sym != nil && paramSymbols[sym] {
					foundParam = true
					return
				}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				walk(child)
				return foundParam
			})
		}
		walk(expr)
		return foundParam
	}
	collectFnParamSymbols := func(fn *ast.Node, base map[*ast.Symbol]bool) map[*ast.Symbol]bool {
		if fn == nil {
			return base
		}
		params := fn.Parameters()
		if len(params) == 0 {
			return base
		}
		out := make(map[*ast.Symbol]bool, len(base)+len(params))
		for s := range base {
			out[s] = true
		}
		var addBinding func(*ast.Node)
		addBinding = func(n *ast.Node) {
			if n == nil {
				return
			}
			switch n.Kind {
			case ast.KindIdentifier:
				sym := n.Parent.Symbol()
				if sym == nil {
					sym = resolveRstestRootSymbol(analysis.ctx, n)
				}
				if sym != nil {
					out[sym] = true
				}
			case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
				if bp := n.AsBindingPattern(); bp != nil && bp.Elements != nil {
					for _, el := range bp.Elements.Nodes {
						if el != nil && el.Kind == ast.KindBindingElement {
							addBinding(el.Name())
						}
					}
				}
			}
		}
		for _, p := range params {
			if p != nil && p.Name() != nil {
				addBinding(p.Name())
			}
		}
		return out
	}
	visitedBodies := map[*ast.Node]bool{}
	var visitBody func(body *ast.Node, ownThis bool, outerAliases map[*ast.Symbol]bool, paramSymbols map[*ast.Symbol]bool)
	visitBody = func(body *ast.Node, ownThis bool, outerAliases map[*ast.Symbol]bool, paramSymbols map[*ast.Symbol]bool) {
		if body == nil || visitedBodies[body] {
			return
		}
		visitedBodies[body] = true
		var thisAliases map[*ast.Symbol]bool
		if len(outerAliases) > 0 {
			thisAliases = make(map[*ast.Symbol]bool, len(outerAliases))
			for sym := range outerAliases {
				thisAliases[sym] = true
			}
		}
		var localFns map[*ast.Symbol][]*ast.Node
		addLocalFn := func(sym *ast.Symbol, fn *ast.Node) {
			if sym == nil || fn == nil {
				return
			}
			if localFns == nil {
				localFns = map[*ast.Symbol][]*ast.Node{}
			}
			localFns[sym] = append(localFns[sym], fn)
		}
		var collectLocalBindings func(*ast.Node)
		collectLocalBindings = func(node *ast.Node) {
			if node == nil {
				return
			}
			if node != body && testFramework.IsFunction(node) {
				if node.Kind == ast.KindFunctionDeclaration && node.Name() != nil && node.Name().Kind == ast.KindIdentifier {
					sym := node.Symbol()
					if sym == nil {
						sym = resolveRstestRootSymbol(analysis.ctx, node.Name())
					}
					addLocalFn(sym, node)
				}
				if node.Kind != ast.KindArrowFunction || !isImmediatelyInvokedFunction(node) {
					return
				}
			}
			switch node.Kind {
			case ast.KindVariableDeclaration:
				vd := node.AsVariableDeclaration()
				if vd != nil && vd.Initializer != nil && vd.Name() != nil && vd.Name().Kind == ast.KindIdentifier {
					init := internalUtils.SkipAssertionsAndParens(vd.Initializer)
					sym := node.Symbol()
					if sym == nil {
						sym = resolveRstestRootSymbol(analysis.ctx, vd.Name())
					}
					if sym != nil && init != nil {
						if ownThis && init.Kind == ast.KindThisKeyword {
							if thisAliases == nil {
								thisAliases = map[*ast.Symbol]bool{}
							}
							thisAliases[sym] = true
						} else if init.Kind == ast.KindArrowFunction || init.Kind == ast.KindFunctionExpression {
							addLocalFn(sym, init)
						}
					}
				}
			case ast.KindBinaryExpression:
				bin := node.AsBinaryExpression()
				if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken {
					if lhs := internalUtils.SkipAssertionsAndParens(bin.Left); lhs != nil && lhs.Kind == ast.KindIdentifier {
						if rhs := internalUtils.SkipAssertionsAndParens(bin.Right); rhs != nil {
							sym := resolveRstestRootSymbol(analysis.ctx, lhs)
							if sym != nil {
								if ownThis && rhs.Kind == ast.KindThisKeyword {
									if thisAliases == nil {
										thisAliases = map[*ast.Symbol]bool{}
									}
									thisAliases[sym] = true
								} else if rhs.Kind == ast.KindArrowFunction || rhs.Kind == ast.KindFunctionExpression {
									addLocalFn(sym, rhs)
								}
							}
						}
					}
				}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				collectLocalBindings(child)
				return false
			})
		}
		collectLocalBindings(body)
		isThis := func(e *ast.Node) bool {
			e = internalUtils.SkipAssertionsAndParens(e)
			if e == nil {
				return false
			}
			if ownThis && e.Kind == ast.KindThisKeyword {
				return true
			}
			if e.Kind == ast.KindIdentifier && len(thisAliases) > 0 {
				if sym := resolveRstestRootSymbol(analysis.ctx, e); sym != nil && thisAliases[sym] {
					return true
				}
			}
			return false
		}
		isConsoleCallee := func(callee *ast.Node) bool {
			if callee == nil || !ast.IsAccessExpression(callee) {
				return false
			}
			recv := internalUtils.SkipAssertionsAndParens(callee.Expression())
			return recv != nil && recv.Kind == ast.KindIdentifier && recv.AsIdentifier().Text == "console"
		}
		scanThisPropWrites(body, isThis)
		recordDirectWriteExpr := func(rhs *ast.Node) {
			exprs = append(exprs, rhs)
			if !isSimplePrimitiveLiteralExpr(rhs) && exprReferencesParams(rhs, paramSymbols) {
				unresolved = true
			}
		}
		visitLocalFn := func(fn *ast.Node) {
			if fn == nil || fn.Body() == nil {
				return
			}
			fnOwnThis := ownThis && fn.Kind == ast.KindArrowFunction
			visitBody(fn.Body(), fnOwnThis, thisAliases, collectFnParamSymbols(fn, paramSymbols))
		}
		var visit func(*ast.Node)
		visit = func(node *ast.Node) {
			if node == nil {
				return
			}
			if node != body && testFramework.IsFunction(node) {
				if isImmediatelyInvokedFunction(node) {
					visitLocalFn(node)
				}
				return
			}
			switch node.Kind {
			case ast.KindBinaryExpression:
				bin := node.AsBinaryExpression()
				if bin != nil && bin.OperatorToken != nil && ast.IsAssignmentOperator(bin.OperatorToken.Kind) &&
					!internalUtils.IsDefaultValueInDestructuringAssignment(node) {
					lhs := internalUtils.SkipAssertionsAndParens(bin.Left)
					if rootAccess, isDirect, ok := thisAccessRootProp(lhs, isThis); ok {
						pName, pKnown := analysis.evalExpectConfigAccessName(rootAccess)
						if matchesTarget(pName, pKnown) {
							if isDirect {
								recordDirectWriteExpr(bin.Right)
							} else if !isSimplePrimitiveLiteralExpr(bin.Right) {
								unresolved = true
							}
						}
					}
				}
			case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
				if internalUtils.IsWriteReference(node) {
					outer := outermostExpressionWrapper(node)
					isDirectBinLHS := outer != nil && outer.Parent != nil &&
						outer.Parent.Kind == ast.KindBinaryExpression &&
						outer.Parent.AsBinaryExpression().Left == outer &&
						!internalUtils.IsDefaultValueInDestructuringAssignment(outer.Parent)
					if !isDirectBinLHS {
						if rootAccess, _, ok := thisAccessRootProp(node, isThis); ok {
							pName, pKnown := analysis.evalExpectConfigAccessName(rootAccess)
							if matchesTarget(pName, pKnown) {
								unresolved = true
							}
						}
					}
				}
			case ast.KindDeleteExpression:
				if target := internalUtils.SkipAssertionsAndParens(node.AsDeleteExpression().Expression); target != nil {
					if rootAccess, isDirect, ok := thisAccessRootProp(target, isThis); ok {
						pName, pKnown := analysis.evalExpectConfigAccessName(rootAccess)
						if isDirect {
							if !pKnown {
								unresolved = true
							} else if allGetters || hasDynamicPrecedingKey || precedingKeySet[pName] {
								for _, sg := range gs.shadowedProtoGetters[pName] {
									visitBody(sg.Body(), true, nil, nil)
								}
								if gs.hasProtoDynamic {
									for _, dg := range gs.protoDynamicGetters {
										visitBody(dg.Body(), true, nil, nil)
									}
									if gs.hasOpaqueProto {
										unresolved = true
									}
								}
							}
						} else if matchesTarget(pName, pKnown) {
							unresolved = true
						}
					}
				}
			case ast.KindCallExpression:
				args := node.Arguments()
				callee := internalUtils.SkipAssertionsAndParens(node.Expression())
				handledBuiltin := false
				if len(localFns) > 0 && callee != nil {
					if callee.Kind == ast.KindIdentifier {
						if sym := resolveRstestRootSymbol(analysis.ctx, callee); sym != nil {
							for _, fn := range localFns[sym] {
								visitLocalFn(fn)
							}
						}
					} else if ast.IsAccessExpression(callee) {
						if mName, ok := internalUtils.AccessExpressionStaticName(callee); ok && (mName == "call" || mName == "apply") {
							if recv := internalUtils.SkipAssertionsAndParens(callee.Expression()); recv != nil && recv.Kind == ast.KindIdentifier {
								if sym := resolveRstestRootSymbol(analysis.ctx, recv); sym != nil {
									for _, fn := range localFns[sym] {
										visitLocalFn(fn)
									}
								}
							}
						}
					}
				}
				if rootAccess, isDirect, ok := thisAccessRootProp(callee, isThis); ok {
					if isDirect {
						mName, mOK := analysis.evalExpectConfigAccessName(callee)
						if mOK && mName == "__defineGetter__" && len(args) >= 2 && !hasSpreadBeforeIndex(args, 2) {
							kName, kOK := analysis.staticStringEvaluator().EvalToString(internalUtils.SkipAssertionsAndParens(args[0]))
							if matchesTarget(kName, kOK) {
								fnArg := internalUtils.SkipAssertionsAndParens(args[1])
								if fnArg != nil && (fnArg.Kind == ast.KindArrowFunction || fnArg.Kind == ast.KindFunctionExpression) {
									for _, rExpr := range collectGetterReturnExprs(fnArg) {
										recordDirectWriteExpr(rExpr)
									}
								} else {
									unresolved = true
								}
							}
						} else if mOK && len(args) == 0 && !allGetters && !mutableObj && !gettersWriteDynamicThisProp && !thisPropWrittenInGetters[mName] {
							if methodFn := analysis.lookupOwnLiteralInlineMethod(objExpr, mName); methodFn != nil && methodFn.Body() != nil {
								visitBody(methodFn.Body(), true, nil, nil)
							} else {
								unresolved = true
							}
						} else {
							unresolved = true
						}
					} else {
						pName, pKnown := analysis.evalExpectConfigAccessName(rootAccess)
						if matchesTarget(pName, pKnown) {
							unresolved = true
						}
					}
				}
				if len(args) >= 2 && isThis(args[0]) {
					if analysis.isGlobalBuiltinMemberCall(callee, "Object", "assign") {
						handledBuiltin = true
						for _, src := range args[1:] {
							if src == nil || src.Kind == ast.KindSpreadElement {
								unresolved = true
								continue
							}
							if srcObj, mutableSrc := analysis.resolveObjectLiteralForBoundLookup(src); srcObj != nil {
								var propExprs []*ast.Node
								if targetProp == "" {
									propExprs = analysis.allObjectLiteralPropertyExprs(srcObj)
								} else {
									propExprs = analysis.lookupObjectLiteralProperty(srcObj, targetProp, 1).possibleExprs
								}
								for _, pe := range propExprs {
									recordDirectWriteExpr(pe)
								}
								if mutableSrc {
									containers = append(containers, src)
								}
							} else {
								containers = append(containers, src)
								if exprReferencesParams(src, paramSymbols) {
									unresolved = true
								}
							}
						}
					} else if analysis.isGlobalBuiltinMemberCall(callee, "Reflect", "set") {
						handledBuiltin = true
						if hasSpreadBeforeIndex(args, 3) || len(args) < 3 {
							unresolved = true
						} else {
							kName, kOK := analysis.staticStringEvaluator().EvalToString(internalUtils.SkipAssertionsAndParens(args[1]))
							if matchesTarget(kName, kOK) {
								recordDirectWriteExpr(args[2])
							}
						}
					} else if analysis.isGlobalBuiltinMemberCall(callee, "Object", "defineProperty") ||
						analysis.isGlobalBuiltinMemberCall(callee, "Reflect", "defineProperty") {
						handledBuiltin = true
						if hasSpreadBeforeIndex(args, 3) || len(args) < 3 {
							unresolved = true
						} else {
							kName, kOK := analysis.staticStringEvaluator().EvalToString(internalUtils.SkipAssertionsAndParens(args[1]))
							if matchesTarget(kName, kOK) {
								descObj, mutableDesc := analysis.resolveObjectLiteralForBoundLookup(args[2])
								dExprs, complete := analysis.collectDescriptorValueAndGetExprs(descObj)
								for _, de := range dExprs {
									recordDirectWriteExpr(de)
								}
								if !complete || mutableDesc {
									unresolved = true
								}
							}
						}
					} else if analysis.isGlobalBuiltinMemberCall(callee, "Object", "defineProperties") {
						handledBuiltin = true
						if hasSpreadBeforeIndex(args, 2) {
							unresolved = true
						} else {
							propsObj, mutableProps := analysis.resolveObjectLiteralForBoundLookup(args[1])
							if propsObj == nil || mutableProps || propsObj.AsObjectLiteralExpression().Properties == nil {
								unresolved = true
							} else {
								for _, p := range propsObj.AsObjectLiteralExpression().Properties.Nodes {
									if p == nil {
										continue
									}
									if p.Kind != ast.KindPropertyAssignment || p.Name() == nil {
										unresolved = true
										continue
									}
									kName, kOK, _ := analysis.evalPropertyNameKey(p.Name())
									if matchesTarget(kName, kOK) {
										descObj, mutableDesc := analysis.resolveObjectLiteralForBoundLookup(p.AsPropertyAssignment().Initializer)
										dExprs, complete := analysis.collectDescriptorValueAndGetExprs(descObj)
										for _, de := range dExprs {
											recordDirectWriteExpr(de)
										}
										if !complete || mutableDesc {
											unresolved = true
										}
									}
								}
							}
						}
					}
				}
				if !handledBuiltin && !isConsoleCallee(callee) {
					for _, arg := range args {
						if len(localFns) > 0 {
							if uArg := internalUtils.SkipAssertionsAndParens(arg); uArg != nil && uArg.Kind == ast.KindIdentifier {
								if sym := resolveRstestRootSymbol(analysis.ctx, uArg); sym != nil {
									for _, fn := range localFns[sym] {
										visitLocalFn(fn)
									}
								}
							}
						}
						if isThis(arg) {
							unresolved = true
							break
						}
						if rootAccess, _, ok := thisAccessRootProp(arg, isThis); ok {
							pName, pKnown := analysis.evalExpectConfigAccessName(rootAccess)
							if matchesTarget(pName, pKnown) {
								unresolved = true
								break
							}
						}
					}
				}
			}
			node.ForEachChild(func(child *ast.Node) bool {
				visit(child)
				return false
			})
		}
		visit(body)
	}
	for _, g := range getters {
		visitBody(g.Body(), true, nil, nil)
	}
	return exprs, containers, unresolved
}

// lookupObjectLiteralProperty inspects `sourceExpr` when it is an object literal
// for static property `propName`, including spreads, getters, and non-computed
// `__proto__` chains.
func (analysis *RstestCallAnalysis) lookupObjectLiteralProperty(
	sourceExpr *ast.Node,
	propName string,
	depth int,
) objectLiteralPropertyLookup {
	return analysis.lookupObjectLiteralPropertyWithGetterRisk(sourceExpr, propName, false, depth)
}

func (analysis *RstestCallAnalysis) lookupObjectLiteralPropertyWithGetterRisk(
	sourceExpr *ast.Node,
	propName string,
	priorGetterCanRun bool,
	depth int,
) objectLiteralPropertyLookup {
	if depth > 4 {
		return objectLiteralPropertyLookup{mayHaveUnknownSource: true}
	}
	obj := internalUtils.SkipAssertionsAndParens(sourceExpr)
	if obj == nil || obj.Kind != ast.KindObjectLiteralExpression || obj.AsObjectLiteralExpression().Properties == nil {
		return objectLiteralPropertyLookup{}
	}
	var res objectLiteralPropertyLookup
	var protoExpr *ast.Node
	hasGetter := false
	for _, prop := range obj.AsObjectLiteralExpression().Properties.Nodes {
		if prop == nil {
			continue
		}
		if prop.Kind == ast.KindSpreadAssignment {
			// A later spread may overwrite the property with `undefined` or
			// another value (including a bound callable from a literal/const
			// spread target), while leaving any earlier explicit value intact
			// if the spread source lacks `propName`.
			res.exactExpr = nil
			res.definitelyDefined = false
			res.mayHaveUnknownSource = true
			res.fromNonFreshSource = true
			if spreadObj, _ := analysis.resolveObjectLiteralForBoundLookup(prop.AsSpreadAssignment().Expression); spreadObj != nil {
				spreadLookup := analysis.lookupObjectLiteralProperty(spreadObj, propName, depth+1)
				res.possibleExprs = append(res.possibleExprs, spreadLookup.possibleExprs...)
			}
			continue
		}
		pName := prop.Name()
		if pName == nil {
			res.exactExpr = nil
			res.definitelyDefined = false
			res.mayHaveUnknownSource = true
			continue
		}
		name, known, staticKey := analysis.evalPropertyNameKey(pName)
		if !known {
			if !staticKey {
				res.exactExpr = nil
				res.definitelyDefined = false
				res.mayHaveUnknownSource = true
				switch prop.Kind {
				case ast.KindPropertyAssignment:
					res.possibleExprs = append(res.possibleExprs, prop.AsPropertyAssignment().Initializer)
				case ast.KindShorthandPropertyAssignment:
					res.possibleExprs = append(res.possibleExprs, prop.AsShorthandPropertyAssignment().Name())
				case ast.KindGetAccessor:
					res.possibleExprs = append(res.possibleExprs, collectGetterReturnExprs(prop)...)
				}
			}
			continue
		}
		// Per ECMAScript B.3.1, non-computed `__proto__: expr` / `"__proto__": expr`
		// sets the object's prototype rather than defining an own `"__proto__"` property.
		if prop.Kind == ast.KindPropertyAssignment && name == "__proto__" && pName.Kind != ast.KindComputedPropertyName {
			protoExpr = prop.AsPropertyAssignment().Initializer
			continue
		}
		if name != propName {
			continue
		}
		res.foundExplicit = true
		res.mayHaveUnknownSource = false
		res.fromNonFreshSource = false
		switch prop.Kind {
		case ast.KindMethodDeclaration:
			hasGetter = false
			res.exactExpr = nil
			res.possibleExprs = nil
			res.definitelyDefined = true
		case ast.KindPropertyAssignment:
			hasGetter = false
			val := prop.AsPropertyAssignment().Initializer
			res.exactExpr = val
			res.possibleExprs = []*ast.Node{val}
			res.definitelyDefined = analysis.isDefinitelyDefinedExpr(val, 0, nil)
		case ast.KindShorthandPropertyAssignment:
			hasGetter = false
			val := prop.AsShorthandPropertyAssignment().Name()
			res.exactExpr = val
			res.possibleExprs = []*ast.Node{val}
			res.definitelyDefined = analysis.isDefinitelyDefinedExpr(val, 0, nil)
		case ast.KindGetAccessor:
			hasGetter = true
			res.exactExpr = nil
			res.possibleExprs = collectGetterReturnExprs(prop)
			res.definitelyDefined = false
		case ast.KindSetAccessor:
			res.exactExpr = nil
			if !hasGetter {
				res.possibleExprs = nil
			}
			res.definitelyDefined = false
		default:
			hasGetter = false
			res.exactExpr = nil
			res.possibleExprs = nil
			res.definitelyDefined = false
		}
	}
	// If an earlier destructured property can invoke a getter on `obj` (or its
	// `__proto__`), that getter can mutate `this[propName] = undefined` or
	// `delete this[propName]` (exposing `__proto__[propName]`) before `propName`
	// is read. An unread getter elsewhere on `obj` cannot execute beforehand.
	if priorGetterCanRun {
		res.definitelyDefined = false
		res.exactExpr = nil
		res.mayHaveUnknownSource = true
	}
	// Resolve inherited properties from non-computed `__proto__: protoExpr` when
	// no own property shadows it, OR when an earlier getter could `delete` the
	// own property before `propName` is read.
	if protoExpr != nil && (!res.foundExplicit || priorGetterCanRun) && propName != "__proto__" {
		unwrappedProto := internalUtils.SkipAssertionsAndParens(protoExpr)
		if protoObj, mutableProto := analysis.resolveObjectLiteralForBoundLookup(protoExpr); protoObj != nil {
			protoLookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(protoObj, propName, priorGetterCanRun, depth+1)
			res.possibleExprs = append(res.possibleExprs, protoLookup.possibleExprs...)
			res.exposedProtoExprs = append(res.exposedProtoExprs, protoLookup.exposedProtoExprs...)
			if unwrappedProto == nil || unwrappedProto.Kind != ast.KindObjectLiteralExpression || protoLookup.fromNonFreshSource {
				res.fromNonFreshSource = true
			}
			if mutableProto || protoLookup.mayHaveUnknownSource {
				res.mayHaveUnknownSource = true
			} else if !res.foundExplicit && !res.mayHaveUnknownSource && !priorGetterCanRun && protoLookup.definitelyDefined {
				res.definitelyDefined = true
				res.exactExpr = protoLookup.exactExpr
			}
		} else if unwrappedProto != nil && unwrappedProto.Kind != ast.KindNullKeyword {
			res.exposedProtoExprs = append(res.exposedProtoExprs, protoExpr)
			res.mayHaveUnknownSource = true
			res.fromNonFreshSource = true
		}
	}
	return res
}

func (analysis *RstestCallAnalysis) isDefinitelyDefinedExpr(expr *ast.Node, depth int, visited map[*ast.Symbol]bool) bool {
	if depth > 8 || expr == nil {
		return false
	}
	expr = internalUtils.SkipAssertionsAndParens(expr)
	if expr == nil {
		return false
	}
	switch expr.Kind {
	case ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindClassExpression,
		ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
		ast.KindNewExpression, ast.KindRegularExpressionLiteral,
		ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression,
		ast.KindNumericLiteral, ast.KindBigIntLiteral,
		ast.KindTrueKeyword, ast.KindFalseKeyword, ast.KindNullKeyword:
		return true
	case ast.KindIdentifier:
		sym := resolveRstestRootSymbol(analysis.ctx, expr)
		if sym == nil || visited[sym] {
			return false
		}
		// Require a single concrete same-file value declaration (ignore type
		// declarations and function overload signatures).
		var decl *ast.Node
		for _, d := range sym.Declarations {
			if d == nil {
				continue
			}
			switch d.Kind {
			case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindTypeParameter:
				continue
			case ast.KindFunctionDeclaration:
				if d.AsFunctionDeclaration().Body == nil {
					continue
				}
			}
			if decl != nil {
				return false
			}
			decl = d
		}
		if decl == nil || analysis.ctx.SourceFile == nil || ast.GetSourceFileOfNode(decl) != analysis.ctx.SourceFile {
			return false
		}
		if internalUtils.IsInAmbientContext(decl) || ast.HasSyntacticModifier(decl, ast.ModifierFlagsAmbient) {
			return false
		}
		switch decl.Kind {
		case ast.KindFunctionDeclaration:
			// Only unconditionally hoisted function declarations (top-level or
			// direct function body) are guaranteed initialized; block/Annex B
			// conditional function declarations can evaluate to `undefined`.
			if decl.Parent == nil {
				return false
			}
			switch decl.Parent.Kind {
			case ast.KindSourceFile, ast.KindModuleBlock:
			case ast.KindBlock:
				if decl.Parent.Parent == nil || !testFramework.IsFunction(decl.Parent.Parent) {
					return false
				}
			default:
				return false
			}
			return analysis.hasNoExternalWrites(sym, decl)
		case ast.KindClassDeclaration:
			return analysis.hasNoExternalWrites(sym, decl)
		case ast.KindVariableDeclaration:
			if ast.IsVarConst(decl) && decl.AsVariableDeclaration().Initializer != nil {
				init := decl.AsVariableDeclaration().Initializer
				if nodeContains(init, expr) {
					return false
				}
				if visited == nil {
					visited = map[*ast.Symbol]bool{}
				}
				visited[sym] = true
				defer delete(visited, sym)
				return analysis.isDefinitelyDefinedExpr(init, depth+1, visited)
			}
		}
	}
	return false
}

func (analysis *RstestCallAnalysis) hasNoExternalWrites(sym *ast.Symbol, decl *ast.Node) bool {
	if analysis.ctx.Refs == nil || sym == nil {
		return false
	}
	if cached, ok := analysis.symbolNoExternalWrites[sym]; ok {
		return cached
	}
	clean := true
	for _, ref := range analysis.ctx.Refs.References(sym) {
		if ref == nil || !internalUtils.IsWriteReference(ref) {
			continue
		}
		if ref.Parent == decl && decl.Name() == ref {
			continue
		}
		clean = false
		break
	}
	if analysis.symbolNoExternalWrites == nil {
		analysis.symbolNoExternalWrites = map[*ast.Symbol]bool{}
	}
	analysis.symbolNoExternalWrites[sym] = clean
	return clean
}

func (analysis *RstestCallAnalysis) assignmentDestructuredExpectConfigMember(prop, defaultExpr *ast.Node) expectConfigKind {
	if prop == nil || prop.Parent == nil || prop.Parent.Kind != ast.KindObjectLiteralExpression {
		return expectConfigNone
	}
	obj := prop.Parent
	if prop.Kind == ast.KindSpreadAssignment {
		container := analysis.expectConfigAssignmentObjectKind(obj)
		return container & (expectConfigReceiver | expectConfigExpect)
	}
	propName := prop.Name()
	if propName == nil {
		return expectConfigNone
	}
	var out expectConfigKind
	member, known, staticKey := analysis.evalPropertyNameKey(propName)
	defaultSkipped := false
	sources, definiteSrc, sourcesMayBeMutable, sourcesUnresolvedGetter := analysis.assignmentObjectSourceExprs(obj)
	if sourcesUnresolvedGetter {
		out |= expectConfigDynamicMember
	}
	if known {
		if definiteSrc != nil {
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(definiteSrc); objExpr != nil && !mutableObj {
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(obj, prop, objExpr, false)
				defLookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter, 0)
				defaultSkipped = defLookup.definitelyDefined
			}
		}
		for _, srcExpr := range sources {
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(srcExpr); objExpr != nil {
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(obj, prop, objExpr, mutableObj)
				if mutableObj && !sourcesMayBeMutable && !priorGetter && len(sources) == 1 {
					if unwrapped := internalUtils.SkipAssertionsAndParens(srcExpr); unwrapped != nil && unwrapped.Kind == ast.KindIdentifier {
						if sym := resolveRstestRootSymbol(analysis.ctx, unwrapped); sym != nil {
							if replacement, replaced := analysis.unconditionalPropertyReplacementBefore(sym, objExpr, member, unwrapped); replaced {
								out |= analysis.expectConfigMember(replacement)
								defaultSkipped = analysis.isDefinitelyDefinedExpr(replacement, 0, nil)
								continue
							}
						}
					}
				}
				lookup := analysis.lookupObjectLiteralPropertyWithGetterRisk(objExpr, member, priorGetter || mutableObj || sourcesMayBeMutable, 0)
				for _, expr := range lookup.possibleExprs {
					out |= analysis.expectConfigMember(expr)
				}
				useAllGetters := sourcesMayBeMutable || (mutableObj && analysis.srcExprObjEscaped(srcExpr, objExpr))
				if priorGetter || useAllGetters || mutableObj {
					gExprs, gContainers, gUnresolved := analysis.collectPrecedingGetterPropWriteExprs(obj, prop, objExpr, member, useAllGetters, mutableObj)
					for _, expr := range gExprs {
						out |= analysis.expectConfigMember(expr)
					}
					for _, cExpr := range gContainers {
						out |= classifyDestructuredExpectConfigProperty(analysis.expectConfigMember(cExpr), member, true, true)
					}
					if gUnresolved {
						out |= expectConfigDynamicMember
					}
				}
				for _, pExpr := range lookup.exposedProtoExprs {
					out |= classifyDestructuredExpectConfigProperty(analysis.expectConfigMember(pExpr), member, true, true)
				}
			}
		}
	} else if !staticKey {
		for _, srcExpr := range sources {
			for _, expr := range analysis.allObjectLiteralPropertyExprs(srcExpr) {
				out |= analysis.expectConfigMember(expr)
			}
			if objExpr, mutableObj := analysis.resolveObjectLiteralForBoundLookup(srcExpr); objExpr != nil {
				priorGetter := analysis.precedingDestructuringMayInvokeGetter(obj, prop, objExpr, mutableObj)
				useAllGetters := sourcesMayBeMutable || (mutableObj && analysis.srcExprObjEscaped(srcExpr, objExpr))
				if priorGetter || useAllGetters || mutableObj {
					gExprs, gContainers, gUnresolved := analysis.collectPrecedingGetterPropWriteExprs(obj, prop, objExpr, "", useAllGetters, mutableObj)
					for _, expr := range gExprs {
						out |= analysis.expectConfigMember(expr)
					}
					for _, cExpr := range gContainers {
						out |= classifyDestructuredExpectConfigProperty(analysis.expectConfigMember(cExpr), "", false, false)
					}
					if gUnresolved {
						out |= expectConfigDynamicMember
					}
				}
			}
		}
	}
	if defaultExpr != nil && !defaultSkipped {
		out |= analysis.expectConfigMember(defaultExpr)
	}
	container := analysis.expectConfigAssignmentObjectKind(obj)
	if container == expectConfigNone {
		return out
	}
	return out | classifyDestructuredExpectConfigProperty(container, member, known, staticKey)
}

func (analysis *RstestCallAnalysis) expectConfigAssignmentObjectKind(obj *ast.Node) expectConfigKind {
	if obj == nil || obj.Parent == nil {
		return expectConfigNone
	}
	outer := internalUtils.OutermostParenthesizedExpression(obj)
	if outer == nil || outer.Parent == nil {
		return expectConfigNone
	}
	switch outer.Parent.Kind {
	case ast.KindBinaryExpression:
		bin := outer.Parent.AsBinaryExpression()
		if bin != nil && bin.OperatorToken != nil && bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Left == outer {
			// Nested destructuring assignment default, e.g.
			// `({ expect: { extend } = {} } = core)`: pass `bin.Right` as the
			// parent PropertyAssignment's default so definite parent properties
			// skip the unused fallback symmetrically with binding patterns.
			if outer.Parent.Parent != nil && outer.Parent.Parent.Kind == ast.KindPropertyAssignment &&
				outer.Parent.Parent.AsPropertyAssignment().Initializer == outer.Parent {
				return analysis.assignmentDestructuredExpectConfigMember(outer.Parent.Parent, bin.Right)
			}
			return analysis.expectConfigMember(bin.Right)
		}
	case ast.KindPropertyAssignment:
		return analysis.assignmentDestructuredExpectConfigMember(outer.Parent, nil)
	}
	return expectConfigNone
}

func classifyDestructuredExpectConfigProperty(container expectConfigKind, member string, known, staticKey bool) expectConfigKind {
	var out expectConfigKind
	if container&expectConfigExpect != 0 {
		if known {
			if member == "extend" && container&expectConfigExpectExtend != 0 {
				out |= expectConfigExtend
			}
			if member == "addEqualityTesters" && container&expectConfigExpectEquality != 0 {
				out |= expectConfigAddEqualityTesters
			}
		} else if !staticKey {
			out |= expectConfigDynamicMember
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
		return analysis.destructuredExpectConfigMember(pattern.Parent)
	case ast.KindParameter:
		if analysis.callbacksRef().ContextParameters[pattern.Parent] {
			return expectConfigReceiver
		}
		return expectConfigNone
	default:
		return expectConfigNone
	}
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
