package main

import (
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/config/target"
)

func explicitConfigCatalogForTest(
	configDirectory string,
	entries rslintconfig.RslintConfig,
) *discovery.ConfigCatalog {
	configDirectory = tspath.NormalizePath(configDirectory)
	return &discovery.ConfigCatalog{
		Configs: map[string]rslintconfig.RslintConfig{
			configDirectory: entries,
		},
		Explicit: true,
	}
}

// Tests that construct authored catalogs supply request observations here;
// production adapters always receive these from discovery.
func prepareCommandCatalogForTest(cwd string, args *lintArgs) {
	catalog := args.ConfigCatalog
	if catalog == nil || catalog.GitPolicy != nil {
		return
	}
	fsys := args.FS
	if fsys == nil {
		fsys = osvfs.FS()
	}
	paths := args.AllowFiles
	if paths == nil {
		for _, scope := range catalog.Scopes {
			paths = append(paths, scope.ExplicitFiles...)
		}
	}
	if paths != nil {
		catalog.TargetFiles = make([]rslintconfig.PathIdentity, len(paths))
		for index, path := range paths {
			catalog.TargetFiles[index] = rslintconfig.FreezeFileIdentity(path, fsys)
		}
	}
	directories := args.AllowDirs
	if len(paths) == 0 && len(directories) == 0 {
		directories = []string{cwd}
	}
	catalog.TargetDirectories = rslintconfig.ResolveDirectoryIdentities(directories, fsys)
	var entries rslintconfig.RslintConfig
	configDirectory := cwd
	if catalog.Explicit {
		for owner, config := range catalog.Configs {
			configDirectory, entries = owner, config
		}
	}
	catalog.GitPolicy = rslintconfig.CollectGitPolicy(entries, configDirectory, fsys, catalog.TargetFiles, catalog.TargetDirectories)
}

func withGitPolicyForTest(request target.Request) target.Request {
	if request.GitPolicy != nil {
		return request
	}
	root := request.ScanRoot
	if root == "" {
		root = request.ConfigDirectory
	}
	if request.Files == nil && request.Directories == nil {
		request.Directories = rslintconfig.ResolveDirectoryIdentities([]string{root}, request.FS)
	}
	request.GitPolicy = rslintconfig.CollectGitPolicy(request.Config, request.ConfigDirectory, request.FS, request.Files, request.Directories)
	return request
}
