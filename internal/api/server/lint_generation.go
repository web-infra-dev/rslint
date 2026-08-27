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
	initial linter.Generation
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
	if snapshot.Empty() {
		return p.initial, nil, nil
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
