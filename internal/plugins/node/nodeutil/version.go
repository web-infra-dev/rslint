package nodeutil

import (
	"github.com/web-infra-dev/rslint/internal/rule"
	"github.com/web-infra-dev/rslint/internal/utils/npmsemver"
	"github.com/web-infra-dev/rslint/internal/utils/packagejson"
)

// ConfiguredNodeVersion follows options, settings.n, settings.node, package
// engines.node, devEngines.runtime, then the upstream >=16.0.0 fallback.
func ConfiguredNodeVersion(ctx rule.RuleContext, options map[string]any) npmsemver.Range {
	for _, raw := range settingValues("version", options, ctx.Settings) {
		if raw == nil || raw == "" || raw == false || raw == float64(0) {
			continue
		}
		if version, ok := npmsemver.Parse(settingString(raw)); ok {
			return version
		}
	}
	if p := ctx.Program(); p != nil {
		if pkg := packagejson.FindNearestValid(p, ctx.SourceFile.FileName()); pkg != nil {
			engines, _ := pkg.Field("engines").(map[string]any)
			if text, ok := engines["node"].(string); ok {
				if version, valid := npmsemver.Parse(text); valid {
					return version
				}
			}
			devEngines, _ := pkg.Field("devEngines").(map[string]any)
			entries, ok := devEngines["runtime"].([]any)
			if !ok {
				entries = []any{devEngines["runtime"]}
			}
			for _, raw := range entries {
				entry, _ := raw.(map[string]any)
				if entry["name"] != "node" {
					continue
				}
				if text, ok := entry["version"].(string); ok {
					if version, valid := npmsemver.Parse(text); valid {
						return version
					}
					break
				}
			}
		}
	}
	version, _ := npmsemver.Parse(">=16.0.0")
	return version
}
