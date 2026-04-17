package object_shorthand

import (
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// https://eslint.org/docs/latest/rules/object-shorthand

// Option modes.
const (
	modeAlways             = "always"
	modeNever              = "never"
	modeMethods            = "methods"
	modeProperties         = "properties"
	modeConsistent         = "consistent"
	modeConsistentAsNeeded = "consistent-as-needed"
)

var jsdocStarRegex = regexp.MustCompile(`^\s*\*`)

type options struct {
	apply                     string
	ignoreConstructors        bool
	methodsIgnorePattern      *regexp.Regexp
	avoidQuotes               bool
	avoidExplicitReturnArrows bool
}

func parseOptions(opts any) options {
	o := options{apply: modeAlways}

	if arr, ok := opts.([]interface{}); ok {
		if len(arr) > 0 {
			if s, ok := arr[0].(string); ok && s != "" {
				o.apply = s
			}
		}
		if len(arr) > 1 {
			if m, ok := arr[1].(map[string]interface{}); ok {
				applyObjectOptions(&o, m)
			}
		}
	} else if s, ok := opts.(string); ok && s != "" {
		o.apply = s
	} else if m := utils.GetOptionsMap(opts); m != nil {
		if s, ok := m["apply"].(string); ok && s != "" {
			o.apply = s
		}
		applyObjectOptions(&o, m)
	}

	return o
}

func applyObjectOptions(o *options, m map[string]interface{}) {
	if v, ok := m["ignoreConstructors"].(bool); ok {
		o.ignoreConstructors = v
	}
	if v, ok := m["avoidQuotes"].(bool); ok {
		o.avoidQuotes = v
	}
	if v, ok := m["avoidExplicitReturnArrows"].(bool); ok {
		o.avoidExplicitReturnArrows = v
	}
	if v, ok := m["methodsIgnorePattern"].(string); ok && v != "" {
		if re, err := regexp.Compile(v); err == nil {
			o.methodsIgnorePattern = re
		}
	}
}

// isStringLiteralKey reports whether the property name node is a string
// literal (e.g. `'foo'`), including the computed form `['foo']`. Mirrors
// ESLint's `isStringLiteral(node.key)` check, which operates on the Literal
// node itself regardless of `node.computed`.
func isStringLiteralKey(nameNode *ast.Node) bool {
	if nameNode == nil {
		return false
	}
	if nameNode.Kind == ast.KindStringLiteral {
		return true
	}
	if nameNode.Kind == ast.KindComputedPropertyName {
		if expr := nameNode.AsComputedPropertyName().Expression; expr != nil {
			return expr.Kind == ast.KindStringLiteral
		}
	}
	return false
}

// propertyKind categorizes an object literal property for the purposes of
// this rule.
type propertyKind int

const (
	propKindOther           propertyKind = iota // getter/setter/spread - skip
	propKindLongformProp                        // { a: b }
	propKindLongformMethod                      // { a: function() {} } or { a: () => {} }
	propKindShorthandProp                       // { a }
	propKindShorthandMethod                     // { a() {} }
)

// propertyValue returns a PropertyAssignment's initializer with any
// enclosing parentheses stripped. ESLint's parser discards parentheses by
// default, so rules written against its AST reference `node.value` directly
// as an Identifier / FunctionExpression / ArrowFunction. tsgo preserves
// parentheses as `ParenthesizedExpression` nodes, so every check that looks
// at the value's shape must go through this unwrap first — otherwise code
// like `{a: (a)}` or `{a: (function(){})}` silently escapes the rule.
func propertyValue(pa *ast.PropertyAssignment) *ast.Node {
	if pa == nil || pa.Initializer == nil {
		return nil
	}
	return ast.SkipParentheses(pa.Initializer)
}

func classify(node *ast.Node) propertyKind {
	switch node.Kind {
	case ast.KindGetAccessor, ast.KindSetAccessor, ast.KindSpreadAssignment:
		return propKindOther
	case ast.KindShorthandPropertyAssignment:
		return propKindShorthandProp
	case ast.KindMethodDeclaration:
		return propKindShorthandMethod
	case ast.KindPropertyAssignment:
		value := propertyValue(node.AsPropertyAssignment())
		if value == nil {
			return propKindOther
		}
		switch value.Kind {
		case ast.KindFunctionExpression, ast.KindArrowFunction:
			return propKindLongformMethod
		}
		return propKindLongformProp
	}
	return propKindOther
}

// canHaveShorthand reports whether a property can have a shorthand form.
// Getters, setters and spread elements cannot.
func canHaveShorthand(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindGetAccessor, ast.KindSetAccessor, ast.KindSpreadAssignment:
		return false
	}
	return true
}

func isShorthandKind(k propertyKind) bool {
	return k == propKindShorthandProp || k == propKindShorthandMethod
}

// isRedundantLongform reports whether a PropertyAssignment could be rewritten
// as a shorthand — i.e. its key and value carry the same name.
func isRedundantLongform(node *ast.Node) bool {
	if node.Kind == ast.KindMethodDeclaration {
		return true
	}
	if node.Kind == ast.KindPropertyAssignment {
		pa := node.AsPropertyAssignment()
		value := propertyValue(pa)
		if value == nil {
			return false
		}
		if value.Kind == ast.KindFunctionExpression {
			// Only anonymous FE counts as redundant (would become shorthand
			// method without losing the function's name).
			return value.AsFunctionExpression().Name() == nil
		}
		if value.Kind == ast.KindIdentifier {
			keyName, ok := utils.GetStaticPropertyName(pa.Name())
			if !ok {
				return false
			}
			return keyName == value.AsIdentifier().Text
		}
	}
	return false
}

// hasCommentsInsideText reports whether any `//` or `/*` sequence appears in
// the half-open source range [start, end). `utils.HasCommentsInRange` only
// checks leading/trailing comments anchored at a single position, so this
// raw-text scan is the simplest way to look for comments sprinkled *between*
// arbitrary children of a node (e.g. between the key, colon and value of a
// PropertyAssignment).
func hasCommentsInsideText(sourceText string, start, end int) bool {
	for i := start; i+1 < end; i++ {
		if sourceText[i] == '/' && (sourceText[i+1] == '/' || sourceText[i+1] == '*') {
			return true
		}
	}
	return false
}

// hasJSDocTypeAnnotationInside returns true when the node contains a JSDoc
// block comment with `@type` — these are type-suppression annotations that
// should block shorthand conversion.
//
// ESLint actually uses two subtly different JSDoc detection rules depending
// on the property-key kind:
//
//   - Identifier key (`{foo: foo}`): body matches `^\s*\*` — tolerates a
//     leading newline/whitespace before the first `*` (non-standard but
//     accepted).
//   - StringLiteral key (`{'foo': foo}`): body starts with `*` exactly — no
//     leading whitespace allowed.
//
// The `strict` parameter selects the StringLiteral-key behavior. Standard
// `/** … */` JSDoc works under both modes; the difference only shows up on
// unusual formats like `/*\n * @type … */` (leading newline before the `*`).
func hasJSDocTypeAnnotationInside(sourceText string, node *ast.Node, strict bool) bool {
	start, end := node.Pos(), node.End()
	if end > len(sourceText) {
		end = len(sourceText)
	}
	for i := start; i+1 < end; i++ {
		if sourceText[i] != '/' || sourceText[i+1] != '*' {
			continue
		}
		closeIdx := strings.Index(sourceText[i+2:end], "*/")
		if closeIdx < 0 {
			break
		}
		body := sourceText[i+2 : i+2+closeIdx]
		matchesJSDoc := strings.HasPrefix(body, "*")
		if !strict && !matchesJSDoc {
			matchesJSDoc = jsdocStarRegex.MatchString(body)
		}
		if matchesJSDoc && strings.Contains(body, "@type") {
			return true
		}
		i = i + 2 + closeIdx + 1 // skip past the closing `*/`
	}
	return false
}

// shouldIgnoreMethodName applies the `methodsIgnorePattern` option.
func shouldIgnoreMethodName(o *options, nameNode *ast.Node) bool {
	if o.methodsIgnorePattern == nil {
		return false
	}
	name, ok := utils.GetStaticPropertyName(nameNode)
	if !ok {
		return false
	}
	return o.methodsIgnorePattern.MatchString(name)
}

// isArgumentsIdentifier reports whether the node is an Identifier whose
// text is literally "arguments". The caller decides whether it is a real
// reference (see the lexical-scope check in the rule body).
func isArgumentsIdentifier(node *ast.Node) bool {
	return node.Kind == ast.KindIdentifier && node.AsIdentifier().Text == "arguments"
}

// isArgumentsShadowedInBlockScope reports whether a block-scoped declaration
// named `arguments` sits in an enclosing block between the given identifier
// and its nearest non-arrow function. When it does, ESLint's scope manager
// resolves the identifier to that shadow instead of the function's implicit
// `arguments`, so the reference is NOT collected as a lexical identifier
// that blocks arrow→method conversion.
//
// The shadow forms we recognize mirror ESLint's scope manager:
//   - `let` / `const` / `using arguments` at the top of an enclosing `Block`
//   - `for (let|const arguments …)` / `for…in` / `for…of` iteration binding
//   - `catch (arguments)` binding
//
// `var arguments` is NOT considered a shadow: ESLint hoists it to the
// function scope where it merges with the implicit `arguments` variable, so
// references still count as lexical identifiers.
func isArgumentsShadowedInBlockScope(identifier *ast.Node) bool {
	for cur := identifier.Parent; cur != nil; cur = cur.Parent {
		switch cur.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression,
			ast.KindMethodDeclaration, ast.KindGetAccessor,
			ast.KindSetAccessor, ast.KindConstructor, ast.KindSourceFile:
			return false

		case ast.KindBlock:
			block := cur.AsBlock()
			if block == nil || block.Statements == nil {
				continue
			}
			for _, stmt := range block.Statements.Nodes {
				if stmt.Kind == ast.KindVariableStatement &&
					declarationListDeclaresBlockScopedArguments(stmt.AsVariableStatement().DeclarationList) {
					return true
				}
			}

		case ast.KindForStatement:
			if declarationListDeclaresBlockScopedArguments(cur.AsForStatement().Initializer) {
				return true
			}
		case ast.KindForInStatement:
			if declarationListDeclaresBlockScopedArguments(cur.AsForInOrOfStatement().Initializer) {
				return true
			}
		case ast.KindForOfStatement:
			if declarationListDeclaresBlockScopedArguments(cur.AsForInOrOfStatement().Initializer) {
				return true
			}

		case ast.KindCatchClause:
			cc := cur.AsCatchClause()
			if cc == nil || cc.VariableDeclaration == nil {
				continue
			}
			name := cc.VariableDeclaration.Name()
			if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "arguments" {
				return true
			}
		}
	}
	return false
}

// declarationListDeclaresBlockScopedArguments reports whether a node that is
// (or wraps) a VariableDeclarationList has a `let` / `const` / `using`
// declarator named `arguments`. Accepts either a VariableDeclarationList node
// directly (for `for` initializers) or nil (returns false).
func declarationListDeclaresBlockScopedArguments(node *ast.Node) bool {
	if node == nil || node.Kind != ast.KindVariableDeclarationList {
		return false
	}
	if node.Flags&ast.NodeFlagsBlockScoped == 0 {
		return false
	}
	list := node.AsVariableDeclarationList()
	if list == nil || list.Declarations == nil {
		return false
	}
	for _, decl := range list.Declarations.Nodes {
		vd := decl.AsVariableDeclaration()
		if vd == nil {
			continue
		}
		name := vd.Name()
		if name != nil && name.Kind == ast.KindIdentifier && name.AsIdentifier().Text == "arguments" {
			return true
		}
	}
	return false
}

var ObjectShorthandRule = rule.Rule{
	Name: "object-shorthand",
	Run: func(ctx rule.RuleContext, optionsAny any) rule.RuleListeners {
		opts := parseOptions(optionsAny)
		sourceText := ctx.SourceFile.Text()

		applyToMethods := opts.apply == modeMethods || opts.apply == modeAlways
		applyToProps := opts.apply == modeProperties || opts.apply == modeAlways
		applyNever := opts.apply == modeNever
		applyConsistent := opts.apply == modeConsistent
		applyConsistentAsNeeded := opts.apply == modeConsistentAsNeeded

		// Lexical scope stack — each entry holds the set of arrow functions
		// that live in the current (non-arrow) lexical scope. The program
		// scope is seeded here because the linter does not fire a listener on
		// the SourceFile node itself.
		lexicalScopeStack := []map[*ast.Node]bool{{}}
		arrowsWithLexicalIdentifiers := map[*ast.Node]bool{}

		enterScope := func() {
			lexicalScopeStack = append(lexicalScopeStack, map[*ast.Node]bool{})
		}
		exitScope := func() {
			if len(lexicalScopeStack) > 0 {
				lexicalScopeStack = lexicalScopeStack[:len(lexicalScopeStack)-1]
			}
		}
		markCurrentLexical := func() {
			if len(lexicalScopeStack) == 0 {
				return
			}
			for arrow := range lexicalScopeStack[len(lexicalScopeStack)-1] {
				arrowsWithLexicalIdentifiers[arrow] = true
			}
		}

		// Messages.
		msgExpectedPropertyShorthand := rule.RuleMessage{
			Id:          "expectedPropertyShorthand",
			Description: "Expected property shorthand.",
		}
		msgExpectedMethodShorthand := rule.RuleMessage{
			Id:          "expectedMethodShorthand",
			Description: "Expected method shorthand.",
		}
		msgExpectedPropertyLongform := rule.RuleMessage{
			Id:          "expectedPropertyLongform",
			Description: "Expected longform property syntax.",
		}
		msgExpectedMethodLongform := rule.RuleMessage{
			Id:          "expectedMethodLongform",
			Description: "Expected longform method syntax.",
		}
		msgExpectedLiteralMethodLongform := rule.RuleMessage{
			Id:          "expectedLiteralMethodLongform",
			Description: "Expected longform method syntax for string literal keys.",
		}
		msgExpectedAllPropertiesShorthanded := rule.RuleMessage{
			Id:          "expectedAllPropertiesShorthanded",
			Description: "Expected shorthand for all properties.",
		}
		msgUnexpectedMix := rule.RuleMessage{
			Id:          "unexpectedMix",
			Description: "Unexpected mix of shorthand and non-shorthand properties.",
		}

		// -------- Autofix helpers --------

		// keyText returns the full textual representation of a property name,
		// including the surrounding brackets for ComputedPropertyName.
		keyText := func(nameNode *ast.Node) string {
			if nameNode == nil {
				return ""
			}
			r := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
			return sourceText[r.Pos():r.End()]
		}

		// fixShorthandProperty rewrites `{ key: value }` into `{ value }` when
		// key === value.name. Returns nil when any comment is present inside the
		// PropertyAssignment — replacing the whole node would drop it.
		fixShorthandProperty := func(node *ast.Node, valueName string) []rule.RuleFix {
			if hasCommentsInsideText(sourceText, node.Pos(), node.End()) {
				return nil
			}
			return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, node, valueName)}
		}

		// fixPropertyToLongform rewrites `{ x }` into `{ x: x }` for the
		// ShorthandPropertyAssignment case (`"never"` option).
		fixPropertyToLongform := func(node *ast.Node) []rule.RuleFix {
			sp := node.AsShorthandPropertyAssignment()
			if sp == nil {
				return nil
			}
			name := sp.Name()
			if name == nil || name.Kind != ast.KindIdentifier {
				return nil
			}
			ident := name.AsIdentifier().Text
			return []rule.RuleFix{rule.RuleFixInsertAfter(name, ": "+ident)}
		}

		// fixMethodToLongform rewrites a MethodDeclaration (`{ foo() {} }`) into
		// `{ foo: function() {} }` for the `"never"` / `avoidQuotes` cases.
		fixMethodToLongform := func(node *ast.Node) []rule.RuleFix {
			method := node.AsMethodDeclaration()
			if method == nil || method.Body == nil {
				return nil
			}

			isAsync := ast.HasSyntacticModifier(node, ast.ModifierFlagsAsync)
			isGenerator := method.AsteriskToken != nil

			nameNode := method.Name()
			if nameNode == nil {
				return nil
			}

			// Range to replace: start of node → end of name.
			nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			nameRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)

			keyStr := keyText(nameNode)
			header := "function"
			if isAsync {
				header = "async function"
			}
			if isGenerator {
				header += "*"
			}

			replaceRange := core.NewTextRange(nodeRange.Pos(), nameRange.End())
			return []rule.RuleFix{rule.RuleFixReplaceRange(replaceRange, keyStr+": "+header)}
		}

		// fixFunctionToMethod rewrites a PropertyAssignment whose initializer
		// is a FunctionExpression into a MethodDeclaration shorthand. Handles
		// `{a: (function(){})}` by unwrapping the outer parentheses.
		fixFunctionToMethod := func(node *ast.Node) []rule.RuleFix {
			pa := node.AsPropertyAssignment()
			fn := propertyValue(pa)
			if fn == nil || fn.Kind != ast.KindFunctionExpression {
				return nil
			}
			nameNode := pa.Name()
			if nameNode == nil {
				return nil
			}
			// Disallow fix if a comment sits between the key and the function.
			keyRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
			fnRange := utils.TrimNodeTextRange(ctx.SourceFile, fn)
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(keyRange.End(), fnRange.Pos())) {
				return nil
			}

			fe := fn.AsFunctionExpression()
			if fe == nil || fe.Body == nil {
				return nil
			}
			isAsync := ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync)
			isGenerator := fe.AsteriskToken != nil

			prefix := ""
			if isAsync {
				prefix += "async "
			}
			if isGenerator {
				prefix += "*"
			}

			// The tail we want to keep is everything after `function` (and
			// `*` when generator). For generators, AsteriskToken gives us a
			// precise end position; otherwise we scan for the `function`
			// keyword past any leading modifiers (e.g. `async`).
			var headerEnd int
			if isGenerator {
				headerEnd = fe.AsteriskToken.End()
			} else {
				kwEnd, ok := positionAfterFunctionKeyword(ctx.SourceFile, fn)
				if !ok {
					return nil
				}
				headerEnd = kwEnd
			}

			nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			keyStr := keyText(nameNode)
			tail := sourceText[headerEnd:fnRange.End()]

			return []rule.RuleFix{rule.RuleFixReplaceRange(
				core.NewTextRange(nodeRange.Pos(), nodeRange.End()),
				prefix+keyStr+tail,
			)}
		}

		// fixArrowToMethod rewrites `{ foo: (a) => { return; } }` into
		// `{ foo(a) { return; } }`. Also handles extra parentheses around the
		// arrow (`{ foo: ((a) => { … }) }`). The caller is responsible for
		// verifying the arrow has no lexical identifiers.
		fixArrowToMethod := func(node *ast.Node) []rule.RuleFix {
			pa := node.AsPropertyAssignment()
			fn := propertyValue(pa)
			if fn == nil || fn.Kind != ast.KindArrowFunction {
				return nil
			}
			arrow := fn.AsArrowFunction()
			if arrow == nil || arrow.Body == nil {
				return nil
			}
			if arrow.Body.Kind != ast.KindBlock {
				return nil
			}

			nameNode := pa.Name()
			if nameNode == nil {
				return nil
			}
			keyRange := utils.TrimNodeTextRange(ctx.SourceFile, nameNode)
			fnRange := utils.TrimNodeTextRange(ctx.SourceFile, fn)
			if utils.HasCommentsInRange(ctx.SourceFile, core.NewTextRange(keyRange.End(), fnRange.Pos())) {
				return nil
			}

			isAsync := ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync)
			if arrow.EqualsGreaterThanToken == nil {
				return nil
			}

			// The AST already identifies the `=>` token, so we can slice the
			// arrow's source text around it directly. This preserves TypeScript
			// type annotations and default values without re-parsing.
			paramsStart := fnRange.Pos()
			if mods := arrow.Modifiers(); isAsync && mods != nil && len(mods.Nodes) > 0 {
				// Skip past the `async` modifier so the params start at `(`.
				paramsStart = mods.End()
			}

			arrowTokenPos := arrow.EqualsGreaterThanToken.Pos()
			arrowTokenEnd := arrow.EqualsGreaterThanToken.End()

			paramsText := strings.TrimSpace(sourceText[paramsStart:arrowTokenPos])
			bodyText := strings.TrimLeft(sourceText[arrowTokenEnd:fnRange.End()], " \t")

			// Wrap a single identifier parameter in parentheses: `x => …`.
			if len(paramsText) == 0 || paramsText[0] != '(' {
				paramsText = "(" + paramsText + ")"
			}

			prefix := ""
			if isAsync {
				prefix = "async "
			}

			nodeRange := utils.TrimNodeTextRange(ctx.SourceFile, node)
			keyStr := keyText(nameNode)
			replacement := prefix + keyStr + paramsText + " " + bodyText

			return []rule.RuleFix{rule.RuleFixReplaceRange(
				core.NewTextRange(nodeRange.Pos(), nodeRange.End()),
				replacement,
			)}
		}

		// -------- Reporting --------

		reportMix := func(obj *ast.Node) {
			ctx.ReportNode(obj, msgUnexpectedMix)
		}
		reportAllShorthand := func(obj *ast.Node) {
			ctx.ReportNode(obj, msgExpectedAllPropertiesShorthanded)
		}

		checkConsistency := func(obj *ast.Node, checkRedundancy bool) {
			ol := obj.AsObjectLiteralExpression()
			if ol == nil || ol.Properties == nil {
				return
			}
			var considered []*ast.Node
			for _, p := range ol.Properties.Nodes {
				if canHaveShorthand(p) {
					considered = append(considered, p)
				}
			}
			if len(considered) == 0 {
				return
			}
			shorthandCount := 0
			for _, p := range considered {
				if isShorthandKind(classify(p)) {
					shorthandCount++
				}
			}
			if shorthandCount == len(considered) {
				return // all shorthand — consistent
			}
			if shorthandCount > 0 {
				reportMix(obj)
				return
			}
			if !checkRedundancy {
				return
			}
			// All longform; report if every property is redundant.
			allRedundant := true
			for _, p := range considered {
				if !isRedundantLongform(p) {
					allRedundant = false
					break
				}
			}
			if allRedundant {
				reportAllShorthand(obj)
			}
		}

		// -------- Property-level checks --------

		handleProperty := func(node *ast.Node) {
			// Only properties directly inside an object literal are considered.
			if node.Parent == nil || node.Parent.Kind != ast.KindObjectLiteralExpression {
				return
			}

			kind := classify(node)
			if kind == propKindOther {
				return
			}

			isConcise := kind == propKindShorthandProp || kind == propKindShorthandMethod

			// Computed keys can only fail the method checks.
			if node.Kind == ast.KindPropertyAssignment {
				pa := node.AsPropertyAssignment()
				if pa != nil && pa.Name() != nil && pa.Name().Kind == ast.KindComputedPropertyName {
					if pa.Initializer != nil &&
						pa.Initializer.Kind != ast.KindFunctionExpression &&
						pa.Initializer.Kind != ast.KindArrowFunction {
						return
					}
				}
			}

			if isConcise {
				// Shorthand — may need to be converted to longform.
				if kind == propKindShorthandMethod {
					method := node.AsMethodDeclaration()
					var keyNode *ast.Node
					if method != nil {
						keyNode = method.Name()
					}
					if applyNever || (opts.avoidQuotes && isStringLiteralKey(keyNode)) {
						msg := msgExpectedMethodLongform
						if !applyNever && opts.avoidQuotes {
							msg = msgExpectedLiteralMethodLongform
						}
						if fixes := fixMethodToLongform(node); fixes != nil {
							ctx.ReportNodeWithFixes(node, msg, fixes...)
						} else {
							ctx.ReportNode(node, msg)
						}
					}
				} else if applyNever {
					// `{ x }` → `{ x: x }`
					if fixes := fixPropertyToLongform(node); fixes != nil {
						ctx.ReportNodeWithFixes(node, msgExpectedPropertyLongform, fixes...)
					} else {
						ctx.ReportNode(node, msgExpectedPropertyLongform)
					}
				}
				return
			}

			// Longform — may need to be converted to shorthand.
			pa := node.AsPropertyAssignment()
			value := propertyValue(pa) // unwraps parentheses
			if value == nil {
				return
			}
			valueKind := value.Kind

			// Method-shorthand: `{ foo: function() {} }` / `{ foo: () => {} }`
			if applyToMethods && (valueKind == ast.KindFunctionExpression || valueKind == ast.KindArrowFunction) {
				if valueKind == ast.KindFunctionExpression {
					// Named FunctionExpression: skip.
					if value.AsFunctionExpression().Name() != nil {
						return
					}
				}

				nameNode := pa.Name()
				if opts.ignoreConstructors && nameNode != nil && nameNode.Kind == ast.KindIdentifier {
					if utils.IsConstructorName(nameNode.AsIdentifier().Text) {
						return
					}
				}
				if shouldIgnoreMethodName(&opts, nameNode) {
					return
				}
				if opts.avoidQuotes && isStringLiteralKey(nameNode) {
					return
				}

				if valueKind == ast.KindFunctionExpression {
					if fixes := fixFunctionToMethod(node); fixes != nil {
						ctx.ReportNodeWithFixes(node, msgExpectedMethodShorthand, fixes...)
					} else {
						ctx.ReportNode(node, msgExpectedMethodShorthand)
					}
					return
				}

				// ArrowFunction — only when avoidExplicitReturnArrows is enabled
				// and the body is a block, and the arrow does not use lexical
				// identifiers.
				arrow := value.AsArrowFunction()
				if arrow == nil || arrow.Body == nil || arrow.Body.Kind != ast.KindBlock {
					return
				}
				if !opts.avoidExplicitReturnArrows {
					return
				}
				if arrowsWithLexicalIdentifiers[value] {
					return
				}
				if fixes := fixArrowToMethod(node); fixes != nil {
					ctx.ReportNodeWithFixes(node, msgExpectedMethodShorthand, fixes...)
				} else {
					ctx.ReportNode(node, msgExpectedMethodShorthand)
				}
				return
			}

			// Property-shorthand: `{ foo: foo }` → `{ foo }`
			if !applyToProps {
				return
			}
			if valueKind != ast.KindIdentifier {
				return
			}
			valueIdent := value.AsIdentifier()
			if valueIdent == nil {
				return
			}
			nameNode := pa.Name()
			if nameNode == nil {
				return
			}

			switch nameNode.Kind {
			case ast.KindIdentifier:
				if nameNode.AsIdentifier().Text != valueIdent.Text {
					return
				}
				// Identifier-key branch uses the tolerant `^\s*\*` check —
				// leading whitespace/newline before the first `*` is allowed
				// (matches ESLint's `JSDOC_COMMENT_REGEX`).
				if hasJSDocTypeAnnotationInside(sourceText, node, false /*strict*/) {
					return
				}
				if fixes := fixShorthandProperty(node, valueIdent.Text); fixes != nil {
					ctx.ReportNodeWithFixes(node, msgExpectedPropertyShorthand, fixes...)
				} else {
					ctx.ReportNode(node, msgExpectedPropertyShorthand)
				}
			case ast.KindStringLiteral:
				if nameNode.AsStringLiteral().Text != valueIdent.Text {
					return
				}
				if opts.avoidQuotes {
					return
				}
				// StringLiteral-key branch uses the strict `startsWith("*")`
				// check — matches ESLint's tighter predicate in this branch.
				if hasJSDocTypeAnnotationInside(sourceText, node, true /*strict*/) {
					return
				}
				if fixes := fixShorthandProperty(node, valueIdent.Text); fixes != nil {
					ctx.ReportNodeWithFixes(node, msgExpectedPropertyShorthand, fixes...)
				} else {
					ctx.ReportNode(node, msgExpectedPropertyShorthand)
				}
			}
		}

		// -------- Listeners --------

		listeners := rule.RuleListeners{
			ast.KindFunctionDeclaration:                    func(n *ast.Node) { enterScope() },
			rule.ListenerOnExit(ast.KindFunctionDeclaration): func(n *ast.Node) { exitScope() },
			ast.KindFunctionExpression:                     func(n *ast.Node) { enterScope() },
			rule.ListenerOnExit(ast.KindFunctionExpression): func(n *ast.Node) { exitScope() },
			ast.KindMethodDeclaration: func(n *ast.Node) {
				enterScope()
				// Also run the property-level check (only when inside an
				// object literal) on enter: MethodDeclaration has no children
				// whose traversal matters for this check.
				handleProperty(n)
			},
			rule.ListenerOnExit(ast.KindMethodDeclaration): func(n *ast.Node) { exitScope() },
			ast.KindGetAccessor:                            func(n *ast.Node) { enterScope() },
			rule.ListenerOnExit(ast.KindGetAccessor):       func(n *ast.Node) { exitScope() },
			ast.KindSetAccessor:                            func(n *ast.Node) { enterScope() },
			rule.ListenerOnExit(ast.KindSetAccessor):       func(n *ast.Node) { exitScope() },
			ast.KindConstructor:                            func(n *ast.Node) { enterScope() },
			rule.ListenerOnExit(ast.KindConstructor):       func(n *ast.Node) { exitScope() },

			ast.KindArrowFunction: func(n *ast.Node) {
				if len(lexicalScopeStack) > 0 {
					lexicalScopeStack[len(lexicalScopeStack)-1][n] = true
				}
			},
			rule.ListenerOnExit(ast.KindArrowFunction): func(n *ast.Node) {
				if len(lexicalScopeStack) > 0 {
					delete(lexicalScopeStack[len(lexicalScopeStack)-1], n)
				}
			},

			ast.KindThisKeyword:  func(n *ast.Node) { markCurrentLexical() },
			ast.KindSuperKeyword: func(n *ast.Node) { markCurrentLexical() },
			ast.KindMetaProperty: func(n *ast.Node) {
				// `new.target`
				mp := n.AsMetaProperty()
				if mp != nil && mp.KeywordToken == ast.KindNewKeyword {
					markCurrentLexical()
				}
			},
			ast.KindIdentifier: func(n *ast.Node) {
				if !isArgumentsIdentifier(n) {
					return
				}
				// ESLint only collects `arguments` references seen inside a
				// non-arrow function — the Program scope has no implicit
				// `arguments` binding, so bare `arguments` at module / global
				// level doesn't block arrow→method conversion. We approximate
				// this with a depth check: stack[0] is the seeded program
				// scope, so depth > 1 means at least one enclosing function
				// (FunctionDeclaration, FunctionExpression, MethodDeclaration,
				// GetAccessor, SetAccessor, Constructor). Arrow functions do
				// NOT push a new entry, so they are correctly transparent.
				if len(lexicalScopeStack) <= 1 {
					return
				}
				// A block-scoped `let`/`const`/`using arguments` between the
				// reference and the enclosing function hides the function's
				// implicit `arguments` — ESLint resolves the reference to
				// that shadow, so it doesn't count as a lexical identifier.
				if isArgumentsShadowedInBlockScope(n) {
					return
				}
				markCurrentLexical()
			},

			// Property-level checks on exit so that descendant-tracking for
			// lexical identifiers inside arrow values is complete.
			rule.ListenerOnExit(ast.KindPropertyAssignment): func(n *ast.Node) {
				handleProperty(n)
			},
			ast.KindShorthandPropertyAssignment: func(n *ast.Node) {
				// Skip destructuring binding: `let {a, b} = o`.
				if n.Parent != nil && n.Parent.Kind != ast.KindObjectLiteralExpression {
					return
				}
				handleProperty(n)
			},

			ast.KindObjectLiteralExpression: func(n *ast.Node) {
				if applyConsistent {
					checkConsistency(n, false)
				} else if applyConsistentAsNeeded {
					checkConsistency(n, true)
				}
			},
		}

		return listeners
	},
}

// positionAfterFunctionKeyword returns the end offset of the `function`
// keyword inside a FunctionExpression. Used when splicing a property value
// like `async function foo()` into a method shorthand — everything past this
// position (parameters, body, type annotations) is preserved verbatim.
//
// Returns (-1, false) if the token at the expected position is not
// `function`, which only happens with malformed input.
func positionAfterFunctionKeyword(sourceFile *ast.SourceFile, fn *ast.Node) (int, bool) {
	pos := fn.Pos()
	// The FunctionExpression's range starts before any modifiers; advance past
	// the modifier list (e.g. `async`) if present.
	if fe := fn.AsFunctionExpression(); fe != nil {
		if mods := fe.Modifiers(); mods != nil && len(mods.Nodes) > 0 {
			pos = mods.End()
		}
	}
	kwRange := scanner.GetRangeOfTokenAtPosition(sourceFile, pos)
	if scanner.ScanTokenAtPosition(sourceFile, pos) != ast.KindFunctionKeyword {
		return -1, false
	}
	return kwRange.End(), true
}
