package api

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/web-infra-dev/rslint/internal/ipc"
)

type serviceTestHandler struct {
	lintCalls atomic.Int32
}

func (h *serviceTestHandler) HandleLint(LintRequest) (*LintResponse, error) {
	h.lintCalls.Add(1)
	return &LintResponse{Diagnostics: []Diagnostic{}, LintedFiles: []string{}}, nil
}

func (h *serviceTestHandler) HandleGetAstInfo(GetAstInfoRequest) (*GetAstInfoResponse, error) {
	return &GetAstInfoResponse{}, nil
}

type reverseServiceTestHandler struct {
	serviceTestHandler
	reverseCalls atomic.Int32
	reverseErr   chan error
	requests     chan LintRequest
}

type configDiscoveryServiceTestHandler struct {
	serviceTestHandler
	configDiscoveryCalls atomic.Int32
	sawCapabilityView    atomic.Bool
	configLoadCapability atomic.Bool
	pluginLintCapability atomic.Bool
}

func (h *configDiscoveryServiceTestHandler) HandleLintWithContext(_ context.Context, _ LintRequest, requester Requester) (*LintResponse, error) {
	h.configDiscoveryCalls.Add(1)
	if capabilityRequester, ok := requester.(PeerCapabilityRequester); ok {
		h.sawCapabilityView.Store(true)
		h.configLoadCapability.Store(capabilityRequester.PeerSupportsCapability(CapabilityReverseConfigLoad))
		h.pluginLintCapability.Store(capabilityRequester.PeerSupportsCapability(CapabilityReversePluginLint))
	}
	return &LintResponse{Diagnostics: []Diagnostic{}, LintedFiles: []string{}}, nil
}

func (h *reverseServiceTestHandler) HandleLintWithContext(ctx context.Context, req LintRequest, requester Requester) (*LintResponse, error) {
	h.reverseCalls.Add(1)
	if h.requests != nil {
		h.requests <- req
	}
	msg, err := requester.SendRequest(ctx, KindPluginLint, struct {
		Text string `json:"text"`
	}{Text: "from-go"})
	if h.reverseErr != nil {
		h.reverseErr <- err
	}
	if err != nil {
		return nil, err
	}
	var result struct {
		Text string `json:"text"`
	}
	if err := msg.Decode(&result); err != nil {
		return nil, err
	}
	return &LintResponse{
		Diagnostics: []Diagnostic{{RuleName: "test", Message: result.Text}},
		LintedFiles: []string{},
	}, nil
}

type serviceChannelPair struct {
	peer          *ipc.Channel
	serviceDone   chan error
	serviceClosed chan struct{}
	peerToService *io.PipeWriter
	serviceToPeer *io.PipeWriter
	readers       []*io.PipeReader
}

func newServiceChannelPair(t *testing.T, handler Handler, peerHandler ipc.InboundHandler) *serviceChannelPair {
	t.Helper()
	peerToServiceR, peerToServiceW := io.Pipe()
	serviceToPeerR, serviceToPeerW := io.Pipe()

	service := NewService(peerToServiceR, serviceToPeerW, handler)
	peer := ipc.NewChannel(serviceToPeerR, peerToServiceW)
	if peerHandler != nil {
		peer.SetInboundHandler(peerHandler)
	}

	done := make(chan error, 1)
	closed := make(chan struct{})
	go func() {
		done <- service.Start()
		close(closed)
	}()
	peer.Start()

	pair := &serviceChannelPair{
		peer:          peer,
		serviceDone:   done,
		serviceClosed: closed,
		peerToService: peerToServiceW,
		serviceToPeer: serviceToPeerW,
		readers:       []*io.PipeReader{peerToServiceR, serviceToPeerR},
	}
	t.Cleanup(func() {
		_ = peer.Close()
		_ = peerToServiceW.Close()
		_ = serviceToPeerW.Close()
		for _, reader := range pair.readers {
			_ = reader.Close()
		}
		select {
		case <-closed:
		case <-time.After(2 * time.Second):
			t.Errorf("API service did not stop during cleanup")
		}
	})
	return pair
}

func requestContext(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	t.Cleanup(cancel)
	return ctx
}

type gatedAPIWriter struct {
	mu          sync.Mutex
	buf         bytes.Buffer
	calls       int
	bodyEntered chan struct{}
	releaseBody chan struct{}
	releaseOnce sync.Once
}

func newGatedAPIWriter() *gatedAPIWriter {
	return &gatedAPIWriter{
		bodyEntered: make(chan struct{}),
		releaseBody: make(chan struct{}),
	}
}

func (w *gatedAPIWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	w.calls++
	call := w.calls
	w.mu.Unlock()
	if call == 2 {
		close(w.bodyEntered)
		<-w.releaseBody
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buf.Write(p)
}

func (w *gatedAPIWriter) release() {
	w.releaseOnce.Do(func() { close(w.releaseBody) })
}

func (w *gatedAPIWriter) bytes() []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	return bytes.Clone(w.buf.Bytes())
}

func TestService_ExitWaitsForCompleteAcknowledgement(t *testing.T) {
	inR, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close() })
	writer := newGatedAPIWriter()
	t.Cleanup(writer.release)
	service := NewService(inR, writer, &serviceTestHandler{})

	serviceDone := make(chan error, 1)
	go func() { serviceDone <- service.Start() }()

	request, err := ipc.NewMessage(ipc.KindExit, 17, struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if err := ipc.WriteFrame(inW, request); err != nil {
		t.Fatalf("write exit request: %v", err)
	}
	select {
	case <-writer.bodyEntered:
	case <-time.After(2 * time.Second):
		t.Fatal("exit acknowledgement did not reach the body-write gate")
	}
	select {
	case err := <-serviceDone:
		t.Fatalf("service returned before the complete exit acknowledgement: %v", err)
	default:
	}

	writer.release()
	select {
	case err := <-serviceDone:
		if err != nil {
			t.Fatalf("service exit: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("service did not return after the complete exit acknowledgement")
	}

	response, err := ipc.ReadFrame(bufio.NewReader(bytes.NewReader(writer.bytes())))
	if err != nil {
		t.Fatalf("read exit acknowledgement: %v", err)
	}
	if response.Kind != ipc.KindResponse || response.ID != 17 {
		t.Fatalf("exit acknowledgement = kind %q id %d", response.Kind, response.ID)
	}
}

type failingAPIWriter struct{}

func (failingAPIWriter) Write([]byte) (int, error) {
	return 0, errors.New("api writer boom")
}

func TestService_ExitWriteFailureIsNotSuccess(t *testing.T) {
	inR, inW := io.Pipe()
	t.Cleanup(func() { _ = inW.Close() })
	service := NewService(inR, failingAPIWriter{}, &serviceTestHandler{})

	serviceDone := make(chan error, 1)
	go func() { serviceDone <- service.Start() }()
	request, err := ipc.NewMessage(ipc.KindExit, 23, struct{}{})
	if err != nil {
		t.Fatal(err)
	}
	if err := ipc.WriteFrame(inW, request); err != nil {
		t.Fatalf("write exit request: %v", err)
	}

	select {
	case err := <-serviceDone:
		if err == nil {
			t.Fatal("exit acknowledgement write failure returned success")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("service hung after the exit acknowledgement write failed")
	}
}

func TestService_BidirectionalLintKeepsReadLoopRunning(t *testing.T) {
	handler := &reverseServiceTestHandler{requests: make(chan LintRequest, 1)}
	pair := newServiceChannelPair(t, handler, func(_ context.Context, msg *ipc.Message) (any, error) {
		if msg.Kind != KindPluginLint {
			t.Fatalf("unexpected reverse request kind %q", msg.Kind)
		}
		var req struct {
			Text string `json:"text"`
		}
		if err := msg.Decode(&req); err != nil {
			return nil, err
		}
		return struct {
			Text string `json:"text"`
		}{Text: req.Text + "-and-back"}, nil
	})

	ctx := requestContext(t)
	handshake, err := pair.peer.SendRequest(ctx, ipc.KindHandshake, HandshakeRequest{
		Version:      Version,
		Capabilities: []string{CapabilityReversePluginLint},
	})
	if err != nil {
		t.Fatalf("handshake: %v", err)
	}
	var handshakeResult HandshakeResponse
	if err := handshake.Decode(&handshakeResult); err != nil {
		t.Fatalf("decode handshake: %v", err)
	}
	if !handshakeResult.OK || handshakeResult.Version != Version {
		t.Fatalf("unexpected handshake response: %+v", handshakeResult)
	}
	if len(handshakeResult.Capabilities) != 2 ||
		handshakeResult.Capabilities[0] != CapabilityReversePluginLint ||
		handshakeResult.Capabilities[1] != CapabilityReverseConfigLoad {
		t.Fatalf("bidirectional handler did not advertise both reverse capabilities: %+v", handshakeResult.Capabilities)
	}

	msg, err := pair.peer.SendRequest(ctx, KindLint, LintRequest{EslintPlugins: []EslintPluginEntry{{
		Prefix:    "community",
		RuleNames: []string{"rule"},
	}}})
	if err != nil {
		t.Fatalf("lint request deadlocked or failed: %v", err)
	}
	var lintResult LintResponse
	if err := msg.Decode(&lintResult); err != nil {
		t.Fatalf("decode lint response: %v", err)
	}
	if handler.reverseCalls.Load() != 1 {
		t.Fatalf("expected one context-aware lint call, got %d", handler.reverseCalls.Load())
	}
	received := <-handler.requests
	if len(received.EslintPlugins) != 1 || received.EslintPlugins[0].Prefix != "community" ||
		len(received.EslintPlugins[0].RuleNames) != 1 || received.EslintPlugins[0].RuleNames[0] != "rule" {
		t.Fatalf("eslintPlugins metadata did not survive the wire: %+v", received.EslintPlugins)
	}
	if len(lintResult.Diagnostics) != 1 || lintResult.Diagnostics[0].Message != "from-go-and-back" {
		t.Fatalf("unexpected reverse RPC result: %+v", lintResult.Diagnostics)
	}

	if _, err := pair.peer.SendRequest(ctx, KindGetAstInfo, GetAstInfoRequest{}); err != nil {
		t.Fatalf("getAstInfo compatibility request: %v", err)
	}
	if _, err := pair.peer.SendRequest(ctx, ipc.KindExit, struct{}{}); err != nil {
		t.Fatalf("exit request: %v", err)
	}
	select {
	case err := <-pair.serviceDone:
		if err != nil {
			t.Fatalf("service exit: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("service did not stop after writing the exit acknowledgement")
	}
}

func TestService_LegacyHandlerFallback(t *testing.T) {
	handler := &serviceTestHandler{}
	pair := newServiceChannelPair(t, handler, nil)
	ctx := requestContext(t)
	message, err := pair.peer.SendRequest(ctx, ipc.KindHandshake, HandshakeRequest{Version: Version})
	if err != nil {
		t.Fatalf("handshake: %v", err)
	}
	var response HandshakeResponse
	if err := message.Decode(&response); err != nil {
		t.Fatalf("decode handshake: %v", err)
	}
	if len(response.Capabilities) != 0 {
		t.Fatalf("legacy handler advertised unsupported capabilities: %+v", response.Capabilities)
	}

	if _, err := pair.peer.SendRequest(ctx, KindLint, LintRequest{}); err != nil {
		t.Fatalf("legacy lint handler: %v", err)
	}
	if handler.lintCalls.Load() != 1 {
		t.Fatalf("expected legacy HandleLint once, got %d", handler.lintCalls.Load())
	}
	if _, err := pair.peer.SendRequest(ctx, ipc.KindExit, struct{}{}); err != nil {
		t.Fatalf("exit request: %v", err)
	}
}

func TestService_TransportShutdownCancelsReverseRequest(t *testing.T) {
	reverseStarted := make(chan struct{})
	handler := &reverseServiceTestHandler{reverseErr: make(chan error, 1)}
	pair := newServiceChannelPair(t, handler, func(ctx context.Context, msg *ipc.Message) (any, error) {
		if msg.Kind != KindPluginLint {
			return nil, errors.New("unexpected reverse request")
		}
		close(reverseStarted)
		<-ctx.Done()
		return nil, ctx.Err()
	})

	ctx := requestContext(t)
	if _, err := pair.peer.SendRequest(ctx, ipc.KindHandshake, HandshakeRequest{
		Version:      Version,
		Capabilities: []string{CapabilityReversePluginLint},
	}); err != nil {
		t.Fatalf("handshake: %v", err)
	}
	go func() {
		_, _ = pair.peer.SendRequest(ctx, KindLint, LintRequest{})
	}()
	select {
	case <-reverseStarted:
	case <-time.After(2 * time.Second):
		t.Fatal("service never issued reverse pluginLint request")
	}

	// Closing the host's write half gives the service a clean peer EOF while a
	// reverse request is pending. Channel shutdown must wake SendRequest.
	if err := pair.peerToService.Close(); err != nil {
		t.Fatalf("close peer transport: %v", err)
	}
	select {
	case err := <-handler.reverseErr:
		if err == nil {
			t.Fatal("expected reverse request to fail when the transport closes")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("reverse request remained blocked after transport shutdown")
	}
	select {
	case err := <-pair.serviceDone:
		if err != nil {
			t.Fatalf("clean peer EOF should stop service without error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("service did not stop after peer EOF")
	}
}

func TestService_RejectsProtocolMismatch(t *testing.T) {
	handler := &serviceTestHandler{}
	pair := newServiceChannelPair(t, handler, nil)
	ctx := requestContext(t)

	message, err := pair.peer.SendRequest(ctx, ipc.KindHandshake, HandshakeRequest{Version: "1.0.0"})
	if err != nil {
		t.Fatalf("handshake request: %v", err)
	}
	var response HandshakeResponse
	if err := message.Decode(&response); err != nil {
		t.Fatalf("decode handshake: %v", err)
	}
	if response.OK || response.Version != Version {
		t.Fatalf("unexpected mismatch response: %+v", response)
	}
	if _, err := pair.peer.SendRequest(ctx, KindLint, LintRequest{}); err == nil {
		t.Fatal("lint should be rejected after a mismatched handshake")
	}
	_, _ = pair.peer.SendRequest(ctx, ipc.KindExit, struct{}{})
}

func TestService_RequiresPeerCapabilityForPluginLint(t *testing.T) {
	handler := &reverseServiceTestHandler{}
	pair := newServiceChannelPair(t, handler, nil)
	ctx := requestContext(t)

	if _, err := pair.peer.SendRequest(ctx, ipc.KindHandshake, HandshakeRequest{Version: Version}); err != nil {
		t.Fatalf("handshake: %v", err)
	}
	_, err := pair.peer.SendRequest(ctx, KindLint, LintRequest{
		EslintPlugins: []EslintPluginEntry{{Prefix: "community", RuleNames: []string{"rule"}}},
	})
	if err == nil || !strings.Contains(err.Error(), CapabilityReversePluginLint) {
		t.Fatalf("expected missing-capability error, got %v", err)
	}
	_, _ = pair.peer.SendRequest(ctx, ipc.KindExit, struct{}{})
}

func TestService_ConfigDiscoveryRequiresAdvertisedCapability(t *testing.T) {
	tests := []struct {
		name         string
		capabilities []string
		wantError    string
		wantCalls    int32
	}{
		{
			name:      "missing capability",
			wantError: CapabilityReverseConfigLoad,
		},
		{
			name:         "advertised capability",
			capabilities: []string{CapabilityReverseConfigLoad},
			wantCalls:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			handler := &configDiscoveryServiceTestHandler{}
			pair := newServiceChannelPair(t, handler, nil)
			ctx := requestContext(t)
			if _, err := pair.peer.SendRequest(ctx, ipc.KindHandshake, HandshakeRequest{
				Version:      Version,
				Capabilities: test.capabilities,
			}); err != nil {
				t.Fatalf("handshake: %v", err)
			}

			_, err := pair.peer.SendRequest(ctx, KindLint, LintRequest{
				ConfigDiscovery: &ConfigDiscoveryRequest{},
			})
			if test.wantError == "" {
				if err != nil {
					t.Fatalf("lint with config discovery capability: %v", err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("error = %v, want capability error containing %q", err, test.wantError)
			}
			if got := handler.configDiscoveryCalls.Load(); got != test.wantCalls {
				t.Fatalf("config discovery handler calls = %d, want %d", got, test.wantCalls)
			}
			if test.wantCalls > 0 {
				if !handler.sawCapabilityView.Load() {
					t.Fatal("bidirectional handler did not receive the peer capability view")
				}
				if !handler.configLoadCapability.Load() {
					t.Fatal("peer capability view lost reverseConfigLoadV1")
				}
				if handler.pluginLintCapability.Load() {
					t.Fatal("peer capability view invented reversePluginLint")
				}
			}
			_, _ = pair.peer.SendRequest(ctx, ipc.KindExit, struct{}{})
		})
	}
}

func TestService_RejectsPluginMetadataForLegacyHandler(t *testing.T) {
	handler := &serviceTestHandler{}
	pair := newServiceChannelPair(t, handler, nil)
	ctx := requestContext(t)

	if _, err := pair.peer.SendRequest(ctx, ipc.KindHandshake, HandshakeRequest{
		Version:      Version,
		Capabilities: []string{CapabilityReversePluginLint},
	}); err != nil {
		t.Fatalf("handshake: %v", err)
	}
	_, err := pair.peer.SendRequest(ctx, KindLint, LintRequest{
		EslintPlugins: []EslintPluginEntry{{Prefix: "community", RuleNames: []string{"rule"}}},
	})
	if err == nil || !strings.Contains(err.Error(), "handler does not support reversePluginLint") {
		t.Fatalf("expected unsupported-handler error, got %v", err)
	}
	if handler.lintCalls.Load() != 0 {
		t.Fatalf("legacy HandleLint must not receive plugin metadata, got %d calls", handler.lintCalls.Load())
	}
	_, _ = pair.peer.SendRequest(ctx, ipc.KindExit, struct{}{})
}

func TestService_TruncatedFrameIsNotCleanShutdown(t *testing.T) {
	// Declares a 10-byte JSON body but supplies one byte before EOF.
	reader := bytes.NewReader([]byte{10, 0, 0, 0, '{'})
	err := NewService(reader, io.Discard, &serviceTestHandler{}).Start()
	if !errors.Is(err, io.ErrUnexpectedEOF) {
		t.Fatalf("expected truncated frame error, got %v", err)
	}
}
