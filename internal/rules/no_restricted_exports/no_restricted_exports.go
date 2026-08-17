package no_restricted_exports

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

//go:embed no_restricted_exports.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/no-restricted-exports

type restrictDefaultExportsOptions struct {
	direct        bool
	named         bool
	defaultFrom   bool
	namedFrom     bool
	namespaceFrom bool
}

type options struct {
	restrictedNames        map[string]bool
	restrictedNamesPattern *esregexp.RegExp
	restrictDefaultExports *restrictDefaultExportsOptions
}

func parseOptions(optionsList []any) options {
	out := options{}
	if len(optionsList) == 0 {
		return out
	}
	optsMap, _ := optionsList[0].(map[string]any)
	if optsMap == nil {
		return out
	}

	if names, ok := optsMap["restrictedNamedExports"].([]any); ok {
		out.restrictedNames = make(map[string]bool, len(names))
		for _, n := range names {
			if s, ok := n.(string); ok {
				out.restrictedNames[s] = true
			}
		}
	}

	if pattern, ok := optsMap["restrictedNamedExportsPattern"].(string); ok && pattern != "" {
		if re, err := esregexp.Compile(pattern, "u"); err == nil {
			out.restrictedNamesPattern = re
		}
	}

	if rde, ok := optsMap["restrictDefaultExports"].(map[string]any); ok {
		restrict := &restrictDefaultExportsOptions{}
		if v, ok := rde["direct"].(bool); ok {
			restrict.direct = v
		}
		if v, ok := rde["named"].(bool); ok {
			restrict.named = v
		}
		if v, ok := rde["defaultFrom"].(bool); ok {
			restrict.defaultFrom = v
		}
		if v, ok := rde["namedFrom"].(bool); ok {
			restrict.namedFrom = v
		}
		if v, ok := rde["namespaceFrom"].(bool); ok {
			restrict.namespaceFrom = v
		}
		out.restrictDefaultExports = restrict
	}

	return out
}

func restrictedNamedMessage(name string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "restrictedNamed",
		Description: fmt.Sprintf("'%s' is restricted from being used as an exported name.", name),
	}
}

var restrictedDefaultMessage = rule.RuleMessage{
	Id:          "restrictedDefault",
	Description: "Exporting 'default' is restricted.",
}

// isRestrictedName reports whether name is disallowed by either the exact
// restrictedNamedExports set or the restrictedNamedExportsPattern regex.
// The pattern never applies to "default" — restrictDefaultExports is the
// dedicated mechanism for that name.
func (o options) isRestrictedName(name string) bool {
	if name != "default" && o.restrictedNamesPattern.Test(name) {
		return true
	}
	return o.restrictedNames[name]
}

// checkNamed handles an exported name that can never literally be "default"
// (a function/class id or a declared variable binding — "default" is a
// reserved word and cannot be an identifier), so only the restrictedNamed
// path applies.
func (o options) checkNamed(ctx rule.RuleContext, nameNode *ast.Node) {
	name, ok := utils.GetStaticPropertyName(nameNode)
	if !ok {
		return
	}
	if o.isRestrictedName(name) {
		ctx.ReportNode(nameNode, restrictedNamedMessage(name))
	}
}

// checkExportAllName handles the exported alias of `export * as name from
// 'mod'`, where a "default" alias is governed by restrictDefaultExports.namespaceFrom.
func (o options) checkExportAllName(ctx rule.RuleContext, nameNode *ast.Node) {
	name, ok := utils.GetStaticPropertyName(nameNode)
	if !ok {
		return
	}
	if o.isRestrictedName(name) {
		ctx.ReportNode(nameNode, restrictedNamedMessage(name))
		return
	}
	if name == "default" && o.restrictDefaultExports != nil && o.restrictDefaultExports.namespaceFrom {
		ctx.ReportNode(nameNode, restrictedDefaultMessage)
	}
}

// checkSpecifierName handles a `export { local as exported }` (optionally
// `from 'mod'`) specifier. localNode is the specifier's local/source-side
// name (the property name before "as", or the exported name itself when
// there is no "as"). isSourceSpecified is true for re-exports (`from 'mod'`).
func (o options) checkSpecifierName(ctx rule.RuleContext, nameNode *ast.Node, localNode *ast.Node, isSourceSpecified bool) {
	name, ok := utils.GetStaticPropertyName(nameNode)
	if !ok {
		return
	}
	if o.isRestrictedName(name) {
		ctx.ReportNode(nameNode, restrictedNamedMessage(name))
		return
	}
	if name != "default" || o.restrictDefaultExports == nil {
		return
	}
	if !isSourceSpecified {
		if o.restrictDefaultExports.named {
			ctx.ReportNode(nameNode, restrictedDefaultMessage)
		}
		return
	}
	localName, _ := utils.GetStaticPropertyName(localNode)
	if (localName == "default" && o.restrictDefaultExports.defaultFrom) ||
		(localName != "default" && o.restrictDefaultExports.namedFrom) {
		ctx.ReportNode(nameNode, restrictedDefaultMessage)
	}
}

// checkNamedExportsClause walks `export { a, b as c }` (optionally
// `from 'mod'`), reporting each specifier's exported name.
func checkNamedExportsClause(ctx rule.RuleContext, opts options, namedExports *ast.NamedExports, isSourceSpecified bool) {
	if namedExports.Elements == nil {
		return
	}
	for _, specifierNode := range namedExports.Elements.Nodes {
		if specifierNode.Kind != ast.KindExportSpecifier {
			continue
		}
		spec := specifierNode.AsExportSpecifier()
		exportedNode := spec.Name()
		localNode := spec.PropertyName
		if localNode == nil {
			localNode = exportedNode
		}
		opts.checkSpecifierName(ctx, exportedNode, localNode, isSourceSpecified)
	}
}

// checkDefaultExportedDeclaration handles `export default function foo() {}`,
// `export default class Foo {}` and `export default interface Foo {}` — the
// declaration forms of a default export, represented as a Function/Class/
// InterfaceDeclaration carrying both the export and default modifiers rather
// than as an ExportAssignment. Upstream reports an ExportDefaultDeclaration
// for what it declares without looking inside, so a bodiless declaration such
// as `export default function foo(): void;` is reported like any other.
func checkDefaultExportedDeclaration(ctx rule.RuleContext, opts options, node *ast.Node) {
	if opts.restrictDefaultExports != nil && opts.restrictDefaultExports.direct {
		ctx.ReportNode(node, restrictedDefaultMessage)
	}
}

// checkNamedExportedDeclaration handles `export function foo() {}` and
// `export class Foo {}` — named (non-default) function/class declarations.
// TypeScript's forgiving parser can produce a nameless declaration here for
// invalid input (e.g. `export class extends Base {}`); guard against that
// rather than reporting on a missing name.
//
// A function declaration with no body — an overload signature, a `declare
// function`, or any function in a `.d.ts` — is a TSDeclareFunction upstream,
// a type the declaration.type switch does not match, so it goes unchecked. A
// class keeps its ClassDeclaration type whether or not it is declared, so
// `export declare class Foo {}` stays checked.
func checkNamedExportedDeclaration(ctx rule.RuleContext, opts options, node *ast.Node) {
	if node.Kind == ast.KindFunctionDeclaration && node.AsFunctionDeclaration().Body == nil {
		return
	}
	name := node.Name()
	if name == nil {
		return
	}
	opts.checkNamed(ctx, name)
}

var NoRestrictedExportsRule = rule.Rule{
	Name:   "no-restricted-exports",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, optionsList []any) rule.RuleListeners {
		opts := parseOptions(optionsList)
		if len(opts.restrictedNames) == 0 && opts.restrictedNamesPattern == nil && opts.restrictDefaultExports == nil {
			return nil
		}

		checkFunctionOrClassDeclaration := func(node *ast.Node) {
			flags := node.ModifierFlags()
			if flags&ast.ModifierFlagsExport == 0 {
				return
			}
			if flags&ast.ModifierFlagsDefault != 0 {
				checkDefaultExportedDeclaration(ctx, opts, node)
				return
			}
			checkNamedExportedDeclaration(ctx, opts, node)
		}

		// An interface is the one declaration kind besides a function and a
		// class that TypeScript lets carry the default modifier, and the only
		// one reachable here: `export default enum`, `export default type` and
		// `export default namespace` are all parse errors. Its named form
		// (`export interface Foo {}`) is a TS-only declaration type upstream
		// leaves unchecked, so only the default modifier reports.
		checkDefaultInterfaceDeclaration := func(node *ast.Node) {
			flags := node.ModifierFlags()
			if flags&ast.ModifierFlagsExport == 0 || flags&ast.ModifierFlagsDefault == 0 {
				return
			}
			checkDefaultExportedDeclaration(ctx, opts, node)
		}

		return rule.RuleListeners{
			// `export * from 'mod'` / `export * as name from 'mod'` /
			// `export { a, b as c }` / `export { a, b as c } from 'mod'`.
			ast.KindExportDeclaration: func(node *ast.Node) {
				exportDecl := node.AsExportDeclaration()
				if exportDecl.ExportClause == nil {
					// Bare `export * from 'mod'` re-exports everything but
					// has no exported name of its own to check.
					return
				}
				isSourceSpecified := exportDecl.ModuleSpecifier != nil
				switch exportDecl.ExportClause.Kind {
				case ast.KindNamespaceExport:
					opts.checkExportAllName(ctx, exportDecl.ExportClause.AsNamespaceExport().Name())
				case ast.KindNamedExports:
					checkNamedExportsClause(ctx, opts, exportDecl.ExportClause.AsNamedExports(), isSourceSpecified)
				}
			},
			// `export default <expression>`, e.g. `export default foo;`,
			// `export default 42;`, or `export default (function foo() {});`.
			ast.KindExportAssignment: func(node *ast.Node) {
				exportAssignment := node.AsExportAssignment()
				if exportAssignment.IsExportEquals {
					// `export = foo;` is a TS-specific CommonJS-style export,
					// not an ES default export.
					return
				}
				checkDefaultExportedDeclaration(ctx, opts, node)
			},
			// `export function foo() {}` / `export default function foo() {}`.
			ast.KindFunctionDeclaration: checkFunctionOrClassDeclaration,
			// `export class Foo {}` / `export default class Foo {}`.
			ast.KindClassDeclaration: checkFunctionOrClassDeclaration,
			// `export default interface Foo {}`.
			ast.KindInterfaceDeclaration: checkDefaultInterfaceDeclaration,
			// `export var/let/const a = 1, { b } = obj, [c] = arr;`.
			ast.KindVariableStatement: func(node *ast.Node) {
				if node.ModifierFlags()&ast.ModifierFlagsExport == 0 {
					return
				}
				declList := node.AsVariableStatement().DeclarationList
				if declList == nil || !ast.IsVariableDeclarationList(declList) {
					return
				}
				for _, decl := range declList.AsVariableDeclarationList().Declarations.Nodes {
					if !ast.IsVariableDeclaration(decl) {
						continue
					}
					name := decl.AsVariableDeclaration().Name()
					if name == nil {
						continue
					}
					utils.CollectBindingNames(name, func(ident *ast.Node, _ string) {
						opts.checkNamed(ctx, ident)
					})
				}
			},
		}
	},
}
