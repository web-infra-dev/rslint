package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/utils"
)

// cliFinalChangeCommitter performs the CLI's one terminal disk projection.
// All fix rounds happen in linter-owned memory before this adapter is called.
type cliFinalChangeCommitter struct{}

func (cliFinalChangeCommitter) CommitFinalChanges(
	ctx context.Context,
	changes []linter.FileChange,
) (linter.CommitResult, error) {
	result := linter.CommitResult{}
	if len(changes) == 0 {
		return result, nil
	}

	// Preflight the complete delta before the first write. Multi-file disk
	// replacement cannot be globally atomic, but stale input must not create an
	// avoidable partial commit.
	var preflightErrors []error
	for _, change := range changes {
		if err := ctx.Err(); err != nil {
			preflightErrors = append(preflightErrors, err)
			break
		}
		current, ok := readPhysicalFileText(change.Path)
		if !ok {
			preflightErrors = append(preflightErrors, fmt.Errorf("read fix target %q", change.Path))
			continue
		}
		if current != change.Before {
			preflightErrors = append(preflightErrors, fmt.Errorf("fix target %q changed after lint generation", change.Path))
		}
	}
	if err := errors.Join(preflightErrors...); err != nil {
		return result, err
	}

	var writeErrors []error
	for _, change := range changes {
		if err := ctx.Err(); err != nil {
			writeErrors = append(writeErrors, err)
			break
		}
		if err := os.WriteFile(change.Path, []byte(change.After), 0o644); err != nil {
			writeErrors = append(writeErrors, fmt.Errorf("write fixed file %q: %w", change.Path, err))
			continue
		}
		result.ConfirmedPaths = append(result.ConfirmedPaths, change.Path)
	}
	return result, errors.Join(writeErrors...)
}

func readPhysicalFileText(path string) (string, bool) {
	fsys := bundled.WrapFS(osvfs.FS())
	text, ok := fsys.ReadFile(path)
	if !ok {
		return "", false
	}
	return utils.RestoreSourceBOM(fsys, path, text), true
}
