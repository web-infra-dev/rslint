package server

import (
	"context"
	"errors"
	"strings"
	"testing"

	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rules"
)

func TestPluginConfigResolverUsesAPIRoutingOverride(t *testing.T) {
	const (
		configDirectory = "/repo"
		sourcePath      = "/repo/src/a.ts"
		routingKey      = "opaque-api-config-key"
	)
	configMap := map[string]rslintconfig.RslintConfig{
		configDirectory: {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
	}
	lintResolver := configLint.NewResolver(configLint.ResolverOptions{
		ConfigsByOwner: configMap,
		Catalog:        rules.All(),
		TargetsBySourcePath: map[string]target.File{
			sourcePath: {
				PathIdentity: rslintconfig.PathIdentity{
					Path:          sourcePath,
					CanonicalPath: sourcePath,
				},
				ConfigDirectory: configDirectory,
			},
		},
		PathSpaces: rslintconfig.NewPathSpaceSnapshot(configMap, nil),
	})
	resolver := eslintPluginConfigResolver{
		lintResolver:           lintResolver,
		pluginConfigKeyByOwner: map[string]string{configDirectory: routingKey},
	}

	if got := resolver.resolve(sourcePath).ConfigKey; got != routingKey {
		t.Errorf("routing key = %q, want API override %q", got, routingKey)
	}
}

func TestReportEslintPluginDispatchOutcome(t *testing.T) {
	var stderr strings.Builder
	writeEslintPluginDispatchOutcome(&stderr, linter.EslintPluginDispatchOutcome{
		DispatchError: errors.New("transport closed"),
	})
	if want := "rslint: eslint-plugin lint error: transport closed\n"; stderr.String() != want {
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
