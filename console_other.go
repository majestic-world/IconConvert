//go:build !windows

package main

// isDropInvocation is Windows-only; elsewhere the tool always runs from a terminal.
func isDropInvocation() bool { return false }
