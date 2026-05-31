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
