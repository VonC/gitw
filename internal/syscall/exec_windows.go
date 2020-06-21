// +build windows

package syscall

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"syscall"
)

// ExecCmd starts a sh -c 'scmd' session.
// If scmd ends with &, don't wait for result (background process)
func ExecCmd(scmd string) (berr *bytes.Buffer, bout *bytes.Buffer, err error) {
	berr = &bytes.Buffer{}
	bout = &bytes.Buffer{}
	cmd := exec.Command("cmd", "/C", scmd)
	cmd.SysProcAttr = &syscall.SysProcAttr{}
	cmd.SysProcAttr.CmdLine = "cmd /C " + scmd
	//fmt.Printf("Execute '%s'\n%+v\n", scmd, cmd.SysProcAttr.CmdLine)
	log.Printf("Execute '%s'\n", scmd)
	fmt.Printf("Execute '%s'\n", scmd)
	cmd.Stderr = berr
	cmd.Stdout = bout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("stdin error %s [%s]", err, berr.String())
		return berr, bout, err
	}
	err = stdin.Close()
	if err != nil {
		log.Printf("Close error ion stdin %s", err)
		return berr, bout, err
	}
	err = cmd.Start()
	if err != nil {
		log.Printf("start error %s [%s]", err, berr.String())
		return berr, bout, err
	}
	if strings.HasPrefix(scmd, "start ") && strings.Contains(scmd, " /B ") {
		return berr, bout, nil
	}
	err = cmd.Wait()
	if err != nil {
		log.Printf("exit error %s [%s]", err, berr.String())
	}
	return berr, bout, err
}
