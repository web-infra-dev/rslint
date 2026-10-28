package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/program/loader"
)

// cliGenerationProvider projects the pipeline's immutable memory snapshot onto
// a request-local overlay and rebuilds the exact Program and target generation.
// It never owns fix rounds or mutates disk.
type cliGenerationProvider struct {
	initial        func(context.Context) (loader.LoadResult, vfs.FS, error)
	loads          func(context.Context) ([]linter.GenerationLoad, error)
	singleThreaded bool
	rebuild        func(context.Context, linter.SourceSnapshot) (loader.LoadResult, vfs.FS, error)
	generation     func(loader.LoadResult, vfs.FS) linter.Generation
}

func (p *cliGenerationProvider) AcquireGeneration(
	ctx context.Context,
	snapshot linter.SourceSnapshot,
) (linter.Generation, linter.ReleaseFunc, error) {
	if err := ctx.Err(); err != nil {
		return linter.Generation{}, nil, err
	}
	if p == nil || p.generation == nil {
		return linter.Generation{}, nil, errors.New("CLI lint generation provider is not configured")
	}
	if snapshot.Empty() && p.initial != nil {
		binding, fsys, err := p.initial(ctx)
		if err != nil {
			return linter.Generation{}, nil, err
		}
		generation := p.generation(binding, fsys)
		// Transfer this request-local input at acquisition. A ReleaseFunc is
		// reserved for an actual producer lease, such as an LSP request borrow.
		p.initial = nil
		return generation, nil, nil
	}
	if p.rebuild == nil {
		return linter.Generation{}, nil, errors.New("rebuild CLI lint generation: provider is not configured")
	}
	binding, fsys, err := p.rebuild(ctx, snapshot)
	if err != nil {
		return linter.Generation{}, nil, fmt.Errorf("rebuild CLI lint generation: %w", err)
	}
	if len(binding.Programs) == 0 {
		return linter.Generation{}, nil, errors.New("rebuild CLI lint generation: no programs returned")
	}
	return p.generation(binding, fsys), nil, nil
}

func (p *cliGenerationProvider) PlanGeneration(ctx context.Context, snapshot linter.SourceSnapshot) (linter.GenerationPlan, error) {
	plan := linter.GenerationPlan{SingleThreaded: p.singleThreaded}
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
