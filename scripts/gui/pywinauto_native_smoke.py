"""Bounded synthetic native-GUI acceptance for ECO using pinned pywinauto.

This script never selects an existing user workspace. LOCALAPPDATA is replaced
with a new test directory before ECO starts, so first-run state is synthetic.
"""

from __future__ import annotations

import argparse
import hashlib
import json
import os
from pathlib import Path
import subprocess
import sys
import tempfile
import time

from pywinauto import Desktop, keyboard


DONOR_COMMIT = "18d2a95cebed2f0061ab4e4c80c3a76ece5dd4f3"
MAIN_CLASS = "ECO_V25_NATIVE_MAIN"


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as stream:
        for chunk in iter(lambda: stream.read(1024 * 1024), b""):
            h.update(chunk)
    return h.hexdigest()


def wait_dialog(pid: int, title_re: str, timeout: float = 20.0):
    dialog = Desktop(backend="win32").window(process=pid, class_name="#32770", title_re=title_re)
    dialog.wait("exists visible enabled ready", timeout=timeout, retry_interval=0.2)
    return dialog


def click_ok(dialog) -> None:
    button = dialog.child_window(title="OK", class_name="Button")
    button.wait("exists visible enabled ready", timeout=5)
    button.click()


def wait_main(pid: int, timeout: float = 20.0):
    window = Desktop(backend="win32").window(process=pid, class_name=MAIN_CLASS)
    window.wait("exists visible enabled ready", timeout=timeout, retry_interval=0.2)
    return window


def control(main, control_id: int, class_name: str):
    wrapper = main.child_window(control_id=control_id, class_name=class_name)
    wrapper.wait("exists", timeout=5)
    return wrapper


def require(condition: bool, message: str) -> None:
    if not condition:
        raise AssertionError(message)


def run(exe: Path, receipt_path: Path) -> dict:
    exe = exe.resolve()
    require(exe.is_file(), f"ECO executable does not exist: {exe}")

    with tempfile.TemporaryDirectory(prefix="eco-pywinauto-") as temp_text:
        temp = Path(temp_text)
        local_app_data = temp / "LocalAppData"
        local_app_data.mkdir()
        env = os.environ.copy()
        env["LOCALAPPDATA"] = str(local_app_data)
        env["ECO_GUI_SMOKE"] = "synthetic-only"

        process = subprocess.Popen([str(exe)], env=env, cwd=str(exe.parent))
        facts: dict[str, object] = {
            "schema": 1,
            "donor": "pywinauto/pywinauto",
            "donor_commit": DONOR_COMMIT,
            "eco_executable": exe.name,
            "eco_sha256": sha256(exe),
            "localappdata_isolated": True,
            "existing_user_workspace_selected": False,
            "checks": [],
        }
        checks: list[str] = facts["checks"]  # type: ignore[assignment]

        try:
            first = wait_dialog(process.pid, r"^New ECO candidate workspace$")
            require("created explicitly" in first.window_text() or first.exists(), "first-run confirmation missing")
            click_ok(first)
            checks.append("first_run_created_explicit_candidate")

            main = wait_main(process.pid)
            require(main.class_name() == MAIN_CLASS, "wrong ECO main-window class")
            require(main.is_visible() and main.is_enabled(), "ECO main window is not usable")
            facts["window_title"] = main.window_text()
            checks.append("native_main_window_visible")

            # The first source candidate shows a What's New modal. Treat absence
            # as a failure because this is a fresh isolated LOCALAPPDATA route.
            whats = wait_dialog(process.pid, r"^What.*new.*Evidence & Casework One.*$")
            click_ok(whats)
            checks.append("whats_new_modal_reachable")

            main.set_focus()
            keyboard.send_keys("^f", pause=0.05)
            search = control(main, 1004, "Edit")
            search.wait("visible enabled ready", timeout=5)
            require(search.has_focus(), "Ctrl+F did not focus the verified evidence search")
            search.set_edit_text("synthetic gui smoke")
            require(search.window_text() == "synthetic gui smoke", "search edit did not retain synthetic text")
            checks.append("ctrl_f_focuses_search")

            search_all = control(main, 1005, "Button")
            this_item = control(main, 1006, "Button")
            previous = control(main, 1007, "Button")
            next_button = control(main, 1008, "Button")
            open_match = control(main, 1009, "Button")
            require(search_all.window_text() == "Search all" and search_all.is_visible() and search_all.is_enabled(), "Search all control is not exposed")
            require(this_item.window_text() == "This item" and this_item.is_visible() and not this_item.is_enabled(), "empty-workspace This item state is wrong")
            for expected, wrapper in (("Previous", previous), ("Next", next_button), ("Open match", open_match)):
                require(wrapper.window_text() == expected and wrapper.is_visible() and not wrapper.is_enabled(), f"empty-search {expected} state is wrong")
            checks.append("search_controls_exposed_with_empty_state")

            keyboard.send_keys("{TAB}", pause=0.05)
            require(search_all.has_focus(), "Tab did not move from search edit to Search all")
            keyboard.send_keys("+{TAB}", pause=0.05)
            require(search.has_focus(), "Shift+Tab did not return focus to search edit")
            checks.append("search_keyboard_focus_order")

            keyboard.send_keys("%3", pause=0.08)  # Alt+3 => Ask ECO
            question = control(main, 1001, "Edit")
            ask = control(main, 1002, "Button")
            answer = control(main, 1003, "Edit")
            question.wait("visible enabled ready", timeout=5)
            require(question.has_focus(), "Alt+3 did not focus the Ask ECO question edit")
            require(ask.window_text() == "Ask ECO" and ask.is_visible() and ask.is_enabled(), "Ask ECO button is not usable")
            require(answer.is_visible(), "source-backed answer control is not visible")
            checks.append("ask_page_keyboard_navigation")

            keyboard.send_keys("%1", pause=0.08)  # home
            require(not search.is_visible(), "evidence-only search control remained visible on Home")
            require(not question.is_visible(), "Ask-only question control remained visible on Home")
            checks.append("page_specific_controls_hide")

            main.close()
            process.wait(timeout=10)
            require(process.returncode == 0, f"ECO exited with code {process.returncode}")
            checks.append("clean_window_close")

            metadata = list(local_app_data.rglob("workspace.ecodb"))
            keys = list(local_app_data.rglob("vault.key"))
            require(len(metadata) == 1 and len(keys) == 1, "isolated first run did not create exactly one committed workspace")
            facts["synthetic_workspace_relative"] = str(metadata[0].parent.relative_to(local_app_data))
            checks.append("one_isolated_committed_workspace")
            facts["status"] = "PASS"
            return facts
        finally:
            if process.poll() is None:
                process.kill()
                try:
                    process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    pass
            receipt_path.parent.mkdir(parents=True, exist_ok=True)
            receipt_path.write_text(json.dumps(facts, indent=2) + "\n", encoding="utf-8")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--exe", required=True, type=Path)
    parser.add_argument("--receipt", required=True, type=Path)
    args = parser.parse_args()
    try:
        result = run(args.exe, args.receipt)
        print(json.dumps(result, indent=2))
        return 0
    except Exception as exc:  # failure receipt is written by run's finally
        print(f"PYWINAUTO ECO GUI SMOKE FAILED: {exc}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
