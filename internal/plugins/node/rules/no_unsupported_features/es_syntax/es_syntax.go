package es_syntax

import (
	"cmp"
	_ "embed"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/plugins/node/nodeutil"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
)

//go:embed es_syntax.schema.json
var schemaJSON []byte

// https://github.com/eslint-community/eslint-plugin-n/blob/v18.3.0/lib/rules/no-unsupported-features/es-syntax.js
var ESSyntaxRule = rule.Rule{
	Name:   "node/no-unsupported-features/es-syntax",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		var opts map[string]any
		if len(options) > 0 {
			opts, _ = options[0].(map[string]any)
		}
		version := nodeutil.ConfiguredNodeVersion(ctx, opts)
		ignores, _ := opts["ignores"].([]any)
		ignored := make(map[string]bool, len(ignores))
		for _, value := range ignores {
			if name, ok := value.(string); ok {
				ignored[name] = true
			}
		}
		check := &syntaxChecker{ctx: ctx, version: version.Raw(), enabled: map[string]featureCheck{}, prototypes: map[string][]int{}}
		supported := map[string]bool{}
		supports := func(raw string) bool {
			if value, ok := supported[raw]; ok {
				return value
			}
			value := version.IsSubsetOf(raw)
			supported[raw] = value
			return value
		}
		for index, f := range features {
			if len(ignored) > 0 && (ignored[f.name] || ignored["no-"+f.name] || ignored[camelName(f.name)] || slices.ContainsFunc(f.aliases, func(name string) bool { return ignored[name] })) {
				continue
			}
			required := f.supported
			if f.sloppy != "" {
				required = f.sloppy
			}
			if !supports(required) {
				check.enabled[f.name] = featureCheck{index: index, strictSupported: f.sloppy != "" && supports(f.supported)}
				// Index only enabled prototype features once per file. Most member
				// names have no matching feature; each feature is listed only once.
				for _, prototype := range f.prototypes {
					for _, name := range prototype.properties {
						indices := check.prototypes[name]
						if len(indices) == 0 || indices[len(indices)-1] != index {
							check.prototypes[name] = append(indices, index)
						}
					}
				}
			}
		}
		if len(check.enabled) == 0 {
			return nil
		}
		listeners := rule.RuleListeners{}
		for _, kind := range []ast.Kind{
			ast.KindArrowFunction, ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor,
			ast.KindNumericLiteral, ast.KindBigIntLiteral, ast.KindStringLiteral, ast.KindRegularExpressionLiteral, ast.KindIdentifier, ast.KindPrivateIdentifier,
			ast.KindVariableStatement, ast.KindVariableDeclarationList, ast.KindVariableDeclaration, ast.KindParameter, ast.KindBindingElement,
			ast.KindClassDeclaration, ast.KindClassExpression, ast.KindClassStaticBlockDeclaration, ast.KindPropertyDeclaration,
			ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
			ast.KindObjectBindingPattern, ast.KindArrayBindingPattern, ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression,
			ast.KindForOfStatement, ast.KindForInStatement, ast.KindAwaitExpression, ast.KindCatchClause,
			ast.KindBinaryExpression, ast.KindCallExpression, ast.KindNewExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression,
			ast.KindSpreadAssignment, ast.KindMetaProperty, ast.KindSuperKeyword,
			ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression, ast.KindTemplateHead, ast.KindTemplateMiddle, ast.KindTemplateTail, ast.KindTaggedTemplateExpression,
			ast.KindImportDeclaration, ast.KindExportDeclaration, ast.KindExportAssignment,
			ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration, ast.KindEnumDeclaration, ast.KindModuleDeclaration, ast.KindImportEqualsDeclaration,
		} {
			listeners[kind] = check.visit
		}
		// The same tsgo nodes represent expressions and assignment patterns.
		listeners[rule.ListenerOnNotAllowPattern(ast.KindArrayLiteralExpression)] = func(node *ast.Node) {
			for _, element := range node.AsArrayLiteralExpression().Elements.Nodes {
				if element.Kind == ast.KindSpreadElement {
					check.report("spread-elements", element)
				}
			}
		}
		listeners[rule.ListenerOnNotAllowPattern(ast.KindObjectLiteralExpression)] = func(node *ast.Node) {
			for _, property := range node.Properties() {
				if property.Kind == ast.KindShorthandPropertyAssignment {
					check.report("property-shorthands", property)
				}
			}
		}
		listeners[rule.ListenerOnExit(ast.KindEndOfFile)] = func(*ast.Node) {
			check.checkBuiltins()
			text := ctx.SourceFile.Text()
			if shebang := scanner.GetShebang(text); shebang != "" {
				// ESLint treats hashbangs terminated by U+2028/U+2029 as line comments.
				if rest := text[len(shebang):]; !strings.HasPrefix(rest, "\u2028") && !strings.HasPrefix(rest, "\u2029") {
					check.reportRange("hashbang", ctx.SourceFile.AsNode(), core.NewTextRange(0, len(shebang)))
				}
			}
			slices.SortStableFunc(check.diagnostics, func(a, b syntaxDiagnostic) int {
				return cmp.Compare(a.loc.Pos(), b.loc.Pos())
			})
			for _, diagnostic := range check.diagnostics {
				ctx.ReportRange(diagnostic.loc, diagnostic.message)
			}
		}
		return listeners
	},
}

type featureCheck struct {
	index           int
	strictSupported bool
}
type syntaxDiagnostic struct {
	loc     core.TextRange
	message rule.RuleMessage
}
type syntaxChecker struct {
	ctx             rule.RuleContext
	version         string
	enabled         map[string]featureCheck
	prototypes      map[string][]int
	diagnostics     []syntaxDiagnostic
	evaluator       *utils.StaticStringEvaluator
	expressionTypes map[*ast.Node]string
}

func camelName(name string) string {
	parts := strings.Split(name, "-")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 && parts[i][0] >= 'a' && parts[i][0] <= 'z' {
			parts[i] = string(parts[i][0]-'a'+'A') + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func (c *syntaxChecker) report(name string, node *ast.Node) {
	if _, enabled := c.enabled[name]; node != nil && enabled {
		node = ast.SkipParentheses(node)
		c.reportRange(name, node, utils.TrimNodeTextRange(c.ctx.SourceFile, node))
	}
}
func (c *syntaxChecker) reportRange(name string, node *ast.Node, loc core.TextRange) {
	check, ok := c.enabled[name]
	if !ok {
		return
	}
	f := features[check.index]
	supported := f.supported
	if f.sloppy != "" {
		if utils.IsInStrictModeWithSourceType(node, c.ctx.SourceFile, c.ctx.LanguageOptions.EffectiveSourceType()) {
			if check.strictSupported {
				return
			}
		} else {
			supported = f.sloppy
		}
	}
	id := "not-supported-till"
	message := "'" + name + "' is not supported until Node.js " + supported + "."
	if supported == "<0" {
		id = "not-supported-yet"
		message = "'" + name + "' is not supported in Node.js."
	}
	c.diagnostics = append(c.diagnostics, syntaxDiagnostic{loc, rule.RuleMessage{Id: id, Description: message + " The configured version range is '" + c.version + "'."}})
}

func (c *syntaxChecker) static() *utils.StaticStringEvaluator {
	if c.evaluator == nil {
		c.evaluator = utils.NewStaticStringEvaluatorWithReferenceResolver(nil, c.ctx.SourceFile, c.ctx.Refs)
	}
	return c.evaluator
}
