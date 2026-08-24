package rule

import (
	"slices"
	"sync"
	"testing"
)

func requireCatalogPanic(t *testing.T, run func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic")
		}
	}()
	run()
}

func TestCatalogRejectsNilReceiver(t *testing.T) {
	var catalog *Catalog
	t.Run("ForESLintPlugins", func(t *testing.T) {
		requireCatalogPanic(t, func() { catalog.ForESLintPlugins(nil) })
	})
	t.Run("Lookup", func(t *testing.T) {
		requireCatalogPanic(t, func() { catalog.Lookup("rule") })
	})
	t.Run("AllRules", func(t *testing.T) {
		requireCatalogPanic(t, func() { catalog.AllRules() })
	})
	t.Run("RuleNamesForNamespace", func(t *testing.T) {
		requireCatalogPanic(t, func() { catalog.RuleNamesForNamespace("") })
	})
}

func TestNamespace(t *testing.T) {
	tests := []struct {
		ruleName string
		expected string
	}{
		{ruleName: "@typescript-eslint/no-explicit-any", expected: "@typescript-eslint"},
		{ruleName: "@scope/plugin/rule", expected: "@scope/plugin"},
		{ruleName: "import/no-unresolved", expected: "import"},
		{ruleName: "plugin/", expected: "plugin"},
		{ruleName: "no-debugger", expected: ""},
	}
	for _, test := range tests {
		if actual := Namespace(test.ruleName); actual != test.expected {
			t.Errorf("Namespace(%q) = %q, want %q", test.ruleName, actual, test.expected)
		}
	}
}

func TestCatalogRuleNamesForNamespaceReturnsCopy(t *testing.T) {
	catalog := NewCatalog(
		Rule{Name: "core"},
		Rule{Name: "plugin/first"},
		Rule{Name: "plugin/second"},
	)
	coreNames := catalog.RuleNamesForNamespace("")
	if len(coreNames) != 1 || coreNames[0] != "core" {
		t.Fatalf("core names = %v, want [core]", coreNames)
	}
	pluginNames := catalog.RuleNamesForNamespace("plugin")
	if len(pluginNames) != 2 || pluginNames[0] != "plugin/first" || pluginNames[1] != "plugin/second" {
		t.Fatalf("plugin names = %v, want sorted names", pluginNames)
	}
	pluginNames[0] = "changed"
	for _, ruleName := range catalog.RuleNamesForNamespace("plugin") {
		if ruleName == "changed" {
			t.Fatal("mutating returned rule names changed catalog")
		}
	}
}

func TestCatalogAllRulesReturnsCopy(t *testing.T) {
	catalog := NewCatalog(Rule{Name: "base"})
	rules := catalog.AllRules()
	delete(rules, "base")
	rules["injected"] = Rule{Name: "injected"}

	if _, ok := catalog.Lookup("base"); !ok {
		t.Fatal("deleting from All result mutated catalog")
	}
	if _, ok := catalog.Lookup("injected"); ok {
		t.Fatal("adding to All result mutated catalog")
	}
}

func TestCatalogDuplicateNameUsesLastRule(t *testing.T) {
	catalog := NewCatalog(
		Rule{Name: "duplicate", RequiresTypeInfo: false},
		Rule{Name: "duplicate", RequiresTypeInfo: true},
	)
	ruleImpl, ok := catalog.Lookup("duplicate")
	if !ok || !ruleImpl.RequiresTypeInfo {
		t.Fatalf("Lookup duplicate = (%+v, %t), want last rule", ruleImpl, ok)
	}
}

func TestCatalogForESLintPluginsReplacesTheExactPluginSet(t *testing.T) {
	base := NewCatalog(Rule{Name: "go-rule"})
	first, shadowed := base.ForESLintPlugins([]ESLintPluginMetadata{{
		Prefix:    "community",
		RuleNames: []string{"first"},
	}})
	if len(shadowed) != 0 {
		t.Fatalf("shadowed = %v, want none", shadowed)
	}
	if _, ok := first.Lookup("community/first"); !ok {
		t.Fatal("derived catalog is missing plugin rule")
	}
	if _, ok := base.Lookup("community/first"); ok {
		t.Fatal("plugin rule leaked into base catalog")
	}

	second, _ := first.ForESLintPlugins([]ESLintPluginMetadata{{
		Prefix:    "replacement",
		RuleNames: []string{"second"},
	}})
	if _, ok := second.Lookup("community/first"); ok {
		t.Fatal("later catalog retained a plugin rule outside its exact plugin set")
	}
	if _, ok := second.Lookup("replacement/second"); !ok {
		t.Fatal("later catalog is missing its replacement plugin rule")
	}
	withoutPlugins, _ := second.ForESLintPlugins(nil)
	if _, ok := withoutPlugins.Lookup("replacement/second"); ok {
		t.Fatal("empty plugin set retained a plugin rule")
	}
	if _, ok := withoutPlugins.Lookup("go-rule"); !ok {
		t.Fatal("replacing plugin rules removed a Go rule")
	}
}

func TestCatalogForESLintPluginsKeepsGoRule(t *testing.T) {
	builtIn := Rule{Name: "community/check", RequiresTypeInfo: true}
	derived, shadowed := NewCatalog(builtIn).ForESLintPlugins([]ESLintPluginMetadata{{
		Prefix:    "community",
		RuleNames: []string{"check"},
	}})
	if len(shadowed) != 1 || shadowed[0] != builtIn.Name {
		t.Fatalf("shadowed = %v, want [%q]", shadowed, builtIn.Name)
	}
	got, ok := derived.Lookup(builtIn.Name)
	if !ok || got.IsEslintPluginRule || !got.RequiresTypeInfo {
		t.Fatalf("Lookup(%q) = (%+v, %t), want Go rule", builtIn.Name, got, ok)
	}
}

func TestCatalogForESLintPluginsKeepsMetadataOrderForGoRuleCollisions(t *testing.T) {
	base := NewCatalog(
		Rule{Name: "second/check"},
		Rule{Name: "first/check"},
	)
	_, shadowed := base.ForESLintPlugins([]ESLintPluginMetadata{
		{Prefix: "first", RuleNames: []string{"check"}},
		{Prefix: "second", RuleNames: []string{"check"}},
	})
	want := []string{"first/check", "second/check"}
	if !slices.Equal(shadowed, want) {
		t.Fatalf("shadowed = %v, want %v", shadowed, want)
	}
}

func TestCatalogForESLintPluginsDoesNotReportDuplicatePluginMetadataAsShadowed(t *testing.T) {
	derived, shadowed := NewCatalog().ForESLintPlugins([]ESLintPluginMetadata{
		{Prefix: "community", RuleNames: []string{"check", "check"}},
		{Prefix: "community", RuleNames: []string{"check"}},
	})
	if len(shadowed) != 0 {
		t.Fatalf("shadowed = %v, want none", shadowed)
	}
	ruleImpl, ok := derived.Lookup("community/check")
	if !ok || !ruleImpl.IsEslintPluginRule {
		t.Fatalf("Lookup(community/check) = (%+v, %t), want plugin placeholder", ruleImpl, ok)
	}
}

func TestCatalogForESLintPluginsIgnoresEmptyPrefixButKeepsEmptyRuleName(t *testing.T) {
	derived, _ := NewCatalog().ForESLintPlugins([]ESLintPluginMetadata{
		{Prefix: "", RuleNames: []string{"ignored"}},
		{Prefix: "community", RuleNames: []string{"", "check"}},
	})
	if _, ok := derived.Lookup("/ignored"); ok {
		t.Fatal("empty plugin prefix became resolvable")
	}
	for _, ruleName := range []string{"community/", "community/check"} {
		ruleImpl, ok := derived.Lookup(ruleName)
		if !ok || !ruleImpl.IsEslintPluginRule {
			t.Errorf("Lookup(%q) = (%+v, %t), want plugin placeholder", ruleName, ruleImpl, ok)
		}
	}
}

func TestCatalogSupportsConcurrentReads(t *testing.T) {
	catalog := NewCatalog(Rule{Name: "base"})
	const readers = 32
	var wait sync.WaitGroup
	wait.Add(readers)
	for range readers {
		go func() {
			defer wait.Done()
			if _, ok := catalog.Lookup("base"); !ok {
				t.Errorf("catalog is missing base rule")
			}
		}()
	}
	wait.Wait()
}
