package plain

import (
	"github.com/Merovius/bor/sandbox"
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

func (d Driver) Command(name string, arg ...string) sandbox.Cmd {
	return Cmd{exec.Command(name, arg...)}
}

func init() {
	sandbox.Register("plain", Driver{})
}
