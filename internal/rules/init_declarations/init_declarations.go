package init_declarations

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed init_declarations.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/init-declarations
//
// Upstream listens on VariableDeclaration/TSModuleDeclaration ESTree nodes and
// tracks ambient-namespace nesting with a manually toggled boolean. rslint
// listens on the tsgo-equivalent VariableDeclarationList instead — a
// VariableDeclarationList sits directly under ForStatement / ForInStatement /
// ForOfStatement for loop initializers (no enclosing VariableStatement), which
// is why the loop-shape checks below inspect its parent rather than a
// grandparent — and uses utils.IsInAmbientContext, which the binder derives
// from `declare` modifiers and .d.ts files, instead of a hand-rolled
// enter/exit toggle over TSModuleDeclaration.
var InitDeclarationsRule = rule.Rule{
	Name:   "init-declarations",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				checkVariableDeclarationList(ctx, node, opts)
			},
		}
	},
}

type initDeclarationsOptions struct {
	mode              string // "always" or "never"
	ignoreForLoopInit bool
}

// parseOptions reads ESLint's `["<mode>", { ...sub-options }]` options array:
// the mode string in slot 0 (defaulting to "always", matching upstream's
// `meta.defaultOptions`), and the sub-option object — only meaningful for
// "never" — in slot 1.
func parseOptions(options []any) initDeclarationsOptions {
	opts := initDeclarationsOptions{mode: "always"}

	if len(options) == 0 {
		return opts
	}

	if modeStr, _ := options[0].(string); modeStr == "always" || modeStr == "never" {
		opts.mode = modeStr
	}
	if len(options) > 1 {
		if subOpts, ok := options[1].(map[string]any); ok {
			if b, ok := subOpts["ignoreForLoopInit"].(bool); ok {
				opts.ignoreForLoopInit = b
			}
		}
	}

	return opts
}

func checkVariableDeclarationList(ctx rule.RuleContext, node *ast.Node, opts initDeclarationsOptions) {
	declList := node.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return
	}

	// Upstream: `if (node.declare || insideDeclaredNamespace) return;` — an
	// ambient declaration list (either `declare var/let/const ...` itself, or
	// any declaration nested inside a `declare namespace`/`declare module`/
	// ambient `.d.ts` context) is exempt in both modes.
	if utils.IsInAmbientContext(node) {
		return
	}

	parent := node.Parent
	inForLoop := parent != nil && isForLoopKind(parent.Kind)
	isIgnoredForLoop := opts.ignoreForLoopInit && inForLoop
	kind := utils.GetVarDeclListKind(node)
	isConstantBinding := kind == "const" || kind == "using" || kind == "await using"

	for _, decl := range declList.Declarations.Nodes {
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil {
			continue
		}

		// Upstream only reports for `id.type === "Identifier"`; destructuring
		// patterns (`var {a} = x`, `var [a] = x`) are silently skipped.
		nameNode := varDecl.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			continue
		}

		// Upstream's isInitialized(): a for-loop binding (`for (var i ...)`,
		// `for (var k in x)`, `for (var v of x)`) is always considered
		// initialized, even without an explicit initializer — the loop's
		// init/left slot IS the declaration by AST construction, so upstream's
		// `block.init === declaration` / `block.left === declaration` checks
		// are trivially true whenever a VariableDeclarationList's parent is a
		// for-loop. Otherwise, initialized means an explicit initializer.
		initialized := inForLoop || varDecl.Initializer != nil

		var messageId string
		switch {
		case opts.mode == "always" && !initialized:
			// Note: "always" applies uniformly to every kind, including
			// const/using/await using — those require initializers to parse
			// at all, and an explicit `declare`-only exemption is handled
			// above, so this branch is unreachable for them in practice.
			messageId = "initialized"
		case opts.mode == "never" && !isConstantBinding && initialized && !isIgnoredForLoop:
			messageId = "notInitialized"
		}

		if messageId == "" {
			continue
		}

		idName := nameNode.AsIdentifier().Text
		ctx.ReportNode(decl, buildMessage(messageId, idName))
	}
}

func buildMessage(messageId, idName string) rule.RuleMessage {
	var desc string
	if messageId == "initialized" {
		desc = "Variable '" + idName + "' should be initialized on declaration."
	} else {
		desc = "Variable '" + idName + "' should not be initialized on declaration."
	}
	return rule.RuleMessage{
		Id:          messageId,
		Description: desc,
		Data:        map[string]string{"idName": idName},
	}
}

// isForLoopKind reports whether kind is a for-loop statement kind. A
// VariableDeclarationList can only sit in the initializer / left slot of
// these three kinds, so a parent of one of these kinds already implies the
// declaration list IS the loop binding.
func isForLoopKind(kind ast.Kind) bool {
	switch kind {
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return true
	}
	return false
}
