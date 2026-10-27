package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/web-infra-dev/rslint/internal/linter"
)

// apiGenerationProvider materializes immutable request-local generations from
// the pipeline-owned memory snapshot. The API does not own fix rounds or
// mutate its base request VFS.
type apiGenerationProvider struct {
	initial func(context.Context) (linter.Generation, error)
	loads   func(context.Context) ([]linter.GenerationLoad, error)
	rebuild func(context.Context, linter.SourceSnapshot) (linter.Generation, error)
}

func (p *apiGenerationProvider) AcquireGeneration(
	ctx context.Context,
	snapshot linter.SourceSnapshot,
) (linter.Generation, linter.ReleaseFunc, error) {
	if err := ctx.Err(); err != nil {
		return linter.Generation{}, nil, err
	}
	if p == nil {
		return linter.Generation{}, nil, errors.New("API lint generation provider is not configured")
	}
	if snapshot.Empty() && p.initial != nil {
		generation, err := p.initial(ctx)
		if err != nil {
			return linter.Generation{}, nil, err
		}
		// Transfer our ownership while published generation contents remain
		// immutable. There is no external borrow to finalize for an API request.
		p.initial = nil
		return generation, nil, nil
	}
	if p.rebuild == nil {
		return linter.Generation{}, nil, errors.New("rebuild API lint generation: provider is not configured")
	}
	generation, err := p.rebuild(ctx, snapshot)
	if err != nil {
		return linter.Generation{}, nil, fmt.Errorf("rebuild API lint generation: %w", err)
	}
	return generation, nil, nil
}

func (p *apiGenerationProvider) PlanGeneration(ctx context.Context, snapshot linter.SourceSnapshot) (linter.GenerationPlan, error) {
	plan := linter.GenerationPlan{}
	if err := ctx.Err(); err != nil {
		return plan, err
	}
	if snapshot.Empty() && p.initial != nil && p.loads != nil {
		loads, err := p.loads(ctx)
		if err != nil {
			return plan, err
		}
		plan.Loads = loads
	}
	if len(plan.Loads) == 0 {
		plan.Loads = []linter.GenerationLoad{{Acquire: func(ctx context.Context) (linter.Generation, linter.ReleaseFunc, error) {
			return p.AcquireGeneration(ctx, snapshot)
		}}}
	}
	return plan, nil
}
