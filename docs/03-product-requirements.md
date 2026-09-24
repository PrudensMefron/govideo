# GoVideo — Product Requirements

## MVP product definition

GoVideo MVP is a Linux/Windows desktop application capable of analyzing online video URLs and producing video and/or audio files with visible progress.

The MVP should also convert a user-selected local video file into audio.

## Functional requirements

### FR-001 — Analyze URL

The user can paste an HTTP/HTTPS media/page URL.

The backend analyzes it and returns normalized media metadata.

Minimum response:

- title;
- duration when available;
- thumbnail when available;
- formats/quality data when useful;
- size estimate when available;
- whether extraction succeeded.

### FR-002 — Select output mode

Before downloading, the user can choose:

- Video
- Audio
- Both

The selected mode must be part of the backend request.

### FR-003 — Video download

The application downloads a video to a user-selected destination.

Where video and audio streams are separate, FFmpeg may be required for merge/post-processing.

### FR-004 — Audio download

The application can produce an audio/music file from an online source.

Implementation may involve yt-dlp format selection plus FFmpeg extraction/transcoding.

Audio format and quality should eventually be user-selectable.

The first implementation may support a single documented default format.

### FR-005 — Both

The application can produce both a video output and an audio output from the same analyzed media.

This is one user job with multiple artifacts.

The UI must not force the user to start two unrelated jobs.

### FR-006 — Local video to audio

The user can select a local video file through the desktop GUI.

The backend converts/extracts its audio through FFmpeg and writes the output to a selected destination.

### FR-007 — Progress

Long-running operations expose progress to the GUI.

When available, progress includes:

- percentage;
- downloaded/processed bytes;
- total size;
- speed;
- ETA;
- current phase.

Possible phases include:

```text
analyzing
downloading
merging
extracting-audio
transcoding
finalizing
```

### FR-008 — Cancellation

An active job can be cancelled.

Cancellation should terminate relevant child processes and transition the job into a cancelled terminal state.

### FR-009 — Concurrent jobs

Multiple jobs can exist simultaneously.

Execution concurrency must be bounded.

The exact default concurrency limit is a configuration decision, not a hard-coded product assumption in this document.

### FR-010 — Managed yt-dlp

Normal users should not have to install `yt-dlp`.

GoVideo manages its own compatible executable.

### FR-011 — Managed FFmpeg

The desired product experience is also to avoid requiring manual FFmpeg installation.

Initial development may temporarily detect a system FFmpeg, but the product architecture should support a managed FFmpeg installation.

### FR-012 — Dependency status

The application can report whether required engines are:

- installed;
- valid;
- missing;
- outdated, when update checking is implemented;
- failed to install.

### FR-013 — Destination selection

The user can select output directories through native/Wails file dialogs.

The backend validates writable destinations.

### FR-014 — Error presentation

Backend errors must preserve technical detail.

The GUI presents a useful user-facing summary and may expose technical details separately.

## UX model

Suggested primary application layout:

```text
┌─────────────────────────────────────────────────────────┐
│ GoVideo                                                 │
├─────────────────────────────────────────────────────────┤
│ URL                                                     │
│ [ https://...                                      ]    │
│                                      [ Analyze ]         │
├─────────────────────────────────────────────────────────┤
│ Thumbnail   Title                                       │
│             Duration / estimated size                   │
│                                                         │
│ Output:  (•) Video  ( ) Audio  ( ) Both                │
│ Quality: [ Best ▼ ]                                     │
│ Audio:   [ default ▼ ]                                  │
│ Destination: ~/Videos                    [ Browse ]      │
│                                          [ Download ]    │
├─────────────────────────────────────────────────────────┤
│ Jobs                                                    │
│ Video A     Downloading  ███████░ 72%     [ Cancel ]    │
│ Video B     Processing   █████████░ 91%                  │
└─────────────────────────────────────────────────────────┘
```

A separate action/view can expose:

```text
Convert local video → audio
```

Do not treat this ASCII layout as a pixel-perfect design requirement.

## Platform requirements

### Linux

Primary early development target.

Target architecture initially may be AMD64.

Wails v3 currently uses GTK4 + WebKitGTK 6.0 by default, with platform packaging implications.

Managed yt-dlp currently starts with the official Linux standalone asset.

### Windows

Required product platform.

Wails uses WebView2.

Dependency manager must select Windows-specific `yt-dlp` and FFmpeg assets.

Paths, executable names, permissions, process behavior and packaging must be isolated from Linux assumptions.

## Non-functional requirements

### Responsiveness

No blocking network, download, conversion, process wait, or filesystem-heavy work on the GUI event path.

### Reliability

Interrupted dependency downloads must not leave valid-looking partial binaries.

### Testability

Core behavior should be testable without launching Wails.

Mappers/parsers should be testable without internet access.

### Maintainability

Third-party schemas should be isolated in adapters.

### Security

No DRM bypass.

Never construct shell commands by concatenating user input.

Use argument arrays and validated filesystem paths.

### Observability

Errors should retain enough context to diagnose:

- external command;
- phase;
- exit failure;
- stderr where safe;
- job ID.

Avoid logging sensitive cookies/authorization headers.
