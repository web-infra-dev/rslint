package no_restricted_globals

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/no-restricted-globals

var defaultGlobalObjects = []string{"globalThis", "self", "window"}

// globalEntry records the optional custom message configured for a
// restricted global name.
type globalEntry struct {
	hasMessage bool
	message    string
}

type options struct {
	restrictedGlobals map[string]globalEntry
	checkGlobalObject bool
	globalObjects     map[string]bool
}

// normalizeOptionsList mirrors ESLint's `context.options`: rslint's config
// loader unwraps a single non-array option to a bare value (string or map),
// so this restores the array shape ESLint's rule body assumes.
func normalizeOptionsList(opts any) []interface{} {
	if opts == nil {
		return nil
	}
	if list, ok := opts.([]interface{}); ok {
		return list
	}
	return []interface{}{opts}
}

func parseOptions(opts any) options {
	optionsList := normalizeOptionsList(opts)

	isGlobalsObject := false
	var globalsObjectMap map[string]interface{}
	if len(optionsList) > 0 {
		if m, ok := optionsList[0].(map[string]interface{}); ok {
			if _, has := m["globals"]; has {
				isGlobalsObject = true
				globalsObjectMap = m
			}
		}
	}

	var rawGlobals []interface{}
	checkGlobalObject := false
	var userGlobalObjects []string
	if isGlobalsObject {
		if g, ok := globalsObjectMap["globals"].([]interface{}); ok {
			rawGlobals = g
		}
		if b, ok := globalsObjectMap["checkGlobalObject"].(bool); ok {
			checkGlobalObject = b
		}
		if list, ok := globalsObjectMap["globalObjects"].([]interface{}); ok {
			for _, item := range list {
				if s, ok := item.(string); ok {
					userGlobalObjects = append(userGlobalObjects, s)
				}
			}
		}
	} else {
		rawGlobals = optionsList
	}

	restrictedGlobals := make(map[string]globalEntry, len(rawGlobals))
	for _, item := range rawGlobals {
		switch v := item.(type) {
		case string:
			restrictedGlobals[v] = globalEntry{}
		case map[string]interface{}:
			name, _ := v["name"].(string)
			if name == "" {
				continue
			}
			if msg, ok := v["message"].(string); ok {
				restrictedGlobals[name] = globalEntry{hasMessage: true, message: msg}
			} else {
				restrictedGlobals[name] = globalEntry{}
			}
		}
	}

	globalObjects := make(map[string]bool, len(defaultGlobalObjects)+len(userGlobalObjects))
	for _, name := range defaultGlobalObjects {
		globalObjects[name] = true
	}
	for _, name := range userGlobalObjects {
		globalObjects[name] = true
	}

	return options{
		restrictedGlobals: restrictedGlobals,
		checkGlobalObject: checkGlobalObject,
		globalObjects:     globalObjects,
	}
}

var NoRestrictedGlobalsRule = rule.Rule{
	Name: "no-restricted-globals",
	Run: func(ctx rule.RuleContext, rawOptions any) rule.RuleListeners {
		opts := parseOptions(rawOptions)

		// If no globals are restricted, there's nothing to check.
		if len(opts.restrictedGlobals) == 0 {
			return rule.RuleListeners{}
		}

		report := func(node *ast.Node, name string) {
			entry := opts.restrictedGlobals[name]
			if entry.hasMessage {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "customMessage",
					Description: fmt.Sprintf("Unexpected use of '%s'. %s", name, entry.message),
				})
			} else {
				ctx.ReportNode(node, rule.RuleMessage{
					Id:          "defaultMessage",
					Description: fmt.Sprintf("Unexpected use of '%s'.", name),
				})
			}
		}

		return rule.RuleListeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if shouldSkip(node) {
					return
				}
				name := node.Text()

				// Direct reference to a restricted global identifier.
				//
				// NOTE: Unlike ESLint, rslint does not model `languageOptions.globals` /
				// environment configuration, so we can't distinguish "declared by an
				// environment/global comment" from "genuinely undeclared" — any
				// identifier that isn't locally shadowed is treated as a reference to
				// the global. This is a superset of ESLint's behavior for this branch:
				// ESLint's own logic doesn't require the name to be a *recognized*
				// global here either (both its "declared elsewhere" and "through"
				// branches only require "resolves to global scope"), so this matches.
				if _, restricted := opts.restrictedGlobals[name]; restricted &&
					!isInTypeContext(node) && !utils.IsShadowed(node, name) {
					report(node, name)
				}

				// checkGlobalObject: `window.foo`, `self['foo']`, `globalThis.foo`, ...
				if opts.checkGlobalObject && opts.globalObjects[name] && !utils.IsShadowed(node, name) {
					checkGlobalObjectAccess(node, name, opts, report)
				}
			},
		}
	},
}

// checkGlobalObjectAccess walks up a chain of member accesses rooted at a
// global-object identifier (e.g. `window` in `window.window.foo()`, which
// tsgo represents as nested PropertyAccessExpressions), skipping repeated
// self-access hops, and reports the final accessed property if its name is
// restricted. Mirrors ESLint's `astUtils.getVariableByName` + reference walk
// in the rule's `Program:exit` handler.
func checkGlobalObjectAccess(node *ast.Node, globalObjectName string, opts options, report func(*ast.Node, string)) {
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	for utils.IsSpecificMemberAccess(parent, "", globalObjectName) {
		parent = ast.WalkUpParenthesizedExpressions(parent.Parent)
	}

	propName, propNode, ok := staticMemberProperty(parent)
	if !ok {
		return
	}
	if _, restricted := opts.restrictedGlobals[propName]; restricted {
		report(propNode, propName)
	}
}

// staticMemberProperty extracts the statically-known property name and its
// reporting node from a member access (`obj.prop` or `obj['prop']`),
// unwrapping parentheses first.
func staticMemberProperty(node *ast.Node) (string, *ast.Node, bool) {
	node = ast.SkipParentheses(node)
	if node == nil {
		return "", nil, false
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		name := node.AsPropertyAccessExpression().Name()
		propName, ok := utils.GetStaticPropertyName(name)
		if !ok {
			return "", nil, false
		}
		return propName, name, true
	case ast.KindElementAccessExpression:
		arg := ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
		val, ok := utils.GetStaticExpressionValue(arg)
		if !ok {
			return "", nil, false
		}
		return val, arg, true
	}
	return "", nil, false
}

// shouldSkip reports whether node occupies a "name" position (declaration
// name, member/property name, label, or the original name in an aliased
// import/re-export specifier) rather than a genuine value-reference
// position that ESLint's scope analysis would surface as a variable
// reference.
func shouldSkip(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return true
	}

	// Skip declaration names (var/let/const/function/class/enum/import/
	// parameter names, and member names: methods, properties, accessors,
	// signatures, object-literal keys). ShorthandPropertyAssignment names
	// (`{foo}`) are declaration-shaped but also read the outer binding, so
	// they must NOT be skipped.
	if ast.IsDeclarationName(node) && parent.Kind != ast.KindShorthandPropertyAssignment {
		return true
	}

	// Skip property names in member access (obj.prop -> skip "prop").
	if parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Name() == node {
		return true
	}

	// Skip the destructuring source key ({key: newName} -> skip "key"); the
	// binding target "newName" is a declaration, handled above.
	if parent.Kind == ast.KindBindingElement && parent.PropertyName() == node {
		return true
	}

	// Skip the original name in aliased imports (import { Original as Alias }
	// -> skip "Original"); it names a module export, not a scope reference.
	if parent.Kind == ast.KindImportSpecifier && parent.PropertyName() == node {
		return true
	}

	// Skip the original name in re-export aliases (export { Original as
	// Alias } from 'module'). Without `from`, `export { X as Y }` reads the
	// local X, so only skip when it's an actual re-export.
	if parent.Kind == ast.KindExportSpecifier && parent.PropertyName() == node &&
		utils.IsReExportSpecifier(parent) {
		return true
	}

	// Skip label identifiers; labels are a separate namespace from variables.
	if parent.Kind == ast.KindLabeledStatement ||
		parent.Kind == ast.KindBreakStatement ||
		parent.Kind == ast.KindContinueStatement {
		return true
	}

	return false
}

// isInTypeContext mirrors ESLint's TYPE_NODES check: a restricted-global
// reference is not reported when it names a type instead of a value, i.e.
// its immediate parent is a type reference, a type query (`typeof X` in a
// type position), a qualified type name (`NS.Test`), or a heritage clause
// entry (`implements X` / `interface Y extends X`) — but NOT a class
// `extends` clause, whose superclass expression is evaluated at runtime.
func isInTypeContext(node *ast.Node) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindTypeReference, ast.KindTypeQuery, ast.KindQualifiedName:
		return true
	case ast.KindExpressionWithTypeArguments:
		return !utils.IsClassExtendsHeritageClause(parent)
	}
	return false
}
