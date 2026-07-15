package discovery

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type predicateEvaluation struct {
	ctx   context.Context
	call  ConfigPredicateCall
	ready chan struct{}
	value bool
	err   error
}

// predicateCoordinator batches only predicates that independent ConfigArray
// evaluations have currently reached. No OR/AND/ignore successor is enqueued
// until its predecessor resolves, preserving JavaScript short-circuit effects.
type predicateCoordinator struct {
	lifecycleCtx  context.Context
	cancel        context.CancelFunc
	loader        ConfigModuleLoader
	transactionID string
	queue         chan *predicateEvaluation
	sequence      atomic.Uint64
	closeOnce     sync.Once
	waitGroup     sync.WaitGroup
}

func newPredicateCoordinator(
	_ context.Context,
	loader ConfigModuleLoader,
	transactionID string,
) *predicateCoordinator {
	lifecycleCtx, cancel := context.WithCancel(context.Background())
	coordinator := &predicateCoordinator{
		lifecycleCtx:  lifecycleCtx,
		cancel:        cancel,
		loader:        loader,
		transactionID: transactionID,
		queue:         make(chan *predicateEvaluation, 128),
	}
	coordinator.waitGroup.Add(1)
	go coordinator.run()
	return coordinator
}

func (coordinator *predicateCoordinator) ResolveConfigPredicate(
	ctx context.Context,
	predicateID string,
	filePath string,
	directory bool,
) (bool, error) {
	if coordinator == nil || coordinator.loader == nil {
		return false, errors.New("live config predicates require an evaluateConfigPredicates host")
	}
	callID := fmt.Sprintf("predicate-%09d", coordinator.sequence.Add(1))
	evaluation := &predicateEvaluation{
		ctx: ctx,
		call: ConfigPredicateCall{
			CallID:       callID,
			PredicateID:  predicateID,
			AbsolutePath: filePath,
			Directory:    directory,
		},
		ready: make(chan struct{}),
	}
	select {
	case coordinator.queue <- evaluation:
	case <-ctx.Done():
		return false, ctx.Err()
	case <-coordinator.lifecycleCtx.Done():
		return false, coordinator.lifecycleCtx.Err()
	}
	select {
	case <-evaluation.ready:
		return evaluation.value, evaluation.err
	case <-ctx.Done():
		return false, ctx.Err()
	case <-coordinator.lifecycleCtx.Done():
		return false, coordinator.lifecycleCtx.Err()
	}
}

func (coordinator *predicateCoordinator) Close() {
	if coordinator == nil {
		return
	}
	coordinator.closeOnce.Do(func() {
		// A catalog may be closed while a later LSP query is enqueueing. Cancel
		// the lifecycle instead of closing queue so that race cannot panic with
		// "send on closed channel". Resolve callers observe lifecycleCtx and the
		// worker cancels any in-flight reverse request before it exits.
		coordinator.cancel()
		coordinator.waitGroup.Wait()
	})
}

func (coordinator *predicateCoordinator) run() {
	defer coordinator.waitGroup.Done()
	for {
		var first *predicateEvaluation
		select {
		case <-coordinator.lifecycleCtx.Done():
			return
		case first = <-coordinator.queue:
		}
		batch := []*predicateEvaluation{first}
		runtime.Gosched()
	drain:
		for {
			select {
			case <-coordinator.lifecycleCtx.Done():
				coordinator.fail(batch, coordinator.lifecycleCtx.Err())
				return
			case evaluation := <-coordinator.queue:
				batch = append(batch, evaluation)
			default:
				break drain
			}
		}
		coordinator.perform(batch)
	}
}

func (coordinator *predicateCoordinator) fail(batch []*predicateEvaluation, err error) {
	for _, evaluation := range batch {
		evaluation.err = err
		close(evaluation.ready)
	}
}

func (coordinator *predicateCoordinator) perform(batch []*predicateEvaluation) {
	active := make([]*predicateEvaluation, 0, len(batch))
	request := ConfigPredicateBatchRequest{
		TransactionID: coordinator.transactionID,
		Calls:         make([]ConfigPredicateCall, 0, len(batch)),
	}
	for _, evaluation := range batch {
		if err := evaluation.ctx.Err(); err != nil {
			evaluation.err = err
			close(evaluation.ready)
			continue
		}
		active = append(active, evaluation)
		request.Calls = append(request.Calls, evaluation.call)
	}
	if len(active) == 0 {
		return
	}

	response, err := coordinator.loader.EvaluateConfigPredicates(coordinator.lifecycleCtx, request)
	var results map[string]ConfigPredicateResult
	if err == nil {
		results, err = validateConfigPredicateBatch(request, response)
	}
	for _, evaluation := range active {
		if err != nil {
			evaluation.err = err
			close(evaluation.ready)
			continue
		}
		result := results[evaluation.call.CallID]
		if result.Status == "failed" {
			evaluation.err = fmt.Errorf("config predicate %q failed: %s", evaluation.call.PredicateID, result.Error.Message)
		} else {
			evaluation.value = result.Value
		}
		close(evaluation.ready)
	}
}

var _ rslintconfig.ConfigPredicateResolver = (*predicateCoordinator)(nil)
