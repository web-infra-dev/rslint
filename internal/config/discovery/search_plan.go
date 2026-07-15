package discovery

import (
	"errors"
	"fmt"
	"math"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
)

// ExplicitFileSearch is an input that resolved to an existing regular file.
// Path intentionally preserves the lexical path used by the caller. Stat may
// follow a final symlink, but planning never replaces Path with its real path.
type ExplicitFileSearch struct {
	Path     string
	RawInput string
	Order    int
}

// GlobSearch describes one config-aware filesystem walk. Patterns are
// absolute lexical patterns while RawPatterns retain the corresponding
// normalized user inputs for diagnostics.
type GlobSearch struct {
	ID          uint32
	BasePath    string
	Patterns    []string
	RawPatterns []string
	Order       int
}

// SearchPlan is the ordered, stat-classified form of lint path inputs.
// Searches with the same lexical base are merged. Nested bases are deliberately
// retained because each base is an independently admitted ESLint search root.
type SearchPlan struct {
	ExplicitFiles []ExplicitFileSearch
	GlobSearches  []GlobSearch
}

// SearchPlanOptions controls the input behavior that ESLint exposes around
// findFiles().
type SearchPlanOptions struct {
	// GlobInputPaths permits nonexistent glob-shaped inputs to become searches.
	// When false, such an input is treated as a missing literal path.
	GlobInputPaths bool
	// ErrorOnUnmatchedPattern reports the first missing literal input. Glob
	// searches themselves report unmatched patterns after walking.
	ErrorOnUnmatchedPattern bool
	// SingleThreaded disables the stat fan-out used for independent inputs.
	SingleThreaded bool
}

var ErrNoFilesFound = errors.New("no files found")

// NoFilesFoundError is returned for a literal input that does not exist. It is
// intentionally distinct from a glob search that exists but yields no files;
// the latter can only be decided by the search walker after ignore processing.
type NoFilesFoundError struct {
	Pattern      string
	GlobDisabled bool
}

func (err *NoFilesFoundError) Error() string {
	if err == nil {
		return ErrNoFilesFound.Error()
	}
	suffix := ""
	if err.GlobDisabled {
		suffix = " (glob was disabled)"
	}
	return fmt.Sprintf("No files matching '%s' were found%s.", err.Pattern, suffix)
}

func (err *NoFilesFoundError) Unwrap() error {
	return ErrNoFilesFound
}

// BuildSearchPlan performs the same stat-first input classification as ESLint
// v10 findFiles(). Existing files and directories always win over glob syntax
// in their names. The function performs no filesystem walk and no realpath
// lookup; both config-aware traversal and unmatched-glob reporting happen in a
// later phase.
func BuildSearchPlan(
	fsys vfs.FS,
	cwd string,
	inputs []string,
	opts SearchPlanOptions,
) (*SearchPlan, error) {
	if fsys == nil {
		return nil, errors.New("search planning requires a filesystem")
	}
	if cwd == "" {
		return nil, errors.New("search planning requires a working directory")
	}

	if runtime.GOOS == "windows" {
		cwd = tspath.NormalizePath(cwd)
	} else {
		cwd = filepath.Clean(cwd)
	}
	plan := &SearchPlan{
		ExplicitFiles: make([]ExplicitFileSearch, 0, len(inputs)),
		// ESLint pre-seeds its search map with cwd. If a later input groups into
		// cwd, that search stays ahead of bases first encountered earlier.
		GlobSearches: []GlobSearch{{BasePath: cwd, Order: math.MaxInt}},
	}
	searchIndexByBase := map[string]int{cwd: 0}
	missing := make([]*NoFilesFoundError, 0)
	rawInputs := make([]string, len(inputs))
	absoluteInputs := make([]string, len(inputs))
	stats := make([]vfs.FileInfo, len(inputs))
	statInput := func(index int) {
		// ESLint stats the original host spelling, then normalizes a separate
		// copy before glob classification and diagnostics.
		rawInputs[index] = tspath.NormalizeSlashes(inputs[index])
		absoluteInputs[index] = resolveSearchInput(cwd, inputs[index])
		stats[index] = fsys.Stat(absoluteInputs[index])
	}
	if opts.SingleThreaded || len(inputs) < 2 {
		for index := range inputs {
			statInput(index)
		}
	} else {
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(inputs))
		for index := range inputs {
			go func() {
				defer waitGroup.Done()
				statInput(index)
			}()
		}
		waitGroup.Wait()
	}

	for order := range inputs {
		rawInput := rawInputs[order]
		absoluteInput := absoluteInputs[order]
		info := stats[order]

		if info != nil {
			switch {
			case info.Mode().IsRegular():
				plan.ExplicitFiles = append(plan.ExplicitFiles, ExplicitFileSearch{
					Path:     absoluteInput,
					RawInput: rawInput,
					Order:    order,
				})
			case info.IsDir():
				appendGlobSearch(
					plan,
					searchIndexByBase,
					absoluteInput,
					// ESLint retains the native spelling as the walk base but
					// normalizes the independently stored glob to POSIX slashes.
					// The distinction is observable for POSIX names containing a
					// backslash, where path.relative() later produces an unmatched
					// sibling-style pattern.
					appendSearchGlobstar(tspath.NormalizeSlashes(absoluteInput)),
					rawInput,
					order,
				)
			}
			continue
		}

		if opts.GlobInputPaths && isGlobPattern(rawInput) {
			basePath := resolveSearchInput(cwd, GlobParent(rawInput))
			absoluteInput = resolveSearchInput(cwd, rawInput)
			appendGlobSearch(
				plan,
				searchIndexByBase,
				basePath,
				absoluteInput,
				rawInput,
				order,
			)
			continue
		}

		missing = append(missing, &NoFilesFoundError{
			Pattern:      rawInput,
			GlobDisabled: !opts.GlobInputPaths && isGlobPattern(rawInput),
		})
	}

	if opts.ErrorOnUnmatchedPattern && len(missing) > 0 {
		return nil, missing[0]
	}
	filteredSearches := plan.GlobSearches[:0]
	for _, search := range plan.GlobSearches {
		if len(search.Patterns) == 0 {
			continue
		}
		search.ID = uint32(len(filteredSearches))
		filteredSearches = append(filteredSearches, search)
	}
	plan.GlobSearches = filteredSearches
	return plan, nil
}

func appendSearchGlobstar(directory string) string {
	if strings.HasSuffix(directory, "/") {
		return directory + "**"
	}
	return directory + "/**"
}

// resolveSearchInput mirrors Node path.resolve for the host platform. On
// POSIX, backslash is a valid filename byte and also Minimatch's escape marker;
// normalizing it before stat would silently target a different path.
func resolveSearchInput(cwd string, input string) string {
	if runtime.GOOS == "windows" {
		return tspath.ResolvePath(cwd, tspath.NormalizeSlashes(input))
	}
	if filepath.IsAbs(input) {
		return filepath.Clean(input)
	}
	return filepath.Join(cwd, input)
}

func appendGlobSearch(
	plan *SearchPlan,
	searchIndexByBase map[string]int,
	basePath string,
	pattern string,
	rawPattern string,
	order int,
) {
	if index, ok := searchIndexByBase[basePath]; ok {
		search := &plan.GlobSearches[index]
		if len(search.Patterns) == 0 {
			search.Order = order
		}
		search.Patterns = append(search.Patterns, pattern)
		search.RawPatterns = append(search.RawPatterns, rawPattern)
		return
	}

	index := len(plan.GlobSearches)
	searchIndexByBase[basePath] = index
	plan.GlobSearches = append(plan.GlobSearches, GlobSearch{
		ID:          uint32(index),
		BasePath:    basePath,
		Patterns:    []string{pattern},
		RawPatterns: []string{rawPattern},
		Order:       order,
	})
}
