package utils

import "testing"

// TestModuleSettingsKeySeparatesListElements locks in that the cache key keeps
// list boundaries: two ignore lists whose elements merely concatenate alike
// must not share one compiled-settings cache entry.
func TestModuleSettingsKeySeparatesListElements(t *testing.T) {
	t.Parallel()

	joined := moduleSettingsKey(map[string]interface{}{"import/ignore": []interface{}{"foo bar"}})
	split := moduleSettingsKey(map[string]interface{}{"import/ignore": []interface{}{"foo", "bar"}})
	if joined == split {
		t.Fatalf("moduleSettingsKey collided on %q", joined)
	}

	typed := moduleSettingsKey(map[string]interface{}{"import/ignore": []string{"foo", "bar"}})
	if typed != split {
		t.Fatalf("moduleSettingsKey separated equal settings: %q vs %q", typed, split)
	}
}
