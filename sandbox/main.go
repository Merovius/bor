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

// Driver implement everything we need from a sandbox
type Driver interface {
	Command(name string, arg ...string) Cmd // Create a command (don't run it yet)
	Config(*goconf.ConfigFile) error        // Called at the beginning, can be used to define own configuration variables. If the return value is not nil, bor outputs it and aborts. See the easysandbox package for an example
}

// Cmd implements everything we need to be able to do to a running command
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

// ProcessState is basically a wrapper around *os.ProcessState. We can not use
// that, because not all sandboxes must have a process running on this machine
// (for example using VMs for sandboxing)
type ProcessState interface {
	Exited() bool
	Pid() int
	String() string
	Success() bool
	SystemTime() time.Duration
	UserTime() time.Duration
}

// Register a driver with a given name. It is an error to register a name
// that is already taken
func Register(name string, driver Driver) error {
	if _, exists := drivers[name]; exists {
		return fmt.Errorf("Sandbox driver %s already registered")
	}
	drivers[name] = driver
	return nil
}

// Command wraps the Command-method of the given driver
func Command(driver string, name string, arg ...string) Cmd {
	if dr, ok := drivers[driver]; ok {
		return dr.Command(name, arg...)
	}
	panic(fmt.Errorf("No such Sandbox driver: %s", driver))
}

// Config calls the Config-method of all registered drivers
func Config(cfg *goconf.ConfigFile) error {
	for _, dr := range drivers {
		err := dr.Config(cfg)
		if err != nil {
			return err
		}
	}
	return nil
}

// TimeoutError reperesents an error due to a timeout
type TimeoutError struct{}

// Error returns the string "Timeout"
func (e TimeoutError) Error() string {
	return "Timeout"
}

// TimeoutCombinedOutput works like exec.Cmd.CombinedOutput(), except a timeout
// is given, after which the Process is automatically killed
func TimeoutCombinedOutput(cmd Cmd, timeout time.Duration) ([]byte, error) {
	// We need to buffer the output
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
