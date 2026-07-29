package cmd

import (
	"testing"
)

func TestUnixSocketPathFromEntry(t *testing.T) {
	tests := []struct {
		name  string
		entry string
		want  string
	}{
		{
			name:  "unix socket entry",
			entry: "proj:asia-northeast1:inst?unix-socket-path=/Users/me/.cloudsql/sockets/proj_asia-northeast1_inst --auto-iam-authn --private-ip --quiet",
			want:  "/Users/me/.cloudsql/sockets/proj_asia-northeast1_inst",
		},
		{
			name:  "unix socket entry with quotes",
			entry: "'proj:asia-northeast1:inst?unix-socket-path=/run/user/1000/cloudsql/proj_asia-northeast1_inst' --auto-iam-authn --quiet",
			want:  "/run/user/1000/cloudsql/proj_asia-northeast1_inst",
		},
		{
			name:  "tcp entry has no socket path",
			entry: "proj:asia-northeast1:inst --auto-iam-authn --private-ip --quiet --port=5432",
			want:  "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := unixSocketPathFromEntry(tt.entry); got != tt.want {
				t.Errorf("unixSocketPathFromEntry(%q) = %q, want %q", tt.entry, got, tt.want)
			}
		})
	}
}
