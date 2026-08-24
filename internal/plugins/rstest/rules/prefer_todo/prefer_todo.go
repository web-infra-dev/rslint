package prefer_todo

import (
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

func messageEmptyTest() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "emptyTest",
		Description: "Prefer todo test case over empty test case",
	}
}

func messageUnimplementedTest() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unimplementedTest",
		Description: "Prefer todo test case over unimplemented test case",
	}
}

type callbackKind uint8

const (
	callbackImplementedOrUnknown callbackKind = iota
	callbackMissing
	callbackInlineEmpty
)

type callbackClassification struct {
	kind          callbackKind
	callbackIndex int
	optionsIndex  int
	timeoutIndex  int
	// reportOnly is true when the callback is certainly empty but the call has
	// no supported Rstest overload that can be rewritten safely.
	reportOnly bool
}

// PreferTodoRule ports eslint-plugin-jest's prefer-todo semantics to Rstest's
// wider registration surface. Unlike the Jest rule, it must classify Rstest's
// `(title, fn, timeout)` and `(title, options, fn?)` overloads conservatively.
//
// Reporting and fixing are two separate decisions. A registration the parser
// has proved to be a bare Rstest test is reported even when no safe rewrite
// exists — a same-file alias that hides `.skip`, or a repeated `.skip` chain,
// both report without a fix. A call site the parser cannot prove is still an
// Rstest test at run time, such as a whole-module CommonJS namespace whose
// properties are writable, is not reported at all.
var PreferTodoRule = rule.Rule{
	Name:   "rstest/prefer-todo",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		analysis := rstestUtils.GetRstestCallAnalysis(ctx)
		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				parsed := analysis.ParseFnCall(node)
				if parsed == nil || parsed.Kind != rstestUtils.RstestFnTypeTest {
					return
				}
				if parsed.Todo || parsed.IsParameterized() {
					return
				}

				call := node.AsCallExpression()
				if call == nil || call.Arguments == nil {
					return
				}
				optionalCall := call.QuestionDotToken != nil ||
					ast.IsOptionalChain(node) ||
					ast.IsOptionalChain(ast.SkipParentheses(call.Expression))
				args := call.Arguments.Nodes
				if len(args) == 0 {
					return
				}

				title := ast.SkipParentheses(args[0])
				if !isStringNode(title) {
					return
				}

				callSite, ok := classifyCallSite(ctx, call, parsed)
				if !ok {
					return
				}

				classification := classifyCallback(args)
				switch classification.kind {
				case callbackMissing:
					ctx.ReportNodeWithDeferredFixes(node, messageUnimplementedTest(), func() []rule.RuleFix {
						if optionalCall || classification.reportOnly {
							return nil
						}
						return buildFixes(ctx, call, callSite, classification)
					})
				case callbackInlineEmpty:
					ctx.ReportNodeWithDeferredFixes(node, messageEmptyTest(), func() []rule.RuleFix {
						if optionalCall || classification.reportOnly {
							return nil
						}
						return buildFixes(ctx, call, callSite, classification)
					})
				}
			},
		}
	},
}

type callSiteClassification struct {
	bareRoot  bool
	skipEntry *rstestUtils.ParsedRstestFnMemberEntry
	// fixable is false for a call site that is reported but has no rewrite
	// that delivers a todo test.
	fixable bool
}

func classifyCallSite(
	ctx rule.RuleContext,
	call *ast.CallExpression,
	parsed *rstestUtils.ParsedRstestFnCall,
) (callSiteClassification, bool) {
	if parsed == nil {
		return callSiteClassification{}, false
	}

	if len(parsed.Members) == 0 {
		if isSafeBareCallSiteReference(ctx, call.Expression) {
			return callSiteClassification{bareRoot: true, fixable: true}, true
		}
		if parsed.Skipped {
			// `const skipped = test.skip; skipped('case')`. The parser resolved
			// the alias, so this is a skipped Rstest test with no body, but the
			// `.skip` lives in the alias initializer and there is no accessor at
			// the call site to rewrite. Report it without a fix.
			return callSiteClassification{}, true
		}
		return callSiteClassification{}, false
	}

	var skipEntry *rstestUtils.ParsedRstestFnMemberEntry
	repeatedSkip := false
	for i, member := range parsed.Members {
		if member != "skip" {
			return callSiteClassification{}, false
		}
		if skipEntry != nil {
			repeatedSkip = true
			continue
		}
		skipEntry = &parsed.MemberEntries[i]
	}
	if skipEntry == nil {
		return callSiteClassification{}, false
	}
	receiver, _ := testFramework.AccessorReceiverAndParent(skipEntry)
	if !isSafeBareCallSiteReference(ctx, receiver) {
		return callSiteClassification{}, false
	}
	if repeatedSkip {
		// Rewriting either accessor in `test.skip.skip()` leaves the other skip
		// active, so no one-accessor fix produces a todo. Report without one.
		return callSiteClassification{}, true
	}
	return callSiteClassification{skipEntry: skipEntry, fixable: true}, true
}

func isSafeBareCallSiteReference(ctx rule.RuleContext, node *ast.Node) bool {
	if node == nil {
		return false
	}
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	switch node.Kind {
	case ast.KindIdentifier:
		return isSafeBareIdentifierReference(ctx, node, nil, 0)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return isSafeBarePropertyReference(ctx, node, nil, 0)
	default:
		return false
	}
}

func isSafeBarePropertyReference(
	ctx rule.RuleContext,
	node *ast.Node,
	visited map[*ast.Symbol]bool,
	depth int,
) bool {
	node = ast.SkipParentheses(node)
	if node == nil || depth > 16 {
		return false
	}

	var (
		receiver *ast.Node
		name     string
	)
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		if property == nil || property.Name() == nil {
			return false
		}
		receiver = property.Expression
		name = property.Name().Text()
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		if element == nil {
			return false
		}
		receiver = element.Expression
		var ok bool
		name, ok = internalUtils.GetStaticStringLiteralValue(
			ast.SkipParentheses(element.ArgumentExpression),
		)
		if !ok {
			return false
		}
	default:
		return false
	}

	if name != "test" && name != "it" {
		return false
	}
	return isSafeRstestNamespaceReceiver(ctx, receiver, visited, depth)
}

func isSafeRstestNamespaceReceiver(
	ctx rule.RuleContext,
	node *ast.Node,
	visited map[*ast.Symbol]bool,
	depth int,
) bool {
	node = ast.SkipParentheses(node)
	if node == nil || depth > 16 {
		return false
	}
	if isImportMetaRstest(node) {
		return true
	}
	if node.Kind != ast.KindIdentifier || ctx.Refs == nil {
		return false
	}

	symbol := ctx.Refs.Resolve(node)
	if symbol == nil {
		return false
	}
	if visited != nil && visited[symbol] {
		return false
	}
	if isRstestNamespaceImportSymbol(symbol) {
		return true
	}

	if visited == nil {
		visited = map[*ast.Symbol]bool{}
	}
	visited[symbol] = true
	defer delete(visited, symbol)

	sameFileDecls := 0
	for _, declaration := range symbol.Declarations {
		if declaration == nil || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
			continue
		}
		sameFileDecls++
		if !isSafeRstestNamespaceAliasDeclaration(ctx, declaration, visited, depth+1) {
			return false
		}
	}
	return sameFileDecls > 0
}

// isRstestNamespaceImportSymbol deliberately accepts only ESM namespace
// imports. IsModuleNamespaceSymbol also accepts whole-module require bindings,
// but those objects have mutable properties: `const core = require(...)` does
// not prevent `core.test = custom`. Without proving the property is unwritten,
// a fix on `core.test()` could therefore rewrite an unrelated function call.
func isRstestNamespaceImportSymbol(symbol *ast.Symbol) bool {
	if symbol == nil {
		return false
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil || declaration.Kind != ast.KindNamespaceImport {
			continue
		}
		importDeclaration := testFramework.FindImportDeclaration(declaration)
		if importDeclaration == nil || importDeclaration.ModuleSpecifier == nil {
			continue
		}
		switch importDeclaration.ModuleSpecifier.Text() {
		case rstestUtils.RstestImportModule, rstestUtils.RstestPlaywrightImportModule:
			return true
		}
	}
	return false
}

func isSafeRstestNamespaceAliasDeclaration(
	ctx rule.RuleContext,
	declaration *ast.Node,
	visited map[*ast.Symbol]bool,
	depth int,
) bool {
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	if declaration.Parent == nil ||
		declaration.Parent.Kind != ast.KindVariableDeclarationList ||
		declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}

	initializer := ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer)
	if initializer == nil {
		return false
	}
	if isImportMetaRstest(initializer) {
		return true
	}

	switch initializer.Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return isSafeRstestNamespaceReceiver(ctx, initializer, visited, depth)
	default:
		return false
	}
}

func isSafeBareIdentifierReference(
	ctx rule.RuleContext,
	identifier *ast.Node,
	visited map[*ast.Symbol]bool,
	depth int,
) bool {
	if identifier == nil || identifier.Kind != ast.KindIdentifier || depth > 16 {
		return false
	}
	if ctx.Refs == nil {
		text := identifier.AsIdentifier().Text
		return text == "test" || text == "it"
	}

	symbol := ctx.Refs.Resolve(identifier)
	if symbol == nil {
		text := identifier.AsIdentifier().Text
		return text == "test" || text == "it"
	}
	if visited != nil && visited[symbol] {
		return false
	}

	name, _, mode := testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
		identifier.AsIdentifier().Text,
		identifier,
		symbol,
		ctx.SourceFile,
		rstestUtils.RstestImportModule,
	)
	if mode == rstestUtils.RSTEST_IMPORT_MODE && (name == "test" || name == "it") {
		return true
	}
	if mode == rstestUtils.RSTEST_GLOBAL_MODE &&
		(name == "test" || name == "it") &&
		!internalUtils.IsValueSymbolDeclaredInFile(symbol, ctx.SourceFile) {
		return true
	}
	name, _, mode = testFramework.ResolveFunctionIdentifierReferenceFromSymbol(
		identifier.AsIdentifier().Text,
		identifier,
		symbol,
		ctx.SourceFile,
		rstestUtils.RstestPlaywrightImportModule,
	)
	if mode == rstestUtils.RSTEST_IMPORT_MODE && name == "test" {
		return true
	}

	if visited == nil {
		visited = map[*ast.Symbol]bool{}
	}
	visited[symbol] = true
	defer delete(visited, symbol)

	sameFileDecls := 0
	for _, declaration := range symbol.Declarations {
		if declaration == nil || ast.GetSourceFileOfNode(declaration) != ctx.SourceFile {
			continue
		}
		sameFileDecls++
		if !isSafeBareAliasDeclaration(ctx, declaration, visited, depth+1) {
			return false
		}
	}
	return sameFileDecls > 0
}

func isSafeBareAliasDeclaration(
	ctx rule.RuleContext,
	declaration *ast.Node,
	visited map[*ast.Symbol]bool,
	depth int,
) bool {
	if declaration != nil && declaration.Kind == ast.KindBindingElement {
		return isSafeBareBindingElement(ctx, declaration, visited, depth)
	}
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	if declaration.Parent == nil ||
		declaration.Parent.Kind != ast.KindVariableDeclarationList ||
		declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}

	initializer := ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer)
	if initializer == nil {
		return false
	}
	switch initializer.Kind {
	case ast.KindIdentifier:
		return isSafeBareIdentifierReference(ctx, initializer, visited, depth)
	case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return isSafeBarePropertyReference(ctx, initializer, visited, depth)
	default:
		return false
	}
}

// isSafeBareBindingElement accepts a `test` or `it` pulled straight out of an
// Rstest namespace by a `const` object destructure, such as
// `const { test } = import.meta.rstest`. The property being read has to be the
// bare test API itself, so a rename keeps the original property name as the
// thing that must be `test` or `it`. A default value, a rest element, an array
// pattern, or a nested pattern all break that guarantee and are refused.
func isSafeBareBindingElement(
	ctx rule.RuleContext,
	declaration *ast.Node,
	visited map[*ast.Symbol]bool,
	depth int,
) bool {
	if declaration == nil || declaration.Kind != ast.KindBindingElement || depth > 16 {
		return false
	}
	binding := declaration.AsBindingElement()
	if binding == nil || binding.DotDotDotToken != nil || binding.Initializer != nil {
		return false
	}
	if binding.Name() == nil || binding.Name().Kind != ast.KindIdentifier {
		return false
	}

	name := ""
	if binding.PropertyName != nil {
		var ok bool
		name, ok = internalUtils.GetStaticStringLiteralValue(binding.PropertyName)
		if !ok {
			if binding.PropertyName.Kind != ast.KindIdentifier {
				return false
			}
			name = binding.PropertyName.AsIdentifier().Text
		}
	} else {
		name = binding.Name().AsIdentifier().Text
	}
	if name != "test" && name != "it" {
		return false
	}

	// Only a direct property of the namespace qualifies. Walking up through
	// nested patterns would accept `const { runner: { test } } = ns`, where the
	// binding names a property of something else entirely.
	pattern := declaration.Parent
	if pattern == nil || pattern.Kind != ast.KindObjectBindingPattern {
		return false
	}
	variable := pattern.Parent
	if variable == nil || variable.Kind != ast.KindVariableDeclaration {
		return false
	}
	if variable.Parent == nil ||
		variable.Parent.Kind != ast.KindVariableDeclarationList ||
		variable.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}

	initializer := ast.SkipParentheses(variable.AsVariableDeclaration().Initializer)
	if initializer == nil {
		return false
	}
	return isSafeRstestNamespaceReceiver(ctx, initializer, visited, depth+1)
}

func isImportMetaRstest(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil {
		return false
	}

	var (
		expression *ast.Node
		name       string
		ok         bool
	)
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		property := node.AsPropertyAccessExpression()
		if property == nil || property.Name() == nil {
			return false
		}
		expression = property.Expression
		name = property.Name().Text()
	case ast.KindElementAccessExpression:
		element := node.AsElementAccessExpression()
		if element == nil {
			return false
		}
		expression = element.Expression
		name, ok = internalUtils.GetStaticStringLiteralValue(
			ast.SkipParentheses(element.ArgumentExpression),
		)
		if !ok {
			return false
		}
	default:
		return false
	}
	return name == "rstest" && isImportMeta(expression)
}

// isImportMeta matches `import.meta` itself. Anything with `import.meta` merely
// at the root of a longer chain, such as `import.meta.env.rstest` or
// `import.meta.glob('x').rstest`, names some other object and is not the Rstest
// namespace.
func isImportMeta(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node == nil || node.Kind != ast.KindMetaProperty {
		return false
	}
	meta := node.AsMetaProperty()
	return meta != nil && meta.Name() != nil && meta.Name().Text() == "meta"
}

func classifyCallback(args []*ast.Node) callbackClassification {
	if len(args) == 1 {
		return callbackClassification{
			kind:          callbackMissing,
			callbackIndex: -1,
			optionsIndex:  -1,
			timeoutIndex:  -1,
		}
	}
	if len(args) == 0 {
		return callbackClassification{
			kind:          callbackImplementedOrUnknown,
			callbackIndex: -1,
			optionsIndex:  -1,
			timeoutIndex:  -1,
		}
	}
	if len(args) > 3 {
		classification := classifyCallback(args[:3])
		if classification.kind == callbackInlineEmpty {
			classification.reportOnly = true
		}
		return classification
	}

	second := skipParentheses(args[1])
	var third *ast.Node
	if len(args) >= 3 {
		third = skipParentheses(args[2])
	}

	secondInlineFn, secondEmpty := inlineFunctionState(second)
	thirdInlineFn, thirdEmpty := inlineFunctionState(third)

	if secondInlineFn {
		if secondEmpty {
			if thirdInlineFn {
				return callbackClassification{
					kind:          callbackInlineEmpty,
					callbackIndex: 1,
					optionsIndex:  -1,
					timeoutIndex:  -1,
					reportOnly:    true,
				}
			}
			timeoutIndex := -1
			if len(args) == 3 {
				timeoutIndex = 2
			}
			return callbackClassification{
				kind:          callbackInlineEmpty,
				callbackIndex: 1,
				optionsIndex:  -1,
				timeoutIndex:  timeoutIndex,
			}
		}
		return callbackClassification{kind: callbackImplementedOrUnknown, callbackIndex: -1, optionsIndex: -1, timeoutIndex: -1}
	}

	if isObjectLiteral(second) {
		if len(args) == 2 {
			return callbackClassification{
				kind:          callbackMissing,
				callbackIndex: -1,
				optionsIndex:  1,
				timeoutIndex:  -1,
			}
		}
		if thirdInlineFn && thirdEmpty {
			return callbackClassification{
				kind:          callbackInlineEmpty,
				callbackIndex: 2,
				optionsIndex:  1,
				timeoutIndex:  -1,
			}
		}
		return callbackClassification{kind: callbackImplementedOrUnknown, callbackIndex: -1, optionsIndex: -1, timeoutIndex: -1}
	}

	return callbackClassification{kind: callbackImplementedOrUnknown, callbackIndex: -1, optionsIndex: -1, timeoutIndex: -1}
}

func skipParentheses(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	return ast.SkipParentheses(node)
}

func inlineFunctionState(node *ast.Node) (bool, bool) {
	if node == nil {
		return false, false
	}
	switch node.Kind {
	case ast.KindArrowFunction:
		fn := node.AsArrowFunction()
		if fn == nil || fn.Body == nil || fn.Body.Kind != ast.KindBlock {
			return true, false
		}
		block := fn.Body.AsBlock()
		return true, block != nil && (block.Statements == nil || len(block.Statements.Nodes) == 0)
	case ast.KindFunctionExpression:
		fn := node.AsFunctionExpression()
		if fn == nil || fn.Body == nil || fn.Body.Kind != ast.KindBlock {
			return true, false
		}
		block := fn.Body.AsBlock()
		return true, block != nil && (block.Statements == nil || len(block.Statements.Nodes) == 0)
	default:
		return false, false
	}
}

func isObjectLiteral(node *ast.Node) bool {
	return node != nil && node.Kind == ast.KindObjectLiteralExpression
}

func isStringNode(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression:
		return true
	default:
		return false
	}
}

func buildFixes(
	ctx rule.RuleContext,
	call *ast.CallExpression,
	callSite callSiteClassification,
	classification callbackClassification,
) []rule.RuleFix {
	if ctx.SourceFile == nil || call == nil || !callSite.fixable {
		return nil
	}

	fixes := make([]rule.RuleFix, 0, 2)
	calleeFix, ok := buildCalleeFix(ctx.SourceFile, call, callSite)
	if !ok {
		return nil
	}
	fixes = append(fixes, calleeFix)

	if classification.kind == callbackInlineEmpty {
		argFixes, ok := buildArgumentFixes(ctx.SourceFile, call, classification)
		if !ok {
			return nil
		}
		fixes = append(fixes, argFixes...)
	}

	if len(fixes) == 0 {
		return nil
	}
	return fixes
}

func buildCalleeFix(
	sourceFile *ast.SourceFile,
	call *ast.CallExpression,
	callSite callSiteClassification,
) (rule.RuleFix, bool) {
	if callSite.bareRoot {
		callee := internalUtils.OutermostParenthesizedExpression(call.Expression)
		return rule.RuleFixInsertAfter(callee, ".todo"), true
	}
	if callSite.skipEntry == nil {
		return rule.RuleFix{}, false
	}
	valueRange, replacement, ok := testFramework.AccessorReplacement(
		sourceFile,
		callSite.skipEntry.Node,
		"todo",
	)
	if !ok {
		return rule.RuleFix{}, false
	}
	return rule.RuleFixReplaceRange(valueRange, replacement), true
}

func buildArgumentFixes(
	sourceFile *ast.SourceFile,
	call *ast.CallExpression,
	classification callbackClassification,
) ([]rule.RuleFix, bool) {
	args := call.Arguments
	if args == nil || len(args.Nodes) == 0 || classification.callbackIndex < 0 {
		return nil, false
	}
	nodes := args.Nodes
	if classification.callbackIndex >= len(nodes) {
		return nil, false
	}

	text := sourceFile.Text()
	callback := internalUtils.OutermostParenthesizedExpression(nodes[classification.callbackIndex])
	if callback == nil {
		return nil, false
	}
	callbackRange := internalUtils.TrimNodeTextRange(sourceFile, callback)

	switch {
	case classification.callbackIndex == 2 && classification.optionsIndex == 1:
		startRange, ok := findCommaBefore(sourceFile, call.AsNode(), callbackRange.Pos())
		if !ok {
			return nil, false
		}
		start := startRange.Pos()
		end := callbackRange.End()
		if commaAfter, ok := findCommaAfter(sourceFile, call.AsNode(), callbackRange.End()); ok {
			end = commaAfter.End()
		}
		if start >= end || end > len(text) {
			return nil, false
		}
		return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, end))}, true
	case classification.callbackIndex == 1 && classification.optionsIndex == -1:
		startRange, ok := findCommaBefore(sourceFile, call.AsNode(), callbackRange.Pos())
		if !ok {
			return nil, false
		}
		start := startRange.Pos()
		end := callbackRange.End()
		if classification.timeoutIndex >= 0 {
			if classification.timeoutIndex >= len(nodes) {
				return nil, false
			}
			timeout := internalUtils.OutermostParenthesizedExpression(nodes[classification.timeoutIndex])
			if timeout == nil {
				return nil, false
			}
			timeoutRange := internalUtils.TrimNodeTextRange(sourceFile, timeout)
			end = timeoutRange.End()
			if commaAfter, ok := findCommaAfter(sourceFile, call.AsNode(), timeoutRange.End()); ok {
				end = commaAfter.End()
			}
		} else if commaAfter, ok := findCommaAfter(sourceFile, call.AsNode(), callbackRange.End()); ok {
			end = commaAfter.End()
		}
		if start >= end || end > len(text) {
			return nil, false
		}
		return []rule.RuleFix{rule.RuleFixRemoveRange(core.NewTextRange(start, end))}, true
	default:
		return nil, false
	}
}

func findCommaBefore(sourceFile *ast.SourceFile, node *ast.Node, pos int) (core.TextRange, bool) {
	if sourceFile == nil || node == nil {
		return core.TextRange{}, false
	}
	var lastComma core.TextRange
	found := false
	for _, token := range internalUtils.TokensOfNode(sourceFile, node) {
		if token.Kind == ast.KindCommaToken && token.End <= pos {
			lastComma = token.Range()
			found = true
		}
	}
	return lastComma, found
}

func findCommaAfter(sourceFile *ast.SourceFile, node *ast.Node, pos int) (core.TextRange, bool) {
	if sourceFile == nil || node == nil {
		return core.TextRange{}, false
	}
	for _, token := range internalUtils.TokensOfNode(sourceFile, node) {
		if token.Kind == ast.KindCommaToken && token.Start >= pos {
			return token.Range(), true
		}
	}
	return core.TextRange{}, false
}
