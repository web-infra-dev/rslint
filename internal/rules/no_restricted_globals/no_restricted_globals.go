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

func parseOptions(optionsList []any) options {
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
			if v == "" {
				continue
			}
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
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

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

				// Direct reference to a restricted global identifier. ESLint's own
				// "declared elsewhere" and "through" branches only require that the
				// reference resolves to global scope, never that the name is a
				// recognized global — so a configured global (even one set to
				// `off`) can't suppress this rule, and "not locally shadowed" is
				// the only check needed here.
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

func shouldSkip(node *ast.Node) bool {
	return node == nil || node.Parent == nil || utils.IsNonReferenceIdentifier(node)
}

// isInTypeContext excludes type-only references while retaining class
// `extends`, whose superclass expression is evaluated at runtime.
func isInTypeContext(node *ast.Node) bool {
	return utils.IsIdentifierInTypeReference(node)
}
