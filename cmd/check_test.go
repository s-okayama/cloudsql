package cmd

import (
	"net"
	"testing"
)

func TestFindAvailablePort_Free(t *testing.T) {
	port := findAvailablePort(19876)
	if port != 19876 {
		t.Errorf("expected 19876, got %d", port)
	}
}

func TestFindAvailablePort_Occupied(t *testing.T) {
	ln, err := net.Listen("tcp4", "127.0.0.1:19877")
	if err != nil {
		t.Fatalf("failed to occupy port: %v", err)
	}
	defer ln.Close()

	port := findAvailablePort(19877)
	if port == 19877 {
		t.Errorf("expected a different port, got 19877")
	}
	if port < 19878 || port > 19977 {
		t.Errorf("expected port in range 19878-19977, got %d", port)
	}
}

func TestFindAvailablePort_SkipsMultiple(t *testing.T) {
	ln1, err := net.Listen("tcp4", "127.0.0.1:19880")
	if err != nil {
		t.Fatalf("failed to occupy port: %v", err)
	}
	defer ln1.Close()

	ln2, err := net.Listen("tcp4", "127.0.0.1:19881")
	if err != nil {
		t.Fatalf("failed to occupy port: %v", err)
	}
	defer ln2.Close()

	port := findAvailablePort(19880)
	if port != 19882 {
		t.Errorf("expected 19882, got %d", port)
	}
}
