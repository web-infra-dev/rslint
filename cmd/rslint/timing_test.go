package main

import (
	"testing"
)

func TestTimingFlagSet(t *testing.T) {
	cases := []struct {
		value       string
		wantErr     bool
		wantEnabled bool
		wantLimit   int
	}{
		{value: "true", wantEnabled: true},
		{value: "all", wantEnabled: true},
		{value: "ALL", wantEnabled: true},
		{value: "false"},
		{value: "10", wantEnabled: true, wantLimit: 10},
		{value: "0", wantErr: true},
		{value: "-3", wantErr: true},
		{value: "ten", wantErr: true},
	}
	for _, c := range cases {
		var enabled bool
		var limit int
		err := (&timingFlag{enabled: &enabled, limit: &limit}).Set(c.value)
		if (err != nil) != c.wantErr {
			t.Errorf("Set(%q) error = %v, wantErr %v", c.value, err, c.wantErr)
			continue
		}
		if err == nil && (enabled != c.wantEnabled || limit != c.wantLimit) {
			t.Errorf("Set(%q) = enabled %v limit %d, want enabled %v limit %d",
				c.value, enabled, limit, c.wantEnabled, c.wantLimit)
		}
	}
}
