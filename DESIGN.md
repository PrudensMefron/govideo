# GoVideo Design System

GoVideo is a calm, focused desktop utility. Its visual language uses open layouts, strong alignment and a single cobalt action color; it avoids dashboard clutter and decorative cards.

## Tokens

- Canvas: `#f4f8ff`; surface: `#ffffff`; elevated surface: `#f8faff`.
- Ink: `#101828`; muted: `#667085`; border: `#d8e1ef`.
- Brand: `#0866f5`; brand hover: `#0758d4`; focus: `#73a7ff`.
- Success: `#079455`; warning: `#dc6803`; danger: `#d92d20`.
- Radius: 10px controls, 14px panels, 18px primary surface.
- Shadow: `0 18px 50px rgba(23, 53, 96, .10)` only for overlays and the app surface.
- Spacing follows a 4px base. Main gutters are 24–40px.
- Motion: 140–180ms ease-out; disabled under `prefers-reduced-motion`.

## Typography and components

Inter is the primary family with a system sans fallback. Body text is 14–16px; screen headings are 26–32px. Controls are at least 40px high, focus is always visible, and icons use a consistent 2px outline. Buttons combine emphasis (`solid`, `outline`, `ghost`) with semantic intent.

The app shell has a compact top bar, two main destinations, and one settings action. Home uses an open work surface. Jobs use rows rather than card grids. Dialogs are native `<dialog>` elements with an opaque surface and focus restoration. Narrow layouts stack controls and preserve 44px targets.

## Themes

`govideo-light` and `govideo-dark` map the same semantic tokens. The system preference is applied before Vue mounts; users can override it in Settings.
