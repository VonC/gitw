package version

import (
	_ "embed"
	"fmt"
	"log"
	"strings"
)

var (
	// Version : current version
	Version string = strings.TrimSpace(version)
	//go:embed version.txt
	version string
	ExeDir  string
)

// String displays all the version values
func String() string {
	vData := strings.Split(Version, "\n")
	if len(vData) != 4 {
		log.Fatalf("Embedded version data is badly formed or missing: %d lines instead of 4", len(vData))
	}
	res := ""
	res = res + fmt.Sprintf("Version     : %s\n", vData[0])
	res = res + fmt.Sprintf("BuildDate   : %s\n", vData[1])
	res = res + fmt.Sprintf("BuildNumber : %s\n", vData[2])
	res = res + fmt.Sprintf("Git Tag     : %s\n", vData[3])
	res = res + fmt.Sprintf("ExeDir      : %s\n", ExeDir)
	return res
}
