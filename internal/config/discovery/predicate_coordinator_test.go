package discovery

import (
	"context"
	"sync"
	"testing"
	"time"
)

type recordingPredicateLoader struct {
	mu       sync.Mutex
	requests []ConfigPredicateBatchRequest
}

func (loader *recordingPredicateLoader) LoadConfigs(
	context.Context,
	ConfigLoadBatchRequest,
) (ConfigLoadBatchResponse, error) {
	panic("unexpected LoadConfigs call")
}

func (loader *recordingPredicateLoader) ActivateConfigs(
	context.Context,
	ConfigActivationRequest,
) (ConfigActivationResponse, error) {
	panic("unexpected ActivateConfigs call")
}

func (loader *recordingPredicateLoader) EvaluateConfigPredicates(
	_ context.Context,
	request ConfigPredicateBatchRequest,
) (ConfigPredicateBatchResponse, error) {
	loader.mu.Lock()
	loader.requests = append(loader.requests, request)
	loader.mu.Unlock()
	response := ConfigPredicateBatchResponse{TransactionID: request.TransactionID}
	for _, call := range request.Calls {
		response.Results = append(response.Results, ConfigPredicateResult{
			CallID: call.CallID,
			Status: "evaluated",
			Value:  call.PredicateID == "truthy",
		})
	}
	return response, nil
}

func TestPredicateCoordinatorBatchesOnlyReachedIndependentCalls(t *testing.T) {
	loader := &recordingPredicateLoader{}
	lifecycleCtx, cancel := context.WithCancel(context.Background())
	coordinator := &predicateCoordinator{
		lifecycleCtx:  lifecycleCtx,
		cancel:        cancel,
		loader:        loader,
		transactionID: "tx",
		queue:         make(chan *predicateEvaluation, 8),
	}
	type result struct {
		value bool
		err   error
	}
	results := make(chan result, 2)
	for _, predicateID := range []string{"truthy", "falsy"} {
		go func() {
			value, err := coordinator.ResolveConfigPredicate(context.Background(), predicateID, "/repo/file.ts", false)
			results <- result{value, err}
		}()
	}

	deadline := time.After(5 * time.Second)
	for len(coordinator.queue) != 2 {
		select {
		case <-deadline:
			t.Fatalf("queued calls = %d, want 2", len(coordinator.queue))
		default:
			time.Sleep(time.Millisecond)
		}
	}
	coordinator.waitGroup.Add(1)
	go coordinator.run()

	seen := map[bool]bool{}
	for range 2 {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		seen[result.value] = true
	}
	coordinator.Close()
	if !seen[true] || !seen[false] {
		t.Fatalf("results = %v, want true and false", seen)
	}
	loader.mu.Lock()
	requests := append([]ConfigPredicateBatchRequest(nil), loader.requests...)
	loader.mu.Unlock()
	if len(requests) != 1 || len(requests[0].Calls) != 2 {
		t.Fatalf("requests = %+v, want one two-call batch", requests)
	}
}

func TestPredicateCoordinatorOutlivesCatalogBuildContext(t *testing.T) {
	loader := &recordingPredicateLoader{}
	buildCtx, cancelBuild := context.WithCancel(context.Background())
	coordinator := newPredicateCoordinator(buildCtx, loader, "tx-committed")
	cancelBuild()
	defer coordinator.Close()

	value, err := coordinator.ResolveConfigPredicate(
		context.Background(),
		"truthy",
		"/repo/later.ts",
		false,
	)
	if err != nil {
		t.Fatalf("committed predicate after build cancellation: %v", err)
	}
	if !value {
		t.Fatal("committed predicate returned false, want true")
	}
}
