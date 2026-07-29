package cmd

import (
	"testing"
	"time"
)

func TestSuperviseProcess_KillsChildAfterTTL(t *testing.T) {
	start := time.Now()
	superviseProcess(200*time.Millisecond, []string{"sleep", "30"})
	elapsed := time.Since(start)
	if elapsed < 200*time.Millisecond {
		t.Errorf("returned before ttl elapsed: %v", elapsed)
	}
	if elapsed > 5*time.Second {
		t.Errorf("child was not killed after ttl: %v", elapsed)
	}
}

func TestSuperviseProcess_ExitsWhenChildExits(t *testing.T) {
	start := time.Now()
	superviseProcess(30*time.Second, []string{"true"})
	if elapsed := time.Since(start); elapsed > 2*time.Second {
		t.Errorf("did not return promptly after child exit: %v", elapsed)
	}
}
