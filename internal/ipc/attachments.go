package ipc

import (
	"errors"
	"unicode/utf8"
)

// AttachmentLimit is the preferred complete-text batch budget. Zero means
// that this channel keeps attachments inline. It exposes no platform handles.
func (c *Channel) AttachmentLimit() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed || c.sources == nil {
		return 0
	}
	return SourceSlotSize
}

// initializeSources runs on the read loop before invoking an inbound request
// handler. The first request fixes the backend, including an absent or failed
// mapping; later requests cannot change transport under running application code.
func (c *Channel) initializeSources(transport *TransportMetadata) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return c.closeErr
	}
	if c.sourcesInitialized {
		if transport != nil && transport.Mapping != nil {
			return errors.New("ipc: source mapping already initialized")
		}
		return nil
	}
	c.sourcesInitialized = true
	if transport == nil || transport.Mapping == nil {
		return nil
	}
	// Optional acceleration: unsupported versions/platforms, invalid handles,
	// and mapping failures preserve the complete inline attachment path.
	c.sources, _ = openSourcePool(*transport.Mapping)
	return nil
}

// attachTexts retains string headers, not copies of source bytes. The pool is
// the sole extra byte copy when sharing succeeds; every other text stays in
// its original inline form, including empty strings and invalid UTF-8.
func attachTexts(msg *Message, sources *sourcePool, texts []string) *SourceBatch {
	if len(texts) == 0 {
		return nil
	}
	msg.Attachments = make([]TextAttachment, len(texts))
	for i, text := range texts {
		msg.Attachments[i].Text = &text
	}
	if sources == nil {
		return nil
	}
	parts := make([]string, 0, len(texts))
	ranges := make([]*SourceRange, len(texts))
	var length uint32
	for i, text := range texts {
		if len(text) > SourceSlotSize-int(length) || !utf8.ValidString(text) {
			continue
		}
		size := uint32(len(text))
		parts = append(parts, text)
		ranges[i] = &SourceRange{Offset: length, Length: size}
		length += size
	}
	batch, ok := sources.store(parts)
	if !ok {
		return nil
	}
	for i, sourceRange := range ranges {
		if sourceRange != nil {
			msg.Attachments[i] = TextAttachment{Range: sourceRange}
		}
	}
	msg.Transport = &TransportMetadata{Batch: &batch}
	return &batch
}
