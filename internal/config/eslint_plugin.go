package config

import (
	"fmt"
	"os"

	"github.com/web-infra-dev/rslint/internal/rule"
)

// EslintPluginEntry is the wire-format descriptor for a single ESLint
// plugin the user declared via `eslintPlugins: { <prefix>: <pluginObj> }`.
// The Node side (`packages/rslint/src/config-loader.ts`) extracts it from
// the live plugin instance; the Go side never sees the live object, only
// this serializable form, forwarded to the Worker via the IPC `init`
// payload.
//
// The Go runtime does NOT execute JS plugin code and does NOT introspect
// any field beyond Prefix + RuleNames — those are the only two the
// placeholder-rule registry consumes:
//
//   - Prefix:    user-chosen namespace (e.g. "uc"); rules are referenced
//     in the `rules` block as "<prefix>/<ruleName>".
//   - RuleNames: keys of `plugin.rules`; the registry creates placeholder
//     entries "<prefix>/<ruleName>" without ever loading the plugin.
//
// The remaining fields (Specifier, Version, Options, ResolvedPath) are
// pass-through metadata only: decoded off the wire and kept so a future
// change can use them without a wire-format break, but Go ignores them
// today (it never re-sends them; the WorkerPool that could use them is
// owned Node-side).
type EslintPluginEntry struct {
	Prefix    string                 `json:"prefix"`
	Specifier string                 `json:"specifier,omitempty"`
	Version   string                 `json:"version,omitempty"`
	RuleNames []string               `json:"ruleNames"`
	Options   map[string]interface{} `json:"options,omitempty"`
	// ResolvedPath: absolute filesystem path of the plugin module as
	// computed by the Node side at config load. Pass-through only.
	ResolvedPath string `json:"resolvedPath,omitempty"`
}

// FullRuleNames returns the rule identifiers as they appear in user `rules`
// configuration: "<prefix>/<ruleName>" for each name in RuleNames.
func (e EslintPluginEntry) FullRuleNames() []string {
	out := make([]string, 0, len(e.RuleNames))
	for _, name := range e.RuleNames {
		out = append(out, e.Prefix+"/"+name)
	}
	return out
}

// RegisterEslintPluginRules populates the shared GlobalRuleRegistry with
// placeholder entries for every rule in every supplied plugin entry. It is
// called once per run, AFTER RegisterAllRules() has populated native rules,
// so native rules win every same-name conflict deterministically.
//
// Behavior:
//   - For each <prefix>/<ruleName> not already in the registry: register a
//     placeholder rule with IsEslintPluginRule=true and a no-op Run. The real
//     execution is dispatched by the linter at Program-batch level via IPC
//     (M3); placeholder rules merely make them visible to the rule resolver
//     so user `rules: { 'uc/no-null': 'error' }` configurations don't get
//     reported as "unknown rule".
//   - Same-name conflicts (the registry already has a native rule with the
//     same prefix+name) are skipped silently in the registry — but a single
//     stderr line per skipped rule is emitted so users can observe that
//     their JS plugin's same-named rule won't run. Native always wins.
//
// Idempotency: calling this twice with the same entries is safe — the second
// call is a no-op for every entry whose rule names are already registered
// (whether native or placeholder).
//
// Side effects: writes to GlobalRuleRegistry; emits stderr lines on conflict.
// Determinism: rules within a plugin are processed in input slice order;
// plugins are processed in input slice order. The conflict warning lines
// follow the same order — important for golden tests.
func RegisterEslintPluginRules(entries []EslintPluginEntry) {
	for _, entry := range entries {
		if entry.Prefix == "" {
			fmt.Fprintf(os.Stderr,
				"[rslint] skipping eslintPlugin entry with empty prefix (rules=%v)\n",
				entry.RuleNames)
			continue
		}

		for _, ruleName := range entry.RuleNames {
			fullName := entry.Prefix + "/" + ruleName

			if existing, exists := GlobalRuleRegistry.GetRule(fullName); exists {
				if existing.IsEslintPluginRule {
					// Same name already registered as a plugin placeholder
					// (different entry contributed it earlier this run, or
					// the same entry was registered twice). Silent dedupe —
					// per-file dispatch picks the right instance via
					// CompatLintFile.ConfigKey, so the registry just needs
					// the name to exist.
					continue
				}
				// A native rule with the same fully-qualified name takes
				// precedence; warn the user once so the shadowing is
				// visible.
				fmt.Fprintf(os.Stderr,
					"[rslint] skipped JS rule %s because a native implementation exists; native takes precedence\n",
					fullName)
				continue
			}

			GlobalRuleRegistry.Register(fullName, rule.Rule{
				Name:               fullName,
				RequiresTypeInfo:   false, // ESLint plugins in the Worker are not type-aware
				IsEslintPluginRule: true,
				Run: func(ctx rule.RuleContext, options any) rule.RuleListeners {
					// Placeholder. The linter MUST NOT invoke this; it should
					// route the rule into a per-Program lintEslintPlugin
					// batch instead. Returning nil keeps the listener
					// dispatch table empty if Run is mistakenly called.
					return nil
				},
			})
		}
	}
}

// (The Go runtime no longer owns plugin identity — it cannot execute
// JS plugin code, so it cannot reason about which two entries represent
// the same plugin instance. Cross-config rule-name coalescing happens
// on the Node side in `packages/rslint/src/cli.ts::runViaEngine` before
// the IPC `init` payload is sent. Per-file dispatch from Go to the
// runner is keyed by config directory (CompatLintFile.ConfigKey); the
// runner worker imports the user's rslint config files directly and
// uses that key to pick the right loaded plugin set.)
