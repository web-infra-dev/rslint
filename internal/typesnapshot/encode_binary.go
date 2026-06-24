package typesnapshot

import (
	"encoding/binary"
	"sort"
)

// Binary wire format for a Snapshot. Little-endian throughout, matching both
// the IPC frame's LE length prefix and the JS-side DataView reader (which uses
// littleEndian=true). The format is RANDOM-ACCESS by design:
//
//   - node2type is a fixed-width array sorted by (tokenStart, end); the worker
//     binary-searches it for getTypeAtLocation instead of building a Map of
//     every value node (the eager Map over ~276k entries on a 54k-line file was
//     the M0.5 bottleneck — see snapshot.go's depth-policy note).
//   - types is a fixed-width index sorted by type-id pointing into a blob
//     section; the worker decodes a type block lazily, only for the type-ids a
//     rule actually touches (internType is already lazy — this just moves its
//     data source from a JS object to an offset read).
//
// This replaces the JSON array-of-objects that forced an eager parse +
// structuredClone + full Map build on the worker. Go's encoding/json base64s a
// []byte field automatically, so the bytes ride the existing JSON IPC frame
// (frame.go has no binary-payload channel) without touching the transport.
//
// Layout:
//
//	header (24 bytes):
//	  magic         u32   // "RSNP" — guards against a stale/foreign payload
//	  version       u32
//	  n2tCount      u32
//	  typCount      u32
//	  primUndefined i32
//	  primNull      i32
//	node2type (n2tCount × 12 bytes), sorted by (tokenStart, end):
//	  tokenStart i32, end i32, typeId i32
//	typesIndex (typCount × 12 bytes), sorted by typeId:
//	  typeId i32, blobOffset u32 (from blob start), blobLen u32
//	blob (concatenated type blocks); per block:
//	  flags i32
//	  flagBits u8            // bit0 = isArray, bit1 = isTuple
//	  nameLen u32, name [nameLen]byte (utf8)
//	  memberCount u32,  memberTypes   [memberCount]i32
//	  typeArgCount u32, typeArgs      [typeArgCount]i32
//	  callSigCount u32, callSigReturns[callSigCount]i32
const (
	snapshotMagic   uint32 = 0x52534E50 // "RSNP"
	snapshotVersion uint32 = 1
	headerSize             = 24
	n2tEntrySize           = 12
	idxEntrySize           = 12
)

// EncodeBinary serializes snap into the random-access binary wire format above.
func EncodeBinary(snap Snapshot) []byte {
	// node2type sorted by (tokenStart, end) so the worker can binary-search it.
	n2t := make([]NodeTypeEntry, len(snap.Node2Type))
	copy(n2t, snap.Node2Type)
	sort.Slice(n2t, func(i, j int) bool {
		if n2t[i].TokenStart != n2t[j].TokenStart {
			return n2t[i].TokenStart < n2t[j].TokenStart
		}
		if n2t[i].End != n2t[j].End {
			return n2t[i].End < n2t[j].End
		}
		// Build dedupes (tokenStart,end) collisions, so equal keys shouldn't reach
		// here; tie-break on typeId anyway for a deterministic order (defensive).
		return n2t[i].TypeID < n2t[j].TypeID
	})

	// type-ids sorted for the binary-searchable index.
	ids := make([]int, 0, len(snap.Types))
	for id := range snap.Types {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	// Build the blob first so each type's (offset, len) is known for the index.
	var blob []byte
	type idxEntry struct{ id, off, length int }
	index := make([]idxEntry, 0, len(ids))
	for _, id := range ids {
		off := len(blob)
		blob = appendTypeBlock(blob, snap.Types[id])
		index = append(index, idxEntry{id: id, off: off, length: len(blob) - off})
	}

	out := make([]byte, headerSize, headerSize+len(n2t)*n2tEntrySize+len(index)*idxEntrySize+len(blob))
	binary.LittleEndian.PutUint32(out[0:], snapshotMagic)
	binary.LittleEndian.PutUint32(out[4:], snapshotVersion)
	binary.LittleEndian.PutUint32(out[8:], uint32(len(n2t)))
	binary.LittleEndian.PutUint32(out[12:], uint32(len(index)))
	binary.LittleEndian.PutUint32(out[16:], uint32(int32(snap.PrimTypes.Undefined)))
	binary.LittleEndian.PutUint32(out[20:], uint32(int32(snap.PrimTypes.Null)))

	for _, e := range n2t {
		out = binary.LittleEndian.AppendUint32(out, uint32(int32(e.TokenStart)))
		out = binary.LittleEndian.AppendUint32(out, uint32(int32(e.End)))
		out = binary.LittleEndian.AppendUint32(out, uint32(int32(e.TypeID)))
	}
	for _, e := range index {
		out = binary.LittleEndian.AppendUint32(out, uint32(int32(e.id)))
		out = binary.LittleEndian.AppendUint32(out, uint32(e.off))
		out = binary.LittleEndian.AppendUint32(out, uint32(e.length))
	}
	out = append(out, blob...)
	return out
}

func appendTypeBlock(b []byte, t TypeBlock) []byte {
	b = binary.LittleEndian.AppendUint32(b, uint32(int32(t.Flags)))
	var bits byte
	if t.IsArray {
		bits |= 1
	}
	if t.IsTuple {
		bits |= 2
	}
	b = append(b, bits)
	name := []byte(t.Name)
	b = binary.LittleEndian.AppendUint32(b, uint32(len(name)))
	b = append(b, name...)
	b = appendIDs(b, t.MemberTypes)
	b = appendIDs(b, t.TypeArgs)
	b = appendIDs(b, t.CallSigReturns)
	return b
}

func appendIDs(b []byte, ids []TypeID) []byte {
	b = binary.LittleEndian.AppendUint32(b, uint32(len(ids)))
	for _, id := range ids {
		b = binary.LittleEndian.AppendUint32(b, uint32(int32(id)))
	}
	return b
}

// DecodeBinary reconstructs a Snapshot from EncodeBinary output. The production
// reader is the JS worker (parser-services-from-snapshot.ts); this Go decoder
// exists so the round-trip is verifiable in a Go test and so the byte layout
// has one authoritative reference implementation. It is defensive (bounds-
// checked) and returns ok=false on any truncation rather than panicking.
func DecodeBinary(data []byte) (Snapshot, bool) {
	s := Snapshot{Types: map[TypeID]TypeBlock{}}
	if len(data) < headerSize || binary.LittleEndian.Uint32(data[0:]) != snapshotMagic {
		return s, false
	}
	if binary.LittleEndian.Uint32(data[4:]) != snapshotVersion {
		return s, false
	}
	n2tCount := int(binary.LittleEndian.Uint32(data[8:]))
	typCount := int(binary.LittleEndian.Uint32(data[12:]))
	s.PrimTypes.Undefined = int(int32(binary.LittleEndian.Uint32(data[16:])))
	s.PrimTypes.Null = int(int32(binary.LittleEndian.Uint32(data[20:])))

	off := headerSize
	if off+n2tCount*n2tEntrySize+typCount*idxEntrySize > len(data) {
		return s, false
	}
	for i := 0; i < n2tCount; i++ {
		s.Node2Type = append(s.Node2Type, NodeTypeEntry{
			TokenStart: int(int32(binary.LittleEndian.Uint32(data[off:]))),
			End:        int(int32(binary.LittleEndian.Uint32(data[off+4:]))),
			TypeID:     int(int32(binary.LittleEndian.Uint32(data[off+8:]))),
		})
		off += n2tEntrySize
	}

	type idxEntry struct{ id, boff, blen int }
	index := make([]idxEntry, typCount)
	for i := 0; i < typCount; i++ {
		index[i] = idxEntry{
			id:   int(int32(binary.LittleEndian.Uint32(data[off:]))),
			boff: int(binary.LittleEndian.Uint32(data[off+4:])),
			blen: int(binary.LittleEndian.Uint32(data[off+8:])),
		}
		off += idxEntrySize
	}

	blobStart := off
	for _, e := range index {
		if blobStart+e.boff+e.blen > len(data) {
			return s, false
		}
		block, ok := decodeTypeBlock(data[blobStart+e.boff : blobStart+e.boff+e.blen])
		if !ok {
			return s, false
		}
		block.ID = e.id
		s.Types[e.id] = block
	}
	return s, true
}

func decodeTypeBlock(b []byte) (TypeBlock, bool) {
	var t TypeBlock
	if len(b) < 5 {
		return t, false
	}
	t.Flags = int(int32(binary.LittleEndian.Uint32(b[0:])))
	bits := b[4]
	t.IsArray = bits&1 != 0
	t.IsTuple = bits&2 != 0
	p := 5
	nameLen := int(binary.LittleEndian.Uint32(b[p:]))
	p += 4
	if p+nameLen > len(b) {
		return t, false
	}
	t.Name = string(b[p : p+nameLen])
	p += nameLen
	var ok bool
	if t.MemberTypes, p, ok = readIDs(b, p); !ok {
		return t, false
	}
	if t.TypeArgs, p, ok = readIDs(b, p); !ok {
		return t, false
	}
	if t.CallSigReturns, p, ok = readIDs(b, p); !ok {
		return t, false
	}
	return t, true
}

func readIDs(b []byte, p int) ([]TypeID, int, bool) {
	if p+4 > len(b) {
		return nil, p, false
	}
	n := int(binary.LittleEndian.Uint32(b[p:]))
	p += 4
	if p+n*4 > len(b) {
		return nil, p, false
	}
	if n == 0 {
		return nil, p, true
	}
	ids := make([]TypeID, n)
	for i := 0; i < n; i++ {
		ids[i] = int(int32(binary.LittleEndian.Uint32(b[p:])))
		p += 4
	}
	return ids, p, true
}
