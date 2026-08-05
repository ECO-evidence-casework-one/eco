package v38adapterproof

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
)

// terminateCommand is the exact production primitive used by the recovered
// V38 worker lease. A successful return proves that Process.Kill completed and
// that the goroutine owning cmd.Wait observed process termination.
func terminateCommand(ctx context.Context, cmd *exec.Cmd, done <-chan error) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	if done == nil {
		return errors.New("worker completion channel is unavailable")
	}
	killErr := cmd.Process.Kill()
	if killErr != nil && !errors.Is(killErr, os.ErrProcessDone) {
		return fmt.Errorf("kill worker process: %w", killErr)
	}
	select {
	case <-done:
		if cmd.ProcessState == nil {
			return errors.New("worker process termination was not confirmed by Wait")
		}
		return nil
	case <-ctx.Done():
		return fmt.Errorf("confirm worker process termination: %w", ctx.Err())
	}
}
