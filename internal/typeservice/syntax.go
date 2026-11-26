package typeservice

import (
	"errors"
	"slices"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/printer"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/web-infra-dev/rslint/internal/program"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// Only the requested declaration's ancestors and declaration names are copied.
// Descendant syntax stays demand-loaded through the immutable source transport.
func nodeFacts(file *ast.SourceFile, node *ast.Node) NodeFacts {
	facts := NodeFacts{}
	seen := map[*ast.Node]bool{}
	table := encoder.GetNodeIndexTable(file)
	reference := func(n *ast.Node) *NodeRef { return indexedNodeReference(file, table, n) }
	add := func(n *ast.Node) *ast.Node {
		if n == nil {
			return nil
		}
		ref := reference(n)
		if int(ref.Index) < len(table.Nodes) {
			n = table.Nodes[ref.Index]
		}
		if seen[n] {
			return n
		}
		seen[n] = true
		fact := NodeFact{
			Node:   *ref,
			Kind:   strings.TrimPrefix(n.Kind.String(), "Kind"),
			Flags:  uint32(n.Flags),
			Parent: reference(syntaxParent(n)),
		}
		if n.Flags&ast.NodeFlagsPossiblyContainsDeprecatedTag != 0 && ast.GetJSDocDeprecatedTag(n) == nil {
			fact.Flags &^= uint32(ast.NodeFlagsPossiblyContainsDeprecatedTag)
		}
		if ast.IsModuleDeclaration(n) {
			fact.ModuleKeyword = uint32(n.AsModuleDeclaration().Keyword)
			if n.Parent != nil && ast.IsModuleDeclaration(n.Parent) {
				fact.Flags |= uint32(ast.NodeFlagsNestedNamespace)
			}
		}
		if ast.IsIdentifier(n) || ast.IsPrivateIdentifier(n) || ast.IsStringLiteral(n) || ast.IsNumericLiteral(n) {
			text := n.Text()
			fact.Text = &text
		}
		fact.Name = reference(syntaxName(n))
		facts = append(facts, fact)
		return n
	}
	for current := node; current != nil; current = syntaxParent(current) {
		current = add(current)
		add(syntaxName(current))
	}
	return facts
}

func syntaxParent(node *ast.Node) *ast.Node {
	parent := node.Parent
	for parent != nil && (parent.Kind == ast.KindAsExpression || parent.Kind == ast.KindSatisfiesExpression) && utils.JSDocTypeCastExpression(parent) != nil {
		parent = parent.Parent
	}
	return parent
}

func syntaxName(node *ast.Node) *ast.Node {
	if node.Kind == ast.KindJSDocSeeTag {
		return node.AsJSDocSeeTag().NameExpression
	}
	name := node.Name()
	if node.Kind == ast.KindJSDocTypedefTag || node.Kind == ast.KindJSDocCallbackTag {
		for name != nil && ast.IsModuleDeclaration(name) {
			if name.AsModuleDeclaration().Body == nil {
				return name.Name()
			}
			name = name.AsModuleDeclaration().Body
		}
	}
	return name
}

func nodeReference(p *program.Program, n *ast.Node) *NodeRef {
	if n == nil {
		return nil
	}
	f := ast.GetSourceFileOfNode(n)
	if f == nil || !p.OwnsSourceFile(f) {
		panic("declaration is outside Program")
	}
	return indexedNodeReference(f, encoder.GetNodeIndexTable(f), n)
}

func indexedNodeReference(f *ast.SourceFile, table *encoder.NodeIndexTable, n *ast.Node) *NodeRef {
	if n == nil {
		return nil
	}
	index := uint32(0)
	if n.Flags&ast.NodeFlagsJavaScriptFile != 0 && n.Flags&(ast.NodeFlagsJSDoc|ast.NodeFlagsReparsed) != 0 {
		index = jsDocNodeIndex(n, f, table)
	}
	if index == 0 {
		index = table.GetIndex(n)
	}
	if index == 0 {
		panic("declaration has no compiler node handle")
	}
	return &NodeRef{File: f.FileName(), Index: index}
}

func resolveNode(f *ast.SourceFile, node NodeRef) (*ast.Node, error) {
	if f == nil {
		return nil, errors.New("foreign source handle")
	}
	table := encoder.GetNodeIndexTable(f)
	if node.Index >= uint32(len(table.Nodes)) || table.Nodes[node.Index] == nil {
		return nil, errors.New("invalid compiler node handle")
	}
	parent := table.Nodes[node.Index]
	if node.Token == nil {
		return parent, nil
	}
	token := node.Token
	positions := f.GetPositionMap()
	pos, end := positions.UTF16ToUTF8(token.Pos), positions.UTF16ToUTF8(token.End)
	if pos < parent.Pos() || end < pos || end > parent.End() {
		return nil, errors.New("invalid compiler token range")
	}
	if token.Kind == ast.KindSyntaxList {
		list := ast.NewNodeFactory(ast.NodeFactoryHooks{}).NewSyntaxList(nil)
		list.Loc, list.Parent = core.NewTextRange(pos, end), parent
		return list, nil
	}
	if !ast.IsTokenKind(token.Kind) {
		return nil, errors.New("invalid compiler token kind")
	}
	tokenStart, flags, ok := scanValidatedToken(f, token.Kind, pos, end)
	if !ok {
		return nil, errors.New("compiler token does not match source")
	}
	return createEphemeralToken(f, token.Kind, pos, tokenStart, end, parent, flags), nil
}

// scanValidatedToken verifies that a client-supplied token kind and range match
// the actual source text without writing into the shared SourceFile tokenCache.
func scanValidatedToken(f *ast.SourceFile, kind ast.Kind, pos, end int) (int, ast.TokenFlags, bool) {
	s := scanner.GetScannerForSourceFile(f, pos)
	scanned := s.Token()
	if (scanned == kind || (kind == ast.KindIdentifier && ast.IsKeywordKind(scanned))) && s.TokenEnd() == end {
		return s.TokenStart(), s.TokenFlags(), true
	}
	switch kind {
	case ast.KindGreaterThanToken, ast.KindGreaterThanGreaterThanToken, ast.KindGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken, ast.KindGreaterThanGreaterThanGreaterThanToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken:
		if s.ReScanGreaterThanToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindLessThanToken:
		if s.ReScanLessThanToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindQuestionToken:
		if s.ReScanQuestionToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindRegularExpressionLiteral:
		if s.ReScanSlashToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindTemplateMiddle, ast.KindTemplateTail:
		if s.ReScanTemplateToken(false) == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces, ast.KindLessThanSlashToken:
		if s.ReScanJsxToken(true) == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindHashToken:
		if s.ReScanHashToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindAsteriskEqualsToken:
		if s.ReScanAsteriskEqualsToken() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	case ast.KindStringLiteral:
		if s.ReScanJsxAttributeValue() == kind && s.TokenEnd() == end {
			return s.TokenStart(), s.TokenFlags(), true
		}
	}
	return 0, ast.TokenFlagsNone, false
}

// createEphemeralToken builds a request-local token node without mutating
// SourceFile.tokenCache.
func createEphemeralToken(f *ast.SourceFile, kind ast.Kind, pos, tokenStart, end int, parent *ast.Node, flags ast.TokenFlags) *ast.Node {
	factory := ast.NewNodeFactory(ast.NodeFactoryHooks{})
	text := f.Text()[tokenStart:end]
	var tok *ast.Node
	switch kind {
	case ast.KindNumericLiteral:
		tok = factory.NewNumericLiteral(text, flags)
	case ast.KindBigIntLiteral:
		tok = factory.NewBigIntLiteral(text, flags)
	case ast.KindStringLiteral:
		tok = factory.NewStringLiteral(text, flags)
	case ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces:
		tok = factory.NewJsxText(text, kind == ast.KindJsxTextAllWhiteSpaces)
	case ast.KindRegularExpressionLiteral:
		tok = factory.NewRegularExpressionLiteral(text, flags)
	case ast.KindNoSubstitutionTemplateLiteral:
		tok = factory.NewNoSubstitutionTemplateLiteral(text, flags)
	case ast.KindTemplateHead:
		tok = factory.NewTemplateHead(text, "", flags)
	case ast.KindTemplateMiddle:
		tok = factory.NewTemplateMiddle(text, "", flags)
	case ast.KindTemplateTail:
		tok = factory.NewTemplateTail(text, "", flags)
	case ast.KindIdentifier:
		tok = factory.NewIdentifier(text)
	case ast.KindPrivateIdentifier:
		tok = factory.NewPrivateIdentifier(text)
	default:
		tok = factory.NewToken(kind)
	}
	tok.Loc = core.NewTextRange(pos, end)
	tok.Parent = parent
	return tok
}

func encodeGeneratedNode(node *ast.Node, builder *checker.NodeBuilder, enclosing *ast.Node) (Result, error) {
	if node == nil {
		return &EncodedTypeNode{}, nil
	}
	tree, table, err := encoder.EncodeNode(node, nil)
	if err != nil {
		return nil, err
	}
	result := &EncodedTypeNode{tree: tree}
	emit := builder.EmitContext()
	for index, generated := range table.Nodes {
		if flags := emit.EmitFlags(generated); flags != 0 {
			if result.emitFlags == nil {
				result.emitFlags = make([]uint32, len(table.Nodes))
			}
			if flags&printer.EFSingleLine != 0 {
				result.emitFlags[index] |= 1
			}
			if flags&printer.EFNoIndentation != 0 {
				result.emitFlags[index] |= 2
			}
			if flags&printer.EFNoAsciiEscaping != 0 {
				result.emitFlags[index] |= 4
			}
		}
		origin := emit.ParseNode(generated)
		if origin != nil && ast.GetSourceFileOfNode(origin) != nil {
			if result.parseNodes == nil {
				result.parseNodes = make([]*ast.Node, len(table.Nodes))
			}
			result.parseNodes[index] = origin
		}
	}
	prepareGeneratedSyntax(tree, table.Nodes, result.parseNodes, ast.GetSourceFileOfNode(enclosing))
	return result, nil
}

func isJSDocOverloadDeclaration(decl *ast.Node) bool {
	return decl.Flags&ast.NodeFlagsReparsed != 0 && (ast.IsFunctionDeclaration(decl) || ast.IsMethodDeclaration(decl) || ast.IsConstructorDeclaration(decl))
}

// Choose the authored value, or the first declaration for a type-only link.
func symbolTargetDeclaration(symbol *ast.Symbol) *ast.Node {
	value := symbol.ValueDeclaration
	if value != nil && !isJSDocOverloadDeclaration(value) {
		return value
	}
	afterValue := value == nil
	for _, decl := range symbol.Declarations {
		if afterValue && !isJSDocOverloadDeclaration(decl) {
			return decl
		}
		if decl == value {
			afterValue = true
		}
	}
	return value
}

func symbolDeclarations(symbol *ast.Symbol, snapshot bool) ([]*ast.Node, *ast.Node) {
	nodes := symbol.Declarations
	value := symbol.ValueDeclaration
	var kept []*ast.Node
	start := 0
	useImplementation := false
	// The JS parser inserts generated overloads immediately before their
	// implementation. Usually all authored declarations form a single suffix.
	for i, decl := range nodes {
		if isJSDocOverloadDeclaration(decl) {
			if start < i {
				kept = append(kept, nodes[start:i]...)
			}
			start = i + 1
			useImplementation = useImplementation || decl == value
		} else if useImplementation {
			value = decl
			useImplementation = false
		}
	}
	if kept != nil {
		return append(kept, nodes[start:]...), value
	}
	declarations := nodes[start:]
	if snapshot {
		declarations = slices.Clone(declarations)
	}
	return declarations, value
}
