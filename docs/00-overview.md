# GoVideo — Product Overview

## Purpose

GoVideo is a desktop media utility focused on making common video download and conversion tasks simple for end users.

The application began as a single Python script that used `yt-dlp` and Playwright. The project is being redesigned as a proper desktop application with a reusable Go core and a graphical interface.

The target application should feel like a normal desktop utility rather than a wrapper that exposes command-line tools.

## Main use cases

### Download online video

The user pastes a URL.

GoVideo analyzes the source, shows metadata and then allows the user to choose what to produce:

- Video
- Audio/music
- Both

The user selects a destination and starts the operation.

### Convert an existing local video to audio

The user selects a local video file from the GUI.

GoVideo extracts or converts the audio and saves the requested audio output.

This workflow does not require `yt-dlp`; it is primarily an FFmpeg workflow.

### Multiple operations

The application should eventually support a queue/list of jobs with:

- progress;
- speed;
- ETA when available;
- status;
- cancellation;
- failures;
- completed output path.

Concurrency must be bounded and controlled by the Go backend.

## Product goals

- Simple installation.
- No manual `yt-dlp` installation for normal users.
- Attractive desktop GUI.
- Responsive UI while downloads and conversions run.
- Clean separation between UI and media logic.
- Linux and Windows support.
- Replaceable frontend layer without replacing the core.
- Robust handling of external dependencies.
- Good error messages for ordinary users while preserving detailed backend errors for diagnostics.

## Non-goals for the initial versions

- DRM circumvention.
- Reimplementing `yt-dlp` extractors in Go.
- Becoming a full video editor.
- Advanced timeline editing.
- Media library/server functionality.
- Mobile support.
- macOS support in the initial platform scope.

## High-level technology

```text
Desktop shell     Wails 3
Frontend          Vue 3 + TypeScript
Backend/Core      Go
Media extraction  yt-dlp
Media processing  FFmpeg
Browser fallback  Browser automation, planned
Platforms         Linux + Windows
```

## Wails

GoVideo targets Wails 3.

As of 2026-09-24, Wails 3 is in beta. Desktop APIs have a beta compatibility commitment, but the project should pin a known working beta version rather than blindly tracking `latest`.

Wails uses the native system WebView:

- Linux: WebKitGTK stack;
- Windows: WebView2.

The Wails layer should expose application services and events, not contain core download logic.

## Naming

The application and new architecture are named **GoVideo**.

Legacy repository/script references may still use `downvideo`. Rename work should be deliberate and should not obscure history or break imports without a controlled migration.
