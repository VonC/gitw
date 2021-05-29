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

// https://regex101.com/r/67F20E/1
var repid = regexp.MustCompile(`(?m)^(?P<cmd>.*?)\s+\d+\s+(?P<pid>\d+)$`)

func GetParentPS(apid Pid) (*Ps, error) {
	spid := string(apid)
	if spid == "" {
		spid = fmt.Sprintf("%d", os.Getpid())
	}
	serr, sout, err := syscall.ExecCmd(fmt.Sprintf(`wmic process get processid,parentprocessid,executablepath|C:\Windows\System32\find.exe "%s"`, spid))
	if err != nil || serr.String() != "" {
		err := fmt.Errorf("error: unable to get PID of current bash session: '%+v', serr '%s'", err, serr.String())
		return nil, err
	}
	var p *Ps
	matches := xregexp.FindNamedMatches(repid, sout.String())
	if len(matches) > 0 {
		p = &Ps{pid: Pid(matches["pid"]), cmd: matches["cmd"]}
	} else {
		return nil, fmt.Errorf("error: unable to get ps from: '%s'", sout.String())
	}
	return p, err
}
