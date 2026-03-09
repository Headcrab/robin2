package robin

import (
	"errors"
	"net"
	"os"
	"syscall"
	"testing"
)

func TestIsClientDisconnect(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "broken pipe syscall",
			err:  &net.OpError{Err: &os.SyscallError{Syscall: "write", Err: syscall.EPIPE}},
			want: true,
		},
		{
			name: "connection reset syscall",
			err:  &net.OpError{Err: &os.SyscallError{Syscall: "write", Err: syscall.ECONNRESET}},
			want: true,
		},
		{
			name: "plain broken pipe text",
			err:  errors.New("write tcp [::1]:8008->[::1]:49688: write: broken pipe"),
			want: true,
		},
		{
			name: "plain reset text",
			err:  errors.New("write tcp [::1]:8008->[::1]:56588: write: connection reset by peer"),
			want: true,
		},
		{
			name: "other error",
			err:  errors.New("something else"),
			want: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if got := isClientDisconnect(tc.err); got != tc.want {
				t.Fatalf("isClientDisconnect() = %v, want %v", got, tc.want)
			}
		})
	}
}
