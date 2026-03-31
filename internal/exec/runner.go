package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	osexec "os/exec"
	"time"
)

const (
	DefaultTimeout        = 2 * time.Second
	DefaultMaxStreamBytes = 1 << 20 // 1 MiB per stream
)

// CommandFunc runs a resolved binary with explicit argv. Stdin is not used (nil).
type CommandFunc func(ctx context.Context, path string, args []string, stdout, stderr io.Writer) error

// Runner executes programs without a shell, with deadlines and capped streams.
type Runner struct {
	Timeout        time.Duration
	MaxStdoutBytes int64
	MaxStderrBytes int64
	LookPath       func(string) (string, error)
	// RunCommand replaces os/exec when set (tests).
	RunCommand CommandFunc
}

// NewRunner returns a runner with library defaults.
func NewRunner() *Runner {
	return &Runner{
		Timeout:        DefaultTimeout,
		MaxStdoutBytes: DefaultMaxStreamBytes,
		MaxStderrBytes: DefaultMaxStreamBytes,
		LookPath:       osexec.LookPath,
	}
}

func (r *Runner) timeout() time.Duration {
	if r == nil || r.Timeout <= 0 {
		return DefaultTimeout
	}
	return r.Timeout
}

func (r *Runner) capStdout() int64 {
	if r == nil || r.MaxStdoutBytes <= 0 {
		return DefaultMaxStreamBytes
	}
	return r.MaxStdoutBytes
}

func (r *Runner) capStderr() int64 {
	if r == nil || r.MaxStderrBytes <= 0 {
		return DefaultMaxStreamBytes
	}
	return r.MaxStderrBytes
}

// Run resolves name on PATH, runs [path] + args with exec.CommandContext, and captures streams.
// args must be non-nil (empty slice allowed). No shell is involved.
func (r *Runner) Run(name string, args []string) (stdout, stderr []byte, err error) {
	if name == "" {
		return nil, nil, &Error{Kind: InvalidArgument, Message: "executable name is empty"}
	}
	if args == nil {
		return nil, nil, &Error{Kind: InvalidArgument, Message: "args must be non-nil (use empty slice)"}
	}
	if r == nil {
		r = NewRunner()
	}
	lp := r.LookPath
	if lp == nil {
		lp = osexec.LookPath
	}
	path, lerr := lp(name)
	if lerr != nil {
		return nil, nil, &Error{
			Kind:    NotFound,
			Message: fmt.Sprintf("resolve %q: %v", name, lerr),
			Cause:   lerr,
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout())
	defer cancel()

	out := &CappedWriter{Stream: "stdout", Limit: r.capStdout()}
	errBuf := &CappedWriter{Stream: "stderr", Limit: r.capStderr()}

	run := r.RunCommand
	if run == nil {
		run = defaultRunCommand
	}
	runErr := run(ctx, path, args, out, errBuf)
	stdout, stderr = out.Bytes(), errBuf.Bytes()

	if runErr != nil {
		if errors.Is(runErr, context.DeadlineExceeded) || ctx.Err() == context.DeadlineExceeded {
			return stdout, stderr, &Error{
				Kind:    Timeout,
				Message: fmt.Sprintf("deadline exceeded (%v)", r.timeout()),
				Cause:   runErr,
			}
		}
		var te *Error
		if errors.As(runErr, &te) && te.Kind == TooLarge {
			return stdout, stderr, te
		}
		var exit *osexec.ExitError
		if errors.As(runErr, &exit) {
			return stdout, stderr, &Error{
				Kind:    ExecFailed,
				Message: fmt.Sprintf("exit code %d", exit.ExitCode()),
				Cause:   runErr,
			}
		}
		return stdout, stderr, &Error{
			Kind:    ExecFailed,
			Message: runErr.Error(),
			Cause:   runErr,
		}
	}
	return stdout, stderr, nil
}

func defaultRunCommand(ctx context.Context, path string, args []string, stdout, stderr io.Writer) error {
	cmd := osexec.CommandContext(ctx, path, args...)
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
