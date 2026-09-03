package no_undef

import (
	_ "embed"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

//go:embed no_undef.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/no-undef
//
// Like ESLint, this rule is deliberately independent of TypeScript's checker
// environment. It resolves declarations in the current file, then applies the
// versioned globals selected by languageOptions and the config/inline globals
// exposed on RuleContext. For TypeScript syntax, it also mirrors the parser's
// default esnext type-space globals. DOM, Node, ambient .d.ts, and cross-file
// declarations are not silently supplied by a TypeChecker.

type options struct {
	checkTypeof bool
}

func parseOptions(rawOptions []any) options {
	result := options{checkTypeof: false}
	if len(rawOptions) == 0 {
		return result
	}
	optsMap, _ := rawOptions[0].(map[string]any)
	if v, ok := optsMap["typeof"].(bool); ok {
		result.checkTypeof = v
	}
	return result
}

var NoUndefRule = rule.Rule{
	Name:   "no-undef",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, rawOptions []any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		if ctx.Refs == nil {
			return rule.RuleListeners{}
		}

		listeners := rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if !scope.IsReferenceIdentifier(node) || (!opts.checkTypeof && isTypeOfOperand(node)) {
					return
				}
				reportUndefinedReference(&ctx, node, node.Text())
			},
		}
		// TSESTree represents bare `<this>` tag names as JSXIdentifiers and
		// records them as references. tsgo preserves `this` as a keyword, so it
		// needs its own listener; member tags such as `<this.Component>` are not
		// references and are rejected by the shared scope classifier.
		if ctx.SourceFile.ScriptKind == core.ScriptKindTSX {
			listeners[ast.KindThisKeyword] = func(node *ast.Node) {
				if scope.IsTypeScriptJsxThisReference(node) {
					reportUndefinedReference(&ctx, node, "this")
				}
			}
		}
		return listeners
	},
}

func reportUndefinedReference(ctx *rule.RuleContext, node *ast.Node, name string) {
	if ctx.Globals.Access(name).IsDeclared() {
		return
	}
	referenceSpace := noUndefReferenceSpace(node)
	// An explicit `off` removes only implicit globals. A declaration or
	// import authored in this file still defines its references.
	if ctx.Refs.ResolveInFileWithMeaning(node, referenceSpace.DeclarationMeaning()) != nil {
		return
	}
	// CommonJS's wrapper-level `arguments` binding is lexical rather than
	// a configurable global, so it remains defined even when a same-named
	// global is disabled.
	if ctx.Refs.HasImplicitWrapperBinding(name) {
		return
	}
	// typescript-eslint's scope manager seeds its default `esnext`
	// library in a separate type namespace. A name such as `Record`
	// therefore resolves in `type T = Record<string, unknown>`, but the
	// same spelling in `Record;` remains undefined. Keep this check after
	// same-file resolution so authored declarations retain normal scope
	// behavior, and do not consult the TypeChecker: ambient, cross-file,
	// and host-library symbols are outside core no-undef's inputs.
	if referenceSpace.IncludesType() && rule.IsDefaultTypeScriptTypeGlobal(name) {
		return
	}

	ctx.ReportNode(node, rule.RuleMessage{
		Id:          "undef",
		Description: "'" + name + "' is not defined.",
	})
}

// noUndefReferenceSpace preserves the rule's established Espree-facing
// behavior for JavaScript while sharing typescript-eslint's scope classifier
// everywhere else. Main has historically treated a parenthesized JavaScript
// default export as a value reference even though TSESTree erases the
// parentheses and assigns both reference capabilities.
func noUndefReferenceSpace(node *ast.Node) scope.ReferenceSpace {
	if ast.IsInJSFile(node) && node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		if exportParent := ast.WalkUpParenthesizedExpressions(node.Parent); exportParent != nil &&
			exportParent.Kind == ast.KindExportAssignment {
			if assignment := exportParent.AsExportAssignment(); assignment != nil &&
				ast.SkipParentheses(assignment.Expression) == node {
				return scope.ReferenceValue
			}
		}
	}
	return scope.ESLintReferenceSpace(node)
}

// isTypeOfOperand checks if the identifier is the sole operand of a typeof
// expression, walking through any ParenthesizedExpression wrappers.
// typeof x      → Identifier.Parent = TypeOfExpression → true
// typeof (x)    → Identifier.Parent = ParenExpr.Parent = TypeOfExpression → true
// typeof (a+b)  → Identifier.Parent = BinaryExpression → false
func isTypeOfOperand(node *ast.Node) bool {
	current := node.Parent
	for current != nil && current.Kind == ast.KindParenthesizedExpression {
		current = current.Parent
	}
	return current != nil && current.Kind == ast.KindTypeOfExpression
}
