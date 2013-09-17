// Package plain implements the sandbox-interface just wrapping os/exec (i.e.
// providing no sandboxing)
package plain

import (
	goconf "code.google.com/p/goconf/conf"
	"github.com/Merovius/bor/sandbox"
	"io"
	"os/exec"
)

type Driver struct{}

type Cmd struct {
	*exec.Cmd
}

func (c Cmd) Dir() string {
	return c.Cmd.Dir
}

func (c Cmd) SetDir(dir string) {
	c.Cmd.Dir = dir
}

func (c Cmd) ProcessState() sandbox.ProcessState {
	return c.Cmd.ProcessState
}

func (c Cmd) Kill() error {
	return c.Cmd.Process.Kill()
}

func (c Cmd) SetStderr(w io.Writer) {
	c.Cmd.Stderr = w
}

func (c Cmd) SetStdout(w io.Writer) {
	c.Cmd.Stdout = w
}

func (d Driver) Command(name string, arg ...string) sandbox.Cmd {
	return Cmd{exec.Command(name, arg...)}
}

func (d Driver) Config(_ *goconf.ConfigFile) error {
	return nil
}

func init() {
	sandbox.Register("plain", Driver{})
}
