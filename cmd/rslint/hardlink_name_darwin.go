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

// The kernel can report the name of another hard link. Only consider directory
// entries in the requested parent that identify the same object. An empty name
// means that the underlying lookup could not be attributed to one entry.
func darwinHardLinkEntryName(parent, requested, reportedPath string, original os.FileInfo) string {
	const pcCaseSensitive = 11 // Darwin's _PC_CASE_SENSITIVE selector.
	if filepath.Join(parent, requested) == reportedPath {
		return requested
	}
	entries, err := os.ReadDir(parent)
	if err != nil {
		// A reported link in another parent says nothing about this entry's
		// spelling, even when both names open the same inode.
		if reportedPath == "" {
			return ""
		}
		dir, _ := os.Stat(parent)
		reportedDir, _ := os.Stat(filepath.Dir(reportedPath))
		if !os.SameFile(dir, reportedDir) {
			return ""
		}
		reported := filepath.Base(reportedPath)
		if requested == reported {
			return reported
		}
		// On a case-insensitive volume an ASCII case change cannot name a
		// separate entry, even when the parent allows search but not listing.
		if strings.EqualFold(requested, reported) && !strings.ContainsFunc(requested+reported, func(r rune) bool { return r > unicode.MaxASCII }) {
			sensitive, err := unix.Pathconf(parent, pcCaseSensitive)
			candidate, _ := os.Lstat(reportedPath)
			if err == nil && sensitive == 0 && os.SameFile(original, candidate) {
				return reported
			}
		}
		return ""
	}
	for _, entry := range entries {
		if entry.Name() == requested {
			return requested
		}
	}
	var names []string
	for _, entry := range entries {
		if entry.Type() != original.Mode().Type() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return ""
		}
		if os.SameFile(original, info) {
			names = append(names, entry.Name())
		}
	}
	if len(names) == 1 {
		return names[0]
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
				return ""
			}
			match = name
		}
	}
	return match
}
