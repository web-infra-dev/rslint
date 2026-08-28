package consistent_rstest_namespace

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/microsoft/typescript-go/shim/ast"
	"github.com/microsoft/typescript-go/shim/core"
	rstestUtils "github.com/web-infra-dev/rslint/internal/plugins/rstest/utils"
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils"
	testFramework "github.com/web-infra-dev/rslint/internal/utils/test_framework"
)

//go:embed consistent_rstest_namespace.schema.json
var schemaJSON []byte

const (
	namespaceRs     = "rs"
	namespaceRstest = "rstest"
)

func parseOptions(rawOptions []any) string {
	if len(rawOptions) == 0 {
		return namespaceRs
	}
	optionsMap, _ := rawOptions[0].(map[string]any)
	if value, ok := optionsMap["fn"].(string); ok && value == namespaceRstest {
		return namespaceRstest
	}
	return namespaceRs
}

func oppositeNamespace(name string) string {
	if name == namespaceRs {
		return namespaceRstest
	}
	return namespaceRs
}

func buildConsistentNamespaceMessage(preferred, disallowed string) rule.RuleMessage {
	return rule.RuleMessage{
		Id:          "consistentNamespace",
		Description: fmt.Sprintf("Prefer using `%s` instead of `%s`", preferred, disallowed),
		Data: map[string]string{
			"namespace":         preferred,
			"oppositeNamespace": disallowed,
		},
	}
}

var ConsistentRstestNamespaceRule = rule.Rule{
	Name:             "rstest/consistent-rstest-namespace",
	RequiresTypeInfo: true,
	Schema:           rule.NewSchema(schemaJSON),
	Run: func(ctx rule.RuleContext, options []any) rule.RuleListeners {
		preferred := parseOptions(options)
		disallowed := oppositeNamespace(preferred)
		message := buildConsistentNamespaceMessage(preferred, disallowed)
		file := &fileNamespaces{ctx: ctx, preferred: preferred, disallowed: disallowed}

		// One namespace object can head several calls — `rs.mocked(fn)` and the
		// `.mockReturnValue(value)` that follows it share the same identifier —
		// so the identifier, not the call, is what gets reported once.
		reportedRoots := map[*ast.Node]bool{}

		return rule.RuleListeners{
			ast.KindImportDeclaration: func(node *ast.Node) {
				declaration := node.AsImportDeclaration()
				if declaration == nil ||
					declaration.ModuleSpecifier == nil ||
					declaration.ModuleSpecifier.Text() != rstestUtils.RstestImportModule {
					return
				}
				elements := namedImportElements(declaration)
				for index, element := range elements {
					if plainImportedName(element) != disallowed {
						continue
					}
					ctx.ReportNodeWithDeferredFixes(element, message, func() []rule.RuleFix {
						return file.importFixes(elements, index)
					})
				}
			},
			ast.KindCallExpression: func(node *ast.Node) {
				root := namespaceCallRoot(node)
				if root == nil ||
					root.AsIdentifier().Text != disallowed ||
					reportedRoots[root] {
					return
				}
				mode, ok := file.resolveNamespace(root)
				if !ok {
					return
				}
				reportedRoots[root] = true
				ctx.ReportNodeWithDeferredFixes(root, message, func() []rule.RuleFix {
					if mode == rstestUtils.RSTEST_IMPORT_MODE && !file.bindingRewritable() {
						return nil
					}
					return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, root, preferred)}
				})
			},
		}
	},
}

// namespaceCallRoot returns the identifier a call is headed by when the call
// goes through a member of that identifier. A direct `rstest(...)` call has no
// namespace object to rewrite, so it is not a root.
func namespaceCallRoot(node *ast.Node) *ast.Node {
	call := node.AsCallExpression()
	if call == nil {
		return nil
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee == nil || callee.Kind == ast.KindIdentifier {
		return nil
	}
	root := testFramework.ResolveFirstIdentifier(callee)
	if root == nil || root.Kind != ast.KindIdentifier {
		return nil
	}
	return root
}

// isNamespaceCallRoot is the reverse question: given an identifier, does it
// head a member call? It is what tells a use the rule rewrites apart from one
// it leaves alone, such as `const mock = rstest.fn` or `[rstest]`.
func isNamespaceCallRoot(identifier *ast.Node) bool {
	current := identifier
	throughMember := false
	for parent := current.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression, ast.KindNonNullExpression:
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression != current {
				return false
			}
			throughMember = true
		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression != current {
				return false
			}
			throughMember = true
		case ast.KindCallExpression:
			return throughMember && parent.AsCallExpression().Expression == current
		default:
			return false
		}
		current = parent
	}
	return false
}

func namedImportElements(declaration *ast.ImportDeclaration) []*ast.Node {
	if declaration.ImportClause == nil || declaration.ImportClause.IsTypeOnly() {
		return nil
	}
	clause := declaration.ImportClause.AsImportClause()
	if clause == nil ||
		clause.NamedBindings == nil ||
		clause.NamedBindings.Kind != ast.KindNamedImports {
		return nil
	}
	named := clause.NamedBindings.AsNamedImports()
	if named == nil || named.Elements == nil {
		return nil
	}
	return named.Elements.Nodes
}

// plainImportedName returns the name an import specifier binds when it binds it
// under its own name. An aliased specifier returns "", because `rstest as
// testUtils` introduces a name of the author's choosing rather than one of the
// two spellings this rule is about.
func plainImportedName(element *ast.Node) string {
	specifier := element.AsImportSpecifier()
	if specifier == nil || specifier.IsTypeOnly || specifier.PropertyName != nil {
		return ""
	}
	name := specifier.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.AsIdentifier().Text
}

// fileNamespaces answers the two file-wide questions the fixes depend on:
// whether the preferred spelling is already imported, and whether every use of
// the imported disallowed spelling is one the rule rewrites. Both are computed
// on the first fix demand, so a file that only reports never pays for them.
type fileNamespaces struct {
	ctx        rule.RuleContext
	preferred  string
	disallowed string

	scanned           bool
	preferredImported bool
	disallowedImport  *ast.Node
	rewritable        bool
}

func (file *fileNamespaces) resolveNamespace(identifier *ast.Node) (rstestUtils.RstestImportMode, bool) {
	original, _, mode := testFramework.ResolveFunctionIdentifierReference(
		file.disallowed,
		identifier,
		file.ctx.TypeChecker,
		file.ctx.SourceFile,
		rstestUtils.RstestImportModule,
	)
	return mode, original == file.disallowed
}

// bindingRewritable reports whether rewriting the imported disallowed spelling
// leaves the file consistent. It does not when the binding is one the import
// listener never renames — a destructured `require`, or a specifier written as
// `rstest as rstest` — and it does not when a use the rule does not rewrite
// survives, since `export { rstest }` or `const { fn } = rstest` would still
// name a binding the fix has renamed away or dropped.
func (file *fileNamespaces) bindingRewritable() bool {
	file.scan()
	return file.disallowedImport != nil && file.rewritable
}

func (file *fileNamespaces) importFixes(elements []*ast.Node, index int) []rule.RuleFix {
	file.scan()
	if !file.rewritable {
		return nil
	}
	element := elements[index]
	if !file.preferredImported {
		return []rule.RuleFix{rule.RuleFixReplace(file.ctx.SourceFile, element, file.preferred)}
	}
	// Renaming would declare the preferred spelling twice, so the disallowed
	// specifier goes instead. Dropping the last specifier of a declaration would
	// turn it into a bare side-effect import, which is a different statement
	// than the one written, so that is reported without a fix.
	if len(elements) < 2 {
		return nil
	}
	return []rule.RuleFix{rule.RuleFixRemoveRange(specifierRemovalRange(file.ctx.SourceFile, elements, index))}
}

func (file *fileNamespaces) scan() {
	if file.scanned {
		return
	}
	file.scanned = true
	file.rewritable = true
	sourceFile := file.ctx.SourceFile
	if sourceFile == nil {
		return
	}

	for _, statement := range sourceFile.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		declaration := statement.AsImportDeclaration()
		if declaration == nil ||
			declaration.ModuleSpecifier == nil ||
			declaration.ModuleSpecifier.Text() != rstestUtils.RstestImportModule {
			continue
		}
		for _, element := range namedImportElements(declaration) {
			switch plainImportedName(element) {
			case file.preferred:
				file.preferredImported = true
			case file.disallowed:
				if file.disallowedImport == nil {
					file.disallowedImport = element
				}
			}
		}
	}

	if file.disallowedImport == nil {
		return
	}
	boundName := file.disallowedImport.AsImportSpecifier().Name()
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil || !file.rewritable {
			return
		}
		if file.isSurvivingUse(node, boundName) {
			file.rewritable = false
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())
}

// isSurvivingUse reports whether node names the imported disallowed spelling
// somewhere the rule's fixes do not reach.
func (file *fileNamespaces) isSurvivingUse(node *ast.Node, boundName *ast.Node) bool {
	if node.Kind == ast.KindExportSpecifier {
		specifier := node.AsExportSpecifier()
		return specifier != nil &&
			specifier.PropertyName == nil &&
			specifier.Name() != nil &&
			specifier.Name().Text() == file.disallowed
	}
	if node.Kind != ast.KindIdentifier ||
		node == boundName ||
		node.AsIdentifier().Text != file.disallowed ||
		ast.IsIdentifierName(node) ||
		ast.IsDeclarationNameOrImportPropertyName(node) ||
		isNamespaceCallRoot(node) {
		return false
	}
	_, ok := file.resolveNamespace(node)
	return ok
}

// specifierRemovalRange spans the specifier at index together with the comma
// that separates it from its neighbour. A comment written for the specifier
// that stays is kept; one written for the specifier that goes is removed with
// it.
func specifierRemovalRange(sourceFile *ast.SourceFile, elements []*ast.Node, index int) core.TextRange {
	elementRange := utils.TrimNodeTextRange(sourceFile, elements[index])
	if index > 0 {
		previousEnd := utils.TrimNodeTextRange(sourceFile, elements[index-1]).End()
		return core.NewTextRange(previousEnd, elementRange.End())
	}
	nextStart := utils.TrimNodeTextRange(sourceFile, elements[index+1]).Pos()
	return core.NewTextRange(elementRange.Pos(), commentAwareEnd(sourceFile.Text(), elementRange.End(), nextStart))
}

func commentAwareEnd(text string, start int, end int) int {
	if start < 0 || end > len(text) || start >= end {
		return end
	}
	if offset := strings.Index(text[start:end], "//"); offset >= 0 {
		end = start + offset
	}
	if offset := strings.Index(text[start:end], "/*"); offset >= 0 {
		end = start + offset
	}
	return end
}
