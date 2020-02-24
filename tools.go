package minidoc

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"
)

func WriteToFile(filepath, content string) (done bool) {
	file, err := os.Create(filepath)
	if err != nil {
		log.Errorf("creating file: %v", err)
		return true
	}
	defer file.Close()
	_, err = fmt.Fprintf(file, content)
	if err != nil {
		log.Errorf("writing content: %v", err)
		return true
	}
	return false
}

func ReadFromFile(filepath string) (string, error) {
	dat, err := ioutil.ReadFile(filepath)
	if err != nil {
		log.Errorf("reading content: %v", err)
		return "", err
	}

	return string(dat), nil
}

func DeleteFile(filepath string) (success bool) {
	err := os.Remove(filepath)
	if err != nil {
		log.Errorf("creating file: %v", err)
		return false
	}
	return true
}

// this works perfectly
func OpenVim(app *SimpleApp, filepath string) {
	app.Suspend(func() {
		cmd := exec.Command("vim", filepath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Errorf("opening vi: %v", err)
		}
		log.Debug("returning the control back")
	})
}

// works create, all key inputs works, exit the minidoc since this seems to be replace the process
func (s *Search) openVim(filepath string) {
	binary, lookErr := exec.LookPath("vim")
	if lookErr != nil {
		panic(lookErr)
	}

	args := []string{"vim", filepath}

	env := os.Environ()
	execErr := syscall.Exec(binary, args, env)
	if execErr != nil {
		panic(execErr)
	}
	s.App.Draw()
}

func (s *Search) openVimForkExec(filepath string) {
	cmd := "vim"
	binary, lookErr := exec.LookPath(cmd)
	if lookErr != nil {
		panic(lookErr)
	}
	//fmt.Println(binary)

	os.Remove("/tmp/stdin")
	os.Remove("/tmp/stdout")
	os.Remove("/tmp/stderr")

	fstdin, err1 := os.Create("/tmp/stdin")
	fstdout, err2 := os.Create("/tmp/stdout")
	fstderr, err3 := os.Create("/tmp/stderr")
	if err1 != nil || err2 != nil || err3 != nil {
		log.Errorf("%v %v %v", err1, err2, err3)
		panic("WOW")
	}

	env := os.Environ()

	argv := []string{filepath}
	procAttr := syscall.ProcAttr{
		Dir:   "/tmp",
		Files: []uintptr{fstdin.Fd(), fstdout.Fd(), fstderr.Fd()},
		Env:   env,
		Sys: &syscall.SysProcAttr{
			Foreground: false,
		},
	}

	pid, err := syscall.ForkExec(binary, argv, &procAttr)
	log.Debugf("pid=%d err=%v", pid, err)
	s.App.Draw()
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}
