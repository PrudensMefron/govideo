# GoVideo — Architecture

## Architectural objective

GoVideo should be a desktop application whose UI can evolve without forcing a rewrite of media/download logic.

The central architectural constraint is:

> Wails and Vue are clients of the application core, not the owners of it.

## Layers

```text
┌─────────────────────────────────────────────┐
│ Vue 3 + TypeScript                         │
│ Views / components / presentation state    │
└──────────────────────┬──────────────────────┘
                       │ generated bindings/events
┌──────────────────────▼──────────────────────┐
│ Wails 3 adapter                            │
│ Desktop services, dialogs, event bridge     │
└──────────────────────┬──────────────────────┘
                       │
┌──────────────────────▼──────────────────────┐
│ Application layer                           │
│ Analyze, download, convert, cancel, jobs    │
└──────────────────────┬──────────────────────┘
                       │
┌──────────────────────▼──────────────────────┐
│ Core/domain                                 │
│ Media, source, jobs, requests, events       │
└──────────────────────┬──────────────────────┘
                       │ ports/interfaces
          ┌────────────┼────────────┐
          │            │            │
┌─────────▼───┐ ┌──────▼─────┐ ┌──▼──────────────┐
│ yt-dlp      │ │ FFmpeg      │ │ Browser/media   │
│ adapter     │ │ adapter     │ │ discovery       │
└─────────────┘ └────────────┘ └─────────────────┘
```

## Proposed package layout

Exact names may be adapted to the actual Wails scaffold.

```text
govideo/
├── cmd/
│   └── govideo/
│       └── main.go
│
├── internal/
│   ├── core/
│   │   ├── source.go
│   │   ├── media.go
│   │   ├── job.go
│   │   ├── event.go
│   │   ├── output.go
│   │   └── ports.go
│   │
│   ├── app/
│   │   ├── service.go
│   │   ├── analyze.go
│   │   ├── download.go
│   │   ├── convert.go
│   │   └── jobs.go
│   │
│   ├── jobs/
│   │   ├── manager.go
│   │   └── worker.go
│   │
│   ├── extractor/
│   │   └── ytdlp/
│   │       ├── adapter.go
│   │       ├── dto.go
│   │       ├── mapper.go
│   │       └── progress.go
│   │
│   ├── media/
│   │   └── ffmpeg/
│   │       ├── adapter.go
│   │       └── progress.go
│   │
│   ├── deps/
│   │   ├── manager.go
│   │   ├── ytdlp_linux.go
│   │   ├── ytdlp_windows.go
│   │   ├── ffmpeg_linux.go
│   │   └── ffmpeg_windows.go
│   │
│   ├── browser/
│   │   └── ...
│   │
│   └── desktop/
│       └── ...
│
├── frontend/
│   └── Vue application
│
├── docs/
└── go.mod
```

This is a target shape, not a command to create empty packages prematurely.

## Domain model

### Source

Represents where online media comes from.

Conceptually:

```go
type Source struct {
    URL     string
    PageURL string
    Headers map[string]string
}
```

Headers may be required when browser discovery produces a stream that needs Referer/User-Agent/Cookie context.

### Media

Represents normalized metadata.

It should not expose raw yt-dlp data structures.

Typical fields:

- ID;
- title;
- webpage URL;
- thumbnail;
- duration;
- estimated size;
- DRM flag;
- available formats.

### Output mode

Online download intent must be modeled explicitly.

Recommended shape:

```go
type OutputMode string

const (
    OutputVideo OutputMode = "video"
    OutputAudio OutputMode = "audio"
    OutputBoth  OutputMode = "both"
)
```

This is preferable to independent `downloadVideo bool` / `downloadAudio bool` flags because invalid combinations become harder to represent.

### Audio options

Do not force audio format decisions into the first domain model unless required.

A future request may contain:

```go
type AudioOptions struct {
    Format  AudioFormat
    Bitrate int
}
```

Initial supported format should be chosen deliberately during implementation.

### Job

A job is the unit of long-running execution.

A job should contain or reference:

- ID;
- source or local file request;
- normalized media;
- requested output mode;
- destination;
- state;
- progress;
- created/updated timestamps;
- failure;
- cancellation context.

One user action selecting "Both" should remain one logical job even if multiple external tool invocations occur.

## Application services

The Wails layer should call application services.

Conceptual APIs:

```text
AnalyzeURL(...)
StartDownload(...)
ConvertLocalVideo(...)
CancelJob(...)
ListJobs(...)
GetJob(...)
CheckDependencies(...)
```

Exact Go signatures should use typed request/response structs.

## yt-dlp boundary

Use `yt-dlp` as an external process.

Prefer:

```go
exec.CommandContext(ctx, binary, args...)
```

Do not invoke a shell.

Probe media using structured output.

Download progress should also be emitted in a machine-parseable format wherever possible.

Map yt-dlp DTOs into GoVideo domain objects at the adapter boundary.

## FFmpeg boundary

FFmpeg should be responsible for media-processing tasks such as:

- merging streams;
- extracting audio;
- transcoding audio;
- converting local video to audio.

Do not make FFmpeg details part of Vue components.

The FFmpeg adapter should support cancellation through `context.Context`.

Progress should eventually be normalized into the same job/event model used by downloads.

## Job manager

A job manager should own concurrency.

Desired behavior:

```text
UI StartDownload
      ↓
Application service
      ↓
Job created
      ↓
Job manager
      ↓
bounded worker slot
      ↓
adapter work
      ↓
events/state updates
      ↓
Wails event bridge
      ↓
Vue
```

Use a bounded semaphore/worker pool rather than launching unlimited work.

Do not hold locks while blocking on external processes or network operations.

## Events

The backend should expose meaningful application events, not raw process output.

Examples:

```text
job:created
job:updated
job:progress
job:completed
job:failed
job:cancelled
dependency:status
```

Payloads should use Wails-friendly typed structs.

The exact Wails v3 event API should be verified against the pinned version before implementation.

## Application updater boundary

Updating GoVideo itself is a desktop-adapter responsibility, separate from the
managed `yt-dlp`/FFmpeg dependencies and from media jobs:

```text
Vue settings/update notice
          ↓ typed intent/status
Wails DesktopService
          ↓
internal/desktop/update
          ↓
Wails updater + mandatory SHA256SUMS
          ↓
PrudensMefron/govideo GitHub Releases
```

The updater must not enter `internal/core` or the media job manager. The
frontend must not choose download URLs or execute release assets. See
`07-auto-updater.md` for the release contract, verification model and Linux/Windows
installation constraints.

## Browser fallback

The Python implementation already demonstrated a fallback strategy:

```text
direct yt-dlp probe
        ↓ failure
browser opens target page
        ↓
network requests observed
        ↓
HLS / DASH / direct media candidates
        ↓
candidate scoring
        ↓
yt-dlp validates candidates
```

This remains a planned capability.

It should live behind a media discovery interface rather than inside the UI or yt-dlp adapter.

No DRM circumvention.

## Filesystem rules

- User destinations are user-controlled.
- Managed dependencies belong in application data directories.
- Temporary downloads/installations must use safe temporary files.
- Dependency replacement should be atomic.
- External metadata must not be trusted directly as arbitrary paths.

## Dependency direction

Allowed:

```text
desktop -> app -> core
extractor/ytdlp -> core
media/ffmpeg -> core
jobs -> core
deps -> stdlib/platform
```

Avoid:

```text
core -> Wails
core -> Vue
core -> yt-dlp package internals
core -> ffmpeg command implementation
```

## Implemented package boundaries

`internal/app` owns orchestration, bounded execution and persistence; `internal/deps` owns dependency state and managed yt-dlp; `internal/extractor/ytdlp` owns inspection/download; `internal/media/ffmpeg` owns local conversion; and `internal/platformfs` owns collision-safe output allocation. `desktopservice.go` is the Wails-only DTO/dialog adapter. A “both” request remains one job while the application layer runs video and audio as separate argument-safe invocations.

All adapters create subprocesses through `internal/process.CommandContext`. Its Windows implementation combines `syscall.SysProcAttr.HideWindow` with the `CREATE_NO_WINDOW` creation flag, so direct engine execution never opens a console window.

## Audio editing boundary

`core.TrimRequest` defines seconds to remove from the start/end. `core.TrimDuration`
validates finite nonnegative values and requires at least 100ms of retained audio.
`app.AudioEditor` coordinates the `AudioEngine` port: probe, encode the retained
interval, then decode that encoded artifact to WAV for WebView auditioning. The
FFmpeg adapter owns codecs/arguments; application/core code has no Wails imports.

Previews are a separate ephemeral editing session, not queued downloads. One
render/save may execute at a time. Cancellation flows from Wails call context to
FFmpeg; leaving the editor discards playback and cancels active rendering. A
successful save records a completed `audio_trim` activity in the existing history.

Private temporary directories hold previews. Only an opaque, random token is
exposed; `internal/desktop/audio` adapts `OpenPreview(token)` to a loopback HTTP server
with `http.ServeContent`, GET/HEAD and byte-range seeking. Its listener binds only
`127.0.0.1:0`, checks Host and closes on shutdown. The desktop DTO adds the URL;
the core never owns transport addresses. It never accepts raw
paths from HTTP. Preview files are discarded when replaced, explicitly cleared,
saved or on normal application shutdown; crash leftovers remain OS temporary data.

New copies use exclusive creation. Replacement stages a synchronized file in the
original directory and rechecks identity/content immediately before rename. No
background action replaces the original. Editing changes invalidate the frontend
preview so saved output always corresponds to the tested selection. Regeneration
awaits asynchronous discard before requesting another render, because both use
the same backend operation lock.

`Service.ResolveArtifact` authorizes completed audio/video outputs before
`internal/desktop/fileopen` invokes `xdg-open` (Linux) or FileProtocolHandler
(Windows) via argument slices and the shared hidden-console process launcher.
