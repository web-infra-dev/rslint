package minimatch

import "testing"

// TestCompiledPatternsAreBounded covers the bound on how long one name may
// spend in one pattern part. The extended glob syntax compiles to overlapping
// alternations, which regexp2 walks by backtracking and so can explore
// exponentially many ways over a name that ends up not matching.
//
// This reads the bound off the compiled patterns rather than handing the
// matcher a name that needs it: running the pathological case to watch it stop
// costs a core for as long as the bound allows, every time the suite runs.
func TestCompiledPatternsAreBounded(t *testing.T) {
	if matchTimeout <= 0 {
		t.Fatalf("matchTimeout = %s, want a finite bound", matchTimeout)
	}
	if braceShortcut.MatchTimeout != matchTimeout {
		t.Errorf("braceShortcut.MatchTimeout = %s, want %s", braceShortcut.MatchTimeout, matchTimeout)
	}

	// Between them these cover a part compiled through every route that
	// produces a regexp: an extended glob list, a negated list, a character
	// class, a plain wildcard, and a list left open at the end of the part.
	patterns := []string{
		"+(a|aa)",
		"*.!(x).!(y|z)",
		"/src/!(a!(a|b)",
		"[a-z]*",
		"?(a|b)/**/*.ts",
		"{a,b}/*",
	}

	for _, pattern := range patterns {
		t.Run(pattern, func(t *testing.T) {
			m := New(pattern, Options{})
			compiled := 0
			for _, row := range m.set {
				for _, part := range row {
					if part.re == nil {
						continue
					}
					compiled++
					if part.re.MatchTimeout != matchTimeout {
						t.Errorf("part %q of %q has MatchTimeout %s, want %s",
							part.re.String(), pattern, part.re.MatchTimeout, matchTimeout)
					}
				}
			}
			if compiled == 0 {
				t.Errorf("%q compiled to no regexp at all, so the bound covers nothing", pattern)
			}
		})
	}
}
