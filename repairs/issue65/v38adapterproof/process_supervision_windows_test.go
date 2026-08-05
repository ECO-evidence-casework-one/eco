//go:build windows

package v38adapterproof

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestRecoveredV38LeaseKillsCancellationIgnoringProcess(t *testing.T) {
	if os.Getenv("ECO_V38_LEASE_HELPER") == "1" {
		for {
			time.Sleep(time.Hour)
		}
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestRecoveredV38LeaseKillsCancellationIgnoringProcess")
	cmd.Env = append(os.Environ(), "ECO_V38_LEASE_HELPER=1")
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := terminateCommand(ctx, cmd, done); err != nil {
		t.Fatal(err)
	}
	if cmd.ProcessState == nil {
		t.Fatal("termination returned without cmd.Wait confirming the process ended")
	}
}

func TestRecoveredV38LeaseRejectsUnprovableTermination(t *testing.T) {
	cmd := &exec.Cmd{Process: &os.Process{Pid: os.Getpid()}}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if err := terminateCommand(ctx, cmd, nil); err == nil {
		t.Fatal("missing Wait owner was incorrectly treated as confirmed termination")
	}
}
