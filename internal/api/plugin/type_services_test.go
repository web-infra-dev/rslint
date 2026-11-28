package plugin

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/typeservice"
)

func roundTripIPCResponse(t *testing.T, payload any) *ipc.Message {
	t.Helper()
	inR, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close() })
	var out bytes.Buffer
	receiptCh := make(chan *ipc.ResponseReceipt, 1)
	ch := ipc.NewChannel(inR, &out)
	ch.SetInboundHandler(func(context.Context, *ipc.Message) (any, error) {
		tracked, receipt := ipc.TrackResponse(payload)
		receiptCh <- receipt
		return tracked, nil
	})
	ch.Start()
	t.Cleanup(func() { _ = ch.Close() })
	req, err := ipc.NewMessage("typeService", 1, map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	if err := ipc.WriteFrame(inW, req); err != nil {
		t.Fatal(err)
	}
	receipt := <-receiptCh
	<-receipt.Done()
	if err := receipt.Err(); err != nil {
		t.Fatalf("response receipt: %v", err)
	}
	if err := ch.Close(); err != nil {
		t.Fatal(err)
	}
	msg, err := ipc.ReadFrame(bufio.NewReader(bytes.NewReader(out.Bytes())))
	if err != nil {
		t.Fatalf("read response frame: %v", err)
	}
	return msg
}

func TestHandleTypeServicesDecodeAndClosedStatus(t *testing.T) {
	badMsg := &ipc.Message{Kind: "typeService", ID: 1, Data: []byte(`{invalid`)}
	if _, err := HandleTypeServices(context.Background(), badMsg); err == nil || !strings.Contains(err.Error(), "decode typeService request") {
		t.Fatalf("invalid JSON error = %v, want decode error", err)
	}

	reqMsg, err := ipc.NewMessage("typeService", 2, linter.EslintPluginTypeRequest{
		Session:    "nonexistent-session",
		Generation: 1,
		Open:       true,
	})
	if err != nil {
		t.Fatal(err)
	}
	res, err := HandleTypeServices(context.Background(), reqMsg)
	if err != nil {
		t.Fatalf("closed session returned transport error: %v", err)
	}
	frame := roundTripIPCResponse(t, res)
	var status struct {
		Status string `json:"status"`
	}
	if err := frame.Decode(&status); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if status.Status != "closed" {
		t.Fatalf("status = %q, want closed", status.Status)
	}
}

func TestEncodeTypeServiceWireResponseInlineAndAttachments(t *testing.T) {
	t.Run("small packets stay inline without attachments", func(t *testing.T) {
		small := bytes.Repeat([]byte{0x42}, minSharedTypeServiceBytes-1)
		encoded := encodeTypeServiceWireResponse(&linter.EslintPluginTypeResult{
			Packets: [][]byte{small},
		})
		frame := roundTripIPCResponse(t, encoded)
		if len(frame.Attachments) != 0 {
			t.Fatalf("expected 0 attachments for small inline packet, got %d", len(frame.Attachments))
		}
		var wire typeServiceWireResponse
		if err := frame.Decode(&wire); err != nil {
			t.Fatalf("decode inline response: %v", err)
		}
		if len(wire.Packets) != 1 || !bytes.Equal(wire.Packets[0], small) || len(wire.PacketAttachments) != 0 {
			t.Fatalf("unexpected inline wire response: packets=%d attachments=%v", len(wire.Packets), wire.PacketAttachments)
		}
	})

	t.Run("large packets, projection, and source use binary attachments", func(t *testing.T) {
		large := bytes.Repeat([]byte{0x7a}, minSharedTypeServiceBytes)
		projTree := []byte{1, 2, 3, 4}
		srcTree := []byte{5, 6, 7}
		projection := &linter.EslintPluginTypeProjection{Tree: projTree}
		projection.Origins = []uint32{7, 11}
		projection.Reverse = map[uint32]uint32{11: 1}
		encoded := encodeTypeServiceWireResponse(&linter.EslintPluginTypeResult{
			Descriptor: &typeservice.Descriptor{Generation: 3},
			Projection: projection,
			Packets:    [][]byte{large},
			Source:     &linter.EslintPluginTypeSource{FileName: "/a.ts", Path: "/a.ts", Tree: srcTree},
		})
		frame := roundTripIPCResponse(t, encoded)
		if len(frame.Attachments) != 3 {
			t.Fatalf("attachments = %d, want 3", len(frame.Attachments))
		}
		var wire typeServiceWireResponse
		if err := frame.Decode(&wire); err != nil {
			t.Fatalf("decode attachment response: %v", err)
		}
		var rawWire map[string]any
		if err := json.Unmarshal(frame.Data, &rawWire); err != nil {
			t.Fatalf("unmarshal raw wire response: %v", err)
		}
		projJSON, ok := rawWire["projection"].(map[string]any)
		if !ok {
			t.Fatalf("projection JSON object missing: %+v", rawWire)
		}
		originsJSON, ok := projJSON["origins"].([]any)
		if !ok || len(originsJSON) != 2 || originsJSON[0] != float64(7) || originsJSON[1] != float64(11) {
			t.Fatalf("projection.origins = %#v, want [7 11]", projJSON["origins"])
		}
		reverseJSON, ok := projJSON["reverse"].(map[string]any)
		if !ok || reverseJSON["11"] != float64(1) {
			t.Fatalf("projection.reverse = %#v, want {\"11\": 1}", projJSON["reverse"])
		}
		if _, hasTree := projJSON["tree"]; hasTree {
			t.Fatalf("projection JSON unexpectedly included inline tree: %+v", projJSON)
		}
		if _, hasWrapper := projJSON["EslintPluginTypeProjectionMeta"]; hasWrapper {
			t.Fatalf("projection JSON wrapped embedded meta struct: %+v", projJSON)
		}
		if wire.Descriptor == nil || wire.Descriptor.Generation != 3 {
			t.Fatalf("descriptor = %+v, want generation 3", wire.Descriptor)
		}
		if wire.Projection == nil || wire.Projection.TreeAttachment != 0 || frame.Attachments[0].Bytes == nil || !bytes.Equal(*frame.Attachments[0].Bytes, projTree) {
			t.Fatalf("projection attachment mismatch: %+v", wire.Projection)
		}
		if len(wire.Packets) != 0 || len(wire.PacketAttachments) != 1 || wire.PacketAttachments[0] != 1 || frame.Attachments[1].Bytes == nil || !bytes.Equal(*frame.Attachments[1].Bytes, large) {
			t.Fatalf("packet attachment mismatch: %+v", wire.PacketAttachments)
		}
		if wire.Source == nil || wire.Source.TreeAttachment == nil || *wire.Source.TreeAttachment != 2 || frame.Attachments[2].Bytes == nil || !bytes.Equal(*frame.Attachments[2].Bytes, srcTree) {
			t.Fatalf("source attachment mismatch: %+v", wire.Source)
		}
	})

	t.Run("metadata-only source omits tree attachment", func(t *testing.T) {
		encoded := encodeTypeServiceWireResponse(&linter.EslintPluginTypeResult{
			Source: &linter.EslintPluginTypeSource{FileName: "/meta.ts", Path: "/meta.ts"},
		})
		frame := roundTripIPCResponse(t, encoded)
		if len(frame.Attachments) != 0 {
			t.Fatalf("expected 0 attachments for metadata-only source, got %d", len(frame.Attachments))
		}
		var wire typeServiceWireResponse
		if err := frame.Decode(&wire); err != nil {
			t.Fatalf("decode metadata source: %v", err)
		}
		if wire.Source == nil || wire.Source.FileName != "/meta.ts" || wire.Source.TreeAttachment != nil {
			t.Fatalf("unexpected source reply: %+v", wire.Source)
		}
	})
}
