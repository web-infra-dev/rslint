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
	initial    loader.LoadResult
	initialFS  vfs.FS
	rebuild    func(context.Context, linter.SourceSnapshot) (loader.LoadResult, vfs.FS, error)
	generation func(loader.LoadResult, vfs.FS) linter.Generation
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
	if snapshot.Empty() {
		return p.generation(p.initial, p.initialFS), nil, nil
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
