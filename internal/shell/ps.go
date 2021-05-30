package shell

import (
	"strings"
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
