//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/windows"
)

func getResetConsoleMore() func() {
	// https://github.com/charmbracelet/bubbletea/issues/121
	// https://github.com/erikgeiser/coninput/blob/main/example/main.go
	con, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		log.Fatalf("get stdin handle: %s", err)
	}

	var originalConsoleMode uint32
	err = windows.GetConsoleMode(con, &originalConsoleMode)
	if err != nil {
		log.Fatalf("get console mode: %s", err)
	}

	return func() {
		fmt.Println("resetConsole Windows")
		resetErr := windows.SetConsoleMode(con, originalConsoleMode)
		if err == nil && resetErr != nil {
			log.Fatalf("reset console mode: %s", resetErr)
		}
	}
}
