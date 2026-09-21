package loader

import (
	"errors"
	"fmt"
	"runtime"
	"sync"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs"
	"github.com/web-infra-dev/rslint/internal/config/target"
)

// ProjectCandidate describes a project without reading its configuration or
// acquiring its Program. Service discovery supplies a single selected candidate
// with SourceReferences enabled; explicit projects retain their authored options.
type ProjectCandidate struct {
	ConfigPath       string
	SourceReferences bool
}

// ProjectSelectionRequest supplies already-resolved candidate order and frozen
// targets from one filesystem generation. Adapters own Program construction,
// reuse and invalidation; they do not decide target membership or fallback.
type ProjectSelectionRequest struct {
	Targets          []target.File
	CandidateIndexes [][]int
	Candidates       []ProjectCandidate
	FS               vfs.FS
	SingleThreaded   bool
	// Metadata may be prefetched concurrently. Only the ordered selection may
	// report its errors. A nil result without an error means unavailable.
	Metadata func(index int) (*tsoptions.ParsedCommandLine, error)
	// Program must acquire the generation described by the selected metadata.
	Program func(index int, parsed *tsoptions.ParsedCommandLine) (*compiler.Program, error)
}

// ProjectSourceSelection is the configured source for one input target. A
// CandidateIndex of -1 leaves source-only parsing to the integration adapter.
type ProjectSourceSelection struct {
	CandidateIndex int
	Program        *compiler.Program
	SourceFile     *ast.SourceFile
	DirectRoot     bool
}

type projectSelectionSlot struct {
	parseOnce      sync.Once
	config         *tsoptions.ParsedCommandLine
	parseErr       error
	exactRoots     map[string]struct{}
	canonicalOnce  sync.Once
	canonicalRoots map[string]struct{}
	buildOnce      sync.Once
	program        *compiler.Program
	buildErr       error
	lookupOnce     sync.Once
	lookupMu       sync.Mutex
	lookup         *programFileIndex
}

type projectSelection struct {
	request ProjectSelectionRequest
	slots   []projectSelectionSlot
	// Existing target/source identity machinery is shared across root probes.
	// Only this request owns the index; no editor or loader cache retains it.
	rootIdentityMu sync.Mutex
	rootIdentities *programFileIndex
}

func (selection *projectSelection) metadata(index int) (*projectSelectionSlot, error) {
	slot := &selection.slots[index]
	slot.parseOnce.Do(func() {
		if selection.request.Metadata == nil {
			slot.parseErr = errors.New("project metadata loader is unavailable")
			return
		}
		slot.config, slot.parseErr = selection.request.Metadata(index)
		if slot.parseErr != nil || slot.config == nil {
			return
		}
		slot.exactRoots = make(map[string]struct{}, len(slot.config.FileNames()))
		for _, file := range slot.config.FileNames() {
			slot.exactRoots[exactPathID(tspath.NormalizePath(file))] = struct{}{}
		}
	})
	if slot.parseErr != nil {
		return nil, fmt.Errorf("parse TypeScript config %q: %w", selection.request.Candidates[index].ConfigPath, slot.parseErr)
	}
	return slot, nil
}

func (selection *projectSelection) canonicalRoots(slot *projectSelectionSlot) map[string]struct{} {
	slot.canonicalOnce.Do(func() {
		selection.rootIdentityMu.Lock()
		defer selection.rootIdentityMu.Unlock()
		identities := selection.rootIdentities
		slot.canonicalRoots = make(map[string]struct{}, len(slot.exactRoots))
		var unknown []string
		var unknownIDs []string
		for _, file := range slot.config.FileNames() {
			file = tspath.NormalizePath(file)
			id := exactPathID(file)
			if canonical, known := identities.canonicalBySourcePath[id]; known {
				slot.canonicalRoots[canonical] = struct{}{}
			} else {
				unknown = append(unknown, file)
				unknownIDs = append(unknownIDs, id)
			}
		}
		for index, canonical := range identities.canonicalSourcePathIDs(unknown) {
			identities.canonicalBySourcePath[unknownIDs[index]] = canonical
			slot.canonicalRoots[canonical] = struct{}{}
		}
	})
	return slot.canonicalRoots
}

// matchRoots checks exact and physical roots before advancing to the next
// candidate, regardless of which metadata prefetch completed first.
// All targets seed identity lookup, including targets in other candidate groups.
func (selection *projectSelection) matchRoots(index int, pending []int, owners []int, enqueueBuild func(int)) ([]int, error) {
	slot, err := selection.metadata(index)
	if err != nil || slot.config == nil {
		return pending, err
	}
	enqueued := false
	accept := func(targetIndex int) {
		owners[targetIndex] = index
		if enqueueBuild != nil && !enqueued && selection.supportsParsedTarget(index, slot.config, selection.request.Targets[targetIndex]) {
			// This target has a confirmed owner. Its Program need not wait for
			// the remaining targets' physical root identities to be resolved.
			enqueueBuild(index)
			enqueued = true
		}
	}
	unresolved := pending[:0]
	for _, targetIndex := range pending {
		file := selection.request.Targets[targetIndex]
		if _, exact := slot.exactRoots[exactPathID(file.Path)]; exact {
			accept(targetIndex)
		} else {
			unresolved = append(unresolved, targetIndex)
		}
	}
	pending = unresolved
	if len(pending) == 0 || selection.request.FS == nil || len(slot.exactRoots) == 0 {
		return pending, nil
	}
	canonicalRoots := selection.canonicalRoots(slot)
	unresolved = pending[:0]
	for _, targetIndex := range pending {
		file := selection.request.Targets[targetIndex]
		_, found := canonicalRoots[exactPathID(file.CanonicalPath)]
		if file.CanonicalPath != "" && found {
			accept(targetIndex)
		} else {
			unresolved = append(unresolved, targetIndex)
		}
	}
	return unresolved, nil
}

func (selection *projectSelection) supportsTarget(index int, file target.File) (bool, error) {
	slot, err := selection.metadata(index)
	if err != nil || slot.config == nil {
		return false, err
	}
	return selection.supportsParsedTarget(index, slot.config, file), nil
}

func (selection *projectSelection) supportsParsedTarget(index int, parsed *tsoptions.ParsedCommandLine, file target.File) bool {
	caseSensitive := selection.request.FS == nil || selection.request.FS.UseCaseSensitiveFileNames()
	return selection.request.Candidates[index].SourceReferences ||
		projectSupportsTarget(parsed.CompilerOptions(), file, caseSensitive)
}

func (selection *projectSelection) build(index int) error {
	slot := &selection.slots[index]
	slot.buildOnce.Do(func() {
		if selection.request.Program == nil {
			slot.buildErr = errors.New("project Program loader is unavailable")
			return
		}
		metadata, err := selection.metadata(index)
		if err != nil || metadata.config == nil {
			slot.buildErr = err
			return
		}
		slot.program, slot.buildErr = selection.request.Program(index, metadata.config)
	})
	if slot.buildErr != nil {
		return fmt.Errorf("create TypeScript Program from %q: %w", selection.request.Candidates[index].ConfigPath, slot.buildErr)
	}
	return nil
}

func (selection *projectSelection) source(index int, file target.File) *ast.SourceFile {
	slot := &selection.slots[index]
	if slot.program == nil {
		return nil
	}
	// Editor Programs can carry internal construction options. Another target
	// causing acquisition must not broaden this target's authored eligibility.
	if !selection.supportsParsedTarget(index, slot.config, file) {
		return nil
	}
	slot.lookupOnce.Do(func() {
		slot.lookup = newProgramFileIndex([]*compiler.Program{slot.program},
			selection.request.Targets, selection.request.FS, selection.request.SingleThreaded)
	})
	slot.lookupMu.Lock()
	defer slot.lookupMu.Unlock()
	return slot.lookup.sourceFileForTarget([]int{0}, 0, file)
}

func (selection *projectSelection) missingServiceSource(index int, file target.File) error {
	return fmt.Errorf("project root %q from %q was absent from its TypeScript Program",
		file.Path, selection.request.Candidates[index].ConfigPath)
}

// queueDirectBuilds overlaps confirmed root construction with metadata still
// needed by other targets. It never acquires an unselected candidate's Program.
// The ordered consumer reports errors after all submitted builds have joined.
func (selection *projectSelection) queueDirectBuilds() (enqueue func(int), wait func()) {
	workersN := min(runtime.GOMAXPROCS(0), len(selection.slots))
	if selection.request.SingleThreaded || workersN <= 1 {
		return nil, nil
	}
	jobs := make(chan int, workersN)
	queued := make([]sync.Once, len(selection.slots))
	var workers sync.WaitGroup
	workers.Add(workersN)
	for range workersN {
		go func() {
			defer workers.Done()
			for index := range jobs {
				_ = selection.build(index)
			}
		}()
	}
	return func(index int) {
			queued[index].Do(func() { jobs <- index })
		}, func() {
			close(jobs)
			workers.Wait()
		}
}

func forEachSelectedProject(singleThreaded bool, indexes []int, task func(int)) {
	workersN := min(runtime.GOMAXPROCS(0), len(indexes))
	if singleThreaded || workersN <= 1 {
		for _, index := range indexes {
			task(index)
		}
		return
	}
	jobs := make(chan int, workersN)
	var workers sync.WaitGroup
	workers.Add(workersN)
	for range workersN {
		go func() {
			defer workers.Done()
			for index := range jobs {
				task(index)
			}
		}()
	}
	for _, index := range indexes {
		jobs <- index
	}
	close(jobs)
	workers.Wait()
}

func runTargetProjectTasks(groups []projectTargetGroup, singleThreaded bool, task func(projectTargetGroup) error) error {
	if singleThreaded || runtime.GOMAXPROCS(0) <= 1 || len(groups) <= 1 {
		for _, group := range groups {
			if err := task(group); err != nil {
				return err
			}
		}
		return nil
	}
	indexes := make([]int, len(groups))
	errs := make([]error, len(groups))
	for index := range indexes {
		indexes[index] = index
	}
	forEachSelectedProject(singleThreaded, indexes, func(index int) { errs[index] = task(groups[index]) })
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}

// prefetchRootMetadata restores the declaration-prefix concurrency used by
// focused lint. The nearest containing directory is only a latency hint, never
// evidence of ownership. The ordered consumer can use each ready slot without
// waiting for the whole prefix. Its group must join this work before returning.
func (selection *projectSelection) prefetchRootMetadata(indexes, pending []int) func() {
	caseSensitive := selection.request.FS == nil || selection.request.FS.UseCaseSensitiveFileNames()
	options := tspath.ComparePathsOptions{UseCaseSensitiveFileNames: caseSensitive}
	end := 1 // The first candidate has already been consumed.
	for _, targetIndex := range pending {
		file := selection.request.Targets[targetIndex]
		predicted, longestDirectory := -1, -1
		for position, index := range indexes {
			directory := tspath.GetDirectoryPath(selection.request.Candidates[index].ConfigPath)
			if len(directory) > longestDirectory && tspath.ContainsPath(directory, file.Path, options) {
				predicted, longestDirectory = position, len(directory)
			}
		}
		end = max(end, predicted+1)
	}
	if end <= 2 {
		// At most one unread slot offers no metadata parallelism. Leave it
		// to the ordered consumer instead of adding a goroutine and a join.
		return nil
	}
	prefix := indexes[1:end]
	done := make(chan struct{})
	go func() {
		defer close(done)
		forEachSelectedProject(false, prefix, func(index int) {
			// Store failures in their slots; only a reached candidate can fail
			// the request. Prefetch never acquires a Program or selects an owner.
			_, _ = selection.metadata(index)
		})
	}()
	return func() { <-done }
}

// SelectProjectSources applies one project selection policy for CLI, API and
// editor requests. It never changes targets, reports unused metadata failures,
// or constructs source-only Programs. Metadata-root priority, extension
// eligibility, actual source membership and service errors are decided here.
func SelectProjectSources(request ProjectSelectionRequest) ([]ProjectSourceSelection, error) {
	if len(request.CandidateIndexes) != len(request.Targets) {
		return nil, errors.New("project candidate lists must match the target count")
	}
	candidatesByTarget := make(map[target.File][]int, len(request.Targets))
	result := make([]ProjectSourceSelection, len(request.Targets))
	owners := make([]int, len(request.Targets))
	for index, file := range request.Targets {
		result[index].CandidateIndex = -1
		owners[index] = -1
		for _, candidate := range request.CandidateIndexes[index] {
			if candidate < 0 || candidate >= len(request.Candidates) {
				return nil, fmt.Errorf("invalid project candidate %d for %q", candidate, file.Path)
			}
		}
		candidatesByTarget[file] = request.CandidateIndexes[index]
	}
	selection := &projectSelection{
		request:        request,
		slots:          make([]projectSelectionSlot, len(request.Candidates)),
		rootIdentities: newProgramFileIndex(nil, request.Targets, request.FS, request.SingleThreaded),
	}
	selection.rootIdentities.initialize()
	groups := groupTargetsByProjects(request.Targets, candidatesByTarget, func(string) []int { return nil })
	enqueueBuild, waitForBuilds := selection.queueDirectBuilds()
	err := func() error {
		if waitForBuilds != nil {
			defer waitForBuilds()
		}
		return runTargetProjectTasks(groups, request.SingleThreaded, func(group projectTargetGroup) error {
			pending := append([]int(nil), group.targetIndexes...)
			for position, candidate := range group.projectIndexes {
				if len(pending) == 0 {
					break
				}
				// Give the first candidate a chance to finish without lookahead.
				if position == 1 && !request.SingleThreaded && runtime.GOMAXPROCS(0) > 1 {
					if waitForMetadata := selection.prefetchRootMetadata(group.projectIndexes, pending); waitForMetadata != nil {
						defer waitForMetadata()
					}
				}
				var err error
				pending, err = selection.matchRoots(candidate, pending, owners, enqueueBuild)
				if err != nil {
					return err
				}
			}
			return nil
		})
	}()
	if err != nil {
		return nil, err
	}

	var direct []int
	seen := make([]bool, len(request.Candidates))
	for targetIndex, candidate := range owners {
		if candidate < 0 {
			continue
		}
		supported, err := selection.supportsTarget(candidate, request.Targets[targetIndex])
		if err != nil {
			return nil, err
		}
		if !supported {
			owners[targetIndex] = -1
			continue
		}
		if !seen[candidate] {
			seen[candidate] = true
			direct = append(direct, candidate)
		}
	}
	if enqueueBuild == nil {
		forEachSelectedProject(request.SingleThreaded, direct, func(index int) { _ = selection.build(index) })
	}
	for _, index := range direct {
		if err := selection.build(index); err != nil {
			return nil, err
		}
	}
	for targetIndex, candidate := range owners {
		if candidate < 0 {
			continue
		}
		file := request.Targets[targetIndex]
		if source := selection.source(candidate, file); source != nil {
			result[targetIndex] = ProjectSourceSelection{candidate, selection.slots[candidate].program, source, true}
		} else if request.Candidates[candidate].SourceReferences {
			return nil, selection.missingServiceSource(candidate, file)
		}
	}

	// A rejected first metadata root falls back in declaration order. Do not
	// promote a later direct root over an earlier project's imported source.
	err = runTargetProjectTasks(groups, request.SingleThreaded, func(group projectTargetGroup) error {
		var pending []int
		for _, targetIndex := range group.targetIndexes {
			if result[targetIndex].CandidateIndex < 0 {
				pending = append(pending, targetIndex)
			}
		}
		for _, candidate := range group.projectIndexes {
			if len(pending) == 0 {
				break
			}
			supported := false
			for _, targetIndex := range pending {
				var err error
				supported, err = selection.supportsTarget(candidate, request.Targets[targetIndex])
				if err != nil {
					return err
				}
				if supported {
					break
				}
			}
			if !supported {
				continue
			}
			if err := selection.build(candidate); err != nil {
				return err
			}
			unresolved := pending[:0]
			for _, targetIndex := range pending {
				file := request.Targets[targetIndex]
				if source := selection.source(candidate, file); source != nil {
					result[targetIndex] = ProjectSourceSelection{candidate, selection.slots[candidate].program, source, false}
				} else if request.Candidates[candidate].SourceReferences {
					return selection.missingServiceSource(candidate, file)
				} else {
					unresolved = append(unresolved, targetIndex)
				}
			}
			pending = unresolved
		}
		return nil
	})
	return result, err
}
