//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows"
)

func getOriginalConsoleMode() (windows.Handle, uint32) {
	con, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		log.Fatalf("get stdin handle: %s", err)
	}

	var originalConsoleMode uint32
	err = windows.GetConsoleMode(con, &originalConsoleMode)
	if err != nil {
		log.Fatalf("get console mode: %s", err)
	}
	return con, originalConsoleMode
}

func resetConsoleMore(con windows.Handle, originalConsoleMode uint32) {
	// https://github.com/charmbracelet/bubbletea/issues/121
	// https://github.com/erikgeiser/coninput/blob/main/example/main.go
	ccon, ccor := getOriginalConsoleMode()
	fmt.Printf("Restore con %d (vs. current %d), orig %d (vs. current %d)", con, ccon, originalConsoleMode, ccor)
	resetErr := windows.SetConsoleMode(con, 992)
	resetErr = windows.SetConsoleMode(con, 503)
	if resetErr != nil {
		log.Fatalf("reset console mode: %s", resetErr)
	}
}
