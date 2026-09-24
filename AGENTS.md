# GoVideo — Agent Instructions

This file is the primary implementation guide for AI coding agents working on GoVideo.

## 1. Product identity

**GoVideo** is a cross-platform desktop application for downloading and converting online video/media.

Initial target platforms:

- Linux
- Windows

The application is no longer planned as a TUI. The primary user interface is a desktop GUI built with **Wails 3 + Vue 3**.

The backend/core is written in **Go** and must remain independent from the GUI wherever practical.

## 2. Core product capabilities

GoVideo must support the following user workflows:

1. Paste an online media URL, inspect it, and download it as:
   - video;
   - audio/music;
   - both video and audio.

2. Show relevant metadata before downloading, including when available:
   - title;
   - duration;
   - thumbnail;
   - estimated size;
   - available formats/qualities;
   - source/extraction status.

3. Track download progress and post-processing progress.

4. Allow multiple jobs and concurrent downloads with bounded concurrency.

5. Allow cancellation.

6. Allow the user to select a local video file and convert/extract its audio.

7. Manage required external runtime dependencies such as:
   - `yt-dlp`;
   - `ffmpeg`;
   - browser/browser automation runtime if browser fallback is implemented.

8. Avoid requiring the user to manually install `yt-dlp`.

## 3. Architecture rule

The GUI must not become the application core.

The desired dependency direction is:

```text
Vue 3 UI
   │
Wails services / bridge
   │
Application layer
   │
Core/domain
   │
Ports/interfaces
   │
Adapters
   ├── yt-dlp
   ├── ffmpeg
   ├── filesystem
   └── browser discovery (planned)
```

The Core must not import Vue, JavaScript-specific concepts, Wails-specific UI behavior, or frontend state.

Wails is an adapter/interface layer around Go services.

## 4. Source of truth

When implementation and documentation disagree:

1. Existing compiling code is the source of truth for exact names, paths, signatures, and currently implemented behavior.
2. These documents are the source of truth for architectural direction and planned product behavior.
3. Do not rewrite working code only to make it match an example shown in documentation.
4. If a documented architectural constraint conflicts with existing code, prefer a small refactor that preserves behavior.
5. Never silently invent completed features. Update `docs/01-current-state.md` when a milestone is actually implemented.

## 5. Implementation principles

Prefer:

- small Go packages with explicit responsibilities;
- interfaces at real system boundaries, not interfaces for every struct;
- `context.Context` for cancellable work;
- typed domain objects instead of passing raw `map[string]any`;
- structured events/state instead of parsing human-facing output;
- atomic filesystem writes for managed binaries;
- explicit errors with wrapping via `%w`;
- tests around parsers, mappers, dependency management, and job state;
- platform-specific files using Go build constraints when platform behavior diverges.

Avoid:

- putting download logic in Wails methods;
- returning raw `yt-dlp` JSON to Vue;
- calling external tools directly from Vue;
- global mutable job state without synchronization;
- unbounded goroutine creation;
- parsing ordinary human-readable `yt-dlp` progress output when structured output is available;
- coupling domain types to a specific downloader;
- automatic DRM circumvention.

## 6. External tools

### yt-dlp

Treat `yt-dlp` as a managed external engine.

Do not embed it in the Go executable.

The dependency manager should:

- detect whether the managed binary exists and works;
- download the official standalone binary when missing;
- place it in GoVideo's application data directory;
- validate it by executing `--version`;
- eventually validate downloads with release checksums;
- eventually support updates.

Linux support is implemented/planned first. Windows support follows through a platform-specific asset selection layer.

### ffmpeg

`ffmpeg` is required for important product features:

- combining separate video/audio streams;
- extracting audio;
- transcoding local files;
- producing audio-only downloads when conversion is needed.

It should eventually be managed similarly to `yt-dlp`, but do not conflate their installers.

## 7. Job model

Long-running work should be represented as jobs.

Expected lifecycle:

```text
pending
  ↓
analyzing
  ↓
ready
  ↓
downloading
  ↓
processing
  ↓
completed
```

Terminal states also include:

```text
failed
cancelled
```

A job should own enough information to represent:

- source;
- detected media;
- requested output;
- destination;
- progress;
- status;
- error;
- timestamps;
- cancellation.

The GUI observes job state/events. It must not own execution.

## 8. Output modes

The domain must explicitly represent output intent rather than encoding it in UI booleans.

Conceptually:

```go
type OutputMode string

const (
    OutputVideo OutputMode = "video"
    OutputAudio OutputMode = "audio"
    OutputBoth  OutputMode = "both"
)
```

Exact naming may differ if existing code already establishes another convention.

A "both" operation should represent one user job even if the adapter internally requires multiple outputs or post-processing steps.

## 9. Local conversion

Local video-to-audio conversion is a first-class feature, not a special UI hack.

It should eventually use an application service such as:

```text
ConvertLocalMedia(...)
```

backed by an FFmpeg adapter.

The frontend selects a file through Wails/native dialogs, then passes a safe path/request to the Go service.

## 10. Frontend

Target frontend:

- Vue 3;
- TypeScript;
- Vite through Wails tooling;
- Composition API;
- small stores/composables for UI state;
- Wails-generated bindings/types where appropriate.

Do not duplicate authoritative backend job state in complicated frontend state machines.

## 11. Platform scope

Current product scope:

- Linux
- Windows

Linux implementation may land first.

Do not introduce Windows-specific assumptions into core packages. Platform-specific dependency logic should be isolated behind files/packages such as:

```text
ytdlp_linux.go
ytdlp_windows.go
```

or equivalent strategy code.

## 12. Security and safety

- Do not implement DRM bypass.
- Do not execute shell strings constructed from user input.
- Prefer `exec.CommandContext` with argument slices.
- Validate file paths and destinations at Go boundaries.
- Never trust metadata returned by remote sites for filesystem paths without sanitization.
- Download managed binaries only from trusted official release sources.
- Verify checksums before considering dependency installation production-ready.

## 13. Before editing

Before making significant changes:

1. inspect the current repository tree;
2. inspect `go.mod`;
3. inspect existing packages and types;
4. read `docs/01-current-state.md`;
5. read the relevant architecture document;
6. preserve existing working behavior unless the task explicitly changes it.

After implementing a milestone, update `docs/01-current-state.md` and, if architecture changed, `docs/02-architecture.md`.
