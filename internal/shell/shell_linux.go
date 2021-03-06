// +build linux

package shell

import (
	"fmt"
	"gitw/internal/syscall"
)

func CleanOldBashFiles(verbose bool) error {
	// https://unix.stackexchange.com/a/112407/7490 https://unix.stackexchange.com/questions/92346/why-does-find-mtime-1-only-return-files-older-than-2-days
	cmd := `find /tmp -maxdepth 1 -user $USER -type f -mtime +1 -name "bash.*" -exec rm -f {} \;`
	serr, sout, err := syscall.ExecCmd(cmd)
	if verbose {
		fmt.Printf("sout(%s)='%s'\n", cmd, sout.String())
	}
	if err != nil || sout.String() != "" || serr.String() != "" {
		return fmt.Errorf("Error: unable to clean old /tmp/bash files: '%+v', serr '%s'", err, serr.String())
	}
	return nil
}

func TempPath() string {
	return "/tmp"
}
