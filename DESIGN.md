# GoVideo Design System

GoVideo is a calm, focused desktop utility. Its visual language uses open layouts, strong alignment and a single cobalt action color; it avoids dashboard clutter and decorative cards.

## Tokens

- Canvas: `#f5f7fb`; surface: `#ffffff`.
- Ink: `#172538`; muted: semantic content at 70%; border: `#dce4ef`.
- Brand: `#0866e6`; focus follows DaisyUI primary.
- Success: `#07834c`; warning: `#aa6200`; danger: `#bd2834`.
- Radius: 8px controls, 16px panels.
- Shadow: `0 18px 50px rgba(23, 53, 96, .10)` only for overlays and the app surface.
- Spacing follows a 4px base. Main gutters are 24–40px.
- Motion: 140–180ms ease-out; disabled under `prefers-reduced-motion`.

## Typography and components

Inter is the primary family with a system sans fallback. Body text is 14–16px; screen headings are 26–32px. Controls are at least 40px high, focus is always visible, and icons use a consistent 2px outline. Buttons combine emphasis (`solid`, `outline`, `ghost`) with semantic intent.

The app shell has a compact top bar, two main destinations, and one settings action. Home uses an open work surface. Jobs use rows rather than card grids. Dialogs are native `<dialog>` elements with an opaque surface and focus restoration. Narrow layouts stack controls and preserve 44px targets.

DaisyUI is the canonical component implementation: `btn`, `input`, `select`, `radio`, `badge`, `alert`, `loading`, `list`, `radial-progress`, `modal` and `modal-box`. Custom CSS defines application layout and theme tokens, not replacement controls. Activity rings are approximately 50px, with blue active progress and green completion; failed/cancelled tasks show a status icon instead of a misleading percentage.

## Themes

`govideo-light` and `govideo-dark` map the same semantic tokens. The system preference is applied before Vue mounts; users can override it in Settings.
