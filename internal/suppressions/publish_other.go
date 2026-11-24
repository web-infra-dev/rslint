//go:build unix && !aix && !darwin && !linux

package suppressions

func renameNew(from, to string) error { return linkNew(from, to) }
