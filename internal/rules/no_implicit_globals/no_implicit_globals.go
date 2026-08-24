package no_implicit_globals

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
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
// NOTE: Unlike ESLint, rslint does not expose languageOptions.sourceType or
// parserOptions.ecmaFeatures.globalReturn (see PORT_RULE.md's framework-gap
// list). "Is this the global scope" is instead derived from the file's actual
// module-ness: ast.IsExternalModule (module syntax, or an extension TypeScript
// forces module-ness on) combined with the resolved language defaults for the
// file extension (ctx.Refs.HasNonGlobalTopLevelScope, which covers .cjs's
// CommonJS wrapper and .js/.mjs's default module treatment) — the same
// combination no_redeclare already uses for its own "is this the Program's
// global scope" check.
var NoImplicitGlobalsRule = rule.Rule{
	Name:   "no-implicit-globals",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		if ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}
		opts := parseOptions(rawOptions)

		hasNonGlobalTopLevelScope := ast.IsExternalModule(ctx.SourceFile) ||
			(ctx.Refs != nil && ctx.Refs.HasNonGlobalTopLevelScope())

		// ESLint records an implicit global only for a write in sloppy-mode
		// code, so top-level strictness decides whether a leak is reportable.
		// It comes from the same module-ness the declaration checks use, minus
		// CommonJS: a .cjs file has its own top-level scope without being an ES
		// module, so its top level stays sloppy until it uses module syntax
		// itself.
		strictTopLevel := hasNonGlobalTopLevelScope &&
			(utils.HasModuleSyntax(ctx.SourceFile) || !utils.IsCommonJSFileExtension(ctx.SourceFile.FileName()))

		listeners := rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				checkImplicitGlobalWrite(ctx, node, strictTopLevel)
			},
		}

		if hasNonGlobalTopLevelScope {
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
	if functionDeclarationScope(ctx, node) != sourceFileNode {
		return
	}
	reportDeclaration(ctx, node, nameNode.Text(), "function", false)
}

// functionDeclarationScope returns the scope a function declaration binds its
// name in. eslint-scope only creates block scopes from ES2015 on, so under
// `languageOptions.ecmaVersion` 3 or 5 a block-level function declaration binds
// in the enclosing variable scope instead, and `{ function foo() {} }` is a
// global declaration.
func functionDeclarationScope(ctx rule.RuleContext, node *ast.Node) *ast.Node {
	if ctx.LanguageOptions.EffectiveECMAVersion() < 2015 {
		return utils.FindEnclosingScope(node)
	}
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

// checkImplicitGlobalWrite handles the two reference-driven diagnostics that
// apply file-wide, independently of module-ness: an assignment to a read-only
// global (assignmentToReadonlyGlobal) and an assignment to a name that is
// neither declared anywhere in the file nor a known global at all, outside
// strict-mode code (globalVariableLeak). Both only ever consider a pure `=`
// assignment or bare for-in/for-of loop variable — the same shape ESLint's
// own reference.isWrite() && !reference.isRead() filter selects, which
// excludes compound assignments (foo += 1) and update expressions (foo++).
//
// The two diagnostics differ on `/* exported */`: upstream skips an exported
// variable before reaching its reference loop, so the readonly assignment goes
// unreported, while the leak is collected from the global scope's implicit
// variables, which the directive never touches.
//
// Both stop at any name the file itself defines, which in a global script means
// any top-level declaration whatsoever: ESLint keeps that file's declarations
// and the configured globals in one global-scope variable, so a type-only
// `interface foo {}` gives `foo` a definition and neither an implicit variable
// (the leak) nor a definition-less read-only global (the assignment) remains.
// Inner scopes hold separate variables that a value reference never resolves
// to, so the same declaration inside a function or namespace leaves the name
// global — the distinction IsGlobalNameReference draws.
func checkImplicitGlobalWrite(ctx rule.RuleContext, node *ast.Node, strictTopLevel bool) {
	root, writes := findPureAssignmentRoot(node)
	if root == nil {
		return
	}
	name := node.Text()
	syntheticPatternName := isSyntheticPatternName(node)
	if ctx.Refs != nil && !syntheticPatternName &&
		!ctx.Refs.IsGlobalNameReference(node, name, ast.SymbolFlagsValue) {
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
		strictnessNode := node
		if syntheticPatternName {
			// PatternVisitor's synthetic write belongs to the containing
			// assignment, outside a function/class expression's own scope.
			strictnessNode = root
		}
		if !strictTopLevel && !utils.IsInStrictMode(strictnessNode, ctx.SourceFile) {
			for range writes {
				ctx.ReportNode(root, globalVariableLeakMessage)
			}
		}
	}
}

// isSyntheticPatternName reports a function/class expression name that
// PatternVisitor reaches as an assignment target while traversing a
// parser-accepted expression-shaped pattern. The ordinary reference index
// correctly resolves such a name to its expression-local self-binding, but
// scope-manager's synthetic pattern traversal records a separate global write.
func isSyntheticPatternName(node *ast.Node) bool {
	if node == nil || node.Parent == nil {
		return false
	}
	parent := node.Parent
	if parent.Kind != ast.KindFunctionExpression && parent.Kind != ast.KindClassExpression {
		return false
	}
	return parent.Name() == node && utils.IsInDestructuringAssignment(parent)
}

// findPureAssignmentRoot walks from a candidate write-target identifier up
// through destructuring containers and transparent TS wrappers (parens,
// non-null assertions, `as` and `<T>` assertions) to the enclosing pure `=`
// assignment or for-in/for-of statement, mirroring upstream's ASSIGNMENT_NODES
// walk of reference.identifier.parent. Alongside the root it returns how many
// write references the scope manager records for the identifier: the assignment
// itself, plus one for every destructuring default the identifier sits under,
// since eslint-scope adds a write per enclosing AssignmentPattern before the
// one for the assignment. `[[foo = 1] = []] = arr` therefore reports three
// times over the whole assignment. It returns nil for anything that isn't a pure
// write target: compound/logical assignment operators and update expressions
// read as well as write, so ESLint's ordinary reference analysis does not treat
// them as leak/readonly-assignment candidates. PatternVisitor is the exception:
// a compound assignment inside an expression-shaped pattern contributes one
// write for its left side before the containing pattern contributes another.
//
// TypeScript expression wrappers are where the walk stops mirroring the AST and
// starts mirroring the scope manager. An AssignmentExpression's Left is
// unwrapped exactly once — a TSAsExpression, TSTypeAssertion or
// TSNonNullExpression, never a TSSatisfiesExpression — and what remains has to
// be a pattern, so `(foo as any) = 1` is a write while `(foo as any)! = 1`,
// `foo!! = 1` and `(foo satisfies any) = 1` are not. A wrapper around a
// destructuring target does not survive that test either: `[foo]` is no longer a
// destructuring pattern once an assertion wraps it, so `([foo] as any) = arr` is
// no write at all. Every other path — a for-in/for-of head, anything nested
// inside a pattern — visits the pattern directly and accepts a wrapped target
// however deeply it is nested, `satisfies` included: `[foo satisfies any] = arr`
// is a write.
func findPureAssignmentRoot(node *ast.Node) (*ast.Node, int) {
	if ast.IsPartOfTypeNode(node) {
		return nil, 0
	}
	current := node
	writes := 1
	// wrappers counts the TS expression wrappers the walk has climbed since the
	// last pattern node, which is what an AssignmentExpression's Left has to
	// shed in its single unwrap; satisfies records whether any of them is the
	// one wrapper that unwrap never sheds.
	wrappers := 0
	satisfies := false
	// enterPattern moves the walk onto a pattern node. Only the wrappers
	// between the last pattern node and the assignment have to survive that
	// single unwrap, so reaching one clears the wrapper state.
	enterPattern := func(pattern *ast.Node) {
		wrappers = 0
		satisfies = false
		current = pattern
	}
	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil {
				return nil, 0
			}
			if binary.OperatorToken.Kind != ast.KindEqualsToken {
				if !utils.IsInDestructuringAssignment(parent) {
					return nil, 0
				}
				if ast.IsCompoundAssignmentOperator(binary.OperatorToken.Kind) {
					if binary.Left != current {
						return nil, 0
					}
					writes++
				}
				current = parent
				continue
			}
			if binary.Left != current {
				return nil, 0
			}
			if utils.IsInDestructuringAssignment(parent) {
				writes++
				enterPattern(parent)
				continue
			}
			if wrappers > 1 || satisfies {
				return nil, 0
			}
			return parent, writes

		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := parent.AsForInOrOfStatement()
			if stmt == nil || stmt.Initializer != current {
				return nil, 0
			}
			return parent, writes

		case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
			if !utils.IsInDestructuringAssignment(parent) {
				return nil, 0
			}
			enterPattern(parent)

		case ast.KindShorthandPropertyAssignment:
			shorthand := parent.AsShorthandPropertyAssignment()
			if shorthand == nil || shorthand.Name() != current {
				return nil, 0
			}
			// `{ foo = 1 } = obj` keeps its default on the shorthand rather
			// than in a nested assignment, but the scope manager still sees an
			// AssignmentPattern and records its extra write.
			if shorthand.ObjectAssignmentInitializer != nil {
				writes++
			}
			enterPattern(parent)

		case ast.KindPropertyAssignment:
			propAssignment := parent.AsPropertyAssignment()
			if propAssignment == nil || propAssignment.Initializer != current {
				return nil, 0
			}
			enterPattern(parent)

		case ast.KindSpreadElement, ast.KindSpreadAssignment:
			enterPattern(parent)

		case ast.KindNonNullExpression, ast.KindAsExpression, ast.KindTypeAssertionExpression:
			wrappers++
			current = parent

		case ast.KindSatisfiesExpression:
			satisfies = true
			current = parent

		case ast.KindParenthesizedExpression:
			current = parent

		case ast.KindCallExpression:
			// PatternVisitor treats call arguments as right-hand reads, but
			// continues through the callee. This makes `[fn(arg)] = value`
			// a write to `fn` only, despite the unusual parser-accepted target.
			call := parent.AsCallExpression()
			if call == nil || call.Expression != current || !utils.IsInDestructuringAssignment(parent) {
				return nil, 0
			}
			current = parent

		case ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
			// A member expression writes its property, not either identifier
			// inside the expression. PatternVisitor records both as reads.
			return nil, 0

		default:
			// @typescript-eslint's PatternVisitor falls back to ordinary child
			// traversal for parser-accepted expression-shaped pattern elements.
			// Keep climbing while this expression remains inside the assignment
			// pattern; IsInDestructuringAssignment excludes default RHS values
			// and computed property keys.
			if !utils.IsInDestructuringAssignment(parent) {
				return nil, 0
			}
			current = parent
		}
	}
	return nil, 0
}
