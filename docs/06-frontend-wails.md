# GoVideo — Wails 3 and Vue 3 Frontend Guide

## Implemented bridge

`DesktopService` exposes analysis, download, batch conversion, paged job listing, cancel/retry, dependency status/retry, native input/destination selection, artifact reveal and settings. The app registers `job:updated`, `dependency:updated`, and `files:dropped`; the frontend fetches snapshots before subscribing. Native `<dialog>` elements replace browser alert/confirm/prompt APIs.

## Purpose

The frontend exists to present and control application state.

It should not implement media extraction, download orchestration, FFmpeg logic, dependency installation, or job execution.

## Stack

- Wails 3
- Vue 3
- TypeScript
- Composition API
- Vite/Wails frontend tooling

Use the Wails-generated bindings/models where practical.

## Wails status

As of 2026-09-24, Wails v3 is in beta.

The project should pin a specific beta version after validating it instead of silently upgrading every developer/build to the newest beta.

Wails 3 requires a modern Go toolchain; verify the requirement against the pinned version before CI/release setup.

## Backend service boundary

Expose coarse application operations.

Good:

```text
AnalyzeURL
StartDownload
ConvertLocalVideo
CancelJob
ListJobs
CheckDependencies
SelectDestination / native dialog helpers
```

Bad:

```text
RunArbitraryCommand
RunYTDLPWithArgs
RunFFmpegWithArgs
GetRawYTDLPJSON
```

The frontend should express intent, not command-line syntax.

## Request DTO example

Conceptually:

```ts
type StartDownloadRequest = {
  sourceURL: string
  destination: string
  outputMode: "video" | "audio" | "both"
  videoQuality?: string
  audioFormat?: string
}
```

The authoritative generated type should ultimately come from the Go/Wails model rather than being duplicated by hand.

## Views

Initial application can be kept compact.

Suggested views/areas:

### Home / Download

- URL input
- Analyze
- media preview
- output options
- destination
- Start

### Jobs

- active jobs
- progress
- phases
- cancellation
- completed jobs

This can initially be part of the home view instead of a separate route.

### Local conversion

- select video file
- select audio output options
- destination
- convert
- progress

### Settings

Later:

- default destination
- concurrency limit
- dependency status
- yt-dlp update behavior
- browser profile/fallback settings

Do not build settings screens before corresponding backend behavior exists.

## State management

Prefer local Vue state/composables for simple screens.

Introduce Pinia only if application-wide frontend state becomes complex enough to justify it.

Backend jobs are authoritative.

A frontend job representation should be treated as a projection/cache of backend state.

## Events

For long-running work, prefer backend → frontend events rather than polling every few milliseconds.

Example flow:

```text
Go worker
   ↓
job state/progress
   ↓
Wails event
   ↓
Vue listener
   ↓
update visible job
```

On initial page load/reconnect, fetch current job snapshots so event loss does not permanently desynchronize the UI.

## File dialogs

Use Wails/native dialogs for:

- local input video selection;
- output directory selection;
- possibly output file selection.

Do not implement a custom filesystem explorer unless there is a concrete product need.

## Errors

Separate:

- user-facing message;
- technical details.

Example:

```text
Download failed

The source could not be downloaded.

Details:
yt-dlp exited with status 1: ...
```

Do not expose cookies, authorization headers, or other sensitive browser context in error views.

## UI responsiveness

Never await a long-running download in a way that makes the user wait for the Wails method call to complete before receiving a Job ID.

Preferred model:

```text
StartDownload()
   ↓
returns Job immediately
   ↓
work continues in Go
   ↓
events update UI
```

## Styling direction

The product should look like a small modern desktop utility.

Priorities:

- clear information hierarchy;
- useful progress state;
- compact controls;
- obvious output mode;
- no web-dashboard bloat;
- good light/dark behavior eventually.

DaisyUI is the canonical component library. Do not couple Core work to styling.

## Audio editor and playback bindings

`AudioTrimView` and `useAudioTrim` implement the third Home operation, using
generated types and these desktop methods:

- `SelectInputAudio`, `ListAudioArtifacts`, `InspectAudio(context, path)`;
- `CreateAudioPreview(context, core.TrimRequest)` → `AudioPreviewDTO`, including `playbackURL`;
- `DiscardAudioPreview(id)`;
- `SaveAudioPreview(context, app.SaveAudioRequest)` → `app.SavedAudio`;
- `OpenArtifact(path)` for the associated media application.

Wails injects `context.Context`; callers cancel inspection/rendering through the
generated `CancellablePromise`. Saving publishes an already rendered file and is
not cancelled midway by page navigation. Preview regeneration and changing any
cut invalidate the old token/player. Actual download/conversion execution remains
in the existing job manager.

The shared `AudioPlayer` wraps an HTML audio element with DaisyUI play/pause and
native range inputs for keyboard seeking/volume. It exposes that element to the
editor for cleanup before saving or regenerating; theme colors use primary tokens.
Its source comes from `playbackURL`: an ephemeral loopback
HTTP server at `http://127.0.0.1:<port>/audio-preview/<token>`. This avoids the
native WebKitGTK media-source failure observed under the Wails custom URI scheme.
Only token-authorized preview WAVs are served; Host is validated and no external
network interface is bound. The listener closes on shutdown. Never
use a `file://` URL or expose a general filesystem-serving endpoint. WAV is decoded
from the prepared final artifact to avoid depending on WebView Opus/AAC support.
The full audio is streamed from disk, not serialized through a binding as base64.
Discard calls are serialized and awaited before regeneration. Test that ordering
with `node tests/audio-trim.test.cjs` from `frontend/`.

Activities use a native button in the title for keyboard activation and a row
click for pointer activation. Nested buttons do not propagate into playback;
multi-output activities use a DaisyUI/native dialog to choose which artifact opens.
