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
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/ipc"
)

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

	// Run this assertion in a fresh test process so no earlier API request can
	// have populated the process-global rule registry and hidden an ordering bug.
	cmd := exec.Command(os.Args[0], "-test.run=^TestHandleLintFirstRequestValidatesRuleOptions$")
	cmd.Env = append(os.Environ(), apiFirstRuleOptionsValidationProcess+"=1")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fresh-process API validation failed: %v\n%s", err, output)
	}
}

func TestHandleLintValidatesUnknownRuleNames(t *testing.T) {
	root := t.TempDir()
	_, err := (&IPCHandler{}).HandleLint(api.LintRequest{
		Config: json.RawMessage(`[{
			"rules": { "non-existent-rule-name": "error" }
		}]`),
		ConfigDirectory:  root,
		WorkingDirectory: root,
	})
	if err == nil || !strings.Contains(err.Error(), `unknown rule "non-existent-rule-name"`) {
		t.Fatalf("API request did not validate rule names: %v", err)
	}
}

func TestHandleLintResolvesMountedPluginRuleNames(t *testing.T) {
	// A rule mounted via the request's own plugin entries must pass name
	// validation even though it never has a native implementation — the
	// entries are the source of truth, not the process-global registry.
	root := t.TempDir()
	_, err := (&IPCHandler{}).HandleLint(api.LintRequest{
		Config: json.RawMessage(`[{
			"plugins": ["test-mounted"],
			"rules": { "test-mounted/no-foo": "error" }
		}]`),
		EslintPlugins: []api.EslintPluginEntry{
			{Prefix: "test-mounted", RuleNames: []string{"no-foo"}},
		},
		ConfigDirectory:  root,
		WorkingDirectory: root,
	})
	if err != nil {
		t.Fatalf("mounted plugin rule failed name validation: %v", err)
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
