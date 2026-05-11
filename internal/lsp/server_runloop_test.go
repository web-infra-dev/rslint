package lsp

import (
	"errors"
	"fmt"
	"io"
	"testing"
)

// #4 regression: a clean shutdown (SIGINT/SIGTERM) must NOT surface an
// error, while a real (non-EOF) error with no shutdown signal MUST
// propagate. The original code judged the errgroup-derived ctx — which
// errgroup cancels before Wait() returns, so its Err() is ALWAYS non-nil —
// so the `ctx.Err() != nil` guard was always true and a clean shutdown's
// spurious context.Canceled got propagated as a fatal error (non-zero exit
// on graceful stop). runLoopError makes the decision off the SIGNAL ctx
// instead; `signalled` is that signal ctx's state.
func TestRunLoopError(t *testing.T) {
	realErr := errors.New("write failed")
	cases := []struct {
		name      string
		err       error
		signalled bool
		want      error
	}{
		{"real error, no signal -> propagate", realErr, false, realErr},
		{"real error, signalled -> swallow (shutdown fallout)", realErr, true, nil},
		{"EOF, no signal -> swallow (client disconnect)", io.EOF, false, nil},
		{"wrapped EOF, no signal -> swallow", fmt.Errorf("read: %w", io.EOF), false, nil},
		{"real error wrapped, no signal -> propagate", fmt.Errorf("x: %w", realErr), false, fmt.Errorf("x: %w", realErr)},
		{"nil error -> nil", nil, false, nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := runLoopError(tc.err, tc.signalled)
			if (got == nil) != (tc.want == nil) {
				t.Fatalf("runLoopError(%v, %v) = %v, want %v", tc.err, tc.signalled, got, tc.want)
			}
			if got != nil && got.Error() != tc.want.Error() {
				t.Errorf("runLoopError(%v, %v) = %q, want %q", tc.err, tc.signalled, got, tc.want)
			}
		})
	}
}
