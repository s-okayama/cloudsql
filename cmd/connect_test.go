package cmd

import (
	"testing"
)

func TestFindExistingProxy_NoProxy(t *testing.T) {
	port := findExistingProxy("nonexistent-project:region:nonexistent-instance")
	if port != 0 {
		t.Errorf("expected 0 for no existing proxy, got %d", port)
	}
}

func TestUnixSocketPathFor_Postgres(t *testing.T) {
	got := unixSocketPathFor("POSTGRES_15", "/base", "proj:asia:inst")
	want := "/base/proj_asia_inst/.s.PGSQL.5432"
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

func TestUnixSocketPathFor_MySQL(t *testing.T) {
	got := unixSocketPathFor("MYSQL_8_0", "/base", "proj:asia:inst")
	want := "/base/proj_asia_inst"
	if got != want {
		t.Errorf("expected %s, got %s", want, got)
	}
}

func TestDefaultUnixSocketDir_XDGRuntimeDir(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")
	if got := defaultUnixSocketDir(); got != "/run/user/1000/cloudsql" {
		t.Errorf("expected /run/user/1000/cloudsql, got %s", got)
	}
}

func TestDefaultUnixSocketDir_FallbackToHome(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", "")
	t.Setenv("HOME", "/Users/testuser")
	if got := defaultUnixSocketDir(); got != "/Users/testuser/.cloudsql/sockets" {
		t.Errorf("expected /Users/testuser/.cloudsql/sockets, got %s", got)
	}
}
