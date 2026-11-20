// cspell:ignore APFS Fstypename

package main

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"golang.org/x/sys/unix"
	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

const pcCaseSensitive = 11 // Darwin's _PC_CASE_SENSITIVE selector.

func darwinSameParent(left, right string) bool {
	dir, other := filepath.Dir(left), filepath.Dir(right)
	if dir == other {
		return true
	}
	info, _ := os.Stat(dir)
	otherInfo, _ := os.Stat(other)
	return os.SameFile(info, otherInfo)
}

// A search-only parent can prevent reading a hard link's stored spelling.
// Verify both its name on this volume and the object reached without following it.
func darwinSameCaseLookup(left, right string) bool {
	name, other := filepath.Base(left), filepath.Base(right)
	if name != other && !darwinEqualLookupNames(filepath.Dir(left), name, other) {
		return false
	}
	info, _ := os.Lstat(left)
	otherInfo, _ := os.Lstat(right)
	return os.SameFile(info, otherInfo)
}

func darwinEqualLookupNames(parent, left, right string) bool {
	sensitive, err := unix.Pathconf(parent, pcCaseSensitive)
	if err != nil {
		return false
	}
	if !strings.ContainsFunc(left+right, func(r rune) bool { return r > unicode.MaxASCII }) {
		return sensitive == 0 && strings.EqualFold(left, right)
	}
	var volume unix.Statfs_t
	if unix.Statfs(parent, &volume) != nil || unix.ByteSliceToString(volume.Fstypename[:]) != "apfs" {
		return false
	}
	// Both APFS variants accept canonical equivalents. Restrict this proof to
	// fully composed names: APFS compares some remaining mark sequences
	// differently from standard normalization and case folding.
	left, right = norm.NFC.String(left), norm.NFC.String(right)
	if strings.ContainsFunc(left+right, unicode.IsMark) {
		return false
	}
	if sensitive != 0 || left == right {
		return left == right
	}
	fold := cases.Fold()
	return strings.EqualFold(fold.String(left), fold.String(right))
}

// The kernel can report the name of another hard link. Only consider directory
// entries in the requested parent that identify the same object. An empty name
// means that the underlying lookup could not be attributed to one entry.
func darwinHardLinkEntryName(parent, requested, reportedPath string, original os.FileInfo) (string, error) {
	path := filepath.Join(parent, requested)
	if path == reportedPath {
		return requested, nil
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		// A reported link in another parent says nothing about this entry's
		// spelling, even when both names open the same inode.
		if reportedPath != "" && darwinSameParent(path, reportedPath) && darwinSameCaseLookup(path, reportedPath) {
			return filepath.Base(reportedPath), nil
		}
		return "", err
	}
	for _, entry := range entries {
		if entry.Name() == requested {
			return requested, nil
		}
	}
	var names []string
	for _, entry := range entries {
		if entry.Type() != original.Mode().Type() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return "", err
		}
		if os.SameFile(original, info) {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 1 {
		return names[0], nil
	}
	spelling := norm.NFD.String
	if sensitive, err := unix.Pathconf(parent, pcCaseSensitive); err == nil && sensitive == 0 {
		fold := cases.Fold()
		spelling = func(s string) string { return fold.String(norm.NFD.String(s)) }
	}
	want := spelling(requested)
	match := ""
	for _, name := range names {
		if spelling(name) == want {
			if match != "" {
				return "", nil
			}
			match = name
		}
	}
	return match, nil
}
