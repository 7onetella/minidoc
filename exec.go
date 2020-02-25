package minidoc

import (
	"bufio"
	"fmt"
	"os/exec"
)

// Execute execute
func Execute(args []string) (string, error) {
	output, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	return string(output), err
}

// Exec exec.Command
func Exec(args []string) error {
	cmd := exec.Command(args[0], args[1:]...)

	stdOut, _ := cmd.StdoutPipe()
	stdErr, _ := cmd.StderrPipe()
	//stdIn, _ := cmd.StdinPipe()

	err := cmd.Start()
	if err != nil {
		fmt.Println("terminated early: " + err.Error())
		return err
	}

	go func() {
		scanner := bufio.NewScanner(stdOut)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stdErr)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
	}()

	// if exec.Command calls a service that blocks then this code will never be reached
	err = cmd.Wait()

	return err
}
