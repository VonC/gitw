// +build windows

package shell

import (
	"fmt"
	"gitw/internal/syscall"
	"os"
)

func CleanOldBashFiles(verbose bool) error {
	cmd := `find %TMP% -maxdepth 1 -type f -mtime -1 -name "bash.*" -exec rm -f {} ;`
	serr, sout, err := syscall.ExecCmd(cmd)
	if verbose {
		fmt.Printf("sout(%s)='%s'\n", cmd, sout.String())
	}
	if err != nil || sout.String() != "" || serr.String() != "" {
		return fmt.Errorf("error: unable to clean old /tmp/bash files: '%+v', serr '%s'", err, serr.String())
	}
	return nil
}

func TempPath() string {
	return os.Getenv("TMP")
}
