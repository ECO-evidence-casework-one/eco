package eco

// Test-only fixture convenience retaining prior create-or-open setup behaviour.
// Application code cannot invoke this helper. Strict-open contract tests do not use it.
func openTestVault(root string) (*Vault, error) {
	if err := CheckWorkspaceRecoveryState(root); err != nil {
		return nil, err
	}
	if err := preflightWorkspaceFormat(root); err != nil {
		return nil, err
	}
	owner, err := acquireOrCreateWorkspaceRootOwner(root)
	if err != nil {
		return nil, err
	}
	return openOwnedVault(owner, true)
}
