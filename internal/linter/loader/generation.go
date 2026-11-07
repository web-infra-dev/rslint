// Package loader adapts request-scoped Program loads to lint generations.
// Program construction and lint scheduling remain in their owning packages.
package loader

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	programLoader "github.com/web-infra-dev/rslint/internal/program/loader"
)

// GenerationLoad preserves deferred loading and projects a complete binding
// through the entry point's generation adapter. Request-owned Programs do not
// carry a producer lease, so successful acquisition returns no ReleaseFunc.
func GenerationLoad(
	load func(context.Context) (programLoader.LoadResult, error),
	fsys vfs.FS,
	generation func(programLoader.LoadResult, vfs.FS) linter.Generation,
) linter.GenerationLoad {
	return linter.GenerationLoad{
		Acquire: func(ctx context.Context) (linter.Generation, linter.ReleaseFunc, error) {
			if err := ctx.Err(); err != nil {
				return linter.Generation{}, nil, err
			}
			binding, err := load(ctx)
			if err != nil {
				return linter.Generation{}, nil, err
			}
			return generation(binding, fsys), nil, nil
		},
	}
}

// GenerationLoads maps independent Program loads to their admission metadata
// and deferred generation acquisition.
func GenerationLoads(
	loads []programLoader.ProgramLoad,
	fsys vfs.FS,
	generation func(programLoader.LoadResult, vfs.FS) linter.Generation,
) []linter.GenerationLoad {
	result := make([]linter.GenerationLoad, len(loads))
	for index, load := range loads {
		result[index] = GenerationLoad(func(context.Context) (programLoader.LoadResult, error) {
			return load.Load()
		}, fsys, generation)
		result[index].TargetCount = load.TargetCount
		if load.Options != nil && load.Options.Checkers != nil {
			result[index].CheckerCount = max(1, *load.Options.Checkers)
		}
	}
	return result
}
