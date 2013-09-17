// Package easysandbox implements the sandbox.Driver interface, injecting
// SECCOMPv1-mode into run processes by means of the EasySandbox library. This
// means, that the process can do no syscall but read(), write() (on
// stdin/stdout/stderr), exit() and sigreturn(), effectively minimizing the
// effect it can have on the rest of the system.
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
	id    = 0
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
func NewOffsetReadCloser(r io.ReadCloser, n int) *OffsetReadCloser {
	return &OffsetReadCloser{r, 0, n}
}

// Read implements the io.Reader interface
func (r *OffsetReadCloser) Read(p []byte) (n int, err error) {
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

// Close implements the io.Closer interface
func (r *OffsetReadCloser) Close() error {
	return r.r.Close()
}

// OffsetWriter wraps an io.Writer, throwing away a number of bytes at the start
type OffsetWriter struct {
	w io.Writer
	i int
	n int
}

// NewOffsetWriter returns an OffsetWriter wrapping w and throwing away the
// first n bytes
func NewOffsetWriter(w io.Writer, n int) *OffsetWriter {
	return &OffsetWriter{w, 0, n}
}

// Write implements the io.Writer interface
func (w *OffsetWriter) Write(p []byte) (n int, err error) {
	if w.n == w.i {
		return w.w.Write(p)
	}

	todo := w.n - w.i

	if len(p) <= todo {
		w.i += len(p)
		return len(p), nil
	}

	m, err := w.w.Write(p[todo:])
	w.i += m

	return m + todo, nil
}

// Driver implements the sandbox-interface
type Driver struct{}

// Cmd wraps a *exec.Cmd in delicious sandboxing
type Cmd struct {
	*exec.Cmd
}

// CombinedOutput returns the combined stderr and stdout of the command,
// stripping away the magic of EasySandbox
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

// Output returns the stdout of the command, stripping away the magic of
// EasySandbox
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

// StderrPipe returns a pipe connected to stderr and set up to throw away the
// magic of EasySandbox
func (c Cmd) StderrPipe() (io.ReadCloser, error) {
	pipe, err := c.Cmd.StderrPipe()
	r := NewOffsetReadCloser(pipe, len(magic))
	return r, err
}

// StdoutPipe returns a pipe connected to stdout and set up to throw away the
// magic of EasySandbox
func (c Cmd) StdoutPipe() (io.ReadCloser, error) {
	pipe, err := c.Cmd.StdoutPipe()
	r := NewOffsetReadCloser(pipe, len(magic))
	return r, err
}

// Dir returns the current working directory of the command
func (c Cmd) Dir() string {
	return c.Cmd.Dir
}

// SetDir sets the working directory of the command
func (c Cmd) SetDir(dir string) {
	c.Cmd.Dir = dir
}

// ProcessState returns the *os.ProcessState of the underlying *exec.Cmd
func (c Cmd) ProcessState() sandbox.ProcessState {
	return c.Cmd.ProcessState
}

// Kill sends a SIGKILL to the underlying *os.Process
func (c Cmd) Kill() error {
	return c.Cmd.Process.Kill()
}

// SetStdout sets the stdout to the given writer, throwing away the EasySandbox
// magic
func (c Cmd) SetStdout(w io.Writer) {
	stdout := NewOffsetWriter(w, len(magic))
	c.Cmd.Stdout = stdout
}

// SetStdout sets the stderr to the given writer, throwing away the EasySandbox
// magic
func (c Cmd) SetStderr(w io.Writer) {
	stderr := NewOffsetWriter(w, len(magic))
	c.Cmd.Stderr = stderr
}

// Command returns a new command-struct, with the necessary environment for
// EasySandbox
func (d Driver) Command(name string, arg ...string) sandbox.Cmd {
	ret := Cmd{exec.Command(name, arg...)}
	ret.Cmd.Env = []string{
		"LD_PRELOAD=" + path,
		fmt.Sprintf("EASYSANDBOX_HEAPSIZE=%d", heap),
	}

	return ret
}

// Config configures the easysandbox package. It extracts the location of the
// shared object and the wanted heapsize
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

// init registers the easysandbox driver
func init() {
	sandbox.Register("easysandbox", Driver{})
}
