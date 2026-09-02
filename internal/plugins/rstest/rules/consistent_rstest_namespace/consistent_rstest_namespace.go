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
		reportInvocation := func(expression *ast.Node) {
			root := namespaceInvocationRoot(expression)
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
				if !file.preferredNameFree() {
					return nil
				}
				if mode == rstestUtils.RSTEST_IMPORT_MODE && !file.bindingRewritable() {
					return nil
				}
				return []rule.RuleFix{rule.RuleFixReplace(ctx.SourceFile, root, preferred)}
			})
		}

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
				call := node.AsCallExpression()
				if call != nil {
					reportInvocation(call.Expression)
				}
			},
			ast.KindTaggedTemplateExpression: func(node *ast.Node) {
				tagged := node.AsTaggedTemplateExpression()
				if tagged != nil {
					reportInvocation(tagged.Tag)
				}
			},
		}
	},
}

// namespaceInvocationRoot returns the identifier an invocation is headed by
// when it goes through a member of that identifier. A direct `rstest(...)` call
// or tagged template has no namespace object to rewrite, so it is not a root.
func namespaceInvocationRoot(expression *ast.Node) *ast.Node {
	throughMember := false
	for expression != nil {
		switch expression.Kind {
		case ast.KindIdentifier:
			if throughMember {
				return expression
			}
			return nil
		case ast.KindParenthesizedExpression:
			expression = expression.AsParenthesizedExpression().Expression
		case ast.KindNonNullExpression:
			expression = expression.AsNonNullExpression().Expression
		case ast.KindAsExpression:
			expression = expression.AsAsExpression().Expression
		case ast.KindTypeAssertionExpression:
			expression = expression.AsTypeAssertion().Expression
		case ast.KindSatisfiesExpression:
			expression = expression.AsSatisfiesExpression().Expression
		case ast.KindCallExpression:
			expression = expression.AsCallExpression().Expression
		case ast.KindTaggedTemplateExpression:
			expression = expression.AsTaggedTemplateExpression().Tag
		case ast.KindPropertyAccessExpression:
			throughMember = true
			expression = expression.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			throughMember = true
			expression = expression.AsElementAccessExpression().Expression
		default:
			return nil
		}
	}
	return nil
}

// isNamespaceInvocationRoot is the reverse question: given an identifier, does
// it head a member call or tagged template? It is what tells a use the rule
// rewrites apart from one it leaves alone, such as `const mock = rstest.fn` or
// `[rstest]`.
func isNamespaceInvocationRoot(identifier *ast.Node) bool {
	current := identifier
	throughMember := false
	for parent := current.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			if parent.AsParenthesizedExpression().Expression != current {
				return false
			}
		case ast.KindNonNullExpression:
			if parent.AsNonNullExpression().Expression != current {
				return false
			}
		case ast.KindAsExpression:
			if parent.AsAsExpression().Expression != current {
				return false
			}
		case ast.KindTypeAssertionExpression:
			if parent.AsTypeAssertion().Expression != current {
				return false
			}
		case ast.KindSatisfiesExpression:
			if parent.AsSatisfiesExpression().Expression != current {
				return false
			}
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
		case ast.KindTaggedTemplateExpression:
			return throughMember && parent.AsTaggedTemplateExpression().Tag == current
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

// fileNamespaces answers the file-wide questions the fixes depend on: whether
// the preferred spelling is already imported, whether the preferred spelling is
// free for the fix to write, and whether every use of the imported disallowed
// spelling is one the rule rewrites. All are computed on the first fix demand,
// so a file that only reports never pays for them.
type fileNamespaces struct {
	ctx        rule.RuleContext
	preferred  string
	disallowed string

	scanned           bool
	preferredImported bool
	preferredFree     bool
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

// preferredNameFree reports whether the preferred spelling can be written into
// the file at all. It cannot when something other than an Rstest namespace
// import already declares that name: a top-level declaration would collide with
// a renamed import, and a declaration in any nested scope would capture a
// rewritten call. Scopes are not distinguished, because a fix withheld from a
// file that would have accepted it is a smaller harm than one that points a
// call at the binding next to it.
func (file *fileNamespaces) preferredNameFree() bool {
	file.scan()
	return file.preferredFree
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
	if !file.rewritable || !file.preferredFree {
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
	file.preferredFree = true
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

	// The walk answers both remaining questions at once. The preferred spelling
	// has to be checked even when the disallowed one is a global, since a call
	// rewritten in global mode is captured by a local binding just the same.
	var boundName *ast.Node
	if file.disallowedImport != nil {
		boundName = file.disallowedImport.AsImportSpecifier().Name()
	}
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if node == nil {
			return
		}
		if file.preferredFree && file.declaresPreferredName(node) {
			file.preferredFree = false
		}
		if boundName != nil && file.rewritable && file.isSurvivingUse(node, boundName) {
			file.rewritable = false
		}
		if !file.preferredFree && (boundName == nil || !file.rewritable) {
			return
		}
		node.ForEachChild(func(child *ast.Node) bool {
			visit(child)
			return false
		})
	}
	visit(sourceFile.AsNode())
}

// declaresPreferredName reports whether node declares the preferred spelling as
// something the fix would collide with. An `@rstest/core` import of either
// spelling is excluded: it binds the very namespace the fix rewrites to, so a
// call rewritten to reach it reaches the right object.
func (file *fileNamespaces) declaresPreferredName(node *ast.Node) bool {
	if node.Kind != ast.KindIdentifier ||
		node.AsIdentifier().Text != file.preferred ||
		!ast.IsDeclarationName(node) {
		return false
	}
	if node.Parent.Kind != ast.KindImportSpecifier {
		return declaresBinding(node.Parent)
	}
	return !isRstestNamespaceSpecifier(node.Parent)
}

// declaresBinding tells the declarations that introduce a name into a scope
// apart from the ones that only name a member, such as an object literal
// property or a class method.
func declaresBinding(declaration *ast.Node) bool {
	switch declaration.Kind {
	case ast.KindVariableDeclaration,
		ast.KindParameter,
		ast.KindBindingElement,
		ast.KindFunctionDeclaration,
		ast.KindFunctionExpression,
		ast.KindClassDeclaration,
		ast.KindClassExpression,
		ast.KindEnumDeclaration,
		ast.KindModuleDeclaration,
		ast.KindImportClause,
		ast.KindNamespaceImport,
		ast.KindImportEqualsDeclaration:
		return true
	default:
		return false
	}
}

// isRstestNamespaceSpecifier reports whether an import specifier brings in one
// of the two spellings of the Rstest namespace from `@rstest/core`.
func isRstestNamespaceSpecifier(element *ast.Node) bool {
	specifier := element.AsImportSpecifier()
	if specifier == nil || specifier.IsTypeOnly {
		return false
	}
	declaration := testFramework.FindImportDeclaration(element)
	if declaration == nil ||
		declaration.ModuleSpecifier == nil ||
		declaration.ModuleSpecifier.Text() != rstestUtils.RstestImportModule {
		return false
	}
	imported := specifier.Name()
	if specifier.PropertyName != nil {
		imported = specifier.PropertyName
	}
	if imported == nil || imported.Kind != ast.KindIdentifier {
		return false
	}
	name := imported.AsIdentifier().Text
	return name == namespaceRs || name == namespaceRstest
}

// isSurvivingUse reports whether node names the imported disallowed spelling
// somewhere the rule's fixes do not reach.
func (file *fileNamespaces) isSurvivingUse(node *ast.Node, boundName *ast.Node) bool {
	if node.Kind == ast.KindExportSpecifier {
		return exportSpecifierLocalName(node) == file.disallowed
	}
	if node.Kind != ast.KindIdentifier ||
		node == boundName ||
		node.AsIdentifier().Text != file.disallowed ||
		ast.IsIdentifierName(node) ||
		ast.IsDeclarationNameOrImportPropertyName(node) ||
		isNamespaceInvocationRoot(node) {
		return false
	}
	_, ok := file.resolveNamespace(node)
	return ok
}

// exportSpecifierLocalName returns the name an export specifier reads from the
// enclosing module. It is the property name when the export is aliased, since
// `export { rstest as helper }` and `export type { rstest as Namespace }` both
// name the imported binding on their left-hand side, and the exported name
// otherwise.
func exportSpecifierLocalName(node *ast.Node) string {
	specifier := node.AsExportSpecifier()
	if specifier == nil {
		return ""
	}
	local := specifier.PropertyName
	if local == nil {
		local = specifier.Name()
	}
	if local == nil || local.Kind != ast.KindIdentifier {
		return ""
	}
	return local.AsIdentifier().Text
}

// specifierRemovalRange spans the specifier at index together with the comma
// that separates it from its neighbour. A comment written for the specifier
// that stays is kept; one written for the specifier that goes is removed with
// it. Which comment is whose is read from the separating comma: the comments
// on the removed specifier's side of it go with the specifier.
func specifierRemovalRange(sourceFile *ast.SourceFile, elements []*ast.Node, index int) core.TextRange {
	elementRange := utils.TrimNodeTextRange(sourceFile, elements[index])
	if index > 0 {
		// The comma this range starts at is the one before the specifier, so the
		// comments that follow it run up to the next comma, or to the end of the
		// list when the removed specifier is the last one.
		previousEnd := utils.TrimNodeTextRange(sourceFile, elements[index-1]).End()
		limit := specifierListEnd(sourceFile, elements, index)
		return core.NewTextRange(previousEnd, commentsEnd(sourceFile.Text(), elementRange.End(), limit))
	}
	nextStart := utils.TrimNodeTextRange(sourceFile, elements[index+1]).Pos()
	return core.NewTextRange(elementRange.Pos(), firstSpecifierRemovalEnd(sourceFile.Text(), elementRange.End(), nextStart))
}

// specifierListEnd is how far after the specifier at index its own comments can
// reach: the next specifier for a specifier in the middle of the list, and the
// closing brace of the list for the last one.
func specifierListEnd(sourceFile *ast.SourceFile, elements []*ast.Node, index int) int {
	if index+1 < len(elements) {
		return utils.TrimNodeTextRange(sourceFile, elements[index+1]).Pos()
	}
	if parent := elements[index].Parent; parent != nil {
		return utils.TrimNodeTextRange(sourceFile, parent).End()
	}
	return len(sourceFile.Text())
}

// commentsEnd walks the whitespace and comments that begin at start and returns
// the end of the last comment it consumed, or start when it meets anything else
// first — the separating comma or the closing brace of the list.
func commentsEnd(text string, start int, limit int) int {
	if start < 0 || limit > len(text) || start >= limit {
		return start
	}
	end := start
	for position := start; position < limit; {
		rest := text[position:limit]
		switch {
		case rest[0] == ' ' || rest[0] == '\t' || rest[0] == '\r' || rest[0] == '\n':
			position++
		case strings.HasPrefix(rest, "//"):
			length := strings.IndexByte(rest, '\n')
			if length < 0 {
				length = len(rest)
			}
			position += length
			end = position
		case strings.HasPrefix(rest, "/*"):
			length := strings.Index(rest, "*/")
			if length < 0 {
				return end
			}
			position += length + 2
			end = position
		default:
			return end
		}
	}
	return end
}

func firstSpecifierRemovalEnd(text string, start int, end int) int {
	if start < 0 || end > len(text) || start >= end {
		return end
	}
	between := text[start:end]
	comma := strings.IndexByte(between, ',')
	if comma < 0 {
		return end
	}
	comment := firstCommentOffset(between)
	if comment < 0 {
		return end
	}
	if comment > comma {
		return start + comment
	}
	if afterComma := firstCommentOffset(between[comma+1:]); afterComma >= 0 {
		return start + comma + 1 + afterComma
	}
	return end
}

func firstCommentOffset(text string) int {
	offset := -1
	if line := strings.Index(text, "//"); line >= 0 {
		offset = line
	}
	if block := strings.Index(text, "/*"); block >= 0 && (offset < 0 || block < offset) {
		offset = block
	}
	return offset
}
