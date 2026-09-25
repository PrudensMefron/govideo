# GoVideo — Current State

Last documentation update: 2026-09-25.

## Audio trimming and default-app playback (2026-09-25)

Native playback follow-up: WebKitGTK reproduced a media-source error for a WAV
served under `wails://`. Preview playback now uses an ephemeral HTTP listener
bound exclusively to `127.0.0.1`, with opaque tokens, Host validation and Range
support. A native WebKitGTK test loaded, played, sought and reached the end of a
test WAV over loopback. Frontend regeneration now waits for prior preview cleanup,
eliminating the competing discard/render calls that triggered the busy error.
Regression tests cover that ordering, cancellation during cleanup and loopback
token/Host/range behavior. Native Windows playback remains unverified.

The preview player now uses the shared `AudioPlayer` Vue component: DaisyUI blue
play/pause and range controls for position/volume, with a themed blue surface.
The same HTML audio element and loopback stream still handle decoding/playback.
Native WebKitGTK verification covered play, pause, seeking and light/dark rendering.

- New **Recortar música** operation alongside download/conversion: select an audio artifact from the complete history or a local file through the native picker; inspect with FFprobe; remove seconds from the beginning and backwards from the end.
- Cancellable FFmpeg previews, bounded to one at a time, support MP3, M4A, Opus, WAV, FLAC, Ogg and AAC. The original is untouched until explicit replacement. A private WAV playback endpoint supports byte-range seeking in the embedded audio player; the saved artifact retains the source container and is encoded once before auditioning.
- Copies are saved beside the source with `- recortada` and exclusive incremental names. Replacement requires a confirmation dialog, a matching source fingerprint and a complete synchronized temporary file beside the original. Saved edits enter persistent Activities as `audio_trim`.
- Clicking a completed activity (or activating its title with the keyboard) opens the associated media application. Jobs with multiple outputs offer an explicit choice; folder buttons remain separate. Backend validation rejects missing, untracked and unsupported paths. Windows helpers retain hidden-console flags.
- Verification: Go race suite; real FFmpeg sample-exact WAV interval and seven-format encoding tests; copy/replacement/cancellation/concurrency/source-change tests; HTTP Range/token isolation tests; frontend production typecheck/build; Linux build and Windows AMD64 cross-build. Chromium with the actual Wails server/backend exercised preview playback/seeking, stale-preview invalidation, validation errors and save-to-history at 1180/760/320px. Native OS pickers, native WebKit/WebView2 playback and Windows file associations still require platform runtime verification.

## Frontend component cleanup (2026-09-25)

- Split the Vue shell into Home/Activities views, Media/Settings dialogs and an injected application composable, retaining generated Wails bindings.
- DaisyUI owns buttons, inputs, selects, radios, badges, alerts, modal surfaces and radial progress. Removed accumulated stylesheet overrides and collisions with reserved `progress`/`status` classes.
- Activity rows reserve consistent columns for title, progress and actions; narrow layouts explicitly place each grid item. Progress is blue while running and green when completed; unknown totals have no fabricated numeric percentage. Terminal rows omit stale phase/speed/ETA.
- Settings support saving/error feedback and discard unsaved edits on close; analysis captures the focus-return target before asynchronous work.
- Verified production typecheck/build and rendered Chromium fixtures at 1190, 760 and 320px, including real radial dimensions/colors, responsive layout, light/dark dialogs, Escape and focus restoration. This does not replace native Wails/Linux/Windows or real download end-to-end verification.

## Implemented MVP foundation (verified in this working tree)

- module renamed to `github.com/PrudensMefron/govideo`; Wails/runtime pinned at `v3.0.0-beta.20`;
- explicit output/audio/job/progress/artifact/error domain types and guarded job transitions;
- FIFO job manager with two workers, cancellation, retry, atomic JSON history and interrupted-job recovery;
- structured yt-dlp inspection/download progress, FFprobe validation and FFmpeg local audio conversion;
- managed yt-dlp `2026.08.19` for Linux/Windows with official SHA-256 verification and atomic install;
- typed Wails bridge, job/dependency events, native dialogs, file drop and artifact reveal;
- all direct yt-dlp, FFmpeg, FFprobe and dependency-validation processes use a shared launcher; on Windows it sets both `HideWindow` and `CREATE_NO_WINDOW`, preventing terminal flashes;
- Vue 3 pt-BR workflow for download, batch conversion, activities and settings, with generated bindings, Tailwind/DaisyUI, light/dark themes and pnpm/npm lockfiles;
- Wails updater connected directly to the public `PrudensMefron/govideo` GitHub Releases feed, with mandatory SHA-256 verification, delayed background checks, manual settings UI and hidden Windows helper process;
- GitHub Actions CI/CD using npm and Wails `beta.20`, running only for pushes to `master` after a merge, deriving the next patch version, creating the tag and draft release, and building Linux AMD64 and Windows AMD64.

Remaining hardening: managed FFmpeg archives/checksums (FFmpeg is currently detected from `PATH`), process-group/Windows Job Object termination, event coalescing, platform-native Downloads discovery, broader frontend tests, and Windows package verification. Browser fallback remains post-MVP.

The updater and release contract are documented in `07-auto-updater.md`. Real `N → N+1` update tests and independent release signing remain required before the first stable publication.

This document deliberately separates verified repository state, migration work already designed/implemented during development, and planned work.

## 1. Verified legacy repository state

The GitHub repository historically associated with this project is:

```text
PrudensMefron/downvideo
```

The remote repository state inspected on 2026-09-24 still exposes the legacy implementation:

```text
downvideo
requirements.txt
```

The legacy `downvideo` file is a Python executable script.

Its important existing behavior includes:

- URL sanitization;
- support for HTTP/HTTPS input;
- special handling for `blob:` URLs;
- direct `yt-dlp` probing;
- media metadata inspection;
- file-size estimation;
- DRM detection;
- optional browser fallback through Playwright/Chromium;
- network request observation to locate media candidates;
- candidate scoring for HLS/DASH/direct media;
- custom output directory selection;
- download progress output;
- post-processing output;
- FFmpeg availability warning.

The Python implementation is valuable as behavioral reference but is not the desired final architecture.

## 2. Go migration state defined so far

The new codebase is being designed around Go.

The following Core concepts have already been defined in the migration work and should be preserved if present in the working tree:

### Domain

- `Source`
- `Media`
- `Format`
- `Job`
- `JobID`
- `JobStatus`
- `DownloadRequest`
- typed download events

### Core boundaries

- `Extractor`
- `Downloader`
- event handler/event stream concept

### Source validation

The Core design supports:

- HTTP URLs;
- HTTPS URLs;
- `blob:` sources that require a page URL/browser discovery.

### Job statuses

Current conceptual statuses:

```text
pending
analyzing
ready
downloading
processing
completed
failed
cancelled
```

### yt-dlp adapter

A Go adapter design has been established around:

```text
yt-dlp --dump-single-json
```

The intended behavior is:

1. invoke `yt-dlp` via `exec.CommandContext`;
2. request structured JSON;
3. deserialize only the fields GoVideo needs;
4. map external DTOs into `core.Media`;
5. keep raw yt-dlp schema out of the Core.

The adapter uses the format selector conceptually equivalent to:

```text
bv*+ba/b
```

and preserves the legacy behaviors for:

- format mapping;
- duration;
- estimated size;
- codec presence;
- DRM flags.

### Dependency manager

A dependency manager has been designed/started for managed `yt-dlp`.

Initial Linux target:

```text
linux/amd64
```

Desired application-data path:

```text
$XDG_DATA_HOME/govideo/bin/yt-dlp
```

or, when `XDG_DATA_HOME` is unset:

```text
~/.local/share/govideo/bin/yt-dlp
```

The first Linux implementation is expected to:

- detect platform;
- check whether the file exists;
- execute `yt-dlp --version`;
- download the official standalone `yt-dlp_linux` asset if required;
- write to a temporary file;
- sync and close it;
- apply executable permissions;
- atomically rename into place;
- revalidate the installed executable.

Checksum validation remains an important production hardening step.

## 3. Important repository reconciliation rule

The current GitHub remote may lag behind local migration work.

An agent running inside the actual working tree must inspect the repository before assuming that every Go file described above exists.

If a described component is absent:

- do not mark it as completed;
- implement it only when it belongs to the current task;
- preserve this architecture unless there is a concrete reason to change it.

## 4. Product direction changed after the initial Core design

The initial UI idea was Bubble Tea/TUI.

That direction is obsolete.

The current UI target is:

```text
Wails 3 + Vue 3 + TypeScript
```

Do not spend implementation effort creating a Bubble Tea frontend unless explicitly requested for debugging or a separate tool.

The reusable Go Core concept remains valid and becomes even more important with Wails.

## 5. New functional requirements

The application is no longer only a video downloader.

### Online media output modes

For an analyzed URL, the user must be able to choose:

```text
Video
Audio
Both
```

The Core should model this explicitly.

### Local media conversion

The GUI must allow selecting a local video and extracting/converting it to audio.

This should be implemented through FFmpeg, not through a fake online-download workflow.

### Multi-platform

Initial platform targets are:

```text
Linux
Windows
```

Linux dependency management may be implemented first.

Platform-specific external binary management must not leak into domain packages.

## 6. Not yet considered complete

- managed FFmpeg installation, checksum validation and independent updates;
- process-tree cancellation through Linux process groups and Windows Job Objects;
- browser fallback port to Go;
- independent updater artifact signing and real `N → N+1` update tests;
- per-user distribution policy compatible with self-replacement on Windows and Linux;
- full native Windows package verification;
- broader frontend interaction and accessibility tests.

This list should shrink as work lands.
