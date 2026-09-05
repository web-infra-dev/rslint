package no_implicit_globals

import (
	_ "embed"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_implicit_globals.schema.json
var schemaJSON []byte

type options struct {
	lexicalBindings bool
}

func parseOptions(rawOptions []any) options {
	if len(rawOptions) == 0 {
		return options{}
	}
	m, _ := rawOptions[0].(map[string]any)
	lexical, _ := m["lexicalBindings"].(bool)
	return options{lexicalBindings: lexical}
}

func globalNonLexicalBindingMessage(kind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "globalNonLexicalBinding",
		Description: "Unexpected " + kind + " declaration in the global scope, wrap in an IIFE for a local variable, assign as global property for a global variable.",
	}
}

func globalLexicalBindingMessage(kind string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "globalLexicalBinding",
		Description: "Unexpected " + kind + " declaration in the global scope, wrap in a block or in an IIFE.",
	}
}

var redeclarationOfReadonlyGlobalMessage = rule.RuleMessage{
	Id:          "redeclarationOfReadonlyGlobal",
	Description: "Unexpected redeclaration of read-only global variable.",
}

var assignmentToReadonlyGlobalMessage = rule.RuleMessage{
	Id:          "assignmentToReadonlyGlobal",
	Description: "Unexpected assignment to read-only global variable.",
}

var globalVariableLeakMessage = rule.RuleMessage{
	Id:          "globalVariableLeak",
	Description: "Global variable leak, declare the variable if it is intended to be local.",
}

// NoImplicitGlobalsRule disallows var/function/const/let/class declarations in
// the global scope, assignments to undeclared (implicit global) variables, and
// redeclarations of or assignments to read-only global variables.
//
// NOTE: rslint does not expose parserOptions.ecmaFeatures.globalReturn. Whether
// this is a global program scope comes from the normalized sourceType and the
// matching RefStore initialization; this is the same answer no_redeclare uses
// for its own Program-scope check.
var NoImplicitGlobalsRule = rule.Rule{
	Name:   "no-implicit-globals",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		if ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}
		opts := parseOptions(rawOptions)

		hasNonGlobalProgramScope := ctx.Refs != nil && ctx.Refs.HasNonGlobalProgramScope()

		// ESLint records an implicit global only for a write in sloppy-mode
		// code, so top-level strictness decides whether a leak is reportable.
		// It comes from the resolved source goal rather than syntax alone:
		// CommonJS stays sloppy even when the parser accepts import/export.
		writeContext := writeContextCache{}
		listeners := rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				checkImplicitGlobalWrite(ctx, node, &writeContext)
			},
		}

		if hasNonGlobalProgramScope {
			return listeners
		}

		sourceFileNode := ctx.SourceFile.AsNode()

		listeners[ast.KindVariableDeclarationList] = func(node *ast.Node) {
			checkVariableDeclarationList(ctx, node, sourceFileNode, opts)
		}
		listeners[ast.KindFunctionDeclaration] = func(node *ast.Node) {
			checkFunctionDeclaration(ctx, node, sourceFileNode)
		}
		if opts.lexicalBindings {
			listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
				checkClassDeclaration(ctx, node, sourceFileNode)
			}
		}

		return listeners
	},
}

// reportDeclaration applies the readonly-global / writable-global / plain
// precedence shared by every declaration kind this rule inspects, then
// reports at declNode — the shared position for every name bound by one
// declarator, matching ESLint's own def.node quirk for destructured bindings
// (see no_implicit_globals_extras_test.go for the locked-in shape).
func reportDeclaration(ctx rule.RuleContext, declNode *ast.Node, name string, kind string, lexical bool) {
	if ctx.Exported.Has(name) {
		return
	}
	switch ctx.Globals.Access(name) {
	case utils.GlobalAccessWritable:
		return
	case utils.GlobalAccessReadonly:
		ctx.ReportNode(declNode, redeclarationOfReadonlyGlobalMessage)
		return
	}
	if lexical {
		ctx.ReportNode(declNode, globalLexicalBindingMessage(kind))
	} else {
		ctx.ReportNode(declNode, globalNonLexicalBindingMessage(kind))
	}
}

func checkVariableDeclarationList(ctx rule.RuleContext, node *ast.Node, sourceFileNode *ast.Node, opts options) {
	if node.Flags&ast.NodeFlagsAmbient != 0 {
		// `declare var x;` (bare or inside `declare global { ... }`) produces
		// no runtime binding at all, so it can't leak or shadow anything.
		// Upstream reports the bare form — the scope manager gives it an
		// ordinary Variable def — but "wrap in an IIFE" is not advice an
		// ambient declaration can act on; see the rule doc's differences list.
		return
	}
	isVar := utils.IsVarKeyword(node)
	if !isVar && !opts.lexicalBindings {
		return
	}

	var kind string
	lexical := !isVar
	if isVar {
		if utils.FindEnclosingScope(node) != sourceFileNode {
			return
		}
		kind = "'var'"
	} else {
		if ast.GetEnclosingBlockScopeContainer(node) != sourceFileNode {
			return
		}
		var ok bool
		kind, ok = lexicalDeclarationKind(node.Flags)
		if !ok {
			// `using` / `await using` declarations have no ESLint analog;
			// this rule predates the proposal entirely.
			return
		}
	}

	list := node.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil {
		return
	}
	for _, declarator := range list.Declarations.Nodes {
		decl := declarator.AsVariableDeclaration()
		if decl == nil || decl.Name() == nil {
			continue
		}
		utils.CollectBindingNames(decl.Name(), func(_ *ast.Node, name string) {
			reportDeclaration(ctx, declarator, name, kind, lexical)
		})
	}
}

// lexicalDeclarationKind reports the quoted `'let'`/`'const'` message kind for
// a VariableDeclarationList's flags, or ok=false for `using`/`await using`.
func lexicalDeclarationKind(flags ast.NodeFlags) (kind string, ok bool) {
	switch flags & ast.NodeFlagsBlockScoped {
	case ast.NodeFlagsLet:
		return "'let'", true
	case ast.NodeFlagsConst:
		return "'const'", true
	default:
		return "", false
	}
}

func checkFunctionDeclaration(ctx rule.RuleContext, node *ast.Node, sourceFileNode *ast.Node) {
	fn := node.AsFunctionDeclaration()
	if fn == nil || node.Body() == nil {
		// Only a real function declaration binds a name: an overload signature
		// shares the implementation's binding and an ambient one has no runtime
		// binding at all. Upstream counts one declaration per signature and
		// reports ambient forms too; see the rule doc's differences list.
		return
	}
	nameNode := fn.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return
	}
	if functionDeclarationScope(node) != sourceFileNode {
		return
	}
	reportDeclaration(ctx, node, nameNode.Text(), "function", false)
}

// functionDeclarationScope returns the scope supplied by the TypeScript
// parser used by rslint. Its scope manager retains block-level function scope
// even when languageOptions.ecmaVersion is 3 or 5.
func functionDeclarationScope(node *ast.Node) *ast.Node {
	return ast.GetEnclosingBlockScopeContainer(node)
}

func checkClassDeclaration(ctx rule.RuleContext, node *ast.Node, sourceFileNode *ast.Node) {
	if node.Flags&ast.NodeFlagsAmbient != 0 {
		// `declare class Foo {}` produces no runtime binding; unlike a function
		// declaration a class always has a body, so the ambient flag is the
		// only signal. Upstream reports it; see the rule doc's differences
		// list.
		return
	}
	cls := node.AsClassDeclaration()
	if cls == nil {
		return
	}
	nameNode := cls.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return
	}
	if ast.GetEnclosingBlockScopeContainer(node) != sourceFileNode {
		return
	}
	reportDeclaration(ctx, node, nameNode.Text(), "class", true)
}

// checkImplicitGlobalWrite reports pure runtime writes to undeclared names and
// read-only globals. TypeScript type syntax is deliberately excluded: an
// assertion around a value can be an assignment target after erasure, but the
// identifiers inside the assertion's type never are.
func checkImplicitGlobalWrite(ctx rule.RuleContext, node *ast.Node, writeContext *writeContextCache) {
	root, writes := findPureAssignmentRoot(node, writeContext)
	if root == nil {
		return
	}
	name := node.Text()
	if ctx.Refs != nil && !ctx.Refs.IsGlobalNameReference(node, name, ast.SymbolFlagsValue|ast.SymbolFlagsAlias) {
		return
	}
	switch ctx.Globals.Access(name) {
	case utils.GlobalAccessReadonly:
		if ctx.Exported.Has(name) {
			return
		}
		for range writes {
			ctx.ReportNode(root, assignmentToReadonlyGlobalMessage)
		}
	case utils.GlobalAccessWritable:
		// Writable globals may be freely assigned.
	default:
		if !utils.IsInStrictModeWithSourceType(root, ctx.SourceFile, ctx.LanguageOptions.EffectiveSourceType()) {
			for range writes {
				ctx.ReportNode(root, globalVariableLeakMessage)
			}
		}
	}
}

// findPureAssignmentRoot delegates assignment-target discovery to tsgo. The
// only bridge needed is for erased TypeScript assertions, which
// GetAssignmentTarget intentionally does not cross. Default values retain the
// existing ESLint-compatible diagnostic multiplicity.
func findPureAssignmentRoot(node *ast.Node, writeContext *writeContextCache) (*ast.Node, int) {
	if node == nil || node.Kind != ast.KindIdentifier {
		return nil, 0
	}

	root := assignmentTargetThroughAssertions(node)
	if root == nil || writeContext.isErasedSyntax(node, root) || writeContext.isInsideWrappedPattern(node, root) ||
		isRestAssignmentRoot(root) {
		return nil, 0
	}
	writes := 1
	for utils.IsDefaultValueInDestructuringAssignment(root) {
		writes++
		root = assignmentTargetThroughAssertions(root)
		if root == nil {
			return nil, 0
		}
	}
	// A standalone `=` recovered inside an invalid pattern element is not an
	// executable assignment. Valid default initializers have already been
	// lifted to their enclosing assignment; IsInDestructuringAssignment stops
	// at default right-hand sides and computed property names, which remain
	// ordinary runtime expressions.
	if isRecoveredInvalidPatternRoot(root) {
		return nil, 0
	}
	switch root.Kind {
	case ast.KindBinaryExpression:
		binary := root.AsBinaryExpression()
		if binary == nil || binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
			return nil, 0
		}
	case ast.KindForInStatement, ast.KindForOfStatement:
		// A non-declaration loop initializer is a pure write target.
	default:
		return nil, 0
	}

	for current := node.Parent; current != nil && current != root; current = current.Parent {
		if current.Kind == ast.KindShorthandPropertyAssignment {
			shorthand := current.AsShorthandPropertyAssignment()
			if shorthand != nil && shorthand.ObjectAssignmentInitializer != nil {
				writes++
			}
		}
	}
	return root, writes
}

// assignmentTargetThroughAssertions delegates the assignment-target walk to
// tsgo and only steps across its erased assertion outer expressions. In
// particular, it does not cross an instantiation expression: TypeScript rejects
// one used as an assignment target.
func assignmentTargetThroughAssertions(node *ast.Node) *ast.Node {
	for current := node; current != nil; {
		if target := ast.GetAssignmentTarget(current); target != nil {
			return target
		}
		parent := current.Parent
		if parent == nil || !ast.IsOuterExpression(parent, ast.OEKAssertions|ast.OEKParentheses) || parent.Expression() != current {
			return nil
		}
		current = parent
	}
	return nil
}

// isInsideWrappedPattern rejects every write nested in an array/object pattern
// that is itself wrapped in parentheses or a TypeScript assertion. The target
// check keeps ordinary parenthesized array/object values inside a member
// receiver or index expression out of this recovery-only boundary.
func (c *writeContextCache) isInsideWrappedPattern(node *ast.Node, root *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		if current == c.wrappedKey {
			if c.wrapped {
				// A positive result belongs to the wrapped container, not the
				// assignment root: a sibling pattern may remain executable.
				return true
			}
			return c.rememberWrapped(root, false)
		}
		if !ast.IsAssignmentPattern(current) {
			continue
		}
		if isWrappedPatternContainer(current) {
			return c.rememberWrapped(current, true)
		}
	}
	return c.rememberWrapped(root, false)
}

func isWrappedPatternContainer(node *ast.Node) bool {
	current := node
	wrapped := false
	for current.Parent != nil {
		parent := current.Parent
		if !ast.IsOuterExpression(parent, ast.OEKAssertions|ast.OEKParentheses|ast.OEKExpressionsWithTypeArguments) {
			break
		}
		if parent.Expression() != current {
			break
		}
		wrapped = true
		current = parent
	}
	return wrapped && ast.GetAssignmentTarget(current) != nil
}

// isRestAssignmentRoot rejects a recovered default/assignment used directly
// as an array or object rest target. Assignments evaluated inside a valid
// member target stop at that member expression and remain reportable.
func isRestAssignmentRoot(node *ast.Node) bool {
	current := node
	for current.Parent != nil &&
		ast.IsOuterExpression(current.Parent, ast.OEKAssertions|ast.OEKParentheses) &&
		current.Parent.Expression() == current {
		current = current.Parent
	}
	parent := current.Parent
	return parent != nil && (parent.Kind == ast.KindSpreadElement || parent.Kind == ast.KindSpreadAssignment) &&
		parent.Expression() == current && ast.GetAssignmentTarget(current) != nil
}

// isRecoveredInvalidPatternRoot distinguishes a standalone assignment buried
// in an invalid pattern element from one evaluated while resolving a valid
// member target. Default right-hand sides and computed property names already
// return false from IsInDestructuringAssignment.
func isRecoveredInvalidPatternRoot(node *ast.Node) bool {
	if !utils.IsInDestructuringAssignment(node) {
		return false
	}
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsArrayLiteralOrObjectLiteralDestructuringPattern(current) {
			// Do not let a member expression outside this pattern container make
			// its invalid contents look executable.
			return true
		}
		if ast.IsAccessExpression(current) &&
			!ast.IsOptionalChain(current) && assignmentTargetThroughAssertions(current) != nil {
			return false
		}
	}
	return true
}

// writeContextCache excludes identifiers in syntax that cannot contribute a
// runtime write. NodeFlagsAmbient is propagated by the parser to descendants
// of declaration files and `declare` declarations. Interfaces, bodyless
// function-like declarations, and abstract properties need explicit checks
// because recovery expressions may still appear beneath their AST nodes.
type writeContextCache struct {
	erasedRoot, wrappedKey *ast.Node
	erased, wrapped        bool
}

// isErasedSyntax checks whether the path from a target identifier through its
// assignment root is erased TypeScript syntax, then classifies the root's
// surrounding context. Remembering only the previous root makes nested
// assignment chains linear without allocating a per-file node map: preorder
// traversal visits the enclosing root before the roots nested inside it.
func (c *writeContextCache) isErasedSyntax(node *ast.Node, root *ast.Node) bool {
	for current := node; current != nil && current != root; current = current.Parent {
		if ast.IsPartOfTypeNode(current) {
			return true
		}
	}

	ambient := false
	emittedPropertyPart := false
	for current := root; current != nil; current = current.Parent {
		if current == c.erasedRoot {
			return c.rememberErased(root, c.erased)
		}
		ambient = ambient || utils.IsInAmbientContext(current)
		if current.Kind == ast.KindDecorator || current.Kind == ast.KindComputedPropertyName {
			emittedPropertyPart = true
		}
		if ast.IsPartOfTypeNode(current) ||
			current.Kind == ast.KindInterfaceDeclaration ||
			(ast.IsFunctionLike(current) && current.Body() == nil) {
			return c.rememberErased(root, true)
		}
		if current.Kind != ast.KindPropertyDeclaration || (!ambient && !ast.HasAbstractModifier(current)) {
			continue
		}
		owner := current.Parent
		if emittedPropertyPart && ast.HasDecorators(current) && owner != nil && ast.IsClassLike(owner) &&
			!utils.IsInAmbientContext(owner) {
			// With legacy decorators enabled, TypeScript emits the decorator
			// expression and computed name even though the property is erased.
			return c.rememberErased(root, false)
		}
		return c.rememberErased(root, true)
	}
	return c.rememberErased(root, ambient)
}

func (c *writeContextCache) rememberErased(root *ast.Node, erased bool) bool {
	c.erasedRoot = root
	c.erased = erased
	return erased
}

func (c *writeContextCache) rememberWrapped(key *ast.Node, wrapped bool) bool {
	c.wrappedKey = key
	c.wrapped = wrapped
	return wrapped
}
