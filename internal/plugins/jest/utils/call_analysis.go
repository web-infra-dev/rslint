package utils

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
)

type jestCallAnalysisFileCacheKey struct{}

type jestCallParseResult struct {
	parsed *ParsedJestFnCall
	reason string
}

// JestCallAnalysis owns reusable Jest call, parse-reason, and callback results
// for one file. It intentionally models only Jest's existing semantics; the
// larger Rstest analysis has framework-specific provenance and execution-mode
// responsibilities that do not belong here.
type JestCallAnalysis struct {
	ctx                     rule.RuleContext
	fnCalls                 map[*ast.Node]jestCallParseResult
	calls                   []*ast.Node
	functions               map[string]*ast.Node
	indexed                 bool
	callbackInfos           map[*ast.Node]jestCallbackInfo
	callbacks               JestTestCallbacks
	callbacksOK             bool
	registrationCallbacks   map[*ast.Node]bool
	registrationCallbacksOK bool
}

// GetJestCallAnalysis returns the analysis shared by every migrated Jest rule
// that lints the same file. Manually constructed contexts without a file cache
// retain standalone behavior for parser and rule tests.
func GetJestCallAnalysis(ctx rule.RuleContext) *JestCallAnalysis {
	// Settings affects settings.jest.globalAliases. It is file-level linter
	// configuration, unlike rule options, and is shared by every rule in the
	// pass. Keep only fields the analysis reads; in particular, never retain the
	// rule-specific reporter attached to ctx.
	analysisCtx := rule.RuleContext{
		SourceFile:  ctx.SourceFile,
		Settings:    ctx.Settings,
		TypeChecker: ctx.TypeChecker,
		Refs:        ctx.Refs,
	}
	return rule.CachedByFile(
		ctx,
		jestCallAnalysisFileCacheKey{},
		func() *JestCallAnalysis {
			return &JestCallAnalysis{
				ctx:           analysisCtx,
				fnCalls:       map[*ast.Node]jestCallParseResult{},
				callbackInfos: map[*ast.Node]jestCallbackInfo{},
			}
		},
	)
}

// ParseFnCall parses any Jest API call once. The result map stores misses and
// broken-expect reasons as well as successful parses because ordinary calls
// dominate real source files and are asked about by many recommended rules.
func (analysis *JestCallAnalysis) ParseFnCall(node *ast.Node) *ParsedJestFnCall {
	return analysis.parseFnCallResult(node).parsed
}

// ParseTestCall is the test-only entry point used by callback consumers.
func (analysis *JestCallAnalysis) ParseTestCall(node *ast.Node) *ParsedJestFnCall {
	parsed := analysis.ParseFnCall(node)
	if parsed == nil || parsed.Kind != JestFnTypeTest {
		return nil
	}
	return parsed
}

// ParseExpectCall returns a successfully parsed Jest expect call.
func (analysis *JestCallAnalysis) ParseExpectCall(node *ast.Node) *ParsedJestFnCall {
	parsed := analysis.ParseFnCall(node)
	if parsed == nil || parsed.Kind != JestFnTypeExpect {
		return nil
	}
	return parsed
}

// ParseExpectCallWithReason additionally returns why a Jest expect chain was
// invalid. valid-expect consumes the same cached parse as the other rules
// instead of rebuilding its member chain and resolving its root a second time.
func (analysis *JestCallAnalysis) ParseExpectCallWithReason(
	node *ast.Node,
) (*ParsedJestFnCall, string) {
	result := analysis.parseFnCallResult(node)
	if result.parsed == nil || result.parsed.Kind != JestFnTypeExpect {
		return nil, result.reason
	}
	return result.parsed, ExpectParseReasonNone
}

func (analysis *JestCallAnalysis) parseFnCallResult(node *ast.Node) jestCallParseResult {
	if node == nil || node.Kind != ast.KindCallExpression {
		return jestCallParseResult{}
	}
	if result, ok := analysis.fnCalls[node]; ok {
		return result
	}
	result := parseJestFnCallWithReason(node, analysis.ctx)
	analysis.fnCalls[node] = result
	return result
}

// TestCallback returns the named callback associated with a final Jest test
// registration. Inline callbacks deliberately return the zero value,
// preserving expect-expect's existing contract.
func (analysis *JestCallAnalysis) TestCallback(node *ast.Node) (*ast.Node, string) {
	if analysis.ParseTestCall(node) == nil {
		return nil, ""
	}
	info := analysis.testCallbackInfo(node)
	if info.name == "" {
		return nil, ""
	}
	return info.functionNode, info.name
}

// Callbacks returns the callback classification shared by callback-aware Jest
// rules. It is built lazily because most individual Jest rules do not need it.
func (analysis *JestCallAnalysis) Callbacks() JestTestCallbacks {
	if !analysis.callbacksOK {
		analysis.callbacks = collectJestTestCallbacks(analysis)
		analysis.callbacksOK = true
	}
	return analysis.callbacks
}

// RegistrationCallbacks returns every function that a test or describe
// registration invokes as its callback. It includes inline and named
// callbacks, and reuses the analysis's cached call and function indexes.
func (analysis *JestCallAnalysis) RegistrationCallbacks() map[*ast.Node]bool {
	if analysis.registrationCallbacksOK {
		return analysis.registrationCallbacks
	}
	callbacks := map[*ast.Node]bool{}
	analysis.indexSourceFile()
	for _, node := range analysis.calls {
		parsed := analysis.ParseFnCall(node)
		if parsed == nil || (parsed.Kind != JestFnTypeTest && parsed.Kind != JestFnTypeDescribe) {
			continue
		}
		info := analysis.testCallbackInfo(node)
		if info.functionNode == nil {
			info.functionNode = analysis.fallbackCallbackFunction(info.name)
		}
		if info.functionNode != nil {
			callbacks[info.functionNode] = true
		}
	}
	analysis.registrationCallbacks = callbacks
	analysis.registrationCallbacksOK = true
	return analysis.registrationCallbacks
}

func (analysis *JestCallAnalysis) testCallbackInfo(node *ast.Node) jestCallbackInfo {
	if info, ok := analysis.callbackInfos[node]; ok {
		return info
	}
	info := resolveJestTestCallback(analysis.ctx, node.AsCallExpression())
	analysis.callbackInfos[node] = info
	return info
}

func (analysis *JestCallAnalysis) fallbackCallbackFunction(name string) *ast.Node {
	if name == "" {
		return nil
	}
	analysis.indexSourceFile()
	return analysis.functions[name]
}

// indexSourceFile collects the two facts callback discovery needs in one walk:
// every call and the first function declaration/initializer for each name. The
// latter preserves the old fallback's source-order behavior when no checker
// declaration is available.
func (analysis *JestCallAnalysis) indexSourceFile() {
	if analysis.indexed {
		return
	}
	analysis.indexed = true
	analysis.functions = map[string]*ast.Node{}
	if analysis.ctx.SourceFile == nil {
		return
	}

	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		switch node.Kind {
		case ast.KindCallExpression:
			analysis.calls = append(analysis.calls, node)
		case ast.KindFunctionDeclaration:
			declaration := node.AsFunctionDeclaration()
			if declaration != nil && declaration.Name() != nil {
				analysis.recordFunction(declaration.Name().Text(), node)
			}
		case ast.KindVariableDeclaration:
			declaration := node.AsVariableDeclaration()
			if declaration != nil && declaration.Name() != nil &&
				declaration.Name().Kind == ast.KindIdentifier &&
				declaration.Initializer != nil {
				initializer := ast.SkipParentheses(declaration.Initializer)
				if ast.IsFunctionExpressionOrArrowFunction(initializer) {
					analysis.recordFunction(declaration.Name().Text(), initializer)
				}
			}
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(analysis.ctx.SourceFile.AsNode())
}

func (analysis *JestCallAnalysis) recordFunction(name string, function *ast.Node) {
	if name == "" || function == nil || analysis.functions[name] != nil {
		return
	}
	analysis.functions[name] = function
}
