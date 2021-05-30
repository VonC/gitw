// +build linux
package shell

import (
	"fmt"
	"gitw/internal/syscall"
	"gitw/internal/xregexp"
	"regexp"
	"time"
)

//   PPID
// 64443 ps -p 64447 -o ppid,cmd=
var repid = regexp.MustCompile(`(?m)^\s*?(?P<pid>\d+)\s+(?P<cmd>.*?)$`)

func GetParentPS(apid Pid, verbose bool) (*Ps, error) {
	spid := string(apid)
	if spid == "" {
		spid = "$(echo $$)"
	}
	serr, sout, err := syscall.ExecCmd(fmt.Sprintf("ps -p %s -o ppid,cmd=", spid))
	if err != nil || serr.String() != "" {
		err := fmt.Errorf("Error: unable to get PID of current bash session: '%+v', serr '%s'", err, serr.String())
		return nil, err
	}
	// fmt.Println("===" + sout.String())
	var p *Ps
	matches := xregexp.FindNamedMatches(repid, sout.String())
	if len(matches) > 0 {
		p = &Ps{pid: Pid(matches["pid"]), cmd: matches["cmd"]}
	} else {
		return nil, fmt.Errorf("Error: unable to get ps from: '%s'", sout.String())
	}
	return p, err
}

func GetPSStartDate(apid Pid, verbose bool) (*time.Time, error) {
	cmd := `ps -ewo pid,lstart|grep -E "^\s+?%s "`
	cmd = fmt.Sprintf(cmd, apid)
	serr, sout, err := syscall.ExecCmd(cmd)
	if verbose {
		fmt.Printf("sout(%s)='%s'\n", cmd, sout.String())
	}
	if err != nil {
		return nil, fmt.Errorf("Error: unable to get bash time: '%+v', serr '%s'", err, serr.String())
	}
	//  25301 Wed Aug 28 10:16:24 2019 -bash
	resession := regexp.MustCompile(`(?m)^\s*?\d+\s+?(?P<date>.*?)$`)
	matches := xregexp.FindNamedMatches(resession, sout.String())
	if len(matches) > 0 {
		if verbose {
			fmt.Printf("Session date '%s'\n", matches["date"])
		}
	} else {
		return nil, fmt.Errorf("Error: unable to get bash time from: '%s'", sout.String())
	}
	if verbose {
		fmt.Printf("stime='%s'\n", matches["date"])
	}
	// Wed Aug 28 10:16:24 2019
	date, err := time.Parse("Mon Jan 2 15:04:05 2006", matches["date"])
	if err != nil {
		return nil, fmt.Errorf("Error: unable to parse time: '%+v', serr '%s'", matches["date"], err)
	}
	if verbose {
		fmt.Printf("date='%+v'\n", date)
	}
	return &date, nil
}
