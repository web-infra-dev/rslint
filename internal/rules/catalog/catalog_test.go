package catalog

import (
	"strings"
	"sync"
	"testing"

	"github.com/web-infra-dev/rslint/internal/rule"
)

func TestNativeRuleCollectionsHaveUniqueNamesAndExpectedNamespaces(t *testing.T) {
	seen := make(map[string]string)
	for _, collection := range nativeRuleCollections {
		rules := collection.allRules()
		if len(rules) == 0 {
			t.Errorf("native rule collection %q is empty", collection.namespace)
		}
		for _, ruleImpl := range rules {
			if previousNamespace, exists := seen[ruleImpl.Name]; exists {
				t.Errorf("native rule %q occurs in both %q and %q", ruleImpl.Name, previousNamespace, collection.namespace)
				continue
			}
			seen[ruleImpl.Name] = collection.namespace
			if collection.namespace == "" {
				if strings.Contains(ruleImpl.Name, "/") {
					t.Errorf("core rule %q contains a plugin namespace", ruleImpl.Name)
				}
				continue
			}
			if !strings.HasPrefix(ruleImpl.Name, collection.namespace+"/") {
				t.Errorf("rule %q does not belong to plugin namespace %q", ruleImpl.Name, collection.namespace)
			}
		}
	}
	if got := len(Native().AllRules()); got != len(seen) {
		t.Fatalf("native catalog contains %d rules, want %d unique source rules", got, len(seen))
	}
}

func TestWithESLintPluginsRequiresBaseCatalog(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected nil base catalog to panic")
		}
	}()
	WithESLintPlugins(nil, nil)
}

func TestWithESLintPluginsIsScopedToDerivedCatalog(t *testing.T) {
	base := rule.NewCatalog(rule.Rule{Name: "native"})
	first, shadowed := WithESLintPlugins(base, []rule.ESLintPluginMetadata{{
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

	second, _ := WithESLintPlugins(base, nil)
	if _, ok := second.Lookup("community/first"); ok {
		t.Fatal("plugin rule leaked into later catalog")
	}
}

func TestWithESLintPluginsKeepsNativeRule(t *testing.T) {
	native := rule.Rule{Name: "community/check", RequiresTypeInfo: true}
	derived, shadowed := WithESLintPlugins(rule.NewCatalog(native), []rule.ESLintPluginMetadata{{
		Prefix:    "community",
		RuleNames: []string{"check"},
	}})
	if len(shadowed) != 1 || shadowed[0] != native.Name {
		t.Fatalf("shadowed = %v, want [%q]", shadowed, native.Name)
	}
	got, ok := derived.Lookup(native.Name)
	if !ok || got.IsEslintPluginRule || !got.RequiresTypeInfo {
		t.Fatalf("Lookup(%q) = (%+v, %t), want native rule", native.Name, got, ok)
	}
}

func TestNativeSupportsConcurrentReads(t *testing.T) {
	catalog := Native()
	const readers = 32
	var wait sync.WaitGroup
	wait.Add(readers)
	for range readers {
		go func() {
			defer wait.Done()
			if _, ok := catalog.Lookup("no-debugger"); !ok {
				t.Errorf("native catalog is missing no-debugger")
			}
		}()
	}
	wait.Wait()
}
