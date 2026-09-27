package ipc

// Go owns the source-storage policy. The peer requests these settings over the
// existing channel before constructing its mapping; no build-time bindings are
// needed to keep the native reader's layout in sync.

const (
	FrameHeaderSize = 4
	MaxFrameSize    = 256 * 1024 * 1024

	SourceVersion           uint32 = 1
	SourceSlotCount                = 16
	SourceSlotSize                 = 16 * 1024 * 1024
	SourceHeaderSize               = 4096
	SourcePublicationStride        = 4
	SourceCapacity                 = SourceHeaderSize + SourceSlotCount*SourceSlotSize
	SourceInheritedFD              = 3
	SourceMaxGeneration            = ^uint32(0)
)

// SourceConfiguration is the runtime layout supplied to the native reader.
// Capacity is derived from HeaderSize + SlotCount*SlotSize by each consumer.
// Version identifies the publication algorithm, not a second set of defaults.
type SourceConfiguration struct {
	Version           uint32 `json:"version"`
	SlotCount         uint32 `json:"slotCount"`
	SlotSize          uint32 `json:"slotSize"`
	HeaderSize        uint32 `json:"headerSize"`
	PublicationStride uint32 `json:"publicationStride"`
}

func sourceConfiguration() SourceConfiguration {
	return SourceConfiguration{
		Version:           SourceVersion,
		SlotCount:         SourceSlotCount,
		SlotSize:          SourceSlotSize,
		HeaderSize:        SourceHeaderSize,
		PublicationStride: SourcePublicationStride,
	}
}

// SourceDescriptor identifies an OS mapping owned by the native reader.
// The Node transport replaces a local FD with SourceInheritedFD for the child.
type SourceDescriptor struct {
	Version   uint32 `json:"version"`
	FD        int32  `json:"fd,omitempty"`
	Handle    string `json:"handle,omitempty"`
	ProcessID uint32 `json:"processId,omitempty"`
}

// SourceBatch identifies a published region until its acknowledged release.
type SourceBatch struct {
	Slot       uint32 `json:"slot"`
	Generation uint32 `json:"generation"`
	Length     uint32 `json:"length"`
}

// SourceRange identifies one complete source within a published batch.
type SourceRange struct {
	Offset uint32 `json:"offset"`
	Length uint32 `json:"length"`
}

// TextAttachment carries exactly one complete inline text or published range.
// A pointer preserves empty inline strings on the wire.
type TextAttachment struct {
	Text  *string      `json:"text,omitempty"`
	Range *SourceRange `json:"range,omitempty"`
}

// TransportMetadata belongs to the envelope, never the application payload.
// Mapping is bootstrap-only; Released acknowledges this request's exact Batch.
type TransportMetadata struct {
	Mapping  *SourceDescriptor `json:"mapping,omitempty"`
	Batch    *SourceBatch      `json:"batch,omitempty"`
	Released *SourceBatch      `json:"released,omitempty"`
}
