package term

import "testing"

// TestResolveColorEnabled pins every precedence tier and the tier-vs-tier
// orderings that differ from the pre-refactor behavior. Inputs are fully
// explicit per case — no t.Setenv, no process env reads, no mutation of
// fatih/color globals (writing color.NoColor from a test poisons other tests
// in the same binary).
func TestResolveColorEnabled(t *testing.T) {
	t.Parallel()

	// fc builds the FORCE_COLOR (value, set) pair.
	fc := func(v string) ColorInputs {
		return ColorInputs{ForceColorEnv: v, ForceColorEnvSet: true}
	}

	cases := []struct {
		name string
		in   ColorInputs
		want bool
	}{
		// tier 7: the TTY fact baseline
		{"all absent, not a TTY", ColorInputs{}, false},
		{"all absent, TTY", ColorInputs{StdoutIsTTY: true}, true},
		{"exotic TERM on a TTY still colors (documented tier-7 deviation)",
			ColorInputs{Term: "xterm-nonsense", StdoutIsTTY: true}, true},

		// tier 1: --no-color beats everything
		{"--no-color beats --force-color and TTY",
			ColorInputs{NoColorFlag: true, ForceColorFlag: true, StdoutIsTTY: true}, false},
		{"--no-color beats FORCE_COLOR=1",
			ColorInputs{NoColorFlag: true, ForceColorEnv: "1", ForceColorEnvSet: true}, false},
		{"--no-color beats GITHUB_ACTIONS",
			ColorInputs{NoColorFlag: true, GithubActionsEnv: "true"}, false},

		// tier 2: --force-color beats env and the TTY fact
		{"--force-color on a pipe", ColorInputs{ForceColorFlag: true}, true},
		{"--force-color beats NO_COLOR env",
			ColorInputs{ForceColorFlag: true, NoColorEnvSet: true}, true},
		{"--force-color beats FORCE_COLOR=0",
			ColorInputs{ForceColorFlag: true, ForceColorEnv: "0", ForceColorEnvSet: true}, true},
		{"--force-color beats TERM=dumb",
			ColorInputs{ForceColorFlag: true, Term: "dumb"}, true},

		// tier 3: FORCE_COLOR is value-aware and terminal
		{"FORCE_COLOR set-but-empty enables", fc(""), true},
		{"FORCE_COLOR=1 enables", fc("1"), true},
		{"FORCE_COLOR=true enables", fc("true"), true},
		{"FORCE_COLOR=2 enables", fc("2"), true},
		{"FORCE_COLOR=3 enables", fc("3"), true},
		{"FORCE_COLOR=0 disables even on a TTY",
			ColorInputs{ForceColorEnv: "0", ForceColorEnvSet: true, StdoutIsTTY: true}, false},
		{"FORCE_COLOR=false disables", fc("false"), false},
		{"FORCE_COLOR unrecognized value disables (Node semantics)", fc("yes"), false},
		{"FORCE_COLOR=1 beats NO_COLOR (Node: NO_COLOR ignored when FORCE_COLOR set)",
			ColorInputs{ForceColorEnv: "1", ForceColorEnvSet: true, NoColorEnvSet: true}, true},
		{"FORCE_COLOR=0 with NO_COLOR also set disables",
			ColorInputs{ForceColorEnv: "0", ForceColorEnvSet: true, NoColorEnvSet: true}, false},
		{"FORCE_COLOR=1 beats TERM=dumb",
			ColorInputs{ForceColorEnv: "1", ForceColorEnvSet: true, Term: "dumb"}, true},
		{"FORCE_COLOR=0 beats GITHUB_ACTIONS",
			ColorInputs{ForceColorEnv: "0", ForceColorEnvSet: true, GithubActionsEnv: "true"}, false},

		// tier 4: NO_COLOR (presence, including empty value)
		{"NO_COLOR set on a TTY disables", ColorInputs{NoColorEnvSet: true, StdoutIsTTY: true}, false},
		{"NO_COLOR set-but-empty disables (Node semantics)", ColorInputs{NoColorEnvSet: true}, false},
		{"NO_COLOR beats GITHUB_ACTIONS",
			ColorInputs{NoColorEnvSet: true, GithubActionsEnv: "true"}, false},

		// tier 5: TERM=dumb beats GITHUB_ACTIONS and the TTY fact
		{"TERM=dumb on a TTY disables", ColorInputs{Term: "dumb", StdoutIsTTY: true}, false},
		{"TERM=dumb beats GITHUB_ACTIONS",
			ColorInputs{Term: "dumb", GithubActionsEnv: "true", StdoutIsTTY: true}, false},

		// tier 6: GITHUB_ACTIONS forces on for piped CI logs (non-empty check:
		// any value counts, set-but-empty falls through to the TTY fact)
		{"GITHUB_ACTIONS on a pipe enables", ColorInputs{GithubActionsEnv: "true"}, true},
		{"GITHUB_ACTIONS=false still enables (non-empty semantics)",
			ColorInputs{GithubActionsEnv: "false"}, true},
		{"GITHUB_ACTIONS set-but-empty falls through to the TTY fact",
			ColorInputs{GithubActionsEnv: ""}, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := ResolveColorEnabled(tc.in); got != tc.want {
				t.Errorf("ResolveColorEnabled(%+v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}
