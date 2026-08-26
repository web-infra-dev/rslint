package rules

import "testing"

func TestAllContainsEveryGoRuleExactlyOnce(t *testing.T) {
	sources := allRules()
	sharedCatalog := All()
	if sharedCatalog != All() {
		t.Fatal("All returned different catalog snapshots")
	}
	catalog := sharedCatalog.AllRules()
	if len(catalog) != len(sources) {
		t.Fatalf("rule catalog contains %d rules from %d source entries; a rule name is duplicated", len(catalog), len(sources))
	}
	for _, ruleImpl := range sources {
		if ruleImpl.IsEslintPluginRule {
			t.Errorf("Go rule %q is incorrectly marked as an object-form ESLint-plugin placeholder", ruleImpl.Name)
		}
		if _, ok := catalog[ruleImpl.Name]; !ok {
			t.Errorf("rule catalog is missing %q", ruleImpl.Name)
		}
	}
}
