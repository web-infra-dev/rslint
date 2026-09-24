package utils

import (
	"fmt"
	"strings"
	"unicode/utf16"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

// LiteralModuleSource applies moduleVisitor's string and AMD filters to the
// shared syntax collection. Callers choose whether to include type-only imports.
func LiteralModuleSource(ref modules.Source) *ast.Node {
	if utils.IsJSDocSyntaxNode(ref.Declaration) {
		return nil
	}
	if ref.Kind == modules.ModuleReferenceAMD && len(ref.Declaration.AsCallExpression().Arguments.Nodes) != 2 {
		return nil
	}
	source := utils.ESTreeRuntimeExpression(ref.Specifier)
	if source == nil || source.Kind != ast.KindStringLiteral {
		return nil
	}
	if ref.Kind == modules.ModuleReferenceAMD && (source.Text() == "require" || source.Text() == "exports") {
		return nil
	}
	return source
}

type VisitModulesOptions struct {
	Commonjs bool
	AMD      bool
	ESModule bool
	Ignore   []string
}

// See https://github.com/import-js/eslint-plugin-import/blob/v2.32.0/utils/moduleVisitor.js
func VisitModules(visitor func(source *ast.StringLiteralLike, node *ast.Node), options VisitModulesOptions) rule.RuleListeners {
	visitors := rule.RuleListeners{}
	ignored := make([]*esregexp.RegExp, 0, len(options.Ignore))
	for _, pattern := range options.Ignore {
		// Rule schemas reject invalid user patterns before listeners are built.
		if expression, err := esregexp.Compile(moduleIgnorePattern(pattern), ""); err == nil {
			ignored = append(ignored, expression)
		}
	}

	checkSourceValue := func(source *ast.StringLiteralLike, node *ast.Node) {
		if source == nil {
			return
		}

		if len(ignored) > 0 {
			units := ecmascript.StringCodeUnitRunes(source.Text())
			for _, pattern := range ignored {
				matched, err := pattern.Unwrap().MatchRunes(units)
				// An ignore pattern that times out must not produce false positives.
				if matched || err != nil {
					return
				}
			}
		}

		visitor(source, node)
	}

	checkSource := func(node *ast.Node) {
		checkSourceValue(node.ModuleSpecifier(), node)
	}

	// for esmodule dynamic `import()` calls
	checkImportCall := func(node *ast.Node) {
		call := node.AsCallExpression()

		if call.Expression.Kind != ast.KindImportKeyword {
			return
		}

		// A recovered parse of incomplete source can leave `import()` with no
		// arguments at all.
		if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
			return
		}

		modulePath := utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])
		// Upstream accepts string Literals, not static TemplateLiterals.
		if modulePath == nil || modulePath.Kind != ast.KindStringLiteral {
			return
		}

		checkSourceValue(modulePath, call.AsNode())
	}

	// for CommonJS `require` calls
	checkCommon := func(call *ast.CallExpression) {
		checkSourceValue(CommonJSRequireSource(call), call.AsNode())
	}

	checkAMD := func(call *ast.CallExpression) {
		callee := utils.ESTreeCallCallee(call.Expression)
		if callee == nil || !ast.IsIdentifier(callee) ||
			(callee.Text() != "require" && callee.Text() != "define") ||
			call.Arguments == nil || len(call.Arguments.Nodes) != 2 {
			return
		}
		modules := utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])
		if modules == nil || modules.Kind != ast.KindArrayLiteralExpression {
			return
		}
		for _, element := range modules.AsArrayLiteralExpression().Elements.Nodes {
			source := utils.ESTreeRuntimeExpression(element)
			if source == nil || source.Kind != ast.KindStringLiteral ||
				source.Text() == "require" || source.Text() == "exports" {
				continue
			}
			checkSourceValue(source, source)
		}
	}

	if options.ESModule {
		visitors[ast.KindJSImportDeclaration] = checkSource
		visitors[ast.KindImportDeclaration] = checkSource
		visitors[ast.KindExportDeclaration] = checkSource
		visitors[ast.KindCallExpression] = checkImportCall
		// There is not `ImportExpression` in TypeScript
	}

	if options.Commonjs || options.AMD {
		currentCallExpression, ok := visitors[ast.KindCallExpression]

		visitors[ast.KindCallExpression] = func(node *ast.Node) {
			if ok {
				currentCallExpression(node)
			}
			if options.Commonjs {
				checkCommon(node.AsCallExpression())
			}
			if options.AMD {
				checkAMD(node.AsCallExpression())
			}
		}
	}

	return visitors
}

// Module ignore patterns have no Unicode flag. Spell supplementary literals as
// surrogate escapes so both the pattern and subject match UTF-16 code units.
func moduleIgnorePattern(pattern string) string {
	var source strings.Builder
	source.Grow(len(pattern))
	inClass := false
	for offset := 0; offset < len(pattern); {
		remaining := pattern[offset:]
		// Capture names are identifiers, not characters to match.
		if !inClass && (strings.HasPrefix(remaining, `\k<`) ||
			strings.HasPrefix(remaining, "(?<") && len(remaining) > 3 && remaining[3] != '=' && remaining[3] != '!') {
			if end := strings.IndexByte(remaining, '>'); end >= 0 {
				source.WriteString(remaining[:end+1])
				offset += end + 1
				continue
			}
		}
		r, size := ecmascript.DecodeStringRune(remaining)
		if r == '\\' && offset+size < len(pattern) {
			next, nextSize := ecmascript.DecodeStringRune(pattern[offset+size:])
			if next <= 0xffff && !utf16.IsSurrogate(next) {
				source.WriteString(pattern[offset : offset+size+nextSize])
				offset += size + nextSize
				continue
			}
			// A backslash before a supplementary literal is an identity escape.
			offset += size
			r, size = next, nextSize
		}
		switch {
		case r > 0xffff:
			high, low := utf16.EncodeRune(r)
			fmt.Fprintf(&source, `\u%04x\u%04x`, high, low)
		case utf16.IsSurrogate(r):
			fmt.Fprintf(&source, `\u%04x`, r)
		default:
			source.WriteString(pattern[offset : offset+size])
			switch r {
			case '[':
				inClass = true
			case ']':
				inClass = false
			}
		}
		offset += size
	}
	return source.String()
}

// CommonJSRequireSource returns the string literal in a direct require call.
// Import rules recognize the call syntactically, including shadowed require,
// without evaluating arguments or following aliases. Parentheses and JSDoc
// casts are transparent; TypeScript assertions and template literals are not.
// Callers decide whether an optional call is eligible in their AST context.
func CommonJSRequireSource(call *ast.CallExpression) *ast.StringLiteralLike {
	callee := utils.ESTreeCallCallee(call.Expression)
	if callee == nil || !ast.IsIdentifier(callee) || callee.Text() != "require" || call.Arguments == nil || len(call.Arguments.Nodes) != 1 {
		return nil
	}
	source := utils.ESTreeRuntimeExpression(call.Arguments.Nodes[0])
	if source == nil || !ast.IsStringLiteral(source) {
		return nil
	}
	return source
}
