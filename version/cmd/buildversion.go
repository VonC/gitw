package main

import (
	"fmt"
	"gitw/internal/syscall"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	// Load the file content
	vFileData, _ := os.ReadFile("version/version.txt")

	// Convert from Byte array to string and split
	// on newlines. We now have a slice of strings
	vLines := strings.Split(string(vFileData), "\n")

	if len(vLines) != 4 {
		log.Fatalf("version.txt incorrect: expected 4 lines, got %d", len(vLines))
	}

	// Generate a timestamp.
	bTime := time.Now().Format("20060102-150405")

	// Load the count from the 3rd line of the file
	// It's a string so we need to convert to integer
	// Then increment it by 1
	bNum, err := strconv.Atoi(vLines[2])
	if err != nil {
		log.Fatalf("Unable to convert build number '%s': %+v", vLines[2], err)
	}
	bNum++

	// https://medium.com/@joshroppo/setting-go-1-5-variables-at-compile-time-for-versioning-5b30a965d33e
	// https://stackoverflow.com/questions/38087256/dynamic-version-from-git-with-go-get
	gitout, gitver, err := syscall.ExecCmd("git describe --long --tags --always --dirty")
	if err != nil || gitout.String() != "" {
		log.Fatalf("Unable to git describe: '%s' (%+v)", gitout.String(), err)
	}

	// Generate a single string to write back to the file
	// Note, we didn't change the version string
	outStr := vLines[0] + "\n" + bTime + "\n" + fmt.Sprint(bNum) + "\n" + strings.TrimSpace(gitver.String())

	// Write the data back to the file.
	_ = os.WriteFile("version/version.txt", []byte(outStr), 0777)
}
