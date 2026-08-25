package main

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/typescript-go/shim/tspath"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	configLint "github.com/web-infra-dev/rslint/internal/config/lint"
	"github.com/web-infra-dev/rslint/internal/config/target"
	"github.com/web-infra-dev/rslint/internal/linter"
	"github.com/web-infra-dev/rslint/internal/rule"
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

func TestPluginConfigResolverUsesAPIRoutingOverride(t *testing.T) {
	const (
		configDirectory = "/repo"
		sourcePath      = "/repo/src/a.ts"
		routingKey      = "opaque-api-config-key"
	)
	resolver := eslintPluginConfigResolver{
		lintResolver: newPluginLintResolverForTest(configLint.ResolverOptions{
			ConfigsByOwner: map[string]rslintconfig.RslintConfig{
				configDirectory: {{Rules: rslintconfig.Rules{"no-debugger": "error"}}},
			},
			TargetsBySourcePath: map[string]target.File{
				sourcePath: pluginTargetForTest(sourcePath, configDirectory),
			},
		}),
		pluginConfigDirectoryByOwner: map[string]string{configDirectory: routingKey},
	}

	if got := resolver.resolve(sourcePath).ConfigKey; got != routingKey {
		t.Errorf("routing key = %q, want API override %q", got, routingKey)
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

func TestDispatchEslintPluginRulesAsyncReportsFailureBeforeResultIsObserved(t *testing.T) {
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %v", err)
	}
	previous := os.Stderr
	os.Stderr = writer
	defer func() {
		os.Stderr = previous
		_ = writer.Close()
		_ = reader.Close()
	}()

	type readResult struct {
		line string
		err  error
	}
	lineRead := make(chan readResult, 1)
	go func() {
		line, readErr := bufio.NewReader(reader).ReadString('\n')
		lineRead <- readResult{line: strings.ReplaceAll(line, "\r\n", "\n"), err: readErr}
	}()

	dispatchError := errors.New("transport closed")
	diagnosticsCh := dispatchEslintPluginRulesAsync(
		context.Background(),
		func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			return nil, dispatchError
		},
		[]linter.EslintPluginFileInput{{
			Path: "/repo/a.ts",
			Rules: []rule.ConfiguredRule{{
				Name:               "external/rule",
				Severity:           rule.SeverityError,
				IsEslintPluginRule: true,
			}},
		}},
		false,
		linter.SuggestionsModeOff,
		nil,
	)

	select {
	case result := <-lineRead:
		if result.err != nil {
			t.Fatalf("read stderr: %v", result.err)
		}
		if want := "rslint: eslint-plugin lint error: transport closed\n"; result.line != want {
			t.Fatalf("stderr = %q, want %q", result.line, want)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("dispatch failure was not written before the result channel was observed")
	}

	diagnostics := <-diagnosticsCh
	if len(diagnostics) != 1 || diagnostics[0].RuleName != "rslint/plugin-lint-error" {
		t.Fatalf("diagnostics = %v, want the dispatch failure diagnostic", diagnostics)
	}
}

func TestDispatchEslintPluginRulesAsyncKeepsCancellationSilent(t *testing.T) {
	var diagnostics []rule.RuleDiagnostic
	stderr := captureStderrForTest(t, func() {
		diagnostics = <-dispatchEslintPluginRulesAsync(
			context.Background(),
			func(context.Context, linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
				return nil, context.Canceled
			},
			[]linter.EslintPluginFileInput{{
				Path: "/repo/a.ts",
				Rules: []rule.ConfiguredRule{{
					Name:               "external/rule",
					Severity:           rule.SeverityError,
					IsEslintPluginRule: true,
				}},
			}},
			false,
			linter.SuggestionsModeOff,
			nil,
		)
	})
	if len(diagnostics) != 0 {
		t.Errorf("cancellation diagnostics = %v, want none", diagnostics)
	}
	if stderr != "" {
		t.Errorf("cancellation stderr = %q, want none", stderr)
	}
}

func captureStderrForTest(t *testing.T, action func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %v", err)
	}
	previous := os.Stderr
	os.Stderr = writer
	action()
	os.Stderr = previous
	if err := writer.Close(); err != nil {
		t.Fatalf("close stderr writer: %v", err)
	}
	output, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stderr reader: %v", err)
	}
	return strings.ReplaceAll(string(output), "\r\n", "\n")
}

func pluginTargetForTest(path string, configDirectory string) target.File {
	return target.File{
		PathIdentity:    rslintconfig.PathIdentity{Path: path, CanonicalPath: path},
		ConfigDirectory: configDirectory,
	}
}
