package shell

import (
	"fmt"
	"gitw/internal/syscall"
	"log"
	"strconv"
	"strings"
	"time"
)

type Pid string

type Ps struct {
	pid Pid
	cmd string
}

func (p *Ps) isGitW() bool {
	return strings.Contains(p.cmd, "gitw")
}

func GetBashPID(verbose bool) (Pid, error) {
	p, err := GetParentPS(Pid(""), verbose)
	// fmt.Printf("ps ='%+v'\nerr='%+v'\n", p, err)
	if err != nil {
		return Pid(""), err
	}
	for !p.isGitW() {
		p, err = GetParentPS(p.pid, verbose)
		// fmt.Printf("psng ='%+v'\nerr='%+v'\n", p, err)
		if err != nil {
			return Pid(""), err
		}
	}
	var bpid Pid
	for p.isGitW() {
		bpid = p.pid
		p, err = GetParentPS(p.pid, verbose)
		// fmt.Printf("psig ='%+v'\nerr='%+v'\n", p, err)
		if err != nil {
			return Pid(""), err
		}
	}
	//fmt.Println(">>>" + string(bpid))
	return bpid, nil
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
