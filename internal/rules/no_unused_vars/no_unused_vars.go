package no_unused_vars

import (
	_ "embed"
	"fmt"
	"sort"
	"strings"

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
	allUsages         map[*ast.Symbol][]*ast.Node
	writeRefs         map[*ast.Symbol][]*ast.Node
	unresolvedRefs    map[string][]*ast.Node
	referencesByName  map[string][]*ast.Node
	seenMergedSymbols map[*ast.Symbol]bool
	reportedUnused    map[*ast.Node]bool
	reporter          *diagnosticReporter
	tokens            []utils.SourceToken
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

// isInTypeContext checks if an identifier is inside a type-only position
// (type reference, type alias body, interface body, etc.). Used to detect
// variables that are "only used as a type" — runtime values referenced solely
// in type annotations should be reported with a specific message.
func isInTypeContext(node *ast.Node) bool {
	parent := node.Parent
	for parent != nil {
		switch parent.Kind {
		case ast.KindTypeReference,
			ast.KindTypeAliasDeclaration,
			ast.KindInterfaceDeclaration,
			ast.KindTypeParameter,
			ast.KindTypeQuery,
			ast.KindTypeOperator,
			ast.KindIndexedAccessType,
			ast.KindConditionalType,
			ast.KindInferType,
			ast.KindTypeLiteral,
			ast.KindMappedType:
			return true
			// Note: KindAsExpression, KindTypeAssertionExpression, KindSatisfiesExpression
			// are NOT included here. Their expression operand is a value context;
			// only the type annotation part is a type context. Since we walk up
			// from the identifier, a value operand will pass through these nodes
			// and continue upward without being misclassified as type-only.
		}
		parent = parent.Parent
	}
	return false
}

// isDeclarationName delegates to utils.IsDeclarationIdentifier.
func isDeclarationName(node *ast.Node) bool {
	return utils.IsDeclarationIdentifier(node)
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
		if forStmt != nil && forStmt.Statement != nil && forInBodyStartsWithReturn(forStmt.Statement) {
			return false // Not write-only — the variable is meaningfully used
		}
		return true
	}
	return false
}

// isUpdateTarget checks if the identifier is the operand of a prefix/postfix
// increment or decrement (++x, x++, --x, x--).
func isUpdateTarget(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	if parent.Kind == ast.KindPrefixUnaryExpression {
		op := parent.AsPrefixUnaryExpression().Operator
		return op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken
	}
	if parent.Kind == ast.KindPostfixUnaryExpression {
		op := parent.AsPostfixUnaryExpression().Operator
		return op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken
	}
	return false
}

// isCompoundAssignmentTarget checks if the identifier is the LHS of a compound
// assignment (+=, -=, *=, etc.) but NOT a simple = or logical assignment.
func isCompoundAssignmentTarget(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	if parent.Kind != ast.KindBinaryExpression {
		return false
	}
	bin := parent.AsBinaryExpression()
	if bin == nil || bin.Left != node {
		return false
	}
	op := bin.OperatorToken.Kind
	return ast.IsCompoundAssignment(op) && op != ast.KindEqualsToken
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
		switch current.Kind {
		case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement,
			ast.KindWhileStatement, ast.KindDoStatement:
			return true
		}
	}
	return false
}

// nearestVariableScope returns the nearest enclosing function-like node containing
// node, or nil if node sits at the top level (module/global scope). Blocks don't
// introduce a new variable scope, matching escope's notion of a "variable scope".
func nearestVariableScope(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLike(current) {
			return current
		}
	}
	return nil
}

func lastWriteInDeclarationScope(definition *ast.Node, refs []*ast.Node) *ast.Node {
	declarationScope := nearestVariableScope(definition)
	for index := len(refs) - 1; index >= 0; index-- {
		if nearestVariableScope(refs[index]) == declarationScope {
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
func isSelfModifyingReference(node *ast.Node, sym *ast.Symbol, checker *checker.Checker, declNode *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent

	// Case 1: a++ or a-- (update expression as a statement)
	if parent.Kind == ast.KindPrefixUnaryExpression {
		if op := parent.AsPrefixUnaryExpression().Operator; op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken {
			return isUnusedExpression(parent)
		}
	}
	if parent.Kind == ast.KindPostfixUnaryExpression {
		if op := parent.AsPostfixUnaryExpression().Operator; op == ast.KindPlusPlusToken || op == ast.KindMinusMinusToken {
			return isUnusedExpression(parent)
		}
	}

	// Case 2: a += expr (compound assignment — the LHS identifier is both read and written).
	// Logical assignments (??=, &&=, ||=) are NOT self-modifying because they conditionally
	// assign and ESLint considers them as meaningful usage.
	if parent.Kind == ast.KindBinaryExpression {
		bin := parent.AsBinaryExpression()
		if bin != nil && ast.IsCompoundAssignment(bin.OperatorToken.Kind) && bin.Left == node {
			op := bin.OperatorToken.Kind
			if op == ast.KindBarBarEqualsToken || op == ast.KindAmpersandAmpersandEqualsToken || op == ast.KindQuestionQuestionEqualsToken {
				return false // Logical assignment — not self-modifying
			}
			return isUnusedExpression(parent)
		}
	}

	// Case 3: cb = (function(a) { return cb(1+a); })() — identifier inside IIFE body
	// that's assigned back to the same variable. Walk up from the identifier to find
	// if it's inside a function whose call result is assigned to the same variable.
	if isInsideFunctionAssignedToSelf(node, sym, checker) {
		return true
	}

	// Case 4: a = <expr containing a> (identifier appears in the RHS of an assignment
	// to the same variable). Covers: `a = a + 1`, `a = a.filter(...)`, `a = a.concat(a)`.
	// Walk up through expressions, but only when the identifier is the "subject" of the
	// expression chain (e.g., `a.method()` where `a` is the object). Stop when the
	// identifier is used as an argument to another function (e.g., `f(a)` or `setTimeout(fn, a)`).
	current := node
	for current.Parent != nil {
		p := current.Parent
		if p.Kind == ast.KindBinaryExpression {
			bin := p.AsBinaryExpression()
			if bin != nil && ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
				lhsSym := checker.GetSymbolAtLocation(bin.Left)
				if lhsSym == sym {
					if declNode != nil && (nearestVariableScope(p) != nearestVariableScope(declNode) || isInsideLoop(p)) {
						return false
					}
					return isUnusedExpression(p)
				}
				break
			}
			// Arithmetic/logic: keep walking
			current = p
			continue
		}
		if p.Kind == ast.KindParenthesizedExpression {
			current = p
			continue
		}
		// PropertyAccessExpression: a.method — walk up only if `current` is the object (a)
		if p.Kind == ast.KindPropertyAccessExpression {
			pae := p.AsPropertyAccessExpression()
			if pae != nil && pae.Expression == current {
				current = p
				continue
			}
			break
		}
		// CallExpression: a.method(...) or a.method(a)
		if p.Kind == ast.KindCallExpression {
			ce := p.AsCallExpression()
			if ce != nil {
				if ce.Expression == current {
					// current is the callee (a.method) → walk up
					current = p
					continue
				}
				// current is an argument. If the callee is a method on the same variable
				// (e.g., x.concat(x)), continue walking — the value is still consumed
				// by the same variable's method chain. Otherwise break.
				if isMethodCallOnSameSymbol(ce.Expression, sym, checker) {
					current = p
					continue
				}
				// current is an argument to a function call (e.g., x in foo(x)).
				// If the identifier is NOT inside a "storable function" (a non-IIFE function
				// expression that could be captured as a callback), check if the call result
				// is assigned to the same variable (e.g., x = foo(x)).
				if !isInsideStorableFunction(node, p) {
					current = p
					continue
				}
			}
			break
		}
		// Stop at other expression types (conditionals, template literals, etc.)
		break
	}

	return false
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
		case ast.KindModuleDeclaration, ast.KindFunctionDeclaration:
			body = definition.Body()
		case ast.KindInterfaceDeclaration, ast.KindTypeAliasDeclaration, ast.KindEnumDeclaration:
			// Self-referencing types/enums: `interface Foo { baz: Foo['bar'] }`,
			// `type Foo = { bar: Foo }`, `enum Foo { B = Foo.A }`.
			body = definition
		case ast.KindVariableDeclaration:
			// For `var a = function() { a() }` or `const a = () => { a() }`,
			// the definition is the VariableDeclaration. The initializer is the
			// function expression whose body contains the self-reference.
			varDecl := definition.AsVariableDeclaration()
			if varDecl != nil && varDecl.Initializer != nil {
				init := varDecl.Initializer
				// The initializer could be a function expression or arrow function
				body = init.Body()
			}
		default:
			continue
		}
		if body == nil {
			continue
		}
		current := usage
		for current != nil {
			if current == body {
				return true
			}
			current = current.Parent
		}
	}
	return false
}

// isInsideStorableFunction checks if the identifier `node` is inside a function
// expression/arrow function between `node` and `boundary` that is NOT an IIFE.
// Such a function could be stored as a callback and called later, so its reference
// to the variable is not necessarily self-modifying.
// Example of storable: `_timer = setTimeout(function(){ clearInterval(_timer) }, ...)`
// Example of non-storable (IIFE): `cb = (function(a){ return cb(1+a) })()`
func isInsideStorableFunction(node *ast.Node, boundary *ast.Node) bool {
	current := node.Parent
	for current != nil && current != boundary {
		if ast.IsFunctionLike(current) {
			// Check if this function is an IIFE (immediately invoked)
			if ast.GetImmediatelyInvokedFunctionExpression(current) != nil {
				// IIFE — not storable, continue walking up
				current = current.Parent
				continue
			}
			// Non-IIFE function expression — it's storable
			return true
		}
		current = current.Parent
	}
	return false
}

// isMethodCallOnSameSymbol checks if the callee expression is a method call
// on the same variable (e.g., `x.concat` where `x` is the symbol).
func isMethodCallOnSameSymbol(callee *ast.Node, sym *ast.Symbol, checker *checker.Checker) bool {
	if callee == nil || callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	obj := callee.AsPropertyAccessExpression().Expression
	if obj == nil {
		return false
	}
	objSym := checker.GetSymbolAtLocation(obj)
	return objSym == sym
}

// isInsideFunctionAssignedToSelf checks if the identifier is inside a function expression
// whose result (directly or via IIFE) is assigned to the same variable.
// Covers: `cb = (function(a) { return cb(1+a); })()` (IIFE)
//
//	`cb = (0, function(a) { cb(1+a); })` (non-IIFE, assigned directly)
//	`cb = (function(a) { cb(1+a); }, cb)` (discarded in comma left operand)
func isInsideFunctionAssignedToSelf(node *ast.Node, sym *ast.Symbol, checker *checker.Checker) bool {
	current := node
	for current != nil {
		if current.Kind == ast.KindFunctionExpression || current.Kind == ast.KindArrowFunction {
			// Walk up from the function expression through parens, commas, and calls
			// to find the enclosing assignment.
			ancestor := current.Parent
			for ancestor != nil {
				if ancestor.Kind == ast.KindParenthesizedExpression {
					ancestor = ancestor.Parent
					continue
				}
				if ancestor.Kind == ast.KindBinaryExpression {
					bin := ancestor.AsBinaryExpression()
					if bin != nil && bin.OperatorToken.Kind == ast.KindCommaToken {
						// Comma expression: continue walking up
						ancestor = ancestor.Parent
						continue
					}
					if bin != nil && ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
						lhsSym := checker.GetSymbolAtLocation(bin.Left)
						if lhsSym == sym {
							if isInsideStorableFunction(node, current) {
								return false
							}
							return isUnusedExpression(ancestor)
						}
					}
					break
				}
				if ancestor.Kind == ast.KindCallExpression {
					ce := ancestor.AsCallExpression()
					// Only walk through IIFE (function is the callee), not function-as-argument
					callee := ce.Expression
					for callee != nil && callee.Kind == ast.KindParenthesizedExpression {
						callee = callee.AsParenthesizedExpression().Expression
					}
					if callee == current {
						ancestor = ancestor.Parent
						continue
					}
					break
				}
				break
			}
		}
		current = current.Parent
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

// hasArrayDestructuringWrite checks if the variable has any write reference
// via array destructuring assignment (e.g., `[_x] = arr`).
func hasArrayDestructuringWrite(writeRefs map[*ast.Symbol][]*ast.Node, sym *ast.Symbol) bool {
	refs, exists := writeRefs[sym]
	if !exists {
		return false
	}
	for _, ref := range refs {
		parent := ref.Parent
		for parent != nil {
			if parent.Kind == ast.KindArrayLiteralExpression {
				return true
			}
			if parent.Kind != ast.KindSpreadElement &&
				parent.Kind != ast.KindParenthesizedExpression {
				break
			}
			parent = parent.Parent
		}
	}
	return false
}

func hasObjectRestSiblingWrite(writeRefs map[*ast.Symbol][]*ast.Node, sym *ast.Symbol) bool {
	if sym == nil {
		return false
	}
	for _, ref := range writeRefs[sym] {
		isRestTarget := false
		for current := ref.Parent; current != nil; current = current.Parent {
			if current.Kind == ast.KindSpreadAssignment {
				isRestTarget = true
			}
			if current.Kind != ast.KindObjectLiteralExpression {
				continue
			}
			if !utils.IsInDestructuringAssignment(current) {
				break
			}
			if isRestTarget {
				break
			}
			object := current.AsObjectLiteralExpression()
			if object == nil || object.Properties == nil || len(object.Properties.Nodes) == 0 {
				break
			}
			return object.Properties.Nodes[len(object.Properties.Nodes)-1].Kind == ast.KindSpreadAssignment
		}
	}
	return false
}

// isParameterInWithoutBodyDeclaration checks if a parameter is in a function-like
// declaration that has no body (overload signatures, abstract methods, type-level
// constructs). Such parameters are purely declarative and should not be reported.
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

	// Check if the module declaration has the Ambient (declare) flag.
	isAmbient := ast.GetCombinedModifierFlags(moduleDecl)&ast.ModifierFlagsAmbient != 0
	// Also check if any ancestor is a global scope augmentation (declare global { ... }).
	// GetCombinedModifierFlags doesn't walk up through ModuleBlock→ModuleDeclaration,
	// so we need to check ancestors explicitly.
	if !isAmbient {
		if ast.FindAncestor(node.Parent, func(n *ast.Node) bool {
			return ast.IsGlobalScopeAugmentation(n)
		}) != nil {
			isAmbient = true
		}
	}
	// Also check if any ancestor ModuleDeclaration has the Ambient flag.
	if !isAmbient {
		if ast.FindAncestor(moduleDecl.Parent, func(n *ast.Node) bool {
			return n.Kind == ast.KindModuleDeclaration &&
				ast.HasSyntacticModifier(n, ast.ModifierFlagsAmbient)
		}) != nil {
			isAmbient = true
		}
	}
	// In .d.ts files, all declarations are implicitly ambient.
	if !isAmbient {
		sf := ast.GetSourceFileOfNode(node)
		if sf != nil && sf.IsDeclarationFile {
			isAmbient = true
		}
	}

	if !isAmbient {
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

func isTopLevelDeclaration(node *ast.Node) bool {
	if node == nil {
		return false
	}
	switch node.Kind {
	case ast.KindVariableDeclaration,
		ast.KindFunctionDeclaration,
		ast.KindClassDeclaration,
		ast.KindInterfaceDeclaration,
		ast.KindTypeAliasDeclaration,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration:
		return true
	}
	return false
}

func isParameterNode(node *ast.Node) bool {
	return ast.FindAncestorKind(node, ast.KindParameter) != nil
}

func isCaughtErrorNode(node *ast.Node) bool {
	return ast.FindAncestorKind(node, ast.KindCatchClause) != nil
}

func isUsingDeclaration(varDeclNode *ast.Node) bool {
	return ast.IsVarUsing(varDeclNode) || ast.IsVarAwaitUsing(varDeclNode)
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

func isDestructuredArrayElement(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	// Walk up through BindingElements to find the containing pattern.
	for parent != nil {
		if parent.Kind == ast.KindArrayBindingPattern {
			return true
		}
		if parent.Kind != ast.KindBindingElement {
			break
		}
		parent = parent.Parent
	}
	return false
}

// matchesIgnorePattern checks if a variable name matches its category's
// ignore pattern, and whether the match should result in ignoring or
// reporting (when reportUsedIgnorePattern is true and the variable is used).
// Returns: (shouldIgnore bool, matchesPattern bool, matched variable type)
func matchesIgnorePattern(varName string, varInfo *VariableInfo, opts Config, writeRefs map[*ast.Symbol][]*ast.Node, sym *ast.Symbol) (bool, bool, variableType) {
	var re *regexp2.Regexp
	kind := variableTypeVariable

	if isParameterNode(varInfo.Definition) {
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

	matched := re != nil && utils.Regexp2MatchString(re, varName)

	// destructuredArrayIgnorePattern applies to array-destructured elements,
	// checking both the declaration site AND assignment sites (e.g., `let _x; [_x] = arr`).
	if !matched && opts.destructuredArrayIgnoreRe != nil {
		if isDestructuredArrayElement(varInfo.Definition) || hasArrayDestructuringWrite(writeRefs, sym) {
			matched = utils.Regexp2MatchString(opts.destructuredArrayIgnoreRe, varName)
			if matched {
				kind = variableTypeArrayDestructure
			}
		}
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
	if opts.DestructuredArrayIgnorePattern != "" && isDestructuredArrayElement(definition) {
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

// isExported checks if a variable is exported from the module. Exported variables
// are excluded from unused-var reporting. Checks: export modifier on the declaration,
// export modifier on any merged declaration (declaration merging), parent
// VariableStatement export, re-export via `export { name }`, and ancestor
// ExportDeclaration nodes on the variable or its references.
func isExported(ctx rule.RuleContext, varInfo *VariableInfo) bool {
	if varInfo.Variable == nil {
		return false
	}

	if varInfo.Definition != nil {
		modifierFlags := ast.GetCombinedModifierFlags(varInfo.Definition)
		if modifierFlags&ast.ModifierFlagsExport != 0 {
			return true
		}

		// Declaration merging: if ANY declaration of the symbol is exported,
		// the variable is considered exported (e.g., `interface Foo {} export const Foo = ...`)
		sym := ctx.TypeChecker.GetSymbolAtLocation(varInfo.Variable)
		if sym != nil && len(sym.Declarations) > 1 {
			for _, decl := range sym.Declarations {
				if ast.GetCombinedModifierFlags(decl)&ast.ModifierFlagsExport != 0 {
					return true
				}
				// Also check parent VariableStatement for export
				parent := decl.Parent
				for parent != nil && (parent.Kind == ast.KindVariableDeclarationList || parent.Kind == ast.KindVariableStatement) {
					if ast.GetCombinedModifierFlags(parent)&ast.ModifierFlagsExport != 0 {
						return true
					}
					parent = parent.Parent
				}
			}
		}

		if isTopLevelDeclaration(varInfo.Definition) {
			parent := varInfo.Definition.Parent
			for parent != nil {
				switch parent.Kind {
				case ast.KindVariableDeclarationList, ast.KindVariableStatement:
					modifierFlags := ast.GetCombinedModifierFlags(parent)
					if modifierFlags&ast.ModifierFlagsExport != 0 {
						return true
					}
				case ast.KindSourceFile, ast.KindModuleBlock:
					goto doneParentWalk
				}
				// Stop at function/class boundaries — a variable inside a function
				// is not exported even if the containing function is.
				if ast.IsFunctionLike(parent) || parent.Kind == ast.KindClassDeclaration {
					break
				}
				parent = parent.Parent
			}
		}
	}
doneParentWalk:

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

	// Check if the symbol is re-exported via `export { name }`.
	// The export specifier creates a different symbol chain, so we resolve
	// both sides through SkipAlias and compare.
	if varInfo.Definition != nil {
		sym := ctx.TypeChecker.GetSymbolAtLocation(varInfo.Variable)
		if sym != nil {
			sf := ast.GetSourceFileOfNode(varInfo.Variable)
			if isReExportedSymbol(ctx, sym, sf.AsNode()) {
				return true
			}
		}
	}

	return false
}

// isReExportedSymbol checks if a symbol is re-exported via `export { name }` or
// `export { name as alias }`. Resolves both the export specifier's symbol and the
// original symbol through SkipAlias to handle re-exports of imported bindings.
func isReExportedSymbol(ctx rule.RuleContext, sym *ast.Symbol, sourceFile *ast.Node) bool {
	found := false
	sourceFile.ForEachChild(func(child *ast.Node) bool {
		if child.Kind != ast.KindExportDeclaration {
			return false
		}
		exportDecl := child.AsExportDeclaration()
		if exportDecl == nil || exportDecl.ExportClause == nil {
			return false
		}
		// `export { ... } from 'mod'` only re-exports module bindings, never
		// in-scope locals — skip these declarations entirely so a local that
		// happens to share a name with a module export is not falsely treated
		// as re-exported.
		if exportDecl.ModuleSpecifier != nil {
			return false
		}
		if !ast.IsNamedExports(exportDecl.ExportClause) {
			return false
		}
		namedExports := exportDecl.ExportClause.AsNamedExports()
		if namedExports == nil || namedExports.Elements == nil {
			return false
		}
		for _, spec := range namedExports.Elements.Nodes {
			exportSpec := spec.AsExportSpecifier()
			if exportSpec == nil {
				continue
			}
			// When renamed (`export { join as myJoin }`), PropertyName references the local binding.
			// When not renamed (`export { join }`), Name's symbol differs from the import binding,
			// so we use GetSymbolAtLocation on PropertyName which resolves to the local symbol.
			refNode := exportSpec.PropertyName
			if refNode != nil {
				exportSym := ctx.TypeChecker.GetSymbolAtLocation(refNode)
				if exportSym == sym {
					found = true
					return true
				}
			} else {
				// No rename: compare by name text since the export specifier creates a different symbol
				exportName := exportSpec.Name()
				if exportName != nil && ast.IsIdentifier(exportName) && sym.Name == exportName.AsIdentifier().Text {
					found = true
					return true
				}
			}
		}
		return false
	})
	return found
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
	text := file.Text()
	for i := pos; i < len(text); i++ {
		ch := text[i]
		if ch == ',' {
			return i + 1
		}
		if ch != ' ' && ch != '\t' {
			break
		}
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

func isLoopKind(kind ast.Kind) bool {
	switch kind {
	case ast.KindForStatement,
		ast.KindForInStatement,
		ast.KindForOfStatement,
		ast.KindWhileStatement,
		ast.KindDoStatement:
		return true
	default:
		return false
	}
}

func isSingleStatementBody(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	switch node.Parent.Kind {
	case ast.KindIfStatement, ast.KindWithStatement:
		return true
	default:
		return isLoopKind(node.Parent.Kind)
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
	inLoopInitializer := listParent != nil && isLoopKind(listParent.Kind)
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
		fix, ok = fixFunctionParameter(ctx, definition, ac)
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
		fix, ok = removeNodeRange(ctx.SourceFile, definition), true
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

// isPropertyNameLikePosition reports whether an identifier appears in a syntactic
// position where it names a property/label/attribute rather than referring to a
// declared value or type. Such identifiers must NOT be added to `unresolvedRefs`
// even when the type checker fails to resolve them — otherwise the name-based
// fallback in processVariable will mistake them for usages of an unrelated
// same-named local variable (e.g. `obj.name` on an `any`-typed receiver
// polluting the lookup of an unused local `const name`).
func isPropertyNameLikePosition(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindPropertyAccessExpression:
		// `obj.name` — node is the `.name` part on the right.
		pae := parent.AsPropertyAccessExpression()
		return pae != nil && pae.Name() == node
	case ast.KindQualifiedName:
		// `Foo.Bar` in type position — node is the `.Bar` part on the right.
		qn := parent.AsQualifiedName()
		return qn != nil && qn.Right == node
	case ast.KindPropertyAssignment:
		// `{ name: value }` in object literal — node is the property key.
		pa := parent.AsPropertyAssignment()
		return pa != nil && pa.Name() == node
	case ast.KindBindingElement:
		// `const { name: alias } = obj` — node is the source property name.
		// The destination `alias` is a declaration name (handled separately).
		be := parent.AsBindingElement()
		return be != nil && be.PropertyName != nil && be.PropertyName == node
	case ast.KindImportSpecifier:
		// `import { name as alias } from 'mod'` — node is the source export
		// name (PropertyName), which references the module's exported binding,
		// not any in-scope variable. When the module is unresolvable the symbol
		// lookup fails and would otherwise pollute unresolvedRefs[name].
		is := parent.AsImportSpecifier()
		return is != nil && is.PropertyName != nil && is.PropertyName == node
	case ast.KindExportSpecifier:
		// ExportSpecifier semantics depend on whether the enclosing
		// ExportDeclaration is a re-export (`export { ... } from 'mod'`) or
		// a local export (no `from`):
		//   * Local export: both `Name` and `PropertyName` are references
		//     to in-scope locals. We must NOT exclude them — otherwise an
		//     unresolved local `name` reference would be missed.
		//   * Re-export: both `Name` and `PropertyName` name module-level
		//     bindings, never in-scope locals. They must be excluded so an
		//     unresolved module specifier does not pollute the lookup of
		//     a same-named local elsewhere in the file.
		es := parent.AsExportSpecifier()
		if es == nil {
			return false
		}
		exportDecl := ast.FindAncestorKind(parent, ast.KindExportDeclaration)
		if exportDecl == nil {
			return false
		}
		if exportDecl.AsExportDeclaration().ModuleSpecifier == nil {
			return false
		}
		return es.PropertyName == node || es.Name() == node
	case ast.KindJsxAttribute:
		// `<X name="..." />` — node is the attribute name.
		attr := parent.AsJsxAttribute()
		return attr != nil && attr.Name() == node
	case ast.KindJsxNamespacedName:
		// `<X xml:lang="en" />` — both `xml` (Namespace) and `lang` (Name)
		// are JSX-namespace components, never in-scope value references.
		jnn := parent.AsJsxNamespacedName()
		return jnn != nil && (jnn.Namespace == node || jnn.Name() == node)
	case ast.KindImportAttribute:
		// `import 'mod' with { type: 'json' }` — the `type` here is an
		// import-attribute key, not a value reference.
		ia := parent.AsImportAttribute()
		return ia != nil && ia.Name() == node
	case ast.KindLabeledStatement:
		// `name: while(...)` — label declaration, not a value reference.
		ls := parent.AsLabeledStatement()
		return ls != nil && ls.Label == node
	case ast.KindBreakStatement, ast.KindContinueStatement:
		// `break name` / `continue name` — label reference (separate namespace).
		return true
	case ast.KindMetaProperty:
		// `new.target` / `import.meta` — node is the keyword.name.
		mp := parent.AsMetaProperty()
		return mp != nil && mp.Name() == node
	}
	return false
}

// collectSymbolUsages walks the entire source file AST and collects:
//   - usages: maps each symbol to its usage reference nodes (read references)
//   - writeRefs: maps each symbol to its write-only reference nodes (assignments)
//   - unresolvedRefs: maps identifier text to nodes where GetSymbolAtLocation returns nil
//
// After the walk, it calls markJsxFactoryUsed to handle JSX factory implicit usage.
func collectSymbolUsages(ctx rule.RuleContext, sourceFile *ast.Node, usages map[*ast.Symbol][]*ast.Node, writeRefs map[*ast.Symbol][]*ast.Node, unresolvedRefs map[string][]*ast.Node, referencesByName map[string][]*ast.Node, coreSemantics bool) {
	var walk func(*ast.Node)
	walk = func(node *ast.Node) {
		if node == nil {
			return
		}

		if ast.IsIdentifier(node) && !isDeclarationName(node) {
			// Track write-only references separately for report position.
			// Simple assignments (=) are write-only and don't count as usage.
			if isPartOfAssignment(node) {
				sym := ctx.TypeChecker.GetSymbolAtLocation(node)
				if sym != nil {
					writeRefs[sym] = append(writeRefs[sym], node)
				}
				if coreSemantics && node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
					valueSymbol := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(node.Parent)
					if valueSymbol != nil && valueSymbol != sym {
						writeRefs[valueSymbol] = append(writeRefs[valueSymbol], node)
					}
				}
				node.ForEachChild(func(child *ast.Node) bool {
					walk(child)
					return false
				})
				return
			}
			if !isPropertyNameLikePosition(node) {
				name := node.AsIdentifier().Text
				referencesByName[name] = append(referencesByName[name], node)
			}
			// Compound assignments (+=, -=, etc.) and update expressions (++, --)
			// are both read and write. Track as writeRef for report position,
			// but don't return early — the node is still recorded as a usage below.
			if isCompoundAssignmentTarget(node) || isUpdateTarget(node) {
				sym := ctx.TypeChecker.GetSymbolAtLocation(node)
				if sym != nil {
					writeRefs[sym] = append(writeRefs[sym], node)
				}
			}
			sym := ctx.TypeChecker.GetSymbolAtLocation(node)
			if sym != nil {
				// Store usage under both the original symbol and the resolved alias.
				// This ensures import specifiers match correctly: the declaration site
				// uses the original (pre-alias) symbol, while re-exports or other
				// alias chains still work via the resolved symbol.
				usages[sym] = append(usages[sym], node)
				resolved := ctx.TypeChecker.SkipAlias(sym)
				if resolved != sym {
					usages[resolved] = append(usages[resolved], node)
				}
			} else if !isPropertyNameLikePosition(node) {
				// TypeChecker is the source of truth; this branch is only a
				// narrow fallback for residual cases where GetSymbolAtLocation
				// returns nil but the identifier IS a value/type reference
				// (e.g., empty namespaces — see TestNoUnusedVarsPatterns'
				// `namespace _Foo {} export const x = _Foo;` invalid case,
				// which still depends on the name-based lookup).
				//
				// Identifiers in pure property/label/attribute positions can
				// never refer to a top-level declared symbol, so excluding
				// them here prevents `obj.name` (any-typed) from polluting
				// the lookup of an unused local `name`.
				idText := node.AsIdentifier().Text
				unresolvedRefs[idText] = append(unresolvedRefs[idText], node)
			}
			// For shorthand properties like { stats }, the identifier serves
			// as both the property name and the value reference. GetSymbolAtLocation
			// returns the property symbol, but we also need the value symbol to
			// track usage of the referenced variable (especially for imports).
			if node.Parent != nil && node.Parent.Kind == ast.KindShorthandPropertyAssignment {
				valSym := ctx.TypeChecker.GetShorthandAssignmentValueSymbol(node.Parent)
				if valSym != nil {
					usages[valSym] = append(usages[valSym], node)
					resolved := ctx.TypeChecker.SkipAlias(valSym)
					if resolved != valSym {
						usages[resolved] = append(usages[resolved], node)
					}
				}
			}
		}

		node.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	walk(sourceFile)

	// TypeScript only creates a reference from JSX to the factory function in
	// the classic `jsx: "react"` runtime, and even then via an implicit
	// `React.createElement` call that has no identifier node for the AST walk
	// above to find. For every other mode (preserve, react-native, react-jsx)
	// the factory import has no textual reference at all. Mark it as used,
	// matching @typescript-eslint/parser's `jsxPragma` behavior (the factory
	// is considered used whenever the file contains JSX, in any runtime).
	markJsxFactoryUsed(ctx, sourceFile, usages)
}

// markJsxFactoryUsed checks if the source file contains JSX and, if so, marks
// the JSX factory (and fragment factory) imports as used. This runs for every
// jsx mode: TS never produces an AST identifier reference to the factory, so
// without this an `import React` whose only "use" is JSX would be falsely
// reported. Mirrors @typescript-eslint/parser, which treats the jsxPragma
// (default "React") as used whenever JSX is present, regardless of runtime.
func markJsxFactoryUsed(ctx rule.RuleContext, sourceFile *ast.Node, usages map[*ast.Symbol][]*ast.Node) {
	if ctx.Program == nil {
		return
	}
	opts := ctx.Program.Options()
	if opts == nil {
		return
	}
	firstJsx, firstFragment := findJsxNodes(sourceFile)
	if firstJsx == nil && firstFragment == nil {
		return
	}
	// Any JSX (element or fragment) marks the factory as used
	factoryName := "React"
	if opts.JsxFactory != "" {
		factoryName = strings.Split(opts.JsxFactory, ".")[0]
	}
	refNode := firstJsx
	if refNode == nil {
		refNode = firstFragment
	}
	markImportByNameAsUsed(ctx, sourceFile, factoryName, refNode, usages)
	// JSX fragments additionally mark the fragment factory as used
	if firstFragment != nil && opts.JsxFragmentFactory != "" {
		fragmentFactoryName := strings.Split(opts.JsxFragmentFactory, ".")[0]
		markImportByNameAsUsed(ctx, sourceFile, fragmentFactoryName, firstFragment, usages)
	}
}

// markImportByNameAsUsed finds an import with the given name and adds refNode
// as a usage reference for its symbol. We use refNode (a JSX element/fragment)
// instead of the import's own name node because processVariable filters out
// usages where usage.Pos() == declaration.Pos() (self-reference filtering).
func markImportByNameAsUsed(ctx rule.RuleContext, sourceFile *ast.Node, name string, refNode *ast.Node, usages map[*ast.Symbol][]*ast.Node) {
	sf := sourceFile.AsSourceFile()
	if sf == nil {
		return
	}
	sf.AsNode().ForEachChild(func(child *ast.Node) bool {
		// Handle import React = require('react')
		if child.Kind == ast.KindImportEqualsDeclaration {
			ieq := child.AsImportEqualsDeclaration()
			if ieq != nil && ieq.Name() != nil && ieq.Name().AsIdentifier().Text == name {
				sym := ctx.TypeChecker.GetSymbolAtLocation(ieq.Name())
				if sym != nil {
					usages[sym] = append(usages[sym], refNode)
				}
				return true
			}
			return false
		}
		if child.Kind != ast.KindImportDeclaration {
			return false
		}
		importDecl := child.AsImportDeclaration()
		if importDecl == nil || importDecl.ImportClause == nil {
			return false
		}
		clause := importDecl.ImportClause
		// Check default import: import React from '...'
		if clause.Name() != nil && clause.Name().AsIdentifier().Text == name {
			sym := ctx.TypeChecker.GetSymbolAtLocation(clause.Name())
			if sym != nil {
				usages[sym] = append(usages[sym], refNode)
			}
			return true
		}
		if clause.AsImportClause().NamedBindings != nil {
			bindings := clause.AsImportClause().NamedBindings
			// Check namespace import: import * as React from '...'
			if bindings.Kind == ast.KindNamespaceImport {
				nsImport := bindings.AsNamespaceImport()
				if nsImport != nil && nsImport.Name() != nil && nsImport.Name().AsIdentifier().Text == name {
					sym := ctx.TypeChecker.GetSymbolAtLocation(nsImport.Name())
					if sym != nil {
						usages[sym] = append(usages[sym], refNode)
					}
					return true
				}
			}
			// Check named imports: import { h } from '...'
			if bindings.Kind == ast.KindNamedImports {
				bindings.ForEachChild(func(spec *ast.Node) bool {
					if spec.Kind == ast.KindImportSpecifier {
						specName := spec.AsImportSpecifier().Name()
						if specName != nil && specName.AsIdentifier().Text == name {
							sym := ctx.TypeChecker.GetSymbolAtLocation(specName)
							if sym != nil {
								usages[sym] = append(usages[sym], refNode)
							}
							return true
						}
					}
					return false
				})
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
		if n == nil || (firstJsx != nil && firstFragment != nil) {
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

// processVariable determines whether a single declared variable/parameter/import
// is unused and, if so, reports it. The decision pipeline:
//  1. Resolve the symbol and look up usages (original sym → SkipAlias → unresolved fallback)
//  2. Filter out self-references (same position, self-modifying, inside own declaration body)
//  3. Classify remaining usages as value or type-only
//  4. Apply ignore patterns (varsIgnorePattern, argsIgnorePattern, etc.)
//  5. Skip exported symbols (except for reportUsedIgnorePattern)
//  6. Apply "after-used" logic for parameters
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
		// Fallback: check unresolved references by name.
		// This handles cases like empty namespaces where GetSymbolAtLocation
		// returns nil for the reference but the namespace symbol is valid.
		if !exists && !isImportDef {
			if unresolved, ok := ac.unresolvedRefs[name]; ok {
				for _, ref := range unresolved {
					if ref.Pos() != nameNode.Pos() && !isInsideOwnDeclaration(ref, definition) {
						usageNodes = append(usageNodes, ref)
						exists = true
					}
				}
			}
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
				if usage.Pos() != varInfo.Variable.Pos() &&
					!isSelfModifyingReference(usage, sym, ctx.TypeChecker, nameNode) &&
					!isInsideAnyOwnDeclaration(usage, allDecls) {
					filteredUsages = append(filteredUsages, usage)
				}
			}

			if len(filteredUsages) > 0 {
				onlyUsedAsType := true
				for _, usage := range filteredUsages {
					if !isInTypeContext(usage) {
						onlyUsedAsType = false
						break
					}
				}
				varInfo.Used = !onlyUsedAsType
				varInfo.OnlyUsedAsType = onlyUsedAsType
			}
		}
	}

	// vars: "local" — skip top-level (global scope) variable declarations.
	if opts.Vars == "local" && !isParameterNode(definition) && !isCaughtErrorNode(definition) {
		if definition != nil && definition.Parent != nil {
			parent := definition.Parent
			// VariableDeclaration → VariableDeclarationList → VariableStatement → SourceFile
			for parent != nil && (parent.Kind == ast.KindVariableDeclarationList || parent.Kind == ast.KindVariableStatement) {
				parent = parent.Parent
			}
			if parent != nil && parent.Kind == ast.KindSourceFile {
				return
			}
		}
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
		varInfo.Used = true
		varInfo.OnlyUsedAsType = false
	}
	if !varInfo.Used && opts.IgnoreRestSiblings && hasObjectRestSiblingWrite(ac.writeRefs, sym) {
		return
	}

	// Check ignore patterns (varsIgnorePattern / argsIgnorePattern / caughtErrorsIgnorePattern).
	// If the variable matches its category's pattern and is unused → ignore silently.
	// If it matches but IS used and reportUsedIgnorePattern is true → report as usedIgnoredVar.
	shouldIgnore, matchedPattern, matchedType := matchesIgnorePattern(name, varInfo, opts, ac.writeRefs, sym)
	if shouldIgnore {
		return
	}

	if isExported(ctx, varInfo) {
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
			ac.reporter.reportNode(reportNode, buildUsedIgnoredVarMessage(name, additional))
		}
		return
	}

	// "after-used" for parameters: skip unused params before the last used param.
	// Only applies to direct Parameter nodes, not destructured elements within them.
	if !varInfo.Used && definition != nil && definition.Kind == ast.KindParameter && opts.Args == "after-used" {
		param := definition.AsParameterDeclaration()
		if param != nil && param.Initializer == nil && isBeforeLastUsedParam(ctx, definition, ac.allUsages) {
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
		ac.reporter.reportNode(varInfo.Variable, buildUsedIgnoredVarMessage(name, usedIgnoredAdditional))
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
				ac.reporter.reportNodeWithSuggestions(reportNode, message, suggestion)
				return
			}
		}
		ac.reporter.reportNode(reportNode, message)
	}
}

func newRule(flavor ruleFlavor) rule.Rule {
	schema := (*rule.Schema)(nil)
	if !flavor.typescript {
		schema = rule.NewSchema(schemaJSON)
	}

	return rule.Rule{
		Name:             "no-unused-vars",
		RequiresTypeInfo: flavor.typescript,
		Schema:           schema,
		Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
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
				allUsages:         make(map[*ast.Symbol][]*ast.Node),
				writeRefs:         make(map[*ast.Symbol][]*ast.Node),
				unresolvedRefs:    make(map[string][]*ast.Node),
				referencesByName:  make(map[string][]*ast.Node),
				seenMergedSymbols: make(map[*ast.Symbol]bool),
				reportedUnused:    make(map[*ast.Node]bool),
				reporter:          reporter,
				tokens:            tokens,
			}
			collected := false

			seenWithoutBodyFuncSymbols := make(map[*ast.Symbol]bool)

			ensureCollected := func(node *ast.Node) {
				if !collected {
					sourceFile := ast.GetSourceFileOfNode(node)
					collectSymbolUsages(ctx, sourceFile.AsNode(), ac.allUsages, ac.writeRefs, ac.unresolvedRefs, ac.referencesByName, !flavor.typescript)
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
					hasRestSibling := false
					if opts.IgnoreRestSiblings {
						nameNode.ForEachChild(func(child *ast.Node) bool {
							if child.Kind == ast.KindBindingElement {
								elem := child.AsBindingElement()
								if elem != nil && elem.DotDotDotToken != nil {
									hasRestSibling = true
									return true
								}
							}
							return false
						})
					}
					nameNode.ForEachChild(func(child *ast.Node) bool {
						if child.Kind == ast.KindBindingElement {
							elem := child.AsBindingElement()
							if elem != nil && elem.Name() != nil {
								if hasRestSibling && elem.DotDotDotToken == nil {
									return false
								}
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
					if isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node) {
						return
					}
					if opts.IgnoreUsingDeclarations && isUsingDeclaration(node) {
						return
					}
					// Skip for-in/for-of declarations whose body starts with return.
					// E.g., `for (var name in obj) { return true; }` — the variable
					// is considered "used" (checking property existence).
					if forStmt := isForInOfDeclaration(node); forStmt != nil {
						body := forStmt.AsForInOrOfStatement().Statement
						if forInBodyStartsWithReturn(body) {
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
						if isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node) {
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
					if ast.IsGlobalScopeAugmentation(node) {
						return
					}
					if ast.FindAncestor(node.Parent, func(n *ast.Node) bool { return ast.IsGlobalScopeAugmentation(n) }) != nil {
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
					if isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node) {
						return
					}

					// Skip namespace augmentations — if the namespace symbol has
					// declarations outside this file, it's augmenting an existing
					// namespace (e.g., `declare namespace NodeJS { ... }`).
					sym := ctx.TypeChecker.GetSymbolAtLocation(nameNode)
					if sym != nil && len(sym.Declarations) > 1 {
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
					if isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node) {
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
					if isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node) {
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
					if isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node) {
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
					if isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node) {
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

					if isParameterInWithoutBodyDeclaration(node) {
						return
					}

					// Skip TypeScript's `this` parameter (type annotation only, not a real param).
					// In tsgo, the `this` parameter name is parsed as an Identifier with text "this".
					if paramDecl.Name() != nil &&
						(paramDecl.Name().Kind == ast.KindThisKeyword ||
							(ast.IsIdentifier(paramDecl.Name()) && paramDecl.Name().AsIdentifier().Text == "this")) {
						return
					}

					// Skip constructor parameter properties (private/protected/public/readonly params).
					// These are promoted to class fields and are inherently "used".
					if ast.HasSyntacticModifier(node, ast.ModifierFlagsParameterPropertyModifier) {
						return
					}

					if paramDecl.Name() != nil {
						processBindingName(paramDecl.Name(), node)
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
					// Generic type parameter declarations: `<T>`, `<T = unknown>`, `<T extends U>`.
					// Skip nodes that syntactically share KindTypeParameter in tsgo but aren't
					// parameter declarations: `infer T`, mapped-type `[P in K]`, JSDoc @template.
					parent := node.Parent
					if parent != nil {
						switch parent.Kind {
						case ast.KindInferType, ast.KindMappedType, ast.KindJSDocTemplateTag:
							return
						}
					}
					if isInsideAmbientModuleBlock(node) || isInDtsWithoutExplicitExports(node) {
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
			}

			if !flavor.typescript {
				ensureCollected(ctx.SourceFile.AsNode())
				if opts.Vars != "local" {
					for _, inlineGlobal := range ctx.InlineGlobals {
						if !inlineGlobal.Declared || len(inlineGlobal.NameRanges) == 0 {
							continue
						}
						if len(ac.referencesByName[inlineGlobal.Name]) > 0 {
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
