//go:build unix && !aix && !darwin && !linux

package suppressions

import "os"

func copyAccessACL(_, _ *os.File) error { return nil }

func renameNew(from, to string) error { return linkNew(from, to) }
