// Package typeservice exposes caller-owned checker sessions without selecting
// projects or executing rules.
package typeservice

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/api/encoder"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

func prepareGeneratedSyntax(tree []byte, nodes, origins []*ast.Node, file *ast.SourceFile) {
	records := tree[binary.LittleEndian.Uint32(tree[encoder.HeaderOffsetNodes:]):]
	if origins != nil {
		restoreGeneratedLists(records, nodes, origins)
	}
	if file == nil {
		return
	}
	var positions *ast.PositionMap
	for offset := encoder.NodeSize; offset < len(records); offset += encoder.NodeSize {
		for _, field := range [...]int{encoder.NodeOffsetPos, encoder.NodeOffsetEnd} {
			data := records[offset+field:]
			pos := int32(binary.LittleEndian.Uint32(data))
			if pos < 0 {
				continue
			}
			if positions == nil {
				positions = file.GetPositionMap()
				if positions.IsAsciiOnly() {
					return
				}
			}
			binary.LittleEndian.PutUint32(data, uint32(positions.UTF8ToUTF16(int(pos))))
		}
	}
}

// A copied node keeps its parse origin. Restore its lists' comma flags and,
// when the builder kept the original span, their bounds before UTF-16 conversion.
func restoreGeneratedLists(records []byte, nodes, origins []*ast.Node) {
	var fields, generated, source []*ast.NodeList
	visitor := ast.NodeVisitor{Hooks: ast.NodeVisitorHooks{
		VisitNodes: func(list *ast.NodeList, _ *ast.NodeVisitor) *ast.NodeList {
			fields = append(fields, list)
			return list
		},
		VisitModifiers: func(modifiers *ast.ModifierList, _ *ast.NodeVisitor) *ast.ModifierList {
			var list *ast.NodeList
			if modifiers != nil && len(modifiers.Nodes) != 0 {
				list = &modifiers.NodeList
			}
			fields = append(fields, list)
			return modifiers
		},
	}}
	listFields := func(node *ast.Node, reuse []*ast.NodeList) []*ast.NodeList {
		fields = reuse[:0]
		node.VisitEachChild(&visitor)
		return fields
	}
	for index, origin := range origins {
		if origin == nil || index+1 == len(nodes) {
			continue
		}
		owner := nodes[index]
		if owner.Kind != origin.Kind || (owner.Pos() >= 0 && owner.Loc != origin.Loc) {
			continue
		}
		first := index + 1
		if int(binary.LittleEndian.Uint32(records[first*encoder.NodeSize+encoder.NodeOffsetParent:])) != index {
			continue
		}
		// NodeLists occupy nil slots in the encoder's index, in visitor order.
		slot := 0
		for child := first; child != 0; {
			record := records[child*encoder.NodeSize:]
			isList := nodes[child] == nil
			child = int(binary.LittleEndian.Uint32(record[encoder.NodeOffsetNext:]))
			if !isList {
				continue
			}
			if slot == 0 {
				generated = listFields(owner, generated)
				source = listFields(origin, source)
			}
			for generated[slot] == nil {
				slot++
			}
			original := source[slot]
			slot++
			if original == nil || int32(binary.LittleEndian.Uint32(record[encoder.NodeOffsetPos:])) >= 0 {
				continue
			}
			if owner.Pos() >= 0 {
				binary.LittleEndian.PutUint32(record[encoder.NodeOffsetPos:], uint32(original.Pos()))
				binary.LittleEndian.PutUint32(record[encoder.NodeOffsetEnd:], uint32(original.End()))
			}
			if original.HasTrailingComma() {
				binary.LittleEndian.PutUint32(record[encoder.NodeOffsetFlags:], 1)
			}
		}
	}
}

const (
	packetHandle uint32 = iota + 1
	packetBoolean
	packetText
	packetHandles
	packetType
	packetSymbol
	packetSignature
	packetNodes
	packetIndexInfos
	packetTypePredicate
	packetNamedSymbols
	packetDisplayParts
	packetJSDocTags
	packetTypeNode
	packetDiagnostics
	packetLiteral
	packetOptionalHandles
	packetStrings
	packetTypeDetails
	packetDeclaredMembers
)

// EncodePacket has a finite record layout: header/kind, fixed fields, then
// offset-addressed lists and an interned UTF-16 string pool. No arbitrary map,
// recursive value graph, reflection, or Go pointer enters a packet.
// These bytes are owned. A transport may lease them only for its call; a client
// copies fallback packets before retaining them in its disposable fact store.
func EncodePacket(result Result) ([]byte, error) {
	var kind, words uint32
	switch result.(type) {
	case Literal:
		kind, words = packetLiteral, 3
	case *DeclaredMembers:
		kind, words = packetDeclaredMembers, 4
	case *TypeDetails:
		kind, words = packetTypeDetails, 21
	case Strings:
		kind, words = packetStrings, 1
	case OptionalHandles:
		kind, words = packetOptionalHandles, 1
	case Handle:
		kind, words = packetHandle, 1
	case Boolean:
		kind, words = packetBoolean, 1
	case Text:
		kind, words = packetText, 1
	case Handles:
		kind, words = packetHandles, 1
	case *TypeFact:
		kind, words = packetType, 13
	case *SymbolFact:
		kind, words = packetSymbol, 7
	case *SignatureFact:
		kind, words = packetSignature, 5
	case NodeFacts:
		kind, words = packetNodes, 1
	case IndexInfos:
		kind, words = packetIndexInfos, 1
	case *TypePredicateFact:
		kind, words = packetTypePredicate, 4
	case NamedSymbols:
		kind, words = packetNamedSymbols, 1
	case DisplayParts:
		kind, words = packetDisplayParts, 1
	case JSDocTags:
		kind, words = packetJSDocTags, 1
	case *EncodedTypeNode:
		kind, words = packetTypeNode, 3
	case SuggestionDiagnostics:
		kind, words = packetDiagnostics, 1
	default:
		return nil, errors.New("unsupported semantic fact")
	}
	w := packetWriter{data: make([]byte, 8+words*4), strings: make(map[string]uint32)}
	copy(w.data, "RTS5")
	w.put(4, kind)
	switch fact := result.(type) {
	case Literal:
		w.put(8, fact.Kind)
		switch fact.Kind {
		case 1:
			w.put(12, w.text(fact.Text))
		case 2:
			binary.LittleEndian.PutUint64(w.data[12:], math.Float64bits(fact.Number))
		}
	case *DeclaredMembers:
		if fact != nil {
			w.put(8, w.ids(fact.properties))
			w.put(12, w.ids(fact.calls))
			w.put(16, w.ids(fact.constructs))
			w.put(20, w.indexInfos(fact.indices))
		}
	case *TypeDetails:
		w.put(8, w.node(fact.patternRef))
		w.put(12, w.node(fact.nodeRef))
		w.put(20, fact.regular)
		w.put(24, fact.base)
		w.put(28, fact.constraint)
		w.put(32, w.ids(fact.outer))
		w.put(36, w.ids(fact.local))
		w.put(40, w.ids(fact.elements))
		w.put(44, fact.minPlusOne)
		w.put(48, fact.fixedPlusOne)
		w.put(52, w.nodeList(fact.labelRefs))
		if fact.texts != nil {
			ids := make([]uint32, len(fact.texts))
			for i, v := range fact.texts {
				ids[i] = w.text(v)
			}
			w.put(56, w.ids(ids))
		}
		w.put(60, w.node(fact.rootRef))
		w.put(64, fact.rootCheck)
		w.put(68, fact.rootExtends)
		w.put(72, fact.rootAlias)
		if fact.distributive {
			w.put(76, 1)
		}
		w.put(80, w.ids(fact.infer))
		w.put(84, w.ids(fact.rootOuter))
		w.put(88, w.ids(fact.rootArguments))
	case Strings:
		ids := make([]uint32, len(fact))
		for i, value := range fact {
			ids[i] = w.text(value)
		}
		w.put(8, w.ids(ids))
	case OptionalHandles:
		w.put(8, w.ids(fact))
	case Handle:
		w.put(8, uint32(fact))
	case Boolean:
		if fact {
			w.put(8, 1)
		}
	case Text:
		w.put(8, w.text(string(fact)))
	case Handles:
		if fact == nil {
			fact = Handles{}
		}
		w.put(8, w.ids(fact))
	case *TypeFact:
		flags := make([]uint32, 0, len(fact.Flags))
		for _, flag := range fact.Flags {
			flags = append(flags, w.text(flag))
		}
		w.put(8, w.ids(flags))
		w.put(12, fact.Symbol)
		w.put(16, fact.AliasSymbol)
		if fact.IntrinsicName != "" {
			w.put(20, w.text(fact.IntrinsicName))
		}
		w.put(24, fact.Literal.Kind)
		switch fact.Literal.Kind {
		case 1:
			w.put(28, w.text(fact.Literal.Text))
		case 2:
			offset := uint32(len(w.data))
			w.data = binary.LittleEndian.AppendUint64(w.data, math.Float64bits(fact.Literal.Number))
			w.put(28, offset)
		case 3:
			offset := uint32(len(w.data))
			w.data = append(w.data, make([]byte, 8)...)
			if fact.Literal.Negative {
				w.put(offset, 1)
			}
			w.put(offset+4, w.text(fact.Literal.Text))
			w.put(28, offset)
		}
		w.put(32, w.ids(fact.TypeOperands))
		if fact.EscapedName != "" {
			w.put(36, w.text(fact.EscapedName))
		}
		if fact.ObjectFlags != nil {
			flags := make([]uint32, 0, len(fact.ObjectFlags))
			for _, flag := range fact.ObjectFlags {
				flags = append(flags, w.text(flag))
			}
			w.put(40, w.ids(flags))
		}
		w.put(44, fact.Target)
		w.put(48, w.ids(fact.AliasTypeArguments))
		w.put(52, w.ids(fact.TypeParameters))
		if fact.Tuple != nil {
			offset := uint32(len(w.data))
			w.data = append(w.data, make([]byte, 8)...)
			w.put(offset, fact.Tuple.CombinedFlags)
			if fact.Tuple.Readonly {
				w.put(offset+4, 1)
			}
			w.put(56, offset)
		}
	case *SymbolFact:
		w.put(8, w.text(fact.Name))
		w.put(12, w.text(fact.EscapedName))
		w.put(16, fact.Flags)
		if fact.Declarations != nil {
			offset := uint32(len(w.data))
			w.data = append(w.data, make([]byte, 4+len(fact.Declarations)*8)...)
			w.put(offset, uint32(len(fact.Declarations)))
			for i, node := range fact.Declarations {
				at := offset + 4 + uint32(i)*8
				w.put(at, w.text(node.File))
				w.put(at+4, node.Index)
			}
			w.put(20, offset)
		}
		w.put(24, w.node(fact.ValueDeclaration))
		if fact.Flags&uint32(ast.SymbolFlagsTransient) != 0 {
			w.put(28, fact.CheckFlags+1)
		}
		var tables uint32
		if fact.HasMembers {
			tables |= 1
		}
		if fact.HasExports {
			tables |= 2
		}
		if fact.HasGlobalExports {
			tables |= 4
		}
		w.put(32, tables)
	case *EncodedTypeNode:
		w.put(8, uint32(len(fact.tree)))
		w.data = append(w.data, fact.tree...)
		w.put(12, w.nodeList(fact.origins))
		w.put(16, w.ids(fact.emitFlags))
	case SuggestionDiagnostics:
		w.put(8, w.diagnostics(fact, make(map[string]uint32)))
	case NodeFacts:
		offset := uint32(len(w.data))
		w.data = append(w.data, make([]byte, 4+len(fact)*28)...)
		w.put(offset, uint32(len(fact)))
		for i, node := range fact {
			at := offset + 4 + uint32(i)*28
			w.put(at, w.node(&node.Node))
			w.put(at+4, w.text(node.Kind))
			w.put(at+8, node.Flags)
			w.put(at+12, w.node(node.Parent))
			w.put(at+16, w.node(node.Name))
			if node.Text != nil {
				w.put(at+20, w.text(*node.Text))
			}
			w.put(at+24, node.ModuleKeyword)
		}
		w.put(8, offset)
	case *SignatureFact:
		w.put(8, w.ids(fact.Parameters))
		w.put(12, w.node(fact.Declaration))
		w.put(16, w.ids(fact.TypeParameters))
		w.put(20, fact.ThisParameter)
		w.put(24, uint32(fact.MinArgumentCount))
	case IndexInfos:
		w.put(8, w.indexInfos(fact))
	case *TypePredicateFact:
		if fact != nil {
			w.put(8, fact.Kind+1)
			if fact.ParameterName != nil {
				w.put(12, w.text(*fact.ParameterName))
				w.put(16, uint32(fact.ParameterIndex))
			}
			w.put(20, fact.Type)
		}
	case NamedSymbols:
		offset := uint32(len(w.data))
		w.data = append(w.data, make([]byte, 4+len(fact)*8)...)
		w.put(offset, uint32(len(fact)))
		for i, symbol := range fact {
			at := offset + 4 + uint32(i)*8
			w.put(at, symbol.ID)
			w.put(at+4, w.text(symbol.Name))
		}
		w.put(8, offset)
	case DisplayParts:
		w.put(8, w.displayParts(fact))
	case JSDocTags:
		offset := uint32(len(w.data))
		w.data = append(w.data, make([]byte, 4+len(fact)*8)...)
		w.put(offset, uint32(len(fact)))
		for i, tag := range fact {
			at := offset + 4 + uint32(i)*8
			w.put(at, w.text(tag.Name))
			w.put(at+4, w.displayParts(tag.Text))
		}
		w.put(8, offset)
	}
	if uint64(len(w.data)) > math.MaxUint32 {
		return nil, errors.New("semantic fact exceeds packet limit")
	}
	return w.data, nil
}

type packetWriter struct {
	data    []byte
	strings map[string]uint32
}

func (w *packetWriter) indexInfos(infos IndexInfos) uint32 {
	offset := uint32(len(w.data))
	w.data = append(w.data, make([]byte, 4+len(infos)*24)...)
	w.put(offset, uint32(len(infos)))
	for i, info := range infos {
		at := offset + 4 + uint32(i)*24
		w.put(at, info.ID)
		w.put(at+4, info.KeyType)
		w.put(at+8, info.Type)
		if info.IsReadonly {
			w.put(at+12, 1)
		}
		w.put(at+16, w.node(info.Declaration))
		w.put(at+20, w.nodeList(info.Components))
	}
	return offset
}

func (w *packetWriter) diagnostics(facts SuggestionDiagnostics, configs map[string]uint32) uint32 {
	if len(facts) == 0 {
		return 0
	}
	at := uint32(len(w.data))
	w.data = append(w.data, make([]byte, 4+40*len(facts))...)
	w.put(at, uint32(len(facts)))
	for i, d := range facts {
		pos := at + 4 + uint32(i)*40
		if d.Config != nil {
			offset := configs[d.Config.ID]
			if offset == 0 {
				offset = w.diagnosticConfig(d.Config)
				configs[d.Config.ID] = offset
			}
			w.put(pos, offset)
		} else if d.File != "" {
			w.put(pos, w.text(d.File))
		}
		w.put(pos+4, d.Start)
		w.put(pos+8, d.Length)
		w.put(pos+12, d.Category)
		w.put(pos+16, d.Code)
		w.put(pos+20, w.text(d.Message))
		if d.Source != "" {
			w.put(pos+24, w.text(d.Source))
		}
		var flags uint32
		if d.Deprecated {
			flags |= 1
		}
		if d.Unnecessary {
			flags |= 2
		}
		if d.Config != nil {
			flags |= 4
		}
		w.put(pos+28, flags)
		w.put(pos+32, w.diagnostics(d.Chain, configs))
		w.put(pos+36, w.diagnostics(d.Related, configs))
	}
	return at
}

func (w *packetWriter) diagnosticConfig(info *configSourceInfo) uint32 {
	at := uint32(len(w.data))
	w.data = append(w.data, make([]byte, 16)...)
	w.put(at, w.text(info.Name))
	w.put(at+4, w.text(string(info.Path)))
	w.put(at+12, w.text(info.ID))
	if len(info.ExtendedSourceFiles) != 0 {
		files := make([]uint32, len(info.ExtendedSourceFiles))
		for i, file := range info.ExtendedSourceFiles {
			files[i] = w.text(file)
		}
		w.put(at+8, w.ids(files))
	}
	return at
}

func (w *packetWriter) put(at, value uint32) { binary.LittleEndian.PutUint32(w.data[at:], value) }
func (w *packetWriter) text(text string) uint32 {
	if offset, ok := w.strings[text]; ok {
		return offset
	}
	offset := uint32(len(w.data))
	units := ecmascript.StringCodeUnits(text)
	w.data = binary.LittleEndian.AppendUint32(w.data, uint32(len(units)))
	for _, unit := range units {
		w.data = binary.LittleEndian.AppendUint16(w.data, unit)
	}
	w.strings[text] = offset
	return offset
}
func (w *packetWriter) ids(ids []uint32) uint32 {
	if ids == nil {
		return 0
	}
	offset := uint32(len(w.data))
	w.data = binary.LittleEndian.AppendUint32(w.data, uint32(len(ids)))
	for _, id := range ids {
		w.data = binary.LittleEndian.AppendUint32(w.data, id)
	}
	return offset
}
func (w *packetWriter) displayParts(parts DisplayParts) uint32 {
	if parts == nil {
		return 0
	}
	offset := uint32(len(w.data))
	w.data = append(w.data, make([]byte, 4+len(parts)*12)...)
	w.put(offset, uint32(len(parts)))
	for i, part := range parts {
		at := offset + 4 + uint32(i)*12
		w.put(at, w.text(part.Text))
		w.put(at+4, w.text(part.Kind))
		if target := part.Target; target != nil {
			span := uint32(len(w.data))
			w.data = append(w.data, make([]byte, 12)...)
			w.put(span, w.text(target.File))
			w.put(span+4, target.Start)
			w.put(span+8, target.Length)
			w.put(at+8, span)
		}
	}
	return offset
}
func (w *packetWriter) nodeList(nodes []*NodeRef) uint32 {
	if nodes == nil {
		return 0
	}
	ids := make([]uint32, len(nodes))
	for i, n := range nodes {
		ids[i] = w.node(n)
	}
	return w.ids(ids)
}
func (w *packetWriter) node(node *NodeRef) uint32 {
	if node == nil {
		return 0
	}
	offset := uint32(len(w.data))
	w.data = append(w.data, make([]byte, 8)...)
	// Config nodes carry the identity of the parse after the signed index.
	index := int32(node.Index)
	if node.Config != "" {
		w.data = append(w.data, make([]byte, 4)...)
		w.put(offset+8, w.text(node.Config))
		index = -index
	}
	w.put(offset, w.text(node.File))
	w.put(offset+4, uint32(index))
	return offset
}

// EncodeReply keeps direct immutable object records beside the scalar answer.
// Packet slices use independent offset origins and can be retained or evicted
// separately by the worker. No recursive graph or shared native pointer exists.
func EncodeReply(result Result, records []Record) ([]byte, error) {
	packet, err := EncodePacket(result)
	if err != nil || len(records) == 0 {
		return packet, err
	}
	data := make([]byte, 0, 12+len(records)*16+len(packet))
	data = append(data, make([]byte, 12+len(records)*16)...)
	copy(data, "RTB1")
	binary.LittleEndian.PutUint32(data[4:], uint32(len(packet)))
	binary.LittleEndian.PutUint32(data[8:], uint32(len(records)))
	data = append(data, packet...)
	for i, record := range records {
		encoded, err := EncodePacket(record.Fact)
		if err != nil {
			return nil, err
		}
		at := 12 + i*16
		var kind uint32
		switch record.Op {
		case "typeRecord":
			kind = 1
		case "symbolRecord":
			kind = 2
		case "signatureRecord":
			kind = 3
		default:
			return nil, errors.New("invalid supplemental fact")
		}
		binary.LittleEndian.PutUint32(data[at:], kind)
		binary.LittleEndian.PutUint32(data[at+4:], record.ID)
		binary.LittleEndian.PutUint32(data[at+8:], uint32(len(data)))
		binary.LittleEndian.PutUint32(data[at+12:], uint32(len(encoded)))
		data = append(data, encoded...)
	}
	return data, nil
}
