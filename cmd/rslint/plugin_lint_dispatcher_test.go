//go:build !js

package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/ipc/sharedsource"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// Wire tests record the storage contract; pool ownership/concurrency is tested
// in sharedsource, and engine integration exercises real OS mappings.
type recordedPluginSources struct {
	parts       []string
	generation  uint32
	unavailable bool
	released    []sharedsource.Batch
	closed      bool
}

func (s *recordedPluginSources) Store(parts []string) (sharedsource.Batch, bool) {
	if s.unavailable || len(parts) == 0 {
		return sharedsource.Batch{}, false
	}
	s.parts = parts
	s.generation++
	batch := sharedsource.Batch{Generation: s.generation}
	for _, part := range parts {
		batch.Length += uint32(len(part))
	}
	return batch, true
}
func (s *recordedPluginSources) Release(batch sharedsource.Batch) {
	s.released = append(s.released, batch)
}
func (s *recordedPluginSources) Close() error { s.closed = true; return nil }

func TestPluginSourceSnapshots(t *testing.T) {
	store := &recordedPluginSources{}
	dispatcher := &pluginLintDispatcher{sources: store}
	texts := []string{"", "\ufeff" + "const café = '😀';\r\n// \x00", string([]byte{0xff}), strings.Repeat("x", sharedsource.SlotSize+1)}
	files := make([]linter.EslintPluginLintFile, len(texts)+1)
	for i := range texts {
		files[i].Text = &texts[i]
	}
	req := linter.EslintPluginLintRequest{Files: files}
	wire, batch := dispatcher.encode(req)
	encoded := wire.(pluginSourceRequest)
	if batch == nil || batch.Length != uint32(len(texts[1])) {
		t.Fatalf("wrong snapshot length: %+v", batch)
	}
	for i := range 2 {
		file := encoded.Files[i]
		if file.Text != nil || file.SourceRange == nil || req.Files[i].Text == nil {
			t.Fatalf("source lost or original request mutated: %+v", file)
		}
		r := file.SourceRange
		if got := strings.Join(store.parts, "")[r.Offset : r.Offset+r.Length]; got != texts[i] {
			t.Fatalf("source bytes changed: %q", got)
		}
	}
	for i := 2; i < len(files); i++ {
		if encoded.Files[i].Text != req.Files[i].Text || encoded.Files[i].SourceRange != nil {
			t.Fatalf("inline fallback changed file %d", i)
		}
	}
	store.unavailable = true
	wire, batch = dispatcher.encode(req)
	if batch != nil {
		t.Fatal("exhausted storage did not fall back")
	}
	inline := wire.(linter.EslintPluginLintRequest)
	for i := range req.Files {
		if inline.Files[i].Text != req.Files[i].Text {
			t.Fatal("fallback lost the original snapshot")
		}
	}
}

func TestPluginSourceReleaseAcknowledgement(t *testing.T) {
	for _, mode := range []string{"release", "missing", "wrong-slot", "wrong-generation", "wrong-length", "error", "cancel"} {
		t.Run(mode, func(t *testing.T) {
			store := &recordedPluginSources{}
			dispatcher := &pluginLintDispatcher{sources: store}
			client, peer := newCLIChannelPair(t)
			dispatcher.channel = client
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
				var req pluginSourceRequest
				if err := msg.Decode(&req); err != nil {
					return nil, err
				}
				if req.SourceBatch == nil || req.Files[0].Text != nil {
					return nil, errors.New("source not shared")
				}
				res := pluginSourceResponse{}
				switch mode {
				case "release":
					res.ReleasedSource = req.SourceBatch
				case "wrong-slot", "wrong-generation", "wrong-length":
					wrong := *req.SourceBatch
					switch mode {
					case "wrong-slot":
						wrong.Slot++
					case "wrong-generation":
						wrong.Generation++
					case "wrong-length":
						wrong.Length++
					}
					res.ReleasedSource = &wrong
				case "error":
					return nil, errors.New("worker failed")
				case "cancel":
					cancel()
					return nil, context.Canceled
				}
				return res, nil
			})
			client.Start()
			peer.Start()
			source := "debugger;"
			_, err := dispatcher.dispatch(ctx, linter.EslintPluginLintRequest{Files: []linter.EslintPluginLintFile{{Text: &source}}})
			if (mode == "error" || mode == "cancel") != (err != nil) {
				t.Fatalf("unexpected dispatch result: %v", err)
			}
			if (len(store.released) == 1) != (mode == "release") {
				t.Fatalf("unsafe reuse after %s", mode)
			}
		})
	}
}

func TestPluginSourceLargeBatch(t *testing.T) {
	store := &recordedPluginSources{}
	dispatcher := &pluginLintDispatcher{sources: store}
	client, peer := newCLIChannelPair(t)
	dispatcher.channel = client
	requests := 0
	peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		var req pluginSourceRequest
		if err := msg.Decode(&req); err != nil {
			return nil, err
		}
		requests++
		if len(req.Files) != 1 || req.Files[0].Text != nil || req.SourceBatch == nil {
			return nil, errors.New("large batch was not split into shared snapshots")
		}
		return pluginSourceResponse{
			EslintPluginLintResult: linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{FilePath: req.Files[0].Path}}},
			ReleasedSource:         req.SourceBatch,
		}, nil
	})
	client.Start()
	peer.Start()
	text := strings.Repeat("x", sharedsource.SlotSize/2+1)
	req := linter.EslintPluginLintRequest{Files: []linter.EslintPluginLintFile{
		{Path: "first.ts", Text: &text}, {Path: "second.ts", Text: &text}, {Path: "third.ts", Text: &text},
	}}
	result, err := dispatcher.dispatch(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if requests != 3 || len(result.Results) != 3 {
		t.Fatalf("lost batch results: %+v", result)
	}
	for i, file := range req.Files {
		if result.Results[i].FilePath != file.Path {
			t.Fatal("changed batch result order")
		}
	}
}

func TestPluginSourceCancellationAndUnavailableMapping(t *testing.T) {
	store := &recordedPluginSources{}
	dispatcher := &pluginLintDispatcher{sources: store}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := dispatcher.dispatch(ctx, linter.EslintPluginLintRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled dispatch: %v", err)
	}
	if store.generation != 0 {
		t.Fatal("cancelled dispatch wrote source bytes")
	}
	dispatcher.close()
	if !store.closed {
		t.Fatal("dispatcher did not close its source storage")
	}
	fallback := newPluginLintDispatcher(nil, &sharedsource.Descriptor{Version: 99})
	defer fallback.close()
	if fallback.sources != nil {
		t.Fatal("unsupported mapping did not select inline transport")
	}
}
