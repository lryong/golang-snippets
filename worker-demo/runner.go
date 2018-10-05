package main

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

// runner is to run os commands giving command line, env and log file
// it's an alternative to pyhton-sh or go-sh

var errProcessNotStarted = errors.New("Process Not Started")

type cmdJob struct {
	sync.Mutex
	cmd        *exec.Cmd
	workingDir string
	env        map[string]string
	logFile    *os.File
	finished   chan empty
	provider   mirrorProvider
	retErr     error
}

func newCmdJob(provider mirrorProvider, cmdAndArgs []string, workingDir string, env map[string]string) *cmdJob {
	return nil
}

func (c *cmdJob) Start() error {
	c.finished = make(chan empty, 1)
	return c.cmd.Start()
}

func (c *cmdJob) Wait() error {
	c.Lock()
	defer c.Unlock()

	select {
	case <-c.finished:
		return c.retErr
	default:
		err := c.cmd.Wait()
		if c.cmd.Stdout != nil {
			c.cmd.Stdout.(*os.File).Close()
		}
		c.retErr = err
		close(c.finished)
		return err
	}
}

func (c *cmdJob) Terminate() error {
	if c.cmd == nil || c.cmd.Process == nil {
		return errProcessNotStarted
	}

	if d := c.provider.Docker(); d != nil {
		sh.Command(
			"docker", "stop", "-t", "2", d.Name(),
		).Run()
		return nil
	}

	err := unix.Kill(c.cmd.Process.Pid, syscall.SIGTERM)
	if err != nil {
		return err
	}

	select {
	case <-time.After(2 * time.Second):
		unix.Kill(c.cmd.Process.Pid, syscall.SIGKILL)
		return errors.New("SIGTERM failed to kill the job")
	case <-c.finished:
		return nil
	}
}

func newEnviron(env map[string]string, inherit bool) []string {
	environ := make([]string, 0, len(env))
	if inherit {
		for _, line := range os.Environ() {
			// if os enviroment and env collapses,
			// omit the os one
			k := strings.Split(line, "=")[0]
			if _, ok := env[k]; ok {
				continue
			}
			environ = append(environ, line)
		}
	}
	for k, v := range env {
		environ = append(environ, k+"="+v)
	}
	return environ
}
