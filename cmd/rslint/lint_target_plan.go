package main

import (
	"fmt"

	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
)

type resolvedLintTarget struct {
	Path           string
	CanonicalPath  string
	OwnerConfigDir string
}

type lintTargetPlan struct {
	Targets []resolvedLintTarget
}

func configsForLintTargetPlan(
	configMap map[string]rslintconfig.RslintConfig,
	plan lintTargetPlan,
) map[string]rslintconfig.RslintConfig {
	if len(configMap) == 0 || len(plan.Targets) == 0 {
		return nil
	}
	active := make(map[string]rslintconfig.RslintConfig)
	for _, target := range plan.Targets {
		if entries, ok := configMap[target.OwnerConfigDir]; ok {
			active[target.OwnerConfigDir] = entries
		}
	}
	return active
}

func resolveLintTargetPlan(
	configMap map[string]rslintconfig.RslintConfig,
	rslintConfig rslintconfig.RslintConfig,
	currentDirectory string,
	configTargetScopes map[string]rslintconfig.LintDiscoveryScope,
	fsys vfs.FS,
	allowFiles []string,
	allowDirs []string,
	singleThreaded bool,
) (lintTargetPlan, error) {
	type targetWithOwner struct {
		path          string
		canonicalPath string
		owner         string
	}
	var targetFiles []targetWithOwner
	if configMap != nil {
		for _, target := range rslintconfig.DiscoverLintTargetsMultiConfig(
			configMap,
			configTargetScopes,
			fsys,
			allowFiles,
			allowDirs,
			singleThreaded,
		) {
			targetFiles = append(targetFiles, targetWithOwner{
				path:          target.Path,
				canonicalPath: target.CanonicalPath,
				owner:         target.ConfigDirectory,
			})
		}
	} else {
		for _, target := range rslintconfig.DiscoverLintTargets(rslintConfig, currentDirectory, fsys, allowFiles, allowDirs, singleThreaded) {
			targetFiles = append(targetFiles, targetWithOwner{
				path:          target.Path,
				canonicalPath: target.CanonicalPath,
				owner:         currentDirectory,
			})
		}
	}

	plan := lintTargetPlan{Targets: make([]resolvedLintTarget, 0, len(targetFiles))}
	seenCanonical := make(map[string]resolvedLintTarget, len(targetFiles))
	for _, discovered := range targetFiles {
		targetPath := discovered.path
		ownerConfigDir := discovered.owner
		canonicalPath := tspath.NormalizePath(targetPath)
		if discovered.canonicalPath != "" {
			canonicalPath = tspath.NormalizePath(discovered.canonicalPath)
		}
		if discovered.canonicalPath == "" && fsys != nil {
			if realPath := fsys.Realpath(canonicalPath); realPath != "" {
				canonicalPath = tspath.NormalizePath(realPath)
			}
		}
		canonicalKey := exactFilesystemPathID(canonicalPath)
		target := resolvedLintTarget{
			Path:           tspath.NormalizePath(targetPath),
			CanonicalPath:  canonicalPath,
			OwnerConfigDir: tspath.NormalizePath(ownerConfigDir),
		}
		if existing, exists := seenCanonical[canonicalKey]; exists {
			if canonicalFilesystemPathID(existing.OwnerConfigDir, fsys) != canonicalFilesystemPathID(target.OwnerConfigDir, fsys) {
				return lintTargetPlan{}, fmt.Errorf(
					"lint target aliases %q and %q resolve to the same file but are governed by different configs %q and %q",
					existing.Path,
					target.Path,
					existing.OwnerConfigDir,
					target.OwnerConfigDir,
				)
			}
			continue
		}
		seenCanonical[canonicalKey] = target
		plan.Targets = append(plan.Targets, target)
	}
	return plan, nil
}

func preferredCallerTargetPaths(plan lintTargetPlan) map[string]string {
	if len(plan.Targets) == 0 {
		return nil
	}
	preferred := make(map[string]string, len(plan.Targets))
	for _, target := range plan.Targets {
		canonicalID := exactFilesystemPathID(target.CanonicalPath)
		if _, exists := preferred[canonicalID]; !exists {
			preferred[canonicalID] = target.Path
		}
	}
	return preferred
}
