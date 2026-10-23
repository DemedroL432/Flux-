# PulseShield Quick Test Checklist

Use this checklist to verify that proxy and DPI work independently and together.

## 1) Smoke start

- Launch `PulseShield.exe`
- Confirm UI opens and shows current status
- Open `Diagnostics` and run `Self-check`

Expected:
- no startup crash
- config and logs paths are present

## 2) DPI-only mode

- Disable proxy
- Enable DPI
- Choose a Zapret profile
- Press `Start` in DPI section

Expected:
- DPI status becomes running
- Proxy remains disabled
- App stays responsive

## 3) Proxy-only mode

- Disable DPI
- Enable proxy
- Select active VLESS server
- Press `Apply network state`

Expected:
- proxy connection starts
- DPI remains stopped
- `Test proxy` in Diagnostics succeeds (if server is reachable)

## 4) Combined mode (no conflict)

- Enable both proxy and DPI
- Press `Apply network state`

Expected:
- both features stay active simultaneously
- no forced stop of one by another
- logs for both core and DPI are updating

## 5) Routing profile behavior

- Select `proxyMode = direct_with_rules`
- Pick routing profile
- Apply

Expected:
- unmatched traffic defaults to direct
- rules with `action: proxy` still route through proxy

Then switch to `system_proxy`:

Expected:
- unmatched traffic defaults to proxy

## 6) Stop all

- Press `Stop All`

Expected:
- Proxy and DPI both stop
- UI reflects disabled state

## 7) Routing editor/import/export

- Open Routing page
- `Load JSON`
- Change one rule
- `Apply JSON`
- `Export file`
- `Import file` back

Expected:
- profile updates without validation errors
- JSON round-trip works

