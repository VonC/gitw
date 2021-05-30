// +build windows

package shell

import (
	"fmt"
	"gitw/internal/syscall"
	"gitw/internal/xregexp"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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

func GetPSStartDate(apid Pid, verbose bool) (*time.Time, error) {
	//cmd := `ps -ewo pid,lstart|grep -E "^\s+?%s "`
	cmd := `wmic process get handle,CreationDate|grep -E "\b%s\s*?$"`
	cmd = fmt.Sprintf(cmd, apid)
	serr, sout, err := syscall.ExecCmd(cmd)
	if verbose {
		fmt.Printf("sout(%s)='%s'\n", cmd, sout.String())
	}
	if err != nil {
		return nil, fmt.Errorf("error: unable to get ps start time: '%+v', serr '%s'", err, serr.String())
	}
	//  20210515081628.338329+120  21992
	res := strings.Split(sout.String(), " ")
	sdate := res[0]
	if verbose {
		fmt.Printf("stime='%s'\n", sdate)
	}
	dd := strings.Split(sdate, "+")
	offset := ""
	sep := ""
	if len(dd) == 2 {
		sdate = dd[0]
		offset = dd[1]
		sep = "+"
	} else {
		dd = strings.Split(sdate, "-")
		if len(dd) == 2 {
			sdate = dd[0]
			offset = dd[1]
			sep = "-"
		} else {
			log.Fatalf("Unable to extract offset from sdate '%s'", sdate)
		}
	}
	var ioffset int
	if ioffset, err = strconv.Atoi(offset); err != nil {
		log.Fatalf("Unable to convert offset '%s' into int: %+v", offset, err)
	}
	ioffset = ioffset / 60
	sdate = fmt.Sprintf("%s%s%02d", sdate, sep, ioffset)
	// Wed Aug 28 10:16:24 2019
	date, err := time.Parse("20060202150405.999999-07", sdate)
	if err != nil {
		return nil, fmt.Errorf("error: unable to parse time: '%+v', serr '%s'", sdate, err)
	}

	return &date, nil
}
