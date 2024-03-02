package ssh

import (
	"bytes"
	"github.com/cockroachdb/errors"
	"io"
	"reflect"
	"testing"
)

func TestReadSingleUntil(t *testing.T) {
	type Input struct {
		in  []byte
		sep []byte
	}
	type Result struct {
		out []byte
		err error
	}
	tests := []struct {
		in  Input
		out Result
	}{
		{in: Input{in: []byte("123\n456"), sep: []byte("\n")}, out: Result{out: []byte("123\n"), err: nil}},
		{in: Input{in: []byte("123\n456\n789"), sep: []byte("\n")}, out: Result{out: []byte("123\n"), err: nil}},
		{in: Input{in: []byte("123456"), sep: []byte("\n")}, out: Result{out: []byte("123456"), err: io.EOF}},
	}

	for _, tc := range tests {
		t.Run(string(tc.in.in), func(t *testing.T) {
			b, err := readSingleUntil(bytes.NewReader(tc.in.in), tc.in.sep)

			if !reflect.DeepEqual(b, tc.out.out) || !errors.Is(err, tc.out.err) {
				t.Fatalf(`readSingleUntil(%v, %v) = %v, %s, want %v, %s`, tc.in.in, tc.in.sep, b, err, tc.out.out, tc.out.err)
			}
		})
	}
}
