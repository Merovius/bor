package sandbox

import (
	"bytes"
	goconf "code.google.com/p/goconf/conf"
	"fmt"
	"io"
	"time"
)

var (
	drivers = make(map[string]Driver)
)

type Driver interface {
	Command(name string, arg ...string) Cmd
	Config(*goconf.ConfigFile) error
}

type Cmd interface {
	CombinedOutput() ([]byte, error)
	Output() ([]byte, error)
	Run() error
	Start() error
	StderrPipe() (io.ReadCloser, error)
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	Wait() error
	Kill() error

	Dir() string
	SetDir(string)
	SetStdout(io.Writer)
	SetStderr(io.Writer)

	ProcessState() ProcessState
}

type ProcessState interface {
	Exited() bool
	Pid() int
	String() string
	Success() bool
	SystemTime() time.Duration
	UserTime() time.Duration
}

func Register(name string, driver Driver) error {
	if _, exists := drivers[name]; exists {
		return fmt.Errorf("Sandbox driver %s already registered")
	}
	drivers[name] = driver
	return nil
}

func Command(driver string, name string, arg ...string) Cmd {
	if dr, ok := drivers[driver]; ok {
		return dr.Command(name, arg...)
	}
	panic(fmt.Errorf("No such Sandbox driver: %s", driver))
}

func Config(cfg *goconf.ConfigFile) error {
	for _, dr := range drivers {
		err := dr.Config(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

type TimeoutError struct{}

func (e TimeoutError) Error() string {
	return "Timeout"
}

func TimeoutCombinedOutput(cmd Cmd, timeout time.Duration) ([]byte, error) {
	outbuf := bytes.NewBuffer(make([]byte, 0, 8388608))

	cmd.SetStdout(outbuf)
	cmd.SetStderr(outbuf)

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	to := time.After(timeout)
	ch := make(chan error)

	go func() {
		err := cmd.Wait()
		ch <- err
	}()

	select {
	case <-to:
		cmd.Kill()
		return outbuf.Bytes(), TimeoutError{}
	case err = <-ch:
		return outbuf.Bytes(), err
	}
}
