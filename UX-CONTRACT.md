# GoVideo UX Contract

Native selects are intentional: GoVideo accepts platform-owned popup geometry because it is a Linux/Windows desktop utility and benefits from familiar OS keyboard behavior. Native dialogs select files and folders; application modals use `<dialog>`.

| Capability | Canonical owner | Source of truth | Allowed variants | Verification |
|---|---|---|---|---|
| Select/Listbox | Native `select` | `frontend/src/App.vue` | Compact form fields | Keyboard and platform popup check |
| Form | Vue semantic form controls | `frontend/src/App.vue` | URL, media options, settings | Typecheck and keyboard check |
| Scrollbar | Global CSS baseline | `frontend/src/style.css` | Browser engine fallback | Narrow/overflow check |
