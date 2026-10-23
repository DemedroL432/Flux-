# Beta testing guide

## Build artifacts

- Portable folder: `dist/pulseshield-win64`
- NSIS installer (after `.\build.ps1 -Version x.y.z -BuildInstaller`): `dist/PulseShield-vx.y.z-Setup.exe`

## What testers should do

1. Add a server:
   - Copy a `vless://...` link to clipboard
   - `Servers` → `Add from clipboard`
   - Select server → `Set selected active`
2. Enable and apply:
   - `Proxy enabled` ON, mode `system_proxy`
   - (Optional) `DPI bypass enabled` ON
   - Go to `Dashboard` → `Apply state`
3. Verify:
   - Open `Diagnostics` → `Test connection (via local proxy)`
   - Confirm apps start working through proxy
4. If something breaks:
   - `Diagnostics` → `Export diagnostics bundle (zip)`
   - Send the resulting `.zip` to the developer

## Developer notes

- Logs are written to `%LocalAppData%\\PulseShield\\logs`
- Runtime config is written to `%LocalAppData%\\PulseShield\\runtime`
