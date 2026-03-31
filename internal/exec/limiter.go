package exec

import (
	"bytes"
	"fmt"
)

// CappedWriter accumulates bytes into Buf up to Limit (total). Further input returns *Error with Kind TooLarge.
type CappedWriter struct {
	Stream string
	Limit  int64
	Buf    bytes.Buffer
	n      int64
}

// Bytes returns captured output (possibly truncated when an overflow error was returned).
func (w *CappedWriter) Bytes() []byte {
	if w == nil {
		return nil
	}
	return w.Buf.Bytes()
}

// Write implements io.Writer.
func (w *CappedWriter) Write(p []byte) (int, error) {
	if w.Limit <= 0 {
		return w.Buf.Write(p)
	}
	remain := w.Limit - w.n
	if remain <= 0 {
		return 0, &Error{
			Kind:    TooLarge,
			Message: fmt.Sprintf("%s exceeded byte cap (%d)", w.Stream, w.Limit),
		}
	}
	take := p
	if int64(len(take)) > remain {
		take = p[:remain]
	}
	nw, err := w.Buf.Write(take)
	w.n += int64(nw)
	if err != nil {
		return nw, err
	}
	if int64(len(p)) > int64(len(take)) {
		return len(take), &Error{
			Kind:    TooLarge,
			Message: fmt.Sprintf("%s exceeded byte cap (%d)", w.Stream, w.Limit),
		}
	}
	return len(p), nil
}
