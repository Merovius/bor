// This package implements the sandbox.Driver interface, injecting SECCOMPv1-mode
// into run processes by means of the EasySandbox library. This means, that the
// process can do no syscall but read(), write() (on stdin/stdout/stderr),
// exit() and sigreturn(), effectively minimizing the effect it can have on the
// rest of the system.
//
// Therefore this package is unfit for running processes that are
// multithreaded, need to read from files, open network connection or do
// anything else requiring more than those basic syscalls.
//
// The EasySandbox library outputs the magic string "<<entering SECCOMP
// mode>>\n" on stdout/stderr, for technical reasons. These are automatically
// filtered out. For the same technical reason, processes doing unbuffered IO
// with read() on stdin will possibly not be able to read the first byte input.
package easysandbox

import (
	"bytes"
	goconf "code.google.com/p/goconf/conf"
	"fmt"
	"github.com/Merovius/bor/sandbox"
	"io"
	"os/exec"
)

var (
	magic = []byte("<<entering SECCOMP mode>>\n")
	path  = "/usr/lib/EasySandbox/EasySandbox.so"
	heap  = 8388608
)

// EasySandboxError represents an error with the EasySandbox-library (signified
// by missing the magic string output by EasySandbox on either stderr or stdin)
type EasySandboxError string

// Implement the builtin error interface
func (err EasySandboxError) Error() string {
	return string(err)
}

// OffsetReadCloser wraps an io.ReadCloser throwing away a number of bytes at
// the start
type OffsetReadCloser struct {
	r io.ReadCloser
	i int
	n int
}

// NewOffsetReadCloser returns an OffsetReadCloser wrapping r and throwing away
// the first n bytes
func NewOffsetReadCloser(r io.ReadCloser, n int) OffsetReadCloser {
	return OffsetReadCloser{r, 0, n}
}

// Implement the io.Reader interface
func (r OffsetReadCloser) Read(p []byte) (n int, err error) {
	if r.n == r.i {
		return r.r.Read(p)
	}
	buf := make([]byte, r.n-r.i)

	m, err := r.r.Read(buf)
	if m == r.n-r.i {
		r.i = r.n
		return r.r.Read(p)
	}

	r.i += m
	return 0, err
}

// Implement the io.Closer interface
func (r OffsetReadCloser) Close() error {
	return r.r.Close()
}

// Driver implements the sandbox-interface
type Driver struct{}

type Cmd struct {
	*exec.Cmd
}

func (c Cmd) CombinedOutput() ([]byte, error) {
	out, err := c.Cmd.CombinedOutput()

	if !bytes.HasPrefix(out, magic) {
		return out, EasySandboxError("Magic not found")
	}
	out = out[len(magic):]

	if !bytes.HasPrefix(out, magic) {
		return out, EasySandboxError("Magic found only once")
	}
	out = out[len(magic):]

	return out, err
}

func (c Cmd) Output() ([]byte, error) {
	out, err := c.Cmd.Output()
	if err != nil {
		return out, err
	}

	if !bytes.HasPrefix(out, magic) {
		return out, EasySandboxError("Magic not found")
	}
	out = out[len(magic):]

	return out, nil
}

func (c Cmd) StderrPipe() (io.ReadCloser, error) {
	pipe, err := c.Cmd.StderrPipe()
	r := NewOffsetReadCloser(pipe, len(magic))
	return r, err
}

func (c Cmd) StdoutPipe() (io.ReadCloser, error) {
	pipe, err := c.Cmd.StdoutPipe()
	r := NewOffsetReadCloser(pipe, len(magic))
	return r, err
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
	ret := Cmd{exec.Command(name, arg...)}
	ret.Cmd.Env = []string{
		"LD_PRELOAD=" + path,
		fmt.Sprintf("EASYSANDBOX_HEAPSIZE=%d", heap),
	}

	return ret
}

func (d Driver) Config(cfg *goconf.ConfigFile) error {
	if str, err := cfg.GetString("easysandbox", "Location"); err == nil {
		path = str
	} else {
		return fmt.Errorf("No Location for EasySandbox.so given")
	}

	if num, err := cfg.GetInt("easysandbox", "HeapSize"); err == nil {
		heap = num
	}

	return nil
}

func init() {
	sandbox.Register("easysandbox", Driver{})
}
