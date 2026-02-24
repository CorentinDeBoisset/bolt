package cmdrunr

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
)

var ErrPlannedKill = errors.New("process planned to be killed")

type SafeBuffer struct {
	buf bytes.Buffer
	mtx sync.RWMutex
}

func (s *SafeBuffer) Write(p []byte) (n int, err error) {
	s.mtx.Lock()
	defer s.mtx.Unlock()
	return s.buf.Write(p)
}

func (s *SafeBuffer) Bytes() []byte {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	output := make([]byte, s.buf.Len())
	_ = copy(output, s.buf.Bytes())
	return output
}

func getCmdPath(basePath, cmdPath string) string {
	if filepath.IsAbs(cmdPath) {
		return cmdPath
	} else {
		return filepath.Join(basePath, cmdPath)
	}
}

func RunCommand(ctx context.Context, basePath, path, cmd string, output *SafeBuffer, width, height int) bool {
	// Note: this only works on Unix platforms
	// TODO: maybe add support for windows (or not... :shrug:)
	_, _ = fmt.Fprintf(output, "> %s\n", cmd)
	task := exec.CommandContext(ctx, "/bin/sh", "-c", cmd)
	task.Dir = getCmdPath(basePath, path)

	// Pass the environment to the child processes, and set the WIDTH/HEIGHT env variables
	task.Env = append(os.Environ(), fmt.Sprintf("COLUMNS=%d", width), fmt.Sprintf("LINES=%d", height))
	task.Stdout = output
	task.Stderr = output
	task.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	task.Cancel = func() error {
		return syscall.Kill(-task.Process.Pid, syscall.SIGKILL)
	}

	if err := task.Start(); err != nil {
		_, _ = fmt.Fprintf(output, "\n\nThe command could not start due to the following error:\n%s", err.Error())
		return false
	}

	err := task.Wait()
	if err != nil {
		if errors.Is(context.Cause(ctx), ErrPlannedKill) {
			_, _ = fmt.Fprintf(output, "\nService killed\n\n")
			return true
		} else {
			_, _ = fmt.Fprintf(output, "\n\nThe command failed with the following error:\n%s", err.Error())
			return false
		}
	}

	_, _ = fmt.Fprintf(output, "\n\n")

	return true
}
