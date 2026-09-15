package scopeanalysis

import (
	"slices"
	"strconv"
	"strings"

	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

type fileStateKey struct{}

type managerKey struct {
	collectReferences bool
	filtered          bool
	names             string
}

type fileState struct {
	managers           map[managerKey]*scope.Manager
	declarations       *scope.Manager
	completeReferences *scope.Manager
}

// Get shares read-only scope graphs without widening filtered reference requests.
func Get(ctx rule.RuleContext, options scope.Options) *scope.Manager {
	state := stateFor(ctx)
	key := keyFor(options)
	if manager := state.managers[key]; manager != nil {
		return manager
	}
	manager := scope.Build(ctx.SourceFile, options)
	state.managers[key] = manager
	if options.CollectReferences && options.ReferenceNames == nil {
		state.completeReferences = manager
		state.declarations = manager
	} else if state.declarations == nil {
		state.declarations = manager
	}
	return manager
}

// References returns a graph containing at least the requested names.
func References(ctx rule.RuleContext, names map[string]struct{}) *scope.Manager {
	state := stateFor(ctx)
	if state.completeReferences != nil {
		return state.completeReferences
	}
	return Get(ctx, scope.Options{CollectReferences: true, ReferenceNames: names})
}

// Declarations returns a graph whose declaration tree is complete.
// Its reference slices may be absent or filtered and must not be inspected.
func Declarations(ctx rule.RuleContext) *scope.Manager {
	state := stateFor(ctx)
	if state.declarations != nil {
		return state.declarations
	}
	return Get(ctx, scope.Options{})
}

func stateFor(ctx rule.RuleContext) *fileState {
	return rule.CachedByFile(ctx, fileStateKey{}, newFileState)
}

func newFileState() *fileState {
	return &fileState{managers: make(map[managerKey]*scope.Manager)}
}

func keyFor(options scope.Options) managerKey {
	key := managerKey{collectReferences: options.CollectReferences}
	if !options.CollectReferences || options.ReferenceNames == nil {
		return key
	}
	key.filtered = true
	names := make([]string, 0, len(options.ReferenceNames))
	for name := range options.ReferenceNames {
		names = append(names, name)
	}
	slices.Sort(names)
	var encoded strings.Builder
	for _, name := range names {
		encoded.WriteString(strconv.Quote(name))
	}
	key.names = encoded.String()
	return key
}
