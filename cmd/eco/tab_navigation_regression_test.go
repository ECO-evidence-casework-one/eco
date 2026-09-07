package main

import (
	"os"
	"strings"
	"testing"
)

func TestNativeTabNavigationUsesWindowsDialogSelectionOnlyForTab(t *testing.T) {
	b, err := os.ReadFile("tab_navigation_windows.go")
	if err != nil {
		t.Fatalf("read Tab navigation source: %v", err)
	}
	src := string(b)
	for _, required := range []string{
		`NewProc("SetWindowsHookExW")`,
		`NewProc("IsDialogMessageW")`,
		`whGetMessage`,
		`msg.Message == WM_KEYDOWN && msg.WParam == vkTab`,
		`procIsDialogMessageW.Call(app.hwnd, lparam)`,
		`msg.Message = wmNull`,
		`procCallNextHookEx.Call`,
	} {
		if !strings.Contains(src, required) {
			t.Fatalf("missing Tab-navigation safeguard %q", required)
		}
	}
	if strings.Contains(src, "VK_LEFT") || strings.Contains(src, "VK_RIGHT") || strings.Contains(src, "VK_UP") || strings.Contains(src, "VK_DOWN") {
		t.Fatal("first accessibility slice must not route arrow keys through dialog navigation")
	}
}

func TestExistingNativeAskAndSearchControlsRemainTabStops(t *testing.T) {
	src := windowsSource(t)
	for _, fragment := range []string{
		`"EDIT", "", WS_CHILD|WS_TABSTOP|ES_LEFT|ES_AUTOHSCROLL`,
		`"BUTTON", "Ask ECO", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON`,
		`"BUTTON", "Search all", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON`,
		`"BUTTON", "This item", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON`,
		`"BUTTON", "Previous", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON`,
		`"BUTTON", "Next", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON`,
		`"BUTTON", "Open match", WS_CHILD|WS_TABSTOP|BS_PUSHBUTTON`,
	} {
		if !strings.Contains(src, fragment) {
			t.Fatalf("expected native Tab-stop control %q", fragment)
		}
	}
}
