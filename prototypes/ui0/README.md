# ECO UI-0 framework gate

This directory is deliberately isolated from the production ECO application. It exists to answer one question before the next full candidate is built: **which modern native UI foundation actually meets ECO's hard requirements on Windows and on the Acer baseline?**

The same synthetic Matter projection and interaction tests are implemented in two candidates:

- `qt/` — Qt 6 / QML candidate
- `slint/` — Slint 1.17 candidate

Neither candidate is a release, and neither may read real evidence.

## Required UI-0 behaviours

Both candidates must demonstrate:

- Matter Workspace rather than a database-first home screen;
- live type-ahead search (`w` → `wa` → `war` immediately finds warranty content);
- selectable/copyable AI transcript text;
- a composer that remains reachable at 1366×768;
- responsive collapse of the context inspector;
- keyboard focus and ordinary desktop text behaviour;
- useful review finding cards with provenance and `Why does ECO think this?`;
- zero network requirement at runtime;
- synthetic information only.

The framework winner is selected only after Windows build, accessibility, packaging, memory, scaling and Acer measurements. This prototype does not pre-approve either framework.
