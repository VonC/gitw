//go:build windows
// +build windows

package main

import (
	"fmt"
	"log"

	"github.com/erikgeiser/coninput"
	"github.com/muesli/termenv"
	"golang.org/x/sys/windows"
)

func getOriginalConsoleMode() (uintptr, uint32) {
	con, err := windows.GetStdHandle(windows.STD_INPUT_HANDLE)
	if err != nil {
		log.Fatalf("get stdin handle: %s", err)
	}

	var originalConsoleMode uint32
	err = windows.GetConsoleMode(con, &originalConsoleMode)
	if err != nil {
		log.Fatalf("get console mode: %s", err)
	}
	return uintptr(con), originalConsoleMode
}

func resetConsoleMore(con uintptr, originalConsoleMode uint32) {
	// https://github.com/charmbracelet/bubbletea/issues/121
	// https://github.com/erikgeiser/coninput/blob/main/example/main.go
	// https://github.com/microsoft/terminal/issues/8750#issuecomment-759088381
	ccon, ccor := getOriginalConsoleMode()
	fmt.Printf("Restore con %d (vs. current %d), orig %d (vs. current %d)\n", con, ccon, originalConsoleMode, ccor)

	fmt.Println("\noriginalConsoleMode:", coninput.DescribeInputMode(originalConsoleMode))
	fmt.Println("\nccor:", coninput.DescribeInputMode(ccor))

	resetErr := windows.SetConsoleMode(windows.Handle(con), originalConsoleMode)
	if resetErr != nil {
		log.Fatalf("reset console mode: %s", resetErr)
	}
	//windows.SetConsoleMode(1748, 7)
	//windows.SetConsoleMode(84, 7)
}

func GetColorProfile() termenv.Profile {
	return termenv.ANSI
}
