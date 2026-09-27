package pluginlint

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
)

type testChannel struct {
	limit int
	send  func(context.Context, sourceRequest, []string) (*ipc.Message, error)
}

func (c *testChannel) AttachmentLimit() int { return c.limit }
func (c *testChannel) SendRequest(ctx context.Context, kind ipc.MessageKind, data any, sources ...string) (*ipc.Message, error) {
	if kind != "pluginLint" {
		return nil, fmt.Errorf("unexpected request kind: %s", kind)
	}
	return c.send(ctx, data.(sourceRequest), sources)
}

func testClient(send func(context.Context, sourceRequest, []string) (*ipc.Message, error)) *Client {
	return &Client{channel: &testChannel{limit: 16, send: send}, inFlight: make(chan struct{}, 8)}
}

func resultFor(req sourceRequest) (*ipc.Message, error) {
	var result linter.EslintPluginLintResult
	for _, file := range req.Files {
		result.Results = append(result.Results, linter.EslintPluginFileResult{FilePath: file.Path})
	}
	return ipc.NewMessage(ipc.KindResponse, 1, result)
}

func TestCompleteSnapshots(t *testing.T) {
	texts := []string{"", "\ufeff" + "const café = '😀';\r\n// \x00", string([]byte{0xff}), strings.Repeat("x", 32)}
	files := make([]linter.EslintPluginLintFile, len(texts)+1)
	for i := range texts {
		files[i].Text = &texts[i]
	}
	req := linter.EslintPluginLintRequest{Files: files}
	wire, attachments := encode(req)
	for i := range texts {
		file := wire.Files[i]
		if file.Text != nil || file.SourceIndex == nil || int(*file.SourceIndex) != i || attachments[*file.SourceIndex] != texts[i] || req.Files[i].Text != &texts[i] {
			t.Fatalf("lost or mutated snapshot %d", i)
		}
	}
	if wire.Files[len(texts)].SourceIndex != nil || len(attachments) != len(texts) {
		t.Fatal("source-less file acquired an attachment")
	}
}

func TestSplitPreservesMetadataAndOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var requests atomic.Int32
	allStarted := make(chan struct{})
	client := testClient(func(ctx context.Context, req sourceRequest, sources []string) (*ipc.Message, error) {
		if requests.Add(1) == 3 {
			close(allStarted)
		}
		if len(req.Files) != 1 || len(sources) != 1 || sources[0] != "123456789" || req.Files[0].Text != nil || req.Files[0].SourceIndex == nil {
			return nil, errors.New("source batch not split at the transport budget")
		}
		if req.Generation != "generation" || !req.CollectFixes || !req.CollectTiming || req.SuggestionsMode != linter.SuggestionsModeEager || len(req.Rules) != 1 {
			return nil, errors.New("split lost metadata")
		}
		if req.Files[0].Path == "first.ts" {
			select {
			case <-allStarted:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
		return resultFor(req)
	})
	text := "123456789"
	req := linter.EslintPluginLintRequest{
		Generation: "generation",
		Files:      []linter.EslintPluginLintFile{{Path: "first.ts", Text: &text}, {Path: "second.ts", Text: &text}, {Path: "third.ts", Text: &text}},
		Rules:      map[string]linter.EslintPluginRuleConfig{"plugin/rule": {}}, CollectFixes: true,
		SuggestionsMode: linter.SuggestionsModeEager, CollectTiming: true,
	}
	result, err := client.Dispatch(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	if requests.Load() != 3 || len(result.Results) != 3 {
		t.Fatal("lost split results")
	}
	for i, file := range req.Files {
		if result.Results[i].FilePath != file.Path {
			t.Fatal("changed result order")
		}
	}
}

func TestInlineAndOversizedAttachments(t *testing.T) {
	for _, limit := range []int{0, 16} {
		t.Run(strconv.Itoa(limit), func(t *testing.T) {
			text := strings.Repeat("x", 32)
			client := testClient(func(_ context.Context, req sourceRequest, sources []string) (*ipc.Message, error) {
				if len(req.Files) != 3 || len(sources) != 2 || sources[0] != text || sources[1] != text {
					t.Fatal("adapter truncated oversized sources")
				}
				return resultFor(req)
			})
			client.channel.(*testChannel).limit = limit
			_, err := client.Dispatch(context.Background(), linter.EslintPluginLintRequest{Files: []linter.EslintPluginLintFile{{Text: &text}, {}, {Text: &text}}})
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCancellationBeforeSend(t *testing.T) {
	client := testClient(func(context.Context, sourceRequest, []string) (*ipc.Message, error) {
		t.Fatal("cancelled request sent")
		return nil, errors.New("cancelled request sent")
	})
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := client.Dispatch(ctx, linter.EslintPluginLintRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled dispatch: %v", err)
	}
}

func TestConcurrencyBoundAcrossLogicalBatches(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	arrivals := make(chan struct{}, 24)
	unblock := make(chan struct{})
	var once sync.Once
	release := func() { once.Do(func() { close(unblock) }) }
	defer release()
	var active, peak atomic.Int32
	client := testClient(func(ctx context.Context, req sourceRequest, _ []string) (*ipc.Message, error) {
		n := active.Add(1)
		defer active.Add(-1)
		for prev := peak.Load(); prev < n && !peak.CompareAndSwap(prev, n); prev = peak.Load() {
		}
		arrivals <- struct{}{}
		select {
		case <-unblock:
		case <-ctx.Done():
			return nil, ctx.Err()
		}
		return resultFor(req)
	})
	text := "123456789"
	done := make(chan error, 2)
	for range 2 {
		var req linter.EslintPluginLintRequest
		for range 12 {
			req.Files = append(req.Files, linter.EslintPluginLintFile{Text: &text})
		}
		go func() { _, err := client.Dispatch(ctx, req); done <- err }()
	}
	for range 8 {
		select {
		case <-arrivals:
		case <-ctx.Done():
			t.Fatal("requests did not fill the budget")
		}
	}
	select {
	case <-arrivals:
		t.Error("logical requests multiplied concurrency")
	case <-time.After(50 * time.Millisecond):
	}
	release()
	for range 2 {
		if err := <-done; err != nil {
			t.Fatal(err)
		}
	}
	if peak.Load() > 8 {
		t.Fatalf("%d requests in flight", peak.Load())
	}
}

func TestCancellationJoinsStartedCalls(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{}, 16)
	finish := make(chan struct{})
	var writes atomic.Int32
	client := testClient(func(ctx context.Context, _ sourceRequest, _ []string) (*ipc.Message, error) {
		writes.Add(1)
		started <- struct{}{}
		<-finish // Models a writer that still owns the immutable source snapshot.
		return nil, ctx.Err()
	})
	text := "123456789"
	var req linter.EslintPluginLintRequest
	for range 12 {
		req.Files = append(req.Files, linter.EslintPluginLintFile{Text: &text})
	}
	done := make(chan error, 1)
	go func() { _, err := client.Dispatch(ctx, req); done <- err }()
	for range 8 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			close(finish)
			t.Fatal("calls did not start")
		}
	}
	cancel()
	select {
	case err := <-done:
		close(finish)
		t.Fatalf("returned with active writers: %v", err)
	case <-time.After(50 * time.Millisecond):
	}
	close(finish)
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled dispatch: %v", err)
	}
	if writes.Load() != 8 || len(client.inFlight) != 0 {
		t.Fatal("cancelled dispatch started later parts or retained permits")
	}
}

func TestSplitErrorsKeepRequestOrder(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	later := make(chan struct{})
	client := testClient(func(ctx context.Context, req sourceRequest, _ []string) (*ipc.Message, error) {
		if req.Files[0].Path == "first.ts" {
			select {
			case <-later:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
			return nil, errors.New("first request failed")
		}
		defer close(later)
		return ipc.NewMessage(ipc.KindResponse, 1, map[string]any{"results": "invalid"})
	})
	text := "123456789"
	_, err := client.Dispatch(ctx, linter.EslintPluginLintRequest{Files: []linter.EslintPluginLintFile{{Path: "first.ts", Text: &text}, {Path: "second.ts", Text: &text}}})
	if err == nil || err.Error() != "first request failed" {
		t.Fatalf("error selected by completion order: %v", err)
	}
}

func TestSplitFailureOutranksCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	started := make(chan struct{}, 2)
	client := testClient(func(ctx context.Context, req sourceRequest, _ []string) (*ipc.Message, error) {
		started <- struct{}{}
		<-ctx.Done()
		if req.Files[0].Path == "second.ts" {
			panic("later writer failed")
		}
		return nil, ctx.Err()
	})
	text := "123456789"
	done := make(chan error, 1)
	go func() {
		_, err := client.Dispatch(ctx, linter.EslintPluginLintRequest{Files: []linter.EslintPluginLintFile{{Path: "first.ts", Text: &text}, {Path: "second.ts", Text: &text}}})
		done <- err
	}()
	for range 2 {
		select {
		case <-started:
		case <-time.After(5 * time.Second):
			t.Fatal("calls did not start")
		}
	}
	cancel()
	if err := <-done; err == nil || !strings.Contains(err.Error(), "later writer failed") {
		t.Fatalf("cancellation masked a failure: %v", err)
	}
	if len(client.inFlight) != 0 {
		t.Fatal("panic retained a permit")
	}
}
