package init_declarations

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed init_declarations.schema.json
var schemaJSON []byte

// InitDeclarationsRule mirrors @typescript-eslint/init-declarations, which wraps
// the ESLint core init-declarations rule and additionally:
//   - Skips `declare const|let|var` declarations.
//   - Skips bindings inside an ancestor `declare namespace { ... }` (or any
//     other ambient `ModuleDeclaration` — `declare module 'm'`,
//     `declare global`).
//   - When reporting on a declarator whose `Initializer == nil`, narrows the
//     diagnostic range to just the identifier (excluding the type annotation),
//     matching typescript-eslint's `getReportLoc`.
//
// https://typescript-eslint.io/rules/init-declarations
// Upstream wrapper: packages/eslint-plugin/src/rules/init-declarations.ts
// Upstream base rule: eslint/lib/rules/init-declarations.js
var InitDeclarationsRule = rule.CreateRule(rule.Rule{
	Name:   "init-declarations",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		return rule.RuleListeners{
			// Listen on VariableDeclarationList rather than VariableStatement:
			// for-loop initializers (`for (var i = 0; ...)`) are direct children
			// of ForStatement / ForInStatement / ForOfStatement and have no
			// enclosing VariableStatement.
			ast.KindVariableDeclarationList: func(node *ast.Node) {
				checkVariableDeclarationList(ctx, node, opts)
			},
		}
	},
})

type initDeclarationsOptions struct {
	mode              string // "always" or "never"
	ignoreForLoopInit bool
}

// parseOptions reads ESLint's `["<mode>", { ...sub-options }]` options array:
// the mode string in slot 0 (defaulting to "always"), and the sub-option
// object — only meaningful for "never" — in slot 1.
func parseOptions(options []any) initDeclarationsOptions {
	opts := initDeclarationsOptions{mode: "always"}

	if len(options) == 0 {
		return opts
	}

	if modeStr, _ := options[0].(string); modeStr == "always" || modeStr == "never" {
		opts.mode = modeStr
	}
	if len(options) > 1 {
		subOpts, _ := options[1].(map[string]interface{})
		if b, ok := subOpts["ignoreForLoopInit"].(bool); ok {
			opts.ignoreForLoopInit = b
		}
	}

	return opts
}

func checkVariableDeclarationList(ctx rule.RuleContext, node *ast.Node, opts initDeclarationsOptions) {
	declList := node.AsVariableDeclarationList()
	if declList == nil || declList.Declarations == nil {
		return
	}

	parent := node.Parent
	if parent == nil {
		return
	}

	if utils.IsInAmbientContext(node) {
		return
	}

	// CONSTANT_BINDINGS in upstream = {const, using, await using}. They require
	// an initializer at parse time, so "never" mode must never report them as
	// `notInitialized`. utils.GetVarDeclListKind centralizes the
	// `await using = NodeFlagsConst|NodeFlagsUsing` encoding so we don't have
	// to repeat it here.
	kind := utils.GetVarDeclListKind(node)
	isConstantBinding := kind == "const" || kind == "using" || kind == "await using"

	inForLoop := isForLoopParent(parent)

	for _, decl := range declList.Declarations.Nodes {
		varDecl := decl.AsVariableDeclaration()
		if varDecl == nil {
			continue
		}

		nameNode := varDecl.Name()
		// Upstream only reports for `id.type === "Identifier"`; destructuring
		// patterns (`{a} = ...`, `[a] = ...`) are silently skipped.
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			continue
		}
		idName := nameNode.AsIdentifier().Text

		hasExplicitInit := varDecl.Initializer != nil
		// Mirror ESLint's `isInitialized`: for-loop bindings are considered
		// initialized regardless of `Initializer` — `for (var i;;)` should not
		// trip "always", and `for (var x in arr)` should still trip "never".
		initialized := hasExplicitInit || inForLoop

		var messageId string
		switch {
		case opts.mode == "always" && !initialized:
			messageId = "initialized"
		case opts.mode == "never" && initialized && !isConstantBinding:
			if opts.ignoreForLoopInit && inForLoop {
				continue
			}
			messageId = "notInitialized"
		}
		if messageId == "" {
			continue
		}

		msg := buildMessage(messageId, idName)

		// Mirror typescript-eslint's `getReportLoc`: when the declarator has no
		// explicit init, narrow the report to the identifier so the diagnostic
		// doesn't underline a trailing type annotation. This covers BOTH paths
		// that produce a no-init report:
		//   - "always" + !initialized → "initialized"
		//   - "never" + in-for-loop + Initializer==nil → "notInitialized"
		//     (e.g. `for (var x: T;;)` or `for (var x in arr)`)
		if !hasExplicitInit {
			ctx.ReportNode(nameNode, msg)
		} else {
			// Declarator has an init expression — report the full declarator
			// (id + type + init) to match upstream's diagnostic ranges.
			ctx.ReportNode(decl, msg)
		}
	}
}

func buildMessage(messageId, idName string) rule.RuleMessage {
	var desc string
	if messageId == "initialized" {
		desc = fmt.Sprintf("Variable '%s' should be initialized on declaration.", idName)
	} else {
		desc = fmt.Sprintf("Variable '%s' should not be initialized on declaration.", idName)
	}
	return rule.RuleMessage{
		Id:          messageId,
		Description: desc,
		Data:        map[string]string{"idName": idName},
	}
}

// isForLoopParent reports whether the VariableDeclarationList's parent is a
// for-loop statement. A VariableDeclarationList can only sit in the
// initializer / left slot of these three kinds, so reaching this state already
// implies the declaration list IS the loop binding.
func isForLoopParent(parent *ast.Node) bool {
	switch parent.Kind {
	case ast.KindForStatement, ast.KindForInStatement, ast.KindForOfStatement:
		return true
	}
	return false
}
