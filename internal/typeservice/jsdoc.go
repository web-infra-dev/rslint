package typeservice

import (
	"cmp"
	"slices"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
)

func documentation(c *checker.Checker, declarations []*ast.Node, active map[*ast.Symbol]bool) DisplayParts {
	declarations = uniqueDeclarations(declarations)
	var parts DisplayParts
	var comments []DisplayParts
	for _, decl := range declarations {
		nodes := jsDocCommentNodes(decl)
		if decl.Kind == ast.KindJSDocParameterTag || decl.Kind == ast.KindJSDocPropertyTag {
			nodes = []*ast.Node{decl}
		}
		if decl.Kind == ast.KindJSDocTypedefTag || decl.Kind == ast.KindJSDocCallbackTag {
			nodes = []*ast.Node{decl, decl.Parent}
		}
		if decl.Kind == ast.KindJSDocSignature && decl.Parent.Kind == ast.KindJSDocOverloadTag {
			nodes = []*ast.Node{decl.Parent.Parent}
		}
		for _, node := range nodes {
			var inheritDoc *ast.Node
			if node.Kind == ast.KindJSDoc {
				tags := docTags(node)
				if decl.Kind != ast.KindJSDocCallbackTag && decl.Kind != ast.KindJSDocTypedefTag && belongsToTypeAlias(tags) {
					continue
				}
				if index := slices.IndexFunc(tags, isInheritDoc); index >= 0 {
					inheritDoc = tags[index]
				}
			}
			if node.CommentList() == nil && inheritDoc == nil {
				continue
			}
			comment := commentParts(c, node)
			if inheritDoc != nil {
				comment = append(comment, commentParts(c, inheritDoc)...)
			}
			if !slices.ContainsFunc(comments, func(existing DisplayParts) bool {
				return slices.EqualFunc(existing, comment, func(a, b DisplayPart) bool {
					return a.Text == b.Text && a.Kind == b.Kind
				})
			}) {
				comments = append(comments, comment)
			}
		}
	}
	for i, comment := range comments {
		if i > 0 {
			parts = append(parts, DisplayPart{Kind: "lineBreak", Text: "\n"})
		}
		parts = append(parts, comment...)
	}
	if c != nil && (len(parts) == 0 || hasInheritDoc(declarations)) {
		for _, inherited := range inheritedDeclarations(c, declarations, active, false, documentation) {
			if len(parts) != 0 {
				inherited = append(inherited, DisplayPart{Kind: "lineBreak", Text: "\n"})
			}
			parts = append(inherited, parts...)
		}
	}
	return parts
}

func jsDocTags(c *checker.Checker, declarations []*ast.Node, active map[*ast.Symbol]bool) JSDocTags {
	declarations = uniqueDeclarations(declarations)
	var tags JSDocTags
	for _, decl := range declarations {
		nodes := declarationTags(decl)
		if !belongsToTypeAlias(nodes) {
			tags = appendTags(c, tags, nodes)
		}
	}
	if c != nil && (len(tags) == 0 || hasInheritDoc(declarations)) {
		for _, inherited := range inheritedDeclarations(c, declarations, active, true, jsDocTags) {
			tags = append(inherited, tags...)
		}
	}
	return tags
}

func uniqueDeclarations(declarations []*ast.Node) []*ast.Node {
	if len(declarations) == 1 && declarations[0] == nil {
		return nil
	}
	if len(declarations) < 2 {
		return declarations
	}
	unique := make([]*ast.Node, 0, len(declarations))
	seen := make(map[*ast.Node]bool, len(declarations))
	for _, declaration := range declarations {
		if declaration != nil && !seen[declaration] {
			seen[declaration] = true
			unique = append(unique, declaration)
		}
	}
	return unique
}

// The parser attaches comments to their syntax owners. A variable declaration
// and its initializer may share the variable statement's comment. Earlier
// blocks on one owner contribute only JavaScript overload tags.
func jsDocCommentNodes(decl *ast.Node) []*ast.Node {
	var nodes []*ast.Node
	collect := func(docs []*ast.Node) {
		for i, doc := range docs {
			tags := docTags(doc)
			if i != len(docs)-1 {
				for _, tag := range tags {
					if tag.Kind == ast.KindJSDocOverloadTag {
						nodes = append(nodes, tag)
					}
				}
			} else if doc.Parent != nil && doc.Parent.Kind == ast.KindParenthesizedExpression && doc.Parent != decl &&
				slices.ContainsFunc(tags, isJSDocTypeAssertion) {
				for _, tag := range tags {
					if !isJSDocTypeAssertion(tag) {
						nodes = append(nodes, tag)
					}
				}
			} else {
				nodes = append(nodes, doc)
			}
		}
	}
	if ast.IsVariableLike(decl) {
		if initializer := decl.Initializer(); initializer != nil {
			collect(initializer.JSDoc(nil))
		}
	}
	for node := decl; node != nil && node.Parent != nil; node = ast.GetNextJSDocCommentLocation(node) {
		collect(node.JSDoc(nil))
		if node.Kind == ast.KindParameter || node.Kind == ast.KindTypeParameter {
			nodes = append(nodes, parameterTags(node)...)
			break
		}
	}
	return nodes
}

func isJSDocTypeAssertion(tag *ast.Node) bool {
	return tag.Kind == ast.KindJSDocTypeTag || tag.Kind == ast.KindJSDocSatisfiesTag
}

func parameterTags(parameter *ast.Node) []*ast.Node {
	var matches []*ast.Node
	name := parameter.Name()
	tags := declarationTags(parameter.Parent)
	if parameter.Kind == ast.KindTypeParameter {
		for _, tag := range tags {
			if tag.Kind == ast.KindJSDocTemplateTag && slices.ContainsFunc(tag.TypeParameters(), func(decl *ast.Node) bool {
				return decl.Name().Text() == name.Text()
			}) {
				matches = append(matches, tag)
			}
		}
		return matches
	}
	if !ast.IsIdentifier(name) {
		position := slices.Index(parameter.Parent.Parameters(), parameter)
		for _, tag := range tags {
			if tag.Kind == ast.KindJSDocParameterTag {
				if position == 0 {
					return []*ast.Node{tag}
				}
				position--
			}
		}
		return nil
	}
	for _, tag := range tags {
		if tag.Kind == ast.KindJSDocParameterTag && ast.IsIdentifier(tag.Name()) && tag.Name().Text() == name.Text() {
			matches = append(matches, tag)
		}
	}
	return matches
}

func declarationTags(decl *ast.Node) []*ast.Node {
	var tags []*ast.Node
	for _, node := range jsDocCommentNodes(decl) {
		if node.Kind == ast.KindJSDoc {
			tags = append(tags, docTags(node)...)
		} else {
			tags = append(tags, node)
		}
	}
	return tags
}

func docTags(node *ast.Node) []*ast.Node {
	if list := node.AsJSDoc().Tags; list != nil {
		return list.Nodes
	}
	return nil
}

// tsgo binds reparsed declarations and encodes their authored JSDoc separately.
// Keep their indices in the existing node table. Only a queried JS file needs them.
type jsDocIndices struct {
	source, bound []uint32
}

var jsDocIndicesKey = ast.NewSourceFileDataKey[*jsDocIndices]()

func fileJSDocIndices(file *ast.SourceFile, table *encoder.NodeIndexTable) *jsDocIndices {
	return ast.GetOrComputeSourceFileData(file, jsDocIndicesKey, func(*ast.SourceFile) *jsDocIndices {
		var indices jsDocIndices
		inDoc := false
		for i, n := range table.Nodes {
			if n == nil {
				continue
			}
			if n.Kind == ast.KindJSDoc {
				inDoc = n.Flags&ast.NodeFlagsReparsed == 0
			} else if n.Flags&ast.NodeFlagsJSDoc == 0 {
				inDoc = false
			}
			if inDoc {
				indices.source = append(indices.source, uint32(i))
			} else if n.Flags&ast.NodeFlagsReparsed != 0 {
				indices.bound = append(indices.bound, uint32(i))
			}
		}
		order := func(i, j uint32) int {
			return cmp.Compare(jsDocLookupPos(table.Nodes[i]), jsDocLookupPos(table.Nodes[j]))
		}
		slices.SortStableFunc(indices.source, order)
		slices.SortStableFunc(indices.bound, order)
		return &indices
	})
}

func jsDocLookupPos(node *ast.Node) int {
	if node.Kind == ast.KindJSDocSignature && node.Parent.Kind == ast.KindJSDocOverloadTag {
		return node.Parent.TagName().Pos()
	}
	return node.Pos()
}

func jsDocAtPosition(indices []uint32, table *encoder.NodeIndexTable, pos int) []uint32 {
	end := sort.Search(len(indices), func(i int) bool { return jsDocLookupPos(table.Nodes[indices[i]]) > pos })
	start := end
	for start > 0 && jsDocLookupPos(table.Nodes[indices[start-1]]) == pos {
		start--
	}
	return indices[start:end]
}

func jsDocNodeIndex(node *ast.Node, file *ast.SourceFile, table *encoder.NodeIndexTable) uint32 {
	indices := fileJSDocIndices(file, table)
	candidates := jsDocAtPosition(indices.source, table, jsDocLookupPos(node))
	var match uint32
	// The encoder emits shared comments on reparsed declarations before their source owner.
	for i := len(candidates) - 1; i >= 0; i-- {
		idx := candidates[i]
		source := table.Nodes[idx]
		if source == node || node.Flags&ast.NodeFlagsReparsed != 0 && isSourceJSDocNode(source, node) {
			return idx
		}
		// Descendants of a deep clone keep their source flags.
		if match == 0 && source.Kind == node.Kind && source.Loc == node.Loc {
			match = idx
		}
	}
	return match
}

func boundJSDocDeclaration(source *ast.Node) *ast.Node {
	if source.Flags&ast.NodeFlagsJavaScriptFile != 0 {
		file := ast.GetSourceFileOfNode(source)
		table := encoder.GetNodeIndexTable(file)
		indices := fileJSDocIndices(file, table)
		for _, idx := range jsDocAtPosition(indices.bound, table, jsDocLookupPos(source)) {
			if node := table.Nodes[idx]; isSourceJSDocNode(source, node) {
				return node
			}
		}
	}
	return nil
}

func isSourceJSDocNode(source, node *ast.Node) bool {
	if source.Kind == ast.KindJSDocSignature && source.Parent.Kind == ast.KindJSDocOverloadTag {
		// The native overload declaration uses just the @overload tag name's range.
		if ast.IsFunctionLike(node) && node.Loc == source.Parent.TagName().Loc {
			return true
		}
	}
	if source.Loc != node.Loc {
		return false
	}
	if source.Kind == node.Kind {
		return true
	}
	switch source.Kind {
	case ast.KindJSDocTypedefTag, ast.KindJSDocCallbackTag:
		return node.Kind == ast.KindJSTypeAliasDeclaration
	case ast.KindJSDocImportTag:
		return node.Kind == ast.KindJSImportDeclaration
	case ast.KindJSDocPropertyTag, ast.KindJSDocParameterTag:
		return node.Kind == ast.KindPropertySignature || node.Kind == ast.KindParameter
	case ast.KindJSDocSignature:
		return node.Kind == ast.KindFunctionType
	case ast.KindJSDocTypeLiteral:
		return node.Kind == ast.KindTypeLiteral || node.Kind == ast.KindArrayType
	}
	return false
}

func belongsToTypeAlias(tags []*ast.Node) bool {
	return slices.ContainsFunc(tags, func(tag *ast.Node) bool {
		return tag.Kind == ast.KindJSDocTypedefTag || tag.Kind == ast.KindJSDocCallbackTag
	}) && !slices.ContainsFunc(tags, func(tag *ast.Node) bool {
		return tag.Kind == ast.KindJSDocParameterTag || tag.Kind == ast.KindJSDocReturnTag
	})
}

func isInheritDoc(tag *ast.Node) bool {
	name := tag.TagName().Text()
	// cspell:ignore inheritdoc
	return name == "inheritDoc" || name == "inheritdoc"
}

func hasInheritDoc(declarations []*ast.Node) bool {
	return slices.ContainsFunc(declarations, func(decl *ast.Node) bool {
		return slices.ContainsFunc(declarationTags(decl), isInheritDoc)
	})
}

func inheritedDeclarations[T ~[]E, E any](c *checker.Checker, declarations []*ast.Node, active map[*ast.Symbol]bool, tags bool, read func(*checker.Checker, []*ast.Node, map[*ast.Symbol]bool) T) []T {
	var result []T
	seen := make(map[*ast.Symbol]bool)
	for _, declaration := range declarations {
		owner, member := declaration.Parent, declaration.Symbol()
		if owner == nil || member == nil {
			continue
		}
		if owner.Kind == ast.KindConstructor {
			owner = owner.Parent
		}
		if owner == nil || !ast.IsClassOrInterfaceLike(owner) {
			continue
		}
		bases := ast.GetExtendsHeritageClauseElements(owner)
		if ast.IsClassLike(owner) {
			bases = append(slices.Clone(bases), ast.GetImplementsHeritageClauseElements(owner)...)
		}
		accessor := declaration.Kind == ast.KindGetAccessor || declaration.Kind == ast.KindSetAccessor
		for _, base := range bases {
			t := c.GetTypeAtLocation(base)
			if t == nil {
				continue
			}
			if ast.HasStaticModifier(declaration) && t.Symbol() != nil {
				t = c.GetTypeOfSymbol(t.Symbol())
			}
			symbol := c.GetPropertyOfType(t, member.Name)
			if symbol == nil || seen[symbol] {
				continue
			}
			seen[symbol] = true
			if !accessor && tags && len(symbol.Declarations) != 1 {
				continue
			}
			if active[symbol] {
				break
			}
			active[symbol] = true
			var parts T
			if accessor {
				matching := slices.DeleteFunc(slices.Clone(symbol.Declarations), func(decl *ast.Node) bool {
					return decl.Kind != declaration.Kind
				})
				parts = read(c, matching, active)
			}
			if len(parts) == 0 {
				parts = read(c, symbol.Declarations, active)
			}
			result = append(result, parts)
			delete(active, symbol)
			break
		}
	}
	return result
}

func appendTags(c *checker.Checker, result JSDocTags, tags []*ast.Node) JSDocTags {
	for _, tag := range tags {
		result = append(result, JSDocTag{Name: tag.TagName().Text(), Text: tagParts(c, tag)})
		if (tag.Kind == ast.KindJSDocParameterTag || tag.Kind == ast.KindJSDocPropertyTag) && tag.AsJSDocParameterOrPropertyTag().IsNameFirst {
			expression := tag.AsJSDocParameterOrPropertyTag().TypeExpression
			if expression != nil {
				if typ := expression.Type(); typ != nil && typ.Kind == ast.KindJSDocTypeLiteral {
					result = appendTags(c, result, typ.AsJSDocTypeLiteral().JSDocPropertyTags)
				}
			}
		}
	}
	return result
}

func tagParts(c *checker.Checker, tag *ast.Node) DisplayParts {
	var node *ast.Node
	kind := "text"
	switch tag.Kind {
	case ast.KindJSDocThrowsTag:
		node = tag.AsJSDocThrowsTag().TypeExpression
	case ast.KindJSDocImplementsTag:
		node = tag.AsJSDocImplementsTag().ClassName
	case ast.KindJSDocAugmentsTag:
		node = tag.AsJSDocAugmentsTag().ClassName
	case ast.KindJSDocTypeTag:
		node = tag.AsJSDocTypeTag().TypeExpression
	case ast.KindJSDocSatisfiesTag:
		node = tag.AsJSDocSatisfiesTag().TypeExpression
	case ast.KindJSDocSeeTag:
		node = tag.AsJSDocSeeTag().NameExpression
	case ast.KindJSDocTypedefTag, ast.KindJSDocCallbackTag:
		node, kind = tag.Name(), "aliasName"
	case ast.KindJSDocParameterTag:
		node, kind = tag.Name(), "parameterName"
	case ast.KindJSDocPropertyTag:
		node, kind = tag.Name(), "propertyName"
	case ast.KindJSDocTemplateTag:
		var parts DisplayParts
		if constraint := tag.AsJSDocTemplateTag().Constraint; constraint != nil {
			parts = append(parts, DisplayPart{Kind: "text", Text: scanner.GetTextOfNode(constraint)})
		}
		for i, parameter := range tag.TypeParameters() {
			if i > 0 {
				parts = append(parts, DisplayPart{Kind: "punctuation", Text: ","})
			}
			if len(parts) != 0 {
				parts = append(parts, DisplayPart{Kind: "space", Text: " "})
			}
			parts = append(parts, DisplayPart{Kind: "typeParameterName", Text: scanner.GetTextOfNode(parameter)})
		}
		if comment := commentParts(c, tag); len(comment) != 0 {
			parts = append(parts, DisplayPart{Kind: "space", Text: " "})
			parts = append(parts, comment...)
		}
		return parts
	}
	comment := commentParts(c, tag)
	if node == nil {
		return comment
	}
	name := scanner.GetTextOfNode(node)
	if len(comment) == 0 {
		return DisplayParts{{Kind: "text", Text: name}}
	}
	if name == "http" || name == "https" {
		return append(DisplayParts{{Kind: "text", Text: name}}, comment...)
	}
	return append(DisplayParts{{Kind: kind, Text: name}, {Kind: "space", Text: " "}}, comment...)
}

func commentParts(c *checker.Checker, node *ast.Node) DisplayParts {
	var parts DisplayParts
	for _, comment := range node.Comments() {
		switch comment.Kind {
		case ast.KindJSDocText:
			parts = append(parts, DisplayPart{Kind: "text", Text: comment.Text()})
		case ast.KindJSDocLink, ast.KindJSDocLinkCode, ast.KindJSDocLinkPlain:
			parts = append(parts, linkParts(c, comment)...)
		}
	}
	return parts
}

func linkParts(c *checker.Checker, link *ast.Node) DisplayParts {
	kind := "link"
	switch link.Kind {
	case ast.KindJSDocLinkCode:
		kind = "linkcode"
	case ast.KindJSDocLinkPlain:
		kind = "linkplain"
	}
	parts := DisplayParts{{Kind: "link", Text: "{@" + kind + " "}}
	name, text := link.Name(), link.Text()
	if name == nil {
		if text != "" {
			parts = append(parts, DisplayPart{Kind: "linkText", Text: text})
		}
	} else {
		suffix := linkNameSuffix(text)
		nameText, label := scanner.GetTextOfNode(name)+text[:suffix], text[suffix:]
		if rest, found := strings.CutPrefix(label, "|"); found {
			label = strings.TrimLeft(rest, " ")
		}
		if target := linkTarget(c, name); target != nil {
			parts = append(parts, DisplayPart{Kind: "linkName", Text: nameText, Target: target})
			if label != "" {
				parts = append(parts, DisplayPart{Kind: "linkText", Text: label})
			}
		} else {
			if suffix == 0 || strings.HasPrefix(text[suffix:], "|") && !strings.HasSuffix(nameText, " ") {
				nameText += " "
			}
			parts = append(parts, DisplayPart{Kind: "linkText", Text: nameText + label})
		}
	}
	return append(parts, DisplayPart{Kind: "link", Text: "}"})
}

func linkNameSuffix(text string) int {
	switch {
	case strings.HasPrefix(text, "://"):
		if bar := strings.IndexByte(text, '|'); bar >= 0 {
			return bar
		}
		return len(text)
	case strings.HasPrefix(text, "()"):
		return 2
	case strings.HasPrefix(text, "<"):
		depth := 0
		for i := range len(text) {
			switch text[i] {
			case '<':
				depth++
			case '>':
				depth--
				if depth == 0 {
					return i + 1
				}
			}
		}
	}
	return 0
}

func linkTarget(c *checker.Checker, name *ast.Node) *DisplayPartTarget {
	if c == nil {
		return nil
	}
	symbol := c.GetSymbolAtLocation(name)
	if symbol == nil {
		return nil
	}
	symbol = c.SkipAlias(symbol)
	declaration := symbol.ValueDeclaration
	if declaration == nil && len(symbol.Declarations) != 0 {
		declaration = symbol.Declarations[0]
	}
	if declaration == nil {
		return nil
	}
	file := ast.GetSourceFileOfNode(declaration)
	if file == nil {
		return nil
	}
	start, end := scanner.GetTokenPosOfNode(declaration, file, false), declaration.End()
	text := file.Text()
	return &DisplayPartTarget{File: file.FileName(), Start: uint32(core.UTF16Len(text[:start])), Length: uint32(core.UTF16Len(text[start:end]))}
}
