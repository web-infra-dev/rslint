// Package typeservice exposes caller-owned checker sessions without selecting
// projects or executing rules.
package typeservice

import (
	"encoding/binary"
	"errors"
	"math"

	"github.com/web-infra-dev/rslint/internal/utils/ecmascript"
)

const (
	packetHandle uint32 = iota + 1
	packetBoolean
	packetText
	packetHandles
	packetType
	packetSymbol
	packetSignature
)

// EncodePacket has a finite record layout: header/kind, fixed fields, then
// offset-addressed lists and an interned UTF-16 string pool. No arbitrary map,
// recursive value graph, reflection, or Go pointer enters a packet.
// These bytes are owned. A transport may lease them only for its call; a client
// copies fallback packets before retaining them in its disposable fact store.
func EncodePacket(result Result) ([]byte, error) {
	var kind, words uint32
	switch result.(type) {
	case Handle:
		kind, words = packetHandle, 1
	case Boolean:
		kind, words = packetBoolean, 1
	case Text:
		kind, words = packetText, 1
	case Handles:
		kind, words = packetHandles, 1
	case *TypeFact:
		kind, words = packetType, 8
	case *SymbolFact:
		kind, words = packetSymbol, 5
	case *SignatureFact:
		kind, words = packetSignature, 2
	default:
		return nil, errors.New("unsupported semantic fact")
	}
	w := packetWriter{data: make([]byte, 8+words*4), strings: make(map[string]uint32)}
	copy(w.data, "RTS3")
	w.put(4, kind)
	switch fact := result.(type) {
	case Handle:
		w.put(8, uint32(fact))
	case Boolean:
		if fact {
			w.put(8, 1)
		}
	case Text:
		w.put(8, w.text(string(fact)))
	case Handles:
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
		w.put(32, w.ids(fact.Types))
		if fact.EscapedName != "" {
			w.put(36, w.text(fact.EscapedName))
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
	case *SignatureFact:
		w.put(8, w.ids(fact.Parameters))
		w.put(12, w.node(fact.Declaration))
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
func (w *packetWriter) node(node *NodeRef) uint32 {
	if node == nil {
		return 0
	}
	offset := uint32(len(w.data))
	w.data = append(w.data, make([]byte, 8)...)
	w.put(offset, w.text(node.File))
	w.put(offset+4, node.Index)
	return offset
}
