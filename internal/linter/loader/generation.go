// Package loader adapts request-scoped Program loads to lint generations.
// Program construction and lint scheduling remain in their owning packages.
package loader

import (
	"context"

	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	programLoader "github.com/web-infra-dev/rslint/internal/program/loader"
)

// GenerationLoads preserves deferred loading and projects each loaded binding
// through the entry point's generation adapter. Request-owned Programs do not
// carry a producer lease, so successful acquisition returns no ReleaseFunc.
func GenerationLoads(
	loads []programLoader.ProgramLoad,
	fsys vfs.FS,
	generation func(programLoader.LoadResult, vfs.FS) linter.Generation,
) []linter.GenerationLoad {
	result := make([]linter.GenerationLoad, len(loads))
	for index, load := range loads {
		checkers := 0
		if load.Options != nil && load.Options.Checkers != nil {
			checkers = max(1, *load.Options.Checkers)
		}
		result[index] = linter.GenerationLoad{
			TargetCount:  load.TargetCount,
			CheckerCount: checkers,
			Acquire: func(ctx context.Context) (linter.Generation, linter.ReleaseFunc, error) {
				if err := ctx.Err(); err != nil {
					return linter.Generation{}, nil, err
				}
				binding, err := load.Load()
				if err != nil {
					return linter.Generation{}, nil, err
				}
				return generation(binding, fsys), nil, nil
			},
		}
	}
	return result
}
