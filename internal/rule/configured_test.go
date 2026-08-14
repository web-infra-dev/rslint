package rule

import "testing"

func TestFilterNonTypeAwareRules(t *testing.T) {
	tests := []struct {
		name  string
		rules []ConfiguredRule
		want  []string
	}{
		{
			name: "mixed",
			rules: []ConfiguredRule{
				{Name: "syntax-rule"},
				{Name: "type-rule", RequiresTypeInfo: true},
				{Name: "another-syntax"},
			},
			want: []string{"syntax-rule", "another-syntax"},
		},
		{
			name: "all type aware",
			rules: []ConfiguredRule{
				{Name: "type-rule-1", RequiresTypeInfo: true},
				{Name: "type-rule-2", RequiresTypeInfo: true},
			},
		},
		{
			name:  "all syntax",
			rules: []ConfiguredRule{{Name: "rule-a"}, {Name: "rule-b"}},
			want:  []string{"rule-a", "rule-b"},
		},
		{name: "empty"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filtered := FilterNonTypeAwareRules(test.rules)
			if len(filtered) != len(test.want) {
				t.Fatalf("got %d rules, want %d", len(filtered), len(test.want))
			}
			for index, name := range test.want {
				if filtered[index].Name != name {
					t.Fatalf("rule %d = %q, want %q", index, filtered[index].Name, name)
				}
			}
		})
	}
}

func TestCreateRulePreservesRequiresTypeInfo(t *testing.T) {
	configured := CreateRule(Rule{
		Name:             "test-rule",
		RequiresTypeInfo: true,
		Run:              func(RuleContext, []any) RuleListeners { return nil },
	})

	if configured.Name != "@typescript-eslint/test-rule" {
		t.Fatalf("unexpected name: %s", configured.Name)
	}
	if !configured.RequiresTypeInfo {
		t.Fatal("RequiresTypeInfo should survive CreateRule")
	}
}
