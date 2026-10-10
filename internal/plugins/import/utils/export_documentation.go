package utils

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/rule"
	rslint_utils "github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

// Deprecation is immutable documentation metadata shared by export-map readers.
type Deprecation struct{ Description string }

// exportDocSource records authored nodes without scanning their comments.
// Existing export-map consumers pay no documentation parsing cost.
type exportDocSource struct {
	file  *ast.SourceFile
	nodes []*ast.Node
}

type exportDocumentation struct {
	file         *ast.SourceFile
	comments     []*ast.CommentRange
	styles       []string
	deprecations rslint_utils.LazyMap[*exportDocSource, *Deprecation]
	module       *Deprecation
}

func newExportDocumentation(file *ast.SourceFile, styles []string) *exportDocumentation {
	docs := &exportDocumentation{file: file, comments: rule.NewCommentStore(file).All(), styles: styles}
	docs.module = docs.moduleDeprecation()
	return docs
}

func (index *ModuleIndex) documentationOf(file *ast.SourceFile) *exportDocumentation {
	return index.documentation.Get(file, func() *exportDocumentation {
		return newExportDocumentation(file, index.settings.docStyles)
	})
}

// Deprecation reads a module's first @module block, independently of docstyle.
func (m *ExportMap) Deprecation(ctx rule.RuleContext) *Deprecation {
	if m == nil || m.docFile == nil || !exportDocumentationAvailable(ctx, m.docFile) {
		return nil
	}
	return IndexFor(ctx).documentationOf(m.docFile).module
}

// Deprecation reads export documentation using this consumer's import/docstyle.
// The index separates settings and caches both successful and absent results.
func (m *ExportMeta) Deprecation(ctx rule.RuleContext) *Deprecation {
	if m == nil || m.doc == nil || len(m.doc.nodes) == 0 || !exportDocumentationAvailable(ctx, m.doc.file) {
		return nil
	}
	docs := IndexFor(ctx).documentationOf(m.doc.file)
	// Namespace variables share their first candidate but have different fallback
	// comments. Cache the complete source, not just that first namespace node.
	return docs.deprecations.Get(m.doc, func() *Deprecation {
		return docs.capture(m.doc.nodes...)
	})
}

func exportDocumentationAvailable(ctx rule.RuleContext, file *ast.SourceFile) bool {
	return exportExtensionAllowed(ctx.Settings, file.FileName()) && ctx.Program().IsValid() &&
		len(ctx.Program().SyntacticDiagnostics(context.Background(), file)) == 0
}

// DefaultDeprecation includes documentation on TypeScript export assignments
// whose synthetic default is deliberately excluded from ordinary Names/Size.
func (m *ExportMap) DefaultDeprecation(ctx rule.RuleContext) *Deprecation {
	if m == nil {
		return nil
	}
	if meta := m.Get(defaultExportName); meta != nil {
		return meta.Deprecation(ctx)
	}
	return (&ExportMeta{doc: m.defaultDoc}).Deprecation(ctx)
}

func (local *localExports) captureTypeScriptExportDocs(file *ast.SourceFile, export *ast.Node, first int) {
	expression := export.Name()
	if export.Kind == ast.KindExportAssignment {
		expression = export.AsExportAssignment().Expression
	}
	name := referencedIdentifierText(expression)
	found := false
	for _, decl := range file.Statements.Nodes {
		if isExportedDeclaration(decl) {
			continue
		}
		matches := false
		switch decl.Kind {
		case ast.KindVariableStatement:
			for _, variable := range decl.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
				if variable.Name().Kind == ast.KindIdentifier && variable.Name().Text() == name {
					matches = true
				}
			}
		case ast.KindClassDeclaration, ast.KindEnumDeclaration, ast.KindTypeAliasDeclaration, ast.KindInterfaceDeclaration, ast.KindModuleDeclaration:
			matches = decl.Name() != nil && decl.Name().Text() == name
		case ast.KindFunctionDeclaration:
			// Upstream only includes TSDeclareFunction, not runtime functions.
			matches = decl.Body() == nil && decl.Name() != nil && decl.Name().Text() == name
		}
		if !matches {
			continue
		}
		found = true
		if decl.Kind != ast.KindModuleDeclaration {
			local.DefaultDoc = &exportDocSource{file: file, nodes: []*ast.Node{decl}}
			continue
		}
		body := decl.AsModuleDeclaration().Body
		if body == nil || body.Kind != ast.KindModuleBlock {
			continue
		}
		for _, member := range body.AsModuleBlock().Statements.Nodes {
			nodes := []*ast.Node{member}
			if member.Kind == ast.KindVariableStatement {
				nodes = []*ast.Node{decl, member}
			}
			doc := &exportDocSource{file: file, nodes: nodes}
			for _, name := range exportedDeclarationNames(member) {
				for i := first; i < len(local.Steps); i++ {
					if local.Steps[i].Docs == nil {
						local.Steps[i].Docs = make(map[string]*exportDocSource)
					}
					local.Steps[i].Docs[name] = doc
				}
			}
		}
	}
	if !found {
		local.DefaultDoc = &exportDocSource{file: file, nodes: []*ast.Node{export}}
	}
}

// Upstream captureDoc stops at the first node with any leading comments, even
// when those comments contain no documentation. This differs from tsgo's JSDoc
// inheritance, and also accepts ordinary block comments and TomDoc line comments.
func (d *exportDocumentation) capture(nodes ...*ast.Node) *Deprecation {
	for _, node := range nodes {
		if node == nil {
			continue
		}
		start := rslint_utils.TrimNodeTextRange(d.file, node).Pos()
		end := sortCommentEnd(d.comments, start)
		first := end
		cursor := start
		for first > 0 {
			c := d.comments[first-1]
			if ecmascript.StringTrim(d.file.Text()[c.End():cursor]) != "" {
				break
			}
			cursor = c.Pos()
			first--
		}
		if first == end {
			continue
		}
		comments := d.comments[first:end]
		var dep *Deprecation
		seenStyles := make(map[string]bool, len(d.styles))
		for _, style := range d.styles {
			if seenStyles[style] {
				continue
			}
			seenStyles[style] = true
			switch style {
			case "jsdoc":
				for _, c := range comments {
					if c.Kind == ast.KindMultiLineCommentTrivia {
						dep, _ = parseExportDoc(d.commentText(c))
					}
				}
			case "tomdoc":
				var lines []string
				for _, c := range comments {
					line := ecmascript.StringTrim(d.commentText(c))
					if line == "" {
						break
					}
					lines = append(lines, line)
				}
				text := strings.Join(lines, " ")
				for _, status := range []string{"Public:", "Internal:", "Deprecated:"} {
					if !strings.HasPrefix(text, status) {
						continue
					}
					description := strings.TrimLeftFunc(text[len(status):], ecmascript.IsWhiteSpaceOrLineTerminator)
					if end := strings.IndexAny(description, "\r\n\u2028\u2029"); end >= 0 {
						description = description[:end]
					}
					if description != "" {
						dep = nil
						if status == "Deprecated:" {
							dep = &Deprecation{Description: description}
						}
					}
				}
			}
		}
		return dep
	}
	return nil
}

func sortCommentEnd(comments []*ast.CommentRange, pos int) int {
	// Binary search avoids rescanning the file's preceding comments per export.
	lo, hi := 0, len(comments)
	for lo < hi {
		mid := (lo + hi) / 2
		if comments[mid].End() <= pos {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

func (d *exportDocumentation) commentText(c *ast.CommentRange) string {
	text := d.file.Text()[c.Pos():c.End()]
	if c.Kind == ast.KindMultiLineCommentTrivia {
		return text[2 : len(text)-2]
	}
	return text[2:]
}

func (d *exportDocumentation) moduleDeprecation() *Deprecation {
	// Module docs are always JSDoc, regardless of import/docstyle. The first
	// block containing @module wins, even if it has no @deprecated tag.
	for _, c := range d.comments {
		if c.Kind != ast.KindMultiLineCommentTrivia {
			continue
		}
		dep, module := parseExportDoc(d.commentText(c))
		if module {
			return dep
		}
	}
	return nil
}

func captureExportDocSources(file *ast.SourceFile, step *exportStep, stmt *ast.Node) {
	if step.Kind != exportStepLocalDefault && (step.Kind != exportStepNames || !isExportedDeclaration(stmt)) {
		return
	}
	step.Docs = make(map[string]*exportDocSource)
	doc := func(nodes ...*ast.Node) *exportDocSource { return &exportDocSource{file: file, nodes: nodes} }
	if step.Kind == exportStepLocalDefault {
		step.Docs[defaultExportName] = doc(stmt)
		return
	}
	if stmt.Kind == ast.KindVariableStatement && !ast.HasSyntacticModifier(stmt, ast.ModifierFlagsDefault) {
		for _, decl := range stmt.AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes {
			dep := doc(decl, stmt)
			rslint_utils.CollectBindingNames(decl.Name(), func(_ *ast.Node, name string) { step.Docs[name] = dep })
		}
		return
	}
	for _, name := range step.Names {
		step.Docs[name] = doc(stmt)
	}
}

// Only the @module and @deprecated tags affect this rule. Use ECMAScript
// whitespace for Doctrine's unwrap and description operations. Unlike Doctrine,
// malformed unrelated type tags do not terminate this metadata extraction.
func parseExportDoc(text string) (*Deprecation, bool) {
	type tag struct {
		title      string
		start, end int
	}
	var tags []tag
	var unwrapped strings.Builder
	for len(text) > 0 {
		line, separator := text, ""
		if end := strings.IndexAny(text, "\r\n\u2028\u2029"); end >= 0 {
			_, size := utf8.DecodeRuneInString(text[end:])
			if text[end] == '\r' && end+1 < len(text) && text[end+1] == '\n' {
				size++
			}
			line, separator, text = text[:end], text[end:end+size], text[end+size:]
		} else {
			text = ""
		}
		line = strings.TrimLeftFunc(line, ecmascript.IsWhiteSpaceOrLineTerminator)
		if strings.HasPrefix(line, "*") {
			line = line[1:]
			if r, size := utf8.DecodeRuneInString(line); size > 0 && ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				line = line[size:]
			}
		}
		content := strings.TrimLeftFunc(line, ecmascript.IsWhiteSpaceOrLineTerminator)
		start := unwrapped.Len() + len(line) - len(content)
		if strings.HasPrefix(content, "@") {
			end := 1
			for end < len(content) && ((content[end] >= 'a' && content[end] <= 'z') || (content[end] >= 'A' && content[end] <= 'Z') || (content[end] >= '0' && content[end] <= '9')) {
				end++
			}
			tags = append(tags, tag{title: content[1:end], start: start, end: start + end})
		}
		unwrapped.WriteString(line)
		unwrapped.WriteString(separator)
	}
	text = unwrapped.String()
	var dep *Deprecation
	module := false
	for i, current := range tags {
		if current.title == "module" {
			module = true
		}
		if current.title != "deprecated" || dep != nil {
			continue
		}
		end := len(text)
		if i+1 < len(tags) {
			end = tags[i+1].start
		}
		description := ecmascript.StringTrim(text[current.end:end])
		if len(description) > 1 && description[0] == '-' {
			r, size := utf8.DecodeRuneInString(description[1:])
			if ecmascript.IsWhiteSpaceOrLineTerminator(r) {
				description = description[1+size:]
			}
		}
		dep = &Deprecation{Description: description}
	}
	return dep, module
}
