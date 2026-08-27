package utils

import (
	"github.com/microsoft/typescript-go/shim/ast"
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
	calls       []*ast.Node
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
	callbacks     RstestTestCallbacks
	callbacksOK   bool
	hasTests      bool
}

type rstestCallAnalysisFileCacheKey struct{}

// GetRstestCallAnalysis returns the analysis shared by every Rstest rule that
// lints the same file. Manually-constructed contexts without a file cache keep
// the standalone behavior used by rule and parser tests.
func GetRstestCallAnalysis(ctx rule.RuleContext) *RstestCallAnalysis {
	analysisCtx := rule.RuleContext{
		SourceFile:  ctx.SourceFile,
		TypeChecker: ctx.TypeChecker,
	}
	return rule.CachedByFile(
		ctx,
		rstestCallAnalysisFileCacheKey{},
		func() *RstestCallAnalysis { return newRstestCallAnalysis(analysisCtx) },
	)
}

func newRstestCallAnalysis(ctx rule.RuleContext) *RstestCallAnalysis {
	analysis := &RstestCallAnalysis{
		ctx:           ctx,
		candidates:    cloneRstestCandidateSeeds(),
		fnCalls:       map[*ast.Node]*ParsedRstestFnCall{},
		expectCalls:   map[*ast.Node]*ParsedRstestExpectCall{},
		isExpect:      map[*ast.Node]bool{},
		expectRoots:   map[*ast.Symbol]rstestExpectRoot{},
		functions:     map[string]rstestFunctionEntry{},
		callbackInfos: map[*ast.Node]rstestCallbackInfo{},
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
	info := resolveRstestTestCallback(analysis.ctx, node.AsCallExpression())
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

func (analysis *RstestCallAnalysis) Callbacks() RstestTestCallbacks {
	return *analysis.callbacksRef()
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
)

const rstestCandidateAll = rstestCandidateFn | rstestCandidateTest | rstestCandidateExpect

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
		switch node.Kind {
		case ast.KindImportDeclaration:
			analysis.collectImportCandidates(node.AsImportDeclaration())
		case ast.KindVariableDeclaration:
			analysis.collectVariableCandidates(node, &aliases)
			analysis.recordFunction(node)
		case ast.KindFunctionDeclaration:
			analysis.recordFunction(node)
		case ast.KindCallExpression:
			analysis.calls = append(analysis.calls, node)
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
	initializer := ast.SkipParentheses(declaration.Initializer)
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
				var ok bool
				importedName, ok = staticRstestBindingPropertyName(binding.PropertyName)
				if !ok {
					continue
				}
			}
			analysis.candidates[binding.Name().Text()] |=
				rstestImportedCandidateKind(importedName)
		}
		return
	}
	if name.Kind == ast.KindObjectBindingPattern {
		if node.Parent == nil || node.Parent.Kind != ast.KindVariableDeclarationList ||
			node.Parent.Flags&ast.NodeFlagsConst == 0 {
			return
		}
		root := testFramework.ResolveFirstIdentifier(initializer)
		if root == nil || root.Kind != ast.KindIdentifier {
			return
		}
		pattern := name.AsBindingPattern()
		if pattern == nil || pattern.Elements == nil {
			return
		}
		for _, element := range pattern.Elements.Nodes {
			binding := element.AsBindingElement()
			if binding == nil || binding.DotDotDotToken != nil || binding.Name() == nil ||
				binding.Name().Kind != ast.KindIdentifier {
				continue
			}
			*aliases = append(*aliases, rstestAliasCandidate{
				localName: binding.Name().AsIdentifier().Text,
				rootName:  root.AsIdentifier().Text,
			})
		}
		return
	}
	if name.Kind != ast.KindIdentifier || node.Parent == nil ||
		node.Parent.Kind != ast.KindVariableDeclarationList ||
		node.Parent.Flags&ast.NodeFlagsConst == 0 {
		return
	}
	localName := name.AsIdentifier().Text
	if isRstestRequireCall(initializer) || isImportMetaRstest(initializer) {
		analysis.candidates[localName] |= rstestCandidateAll
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
	return testFramework.IsModuleRequireCall(node, RstestImportModule) ||
		testFramework.IsModuleRequireCall(node, RstestPlaywrightImportModule)
}

func isRstestImportModule(name string) bool {
	return name == RstestImportModule || name == RstestPlaywrightImportModule
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
	return kind
}
