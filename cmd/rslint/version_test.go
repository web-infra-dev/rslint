package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"testing"

	"github.com/web-infra-dev/rslint/internal/buildinfo"
)

func TestRunVersion(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		args []string
		code int
	}{
		{name: "text"},
		{name: "json", args: []string{"--json"}},
		{name: "unknown flag", args: []string{"--wat"}, code: 2},
		{name: "positional", args: []string{"file.ts"}, code: 2},
		{name: "invalid boolean", args: []string{"--json=wat"}, code: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var stdout, stderr bytes.Buffer
			if code := runVersion(tc.args, &stdout, &stderr); code != tc.code {
				t.Fatalf("exit = %d, want %d; stderr: %s", code, tc.code, &stderr)
			}
			if tc.code != 0 {
				if stdout.Len() != 0 || stderr.Len() == 0 {
					t.Fatalf("invalid arguments must only write stderr: stdout=%q stderr=%q", &stdout, &stderr)
				}
				return
			}
			if stderr.Len() != 0 {
				t.Fatalf("unexpected stderr: %s", &stderr)
			}
			if tc.name == "json" {
				var info buildinfo.Info
				decoder := json.NewDecoder(&stdout)
				if err := decoder.Decode(&info); err != nil {
					t.Fatal(err)
				}
				if info.String() != buildinfo.Current().String() {
					t.Fatalf("wrong embedded binding: %+v", info)
				}
				if err := decoder.Decode(&info); !errors.Is(err, io.EOF) {
					t.Fatalf("extra data after version JSON: %v", err)
				}
			} else if stdout.String() != buildinfo.Current().String()+"\n" {
				t.Fatalf("unexpected version output: %q", &stdout)
			}
		})
	}
}
