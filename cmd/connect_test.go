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
