package no_underscore_dangle

import (
	_ "embed"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed no_underscore_dangle.schema.json
var schemaJSON []byte

type noUnderscoreDangleOptions struct {
	allow                      []string
	allowAfterSuper            bool
	allowAfterThis             bool
	allowAfterThisConstructor  bool
	allowFunctionParams        bool
	allowInArrayDestructuring  bool
	allowInObjectDestructuring bool
	enforceInClassFields       bool
	enforceInMethodNames       bool
}

func parseOptions(options []any) noUnderscoreDangleOptions {
	opts := noUnderscoreDangleOptions{
		allowFunctionParams:        true,
		allowInArrayDestructuring:  true,
		allowInObjectDestructuring: true,
	}
	if len(options) == 0 {
		return opts
	}

	optsMap, _ := options[0].(map[string]any)
	if allow, ok := optsMap["allow"].([]any); ok {
		for _, entry := range allow {
			if name, ok := entry.(string); ok {
				opts.allow = append(opts.allow, name)
			}
		}
	}
	if value, ok := optsMap["allowAfterSuper"].(bool); ok {
		opts.allowAfterSuper = value
	}
	if value, ok := optsMap["allowAfterThis"].(bool); ok {
		opts.allowAfterThis = value
	}
	if value, ok := optsMap["allowAfterThisConstructor"].(bool); ok {
		opts.allowAfterThisConstructor = value
	}
	if value, ok := optsMap["allowFunctionParams"].(bool); ok {
		opts.allowFunctionParams = value
	}
	if value, ok := optsMap["allowInArrayDestructuring"].(bool); ok {
		opts.allowInArrayDestructuring = value
	}
	if value, ok := optsMap["allowInObjectDestructuring"].(bool); ok {
		opts.allowInObjectDestructuring = value
	}
	if value, ok := optsMap["enforceInClassFields"].(bool); ok {
		opts.enforceInClassFields = value
	}
	if value, ok := optsMap["enforceInMethodNames"].(bool); ok {
		opts.enforceInMethodNames = value
	}

	return opts
}

// https://eslint.org/docs/latest/rules/no-underscore-dangle
var NoUnderscoreDangleRule = rule.Rule{
	Name:   "no-underscore-dangle",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		opts := parseOptions(options)

		checkFunctionParameters := func(node *ast.Node) {
			checkParameters(ctx, node, opts)
		}

		return rule.RuleListeners{
			ast.KindFunctionDeclaration: func(node *ast.Node) {
				checkFunctionDeclarationName(ctx, node, opts)
				checkParameters(ctx, node, opts)
			},
			ast.KindFunctionExpression: checkFunctionParameters,
			ast.KindArrowFunction:      checkFunctionParameters,
			ast.KindConstructor:        checkFunctionParameters,
			ast.KindMethodDeclaration: func(node *ast.Node) {
				checkMemberName(ctx, node, opts.enforceInMethodNames, opts)
				checkParameters(ctx, node, opts)
			},
			ast.KindGetAccessor: func(node *ast.Node) {
				checkAccessorName(ctx, node, opts)
				checkParameters(ctx, node, opts)
			},
			ast.KindSetAccessor: func(node *ast.Node) {
				checkAccessorName(ctx, node, opts)
				checkParameters(ctx, node, opts)
			},
			ast.KindPropertyDeclaration: func(node *ast.Node) {
				checkClassField(ctx, node, opts)
			},
			ast.KindVariableDeclaration: func(node *ast.Node) {
				checkVariableDeclaration(ctx, node, opts)
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				checkMemberAccess(ctx, node, opts)
			},
			ast.KindElementAccessExpression: func(node *ast.Node) {
				checkMemberAccess(ctx, node, opts)
			},
		}
	},
}

// hasDanglingUnderscore mirrors upstream's helper of the same name: the bare
// `_` name is always exempt, anything else with a leading or trailing `_` is
// dangling.
func hasDanglingUnderscore(identifier string) bool {
	return identifier != "_" &&
		(strings.HasPrefix(identifier, "_") || strings.HasSuffix(identifier, "_"))
}

func (opts noUnderscoreDangleOptions) isAllowed(identifier string) bool {
	for _, allowed := range opts.allow {
		if allowed == identifier {
			return true
		}
	}
	return false
}

// reports the diagnostic when `identifier` dangles and is not allow-listed.
// `display` is what the message shows, which differs from `identifier` for
// private names (`#foo` is reported, `foo` is matched against `allow`).
func reportIfDangling(ctx rule.RuleContext, reportRange core.TextRange, identifier, display string, opts noUnderscoreDangleOptions) {
	if !hasDanglingUnderscore(identifier) || opts.isAllowed(identifier) {
		return
	}
	ctx.ReportRange(reportRange, rule.RuleMessage{
		Id:          "unexpectedUnderscore",
		Description: "Unexpected dangling '_' in '" + display + "'.",
		Data:        map[string]string{"identifier": display},
	})
}

// estreeName returns the name an ESTree `Identifier` / `PrivateIdentifier`
// key or property would carry, and whether it came from a private name. Every
// other node kind (string / numeric literal keys, computed non-identifier
// expressions, …) has no `name` property at all, which is how upstream's
// `typeof identifier !== "undefined"` guards skip them.
func estreeName(node *ast.Node) (identifier string, private bool, ok bool) {
	if node == nil {
		return "", false, false
	}
	switch node.Kind {
	case ast.KindIdentifier:
		return node.AsIdentifier().Text, false, true
	case ast.KindPrivateIdentifier:
		// tsgo keeps the `#` in the token text; ESTree's PrivateIdentifier
		// `name` drops it. Only the member-name messages re-add it.
		return strings.TrimPrefix(node.AsPrivateIdentifier().Text, "#"), true, true
	}
	return "", false, false
}

// memberNameNode unwraps a member name the way ESTree exposes it: a computed
// key holds the inner expression directly, and parentheses are invisible.
func memberNameNode(node *ast.Node) *ast.Node {
	nameNode := node.Name()
	if nameNode != nil && nameNode.Kind == ast.KindComputedPropertyName {
		return ast.SkipParentheses(nameNode.AsComputedPropertyName().Expression)
	}
	return nameNode
}

// checkFunctionDeclarationName ports checkForDanglingUnderscoreInFunction's
// `node.type === "FunctionDeclaration" && node.id` arm. A body-less function
// declaration (an overload signature or a `declare function`) parses as
// TSDeclareFunction upstream, a kind the rule never listens on, so it is
// skipped here too.
func checkFunctionDeclarationName(ctx rule.RuleContext, node *ast.Node, opts noUnderscoreDangleOptions) {
	if node.Body() == nil {
		return
	}
	nameNode := node.Name()
	if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
		return
	}
	identifier := nameNode.AsIdentifier().Text
	reportIfDangling(ctx, functionDeclarationRange(ctx.SourceFile, node), identifier, identifier, opts)
}

// functionDeclarationRange is the span ESTree gives a FunctionDeclaration:
// `export` / `export default` sit on the enclosing export declaration there,
// so the reported range starts at the `async` keyword or the `function`
// keyword, while tsgo keeps those modifiers on the declaration itself.
func functionDeclarationRange(sourceFile *ast.SourceFile, node *ast.Node) core.TextRange {
	nodeRange := utils.TrimNodeTextRange(sourceFile, node)
	pos := nodeRange.Pos()
	if mods := node.Modifiers(); mods != nil {
		for _, modifier := range mods.Nodes {
			switch modifier.Kind {
			case ast.KindExportKeyword, ast.KindDefaultKeyword, ast.KindDeclareKeyword:
				pos = scanner.SkipTrivia(sourceFile.Text(), modifier.End())
			}
		}
	}
	return core.NewTextRange(pos, nodeRange.End())
}

// checkParameters ports checkForDanglingUnderscoreInFunctionParameters. Only
// function forms that ESTree models as FunctionDeclaration /
// FunctionExpression / ArrowFunctionExpression reach here, which is why
// body-less signatures — overloads and `abstract` members — bail out: upstream
// gives those a TSEmptyBodyFunctionExpression / TSDeclareFunction value the
// rule has no listener for.
func checkParameters(ctx rule.RuleContext, node *ast.Node, opts noUnderscoreDangleOptions) {
	if opts.allowFunctionParams || node.Body() == nil {
		return
	}

	for _, param := range node.Parameters() {
		// A parameter property (`constructor(private _x) {}`) is a
		// TSParameterProperty upstream, not an Identifier, so it never
		// matches the Identifier arm below.
		if ast.HasSyntacticModifier(param, ast.ModifierFlagsParameterPropertyModifier) {
			continue
		}
		nameNode := param.Name()
		if nameNode == nil || nameNode.Kind != ast.KindIdentifier {
			continue
		}
		identifier := nameNode.AsIdentifier().Text
		reportIfDangling(ctx, parameterRange(ctx.SourceFile, param), identifier, identifier, opts)
	}
}

// parameterRange is the span ESTree gives a parameter. Decorators are part of
// the parameter node in tsgo; upstream keeps them out of a plain `_x` /
// `_x = 1` parameter but inside a rest parameter, so `@dec ..._x` reports from
// the decorator while `@dec _x` reports from the name.
func parameterRange(sourceFile *ast.SourceFile, param *ast.Node) core.TextRange {
	nodeRange := utils.TrimNodeTextRange(sourceFile, param)
	if param.AsParameterDeclaration().DotDotDotToken != nil || !ast.HasDecorators(param) {
		return nodeRange
	}

	pos := nodeRange.Pos()
	if mods := param.Modifiers(); mods != nil {
		for _, modifier := range mods.Nodes {
			if modifier.Kind == ast.KindDecorator {
				pos = scanner.SkipTrivia(sourceFile.Text(), modifier.End())
			}
		}
	}
	return core.NewTextRange(pos, nodeRange.End())
}

// checkVariableDeclaration ports
// checkForDanglingUnderscoreInVariableExpression: every name the declarator
// binds is checked, and each report lands on the declarator itself. A catch
// clause parameter is a VariableDeclaration in tsgo but a CatchClause `param`
// upstream, so it is skipped.
func checkVariableDeclaration(ctx rule.RuleContext, node *ast.Node, opts noUnderscoreDangleOptions) {
	if node.Parent != nil && node.Parent.Kind == ast.KindCatchClause {
		return
	}

	reportRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
	// Upstream iterates the declarator's *variables*, so a name bound more than
	// once by one declarator (`var {a: _x, b: _x} = o`) is checked once, using
	// the first binding.
	seen := make(map[string]struct{})
	utils.CollectBindingNames(node.Name(), func(ident *ast.Node, identifier string) {
		if _, duplicate := seen[identifier]; duplicate {
			return
		}
		seen[identifier] = struct{}{}
		if !hasDanglingUnderscore(identifier) || opts.isAllowed(identifier) {
			return
		}
		// Upstream walks up from the bound name to the nearest
		// VariableDeclarator / ArrayPattern / ObjectPattern, so a name nested
		// under a default value or a rest element belongs to the pattern that
		// encloses it.
		switch nearestBindingContainer(ident) {
		case ast.KindArrayBindingPattern:
			if opts.allowInArrayDestructuring {
				return
			}
		case ast.KindObjectBindingPattern:
			if opts.allowInObjectDestructuring {
				return
			}
		}
		ctx.ReportRange(reportRange, rule.RuleMessage{
			Id:          "unexpectedUnderscore",
			Description: "Unexpected dangling '_' in '" + identifier + "'.",
			Data:        map[string]string{"identifier": identifier},
		})
	})
}

func nearestBindingContainer(ident *ast.Node) ast.Kind {
	for current := ident.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindArrayBindingPattern, ast.KindObjectBindingPattern, ast.KindVariableDeclaration:
			return current.Kind
		}
	}
	return ast.KindUnknown
}

// checkMemberAccess ports checkForDanglingUnderscoreInMemberExpression.
// ESTree's MemberExpression covers both dotted and computed access, and its
// `property` is the raw key node — so `foo[_bar]` is checked exactly like
// `foo._bar`, while `foo['_bar']` has no `name` and is skipped.
func checkMemberAccess(ctx rule.RuleContext, node *ast.Node, opts noUnderscoreDangleOptions) {
	// A private name reports as `_foo`, not `#_foo`, here: only the method and
	// class-field messages re-add the `#`.
	identifier, _, ok := estreeName(memberPropertyNode(node))
	// Member access is among the most common nodes in a program, and nothing
	// below matters until the name actually dangles.
	if !ok || !hasDanglingUnderscore(identifier) {
		return
	}
	// Special-cased upstream for member access only: `foo.__proto__` stays.
	if identifier == "__proto__" {
		return
	}
	// `<Foo._bar />` is a JSXMemberExpression upstream, a kind the rule has no
	// listener for, while tsgo spells a JSX tag name with the same
	// PropertyAccessExpression it uses for value member access.
	if isJsxTagName(node) {
		return
	}

	object := ast.SkipParentheses(memberObjectNode(node))
	if opts.allowAfterThis && object.Kind == ast.KindThisKeyword {
		return
	}
	if opts.allowAfterSuper && object.Kind == ast.KindSuperKeyword {
		return
	}
	if opts.allowAfterThisConstructor && isThisConstructorReference(object) {
		return
	}

	reportIfDangling(ctx, utils.TrimNodeTextRange(ctx.SourceFile, node), identifier, identifier, opts)
}

// isJsxTagName reports whether node is part of the dotted tag name of a JSX
// element — `<a.b.c />` nests one PropertyAccessExpression inside another, so
// every link has to climb to the outermost one before asking.
func isJsxTagName(node *ast.Node) bool {
	current := node
	for current.Parent != nil && current.Parent.Kind == ast.KindPropertyAccessExpression {
		current = current.Parent
	}
	parent := current.Parent
	return parent != nil &&
		(ast.IsJsxOpeningElement(parent) || ast.IsJsxSelfClosingElement(parent) || ast.IsJsxClosingElement(parent))
}

func memberPropertyNode(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindElementAccessExpression {
		return ast.SkipParentheses(node.AsElementAccessExpression().ArgumentExpression)
	}
	return node.AsPropertyAccessExpression().Name()
}

func memberObjectNode(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindElementAccessExpression {
		return node.AsElementAccessExpression().Expression
	}
	return node.AsPropertyAccessExpression().Expression
}

// isThisConstructorReference ports the helper of the same name: the receiver
// is itself a member access whose key is named `constructor` on `this`. The
// key check reads the ESTree `property.name`, so a computed `this[constructor]`
// counts while a quoted `this['constructor']` does not.
func isThisConstructorReference(object *ast.Node) bool {
	if object.Kind != ast.KindPropertyAccessExpression && object.Kind != ast.KindElementAccessExpression {
		return false
	}
	name, _, ok := estreeName(memberPropertyNode(object))
	if !ok || name != "constructor" {
		return false
	}
	return ast.SkipParentheses(memberObjectNode(object)).Kind == ast.KindThisKeyword
}

// checkAccessorName routes a getter/setter to the method-name check only when
// it is a class member: upstream sees class accessors as MethodDefinition but
// object-literal accessors as a `Property` with `method: false`, which
// checkForDanglingUnderscoreInMethod never reports.
func checkAccessorName(ctx rule.RuleContext, node *ast.Node, opts noUnderscoreDangleOptions) {
	if node.Parent == nil || !ast.IsClassLike(node.Parent) {
		return
	}
	checkMemberName(ctx, node, opts.enforceInMethodNames, opts)
}

// checkClassField ports checkForDanglingUnderscoreInClassField.
func checkClassField(ctx rule.RuleContext, node *ast.Node, opts noUnderscoreDangleOptions) {
	// `accessor _x = 1` is an AccessorProperty upstream, a kind the rule has
	// no listener for.
	if ast.HasAccessorModifier(node) {
		return
	}
	checkMemberName(ctx, node, opts.enforceInClassFields, opts)
}

// checkMemberName reports a class member or object-literal method whose key
// dangles, gated on the option that governs that member kind.
func checkMemberName(ctx rule.RuleContext, node *ast.Node, enforce bool, opts noUnderscoreDangleOptions) {
	if !enforce {
		return
	}
	// An `abstract` member is a TSAbstractMethodDefinition /
	// TSAbstractPropertyDefinition upstream — again, no listener.
	if ast.HasAbstractModifier(node) {
		return
	}
	identifier, private, ok := estreeName(memberNameNode(node))
	if !ok {
		return
	}
	display := identifier
	if private {
		display = "#" + identifier
	}
	reportIfDangling(ctx, utils.TrimNodeTextRange(ctx.SourceFile, node), identifier, display, opts)
}
