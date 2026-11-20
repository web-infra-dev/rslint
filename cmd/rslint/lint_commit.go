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

// An in-place write must agree with the final text verified for every
// selected hard link to that file, including names with no net edit.
func (committer cliFinalChangeCommitter) checkHardLinkConflicts(ctx context.Context, changes []linter.FileChange, destinations []string) error {
	if len(committer.canonicalByPath) < 2 {
		return nil
	}
	type changedFile struct {
		change *linter.FileChange
		info   os.FileInfo
	}
	// Size and modification time only narrow comparisons. SameFile decides
	// identity without assuming a platform-specific layout for FileInfo.Sys.
	type stamp struct{ size, modified int64 }
	stampOf := func(info os.FileInfo) stamp { return stamp{info.Size(), info.ModTime().UnixNano()} }
	byStamp := make(map[stamp][]changedFile, len(changes))
	byPath := make(map[string]changedFile, len(changes))
	for index := range changes {
		if err := ctx.Err(); err != nil {
			return err
		}
		change := &changes[index]
		info, err := statFixFile(destinations[index])
		if err != nil {
			return fmt.Errorf("stat fix target %q: %w", change.Path, err)
		}
		file := changedFile{change, info}
		byPath[rslintconfig.ExactPathID(change.Path)] = file
		key := stampOf(info)
		byStamp[key] = append(byStamp[key], file)
	}
	var conflict [2]string
	for path, physical := range committer.canonicalByPath {
		if err := ctx.Err(); err != nil {
			return err
		}
		current := byPath[path]
		info := current.info
		if info == nil {
			var err error
			info, err = statFixFile(physical)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return fmt.Errorf("stat selected lint target %q: %w", path, err)
			}
		}
		for _, fixed := range byStamp[stampOf(info)] {
			if current.change == fixed.change || !os.SameFile(info, fixed.info) {
				continue
			}
			expected := fixed.change.Before
			if current.change != nil {
				expected = current.change.After
			}
			if expected != fixed.change.After {
				pair := [2]string{min(fixed.change.Path, path), max(fixed.change.Path, path)}
				if conflict[0] == "" || slices.Compare(pair[:], conflict[:]) < 0 {
					conflict = pair
				}
			}
		}
	}
	if conflict[0] != "" {
		return fmt.Errorf("hard-linked lint targets %q and %q require different final contents", conflict[0], conflict[1])
	}
	return nil
}

// On Windows a hard link's directory entry can have stale size and
// timestamps. A handle supplies the metadata for the underlying file.
func statFixFile(path string) (os.FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	return info, errors.Join(err, file.Close())
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
