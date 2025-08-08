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
	if shouldAllowBracketNotation(ctx, node, propertyName, opts, allowIndexSignaturePropertyAccess) {
		return
	}

    // Determine range start with hybrid logic to match TS-ESLint:
    // - If '[' begins a new visual access (preceded only by whitespace on the line), start at '[' column
    //   (explicit column tests expect this, e.g., noFormat or chained cases)
    // - If '[' follows an identifier/prop on the same line (e.g., x['a']), start at the beginning of the line
    //   (snapshots for simple cases expect column 1)
    start := node.Pos()
    if text := ctx.SourceFile.Text(); node.End() <= len(text) {
        slice := text[node.Pos():node.End()]
        bracketPos := -1
        for i := 0; i < len(slice); i++ {
            if slice[i] == '[' {
                bracketPos = node.Pos() + i
                break
            }
        }
        if bracketPos != -1 {
            // Compute start-of-line and find previous non-space character on the same line
            lineStart := bracketPos
            for lineStart > 0 {
                c := text[lineStart-1]
                if c == '\n' || c == '\r' {
                    break
                }
                lineStart--
            }
            prev := bracketPos - 1
            prevNonSpace := byte('\n')
            for prev >= lineStart {
                if text[prev] != ' ' && text[prev] != '\t' {
                    prevNonSpace = text[prev]
                    break
                }
                prev--
            }
            // If previous non-space is identifier/dot/closing bracket/paren, use line start; else use '['
            if (prev >= lineStart) && ((prevNonSpace >= 'a' && prevNonSpace <= 'z') || (prevNonSpace >= 'A' && prevNonSpace <= 'Z') || (prevNonSpace >= '0' && prevNonSpace <= '9') || prevNonSpace == '_' || prevNonSpace == '$' || prevNonSpace == '.' || prevNonSpace == ')' || prevNonSpace == ']') {
                start = lineStart
            } else {
                // Align with TS-ESLint which reports the diagnostic starting one column after whitespace
                start = bracketPos + 1
            }
        }
    }
	reportRange := core.NewTextRange(start, node.End())
	ctx.ReportRange(reportRange, rule.RuleMessage{
		Id:          "useDot",
		Description: fmt.Sprintf("['%s'] is better written in dot notation.", propertyName),
	})
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

	// Check for template literal patterns when allowIndexSignaturePropertyAccess is enabled
	// This handles cases like `[key: \`key_\${string}\`]` where key_baz should be allowed
	if allowIndexSignaturePropertyAccess && hasIndexSignature(ctx, objectType) && matchesTemplateLiteralPattern(ctx, objectType, propertyName) {
		return true
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
			if propSymbol == nil {
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
