package tsdecls

import (
	"bytes"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTsdecls(t *testing.T) {
	want, err := os.ReadFile("_testdata/want.ts")
	if err != nil {
		t.Fatal(err)
	}
	got := new(bytes.Buffer)
	if err = Write(got, "_testdata", "Server"); err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(string(want), got.String()); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}
