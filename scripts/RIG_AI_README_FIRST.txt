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
- Shows Qwen download size, percentage and transfer speed about every 15 seconds.
- Stops/retries a transfer that makes no progress for 3 minutes.
- Keeps partial Qwen bytes and resumes them on the next attempt/run when possible.
- Checks the exact model/runtime fingerprints.
- Makes Qwen generate a real offline test answer.
- Opens ECO only if all required checks pass.

WHERE IT GOES

If your E: drive is available, the requested preview location is:
E:\ECO_RIG_AI_PREVIEW

Otherwise it uses an ECO_RIG_AI_PREVIEW folder next to these setup files.

If Windows still has a disposable file from a failed attempt locked, the corrected
setup will leave that failed folder untouched and automatically use a fresh sibling
folder such as:
E:\ECO_RIG_AI_PREVIEW.retry-...

The setup window always prints:
Actual preview location: ...

Use that actual location for AI_SETUP_RESULT.txt and START_ECO_WITH_AI.cmd.

RETRY AFTER A STOPPED OR INTERRUPTED SETUP

The corrected setup recognises both a recorded STOPPED setup and an interrupted setup
that was closed before AI_SETUP_RESULT.txt could be written. It first attempts the
normal timestamped archive. If Windows refuses because an old work file is still in
use, the failed directory is preserved in place instead of making the whole setup fail.

If the interrupted attempt contains:
AIAssets\qwen2.5-1.5b-instruct-q4_k_m.gguf.part

the partial model download is moved intact into the clean retry seed and the qualified
v3 preparer resumes it rather than deliberately starting from zero. The setup never
starts a second downloader against a partial file that is still locked.

Do not delete the previous attempt yourself.

IF IT WORKS

The setup window will say:
ECO RIG AI SETUP PASS
Real offline Qwen generation: PASS

ECO should then open. In Ask ECO, the answer area should identify Qwen as ready/used rather than making you guess.

IF IT STOPS

Do not change Defender, Smart App Control or PowerShell machine policy.
Use the folder shown after "Actual preview location". Open AI_SETUP_RESULT.txt there
and send that file back to the ECO development chat.
Any .part model file is intentionally retained for another safe resume attempt.

TEST BOUNDARY

This is still a private developer preview. Use synthetic/test material, not real sensitive evidence, until the wider acceptance work is complete.
