from pathlib import Path


def replace_once(path: str, old: str, new: str) -> None:
    p = Path(path)
    text = p.read_text(encoding="utf-8")
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"expected exactly one anchor in {path}, found {count}")
    p.write_text(text.replace(old, new, 1), encoding="utf-8")


replace_once(
    "cmd/eco/winapi_windows.go",
    '''\tMB_YESNO             = 0x00000004\n\tMB_ICONQUESTION      = 0x00000020\n\tIDYES                = 6\n''',
    '''\tMB_YESNO             = 0x00000004\n\tMB_YESNOCANCEL       = 0x00000003\n\tMB_ICONQUESTION      = 0x00000020\n\tIDYES                = 6\n\tIDNO                 = 7\n\tIDCANCEL             = 2\n''',
)

replace_once(
    "cmd/eco/main_windows.go",
    '''\troot := os.Getenv("LOCALAPPDATA")\n\tif root == "" {\n\t\troot, _ = os.UserConfigDir()\n\t}\n\troot = filepath.Join(root, "EvidenceCaseworkOne", "V25N2")\n\tv, err := eco.OpenVault(root)\n''',
    '''\tbase := os.Getenv("LOCALAPPDATA")\n\tif base == "" {\n\t\tbase, _ = os.UserConfigDir()\n\t}\n\troot, proceed := chooseDevelopmentWorkspace(base)\n\tif !proceed {\n\t\treturn\n\t}\n\tv, err := eco.OpenVault(root)\n''',
)

replace_once(
    "cmd/eco/main_windows.go",
    '''\tif err != nil {\n\t\tmessageBox(0, "ECO vault could not open", err.Error(), MB_OK|MB_ICONERROR)\n\t\treturn\n\t}\n\tinitialView := v.Snapshot()\n''',
    '''\tif err != nil {\n\t\tmessageBox(0, "ECO vault could not open", err.Error(), MB_OK|MB_ICONERROR)\n\t\treturn\n\t}\n\tdefer func() { _ = v.Close() }()\n\tinitialView := v.Snapshot()\n''',
)

insert_anchor = '''}\n\nfunc registerClasses() {\n'''
startup_helpers = r'''}

func chooseDevelopmentWorkspace(base string) (string, bool) {
	candidate := eco.DefaultDevelopmentWorkspaceRoot(base)
	candidateInfo, candidateErr := os.Stat(candidate)
	if candidateErr != nil && !os.IsNotExist(candidateErr) {
		messageBox(0, "ECO workspace check failed", candidateErr.Error(), MB_OK|MB_ICONERROR)
		return "", false
	}
	if os.IsNotExist(candidateErr) {
		legacy := filepath.Join(base, "EvidenceCaseworkOne", "V25N2")
		if eco.ValidateExistingWorkspaceRoot(legacy) == nil {
			choice := messageBox(0, "Choose ECO development workspace", "This exact source candidate has no workspace yet, so its normal start is clean.\r\n\r\nA previous V25N2 development workspace also exists.\r\n\r\nYes — Start this candidate clean\r\nNo — Open an existing ECO workspace\r\nCancel — Exit without opening anything", MB_YESNOCANCEL|MB_ICONQUESTION)
			switch choice {
			case IDYES:
				return candidate, true
			case IDNO:
				return chooseExistingWorkspace()
			default:
				return "", false
			}
		}
		return candidate, true
	}
	if candidateInfo == nil || !candidateInfo.IsDir() {
		messageBox(0, "ECO candidate workspace is invalid", "The candidate workspace route exists but is not a directory. ECO will not replace or alter it.", MB_OK|MB_ICONERROR)
		return "", false
	}

	for {
		continueOK := eco.ValidateExistingWorkspaceRoot(candidate) == nil
		if continueOK {
			choice := messageBox(0, "Choose ECO development workspace", "This exact source candidate already has a development workspace.\r\n\r\nYes — Continue this candidate\r\nNo — Open existing / Start clean\r\nCancel — Exit without opening anything", MB_YESNOCANCEL|MB_ICONQUESTION)
			switch choice {
			case IDYES:
				return candidate, true
			case IDCANCEL:
				return "", false
			}
		} else {
			choice := messageBox(0, "Candidate workspace needs a decision", "This candidate route exists but is not a complete ECO workspace. ECO will not silently repair or overwrite it.\r\n\r\nYes — Open another existing ECO workspace\r\nNo — Preserve this incomplete route and start clean\r\nCancel — Exit", MB_YESNOCANCEL|MB_ICONQUESTION)
			switch choice {
			case IDYES:
				return chooseExistingWorkspace()
			case IDNO:
				return startCleanCandidate(candidate)
			default:
				return "", false
			}
		}

		choice := messageBox(0, "Open existing or start clean", "Choose what ECO should do with development state.\r\n\r\nYes — Open an existing ECO workspace\r\nNo — Preserve this candidate workspace and start clean\r\nCancel — Go back", MB_YESNOCANCEL|MB_ICONQUESTION)
		switch choice {
		case IDYES:
			root, ok := chooseExistingWorkspace()
			if ok {
				return root, true
			}
		case IDNO:
			return startCleanCandidate(candidate)
		case IDCANCEL:
			continue
		}
	}
}

func chooseExistingWorkspace() (string, bool) {
	for {
		root := openFolderDialogWithTitle(0, "Open an existing ECO workspace — the selected folder will be checked before it is opened")
		if root == "" {
			return "", false
		}
		if err := eco.ValidateExistingWorkspaceRoot(root); err != nil {
			messageBox(0, "That folder is not an existing ECO workspace", err.Error()+"\r\n\r\nNothing was created or changed in the selected folder.", MB_OK|MB_ICONERROR)
			continue
		}
		return root, true
	}
}

func startCleanCandidate(candidate string) (string, bool) {
	archive, err := eco.ArchiveDevelopmentWorkspaceForCleanStart(candidate)
	if err != nil {
		messageBox(0, "ECO could not start clean", err.Error(), MB_OK|MB_ICONERROR)
		return "", false
	}
	if archive != "" {
		messageBox(0, "Prior candidate workspace preserved", "ECO did not delete the previous workspace. It was preserved at:\r\n\r\n"+archive+"\r\n\r\nA fresh candidate workspace will now be created.", MB_OK|MB_ICONINFORMATION)
	}
	return candidate, true
}

func registerClasses() {
'''
replace_once("cmd/eco/main_windows.go", insert_anchor, startup_helpers)

replace_once(
    "cmd/eco/ui_regression_test.go",
    '''func TestNativeSourceContainsNoBrowserOrLocalhostRuntime(t *testing.T) {\n''',
    '''func TestDevelopmentStartupStateIsExplicitAndCandidateBound(t *testing.T) {\n\tsrc := windowsSource(t)\n\tfor _, required := range []string{"chooseDevelopmentWorkspace", "DefaultDevelopmentWorkspaceRoot", "ValidateExistingWorkspaceRoot", "ArchiveDevelopmentWorkspaceForCleanStart", "Continue this candidate", "Open an existing ECO workspace", "start clean", "defer func() { _ = v.Close() }()"} {\n\t\tif !strings.Contains(src, required) {\n\t\t\tt.Fatalf("missing explicit startup-state control %q", required)\n\t\t}\n\t}\n\tif strings.Contains(src, `root = filepath.Join(root, "EvidenceCaseworkOne", "V25N2")`) {\n\t\tt.Fatal("Windows startup still silently selects the old shared V25N2 workspace")\n\t}\n}\n\nfunc TestNativeSourceContainsNoBrowserOrLocalhostRuntime(t *testing.T) {\n''',
)

replace_once(
    "cmd/eco/main_other.go",
    '''\troot := filepath.Join(os.TempDir(), "eco-v25-n1-dev-vault")\n\tv, err := eco.OpenVault(root)\n''',
    '''\troot := eco.DefaultDevelopmentWorkspaceRoot(os.TempDir())\n\tv, err := eco.OpenVault(root)\n''',
)
replace_once(
    "cmd/eco/main_other.go",
    '''\tif err != nil {\n\t\tpanic(err)\n\t}\n\tfmt.Println(eco.BuildName)\n''',
    '''\tif err != nil {\n\t\tpanic(err)\n\t}\n\tdefer v.Close()\n\tfmt.Println(eco.BuildName)\n''',
)
