package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/rules/catalog"
)

// The rule-options validation step runs after configuration is fully
// resolved and before any linting starts: a config with schema-invalid rule
// options must fail fast, report every failure at once, and never produce
// lint diagnostics.
func TestCLIInvalidRuleOptionsFailFastBeforeLinting(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rslint.jsonc"), []byte(`[
		{
			"files": ["*.js"],
			"rules": {
				"no-console": ["error", { "allow": "warn" }],
				"eqeqeq": ["error", "sometimes"],
				"no-debugger": "error"
			}
		}
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
		Config:         "rslint.jsonc",
		Format:         "default",
		NoColor:        true,
		SingleThreaded: true,
	})
	if code != 1 {
		t.Fatalf("expected exit code 1 for invalid rule options, got code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
	// Every failure is reported, not just the first.
	if !strings.Contains(stderr, `invalid options for rule "no-console"`) {
		t.Errorf("expected stderr to name no-console, got %q", stderr)
	}
	if !strings.Contains(stderr, `invalid options for rule "eqeqeq"`) {
		t.Errorf("expected stderr to name eqeqeq, got %q", stderr)
	}
	// Linting never started: the valid no-debugger rule produced no diagnostic.
	if strings.Contains(stdout, "no-debugger") {
		t.Errorf("expected no lint diagnostics before validation passes, stdout=%q", stdout)
	}
}

// The same config with schema-valid options must sail through the validation
// step and lint normally.
func TestCLIValidRuleOptionsLintNormally(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "rslint.jsonc"), []byte(`[
		{
			"files": ["*.js"],
			"rules": {
				"no-console": ["error", { "allow": ["warn"] }],
				"eqeqeq": ["error", "smart"],
				"no-debugger": "error"
			}
		}
	]`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("debugger;\n"), 0o644); err != nil {
		t.Fatalf("write app.js: %v", err)
	}

	code, stdout, stderr := runLintPipelineForTest(t, dir, lintArgs{
		Config:         "rslint.jsonc",
		Format:         "default",
		NoColor:        true,
		SingleThreaded: true,
	})
	if strings.Contains(stderr, "invalid options") {
		t.Fatalf("expected schema validation to pass, stderr=%q", stderr)
	}
	if code != 1 || !strings.Contains(stdout, "no-debugger") {
		t.Fatalf("expected the lint itself to run and flag no-debugger, got code=%d stdout=%q stderr=%q", code, stdout, stderr)
	}
}

func TestValidateResolvedRuleOptionsReturnsNormalizedSingleConfig(t *testing.T) {
	inputOptions := map[string]any{"values": []any{"original"}}
	input := rslintconfig.RslintConfig{{
		Rules: rslintconfig.Rules{
			"unknown-rule": []any{"error", inputOptions},
		},
	}}

	normalizedMap, normalized, messages := validateResolvedRuleOptions(nil, input, catalog.Native())
	if normalizedMap != nil {
		t.Fatalf("single-config mode returned a non-nil config map: %#v", normalizedMap)
	}
	if len(messages) != 0 {
		t.Fatalf("unexpected validation messages: %v", messages)
	}

	normalizedOptions := normalized[0].Rules["unknown-rule"].([]any)[1].(map[string]any)
	normalizedOptions["values"].([]any)[0] = "changed"
	if got := inputOptions["values"].([]any)[0]; got != "original" {
		t.Fatalf("helper returned the input config instead of its normalized copy: %#v", got)
	}
}

func TestValidateResolvedRuleOptionsPreservesMultiConfigMode(t *testing.T) {
	normalizedMap, _, messages := validateResolvedRuleOptions(
		map[string]rslintconfig.RslintConfig{},
		rslintconfig.RslintConfig{{Rules: rslintconfig.Rules{"unused": "error"}}},
		catalog.Native(),
	)
	if normalizedMap == nil || len(normalizedMap) != 0 {
		t.Fatalf("non-nil empty config map changed mode: %#v", normalizedMap)
	}
	if len(messages) != 0 {
		t.Fatalf("unexpected validation messages: %v", messages)
	}

	inputOptions := map[string]any{"values": []any{"original"}}
	inputMap := map[string]rslintconfig.RslintConfig{
		"/workspace/a": {{
			Rules: rslintconfig.Rules{
				"unknown-rule": []any{"error", inputOptions},
			},
		}},
	}
	normalizedMap, _, messages = validateResolvedRuleOptions(inputMap, nil, catalog.Native())
	if len(messages) != 0 {
		t.Fatalf("unexpected validation messages: %v", messages)
	}
	normalizedOptions := normalizedMap["/workspace/a"][0].Rules["unknown-rule"].([]any)[1].(map[string]any)
	normalizedOptions["values"].([]any)[0] = "changed"
	if got := inputOptions["values"].([]any)[0]; got != "original" {
		t.Fatalf("multi-config helper returned an aliased input value: %#v", got)
	}
}

const apiFirstRuleOptionsValidationProcess = "RSLINT_TEST_API_FIRST_RULE_OPTIONS_VALIDATION"

func TestHandleLintFirstRequestValidatesRuleOptions(t *testing.T) {
	if os.Getenv(apiFirstRuleOptionsValidationProcess) == "1" {
		root := t.TempDir()
		_, err := (&IPCHandler{}).HandleLint(api.LintRequest{
			Config: json.RawMessage(`[{
				"rules": { "no-console": ["error", { "allow": "warn" }] }
			}]`),
			ConfigDirectory:  root,
			WorkingDirectory: root,
		})
		if err == nil || !strings.Contains(err.Error(), `invalid options for rule "no-console"`) {
			t.Fatalf("first API request did not validate rule options: %v", err)
		}
		return
	}

	// Run this assertion in a fresh test process so first-request catalog setup
	// cannot accidentally depend on initialization performed by an earlier call.
	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleLintFirstRequestValidatesRuleOptions$")
	cmd.Env = append(os.Environ(), apiFirstRuleOptionsValidationProcess+"=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fresh-process API validation failed: %v\n%s", err, output)
	}
}

func TestHandleLintDiscoveredConfigValidatesRuleOptions(t *testing.T) {
	root := t.TempDir()
	writeProgramTestFiles(t, root, map[string]string{
		"rslint.config.js": "export default [];\n",
		"input.js":         "console.log('test');\n",
	})
	entries := mustAPIConfig(t, `[{"rules":{"no-console":["error",{"allow":"warn"}]}}]`)
	loadedIDs := make(map[string]struct{})
	requester := apiRequesterFunc(func(_ context.Context, kind ipc.MessageKind, payload any) (*ipc.Message, error) {
		switch kind {
		case api.KindLoadConfigs:
			request := payload.(discovery.ConfigLoadBatchRequest)
			response := discovery.ConfigLoadBatchResponse{TransactionID: request.TransactionID}
			for _, candidate := range request.Candidates {
				loadedIDs[candidate.ID] = struct{}{}
				response.Results = append(response.Results, discovery.ConfigLoadResult{
					ID:      candidate.ID,
					Status:  "loaded",
					Entries: entries,
				})
			}
			return ipc.NewMessage(ipc.KindResponse, 1, response)
		case api.KindActivateConfigs:
			request := payload.(discovery.ConfigActivationRequest)
			for _, id := range request.EffectiveConfigIDs {
				if _, ok := loadedIDs[id]; !ok {
					return nil, fmt.Errorf("activation contains unknown candidate ID %q", id)
				}
			}
			return ipc.NewMessage(ipc.KindResponse, 1, discovery.ConfigActivationResponse{
				TransactionID: request.TransactionID,
			})
		default:
			return nil, fmt.Errorf("unexpected reverse request kind %q", kind)
		}
	})

	_, err := (&IPCHandler{}).HandleLintWithContext(context.Background(), api.LintRequest{
		Files:            []string{filepath.Join(root, "input.js")},
		WorkingDirectory: root,
		ConfigDiscovery:  &api.ConfigDiscoveryRequest{},
	}, requester)
	if err == nil || !strings.Contains(err.Error(), `invalid options for rule "no-console"`) {
		t.Fatalf("discovered config did not validate rule options: %v", err)
	}
}

func TestHandleLintDiscoveryOverrideValidatesRuleOptionsWithoutCandidate(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "input.js")
	if err := os.WriteFile(target, []byte("console.log('test');\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	requester := apiRequesterFunc(func(context.Context, ipc.MessageKind, any) (*ipc.Message, error) {
		return nil, errors.New("empty discovery must not issue reverse requests")
	})

	_, err := (&IPCHandler{}).HandleLintWithContext(context.Background(), api.LintRequest{
		Files:            []string{target},
		WorkingDirectory: root,
		ConfigDiscovery: &api.ConfigDiscoveryRequest{
			OverrideConfig: json.RawMessage(`[{"rules":{"no-console":["error",{"allow":"warn"}]}}]`),
		},
	}, requester)
	if err == nil || !strings.Contains(err.Error(), `invalid options for rule "no-console"`) {
		t.Fatalf("discovery override did not validate rule options: %v", err)
	}
}
