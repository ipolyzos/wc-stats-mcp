package exec

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	osexec "os/exec"
	"testing"
	"time"
)

func TestRunner_invalidName(t *testing.T) {
	r := NewRunner()
	_, _, err := r.Run("", []string{})
	k, ok := KindOf(err)
	if !ok || k != InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v (%v)", err, k)
	}
}

func TestRunner_nilArgs(t *testing.T) {
	r := NewRunner()
	_, _, err := r.Run("true", nil)
	k, ok := KindOf(err)
	if !ok || k != InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
	}
}

func TestRunner_notFound(t *testing.T) {
	r := &Runner{
		Timeout:        time.Second,
		MaxStdoutBytes: 1024,
		MaxStderrBytes: 1024,
		LookPath: func(string) (string, error) {
			return "", os.ErrNotExist
		},
	}
	_, _, err := r.Run("no-such-binary-xyz", []string{})
	k, ok := KindOf(err)
	if !ok || k != NotFound {
		t.Fatalf("expected NotFound, got %v", err)
	}
}

func TestRunner_timeout_fakeCommand(t *testing.T) {
	r := &Runner{
		Timeout:        5 * time.Millisecond,
		MaxStdoutBytes: 4096,
		MaxStderrBytes: 4096,
		LookPath:       func(string) (string, error) { return "/ignored", nil },
		RunCommand: func(ctx context.Context, path string, args []string, stdout, stderr io.Writer) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	_, _, err := r.Run("any", []string{})
	k, ok := KindOf(err)
	if !ok || k != Timeout {
		t.Fatalf("expected Timeout, got kind=%v err=%v", k, err)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected unwrap deadline: %v", err)
	}
}

func TestRunner_stdoutTooLarge_fakeCommand(t *testing.T) {
	r := &Runner{
		Timeout:        time.Second,
		MaxStdoutBytes: 100,
		MaxStderrBytes: 100,
		LookPath:       func(string) (string, error) { return "/ignored", nil },
		RunCommand: func(ctx context.Context, path string, args []string, stdout, stderr io.Writer) error {
			_, err := io.Copy(stdout, bytes.NewReader(bytes.Repeat([]byte("x"), 500)))
			return err
		},
	}
	out, _, err := r.Run("any", []string{})
	k, ok := KindOf(err)
	if !ok || k != TooLarge {
		t.Fatalf("expected TooLarge, got %v", err)
	}
	if len(out) != 100 {
		t.Fatalf("expected 100 bytes kept, got %d", len(out))
	}
}

func TestRunner_trueWhenAvailable(t *testing.T) {
	if _, err := osexec.LookPath("true"); err != nil {
		t.Skip("true not on PATH:", err)
	}
	r := NewRunner()
	r.Timeout = 5 * time.Second
	out, _, err := r.Run("true", []string{})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("unexpected stdout: %q", out)
	}
}

func TestRunner_falseWhenAvailable(t *testing.T) {
	if _, err := osexec.LookPath("false"); err != nil {
		t.Skip("false not on PATH:", err)
	}
	r := NewRunner()
	r.Timeout = 5 * time.Second
	_, _, err := r.Run("false", []string{})
	k, ok := KindOf(err)
	if !ok || k != ExecFailed {
		t.Fatalf("expected ExecFailed, got %v", err)
	}
}
