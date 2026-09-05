package no_restricted_globals

import (
	_ "embed"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_restricted_globals.schema.json
var schemaJSON []byte

// https://eslint.org/docs/latest/rules/no-restricted-globals

var defaultGlobalObjects = []string{"globalThis", "self", "window"}

// globalEntry records the optional custom message configured for a
// restricted global name.
type globalEntry struct {
	hasMessage bool
	message    string
}

type sourceBindingState uint8

const (
	sourceBindingUnchecked sourceBindingState = iota
	sourceBindingAbsent
	sourceBindingPresent
)

type options struct {
	restrictedGlobals map[string]globalEntry
	checkGlobalObject bool
	globalObjects     map[string]sourceBindingState
}

func parseOptions(optionsList []any, sourceFile *ast.SourceFile) (options, bool) {
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
	mayUseRestrictedGlobal := sourceFile == nil || sourceFile.AsNode().Kind != ast.KindSourceFile
	for _, item := range rawGlobals {
		var name string
		switch v := item.(type) {
		case string:
			if v == "" {
				continue
			}
			name = v
			restrictedGlobals[name] = globalEntry{}
		case map[string]interface{}:
			name, _ = v["name"].(string)
			if name == "" {
				continue
			}
			if msg, ok := v["message"].(string); ok {
				restrictedGlobals[name] = globalEntry{hasMessage: true, message: msg}
			} else {
				restrictedGlobals[name] = globalEntry{}
			}
		}
		if !mayUseRestrictedGlobal && name != "" {
			mayUseRestrictedGlobal = sourceFile.HasIdentifier(name)
		}
	}

	var globalObjects map[string]sourceBindingState
	if checkGlobalObject {
		globalObjects = make(map[string]sourceBindingState, len(defaultGlobalObjects)+len(userGlobalObjects))
		for _, name := range defaultGlobalObjects {
			globalObjects[name] = sourceBindingUnchecked
		}
		for _, name := range userGlobalObjects {
			globalObjects[name] = sourceBindingUnchecked
		}
	}

	return options{
		restrictedGlobals: restrictedGlobals,
		checkGlobalObject: checkGlobalObject,
		globalObjects:     globalObjects,
	}, mayUseRestrictedGlobal
}

var NoRestrictedGlobalsRule = rule.Rule{
	Name:   "no-restricted-globals",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts, mayUseRestrictedGlobal := parseOptions(options, ctx.SourceFile)

		// If no globals are restricted, there's nothing to check.
		if len(opts.restrictedGlobals) == 0 {
			return nil
		}
		if opts.checkGlobalObject {
			for name := range opts.globalObjects {
				if !ctx.Globals.Access(name).IsDeclared() {
					delete(opts.globalObjects, name)
				} else if !mayUseRestrictedGlobal {
					mayUseRestrictedGlobal = ctx.SourceFile.HasIdentifier(name)
				}
			}
		}
		if !mayUseRestrictedGlobal {
			return nil
		}

		report := func(node *ast.Node, name string, entry globalEntry) {
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
				entry, restricted := opts.restrictedGlobals[name]

				// Direct reference to a restricted global identifier. ESLint's own
				// "declared elsewhere" and "through" branches only require that the
				// reference resolves to global scope, never that the name is a
				// recognized global — so a configured global (even one set to
				// `off`) can't suppress this rule, and "not locally shadowed" is
				// the only check needed here.
				if restricted && !isInTypeContext(node) && !utils.IsShadowed(node, name) {
					report(node, name, entry)
				}

				// checkGlobalObject: `window.foo`, `self['foo']`, `globalThis.foo`, ...
				if opts.checkGlobalObject {
					if binding, globalObject := opts.globalObjects[name]; globalObject &&
						isUnshadowedGlobalObject(opts.globalObjects, ctx.SourceFile, node, name, binding) {
						checkGlobalObjectAccess(node, name, opts, report)
					}
				}
			},
		}
	},
}

func isUnshadowedGlobalObject(
	globalObjects map[string]sourceBindingState,
	sourceFile *ast.SourceFile,
	node *ast.Node,
	name string,
	binding sourceBindingState,
) bool {
	if binding == sourceBindingUnchecked {
		if sourceHasValueBinding(sourceFile, name) {
			binding = sourceBindingPresent
		} else {
			binding = sourceBindingAbsent
		}
		globalObjects[name] = binding
	}
	return binding == sourceBindingAbsent || !utils.IsShadowed(node, name)
}

// sourceHasValueBinding reports whether any lexical scope in sourceFile
// declares name in the value namespace. The binder links locals containers in
// source order, so a miss proves that no reference in the file can shadow the
// configured global object. Files without binder metadata conservatively fall
// back to the exact per-reference shadowing check.
func sourceHasValueBinding(sourceFile *ast.SourceFile, name string) bool {
	if sourceFile == nil || !sourceFile.IsBound() {
		return true
	}
	for container := sourceFile.AsNode(); container != nil; {
		data := container.LocalsContainerData()
		if data == nil {
			return true
		}
		if len(data.Locals) != 0 &&
			utils.IsValueSymbolDeclaredInFile(data.Locals[name], sourceFile) {
			return true
		}
		if (container.Kind == ast.KindFunctionExpression || container.Kind == ast.KindClassExpression) &&
			container.Name() != nil && container.Name().Text() == name {
			return true
		}
		container = data.NextContainer
	}
	return sourceHasNamedExpressionBinding(sourceFile, name)
}

func sourceHasNamedExpressionBinding(sourceFile *ast.SourceFile, name string) bool {
	found := false
	var visit func(*ast.Node) bool
	visit = func(node *ast.Node) bool {
		if (node.Kind == ast.KindFunctionExpression || node.Kind == ast.KindClassExpression) &&
			node.Name() != nil && node.Name().Text() == name {
			found = true
			return true
		}
		node.ForEachChild(visit)
		return found
	}
	sourceFile.AsNode().ForEachChild(visit)
	return found
}

// checkGlobalObjectAccess walks up a chain of member accesses rooted at a
// global-object identifier (e.g. `window` in `window.window.foo()`, which
// tsgo represents as nested PropertyAccessExpressions), skipping repeated
// self-access hops, and reports the final accessed property if its name is
// restricted. Mirrors ESLint's `astUtils.getVariableByName` + reference walk
// in the rule's `Program:exit` handler.
func checkGlobalObjectAccess(node *ast.Node, globalObjectName string, opts options, report func(*ast.Node, string, globalEntry)) {
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	for utils.IsSpecificMemberAccess(parent, "", globalObjectName) ||
		(utils.IsHeritageQualifiedName(parent) && parent.AsQualifiedName().Right.Text() == globalObjectName) {
		parent = ast.WalkUpParenthesizedExpressions(parent.Parent)
	}

	propName, propNode, ok := staticMemberProperty(parent)
	if !ok {
		return
	}
	if entry, restricted := opts.restrictedGlobals[propName]; restricted {
		report(propNode, propName, entry)
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
	case ast.KindQualifiedName:
		if utils.IsHeritageQualifiedName(node) {
			name := node.AsQualifiedName().Right
			return name.Text(), name, true
		}
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
	// Member names are a frequent collision with common restricted names such
	// as `name` and `length`, and checking them does not require the broader
	// declaration-name classification below.
	if parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Name() == node {
		return true
	}
	if parent.Kind == ast.KindQualifiedName && parent.AsQualifiedName().Right == node {
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

	// Skip lowercase/hyphenated JSX tag names (`<foo />`, `<foo-bar />`):
	// these name an intrinsic HTML element, not a variable reference, and
	// are never added as a scope reference. A capitalized tag name (`<Foo />`)
	// *is* a reference to a component variable and must not be skipped.
	if isJsxIntrinsicTagName(node, parent) {
		return true
	}

	return false
}

// isJsxIntrinsicTagName reports whether node is the bare-identifier tag name
// of a JSX opening/self-closing/closing element that names an intrinsic
// element (lowercase first letter, or containing a hyphen) rather than a
// component reference.
func isJsxIntrinsicTagName(node *ast.Node, parent *ast.Node) bool {
	var tagName *ast.Node
	switch parent.Kind {
	case ast.KindJsxOpeningElement:
		tagName = parent.AsJsxOpeningElement().TagName
	case ast.KindJsxSelfClosingElement:
		tagName = parent.AsJsxSelfClosingElement().TagName
	case ast.KindJsxClosingElement:
		tagName = parent.AsJsxClosingElement().TagName
	default:
		return false
	}
	return tagName == node && scanner.IsIntrinsicJsxName(node.Text())
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
	case ast.KindTypeReference, ast.KindTypeQuery:
		return true
	case ast.KindQualifiedName:
		return !utils.IsHeritageQualifiedName(parent)
	case ast.KindExpressionWithTypeArguments:
		return !utils.IsClassExtendsHeritageClause(parent)
	}
	return false
}
