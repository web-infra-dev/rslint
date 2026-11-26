package utils

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
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
	// Expect customization is computed lazily from calls because only rules
	// whose safety depends on the built-in matcher implementations need it. The
	// source-file call index is already shared by every Rstest rule, so this never
	// adds another AST walk.
	overriddenExpectMatchers    map[string]bool
	allExpectMatchersOverridden bool
	hasCustomEqualityTesters    bool
	expectCustomizationOK       bool
	expectConfigAliases         map[*ast.Symbol]expectConfigKind
	expectConfigSymbolInfo      map[*ast.Symbol]*expectConfigSymbolSummary
	expectConfigFnReach         map[*ast.Symbol]*expectConfigFnSummary
	expectConfigMustOverwrite   map[*ast.Symbol]*expectConfigOverwriteSummary
	expectConfigVisiting        map[*ast.Symbol]bool
	expectConfigCycleHits       int
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
		var arg *ast.Node
		hasArg := false
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
					arg = internalUtils.SkipAssertionsAndParens(args[0])
					hasArg = true
				}
			}
		}
		if kind == expectConfigNone {
			kind, arg, hasArg = analysis.expectConfigCallTarget(call)
		}
		if !hasArg || kind == expectConfigNone {
			continue
		}
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
				continue
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
				continue
			default:
				if _, isString := analysis.staticStringEvaluator().Eval(arg); isString {
					continue
				}
				analysis.allExpectMatchersOverridden = true
				analysis.hasCustomEqualityTesters = true
			}
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

func (analysis *RstestCallAnalysis) extractBoundCallFirstArg(node *ast.Node, depth int) (*ast.Node, bool) {
	if depth > 4 || node == nil {
		return nil, false
	}
	node = internalUtils.SkipAssertionsAndParens(node)
	if node == nil {
		return nil, false
	}
	if internalUtils.IsCommaOperator(node) {
		return analysis.extractBoundCallFirstArg(node.AsBinaryExpression().Right, depth+1)
	}
	if node.Kind == ast.KindCallExpression {
		callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if method, ok := analysis.evalExpectConfigAccessName(callee); ok && method == "bind" {
				bArgs := node.Arguments()
				if hasSpreadBeforeIndex(bArgs, 1) {
					return nil, true
				}
				if len(bArgs) >= 2 {
					return internalUtils.SkipAssertionsAndParens(bArgs[1]), true
				}
				return analysis.extractBoundCallFirstArg(callee.Expression(), depth+1)
			}
		}
		return nil, false
	}
	if node.Kind == ast.KindIdentifier {
		sym := resolveRstestRootSymbol(analysis.ctx, node)
		if sym == nil {
			return nil, false
		}
		summary := analysis.getExpectConfigSymbolSummary(sym)
		if len(summary.writes) == 1 && summary.writes[0].rhs != nil {
			return analysis.extractBoundCallFirstArg(summary.writes[0].rhs, depth+1)
		}
	}
	return nil, false
}

func (analysis *RstestCallAnalysis) expectConfigCallTarget(call *ast.Node) (kind expectConfigKind, arg *ast.Node, hasArg bool) {
	if call == nil || call.Kind != ast.KindCallExpression {
		return expectConfigNone, nil, false
	}
	args := call.Arguments()
	callee := internalUtils.SkipAssertionsAndParens(call.Expression())
	if callee == nil {
		return expectConfigNone, nil, false
	}
	// Handle `Reflect.apply(expect.extend, expect, [matchers])`.
	if analysis.isGlobalBuiltinMemberCall(callee, "Reflect", "apply") {
		if len(args) > 0 && args[0] != nil && args[0].Kind != ast.KindSpreadElement {
			target := analysis.expectConfigCallMember(args[0])
			if target != expectConfigNone {
				if hasSpreadBeforeIndex(args, 2) {
					return target, nil, true
				}
				if len(args) < 3 {
					return expectConfigNone, nil, false
				}
				argNode, ok := firstArgFromApplyArray(args[2])
				return target, argNode, ok
			}
		}
	}
	// Handle `expect.extend.call(expect, matchers)` and `expect.extend.apply(expect, [matchers])`.
	if ast.IsAccessExpression(callee) {
		if method, ok := analysis.evalExpectConfigAccessName(callee); ok && (method == "call" || method == "apply") {
			target := analysis.expectConfigCallMember(callee.Expression())
			if target != expectConfigNone {
				if boundArg, hasBound := analysis.extractBoundCallFirstArg(callee.Expression(), 0); hasBound {
					return target, boundArg, true
				}
				if hasSpreadBeforeIndex(args, 1) {
					return target, nil, true
				}
				if len(args) < 2 {
					return expectConfigNone, nil, false
				}
				if method == "call" {
					return target, internalUtils.SkipAssertionsAndParens(args[1]), true
				}
				argNode, ok := firstArgFromApplyArray(args[1])
				return target, argNode, ok
			}
		}
	}
	kind = analysis.expectConfigCallMember(callee)
	if kind == expectConfigNone {
		return expectConfigNone, nil, false
	}
	if boundArg, hasBound := analysis.extractBoundCallFirstArg(callee, 0); hasBound {
		return kind, boundArg, true
	}
	if len(args) == 0 {
		return expectConfigNone, nil, false
	}
	if args[0] != nil && args[0].Kind == ast.KindSpreadElement {
		return kind, nil, true
	}
	return kind, internalUtils.SkipAssertionsAndParens(args[0]), true
}

func (analysis *RstestCallAnalysis) recordExtendedMatchers(matchers *ast.Node) {
	if matchers == nil || matchers.Kind != ast.KindObjectLiteralExpression ||
		matchers.AsObjectLiteralExpression().Properties == nil {
		analysis.allExpectMatchersOverridden = true
		return
	}
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
	obj := objNode.AsObjectLiteralExpression()
	if obj == nil || obj.Properties == nil {
		return kind
	}
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
		name, ok := internalUtils.GetStaticPropertyName(propName)
		if !ok {
			name, ok = analysis.staticStringEvaluator().EvalPropertyName(propName)
		}
		if !ok {
			continue
		}
		var valKind expectConfigKind
		switch prop.Kind {
		case ast.KindPropertyAssignment:
			valKind = analysis.expectConfigMember(prop.AsPropertyAssignment().Initializer)
		case ast.KindShorthandPropertyAssignment:
			valKind = analysis.expectConfigIdentifierMember(prop.AsShorthandPropertyAssignment().Name())
		}
		switch name {
		case "extend":
			kind &^= expectConfigExpectExtend
			if valKind&expectConfigExtend != 0 {
				kind |= expectConfigExpectExtend
			}
		case "addEqualityTesters":
			kind &^= expectConfigExpectEquality
			if valKind&expectConfigAddEqualityTesters != 0 {
				kind |= expectConfigExpectEquality
			}
		case "expect":
			kind &^= expectConfigReceiver
			if valKind&expectConfigExpect != 0 {
				kind |= expectConfigReceiver
			}
		}
	}
	return kind
}

// expectConfigObjectLiteralSpreadMember resolves shallow-copy literals such as
// `{ ...expect }` or `{ ...r }`, clearing only the specific overwritten member
// bits unless the property assigns back a genuine expect/extend/tester value.
func (analysis *RstestCallAnalysis) expectConfigObjectLiteralSpreadMember(node *ast.Node) expectConfigKind {
	return analysis.applyObjectLiteralPropertiesToKind(expectConfigNone, node)
}

func (analysis *RstestCallAnalysis) expectConfigObjectAssignMember(args []*ast.Node) expectConfigKind {
	if len(args) == 0 {
		return expectConfigNone
	}
	var kind expectConfigKind
	for _, arg := range args {
		a := internalUtils.SkipAssertionsAndParens(arg)
		if a == nil || a.Kind == ast.KindSpreadElement {
			continue
		}
		if a.Kind == ast.KindObjectLiteralExpression {
			kind = analysis.applyObjectLiteralPropertiesToKind(kind, a)
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
	rhs             *ast.Node
	initialKind     expectConfigKind
	evaluatedMember expectConfigKind
}

type expectConfigBlockKill struct {
	pos     int
	stmtEnd int
}

type expectConfigSymbolSummary struct {
	writes              []expectConfigWrite
	blockKills          map[*ast.Node][]expectConfigBlockKill
	configWritesByFn    map[*ast.Node][]int // indices into writes sorted by exprEnd
	allConfigMask       expectConfigKind
	hasAnyConfigSource  bool
	evaluatedSources    bool
	shallowAliasTarget  *ast.Symbol
	isDirectShallowCopy bool
}

type expectConfigOverwriteKey struct {
	block *ast.Node
	bit   expectConfigKind
}

type expectConfigOverwriteEvent struct {
	pos     int
	stmtEnd int
	clears  bool
}

// expectConfigOverwriteSummary stores, for each (block, config bit), the
// sorted sequence of property overwrite/delete (`clears: true`) and restore
// (`clears: false`) events in that block. Queries binary-search the latest
// event before `useStmt` (and after the reaching write), running in O(log E)
// with sound MUST-overwrite semantics.
type expectConfigOverwriteSummary struct {
	events map[expectConfigOverwriteKey][]expectConfigOverwriteEvent
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
		return expectConfigNone
	}
	if analysis.expectConfigVisiting == nil {
		analysis.expectConfigVisiting = map[*ast.Symbol]bool{}
	}
	analysis.expectConfigVisiting[symbol] = true
	defer delete(analysis.expectConfigVisiting, symbol)

	summary := analysis.getExpectConfigSymbolSummary(symbol)
	if len(summary.writes) == 0 {
		analysis.cacheExpectConfigAlias(symbol, expectConfigNone)
		return expectConfigNone
	}

	// Evaluate every write's underlying config kind ONCE per symbol — unless
	// an active ancestor cycle truncated this evaluation, in which case we keep
	// results local so the cache is never permanently poisoned.
	var localEvaluated []expectConfigKind
	var allMask expectConfigKind
	if !summary.evaluatedSources {
		cycleBefore := analysis.expectConfigCycleHits
		localEvaluated = make([]expectConfigKind, len(summary.writes))
		for i := range summary.writes {
			k := analysis.evalExpectConfigWriteStatic(&summary.writes[i])
			localEvaluated[i] = k
			allMask |= k
		}
		if analysis.expectConfigCycleHits == cycleBefore {
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
		kind |= analysis.assignmentDestructuredExpectConfigMember(write.assignDestruct)
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
	// `Object.assign({}, expect)` also allocates a shallow copy.
	if node.Kind == ast.KindCallExpression {
		callee := internalUtils.SkipAssertionsAndParens(node.AsCallExpression().Expression)
		if callee != nil && ast.IsAccessExpression(callee) {
			if name, ok := internalUtils.AccessExpressionStaticName(callee); ok && name == "assign" {
				if recv := internalUtils.SkipAssertionsAndParens(internalUtils.AccessExpressionObject(callee)); recv != nil &&
					recv.Kind == ast.KindIdentifier && recv.AsIdentifier().Text == "Object" {
					return true
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
								exprEnd = outer.Parent.Parent.End()
							}
						}
					}
				case ast.KindShorthandPropertyAssignment:
					assignDestruct = outer.Parent
					exprEnd = enclosingDestructuringAssignmentEnd(outer.Parent)
					if sp := outer.Parent.AsShorthandPropertyAssignment(); sp != nil && sp.ObjectAssignmentInitializer != nil {
						rhs = sp.ObjectAssignmentInitializer
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
			shallow := isObjectLiteralSpread(rhs) || (assignDestruct != nil && assignDestruct.Kind == ast.KindSpreadAssignment)
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
				rhs:            rhs,
			})
		}
	}
	// Index unconditional writes per block in sorted order so kill checks run in O(log W).
	for i := range summary.writes {
		w := &summary.writes[i]
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
	var mask expectConfigKind
	visited := map[*ast.Symbol]bool{}
	first := true
	for cur := symbol; cur != nil && !visited[cur]; {
		visited[cur] = true
		var minBlock *ast.Node
		minPos := -1
		if first && write != nil {
			minBlock = write.block
			minPos = write.exprEnd
		}
		first = false
		mask |= analysis.symbolMustOverwrittenMaskAt(cur, useNode, minBlock, minPos)
		if info := analysis.expectConfigSymbolInfo[cur]; info != nil {
			cur = info.shallowAliasTarget
		} else {
			break
		}
	}
	return mask
}

func (analysis *RstestCallAnalysis) symbolMustOverwrittenMaskAt(
	symbol *ast.Symbol,
	useNode *ast.Node,
	writeBlock *ast.Node,
	writeExprEnd int,
) expectConfigKind {
	summary := analysis.getExpectConfigOverwriteSummary(symbol)
	if len(summary.events) == 0 {
		return expectConfigNone
	}
	var mask expectConfigKind
	var decided expectConfigKind
	// Walk upward from useNode across enclosing statements/blocks within the
	// same function (never crossing a function boundary into an outer scope).
	cur := useNode
	for cur != nil {
		stmt, block := enclosingStatementAndList(cur)
		if stmt == nil || block == nil {
			break
		}
		stmtPos := stmt.Pos()
		for _, bit := range []expectConfigKind{expectConfigExpectExtend, expectConfigExpectEquality, expectConfigReceiver} {
			if decided&bit != 0 {
				continue
			}
			events := summary.events[expectConfigOverwriteKey{block: block, bit: bit}]
			if len(events) == 0 {
				continue
			}
			// Find the latest event whose statement finished before `stmtPos`.
			idx := sort.Search(len(events), func(i int) bool {
				return events[i].stmtEnd > stmtPos
			})
			if idx > 0 {
				ev := events[idx-1]
				if block == writeBlock && ev.pos < writeExprEnd {
					// Overwrite happened before this reaching write (e.g. before
					// `mock = { ...expect }` replaced the object).
					decided |= bit
					continue
				}
				decided |= bit
				if ev.clears {
					mask |= bit
				}
			}
		}
		if block == writeBlock || block.Parent == nil || testFramework.IsFunction(block.Parent) ||
			block.Kind == ast.KindSourceFile || block.Kind == ast.KindModuleBlock {
			break
		}
		cur = block
	}
	return mask
}

func (analysis *RstestCallAnalysis) getExpectConfigOverwriteSummary(symbol *ast.Symbol) *expectConfigOverwriteSummary {
	if summary, ok := analysis.expectConfigMustOverwrite[symbol]; ok {
		return summary
	}
	summary := &expectConfigOverwriteSummary{}
	record := func(block *ast.Node, pos, stmtEnd int, bit expectConfigKind, clears bool) {
		if block == nil || bit == expectConfigNone {
			return
		}
		if summary.events == nil {
			summary.events = map[expectConfigOverwriteKey][]expectConfigOverwriteEvent{}
		}
		key := expectConfigOverwriteKey{block: block, bit: bit}
		summary.events[key] = append(summary.events[key], expectConfigOverwriteEvent{
			pos:     pos,
			stmtEnd: stmtEnd,
			clears:  clears,
		})
	}
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
						if !propStatic {
							break
						}
						bit := expectConfigOverwriteBit(propName)
						if bit == expectConfigNone {
							break
						}
						// If writing back a genuine expect/extend/tester value
						// (even conditionally), record a restore (`clears: false`)
						// so any earlier overwrite is invalidated.
						if isWrite && !isDelete && outerAccess != nil && outerAccess.Parent != nil && outerAccess.Parent.Kind == ast.KindBinaryExpression {
							if bin := outerAccess.Parent.AsBinaryExpression(); bin != nil && bin.OperatorToken != nil &&
								bin.OperatorToken.Kind == ast.KindEqualsToken && bin.Left == outerAccess {
								if analysis.overwritePreservesConfigBit(bit, bin.Right) {
									record(block, stmt.Pos(), stmt.End(), bit, false)
									break
								}
							}
						}
						if unconditional {
							record(block, stmt.Pos(), stmt.End(), bit, true)
						}
						break
					}
					continue
				}
				if firstAccess == nil && parent.Kind == ast.KindCallExpression {
					for bit, clears := range analysis.mutatingBuiltinOverwriteChanges(parent, current) {
						if !clears {
							record(block, stmt.Pos(), stmt.End(), bit, false)
						} else if unconditional {
							record(block, stmt.Pos(), stmt.End(), bit, true)
						}
					}
				}
				break
			}
		}
	}
	for _, events := range summary.events {
		sort.Slice(events, func(i, j int) bool {
			if events[i].stmtEnd != events[j].stmtEnd {
				return events[i].stmtEnd < events[j].stmtEnd
			}
			return events[i].pos < events[j].pos
		})
	}
	if analysis.expectConfigMustOverwrite == nil {
		analysis.expectConfigMustOverwrite = map[*ast.Symbol]*expectConfigOverwriteSummary{}
	}
	analysis.expectConfigMustOverwrite[symbol] = summary
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
		return k&expectConfigExtend != 0
	case expectConfigExpectEquality:
		return k&expectConfigAddEqualityTesters != 0
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

// propertyDescriptorChange classifies an `Object.defineProperty` /
// `Reflect.defineProperty` / `Object.defineProperties` descriptor object.
// Attribute-only descriptors such as `{ enumerable: false }` keep the existing
// property value (ok = false). Descriptors specifying `value` or `get` return
// `(clears, true)` where `clears` is false when `value` restores the real API.
func (analysis *RstestCallAnalysis) propertyDescriptorChange(descNode *ast.Node, bit expectConfigKind) (clears bool, ok bool) {
	desc := internalUtils.SkipAssertionsAndParens(descNode)
	if desc == nil || desc.Kind != ast.KindObjectLiteralExpression || desc.AsObjectLiteralExpression().Properties == nil {
		return false, false
	}
	hasValue := false
	valuePreserves := false
	hasGetter := false
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
		}
	}
	if hasValue {
		return !valuePreserves, true
	}
	if hasGetter {
		return true, true
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
			if src == nil || src.Kind == ast.KindSpreadElement {
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
				preserves := false
				switch prop.Kind {
				case ast.KindPropertyAssignment:
					preserves = analysis.overwritePreservesConfigBit(bit, prop.AsPropertyAssignment().Initializer)
				case ast.KindShorthandPropertyAssignment:
					preserves = analysis.overwritePreservesConfigBit(bit, prop.AsShorthandPropertyAssignment().Name())
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
	if binding.Initializer != nil {
		out |= analysis.expectConfigMember(binding.Initializer)
	}
	pattern := declaration.Parent
	if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return out
	}
	container := analysis.expectConfigBindingPatternKind(pattern)
	if container == expectConfigNone {
		return out
	}
	if binding.DotDotDotToken != nil {
		return out | (container & (expectConfigReceiver | expectConfigExpect))
	}
	member, known, staticKey := analysis.evalBindingElementPropertyName(binding)
	return out | classifyDestructuredExpectConfigProperty(container, member, known, staticKey)
}

func (analysis *RstestCallAnalysis) assignmentDestructuredExpectConfigMember(prop *ast.Node) expectConfigKind {
	if prop == nil || prop.Parent == nil || prop.Parent.Kind != ast.KindObjectLiteralExpression {
		return expectConfigNone
	}
	obj := prop.Parent
	container := analysis.expectConfigAssignmentObjectKind(obj)
	if container == expectConfigNone {
		return expectConfigNone
	}
	if prop.Kind == ast.KindSpreadAssignment {
		return container & (expectConfigReceiver | expectConfigExpect)
	}
	propName := prop.Name()
	if propName == nil {
		return expectConfigNone
	}
	member, known, staticKey := analysis.evalPropertyNameKey(propName)
	return classifyDestructuredExpectConfigProperty(container, member, known, staticKey)
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
			// `({ expect: { extend } = {} } = core)`: combine the default RHS
			// with the parent PropertyAssignment's container.
			if outer.Parent.Parent != nil && outer.Parent.Parent.Kind == ast.KindPropertyAssignment &&
				outer.Parent.Parent.AsPropertyAssignment().Initializer == outer.Parent {
				return analysis.assignmentDestructuredExpectConfigMember(outer.Parent.Parent) | analysis.expectConfigMember(bin.Right)
			}
			return analysis.expectConfigMember(bin.Right)
		}
	case ast.KindPropertyAssignment:
		return analysis.assignmentDestructuredExpectConfigMember(outer.Parent)
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
