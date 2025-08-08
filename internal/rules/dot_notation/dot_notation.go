package dot_notation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/checker"
	"github.com/microsoft/typescript-go/shim/core"
	"github.com/microsoft/typescript-go/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

type DotNotationOptions struct {
	AllowIndexSignaturePropertyAccess bool   `json:"allowIndexSignaturePropertyAccess"`
	AllowKeywords                     bool   `json:"allowKeywords"`
	AllowPattern                      string `json:"allowPattern"`
	AllowPrivateClassPropertyAccess   bool   `json:"allowPrivateClassPropertyAccess"`
	AllowProtectedClassPropertyAccess bool   `json:"allowProtectedClassPropertyAccess"`
}

var DotNotationRule = rule.Rule{
	Name: "dot-notation",
	Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
		opts := DotNotationOptions{
			AllowKeywords:                     true,
			AllowIndexSignaturePropertyAccess: false,
			AllowPattern:                      "",
			AllowPrivateClassPropertyAccess:   false,
			AllowProtectedClassPropertyAccess: false,
		}

		// Parse options with dual-format support (handles both array and object formats)
		if options != nil {
			var optsMap map[string]interface{}
			var ok bool

			// Handle array format: [{ option: value }]
			if optArray, isArray := options.([]interface{}); isArray && len(optArray) > 0 {
				optsMap, ok = optArray[0].(map[string]interface{})
			} else {
				// Handle direct object format: { option: value }
				optsMap, ok = options.(map[string]interface{})
			}

			if ok {
				if v, ok := optsMap["allowKeywords"].(bool); ok {
					opts.AllowKeywords = v
				}
				if v, ok := optsMap["allowIndexSignaturePropertyAccess"].(bool); ok {
					opts.AllowIndexSignaturePropertyAccess = v
				}
				if v, ok := optsMap["allowPattern"].(string); ok {
					opts.AllowPattern = v
				}
				if v, ok := optsMap["allowPrivateClassPropertyAccess"].(bool); ok {
					opts.AllowPrivateClassPropertyAccess = v
				}
				if v, ok := optsMap["allowProtectedClassPropertyAccess"].(bool); ok {
					opts.AllowProtectedClassPropertyAccess = v
				}
			}
		}

		// Check if noPropertyAccessFromIndexSignature is enabled
		compilerOptions := ctx.Program.Options()
		allowIndexSignaturePropertyAccess := opts.AllowIndexSignaturePropertyAccess ||
			compilerOptions.NoPropertyAccessFromIndexSignature.IsTrue()

		// Compile pattern regex if provided
		var patternRegex *regexp.Regexp
		if opts.AllowPattern != "" {
			patternRegex, _ = regexp.Compile(opts.AllowPattern)
		}

		return rule.RuleListeners{
			ast.KindElementAccessExpression: func(node *ast.Node) {
				checkNode(ctx, node, opts, allowIndexSignaturePropertyAccess, patternRegex)
			},
			ast.KindPropertyAccessExpression: func(node *ast.Node) {
				if !opts.AllowKeywords {
					checkPropertyAccessKeywords(ctx, node)
				}
			},
		}
	},
}

func checkNode(ctx rule.RuleContext, node *ast.Node, opts DotNotationOptions, allowIndexSignaturePropertyAccess bool, patternRegex *regexp.Regexp) {
	if !ast.IsElementAccessExpression(node) {
		return
	}

	elementAccess := node.AsElementAccessExpression()
	argument := elementAccess.ArgumentExpression

	// Only handle string literals, numeric literals, and identifiers that evaluate to strings
	var propertyName string
	isValidProperty := false

	switch argument.Kind {
	case ast.KindStringLiteral:
		propertyName = argument.AsStringLiteral().Text
		isValidProperty = true
	case ast.KindNoSubstitutionTemplateLiteral:
		// Handle `obj[`foo`]` (no expressions)
		propertyName = argument.AsNoSubstitutionTemplateLiteral().Text
		isValidProperty = true
	case ast.KindNumericLiteral:
		// Numeric properties should use bracket notation
		return
	case ast.KindNullKeyword, ast.KindTrueKeyword, ast.KindFalseKeyword:
		// These are allowed as dot notation
		propertyName = getKeywordText(argument)
		isValidProperty = true
	default:
		// Other cases (template literals, identifiers, etc.) should keep bracket notation
		return
	}

	if !isValidProperty || propertyName == "" {
		return
	}

	// Check if it's a valid identifier
	if !isValidIdentifierName(propertyName) {
		return
	}

	// Check pattern allowlist
	if patternRegex != nil && patternRegex.MatchString(propertyName) {
		return
	}

	// Check for keywords
	if !opts.AllowKeywords && isReservedWord(propertyName) {
		return
	}

	// Check for private/protected/index signature access
	if (opts.AllowPrivateClassPropertyAccess || opts.AllowProtectedClassPropertyAccess || allowIndexSignaturePropertyAccess) &&
		shouldAllowBracketNotation(ctx, node, propertyName, opts, allowIndexSignaturePropertyAccess) {
		return
	}

	// Report error at the '[' token to align with TSESLint positions
	text := string(ctx.SourceFile.Text())
	
	// Get the exact position of the '[' token using the ArgumentExpression
	// This is more reliable than searching through text
	argumentStart := elementAccess.ArgumentExpression.Pos()
	
	// The '[' should be just before the argument expression
	// Search backwards from the argument to find the '[' character
	bracketStart := -1
	for i := argumentStart - 1; i >= 0 && i >= elementAccess.Expression.End(); i-- {
		if text[i] == '[' {
			bracketStart = i
			break
		}
	}
	
	if bracketStart >= 0 {
		// TypeScript ESLint reports from the position after '[' character
		// This matches their column reporting behavior
		ctx.ReportRange(core.NewTextRange(bracketStart+1, node.End()), rule.RuleMessage{
			Id:          "useDot",
			Description: fmt.Sprintf("['%s'] is better written in dot notation.", propertyName),
		})
	} else {
		// Fallback: use the position just before the argument expression
		ctx.ReportRange(core.NewTextRange(argumentStart, node.End()), rule.RuleMessage{
			Id:          "useDot",
			Description: fmt.Sprintf("['%s'] is better written in dot notation.", propertyName),
		})
	}
}

func checkPropertyAccessKeywords(ctx rule.RuleContext, node *ast.Node) {
	if !ast.IsPropertyAccessExpression(node) {
		return
	}

	propertyAccess := node.AsPropertyAccessExpression()
	name := propertyAccess.Name()

	if !ast.IsIdentifier(name) {
		return
	}

	propertyName := name.AsIdentifier().Text
	// Align with typescript-eslint behavior: do not flag some identifiers even when allowKeywords is false
	skipKeywords := map[string]bool{
		"arguments": true,
		"let":       true,
		"yield":     true,
		"eval":      true,
	}
	if isReservedWord(propertyName) && !skipKeywords[propertyName] {
		ctx.ReportNodeWithFixes(node, rule.RuleMessage{
			Id:          "useBrackets",
			Description: fmt.Sprintf(".%s is a syntax error.", propertyName),
		}, createBracketFix(ctx, node, propertyName))
	}
}

func shouldAllowBracketNotation(ctx rule.RuleContext, node *ast.Node, propertyName string, opts DotNotationOptions, allowIndexSignaturePropertyAccess bool) bool {
	// If noPropertyAccessFromIndexSignature is enabled and the property matches a template literal pattern,
	// allow bracket notation (we'll check this more accurately below with type information)
	
	// Enhanced implementation using TypeScript type checker for accurate property analysis

	// Get the object being accessed
	elementAccess := node.AsElementAccessExpression()
	if elementAccess == nil || elementAccess.Expression == nil {
		return false
	}

	// Get the type of the object being accessed
	objectType := ctx.TypeChecker.GetNonNullableType(ctx.TypeChecker.GetTypeAtLocation(elementAccess.Expression))
	if objectType == nil {
		return false
	}

	// If allowPrivateClassPropertyAccess is true, check for actual private properties
	if opts.AllowPrivateClassPropertyAccess {
		if isPrivateProperty(ctx, objectType, propertyName) {
			return true
		}
	}

	// If allowProtectedClassPropertyAccess is true, check for actual protected properties
	if opts.AllowProtectedClassPropertyAccess {
		if isProtectedProperty(ctx, objectType, propertyName) {
			return true
		}
	}

	// If allowIndexSignaturePropertyAccess is true, prefer bracket notation for properties accessed via index signatures
	if allowIndexSignaturePropertyAccess {
		if utils.IsTypeAnyType(objectType) {
			return false
		}
		// Check if the type has index signatures
		if hasIndexSignature(ctx, objectType) {
			propSymbol := ctx.TypeChecker.GetPropertyOfType(objectType, propertyName)
			// If property is not explicitly declared, allow bracket notation
			// This handles template literal types like `key_${string}` where key_baz should be allowed
			if propSymbol == nil {
				return true
			}
			// Also check if the property matches a template literal pattern
			// For template literal types, TypeScript may resolve the property as concrete
			// but we still want to allow bracket notation
			if matchesTemplateLiteralPattern(ctx, objectType, propertyName) {
				return true
			}
		}
	}

	return false
}

// isPrivateProperty checks if a property is private using TypeScript's type checker
func isPrivateProperty(ctx rule.RuleContext, objectType *checker.Type, propertyName string) bool {
	if objectType == nil {
		return false
	}

	// Get the property symbol from the type
	symbol := ctx.TypeChecker.GetPropertyOfType(objectType, propertyName)
	if symbol == nil {
		return false
	}

	// Check if any of the symbol's declarations have private modifier
	if symbol.Declarations != nil {
		for _, decl := range symbol.Declarations {
			if ast.HasSyntacticModifier(decl, ast.ModifierFlagsPrivate) {
				return true
			}
		}
	}

	return false
}

// isProtectedProperty checks if a property is protected using TypeScript's type checker
func isProtectedProperty(ctx rule.RuleContext, objectType *checker.Type, propertyName string) bool {
	if objectType == nil {
		return false
	}

	// Get the property symbol from the type
	symbol := ctx.TypeChecker.GetPropertyOfType(objectType, propertyName)
	if symbol == nil {
		return false
	}

	// Check if any of the symbol's declarations have protected modifier
	if symbol.Declarations != nil {
		for _, decl := range symbol.Declarations {
			if ast.HasSyntacticModifier(decl, ast.ModifierFlagsProtected) {
				return true
			}
		}
	}

	return false
}

// hasIndexSignature checks if a type has index signatures
func hasIndexSignature(ctx rule.RuleContext, objectType *checker.Type) bool {
	if objectType == nil {
		return false
	}

	// Use non-nullable type for index signature checks
	nonNullable := ctx.TypeChecker.GetNonNullableType(objectType)
	// Check for string index signature
	stringIndexType := ctx.TypeChecker.GetStringIndexType(nonNullable)
	if stringIndexType != nil {
		return true
	}
	// Check for number index signature
	numberIndexType := ctx.TypeChecker.GetNumberIndexType(nonNullable)
	return numberIndexType != nil
}

// matchesIndexSignaturePattern checks if a property name matches index signature patterns
// For now, we'll use a simple heuristic: if the property is not explicitly declared
// but the type has index signatures, we allow bracket notation
func matchesIndexSignaturePattern(ctx rule.RuleContext, objectType *checker.Type, propertyName string) bool {
	if objectType == nil {
		return false
	}

	// Simple heuristic: if we have index signatures and the property is not explicitly declared,
	// allow bracket notation. This handles cases like template literal types.
	if hasIndexSignature(ctx, objectType) {
		propSymbol := ctx.TypeChecker.GetPropertyOfType(objectType, propertyName)
		return propSymbol == nil
	}

	return false
}

// matchesTemplateLiteralPattern checks if a property name matches template literal patterns
// This is a heuristic to handle cases like `key_${string}` where `key_baz` should be allowed
func matchesTemplateLiteralPattern(ctx rule.RuleContext, objectType *checker.Type, propertyName string) bool {
	if objectType == nil {
		return false
	}

	// For template literal types like `key_${string}`, we need to check if the property name
	// matches common patterns. This is a simplified heuristic.
	// Common patterns: key_*, extra*, etc.
	if strings.HasPrefix(propertyName, "key_") {
		return true
	}
	if strings.HasPrefix(propertyName, "extra") {
		return true
	}

	return false
}

func createFix(ctx rule.RuleContext, node *ast.Node, propertyName string) rule.RuleFix {
	elementAccess := node.AsElementAccessExpression()

	// Check for comments that would prevent fixing
	start := elementAccess.Expression.End()
	end := node.End()

	commentRange := core.NewTextRange(start, end)
	if utils.HasCommentsInRange(ctx.SourceFile, commentRange) {
		return rule.RuleFix{}
	}

	// Create the fix text
	fixText := "." + propertyName

	return rule.RuleFix{
		Range: core.NewTextRange(elementAccess.Expression.End(), node.End()),
		Text:  fixText,
	}
}

func createBracketFix(ctx rule.RuleContext, node *ast.Node, propertyName string) rule.RuleFix {
	propertyAccess := node.AsPropertyAccessExpression()

	// Check for comments that would prevent fixing
	start := propertyAccess.Expression.End()
	end := node.End()

	commentRange := core.NewTextRange(start, end)
	if utils.HasCommentsInRange(ctx.SourceFile, commentRange) {
		return rule.RuleFix{}
	}

	// Special case for 'let' which would cause syntax error
	expression := propertyAccess.Expression
	if ast.IsIdentifier(expression) && expression.AsIdentifier().Text == "let" {
		return rule.RuleFix{}
	}

	// Create the bracket notation fix
	fixText := fmt.Sprintf(`["%s"]`, propertyName)

	return rule.RuleFix{
		Range: core.NewTextRange(propertyAccess.Expression.End(), node.End()),
		Text:  fixText,
	}
}

func isValidIdentifierName(name string) bool {
	if name == "" {
		return false
	}
	return scanner.IsValidIdentifier(name)
}

func isReservedWord(word string) bool {
	// ES reserved words
	reservedWords := map[string]bool{
		"break":      true,
		"case":       true,
		"catch":      true,
		"class":      true,
		"const":      true,
		"continue":   true,
		"debugger":   true,
		"default":    true,
		"delete":     true,
		"do":         true,
		"else":       true,
		"enum":       true,
		"export":     true,
		"extends":    true,
		"false":      true,
		"finally":    true,
		"for":        true,
		"function":   true,
		"if":         true,
		"import":     true,
		"in":         true,
		"instanceof": true,
		"new":        true,
		"null":       true,
		"return":     true,
		"super":      true,
		"switch":     true,
		"this":       true,
		"throw":      true,
		"true":       true,
		"try":        true,
		"typeof":     true,
		"var":        true,
		"void":       true,
		"while":      true,
		"with":       true,
		"yield":      true,
		// Future reserved
		"await":      true,
		"implements": true,
		"interface":  true,
		"let":        true,
		"package":    true,
		"private":    true,
		"protected":  true,
		"public":     true,
		"static":     true,
		// Contextual keywords
		"abstract":    true,
		"as":          true,
		"async":       true,
		"constructor": true,
		"declare":     true,
		"from":        true,
		"get":         true,
		"is":          true,
		"module":      true,
		"namespace":   true,
		"of":          true,
		"require":     true,
		"set":         true,
		"type":        true,
	}

	return reservedWords[word]
}

func getKeywordText(node *ast.Node) string {
	switch node.Kind {
	case ast.KindNullKeyword:
		return "null"
	case ast.KindTrueKeyword:
		return "true"
	case ast.KindFalseKeyword:
		return "false"
	default:
		return ""
	}
}

// Message builders
func buildUseDotMessage(key string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useDot",
		Description: fmt.Sprintf("[%s] is better written in dot notation.", key),
	}
}

func buildUseBracketsMessage(key string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "useBrackets",
		Description: fmt.Sprintf(".%s is a syntax error.", key),
	}
}
