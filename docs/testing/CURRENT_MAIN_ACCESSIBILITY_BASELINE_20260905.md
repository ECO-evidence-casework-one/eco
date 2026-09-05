# ECO current-main accessibility baseline

Source commit: `dbe458901bbfc8eb6bb9b91e781ddbc97307c29b`
Standard child HWNDs inspected: **9**
Ask/search native semantic controls: **PASS**
Known custom-painted interactive labels with no child HWND semantic peer: **10 / 10**

## Missing native semantic peers in this baseline
- Home
- Evidence
- Matters
- Review
- Changes
- Trust & settings
- +  Add files
- +  Add folder
- Paste image
- Open native preview

## Interpretation
- Existing native EDIT/BUTTON controls expose established Windows accessibility contracts.
- The listed custom-painted interactions are not represented as child HWND semantic controls and therefore are not qualified for Narrator/NVDA from this evidence.
- Hosted-runner MSAA evidence is a structural baseline only; physical Windows 11 Narrator/NVDA remains mandatory.
