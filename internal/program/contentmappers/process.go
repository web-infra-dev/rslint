package contentmappers

import (
	"errors"
	"io"
	"os/exec"
	"time"
)

// spawnProcess mirrors ts-go's own mapper spawner: Read is the child's stdout,
// Write its stdin, and Close kills and reaps it.
func spawnProcess(command []string, dir string, stderr io.Writer) (io.ReadWriteCloser, error) {
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Dir = dir
	cmd.Stderr = stderr
	cmd.WaitDelay = time.Second
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	return &childProcess{cmd: cmd, stdin: stdin, stdout: stdout}, nil
}

type childProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.Reader
}

func (p *childProcess) Read(b []byte) (int, error)  { return p.stdout.Read(b) }
func (p *childProcess) Write(b []byte) (int, error) { return p.stdin.Write(b) }

func (p *childProcess) ExitCode() (int, bool) {
	if p.cmd.ProcessState == nil {
		return 0, false
	}
	return p.cmd.ProcessState.ExitCode(), true
}

func (p *childProcess) Close() error {
	_ = p.stdin.Close()
	_ = p.cmd.Process.Kill()
	err := p.cmd.Wait()
	if _, ok := errors.AsType[*exec.ExitError](err); ok {
		return nil
	}
	if errors.Is(err, exec.ErrWaitDelay) {
		return nil
	}
	return err
}
