# GoVideo — Current State

Last documentation update: 2026-09-24.

## Implemented MVP foundation (verified in this working tree)

- module renamed to `github.com/PrudensMefron/govideo`; Wails/runtime pinned at `v3.0.0-beta.20`;
- explicit output/audio/job/progress/artifact/error domain types and guarded job transitions;
- FIFO job manager with two workers, cancellation, retry, atomic JSON history and interrupted-job recovery;
- structured yt-dlp inspection/download progress, FFprobe validation and FFmpeg local audio conversion;
- managed yt-dlp `2026.08.19` for Linux/Windows with official SHA-256 verification and atomic install;
- typed Wails bridge, job/dependency events, native dialogs, file drop and artifact reveal;
- all direct yt-dlp, FFmpeg, FFprobe and dependency-validation processes use a shared launcher; on Windows it sets both `HideWindow` and `CREATE_NO_WINDOW`, preventing terminal flashes;
- Vue 3 pt-BR workflow for download, batch conversion, activities and settings, with generated bindings, Tailwind/DaisyUI, light/dark themes and a pnpm lockfile/build workflow.

Remaining hardening: managed FFmpeg archives/checksums (FFmpeg is currently detected from `PATH`), process-group/Windows Job Object termination, event coalescing, platform-native Downloads discovery, broader frontend tests, and Windows package verification. Browser fallback and update discovery remain post-MVP.

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

Unless the working tree proves otherwise, do not consider these completed:

- Wails 3 shell;
- Vue 3 GUI;
- Wails service bridge;
- job manager;
- bounded concurrent worker execution;
- structured live download progress;
- cancellation end-to-end;
- FFmpeg dependency manager;
- audio-only online workflow;
- "both" workflow;
- local video-to-audio conversion;
- Windows dependency management;
- browser fallback port to Go;
- release checksum validation;
- yt-dlp update flow;
- persistent user settings;
- download history.

This list should shrink as work lands.
