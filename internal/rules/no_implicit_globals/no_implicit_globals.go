package no_implicit_globals

import (
	_ "embed"
	"regexp"
	"strings"

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
// module-ness: ast.IsExternalModule (real import/export syntax) combined with
// the resolved language defaults for the file extension (ctx.Refs
// .HasNonGlobalTopLevelScope, which covers .cjs's CommonJS wrapper and
// .js/.mjs's default module treatment) — the same combination no_redeclare
// already uses for its own "is this the Program's global scope" check.
var NoImplicitGlobalsRule = rule.Rule{
	Name:   "no-implicit-globals",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		if ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}
		opts := parseOptions(rawOptions)

		listeners := rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				checkImplicitGlobalWrite(ctx, node)
			},
		}

		hasNonGlobalTopLevelScope := ast.IsExternalModule(ctx.SourceFile) ||
			(ctx.Refs != nil && ctx.Refs.HasNonGlobalTopLevelScope())
		if hasNonGlobalTopLevelScope {
			return listeners
		}

		sourceFileNode := ctx.SourceFile.AsNode()
		exported := parseExportedNames(ctx)

		listeners[ast.KindVariableDeclarationList] = func(node *ast.Node) {
			checkVariableDeclarationList(ctx, node, sourceFileNode, opts, exported)
		}
		listeners[ast.KindFunctionDeclaration] = func(node *ast.Node) {
			checkFunctionDeclaration(ctx, node, sourceFileNode, exported)
		}
		if opts.lexicalBindings {
			listeners[ast.KindClassDeclaration] = func(node *ast.Node) {
				checkClassDeclaration(ctx, node, sourceFileNode, exported)
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
func reportDeclaration(ctx rule.RuleContext, declNode *ast.Node, name string, kind string, lexical bool, exported map[string]bool) {
	if exported[name] {
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

func checkVariableDeclarationList(ctx rule.RuleContext, node *ast.Node, sourceFileNode *ast.Node, opts options, exported map[string]bool) {
	if node.Flags&ast.NodeFlagsAmbient != 0 {
		// `declare var x;` (bare or inside `declare global { ... }`) produces
		// no runtime binding at all, so it can't leak or shadow anything —
		// an ESLint-invisible TypeScript-only shape.
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
			reportDeclaration(ctx, declarator, name, kind, lexical, exported)
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

func checkFunctionDeclaration(ctx rule.RuleContext, node *ast.Node, sourceFileNode *ast.Node, exported map[string]bool) {
	fn := node.AsFunctionDeclaration()
	if fn == nil || node.Body() == nil {
		// Bodyless overload/ambient signatures are a TypeScript-only shape
		// with no ESLint analog; only a real function declaration binds a name.
		return
	}
	nameNode := fn.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return
	}
	if ast.GetEnclosingBlockScopeContainer(node) != sourceFileNode {
		return
	}
	reportDeclaration(ctx, node, nameNode.Text(), "function", false, exported)
}

func checkClassDeclaration(ctx rule.RuleContext, node *ast.Node, sourceFileNode *ast.Node, exported map[string]bool) {
	if node.Flags&ast.NodeFlagsAmbient != 0 {
		// `declare class Foo {}` produces no runtime binding, unlike a
		// bodyless function declaration a class always has a body — an
		// ESLint-invisible TypeScript-only shape.
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
	reportDeclaration(ctx, node, nameNode.Text(), "class", true, exported)
}

// checkImplicitGlobalWrite handles the two reference-driven diagnostics that
// apply file-wide, independently of module-ness: an assignment to a read-only
// global (assignmentToReadonlyGlobal) and an assignment to a name that is
// neither declared anywhere in the file nor a known global at all, outside
// strict-mode code (globalVariableLeak). Both only ever consider a pure `=`
// assignment or bare for-in/for-of loop variable — the same shape ESLint's
// own reference.isWrite() && !reference.isRead() filter selects, which
// excludes compound assignments (foo += 1) and update expressions (foo++).
func checkImplicitGlobalWrite(ctx rule.RuleContext, node *ast.Node) {
	root := findPureAssignmentRoot(node)
	if root == nil {
		return
	}
	if ctx.Refs != nil && ctx.Refs.IsDefinedInFile(node) {
		return
	}
	name := node.Text()
	switch ctx.Globals.Access(name) {
	case utils.GlobalAccessReadonly:
		ctx.ReportNode(root, assignmentToReadonlyGlobalMessage)
	case utils.GlobalAccessWritable:
		// Writable globals may be freely assigned.
	default:
		if !utils.IsInStrictMode(node, ctx.SourceFile) {
			ctx.ReportNode(root, globalVariableLeakMessage)
		}
	}
}

// findPureAssignmentRoot walks from a candidate write-target identifier up
// through destructuring containers and transparent TS wrappers (parens,
// non-null assertions) to the enclosing pure `=` assignment or for-in/for-of
// statement, mirroring upstream's ASSIGNMENT_NODES walk of
// reference.identifier.parent. It returns nil for anything that isn't a pure
// write target: compound/logical assignment operators and update expressions
// read as well as write, so ESLint's scope analysis (and this walk) does not
// treat them as leak/readonly-assignment candidates. It also returns nil when
// the write happens through an opaque TS wrapper (`as`, `<T>`, `satisfies`) —
// ESLint's scope manager only recognizes Identifier/ObjectPattern/ArrayPattern
// (and their spreads) as assignment patterns, so a write reaching an
// AssignmentExpression's Left through one of those wrappers is treated as a
// plain read-write reference instead, the same reasoning behind
// no_global_assign's isWriteThroughTypeAssertion exclusion.
func findPureAssignmentRoot(node *ast.Node) *ast.Node {
	current := node
	for current != nil && current.Parent != nil {
		parent := current.Parent
		switch parent.Kind {
		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary == nil || binary.OperatorToken == nil ||
				binary.OperatorToken.Kind != ast.KindEqualsToken || binary.Left != current {
				return nil
			}
			if utils.IsDefaultValueInDestructuringAssignment(parent) {
				current = parent
				continue
			}
			return parent

		case ast.KindForInStatement, ast.KindForOfStatement:
			stmt := parent.AsForInOrOfStatement()
			if stmt == nil || stmt.Initializer != current {
				return nil
			}
			return parent

		case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression:
			if !utils.IsInDestructuringAssignment(parent) {
				return nil
			}
			current = parent

		case ast.KindShorthandPropertyAssignment:
			shorthand := parent.AsShorthandPropertyAssignment()
			if shorthand == nil || shorthand.Name() != current {
				return nil
			}
			current = parent

		case ast.KindPropertyAssignment:
			propAssignment := parent.AsPropertyAssignment()
			if propAssignment == nil || propAssignment.Initializer != current {
				return nil
			}
			current = parent

		case ast.KindSpreadElement, ast.KindSpreadAssignment,
			ast.KindParenthesizedExpression, ast.KindNonNullExpression:
			current = parent

		default:
			return nil
		}
	}
	return nil
}

var exportedDirectiveJustificationPattern = regexp.MustCompile(`\s-{2,}\s`)

// parseExportedNames collects every name listed by a `/* exported name1,
// name2 */` directive comment, mirroring @eslint/plugin-kit's parseListConfig:
// comma-separated entries, each trimmed and unwrapped of one matching layer of
// quotes. A name in this set is intentionally global (assigned to elsewhere,
// e.g. by a bundler) and is exempt from the declaration-in-global-scope
// diagnostics — but not from the readonly-global or leak diagnostics, which
// upstream keeps unconditional (see reportDeclaration's caller).
func parseExportedNames(ctx rule.RuleContext) map[string]bool {
	text := ctx.SourceFile.Text()
	if !strings.Contains(text, "exported") {
		return nil
	}

	var names map[string]bool
	for _, comment := range ctx.Comments.All() {
		if comment.Kind != ast.KindMultiLineCommentTrivia {
			continue
		}
		start, end := comment.Pos()+2, comment.End()
		if end-start >= 2 && text[end-2:end] == "*/" {
			end -= 2
		}
		if start < 0 || end > len(text) || start > end {
			continue
		}
		content := strings.TrimSpace(text[start:end])
		rest, ok := strings.CutPrefix(content, "exported")
		if !ok {
			continue
		}
		if rest != "" && !isDirectiveSeparator(rest[0]) {
			continue
		}
		if loc := exportedDirectiveJustificationPattern.FindStringIndex(rest); loc != nil {
			rest = rest[:loc[0]]
		}
		for _, part := range strings.Split(rest, ",") {
			name := stripMatchingQuotes(strings.TrimSpace(part))
			if name == "" {
				continue
			}
			if names == nil {
				names = make(map[string]bool)
			}
			names[name] = true
		}
	}
	return names
}

func isDirectiveSeparator(b byte) bool {
	switch b {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	default:
		return false
	}
}

func stripMatchingQuotes(s string) string {
	if len(s) >= 2 {
		first, last := s[0], s[len(s)-1]
		if (first == '\'' || first == '"') && first == last {
			return s[1 : len(s)-1]
		}
	}
	return s
}
