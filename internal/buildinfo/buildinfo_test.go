package buildinfo

import "testing"

func TestInfoString(t *testing.T) {
	t.Parallel()
	release := "7.0.2"
	for _, tc := range []struct {
		name    string
		release *string
		want    string
	}{
		{name: "commit only", want: "rslint 0.9.2\nTypeScript (abc)"},
		{name: "exact release", release: &release, want: "rslint 0.9.2\nTypeScript 7.0.2 (abc)"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			info := Info{Version: "0.9.2", TypeScript: TypeScript{Commit: "abc", ReleaseVersion: tc.release}}
			if got := info.String(); got != tc.want {
				t.Fatalf("String() = %q, want %q", got, tc.want)
			}
		})
	}
}
