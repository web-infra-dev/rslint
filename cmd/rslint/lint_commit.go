package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/utils"
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
	if err := committer.checkHardLinkConflicts(ctx, changes, destinations); err != nil {
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

// An in-place write must agree with the final text verified for every selected
// name for the file. Bind mounts can share a file even when its link count is one.
func (committer cliFinalChangeCommitter) checkHardLinkConflicts(ctx context.Context, changes []linter.FileChange, destinations []string) error {
	if len(committer.canonicalByPath) < 2 {
		return nil
	}
	type changedFile struct {
		change *linter.FileChange
		id     fixFileID
	}
	// The first changed name in lexical order is enough to witness any
	// disagreement, whether another name has a different edit or no edit.
	firstByID := make(map[fixFileID]*linter.FileChange, len(changes))
	byPath := make(map[string]changedFile, len(changes))
	for index := range changes {
		if err := ctx.Err(); err != nil {
			return err
		}
		change := &changes[index]
		if change.Before == change.After {
			continue
		}
		id, err := statFixFile(destinations[index])
		if err != nil {
			return fmt.Errorf("stat fix target %q: %w", change.Path, err)
		}
		byPath[rslintconfig.ExactPathID(change.Path)] = changedFile{change, id}
		if first := firstByID[id]; first == nil || change.Path < first.Path {
			firstByID[id] = change
		}
	}
	var conflict [2]string
	for path, physical := range committer.canonicalByPath {
		if err := ctx.Err(); err != nil {
			return err
		}
		current := byPath[path]
		if current.change == nil {
			var err error
			current.id, err = statFixFile(physical)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return fmt.Errorf("stat selected lint target %q: %w", path, err)
			}
		}
		first := firstByID[current.id]
		if first == nil || (current.change != nil && current.change.After == first.After) {
			continue
		}
		pair := [2]string{min(first.Path, path), max(first.Path, path)}
		if conflict[0] == "" || slices.Compare(pair[:], conflict[:]) < 0 {
			conflict = pair
		}
	}
	if conflict[0] != "" {
		return fmt.Errorf("hard-linked lint targets %q and %q require different final contents", conflict[0], conflict[1])
	}
	return nil
}

func readPhysicalFileText(path string) (string, bool) {
	fsys := bundled.WrapFS(osvfs.FS())
	text, ok := fsys.ReadFile(path)
	if !ok {
		return "", false
	}
	text, err := utils.RestoreSourceBOM(fsys, path, text)
	return text, err == nil
}
