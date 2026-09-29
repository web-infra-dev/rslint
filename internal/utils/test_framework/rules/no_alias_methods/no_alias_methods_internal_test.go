package no_alias_methods

import "testing"

func TestMatcherAliases(t *testing.T) {
	if len(MatcherAliases) != 11 {
		t.Errorf("expected 11 matcher aliases, got %d", len(MatcherAliases))
	}
	if got := MatcherAliases["toBeCalled"]; got != "toHaveBeenCalled" {
		t.Errorf("toBeCalled should map to toHaveBeenCalled, got %q", got)
	}
	for alias, canonical := range MatcherAliases {
		if _, isAlias := MatcherAliases[canonical]; isAlias {
			t.Errorf("canonical name %q (for alias %q) must not itself be an alias", canonical, alias)
		}
	}
}
