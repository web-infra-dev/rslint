package no_restricted_imports

import (
	"fmt"
	"slices"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/dlclark/regexp2"
	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// specifierInfo holds the node and type-only status of an import/export specifier,
// used to report errors on the correct AST location.
type specifierInfo struct {
	node       *ast.Node
	isTypeOnly bool
}

// importNameEntry is a single name→specifiers mapping, preserving insertion order.
type importNameEntry struct {
	name       string
	specifiers []specifierInfo
}

// orderedImportNames preserves the insertion order of import names so that
// errors are reported in source-code order (Go maps have non-deterministic iteration).
type orderedImportNames struct {
	entries []importNameEntry
	index   map[string]int // name → index into entries
}

func newOrderedImportNames() *orderedImportNames {
	return &orderedImportNames{index: make(map[string]int)}
}

func (o *orderedImportNames) add(name string, spec specifierInfo) {
	if idx, ok := o.index[name]; ok {
		o.entries[idx].specifiers = append(o.entries[idx].specifiers, spec)
	} else {
		o.index[name] = len(o.entries)
		o.entries = append(o.entries, importNameEntry{name: name, specifiers: []specifierInfo{spec}})
	}
}

// restrictedPathEntry represents one entry from the "paths" config option.
// A single module path can have multiple entries with different importNames restrictions.
type restrictedPathEntry struct {
	message          string
	importNames      []string
	allowImportNames []string
	allowTypeImports bool
}

// restrictedPatternGroup represents one entry from the "patterns" config option.
// It matches import sources via either gitignore-style globs or regex.
type restrictedPatternGroup struct {
	matcher                *ignoreMatcher  // gitignore-style matcher (from "group")
	regexMatcher           *regexp2.Regexp // regex matcher (from "regex"); uses regexp2 for JS-compatible lookahead/lookbehind
	customMessage          string
	importNames            []string
	importNamePattern      *regexp2.Regexp
	allowImportNames       []string
	allowImportNamePattern *regexp2.Regexp
	allowTypeImports       bool
}

// --- Gitignore-style Pattern Matcher ---
//
// Implements matching compatible with the npm `ignore` package used by ESLint.
// Key behaviors:
//   - Pattern without "/" matches at any depth (e.g., "bar" matches "foo/bar")
//   - Pattern with "/" is anchored to root (e.g., "foo/bar" matches only "foo/bar")
//   - "!" prefix negates a pattern; later patterns override earlier ones
//   - "#" prefix is a comment (skipped)
//   - Case-insensitive by default

type ignoreMatcher struct {
	patterns      []ignorePattern
	caseSensitive bool
}

type ignorePattern struct {
	negated bool
	glob    string
}

func newIgnoreMatcher(patterns []string, caseSensitive bool) *ignoreMatcher {
	m := &ignoreMatcher{caseSensitive: caseSensitive}
	for _, raw := range patterns {
		p := raw

		negated := false
		if strings.HasPrefix(p, "!") {
			negated = true
			p = p[1:]
		}

		if p == "" || strings.HasPrefix(p, "#") {
			continue
		}

		// Escaped # → literal #
		if strings.HasPrefix(p, "\\#") {
			p = "#" + p[2:]
		}

		p = strings.TrimSuffix(p, "/")
		if p == "" {
			continue
		}

		// No "/" → match at any depth (gitignore: unrooted pattern)
		if !strings.Contains(p, "/") {
			p = "**/" + p
		} else if strings.HasPrefix(p, "/") {
			p = p[1:] // Leading "/" → anchored to root
		}

		m.patterns = append(m.patterns, ignorePattern{negated: negated, glob: p})
	}
	return m
}

// ignores returns true if path is matched (and not negated) by any pattern.
func (m *ignoreMatcher) ignores(path string) bool {
	ignored := false
	for _, p := range m.patterns {
		glob := p.glob
		testPath := path
		if !m.caseSensitive {
			glob = strings.ToLower(glob)
			testPath = strings.ToLower(testPath)
		}
		if matchIgnore(glob, testPath) {
			ignored = !p.negated
		}
	}
	return ignored
}

// matchIgnore checks for an exact match, then tries directory-prefix matching:
// if any prefix of path (at a '/' boundary) matches the glob, the full path is
// considered matched (gitignore directory semantics — matched directories include
// all their contents). We skip empty-segment boundaries from "//" so that
// doublestar's `*` (which matches zero chars) doesn't produce false positives.
func matchIgnore(glob, path string) bool {
	if matched, _ := doublestar.Match(glob, path); matched {
		return true
	}
	for i := len(path) - 1; i > 0; i-- {
		if path[i] == '/' && path[i-1] != '/' {
			if matched, _ := doublestar.Match(glob, path[:i]); matched {
				return true
			}
		}
	}
	return false
}

// --- Public API for the typescript-eslint variant ---
//
// The exports below let the @typescript-eslint/no-restricted-imports wrapper
// reuse the rule's restriction logic while overriding source extraction (trim)
// and ImportEqualsDeclaration handling (synthesize a default specifier so that
// `importNames: ['default']` and `allowImportNames` apply to `import x = require(...)`,
// matching upstream typescript-eslint behavior).

// SpecifierInfo describes a single import/export specifier — its AST location
// and whether it is type-only. Exported for variants that need to synthesize
// specifiers (e.g. typescript-eslint's import-equals → default-specifier).
// Use NewSpecifierInfo to construct.
type SpecifierInfo = specifierInfo

// NewSpecifierInfo constructs a SpecifierInfo. Use this instead of struct
// literals — the underlying fields are unexported.
func NewSpecifierInfo(node *ast.Node, isTypeOnly bool) SpecifierInfo {
	return specifierInfo{node: node, isTypeOnly: isTypeOnly}
}

// OrderedImportNames is the insertion-ordered map of name → specifiers used by
// the rule's restriction checks. Use NewOrderedImportNames to construct an
// empty instance and Add to populate it.
type OrderedImportNames = orderedImportNames

// NewOrderedImportNames returns an empty OrderedImportNames suitable for use
// with Engine.Check.
func NewOrderedImportNames() *OrderedImportNames { return newOrderedImportNames() }

// Add appends spec under the given name, preserving insertion order on the
// first occurrence and grouping further specifiers under the existing entry.
func (o *OrderedImportNames) Add(name string, spec SpecifierInfo) {
	o.add(name, spec)
}

// Engine encapsulates parsed restriction options. It is the building block
// shared by NoRestrictedImportsRule (core) and the typescript-eslint variant.
// Construct with NewEngine and call Check per declaration.
type Engine struct {
	grouped  map[string][]restrictedPathEntry
	patterns []restrictedPatternGroup
}

// NewEngine parses options and returns an Engine.
func NewEngine(options any) *Engine {
	g, p := parseOptions(options)
	return &Engine{grouped: g, patterns: p}
}

// IsActive reports whether the engine has any restrictions configured. When
// false, the listener can return immediately.
func (e *Engine) IsActive() bool {
	return len(e.grouped) > 0 || len(e.patterns) > 0
}

// Check applies the rule's path and pattern restrictions to a single
// declaration. Source must already be trimmed/normalized; importNames must
// contain whatever specifiers the variant wants to be eligible for
// importNames / allowImportNames / importNamePattern matching (an empty map is
// the ESLint-base default for ImportEquals).
func (e *Engine) Check(ctx *rule.RuleContext, node *ast.Node, source string, importNames *OrderedImportNames) {
	checkRestrictedPathAndReport(ctx, node, source, importNames, e.grouped)
	for i := range e.patterns {
		if isRestrictedPattern(source, &e.patterns[i]) {
			reportPathForPatterns(ctx, node, &e.patterns[i], importNames, source)
		}
	}
}

// ExtractImportNames returns the importNames map for an ImportDeclaration.
func ExtractImportNames(decl *ast.ImportDeclaration) *OrderedImportNames {
	return extractImportNames(decl)
}

// ExtractExportNames returns the importNames map for an ExportDeclaration.
func ExtractExportNames(decl *ast.ExportDeclaration) *OrderedImportNames {
	return extractExportNames(decl)
}

// IsTypeOnlyDeclaration reports whether the entire import/export declaration
// is type-only (`import type`, `import { type ... }` for ALL specifiers,
// `import type x = require(...)`, or the equivalent export forms).
func IsTypeOnlyDeclaration(node *ast.Node) bool {
	return isTypeOnlyDeclaration(node)
}

// BuildAllowTypeImportSourceFilter returns a predicate reporting whether an
// import source matches any path entry or pattern entry that has
// allowTypeImports=true. The predicate is nil if no such entry exists.
//
// This implements the typescript-eslint short-circuit: when the entire
// declaration is type-only AND the predicate returns true for the source, the
// whole declaration is exempted regardless of any other matching entries.
// Without this short-circuit, conflicting duplicate entries (e.g. two `paths`
// for the same name with allowTypeImports both true and false) would diverge
// from upstream — upstream skips on any "true", rslint core checks per-entry.
func BuildAllowTypeImportSourceFilter(options any) func(source string) bool {
	grouped, patterns := parseOptions(options)
	if len(grouped) == 0 && len(patterns) == 0 {
		return nil
	}

	var allowedNames map[string]struct{}
	for name, entries := range grouped {
		for _, e := range entries {
			if e.allowTypeImports {
				if allowedNames == nil {
					allowedNames = make(map[string]struct{})
				}
				allowedNames[name] = struct{}{}
				break
			}
		}
	}

	allowedPatterns := make([]*restrictedPatternGroup, 0)
	for i := range patterns {
		if patterns[i].allowTypeImports {
			allowedPatterns = append(allowedPatterns, &patterns[i])
		}
	}

	if len(allowedNames) == 0 && len(allowedPatterns) == 0 {
		return nil
	}

	return func(source string) bool {
		if _, ok := allowedNames[source]; ok {
			return true
		}
		for _, p := range allowedPatterns {
			if isRestrictedPattern(source, p) {
				return true
			}
		}
		return false
	}
}

// --- Options Parsing ---
//
// The rule accepts two option formats (matching ESLint):
//   1. Array of strings/objects: ["fs", { name: "foo", importNames: ["bar"] }]
//   2. Single object with paths/patterns: [{ paths: [...], patterns: [...] }]

func parseOptions(options any) (groupedPaths map[string][]restrictedPathEntry, patternGroups []restrictedPatternGroup) {
	groupedPaths = make(map[string][]restrictedPathEntry)

	// Handle the case where the config parser unwraps a single-element array,
	// delivering a map directly instead of [map]. Wrap it back into an array.
	if obj, ok := options.(map[string]interface{}); ok {
		options = []interface{}{obj}
	}

	arr, ok := options.([]interface{})
	if !ok || len(arr) == 0 {
		return groupedPaths, patternGroups
	}

	// Detect format 2: first element is an object with "paths" or "patterns" key
	isPathAndPatternsObject := false
	if firstObj, ok := arr[0].(map[string]interface{}); ok {
		_, hasPaths := firstObj["paths"]
		_, hasPatterns := firstObj["patterns"]
		if hasPaths || hasPatterns {
			isPathAndPatternsObject = true
		}
	}

	var restrictedPathsList []interface{}
	var restrictedPatternsList []interface{}

	if isPathAndPatternsObject {
		obj, _ := arr[0].(map[string]interface{})
		if paths, ok := obj["paths"]; ok {
			if pathsArr, ok := paths.([]interface{}); ok {
				restrictedPathsList = pathsArr
			}
		}
		if patterns, ok := obj["patterns"]; ok {
			if patternsArr, ok := patterns.([]interface{}); ok {
				restrictedPatternsList = patternsArr
			}
		}
	} else {
		restrictedPathsList = arr
	}

	// Parse restricted paths
	for _, item := range restrictedPathsList {
		switch v := item.(type) {
		case string:
			groupedPaths[v] = append(groupedPaths[v], restrictedPathEntry{})
		case map[string]interface{}:
			name, _ := v["name"].(string)
			rp := restrictedPathEntry{}
			if msg, ok := v["message"].(string); ok {
				rp.message = msg
			}
			rp.importNames = utils.ToStringSlice(v["importNames"])
			rp.allowImportNames = utils.ToStringSlice(v["allowImportNames"])
			if b, ok := v["allowTypeImports"].(bool); ok {
				rp.allowTypeImports = b
			}
			groupedPaths[name] = append(groupedPaths[name], rp)
		}
	}

	// Parse restricted patterns
	if len(restrictedPatternsList) > 0 {
		if _, isString := restrictedPatternsList[0].(string); isString {
			// String array → single implicit group
			var group []string
			for _, p := range restrictedPatternsList {
				if s, ok := p.(string); ok {
					group = append(group, s)
				}
			}
			patternGroups = append(patternGroups, restrictedPatternGroup{
				matcher: newIgnoreMatcher(group, false),
			})
		} else {
			// Object array → one group per object
			for _, item := range restrictedPatternsList {
				obj, ok := item.(map[string]interface{})
				if !ok {
					continue
				}
				pg := restrictedPatternGroup{}
				caseSensitive := false
				if cs, ok := obj["caseSensitive"].(bool); ok {
					caseSensitive = cs
				}
				if group := utils.ToStringSlice(obj["group"]); len(group) > 0 {
					pg.matcher = newIgnoreMatcher(group, caseSensitive)
				}
				if regexStr, ok := obj["regex"].(string); ok {
					opts := regexp2.RegexOptions(regexp2.IgnoreCase | regexp2.Unicode)
					if caseSensitive {
						opts = regexp2.Unicode
					}
					if re, err := regexp2.Compile(regexStr, opts); err == nil {
						pg.regexMatcher = re
					}
				}
				if msg, ok := obj["message"].(string); ok {
					pg.customMessage = msg
				}
				pg.importNames = utils.ToStringSlice(obj["importNames"])
				pg.allowImportNames = utils.ToStringSlice(obj["allowImportNames"])
				if pattern, ok := obj["importNamePattern"].(string); ok && pattern != "" {
					if re, err := regexp2.Compile(pattern, regexp2.Unicode); err == nil {
						pg.importNamePattern = re
					}
				}
				if pattern, ok := obj["allowImportNamePattern"].(string); ok && pattern != "" {
					if re, err := regexp2.Compile(pattern, regexp2.Unicode); err == nil {
						pg.allowImportNamePattern = re
					}
				}
				if b, ok := obj["allowTypeImports"].(bool); ok {
					pg.allowTypeImports = b
				}
				patternGroups = append(patternGroups, pg)
			}
		}
	}
	return groupedPaths, patternGroups
}

// --- Rule Definition ---

var NoRestrictedImportsRule = rule.Rule{
	Name: "no-restricted-imports",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		groupedPaths, patternGroups := parseOptions(options)

		if len(groupedPaths) == 0 && len(patternGroups) == 0 {
			return rule.RuleListeners{}
		}

		checkNode := func(node *ast.Node, importSource string, importNames *orderedImportNames) {
			checkRestrictedPathAndReport(&ctx, node, importSource, importNames, groupedPaths)
			for i := range patternGroups {
				if isRestrictedPattern(importSource, &patternGroups[i]) {
					reportPathForPatterns(&ctx, node, &patternGroups[i], importNames, importSource)
				}
			}
		}

		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				importDecl := node.AsImportDeclaration()
				if importDecl.ModuleSpecifier == nil {
					return
				}
				importSource := strings.TrimSpace(utils.GetStaticStringValue(importDecl.ModuleSpecifier))
				if importSource == "" {
					return
				}
				checkNode(node, importSource, extractImportNames(importDecl))
			},
			ast.KindExportDeclaration: func(node *ast.Node) {
				exportDecl := node.AsExportDeclaration()
				if exportDecl.ModuleSpecifier == nil {
					return
				}
				importSource := strings.TrimSpace(utils.GetStaticStringValue(exportDecl.ModuleSpecifier))
				if importSource == "" {
					return
				}
				checkNode(node, importSource, extractExportNames(exportDecl))
			},
			ast.KindImportEqualsDeclaration: func(node *ast.Node) {
				ieDecl := node.AsImportEqualsDeclaration()
				if ieDecl.ModuleReference == nil || ieDecl.ModuleReference.Kind != ast.KindExternalModuleReference {
					return
				}
				extRef := ieDecl.ModuleReference.AsExternalModuleReference()
				if extRef.Expression == nil {
					return
				}
				// Trim to match how ImportDeclaration / ExportDeclaration sources are
				// normalized — and to match typescript-eslint's wrapper, which
				// synthesizes an ImportDeclaration before checking and thus trims.
				// ESLint base does NOT trim require() source, but in practice nobody
				// writes whitespace inside require('...') so the divergence is moot
				// and the consistency is worth more.
				importSource := strings.TrimSpace(utils.GetStaticStringValue(extRef.Expression))
				if importSource == "" {
					return
				}
				checkNode(node, importSource, newOrderedImportNames())
			},
		}
	},
}

// --- Import/Export Name Extraction ---

// extractImportNames builds a map from import name → specifier info for an ImportDeclaration.
// - Default import → name "default"
// - Namespace import (import * as ns) → name "*"
// - Named imports → the source name before "as" (e.g., import { Foo as Bar } → "Foo")
func extractImportNames(importDecl *ast.ImportDeclaration) *orderedImportNames {
	importNames := newOrderedImportNames()
	if importDecl.ImportClause == nil {
		return importNames
	}
	ic := importDecl.ImportClause.AsImportClause()
	if ic == nil {
		return importNames
	}
	isWholeTypeOnly := ic.PhaseModifier == ast.KindTypeKeyword

	// Default import: import foo from 'bar'
	if ic.Name() != nil {
		importNames.add("default", specifierInfo{
			node:       importDecl.ImportClause,
			isTypeOnly: isWholeTypeOnly,
		})
	}

	if ic.NamedBindings == nil {
		return importNames
	}
	switch ic.NamedBindings.Kind {
	case ast.KindNamespaceImport:
		importNames.add("*", specifierInfo{
			node:       ic.NamedBindings,
			isTypeOnly: isWholeTypeOnly,
		})
	case ast.KindNamedImports:
		ni := ic.NamedBindings.AsNamedImports()
		if ni.Elements != nil {
			for _, elem := range ni.Elements.Nodes {
				spec := elem.AsImportSpecifier()
				if spec == nil {
					continue
				}
				name := specifierSourceName(spec.PropertyName, spec.Name())
				importNames.add(name, specifierInfo{
					node:       elem,
					isTypeOnly: spec.IsTypeOnly || isWholeTypeOnly,
				})
			}
		}
	}
	return importNames
}

// extractExportNames builds a map from export source name → specifier info for an ExportDeclaration.
// - export * from 'foo' → name "*"
// - export * as ns from 'foo' → name "*"
// - export { Foo as Bar } from 'foo' → name "Foo" (the source module's export name)
func extractExportNames(exportDecl *ast.ExportDeclaration) *orderedImportNames {
	importNames := newOrderedImportNames()

	if exportDecl.ExportClause == nil {
		// export * from 'foo' — find the * token node for accurate error position
		importNames.add("*", specifierInfo{
			node:       findStarToken(exportDecl),
			isTypeOnly: exportDecl.IsTypeOnly,
		})
		return importNames
	}

	switch exportDecl.ExportClause.Kind {
	case ast.KindNamespaceExport:
		// export * as ns from 'foo' — use * token for error position (matching ESLint)
		importNames.add("*", specifierInfo{
			node:       findStarToken(exportDecl),
			isTypeOnly: exportDecl.IsTypeOnly,
		})
	case ast.KindNamedExports:
		ne := exportDecl.ExportClause.AsNamedExports()
		if ne.Elements != nil {
			for _, elem := range ne.Elements.Nodes {
				spec := elem.AsExportSpecifier()
				if spec == nil {
					continue
				}
				name := specifierSourceName(spec.PropertyName, spec.Name())
				importNames.add(name, specifierInfo{
					node:       elem,
					isTypeOnly: spec.IsTypeOnly || exportDecl.IsTypeOnly,
				})
			}
		}
	}
	return importNames
}

// specifierSourceName returns the "source" name of an import/export specifier.
// For "import { Foo as Bar }", propertyName is Foo (the module export) and nameNode is Bar (the local binding).
// When there is no "as" rename, propertyName is nil and nameNode is used for both.
func specifierSourceName(propertyName *ast.Node, nameNode *ast.Node) string {
	if propertyName != nil {
		if name, ok := utils.GetStaticPropertyName(propertyName); ok {
			return name
		}
	}
	if nameNode != nil {
		if name, ok := utils.GetStaticPropertyName(nameNode); ok {
			return name
		}
	}
	return ""
}

// findStarToken scans the ExportDeclaration to locate the `*` token node.
// For `export * from 'foo'`, the `*` is a token without its own AST node; we
// use the scanner to find it and GetOrCreateToken to materialize a node so that
// ReportNode can highlight just the `*` (matching ESLint's behavior).
func findStarToken(exportDecl *ast.ExportDeclaration) *ast.Node {
	node := exportDecl.AsNode()
	sf := ast.GetSourceFileOfNode(node)
	if sf == nil {
		return nil
	}
	s := scanner.GetScannerForSourceFile(sf, node.Pos())
	for s.TokenStart() < node.End() {
		if s.Token() == ast.KindAsteriskToken {
			return sf.GetOrCreateToken(s.Token(), s.TokenFullStart(), s.TokenEnd(), node, ast.TokenFlagsNone)
		}
		s.Scan()
	}
	return nil
}

// --- Path Restriction Checking ---

func checkRestrictedPathAndReport(
	ctx *rule.RuleContext,
	node *ast.Node,
	importSource string,
	importNames *orderedImportNames,
	groupedPaths map[string][]restrictedPathEntry,
) {
	entries, ok := groupedPaths[importSource]
	if !ok {
		return
	}

	for _, entry := range entries {
		if entry.allowTypeImports && isTypeOnlyDeclaration(node) {
			continue
		}

		// No importNames or allowImportNames → restrict the entire import path
		if len(entry.importNames) == 0 && len(entry.allowImportNames) == 0 {
			messageId := "path"
			msg := fmt.Sprintf("'%s' import is restricted from being used.", importSource)
			if entry.message != "" {
				messageId = "pathWithCustomMessage"
				msg += " " + entry.message
			}
			ctx.ReportNode(node, rule.RuleMessage{Id: messageId, Description: msg})
			continue
		}

		for _, e := range importNames.entries {
			if e.name == "*" {
				reportStarForPath(ctx, node, e.specifiers[0], entry, importSource)
				continue
			}
			if len(entry.importNames) > 0 && slices.Contains(entry.importNames, e.name) {
				reportNameForPath(ctx, node, e.specifiers, entry, e.name, importSource, "importName", "importNameWithCustomMessage",
					fmt.Sprintf("'%s' import from '%s' is restricted.", e.name, importSource))
			}
			if len(entry.allowImportNames) > 0 && !slices.Contains(entry.allowImportNames, e.name) {
				reportNameForPath(ctx, node, e.specifiers, entry, e.name, importSource, "allowedImportName", "allowedImportNameWithCustomMessage",
					fmt.Sprintf("'%s' import from '%s' is restricted because only %s %s allowed.",
						e.name, importSource, formatEnglishList(entry.allowImportNames), isOrAre(entry.allowImportNames)))
			}
		}
	}
}

// reportStarForPath reports an error for a namespace ("*") import against a path restriction.
func reportStarForPath(ctx *rule.RuleContext, node *ast.Node, spec specifierInfo, entry restrictedPathEntry, importSource string) {
	reportNode := spec.node
	if reportNode == nil {
		reportNode = node
	}
	if len(entry.importNames) > 0 {
		messageId := "everything"
		msg := fmt.Sprintf("* import is invalid because %s from '%s' %s restricted.",
			formatEnglishList(entry.importNames), importSource, isOrAre(entry.importNames))
		if entry.message != "" {
			messageId = "everythingWithCustomMessage"
			msg += " " + entry.message
		}
		ctx.ReportNode(reportNode, rule.RuleMessage{Id: messageId, Description: msg})
	} else if len(entry.allowImportNames) > 0 {
		messageId := "everythingWithAllowImportNames"
		msg := fmt.Sprintf("* import is invalid because only %s from '%s' %s allowed.",
			formatEnglishList(entry.allowImportNames), importSource, isOrAre(entry.allowImportNames))
		if entry.message != "" {
			messageId = "everythingWithAllowImportNamesAndCustomMessage"
			msg += " " + entry.message
		}
		ctx.ReportNode(reportNode, rule.RuleMessage{Id: messageId, Description: msg})
	}
}

// reportNameForPath reports errors for each non-type-only specifier with the given messageIds.
func reportNameForPath(ctx *rule.RuleContext, node *ast.Node, specifiers []specifierInfo, entry restrictedPathEntry, name string, importSource string, messageId string, messageIdCustom string, baseMsg string) {
	for _, spec := range specifiers {
		if entry.allowTypeImports && spec.isTypeOnly {
			continue
		}
		id := messageId
		msg := baseMsg
		if entry.message != "" {
			id = messageIdCustom
			msg += " " + entry.message
		}
		reportNode := spec.node
		if reportNode == nil {
			reportNode = node
		}
		ctx.ReportNode(reportNode, rule.RuleMessage{Id: id, Description: msg})
	}
}

// --- Pattern Restriction Checking ---

func isRestrictedPattern(importSource string, group *restrictedPatternGroup) bool {
	if group.regexMatcher != nil {
		matched, _ := group.regexMatcher.MatchString(importSource)
		return matched
	}
	if group.matcher != nil {
		return group.matcher.ignores(importSource)
	}
	return false
}

func reportPathForPatterns(
	ctx *rule.RuleContext,
	node *ast.Node,
	group *restrictedPatternGroup,
	importNames *orderedImportNames,
	importSource string,
) {
	if group.allowTypeImports && isTypeOnlyDeclaration(node) {
		return
	}

	customMessage := group.customMessage

	// No specific import name restrictions → report the whole import
	if len(group.importNames) == 0 && len(group.allowImportNames) == 0 &&
		group.importNamePattern == nil && group.allowImportNamePattern == nil {
		messageId := "patterns"
		msg := fmt.Sprintf("'%s' import is restricted from being used by a pattern.", importSource)
		if customMessage != "" {
			messageId = "patternWithCustomMessage"
			msg += " " + customMessage
		}
		ctx.ReportNode(node, rule.RuleMessage{Id: messageId, Description: msg})
		return
	}

	for _, e := range importNames.entries {
		if e.name == "*" {
			reportStarForPattern(ctx, node, e.specifiers[0], group, importSource)
			continue
		}

		// Check restricted import names (by name list or regex pattern)
		if (len(group.importNames) > 0 && slices.Contains(group.importNames, e.name)) ||
			(group.importNamePattern != nil && regexp2Match(group.importNamePattern, e.name)) {
			reportSpecifiersForPattern(ctx, node, e.specifiers, group, e.name, importSource,
				"patternAndImportName", "patternAndImportNameWithCustomMessage",
				fmt.Sprintf("'%s' import from '%s' is restricted from being used by a pattern.", e.name, importSource))
		}

		if len(group.allowImportNames) > 0 && !slices.Contains(group.allowImportNames, e.name) {
			reportSpecifiersForPattern(ctx, node, e.specifiers, group, e.name, importSource,
				"allowedImportName", "allowedImportNameWithCustomMessage",
				fmt.Sprintf("'%s' import from '%s' is restricted because only %s %s allowed.",
					e.name, importSource, formatEnglishList(group.allowImportNames), isOrAre(group.allowImportNames)))
		} else if group.allowImportNamePattern != nil && !regexp2Match(group.allowImportNamePattern, e.name) {
			reportSpecifiersForPattern(ctx, node, e.specifiers, group, e.name, importSource,
				"allowedImportNamePattern", "allowedImportNamePatternWithCustomMessage",
				fmt.Sprintf("'%s' import from '%s' is restricted because only imports that match the pattern '%s' are allowed from '%s'.",
					e.name, importSource, jsRegexString(group.allowImportNamePattern), importSource))
		}
	}
}

// reportStarForPattern reports an error for a namespace ("*") import against a pattern restriction.
func reportStarForPattern(ctx *rule.RuleContext, node *ast.Node, spec specifierInfo, group *restrictedPatternGroup, importSource string) {
	reportNode := spec.node
	if reportNode == nil {
		reportNode = node
	}
	customMessage := group.customMessage

	if len(group.importNames) > 0 {
		messageId := "patternAndEverything"
		msg := fmt.Sprintf("* import is invalid because %s from '%s' %s restricted from being used by a pattern.",
			formatEnglishList(group.importNames), importSource, isOrAre(group.importNames))
		if customMessage != "" {
			messageId = "patternAndEverythingWithCustomMessage"
			msg += " " + customMessage
		}
		ctx.ReportNode(reportNode, rule.RuleMessage{Id: messageId, Description: msg})
	} else if len(group.allowImportNames) > 0 {
		messageId := "everythingWithAllowImportNames"
		msg := fmt.Sprintf("* import is invalid because only %s from '%s' %s allowed.",
			formatEnglishList(group.allowImportNames), importSource, isOrAre(group.allowImportNames))
		if customMessage != "" {
			messageId = "everythingWithAllowImportNamesAndCustomMessage"
			msg += " " + customMessage
		}
		ctx.ReportNode(reportNode, rule.RuleMessage{Id: messageId, Description: msg})
	} else if group.allowImportNamePattern != nil {
		messageId := "everythingWithAllowedImportNamePattern"
		msg := fmt.Sprintf("* import is invalid because only imports that match the pattern '%s' from '%s' are allowed.",
			jsRegexString(group.allowImportNamePattern), importSource)
		if customMessage != "" {
			messageId = "everythingWithAllowedImportNamePatternWithCustomMessage"
			msg += " " + customMessage
		}
		ctx.ReportNode(reportNode, rule.RuleMessage{Id: messageId, Description: msg})
	} else if group.importNamePattern != nil {
		messageId := "patternAndEverythingWithRegexImportName"
		msg := fmt.Sprintf("* import is invalid because import name matching '%s' pattern from '%s' is restricted from being used.",
			jsRegexString(group.importNamePattern), importSource)
		if customMessage != "" {
			messageId = "patternAndEverythingWithRegexImportNameAndCustomMessage"
			msg += " " + customMessage
		}
		ctx.ReportNode(reportNode, rule.RuleMessage{Id: messageId, Description: msg})
	}
}

// reportSpecifiersForPattern reports errors for each non-type-only specifier with the given messageIds.
func reportSpecifiersForPattern(ctx *rule.RuleContext, node *ast.Node, specifiers []specifierInfo, group *restrictedPatternGroup, name string, importSource string, messageId string, messageIdCustom string, baseMsg string) {
	for _, spec := range specifiers {
		if group.allowTypeImports && spec.isTypeOnly {
			continue
		}
		id := messageId
		msg := baseMsg
		if group.customMessage != "" {
			id = messageIdCustom
			msg += " " + group.customMessage
		}
		reportNode := spec.node
		if reportNode == nil {
			reportNode = node
		}
		ctx.ReportNode(reportNode, rule.RuleMessage{Id: id, Description: msg})
	}
}

// --- Type-Only Declaration Checking ---

// isTypeOnlyDeclaration checks whether an entire import/export declaration is type-only.
// This is used for the allowTypeImports option at the declaration level.
// For import declarations: true if "import type ..." or all specifiers have individual "type" keyword.
// For export declarations: true if "export type ..." or all specifiers have individual "type" keyword.
// For import equals: true if "import type x = require(...)".
func isTypeOnlyDeclaration(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindImportDeclaration:
		importDecl := node.AsImportDeclaration()
		if importDecl.ImportClause == nil {
			return false
		}
		ic := importDecl.ImportClause.AsImportClause()
		if ic == nil {
			return false
		}
		// "import type ..." at the clause level
		if ic.PhaseModifier == ast.KindTypeKeyword {
			return true
		}
		// All individual specifiers are "type" (e.g., import { type A, type B })
		return allImportSpecifiersTypeOnly(ic)
	case ast.KindExportDeclaration:
		exportDecl := node.AsExportDeclaration()
		if exportDecl.IsTypeOnly {
			return true
		}
		return allExportSpecifiersTypeOnly(exportDecl)
	case ast.KindImportEqualsDeclaration:
		return node.AsImportEqualsDeclaration().IsTypeOnly
	}
	return false
}

func allImportSpecifiersTypeOnly(ic *ast.ImportClause) bool {
	// Default imports cannot be individually type-only
	if ic.Name() != nil {
		return false
	}
	if ic.NamedBindings == nil || ic.NamedBindings.Kind != ast.KindNamedImports {
		return false
	}
	ni := ic.NamedBindings.AsNamedImports()
	if ni.Elements == nil || len(ni.Elements.Nodes) == 0 {
		return false
	}
	for _, elem := range ni.Elements.Nodes {
		if !elem.IsTypeOnly() {
			return false
		}
	}
	return true
}

func allExportSpecifiersTypeOnly(exportDecl *ast.ExportDeclaration) bool {
	if exportDecl.ExportClause == nil || exportDecl.ExportClause.Kind != ast.KindNamedExports {
		return false
	}
	ne := exportDecl.ExportClause.AsNamedExports()
	if ne.Elements == nil || len(ne.Elements.Nodes) == 0 {
		return false
	}
	for _, elem := range ne.Elements.Nodes {
		if !elem.IsTypeOnly() {
			return false
		}
	}
	return true
}

// --- Utility Functions ---

// formatEnglishList formats names as a single-quoted English list using Intl.ListFormat("en-US") style:
// ["a"] → "'a'", ["a","b"] → "'a' and 'b'", ["a","b","c"] → "'a', 'b', and 'c'"
func formatEnglishList(names []string) string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = "'" + name + "'"
	}
	switch len(quoted) {
	case 0:
		return ""
	case 1:
		return quoted[0]
	case 2:
		return quoted[0] + " and " + quoted[1]
	default:
		return strings.Join(quoted[:len(quoted)-1], ", ") + ", and " + quoted[len(quoted)-1]
	}
}

// jsRegexString formats a Go regexp pattern in JavaScript RegExp.toString() notation: /pattern/u.
// ESLint creates these patterns with `new RegExp(pattern, "u")` and interpolates them via toString().
// jsRegexString formats a regex pattern in JavaScript RegExp.toString() notation: /pattern/u.
func jsRegexString(re *regexp2.Regexp) string {
	return "/" + re.String() + "/u"
}

// regexp2Match wraps regexp2.MatchString, discarding the error (timeout).
func regexp2Match(re *regexp2.Regexp, s string) bool {
	matched, _ := re.MatchString(s)
	return matched
}

func isOrAre(names []string) string {
	if len(names) == 1 {
		return "is"
	}
	return "are"
}
