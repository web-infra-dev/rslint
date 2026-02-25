package inspector

import (
	"github.com/microsoft/typescript-go/shim/ast"
)

// BuildShallowNodeInfo builds minimal NodeInfo (only kind, pos, end) for nested nodes
// Note: ID is not included in shallow info as it changes on each request
func (b *Builder) BuildShallowNodeInfo(node *ast.Node) *NodeInfo {
	if node == nil {
		return nil
	}
	info := &NodeInfo{
		Kind:     int(node.Kind),
		KindName: KindToString(node.Kind),
		Flags:    int(node.Flags),
	}
	// Get flag names
	if node.Flags != 0 {
		info.FlagNames = GetNodeFlagNames(node.Flags)
	}
	// Set position - always use actual position from the node
	info.Pos = node.Pos()
	info.End = node.End()
	// Check if node is from external file (like lib.d.ts)
	nodeFile := ast.GetSourceFileOfNode(node)
	if nodeFile != nil && nodeFile != b.sourceFile {
		// External file - set FileName for fetching from that file
		info.FileName = nodeFile.FileName()
	}
	// Get text for identifiers and literals
	if text := GetNodeText(node); text != "" {
		info.Text = text
	}
	return info
}

// BuildNodeInfo builds NodeInfo from an AST node
func (b *Builder) BuildNodeInfo(node *ast.Node) *NodeInfo {
	if node == nil {
		return nil
	}

	info := &NodeInfo{
		Id:        uint64(ast.GetNodeId(node)),
		Kind:      int(node.Kind),
		KindName:  KindToString(node.Kind),
		Pos:       node.Pos(),
		End:       node.End(),
		Flags:     int(node.Flags),
		FlagNames: GetNodeFlagNames(node.Flags),
	}

	// Get text for identifiers and literals
	if text := GetNodeText(node); text != "" {
		info.Text = text
	}

	// Parent - now fetchable using kind filter even if positions overlap
	if node.Parent != nil {
		info.Parent = b.BuildShallowNodeInfo(node.Parent)
	}

	// Common node properties (shallow)
	b.buildNodeChildren(node, info)

	// Build locals (symbols declared in this node's scope)
	b.buildLocals(node, info)

	return info
}

// buildLocals builds the locals (symbols declared in this node's scope)
func (b *Builder) buildLocals(node *ast.Node, info *NodeInfo) {
	// Safely get locals - may panic for some node types
	var locals ast.SymbolTable
	func() {
		defer func() { recover() }()
		locals = node.Locals()
	}()

	if len(locals) == 0 {
		return
	}

	info.Locals = make([]*SymbolInfo, 0, len(locals))
	for _, symbol := range locals {
		if symbolInfo := b.BuildSymbolInfo(symbol); symbolInfo != nil {
			info.Locals = append(info.Locals, symbolInfo)
		}
	}
}

// buildNodeList builds a slice of shallow NodeInfo from a NodeList and adds list metadata
func (b *Builder) buildNodeList(list *ast.NodeList, metaName string, info *NodeInfo) []*NodeInfo {
	if list == nil || len(list.Nodes) == 0 {
		return nil
	}
	result := make([]*NodeInfo, 0, len(list.Nodes))
	for _, node := range list.Nodes {
		result = append(result, b.BuildShallowNodeInfo(node))
	}
	AddListMeta(info, metaName, list)
	return result
}

// buildNodeChildren populates node-specific child properties
func (b *Builder) buildNodeChildren(node *ast.Node, info *NodeInfo) {
	// Name (for declarations) - safe call as Name() may panic on some node types
	SafeCall(func() {
		if name := node.Name(); name != nil {
			info.Name = b.BuildShallowNodeInfo(name.AsNode())
		}
	})

	// Expression - safe call as Expression() may panic on some node types
	SafeCall(func() {
		if expr := node.Expression(); expr != nil {
			info.Expression = b.BuildShallowNodeInfo(expr)
		}
	})

	// Body - safe call as Body() may panic on some node types
	SafeCall(func() {
		if body := node.Body(); body != nil {
			info.Body = b.BuildShallowNodeInfo(body)
		}
	})

	// Type annotation - safe call as Type() may panic on some node types
	SafeCall(func() {
		if typeNode := node.Type(); typeNode != nil {
			info.TypeNode = b.BuildShallowNodeInfo(typeNode)
		}
	})

	// Initializer - safe call as Initializer() may panic on some node types
	SafeCall(func() {
		if init := node.Initializer(); init != nil {
			info.Initializer = b.BuildShallowNodeInfo(init)
		}
	})

	// BinaryExpression specific
	if node.Kind == ast.KindBinaryExpression {
		if be := node.AsBinaryExpression(); be != nil {
			info.Left = b.BuildShallowNodeInfo(be.Left)
			info.Right = b.BuildShallowNodeInfo(be.Right)
			info.OperatorToken = b.BuildShallowNodeInfo(be.OperatorToken.AsNode())
		}
	}

	// PrefixUnaryExpression specific
	if node.Kind == ast.KindPrefixUnaryExpression {
		if pue := node.AsPrefixUnaryExpression(); pue != nil {
			info.Operand = b.BuildShallowNodeInfo(pue.Operand)
		}
	}

	// PostfixUnaryExpression specific
	if node.Kind == ast.KindPostfixUnaryExpression {
		if pue := node.AsPostfixUnaryExpression(); pue != nil {
			info.Operand = b.BuildShallowNodeInfo(pue.Operand)
		}
	}

	// ConditionalExpression specific
	if node.Kind == ast.KindConditionalExpression {
		if ce := node.AsConditionalExpression(); ce != nil {
			info.Condition = b.BuildShallowNodeInfo(ce.Condition)
			info.WhenTrue = b.BuildShallowNodeInfo(ce.WhenTrue)
			info.WhenFalse = b.BuildShallowNodeInfo(ce.WhenFalse)
		}
	}

	// IfStatement specific
	if node.Kind == ast.KindIfStatement {
		if is := node.AsIfStatement(); is != nil {
			info.Condition = b.BuildShallowNodeInfo(is.Expression)
			info.ThenStatement = b.BuildShallowNodeInfo(is.ThenStatement)
			if is.ElseStatement != nil {
				info.ElseStatement = b.BuildShallowNodeInfo(is.ElseStatement)
			}
		}
	}

	// ClassDeclaration specific
	if node.Kind == ast.KindClassDeclaration {
		if cd := node.AsClassDeclaration(); cd != nil {
			info.HeritageClauses = b.buildNodeList(cd.HeritageClauses, "HeritageClauses", info)
			info.Members = b.buildNodeList(cd.Members, "Members", info)
			info.TypeParameters = b.buildNodeList(cd.TypeParameters, "TypeParameters", info)
		}
	}

	// ClassExpression specific
	if node.Kind == ast.KindClassExpression {
		if ce := node.AsClassExpression(); ce != nil {
			info.HeritageClauses = b.buildNodeList(ce.HeritageClauses, "HeritageClauses", info)
			info.Members = b.buildNodeList(ce.Members, "Members", info)
			info.TypeParameters = b.buildNodeList(ce.TypeParameters, "TypeParameters", info)
		}
	}

	// InterfaceDeclaration specific
	if node.Kind == ast.KindInterfaceDeclaration {
		if id := node.AsInterfaceDeclaration(); id != nil {
			info.HeritageClauses = b.buildNodeList(id.HeritageClauses, "HeritageClauses", info)
			info.Members = b.buildNodeList(id.Members, "Members", info)
			info.TypeParameters = b.buildNodeList(id.TypeParameters, "TypeParameters", info)
		}
	}

	// TypeLiteral (object type) specific
	if node.Kind == ast.KindTypeLiteral {
		if tl := node.AsTypeLiteralNode(); tl != nil {
			info.Members = b.buildNodeList(tl.Members, "Members", info)
		}
	}

	// FunctionDeclaration, MethodDeclaration, ArrowFunction, etc.
	SafeCall(func() {
		if sig := node.FunctionLikeData(); sig != nil {
			info.Parameters = b.buildNodeList(sig.Parameters, "Parameters", info)
			info.TypeParameters = b.buildNodeList(sig.TypeParameters, "TypeParameters", info)
		}
	})

	// CallExpression specific
	if node.Kind == ast.KindCallExpression {
		if ce := node.AsCallExpression(); ce != nil {
			if ce.QuestionDotToken != nil {
				info.QuestionDotToken = b.BuildShallowNodeInfo(ce.QuestionDotToken.AsNode())
			}
			info.TypeArguments = b.buildNodeList(ce.TypeArguments, "TypeArguments", info)
			info.Arguments = b.buildNodeList(ce.Arguments, "Arguments", info)
		}
	}

	// NewExpression specific
	if node.Kind == ast.KindNewExpression {
		if ne := node.AsNewExpression(); ne != nil {
			info.TypeArguments = b.buildNodeList(ne.TypeArguments, "TypeArguments", info)
			info.Arguments = b.buildNodeList(ne.Arguments, "Arguments", info)
		}
	}

	// Block specific
	if node.Kind == ast.KindBlock {
		if block := node.AsBlock(); block != nil {
			info.Statements = b.buildNodeList(block.Statements, "Statements", info)
		}
	}

	// SourceFile specific
	if node.Kind == ast.KindSourceFile {
		if sf := node.AsSourceFile(); sf != nil {
			info.Statements = b.buildNodeList(sf.Statements, "Statements", info)
		}
	}

	// ObjectLiteralExpression specific
	if node.Kind == ast.KindObjectLiteralExpression {
		if ole := node.AsObjectLiteralExpression(); ole != nil {
			info.Properties = b.buildNodeList(ole.Properties, "Properties", info)
		}
	}

	// ArrayLiteralExpression specific
	if node.Kind == ast.KindArrayLiteralExpression {
		if ale := node.AsArrayLiteralExpression(); ale != nil {
			info.Elements = b.buildNodeList(ale.Elements, "Elements", info)
		}
	}

	// VariableStatement specific
	if node.Kind == ast.KindVariableStatement {
		if vs := node.AsVariableStatement(); vs != nil {
			if vs.DeclarationList != nil {
				info.DeclarationList = b.BuildShallowNodeInfo(vs.DeclarationList.AsNode())
			}
		}
	}

	// VariableDeclarationList specific
	if node.Kind == ast.KindVariableDeclarationList {
		if vdl := node.AsVariableDeclarationList(); vdl != nil {
			info.Declarations = b.buildNodeList(vdl.Declarations, "Declarations", info)
		}
	}

	// EnumDeclaration specific
	if node.Kind == ast.KindEnumDeclaration {
		if ed := node.AsEnumDeclaration(); ed != nil {
			info.Members = b.buildNodeList(ed.Members, "Members", info)
		}
	}

	// TypeAliasDeclaration specific
	if node.Kind == ast.KindTypeAliasDeclaration {
		if tad := node.AsTypeAliasDeclaration(); tad != nil {
			info.TypeParameters = b.buildNodeList(tad.TypeParameters, "TypeParameters", info)
		}
	}

	// ModuleDeclaration specific
	if node.Kind == ast.KindModuleDeclaration {
		if md := node.AsModuleDeclaration(); md != nil {
			if md.Body != nil {
				info.Body = b.BuildShallowNodeInfo(md.Body)
			}
		}
	}

	// ModuleBlock specific
	if node.Kind == ast.KindModuleBlock {
		if mb := node.AsModuleBlock(); mb != nil {
			info.Statements = b.buildNodeList(mb.Statements, "Statements", info)
		}
	}

	// ImportDeclaration specific
	if node.Kind == ast.KindImportDeclaration {
		if id := node.AsImportDeclaration(); id != nil {
			if id.ImportClause != nil {
				info.ImportClause = b.BuildShallowNodeInfo(id.ImportClause.AsNode())
			}
			if id.ModuleSpecifier != nil {
				info.ModuleSpecifier = b.BuildShallowNodeInfo(id.ModuleSpecifier)
			}
		}
	}

	// ImportClause specific
	if node.Kind == ast.KindImportClause {
		if ic := node.AsImportClause(); ic != nil {
			if ic.NamedBindings != nil {
				info.NamedBindings = b.BuildShallowNodeInfo(ic.NamedBindings)
			}
		}
	}

	// NamedImports specific
	if node.Kind == ast.KindNamedImports {
		if ni := node.AsNamedImports(); ni != nil {
			info.Elements = b.buildNodeList(ni.Elements, "Elements", info)
		}
	}

	// ExportDeclaration specific
	if node.Kind == ast.KindExportDeclaration {
		if ed := node.AsExportDeclaration(); ed != nil {
			if ed.ExportClause != nil {
				info.ExportClause = b.BuildShallowNodeInfo(ed.ExportClause)
			}
			if ed.ModuleSpecifier != nil {
				info.ModuleSpecifier = b.BuildShallowNodeInfo(ed.ModuleSpecifier)
			}
		}
	}

	// NamedExports specific
	if node.Kind == ast.KindNamedExports {
		if ne := node.AsNamedExports(); ne != nil {
			info.Elements = b.buildNodeList(ne.Elements, "Elements", info)
		}
	}

	// ForStatement specific
	if node.Kind == ast.KindForStatement {
		if fs := node.AsForStatement(); fs != nil {
			if fs.Initializer != nil {
				info.Initializer = b.BuildShallowNodeInfo(fs.Initializer)
			}
			if fs.Condition != nil {
				info.Condition = b.BuildShallowNodeInfo(fs.Condition)
			}
			if fs.Incrementor != nil {
				info.Incrementor = b.BuildShallowNodeInfo(fs.Incrementor)
			}
			if fs.Statement != nil {
				info.Statement = b.BuildShallowNodeInfo(fs.Statement)
			}
		}
	}

	// WhileStatement specific
	if node.Kind == ast.KindWhileStatement {
		if ws := node.AsWhileStatement(); ws != nil {
			info.Expression = b.BuildShallowNodeInfo(ws.Expression)
			if ws.Statement != nil {
				info.Statement = b.BuildShallowNodeInfo(ws.Statement)
			}
		}
	}

	// DoStatement specific
	if node.Kind == ast.KindDoStatement {
		if ds := node.AsDoStatement(); ds != nil {
			if ds.Statement != nil {
				info.Statement = b.BuildShallowNodeInfo(ds.Statement)
			}
			info.Expression = b.BuildShallowNodeInfo(ds.Expression)
		}
	}

	// SwitchStatement specific
	if node.Kind == ast.KindSwitchStatement {
		if ss := node.AsSwitchStatement(); ss != nil {
			info.Expression = b.BuildShallowNodeInfo(ss.Expression)
			if ss.CaseBlock != nil {
				info.CaseBlock = b.BuildShallowNodeInfo(ss.CaseBlock.AsNode())
			}
		}
	}

	// CaseBlock specific
	if node.Kind == ast.KindCaseBlock {
		if cb := node.AsCaseBlock(); cb != nil {
			info.Clauses = b.buildNodeList(cb.Clauses, "Clauses", info)
		}
	}

	// TryStatement specific
	if node.Kind == ast.KindTryStatement {
		if ts := node.AsTryStatement(); ts != nil {
			if ts.TryBlock != nil {
				info.TryBlock = b.BuildShallowNodeInfo(ts.TryBlock.AsNode())
			}
			if ts.CatchClause != nil {
				info.CatchClause = b.BuildShallowNodeInfo(ts.CatchClause.AsNode())
			}
			if ts.FinallyBlock != nil {
				info.FinallyBlock = b.BuildShallowNodeInfo(ts.FinallyBlock.AsNode())
			}
		}
	}

	// CatchClause specific
	if node.Kind == ast.KindCatchClause {
		if cc := node.AsCatchClause(); cc != nil {
			if cc.VariableDeclaration != nil {
				info.VariableDeclaration = b.BuildShallowNodeInfo(cc.VariableDeclaration.AsNode())
			}
			if cc.Block != nil {
				info.Block = b.BuildShallowNodeInfo(cc.Block.AsNode())
			}
		}
	}

	// ReturnStatement specific
	if node.Kind == ast.KindReturnStatement {
		if rs := node.AsReturnStatement(); rs != nil {
			if rs.Expression != nil {
				info.Expression = b.BuildShallowNodeInfo(rs.Expression)
			}
		}
	}

	// ThrowStatement specific
	if node.Kind == ast.KindThrowStatement {
		if ts := node.AsThrowStatement(); ts != nil {
			info.Expression = b.BuildShallowNodeInfo(ts.Expression)
		}
	}

	// PropertyAccessExpression specific
	if node.Kind == ast.KindPropertyAccessExpression {
		if pae := node.AsPropertyAccessExpression(); pae != nil {
			info.Expression = b.BuildShallowNodeInfo(pae.Expression)
			info.Name = b.BuildShallowNodeInfo(pae.Name())
		}
	}

	// ElementAccessExpression specific
	if node.Kind == ast.KindElementAccessExpression {
		if eae := node.AsElementAccessExpression(); eae != nil {
			info.Expression = b.BuildShallowNodeInfo(eae.Expression)
			info.ArgumentExpression = b.BuildShallowNodeInfo(eae.ArgumentExpression)
		}
	}

	// PropertyAssignment specific
	if node.Kind == ast.KindPropertyAssignment {
		if pa := node.AsPropertyAssignment(); pa != nil {
			info.Name = b.BuildShallowNodeInfo(pa.Name())
			info.Initializer = b.BuildShallowNodeInfo(pa.Initializer)
		}
	}

	// ShorthandPropertyAssignment specific
	if node.Kind == ast.KindShorthandPropertyAssignment {
		if spa := node.AsShorthandPropertyAssignment(); spa != nil {
			info.Name = b.BuildShallowNodeInfo(spa.Name())
			if spa.EqualsToken != nil {
				info.EqualsToken = b.BuildShallowNodeInfo(spa.EqualsToken.AsNode())
			}
			if spa.ObjectAssignmentInitializer != nil {
				info.ObjectAssignmentInitializer = b.BuildShallowNodeInfo(spa.ObjectAssignmentInitializer)
			}
		}
	}

	// SpreadAssignment specific
	if node.Kind == ast.KindSpreadAssignment {
		if sa := node.AsSpreadAssignment(); sa != nil {
			info.Expression = b.BuildShallowNodeInfo(sa.Expression)
		}
	}

	// SpreadElement specific
	if node.Kind == ast.KindSpreadElement {
		if se := node.AsSpreadElement(); se != nil {
			info.Expression = b.BuildShallowNodeInfo(se.Expression)
		}
	}

	// TemplateExpression specific
	if node.Kind == ast.KindTemplateExpression {
		if te := node.AsTemplateExpression(); te != nil {
			info.Head = b.BuildShallowNodeInfo(te.Head.AsNode())
			info.TemplateSpans = b.buildNodeList(te.TemplateSpans, "TemplateSpans", info)
		}
	}

	// TemplateSpan specific
	if node.Kind == ast.KindTemplateSpan {
		if ts := node.AsTemplateSpan(); ts != nil {
			info.Expression = b.BuildShallowNodeInfo(ts.Expression)
			info.Literal = b.BuildShallowNodeInfo(ts.Literal.AsNode())
		}
	}

	// TaggedTemplateExpression specific
	if node.Kind == ast.KindTaggedTemplateExpression {
		if tte := node.AsTaggedTemplateExpression(); tte != nil {
			info.Tag = b.BuildShallowNodeInfo(tte.Tag)
			info.Template = b.BuildShallowNodeInfo(tte.Template)
		}
	}

	// AsExpression specific (type assertion)
	if node.Kind == ast.KindAsExpression {
		if ae := node.AsAsExpression(); ae != nil {
			info.Expression = b.BuildShallowNodeInfo(ae.Expression)
			info.TypeNode = b.BuildShallowNodeInfo(ae.Type)
		}
	}

	// TypeAssertionExpression specific
	if node.Kind == ast.KindTypeAssertionExpression {
		if tae := node.AsTypeAssertion(); tae != nil {
			info.TypeNode = b.BuildShallowNodeInfo(tae.Type)
			info.Expression = b.BuildShallowNodeInfo(tae.Expression)
		}
	}

	// ParenthesizedExpression specific
	if node.Kind == ast.KindParenthesizedExpression {
		if pe := node.AsParenthesizedExpression(); pe != nil {
			info.Expression = b.BuildShallowNodeInfo(pe.Expression)
		}
	}

	// AwaitExpression specific
	if node.Kind == ast.KindAwaitExpression {
		if ae := node.AsAwaitExpression(); ae != nil {
			info.Expression = b.BuildShallowNodeInfo(ae.Expression)
		}
	}

	// YieldExpression specific
	if node.Kind == ast.KindYieldExpression {
		if ye := node.AsYieldExpression(); ye != nil {
			if ye.Expression != nil {
				info.Expression = b.BuildShallowNodeInfo(ye.Expression)
			}
		}
	}

	// TypeOfExpression specific
	if node.Kind == ast.KindTypeOfExpression {
		if toe := node.AsTypeOfExpression(); toe != nil {
			info.Expression = b.BuildShallowNodeInfo(toe.Expression)
		}
	}

	// DeleteExpression specific
	if node.Kind == ast.KindDeleteExpression {
		if de := node.AsDeleteExpression(); de != nil {
			info.Expression = b.BuildShallowNodeInfo(de.Expression)
		}
	}

	// VoidExpression specific
	if node.Kind == ast.KindVoidExpression {
		if ve := node.AsVoidExpression(); ve != nil {
			info.Expression = b.BuildShallowNodeInfo(ve.Expression)
		}
	}

	// ParameterDeclaration specific
	if node.Kind == ast.KindParameter {
		if pd := node.AsParameterDeclaration(); pd != nil {
			if pd.QuestionToken != nil {
				info.QuestionToken = b.BuildShallowNodeInfo(pd.QuestionToken.AsNode())
			}
			if pd.DotDotDotToken != nil {
				info.DotDotDotToken = b.BuildShallowNodeInfo(pd.DotDotDotToken.AsNode())
			}
		}
	}

	// ArrowFunction specific
	if node.Kind == ast.KindArrowFunction {
		if af := node.AsArrowFunction(); af != nil {
			if af.EqualsGreaterThanToken != nil {
				info.EqualsGreaterThanToken = b.BuildShallowNodeInfo(af.EqualsGreaterThanToken.AsNode())
			}
		}
	}

	// TypeParameterDeclaration specific
	if node.Kind == ast.KindTypeParameter {
		if tpd := node.AsTypeParameter(); tpd != nil {
			if tpd.Constraint != nil {
				info.Constraint = b.BuildShallowNodeInfo(tpd.Constraint.AsNode())
			}
			if tpd.DefaultType != nil {
				info.DefaultType = b.BuildShallowNodeInfo(tpd.DefaultType.AsNode())
			}
		}
	}

	// PropertyDeclaration specific (PostfixToken: ? or !)
	if node.Kind == ast.KindPropertyDeclaration {
		if pd := node.AsPropertyDeclaration(); pd != nil {
			if pd.PostfixToken != nil {
				tokenNode := b.BuildShallowNodeInfo(pd.PostfixToken.AsNode())
				switch pd.PostfixToken.Kind {
				case ast.KindQuestionToken:
					info.QuestionToken = tokenNode
				case ast.KindExclamationToken:
					info.ExclamationToken = tokenNode
				}
			}
		}
	}

	// PropertySignature specific - directly access fields since generic methods may not work
	if node.Kind == ast.KindPropertySignature {
		if ps := node.AsPropertySignatureDeclaration(); ps != nil {
			// Name
			if name := ps.Name(); name != nil {
				info.Name = b.BuildShallowNodeInfo(name.AsNode())
			}
			// QuestionToken (optional property marker)
			if ps.PostfixToken != nil && ps.PostfixToken.Kind == ast.KindQuestionToken {
				info.QuestionToken = b.BuildShallowNodeInfo(ps.PostfixToken.AsNode())
			}
			// Type
			if ps.Type != nil {
				info.TypeNode = b.BuildShallowNodeInfo(ps.Type.AsNode())
			}
			// Initializer (for error reporting)
			if ps.Initializer != nil {
				info.Initializer = b.BuildShallowNodeInfo(ps.Initializer)
			}
		}
	}

	// MethodDeclaration specific (PostfixToken: ?)
	if node.Kind == ast.KindMethodDeclaration {
		if md := node.AsMethodDeclaration(); md != nil {
			if md.PostfixToken != nil {
				tokenKind := md.PostfixToken.Kind
				tokenNode := b.BuildShallowNodeInfo(md.PostfixToken.AsNode())
				if tokenKind == ast.KindQuestionToken {
					info.QuestionToken = tokenNode
				}
			}
		}
	}

	// MethodSignature specific - directly access fields since generic methods may not work
	if node.Kind == ast.KindMethodSignature {
		if ms := node.AsMethodSignatureDeclaration(); ms != nil {
			// Name
			if name := ms.Name(); name != nil {
				info.Name = b.BuildShallowNodeInfo(name.AsNode())
			}
			// QuestionToken (optional method marker)
			if ms.PostfixToken != nil && ms.PostfixToken.Kind == ast.KindQuestionToken {
				info.QuestionToken = b.BuildShallowNodeInfo(ms.PostfixToken.AsNode())
			}
			// TypeParameters
			info.TypeParameters = b.buildNodeList(ms.TypeParameters, "TypeParameters", info)
			// Parameters
			info.Parameters = b.buildNodeList(ms.Parameters, "Parameters", info)
			// Return type
			if ms.Type != nil {
				info.TypeNode = b.BuildShallowNodeInfo(ms.Type.AsNode())
			}
		}
	}

	// CallSignature specific - directly access fields since generic methods may not work
	if node.Kind == ast.KindCallSignature {
		if cs := node.AsCallSignatureDeclaration(); cs != nil {
			// TypeParameters
			info.TypeParameters = b.buildNodeList(cs.TypeParameters, "TypeParameters", info)
			// Parameters
			info.Parameters = b.buildNodeList(cs.Parameters, "Parameters", info)
			// Return type
			if cs.Type != nil {
				info.TypeNode = b.BuildShallowNodeInfo(cs.Type.AsNode())
			}
		}
	}

	// ConstructSignature specific - directly access fields since generic methods may not work
	if node.Kind == ast.KindConstructSignature {
		if cs := node.AsConstructSignatureDeclaration(); cs != nil {
			// TypeParameters
			info.TypeParameters = b.buildNodeList(cs.TypeParameters, "TypeParameters", info)
			// Parameters
			info.Parameters = b.buildNodeList(cs.Parameters, "Parameters", info)
			// Return type
			if cs.Type != nil {
				info.TypeNode = b.BuildShallowNodeInfo(cs.Type.AsNode())
			}
		}
	}

	// IndexSignature specific - directly access fields since generic methods may not work
	if node.Kind == ast.KindIndexSignature {
		if is := node.AsIndexSignatureDeclaration(); is != nil {
			// Parameters (the index parameter)
			info.Parameters = b.buildNodeList(is.Parameters, "Parameters", info)
			// Return type (the value type)
			if is.Type != nil {
				info.TypeNode = b.BuildShallowNodeInfo(is.Type.AsNode())
			}
		}
	}

	// Modifiers (for declarations)
	SafeCall(func() {
		if modifiers := node.Modifiers(); modifiers != nil {
			info.Modifiers = b.buildNodeList(&modifiers.NodeList, "Modifiers", info)
		}
	})
}
