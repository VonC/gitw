//go:build (linux && !android && !nox11) || freebsd || openbsd || !windows
// +build linux,!android,!nox11 freebsd openbsd !windows

package main

import "fmt"

func getOriginalConsoleMode() (windows.Handle, uint32) {
	return 0, 0
}

func resetConsoleMore(con windows.Handle, originalConsoleMode uint32) {
	return
}
