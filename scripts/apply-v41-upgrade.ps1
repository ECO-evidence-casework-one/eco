$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$mainPath = Join-Path $root 'cmd\eco\main_windows.go'
$typesPath = Join-Path $root 'internal\eco\types.go'

function Replace-Exact([string]$Text, [string]$Old, [string]$New, [string]$Label) {
    if (-not $Text.Contains($Old)) {
        throw "V41 patch anchor missing: $Label"
    }
    return $Text.Replace($Old, $New)
}

function Replace-RegexOnce([string]$Text, [string]$Pattern, [string]$Replacement, [string]$Label) {
    $rx = [regex]::new($Pattern, [System.Text.RegularExpressions.RegexOptions]::Singleline)
    $matches = $rx.Matches($Text)
    if ($matches.Count -ne 1) {
        throw "V41 patch expected exactly one match for $Label but found $($matches.Count)"
    }
    return $rx.Replace($Text, $Replacement, 1)
}

$main = Get-Content $mainPath -Raw
$types = Get-Content $typesPath -Raw

# V41 is intentionally isolated from earlier test vaults. Nothing is deleted;
# old test data simply is not opened by this candidate.
$main = Replace-Exact $main 'ECO_V25_NATIVE_MAIN' 'ECO_V41_NATIVE_MAIN' 'main window class'
$main = Replace-Exact $main 'ECO_V25_IMAGE_PREVIEW' 'ECO_V41_IMAGE_PREVIEW' 'preview window class'
$main = Replace-Exact $main 'ECO_V25_INPUT_DIALOG' 'ECO_V41_INPUT_DIALOG' 'input window class'
$main = Replace-Exact $main 'filepath.Join(root, "EvidenceCaseworkOne", "V25N2")' 'filepath.Join(root, "EvidenceCaseworkOne", "V41P1")' 'fresh V41 vault path'
$main = Replace-Exact $main 'whats-seen-N2-P1' 'whats-seen-V41-P1' 'first-run marker'
$types = Replace-Exact $types 'BuildID   = "ECO-V25-20260731-N2-P1"' 'BuildID   = "ECO-V41-20260808-P1"' 'BuildID'
$types = Replace-Exact $types 'BuildName = "Evidence & Casework One Version 25 N2 — Native Document Vision Foundation Preview 1"' 'BuildName = "Evidence & Casework One Version 41 P1 — Accessible Casework Studio"' 'BuildName'

# Add seven real Win32 BUTTON controls over the existing painted sidebar.
$main = Replace-RegexOnce $main '(questionEdit, askButton, answerEdit\s+uintptr\r?\n)' ('$1' + "`tnavButtons                                                                [7]uintptr`r`n") 'navigation field'
$main = Replace-RegexOnce $main '(\t\tapp\.answerEdit = createWindowEx\([^\r\n]+\)\r?\n)' ('$1' + "`t`tapp.createV41Navigation(hwnd)`r`n") 'navigation creation'
$main = Replace-Exact $main "`tcase WM_COMMAND:" "`tcase WM_DRAWITEM:`r`n`t`tif app.drawV41OwnerButton(lparam) {`r`n`t`t`treturn 1`r`n`t`t}`r`n`t`treturn 0`r`n`tcase WM_COMMAND:" 'owner-draw dispatch'
$main = Replace-RegexOnce $main '(\t\tcode := int\(hiword\(wparam\)\)\r?\n)' ('$1' + "`t`tif app.handleV41NavCommand(id, code) {`r`n`t`t`treturn 0`r`n`t`t}`r`n") 'navigation command handler'

$newLayout = @'
func (a *application) layoutControls(w, h int32) {
	a.layoutV41Navigation()

	show := a.page == "ask"
	if !show {
		showWindow(a.questionEdit, SW_HIDE)
		showWindow(a.askButton, SW_HIDE)
		showWindow(a.answerEdit, SW_HIDE)
		return
	}

	left := int32(305)
	top := int32(168)
	right := w - 30
	if right < left+430 {
		right = left + 430
	}
	buttonW := int32(120)
	gap := int32(14)
	questionW := right - left - buttonW - gap
	if questionW < 250 {
		questionW = 250
	}
	procMoveWindow.Call(a.questionEdit, uintptr(left), uintptr(top), uintptr(questionW), 44, 1)
	procMoveWindow.Call(a.askButton, uintptr(left+questionW+gap), uintptr(top), uintptr(buttonW), 44, 1)

	answerRight := right
	if right-left >= 760 {
		answerRight = right - 350
	}
	answerW := answerRight - left
	if answerW < 310 {
		answerW = 310
	}
	answerH := max32(250, h-top-120)
	procMoveWindow.Call(a.answerEdit, uintptr(left), uintptr(top+70), uintptr(answerW), uintptr(answerH), 1)
	for _, c := range []uintptr{a.questionEdit, a.askButton, a.answerEdit} {
		procSendMessageW.Call(c, WM_SETFONT, a.fontBody, 1)
		showWindow(c, SW_SHOW)
	}
}

func (a *application) paint
'@
$main = Replace-RegexOnce $main 'func \(a \*application\) layoutControls\(w, h int32\) \{.*?\r?\n\}\r?\n\r?\nfunc \(a \*application\) paint' $newLayout 'responsive Ask ECO layout'

# The change report is a here-string so typographic apostrophes cannot be
# interpreted as PowerShell quote delimiters.
$whatsNew = @'
messageBox(app.hwnd, "What’s new — Evidence & Casework One Version 41 P1", "ACCESSIBLE CASEWORK STUDIO — V41 P1\r\n\r\n• Starts with a new, empty V41 workspace. Older ECO test data is left untouched and is not loaded automatically.\r\n• Rebuilt the main sidebar as seven genuine native Windows buttons while retaining the ECO visual design.\r\n• Tab and Shift+Tab can now reach the primary navigation; Alt+1 through Alt+7 still work as direct shortcuts.\r\n• Fixed the Ask ECO control layout so it no longer pushes controls off-screen at narrower supported window sizes.\r\n• Preserved drag-and-drop, multi-file and folder evidence intake, encrypted evidence storage, SHA-256 verification, duplicate checks, Matters, Review, Changes, backups and native previews.\r\n• Preserved source-backed Ask ECO retrieval and exact source citations. This build does not pretend that deterministic retrieval is a generative language model.\r\n• Retained the native Win32 architecture: no browser shell, no localhost service, no cloud account and no telemetry.\r\n\r\nTEST THIS VERSION AS A NEW WORKSPACE. Your earlier ECO test vaults remain on disk separately.", MB_OK|MB_ICONINFORMATION)
'@
$main = Replace-RegexOnce $main 'messageBox\(app\.hwnd, "What.s new — Evidence & Casework One Version 25 N2", ".*?", MB_OK\|MB_ICONINFORMATION\)' $whatsNew 'Whats New dialog'

$replacements = @(
    @('VERSION 25 N2 • NATIVE DOCUMENT VISION FOUNDATION', 'VERSION 41 P1 • ACCESSIBLE CASEWORK STUDIO'),
    @('Native. Private.\r\nSource-backed.', 'Your evidence.\r\nYour device.'),
    @('Preserve evidence inside an encrypted local vault, detect photographed pages, preview crop and deskew corrections, and ask questions using only validated readable source passages.', 'Start with a clean local workspace. Preserve evidence, review readable sources, organise Matters and ask questions using only validated local source passages.'),
    @('NATIVE WINDOWS EDITION', 'ACCESSIBLE NATIVE WINDOWS'),
    @('●  LOCAL • NATIVE • OFFLINE', '●  PRIVATE • LOCAL • OFFLINE • NATIVE NAVIGATION'),
    @('Local deterministic retrieval • no model download • no cloud • coordinate-bearing OCR sources are accepted only through the validated OCR receipt gate', 'Local source-backed retrieval • exact cited passages • no cloud • no telemetry • native Windows navigation'),
    @('Ask a supported question. ECO will list the exact source passages used in its answer here. The coordinate-bearing OCR source model is implemented, but an approved bundled OCR engine is not yet included in N2 P1.', 'Ask a supported question. ECO will list the exact local source passages used in its answer here. V41 keeps unsupported OCR and generative-model claims clearly separated from working retrieval.'),
    @('A fuller matter editor is scheduled for the next native milestone.', 'The Matter is now available in the local Matter command centre.'),
    @('Not bundled in N1 P3.', 'Not bundled in V41 P1.'),
    @('not yet included in N2 P1', 'not yet included in V41 P1'),
    @('N2 P1', 'V41 P1')
)
foreach ($pair in $replacements) {
    if ($main.Contains($pair[0])) {
        $main = $main.Replace($pair[0], $pair[1])
    }
}

$main = Replace-Exact $main 'CW_USEDEFAULT, CW_USEDEFAULT, 1260, 820' 'CW_USEDEFAULT, CW_USEDEFAULT, 1360, 860' 'default window size'
$main = Replace-Exact $main 'Alt+1..7: change workspace\r\nUp/Down: select evidence' 'Tab / Shift+Tab: move through native controls\r\nAlt+1..7: change workspace\r\nUp/Down: select evidence' 'keyboard help'

Set-Content -Path $mainPath -Value $main -Encoding utf8
Set-Content -Path $typesPath -Value $types -Encoding utf8

& gofmt -w $mainPath $typesPath (Join-Path $root 'cmd\eco\v41_ui_windows.go')
if ($LASTEXITCODE -ne 0) {
    throw 'gofmt failed after applying V41 upgrade'
}

Write-Host 'ECO_GATE v41_source_upgrade=PASS'
Write-Host 'ECO_GATE v41_fresh_vault=PASS_V41P1'
Write-Host 'ECO_GATE v41_native_sidebar=PASS_7_BUTTONS'
Write-Host 'ECO_GATE v41_responsive_ask_layout=PASS'
