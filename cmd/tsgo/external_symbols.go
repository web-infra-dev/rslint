package main

import (
	"sort"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/compiler"
	"github.com/microsoft/typescript-go/shim/module"
	"github.com/microsoft/typescript-go/shim/tspath"
)

const globalNamespace = "global"

type externalModuleRoot struct {
	namespace string
	prefix    string
	fileName  string
}

type externalSymbolKey struct {
	namespace string
	symbolID  ast.SymbolId
}

type externalSymbolCollector struct {
	tc        *checker.Checker
	semantic  *Semantic
	expanded  map[externalSymbolKey]string
	collected map[ast.SymbolId]int
}

func collectExternalSymbols(program *compiler.Program, tc *checker.Checker, semantic *Semantic) {
	if program == nil || tc == nil || semantic == nil {
		return
	}

	collector := externalSymbolCollector{
		tc:        tc,
		semantic:  semantic,
		expanded:  make(map[externalSymbolKey]string),
		collected: make(map[ast.SymbolId]int),
	}
	collector.collectGlobals()
	collector.collectDependencies(program)
}

func (c *externalSymbolCollector) collectGlobals() {
	globalThis := c.tc.GetGlobalSymbol("globalThis", ast.SymbolFlagsModule, nil)
	if globalThis == nil {
		return
	}
	c.record(globalNamespace, "globalThis", globalThis)
	globalThisID := ast.GetSymbolId(globalThis)
	c.expanded[externalSymbolKey{namespace: globalNamespace, symbolID: globalThisID}] = "globalThis"

	ty := c.tc.GetTypeOfSymbol(globalThis)
	if ty == nil {
		return
	}
	properties := c.tc.GetPropertiesOfType(ty)
	sort.Slice(properties, func(i int, j int) bool {
		return properties[i].Name < properties[j].Name
	})
	for _, property := range properties {
		c.collect(globalNamespace, property.Name, property)
	}
}

func (c *externalSymbolCollector) collectDependencies(program *compiler.Program) {
	roots := make(map[externalModuleRoot]struct{})
	for _, resolutions := range program.GetResolvedModules() {
		for key, resolution := range resolutions {
			if resolution == nil || !resolution.IsResolved() || !resolution.IsExternalLibraryImport {
				continue
			}
			namespace, prefix, ok := externalModuleName(key.Name)
			if !ok {
				continue
			}
			roots[externalModuleRoot{
				namespace: namespace,
				prefix:    prefix,
				fileName:  resolution.ResolvedFileName,
			}] = struct{}{}
		}
	}

	sortedRoots := make([]externalModuleRoot, 0, len(roots))
	for root := range roots {
		sortedRoots = append(sortedRoots, root)
	}
	sort.Slice(sortedRoots, func(i int, j int) bool {
		left := sortedRoots[i]
		right := sortedRoots[j]
		if left.namespace != right.namespace {
			return left.namespace < right.namespace
		}
		if left.prefix != right.prefix {
			return left.prefix < right.prefix
		}
		return left.fileName < right.fileName
	})

	for _, root := range sortedRoots {
		file := program.GetSourceFileForResolvedModule(root.fileName)
		if file == nil || file.Symbol == nil {
			continue
		}
		exports := c.tc.GetExportsAndPropertiesOfModule(file.Symbol)
		sort.Slice(exports, func(i int, j int) bool {
			return exports[i].Name < exports[j].Name
		})
		for _, symbol := range exports {
			name := joinExternalName(root.prefix, symbol.Name)
			c.collect(root.namespace, name, symbol)
		}
	}
}

func (c *externalSymbolCollector) collect(namespace string, name string, symbol *ast.Symbol) {
	if symbol == nil || name == "" || strings.HasPrefix(name, ast.InternalSymbolNamePrefix) {
		return
	}
	if symbol.Flags&ast.SymbolFlagsAlias != 0 {
		if target, ok := c.tc.ResolveAlias(symbol); ok {
			symbol = target
		}
	}
	if symbol == nil || symbol.Flags&ast.SymbolFlagsValue == 0 {
		return
	}

	c.record(namespace, name, symbol)

	symbolID := ast.GetSymbolId(symbol)
	expandedKey := externalSymbolKey{namespace: namespace, symbolID: symbolID}
	if expandedName, exists := c.expanded[expandedKey]; exists && !preferShorterName(name, expandedName) {
		return
	}
	c.expanded[expandedKey] = name

	ty := c.tc.GetTypeOfSymbol(symbol)
	if ty == nil {
		return
	}
	properties := c.tc.GetPropertiesOfType(ty)
	sort.Slice(properties, func(i int, j int) bool {
		return properties[i].Name < properties[j].Name
	})
	for _, property := range properties {
		c.collect(namespace, joinExternalName(name, property.Name), property)
	}
}

func (c *externalSymbolCollector) record(namespace string, name string, symbol *ast.Symbol) {
	symbolID := ast.GetSymbolId(symbol)
	external := ExternalSymbol{
		SymbolId:  symbolID,
		Namespace: []byte(namespace),
		Name:      []byte(name),
	}
	if index, exists := c.collected[symbolID]; exists {
		if preferShorterSymbol(external, c.semantic.ExternalSymbols[index]) {
			c.semantic.ExternalSymbols[index] = external
		}
	} else {
		c.collected[symbolID] = len(c.semantic.ExternalSymbols)
		c.semantic.ExternalSymbols = append(c.semantic.ExternalSymbols, external)
	}
}

func externalModuleName(moduleName string) (namespace string, prefix string, ok bool) {
	if moduleName == "" || tspath.IsExternalModuleNameRelative(moduleName) || strings.HasPrefix(moduleName, "#") {
		return "", "", false
	}
	if strings.HasPrefix(moduleName, "node:") {
		return moduleName, "", true
	}

	namespace, prefix = module.ParsePackageName(moduleName)
	if namespace == "" {
		return "", "", false
	}
	prefix = strings.ReplaceAll(prefix, "/", ".")
	return namespace, prefix, true
}

func joinExternalName(prefix string, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}

func preferShorterSymbol(candidate ExternalSymbol, current ExternalSymbol) bool {
	candidateNamespace := string(candidate.Namespace)
	currentNamespace := string(current.Namespace)
	if candidateNamespace != currentNamespace {
		if candidateNamespace == globalNamespace {
			return true
		}
		if currentNamespace == globalNamespace {
			return false
		}
		return candidateNamespace < currentNamespace
	}

	candidateName := string(candidate.Name)
	currentName := string(current.Name)
	return preferShorterName(candidateName, currentName)
}

func preferShorterName(candidateName string, currentName string) bool {
	candidateDepth := strings.Count(candidateName, ".")
	currentDepth := strings.Count(currentName, ".")
	if candidateDepth != currentDepth {
		return candidateDepth < currentDepth
	}
	if len(candidateName) != len(currentName) {
		return len(candidateName) < len(currentName)
	}
	return candidateName < currentName
}
