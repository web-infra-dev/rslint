package main

import (
	"path/filepath"
	"testing"
)

func TestGeneratedProtocolIsCurrent(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	if err := run(root, true); err != nil {
		t.Fatal(err)
	}
}
