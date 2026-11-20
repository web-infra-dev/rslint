//go:build unix || js || wasip1

package main

import "syscall"

type fixFileID struct {
	device uint64
	inode  uint64
}

func statFixFile(path string) (fixFileID, error) {
	var stat syscall.Stat_t
	if err := syscall.Stat(path, &stat); err != nil {
		return fixFileID{}, err
	}
	return fixFileID{uint64(stat.Dev), stat.Ino}, nil
}
