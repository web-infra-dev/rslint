package scopeanalysis

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/scope"
)

// Provider owns scope analysis state for one file.
type Provider struct{ cache *scope.Cache }

// For returns a provider backed by the RuleContext's file cache.
func For(ctx rule.RuleContext) Provider { return Provider{cache: ctx.ScopeAnalysis()} }

// Get shares read-only scope graphs without widening filtered reference requests.
func Get(ctx rule.RuleContext, options scope.Options) *scope.Manager { return For(ctx).Get(options) }

// Get returns the graph for the exact options.
func (provider Provider) Get(options scope.Options) *scope.Manager {
	return provider.cache.Get(options)
}

// References returns a graph containing at least the requested names.
func References(ctx rule.RuleContext, names map[string]struct{}) *scope.Manager {
	return For(ctx).References(names)
}

// References returns a graph containing at least the requested names.
func (provider Provider) References(names map[string]struct{}) *scope.Manager {
	return provider.cache.References(names)
}

// Declarations returns a graph whose declaration tree is complete.
// Its reference slices may be absent or filtered and must not be inspected.
func Declarations(ctx rule.RuleContext) *scope.Manager { return For(ctx).Declarations() }

// Declarations returns a graph whose declaration tree is complete.
// Its reference slices may be absent or filtered and must not be inspected.
func (provider Provider) Declarations() *scope.Manager { return provider.cache.Declarations() }
