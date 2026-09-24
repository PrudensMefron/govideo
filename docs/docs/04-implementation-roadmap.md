# GoVideo — Implementation Roadmap

This roadmap is ordered to minimize rework.

Do not start all milestones at once.

## Phase 0 — Repository reconciliation

Before new work:

- confirm project/module has been renamed to GoVideo where intended;
- inspect current `go.mod`;
- inspect what migration code actually exists locally;
- preserve the Python script as legacy behavioral reference until feature parity is sufficient;
- ensure documentation is present in the working tree.

Exit criteria:

- actual code state is reflected in `01-current-state.md`.

## Phase 1 — Core foundation

Implement/confirm:

- `Source`;
- URL validation;
- `Media`;
- `Format`;
- `Job`;
- `JobStatus`;
- `OutputMode`;
- typed errors;
- adapter ports.

Add tests.

Exit criteria:

- core package has no Wails dependency;
- core tests pass.

## Phase 2 — Managed yt-dlp on Linux

Implement:

- application data directory;
- `yt-dlp` managed path;
- Linux/AMD64 asset selection;
- existence/regular-file checks;
- `--version` validation;
- HTTP download;
- temporary file;
- sync/close;
- `chmod`;
- atomic rename;
- post-install validation.

Then harden:

- SHA-256 checksum verification;
- explicit HTTP status validation;
- reasonable timeouts;
- clear install errors.

Exit criteria:

- fresh Linux machine/user profile can launch GoVideo backend and obtain a working managed yt-dlp without manual installation.

## Phase 3 — yt-dlp probe adapter

Implement:

- `exec.CommandContext`;
- structured JSON probe;
- DTO;
- mapper;
- duration conversion;
- format mapping;
- estimated size;
- DRM flags;
- adapter tests using fixtures.

Exit criteria:

- an online URL can produce a normalized `core.Media`.

## Phase 4 — FFmpeg strategy

Design and implement:

- FFmpeg port;
- dependency detection/manager;
- initial Linux strategy;
- process execution;
- error capture;
- cancellation.

Do not block local-conversion work on sophisticated update logic.

Exit criteria:

- Go code can invoke a known FFmpeg binary through a typed adapter.

## Phase 5 — Output requests

Create typed download/output configuration.

Support:

```text
video
audio
both
```

Decide initial audio output format explicitly.

Implement request-to-yt-dlp argument mapping.

Exit criteria:

- backend can produce video-only/default-video, audio output, and both artifacts through programmatic calls.

## Phase 6 — Local conversion

Implement:

```text
local video → FFmpeg → audio
```

Include:

- input validation;
- output path;
- overwrite policy;
- cancellation;
- progress if practical.

Exit criteria:

- backend test/manual CLI harness can convert a local video without Wails.

## Phase 7 — Job manager

Implement:

- synchronized job registry;
- ID creation;
- lifecycle transitions;
- bounded concurrency;
- per-job `context.Context`;
- cancellation;
- progress snapshots/events;
- terminal states.

Define transition rules to prevent impossible states.

Exit criteria:

- multiple programmatic jobs can run safely;
- concurrency limit works;
- cancellation works.

## Phase 8 — Wails 3 scaffold

Pin a known Wails 3 beta version.

Create:

- Wails application;
- Vue 3 + TypeScript frontend;
- main window;
- Go application services;
- generated frontend bindings.

Do not move Core logic into service methods.

Exit criteria:

- Vue can call a trivial backend health/dependency method.

## Phase 9 — Online download GUI

Build:

- URL input;
- Analyze button;
- metadata display;
- thumbnail;
- output mode selection;
- destination picker;
- Download button;
- job list;
- progress;
- cancellation;
- error states.

Exit criteria:

- primary online workflow can be completed entirely through GUI.

## Phase 10 — Local conversion GUI

Build:

- local file picker;
- detected file summary;
- output audio settings;
- destination;
- conversion progress;
- cancellation.

Exit criteria:

- local video-to-audio workflow works through GUI.

## Phase 11 — Windows support

Implement/test:

- Windows managed yt-dlp asset;
- Windows executable naming/path handling;
- FFmpeg strategy;
- Wails/WebView2 behavior;
- process cancellation;
- packaging.

Exit criteria:

- clean Windows system/user can run packaged app and complete major workflows.

## Phase 12 — Browser fallback

Port/rebuild legacy fallback:

- browser launch;
- optional persistent profile;
- network request observation;
- candidate scoring;
- headers/cookies required for candidate reuse;
- yt-dlp candidate validation.

Keep behind a discovery interface.

Exit criteria:

- supported pages that fail direct extraction can recover through browser discovery.

## Phase 13 — Dependency updates and production hardening

Implement:

- version checks;
- managed update policy;
- safe replacement;
- checksum verification for every download;
- rollback strategy if useful;
- dependency UI status.

Also:

- packaging;
- release signing strategy;
- error reporting/log locations;
- update strategy for GoVideo itself.

## Working rule for Codex

After completing a phase or a substantial subphase:

1. run tests;
2. update `docs/01-current-state.md`;
3. mark only actually working behavior as implemented;
4. document newly introduced architectural decisions.
