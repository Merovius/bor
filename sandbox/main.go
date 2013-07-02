package sandbox

import (
	"fmt"
	"io"
	"time"
)

var (
	drivers = make(map[string]Driver)
)

type Driver interface {
	Command(name string, arg ...string) Cmd
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

	Dir() string
	SetDir(string)

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