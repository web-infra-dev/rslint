package no_duplicate_imports

import (
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type declarationKind int

const (
	declImport declarationKind = iota
	declExport
)

// importExportType mirrors ESLint's getImportExportType return values:
//   - typeImportDefault       → "ImportDefaultSpecifier"
//   - typeNamed               → "ImportSpecifier" / "ExportSpecifier"
//   - typeNamespace           → "ImportNamespaceSpecifier" / "ExportNamespaceSpecifier"
//   - typeSideEffectImport    → "SideEffectImport" (`import "foo"`)
//   - typeExportAll           → "ExportAll" (`export * from "foo"` without `as`)
type importExportType int

const (
	typeImportDefault importExportType = iota
	typeNamed
	typeNamespace
	typeSideEffectImport
	typeExportAll
)

type entry struct {
	node       *ast.Node
	declKind   declarationKind
	ieType     importExportType
	isTypeOnly bool
}

type options struct {
	includeExports           bool
	allowSeparateTypeImports bool
}

func parseOptions(raw any) options {
	opts := options{}
	m := utils.GetOptionsMap(raw)
	if m == nil {
		return opts
	}
	if v, ok := m["includeExports"].(bool); ok {
		opts.includeExports = v
	}
	if v, ok := m["allowSeparateTypeImports"].(bool); ok {
		opts.allowSeparateTypeImports = v
	}
	return opts
}

// getImportExportType maps a declaration to the closest ESTree-equivalent
// specifier kind that ESLint's no-duplicate-imports inspects.
//
// For ImportDeclaration: ESLint scans `node.specifiers` and returns the FIRST
// specifier whose type is named or namespace; if none match, the first
// specifier; if no specifiers at all, "SideEffectImport". In tsgo, named or
// namespace specifiers are both reachable through `ImportClause.NamedBindings`,
// so a single check suffices — the named/namespace form takes precedence over
// a default-only import (`import foo, { bar }` reports as named, matching
// ESLint's "first named/namespace wins").
//
// For ExportDeclaration: nil ExportClause → "ExportAll" (`export * from`); a
// NamespaceExport clause → "ExportNamespaceSpecifier" (`export * as ns from`);
// a NamedExports clause → "ExportSpecifier" (`export { x } from`).
func getImportExportType(node *ast.Node) importExportType {
	switch node.Kind {
	case ast.KindImportDeclaration:
		ic := node.AsImportDeclaration().ImportClause
		if ic == nil {
			return typeSideEffectImport
		}
		clause := ic.AsImportClause()
		if clause == nil {
			return typeSideEffectImport
		}
		// `import {} from "mod"` / `import foo, {} from "mod"` — ESLint sees an
		// empty `specifiers` array and falls through to "SideEffectImport" /
		// "ImportDefaultSpecifier" respectively. Treat an empty NamedImports as
		// "no named binding" to match.
		if nb := clause.NamedBindings; nb != nil {
			switch nb.Kind {
			case ast.KindNamespaceImport:
				return typeNamespace
			case ast.KindNamedImports:
				ni := nb.AsNamedImports()
				if ni != nil && ni.Elements != nil && len(ni.Elements.Nodes) > 0 {
					return typeNamed
				}
			}
		}
		if clause.Name() != nil {
			return typeImportDefault
		}
		return typeSideEffectImport
	case ast.KindExportDeclaration:
		ed := node.AsExportDeclaration()
		if ec := ed.ExportClause; ec != nil {
			switch ec.Kind {
			case ast.KindNamespaceExport:
				return typeNamespace
			case ast.KindNamedExports:
				// `export {} from "mod"` — ESLint's `node.specifiers` is empty
				// and `node.type` is "ExportNamedDeclaration" (not
				// "ExportAllDeclaration"), so it returns "SideEffectImport".
				ne := ec.AsNamedExports()
				if ne != nil && ne.Elements != nil && len(ne.Elements.Nodes) > 0 {
					return typeNamed
				}
				return typeSideEffectImport
			}
		}
		return typeExportAll
	}
	return typeSideEffectImport
}

// isWholeDeclTypeOnly mirrors ESLint's `node.importKind === "type"` /
// `node.exportKind === "type"` — a declaration-level marker only, NOT the
// per-specifier `type` keyword (`import { type X }`).
func isWholeDeclTypeOnly(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindImportDeclaration:
		ic := node.AsImportDeclaration().ImportClause
		if ic == nil {
			return false
		}
		return ic.AsImportClause().PhaseModifier == ast.KindTypeKeyword
	case ast.KindExportDeclaration:
		return node.AsExportDeclaration().IsTypeOnly
	}
	return false
}

func canBeMerged(a, b *entry) bool {
	if a.isTypeOnly && b.isTypeOnly {
		isDefault1 := a.ieType == typeImportDefault
		isDefault2 := b.ieType == typeImportDefault
		isNamed1 := a.ieType == typeNamed
		isNamed2 := b.ieType == typeNamed
		if (isDefault1 && isNamed2) || (isDefault2 && isNamed1) {
			return false
		}
	}

	if (a.ieType == typeExportAll && b.ieType != typeExportAll && b.ieType != typeSideEffectImport) ||
		(a.ieType != typeExportAll && a.ieType != typeSideEffectImport && b.ieType == typeExportAll) {
		return false
	}

	if (a.ieType == typeNamespace && b.ieType == typeNamed) ||
		(b.ieType == typeNamespace && a.ieType == typeNamed) {
		return false
	}

	return true
}

func shouldReport(node *entry, previous []*entry, allowSeparateTypeImports bool) bool {
	for _, prev := range previous {
		if allowSeparateTypeImports && node.isTypeOnly != prev.isTypeOnly {
			continue
		}
		if canBeMerged(node, prev) {
			return true
		}
	}
	return false
}

func filterByKind(entries []*entry, kind declarationKind) []*entry {
	out := make([]*entry, 0, len(entries))
	for _, e := range entries {
		if e.declKind == kind {
			out = append(out, e)
		}
	}
	return out
}

// getModule returns the trimmed module specifier for an import/export
// declaration, or "" when no static module specifier is present (e.g.
// `export { foo }` without source, or a non-string-literal specifier).
func getModule(node *ast.Node) string {
	var spec *ast.Node
	switch node.Kind {
	case ast.KindImportDeclaration:
		spec = node.AsImportDeclaration().ModuleSpecifier
	case ast.KindExportDeclaration:
		spec = node.AsExportDeclaration().ModuleSpecifier
	default:
		return ""
	}
	if spec == nil {
		return ""
	}
	return strings.TrimSpace(utils.GetStaticStringValue(spec))
}

func reportMessage(ctx *rule.RuleContext, node *ast.Node, id, module string) {
	var desc string
	switch id {
	case "import":
		desc = fmt.Sprintf("'%s' import is duplicated.", module)
	case "importAs":
		desc = fmt.Sprintf("'%s' import is duplicated as export.", module)
	case "export":
		desc = fmt.Sprintf("'%s' export is duplicated.", module)
	case "exportAs":
		desc = fmt.Sprintf("'%s' export is duplicated as import.", module)
	}
	ctx.ReportNode(node, rule.RuleMessage{Id: id, Description: desc})
}

// https://eslint.org/docs/latest/rules/no-duplicate-imports
var NoDuplicateImportsRule = rule.Rule{
	Name: "no-duplicate-imports",
	Run: func(ctx rule.RuleContext, rawOpts any) rule.RuleListeners {
		opts := parseOptions(rawOpts)
		modules := map[string][]*entry{}

		handle := func(declKind declarationKind) func(node *ast.Node) {
			return func(node *ast.Node) {
				module := getModule(node)
				if module == "" {
					return
				}
				current := &entry{
					node:       node,
					declKind:   declKind,
					ieType:     getImportExportType(node),
					isTypeOnly: isWholeDeclTypeOnly(node),
				}

				if previous, ok := modules[module]; ok {
					importNodes := filterByKind(previous, declImport)
					var exportNodes []*entry
					if opts.includeExports {
						exportNodes = filterByKind(previous, declExport)
					}
					var messageIds []string
					switch declKind {
					case declImport:
						if shouldReport(current, importNodes, opts.allowSeparateTypeImports) {
							messageIds = append(messageIds, "import")
						}
						if opts.includeExports && shouldReport(current, exportNodes, opts.allowSeparateTypeImports) {
							messageIds = append(messageIds, "importAs")
						}
					case declExport:
						if shouldReport(current, exportNodes, opts.allowSeparateTypeImports) {
							messageIds = append(messageIds, "export")
						}
						if shouldReport(current, importNodes, opts.allowSeparateTypeImports) {
							messageIds = append(messageIds, "exportAs")
						}
					}
					for _, id := range messageIds {
						reportMessage(&ctx, node, id, module)
					}
				}

				modules[module] = append(modules[module], current)
			}
		}

		listeners := rule.RuleListeners{
			ast.KindImportDeclaration: handle(declImport),
		}
		if opts.includeExports {
			// In tsgo, ExportDeclaration covers both ESLint's
			// ExportNamedDeclaration (with source) and ExportAllDeclaration.
			// `export const`, `export function`, `export default …` use
			// other kinds (VariableStatement / FunctionDeclaration /
			// ExportAssignment) and have no module specifier, so they're
			// already filtered out by the `module == ""` guard.
			listeners[ast.KindExportDeclaration] = handle(declExport)
		}
		return listeners
	},
}
