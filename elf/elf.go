package elf

import (
	"os"
	"fmt"
	"exec"
)

func AssembleAndLink(fn string, code []byte) (err os.Error) {
	o,err := os.Open(fn+".S", os.O_WRONLY + os.O_CREAT + os.O_TRUNC, 0666)
	if err != nil {
		o.Close()
		return
	}
	_,err = o.Write(code)
	o.Close()
	if err != nil {
		return
	}
	err = justrun("as", "--32", "--fatal-warnings", "-o", fn+".o", fn+".S")
	if err != nil {
		return
	}
	err = justrun("ld", "-s", "-o", fn, fn+".o")
	return
}

func justrun(cmd string, args ...string) os.Error {
	abscmd,err := exec.LookPath(cmd)
	if err != nil { return os.NewError("Couldn't find "+cmd+": "+err.String()) }
	
	cmdargs := make([]string, len(args)+1)
	cmdargs[0] = cmd
	for i,a := range args {
		cmdargs[i+1] = a
	}
	pid, err := exec.Run(abscmd, cmdargs, nil, "",
		exec.PassThrough, exec.PassThrough, exec.PassThrough)
	if err != nil { return err }
	wmsg,err := pid.Wait(0)
	if err != nil { return err }
	if wmsg.ExitStatus() != 0 {
		return os.NewError(cmd+" exited with status "+fmt.Sprint(wmsg.ExitStatus()))
	}
	return nil
}
