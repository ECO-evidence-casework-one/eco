package eco

import (
	"bufio"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestWorkspaceOwnerBlocksSecondAcquisition(t *testing.T) {
	root := t.TempDir()
	first, err := acquireWorkspaceRootOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := acquireWorkspaceRootOwner(root)
	if second != nil {
		_ = second.Close()
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		t.Fatalf("second acquisition error=%v", err)
	}
}

func TestWorkspaceOwnerCloseAllowsReacquisition(t *testing.T) {
	root := t.TempDir()
	first, err := acquireWorkspaceRootOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := first.Close(); err != nil {
		t.Fatal(err)
	}
	second, err := acquireWorkspaceRootOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := second.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestWorkspaceOwnerBlocksOtherProcess(t *testing.T) {
	if os.Getenv("ECO_WORKSPACE_OWNER_HELPER") == "1" {
		t.Skip("parent-only test")
	}
	root := t.TempDir()
	cmd := exec.Command(os.Args[0], "-test.run=TestWorkspaceOwnerHelperProcess")
	cmd.Env = append(os.Environ(), "ECO_WORKSPACE_OWNER_HELPER=1", "ECO_WORKSPACE_OWNER_ROOT="+root)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(stdout)
	if !scanner.Scan() || scanner.Text() != "READY" {
		_ = cmd.Process.Kill()
		t.Fatalf("helper did not become ready: %q err=%v", scanner.Text(), scanner.Err())
	}
	lease, err := acquireWorkspaceRootOwner(root)
	if lease != nil {
		_ = lease.Close()
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		_ = stdin.Close()
		_ = cmd.Wait()
		t.Fatalf("cross-process acquisition error=%v", err)
	}
	_ = stdin.Close()
	if err := cmd.Wait(); err != nil {
		t.Fatal(err)
	}
	lease, err = acquireWorkspaceRootOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	_ = lease.Close()
}

func TestWorkspaceOwnerHelperProcess(t *testing.T) {
	if os.Getenv("ECO_WORKSPACE_OWNER_HELPER") != "1" {
		return
	}
	lease, err := acquireWorkspaceRootOwner(os.Getenv("ECO_WORKSPACE_OWNER_ROOT"))
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	_, _ = os.Stdout.WriteString("READY\n")
	_, _ = bufio.NewReader(os.Stdin).ReadByte()
}

func TestWorkspaceOwnerLockFileIsInsideRoot(t *testing.T) {
	root := t.TempDir()
	lease, err := acquireWorkspaceRootOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	if _, err := os.Stat(filepath.Join(root, workspaceOwnerFilename)); err != nil {
		t.Fatal(err)
	}
	if err := lease.revalidate(); err != nil {
		t.Fatal(err)
	}
}

func TestAcquireOrCreateWorkspaceRootOwnerCreatesRootUnderClaim(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "new-workspace")
	lease, err := acquireOrCreateWorkspaceRootOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		t.Fatalf("root not created safely: info=%v err=%v", info, err)
	}
	if err := lease.revalidate(); err != nil {
		t.Fatal(err)
	}
}

func TestConcurrentWorkspaceCreationHasOneWritableOwner(t *testing.T) {
	parent := t.TempDir()
	root := filepath.Join(parent, "new-workspace")
	first, err := acquireOrCreateWorkspaceRootOwner(root)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	second, err := acquireOrCreateWorkspaceRootOwner(root)
	if second != nil {
		_ = second.Close()
	}
	if !errors.Is(err, ErrWorkspaceInUse) {
		t.Fatalf("concurrent create/open error=%v", err)
	}
}
