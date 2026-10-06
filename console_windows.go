package main

import (
	"unsafe"

	"golang.org/x/sys/windows"
)

var procGetConsoleProcessList = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleProcessList")

// isDropInvocation reports whether this process owns its console alone, which
// happens when Explorer launches it (a Drop Invocation, or a double-click)
// rather than a terminal that shares its console.
func isDropInvocation() bool {
	var pids [2]uint32
	n, _, _ := procGetConsoleProcessList.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	return n == 1
}
