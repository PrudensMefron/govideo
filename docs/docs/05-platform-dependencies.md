# GoVideo — Platforms and Managed Dependencies

## Current implementation

yt-dlp is pinned to `2026.08.19`. The Linux and Windows official standalone assets use embedded official SHA-256 values, bounded HTTPS download, a temporary file, fsync, digest validation, atomic rename, and post-install version validation. Installation is serialized by the dependency manager.

FFmpeg is currently detected from `PATH` and surfaced with attribution to `ffmpeg.org/legal.html`. The pinned BtbN LGPL archive installer described below remains planned and is not reported as managed.

## Supported product platforms

Initial product scope:

```text
Linux
Windows
```

Development may be Linux-first, but platform assumptions must remain isolated.

## Application data

Managed executables should live in an application-owned data directory, not in the repository and not necessarily in the user's global PATH.

### Linux

Follow XDG conventions.

Preferred root:

```text
$XDG_DATA_HOME/govideo
```

Fallback:

```text
~/.local/share/govideo
```

Example:

```text
~/.local/share/govideo/
├── bin/
│   ├── yt-dlp
│   └── ffmpeg
└── ...
```

### Windows

Use the appropriate per-user application data location.

Do not manually approximate Windows paths from environment variables in multiple packages. Centralize platform path resolution.

A likely shape is:

```text
<user app data>/GoVideo/bin/
```

Exact resolution should use Go/platform facilities and be tested on Windows.

## yt-dlp

### Strategy

`yt-dlp` is a managed external dependency.

It is intentionally not embedded into the Go binary because:

- it changes frequently;
- site extractor fixes are released independently of GoVideo;
- managed replacement avoids requiring a GoVideo release for every yt-dlp update.

### Linux initial implementation

Initial supported tuple:

```text
GOOS=linux
GOARCH=amd64
```

Use the official standalone Linux asset.

Installation pipeline:

```text
check
  ↓ missing/invalid
download official asset
  ↓
temporary file
  ↓
SHA-256 verification (required before production-ready)
  ↓
sync + close
  ↓
permissions
  ↓
atomic rename
  ↓
execute --version
  ↓
ready
```

### Windows planned implementation

Use the official Windows standalone executable asset.

Differences that must be isolated:

- `.exe` name;
- asset selection;
- executable permission semantics;
- path conventions;
- process behavior.

## FFmpeg

FFmpeg is a separate managed dependency.

Do not assume the yt-dlp binary provides FFmpeg.

FFmpeg enables:

- merge of separated streams;
- extraction of audio;
- transcoding;
- local video → audio;
- some requested output formats.

The dependency manager should eventually locate:

```text
ffmpeg
ffprobe
```

as required by the implementation.

### Distribution caution

Before shipping FFmpeg builds with GoVideo, choose an appropriate distribution source and confirm licensing/build configuration implications.

Do not randomly download executable archives from third-party mirrors inside production code.

## Browser runtime

Browser fallback is planned, not foundational.

Possible implementation may require browser automation/runtime assets.

Keep browser runtime management separate from yt-dlp/FFmpeg.

## Wails platform runtime

### Linux

Wails 3 uses the system WebKitGTK stack rather than bundling Chromium.

Current Wails 3 beta documentation specifies GTK4 + WebKitGTK 6.0 as the default Linux stack.

Linux packaging must account for runtime/system dependencies.

### Windows

Wails uses WebView2.

WebView2 availability/install behavior is a packaging concern separate from GoVideo's media dependencies.

## Version policy

### Wails

Pin a known working Wails 3 beta release during the beta period.

Do not use an unconstrained `latest` in reproducible builds.

### yt-dlp

The application should be able to update yt-dlp independently of GoVideo.

Initial implementation may only install when missing.

Update checks can be added later.

### FFmpeg

Pin/track a known compatible build strategy.

Do not update silently in the middle of an active media job.

## Dependency manager API direction

Keep dependency-specific operations behind a manager/service.

Conceptually:

```go
type DependencyStatus struct {
    Name      string
    Path      string
    Version   string
    Installed bool
    Valid     bool
}

func (m *Manager) EnsureYTDLP(ctx context.Context) (DependencyStatus, error)
func (m *Manager) EnsureFFmpeg(ctx context.Context) (DependencyStatus, error)
```

Exact signatures may evolve.

Avoid one giant generic installer with dozens of switches if per-tool behavior differs materially.

## Concurrency

Only one installation/update per dependency should run at a time.

Two simultaneous GUI requests must not race to replace `yt-dlp`.

Use synchronization or singleflight-style behavior around dependency installation.

## Security

Dependency downloads are executable code.

Required hardening:

- HTTPS;
- trusted official source;
- explicit status checking;
- maximum/expected-size sanity where useful;
- SHA-256 verification;
- atomic replacement;
- no execution before validation;
- never log secrets from authenticated browser/media headers.
