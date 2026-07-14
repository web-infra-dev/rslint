package config

import (
	"testing"
)

// TestAllRules_DeclaredSchemasCompile compiles every registered rule's
// declared options schema. Schemas compile lazily at first use (so a bad
// schema would otherwise only surface when its rule is enabled by some
// user's config); this sweep front-loads that failure into CI, playing the
// role a MustCompile-at-init would — without making every rslint process pay
// startup compilation for hundreds of schemas.
func TestAllRules_DeclaredSchemasCompile(t *testing.T) {
	RegisterAllRules()
	for name, ruleImpl := range GlobalRuleRegistry.GetAllRules() {
		if ruleImpl.Schema == nil {
			continue
		}
		if _, err := ruleImpl.Schema.Compile(); err != nil {
			t.Errorf("rule %s: options schema failed to compile: %v", name, err)
		}
	}
}
