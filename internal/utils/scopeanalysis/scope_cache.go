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

// Provider owns scope analysis state for one file.
type Provider struct {
	ctx   rule.RuleContext
	state *fileState
}

// For returns a provider backed by the RuleContext's file cache.
func For(ctx rule.RuleContext) Provider {
	return Provider{ctx: ctx, state: stateFor(ctx)}
}

// Get shares read-only scope graphs without widening filtered reference requests.
func Get(ctx rule.RuleContext, options scope.Options) *scope.Manager {
	return For(ctx).Get(options)
}

// Get returns the graph for the exact options.
func (provider Provider) Get(options scope.Options) *scope.Manager {
	key := keyFor(options)
	if manager := provider.state.managers[key]; manager != nil {
		return manager
	}
	manager := scope.Build(provider.ctx.SourceFile, options)
	provider.state.managers[key] = manager
	if options.CollectReferences && options.ReferenceNames == nil {
		provider.state.completeReferences = manager
		provider.state.declarations = manager
	} else if provider.state.declarations == nil {
		provider.state.declarations = manager
	}
	return manager
}

// References returns a graph containing at least the requested names.
func References(ctx rule.RuleContext, names map[string]struct{}) *scope.Manager {
	return For(ctx).References(names)
}

// References returns a graph containing at least the requested names.
func (provider Provider) References(names map[string]struct{}) *scope.Manager {
	if provider.state.completeReferences != nil {
		return provider.state.completeReferences
	}
	return provider.Get(scope.Options{CollectReferences: true, ReferenceNames: names})
}

// Declarations returns a graph whose declaration tree is complete.
// Its reference slices may be absent or filtered and must not be inspected.
func Declarations(ctx rule.RuleContext) *scope.Manager {
	return For(ctx).Declarations()
}

// Declarations returns a graph whose declaration tree is complete.
// Its reference slices may be absent or filtered and must not be inspected.
func (provider Provider) Declarations() *scope.Manager {
	if provider.state.declarations != nil {
		return provider.state.declarations
	}
	return provider.Get(scope.Options{})
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
