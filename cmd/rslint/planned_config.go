package main

import (
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/program/loader"
	"github.com/web-infra-dev/rslint/internal/rule"
)

// plannedLintConfigResolver projects the config snapshot already frozen for
// each lint target. It never re-runs config ownership, files, or ignores after
// Program binding.
type plannedLintConfigResolver struct {
	plan    *rslintconfig.LintProjectPlan
	binding *loader.LoadResult
}

func (resolver plannedLintConfigResolver) targetForFile(filePath string) (*rslintconfig.PlannedLintTarget, bool) {
	if resolver.plan == nil || resolver.binding == nil {
		return nil, false
	}
	targetIndex, ok := resolver.binding.TargetIndexForSourcePath(filePath)
	if !ok || targetIndex < 0 || targetIndex >= len(resolver.plan.Targets) {
		return nil, false
	}
	return &resolver.plan.Targets[targetIndex], true
}

func (resolver plannedLintConfigResolver) ConfigForFile(filePath string) *rslintconfig.MergedConfig {
	target, ok := resolver.targetForFile(filePath)
	if !ok || target.Effective == nil {
		return nil
	}
	return target.Effective.MergedConfig()
}

func (resolver plannedLintConfigResolver) EnabledRulesForFile(filePath string) []rule.ConfiguredRule {
	target, ok := resolver.targetForFile(filePath)
	if !ok || target.Effective == nil {
		return nil
	}
	return target.Effective.EnabledRules()
}

func (resolver plannedLintConfigResolver) OwnerForFile(filePath string) string {
	target, ok := resolver.targetForFile(filePath)
	if !ok {
		return ""
	}
	return target.Target.ConfigDirectory
}
