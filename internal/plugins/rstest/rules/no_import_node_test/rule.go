package no_import_node

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	internalUtils "github.com/web-infra-dev/rslint/internal/utils"
)

const (
	nodeTestModule     = "node:test"
	nodeTestSubpathPfx = nodeTestModule + "/"
)

// isNodeTestModule reports whether a module specifier resolves to Node's test
// runner: `node:test` itself, or one of its sub-paths such as
// `node:test/reporters`. Only the exact `node:test` specifier has a candidate
// replacement in the Rstest API, but every form is reported.
func isNodeTestModule(specifier string) bool {
	return specifier == nodeTestModule || strings.HasPrefix(specifier, nodeTestSubpathPfx)
}

// replacementModuleForFile picks the specifier the fix writes.
//
// The same test API is reachable as `@rstest/core` and as `rstack/test`, and a
// project set up through the Rstack CLI depends on `rstack` alone: writing
// `@rstest/core` into such a file produces an import that need not resolve.
// The file's own spelling is therefore what the fix follows, and only a file
// that says `rstack/test` and never says `@rstest/core` gets the Rstack
// spelling. A file with no Rstest reference at all — one relying on the test
// globals without a type reference, say — keeps `@rstest/core`, the specifier
// that resolves in every project that depends on Rstest directly.
func replacementModuleForFile(sourceFile *ast.SourceFile) string {
	if sourceFile == nil {
		return rstestUtils.RstestImportModule
	}

	sawRstack := false
	for _, directive := range sourceFile.TypeReferenceDirectives {
		if directive == nil {
			continue
		}
		switch {
		case directive.FileName == rstestUtils.RstestImportModule ||
			strings.HasPrefix(directive.FileName, rstestUtils.RstestImportModule+"/"):
			return rstestUtils.RstestImportModule
		case directive.FileName == rstestUtils.RstackTestImportModule ||
			strings.HasPrefix(directive.FileName, rstestUtils.RstackTestImportModule+"/"):
			sawRstack = true
		}
	}

	// A `require()` can sit anywhere — inside a function body, as the object of
	// a property access — so the whole file is walked rather than only its
	// top-level statements.
	sawRstest := false
	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if specifier, ok := referencedModuleSpecifier(node); ok {
			switch specifier {
			case rstestUtils.RstestImportModule:
				sawRstest = true
				return true
			case rstestUtils.RstackTestImportModule:
				sawRstack = true
			}
		}
		return node.ForEachChild(visit)
	}
	sourceFile.AsNode().ForEachChild(visit)

	if sawRstest {
		return rstestUtils.RstestImportModule
	}
	if sawRstack {
		return rstestUtils.RstackTestImportModule
	}
	return rstestUtils.RstestImportModule
}

// referencedModuleSpecifier returns the module a node names, in any of the
// three static forms a test file writes: `import ... from 'm'`,
// `import x = require('m')`, and a `require('m')` call.
func referencedModuleSpecifier(node *ast.Node) (string, bool) {
	switch node.Kind {
	case ast.KindImportDeclaration, ast.KindJSImportDeclaration:
		declaration := node.AsImportDeclaration()
		if declaration == nil || declaration.ModuleSpecifier == nil ||
			!isStringNode(declaration.ModuleSpecifier) {
			return "", false
		}
		return declaration.ModuleSpecifier.Text(), true

	case ast.KindImportEqualsDeclaration:
		declaration := node.AsImportEqualsDeclaration()
		if declaration == nil || declaration.ModuleReference == nil ||
			declaration.ModuleReference.Kind != ast.KindExternalModuleReference {
			return "", false
		}
		reference := declaration.ModuleReference.AsExternalModuleReference()
		if reference == nil || reference.Expression == nil {
			return "", false
		}
		specifier := ast.SkipParentheses(reference.Expression)
		if specifier == nil || !isStringNode(specifier) {
			return "", false
		}
		return specifier.Text(), true

	case ast.KindCallExpression:
		return requiredModuleSpecifier(node)

	default:
		return "", false
	}
}

// requiredModuleSpecifier returns the module a `require('m')` call names.
func requiredModuleSpecifier(node *ast.Node) (string, bool) {
	if node == nil {
		return "", false
	}
	node = ast.SkipParentheses(node)
	if node == nil || !ast.IsRequireCall(node, true /* requireStringLiteralLikeArgument */) {
		return "", false
	}
	arguments := node.Arguments()
	if len(arguments) == 0 || arguments[0] == nil {
		return "", false
	}
	specifier := ast.SkipParentheses(arguments[0])
	if specifier == nil || !isStringNode(specifier) {
		return "", false
	}
	return specifier.Text(), true
}

var safeImportedNames = map[string]bool{
	"describe": true,
	"it":       true,
	"test":     true,
}

func noImportNodeTestMessage() rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "noImportNodeTest",
		Description: "Do not import the Node test runner in Rstest test files",
	}
}

func isStringNode(node *ast.Node) bool {
	return node.Kind == ast.KindStringLiteral ||
		node.Kind == ast.KindNoSubstitutionTemplateLiteral
}

func canSafelyReplaceModule(declaration *ast.ImportDeclaration) bool {
	if declaration == nil || declaration.ImportClause == nil {
		return false
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil || clause.Name() != nil || clause.NamedBindings == nil ||
		clause.NamedBindings.Kind != ast.KindNamedImports {
		return false
	}
	named := clause.NamedBindings.AsNamedImports()
	if named == nil || named.Elements == nil || len(named.Elements.Nodes) == 0 {
		return false
	}
	for _, element := range named.Elements.Nodes {
		specifier := element.AsImportSpecifier()
		if specifier == nil || specifier.Name() == nil {
			return false
		}
		importedName := specifier.Name().Text()
		if specifier.PropertyName != nil {
			importedName = specifier.PropertyName.Text()
		}
		if !safeImportedNames[importedName] {
			return false
		}
	}
	return true
}

func replacementForModuleSpecifier(sourceFile *ast.SourceFile, node *ast.Node, replacementModule string) (string, bool) {
	if sourceFile == nil || node == nil {
		return "", false
	}
	trimmed := internalUtils.TrimNodeTextRange(sourceFile, node)
	text := sourceFile.Text()
	if trimmed.Pos() < 0 || trimmed.End() > len(text) || trimmed.End()-trimmed.Pos() < 2 {
		return "", false
	}
	raw := text[trimmed.Pos():trimmed.End()]
	quote := raw[0]
	if (quote != '\'' && quote != '"') || raw[len(raw)-1] != quote {
		return "", false
	}
	return string(quote) + replacementModule + string(quote), true
}

var NoImportNodeTestRule = rule.Rule{
	Name:   "rstest/no-import-node-test",
	Schema: rule.EmptyArraySchema,
	Run: func(ctx rule.RuleContext, _ []any) rule.RuleListeners {
		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				// The parser accepts an arbitrary expression as the module
				// specifier and only rejects non-literals in a later grammar
				// check, so recoverable-but-invalid source such as
				// `import x from a.b` reaches this listener. `Text()` panics on
				// those kinds, so the specifier has to be narrowed first.
				if declaration == nil || declaration.ModuleSpecifier == nil ||
					!isStringNode(declaration.ModuleSpecifier) {
					return
				}
				specifier := declaration.ModuleSpecifier.Text()
				if !isNodeTestModule(specifier) {
					return
				}
				message := noImportNodeTestMessage()
				if specifier != nodeTestModule || !canSafelyReplaceModule(declaration) {
					ctx.ReportNode(node, message)
					return
				}
				replacement, ok := replacementForModuleSpecifier(
					ctx.SourceFile,
					declaration.ModuleSpecifier,
					replacementModuleForFile(ctx.SourceFile),
				)
				if !ok {
					ctx.ReportNode(node, message)
					return
				}
				ctx.ReportNodeWithFixes(
					node,
					message,
					rule.RuleFixReplaceRange(
						internalUtils.TrimNodeTextRange(ctx.SourceFile, declaration.ModuleSpecifier),
						replacement,
					),
				)
			},
			// `import nodeTest = require('node:test')` is a static runner import
			// too, but TypeScript parses it as its own declaration kind rather
			// than a call expression, so it needs its own listener. Like
			// `require()`, it is reported without a fix.
			ast.KindImportEqualsDeclaration: func(node *ast.Node) {
				declaration := node.AsImportEqualsDeclaration()
				if declaration == nil || declaration.ModuleReference == nil ||
					declaration.ModuleReference.Kind != ast.KindExternalModuleReference {
					return
				}
				reference := declaration.ModuleReference.AsExternalModuleReference()
				if reference == nil || reference.Expression == nil {
					return
				}
				specifier := ast.SkipParentheses(reference.Expression)
				if specifier == nil || !isStringNode(specifier) || !isNodeTestModule(specifier.Text()) {
					return
				}
				ctx.ReportNode(node, noImportNodeTestMessage())
			},
			// `require('node:test')` is reported as well, but never fixed: the
			// binding forms a require can take do not map onto a specifier
			// rewrite the way a static import does.
			ast.KindCallExpression: func(node *ast.Node) {
				// Parentheses are transparent, so `(require)('node:test')` and
				// `require(('node:test'))` are the same import as the bare form.
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if callee == nil || callee.Kind != ast.KindIdentifier || callee.AsIdentifier().Text != "require" {
					return
				}
				arguments := node.Arguments()
				if len(arguments) == 0 {
					return
				}
				firstArg := ast.SkipParentheses(arguments[0])
				if firstArg == nil || !isStringNode(firstArg) || !isNodeTestModule(firstArg.Text()) {
					return
				}
				ctx.ReportNode(firstArg, noImportNodeTestMessage())
			},
		}
	},
}
