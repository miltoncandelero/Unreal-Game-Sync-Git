package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

var (
	ErrApplicationNotFound = errors.New("application not found")
	//ErrApplicationFailed   = errors.New("application execution failed")
)

type ErrExec struct {
	ExitCode  int
	Output    string
	ErrOutput string
	Cmd       string
	Args      []string
}

func (e ErrExec) Error() string {
	out := e.Output
	if len(e.ErrOutput) > len(out) {
		out = e.ErrOutput
	}
	return fmt.Sprintf("application execution failed: %s %s: rc %d: %s",
		e.Cmd,
		strings.Join(e.Args, " "),
		e.ExitCode,
		out,
	)
}

func Execute(workingDir, cmd string, args ...string) ([]string, error) {

	outStr, err := ExecuteOneLine(workingDir, cmd, args...)

	if err != nil {
		return nil, err
	}

	lines := strings.Split(outStr, "\n")
	for idx, line := range lines {
		lines[idx] = strings.TrimSpace(line)
	}

	return lines, nil
}

func ExecuteOneLine(workingDir, cmd string, args ...string) (string, error) {
	_, err := exec.LookPath(cmd)
	if err != nil {
		return "", fmt.Errorf("%w: %s", ErrApplicationNotFound, cmd)
	}

	c := exec.Command(cmd, args...)
	if workingDir != "" {
		c.Dir = workingDir
	}
	c.Env = os.Environ()

	// combined contains stdout and stderr but stderr only contains stderr output
	combinedOut := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}

	c.Stderr = io.MultiWriter(combinedOut, stderrBuf)
	c.Stdout = combinedOut

	// fmt.Printf("Executing: %s\n", c.String())

	err = c.Run()
	if err != nil {
		return "", ErrExec{
			ExitCode:  c.ProcessState.ExitCode(),
			Output:    strings.TrimSpace(combinedOut.String()),
			ErrOutput: strings.TrimSpace(stderrBuf.String()),
			Cmd:       cmd,
			Args:      args,
		}
	}

	outStr := combinedOut.String()
	return outStr, nil
}
