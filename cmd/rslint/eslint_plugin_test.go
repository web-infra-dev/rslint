package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rules"
)

func newPluginLintResolverForTest(options configLint.ResolverOptions) *configLint.Resolver {
	options.Catalog = rules.All()
	configs := options.ConfigsByOwner
	if configs == nil {
		configs = map[string]rslintconfig.RslintConfig{
			options.ConfigDirectory: options.Config,
		}
	}
	options.PathSpaces = rslintconfig.NewPathSpaceSnapshot(configs, options.FS)
	return configLint.NewResolver(options)
}

func TestPluginConfigResolverUsesOwnerAsRoutingKey(t *testing.T) {
	configDirectory := tspath.NormalizePath(`C:\proj`)
	sourcePath := configDirectory + "/src/a.ts"
	configMap := map[string]rslintconfig.RslintConfig{
		configDirectory: {{
			Settings: rslintconfig.Settings{"owner": "go"},
			Rules:    rslintconfig.Rules{"no-debugger": "error"},
		}},
	}
	resolver := eslintPluginConfigResolver{
		lintResolver: newPluginLintResolverForTest(configLint.ResolverOptions{
			ConfigsByOwner: configMap,
			TargetsBySourcePath: map[string]target.File{
				sourcePath: pluginTargetForTest(sourcePath, configDirectory),
			},
		}),
	}

	resolved := resolver.resolve(sourcePath)
	if resolved.ConfigKey != configDirectory {
		t.Errorf("routing key = %q, want owner %q", resolved.ConfigKey, configDirectory)
	}
	if resolved.Settings["owner"] != "go" {
		t.Errorf("settings = %v, want the resolved owner config", resolved.Settings)
	}
}

func TestPluginConfigResolverPreservesResolutionModes(t *testing.T) {
	configMap := map[string]rslintconfig.RslintConfig{
		"/repo": {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
	}

	multi := eslintPluginConfigResolver{
		lintResolver: newPluginLintResolverForTest(configLint.ResolverOptions{
			ConfigsByOwner: configMap,
		}),
	}
	if got := multi.resolve("/elsewhere/a.ts"); got.ConfigKey != "" || got.LanguageOptions != nil || got.Settings != nil {
		t.Errorf("unbound multi-config source = %+v, want empty projection", got)
	}

	single := eslintPluginConfigResolver{
		lintResolver: newPluginLintResolverForTest(configLint.ResolverOptions{
			Config:          configMap["/repo"],
			ConfigDirectory: "/repo",
		}),
	}
	if got := single.resolve("/repo/a.ts").ConfigKey; got != "/repo" {
		t.Errorf("single-config routing key = %q, want /repo", got)
	}
}

func TestReportEslintPluginDispatchOutcome(t *testing.T) {
	var stderr strings.Builder
	writeEslintPluginDispatchOutcome(&stderr, linter.EslintPluginDispatchOutcome{
		Notices: []linter.EslintPluginProtocolNotice{
			{Kind: linter.EslintPluginMissingFileResult, FilePath: "/repo/a.ts"},
			{Kind: linter.EslintPluginUnconfiguredDiagnostic, FilePath: "/repo/b.ts", RuleName: "plugin/extra"},
		},
		DispatchError: errors.New("transport closed"),
	})
	if want := "rslint: plugin-lint returned no result for \"/repo/a.ts\"\n" +
		"rslint: plugin diagnostic for unconfigured rule \"plugin/extra\" in \"/repo/b.ts\"\n" +
		"rslint: eslint-plugin lint error: transport closed\n"; stderr.String() != want {
		t.Fatalf("stderr = %q, want %q", stderr.String(), want)
	}

	stderr.Reset()
	writeEslintPluginDispatchOutcome(&stderr, linter.EslintPluginDispatchOutcome{
		DispatchError: context.Canceled,
	})
	if stderr.Len() != 0 {
		t.Fatalf("cancellation stderr = %q, want none", stderr.String())
	}
}

func pluginTargetForTest(path string, configDirectory string) target.File {
	return target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: path, CanonicalPath: path},
		ConfigDirectory: configDirectory,
	}
}
