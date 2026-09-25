# GoVideo UX Contract

Native selects are intentional: GoVideo accepts platform-owned popup geometry because it is a Linux/Windows desktop utility and benefits from familiar OS keyboard behavior. Native dialogs select files and folders; application modals use `<dialog>`.

| Capability | Canonical owner | Source of truth | Allowed variants | Verification |
|---|---|---|---|---|
| Select/Listbox | DaisyUI `select` on native `select` | HomeView, MediaDialog, SettingsDialog | Compact form fields | Keyboard and platform popup check |
| Form | DaisyUI `input`, `radio`, `btn` | `frontend/src/views/`, `frontend/src/components/` | URL, media options, settings | Typecheck and keyboard check |
| Job progress | DaisyUI `radial-progress` | `frontend/src/views/ActivitiesView.vue` | Blue active, green completed, unknown total without numeric value | Render fixtures at desktop and narrow widths |
| Modal | Native dialog + DaisyUI `modal` / `modal-box` | MediaDialog, SettingsDialog | Scrollable body, fixed header/footer | Escape, focus return, overflow |
| Scrollbar | Global CSS baseline | `frontend/src/style.css` | Browser engine fallback | Narrow/overflow check |

Do not use DaisyUI component names (`progress`, `status`) for unrelated layout wrappers. Activity layout uses `gv-job-progress` and explicit grid placement so DaisyUI list-row rules cannot overlap content at responsive breakpoints. Business state and Wails subscriptions belong to `useGoVideo`, not view components.
