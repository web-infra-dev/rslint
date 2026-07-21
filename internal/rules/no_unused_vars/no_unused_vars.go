package no_unused_vars

import (
	_ "embed"
	"fmt"
	"sort"

	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no-unused-vars.schema.json
var schemaJSON []byte

type ruleFlavor struct {
	typescript      bool
	coreSuggestions bool
}

type EnableAutofixRemoval struct {
	Imports bool `json:"imports"`
}

type Config struct {
	Vars                           string               `json:"vars"`
	VarsIgnorePattern              string               `json:"varsIgnorePattern"`
	Args                           string               `json:"args"`
	ArgsIgnorePattern              string               `json:"argsIgnorePattern"`
	CaughtErrors                   string               `json:"caughtErrors"`
	CaughtErrorsIgnorePattern      string               `json:"caughtErrorsIgnorePattern"`
	DestructuredArrayIgnorePattern string               `json:"destructuredArrayIgnorePattern"`
	IgnoreRestSiblings             bool                 `json:"ignoreRestSiblings"`
	IgnoreClassWithStaticInitBlock bool                 `json:"ignoreClassWithStaticInitBlock"`
	IgnoreUsingDeclarations        bool                 `json:"ignoreUsingDeclarations"`
	ReportUsedIgnorePattern        bool                 `json:"reportUsedIgnorePattern"`
	EnableAutofixRemoval           EnableAutofixRemoval `json:"enableAutofixRemoval"`

	varsIgnoreRe              *regexp2.Regexp
	argsIgnoreRe              *regexp2.Regexp
	caughtErrorsIgnoreRe      *regexp2.Regexp
	destructuredArrayIgnoreRe *regexp2.Regexp
}

type variableType string

const (
	variableTypeArrayDestructure variableType = "array-destructure"
	variableTypeCatchClause      variableType = "catch-clause"
	variableTypeParameter        variableType = "parameter"
	variableTypeVariable         variableType = "variable"
)

type analysisContext struct {
	allUsages          map[*ast.Symbol][]*ast.Node
	writeRefs          map[*ast.Symbol][]*ast.Node
	localExportTargets map[*ast.Symbol]bool
	globalRefsByName   map[string][]*ast.Node
	exportedNames      map[string]bool
	refIndex           *utils.ReferenceIndex
	seenMergedSymbols  map[*ast.Symbol]bool
	reportedUnused     map[*ast.Node]bool
	reporter           *diagnosticReporter
	tokens             []utils.SourceToken
}

type pendingDiagnostic struct {
	node        *ast.Node
	textRange   core.TextRange
	usesRange   bool
	message     rule.RuleMessage
	suggestions []rule.RuleSuggestion
}

type diagnosticReporter struct {
	ctx          rule.RuleContext
	deferReports bool
	pending      []pendingDiagnostic
}

func (r *diagnosticReporter) reportNode(node *ast.Node, message rule.RuleMessage) {
	if !r.deferReports {
		r.ctx.ReportNode(node, message)
		return
	}
	r.pending = append(r.pending, pendingDiagnostic{node: node, message: message})
}

func (r *diagnosticReporter) reportNodeWithSuggestions(node *ast.Node, message rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
	if !r.deferReports {
		r.ctx.ReportNodeWithSuggestions(node, message, suggestions...)
		return
	}
	r.pending = append(r.pending, pendingDiagnostic{
		node:        node,
		message:     message,
		suggestions: suggestions,
	})
}

func (r *diagnosticReporter) reportRange(textRange core.TextRange, message rule.RuleMessage) {
	if !r.deferReports {
		r.ctx.ReportRange(textRange, message)
		return
	}
	r.pending = append(r.pending, pendingDiagnostic{
		textRange: textRange,
		usesRange: true,
		message:   message,
	})
}

func (r *diagnosticReporter) reportRangeWithSuggestions(textRange core.TextRange, message rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
	if !r.deferReports {
		r.ctx.ReportRangeWithSuggestions(textRange, message, suggestions...)
		return
	}
	r.pending = append(r.pending, pendingDiagnostic{
		textRange:   textRange,
		usesRange:   true,
		message:     message,
		suggestions: suggestions,
	})
}

func (r *diagnosticReporter) flush() {
	sort.SliceStable(r.pending, func(i, j int) bool {
		return r.position(r.pending[i]) < r.position(r.pending[j])
	})
	for _, diagnostic := range r.pending {
		if diagnostic.usesRange {
			if len(diagnostic.suggestions) > 0 {
				r.ctx.ReportRangeWithSuggestions(diagnostic.textRange, diagnostic.message, diagnostic.suggestions...)
			} else {
				r.ctx.ReportRange(diagnostic.textRange, diagnostic.message)
			}
		} else if len(diagnostic.suggestions) > 0 {
			r.ctx.ReportNodeWithSuggestions(diagnostic.node, diagnostic.message, diagnostic.suggestions...)
		} else {
			r.ctx.ReportNode(diagnostic.node, diagnostic.message)
		}
	}
	r.pending = nil
}

func (r *diagnosticReporter) position(diagnostic pendingDiagnostic) int {
	if diagnostic.usesRange {
		return diagnostic.textRange.Pos()
	}
	return utils.TrimNodeTextRange(r.ctx.SourceFile, diagnostic.node).Pos()
}

// eslintCoreDefinitionNameRange reproduces the range carried by a definition
// identifier from @typescript-eslint/parser. For a non-rest parameter or a
// simple variable declaration, that range includes its optional/type suffix;
// binding-pattern names and rest parameters keep the bare identifier range.
func eslintCoreDefinitionNameRange(sourceFile *ast.SourceFile, nameNode *ast.Node, definition *ast.Node) (core.TextRange, bool) {
	if nameNode == nil {
		return core.TextRange{}, false
	}
	nameRange := utils.TrimNodeTextRange(sourceFile, nameNode)
	if definition == nil || definition.Name() != nameNode ||
		(!ast.IsIdentifier(nameNode) && nameNode.Kind != ast.KindThisKeyword) {
		return nameRange, false
	}

	end := nameRange.End()
	switch definition.Kind {
	case ast.KindParameter:
		parameter := definition.AsParameterDeclaration()
		if parameter == nil || parameter.DotDotDotToken != nil {
			return nameRange, false
		}
		if parameter.Type != nil {
			end = parameter.Type.End()
		} else if parameter.QuestionToken != nil {
			end = parameter.QuestionToken.End()
		}
	case ast.KindVariableDeclaration:
		declaration := definition.AsVariableDeclaration()
		if declaration == nil {
			return nameRange, false
		}
		if declaration.Type != nil {
			end = declaration.Type.End()
		} else if declaration.ExclamationToken != nil {
			end = declaration.ExclamationToken.End()
		}
	default:
		return nameRange, false
	}

	if end <= nameRange.End() {
		return nameRange, false
	}
	return core.NewTextRange(nameRange.Pos(), end), true
}

func reportVariableDiagnostic(ctx rule.RuleContext, reporter *diagnosticReporter, nameNode *ast.Node, reportNode *ast.Node, definition *ast.Node, flavor ruleFlavor, message rule.RuleMessage, suggestions ...rule.RuleSuggestion) {
	if !flavor.typescript && reportNode == nameNode {
		if textRange, extended := eslintCoreDefinitionNameRange(ctx.SourceFile, nameNode, definition); extended {
			if len(suggestions) > 0 {
				reporter.reportRangeWithSuggestions(textRange, message, suggestions...)
			} else {
				reporter.reportRange(textRange, message)
			}
			return
		}
	}
	if len(suggestions) > 0 {
		reporter.reportNodeWithSuggestions(reportNode, message, suggestions...)
	} else {
		reporter.reportNode(reportNode, message)
	}
}

type VariableInfo struct {
	Variable       *ast.Node
	Used           bool
	OnlyUsedAsType bool
	References     []*ast.Node
	Definition     *ast.Node
}

func parseOptions(options []any) Config {
	config := Config{
		Vars:         "all",
		Args:         "after-used",
		CaughtErrors: "all",
	}

	if len(options) > 0 {
		if vars, ok := options[0].(string); ok {
			config.Vars = vars
			return compilePatterns(config)
		}
	}

	if optsMap := utils.GetOptionsMap(options); optsMap != nil {
		parseOptionsFromMap(optsMap, &config)
	}

	return compilePatterns(config)
}

func parseOptionsFromMap(optsMap map[string]interface{}, config *Config) {
	if val, ok := optsMap["vars"].(string); ok {
		config.Vars = val
	}
	if val, ok := optsMap["varsIgnorePattern"].(string); ok {
		config.VarsIgnorePattern = val
	}
	if val, ok := optsMap["args"].(string); ok {
		config.Args = val
	}
	if val, ok := optsMap["argsIgnorePattern"].(string); ok {
		config.ArgsIgnorePattern = val
	}
	if val, ok := optsMap["caughtErrors"].(string); ok {
		config.CaughtErrors = val
	}
	if val, ok := optsMap["caughtErrorsIgnorePattern"].(string); ok {
		config.CaughtErrorsIgnorePattern = val
	}
	if val, ok := optsMap["destructuredArrayIgnorePattern"].(string); ok {
		config.DestructuredArrayIgnorePattern = val
	}
	if val, ok := optsMap["ignoreRestSiblings"].(bool); ok {
		config.IgnoreRestSiblings = val
	}
	if val, ok := optsMap["ignoreClassWithStaticInitBlock"].(bool); ok {
		config.IgnoreClassWithStaticInitBlock = val
	}
	if val, ok := optsMap["ignoreUsingDeclarations"].(bool); ok {
		config.IgnoreUsingDeclarations = val
	}
	if val, ok := optsMap["reportUsedIgnorePattern"].(bool); ok {
		config.ReportUsedIgnorePattern = val
	}
	if val, ok := optsMap["enableAutofixRemoval"].(map[string]interface{}); ok {
		if imports, ok := val["imports"].(bool); ok {
			config.EnableAutofixRemoval.Imports = imports
		}
	}
}

func compilePatterns(config Config) Config {
	if config.VarsIgnorePattern != "" {
		config.varsIgnoreRe, _ = utils.CompileRegexp2(config.VarsIgnorePattern, utils.JSUnicodeRegexOptions)
	}
	if config.ArgsIgnorePattern != "" {
		config.argsIgnoreRe, _ = utils.CompileRegexp2(config.ArgsIgnorePattern, utils.JSUnicodeRegexOptions)
	}
	if config.CaughtErrorsIgnorePattern != "" {
		config.caughtErrorsIgnoreRe, _ = utils.CompileRegexp2(config.CaughtErrorsIgnorePattern, utils.JSUnicodeRegexOptions)
	}
	if config.DestructuredArrayIgnorePattern != "" {
		config.destructuredArrayIgnoreRe, _ = utils.CompileRegexp2(config.DestructuredArrayIgnorePattern, utils.JSUnicodeRegexOptions)
	}
	return config
}

// isPartOfAssignment checks if an identifier is a write-only target in an
// assignment (simple =) or for-in/for-of initializer. Uses the TypeScript-go
// public API GetAssignmentTarget which handles all destructuring patterns,
// parenthesized expressions, and for-in/for-of loops.
func isPartOfAssignment(node *ast.Node) bool {
	target := ast.GetAssignmentTarget(node)
	if target == nil {
		return false
	}
	// For simple assignment (=), the target is the LHS identifier → write-only.
	// Compound assignments (+=, etc.) also read, so they are NOT write-only.
	if target.Kind == ast.KindBinaryExpression {
		bin := target.AsBinaryExpression()
		return bin != nil && bin.OperatorToken.Kind == ast.KindEqualsToken
	}
	// For-in/for-of initializers are write targets, UNLESS the first statement
	// in the loop body is a ReturnStatement (pattern for checking property existence).
	// This matches ESLint's isForInOfRef() logic (see #2342).
	if target.Kind == ast.KindForInStatement || target.Kind == ast.KindForOfStatement {
		forStmt := target.AsForInOrOfStatement()
		if forStmt != nil && forStmt.Statement != nil &&
			forInBodyStartsWithReturn(forStmt.Statement) && isDirectForInOfTarget(node, target) {
			return false // Not write-only — the variable is meaningfully used
		}
		return true
	}
	return false
}

// isDirectForInOfTarget mirrors the narrow shape recognized by ESLint's
// isForInOfRef: only `for (x in/of value)` and `for (let x in/of value)` get
// the return-first-body exception. A binding nested in `[x]` or `{x}` remains
// a write-only loop target even when the body starts with return.
func isDirectForInOfTarget(node *ast.Node, loop *ast.Node) bool {
	if node == nil || !ast.IsForInOrOfStatement(loop) {
		return false
	}
	statement := loop.AsForInOrOfStatement()
	if statement == nil || statement.Initializer == nil {
		return false
	}
	initializer := ast.SkipParentheses(statement.Initializer)
	if initializer == node {
		return true
	}
	if initializer == nil || initializer.Kind != ast.KindVariableDeclarationList {
		return false
	}
	declarations := initializer.AsVariableDeclarationList().Declarations
	if declarations == nil || len(declarations.Nodes) != 1 {
		return false
	}
	declaration := declarations.Nodes[0].AsVariableDeclaration()
	if declaration == nil {
		return false
	}
	name := declaration.Name()
	return name != nil && ast.SkipParentheses(name) == node
}

// isUpdateTarget checks if the identifier is the operand of a prefix/postfix
// increment or decrement (++x, x++, --x, x--).
func isUpdateTarget(node *ast.Node) bool {
	return updateExpressionForTarget(node) != nil
}

func updateExpressionForTarget(node *ast.Node) *ast.Node {
	target := ast.GetAssignmentTarget(node)
	if target == nil {
		return nil
	}
	if target.Kind == ast.KindPrefixUnaryExpression || target.Kind == ast.KindPostfixUnaryExpression {
		return target
	}
	return nil
}

// isCompoundAssignmentTarget checks if the identifier is the LHS of a compound
// assignment, including logical assignments but excluding simple `=`.
func isCompoundAssignmentTarget(node *ast.Node) bool {
	return compoundAssignmentForTarget(node) != nil
}

func compoundAssignmentForTarget(node *ast.Node) *ast.Node {
	target := ast.GetAssignmentTarget(node)
	if target == nil || target.Kind != ast.KindBinaryExpression {
		return nil
	}
	bin := target.AsBinaryExpression()
	op := bin.OperatorToken.Kind
	if !ast.IsCompoundAssignment(op) || op == ast.KindEqualsToken {
		return nil
	}
	return target
}

// hasAssignment checks if a variable declaration has an initializer or any
// write references. Used to distinguish "assigned a value but never used"
// from "defined but never used" in error messages.
func hasAssignment(definition *ast.Node, sym *ast.Symbol, writeRefs map[*ast.Symbol][]*ast.Node, flavor ruleFlavor) bool {
	if definition != nil {
		switch definition.Kind {
		case ast.KindVariableDeclaration:
			// Variable with initializer: `const x = 1`, `let x = 1`
			varDecl := definition.AsVariableDeclaration()
			if varDecl != nil && (varDecl.Initializer != nil || isForInOfDeclaration(definition) != nil) {
				return true
			}
		case ast.KindBindingElement:
			// @typescript-eslint historically classifies destructured bindings
			// as assigned. ESLint core classifies variable-pattern bindings as
			// assigned, but parameter and catch-clause patterns as defined.
			if flavor.typescript {
				return true
			}
			if bindingElement := definition.AsBindingElement(); bindingElement != nil && bindingElement.Initializer != nil {
				return true
			}
		bindingAncestors:
			for current := definition.Parent; current != nil; current = current.Parent {
				switch current.Kind {
				case ast.KindParameter, ast.KindCatchClause:
					break bindingAncestors
				case ast.KindVariableDeclaration:
					variable := current.AsVariableDeclaration()
					if variable != nil && (variable.Initializer != nil || isForInOfDeclaration(current) != nil) {
						return true
					}
					break bindingAncestors
				}
			}
		case ast.KindParameter:
			// Parameters with default values: `function f(x = 1)`
			paramDecl := definition.AsParameterDeclaration()
			if paramDecl != nil && paramDecl.Initializer != nil {
				return true
			}
		}
	}
	if sym != nil {
		if refs, exists := writeRefs[sym]; exists && len(refs) > 0 {
			return true
		}
	}
	return false
}

// isInsideLoop checks whether node is lexically inside a loop construct (for,
// for-in, for-of, while, do-while) without crossing a function boundary first.
// Mirrors ESLint's astUtils.isInLoop, used to decide whether a self-referencing
// accumulator assignment (`x = f(x)` inside a loop) can still be observed on a
// later iteration — and so counts as a real use rather than self-modification.
func isInsideLoop(node *ast.Node) bool {
	for current := node; current != nil && !ast.IsFunctionLike(current); current = current.Parent {
		if ast.IsIterationStatement(current, false) {
			return true
		}
	}
	return false
}

func lastWriteInDeclarationScope(definition *ast.Node, refs []*ast.Node) *ast.Node {
	declarationScope := utils.FindEnclosingScope(definition)
	for index := len(refs) - 1; index >= 0; index-- {
		if utils.FindEnclosingScope(refs[index]) == declarationScope {
			return refs[index]
		}
	}
	return nil
}

func lastWriteToReport(definition *ast.Node, refs []*ast.Node, flavor ruleFlavor) *ast.Node {
	if flavor.typescript {
		if len(refs) == 0 {
			return nil
		}
		return refs[len(refs)-1]
	}
	return lastWriteInDeclarationScope(definition, refs)
}

// isSelfModifyingReference checks if a read reference to a variable is ONLY
// used to modify the same variable, with the result not used elsewhere.
// Examples: `a = a + 1;`, `a++;`, `a += 1;` (as statements, not sub-expressions).
//
// declNode anchors the variable's own declaration site. A self-referencing
// assignment (`x = f(x)`) does NOT count as self-modification — i.e. it IS a
// real use — when the assignment happens in a different function scope than
// the declaration, or inside a loop: the written value can be observed later
// (a closure capturing it, or the next loop iteration), so it's a genuine
// read-modify-write accumulator rather than a discarded self-reference.
func isSelfModifyingReference(node *ast.Node, sym *ast.Symbol, name string, checker *checker.Checker, declNode *ast.Node, sourceFile *ast.SourceFile) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	// Case 1: a++ or a-- (update expression as a statement)
	if update := updateExpressionForTarget(node); update != nil {
		return isUnusedExpression(update)
	}

	// Case 2: a += expr (compound assignment — the LHS identifier is both read and written).
	// Logical assignments (??=, &&=, ||=) are NOT self-modifying because they conditionally
	// assign and ESLint considers them as meaningful usage.
	if assignment := compoundAssignmentForTarget(node); assignment != nil {
		op := assignment.AsBinaryExpression().OperatorToken.Kind
		if op == ast.KindBarBarEqualsToken || op == ast.KindAmpersandAmpersandEqualsToken || op == ast.KindQuestionQuestionEqualsToken {
			return false // Logical assignment — not self-modifying
		}
		return isUnusedExpression(assignment)
	}

	// Case 3: a = <any expression tree containing a>. Walking assignment
	// ancestors instead of enumerating expression kinds covers element access,
	// calls/new, templates, conditionals, containers, and TypeScript wrappers.
	// The assignment must discard its result and execute in the declaration's
	// variable scope; loop-carried and cross-scope updates can be observed later.
	declarationScope := variableScope(declNode, sourceFile)
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind != ast.KindBinaryExpression {
			continue
		}
		binary := current.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil || !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			continue
		}
		if !ast.IsNodeDescendantOf(node, binary.Right) ||
			!referenceMatchesVariable(binary.Left, sym, name, checker, sourceFile) ||
			!isUnusedExpression(current) ||
			utils.FindEnclosingScope(current) != declarationScope ||
			isInsideLoop(current) {
			continue
		}
		return !isInsideOfStorableFunction(node, binary.Right)
	}

	return false
}

func variableScope(declaration *ast.Node, sourceFile *ast.SourceFile) *ast.Node {
	if declaration != nil {
		return utils.FindEnclosingScope(declaration)
	}
	if sourceFile != nil {
		return sourceFile.AsNode()
	}
	return nil
}

// referenceMatchesVariable compares a direct assignment target with either a
// checker symbol or an inline-global name. Inline globals intentionally reject
// symbols declared in this source file so shadowing never counts as global use.
func referenceMatchesVariable(node *ast.Node, sym *ast.Symbol, name string, typeChecker *checker.Checker, sourceFile *ast.SourceFile) bool {
	// ESTree erases parentheses around an assignment target but retains
	// TypeScript assertion nodes. Unwrap only parentheses to keep that shape.
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsIdentifier(node) || node.Text() != name {
		return false
	}
	referenceSymbol := utils.GetReferenceSymbol(node, typeChecker)
	if sym != nil {
		return referenceSymbol == sym
	}
	return referenceSymbol == nil || !utils.IsSymbolDeclaredInFile(referenceSymbol, sourceFile)
}

// isUnusedExpression checks if an expression's result is not consumed.
// Walks up through parentheses and comma expressions to find the ultimate
// consumer. Returns true if the expression ends up in an ExpressionStatement
// (result discarded) or as the left operand of a comma (result discarded).
func isUnusedExpression(node *ast.Node) bool {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		// Direct statement: result is discarded
		if ast.IsExpressionStatement(parent) {
			return true
		}
		// Parenthesized expression: keep walking up
		if parent.Kind == ast.KindParenthesizedExpression {
			current = parent
			continue
		}
		// Comma expression
		if parent.Kind == ast.KindBinaryExpression {
			bin := parent.AsBinaryExpression()
			if bin != nil && bin.OperatorToken.Kind == ast.KindCommaToken {
				if bin.Left == current {
					// Left operand of comma: result is discarded
					return true
				}
				// Right operand of comma: the comma's value IS this operand,
				// but check if the comma itself is unused
				current = parent
				continue
			}
		}
		return false
	}
	return false
}

// isInsideOwnDeclaration checks if a usage reference is inside the body of its
// own declaration. This covers:
//   - namespace self-reference: `namespace Foo { Foo.Bar }` — Foo used only inside itself
//   - recursive function: `function foo() { return foo(); }` — foo calls only itself
func isInsideOwnDeclaration(usage *ast.Node, definition *ast.Node) bool {
	return isInsideAnyOwnDeclaration(usage, []*ast.Node{definition})
}

// isInsideAnyOwnDeclaration checks if a usage reference is inside the body of
// any of the given declarations. Used for declaration merging (e.g., multiple
// interface declarations for the same symbol).
func isInsideAnyOwnDeclaration(usage *ast.Node, definitions []*ast.Node) bool {
	for _, definition := range definitions {
		if definition == nil {
			continue
		}
		var body *ast.Node
		switch definition.Kind {
		case ast.KindModuleDeclaration:
			body = definition.Body()
		case ast.KindFunctionDeclaration:
			// The function scope includes parameter initializers as well as the
			// body, so recursive defaults such as `function f(x = f) {}` are self-use.
			body = definition
		case ast.KindClassDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeAliasDeclaration,
			ast.KindEnumDeclaration:
			// Class declarations have a class-local name binding, just like
			// ESLint's class scope. A reference from extends, a computed key,
			// a field, method, or static block therefore does not consume the
			// declaration in the enclosing scope.
			// Self-referencing types/enums: `interface Foo { baz: Foo['bar'] }`,
			// `type Foo = { bar: Foo }`, `enum Foo { B = Foo.A }`.
			body = definition
		case ast.KindVariableDeclaration:
			// For `var a = function() { a() }` or `const a = () => { a() }`,
			// the definition is the VariableDeclaration. The initializer is the
			// function expression whose body contains the self-reference.
			varDecl := definition.AsVariableDeclaration()
			if varDecl != nil && varDecl.Initializer != nil {
				// ESTree erases parentheses around an initializer; mirror that
				// behavior without erasing TS assertion wrappers, which are semantic
				// nodes in @typescript-eslint's AST.
				initializer := ast.SkipParentheses(varDecl.Initializer)
				if initializer != nil && ast.IsFunctionLike(initializer) {
					body = initializer
				}
			}
		default:
			continue
		}
		if body == nil {
			continue
		}
		if ast.IsNodeDescendantOf(usage, body) {
			return true
		}
	}
	return false
}

// isInsideOfStorableFunction reports whether the nearest function containing
// a RHS reference can escape through another value path. A function invoked as
// the RHS callee is not storable; one passed as an argument or assigned/yielded
// within the RHS can execute later and therefore makes the reference a real use.
func isInsideOfStorableFunction(node *ast.Node, rhs *ast.Node) bool {
	function := ast.FindAncestor(node.Parent, func(current *ast.Node) bool {
		return ast.IsFunctionLike(current)
	})
	return function != nil && ast.IsNodeDescendantOf(function, rhs) && isStorableFunction(function, rhs)
}

func isStorableFunction(function *ast.Node, rhs *ast.Node) bool {
	node := function
	for parent := function.Parent; parent != nil && ast.IsNodeDescendantOf(parent, rhs); parent = parent.Parent {
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary != nil && binary.OperatorToken != nil {
				operator := binary.OperatorToken.Kind
				if operator == ast.KindCommaToken {
					if binary.Right != node {
						return false
					}
				} else if ast.IsAssignmentOperator(operator) {
					return true
				}
			}
		case ast.KindCallExpression:
			return parent.AsCallExpression().Expression != node
		case ast.KindNewExpression:
			return parent.AsNewExpression().Expression != node
		case ast.KindTaggedTemplateExpression, ast.KindYieldExpression:
			return true
		default:
			// IsStatement includes declaration statements. Include every Block
			// as ESTree's BlockStatement too; tsgo intentionally excludes function,
			// try, and catch bodies from IsStatement. Do not use ast.IsDeclaration,
			// which would also classify property assignments as declarations.
			if parent.Kind == ast.KindBlock || ast.IsStatement(parent) {
				return true
			}
		}
		node = parent
	}
	return false
}

// isParamUsed checks if a parameter name (Identifier or binding pattern) has any usage.
// Recursively checks destructured binding elements.
func isParamUsed(ctx rule.RuleContext, nameNode *ast.Node, allUsages map[*ast.Symbol][]*ast.Node) bool {
	if nameNode == nil {
		return false
	}
	if ast.IsIdentifier(nameNode) {
		sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
		if sym == nil {
			return false
		}
		resolved := ctx.TypeChecker.SkipAlias(sym)
		if usageNodes, exists := allUsages[resolved]; exists {
			for _, usage := range usageNodes {
				if usage.Pos() != nameNode.Pos() {
					return true
				}
			}
		}
		// Also check original sym for import specifiers
		if usageNodes, exists := allUsages[sym]; exists {
			for _, usage := range usageNodes {
				if usage.Pos() != nameNode.Pos() {
					return true
				}
			}
		}
		return false
	}
	// Binding pattern: recursively check each element
	if nameNode.Kind == ast.KindObjectBindingPattern || nameNode.Kind == ast.KindArrayBindingPattern {
		found := false
		nameNode.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				elem := child.AsBindingElement()
				if elem != nil && isParamUsed(ctx, elem.Name(), allUsages) {
					found = true
					return true
				}
			}
			return false
		})
		return found
	}
	return false
}

// forInBodyStartsWithReturn checks if the body of a for-in/for-of loop
// starts with a ReturnStatement. This matches ESLint's isForInOfRef() logic:
// a for-in/for-of variable is considered "used" only when the FIRST statement
// in the loop body is a return statement (pattern for checking property existence).
func forInBodyStartsWithReturn(body *ast.Node) bool {
	if body == nil {
		return false
	}
	// for (x in obj) return; — body IS the return statement
	if body.Kind == ast.KindReturnStatement {
		return true
	}
	// for (x in obj) { return; } — body is a Block, check first statement
	if body.Kind == ast.KindBlock {
		block := body.AsBlock()
		if block != nil && block.Statements != nil && len(block.Statements.Nodes) > 0 {
			return block.Statements.Nodes[0].Kind == ast.KindReturnStatement
		}
	}
	return false
}

// isForInOfDeclaration checks if a VariableDeclaration is the initializer of a
// for-in/for-of loop (e.g., `for (var name in obj)`).
func isForInOfDeclaration(node *ast.Node) *ast.Node {
	if node == nil || node.Parent == nil {
		return nil
	}
	// VariableDeclaration → VariableDeclarationList → ForInStatement/ForOfStatement
	declList := node.Parent
	if declList == nil || declList.Kind != ast.KindVariableDeclarationList {
		return nil
	}
	forStmt := declList.Parent
	if forStmt == nil {
		return nil
	}
	if forStmt.Kind == ast.KindForInStatement || forStmt.Kind == ast.KindForOfStatement {
		return forStmt
	}
	return nil
}

// hasDirectArrayDestructuringWrite checks if the variable has any direct write
// reference via array destructuring assignment (e.g., `[_x] = arr`).
func hasDirectArrayDestructuringWrite(writeRefs map[*ast.Symbol][]*ast.Node, sym *ast.Symbol) bool {
	refs, exists := writeRefs[sym]
	if !exists {
		return false
	}
	for _, ref := range refs {
		// ESTree checks `ref.identifier.parent.type === "ArrayPattern"`.
		// A rest element or default assignment adds an intermediate parent,
		// so those forms intentionally do not qualify.
		if ref.Parent != nil && ref.Parent.Kind == ast.KindArrayLiteralExpression &&
			utils.IsInDestructuringAssignment(ref.Parent) {
			return true
		}
	}
	return false
}

// isDirectArrayDestructuredIdentifier recreates ESTree's direct-parent test
// for a declaration binding. tsgo stores defaults and rest markers on the
// BindingElement itself, while ESTree wraps their identifiers in
// AssignmentPattern/RestElement nodes; exclude those wrappers explicitly.
func isDirectArrayDestructuredIdentifier(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindBindingElement || node.Parent == nil ||
		node.Parent.Kind != ast.KindArrayBindingPattern {
		return false
	}
	element := node.AsBindingElement()
	return element != nil && element.DotDotDotToken == nil && element.Initializer == nil &&
		element.Name() != nil && ast.IsIdentifier(element.Name())
}

// hasObjectRestSiblingDefinition matches ESLint's direct Property sibling
// shape. Nested bindings, defaults, and the rest binding itself do not qualify.
func hasObjectRestSiblingDefinition(definition *ast.Node) bool {
	if definition == nil || definition.Kind != ast.KindBindingElement || definition.Parent == nil ||
		definition.Parent.Kind != ast.KindObjectBindingPattern {
		return false
	}
	element := definition.AsBindingElement()
	if element == nil || element.DotDotDotToken != nil || element.Initializer != nil ||
		element.Name() == nil || !ast.IsIdentifier(element.Name()) {
		return false
	}
	pattern := definition.Parent.AsBindingPattern()
	if pattern == nil || pattern.Elements == nil || len(pattern.Elements.Nodes) == 0 {
		return false
	}
	last := pattern.Elements.Nodes[len(pattern.Elements.Nodes)-1]
	if last.Kind != ast.KindBindingElement {
		return false
	}
	lastElement := last.AsBindingElement()
	return lastElement != nil && lastElement.DotDotDotToken != nil
}

func hasObjectRestSiblingWrite(writeRefs map[*ast.Symbol][]*ast.Node, sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	for _, ref := range writeRefs[sym] {
		property := ref.Parent
		for property != nil && property.Kind == ast.KindParenthesizedExpression {
			property = property.Parent
		}
		if property == nil ||
			(property.Kind != ast.KindPropertyAssignment && property.Kind != ast.KindShorthandPropertyAssignment) {
			continue
		}
		if property.Kind == ast.KindShorthandPropertyAssignment &&
			property.AsShorthandPropertyAssignment().ObjectAssignmentInitializer != nil {
			// In ESTree the identifier is wrapped by an AssignmentPattern,
			// so it is not the direct Property child checked by hasRestSibling.
			continue
		}
		objectNode := property.Parent
		if objectNode == nil || objectNode.Kind != ast.KindObjectLiteralExpression ||
			!utils.IsInDestructuringAssignment(objectNode) {
			continue
		}
		object := objectNode.AsObjectLiteralExpression()
		if object != nil && object.Properties != nil && len(object.Properties.Nodes) > 0 &&
			object.Properties.Nodes[len(object.Properties.Nodes)-1].Kind == ast.KindSpreadAssignment {
			return true
		}
	}
	return false
}

// isParameterInWithoutBodyDeclaration checks if a parameter is in a function-like
// declaration that has no body (overload signatures, abstract methods, type-level
// constructs). The TypeScript extension rule ignores these definitions, while
// ESLint core checks the parser-created variables in their signature scope.
func isParameterInWithoutBodyDeclaration(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	switch parent.Kind {
	case ast.KindFunctionDeclaration,
		ast.KindMethodDeclaration,
		ast.KindMethodSignature,
		ast.KindConstructor:
		return parent.Body() == nil
	// Setter parameters are required by syntax and should never be reported.
	case ast.KindSetAccessor:
		return true
	// Type-level function-like constructs never have a body.
	// Parameters in these are part of type signatures.
	case ast.KindFunctionType,
		ast.KindConstructorType,
		ast.KindCallSignature,
		ast.KindConstructSignature,
		ast.KindIndexSignature:
		return true
	}
	return false
}

// isInsideAmbientModuleBlock checks if the node is inside an ambient (declare)
// module/namespace block. Non-declare namespaces (e.g., `export namespace Foo { ... }`)
// should still have their contents checked for unused vars.
//
// Special case: if the immediate parent module block has explicit exports
// (export =, export default, or export { ... }), non-exported declarations are
// private and should be checked — return false for them.
func isInsideAmbientModuleBlock(node *ast.Node) bool {
	moduleBlock := ast.FindAncestorKind(node, ast.KindModuleBlock)
	if moduleBlock == nil {
		return false
	}
	// The ModuleBlock's parent is the ModuleDeclaration.
	moduleDecl := moduleBlock.Parent
	if moduleDecl == nil || moduleDecl.Kind != ast.KindModuleDeclaration {
		return false
	}

	// tsgo propagates NodeFlagsAmbient through declaration files, `declare`
	// modules, nested namespaces, and global augmentations.
	if !utils.IsInAmbientContext(node) {
		return false
	}
	// If the module block has explicit exports, non-exported declarations
	// should be checked (return false).
	if moduleBlockHasExplicitExports(moduleBlock) {
		return false
	}
	return true
}

// moduleBlockHasExplicitExports checks if a ModuleBlock contains any explicit export
// statements (export =, export default, export { ... }, export * from '...').
// Note: export modifier on declarations (export const, export namespace, etc.) does NOT count.
func moduleBlockHasExplicitExports(moduleBlock *ast.Node) bool {
	found := false
	moduleBlock.ForEachChild(func(child *ast.Node) bool {
		switch child.Kind {
		case ast.KindExportAssignment:
			// export = x or export default x
			found = true
			return true
		case ast.KindExportDeclaration:
			// export { ... } or export {} or export * from '...'
			found = true
			return true
		}
		return false
	})
	return found
}

// isInDtsWithoutExplicitExports checks if a node is in a .d.ts file where
// it should be implicitly ambient (not reported). In .d.ts files:
// - Top-level declarations without explicit exports are globally visible → skip
// - Top-level declarations with explicit exports are module-scoped → check
// - Declarations inside namespaces are handled by isInsideAmbientModuleBlock
func isInDtsWithoutExplicitExports(node *ast.Node) bool {
	sf := ast.GetSourceFileOfNode(node)
	if sf == nil || !sf.IsDeclarationFile {
		return false
	}
	// Find the containing scope: either a ModuleBlock or the SourceFile
	// If we're inside a module block, isInsideAmbientModuleBlock handles it.
	moduleBlock := ast.FindAncestorKind(node, ast.KindModuleBlock)
	if moduleBlock != nil {
		return false // handled by isInsideAmbientModuleBlock
	}
	// We're at the top level of a .d.ts file.
	// Check if the source file has explicit exports.
	sourceNode := sf.AsNode()
	hasExplicit := false
	sourceNode.ForEachChild(func(child *ast.Node) bool {
		switch child.Kind {
		case ast.KindExportAssignment:
			hasExplicit = true
			return true
		case ast.KindExportDeclaration:
			hasExplicit = true
			return true
		}
		return false
	})
	return !hasExplicit
}

func isParameterNode(node *ast.Node) bool {
	return node != nil && ast.IsPartOfParameterDeclaration(node)
}

func isCaughtErrorNode(node *ast.Node) bool {
	return node != nil && ast.IsCatchClauseVariableDeclarationOrBindingElement(node)
}

func isUsingDeclaration(definition *ast.Node) bool {
	if definition != nil && definition.Kind == ast.KindBindingElement {
		definition = utils.EnclosingVariableDeclarationOfBindingElement(definition)
	}
	if definition == nil || definition.Kind != ast.KindVariableDeclaration {
		return false
	}
	return ast.IsVarUsing(definition) || ast.IsVarAwaitUsing(definition)
}

func hasStaticInitBlock(classNode *ast.Node) bool {
	found := false
	classNode.ForEachChild(func(child *ast.Node) bool {
		if child.Kind == ast.KindClassStaticBlockDeclaration {
			found = true
			return true
		}
		return false
	})
	return found
}

// matchesIgnorePattern checks if a variable name matches its category's
// ignore pattern, and whether the match should result in ignoring or
// reporting (when reportUsedIgnorePattern is true and the variable is used).
// Returns: (shouldIgnore bool, matchesPattern bool, matched variable type)
func matchesIgnorePattern(varName string, varInfo *VariableInfo, opts Config, writeRefs map[*ast.Symbol][]*ast.Node, sym *ast.Symbol) (bool, bool, variableType) {
	var re *regexp2.Regexp
	kind := variableTypeVariable
	matched := false

	// ESLint evaluates destructuredArrayIgnorePattern before the variable's
	// ordinary category. This matters when both patterns match, and when
	// args/caughtErrors is "none" together with reportUsedIgnorePattern.
	if opts.destructuredArrayIgnoreRe != nil &&
		(isDirectArrayDestructuredIdentifier(varInfo.Definition) || hasDirectArrayDestructuringWrite(writeRefs, sym)) &&
		utils.Regexp2MatchString(opts.destructuredArrayIgnoreRe, varName) {
		kind = variableTypeArrayDestructure
		matched = true
	} else if isParameterNode(varInfo.Definition) {
		kind = variableTypeParameter
		if opts.Args == "none" {
			return true, false, kind
		}
		re = opts.argsIgnoreRe
	} else if isCaughtErrorNode(varInfo.Definition) {
		kind = variableTypeCatchClause
		if opts.CaughtErrors == "none" {
			return true, false, kind
		}
		re = opts.caughtErrorsIgnoreRe
	} else {
		re = opts.varsIgnoreRe
	}

	if !matched {
		matched = re != nil && utils.Regexp2MatchString(re, varName)
	}

	if !matched {
		return false, false, kind
	}

	// Pattern matches. If used + reportUsedIgnorePattern, don't ignore — report instead.
	if varInfo.Used && opts.ReportUsedIgnorePattern {
		return false, true, kind
	}

	return true, true, kind
}

func ignorePatternAdditional(kind variableType, opts Config, used bool) string {
	var description, pattern string
	switch kind {
	case variableTypeArrayDestructure:
		description = "elements of array destructuring"
		pattern = opts.DestructuredArrayIgnorePattern
	case variableTypeCatchClause:
		description = "caught errors"
		pattern = opts.CaughtErrorsIgnorePattern
	case variableTypeParameter:
		description = "args"
		pattern = opts.ArgsIgnorePattern
	default:
		description = "vars"
		pattern = opts.VarsIgnorePattern
	}
	if pattern == "" {
		return ""
	}
	if used {
		return fmt.Sprintf(". Used %s must not match /%s/u", description, pattern)
	}
	return fmt.Sprintf(". Allowed unused %s must match /%s/u", description, pattern)
}

func definitionVariableType(definition *ast.Node, opts Config) variableType {
	if opts.DestructuredArrayIgnorePattern != "" && isDirectArrayDestructuredIdentifier(definition) {
		return variableTypeArrayDestructure
	}
	if isCaughtErrorNode(definition) {
		return variableTypeCatchClause
	}
	if isParameterNode(definition) {
		return variableTypeParameter
	}
	return variableTypeVariable
}

// isBeforeLastUsedParam checks if an unused parameter appears before a later
// parameter that is used (or has a default value). Used for the "after-used"
// args option: unused parameters before the last used one are allowed because
// they serve as positional placeholders.
func isBeforeLastUsedParam(ctx rule.RuleContext, paramNode *ast.Node, allUsages map[*ast.Symbol][]*ast.Node) bool {
	if paramNode == nil || paramNode.Parent == nil {
		return false
	}

	funcLike := paramNode.Parent
	params := funcLike.Parameters()
	if len(params) == 0 {
		return false
	}

	paramIndex := -1
	for i, p := range params {
		if p.AsNode() == paramNode {
			paramIndex = i
			break
		}
	}
	if paramIndex < 0 {
		return false
	}

	for i := paramIndex + 1; i < len(params); i++ {
		sibling := params[i]

		// A parameter with a default value (initializer) counts as a
		// meaningful position marker. ESLint's after-used skips params
		// before a later param that has a default value.
		if sibling.AsNode().Initializer() != nil {
			return true
		}

		if isParamUsed(ctx, sibling.Name(), allUsages) {
			return true
		}
	}

	return false
}

// hasExplicitExportModifier checks only source-written export modifiers. The
// checker's combined flags also mark ambient namespace members and enum members
// as implicitly exported, but ESLint core does not treat those flags as uses.
func hasExplicitExportModifier(node *ast.Node) bool {
	if node == nil {
		return false
	}
	root := ast.GetRootDeclaration(node)
	if root != nil {
		node = root
	}
	if ast.HasSyntacticModifier(node, ast.ModifierFlagsExport) {
		return true
	}
	if node.Kind != ast.KindVariableDeclaration || node.Parent == nil ||
		node.Parent.Kind != ast.KindVariableDeclarationList || node.Parent.Parent == nil {
		return false
	}
	return node.Parent.Parent.Kind == ast.KindVariableStatement &&
		ast.HasSyntacticModifier(node.Parent.Parent, ast.ModifierFlagsExport)
}

// isExported checks if a variable is exported from the module. Exported variables
// are excluded from unused-var reporting. Checks: export modifier on the declaration,
// export modifier on any merged declaration (declaration merging), parent
// VariableStatement export, and re-export via `export { name }`.
func isExported(ctx rule.RuleContext, varInfo *VariableInfo, localExportTargets map[*ast.Symbol]bool, flavor ruleFlavor) bool {
	if varInfo.Variable == nil {
		return false
	}
	if !flavor.typescript && varInfo.Definition != nil && varInfo.Definition.Kind == ast.KindEnumMember {
		// Exporting the containing enum consumes the enum binding, not each
		// member variable created by @typescript-eslint/scope-manager.
		return false
	}

	if varInfo.Definition != nil {
		exported := hasExplicitExportModifier(varInfo.Definition)
		if flavor.typescript {
			// The extension rule follows TypeScript's implicit ambient exports.
			exported = ast.GetCombinedModifierFlags(varInfo.Definition)&ast.ModifierFlagsExport != 0
		}
		if exported {
			return true
		}

		// Declaration merging: if ANY declaration of the symbol is exported,
		// the variable is considered exported (e.g., `interface Foo {} export const Foo = ...`)
		sym := ctx.TypeChecker.GetSymbolAtLocation(varInfo.Variable)
		if sym != nil && len(sym.Declarations) > 1 {
			for _, decl := range sym.Declarations {
				if !flavor.typescript && ast.GetSourceFileOfNode(decl) != ctx.SourceFile {
					// The TS checker merges module augmentations with declarations
					// from the augmented package. ESLint core's scope manager keeps
					// the local augmentation binding separate.
					continue
				}
				declExported := hasExplicitExportModifier(decl)
				if flavor.typescript {
					declExported = ast.GetCombinedModifierFlags(decl)&ast.ModifierFlagsExport != 0
				}
				if declExported {
					return true
				}
			}
		}
	}

	parent := varInfo.Variable.Parent
	for parent != nil {
		if parent.Kind == ast.KindExportDeclaration {
			return true
		}
		parent = parent.Parent
	}

	for _, ref := range varInfo.References {
		refParent := ref.Parent
		for refParent != nil {
			if refParent.Kind == ast.KindExportDeclaration {
				return true
			}
			refParent = refParent.Parent
		}
	}

	// Local export declarations may live inside nested namespaces. The source
	// walk records their checker-resolved targets once, so this lookup handles
	// arbitrary nesting without rescanning the file for every variable.
	if varInfo.Definition != nil {
		sym := ctx.TypeChecker.GetSymbolAtLocation(varInfo.Variable)
		if sym != nil && localExportTargets[sym] {
			return true
		}
	}

	return false
}

func buildUnusedVarMessage(varName string, hasAssignment bool, additional string) rule.RuleMessage {
	action := "defined"
	if hasAssignment {
		action = "assigned a value"
	}
	return rule.RuleMessage{
		Id:          "unusedVar",
		Description: fmt.Sprintf("'%s' is %s but never used%s.", varName, action, additional),
		Data: map[string]string{
			"varName":    varName,
			"action":     action,
			"additional": additional,
		},
	}
}

func buildUsedOnlyAsTypeMessage(varName string, hasAssignment bool) rule.RuleMessage {
	desc := fmt.Sprintf("'%s' is defined but only used as a type.", varName)
	if hasAssignment {
		desc = fmt.Sprintf("'%s' is assigned a value but only used as a type.", varName)
	}
	return rule.RuleMessage{
		Id:          "usedOnlyAsType",
		Description: desc,
		Data: map[string]string{
			"varName": varName,
		},
	}
}

func buildUsedIgnoredVarMessage(varName string, additional string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "usedIgnoredVar",
		Description: fmt.Sprintf("'%s' is marked as ignored but is used%s.", varName, additional),
		Data: map[string]string{
			"varName":    varName,
			"additional": additional,
		},
	}
}

func buildRemoveVarMessage(varName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeVar",
		Description: fmt.Sprintf("Remove unused variable '%s'.", varName),
		Data: map[string]string{
			"varName": varName,
		},
	}
}

func buildRemoveUnusedImportMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeUnusedImportDeclaration",
		Description: "Remove unused import declaration.",
	}
}

func buildRemoveUnusedVarMessage(varName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "removeUnusedVar",
		Description: fmt.Sprintf("Remove unused variable \"%s\".", varName),
	}
}

// getImportRemoveFix generates a fix that removes an unused import binding.
// It handles four import kinds: ImportSpecifier, ImportClause (default import),
// NamespaceImport, and ImportEqualsDeclaration.
func getImportRemoveFix(ctx rule.RuleContext, definition *ast.Node, reportedUnused map[*ast.Node]bool) (fixes []rule.RuleFix, suggestionMsg rule.RuleMessage) {
	file := ctx.SourceFile
	switch definition.Kind {
	case ast.KindImportEqualsDeclaration:
		// import X = require('foo') → remove entire statement
		return []rule.RuleFix{removeImportLine(file, definition)}, buildRemoveUnusedImportMessage()

	case ast.KindNamespaceImport:
		// import * as ns from 'foo' → remove entire import declaration
		importDecl := getImportDeclaration(definition)
		if importDecl != nil {
			return []rule.RuleFix{removeImportLine(file, importDecl)}, buildRemoveUnusedImportMessage()
		}

	case ast.KindImportClause:
		// import Foo from 'foo' (default import)
		importDecl := getImportDeclaration(definition)
		if importDecl == nil {
			break
		}
		clause := definition.AsImportClause()
		if clause == nil {
			break
		}
		// If no named bindings, or all specifiers are unused → remove entire declaration
		if clause.NamedBindings == nil || allImportSpecifiersUnused(clause, reportedUnused) {
			return []rule.RuleFix{removeImportLine(file, importDecl)}, buildRemoveUnusedImportMessage()
		}
		// Otherwise remove just the default specifier and trailing comma
		// `import Foo, { Used } from 'foo'` → `import { Used } from 'foo'`
		nameNode := clause.Name()
		if nameNode == nil {
			break
		}
		// Find the comma after the default specifier
		commaEnd := findCommaAfter(file, nameNode.End())
		if commaEnd > 0 {
			return []rule.RuleFix{rule.RuleFixRemoveRange(nameNode.Loc.WithEnd(commaEnd))}, buildRemoveUnusedVarMessage(nameNode.AsIdentifier().Text)
		}

	case ast.KindImportSpecifier:
		// import { Foo } from 'foo' (named import specifier)
		importDecl := getImportDeclaration(definition)
		if importDecl == nil {
			break
		}
		clause := importDecl.AsImportDeclaration().ImportClause
		if clause == nil {
			break
		}
		importClause := clause.AsImportClause()
		if importClause == nil {
			break
		}
		// If all specifiers in this declaration are unused → remove entire declaration
		if allImportSpecifiersUnused(importClause, reportedUnused) {
			return []rule.RuleFix{removeImportLine(file, importDecl)}, buildRemoveUnusedImportMessage()
		}
		// Otherwise remove just this specifier with its leading or trailing comma
		return []rule.RuleFix{removeSpecifierWithComma(file, definition)}, buildRemoveUnusedVarMessage(definition.AsImportSpecifier().Name().AsIdentifier().Text)
	}
	return nil, rule.RuleMessage{}
}

func getImportDeclaration(node *ast.Node) *ast.Node {
	current := node
	for current != nil {
		if current.Kind == ast.KindImportDeclaration {
			return current
		}
		current = current.Parent
	}
	return nil
}

func allImportSpecifiersUnused(clause *ast.ImportClause, reportedUnused map[*ast.Node]bool) bool {
	// Check default import
	if clause.Name() != nil && !reportedUnused[clause.AsNode()] {
		return false
	}
	// Check named bindings
	if clause.NamedBindings != nil {
		nb := clause.NamedBindings
		if nb.Kind == ast.KindNamespaceImport {
			return reportedUnused[nb]
		}
		if nb.Kind == ast.KindNamedImports {
			namedImports := nb.AsNamedImports()
			if namedImports != nil && namedImports.Elements != nil {
				for _, spec := range namedImports.Elements.Nodes {
					if !reportedUnused[spec] {
						return false
					}
				}
			}
		}
	}
	return true
}

func removeImportLine(file *ast.SourceFile, node *ast.Node) rule.RuleFix {
	// Remove the entire line including trailing newline
	text := file.Text()
	start := node.Pos()
	end := node.End()
	// Expand to include trailing newline
	if end < len(text) && text[end] == '\n' {
		end++
	} else if end+1 < len(text) && text[end] == '\r' && text[end+1] == '\n' {
		end += 2
	}
	return rule.RuleFixRemoveRange(node.Loc.WithPos(start).WithEnd(end))
}

func findCommaAfter(file *ast.SourceFile, pos int) int {
	token, ok := utils.TokenAtOrAfter(file, pos)
	if ok && token.Kind == ast.KindCommaToken {
		return token.End
	}
	return -1
}

func removeSpecifierWithComma(file *ast.SourceFile, specNode *ast.Node) rule.RuleFix {
	text := file.Text()
	start := specNode.Pos()
	end := specNode.End()

	// Skip leading trivia (whitespace) to get the actual text start
	textStart := start
	for textStart < end && (text[textStart] == ' ' || text[textStart] == '\t' || text[textStart] == '\n' || text[textStart] == '\r') {
		textStart++
	}

	// Try to remove leading comma: `, Specifier` (preferred — avoids double space)
	leadingComma := -1
	for i := start - 1; i >= 0; i-- {
		ch := text[i]
		if ch == ',' {
			leadingComma = i
			break
		}
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			break
		}
	}
	if leadingComma >= 0 {
		return rule.RuleFixRemoveRange(specNode.Loc.WithPos(leadingComma).WithEnd(end))
	}
	// Fallback: remove trailing comma: `Specifier, ` (for first specifier in list)
	trailingEnd := findCommaAfter(file, end)
	if trailingEnd > 0 {
		// Also skip space after comma
		for trailingEnd < len(text) && (text[trailingEnd] == ' ' || text[trailingEnd] == '\t') {
			trailingEnd++
		}
		// Use textStart (not start) to preserve leading whitespace after `{`
		return rule.RuleFixRemoveRange(specNode.Loc.WithPos(textStart).WithEnd(trailingEnd))
	}
	// Last resort: just remove the specifier
	return rule.RuleFixRemoveRange(specNode.Loc.WithPos(textStart).WithEnd(end))
}

func tokenBefore(tokens []utils.SourceToken, pos int, skips int) (utils.SourceToken, bool) {
	index := sort.Search(len(tokens), func(index int) bool {
		return tokens[index].End > pos
	}) - 1 - skips
	if index < 0 {
		return utils.SourceToken{}, false
	}
	return tokens[index], true
}

func tokenAfter(tokens []utils.SourceToken, pos int, skips int) (utils.SourceToken, bool) {
	index := sort.Search(len(tokens), func(index int) bool {
		return tokens[index].Start >= pos
	}) + skips
	if index >= len(tokens) {
		return utils.SourceToken{}, false
	}
	return tokens[index], true
}

func removeNodeRange(sourceFile *ast.SourceFile, node *ast.Node) rule.RuleFix {
	return rule.RuleFixRemoveRange(utils.TrimNodeTextRange(sourceFile, node))
}

func isSingleStatementBody(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindIfStatement, ast.KindWithStatement:
		return true
	default:
		return ast.IsIterationStatement(node.Parent, false)
	}
}

func bindingElementCount(pattern *ast.Node) int {
	if pattern == nil || (pattern.Kind != ast.KindObjectBindingPattern && pattern.Kind != ast.KindArrayBindingPattern) {
		return 0
	}
	count := 0
	for _, element := range pattern.AsBindingPattern().Elements.Nodes {
		if element.Kind == ast.KindBindingElement {
			count++
		}
	}
	return count
}

func fixVariableDeclaration(ctx rule.RuleContext, declaration *ast.Node, ac *analysisContext) (rule.RuleFix, bool) {
	if declaration == nil || declaration.Kind != ast.KindVariableDeclaration || isCaughtErrorNode(declaration) {
		return rule.RuleFix{}, false
	}
	declarationList := declaration.Parent
	if declarationList == nil || declarationList.Kind != ast.KindVariableDeclarationList {
		return rule.RuleFix{}, false
	}
	list := declarationList.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil {
		return rule.RuleFix{}, false
	}
	declarations := list.Declarations.Nodes
	if len(declarations) == 0 {
		return rule.RuleFix{}, false
	}

	listParent := declarationList.Parent
	inLoopInitializer := listParent != nil && ast.IsIterationStatement(listParent, false)
	name := declaration.AsVariableDeclaration().Name()
	if inLoopInitializer && (len(declarations) == 1 || (name != nil && ast.IsBindingPattern(name))) {
		return rule.RuleFix{}, false
	}

	declarationRange := utils.TrimNodeTextRange(ctx.SourceFile, declaration)
	if len(declarations) > 1 {
		if before, ok := tokenBefore(ac.tokens, declarationRange.Pos(), 0); ok && before.Text == "," {
			return rule.RuleFixRemoveRange(core.NewTextRange(before.Start, declarationRange.End())), true
		}
		if after, ok := tokenAfter(ac.tokens, declarationRange.End(), 0); ok && after.Text == "," {
			return rule.RuleFixRemoveRange(core.NewTextRange(declarationRange.Pos(), after.End)), true
		}
		return rule.RuleFix{}, false
	}

	if listParent == nil || listParent.Kind != ast.KindVariableStatement {
		return rule.RuleFix{}, false
	}
	if isSingleStatementBody(listParent) {
		return rule.RuleFixReplaceRange(utils.TrimNodeTextRange(ctx.SourceFile, listParent), ";"), true
	}

	statementRange := utils.TrimNodeTextRange(ctx.SourceFile, listParent)
	next, hasNext := tokenAfter(ac.tokens, statementRange.End(), 0)
	if hasNext {
		previous, hasPrevious := tokenBefore(ac.tokens, statementRange.Pos(), 0)
		if next.Kind == ast.KindStringLiteral ||
			(hasPrevious && previous.Text != ";" && previous.Text != "{") {
			return rule.RuleFix{}, false
		}
	}
	return rule.RuleFixRemoveRange(statementRange), true
}

func fixFunctionParameter(ctx rule.RuleContext, parameter *ast.Node, ac *analysisContext) (rule.RuleFix, bool) {
	if parameter == nil || parameter.Kind != ast.KindParameter || parameter.Parent == nil || !ast.IsFunctionLike(parameter.Parent) {
		return rule.RuleFix{}, false
	}
	functionNode := parameter.Parent
	parameters := functionNode.Parameters()
	if len(parameters) == 0 {
		return rule.RuleFix{}, false
	}

	parameterRange := utils.TrimNodeTextRange(ctx.SourceFile, parameter)
	if len(parameters) == 1 {
		if functionNode.Kind == ast.KindArrowFunction {
			after, hasAfter := tokenAfter(ac.tokens, parameterRange.End(), 0)
			functionRange := utils.TrimNodeTextRange(ctx.SourceFile, functionNode)
			if hasAfter && after.Text == "=>" && functionRange.Pos() == parameterRange.Pos() {
				return rule.RuleFixReplaceRange(parameterRange, "()"), true
			}
		}
		return rule.RuleFixRemoveRange(parameterRange), true
	}

	before, hasBefore := tokenBefore(ac.tokens, parameterRange.Pos(), 0)
	after, hasAfter := tokenAfter(ac.tokens, parameterRange.End(), 0)
	if hasBefore && before.Text == "(" && hasAfter && after.Text == "," {
		return rule.RuleFixRemoveRange(core.NewTextRange(parameterRange.Pos(), after.End)), true
	}
	if hasBefore && before.Text == "," {
		return rule.RuleFixRemoveRange(core.NewTextRange(before.Start, parameterRange.End())), true
	}
	return rule.RuleFix{}, false
}

func fixCollapsedBindingPattern(ctx rule.RuleContext, pattern *ast.Node, ac *analysisContext) (rule.RuleFix, bool) {
	if pattern == nil || pattern.Parent == nil {
		return rule.RuleFix{}, false
	}
	switch pattern.Parent.Kind {
	case ast.KindBindingElement:
		return fixBindingElement(ctx, pattern.Parent, ac)
	case ast.KindVariableDeclaration:
		return fixVariableDeclaration(ctx, pattern.Parent, ac)
	case ast.KindParameter:
		return fixFunctionParameter(ctx, pattern.Parent, ac)
	default:
		return rule.RuleFix{}, false
	}
}

func fixBindingElement(ctx rule.RuleContext, element *ast.Node, ac *analysisContext) (rule.RuleFix, bool) {
	if element == nil || element.Kind != ast.KindBindingElement || element.Parent == nil {
		return rule.RuleFix{}, false
	}
	pattern := element.Parent
	if pattern.Kind != ast.KindObjectBindingPattern && pattern.Kind != ast.KindArrayBindingPattern {
		return rule.RuleFix{}, false
	}
	if bindingElementCount(pattern) == 1 {
		return fixCollapsedBindingPattern(ctx, pattern, ac)
	}

	elementRange := utils.TrimNodeTextRange(ctx.SourceFile, element)
	before, hasBefore := tokenBefore(ac.tokens, elementRange.Pos(), 0)
	after, hasAfter := tokenAfter(ac.tokens, elementRange.End(), 0)

	if pattern.Kind == ast.KindArrayBindingPattern {
		if hasBefore && before.Text == "," && hasAfter && after.Text == "]" {
			return rule.RuleFixRemoveRange(core.NewTextRange(before.Start, elementRange.End())), true
		}
		return rule.RuleFixRemoveRange(elementRange), true
	}

	if hasBefore && before.Text == "{" && hasAfter && after.Text == "," {
		return rule.RuleFixRemoveRange(core.NewTextRange(elementRange.Pos(), after.End)), true
	}
	if hasBefore && before.Text == "," {
		return rule.RuleFixRemoveRange(core.NewTextRange(before.Start, elementRange.End())), true
	}
	return rule.RuleFix{}, false
}

func fixImportBinding(ctx rule.RuleContext, definition *ast.Node, ac *analysisContext) (rule.RuleFix, bool) {
	importDeclaration := getImportDeclaration(definition)
	if importDeclaration == nil {
		return rule.RuleFix{}, false
	}
	importData := importDeclaration.AsImportDeclaration()
	if importData == nil || importData.ImportClause == nil || importData.ModuleSpecifier == nil {
		return rule.RuleFix{}, false
	}
	clause := importData.ImportClause.AsImportClause()
	if clause == nil {
		return rule.RuleFix{}, false
	}
	moduleRange := utils.TrimNodeTextRange(ctx.SourceFile, importData.ModuleSpecifier)

	switch definition.Kind {
	case ast.KindImportClause:
		name := clause.Name()
		if name == nil {
			return rule.RuleFix{}, false
		}
		nameRange := utils.TrimNodeTextRange(ctx.SourceFile, name)
		if clause.NamedBindings == nil {
			return rule.RuleFixRemoveRange(core.NewTextRange(nameRange.Pos(), moduleRange.Pos())), true
		}
		if comma, ok := tokenAfter(ac.tokens, nameRange.End(), 0); ok && comma.Text == "," {
			return rule.RuleFixRemoveRange(core.NewTextRange(nameRange.Pos(), comma.End)), true
		}

	case ast.KindNamespaceImport:
		namespaceRange := utils.TrimNodeTextRange(ctx.SourceFile, definition)
		if clause.Name() != nil {
			if comma, ok := tokenBefore(ac.tokens, namespaceRange.Pos(), 0); ok && comma.Text == "," {
				return rule.RuleFixRemoveRange(core.NewTextRange(comma.Start, namespaceRange.End())), true
			}
			return rule.RuleFix{}, false
		}
		return rule.RuleFixRemoveRange(core.NewTextRange(namespaceRange.Pos(), moduleRange.Pos())), true

	case ast.KindImportSpecifier:
		if clause.NamedBindings == nil || clause.NamedBindings.Kind != ast.KindNamedImports {
			return rule.RuleFix{}, false
		}
		namedImports := clause.NamedBindings.AsNamedImports()
		if namedImports == nil || namedImports.Elements == nil {
			return rule.RuleFix{}, false
		}
		specifiers := namedImports.Elements.Nodes
		if len(specifiers) == 1 {
			if clause.Name() == nil {
				return removeNodeRange(ctx.SourceFile, importDeclaration), true
			}
			namedRange := utils.TrimNodeTextRange(ctx.SourceFile, clause.NamedBindings)
			if comma, ok := tokenBefore(ac.tokens, namedRange.Pos(), 0); ok && comma.Text == "," {
				return rule.RuleFixRemoveRange(core.NewTextRange(comma.Start, namedRange.End())), true
			}
			return rule.RuleFix{}, false
		}

		specifierRange := utils.TrimNodeTextRange(ctx.SourceFile, definition)
		before, hasBefore := tokenBefore(ac.tokens, specifierRange.Pos(), 0)
		after, hasAfter := tokenAfter(ac.tokens, specifierRange.End(), 0)
		if hasBefore && before.Text == "{" && hasAfter && after.Text == "," {
			return rule.RuleFixRemoveRange(core.NewTextRange(specifierRange.Pos(), after.End)), true
		}
		if hasBefore && before.Text == "," {
			return rule.RuleFixRemoveRange(core.NewTextRange(before.Start, specifierRange.End())), true
		}
	}

	return rule.RuleFix{}, false
}

// fixCommaSeparatedName mirrors ESLint core's generic identifier suggestion
// for TS list bindings such as type parameters and enum members. It removes an
// adjacent comma with the name when possible, but deliberately leaves any
// initializer/constraint that follows the identifier untouched.
func fixCommaSeparatedName(ctx rule.RuleContext, nameNode *ast.Node, ac *analysisContext) rule.RuleFix {
	nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
	if before, ok := tokenBefore(ac.tokens, nameRange.Pos(), 0); ok && before.Text == "," {
		return rule.RuleFixRemoveRange(core.NewTextRange(before.Start, nameRange.End()))
	}
	if after, ok := tokenAfter(ac.tokens, nameRange.End(), 0); ok && after.Text == "," {
		return rule.RuleFixRemoveRange(core.NewTextRange(nameRange.Pos(), after.End))
	}
	return rule.RuleFixRemoveRange(nameRange)
}

func getCoreRemoveSuggestion(ctx rule.RuleContext, nameNode *ast.Node, definition *ast.Node, sym *ast.Symbol, ac *analysisContext) (rule.RuleSuggestion, bool) {
	if sym != nil && len(ac.writeRefs[sym]) > 0 {
		return rule.RuleSuggestion{}, false
	}
	if definition == nil {
		return rule.RuleSuggestion{}, false
	}

	var (
		fix rule.RuleFix
		ok  bool
	)
	switch definition.Kind {
	case ast.KindVariableDeclaration:
		fix, ok = fixVariableDeclaration(ctx, definition, ac)
	case ast.KindBindingElement:
		fix, ok = fixBindingElement(ctx, definition, ac)
	case ast.KindParameter:
		if ast.HasSyntacticModifier(definition, ast.ModifierFlagsParameterPropertyModifier) {
			// The TS parser exposes only the identifier plus its optional/type
			// suffix as this core rule's fixable variable range. Preserve that
			// behavior even though removing it can leave a modifier behind.
			fixRange, _ := eslintCoreDefinitionNameRange(ctx.SourceFile, nameNode, definition)
			fix, ok = rule.RuleFixRemoveRange(fixRange), true
		} else {
			fix, ok = fixFunctionParameter(ctx, definition, ac)
		}
	case ast.KindFunctionDeclaration:
		if definition.Body() == nil {
			// Overload/ambient function definitions use the function identifier
			// itself as ESLint's removal range, not the whole declaration.
			fix, ok = removeNodeRange(ctx.SourceFile, nameNode), true
		} else {
			fix, ok = removeNodeRange(ctx.SourceFile, definition), true
		}
	case ast.KindClassDeclaration:
		fix, ok = removeNodeRange(ctx.SourceFile, definition), true
	case ast.KindTypeParameter, ast.KindEnumMember:
		fix, ok = fixCommaSeparatedName(ctx, nameNode, ac), true
	case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindModuleDeclaration, ast.KindEnumDeclaration:
		fix, ok = removeNodeRange(ctx.SourceFile, nameNode), true
	case ast.KindImportClause, ast.KindImportSpecifier, ast.KindNamespaceImport:
		fix, ok = fixImportBinding(ctx, definition, ac)
	}
	if !ok {
		return rule.RuleSuggestion{}, false
	}

	name := ""
	if nameNode != nil && ast.IsIdentifier(nameNode) {
		name = nameNode.AsIdentifier().Text
	}
	return rule.RuleSuggestion{
		Message:  buildRemoveVarMessage(name),
		FixesArr: []rule.RuleFix{fix},
	}, true
}

// collectLocalExportTargets adds the exact local symbols consumed by one named
// export declaration. It deliberately ignores source-bearing re-exports: in
// `export { A } from "mod"`, A does not refer to an in-scope declaration.
func collectLocalExportTargets(typeChecker *checker.Checker, node *ast.Node, targets map[*ast.Symbol]bool) {
	if node == nil || node.Kind != ast.KindExportDeclaration {
		return
	}
	exportDecl := node.AsExportDeclaration()
	if exportDecl == nil || exportDecl.ExportClause == nil || !ast.IsNamedExports(exportDecl.ExportClause) {
		return
	}
	namedExports := exportDecl.ExportClause.AsNamedExports()
	if namedExports == nil || namedExports.Elements == nil {
		return
	}
	for _, spec := range namedExports.Elements.Nodes {
		if spec.AsExportSpecifier() == nil || utils.IsReExportSpecifier(spec) {
			continue
		}
		// The checker resolves aliases and lexical shadows for both
		// `export { A as B }` and `export type { A as B }`.
		if target := typeChecker.GetExportSpecifierLocalTargetSymbol(spec); target != nil {
			targets[target] = true
		}
	}
}

// collectSymbolUsages walks the entire source file AST and collects:
//   - usages: maps each symbol to its usage reference nodes (read references)
//   - writeRefs: maps each symbol to its write-only reference nodes (assignments)
//   - localExportTargets: local symbols consumed by named export declarations
//   - globalRefsByName: references that are not shadowed by a local declaration
//
// For the TypeScript flavor, the walk also marks implicit JSX factory usage.
func collectSymbolUsages(ctx rule.RuleContext, sourceFile *ast.Node, usages map[*ast.Symbol][]*ast.Node, writeRefs map[*ast.Symbol][]*ast.Node, localExportTargets map[*ast.Symbol]bool, globalRefsByName map[string][]*ast.Node, coreSemantics bool) {
	sf := sourceFile.AsSourceFile()
	addUsage := func(sym *ast.Symbol, node *ast.Node) {
		if sym == nil {
			return
		}
		usages[sym] = append(usages[sym], node)
		resolved := ctx.TypeChecker.SkipAlias(sym)
		if resolved != nil && resolved != sym {
			usages[resolved] = append(usages[resolved], node)
		}
	}

	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}
		collectLocalExportTargets(ctx.TypeChecker, node, localExportTargets)

		if ast.IsIdentifier(node) && !utils.IsNonReferenceIdentifier(node) {
			sym := utils.GetReferenceSymbol(node, ctx.TypeChecker)
			if coreSemantics && (sym == nil || !utils.IsSymbolDeclaredInFile(sym, sf)) {
				globalRefsByName[node.Text()] = append(globalRefsByName[node.Text()], node)
			}

			// Track write-only references separately for report position.
			// Simple assignments (=) are write-only and don't count as usage.
			if isPartOfAssignment(node) {
				writeSymbol := sym
				if !coreSemantics && node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
					// @typescript-eslint's scope manager associates object-pattern
					// shorthand writes with the property symbol for report-location
					// purposes. Core resolves the shorthand's value binding instead.
					writeSymbol = ctx.TypeChecker.GetSymbolAtLocation(node)
				}
				if writeSymbol != nil {
					writeRefs[writeSymbol] = append(writeRefs[writeSymbol], node)
				}
				node.ForEachChild(func(child *ast.Node) bool {
					walk(child)
					return false
				})
				return
			}
			// Compound assignments (+=, -=, etc.) and update expressions (++, --)
			// are both read and write. Track as writeRef for report position,
			// but don't return early — the node is still recorded as a usage below.
			if isCompoundAssignmentTarget(node) || isUpdateTarget(node) {
				if sym != nil {
					writeRefs[sym] = append(writeRefs[sym], node)
				}
			}
			addUsage(sym, node)
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sourceFile)

	// @typescript-eslint's scope manager marks the configured JSX factory as
	// used implicitly. Espree does not, so ESLint core still reports an otherwise
	// unused React import; preserve that distinction between the two flavors.
	if !coreSemantics {
		markJsxFactoryUsed(ctx, sourceFile, usages)
	}
}

// markJsxFactoryUsed checks if the source file contains JSX and, if so, marks
// the JSX factory (and fragment factory) imports as used. This runs for every
// jsx mode: TS never produces an AST identifier reference to the factory, so
// without this an `import React` whose only "use" is JSX would be falsely
// reported. Mirrors @typescript-eslint/parser, which treats the jsxPragma
// (default "React") as used whenever JSX is present, regardless of runtime.
func markJsxFactoryUsed(ctx rule.RuleContext, sourceFile *ast.Node, usages map[*ast.Symbol][]*ast.Node) {
	if sourceFile == nil || sourceFile.SubtreeFacts()&ast.SubtreeContainsJsx == 0 {
		return
	}
	firstJsx, firstFragment := findJsxNodes(sourceFile)
	if firstJsx == nil && firstFragment == nil {
		return
	}
	// The checker folds file-level @jsx pragmas, compiler options, and the
	// default React namespace into one public lookup. Use the source file as
	// the location so a fragment-specific pragma cannot replace the element
	// factory namespace when the file contains fragments only.
	factoryName := ctx.TypeChecker.GetJsxNamespace(sourceFile)
	refNode := firstJsx
	if refNode == nil {
		refNode = firstFragment
	}
	if factoryName != "" {
		markImportByNameAsUsed(ctx, sourceFile, factoryName, refNode, usages)
	}
	// JSX fragments additionally mark the fragment factory as used
	if firstFragment != nil {
		fragmentFactoryName := ctx.TypeChecker.GetJsxFragmentFactory(firstFragment)
		if fragmentFactoryName != "" {
			markImportByNameAsUsed(ctx, sourceFile, fragmentFactoryName, firstFragment, usages)
		}
	}
}

// markImportByNameAsUsed finds an import with the given name and adds refNode
// as a usage reference for its symbol. We use refNode (a JSX element/fragment)
// instead of the import's own name node because processVariable filters out
// usages where usage.Pos() == declaration.Pos() (self-reference filtering).
func markImportByNameAsUsed(ctx rule.RuleContext, sourceFile *ast.Node, name string, refNode *ast.Node, usages map[*ast.Symbol][]*ast.Node) {
	sourceFile.ForEachChild(func(child *ast.Node) bool {
		for _, binding := range utils.GetImportBindingNodes(child) {
			if binding != nil && binding.Text() == name {
				if sym := ctx.TypeChecker.GetSymbolAtLocation(binding); sym != nil {
					usages[sym] = append(usages[sym], refNode)
				}
				return true
			}
		}
		return false
	})
}

// findJsxNodes returns the first JSX element/self-closing node and the first
// JSX fragment node found in the AST. Either may be nil if not present.
func findJsxNodes(node *ast.Node) (firstJsx *ast.Node, firstFragment *ast.Node) {
	var walk func(*ast.Node)
	walk = func(n *ast.Node) {
		if n == nil || n.SubtreeFacts()&ast.SubtreeContainsJsx == 0 ||
			(firstJsx != nil && firstFragment != nil) {
			return
		}
		switch n.Kind {
		case ast.KindJsxElement, ast.KindJsxSelfClosingElement:
			if firstJsx == nil {
				firstJsx = n
			}
		case ast.KindJsxFragment:
			// Fragments also require the JSX factory (e.g., React.createElement),
			// so they count as a JSX element for factory-marking purposes.
			if firstJsx == nil {
				firstJsx = n
			}
			if firstFragment == nil {
				firstFragment = n
			}
		}
		n.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return firstJsx != nil && firstFragment != nil
		})
	}
	walk(node)
	return
}

// hasInlineGlobalUse applies the same read semantics as declared variables to
// `/* global name */` entries, which have no declaration node or checker symbol.
// Local shadows were removed while collecting globalRefsByName.
func hasInlineGlobalUse(ctx rule.RuleContext, name string, references []*ast.Node) bool {
	for _, reference := range references {
		if isPartOfAssignment(reference) {
			continue
		}
		if isSelfModifyingReference(reference, nil, name, ctx.TypeChecker, nil, ctx.SourceFile) {
			continue
		}
		return true
	}
	return false
}

// referenceFallbackBoundary scopes the name-index fallback to the declaration's
// actual var or lexical scope when checker symbol identity is unavailable.
func referenceFallbackBoundary(definition *ast.Node) *ast.Node {
	if definition == nil {
		return nil
	}
	root := ast.GetRootDeclaration(definition)
	if root != nil && root.Kind == ast.KindVariableDeclaration && root.Parent != nil &&
		root.Parent.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(root.Parent) {
		// `var` escapes an enclosing block/loop/catch into its variable scope.
		return utils.FindEnclosingScope(definition)
	}
	// let/const/using, block functions/classes, catch variables, and loop
	// bindings must keep the name-based checker fallback inside their lexical
	// scope. This prevents a later same-named export from consuming a shadow.
	return ast.GetEnclosingBlockScopeContainer(definition)
}

// coreAmbientModuleBoundary returns the lexical module/namespace scope used by
// ESLint core when @typescript-eslint/parser supplies the scopes. The checker
// can merge an augmentation member with an outside/global symbol, so core must
// ignore those outside references for this declaration.
func coreAmbientModuleBoundary(definition *ast.Node, flavor ruleFlavor) *ast.Node {
	if flavor.typescript || definition == nil || !utils.IsInAmbientContext(definition) {
		return nil
	}
	return ast.FindAncestorKind(definition, ast.KindModuleBlock)
}

func isCoreTypeOnlyParameter(definition *ast.Node, flavor ruleFlavor) bool {
	return !flavor.typescript && definition != nil && definition.Kind == ast.KindParameter &&
		isParameterInWithoutBodyDeclaration(definition)
}

func coreTypeDeclarationSelfReferenceCounts(definition *ast.Node, flavor ruleFlavor) bool {
	return !flavor.typescript && definition != nil &&
		(definition.Kind == ast.KindInterfaceDeclaration || definition.Kind == ast.KindTypeAliasDeclaration)
}

// isScriptGlobalDefinition models ESLint's global scope for vars:"local" and
// `/* exported */`. `var` uses its enclosing variable scope even when nested
// in a block or loop; lexical declarations use tsgo's block-scope container.
func isScriptGlobalDefinition(sourceFile *ast.SourceFile, definition *ast.Node) bool {
	if sourceFile == nil || definition == nil || ast.IsExternalModule(sourceFile) {
		return false
	}
	root := ast.GetRootDeclaration(definition)
	if root == nil {
		return false
	}
	if root.Kind == ast.KindVariableDeclaration && root.Parent != nil &&
		root.Parent.Kind == ast.KindVariableDeclarationList && utils.IsVarKeyword(root.Parent) {
		return utils.FindEnclosingScope(root) == sourceFile.AsNode()
	}
	return ast.GetEnclosingBlockScopeContainer(root) == sourceFile.AsNode()
}

// processVariable determines whether a single declared variable/parameter/import
// is unused and, if so, reports it. The decision pipeline:
//  1. Resolve the symbol and look up usages (original sym → SkipAlias → shared reference index)
//  2. Filter out self-references (same position, self-modifying, inside own declaration body)
//  3. Classify remaining usages as value or type-only
//  4. Apply global/directive and value-vs-type semantics
//  5. Apply ignore patterns (varsIgnorePattern, argsIgnorePattern, etc.)
//  6. Skip exports, "after-used" parameters, and option-specific suppressions
//  7. Report at the last write-reference position (or declaration name as fallback)
func processVariable(ctx rule.RuleContext, nameNode *ast.Node, name string, definition *ast.Node, opts Config, ac *analysisContext, flavor ruleFlavor) {
	varInfo := &VariableInfo{
		Variable:       nameNode,
		Used:           false,
		OnlyUsedAsType: false,
		References:     []*ast.Node{},
		Definition:     definition,
	}

	sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
	// For declaration merging (interface + const, etc.), only process once.
	if sym != nil && len(sym.Declarations) > 1 {
		if ac.seenMergedSymbols[sym] {
			return
		}
		ac.seenMergedSymbols[sym] = true
	}
	coreTypeOnlyParameter := isCoreTypeOnlyParameter(definition, flavor)
	coreTypeSelfReference := coreTypeDeclarationSelfReferenceCounts(definition, flavor)
	var coreTypeOnlyBoundary *ast.Node
	if coreTypeOnlyParameter {
		// A signature parameter can be referenced by a type predicate in the
		// same signature, but an identically named value outside that signature
		// belongs to another ESLint scope.
		coreTypeOnlyBoundary = definition.Parent
	}
	coreModuleBoundary := coreAmbientModuleBoundary(definition, flavor)
	if sym != nil {
		// Look up usages by original symbol first. For import specifiers from
		// resolvable modules, SkipAlias collapses all specifiers into the same
		// module symbol, so we must NOT fall back to the resolved symbol for imports.
		usageNodes, exists := ac.allUsages[sym]
		isImportDef := definition != nil && (definition.Kind == ast.KindImportSpecifier ||
			definition.Kind == ast.KindImportClause ||
			definition.Kind == ast.KindNamespaceImport ||
			definition.Kind == ast.KindImportEqualsDeclaration)
		if !exists && !isImportDef {
			resolved := ctx.TypeChecker.SkipAlias(sym)
			if resolved != sym {
				usageNodes, exists = ac.allUsages[resolved]
			}
		}
		// Checker edge cases (notably empty namespaces) can lose symbol
		// identity. Reuse the shared, shadow-aware name index as a narrow
		// fallback and keep references inside the declaration's own container.
		if !exists && !isImportDef && !coreTypeOnlyParameter && ac.refIndex != nil {
			boundary := referenceFallbackBoundary(definition)
			if coreModuleBoundary != nil {
				boundary = coreModuleBoundary
			}
			ac.refIndex.ForEachReferenceByName(name, boundary, func(ref *ast.Node) bool {
				if boundary != nil && !ast.IsNodeDescendantOf(ref, boundary) {
					return false
				}
				if ref.Pos() != nameNode.Pos() && !isPartOfAssignment(ref) &&
					!isInsideOwnDeclaration(ref, definition) {
					usageNodes = append(usageNodes, ref)
					exists = true
				}
				return false
			})
		}
		// For declaration merging (e.g., multiple interfaces with same name),
		// collect all declarations so self-references in ANY declaration body
		// are correctly filtered out.
		allDecls := []*ast.Node{definition}
		if len(sym.Declarations) > 1 {
			allDecls = sym.Declarations
		}
		if exists {
			varInfo.References = usageNodes

			filteredUsages := []*ast.Node{}
			for _, usage := range usageNodes {
				if (coreModuleBoundary == nil || ast.IsNodeDescendantOf(usage, coreModuleBoundary)) &&
					(coreTypeOnlyBoundary == nil || ast.IsNodeDescendantOf(usage, coreTypeOnlyBoundary)) &&
					usage.Pos() != varInfo.Variable.Pos() &&
					!isSelfModifyingReference(usage, sym, name, ctx.TypeChecker, nameNode, ctx.SourceFile) &&
					(coreTypeSelfReference || !isInsideAnyOwnDeclaration(usage, allDecls)) {
					filteredUsages = append(filteredUsages, usage)
				}
			}

			if len(filteredUsages) > 0 {
				onlyUsedAsType := true
				for _, usage := range filteredUsages {
					if !utils.IsIdentifierInTypeReference(usage) {
						onlyUsedAsType = false
						break
					}
				}
				varInfo.Used = !onlyUsedAsType
				varInfo.OnlyUsedAsType = onlyUsedAsType
			}
		}
	}

	scriptGlobal := isScriptGlobalDefinition(ctx.SourceFile, definition)
	if scriptGlobal && ac.exportedNames[name] {
		// ESLint marks exported-directive globals as used before this rule runs.
		// Preserve that ordering so reportUsedIgnorePattern can still diagnose a
		// marked name that matches an ignore pattern.
		varInfo.Used = true
		varInfo.OnlyUsedAsType = false
	}

	// vars: "local" skips only the script global scope. ES module top-level
	// bindings live in a module scope and must still be checked.
	if opts.Vars == "local" && scriptGlobal {
		return
	}

	// For type-level declarations and imports, being used in a type context
	// IS valid usage — don't report "only used as type".
	// This must happen BEFORE matchesIgnorePattern so reportUsedIgnorePattern works correctly.
	isTypeOrImportDeclaration := definition != nil && (definition.Kind == ast.KindInterfaceDeclaration ||
		definition.Kind == ast.KindTypeAliasDeclaration ||
		definition.Kind == ast.KindEnumDeclaration ||
		definition.Kind == ast.KindTypeParameter ||
		definition.Kind == ast.KindImportSpecifier ||
		definition.Kind == ast.KindImportClause ||
		definition.Kind == ast.KindNamespaceImport ||
		definition.Kind == ast.KindImportEqualsDeclaration)
	if isTypeOrImportDeclaration && varInfo.OnlyUsedAsType {
		varInfo.Used = true
		varInfo.OnlyUsedAsType = false
	}
	if !flavor.typescript && varInfo.OnlyUsedAsType {
		// TypeScript's scope manager presents type references to ESLint's core
		// rule as uses. Preserve that base-rule behavior; the TypeScript flavor
		// retains its more specific "only used as a type" diagnostic.
		varInfo.Used = true
		varInfo.OnlyUsedAsType = false
	}
	// Check ignore patterns (varsIgnorePattern / argsIgnorePattern / caughtErrorsIgnorePattern).
	// If the variable matches its category's pattern and is unused → ignore silently.
	// If it matches but IS used and reportUsedIgnorePattern is true → report as usedIgnoredVar.
	shouldIgnore, matchedPattern, matchedType := matchesIgnorePattern(name, varInfo, opts, ac.writeRefs, sym)
	if shouldIgnore {
		return
	}

	if isExported(ctx, varInfo, ac.localExportTargets, flavor) {
		// Even exported variables should be reported if they match an ignore pattern
		// and reportUsedIgnorePattern is true (e.g., `export const x = _Foo`).
		if matchedPattern && varInfo.Used && opts.ReportUsedIgnorePattern {
			reportNode := varInfo.Variable
			if sym != nil {
				if refs, exists := ac.writeRefs[sym]; exists && len(refs) > 0 {
					if lastWrite := lastWriteToReport(definition, refs, flavor); lastWrite != nil {
						reportNode = lastWrite
					}
				}
			}
			additional := ""
			if !flavor.typescript {
				additional = ignorePatternAdditional(matchedType, opts, true)
			}
			reportVariableDiagnostic(ctx, ac.reporter, nameNode, reportNode, definition, flavor, buildUsedIgnoredVarMessage(name, additional))
		}
		return
	}

	// "after-used" for parameters: skip unused params before the last used param.
	// Only applies to direct Parameter nodes, not destructured elements within them.
	if !varInfo.Used && !coreTypeOnlyParameter && definition != nil && definition.Kind == ast.KindParameter && opts.Args == "after-used" {
		param := definition.AsParameterDeclaration()
		if param != nil && param.Initializer == nil && isBeforeLastUsedParam(ctx, definition, ac.allUsages) {
			return
		}
	}

	// These options suppress only otherwise-unused bindings. Keep them after
	// ignore-pattern handling so reportUsedIgnorePattern can still report a
	// used `using` binding or a used object-rest sibling.
	if !varInfo.Used {
		if opts.IgnoreUsingDeclarations && isUsingDeclaration(definition) {
			return
		}
		if opts.IgnoreRestSiblings &&
			(hasObjectRestSiblingDefinition(definition) || hasObjectRestSiblingWrite(ac.writeRefs, sym)) {
			return
		}
	}

	// ESLint reports at the last write reference position (e.g., `a = a + 1` reports
	// at the LHS `a`). Fall back to the declaration name node if no write refs found.
	reportNode := varInfo.Variable
	if sym != nil {
		if refs, exists := ac.writeRefs[sym]; exists && len(refs) > 0 {
			if lastWrite := lastWriteToReport(definition, refs, flavor); lastWrite != nil {
				reportNode = lastWrite
			}
		}
	}

	assigned := hasAssignment(definition, sym, ac.writeRefs, flavor)
	unusedAdditional := ""
	usedIgnoredAdditional := ""
	if !flavor.typescript {
		unusedAdditional = ignorePatternAdditional(definitionVariableType(definition, opts), opts, false)
		usedIgnoredAdditional = ignorePatternAdditional(matchedType, opts, true)
	}

	if matchedPattern && varInfo.Used && opts.ReportUsedIgnorePattern {
		reportVariableDiagnostic(ctx, ac.reporter, nameNode, varInfo.Variable, definition, flavor, buildUsedIgnoredVarMessage(name, usedIgnoredAdditional))
	} else if flavor.typescript && varInfo.OnlyUsedAsType && opts.Vars == "all" {
		ac.reporter.reportNode(reportNode, buildUsedOnlyAsTypeMessage(name, assigned))
	} else if !varInfo.Used {
		isImport := definition != nil && (definition.Kind == ast.KindImportSpecifier ||
			definition.Kind == ast.KindImportClause ||
			definition.Kind == ast.KindNamespaceImport ||
			definition.Kind == ast.KindImportEqualsDeclaration)

		// Track unused imports for allImportSpecifiersUnused check
		if isImport {
			ac.reportedUnused[definition] = true
		}

		// Generate import removal fix/suggestion
		if isImport && flavor.typescript {
			fixes, suggestionMsg := getImportRemoveFix(ctx, definition, ac.reportedUnused)
			if len(fixes) > 0 {
				rule.ReportNodeWithFixesOrSuggestions(ctx, reportNode, opts.EnableAutofixRemoval.Imports, buildUnusedVarMessage(name, assigned, unusedAdditional), suggestionMsg, fixes...)
				return
			}
		}
		message := buildUnusedVarMessage(name, assigned, unusedAdditional)
		if flavor.coreSuggestions {
			if suggestion, ok := getCoreRemoveSuggestion(ctx, nameNode, definition, sym, ac); ok {
				reportVariableDiagnostic(ctx, ac.reporter, nameNode, reportNode, definition, flavor, message, suggestion)
				return
			}
		}
		reportVariableDiagnostic(ctx, ac.reporter, nameNode, reportNode, definition, flavor, message)
	}
}

func newRule(flavor ruleFlavor) rule.Rule {
	schema := (*rule.Schema)(nil)
	if !flavor.typescript {
		schema = rule.NewSchema(schemaJSON)
	}

	return rule.Rule{
		Name:                "no-unused-vars",
		RequiresBindingInfo: !flavor.typescript,
		RequiresTypeInfo:    flavor.typescript,
		Schema:              schema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
			if !flavor.typescript && ctx.TypeChecker == nil {
				// Core needs lexical symbols, not trusted project type information.
				// Gap/inferred files receive the linter's binding-only checker.
				ctx.TypeChecker = ctx.BindingChecker
			}
			if ctx.TypeChecker == nil {
				return rule.RuleListeners{}
			}
			opts := parseOptions(options)
			reporter := &diagnosticReporter{
				ctx:          ctx,
				deferReports: !flavor.typescript,
			}
			var tokens []utils.SourceToken
			if flavor.coreSuggestions {
				tokens = utils.TokensOfNode(ctx.SourceFile, ctx.SourceFile.AsNode())
			}

			ac := &analysisContext{
				allUsages:          make(map[*ast.Symbol][]*ast.Node),
				writeRefs:          make(map[*ast.Symbol][]*ast.Node),
				localExportTargets: make(map[*ast.Symbol]bool),
				globalRefsByName:   make(map[string][]*ast.Node),
				exportedNames:      rule.ParseExportedNames(ctx.SourceFile, ctx.Comments),
				refIndex:           utils.NewReferenceIndex(ctx.SourceFile, ctx.TypeChecker),
				seenMergedSymbols:  make(map[*ast.Symbol]bool),
				reportedUnused:     make(map[*ast.Node]bool),
				reporter:           reporter,
				tokens:             tokens,
			}
			collected := false

			seenWithoutBodyFuncSymbols := make(map[*ast.Symbol]bool)

			ensureCollected := func(node *ast.Node) {
				if !collected {
					sourceFile := ast.GetSourceFileOfNode(node)
					collectSymbolUsages(ctx, sourceFile.AsNode(), ac.allUsages, ac.writeRefs, ac.localExportTargets, ac.globalRefsByName, !flavor.typescript)
					collected = true
				}
			}

			// processBindingName handles both simple identifiers and destructuring patterns.
			var processBindingName func(nameNode *ast.Node, definition *ast.Node)
			processBindingName = func(nameNode *ast.Node, definition *ast.Node) {
				if nameNode == nil {
					return
				}
				if ast.IsIdentifier(nameNode) {
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(definition)
					processVariable(ctx, nameNode, identifier.Text, definition, opts, ac, flavor)
				} else if nameNode.Kind == ast.KindObjectBindingPattern || nameNode.Kind == ast.KindArrayBindingPattern {
					nameNode.ForEachChild(func(child *ast.Node) bool {
						if child.Kind == ast.KindBindingElement {
							elem := child.AsBindingElement()
							if elem != nil && elem.Name() != nil {
								processBindingName(elem.Name(), child)
							}
						}
						return false
					})
				}
			}

			listeners := rule.RuleListeners{
				ast.KindVariableDeclaration: func(node *ast.Node) {
					varDecl := node.AsVariableDeclaration()
					if varDecl == nil {
						return
					}
					if flavor.typescript && (isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node)) {
						return
					}
					// Skip for-in/for-of declarations whose body starts with return.
					// E.g., `for (var name in obj) { return true; }` — the variable
					// is considered "used" (checking property existence).
					if forStmt := isForInOfDeclaration(node); forStmt != nil {
						body := forStmt.AsForInOrOfStatement().Statement
						if ast.IsIdentifier(varDecl.Name()) && forInBodyStartsWithReturn(body) {
							return
						}
					}
					processBindingName(varDecl.Name(), node)
				},

				ast.KindFunctionDeclaration: func(node *ast.Node) {
					funcDecl := node.AsFunctionDeclaration()
					if funcDecl == nil {
						return
					}
					if funcDecl.Name() == nil || !ast.IsIdentifier(funcDecl.Name()) {
						return
					}

					nameNode := funcDecl.Name()
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}

					ensureCollected(node)

					if node.Body() == nil {
						if flavor.typescript && (isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node)) {
							return
						}
						sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
						if sym != nil {
							resolved := ctx.TypeChecker.SkipAlias(sym)
							if seenWithoutBodyFuncSymbols[resolved] {
								return
							}
							seenWithoutBodyFuncSymbols[resolved] = true
						}
					}

					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindModuleDeclaration: func(node *ast.Node) {
					moduleDecl := node.AsModuleDeclaration()
					if moduleDecl == nil {
						return
					}

					// Skip global scope augmentations: `declare global { ... }`
					// Also skip any namespace inside `declare global` — they're global type augmentations.
					// `global` in `declare global { ... }` names the augmentation
					// construct; it is not a variable in either ESLint scope model.
					if ast.IsGlobalScopeAugmentation(node) {
						return
					}
					if flavor.typescript && ast.FindAncestor(node.Parent, func(n *ast.Node) bool { return ast.IsGlobalScopeAugmentation(n) }) != nil {
						return
					}

					nameNode := moduleDecl.Name()
					if nameNode == nil {
						return
					}
					// Skip module augmentations: `declare module 'foo' { ... }`
					if nameNode.Kind == ast.KindStringLiteral {
						return
					}
					if !ast.IsIdentifier(nameNode) {
						return
					}
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}

					// Skip dotted namespace declarations like `namespace foo.bar` —
					// the outer `foo` is just a container, not a standalone declaration.
					if node.Body() != nil && node.Body().Kind == ast.KindModuleDeclaration {
						return
					}

					// Skip inner namespaces inside ambient (declare) namespace declarations —
					// they're part of the type definition, not standalone declarations.
					if flavor.typescript && (isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node)) {
						return
					}

					// Skip namespace augmentations — if the namespace symbol has
					// declarations outside this file, it's augmenting an existing
					// namespace (e.g., `declare namespace NodeJS { ... }`).
					sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
					if flavor.typescript && sym != nil && len(sym.Declarations) > 1 {
						for _, decl := range sym.Declarations {
							if decl != node {
								return
							}
						}
					}

					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindClassDeclaration: func(node *ast.Node) {
					classDecl := node.AsClassDeclaration()
					if classDecl == nil || classDecl.Name() == nil || !ast.IsIdentifier(classDecl.Name()) {
						return
					}
					if flavor.typescript && (isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node)) {
						return
					}
					if opts.IgnoreClassWithStaticInitBlock && hasStaticInitBlock(node) {
						return
					}
					nameNode := classDecl.Name()
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindInterfaceDeclaration: func(node *ast.Node) {
					interfaceDecl := node.AsInterfaceDeclaration()
					if interfaceDecl == nil || interfaceDecl.Name() == nil || !ast.IsIdentifier(interfaceDecl.Name()) {
						return
					}
					if flavor.typescript && (isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node)) {
						return
					}
					nameNode := interfaceDecl.Name()
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindTypeAliasDeclaration: func(node *ast.Node) {
					typeAlias := node.AsTypeAliasDeclaration()
					if typeAlias == nil || typeAlias.Name() == nil || !ast.IsIdentifier(typeAlias.Name()) {
						return
					}
					if flavor.typescript && (isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node)) {
						return
					}
					nameNode := typeAlias.Name()
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindEnumDeclaration: func(node *ast.Node) {
					enumDecl := node.AsEnumDeclaration()
					if enumDecl == nil || enumDecl.Name() == nil || !ast.IsIdentifier(enumDecl.Name()) {
						return
					}
					if flavor.typescript && (isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node)) {
						return
					}
					nameNode := enumDecl.Name()
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindParameter: func(node *ast.Node) {
					paramDecl := node.AsParameterDeclaration()
					if paramDecl == nil {
						return
					}
					// JSDoc function-type names live in comments and are not variables
					// exposed by @typescript-eslint/parser's ESTree scope graph.
					if node.Flags&ast.NodeFlagsJSDoc != 0 {
						return
					}
					// Index-signature keys are property placeholders, not variables in
					// either ESLint scope model.
					if node.Parent != nil && node.Parent.Kind == ast.KindIndexSignature {
						return
					}

					if flavor.typescript && isParameterInWithoutBodyDeclaration(node) {
						return
					}
					// ESLint skips setter arguments because setters require exactly one
					// parameter even when the body does not read it.
					if node.Parent != nil && node.Parent.Kind == ast.KindSetAccessor {
						return
					}

					// Skip TypeScript's `this` parameter (type annotation only, not a real param).
					// In tsgo, the `this` parameter name is parsed as an Identifier with text "this".
					if flavor.typescript && paramDecl.Name() != nil &&
						(paramDecl.Name().Kind == ast.KindThisKeyword ||
							(ast.IsIdentifier(paramDecl.Name()) && paramDecl.Name().AsIdentifier().Text == "this")) {
						return
					}

					// Skip constructor parameter properties (private/protected/public/readonly params).
					// These are promoted to class fields and are inherently "used".
					if flavor.typescript && ast.HasSyntacticModifier(node, ast.ModifierFlagsParameterPropertyModifier) {
						return
					}

					if nameNode := paramDecl.Name(); nameNode != nil {
						if !flavor.typescript && nameNode.Kind == ast.KindThisKeyword {
							ensureCollected(node)
							processVariable(ctx, nameNode, "this", node, opts, ac, flavor)
							return
						}
						processBindingName(nameNode, node)
					}
				},

				// Note: catch clause variables are processed by KindVariableDeclaration above,
				// since CatchClause.VariableDeclaration is a VariableDeclaration node.
				// isCaughtErrorNode() detects them by checking for a CatchClause ancestor.
				ast.KindImportSpecifier: func(node *ast.Node) {
					importSpec := node.AsImportSpecifier()
					if importSpec == nil {
						return
					}
					nameNode := importSpec.Name()
					if nameNode == nil || !ast.IsIdentifier(nameNode) {
						return
					}
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindImportClause: func(node *ast.Node) {
					// Default import: `import Foo from './foo'`
					importClause := node.AsImportClause()
					if importClause == nil {
						return
					}
					nameNode := importClause.Name()
					if nameNode == nil || !ast.IsIdentifier(nameNode) {
						return
					}
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindNamespaceImport: func(node *ast.Node) {
					// Namespace import: `import * as ns from './foo'`
					nsImport := node.AsNamespaceImport()
					if nsImport == nil {
						return
					}
					nameNode := nsImport.Name()
					if nameNode == nil || !ast.IsIdentifier(nameNode) {
						return
					}
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindImportEqualsDeclaration: func(node *ast.Node) {
					// `import X = require('foo')` or `import X = Namespace.Y`
					importEquals := node.AsImportEqualsDeclaration()
					if importEquals == nil {
						return
					}
					nameNode := importEquals.Name()
					if nameNode == nil || !ast.IsIdentifier(nameNode) {
						return
					}
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindTypeParameter: func(node *ast.Node) {
					// JSDoc @template names are comment metadata, not ESTree variables.
					if node.Flags&ast.NodeFlagsJSDoc != 0 {
						return
					}
					// Generic type parameter declarations: `<T>`, `<T = unknown>`, `<T extends U>`.
					// Skip nodes that syntactically share KindTypeParameter in tsgo but aren't
					// parameter declarations: `infer T`, mapped-type `[P in K]`, JSDoc @template.
					parent := node.Parent
					if flavor.typescript && parent != nil {
						switch parent.Kind {
						case ast.KindInferType, ast.KindMappedType, ast.KindJSDocTemplateTag:
							return
						}
					}
					if flavor.typescript && (isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node)) {
						return
					}
					typeParam := node.AsTypeParameterDeclaration()
					if typeParam == nil {
						return
					}
					nameNode := typeParam.Name()
					if nameNode == nil || !ast.IsIdentifier(nameNode) {
						return
					}
					identifier := nameNode.AsIdentifier()
					if identifier == nil {
						return
					}
					ensureCollected(node)
					processVariable(ctx, nameNode, identifier.Text, node, opts, ac, flavor)
				},

				ast.KindEnumMember: func(node *ast.Node) {
					if flavor.typescript {
						return
					}
					member := node.AsEnumMember()
					if member == nil || member.Name() == nil || !ast.IsIdentifier(member.Name()) {
						return
					}
					nameNode := member.Name()
					ensureCollected(node)
					processVariable(ctx, nameNode, nameNode.Text(), node, opts, ac, flavor)
				},
			}

			if !flavor.typescript {
				ensureCollected(ctx.SourceFile.AsNode())
				if opts.Vars != "local" {
					for _, inlineGlobal := range ctx.InlineGlobals {
						if !inlineGlobal.Declared || len(inlineGlobal.NameRanges) == 0 {
							continue
						}
						if ac.exportedNames[inlineGlobal.Name] {
							continue
						}
						if hasInlineGlobalUse(ctx, inlineGlobal.Name, ac.globalRefsByName[inlineGlobal.Name]) {
							continue
						}
						reporter.reportRange(
							inlineGlobal.NameRanges[0],
							buildUnusedVarMessage(inlineGlobal.Name, false, ""),
						)
					}
				}

				statements := ctx.SourceFile.Statements
				if statements == nil || len(statements.Nodes) == 0 {
					reporter.flush()
				} else {
					lastTopLevelNode := statements.Nodes[len(statements.Nodes)-1]
					listeners[rule.ListenerOnExit(lastTopLevelNode.Kind)] = func(node *ast.Node) {
						if node == lastTopLevelNode {
							reporter.flush()
						}
					}
				}
			}

			return listeners
		},
	}
}

// NoUnusedVarsRule implements ESLint core's no-unused-vars rule.
var NoUnusedVarsRule = newRule(ruleFlavor{coreSuggestions: true})

// NewTypeScriptRule returns the unprefixed base rule used by the
// @typescript-eslint wrapper.
func NewTypeScriptRule() rule.Rule {
	return newRule(ruleFlavor{typescript: true})
}
