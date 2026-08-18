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

		// isSimpleThisAssignment reports whether ref is the left-hand side of
		// a plain `=` assignment whose right-hand side is `this`, matching
		// upstream's `write.type === "ThisExpression" && write.parent.operator === "="`.
		// ref may sit inside one or more parenthesized wrappers (`(that) =
		// this`), which ESTree has no node for but tsgo does — WalkUpParenthesizedExpressions
		// skips past them to find the real parent, same as ESLint sees.
		isSimpleThisAssignment := func(ref *ast.Node) bool {
			parent := ast.WalkUpParenthesizedExpressions(ref.Parent)
			if parent == nil || parent.Kind != ast.KindBinaryExpression {
				return false
			}
			bin := parent.AsBinaryExpression()
			if ast.SkipParentheses(bin.Left) != ref || bin.OperatorToken.Kind != ast.KindEqualsToken {
				return false
			}
			return ast.SkipParentheses(bin.Right).Kind == ast.KindThisKeyword
		}

		// checkWasAssigned mirrors upstream's checkWasAssigned: it looks up
		// alias among scopeNode's own directly-declared bindings and, unless
		// already initialized or later assigned `this` in that exact scope,
		// reports every declaration of it.
		checkWasAssigned := func(scopeNode *ast.Node, alias string) {
			locals := ast.GetLocals(scopeNode)
			if locals == nil {
				return
			}
			sym := locals[alias]
			if sym == nil || sym.Flags&aliasSymbolFlags == 0 {
				return
			}
			// The two halves of an exported declaration each hold one of the
			// things this check needs — the local the declarations, the export
			// symbol the references — and eslint-scope has a single variable
			// for both. The export symbol carries the declarations too, so it
			// is the identity to go on with.
			if sym.ExportSymbol != nil {
				sym = sym.ExportSymbol
			}

			for _, decl := range sym.Declarations {
				if declarationHasInitializer(decl) {
					return
				}
			}

			if ctx.Refs != nil {
				for _, ref := range ctx.Refs.References(sym) {
					if referenceScopeNode(ref) != scopeNode {
						continue
					}
					if isSimpleThisAssignment(ref) {
						return
					}
				}
			}

			for _, decl := range sym.Declarations {
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
				left := ast.SkipParentheses(bin.Left)
				if left.Kind != ast.KindIdentifier {
					return
				}
				checkAssignment(node, left.Text(), bin.Right, bin.OperatorToken.Kind != ast.KindEqualsToken)
			},
		}
		for _, kind := range scopeKinds {
			listeners[rule.ListenerOnExit(kind)] = func(node *ast.Node) {
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

// aliasSymbolFlags are the symbol flags a designated alias's binding can
// carry in a scope's locals. A plain declaration carries SymbolFlagsValue,
// but in a module the binder splits an exported one in two: the export
// symbol keeps the real flags while the local left behind carries only
// SymbolFlagsExportValue, and an import binding carries only
// SymbolFlagsAlias. All three are variables of the scope as eslint-scope
// sees it; a type-only declaration, which carries none of them, is not.
const aliasSymbolFlags = ast.SymbolFlagsValue | ast.SymbolFlagsExportValue | ast.SymbolFlagsAlias

// declarationTextRange is the range upstream reports a declaration at.
// ESTree wraps an exported declaration in an ExportNamedDeclaration /
// ExportDefaultDeclaration whose inner declaration begins past the `export`
// and `default` keywords, while tsgo keeps both as modifiers of the
// declaration itself, so they are skipped here. Every other modifier
// (`declare`, `abstract`, ...) is part of ESTree's node and is kept.
func declarationTextRange(sourceFile *ast.SourceFile, decl *ast.Node) core.TextRange {
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
// container, except that a loop whose header declares no block-scoped
// binding is walked past: tsgo counts every `for`/`for-in`/`for-of`
// statement as a container, while eslint-scope opens a scope for one only
// when its header declares with `let`, `const` or `using`.
func referenceScopeNode(ref *ast.Node) *ast.Node {
	scopeNode := ast.GetEnclosingBlockScopeContainer(ref)
	for scopeNode != nil && isScopelessLoop(scopeNode) {
		scopeNode = ast.GetEnclosingBlockScopeContainer(scopeNode)
	}
	return scopeNode
}

// isScopelessLoop reports whether node is a loop statement eslint-scope
// gives no scope of its own, which is any whose header declares nothing or
// declares with `var`.
func isScopelessLoop(node *ast.Node) bool {
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
