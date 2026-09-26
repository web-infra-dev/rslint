//go:build !js

package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/ipc/sharedsource"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// Wire tests record the storage contract; pool ownership/concurrency is tested
// in sharedsource, and engine integration exercises real OS mappings.
type recordedPluginSources struct {
	mu          sync.Mutex
	parts       []string
	generation  uint32
	unavailable bool
	released    []sharedsource.Batch
	closed      bool
	beforeStore func([]string)
}

func (s *recordedPluginSources) Store(parts []string) (sharedsource.Batch, bool) {
	if s.beforeStore != nil {
		s.beforeStore(parts)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.released = append(s.released, batch)
}
func (s *recordedPluginSources) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closed = true
	return nil
}

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
			client, peer := newCLIChannelPair(t)
			dispatcher := newPluginLintDispatcher(client, nil)
			dispatcher.sources = store
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
	client, peer := newCLIChannelPair(t)
	dispatcher := newPluginLintDispatcher(client, nil)
	dispatcher.sources = store
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var requests atomic.Int32
	allStarted := make(chan struct{})
	peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		var req pluginSourceRequest
		if err := msg.Decode(&req); err != nil {
			return nil, err
		}
		if requests.Add(1) == 3 {
			close(allStarted)
		}
		if len(req.Files) != 1 || req.Files[0].Text != nil || req.SourceBatch == nil {
			return nil, errors.New("large batch was not split into shared snapshots")
		}
		if req.Generation != "generation" || !req.CollectFixes || !req.CollectTiming || req.SuggestionsMode != linter.SuggestionsModeEager || len(req.Rules) != 1 {
			return nil, errors.New("split lost request metadata")
		}
		if req.Files[0].Path == "first.ts" {
			// Later parts must reach the worker pool while this file is slow.
			select {
			case <-allStarted:
			case <-ctx.Done():
				return nil, errors.New("storage splitting serialized plugin execution")
			}
		}
		return pluginSourceResponse{
			EslintPluginLintResult: linter.EslintPluginLintResult{Results: []linter.EslintPluginFileResult{{FilePath: req.Files[0].Path}}},
			ReleasedSource:         req.SourceBatch,
		}, nil
	})
	client.Start()
	peer.Start()
	text := strings.Repeat("x", sharedsource.SlotSize/2+1)
	req := linter.EslintPluginLintRequest{
		Generation: "generation",
		Files: []linter.EslintPluginLintFile{
			{Path: "first.ts", Text: &text}, {Path: "second.ts", Text: &text}, {Path: "third.ts", Text: &text},
		},
		Rules: map[string]linter.EslintPluginRuleConfig{"plugin/rule": {}}, CollectFixes: true,
		SuggestionsMode: linter.SuggestionsModeEager, CollectTiming: true,
	}
	result, err := dispatcher.dispatch(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 3 || len(result.Results) != 3 {
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
	dispatcher := newPluginLintDispatcher(nil, nil)
	dispatcher.sources = store
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

func TestPluginSourceConcurrencyBoundAcrossLogicalBatches(t *testing.T) {
	client, peer := newCLIChannelPair(t)
	dispatcher := newPluginLintDispatcher(client, nil)
	dispatcher.sources = &recordedPluginSources{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	arrivals := make(chan struct{}, 24)
	unblock := make(chan struct{})
	var unblockOnce sync.Once
	release := func() { unblockOnce.Do(func() { close(unblock) }) }
	defer release()
	var active, peak atomic.Int32
	peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		var req pluginSourceRequest
		if err := msg.Decode(&req); err != nil {
			return nil, err
		}
		n := active.Add(1)
		defer active.Add(-1)
		for previous := peak.Load(); previous < n && !peak.CompareAndSwap(previous, n); previous = peak.Load() {
		}
		arrivals <- struct{}{}
		select {
		case <-unblock:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return pluginSourceResponse{ReleasedSource: req.SourceBatch}, nil
	})
	client.Start()
	peer.Start()
	text := strings.Repeat("x", sharedsource.SlotSize/2+1)
	dispatchErrors := make(chan error, 2)
	for batch := range 2 {
		var req linter.EslintPluginLintRequest
		for file := range 12 {
			req.Files = append(req.Files, linter.EslintPluginLintFile{Path: fmt.Sprintf("%d-%d.ts", batch, file), Text: &text})
		}
		go func() {
			_, err := dispatcher.dispatch(ctx, req)
			dispatchErrors <- err
		}()
	}
	for range 8 {
		select {
		case <-arrivals:
		case <-ctx.Done():
			t.Fatal("the dispatcher did not fill the available worker capacity")
		}
	}
	select {
	case <-arrivals:
		t.Error("logical batches multiplied the shared in-flight limit")
	case <-time.After(50 * time.Millisecond):
	}
	release()
	for range 2 {
		if err := <-dispatchErrors; err != nil {
			t.Fatal(err)
		}
	}
	if peak.Load() > 8 {
		t.Fatalf("observed %d in-flight source requests", peak.Load())
	}
}

func TestPluginSourceCancellationJoinsStartedWriters(t *testing.T) {
	client, peer := newCLIChannelPair(t)
	dispatcher := newPluginLintDispatcher(client, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{}, 16)
	finishWrites := make(chan struct{})
	var writes atomic.Int32
	store := &recordedPluginSources{beforeStore: func([]string) {
		writes.Add(1)
		started <- struct{}{}
		<-finishWrites
	}}
	dispatcher.sources = store
	peer.SetInboundHandler(func(_ context.Context, _ *ipc.Message) (any, error) {
		return pluginSourceResponse{}, nil
	})
	client.Start()
	peer.Start()
	text := strings.Repeat("x", sharedsource.SlotSize/2+1)
	var req linter.EslintPluginLintRequest
	for range 12 {
		req.Files = append(req.Files, linter.EslintPluginLintFile{Text: &text})
	}
	done := make(chan error, 1)
	go func() {
		_, err := dispatcher.dispatch(ctx, req)
		done <- err
	}()
	for range 8 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			close(finishWrites)
			t.Fatal("writers did not start")
		}
	}
	cancel()
	select {
	case err := <-done:
		close(finishWrites)
		t.Fatalf("returned while source writers still owned the snapshot: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(finishWrites)
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled dispatch: %v", err)
	}
	if writes.Load() != 8 || len(dispatcher.inFlight) != 0 {
		t.Fatal("cancelled dispatch started later parts or retained an in-flight permit")
	}
}

func TestPluginSourceSplitErrorsKeepRequestOrder(t *testing.T) {
	client, peer := newCLIChannelPair(t)
	dispatcher := newPluginLintDispatcher(client, nil)
	dispatcher.sources = &recordedPluginSources{}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	laterFailed := make(chan *ipc.ResponseReceipt, 1)
	peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
		var req pluginSourceRequest
		if err := msg.Decode(&req); err != nil {
			return nil, err
		}
		if req.Files[0].Path == "first.ts" {
			select {
			case receipt := <-laterFailed:
				<-receipt.Done()
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			return nil, errors.New("first request failed")
		}
		// A malformed normal response fails decoding without sealing the peer;
		// its receipt ensures the later part completes before the first error.
		response, receipt := ipc.TrackResponse(map[string]any{"results": "invalid"})
		laterFailed <- receipt
		return response, nil
	})
	client.Start()
	peer.Start()
	text := strings.Repeat("x", sharedsource.SlotSize/2+1)
	_, err := dispatcher.dispatch(ctx, linter.EslintPluginLintRequest{Files: []linter.EslintPluginLintFile{
		{Path: "first.ts", Text: &text}, {Path: "second.ts", Text: &text},
	}})
	if err == nil || err.Error() != "first request failed" {
		t.Fatalf("error selected by completion order: %v", err)
	}
}

func TestPluginSourceSplitFailureOutranksCancellation(t *testing.T) {
	client, peer := newCLIChannelPair(t)
	dispatcher := newPluginLintDispatcher(client, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	firstArrived, secondWriting := make(chan struct{}), make(chan struct{})
	failWriter := make(chan struct{})
	dispatcher.sources = &recordedPluginSources{beforeStore: func(parts []string) {
		if parts[0][0] == 'b' {
			close(secondWriting)
			<-failWriter
			panic("later writer failed")
		}
	}}
	peer.SetInboundHandler(func(peerCtx context.Context, _ *ipc.Message) (any, error) {
		close(firstArrived)
		// No reply before cancellation: the first part must return the real
		// context.Canceled sentinel rather than a stringified peer error.
		<-peerCtx.Done()
		return nil, peerCtx.Err()
	})
	client.Start()
	peer.Start()
	first, second := strings.Repeat("a", sharedsource.SlotSize/2+1), strings.Repeat("b", sharedsource.SlotSize/2+1)
	done := make(chan error, 1)
	go func() {
		_, err := dispatcher.dispatch(ctx, linter.EslintPluginLintRequest{Files: []linter.EslintPluginLintFile{
			{Text: &first}, {Text: &second},
		}})
		done <- err
	}()
	for _, entered := range []<-chan struct{}{firstArrived, secondWriting} {
		select {
		case <-entered:
		case <-time.After(5 * time.Second):
			close(failWriter)
			t.Fatal("both parts did not start")
		}
	}
	cancel()
	close(failWriter)
	if err := <-done; err == nil || !strings.Contains(err.Error(), "later writer failed") {
		t.Fatalf("earlier cancellation masked a later writer panic: %v", err)
	}
	if len(dispatcher.inFlight) != 0 {
		t.Fatal("panic retained an in-flight permit")
	}
}
