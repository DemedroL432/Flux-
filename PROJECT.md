# Project: PulseShield (Flux) Desktop Client

## Architecture
- **Desktop Shell**: Go backend with Wails v2 runtime embedding React/Vite web application. Frameless fixed 1280×720 window with custom HTML/CSS drag regions (`--wails-draggable: drag`).
- **Frontend Layer**: React 18 + Vite 3 + TypeScript. Cyberpunk Glassmorphism Graphite theme (`GraphiteTheme.css`). Pages: Dashboard, Servers, Routing, DPI, Diagnostics, Logs, Settings. Monolithic right panel HUD layout.
- **Proxy Engine**: sing-box (VLESS-reality / shadowsocks), configuration generation, route rules, TCP liveness probing, port kill cleanup (`KillStraySingBoxOnListenPort`).
- **DPI Bypass Engine**: Zapret (`winws2.exe`, WinDivert kernel driver), strategy presets (`third_party/zapret/presets/`), domain/IP hostlists (`third_party/zapret/lists/`), filter files.
- **Routing Engine**: Strategy profiles (`bypass-russia` -> `rf-smart-split`, `all-proxy` -> `default`, `all-direct`). Final outbounds (`proxy` vs `direct`).
- **Packaging & Build**: `build.ps1` invoking `scripts/build_windows.ps1` with Wails CLI, Go compiler, and portable zip compression producing `dist/pulseshield-win64` and `dist/Flux-v0.7.3-beta-win64.zip`.

## Feature Inventory
| # | Feature | Description | Milestone | Source |
|---|---------|-------------|-----------|--------|
| 1 | Window Framing & Geometry | 1280x720 frameless window, non-resizable, draggable topbar and sidebar brand | M1 | R1, ORIGINAL_REQUEST §14 |
| 2 | Dashboard Scrollbar Elimination | Budgeted layout height, overflow-y: hidden, absolute x-axis chart positioning | M1 | R1, ORIGINAL_REQUEST §15, §33 |
| 3 | Services Health Status 4-Column Grid | Clean 2-row card: header row + 4 symmetric columns (sing-box, WinDivert, System Proxy, DNS) without text clipping or wrapping | M1 | R1, ORIGINAL_REQUEST §15, §33 |
| 4 | Vector SVG Country Flags | Reusable CountryFlag component with SVG tricolor for Russia (RU), DE, NL, FI, GB, US, World; remove emoji glyph fallbacks | M1 | R1, ORIGINAL_REQUEST §16, §34 |
| 5 | Flux Typography & Branding | Space Grotesk font styling for .flux-hero-wordmark to eliminate fuzzy rendering | M1 | R1, ORIGINAL_REQUEST §34 |
| 6 | Monolithic Right Panel | Unified .flux-right-monolith container with hairline dividers across all 7 pages | M1 | R1, ORIGINAL_REQUEST §35 |
| 7 | WinDivert & Zapret Process Lifecycle | Startup stray process cleanup for winws/winws2, driver stop/delete (Monkey, WinDivert), fix Runner.Stop() nil guard | M2 | R2, ORIGINAL_REQUEST §19, §38 |
| 8 | Routing Profile Compatibility | Support bypass-russia, all-proxy, all-direct aliases in Go backend matching UI selectors | M2 | R2, ORIGINAL_REQUEST §20 |
| 9 | Health Check Event Synchronization | Immediate GetServiceHealth trigger upon runtime-apply / setDpi, dynamic port binding | M2 | R2, ORIGINAL_REQUEST §19, §38 |
| 10 | Zapret Preset Integrity | Validation of all 93 preset references via check_zapret_presets.ps1 (0 missing) | M2 | R2, ORIGINAL_REQUEST §21, §39 |
| 11 | Zone.Identifier Cleanup | Removal of 26 corrupted alternate stream files in third_party/zapret | M3 | R3, Explorer Survey |
| 12 | Version Metadata Alignment | Synchronize version 0.7.3-beta across wails.json, package.json, main.go, build.ps1 | M3 | R3, Explorer Survey |
| 13 | Portable ZIP Packaging | Add Compress-Archive step in build.ps1 / build_windows.ps1 to generate dist/Flux-v0.7.3-beta-win64.zip | M3 | R3, ORIGINAL_REQUEST §26, §43 |
| 14 | Automated Verification & Build Pipeline | Full TypeScript/Vite check, preset check, release build Flux.exe | M3 | R3, ORIGINAL_REQUEST §24-26, §31, §42 |

## Milestones
| # | Name | Scope | Dependencies | Status |
|---|------|-------|-------------|--------|
| M1 | UI/UX & Design System Refinement | Features 1-6: SVG flags, 4-col health card, dashboard no-scroll, monolithic right panel, Space Grotesk logo, frontend build | none | DONE |
| M2 | Network Core Lifecycle & Routing | Features 7-10: WinDivert/winws process cleanup, Stop() fix, routing profile aliases, health sync, preset validator | M1 | DONE |
| M3 | Release Build & Packaging Verification | Features 11-14: Zone.Identifier cleanup, version alignment, ZIP packaging in build scripts, full release build & verification | M1, M2 | IN_PROGRESS |

## Interface Contracts
### Frontend ↔ Go Backend (Wails Runtime)
- `GetUIState()` -> returns `UIState` with `RoutingProfile` ("bypass-russia" | "all-proxy" | "all-direct" | "rf-smart-split" | "default").
- `SetRoutingProfile(name string)` -> accepts "bypass-russia", "all-proxy", "all-direct" and maps to backend profile rules.
- `GetServiceHealth()` -> returns `ServiceHealth`:
  - `ProxyEnabled`: bool
  - `ProxyListenPort`: int (e.g. 8080)
  - `ProxyLocalOK`: bool
  - `VLESSRemote`: bool
  - `DPIRunning`: bool
  - `ActiveDPIProfile`: string
  - `DpiTool`: string
- `ApplyRuntime(req)` / `ApplyRuntimeDeferred(req)`: non-blocking proxy/DPI apply, emits `runtime-apply` Wails event.

### CountryFlag Component Contract
- Props: `{ code?: string, name?: string, className?: string, style?: CSSProperties }`
- Supported codes/names: "ru" / "russia" / "россия" / "москва" -> SVG Russian Tricolor (White-Blue-Red).
- Supported: "de" / "germany", "nl" / "netherlands", "fi" / "finland", "gb" / "uk", "us" / "usa", fallback -> World SVG.

## Code Layout
- `ui/PulseShieldDesktop/frontend/src/`:
  - `components/CountryFlag.tsx`: Vector SVG flags component.
  - `components/DesktopTopBar.tsx`, `components/DesktopSidebar.tsx`, `components/DesktopBanner.tsx`.
  - `pages/DashboardPage.tsx`: Dashboard HUD, 4-col services health, no scrollbar.
  - `pages/ServersPage.tsx`, `NetworkPage.tsx`, `DpiPage.tsx`, `DiagnosticsPage.tsx`, `LogsPage.tsx`, `SettingsPage.tsx`: Monolithic right panels.
  - `GraphiteTheme.css`: Cyberpunk styling, CSS grid, drag zones, typography.
- `app/`:
  - `config/routing_presets.go`, `models/models.go`: Routing profile mapping and aliases.
  - `dpi/runner.go`, `dpi/paths.go`: WinDivert / winws2 process management and cleanup.
  - `core/manager.go`, `core/adapter.go`: sing-box lifecycle.
- `ui/PulseShieldDesktop/`:
  - `app.go`: Wails bindings, health checks, event dispatch.
  - `main.go`: Window geometry, Wails lifecycle hooks.
  - `wails.json`: Project version and build config.
- `scripts/`:
  - `build_windows.ps1`, `build.ps1`: Release build and zip packaging.
  - `check_zapret_presets.ps1`: Zapret preset validator.
