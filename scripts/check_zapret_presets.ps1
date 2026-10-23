# Validates preset file references against third_party/zapret (same layout as dist/.../zapret).
# Run from repo root: pwsh -File scripts/check_zapret_presets.ps1
# Exit 0 = OK, 1 = missing files or bad service_config.

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$base = Join-Path $root "third_party\zapret"
if (-not (Test-Path $base)) {
  Write-Error "Zapret bundle not found: $base"
  exit 1
}

$refs = @{}
Get-ChildItem (Join-Path $base "presets\*.txt") -ErrorAction SilentlyContinue | ForEach-Object {
  Get-Content $_.FullName -Encoding UTF8 | ForEach-Object {
    $line = $_
    if ($line -match '^\s*#') { return }
    foreach ($m in [regex]::Matches($line, '@(bin/[^\s#]+|lua/[^\s#]+|windivert\.filter/[^\s#]+)')) {
      $refs[$m.Groups[1].Value] = $true
    }
    foreach ($m in [regex]::Matches($line, '(?:hostlist|ipset|ipset-exclude|hostlist-exclude)=(lists/[^\s#]+)')) {
      $refs[$m.Groups[1].Value] = $true
    }
  }
}

$missing = [System.Collections.Generic.List[string]]::new()
foreach ($rel in $refs.Keys | Sort-Object) {
  $full = Join-Path $base $rel
  if (-not (Test-Path -LiteralPath $full)) {
    [void]$missing.Add($rel)
  }
}

$svc = Join-Path $base "lists\service_config.txt"
$badSvc = $false
if (Test-Path $svc) {
  $raw = Get-Content $svc -Raw -Encoding UTF8
  if ($raw -match '(?i)[a-z]:\\') {
    $badSvc = $true
    Write-Warning "lists/service_config.txt contains a Windows absolute path (foreign install metadata)."
  }
}

Write-Host "Presets: scanned references: $($refs.Count)"
if ($missing.Count -gt 0) {
  Write-Host "MISSING ($($missing.Count)):" -ForegroundColor Red
  $missing | ForEach-Object { Write-Host "  $_" }
  exit 1
}
Write-Host "All referenced paths exist under zapret root." -ForegroundColor Green
if ($badSvc) { exit 1 }
exit 0
