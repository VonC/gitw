//go:build (linux && !android && !nox11) || freebsd || openbsd || !windows
// +build linux,!android,!nox11 freebsd openbsd !windows

package main

func getOriginalConsoleMode() (uintptr, uint32) {
	return 0, 0
}

func resetConsoleMore(con uintptr, originalConsoleMode uint32) {
	return
}

func GetColorProfile() termenv.Profile {
	return 0
}
