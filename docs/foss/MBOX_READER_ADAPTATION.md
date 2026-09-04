# MBOX reader adaptation

Date: 2026-09-04

ECO uses `emersion/go-mbox` (MIT) at exact acquired commit `1345da99f1254a23f517ffdc979f92359442473d` for MBOX message framing instead of inventing a mailbox parser. The source was acquired in FOSS donor Wave 2 with source-archive SHA-256 `b96b0ef7939de0fbe93557e7f8228f23ef484452e1909b74dc788415e7ab0566`.

ECO remains responsible for evidence-specific safety around the donor: the mailbox is streamed rather than loaded wholesale; automatic scanning is limited to 256 MiB, 10,000 messages and 8 MiB per individual message; extracted readable output remains under ECO's 24 MiB text bound; parser diagnostics are bounded; and malformed mailboxes fail closed while the preserved original remains authoritative.

Each framed message is passed through ECO's existing `net/mail` / MIME readable-text path, so MBOX and EML share the same header/body handling rather than duplicating MIME logic. The feature is read-only and introduces no network behavior. A Go fuzz target preserves hostile-input coverage for the MBOX boundary parser, while normal tests execute its seed corpus deterministically.
