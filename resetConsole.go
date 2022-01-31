//go:build (linux && !android && !nox11) || freebsd || openbsd || !windows
// +build linux,!android,!nox11 freebsd openbsd !windows

package main

import "fmt"

func getResetConsoleMore() func() {
	return func() {
		fmt.Println("resetConsole Linux")
	}
}
