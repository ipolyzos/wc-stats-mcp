package version

import "testing"

func TestString_nonEmpty(t *testing.T) {
	if String() == "" {
		t.Fatal("expected non-empty version string")
	}
}
