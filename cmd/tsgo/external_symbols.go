package main

import (
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/module"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
)

const globalNamespace = "global"

type externalSymbolKey struct {
	namespace string
	symbolID  ast.SymbolId
}
type externalSymbolNameKey struct {
	externalSymbolKey
	name string
}
type externalSymbolCollector struct {
	tc        *checker.Checker
	semantic  *Semantic
	active    map[externalSymbolKey]bool
	collected map[externalSymbolNameKey]bool
}

// collectExternalSymbols appends shallow global names and qualified external
// declarations to the semantic output.
func collectExternalSymbols(program *compiler.Program, tc *checker.Checker, semantic *Semantic) {
	if program == nil || tc == nil || semantic == nil {
		return
	}
	c := externalSymbolCollector{
		tc:        tc,
		semantic:  semantic,
		active:    make(map[externalSymbolKey]bool),
		collected: make(map[externalSymbolNameKey]bool),
	}
	c.collectGlobals(program)
	c.collectDependencies(program)
}

// collectGlobals records top-level global value names without expanding their types or members.
//
// JavaScript example:
//
//	console.log("hello");
//	Math.abs(-1);
//
// This pass records "console" and "Math" in "global", without adding paths such
// as "console.log" or "Math.abs"; standard library declarations are collected separately.
// Ambient module declarations such as declare module "buffer" are excluded;
// actual global values such as "Buffer" and "process" remain eligible.
func (c *externalSymbolCollector) collectGlobals(program *compiler.Program) {
	files := program.GetSourceFiles()
	if len(files) == 0 {
		return
	}
	// Read scope symbols without inspecting the types of global values.
	// Resolve each name in the global scope to exclude module-local declarations.
	symbols := c.tc.GetSymbolsInScope(files[0].AsNode(), ast.SymbolFlagsValue)
	sort.Slice(symbols, func(i, j int) bool { return symbols[i].Name < symbols[j].Name })
	for _, symbol := range symbols {
		global := c.tc.GetGlobalSymbol(symbol.Name, ast.SymbolFlagsValue, nil)
		global = c.resolveSymbol(global)
		if global == nil || global.IsExternalModule() {
			continue
		}
		c.record(globalNamespace, symbol.Name, global)
	}
}

// collectDependencies collects declarations from every loaded standard library
// and external dependency file, including non-exported declarations.
//
// JavaScript example:
//
//	// app.js
//	import "pkg";
//	// pkg/index.js
//	import "./internal.js";
//	// pkg/internal.js
//	const hidden = 1;
//	export {};
//
// "hidden" is collected under "pkg" even though the entry does not export it;
// the compiler's file classification determines which files are scanned.
func (c *externalSymbolCollector) collectDependencies(program *compiler.Program) {
	for _, file := range program.GetSourceFiles() {
		if program.IsSourceFileDefaultLibrary(file.Path()) {
			c.collectLibrary(file)
			continue
		}
		if !program.IsSourceFileFromExternalLibrary(file) {
			continue
		}
		namespace := externalFileNamespace(program, file)
		c.collectTable(namespace, "", file.Locals)
		if file.Symbol != nil {
			c.collectExports(namespace, "", file.Symbol)
		}
	}
}

// externalFileNamespace reads a dependency's package name from compiler metadata,
// using its file name when the package has no declared name.
func externalFileNamespace(program *compiler.Program, file *ast.SourceFile) string {
	directory := program.GetSourceFileMetaData(file.Path()).PackageJsonDirectory
	if directory == "" {
		directory = program.GetNearestAncestorDirectoryWithPackageJson(tspath.GetDirectoryPath(file.FileName()))
	}
	if info := program.GetPackageJsonInfo(tspath.CombinePaths(directory, "package.json")); info.Exists() {
		if name, ok := info.Contents.Name.GetValue(); ok && name != "" {
			return module.GetPackageNameFromTypesPackageName(name)
		}
	}
	return file.FileName()
}

// collectLibrary collects standard library declarations and names constructor
// members using their explicitly declared prototype relationship.
//
// TypeScript example:
//
//	interface String { slice(start?: number): string; }
//	interface StringConstructor { readonly prototype: String; }
//	declare var String: StringConstructor;
//
// "String.slice" and "String.prototype.slice" share the interface member's ID;
// constructor members also receive names under "String" without walking types.
func (c *externalSymbolCollector) collectLibrary(file *ast.SourceFile) {
	c.collectTable(globalNamespace, "", file.Locals)
	names := make([]string, 0, len(file.Locals))
	for name := range file.Locals {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		symbol := c.resolveSymbol(file.Locals[name])
		if symbol == nil || symbol.Flags&ast.SymbolFlagsInterface == 0 {
			continue
		}
		for _, declaration := range symbol.Declarations {
			if declaration.Kind != ast.KindVariableDeclaration {
				continue
			}
			ty := declaration.Type()
			if ty == nil || ty.Kind != ast.KindTypeReference {
				continue
			}
			constructor := c.resolveSymbol(c.tc.GetSymbolAtLocation(ty.AsTypeReferenceNode().TypeName))
			if constructor == nil {
				continue
			}
			prototype := constructor.Members["prototype"]
			if prototype == nil || prototype.ValueDeclaration == nil {
				continue
			}
			prototypeType := prototype.ValueDeclaration.Type()
			if prototypeType != nil && prototypeType.Kind == ast.KindTypeReference &&
				c.resolveSymbol(c.tc.GetSymbolAtLocation(prototypeType.AsTypeReferenceNode().TypeName)) == symbol {
				c.collectTable(globalNamespace, joinExternalName(name, "prototype"), symbol.Members)
				c.collectTable(globalNamespace, name, constructor.Members)
			}
		}
	}
}

// resolveSymbol returns the canonical merged export or alias target so different
// names share the same symbol ID.
//
// JavaScript example:
//
//	const value = 1;
//	export { value as renamed };
//
// "value" and "renamed" resolve to the same symbol; callers retain each name.
func (c *externalSymbolCollector) resolveSymbol(symbol *ast.Symbol) *ast.Symbol {
	if symbol == nil {
		return nil
	}
	symbol = c.tc.GetMergedSymbol(symbol)
	if symbol.ExportSymbol != nil {
		symbol = symbol.ExportSymbol
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		if target, ok := c.tc.ResolveAlias(symbol); ok {
			return c.tc.GetMergedSymbol(target)
		}
		return nil
	}
	return c.tc.GetMergedSymbol(symbol)
}

// collectTable collects a declaration table in name order, qualifying names with
// the prefix and flattening module wrappers.
//
// JavaScript example:
//
//	const api = { run() {} };
//
// Collecting the initializer's member table with prefix "api" produces "api.run".
func (c *externalSymbolCollector) collectTable(namespace, prefix string, table ast.SymbolTable) {
	names := make([]string, 0, len(table))
	for name := range table {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		symbol := table[name]
		if strings.HasPrefix(name, ast.InternalSymbolNamePrefix) {
			name = ast.SymbolName(symbol)
			if strings.HasPrefix(name, ast.InternalSymbolNamePrefix) {
				continue
			}
		}
		if symbol.IsExternalModule() || name == ast.InternalSymbolNameExportEquals {
			// Ambient module wrappers and export assignments describe the module itself.
			c.collect(namespace, prefix, symbol)
		} else {
			c.collect(namespace, joinExternalName(prefix, name), symbol)
		}
	}
}

// collectExports collects a module's resolved exports under their exported names,
// including re-exports.
//
// JavaScript example:
//
//	export { value as renamed } from "./internal.js";
//
// The exported path is "renamed", while its symbol ID belongs to "value".
func (c *externalSymbolCollector) collectExports(namespace, prefix string, symbol *ast.Symbol) {
	exports := c.tc.GetExportsOfModule(symbol)
	sort.Slice(exports, func(i, j int) bool { return exports[i].Name < exports[j].Name })
	for _, exported := range exports {
		name := prefix
		if exported.Name != ast.InternalSymbolNameExportEquals {
			name = joinExternalName(prefix, exported.Name)
		}
		c.collect(namespace, name, exported)
	}
}

// collect records a symbol and recursively collects its declared members, stopping
// declaration cycles without following value types.
//
// JavaScript example:
//
//	class Foo {
//	  static create() { const local = new Foo(); return local; }
//	  run() {}
//	}
//
// Members include "Foo.create" and "Foo.prototype.run"; function-local "local"
// is excluded. Explicit object and type literals also contribute their written
// members, but type references and function return types do not introduce paths.
//
// TypeScript example of a declaration cycle:
//
//	namespace Loop {
//	  export import Self = Loop;
//	  export const value = 1;
//	}
//
// "Loop.Self" retains the ID of "Loop", but its members are not expanded again
// while "Loop" is being collected, avoiding paths such as "Loop.Self.Self".
func (c *externalSymbolCollector) collect(namespace, name string, symbol *ast.Symbol) {
	if strings.HasPrefix(name, ast.InternalSymbolNamePrefix) {
		return
	}
	symbol = c.resolveSymbol(symbol)
	if symbol == nil || symbol.Flags&ast.SymbolFlagsTypeParameter != 0 {
		return
	}
	c.record(namespace, name, symbol)
	key := externalSymbolKey{namespace, ast.GetSymbolId(symbol)}
	if c.active[key] {
		return
	}
	c.active[key] = true
	defer delete(c.active, key)
	if symbol.Flags&ast.SymbolFlagsModule != 0 {
		c.collectExports(namespace, name, symbol)
		for _, declaration := range symbol.Declarations {
			if declaration.Kind == ast.KindModuleDeclaration {
				c.collectTable(namespace, name, declaration.Locals())
			}
		}
	} else {
		c.collectTable(namespace, name, symbol.Exports)
	}
	membersName := name
	if symbol.Flags&ast.SymbolFlagsClass != 0 {
		membersName = joinExternalName(name, "prototype")
	}
	c.collectTable(namespace, membersName, symbol.Members)
	// Traverse only members written in a declaration, never the type of a value.
	// Type references, return types and function bodies do not introduce paths.
	for _, declaration := range symbol.Declarations {
		switch declaration.Kind {
		case ast.KindVariableDeclaration, ast.KindPropertyDeclaration, ast.KindPropertySignature,
			ast.KindPropertyAssignment, ast.KindTypeAliasDeclaration:
		default:
			continue
		}
		nodes := []*ast.Node{declaration.Type()}
		if declaration.Kind != ast.KindTypeAliasDeclaration {
			nodes = append(nodes, declaration.Initializer())
		}
		for _, node := range nodes {
			if node != nil && (node.Kind == ast.KindTypeLiteral || node.Kind == ast.KindObjectLiteralExpression) {
				if declared := node.Symbol(); declared != nil {
					c.collectTable(namespace, name, declared.Members)
				}
			}
		}
	}
}

// record appends each namespace, symbol ID and name combination once, retaining
// distinct names for the same symbol.
//
// JavaScript example:
//
//	const api = {};
//	export { api as other };
//
// "api" and "other" produce separate records with the same symbol ID; recording
// either name again in the same namespace does not add a duplicate.
func (c *externalSymbolCollector) record(namespace, name string, symbol *ast.Symbol) {
	if symbol == nil || name == "" || strings.HasPrefix(name, ast.InternalSymbolNamePrefix) {
		return
	}
	symbolID := ast.GetSymbolId(symbol)
	key := externalSymbolNameKey{externalSymbolKey{namespace, symbolID}, name}
	if c.collected[key] {
		return
	}
	c.collected[key] = true
	c.semantic.ExternalSymbols = append(c.semantic.ExternalSymbols, ExternalSymbol{
		SymbolId: symbolID, Namespace: []byte(namespace), Name: []byte(name),
	})
}

// joinExternalName joins a declaration name to an optional dotted prefix.
func joinExternalName(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
