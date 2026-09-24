# GoVideo Documentation Pack

This documentation pack captures the current product direction and implementation architecture for GoVideo.

It is intended primarily to guide AI coding agents such as Codex, while remaining useful to human contributors.

## Files

- `AGENTS.md` — mandatory coding-agent instructions and architectural rules.
- `docs/00-overview.md` — product purpose and scope.
- `docs/01-current-state.md` — what exists, what has been designed, and what is still missing.
- `docs/02-architecture.md` — target backend/frontend architecture and boundaries.
- `docs/03-product-requirements.md` — functional and non-functional requirements.
- `docs/04-implementation-roadmap.md` — ordered implementation plan.
- `docs/05-platform-dependencies.md` — Linux/Windows and external dependency strategy.
- `docs/06-frontend-wails.md` — Wails 3 + Vue 3 frontend guidance.
- `docs/07-auto-updater.md` — GitHub Releases updater and draft-release pipeline.

## Recommended agent reading order

```text
AGENTS.md
docs/01-current-state.md
docs/00-overview.md
docs/02-architecture.md
docs/03-product-requirements.md
docs/04-implementation-roadmap.md
```

Then read the task-relevant platform/frontend document.

## Maintenance rule

Documentation must not claim a feature is implemented merely because it is planned.

When code lands, update `docs/01-current-state.md`.
