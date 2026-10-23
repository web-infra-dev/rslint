package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/utils"
	"github.com/web-infra-dev/rslint/internal/utils/pathutil"
)

// cliFinalChangeCommitter performs the CLI's one terminal disk projection.
// All fix rounds happen in linter-owned memory before this adapter is called.
type cliFinalChangeCommitter struct {
	canonicalByPath map[string]string
}

func newCLIFinalChangeCommitter(files []target.File) cliFinalChangeCommitter {
	paths := make(map[string]string, len(files))
	for _, file := range files {
		paths[rslintconfig.ExactPathID(file.Path)] = tspath.NormalizePath(file.CanonicalPath)
	}
	return cliFinalChangeCommitter{canonicalByPath: paths}
}

func (committer cliFinalChangeCommitter) CommitFinalChanges(
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
	destinations := make([]string, len(changes))
	for index, change := range changes {
		if err := ctx.Err(); err != nil {
			preflightErrors = append(preflightErrors, err)
			break
		}
		physical, selected := committer.canonicalByPath[rslintconfig.ExactPathID(change.Path)]
		if !selected {
			preflightErrors = append(preflightErrors, fmt.Errorf("fix target %q is not a selected lint target", change.Path))
			continue
		}
		destinations[index] = physical
		current, ok := readPhysicalFileText(physical)
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
	for index, change := range changes {
		if err := ctx.Err(); err != nil {
			writeErrors = append(writeErrors, err)
			break
		}
		if err := os.WriteFile(destinations[index], []byte(change.After), 0o644); err != nil {
			writeErrors = append(writeErrors, fmt.Errorf("write fixed file %q: %w", change.Path, err))
			continue
		}
		result.ConfirmedPaths = append(result.ConfirmedPaths, change.Path)
	}
	return result, errors.Join(writeErrors...)
}

func readPhysicalFileText(path string) (string, bool) {
	fsys := bundled.WrapFS(osvfs.FS())
	text, ok := fsys.ReadFile(pathutil.File(path))
	if !ok {
		return "", false
	}
	return utils.RestoreSourceBOM(fsys, path, text), true
}
