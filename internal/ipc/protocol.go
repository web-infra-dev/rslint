package ipc

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
)

// Go owns the shared-memory policy. The peer requests these settings over the
// existing channel before constructing its mapping; no build-time bindings are
// needed to keep the native reader's layout in sync.

const (
	FrameHeaderSize = 4
	MaxFrameSize    = 256 * 1024 * 1024

	MemoryVersion           uint32 = 1
	MemorySlotCount                = 16
	MemorySlotSize                 = 16 * 1024 * 1024
	MemoryHeaderSize               = 4096
	MemoryPublicationStride        = 4
	MemoryCapacity                 = MemoryHeaderSize + MemorySlotCount*MemorySlotSize
	MemoryInheritedFD              = 3
	MemoryMaxGeneration            = ^uint32(0)
)

// MemoryConfiguration is the runtime layout supplied to the native reader.
// Capacity is derived from HeaderSize + SlotCount*SlotSize by each consumer.
// Version identifies the publication algorithm, not a second set of defaults.
type MemoryConfiguration struct {
	Version           uint32 `json:"version"`
	SlotCount         uint32 `json:"slotCount"`
	SlotSize          uint32 `json:"slotSize"`
	HeaderSize        uint32 `json:"headerSize"`
	PublicationStride uint32 `json:"publicationStride"`
}

func memoryConfiguration() MemoryConfiguration {
	return MemoryConfiguration{
		Version:           MemoryVersion,
		SlotCount:         MemorySlotCount,
		SlotSize:          MemorySlotSize,
		HeaderSize:        MemoryHeaderSize,
		PublicationStride: MemoryPublicationStride,
	}
}

// MemoryMapping identifies an OS mapping owned by the native reader.
// Node reports the descriptor's child stdio position. Private bootstrap reserves
// the first extra stdio slot for this mapping.
type MemoryMapping struct {
	Version   uint32 `json:"version"`
	FD        int32  `json:"fd,omitempty"`
	Handle    string `json:"handle,omitempty"`
	ProcessID uint32 `json:"processId,omitempty"`
}

// MemoryBatch identifies a published region until its acknowledged release.
type MemoryBatch struct {
	Slot       uint32 `json:"slot"`
	Generation uint32 `json:"generation"`
	Length     uint32 `json:"length"`
}

// MemoryRange identifies complete bytes in the logical concatenation of a
// request's batches, using each batch's actual length rather than slot capacity.
type MemoryRange struct {
	Offset uint32 `json:"offset"`
	Length uint32 `json:"length"`
}

// AttachmentData carries exactly one complete inline text, inline byte buffer
// (JSON base64), or shared range. Pointers preserve empty inline values.
type AttachmentData struct {
	Text  *string      `json:"text,omitempty"`
	Bytes *[]byte      `json:"bytes,omitempty"`
	Range *MemoryRange `json:"range,omitempty"`
}

// Decode the union before typed pointers can collapse an explicit null into an
// absent field. Empty text and binary values remain distinct, valid variants.
func (a *AttachmentData) UnmarshalJSON(data []byte) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	count := 0
	for _, key := range []string{"text", "bytes", "range"} {
		if _, exists := fields[key]; exists {
			count++
		}
	}
	if count != 1 {
		return errors.New("ipc: attachment must contain exactly one of text, bytes, or range")
	}
	*a = AttachmentData{}
	for _, key := range []string{"text", "bytes"} {
		if raw, exists := fields[key]; exists {
			if trimmed := bytes.TrimSpace(raw); len(trimmed) == 0 || trimmed[0] != '"' {
				return errors.New("ipc: inline attachment must be a string")
			}
			var value string
			if err := json.Unmarshal(raw, &value); err != nil {
				return err
			}
			if key == "text" {
				a.Text = &value
				return nil
			}
			decoded, err := base64.StdEncoding.Strict().DecodeString(value)
			if err != nil || strings.ContainsAny(value, "\r\n") {
				return errors.New("ipc: invalid base64 attachment bytes")
			}
			a.Bytes = &decoded
			return nil
		}
	}
	raw := bytes.TrimSpace(fields["range"])
	if len(raw) == 0 || raw[0] != '{' {
		return errors.New("ipc: attachment range must be an object")
	}
	return json.Unmarshal(raw, &a.Range)
}

// TransportMetadata belongs to the envelope, never the application payload.
// Mapping is bootstrap-only. Released must match the request's complete ordered
// batch list; partial, reordered or unsolicited acknowledgements grant no reuse.
type TransportMetadata struct {
	SharedMemory uint32         `json:"sharedMemory,omitempty"`
	Mapping      *MemoryMapping `json:"mapping,omitempty"`
	Batches      []MemoryBatch  `json:"batches,omitempty"`
	Released     []MemoryBatch  `json:"released,omitempty"`
}

func (m *TransportMetadata) UnmarshalJSON(data []byte) error {
	type fields TransportMetadata
	*m = TransportMetadata{}
	decoded := struct {
		*fields
		Batches json.RawMessage `json:"batches"`
	}{fields: (*fields)(m)}
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	if decoded.Batches != nil {
		if err := json.Unmarshal(decoded.Batches, &m.Batches); err != nil {
			return err
		}
		if len(m.Batches) == 0 {
			return errors.New("ipc: memory batches must be a nonempty array")
		}
	}
	return nil
}
