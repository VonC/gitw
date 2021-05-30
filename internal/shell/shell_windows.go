// +build windows

package shell

import (
	"fmt"
	"gitw/internal/syscall"
	"gitw/internal/xregexp"
	"os"
	"regexp"
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

// https://regex101.com/r/H5iUEd/1
var repid = regexp.MustCompile(`(?m)^(?P<cmd>.*?)\s+(?P<ppid>\d+)\s+\d+\s*?$`)

func GetParentPS(apid Pid, verbose bool) (*Ps, error) {
	spid := string(apid)
	if spid == "" {
		spid = fmt.Sprintf("%d", os.Getpid())
	}
	cmd := fmt.Sprintf(`wmic process get processid,parentprocessid,executablepath|C:\Windows\System32\find.exe "%s"`, spid)
	serr, sout, err := syscall.ExecCmd(cmd)
	if err != nil || serr.String() != "" {
		err := fmt.Errorf("error: unable to get PID of current bash session: '%+v', serr '%s'", err, serr.String())
		return nil, err
	}
	var p *Ps
	res := sout.String()
	matches := xregexp.FindNamedMatches(repid, res)
	if verbose {
		fmt.Printf("cmd='%s'\nparentps of '%s'='%+v'\nregexp='%s'\nmatches='%+v'\n", cmd, spid, res, repid.String(), matches)
	}
	if len(matches) > 0 {
		p = &Ps{pid: Pid(matches["ppid"]), cmd: matches["cmd"]}
	} else {
		return nil, fmt.Errorf("error: unable to get ps from: '%s'", res)
	}
	return p, err
}

func TempPath() string {
	return os.Getenv("TMP")
}
