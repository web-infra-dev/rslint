package main

import (
	"fmt"
	"strconv"
	"strings"
)

// timingFlag parses --timing. IsBoolFlag makes the value optional: a bare
// --timing enables the full table, --timing=N keeps only the top N rules
// (the value must use the = form, or N would swallow a positional path).
type timingFlag struct {
	enabled *bool
	limit   *int
}

func (f *timingFlag) IsBoolFlag() bool { return true }

func (f *timingFlag) String() string { return "" }

func (f *timingFlag) Set(value string) error {
	switch strings.ToLower(value) {
	case "true", "all":
		*f.enabled = true
		*f.limit = 0
		return nil
	case "false":
		*f.enabled = false
		*f.limit = 0
		return nil
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fmt.Errorf("expected \"all\" or a positive rule count, got %q", value)
	}
	*f.enabled = true
	*f.limit = n
	return nil
}
