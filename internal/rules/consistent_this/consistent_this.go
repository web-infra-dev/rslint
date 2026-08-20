package consistent_this

import (
	_ "embed"
	"fmt"
	"slices"
	"sort"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed consistent_this.schema.json
var schemaJSON []byte

// scopeKinds are every function-like node kind that gets its own "was the
// alias assigned in this exact scope" check, mirroring upstream's
// "FunctionExpression:exit" / "FunctionDeclaration:exit" listeners. Arrow
// functions are intentionally excluded: upstream never listens for
// "ArrowFunctionExpression:exit", so an alias declared (and never assigned)
// inside an arrow function body is never checked.
var scopeKinds = []ast.Kind{
	ast.KindFunctionDeclaration,
	ast.KindFunctionExpression,
	ast.KindMethodDeclaration,
	ast.KindConstructor,
	ast.KindGetAccessor,
	ast.KindSetAccessor,
}

// scopeVariableKinds are the declaration kinds eslint-scope makes a variable
// of its scope for. Class and interface members, enum members and object
// literal properties are named declarations too, but none of them is a
// variable of any scope, so they are left out.
var scopeVariableKinds = []ast.Kind{
	ast.KindVariableDeclaration,
	ast.KindBindingElement,
	ast.KindParameter,
	ast.KindFunctionDeclaration,
	ast.KindClassDeclaration,
	ast.KindImportClause,
	ast.KindImportSpecifier,
	ast.KindNamespaceImport,
	ast.KindImportEqualsDeclaration,
	ast.KindTypeAliasDeclaration,
	ast.KindInterfaceDeclaration,
	ast.KindEnumDeclaration,
	ast.KindModuleDeclaration,
	ast.KindTypeParameter,
}

func aliasNotAssignedToThisMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "aliasNotAssignedToThis",
		Description: fmt.Sprintf("Designated alias '%s' is not assigned to 'this'.", name),
		Data:        map[string]string{"name": name},
	}
}

func unexpectedAliasMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "unexpectedAlias",
		Description: fmt.Sprintf("Unexpected alias '%s' for 'this'.", name),
		Data:        map[string]string{"name": name},
	}
}

// diagnostic is one pending report. Upstream emits the per-assignment reports
// while it walks and the was-it-assigned reports at each scope's exit
// (including Program scope, which only resolves once the whole file — and,
// for var, every sibling declaration — has been seen), then ESLint sorts
// everything by position before handing it to the user. rslint has no
// traversal hook that fires after the whole file, so every report is
// collected here and sorted immediately before being emitted, reproducing
// that order.
type diagnostic struct {
	textRange core.TextRange
	message   rule.RuleMessage
}

// https://eslint.org/docs/latest/rules/consistent-this
var ConsistentThisRule = rule.Rule{
	Name:   "consistent-this",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		aliases := parseOptions(options)
		var diagnostics []diagnostic

		reportRange := func(textRange core.TextRange, msg rule.RuleMessage) {
			diagnostics = append(diagnostics, diagnostic{
				textRange: textRange,
				message:   msg,
			})
		}

		report := func(node *ast.Node, msg rule.RuleMessage) {
			reportRange(utils.TrimNodeTextRange(ctx.SourceFile, node), msg)
		}

		// checkAssignment mirrors upstream's checkAssignment: node is the
		// VariableDeclaration or assignment BinaryExpression being reported
		// on, name is the identifier being assigned, value is the
		// (un-skipped) right-hand side, and hasNonEqualsOperator is true
		// only for a compound assignment (e.g. `self += this`).
		checkAssignment := func(node *ast.Node, name string, value *ast.Node, hasNonEqualsOperator bool) {
			isThis := ast.SkipParentheses(value).Kind == ast.KindThisKeyword
			if slices.Contains(aliases, name) {
				if !isThis || hasNonEqualsOperator {
					report(node, aliasNotAssignedToThisMessage(name))
				}
			} else if isThis {
				report(node, unexpectedAliasMessage(name))
			}
		}

		// isThisAssignmentTarget reports whether ref is a target of a plain `=`
		// assignment whose right-hand side is `this`, matching upstream's
		// `write.type === "ThisExpression" && write.parent.operator === "="`.
		// eslint-scope hands every name in a destructuring target the whole
		// assignment's right-hand side as its write expression, so
		// `({ self } = this)` and `[self] = this` count exactly as
		// `self = this` does.
		isThisAssignmentTarget := func(ref *ast.Node) bool {
			root := rootAssignmentOfTarget(ref)
			if root == nil {
				return false
			}
			bin := root.AsBinaryExpression()
			return bin.OperatorToken.Kind == ast.KindEqualsToken &&
				ast.SkipParentheses(bin.Right).Kind == ast.KindThisKeyword
		}

		// aliasDeclarations indexes every declaration named after one of the
		// aliases, keyed by that name, as the walk reaches it. A scope's own
		// declarations are all in by the time its exit fires, and the whole
		// file's are in by the time flush runs. The binder's own tables stay
		// the source of truth for which scope a declaration belongs to; this
		// index only recovers the declarations those tables drop, which
		// findScopeVariable explains.
		aliasDeclarations := map[string][]*ast.Node{}
		collectDeclaration := func(node *ast.Node) {
			name := ast.GetNameOfDeclaration(node)
			if name == nil || name.Kind != ast.KindIdentifier {
				return
			}
			if text := name.Text(); slices.Contains(aliases, text) {
				aliasDeclarations[text] = append(aliasDeclarations[text], node)
			}
		}

		// findScopeVariable collects what eslint-scope would hold in one
		// variable: every declaration of alias belonging to scopeNode, and
		// every symbol carrying references to it. Two shapes make the binder's
		// locals table an incomplete answer on its own:
		//
		//   - Redeclaring a name with a conflicting meaning (`var self;
		//     function self() {}`) leaves only the last declaration in the
		//     table; eslint-scope keeps one variable holding both.
		//   - `export import self = require("x")` is filed under the module's
		//     exports and never reaches locals at all.
		//
		// Both are recovered from the file-wide declaration index, using the
		// locals tables to decide which scope each recovered declaration
		// belongs to.
		findScopeVariable := func(scopeNode *ast.Node, alias string) ([]*ast.Node, []*ast.Symbol) {
			var decls []*ast.Node
			var symbols []*ast.Symbol
			addSymbol := func(sym *ast.Symbol) {
				if sym != nil && !slices.Contains(symbols, sym) {
					symbols = append(symbols, sym)
				}
			}

			var sym *ast.Symbol
			if ast.IsLocalsContainer(scopeNode) {
				sym = ast.GetLocals(scopeNode)[alias]
			}
			if sym == nil {
				sym = exportedImportEqualsSymbol(scopeNode, alias)
			}
			if sym != nil {
				// The two halves of an exported declaration each hold one of
				// the things this check needs — the local the declarations,
				// the export symbol the references — and eslint-scope has a
				// single variable for both. The export symbol carries the
				// declarations too, so it is the identity to go on with.
				if sym.ExportSymbol != nil {
					sym = sym.ExportSymbol
				}
				decls = append(decls, sym.Declarations...)
				addSymbol(sym)
			}

			for _, decl := range aliasDeclarations[alias] {
				if slices.Contains(decls, decl) || declarationScopeNode(decl, alias) != scopeNode {
					continue
				}
				decls = append(decls, decl)
				addSymbol(decl.Symbol())
			}
			return decls, symbols
		}

		// checkWasAssigned mirrors upstream's checkWasAssigned: it looks up
		// alias among scopeNode's own directly-declared bindings and, unless
		// already initialized or later assigned `this` in that exact scope,
		// reports every declaration of it. Every local of the scope counts,
		// type-space ones included — typescript-eslint's scope manager makes a
		// `type`/`interface`/`namespace` declaration and a type parameter
		// variables of their scope just like a value declaration.
		checkWasAssigned := func(scopeNode *ast.Node, alias string) {
			decls, symbols := findScopeVariable(scopeNode, alias)
			if len(decls) == 0 {
				return
			}

			for _, decl := range decls {
				if declarationHasInitializer(decl) {
					return
				}
			}

			if ctx.Refs != nil {
				for _, sym := range symbols {
					for _, ref := range ctx.Refs.References(sym) {
						if referenceScopeNode(ref) != scopeNode {
							continue
						}
						if isThisAssignmentTarget(ref) {
							return
						}
					}
				}
			}

			for _, decl := range decls {
				reportRange(declarationTextRange(ctx.SourceFile, decl), aliasNotAssignedToThisMessage(alias))
			}
		}

		ensureWasAssigned := func(scopeNode *ast.Node) {
			for _, alias := range aliases {
				checkWasAssigned(scopeNode, alias)
			}
		}

		flush := func() {
			ensureWasAssigned(ctx.SourceFile.AsNode())
			sort.SliceStable(diagnostics, func(i, j int) bool {
				return diagnostics[i].textRange.Pos() < diagnostics[j].textRange.Pos()
			})
			for _, d := range diagnostics {
				ctx.ReportRange(d.textRange, d.message)
			}
		}

		listeners := rule.RuleListeners{
			ast.KindVariableDeclaration: func(node *ast.Node) {
				decl := node.AsVariableDeclaration()
				if decl.Initializer == nil {
					return
				}
				name := decl.Name()
				if name.Kind != ast.KindIdentifier {
					return
				}
				checkAssignment(node, name.Text(), decl.Initializer, false)
			},
			ast.KindBinaryExpression: func(node *ast.Node) {
				bin := node.AsBinaryExpression()
				if !ast.IsAssignmentOperator(bin.OperatorToken.Kind) {
					return
				}
				// An assignment written inside a destructuring target is a
				// default value, which ESTree parses as an AssignmentPattern
				// rather than an AssignmentExpression: `[self = 1] = this`
				// assigns `this`, not `1`.
				if rootAssignmentOfTarget(node) != nil {
					return
				}
				left := ast.SkipParentheses(bin.Left)
				if left.Kind != ast.KindIdentifier {
					return
				}
				checkAssignment(node, left.Text(), bin.Right, bin.OperatorToken.Kind != ast.KindEqualsToken)
			},
		}
		for _, kind := range scopeVariableKinds {
			previous := listeners[kind]
			listeners[kind] = func(node *ast.Node) {
				if previous != nil {
					previous(node)
				}
				collectDeclaration(node)
			}
		}
		for _, kind := range scopeKinds {
			listeners[rule.ListenerOnExit(kind)] = func(node *ast.Node) {
				// A signature with no body is a TSDeclareFunction or a
				// TSEmptyBodyFunctionExpression upstream — an overload, an
				// ambient declaration, an abstract method — and neither of
				// those is a FunctionDeclaration or a FunctionExpression, so
				// upstream's exit listeners never fire for one.
				if node.Body() == nil {
					return
				}
				ensureWasAssigned(node)
			}
		}

		// There is no listener that fires once after the whole file has been
		// walked (rslint never visits the SourceFile node itself), so the
		// Program-scope check — and the final sort — piggybacks on the exit
		// of the file's last top-level statement instead, the same technique
		// consistent-return uses for its own end-of-file pass.
		statements := ctx.SourceFile.Statements
		if statements == nil || len(statements.Nodes) == 0 {
			return listeners
		}
		lastTopLevelNode := statements.Nodes[len(statements.Nodes)-1]
		exitKind := rule.ListenerOnExit(lastTopLevelNode.Kind)
		previous := listeners[exitKind]
		listeners[exitKind] = func(node *ast.Node) {
			if previous != nil {
				previous(node)
			}
			if node == lastTopLevelNode {
				flush()
			}
		}

		return listeners
	},
}

// declarationTextRange is the range upstream reports a declaration at.
// ESTree wraps an exported declaration in an ExportNamedDeclaration /
// ExportDefaultDeclaration whose inner declaration begins past the `export`
// and `default` keywords, while tsgo keeps both as modifiers of the
// declaration itself, so they are skipped here. Every other modifier
// (`declare`, `abstract`, ...) is part of ESTree's node and is kept.
func declarationTextRange(sourceFile *ast.SourceFile, decl *ast.Node) core.TextRange {
	// eslint-scope records a parameter's definition node as the function the
	// parameter belongs to rather than the parameter itself, so that is what
	// upstream reports — for a destructured parameter too, whose names are
	// definitions of the same shape.
	if root := enclosingParameter(decl); root != nil {
		fn := root.Parent
		switch fn.Kind {
		case ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			// A member's function is the FunctionExpression ESTree hangs off
			// it, which holds only the signature and the body — the name and
			// the modifiers stay on the member.
			return core.NewTextRange(memberSignatureStart(sourceFile, fn), fn.End())
		}
		decl = fn
	}

	// A default import's ESTree node is the ImportDefaultSpecifier, which is
	// the name alone. tsgo's ImportClause additionally spans the `type`
	// keyword of a type-only import, which upstream leaves on the enclosing
	// ImportDeclaration, so the name is reported directly.
	if decl.Kind == ast.KindImportClause {
		if name := decl.Name(); name != nil {
			decl = name
		}
	}
	pos := decl.Pos()
	if modifiers := decl.Modifiers(); modifiers != nil {
		for _, modifier := range modifiers.Nodes {
			if modifier.Kind != ast.KindExportKeyword && modifier.Kind != ast.KindDefaultKeyword {
				break
			}
			pos = modifier.End()
		}
	}
	return scanner.GetRangeOfTokenAtPosition(sourceFile, pos).WithEnd(decl.End())
}

// referenceScopeNode is the scope a reference belongs to, standing in for
// upstream's `reference.from`. It is the nearest enclosing block-scope
// container, with three corrections where tsgo's containers and
// eslint-scope's scopes disagree: a loop whose header declares no
// block-scoped binding is walked past, an object-literal member is walked
// past for a reference in its computed key, and a `with` body or an `enum`
// body — no container of tsgo's at all — is a scope of its own.
func referenceScopeNode(ref *ast.Node) *ast.Node {
	scopeNode := ast.GetEnclosingBlockScopeContainer(ref)
	for scopeNode != nil && (isLoopWithoutOwnScope(scopeNode) || isComputedKeyOfObjectLiteralMember(scopeNode, ref)) {
		scopeNode = ast.GetEnclosingBlockScopeContainer(scopeNode)
	}
	if unmapped := enclosingUnmappedScope(ref, scopeNode); unmapped != nil {
		return unmapped
	}
	return scopeNode
}

// isComputedKeyOfObjectLiteralMember reports whether ref sits in the computed
// key of node, an object-literal method or accessor. tsgo counts such a key
// as part of the member that it names, while eslint-scope evaluates it in the
// scope the object literal itself is written in — the member's own scope
// holds only its body. A class member's computed key is left alone: it does
// belong to a scope of its own, the class scope, which is likewise never the
// enclosing function's.
func isComputedKeyOfObjectLiteralMember(node *ast.Node, ref *ast.Node) bool {
	switch node.Kind {
	case ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
	default:
		return false
	}
	if node.Parent == nil || node.Parent.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	name := node.Name()
	if name == nil || name.Kind != ast.KindComputedPropertyName {
		return false
	}
	return name.Pos() <= ref.Pos() && ref.End() <= name.End()
}

// enclosingUnmappedScope returns the innermost `with` statement, `enum`
// declaration, or class holding ref, looking no further out than stopNode.
// eslint-scope opens a scope for a `with` body, for an enum body, and for a
// class, none of which tsgo counts as a container at all, so without this a
// reference written directly under one — `with (obj) self = this`,
// `enum E { A = (self = this, 1) }`, `class C extends (self = this, Object) {}`
// — would be attributed to the enclosing function.
func enclosingUnmappedScope(ref *ast.Node, stopNode *ast.Node) *ast.Node {
	for node := ref; node != nil && node != stopNode; node = node.Parent {
		parent := node.Parent
		if parent == nil {
			break
		}
		switch parent.Kind {
		case ast.KindWithStatement:
			// A `with` statement's object expression is evaluated in the
			// enclosing scope, so only its body is matched.
			if parent.AsWithStatement().Statement == node {
				return parent
			}
		case ast.KindEnumDeclaration:
			// An enum's own name is bound in the enclosing scope, so only
			// its members are matched.
			if parent.Name() != node {
				return parent
			}
		case ast.KindClassDeclaration, ast.KindClassExpression:
			// A class's name is bound in the enclosing scope as well, and its
			// decorators are evaluated there; the heritage clauses and the
			// class body are the class scope's own.
			if parent.Name() != node && node.Kind != ast.KindDecorator {
				return parent
			}
		}
	}
	return nil
}

// enclosingParameter returns the parameter decl declares a name of, walking
// out of the binding pattern of a destructured parameter, or nil if decl is
// not a parameter name at all.
func enclosingParameter(decl *ast.Node) *ast.Node {
	for node := decl; node != nil; node = node.Parent {
		switch node.Kind {
		case ast.KindParameter:
			return node
		case ast.KindBindingElement, ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		default:
			return nil
		}
	}
	return nil
}

// memberSignatureStart is where the FunctionExpression a class or
// object-literal member holds begins: at the member's type parameter list if
// it has one and at its parameter list otherwise, past the modifiers, the
// name and any optional marker.
func memberSignatureStart(sourceFile *ast.SourceFile, member *ast.Node) int {
	pos := member.Pos()
	if name := member.Name(); name != nil {
		pos = name.End()
	}
	for {
		tokenRange := scanner.GetRangeOfTokenAtPosition(sourceFile, pos)
		if tokenRange.End() <= pos {
			return member.Pos()
		}
		switch sourceFile.Text()[tokenRange.Pos():tokenRange.End()] {
		case "(", "<":
			return tokenRange.Pos()
		}
		pos = tokenRange.End()
	}
}

// rootAssignmentOfTarget returns the assignment ref is a target of, or nil if
// ref is not an assignment target at all. A destructured name reaches its
// assignment through the pattern it is written in, and an assignment nested
// inside that pattern is a default value rather than the assignment itself
// (`[self = 1] = this`), so the walk keeps going and the outermost assignment
// found wins — the one eslint-scope records as the write. Parenthesized
// wrappers, which ESTree has no node for but tsgo does, are walked past on the
// way.
func rootAssignmentOfTarget(ref *ast.Node) *ast.Node {
	var root *ast.Node
	for node := ref; node.Parent != nil; node = node.Parent {
		parent := node.Parent
		switch parent.Kind {
		case ast.KindParenthesizedExpression,
			ast.KindArrayLiteralExpression,
			ast.KindObjectLiteralExpression,
			ast.KindSpreadElement,
			ast.KindSpreadAssignment,
			ast.KindShorthandPropertyAssignment:
		case ast.KindPropertyAssignment:
			// `({ self: x } = this)` writes to x, not to the key `self`.
			if parent.AsPropertyAssignment().Initializer != node {
				return root
			}
		case ast.KindBinaryExpression:
			bin := parent.AsBinaryExpression()
			if !ast.IsAssignmentOperator(bin.OperatorToken.Kind) ||
				ast.SkipParentheses(bin.Left) != ast.SkipParentheses(node) {
				return root
			}
			root = parent
		default:
			return root
		}
	}
	return root
}

// declarationScopeNode is the scope decl's binding belongs to: the nearest
// enclosing container whose locals table binds alias. Reading it off the
// binder's tables rather than off decl's own nesting keeps hoisting right —
// a `var` written inside a block binds in the enclosing function — and keeps
// a declaration the binder dropped attributed to the same scope as the
// declaration that displaced it.
func declarationScopeNode(decl *ast.Node, alias string) *ast.Node {
	for node := decl.Parent; node != nil; node = node.Parent {
		if !ast.IsLocalsContainer(node) {
			continue
		}
		if _, ok := ast.GetLocals(node)[alias]; ok {
			return node
		}
	}
	return nil
}

// exportedImportEqualsSymbol is the `export import X = require("y")` binding
// of scopeNode, which the binder files under the scope's exports and leaves
// out of its locals entirely. Only an import-equals declaration is matched:
// the other export shapes sharing that table — `export { X } from "y"`,
// `export * as X from "y"` — bind no local name, and eslint-scope gives them
// no variable either.
func exportedImportEqualsSymbol(scopeNode *ast.Node, alias string) *ast.Symbol {
	symbol := scopeNode.Symbol()
	if symbol == nil {
		return nil
	}
	sym := symbol.Exports[alias]
	if sym == nil || len(sym.Declarations) == 0 {
		return nil
	}
	for _, decl := range sym.Declarations {
		if decl.Kind != ast.KindImportEqualsDeclaration {
			return nil
		}
	}
	return sym
}

// isLoopWithoutOwnScope reports whether node is a loop statement eslint-scope
// gives no scope of its own, which is any whose header declares nothing or
// declares with `var`.
func isLoopWithoutOwnScope(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
	default:
		return false
	}
	initializer := node.Initializer()
	if initializer == nil || initializer.Kind != ast.KindVariableDeclarationList {
		return true
	}
	return initializer.Flags&ast.NodeFlagsBlockScoped == 0
}

// declarationHasInitializer reports whether decl — one of a symbol's
// Declarations — was given a value at the point it was declared. For a plain
// `var x = v` declarator this is decl's own Initializer. A destructured
// binding (`var {x} = v` / `var [x] = v`) has no Initializer of its own in
// tsgo — unlike upstream, where eslint-scope's `def.node` for a destructured
// name is the whole enclosing VariableDeclarator — so it walks up to that
// declarator and checks its Initializer instead, matching upstream's
// leniency: any declarator init counts, even one that isn't `this`.
func declarationHasInitializer(decl *ast.Node) bool {
	switch decl.Kind {
	case ast.KindVariableDeclaration:
		return decl.AsVariableDeclaration().Initializer != nil
	case ast.KindBindingElement:
		enclosing := utils.EnclosingVariableDeclarationOfBindingElement(decl)
		return enclosing != nil && enclosing.AsVariableDeclaration().Initializer != nil
	default:
		return false
	}
}

func parseOptions(options []any) []string {
	if len(options) == 0 {
		return []string{"that"}
	}
	aliases := make([]string, 0, len(options))
	for _, opt := range options {
		if s, ok := opt.(string); ok {
			aliases = append(aliases, s)
		}
	}
	return aliases
}
