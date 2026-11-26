package estree

import (
	"fmt"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	esregexp "github.com/web-infra-dev/rslint/internal/utils/ecmascript/regexp"
)

// TreeOptions configures in-memory ESTree projection via Builder.WithTree.
type TreeOptions struct {
	SourceType string
}

// ValueKind discriminates the runtime value returned by Tree.Field.
type ValueKind uint8

const (
	ValueUndefined ValueKind = iota
	ValueNull
	ValueBool
	ValueNumber
	ValueString
	ValueBigInt
	ValueRegExp
	ValueRegexObject
	ValueTemplateValue
	ValueNode
	ValueList
)

// Value represents one field value read from a projected Tree node.
type Value struct {
	Kind ValueKind
	Bool bool
	Num  float64
	Str  string
	Node NodeID
	list List
}

// List returns the read-only child list handle when Kind == ValueList.
func (v Value) List() List {
	return v.list
}

// List is a read-only bounded view of an ESTree child array inside a Tree.
// Sparse array holes (elisions) are preserved as NoNode entries.
type List struct {
	elems []uint32
}

// Len returns the number of slots in the list, including sparse holes.
func (l List) Len() int {
	return len(l.elems)
}

// At returns the child NodeID at index, or NoNode for a sparse array hole.
func (l List) At(index int) NodeID {
	return NodeID(l.elems[index])
}

type nodeLink struct {
	parent      NodeID
	listOffset  uint32 // valid only when listIndex >= 0
	listIndex   int32  // -1 when not an element of a list
	parentField FieldID
}

// Tree is a read-only in-memory TSESTree view scoped to a Builder.WithTree
// callback. Node ranges are UTF-8 byte offsets into the source file.
type Tree struct {
	root    NodeID
	nodes   []nodeRecord
	values  []uint32
	lists   []uint32
	strings []string
	numbers []float64
	links   []nodeLink
}

// WithTree projects file into a read-only in-memory TSESTree with UTF-8 byte
// ranges and passes it to fn. The Tree borrows Builder scratch storage and must
// not escape fn. Source-backed references are always cleared before returning,
// including if projection fails or fn panics.
func (b *Builder) WithTree(file *ast.SourceFile, options TreeOptions, fn func(*Tree)) error {
	if options.SourceType != "module" && options.SourceType != "script" && options.SourceType != "commonjs" {
		return fmt.Errorf("invalid ESTree sourceType %q", options.SourceType)
	}
	defer b.p.clearSource()

	root, err := b.projectTree(file, options)
	if err != nil {
		return err
	}

	p := &b.p
	tree := Tree{
		root:    NodeID(root),
		nodes:   p.nodes,
		values:  p.values,
		lists:   p.lists,
		strings: p.strings,
		numbers: p.numbers,
		links:   p.links,
	}
	fn(&tree)
	return nil
}

func (b *Builder) projectTree(file *ast.SourceFile, options TreeOptions) (root uint32, err error) {
	defer func() {
		if failure := recover(); failure != nil {
			if pe, ok := failure.(projectionError); ok {
				root = 0
				err = pe.err
				return
			}
			panic(failure)
		}
	}()

	p := &b.p
	p.ctx = nil
	p.file = file
	p.text = file.Text()
	p.isJS = file.ScriptKind == core.ScriptKindJS || file.ScriptKind == core.ScriptKindJSX
	p.sourceType = options.SourceType
	if options.SourceType == "commonjs" && !p.isJS {
		p.sourceType = "script"
	}
	p.mapping = nil
	p.treeMode = true
	p.needTokens = false
	p.lastTriviaPos = -1
	p.lastTriviaStart = 0
	// Empty offsets table makes p.offsets.At(offset) return UTF-8 byte offsets.
	p.offsets.entries = p.offsets.entries[:0]
	p.offsets.delta = 0

	nodeCapacity := min(file.NodeCount, maxInitialFrameBytes/14)
	p.nodes = slices.Grow(p.nodes[:0], nodeCapacity)
	p.values = slices.Grow(p.values[:0], min(nodeCapacity*4, maxInitialFrameBytes/4))
	p.lists = slices.Grow(p.lists[:0], nodeCapacity)
	p.emptyList = 0
	p.strings = slices.Grow(p.strings[:0], nodeCapacity)
	p.stringBytes = 0
	p.numbers = p.numbers[:0]
	p.tokenCovers = p.tokenCovers[:0]
	p.listScratch = p.listScratch[:0]

	root = p.convert(file.AsNode(), false)
	p.linkTree(NodeID(root))
	return root, nil
}

func (p *estreeProjector) linkTree(root NodeID) {
	count := len(p.nodes)
	if cap(p.links) < count {
		p.links = make([]nodeLink, count)
	} else {
		p.links = p.links[:count]
	}
	for i := range p.links {
		p.links[i] = nodeLink{listIndex: -1}
	}
	if root == NoNode {
		return
	}
	var stackBuf [64]NodeID
	stack := append(stackBuf[:0], root)
	for len(stack) > 0 {
		id := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		rec := p.nodes[id-1]
		base := rec.valueStart
		for _, cs := range childSlots[rec.typeID] {
			encoded := p.values[base+uint32(cs.slot)]
			tag := encoded >> tagShift
			payload := encoded & payloadMask
			switch tag {
			case tagNode:
				child := NodeID(payload)
				if child == NoNode {
					continue
				}
				p.links[child-1] = nodeLink{
					parent:      id,
					listIndex:   -1,
					parentField: cs.field,
				}
				stack = append(stack, child)
			case tagList:
				length := p.lists[payload]
				elemStart := payload + 1
				for idx := range length {
					child := NodeID(p.lists[elemStart+idx])
					if child == NoNode {
						continue
					}
					p.links[child-1] = nodeLink{
						parent:      id,
						listOffset:  payload,
						listIndex:   int32(idx),
						parentField: cs.field,
					}
					stack = append(stack, child)
				}
			}
		}
	}
}

func (t *Tree) validNode(id NodeID) bool {
	return id != NoNode && int(id) <= len(t.nodes)
}

func (t *Tree) listAt(payload uint32) List {
	length := t.lists[payload]
	start := payload + 1
	end := start + length
	return List{elems: t.lists[start:end:end]}
}

// Root returns the root Program NodeID.
func (t *Tree) Root() NodeID {
	return t.root
}

// NodeCount returns the total number of projected nodes.
func (t *Tree) NodeCount() int {
	return len(t.nodes)
}

// Type returns the ESTree TypeID for id, or UnknownType if id is invalid.
func (t *Tree) Type(id NodeID) TypeID {
	if !t.validNode(id) {
		return UnknownType
	}
	return TypeID(t.nodes[id-1].typeID)
}

// Range returns the UTF-8 byte TextRange for id.
func (t *Tree) Range(id NodeID) core.TextRange {
	if !t.validNode(id) {
		return core.TextRange{}
	}
	rec := t.nodes[id-1]
	return core.NewTextRange(int(rec.start), int(rec.end))
}

// Parent returns the parent NodeID of id, or NoNode for Program or invalid id.
func (t *Tree) Parent(id NodeID) NodeID {
	if !t.validNode(id) {
		return NoNode
	}
	return t.links[id-1].parent
}

// ParentField returns the FieldID on Parent(id) through which id is attached.
func (t *Tree) ParentField(id NodeID) FieldID {
	if !t.validNode(id) {
		return NoField
	}
	return t.links[id-1].parentField
}

// SiblingList returns the parent's list containing id and id's 0-based raw
// index in that list (preserving sparse array holes).
func (t *Tree) SiblingList(id NodeID) (List, int, bool) {
	if !t.validNode(id) {
		return List{}, -1, false
	}
	link := t.links[id-1]
	if link.listIndex < 0 {
		return List{}, -1, false
	}
	return t.listAt(link.listOffset), int(link.listIndex), true
}

// ForEachChild visits every non-null child of id in canonical
// @typescript-eslint/visitor-keys order. Stops early if fn returns true.
func (t *Tree) ForEachChild(id NodeID, fn func(NodeID) bool) bool {
	if !t.validNode(id) {
		return false
	}
	rec := t.nodes[id-1]
	base := rec.valueStart
	for _, cs := range childSlots[rec.typeID] {
		encoded := t.values[base+uint32(cs.slot)]
		tag := encoded >> tagShift
		payload := encoded & payloadMask
		switch tag {
		case tagNode:
			child := NodeID(payload)
			if child != NoNode && fn(child) {
				return true
			}
		case tagList:
			length := t.lists[payload]
			elemStart := payload + 1
			for idx := range length {
				child := NodeID(t.lists[elemStart+idx])
				if child != NoNode && fn(child) {
					return true
				}
			}
		}
	}
	return false
}

// AppendChildren appends every non-null child of id in canonical
// @typescript-eslint/visitor-keys order to dst and returns the updated slice.
func (t *Tree) AppendChildren(id NodeID, dst []NodeID) []NodeID {
	if !t.validNode(id) {
		return dst
	}
	rec := t.nodes[id-1]
	base := rec.valueStart
	for _, cs := range childSlots[rec.typeID] {
		encoded := t.values[base+uint32(cs.slot)]
		tag := encoded >> tagShift
		payload := encoded & payloadMask
		switch tag {
		case tagNode:
			if child := NodeID(payload); child != NoNode {
				dst = append(dst, child)
			}
		case tagList:
			length := t.lists[payload]
			elemStart := payload + 1
			for idx := range length {
				if child := NodeID(t.lists[elemStart+idx]); child != NoNode {
					dst = append(dst, child)
				}
			}
		}
	}
	return dst
}

// Field reads a public ESTree field from node id. Internal $-prefixed wire
// fields are synthesized into their standard ESTree shapes (Literal.regex,
// Literal.value for BigInt/RegExp, and TemplateElement.value).
func (t *Tree) Field(id NodeID, field FieldID) Value {
	if !t.validNode(id) || field == NoField || int(field) >= maxFieldID {
		return Value{Kind: ValueUndefined}
	}
	rec := t.nodes[id-1]
	tid := TypeID(rec.typeID)

	if tid == typeLiteral {
		switch field {
		case fieldValue:
			if bigintEnc := t.values[rec.valueStart+slotLiteralBigInt]; bigintEnc>>tagShift == tagString {
				return Value{Kind: ValueBigInt, Str: t.strings[bigintEnc&payloadMask]}
			}
			if patEnc := t.values[rec.valueStart+slotLiteralRegexPattern]; patEnc>>tagShift == tagString {
				flagsEnc := t.values[rec.valueStart+slotLiteralRegexFlags]
				if flagsEnc>>tagShift != tagString {
					return Value{Kind: ValueNull}
				}
				pattern := t.strings[patEnc&payloadMask]
				flags := t.strings[flagsEnc&payloadMask]
				if isValidJSRegExp(pattern, flags) {
					return Value{Kind: ValueRegExp, Node: id}
				}
				return Value{Kind: ValueNull}
			}
		case fieldRegex:
			if t.values[rec.valueStart+slotLiteralRegexPattern]>>tagShift == tagString {
				return Value{Kind: ValueRegexObject, Node: id}
			}
			return Value{Kind: ValueUndefined}
		}
	} else if tid == typeTemplateElement && field == fieldValue {
		return Value{Kind: ValueTemplateValue, Node: id}
	}

	slot := fieldSlot[tid][field]
	if slot < 0 {
		return Value{Kind: ValueUndefined}
	}
	return t.decodeValue(t.values[rec.valueStart+uint32(slot)])
}

func (t *Tree) decodeValue(encoded uint32) Value {
	payload := encoded & payloadMask
	switch encoded >> tagShift {
	case tagNull:
		return Value{Kind: ValueNull}
	case tagNode:
		if payload == 0 {
			return Value{Kind: ValueNull}
		}
		return Value{Kind: ValueNode, Node: NodeID(payload)}
	case tagList:
		return Value{Kind: ValueList, list: t.listAt(payload)}
	case tagString:
		return Value{Kind: ValueString, Str: t.strings[payload]}
	case tagFalse:
		return Value{Kind: ValueBool, Bool: false}
	case tagTrue:
		return Value{Kind: ValueBool, Bool: true}
	case tagNumber:
		return Value{Kind: ValueNumber, Num: t.numbers[payload]}
	default:
		return Value{Kind: ValueUndefined}
	}
}

// RegexInfo returns the authored pattern and flags of a RegExp Literal node.
func (t *Tree) RegexInfo(id NodeID) (pattern, flags string, ok bool) {
	if !t.validNode(id) || TypeID(t.nodes[id-1].typeID) != typeLiteral {
		return "", "", false
	}
	base := t.nodes[id-1].valueStart
	patEnc := t.values[base+slotLiteralRegexPattern]
	flagsEnc := t.values[base+slotLiteralRegexFlags]
	if patEnc>>tagShift != tagString || flagsEnc>>tagShift != tagString {
		return "", "", false
	}
	return t.strings[patEnc&payloadMask], t.strings[flagsEnc&payloadMask], true
}

// TemplateValue returns the cooked and raw strings of a TemplateElement node.
func (t *Tree) TemplateValue(id NodeID) (cooked string, cookedNull bool, raw string, ok bool) {
	if !t.validNode(id) || TypeID(t.nodes[id-1].typeID) != typeTemplateElement {
		return "", false, "", false
	}
	base := t.nodes[id-1].valueStart
	cookedEnc := t.values[base+slotTemplateCooked]
	rawEnc := t.values[base+slotTemplateRaw]
	switch cookedEnc >> tagShift {
	case tagNull:
		cookedNull = true
	case tagString:
		cooked = t.strings[cookedEnc&payloadMask]
	}
	if rawEnc>>tagShift == tagString {
		raw = t.strings[rawEnc&payloadMask]
	}
	return cooked, cookedNull, raw, true
}

// isValidJSRegExp checks whether new RegExp(pattern, flags) would succeed in
// JavaScript without rejecting ES2024 `v`-flag literals that esregexp does not
// compile.
func isValidJSRegExp(pattern, flags string) bool {
	var seen uint8
	for i := range len(flags) {
		var bit uint8
		switch flags[i] {
		case 'd':
			bit = 1 << 0
		case 'g':
			bit = 1 << 1
		case 'i':
			bit = 1 << 2
		case 'm':
			bit = 1 << 3
		case 's':
			bit = 1 << 4
		case 'u':
			bit = 1 << 5
		case 'v':
			bit = 1 << 6
		case 'y':
			bit = 1 << 7
		default:
			return false
		}
		if seen&bit != 0 {
			return false
		}
		seen |= bit
	}
	// `u` and `v` flags are mutually exclusive in ECMAScript.
	if seen&(1<<5) != 0 && seen&(1<<6) != 0 {
		return false
	}
	// esregexp does not implement the ES2024 `v` flag yet; do not treat valid
	// `/.../v` literals as null.
	if seen&(1<<6) != 0 {
		return true
	}
	_, err := esregexp.Compile(pattern, flags)
	return err == nil
}

// triggerKindsTable maps provably closed ESTree TypeIDs to the complete set of
// tsgo ast.Kind values that estreeProjector can project into that TypeID.
// Any TypeID omitted here (including synthesized, multi-source, or ubiquitous
// types) returns ok=false from TriggerKinds so callers run eagerly.
var triggerKindsTable = [schemaCount][]ast.Kind{
	typeBlockStatement:                  {ast.KindBlock},
	typeTSModuleBlock:                   {ast.KindModuleBlock},
	typeEmptyStatement:                  {ast.KindEmptyStatement},
	typeDebuggerStatement:               {ast.KindDebuggerStatement},
	typeExpressionStatement:             {ast.KindExpressionStatement},
	typeReturnStatement:                 {ast.KindReturnStatement},
	typeThrowStatement:                  {ast.KindThrowStatement},
	typeIfStatement:                     {ast.KindIfStatement},
	typeWhileStatement:                  {ast.KindWhileStatement},
	typeDoWhileStatement:                {ast.KindDoStatement},
	typeForStatement:                    {ast.KindForStatement},
	typeForInStatement:                  {ast.KindForInStatement},
	typeForOfStatement:                  {ast.KindForOfStatement},
	typeBreakStatement:                  {ast.KindBreakStatement},
	typeContinueStatement:               {ast.KindContinueStatement},
	typeWithStatement:                   {ast.KindWithStatement},
	typeSwitchStatement:                 {ast.KindSwitchStatement},
	typeSwitchCase:                      {ast.KindCaseClause, ast.KindDefaultClause},
	typeTryStatement:                    {ast.KindTryStatement},
	typeCatchClause:                     {ast.KindCatchClause},
	typeLabeledStatement:                {ast.KindLabeledStatement},
	typeVariableDeclaration:             {ast.KindVariableStatement, ast.KindVariableDeclarationList},
	typeVariableDeclarator:              {ast.KindVariableDeclaration},
	typeFunctionDeclaration:             {ast.KindFunctionDeclaration},
	typeTSDeclareFunction:               {ast.KindFunctionDeclaration},
	typeArrowFunctionExpression:         {ast.KindArrowFunction},
	typePrivateIdentifier:               {ast.KindPrivateIdentifier},
	typeSuper:                           {ast.KindSuperKeyword},
	typeBinaryExpression:                {ast.KindBinaryExpression},
	typeLogicalExpression:               {ast.KindBinaryExpression},
	typeAssignmentExpression:            {ast.KindBinaryExpression},
	typeSequenceExpression:              {ast.KindBinaryExpression},
	typeConditionalExpression:           {ast.KindConditionalExpression},
	typeUnaryExpression:                 {ast.KindPrefixUnaryExpression, ast.KindDeleteExpression, ast.KindVoidExpression, ast.KindTypeOfExpression},
	typeUpdateExpression:                {ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression},
	typeAwaitExpression:                 {ast.KindAwaitExpression},
	typeYieldExpression:                 {ast.KindYieldExpression},
	typeTSAsExpression:                  {ast.KindAsExpression},
	typeTSSatisfiesExpression:           {ast.KindSatisfiesExpression},
	typeTSNonNullExpression:             {ast.KindNonNullExpression},
	typeTSTypeAssertion:                 {ast.KindTypeAssertionExpression},
	typeChainExpression:                 {ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression, ast.KindNonNullExpression},
	typeMemberExpression:                {ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindQualifiedName},
	typeCallExpression:                  {ast.KindCallExpression},
	typeNewExpression:                   {ast.KindNewExpression},
	typeImportExpression:                {ast.KindCallExpression},
	typeArrayExpression:                 {ast.KindArrayLiteralExpression},
	typeSpreadElement:                   {ast.KindSpreadElement, ast.KindSpreadAssignment},
	typeTemplateLiteral:                 {ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression},
	typeTaggedTemplateExpression:        {ast.KindTaggedTemplateExpression},
	typeClassDeclaration:                {ast.KindClassDeclaration},
	typeClassExpression:                 {ast.KindClassExpression},
	typeClassBody:                       {ast.KindClassDeclaration, ast.KindClassExpression},
	typeMethodDefinition:                {ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor},
	typePropertyDefinition:              {ast.KindPropertyDeclaration},
	typeStaticBlock:                     {ast.KindClassStaticBlockDeclaration},
	typeTSParameterProperty:             {ast.KindParameter},
	typeDecorator:                       {ast.KindDecorator},
	typeImportDeclaration:               {ast.KindImportDeclaration},
	typeImportDefaultSpecifier:          {ast.KindImportClause},
	typeImportNamespaceSpecifier:        {ast.KindNamespaceImport},
	typeImportSpecifier:                 {ast.KindImportSpecifier},
	typeExportAllDeclaration:            {ast.KindExportDeclaration},
	typeExportSpecifier:                 {ast.KindExportSpecifier},
	typeTSExportAssignment:              {ast.KindExportAssignment},
	typeImportAttribute:                 {ast.KindImportAttribute},
	typeTSInterfaceDeclaration:          {ast.KindInterfaceDeclaration},
	typeTSInterfaceBody:                 {ast.KindInterfaceDeclaration},
	typeTSTypeAliasDeclaration:          {ast.KindTypeAliasDeclaration, ast.KindJSTypeAliasDeclaration},
	typeTSPropertySignature:             {ast.KindPropertySignature},
	typeTSIndexSignature:                {ast.KindIndexSignature},
	typeTSTypeReference:                 {ast.KindTypeReference},
	typeTSTypeLiteral:                   {ast.KindTypeLiteral},
	typeTSUnionType:                     {ast.KindUnionType},
	typeTSIntersectionType:              {ast.KindIntersectionType},
	typeTSArrayType:                     {ast.KindArrayType},
	typeTSLiteralType:                   {ast.KindLiteralType},
	typeTSTypeParameter:                 {ast.KindTypeParameter},
	typeTSStringKeyword:                 {ast.KindStringKeyword},
	typeTSNumberKeyword:                 {ast.KindNumberKeyword},
	typeTSBooleanKeyword:                {ast.KindBooleanKeyword},
	typeTSAnyKeyword:                    {ast.KindAnyKeyword},
	typeTSUnknownKeyword:                {ast.KindUnknownKeyword},
	typeTSVoidKeyword:                   {ast.KindVoidKeyword},
	typeTSUndefinedKeyword:              {ast.KindUndefinedKeyword},
	typeTSNeverKeyword:                  {ast.KindNeverKeyword},
	typeTSObjectKeyword:                 {ast.KindObjectKeyword},
	typeTSSymbolKeyword:                 {ast.KindSymbolKeyword},
	typeTSBigIntKeyword:                 {ast.KindBigIntKeyword},
	typeTSIntrinsicKeyword:              {ast.KindIntrinsicKeyword},
	typeTSFunctionType:                  {ast.KindFunctionType},
	typeTSConstructorType:               {ast.KindConstructorType},
	typeTSTupleType:                     {ast.KindTupleType},
	typeTSNamedTupleMember:              {ast.KindNamedTupleMember},
	typeTSOptionalType:                  {ast.KindOptionalType},
	typeTSIndexedAccessType:             {ast.KindIndexedAccessType},
	typeTSTypeOperator:                  {ast.KindTypeOperator},
	typeJSXElement:                      {ast.KindJsxElement, ast.KindJsxSelfClosingElement},
	typeJSXOpeningElement:               {ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement},
	typeJSXClosingElement:               {ast.KindJsxClosingElement},
	typeJSXFragment:                     {ast.KindJsxFragment},
	typeJSXOpeningFragment:              {ast.KindJsxOpeningFragment},
	typeJSXClosingFragment:              {ast.KindJsxClosingFragment},
	typeJSXMemberExpression:             {ast.KindPropertyAccessExpression},
	typeJSXNamespacedName:               {ast.KindJsxNamespacedName},
	typeJSXAttribute:                    {ast.KindJsxAttribute},
	typeJSXSpreadAttribute:              {ast.KindJsxSpreadAttribute},
	typeJSXExpressionContainer:          {ast.KindJsxExpression},
	typeJSXEmptyExpression:              {ast.KindJsxExpression},
	typeJSXSpreadChild:                  {ast.KindJsxExpression},
	typeJSXText:                         {ast.KindJsxText, ast.KindJsxTextAllWhiteSpaces},
	typeMetaProperty:                    {ast.KindMetaProperty},
	typeTSCallSignatureDeclaration:      {ast.KindCallSignature},
	typeTSConstructSignatureDeclaration: {ast.KindConstructSignature},
	typeTSTypePredicate:                 {ast.KindTypePredicate},
	typeTSImportType:                    {ast.KindImportType},
	typeTSEnumBody:                      {ast.KindEnumDeclaration},
	typeTSEnumDeclaration:               {ast.KindEnumDeclaration},
	typeTSEnumMember:                    {ast.KindEnumMember},
	typeTSModuleDeclaration:             {ast.KindModuleDeclaration},
	typeTSImportEqualsDeclaration:       {ast.KindImportEqualsDeclaration},
	typeTSExternalModuleReference:       {ast.KindExternalModuleReference},
	typeTSConditionalType:               {ast.KindConditionalType},
	typeTSMappedType:                    {ast.KindMappedType},
	typeTSInferType:                     {ast.KindInferType},
	typeTSTemplateLiteralType:           {ast.KindTemplateLiteralType},
	typeTSThisType:                      {ast.KindThisType},
	typeTSAbstractMethodDefinition:      {ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor, ast.KindConstructor},
	typeTSAbstractPropertyDefinition:    {ast.KindPropertyDeclaration},
	typeAccessorProperty:                {ast.KindPropertyDeclaration},
	typeTSAbstractAccessorProperty:      {ast.KindPropertyDeclaration},
	typeTSInstantiationExpression:       {ast.KindExpressionWithTypeArguments},
	typeTSNamespaceExportDeclaration:    {ast.KindNamespaceExportDeclaration},
}

// TriggerKinds returns a copy of the tsgo ast.Kind values that can produce a
// projected node of type id. When ok is false, callers must project eagerly.
func TriggerKinds(id TypeID) ([]ast.Kind, bool) {
	if int(id) >= len(triggerKindsTable) {
		return nil, false
	}
	kinds := triggerKindsTable[id]
	if len(kinds) == 0 {
		return nil, false
	}
	return slices.Clone(kinds), true
}
