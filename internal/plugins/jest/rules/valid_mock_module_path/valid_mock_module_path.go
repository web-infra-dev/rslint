package valid_mock_module_path

import (
	_ "embed"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	jestUtils "github.com/web-infra-dev/rslint/internal/plugins/jest/utils"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/moduleresolver"
	"github.com/web-infra-dev/rslint/internal/utils/modules"
)

//go:embed valid_mock_module_path.schema.json
var schemaJSON []byte

var defaultModuleFileExtensions = []string{".js", ".ts", ".tsx", ".jsx", ".json"}

// requireResolveOptions selects what Node's `require.resolve` accepts: its
// default extensions and the CommonJS export conditions, with package
// `exports` enforced.
var requireResolveOptions = moduleresolver.Options{
	Extensions: []string{".js", ".json", ".node"},
	Conditions: []string{"node", "require"},
}

func buildInvalidMockModulePathMessage(moduleName string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "invalidMockModulePath",
		Description: "Module path " + moduleName + " does not exist or is not exported",
		Data:        map[string]string{"moduleName": moduleName},
	}
}

func parseModuleFileExtensions(options []any) []string {
	if len(options) == 0 {
		return defaultModuleFileExtensions
	}
	optionsMap, ok := options[0].(map[string]any)
	if !ok {
		return defaultModuleFileExtensions
	}
	raw, ok := optionsMap["moduleFileExtensions"].([]any)
	if !ok {
		return defaultModuleFileExtensions
	}
	extensions := make([]string, 0, len(raw))
	for _, value := range raw {
		if extension, ok := value.(string); ok {
			extensions = append(extensions, extension)
		}
	}
	return extensions
}

// calleeAccessorName mirrors upstream's `isSupportedAccessor` on the callee's
// property: an identifier (computed or not), a string literal, or a template
// literal without substitutions.
func calleeAccessorName(callee *ast.Node) (string, bool) {
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		if name == nil || name.Kind != ast.KindIdentifier {
			return "", false
		}
		return name.Text(), true
	case ast.KindElementAccessExpression:
		return staticAccessorName(callee.AsElementAccessExpression().ArgumentExpression)
	}
	return "", false
}

func staticAccessorName(node *ast.Node) (string, bool) {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil {
		return "", false
	}
	switch node.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		return node.Text(), true
	}
	return "", false
}

// propertyKeyName applies the same accessor test to an object key. Like
// upstream, a computed identifier key such as `[virtual]` counts by its name.
func propertyKeyName(name *ast.Node) (string, bool) {
	if name == nil {
		return "", false
	}
	switch name.Kind {
	case ast.KindIdentifier, ast.KindStringLiteral:
		return name.Text(), true
	case ast.KindComputedPropertyName:
		return staticAccessorName(name.AsComputedPropertyName().Expression)
	}
	return "", false
}

// isTruthyLiteral answers JavaScript truthiness for an ESTree `Literal`.
// Any other expression, including a template literal, is not a literal.
func isTruthyLiteral(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindTrueKeyword, ast.KindRegularExpressionLiteral:
		return true
	case ast.KindStringLiteral:
		return node.AsStringLiteral().Text != ""
	case ast.KindNumericLiteral:
		return utils.NormalizeNumericLiteral(node.AsNumericLiteral().Text) != "0"
	case ast.KindBigIntLiteral:
		return utils.NormalizeBigIntLiteral(node.AsBigIntLiteral().Text) != "0"
	}
	return false
}

// hasTrueVirtualProperty reports an object literal with any `virtual` property
// whose value is a truthy literal, which tells Jest the module need not exist.
func hasTrueVirtualProperty(node *ast.Node) bool {
	node = utils.ESTreeRuntimeExpression(node)
	if node == nil || node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		if property.Kind != ast.KindPropertyAssignment {
			continue
		}
		if name, ok := propertyKeyName(property.Name()); !ok || name != "virtual" {
			continue
		}
		value := utils.ESTreeRuntimeExpression(property.AsPropertyAssignment().Initializer)
		if value != nil && isTruthyLiteral(value) {
			return true
		}
	}
	return false
}

// isNestedInCallChain mirrors upstream's parser, which only accepts a Jest
// call that is not itself called, passed as an argument, or accessed: in
// `jest.mock('a').mock('b')` only the outer call is checked.
func isNestedInCallChain(node *ast.Node) bool {
	parent := ast.WalkUpParenthesizedExpressions(node.Parent)
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindCallExpression, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression:
		return true
	}
	return false
}

// localModuleExists resolves a specifier starting with `.` against the linted
// file's directory and accepts any file or directory found at that path, or at
// that path followed by one of the configured extensions.
func localModuleExists(sourceProgram *program.Program, fileName, specifier string, extensions []string) bool {
	fs := sourceProgram.FS()
	if fs == nil {
		return true
	}
	base := tspath.ResolvePath(tspath.GetDirectoryPath(fileName), specifier)
	if tspath.GetRootLength(base) < len(base) {
		base = tspath.RemoveTrailingDirectorySeparator(base)
	}
	for _, extension := range append([]string{""}, extensions...) {
		candidate := tspath.NormalizePath(base + extension)
		if fs.FileExists(candidate) || fs.DirectoryExists(candidate) {
			return true
		}
	}
	return false
}

// packageModuleExists resolves any other specifier as Node's `require.resolve`
// does, searching `node_modules` directories from the linted file upward.
func packageModuleExists(sourceProgram *program.Program, fileName, specifier string) bool {
	if modules.IsNodeBuiltin(specifier) {
		return true
	}
	return moduleresolver.Resolve(sourceProgram, specifier, fileName, requireResolveOptions).Error == ""
}

var ValidMockModulePathRule = rule.Rule{
	Name:   "jest/valid-mock-module-path",
	Schema: rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		sourceProgram := ctx.Program()
		if !sourceProgram.IsValid() || ctx.SourceFile == nil {
			return rule.RuleListeners{}
		}
		analysis := jestUtils.GetJestCallAnalysis(ctx)
		moduleFileExtensions := parseModuleFileExtensions(options)
		fileName := ctx.SourceFile.FileName()

		return rule.RuleListeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				callee := utils.ESTreeCallCallee(call.Expression)
				if callee == nil {
					return
				}
				method, ok := calleeAccessorName(callee)
				if !ok || (method != "mock" && method != "doMock") {
					return
				}
				if isNestedInCallChain(node) {
					return
				}
				jestFnCall := analysis.ParseFnCall(node)
				if jestFnCall == nil || jestFnCall.Kind != jestUtils.JestFnTypeJest {
					return
				}

				arguments := call.Arguments.Nodes
				moduleName := utils.ESTreeRuntimeExpression(arguments[0])
				if moduleName == nil || moduleName.Kind != ast.KindStringLiteral {
					return
				}
				if len(arguments) > 2 && hasTrueVirtualProperty(arguments[2]) {
					return
				}

				specifier := moduleName.AsStringLiteral().Text
				if strings.HasPrefix(specifier, ".") {
					if localModuleExists(sourceProgram, fileName, specifier, moduleFileExtensions) {
						return
					}
				} else if packageModuleExists(sourceProgram, fileName, specifier) {
					return
				}

				ctx.ReportNode(node, buildInvalidMockModulePathMessage(utils.TrimmedNodeText(ctx.SourceFile, moduleName)))
			},
		}
	},
}
