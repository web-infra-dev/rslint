package rule

import (
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
	t.Run("WithRules", func(t *testing.T) {
		requireCatalogPanic(t, func() { catalog.WithRules() })
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
	if len(pluginNames) != 2 {
		t.Fatalf("plugin names = %v, want two rules", pluginNames)
	}
	pluginNames[0] = "changed"
	for _, ruleName := range catalog.RuleNamesForNamespace("plugin") {
		if ruleName == "changed" {
			t.Fatal("mutating returned rule names changed catalog")
		}
	}
}

func TestCatalogWithRulesDoesNotMutateBase(t *testing.T) {
	base := NewCatalog(Rule{Name: "base"})
	derived := base.WithRules(Rule{Name: "plugin", IsEslintPluginRule: true})

	if _, ok := base.Lookup("plugin"); ok {
		t.Fatal("derived rule leaked into base catalog")
	}
	if _, ok := derived.Lookup("base"); !ok {
		t.Fatal("derived catalog lost base rule")
	}
	if _, ok := derived.Lookup("plugin"); !ok {
		t.Fatal("derived catalog does not contain added rule")
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
