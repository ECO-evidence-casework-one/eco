ECO RIG AI PREVIEW
==================

WHAT TO DO

1. Double-click START_RIG_AI_PREVIEW.cmd
2. Leave the window open while it checks/builds/downloads.
3. Do not disable Windows security if Windows or the script stops it.

WHAT THE SETUP DOES

- Uses the controlled ECO source from GitHub.
- Runs ECO's tests before building the app.
- Downloads the checked llama.cpp runtime.
- Downloads Qwen2.5 1.5B Instruct Q4_K_M (about 1.1 GB).
- Checks the exact model/runtime fingerprints.
- Makes Qwen generate a real offline test answer.
- Opens ECO only if all required checks pass.

WHERE IT GOES

If your E: drive is available, the preview is created at:
E:\ECO_RIG_AI_PREVIEW

Otherwise it uses an ECO_RIG_AI_PREVIEW folder next to these setup files.

RETRY AFTER A STOPPED SETUP

If E:\ECO_RIG_AI_PREVIEW already contains AI_SETUP_RESULT.txt whose first line is
ECO RIG AI SETUP STOPPED, the corrected setup preserves that whole failed attempt
under a timestamped .failed-... folder and creates a fresh preview automatically.
Do not delete the old failed attempt yourself.

IF IT WORKS

The setup window will say:
ECO RIG AI SETUP PASS
Real offline Qwen generation: PASS

ECO should then open. In Ask ECO, the answer area should identify Qwen as ready/used rather than making you guess.

IF IT STOPS

Do not change Defender, Smart App Control or PowerShell machine policy.
Open AI_SETUP_RESULT.txt in the preview folder and send that file back to the ECO development chat.

TEST BOUNDARY

This is still a private developer preview. Use synthetic/test material, not real sensitive evidence, until the wider acceptance work is complete.
