package main

import (
	"golang.org/x/sys/windows"
)

// enableVirtualTerminalProcessing arms ANSI rendering on the inherited
// console stderr. Under IPC mode — the only native lint mode — this process's
// stdout is Node's pipe (lint output renders in the Node parent, whose tty
// layer owns VT processing there), so stderr is the only handle this process
// still writes ANSI to directly (the syntactic-error pretty report).
func enableVirtualTerminalProcessing() {
	h, err := windows.GetStdHandle(windows.STD_ERROR_HANDLE)
	if err != nil || h == windows.InvalidHandle {
		return
	}
	fileType, err := windows.GetFileType(h)
	if err != nil || fileType == windows.FILE_TYPE_CHAR {
		var mode uint32
		if err := windows.GetConsoleMode(h, &mode); err != nil {
			return
		}
		if mode&windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING == 0 {
			_ = windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING)
		}
	}
}
