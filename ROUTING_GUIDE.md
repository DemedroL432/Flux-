# PulseShield Routing Config Guide

This guide explains how to configure routing profiles so traffic can go through:
- `proxy` (VLESS outbound),
- `direct` (bypass proxy),
- `block` (drop traffic).

## Hierarchy of truth (merged rule order)

PulseShield builds sing-box `route.rules` by merging layers in **`MergeSingBoxRoutingRules`** (`app/config/default_routing.go`).  
**First matching rule wins** (sing-box evaluates rules top → bottom).

| Priority | Source | What it is |
|----------|--------|------------|
| **1 (highest)** | **Process map** | Per-process routing from the app (e.g. send `chrome.exe` via proxy). Injected as synthetic route rules first. |
| **2** | **User rules** | `userRoutingRules` in `app-config.json` — rows you add in the UI tables («Умный сплит»). |
| **3** | **Active profile** | The selected `routingProfile` and its `rules` array from `routingProfiles`. |
| **4 (lowest)** | **System library** | `DefaultRoutingLibraryRules`: built-in lists (ChatGPT, Netflix, Cloudflare CIDRs, anticheat/direct domains, optional Zapret list files). **Skipped entirely when `strictRoutingMode` is `true`** («чистый гейминг» — only process map + user + profile). |

### Final action (unmatched traffic)

If **no** route rule matches, sing-box uses the **final outbound** from `route.final`:

1. If the active routing profile defines `final` as `direct` or `proxy`, that wins.
2. Otherwise:
   - `proxyMode: direct_with_rules` → final **direct** (save VLESS traffic).
   - `system_proxy` / `tun` → final **proxy** (typical full-VPN behaviour).

See `routeFinalOutbound` in `app/core/singbox_config.go`.

### Strict routing mode (`strictRoutingMode`)

When `true`, the built-in library layer is **not** merged. Use this when you want **only** your profile + manual rules + process map (e.g. competitive gaming — nothing extra from the library should steer traffic).

---

## Where routing is stored

Routing profiles are part of `%LOCALAPPDATA%/PulseShield/app-config.json` in the `routingProfiles` array.

Active profile is selected by:
- `routingProfile`: profile name

Proxy behavior baseline is selected by:
- `proxyMode`: `system_proxy`, `tun`, or `direct_with_rules`

## Important behavior

- **Merged** routing uses the hierarchy above (process map → user → profile → library). Unmatched traffic still follows **final outbound** (profile `final`, then `proxyMode`).
- Rule order matters within each layer; merged list order reflects that priority.

## Rule fields

Each rule supports:
- `name`: label
- `domains`: list of domains (sing-box `domain`)
- `domainSuffixes`: list of suffixes without leading dot, e.g. `ru`, `xn--p1ai` for `.рф` (sing-box `domain_suffix`)
- `cidrs`: list of CIDR ranges
- `processes`: list of process names (for example, `chrome.exe`)
- `action`: one of `proxy`, `direct`, `block`, `dpi-bypass`

Notes:
- `dpi-bypass` is treated as `direct` on proxy routing level (same path as DPI tools on direct TCP).
- Use lowercase domains where possible.

## Built-in presets (merged into config if missing)

| Name | Role |
|------|------|
| `default` | Everything through VLESS except local/example block rules. |
| `rf-smart-split` | RF-oriented split: LAN + `.ru`/`.рф`/`.su` + game launchers + large “DPI-direct” list → direct; `final` → direct so a weak VLESS does not break YouTube-style sites. |
| `selective-proxy` | Only listed domains → proxy; `final` → direct. |
| `gaming-friendly` | Launchers direct; listed domains → proxy; `final` → direct. |

## Example profile: proxy by default, direct for local

```json
{
  "name": "default",
  "rules": [
    {
      "name": "Local LAN direct",
      "cidrs": ["10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"],
      "action": "direct"
    },
    {
      "name": "Windows update direct",
      "domains": ["windowsupdate.com", "download.windowsupdate.com"],
      "action": "direct"
    },
    {
      "name": "Block telemetry sample",
      "domains": ["telemetry.example.com"],
      "action": "block"
    },
    {
      "name": "Browser via proxy",
      "processes": ["chrome.exe", "msedge.exe", "firefox.exe"],
      "action": "proxy"
    }
  ]
}
```

## Example profile: direct by default, proxy only selected apps

Set:
- `proxyMode: "direct_with_rules"`

Then add rules that explicitly send traffic to proxy:

```json
{
  "name": "gaming-direct-base",
  "rules": [
    {
      "name": "Discord proxy",
      "domains": ["discord.com", "discord.gg"],
      "action": "proxy"
    },
    {
      "name": "Game launcher proxy",
      "processes": ["launcher.exe"],
      "action": "proxy"
    }
  ]
}
```

## Quick validation checklist

- Profile name in `routingProfile` exists in `routingProfiles`.
- Each rule has at least one matcher (`domains`, `cidrs`, or `processes`).
- `action` is valid.
- CIDRs are valid notation (`x.x.x.x/n`).
- If using process rules, process names include `.exe`.

