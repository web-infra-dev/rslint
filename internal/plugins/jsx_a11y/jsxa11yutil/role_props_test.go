package jsxa11yutil

import (
	"sort"
	"testing"
)

func TestRoleAriaPropsData(t *testing.T) {
	if len(ariaRolePropNames) > 64 {
		t.Fatalf("%d ARIA properties do not fit in a uint64 mask", len(ariaRolePropNames))
	}
	if !sort.StringsAreSorted(ariaRolePropNames[:]) || !sort.StringsAreSorted(ariaRoleNames[:]) {
		t.Fatal("generated name tables must be sorted for binary search")
	}
	if len(ariaRoleNames) != len(ariaRoleMasks) {
		t.Fatalf("%d role names but %d masks", len(ariaRoleNames), len(ariaRoleMasks))
	}
	memberships := 0
	for _, role := range ariaRoleNames {
		props, ok := LookupRoleAriaProps(role)
		if !ok {
			t.Fatalf("known role %q not found", role)
		}
		memberships += len(props.Names())
	}
	if len(ariaRoleNames) != 139 || memberships != 2833 {
		t.Errorf("got %d roles and %d memberships, want 139 and 2833", len(ariaRoleNames), memberships)
	}
}

func TestLookupRoleAriaProps(t *testing.T) {
	// A known role with no supported properties differs from an unknown role.
	for _, role := range []string{"none", "doc-pullquote"} {
		props, ok := LookupRoleAriaProps(role)
		if !ok || props != 0 {
			t.Errorf("LookupRoleAriaProps(%q) = %d, %v; want 0, true", role, props, ok)
		}
	}
	for _, role := range []string{"", "BUTTON", "Button", "unknown-role"} {
		if _, ok := LookupRoleAriaProps(role); ok {
			t.Errorf("LookupRoleAriaProps(%q) unexpectedly found", role)
		}
	}

	button, ok := LookupRoleAriaProps("button")
	if !ok {
		t.Fatal("button not found")
	}
	if !button.Supports("aria-pressed") {
		t.Error("button should support aria-pressed")
	}
	if button.Supports("aria-checked") || button.Supports("ARIA-PRESSED") || button.Supports("aria-unknown") {
		t.Error("button supports only its own case-sensitive ARIA properties")
	}

	// Abstract, DPUB and Graphics roles are included.
	for _, role := range []string{"command", "widget", "doc-abstract", "graphics-document"} {
		if _, ok := LookupRoleAriaProps(role); !ok {
			t.Errorf("role %q missing", role)
		}
	}
}
